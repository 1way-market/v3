package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Currency codes according to ISO 4217
const (
	CurrencyUSD = "840" // United States Dollar
	CurrencyEUR = "978" // Euro
	CurrencyTRY = "949" // Turkish Lira
	CurrencyRUB = "643" // Russian Ruble
	CurrencyGBP = "826" // British Pound
)

// Price represents a monetary value with its currency
type Price struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

// UnmarshalJSON implements custom JSON unmarshaling to handle currency as both string and number
func (p *Price) UnmarshalJSON(data []byte) error {
	// Try to unmarshal into a temporary struct
	var temp struct {
		Value    float64     `json:"value"`
		Currency json.Number `json:"currency"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	p.Value = temp.Value

	// Convert currency to string
	if temp.Currency != "" {
		if _, err := strconv.Atoi(string(temp.Currency)); err != nil {
			return fmt.Errorf("invalid currency code: %v", temp.Currency)
		}
		p.Currency = string(temp.Currency)
	}

	return nil
}

// Scan implements the sql.Scanner interface for JSONB storage
func (p *Price) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, &p)
}
