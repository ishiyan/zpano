package core

import (
	"math"

	"zpano/streamingkbn"
)

// PartialMoments computes streaming lower/higher partial moments about a
// threshold (target return or minimum acceptable return, MAR), together
// with the moments of the returns strictly above (upper excess) and
// strictly below (lower excess) the threshold.
//
// All internal moment accumulators use ddof=1, bias=true, fisher=true,
// matching scipy's default behavior for kurtosis.
//
// Use [NewPartialMoments] to create an instance.
type PartialMoments struct {
	threshold  float64
	countTotal int

	upperExcessKBN *streamingkbn.RawMomentsKleinKBN
	lowerExcessKBN *streamingkbn.RawMomentsKleinKBN
	lpmKBN         *streamingkbn.RawMomentsKleinKBN
	hpmKBN         *streamingkbn.RawMomentsKleinKBN
}

// NewPartialMoments returns a new empty PartialMoments for the given
// threshold, the target return or minimum acceptable return (MAR) in the
// same periodicity as the returns.
func NewPartialMoments(threshold float64) *PartialMoments {
	return &PartialMoments{
		threshold:      threshold,
		upperExcessKBN: streamingkbn.NewRawMomentsKleinKBN(1, true, true),
		lowerExcessKBN: streamingkbn.NewRawMomentsKleinKBN(1, true, true),
		lpmKBN:         streamingkbn.NewRawMomentsKleinKBN(1, true, true),
		hpmKBN:         streamingkbn.NewRawMomentsKleinKBN(1, true, true),
	}
}

// Threshold returns the threshold (target return or MAR).
func (p *PartialMoments) Threshold() float64 { return p.threshold }

// SetThreshold changes the threshold. It does not recompute accumulated
// state, so it should only be called on an empty accumulator.
func (p *PartialMoments) SetThreshold(threshold float64) { p.threshold = threshold }

// Reset clears all accumulated state (the threshold is kept).
func (p *PartialMoments) Reset() {
	p.countTotal = 0
	p.upperExcessKBN.Reset()
	p.lowerExcessKBN.Reset()
	p.lpmKBN.Reset()
	p.hpmKBN.Reset()
}

// Revert removes a previously added return.
func (p *PartialMoments) Revert(ret float64) {
	p.countTotal--
	// Lower partial moments for the raw returns less target return.
	pm := p.threshold - ret
	if pm < 0 {
		p.upperExcessKBN.Revert(-pm)
		pm = 0
	}
	p.lpmKBN.Revert(pm)

	// Higher partial moments for the raw returns less required return.
	pm = ret - p.threshold
	if pm < 0 {
		p.lowerExcessKBN.Revert(-pm)
		pm = 0
	}
	p.hpmKBN.Revert(pm)
}

// Update adds a return.
func (p *PartialMoments) Update(ret float64) {
	p.countTotal++
	// Lower partial moments for the raw returns less target return.
	pm := p.threshold - ret
	if pm < 0 {
		p.upperExcessKBN.Update(-pm)
		pm = 0
	}
	p.lpmKBN.Update(pm)

	// Higher partial moments for the raw returns less required return.
	pm = ret - p.threshold
	if pm < 0 {
		p.lowerExcessKBN.Update(-pm)
		pm = 0
	}
	p.hpmKBN.Update(pm)
}

// LowerPartialMoment1 returns mean(max(threshold - r, 0)).
func (p *PartialMoments) LowerPartialMoment1() float64 { return p.lpmKBN.X1() }

// LowerPartialMoment2 returns mean(max(threshold - r, 0)²).
func (p *PartialMoments) LowerPartialMoment2() float64 { return p.lpmKBN.X2() }

// LowerPartialMoment3 returns mean(max(threshold - r, 0)³).
func (p *PartialMoments) LowerPartialMoment3() float64 { return p.lpmKBN.X3() }

// LowerPartialMoment4 returns mean(max(threshold - r, 0)⁴).
func (p *PartialMoments) LowerPartialMoment4() float64 { return p.lpmKBN.X4() }

// HigherPartialMoment1 returns mean(max(r - threshold, 0)).
func (p *PartialMoments) HigherPartialMoment1() float64 { return p.hpmKBN.X1() }

