//nolint:dupl
package movingaverageconvergencedivergence

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// MACD is the MACD line value (fast MA - slow MA).
	MACD Output = iota + 1

	// Signal is the signal line value (MA of MACD line).
	Signal

	// Histogram is the histogram value (MACD - signal).
	Histogram

	outputLast
)

const (
	mACDValueStr      = "macdValue"
	signalValueStr    = "signalValue"
	histogramValueStr = "histogramValue"
	unknownStr        = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case MACD:
		return mACDValueStr
	case Signal:
		return signalValueStr
	case Histogram:
		return histogramValueStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= MACD && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown moving average convergence divergence output"
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
		errFmt = "cannot unmarshal '%s': unknown moving average convergence divergence output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case mACDValueStr:
		*o = MACD
	case signalValueStr:
		*o = Signal
	case histogramValueStr:
		*o = Histogram
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
