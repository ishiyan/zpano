package streamingkbn

import (
	"math"
	"testing"
)

func TestLinearRegressionKleinKBN_PerfectFit(t *testing.T) {
	t.Parallel()

	r := NewLinearRegressionKleinKBN()
	for i := 0; i < 5; i++ {
		x := float64(i)
		r.Update(x, 2*x+1)
	}
	if !almostEqual(r.Slope(), 2.0, 1e-13) {
		t.Errorf("slope = %v, want 2.0", r.Slope())
	}
	if !almostEqual(r.Intercept(), 1.0, 1e-13) {
		t.Errorf("intercept = %v, want 1.0", r.Intercept())
	}
	if !almostEqual(r.Correlation(), 1.0, 1e-13) {
		t.Errorf("correlation = %v, want 1.0", r.Correlation())
	}
}

func TestLinearRegressionKleinKBN_ZeroCorrelation(t *testing.T) {
	t.Parallel()

	r := NewLinearRegressionKleinKBN()
	for i := 0; i < 5; i++ {
		r.Update(float64(i), 0.0)
	}
	if !almostEqual(r.Slope(), 0.0, 1e-13) {
		t.Errorf("slope = %v, want 0.0", r.Slope())
	}
	if !math.IsNaN(r.Correlation()) {
		t.Errorf("correlation = %v, want NaN", r.Correlation())
	}
}

func TestLinearRegressionKleinKBN_SinglePoint(t *testing.T) {
	t.Parallel()

	r := NewLinearRegressionKleinKBN()
	r.Update(1.0, 2.0)
	if !math.IsNaN(r.Slope()) {
		t.Errorf("slope = %v, want NaN", r.Slope())
	}
	if !math.IsNaN(r.Intercept()) {
		t.Errorf("intercept = %v, want NaN", r.Intercept())
	}
	if !math.IsNaN(r.Correlation()) {
		t.Errorf("correlation = %v, want NaN", r.Correlation())
	}
}

func TestLinearRegressionKleinKBN_TwoPoints(t *testing.T) {
	t.Parallel()

	r := NewLinearRegressionKleinKBN()
	r.Update(0.0, 1.0)
	r.Update(2.0, 5.0)
	if !almostEqual(r.Slope(), 2.0, 1e-13) {
		t.Errorf("slope = %v, want 2.0", r.Slope())
	}
	if !almostEqual(r.Intercept(), 1.0, 1e-13) {
		t.Errorf("intercept = %v, want 1.0", r.Intercept())
	}
	if !almostEqual(r.Correlation(), 1.0, 1e-13) {
		t.Errorf("correlation = %v, want 1.0", r.Correlation())
	}
}

func TestLinearRegressionKleinKBN_RevertMatchesSingleUpdate(t *testing.T) {
	t.Parallel()

	r := NewLinearRegressionKleinKBN()
	r.Update(1.0, 2.0)
	r.Update(3.0, 4.0)
	r.Revert(3.0, 4.0)

	ref := NewLinearRegressionKleinKBN()
	ref.Update(1.0, 2.0)

	if r.n != ref.n {
		t.Errorf("n = %d, want %d", r.n, ref.n)
	}
	if r.n != ref.n {
		t.Errorf("n = %d != %d", r.n, ref.n)
	}
	if !math.IsNaN(r.Slope()) {
		t.Errorf("slope = %v, want NaN", r.Slope())
	}
	if !math.IsNaN(ref.Slope()) {
		t.Errorf("ref.slope = %v, want NaN", ref.Slope())
	}
}

func TestLinearRegressionKleinKBN_RevertToEmpty(t *testing.T) {
	t.Parallel()

	r := NewLinearRegressionKleinKBN()
	r.Update(1.0, 2.0)
	r.Revert(1.0, 2.0)
	if r.n != 0 {
		t.Errorf("n = %d, want 0", r.n)
	}
	if !math.IsNaN(r.Slope()) {
		t.Errorf("slope = %v, want NaN", r.Slope())
	}
	if !math.IsNaN(r.Intercept()) {
		t.Errorf("intercept = %v, want NaN", r.Intercept())
	}
	if !math.IsNaN(r.Correlation()) {
		t.Errorf("correlation = %v, want NaN", r.Correlation())
	}
}

func TestLinearRegressionKleinKBN_RollingWindow(t *testing.T) {
	t.Parallel()

	data := [][2]float64{{0, 1}, {1, 3}, {2, 5}, {3, 7}, {4, 9}}
	r := NewLinearRegressionKleinKBN()
	for _, p := range data {
		r.Update(p[0], p[1])
	}

	r.Revert(data[0][0], data[0][1])
	r.Revert(data[1][0], data[1][1])
	r.Update(5.0, 11.0)
	r.Update(6.0, 13.0)

	ref := NewLinearRegressionKleinKBN()
	for _, p := range data[2:] {
		ref.Update(p[0], p[1])
	}
	ref.Update(5.0, 11.0)
	ref.Update(6.0, 13.0)

	if r.n != ref.n {
		t.Errorf("n = %d, want %d", r.n, ref.n)
	}
	if !almostEqual(r.Slope(), ref.Slope(), 1e-12) {
		t.Errorf("slope = %v, want %v", r.Slope(), ref.Slope())
	}
	if !almostEqual(r.Intercept(), ref.Intercept(), 1e-12) {
		t.Errorf("intercept = %v, want %v", r.Intercept(), ref.Intercept())
	}
	if !almostEqual(r.Correlation(), ref.Correlation(), 1e-12) {
		t.Errorf("correlation = %v, want %v", r.Correlation(), ref.Correlation())
	}
}

func TestLinearRegressionKleinKBN_NegativeCorrelation(t *testing.T) {
	t.Parallel()

	r := NewLinearRegressionKleinKBN()
	for i := 0; i < 5; i++ {
		x := float64(i)
		r.Update(x, -2*x+1)
	}
	if !almostEqual(r.Slope(), -2.0, 1e-13) {
		t.Errorf("slope = %v, want -2.0", r.Slope())
	}
	if !almostEqual(r.Intercept(), 1.0, 1e-13) {
		t.Errorf("intercept = %v, want 1.0", r.Intercept())
	}
	if !almostEqual(r.Correlation(), -1.0, 1e-13) {
		t.Errorf("correlation = %v, want -1.0", r.Correlation())
	}
}

func TestLinearRegressionKleinKBN_Reset(t *testing.T) {
	t.Parallel()

	r := NewLinearRegressionKleinKBN()
	for i := 0; i < 5; i++ {
		x := float64(i)
		r.Update(x, 2*x+1)
	}
	r.Reset()
	if r.n != 0 {
		t.Errorf("n = %d, want 0", r.n)
	}
	if !math.IsNaN(r.Slope()) {
		t.Errorf("slope = %v, want NaN", r.Slope())
	}
	if !math.IsNaN(r.Intercept()) {
		t.Errorf("intercept = %v, want NaN", r.Intercept())
	}
	if !math.IsNaN(r.Correlation()) {
		t.Errorf("correlation = %v, want NaN", r.Correlation())
	}

	r.Update(0.0, 1.0)
	r.Update(1.0, 3.0)
	if !almostEqual(r.Slope(), 2.0, 1e-13) {
		t.Errorf("slope = %v, want 2.0", r.Slope())
	}
}
