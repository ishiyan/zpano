package data

import (
	"fmt"
	"time"
)

// Bar represents an [open, high, low, close, volume] price bar.
type Bar struct {
	// Time is the date and time of the closing price.
	Time time.Time `json:"time"`

	// Open is the opening price.
	Open float64 `json:"open"`

	// High is the highest price.
	High float64 `json:"high"`

	// Low is the lowest price.
	Low float64 `json:"low"`

	// Close is the closing price.
	Close float64 `json:"close"`

	// Volume is the aggregated volume.
	Volume float64 `json:"volume"`
}

// IsRising indicates whether this is a rising bar, i.e. the opening price is less than the closing price.
func (b *Bar) IsRising() bool {
	return b.Open < b.Close
}

// IsFalling indicates whether this is a falling bar, i.e. the closing price is less than the opening price.
func (b *Bar) IsFalling() bool {
	return b.Close < b.Open
}

// Median is the median price, calculated as
//   (low + high) / 2.
func (b *Bar) Median() float64 {
	return (b.Low + b.High) / 2 //nolint:gomnd
}

// Typical is the typical price, calculated as
//   (low + high + close) / 3.
func (b *Bar) Typical() float64 {
	return (b.Low + b.High + b.Close) / 3 //nolint:gomnd
}

// Weighted is the weighted price, calculated as
//   (low + high + 2*close) / 4.
func (b *Bar) Weighted() float64 {
	return (b.Low + b.High + b.Close + b.Close) / 4 //nolint:gomnd
}

// Average is the weighted price, calculated as
//   (low + high + open + close) / 4.
func (b *Bar) Average() float64 {
	return (b.Low + b.High + b.Open + b.Close) / 4 //nolint:gomnd
}

// String implements the Stringer interface.
func (b *Bar) String() string {
	return fmt.Sprintf("{%s, %f, %f, %f, %f, %f}", b.Time.Format(timeFmtDateTime), b.Open, b.High, b.Low, b.Close, b.Volume)
}
