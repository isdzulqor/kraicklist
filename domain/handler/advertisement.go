package handler

import (
	"kraicklist/config"
	"kraicklist/domain/service"
	"kraicklist/helper/errors"
	"kraicklist/helper/response"
	"net/http"
)

type Advertisement struct {
	conf *config.Config

	adService *service.Advertisement
}

func InitAdvertisement(conf *config.Config, adService *service.Advertisement) *Advertisement {
	return &Advertisement{
		conf:      conf,
		adService: adService,
	}
}

func (h *Advertisement) SearchAds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	keyword := r.FormValue("q")
	if keyword == "" {
		err := errors.ErrorParamInvalid.AppendMessage("q is necessary.")
		response.Failed(ctx, w, errors.GetStatusCode(err), err)
		return
	}

	result, err := h.adService.SearchAds(ctx, keyword)
	if err != nil {
		response.Failed(ctx, w, errors.GetStatusCode(err), err)
		return
	}

	response.Success(ctx, w, http.StatusOK, result)
}
