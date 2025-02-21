package model

import "time"

// Ad represents an advertisement in the system
type Ad struct {
	ID           int64           `json:"id"`
	Title        []MultiLangText `json:"title_multi"`
	Description  []MultiLangText `json:"body_multi,omitempty"`
	Properties   []MultiLangText `json:"properties,omitempty"`
	CategoryIDs  []int64         `json:"category_ids,omitempty"`
	Status       string          `json:"status"`
	Price        *float64        `json:"price,omitempty"`
	SearchVector string          `json:"-"` // Used internally for full-text search
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// AdFilter represents the filter options for ads
type AdFilter struct {
	CategoryIDs []int64  `json:"category_ids,omitempty"`
	Status      string   `json:"status,omitempty"`
	MinPrice    *float64 `json:"min_price,omitempty"`
	MaxPrice    *float64 `json:"max_price,omitempty"`
	SearchQuery string   `json:"search_query,omitempty"`
	Language    Language `json:"language,omitempty"`
	Offset      int      `json:"offset,omitempty"`
	Limit       int      `json:"limit,omitempty"`
}

// AdRepository defines the interface for ad storage operations
type AdRepository interface {
	Create(ad *Ad) error
	Update(ad *Ad) error
	Delete(id int64) error
	GetByID(id int64) (*Ad, error)
	List(filter *AdFilter) ([]*Ad, error)
	Count(filter *AdFilter) (int64, error)
}
