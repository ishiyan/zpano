package williamspercentr

import (
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
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
func (w *WilliamsPercentR) IsPrimed() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.primed
}

// Metadata describes the output data of the indicator.
func (w *WilliamsPercentR) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.WilliamsPercentR,
		Mnemonic:    willrMnemonic,
		Description: willrDescription,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(WilliamsPercentRValue),
				Type:        outputs.ScalarType,
				Mnemonic:    willrMnemonic,
				Description: willrDescription,
			},
		},
	}
}

// Update updates the Williams %R given the next bar's close, high, and low values.
func (w *WilliamsPercentR) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	index := w.circularIndex
	w.lowCircular[index] = low
	w.highCircular[index] = high

	// Advance circular buffer index.
	w.circularIndex++
	if w.circularIndex > w.lengthMinOne {
		w.circularIndex = 0
	}

	if w.length > w.circularCount {
		if w.lengthMinOne == w.circularCount {
			// We have exactly `length` samples; compute for the first time.
			minLow := w.lowCircular[index]
			maxHigh := w.highCircular[index]

			for i := 0; i < w.lengthMinOne; i++ {
				// The value of index is always positive here (we started at lengthMinOne).
				index--

				if temp := w.lowCircular[index]; minLow > temp {
					minLow = temp
				}

				if temp := w.highCircular[index]; maxHigh < temp {
					maxHigh = temp
				}
			}

			if math.Abs(maxHigh-minLow) < math.SmallestNonzeroFloat64 {
				w.value = 0
			} else {
				w.value = -100 * (maxHigh - close) / (maxHigh - minLow)
			}

			w.primed = true
		}

		w.circularCount++

		return w.value
	}

	// Already primed, compute normally with wrapping.
	minLow := w.lowCircular[index]
	maxHigh := w.highCircular[index]

	for i := 0; i < w.lengthMinOne; i++ {
		if index == 0 {
			index = w.lengthMinOne
		} else {
			index--
		}

		if temp := w.lowCircular[index]; minLow > temp {
			minLow = temp
		}

		if temp := w.highCircular[index]; maxHigh < temp {
			maxHigh = temp
		}
	}

	if math.Abs(maxHigh-minLow) < math.SmallestNonzeroFloat64 {
		w.value = 0
	} else {
		w.value = -100 * (maxHigh - close) / (maxHigh - minLow)
	}

	return w.value
}

// UpdateSample updates the Williams %R using a single sample value
// as a substitute for high, low, and close.
func (w *WilliamsPercentR) UpdateSample(sample float64) float64 {
	return w.Update(sample, sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (w *WilliamsPercentR) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: w.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (w *WilliamsPercentR) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: w.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (w *WilliamsPercentR) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: w.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (w *WilliamsPercentR) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: w.Update(v, v, v)}

	return output
}
