package zerolagexponentialmovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// ZeroLagExponentialMovingAverage is Ehler's Zero-lag Exponential Moving Average (ZEMA).
//
// ZEMA = alpha*(Price + gainFactor*(Price - Price[momentumLength ago])) + (1 - alpha)*ZEMA[previous]
//
// The indicator is not primed during the first VelocityMomentumLength updates.
//
// Reference:
//
// Ehlers, John F. (2001). Rocket Science for Traders. Wiley. pp 167-170.
type ZeroLagExponentialMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	alpha          float64
	oneMinAlpha    float64
	gainFactor     float64
	momentumLength int
	momentumWindow []float64
	length         int
	count          int
	value          float64
	primed         bool
}

// NewZeroLagExponentialMovingAverage returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewZeroLagExponentialMovingAverage(p *ZeroLagExponentialMovingAverageParams) (*ZeroLagExponentialMovingAverage, error) {
	const (
		invalid = "invalid zero-lag exponential moving average parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "zema(%.4g, %.4g, %d%s)"
		epsilon = 0.00000001
	)

	sf := p.SmoothingFactor
	if sf <= 0 || sf > 1 {
		return nil, fmt.Errorf(fmts, invalid, "smoothing factor should be in (0, 1]")
	}

	ml := p.VelocityMomentumLength
	if ml < 1 {
		return nil, fmt.Errorf(fmts, invalid, "velocity momentum length should be positive")
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

	gf := p.VelocityGainFactor

	// Build mnemonic.
	mnemonic := fmt.Sprintf(fmtn, sf, gf, ml, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Zero-lag Exponential Moving Average " + mnemonic

	z := &ZeroLagExponentialMovingAverage{
		alpha:          sf,
		oneMinAlpha:    1 - sf,
		gainFactor:     gf,
		momentumLength: ml,
		momentumWindow: make([]float64, ml+1),
		length:         length,
		value:          math.NaN(),
	}

	z.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, z.Update)

	return z, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *ZeroLagExponentialMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *ZeroLagExponentialMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.ZeroLagExponentialMovingAverage,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the indicator given the next sample.
//
// The indicator is not primed during the first VelocityMomentumLength updates.
func (s *ZeroLagExponentialMovingAverage) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		// Shift momentum window left by 1.
		copy(s.momentumWindow, s.momentumWindow[1:])
		s.momentumWindow[s.momentumLength] = sample
		s.value = s.calculate(sample)

		return s.value
	}

	s.momentumWindow[s.count] = sample
	s.count++

	if s.count <= s.momentumLength {
		s.value = sample

		return math.NaN()
	}

	// count == momentumLength + 1: prime the indicator.
	s.value = s.calculate(sample)
	s.primed = true

	return s.value
}

func (s *ZeroLagExponentialMovingAverage) calculate(sample float64) float64 {
	momentum := sample - s.momentumWindow[0]

	return s.alpha*(sample+s.gainFactor*momentum) + s.oneMinAlpha*s.value
}
