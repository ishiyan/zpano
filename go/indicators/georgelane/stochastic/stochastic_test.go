//nolint:testpackage
package stochastic

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

// Test case 1: fastK=5, slowK=3/SMA, slowD=4/SMA.
// begIndex=9, SlowK[0]=38.139, SlowD[0]=36.725.
func TestStochastic_5_SMA3_SMA4_SingleValue(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-2

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()

	ind, err := NewStochastic(&StochasticParams{
		FastKLength: 5,
		SlowKLength: 3,
		SlowDLength: 4,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Feed first 9 bars (indices 0..8).
	for i := 0; i < 9; i++ {
		ind.Update(close[i], high[i], low[i])
	}

	// Index 9: first primed value.
	_, slowK, slowD := ind.Update(close[9], high[9], low[9])

	if math.Abs(slowK-38.139) > tolerance {
		t.Errorf("[9] SlowK: expected ~38.139, got %v", slowK)
	}

	if math.Abs(slowD-36.725) > tolerance {
		t.Errorf("[9] SlowD: expected ~36.725, got %v", slowD)
	}

	if !ind.IsPrimed() {
		t.Error("expected primed at index 9")
	}
}

// Test case 2: fastK=5, slowK=3/SMA, slowD=3/SMA.
// begIndex=8, first: SlowK[0]=24.0128, SlowD[0]=36.254.
func TestStochastic_5_SMA3_SMA3_FirstValue(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-2

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()

	ind, err := NewStochastic(&StochasticParams{
		FastKLength: 5,
		SlowKLength: 3,
		SlowDLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Feed first 8 bars (indices 0..7).
	for i := 0; i < 8; i++ {
		ind.Update(close[i], high[i], low[i])
	}

	// Index 8: first primed value.
	_, slowK, slowD := ind.Update(close[8], high[8], low[8])

	if math.Abs(slowK-24.0128) > tolerance {
		t.Errorf("[8] SlowK: expected ~24.0128, got %v", slowK)
	}

	if math.Abs(slowD-36.254) > tolerance {
		t.Errorf("[8] SlowD: expected ~36.254, got %v", slowD)
	}

	if !ind.IsPrimed() {
		t.Error("expected primed at index 8")
	}
}

// Test case 4: fastK=5, slowK=3/SMA, slowD=3/SMA.
// Last values: SlowK[243]=30.194, SlowD[243]=43.69.
func TestStochastic_5_SMA3_SMA3_LastValue(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-2

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()

	ind, err := NewStochastic(&StochasticParams{
		FastKLength: 5,
		SlowKLength: 3,
		SlowDLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	var slowK, slowD float64

	for i := 0; i < 252; i++ {
		_, slowK, slowD = ind.Update(close[i], high[i], low[i])
	}

	if math.Abs(slowK-30.194) > tolerance {
		t.Errorf("[251] SlowK: expected ~30.194, got %v", slowK)
	}

	if math.Abs(slowD-43.69) > tolerance {
		t.Errorf("[251] SlowD: expected ~43.69, got %v", slowD)
	}
}

// Test case 3: fastK=5, slowK=3/SMA, slowD=4/SMA.
// Last values (output index 242 = input index 251): SlowK=30.194, SlowD=46.641.
func TestStochastic_5_SMA3_SMA4_LastValue(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-2

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()

	ind, err := NewStochastic(&StochasticParams{
		FastKLength: 5,
		SlowKLength: 3,
		SlowDLength: 4,
	})
	if err != nil {
		t.Fatal(err)
	}

	var slowK, slowD float64

	for i := 0; i < 252; i++ {
		_, slowK, slowD = ind.Update(close[i], high[i], low[i])
	}

	if math.Abs(slowK-30.194) > tolerance {
		t.Errorf("[251] SlowK: expected ~30.194, got %v", slowK)
	}

	if math.Abs(slowD-46.641) > tolerance {
		t.Errorf("[251] SlowD: expected ~46.641, got %v", slowD)
	}
}

func TestStochasticIsPrimed(t *testing.T) {
	t.Parallel()

	ind, err := NewStochastic(&StochasticParams{
		FastKLength: 5,
		SlowKLength: 3,
		SlowDLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()

	// begIndex=8 for fastK=5, slowK=3/SMA, slowD=3/SMA.
	// lookback = (5-1) + (3-1) + (3-1) = 8.
	if ind.IsPrimed() {
		t.Error("expected not primed initially")
	}

	for i := 0; i < 8; i++ {
		ind.Update(close[i], high[i], low[i])
		if ind.IsPrimed() {
			t.Errorf("[%d] expected not primed", i)
		}
	}

	ind.Update(close[8], high[8], low[8])

	if !ind.IsPrimed() {
		t.Error("expected primed after index 8")
	}
}

func TestStochasticNaN(t *testing.T) {
	t.Parallel()

	ind, err := NewStochastic(&StochasticParams{
		FastKLength: 5,
		SlowKLength: 3,
		SlowDLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	fastK, slowK, slowD := ind.Update(math.NaN(), 1.0, 1.0)
	if !math.IsNaN(fastK) {
		t.Errorf("expected NaN FastK, got %v", fastK)
	}

	if !math.IsNaN(slowK) {
		t.Errorf("expected NaN SlowK, got %v", slowK)
	}

	if !math.IsNaN(slowD) {
		t.Errorf("expected NaN SlowD, got %v", slowD)
	}
}

func TestStochasticMetadata(t *testing.T) {
	t.Parallel()

	ind, err := NewStochastic(&StochasticParams{
		FastKLength: 5,
		SlowKLength: 3,
		SlowDLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := ind.Metadata()

	if meta.Identifier != core.Stochastic {
		t.Errorf("expected identifier Stochastic, got %v", meta.Identifier)
	}

	exp := "stoch(5/SMA3/SMA3)"
	if meta.Mnemonic != exp {
		t.Errorf("expected mnemonic '%s', got '%s'", exp, meta.Mnemonic)
	}

	if len(meta.Outputs) != 3 {
		t.Fatalf("expected 3 outputs, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(FastK) {
		t.Errorf("expected output 0 kind %d, got %d", FastK, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected scalar output type, got %v", meta.Outputs[0].Shape)
	}

	if meta.Outputs[1].Kind != int(SlowK) {
		t.Errorf("expected output 1 kind %d, got %d", SlowK, meta.Outputs[1].Kind)
	}

	if meta.Outputs[2].Kind != int(SlowD) {
		t.Errorf("expected output 2 kind %d, got %d", SlowD, meta.Outputs[2].Kind)
	}
}

func TestStochasticUpdateBar(t *testing.T) {
	t.Parallel()

	const tolerance = 1e-2

	high := testInputHigh()
	low := testInputLow()
	close := testInputClose()

	ind, err := NewStochastic(&StochasticParams{
		FastKLength: 5,
		SlowKLength: 3,
		SlowDLength: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	for i := 0; i < 8; i++ {
		bar := &entities.Bar{Time: tm, High: high[i], Low: low[i], Close: close[i]}
		out := ind.UpdateBar(bar)

		v := out[2].(entities.Scalar).Value //nolint:forcetypeassert
		if !math.IsNaN(v) {
			t.Errorf("[%d] expected NaN SlowD, got %v", i, v)
		}
	}

	bar := &entities.Bar{Time: tm, High: high[8], Low: low[8], Close: close[8]}
	out := ind.UpdateBar(bar)

	slowK := out[1].(entities.Scalar).Value //nolint:forcetypeassert
	slowD := out[2].(entities.Scalar).Value //nolint:forcetypeassert

	if math.Abs(slowK-24.0128) > tolerance {
		t.Errorf("[8] SlowK: expected ~24.0128, got %v", slowK)
	}

	if math.Abs(slowD-36.254) > tolerance {
		t.Errorf("[8] SlowD: expected ~36.254, got %v", slowD)
	}
}

func TestStochasticInvalidParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                                  string
		fastKLength, slowKLength, slowDLength int
	}{
		{"fastK too small", 0, 3, 3},
		{"slowK too small", 5, 0, 3},
		{"slowD too small", 5, 3, 0},
		{"fastK negative", -1, 3, 3},
	}

	for _, tt := range tests {
		_, err := NewStochastic(&StochasticParams{
			FastKLength: tt.fastKLength,
			SlowKLength: tt.slowKLength,
			SlowDLength: tt.slowDLength,
		})
		if err == nil {
			t.Errorf("%s: expected error, got nil", tt.name)
		}
	}
}
