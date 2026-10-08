package core

import (
	"math"

	"zpano/streamingkbn"
)

// Capture is a streaming accumulator of upside/downside capture ratios
// (geometric and arithmetic), up/down number ratios and up/down
// percentage ratios of an asset relative to a benchmark.
//
// Following the PerformanceAnalytics conventions, upside periods have
// benchmark > 0; downside capture and down-number use benchmark <= 0;
// down-percentage uses benchmark < 0.
//
// Revert may remove any previously added pair, so the type works for
// FIFO rolling windows. The zero value is ready to use.
type Capture struct {
	logretASumUp streamingkbn.KleinKBNAccumulator
	logretBSumUp streamingkbn.KleinKBNAccumulator
	logretASumDn streamingkbn.KleinKBNAccumulator
	logretBSumDn streamingkbn.KleinKBNAccumulator
	aSumUp       streamingkbn.KleinKBNAccumulator
	bSumUp       streamingkbn.KleinKBNAccumulator
	aSumDn       streamingkbn.KleinKBNAccumulator
	bSumDn       streamingkbn.KleinKBNAccumulator
	aNumUp       int
	bNumUp       int
	aNumDn       int
	bNumDn       int
	aPercUp      int
	bPercUp      int
	aPercDn      int
	bPercDn      int
}

// NewCapture returns a new empty Capture.
func NewCapture() *Capture {
	return &Capture{}
}

// Reset clears all accumulated state.
func (c *Capture) Reset() {
	*c = Capture{}
}

func captureLogret(ret float64) float64 {
	if ret != 0 {
		return math.Log1p(ret)
	}
	return 0
}

// Revert removes a previously added (asset, benchmark) return pair.
func (c *Capture) Revert(retAsset, retBenchmark float64) {
	if retBenchmark > 0 { // Upside
		// Geometric
		c.logretASumUp.Revert(captureLogret(retAsset))
		c.logretBSumUp.Revert(captureLogret(retBenchmark))
		// Arithmetic
		c.aSumUp.Revert(retAsset)
		c.bSumUp.Revert(retBenchmark)
		// Number
		c.bNumUp--
		if retAsset > 0 {
			c.aNumUp--
		}
		// Percentage
		c.bPercUp--
		if retAsset > retBenchmark {
			c.aPercUp--
		}
	} else { // Downside
		// Geometric
		c.logretASumDn.Revert(captureLogret(retAsset))
		c.logretBSumDn.Revert(captureLogret(retBenchmark))
		// Arithmetic
		c.aSumDn.Revert(retAsset)
		c.bSumDn.Revert(retBenchmark)
		// Number
		c.bNumDn--
		if retAsset < 0 {
			c.aNumDn--
		}
		// Percentage
		if retBenchmark < 0 {
			c.bPercDn--
			if retAsset > retBenchmark {
				c.aPercDn--
			}
		}
	}
}

// Update adds an (asset, benchmark) return pair.
func (c *Capture) Update(retAsset, retBenchmark float64) {
	if retBenchmark > 0 { // Upside
		// Geometric
		c.logretASumUp.Update(captureLogret(retAsset))
		c.logretBSumUp.Update(captureLogret(retBenchmark))
		// Arithmetic
		c.aSumUp.Update(retAsset)
		c.bSumUp.Update(retBenchmark)
		// Counts
		c.bNumUp++
		if retAsset > 0 {
			c.aNumUp++
		}
		// Perc
		c.bPercUp++
		if retAsset > retBenchmark {
			c.aPercUp++
		}
	} else { // Downside
		// Geometric
		c.logretASumDn.Update(captureLogret(retAsset))
		c.logretBSumDn.Update(captureLogret(retBenchmark))
		// Arithmetic
		c.aSumDn.Update(retAsset)
		c.bSumDn.Update(retBenchmark)
		// Counts
		c.bNumDn++
		if retAsset < 0 {
			c.aNumDn++
		}
		// Perc
		if retBenchmark < 0 {
			c.bPercDn++
			if retAsset > retBenchmark {
				c.aPercDn++
			}
		}
	}
}

// UpsideCaptureRatioGeometric returns the ratio of the compounded asset
// return to the compounded benchmark return over upside periods
// (NaN when the compounded benchmark return is zero).
func (c *Capture) UpsideCaptureRatioGeometric() float64 {
	aCum := math.Expm1(c.logretASumUp.Value())
	bCum := math.Expm1(c.logretBSumUp.Value())
	if bCum != 0 {
		return aCum / bCum
	}
	return math.NaN()
}

// UpsideCaptureRatioArithmetic returns the ratio of the summed asset
// returns to the summed benchmark returns over upside periods
// (NaN when the benchmark sum is zero).
func (c *Capture) UpsideCaptureRatioArithmetic() float64 {
	bSum := c.bSumUp.Value()
	if bSum != 0 {
		return c.aSumUp.Value() / bSum
	}
	return math.NaN()
}

// DownsideCaptureRatioGeometric returns the ratio of the compounded asset
// return to the compounded benchmark return over downside periods
// (NaN when the compounded benchmark return is zero).
func (c *Capture) DownsideCaptureRatioGeometric() float64 {
	aCum := math.Expm1(c.logretASumDn.Value())
	bCum := math.Expm1(c.logretBSumDn.Value())
	if bCum != 0 {
		return aCum / bCum
	}
	return math.NaN()
}

// DownsideCaptureRatioArithmetic returns the ratio of the summed asset
// returns to the summed benchmark returns over downside periods
// (NaN when the benchmark sum is zero).
func (c *Capture) DownsideCaptureRatioArithmetic() float64 {
	bSum := c.bSumDn.Value()
	if bSum != 0 {
		return c.aSumDn.Value() / bSum
	}
	return math.NaN()
}

// UpNumberRatio returns the fraction of upside periods in which the asset
// return is positive (NaN when there are no upside periods).
func (c *Capture) UpNumberRatio() float64 {
	bNum := c.bNumUp
	if bNum != 0 {
		return float64(c.aNumUp) / float64(bNum)
	}
	return math.NaN()
}

// DownNumberRatio returns the fraction of downside periods in which the
// asset return is negative (NaN when there are no downside periods).
func (c *Capture) DownNumberRatio() float64 {
	bNum := c.bNumDn
	if bNum != 0 {
		return float64(c.aNumDn) / float64(bNum)
	}
	return math.NaN()
}

// UpPercentageRatio returns the fraction of upside periods in which the
// asset outperforms the benchmark (NaN when there are no upside periods).
func (c *Capture) UpPercentageRatio() float64 {
	bPerc := c.bPercUp
	if bPerc != 0 {
		return float64(c.aPercUp) / float64(bPerc)
	}
	return math.NaN()
}

// DownPercentageRatio returns the fraction of strictly negative benchmark
// periods in which the asset outperforms the benchmark (NaN when there
// are no such periods).
func (c *Capture) DownPercentageRatio() float64 {
	bPerc := c.bPercDn
	if bPerc != 0 {
		return float64(c.aPercDn) / float64(bPerc)
	}
	return math.NaN()
}
