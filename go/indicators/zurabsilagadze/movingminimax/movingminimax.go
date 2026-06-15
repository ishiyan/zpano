package movingminimax

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// MovingMiniMax is Zurab Silagadze's Moving Mini-Max (MMM) indicator.
//
// A nonlinear indicator for technical analysis that emphasizes local maximums and minimums
// in a price series with inherent smoothing. The algorithm is borrowed from gamma-ray
// spectroscopy peak finding and models price exploration as a quantum particle that can
// tunnel through small noise barriers but is stopped by genuine trend reversals.
//
// Reference:
//
// Silagadze, Z. K. (2011). Moving Mini-Max -- a new indicator for technical analysis.
// IFTA Journal 11, 46-49. arXiv:0802.0984v2.
type MovingMiniMax struct {
	mu sync.RWMutex

	mnemonic    string
	description string

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc

	m          int
	n          int
	numExtrema int

	window []float64
	bufPos int
	count  int

	primed bool
}

// peak holds one detected peak as a (strength, index) pair.
type peak struct {
	strength float64
	index    int
}

// level holds one detected support/resistance level.
type level struct {
	price    float64
	offset   int
	strength float64
}

// result holds one computed MMM output set.
type result struct {
	up          float64
	down        float64
	resistances []level
	supports    []level
	upDist      []float64
	downDist    []float64
	valid       bool
}

// NewMovingMiniMax returns an instance of the indicator created using supplied parameters.
func NewMovingMiniMax(p *Params) (*MovingMiniMax, error) {
	const (
		invalid           = "invalid moving mini-max parameters"
		fmts              = "%s: %s"
		fmtw              = "%s: %w"
		defaultM          = 5
		defaultN          = 50
		defaultNumExtrema = 3
	)

	m := p.M
	if m == 0 {
		m = defaultM
	}

	n := p.N
	if n == 0 {
		n = defaultN
	}

	numExtrema := p.NumExtrema
	if numExtrema == 0 {
		numExtrema = defaultNumExtrema
	}

	if m < 1 {
		return nil, fmt.Errorf(fmts, invalid, "m should be >= 1")
	}

	if n <= 2*m {
		return nil, fmt.Errorf(fmts, invalid, "n should be > 2*m")
	}

	if numExtrema < 1 {
		return nil, fmt.Errorf(fmts, invalid, "num extrema should be >= 1")
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

	mnemonic := fmt.Sprintf("mmm(%d,%d,%d%s)",
		m, n, numExtrema, core.ComponentTripleMnemonic(bc, qc, tc))

	return &MovingMiniMax{
		mnemonic:    mnemonic,
		description: "Moving mini-max " + mnemonic,
		barFunc:     barFunc,
		quoteFunc:   quoteFunc,
		tradeFunc:   tradeFunc,
		m:           m,
		n:           n,
		numExtrema:  numExtrema,
		window:      make([]float64, n),
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *MovingMiniMax) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *MovingMiniMax) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.MovingMiniMax,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonic + " up", Description: s.description + " up value"},
			{Mnemonic: s.mnemonic + " down", Description: s.description + " down value"},
			{Mnemonic: s.mnemonic + " resistances", Description: s.description + " resistances"},
			{Mnemonic: s.mnemonic + " supports", Description: s.description + " supports"},
			{Mnemonic: s.mnemonic + " up dist", Description: s.description + " up distribution"},
			{Mnemonic: s.mnemonic + " down dist", Description: s.description + " down distribution"},
		},
	)
}

