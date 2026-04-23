package autocorrelationperiodogram

import "math"

// estimator implements the Ehlers autocorrelation periodogram of EasyLanguage
// listing 8-3. It is an unexported implementation detail used only by the
// AutoCorrelationPeriodogram.
//
// Pipeline per input sample:
//  1. 2-pole Butterworth highpass filter tuned to MaxPeriod.
//  2. 2-pole Super Smoother filter tuned to MinPeriod.
//  3. For each Lag in [0..MaxPeriod], compute Pearson correlation Corr[Lag]
//     between Filt[0..M-1] and Filt[Lag..Lag+M-1] with M = AveragingLength.
//  4. For each Period in [MinPeriod..MaxPeriod], compute the DFT coefficients
//     CosinePart[P] = Σ_{N=3..MaxPeriod} Corr[N]·cos(2πN/P) and SinePart[P]
//     analogously; SqSum[P] = Cos² + Sin².
//  5. Smooth: R[P] = 0.2·SqSum[P]^(2 or 1) + 0.8·R_previous[P].
//  6. Normalize by fast-attack / slow-decay AGC against MaxPwr; the output
//     spectrum holds Pwr[P] = R[P] / MaxPwr.
type estimator struct {
	minPeriod       int
	maxPeriod       int
	averagingLength int
	lengthSpectrum  int
	filtBufferLen   int

	isSpectralSquaring              bool
	isSmoothing                     bool
	isAutomaticGainControl          bool
	automaticGainControlDecayFactor float64

	// Pre-filter coefficients (scalar).
	coeffHP0 float64
	coeffHP1 float64
	coeffHP2 float64
	ssC1     float64
	ssC2     float64
	ssC3     float64

	// DFT basis tables. cosTab[p-minPeriod][n] = cos(2πn/p), sinTab similar,
	// for n in [0..maxPeriod], p in [minPeriod..maxPeriod]. The DFT only
	// sums over n in [dftLagStart..maxPeriod], with dftLagStart = 3 per EL.
	cosTab [][]float64
	sinTab [][]float64

	// Pre-filter state.
	close0, close1, close2 float64
	hp0, hp1, hp2          float64

	// Filt history: filt[k] = Filt k bars ago (0 = current). Length =
	// maxPeriod + averagingLength.
	filt []float64

	// Per-lag Pearson correlation coefficients, indexed by lag [0..maxPeriod].
	corr []float64

	// Smoothed power per period bin [0..lengthSpectrum), bin i -> period minPeriod+i.
	rPrevious []float64

	// Normalized spectrum values (output), indexed [0..lengthSpectrum).
	spectrum []float64

	spectrumMin         float64
	spectrumMax         float64
	previousSpectrumMax float64
}

// newEstimator creates a new autocorrelation periodogram estimator.
func newEstimator(
	minPeriod, maxPeriod, averagingLength int,
	isSpectralSquaring bool,
	isSmoothing bool,
	isAutomaticGainControl bool,
	automaticGainControlDecayFactor float64,
) *estimator {
	const (
		twoPi       = 2 * math.Pi
		dftLagStart = 3 // EL hardcodes the DFT inner sum to start at N=3.
	)

	lengthSpectrum := maxPeriod - minPeriod + 1
	filtBufferLen := maxPeriod + averagingLength
	corrLen := maxPeriod + 1

	// Highpass coefficients, cutoff at MaxPeriod.
	omegaHP := 0.707 * twoPi / float64(maxPeriod)
	alphaHP := (math.Cos(omegaHP) + math.Sin(omegaHP) - 1) / math.Cos(omegaHP)
	cHP0 := (1 - alphaHP/2) * (1 - alphaHP/2)
	cHP1 := 2 * (1 - alphaHP)
	cHP2 := (1 - alphaHP) * (1 - alphaHP)

	// SuperSmoother coefficients, period = MinPeriod.
	a1 := math.Exp(-1.414 * math.Pi / float64(minPeriod))
	b1 := 2 * a1 * math.Cos(1.414*math.Pi/float64(minPeriod))
	ssC2 := b1
	ssC3 := -a1 * a1
	ssC1 := 1 - ssC2 - ssC3

	// DFT basis tables; cosTab[i][n] for period = minPeriod+i and lag n in [0..maxPeriod].
	cosTab := make([][]float64, lengthSpectrum)
	sinTab := make([][]float64, lengthSpectrum)

	for i := 0; i < lengthSpectrum; i++ {
		period := minPeriod + i

		cosTab[i] = make([]float64, corrLen)
		sinTab[i] = make([]float64, corrLen)

		for n := dftLagStart; n < corrLen; n++ {
			angle := twoPi * float64(n) / float64(period)
			cosTab[i][n] = math.Cos(angle)
			sinTab[i][n] = math.Sin(angle)
		}
	}

	return &estimator{
		minPeriod:                       minPeriod,
		maxPeriod:                       maxPeriod,
		averagingLength:                 averagingLength,
		lengthSpectrum:                  lengthSpectrum,
		filtBufferLen:                   filtBufferLen,
		isSpectralSquaring:              isSpectralSquaring,
		isSmoothing:                     isSmoothing,
		isAutomaticGainControl:          isAutomaticGainControl,
		automaticGainControlDecayFactor: automaticGainControlDecayFactor,
		coeffHP0:                        cHP0,
		coeffHP1:                        cHP1,
		coeffHP2:                        cHP2,
		ssC1:                            ssC1,
		ssC2:                            ssC2,
		ssC3:                            ssC3,
		cosTab:                          cosTab,
		sinTab:                          sinTab,
		filt:                            make([]float64, filtBufferLen),
		corr:                            make([]float64, corrLen),
		rPrevious:                       make([]float64, lengthSpectrum),
		spectrum:                        make([]float64, lengthSpectrum),
	}
}

