package repository

import (
	"context"
	"kraicklist/config"
	"kraicklist/domain/model"
	"strings"
)

type Advertisement struct {
	conf *config.Config

	adsData *model.Advertisements
}

func InitAdvertisement(conf *config.Config, adsData *model.Advertisements) *Advertisement {
	return &Advertisement{
		conf:    conf,
		adsData: adsData,
	}
}

func (ad *Advertisement) SearchAds(ctx context.Context, query string) (out model.Advertisements, err error) {
	for _, ad := range *ad.adsData {
		if strings.Contains(ad.Title, query) || strings.Contains(ad.Content, query) {
			out = append(out, ad)
		}
	}
	return
}
