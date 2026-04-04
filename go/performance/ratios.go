package performance

import (
	"math"
	"sort"
	"time"

	"portf_py/daycounting"
	"portf_py/daycounting/conventions"
)

const sqrt2 = 1.4142135623730950488016887242097

// Ratios accumulates portfolio returns incrementally and computes
// various financial performance ratios at each step.
type Ratios struct {
	periodicity        Periodicity
	periodsPerAnnum    int
	daysPerPeriod      float64
	riskFreeRate       float64
	requiredReturn     float64
	dayCountConvention conventions.DayCountConvention
	rollingWindow      *int
	minPeriods         *int

	fractionalPeriods []float64
	returns           []float64
	sampleCount       int

	logretSum                     float64
	drawdownsCumulative           []float64
	drawdownsCumulativeMin        float64
	drawdownsPeaks                []float64
	drawdownsPeaksPeak            int
	drawdownContinuous            []float64
	drawdownContinuousFinal       []float64
	drawdownContinuousFinalized   bool
	drawdownContinuousPeak        int
	drawdownContinuousInside      bool
	cumulativeReturnPlus1         float64
	cumulativeReturnPlus1Max      float64
	cumulativeReturnGeometricMean *float64
	returnsMean                   *float64
	returnsStd                    *float64
	returnsAutocorrPenalty        float64
	excessMean                    *float64
	excessStd                     *float64
	excessAutocorrPenalty         float64
	requiredMean                  *float64
	requiredLPM1                  *float64
	requiredLPM2                  *float64
	requiredLPM3                  *float64
	requiredHPM1                  *float64
	requiredHPM2                  *float64
	requiredHPM3                  *float64
	requiredAutocorrPenalty       float64
	avgReturn                     *float64
	avgWin                        *float64
	avgLoss                       *float64
	winRate                       *float64
	totalDuration                 float64

	resetCalled bool
}

// New creates a new Ratios instance with the specified parameters.
//
// The annual rates are de-annualized to per-period rates based on the periodicity.
// rollingWindow, if non-nil, limits computations to the last N returns.
// minPeriods, if non-nil and > 0, causes all ratio methods to return nil
// until at least that many samples have been added.
func New(
	periodicity Periodicity,
	annualRiskFreeRate float64,
	annualTargetReturn float64,
	dayCountConvention conventions.DayCountConvention,
	rollingWindow *int,
	minPeriods *int,
) *Ratios {
	ppa := periodicity.PeriodsPerAnnum()
	dpp := periodicity.DaysPerPeriod()

	rfr := annualRiskFreeRate
	if annualRiskFreeRate != 0 && ppa != 1 {
		rfr = math.Pow(1+annualRiskFreeRate, 1.0/float64(ppa)) - 1
	}

	rr := annualTargetReturn
	if annualTargetReturn != 0 && ppa != 1 {
		rr = math.Pow(1+annualTargetReturn, 1.0/float64(ppa)) - 1
	}

	// Treat nil or <=0 minPeriods as no minimum
	var mp *int
	if minPeriods != nil && *minPeriods > 0 {
		mp = minPeriods
	}

	return &Ratios{
		periodicity:        periodicity,
		periodsPerAnnum:    ppa,
		daysPerPeriod:      dpp,
		riskFreeRate:       rfr,
		requiredReturn:     rr,
		dayCountConvention: dayCountConvention,
		rollingWindow:      rollingWindow,
		minPeriods:         mp,
	}
}

