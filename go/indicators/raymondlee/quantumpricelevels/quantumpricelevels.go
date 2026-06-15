package quantumpricelevels

import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// cbrt returns the signed real cube root via Pow (matches the reference implementation).
func cbrt(x float64) float64 {
	if x >= 0.0 {
		return math.Pow(x, 1.0/3.0)
	}

	return -math.Pow(-x, 1.0/3.0)
}

// computeK0 returns the K0 constant for energy level n (Dasgupta et al. 2007).
func computeK0(n int) float64 {
	fn := float64(n)
	numerator := 1.1924 + 33.2383*fn + 56.2169*fn*fn
	denominator := 1.0 + 43.6106*fn

	return math.Pow(numerator/denominator, 1.0/3.0)
}

// QuantumPriceLevels is Raymond Lee's Quantum Price Levels (QPL) indicator.
//
// It computes discrete support/resistance price levels from a quantum-finance analogy:
// the market is modelled as a quantum anharmonic oscillator, and the discrete energy
// eigenvalues of the system map to price levels above and below the current price.
//
// Reference:
//
// Lee, R. S. T. (2021). Quantum Finance Forecast System with Quantum Anharmonic
// Oscillator Model for Quantum Price Level Modeling. IAJER, 4(02), 1-21.
type QuantumPriceLevels struct {
	mu sync.RWMutex

	mnemonic    string
	description string

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	lookback    int
	numLevels   int
	numBins     int
	scaleFactor float64

	k []float64 // Pre-computed K0 constants.

	returns   []float64
	bufPos    int
	count     int
	prevPrice float64
	havePrev  bool

	primed bool
}

// result holds one computed QPL output set.
type result struct {
	lambda      float64
	sigma       float64
	nqpr        []float64
	resistances []float64
	supports    []float64
	valid       bool
}

// NewQuantumPriceLevels returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen
func NewQuantumPriceLevels(p *Params) (*QuantumPriceLevels, error) {
	const (
		invalid          = "invalid quantum price levels parameters"
		fmts             = "%s: %s"
		fmtw             = "%s: %w"
		defaultLookback  = 2048
		defaultNumLevels = 21
		defaultNumBins   = 100
		defaultScale     = 0.21
	)

	lookback := p.Lookback
	if lookback == 0 {
		lookback = defaultLookback
	}

	numLevels := p.NumLevels
	if numLevels == 0 {
		numLevels = defaultNumLevels
	}

	numBins := p.NumBins
	if numBins == 0 {
		numBins = defaultNumBins
	}

	scaleFactor := p.ScaleFactor
	if scaleFactor == 0.0 {
		scaleFactor = defaultScale
	}

	if lookback < 2 {
		return nil, fmt.Errorf(fmts, invalid, "lookback should be >= 2")
	}

	if numLevels < 1 {
		return nil, fmt.Errorf(fmts, invalid, "num levels should be >= 1")
	}

	if numBins < 2 {
		return nil, fmt.Errorf(fmts, invalid, "num bins should be >= 2")
	}

	if scaleFactor <= 0.0 {
		return nil, fmt.Errorf(fmts, invalid, "scale factor should be > 0")
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

	k := make([]float64, numLevels)
	for n := 0; n < numLevels; n++ {
		k[n] = computeK0(n)
	}

	mnemonic := fmt.Sprintf("qpl(%d,%d,%d,%s%s)",
		lookback, numLevels, numBins, formatScale(scaleFactor), core.ComponentTripleMnemonic(bc, qc, tc))

	return &QuantumPriceLevels{
		mnemonic:    mnemonic,
		description: "Quantum price levels " + mnemonic,
		barFunc:     barFunc,
		quoteFunc:   quoteFunc,
		tradeFunc:   tradeFunc,
		lookback:    lookback,
		numLevels:   numLevels,
		numBins:     numBins,
		scaleFactor: scaleFactor,
		k:           k,
		returns:     make([]float64, lookback),
	}, nil
}

// formatScale formats the scale factor compactly (e.g. 0.21, 0.1, 0.42).
func formatScale(v float64) string {
	return fmt.Sprintf("%g", v)
}

// IsPrimed indicates whether the indicator is primed.
func (s *QuantumPriceLevels) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *QuantumPriceLevels) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.QuantumPriceLevels,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " lambda", Description: s.description + " anharmonic coefficient"},
			{Mnemonic: s.mnemonic + " stddev", Description: s.description + " return standard deviation"},
			{Mnemonic: s.mnemonic + " nqpr", Description: s.description + " normalized multipliers"},
			{Mnemonic: s.mnemonic + " resistances", Description: s.description + " resistance levels"},
			{Mnemonic: s.mnemonic + " supports", Description: s.description + " support levels"},
		},
	)
}

