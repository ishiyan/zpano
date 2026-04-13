package entities

import "time"

type Temporal interface {
	DateTime() time.Time
}