// Reset initializes/resets all internal state for accumulation.
func (r *Ratios) Reset() {
	r.fractionalPeriods = []float64{}
	r.returns = []float64{}
	r.sampleCount = 0
	r.logretSum = 0
	r.drawdownsCumulative = []float64{}
	r.drawdownsCumulativeMin = math.Inf(1)
	r.drawdownsPeaks = []float64{}
	r.drawdownsPeaksPeak = 0
	r.drawdownContinuous = []float64{}
	r.drawdownContinuousFinal = []float64{}
	r.drawdownContinuousFinalized = false
	r.drawdownContinuousPeak = 1
	r.drawdownContinuousInside = false
	r.cumulativeReturnPlus1 = 1
	r.cumulativeReturnPlus1Max = math.Inf(-1)
	r.totalDuration = 0
	r.cumulativeReturnGeometricMean = nil
	r.returnsMean = nil
	r.returnsStd = nil
	r.returnsAutocorrPenalty = 1
	r.excessMean = nil
	r.excessStd = nil
	r.excessAutocorrPenalty = 1
	r.requiredMean = nil
	r.requiredLPM1 = nil
	r.requiredLPM2 = nil
	r.requiredLPM3 = nil
	r.requiredHPM1 = nil
	r.requiredHPM2 = nil
	r.requiredHPM3 = nil
	r.requiredAutocorrPenalty = 1
	r.avgReturn = nil
	r.avgWin = nil
	r.avgLoss = nil
	r.winRate = nil
	r.resetCalled = true
}

