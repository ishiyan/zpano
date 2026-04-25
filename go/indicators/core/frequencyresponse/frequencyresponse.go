package frequencyresponse

import (
	"fmt"
	"math"

	"zpano/indicators/core"
)

// Component contains a single calculated filter frequency response component data.
type Component struct {
	Data []float64
	Min  float64
	Max  float64
}

func newComponent(length int) Component {
	return Component{
		Data: make([]float64, length),
		Min:  math.Inf(-1),
		Max:  math.Inf(1),
	}
}

// FrequencyResponse contains calculated filter frequency response data.
//
// All slices have the same spectrum length.
type FrequencyResponse struct {
	// Label is the mnemonic of the filter used to calculate the frequency response.
	Label string

	// NormalizedFrequency is a frequency in units of cycles per 2 samples, 1 being the Nyquist frequency.
	NormalizedFrequency []float64

	// PowerPercent is spectrum power in percentages from a maximum value.
	PowerPercent Component

	// PowerDecibel is spectrum power in decibels.
	PowerDecibel Component

	// AmplitudePercent is spectrum amplitude in percentages from a maximum value.
	AmplitudePercent Component

	// AmplitudeDecibel is spectrum amplitude in decibels.
	AmplitudeDecibel Component

	// PhaseDegrees is phase in degrees in range [-180, 180].
	PhaseDegrees Component

	// PhaseDegreesUnwrapped is phase in degrees unwrapped.
	PhaseDegreesUnwrapped Component
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
// The phaseDegreesUnwrappingLimit controls the phase unwrapping threshold (use 179 as default).
//
// The impulse signal length should be an integer of a power of 2 and be greater than 4.
// Realistic values are 512, 1024, 2048, 4096.
func Calculate(signalLength int, filter Updater, warmup int, phaseDegreesUnwrappingLimit float64) (*FrequencyResponse, error) {
	if !isValidSignalLength(signalLength) {
		const format = "length should be power of 2 and not less than 4: %d"
		return nil, fmt.Errorf(format, signalLength)
	}

	spectrumLength := signalLength/2 - 1

	fr := &FrequencyResponse{
		Label:                 filter.Metadata().Mnemonic,
		NormalizedFrequency:   make([]float64, spectrumLength),
		PowerPercent:          newComponent(spectrumLength),
		PowerDecibel:          newComponent(spectrumLength),
		AmplitudePercent:      newComponent(spectrumLength),
		AmplitudeDecibel:      newComponent(spectrumLength),
		PhaseDegrees:          newComponent(spectrumLength),
		PhaseDegreesUnwrapped: newComponent(spectrumLength),
	}

	prepareFrequencyDomain(spectrumLength, fr.NormalizedFrequency)

	signal := prepareFilteredSignal(signalLength, filter, warmup)
	directRealFastFourierTransform(signal)
	parseSpectrum(spectrumLength, signal, &fr.PowerPercent, &fr.AmplitudePercent,
		&fr.PhaseDegrees, &fr.PhaseDegreesUnwrapped, phaseDegreesUnwrappingLimit)
	toDecibels(spectrumLength, &fr.PowerPercent, &fr.PowerDecibel)
	toPercents(spectrumLength, &fr.PowerPercent, &fr.PowerPercent)
	toDecibels(spectrumLength, &fr.AmplitudePercent, &fr.AmplitudeDecibel)
	toPercents(spectrumLength, &fr.AmplitudePercent, &fr.AmplitudePercent)

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

func parseSpectrum(length int, signal []float64, power, amplitude, phase, phaseUnwrapped *Component, phaseDegreesUnwrappingLimit float64) {
	const rad2deg = float64(180) / math.Pi

	pmin := math.Inf(1)
	pmax := math.Inf(-1)
	amin := math.Inf(1)
	amax := math.Inf(-1)

	for i, k := 0, 2; i < length; i++ {
		re := signal[k]
		k++
		im := signal[k]
		k++

		// Wrapped phase -- atan2 returns radians in the [-π, π] range.
		// We convert them into [-180, 180] degree range.
		phase.Data[i] = -math.Atan2(im, re) * rad2deg
		phaseUnwrapped.Data[i] = 0

		pwr := re*re + im*im
		power.Data[i] = pwr
		pmin = math.Min(pmin, pwr)
		pmax = math.Max(pmax, pwr)

		amp := math.Sqrt(pwr)
		amplitude.Data[i] = amp
		amin = math.Min(amin, amp)
		amax = math.Max(amax, amp)
	}

	unwrapPhaseDegrees(length, phase.Data, phaseUnwrapped, phaseDegreesUnwrappingLimit)
	phase.Min = -180
	phase.Max = 180
	power.Min = pmin
	power.Max = pmax
	amplitude.Min = amin
	amplitude.Max = amax
}

func unwrapPhaseDegrees(length int, wrapped []float64, unwrapped *Component, limit float64) {
	k := 0.0

	min := wrapped[0]
	max := min
	unwrapped.Data[0] = min

	for i := 1; i < length; i++ {
		w := wrapped[i]
		increment := wrapped[i] - wrapped[i-1]

		if increment > limit {
			k -= increment
		} else if increment < -limit {
			k += increment
		}

		w += k
		min = math.Min(min, w)
		max = math.Max(max, w)
		unwrapped.Data[i] = w
	}

	unwrapped.Min = min
	unwrapped.Max = max
}

func toDecibels(length int, src, tgt *Component) {
	const (
		five     = 5
		ten      = 10
		twenty   = 20
		hundreed = 100
	)

	dbmin := math.Inf(1)
	dbmax := math.Inf(-1)

	base := src.Data[0]
	if base < math.SmallestNonzeroFloat64 {
		base = src.Max
	}

	for i := 0; i < length; i++ {
		db := twenty * math.Log10(src.Data[i]/base)
		dbmin = math.Min(dbmin, db)
		dbmax = math.Max(dbmax, db)
		tgt.Data[i] = db
	}

	// If dbmin falls into one of [-100, -90), [-90, -80), ..., [-10, 0)
	// intervals, set it to the minimum value of the interval.
	for i := ten; i > 0; i-- {
		min := -float64(i) * ten
		max := -float64(i-1) * ten

		if dbmin >= min && dbmin < max {
			dbmin = min

			break
		}
	}

	// Limit all minimal decibel values to -100.
	if dbmin < -hundreed {
		dbmin = -hundreed

		for i := 0; i < length; i++ {
			if tgt.Data[i] < -hundreed {
				tgt.Data[i] = -hundreed
			}
		}
	}

	// If dbmax falls into one of [0, 5), [5, 10)
	// intervals, set it to the maximum value of the interval.
	for i := 2; i > 0; i-- {
		max := float64(i) * five
		min := float64(i-1) * five

		if dbmax >= min && dbmax < max {
			dbmax = max

			break
		}
	}

	// Limit all maximal decibel values to 10.
	if dbmax > ten {
		dbmax = ten

		for i := 0; i < length; i++ {
			if tgt.Data[i] > ten {
				tgt.Data[i] = ten
			}
		}
	}

	tgt.Min = dbmin
	tgt.Max = dbmax
}

func toPercents(length int, src, tgt *Component) {
	const (
		ten        = 10
		hundreed   = 100
		twohundred = 200
	)

	pctmin := 0.0
	pctmax := math.Inf(-1)

	base := src.Data[0]
	if base < math.SmallestNonzeroFloat64 {
		base = src.Max
	}

	for i := 0; i < length; i++ {
		pct := hundreed * src.Data[i] / base
		pctmax = math.Max(pctmax, pct)
		tgt.Data[i] = pct
	}

	// If pctmax falls into one of [100, 110), [110, 120), ..., [190, 200)
	// intervals, set it to the maximum value of the interval.
	for i := 0; i < ten; i++ {
		min := hundreed + float64(i)*ten
		max := hundreed + float64(i+1)*ten

		if pctmax >= min && pctmax < max {
			pctmax = max

			break
		}
	}

	// Limit all maximal percentage values to 200.
	if pctmax > twohundred {
		pctmax = twohundred

		for i := 0; i < length; i++ {
			if tgt.Data[i] > twohundred {
				tgt.Data[i] = twohundred
			}
		}
	}

	tgt.Min = pctmin
	tgt.Max = pctmax
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
