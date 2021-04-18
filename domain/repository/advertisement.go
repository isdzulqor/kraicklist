package repository

import (
	"context"
	"fmt"

	"github.com/isdzulqor/kraicklist/config"
	"github.com/isdzulqor/kraicklist/domain/model"
	"github.com/isdzulqor/kraicklist/external/index"
)

type Advertisement struct {
	conf *config.Config

	bleveIndex *index.BleveIndex
	esIndex    *index.ElasticIndex
}

func InitAdvertisement(conf *config.Config, bleveIndex *index.BleveIndex, esIndex *index.ElasticIndex) *Advertisement {
	return &Advertisement{
		conf:       conf,
		bleveIndex: bleveIndex,
		esIndex:    esIndex,
	}
}

func (ad *Advertisement) SearchAds(ctx context.Context, query string) (out model.Advertisements, err error) {
	if ad.conf.IndexerActivated == index.IndexElastic {
		esQuery := index.ElasticRootQuery{}
		esQuery.ConstructElasticMultiMatchQuery(query, "title", "content", "tags")
		if _, err = ad.esIndex.SearchQuery(ctx, esQuery, &out); err != nil {
			// TODO: error handler
		}
		return
	}

	if _, err = ad.bleveIndex.SearchQuery(ctx, query, &out); err != nil {
		// TODO: error handler
	}
	return
}

func (ad *Advertisement) IndexAds(ctx context.Context, in model.Advertisements) (err error) {
	var (
		elasticDocs index.ElasticDocs
		bleveDocs   index.BleveDocs
	)
	// indexing using elastic
	if ad.conf.IndexerActivated == index.IndexElastic {
		if elasticDocs, err = in.ToElasticDocs(); err != nil {
			err = fmt.Errorf("failed to convert to elasticDocs, err:%v", err)
			return
		}
		// TODO: error handler
		var errorElasticDocs *index.ElasticDocErrors
		errorElasticDocs, err = ad.esIndex.BulkIndexDocs(ctx, elasticDocs)
		if err != nil {
			return
		}
		if errorElasticDocs != nil {
			err = errorElasticDocs.ToError()
			return
		}
		return
	}

	// using bleve
	if bleveDocs, err = in.ToBleveDocs(); err != nil {
		err = fmt.Errorf("failed to convert to bleveDocs, err:%v", err)
		return
	}
	if errorDocs := ad.bleveIndex.BulkIndex(ctx, bleveDocs); errorDocs != nil {
		err = errorDocs.ToError()
		return
	}
	return
}
