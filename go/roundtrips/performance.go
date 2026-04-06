package roundtrips

import (
	"math"
	"time"

	"portf_py/daycounting"
	"portf_py/daycounting/conventions"
)

// sliceMean computes the arithmetic mean of a float64 slice.
// Returns 0 for empty slices.
func sliceMean(s []float64) float64 {
	n := len(s)
	if n == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range s {
		sum += v
	}
	return sum / float64(n)
}

// sliceStdPop computes the population standard deviation (ddof=0).
// Returns 0 for empty slices.
func sliceStdPop(s []float64) float64 {
	n := len(s)
	if n == 0 {
		return 0
	}
	m := sliceMean(s)
	sum := 0.0
	for _, v := range s {
		d := v - m
		sum += d * d
	}
	return math.Sqrt(sum / float64(n))
}

// maxConsecutive counts the longest run of consecutive true values in a bool slice.
func maxConsecutive(bools []bool) int {
	maxStreak := 0
	current := 0
	for _, b := range bools {
		if b {
			current++
			if current > maxStreak {
				maxStreak = current
			}
		} else {
			current = 0
		}
	}
	return maxStreak
}

// RoundtripPerformance calculates roundtrip performance statistics.
type RoundtripPerformance struct {
	InitialBalance     float64
	AnnualRiskFreeRate float64
	AnnualTargetReturn float64
	DayCountConvention conventions.DayCountConvention

	Roundtrips                   []Roundtrip
	ReturnsOnInvestments         []float64
	SortinoDownsideReturns       []float64
	ReturnsOnInvestmentsAnnual   []float64
	SortinoDownsideReturnsAnnual []float64

	FirstTime          *time.Time
	LastTime           *time.Time
	MaxNetPnl          float64
	MaxDrawdown        float64
	MaxDrawdownPercent float64

	TotalCommission             float64
	GrossWinningCommission      float64
	GrossLoosingCommission      float64
	NetWinningCommission        float64
	NetLoosingCommission        float64
	GrossWinningLongCommission  float64
	GrossLoosingLongCommission  float64
	NetWinningLongCommission    float64
	NetLoosingLongCommission    float64
	GrossWinningShortCommission float64
	GrossLoosingShortCommission float64
	NetWinningShortCommission   float64
	NetLoosingShortCommission   float64

	netPnl               float64
	grossPnl             float64
	grossWinningPnl      float64
	grossLoosingPnl      float64
	netWinningPnl        float64
	netLoosingPnl        float64
	grossLongPnl         float64
	grossShortPnl        float64
	netLongPnl           float64
	netShortPnl          float64
	grossLongWinningPnl  float64
	grossLongLoosingPnl  float64
	netLongWinningPnl    float64
	netLongLoosingPnl    float64
	grossShortWinningPnl float64
	grossShortLoosingPnl float64
	netShortWinningPnl   float64
	netShortLoosingPnl   float64

	totalCount             int
	longCount              int
	shortCount             int
	grossWinningCount      int
	grossLoosingCount      int
	netWinningCount        int
	netLoosingCount        int
	grossLongWinningCount  int
	grossLongLoosingCount  int
	netLongWinningCount    int
	netLongLoosingCount    int
	grossShortWinningCount int
	grossShortLoosingCount int
	netShortWinningCount   int
	netShortLoosingCount   int

	durationSec                  float64
	durationSecLong              float64
	durationSecShort             float64
	durationSecGrossWinning      float64
	durationSecGrossLoosing      float64
	durationSecNetWinning        float64
	durationSecNetLoosing        float64
	durationSecGrossLongWinning  float64
	durationSecGrossLongLoosing  float64
	durationSecNetLongWinning    float64
	durationSecNetLongLoosing    float64
	durationSecGrossShortWinning float64
	durationSecGrossShortLoosing float64
	durationSecNetShortWinning   float64
	durationSecNetShortLoosing   float64
	totalDurationAnnualized      float64

	totalMae      float64
	totalMfe      float64
	totalEff      float64
	totalEffEntry float64
	totalEffExit  float64

	roiMean    *float64
	roiStd     *float64
	roiTdd     *float64
	roiannMean *float64
	roiannStd  *float64
	roiannTdd  *float64
}

// NewRoundtripPerformance creates a new RoundtripPerformance tracker.
func NewRoundtripPerformance(
	initialBalance float64,
	annualRiskFreeRate float64,
	annualTargetReturn float64,
	dayCountConvention conventions.DayCountConvention,
) *RoundtripPerformance {
	return &RoundtripPerformance{
		InitialBalance:     initialBalance,
		AnnualRiskFreeRate: annualRiskFreeRate,
		AnnualTargetReturn: annualTargetReturn,
		DayCountConvention: dayCountConvention,
	}
}

