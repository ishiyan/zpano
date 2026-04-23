package movingaverageconvergencedivergence

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/common/exponentialmovingaverage"
	"zpano/indicators/common/simplemovingaverage"
	"zpano/indicators/core"
)

// lineUpdater is an interface for indicators that accept a single scalar and return a value.
type lineUpdater interface {
	Update(float64) float64
	IsPrimed() bool
}

// MovingAverageConvergenceDivergence is Gerald Appel's MACD indicator.
//
// MACD is calculated by subtracting the slow moving average from the fast moving average.
// A signal line (moving average of MACD) and histogram (MACD minus signal) are also produced.
//
// The indicator produces three outputs:
//   - MACD: fast MA - slow MA
//   - Signal: MA of the MACD line
//   - Histogram: MACD - Signal
//
// Reference:
//
// Appel, Gerald (2005). Technical Analysis: Power Tools for Active Investors.
type MovingAverageConvergenceDivergence struct {
	mu sync.RWMutex

	fastMA   lineUpdater
	slowMA   lineUpdater
	signalMA lineUpdater

	macdValue      float64
	signalValue    float64
	histogramValue float64
	primed         bool

	// fastDelay is the number of initial samples to skip before feeding the fast MA.
	// This aligns the fast MA's SMA seed window with the slow MA's, matching TaLib's
	// batch algorithm where both MAs start producing output at the same index.
	fastDelay int
	fastCount int

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	mnemonic string
}

// NewMovingAverageConvergenceDivergence returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewMovingAverageConvergenceDivergence(
	p *MovingAverageConvergenceDivergenceParams,
) (*MovingAverageConvergenceDivergence, error) {
	const (
		invalid           = "invalid moving average convergence divergence parameters"
		fmts              = "%s: %s"
		fmtw              = "%s: %w"
		minLength         = 2
		minSignalLength   = 1
		defaultFastLength = 12
		defaultSlowLength = 26
		defaultSignalLen  = 9
	)

	fastLength := p.FastLength
	if fastLength == 0 {
		fastLength = defaultFastLength
	}

	slowLength := p.SlowLength
	if slowLength == 0 {
		slowLength = defaultSlowLength
	}

	signalLength := p.SignalLength
	if signalLength == 0 {
		signalLength = defaultSignalLen
	}

	if fastLength < minLength {
		return nil, fmt.Errorf(fmts, invalid, "fast length should be greater than 1")
	}

	if slowLength < minLength {
		return nil, fmt.Errorf(fmts, invalid, "slow length should be greater than 1")
	}

	if signalLength < minSignalLength {
		return nil, fmt.Errorf(fmts, invalid, "signal length should be greater than 0")
	}

	// Auto-swap fast/slow if needed (matches TaLib behavior).
	if slowLength < fastLength {
		fastLength, slowLength = slowLength, fastLength
	}

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

	// Default FirstIsAverage to true (TA-Lib compatible: EMA seeded with SMA).
	firstIsAverage := true
	if p.FirstIsAverage != nil {
		firstIsAverage = *p.FirstIsAverage
	}

	var fastMA, slowMA, signalMA lineUpdater

	// Create fast and slow MAs.
	fastMA, err = newMA(p.MovingAverageType, fastLength, firstIsAverage)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	slowMA, err = newMA(p.MovingAverageType, slowLength, firstIsAverage)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	// Create signal MA.
	signalMA, err = newMA(p.SignalMovingAverageType, signalLength, firstIsAverage)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	mnemonic := buildMnemonic(fastLength, slowLength, signalLength,
		p.MovingAverageType, p.SignalMovingAverageType, bc, qc, tc)

	return &MovingAverageConvergenceDivergence{
		fastMA:         fastMA,
		slowMA:         slowMA,
		signalMA:       signalMA,
		fastDelay:      slowLength - fastLength,
		macdValue:      math.NaN(),
		signalValue:    math.NaN(),
		histogramValue: math.NaN(),
		barFunc:        barFunc,
		quoteFunc:      quoteFunc,
		tradeFunc:      tradeFunc,
		mnemonic:       mnemonic,
	}, nil
}

