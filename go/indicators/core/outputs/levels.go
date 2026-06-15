package outputs

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Level is a single entry of a Levels output, expressed as a value with an
// optional bar offset and an optional strength.
type Level struct {
	// Value is the value (e.g. a price level or a multiplier) at this entry.
	Value float64 `json:"value"`

	// Offset is the number of bars back from the Levels' Time at which this
	// level occurs (0 = the current bar). For levels that are not anchored to a
	// past bar (e.g. theoretical price levels), Offset is 0.
	Offset int `json:"offset"`

	// Strength is an optional significance measure for this level (higher = more
	// significant). It is NaN when not applicable.
	Strength float64 `json:"strength"`
}

// Levels holds a time stamp (anchoring the current bar) and a variable-length
// set of levels, typically a ranked set of support/resistance price levels.
//
// Each Update emits a fresh, self-contained Levels; renderers should replace the
// previous set of this indicator with the new one. This provides an immutable,
// streaming-friendly model for indicators whose level set may change as new bars
// arrive (e.g. support/resistance, pivots, Fibonacci grids, quantum price levels).
type Levels struct {
	// Time is the date and time (x) of the bar that anchors this set
	// (i.e. the bar at offset 0).
	Time time.Time `json:"time"`

	// Levels is the set of levels. The slice may be empty if the indicator has
	// not produced any levels yet.
	Levels []Level `json:"levels"`
}

// NewLevel creates a level with the given value, offset and strength.
func NewLevel(value float64, offset int, strength float64) Level {
	return Level{
		Value:    value,
		Offset:   offset,
		Strength: strength,
	}
}

// NewValueLevel creates a level with the given value, an offset of 0 and a NaN
// strength. It is a convenience for levels that carry only a value (e.g. a
// theoretical price level).
func NewValueLevel(value float64) Level {
	return Level{
		Value:    value,
		Offset:   0,
		Strength: math.NaN(),
	}
}

// NewLevels creates a new Levels with the given time and entries.
func NewLevels(time time.Time, levels []Level) *Levels {
	return &Levels{
		Time:   time,
		Levels: levels,
	}
}

// NewEmptyLevels creates a new empty Levels with no entries.
func NewEmptyLevels(time time.Time) *Levels {
	return &Levels{
		Time:   time,
		Levels: []Level{},
	}
}

// IsEmpty indicates whether this Levels has no entries.
func (l *Levels) IsEmpty() bool {
	return len(l.Levels) == 0
}

// String implements the Stringer interface.
func (l *Levels) String() string {
	var sb strings.Builder

	sb.WriteString("{")
	sb.WriteString(l.Time.Format(timeFmt))
	sb.WriteString(", [")

	for i, lv := range l.Levels {
		if i > 0 {
			sb.WriteString(" ")
		}

		fmt.Fprintf(&sb, "(%f, %d, %f)", lv.Value, lv.Offset, lv.Strength)
	}

	sb.WriteString("]}")

	return sb.String()
}
