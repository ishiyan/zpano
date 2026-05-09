//nolint:testpackage
package hilberttransformer

import (
	"math"
	"testing"
)


//nolint:funlen, cyclop, dupl
func TestHomodyneDiscriminatorEstimatorUnrolledUpdate(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-8 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	input := testHomodyneDiscriminatorEstimatorUnrolledInput()

	t.Run("reference implementation: wma smoothed (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		hdeu := testHomodyneDiscriminatorEstimatorUnrolledCreateDefault()
		exp := testHomodyneDiscriminatorEstimatorUnrolledExpectedSmoothed()

		const (
			lprimed        = 3
			notPrimedValue = 0
			index          = 99999
		)

		for i := range lprimed {
			hdeu.Update(input[i])
			check(i, notPrimedValue, hdeu.Smoothed())
		}

		for i := lprimed; i < len(input); i++ {
			hdeu.Update(input[i])
			check(i, exp[i], hdeu.Smoothed())
		}

		previous := hdeu.Smoothed()
		hdeu.Update(math.NaN())
		check(index, previous, hdeu.Smoothed())
	})

	t.Run("reference implementation: detrended (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		hdeu := testHomodyneDiscriminatorEstimatorUnrolledCreateDefault()
		exp := testHomodyneDiscriminatorEstimatorUnrolledExpectedDetrended()

		const (
			lprimed        = 3
			notPrimedValue = 0
			index          = 99999
		)

		for i := range lprimed {
			hdeu.Update(input[i])
			check(i, notPrimedValue, hdeu.Detrended())
		}

		for i := lprimed; i < len(input); i++ {
			hdeu.Update(input[i])
			check(i, exp[i], hdeu.Detrended())
		}

		previous := hdeu.Detrended()
		hdeu.Update(math.NaN())
		check(index, previous, hdeu.Detrended())
	})

	t.Run("reference implementation: quadrature (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		hdeu := testHomodyneDiscriminatorEstimatorUnrolledCreateDefault()
		exp := testHomodyneDiscriminatorEstimatorUnrolledExpectedQuadrature()

		const (
			lprimed        = 3
			notPrimedValue = 0
			index          = 99999
		)

		for i := range lprimed {
			hdeu.Update(input[i])
			check(i, notPrimedValue, hdeu.Quadrature())
		}

		for i := lprimed; i < len(input); i++ {
			hdeu.Update(input[i])
			check(i, exp[i], hdeu.Quadrature())
		}

		previous := hdeu.Quadrature()
		hdeu.Update(math.NaN())
		check(index, previous, hdeu.Quadrature())
	})

	t.Run("reference implementation: in-phase (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		hdeu := testHomodyneDiscriminatorEstimatorUnrolledCreateDefault()
		exp := testHomodyneDiscriminatorEstimatorUnrolledExpectedInPhase()

		const (
			lprimed        = 3
			notPrimedValue = 0
			index          = 99999
		)

		for i := range lprimed {
			hdeu.Update(input[i])
			check(i, notPrimedValue, hdeu.InPhase())
		}

		for i := lprimed; i < len(input); i++ {
			hdeu.Update(input[i])
			check(i, exp[i], hdeu.InPhase())
		}

		previous := hdeu.InPhase()
		hdeu.Update(math.NaN())
		check(index, previous, hdeu.InPhase())
	})

	t.Run("reference implementation: period (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		hdeu := testHomodyneDiscriminatorEstimatorUnrolledCreateDefault()
		exp := testHomodyneDiscriminatorEstimatorUnrolledExpectedPeriod()

		const (
			lprimed        = 3
			notPrimedValue = 6
			index          = 99999
		)

		for i := range lprimed {
			hdeu.Update(input[i])
			check(i, notPrimedValue, hdeu.Period())
		}

		for i := lprimed; i < len(input); i++ {
			hdeu.Update(input[i])
			check(i, exp[i], hdeu.Period())
		}

		previous := hdeu.Period()
		hdeu.Update(math.NaN())
		check(index, previous, hdeu.Period())
	})
}

//nolint:dupl
func TestHomodyneDiscriminatorEstimatorUnrolledPeriod(t *testing.T) {
	t.Parallel()

	check := func(exp, act, epsilon float64) {
		t.Helper()

		if math.Abs(exp-act) > epsilon {
			t.Errorf("period is incorrect: expected %v, actual %v", exp, act)
		}
	}

	update := func(omega float64) *HomodyneDiscriminatorEstimatorUnrolled {
		t.Helper()

		const updates = 512

		hdeu := testHomodyneDiscriminatorEstimatorUnrolledCreateDefault()
		for i := range updates {
			hdeu.Update(math.Sin(omega * float64(i)))
		}

		return hdeu
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

		hdeu := update(omega)
		check(float64(hdeu.MinPeriod()), float64(hdeu.Period()), epsilon)
	})

	t.Run("max period of sin input", func(t *testing.T) {
		t.Parallel()

		const (
			period  = 60
			omega   = 2 * math.Pi / period
			epsilon = 1e-14
		)

		hdeu := update(omega)
		check(float64(hdeu.MaxPeriod()), float64(hdeu.Period()), epsilon)
	})
}