func newMA(maType MovingAverageType, length int, firstIsAverage bool) (lineUpdater, error) {
	switch maType {
	case SMA:
		return simplemovingaverage.NewSimpleMovingAverage(
			&simplemovingaverage.SimpleMovingAverageParams{Length: length})
	default: // EMA (zero value)
		return exponentialmovingaverage.NewExponentialMovingAverageLength(
			&exponentialmovingaverage.ExponentialMovingAverageLengthParams{
				Length:         length,
				FirstIsAverage: firstIsAverage,
			})
	}
}

func maLabel(maType MovingAverageType) string {
	if maType == SMA {
		return "SMA"
	}

	return "EMA"
}

func buildMnemonic(
	fastLen, slowLen, signalLen int,
	maType, signalMAType MovingAverageType,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) string {
	// Default: macd(12,26,9) when both MA types are EMA.
	// Non-default: macd(12,26,9,SMA,EMA) showing both types.
	suffix := ""
	if maType != EMA || signalMAType != EMA {
		suffix = fmt.Sprintf(",%s,%s", maLabel(maType), maLabel(signalMAType))
	}

	return fmt.Sprintf("macd(%d,%d,%d%s%s)", fastLen, slowLen, signalLen, suffix,
		core.ComponentTripleMnemonic(bc, qc, tc))
}

// IsPrimed indicates whether the indicator is primed.
func (s *MovingAverageConvergenceDivergence) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *MovingAverageConvergenceDivergence) Metadata() core.Metadata {
	desc := "Moving Average Convergence Divergence " + s.mnemonic

	return core.BuildMetadata(
		core.MovingAverageConvergenceDivergence,
		s.mnemonic,
		desc,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " macd", Description: desc + " MACD"},
			{Mnemonic: s.mnemonic + " signal", Description: desc + " Signal"},
			{Mnemonic: s.mnemonic + " histogram", Description: desc + " Histogram"},
		},
	)
}

// Update updates the indicator given the next sample value.
// Returns macd, signal, histogram values.
//
//nolint:nonamedreturns
func (s *MovingAverageConvergenceDivergence) Update(sample float64) (macd, signal, histogram float64) {
	nan := math.NaN()

	if math.IsNaN(sample) {
		return nan, nan, nan
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Feed the slow MA every sample.
	slow := s.slowMA.Update(sample)

	// Delay the fast MA to align SMA seed windows (matches TaLib batch algorithm).
	var fast float64

	if s.fastCount < s.fastDelay {
		s.fastCount++
		fast = nan
	} else {
		fast = s.fastMA.Update(sample)
	}

	if math.IsNaN(fast) || math.IsNaN(slow) {
		s.macdValue = nan
		s.signalValue = nan
		s.histogramValue = nan

		return nan, nan, nan
	}

	macd = fast - slow
	s.macdValue = macd

	signal = s.signalMA.Update(macd)

	if math.IsNaN(signal) {
		s.signalValue = nan
		s.histogramValue = nan

		return macd, nan, nan
	}

	s.signalValue = signal
	histogram = macd - signal
	s.histogramValue = histogram
	s.primed = s.fastMA.IsPrimed() && s.slowMA.IsPrimed() && s.signalMA.IsPrimed()

	return macd, signal, histogram
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *MovingAverageConvergenceDivergence) UpdateScalar(sample *entities.Scalar) core.Output {
	macd, signal, histogram := s.Update(sample.Value)

	const outputCount = 3

	output := make([]any, outputCount)
	output[0] = entities.Scalar{Time: sample.Time, Value: macd}
	output[1] = entities.Scalar{Time: sample.Time, Value: signal}
	output[2] = entities.Scalar{Time: sample.Time, Value: histogram}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *MovingAverageConvergenceDivergence) UpdateBar(sample *entities.Bar) core.Output {
	v := s.barFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *MovingAverageConvergenceDivergence) UpdateQuote(sample *entities.Quote) core.Output {
	v := s.quoteFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *MovingAverageConvergenceDivergence) UpdateTrade(sample *entities.Trade) core.Output {
	v := s.tradeFunc(sample)

	return s.UpdateScalar(&entities.Scalar{Time: sample.Time, Value: v})
}
