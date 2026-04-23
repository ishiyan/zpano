// Package adaptivetrendandcyclefilter implements the Adaptive Trend and Cycle
// Filter (ATCF) suite by Vladimir Kravchuk.
//
// The suite is a bank of five Finite Impulse Response (FIR) filters applied
// to the same input series plus three composite outputs derived from them:
//
//   - FATL (Fast Adaptive Trend Line)       — 39-tap FIR.
//   - SATL (Slow Adaptive Trend Line)       — 65-tap FIR.
//   - RFTL (Reference Fast Trend Line)      — 44-tap FIR.
//   - RSTL (Reference Slow Trend Line)      — 91-tap FIR.
//   - RBCI (Range Bound Channel Index)      — 56-tap FIR.
//   - FTLM (Fast Trend Line Momentum)       = FATL − RFTL.
//   - STLM (Slow Trend Line Momentum)       = SATL − RSTL.
//   - PCCI (Perfect Commodity Channel Index)= input − FATL.
//
// Each FIR filter emits NaN until its own window fills. Indicator-level
// IsPrimed mirrors RSTL (the longest pole, 91 samples).
//
// Reference: Vladimir Kravchuk, "New adaptive method of following the
// tendency and market cycles", Currency Speculator magazine, 2000.
package adaptivetrendandcyclefilter

import (
	"fmt"
	"math"
	"sync"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
)

// firFilter is the internal FIR engine shared by all five ATCF lines.
//
// It holds a fixed window (length = len(coeffs)) and computes
// Σ window[i]·coeffs[i] on every update once primed. Before priming, value
// is NaN and incoming samples fill the window one slot at a time.
type firFilter struct {
	window []float64
	coeffs []float64
	count  int
	primed bool
	value  float64
}

func newFirFilter(coeffs []float64) *firFilter {
	return &firFilter{
		window: make([]float64, len(coeffs)),
		coeffs: coeffs,
		value:  math.NaN(),
	}
}

func (s *firFilter) isPrimed() bool { return s.primed }

func (s *firFilter) update(sample float64) float64 {
	if s.primed {
		copy(s.window, s.window[1:])
		s.window[len(s.window)-1] = sample

		sum := 0.0
		for i := range s.window {
			sum += s.window[i] * s.coeffs[i]
		}

		s.value = sum

		return s.value
	}

	s.window[s.count] = sample
	s.count++

	if s.count == len(s.window) {
		s.primed = true

		sum := 0.0
		for i := range s.window {
			sum += s.window[i] * s.coeffs[i]
		}

		s.value = sum
	}

	return s.value
}

// AdaptiveTrendAndCycleFilter is Vladimir Kravchuk's combined ATCF suite.
//
// It exposes eight scalar outputs (five FIR filters + three composites).
type AdaptiveTrendAndCycleFilter struct {
	mu sync.RWMutex

	mnemonic    string
	description string

	mnemonicFatl, descriptionFatl string
	mnemonicSatl, descriptionSatl string
	mnemonicRftl, descriptionRftl string
	mnemonicRstl, descriptionRstl string
	mnemonicRbci, descriptionRbci string
	mnemonicFtlm, descriptionFtlm string
	mnemonicStlm, descriptionStlm string
	mnemonicPcci, descriptionPcci string

	fatl *firFilter
	satl *firFilter
	rftl *firFilter
	rstl *firFilter
	rbci *firFilter

	ftlmValue float64
	stlmValue float64
	pcciValue float64

	barFunc   entities.BarFunc
	quoteFunc entities.QuoteFunc
	tradeFunc entities.TradeFunc
}

// NewAdaptiveTrendAndCycleFilterDefault returns an instance created with default parameters.
func NewAdaptiveTrendAndCycleFilterDefault() (*AdaptiveTrendAndCycleFilter, error) {
	return NewAdaptiveTrendAndCycleFilterParams(&Params{})
}

