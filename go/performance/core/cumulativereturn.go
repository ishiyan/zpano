package core

import (
	"errors"
	"math"

	"zpano/streamingkbn"
)

// CumulativeReturn computes streaming cumulative (geometric) returns.
//
// It accumulates the sum of log returns, log(1 + r), with compensated
// summation. Because only a sum is stored, Revert may remove any
// previously added return, so the type works for FIFO rolling windows:
// the caller owns the window and feeds evicted returns to Revert.
//
// The zero value is ready to use.
type CumulativeReturn struct {
	cumlogretSum streamingkbn.KleinKBNAccumulator
	count        int
}

// NewCumulativeReturn returns a new empty CumulativeReturn.
func NewCumulativeReturn() *CumulativeReturn {
	return &CumulativeReturn{}
}

// Reset clears all accumulated state.
func (c *CumulativeReturn) Reset() {
	c.cumlogretSum.Reset()
	c.count = 0
}

func cumulativeLogret(ret float64) float64 {
	if ret != 0 {
		return math.Log1p(ret)
	}
	return 0
}

// Revert removes a previously added return.
//
// Returns an error if there are no returns.
func (c *CumulativeReturn) Revert(ret float64) error {
	if c.count <= 0 {
		return errors.New("Cannot revert from an empty accumulator")
	}
	c.count--
	c.cumlogretSum.Revert(cumulativeLogret(ret))
	return nil
}

// Update adds a return, expressed as a decimal (must be > -1).
func (c *CumulativeReturn) Update(ret float64) {
	c.count++
	c.cumlogretSum.Update(cumulativeLogret(ret))
}

// Count returns the number of accumulated returns.
func (c *CumulativeReturn) Count() int {
	return c.count
}

// CumulativeGeometricReturn returns the cumulative geometric return,
// prod(1 + r) - 1 (0.0 when empty).
func (c *CumulativeReturn) CumulativeGeometricReturn() float64 {
	return math.Expm1(c.cumlogretSum.Value())
}

// GeometricMeanReturn returns the geometric mean of the returns,
// prod(1 + r)^(1/n) - 1 (NaN when empty).
func (c *CumulativeReturn) GeometricMeanReturn() float64 {
	if c.count > 0 {
		return math.Expm1(c.cumlogretSum.Value() / float64(c.count))
	}
	return math.NaN()
}

// AnnualizedGeometricMeanReturn returns the annualized geometric mean,
// prod(1 + r)^(periodsPerYear/n) - 1 (NaN when empty).
func (c *CumulativeReturn) AnnualizedGeometricMeanReturn(periodsPerYear float64) float64 {
	if c.count == 0 {
		return math.NaN()
	}
	return math.Expm1(c.cumlogretSum.Value() * periodsPerYear / float64(c.count))
}