// AddReturn adds a new return observation and updates all internal state.
func (r *Ratios) AddReturn(
	returnVal float64,
	returnBenchmark float64,
	value float64,
	timeStart time.Time,
	timeEnd time.Time,
) {
	var fractionalPeriod float64
	if r.periodicity == Annual {
		fp, err := daycounting.YearFrac(timeStart, timeEnd, r.dayCountConvention)
		if err != nil {
			return
		}
		fractionalPeriod = fp
	} else {
		fp, err := daycounting.DayFrac(timeStart, timeEnd, r.dayCountConvention)
		if err != nil {
			return
		}
		fractionalPeriod = fp / r.daysPerPeriod
	}

	r.fractionalPeriods = append(r.fractionalPeriods, fractionalPeriod)
	if fractionalPeriod == 0 {
		return
	}
	r.totalDuration += fractionalPeriod
	r.sampleCount++

	// Normalized return
	ret := returnVal / fractionalPeriod
	r.returns = append(r.returns, ret)

	// Window slice: use last rollingWindow returns, or all if not set
	w := r.returns
	if r.rollingWindow != nil {
		n := *r.rollingWindow
		if len(r.returns) > n {
			w = r.returns[len(r.returns)-n:]
		}
	}
	l := len(w)
	lf := float64(l)

	// Returns mean
	mean := sliceMean(w)
	r.returnsMean = &mean

	// Returns std (ddof=1, sample)
	if l > 1 {
		s := sliceStdDdof1(w, mean)
		r.returnsStd = &s
	} else {
		r.returnsStd = nil
	}

	r.returnsAutocorrPenalty = r.autocorrPenalty(w)

	// Average return, win rate, avg win, avg loss
	nonZero := filterNonZero(w)
	lenNonZero := len(nonZero)
	if lenNonZero > 0 {
		m := sliceMean(nonZero)
		r.avgReturn = &m

		positive := filterPositive(w)
		lenPos := len(positive)
		wr := float64(lenPos) / float64(lenNonZero)
		r.winRate = &wr

		if lenPos > 0 {
			m2 := sliceMean(positive)
			r.avgWin = &m2
		} else {
			r.avgWin = nil
		}

		negative := filterNegative(w)
		lenNeg := len(negative)
		if lenNeg > 0 {
			m3 := sliceMean(negative)
			r.avgLoss = &m3
		} else {
			r.avgLoss = nil
		}
	} else {
		r.avgReturn = nil
		r.winRate = nil
		r.avgWin = nil
		r.avgLoss = nil
	}

	// Excess returns (returns less risk-free rate)
	if r.riskFreeRate == 0 {
		r.excessMean = r.returnsMean
		r.excessStd = r.returnsStd
		r.excessAutocorrPenalty = r.returnsAutocorrPenalty
	} else {
		excess := make([]float64, l)
		for i, v := range w {
			excess[i] = v - r.riskFreeRate
		}
		em := sliceMean(excess)
		r.excessMean = &em
		if l > 1 {
			es := sliceStdDdof1(excess, em)
			r.excessStd = &es
		} else {
			r.excessStd = nil
		}
		r.excessAutocorrPenalty = r.autocorrPenalty(excess)
	}

	// Lower partial moments for the raw returns (less required return)
	var tmp2 []float64
	if r.requiredReturn == 0 {
		tmp2 = make([]float64, l)
		for i, v := range w {
			tmp2[i] = -v
		}
	} else {
		tmp2 = make([]float64, l)
		for i, v := range w {
			tmp2[i] = r.requiredReturn - v
		}
	}
	// Clip to min 0
	for i, v := range tmp2 {
		if v < 0 {
			tmp2[i] = 0
		}
	}
	lpm1 := sliceSum(tmp2) / lf
	r.requiredLPM1 = &lpm1
	lpm2 := sliceSumPow(tmp2, 2) / lf
	r.requiredLPM2 = &lpm2
	lpm3 := sliceSumPow(tmp2, 3) / lf
	r.requiredLPM3 = &lpm3

	// Higher partial moments for the raw returns (less required return)
	var tmp3 []float64
	if r.requiredReturn == 0 {
		tmp3 = make([]float64, l)
		copy(tmp3, w)
		rm := *r.returnsMean
		r.requiredMean = &rm
		r.requiredAutocorrPenalty = r.returnsAutocorrPenalty
	} else {
		tmp3 = make([]float64, l)
		for i, v := range w {
			tmp3[i] = v - r.requiredReturn
		}
		rm := sliceMean(tmp3)
		r.requiredMean = &rm
		r.requiredAutocorrPenalty = r.autocorrPenalty(tmp3)
	}
	// Clip to min 0
	for i, v := range tmp3 {
		if v < 0 {
			tmp3[i] = 0
		}
	}
	hpm1 := sliceSum(tmp3) / lf
	r.requiredHPM1 = &hpm1
	hpm2 := sliceSumPow(tmp3, 2) / lf
	r.requiredHPM2 = &hpm2
	hpm3 := sliceSumPow(tmp3, 3) / lf
	r.requiredHPM3 = &hpm3

	// Cumulative returns — recompute from window
	wStart := len(r.returns) - l
	logretSum := 0.0
	for j := wStart; j < len(r.returns); j++ {
		fpJ := r.fractionalPeriods[j]
		if fpJ != 0 {
			logretSum += math.Log(w[j-wStart] + 1)
		}
	}
	r.logretSum = logretSum
	cmr := math.Exp(logretSum)
	r.cumulativeReturnPlus1 = cmr
	if l >= 1 {
		gm := math.Pow(cmr, 1.0/lf) - 1
		r.cumulativeReturnGeometricMean = &gm
	}
	r.cumulativeReturnPlus1Max = math.Inf(-1)
	cumr := 1.0
	for j := 0; j < l; j++ {
		cumr *= (w[j] + 1)
		if cumr > r.cumulativeReturnPlus1Max {
			r.cumulativeReturnPlus1Max = cumr
		}
	}

	// Drawdowns from peaks to valleys (cumulative returns) — recompute from window
	r.drawdownsCumulative = make([]float64, 0, l)
	r.drawdownsCumulativeMin = math.Inf(1)
	cumr = 1.0
	cumrMax := math.Inf(-1)
	for j := 0; j < l; j++ {
		cumr *= (w[j] + 1)
		if cumr > cumrMax {
			cumrMax = cumr
		}
		dd := cumr/cumrMax - 1
		r.drawdownsCumulative = append(r.drawdownsCumulative, dd)
		if r.drawdownsCumulativeMin > dd {
			r.drawdownsCumulativeMin = dd
		}
	}

	// Drawdown peaks (used in pain index, ulcer index) — recompute from window
	r.drawdownsPeaks = make([]float64, 0, l)
	r.drawdownsPeaksPeak = 0
	for j := 0; j < l; j++ {
		ddPeak := 1.0
		for k := r.drawdownsPeaksPeak + 1; k <= j; k++ {
			ddPeak *= (1 + w[k]*0.01)
		}
		if ddPeak > 1 {
			r.drawdownsPeaksPeak = j
			r.drawdownsPeaks = append(r.drawdownsPeaks, 0)
		} else {
			r.drawdownsPeaks = append(r.drawdownsPeaks, (ddPeak-1)*100)
		}
	}

	// Drawdown continuous (used in Burke ratio) — recompute from window
	r.drawdownContinuous = make([]float64, 0)
	r.drawdownContinuousFinal = make([]float64, 0)
	r.drawdownContinuousFinalized = false
	r.drawdownContinuousPeak = 1
	r.drawdownContinuousInside = false
	for j := 1; j < l; j++ {
		if w[j] < 0 {
			if !r.drawdownContinuousInside {
				r.drawdownContinuousInside = true
				r.drawdownContinuousPeak = j - 1
			}
			r.drawdownContinuous = append(r.drawdownContinuous, 0)
		} else {
			if r.drawdownContinuousInside {
				ddC := 1.0
				j1 := r.drawdownContinuousPeak + 1
				for k := j1; k < j; k++ {
					ddC *= (1 + w[k]*0.01)
				}
				r.drawdownContinuous = append(r.drawdownContinuous, (ddC-1)*100)
				r.drawdownContinuousInside = false
			} else {
				r.drawdownContinuous = append(r.drawdownContinuous, 0)
			}
		}
	}

	// Suppress unused variable warnings
	_ = returnBenchmark
	_ = value
}

