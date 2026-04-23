//nolint:testpackage
package t3exponentialmovingaverage

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
// /************/
// /*  T3 TEST */
// /************/
// { 1, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,      0,  85.73, 24,  252-24  }, /* First Value */
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,      1,  84.37, 24,  252-24  },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 252-26, 109.03, 24,  252-24  },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 252-25, 108.88, 24,  252-24  }, /* Last Value */

func testT3ExponentialMovingAverageTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func testT3ExponentialMovingAverageInput() []float64 { //nolint:dupl
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

// Expected data is taken from the TA-L:ib (http://ta-lib.org/): test_T3.xls, T3(5,0.7) column.

//nolint:lll
func testT3ExponentialMovingAverageExpected() []float64 {
	return []float64{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		85.72987567371790, 84.36536238957790, 83.48212833499610, 83.64878022075270, 84.20774780003610, 84.81400226243770,
		85.21162459867090, 85.61977953584720, 85.88366564657920, 86.37507196062170, 86.96360606377290, 87.33715584797100,
		87.47581538320310, 87.22704246772630, 86.66138343888210, 85.93988420535540, 85.17149180117680, 84.72447668908230,
		85.01865520943790, 85.83612757497730, 87.03563064369770, 88.27257526381620, 89.41758978061620, 90.07559145334440,
		90.52788384532180, 90.79849305637170, 90.75816999615070, 90.48012595902110, 89.56721885323770, 88.19961294746030,
		86.64012904487460, 85.41932552383290, 84.73734286979860, 84.55421441417930, 85.00890603797630, 85.87022403441470,
		86.77016654750900, 87.51893615115180, 88.47034581811800, 89.45474018834570, 90.52574408904200, 91.57169411417090,
		92.42614882677630, 92.87068016379410, 92.77362350686670, 92.29390466156780, 91.56310318186090, 90.34706085030460,
		88.70815109494770, 87.17399400070270, 86.08896648247290, 86.77721477010180, 89.01690920285690, 92.50578045524320,
		96.44661245628320, 99.68374118082500, 101.92353915902500, 103.51864520203800, 104.78791133759700, 105.75424271690200,
		106.42923203569800, 106.67331521810400, 107.05684930109700, 107.65552066563500, 108.45831824703700, 109.54712933016300,
		111.89434730682000, 114.52574405349600, 116.75228613125600, 118.41726647291900, 119.37242887772400, 119.58121746519900,
		119.15016873725100, 117.99712010891400, 116.35343432693000, 115.50133109382700, 115.19900149595300, 115.20832852971600,
		114.89045855286400, 114.49505935654600, 114.04873581207800, 113.98299653094000, 114.76120631137200, 115.59884277390800,
		116.28205585323000, 116.55693390132200, 116.39983049399400, 116.13322166220300, 115.93071030281200, 116.38166637287000,
		117.22757950324600, 118.23631535248800, 119.65852250290400, 121.06189350073800, 122.17001559831100, 122.88208870818400,
		123.32075513993400, 123.49845436008200, 123.73526507154800, 124.57544794795000, 125.98775939563200, 127.72677347766200,
		129.27852779306300, 130.66084445385100, 131.92850098151000, 133.38763285852100, 134.90304940294100, 136.26189527105800,
		137.25635611457500, 137.76465314041100, 137.88397712608300, 137.57294795595100, 136.27097486999100, 134.48409991908300,
		132.08334163347200, 129.64905881153300, 127.34920616402000, 125.78806210911800, 125.20525121026800, 124.97471519964800,
		124.91187781656600, 124.53236452951700, 123.58691150647000, 122.26362413973900, 121.44605573567300, 121.22904968429300,
		121.26833012683900, 121.06480537490900, 121.11936168993500, 121.16873672766900, 121.42906776612200, 122.30150235050600,
		123.65994185435500, 124.60369529647400, 124.92167475815800, 124.65108838183800, 124.37460034158500, 123.94458586126200,
		123.47997231186300, 123.13057512793100, 123.03814664421900, 123.04147396438600, 123.23225566873300, 123.84790124323500,
		124.53741993908800, 125.48193174896000, 126.88001089241300, 128.29156445167800, 129.94307874395900, 131.63064617976500,
		132.80966556857300, 133.55644670051600, 133.80913491375100, 133.48924486083300, 132.29305889081000, 131.18723345437400,
		130.07504168974000, 128.82765488259600, 127.23236569221000, 125.90878699219100, 124.77446308473800, 123.92185993031100,
		122.95122696884300, 122.05570316734200, 120.96585380865100, 120.08120106718900, 119.74698117219900, 119.52440496847700,
		118.98986994944600, 117.91920934928700, 116.69443222363700, 115.09004689050000, 112.81604242902500, 110.56231508883200,
		108.80117430219600, 107.52097068017500, 106.69052592372500, 106.21332619981500, 104.12646621492700, 101.28478214721100,
		98.45909805366300, 96.29196542448850, 94.59404389669660, 93.56678784171560, 93.49539073445030, 93.90336034189440,
		94.25829840322930, 94.40922959431940, 94.07888457524260, 93.30020230129970, 92.78070057794470, 92.57408795701960,
		92.98810688816150, 93.56080911108470, 94.18829647483600, 94.55572153094430, 94.73726505383050, 94.70162919926040,
		95.03014981975750, 96.34501640344490, 98.68718019257840, 101.19176302440600, 103.18034967184000, 104.57401565486800,
		105.34469118495800, 105.51082025160400, 105.32911720489200, 105.20951060163700, 105.97637243377600, 107.82138781238700,
		110.21769922957300, 112.77811195271300, 114.51075261504700, 114.85539096796300, 114.29440592244600, 113.27318258016400,
		111.89052001075400, 110.70018738987500, 109.94200614625900, 109.47043806636400, 109.29969950462100, 109.08364080071000,
		108.87294256387700, 108.83218146601300, 108.93829305405100, 109.02476604337200, 109.03210341275800, 108.87915000449300,
	}
}

func TestT3ExponentialMovingAverageUpdate(t *testing.T) { //nolint: funlen, cyclop
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-3 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testT3ExponentialMovingAverageInput()

	const (
		l       = 5
		lprimed = 6*l - 6
	)

	t.Run("length = 5, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		const (
			i250value = 109.032 // Index=250 value.
			i251value = 108.88  // Index=251 (last) value.
		)

		t3 := testT3ExponentialMovingAverageCreateLength(l, true, 0.7)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, t3.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := t3.Update(input[i])

			switch i {
			case 250:
				check(i, i250value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, t3.Update(math.NaN()))
	})

	t.Run("length = 5, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()

		const (
			i24value  = 85.749  // Index=24 value.
			i25value  = 84.380  // Index=25 value.
			i250value = 109.032 // Index=250 value.
			i251value = 108.88  // Index=251 (last) value.
		)

		t3 := testT3ExponentialMovingAverageCreateLength(l, false, 0.7)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, t3.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := t3.Update(input[i])

			switch i {
			case 24:
				check(i, i24value, act)
			case 25:
				check(i, i25value, act)
			case 250:
				check(i, i250value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, t3.Update(math.NaN()))
	})

	t.Run("length = 5, firstIsAverage = true (t3.xls)", func(t *testing.T) {
		t.Parallel()

		const (
			firstCheck = lprimed
		)

		t3 := testT3ExponentialMovingAverageCreateLength(l, true, 0.7)

		exp := testT3ExponentialMovingAverageExpected()

		for i := 0; i < lprimed; i++ {
			checkNaN(i, t3.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := t3.Update(input[i])

			if i >= firstCheck {
				check(i, exp[i], act)
			}
		}

		checkNaN(0, t3.Update(math.NaN()))
	})
}

func TestT3ExponentialMovingAverageUpdateEntity(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		l        = 2
		lprimed  = 6*l - 6
		inp      = 3.
		expFalse = 1.6675884773662544
		expTrue  = 1.6901728395061721
	)

	time := testT3ExponentialMovingAverageTime()
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
		t3 := testT3ExponentialMovingAverageCreateLength(l, false, 0.7)

		for i := 0; i < lprimed; i++ {
			t3.Update(0.)
		}

		check(expFalse, t3.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		t3 := testT3ExponentialMovingAverageCreateLength(l, true, 0.7)

		for i := 0; i < lprimed; i++ {
			t3.Update(0.)
		}

		check(expTrue, t3.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		t3 := testT3ExponentialMovingAverageCreateLength(l, false, 0.7)

		for i := 0; i < lprimed; i++ {
			t3.Update(0.)
		}

		check(expFalse, t3.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		t3 := testT3ExponentialMovingAverageCreateLength(l, true, 0.7)

		for i := 0; i < lprimed; i++ {
			t3.Update(0.)
		}

		check(expTrue, t3.UpdateTrade(&r))
	})
}

func TestT3ExponentialMovingAverageIsPrimed(t *testing.T) {
	t.Parallel()

	input := testT3ExponentialMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const (
		l       = 5
		lprimed = 6*l - 6
	)

	t.Run("length = 5, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		t3 := testT3ExponentialMovingAverageCreateLength(l, true, 0.7)

		check(0, false, t3.IsPrimed())

		for i := 0; i < lprimed; i++ {
			t3.Update(input[i])
			check(i+1, false, t3.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			t3.Update(input[i])
			check(i+1, true, t3.IsPrimed())
		}
	})

	t.Run("length = 5, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()

		t3 := testT3ExponentialMovingAverageCreateLength(l, false, 0.7)

		check(0, false, t3.IsPrimed())

		for i := 0; i < lprimed; i++ {
			t3.Update(input[i])
			check(i+1, false, t3.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			t3.Update(input[i])
			check(i+1, true, t3.IsPrimed())
		}
	})
}

func TestT3ExponentialMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("length = 10, v=0.3333, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		t3 := testT3ExponentialMovingAverageCreateLength(10, true, 0.3333)
		act := t3.Metadata()

		check("Identifier", core.T3ExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "t3(10, 0.33330000)", act.Mnemonic)
		check("Description", "T3 exponential moving average t3(10, 0.33330000)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "t3(10, 0.33330000)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "T3 exponential moving average t3(10, 0.33330000)", act.Outputs[0].Description)
	})

	t.Run("alpha = 2/11 = 0.18181818..., v=0.3333333, firstIsAverage = false", func(t *testing.T) {
		t.Parallel()

		// α = 2 / (ℓ + 1) = 2/11 = 0.18181818...
		const alpha = 2. / 11.

		t3 := testT3ExponentialMovingAverageCreateAlpha(alpha, false, 0.3333333)
		act := t3.Metadata()

		check("Identifier", core.T3ExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "t3(10, 0.18181818, 0.33333330)", act.Mnemonic)
		check("Description", "T3 exponential moving average t3(10, 0.18181818, 0.33333330)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "t3(10, 0.18181818, 0.33333330)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "T3 exponential moving average t3(10, 0.18181818, 0.33333330)", act.Outputs[0].Description)
	})

	t.Run("length with non-default bar component", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageLengthParams{
			Length: 10, VolumeFactor: 0.7, FirstIsAverage: true, BarComponent: entities.BarMedianPrice,
		}

		t3, _ := NewT3ExponentialMovingAverageLength(&params)
		act := t3.Metadata()

		check("Mnemonic", "t3(10, 0.70000000, hl/2)", act.Mnemonic)
		check("Description", "T3 exponential moving average t3(10, 0.70000000, hl/2)", act.Description)
	})

	t.Run("alpha with non-default quote component", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2. / 11., VolumeFactor: 0.7, FirstIsAverage: false, QuoteComponent: entities.QuoteBidPrice,
		}

		t3, _ := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		act := t3.Metadata()

		check("Mnemonic", "t3(10, 0.18181818, 0.70000000, b)", act.Mnemonic)
		check("Description", "T3 exponential moving average t3(10, 0.18181818, 0.70000000, b)", act.Description)
	})
}

