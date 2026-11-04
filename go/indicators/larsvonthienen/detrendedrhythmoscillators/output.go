// SPDX-License-Identifier: MPL-2.0
// Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
// Numerical source: Lars von Thienen / WhenToTrade (2017), MPL-2.0.
// Research adaptations are documented in description.md.
package detrendedrhythmoscillators

import (
	"encoding/json"
	"fmt"
)

// Output describes the registered plot components.
type Output int

const (
	// Value is the value.
	Value Output = iota + 1
	// Direction is the direction.
	Direction
	// CycleLength is the cycle length.
	CycleLength
	// PivotValue is the pivot value.
	PivotValue
	// PivotOffset is the pivot offset.
	PivotOffset
	outputLast
)

// String implements fmt.Stringer.
func (o Output) String() string {
	switch o {
	case Value:
		return "value"
	case Direction:
		return "direction"
	case CycleLength:
		return "cycle_length"
	case PivotValue:
		return "pivot_value"
	case PivotOffset:
		return "pivot_offset"
	default:
		return "unknown"
	}
}

// IsKnown identifies a registered output.
func (o Output) IsKnown() bool { return o >= 1 && o < outputLast }

// MarshalJSON implements json.Marshaler.
func (o Output) MarshalJSON() ([]byte, error) {
	if !o.IsKnown() {
		return nil, fmt.Errorf("unknown detrended-rhythm-oscillators output")
	}
	return json.Marshal(o.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *Output) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	for i := Output(1); i < outputLast; i++ {
		if i.String() == value {
			*o = i
			return nil
		}
	}
	return fmt.Errorf("unknown detrended-rhythm-oscillators output: %s", value)
}
