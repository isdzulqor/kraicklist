package seed

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"kraicklist/config"
	"kraicklist/domain/model"
	"kraicklist/helper/logging"
	"os"
	"strings"
)

func Exec() model.Advertisements {
	conf := config.Get()

	ctx := context.Background()

	logging.Init(strings.ToUpper(conf.LogLevel))

	logging.InfoContext(ctx, "preparing data seed...")

	ads, err := loadAdsData(conf.Advertisement.MasterDataPath)
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}

	logging.InfoContext(ctx, "data seed is finished")
	return ads
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
