package core

import (
	"zpano/indicators/core/outputs/shape"
)

// OutputDescriptor classifies a single indicator output for charting / discovery.
type OutputDescriptor struct {
	// Kind is an integer representation of the output enumeration of a related indicator.
	Kind int `json:"kind"`

	// Shape is the data shape of this output.
	Shape shape.Shape `json:"shape"`

	// Role is the semantic role of this output.
	Role Role `json:"role"`

	// Pane is the chart pane on which this output is drawn.
	Pane Pane `json:"pane"`
}
