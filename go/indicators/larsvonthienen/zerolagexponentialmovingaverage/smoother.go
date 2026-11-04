// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package zerolagexponentialmovingaverage

import (
	"math"

	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// ZLEMA is the prepared first-price-prehistory smoother, distinct from Ehlers' correction filter.
type ZLEMA struct {
	alpha, first, value float64
	lag                 int
	history             []float64
}

// NewZLEMA constructs the reusable prepared engine.
func NewZLEMA(length int) *ZLEMA {
	if err := harmoniccycle.Integer(length, "length", 1); err != nil {
		panic(err)
	}
	return &ZLEMA{alpha: 2 / float64(length+1), lag: (length - 1) / 2, first: math.NaN(), value: math.NaN(), history: make([]float64, 0, (length-1)/2+1)}
}

// Update accepts an already validated finite observation.
func (s *ZLEMA) Update(price float64) float64 {
	if math.IsNaN(s.first) {
		s.first = price
		s.value = price
		s.history = harmoniccycle.Push(s.history, price, s.lag+1)
		return price
	}
	history := append(append([]float64{}, s.history...), price)
	lagged := s.first
	if len(history) > s.lag {
		lagged = history[len(history)-1-s.lag]
	}
	adjusted := harmoniccycle.Finite(2*price - lagged)
	value := harmoniccycle.Finite(s.alpha*adjusted + (1-s.alpha)*s.value)
	s.history = harmoniccycle.Push(s.history, price, s.lag+1)
	s.value = value
	return value
}
