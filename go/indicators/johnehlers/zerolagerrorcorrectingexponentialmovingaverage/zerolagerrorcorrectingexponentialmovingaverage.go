package zerolagerrorcorrectingexponentialmovingaverage

//nolint: gofumpt
import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
)

const leastErrorSentinel = 1_000_000.0

// ZeroLagErrorCorrectingExponentialMovingAverage implements Ehlers and Way's
// adaptive zero-lag error-correcting exponential moving average (ZECEMA) from
// the November 2010 TASC article "Zero Lag (Well, Almost)".
//
// Starting from explicit zero EMA and EC states, each accepted finite sample
// updates the baseline EMA once and searches integer-tenths gains i/10 for
// i in [-GainLimit, +GainLimit] against the frozen previous-bar EC using strict
// improvement (`error < leastError`) from the literal 1,000,000 sentinel.
// Both EC (`Value`) and baseline EMA (`Ema`) outputs are available from the
// first accepted finite observation.
//
// Reference:
//
// John F. Ehlers and Ric Way, "Zero Lag (Well, Almost)", TASC, Nov 2010, v28.11, pp. 30-35.
type ZeroLagErrorCorrectingExponentialMovingAverage struct {
	mu              sync.RWMutex
	mnemonic        string
	description     string
	mnemonicEma     string
	descriptionEma  string
	length          int
	gainLimit       int
	alpha           float64
	complement      float64
	count           int
	emaState        float64
	ecState         float64
	bestGainState   float64
	leastErrorState float64
	currentValue    float64
	currentEma      float64
	barFunc         entities.BarFunc
	quoteFunc       entities.QuoteFunc
	tradeFunc       entities.TradeFunc
}

// NewZeroLagErrorCorrectingExponentialMovingAverageDefault creates an instance with Ehlers defaults (Length=20, GainLimit=50).
func NewZeroLagErrorCorrectingExponentialMovingAverageDefault() (*ZeroLagErrorCorrectingExponentialMovingAverage, error) {
	return NewZeroLagErrorCorrectingExponentialMovingAverageLength(DefaultLengthParams())
}

// NewZeroLagErrorCorrectingExponentialMovingAverage returns an instance created from length-based parameters.
func NewZeroLagErrorCorrectingExponentialMovingAverage(params *LengthParams) (*ZeroLagErrorCorrectingExponentialMovingAverage, error) {
	return NewZeroLagErrorCorrectingExponentialMovingAverageLength(params)
}

// NewZeroLagErrorCorrectingExponentialMovingAverageLength returns an instance created from length-based parameters.
func NewZeroLagErrorCorrectingExponentialMovingAverageLength(params *LengthParams) (*ZeroLagErrorCorrectingExponentialMovingAverage, error) {
	const invalid = "invalid zero-lag error-correcting exponential moving average parameters"

	if params == nil {
		return nil, fmt.Errorf("%s: params must not be nil", invalid)
	}
	if params.Length < 1 {
		return nil, fmt.Errorf("%s: length must be an integer >= 1", invalid)
	}

	alpha := 2.0 / (float64(params.Length) + 1.0)
	complement := 1.0 - alpha
	if !(alpha > 0.0 && alpha <= 1.0) || !(complement >= 0.0 && complement < 1.0) {
		return nil, fmt.Errorf("%s: length must yield representably stable EMA coefficients", invalid)
	}

	return newZeroLagErrorCorrectingExponentialMovingAverage(
		params.Length,
		params.GainLimit,
		alpha,
		complement,
		params.Mnemonic(),
		params.EmaMnemonic(),
		params.BarComponent,
		params.QuoteComponent,
		params.TradeComponent,
	)
}

// NewZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactor returns an instance created from smoothing-factor parameters.
func NewZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactor(params *SmoothingFactorParams) (*ZeroLagErrorCorrectingExponentialMovingAverage, error) {
	const invalid = "invalid zero-lag error-correcting exponential moving average parameters"

	if params == nil {
		return nil, fmt.Errorf("%s: params must not be nil", invalid)
	}

	alpha := params.SmoothingFactor
	if math.IsNaN(alpha) || math.IsInf(alpha, 0) || !(alpha > 0.0 && alpha <= 1.0) {
		return nil, fmt.Errorf("%s: smoothing factor must be in (0, 1]", invalid)
	}
	complement := 1.0 - alpha
	if !(complement >= 0.0 && complement < 1.0) {
		return nil, fmt.Errorf("%s: smoothing factor must yield representably stable EMA coefficients", invalid)
	}

	return newZeroLagErrorCorrectingExponentialMovingAverage(
		equivalentLengthFromAlpha(alpha),
		params.GainLimit,
		alpha,
		complement,
		params.Mnemonic(),
		params.EmaMnemonic(),
		params.BarComponent,
		params.QuoteComponent,
		params.TradeComponent,
	)
}