// update processes one price and returns the computed result.
//
//nolint:funlen,gocognit,cyclop
func (s *QuantumPriceLevels) update(sample float64) result {
	// First price: just store it; no return yet.
	if !s.havePrev {
		s.prevPrice = sample
		s.havePrev = true
		s.primed = false

		return result{}
	}

	// Inverse return ratio (Lee's convention).
	newReturn := 1.0
	if sample > 0.0 {
		newReturn = s.prevPrice / sample
	}

	s.prevPrice = sample

	// Store in ring buffer.
	if s.count < s.lookback {
		s.returns[s.count] = newReturn
		s.count++
	} else {
		s.returns[s.bufPos] = newReturn
		s.bufPos = (s.bufPos + 1) % s.lookback
	}

	if s.count < s.lookback {
		s.primed = false

		return result{}
	}

	s.primed = true

	lookback := s.lookback
	numBins := s.numBins
	numLevels := s.numLevels
	scaleFactor := s.scaleFactor

	// Statistics (population mu, sigma).
	sumR := 0.0
	for i := 0; i < lookback; i++ {
		sumR += s.returns[i]
	}

	mu := sumR / float64(lookback)

	sumVar := 0.0
	for i := 0; i < lookback; i++ {
		diff := s.returns[i] - mu
		sumVar += diff * diff
	}

	sigma := math.Sqrt(sumVar / float64(lookback))
	if sigma == 0.0 {
		return result{}
	}

	// Histogram centred at r = 1.
	halfBins := numBins / 2
	dr := 3.0 * sigma / float64(halfBins)
	leftBoundary := 1.0 - float64(halfBins)*dr

	q := make([]int, numBins)
	totalCount := 0

	for i := 0; i < lookback; i++ {
		r := s.returns[i]
		binIndex := int((r - leftBoundary) / dr)

		if binIndex >= 0 && binIndex < numBins {
			q[binIndex]++
			totalCount++
		}
	}

	if totalCount == 0 {
		return result{}
	}

	// Ground state (peak bin).
	maxQ := 0.0
	maxQno := 0

	for kk := 0; kk < numBins; kk++ {
		nq := float64(q[kk]) / float64(totalCount)
		if nq > maxQ {
			maxQ = nq
			maxQno = kk
		}
	}

	if maxQno == 0 || maxQno == numBins-1 {
		return result{}
	}

	// lambda via FDM.
	phiPlus1 := float64(q[maxQno+1]) / float64(totalCount)
	phiMinus1 := float64(q[maxQno-1]) / float64(totalCount)

	rPeak := leftBoundary + float64(maxQno)*dr
	r0 := rPeak - dr/2.0
	rPlus1 := r0 + dr
	rMinus1 := r0 - dr

	lUp := (rMinus1*rMinus1)*phiMinus1 - (rPlus1*rPlus1)*phiPlus1
	lDw := (rPlus1*rPlus1*rPlus1*rPlus1)*phiPlus1 - (rMinus1*rMinus1*rMinus1*rMinus1)*phiMinus1

	if lDw == 0.0 {
		return result{}
	}

	lambda := math.Abs(lUp / lDw)

	// Energy levels via Cardano.
	qfel := make([]float64, numLevels)

	for n := 0; n < numLevels; n++ {
		twoNPlus1 := float64(2*n + 1)
		p := -(twoNPlus1 * twoNPlus1)
		qCoef := -lambda * (twoNPlus1 * twoNPlus1 * twoNPlus1) * (s.k[n] * s.k[n] * s.k[n])
		discriminant := (qCoef*qCoef)/4.0 + (p*p*p)/27.0

		if discriminant < 0.0 {
			return result{}
		}

		sqrtD := math.Sqrt(discriminant)
		u := cbrt(-qCoef/2.0 + sqrtD)
		v := cbrt(-qCoef/2.0 - sqrtD)
		qfel[n] = u + v
	}

	if qfel[0] == 0.0 {
		return result{}
	}

	// NQPR and projection from the current price.
	nqpr := make([]float64, numLevels)
	resistances := make([]float64, numLevels)
	supports := make([]float64, numLevels)

	for n := 0; n < numLevels; n++ {
		qpr := qfel[n] / qfel[0]
		nqpr[n] = 1.0 + scaleFactor*sigma*qpr
		resistances[n] = sample * nqpr[n]
		supports[n] = sample / nqpr[n]
	}

	return result{
		lambda:      lambda,
		sigma:       sigma,
		nqpr:        nqpr,
		resistances: resistances,
		supports:    supports,
		valid:       true,
	}
}

// wrap builds the output slice for the given time and result.
func (s *QuantumPriceLevels) wrap(t time.Time, r result) core.Output {
	nan := math.NaN()

	lambda := nan
	sigma := nan

	if r.valid {
		lambda = r.lambda
		sigma = r.sigma
	}

	output := make([]any, 5)
	output[0] = entities.Scalar{Time: t, Value: lambda}
	output[1] = entities.Scalar{Time: t, Value: sigma}
	output[2] = levelsOf(t, r.nqpr)
	output[3] = levelsOf(t, r.resistances)
	output[4] = levelsOf(t, r.supports)

	return output
}

// levelsOf builds a *outputs.Levels from a slice of values (offset 0, NaN strength).
func levelsOf(t time.Time, values []float64) *outputs.Levels {
	if len(values) == 0 {
		return outputs.NewEmptyLevels(t)
	}

	entries := make([]outputs.Level, len(values))
	for i, v := range values {
		entries[i] = outputs.NewValueLevel(v)
	}

	return outputs.NewLevels(t, entries)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *QuantumPriceLevels) UpdateScalar(sample *entities.Scalar) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.wrap(sample.Time, s.update(sample.Value))
}

// UpdateBar updates the indicator given the next bar sample.
func (s *QuantumPriceLevels) UpdateBar(sample *entities.Bar) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.wrap(sample.Time, s.update(s.barFunc(sample)))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *QuantumPriceLevels) UpdateQuote(sample *entities.Quote) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.wrap(sample.Time, s.update(s.quoteFunc(sample)))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *QuantumPriceLevels) UpdateTrade(sample *entities.Trade) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.wrap(sample.Time, s.update(s.tradeFunc(sample)))
}
