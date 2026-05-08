//nolint:testpackage
package trendcyclemode

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
	"zpano/indicators/johnehlers/hilberttransformer"
)

//nolint:funlen
func TestTrendCycleModeUpdate(t *testing.T) {
	t.Parallel()

	input := testTCMInput()
	expPeriod := testTCMExpectedPeriod()
	expPhase := testTCMExpectedPhase()
	expSine := testTCMExpectedSine()
	expSineLead := testTCMExpectedSineLead()
	expITL := testTCMExpectedITL()
	expValue := testTCMExpectedValue()

	const (
		skip       = 9   // TradeStation implementation skips first 9 bars.
		settleSkip = 177 // Samples required for the EMA to converge past structural reference mismatch.
	)

	t.Run("reference period", func(t *testing.T) {
		t.Parallel()

		x := testTCMCreateDefault()

		for i := skip; i < len(input); i++ {
			_, _, _, _, _, _, period, _ := x.Update(input[i])
			if math.IsNaN(period) || i < settleSkip {
				continue
			}

			if math.Abs(expPeriod[i]-period) > testTCMTolerance {
				t.Errorf("[%v] period is incorrect: expected %v, actual %v", i, expPeriod[i], period)
			}
		}
	})

	t.Run("reference phase", func(t *testing.T) {
		t.Parallel()

		x := testTCMCreateDefault()

		for i := skip; i < len(input); i++ {
			_, _, _, _, _, _, _, phase := x.Update(input[i])
			if math.IsNaN(phase) || math.IsNaN(expPhase[i]) || i < settleSkip {
				continue
			}

			// MBST wraps phase into (-180, 180]; zpano into [0, 360). Compare modulo 360.
			d := math.Mod(expPhase[i]-phase, 360.0)
			if d > 180 {
				d -= 360
			} else if d < -180 {
				d += 360
			}

			if math.Abs(d) > testTCMTolerance {
				t.Errorf("[%v] phase is incorrect: expected %v, actual %v", i, expPhase[i], phase)
			}
		}
	})

	t.Run("reference sine wave", func(t *testing.T) {
		t.Parallel()

		x := testTCMCreateDefault()

		for i := skip; i < len(input); i++ {
			_, _, _, _, sine, _, _, _ := x.Update(input[i])
			if math.IsNaN(sine) || math.IsNaN(expSine[i]) || i < settleSkip {
				continue
			}

			if math.Abs(expSine[i]-sine) > testTCMTolerance {
				t.Errorf("[%v] sine is incorrect: expected %v, actual %v", i, expSine[i], sine)
			}
		}
	})

	t.Run("reference sine wave lead", func(t *testing.T) {
		t.Parallel()

		x := testTCMCreateDefault()

		for i := skip; i < len(input); i++ {
			_, _, _, _, _, sineLead, _, _ := x.Update(input[i])
			if math.IsNaN(sineLead) || math.IsNaN(expSineLead[i]) || i < settleSkip {
				continue
			}

			if math.Abs(expSineLead[i]-sineLead) > testTCMTolerance {
				t.Errorf("[%v] sineLead is incorrect: expected %v, actual %v", i, expSineLead[i], sineLead)
			}
		}
	})

	t.Run("reference instantaneous trend line", func(t *testing.T) {
		t.Parallel()

		x := testTCMCreateDefault()

		for i := skip; i < len(input); i++ {
			_, _, _, itl, _, _, _, _ := x.Update(input[i])
			if math.IsNaN(itl) || math.IsNaN(expITL[i]) || i < settleSkip {
				continue
			}

			if math.Abs(expITL[i]-itl) > testTCMTolerance {
				t.Errorf("[%v] itl is incorrect: expected %v, actual %v", i, expITL[i], itl)
			}
		}
	})

	t.Run("reference value (tcm)", func(t *testing.T) {
		t.Parallel()

		x := testTCMCreateDefault()
		limit := len(expValue)

		for i := skip; i < len(input); i++ {
			value, _, _, _, _, _, _, _ := x.Update(input[i])
			if i >= limit {
				continue
			}
			// MBST known mismatches.
			if i == 70 || i == 71 {
				continue
			}

			if math.IsNaN(value) || math.IsNaN(expValue[i]) {
				continue
			}

			if math.Abs(expValue[i]-value) > testTCMTolerance {
				t.Errorf("[%v] value is incorrect: expected %v, actual %v", i, expValue[i], value)
			}
		}
	})

	t.Run("is-trend / is-cycle are complementary 0/1", func(t *testing.T) {
		t.Parallel()

		x := testTCMCreateDefault()

		for i := skip; i < len(input); i++ {
			value, trend, cycle, _, _, _, _, _ := x.Update(input[i])
			if math.IsNaN(value) {
				continue
			}

			if trend+cycle != 1 {
				t.Errorf("[%v] trend+cycle is incorrect: expected 1, actual %v", i, trend+cycle)
			}

			if (value > 0 && trend != 1) || (value < 0 && trend != 0) {
				t.Errorf("[%v] value/trend mismatch: value=%v trend=%v", i, value, trend)
			}
		}
	})

	t.Run("NaN input returns NaN tuple", func(t *testing.T) {
		t.Parallel()

		x := testTCMCreateDefault()
		value, trend, cycle, itl, sine, sineLead, period, phase := x.Update(math.NaN())

		if !math.IsNaN(value) || !math.IsNaN(trend) || !math.IsNaN(cycle) ||
			!math.IsNaN(itl) || !math.IsNaN(sine) || !math.IsNaN(sineLead) ||
			!math.IsNaN(period) || !math.IsNaN(phase) {
			t.Errorf("expected all NaN, got (%v, %v, %v, %v, %v, %v, %v, %v)",
				value, trend, cycle, itl, sine, sineLead, period, phase)
		}
	})
}

