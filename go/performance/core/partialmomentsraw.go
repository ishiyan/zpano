package core

import (
	"zpano/streamingkbn"
)

// RawPartialMoments computes streaming raw (unnormalized, threshold 0)
// lower/higher partial moments, together with counts and sums of the
// positive and negative returns.
//
// The zero value is ready to use.
type RawPartialMoments struct {
	count    int
	countPos int
	countNeg int
	lpmKBN   streamingkbn.KleinKBNAccumulator
	hpmKBN   streamingkbn.KleinKBNAccumulator
	posKBN   streamingkbn.KleinKBNAccumulator
	negKBN   streamingkbn.KleinKBNAccumulator
}

// NewRawPartialMoments returns a new empty RawPartialMoments.
func NewRawPartialMoments() *RawPartialMoments {
	return &RawPartialMoments{}
}

// Reset clears all accumulated state.
func (p *RawPartialMoments) Reset() {
	*p = RawPartialMoments{}
}

// Revert removes a previously added return.
func (p *RawPartialMoments) Revert(ret float64) {
	p.count--
	// Lower partial moment
	pm := -ret
	if pm < 0 {
		pm = 0
	}
	p.lpmKBN.Revert(pm)

	// Higher partial moment
	pm = ret
	if pm < 0 {
		pm = 0
	}
	p.hpmKBN.Revert(pm)

	if ret > 0 {
		p.countPos--
		p.posKBN.Revert(ret)
	} else if ret < 0 {
		p.countNeg--
		p.negKBN.Revert(ret)
	}
}

// Update adds a return.
func (p *RawPartialMoments) Update(ret float64) {
	p.count++
	// Lower partial moment
	pm := -ret
	if pm < 0 {
		pm = 0
	}
	p.lpmKBN.Update(pm)

	// Higher partial moment
	pm = ret
	if pm < 0 {
		pm = 0
	}
	p.hpmKBN.Update(pm)

	if ret > 0 {
		p.countPos++
		p.posKBN.Update(ret)
	} else if ret < 0 {
		p.countNeg++
		p.negKBN.Update(ret)
	}
}

// Count returns the number of accumulated returns.
func (p *RawPartialMoments) Count() int { return p.count }

// LowerPartialMoment1 returns sum(max(-r, 0)).
func (p *RawPartialMoments) LowerPartialMoment1() float64 { return p.lpmKBN.Value() }

// HigherPartialMoment1 returns sum(max(r, 0)).
func (p *RawPartialMoments) HigherPartialMoment1() float64 { return p.hpmKBN.Value() }

// CountNegative returns the number of negative returns.
func (p *RawPartialMoments) CountNegative() int { return p.countNeg }

// SumNegative returns the sum of negative returns.
func (p *RawPartialMoments) SumNegative() float64 { return p.negKBN.Value() }

// CountPositive returns the number of positive returns.
func (p *RawPartialMoments) CountPositive() int { return p.countPos }

// SumPositive returns the sum of positive returns.
func (p *RawPartialMoments) SumPositive() float64 { return p.posKBN.Value() }
