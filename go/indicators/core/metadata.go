package core

import (
	"zpano/indicators/core/outputs"
)

// Metadata describes a type and outputs of an indicator.
type Metadata struct {
	// Type identifies a type this indicator.
	Type Type `json:"type"`

	// Mnemonic is a short name (mnemonic) of this indicator.
	Mnemonic string `json:"mnemonic"`

	// Description is a description of this indicator.
	Description string `json:"description"`

	// Outputs is a slice of metadata for individual outputs.
	Outputs []outputs.Metadata `json:"outputs"`
}
