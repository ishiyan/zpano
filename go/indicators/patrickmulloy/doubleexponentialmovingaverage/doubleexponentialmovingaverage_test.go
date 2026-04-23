//nolint:testpackage
package doubleexponentialmovingaverage

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
//   /*  DEMA TEST - Metastock      */
//   /*******************************/
//
//   /* No output value. */
//   { 0, TA_ANY_MA_TEST, 0, 1, 1,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 0, 0, 0, 0},
//#ifndef TA_FUNC_NO_RANGE_CHECK
//   { 0, TA_ANY_MA_TEST, 0, 0, 251,  0, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_BAD_PARAM, 0, 0, 0, 0 },
//#endif
//
//   /* Test with period 14 */
//   { 0, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   0,  83.785, 26, 252-26 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1,  84.768, 26, 252-26 },
//   { 0, TA_ANY_MA_TEST, 0, 0, 251, 14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 252-27, 109.467, 26, 252-26 }, /* Last Value */
//
//   /* Test with 1 unstable price bar. Test for period 2, 14 */
//   { 1, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   0,  93.960, 4, 252-4 }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   1,  94.522, 4, 252-4 },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  2, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 252-5, 107.94, 4, 252-4 }, /* Last Value */
//
//   { 1, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    0,  84.91,  (13*2)+2, 252-((13*2)+2) }, /* First Value */
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    1,  84.97,  (13*2)+2, 252-((13*2)+2) },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    2,  84.80,  (13*2)+2, 252-((13*2)+2) },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,    3,  85.14,  (13*2)+2, 252-((13*2)+2) },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS,   20,  89.83,  (13*2)+2, 252-((13*2)+2) },
//   { 0, TA_ANY_MA_TEST, 1, 0, 251,  14, TA_MAType_DEMA, TA_COMPATIBILITY_METASTOCK, TA_SUCCESS, 252-((13*2)+2+1), 109.4676, (13*2)+2, 252-((13*2)+2) }, /* Last Value */

func testDoubleExponentialMovingAverageTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

