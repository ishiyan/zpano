package instantaneoustrendline

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

// InstantaneousTrendLine (Ehler's Instantaneous Trend Line, iTrend) is described
// in Ehler's book "Cybernetic Analysis for Stocks and Futures" (2004):
//
//	H(z) = ((α-α²/4) + α²z⁻¹/2 - (α-3α²/4)z⁻²) / (1 - 2(1-α)z⁻¹ + (1-α)²z⁻²)
//
// which is a complementary low-pass filter found by subtracting the CyberCycle
// high-pass filter from unity.
//
// The Instantaneous Trend Line has zero lag and about the same smoothing as an
// Exponential Moving Average with the same α.
//
// The indicator has two outputs: the trend line value and a trigger line.
//
// Reference:
//
//	Ehlers, John F. (2004). Cybernetic Analysis for Stocks and Futures. Wiley.
type InstantaneousTrendLine struct {
	mu                 sync.RWMutex
	length             int
	smoothingFactor    float64
	mnemonic           string
	description        string
	mnemonicTrig       string
	descriptionTrig    string
	coeff1             float64
	coeff2             float64
	coeff3             float64
	coeff4             float64
	coeff5             float64
	count              int
	previousSample1    float64
	previousSample2    float64
	previousTrendLine1 float64
	previousTrendLine2 float64
	trendLine          float64
	triggerLine        float64
	primed             bool
	barFunc            entities.BarFunc
	quoteFunc          entities.QuoteFunc
	tradeFunc          entities.TradeFunc
}

// NewInstantaneousTrendLineLength returns an instance of the indicator
// created using supplied parameters based on length.
func NewInstantaneousTrendLineLength(p *LengthParams) (*InstantaneousTrendLine, error) {
	return newInstantaneousTrendLine(p.Length, math.NaN(),
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

// NewInstantaneousTrendLineSmoothingFactor returns an instance of the indicator
// created using supplied parameters based on smoothing factor.
func NewInstantaneousTrendLineSmoothingFactor(p *SmoothingFactorParams) (*InstantaneousTrendLine, error) {
	return newInstantaneousTrendLine(0, p.SmoothingFactor,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

//nolint:funlen,cyclop
func newInstantaneousTrendLine(length int, alpha float64,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*InstantaneousTrendLine, error) {
	const (
		invalid  = "invalid instantaneous trend line parameters"
		fmts     = "%s: %s"
		fmtw     = "%s: %w"
		fmtn     = "iTrend(%d%s)"
		fmtnTrig = "iTrendTrigger(%d%s)"
		descr    = "Instantaneous Trend Line "
		descrTr  = "Instantaneous Trend Line trigger "
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

	// Resolve defaults for component functions.
	// InstantaneousTrendLine default bar component is MedianPrice, not ClosePrice.
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
	mnemonicTrig := fmt.Sprintf(fmtnTrig, length, componentMnemonic)

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
	// H(z) = ((α-α²/4) + α²z⁻¹/2 - (α-3α²/4)z⁻²) / (1 - 2(1-α)z⁻¹ + (1-α)²z⁻²)
	a2 := alpha * alpha
	c1 := alpha - a2/4
	c2 := a2 / 2
	c3 := -(alpha - 3*a2/4)

	x := 1 - alpha
	c4 := 2 * x
	c5 := -(x * x)

	return &InstantaneousTrendLine{
		length:          length,
		smoothingFactor: alpha,
		mnemonic:        mnemonic,
		description:     descr + mnemonic,
		mnemonicTrig:    mnemonicTrig,
		descriptionTrig: descrTr + mnemonicTrig,
		coeff1:          c1,
		coeff2:          c2,
		coeff3:          c3,
		coeff4:          c4,
		coeff5:          c5,
		trendLine:       math.NaN(),
		triggerLine:     math.NaN(),
		primed:          false,
		barFunc:         barFunc,
		quoteFunc:       quoteFunc,
		tradeFunc:       tradeFunc,
	}, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *InstantaneousTrendLine) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
func (s *InstantaneousTrendLine) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.InstantaneousTrendLine,
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

// Update updates the value of the instantaneous trend line given the next sample.
//
//nolint:funlen,cyclop
func (s *InstantaneousTrendLine) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		s.trendLine = s.coeff1*sample + s.coeff2*s.previousSample1 + s.coeff3*s.previousSample2 +
			s.coeff4*s.previousTrendLine1 + s.coeff5*s.previousTrendLine2
		s.triggerLine = 2*s.trendLine - s.previousTrendLine2

		s.previousSample2 = s.previousSample1
		s.previousSample1 = sample
		s.previousTrendLine2 = s.previousTrendLine1
		s.previousTrendLine1 = s.trendLine

		return s.trendLine
	}

	s.count++

	switch s.count {
	case 1:
		s.previousSample2 = sample

		return math.NaN()
	case 2:
		s.previousSample1 = sample

		return math.NaN()
	case 3:
		s.previousTrendLine2 = (sample + 2*s.previousSample1 + s.previousSample2) / 4

		s.previousSample2 = s.previousSample1
		s.previousSample1 = sample

		return math.NaN()
	case 4:
		s.previousTrendLine1 = (sample + 2*s.previousSample1 + s.previousSample2) / 4

		s.previousSample2 = s.previousSample1
		s.previousSample1 = sample

		return math.NaN()
	case 5:
		s.trendLine = s.coeff1*sample + s.coeff2*s.previousSample1 + s.coeff3*s.previousSample2 +
			s.coeff4*s.previousTrendLine1 + s.coeff5*s.previousTrendLine2
		s.triggerLine = 2*s.trendLine - s.previousTrendLine2

		s.previousSample2 = s.previousSample1
		s.previousSample1 = sample
		s.previousTrendLine2 = s.previousTrendLine1
		s.previousTrendLine1 = s.trendLine
		s.primed = true

		return s.trendLine
	}

	return math.NaN()
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *InstantaneousTrendLine) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *InstantaneousTrendLine) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *InstantaneousTrendLine) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *InstantaneousTrendLine) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *InstantaneousTrendLine) updateEntity(
	time time.Time, sample float64,
) core.Output {
	const length = 2

	output := make([]any, length)
	v := s.Update(sample)

	trig := s.triggerLine
	if math.IsNaN(v) {
		trig = math.NaN()
	}

	i := 0
	output[i] = entities.Scalar{Time: time, Value: v}
	i++
	output[i] = entities.Scalar{Time: time, Value: trig}

	return output
}
