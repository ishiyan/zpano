//nolint:dupl
package fractalgeneralizeddimensionindex

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Fgdi is the fractal graph dimension value.
	Fgdi Output = iota + 1
	// Upper is the upper band (fgdi + stddev).
	Upper
	// Lower is the lower band (fgdi - stddev).
	Lower
	// Stddev is the standard deviation of the dimension estimate.
	Stddev
	// Band is the lower/upper band pair.
	Band
	outputLast
)

const (
	fgdiStr    = "fgdi"
	upperStr   = "upper"
	lowerStr   = "lower"
	stddevStr  = "stddev"
	bandStr    = "band"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Fgdi:
		return fgdiStr
	case Upper:
		return upperStr
	case Lower:
		return lowerStr
	case Stddev:
		return stddevStr
	case Band:
		return bandStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Fgdi && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown fractal graph dimension index output"
		extra  = 2
		dqc    = '"'
	)

	s := o.String()
	if s == unknownStr {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *Output) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown fractal graph dimension index output"
		dqs    = "\""
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case fgdiStr:
		*o = Fgdi
	case upperStr:
		*o = Upper
	case lowerStr:
		*o = Lower
	case stddevStr:
		*o = Stddev
	case bandStr:
		*o = Band
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