// HigherPartialMoment2 returns mean(max(r - threshold, 0)²).
func (p *PartialMoments) HigherPartialMoment2() float64 { return p.hpmKBN.X2() }

// HigherPartialMoment3 returns mean(max(r - threshold, 0)³).
func (p *PartialMoments) HigherPartialMoment3() float64 { return p.hpmKBN.X3() }

// HigherPartialMoment4 returns mean(max(r - threshold, 0)⁴).
func (p *PartialMoments) HigherPartialMoment4() float64 { return p.hpmKBN.X4() }

// DownsideFrequency returns the proportion of returns below the threshold
// (NaN when empty).
func (p *PartialMoments) DownsideFrequency() float64 {
	total := p.countTotal
	if total == 0 {
		return math.NaN()
	}
	return float64(p.lowerExcessKBN.N()) / float64(total)
}

// UpsideFrequency returns the proportion of returns above the threshold
// (NaN when empty).
func (p *PartialMoments) UpsideFrequency() float64 {
	total := p.countTotal
	if total == 0 {
		return math.NaN()
	}
	return float64(p.upperExcessKBN.N()) / float64(total)
}

// DownsidePotential returns the mean of the lower partial moments (also
// called shortfall).
func (p *PartialMoments) DownsidePotential() float64 { return p.lpmKBN.Mean() }

// TotalCount returns the number of accumulated returns.
func (p *PartialMoments) TotalCount() int { return p.countTotal }

// UpperExcessCount returns the number of returns above the threshold.
func (p *PartialMoments) UpperExcessCount() int { return p.upperExcessKBN.N() }

// LowerExcessCount returns the number of returns below the threshold.
func (p *PartialMoments) LowerExcessCount() int { return p.lowerExcessKBN.N() }

// UpperExcessMoment1 returns the mean of (r - threshold) over returns
// above the threshold.
func (p *PartialMoments) UpperExcessMoment1() float64 { return p.upperExcessKBN.X1() }

// UpperExcessMoment1Sum returns the sum of (r - threshold) over returns
// above the threshold.
func (p *PartialMoments) UpperExcessMoment1Sum() float64 { return p.upperExcessKBN.X1Sum() }

// UpperExcessMoment2 returns the mean of (r - threshold)² over returns
// above the threshold.
func (p *PartialMoments) UpperExcessMoment2() float64 { return p.upperExcessKBN.X2() }

// UpperExcessMoment2Sum returns the sum of (r - threshold)² over returns
// above the threshold.
func (p *PartialMoments) UpperExcessMoment2Sum() float64 { return p.upperExcessKBN.X2Sum() }

// UpperExcessMoment3 returns the mean of (r - threshold)³ over returns
// above the threshold.
func (p *PartialMoments) UpperExcessMoment3() float64 { return p.upperExcessKBN.X3() }

// UpperExcessMoment4 returns the mean of (r - threshold)⁴ over returns
// above the threshold.
func (p *PartialMoments) UpperExcessMoment4() float64 { return p.upperExcessKBN.X4() }

// LowerExcessMoment1 returns the mean of (threshold - r) over returns
// below the threshold.
func (p *PartialMoments) LowerExcessMoment1() float64 { return p.lowerExcessKBN.X1() }

// LowerExcessMoment2 returns the mean of (threshold - r)² over returns
// below the threshold.
func (p *PartialMoments) LowerExcessMoment2() float64 { return p.lowerExcessKBN.X2() }

// LowerExcessMoment2Sum returns the sum of (threshold - r)² over returns
// below the threshold.
func (p *PartialMoments) LowerExcessMoment2Sum() float64 { return p.lowerExcessKBN.X2Sum() }

// LowerExcessMoment3 returns the mean of (threshold - r)³ over returns
// below the threshold.
func (p *PartialMoments) LowerExcessMoment3() float64 { return p.lowerExcessKBN.X3() }

// LowerExcessMoment4 returns the mean of (threshold - r)⁴ over returns
// below the threshold.
func (p *PartialMoments) LowerExcessMoment4() float64 { return p.lowerExcessKBN.X4() }
