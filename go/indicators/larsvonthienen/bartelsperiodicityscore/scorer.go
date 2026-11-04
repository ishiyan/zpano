// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package bartelsperiodicityscore

import (
	"math"

	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// BartelsOutput is the segment harmonic consistency evidence.
type BartelsOutput struct{ Score, GenuinePercent, Segments float64 }

// BartelsScore scores the latest caller-selected complete segments; no-energy evidence is unavailable.
func BartelsScore(values []float64, period, maxSegments, minSegments int) BartelsOutput {
	if err := harmoniccycle.Integer(period, "period", 3); err != nil {
		panic(err)
	}
	if maxSegments < 2 || minSegments < 2 || minSegments > maxSegments {
		panic("invalid Bartels segments")
	}
	for _, value := range values {
		harmoniccycle.Sample(value)
		if math.IsNaN(value) {
			panic("Bartels batch data must be finite")
		}
	}
	segments := min(len(values)/period, maxSegments)
	nan := math.NaN()
	if segments < minSegments {
		return BartelsOutput{nan, nan, float64(segments)}
	}
	var sumA, sumB, sumEnergy float64
	for segment := range segments {
		var a, b float64
		for j := range period {
			x := segment*period + j
			angle := (360 / float64(period) * float64(x+1)) / 180 * math.Pi
			a += math.Sin(angle) * values[x]
			b += math.Cos(angle) * values[x]
		}
		sumA += a
		sumB += b
		sumEnergy = harmoniccycle.Finite(sumEnergy + a*a + b*b)
	}
	if sumEnergy == 0 {
		return BartelsOutput{nan, nan, float64(segments)}
	}
	averageA, averageB := sumA/float64(segments), sumB/float64(segments)
	amplitude := math.Sqrt(averageA*averageA + averageB*averageB)
	rmsAmplitude := math.Sqrt(sumEnergy / float64(segments))
	reference := rmsAmplitude / math.Sqrt(float64(segments))
	ratio := amplitude / reference
	score := math.Exp(-ratio * ratio)
	return BartelsOutput{score, 100 * (1 - score), float64(segments)}
}
