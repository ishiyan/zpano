// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclepolarity

import (
	"encoding/json"
	"fmt"
)

// Output describes the registered plot components.
type Output int

const (
	// Polarity is the polarity.
	Polarity Output = iota + 1
	// ShortValue is the short value.
	ShortValue
	// LongValue is the long value.
	LongValue
	// ShortSlope is the short slope.
	ShortSlope
	// LongSlope is the long slope.
	LongSlope
	outputLast
)

// String implements fmt.Stringer.
func (o Output) String() string {
	switch o {
	case Polarity:
		return "polarity"
	case ShortValue:
		return "short_value"
	case LongValue:
		return "long_value"
	case ShortSlope:
		return "short_slope"
	case LongSlope:
		return "long_slope"
	default:
		return "unknown"
	}
}

// IsKnown identifies a registered output.
func (o Output) IsKnown() bool { return o >= 1 && o < outputLast }

// MarshalJSON implements json.Marshaler.
func (o Output) MarshalJSON() ([]byte, error) {
	if !o.IsKnown() {
		return nil, fmt.Errorf("unknown cycle-polarity output")
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
	return fmt.Errorf("unknown cycle-polarity output: %s", value)
}
