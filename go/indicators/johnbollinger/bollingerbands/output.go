//nolint:dupl
package bollingerbands

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Lower is the lower band value.
	Lower Output = iota + 1

	// Middle is the middle band (moving average) value.
	Middle

	// Upper is the upper band value.
	Upper

	// BandWidth is the band width value.
	BandWidth

	// PercentBand is the percent band (%B) value.
	PercentBand

	// Band is the lower/upper band.
	Band

	outputLast
)

const (
	lowerValueStr  = "lowerValue"
	middleValueStr = "middleValue"
	upperValueStr  = "upperValue"
	bandWidthStr   = "bandWidth"
	percentBandStr = "percentBand"
	bandStr        = "band"
	unknownStr     = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Lower:
		return lowerValueStr
	case Middle:
		return middleValueStr
	case Upper:
		return upperValueStr
	case BandWidth:
		return bandWidthStr
	case PercentBand:
		return percentBandStr
	case Band:
		return bandStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Lower && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown bollinger bands output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
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
		errFmt = "cannot unmarshal '%s': unknown bollinger bands output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case lowerValueStr:
		*o = Lower
	case middleValueStr:
		*o = Middle
	case upperValueStr:
		*o = Upper
	case bandWidthStr:
		*o = BandWidth
	case percentBandStr:
		*o = PercentBand
	case bandStr:
		*o = Band
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
