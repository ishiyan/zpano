// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package phasingoscillator

import (
	"encoding/json"
	"fmt"
)

// Output describes the registered plot components.
type Output int

const (
	// Value is the value.
	Value Output = iota + 1
	// ShortScore is the short score.
	ShortScore
	// LongScore is the long score.
	LongScore
	// ShortPeriod is the short period.
	ShortPeriod
	// LongPeriod is the long period.
	LongPeriod
	outputLast
)

// String implements fmt.Stringer.
func (o Output) String() string {
	switch o {
	case Value:
		return "value"
	case ShortScore:
		return "short_score"
	case LongScore:
		return "long_score"
	case ShortPeriod:
		return "short_period"
	case LongPeriod:
		return "long_period"
	default:
		return "unknown"
	}
}

// IsKnown identifies a registered output.
func (o Output) IsKnown() bool { return o >= 1 && o < outputLast }

// MarshalJSON implements json.Marshaler.
func (o Output) MarshalJSON() ([]byte, error) {
	if !o.IsKnown() {
		return nil, fmt.Errorf("unknown phasing-oscillator output")
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
	return fmt.Errorf("unknown phasing-oscillator output: %s", value)
}
