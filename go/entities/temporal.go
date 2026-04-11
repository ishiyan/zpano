package data

import "time"

type Temporal interface {
	DateTime() time.Time
}
