//nolint:dupl
package moneyflowindex

import (
	"bytes"
	"fmt"
)

// MoneyFlowIndexOutput describes the outputs of the indicator.
type MoneyFlowIndexOutput int

const (
	// The scalar value of the money flow index.
	MoneyFlowIndexValue MoneyFlowIndexOutput = iota + 1
	moneyFlowIndexLast
)

const (
	moneyFlowIndexOutputValue   = "value"
	moneyFlowIndexOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o MoneyFlowIndexOutput) String() string {
	switch o {
	case MoneyFlowIndexValue:
		return moneyFlowIndexOutputValue
	default:
		return moneyFlowIndexOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o MoneyFlowIndexOutput) IsKnown() bool {
	return o >= MoneyFlowIndexValue && o < moneyFlowIndexLast
}

// MarshalJSON implements the Marshaler interface.
func (o MoneyFlowIndexOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown money flow index output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == moneyFlowIndexOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *MoneyFlowIndexOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown money flow index output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case moneyFlowIndexOutputValue:
		*o = MoneyFlowIndexValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
