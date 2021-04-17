package api

import (
	"context"
	"kraicklist/config"
	"kraicklist/domain/handler"
	"kraicklist/domain/repository"
	"kraicklist/domain/service"
	"kraicklist/external/index"
	"kraicklist/helper/logging"
	"net/http"
	"strings"
)

func Exec() {
	conf := config.Get()
	conf.PrintPretty()

	ctx := context.Background()

	logging.Init(strings.ToUpper(conf.LogLevel))

	var (
		err error

		bleveIndex   *index.BleveIndex
		elasticIndex *index.ElasticIndex
	)

	// initialize dependencies
	switch conf.Advertisement.Indexer {
	case index.IndexBleve:
		bleveIndex, err = index.InitBleveIndex(ctx, conf.Advertisement.Bleve.IndexName)
		if err != nil {
			logging.FatalContext(ctx, "%v", err)
		}
	case index.IndexElastic:
		elasticIndex, err = index.InitESIndex(ctx,
			conf.Advertisement.Elastic.Host,
			conf.Advertisement.Elastic.Username,
			conf.Advertisement.Elastic.Password,
			conf.Advertisement.Elastic.IndexName)
		if err != nil {
			logging.FatalContext(ctx, "%v", err)
		}
	default:
		logging.FatalContext(ctx, "Indexer for %s is invalid", conf.Advertisement.Indexer)
	}
	// initialize repo
	adRepo := repository.InitAdvertisement(conf, bleveIndex, elasticIndex)

	// initialize service
	adService := service.InitAdvertisement(adRepo)

	// initialize handlers
	adHandler := handler.InitAdvertisement(conf, adService)
	rootHandler := handler.Root{
		Advertisement: adHandler,
	}

	// starting server
	logging.InfoContext(ctx, "Starting HTTP on port %s", conf.Port)
	router := createRouter(ctx, rootHandler)
	if err := http.ListenAndServe(":"+conf.Port, router); err != nil {
		logging.FatalContext(ctx, "Failed starting HTTP - %v", err)
	}
}
