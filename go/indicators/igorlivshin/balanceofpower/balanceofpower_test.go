//nolint:testpackage
package balanceofpower

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

// First 20 entries from the C# TA-Lib test data.
func testOpen() []float64 {
	return []float64{
		92.500, 91.500, 95.155, 93.970, 95.500, 94.500, 95.000, 91.500, 91.815, 91.125,
		93.875, 97.500, 98.815, 92.000, 91.125, 91.875, 93.405, 89.750, 89.345, 92.250,
	}
}

func testHigh() []float64 {
	return []float64{
		93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000,
		96.250000, 99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000,
	}
}

func testLow() []float64 {
	return []float64{
		90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000,
		92.750000, 96.315000, 96.030000, 88.815000, 86.750000, 90.940000, 88.905000, 88.780000, 89.250000, 89.750000,
	}
}

func testClose() []float64 {
	return []float64{
		91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
		96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
	}
}

func testExpected() []float64 {
	return []float64{
		-0.400000000000000, 0.937765205091938, -0.367058823529412, 0.418215613382900, -0.540031397174254,
		0.102459016393443, -0.823333333333333, 0.314861460957179, -0.495049504950495, 0.632941176470588,
		0.642857142857143, -0.075528700906344, -0.101777059773828, -0.540025412960610, -0.027382256297919,
		0.406926406926406, -0.944444444444444, -0.216000000000001, 0.838235294117648, -0.950000000000000,
	}
}

func testTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func roundTo(v float64, digits int) float64 {
	p := math.Pow(10, float64(digits))
	return math.Round(v*p) / p
}

func TestBalanceOfPowerOHLC(t *testing.T) {
	t.Parallel()

	const digits = 9

	open := testOpen()
	high := testHigh()
	low := testLow()
	close := testClose()
	expected := testExpected()
	count := len(open)

	bop, err := NewBalanceOfPower(&BalanceOfPowerParams{})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < count; i++ {
		v := bop.UpdateOHLC(open[i], high[i], low[i], close[i])
		if math.IsNaN(v) {
			t.Errorf("[%d] expected non-NaN, got NaN", i)
			continue
		}

		if !bop.IsPrimed() {
			t.Errorf("[%d] expected primed", i)
		}

		got := roundTo(v, digits)
		exp := roundTo(expected[i], digits)

		if got != exp {
			t.Errorf("[%d] expected %v, got %v", i, exp, got)
		}
	}
}

func TestBalanceOfPowerIsPrimed(t *testing.T) {
	t.Parallel()

	bop, err := NewBalanceOfPower(&BalanceOfPowerParams{})
	if err != nil {
		t.Fatal(err)
	}

	// Always primed.
	if !bop.IsPrimed() {
		t.Error("expected primed initially")
	}

	bop.UpdateOHLC(92.5, 93.25, 90.75, 91.5)
	if !bop.IsPrimed() {
		t.Error("expected still primed after update")
	}
}

func TestBalanceOfPowerNaN(t *testing.T) {
	t.Parallel()

	bop, err := NewBalanceOfPower(&BalanceOfPowerParams{})
	if err != nil {
		t.Fatal(err)
	}

	v := bop.Update(math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN sample, got %v", v)
	}

	v = bop.UpdateOHLC(math.NaN(), 1.0, 2.0, 3.0)
	if !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN open, got %v", v)
	}

	v = bop.UpdateOHLC(1.0, math.NaN(), 2.0, 3.0)
	if !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN high, got %v", v)
	}

	v = bop.UpdateOHLC(1.0, 2.0, math.NaN(), 3.0)
	if !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN low, got %v", v)
	}

	v = bop.UpdateOHLC(1.0, 2.0, 3.0, math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN close, got %v", v)
	}
}

func TestBalanceOfPowerZeroRange(t *testing.T) {
	t.Parallel()

	bop, err := NewBalanceOfPower(&BalanceOfPowerParams{})
	if err != nil {
		t.Fatal(err)
	}

	// When H == L, range < epsilon, result is 0.
	v := bop.UpdateOHLC(0.001, 0.001, 0.001, 0.001)
	if v != 0 {
		t.Errorf("expected 0 for zero range, got %v", v)
	}
}

func TestBalanceOfPowerScalarAlwaysZero(t *testing.T) {
	t.Parallel()

	bop, err := NewBalanceOfPower(&BalanceOfPowerParams{})
	if err != nil {
		t.Fatal(err)
	}

	// Scalar update uses same value for O/H/L/C, so range=0, BOP=0.
	v := bop.Update(50.0)
	if v != 0 {
		t.Errorf("expected 0 for scalar update, got %v", v)
	}

	v = bop.Update(100.0)
	if v != 0 {
		t.Errorf("expected 0 for scalar update, got %v", v)
	}
}

func TestBalanceOfPowerMetadata(t *testing.T) {
	t.Parallel()

	bop, err := NewBalanceOfPower(&BalanceOfPowerParams{})
	if err != nil {
		t.Fatal(err)
	}

	meta := bop.Metadata()

	if meta.Identifier != core.BalanceOfPower {
		t.Errorf("expected identifier BalanceOfPower, got %v", meta.Identifier)
	}

	exp := "bop"
	if meta.Mnemonic != exp {
		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
	}

	if meta.Description != "Balance of Power" {
		t.Errorf("expected description 'Balance of Power', got '%s'", meta.Description)
	}

	if len(meta.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(Value) {
		t.Errorf("expected output kind %d, got %d", Value, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected scalar output type, got %v", meta.Outputs[0].Shape)
	}
}

func TestBalanceOfPowerUpdateBar(t *testing.T) {
	t.Parallel()

	const digits = 9

	open := testOpen()
	high := testHigh()
	low := testLow()
	close := testClose()
	expected := testExpected()

	bop, err := NewBalanceOfPower(&BalanceOfPowerParams{})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	for i := 0; i < len(open); i++ {
		bar := &entities.Bar{
			Time: tm, Open: open[i], High: high[i], Low: low[i], Close: close[i],
		}
		out := bop.UpdateBar(bar)

		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
		got := roundTo(v, digits)
		exp := roundTo(expected[i], digits)

		if got != exp {
			t.Errorf("[%d] expected %v, got %v", i, exp, got)
		}
	}
}