// autocorrPenalty is a stub returning 1, matching all implementations.
func (r *Ratios) autocorrPenalty(returns []float64) float64 {
	return 1
}

// isPrimed returns true if enough samples have been added to satisfy minPeriods.
func (r *Ratios) isPrimed() bool {
	if r.minPeriods == nil {
		return true
	}
	return r.sampleCount >= *r.minPeriods
}

// windowReturns returns the windowed slice of returns.
func (r *Ratios) windowReturns() []float64 {
	if r.rollingWindow == nil {
		return r.returns
	}
	n := *r.rollingWindow
	if len(r.returns) > n {
		return r.returns[len(r.returns)-n:]
	}
	return r.returns
}

// CumulativeReturn returns the cumulative geometric return.
func (r *Ratios) CumulativeReturn() float64 {
	return r.cumulativeReturnPlus1 - 1
}

// DrawdownsCumulative returns the drawdowns from peaks to valleys on cumulative geometric returns.
func (r *Ratios) DrawdownsCumulative() []float64 {
	return r.drawdownsCumulative
}

// MinDrawdownsCumulative returns the minimum (most negative) cumulative drawdown.
func (r *Ratios) MinDrawdownsCumulative() float64 {
	return r.drawdownsCumulativeMin
}

// WorstDrawdownsCumulative returns the absolute value of the worst cumulative drawdown.
func (r *Ratios) WorstDrawdownsCumulative() float64 {
	return math.Abs(r.drawdownsCumulativeMin)
}

// DrawdownsPeaks returns the drawdowns from peaks (used in pain/ulcer indices).
func (r *Ratios) DrawdownsPeaks() []float64 {
	return r.drawdownsPeaks
}

// DrawdownsContinuous returns drawdowns on continuous uninterrupted losing regions.
// If peaksOnly is true, returns only the non-zero values.
// If maxPeaks > 0 and peaksOnly is true, returns at most maxPeaks values (sorted ascending).
func (r *Ratios) DrawdownsContinuous(peaksOnly bool, maxPeaks int) []float64 {
	r.finalizeContinuousDrawdown()
	if !peaksOnly {
		return r.drawdownContinuousFinal
	}
	var drawdowns []float64
	for _, v := range r.drawdownContinuousFinal {
		if v != 0 {
			drawdowns = append(drawdowns, v)
		}
	}
	if maxPeaks > 0 && len(drawdowns) > 0 {
		sort.Float64s(drawdowns)
		if len(drawdowns) > maxPeaks {
			drawdowns = drawdowns[:maxPeaks]
		}
	}
	return drawdowns
}

