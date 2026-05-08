//nolint:testpackage
package hilberttransformerinstantaneoustrendline

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
	"zpano/indicators/johnehlers/hilberttransformer"
)

func TestHilbertTransformerInstantaneousTrendLineUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	input := testHTITLInput()
	expPeriod := testHTITLExpectedPeriod()
	expValue := testHTITLExpectedValue()

	const (
		skip       = 9   // TradeStation implementation skips first 9 bars.
		settleSkip = 177 // Samples required for the EMA to converge past structural reference mismatch.
	)

	t.Run("reference value (MBST InstantaneousTrendLineTest)", func(t *testing.T) {
		t.Parallel()

		x := testHTITLCreateDefault()

		for i := skip; i < len(input); i++ {
			value, _ := x.Update(input[i])
			if math.IsNaN(value) || i < settleSkip {
				continue
			}

			if math.IsNaN(expValue[i]) {
				continue
			}

			if math.Abs(expValue[i]-value) > testHTITLTolerance {
				t.Errorf("[%v] value is incorrect: expected %v, actual %v", i, expValue[i], value)
			}
		}
	})

	t.Run("reference period (test_MAMA.xsl, Period Adjustment)", func(t *testing.T) {
		t.Parallel()

		x := testHTITLCreateDefault()

		for i := skip; i < len(input); i++ {
			_, period := x.Update(input[i])
			if math.IsNaN(period) || i < settleSkip {
				continue
			}

			if math.Abs(expPeriod[i]-period) > testHTITLTolerance {
				t.Errorf("[%v] period is incorrect: expected %v, actual %v", i, expPeriod[i], period)
			}
		}
	})

	t.Run("NaN input returns NaN pair", func(t *testing.T) {
		t.Parallel()

		x := testHTITLCreateDefault()
		value, period := x.Update(math.NaN())

		if !math.IsNaN(value) || !math.IsNaN(period) {
			t.Errorf("expected (NaN, NaN), actual (%v, %v)", value, period)
		}
	})
}

func TestHilbertTransformerInstantaneousTrendLineIsPrimed(t *testing.T) {
	t.Parallel()

	input := testHTITLInput()

	t.Run("primes somewhere in the input sequence", func(t *testing.T) {
		t.Parallel()

		x := testHTITLCreateDefault()

		if x.IsPrimed() {
			t.Error("expected not primed at start")
		}

		primedAt := -1

		for i := range input {
			x.Update(input[i])

			if x.IsPrimed() && primedAt < 0 {
				primedAt = i
			}
		}

		if primedAt < 0 {
			t.Error("expected indicator to become primed within the input sequence")
		}

		if !x.IsPrimed() {
			t.Error("expected primed at end")
		}
	})
}

func TestHilbertTransformerInstantaneousTrendLineMetadata(t *testing.T) {
	t.Parallel()

	const (
		descrValue = "Hilbert transformer instantaneous trend line "
		descrDCP   = "Dominant cycle period "
	)

	x := testHTITLCreateDefault()
	act := x.Metadata()

	check := func(what string, exp, a any) {
		t.Helper()

		if exp != a {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, a)
		}
	}

	mnemonic := "htitl(0.330, 4, 1.000, hl/2)"
	mnemonicDCP := "dcp(0.330, hl/2)"

	check("Identifier", core.HilbertTransformerInstantaneousTrendLine, act.Identifier)
	check("Mnemonic", mnemonic, act.Mnemonic)
	check("Description", descrValue+mnemonic, act.Description)
	check("len(Outputs)", 2, len(act.Outputs))

	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mnemonic, act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", descrValue+mnemonic, act.Outputs[0].Description)

	check("Outputs[1].Kind", int(DominantCyclePeriod), act.Outputs[1].Kind)
	check("Outputs[1].Shape", shape.Scalar, act.Outputs[1].Shape)
	check("Outputs[1].Mnemonic", mnemonicDCP, act.Outputs[1].Mnemonic)
	check("Outputs[1].Description", descrDCP+mnemonicDCP, act.Outputs[1].Description)
}

func TestHilbertTransformerInstantaneousTrendLineUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		primeCount = 200
		inp        = 100.
		outputLen  = 2
	)

	tm := testHTITLTime()
	check := func(act core.Output) {
		t.Helper()

		if len(act) != outputLen {
			t.Errorf("len(output) is incorrect: expected %v, actual %v", outputLen, len(act))

			return
		}

		for i := 0; i < outputLen; i++ {
			s, ok := act[i].(entities.Scalar)
			if !ok {
				t.Errorf("output[%d] is not a scalar", i)

				continue
			}

			if s.Time != tm {
				t.Errorf("output[%d] time is incorrect: expected %v, actual %v", i, tm, s.Time)
			}
		}
	}

	input := testHTITLInput()

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		x := testHTITLCreateDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)])
		}

		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp, Low: inp, Close: inp}
		x := testHTITLCreateDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)])
		}

		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x := testHTITLCreateDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)])
		}

		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x := testHTITLCreateDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)])
		}

		check(x.UpdateTrade(&r))
	})
}

