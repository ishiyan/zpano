//nolint:dupl
package aroon

import (
	"bytes"
	"fmt"
)

// AroonOutput describes the outputs of the indicator.
type AroonOutput int

const (
	// The Aroon Up line.
	AroonUp AroonOutput = iota + 1

	// The Aroon Down line.
	AroonDown

	// The Aroon Oscillator (AroonUp - AroonDown).
	AroonOsc

	aroonLast
)

const (
	aroonOutputUp      = "up"
	aroonOutputDown    = "down"
	aroonOutputOsc     = "osc"
	aroonOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o AroonOutput) String() string {
	switch o {
	case AroonUp:
		return aroonOutputUp
	case AroonDown:
		return aroonOutputDown
	case AroonOsc:
		return aroonOutputOsc
	default:
		return aroonOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o AroonOutput) IsKnown() bool {
	return o >= AroonUp && o < aroonLast
}

// MarshalJSON implements the Marshaler interface.
func (o AroonOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown aroon output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == aroonOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *AroonOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown aroon output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case aroonOutputUp:
		*o = AroonUp
	case aroonOutputDown:
		*o = AroonDown
	case aroonOutputOsc:
		*o = AroonOsc
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
