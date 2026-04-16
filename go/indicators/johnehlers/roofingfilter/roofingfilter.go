package roofingfilter

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
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

	count       int
	samplePrev  float64
	samplePrev2 float64
	hpPrev      float64
	hpPrev2     float64
	ssPrev      float64
	ssPrev2     float64
	zmPrev      float64
	value       float64
	primed      bool
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
func (r *RoofingFilter) IsPrimed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.primed
}

// Metadata describes an output data of the indicator.
func (r *RoofingFilter) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.RoofingFilter,
		Mnemonic:    r.LineIndicator.Mnemonic,
		Description: r.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(RoofingFilterValue),
				Type:        outputs.ScalarType,
				Mnemonic:    r.LineIndicator.Mnemonic,
				Description: r.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the roofing filter given the next sample.
func (r *RoofingFilter) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.hasTwoPole {
		return r.update2Pole(sample)
	}

	return r.update1Pole(sample)
}

func (r *RoofingFilter) update1Pole(sample float64) float64 {
	var hp, ss, zm float64

	if r.primed {
		hp = r.hpCoeff1*(sample-r.samplePrev) + r.hpCoeff2*r.hpPrev
		ss = r.ssCoeff1*(hp+r.hpPrev) + r.ssCoeff2*r.ssPrev + r.ssCoeff3*r.ssPrev2

		if r.hasZeroMean {
			zm = r.hpCoeff1*(ss-r.ssPrev) + r.hpCoeff2*r.zmPrev
			r.value = zm
		} else {
			r.value = ss
		}
	} else {
		r.count++

		if r.count == 1 {
			hp = 0
			ss = 0
		} else {
			hp = r.hpCoeff1*(sample-r.samplePrev) + r.hpCoeff2*r.hpPrev
			ss = r.ssCoeff1*(hp+r.hpPrev) + r.ssCoeff2*r.ssPrev + r.ssCoeff3*r.ssPrev2

			if r.hasZeroMean {
				zm = r.hpCoeff1*(ss-r.ssPrev) + r.hpCoeff2*r.zmPrev
				if r.count == 5 { //nolint:mnd
					r.primed = true
					r.value = zm
				}
			} else if r.count == 4 { //nolint:mnd
				r.primed = true
				r.value = ss
			}
		}
	}

	r.samplePrev = sample
	r.hpPrev = hp
	r.ssPrev2 = r.ssPrev
	r.ssPrev = ss

	if r.hasZeroMean {
		r.zmPrev = zm
	}

	return r.value
}

func (r *RoofingFilter) update2Pole(sample float64) float64 {
	var hp, ss float64

	if r.primed {
		hp = r.hpCoeff1*(sample-2*r.samplePrev+r.samplePrev2) +
			r.hpCoeff2*r.hpPrev - r.hpCoeff3*r.hpPrev2
		ss = r.ssCoeff1*(hp+r.hpPrev) + r.ssCoeff2*r.ssPrev + r.ssCoeff3*r.ssPrev2
		r.value = ss
	} else {
		r.count++

		if r.count < 4 { //nolint:mnd
			hp = 0
			ss = 0
		} else {
			hp = r.hpCoeff1*(sample-2*r.samplePrev+r.samplePrev2) +
				r.hpCoeff2*r.hpPrev - r.hpCoeff3*r.hpPrev2
			ss = r.ssCoeff1*(hp+r.hpPrev) + r.ssCoeff2*r.ssPrev + r.ssCoeff3*r.ssPrev2

			if r.count == 5 { //nolint:mnd
				r.primed = true
				r.value = ss
			}
		}
	}

	r.samplePrev2 = r.samplePrev
	r.samplePrev = sample
	r.hpPrev2 = r.hpPrev
	r.hpPrev = hp
	r.ssPrev2 = r.ssPrev
	r.ssPrev = ss

	return r.value
}
