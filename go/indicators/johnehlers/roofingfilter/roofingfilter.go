package roofingfilter

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// RoofingFilter is Ehler's Roofing Filter described in Ehler's book
// "Cycle Analytics for Traders" (2013).
//
// The Roofing Filter is comprised of a high-pass filter and a Super Smoother.
// Given the longest (Λ) and the shortest (λ) cycle periods in bars,
// the high-pass filter passes cyclic components whose periods are shorter than the longest one,
// and the Super Smoother filter attenuates cycle periods shorter than the shortest one.
//
// Three flavours are available:
//   - 1-pole high-pass filter (default)
//   - 1-pole high-pass filter with zero-mean
//   - 2-pole high-pass filter
//
// Reference:
//
// Ehlers, John F. (2013). Cycle Analytics for Traders. Wiley.
type RoofingFilter struct {
	mu sync.RWMutex
	core.LineIndicator
	hpCoeff1 float64
	hpCoeff2 float64
	hpCoeff3 float64
	ssCoeff1 float64
	ssCoeff2 float64
	ssCoeff3 float64

	hasTwoPole  bool
	hasZeroMean bool

	count           int
	samplePrevious  float64
	samplePrevious2 float64
	hpPrevious      float64
	hpPrevious2     float64
	ssPrevious      float64
	ssPrevious2     float64
	zmPrevious      float64
	value           float64
	primed          bool
}

// NewRoofingFilter returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewRoofingFilter(p *RoofingFilterParams) (*RoofingFilter, error) {
	const (
		invalid = "invalid roofing filter parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		minPer  = 2
	)

	shortest := p.ShortestCyclePeriod
	if shortest < minPer {
		return nil, fmt.Errorf(fmts, invalid, "shortest cycle period should be greater than 1")
	}

	longest := p.LongestCyclePeriod
	if longest <= shortest {
		return nil, fmt.Errorf(fmts, invalid, "longest cycle period should be greater than shortest")
	}

	// Resolve defaults for component functions.
	// RoofingFilter default bar component is MedianPrice, not ClosePrice.
	bc := p.BarComponent
	if bc == 0 {
		bc = entities.BarMedianPrice
	}

	qc := p.QuoteComponent
	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	tc := p.TradeComponent
	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	// Calculate high-pass filter coefficients.
	var hpCoeff1, hpCoeff2, hpCoeff3 float64

	if p.HasTwoPoleHighpassFilter {
		// 2-pole high-pass: α = (cos(π√2/Λ) + sin(π√2/Λ) - 1) / cos(π√2/Λ)
		angle := math.Sqrt2 / 2 * 2 * math.Pi / float64(longest) // = π√2/Λ
		cosAngle := math.Cos(angle)
		alpha := (math.Sin(angle) + cosAngle - 1) / cosAngle
		beta := 1 - alpha/2
		hpCoeff1 = beta * beta
		beta2 := 1 - alpha
		hpCoeff2 = 2 * beta2
		hpCoeff3 = beta2 * beta2
	} else {
		// 1-pole high-pass: α = (cos(2π/Λ) + sin(2π/Λ) - 1) / cos(2π/Λ)
		angle := 2 * math.Pi / float64(longest)
		cosAngle := math.Cos(angle)
		alpha := (math.Sin(angle) + cosAngle - 1) / cosAngle
		hpCoeff1 = 1 - alpha/2
		hpCoeff2 = 1 - alpha
	}

	// Calculate super smoother coefficients.
	// Uses literal 1.414 (not math.Sqrt2) to match C# reference.
	beta := 1.414 * math.Pi / float64(shortest)
	alpha := math.Exp(-beta)
	ssCoeff2 := 2 * alpha * math.Cos(beta)
	ssCoeff3 := -alpha * alpha
	ssCoeff1 := (1 - ssCoeff2 - ssCoeff3) / 2

	// Build mnemonic.
	poles := 1
	if p.HasTwoPoleHighpassFilter {
		poles = 2
	}

	zm := ""
	if p.HasZeroMean && !p.HasTwoPoleHighpassFilter {
		zm = "zm"
	}

	mnemonic := fmt.Sprintf("roof%dhp%s(%d, %d%s)", poles, zm, shortest, longest, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Roofing Filter " + mnemonic

	rf := &RoofingFilter{
		hpCoeff1:    hpCoeff1,
		hpCoeff2:    hpCoeff2,
		hpCoeff3:    hpCoeff3,
		ssCoeff1:    ssCoeff1,
		ssCoeff2:    ssCoeff2,
		ssCoeff3:    ssCoeff3,
		hasTwoPole:  p.HasTwoPoleHighpassFilter,
		hasZeroMean: p.HasZeroMean && !p.HasTwoPoleHighpassFilter,
		value:       math.NaN(),
	}

	rf.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, rf.Update)

	return rf, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *RoofingFilter) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
