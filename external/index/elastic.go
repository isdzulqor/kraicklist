package index

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"kraicklist/helper/errors"
	"kraicklist/helper/jsons"
	"kraicklist/helper/logging"
	"strings"
	"sync/atomic"
	"time"

	es7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esutil"
)

const (
	prefixElastic = "external-elastic:"
	IndexElastic  = "elastic"
)

type ElasticDoc struct {
	ID   string
	Data interface{}
}

type ElasticDocs []ElasticDoc

type ElasticDocError struct {
	DocID string
	err   error
}

type ElasticRootQuery struct {
	Query interface{} `json:"query"`
}

func (e *ElasticRootQuery) ConstructElasticMultiMatchQuery(query string, fields ...string) {
	e.Query = map[string]interface{}{
		"multi_match": ElasticMultiMatchQuery{
			Query:        query,
			Fields:       fields,
			Fuzziness:    "AUTO", // TODO: revisit for flexibility
			PrefixLength: 2,      // TODO: revisit for flexibility, The number of initial characters which will not be “fuzzified”
		}}
}

type ElasticMultiMatchQuery struct {
	Query        string   `json:"query"`
	Fields       []string `json:"fields"`
	Fuzziness    string   `json:"fuzziness"`
	PrefixLength int      `json:"prefix_length"`
}

type ElasticQueryResult struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore interface{} `json:"max_score"`
		// Hits     []interface{} `json:"hits"`
		Hits []struct {
			Index  string      `json:"_index"`
			Type   string      `json:"_type"`
			ID     string      `json:"_id"`
			Score  float64     `json:"_score"`
			Source interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (q ElasticQueryResult) GetHitSources() (documentFields []interface{}) {
	for _, hit := range q.Hits.Hits {
		documentFields = append(documentFields, hit.Source)
	}
	return
}

type ElasticDocErrors []ElasticDocError

// TODO: convert to error lib
func (errorDocs ElasticDocErrors) ToError() error {
	return nil
}

type ElasticIndex struct {
	esClient  *es7.Client
	indexName string
}

func InitESIndex(ctx context.Context, elasticHost []string, username, password, indexName string) (*ElasticIndex, error) {
	es7, err := es7.NewClient(es7.Config{
		Addresses: elasticHost,
		Username:  username,
		Password:  password,
	})
	if err != nil {
		return nil, fmt.Errorf("%s failed to initiate ES7 client, err: %v", prefixElastic, err)
	}

	logging.InfoContext(ctx, "%s ES7 client is initialized", prefixElastic)
	return &ElasticIndex{
		esClient:  es7,
		indexName: indexName,
	}, nil
}

func (es *ElasticIndex) BulkIndexDocs(ctx context.Context, docs ElasticDocs) (docErrors *ElasticDocErrors, err error) {
	var countSuccessful uint64
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         es.indexName,
		Client:        es.esClient,
		NumWorkers:    100,              // The number of worker goroutines
		FlushBytes:    int(5e+6),        // The flush threshold in bytes
		FlushInterval: 30 * time.Second, // The periodic flush interval
	})
	if err != nil {
		logging.ErrContext(ctx, "failed to init bulk indexer, err: %v", err)
		err = errors.ErrorThirdParty
		return
	}

	for _, doc := range docs {
		data, _ := json.Marshal(doc.Data)
		err = bi.Add(
			ctx,
			esutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: doc.ID,
				Body:       bytes.NewReader(data),
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(&countSuccessful, 1)
				},
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if docErrors == nil {
						docErrors = &ElasticDocErrors{}
					}
					if err != nil {
						logging.WarnContext(ctx, "failed to index doc with ID %d, err: %v", doc.ID, err)
					} else {
						err = fmt.Errorf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
						logging.WarnContext(ctx, "failed to index doc with ID %d, err: %v", doc.ID, err)
					}
					*docErrors = append(*docErrors, ElasticDocError{
						DocID: doc.ID,
						err:   err,
					})
				},
			},
		)
		if err != nil {
			logging.ErrContext(ctx, "failed to add doc %s on bulk index, err: %v", doc.ID, err)
		}
	}
	if err = bi.Close(ctx); err != nil {
		logging.ErrContext(ctx, "failed to close bulk index, err: %v", err)
		return
	}

	biStats := bi.Stats()
	if biStats.NumFailed == biStats.NumAdded {
		logging.WarnContext(ctx, "%s failed to index docs, bulk index ES7 status: %v", prefixElastic,
			jsons.ToStringJsonNoError(biStats))
		err = errors.ErrorThirdParty
		return
	}

	logging.DebugContext(ctx, "successfully indexing %d docs", biStats.NumIndexed)
	return
}

func (es *ElasticIndex) SearchQuery(ctx context.Context, query ElasticRootQuery, dest interface{}) (result ElasticQueryResult, err error) {
	data, err := json.Marshal(query)
	if err != nil {
		err = fmt.Errorf("%s failed to marshal ElasticRootQuery", prefixElastic)
		return
	}

	// Perform the search request.
	res, err := es.esClient.Search(
		es.esClient.Search.WithContext(ctx),
		es.esClient.Search.WithIndex(es.indexName),
		es.esClient.Search.WithBody(strings.NewReader(string(data))),
		es.esClient.Search.WithTrackTotalHits(true),
		es.esClient.Search.WithPretty(),
	)
	if err != nil {
		logging.ErrContext(ctx, "failed to search, err: %v", err)
		err = errors.ErrorThirdParty
		return
	}
	defer res.Body.Close()

	if res.IsError() {
		logging.WarnContext(ctx, "failed to search query, resp: %s", res.String())
		err = errors.ErrorThirdParty
		return
	}

	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		logging.ErrContext(ctx, "failed to decode, err: %v", err)
		err = fmt.Errorf("%s %v", prefixElastic, err)
		return
	}

	bytes, err := json.Marshal(result.GetHitSources())
	if err != nil {
		err = fmt.Errorf("%s %v", prefixElastic, err)
		return
	}
	if err = json.Unmarshal(bytes, dest); err != nil {
		err = fmt.Errorf("%s %v", prefixElastic, err)
		return
	}

	return
}

func (es *ElasticIndex) DeleteIndex(ctx context.Context) (err error) {
	res, err := es.esClient.Indices.Delete([]string{es.indexName},
		es.esClient.Indices.Delete.WithIgnoreUnavailable(true))
	if err != nil {
		err = fmt.Errorf("%s cannot delete index, err: %v", prefixElastic, err)
		return
	}
	if res.IsError() {
		err = fmt.Errorf("%s cannot delete index, err resp: %v", prefixElastic, res.String())
	}
	defer res.Body.Close()
	return
}

func (es *ElasticIndex) Ping() error {
	_, err := es.esClient.Cluster.Health()
	if err != nil {
		err = fmt.Errorf("%s %v", prefixElastic, err)
		logging.ErrContext(context.Background(), "%v", err)
		return err
	}
	return nil
}

func (es *ElasticIndex) PingWithRetry(retry int, waitTime time.Duration) (err error) {
	ctx := context.Background()
	for i := retry; i > 0; i-- {
		if err = es.Ping(); err == nil || i == 1 {
			break
		}
		logging.WarnContext(ctx, "can't ping to elastic cluster, Wait 5 seconds. %d retries left...", i-1)
		time.Sleep(5 * time.Second)
	}
	if err == nil {
		logging.DebugContext(ctx, "successfully ping to elastic cluster")
	}
	return
}
