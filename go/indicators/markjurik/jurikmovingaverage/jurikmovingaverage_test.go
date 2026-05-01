//nolint:testpackage
package jurikmovingaverage

import (
	"math"
	"testing"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

//nolint:funlen
func TestJurikMovingAverageUpdate(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-13 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	run := func(jma *JurikMovingAverage, input []float64, output []float64) {
		t.Helper()

		const lenPrimed = 30

		for i := 0; i < len(input); i++ {
			act := jma.Update(input[i])

			if i < lenPrimed {
				checkNaN(i, act)
			} else {
				check(i, output[i], act)
			}
		}

		checkNaN(0, jma.Update(math.NaN()))
	}

	input := testJurikMovingAverageInput()

	t.Run("length = 20, phase = -100", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(20, -100)
		run(jma, input, testJurikMovingAverageLength20PhaseMin100Output())
	})

	t.Run("length = 20, phase = -30", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(20, -30)
		run(jma, input, testJurikMovingAverageLength20PhaseMin30Output())
	})

	t.Run("length = 20, phase = 0", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(20, 0)
		run(jma, input, testJurikMovingAverageLength20Phase0Output())
	})

	t.Run("length = 20, phase = 30", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(20, 30)
		run(jma, input, testJurikMovingAverageLength20Phase30Output())
	})

	t.Run("length = 20, phase = 100", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(20, 100)
		run(jma, input, testJurikMovingAverageLength20Phase100Output())
	})

	t.Run("length = 2, phase = 1", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(2, 1)
		run(jma, input, testJurikMovingAverageLength2Phase1Output())
	})

	t.Run("length = 5, phase = 1", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(5, 1)
		run(jma, input, testJurikMovingAverageLength5Phase1Output())
	})

	t.Run("length = 10, phase = 1", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(10, 1)
		run(jma, input, testJurikMovingAverageLength10Phase1Output())
	})

	t.Run("length = 25, phase = 1", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(25, 1)
		run(jma, input, testJurikMovingAverageLength25Phase1Output())
	})

	t.Run("length = 50, phase = 1", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(50, 1)
		run(jma, input, testJurikMovingAverageLength50Phase1Output())
	})

	t.Run("length = 75, phase = 1", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(75, 1)
		run(jma, input, testJurikMovingAverageLength75Phase1Output())
	})

	t.Run("length = 100, phase = 1", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(100, 1)
		run(jma, input, testJurikMovingAverageLength100Phase1Output())
	})
}

//nolint:funlen
func TestJurikMovingAverageUpdateEntity(t *testing.T) {
	t.Parallel()

	const (
		l         = 10
		phase     = 11
		inp       = 3
		lenPrimed = 30
	)

	time := testJurikMovingAverageTime()
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
	}

	update := func(jma *JurikMovingAverage) {
		t.Helper()

		for range lenPrimed {
			jma.Update(inp)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(l, phase)
		update(jma)

		s := entities.Scalar{Time: time, Value: inp}
		check(jma.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(l, phase)
		update(jma)

		b := entities.Bar{Time: time, Close: inp}
		check(jma.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(l, phase)
		update(jma)

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		check(jma.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(l, phase)
		update(jma)

		r := entities.Trade{Time: time, Price: inp}
		check(jma.UpdateTrade(&r))
	})
}

func TestJurikMovingAverageIsPrimed(t *testing.T) {
	t.Parallel()

	input := testJurikMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 10, phase = 30", func(t *testing.T) {
		t.Parallel()

		const lenPrimed = 30

		jma := testJurikMovingAverageCreate(10, 30)
		check(0, false, jma.IsPrimed())

		for i := 0; i < lenPrimed; i++ {
			jma.Update(input[i])
			check(i+1, false, jma.IsPrimed())
		}

		for i := lenPrimed; i < len(input); i++ {
			jma.Update(input[i])
			check(i+1, true, jma.IsPrimed())
		}
	})
}

func TestJurikMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("length = 10, fphase = 30", func(t *testing.T) {
		t.Parallel()

		jma := testJurikMovingAverageCreate(10, 30)
		act := jma.Metadata()

		check("Identifier", core.JurikMovingAverage, act.Identifier)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(MovingAverage), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "jma(10, 30)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Jurik moving average jma(10, 30)", act.Outputs[0].Description)
	})
}

