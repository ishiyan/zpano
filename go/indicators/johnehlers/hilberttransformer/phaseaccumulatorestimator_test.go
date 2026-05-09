//nolint:testpackage
package hilberttransformer

import (
	"math"
	"testing"
)


//nolint:funlen, cyclop
func TestPhaseAccumulatorEstimatorUpdate(t *testing.T) {
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

		pae := testPhaseAccumulatorEstimatorCreateDefault()
		exp := testSharedExpectedSmoothed()

		const (
			lprimed        = 3
			notPrimedValue = 0
			index          = 99999
		)

		for i := range lprimed {
			pae.Update(input[i])
			check(i, notPrimedValue, pae.Smoothed())
		}

		for i := lprimed; i < len(input); i++ {
			pae.Update(input[i])
			check(i, exp[i], pae.Smoothed())
		}

		previous := pae.Smoothed()
		pae.Update(math.NaN())
		check(index, previous, pae.Smoothed())
	})

	t.Run("reference implementation: detrended (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		pae := testPhaseAccumulatorEstimatorCreateDefault()
		exp := testSharedExpectedDetrended()

		const (
			lprimed        = 9
			notPrimedValue = 0
			index          = 99999

			// This should have been len(input), but after 24, the calculated
			// period is different from the expected data produced by homodyne
			// discriminator. This makes the detrended data also different.
			last = 24
		)

		for i := range lprimed {
			pae.Update(input[i])
			check(i, notPrimedValue, pae.Detrended())
		}

		for i := lprimed; i < last; i++ {
			pae.Update(input[i])
			check(i, exp[i], pae.Detrended())
		}

		previous := pae.Detrended()
		pae.Update(math.NaN())
		check(index, previous, pae.Detrended())
	})

	t.Run("reference implementation: quadrature (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		pae := testPhaseAccumulatorEstimatorCreateDefault()
		exp := testSharedExpectedQuadrature()

		const (
			lprimed        = 15
			notPrimedValue = 0
			index          = 99999

			// This should have been len(input), but after 24, the calculated
			// period is different from the expected data produced by homodyne
			// discriminator. This makes the quadrature data also different.
			last = 24
		)

		for i := range lprimed {
			pae.Update(input[i])
			check(i, notPrimedValue, pae.Quadrature())
		}

		for i := lprimed; i < last; i++ {
			pae.Update(input[i])
			check(i, exp[i], pae.Quadrature())
		}

		previous := pae.Quadrature()
		pae.Update(math.NaN())
		check(index, previous, pae.Quadrature())
	})

	t.Run("reference implementation: in-phase (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		pae := testPhaseAccumulatorEstimatorCreateDefault()
		exp := testSharedExpectedInPhase()

		const (
			lprimed        = 15
			notPrimedValue = 0
			index          = 99999

			// This should have been len(input), but after 24, the calculated
			// period is different from the expected data produced by homodyne
			// discriminator. This makes the in-phase data also different.
			last = 24
		)

		for i := range lprimed {
			pae.Update(input[i])
			check(i, notPrimedValue, pae.InPhase())
		}

		for i := lprimed; i < last; i++ {
			pae.Update(input[i])
			check(i, exp[i], pae.InPhase())
		}

		previous := pae.InPhase()
		pae.Update(math.NaN())
		check(index, previous, pae.InPhase())
	})

	t.Run("reference implementation: period (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		pae := testPhaseAccumulatorEstimatorCreateDefault()
		exp := testSharedExpectedPeriod()

		const (
			lprimed        = 18
			notPrimedValue = 6
			index          = 99999

			// This should have been len(input), but after 18, the calculated
			// period is different from the expected data produced by homodyne
			// discriminator.
			last = 18
		)

		for i := range lprimed {
			pae.Update(input[i])
			check(i, notPrimedValue, pae.Period())
		}

		for i := lprimed; i < last; i++ {
			pae.Update(input[i])
			check(i, exp[i], pae.Period())
		}

		previous := pae.Period()
		pae.Update(math.NaN())
		check(index, previous, pae.Period())
	})
}

//nolint:dupl
func TestPhaseAccumulatorEstimatorPeriod(t *testing.T) {
	t.Parallel()

	check := func(exp, act, epsilon float64) {
		t.Helper()

		if math.Abs(exp-act) > epsilon {
			t.Errorf("period is incorrect: expected %v, actual %v", exp, act)
		}
	}

	update := func(omega float64) *PhaseAccumulatorEstimator {
		t.Helper()

		const updates = 512

		pae := testPhaseAccumulatorEstimatorCreateDefault()
		for i := range updates {
			pae.Update(math.Sin(omega * float64(i)))
		}

		return pae
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
			epsilon = 1e-14
		)

		pae := update(omega)
		check(float64(pae.MinPeriod()), float64(pae.Period()), epsilon)
	})

	t.Run("max period of sin input", func(t *testing.T) {
		t.Parallel()

		const (
			period  = 60
			omega   = 2 * math.Pi / period
			epsilon = 12.5e0
		)

		pae := update(omega)
		check(float64(pae.MaxPeriod()), float64(pae.Period()), epsilon)
	})
}

