//nolint:testpackage
package chandemomentumoscillator

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

//nolint:lll
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// TA Lib uses incorrect CMO calculation which is incompatible with Chande's book.
// TA Lib smooths the CMO values in the same way as RSI does, but Chande didn't smooth the values.
// We don't use TA Lib test data here.

func testChandeMomentumOscillatorTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

//nolint:dupl
func testChandeMomentumOscillatorBookLength10Input() []float64 {
	// Chande, T.S. and Kroll, S.
	// The New Technical Trader: Boost Your Profit by Plugging Into the Latest Indicators
	// ISBN 9780471597803
	// Wiley Finance, 1994
	// Page 96, Table 5.1 Calculating 10-day CMO Using T-Bond06/93 Data
	return []float64{
		101.0313, 101.0313, 101.1250, 101.9687, 102.7813,
		103.0000, 102.9687, 103.0625, 102.9375, 102.7188,
		102.7500, 102.9063, 102.9687,
	}
}

//nolint:dupl
func testChandeMomentumOscillatorBookLength10Output() []float64 {
	return []float64{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),

		// 69.62, 71.43, 71.08,
		69.61963786608334, 71.42857142857143, 71.08377992828775,
	}
}

//nolint:dupl
func testChandeMomentumOscillatorInput() []float64 {
	return []float64{
		91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
		96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
		88.375000, 87.625000, 84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000,
		85.250000, 87.125000, 85.815000, 88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000,
		83.375000, 85.500000, 89.190000, 89.440000, 91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000,
		89.030000, 88.815000, 84.280000, 83.500000, 82.690000, 84.750000, 85.655000, 86.190000, 88.940000, 89.280000,
		88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000, 93.155000, 91.720000, 90.000000, 89.690000,
		88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000, 104.940000, 106.000000, 102.500000,
		102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000,
		112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000,
		110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
		116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000,
		124.750000, 123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000,
		132.250000, 131.000000, 132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000,
		136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000,
		125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000, 122.190000, 119.310000,
		123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000, 124.440000,
		122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
		132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000,
		130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000,
		117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
		107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
		94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000,
		95.000000, 95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000,
		105.000000, 104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000,
		113.370000, 109.000000, 109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000,
		108.000000, 108.620000, 109.750000, 109.810000, 109.000000, 108.750000, 107.870000,
	}
}

func TestChandeMomentumOscillatorUpdate(t *testing.T) { //nolint: funlen
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

	t.Run("length = 10 (book)", func(t *testing.T) {
		t.Parallel()
		input := testChandeMomentumOscillatorBookLength10Input()
		output := testChandeMomentumOscillatorBookLength10Output()
		cmo := testChandeMomentumOscillatorCreate(10)

		for i := 0; i < 10; i++ {
			checkNaN(i, cmo.Update(input[i]))
		}

		for i := 10; i < len(input); i++ {
			act := cmo.Update(input[i])
			check(i, output[i], act)
		}

		checkNaN(0, cmo.Update(math.NaN()))
	})
}

func TestChandeMomentumOscillatorUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		inp = 3.
		exp = 100.
	)

	time := testChandeMomentumOscillatorTime()
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

		if math.Abs(s.Value-exp) > 1e-13 {
			t.Errorf("value is incorrect: expected %v, actual %v", exp, s.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: time, Value: inp}
		cmo := testChandeMomentumOscillatorCreate(l)
		cmo.Update(0.)
		cmo.Update(0.)
		check(cmo.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		cmo := testChandeMomentumOscillatorCreate(l)
		cmo.Update(0.)
		cmo.Update(0.)
		check(cmo.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		cmo := testChandeMomentumOscillatorCreate(l)
		cmo.Update(0.)
		cmo.Update(0.)
		check(cmo.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		cmo := testChandeMomentumOscillatorCreate(l)
		cmo.Update(0.)
		cmo.Update(0.)
		check(cmo.UpdateTrade(&r))
	})
}

func TestChandeMomentumOscillatorIsPrimed(t *testing.T) { //nolint:funlen
	t.Parallel()

	input := testChandeMomentumOscillatorInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		cmo := testChandeMomentumOscillatorCreate(1)

		check(-1, false, cmo.IsPrimed())

		for i := 0; i < 1; i++ {
			cmo.Update(input[i])
			check(i, false, cmo.IsPrimed())
		}

		for i := 1; i < len(input); i++ {
			cmo.Update(input[i])
			check(i, true, cmo.IsPrimed())
		}
	})

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		cmo := testChandeMomentumOscillatorCreate(2)

		check(-1, false, cmo.IsPrimed())

		for i := 0; i < 2; i++ {
			cmo.Update(input[i])
			check(i, false, cmo.IsPrimed())
		}

		for i := 2; i < len(input); i++ {
			cmo.Update(input[i])
			check(i, true, cmo.IsPrimed())
		}
	})

	t.Run("length = 3", func(t *testing.T) {
		t.Parallel()
		cmo := testChandeMomentumOscillatorCreate(3)

		check(-1, false, cmo.IsPrimed())

		for i := 0; i < 3; i++ {
			cmo.Update(input[i])
			check(i, false, cmo.IsPrimed())
		}

		for i := 3; i < len(input); i++ {
			cmo.Update(input[i])
			check(i, true, cmo.IsPrimed())
		}
	})

	t.Run("length = 5", func(t *testing.T) {
		t.Parallel()
		cmo := testChandeMomentumOscillatorCreate(5)

		check(-1, false, cmo.IsPrimed())

		for i := 0; i < 5; i++ {
			cmo.Update(input[i])
			check(i, false, cmo.IsPrimed())
		}

		for i := 5; i < len(input); i++ {
			cmo.Update(input[i])
			check(i, true, cmo.IsPrimed())
		}
	})

	t.Run("length = 10", func(t *testing.T) {
		t.Parallel()
		cmo := testChandeMomentumOscillatorCreate(10)

		check(-1, false, cmo.IsPrimed())

		for i := 0; i < 10; i++ {
			cmo.Update(input[i])
			check(i, false, cmo.IsPrimed())
		}

		for i := 10; i < len(input); i++ {
			cmo.Update(input[i])
			check(i, true, cmo.IsPrimed())
		}
	})
}

