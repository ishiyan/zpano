//nolint:testpackage
package hilberttransformer

import (
	"math"
	"testing"
)


//nolint:funlen, cyclop, dupl
func TestHomodyneDiscriminatorEstimatorUpdate(t *testing.T) {
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

		hde := testHomodyneDiscriminatorEstimatorCreateDefault()
		exp := testSharedExpectedSmoothed()

		const (
			lprimed        = 3
			notPrimedValue = 0
			index          = 99999
		)

		for i := range lprimed {
			hde.Update(input[i])
			check(i, notPrimedValue, hde.Smoothed())
		}

		for i := lprimed; i < len(input); i++ {
			hde.Update(input[i])
			check(i, exp[i], hde.Smoothed())
		}

		previous := hde.Smoothed()
		hde.Update(math.NaN())
		check(index, previous, hde.Smoothed())
	})

	t.Run("reference implementation: detrended (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		hde := testHomodyneDiscriminatorEstimatorCreateDefault()
		exp := testSharedExpectedDetrended()

		const (
			lprimed        = 9
			notPrimedValue = 0
			index          = 99999
		)

		for i := range lprimed {
			hde.Update(input[i])
			check(i, notPrimedValue, hde.Detrended())
		}

		for i := lprimed; i < len(input); i++ {
			hde.Update(input[i])
			check(i, exp[i], hde.Detrended())
		}

		previous := hde.Detrended()
		hde.Update(math.NaN())
		check(index, previous, hde.Detrended())
	})

	t.Run("reference implementation: quadrature (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		hde := testHomodyneDiscriminatorEstimatorCreateDefault()
		exp := testSharedExpectedQuadrature()

		const (
			lprimed        = 15
			notPrimedValue = 0
			index          = 99999
		)

		for i := range lprimed {
			hde.Update(input[i])
			check(i, notPrimedValue, hde.Quadrature())
		}

		for i := lprimed; i < len(input); i++ {
			hde.Update(input[i])
			check(i, exp[i], hde.Quadrature())
		}

		previous := hde.Quadrature()
		hde.Update(math.NaN())
		check(index, previous, hde.Quadrature())
	})

	t.Run("reference implementation: in-phase (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		hde := testHomodyneDiscriminatorEstimatorCreateDefault()
		exp := testSharedExpectedInPhase()

		const (
			lprimed        = 15
			notPrimedValue = 0
			index          = 99999
		)

		for i := range lprimed {
			hde.Update(input[i])
			check(i, notPrimedValue, hde.InPhase())
		}

		for i := lprimed; i < len(input); i++ {
			hde.Update(input[i])
			check(i, exp[i], hde.InPhase())
		}

		previous := hde.InPhase()
		hde.Update(math.NaN())
		check(index, previous, hde.InPhase())
	})

	t.Run("reference implementation: period (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		hde := testHomodyneDiscriminatorEstimatorCreateDefault()
		exp := testSharedExpectedPeriod()

		const (
			lprimed        = 23
			notPrimedValue = 6
			index          = 99999
		)

		for i := range lprimed {
			hde.Update(input[i])
			check(i, notPrimedValue, hde.Period())
		}

		for i := lprimed; i < len(input); i++ {
			hde.Update(input[i])
			check(i, exp[i], hde.Period())
		}

		previous := hde.Period()
		hde.Update(math.NaN())
		check(index, previous, hde.Period())
	})
}

//nolint:dupl
func TestHomodyneDiscriminatorEstimatorPeriod(t *testing.T) {
	t.Parallel()

	check := func(exp, act, epsilon float64) {
		t.Helper()

		if math.Abs(exp-act) > epsilon {
			t.Errorf("period is incorrect: expected %v, actual %v", exp, act)
		}
	}

	update := func(omega float64) *HomodyneDiscriminatorEstimator {
		t.Helper()

		const updates = 512

		hde := testHomodyneDiscriminatorEstimatorCreateDefault()
		for i := range updates {
			hde.Update(math.Sin(omega * float64(i)))
		}

		return hde
	}

	t.Run("period of sin input", func(t *testing.T) {
		t.Parallel()

		const (
			period  = 30
			omega   = 2 * math.Pi / period
			epsilon = 1e-2
		)

		check(period, update(omega).Period(), epsilon)
	})

	t.Run("min period of sin input", func(t *testing.T) {
		t.Parallel()

		const (
			period  = 3
			omega   = 2 * math.Pi / period
			epsilon = 1e-14
		)

		hde := update(omega)
		check(float64(hde.MinPeriod()), float64(hde.Period()), epsilon)
	})

	t.Run("max period of sin input", func(t *testing.T) {
		t.Parallel()

		const (
			period  = 60
			omega   = 2 * math.Pi / period
			epsilon = 1e-14
		)

		hde := update(omega)
		check(float64(hde.MaxPeriod()), float64(hde.Period()), epsilon)
	})
}

