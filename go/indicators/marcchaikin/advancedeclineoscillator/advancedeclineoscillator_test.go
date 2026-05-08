//nolint:testpackage
package advancedeclineoscillator

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestAdvanceDeclineOscillatorEMA(t *testing.T) {
	t.Parallel()

	const digits = 2

	highs := testHighs()
	lows := testLows()
	closes := testCloses()
	volumes := testVolumes()
	expected := testExpectedEMA()
	count := len(highs)

	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
		FastLength:        3,
		SlowLength:        10,
		MovingAverageType: EMA,
	})
	if err != nil {
		t.Fatal(err)
	}

	// EMA with length 10 has lookback = 9. First 9 values are NaN.
	for i := 0; i < count; i++ {
		v := adosc.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])

		if i < 9 {
			if !math.IsNaN(v) {
				t.Errorf("[%d] expected NaN, got %v", i, v)
			}

			if adosc.IsPrimed() {
				t.Errorf("[%d] expected not primed", i)
			}

			continue
		}

		if math.IsNaN(v) {
			t.Errorf("[%d] expected non-NaN, got NaN", i)
			continue
		}

		if !adosc.IsPrimed() {
			t.Errorf("[%d] expected primed", i)
		}

		got := roundTo(v, digits)
		exp := roundTo(expected[i], digits)

		if got != exp {
			t.Errorf("[%d] expected %v, got %v", i, exp, got)
		}
	}
}

func TestAdvanceDeclineOscillatorSMA(t *testing.T) {
	t.Parallel()

	const digits = 2

	highs := testHighs()
	lows := testLows()
	closes := testCloses()
	volumes := testVolumes()
	expected := testExpectedSMA()
	count := len(highs)

	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
		FastLength:        3,
		SlowLength:        10,
		MovingAverageType: SMA,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < count; i++ {
		v := adosc.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])

		if i < 9 {
			if !math.IsNaN(v) {
				t.Errorf("[%d] expected NaN, got %v", i, v)
			}

			if adosc.IsPrimed() {
				t.Errorf("[%d] expected not primed", i)
			}

			continue
		}

		if math.IsNaN(v) {
			t.Errorf("[%d] expected non-NaN, got NaN", i)
			continue
		}

		if !adosc.IsPrimed() {
			t.Errorf("[%d] expected primed", i)
		}

		got := roundTo(v, digits)
		exp := roundTo(expected[i], digits)

		if got != exp {
			t.Errorf("[%d] expected %v, got %v", i, exp, got)
		}
	}
}

func TestAdvanceDeclineOscillatorTaLibSpotChecks(t *testing.T) {
	t.Parallel()

	const digits = 2

	highs := testHighs()
	lows := testLows()
	closes := testCloses()
	volumes := testVolumes()
	count := len(highs)

	// ADOSC(3,10) EMA spot checks.
	t.Run("ADOSC_3_10", func(t *testing.T) {
		t.Parallel()

		adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
			FastLength:        3,
			SlowLength:        10,
			MovingAverageType: EMA,
		})
		if err != nil {
			t.Fatal(err)
		}

		var values []float64
		for i := 0; i < count; i++ {
			v := adosc.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])
			values = append(values, v)
		}

		// TA-Lib: begIndex=9, so output[0] corresponds to index 9.
		spotChecks := []struct {
			index    int
			expected float64
		}{
			{9, 841238.33},
			{10, 2255663.07},
			{250, -526700.32},
			{251, -1139932.73},
		}

		for _, sc := range spotChecks {
			got := roundTo(values[sc.index], digits)
			exp := roundTo(sc.expected, digits)

			if got != exp {
				t.Errorf("spot check [%d]: expected %v, got %v", sc.index, exp, got)
			}
		}
	})

	// ADOSC(5,2) EMA spot checks.
	t.Run("ADOSC_5_2", func(t *testing.T) {
		t.Parallel()

		adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
			FastLength:        5,
			SlowLength:        2,
			MovingAverageType: EMA,
		})
		if err != nil {
			t.Fatal(err)
		}

		var values []float64
		for i := 0; i < count; i++ {
			v := adosc.UpdateHLCV(highs[i], lows[i], closes[i], volumes[i])
			values = append(values, v)
		}

		// begIndex=4, output[0] at index 4.
		got := roundTo(values[4], digits)
		exp := roundTo(585361.29, digits)

		if got != exp {
			t.Errorf("ADOSC(5,2) spot check [4]: expected %v, got %v", exp, got)
		}
	})
}

