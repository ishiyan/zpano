//nolint:dupl
package averagetruerange

import (
	"bytes"
	"fmt"
)

// AverageTrueRangeOutput describes the outputs of the indicator.
type AverageTrueRangeOutput int

const (
	// The scalar value of the average true range.
	AverageTrueRangeValue AverageTrueRangeOutput = iota + 1
	averageTrueRangeLast
)

const (
	averageTrueRangeOutputValue   = "value"
	averageTrueRangeOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o AverageTrueRangeOutput) String() string {
	switch o {
	case AverageTrueRangeValue:
		return averageTrueRangeOutputValue
	default:
		return averageTrueRangeOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o AverageTrueRangeOutput) IsKnown() bool {
	return o >= AverageTrueRangeValue && o < averageTrueRangeLast
}

// MarshalJSON implements the Marshaler interface.
func (o AverageTrueRangeOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown average true range output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == averageTrueRangeOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *AverageTrueRangeOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown average true range output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case averageTrueRangeOutputValue:
		*o = AverageTrueRangeValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
