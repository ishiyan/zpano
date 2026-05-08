//nolint:testpackage
package arnaudlegouxmovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func testAlmaTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func testAlmaCreate(window int, sigma float64, offset float64) *ArnaudLegouxMovingAverage {
	alma, _ := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{
		Window: window,
		Sigma:  sigma,
		Offset: offset,
	})

	return alma
}

func testAlmaRun(t *testing.T, name string, window int, sigma float64, offset float64, expected []float64) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		t.Parallel()

		alma := testAlmaCreate(window, sigma, offset)

		warmup := window - 1
		if window == 1 {
			warmup = 0
		}

		for i := 0; i < warmup; i++ {
			act := alma.Update(testInput[i])
			if !math.IsNaN(act) {
				t.Errorf("[%v] expected NaN, got %v", i, act)
			}
		}

		for i := warmup; i < len(testInput); i++ {
			act := alma.Update(testInput[i])
			exp := expected[i]

			if math.IsNaN(exp) {
				if !math.IsNaN(act) {
					t.Errorf("[%v] expected NaN, got %v", i, act)
				}
			} else if math.Abs(act-exp) > 1e-13 {
				t.Errorf("[%v] expected %v, got %v", i, exp, act)
			}
		}

		// NaN passthrough.
		act := alma.Update(math.NaN())
		if !math.IsNaN(act) {
			t.Errorf("expected NaN passthrough, got %v", act)
		}
	})
}

func TestArnaudLegouxMovingAverageUpdate(t *testing.T) {
	t.Parallel()

	testAlmaRun(t, "w9 s6 o0.85 (default)", 9, 6.0, 0.85, expectedW9_S6_O0_85)
	testAlmaRun(t, "w9 s6 o0.5", 9, 6.0, 0.5, expectedW9_S6_O0_5)
	testAlmaRun(t, "w10 s6 o0.85", 10, 6.0, 0.85, expectedW10_S6_O0_85)
	testAlmaRun(t, "w5 s6 o0.9", 5, 6.0, 0.9, expectedW5_S6_O0_9)
	testAlmaRun(t, "w1 s6 o0.85", 1, 6.0, 0.85, expectedW1_S6_O0_85)
	testAlmaRun(t, "w3 s6 o0.85", 3, 6.0, 0.85, expectedW3_S6_O0_85)
	testAlmaRun(t, "w21 s6 o0.85", 21, 6.0, 0.85, expectedW21_S6_O0_85)
	testAlmaRun(t, "w50 s6 o0.85", 50, 6.0, 0.85, expectedW50_S6_O0_85)
	testAlmaRun(t, "w9 s6 o0", 9, 6.0, 0.0, expectedW9_S6_O0)
	testAlmaRun(t, "w9 s6 o1", 9, 6.0, 1.0, expectedW9_S6_O1)
	testAlmaRun(t, "w9 s2 o0.85", 9, 2.0, 0.85, expectedW9_S2_O0_85)
	testAlmaRun(t, "w9 s20 o0.85", 9, 20.0, 0.85, expectedW9_S20_O0_85)
	testAlmaRun(t, "w9 s0.5 o0.85", 9, 0.5, 0.85, expectedW9_S0_5_O0_85)
	testAlmaRun(t, "w15 s4 o0.7", 15, 4.0, 0.7, expectedW15_S4_O0_7)
}