// Reset resets all state back to initial values.
func (p *RoundtripPerformance) Reset() {
	p.Roundtrips = nil
	p.ReturnsOnInvestments = nil
	p.SortinoDownsideReturns = nil
	p.ReturnsOnInvestmentsAnnual = nil
	p.SortinoDownsideReturnsAnnual = nil

	p.FirstTime = nil
	p.LastTime = nil
	p.MaxNetPnl = 0
	p.MaxDrawdown = 0
	p.MaxDrawdownPercent = 0

	p.TotalCommission = 0
	p.GrossWinningCommission = 0
	p.GrossLoosingCommission = 0
	p.NetWinningCommission = 0
	p.NetLoosingCommission = 0
	p.GrossWinningLongCommission = 0
	p.GrossLoosingLongCommission = 0
	p.NetWinningLongCommission = 0
	p.NetLoosingLongCommission = 0
	p.GrossWinningShortCommission = 0
	p.GrossLoosingShortCommission = 0
	p.NetWinningShortCommission = 0
	p.NetLoosingShortCommission = 0

	p.netPnl = 0
	p.grossPnl = 0
	p.grossWinningPnl = 0
	p.grossLoosingPnl = 0
	p.netWinningPnl = 0
	p.netLoosingPnl = 0
	p.grossLongPnl = 0
	p.grossShortPnl = 0
	p.netLongPnl = 0
	p.netShortPnl = 0
	p.grossLongWinningPnl = 0
	p.grossLongLoosingPnl = 0
	p.netLongWinningPnl = 0
	p.netLongLoosingPnl = 0
	p.grossShortWinningPnl = 0
	p.grossShortLoosingPnl = 0
	p.netShortWinningPnl = 0
	p.netShortLoosingPnl = 0

	p.totalCount = 0
	p.longCount = 0
	p.shortCount = 0
	p.grossWinningCount = 0
	p.grossLoosingCount = 0
	p.netWinningCount = 0
	p.netLoosingCount = 0
	p.grossLongWinningCount = 0
	p.grossLongLoosingCount = 0
	p.netLongWinningCount = 0
	p.netLongLoosingCount = 0
	p.grossShortWinningCount = 0
	p.grossShortLoosingCount = 0
	p.netShortWinningCount = 0
	p.netShortLoosingCount = 0

	p.durationSec = 0
	p.durationSecLong = 0
	p.durationSecShort = 0
	p.durationSecGrossWinning = 0
	p.durationSecGrossLoosing = 0
	p.durationSecNetWinning = 0
	p.durationSecNetLoosing = 0
	p.durationSecGrossLongWinning = 0
	p.durationSecGrossLongLoosing = 0
	p.durationSecNetLongWinning = 0
	p.durationSecNetLongLoosing = 0
	p.durationSecGrossShortWinning = 0
	p.durationSecGrossShortLoosing = 0
	p.durationSecNetShortWinning = 0
	p.durationSecNetShortLoosing = 0
	p.totalDurationAnnualized = 0

	p.totalMae = 0
	p.totalMfe = 0
	p.totalEff = 0
	p.totalEffEntry = 0
	p.totalEffExit = 0

	p.roiMean = nil
	p.roiStd = nil
	p.roiTdd = nil
	p.roiannMean = nil
	p.roiannStd = nil
	p.roiannTdd = nil
}

