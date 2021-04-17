package model

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
