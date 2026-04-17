//nolint:dupl
package balanceofpower

import (
	"bytes"
	"fmt"
)

// BalanceOfPowerOutput describes the outputs of the indicator.
type BalanceOfPowerOutput int

const (
	// The scalar value of the balance of power.
	BalanceOfPowerValue BalanceOfPowerOutput = iota + 1
	balanceOfPowerLast
)

const (
	balanceOfPowerOutputValue   = "value"
	balanceOfPowerOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o BalanceOfPowerOutput) String() string {
	switch o {
	case BalanceOfPowerValue:
		return balanceOfPowerOutputValue
	default:
		return balanceOfPowerOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o BalanceOfPowerOutput) IsKnown() bool {
	return o >= BalanceOfPowerValue && o < balanceOfPowerLast
}

// MarshalJSON implements the Marshaler interface.
func (o BalanceOfPowerOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown balance of power output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == balanceOfPowerOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *BalanceOfPowerOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown balance of power output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case balanceOfPowerOutputValue:
		*o = BalanceOfPowerValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
