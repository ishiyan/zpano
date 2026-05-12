// Package core provides the CandlestickPatterns engine struct and all helper
// methods needed by the pattern implementations in the patterns/ subpackage.
package core

import (
	"zpano/fuzzy"
)

// Minimum history size: 5-candle patterns + 10 default criterion period + 5 margin.
const MinHistory = 20

// Options configures the CandlestickPatterns engine.
type Options struct {
	LongBody        *Criterion
	VeryLongBody    *Criterion
	ShortBody       *Criterion
	DojiBody        *Criterion
	LongShadow      *Criterion
	VeryLongShadow  *Criterion
	ShortShadow     *Criterion
	VeryShortShadow *Criterion
	Near            *Criterion
	Far             *Criterion
	Equal           *Criterion
	FuzzRatio       float64
	Shape           fuzzy.MembershipShape
}

// CandlestickPatterns is the candlestick pattern recognition engine.
//
// Provides streaming bar-by-bar evaluation of 61 Japanese candlestick patterns.
// Call Update(open, high, low, close) for each new bar, then call any pattern
// method to get the result for the current bar.
//
// Pattern methods return:
//
//	+100 for bullish match, -100 for bearish match, 0 for no match.
//	Some patterns return +50/-50 for unconfirmed signals (Hikkake, HikkakeModified).
type CandlestickPatterns struct {
	// Fuzzy configuration.
	FuzzRatio float64
	Shape     fuzzy.MembershipShape

	// Criteria states.
	LongBody        *CriterionState
	VeryLongBody    *CriterionState
	ShortBody       *CriterionState
	DojiBody        *CriterionState
	LongShadow      *CriterionState
	VeryLongShadow  *CriterionState
	ShortShadow     *CriterionState
	VeryShortShadow *CriterionState
	Near            *CriterionState
	Far             *CriterionState
	Equal           *CriterionState
	AllStates       []*CriterionState

	// Ring buffer of recent bars.
	History    []OHLC
	HistSize   int
	HistStart  int
	HistLen    int
	Count      int

	// Stateful pattern state: hikkake_modified
	HikmodPatternResult float64
	HikmodPatternIdx    int
	HikmodConfirmed     bool
	HikmodLastSignal    float64
}

func criterionOrDefault(c *Criterion, def Criterion) Criterion {
	if c != nil {
		return c.Copy()
	}
	return def.Copy()
}

// New creates a new CandlestickPatterns engine with the given options.
// Pass nil for default options.
func New(opts *Options) *CandlestickPatterns {
	var fuzzRatio float64 = 0.2
	var shape fuzzy.MembershipShape = fuzzy.Sigmoid

	var longBody, veryLongBody, shortBody, dojiBody Criterion
	var longShadow, veryLongShadow, shortShadow, veryShortShadow Criterion
	var nearC, farC, equalC Criterion

	if opts != nil {
		if opts.FuzzRatio != 0 {
			fuzzRatio = opts.FuzzRatio
		}
		shape = opts.Shape
		longBody = criterionOrDefault(opts.LongBody, DefaultLongBody)
		veryLongBody = criterionOrDefault(opts.VeryLongBody, DefaultVeryLongBody)
		shortBody = criterionOrDefault(opts.ShortBody, DefaultShortBody)
		dojiBody = criterionOrDefault(opts.DojiBody, DefaultDojiBody)
		longShadow = criterionOrDefault(opts.LongShadow, DefaultLongShadow)
		veryLongShadow = criterionOrDefault(opts.VeryLongShadow, DefaultVeryLongShadow)
		shortShadow = criterionOrDefault(opts.ShortShadow, DefaultShortShadow)
		veryShortShadow = criterionOrDefault(opts.VeryShortShadow, DefaultVeryShortShadow)
		nearC = criterionOrDefault(opts.Near, DefaultNear)
		farC = criterionOrDefault(opts.Far, DefaultFar)
		equalC = criterionOrDefault(opts.Equal, DefaultEqual)
	} else {
		longBody = DefaultLongBody.Copy()
		veryLongBody = DefaultVeryLongBody.Copy()
		shortBody = DefaultShortBody.Copy()
		dojiBody = DefaultDojiBody.Copy()
		longShadow = DefaultLongShadow.Copy()
		veryLongShadow = DefaultVeryLongShadow.Copy()
		shortShadow = DefaultShortShadow.Copy()
		veryShortShadow = DefaultVeryShortShadow.Copy()
		nearC = DefaultNear.Copy()
		farC = DefaultFar.Copy()
		equalC = DefaultEqual.Copy()
	}

	cp := &CandlestickPatterns{
		FuzzRatio:    fuzzRatio,
		Shape:        shape,
		LongBody:     NewCriterionState(longBody, 5),
		VeryLongBody: NewCriterionState(veryLongBody, 5),
		ShortBody:    NewCriterionState(shortBody, 5),
		DojiBody:     NewCriterionState(dojiBody, 5),
		LongShadow:   NewCriterionState(longShadow, 5),
		VeryLongShadow: NewCriterionState(veryLongShadow, 5),
		ShortShadow:    NewCriterionState(shortShadow, 5),
		VeryShortShadow: NewCriterionState(veryShortShadow, 5),
		Near:  NewCriterionState(nearC, 5),
		Far:   NewCriterionState(farC, 5),
		Equal: NewCriterionState(equalC, 5),
	}

	cp.AllStates = []*CriterionState{
		cp.LongBody, cp.VeryLongBody, cp.ShortBody, cp.DojiBody,
		cp.LongShadow, cp.VeryLongShadow, cp.ShortShadow, cp.VeryShortShadow,
		cp.Near, cp.Far, cp.Equal,
	}

	// History size: largest criterion period + 10, floored at MinHistory.
	maxPeriod := 0
	for _, s := range cp.AllStates {
		if s.Criterion.AveragePeriod > maxPeriod {
			maxPeriod = s.Criterion.AveragePeriod
		}
	}
	historySize := maxPeriod + 10
	if historySize < MinHistory {
		historySize = MinHistory
	}
	cp.History = make([]OHLC, historySize)
	cp.HistSize = historySize

	return cp
}

