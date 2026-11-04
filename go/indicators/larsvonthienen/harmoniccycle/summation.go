// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package harmoniccycle

import (
	"math"
)

// Sum uses non-overlapping partials to retain precision when harmonic values cancel.
func Sum(values []float64) float64 {
	partials := make([]float64, 0, 32)
	for _, value := range values {
		i := 0
		for _, y := range partials {
			if math.Abs(value) < math.Abs(y) {
				value, y = y, value
			}
			high := value + y
			low := y - (high - value)
			if low != 0 {
				partials[i] = low
				i++
			}
			value = high
		}
		partials = partials[:i]
		partials = append(partials, value)
	}
	var sum float64
	for i := len(partials) - 1; i >= 0; i-- {
		sum += partials[i]
	}
	return sum
}
