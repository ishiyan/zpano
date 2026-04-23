package outputs

import (
	"fmt"
	"strings"
	"time"
)

// Point is a single vertex of a Polyline, expressed as (offset, value)
// where offset is the number of bars back from the Polyline's Time
// (0 = the current bar, 1 = the previous bar, and so on).
type Point struct {
	// Offset is the number of bars back from the Polyline's Time.
	Offset int `json:"offset"`

	// Value is the value (y) at this vertex.
	Value float64 `json:"value"`
}

// Polyline holds a time stamp (anchoring the current bar) and an ordered,
// variable-length sequence of points describing a polyline over recent history.
//
// Points are ordered from oldest (largest Offset) to newest (Offset == 0).
//
// Each Update emits a fresh, self-contained Polyline; renderers should
// replace the previous polyline of this indicator with the new one.
// This provides an immutable, streaming-friendly model for indicators
// whose historical overlay may change as new bars arrive (e.g. ZigZag,
// Fibonacci grids, pivot overlays).
type Polyline struct {
	// Time is the date and time (x) of the bar that anchors this polyline
	// (i.e. the bar at offset 0).
	Time time.Time `json:"time"`

	// Points is the ordered sequence of polyline vertices, from oldest to newest.
	// The slice may be empty if the indicator has not produced a polyline yet.
	Points []Point `json:"points"`
}

// NewPolyline creates a new polyline with the given time and points.
// The points are stored as-is; callers are responsible for supplying them
// in the documented old-to-new order.
func NewPolyline(time time.Time, points []Point) *Polyline {
	return &Polyline{
		Time:   time,
		Points: points,
	}
}

// NewEmptyPolyline creates a new empty polyline with no points.
func NewEmptyPolyline(time time.Time) *Polyline {
	return &Polyline{
		Time:   time,
		Points: []Point{},
	}
}

// IsEmpty indicates whether this polyline has no points.
func (p *Polyline) IsEmpty() bool {
	return len(p.Points) == 0
}

// String implements the Stringer interface.
func (p *Polyline) String() string {
	var sb strings.Builder

	sb.WriteString("{")
	sb.WriteString(p.Time.Format(timeFmt))
	sb.WriteString(", [")

	for i, pt := range p.Points {
		if i > 0 {
			sb.WriteString(" ")
		}

		fmt.Fprintf(&sb, "(%d, %f)", pt.Offset, pt.Value)
	}

	sb.WriteString("]}")

	return sb.String()
}
