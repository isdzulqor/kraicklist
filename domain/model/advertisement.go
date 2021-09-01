package model

import (
	"fmt"

	"github.com/isdzulqor/kraicklist/external/index"
)

type Advertisement struct {
	ID        int64       `json:"id"  db:"id"`
	Title     string      `json:"title" db:"title"`
	Content   string      `json:"content" db:"content"`
	ThumbURL  string      `json:"thumb_url" db:"thumb_url"`
	Tags      interface{} `json:"tags" db:"tags"` // TODO: revise to slices, needs to sanitize when indexing
	UpdatedAt int64       `json:"updated_at" db:"updated_at"`
	ImageURLs interface{} `json:"image_urls" db:"image_urls"` // TODO: revise to slices, needs to sanitize when indexing
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

func (ads Advertisements) ToSQLInsert() (query string, args []interface{}, err error) {
	if len(ads) == 0 {
		err = fmt.Errorf("no ads to be converted to elastic docs")
		return
	}

	query = `INSERT INTO advertisement (id, title, content, thumb_url, tags, updated_at, image_urls)`
	query = `INSERT INTO advertisement (id, title, content, title_tokens, content_tokens, thumb_url, tags, updated_at, image_urls)`

	var marks string

	markIncrement := 0
	for i, ad := range ads {
		if i > 0 {
			marks += ", \n"
		}
		args = append(args, ad.ID, ad.Title, ad.Content, ad.Title, ad.Content,
			ad.ThumbURL, fmt.Sprint(ad.Tags), ad.UpdatedAt, fmt.Sprint(ad.ImageURLs))

		marks += fmt.Sprintf("($%d, $%d, $%d, to_tsvector($%d), to_tsvector($%d), $%d, $%d, $%d, $%d)",
			markIncrement+1,
			markIncrement+2,
			markIncrement+3,
			markIncrement+4,
			markIncrement+5,
			markIncrement+6,
			markIncrement+7,
			markIncrement+8,
			markIncrement+9,
		)
		markIncrement += 9
	}
	query = fmt.Sprintf("%s VALUES %s", query, marks)
	return
}
