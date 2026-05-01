package jurikzerolagvelocity

import "encoding/json"

// Output identifies the output of the JurikZeroLagVelocity indicator.
type Output int

const (
	// Value is the main VEL output line.
	Value Output = iota + 1
)

// String returns the string representation.
func (o Output) String() string {
	if o == Value {
		return "value"
	}

	return "unknown"
}

// IsKnown returns true if the output is known.
func (o Output) IsKnown() bool {
	return o == Value
}

// MarshalJSON implements the json.Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (o *Output) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "value":
		*o = Value
	default:
		*o = Value
	}

	return nil
}
