// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package completecyclescanner

import (
	"encoding/json"
	"fmt"
)

// Output describes the registered plot components.
type Output int

const (
	// Spectrum is the spectrum.
	Spectrum Output = iota + 1
	// GenuinePercent is the genuine percent.
	GenuinePercent
	// DominantPeriod is the dominant period.
	DominantPeriod
	// AcceptedCount is the accepted count.
	AcceptedCount
	outputLast
)

// String implements fmt.Stringer.
func (o Output) String() string {
	switch o {
	case Spectrum:
		return "spectrum"
	case GenuinePercent:
		return "genuine_percent"
	case DominantPeriod:
		return "dominant_period"
	case AcceptedCount:
		return "accepted_count"
	default:
		return "unknown"
	}
}

// IsKnown identifies a registered output.
func (o Output) IsKnown() bool { return o >= 1 && o < outputLast }

// MarshalJSON implements json.Marshaler.
func (o Output) MarshalJSON() ([]byte, error) {
	if !o.IsKnown() {
		return nil, fmt.Errorf("unknown complete-cycle-scanner output")
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
	return fmt.Errorf("unknown complete-cycle-scanner output: %s", value)
}
