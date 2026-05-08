//nolint:testpackage
package balanceofpower

import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

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
