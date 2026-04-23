package coronaspectrum

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Value is the Corona spectrum heatmap column.
	Value Output = iota + 1
	// DominantCycle is the weighted-center-of-gravity dominant cycle estimate.
	DominantCycle
	// DominantCycleMedian is the 5-sample median of DominantCycle.
	DominantCycleMedian
	outputLast
)

const (
	valueStr               = "value"
	dominantCycleStr       = "dominantCycle"
	dominantCycleMedianStr = "dominantCycleMedian"
	unknownStr             = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	case DominantCycle:
		return dominantCycleStr
	case DominantCycleMedian:
		return dominantCycleMedianStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Value && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown corona spectrum output"
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
		errFmt = "cannot unmarshal '%s': unknown corona spectrum output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	case dominantCycleStr:
		*o = DominantCycle
	case dominantCycleMedianStr:
		*o = DominantCycleMedian
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