func TestHomodyneDiscriminatorEstimatorUnrolledPrimed(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	input := testHomodyneDiscriminatorEstimatorUnrolledInput()

	t.Run("reference implementation: primed (test_ht_hd.xsl)", func(t *testing.T) {
		t.Parallel()

		hdeu := testHomodyneDiscriminatorEstimatorUnrolledCreateDefault()

		const lprimed = 2 + 7*3

		check(0, false, hdeu.Primed())

		for i := range lprimed {
			hdeu.Update(input[i])
			check(i+1, false, hdeu.Primed())
		}

		for i := lprimed; i < len(input); i++ {
			hdeu.Update(input[i])
			check(i+1, true, hdeu.Primed())
		}
	})

	t.Run("reference implementation: primed with warmup", func(t *testing.T) {
		t.Parallel()

		const lprimed = 50

		hdeu := testHomodyneDiscriminatorEstimatorUnrolledCreateWarmUp(lprimed)

		check(0, false, hdeu.Primed())

		for i := range lprimed {
			hdeu.Update(input[i])
			check(i+1, false, hdeu.Primed())
		}

		for i := lprimed; i < len(input); i++ {
			hdeu.Update(input[i])
			check(i+1, true, hdeu.Primed())
		}
	})
}

func TestNewHomodyneDiscriminatorEstimatorUnrolled(t *testing.T) { //nolint: funlen, maintidx
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

	checkInstance := func(hdeu *HomodyneDiscriminatorEstimatorUnrolled,
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

		const primedCount = 23

		warmUpPeriod := max(warmUp, primedCount)

		check("smoothingLength", length, hdeu.smoothingLength)
		check("SmoothingLength()", length, hdeu.SmoothingLength())
		check("minPeriod", defaultMinPeriod, hdeu.minPeriod)
		check("MinPeriod()", defaultMinPeriod, hdeu.MinPeriod())
		check("maxPeriod", defaultMaxPeriod, hdeu.maxPeriod)
		check("MaxPeriod()", defaultMaxPeriod, hdeu.MaxPeriod())
		check("warmUpPeriod", warmUpPeriod, hdeu.warmUpPeriod)
		check("WarmUpPeriod()", warmUpPeriod, hdeu.WarmUpPeriod())
		check("alphaEmaQuadratureInPhase", alphaQuadratureInPhase, hdeu.alphaEmaQuadratureInPhase)
		check("AlphaEmaQuadratureInPhase()", alphaQuadratureInPhase, hdeu.AlphaEmaQuadratureInPhase())
		check("oneMinAlphaEmaQuadratureInPhase", 1-alphaQuadratureInPhase, hdeu.oneMinAlphaEmaQuadratureInPhase)
		check("alphaEmaPeriod", alphaPeriod, hdeu.alphaEmaPeriod)
		check("AlphaEmaPeriod()", alphaPeriod, hdeu.AlphaEmaPeriod())
		check("oneMinAlphaEmaPeriod", 1-alphaPeriod, hdeu.oneMinAlphaEmaPeriod)
		check("isPrimed", false, hdeu.isPrimed)
		check("isWarmedUp", false, hdeu.isWarmedUp)
		check("period", float64(defaultMinPeriod), hdeu.period)
		check("Period()", float64(defaultMinPeriod), hdeu.Period())
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("err == nil", true, err == nil)
		checkInstance(hdeu, params.SmoothingLength, params.AlphaEmaQuadratureInPhase,
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("err == nil", true, err == nil)
		checkInstance(hdeu, params.SmoothingLength, params.AlphaEmaQuadratureInPhase,
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
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

		hdeu, err := NewHomodyneDiscriminatorEstimatorUnrolled(&params)
		check("hdeu == nil", true, hdeu == nil)
		check("err", errap, err.Error())
	})
}

func testHomodyneDiscriminatorEstimatorUnrolledCreateDefault() *HomodyneDiscriminatorEstimatorUnrolled {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.2,
		AlphaEmaPeriod:            0.2,
	}

	hdeu, _ := NewHomodyneDiscriminatorEstimatorUnrolled(&params)

	return hdeu
}

func testHomodyneDiscriminatorEstimatorUnrolledCreateWarmUp(warmUp int) *HomodyneDiscriminatorEstimatorUnrolled {
	params := CycleEstimatorParams{
		SmoothingLength:           4,
		AlphaEmaQuadratureInPhase: 0.2,
		AlphaEmaPeriod:            0.2,
		WarmUpPeriod:              warmUp,
	}

	hdeu, _ := NewHomodyneDiscriminatorEstimatorUnrolled(&params)

	return hdeu
}
