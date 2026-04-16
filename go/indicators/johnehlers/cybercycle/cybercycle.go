package cybercycle

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

// CyberCycle (Ehler's Cyber Cycle, CC) is described in Ehler's book
// "Cybernetic Analysis for Stocks and Futures" (2004):
//
//	H(z) = ((1-α/2)²(1 - 2z⁻¹ + z⁻²)) / (1 - 2(1-α)z⁻¹ + (1-α)²z⁻²)
//
// which is a complementary high-pass filter found by subtracting the
// Instantaneous Trend Line low-pass filter from unity.
//
// The Cyber Cycle has zero lag and retains the relative cycle amplitude.
//
// The indicator has two outputs: the cycle value and a signal line which
// is an exponential moving average of the cycle value.
//
// Reference:
//
//	Ehlers, John F. (2004). Cybernetic Analysis for Stocks and Futures. Wiley.
type CyberCycle struct {
	mu                sync.RWMutex
	length            int
	smoothingFactor   float64
	signalLag         int
	mnemonic          string
	description       string
	mnemonicSig       string
	descriptionSig    string
	coeff1            float64
	coeff2            float64
	coeff3            float64
	coeff4            float64
	coeff5            float64
	count             int
	previousSample1   float64
	previousSample2   float64
	previousSample3   float64
	smoothed          float64
	previousSmoothed1 float64
	previousSmoothed2 float64
	value             float64
	previousValue1    float64
	previousValue2    float64
	signal            float64
	primed            bool
	barFunc           entities.BarFunc
	quoteFunc         entities.QuoteFunc
	tradeFunc         entities.TradeFunc
}

// NewCyberCycleLength returns an instance of the indicator
// created using supplied parameters based on length.
func NewCyberCycleLength(p *LengthParams) (*CyberCycle, error) {
	return newCyberCycle(p.Length, math.NaN(), p.SignalLag,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

// NewCyberCycleSmoothingFactor returns an instance of the indicator
// created using supplied parameters based on smoothing factor.
func NewCyberCycleSmoothingFactor(p *SmoothingFactorParams) (*CyberCycle, error) {
	return newCyberCycle(0, p.SmoothingFactor, p.SignalLag,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

//nolint:funlen,cyclop
func newCyberCycle(length int, alpha float64, signalLag int,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*CyberCycle, error) {
	const (
		invalid  = "invalid cyber cycle parameters"
		fmts     = "%s: %s"
		fmtw     = "%s: %w"
		fmtn     = "cc(%d%s)"
		fmtns    = "ccSignal(%d%s)"
		descr    = "Cyber Cycle "
		descrSig = "Cyber Cycle signal "
		epsilon  = 0.00000001
		two      = 2.
	)

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if math.IsNaN(alpha) {
		// Length-based construction.
		if length < 1 {
			return nil, fmt.Errorf(fmts, invalid, "length should be a positive integer")
		}

		alpha = two / float64(1+length)
	} else {
		// Smoothing-factor-based construction.
		if alpha < 0 || alpha > 1 {
			return nil, fmt.Errorf(fmts, invalid, "smoothing factor should be in range [0, 1]")
		}

		if alpha < epsilon {
			length = math.MaxInt
		} else {
			length = int(math.Round(two/alpha)) - 1
		}
	}

	if signalLag < 1 {
		return nil, fmt.Errorf(fmts, invalid, "signal lag should be a positive integer")
	}

	// Resolve defaults for component functions.
	// CyberCycle default bar component is MedianPrice, not ClosePrice.
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
	mnemonic := fmt.Sprintf(fmtn, length, componentMnemonic)
	mnemonicSig := fmt.Sprintf(fmtns, length, componentMnemonic)

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	// Calculate coefficients.
	x := 1 - alpha/2
	c1 := x * x

	x = 1 - alpha
	c2 := 2 * x
	c3 := -x * x

	x = 1 / float64(1+signalLag)
	c4 := x
	c5 := 1 - x

	return &CyberCycle{
		length:          length,
		smoothingFactor: alpha,
		signalLag:       signalLag,
		mnemonic:        mnemonic,
		description:     descr + mnemonic,
		mnemonicSig:     mnemonicSig,
		descriptionSig:  descrSig + mnemonicSig,
		coeff1:          c1,
		coeff2:          c2,
		coeff3:          c3,
		coeff4:          c4,
		coeff5:          c5,
		value:           math.NaN(),
		signal:          math.NaN(),
		primed:          false,
		barFunc:         barFunc,
		quoteFunc:       quoteFunc,
		tradeFunc:       tradeFunc,
	}, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *CyberCycle) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
func (s *CyberCycle) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.CyberCycle,
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
				Kind:        int(Signal),
				Type:        outputs.ScalarType,
				Mnemonic:    s.mnemonicSig,
				Description: s.descriptionSig,
			},
		},
	}
}

