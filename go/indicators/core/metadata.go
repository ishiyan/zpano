package core

import (
	"zpano/indicators/core/outputs"
)

// Metadata describes an indicator and its outputs.
type Metadata struct {
	// Identifier identifies this indicator.
	Identifier Identifier `json:"identifier"`

	// Mnemonic is a short name (mnemonic) of this indicator.
	Mnemonic string `json:"mnemonic"`

	// Description is a description of this indicator.
	Description string `json:"description"`

	// Outputs is a slice of metadata for individual outputs.
	Outputs []outputs.Metadata `json:"outputs"`
}
