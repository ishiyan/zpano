//nolint:dupl
package aroon

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// The Aroon Up line.
	Up Output = iota + 1

	// The Aroon Down line.
	Down

	// The Aroon Oscillator (Up - Down).
	Osc

	outputLast
)

const (
	upStr      = "up"
	downStr    = "down"
	oscStr     = "osc"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Up:
		return upStr
	case Down:
		return downStr
	case Osc:
		return oscStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Up && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown aroon output"
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
		errFmt = "cannot unmarshal '%s': unknown aroon output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case upStr:
		*o = Up
	case downStr:
		*o = Down
	case oscStr:
		*o = Osc
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
