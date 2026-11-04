// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package goertzelcyclespectrum

import (
	"encoding/json"
	"fmt"
)

// Output describes the registered plot components.
type Output int

const (
	// Amplitude is the amplitude.
	Amplitude Output = iota + 1
	// Phases is the phases.
	Phases
	// TroughOffsets is the trough offsets.
	TroughOffsets
	// DominantPeriod is the dominant period.
	DominantPeriod
	// DominantAmplitude is the dominant amplitude.
	DominantAmplitude
	// DominantPhase is the dominant phase.
	DominantPhase
	outputLast
)

// String implements fmt.Stringer.
func (o Output) String() string {
	switch o {
	case Amplitude:
		return "amplitude"
	case Phases:
		return "phases"
	case TroughOffsets:
		return "trough_offsets"
	case DominantPeriod:
		return "dominant_period"
	case DominantAmplitude:
		return "dominant_amplitude"
	case DominantPhase:
		return "dominant_phase"
	default:
		return "unknown"
	}
}

// IsKnown identifies a registered output.
func (o Output) IsKnown() bool { return o >= 1 && o < outputLast }

// MarshalJSON implements json.Marshaler.
func (o Output) MarshalJSON() ([]byte, error) {
	if !o.IsKnown() {
		return nil, fmt.Errorf("unknown goertzel-cycle-spectrum output")
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
	return fmt.Errorf("unknown goertzel-cycle-spectrum output: %s", value)
}
