// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package hodrickprescottsmoothinganddetrending

import (
	"encoding/json"
	"fmt"
)

// Output describes the registered plot components.
type Output int

const (
	// Trend is the trend.
	Trend Output = iota + 1
	// Cycle is the cycle.
	Cycle
	outputLast
)

// String implements fmt.Stringer.
func (o Output) String() string {
	switch o {
	case Trend:
		return "trend"
	case Cycle:
		return "cycle"
	default:
		return "unknown"
	}
}

// IsKnown identifies a registered output.
func (o Output) IsKnown() bool { return o >= 1 && o < outputLast }

// MarshalJSON implements json.Marshaler.
func (o Output) MarshalJSON() ([]byte, error) {
	if !o.IsKnown() {
		return nil, fmt.Errorf("unknown hodrick-prescott-smoothing-and-detrending output")
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
	return fmt.Errorf("unknown hodrick-prescott-smoothing-and-detrending output: %s", value)
}
