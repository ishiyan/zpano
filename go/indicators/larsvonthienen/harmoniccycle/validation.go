// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package harmoniccycle

import (
	"fmt"
	"math"
)

// Real checks a finite parameter against its lower bound.
func Real(value float64, name string, minimum float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < minimum {
		return fmt.Errorf("invalid %s: must be finite and >= %g", name, minimum)
	}
	return nil
}

// Integer checks an integer parameter against its lower bound.
func Integer(value int, name string, minimum int) error {
	if value < minimum {
		return fmt.Errorf("invalid %s: must be >= %d", name, minimum)
	}
	return nil
}

// Finite rejects arithmetic overflow. NaNs used as unavailable outputs are handled by callers.
func Finite(value float64) float64 {
	if math.IsInf(value, 0) || math.IsNaN(value) {
		panic("non-finite numerical result")
	}
	return value
}

// Sample rejects infinite observations; NaN represents missing input.
func Sample(value float64) {
	if math.IsInf(value, 0) {
		panic("sample must be finite or NaN")
	}
}

// Push retains the latest capacity values in oldest-first order.
func Push(values []float64, value float64, capacity int) []float64 {
	if len(values) == capacity {
		copy(values, values[1:])
		values[len(values)-1] = value
		return values
	}
	return append(values, value)
}
