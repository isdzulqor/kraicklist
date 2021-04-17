package service

import (
	"context"
	"kraicklist/domain/model"
	"kraicklist/domain/repository"
)

type Advertisement struct {
	adRepo *repository.Advertisement
}

func InitAdvertisement(adRepo *repository.Advertisement) *Advertisement {
	return &Advertisement{
		adRepo: adRepo,
	}
}

func (s *Advertisement) SearchAds(ctx context.Context, keyword string) (out model.Advertisements, err error) {
	return s.adRepo.SearchAds(ctx, keyword)
}

func (s *Advertisement) IndexAds(ctx context.Context, in model.Advertisements) (err error) {
	return s.adRepo.IndexAds(ctx, in)
}
