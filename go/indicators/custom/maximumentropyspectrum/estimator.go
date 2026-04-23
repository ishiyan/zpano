package maximumentropyspectrum

import "math"

// estimator implements the maximum-entropy (Burg) spectrum estimator. It is an
// unexported port of MBST's MaximumEntropySpectrumEstimator used only by the
// MaximumEntropySpectrum indicator.
type estimator struct {
	length                          int
	degree                          int
	spectrumResolution              int
	lengthSpectrum                  int
	minPeriod                       float64
	maxPeriod                       float64
	isAutomaticGainControl          bool
	automaticGainControlDecayFactor float64

	inputSeries          []float64
	inputSeriesMinusMean []float64
	coefficients         []float64 // length = degree
	spectrum             []float64
	period               []float64

	// Pre-computed trigonometric tables, size [lengthSpectrum][degree].
	frequencySinOmega [][]float64
	frequencyCosOmega [][]float64

	// Burg working buffers.
	h   []float64 // length = degree + 1
	g   []float64 // length = degree + 2
	per []float64 // length = length + 1
	pef []float64 // length = length + 1

	mean                float64
	spectrumMin         float64
	spectrumMax         float64
	previousSpectrumMax float64
}

// newEstimator creates a new maximum-entropy spectrum estimator.
func newEstimator(
	length, degree int,
	minPeriod, maxPeriod float64,
	spectrumResolution int,
	isAutomaticGainControl bool,
	automaticGainControlDecayFactor float64,
) *estimator {
	const twoPi = 2 * math.Pi

	lengthSpectrum := int((maxPeriod-minPeriod)*float64(spectrumResolution)) + 1

	e := &estimator{
		length:                          length,
		degree:                          degree,
		spectrumResolution:              spectrumResolution,
		lengthSpectrum:                  lengthSpectrum,
		minPeriod:                       minPeriod,
		maxPeriod:                       maxPeriod,
		isAutomaticGainControl:          isAutomaticGainControl,
		automaticGainControlDecayFactor: automaticGainControlDecayFactor,
		inputSeries:                     make([]float64, length),
		inputSeriesMinusMean:            make([]float64, length),
		coefficients:                    make([]float64, degree),
		spectrum:                        make([]float64, lengthSpectrum),
		period:                          make([]float64, lengthSpectrum),
		frequencySinOmega:               make([][]float64, lengthSpectrum),
		frequencyCosOmega:               make([][]float64, lengthSpectrum),
		h:                               make([]float64, degree+1),
		g:                               make([]float64, degree+2),
		per:                             make([]float64, length+1),
		pef:                             make([]float64, length+1),
	}

	result := float64(spectrumResolution)

	// Spectrum is evaluated from MaxPeriod down to MinPeriod with the configured resolution.
	for i := 0; i < lengthSpectrum; i++ {
		p := maxPeriod - float64(i)/result
		e.period[i] = p
		theta := twoPi / p

		sinRow := make([]float64, degree)
		cosRow := make([]float64, degree)

		for j := 0; j < degree; j++ {
			omega := -float64(j+1) * theta
			sinRow[j] = math.Sin(omega)
			cosRow[j] = math.Cos(omega)
		}

		e.frequencySinOmega[i] = sinRow
		e.frequencyCosOmega[i] = cosRow
	}

	return e
}

// calculate fills mean, inputSeriesMinusMean, coefficients, spectrum, spectrumMin,
// and spectrumMax from the current inputSeries contents.
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

	e.burgEstimate(e.inputSeriesMinusMean)

	// Evaluate the spectrum from the AR coefficients.
	e.spectrumMin = math.MaxFloat64
	if e.isAutomaticGainControl {
		e.spectrumMax = e.automaticGainControlDecayFactor * e.previousSpectrumMax
	} else {
		e.spectrumMax = -math.MaxFloat64
	}

	for i := 0; i < e.lengthSpectrum; i++ {
		real := 1.0
		imag := 0.0

		cosRow := e.frequencyCosOmega[i]
		sinRow := e.frequencySinOmega[i]

		for j := 0; j < e.degree; j++ {
			real -= e.coefficients[j] * cosRow[j]
			imag -= e.coefficients[j] * sinRow[j]
		}

		s := 1.0 / (real*real + imag*imag)
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

// burgEstimate estimates auto-regression coefficients of the configured degree using
// the Burg maximum-entropy method. It is a direct port of the zero-based C reference
// from Paul Bourke's ar.h suite, matching MBST's implementation.
//
//nolint:cyclop
func (e *estimator) burgEstimate(series []float64) {
	for i := 1; i <= e.length; i++ {
		e.pef[i] = 0
		e.per[i] = 0
	}

	for i := 1; i <= e.degree; i++ {
		var sn, sd float64 // numerator, denominator

		jj := e.length - i

		for j := 0; j < jj; j++ {
			t1 := series[j+i] + e.pef[j]
			t2 := series[j] + e.per[j]
			sn -= 2.0 * t1 * t2
			sd += t1*t1 + t2*t2
		}

		t := sn / sd
		e.g[i] = t

		if i != 1 {
			for j := 1; j < i; j++ {
				e.h[j] = e.g[j] + t*e.g[i-j]
			}

			for j := 1; j < i; j++ {
				e.g[j] = e.h[j]
			}

			jj--
		}

		for j := 0; j < jj; j++ {
			e.per[j] += t*e.pef[j] + t*series[j+i]
			e.pef[j] = e.pef[j+1] + t*e.per[j+1] + t*series[j+1]
		}
	}

	for i := 0; i < e.degree; i++ {
		e.coefficients[i] = -e.g[i+1]
	}
}
