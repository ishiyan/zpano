//nolint:dupl
package hurstdifference

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// HurstDiff is the first difference of the FGDI.
	HurstDiff Output = iota + 1

	// Fgdi is the raw FGDI value.
	Fgdi

	outputLast
)

const (
	hurstDiffStr = "hurstDiff"
	fgdiStr      = "fgdi"
	unknownStr   = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case HurstDiff:
		return hurstDiffStr
	case Fgdi:
		return fgdiStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= HurstDiff && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown hurst difference output"
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
		errFmt = "cannot unmarshal '%s': unknown hurst difference output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case hurstDiffStr:
		*o = HurstDiff
	case fgdiStr:
		*o = Fgdi
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