// AddRoundtrip adds a roundtrip to the performance tracker.
func (p *RoundtripPerformance) AddRoundtrip(rt Roundtrip) {
	p.Roundtrips = append(p.Roundtrips, rt)
	p.totalCount++
	comm := rt.Commission
	p.TotalCommission += comm
	secs := rt.Duration.Seconds()
	p.durationSec += secs
	p.totalMae += rt.MaximumAdverseExcursion
	p.totalMfe += rt.MaximumFavorableExcursion
	p.totalEff += rt.TotalEfficiency
	p.totalEffEntry += rt.EntryEfficiency
	p.totalEffExit += rt.ExitEfficiency

	netPnl := rt.NetPnl
	p.netPnl += netPnl
	if netPnl > 0 {
		p.netWinningCount++
		p.netWinningPnl += netPnl
		p.NetWinningCommission += comm
		p.durationSecNetWinning += secs
	} else if netPnl < 0 {
		p.netLoosingCount++
		p.netLoosingPnl += netPnl
		p.NetLoosingCommission += comm
		p.durationSecNetLoosing += secs
	}

	grossPnl := rt.GrossPnl
	p.grossPnl += grossPnl
	if grossPnl > 0 {
		p.grossWinningCount++
		p.grossWinningPnl += grossPnl
		p.GrossWinningCommission += comm
		p.durationSecGrossWinning += secs
	} else if grossPnl < 0 {
		p.grossLoosingCount++
		p.grossLoosingPnl += grossPnl
		p.GrossLoosingCommission += comm
		p.durationSecGrossLoosing += secs
	}

	if rt.Side == Long {
		p.grossLongPnl += grossPnl
		p.netLongPnl += netPnl
		p.longCount++
		p.durationSecLong += secs
		if grossPnl > 0 {
			p.grossLongWinningCount++
			p.grossLongWinningPnl += grossPnl
			p.GrossWinningLongCommission += comm
			p.durationSecGrossLongWinning += secs
		} else if grossPnl < 0 {
			p.grossLongLoosingCount++
			p.grossLongLoosingPnl += grossPnl
			p.GrossLoosingLongCommission += comm
			p.durationSecGrossLongLoosing += secs
		}
		if netPnl > 0 {
			p.netLongWinningCount++
			p.netLongWinningPnl += grossPnl // intentional: uses grossPnl
			p.NetWinningLongCommission += comm
			p.durationSecNetLongWinning += secs
		} else if netPnl < 0 {
			p.netLongLoosingCount++
			p.netLongLoosingPnl += grossPnl // intentional: uses grossPnl
			p.NetLoosingLongCommission += comm
			p.durationSecNetLongLoosing += secs
		}
	} else {
		p.grossShortPnl += grossPnl
		p.netShortPnl += netPnl
		p.shortCount++
		p.durationSecShort += secs
		if grossPnl > 0 {
			p.grossShortWinningCount++
			p.grossShortWinningPnl += grossPnl
			p.GrossWinningShortCommission += comm
			p.durationSecGrossShortWinning += secs
		} else if grossPnl < 0 {
			p.grossShortLoosingCount++
			p.grossShortLoosingPnl += grossPnl
			p.GrossLoosingShortCommission += comm
			p.durationSecGrossShortLoosing += secs
		}
		if netPnl > 0 {
			p.netShortWinningCount++
			p.netShortWinningPnl += grossPnl // intentional: uses grossPnl
			p.NetWinningShortCommission += comm
			p.durationSecNetShortWinning += secs
		} else if netPnl < 0 {
			p.netShortLoosingCount++
			p.netShortLoosingPnl += grossPnl // intentional: uses grossPnl
			p.NetLoosingShortCommission += comm
			p.durationSecNetShortLoosing += secs
		}
	}

	// Update first/last times and duration
	changed := false
	if p.FirstTime == nil || p.FirstTime.After(rt.EntryTime) {
		t := rt.EntryTime
		p.FirstTime = &t
		changed = true
	}
	if p.LastTime == nil || p.LastTime.Before(rt.ExitTime) {
		t := rt.ExitTime
		p.LastTime = &t
		changed = true
	}
	if changed && p.FirstTime != nil && p.LastTime != nil {
		yf, err := daycounting.YearFrac(*p.FirstTime, *p.LastTime, p.DayCountConvention)
		if err == nil {
			p.totalDurationAnnualized = yf
		}
	}

	roi := netPnl / (rt.Quantity * rt.EntryPrice)
	p.ReturnsOnInvestments = append(p.ReturnsOnInvestments, roi)
	m := sliceMean(p.ReturnsOnInvestments)
	p.roiMean = &m
	s := sliceStdPop(p.ReturnsOnInvestments)
	p.roiStd = &s

	downside := roi - p.AnnualRiskFreeRate
	if downside < 0 {
		p.SortinoDownsideReturns = append(p.SortinoDownsideReturns, downside)
		// TDD = sqrt(mean(power(downsides, 2)))
		sumSq := 0.0
		for _, v := range p.SortinoDownsideReturns {
			sumSq += v * v
		}
		tdd := math.Sqrt(sumSq / float64(len(p.SortinoDownsideReturns)))
		p.roiTdd = &tdd
	}

	// Calculate annualized ROI
	yf, err := daycounting.YearFrac(rt.EntryTime, rt.ExitTime, p.DayCountConvention)
	if err == nil && yf != 0 {
		roiann := roi / yf
		p.ReturnsOnInvestmentsAnnual = append(p.ReturnsOnInvestmentsAnnual, roiann)
		m := sliceMean(p.ReturnsOnInvestmentsAnnual)
		p.roiannMean = &m
		s := sliceStdPop(p.ReturnsOnInvestmentsAnnual)
		p.roiannStd = &s

		downsideAnn := roiann - p.AnnualRiskFreeRate
		if downsideAnn < 0 {
			p.SortinoDownsideReturnsAnnual = append(p.SortinoDownsideReturnsAnnual, downsideAnn)
			sumSq := 0.0
			for _, v := range p.SortinoDownsideReturnsAnnual {
				sumSq += v * v
			}
			tdd := math.Sqrt(sumSq / float64(len(p.SortinoDownsideReturnsAnnual)))
			p.roiannTdd = &tdd
		}
	}

	// Calculate max drawdown
	if p.MaxNetPnl < p.netPnl {
		p.MaxNetPnl = p.netPnl
	}
	dd := p.MaxNetPnl - p.netPnl
	if p.MaxDrawdown < dd {
		p.MaxDrawdown = dd
		p.MaxDrawdownPercent = p.MaxDrawdown / (p.InitialBalance + p.MaxNetPnl)
	}
}

// --- ROI statistics ---

// RoiMean returns the mean value for returns on investments.
func (p *RoundtripPerformance) RoiMean() *float64 { return p.roiMean }

// RoiStd returns the standard deviation over returns on investments.
func (p *RoundtripPerformance) RoiStd() *float64 { return p.roiStd }

// RoiTdd returns the target downside deviation over returns on investments.
func (p *RoundtripPerformance) RoiTdd() *float64 { return p.roiTdd }

// RoiannMean returns the mean value for annualized returns on investments.
func (p *RoundtripPerformance) RoiannMean() *float64 { return p.roiannMean }

// RoiannStd returns the standard deviation over annualized returns on investments.
func (p *RoundtripPerformance) RoiannStd() *float64 { return p.roiannStd }

// RoiannTdd returns the target downside deviation over annualized returns on investments.
func (p *RoundtripPerformance) RoiannTdd() *float64 { return p.roiannTdd }

// --- Risk-adjusted ratios ---

// SharpeRatio returns the Sharpe ratio over returns on investments.
func (p *RoundtripPerformance) SharpeRatio() *float64 {
	if p.roiMean == nil || p.roiStd == nil || *p.roiStd == 0 {
		return nil
	}
	v := *p.roiMean / *p.roiStd
	return &v
}

// SharpeRatioAnnual returns the Sharpe ratio over annualized returns on investments.
func (p *RoundtripPerformance) SharpeRatioAnnual() *float64 {
	if p.roiannMean == nil || p.roiannStd == nil || *p.roiannStd == 0 {
		return nil
	}
	v := *p.roiannMean / *p.roiannStd
	return &v
}

// SortinoRatio returns the Sortino ratio over returns on investments.
func (p *RoundtripPerformance) SortinoRatio() *float64 {
	if p.roiMean == nil || p.roiTdd == nil || *p.roiTdd == 0 {
		return nil
	}
	v := (*p.roiMean - p.AnnualRiskFreeRate) / *p.roiTdd
	return &v
}

// SortinoRatioAnnual returns the Sortino ratio over annualized returns on investments.
func (p *RoundtripPerformance) SortinoRatioAnnual() *float64 {
	if p.roiannMean == nil || p.roiannTdd == nil || *p.roiannTdd == 0 {
		return nil
	}
	v := (*p.roiannMean - p.AnnualRiskFreeRate) / *p.roiannTdd
	return &v
}