// UpdateBar feeds a new OHLC bar into the engine (ring buffer + criterion states).
// This does NOT call HikkakeModifiedUpdate — the caller must do that.
func (cp *CandlestickPatterns) UpdateBar(o, h, l, c float64) {
	bar := OHLC{o, h, l, c}
	if cp.HistLen == cp.HistSize {
		cp.History[cp.HistStart] = bar
		cp.HistStart = (cp.HistStart + 1) % cp.HistSize
	} else {
		idx := (cp.HistStart + cp.HistLen) % cp.HistSize
		cp.History[idx] = bar
		cp.HistLen++
	}
	for _, s := range cp.AllStates {
		s.Push(o, h, l, c)
	}
	cp.Count++
}

// Bar gets OHLC of a bar. shift=1 is most recent, shift=2 is one before, etc.
func (cp *CandlestickPatterns) Bar(shift int) OHLC {
	idx := (cp.HistStart + cp.HistLen - shift) % cp.HistSize
	return cp.History[idx]
}

// Has checks if we have at least n bars in history.
func (cp *CandlestickPatterns) Has(n int) bool {
	return cp.HistLen >= n
}

// Enough checks if we have sufficient bars for a pattern requiring nCandles
// plus the maximum average_period of the given criteria.
func (cp *CandlestickPatterns) Enough(nCandles int, criteria ...*CriterionState) bool {
	avail := cp.HistLen - nCandles
	for _, cs := range criteria {
		if avail < cs.Criterion.AveragePeriod {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Criterion average helpers (shift is from the end, 1-based)
// ---------------------------------------------------------------------------

// AvgCS gets the criterion average value at a given shift from the most recent bar.
func (cp *CandlestickPatterns) AvgCS(cs *CriterionState, shift int) float64 {
	b := cp.Bar(shift)
	return cs.Avg(shift, b.O, b.H, b.L, b.C)
}

// ---------------------------------------------------------------------------
// Fuzzy membership helpers
// ---------------------------------------------------------------------------

// MuLess computes fuzzy 'value < avg' membership degree.
func (cp *CandlestickPatterns) MuLess(value float64, cs *CriterionState, shift int) float64 {
	avg := cp.AvgCS(cs, shift)
	w := cp.FuzzRatio * avg
	if avg <= 0.0 {
		w = 0.0
	}
	return fuzzy.MuLess(value, avg, w, cp.Shape)
}

// MuGreater computes fuzzy 'value > avg' membership degree.
func (cp *CandlestickPatterns) MuGreater(value float64, cs *CriterionState, shift int) float64 {
	avg := cp.AvgCS(cs, shift)
	w := cp.FuzzRatio * avg
	if avg <= 0.0 {
		w = 0.0
	}
	return fuzzy.MuGreater(value, avg, w, cp.Shape)
}

// MuNearValue computes fuzzy 'value ≈ target ± avg' membership degree.
func (cp *CandlestickPatterns) MuNearValue(value, target float64, cs *CriterionState, shift int) float64 {
	avg := cp.AvgCS(cs, shift)
	w := cp.FuzzRatio * avg
	if avg <= 0.0 {
		w = 0.0
	}
	return fuzzy.MuNear(value, target, w, cp.Shape)
}

// MuGeRaw computes fuzzy 'value >= threshold' with explicit width (no criterion).
func (cp *CandlestickPatterns) MuGeRaw(value, threshold, width float64) float64 {
	return fuzzy.MuGreaterEqual(value, threshold, width, cp.Shape)
}

// MuGtRaw computes fuzzy 'value > threshold' with explicit width (no criterion).
func (cp *CandlestickPatterns) MuGtRaw(value, threshold, width float64) float64 {
	return fuzzy.MuGreater(value, threshold, width, cp.Shape)
}

// MuLtRaw computes fuzzy 'value < threshold' with explicit width (no criterion).
func (cp *CandlestickPatterns) MuLtRaw(value, threshold, width float64) float64 {
	return fuzzy.MuLess(value, threshold, width, cp.Shape)
}

// MuBullish computes fuzzy degree of bullishness ∈ [0, 1].
func (cp *CandlestickPatterns) MuBullish(o, c float64, shift int) float64 {
	d := cp.MuDirectionRaw(o, c, shift)
	if d > 0.0 {
		return d
	}
	return 0.0
}

// MuBearish computes fuzzy degree of bearishness ∈ [0, 1].
func (cp *CandlestickPatterns) MuBearish(o, c float64, shift int) float64 {
	d := cp.MuDirectionRaw(o, c, shift)
	if -d > 0.0 {
		return -d
	}
	return 0.0
}

// MuDirectionRaw computes raw fuzzy direction ∈ [-1, +1].
func (cp *CandlestickPatterns) MuDirectionRaw(o, c float64, shift int) float64 {
	avg := cp.AvgCS(cp.ShortBody, shift)
	return fuzzy.MuDirection(o, c, avg, 2.0)
}

// HikkakeModifiedUpdate is called from Update() to track stateful hikkake_modified pattern.
func (cp *CandlestickPatterns) HikkakeModifiedUpdate() {
	if cp.Count < 4 {
		return
	}

	b1 := cp.Bar(4)
	b2 := cp.Bar(3)
	b3 := cp.Bar(2)
	b4 := cp.Bar(1)

	// Check for new pattern.
	if b2.H < b1.H && b2.L > b1.L &&
		b3.H < b2.H && b3.L > b2.L {
		nearAvg := cp.AvgCS(cp.Near, 3)
		// Bullish: 4th breaks low, 2nd close near its low.
		if b4.H < b3.H && b4.L < b3.L &&
			b2.C <= b2.L+nearAvg {
			cp.HikmodPatternResult = 100.0
			cp.HikmodPatternIdx = cp.Count
			return
		}
		// Bearish: 4th breaks high, 2nd close near its high.
		if b4.H > b3.H && b4.L > b3.L &&
			b2.C >= b2.H-nearAvg {
			cp.HikmodPatternResult = -100.0
			cp.HikmodPatternIdx = cp.Count
			return
		}
	}

	// No new pattern — check for confirmation.
	if cp.HikmodPatternResult != 0 &&
		cp.Count <= cp.HikmodPatternIdx+3 {
		shift3rd := cp.Count - cp.HikmodPatternIdx + 2
		b3rd := cp.Bar(shift3rd)

		if cp.HikmodPatternResult > 0 && b4.C > b3rd.H {
			cp.HikmodLastSignal = 200.0
			cp.HikmodPatternResult = 0.0
			cp.HikmodPatternIdx = 0
			cp.HikmodConfirmed = true
			return
		}
		if cp.HikmodPatternResult < 0 && b4.C < b3rd.L {
			cp.HikmodLastSignal = -200.0
			cp.HikmodPatternResult = 0.0
			cp.HikmodPatternIdx = 0
			cp.HikmodConfirmed = true
			return
		}
	}

	// If we passed the 3-bar window, reset.
	if cp.HikmodPatternResult != 0 &&
		cp.Count > cp.HikmodPatternIdx+3 {
		cp.HikmodPatternResult = 0.0
		cp.HikmodPatternIdx = 0
	}
}