// calcQValues computes Q_{i,i+1} and Q_{i,i-1} for each position i = 0..n-1.
func calcQValues(window []float64, n, m int, negate bool) ([]float64, []float64) {
	sign := 1.0
	if negate {
		sign = -1.0
	}

	qPlus := make([]float64, n)
	qMinus := make([]float64, n)

	for i := 0; i < n; i++ {
		si := window[i]
		sumPlus := 0.0
		sumMinus := 0.0

		for k := 1; k <= m; k++ {
			sForward := window[n-1]
			if i+k < n {
				sForward = window[i+k]
			}

			sBackward := window[0]
			if i-k >= 0 {
				sBackward = window[i-k]
			}

			argPlus := 0.0
			if denomPlus := sForward + si; denomPlus != 0.0 {
				argPlus = sign * 2.0 * (sForward - si) / denomPlus
			}

			argMinus := 0.0
			if denomMinus := sBackward + si; denomMinus != 0.0 {
				argMinus = sign * 2.0 * (sBackward - si) / denomMinus
			}

			sumPlus += math.Exp(argPlus)
			sumMinus += math.Exp(argMinus)
		}

		qPlus[i] = sumPlus
		qMinus[i] = sumMinus
	}

	return qPlus, qMinus
}

// calcPValues computes transition probabilities P_{i,i+1} and P_{i,i-1} from Q-values.
func calcPValues(qPlus, qMinus []float64, n int) ([]float64, []float64) {
	pPlus := make([]float64, n)
	pMinus := make([]float64, n)

	for i := 0; i < n; i++ {
		denom := qPlus[i] + qMinus[i]
		if denom == 0.0 {
			pPlus[i] = 0.5
			pMinus[i] = 0.5
		} else {
			pPlus[i] = qPlus[i] / denom
			pMinus[i] = qMinus[i] / denom
		}
	}

	return pPlus, pMinus
}

// calcMiniMax computes the normalized mini-max series from transition probabilities.
func calcMiniMax(pPlus, pMinus []float64, n int) []float64 {
	u := make([]float64, n)
	u[0] = 1.0

	for i := 1; i < n; i++ {
		pPrevToI := pPlus[i-1]
		pIToPrev := pMinus[i]

		if pIToPrev == 0.0 {
			u[i] = u[i-1] * 1e10
		} else {
			u[i] = (pPrevToI / pIToPrev) * u[i-1]
		}
	}

	total := 0.0
	for i := 0; i < n; i++ {
		total += u[i]
	}

	minimax := make([]float64, n)
	if total == 0.0 {
		for i := 0; i < n; i++ {
			minimax[i] = 1.0 / float64(n)
		}

		return minimax
	}

	for i := 0; i < n; i++ {
		minimax[i] = u[i] / total
	}

	return minimax
}

// findPeaks finds distinct local peaks, returned sorted by strength descending.
func findPeaks(values []float64, numPeaks, minSeparation int) []peak {
	n := len(values)

	candidates := make([]peak, 0, n)

	for i := 0; i < n; i++ {
		var isPeak bool

		switch {
		case i == 0:
			isPeak = n <= 1 || values[i] >= values[i+1]
		case i == n-1:
			isPeak = values[i] >= values[i-1]
		default:
			isPeak = values[i] >= values[i-1] && values[i] >= values[i+1]
		}

		if isPeak {
			candidates = append(candidates, peak{strength: values[i], index: i})
		}
	}

	// Sort by strength descending; the reference sorts (value, index) tuples in
	// reverse, so ties break on the larger index first.
	sort.SliceStable(candidates, func(a, b int) bool {
		if candidates[a].strength != candidates[b].strength {
			return candidates[a].strength > candidates[b].strength
		}

		return candidates[a].index > candidates[b].index
	})

	selected := make([]peak, 0, numPeaks)

	for _, c := range candidates {
		if len(selected) >= numPeaks {
			break
		}

		tooClose := false

		for _, sel := range selected {
			if abs(c.index-sel.index) < minSeparation {
				tooClose = true

				break
			}
		}

		if !tooClose {
			selected = append(selected, c)
		}
	}

	return selected
}

// abs returns the absolute value of an int.
func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