func testDoubleExponentialMovingAverageInput() []float64 { //nolint:dupl
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

// Input and output data taken from Excel file describing DEMA calculations in
// Technical Analysis of Stocks & Commodities v.12:2 (72-80), Smoothing Data With Less Lag.

func testDoubleExponentialMovingAverageTascInput() []float64 { //nolint:dupl
	return []float64{
		451.61, 455.20, 453.29, 446.48, 446.17, 440.86, 441.88, 451.61, 438.43, 406.33,
		328.45, 323.30, 326.39, 322.97, 312.49, 316.47, 292.92, 302.57, 326.91, 333.19,
		330.47, 338.47, 340.14, 337.59, 344.66, 345.75, 353.27, 357.12, 363.40, 373.37,
		375.48, 381.58, 372.54, 374.64, 381.83, 373.90, 374.04, 379.23, 379.42, 372.48,
		366.03, 366.66, 376.86, 386.25, 386.92, 391.62, 394.69, 394.33, 394.59, 387.35,
		387.33, 387.71, 378.95, 377.42, 374.43, 376.51, 381.60, 383.91, 384.98, 387.71,
		385.67, 384.59, 388.59, 382.79, 381.02, 373.76, 367.58, 366.38, 373.91, 375.21,
		375.80, 377.34, 381.38, 384.74, 387.09, 391.66, 397.96, 406.35, 402.37, 407.19,
		399.96, 403.99, 405.90, 402.19, 400.94, 406.73, 410.71, 417.68, 423.76, 427.55,
		430.74, 434.83, 442.05, 445.21, 451.63, 453.65, 447.21, 448.36, 435.29, 442.42,
		448.90, 449.29, 452.82, 457.42, 462.48, 461.97, 466.75, 471.34, 471.31, 467.57,
		468.07, 472.92, 483.64, 467.29, 470.67, 452.76, 452.97, 456.19, 456.72, 456.63,
		457.10, 456.22, 443.84, 444.57, 454.82, 458.22, 439.72, 440.88, 421.33, 422.21,
		428.84, 429.01, 419.52, 431.02, 436.76, 442.16, 437.25, 435.54, 430.90, 436.31,
		425.79, 417.98, 428.61, 438.10, 448.31, 453.69, 462.13, 460.87, 467.55, 459.33,
		462.29, 460.53, 468.44, 455.27, 442.59, 417.46, 408.03, 393.49, 367.33, 381.21,
		380.38, 374.42, 362.25, 344.51, 347.36, 327.55, 337.36, 334.36, 336.45, 341.95,
		350.85, 349.04, 359.06, 371.54, 368.83, 373.60, 371.20, 367.24, 361.80, 376.99,
		394.28, 417.69, 436.80, 448.71, 448.95, 456.73, 475.11, 466.29, 464.15, 482.30,
		495.79, 501.62, 501.19, 494.64, 492.10, 493.42, 481.38, 492.67, 506.11, 498.54,
		495.07, 485.82, 475.92, 474.05, 492.71, 497.55, 492.69, 505.67, 508.31, 512.47,
		521.06, 525.68, 516.94, 516.71, 527.19, 524.48, 520.40, 519.05, 538.90, 525.13,
		540.93, 548.08, 531.29, 523.47, 523.90, 536.30, 540.90, 535.76, 565.71, 592.65,
		615.70, 626.85, 624.68, 620.21, 634.95, 636.43, 629.75, 633.47, 615.95, 618.62,
		624.28, 604.67, 590.01, 584.24, 591.81, 572.89, 578.14, 585.76, 574.43, 580.30,
		585.31, 585.43, 569.52, 554.20, 547.84, 563.35, 567.80, 570.52, 565.61, 580.83,
		573.74, 573.18, 563.70, 563.56, 573.44, 583.01, 589.12, 577.20, 571.63, 570.52,
		582.61, 597.30, 605.17, 616.82, 637.16, 642.60, 649.49, 661.60, 655.79, 661.29,
		665.88, 676.95, 677.21, 697.15, 701.64, 696.34, 700.98, 690.54, 663.61, 670.77,
		681.37, 692.78, 682.72, 681.54, 669.85, 665.26, 666.78, 658.41, 661.42, 681.44,
		676.37, 694.29, 700.53, 702.01, 693.19, 689.59, 694.81, 704.49, 705.81, 699.73,
		700.24, 704.70, 718.08, 718.26, 730.96, 734.07,
	}
}

func testDoubleExponentialMovingAverageTascExpected() []float64 { //nolint:dupl
	return []float64{
		448.35153846153900, 444.04591642925000, 440.14325876683900, 435.89797197112700, 432.27216108607300, 428.4823387986880,
		425.44660726945700, 424.32221956733400, 421.62769118311400, 414.83033726696900, 397.84560812203200, 382.1684638859920,
		368.84565318219400, 356.70263630619500, 344.61177746844000, 334.63735513171700, 322.62654116603900, 313.5428856197950,
		309.16728032640400, 306.40916999763100, 303.80344210260600, 302.85705842249500, 302.47933821157700, 301.9983608663500,
		302.78990505807500, 303.85376719743900, 306.06855288785400, 308.77561303182900, 312.25765711668100, 316.9458970680750,
		321.57591953906100, 326.71142072727700, 330.12569035653100, 333.58519070887400, 337.80215498511300, 340.5277984044680,
		343.06768310059700, 346.15558049796200, 349.01153873028100, 350.63684427250100, 351.22996990596900, 351.9095633782500,
		354.02828445734900, 357.30961166805800, 360.37719885601600, 363.82330749394400, 367.36827960468400, 370.5060772052240,
		373.36404591786600, 374.89789264177600, 376.27839007919300, 377.57679520685900, 377.49585696812600, 377.2161727060120,
		376.55094999730900, 376.26572641304400, 376.74867888791700, 377.51889916397200, 378.36695413774200, 379.5197751830860,
		380.26204411670400, 380.77401929783400, 381.80453838016000, 381.89783916607600, 381.73021452208800, 380.5468863182740,
		378.61348759035000, 376.72417599183900, 376.12187633061000, 375.77882448754900, 375.56576179912300, 375.6040783595300,
		376.22280474748100, 377.26090183044400, 378.52719410689100, 380.31178107033600, 382.80280950363100, 386.2185592056820,
		388.68968311097500, 391.57096391806000, 393.09545863514500, 395.01505721352400, 396.98253167529700, 398.1889807999980,
		399.06694346081800, 400.65594056530300, 402.61936913600500, 405.34143412934500, 408.61024741373400, 412.0365400242810,
		415.51424471768900, 419.16284342072000, 423.40690763980300, 427.59861203811200, 432.20484084578100, 436.5487004361930,
		439.44795618689600, 442.14025113812000, 442.61800606262200, 444.00456957917000, 446.10907426987700, 447.9842994202380,
		450.10276104066900, 452.59015125823600, 455.47105424993300, 457.90666488479000, 460.69902426372800, 463.7810386872970,
		466.46056524419900, 468.25088326489400, 469.85338700957900, 471.90939701285600, 475.20696719666600, 475.7478636247790,
		476.64738188687900, 474.83071437832700, 473.18207233394000, 472.11709350392900, 471.19226477584600, 470.3071248387330,
		469.54026481122900, 468.68900476125000, 466.12462403384000, 463.90794851838500, 463.36367692547200, 463.3352208853190,
		460.64352553739400, 458.38651086715700, 453.56640083647600, 449.38259511285200, 446.59240449870100, 444.1271112863010,
		440.57671800386600, 439.05789403467900, 438.53051878518600, 438.83903510058600, 438.42129375178200, 437.8127704074620,
		436.61664122799700, 436.33158320045600, 434.58627785064100, 431.92751110847700, 431.08800595307800, 431.7085234483640,
		433.73215766065500, 436.31460313670400, 439.82803250568000, 442.78284900168400, 446.36834819783800, 448.3858840068980,
		450.59792288842700, 452.30746007553500, 454.94667941935000, 455.40457430577800, 453.98746320398800, 449.1249058769170,
		443.43700745040500, 436.29198063782400, 426.19940220104200, 419.20376553191500, 412.87925527082900, 406.4284240852840,
		398.98606031668800, 389.87305401128500, 382.21779865199500, 372.63726944189300, 365.57562357895500, 358.9385986699680,
		353.41171943430700, 349.36487937940300, 347.12436559088100, 344.96571184576400, 344.56683470162000, 346.0847969047800,
		347.13926492858700, 348.84281150631800, 350.09678320049600, 350.72271622664100, 350.57390590624900, 352.6749455854390,
		357.07627619881700, 364.39354167010100, 373.68560854646800, 383.69745952526600, 392.67196847402800, 401.7848631504070,
		412.52129623701200, 420.81120775166300, 427.86271648369300, 436.69665625267000, 446.44505917530300, 455.9052122479300,
		464.20526001391000, 470.58971654845400, 475.83296273212400, 480.60577070724900, 483.05035493444200, 486.7476714258790,
		491.86738627011100, 495.25492019681100, 497.68217537294300, 498.42920662615400, 497.58813352156200, 496.4813801272450,
		498.07149481592000, 500.09582362409200, 501.12405710915600, 503.81106629945500, 506.50156901377500, 509.4116347481980,
		513.14878775920200, 517.05529733511600, 519.20495014115900, 521.00270487807500, 524.01551076430400, 526.2293264513210,
		527.53566046857800, 528.42388559740100, 531.96578568399100, 533.07598323047000, 536.23825968263300, 539.9951729115380,
		540.86520045120800, 540.44295958768200, 540.04917031737300, 541.39194812915300, 543.17146574211000, 543.9513590824800,
		548.84776223941000, 556.98145349335100, 567.43883513908500, 578.25983402526700, 587.49413439256900, 594.9774007739050,
		603.63437180989800, 611.43929555510500, 617.31490619781300, 622.95122585062700, 625.34175380619100, 627.7206193164830,
		630.51764053276000, 630.08382393199600, 627.48102410866800, 624.22079916214600, 622.28625349928600, 617.7613206415570,
		614.38468371611500, 612.37795228502400, 608.89511421428100, 606.55745406191900, 605.12394595290100, 603.8033446470740,
		600.30058855005100, 594.94307837880300, 589.21899584033200, 586.29714732194300, 584.30106513637300, 582.8872439751560,
		580.90678087274000, 581.29620835902200, 580.61608126092800, 579.91534852930600, 577.92521224483300, 576.1209367622890,
		575.91464550901100, 577.08945838235200, 579.00082986700000, 578.99453886348500, 578.18466172678900, 577.2959886714500,
		578.22170195777100, 581.13509399634300, 584.84432478937700, 589.79790660897800, 597.09481554013700, 604.3466515069590,
		611.75677616160500, 620.04441732005000, 626.54887171613800, 633.07047351406800, 639.47136354102200, 646.6832713773490,
		653.06948442562700, 661.52172642490200, 669.60421468806700, 675.95330591677200, 682.17046475665300, 686.1144724602180,
		685.67825370149000, 686.20461784550800, 688.08343693827800, 691.28514912652000, 692.60313200874900, 693.5105212726140,
		692.55507025042700, 690.95652817927700, 689.66188495954300, 687.23231081390200, 685.42060562284400, 686.5917324172170,
		686.84423902464400, 689.55973926270000, 692.80400358606700, 695.83896817833100, 697.21654979349000, 697.8619301917080,
		699.11454339900500, 701.54579865410000, 703.83467566345300, 704.94047734821500, 705.93269153694400, 707.3883885215910,
		710.53046953848900, 713.29140061860800, 717.49751354398500, 721.62112194478800,
	}
}

func TestDoubleExponentialMovingAverageUpdate(t *testing.T) { //nolint:cyclop,funlen,gocognit,gocyclo,maintidx
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

	input := testDoubleExponentialMovingAverageInput()

	t.Run("length = 2, firstIsAverage = true", func(t *testing.T) {
		const (
			i4value   = 94.013 // Index=4 value.
			i5value   = 94.539 // Index=5 value.
			i251value = 107.94 // Index=251 (last) value.
			l         = 2
			lprimed   = 2*l - 2
		)

		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, true)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, dema.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := dema.Update(input[i])

			switch i {
			case 4:
				check(i, i4value, act)
			case 5:
				check(i, i5value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, dema.Update(math.NaN()))
	})

	t.Run("length = 14, firstIsAverage = true", func(t *testing.T) { //nolint:dupl
		const (
			i28value  = 84.347   // Index=28 value.
			i29value  = 84.487   // Index=29 value.
			i30value  = 84.374   // Index=30 value.
			i31value  = 84.772   // Index=31 value.
			i48value  = 89.803   // Index=48 value.
			i251value = 109.4676 // Index=251 (last) value.
			l         = 14
			lprimed   = 2*l - 2
		)

		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, true)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, dema.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := dema.Update(input[i])

			switch i {
			case 28:
				check(i, i28value, act)
			case 29:
				check(i, i29value, act)
			case 30:
				check(i, i30value, act)
			case 31:
				check(i, i31value, act)
			case 48:
				check(i, i48value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, dema.Update(math.NaN()))
	})

	t.Run("length = 2, firstIsAverage = false (Metastock)", func(t *testing.T) {
		const (
			// The very first value is the input value.
			i4value   = 93.977 /*93.960*/ // Index=4 value.
			i5value   = 94.522 // Index=5 value.
			i251value = 107.94 // Index=251 (last) value.
			l         = 2
			lprimed   = 2*l - 2
		)

		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, dema.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := dema.Update(input[i])

			switch i {
			case 4:
				check(i, i4value, act)
			case 5:
				check(i, i5value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, dema.Update(math.NaN()))
	})

	t.Run("length = 14, firstIsAverage = false (Metastock)", func(t *testing.T) { //nolint:dupl
		const (
			// The very first value is the input value.
			i28value  = 84.87    // 84.91 // Index=28 value.
			i29value  = 84.94    // 84.97 // Index=29 value.
			i30value  = 84.77    // 84.80 // Index=30 value.
			i31value  = 85.12    // 85.14 // Index=31 value.
			i48value  = 89.83    // Index=48 value.
			i251value = 109.4676 // Index=251 (last) value.
			l         = 14
			lprimed   = 2*l - 2
		)

		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, dema.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := dema.Update(input[i])

			switch i {
			case 28:
				check(i, i28value, act)
			case 29:
				check(i, i29value, act)
			case 30:
				check(i, i30value, act)
			case 31:
				check(i, i31value, act)
			case 48:
				check(i, i48value, act)
			case 251:
				check(i, i251value, act)
			}
		}

		checkNaN(0, dema.Update(math.NaN()))
	})

	t.Run("length = 26, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()

		const (
			l          = 26
			lprimed    = 2*l - 2
			firstCheck = 216
		)

		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		in := testDoubleExponentialMovingAverageTascInput()
		exp := testDoubleExponentialMovingAverageTascExpected()
		inlen := len(in)

		for i := 0; i < lprimed; i++ {
			checkNaN(i, dema.Update(in[i]))
		}

		for i := lprimed; i < inlen; i++ {
			act := dema.Update(in[i])

			if i >= firstCheck {
				check(i, exp[i], act)
			}
		}

		checkNaN(0, dema.Update(math.NaN()))
	})
}