func TestArnaudLegouxMovingAverageUpdateEntity(t *testing.T) {
	t.Parallel()

	tm := testAlmaTime()

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		alma := testAlmaCreate(1, 6.0, 0.85)
		s := entities.Scalar{Time: tm, Value: 7.0}
		out := alma.UpdateScalar(&s)

		if len(out) != 1 {
			t.Fatalf("expected 1 output, got %d", len(out))
		}

		sc, ok := out[0].(entities.Scalar)
		if !ok {
			t.Fatal("output is not scalar")
		}

		if sc.Value != 7.0 {
			t.Errorf("expected 7.0, got %v", sc.Value)
		}
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		alma := testAlmaCreate(1, 6.0, 0.85)
		b := entities.Bar{Time: tm, Open: 3.0, High: 3.0, Low: 3.0, Close: 5.0, Volume: 1.0}
		out := alma.UpdateBar(&b)

		if len(out) != 1 {
			t.Fatalf("expected 1 output, got %d", len(out))
		}

		sc, ok := out[0].(entities.Scalar)
		if !ok {
			t.Fatal("output is not scalar")
		}

		if sc.Value != 5.0 {
			t.Errorf("expected 5.0, got %v", sc.Value)
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		alma := testAlmaCreate(1, 6.0, 0.85)
		q := entities.Quote{Time: tm, Bid: 3.0, Ask: 5.0}
		out := alma.UpdateQuote(&q)

		if len(out) != 1 {
			t.Fatalf("expected 1 output, got %d", len(out))
		}

		sc, ok := out[0].(entities.Scalar)
		if !ok {
			t.Fatal("output is not scalar")
		}

		if sc.Value != 4.0 {
			t.Errorf("expected 4.0, got %v", sc.Value)
		}
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		alma := testAlmaCreate(1, 6.0, 0.85)
		r := entities.Trade{Time: tm, Price: 9.0}
		out := alma.UpdateTrade(&r)

		if len(out) != 1 {
			t.Fatalf("expected 1 output, got %d", len(out))
		}

		sc, ok := out[0].(entities.Scalar)
		if !ok {
			t.Fatal("output is not scalar")
		}

		if sc.Value != 9.0 {
			t.Errorf("expected 9.0, got %v", sc.Value)
		}
	})
}

func TestArnaudLegouxMovingAverageIsPrimed(t *testing.T) {
	t.Parallel()

	t.Run("window = 9", func(t *testing.T) {
		t.Parallel()

		alma := testAlmaCreate(9, 6.0, 0.85)

		if alma.IsPrimed() {
			t.Error("should not be primed initially")
		}

		for i := 0; i < 8; i++ {
			alma.Update(testInput[i])
			if alma.IsPrimed() {
				t.Errorf("should not be primed at index %d", i)
			}
		}

		alma.Update(testInput[8])
		if !alma.IsPrimed() {
			t.Error("should be primed after 9 updates")
		}
	})

	t.Run("window = 1", func(t *testing.T) {
		t.Parallel()

		alma := testAlmaCreate(1, 6.0, 0.85)

		if alma.IsPrimed() {
			t.Error("should not be primed initially")
		}

		alma.Update(testInput[0])
		if !alma.IsPrimed() {
			t.Error("should be primed after 1 update")
		}
	})
}

func TestArnaudLegouxMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	alma := testAlmaCreate(9, 6.0, 0.85)
	meta := alma.Metadata()

	if meta.Identifier != core.ArnaudLegouxMovingAverage {
		t.Errorf("identifier: expected %v, got %v", core.ArnaudLegouxMovingAverage, meta.Identifier)
	}

	if meta.Mnemonic != "alma(9, 6, 0.85)" {
		t.Errorf("mnemonic: expected alma(9, 6, 0.85), got %v", meta.Mnemonic)
	}

	if meta.Description != "Arnaud Legoux moving average alma(9, 6, 0.85)" {
		t.Errorf("description: expected Arnaud Legoux moving average alma(9, 6, 0.85), got %v", meta.Description)
	}

	if len(meta.Outputs) != 1 {
		t.Fatalf("expected 1 output, got %d", len(meta.Outputs))
	}

	if meta.Outputs[0].Shape != shape.Scalar {
		t.Errorf("output shape: expected %v, got %v", shape.Scalar, meta.Outputs[0].Shape)
	}

	if meta.Outputs[0].Mnemonic != "alma(9, 6, 0.85)" {
		t.Errorf("output mnemonic: expected alma(9, 6, 0.85), got %v", meta.Outputs[0].Mnemonic)
	}
}

