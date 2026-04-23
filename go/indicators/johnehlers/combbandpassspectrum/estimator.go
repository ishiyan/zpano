package combbandpassspectrum

import "math"

// estimator implements the Ehlers comb band-pass spectrum estimator of
// EasyLanguage listing 10-1. It is an unexported implementation detail used
// only by the CombBandPassSpectrum indicator.
//
// Pipeline per input sample:
//  1. 2-pole Butterworth highpass filter tuned to MaxPeriod.
//  2. 2-pole Super Smoother filter tuned to MinPeriod.
//  3. Bank of 2-pole band-pass filters, one per integer period N in
//     [MinPeriod..MaxPeriod]. Power at bin N is the sum over the last N
//     band-pass outputs of (BP[N,m]/Comp)^2, with Comp = N when spectral
//     dilation compensation is on (default) or 1 otherwise.
type estimator struct {
	minPeriod                       int
	maxPeriod                       int
	lengthSpectrum                  int
	isSpectralDilationCompensation  bool
	isAutomaticGainControl          bool
	automaticGainControlDecayFactor float64

	// Pre-filter coefficients (scalar).
	alphaHP  float64 // α₁ for Butterworth highpass
	coeffHP0 float64 // (1 - α/2)^2
	coeffHP1 float64 // 2*(1 - α)
	coeffHP2 float64 // (1 - α)^2
	ssC1     float64 // SuperSmoother c1
	ssC2     float64 // SuperSmoother c2
	ssC3     float64 // SuperSmoother c3

	// Per-bin band-pass coefficients, indexed [0..lengthSpectrum).
	// Bin i corresponds to period N = minPeriod + i.
	periods []int
	beta    []float64 // β₁ = cos(2π/N)
	alpha   []float64 // α₁ = γ₁ - √(γ₁²−1), γ₁ = 1/cos(2π·bw/N)
	comp    []float64 // N when SDC is on, 1 otherwise

	// Pre-filter state (time-indexed: 0 current, 1 one bar ago, 2 two bars ago).
	close0, close1, close2 float64
	hp0, hp1, hp2          float64
	filt0, filt1, filt2    float64

	// Band-pass filter state. bp[i][m] holds band-pass output for bin i at
	// lag m (m=0 current, m=1 one bar ago, ..., m=maxPeriod-1 oldest tracked).
	bp [][]float64

	// Raw (unnormalized) spectrum values, indexed [0..lengthSpectrum), with
	// bin i corresponding to period minPeriod + i.
	spectrum []float64

	spectrumMin         float64
	spectrumMax         float64
	previousSpectrumMax float64
}

