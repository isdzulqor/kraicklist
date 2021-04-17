package seed

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"kraicklist/config"
	"kraicklist/domain/model"
	"kraicklist/external/index"
	"kraicklist/helper/logging"
	"os"
	"strings"
)

func Exec() {
	conf := config.Get()

	ctx := context.Background()

	logging.Init(strings.ToUpper(conf.LogLevel))

	logging.InfoContext(ctx, "preparing data seed...")

	ads, err := loadAdsData(conf.Advertisement.MasterDataPath)
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}

	seedDataWithBleve(ctx, ads, conf)

	logging.InfoContext(ctx, "data seed is finished")
	return
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
