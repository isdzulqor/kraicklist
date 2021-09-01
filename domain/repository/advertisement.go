package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/isdzulqor/kraicklist/config"
	"github.com/isdzulqor/kraicklist/domain/model"
	"github.com/isdzulqor/kraicklist/external/index"
	errorLib "github.com/isdzulqor/kraicklist/helper/errors"
	"github.com/isdzulqor/kraicklist/helper/logging"
	"github.com/jmoiron/sqlx"
)

type Advertisement struct {
	conf *config.Config

	bleveIndex *index.BleveIndex
	esIndex    *index.ElasticIndex

	sqlDB *sqlx.DB `inject:""`
}

func InitAdvertisement(conf *config.Config,
	bleveIndex *index.BleveIndex,
	esIndex *index.ElasticIndex, sqlDB *sqlx.DB) *Advertisement {
	return &Advertisement{
		conf:       conf,
		bleveIndex: bleveIndex,
		esIndex:    esIndex,
		sqlDB:      sqlDB,
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

	if ad.conf.IndexerActivated == index.IndexBleve {
		if _, err = ad.bleveIndex.SearchQuery(ctx, query, &out); err != nil {
			// TODO: error handler
		}
	}

	// SQL Default
	err = ad.searchAdsSQL(ctx, query, &out)
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

	if ad.conf.IndexerActivated == index.IndexBleve {
		// using bleve
		if bleveDocs, err = in.ToBleveDocs(); err != nil {
			err = fmt.Errorf("failed to convert to bleveDocs, err:%v", err)
			return
		}
		if errorDocs := ad.bleveIndex.BulkIndex(ctx, bleveDocs); errorDocs != nil {
			err = errorDocs.ToError()
			return
		}
	}
	return
}

func (ad *Advertisement) searchAdsSQL(ctx context.Context, keyword string, dest interface{}) (err error) {
	seperatedKey := strings.Split(keyword, " ")
	keyword = ""
	for i, key := range seperatedKey {
		if i > 0 {
			keyword += " & "
		}
		keyword += key
	}

	query := `
		SELECT 
			a.id,
			a.title,
			a.content,
			a.thumb_url,
			a.tags,
			a.updated_at,
			a.image_urls
		FROM 
			advertisement a
		WHERE 
			a.content_tokens @@ to_tsquery($1) 
			OR a.title_tokens @@ to_tsquery($1)
	`

	args := []interface{}{
		keyword,
	}

	if err != nil {
		logging.FatalContext(ctx, "%v", err)
	}
	if err = ad.sqlDB.Unsafe().SelectContext(ctx, dest, query, args...); err != nil {
		logging.DebugContext(ctx, errorLib.FormatQueryError(query, args...))
		err = fmt.Errorf("failed to sqlDB.GetContext due: %v", err)
	}
	return
}
