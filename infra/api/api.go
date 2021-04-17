package api

import (
	"context"
	"kraicklist/config"
	"kraicklist/domain/handler"
	"kraicklist/domain/repository"
	"kraicklist/domain/service"
	"kraicklist/external/index"
	"kraicklist/helper/logging"
	"kraicklist/infra/seed"
	"net/http"
	"strings"
)

func Exec() {
	conf := config.Get()
	conf.PrintPretty()

	// data seeding process
	seed.Exec()

	ctx := context.Background()

	logging.Init(strings.ToUpper(conf.LogLevel))

	// initilize dependencies
	bleveIndex, err := index.InitBleveIndex(ctx, conf.Advertisement.Bleve.IndexName)
	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}

	// initialize repo
	adRepo := repository.InitAdvertisement(conf, bleveIndex)

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
