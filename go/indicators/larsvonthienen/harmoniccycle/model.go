// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package harmoniccycle

import (
	"math"
)

// CycleModel stores sine phase at the first sample of the analysis window.
type CycleModel struct{ Period, Amplitude, Phase, TroughOffset float64 }

// Goertzel implements the published single-period GDFT, preserving normalization and arithmetic order.
func Goertzel(values []float64, period float64) CycleModel {
	if err := Real(period, "period", 2.0000000000000004); err != nil {
		panic(err)
	}
	n := len(values)
	if n < 1 {
		panic("Goertzel needs observations")
	}
	omega := 2 * math.Pi / period
	sine, cosine := math.Sin(omega), math.Cos(omega)
	coefficient := 2 * cosine
	var q1, q2 float64
	for _, value := range values {
		Sample(value)
		if math.IsNaN(value) {
			panic("Goertzel batch input must be finite")
		}
		q0 := Finite(coefficient*q1 - q2 + value)
		q2, q1 = q1, q0
	}
	r1, i1 := q1-q2*cosine, q2*sine
	cn, sn := math.Cos(float64(n-1)*omega), math.Sin(float64(n-1)*omega)
	real := Finite(r1*cn + i1*sn)
	imaginary := Finite(i1*cn - r1*sn)
	amplitude := Finite(2 * math.Sqrt(Finite(real*real+imaginary*imaginary)) / float64(n))
	if amplitude == 0 {
		return CycleModel{period, 0, math.NaN(), math.NaN()}
	}
	phase := math.Pi/2 + math.Atan2(imaginary, real)
	offset := period * (0.75 - phase/(2*math.Pi))
	return CycleModel{period, amplitude, phase, offset}
}

// CycleValue evaluates a fitted wave without advancing observations.
func CycleValue(model CycleModel, offset float64) float64 {
	if model.Amplitude == 0 {
		return 0
	}
	return model.Amplitude * math.Sin(2*math.Pi*offset/model.Period+model.Phase)
}

// CycleSlope returns the analytic derivative in price units per bar.
func CycleSlope(model CycleModel, offset float64) float64 {
	if model.Amplitude == 0 {
		return 0
	}
	omega := 2 * math.Pi / model.Period
	return model.Amplitude * omega * math.Cos(omega*offset+model.Phase)
}

// CompositeValue sums configured harmonic models.
func CompositeValue(models []CycleModel, offset float64) float64 {
	var sum float64
	for _, model := range models {
		sum += CycleValue(model, offset)
	}
	return Finite(sum)
}

// CycleStrength returns the disclosed amplitude/period influence metric.
func CycleStrength(amplitude, period float64) float64 {
	if err := Real(amplitude, "amplitude", 0); err != nil {
		panic(err)
	}
	if err := Real(period, "period", 2.0000000000000004); err != nil {
		panic(err)
	}
	return amplitude / period
}

// CyclePolarity reports agreement between short and long slopes.
func CyclePolarity(shortSlope, longSlope, tolerance float64) float64 {
	Sample(shortSlope)
	Sample(longSlope)
	if err := Real(tolerance, "tolerance", 0); err != nil {
		panic(err)
	}
	if math.IsNaN(shortSlope) || math.IsNaN(longSlope) {
		return math.NaN()
	}
	if shortSlope > tolerance && longSlope > tolerance {
		return 1
	}
	if shortSlope < -tolerance && longSlope < -tolerance {
		return -1
	}
	return 0
}

// PhaseState is the documented continuous research reconstruction.
func PhaseState(phase float64) float64 { Sample(phase); return 100 * math.Sin(phase) }