func newZeroLagErrorCorrectingExponentialMovingAverage(
	length int,
	gainLimit int,
	alpha float64,
	complement float64,
	mnemonic string,
	mnemonicEma string,
	barComp entities.BarComponent,
	quoteComp entities.QuoteComponent,
	tradeComp entities.TradeComponent,
) (*ZeroLagErrorCorrectingExponentialMovingAverage, error) {
	const (
		invalid  = "invalid zero-lag error-correcting exponential moving average parameters"
		fmtw     = "%s: %w"
		descr    = "Zero-lag Error-Correcting Exponential Moving Average "
		descrEma = "Zero-lag Error-Correcting Exponential Moving Average baseline EMA "
	)

	if gainLimit < 0 {
		return nil, fmt.Errorf("%s: gain_limit must be a nonnegative integer in tenths", invalid)
	}

	if barComp == 0 {
		barComp = entities.DefaultBarComponent
	}
	if quoteComp == 0 {
		quoteComp = entities.DefaultQuoteComponent
	}
	if tradeComp == 0 {
		tradeComp = entities.DefaultTradeComponent
	}

	barFunc, err := entities.BarComponentFunc(barComp)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}
	quoteFunc, err := entities.QuoteComponentFunc(quoteComp)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}
	tradeFunc, err := entities.TradeComponentFunc(tradeComp)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	instance := &ZeroLagErrorCorrectingExponentialMovingAverage{
		mnemonic:       mnemonic,
		description:    descr + mnemonic,
		mnemonicEma:    mnemonicEma,
		descriptionEma: descrEma + mnemonicEma,
		length:         length,
		gainLimit:      gainLimit,
		alpha:          alpha,
		complement:     complement,
		barFunc:        barFunc,
		quoteFunc:      quoteFunc,
		tradeFunc:      tradeFunc,
	}
	instance.Reset()

	return instance, nil
}

// Reset restores zero startup state while preserving configuration.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.count = 0
	s.emaState = 0.0
	s.ecState = 0.0
	s.bestGainState = 0.0
	s.leastErrorState = 0.0
	s.currentValue = math.NaN()
	s.currentEma = math.NaN()
}

// WarmupBars returns the warmup requirement (0 for this zero-seeded recursive filter).
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) WarmupBars() int {
	return 0
}

// Length returns the configured or equivalent EMA length.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) Length() int {
	return s.length
}

// GainLimit returns the configured integer gain bound in tenths.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) GainLimit() int {
	return s.gainLimit
}

// Alpha returns the exact baseline EMA smoothing factor.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) Alpha() float64 {
	return s.alpha
}

// SamplesSeen returns the count of accepted finite observations since creation or Reset.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) SamplesSeen() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.count
}

// IsPrimed indicates whether the indicator is primed.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.count > 0
}

// Value returns the latest selected EC output (NaN before the first accepted sample).
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) Value() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.currentValue
}

// Ema returns the latest baseline EMA output (NaN before the first accepted sample).
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) Ema() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.currentEma
}

// BestGain returns the latest selected gain in actual gain units (NaN before priming).
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) BestGain() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.count == 0 {
		return math.NaN()
	}
	return s.bestGainState
}

// LeastError returns the latest search least-error state (NaN before priming).
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) LeastError() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.count == 0 {
		return math.NaN()
	}
	return s.leastErrorState
}

// Metadata describes the output data of the indicator.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.ZeroLagErrorCorrectingExponentialMovingAverage,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: s.description},
			{Mnemonic: s.mnemonicEma, Description: s.descriptionEma},
		},
	)
}

func isFinite(number float64) bool {
	return !math.IsNaN(number) && !math.IsInf(number, 0)
}

func (s *ZeroLagErrorCorrectingExponentialMovingAverage) tryUpdateLocked(sample float64) (float64, float64, error) {
	if math.IsInf(sample, 0) {
		return math.NaN(), math.NaN(), errors.New("price must be a finite real or NaN")
	}
	if math.IsNaN(sample) {
		return math.NaN(), math.NaN(), nil
	}

	alpha := s.alpha
	complement := s.complement
	previousEc := s.ecState

	ema := alpha*sample + complement*s.emaState
	if !isFinite(ema) {
		return math.NaN(), math.NaN(), errors.New("baseline EMA arithmetic overflowed")
	}

	leastError := leastErrorSentinel
	bestGain := s.bestGainState
	for index := -s.gainLimit; index <= s.gainLimit; index++ {
		gain := float64(index) / 10.0
		candidate := alpha*(ema+gain*(sample-previousEc)) + complement*previousEc
		errorValue := math.Abs(sample - candidate)
		if !isFinite(candidate) || !isFinite(errorValue) {
			return math.NaN(), math.NaN(), errors.New("gain-search arithmetic overflowed")
		}
		if errorValue < leastError {
			leastError = errorValue
			bestGain = gain
		}
	}

	ec := alpha*(ema+bestGain*(sample-previousEc)) + complement*previousEc
	if !isFinite(ec) {
		return math.NaN(), math.NaN(), errors.New("selected EC arithmetic overflowed")
	}

	s.emaState = ema
	s.ecState = ec
	s.bestGainState = bestGain
	s.leastErrorState = leastError
	s.count++
	s.currentValue = ec
	s.currentEma = ema
	return ec, ema, nil
}

// TryUpdate consumes one scalar sample and returns the EC output, or an error on infinite input or overflow.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) TryUpdate(sample float64) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ec, _, err := s.tryUpdateLocked(sample)
	return ec, err
}

// Update consumes one scalar sample, returning the EC output and panicking on infinite input or overflow.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) Update(sample float64) float64 {
	ec, err := s.TryUpdate(sample)
	if err != nil {
		panic(err)
	}
	return ec
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *ZeroLagErrorCorrectingExponentialMovingAverage) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *ZeroLagErrorCorrectingExponentialMovingAverage) updateEntity(stamp time.Time, sample float64) core.Output {
	const outputLength = 2

	s.mu.Lock()
	defer s.mu.Unlock()

	ec, ema, err := s.tryUpdateLocked(sample)
	if err != nil {
		panic(err)
	}

	output := make(core.Output, outputLength)
	output[0] = entities.Scalar{Time: stamp, Value: ec}
	output[1] = entities.Scalar{Time: stamp, Value: ema}
	return output
}