// NewAdaptiveTrendAndCycleFilterParams returns an instance created with the supplied parameters.
//
//nolint:funlen
func NewAdaptiveTrendAndCycleFilterParams(p *Params) (*AdaptiveTrendAndCycleFilter, error) {
	const (
		invalid = "invalid adaptive trend and cycle filter parameters"
		fmtw    = "%s: %w"

		fmtAll  = "atcf(%s)"
		fmtOne  = "%s(%s)"
		descAll = "Adaptive trend and cycle filter "
		descOne = "%s %s"
	)

	cfg := *p

	bc := cfg.BarComponent
	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	qc := cfg.QuoteComponent
	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	tc := cfg.TradeComponent
	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	barFunc, err := entities.BarComponentFunc(bc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	quoteFunc, err := entities.QuoteComponentFunc(qc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	tradeFunc, err := entities.TradeComponentFunc(tc)
	if err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	componentMnemonic := core.ComponentTripleMnemonic(bc, qc, tc)
	// componentMnemonic has a leading ", " when non-empty; strip it for the
	// per-component sub-mnemonics so they look like "fatl(hl/2)" instead of
	// "fatl(, hl/2)". Keep the top-level "atcf(%s)" format similar.
	topArg := ""
	subArg := ""

	if componentMnemonic != "" {
		topArg = componentMnemonic[2:] // skip leading ", "
		subArg = componentMnemonic[2:]
	}

	mnemonic := fmt.Sprintf(fmtAll, topArg)

	mkSub := func(name, full string) (string, string) {
		m := fmt.Sprintf(fmtOne, name, subArg)
		d := fmt.Sprintf(descOne, full, m)

		return m, d
	}

	mFatl, dFatl := mkSub("fatl", "Fast Adaptive Trend Line")
	mSatl, dSatl := mkSub("satl", "Slow Adaptive Trend Line")
	mRftl, dRftl := mkSub("rftl", "Reference Fast Trend Line")
	mRstl, dRstl := mkSub("rstl", "Reference Slow Trend Line")
	mRbci, dRbci := mkSub("rbci", "Range Bound Channel Index")
	mFtlm, dFtlm := mkSub("ftlm", "Fast Trend Line Momentum")
	mStlm, dStlm := mkSub("stlm", "Slow Trend Line Momentum")
	mPcci, dPcci := mkSub("pcci", "Perfect Commodity Channel Index")

	return &AdaptiveTrendAndCycleFilter{
		mnemonic:        mnemonic,
		description:     descAll + mnemonic,
		mnemonicFatl:    mFatl,
		descriptionFatl: dFatl,
		mnemonicSatl:    mSatl,
		descriptionSatl: dSatl,
		mnemonicRftl:    mRftl,
		descriptionRftl: dRftl,
		mnemonicRstl:    mRstl,
		descriptionRstl: dRstl,
		mnemonicRbci:    mRbci,
		descriptionRbci: dRbci,
		mnemonicFtlm:    mFtlm,
		descriptionFtlm: dFtlm,
		mnemonicStlm:    mStlm,
		descriptionStlm: dStlm,
		mnemonicPcci:    mPcci,
		descriptionPcci: dPcci,
		fatl:            newFirFilter(fatlCoefficients),
		satl:            newFirFilter(satlCoefficients),
		rftl:            newFirFilter(rftlCoefficients),
		rstl:            newFirFilter(rstlCoefficients),
		rbci:            newFirFilter(rbciCoefficients),
		ftlmValue:       math.NaN(),
		stlmValue:       math.NaN(),
		pcciValue:       math.NaN(),
		barFunc:         barFunc,
		quoteFunc:       quoteFunc,
		tradeFunc:       tradeFunc,
	}, nil
}

// IsPrimed indicates whether the indicator is primed.
//
// The indicator is primed when its longest pole, RSTL (91-tap FIR), is primed.
func (s *AdaptiveTrendAndCycleFilter) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.rstl.isPrimed()
}

// Metadata describes the output data of the indicator.
func (s *AdaptiveTrendAndCycleFilter) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.AdaptiveTrendAndCycleFilter,
		s.mnemonic,
		s.description,
		[]core.OutputText{
			{Mnemonic: s.mnemonicFatl, Description: s.descriptionFatl},
			{Mnemonic: s.mnemonicSatl, Description: s.descriptionSatl},
			{Mnemonic: s.mnemonicRftl, Description: s.descriptionRftl},
			{Mnemonic: s.mnemonicRstl, Description: s.descriptionRstl},
			{Mnemonic: s.mnemonicRbci, Description: s.descriptionRbci},
			{Mnemonic: s.mnemonicFtlm, Description: s.descriptionFtlm},
			{Mnemonic: s.mnemonicStlm, Description: s.descriptionStlm},
			{Mnemonic: s.mnemonicPcci, Description: s.descriptionPcci},
		},
	)
}

// Update feeds the next sample to all five FIR filters and recomputes the
// three composite values. Returns the 8 output scalars in enum order:
// FATL, SATL, RFTL, RSTL, RBCI, FTLM, STLM, PCCI.
//
// Each FIR filter emits NaN until its own window fills; composite values
// emit NaN until both their components are primed. NaN input leaves the
// internal state unchanged and all outputs are NaN.
func (s *AdaptiveTrendAndCycleFilter) Update(sample float64) (
	fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci float64,
) {
	if math.IsNaN(sample) {
		nan := math.NaN()

		return nan, nan, nan, nan, nan, nan, nan, nan
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fatl = s.fatl.update(sample)
	satl = s.satl.update(sample)
	rftl = s.rftl.update(sample)
	rstl = s.rstl.update(sample)
	rbci = s.rbci.update(sample)

	if s.fatl.isPrimed() && s.rftl.isPrimed() {
		s.ftlmValue = fatl - rftl
	}

	if s.satl.isPrimed() && s.rstl.isPrimed() {
		s.stlmValue = satl - rstl
	}

	if s.fatl.isPrimed() {
		s.pcciValue = sample - fatl
	}

	return fatl, satl, rftl, rstl, rbci, s.ftlmValue, s.stlmValue, s.pcciValue
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *AdaptiveTrendAndCycleFilter) UpdateScalar(sample *entities.Scalar) core.Output {
	return s.updateEntity(sample.Time, sample.Value)
}

// UpdateBar updates the indicator given the next bar sample.
func (s *AdaptiveTrendAndCycleFilter) UpdateBar(sample *entities.Bar) core.Output {
	return s.updateEntity(sample.Time, s.barFunc(sample))
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *AdaptiveTrendAndCycleFilter) UpdateQuote(sample *entities.Quote) core.Output {
	return s.updateEntity(sample.Time, s.quoteFunc(sample))
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *AdaptiveTrendAndCycleFilter) UpdateTrade(sample *entities.Trade) core.Output {
	return s.updateEntity(sample.Time, s.tradeFunc(sample))
}

func (s *AdaptiveTrendAndCycleFilter) updateEntity(t time.Time, sample float64) core.Output {
	const length = 8

	fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci := s.Update(sample)

	output := make([]any, length)
	output[0] = entities.Scalar{Time: t, Value: fatl}
	output[1] = entities.Scalar{Time: t, Value: satl}
	output[2] = entities.Scalar{Time: t, Value: rftl}
	output[3] = entities.Scalar{Time: t, Value: rstl}
	output[4] = entities.Scalar{Time: t, Value: rbci}
	output[5] = entities.Scalar{Time: t, Value: ftlm}
	output[6] = entities.Scalar{Time: t, Value: stlm}
	output[7] = entities.Scalar{Time: t, Value: pcci}

	return output
}