// Update updates the value of the cyber cycle given the next sample.
//
//nolint:funlen,cyclop
func (s *CyberCycle) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		s.previousSmoothed2 = s.previousSmoothed1
		s.previousSmoothed1 = s.smoothed
		s.smoothed = (sample + 2*s.previousSample1 + 2*s.previousSample2 + s.previousSample3) / 6

		s.previousValue2 = s.previousValue1
		s.previousValue1 = s.value
		s.value = s.coeff1*(s.smoothed-2*s.previousSmoothed1+s.previousSmoothed2) +
			s.coeff2*s.previousValue1 + s.coeff3*s.previousValue2

		s.signal = s.coeff4*s.value + s.coeff5*s.signal

		s.previousSample3 = s.previousSample2
		s.previousSample2 = s.previousSample1
		s.previousSample1 = sample

		return s.value
	}

	s.count++

	switch s.count {
	case 1:
		s.previousSample3 = sample

		return math.NaN()
	case 2:
		s.previousSample2 = sample

		return math.NaN()
	case 3:
		s.signal = s.coeff4 * (sample - 2*s.previousSample2 + s.previousSample3) / 4
		s.previousSample1 = sample

		return math.NaN()
	case 4:
		s.previousSmoothed2 = (sample + 2*s.previousSample1 + 2*s.previousSample2 + s.previousSample3) / 6
		s.signal = s.coeff4*(sample-2*s.previousSample1+s.previousSample2)/4 + s.coeff5*s.signal

		s.previousSample3 = s.previousSample2
		s.previousSample2 = s.previousSample1
		s.previousSample1 = sample

		return math.NaN()
	case 5:
		s.previousSmoothed1 = (sample + 2*s.previousSample1 + 2*s.previousSample2 + s.previousSample3) / 6
		s.signal = s.coeff4*(sample-2*s.previousSample1+s.previousSample2)/4 + s.coeff5*s.signal

		s.previousSample3 = s.previousSample2
		s.previousSample2 = s.previousSample1
		s.previousSample1 = sample

		return math.NaN()
	case 6:
		s.smoothed = (sample + 2*s.previousSample1 + 2*s.previousSample2 + s.previousSample3) / 6
		s.previousValue2 = (sample - 2*s.previousSample1 + s.previousSample2) / 4
		s.signal = s.coeff4*s.previousValue2 + s.coeff5*s.signal

		s.previousSample3 = s.previousSample2
		s.previousSample2 = s.previousSample1
		s.previousSample1 = sample

		return math.NaN()
	case 7:
		s.previousSmoothed2 = s.previousSmoothed1
		s.previousSmoothed1 = s.smoothed
		s.smoothed = (sample + 2*s.previousSample1 + 2*s.previousSample2 + s.previousSample3) / 6
		s.previousValue1 = (sample - 2*s.previousSample1 + s.previousSample2) / 4
		s.signal = s.coeff4*s.previousValue1 + s.coeff5*s.signal

		s.previousSample3 = s.previousSample2
		s.previousSample2 = s.previousSample1
		s.previousSample1 = sample

		return math.NaN()
	case 8:
		s.previousSmoothed2 = s.previousSmoothed1
		s.previousSmoothed1 = s.smoothed
		s.smoothed = (sample + 2*s.previousSample1 + 2*s.previousSample2 + s.previousSample3) / 6

		s.value = s.coeff1*(s.smoothed-2*s.previousSmoothed1+s.previousSmoothed2) +
			s.coeff2*s.previousValue1 + s.coeff3*s.previousValue2

		s.signal = s.coeff4*s.value + s.coeff5*s.signal

		s.previousSample3 = s.previousSample2
		s.previousSample2 = s.previousSample1
		s.previousSample1 = sample
		s.primed = true

		return s.value
	}

	return math.NaN()
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *CyberCycle) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *CyberCycle) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *CyberCycle) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *CyberCycle) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *CyberCycle) updateEntity(
	time time.Time, sample float64,
) core.Output {
	const length = 2

	output := make([]any, length)
	v := s.Update(sample)

	sig := s.signal
	if math.IsNaN(v) {
		sig = math.NaN()
	}

	i := 0
	output[i] = entities.Scalar{Time: time, Value: v}
	i++
	output[i] = entities.Scalar{Time: time, Value: sig}

	return output
}