func TestNewHilbertTransformerInstantaneousTrendLine(t *testing.T) { //nolint: funlen,maintidx
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarClosePrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		errAlpha = "invalid hilbert transformer instantaneous trend line parameters: " +
			"α for additional smoothing should be in range (0, 1]"
		errTLSL = "invalid hilbert transformer instantaneous trend line parameters: " +
			"trend line smoothing length should be 2, 3, or 4"
		errCPM = "invalid hilbert transformer instantaneous trend line parameters: " +
			"cycle part multiplier should be in range (0, 10]"
		errBC = "invalid hilbert transformer instantaneous trend line parameters: 9999: unknown bar component"
		errQC = "invalid hilbert transformer instantaneous trend line parameters: 9999: unknown quote component"
		errTC = "invalid hilbert transformer instantaneous trend line parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		x, err := NewHilbertTransformerInstantaneousTrendLineDefault()
		check("err == nil", true, err == nil)
		check("mnemonic", "htitl(0.330, 4, 1.000, hl/2)", x.mnemonic)
		check("primed", false, x.primed)
		check("trendLineSmoothingLength", 4, x.trendLineSmoothingLength)
	})

	t.Run("tlsl=2", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testHTITLCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 2,
			CyclePartMultiplier:      1.0,
			BarComponent:             bc, QuoteComponent: qc, TradeComponent: tc,
		}

		x, err := NewHilbertTransformerInstantaneousTrendLineParams(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "htitl(0.330, 2, 1.000)", x.mnemonic)
		check("coeff0", 2.0/3.0, x.coeff0)
		check("coeff1", 1.0/3.0, x.coeff1)
	})

	t.Run("tlsl=3, phase accumulator", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.5,
			EstimatorType:            hilberttransformer.PhaseAccumulator,
			EstimatorParams:          testHTITLCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 3,
			CyclePartMultiplier:      0.5,
			BarComponent:             bc, QuoteComponent: qc, TradeComponent: tc,
		}

		x, err := NewHilbertTransformerInstantaneousTrendLineParams(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "htitl(0.500, 3, 0.500, pa(4, 0.200, 0.200))", x.mnemonic)
	})

	t.Run("α ≤ 0", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.0,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testHTITLCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
		}

		x, err := NewHilbertTransformerInstantaneousTrendLineParams(&params)
		check("x == nil", true, x == nil)
		check("err", errAlpha, err.Error())
	})

	t.Run("α > 1", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 1.00000001,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testHTITLCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
		}

		x, err := NewHilbertTransformerInstantaneousTrendLineParams(&params)
		check("x == nil", true, x == nil)
		check("err", errAlpha, err.Error())
	})

	t.Run("tlsl < 2", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testHTITLCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 1,
			CyclePartMultiplier:      1.0,
		}

		x, err := NewHilbertTransformerInstantaneousTrendLineParams(&params)
		check("x == nil", true, x == nil)
		check("err", errTLSL, err.Error())
	})

	t.Run("tlsl > 4", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testHTITLCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 5,
			CyclePartMultiplier:      1.0,
		}

		x, err := NewHilbertTransformerInstantaneousTrendLineParams(&params)
		check("x == nil", true, x == nil)
		check("err", errTLSL, err.Error())
	})

	t.Run("cpMul ≤ 0", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testHTITLCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      0.0,
		}

		x, err := NewHilbertTransformerInstantaneousTrendLineParams(&params)
		check("x == nil", true, x == nil)
		check("err", errCPM, err.Error())
	})

	t.Run("cpMul > 10", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testHTITLCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      10.00001,
		}

		x, err := NewHilbertTransformerInstantaneousTrendLineParams(&params)
		check("x == nil", true, x == nil)
		check("err", errCPM, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testHTITLCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
			BarComponent:             entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		x, err := NewHilbertTransformerInstantaneousTrendLineParams(&params)
		check("x == nil", true, x == nil)
		check("err", errBC, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testHTITLCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
			BarComponent:             bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		x, err := NewHilbertTransformerInstantaneousTrendLineParams(&params)
		check("x == nil", true, x == nil)
		check("err", errQC, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testHTITLCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
			BarComponent:             bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		x, err := NewHilbertTransformerInstantaneousTrendLineParams(&params)
		check("x == nil", true, x == nil)
		check("err", errTC, err.Error())
	})
}
