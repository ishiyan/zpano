//nolint:testpackage
package supersmoother

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestSuperSmootherUpdate(t *testing.T) {
	t.Parallel()

	const (
		skipRows  = 60  // Julia test skips early rows due to lead-in differences.
		tolerance = 0.5 // CSV rounded to 2dp; MBST priming differs from Julia zero-init.
	)

	input := testSuperSmootherInput()
	expected := testSuperSmootherExpected()
	ss := testSuperSmootherCreate(10)

	for i := range input {
		act := ss.Update(input[i])

		if i < 2 {
			// First 2 updates should return NaN (not primed).
			if !math.IsNaN(act) {
				t.Errorf("[%d] expected NaN, got %v", i, act)
			}

			continue
		}

		// Skip early rows where MBST and Julia priming differ.
		if i < skipRows {
			continue
		}

		if math.Abs(act-expected[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
		}
	}

	// NaN passthrough.
	if !math.IsNaN(ss.Update(math.NaN())) {
		t.Error("expected NaN passthrough")
	}
}

func TestSuperSmootherIsPrimed(t *testing.T) {
	t.Parallel()

	input := testSuperSmootherInput()
	ss := testSuperSmootherCreate(10)

	if ss.IsPrimed() {
		t.Error("should not be primed before any updates")
	}

	for i := 0; i < 2; i++ {
		ss.Update(input[i])

		if ss.IsPrimed() {
			t.Errorf("[%d] should not be primed", i)
		}
	}

	ss.Update(input[2])

	if !ss.IsPrimed() {
		t.Error("[2] should be primed")
	}
}

func TestSuperSmootherUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		inp = 100.
	)

	time := testSuperSmootherTime()
	ss := testSuperSmootherCreate(10)

	// Prime the indicator.
	ss.Update(inp)
	ss.Update(inp)
	ss.Update(inp)

	check := func(act core.Output) {
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

		if math.IsNaN(s.Value) {
			t.Error("value should not be NaN after priming")
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: time, Value: inp}
		ss2 := testSuperSmootherCreate(10)
		ss2.Update(inp)
		ss2.Update(inp)
		ss2.Update(inp)
		check(ss2.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, High: inp, Low: inp}
		ss2 := testSuperSmootherCreate(10)
		ss2.Update(inp)
		ss2.Update(inp)
		ss2.Update(inp)
		check(ss2.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		ss2 := testSuperSmootherCreate(10)
		ss2.Update(inp)
		ss2.Update(inp)
		ss2.Update(inp)
		check(ss2.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		ss2 := testSuperSmootherCreate(10)
		ss2.Update(inp)
		ss2.Update(inp)
		ss2.Update(inp)
		check(ss2.UpdateTrade(&r))
	})
}

func TestSuperSmootherMetadata(t *testing.T) {
	t.Parallel()

	ss := testSuperSmootherCreate(10)
	act := ss.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.SuperSmoother, act.Identifier)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "ss(10, hl/2)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Super Smoother ss(10, hl/2)", act.Outputs[0].Description)
}

func TestNewSuperSmoother(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarMedianPrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		period = 10
		errper = "invalid super smoother parameters: shortest cycle period should be greater than 1"
		errbc  = "invalid super smoother parameters: 9999: unknown bar component"
		errqc  = "invalid super smoother parameters: 9999: unknown quote component"
		errtc  = "invalid super smoother parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("period >= 2", func(t *testing.T) {
		t.Parallel()
		params := SuperSmootherParams{
			ShortestCyclePeriod: period, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ss, err := NewSuperSmoother(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ss(10, hl/2)", ss.LineIndicator.Mnemonic)
		check("description", "Super Smoother ss(10, hl/2)", ss.LineIndicator.Description)
		check("primed", false, ss.primed)
	})

	t.Run("period = 1", func(t *testing.T) {
		t.Parallel()
		params := SuperSmootherParams{
			ShortestCyclePeriod: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ss, err := NewSuperSmoother(&params)
		check("ss == nil", true, ss == nil)
		check("err", errper, err.Error())
	})

	t.Run("period = 0", func(t *testing.T) {
		t.Parallel()
		params := SuperSmootherParams{
			ShortestCyclePeriod: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ss, err := NewSuperSmoother(&params)
		check("ss == nil", true, ss == nil)
		check("err", errper, err.Error())
	})

	t.Run("period < 0", func(t *testing.T) {
		t.Parallel()
		params := SuperSmootherParams{
			ShortestCyclePeriod: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ss, err := NewSuperSmoother(&params)
		check("ss == nil", true, ss == nil)
		check("err", errper, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := SuperSmootherParams{
			ShortestCyclePeriod: period, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		ss, err := NewSuperSmoother(&params)
		check("ss == nil", true, ss == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := SuperSmootherParams{
			ShortestCyclePeriod: period, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		ss, err := NewSuperSmoother(&params)
		check("ss == nil", true, ss == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := SuperSmootherParams{
			ShortestCyclePeriod: period, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		ss, err := NewSuperSmoother(&params)
		check("ss == nil", true, ss == nil)
		check("err", errtc, err.Error())
	})

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := SuperSmootherParams{ShortestCyclePeriod: period}

		ss, err := NewSuperSmoother(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ss(10, hl/2)", ss.LineIndicator.Mnemonic)
		check("description", "Super Smoother ss(10, hl/2)", ss.LineIndicator.Description)
	})

	t.Run("only bar component set to open", func(t *testing.T) {
		t.Parallel()
		params := SuperSmootherParams{ShortestCyclePeriod: period, BarComponent: entities.BarOpenPrice}

		ss, err := NewSuperSmoother(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ss(10, o)", ss.LineIndicator.Mnemonic)
	})
}
