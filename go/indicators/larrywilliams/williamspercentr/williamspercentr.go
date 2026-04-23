package williamspercentr

import (
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

const (
	willrMnemonic    = "willr"
	willrDescription = "Williams %R"
	defaultLength    = 14
	minLength        = 2
)

// WilliamsPercentR is Larry Williams' Williams %R momentum indicator.
//
// Williams %R reflects the level of the closing price relative to the
// highest high over a lookback period. The oscillation ranges from 0 to -100;
// readings from 0 to -20 are considered overbought, readings from -80 to -100
// are considered oversold.
//
// The value is calculated as:
//
//	%R = -100 * (HighestHigh - Close) / (HighestHigh - LowestLow)
//
// where HighestHigh and LowestLow are computed over the last `length` bars.
// If HighestHigh equals LowestLow, the value is 0.
//
// The indicator requires bar data (high, low, close). For scalar, quote, and
// trade updates, the single value is used as a substitute for all three.
//
// Reference:
//
// Williams, Larry (1979). How I Made One Million Dollars Last Year Trading Commodities.
type WilliamsPercentR struct {
	mu            sync.RWMutex
	length        int
	lengthMinOne  int
	circularIndex int
	circularCount int
	lowCircular   []float64
	highCircular  []float64
	value         float64
	primed        bool
}

// NewWilliamsPercentR returns a new instance of the Williams %R indicator.
// The length must be >= 2. If length < 2, the default length of 14 is used.
func NewWilliamsPercentR(length int) *WilliamsPercentR {
	if length < minLength {
		length = defaultLength
	}

	return &WilliamsPercentR{
		length:       length,
		lengthMinOne: length - 1,
		lowCircular:  make([]float64, length),
		highCircular: make([]float64, length),
		value:        math.NaN(),
	}
}

// IsPrimed indicates whether the indicator is primed.
func (s *WilliamsPercentR) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *WilliamsPercentR) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.WilliamsPercentR,
		willrMnemonic,
		willrDescription,
		[]core.OutputText{
			{Mnemonic: willrMnemonic, Description: willrDescription},
		},
	)
}

// Update updates the Williams %R given the next bar's close, high, and low values.
func (s *WilliamsPercentR) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	index := s.circularIndex
	s.lowCircular[index] = low
	s.highCircular[index] = high

	// Advance circular buffer index.
	s.circularIndex++
	if s.circularIndex > s.lengthMinOne {
		s.circularIndex = 0
	}

	if s.length > s.circularCount {
		if s.lengthMinOne == s.circularCount {
			// We have exactly `length` samples; compute for the first time.
			minLow := s.lowCircular[index]
			maxHigh := s.highCircular[index]

			for i := 0; i < s.lengthMinOne; i++ {
				// The value of index is always positive here (we started at lengthMinOne).
				index--

				if temp := s.lowCircular[index]; minLow > temp {
					minLow = temp
				}

				if temp := s.highCircular[index]; maxHigh < temp {
					maxHigh = temp
				}
			}

			if math.Abs(maxHigh-minLow) < math.SmallestNonzeroFloat64 {
				s.value = 0
			} else {
				s.value = -100 * (maxHigh - close) / (maxHigh - minLow)
			}

			s.primed = true
		}

		s.circularCount++

		return s.value
	}

	// Already primed, compute normally with wrapping.
	minLow := s.lowCircular[index]
	maxHigh := s.highCircular[index]

	for i := 0; i < s.lengthMinOne; i++ {
		if index == 0 {
			index = s.lengthMinOne
		} else {
			index--
		}

		if temp := s.lowCircular[index]; minLow > temp {
			minLow = temp
		}

		if temp := s.highCircular[index]; maxHigh < temp {
			maxHigh = temp
		}
	}

	if math.Abs(maxHigh-minLow) < math.SmallestNonzeroFloat64 {
		s.value = 0
	} else {
		s.value = -100 * (maxHigh - close) / (maxHigh - minLow)
	}

	return s.value
}

// UpdateSample updates the Williams %R using a single sample value
// as a substitute for high, low, and close.
func (s *WilliamsPercentR) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *WilliamsPercentR) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *WilliamsPercentR) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *WilliamsPercentR) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *WilliamsPercentR) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}
