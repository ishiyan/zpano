package zerolagerrorcorrectingexponentialmovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// ZeroLagErrorCorrectingExponentialMovingAverage is Ehler's adaptive zero-lag
// error-correcting exponential moving average (ZECEMA).
//
// The algorithm iterates over gain values in [-gainLimit, gainLimit] with the given
// gainStep to find the gain that minimizes the error between the sample and the
// error-corrected EMA value.
//
// The indicator is not primed during the first two updates; it primes on the third.
//
// Reference:
//
// John Ehlers and Ric Way, 'Zero Lag (well, almost)', TASC, 2010, v28.11, pp30-35.
type ZeroLagErrorCorrectingExponentialMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	alpha       float64
	oneMinAlpha float64
	gainLimit   float64
	gainStep    float64
	length      int
	count       int
	value       float64
	emaValue    float64
	primed      bool
}

// NewZeroLagErrorCorrectingExponentialMovingAverage returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewZeroLagErrorCorrectingExponentialMovingAverage(p *ZeroLagErrorCorrectingExponentialMovingAverageParams) (*ZeroLagErrorCorrectingExponentialMovingAverage, error) {
	const (
		invalid = "invalid zero-lag error-correcting exponential moving average parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "zecema(%.4g, %.4g, %.4g%s)"
		epsilon = 0.00000001
	)

	sf := p.SmoothingFactor
	if sf <= 0 || sf > 1 {
		return nil, fmt.Errorf(fmts, invalid, "smoothing factor should be in (0, 1]")
	}

	gl := p.GainLimit
	if gl <= 0 {
		return nil, fmt.Errorf(fmts, invalid, "gain limit should be positive")
	}

	gs := p.GainStep
	if gs <= 0 {
		return nil, fmt.Errorf(fmts, invalid, "gain step should be positive")
	}

	// Resolve defaults for component functions.
	bc := p.BarComponent
	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	qc := p.QuoteComponent
	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	tc := p.TradeComponent
	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	// Calculate equivalent length.
	var length int
	if sf < epsilon {
		length = math.MaxInt
	} else {
		length = int(math.Round(2.0/sf)) - 1
	}

	// Build mnemonic.
	mnemonic := fmt.Sprintf(fmtn, sf, gl, gs, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Zero-lag Error-Correcting Exponential Moving Average " + mnemonic

	z := &ZeroLagErrorCorrectingExponentialMovingAverage{
		alpha:       sf,
		oneMinAlpha: 1 - sf,
		gainLimit:   gl,
		gainStep:    gs,
		length:      length,
		value:       math.NaN(),
		emaValue:    math.NaN(),
	}

	z.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, z.Update)

	return z, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.ZeroLagErrorCorrectingExponentialMovingAverage,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the indicator given the next sample.
//
// The indicator is not primed during the first two updates; it primes on the third.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		s.value = s.calculate(sample)

		return s.value
	}

	s.count++

	if s.count == 1 {
		s.emaValue = sample

		return math.NaN()
	}

	if s.count == 2 {
		s.emaValue = s.calculateEma(sample)
		s.value = s.emaValue

		return math.NaN()
	}

	// count == 3: prime the indicator.
	s.value = s.calculate(sample)
	s.primed = true

	return s.value
}

func (s *ZeroLagErrorCorrectingExponentialMovingAverage) calculateEma(sample float64) float64 {
	return s.alpha*sample + s.oneMinAlpha*s.emaValue
}

func (s *ZeroLagErrorCorrectingExponentialMovingAverage) calculate(sample float64) float64 {
	s.emaValue = s.calculateEma(sample)

	leastError := math.MaxFloat64
	bestEC := 0.0

	for gain := -s.gainLimit; gain <= s.gainLimit; gain += s.gainStep {
		ec := s.alpha*(s.emaValue+gain*(sample-s.value)) + s.oneMinAlpha*s.value
		err := math.Abs(sample - ec)

		if leastError > err {
			leastError = err
			bestEC = ec
		}
	}

	return bestEC
}
