// Package shape enumerates the data shapes an indicator output can take.
package shape

import (
	"bytes"
	"errors"
	"fmt"
)

// Shape identifies the data shape of an indicator output.
type Shape int

const (
	// Scalar holds a time stamp and a value.
	Scalar Shape = iota + 1

	// Band holds a time stamp and two values representing upper and lower lines of a band.
	Band

	// Heatmap holds a time stamp and an array of values representing a heat-map column.
	Heatmap

	// Polyline holds a time stamp and an ordered, variable-length sequence of (offset, value) points.
	Polyline
	last
)

const (
	unknown  = "unknown"
	scalar   = "scalar"
	band     = "band"
	heatmap  = "heatmap"
	polyline = "polyline"
)

var errUnknownShape = errors.New("unknown indicator output shape")

// String implements the Stringer interface.
func (s Shape) String() string {
	switch s {
	case Scalar:
		return scalar
	case Band:
		return band
	case Heatmap:
		return heatmap
	case Polyline:
		return polyline
	default:
		return unknown
	}
}

// IsKnown determines if this output shape is known.
func (s Shape) IsKnown() bool {
	return s >= Scalar && s < last
}

// MarshalJSON implements the Marshaler interface.
func (s Shape) MarshalJSON() ([]byte, error) {
	str := s.String()
	if str == unknown {
		return nil, fmt.Errorf("cannot marshal '%s': %w", str, errUnknownShape)
	}

	const extra = 2 // Two bytes for quotes.

	b := make([]byte, 0, len(str)+extra)
	b = append(b, '"')
	b = append(b, str...)
	b = append(b, '"')

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (s *Shape) UnmarshalJSON(data []byte) error {
	d := bytes.Trim(data, "\"")
	str := string(d)

	switch str {
	case scalar:
		*s = Scalar
	case band:
		*s = Band
	case heatmap:
		*s = Heatmap
	case polyline:
		*s = Polyline
	default:
		return fmt.Errorf("cannot unmarshal '%s': %w", str, errUnknownShape)
	}

	return nil
}
