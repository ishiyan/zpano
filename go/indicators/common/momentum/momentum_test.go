//nolint:testpackage
package momentum

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
// Output data, length=14.
// Taken from TA-Lib (http://ta-lib.org/) tests, test_mom.c.
//
// static TA_Test tableTest[] =
// {
//    /**********************/
//    /*      MOM TEST      */
//    /**********************/
//
// #ifndef TA_FUNC_NO_RANGE_CHECK
//    /* Test out of range. */
//    { 0, TA_MOM_TEST, -1, 3, 14, TA_OUT_OF_RANGE_START_INDEX, 0, 0, 0, 0},
//    { 0, TA_MOM_TEST,  3, -1, 14, TA_OUT_OF_RANGE_END_INDEX,   0, 0, 0, 0},
//    { 0, TA_MOM_TEST,  4, 3, 14, TA_OUT_OF_RANGE_END_INDEX,   0, 0, 0, 0},
// #endif
//    { 1, TA_MOM_TEST, 0, 251, 14, TA_SUCCESS,      0, -0.50,  14,  252-14 }, /* First Value */
//    { 0, TA_MOM_TEST, 0, 251, 14, TA_SUCCESS,      1, -2.00,  14,  252-14 },
//    { 0, TA_MOM_TEST, 0, 251, 14, TA_SUCCESS,      2, -5.22,  14,  252-14 },
//    { 0, TA_MOM_TEST, 0, 251, 14, TA_SUCCESS, 252-15, -1.13,  14,  252-14 },  /* Last Value */
//    /* No output value. */
//    { 0, TA_MOM_TEST, 1, 1,  14, TA_SUCCESS, 0, 0, 0, 0},
//    /* One value tests. */
//    { 0, TA_MOM_TEST, 14,  14, 14, TA_SUCCESS, 0, -0.50,     14, 1},
//    /* Index too low test. */
//    { 0, TA_MOM_TEST, 0,  15, 14, TA_SUCCESS, 0, -0.50,     14, 2},
//    { 0, TA_MOM_TEST, 1,  15, 14, TA_SUCCESS, 0, -0.50,     14, 2},
//    { 0, TA_MOM_TEST, 2,  16, 14, TA_SUCCESS, 0, -0.50,     14, 3},
//    { 0, TA_MOM_TEST, 2,  16, 14, TA_SUCCESS, 1, -2.00,     14, 3},
//    { 0, TA_MOM_TEST, 2,  16, 14, TA_SUCCESS, 2, -5.22,     14, 3},
//    { 0, TA_MOM_TEST, 0,  14, 14, TA_SUCCESS, 0, -0.50,     14, 1},
//    { 0, TA_MOM_TEST, 0,  13, 14, TA_SUCCESS, 0, -0.50,     14, 0},
//    /* Middle of data test. */
//    { 0, TA_MOM_TEST, 20,  21, 14, TA_SUCCESS, 0, -4.15,    20, 2 },
//    { 0, TA_MOM_TEST, 20,  21, 14, TA_SUCCESS, 1, -5.12,    20, 2 },

func testMomentumTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

