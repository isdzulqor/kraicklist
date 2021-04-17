package api

import (
	"context"
	"kraicklist/config"
	"kraicklist/domain/handler"
	"kraicklist/domain/repository"
	"kraicklist/domain/service"
	"kraicklist/helper/logging"
	"kraicklist/infra/seed"
	"net/http"
	"strings"
)

func Exec() {
	conf := config.Get()
	conf.PrintPretty()

	ctx := context.Background()

	logging.Init(strings.ToUpper(conf.LogLevel))

	// data seeding
	adsData := seed.Exec()

	// initialize repo
	adRepo := repository.InitAdvertisement(conf, &adsData)

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
