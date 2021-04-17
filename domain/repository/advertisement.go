package repository

import (
	"context"
	"kraicklist/config"
	"kraicklist/domain/model"
	"kraicklist/external/index"
)

type Advertisement struct {
	conf *config.Config

	bleveIndex *index.BleveIndex
}

func InitAdvertisement(conf *config.Config, bleveIndex *index.BleveIndex) *Advertisement {
	return &Advertisement{
		conf:       conf,
		bleveIndex: bleveIndex,
	}
}

func (ad *Advertisement) SearchAds(ctx context.Context, query string) (out model.Advertisements, err error) {
	if _, err = ad.bleveIndex.SearchQuery(ctx, query, &out); err != nil {
		// TODO: error mapping
	}
	return
}
