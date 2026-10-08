package performance

import (
	"errors"
	"math"

	"zpano/performance/core"
	"zpano/streamingkbn"
)

const sqrt2 = 1.4142135623730950488016887242097

// Common periods-per-annum conventions.
const (
	PeriodsPerAnnumYear    = 1
	PeriodsPerAnnumQuarter = 4
	PeriodsPerAnnumMonth   = 12
	PeriodsPerAnnumWeek    = 52
	PeriodsPerAnnumDay     = 252

	// PeriodsPerAnnumMinuteUSEquities is 390 regular-session minutes/day
	// by 252 trading days/year.
	PeriodsPerAnnumMinuteUSEquities = 98280

	// PeriodsPerAnnumMinuteCrypto is 1440 minutes/day by 365 days/year.
	PeriodsPerAnnumMinuteCrypto = 525600
)

var (
	errPeriodsPerAnnum  = errors.New("periods_per_annum must be positive")
	errConfidence       = errors.New("confidence must be between 0 and 1")
	errUpperOrder       = errors.New("upper_order must be 1, 2, 3, or 4")
	errLowerOrder       = errors.New("lower_order must be 1, 2, 3, or 4")
	errAlpha            = errors.New("alpha must be between 0 and 1")
	errBeta             = errors.New("beta must be between 0 and 1")
	errCutoff           = errors.New("cutoff must be between 0.5 and 1.0")
	errStdDevMultiplier = errors.New("std_dev_multiplier must be positive")
)

// Measures is a streaming calculator of time-series performance and risk
// measures of a portfolio and its benchmark.
//
// It operates on return observations with a common, explicitly defined
// observation period; it does not use timestamps and performs no
// resampling. The periodsPerAnnum argument of [NewMeasures] defines the
// annualization convention: the number of return observations assumed
// to represent one year (252 for daily equity returns, 52 weekly,
// 12 monthly, 4 quarterly, 1 annual, 98280 for one-minute US equity
// session returns). Irregular event-based observations, such as tick
// returns, should be transformed into a regular time-based return series
// first.
//
// Use [NewMeasures] to create an instance. A Measures is not safe for
// concurrent use.
type Measures struct {
	periodsPerAnnum     float64
	sqrtPeriodsPerAnnum float64
	annualRiskFreeRate  float64
	riskFreeRate        float64
	targetReturn        float64
	rollingWindowSize   int

	returns          []float64
	returnsBenchmark []float64

	winLoss *core.WinLoss
	capture *core.Capture

	// All RawMomentsKleinKBN use ddof=1, bias=true, fisher=true, which
	// matches scipy's default behavior for kurtosis.
	returnsKBN                *streamingkbn.RawMomentsKleinKBN
	excessReturnsKBN          *streamingkbn.RawMomentsKleinKBN
	benchmarkReturnsKBN       *streamingkbn.RawMomentsKleinKBN
	benchmarkExcessReturnsKBN *streamingkbn.RawMomentsKleinKBN

	sfmRegression    *core.SFMRegression
	activeReturnsKBN *streamingkbn.RawMomentsKleinKBN

	targetReturnsKBN               *streamingkbn.RawMomentsKleinKBN
	targetPartialMoments           *core.PartialMoments
	rawPartialMoments              *core.RawPartialMoments
	benchmarkTargetPartialMoments  *core.PartialMoments
	cumulativeReturn               *core.CumulativeReturn
	cumulativeExcessReturn         *core.CumulativeReturn
	benchmarkCumulativeReturn      *core.CumulativeReturn
	drawdownContinuousRuns         *core.ContinuousDrawdownRuns
	drawdownHighWatermark          *core.HighWaterMarkDrawdown
	drawdownHighWatermarkBenchmark *core.HighWaterMarkDrawdown
	drawdownEpisodes               *core.DrawdownEpisodes
	drawdownEpisodesBenchmark      *core.DrawdownEpisodes
}