func (r *Ratios) finalizeContinuousDrawdown() {
	if r.drawdownContinuousFinalized {
		return
	}
	w := r.windowReturns()
	if r.drawdownContinuousInside {
		ddC := 1.0
		j1 := r.drawdownContinuousPeak + 1
		for j := j1; j < len(w); j++ {
			ddC *= (1 + w[j]*0.01)
		}
		r.drawdownContinuousFinal = make([]float64, len(r.drawdownContinuous)+1)
		copy(r.drawdownContinuousFinal, r.drawdownContinuous)
		r.drawdownContinuousFinal[len(r.drawdownContinuous)] = (ddC - 1) * 100
	} else {
		r.drawdownContinuousFinal = make([]float64, len(r.drawdownContinuous)+1)
		copy(r.drawdownContinuousFinal, r.drawdownContinuous)
		r.drawdownContinuousFinal[len(r.drawdownContinuous)] = 0
	}
	r.drawdownContinuousFinalized = true
}

// Skew returns the population skewness of the returns.
// Returns nil if fewer than 2 returns.
func (r *Ratios) Skew() *float64 {
	if !r.isPrimed() {
		return nil
	}
	w := r.windowReturns()
	if len(w) < 2 {
		return nil
	}
	s := populationSkewness(w)
	return &s
}

// Kurtosis returns the population excess kurtosis of the returns.
// Uses m4/m2^2 - 3, matching scipy.stats.kurtosis(bias=True, fisher=True).
// Returns nil if fewer than 2 returns.
func (r *Ratios) Kurtosis() *float64 {
	if !r.isPrimed() {
		return nil
	}
	w := r.windowReturns()
	if len(w) < 2 {
		return nil
	}
	k := populationExcessKurtosis(w)
	return &k
}

// SharpeRatio calculates the ex-post Sharpe ratio.
// If ignoreRiskFreeRate is true, the ratio is calculated over raw returns.
// If autocorrelationPenalty is true, the autocorrelation penalty is applied.
func (r *Ratios) SharpeRatio(ignoreRiskFreeRate bool, autocorrelationPenalty bool) *float64 {
	if !r.isPrimed() {
		return nil
	}
	if ignoreRiskFreeRate {
		if r.returnsMean == nil || r.returnsStd == nil || *r.returnsStd == 0 {
			return nil
		}
		denom := *r.returnsStd
		if autocorrelationPenalty {
			denom *= r.returnsAutocorrPenalty
		}
		v := *r.returnsMean / denom
		return &v
	}
	if r.excessMean == nil || r.excessStd == nil || *r.excessStd == 0 {
		return nil
	}
	denom := *r.excessStd
	if autocorrelationPenalty {
		denom *= r.excessAutocorrPenalty
	}
	v := *r.excessMean / denom
	return &v
}

// SortinoRatio calculates the Sortino ratio.
// If autocorrelationPenalty is true, the penalty is applied.
// If divideBySqrt2 is true, uses Jack Schwager's version for direct comparison to Sharpe.
func (r *Ratios) SortinoRatio(autocorrelationPenalty bool, divideBySqrt2 bool) *float64 {
	if !r.isPrimed() {
		return nil
	}
	if r.requiredMean == nil || r.requiredLPM2 == nil || *r.requiredLPM2 == 0 {
		return nil
	}
	denom := math.Sqrt(*r.requiredLPM2)
	if autocorrelationPenalty {
		denom *= r.requiredAutocorrPenalty
	}
	if divideBySqrt2 {
		denom *= sqrt2
	}
	v := *r.requiredMean / denom
	return &v
}

// OmegaRatio calculates the Omega ratio.
func (r *Ratios) OmegaRatio() *float64 {
	if !r.isPrimed() {
		return nil
	}
	if r.requiredMean == nil || r.requiredLPM1 == nil || *r.requiredLPM1 == 0 {
		return nil
	}
	v := *r.requiredMean / *r.requiredLPM1 + 1
	return &v
}

