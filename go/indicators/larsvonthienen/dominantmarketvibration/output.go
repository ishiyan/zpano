// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package dominantmarketvibration

import (
	"encoding/json"
	"fmt"
)

// Output describes the registered plot components.
type Output int

const (
	// Period is the period.
	Period Output = iota + 1
	// Amplitude is the amplitude.
	Amplitude
	// Phase is the phase.
	Phase
	// Strength is the strength.
	Strength
	// GenuinePercent is the genuine percent.
	GenuinePercent
	// Value is the value.
	Value
	// Slope is the slope.
	Slope
	// EtaTop is the eta top.
	EtaTop
	// EtaBottom is the eta bottom.
	EtaBottom
	outputLast
)

// String implements fmt.Stringer.
func (o Output) String() string {
	switch o {
	case Period:
		return "period"
	case Amplitude:
		return "amplitude"
	case Phase:
		return "phase"
	case Strength:
		return "strength"
	case GenuinePercent:
		return "genuine_percent"
	case Value:
		return "value"
	case Slope:
		return "slope"
	case EtaTop:
		return "eta_top"
	case EtaBottom:
		return "eta_bottom"
	default:
		return "unknown"
	}
}

// IsKnown identifies a registered output.
func (o Output) IsKnown() bool { return o >= 1 && o < outputLast }

// MarshalJSON implements json.Marshaler.
func (o Output) MarshalJSON() ([]byte, error) {
	if !o.IsKnown() {
		return nil, fmt.Errorf("unknown dominant-market-vibration output")
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
	return fmt.Errorf("unknown dominant-market-vibration output: %s", value)
}
