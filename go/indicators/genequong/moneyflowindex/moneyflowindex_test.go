//nolint:testpackage
package moneyflowindex

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestMoneyFlowIndexWithVolume(t *testing.T) {
	t.Parallel()

	const digits = 9

	tp := testTypicalPrices()
	vol := testVolumes()
	expected := testExpectedMfi()
	count := len(tp)

	mfi, err := NewMoneyFlowIndex(&MoneyFlowIndexParams{Length: 14})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 14; i++ {
		v := mfi.UpdateWithVolume(tp[i], vol[i])
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}

		if mfi.IsPrimed() {
			t.Errorf("[%d] expected not primed", i)
		}
	}

	for i := 14; i < count; i++ {
		v := mfi.UpdateWithVolume(tp[i], vol[i])
		if math.IsNaN(v) {
			t.Errorf("[%d] expected non-NaN, got NaN", i)
			continue
		}

		if !mfi.IsPrimed() {
			t.Errorf("[%d] expected primed", i)
		}

		got := roundTo(v, digits)
		exp := roundTo(expected[i], digits)

		if got != exp {
			t.Errorf("[%d] expected %v, got %v", i, exp, got)
		}
	}
}

func TestMoneyFlowIndexVolume1(t *testing.T) {
	t.Parallel()

	const digits = 9

	tp := testTypicalPrices()
	expected := testExpectedMfiVolume1()
	count := len(tp)

	mfi, err := NewMoneyFlowIndex(&MoneyFlowIndexParams{Length: 14})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 14; i++ {
		v := mfi.Update(tp[i])
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	for i := 14; i < count; i++ {
		v := mfi.Update(tp[i])
		if math.IsNaN(v) {
			t.Errorf("[%d] expected non-NaN, got NaN", i)
			continue
		}

		got := roundTo(v, digits)
		exp := roundTo(expected[i], digits)

		if got != exp {
			t.Errorf("[%d] expected %v, got %v", i, exp, got)
		}
	}
}

func TestMoneyFlowIndexIsPrimed(t *testing.T) {
	t.Parallel()

	mfi, err := NewMoneyFlowIndex(&MoneyFlowIndexParams{Length: 5})
	if err != nil {
		t.Fatal(err)
	}

	if mfi.IsPrimed() {
		t.Error("expected not primed initially")
	}

	// Feed 6 samples (5+1): first stores previousSample, next 5 fill buffer.
	// Primed after 6th sample (i.e., when bufferCount reaches Length).
	for i := 1; i <= 5; i++ {
		mfi.Update(float64(i))
		if mfi.IsPrimed() {
			t.Errorf("[%d] expected not primed", i)
		}
	}

	mfi.Update(5)
	if !mfi.IsPrimed() {
		t.Error("expected primed after length+1 samples")
	}

	mfi.Update(6)
	if !mfi.IsPrimed() {
		t.Error("expected still primed")
	}
}

func TestMoneyFlowIndexNaN(t *testing.T) {
	t.Parallel()

	mfi, err := NewMoneyFlowIndex(&MoneyFlowIndexParams{Length: 5})
	if err != nil {
		t.Fatal(err)
	}

	v := mfi.Update(math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN sample, got %v", v)
	}

	v = mfi.UpdateWithVolume(1.0, math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN volume, got %v", v)
	}

	v = mfi.UpdateWithVolume(math.NaN(), math.NaN())
	if !math.IsNaN(v) {
		t.Errorf("expected NaN for both NaN, got %v", v)
	}
}

func TestMoneyFlowIndexMetadata(t *testing.T) {
	t.Parallel()

	mfi, err := NewMoneyFlowIndex(&MoneyFlowIndexParams{Length: 14})
	if err != nil {
		t.Fatal(err)
	}

	meta := mfi.Metadata()

	if meta.Identifier != core.MoneyFlowIndex {
		t.Errorf("expected identifier MoneyFlowIndex, got %v", meta.Identifier)
	}

	exp := "mfi(14, hlc/3)"
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

func TestMoneyFlowIndexUpdateScalar(t *testing.T) {
	t.Parallel()

	tp := testTypicalPrices()

	mfi, err := NewMoneyFlowIndex(&MoneyFlowIndexParams{Length: 14})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	// UpdateScalar uses volume=1 path.
	for i := 0; i < 14; i++ {
		scalar := &entities.Scalar{Time: tm, Value: tp[i]}
		out := mfi.UpdateScalar(scalar)

		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	scalar := &entities.Scalar{Time: tm, Value: tp[14]}
	out := mfi.UpdateScalar(scalar)

	v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
	if math.IsNaN(v) {
		t.Errorf("[14] expected non-NaN, got NaN")
	}
}

func TestMoneyFlowIndexUpdateBar(t *testing.T) {
	t.Parallel()

	const digits = 9

	// Input OHLCV from TA-Lib test_MF.xls.
	inputHigh := []float64{
		93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000, 96.250000,
		99.625000, 99.125000, 92.750000, 91.315000,
	}
	inputLow := []float64{
		90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000, 92.750000,
		96.315000, 96.030000, 88.815000, 86.750000,
	}
	inputClose := []float64{
		91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000, 96.125000,
		97.250000, 98.500000, 89.875000, 91.000000,
	}
	inputVolume := []float64{
		4077500, 4955900, 4775300, 4155300, 4593100, 3631300, 3382800, 4954200, 4500000, 3397500,
		4204500, 6321400, 10203600, 19043900, 11692000,
	}

	mfi, err := NewMoneyFlowIndex(&MoneyFlowIndexParams{Length: 14})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	for i := 0; i < 14; i++ {
		bar := &entities.Bar{
			Time: tm, High: inputHigh[i], Low: inputLow[i],
			Close: inputClose[i], Volume: inputVolume[i],
		}
		out := mfi.UpdateBar(bar)

		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN, got %v", i, v)
		}
	}

	// Index 14: first value with real volume via UpdateBar.
	bar := &entities.Bar{
		Time: tm, High: inputHigh[14], Low: inputLow[14],
		Close: inputClose[14], Volume: inputVolume[14],
	}
	out := mfi.UpdateBar(bar)

	v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
	if math.IsNaN(v) {
		t.Errorf("[14] expected non-NaN, got NaN")
	}

	expected := testExpectedMfi()
	got := roundTo(v, digits)
	exp := roundTo(expected[14], digits)

	if got != exp {
		t.Errorf("[14] expected %v, got %v", exp, got)
	}
}

func TestMoneyFlowIndexInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		length int
	}{
		{"length zero", 0},
		{"length negative", -8},
	}

	for _, tt := range tests {
		_, err := NewMoneyFlowIndex(&MoneyFlowIndexParams{
			Length: tt.length,
		})
		if err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}

func TestMoneyFlowIndexSmallSum(t *testing.T) {
	t.Parallel()

	// When sum < 1, MFI should be 0.
	mfi, err := NewMoneyFlowIndex(&MoneyFlowIndexParams{Length: 2})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		mfi.UpdateWithVolume(0.001, 0.5)
	}

	if !mfi.IsPrimed() {
		t.Error("expected primed")
	}

	v := mfi.UpdateWithVolume(0.001, 0.5)
	if v != 0 {
		t.Errorf("expected 0 for small sum, got %v", v)
	}
}
