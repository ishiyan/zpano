package core

// Specification contains all info needed to create an indicator.
type Specification struct {
	// Type identifies an indicator type.
	Type Type `json:"type"`

	// Parameters describe parameters to create an indicator.
	// The concrete type is defined by the related indicator, which in turn is defined by the Type field.
	Parameters any `json:"parameters"`

	// Outputs describes kinds of indicator outputs to calculate.
	Outputs []int `json:"outputs"`
}
