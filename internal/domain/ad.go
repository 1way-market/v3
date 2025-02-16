package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// MultiLangText represents text in a specific language
type MultiLangText struct {
	Lang int    `json:"lang"`
	Text string `json:"text"`
}

// MultiLangArray represents an array of multilingual texts
type MultiLangArray []MultiLangText

// Value implements the driver.Valuer interface for JSONB storage
func (m MultiLangArray) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for JSONB storage
func (m *MultiLangArray) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, &m)
}

// Attributes represents dynamic properties of an ad
type Attributes map[string]interface{}

// Value implements the driver.Valuer interface
func (a Attributes) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface
func (a *Attributes) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, &a)
}

// Ad represents the main advertisement entity
type Ad struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Title        MultiLangArray `json:"title_multi" gorm:"type:jsonb;not null;column:title"`
	Description  MultiLangArray `json:"body_multi,omitempty" gorm:"type:jsonb;column:description"`
	Attributes   Attributes     `json:"attributes,omitempty" gorm:"type:jsonb"`
	CategoryIDs  []int          `json:"category_ids,omitempty" gorm:"type:integer[]"`
	Status       AdStatus       `json:"status" gorm:"type:integer;index;default:0"`
	Price        *Price         `json:"price,omitempty" gorm:"type:jsonb"`
	SearchVector string         `json:"-" gorm:"type:tsvector"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// GetText returns the text for the specified language, falling back to English if not found
func (m MultiLangArray) GetText(lang int) string {
	// First try to find exact match
	for _, t := range m {
		if t.Lang == lang {
			return t.Text
		}
	}

	// Fallback to English (lang 2)
	for _, t := range m {
		if t.Lang == 2 {
			return t.Text
		}
	}

	// If no English, return the first available text
	if len(m) > 0 {
		return m[0].Text
	}

	return ""
}

// FilterRequest represents the query parameters for ad filtering
type FilterRequest struct {
	CategoryIDs []int               `form:"categories"`
	Properties  map[string][]string `form:"properties"`
	TextSearch  string              `form:"q"`
	SortBy      string              `form:"sort"`
	PageToken   string              `form:"next_page"`
	PageSize    int                 `form:"page_size"`
	Lang        string              `form:"lang" binding:"required"`
	MinPrice    *float64            `form:"min_price"`
	MaxPrice    *float64            `form:"max_price"`
	Currency    string              `form:"currency"`
	Status      *AdStatus           `form:"status"`
}

// PaginatedResponse represents a paginated list of ads
type PaginatedResponse struct {
	Items      []Ad   `json:"items"`
	NextPage   string `json:"next_page,omitempty"`
	TotalCount int64  `json:"total_count"`
}