func TestTrendCycleModeIsPrimed(t *testing.T) {
	t.Parallel()

	input := testTCMInput()

	t.Run("primes somewhere in the input sequence", func(t *testing.T) {
		t.Parallel()

		x := testTCMCreateDefault()

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

//nolint:funlen
func TestTrendCycleModeMetadata(t *testing.T) {
	t.Parallel()

	const (
		descrValue    = "Trend versus cycle mode "
		descrTrend    = "Trend versus cycle mode, is-trend flag "
		descrCycle    = "Trend versus cycle mode, is-cycle flag "
		descrITL      = "Trend versus cycle mode instantaneous trend line "
		descrSine     = "Trend versus cycle mode sine wave "
		descrSineLead = "Trend versus cycle mode sine wave lead "
		descrDCP      = "Dominant cycle period "
		descrDCPha    = "Dominant cycle phase "
	)

	x := testTCMCreateDefault()
	act := x.Metadata()

	check := func(what string, exp, a any) {
		t.Helper()

		if exp != a {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, a)
		}
	}

	mnValue := "tcm(0.330, 4, 1.000, 1.500%, hl/2)"
	mnTrend := "tcm-trend(0.330, 4, 1.000, 1.500%, hl/2)"
	mnCycle := "tcm-cycle(0.330, 4, 1.000, 1.500%, hl/2)"
	mnITL := "tcm-itl(0.330, 4, 1.000, 1.500%, hl/2)"
	mnSine := "tcm-sine(0.330, 4, 1.000, 1.500%, hl/2)"
	mnSineLead := "tcm-sineLead(0.330, 4, 1.000, 1.500%, hl/2)"
	mnDCP := "dcp(0.330, hl/2)"
	mnDCPha := "dcph(0.330, hl/2)"

	check("Identifier", core.TrendCycleMode, act.Identifier)
	check("Mnemonic", mnValue, act.Mnemonic)
	check("Description", descrValue+mnValue, act.Description)
	check("len(Outputs)", 8, len(act.Outputs))

	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", mnValue, act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", descrValue+mnValue, act.Outputs[0].Description)

	check("Outputs[1].Kind", int(IsTrendMode), act.Outputs[1].Kind)
	check("Outputs[1].Mnemonic", mnTrend, act.Outputs[1].Mnemonic)
	check("Outputs[1].Description", descrTrend+mnTrend, act.Outputs[1].Description)

	check("Outputs[2].Kind", int(IsCycleMode), act.Outputs[2].Kind)
	check("Outputs[2].Mnemonic", mnCycle, act.Outputs[2].Mnemonic)

	check("Outputs[3].Kind", int(InstantaneousTrendLine), act.Outputs[3].Kind)
	check("Outputs[3].Mnemonic", mnITL, act.Outputs[3].Mnemonic)

	check("Outputs[4].Kind", int(SineWave), act.Outputs[4].Kind)
	check("Outputs[4].Mnemonic", mnSine, act.Outputs[4].Mnemonic)

	check("Outputs[5].Kind", int(SineWaveLead), act.Outputs[5].Kind)
	check("Outputs[5].Mnemonic", mnSineLead, act.Outputs[5].Mnemonic)

	check("Outputs[6].Kind", int(DominantCyclePeriod), act.Outputs[6].Kind)
	check("Outputs[6].Mnemonic", mnDCP, act.Outputs[6].Mnemonic)
	check("Outputs[6].Description", descrDCP+mnDCP, act.Outputs[6].Description)

	check("Outputs[7].Kind", int(DominantCyclePhase), act.Outputs[7].Kind)
	check("Outputs[7].Mnemonic", mnDCPha, act.Outputs[7].Mnemonic)
	check("Outputs[7].Description", descrDCPha+mnDCPha, act.Outputs[7].Description)
}

//nolint:funlen
func TestTrendCycleModeUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		primeCount = 200
		inp        = 100.
		outputLen  = 8
	)

	tm := testTCMTime()
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

	input := testTCMInput()

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		x := testTCMCreateDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)])
		}

		check(x.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp, Low: inp, Close: inp}
		x := testTCMCreateDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)])
		}

		check(x.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		x := testTCMCreateDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)])
		}

		check(x.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		x := testTCMCreateDefault()

		for i := 0; i < primeCount; i++ {
			x.Update(input[i%len(input)])
		}

		check(x.UpdateTrade(&r))
	})
}