// NewMeasures creates a new empty Measures.
//
// periodsPerAnnum is the number of return periods per annum used for
// annualization (Python default 252); it must be positive. It also
// determines the conversion of the annual rates to periodic rates: with
// periodsPerAnnum = 12, an annual rate r is converted to a monthly rate
// (1 + r)^(1/12) - 1. The annual rate is used as is when it is zero or
// when periodsPerAnnum is 1.
//
// annualRiskFreeRate is the annual risk-free rate (Python default 0) used
// by measures that require a risk-free rate.
//
// annualTargetReturn is the annual target return, also known as the
// minimum acceptable return (MAR) (Python default 0), used by measures
// that require a target return.
//
// rollingWindowSize is the maximum number of most recent observations
// retained in the rolling window (Python default 0). Zero or a negative
// value specifies an unbounded running window.
//
// Returns an error if periodsPerAnnum is not positive.
func NewMeasures(periodsPerAnnum, annualRiskFreeRate, annualTargetReturn float64,
	rollingWindowSize int,
) (*Measures, error) {
	if periodsPerAnnum <= 0 {
		return nil, errPeriodsPerAnnum
	}

	m := &Measures{
		periodsPerAnnum:     periodsPerAnnum,
		sqrtPeriodsPerAnnum: math.Sqrt(periodsPerAnnum),
		annualRiskFreeRate:  annualRiskFreeRate,
	}

	if annualRiskFreeRate == 0 || periodsPerAnnum == 1 {
		m.riskFreeRate = annualRiskFreeRate
	} else {
		m.riskFreeRate = math.Pow(1+annualRiskFreeRate, 1/periodsPerAnnum) - 1
	}

	if annualTargetReturn == 0 || periodsPerAnnum == 1 {
		m.targetReturn = annualTargetReturn
	} else {
		m.targetReturn = math.Pow(1+annualTargetReturn, 1/periodsPerAnnum) - 1
	}

	if rollingWindowSize < 0 {
		rollingWindowSize = 0
	}
	m.rollingWindowSize = rollingWindowSize

	newKBN := func() *streamingkbn.RawMomentsKleinKBN {
		return streamingkbn.NewRawMomentsKleinKBN(1, true, true)
	}

	m.winLoss = core.NewWinLoss()
	m.capture = core.NewCapture()
	m.returnsKBN = newKBN()
	m.excessReturnsKBN = newKBN()
	m.benchmarkReturnsKBN = newKBN()
	m.benchmarkExcessReturnsKBN = newKBN()
	m.sfmRegression = core.NewSFMRegression(m.riskFreeRate)
	m.activeReturnsKBN = newKBN()
	m.targetReturnsKBN = newKBN()
	m.targetPartialMoments = core.NewPartialMoments(m.targetReturn)
	m.rawPartialMoments = core.NewRawPartialMoments()
	m.benchmarkTargetPartialMoments = core.NewPartialMoments(m.targetReturn)
	m.cumulativeReturn = core.NewCumulativeReturn()
	m.cumulativeExcessReturn = core.NewCumulativeReturn()
	m.benchmarkCumulativeReturn = core.NewCumulativeReturn()
	m.drawdownContinuousRuns = core.NewContinuousDrawdownRuns()
	m.drawdownHighWatermark = core.NewHighWaterMarkDrawdown(m.rollingWindowSize)
	m.drawdownHighWatermarkBenchmark = core.NewHighWaterMarkDrawdown(m.rollingWindowSize)
	m.drawdownEpisodes = core.NewDrawdownEpisodes()
	m.drawdownEpisodesBenchmark = core.NewDrawdownEpisodes()

	return m, nil
}

// PeriodsPerAnnum returns the number of return periods per annum used for
// annualization.
func (m *Measures) PeriodsPerAnnum() float64 { return m.periodsPerAnnum }

// RiskFreeRate returns the periodic risk-free rate derived from the
// annual risk-free rate.
func (m *Measures) RiskFreeRate() float64 { return m.riskFreeRate }

// TargetReturn returns the periodic target return (MAR) derived from the
// annual target return.
func (m *Measures) TargetReturn() float64 { return m.targetReturn }

// RollingWindowSize returns the rolling window size, 0 meaning an
// unbounded window.
func (m *Measures) RollingWindowSize() int { return m.rollingWindowSize }

// Reset clears all accumulated return data and derived streaming state.
//
// After Reset the instance behaves as if no returns had been added. The
// configuration (periods per annum, rates, rolling window size) is
// preserved.
func (m *Measures) Reset() {
	m.returns = m.returns[:0]
	m.returnsBenchmark = m.returnsBenchmark[:0]

	m.winLoss.Reset()
	m.capture.Reset()

	m.returnsKBN.Reset()
	m.excessReturnsKBN.Reset()
	m.benchmarkReturnsKBN.Reset()
	m.benchmarkExcessReturnsKBN.Reset()

	m.sfmRegression.Reset()
	m.activeReturnsKBN.Reset()

	m.targetReturnsKBN.Reset()
	m.targetPartialMoments.Reset()
	m.rawPartialMoments.Reset()
	m.benchmarkTargetPartialMoments.Reset()

	m.cumulativeReturn.Reset()
	m.cumulativeExcessReturn.Reset()
	m.benchmarkCumulativeReturn.Reset()

	m.drawdownContinuousRuns.Reset()
	m.drawdownHighWatermark.Reset()
	m.drawdownHighWatermarkBenchmark.Reset()
	m.drawdownEpisodes.Reset()
	m.drawdownEpisodesBenchmark.Reset()
}

