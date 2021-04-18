package api

import (
	"context"
	"net/http"

	"github.com/isdzulqor/kraicklist/domain/handler"
	"github.com/isdzulqor/kraicklist/infra"

	"github.com/gorilla/mux"
)

func createRouter(ctx context.Context, rootHandler handler.Root) http.Handler {
	router := mux.NewRouter()

	// setup middlewares
	router.Use(infra.LoggingHandler)
	router.Use(infra.RecoverHandler)
	router.Use(infra.CheckShuttingDown(*rootHandler.Health))

	// UI static
	fs := http.FileServer(http.Dir("./static"))
	router.Handle("/", fs)

	// healthcheck endpoint
	router.HandleFunc("/health", rootHandler.Health.GetHealth).Methods("GET")

	// API serve
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/advertisement/search", rootHandler.Advertisement.SearchAds).Methods("GET")
	api.HandleFunc("/advertisement/index", rootHandler.Advertisement.IndexAds).Methods("POST")
	return router
}