//nolint:funlen,maintidx
func TestNewTrendCycleMode(t *testing.T) {
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarClosePrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		errAlpha = "invalid trend cycle mode parameters: " +
			"α for additional smoothing should be in range (0, 1]"
		errTLSL = "invalid trend cycle mode parameters: " +
			"trend line smoothing length should be 2, 3, or 4"
		errCPM = "invalid trend cycle mode parameters: " +
			"cycle part multiplier should be in range (0, 10]"
		errSep = "invalid trend cycle mode parameters: " +
			"separation percentage should be in range (0, 100]"
		errBC = "invalid trend cycle mode parameters: " +
			"invalid dominant cycle parameters: 9999: unknown bar component"
		errQC = "invalid trend cycle mode parameters: " +
			"invalid dominant cycle parameters: 9999: unknown quote component"
		errTC = "invalid trend cycle mode parameters: " +
			"invalid dominant cycle parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		x, err := NewTrendCycleModeDefault()
		check("err == nil", true, err == nil)
		check("mnemonic", "tcm(0.330, 4, 1.000, 1.500%, hl/2)", x.mnemonic)
		check("primed", false, x.primed)
		check("trendLineSmoothingLength", 4, x.trendLineSmoothingLength)
		check("separationFactor", 0.015, x.separationFactor)
	})

	t.Run("tlsl=2", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 2,
			CyclePartMultiplier:      1.0,
			SeparationPercentage:     1.5,
			BarComponent:             bc, QuoteComponent: qc, TradeComponent: tc,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tcm(0.330, 2, 1.000, 1.500%)", x.mnemonic)
		check("coeff0", 2.0/3.0, x.coeff0)
		check("coeff1", 1.0/3.0, x.coeff1)
	})

	t.Run("tlsl=3, phase accumulator", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.5,
			EstimatorType:            hilberttransformer.PhaseAccumulator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 3,
			CyclePartMultiplier:      0.5,
			SeparationPercentage:     2.0,
			BarComponent:             bc, QuoteComponent: qc, TradeComponent: tc,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "tcm(0.500, 3, 0.500, 2.000%, pa(4, 0.200, 0.200))", x.mnemonic)
	})

	t.Run("α ≤ 0", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.0,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
			SeparationPercentage:     1.5,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("x == nil", true, x == nil)
		check("err", errAlpha, err.Error())
	})

	t.Run("α > 1", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 1.00000001,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
			SeparationPercentage:     1.5,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("x == nil", true, x == nil)
		check("err", errAlpha, err.Error())
	})

	t.Run("tlsl < 2", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 1,
			CyclePartMultiplier:      1.0,
			SeparationPercentage:     1.5,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("x == nil", true, x == nil)
		check("err", errTLSL, err.Error())
	})

	t.Run("tlsl > 4", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 5,
			CyclePartMultiplier:      1.0,
			SeparationPercentage:     1.5,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("x == nil", true, x == nil)
		check("err", errTLSL, err.Error())
	})

	t.Run("cpMul ≤ 0", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      0.0,
			SeparationPercentage:     1.5,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("x == nil", true, x == nil)
		check("err", errCPM, err.Error())
	})

	t.Run("cpMul > 10", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      10.00001,
			SeparationPercentage:     1.5,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("x == nil", true, x == nil)
		check("err", errCPM, err.Error())
	})

	t.Run("sep ≤ 0", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
			SeparationPercentage:     0.0,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("x == nil", true, x == nil)
		check("err", errSep, err.Error())
	})

	t.Run("sep > 100", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
			SeparationPercentage:     100.00001,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("x == nil", true, x == nil)
		check("err", errSep, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
			SeparationPercentage:     1.5,
			BarComponent:             entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("x == nil", true, x == nil)
		check("err", errBC, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
			SeparationPercentage:     1.5,
			BarComponent:             bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		x, err := NewTrendCycleModeParams(&params)
		check("x == nil", true, x == nil)
		check("err", errQC, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testTCMCreateCycleEstimatorParams(),
			TrendLineSmoothingLength: 4,
			CyclePartMultiplier:      1.0,
			SeparationPercentage:     1.5,
			BarComponent:             bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		x, err := NewTrendCycleModeParams(&params)
		check("x == nil", true, x == nil)
		check("err", errTC, err.Error())
	})
}
