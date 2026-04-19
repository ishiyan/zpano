//nolint:testpackage
package rateofchangeratio

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
// Output data, length=14.
// Taken from TA-Lib (http://ta-lib.org/) tests, test_mom.c.
//
// ROCR TEST (price/prevPrice):
// { 1, TA_ROCR_TEST, 0, 251, 14, TA_SUCCESS,      0, 0.994536,  14,  252-14 }, /* First Value */
// { 0, TA_ROCR_TEST, 0, 251, 14, TA_SUCCESS,      1, 0.978906,  14,  252-14 },
// { 0, TA_ROCR_TEST, 0, 251, 14, TA_SUCCESS,      2, 0.944689,  14,  252-14 },
// { 0, TA_ROCR_TEST, 0, 251, 14, TA_SUCCESS, 252-15, 0.989633,  14,  252-14 }, /* Last Value */
//
// ROCR100 TEST (price/prevPrice)*100:
// { 1, TA_ROCR100_TEST, 0, 251, 14, TA_SUCCESS,      0, 99.4536,  14,  252-14 }, /* First Value */
// { 0, TA_ROCR100_TEST, 0, 251, 14, TA_SUCCESS,      1, 97.8906,  14,  252-14 },
// { 0, TA_ROCR100_TEST, 0, 251, 14, TA_SUCCESS,      2, 94.4689,  14,  252-14 },
// { 0, TA_ROCR100_TEST, 0, 251, 14, TA_SUCCESS, 252-15, 98.9633,  14,  252-14 }, /* Last Value */

func testRateOfChangeRatioTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

//nolint:dupl
func testRateOfChangeRatioInput() []float64 {
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

func TestRateOfChangeRatioUpdate(t *testing.T) { //nolint: funlen
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

	input := testRateOfChangeRatioInput()

	t.Run("ROCR length = 14", func(t *testing.T) {
		const ( // Values from index=0 to index=13 are NaN.
			i14value  = 0.994536 // Index=14 value (output index 0).
			i15value  = 0.978906 // Index=15 value (output index 1).
			i16value  = 0.944689 // Index=16 value (output index 2).
			i251value = 0.989633 // Index=251 (last) value (output index 237).
		)

		t.Parallel()
		rocr := testRateOfChangeRatioCreate(14, false)

		for i := 0; i < 13; i++ {
			checkNaN(i, rocr.Update(input[i]))
		}

		for i := 13; i < len(input); i++ {
			act := rocr.Update(input[i])

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

		checkNaN(0, rocr.Update(math.NaN()))
	})

	t.Run("ROCR100 length = 14", func(t *testing.T) {
		const ( // Values from index=0 to index=13 are NaN.
			i14value  = 99.4536 // Index=14 value (output index 0).
			i15value  = 97.8906 // Index=15 value (output index 1).
			i16value  = 94.4689 // Index=16 value (output index 2).
			i251value = 98.9633 // Index=251 (last) value (output index 237).
		)

		t.Parallel()
		rocr100 := testRateOfChangeRatioCreate(14, true)

		for i := 0; i < 13; i++ {
			checkNaN(i, rocr100.Update(input[i]))
		}

		for i := 13; i < len(input); i++ {
			act := rocr100.Update(input[i])

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

		checkNaN(0, rocr100.Update(math.NaN()))
	})

	t.Run("ROCR middle of data", func(t *testing.T) {
		// { 0, TA_ROCR_TEST, 20, 21, 14, TA_SUCCESS, 0, 0.955096, 20, 2 },
		// { 0, TA_ROCR_TEST, 20, 21, 14, TA_SUCCESS, 1, 0.944744, 20, 2 },
		// Output index 0 corresponds to input index 20, output index 1 to input index 21.
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(14, false)

		for i := 0; i < len(input); i++ {
			act := rocr.Update(input[i])

			switch i {
			case 20:
				check(i, 0.955096, act)
			case 21:
				check(i, 0.944744, act)
			}
		}
	})

	t.Run("ROCR100 middle of data", func(t *testing.T) {
		// { 0, TA_ROCR100_TEST, 20, 21, 14, TA_SUCCESS, 0, 95.5096, 20, 2 },
		// { 0, TA_ROCR100_TEST, 20, 21, 14, TA_SUCCESS, 1, 94.4744, 20, 2 },
		t.Parallel()
		rocr100 := testRateOfChangeRatioCreate(14, true)

		for i := 0; i < len(input); i++ {
			act := rocr100.Update(input[i])

			switch i {
			case 20:
				check(i, 95.5096, act)
			case 21:
				check(i, 94.4744, act)
			}
		}
	})
}

func TestRateOfChangeRatioUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l   = 2
		inp = 3.
		exp = 1.
	)

	time := testRateOfChangeRatioTime()
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
		rocr := testRateOfChangeRatioCreate(l, false)
		rocr.Update(inp)
		rocr.Update(inp)
		check(rocr.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		rocr := testRateOfChangeRatioCreate(l, false)
		rocr.Update(inp)
		rocr.Update(inp)
		check(rocr.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		rocr := testRateOfChangeRatioCreate(l, false)
		rocr.Update(inp)
		rocr.Update(inp)
		check(rocr.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		rocr := testRateOfChangeRatioCreate(l, false)
		rocr.Update(inp)
		rocr.Update(inp)
		check(rocr.UpdateTrade(&r))
	})
}

func TestRateOfChangeRatioIsPrimed(t *testing.T) { //nolint:funlen
	t.Parallel()

	input := testRateOfChangeRatioInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(1, false)

		check(-1, false, rocr.IsPrimed())

		for i := 0; i < 1; i++ {
			rocr.Update(input[i])
			check(i, false, rocr.IsPrimed())
		}

		for i := 1; i < len(input); i++ {
			rocr.Update(input[i])
			check(i, true, rocr.IsPrimed())
		}
	})

	t.Run("length = 2", func(t *testing.T) {
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(2, false)

		check(-1, false, rocr.IsPrimed())

		for i := 0; i < 2; i++ {
			rocr.Update(input[i])
			check(i, false, rocr.IsPrimed())
		}

		for i := 2; i < len(input); i++ {
			rocr.Update(input[i])
			check(i, true, rocr.IsPrimed())
		}
	})

	t.Run("length = 5", func(t *testing.T) {
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(5, false)

		check(-1, false, rocr.IsPrimed())

		for i := 0; i < 5; i++ {
			rocr.Update(input[i])
			check(i, false, rocr.IsPrimed())
		}

		for i := 5; i < len(input); i++ {
			rocr.Update(input[i])
			check(i, true, rocr.IsPrimed())
		}
	})

	t.Run("length = 10", func(t *testing.T) {
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(10, false)

		check(-1, false, rocr.IsPrimed())

		for i := 0; i < 10; i++ {
			rocr.Update(input[i])
			check(i, false, rocr.IsPrimed())
		}

		for i := 10; i < len(input); i++ {
			rocr.Update(input[i])
			check(i, true, rocr.IsPrimed())
		}
	})
}

func TestRateOfChangeRatioMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("ROCR", func(t *testing.T) {
		t.Parallel()
		rocr := testRateOfChangeRatioCreate(5, false)
		act := rocr.Metadata()

		check("Type", core.RateOfChangeRatio, act.Type)
		check("Mnemonic", "rocr(5)", act.Mnemonic)
		check("Description", "Rate of Change Ratio rocr(5)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Type", outputs.ScalarType, act.Outputs[0].Type)
		check("Outputs[0].Mnemonic", "rocr(5)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Rate of Change Ratio rocr(5)", act.Outputs[0].Description)
	})

	t.Run("ROCR100", func(t *testing.T) {
		t.Parallel()
		rocr100 := testRateOfChangeRatioCreate(5, true)
		act := rocr100.Metadata()

		check("Type", core.RateOfChangeRatio, act.Type)
		check("Mnemonic", "rocr100(5)", act.Mnemonic)
		check("Description", "Rate of Change Ratio 100 Scale rocr100(5)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Type", outputs.ScalarType, act.Outputs[0].Type)
		check("Outputs[0].Mnemonic", "rocr100(5)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Rate of Change Ratio 100 Scale rocr100(5)", act.Outputs[0].Description)
	})
}

func TestNewRateOfChangeRatio(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 5
		errlen                         = "invalid rate of change ratio parameters: length should be positive"
		errbc                          = "invalid rate of change ratio parameters: 9999: unknown bar component"
		errqc                          = "invalid rate of change ratio parameters: 9999: unknown quote component"
		errtc                          = "invalid rate of change ratio parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("ROCR length > 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(5, hl/2)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(5, hl/2)", rocr.LineIndicator.Description)
		check("primed", false, rocr.primed)
		check("hundredScale", false, rocr.hundredScale)
		check("lastIndex", length, rocr.lastIndex)
		check("len(window)", length+1, len(rocr.window))
		check("windowLength", length+1, rocr.windowLength)
		check("windowCount", 0, rocr.windowCount)
	})

	t.Run("ROCR100 length > 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, HundredScale: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr100(5, hl/2)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio 100 Scale rocr100(5, hl/2)", rocr.LineIndicator.Description)
		check("primed", false, rocr.primed)
		check("hundredScale", true, rocr.hundredScale)
	})

	t.Run("length = 1", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(1, hl/2)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(1, hl/2)", rocr.LineIndicator.Description)
		check("primed", false, rocr.primed)
		check("lastIndex", 1, rocr.lastIndex)
		check("len(window)", 2, len(rocr.window))
		check("windowLength", 2, rocr.windowLength)
		check("windowCount", 0, rocr.windowCount)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: 0, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("rocr == nil", true, rocr == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: -1, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("rocr == nil", true, rocr == nil)
		check("err", errlen, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("rocr == nil", true, rocr == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		rocr, err := New(&params)
		check("rocr == nil", true, rocr == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := Params{
			Length: length, BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		rocr, err := New(&params)
		check("rocr == nil", true, rocr == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	t.Run("all components zero", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(5)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(5)", rocr.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, BarComponent: entities.BarMedianPrice}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(5, hl/2)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(5, hl/2)", rocr.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, QuoteComponent: entities.QuoteBidPrice}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(5, b)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(5, b)", rocr.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := Params{Length: length, TradeComponent: entities.TradeVolume}

		rocr, err := New(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "rocr(5, v)", rocr.LineIndicator.Mnemonic)
		check("description", "Rate of Change Ratio rocr(5, v)", rocr.LineIndicator.Description)
	})
}

func testRateOfChangeRatioCreate(length int, hundredScale bool) *RateOfChangeRatio {
	params := Params{Length: length, HundredScale: hundredScale}

	rocr, _ := New(&params)

	return rocr
}
