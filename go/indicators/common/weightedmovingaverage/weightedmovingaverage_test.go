//nolint:testpackage
package weightedmovingaverage

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
// Output data is taken from TA-Lib (http://ta-lib.org/) tests,
//    test_ma.c.
//
// /*******************************/
// /*   WMA TEST  - CLASSIC       */
// /*******************************/
// #ifndef TA_FUNC_NO_RANGE_CHECK
// /* No output value. */
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  0, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_BAD_PARAM, 0, 0, 0, 0 },
// #endif
// /* One value tests. */
// { 0, TA_ANY_MA_TEST, 0, 2,   2,  2, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   0,  94.52,   2, 1 },
// /* Misc tests: period 2, 30 */
// { 1, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   0,   93.71,  1,  252-1  }, /* First Value */
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   1,   94.52,  1,  252-1  },
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   2,   94.85,  1,  252-1  },
// { 0, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 250,  108.16,  1,  252-1  }, /* Last Value */
//
// { 1, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   0,  88.567,  29,  252-29 }, /* First Value */
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   1,  88.233,  29,  252-29 },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   2,  88.034,  29,  252-29 },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,  29,  87.191,  29,  252-29 },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 221, 109.3413, 29,  252-29 },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 30, TA_MAType_WMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 222, 109.3466, 29,  252-29 }, /* Last Value */

func testWeightedMovingAverageTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

//nolint:dupl
func testWeightedMovingAverageInput() []float64 {
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

func TestWeightedMovingAverageUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-2 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testWeightedMovingAverageInput()

	t.Run("length = 2", func(t *testing.T) {
		const (
			i1value   = 93.71  // Index=1 value.
			i2value   = 94.52  // Index=2 value.
			i3value   = 94.855 // Index=3 value.
			i251value = 108.16 // Index=251 (last) value.
		)

		t.Parallel()
		wma := testWeightedMovingAverageCreate(2)

		for i := 0; i < len(input); i++ {
			act := wma.Update(input[i])

			switch i {
			case 0:
				checkNaN(i, act)
			case 1:
				check(i, i1value, act)
			case 2:
				check(i, i2value, act)
			case 3:
				check(i, i3value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, wma.Update(math.NaN()))
	})

	t.Run("length = 30", func(t *testing.T) {
		const ( // The first 29 values are NaN.
			i29value  = 88.5677  // Index=29 value.
			i30value  = 88.2337  // Index=30 value.
			i31value  = 88.034   // Index=31 value.
			i58value  = 87.191   // Index=58 value.
			i250value = 109.3466 // Index=250 value.
			i251value = 109.3413 // Index=251 (last) value.
		)

		t.Parallel()
		wma := testWeightedMovingAverageCreate(30)

		for i := 0; i < 29; i++ {
			checkNaN(i, wma.Update(input[i]))
		}

		for i := 29; i < len(input); i++ {
			act := wma.Update(input[i])

			switch i {
			case 29:
				check(i, i29value, act)
			case 30:
				check(i, i30value, act)
			case 31:
				check(i, i31value, act)
			case 58:
				check(i, i58value, act)
			case 250:
				check(i, i250value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, wma.Update(math.NaN()))
	})
}

func TestWeightedMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		exp = 93.71
	)

	input := testWeightedMovingAverageInput()
	time := testWeightedMovingAverageTime()
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

		if s.Value != exp {
			t.Errorf("value is incorrect: expected %v, actual %v", exp, s.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		s := entities.Scalar{Time: time, Value: input[1]}
		wma := testWeightedMovingAverageCreate(l)
		wma.Update(input[0])
		check(wma.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: input[1]}
		wma := testWeightedMovingAverageCreate(l)
		wma.Update(input[0])
		check(wma.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: input[1], Ask: input[1]}
		wma := testWeightedMovingAverageCreate(l)
		wma.Update(input[0])
		check(wma.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: input[1]}
		wma := testWeightedMovingAverageCreate(l)
		wma.Update(input[0])
		check(wma.UpdateTrade(&r))
	})
}

func TestWeightedMovingAverageIsPrimed(t *testing.T) { //nolint:dupl
	t.Parallel()

	input := testWeightedMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		wma := testWeightedMovingAverageCreate(2)

		check(-1, false, wma.IsPrimed())

		for i := 0; i < 1; i++ {
			wma.Update(input[i])
			check(i, false, wma.IsPrimed())
		}

		for i := 1; i < len(input); i++ {
			wma.Update(input[i])
			check(i, true, wma.IsPrimed())
		}
	})

	t.Run("length = 30", func(t *testing.T) {
		t.Parallel()
		wma := testWeightedMovingAverageCreate(30)

		check(-1, false, wma.IsPrimed())

		for i := 0; i < 29; i++ {
			wma.Update(input[i])
			check(i, false, wma.IsPrimed())
		}

		for i := 29; i < len(input); i++ {
			wma.Update(input[i])
			check(i, true, wma.IsPrimed())
		}
	})
}

func TestWeightedMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	wma := testWeightedMovingAverageCreate(5)
	act := wma.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Type", core.WeightedMovingAverage, act.Type)
	check("Mnemonic", "wma(5)", act.Mnemonic)
	check("Description", "Weighted moving average wma(5)", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(WeightedMovingAverageValue), act.Outputs[0].Kind)
	check("Outputs[0].Type", outputs.ScalarType, act.Outputs[0].Type)
	check("Outputs[0].Mnemonic", "wma(5)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Weighted moving average wma(5)", act.Outputs[0].Description)
}

func TestNewWeightedMovingAverage(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc entities.BarComponent   = entities.BarMedianPrice
		qc entities.QuoteComponent = entities.QuoteMidPrice
		tc entities.TradeComponent = entities.TradePrice

		length  = 5
		divider = float64(length) * float64(length+1) / 2.
		errlen  = "invalid weighted moving average parameters: length should be greater than 1"
		errbc   = "invalid weighted moving average parameters: 9999: unknown bar component"
		errqc   = "invalid weighted moving average parameters: 9999: unknown quote component"
		errtc   = "invalid weighted moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 1", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, hl/2)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, hl/2)", wma.LineIndicator.Description)
		check("primed", false, wma.primed)
		check("lastIndex", length-1, wma.lastIndex)
		check("len(window)", length, len(wma.window))
		check("windowLength", length, wma.windowLength)
		check("divider", divider, wma.divider)
		check("windowCount", 0, wma.windowCount)
		check("windowSum", 0., wma.windowSum)
		check("windowSub", 0., wma.windowSub)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("wma == nil", true, wma == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{Length: length}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5)", wma.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{Length: length, BarComponent: entities.BarMedianPrice}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, hl/2)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, hl/2)", wma.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{Length: length, QuoteComponent: entities.QuoteBidPrice}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, b)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, b)", wma.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{Length: length, TradeComponent: entities.TradeVolume}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, v)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, v)", wma.LineIndicator.Description)
	})

	t.Run("bar and quote components set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteBidPrice,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, o, b)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, o, b)", wma.LineIndicator.Description)
	})

	t.Run("bar and trade components set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, BarComponent: entities.BarHighPrice, TradeComponent: entities.TradeVolume,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, h, v)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, h, v)", wma.LineIndicator.Description)
	})

	t.Run("quote and trade components set", func(t *testing.T) {
		t.Parallel()
		params := WeightedMovingAverageParams{
			Length: length, QuoteComponent: entities.QuoteAskPrice, TradeComponent: entities.TradeVolume,
		}

		wma, err := NewWeightedMovingAverage(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "wma(5, a, v)", wma.LineIndicator.Mnemonic)
		check("description", "Weighted moving average wma(5, a, v)", wma.LineIndicator.Description)
	})
}

func testWeightedMovingAverageCreate(length int) *WeightedMovingAverage {
	params := WeightedMovingAverageParams{
		Length: length,
	}

	wma, _ := NewWeightedMovingAverage(&params)

	return wma
}
