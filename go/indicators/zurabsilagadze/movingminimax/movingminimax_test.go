//nolint:testpackage
package movingminimax

import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
)

const tolerance = 1e-9

// runLast feeds the price series and returns the last computed result while primed.
func runLast(inputs []float64, m, n, numExtrema int) result {
	mmm, _ := NewMovingMiniMax(&Params{M: m, N: n, NumExtrema: numExtrema})

	var last result

	for _, p := range inputs {
		r := mmm.update(p)
		if mmm.primed {
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

func checkLevels(t *testing.T, name string, actual []level, expected []extremum) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("%s: length mismatch (%d vs %d)", name, len(actual), len(expected))
	}

	for i := range expected {
		if math.Abs(actual[i].price-expected[i].price) > tolerance*math.Max(1.0, math.Abs(expected[i].price)) {
			t.Errorf("%s[%d].price: expected %v, got %v", name, i, expected[i].price, actual[i].price)
		}

		if actual[i].offset != expected[i].offset {
			t.Errorf("%s[%d].offset: expected %d, got %d", name, i, expected[i].offset, actual[i].offset)
		}

		if math.Abs(actual[i].strength-expected[i].strength) > tolerance*math.Max(1.0, math.Abs(expected[i].strength)) {
			t.Errorf("%s[%d].strength: expected %v, got %v", name, i, expected[i].strength, actual[i].strength)
		}
	}
}

func check(t *testing.T, name string, r result, expUp, expDown []float64, expRes, expSup []extremum) {
	t.Helper()

	if !r.valid {
		t.Fatalf("%s: no valid primed output", name)
	}

	checkSeries(t, name+" UP", r.upDist, expUp)
	checkSeries(t, name+" DOWN", r.downDist, expDown)
	checkLevels(t, name+" RES", r.resistances, expRes)
	checkLevels(t, name+" SUP", r.supports, expSup)
}

//nolint:funlen
func TestMovingMiniMaxData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		m, n, e           int
		up, down          []float64
		res, sup          []extremum
	}{
		{"m3_n50_e1", 3, 50, 1, expected_M3_N50_E1_Up, expected_M3_N50_E1_Down, expected_M3_N50_E1_Resistances, expected_M3_N50_E1_Supports},
		{"m3_n50_e3", 3, 50, 3, expected_M3_N50_E3_Up, expected_M3_N50_E3_Down, expected_M3_N50_E3_Resistances, expected_M3_N50_E3_Supports},
		{"m3_n100_e1", 3, 100, 1, expected_M3_N100_E1_Up, expected_M3_N100_E1_Down, expected_M3_N100_E1_Resistances, expected_M3_N100_E1_Supports},
		{"m3_n100_e3", 3, 100, 3, expected_M3_N100_E3_Up, expected_M3_N100_E3_Down, expected_M3_N100_E3_Resistances, expected_M3_N100_E3_Supports},
		{"m3_n252_e1", 3, 252, 1, expected_M3_N252_E1_Up, expected_M3_N252_E1_Down, expected_M3_N252_E1_Resistances, expected_M3_N252_E1_Supports},
		{"m3_n252_e3", 3, 252, 3, expected_M3_N252_E3_Up, expected_M3_N252_E3_Down, expected_M3_N252_E3_Resistances, expected_M3_N252_E3_Supports},
		{"m5_n50_e1", 5, 50, 1, expected_M5_N50_E1_Up, expected_M5_N50_E1_Down, expected_M5_N50_E1_Resistances, expected_M5_N50_E1_Supports},
		{"m5_n50_e3", 5, 50, 3, expected_M5_N50_E3_Up, expected_M5_N50_E3_Down, expected_M5_N50_E3_Resistances, expected_M5_N50_E3_Supports},
		{"m5_n100_e1", 5, 100, 1, expected_M5_N100_E1_Up, expected_M5_N100_E1_Down, expected_M5_N100_E1_Resistances, expected_M5_N100_E1_Supports},
		{"m5_n100_e3", 5, 100, 3, expected_M5_N100_E3_Up, expected_M5_N100_E3_Down, expected_M5_N100_E3_Resistances, expected_M5_N100_E3_Supports},
		{"m5_n252_e1", 5, 252, 1, expected_M5_N252_E1_Up, expected_M5_N252_E1_Down, expected_M5_N252_E1_Resistances, expected_M5_N252_E1_Supports},
		{"m5_n252_e3", 5, 252, 3, expected_M5_N252_E3_Up, expected_M5_N252_E3_Down, expected_M5_N252_E3_Resistances, expected_M5_N252_E3_Supports},
		{"m10_n50_e1", 10, 50, 1, expected_M10_N50_E1_Up, expected_M10_N50_E1_Down, expected_M10_N50_E1_Resistances, expected_M10_N50_E1_Supports},
		{"m10_n50_e3", 10, 50, 3, expected_M10_N50_E3_Up, expected_M10_N50_E3_Down, expected_M10_N50_E3_Resistances, expected_M10_N50_E3_Supports},
		{"m10_n100_e1", 10, 100, 1, expected_M10_N100_E1_Up, expected_M10_N100_E1_Down, expected_M10_N100_E1_Resistances, expected_M10_N100_E1_Supports},
		{"m10_n100_e3", 10, 100, 3, expected_M10_N100_E3_Up, expected_M10_N100_E3_Down, expected_M10_N100_E3_Resistances, expected_M10_N100_E3_Supports},
		{"m10_n252_e1", 10, 252, 1, expected_M10_N252_E1_Up, expected_M10_N252_E1_Down, expected_M10_N252_E1_Resistances, expected_M10_N252_E1_Supports},
		{"m10_n252_e3", 10, 252, 3, expected_M10_N252_E3_Up, expected_M10_N252_E3_Down, expected_M10_N252_E3_Resistances, expected_M10_N252_E3_Supports},
		{"m20_n50_e1", 20, 50, 1, expected_M20_N50_E1_Up, expected_M20_N50_E1_Down, expected_M20_N50_E1_Resistances, expected_M20_N50_E1_Supports},
		{"m20_n50_e3", 20, 50, 3, expected_M20_N50_E3_Up, expected_M20_N50_E3_Down, expected_M20_N50_E3_Resistances, expected_M20_N50_E3_Supports},
		{"m20_n100_e1", 20, 100, 1, expected_M20_N100_E1_Up, expected_M20_N100_E1_Down, expected_M20_N100_E1_Resistances, expected_M20_N100_E1_Supports},
		{"m20_n100_e3", 20, 100, 3, expected_M20_N100_E3_Up, expected_M20_N100_E3_Down, expected_M20_N100_E3_Resistances, expected_M20_N100_E3_Supports},
		{"m20_n252_e1", 20, 252, 1, expected_M20_N252_E1_Up, expected_M20_N252_E1_Down, expected_M20_N252_E1_Resistances, expected_M20_N252_E1_Supports},
		{"m20_n252_e3", 20, 252, 3, expected_M20_N252_E3_Up, expected_M20_N252_E3_Down, expected_M20_N252_E3_Resistances, expected_M20_N252_E3_Supports},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := runLast(testInput, tt.m, tt.n, tt.e)
			check(t, tt.name, r, tt.up, tt.down, tt.res, tt.sup)
		})
	}
}

