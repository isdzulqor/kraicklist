package index

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/isdzulqor/kraicklist/helper/logging"

	"github.com/blevesearch/bleve"
)

const (
	prefixBleve = "external-bleve:"
	IndexBleve  = "bleve"
)

type SearchResultCustom bleve.SearchResult

func (s SearchResultCustom) GetDocumentFields() (documentFields []map[string]interface{}) {
	for _, doc := range s.Hits {
		documentFields = append(documentFields, doc.Fields)
	}
	return
}

type BleveDoc struct {
	ID   string
	Data interface{}
}

type BleveDocs []BleveDoc

type BleveDocError struct {
	DocID string
	err   error
}

type BleveDocErrors []BleveDocError

// TODO: convert to error lib
func (errorDocs BleveDocErrors) ToError() error {
	return nil
}

type BleveIndex struct {
	clientIndex bleve.Index
	indexName   string
}

// TODO: utilize context
func InitBleveIndex(ctx context.Context, indexName string) (out *BleveIndex, err error) {
	docPath := "./data/" + indexName
	index, err := bleve.Open(docPath)
	if err != nil {
		logging.WarnContext(ctx, "%s failed to open index %s, will create new one", prefixBleve, docPath)
		if index, err = bleve.New(docPath, bleve.NewIndexMapping()); err != nil {
			err = fmt.Errorf("%s failed to creaete new index %s, err: %v", prefixBleve, docPath, err)
			return
		}
	}
	logging.InfoContext(ctx, "%s bleve index is initialized", prefixBleve)
	out = &BleveIndex{
		clientIndex: index,
		indexName:   indexName,
	}
	return
}

// TODO: debug logging
func (index *BleveIndex) SearchQuery(ctx context.Context, keyword string, dest interface{}) (result SearchResultCustom, err error) {
	if keyword == "" {
		err = fmt.Errorf("keyword can't be empty")
		return
	}

	searchRequest := bleve.NewSearchRequest(bleve.NewQueryStringQuery(keyword))
	searchRequest.Fields = []string{"*"}

	bleveResult, err := index.clientIndex.SearchInContext(ctx, searchRequest)
	if err != nil {
		err = fmt.Errorf("%s %v", prefixBleve, err)
		return
	}
	result = SearchResultCustom(*bleveResult)

	bytes, err := json.Marshal(result.GetDocumentFields())
	if err != nil {
		err = fmt.Errorf("%s %v", prefixBleve, err)
		return
	}
	if err = json.Unmarshal(bytes, dest); err != nil {
		err = fmt.Errorf("%s %v", prefixBleve, err)
		return
	}
	return
}

// TODO: utilize context
func (index *BleveIndex) BulkIndex(ctx context.Context, docs BleveDocs) (docErrors *BleveDocErrors) {
	errorChan := make(chan BleveDocError)
	wg := sync.WaitGroup{}
	for _, doc := range docs {
		wg.Add(1)
		go func(d BleveDoc) {
			logging.DebugContext(ctx, "indexing doc with ID %s", d.ID)

			if err := index.clientIndex.Index(d.ID, d.Data); err != nil {
				errorChan <- BleveDocError{
					DocID: d.ID,
					err:   err,
				}
			}
			wg.Done()
		}(doc)
	}
	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// constructing errors if exist
	for docError := range errorChan {
		if docErrors == nil {
			docErrors = &BleveDocErrors{}
		}
		*docErrors = append(*docErrors, docError)
	}
	return
}

func (index *BleveIndex) Close() error {
	if err := index.clientIndex.Close(); err != nil {
		return fmt.Errorf("%s %v", prefixBleve, err)
	}
	return nil
}
