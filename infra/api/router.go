package api

import (
	"context"
	"kraicklist/domain/handler"
	"net/http"

	"github.com/gorilla/mux"
)

func createRouter(ctx context.Context, rootHandler handler.Root) http.Handler {
	router := mux.NewRouter()

	// UI static
	fs := http.FileServer(http.Dir("./static"))
	router.Handle("/", fs)

	// API serve
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/advertisement/search", rootHandler.Advertisement.SearchAds).Methods("GET")
	return router
}
