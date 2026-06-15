//nolint:testpackage
package quantumpricelevels

import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

const tolerance = 1e-9

// runLast feeds the price series and returns the last computed result while primed.
func runLast(inputs []float64, lookback, numLevels, numBins int, scaleFactor float64) result {
	if lookback == 0 {
		lookback = len(inputs) - 1
	}

	qpl, _ := NewQuantumPriceLevels(&Params{
		Lookback: lookback, NumLevels: numLevels, NumBins: numBins, ScaleFactor: scaleFactor,
	})

	var last result

	for _, p := range inputs {
		r := qpl.update(p)
		if qpl.primed {
			last = r
		}
	}

	return last
}

func checkSeries(t *testing.T, name string, actual, expected []float64) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("%s: length mismatch (%d vs %d)", name, len(actual), len(expected))
	}

	for i := 0; i < len(expected); i++ {
		delta := tolerance * math.Max(1.0, math.Abs(expected[i]))
		if math.Abs(actual[i]-expected[i]) > delta {
			t.Errorf("%s[%d]: expected %v, got %v", name, i, expected[i], actual[i])
		}
	}
}

func check(t *testing.T, name string, r result, expNQPR, expUpper, expLower []float64) {
	t.Helper()

	if !r.valid {
		t.Fatalf("%s: no valid primed output", name)
	}

	checkSeries(t, name+" NQPR", r.nqpr, expNQPR)
	checkSeries(t, name+" UPPER", r.resistances, expUpper)
	checkSeries(t, name+" LOWER", r.supports, expLower)
}

func TestQuantumPriceLevelsBatch(t *testing.T) {
	t.Parallel()

	check(t, "default", runLast(testInput, 0, 21, 100, 0.21), expectedNQPR, expectedUPPER, expectedLOWER)
	check(t, "F0_10", runLast(testInput, 0, 21, 100, 0.10), expectedNQPR_F0_10, expectedUPPER_F0_10, expectedLOWER_F0_10)
	check(t, "F0_42", runLast(testInput, 0, 21, 100, 0.42), expectedNQPR_F0_42, expectedUPPER_F0_42, expectedLOWER_F0_42)
	check(t, "B50", runLast(testInput, 0, 21, 50, 0.21), expectedNQPR_B50, expectedUPPER_B50, expectedLOWER_B50)
	check(t, "B50_F0_10", runLast(testInput, 0, 21, 50, 0.10), expectedNQPR_B50_F0_10, expectedUPPER_B50_F0_10, expectedLOWER_B50_F0_10)
	check(t, "B50_F0_42", runLast(testInput, 0, 21, 50, 0.42), expectedNQPR_B50_F0_42, expectedUPPER_B50_F0_42, expectedLOWER_B50_F0_42)
	check(t, "L5", runLast(testInput, 0, 5, 100, 0.21), expectedNQPR_L5, expectedUPPER_L5, expectedLOWER_L5)
	check(t, "L10", runLast(testInput, 0, 10, 100, 0.21), expectedNQPR_L10, expectedUPPER_L10, expectedLOWER_L10)
	check(t, "L10_B50_F0_42", runLast(testInput, 0, 10, 50, 0.42), expectedNQPR_L10_B50_F0_42, expectedUPPER_L10_B50_F0_42, expectedLOWER_L10_B50_F0_42)
	check(t, "2K", runLast(testInput2K, 0, 21, 100, 0.21), expectedNQPR_2K, expectedUPPER_2K, expectedLOWER_2K)
}

