package autocorrelationindicator

import "math"

// estimator implements the Ehlers autocorrelation indicator of EasyLanguage
// listing 8-2. It is an unexported implementation detail used only by the
// AutoCorrelationIndicator.
//
// Pipeline per input sample:
//  1. 2-pole Butterworth highpass filter tuned to MaxLag.
//  2. 2-pole Super Smoother filter tuned to SmoothingPeriod.
//  3. For each Lag in [MinLag..MaxLag], compute Pearson correlation between
//     Filt[0..M-1] and Filt[Lag..Lag+M-1] where M = AveragingLength when > 0
//     else Lag. Output is rescaled to [0, 1] via 0.5*(r + 1).
type estimator struct {
	minLag          int
	maxLag          int
	averagingLength int
	lengthSpectrum  int
	filtBufferLen   int

	// Pre-filter coefficients (scalar).
	coeffHP0 float64 // (1 - α/2)^2
	coeffHP1 float64 // 2*(1 - α)
	coeffHP2 float64 // (1 - α)^2
	ssC1     float64
	ssC2     float64
	ssC3     float64

	// Pre-filter state.
	close0, close1, close2 float64
	hp0, hp1, hp2          float64

	// Filt history. filt[k] = Filt k bars ago (0 = current). Length is
	// maxLag + max(averagingLength, maxLag), i.e. large enough to index
	// Filt[Lag + M - 1] for Lag up to maxLag. The SuperSmoother recursion
	// reads filt[1] and filt[2] after the rightward shift.
	filt []float64

	// Spectrum values indexed [0..lengthSpectrum), where bin i corresponds
	// to lag = minLag + i. Already scaled to [0, 1] via 0.5*(r + 1).
	spectrum []float64

	spectrumMin float64
	spectrumMax float64
}

// newEstimator creates a new autocorrelation indicator estimator.
func newEstimator(minLag, maxLag, smoothingPeriod, averagingLength int) *estimator {
	const twoPi = 2 * math.Pi

	lengthSpectrum := maxLag - minLag + 1

	mMax := averagingLength
	if averagingLength == 0 {
		mMax = maxLag
	}

	filtBufferLen := maxLag + mMax

	// Highpass coefficients, cutoff at MaxLag.
	omegaHP := 0.707 * twoPi / float64(maxLag)
	alphaHP := (math.Cos(omegaHP) + math.Sin(omegaHP) - 1) / math.Cos(omegaHP)
	cHP0 := (1 - alphaHP/2) * (1 - alphaHP/2)
	cHP1 := 2 * (1 - alphaHP)
	cHP2 := (1 - alphaHP) * (1 - alphaHP)

	// SuperSmoother coefficients, period = SmoothingPeriod.
	a1 := math.Exp(-1.414 * math.Pi / float64(smoothingPeriod))
	b1 := 2 * a1 * math.Cos(1.414*math.Pi/float64(smoothingPeriod))
	ssC2 := b1
	ssC3 := -a1 * a1
	ssC1 := 1 - ssC2 - ssC3

	return &estimator{
		minLag:          minLag,
		maxLag:          maxLag,
		averagingLength: averagingLength,
		lengthSpectrum:  lengthSpectrum,
		filtBufferLen:   filtBufferLen,
		coeffHP0:        cHP0,
		coeffHP1:        cHP1,
		coeffHP2:        cHP2,
		ssC1:            ssC1,
		ssC2:            ssC2,
		ssC3:            ssC3,
		filt:            make([]float64, filtBufferLen),
		spectrum:        make([]float64, lengthSpectrum),
	}
}

// update advances the estimator by one input sample and evaluates the spectrum.
// Callers are responsible for gating on priming; update is safe to call from
// the first bar (the Filt history just carries zeros until the pre-filters settle).
//
//nolint:funlen,cyclop
func (e *estimator) update(sample float64) {
	// Shift close history.
	e.close2 = e.close1
	e.close1 = e.close0
	e.close0 = sample

	// Shift HP history and compute new HP.
	e.hp2 = e.hp1
	e.hp1 = e.hp0
	e.hp0 = e.coeffHP0*(e.close0-2*e.close1+e.close2) +
		e.coeffHP1*e.hp1 -
		e.coeffHP2*e.hp2

	// Shift Filt history rightward: filt[k] <- filt[k-1] for k from last down to 1.
	// After the shift, filt[1] is the previous Filt and filt[2] is two bars ago.
	for k := e.filtBufferLen - 1; k >= 1; k-- {
		e.filt[k] = e.filt[k-1]
	}

	// Compute new Filt (SuperSmoother on HP) and store at index 0.
	e.filt[0] = e.ssC1*(e.hp0+e.hp1)/2 + e.ssC2*e.filt[1] + e.ssC3*e.filt[2]

	// Pearson correlation per lag.
	e.spectrumMin = math.MaxFloat64
	e.spectrumMax = -math.MaxFloat64

	for i := 0; i < e.lengthSpectrum; i++ {
		lag := e.minLag + i

		m := e.averagingLength
		if m == 0 {
			m = lag
		}

		var sx, sy, sxx, syy, sxy float64

		for c := 0; c < m; c++ {
			x := e.filt[c]
			y := e.filt[lag+c]
			sx += x
			sy += y
			sxx += x * x
			syy += y * y
			sxy += x * y
		}

		denom := (float64(m)*sxx - sx*sx) * (float64(m)*syy - sy*sy)

		r := 0.0
		if denom > 0 {
			r = (float64(m)*sxy - sx*sy) / math.Sqrt(denom)
		}

		v := 0.5 * (r + 1)
		e.spectrum[i] = v

		if v < e.spectrumMin {
			e.spectrumMin = v
		}

		if v > e.spectrumMax {
			e.spectrumMax = v
		}
	}
}