func TestPhaseAccumulatorEstimatorPrimed(t *testing.T) {
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

		pae := testPhaseAccumulatorEstimatorCreateDefault()

		const lprimed = 4 + 7*2

		check(0, false, pae.Primed())

		for i := range lprimed {
			pae.Update(input[i])
			check(i+1, false, pae.Primed())
		}

		for i := lprimed; i < len(input); i++ {
			pae.Update(input[i])
			check(i+1, true, pae.Primed())
		}
	})

	t.Run("reference implementation: primed with warmup", func(t *testing.T) {
		t.Parallel()

		const lprimed = 50

		pae := testPhaseAccumulatorEstimatorCreateWarmUp(lprimed)

		check(0, false, pae.Primed())

		for i := range lprimed {
			pae.Update(input[i])
			check(i+1, false, pae.Primed())
		}

		for i := lprimed; i < len(input); i++ {
			pae.Update(input[i])
			check(i+1, true, pae.Primed())
		}
	})
}

func TestNewPhaseAccumulatorEstimator(t *testing.T) { //nolint: funlen, maintidx
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

	checkInstance := func(pae *PhaseAccumulatorEstimator,
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
		smoothingLengthPlus2HtLengthMin1 := smoothingLengthPlus2HtLengthMin2 + 1
		smoothingLengthPlus2HtLength := smoothingLengthPlus2HtLengthMin1 + 1
		warmUpPeriod := max(warmUp, smoothingLengthPlus2HtLength)

		check("smoothingLength", length, pae.smoothingLength)
		check("SmoothingLength()", length, pae.SmoothingLength())
		check("minPeriod", defaultMinPeriod, pae.minPeriod)
		check("MinPeriod()", defaultMinPeriod, pae.MinPeriod())
		check("maxPeriod", defaultMaxPeriod, pae.maxPeriod)
		check("MaxPeriod()", defaultMaxPeriod, pae.MaxPeriod())
		check("warmUpPeriod", warmUpPeriod, pae.warmUpPeriod)
		check("WarmUpPeriod()", warmUpPeriod, pae.WarmUpPeriod())
		check("alphaEmaQuadratureInPhase", alphaQuadratureInPhase, pae.alphaEmaQuadratureInPhase)
		check("AlphaEmaQuadratureInPhase()", alphaQuadratureInPhase, pae.AlphaEmaQuadratureInPhase())
		check("oneMinAlphaEmaQuadratureInPhase", 1-alphaQuadratureInPhase, pae.oneMinAlphaEmaQuadratureInPhase)
		check("alphaEmaPeriod", alphaPeriod, pae.alphaEmaPeriod)
		check("AlphaEmaPeriod()", alphaPeriod, pae.AlphaEmaPeriod())
		check("oneMinAlphaEmaPeriod", 1-alphaPeriod, pae.oneMinAlphaEmaPeriod)
		check("smoothingLengthPlusHtLengthMin1", smoothingLengthPlusHtLengthMin1, pae.smoothingLengthPlusHtLengthMin1)
		check("smoothingLengthPlus2HtLengthMin2", smoothingLengthPlus2HtLengthMin2, pae.smoothingLengthPlus2HtLengthMin2)
		check("smoothingLengthPlus2HtLengthMin1", smoothingLengthPlus2HtLengthMin1, pae.smoothingLengthPlus2HtLengthMin1)
		check("smoothingLengthPlus2HtLength", smoothingLengthPlus2HtLength, pae.smoothingLengthPlus2HtLength)
		check("len(wmaSmoothed)", htLength, len(pae.wmaSmoothed))
		check("len(detrended)", htLength, len(pae.detrended))
		check("len(deltaPhase)", accumulationLength, len(pae.deltaPhase))
		check("len(rawValues)", length, len(pae.rawValues))
		check("len(wmaFactors)", length, len(pae.wmaFactors))
		check("isPrimed", false, pae.isPrimed)
		check("isWarmedUp", false, pae.isWarmedUp)
		check("period", float64(defaultMinPeriod), pae.period)
		check("Period()", float64(defaultMinPeriod), pae.Period())
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("err == nil", true, err == nil)
		checkInstance(pae, params.SmoothingLength, params.AlphaEmaQuadratureInPhase,
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("err == nil", true, err == nil)
		checkInstance(pae, params.SmoothingLength, params.AlphaEmaQuadratureInPhase,
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
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

		pae, err := NewPhaseAccumulatorEstimator(&params)
		check("pae == nil", true, pae == nil)
		check("err", errap, err.Error())
	})
}

func testPhaseAccumulatorEstimatorCreateDefault() *PhaseAccumulatorEstimator {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.15,
		AlphaEmaPeriod:            0.25,
	}

	pae, _ := NewPhaseAccumulatorEstimator(&params)

	return pae
}

func testPhaseAccumulatorEstimatorCreateWarmUp(warmUp int) *PhaseAccumulatorEstimator {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.15,
		AlphaEmaPeriod:            0.25,
		WarmUpPeriod:              warmUp,
	}

	pae, _ := NewPhaseAccumulatorEstimator(&params)

	return pae
}
