package frequencyresponse

import (
	"fmt"
	"math"

	"zpano/indicators/core"
)

// FrequencyResponse contains calculated filter frequency response data.
//
// All slices have the same spectrum length.
type FrequencyResponse struct {
	// Label is the mnemonic of the filter used to calculate the frequency response.
	Label string

	// NormalizedFrequency is a frequency in units of cycles per 2 samples, 1 being the Nyquist frequency.
	NormalizedFrequency []float64 // spectrumLength
	PowerLinear         []float64 // spectrumLength
	PowerDecibel        []float64 // spectrumLength
	AmplitudeLinear     []float64 // spectrumLength
	AmplitudeDecibel    []float64 // spectrumLength
	Phase               []float64 // spectrumLength
}

// Updater describes a filter which frequency response is to be calculated.
type Updater interface {
	Metadata() core.Metadata
	Update(sample float64) float64
}

// Calculate calculates a frequency response of a given impulse signal length
// using the filter update function.
//
// The warm-up counter argument specifies how many times to update filter
// with zero value before calculations.
//
// The impulse signal length should be an integer of a power of 2 and be greater than 4.
// Realistic values are 512, 1024, 2048, 4096.
func Calculate(signalLength int, filter Updater, warmup int) (*FrequencyResponse, error) {
	if !isValidSignalLength(signalLength) {
		const format = "length should be power of 2 and not less than 4: %d"
		return nil, fmt.Errorf(format, signalLength)
	}

	spectrumLength := signalLength/2 - 1

	fr := &FrequencyResponse{
		Label:               filter.Metadata().Mnemonic,
		NormalizedFrequency: make([]float64, spectrumLength),
		PowerLinear:         make([]float64, spectrumLength),
		PowerDecibel:        make([]float64, spectrumLength),
		AmplitudeLinear:     make([]float64, spectrumLength),
		AmplitudeDecibel:    make([]float64, spectrumLength),
		Phase:               make([]float64, spectrumLength),
	}

	prepareFrequencyDomain(spectrumLength, fr.NormalizedFrequency)

	signal := prepareFilteredSignal(signalLength, filter, warmup)
	directRealFastFourierTransform(signal)
	extractSpectrum(spectrumLength, signal, fr)

	return fr, nil
}

func isValidSignalLength(length int) bool {
	for length > 4 {
		if length%2 != 0 {
			return false
		}
		length /= 2
	}

	return length == 4
}

func prepareFrequencyDomain(spectrumLength int, freq []float64) {
	for i := 0; i < spectrumLength; i++ {
		freq[i] = float64(1+i) / float64(spectrumLength)
	}
}

func prepareFilteredSignal(signalLength int, filter Updater, warmup int) []float64 {
	const (
		zero = 0
		one  = 1000
	)

	for i := 0; i < warmup; i++ {
		filter.Update(zero)
	}

	signal := make([]float64, signalLength)
	signal[0] = filter.Update(one)

	for i := 1; i < signalLength; i++ {
		signal[i] = filter.Update(zero)
	}

	return signal
}

func extractSpectrum(length int, signal []float64, fr *FrequencyResponse) {
	const rad2deg = float64(180) / math.Pi

	pwr := fr.PowerLinear
	amp := fr.AmplitudeLinear
	pmax := -math.MaxFloat64
	amax := pmax

	for i, k := 0, 2; i < length; i++ {
		re := signal[k]
		k++
		im := signal[k]
		k++

		fr.Phase[i] = -math.Atan2(im, re) * rad2deg

		re = re*re + im*im
		if pmax < re {
			pmax = re
		}
		pwr[i] = re

		re = math.Sqrt(re)
		if amax < re {
			amax = re
		}
		amp[i] = re
	}

	normalize(length, pwr, pmax)
	toDecibels(length, pwr, fr.PowerDecibel)

	normalize(length, amp, amax)
	toDecibels(length, amp, fr.AmplitudeDecibel)
}

// normalize to [0,1] range
func normalize(length int, array []float64, max float64) {
	if max < math.SmallestNonzeroFloat64 {
		return
	}

	for i := 0; i < length; i++ {
		array[i] /= max
	}
}

