package domain

import (
	"encoding/json"
	"fmt"
)

// AdStatus represents the status of an advertisement
type AdStatus int

const (
	StatusDraft      AdStatus = 0 // Draft
	StatusPending    AdStatus = 1 // Awaiting review
	StatusFromParser AdStatus = 2 // From parser
	StatusActive     AdStatus = 3 // Active
	StatusCompleted  AdStatus = 4 // Completed
	StatusRejected   AdStatus = 5 // Rejected
	StatusApproved   AdStatus = 6 // Approved
	StatusUnknown    AdStatus = 7 // Robot didn't understand
	StatusDuplicate  AdStatus = 8 // Duplicate
)

// String returns the string representation of the status
func (s AdStatus) String() string {
	switch s {
	case StatusDraft:
		return "draft"
	case StatusPending:
		return "pending"
	case StatusFromParser:
		return "from_parser"
	case StatusActive:
		return "active"
	case StatusCompleted:
		return "completed"
	case StatusRejected:
		return "rejected"
	case StatusApproved:
		return "approved"
	case StatusUnknown:
		return "unknown"
	case StatusDuplicate:
		return "duplicate"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler
func (s AdStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(s))
}

// UnmarshalJSON implements json.Unmarshaler
func (s *AdStatus) UnmarshalJSON(data []byte) error {
	var status int
	if err := json.Unmarshal(data, &status); err != nil {
		return fmt.Errorf("invalid status: %v", err)
	}

	*s = AdStatus(status)
	return nil
}
