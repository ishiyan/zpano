package entities

import "time"

// Scalar represents a scalar value.
type Scalar struct {
	Time  time.Time `json:"t"` // The date and time.
	Value float64   `json:"v"` // The value.
}
