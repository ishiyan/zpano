package core

// Specification contains all info needed to create an indicator.
type Specification struct {
	// Identifier identifies an indicator.
	Identifier Identifier `json:"identifier"`

	// Parameters describe parameters to create an indicator.
	// The concrete type is defined by the related indicator, which in turn is defined by the Identifier field.
	Parameters any `json:"parameters"`

	// Outputs describes kinds of indicator outputs to calculate.
	Outputs []int `json:"outputs"`
}