//nolint:dupl
func testMomentumInput() []float64 {
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

func TestMomentumUpdate(t *testing.T) { //nolint: funlen
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

	input := testMomentumInput()

	t.Run("length = 14", func(t *testing.T) {
		const ( // Values from index=0 to index=13 are NaN.
			i14value  = -0.50 // Index=14 value.
			i15value  = -2.00 // Index=15 value.
			i16value  = -5.22 // Index=16 value.
			i251value = -1.13 // Index=251 (last) value.
		)

		t.Parallel()
		mom := testMomentumCreate(14)

		for i := 0; i < 13; i++ {
			checkNaN(i, mom.Update(input[i]))
		}

		for i := 13; i < len(input); i++ {
			act := mom.Update(input[i])

			switch i {
			case 14:
				check(i, i14value, act)
			case 15:
				check(i, i15value, act)
			case 16:
				check(i, i16value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, mom.Update(math.NaN()))
	})
}

func TestMomentumUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		inp = 3.
		exp = 3.
	)

	time := testMomentumTime()
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
		mom := testMomentumCreate(l)
		mom.Update(0.)
		mom.Update(0.)
		check(mom.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		mom := testMomentumCreate(l)
		mom.Update(0.)
		mom.Update(0.)
		check(mom.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		mom := testMomentumCreate(l)
		mom.Update(0.)
		mom.Update(0.)
		check(mom.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		mom := testMomentumCreate(l)
		mom.Update(0.)
		mom.Update(0.)
		check(mom.UpdateTrade(&r))
	})
}

func TestMomentumIsPrimed(t *testing.T) { //nolint:funlen
	t.Parallel()

	input := testMomentumInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		mom := testMomentumCreate(1)

		check(-1, false, mom.IsPrimed())

		for i := 0; i < 1; i++ {
			mom.Update(input[i])
			check(i, false, mom.IsPrimed())
		}

		for i := 1; i < len(input); i++ {
			mom.Update(input[i])
			check(i, true, mom.IsPrimed())
		}
	})

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		mom := testMomentumCreate(2)

		check(-1, false, mom.IsPrimed())

		for i := 0; i < 2; i++ {
			mom.Update(input[i])
			check(i, false, mom.IsPrimed())
		}

		for i := 2; i < len(input); i++ {
			mom.Update(input[i])
			check(i, true, mom.IsPrimed())
		}
	})

	t.Run("length = 3", func(t *testing.T) {
		t.Parallel()
		mom := testMomentumCreate(3)

		check(-1, false, mom.IsPrimed())

		for i := 0; i < 3; i++ {
			mom.Update(input[i])
			check(i, false, mom.IsPrimed())
		}

		for i := 3; i < len(input); i++ {
			mom.Update(input[i])
			check(i, true, mom.IsPrimed())
		}
	})

	t.Run("length = 5", func(t *testing.T) {
		t.Parallel()
		mom := testMomentumCreate(5)

		check(-1, false, mom.IsPrimed())

		for i := 0; i < 5; i++ {
			mom.Update(input[i])
			check(i, false, mom.IsPrimed())
		}

		for i := 5; i < len(input); i++ {
			mom.Update(input[i])
			check(i, true, mom.IsPrimed())
		}
	})

	t.Run("length = 10", func(t *testing.T) {
		t.Parallel()
		mom := testMomentumCreate(10)

		check(-1, false, mom.IsPrimed())

		for i := 0; i < 10; i++ {
			mom.Update(input[i])
			check(i, false, mom.IsPrimed())
		}

		for i := 10; i < len(input); i++ {
			mom.Update(input[i])
			check(i, true, mom.IsPrimed())
		}
	})
}

func TestMomentumMetadata(t *testing.T) {
	t.Parallel()

	mom := testMomentumCreate(5)
	act := mom.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.Momentum, act.Identifier)
	check("Mnemonic", "mom(5)", act.Mnemonic)
	check("Description", "Momentum mom(5)", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "mom(5)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Momentum mom(5)", act.Outputs[0].Description)
}

func TestNewMomentum(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 5
		errlen                         = "invalid momentum parameters: length should be positive"
		errbc                          = "invalid momentum parameters: 9999: unknown bar component"
		errqc                          = "invalid momentum parameters: 9999: unknown quote component"
		errtc                          = "invalid momentum parameters: 9999: unknown trade component"
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

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(5, hl/2)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(5, hl/2)", mom.LineIndicator.Description)
		check("primed", false, mom.primed)
		check("lastIndex", length, mom.lastIndex)
		check("len(window)", length+1, len(mom.window))
		check("windowLength", length+1, mom.windowLength)
		check("windowCount", 0, mom.windowCount)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(1, hl/2)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(1, hl/2)", mom.LineIndicator.Description)
		check("primed", false, mom.primed)
		check("lastIndex", 1, mom.lastIndex)
		check("len(window)", 2, len(mom.window))
		check("windowLength", 2, mom.windowLength)
		check("windowCount", 0, mom.windowCount)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mom, err := New(&params)
		check("mom == nil", true, mom == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		mom, err := New(&params)
		check("mom == nil", true, mom == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		mom, err := New(&params)
		check("mom == nil", true, mom == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		mom, err := New(&params)
		check("mom == nil", true, mom == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		mom, err := New(&params)
		check("mom == nil", true, mom == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length}

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(5)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(5)", mom.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, BarComponent: entities.BarMedianPrice}

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(5, hl/2)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(5, hl/2)", mom.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, QuoteComponent: entities.QuoteBidPrice}

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(5, b)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(5, b)", mom.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, TradeComponent: entities.TradeVolume}

		mom, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "mom(5, v)", mom.LineIndicator.Mnemonic)
		check("description", "Momentum mom(5, v)", mom.LineIndicator.Description)
	})
}

func testMomentumCreate(length int) *Momentum {
	params := Params{Length: length}

	mom, _ := New(&params)

	return mom
}
