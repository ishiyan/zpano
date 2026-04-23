//nolint:testpackage
package rateofchangepercent

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
// ...
// /**********************/
// /*      ROC TEST      */
// /**********************/
// { 1, TA_ROC_TEST, 0, 251, 14, TA_SUCCESS,      0,  -0.546,  14,  252-14 }, /* First Value */
// { 0, TA_ROC_TEST, 0, 251, 14, TA_SUCCESS,      1,  -2.109,  14,  252-14 },
// { 0, TA_ROC_TEST, 0, 251, 14, TA_SUCCESS,      2,  -5.53,   14,  252-14 },
// { 0, TA_ROC_TEST, 0, 251, 14, TA_SUCCESS, 252-15,  -1.0367, 14,  252-14 }, /* Last Value */
//
// ROC% = ROC / 100, so the expected values are:
// -0.00546, -0.02109, -0.0553, -0.010367

func testRateOfChangePercentTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

//nolint:dupl
func testRateOfChangePercentInput() []float64 {
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

func TestRateOfChangePercentUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-4 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testRateOfChangePercentInput()

	t.Run("length = 14", func(t *testing.T) {
		const ( // Values from index=0 to index=13 are NaN.
			i14value  = -0.00546  // Index=14 value.
			i15value  = -0.02109  // Index=15 value.
			i16value  = -0.0553   // Index=16 value.
			i251value = -0.010367 // Index=251 (last) value.
		)

		t.Parallel()
		rocp := testRateOfChangePercentCreate(14)

		for i := 0; i < 13; i++ {
			checkNaN(i, rocp.Update(input[i]))
		}

		for i := 13; i < len(input); i++ {
			act := rocp.Update(input[i])

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

		checkNaN(0, rocp.Update(math.NaN()))
	})
}

func TestRateOfChangePercentUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		inp = 3.
		exp = 0.
	)

	time := testRateOfChangePercentTime()
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
		rocp := testRateOfChangePercentCreate(l)
		rocp.Update(inp)
		rocp.Update(inp)
		check(rocp.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		rocp := testRateOfChangePercentCreate(l)
		rocp.Update(inp)
		rocp.Update(inp)
		check(rocp.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		rocp := testRateOfChangePercentCreate(l)
		rocp.Update(inp)
		rocp.Update(inp)
		check(rocp.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		rocp := testRateOfChangePercentCreate(l)
		rocp.Update(inp)
		rocp.Update(inp)
		check(rocp.UpdateTrade(&r))
	})
}

func TestRateOfChangePercentIsPrimed(t *testing.T) { //nolint:funlen
	t.Parallel()

	input := testRateOfChangePercentInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		rocp := testRateOfChangePercentCreate(1)

		check(-1, false, rocp.IsPrimed())

		for i := 0; i < 1; i++ {
			rocp.Update(input[i])
			check(i, false, rocp.IsPrimed())
		}

		for i := 1; i < len(input); i++ {
			rocp.Update(input[i])
			check(i, true, rocp.IsPrimed())
		}
	})

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		rocp := testRateOfChangePercentCreate(2)

		check(-1, false, rocp.IsPrimed())

		for i := 0; i < 2; i++ {
			rocp.Update(input[i])
			check(i, false, rocp.IsPrimed())
		}

		for i := 2; i < len(input); i++ {
			rocp.Update(input[i])
			check(i, true, rocp.IsPrimed())
		}
	})

	t.Run("length = 5", func(t *testing.T) {
		t.Parallel()
		rocp := testRateOfChangePercentCreate(5)

		check(-1, false, rocp.IsPrimed())

		for i := 0; i < 5; i++ {
			rocp.Update(input[i])
			check(i, false, rocp.IsPrimed())
		}

		for i := 5; i < len(input); i++ {
			rocp.Update(input[i])
			check(i, true, rocp.IsPrimed())
		}
	})

	t.Run("length = 10", func(t *testing.T) {
		t.Parallel()
		rocp := testRateOfChangePercentCreate(10)

		check(-1, false, rocp.IsPrimed())

		for i := 0; i < 10; i++ {
			rocp.Update(input[i])
			check(i, false, rocp.IsPrimed())
		}

		for i := 10; i < len(input); i++ {
			rocp.Update(input[i])
			check(i, true, rocp.IsPrimed())
		}
	})
}

func TestRateOfChangePercentMetadata(t *testing.T) {
	t.Parallel()

	rocp := testRateOfChangePercentCreate(5)
	act := rocp.Metadata()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	check("Identifier", core.RateOfChangePercent, act.Identifier)
	check("Mnemonic", "rocp(5)", act.Mnemonic)
	check("Description", "Rate of Change Percent rocp(5)", act.Description)
	check("len(Outputs)", 1, len(act.Outputs))
	check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
	check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
	check("Outputs[0].Mnemonic", "rocp(5)", act.Outputs[0].Mnemonic)
	check("Outputs[0].Description", "Rate of Change Percent rocp(5)", act.Outputs[0].Description)
}

func TestNewRateOfChangePercent(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 5
		errlen                         = "invalid rate of change percent parameters: length should be positive"
		errbc                          = "invalid rate of change percent parameters: 9999: unknown bar component"
		errqc                          = "invalid rate of change percent parameters: 9999: unknown quote component"
		errtc                          = "invalid rate of change percent parameters: 9999: unknown trade component"
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

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(5, hl/2)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(5, hl/2)", rocp.LineIndicator.Description)
		check("primed", false, rocp.primed)
		check("lastIndex", length, rocp.lastIndex)
		check("len(window)", length+1, len(rocp.window))
		check("windowLength", length+1, rocp.windowLength)
		check("windowCount", 0, rocp.windowCount)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(1, hl/2)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(1, hl/2)", rocp.LineIndicator.Description)
		check("primed", false, rocp.primed)
		check("lastIndex", 1, rocp.lastIndex)
		check("len(window)", 2, len(rocp.window))
		check("windowLength", 2, rocp.windowLength)
		check("windowCount", 0, rocp.windowCount)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocp, err := New(&params)
		check("rocp == nil", true, rocp == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocp, err := New(&params)
		check("rocp == nil", true, rocp == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		rocp, err := New(&params)
		check("rocp == nil", true, rocp == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		rocp, err := New(&params)
		check("rocp == nil", true, rocp == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		rocp, err := New(&params)
		check("rocp == nil", true, rocp == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length}

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(5)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(5)", rocp.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, BarComponent: entities.BarMedianPrice}

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(5, hl/2)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(5, hl/2)", rocp.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, QuoteComponent: entities.QuoteBidPrice}

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(5, b)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(5, b)", rocp.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, TradeComponent: entities.TradeVolume}

		rocp, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocp(5, v)", rocp.LineIndicator.Mnemonic)
		check("description", "Rate of Change Percent rocp(5, v)", rocp.LineIndicator.Description)
	})
}

func testRateOfChangePercentCreate(length int) *RateOfChangePercent {
	params := Params{Length: length}

	rocp, _ := New(&params)

	return rocp
}
