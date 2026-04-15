package outputs

import (
	"fmt"
	"math"
	"time"
)

// Band represents two band values and a time stamp.
type Band struct {
	// Time is the date and time of this band.
	Time time.Time `json:"time"`

	// Lower is the lower value of the band.
	Lower float64 `json:"lower"`

	// Upper is the upper value of the band.
	Upper float64 `json:"upper"`
}

// newBand creates a new band.
// Both lower and upper values should not be NaN.
func NewBand(time time.Time, lower, upper float64) *Band {
	if lower < upper {
		return &Band{
			Time:  time,
			Lower: lower,
			Upper: upper,
		}
	}

	return &Band{
		Time:  time,
		Lower: upper,
		Upper: lower,
	}
}

// newEmptyBand creates a new empty band.
// Both lower and upper values will be equal to NaN.
func NewEmptyBand(time time.Time) *Band {
	nan := math.NaN()

	return &Band{
		Time:  time,
		Lower: nan,
		Upper: nan,
	}
}

// IsEmpty indicates whether this band is not initialized.
func (b *Band) IsEmpty() bool {
	return math.IsNaN(b.Lower) || math.IsNaN(b.Upper)
}

// String implements the Stringer interface.
func (b *Band) String() string {
	return fmt.Sprintf("{%s, %f, %f}", b.Time.Format(timeFmt), b.Lower, b.Upper)
}
