package domain

import (
	"database/sql/driver"
	"encoding/json"
)

// Property represents a property definition
type Property struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Name         string `json:"name"`
	Type         string `json:"type"`       // primitive, reference
	ValueType    string `json:"value_type"` // string, number, boolean
	IsSearchable bool   `json:"is_searchable"`
}

// PropertyValue represents a predefined value for a property
type PropertyValue struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	PropertyID uint   `json:"property_id"`
	Value      string `json:"value"`
}

// AdProperty represents a property value for an ad
type AdProperty struct {
	ID      uint   `json:"ID"`
	Value   string `json:"value,omitempty"`
	ValueID *uint  `json:"value_id,omitempty"`
}

// AdProperties represents a collection of ad properties
type AdProperties []AdProperty

// Value implements the driver.Valuer interface for JSONB storage
func (p AdProperties) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements the sql.Scanner interface for JSONB storage
func (p *AdProperties) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, &p)
}

// PropertyFilter represents a filter by property value
type PropertyFilter struct {
	PropertyID uint     `json:"property_id"`
	Values     []string `json:"values,omitempty"`
	ValueIDs   []uint   `json:"value_ids,omitempty"`
}
