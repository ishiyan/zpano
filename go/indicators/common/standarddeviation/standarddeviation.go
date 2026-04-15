package standarddeviation

//nolint: gofumpt
import (
	"fmt"
	"math"

	"zpano/entities"
	"zpano/indicators/common/variance"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// StandardDeviation computes the standard deviation of the samples within a moving window of length ℓ
// as a square root of variance:
//
//	σ² = (∑xᵢ² - (∑xᵢ)²/ℓ)/ℓ
//
// for the estimation of the population variance, or as:
//
//	σ² = (∑xᵢ² - (∑xᵢ)²/ℓ)/(ℓ-1)
//
// for the unbiased estimation of the sample variance, i={0,…,ℓ-1}.
type StandardDeviation struct {
	core.LineIndicator
	variance *variance.Variance
}

// NewStandardDeviation returns an instnce of the StandardDeviation indicator created using supplied parameters.
func NewStandardDeviation(p *StandardDeviationParams) (*StandardDeviation, error) {
	const (
		fmtn = "stdev.%c(%d%s)"
	)

	// Resolve defaults for component functions.
	// A zero value means "use default, don't show in mnemonic".
	bc := p.BarComponent
	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	qc := p.QuoteComponent
	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	tc := p.TradeComponent
	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	// Create the underlying variance indicator.
	vp := &variance.VarianceParams{
		Length:         p.Length,
		IsUnbiased:     p.IsUnbiased,
		BarComponent:   bc,
		QuoteComponent: qc,
		TradeComponent: tc,
	}

	v, err := variance.NewVariance(vp)
	if err != nil {
		return nil, err
	}

	var (
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, err
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, err
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, err
	}

	// Build mnemonic using resolved components — defaults are omitted by ComponentTripleMnemonic.
	var c byte = 'p'
	if p.IsUnbiased {
		c = 's'
	}

	mnemonic := fmt.Sprintf(fmtn, c, p.Length, core.ComponentTripleMnemonic(bc, qc, tc))

	var desc string
	if p.IsUnbiased {
		desc = "Standard deviation based on unbiased estimation of the sample variance " + mnemonic
	} else {
		desc = "Standard deviation based on estimation of the population variance " + mnemonic
	}

	sd := &StandardDeviation{
		variance: v,
	}

	sd.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, sd.Update)

	return sd, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *StandardDeviation) IsPrimed() bool {
	return s.variance.IsPrimed()
}

// Metadata describes an output data of the indicator.
func (s *StandardDeviation) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.StandardDeviation,
		Mnemonic:    s.LineIndicator.Mnemonic,
		Description: s.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(StandardDeviationValue),
				Type:        outputs.ScalarType,
				Mnemonic:    s.LineIndicator.Mnemonic,
				Description: s.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the standard deviation given the next sample.
//
// The indicator is not primed during the first ℓ-1 updates.
func (s *StandardDeviation) Update(sample float64) float64 {
	v := s.variance.Update(sample)
	if math.IsNaN(v) {
		return v
	}

	return math.Sqrt(v)
}
