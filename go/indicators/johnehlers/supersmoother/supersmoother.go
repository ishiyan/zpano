package supersmoother

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// SuperSmoother is Ehler's two-pole Super Smoother (SS) described in Ehler's book
// "Cybernetic Analysis for Stocks and Futures" (2004)
// and in his presentation "Spectral Dilation" (2013)
//
// Given the shortest (λ) cycle period in bars, the Super Smoother filter
// attenuates cycle periods shorter than this shortest one.
//
//	β = √2·π / λ
//	α = exp(-β)
//	γ₂ = 2α·cos(β)
//	γ₃ = -α²
//	γ₁ = (1 - γ₂ - γ₃) / 2
//
//	SSᵢ = γ₁·(xᵢ + xᵢ₋₁) + γ₂·SSᵢ₋₁ + γ₃·SSᵢ₋₂
//
// The indicator is not primed during the first 2 updates.
//
// Reference:
//
// Ehlers, John F. (2004). Cybernetic Analysis for Stocks and Futures. Wiley. pp 201-205.
// Ehlers, John F. (2013). Spectral dilation: Presented to the MTA in March 2013. Retrieved from www.mesasoftware.com/seminars/SpectralDilation.pdf
type SuperSmoother struct {
	mu sync.RWMutex
	core.LineIndicator
	coeff1      float64
	coeff2      float64
	coeff3      float64
	count       int
	samplePrev  float64
	filterPrev  float64
	filterPrev2 float64
	value       float64
	primed      bool
}

// NewSuperSmoother returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop
func NewSuperSmoother(p *SuperSmootherParams) (*SuperSmoother, error) {
	const (
		invalid = "invalid super smoother parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "ss(%d%s)"
		minPer  = 2
	)

	period := p.ShortestCyclePeriod
	if period < minPer {
		return nil, fmt.Errorf(fmts, invalid, "shortest cycle period should be greater than 1")
	}

	// Resolve defaults for component functions.
	// SuperSmoother default bar component is MedianPrice, not ClosePrice.
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

	// Calculate coefficients.
	beta := math.Sqrt2 * math.Pi / float64(period)
	alpha := math.Exp(-beta)
	gamma2 := 2 * alpha * math.Cos(beta)
	gamma3 := -alpha * alpha
	gamma1 := (1 - gamma2 - gamma3) / 2

	// Build mnemonic using resolved components.
	mnemonic := fmt.Sprintf(fmtn, period, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Super Smoother " + mnemonic

	ss := &SuperSmoother{
		coeff1: gamma1,
		coeff2: gamma2,
		coeff3: gamma3,
		value:  math.NaN(),
	}

	ss.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, ss.Update)

	return ss, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *SuperSmoother) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the super smoother.
func (s *SuperSmoother) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.SuperSmoother,
		Mnemonic:    s.LineIndicator.Mnemonic,
		Description: s.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(SuperSmootherValue),
				Type:        outputs.ScalarType,
				Mnemonic:    s.LineIndicator.Mnemonic,
				Description: s.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the super smoother given the next sample.
//
// The indicator is not primed during the first 2 updates.
func (s *SuperSmoother) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		filter := s.coeff1*(sample+s.samplePrev) +
			s.coeff2*s.filterPrev + s.coeff3*s.filterPrev2
		s.value = filter
		s.samplePrev = sample
		s.filterPrev2 = s.filterPrev
		s.filterPrev = filter

		return s.value
	}

	s.count++

	if s.count == 1 {
		s.samplePrev = sample
		s.filterPrev = sample
		s.filterPrev2 = sample
	}

	filter := s.coeff1*(sample+s.samplePrev) +
		s.coeff2*s.filterPrev + s.coeff3*s.filterPrev2

	if s.count == 3 { //nolint:mnd
		s.primed = true
		s.value = filter
	}

	s.samplePrev = sample
	s.filterPrev2 = s.filterPrev
	s.filterPrev = filter

	return s.value
}