func TestMovingMiniMaxLatestScalars(t *testing.T) {
	t.Parallel()

	r := runLast(testInput, 3, 50, 1)
	if math.Abs(r.up-r.upDist[len(r.upDist)-1]) > 1e-12 {
		t.Errorf("up != upDist[last]: %v vs %v", r.up, r.upDist[len(r.upDist)-1])
	}

	if math.Abs(r.down-r.downDist[len(r.downDist)-1]) > 1e-12 {
		t.Errorf("down != downDist[last]: %v vs %v", r.down, r.downDist[len(r.downDist)-1])
	}
}

func TestMovingMiniMaxMnemonic(t *testing.T) {
	t.Parallel()

	mmm, err := NewMovingMiniMax(DefaultParams())
	if err != nil {
		t.Fatal(err)
	}

	if mmm.mnemonic != "mmm(5,50,3)" {
		t.Errorf("mnemonic: expected mmm(5,50,3), got %s", mmm.mnemonic)
	}
}

func TestMovingMiniMaxMetadata(t *testing.T) {
	t.Parallel()

	mmm, _ := NewMovingMiniMax(DefaultParams())
	meta := mmm.Metadata()

	if meta.Identifier != core.MovingMiniMax {
		t.Errorf("identifier mismatch")
	}

	if len(meta.Outputs) != 6 {
		t.Errorf("expected 6 outputs, got %d", len(meta.Outputs))
	}
}

func TestMovingMiniMaxUpdateScalar(t *testing.T) {
	t.Parallel()

	mmm, _ := NewMovingMiniMax(&Params{M: 5, N: 50, NumExtrema: 3})

	var out core.Output

	tm := entities.Scalar{}
	for _, p := range testInput {
		tm.Value = p
		out = mmm.UpdateScalar(&tm)
	}

	slice := out
	if len(slice) != 6 {
		t.Fatalf("expected 6 outputs")
	}
}

func TestMovingMiniMaxInvalidParams(t *testing.T) {
	t.Parallel()

	if _, err := NewMovingMiniMax(&Params{M: -1, N: 50, NumExtrema: 3}); err == nil {
		t.Error("expected error for m < 1")
	}

	if _, err := NewMovingMiniMax(&Params{M: 5, N: 10, NumExtrema: 3}); err == nil {
		t.Error("expected error for n <= 2*m")
	}

	if _, err := NewMovingMiniMax(&Params{M: 5, N: 50, NumExtrema: -1}); err == nil {
		t.Error("expected error for num extrema < 1")
	}
}