// update advances the estimator by one input sample and evaluates the spectrum.
// Callers are responsible for gating on priming.
//
//nolint:funlen,cyclop,gocognit
func (e *estimator) update(sample float64) {
	const dftLagStart = 3

	// Pre-filter cascade.
	e.close2 = e.close1
	e.close1 = e.close0
	e.close0 = sample

	e.hp2 = e.hp1
	e.hp1 = e.hp0
	e.hp0 = e.coeffHP0*(e.close0-2*e.close1+e.close2) +
		e.coeffHP1*e.hp1 -
		e.coeffHP2*e.hp2

	// Shift Filt history rightward; new Filt at filt[0].
	for k := e.filtBufferLen - 1; k >= 1; k-- {
		e.filt[k] = e.filt[k-1]
	}

	e.filt[0] = e.ssC1*(e.hp0+e.hp1)/2 + e.ssC2*e.filt[1] + e.ssC3*e.filt[2]

	// Pearson correlation per lag [0..maxPeriod], fixed M = averagingLength.
	m := e.averagingLength

	for lag := 0; lag <= e.maxPeriod; lag++ {
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

		e.corr[lag] = r
	}

	// Discrete Fourier transform of the correlation function, per period bin.
	// Then smooth (EL: squared SqSum), then AGC-normalize.
	e.spectrumMin = math.MaxFloat64
	if e.isAutomaticGainControl {
		e.spectrumMax = e.automaticGainControlDecayFactor * e.previousSpectrumMax
	} else {
		e.spectrumMax = -math.MaxFloat64
	}

	// Pass 1: compute raw R values and track the running max for AGC.
	for i := 0; i < e.lengthSpectrum; i++ {
		cosRow := e.cosTab[i]
		sinRow := e.sinTab[i]

		var cosPart, sinPart float64

		for n := dftLagStart; n <= e.maxPeriod; n++ {
			cosPart += e.corr[n] * cosRow[n]
			sinPart += e.corr[n] * sinRow[n]
		}

		sqSum := cosPart*cosPart + sinPart*sinPart

		raw := sqSum
		if e.isSpectralSquaring {
			raw = sqSum * sqSum
		}

		var r float64
		if e.isSmoothing {
			r = 0.2*raw + 0.8*e.rPrevious[i]
		} else {
			r = raw
		}

		e.rPrevious[i] = r
		e.spectrum[i] = r

		if e.spectrumMax < r {
			e.spectrumMax = r
		}
	}

	e.previousSpectrumMax = e.spectrumMax

	// Pass 2: normalize against the running max and track the (normalized) min.
	if e.spectrumMax > 0 {
		for i := 0; i < e.lengthSpectrum; i++ {
			v := e.spectrum[i] / e.spectrumMax
			e.spectrum[i] = v

			if e.spectrumMin > v {
				e.spectrumMin = v
			}
		}
	} else {
		for i := 0; i < e.lengthSpectrum; i++ {
			e.spectrum[i] = 0
		}

		e.spectrumMin = 0
	}
}