func TestQuantumPriceLevelsReferenceProjection(t *testing.T) {
	t.Parallel()

	r := runLast(testInput, 0, 21, 100, 0.21)
	if !r.valid {
		t.Fatal("no valid output")
	}

	project := func(ref float64) ([]float64, []float64) {
		up := make([]float64, len(r.nqpr))
		lo := make([]float64, len(r.nqpr))

		for i, m := range r.nqpr {
			up[i] = ref * m
			lo[i] = ref / m
		}

		return up, lo
	}

	checkSeries(t, "R50_0 NQPR", r.nqpr, expectedNQPR_R50_0)
	up, lo := project(50.0)
	checkSeries(t, "R50_0 UPPER", up, expectedUPPER_R50_0)
	checkSeries(t, "R50_0 LOWER", lo, expectedLOWER_R50_0)

	checkSeries(t, "R1000_0 NQPR", r.nqpr, expectedNQPR_R1000_0)
	up, lo = project(1000.0)
	checkSeries(t, "R1000_0 UPPER", up, expectedUPPER_R1000_0)
	checkSeries(t, "R1000_0 LOWER", lo, expectedLOWER_R1000_0)

	checkSeries(t, "R1_2345 NQPR", r.nqpr, expectedNQPR_R1_2345)
	up, lo = project(1.2345)
	checkSeries(t, "R1_2345 UPPER", up, expectedUPPER_R1_2345)
	checkSeries(t, "R1_2345 LOWER", lo, expectedLOWER_R1_2345)
}

func TestQuantumPriceLevelsStreaming(t *testing.T) {
	t.Parallel()

	check(t, "S100", runLast(testInput, 100, 21, 100, 0.21), expectedNQPR_S100, expectedUPPER_S100, expectedLOWER_S100)
	check(t, "S150_B50", runLast(testInput, 150, 21, 50, 0.21), expectedNQPR_S150_B50, expectedUPPER_S150_B50, expectedLOWER_S150_B50)
	check(t, "S200_F0_42", runLast(testInput, 200, 21, 100, 0.42), expectedNQPR_S200_F0_42, expectedUPPER_S200_F0_42, expectedLOWER_S200_F0_42)
}

func TestQuantumPriceLevelsScalars(t *testing.T) {
	t.Parallel()

	r := runLast(testInput, 0, 21, 100, 0.21)
	if math.Abs(r.lambda-9.739608012591481e-01) > 1e-9 {
		t.Errorf("lambda: expected %v, got %v", 9.739608012591481e-01, r.lambda)
	}

	if math.Abs(r.sigma-2.662021797593086e-02) > 1e-9 {
		t.Errorf("sigma: expected %v, got %v", 2.662021797593086e-02, r.sigma)
	}
}

func TestQuantumPriceLevelsMnemonic(t *testing.T) {
	t.Parallel()

	qpl, err := NewQuantumPriceLevels(DefaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if qpl.mnemonic != "qpl(2048,21,100,0.21)" {
		t.Errorf("mnemonic: expected 'qpl(2048,21,100,0.21)', got '%s'", qpl.mnemonic)
	}
}

func TestQuantumPriceLevelsMetadata(t *testing.T) {
	t.Parallel()

	qpl, _ := NewQuantumPriceLevels(DefaultParams())
	meta := qpl.Metadata()

	if meta.Identifier != core.QuantumPriceLevels {
		t.Errorf("identifier: expected QuantumPriceLevels, got %v", meta.Identifier)
	}

	if len(meta.Outputs) != 5 {
		t.Errorf("outputs: expected 5, got %d", len(meta.Outputs))
	}
}

func TestQuantumPriceLevelsUpdateScalar(t *testing.T) {
	t.Parallel()

	qpl, _ := NewQuantumPriceLevels(&Params{Lookback: 100})

	var out core.Output
	for _, p := range testInput {
		out = qpl.UpdateScalar(&entities.Scalar{Value: p})
	}

	if len(out) != 5 {
		t.Fatalf("outputs: expected 5, got %d", len(out))
	}

	lvls, ok := out[3].(*outputs.Levels)
	if !ok {
		t.Fatalf("output[3]: expected *outputs.Levels")
	}

	if len(lvls.Levels) != 21 {
		t.Errorf("resistances: expected 21 levels, got %d", len(lvls.Levels))
	}
}

func TestQuantumPriceLevelsInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *Params
	}{
		{"lookback too small", &Params{Lookback: 1, NumLevels: 21, NumBins: 100, ScaleFactor: 0.21}},
		{"num levels too small", &Params{Lookback: 100, NumLevels: -1, NumBins: 100, ScaleFactor: 0.21}},
		{"num bins too small", &Params{Lookback: 100, NumLevels: 21, NumBins: 1, ScaleFactor: 0.21}},
		{"scale factor non-positive", &Params{Lookback: 100, NumLevels: 21, NumBins: 100, ScaleFactor: -0.5}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewQuantumPriceLevels(tt.params); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}