// CalmarRatio returns the Calmar ratio over returns on investments.
func (p *RoundtripPerformance) CalmarRatio() *float64 {
	if p.roiMean == nil || p.MaxDrawdownPercent == 0 {
		return nil
	}
	v := *p.roiMean / p.MaxDrawdownPercent
	return &v
}

// CalmarRatioAnnual returns the Calmar ratio over annualized returns on investments.
func (p *RoundtripPerformance) CalmarRatioAnnual() *float64 {
	if p.roiannMean == nil || p.MaxDrawdownPercent == 0 {
		return nil
	}
	v := *p.roiannMean / p.MaxDrawdownPercent
	return &v
}

// --- Rate of return ---

// RateOfReturn returns the rate of return.
func (p *RoundtripPerformance) RateOfReturn() *float64 {
	if p.InitialBalance == 0 {
		return nil
	}
	v := p.netPnl / p.InitialBalance
	return &v
}

// RateOfReturnAnnual returns the annualized rate of return.
func (p *RoundtripPerformance) RateOfReturnAnnual() *float64 {
	if p.totalDurationAnnualized == 0 || p.InitialBalance == 0 {
		return nil
	}
	v := (p.netPnl / p.InitialBalance) / p.totalDurationAnnualized
	return &v
}

// RecoveryFactor returns the recovery factor.
func (p *RoundtripPerformance) RecoveryFactor() *float64 {
	rorann := p.RateOfReturnAnnual()
	if rorann == nil || p.MaxDrawdownPercent == 0 {
		return nil
	}
	v := *rorann / p.MaxDrawdownPercent
	return &v
}

// --- Profit ratios ---

// GrossProfitRatio returns the PnL ratio of gross winning over gross loosing roundtrips.
func (p *RoundtripPerformance) GrossProfitRatio() *float64 {
	if p.grossLoosingPnl == 0 {
		return nil
	}
	v := math.Abs(p.grossWinningPnl / p.grossLoosingPnl)
	return &v
}

// NetProfitRatio returns the PnL ratio of net winning over net loosing roundtrips.
func (p *RoundtripPerformance) NetProfitRatio() *float64 {
	if p.netLoosingPnl == 0 {
		return nil
	}
	v := math.Abs(p.netWinningPnl / p.netLoosingPnl)
	return &v
}

// GrossProfitLongRatio returns the PnL ratio of long gross winning over long gross loosing.
func (p *RoundtripPerformance) GrossProfitLongRatio() *float64 {
	if p.grossLongLoosingPnl == 0 {
		return nil
	}
	v := math.Abs(p.grossLongWinningPnl / p.grossLongLoosingPnl)
	return &v
}

// NetProfitLongRatio returns the PnL ratio of long net winning over long net loosing.
func (p *RoundtripPerformance) NetProfitLongRatio() *float64 {
	if p.netLongLoosingPnl == 0 {
		return nil
	}
	v := math.Abs(p.netLongWinningPnl / p.netLongLoosingPnl)
	return &v
}

// GrossProfitShortRatio returns the PnL ratio of short gross winning over short gross loosing.
func (p *RoundtripPerformance) GrossProfitShortRatio() *float64 {
	if p.grossShortLoosingPnl == 0 {
		return nil
	}
	v := math.Abs(p.grossShortWinningPnl / p.grossShortLoosingPnl)
	return &v
}

// NetProfitShortRatio returns the PnL ratio of short net winning over short net loosing.
func (p *RoundtripPerformance) NetProfitShortRatio() *float64 {
	if p.netShortLoosingPnl == 0 {
		return nil
	}
	v := math.Abs(p.netShortWinningPnl / p.netShortLoosingPnl)
	return &v
}

// --- Counts ---

func (p *RoundtripPerformance) TotalCount() int             { return p.totalCount }
func (p *RoundtripPerformance) LongCount() int              { return p.longCount }
func (p *RoundtripPerformance) ShortCount() int             { return p.shortCount }
func (p *RoundtripPerformance) GrossWinningCount() int      { return p.grossWinningCount }
func (p *RoundtripPerformance) GrossLoosingCount() int      { return p.grossLoosingCount }
func (p *RoundtripPerformance) NetWinningCount() int        { return p.netWinningCount }
func (p *RoundtripPerformance) NetLoosingCount() int        { return p.netLoosingCount }
func (p *RoundtripPerformance) GrossLongWinningCount() int  { return p.grossLongWinningCount }
func (p *RoundtripPerformance) GrossLongLoosingCount() int  { return p.grossLongLoosingCount }
func (p *RoundtripPerformance) NetLongWinningCount() int    { return p.netLongWinningCount }
func (p *RoundtripPerformance) NetLongLoosingCount() int    { return p.netLongLoosingCount }
func (p *RoundtripPerformance) GrossShortWinningCount() int { return p.grossShortWinningCount }
func (p *RoundtripPerformance) GrossShortLoosingCount() int { return p.grossShortLoosingCount }
func (p *RoundtripPerformance) NetShortWinningCount() int   { return p.netShortWinningCount }
func (p *RoundtripPerformance) NetShortLoosingCount() int   { return p.netShortLoosingCount }

// --- Win/loss ratios ---