// KappaRatio calculates the Kappa ratio of a given order.
func (r *Ratios) KappaRatio(order int) *float64 {
	if !r.isPrimed() {
		return nil
	}
	if r.requiredMean == nil {
		return nil
	}
	switch order {
	case 1:
		if r.requiredLPM1 == nil || *r.requiredLPM1 == 0 {
			return nil
		}
		v := *r.requiredMean / *r.requiredLPM1
		return &v
	case 2:
		if r.requiredLPM2 == nil || *r.requiredLPM2 == 0 {
			return nil
		}
		v := *r.requiredMean / math.Sqrt(*r.requiredLPM2)
		return &v
	case 3:
		if r.requiredLPM3 == nil || *r.requiredLPM3 == 0 {
			return nil
		}
		v := *r.requiredMean / math.Cbrt(*r.requiredLPM3)
		return &v
	default:
		w := r.windowReturns()
		l := len(w)
		if l == 0 {
			return nil
		}
		var tmp []float64
		if r.requiredReturn == 0 {
			tmp = make([]float64, l)
			for i, v := range w {
				tmp[i] = -v
			}
		} else {
			tmp = make([]float64, l)
			for i, v := range w {
				tmp[i] = r.requiredReturn - v
			}
		}
		for i, v := range tmp {
			if v < 0 {
				tmp[i] = 0
			}
		}
		lpm := sliceSumPow(tmp, float64(order)) / float64(l)
		if lpm == 0 {
			return nil
		}
		v := *r.requiredMean / math.Pow(lpm, 1.0/float64(order))
		return &v
	}
}

// Kappa3Ratio calculates the Kappa ratio of order 3.
func (r *Ratios) Kappa3Ratio() *float64 {
	if !r.isPrimed() {
		return nil
	}
	if r.requiredMean == nil || r.requiredLPM3 == nil || *r.requiredLPM3 == 0 {
		return nil
	}
	v := *r.requiredMean / math.Cbrt(*r.requiredLPM3)
	return &v
}

// BernardoLedoitRatio calculates the Bernardo-Ledoit ratio.
func (r *Ratios) BernardoLedoitRatio() *float64 {
	if !r.isPrimed() {
		return nil
	}
	w := r.windowReturns()
	l := len(w)
	if l < 1 {
		return nil
	}
	lf := float64(l)

	// LPM1 with threshold=0 (using -returns clipped to min 0)
	tmp := make([]float64, l)
	for i, v := range w {
		tmp[i] = -v
		if tmp[i] < 0 {
			tmp[i] = 0
		}
	}
	lpm1 := sliceSum(tmp) / lf
	if lpm1 == 0 {
		return nil
	}

	// HPM1 with threshold=0 (using returns clipped to min 0)
	for i, v := range w {
		tmp[i] = v
		if tmp[i] < 0 {
			tmp[i] = 0
		}
	}
	hpm1 := sliceSum(tmp) / lf
	v := hpm1 / lpm1
	return &v
}

// UpsidePotentialRatio calculates the upside potential ratio.
// If full is true, uses HPM1/sqrt(LPM2); if false, uses a subset-based calculation.
func (r *Ratios) UpsidePotentialRatio(full bool) *float64 {
	if !r.isPrimed() {
		return nil
	}
	if full {
		if r.requiredHPM1 == nil || r.requiredLPM2 == nil || *r.requiredLPM2 == 0 {
			return nil
		}
		v := *r.requiredHPM1 / math.Sqrt(*r.requiredLPM2)
		return &v
	}
	// Subset version
	w := r.windowReturns()
	var below []float64
	for _, v := range w {
		if v < r.requiredReturn {
			below = append(below, v)
		}
	}
	l := len(below)
	if l < 1 {
		return nil
	}
	lf := float64(l)
	tmp := make([]float64, l)
	for i, v := range below {
		tmp[i] = v - r.requiredReturn
	}
	lpm2 := sliceSumPow(tmp, 2) / lf
	if lpm2 == 0 {
		return nil
	}
	var above []float64
	for _, v := range w {
		if v > r.requiredReturn {
			above = append(above, v-r.requiredReturn)
		}
	}
	if len(above) == 0 {
		return nil
	}
	hpm1 := sliceMean(above)
	v := hpm1 / math.Sqrt(lpm2)
	return &v
}