// update processes one price and returns the computed result.
func (s *MovingMiniMax) update(sample float64) result {
	// Store the sample in the ring buffer.
	if s.count < s.n {
		s.window[s.count] = sample
		s.count++
	} else {
		s.window[s.bufPos] = sample
		s.bufPos = (s.bufPos + 1) % s.n
	}

	if s.count < s.n {
		s.primed = false

		return result{}
	}

	s.primed = true

	n := s.n
	m := s.m

	// Reconstruct the window in chronological order (oldest -> newest).
	window := make([]float64, n)
	for i := 0; i < n; i++ {
		window[i] = s.window[(s.bufPos+i)%n]
	}

	qUpPlus, qUpMinus := calcQValues(window, n, m, false)
	qDnPlus, qDnMinus := calcQValues(window, n, m, true)

	pUpPlus, pUpMinus := calcPValues(qUpPlus, qUpMinus, n)
	pDnPlus, pDnMinus := calcPValues(qDnPlus, qDnMinus, n)

	upDist := calcMiniMax(pUpPlus, pUpMinus, n)
	dnDist := calcMiniMax(pDnPlus, pDnMinus, n)

	minSep := m
	if minSep < 2 {
		minSep = 2
	}

	uPeaks := findPeaks(upDist, s.numExtrema, minSep)
	dPeaks := findPeaks(dnDist, s.numExtrema, minSep)

	resistances := make([]level, len(uPeaks))
	for i, pk := range uPeaks {
		resistances[i] = level{price: window[pk.index], offset: (n - 1) - pk.index, strength: pk.strength}
	}

	supports := make([]level, len(dPeaks))
	for i, pk := range dPeaks {
		supports[i] = level{price: window[pk.index], offset: (n - 1) - pk.index, strength: pk.strength}
	}

	return result{
		up:          upDist[n-1],
		down:        dnDist[n-1],
		resistances: resistances,
		supports:    supports,
		upDist:      upDist,
		downDist:    dnDist,
		valid:       true,
	}
}

// wrap builds the output slice for the given time and result.
func (s *MovingMiniMax) wrap(t time.Time, r result) core.Output {
	nan := math.NaN()

	up := nan
	down := nan

	if r.valid {
		up = r.up
		down = r.down
	}

	output := make([]any, 6)
	output[0] = entities.Scalar{Time: t, Value: up}
	output[1] = entities.Scalar{Time: t, Value: down}
	output[2] = levelsOf(t, r.resistances)
	output[3] = levelsOf(t, r.supports)
	output[4] = polylineOf(t, r.upDist)
	output[5] = polylineOf(t, r.downDist)

	return output
}

// levelsOf builds a *outputs.Levels from a slice of levels.
func levelsOf(t time.Time, levels []level) *outputs.Levels {
	if len(levels) == 0 {
		return outputs.NewEmptyLevels(t)
	}

	entries := make([]outputs.Level, len(levels))
	for i, lv := range levels {
		entries[i] = outputs.NewLevel(lv.price, lv.offset, lv.strength)
	}

	return outputs.NewLevels(t, entries)
}

// polylineOf builds a *outputs.Polyline from a distribution slice (offset 0 = oldest).
func polylineOf(t time.Time, values []float64) *outputs.Polyline {
	if len(values) == 0 {
		return outputs.NewEmptyPolyline(t)
	}

	points := make([]outputs.Point, len(values))
	for i, v := range values {
		points[i] = outputs.Point{Offset: i, Value: v}
	}

	return outputs.NewPolyline(t, points)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *MovingMiniMax) UpdateScalar(sample *entities.Scalar) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.wrap(sample.Time, s.update(sample.Value))
}

// UpdateBar updates the indicator given the next bar sample.
func (s *MovingMiniMax) UpdateBar(sample *entities.Bar) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.wrap(sample.Time, s.update(s.barFunc(sample)))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *MovingMiniMax) UpdateQuote(sample *entities.Quote) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.wrap(sample.Time, s.update(s.quoteFunc(sample)))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *MovingMiniMax) UpdateTrade(sample *entities.Trade) core.Output {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.wrap(sample.Time, s.update(s.tradeFunc(sample)))
}
