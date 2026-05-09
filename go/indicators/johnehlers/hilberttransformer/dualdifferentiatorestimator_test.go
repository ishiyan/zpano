//nolint:testpackage
package hilberttransformer

import (
	"math"
	"testing"
)


//nolint:funlen, cyclop
func TestDualDifferentiatorEstimatorUpdate(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-8 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	input := testSharedInput()

	t.Run("reference implementation: wma smoothed (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		dde := testDualDifferentiatorEstimatorCreateDefault()
		exp := testSharedExpectedSmoothed()

		const (
			lprimed        = 3
			notPrimedValue = 0
			index          = 99999
		)

		for i := range lprimed {
			dde.Update(input[i])
			check(i, notPrimedValue, dde.Smoothed())
		}

		for i := lprimed; i < len(input); i++ {
			dde.Update(input[i])
			check(i, exp[i], dde.Smoothed())
		}

		previous := dde.Smoothed()
		dde.Update(math.NaN())
		check(index, previous, dde.Smoothed())
	})

	t.Run("reference implementation: detrended (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		dde := testDualDifferentiatorEstimatorCreateDefault()
		exp := testSharedExpectedDetrended()

		const (
			lprimed        = 9
			notPrimedValue = 0
			index          = 99999

			// This should have been len(input), but after 23, the calculated
			// period is different from the expected data produced by homodyne
			// discriminator. This makes the detrended data also different.
			last = 23
		)

		for i := range lprimed {
			dde.Update(input[i])
			check(i, notPrimedValue, dde.Detrended())
		}

		for i := lprimed; i < last; i++ {
			dde.Update(input[i])
			check(i, exp[i], dde.Detrended())
		}

		previous := dde.Detrended()
		dde.Update(math.NaN())
		check(index, previous, dde.Detrended())
	})

	t.Run("reference implementation: quadrature (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		dde := testDualDifferentiatorEstimatorCreateDefault()
		exp := testSharedExpectedQuadrature()

		const (
			lprimed        = 15
			notPrimedValue = 0
			index          = 99999

			// This should have been len(input), but after 23, the calculated
			// period is different from the expected data produced by homodyne
			// discriminator. This makes the quadrature data also different.
			last = 23
		)

		for i := range lprimed {
			dde.Update(input[i])
			check(i, notPrimedValue, dde.Quadrature())
		}

		for i := lprimed; i < last; i++ {
			dde.Update(input[i])
			check(i, exp[i], dde.Quadrature())
		}

		previous := dde.Quadrature()
		dde.Update(math.NaN())
		check(index, previous, dde.Quadrature())
	})

	t.Run("reference implementation: in-phase (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		dde := testDualDifferentiatorEstimatorCreateDefault()
		exp := testSharedExpectedInPhase()

		const (
			lprimed        = 15
			notPrimedValue = 0
			index          = 99999

			// This should have been len(input), but after 23, the calculated
			// period is different from the expected data produced by homodyne
			// discriminator. This makes the in-phase data also different.
			last = 23
		)

		for i := range lprimed {
			dde.Update(input[i])
			check(i, notPrimedValue, dde.InPhase())
		}

		for i := lprimed; i < last; i++ {
			dde.Update(input[i])
			check(i, exp[i], dde.InPhase())
		}

		previous := dde.InPhase()
		dde.Update(math.NaN())
		check(index, previous, dde.InPhase())
	})

	t.Run("reference implementation: period (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		dde := testDualDifferentiatorEstimatorCreateDefault()
		exp := testSharedExpectedPeriod()

		const (
			lprimed        = 18
			notPrimedValue = 6
			index          = 99999

			// This should have been len(input), but after 23, the calculated
			// period is different from the expected data produced by homodyne
			// discriminator.
			last = 23
		)

		for i := range lprimed {
			dde.Update(input[i])
			check(i, notPrimedValue, dde.Period())
		}

		for i := lprimed; i < last; i++ {
			dde.Update(input[i])
			check(i, exp[i], dde.Period())
		}

		previous := dde.Period()
		dde.Update(math.NaN())
		check(index, previous, dde.Period())
	})
}