// AddReturn adds one periodic portfolio return ret and the benchmark
// return retBench of the same observation period, both expressed as
// decimals.
//
// The returns must use the periodicity implied by periods per annum; no
// timestamp-based normalization is performed. The benchmark return is
// used by benchmark-relative measures (beta, alpha, tracking error,
// information ratio, systematic risk, capture ratios, ...).
//
// Inputs are not validated; a return <= -1 produces NaN or ±Inf in the
// logarithmic accumulators.
func (m *Measures) AddReturn(ret, retBench float64) {
	rf := m.riskFreeRate
	evicted := m.rollingWindowSize > 0 && len(m.returns) == m.rollingWindowSize
	if evicted {
		retOld := m.returns[0]
		retBenchOld := m.returnsBenchmark[0]
		m.returns = m.returns[1:]
		m.returnsBenchmark = m.returnsBenchmark[1:]
		m.returnsKBN.Revert(retOld)
		// Excess returns (returns less risk-free rate).
		m.excessReturnsKBN.Revert(retOld - rf)
		// Target returns (returns less target return).
		m.targetReturnsKBN.Revert(retOld - m.targetReturn)
		m.targetPartialMoments.Revert(retOld)
		m.rawPartialMoments.Revert(retOld)
		m.benchmarkTargetPartialMoments.Revert(retBenchOld)
		m.winLoss.Revert(retOld)
		m.capture.Revert(retOld, retBenchOld)
		// The accumulators are not empty here, so Revert cannot fail.
		_ = m.cumulativeReturn.Revert(retOld)
		_ = m.cumulativeExcessReturn.Revert(retOld - rf)
		_ = m.benchmarkCumulativeReturn.Revert(retBenchOld)
		// Benchmarks.
		m.benchmarkReturnsKBN.Revert(retBenchOld)
		m.benchmarkExcessReturnsKBN.Revert(retBenchOld - rf)
		m.activeReturnsKBN.Revert(retOld - retBenchOld)
		m.sfmRegression.Revert(retOld, retBenchOld)
		// Drawdowns. The high-watermark drawdown and drawdown episodes
		// have no Revert; they evict the oldest observation themselves.
		m.drawdownContinuousRuns.Revert(retOld) // Burke
	}

	m.returnsKBN.Update(ret)
	// Excess returns (returns less risk-free rate).
	retExcess := ret - rf
	m.excessReturnsKBN.Update(retExcess)
	// Target returns (returns less target return).
	m.targetReturnsKBN.Update(ret - m.targetReturn)
	m.targetPartialMoments.Update(ret)
	m.rawPartialMoments.Update(ret)
	m.benchmarkTargetPartialMoments.Update(retBench)
	m.winLoss.Update(ret)
	m.capture.Update(ret, retBench)
	// Benchmarks.
	m.benchmarkReturnsKBN.Update(retBench)
	retBenchExcess := retBench - rf
	m.benchmarkExcessReturnsKBN.Update(retBenchExcess)
	m.activeReturnsKBN.Update(ret - retBench)
	m.sfmRegression.Update(ret, retBench)

	m.returns = append(m.returns, ret)
	m.returnsBenchmark = append(m.returnsBenchmark, retBench)

	// Cumulative return.
	m.cumulativeReturn.Update(ret)
	m.cumulativeExcessReturn.Update(retExcess)
	m.benchmarkCumulativeReturn.Update(retBench)

	// Drawdown calculation used in Burke.
	m.drawdownContinuousRuns.Update(ret)

	// High-water-mark drawdown and drawdown episodes. Drawdown episodes
	// have no Revert: when an observation leaves the rolling window (or
	// the window's drawdowns were recomputed), the episodes are rebuilt
	// from the window's drawdowns, so that episode indices refer to
	// positions in the window.
	hwm := m.drawdownHighWatermark
	if recalculated := hwm.Update(ret); recalculated || evicted {
		m.drawdownEpisodes.Recalculate(hwm.Drawdowns())
	} else {
		m.drawdownEpisodes.Update(hwm.Drawdown())
	}
	hwm = m.drawdownHighWatermarkBenchmark
	if recalculated := hwm.Update(retBench); recalculated || evicted {
		m.drawdownEpisodesBenchmark.Recalculate(hwm.Drawdowns())
	} else {
		m.drawdownEpisodesBenchmark.Update(hwm.Drawdown())
	}
}

// divOrNaN returns num / den, or NaN when den is zero.
func divOrNaN(num, den float64) float64 {
	if den != 0 {
		return num / den
	}
	return math.NaN()
}

// pySum mirrors Python's built-in sum() for floats (CPython 3.12+ uses
// Neumaier compensated summation).
func pySum(values []float64) float64 {
	result, c := 0.0, 0.0
	for _, x := range values {
		t := result + x
		if math.Abs(result) >= math.Abs(x) {
			c += (result - t) + x
		} else {
			c += (x - t) + result
		}
		result = t
	}
	if c != 0 && !math.IsInf(c, 0) && !math.IsNaN(c) {
		result += c
	}
	return result
}