// CompoundGrowthRate returns the compound (annual) growth rate (CAGR),
// or the geometric mean of the returns.
func (r *Ratios) CompoundGrowthRate() *float64 {
	if !r.isPrimed() {
		return nil
	}
	return r.cumulativeReturnGeometricMean
}

// CalmarRatio calculates the Calmar ratio.
func (r *Ratios) CalmarRatio() *float64 {
	if !r.isPrimed() {
		return nil
	}
	wdd := r.WorstDrawdownsCumulative()
	if wdd == 0 {
		return nil
	}
	if r.cumulativeReturnGeometricMean == nil {
		return nil
	}
	v := *r.cumulativeReturnGeometricMean / wdd
	return &v
}

// SterlingRatio calculates the Sterling ratio with the given annual excess rate.
func (r *Ratios) SterlingRatio(annualExcessRate float64) *float64 {
	if !r.isPrimed() {
		return nil
	}
	excessRate := annualExcessRate
	if annualExcessRate != 0 && r.periodsPerAnnum != 1 {
		excessRate = math.Pow(1+annualExcessRate, 1.0/float64(r.periodsPerAnnum)) - 1
	}
	wdd := r.WorstDrawdownsCumulative() + excessRate
	if wdd == 0 {
		return nil
	}
	if r.cumulativeReturnGeometricMean == nil {
		return nil
	}
	v := *r.cumulativeReturnGeometricMean / wdd
	return &v
}

// BurkeRatio calculates the Burke ratio.
// If modified is true, calculates the modified Burke ratio.
func (r *Ratios) BurkeRatio(modified bool) *float64 {
	if !r.isPrimed() {
		return nil
	}
	if r.cumulativeReturnGeometricMean == nil {
		return nil
	}
	rate := *r.cumulativeReturnGeometricMean - r.riskFreeRate
	drawdowns := r.DrawdownsContinuous(true, 0)
	if len(drawdowns) < 1 {
		return nil
	}
	sumSq := 0.0
	for _, d := range drawdowns {
		sumSq += d * d
	}
	sqrtSumSq := math.Sqrt(sumSq)
	if sqrtSumSq == 0 {
		return nil
	}
	burke := rate / sqrtSumSq
	if modified {
		burke *= math.Sqrt(float64(len(r.windowReturns())))
	}
	v := burke
	return &v
}

// PainIndex calculates the pain index.
func (r *Ratios) PainIndex() *float64 {
	if !r.isPrimed() {
		return nil
	}
	l := len(r.drawdownsPeaks)
	if l < 1 {
		return nil
	}
	// By calculation, all values are <= 0, so we negate the sum
	v := -sliceSum(r.drawdownsPeaks) / float64(l)
	return &v
}

// PainRatio calculates the pain ratio.
func (r *Ratios) PainRatio() *float64 {
	if !r.isPrimed() {
		return nil
	}
	if r.cumulativeReturnGeometricMean == nil {
		return nil
	}
	rate := *r.cumulativeReturnGeometricMean - r.riskFreeRate
	l := len(r.drawdownsPeaks)
	if l < 1 {
		return nil
	}
	painIndex := -sliceSum(r.drawdownsPeaks) / float64(l)
	if painIndex == 0 {
		return nil
	}
	v := rate / painIndex
	return &v
}

// UlcerIndex calculates the ulcer index.
func (r *Ratios) UlcerIndex() *float64 {
	if !r.isPrimed() {
		return nil
	}
	l := len(r.drawdownsPeaks)
	if l < 1 {
		return nil
	}
	sumSq := 0.0
	for _, d := range r.drawdownsPeaks {
		sumSq += d * d
	}
	v := math.Sqrt(sumSq / float64(l))
	return &v
}

