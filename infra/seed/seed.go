package seed

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/isdzulqor/kraicklist/config"
	"github.com/isdzulqor/kraicklist/domain/model"
	"github.com/isdzulqor/kraicklist/external/index"
	"github.com/isdzulqor/kraicklist/helper/logging"
)

func Exec() {
	conf := config.Get()
	conf.PrintPretty()

	ctx := context.Background()

	logging.Init(strings.ToUpper(conf.LogLevel))

	logging.InfoContext(ctx, "preparing data seed...")

	ads, err := loadAdsData(conf.Advertisement.MasterDataPath)
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}

	switch conf.IndexerActivated {
	case index.IndexBleve:
		seedDataWithBleve(ctx, ads, conf)
	case index.IndexElastic:
		seedDataWithElastic(ctx, ads, conf)
	default:
		logging.FatalContext(ctx, "Indexer for %s is invalid", conf.IndexerActivated)
	}

	logging.InfoContext(ctx, "data seed is finished")
}

func loadAdsData(filePath string) (out model.Advertisements, err error) {
	// open file
	file, err := os.Open(filePath)
	if err != nil {
		err = fmt.Errorf("unable to open file due: %v", err)
		return
	}

	defer file.Close()
	// read as gzip
	reader, err := gzip.NewReader(file)
	if err != nil {
		err = fmt.Errorf("unable to initialize gzip reader due: %v", err)
		return
	}

	// read the reader using scanner to contstruct records
	cs := bufio.NewScanner(reader)
	for cs.Scan() {
		ad := model.Advertisement{}
		err = json.Unmarshal(cs.Bytes(), &ad)
		if err != nil {
			continue
		}
		out = append(out, ad)
	}
	return
}

func seedDataWithBleve(ctx context.Context, ads model.Advertisements, conf *config.Config) {
	logging.InfoContext(ctx, "data seeding with bleve index...")

	bleveIndex, err := index.InitBleveIndex(ctx, conf.Advertisement.Bleve.IndexName)
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}

	docs, err := ads.ToBleveDocs()
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
		return
	}

	if docErrors := bleveIndex.BulkIndex(ctx, docs); docErrors != nil {
		logging.ErrContext(ctx, "%v", docErrors.ToError())
	}

	if err := bleveIndex.Close(); err != nil {
		logging.ErrContext(ctx, "%v", err)
	}
}

func seedDataWithElastic(ctx context.Context, ads model.Advertisements, conf *config.Config) {
	logging.InfoContext(ctx, "data seeding with elastic index...")

	esIndex, err := index.InitESIndex(ctx,
		conf.Elastic.Host,
		conf.Elastic.Username,
		conf.Elastic.Password,
		conf.Advertisement.Elastic.IndexName)
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}

	// check es7 cluster readiness with retry
	if err = esIndex.PingWithRetry(conf.Elastic.PingRetry,
		conf.Elastic.PingWaitTime); err != nil {
		logging.WarnContext(ctx, "%v", err)
	}

	if err = esIndex.DeleteIndex(ctx); err != nil {
		logging.WarnContext(ctx, "%v", err)
	}

	docs, err := ads.ToElasticDocs()
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
		return
	}

	docErrors, err := esIndex.BulkIndexDocs(ctx, docs)
	if err != nil {
		logging.ErrContext(ctx, "%v", err)
	}
	if docErrors != nil {
		logging.ErrContext(ctx, "%v", docErrors.ToError())
	}
}
