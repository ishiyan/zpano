//nolint:dupl
package fractalbandshybrideadaptive

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Frasma2 is the fractal adaptive simple moving average center line.
	Frasma2 Output = iota + 1
	// Upper is the upper band (frasma2 + alpha^H * 2*stddev).
	Upper
	// Lower is the lower band (frasma2 - alpha^H * 2*stddev).
	Lower
	// Band is the lower/upper band pair.
	Band
	outputLast
)

const (
	frasma2Str = "frasma2"
	upperStr   = "upper"
	lowerStr   = "lower"
	bandStr    = "band"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Frasma2:
		return frasma2Str
	case Upper:
		return upperStr
	case Lower:
		return lowerStr
	case Band:
		return bandStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Frasma2 && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown fractal bands hybride adaptive output"
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
		errFmt = "cannot unmarshal '%s': unknown fractal bands hybride adaptive output"
		dqs    = "\""
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case frasma2Str:
		*o = Frasma2
	case upperStr:
		*o = Upper
	case lowerStr:
		*o = Lower
	case bandStr:
		*o = Band
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