func (p *RoundtripPerformance) GrossWinningRatio() float64 {
	if p.totalCount > 0 {
		return float64(p.grossWinningCount) / float64(p.totalCount)
	}
	return 0.0
}
func (p *RoundtripPerformance) GrossLoosingRatio() float64 {
	if p.totalCount > 0 {
		return float64(p.grossLoosingCount) / float64(p.totalCount)
	}
	return 0.0
}
func (p *RoundtripPerformance) NetWinningRatio() float64 {
	if p.totalCount > 0 {
		return float64(p.netWinningCount) / float64(p.totalCount)
	}
	return 0.0
}
func (p *RoundtripPerformance) NetLoosingRatio() float64 {
	if p.totalCount > 0 {
		return float64(p.netLoosingCount) / float64(p.totalCount)
	}
	return 0.0
}
func (p *RoundtripPerformance) GrossLongWinningRatio() float64 {
	if p.longCount > 0 {
		return float64(p.grossLongWinningCount) / float64(p.longCount)
	}
	return 0.0
}
func (p *RoundtripPerformance) GrossLongLoosingRatio() float64 {
	if p.longCount > 0 {
		return float64(p.grossLongLoosingCount) / float64(p.longCount)
	}
	return 0.0
}
func (p *RoundtripPerformance) NetLongWinningRatio() float64 {
	if p.longCount > 0 {
		return float64(p.netLongWinningCount) / float64(p.longCount)
	}
	return 0.0
}
func (p *RoundtripPerformance) NetLongLoosingRatio() float64 {
	if p.longCount > 0 {
		return float64(p.netLongLoosingCount) / float64(p.longCount)
	}
	return 0.0
}
func (p *RoundtripPerformance) GrossShortWinningRatio() float64 {
	if p.shortCount > 0 {
		return float64(p.grossShortWinningCount) / float64(p.shortCount)
	}
	return 0.0
}
func (p *RoundtripPerformance) GrossShortLoosingRatio() float64 {
	if p.shortCount > 0 {
		return float64(p.grossShortLoosingCount) / float64(p.shortCount)
	}
	return 0.0
}
func (p *RoundtripPerformance) NetShortWinningRatio() float64 {
	if p.shortCount > 0 {
		return float64(p.netShortWinningCount) / float64(p.shortCount)
	}
	return 0.0
}
func (p *RoundtripPerformance) NetShortLoosingRatio() float64 {
	if p.shortCount > 0 {
		return float64(p.netShortLoosingCount) / float64(p.shortCount)
	}
	return 0.0
}

// --- PnL totals ---

func (p *RoundtripPerformance) TotalGrossPnl() float64        { return p.grossPnl }
func (p *RoundtripPerformance) TotalNetPnl() float64          { return p.netPnl }
func (p *RoundtripPerformance) WinningGrossPnl() float64      { return p.grossWinningPnl }
func (p *RoundtripPerformance) LoosingGrossPnl() float64      { return p.grossLoosingPnl }
func (p *RoundtripPerformance) WinningNetPnl() float64        { return p.netWinningPnl }
func (p *RoundtripPerformance) LoosingNetPnl() float64        { return p.netLoosingPnl }
func (p *RoundtripPerformance) WinningGrossLongPnl() float64  { return p.grossLongWinningPnl }
func (p *RoundtripPerformance) LoosingGrossLongPnl() float64  { return p.grossLongLoosingPnl }
func (p *RoundtripPerformance) WinningNetLongPnl() float64    { return p.netLongWinningPnl }
func (p *RoundtripPerformance) LoosingNetLongPnl() float64    { return p.netLongLoosingPnl }
func (p *RoundtripPerformance) WinningGrossShortPnl() float64 { return p.grossShortWinningPnl }
func (p *RoundtripPerformance) LoosingGrossShortPnl() float64 { return p.grossShortLoosingPnl }
func (p *RoundtripPerformance) WinningNetShortPnl() float64   { return p.netShortWinningPnl }
func (p *RoundtripPerformance) LoosingNetShortPnl() float64   { return p.netShortLoosingPnl }

// --- Average PnL ---

func divOrZero(a float64, b int) float64 {
	if b > 0 {
		return a / float64(b)
	}
	return 0.0
}

func (p *RoundtripPerformance) AverageGrossPnl() float64 { return divOrZero(p.grossPnl, p.totalCount) }
func (p *RoundtripPerformance) AverageNetPnl() float64   { return divOrZero(p.netPnl, p.totalCount) }
func (p *RoundtripPerformance) AverageGrossLongPnl() float64 {
	return divOrZero(p.grossLongPnl, p.longCount)
}
func (p *RoundtripPerformance) AverageNetLongPnl() float64 {
	return divOrZero(p.netLongPnl, p.longCount)
}
func (p *RoundtripPerformance) AverageGrossShortPnl() float64 {
	return divOrZero(p.grossShortPnl, p.shortCount)
}
func (p *RoundtripPerformance) AverageNetShortPnl() float64 {
	return divOrZero(p.netShortPnl, p.shortCount)
}
func (p *RoundtripPerformance) AverageWinningGrossPnl() float64 {
	return divOrZero(p.grossWinningPnl, p.grossWinningCount)
}
func (p *RoundtripPerformance) AverageLoosingGrossPnl() float64 {
	return divOrZero(p.grossLoosingPnl, p.grossLoosingCount)
}
func (p *RoundtripPerformance) AverageWinningNetPnl() float64 {
	return divOrZero(p.netWinningPnl, p.netWinningCount)
}
func (p *RoundtripPerformance) AverageLoosingNetPnl() float64 {
	return divOrZero(p.netLoosingPnl, p.netLoosingCount)
}
func (p *RoundtripPerformance) AverageWinningGrossLongPnl() float64 {
	return divOrZero(p.grossLongWinningPnl, p.grossLongWinningCount)
}
func (p *RoundtripPerformance) AverageLoosingGrossLongPnl() float64 {
	return divOrZero(p.grossLongLoosingPnl, p.grossLongLoosingCount)
}
func (p *RoundtripPerformance) AverageWinningNetLongPnl() float64 {
	return divOrZero(p.netLongWinningPnl, p.netLongWinningCount)
}
func (p *RoundtripPerformance) AverageLoosingNetLongPnl() float64 {
	return divOrZero(p.netLongLoosingPnl, p.netLongLoosingCount)
}
func (p *RoundtripPerformance) AverageWinningGrossShortPnl() float64 {
	return divOrZero(p.grossShortWinningPnl, p.grossShortWinningCount)
}
func (p *RoundtripPerformance) AverageLoosingGrossShortPnl() float64 {
	return divOrZero(p.grossShortLoosingPnl, p.grossShortLoosingCount)
}
func (p *RoundtripPerformance) AverageWinningNetShortPnl() float64 {
	return divOrZero(p.netShortWinningPnl, p.netShortWinningCount)
}
func (p *RoundtripPerformance) AverageLoosingNetShortPnl() float64 {
	return divOrZero(p.netShortLoosingPnl, p.netShortLoosingCount)
}

