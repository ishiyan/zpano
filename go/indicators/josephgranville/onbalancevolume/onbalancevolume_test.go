//nolint:testpackage
package onbalancevolume

import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

// C# test data: 12 entries.
// Prices: [1,2,8,4,9,6,7,13,9,10,3,12]
// Volumes: [100,90,200,150,500,100,300,150,100,300,200,100]
// Expected: [100,190,390,240,740,640,940,1090,990,1290,1090,1190]

func testPrices() []float64 {
	return []float64{1, 2, 8, 4, 9, 6, 7, 13, 9, 10, 3, 12}
}

func testVolumes() []float64 {
	return []float64{100, 90, 200, 150, 500, 100, 300, 150, 100, 300, 200, 100}
}

func testExpected() []float64 {
	return []float64{100, 190, 390, 240, 740, 640, 940, 1090, 990, 1290, 1090, 1190}
}

func testTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func roundTo(v float64, digits int) float64 {
	p := math.Pow(10, float64(digits))
	return math.Round(v*p) / p
}

func TestOnBalanceVolumeWithVolume(t *testing.T) {
	t.Parallel()

	const digits = 1

	prices := testPrices()
	vol := testVolumes()
	expected := testExpected()
	count := len(prices)

	obv, err := NewOnBalanceVolume(&OnBalanceVolumeParams{})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < count; i++ {
		v := obv.UpdateWithVolume(prices[i], vol[i])
		if math.IsNaN(v) {
			t.Errorf("[%d] expected non-NaN, got NaN", i)
			continue
		}

		if !obv.IsPrimed() {
			t.Errorf("[%d] expected primed", i)
		}

		got := roundTo(v, digits)
		exp := roundTo(expected[i], digits)

		if got != exp {
			t.Errorf("[%d] expected %v, got %v", i, exp, got)
		}
	}
}

func TestOnBalanceVolumeIsPrimed(t *testing.T) {
	t.Parallel()

	obv, err := NewOnBalanceVolume(&OnBalanceVolumeParams{})
	if err != nil {
		t.Fatal(err)
	}

	if obv.IsPrimed() {
		t.Error("expected not primed initially")
	}

	obv.UpdateWithVolume(1.0, 100.0)
	if !obv.IsPrimed() {
		t.Error("expected primed after first update")
	}

	obv.UpdateWithVolume(2.0, 50.0)
	if !obv.IsPrimed() {
		t.Error("expected still primed")
	}
}

func TestOnBalanceVolumeNaN(t *testing.T) {
	t.Parallel()

	obv, err := NewOnBalanceVolume(&OnBalanceVolumeParams{})
	if err != nil {
		t.Fatal(err)
	}

	v := obv.Update(math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN sample, got %v", v)
	}

	v = obv.UpdateWithVolume(1.0, math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN volume, got %v", v)
	}

	v = obv.UpdateWithVolume(math.NaN(), math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN for both NaN, got %v", v)
	}
}

func TestOnBalanceVolumeMetadata(t *testing.T) {
	t.Parallel()

	obv, err := NewOnBalanceVolume(&OnBalanceVolumeParams{})
	if err != nil {
		t.Fatal(err)
	}

	meta := obv.Metadata()

	if meta.Identifier != core.OnBalanceVolume {
		t.Errorf("expected identifier OnBalanceVolume, got %v", meta.Identifier)
	}

	exp := "obv"
	if meta.Mnemonic != exp {
		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
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

func TestOnBalanceVolumeUpdateScalar(t *testing.T) {
	t.Parallel()

	obv, err := NewOnBalanceVolume(&OnBalanceVolumeParams{})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	// UpdateScalar uses volume=1 path.
	scalar := &entities.Scalar{Time: tm, Value: 10.0}
	out := obv.UpdateScalar(scalar)

	v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
	if math.IsNaN(v) {
		t.Error("expected non-NaN from UpdateScalar")
	}

	if v != 1.0 {
		t.Errorf("expected 1.0 (volume=1 on first call), got %v", v)
	}
}

func TestOnBalanceVolumeUpdateBar(t *testing.T) {
	t.Parallel()

	const digits = 1

	prices := testPrices()
	vol := testVolumes()
	expected := testExpected()

	obv, err := NewOnBalanceVolume(&OnBalanceVolumeParams{})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	for i := 0; i < len(prices); i++ {
		bar := &entities.Bar{
			Time: tm, Close: prices[i], Volume: vol[i],
		}
		out := obv.UpdateBar(bar)

		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
		got := roundTo(v, digits)
		exp := roundTo(expected[i], digits)

		if got != exp {
			t.Errorf("[%d] expected %v, got %v", i, exp, got)
		}
	}
}

func TestOnBalanceVolumeEqualPrices(t *testing.T) {
	t.Parallel()

	obv, err := NewOnBalanceVolume(&OnBalanceVolumeParams{})
	if err != nil {
		t.Fatal(err)
	}

	v := obv.UpdateWithVolume(10.0, 100.0)
	if v != 100.0 {
		t.Errorf("expected 100.0, got %v", v)
	}

	// Same price: value should not change.
	v = obv.UpdateWithVolume(10.0, 200.0)
	if v != 100.0 {
		t.Errorf("expected 100.0 (unchanged), got %v", v)
	}
}
