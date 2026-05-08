//nolint:testpackage
package mesaadaptivemovingaverage

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

func TestMesaAdaptiveMovingAverageUpdate(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-9 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testMesaAdaptiveMovingAverageInput()

	const (
		lprimed = 26
		f       = 3
		s       = 39
	)

	t.Run("reference implementation: MAMA from test_mama_new.xls", func(t *testing.T) {
		t.Parallel()

		mama := testMesaAdaptiveMovingAverageCreateLength(f, s)
		exp := testMesaAdaptiveMovingAverageExpected()

		for i := range lprimed {
			checkNaN(i, mama.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := mama.Update(input[i])
			check(i, exp[i], act)
		}

		checkNaN(0, mama.Update(math.NaN()))
	})

	t.Run("reference implementation: FAMA from test_mama_new.xls", func(t *testing.T) {
		t.Parallel()

		mama := testMesaAdaptiveMovingAverageCreateLength(f, s)
		exp := testMesaAdaptiveMovingAverageExpectedFama()

		for i := range lprimed {
			checkNaN(i, mama.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			mama.Update(input[i])
			act := mama.fama
			check(i, exp[i], act)
		}
	})
}

func TestMesaAdaptiveMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen,cyclop
	t.Parallel()

	const (
		lprimed      = 26
		fast         = 3
		slow         = 39
		inp          = 3.
		expectedMama = 1.5
		expectedFama = 0.375
	)

	time := testMesaAdaptiveMovingAverageTime()
	check := func(expMama, expFama float64, act core.Output) {
		t.Helper()

		const outputLen = 3

		if len(act) != outputLen {
			t.Errorf("len(output) is incorrect: expected %v, actual %v", outputLen, len(act))
		}

		i := 0

		s0, ok := act[i].(entities.Scalar)
		if !ok {
			t.Error("output[0] is not a scalar")
		}

		i++

		s1, ok := act[i].(entities.Scalar)
		if !ok {
			t.Error("output[1] is not a scalar")
		}

		i++

		s2, ok := act[i].(outputs.Band)
		if !ok {
			t.Error("output[2] is not a band")
		}

		if s0.Time != time {
			t.Errorf("output[0] time is incorrect: expected %v, actual %v", time, s0.Time)
		}

		if s0.Value != expMama {
			t.Errorf("output[0] value is incorrect: expected %v, actual %v", expMama, s0.Value)
		}

		if s1.Time != time {
			t.Errorf("output[1] time is incorrect: expected %v, actual %v", time, s1.Time)
		}

		if s1.Value != expFama {
			t.Errorf("output[1] value is incorrect: expected %v, actual %v", expFama, s1.Value)
		}

		if s2.Time != time {
			t.Errorf("output[2] time is incorrect: expected %v, actual %v", time, s2.Time)
		}

		if s2.Upper != expMama {
			t.Errorf("output[2] upper value is incorrect: expected %v, actual %v", expMama, s2.Upper)
		}

		if s2.Lower != expFama {
			t.Errorf("output[2] lower value is incorrect: expected %v, actual %v", expFama, s2.Lower)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: time, Value: inp}
		mama := testMesaAdaptiveMovingAverageCreateLength(fast, slow)

		for range lprimed {
			mama.Update(0.)
		}

		check(expectedMama, expectedFama, mama.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		mama := testMesaAdaptiveMovingAverageCreateLength(fast, slow)

		for range lprimed {
			mama.Update(0.)
		}

		check(expectedMama, expectedFama, mama.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		mama := testMesaAdaptiveMovingAverageCreateLength(fast, slow)

		for range lprimed {
			mama.Update(0.)
		}

		check(expectedMama, expectedFama, mama.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		mama := testMesaAdaptiveMovingAverageCreateLength(fast, slow)

		for range lprimed {
			mama.Update(0.)
		}

		check(expectedMama, expectedFama, mama.UpdateTrade(&r))
	})
}

func TestMesaAdaptiveMovingAverageIsPrimed(t *testing.T) {
	t.Parallel()

	input := testMesaAdaptiveMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const (
		lprimed = 26
		f       = 3
		s       = 39
	)

	t.Run("fast len = 3, slow len = 39 (mama.xls)", func(t *testing.T) {
		t.Parallel()

		mama := testMesaAdaptiveMovingAverageCreateLength(f, s)

		check(0, false, mama.IsPrimed())

		for i := 0; i < lprimed; i++ {
			mama.Update(input[i])
			check(i+1, false, mama.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			mama.Update(input[i])
			check(i+1, true, mama.IsPrimed())
		}
	})
}

func TestMesaAdaptiveMovingAverageMetadata(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	checkInstance := func(act core.Metadata, mnemonic string) {
		const (
			outputLen = 3
			descr     = "Mesa adaptive moving average "
		)

		mnemonicFama := strings.ReplaceAll(mnemonic, "mama", "fama")
		mnemonicBand := strings.ReplaceAll(mnemonic, "mama", "mama-fama")

		check("Identifier", core.MesaAdaptiveMovingAverage, act.Identifier)
		check("Mnemonic", mnemonic, act.Mnemonic)
		check("Description", descr+mnemonic, act.Description)
		check("len(Outputs)", outputLen, len(act.Outputs))

		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", mnemonic, act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", descr+mnemonic, act.Outputs[0].Description)

		check("Outputs[1].Kind", int(Fama), act.Outputs[1].Kind)
		check("Outputs[1].Shape", shape.Scalar, act.Outputs[1].Shape)
		check("Outputs[1].Mnemonic", mnemonicFama, act.Outputs[1].Mnemonic)
		check("Outputs[1].Description", descr+mnemonicFama, act.Outputs[1].Description)

		check("Outputs[2].Kind", int(Band), act.Outputs[2].Kind)
		check("Outputs[2].Shape", shape.Band, act.Outputs[2].Shape)
		check("Outputs[2].Mnemonic", mnemonicBand, act.Outputs[2].Mnemonic)
		check("Outputs[2].Description", descr+mnemonicBand, act.Outputs[2].Description)
	}

	t.Run("(fast, slow) limit length = (2, 40)", func(t *testing.T) {
		t.Parallel()

		const (
			f = 2
			s = 40
		)

		mama := testMesaAdaptiveMovingAverageCreateLength(f, s)
		act := mama.Metadata()
		checkInstance(act, "mama(2, 40)")
	})

	t.Run("(fast, slow) α = (0.666666666, 0.064516129)", func(t *testing.T) {
		t.Parallel()

		const (
			f = 0.666666666
			s = 0.064516129
		)

		mama := testMesaAdaptiveMovingAverageCreateAlpha(f, s)
		act := mama.Metadata()
		checkInstance(act, "mama(0.667, 0.065)")
	})
}

func TestNewMesaAdaptiveMovingAverage(t *testing.T) { //nolint: funlen,maintidx
	t.Parallel()

	const (
		bc  entities.BarComponent   = entities.BarMedianPrice
		qc  entities.QuoteComponent = entities.QuoteMidPrice
		tc  entities.TradeComponent = entities.TradePrice
		two                         = 2

		errfl = "invalid mesa adaptive moving average parameters: fast limit length should be larger than 1"
		errsl = "invalid mesa adaptive moving average parameters: slow limit length should be larger than 1"
		errfa = "invalid mesa adaptive moving average parameters: fast limit smoothing factor should be in range [0, 1]"
		errsa = "invalid mesa adaptive moving average parameters: slow limit smoothing factor should be in range [0, 1]"
		errbc = "invalid mesa adaptive moving average parameters: 9999: unknown bar component"
		errqc = "invalid mesa adaptive moving average parameters: 9999: unknown quote component"
		errtc = "invalid mesa adaptive moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(mama *MesaAdaptiveMovingAverage,
		mnemonic string, lenFast int, lenSlow int, aFast float64, aSlow float64,
	) {
		if math.IsNaN(aFast) {
			aFast = two / float64(lenFast+1)
		}

		if math.IsNaN(aSlow) {
			aSlow = two / float64(lenSlow+1)
		}

		const descr = "Mesa adaptive moving average "

		mnemonicFama := strings.ReplaceAll(mnemonic, "mama", "fama")
		mnemonicBand := strings.ReplaceAll(mnemonic, "mama", "mama-fama")

		check("mnemonic", mnemonic, mama.mnemonic)
		check("description", descr+mnemonic, mama.description)
		check("mnemonicFama", mnemonicFama, mama.mnemonicFama)
		check("descriptionFama", descr+mnemonicFama, mama.descriptionFama)
		check("mnemonicBand", mnemonicBand, mama.mnemonicBand)
		check("descriptionBand", descr+mnemonicBand, mama.descriptionBand)
		check("primed", false, mama.primed)
		check("alphaFastLimit", aFast, mama.alphaFastLimit)
		check("alphaSlowLimit", aSlow, mama.alphaSlowLimit)
		check("previousPhase", 0., mama.previousPhase)
		check("mama", 0., mama.mama)
		check("fama", 0., mama.fama)
		check("htce != nil", true, mama.htce != nil)
		check("isPhaseCached", false, mama.isPhaseCached)
		check("primed", false, mama.primed)
		check("barFunc == nil", false, mama.barFunc == nil)
		check("quoteFunc == nil", false, mama.quoteFunc == nil)
		check("tradeFunc == nil", false, mama.tradeFunc == nil)
	}

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		const (
			f = 3
			s = 39
		)

		mama, err := NewMesaAdaptiveMovingAverageDefault()
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(3, 39)", f, s, math.NaN(), math.NaN())
	})

	t.Run("both limit lengths > 1, default estimator", func(t *testing.T) {
		t.Parallel()

		const (
			f = 3
			s = 39
		)

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: testMesaAdaptiveMovingAverageCreateCycleEstimatorParams(),
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(3, 39, hl/2)", f, s, math.NaN(), math.NaN())
	})

	t.Run("both limit lengths > 1, default estimator (different length)", func(t *testing.T) {
		t.Parallel()

		const (
			l = 3
			f = 2
			s = 40
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()
		cep.SmoothingLength = l

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(2, 40, hd(3, 0.200, 0.200), hl/2)", f, s, math.NaN(), math.NaN())
	})

	t.Run("both limit lengths > 1, default estimator (different α quad)", func(t *testing.T) {
		t.Parallel()

		const (
			f = 2
			s = 40
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()
		cep.AlphaEmaQuadratureInPhase = 0.567

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(2, 40, hd(4, 0.567, 0.200), hl/2)", f, s, math.NaN(), math.NaN())
	})

	t.Run("both limit lengths > 1, default estimator (different α period)", func(t *testing.T) {
		t.Parallel()

		const (
			f = 2
			s = 40
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()
		cep.AlphaEmaPeriod = 0.567

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(2, 40, hd(4, 0.200, 0.567), hl/2)", f, s, math.NaN(), math.NaN())
	})

	t.Run("both limit lengths > 1, homodyne discriminator unrolled", func(t *testing.T) {
		t.Parallel()

		const (
			f = 2
			s = 40
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminatorUnrolled,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(2, 40, hdu(4, 0.200, 0.200), hl/2)", f, s, math.NaN(), math.NaN())
	})

	t.Run("both limit lengths > 1, phase accumulator", func(t *testing.T) {
		t.Parallel()

		const (
			f = 2
			s = 40
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.PhaseAccumulator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(2, 40, pa(4, 0.200, 0.200), hl/2)", f, s, math.NaN(), math.NaN())
	})

	t.Run("both limit lengths > 1, dual differentiator", func(t *testing.T) {
		t.Parallel()

		const (
			f = 2
			s = 40
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.DualDifferentiator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(2, 40, dd(4, 0.200, 0.200), hl/2)", f, s, math.NaN(), math.NaN())
	})

	t.Run("fast limit length = 1, error", func(t *testing.T) {
		t.Parallel()

		const (
			f = 1
			s = 39
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("mama == nil", true, mama == nil)
		check("err", errfl, err.Error())
	})

	t.Run("fast limit length = 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			f = 0
			s = 39
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("mama == nil", true, mama == nil)
		check("err", errfl, err.Error())
	})

	t.Run("fast limit length < 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			f = -1
			s = 39
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("mama == nil", true, mama == nil)
		check("err", errfl, err.Error())
	})

	t.Run("slow limit length = 1, error", func(t *testing.T) {
		t.Parallel()

		const (
			f = 3
			s = 1
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("mama == nil", true, mama == nil)
		check("err", errsl, err.Error())
	})

	t.Run("slow limit length = 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			f = 3
			s = 0
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("mama == nil", true, mama == nil)
		check("err", errsl, err.Error())
	})

	t.Run("slow limit length < 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			f = 3
			s = -1
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("mama == nil", true, mama == nil)
		check("err", errsl, err.Error())
	})

	t.Run("both smoothing factors 0 ≤ α ≤ 1, default estimator", func(t *testing.T) {
		t.Parallel()

		const (
			f = 0.66666666
			s = 0.33333333
		)

		params := SmoothingFactorParams{
			FastLimitSmoothingFactor: f, SlowLimitSmoothingFactor: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: testMesaAdaptiveMovingAverageCreateCycleEstimatorParams(),
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(0.667, 0.333, hl/2)", 0, 0, f, s)
	})

	t.Run("both smoothing factors 0 ≤ α ≤ 1, default estimator (different length)", func(t *testing.T) {
		t.Parallel()

		const (
			l = 3
			f = 0.66666666
			s = 0.33333333
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()
		cep.SmoothingLength = l

		params := SmoothingFactorParams{
			FastLimitSmoothingFactor: f, SlowLimitSmoothingFactor: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(0.667, 0.333, hd(3, 0.200, 0.200), hl/2)", 0, 0, f, s)
	})

	t.Run("both smoothing factors 0 ≤ α ≤ 1, default estimator (different α quad)", func(t *testing.T) {
		t.Parallel()

		const (
			f = 0.66666666
			s = 0.33333333
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()
		cep.AlphaEmaQuadratureInPhase = 0.567

		params := SmoothingFactorParams{
			FastLimitSmoothingFactor: f, SlowLimitSmoothingFactor: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(0.667, 0.333, hd(4, 0.567, 0.200), hl/2)", 0, 0, f, s)
	})

	t.Run("both smoothing factors 0 ≤ α ≤ 1, default estimator (different α period)", func(t *testing.T) {
		t.Parallel()

		const (
			f = 0.66666666
			s = 0.33333333
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()
		cep.AlphaEmaPeriod = 0.567

		params := SmoothingFactorParams{
			FastLimitSmoothingFactor: f, SlowLimitSmoothingFactor: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(0.667, 0.333, hd(4, 0.200, 0.567), hl/2)", 0, 0, f, s)
	})

	t.Run("both smoothing factors 0 ≤ α ≤ 1, homodyne discriminator unrolled", func(t *testing.T) {
		t.Parallel()

		const (
			f = 0.66666666
			s = 0.33333333
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := SmoothingFactorParams{
			FastLimitSmoothingFactor: f, SlowLimitSmoothingFactor: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminatorUnrolled,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(0.667, 0.333, hdu(4, 0.200, 0.200), hl/2)", 0, 0, f, s)
	})

	t.Run("both smoothing factors 0 ≤ α ≤ 1, phase accumulator", func(t *testing.T) {
		t.Parallel()

		const (
			f = 0.66666666
			s = 0.33333333
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := SmoothingFactorParams{
			FastLimitSmoothingFactor: f, SlowLimitSmoothingFactor: s,
			EstimatorType:   hilberttransformer.PhaseAccumulator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(0.667, 0.333, pa(4, 0.200, 0.200), hl/2)", 0, 0, f, s)
	})

	t.Run("both smoothing factors 0 ≤ α ≤ 1, dual differentiator", func(t *testing.T) {
		t.Parallel()

		const (
			f = 0.66666666
			s = 0.33333333
		)

		cep := testMesaAdaptiveMovingAverageCreateCycleEstimatorParams()

		params := SmoothingFactorParams{
			FastLimitSmoothingFactor: f, SlowLimitSmoothingFactor: s,
			EstimatorType:   hilberttransformer.DualDifferentiator,
			EstimatorParams: cep,
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(mama, "mama(0.667, 0.333, dd(4, 0.200, 0.200), hl/2)", 0, 0, f, s)
	})

	t.Run("α < 0, fastest, error", func(t *testing.T) {
		t.Parallel()

		const (
			f = -0.00000001
			s = 0.33333333
		)

		params := SmoothingFactorParams{
			FastLimitSmoothingFactor: f, SlowLimitSmoothingFactor: s,
			EstimatorType:   hilberttransformer.DualDifferentiator,
			EstimatorParams: testMesaAdaptiveMovingAverageCreateCycleEstimatorParams(),
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageSmoothingFactor(&params)
		check("mama == nil", true, mama == nil)
		check("err", errfa, err.Error())
	})

	t.Run("α > 1, fastest, error", func(t *testing.T) {
		t.Parallel()

		const (
			f = 1.00000001
			s = 0.33333333
		)

		params := SmoothingFactorParams{
			FastLimitSmoothingFactor: f, SlowLimitSmoothingFactor: s,
			EstimatorType:   hilberttransformer.DualDifferentiator,
			EstimatorParams: testMesaAdaptiveMovingAverageCreateCycleEstimatorParams(),
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageSmoothingFactor(&params)
		check("mama == nil", true, mama == nil)
		check("err", errfa, err.Error())
	})

	t.Run("α < 0, slowest, error", func(t *testing.T) {
		t.Parallel()

		const (
			f = 0.66666666
			s = -0.00000001
		)

		params := SmoothingFactorParams{
			FastLimitSmoothingFactor: f, SlowLimitSmoothingFactor: s,
			EstimatorType:   hilberttransformer.DualDifferentiator,
			EstimatorParams: testMesaAdaptiveMovingAverageCreateCycleEstimatorParams(),
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageSmoothingFactor(&params)
		check("mama == nil", true, mama == nil)
		check("err", errsa, err.Error())
	})

	t.Run("α > 1, slowest, error", func(t *testing.T) {
		t.Parallel()

		const (
			f = 0.66666666
			s = 1.00000001
		)

		params := SmoothingFactorParams{
			FastLimitSmoothingFactor: f, SlowLimitSmoothingFactor: s,
			EstimatorType:   hilberttransformer.DualDifferentiator,
			EstimatorParams: testMesaAdaptiveMovingAverageCreateCycleEstimatorParams(),
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageSmoothingFactor(&params)
		check("mama == nil", true, mama == nil)
		check("err", errsa, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		const (
			f = 3
			s = 39
		)

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: testMesaAdaptiveMovingAverageCreateCycleEstimatorParams(),
			BarComponent:    entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("mama == nil", true, mama == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		const (
			f = 3
			s = 39
		)

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: testMesaAdaptiveMovingAverageCreateCycleEstimatorParams(),
			BarComponent:    bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("mama == nil", true, mama == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		const (
			f = 3
			s = 39
		)

		params := LengthParams{
			FastLimitLength: f, SlowLimitLength: s,
			EstimatorType:   hilberttransformer.HomodyneDiscriminator,
			EstimatorParams: testMesaAdaptiveMovingAverageCreateCycleEstimatorParams(),
			BarComponent:    bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		mama, err := NewMesaAdaptiveMovingAverageLength(&params)
		check("mama == nil", true, mama == nil)
		check("err", errtc, err.Error())
	})
}
