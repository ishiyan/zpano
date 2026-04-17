//nolint:dupl
package onbalancevolume

import (
	"bytes"
	"fmt"
)

// OnBalanceVolumeOutput describes the outputs of the indicator.
type OnBalanceVolumeOutput int

const (
	// The scalar value of the on-balance volume.
	OnBalanceVolumeValue OnBalanceVolumeOutput = iota + 1
	onBalanceVolumeLast
)

const (
	onBalanceVolumeOutputValue   = "value"
	onBalanceVolumeOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o OnBalanceVolumeOutput) String() string {
	switch o {
	case OnBalanceVolumeValue:
		return onBalanceVolumeOutputValue
	default:
		return onBalanceVolumeOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o OnBalanceVolumeOutput) IsKnown() bool {
	return o >= OnBalanceVolumeValue && o < onBalanceVolumeLast
}

// MarshalJSON implements the Marshaler interface.
func (o OnBalanceVolumeOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown on-balance volume output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == onBalanceVolumeOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *OnBalanceVolumeOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown on-balance volume output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case onBalanceVolumeOutputValue:
		*o = OnBalanceVolumeValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