// MartinRatio calculates the Martin (Ulcer) ratio.
func (r *Ratios) MartinRatio() *float64 {
	if !r.isPrimed() {
		return nil
	}
	if r.cumulativeReturnGeometricMean == nil {
		return nil
	}
	rate := *r.cumulativeReturnGeometricMean - r.riskFreeRate
	l := len(r.drawdownsPeaks)
	if l < 1 {
		return nil
	}
	sumSq := 0.0
	for _, d := range r.drawdownsPeaks {
		sumSq += d * d
	}
	ulcerIndex := math.Sqrt(sumSq / float64(l))
	if ulcerIndex == 0 {
		return nil
	}
	v := rate / ulcerIndex
	return &v
}

// GainToPainRatio returns Jack Schwager's gain-to-pain ratio.
func (r *Ratios) GainToPainRatio() *float64 {
	if !r.isPrimed() {
		return nil
	}
	if r.requiredLPM1 == nil || *r.requiredLPM1 == 0 {
		return nil
	}
	if r.returnsMean == nil {
		return nil
	}
	v := *r.returnsMean / *r.requiredLPM1
	return &v
}

// RiskOfRuin calculates the risk of ruin.
func (r *Ratios) RiskOfRuin() *float64 {
	if !r.isPrimed() {
		return nil
	}
	if r.winRate == nil {
		return nil
	}
	wr := *r.winRate
	v := math.Pow((1-wr)/(1+wr), float64(len(r.windowReturns())))
	return &v
}

// RiskReturnRatio calculates the return/risk ratio (Sharpe without risk-free rate).
func (r *Ratios) RiskReturnRatio() *float64 {
	if !r.isPrimed() {
		return nil
	}
	if r.returnsMean == nil || r.returnsStd == nil || *r.returnsStd == 0 {
		return nil
	}
	v := *r.returnsMean / *r.returnsStd
	return &v
}

// ---------- helper functions ----------

func sliceSum(s []float64) float64 {
	sum := 0.0
	for _, v := range s {
		sum += v
	}
	return sum
}

func sliceMean(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}
	return sliceSum(s) / float64(len(s))
}

// sliceStdDdof1 computes sample standard deviation (ddof=1).
func sliceStdDdof1(s []float64, mean float64) float64 {
	if len(s) < 2 {
		return 0
	}
	sum := 0.0
	for _, v := range s {
		d := v - mean
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(s)-1))
}

func sliceSumPow(s []float64, power float64) float64 {
	sum := 0.0
	for _, v := range s {
		sum += math.Pow(v, power)
	}
	return sum
}

func filterNonZero(s []float64) []float64 {
	var result []float64
	for _, v := range s {
		if v != 0 {
			result = append(result, v)
		}
	}
	return result
}

func filterPositive(s []float64) []float64 {
	var result []float64
	for _, v := range s {
		if v > 0 {
			result = append(result, v)
		}
	}
	return result
}

func filterNegative(s []float64) []float64 {
	var result []float64
	for _, v := range s {
		if v < 0 {
			result = append(result, v)
		}
	}
	return result
}

// populationSkewness computes the population skewness (matching scipy.stats.skew(bias=True)).
// Formula: m3 / m2^(3/2) where m2 and m3 are central moments.
func populationSkewness(s []float64) float64 {
	n := float64(len(s))
	mean := sliceMean(s)
	m2 := 0.0
	m3 := 0.0
	for _, v := range s {
		d := v - mean
		m2 += d * d
		m3 += d * d * d
	}
	m2 /= n
	m3 /= n
	if m2 == 0 {
		return 0
	}
	return m3 / math.Pow(m2, 1.5)
}

// populationExcessKurtosis computes the population excess kurtosis
// (matching scipy.stats.kurtosis(bias=True, fisher=True)).
// Formula: m4/m2^2 - 3 where m2 and m4 are central moments.
func populationExcessKurtosis(s []float64) float64 {
	n := float64(len(s))
	mean := sliceMean(s)
	m2 := 0.0
	m4 := 0.0
	for _, v := range s {
		d := v - mean
		d2 := d * d
		m2 += d2
		m4 += d2 * d2
	}
	m2 /= n
	m4 /= n
	if m2 == 0 {
		return 0
	}
	return m4/(m2*m2) - 3
}
