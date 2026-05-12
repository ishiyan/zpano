package core

// Criterion specifies a threshold based on the average value of a candlestick range entity.
//
// The criteria are based on parts of the candlestick and common words indicating length
// (short, long, very long), displacement (near, far), or equality (equal).
//
// For streaming efficiency, the criterion maintains a running total that is updated
// incrementally via add() and remove() rather than rescanning the entire history.
type Criterion struct {
	// Entity is the type of range entity to consider (RealBody, HighLow, or Shadows).
	Entity RangeEntity
	// AveragePeriod is the number of previous candlesticks to calculate an average value.
	AveragePeriod int
	// Factor is the coefficient to multiply the average value.
	Factor float64
}

// Copy creates an independent copy.
func (c Criterion) Copy() Criterion {
	return Criterion{c.Entity, c.AveragePeriod, c.Factor}
}

// AverageValueFromTotal computes the criterion threshold from a precomputed running total.
//
// When AveragePeriod > 0, uses the running total.
// When AveragePeriod == 0, uses the current candle's own range value.
func (c Criterion) AverageValueFromTotal(total, o, h, l, cl float64) float64 {
	if c.AveragePeriod > 0 {
		if c.Entity == Shadows {
			return c.Factor * total / (float64(c.AveragePeriod) * 2.0)
		}
		return c.Factor * total / float64(c.AveragePeriod)
	}
	// Period == 0: use the candle's own range value directly.
	return c.Factor * CandleRangeValue(c.Entity, o, h, l, cl)
}

// CandleContribution computes the contribution of a single candle to the running total.
//
// For Shadows entity, this returns the full (upper + lower) shadow sum
// (not yet divided by 2 -- the division happens in AverageValueFromTotal).
func (c Criterion) CandleContribution(o, h, l, cl float64) float64 {
	switch c.Entity {
	case RealBody:
		if cl >= o {
			return cl - o
		}
		return o - cl
	case HighLow:
		return h - l
	default:
		// SHADOWS: upper + lower shadow sum (division by 2 deferred to AverageValueFromTotal)
		if cl >= o {
			return h - cl + o - l
		}
		return h - o + cl - l
	}
}
