//nolint:testpackage
package exponentialmovingaverage

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs/shape"
)

//nolint:lll
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// Output data is taken from TA-Lib (http://ta-lib.org/) tests,
//    test_ma.c.
//
//   /*******************************/
//   /*   EMA TEST - Classic        */
//   /*******************************/
//   /* No output value. */
//   { 0, TA_ANY_MA_TEST, 0, 1, 1,  14, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 0, 0, 0, 0},
//   #ifndef TA_FUNC_NO_RANGE_CHECK
//   { 0, TA_ANY_MA_TEST, 0, 0, 251,  0, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_BAD_PARAM, 0, 0, 0, 0 },
//   #endif
//   /* Misc tests: period 2, 10 */
//   { 1, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   0,  93.15, 1, 251 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   1,  93.96, 1, 251 },
//   { 0, TA_ANY_MA_TEST, 0, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 250, 108.21, 1, 251 }, /* Last Value */
//
//   { 1, TA_ANY_MA_TEST, 0, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,    0,  93.22,  9, 243 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 0, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,    1,  93.75,  9, 243 },
//   { 0, TA_ANY_MA_TEST, 0, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,   20,  86.46,  9, 243 },
//   { 0, TA_ANY_MA_TEST, 0, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,  242, 108.97,  9, 243 }, /* Last Value */
//   /*******************************/
//   /*   EMA TEST - Metastock      */
//   /*******************************/
//   /* No output value. */
//   { 0, TA_ANY_MA_TEST, 0, 1, 1,  14, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 0, 0, 0, 0},
//   #ifndef TA_FUNC_NO_RANGE_CHECK
//   { 0, TA_ANY_MA_TEST, 0, 0, 251,  0, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_BAD_PARAM, 0, 0, 0, 0 },
//   #endif
//   /* Test with 1 unstable price bar. Test for period 2, 10 */
//   { 1, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   0,  94.15, 1+1, 251-1 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1,  94.78, 1+1, 251-1 },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 250-1, 108.21, 1+1, 251-1 }, /* Last Value */
//
//   { 1, TA_ANY_MA_TEST, 1, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    0,  93.24,  9+1, 243-1 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    1,  93.97,  9+1, 243-1 },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   20,  86.23,  9+1, 243-1 },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 242-1, 108.97,  9+1, 243-1 }, /* Last Value */
//
//   /* Test with 2 unstable price bar. Test for period 2, 10 */
//   { 0, TA_ANY_MA_TEST, 2, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   0,  94.78, 1+2, 251-2 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 2, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1,  94.11, 1+2, 251-2 },
//   { 0, TA_ANY_MA_TEST, 2, 0, 251,  2, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 250-2, 108.21, 1+2, 251-2 }, /* Last Value */
//
//   { 0, TA_ANY_MA_TEST, 2, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    0,  93.97,  9+2, 243-2 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 2, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    1,  94.79,  9+2, 243-2 },
//   { 0, TA_ANY_MA_TEST, 2, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   20,  86.39,  9+2, 243-2 },
//   { 0, TA_ANY_MA_TEST, 2, 0, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,  242-2, 108.97,  9+2, 243-2 }, /* Last Value */
//
//   /* Last 3 value with 1 unstable, period 10 */
//   { 0, TA_ANY_MA_TEST, 1, 249, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1, 109.22, 249, 3 },
//   { 0, TA_ANY_MA_TEST, 1, 249, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   2, 108.97, 249, 3 },
//
//   /* Last 3 value with 2 unstable, period 10 */
//   { 0, TA_ANY_MA_TEST, 2, 249, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   2, 108.97, 249, 3 },
//
//   /* Last 3 value with 3 unstable, period 10 */
//   { 0, TA_ANY_MA_TEST, 3, 249, 251,  10, TA_MAType_EMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   2, 108.97, 249, 3 }

func testExponentialMovingAverageTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func testExponentialMovingAverageInput() []float64 { //nolint:dupl
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

func TestExponentialMovingAverageUpdate(t *testing.T) { //nolint: funlen
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

	input := testExponentialMovingAverageInput()

	t.Run("length = 2, firstIsAverage = true", func(t *testing.T) {
		const (
			i1value   = 93.15  // Index=1 value.
			i2value   = 93.96  // Index=2 value.
			i3value   = 94.71  // Index=3 value.
			i251value = 108.21 // Index=251 (last) value.
		)

		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(2, true)

		for i := 0; i < len(input); i++ {
			act := ema.Update(input[i])

			switch i {
			case 0:
				checkNaN(i, act)
				check(i, input[1], act)
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

		checkNaN(0, ema.Update(math.NaN()))
	})

	t.Run("length = 10, firstIsAverage = true", func(t *testing.T) {
		const (
			i9value   = 93.22  // Index=9 value.
			i10value  = 93.75  // Index=10 value.
			i29value  = 86.46  // Index=29 value.
			i251value = 108.97 // Index=251 (last) value.
		)

		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(10, true)

		for i := 0; i < 9; i++ {
			checkNaN(i, ema.Update(input[i]))
		}

		for i := 9; i < len(input); i++ {
			act := ema.Update(input[i])

			switch i {
			case 9:
				check(i, i9value, act)
			case 10:
				check(i, i10value, act)
			case 29:
				check(i, i29value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, ema.Update(math.NaN()))
	})

	t.Run("length = 2, firstIsAverage = false (Metastock)", func(t *testing.T) {
		const (
			// The very first value is the input value.
			i1value   = 93.71  // Index=1 value.
			i2value   = 94.15  // Index=2 value.
			i3value   = 94.78  // Index=3 value.
			i251value = 108.21 // Index=251 (last) value.
		)

		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(2, false)

		for i := 0; i < len(input); i++ {
			act := ema.Update(input[i])

			switch i {
			case 0:
				checkNaN(i, act)
				check(i, input[1], act)
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

		checkNaN(0, ema.Update(math.NaN()))
	})

	t.Run("length = 10, firstIsAverage = false (Metastock)", func(t *testing.T) {
		const (
			// The very first value is the input value.
			i9value   = 92.60  // Index=9 value.
			i10value  = 93.24  // Index=10 value.
			i11value  = 93.97  // Index=11 value.
			i30value  = 86.23  // Index=30 value.
			i251value = 108.97 // Index=251 (last) value.
		)

		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(10, false)

		for i := 0; i < 9; i++ {
			checkNaN(i, ema.Update(input[i]))
		}

		for i := 9; i < len(input); i++ {
			act := ema.Update(input[i])

			switch i {
			case 9:
				check(i, i9value, act)
			case 10:
				check(i, i10value, act)
			case 11:
				check(i, i11value, act)
			case 30:
				check(i, i30value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, ema.Update(math.NaN()))
	})
}

func TestExponentialMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l     = 2
		alpha = 2. / float64(l+1)
		inp   = 3.
		exp   = alpha * inp
	)

	time := testExponentialMovingAverageTime()
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

		s := entities.Scalar{Time: time, Value: inp}
		ema := testExponentialMovingAverageCreateLength(l, false)
		ema.Update(0.)
		ema.Update(0.)
		check(ema.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		ema := testExponentialMovingAverageCreateLength(l, false)
		ema.Update(0.)
		ema.Update(0.)
		check(ema.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		ema := testExponentialMovingAverageCreateLength(l, false)
		ema.Update(0.)
		ema.Update(0.)
		check(ema.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		ema := testExponentialMovingAverageCreateLength(l, false)
		ema.Update(0.)
		ema.Update(0.)
		check(ema.UpdateTrade(&r))
	})
}

func TestExponentialMovingAverageIsPrimed(t *testing.T) {
	t.Parallel()

	input := testExponentialMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 10, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(10, true)

		check(0, false, ema.IsPrimed())

		for i := 0; i < 9; i++ {
			ema.Update(input[i])
			check(i+1, false, ema.IsPrimed())
		}

		for i := 9; i < len(input); i++ {
			ema.Update(input[i])
			check(i+1, true, ema.IsPrimed())
		}
	})

	t.Run("length = 10, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(10, false)

		check(0, false, ema.IsPrimed())

		for i := 0; i < 9; i++ {
			ema.Update(input[i])
			check(i+1, false, ema.IsPrimed())
		}

		for i := 9; i < len(input); i++ {
			ema.Update(input[i])
			check(i+1, true, ema.IsPrimed())
		}
	})
}

func TestExponentialMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("length = 10, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()
		ema := testExponentialMovingAverageCreateLength(10, true)
		act := ema.Metadata()

		check("Identifier", core.ExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "ema(10)", act.Mnemonic)
		check("Description", "Exponential moving average ema(10)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "ema(10)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Exponential moving average ema(10)", act.Outputs[0].Description)
	})

	t.Run("alpha = 2/11 = 0.18181818..., firstIsAverage = false", func(t *testing.T) {
		t.Parallel()

		// α = 2 / (ℓ + 1) = 2/11 = 0.18181818...
		const alpha = 2. / 11.

		ema := testExponentialMovingAverageCreateAlpha(alpha, false)
		act := ema.Metadata()

		check("Identifier", core.ExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "ema(10, 0.18181818)", act.Mnemonic)
		check("Description", "Exponential moving average ema(10, 0.18181818)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "ema(10, 0.18181818)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Exponential moving average ema(10, 0.18181818)", act.Outputs[0].Description)
	})

	t.Run("length with non-default bar component", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{
			Length: 10, FirstIsAverage: true, BarComponent: entities.BarMedianPrice,
		}

		ema, _ := NewExponentialMovingAverageLength(&params)
		act := ema.Metadata()

		check("Mnemonic", "ema(10, hl/2)", act.Mnemonic)
		check("Description", "Exponential moving average ema(10, hl/2)", act.Description)
	})

	t.Run("alpha with non-default quote component", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2. / 11., FirstIsAverage: false, QuoteComponent: entities.QuoteBidPrice,
		}

		ema, _ := NewExponentialMovingAverageSmoothingFactor(&params)
		act := ema.Metadata()

		check("Mnemonic", "ema(10, 0.18181818, b)", act.Mnemonic)
		check("Description", "Exponential moving average ema(10, 0.18181818, b)", act.Description)
	})
}

func TestNewExponentialMovingAverage(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 10
		alpha                          = 2. / 11.

		errlen   = "invalid exponential moving average parameters: length should be positive"
		erralpha = "invalid exponential moving average parameters: smoothing factor should be in range [0, 1]"
		errbc    = "invalid exponential moving average parameters: 9999: unknown bar component"
		errqc    = "invalid exponential moving average parameters: 9999: unknown quote component"
		errtc    = "invalid exponential moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 1, firstIsAverage = false", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{
			Length: length, FirstIsAverage: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", false, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", length, ema.length)
		check("smoothingFactor", alpha, ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("length = 1, firstIsAverage = true", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{
			Length: 1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(1, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(1, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", true, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", 1, ema.length)
		check("smoothingFactor", 1., ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{
			Length: 0, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("ema == nil", true, ema == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{
			Length: -1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("ema == nil", true, ema == nil)
		check("err", errlen, err.Error())
	})

	t.Run("epsilon < α ≤ 1", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, 0.18181818, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, 0.18181818, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", true, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", length, ema.length)
		check("smoothingFactor", alpha, ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("0 < α < epsilon", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 0.000000001, FirstIsAverage: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(199999999, 0.00000001, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(199999999, 0.00000001, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", false, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", 199999999, ema.length) // 2./0.00000001 - 1.
		check("smoothingFactor", 0.00000001, ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("α = 0", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 0, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(199999999, 0.00000001, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(199999999, 0.00000001, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", true, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", 199999999, ema.length) // 2./0.00000001 - 1.
		check("smoothingFactor", 0.00000001, ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("α = 1", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(1, 1.00000000, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(1, 1.00000000, hl/2)", ema.LineIndicator.Description)
		check("firstIsAverage", true, ema.firstIsAverage)
		check("primed", false, ema.primed)
		check("length", 1, ema.length) // 2./1 - 1.
		check("smoothingFactor", 1., ema.smoothingFactor)
		check("count", 0, ema.count)
		check("sum", 0., ema.sum)
		check("value", 0., ema.value)
	})

	t.Run("α < 0", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: -1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("ema == nil", true, ema == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("α > 1", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("ema == nil", true, ema == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("ema == nil", true, ema == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("ema == nil", true, ema == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("ema == nil", true, ema == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero, length", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{Length: length}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10)", ema.LineIndicator.Description)
	})

	t.Run("all components zero, alpha", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageSmoothingFactorParams{SmoothingFactor: alpha}

		ema, err := NewExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, 0.18181818)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, 0.18181818)", ema.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{Length: length, BarComponent: entities.BarMedianPrice}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, hl/2)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, hl/2)", ema.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{Length: length, QuoteComponent: entities.QuoteBidPrice}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, b)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, b)", ema.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := ExponentialMovingAverageLengthParams{Length: length, TradeComponent: entities.TradeVolume}

		ema, err := NewExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "ema(10, v)", ema.LineIndicator.Mnemonic)
		check("description", "Exponential moving average ema(10, v)", ema.LineIndicator.Description)
	})
}

func testExponentialMovingAverageCreateLength(length int, firstIsAverage bool) *ExponentialMovingAverage {
	params := ExponentialMovingAverageLengthParams{
		Length: length, FirstIsAverage: firstIsAverage,
	}

	ema, _ := NewExponentialMovingAverageLength(&params)

	return ema
}

func testExponentialMovingAverageCreateAlpha(alpha float64, firstIsAverage bool) *ExponentialMovingAverage {
	params := ExponentialMovingAverageSmoothingFactorParams{
		SmoothingFactor: alpha, FirstIsAverage: firstIsAverage,
	}

	ema, _ := NewExponentialMovingAverageSmoothingFactor(&params)

	return ema
}