func TestHomodyneDiscriminatorEstimatorPrimed(t *testing.T) {
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

		hde := testHomodyneDiscriminatorEstimatorCreateDefault()

		const lprimed = 4 + 7*3

		check(0, false, hde.Primed())

		for i := range lprimed {
			hde.Update(input[i])
			check(i+1, false, hde.Primed())
		}

		for i := lprimed; i < len(input); i++ {
			hde.Update(input[i])
			check(i+1, true, hde.Primed())
		}
	})

	t.Run("reference implementation: primed with warmup", func(t *testing.T) {
		t.Parallel()

		const lprimed = 50

		hde := testHomodyneDiscriminatorEstimatorCreateWarmUp(lprimed)

		check(0, false, hde.Primed())

		for i := range lprimed {
			hde.Update(input[i])
			check(i+1, false, hde.Primed())
		}

		for i := lprimed; i < len(input); i++ {
			hde.Update(input[i])
			check(i+1, true, hde.Primed())
		}
	})
}

func TestNewHomodyneDiscriminatorEstimator(t *testing.T) { //nolint: funlen, maintidx
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

	checkInstance := func(hde *HomodyneDiscriminatorEstimator,
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
		smoothingLengthPlus3HtLength := smoothingLengthPlus3HtLengthMin1 + 1
		warmUpPeriod := max(warmUp, smoothingLengthPlus3HtLength)

		check("smoothingLength", length, hde.smoothingLength)
		check("SmoothingLength()", length, hde.SmoothingLength())
		check("minPeriod", defaultMinPeriod, hde.minPeriod)
		check("MinPeriod()", defaultMinPeriod, hde.MinPeriod())
		check("maxPeriod", defaultMaxPeriod, hde.maxPeriod)
		check("MaxPeriod()", defaultMaxPeriod, hde.MaxPeriod())
		check("warmUpPeriod", warmUpPeriod, hde.warmUpPeriod)
		check("WarmUpPeriod()", warmUpPeriod, hde.WarmUpPeriod())
		check("alphaEmaQuadratureInPhase", alphaQuadratureInPhase, hde.alphaEmaQuadratureInPhase)
		check("AlphaEmaQuadratureInPhase()", alphaQuadratureInPhase, hde.AlphaEmaQuadratureInPhase())
		check("oneMinAlphaEmaQuadratureInPhase", 1-alphaQuadratureInPhase, hde.oneMinAlphaEmaQuadratureInPhase)
		check("alphaEmaPeriod", alphaPeriod, hde.alphaEmaPeriod)
		check("AlphaEmaPeriod()", alphaPeriod, hde.AlphaEmaPeriod())
		check("oneMinAlphaEmaPeriod", 1-alphaPeriod, hde.oneMinAlphaEmaPeriod)
		check("smoothingLengthPlusHtLengthMin1", smoothingLengthPlusHtLengthMin1, hde.smoothingLengthPlusHtLengthMin1)
		check("smoothingLengthPlus2HtLengthMin2", smoothingLengthPlus2HtLengthMin2, hde.smoothingLengthPlus2HtLengthMin2)
		check("smoothingLengthPlus3HtLengthMin3", smoothingLengthPlus3HtLengthMin3, hde.smoothingLengthPlus3HtLengthMin3)
		check("smoothingLengthPlus3HtLengthMin2", smoothingLengthPlus3HtLengthMin2, hde.smoothingLengthPlus3HtLengthMin2)
		check("smoothingLengthPlus3HtLengthMin1", smoothingLengthPlus3HtLengthMin1, hde.smoothingLengthPlus3HtLengthMin1)
		check("smoothingLengthPlus3HtLength", smoothingLengthPlus3HtLength, hde.smoothingLengthPlus3HtLength)
		check("len(wmaSmoothed)", htLength, len(hde.wmaSmoothed))
		check("len(detrended)", htLength, len(hde.detrended))
		check("len(inPhase)", htLength, len(hde.inPhase))
		check("len(quadrature)", htLength, len(hde.quadrature))
		check("len(jInPhase)", htLength, len(hde.jInPhase))
		check("len(jQuadrature)", htLength, len(hde.jQuadrature))
		check("len(rawValues)", length, len(hde.rawValues))
		check("len(wmaFactors)", length, len(hde.wmaFactors))
		check("isPrimed", false, hde.isPrimed)
		check("isWarmedUp", false, hde.isWarmedUp)
		check("period", float64(defaultMinPeriod), hde.period)
		check("Period()", float64(defaultMinPeriod), hde.Period())
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("err == nil", true, err == nil)
		checkInstance(hde, params.SmoothingLength, params.AlphaEmaQuadratureInPhase,
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("err == nil", true, err == nil)
		checkInstance(hde, params.SmoothingLength, params.AlphaEmaQuadratureInPhase,
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
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

		hde, err := NewHomodyneDiscriminatorEstimator(&params)
		check("hde == nil", true, hde == nil)
		check("err", errap, err.Error())
	})
}

func testHomodyneDiscriminatorEstimatorCreateDefault() *HomodyneDiscriminatorEstimator {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.2,
		AlphaEmaPeriod:            0.2,
	}

	hde, _ := NewHomodyneDiscriminatorEstimator(&params)

	return hde
}

func testHomodyneDiscriminatorEstimatorCreateWarmUp(warmUp int) *HomodyneDiscriminatorEstimator {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.2,
		AlphaEmaPeriod:            0.2,
		WarmUpPeriod:              warmUp,
	}

	hde, _ := NewHomodyneDiscriminatorEstimator(&params)

	return hde
}
