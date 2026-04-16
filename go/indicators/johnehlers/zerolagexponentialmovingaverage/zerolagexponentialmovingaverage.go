package zerolagexponentialmovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// ZeroLagExponentialMovingAverage is Ehler's Zero-lag Exponential Moving Average (ZEMA).
//
// ZEMA = alpha*(Price + gainFactor*(Price - Price[momentumLength ago])) + (1 - alpha)*ZEMA[prev]
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
func (z *ZeroLagExponentialMovingAverage) IsPrimed() bool {
	z.mu.RLock()
	defer z.mu.RUnlock()

	return z.primed
}

// Metadata describes the output data of the indicator.
func (z *ZeroLagExponentialMovingAverage) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.ZeroLagExponentialMovingAverage,
		Mnemonic:    z.LineIndicator.Mnemonic,
		Description: z.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(ZeroLagExponentialMovingAverageValue),
				Type:        outputs.ScalarType,
				Mnemonic:    z.LineIndicator.Mnemonic,
				Description: z.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the indicator given the next sample.
//
// The indicator is not primed during the first VelocityMomentumLength updates.
func (z *ZeroLagExponentialMovingAverage) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	z.mu.Lock()
	defer z.mu.Unlock()

	if z.primed {
		// Shift momentum window left by 1.
		copy(z.momentumWindow, z.momentumWindow[1:])
		z.momentumWindow[z.momentumLength] = sample
		z.value = z.calculate(sample)

		return z.value
	}

	z.momentumWindow[z.count] = sample
	z.count++

	if z.count <= z.momentumLength {
		z.value = sample

		return math.NaN()
	}

	// count == momentumLength + 1: prime the indicator.
	z.value = z.calculate(sample)
	z.primed = true

	return z.value
}

func (z *ZeroLagExponentialMovingAverage) calculate(sample float64) float64 {
	momentum := sample - z.momentumWindow[0]

	return z.alpha*(sample+z.gainFactor*momentum) + z.oneMinAlpha*z.value
}