func TestNewMovingAverage(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc          entities.BarComponent   = entities.BarMedianPrice
		qc          entities.QuoteComponent = entities.QuoteMidPrice
		tc          entities.TradeComponent = entities.TradePrice
		lengthMin1                          = -1
		length0                             = 0
		length1                             = 1
		length10                            = 10
		phaseMin101                         = -101
		phaseMin100                         = -100
		phaseMin30                          = -30
		phase0                              = 0
		phase30                             = 30
		phase100                            = 100
		phase101                            = 101

		errlen = "invalid jurik moving average parameters: length should be positive"
		errpha = "invalid jurik moving average parameters: phase should be in range [-100, 100]"
		errbc  = "invalid jurik moving average parameters: 9999: unknown bar component"
		errqc  = "invalid jurik moving average parameters: 9999: unknown quote component"
		errtc  = "invalid jurik moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(params *JurikMovingAverageParams, mnemonic string) {
		t.Helper()

		jma, err := NewJurikMovingAverage(params)
		check("err == nil", true, err == nil)
		check("mnemonic", mnemonic, jma.LineIndicator.Mnemonic)
		check("description", "Jurik moving average "+mnemonic, jma.LineIndicator.Description)
		check("primed", false, jma.primed)
	}

	checkError := func(params *JurikMovingAverageParams, e string) {
		t.Helper()

		jma, err := NewJurikMovingAverage(params)
		check("jma == nil", true, jma == nil)
		check("err", e, err.Error())
	}

	t.Run("length > 1, phase = 30", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phase30, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}, "jma(10, 30, hl/2)")
	})

	t.Run("length = 1, phase = 30", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length1, Phase: phase30, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}, "jma(1, 30, hl/2)")
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()

		checkError(&JurikMovingAverageParams{
			Length: length0, Phase: phase30, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}, errlen)
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()

		checkError(&JurikMovingAverageParams{
			Length: lengthMin1, Phase: phase30, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}, errlen)
	})

	t.Run("length = 10, phase < -100", func(t *testing.T) {
		t.Parallel()

		checkError(&JurikMovingAverageParams{
			Length: length10, Phase: phaseMin101, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}, errpha)
	})

	t.Run("length = 10, phase > 100", func(t *testing.T) {
		t.Parallel()

		checkError(&JurikMovingAverageParams{
			Length: length10, Phase: phase101, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}, errpha)
	})

	t.Run("length = 10, phase = -100", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phaseMin100, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}, "jma(10, -100, hl/2)")
	})

	t.Run("length = 10, phase = -30", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phaseMin30, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}, "jma(10, -30, hl/2)")
	})

	t.Run("length = 10, phase = 0", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phase0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}, "jma(10, 0, hl/2)")
	})

	t.Run("length = 10, phase = 100", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phase100, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}, "jma(10, 100, hl/2)")
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		checkError(&JurikMovingAverageParams{
			Length: length10, Phase: phase30,
			BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}, errbc)
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		checkError(&JurikMovingAverageParams{
			Length: length10, Phase: phase30,
			BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}, errqc)
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		checkError(&JurikMovingAverageParams{
			Length: length10, Phase: phase30,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}, errtc)
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phase30,
		}, "jma(10, 30)")
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phase30, BarComponent: entities.BarMedianPrice,
		}, "jma(10, 30, hl/2)")
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phase30, QuoteComponent: entities.QuoteBidPrice,
		}, "jma(10, 30, b)")
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phase30, TradeComponent: entities.TradeVolume,
		}, "jma(10, 30, v)")
	})

	t.Run("bar and quote components set", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phase30,
			BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteBidPrice,
		}, "jma(10, 30, o, b)")
	})

	t.Run("bar and trade components set", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phase30,
			BarComponent: entities.BarHighPrice, TradeComponent: entities.TradeVolume,
		}, "jma(10, 30, h, v)")
	})

	t.Run("quote and trade components set", func(t *testing.T) {
		t.Parallel()

		checkInstance(&JurikMovingAverageParams{
			Length: length10, Phase: phase30,
			QuoteComponent: entities.QuoteAskPrice, TradeComponent: entities.TradeVolume,
		}, "jma(10, 30, a, v)")
	})
}

func testJurikMovingAverageCreate(length, phase int) *JurikMovingAverage {
	params := JurikMovingAverageParams{
		Length: length, Phase: phase,
	}

	jma, _ := NewJurikMovingAverage(&params)

	return jma
}
