//nolint:testpackage
package sinewave

//nolint: gofumpt
import (
	"math"
	"strings"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
	"zpano/indicators/johnehlers/hilberttransformer"
)

func TestSineWaveUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	input := testSineWaveInput()
	expPeriod := testSineWaveExpectedPeriod()
	expPhase := testSineWaveExpectedPhase()
	expSine := testSineWaveExpectedSine()
	expSineLead := testSineWaveExpectedSineLead()

	const (
		skip       = 9   // TradeStation implementation skips first 9 bars.
		settleSkip = 177 // Samples required for the EMA to converge past structural reference mismatch.
	)

	t.Run("reference sine (MBST SineWaveTest)", func(t *testing.T) {
		t.Parallel()

		sw := testSineWaveCreateDefault()

		for i := skip; i < len(input); i++ {
			value, _, _, _ := sw.Update(input[i])
			if math.IsNaN(value) || i < settleSkip {
				continue
			}

			if math.IsNaN(expSine[i]) {
				continue
			}

			if math.Abs(expSine[i]-value) > testSineWaveTolerance {
				t.Errorf("[%v] sine is incorrect: expected %v, actual %v", i, expSine[i], value)
			}
		}
	})

	t.Run("reference sine lead (MBST SineWaveLeadTest)", func(t *testing.T) {
		t.Parallel()

		sw := testSineWaveCreateDefault()

		for i := skip; i < len(input); i++ {
			_, lead, _, _ := sw.Update(input[i])
			if math.IsNaN(lead) || i < settleSkip {
				continue
			}

			if math.IsNaN(expSineLead[i]) {
				continue
			}

			if math.Abs(expSineLead[i]-lead) > testSineWaveTolerance {
				t.Errorf("[%v] sine lead is incorrect: expected %v, actual %v", i, expSineLead[i], lead)
			}
		}
	})

	t.Run("reference period (test_MAMA.xsl, Period Adjustment)", func(t *testing.T) {
		t.Parallel()

		sw := testSineWaveCreateDefault()

		for i := skip; i < len(input); i++ {
			_, _, period, _ := sw.Update(input[i])
			if math.IsNaN(period) || i < settleSkip {
				continue
			}

			if math.Abs(expPeriod[i]-period) > testSineWaveTolerance {
				t.Errorf("[%v] period is incorrect: expected %v, actual %v", i, expPeriod[i], period)
			}
		}
	})

	t.Run("reference phase (test_HT.xsl)", func(t *testing.T) {
		t.Parallel()

		sw := testSineWaveCreateDefault()

		for i := skip; i < len(input); i++ {
			_, _, _, phase := sw.Update(input[i])
			if math.IsNaN(phase) || i < settleSkip {
				continue
			}

			if math.IsNaN(expPhase[i]) {
				continue
			}

			if math.Abs(phaseDiff(expPhase[i], phase)) > testSineWaveTolerance {
				t.Errorf("[%v] phase is incorrect: expected %v, actual %v", i, expPhase[i], phase)
			}
		}
	})

	t.Run("NaN input returns NaN quadruple", func(t *testing.T) {
		t.Parallel()

		sw := testSineWaveCreateDefault()
		value, lead, period, phase := sw.Update(math.NaN())

		if !math.IsNaN(value) || !math.IsNaN(lead) || !math.IsNaN(period) || !math.IsNaN(phase) {
			t.Errorf("expected all NaN, actual (%v, %v, %v, %v)", value, lead, period, phase)
		}
	})
}

func TestSineWaveUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		primeCount = 200
		inp        = 100.
	)

	tm := testSineWaveTime()
	check := func(act core.Output) {
		t.Helper()

		const outputLen = 5

		if len(act) != outputLen {
			t.Errorf("len(output) is incorrect: expected %v, actual %v", outputLen, len(act))

			return
		}

		// Outputs 0, 1, 3, 4 are scalars; output 2 is a band.
		scalarIndex := []int{0, 1, 3, 4}
		for _, i := range scalarIndex {
			s, ok := act[i].(entities.Scalar)
			if !ok {
				t.Errorf("output[%d] is not a scalar", i)

				continue
			}

			if s.Time != tm {
				t.Errorf("output[%d] time is incorrect: expected %v, actual %v", i, tm, s.Time)
			}
		}

		b, ok := act[2].(outputs.Band)
		if !ok {
			t.Errorf("output[2] is not a band")

			return
		}

		if b.Time != tm {
			t.Errorf("output[2] time is incorrect: expected %v, actual %v", tm, b.Time)
		}
	}

	input := testSineWaveInput()

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: tm, Value: inp}
		sw := testSineWaveCreateDefault()

		for i := 0; i < primeCount; i++ {
			sw.Update(input[i%len(input)])
		}

		check(sw.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: tm, High: inp, Low: inp, Close: inp}
		sw := testSineWaveCreateDefault()

		for i := 0; i < primeCount; i++ {
			sw.Update(input[i%len(input)])
		}

		check(sw.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: tm, Bid: inp, Ask: inp}
		sw := testSineWaveCreateDefault()

		for i := 0; i < primeCount; i++ {
			sw.Update(input[i%len(input)])
		}

		check(sw.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: tm, Price: inp}
		sw := testSineWaveCreateDefault()

		for i := 0; i < primeCount; i++ {
			sw.Update(input[i%len(input)])
		}

		check(sw.UpdateTrade(&r))
	})
}