func TestNewT3ExponentialMovingAverage(t *testing.T) { //nolint: funlen, maintidx
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 10
		alpha                          = 2. / 11.

		errlen   = "invalid t3 exponential moving average parameters: length should be greater than 1"
		erralpha = "invalid t3 exponential moving average parameters: smoothing factor should be in range [0, 1]"
		errvol   = "invalid t3 exponential moving average parameters: volume factor should be in range [0, 1]"
		errbc    = "invalid t3 exponential moving average parameters: 9999: unknown bar component"
		errqc    = "invalid t3 exponential moving average parameters: 9999: unknown quote component"
		errtc    = "invalid t3 exponential moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(
		t3 *T3ExponentialMovingAverage, mnemonic string, length int, alpha float64, firstIsAverage bool,
	) {
		check("mnemonic", mnemonic, t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average "+mnemonic, t3.LineIndicator.Description)
		check("firstIsAverage", firstIsAverage, t3.firstIsAverage)
		check("primed", false, t3.primed)
		check("length", length, t3.length)
		check("length2", length+length-1, t3.length2)
		check("length3", length+length+length-2, t3.length3)
		check("length4", length+length+length+length-3, t3.length4)
		check("length5", length+length+length+length+length-4, t3.length5)
		check("length6", length+length+length+length+length+length-5, t3.length6)
		check("smoothingFactor", alpha, t3.smoothingFactor)
		check("count", 0, t3.count)
		check("sum", 0., t3.sum)
		check("ema1", 0., t3.ema1)
		check("ema2", 0., t3.ema2)
		check("ema3", 0., t3.ema3)
		check("ema4", 0., t3.ema4)
		check("ema5", 0., t3.ema5)
		check("ema6", 0., t3.ema6)
	}

	t.Run("length > 1, firstIsAverage = false", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0.7, FirstIsAverage: false,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(10, 0.70000000, hl/2)", length, alpha, false)
	})

	t.Run("length = 1, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: 1, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: 0, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: -1, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errlen, err.Error())
	})

	t.Run("epsilon < α ≤ 1", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(10, 0.18181818, 0.70000000, hl/2)", length, alpha, true)
	})

	t.Run("0 < α < epsilon", func(t *testing.T) {
		t.Parallel()

		const (
			alpha  = 0.00000001
			length = 199999999 // 2./0.00000001 - 1.
		)

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, VolumeFactor: 0.7, FirstIsAverage: false,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(199999999, 0.00000001, 0.70000000, hl/2)", length, alpha, false)
	})

	t.Run("α = 0", func(t *testing.T) {
		t.Parallel()

		const (
			alpha  = 0.00000001
			length = 199999999 // 2./0.00000001 - 1.
		)

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 0, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(199999999, 0.00000001, 0.70000000, hl/2)", length, alpha, true)
	})

	t.Run("α = 1", func(t *testing.T) {
		t.Parallel()

		const (
			alpha  = 1
			length = 1 // 2./1 - 1.
		)

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(1, 1.00000000, 0.70000000, hl/2)", length, alpha, true)
	})

	t.Run("α < 0", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: -1, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("α > 1", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2, VolumeFactor: 0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("volume factor = 0.5", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0.5, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(10, 0.50000000, hl/2)", length, alpha, true)
	})

	t.Run("volume factor = 0", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(10, 0.00000000, hl/2)", length, alpha, true)
	})

	t.Run("volume factor = 1", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		checkInstance(t3, "t3(10, 1.00000000, hl/2)", length, alpha, true)
	})

	t.Run("volume factor < 0", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: 3, VolumeFactor: -0.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errvol, err.Error())
	})

	t.Run("volume factor > 1", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageLengthParams{
			Length: 3, VolumeFactor: 1.7, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errvol, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, VolumeFactor: 0.5, FirstIsAverage: true,
			BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, VolumeFactor: 0.5, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		params := T3ExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, VolumeFactor: 0.5, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("t3 == nil", true, t3 == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero, length", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageLengthParams{Length: length, VolumeFactor: 0.7, FirstIsAverage: true}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "t3(10, 0.70000000)", t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average t3(10, 0.70000000)", t3.LineIndicator.Description)
	})

	t.Run("all components zero, alpha", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageSmoothingFactorParams{SmoothingFactor: alpha, VolumeFactor: 0.7}

		t3, err := NewT3ExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "t3(10, 0.18181818, 0.70000000)", t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average t3(10, 0.18181818, 0.70000000)", t3.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0.7, FirstIsAverage: true, BarComponent: entities.BarMedianPrice,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "t3(10, 0.70000000, hl/2)", t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average t3(10, 0.70000000, hl/2)", t3.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0.7, FirstIsAverage: true, QuoteComponent: entities.QuoteBidPrice,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "t3(10, 0.70000000, b)", t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average t3(10, 0.70000000, b)", t3.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := T3ExponentialMovingAverageLengthParams{
			Length: length, VolumeFactor: 0.7, FirstIsAverage: true, TradeComponent: entities.TradeVolume,
		}

		t3, err := NewT3ExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "t3(10, 0.70000000, v)", t3.LineIndicator.Mnemonic)
		check("description", "T3 exponential moving average t3(10, 0.70000000, v)", t3.LineIndicator.Description)
	})
}

func testT3ExponentialMovingAverageCreateLength(
	length int, firstIsAverage bool, volume float64,
) *T3ExponentialMovingAverage {
	params := T3ExponentialMovingAverageLengthParams{
		Length: length, VolumeFactor: volume, FirstIsAverage: firstIsAverage,
	}

	t3, _ := NewT3ExponentialMovingAverageLength(&params)

	return t3
}

func testT3ExponentialMovingAverageCreateAlpha(
	alpha float64, firstIsAverage bool, volume float64,
) *T3ExponentialMovingAverage {
	params := T3ExponentialMovingAverageSmoothingFactorParams{
		SmoothingFactor: alpha, VolumeFactor: volume, FirstIsAverage: firstIsAverage,
	}

	t3, _ := NewT3ExponentialMovingAverageSmoothingFactor(&params)

	return t3
}