func TestDoubleExponentialMovingAverageUpdateEntity(t *testing.T) { //nolint:funlen
	t.Parallel()

	const (
		l        = 2
		lprimed  = 2*l - 2
		inp      = 3.
		expFalse = 2.666666666666667
	)

	time := testDoubleExponentialMovingAverageTime()
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
		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			dema.Update(0.)
		}

		check(expFalse, dema.UpdateScalar(&s))
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		b := entities.Bar{Time: time, Close: inp}
		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			dema.Update(0.)
		}

		check(expFalse, dema.UpdateBar(&b))
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		q := entities.Quote{Time: time, Bid: inp, Ask: inp}
		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			dema.Update(0.)
		}

		check(expFalse, dema.UpdateQuote(&q))
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		r := entities.Trade{Time: time, Price: inp}
		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		for i := 0; i < lprimed; i++ {
			dema.Update(0.)
		}

		check(expFalse, dema.UpdateTrade(&r))
	})
}

func TestDoubleExponentialMovingAverageIsPrimed(t *testing.T) { //nolint:dupl
	t.Parallel()

	input := testDoubleExponentialMovingAverageInput()
	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const (
		l       = 14
		lprimed = 2*l - 2
	)

	t.Run("length = 14, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, true)

		check(0, false, dema.IsPrimed())

		for i := 0; i < lprimed; i++ {
			dema.Update(input[i])
			check(i+1, false, dema.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			dema.Update(input[i])
			check(i+1, true, dema.IsPrimed())
		}
	})

	t.Run("length = 14, firstIsAverage = false (Metastock)", func(t *testing.T) {
		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(l, false)

		check(0, false, dema.IsPrimed())

		for i := 0; i < lprimed; i++ {
			dema.Update(input[i])
			check(i+1, false, dema.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			dema.Update(input[i])
			check(i+1, true, dema.IsPrimed())
		}
	})
}

func TestDoubleExponentialMovingAverageMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("length = 10, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()

		dema := testDoubleExponentialMovingAverageCreateLength(10, true)
		act := dema.Metadata()

		check("Identifier", core.DoubleExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "dema(10)", act.Mnemonic)
		check("Description", "Double exponential moving average dema(10)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "dema(10)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Double exponential moving average dema(10)", act.Outputs[0].Description)
	})

	t.Run("alpha = 2/11 = 0.18181818..., firstIsAverage = false", func(t *testing.T) {
		t.Parallel()

		// α = 2 / (ℓ + 1) = 2/11 = 0.18181818...
		const alpha = 2. / 11.

		dema := testDoubleExponentialMovingAverageCreateAlpha(alpha, false)
		act := dema.Metadata()

		check("Identifier", core.DoubleExponentialMovingAverage, act.Identifier)
		check("Mnemonic", "dema(10, 0.18181818)", act.Mnemonic)
		check("Description", "Double exponential moving average dema(10, 0.18181818)", act.Description)
		check("len(Outputs)", 1, len(act.Outputs))
		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
		check("Outputs[0].Mnemonic", "dema(10, 0.18181818)", act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", "Double exponential moving average dema(10, 0.18181818)", act.Outputs[0].Description)
	})

	t.Run("length with non-default bar component", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{
			Length: 10, FirstIsAverage: true, BarComponent: entities.BarMedianPrice,
		}

		dema, _ := NewDoubleExponentialMovingAverageLength(&params)
		act := dema.Metadata()

		check("Mnemonic", "dema(10, hl/2)", act.Mnemonic)
		check("Description", "Double exponential moving average dema(10, hl/2)", act.Description)
	})

	t.Run("alpha with non-default quote component", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2. / 11., FirstIsAverage: false, QuoteComponent: entities.QuoteBidPrice,
		}

		dema, _ := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		act := dema.Metadata()

		check("Mnemonic", "dema(10, 0.18181818, b)", act.Mnemonic)
		check("Description", "Double exponential moving average dema(10, 0.18181818, b)", act.Description)
	})
}

