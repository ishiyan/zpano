package discretefouriertransformspectrum

import "math"

// estimator implements the discrete Fourier transform power spectrum estimator.
// It is an unexported port of MBST's DiscreteFourierTransformSpectrumEstimator,
// used only by the DiscreteFourierTransformSpectrum indicator.
type estimator struct {
	length                          int
	spectrumResolution              int
	lengthSpectrum                  int
	maxOmegaLength                  int
	minPeriod                       float64
	maxPeriod                       float64
	isSpectralDilationCompensation  bool
	isAutomaticGainControl          bool
	automaticGainControlDecayFactor float64

	inputSeries          []float64
	inputSeriesMinusMean []float64
	spectrum             []float64
	period               []float64

	// Pre-computed trigonometric tables, size [lengthSpectrum][maxOmegaLength].
	// maxOmegaLength equals length (full-window DFT).
	frequencySinOmega [][]float64
	frequencyCosOmega [][]float64

	mean                float64
	spectrumMin         float64
	spectrumMax         float64
	previousSpectrumMax float64
}

// newEstimator creates a new discrete Fourier transform spectrum estimator.
func newEstimator(
	length int,
	minPeriod, maxPeriod float64,
	spectrumResolution int,
	isSpectralDilationCompensation bool,
	isAutomaticGainControl bool,
	automaticGainControlDecayFactor float64,
) *estimator {
	const twoPi = 2 * math.Pi

	lengthSpectrum := int((maxPeriod-minPeriod)*float64(spectrumResolution)) + 1
	maxOmegaLength := length

	e := &estimator{
		length:                          length,
		spectrumResolution:              spectrumResolution,
		lengthSpectrum:                  lengthSpectrum,
		maxOmegaLength:                  maxOmegaLength,
		minPeriod:                       minPeriod,
		maxPeriod:                       maxPeriod,
		isSpectralDilationCompensation:  isSpectralDilationCompensation,
		isAutomaticGainControl:          isAutomaticGainControl,
		automaticGainControlDecayFactor: automaticGainControlDecayFactor,
		inputSeries:                     make([]float64, length),
		inputSeriesMinusMean:            make([]float64, length),
		spectrum:                        make([]float64, lengthSpectrum),
		period:                          make([]float64, lengthSpectrum),
		frequencySinOmega:               make([][]float64, lengthSpectrum),
		frequencyCosOmega:               make([][]float64, lengthSpectrum),
	}

	result := float64(spectrumResolution)

	// Spectrum is evaluated from MaxPeriod down to MinPeriod with the configured resolution.
	for i := 0; i < lengthSpectrum; i++ {
		p := maxPeriod - float64(i)/result
		e.period[i] = p
		theta := twoPi / p

		sinRow := make([]float64, maxOmegaLength)
		cosRow := make([]float64, maxOmegaLength)

		for j := 0; j < maxOmegaLength; j++ {
			omega := float64(j) * theta
			sinRow[j] = math.Sin(omega)
			cosRow[j] = math.Cos(omega)
		}

		e.frequencySinOmega[i] = sinRow
		e.frequencyCosOmega[i] = cosRow
	}

	return e
}

// calculate fills mean, inputSeriesMinusMean, spectrum, spectrumMin, and spectrumMax
// from the current inputSeries contents.
func (e *estimator) calculate() {
	// Subtract the mean from the input series.
	mean := 0.0
	for i := 0; i < e.length; i++ {
		mean += e.inputSeries[i]
	}

	mean /= float64(e.length)

	for i := 0; i < e.length; i++ {
		e.inputSeriesMinusMean[i] = e.inputSeries[i] - mean
	}

	e.mean = mean

	// Evaluate the DFT power spectrum.
	e.spectrumMin = math.MaxFloat64
	if e.isAutomaticGainControl {
		e.spectrumMax = e.automaticGainControlDecayFactor * e.previousSpectrumMax
	} else {
		e.spectrumMax = -math.MaxFloat64
	}

	for i := 0; i < e.lengthSpectrum; i++ {
		sinRow := e.frequencySinOmega[i]
		cosRow := e.frequencyCosOmega[i]

		var sumSin, sumCos float64

		for j := 0; j < e.maxOmegaLength; j++ {
			sample := e.inputSeriesMinusMean[j]
			sumSin += sample * sinRow[j]
			sumCos += sample * cosRow[j]
		}

		s := sumSin*sumSin + sumCos*sumCos
		if e.isSpectralDilationCompensation {
			s /= e.period[i]
		}

		e.spectrum[i] = s

		if e.spectrumMax < s {
			e.spectrumMax = s
		}

		if e.spectrumMin > s {
			e.spectrumMin = s
		}
	}

	e.previousSpectrumMax = e.spectrumMax
}
