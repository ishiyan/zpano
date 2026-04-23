//nolint:testpackage
package hilberttransformer

import (
	"testing"
)

func BenchmarkHomodyneDiscriminatorEstimator(b *testing.B) {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.2,
		AlphaEmaPeriod:            0.2,
	}

	input := testHomodyneDiscriminatorEstimatorInput()

	for b.Loop() {
		hde, _ := NewHomodyneDiscriminatorEstimator(&params)
		for i := range input {
			hde.Update(input[i])
		}
	}
}

func BenchmarkHomodyneDiscriminatorEstimatorUnrolled(b *testing.B) {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.2,
		AlphaEmaPeriod:            0.2,
	}

	input := testHomodyneDiscriminatorEstimatorUnrolledInput()

	for b.Loop() {
		hdeu, _ := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		for i := range input {
			hdeu.Update(input[i])
		}
	}
}

func BenchmarkDualDifferentiatorEstimator(b *testing.B) {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.15,
		AlphaEmaPeriod:            0.25,
	}

	input := testDualDifferentiatorEstimatorInput()

	for b.Loop() {
		dde, _ := NewDualDifferentiatorEstimator(&params)
		for i := range input {
			dde.Update(input[i])
		}
	}
}

func BenchmarkPhaseAccumulatorEstimator(b *testing.B) {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.15,
		AlphaEmaPeriod:            0.15,
	}

	input := testPhaseAccumulatorEstimatorInput()

	for b.Loop() {
		pae, _ := NewPhaseAccumulatorEstimator(&params)
		for i := range input {
			pae.Update(input[i])
		}
	}
}