func TestChandeMomentumOscillatorMetadata(t *testing.T) {
	t.Parallel()

	cmo := testChandeMomentumOscillatorCreate(5)
	act := cmo.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Type", core.ChandeMomentumOscillator, act.Type)
	check("Mnemonic", "cmo(5)", act.Mnemonic)
	check("Description", "Chande Momentum Oscillator cmo(5)", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Type", outputs.ScalarType, act.Outputs[0].Type)
	check("Outputs[0].Mnemonic", "cmo(5)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Chande Momentum Oscillator cmo(5)", act.Outputs[0].Description)
}

func TestNewChandeMomentumOscillator(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 5
		errlen                         = "invalid Chande momentum oscillator parameters: length should be positive"
		errbc                          = "invalid Chande momentum oscillator parameters: 9999: unknown bar component"
		errqc                          = "invalid Chande momentum oscillator parameters: 9999: unknown quote component"
		errtc                          = "invalid Chande momentum oscillator parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(5, hl/2)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(5, hl/2)", cmo.LineIndicator.Description)
		check("primed", false, cmo.primed)
		check("length", length, cmo.length)
		check("len(ringBuffer)", length, len(cmo.ringBuffer))
		check("ringHead", 0, cmo.ringHead)
		check("count", 0, cmo.count)
		check("prevSample", 0., cmo.prevSample)
		check("gainSum", 0., cmo.gainSum)
		check("lossSum", 0., cmo.lossSum)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(1, hl/2)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(1, hl/2)", cmo.LineIndicator.Description)
		check("primed", false, cmo.primed)
		check("length", 1, cmo.length)
		check("len(ringBuffer)", 1, len(cmo.ringBuffer))
		check("ringHead", 0, cmo.ringHead)
		check("count", 0, cmo.count)
		check("prevSample", 0., cmo.prevSample)
		check("gainSum", 0., cmo.gainSum)
		check("lossSum", 0., cmo.lossSum)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		cmo, err := New(&params)
		check("cmo == nil", true, cmo == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		cmo, err := New(&params)
		check("cmo == nil", true, cmo == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		cmo, err := New(&params)
		check("cmo == nil", true, cmo == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		cmo, err := New(&params)
		check("cmo == nil", true, cmo == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		cmo, err := New(&params)
		check("cmo == nil", true, cmo == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length}

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(5)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(5)", cmo.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, BarComponent: entities.BarMedianPrice}

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(5, hl/2)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(5, hl/2)", cmo.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, QuoteComponent: entities.QuoteBidPrice}

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(5, b)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(5, b)", cmo.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, TradeComponent: entities.TradeVolume}

		cmo, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "cmo(5, v)", cmo.LineIndicator.Mnemonic)
		check("description", "Chande Momentum Oscillator cmo(5, v)", cmo.LineIndicator.Description)
	})
}

func testChandeMomentumOscillatorCreate(length int) *ChandeMomentumOscillator {
	params := Params{Length: length}

	cmo, _ := New(&params)

	return cmo
}
