package core

import (
	"fmt"

	"zpano/indicators/core/outputs"
)

// OutputText provides the per-output mnemonic and description
// used when building a Metadata from the descriptor registry.
type OutputText struct {
	// Mnemonic is a short name (mnemonic) of the output.
	Mnemonic string

	// Description is a description of the output.
	Description string
}

// BuildMetadata constructs a Metadata for the indicator with the given identifier
// by joining the registry's per-output Kind and Shape with the supplied
// per-output mnemonic and description.
//
// texts must be in the same order and length as the descriptor's Outputs.
// It panics if no descriptor is registered for the identifier or if the
// length of texts does not match the descriptor's Outputs.
func BuildMetadata(id Identifier, mnemonic, description string, texts []OutputText) Metadata {
	d, ok := DescriptorOf(id)
	if !ok {
		panic(fmt.Sprintf("core.BuildMetadata: no descriptor registered for identifier %v", id))
	}

	if len(texts) != len(d.Outputs) {
		panic(fmt.Sprintf(
			"core.BuildMetadata: identifier %v has %d outputs in descriptor but %d texts were supplied",
			id, len(d.Outputs), len(texts),
		))
	}

	out := make([]outputs.Metadata, len(texts))
	for i, t := range texts {
		out[i] = outputs.Metadata{
			Kind:        d.Outputs[i].Kind,
			Shape:       d.Outputs[i].Shape,
			Mnemonic:    t.Mnemonic,
			Description: t.Description,
		}
	}

	return Metadata{
		Identifier:  id,
		Mnemonic:    mnemonic,
		Description: description,
		Outputs:     out,
	}
}
