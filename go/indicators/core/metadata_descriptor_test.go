package core_test

import (
	"testing"

	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

// TestDescriptorOutputsWellFormed verifies invariants that every
// indicator's Metadata() now relies on via core.BuildMetadata:
//   - each descriptor has at least one output;
//   - output Kinds are strictly ascending, starting at 1
//     (matching the Go-side `iota + 1` per-indicator output enums);
//   - every output has a known Shape.
func TestDescriptorOutputsWellFormed(t *testing.T) {
	known := map[shape.Shape]bool{
		shape.Scalar:   true,
		shape.Band:     true,
		shape.Heatmap:  true,
		shape.Polyline: true,
	}

	for id, desc := range core.Descriptors() {
		t.Run(id.String(), func(t *testing.T) {
			if len(desc.Outputs) == 0 {
				t.Fatalf("descriptor %v has no outputs", id)
			}

			for i, o := range desc.Outputs {
				want := i + 1
				if o.Kind != want {
					t.Errorf("output[%d].Kind = %d, want %d (must start at 1 and be strictly ascending)", i, o.Kind, want)
				}
				if !known[o.Shape] {
					t.Errorf("output[%d].Shape = %v is not a known shape", i, o.Shape)
				}
			}

			if desc.Identifier != id {
				t.Errorf("descriptor.Identifier = %v, keyed under %v", desc.Identifier, id)
			}
		})
	}
}
