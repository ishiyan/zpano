package core

import (
	"bytes"
	"errors"
	"fmt"
)

// Adaptivity classifies whether an indicator adapts its parameters to market conditions.
type Adaptivity int

const (
	// Static denotes an indicator with fixed parameters.
	Static Adaptivity = iota + 1

	// Adaptive denotes an indicator that adapts parameters to market conditions
	// (e.g., via dominant cycle, efficiency ratio, or fractal dimension).
	Adaptive
	adaptivityLast
)

const (
	adaptivityUnknown  = "unknown"
	adaptivityStatic   = "static"
	adaptivityAdaptive = "adaptive"
)

var errUnknownAdaptivity = errors.New("unknown indicator adaptivity")

// String implements the Stringer interface.
func (a Adaptivity) String() string {
	switch a {
	case Static:
		return adaptivityStatic
	case Adaptive:
		return adaptivityAdaptive
	default:
		return adaptivityUnknown
	}
}

// IsKnown determines if this adaptivity is known.
func (a Adaptivity) IsKnown() bool {
	return a >= Static && a < adaptivityLast
}

// MarshalJSON implements the Marshaler interface.
func (a Adaptivity) MarshalJSON() ([]byte, error) {
	s := a.String()
	if s == adaptivityUnknown {
		return nil, fmt.Errorf("cannot marshal '%s': %w", s, errUnknownAdaptivity)
	}

	const extra = 2

	b := make([]byte, 0, len(s)+extra)
	b = append(b, '"')
	b = append(b, s...)
	b = append(b, '"')

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (a *Adaptivity) UnmarshalJSON(data []byte) error {
	d := bytes.Trim(data, "\"")
	s := string(d)

	switch s {
	case adaptivityStatic:
		*a = Static
	case adaptivityAdaptive:
		*a = Adaptive
	default:
		return fmt.Errorf("cannot unmarshal '%s': %w", s, errUnknownAdaptivity)
	}

	return nil
}