func TestSineWaveBandOrdering(t *testing.T) {
	t.Parallel()

	sw := testSineWaveCreateDefault()
	input := testSineWaveInput()

	tm := testSineWaveTime()
	s := entities.Scalar{Time: tm, Value: input[0]}

	// Prime the indicator.
	for i := 0; i < 200; i++ {
		sw.Update(input[i%len(input)])
	}

	out := sw.UpdateScalar(&s)

	value, ok := out[int(Value)-1].(entities.Scalar)
	if !ok {
		t.Fatalf("output[Value] is not a scalar")
	}

	lead, ok := out[int(Lead)-1].(entities.Scalar)
	if !ok {
		t.Fatalf("output[Lead] is not a scalar")
	}

	band, ok := out[int(Band)-1].(outputs.Band)
	if !ok {
		t.Fatalf("output[Band] is not a band")
	}

	if band.Upper != value.Value {
		t.Errorf("Band.Upper != Value: expected %v, actual %v", value.Value, band.Upper)
	}

	if band.Lower != lead.Value {
		t.Errorf("Band.Lower != Lead: expected %v, actual %v", lead.Value, band.Lower)
	}
}

func TestSineWaveIsPrimed(t *testing.T) {
	t.Parallel()

	input := testSineWaveInput()

	t.Run("primes somewhere in the first half of the sequence", func(t *testing.T) {
		t.Parallel()

		sw := testSineWaveCreateDefault()

		if sw.IsPrimed() {
			t.Error("expected not primed at start")
		}

		primedAt := -1

		for i := range input {
			sw.Update(input[i])

			if sw.IsPrimed() && primedAt < 0 {
				primedAt = i
			}
		}

		if primedAt < 0 {
			t.Error("expected indicator to become primed within the input sequence")
		}

		if !sw.IsPrimed() {
			t.Error("expected primed at end")
		}
	})
}

func TestSineWaveMetadata(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	checkInstance := func(act core.Metadata, mnemonic string) {
		const (
			outputLen  = 5
			descrValue = "Sine wave "
			descrLead  = "Sine wave lead "
			descrBand  = "Sine wave band "
			descrDCP   = "Dominant cycle period "
			descrDCPha = "Dominant cycle phase "
		)

		mnemonicLead := strings.ReplaceAll(mnemonic, "sw(", "sw-lead(")
		mnemonicBand := strings.ReplaceAll(mnemonic, "sw(", "sw-band(")
		mnemonicDCP := strings.ReplaceAll(mnemonic, "sw(", "dcp(")
		mnemonicDCPha := strings.ReplaceAll(mnemonic, "sw(", "dcph(")

		check("Identifier", core.SineWave, act.Identifier)
		check("Mnemonic", mnemonic, act.Mnemonic)
		check("Description", descrValue+mnemonic, act.Description)
		check("len(Outputs)", outputLen, len(act.Outputs))

		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", mnemonic, act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", descrValue+mnemonic, act.Outputs[0].Description)

		check("Outputs[1].Kind", int(Lead), act.Outputs[1].Kind)
		check("Outputs[1].Shape", shape.Scalar, act.Outputs[1].Shape)
		check("Outputs[1].Mnemonic", mnemonicLead, act.Outputs[1].Mnemonic)
		check("Outputs[1].Description", descrLead+mnemonicLead, act.Outputs[1].Description)

		check("Outputs[2].Kind", int(Band), act.Outputs[2].Kind)
		check("Outputs[2].Shape", shape.Band, act.Outputs[2].Shape)
		check("Outputs[2].Mnemonic", mnemonicBand, act.Outputs[2].Mnemonic)
		check("Outputs[2].Description", descrBand+mnemonicBand, act.Outputs[2].Description)

		check("Outputs[3].Kind", int(DominantCyclePeriod), act.Outputs[3].Kind)
		check("Outputs[3].Shape", shape.Scalar, act.Outputs[3].Shape)
		check("Outputs[3].Mnemonic", mnemonicDCP, act.Outputs[3].Mnemonic)
		check("Outputs[3].Description", descrDCP+mnemonicDCP, act.Outputs[3].Description)

		check("Outputs[4].Kind", int(DominantCyclePhase), act.Outputs[4].Kind)
		check("Outputs[4].Shape", shape.Scalar, act.Outputs[4].Shape)
		check("Outputs[4].Mnemonic", mnemonicDCPha, act.Outputs[4].Mnemonic)
		check("Outputs[4].Description", descrDCPha+mnemonicDCPha, act.Outputs[4].Description)
	}

	t.Run("default (α=0.33, BarMedianPrice)", func(t *testing.T) {
		t.Parallel()

		sw := testSineWaveCreateDefault()
		act := sw.Metadata()
		checkInstance(act, "sw(0.330, hl/2)")
	})

	t.Run("α=0.5, phase accumulator", func(t *testing.T) {
		t.Parallel()

		sw := testSineWaveCreateAlpha(0.5, hilberttransformer.PhaseAccumulator)
		act := sw.Metadata()
		checkInstance(act, "sw(0.500, pa(4, 0.200, 0.200), hl/2)")
	})
}

