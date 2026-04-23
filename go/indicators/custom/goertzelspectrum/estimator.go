package goertzelspectrum

import "math"

// estimator implements the Goertzel spectrum estimator. It is an unexported port
// of MBST's GoertzelSpectrumEstimator used only by the GoertzelSpectrum indicator.
type estimator struct {
	length                          int
	spectrumResolution              int
	lengthSpectrum                  int
	minPeriod                       float64
	maxPeriod                       float64
	isFirstOrder                    bool
	isSpectralDilationCompensation  bool
	isAutomaticGainControl          bool
	automaticGainControlDecayFactor float64

	inputSeries          []float64
	inputSeriesMinusMean []float64
	spectrum             []float64
	period               []float64

	// Pre-computed trigonometric tables.
	frequencySin  []float64 // first-order only
	frequencyCos  []float64 // first-order only
	frequencyCos2 []float64 // second-order only

	mean                float64
	spectrumMin         float64
	spectrumMax         float64
	previousSpectrumMax float64
}

// newEstimator creates a new Goertzel spectrum estimator.
func newEstimator(
	length int,
	minPeriod, maxPeriod float64,
	spectrumResolution int,
	isFirstOrder, isSpectralDilationCompensation, isAutomaticGainControl bool,
	automaticGainControlDecayFactor float64,
) *estimator {
	const twoPi = 2 * math.Pi

	lengthSpectrum := int((maxPeriod-minPeriod)*float64(spectrumResolution)) + 1

	e := &estimator{
		length:                          length,
		spectrumResolution:              spectrumResolution,
		lengthSpectrum:                  lengthSpectrum,
		minPeriod:                       minPeriod,
		maxPeriod:                       maxPeriod,
		isFirstOrder:                    isFirstOrder,
		isSpectralDilationCompensation:  isSpectralDilationCompensation,
		isAutomaticGainControl:          isAutomaticGainControl,
		automaticGainControlDecayFactor: automaticGainControlDecayFactor,
		inputSeries:                     make([]float64, length),
		inputSeriesMinusMean:            make([]float64, length),
		spectrum:                        make([]float64, lengthSpectrum),
		period:                          make([]float64, lengthSpectrum),
	}

	result := float64(spectrumResolution)

	if isFirstOrder {
		e.frequencySin = make([]float64, lengthSpectrum)
		e.frequencyCos = make([]float64, lengthSpectrum)

		for i := 0; i < lengthSpectrum; i++ {
			period := maxPeriod - float64(i)/result
			e.period[i] = period
			theta := twoPi / period
			e.frequencySin[i] = math.Sin(theta)
			e.frequencyCos[i] = math.Cos(theta)
		}
	} else {
		e.frequencyCos2 = make([]float64, lengthSpectrum)

		for i := 0; i < lengthSpectrum; i++ {
			period := maxPeriod - float64(i)/result
			e.period[i] = period
			e.frequencyCos2[i] = 2 * math.Cos(twoPi/period)
		}
	}

	return e
}

// calculate fills mean, inputSeriesMinusMean, spectrum, spectrumMin, spectrumMax
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

	// Seed with the first bin.
	spectrum := e.goertzelEstimate(0)
	if e.isSpectralDilationCompensation {
		spectrum /= e.period[0]
	}

	e.spectrum[0] = spectrum
	e.spectrumMin = spectrum

	if e.isAutomaticGainControl {
		e.spectrumMax = e.automaticGainControlDecayFactor * e.previousSpectrumMax
		if e.spectrumMax < spectrum {
			e.spectrumMax = spectrum
		}
	} else {
		e.spectrumMax = spectrum
	}

	for i := 1; i < e.lengthSpectrum; i++ {
		spectrum = e.goertzelEstimate(i)
		if e.isSpectralDilationCompensation {
			spectrum /= e.period[i]
		}

		e.spectrum[i] = spectrum

		if e.spectrumMax < spectrum {
			e.spectrumMax = spectrum
		} else if e.spectrumMin > spectrum {
			e.spectrumMin = spectrum
		}
	}

	e.previousSpectrumMax = e.spectrumMax
}

func (e *estimator) goertzelEstimate(j int) float64 {
	if e.isFirstOrder {
		return e.goertzelFirstOrderEstimate(j)
	}

	return e.goertzelSecondOrderEstimate(j)
}

func (e *estimator) goertzelSecondOrderEstimate(j int) float64 {
	cos2 := e.frequencyCos2[j]

	var s1, s2 float64

	for i := 0; i < e.length; i++ {
		s0 := e.inputSeriesMinusMean[i] + cos2*s1 - s2
		s2 = s1
		s1 = s0
	}

	spectrum := s1*s1 + s2*s2 - cos2*s1*s2
	if spectrum < 0 {
		return 0
	}

	return spectrum
}

func (e *estimator) goertzelFirstOrderEstimate(j int) float64 {
	cosTheta := e.frequencyCos[j]
	sinTheta := e.frequencySin[j]

	var yre, yim float64

	for i := 0; i < e.length; i++ {
		re := e.inputSeriesMinusMean[i] + cosTheta*yre - sinTheta*yim
		im := e.inputSeriesMinusMean[i] + cosTheta*yim + sinTheta*yre
		yre = re
		yim = im
	}

	return yre*yre + yim*yim
}
