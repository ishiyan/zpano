package outputs

import (
	"zpano/indicators/core/outputs/shape"
)

// Metadata describes a single indicator output.
type Metadata struct {
	// Kind is an identification of this indicator output.
	// It is an integer representation of an output enumeration of a related indicator.
	Kind int `json:"kind"`

	// Shape describes the data shape of this indicator output.
	Shape shape.Shape `json:"shape"`

	// Mnemonic is a short name (mnemonic) of this indicator output.
	Mnemonic string `json:"mnemonic"`

	// Description is a description of this indicator output.
	Description string `json:"description"`
}