func TestNewSineWave(t *testing.T) { //nolint: funlen,maintidx
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarClosePrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		errAlpha = "invalid sine wave parameters: α for additional smoothing should be in range (0, 1]"
		errBC    = "invalid sine wave parameters: invalid dominant cycle parameters: 9999: unknown bar component"
		errQC    = "invalid sine wave parameters: invalid dominant cycle parameters: 9999: unknown quote component"
		errTC    = "invalid sine wave parameters: invalid dominant cycle parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(sw *SineWave, mnemonic string) {
		const (
			descrValue = "Sine wave "
			descrLead  = "Sine wave lead "
			descrBand  = "Sine wave band "
			descrDCP   = "Dominant cycle period "
			descrDCPha = "Dominant cycle phase "
		)

		mnemonicLead := strings.ReplaceAll(mnemonic, "sw(", "sw-lead(")
		mnemonicBand := strings.ReplaceAll(mnemonic, "sw(", "sw-band(")
		mnemonicDCP := strings.ReplaceAll(mnemonic, "sw(", "dcp(")
		mnemonicDCPha := strings.ReplaceAll(mnemonic, "sw(", "dcph(")

		check("mnemonic", mnemonic, sw.mnemonic)
		check("description", descrValue+mnemonic, sw.description)
		check("mnemonicLead", mnemonicLead, sw.mnemonicLead)
		check("descriptionLead", descrLead+mnemonicLead, sw.descriptionLead)
		check("mnemonicBand", mnemonicBand, sw.mnemonicBand)
		check("descriptionBand", descrBand+mnemonicBand, sw.descriptionBand)
		check("mnemonicDCP", mnemonicDCP, sw.mnemonicDCP)
		check("descriptionDCP", descrDCP+mnemonicDCP, sw.descriptionDCP)
		check("mnemonicDCPhase", mnemonicDCPha, sw.mnemonicDCPhase)
		check("descriptionDCPhase", descrDCPha+mnemonicDCPha, sw.descriptionDCPhase)
		check("primed", false, sw.primed)
		check("dc != nil", true, sw.dc != nil)
		check("barFunc == nil", false, sw.barFunc == nil)
		check("quoteFunc == nil", false, sw.quoteFunc == nil)
		check("tradeFunc == nil", false, sw.tradeFunc == nil)
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		sw, err := NewSineWaveDefault()
		check("err == nil", true, err == nil)
		checkInstance(sw, "sw(0.330, hl/2)")
	})

	t.Run("α=0.5, default estimator, custom components", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.5,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testSineWaveCreateCycleEstimatorParams(),
			BarComponent:             bc, QuoteComponent: qc, TradeComponent: tc,
		}

		sw, err := NewSineWaveParams(&params)
		check("err == nil", true, err == nil)
		checkInstance(sw, "sw(0.500)")
	})

	t.Run("α=0.5, phase accumulator", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.5,
			EstimatorType:            hilberttransformer.PhaseAccumulator,
			EstimatorParams:          testSineWaveCreateCycleEstimatorParams(),
			BarComponent:             bc, QuoteComponent: qc, TradeComponent: tc,
		}

		sw, err := NewSineWaveParams(&params)
		check("err == nil", true, err == nil)
		checkInstance(sw, "sw(0.500, pa(4, 0.200, 0.200))")
	})

	t.Run("α ≤ 0, error", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.0,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testSineWaveCreateCycleEstimatorParams(),
		}

		sw, err := NewSineWaveParams(&params)
		check("sw == nil", true, sw == nil)
		check("err", errAlpha, err.Error())
	})

	t.Run("α > 1, error", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 1.00000001,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testSineWaveCreateCycleEstimatorParams(),
		}

		sw, err := NewSineWaveParams(&params)
		check("sw == nil", true, sw == nil)
		check("err", errAlpha, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testSineWaveCreateCycleEstimatorParams(),
			BarComponent:             entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		sw, err := NewSineWaveParams(&params)
		check("sw == nil", true, sw == nil)
		check("err", errBC, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testSineWaveCreateCycleEstimatorParams(),
			BarComponent:             bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		sw, err := NewSineWaveParams(&params)
		check("sw == nil", true, sw == nil)
		check("err", errQC, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		params := Params{
			AlphaEmaPeriodAdditional: 0.33,
			EstimatorType:            hilberttransformer.HomodyneDiscriminator,
			EstimatorParams:          testSineWaveCreateCycleEstimatorParams(),
			BarComponent:             bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		sw, err := NewSineWaveParams(&params)
		check("sw == nil", true, sw == nil)
		check("err", errTC, err.Error())
	})
}
