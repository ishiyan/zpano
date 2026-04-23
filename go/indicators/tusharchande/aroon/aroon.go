package aroon

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// Aroon is Tushar Chande's Aroon indicator.
//
// The Aroon indicator measures the number of periods since the highest high
// and lowest low within a lookback window. It produces three outputs:
//   - Up: 100 * (Length - periods since highest high) / Length
//   - Down: 100 * (Length - periods since lowest low) / Length
//   - Osc: Up - Down
//
// The indicator requires bar data (high, low). For scalar, quote, and
// trade updates, the single value substitutes for both.
//
// Reference:
//
// Chande, Tushar S. (1995). "The New Technical Trader". John Wiley & Sons.
type Aroon struct {
	mu sync.RWMutex

	length int
	factor float64

	// Circular buffers for high and low values (size = length+1).
	highBuf     []float64
	lowBuf      []float64
	bufferIndex int
	count       int

	// Tracked indices of highest high and lowest low (absolute indices).
	highestIndex int
	lowestIndex  int

	up     float64
	down   float64
	osc    float64
	primed bool

	mnemonic string
}

// NewAroon returns an instance of the indicator created using supplied parameters.
func NewAroon(p *AroonParams) (*Aroon, error) {
	const (
		invalid   = "invalid aroon parameters"
		fmts      = "%s: %s"
		minLength = 2
	)

	if p.Length < minLength {
		return nil, fmt.Errorf(fmts, invalid, "length should be greater than 1")
	}

	mnemonic := fmt.Sprintf("aroon(%d)", p.Length)
	windowSize := p.Length + 1

	return &Aroon{
		length:   p.Length,
		factor:   100.0 / float64(p.Length), //nolint:mnd
		highBuf:  make([]float64, windowSize),
		lowBuf:   make([]float64, windowSize),
		up:       math.NaN(),
		down:     math.NaN(),
		osc:      math.NaN(),
		mnemonic: mnemonic,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (a *Aroon) IsPrimed() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.primed
}

// Metadata describes the output data of the indicator.
func (a *Aroon) Metadata() core.Metadata {
	desc := "Aroon " + a.mnemonic

	return core.BuildMetadata(
		core.Aroon,
		a.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: a.mnemonic + " up", Description: desc + " Up"},
			{Mnemonic: a.mnemonic + " down", Description: desc + " Down"},
			{Mnemonic: a.mnemonic + " osc", Description: desc + " Oscillator"},
		},
	)
}

// Update updates the indicator given the next bar's high and low values.
// Returns Up, Down, and Osc.
func (a *Aroon) Update(high, low float64) (float64, float64, float64) {
	if math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN(), math.NaN(), math.NaN()
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	windowSize := a.length + 1
	today := a.count

	// Store in circular buffer.
	pos := a.bufferIndex
	a.highBuf[pos] = high
	a.lowBuf[pos] = low
	a.bufferIndex = (a.bufferIndex + 1) % windowSize
	a.count++

	// Need at least length+1 bars (indices 0..length).
	if a.count < windowSize {
		return a.up, a.down, a.osc
	}

	trailingIndex := today - a.length

	if a.count == windowSize {
		// First time: scan entire window to find highest/lowest.
		a.highestIndex = trailingIndex
		a.lowestIndex = trailingIndex

		for i := trailingIndex + 1; i <= today; i++ {
			bufPos := i % windowSize

			if a.highBuf[bufPos] >= a.highBuf[a.highestIndex%windowSize] {
				a.highestIndex = i
			}

			if a.lowBuf[bufPos] <= a.lowBuf[a.lowestIndex%windowSize] {
				a.lowestIndex = i
			}
		}
	} else {
		// Subsequent: optimized update.
		// Check if tracked indices fell out of window.
		if a.highestIndex < trailingIndex {
			a.highestIndex = trailingIndex

			for i := trailingIndex + 1; i <= today; i++ {
				bufPos := i % windowSize
				if a.highBuf[bufPos] >= a.highBuf[a.highestIndex%windowSize] {
					a.highestIndex = i
				}
			}
		} else if high >= a.highBuf[a.highestIndex%windowSize] {
			a.highestIndex = today
		}

		if a.lowestIndex < trailingIndex {
			a.lowestIndex = trailingIndex

			for i := trailingIndex + 1; i <= today; i++ {
				bufPos := i % windowSize
				if a.lowBuf[bufPos] <= a.lowBuf[a.lowestIndex%windowSize] {
					a.lowestIndex = i
				}
			}
		} else if low <= a.lowBuf[a.lowestIndex%windowSize] {
			a.lowestIndex = today
		}
	}

	a.up = a.factor * float64(a.length-(today-a.highestIndex))
	a.down = a.factor * float64(a.length-(today-a.lowestIndex))
	a.osc = a.up - a.down

	if !a.primed {
		a.primed = true
	}

	return a.up, a.down, a.osc
}

// UpdateScalar updates the indicator given the next scalar sample.
func (a *Aroon) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value
	up, down, osc := a.Update(v, v)

	output := make([]any, 3) //nolint:mnd
	output[0] = entities.Scalar{Time: sample.Time, Value: up}
	output[1] = entities.Scalar{Time: sample.Time, Value: down}
	output[2] = entities.Scalar{Time: sample.Time, Value: osc}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (a *Aroon) UpdateBar(sample *entities.Bar) core.Output {
	up, down, osc := a.Update(sample.High, sample.Low)

	output := make([]any, 3) //nolint:mnd
	output[0] = entities.Scalar{Time: sample.Time, Value: up}
	output[1] = entities.Scalar{Time: sample.Time, Value: down}
	output[2] = entities.Scalar{Time: sample.Time, Value: osc}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (a *Aroon) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	return a.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (a *Aroon) UpdateTrade(sample *entities.Trade) core.Output {
	return a.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: sample.Price})
}
