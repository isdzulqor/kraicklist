package handler

import (
	"encoding/json"
	"net/http"

	"github.com/isdzulqor/kraicklist/config"
	"github.com/isdzulqor/kraicklist/domain/model"
	"github.com/isdzulqor/kraicklist/domain/service"
	"github.com/isdzulqor/kraicklist/helper/errors"
	"github.com/isdzulqor/kraicklist/helper/logging"
	"github.com/isdzulqor/kraicklist/helper/response"
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
		err := errors.ErrorParamInvalid.AppendMessage("q param is necessary.")
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

func (h *Advertisement) IndexAds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var requestData model.Advertisements

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		logging.DebugContext(ctx, "failed to decode body param err: %v", err)
		err = errors.ErrorParamInvalid
		response.Failed(ctx, w, errors.GetStatusCode(err), err)
		return
	}

	if err := h.adService.IndexAds(ctx, requestData); err != nil {
		response.Failed(ctx, w, errors.GetStatusCode(err), err)
		return
	}

	response.Success(ctx, w, http.StatusOK, "success")
}
