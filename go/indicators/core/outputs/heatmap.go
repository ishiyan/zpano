package outputs

import (
	"fmt"
	"math"
	"time"
)

// Heatmap holds a time stamp (x) and an array of values (z) corresponding to parameter (y) range
// to paint a heatmap column.
type Heatmap struct {
	// Time is the date and time (x) of this heatmap.
	Time time.Time `json:"time"`

	// ParameterFirst is the first parameter (y) value of the heatmap. This value is the same for all heatmap columns.
	// A parameter corresponding to the i-th value can be calculated as:
	//
	// min(ParameterFirst,ParameterLast) + i / ParameterResolution
	ParameterFirst float64 `json:"first"`

	// ParameterLast is the last parameter (y) value of the heatmap. This value is the same for all heatmap columns.
	// A parameter corresponding to the i-th value can be calculated as:
	//
	// min(ParameterFirst,ParameterLast) + i / ParameterResolution
	ParameterLast float64 `json:"last"`

	// ParameterResolution is the resolution of the parameter (y).  This value is the same for all heatmap columns.
	// It is always a positive number.
	// A value of 10 means that heatmap values are evaluated at every 1/10 of the parameter range.
	// A parameter corresponding to the i-th value can be calculated as:
	//
	// min(ParameterFirst,ParameterLast) + i / ParameterResolution
	ParameterResolution float64 `json:"res"`

	// ValueMin is a minimal value (z) of this heatmap column.
	ValueMin float64 `json:"min"`

	// ValueMax is a maximal value (z) of this heatmap column.
	ValueMax float64 `json:"max"`

	// Values is a slice of values (z) of this heatmap column.
	// The length of the slice is the same for all heatmap columns, but may be zero if the heatmap column is empty.
	Values []float64 `json:"values"`
}

// NewHeatmap creates a new heatmap.
func NewHeatmap(time time.Time, parameterFirst, parameterLast, parameterResolution, valueMin, valueMax float64,
	values []float64,
) *Heatmap {
	return &Heatmap{
		Time:                time,
		ParameterFirst:      parameterFirst,
		ParameterLast:       parameterLast,
		ParameterResolution: parameterResolution,
		ValueMin:            valueMin,
		ValueMax:            valueMax,
		Values:              values,
	}
}

// NewEmptyHeatmap creates a new empty heatmap.
// Both min and max values will be equal to NaN, the values slice will be empty.
func NewEmptyHeatmap(time time.Time, parameterFirst, parameterLast, parameterResolution float64) *Heatmap {
	nan := math.NaN()

	return &Heatmap{
		Time:                time,
		ParameterFirst:      parameterFirst,
		ParameterLast:       parameterLast,
		ParameterResolution: parameterResolution,
		ValueMin:            nan,
		ValueMax:            nan,
		Values:              []float64{},
	}
}

// IsEmpty indicates whether this heatmap is not initialized.
func (h *Heatmap) IsEmpty() bool {
	return len(h.Values) < 1
}

// String implements the Stringer interface.
func (h *Heatmap) String() string {
	return fmt.Sprintf("{%s, (%f, %f, %f), (%f, %f), %v}",
		h.Time.Format(timeFmt), h.ParameterFirst, h.ParameterLast, h.ParameterResolution, h.ValueMin, h.ValueMax, h.Values)
}