func TestArnaudLegouxMovingAverageConstructionErrors(t *testing.T) {
	t.Parallel()

	t.Run("window = 0", func(t *testing.T) {
		t.Parallel()
		_, err := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{Window: 0, Sigma: 6.0, Offset: 0.85})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("window = -1", func(t *testing.T) {
		t.Parallel()
		_, err := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{Window: -1, Sigma: 6.0, Offset: 0.85})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("sigma = 0", func(t *testing.T) {
		t.Parallel()
		_, err := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{Window: 9, Sigma: 0, Offset: 0.85})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("sigma = -1", func(t *testing.T) {
		t.Parallel()
		_, err := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{Window: 9, Sigma: -1, Offset: 0.85})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("offset = -0.1", func(t *testing.T) {
		t.Parallel()
		_, err := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{Window: 9, Sigma: 6.0, Offset: -0.1})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("offset = 1.1", func(t *testing.T) {
		t.Parallel()
		_, err := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{Window: 9, Sigma: 6.0, Offset: 1.1})
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestArnaudLegouxMovingAverageMnemonics(t *testing.T) {
	t.Parallel()

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		alma, _ := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{Window: 9, Sigma: 6.0, Offset: 0.85})
		if alma.LineIndicator.Mnemonic != "alma(9, 6, 0.85)" {
			t.Errorf("mnemonic: expected alma(9, 6, 0.85), got %v", alma.LineIndicator.Mnemonic)
		}
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		alma, _ := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{
			Window: 9, Sigma: 6.0, Offset: 0.85, BarComponent: entities.BarMedianPrice,
		})
		if alma.LineIndicator.Mnemonic != "alma(9, 6, 0.85, hl/2)" {
			t.Errorf("mnemonic: expected alma(9, 6, 0.85, hl/2), got %v", alma.LineIndicator.Mnemonic)
		}
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		alma, _ := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{
			Window: 9, Sigma: 6.0, Offset: 0.85, QuoteComponent: entities.QuoteBidPrice,
		})
		if alma.LineIndicator.Mnemonic != "alma(9, 6, 0.85, b)" {
			t.Errorf("mnemonic: expected alma(9, 6, 0.85, b), got %v", alma.LineIndicator.Mnemonic)
		}
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		alma, _ := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{
			Window: 9, Sigma: 6.0, Offset: 0.85, TradeComponent: entities.TradeVolume,
		})
		if alma.LineIndicator.Mnemonic != "alma(9, 6, 0.85, v)" {
			t.Errorf("mnemonic: expected alma(9, 6, 0.85, v), got %v", alma.LineIndicator.Mnemonic)
		}
	})

	t.Run("bar and quote components set", func(t *testing.T) {
		t.Parallel()
		alma, _ := NewArnaudLegouxMovingAverage(&ArnaudLegouxMovingAverageParams{
			Window: 9, Sigma: 6.0, Offset: 0.85,
			BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteBidPrice,
		})
		if alma.LineIndicator.Mnemonic != "alma(9, 6, 0.85, o, b)" {
			t.Errorf("mnemonic: expected alma(9, 6, 0.85, o, b), got %v", alma.LineIndicator.Mnemonic)
		}
	})
}

func TestArnaudLegouxMovingAverageOutputString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		output Output
		want   string
	}{
		{Value, "value"},
		{outputLast, "unknown"},
		{Output(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.output.String(); got != tt.want {
			t.Errorf("Output(%d).String() = %q, want %q", tt.output, got, tt.want)
		}
	}
}

func TestArnaudLegouxMovingAverageOutputIsKnown(t *testing.T) {
	t.Parallel()

	if !Value.IsKnown() {
		t.Error("Value should be known")
	}

	if outputLast.IsKnown() {
		t.Error("outputLast should not be known")
	}

	if Output(99).IsKnown() {
		t.Error("Output(99) should not be known")
	}
}

func TestArnaudLegouxMovingAverageOutputMarshalJSON(t *testing.T) {
	t.Parallel()

	b, err := Value.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(b) != `"value"` {
		t.Errorf("expected \"value\", got %s", string(b))
	}

	_, err = outputLast.MarshalJSON()
	if err == nil {
		t.Error("expected error for unknown output")
	}
}

func TestArnaudLegouxMovingAverageOutputUnmarshalJSON(t *testing.T) {
	t.Parallel()

	var o Output

	err := o.UnmarshalJSON([]byte(`"value"`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if o != Value {
		t.Errorf("expected Value, got %v", o)
	}

	err = o.UnmarshalJSON([]byte(`"unknown"`))
	if err == nil {
		t.Error("expected error for unknown output")
	}
}