//nolint:dupl
func TestDualDifferentiatorEstimatorPeriod(t *testing.T) {
	t.Parallel()

	check := func(exp, act, epsilon float64) {
		t.Helper()

		if math.Abs(exp-act) > epsilon {
			t.Errorf("period is incorrect: expected %v, actual %v", exp, act)
		}
	}

	update := func(omega float64) *DualDifferentiatorEstimator {
		t.Helper()

		const updates = 512

		dde := testDualDifferentiatorEstimatorCreateDefault()
		for i := range updates {
			dde.Update(math.Sin(omega * float64(i)))
		}

		return dde
	}

	t.Run("period of sin input", func(t *testing.T) {
		t.Parallel()

		const (
			period  = 30
			omega   = 2 * math.Pi / period
			epsilon = 1e0
		)

		check(period, update(omega).Period(), epsilon)
	})

	t.Run("min period of sin input", func(t *testing.T) {
		t.Parallel()

		const (
			period  = 3
			omega   = 2 * math.Pi / period
			epsilon = 1.5e0
		)

		dde := update(omega)
		check(float64(dde.MinPeriod()), float64(dde.Period()), epsilon)
	})

	t.Run("max period of sin input", func(t *testing.T) {
		t.Parallel()

		const (
			period  = 60
			omega   = 2 * math.Pi / period
			epsilon = 1e0
		)

		dde := update(omega)
		check(float64(dde.MaxPeriod()), float64(dde.Period()), epsilon)
	})
}

func TestDualDifferentiatorEstimatorPrimed(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	input := testSharedInput()

	t.Run("reference implementation: primed (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		dde := testDualDifferentiatorEstimatorCreateDefault()

		const lprimed = 3 + 7*3

		check(0, false, dde.Primed())

		for i := range lprimed {
			dde.Update(input[i])
			check(i+1, false, dde.Primed())
		}

		for i := lprimed; i < len(input); i++ {
			dde.Update(input[i])
			check(i+1, true, dde.Primed())
		}
	})

	t.Run("reference implementation: primed with warmup", func(t *testing.T) {
		t.Parallel()

		const lprimed = 50

		dde := testDualDifferentiatorEstimatorCreateWarmUp(lprimed)

		check(0, false, dde.Primed())

		for i := range lprimed {
			dde.Update(input[i])
			check(i+1, false, dde.Primed())
		}

		for i := lprimed; i < len(input); i++ {
			dde.Update(input[i])
			check(i+1, true, dde.Primed())
		}
	})
}

