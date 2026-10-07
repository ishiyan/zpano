package streamingkbn

import (
	"fmt"
	"testing"
)

// Reference values for x = baconBenchmark, y = bacon (portfolio), computed
// with exact rational arithmetic (fractions.Fraction on the binary float
// inputs, square roots with 50-digit decimal.Decimal), rounded to the
// nearest float.
const (
	regSlope       = 0.9988502086225746
	regIntercept   = -0.001030120844918352
	regCorrelation = 0.9693858148753051
	regCoMoment    = 0.033844
	regCovariance  = 0.0014101666666666668
	regVarianceX   = 0.0014117899305555557
	regVarianceY   = 0.0014989166666666668
)

// regProps maps the Python property names to the Go getter methods.
var regProps = map[string]func(*LinearRegressionKleinKBN) float64{
	"slope":       (*LinearRegressionKleinKBN).Slope,
	"intercept":   (*LinearRegressionKleinKBN).Intercept,
	"correlation": (*LinearRegressionKleinKBN).Correlation,
	"covariance":  (*LinearRegressionKleinKBN).Covariance,
	"co_moment":   (*LinearRegressionKleinKBN).CoMoment,
	"variance_x":  (*LinearRegressionKleinKBN).VarianceX,
	"variance_y":  (*LinearRegressionKleinKBN).VarianceY,
}

func feedReg(reg *LinearRegressionKleinKBN, xs, ys []float64) *LinearRegressionKleinKBN {
	for i := range min(len(xs), len(ys)) {
		reg.Update(xs[i], ys[i])
	}
	return reg
}

func assertRegAllNaN(t *testing.T, reg *LinearRegressionKleinKBN) {
	t.Helper()
	assertNaN(t, "slope", reg.Slope())
	assertNaN(t, "intercept", reg.Intercept())
	assertNaN(t, "correlation", reg.Correlation())
}

func TestLinearRegressionKleinKBN_Bacon(t *testing.T) {
	t.Parallel()

	reg := feedReg(NewLinearRegressionKleinKBN(), baconBenchmark, bacon)
	assertEqualInt(t, "n", reg.N(), len(bacon))
	assertAlmostEqual(t, "slope", reg.Slope(), regSlope, 14)
	assertAlmostEqual(t, "intercept", reg.Intercept(), regIntercept, 15)
	assertAlmostEqual(t, "correlation", reg.Correlation(), regCorrelation, 14)
	assertAlmostEqual(t, "co_moment", reg.CoMoment(), regCoMoment, 16)
	assertAlmostEqual(t, "covariance", reg.Covariance(), regCovariance, 16)
	assertAlmostEqual(t, "variance_x", reg.VarianceX(), regVarianceX, 16)
	assertAlmostEqual(t, "variance_y", reg.VarianceY(), regVarianceY, 16)
}

func TestLinearRegressionKleinKBN_PerfectFit(t *testing.T) {
	t.Parallel()

	reg := NewLinearRegressionKleinKBN()
	for i := range 5 {
		reg.Update(float64(i), float64(2*i+1))
	}
	assertAlmostEqual(t, "slope", reg.Slope(), 2.0, 13)
	assertAlmostEqual(t, "intercept", reg.Intercept(), 1.0, 13)
	assertAlmostEqual(t, "correlation", reg.Correlation(), 1.0, 15)
	assertAlmostEqual(t, "mean_x", reg.MeanX(), 2.0, 15)
	assertAlmostEqual(t, "mean_y", reg.MeanY(), 5.0, 15)
	assertAlmostEqual(t, "variance_x", reg.VarianceX(), 2.0, 15)
	assertAlmostEqual(t, "co_moment", reg.CoMoment(), 20.0, 13)
	assertAlmostEqual(t, "covariance", reg.Covariance(), 4.0, 13)
}

func TestLinearRegressionKleinKBN_NegativeCorrelation(t *testing.T) {
	t.Parallel()

	reg := NewLinearRegressionKleinKBN()
	for i := range 5 {
		x := float64(i)
		reg.Update(x, -2.0*x+1.0)
	}
	assertAlmostEqual(t, "slope", reg.Slope(), -2.0, 13)
	assertAlmostEqual(t, "intercept", reg.Intercept(), 1.0, 13)
	assertAlmostEqual(t, "correlation", reg.Correlation(), -1.0, 15)
	assertAlmostEqual(t, "covariance", reg.Covariance(), -4.0, 13)
}

func TestLinearRegressionKleinKBN_ConstantY(t *testing.T) {
	t.Parallel()

	reg := NewLinearRegressionKleinKBN()
	for i := range 5 {
		reg.Update(float64(i), 3.0)
	}
	assertAlmostEqual(t, "slope", reg.Slope(), 0.0, 13)
	assertAlmostEqual(t, "intercept", reg.Intercept(), 3.0, 13)
	assertEqual(t, "covariance", reg.Covariance(), 0.0)
	assertNaN(t, "correlation", reg.Correlation())
}

func TestLinearRegressionKleinKBN_ConstantX(t *testing.T) {
	t.Parallel()

	reg := NewLinearRegressionKleinKBN()
	for i := range 5 {
		reg.Update(3.0, float64(i))
	}
	assertEqual(t, "covariance", reg.Covariance(), 0.0)
	assertRegAllNaN(t, reg)
}

