//nolint:testpackage
package bollingerbands

//nolint:gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
)

func testTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

//nolint:funlen,cyclop
func TestBollingerBands_SampleStdDev_Length20_FullData(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	closingPrice := testClosingPrice()
	sma20 := testSma20Expected()
	expLower := testSampleLowerBandExpected()
	expUpper := testSampleUpperBandExpected()
	expBW := testSampleBandWidthExpected()
	expPctB := testSamplePercentBandExpected()

	boolTrue := true
	ind, err := NewBollingerBands(&BollingerBandsParams{
		Length:     20,
		IsUnbiased: &boolTrue,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 252; i++ {
		lower, middle, upper, bw, pctB := ind.Update(closingPrice[i])

		if math.IsNaN(sma20[i]) {
			if !math.IsNaN(lower) {
				t.Errorf("[%d] lower: expected NaN, got %v", i, lower)
			}

			if !math.IsNaN(middle) {
				t.Errorf("[%d] middle: expected NaN, got %v", i, middle)
			}

			if !math.IsNaN(upper) {
				t.Errorf("[%d] upper: expected NaN, got %v", i, upper)
			}

			continue
		}

		if math.Abs(middle-sma20[i]) > tolerance {
			t.Errorf("[%d] middle: expected %v, got %v", i, sma20[i], middle)
		}

		if math.Abs(lower-expLower[i]) > tolerance {
			t.Errorf("[%d] lower: expected %v, got %v", i, expLower[i], lower)
		}

		if math.Abs(upper-expUpper[i]) > tolerance {
			t.Errorf("[%d] upper: expected %v, got %v", i, expUpper[i], upper)
		}

		if math.Abs(bw-expBW[i]) > tolerance {
			t.Errorf("[%d] bandWidth: expected %v, got %v", i, expBW[i], bw)
		}

		if math.Abs(pctB-expPctB[i]) > tolerance {
			t.Errorf("[%d] percentBand: expected %v, got %v", i, expPctB[i], pctB)
		}
	}
}

//nolint:funlen,cyclop
func TestBollingerBands_PopulationStdDev_Length20_FullData(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	closingPrice := testClosingPrice()
	sma20 := testSma20Expected()
	expLower := testPopulationLowerBandExpected()
	expUpper := testPopulationUpperBandExpected()
	expBW := testPopulationBandWidthExpected()
	expPctB := testPopulationPercentBandExpected()

	boolFalse := false
	ind, err := NewBollingerBands(&BollingerBandsParams{
		Length:     20,
		IsUnbiased: &boolFalse,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 252; i++ {
		lower, middle, upper, bw, pctB := ind.Update(closingPrice[i])

		if math.IsNaN(sma20[i]) {
			if !math.IsNaN(lower) {
				t.Errorf("[%d] lower: expected NaN, got %v", i, lower)
			}

			if !math.IsNaN(middle) {
				t.Errorf("[%d] middle: expected NaN, got %v", i, middle)
			}

			if !math.IsNaN(upper) {
				t.Errorf("[%d] upper: expected NaN, got %v", i, upper)
			}

			continue
		}

		if math.Abs(middle-sma20[i]) > tolerance {
			t.Errorf("[%d] middle: expected %v, got %v", i, sma20[i], middle)
		}

		if math.Abs(lower-expLower[i]) > tolerance {
			t.Errorf("[%d] lower: expected %v, got %v", i, expLower[i], lower)
		}

		if math.Abs(upper-expUpper[i]) > tolerance {
			t.Errorf("[%d] upper: expected %v, got %v", i, expUpper[i], upper)
		}

		if math.Abs(bw-expBW[i]) > tolerance {
			t.Errorf("[%d] bandWidth: expected %v, got %v", i, expBW[i], bw)
		}

		if math.Abs(pctB-expPctB[i]) > tolerance {
			t.Errorf("[%d] percentBand: expected %v, got %v", i, expPctB[i], pctB)
		}
	}
}

func TestBollingerBandsIsPrimed(t *testing.T) {
	t.Parallel()

	boolTrue := true
	ind, err := NewBollingerBands(&BollingerBandsParams{
		Length:     20,
		IsUnbiased: &boolTrue,
	})
	if err != nil {
		t.Fatal(err)
	}

	closingPrice := testClosingPrice()

	if ind.IsPrimed() {
		t.Error("expected not primed initially")
	}

	for i := 0; i < 19; i++ {
		ind.Update(closingPrice[i])
		if ind.IsPrimed() {
			t.Errorf("[%d] expected not primed", i)
		}
	}

	ind.Update(closingPrice[19])

	if !ind.IsPrimed() {
		t.Error("expected primed after index 19")
	}
}

func TestBollingerBandsNaN(t *testing.T) {
	t.Parallel()

	ind, err := NewBollingerBands(&BollingerBandsParams{Length: 20})
	if err != nil {
		t.Fatal(err)
	}

	lower, middle, upper, bw, pctB := ind.Update(math.NaN())
	if !math.IsNaN(lower) {
		t.Errorf("expected NaN lower, got %v", lower)
	}

	if !math.IsNaN(middle) {
		t.Errorf("expected NaN middle, got %v", middle)
	}

	if !math.IsNaN(upper) {
		t.Errorf("expected NaN upper, got %v", upper)
	}

	if !math.IsNaN(bw) {
		t.Errorf("expected NaN bandWidth, got %v", bw)
	}

	if !math.IsNaN(pctB) {
		t.Errorf("expected NaN percentBand, got %v", pctB)
	}
}

func TestBollingerBandsMetadata(t *testing.T) {
	t.Parallel()

	ind, err := NewBollingerBands(&BollingerBandsParams{Length: 20})
	if err != nil {
		t.Fatal(err)
	}

	meta := ind.Metadata()

	if meta.Identifier != core.BollingerBands {
		t.Errorf("expected identifier BollingerBands, got %v", meta.Identifier)
	}

	const expectedOutputs = 6
	if len(meta.Outputs) != expectedOutputs {
		t.Fatalf("expected %d outputs, got %d", expectedOutputs, len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(Lower) {
		t.Errorf("expected output 0 kind %d, got %d", Lower, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected scalar output type, got %v", meta.Outputs[0].Shape)
	}

	if meta.Outputs[1].Kind != int(Middle) {
		t.Errorf("expected output 1 kind %d, got %d", Middle, meta.Outputs[1].Kind)
	}

	if meta.Outputs[5].Kind != int(Band) {
		t.Errorf("expected output 5 kind %d, got %d", Band, meta.Outputs[5].Kind)
	}

	if meta.Outputs[5].Shape != shape.Band {
		t.Errorf("expected band output type, got %v", meta.Outputs[5].Shape)
	}
}

//nolint:funlen
func TestBollingerBandsUpdateScalar(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-8

	closingPrice := testClosingPrice()
	sma20 := testSma20Expected()
	expLower := testSampleLowerBandExpected()
	expUpper := testSampleUpperBandExpected()

	boolTrue := true
	ind, err := NewBollingerBands(&BollingerBandsParams{
		Length:     20,
		IsUnbiased: &boolTrue,
	})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	// Feed first 19 samples — all NaN.
	for i := 0; i < 19; i++ {
		out := ind.UpdateScalar(&entities.Scalar{Time: tm, Value: closingPrice[i]})

		v := out[0].(entities.Scalar).Value //nolint:forcetypeassert
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN lower scalar, got %v", i, v)
		}

		band := out[5].(*outputs.Band) //nolint:forcetypeassert
		if !band.IsEmpty() {
			t.Errorf("[%d] expected empty band", i)
		}
	}

	// Feed index 19 — first primed value.
	out := ind.UpdateScalar(&entities.Scalar{Time: tm, Value: closingPrice[19]})

	lower := out[0].(entities.Scalar).Value  //nolint:forcetypeassert
	middle := out[1].(entities.Scalar).Value //nolint:forcetypeassert
	upper := out[2].(entities.Scalar).Value  //nolint:forcetypeassert

	if math.Abs(middle-sma20[19]) > tolerance {
		t.Errorf("[19] middle: expected %v, got %v", sma20[19], middle)
	}

	if math.Abs(lower-expLower[19]) > tolerance {
		t.Errorf("[19] lower: expected %v, got %v", expLower[19], lower)
	}

	if math.Abs(upper-expUpper[19]) > tolerance {
		t.Errorf("[19] upper: expected %v, got %v", expUpper[19], upper)
	}

	band := out[5].(*outputs.Band) //nolint:forcetypeassert
	if band.IsEmpty() {
		t.Error("[19] expected non-empty band")
	}

	if math.Abs(band.Lower-expLower[19]) > tolerance {
		t.Errorf("[19] band.Lower: expected %v, got %v", expLower[19], band.Lower)
	}

	if math.Abs(band.Upper-expUpper[19]) > tolerance {
		t.Errorf("[19] band.Upper: expected %v, got %v", expUpper[19], band.Upper)
	}
}

func TestBollingerBandsInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		length int
	}{
		{"length too small", 1},
		{"length negative", -1},
	}

	for _, tt := range tests {
		_, err := NewBollingerBands(&BollingerBandsParams{Length: tt.length})
		if err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}