func TestAdvanceDeclineOscillatorUpdateBar(t *testing.T) {
	t.Parallel()

	const digits = 2

	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
		FastLength:        3,
		SlowLength:        10,
		MovingAverageType: EMA,
	})
	if err != nil {
		t.Fatal(err)
	}

	tm := testTime()

	highs := testHighs()
	lows := testLows()
	closes := testCloses()
	volumes := testVolumes()
	expected := testExpectedEMA()

	// Feed first 15 bars via UpdateBar. First 9 are NaN (lookback=9), then valid from index 9.
	for i := 0; i < 15; i++ {
		bar := &entities.Bar{
			Time:   tm,
			Open:   highs[i],
			High:   highs[i],
			Low:    lows[i],
			Close:  closes[i],
			Volume: volumes[i],
		}

		output := adosc.UpdateBar(bar)
		scalar := output[0].(entities.Scalar)

		if i < 9 {
			if !math.IsNaN(scalar.Value) {
				t.Errorf("[%d] bar: expected NaN, got %v", i, scalar.Value)
			}

			continue
		}

		got := roundTo(scalar.Value, digits)
		exp := roundTo(expected[i], digits)

		if got != exp {
			t.Errorf("[%d] bar: expected %v, got %v", i, exp, got)
		}
	}
}

func TestAdvanceDeclineOscillatorNaN(t *testing.T) {
	t.Parallel()

	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
		FastLength:        3,
		SlowLength:        10,
		MovingAverageType: EMA,
	})
	if err != nil {
		t.Fatal(err)
	}

	if v := adosc.Update(math.NaN()); !math.IsNaN(v) {
		t.Errorf("expected NaN, got %v", v)
	}

	if v := adosc.UpdateHLCV(math.NaN(), 1, 2, 3); !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN high, got %v", v)
	}

	if v := adosc.UpdateHLCV(1, math.NaN(), 2, 3); !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN low, got %v", v)
	}

	if v := adosc.UpdateHLCV(1, 2, math.NaN(), 3); !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN close, got %v", v)
	}

	if v := adosc.UpdateHLCV(1, 2, 3, math.NaN()); !math.IsNaN(v) {
		t.Errorf("expected NaN for NaN volume, got %v", v)
	}
}

func TestAdvanceDeclineOscillatorNotPrimedInitially(t *testing.T) {
	t.Parallel()

	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
		FastLength:        3,
		SlowLength:        10,
		MovingAverageType: EMA,
	})
	if err != nil {
		t.Fatal(err)
	}

	if adosc.IsPrimed() {
		t.Error("expected not primed initially")
	}

	if v := adosc.Update(math.NaN()); !math.IsNaN(v) {
		t.Errorf("expected NaN, got %v", v)
	}

	// Still not primed after NaN.
	if adosc.IsPrimed() {
		t.Error("expected not primed after NaN")
	}
}

func TestAdvanceDeclineOscillatorMetadata(t *testing.T) {
	t.Parallel()

	adosc, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
		FastLength:        3,
		SlowLength:        10,
		MovingAverageType: EMA,
	})
	if err != nil {
		t.Fatal(err)
	}

	meta := adosc.Metadata()

	if meta.Identifier != core.AdvanceDeclineOscillator {
		t.Errorf("expected identifier AdvanceDeclineOscillator, got %v", meta.Identifier)
	}

	if meta.Mnemonic != "adosc(EMA3/EMA10)" {
		t.Errorf("expected mnemonic 'adosc(EMA3/EMA10)', got '%v'", meta.Mnemonic)
	}

	if meta.Description != "Chaikin Advance-Decline Oscillator adosc(EMA3/EMA10)" {
		t.Errorf("expected description 'Chaikin Advance-Decline Oscillator adosc(EMA3/EMA10)', got '%v'", meta.Description)
	}

	if len(meta.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Kind != int(Value) {
		t.Errorf("expected output kind %d, got %d", Value, meta.Outputs[0].Kind)
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("expected output type Scalar, got %v", meta.Outputs[0].Shape)
	}
}

func TestAdvanceDeclineOscillatorInvalidParams(t *testing.T) {
	t.Parallel()

	// Fast length < 2.
	_, err := NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
		FastLength:        1,
		SlowLength:        10,
		MovingAverageType: EMA,
	})
	if err == nil {
		t.Error("expected error for fast length < 2")
	}

	// Slow length < 2.
	_, err = NewAdvanceDeclineOscillator(&AdvanceDeclineOscillatorParams{
		FastLength:        3,
		SlowLength:        1,
		MovingAverageType: EMA,
	})
	if err == nil {
		t.Error("expected error for slow length < 2")
	}
}
