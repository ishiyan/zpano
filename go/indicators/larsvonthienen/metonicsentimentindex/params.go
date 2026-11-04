// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package metonicsentimentindex

import ()

// Params configures the indicator; explicit zeros retain their numerical meaning.
type Params struct {
	// Length configures length; the reference default is 13.
	Length int
	// Delay1 configures delay1; the reference default is 6939.
	Delay1 int
	// Delay2 configures delay2; the reference default is 4016.
	Delay2 int
	// Delay3 configures delay3; the reference default is 2922.
	Delay3 int
	// Mode configures mode; the reference default is 0.
	Mode int
}

// DefaultParams returns the prepared reference defaults.
func DefaultParams() *Params {
	return &Params{Length: 13, Delay1: 6939, Delay2: 4016, Delay3: 2922, Mode: 0}
}
