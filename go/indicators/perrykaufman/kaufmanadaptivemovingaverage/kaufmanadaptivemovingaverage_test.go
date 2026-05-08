//nolint:testpackage
package kaufmanadaptivemovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestKaufmanAdaptiveMovingAverageUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-8 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testKaufmanAdaptiveMovingAverageInput()

	const (
		l       = 10
		lprimed = 10
		f       = 2
		s       = 30
	)

	t.Run("value, length = 10, fastest = 2, slowest = 30 (kama.xls)", func(t *testing.T) {
		t.Parallel()

		kama := testKaufmanAdaptiveMovingAverageCreateLength(l, f, s)

		exp := testKaufmanAdaptiveMovingAverageExpected()

		for i := range lprimed {
			checkNaN(i, kama.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := kama.Update(input[i])
			check(i, exp[i], act)
		}

		checkNaN(0, kama.Update(math.NaN()))
	})

	t.Run("efficiency ratio, length = 10 (kama.xls)", func(t *testing.T) {
		t.Parallel()

		kama := testKaufmanAdaptiveMovingAverageCreateLength(l, f, s)

		exp := testKaufmanAdaptiveMovingAverageExpectedEr()

		for i := range lprimed {
			checkNaN(i, kama.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			kama.Update(input[i])
			act := kama.efficiencyRatio
			check(i, exp[i], act)
		}

		checkNaN(0, kama.Update(math.NaN()))
	})
}

func TestKaufmanAdaptiveMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l        = 10
		lprimed  = 10
		fastest  = 2
		slowest  = 30
		inp      = 3.
		expected = 1.3333333333333328
	)

	time := testKaufmanAdaptiveMovingAverageTime()
	check := func(exp float64, act core.Output) {
		t.Helper()

		if len(act) != 1 {
			t.Errorf("len(output) is incorrect: expected 1, actual %v", len(act))
		}

		s, ok := act[0].(entities.Scalar)
		if !ok {
			t.Error("output is not scalar")
		}

		if s.Time != time {
			t.Errorf("time is incorrect: expected %v, actual %v", time, s.Time)
		}

		if s.Value != exp {
			t.Errorf("value is incorrect: expected %v, actual %v", exp, s.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: time, Value: inp}
		kama := testKaufmanAdaptiveMovingAverageCreateLength(l, fastest, slowest)

		for range lprimed {
			kama.Update(0.)
		}

		check(expected, kama.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		kama := testKaufmanAdaptiveMovingAverageCreateLength(l, fastest, slowest)

		for range lprimed {
			kama.Update(0.)
		}

		check(expected, kama.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		kama := testKaufmanAdaptiveMovingAverageCreateLength(l, fastest, slowest)

		for range lprimed {
			kama.Update(0.)
		}

		check(expected, kama.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		kama := testKaufmanAdaptiveMovingAverageCreateLength(l, fastest, slowest)

		for range lprimed {
			kama.Update(0.)
		}

		check(expected, kama.UpdateTrade(&r))
	})
}

func TestKaufmanAdaptiveMovingAverageIsPrimed(t *testing.T) {
	t.Parallel()

	input := testKaufmanAdaptiveMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const (
		l       = 10
		lprimed = 10
		f       = 2
		s       = 30
	)

	t.Run("length = 10, fastest = 2, slowest = 30 (kama.xls)", func(t *testing.T) {
		t.Parallel()

		kama := testKaufmanAdaptiveMovingAverageCreateLength(l, f, s)

		check(0, false, kama.IsPrimed())

		for i := range lprimed {
			kama.Update(input[i])
			check(i+1, false, kama.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			kama.Update(input[i])
			check(i+1, true, kama.IsPrimed())
		}
	})
}

func TestKaufmanAdaptiveMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("length = 10, fastest len = 2, slowest len = 30", func(t *testing.T) {
		t.Parallel()

		kama := testKaufmanAdaptiveMovingAverageCreateLength(10, 2, 30)
		act := kama.Metadata()

		check("Identifier", core.KaufmanAdaptiveMovingAverage, act.Identifier)
		check("Mnemonic", "kama(10, 2, 30)", act.Mnemonic)
		check("Description", "Kaufman adaptive moving average kama(10, 2, 30)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "kama(10, 2, 30)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Kaufman adaptive moving average kama(10, 2, 30)", act.Outputs[0].Description)
	})

	t.Run("length = 10, fastest α = 0.666666666, slowest α = 0.064516129", func(t *testing.T) {
		t.Parallel()

		const (
			l = 10
			f = 0.666666666
			s = 0.064516129
		)

		kama := testKaufmanAdaptiveMovingAverageCreateAlpha(l, f, s)
		act := kama.Metadata()

		check("Identifier", core.KaufmanAdaptiveMovingAverage, act.Identifier)
		check("Mnemonic", "kama(10, 0.6667, 0.0645)", act.Mnemonic)
		check("Description", "Kaufman adaptive moving average kama(10, 0.6667, 0.0645)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "kama(10, 0.6667, 0.0645)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Kaufman adaptive moving average kama(10, 0.6667, 0.0645)", act.Outputs[0].Description)
	})

	t.Run("length with non-default bar component", func(t *testing.T) {
		t.Parallel()
		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 10, FastestLength: 2, SlowestLength: 30,
			BarComponent: entities.BarMedianPrice,
		}

		kama, _ := NewKaufmanAdaptiveMovingAverageLength(&params)
		act := kama.Metadata()

		check("Mnemonic", "kama(10, 2, 30, hl/2)", act.Mnemonic)
		check("Description", "Kaufman adaptive moving average kama(10, 2, 30, hl/2)", act.Description)
	})

	t.Run("alpha with non-default quote component", func(t *testing.T) {
		t.Parallel()
		params := KaufmanAdaptiveMovingAverageSmoothingFactorParams{
			EfficiencyRatioLength: 10, FastestSmoothingFactor: 2. / 3., SlowestSmoothingFactor: 2. / 31.,
			QuoteComponent: entities.QuoteBidPrice,
		}

		kama, _ := NewKaufmanAdaptiveMovingAverageSmoothingFactor(&params)
		act := kama.Metadata()

		check("Mnemonic", "kama(10, 0.6667, 0.0645, b)", act.Mnemonic)
		check("Description", "Kaufman adaptive moving average kama(10, 0.6667, 0.0645, b)", act.Description)
	})
}

func TestNewKaufmanAdaptiveMovingAverage(t *testing.T) { //nolint: funlen, maintidx
	t.Parallel()

	const (
		bc  entities.BarComponent   = entities.BarMedianPrice
		qc  entities.QuoteComponent = entities.QuoteMidPrice
		tc  entities.TradeComponent = entities.TradePrice
		two                         = 2

		errelen = "invalid Kaufman adaptive moving average parameters: efficiency ratio length should be larger than 1"
		errflen = "invalid Kaufman adaptive moving average parameters: fastest smoothing length should be larger than 1"
		errslen = "invalid Kaufman adaptive moving average parameters: slowest smoothing length should be larger than 1"
		errfa   = "invalid Kaufman adaptive moving average parameters: fastest smoothing factor should be in range [0, 1]"
		errsa   = "invalid Kaufman adaptive moving average parameters: slowest smoothing factor should be in range [0, 1]"
		errbc   = "invalid Kaufman adaptive moving average parameters: 9999: unknown bar component"
		errqc   = "invalid Kaufman adaptive moving average parameters: 9999: unknown quote component"
		errtc   = "invalid Kaufman adaptive moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(kama *KaufmanAdaptiveMovingAverage,
		mnemonic string, length int, af float64, as float64,
	) {
		check("mnemonic", mnemonic, kama.LineIndicator.Mnemonic)
		check("description", "Kaufman adaptive moving average "+mnemonic, kama.LineIndicator.Description)
		check("primed", false, kama.primed)
		check("efficiencyRatioLength", length, kama.efficiencyRatioLength)
		check("alphaFastest", af, kama.alphaFastest)
		check("alphaSlowest", as, kama.alphaSlowest)
		check("alphaDiff", af-as, kama.alphaDiff)
		check("absoluteDeltaSum", 0., kama.absoluteDeltaSum)
		check("value", true, math.IsNaN(kama.value))
		check("efficiencyRatio", true, math.IsNaN(kama.efficiencyRatio))
		check("windowCount", 0, kama.windowCount)
	}

	t.Run("efficiency ratio length > 1, (fast,slow) len = (2,30)", func(t *testing.T) {
		t.Parallel()

		const (
			l = 10
			f = 2
			s = 30
		)

		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: l, FastestLength: f, SlowestLength: s,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(kama, "kama(10, 2, 30, hl/2)", l,
			float64(two)/float64(1+f), float64(two)/float64(1+s))
	})

	t.Run("efficiency ratio length = 1, error", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 1, FastestLength: 2, SlowestLength: 30,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("kama == nil", true, kama == nil)
		check("err", errelen, err.Error())
	})

	t.Run("efficiency ratio length = 0, error", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 0, FastestLength: 2, SlowestLength: 30,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("kama == nil", true, kama == nil)
		check("err", errelen, err.Error())
	})

	t.Run("efficiency ratio length < 0, error", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: -1, FastestLength: 2, SlowestLength: 30,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("kama == nil", true, kama == nil)
		check("err", errelen, err.Error())
	})

	t.Run("fastest length = 1, error", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 10, FastestLength: 1, SlowestLength: 30,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("kama == nil", true, kama == nil)
		check("err", errflen, err.Error())
	})

	t.Run("slowest length = 1, error", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 10, FastestLength: 2, SlowestLength: 1,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("kama == nil", true, kama == nil)
		check("err", errslen, err.Error())
	})

	t.Run("0 ≤ α ≤ 1, both smoothing factors", func(t *testing.T) {
		t.Parallel()

		const (
			l = 10
			f = 0.66666666
			s = 0.33333333
		)

		params := KaufmanAdaptiveMovingAverageSmoothingFactorParams{
			EfficiencyRatioLength: l, FastestSmoothingFactor: f, SlowestSmoothingFactor: s,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(kama, "kama(10, 0.6667, 0.3333, hl/2)", l, f, s)
	})

	t.Run("α < 0, fastest, error", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageSmoothingFactorParams{
			EfficiencyRatioLength: 10, FastestSmoothingFactor: -0.00000001, SlowestSmoothingFactor: 0.33333333,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageSmoothingFactor(&params)
		check("kama == nil", true, kama == nil)
		check("err", errfa, err.Error())
	})

	t.Run("α > 1, fastest, error", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageSmoothingFactorParams{
			EfficiencyRatioLength: 10, FastestSmoothingFactor: 1.00000001, SlowestSmoothingFactor: 0.33333333,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageSmoothingFactor(&params)
		check("kama == nil", true, kama == nil)
		check("err", errfa, err.Error())
	})

	t.Run("α < 0, slowest, error", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageSmoothingFactorParams{
			EfficiencyRatioLength: 10, FastestSmoothingFactor: 0.66666666, SlowestSmoothingFactor: -0.00000001,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageSmoothingFactor(&params)
		check("kama == nil", true, kama == nil)
		check("err", errsa, err.Error())
	})

	t.Run("α > 1, slowest, error", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageSmoothingFactorParams{
			EfficiencyRatioLength: 10, FastestSmoothingFactor: 0.66666666, SlowestSmoothingFactor: 1.00000001,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageSmoothingFactor(&params)
		check("kama == nil", true, kama == nil)
		check("err", errsa, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 10, FastestLength: 2, SlowestLength: 30,
			BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("kama == nil", true, kama == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 10, FastestLength: 2, SlowestLength: 30,
			BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("kama == nil", true, kama == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 10, FastestLength: 2, SlowestLength: 30,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("kama == nil", true, kama == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero, length", func(t *testing.T) {
		t.Parallel()
		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 10, FastestLength: 2, SlowestLength: 30,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "kama(10, 2, 30)", kama.LineIndicator.Mnemonic)
		check("description", "Kaufman adaptive moving average kama(10, 2, 30)", kama.LineIndicator.Description)
	})

	t.Run("all components zero, alpha", func(t *testing.T) {
		t.Parallel()
		params := KaufmanAdaptiveMovingAverageSmoothingFactorParams{
			EfficiencyRatioLength: 10, FastestSmoothingFactor: 2. / 3., SlowestSmoothingFactor: 2. / 31.,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "kama(10, 0.6667, 0.0645)", kama.LineIndicator.Mnemonic)
		check("description", "Kaufman adaptive moving average kama(10, 0.6667, 0.0645)", kama.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 10, FastestLength: 2, SlowestLength: 30,
			BarComponent: entities.BarMedianPrice,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "kama(10, 2, 30, hl/2)", kama.LineIndicator.Mnemonic)
		check("description", "Kaufman adaptive moving average kama(10, 2, 30, hl/2)", kama.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 10, FastestLength: 2, SlowestLength: 30,
			QuoteComponent: entities.QuoteBidPrice,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "kama(10, 2, 30, b)", kama.LineIndicator.Mnemonic)
		check("description", "Kaufman adaptive moving average kama(10, 2, 30, b)", kama.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := KaufmanAdaptiveMovingAverageLengthParams{
			EfficiencyRatioLength: 10, FastestLength: 2, SlowestLength: 30,
			TradeComponent: entities.TradeVolume,
		}

		kama, err := NewKaufmanAdaptiveMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "kama(10, 2, 30, v)", kama.LineIndicator.Mnemonic)
		check("description", "Kaufman adaptive moving average kama(10, 2, 30, v)", kama.LineIndicator.Description)
	})
}