func (s *RoofingFilter) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.RoofingFilter,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the roofing filter given the next sample.
func (s *RoofingFilter) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.hasTwoPole {
		return s.update2Pole(sample)
	}

	return s.update1Pole(sample)
}

func (s *RoofingFilter) update1Pole(sample float64) float64 {
	var hp, ss, zm float64

	if s.primed {
		hp = s.hpCoeff1*(sample-s.samplePrevious) + s.hpCoeff2*s.hpPrevious
		ss = s.ssCoeff1*(hp+s.hpPrevious) + s.ssCoeff2*s.ssPrevious + s.ssCoeff3*s.ssPrevious2

		if s.hasZeroMean {
			zm = s.hpCoeff1*(ss-s.ssPrevious) + s.hpCoeff2*s.zmPrevious
			s.value = zm
		} else {
			s.value = ss
		}
	} else {
		s.count++

		if s.count == 1 {
			hp = 0
			ss = 0
		} else {
			hp = s.hpCoeff1*(sample-s.samplePrevious) + s.hpCoeff2*s.hpPrevious
			ss = s.ssCoeff1*(hp+s.hpPrevious) + s.ssCoeff2*s.ssPrevious + s.ssCoeff3*s.ssPrevious2

			if s.hasZeroMean {
				zm = s.hpCoeff1*(ss-s.ssPrevious) + s.hpCoeff2*s.zmPrevious
				if s.count == 5 { //nolint:mnd
					s.primed = true
					s.value = zm
				}
			} else if s.count == 4 { //nolint:mnd
				s.primed = true
				s.value = ss
			}
		}
	}

	s.samplePrevious = sample
	s.hpPrevious = hp
	s.ssPrevious2 = s.ssPrevious
	s.ssPrevious = ss

	if s.hasZeroMean {
		s.zmPrevious = zm
	}

	return s.value
}

func (s *RoofingFilter) update2Pole(sample float64) float64 {
	var hp, ss float64

	if s.primed {
		hp = s.hpCoeff1*(sample-2*s.samplePrevious+s.samplePrevious2) +
			s.hpCoeff2*s.hpPrevious - s.hpCoeff3*s.hpPrevious2
		ss = s.ssCoeff1*(hp+s.hpPrevious) + s.ssCoeff2*s.ssPrevious + s.ssCoeff3*s.ssPrevious2
		s.value = ss
	} else {
		s.count++

		if s.count < 4 { //nolint:mnd
			hp = 0
			ss = 0
		} else {
			hp = s.hpCoeff1*(sample-2*s.samplePrevious+s.samplePrevious2) +
				s.hpCoeff2*s.hpPrevious - s.hpCoeff3*s.hpPrevious2
			ss = s.ssCoeff1*(hp+s.hpPrevious) + s.ssCoeff2*s.ssPrevious + s.ssCoeff3*s.ssPrevious2

			if s.count == 5 { //nolint:mnd
				s.primed = true
				s.value = ss
			}
		}
	}

	s.samplePrevious2 = s.samplePrevious
	s.samplePrevious = sample
	s.hpPrevious2 = s.hpPrevious
	s.hpPrevious = hp
	s.ssPrevious2 = s.ssPrevious
	s.ssPrevious = ss

	return s.value
}
