// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclephasescore

import (
	"encoding/json"
	"fmt"
)

// Output describes the registered plot components.
type Output int

const (
	// Value is the value.
	Value Output = iota + 1
	// Period is the period.
	Period
	// Phase is the phase.
	Phase
	outputLast
)

// String implements fmt.Stringer.
func (o Output) String() string {
	switch o {
	case Value:
		return "value"
	case Period:
		return "period"
	case Phase:
		return "phase"
	default:
		return "unknown"
	}
}

// IsKnown identifies a registered output.
func (o Output) IsKnown() bool { return o >= 1 && o < outputLast }

// MarshalJSON implements json.Marshaler.
func (o Output) MarshalJSON() ([]byte, error) {
	if !o.IsKnown() {
		return nil, fmt.Errorf("unknown cycle-phase-score output")
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
	return fmt.Errorf("unknown cycle-phase-score output: %s", value)
}
