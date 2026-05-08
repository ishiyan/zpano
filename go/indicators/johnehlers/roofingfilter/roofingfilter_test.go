//nolint:testpackage
package roofingfilter

//nolint: gofumpt
import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

func TestRoofingFilterUpdate1Pole(t *testing.T) {
	t.Parallel()

	const (
		skipRows  = 30 // Julia/C# priming differs; skip early rows.
		tolerance = 0.5
	)

	input := testInput()
	expected := testExpected71()
	rf := testCreate1Pole()

	for i := range input {
		act := rf.Update(input[i])

		if i < 3 {
			if !math.IsNaN(act) {
				t.Errorf("[%d] expected NaN, got %v", i, act)
			}

			continue
		}

		if i < skipRows {
			continue
		}

		if math.Abs(act-expected[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
		}
	}

	if !math.IsNaN(rf.Update(math.NaN())) {
		t.Error("expected NaN passthrough")
	}
}

func TestRoofingFilterUpdate1PoleZeroMean(t *testing.T) {
	t.Parallel()

	const (
		skipRows  = 30
		tolerance = 0.5
	)

	input := testInput()
	expected := testExpected72()
	rf := testCreate1PoleZeroMean()

	for i := range input {
		act := rf.Update(input[i])

		if i < 4 {
			if !math.IsNaN(act) {
				t.Errorf("[%d] expected NaN, got %v", i, act)
			}

			continue
		}

		if i < skipRows {
			continue
		}

		if math.Abs(act-expected[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
		}
	}
}

func TestRoofingFilterUpdate2Pole(t *testing.T) {
	t.Parallel()

	const (
		skipRows  = 30
		tolerance = 0.5
	)

	input := testInput()
	expected := testExpected73()
	rf := testCreate2Pole()

	for i := range input {
		act := rf.Update(input[i])

		if i < 4 {
			if !math.IsNaN(act) {
				t.Errorf("[%d] expected NaN, got %v", i, act)
			}

			continue
		}

		if i < skipRows {
			continue
		}

		if math.Abs(act-expected[i]) > tolerance {
			t.Errorf("[%d] expected %v, got %v", i, expected[i], act)
		}
	}
}

func TestRoofingFilterIsPrimed1Pole(t *testing.T) {
	t.Parallel()

	input := testInput()
	rf := testCreate1Pole()

	if rf.IsPrimed() {
		t.Error("should not be primed before any updates")
	}

	for i := 0; i < 3; i++ {
		rf.Update(input[i])

		if rf.IsPrimed() {
			t.Errorf("[%d] should not be primed", i)
		}
	}

	rf.Update(input[3])

	if !rf.IsPrimed() {
		t.Error("[3] should be primed")
	}
}

func TestRoofingFilterIsPrimed1PoleZeroMean(t *testing.T) {
	t.Parallel()

	input := testInput()
	rf := testCreate1PoleZeroMean()

	for i := 0; i < 4; i++ {
		rf.Update(input[i])

		if rf.IsPrimed() {
			t.Errorf("[%d] should not be primed", i)
		}
	}

	rf.Update(input[4])

	if !rf.IsPrimed() {
		t.Error("[4] should be primed")
	}
}

func TestRoofingFilterIsPrimed2Pole(t *testing.T) {
	t.Parallel()

	input := testInput()
	rf := testCreate2Pole()

	for i := 0; i < 4; i++ {
		rf.Update(input[i])

		if rf.IsPrimed() {
			t.Errorf("[%d] should not be primed", i)
		}
	}

	rf.Update(input[4])

	if !rf.IsPrimed() {
		t.Error("[4] should be primed")
	}
}

func TestRoofingFilterUpdateEntity(t *testing.T) {
	t.Parallel()

	const inp = 100.

	time := testTime()
	rf := testCreate1Pole()

	// Prime the indicator.
	for i := 0; i < 4; i++ {
		rf.Update(inp)
	}

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
		rf2 := testCreate1Pole()
		for i := 0; i < 4; i++ {
			rf2.Update(inp)
		}
		check(rf2.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, High: inp, Low: inp}
		rf2 := testCreate1Pole()
		for i := 0; i < 4; i++ {
			rf2.Update(inp)
		}
		check(rf2.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		rf2 := testCreate1Pole()
		for i := 0; i < 4; i++ {
			rf2.Update(inp)
		}
		check(rf2.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		rf2 := testCreate1Pole()
		for i := 0; i < 4; i++ {
			rf2.Update(inp)
		}
		check(rf2.UpdateTrade(&r))
	})
}

func TestRoofingFilterMetadata(t *testing.T) {
	t.Parallel()

	rf := testCreate1Pole()
	act := rf.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.RoofingFilter, act.Identifier)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "roof1hp(10, 48, hl/2)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Roofing Filter roof1hp(10, 48, hl/2)", act.Outputs[0].Description)
}

