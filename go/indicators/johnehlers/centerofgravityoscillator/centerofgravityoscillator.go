package centerofgravityoscillator

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// CenterOfGravityOscillator (Ehler's Center of Gravity oscillator, COG) is
// described in Ehler's book "Cybernetic Analysis for Stocks and Futures" (2004).
//
// The center of gravity in a FIR filter is the position of the average price
// within the filter window length:
//
//	CGᵢ = Σ((i+1) * Priceᵢ) / Σ(Priceᵢ), where i = 0…ℓ-1, ℓ being a window size.
//
// The Center of Gravity oscillator has essentially zero lag and retains the
// relative cycle amplitude.
//
// It moves toward the most recent bar (decreases) when prices rise and moves
// away from the most recent bar (increases) when prices fall; thus moving
// exactly opposite to the price direction.
//
// The indicator has two outputs: the oscillator value and a trigger line which
// is the previous value of the oscillator.
//
// Reference:
//
//	Ehlers, John F. (2004). Cybernetic Analysis for Stocks and Futures. Wiley.
type CenterOfGravityOscillator struct {
	mu              sync.RWMutex
	mnemonic        string
	description     string
	mnemonicTrig    string
	descriptionTrig string
	value           float64
	valuePrevious   float64
	denominatorSum  float64
	length          int
	lengthMinOne    int
	windowCount     int
	window          []float64
	primed          bool
	barFunc         entities.BarFunc
	quoteFunc       entities.QuoteFunc
	tradeFunc       entities.TradeFunc
}

// NewCenterOfGravityOscillator returns an instance of the indicator
// created using supplied parameters.
func NewCenterOfGravityOscillator(
	params *Params,
) (*CenterOfGravityOscillator, error) {
	const (
		invalid   = "invalid center of gravity oscillator parameters"
		fmtl      = "%s: length should be a positive integer"
		fmtw      = "%s: %w"
		fmtn      = "cog(%d%s)"
		fmtnt     = "cogTrig(%d%s)"
		descr     = "Center of Gravity oscillator "
		descrTrig = "Center of Gravity trigger "
	)

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if params.Length < 1 {
		return nil, fmt.Errorf(fmtl, invalid)
	}

	// Resolve defaults for component functions.
	// A zero value means "use default, don't show in mnemonic".
	// CoG default bar component is MedianPrice, not ClosePrice.
	bc := params.BarComponent
	qc := params.QuoteComponent
	tc := params.TradeComponent

	if bc == 0 {
		bc = entities.BarMedianPrice
	}

	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	componentMnemonic := core.ComponentTripleMnemonic(bc, qc, tc)
	mnemonic := fmt.Sprintf(fmtn, params.Length, componentMnemonic)
	mnemonicTrig := fmt.Sprintf(fmtnt, params.Length, componentMnemonic)

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	return &CenterOfGravityOscillator{
		mnemonic:        mnemonic,
		description:     descr + mnemonic,
		mnemonicTrig:    mnemonicTrig,
		descriptionTrig: descrTrig + mnemonicTrig,
		length:          params.Length,
		lengthMinOne:    params.Length - 1,
		window:          make([]float64, params.Length),
		windowCount:     0,
		denominatorSum:  0,
		value:           math.NaN(),
		valuePrevious:   math.NaN(),
		primed:          false,
		barFunc:         barFunc,
		quoteFunc:       quoteFunc,
		tradeFunc:       tradeFunc,
	}, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *CenterOfGravityOscillator) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
func (s *CenterOfGravityOscillator) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.CenterOfGravityOscillator,
		Mnemonic:    s.mnemonic,
		Description: s.description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(Value),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonic,
				Description: s.description,
			},
			{
				Kind:        int(Trigger),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonicTrig,
				Description: s.descriptionTrig,
			},
		},
	}
}

// Update updates the value of the center of gravity oscillator given the next sample.
func (s *CenterOfGravityOscillator) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		s.valuePrevious = s.value
		s.value = s.calculate(sample)

		return s.value
	}

	// Not primed.
	if s.length > s.windowCount {
		s.denominatorSum += sample
		s.window[s.windowCount] = sample

		if s.lengthMinOne == s.windowCount {
			sum := 0.0
			if math.Abs(s.denominatorSum) > math.SmallestNonzeroFloat64 {
				for i := 0; i < s.length; i++ {
					sum += float64(1+i) * s.window[i]
				}

				sum /= s.denominatorSum
			}

			s.valuePrevious = sum
		}
	} else {
		s.value = s.calculate(sample)
		s.primed = true

		s.windowCount++

		return s.value
	}

	s.windowCount++

	return math.NaN()
}

func (s *CenterOfGravityOscillator) calculate(sample float64) float64 {
	s.denominatorSum += sample - s.window[0]

	for i := range s.lengthMinOne {
		s.window[i] = s.window[i+1]
	}

	s.window[s.lengthMinOne] = sample

	sum := 0.0
	if math.Abs(s.denominatorSum) > math.SmallestNonzeroFloat64 {
		for i, j := 0, 1; i < s.length; i, j = i+1, j+1 {
			sum += float64(j) * s.window[i]
		}

		sum /= s.denominatorSum
	}

	return sum
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *CenterOfGravityOscillator) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CenterOfGravityOscillator) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CenterOfGravityOscillator) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CenterOfGravityOscillator) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *CenterOfGravityOscillator) updateEntity(
	time time.Time, sample float64,
) core.Output {
	const length = 2

	output := make([]any, length)
	cog := s.Update(sample)

	trig := s.valuePrevious
	if math.IsNaN(cog) {
		trig = math.NaN()
	}

	i := 0
	output[i] = entities.Scalar{Time: time, Value: cog}
	i++
	output[i] = entities.Scalar{Time: time, Value: trig}

	return output
}