func TestNewDualDifferentiatorEstimator(t *testing.T) { //nolint: funlen, maintidx
	t.Parallel()

	const (
		errle = "invalid cycle estimator parameters: SmoothingLength should be in range [2, 4]"
		erraq = "invalid cycle estimator parameters: AlphaEmaQuadratureInPhase should be in range (0, 1)"
		errap = "invalid cycle estimator parameters: AlphaEmaPeriod should be in range (0, 1)"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(dde *DualDifferentiatorEstimator,
		length int, alphaQuadratureInPhase, alphaPeriod float64, warmUp int,
	) {
		t.Helper()

		const (
			c4  = 4
			c02 = 0.2
		)

		if length == 0 {
			length = c4
		}

		if math.IsNaN(alphaQuadratureInPhase) {
			alphaQuadratureInPhase = c02
		}

		if math.IsNaN(alphaPeriod) {
			alphaPeriod = c02
		}

		smoothingLengthPlusHtLengthMin1 := length + htLength - 1
		smoothingLengthPlus2HtLengthMin2 := smoothingLengthPlusHtLengthMin1 + htLength - 1
		smoothingLengthPlus3HtLengthMin3 := smoothingLengthPlus2HtLengthMin2 + htLength - 1
		smoothingLengthPlus3HtLengthMin2 := smoothingLengthPlus3HtLengthMin3 + 1
		smoothingLengthPlus3HtLengthMin1 := smoothingLengthPlus3HtLengthMin2 + 1
		warmUpPeriod := max(warmUp, smoothingLengthPlus3HtLengthMin1)

		check("smoothingLength", length, dde.smoothingLength)
		check("SmoothingLength()", length, dde.SmoothingLength())
		check("minPeriod", defaultMinPeriod, dde.minPeriod)
		check("MinPeriod()", defaultMinPeriod, dde.MinPeriod())
		check("maxPeriod", defaultMaxPeriod, dde.maxPeriod)
		check("MaxPeriod()", defaultMaxPeriod, dde.MaxPeriod())
		check("warmUpPeriod", warmUpPeriod, dde.warmUpPeriod)
		check("WarmUpPeriod()", warmUpPeriod, dde.WarmUpPeriod())
		check("alphaEmaQuadratureInPhase", alphaQuadratureInPhase, dde.alphaEmaQuadratureInPhase)
		check("AlphaEmaQuadratureInPhase()", alphaQuadratureInPhase, dde.AlphaEmaQuadratureInPhase())
		check("oneMinAlphaEmaQuadratureInPhase", 1-alphaQuadratureInPhase, dde.oneMinAlphaEmaQuadratureInPhase)
		check("alphaEmaPeriod", alphaPeriod, dde.alphaEmaPeriod)
		check("AlphaEmaPeriod()", alphaPeriod, dde.AlphaEmaPeriod())
		check("oneMinAlphaEmaPeriod", 1-alphaPeriod, dde.oneMinAlphaEmaPeriod)
		check("smoothingLengthPlusHtLengthMin1", smoothingLengthPlusHtLengthMin1, dde.smoothingLengthPlusHtLengthMin1)
		check("smoothingLengthPlus2HtLengthMin2", smoothingLengthPlus2HtLengthMin2, dde.smoothingLengthPlus2HtLengthMin2)
		check("smoothingLengthPlus3HtLengthMin3", smoothingLengthPlus3HtLengthMin3, dde.smoothingLengthPlus3HtLengthMin3)
		check("smoothingLengthPlus3HtLengthMin2", smoothingLengthPlus3HtLengthMin2, dde.smoothingLengthPlus3HtLengthMin2)
		check("smoothingLengthPlus3HtLengthMin1", smoothingLengthPlus3HtLengthMin1, dde.smoothingLengthPlus3HtLengthMin1)
		check("len(wmaSmoothed)", htLength, len(dde.wmaSmoothed))
		check("len(detrended)", htLength, len(dde.detrended))
		check("len(inPhase)", htLength, len(dde.inPhase))
		check("len(quadrature)", htLength, len(dde.quadrature))
		check("len(jInPhase)", htLength, len(dde.jInPhase))
		check("len(jQuadrature)", htLength, len(dde.jQuadrature))
		check("len(rawValues)", length, len(dde.rawValues))
		check("len(wmaFactors)", length, len(dde.wmaFactors))
		check("isPrimed", false, dde.isPrimed)
		check("isWarmedUp", false, dde.isWarmedUp)
		check("period", float64(defaultMinPeriod), dde.period)
		check("Period()", float64(defaultMinPeriod), dde.Period())
	}

	t.Run("dafault (4, 0.2, 0.2)", func(t *testing.T) {
		t.Parallel()

		const (
			c4  = 4
			c02 = 0.2
		)

		params := CycleEstimatorParams{
			SmoothingLength:           c4,
			AlphaEmaQuadratureInPhase: c02,
			AlphaEmaPeriod:            c02,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("err == nil", true, err == nil)
		checkInstance(dde, params.SmoothingLength, params.AlphaEmaQuadratureInPhase,
			params.AlphaEmaPeriod, 0)
	})

	t.Run("with warm-up (3, 0.11, 0.12, 44)", func(t *testing.T) {
		t.Parallel()

		const (
			c3   = 3
			c011 = 0.11
			c012 = 0.12
			c44  = 44
		)

		params := CycleEstimatorParams{
			SmoothingLength:           c3,
			AlphaEmaQuadratureInPhase: c011,
			AlphaEmaPeriod:            c012,
			WarmUpPeriod:              c44,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("err == nil", true, err == nil)
		checkInstance(dde, params.SmoothingLength, params.AlphaEmaQuadratureInPhase,
			params.AlphaEmaPeriod, params.WarmUpPeriod)
	})

	t.Run("smoothing length = 1, error", func(t *testing.T) {
		t.Parallel()

		const c02 = 0.2

		params := CycleEstimatorParams{
			SmoothingLength:           1,
			AlphaEmaQuadratureInPhase: c02,
			AlphaEmaPeriod:            c02,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", errle, err.Error())
	})

	t.Run("smoothing length = 0, error", func(t *testing.T) {
		t.Parallel()

		const c02 = 0.2

		params := CycleEstimatorParams{
			SmoothingLength:           0,
			AlphaEmaQuadratureInPhase: c02,
			AlphaEmaPeriod:            c02,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", errle, err.Error())
	})

	t.Run("smoothing length < 0, error", func(t *testing.T) {
		t.Parallel()

		const c02 = 0.2

		params := CycleEstimatorParams{
			SmoothingLength:           -1,
			AlphaEmaQuadratureInPhase: c02,
			AlphaEmaPeriod:            c02,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", errle, err.Error())
	})

	t.Run("smoothing length > 4, error", func(t *testing.T) {
		t.Parallel()

		const (
			c5  = 5
			c02 = 0.2
		)

		params := CycleEstimatorParams{
			SmoothingLength:           c5,
			AlphaEmaQuadratureInPhase: c02,
			AlphaEmaPeriod:            c02,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", errle, err.Error())
	})

	t.Run("quad α = 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			c4  = 4
			c02 = 0.2
			c00 = 0.0
		)

		params := CycleEstimatorParams{
			SmoothingLength:           c4,
			AlphaEmaQuadratureInPhase: c00,
			AlphaEmaPeriod:            c02,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", erraq, err.Error())
	})

	t.Run("period α = 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			c4  = 4
			c02 = 0.2
			c00 = 0.0
		)

		params := CycleEstimatorParams{
			SmoothingLength:           c4,
			AlphaEmaQuadratureInPhase: c02,
			AlphaEmaPeriod:            c00,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", errap, err.Error())
	})

	t.Run("quad α < 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			c4   = 4
			c02  = 0.2
			cneg = -0.01
		)

		params := CycleEstimatorParams{
			SmoothingLength:           c4,
			AlphaEmaQuadratureInPhase: cneg,
			AlphaEmaPeriod:            c02,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", erraq, err.Error())
	})

	t.Run("period α < 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			c4   = 4
			c02  = 0.2
			cneg = -0.01
		)

		params := CycleEstimatorParams{
			SmoothingLength:           c4,
			AlphaEmaQuadratureInPhase: c02,
			AlphaEmaPeriod:            cneg,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", errap, err.Error())
	})

	t.Run("quad α = 1, error", func(t *testing.T) {
		t.Parallel()

		const (
			c4  = 4
			c02 = 0.2
			c10 = 1.0
		)

		params := CycleEstimatorParams{
			SmoothingLength:           c4,
			AlphaEmaQuadratureInPhase: c10,
			AlphaEmaPeriod:            c02,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", erraq, err.Error())
	})

	t.Run("period α = 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			c4  = 4
			c02 = 0.2
			c10 = 1.0
		)

		params := CycleEstimatorParams{
			SmoothingLength:           c4,
			AlphaEmaQuadratureInPhase: c02,
			AlphaEmaPeriod:            c10,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", errap, err.Error())
	})

	t.Run("quad α > 1, error", func(t *testing.T) {
		t.Parallel()

		const (
			c4   = 4
			c02  = 0.2
			c101 = 1.01
		)

		params := CycleEstimatorParams{
			SmoothingLength:           c4,
			AlphaEmaQuadratureInPhase: c101,
			AlphaEmaPeriod:            c02,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", erraq, err.Error())
	})

	t.Run("period α = 0, error", func(t *testing.T) {
		t.Parallel()

		const (
			c4   = 4
			c02  = 0.2
			c101 = 1.01
		)

		params := CycleEstimatorParams{
			SmoothingLength:           c4,
			AlphaEmaQuadratureInPhase: c02,
			AlphaEmaPeriod:            c101,
		}

		dde, err := NewDualDifferentiatorEstimator(&params)
		check("dde == nil", true, dde == nil)
		check("err", errap, err.Error())
	})
}

func testDualDifferentiatorEstimatorCreateDefault() *DualDifferentiatorEstimator {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.15,
		AlphaEmaPeriod:            0.15,
	}

	dde, _ := NewDualDifferentiatorEstimator(&params)

	return dde
}

func testDualDifferentiatorEstimatorCreateWarmUp(warmUp int) *DualDifferentiatorEstimator {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.15,
		AlphaEmaPeriod:            0.15,
		WarmUpPeriod:              warmUp,
	}

	dde, _ := NewDualDifferentiatorEstimator(&params)

	return dde
}