func TestNewRoofingFilter(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarMedianPrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		shortest = 10
		longest  = 48
		errshort = "invalid roofing filter parameters: shortest cycle period should be greater than 1"
		errlong  = "invalid roofing filter parameters: longest cycle period should be greater than shortest"
		errbc    = "invalid roofing filter parameters: 9999: unknown bar component"
		errqc    = "invalid roofing filter parameters: 9999: unknown quote component"
		errtc    = "invalid roofing filter parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("valid params", func(t *testing.T) {
		t.Parallel()
		params := RoofingFilterParams{
			ShortestCyclePeriod: shortest, LongestCyclePeriod: longest,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rf, err := NewRoofingFilter(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "roof1hp(10, 48, hl/2)", rf.LineIndicator.Mnemonic)
		check("description", "Roofing Filter roof1hp(10, 48, hl/2)", rf.LineIndicator.Description)
		check("primed", false, rf.primed)
	})

	t.Run("2-pole mnemonic", func(t *testing.T) {
		t.Parallel()
		params := RoofingFilterParams{
			ShortestCyclePeriod: shortest, LongestCyclePeriod: longest,
			HasTwoPoleHighpassFilter: true,
		}

		rf, err := NewRoofingFilter(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "roof2hp(10, 48, hl/2)", rf.LineIndicator.Mnemonic)
	})

	t.Run("zero-mean mnemonic", func(t *testing.T) {
		t.Parallel()
		params := RoofingFilterParams{
			ShortestCyclePeriod: shortest, LongestCyclePeriod: longest,
			HasZeroMean: true,
		}

		rf, err := NewRoofingFilter(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "roof1hpzm(10, 48, hl/2)", rf.LineIndicator.Mnemonic)
	})

	t.Run("shortest < 2", func(t *testing.T) {
		t.Parallel()
		params := RoofingFilterParams{
			ShortestCyclePeriod: 1, LongestCyclePeriod: longest,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rf, err := NewRoofingFilter(&params)
		check("rf == nil", true, rf == nil)
		check("err", errshort, err.Error())
	})

	t.Run("longest <= shortest", func(t *testing.T) {
		t.Parallel()
		params := RoofingFilterParams{
			ShortestCyclePeriod: shortest, LongestCyclePeriod: shortest,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rf, err := NewRoofingFilter(&params)
		check("rf == nil", true, rf == nil)
		check("err", errlong, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := RoofingFilterParams{
			ShortestCyclePeriod: shortest, LongestCyclePeriod: longest,
			BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		rf, err := NewRoofingFilter(&params)
		check("rf == nil", true, rf == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := RoofingFilterParams{
			ShortestCyclePeriod: shortest, LongestCyclePeriod: longest,
			BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		rf, err := NewRoofingFilter(&params)
		check("rf == nil", true, rf == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := RoofingFilterParams{
			ShortestCyclePeriod: shortest, LongestCyclePeriod: longest,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		rf, err := NewRoofingFilter(&params)
		check("rf == nil", true, rf == nil)
		check("err", errtc, err.Error())
	})

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := RoofingFilterParams{ShortestCyclePeriod: shortest, LongestCyclePeriod: longest}

		rf, err := NewRoofingFilter(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "roof1hp(10, 48, hl/2)", rf.LineIndicator.Mnemonic)
	})

	t.Run("bar component set to open", func(t *testing.T) {
		t.Parallel()
		params := RoofingFilterParams{
			ShortestCyclePeriod: shortest, LongestCyclePeriod: longest,
			BarComponent: entities.BarOpenPrice,
		}

		rf, err := NewRoofingFilter(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "roof1hp(10, 48, o)", rf.LineIndicator.Mnemonic)
	})
}
