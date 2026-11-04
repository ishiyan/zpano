// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclicsmoothedrelativestrengthindex

import (
	"math"

	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// WilderRSI is the prepared RSI seed/recurrence; a flat seed returns 100 and length one is valid.
type WilderRSI struct {
	length, count      int
	previous, up, down float64
}

// NewWilderRSI creates a reusable prepared RSI engine.
func NewWilderRSI(length int) *WilderRSI {
	if err := harmoniccycle.Integer(length, "length", 1); err != nil {
		panic(err)
	}
	return &WilderRSI{length: length, previous: math.NaN()}
}

// Update accepts an already validated finite close.
func (s *WilderRSI) Update(price float64) float64 {
	if math.IsNaN(s.previous) {
		s.previous = price
		return math.NaN()
	}
	change := harmoniccycle.Finite(price - s.previous)
	gain, loss := max(change, 0), max(-change, 0)
	if s.count < s.length {
		s.up = harmoniccycle.Finite(s.up + gain)
		s.down = harmoniccycle.Finite(s.down + loss)
		s.count++
		s.previous = price
		if s.count < s.length {
			return math.NaN()
		}
		s.up /= float64(s.length)
		s.down /= float64(s.length)
	} else {
		s.up = harmoniccycle.Finite((s.up*float64(s.length-1) + gain) / float64(s.length))
		s.down = harmoniccycle.Finite((s.down*float64(s.length-1) + loss) / float64(s.length))
		s.previous = price
	}
	if s.down == 0 {
		return 100
	}
	if s.up == 0 {
		return 0
	}
	return 100 - 100/(1+s.up/s.down)
}