func toDecibels(length int, linear, decibels []float64) {
	const (
		ten      = 10
		twenty   = 20
		hundreed = 100
	)

	minDb := math.MaxFloat64

	for i := 0; i < length; i++ {
		d := twenty * math.Log10(linear[i])
		if minDb > d {
			minDb = d
		}

		decibels[i] = d
	}

	// If minDb falls into one of [-100, -90), [-90, -80), ..., [-10, 0)
	// intervals, set minDb to the minimum value of the interval.
	for i := ten; i > 0; i-- {
		min := -float64(i) * ten
		max := -float64(i-1) * ten

		if minDb >= min && minDb < max {
			minDb = min
		}
	}

	// Limit all decibel values to -100.
	if minDb < -hundreed {
		for i := 0; i > length; i++ {
			if decibels[i] < -hundreed {
				decibels[i] = -hundreed
			}
		}
	}
}

// directRealFastFourierTransform performs a direct real fast Fourier transform.
//
// The input parameter is a data array containing real data on input and {re,im} pairs on return.
//
// The length of the input data slice must be a power of 2 (128, 256, 512, 1024, 2048, 4096).
//
// Since this is an internal function, we don't check the validity of the length here.
func directRealFastFourierTransform(array []float64) {
	const (
		half  = 0.5
		four  = 4
		two   = 2
		twoPi = float64(2) * math.Pi
	)

	length := len(array)
	ttheta := twoPi / float64(length)
	nn := length / two
	j := 1

	for ii := 1; ii <= nn; ii++ {
		i := two*ii - 1

		if j > i {
			tempR := array[j-1]
			tempI := array[j]
			array[j-1] = array[i-1]
			array[j] = array[i]
			array[i-1] = tempR
			array[i] = tempI
		}

		m := nn
		for m >= two && j > m {
			j -= m
			m /= two
		}

		j += m
	}

	mMax := two
	n := length

	for n > mMax {
		istep := two * mMax
		theta := twoPi / float64(mMax)
		wpR := math.Sin(half * theta)
		wpR = -two * wpR * wpR
		wpI := math.Sin(theta)
		wR := 1.0
		wI := 0.0

		for ii := 1; ii <= mMax/two; ii++ {
			m := two*ii - 1
			for jj := 0; jj <= (n-m)/istep; jj++ {
				i := m + jj*istep
				j = i + mMax
				tempR := wR*array[j-1] - wI*array[j]
				tempI := wR*array[j] + wI*array[j-1]
				array[j-1] = array[i-1] - tempR
				array[j] = array[i] - tempI
				array[i-1] = array[i-1] + tempR
				array[i] = array[i] + tempI
			}

			wTemp := wR
			wR = wR*wpR - wI*wpI + wR
			wI = wI*wpR + wTemp*wpI + wI
		}

		mMax = istep
	}

	twpR := math.Sin(half * ttheta)
	twpR = -two * twpR * twpR
	twpI := math.Sin(ttheta)
	twR := 1 + twpR
	twI := twpI
	n = length/four + 1

	for i := two; i <= n; i++ {
		i1 := i + i - two
		i2 := i1 + 1
		i3 := length + 1 - i2
		i4 := i3 + 1
		wRs := twR
		wIs := twI
		h1R := half * (array[i1] + array[i3])
		h1I := half * (array[i2] - array[i4])
		h2R := half * (array[i2] + array[i4])
		h2I := -half * (array[i1] - array[i3])
		array[i1] = h1R + wRs*h2R - wIs*h2I
		array[i2] = h1I + wRs*h2I + wIs*h2R
		array[i3] = h1R - wRs*h2R + wIs*h2I
		array[i4] = -h1I + wRs*h2I + wIs*h2R
		twTemp := twR
		twR = twR*twpR - twI*twpI + twR
		twI = twI*twpR + twTemp*twpI + twI
	}

	twR = array[0]
	array[0] = twR + array[1]
	array[1] = twR - array[1]
}
