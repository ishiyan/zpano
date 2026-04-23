package adaptivetrendandcyclefilter

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Fatl is the Fast Adaptive Trend Line (39-tap FIR).
	Fatl Output = iota + 1
	// Satl is the Slow Adaptive Trend Line (65-tap FIR).
	Satl
	// Rftl is the Reference Fast Trend Line (44-tap FIR).
	Rftl
	// Rstl is the Reference Slow Trend Line (91-tap FIR).
	Rstl
	// Rbci is the Range Bound Channel Index (56-tap FIR).
	Rbci
	// Ftlm is the Fast Trend Line Momentum (FATL - RFTL).
	Ftlm
	// Stlm is the Slow Trend Line Momentum (SATL - RSTL).
	Stlm
	// Pcci is the Perfect Commodity Channel Index (sample - FATL).
	Pcci
	outputLast
)

const (
	fatlStr    = "fatl"
	satlStr    = "satl"
	rftlStr    = "rftl"
	rstlStr    = "rstl"
	rbciStr    = "rbci"
	ftlmStr    = "ftlm"
	stlmStr    = "stlm"
	pcciStr    = "pcci"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Fatl:
		return fatlStr
	case Satl:
		return satlStr
	case Rftl:
		return rftlStr
	case Rstl:
		return rstlStr
	case Rbci:
		return rbciStr
	case Ftlm:
		return ftlmStr
	case Stlm:
		return stlmStr
	case Pcci:
		return pcciStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Fatl && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown adaptive trend and cycle filter output"
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
		errFmt = "cannot unmarshal '%s': unknown adaptive trend and cycle filter output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case fatlStr:
		*o = Fatl
	case satlStr:
		*o = Satl
	case rftlStr:
		*o = Rftl
	case rstlStr:
		*o = Rstl
	case rbciStr:
		*o = Rbci
	case ftlmStr:
		*o = Ftlm
	case stlmStr:
		*o = Stlm
	case pcciStr:
		*o = Pcci
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
