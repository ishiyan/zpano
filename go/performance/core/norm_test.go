package core

import (
	"fmt"
	"testing"
)

func TestNormPDFScipyCompliance(t *testing.T) {
	t.Parallel()
	// These values cover symmetry, peak, and tails.
	inputs := []float64{
		-8.0, -5.0, -3.0,
		-2.0, -1.0, 0.0,
		1.0, 2.0, 3.0,
		5.0, 8.0,
	}
	outputs := []float64{
		5.052271083536893e-15, 1.4867195147342979e-06, 0.0044318484119380075,
		0.053990966513188056, 0.24197072451914337, 0.3989422804014327,
		0.24197072451914337, 0.053990966513188056, 0.0044318484119380075,
		1.4867195147342979e-06, 5.052271083536893e-15,
	}

	for i, x := range inputs {
		assertAlmostEqual(t, NormPDF(x), outputs[i], places(15), fmt.Sprintf("step %d", i))
		// Symmetry: NormPDF(x) == NormPDF(-x).
		assertAlmostEqual(t, NormPDF(-x), NormPDF(x), places(15), fmt.Sprintf("step %d symmetry", i))
	}
}

func TestNormCDFScipyCompliance(t *testing.T) {
	t.Parallel()
	// 2 is just inside Acklam's upper-tail threshold
	// (CDF(2) ≈ 0.97725 > 0.97575), so tests cover the branch boundary.
	inputs := []float64{
		-8.0, -5.0, -3.0,
		-2.0, -1.0, 0.0,
		1.0, 2.0, 3.0,
		5.0, 8.0,
	}
	outputs := []float64{
		6.22096057427174e-16, 2.866515718791933e-07, 0.0013498980316300933,
		0.022750131948179195, 0.15865525393145707, 0.5,
		0.8413447460685429, 0.9772498680518208, 0.9986501019683699,
		0.9999997133484281, 0.9999999999999993,
	}

	for i, x := range inputs {
		assertAlmostEqual(t, NormCDF(x), outputs[i], places(15), fmt.Sprintf("step %d", i))
		// Symmetry: NormCDF(x) + NormCDF(-x) = 1.
		assertAlmostEqual(t, NormCDF(x)+NormCDF(-x), 1.0, places(15), fmt.Sprintf("step %d symmetry", i))
	}
}

func mustNormPPF(t *testing.T, p float64) float64 {
	t.Helper()
	z, err := NormPPF(p)
	if err != nil {
		t.Fatalf("NormPPF(%v): %v", p, err)
	}
	return z
}

func TestNormPPFScipyCompliance(t *testing.T) {
	t.Parallel()
	pValues := []float64{
		// Center
		0.5, 0.75, 0.9,
		// Branch boundaries (values around the transition)
		0.02425, 0.025, 0.975, 0.97575,
		// Common quantiles
		0.001, 0.005, 0.01,
		0.025, 0.05, 0.10,
		0.90, 0.95, 0.975,
		0.99, 0.995, 0.999,
		// Extreme tails (exercise numerical stability)
		1e-12, 1e-10, 1e-8,
		0.99999999, 0.9999999999, 0.999999999999,
	}
	// Hardcoded 1-pValues
	pSymmetricValues := []float64{
		// Center
		0.5, 0.25, 0.1,
		// Branch boundaries (values around the transition)
		0.97575, 0.975, 0.025, 0.02425,
		// Common quantiles
		0.999, 0.995, 0.99,
		0.975, 0.95, 0.90,
		0.10, 0.05, 0.025,
		0.01, 0.005, 0.001,
		// Extreme tails (exercise numerical stability)
		0.999999999999, 0.9999999999, 0.99999999,
		1e-8, 1e-10, 1e-12,
	}
	zValues := []float64{
		// Center
		0.0, 0.6744897501960817, 1.2815515655446004,
		// Branch boundaries (values around the transition)
		-1.972961051311885, -1.9599639845400545, 1.959963984540054, 1.972961051311885,
		// Common quantiles
		-3.090232306167813, -2.575829303548901, -2.3263478740408408,
		-1.9599639845400545, -1.6448536269514729, -1.2815515655446004,
		1.2815515655446004, 1.6448536269514722, 1.959963984540054,
		2.3263478740408408, 2.5758293035489004, 3.090232306167,
		// Extreme tails (exercise numerical stability)
		-7.034483825301131, -6.361340902404056, -5.612001244174789,
		5.612001243305505, 6.361340889697422, 7.0344869100478356,
	}

	// Peter Acklam states that relative error < 1.15 × 10^-9 over the
	// entire domain.
	for i, p := range pValues {
		// Extreme tails have much lower accuracy.
		delta := 3e-9
		if i >= 19 {
			delta = 8e-9
		}
		assertAlmostEqual(t, mustNormPPF(t, p), zValues[i], delta, fmt.Sprintf("step %d", i))

		// NormPPF is the inverse of NormCDF: p ≈ Φ(Φ⁻¹(p)).
		assertAlmostEqual(t, NormCDF(mustNormPPF(t, p)), p, delta, fmt.Sprintf("step %d roundtrip", i))

		// Symmetry: NormPPF(p) == -NormPPF(1 - p).
		delta = 1e-13
		if i >= 19 {
			delta = 4e-6
		}
		assertAlmostEqual(t, -mustNormPPF(t, pSymmetricValues[i]), mustNormPPF(t, p), delta,
			fmt.Sprintf("step %d symmetry", i))
	}
}

func TestNormPPFOutOfDomain(t *testing.T) {
	t.Parallel()
	for _, p := range []float64{0, 1, -0.5, 1.5} {
		z, err := NormPPF(p)
		if err == nil || err.Error() != "p must be between 0 and 1 (exclusive)" {
			t.Errorf("NormPPF(%v): expected error, got %v", p, err)
		}
		assertNaN(t, z, fmt.Sprintf("NormPPF(%v)", p))
	}
}