func TestNewDoubleExponentialMovingAverage(t *testing.T) { //nolint:funlen
	t.Parallel()

	const (
		bc     entities.BarComponent   = entities.BarMedianPrice
		qc     entities.QuoteComponent = entities.QuoteMidPrice
		tc     entities.TradeComponent = entities.TradePrice
		length                         = 10
		alpha                          = 2. / 11.

		errlen   = "invalid double exponential moving average parameters: length should be positive"
		erralpha = "invalid double exponential moving average parameters: smoothing factor should be in range [0, 1]"
		errbc    = "invalid double exponential moving average parameters: 9999: unknown bar component"
		errqc    = "invalid double exponential moving average parameters: 9999: unknown quote component"
		errtc    = "invalid double exponential moving average parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("length > 1, firstIsAverage = false", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{
			Length: length, FirstIsAverage: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", false, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", length, dema.length)
		check("length2", length+length-1, dema.length2)
		check("smoothingFactor", alpha, dema.smoothingFactor)
		check("count", 0, dema.count)
		check("sum", 0., dema.sum)
		check("ema1", 0., dema.ema1)
		check("ema2", 0., dema.ema2)
	})

	t.Run("length = 1, firstIsAverage = true", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{
			Length: 1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(1, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(1, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", true, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", 1, dema.length)
		check("length2", 1, dema.length2)
		check("smoothingFactor", 1., dema.smoothingFactor)
	})

	t.Run("length = 0", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{
			Length: 0, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("dema == nil", true, dema == nil)
		check("err", errlen, err.Error())
	})

	t.Run("length < 0", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{
			Length: -1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("dema == nil", true, dema == nil)
		check("err", errlen, err.Error())
	})

	t.Run("epsilon < α ≤ 1", func(t *testing.T) { //nolint:dupl
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, 0.18181818, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, 0.18181818, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", true, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", length, dema.length)
		check("length2", length+length-1, dema.length2)
		check("smoothingFactor", alpha, dema.smoothingFactor)
		check("count", 0, dema.count)
		check("sum", 0., dema.sum)
		check("ema1", 0., dema.ema1)
		check("ema2", 0., dema.ema2)
	})

	t.Run("0 < α < epsilon", func(t *testing.T) { //nolint:dupl
		t.Parallel()

		const (
			alpha  = 0.00000001
			length = 199999999 // 2./0.00000001 - 1.
		)

		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, FirstIsAverage: false, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(199999999, 0.00000001, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(199999999, 0.00000001, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", false, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", length, dema.length)
		check("smoothingFactor", alpha, dema.smoothingFactor)
		check("count", 0, dema.count)
		check("sum", 0., dema.sum)
		check("ema1", 0., dema.ema1)
		check("ema2", 0., dema.ema2)
	})

	t.Run("α = 0", func(t *testing.T) { //nolint:dupl
		t.Parallel()

		const (
			alpha  = 0.00000001
			length = 199999999 // 2./0.00000001 - 1.
		)

		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 0, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(199999999, 0.00000001, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(199999999, 0.00000001, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", true, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", length, dema.length)
		check("smoothingFactor", alpha, dema.smoothingFactor)
		check("count", 0, dema.count)
		check("sum", 0., dema.sum)
		check("ema1", 0., dema.ema1)
		check("ema2", 0., dema.ema2)
	})

	t.Run("α = 1", func(t *testing.T) { //nolint:dupl
		t.Parallel()

		const (
			alpha  = 1
			length = 1 // 2./1 - 1.
		)

		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: alpha, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(1, 1.00000000, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(1, 1.00000000, hl/2)", dema.LineIndicator.Description)
		check("firstIsAverage", true, dema.firstIsAverage)
		check("primed", false, dema.primed)
		check("length", length, dema.length)
		check("smoothingFactor", 1., dema.smoothingFactor)
		check("count", 0, dema.count)
		check("sum", 0., dema.sum)
		check("ema1", 0., dema.ema1)
		check("ema2", 0., dema.ema2)
	})

	t.Run("α < 0", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: -1, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("dema == nil", true, dema == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("α > 1", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 2, FirstIsAverage: true, BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("dema == nil", true, dema == nil)
		check("err", erralpha, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: entities.BarComponent(9999), QuoteComponent: qc, TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("dema == nil", true, dema == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: entities.QuoteComponent(9999), TradeComponent: tc,
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("dema == nil", true, dema == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{
			SmoothingFactor: 1, FirstIsAverage: true,
			BarComponent: bc, QuoteComponent: qc, TradeComponent: entities.TradeComponent(9999),
		}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("dema == nil", true, dema == nil)
		check("err", errtc, err.Error())
	})

	// Zero-value component mnemonic tests.
	// A zero value means "use default, don't show in mnemonic".

	t.Run("all components zero, length", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10)", dema.LineIndicator.Description)
	})

	t.Run("all components zero, alpha", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageSmoothingFactorParams{SmoothingFactor: alpha}

		dema, err := NewDoubleExponentialMovingAverageSmoothingFactor(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, 0.18181818)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, 0.18181818)", dema.LineIndicator.Description)
	})

	t.Run("only bar component set", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true, BarComponent: entities.BarMedianPrice}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, hl/2)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, hl/2)", dema.LineIndicator.Description)
	})

	t.Run("only quote component set", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true, QuoteComponent: entities.QuoteBidPrice}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, b)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, b)", dema.LineIndicator.Description)
	})

	t.Run("only trade component set", func(t *testing.T) {
		t.Parallel()
		params := DoubleExponentialMovingAverageLengthParams{Length: length, FirstIsAverage: true, TradeComponent: entities.TradeVolume}

		dema, err := NewDoubleExponentialMovingAverageLength(&params)
		check("err == nil", true, err == nil)
		check("mnemonic", "dema(10, v)", dema.LineIndicator.Mnemonic)
		check("description", "Double exponential moving average dema(10, v)", dema.LineIndicator.Description)
	})
}

func testDoubleExponentialMovingAverageCreateLength(length int, firstIsAverage bool) *DoubleExponentialMovingAverage {
	params := DoubleExponentialMovingAverageLengthParams{
		Length: length, FirstIsAverage: firstIsAverage,
	}

	dema, _ := NewDoubleExponentialMovingAverageLength(&params)

	return dema
}

func testDoubleExponentialMovingAverageCreateAlpha(alpha float64, firstIsAverage bool) *DoubleExponentialMovingAverage {
	params := DoubleExponentialMovingAverageSmoothingFactorParams{
		SmoothingFactor: alpha, FirstIsAverage: firstIsAverage,
	}

	dema, _ := NewDoubleExponentialMovingAverageSmoothingFactor(&params)

	return dema
}