// --- Average win/loss ratios ---

func (p *RoundtripPerformance) AverageGrossWinningLoosingRatio() float64 {
	w := p.AverageWinningGrossPnl()
	l := p.AverageLoosingGrossPnl()
	if l != 0 {
		return w / l
	}
	return 0.0
}
func (p *RoundtripPerformance) AverageNetWinningLoosingRatio() float64 {
	w := p.AverageWinningNetPnl()
	l := p.AverageLoosingNetPnl()
	if l != 0 {
		return w / l
	}
	return 0.0
}
func (p *RoundtripPerformance) AverageGrossWinningLoosingLongRatio() float64 {
	w := p.AverageWinningGrossLongPnl()
	l := p.AverageLoosingGrossLongPnl()
	if l != 0 {
		return w / l
	}
	return 0.0
}
func (p *RoundtripPerformance) AverageNetWinningLoosingLongRatio() float64 {
	w := p.AverageWinningNetLongPnl()
	l := p.AverageLoosingNetLongPnl()
	if l != 0 {
		return w / l
	}
	return 0.0
}
func (p *RoundtripPerformance) AverageGrossWinningLoosingShortRatio() float64 {
	w := p.AverageWinningGrossShortPnl()
	l := p.AverageLoosingGrossShortPnl()
	if l != 0 {
		return w / l
	}
	return 0.0
}
func (p *RoundtripPerformance) AverageNetWinningLoosingShortRatio() float64 {
	w := p.AverageWinningNetShortPnl()
	l := p.AverageLoosingNetShortPnl()
	if l != 0 {
		return w / l
	}
	return 0.0
}

// --- Profit PnL ratios ---

func (p *RoundtripPerformance) GrossProfitPnlRatio() float64 {
	if p.grossPnl != 0 {
		return p.grossWinningPnl / p.grossPnl
	}
	return 0.0
}
func (p *RoundtripPerformance) NetProfitPnlRatio() float64 {
	if p.netPnl != 0 {
		return p.netWinningPnl / p.netPnl
	}
	return 0.0
}
func (p *RoundtripPerformance) GrossProfitPnlLongRatio() float64 {
	if p.grossLongPnl != 0 {
		return p.grossLongWinningPnl / p.grossLongPnl
	}
	return 0.0
}
func (p *RoundtripPerformance) NetProfitPnlLongRatio() float64 {
	if p.netLongPnl != 0 {
		return p.netLongWinningPnl / p.netLongPnl
	}
	return 0.0
}
func (p *RoundtripPerformance) GrossProfitPnlShortRatio() float64 {
	if p.grossShortPnl != 0 {
		return p.grossShortWinningPnl / p.grossShortPnl
	}
	return 0.0
}
func (p *RoundtripPerformance) NetProfitPnlShortRatio() float64 {
	if p.netShortPnl != 0 {
		return p.netShortWinningPnl / p.netShortPnl
	}
	return 0.0
}

// --- Average duration ---

func (p *RoundtripPerformance) AverageDurationSeconds() float64 {
	return divOrZero(p.durationSec, p.totalCount)
}
func (p *RoundtripPerformance) AverageGrossWinningDurationSeconds() float64 {
	return divOrZero(p.durationSecGrossWinning, p.grossWinningCount)
}
func (p *RoundtripPerformance) AverageGrossLoosingDurationSeconds() float64 {
	return divOrZero(p.durationSecGrossLoosing, p.grossLoosingCount)
}
func (p *RoundtripPerformance) AverageNetWinningDurationSeconds() float64 {
	return divOrZero(p.durationSecNetWinning, p.netWinningCount)
}
func (p *RoundtripPerformance) AverageNetLoosingDurationSeconds() float64 {
	return divOrZero(p.durationSecNetLoosing, p.netLoosingCount)
}
func (p *RoundtripPerformance) AverageLongDurationSeconds() float64 {
	return divOrZero(p.durationSecLong, p.longCount)
}
func (p *RoundtripPerformance) AverageShortDurationSeconds() float64 {
	return divOrZero(p.durationSecShort, p.shortCount)
}
func (p *RoundtripPerformance) AverageGrossWinningLongDurationSeconds() float64 {
	return divOrZero(p.durationSecGrossLongWinning, p.grossLongWinningCount)
}
func (p *RoundtripPerformance) AverageGrossLoosingLongDurationSeconds() float64 {
	return divOrZero(p.durationSecGrossLongLoosing, p.grossLongLoosingCount)
}
func (p *RoundtripPerformance) AverageNetWinningLongDurationSeconds() float64 {
	return divOrZero(p.durationSecNetLongWinning, p.netLongWinningCount)
}
func (p *RoundtripPerformance) AverageNetLoosingLongDurationSeconds() float64 {
	return divOrZero(p.durationSecNetLongLoosing, p.netLongLoosingCount)
}
func (p *RoundtripPerformance) AverageGrossWinningShortDurationSeconds() float64 {
	return divOrZero(p.durationSecGrossShortWinning, p.grossShortWinningCount)
}
func (p *RoundtripPerformance) AverageGrossLoosingShortDurationSeconds() float64 {
	return divOrZero(p.durationSecGrossShortLoosing, p.grossShortLoosingCount)
}
func (p *RoundtripPerformance) AverageNetWinningShortDurationSeconds() float64 {
	return divOrZero(p.durationSecNetShortWinning, p.netShortWinningCount)
}
func (p *RoundtripPerformance) AverageNetLoosingShortDurationSeconds() float64 {
	return divOrZero(p.durationSecNetShortLoosing, p.netShortLoosingCount)
}

