package model

import (
	"fmt"
	"kraicklist/external/index"
)

type Advertisement struct {
	ID        int64       `json:"id"`
	Title     string      `json:"title"`
	Content   string      `json:"content"`
	ThumbURL  string      `json:"thumb_url"`
	Tags      interface{} `json:"tags"` // TODO: revise to slices, needs to sanitize when indexing
	UpdatedAt int64       `json:"updated_at"`
	ImageURLs interface{} `json:"image_urls"` // TODO: revise to slices, needs to sanitize when indexing
}

type Advertisements []Advertisement

func (ads Advertisements) ToBleveDocs() (out index.BleveDocs, err error) {
	if len(ads) == 0 {
		err = fmt.Errorf("no ads to be converted to bleve docs")
		return
	}
	for _, ad := range ads {
		out = append(out, index.BleveDoc{
			ID:   fmt.Sprint(ad.ID),
			Data: ad,
		})
	}
	return
}

func (ads Advertisements) ToElasticDocs() (out index.ElasticDocs, err error) {
	if len(ads) == 0 {
		err = fmt.Errorf("no ads to be converted to elastic docs")
		return
	}
	for _, ad := range ads {
		out = append(out, index.ElasticDoc{
			ID:   fmt.Sprint(ad.ID),
			Data: ad,
		})
	}
	return
}