func TestLinearRegressionKleinKBN_Empty(t *testing.T) {
	t.Parallel()

	reg := NewLinearRegressionKleinKBN()
	assertEqualInt(t, "n", reg.N(), 0)
	assertEqual(t, "co_moment", reg.CoMoment(), 0.0)
	assertNaN(t, "covariance", reg.Covariance())
	assertNaN(t, "variance_x", reg.VarianceX())
	assertRegAllNaN(t, reg)
}

func TestLinearRegressionKleinKBN_SinglePoint(t *testing.T) {
	t.Parallel()

	reg := NewLinearRegressionKleinKBN()
	reg.Update(1.0, 2.0)
	assertEqual(t, "covariance", reg.Covariance(), 0.0)
	assertRegAllNaN(t, reg)
}

func TestLinearRegressionKleinKBN_TwoPoints(t *testing.T) {
	t.Parallel()

	reg := NewLinearRegressionKleinKBN()
	reg.Update(0.0, 1.0)
	reg.Update(2.0, 5.0)
	assertAlmostEqual(t, "slope", reg.Slope(), 2.0, 13)
	assertAlmostEqual(t, "intercept", reg.Intercept(), 1.0, 13)
	assertAlmostEqual(t, "correlation", reg.Correlation(), 1.0, 13)
}

func TestLinearRegressionKleinKBN_RevertMostRecent(t *testing.T) {
	t.Parallel()

	xs := append(append([]float64{}, baconBenchmark...), 0.5)
	ys := append(append([]float64{}, bacon...), -0.5)
	reg := feedReg(NewLinearRegressionKleinKBN(), xs, ys)
	reg.Revert(0.5, -0.5)
	assertEqualInt(t, "n", reg.N(), len(bacon))
	assertAlmostEqual(t, "slope", reg.Slope(), regSlope, 13)
	assertAlmostEqual(t, "intercept", reg.Intercept(), regIntercept, 14)
	assertAlmostEqual(t, "correlation", reg.Correlation(), regCorrelation, 13)
	assertAlmostEqual(t, "covariance", reg.Covariance(), regCovariance, 15)
}

func TestLinearRegressionKleinKBN_RevertOldest(t *testing.T) {
	t.Parallel()

	xs := append([]float64{0.5}, baconBenchmark...)
	ys := append([]float64{-0.5}, bacon...)
	reg := feedReg(NewLinearRegressionKleinKBN(), xs, ys)
	reg.Revert(0.5, -0.5)
	assertAlmostEqual(t, "slope", reg.Slope(), regSlope, 13)
	assertAlmostEqual(t, "intercept", reg.Intercept(), regIntercept, 14)
	assertAlmostEqual(t, "correlation", reg.Correlation(), regCorrelation, 13)
	assertAlmostEqual(t, "covariance", reg.Covariance(), regCovariance, 15)
}

func TestLinearRegressionKleinKBN_RevertToSingle(t *testing.T) {
	t.Parallel()

	reg := NewLinearRegressionKleinKBN()
	reg.Update(1.0, 2.0)
	reg.Update(3.0, 4.0)
	reg.Revert(3.0, 4.0)
	assertEqualInt(t, "n", reg.N(), 1)
	assertAlmostEqual(t, "mean_x", reg.MeanX(), 1.0, 15)
	assertAlmostEqual(t, "mean_y", reg.MeanY(), 2.0, 15)
	assertAlmostEqual(t, "co_moment", reg.CoMoment(), 0.0, 15)
	assertRegAllNaN(t, reg)
}

func TestLinearRegressionKleinKBN_RevertToEmpty(t *testing.T) {
	t.Parallel()

	reg := NewLinearRegressionKleinKBN()
	reg.Update(1.0, 2.0)
	reg.Revert(1.0, 2.0)
	assertEqualInt(t, "n", reg.N(), 0)
	assertRegAllNaN(t, reg)
}

func TestLinearRegressionKleinKBN_RevertEmptyRaises(t *testing.T) {
	t.Parallel()

	reg := NewLinearRegressionKleinKBN()
	assertPanics(t, "Cannot revert from an empty regression", func() { reg.Revert(1.0, 2.0) })
}

func TestLinearRegressionKleinKBN_RollingWindow(t *testing.T) {
	t.Parallel()

	names := []string{"slope", "intercept", "correlation", "covariance", "co_moment",
		"variance_x", "variance_y"}
	const w = 6
	reg := NewLinearRegressionKleinKBN()
	for i := range baconBenchmark {
		reg.Update(baconBenchmark[i], bacon[i])
		if i >= w {
			reg.Revert(baconBenchmark[i-w], bacon[i-w])
		}
		ref := feedReg(NewLinearRegressionKleinKBN(), window(baconBenchmark, i, w), window(bacon, i, w))
		assertEqualInt(t, fmt.Sprintf("step=%d n", i), reg.N(), ref.N())
		for _, name := range names {
			get := regProps[name]
			assertNaNOrAlmostEqual(t, fmt.Sprintf("step=%d %s", i, name), get(reg), get(ref), 13)
		}
	}
}

func TestLinearRegressionKleinKBN_Reset(t *testing.T) {
	t.Parallel()

	reg := NewLinearRegressionKleinKBN()
	for i := range 5 {
		reg.Update(float64(i), float64(2*i+1))
	}
	reg.Reset()
	assertEqualInt(t, "n", reg.N(), 0)
	assertRegAllNaN(t, reg)
	reg.Update(0.0, 1.0)
	reg.Update(1.0, 3.0)
	assertAlmostEqual(t, "slope", reg.Slope(), 2.0, 13)
}