// newEstimator creates a new comb band-pass spectrum estimator.
func newEstimator(
	minPeriod, maxPeriod int,
	bandwidth float64,
	isSpectralDilationCompensation bool,
	isAutomaticGainControl bool,
	automaticGainControlDecayFactor float64,
) *estimator {
	const twoPi = 2 * math.Pi

	lengthSpectrum := maxPeriod - minPeriod + 1

	// Highpass coefficients, cutoff at MaxPeriod. EL uses degrees; we convert
	// directly to radians:  .707*360/MaxPeriod deg = .707*2π/MaxPeriod rad.
	omegaHP := 0.707 * twoPi / float64(maxPeriod)
	alphaHP := (math.Cos(omegaHP) + math.Sin(omegaHP) - 1) / math.Cos(omegaHP)
	cHP0 := (1 - alphaHP/2) * (1 - alphaHP/2)
	cHP1 := 2 * (1 - alphaHP)
	cHP2 := (1 - alphaHP) * (1 - alphaHP)

	// SuperSmoother coefficients, period = MinPeriod. EL: a1 = exp(-1.414π/MinPeriod),
	// b1 = 2*a1*cos(1.414·180/MinPeriod deg) = 2*a1*cos(1.414π/MinPeriod rad).
	a1 := math.Exp(-1.414 * math.Pi / float64(minPeriod))
	b1 := 2 * a1 * math.Cos(1.414*math.Pi/float64(minPeriod))
	ssC2 := b1
	ssC3 := -a1 * a1
	ssC1 := 1 - ssC2 - ssC3

	e := &estimator{
		minPeriod:                       minPeriod,
		maxPeriod:                       maxPeriod,
		lengthSpectrum:                  lengthSpectrum,
		isSpectralDilationCompensation:  isSpectralDilationCompensation,
		isAutomaticGainControl:          isAutomaticGainControl,
		automaticGainControlDecayFactor: automaticGainControlDecayFactor,
		alphaHP:                         alphaHP,
		coeffHP0:                        cHP0,
		coeffHP1:                        cHP1,
		coeffHP2:                        cHP2,
		ssC1:                            ssC1,
		ssC2:                            ssC2,
		ssC3:                            ssC3,
		periods:                         make([]int, lengthSpectrum),
		beta:                            make([]float64, lengthSpectrum),
		alpha:                           make([]float64, lengthSpectrum),
		comp:                            make([]float64, lengthSpectrum),
		bp:                              make([][]float64, lengthSpectrum),
		spectrum:                        make([]float64, lengthSpectrum),
	}

	for i := 0; i < lengthSpectrum; i++ {
		n := minPeriod + i
		beta := math.Cos(twoPi / float64(n))
		gamma := 1 / math.Cos(twoPi*bandwidth/float64(n))
		alpha := gamma - math.Sqrt(gamma*gamma-1)

		e.periods[i] = n
		e.beta[i] = beta
		e.alpha[i] = alpha

		if isSpectralDilationCompensation {
			e.comp[i] = float64(n)
		} else {
			e.comp[i] = 1
		}

		e.bp[i] = make([]float64, maxPeriod)
	}

	return e
}

// update advances the estimator by one input sample and evaluates the spectrum.
// Callers are responsible for gating on priming; update is safe to call from
// the first bar (the BP history just carries zeros until the pre-filters settle).
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

	// Shift Filt history and compute new Filt (SuperSmoother on HP).
	e.filt2 = e.filt1
	e.filt1 = e.filt0
	e.filt0 = e.ssC1*(e.hp0+e.hp1)/2 + e.ssC2*e.filt1 + e.ssC3*e.filt2

	// Band-pass filter bank: shift each bin's BP history rightward, write new
	// BP[i,0], then sum the last N entries (squared, comp-scaled) into Pwr[i].
	diffFilt := e.filt0 - e.filt2

	// AGC seeds the running max with the decayed previous max; floating max
	// starts at -inf.
	e.spectrumMin = math.MaxFloat64
	if e.isAutomaticGainControl {
		e.spectrumMax = e.automaticGainControlDecayFactor * e.previousSpectrumMax
	} else {
		e.spectrumMax = -math.MaxFloat64
	}

	for i := 0; i < e.lengthSpectrum; i++ {
		bpRow := e.bp[i]

		// Rightward shift: bp[i][m] = bp[i][m-1] for m from maxPeriod-1 down to 1.
		for m := e.maxPeriod - 1; m >= 1; m-- {
			bpRow[m] = bpRow[m-1]
		}

		a := e.alpha[i]
		b := e.beta[i]
		bpRow[0] = 0.5*(1-a)*diffFilt + b*(1+a)*bpRow[1] - a*bpRow[2]

		// Pwr[i] = Σ over m in [0..N) of (BP[i,m] / Comp[i])^2.
		n := e.periods[i]
		c := e.comp[i]
		pwr := 0.0

		for m := 0; m < n; m++ {
			v := bpRow[m] / c
			pwr += v * v
		}

		e.spectrum[i] = pwr

		if e.spectrumMax < pwr {
			e.spectrumMax = pwr
		}

		if e.spectrumMin > pwr {
			e.spectrumMin = pwr
		}
	}

	e.previousSpectrumMax = e.spectrumMax
}