// --- Min/max duration ---

func (p *RoundtripPerformance) filterDurationSeconds(filter func(Roundtrip) bool) []float64 {
	var result []float64
	for _, r := range p.Roundtrips {
		if filter(r) {
			result = append(result, r.Duration.Seconds())
		}
	}
	return result
}

func minSlice(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}
	m := s[0]
	for _, v := range s[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func maxSlice(s []float64) float64 {
	if len(s) == 0 {
		return 0
	}
	m := s[0]
	for _, v := range s[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

func (p *RoundtripPerformance) MinimumDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(_ Roundtrip) bool { return true }))
}
func (p *RoundtripPerformance) MaximumDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(_ Roundtrip) bool { return true }))
}
func (p *RoundtripPerformance) MinimumLongDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.Side == Long }))
}
func (p *RoundtripPerformance) MaximumLongDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.Side == Long }))
}
func (p *RoundtripPerformance) MinimumShortDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.Side == Short }))
}
func (p *RoundtripPerformance) MaximumShortDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.Side == Short }))
}
func (p *RoundtripPerformance) MinimumGrossWinningDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl > 0 }))
}
func (p *RoundtripPerformance) MaximumGrossWinningDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl > 0 }))
}
func (p *RoundtripPerformance) MinimumGrossLoosingDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl < 0 }))
}
func (p *RoundtripPerformance) MaximumGrossLoosingDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl < 0 }))
}
func (p *RoundtripPerformance) MinimumNetWinningDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl > 0 }))
}
func (p *RoundtripPerformance) MaximumNetWinningDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl > 0 }))
}
func (p *RoundtripPerformance) MinimumNetLoosingDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl < 0 }))
}
func (p *RoundtripPerformance) MaximumNetLoosingDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl < 0 }))
}
func (p *RoundtripPerformance) MinimumGrossWinningLongDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl > 0 && r.Side == Long }))
}
func (p *RoundtripPerformance) MaximumGrossWinningLongDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl > 0 && r.Side == Long }))
}
func (p *RoundtripPerformance) MinimumGrossLoosingLongDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl < 0 && r.Side == Long }))
}
func (p *RoundtripPerformance) MaximumGrossLoosingLongDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl < 0 && r.Side == Long }))
}
func (p *RoundtripPerformance) MinimumNetWinningLongDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl > 0 && r.Side == Long }))
}
func (p *RoundtripPerformance) MaximumNetWinningLongDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl > 0 && r.Side == Long }))
}
func (p *RoundtripPerformance) MinimumNetLoosingLongDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl < 0 && r.Side == Long }))
}
func (p *RoundtripPerformance) MaximumNetLoosingLongDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl < 0 && r.Side == Long }))
}
func (p *RoundtripPerformance) MinimumGrossWinningShortDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl > 0 && r.Side == Short }))
}
func (p *RoundtripPerformance) MaximumGrossWinningShortDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl > 0 && r.Side == Short }))
}
func (p *RoundtripPerformance) MinimumGrossLoosingShortDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl < 0 && r.Side == Short }))
}
func (p *RoundtripPerformance) MaximumGrossLoosingShortDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.GrossPnl < 0 && r.Side == Short }))
}
func (p *RoundtripPerformance) MinimumNetWinningShortDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl > 0 && r.Side == Short }))
}
func (p *RoundtripPerformance) MaximumNetWinningShortDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl > 0 && r.Side == Short }))
}
func (p *RoundtripPerformance) MinimumNetLoosingShortDurationSeconds() float64 {
	return minSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl < 0 && r.Side == Short }))
}
func (p *RoundtripPerformance) MaximumNetLoosingShortDurationSeconds() float64 {
	return maxSlice(p.filterDurationSeconds(func(r Roundtrip) bool { return r.NetPnl < 0 && r.Side == Short }))
}

// --- MAE / MFE / efficiency ---

func (p *RoundtripPerformance) AverageMaximumAdverseExcursion() float64 {
	return divOrZero(p.totalMae, p.totalCount)
}
func (p *RoundtripPerformance) AverageMaximumFavorableExcursion() float64 {
	return divOrZero(p.totalMfe, p.totalCount)
}
func (p *RoundtripPerformance) AverageEntryEfficiency() float64 {
	return divOrZero(p.totalEffEntry, p.totalCount)
}
func (p *RoundtripPerformance) AverageExitEfficiency() float64 {
	return divOrZero(p.totalEffExit, p.totalCount)
}
func (p *RoundtripPerformance) AverageTotalEfficiency() float64 {
	return divOrZero(p.totalEff, p.totalCount)
}

// filtered average helper
func (p *RoundtripPerformance) filteredAvg(field func(Roundtrip) float64, filter func(Roundtrip) bool, count int) float64 {
	if count == 0 {
		return 0.0
	}
	sum := 0.0
	for _, r := range p.Roundtrips {
		if filter(r) {
			sum += field(r)
		}
	}
	return sum / float64(count)
}

