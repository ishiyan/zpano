//nolint:dupl
package movingminimax

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Up is the up mini-max value at the most recent bar (emphasizes local maxima).
	Up Output = iota + 1

	// Down is the down mini-max value at the most recent bar (emphasizes local minima).
	Down

	// Resistances is the detected resistance levels, sorted by strength (strongest first).
	Resistances

	// Supports is the detected support levels, sorted by strength (strongest first).
	Supports

	// UpDistribution is the full up mini-max probability distribution over the window.
	UpDistribution

	// DownDistribution is the full down mini-max probability distribution over the window.
	DownDistribution

	outputLast
)

const (
	upStr       = "up"
	downStr     = "down"
	resistStr   = "resistances"
	supportsStr = "supports"
	upDistStr   = "upDistribution"
	downDistStr = "downDistribution"
	unknownStr  = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Up:
		return upStr
	case Down:
		return downStr
	case Resistances:
		return resistStr
	case Supports:
		return supportsStr
	case UpDistribution:
		return upDistStr
	case DownDistribution:
		return downDistStr
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
		errFmt = "cannot marshal '%s': unknown moving mini-max output"
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
		errFmt = "cannot unmarshal '%s': unknown moving mini-max output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case upStr:
		*o = Up
	case downStr:
		*o = Down
	case resistStr:
		*o = Resistances
	case supportsStr:
		*o = Supports
	case upDistStr:
		*o = UpDistribution
	case downDistStr:
		*o = DownDistribution
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
