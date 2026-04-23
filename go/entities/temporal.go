package entities

import "time"

// Temporal describes a value that carries a timestamp.
type Temporal interface {
	DateTime() time.Time
}