func (p *RoundtripPerformance) AverageMaximumAdverseExcursionGrossWinning() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.MaximumAdverseExcursion }, func(r Roundtrip) bool { return r.GrossPnl > 0 }, p.grossWinningCount)
}
func (p *RoundtripPerformance) AverageMaximumAdverseExcursionGrossLoosing() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.MaximumAdverseExcursion }, func(r Roundtrip) bool { return r.GrossPnl < 0 }, p.grossLoosingCount)
}
func (p *RoundtripPerformance) AverageMaximumAdverseExcursionNetWinning() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.MaximumAdverseExcursion }, func(r Roundtrip) bool { return r.NetPnl > 0 }, p.netWinningCount)
}
func (p *RoundtripPerformance) AverageMaximumAdverseExcursionNetLoosing() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.MaximumAdverseExcursion }, func(r Roundtrip) bool { return r.NetPnl < 0 }, p.netLoosingCount)
}
func (p *RoundtripPerformance) AverageMaximumFavorableExcursionGrossWinning() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.MaximumFavorableExcursion }, func(r Roundtrip) bool { return r.GrossPnl > 0 }, p.grossWinningCount)
}
func (p *RoundtripPerformance) AverageMaximumFavorableExcursionGrossLoosing() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.MaximumFavorableExcursion }, func(r Roundtrip) bool { return r.GrossPnl < 0 }, p.grossLoosingCount)
}
func (p *RoundtripPerformance) AverageMaximumFavorableExcursionNetWinning() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.MaximumFavorableExcursion }, func(r Roundtrip) bool { return r.NetPnl > 0 }, p.netWinningCount)
}
func (p *RoundtripPerformance) AverageMaximumFavorableExcursionNetLoosing() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.MaximumFavorableExcursion }, func(r Roundtrip) bool { return r.NetPnl < 0 }, p.netLoosingCount)
}
func (p *RoundtripPerformance) AverageEntryEfficiencyGrossWinning() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.EntryEfficiency }, func(r Roundtrip) bool { return r.GrossPnl > 0 }, p.grossWinningCount)
}
func (p *RoundtripPerformance) AverageEntryEfficiencyGrossLoosing() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.EntryEfficiency }, func(r Roundtrip) bool { return r.GrossPnl < 0 }, p.grossLoosingCount)
}
func (p *RoundtripPerformance) AverageEntryEfficiencyNetWinning() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.EntryEfficiency }, func(r Roundtrip) bool { return r.NetPnl > 0 }, p.netWinningCount)
}
func (p *RoundtripPerformance) AverageEntryEfficiencyNetLoosing() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.EntryEfficiency }, func(r Roundtrip) bool { return r.NetPnl < 0 }, p.netLoosingCount)
}
func (p *RoundtripPerformance) AverageExitEfficiencyGrossWinning() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.ExitEfficiency }, func(r Roundtrip) bool { return r.GrossPnl > 0 }, p.grossWinningCount)
}
func (p *RoundtripPerformance) AverageExitEfficiencyGrossLoosing() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.ExitEfficiency }, func(r Roundtrip) bool { return r.GrossPnl < 0 }, p.grossLoosingCount)
}
func (p *RoundtripPerformance) AverageExitEfficiencyNetWinning() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.ExitEfficiency }, func(r Roundtrip) bool { return r.NetPnl > 0 }, p.netWinningCount)
}
func (p *RoundtripPerformance) AverageExitEfficiencyNetLoosing() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.ExitEfficiency }, func(r Roundtrip) bool { return r.NetPnl < 0 }, p.netLoosingCount)
}
func (p *RoundtripPerformance) AverageTotalEfficiencyGrossWinning() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.TotalEfficiency }, func(r Roundtrip) bool { return r.GrossPnl > 0 }, p.grossWinningCount)
}
func (p *RoundtripPerformance) AverageTotalEfficiencyGrossLoosing() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.TotalEfficiency }, func(r Roundtrip) bool { return r.GrossPnl < 0 }, p.grossLoosingCount)
}
func (p *RoundtripPerformance) AverageTotalEfficiencyNetWinning() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.TotalEfficiency }, func(r Roundtrip) bool { return r.NetPnl > 0 }, p.netWinningCount)
}
func (p *RoundtripPerformance) AverageTotalEfficiencyNetLoosing() float64 {
	return p.filteredAvg(func(r Roundtrip) float64 { return r.TotalEfficiency }, func(r Roundtrip) bool { return r.NetPnl < 0 }, p.netLoosingCount)
}

// --- Consecutive streaks ---

func (p *RoundtripPerformance) MaxConsecutiveGrossWinners() int {
	bools := make([]bool, len(p.Roundtrips))
	for i, r := range p.Roundtrips {
		bools[i] = r.GrossPnl > 0
	}
	return maxConsecutive(bools)
}
func (p *RoundtripPerformance) MaxConsecutiveGrossLoosers() int {
	bools := make([]bool, len(p.Roundtrips))
	for i, r := range p.Roundtrips {
		bools[i] = r.GrossPnl < 0
	}
	return maxConsecutive(bools)
}
func (p *RoundtripPerformance) MaxConsecutiveNetWinners() int {
	bools := make([]bool, len(p.Roundtrips))
	for i, r := range p.Roundtrips {
		bools[i] = r.NetPnl > 0
	}
	return maxConsecutive(bools)
}
func (p *RoundtripPerformance) MaxConsecutiveNetLoosers() int {
	bools := make([]bool, len(p.Roundtrips))
	for i, r := range p.Roundtrips {
		bools[i] = r.NetPnl < 0
	}
	return maxConsecutive(bools)
}
