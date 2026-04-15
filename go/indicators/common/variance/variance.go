package variance

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

// Variance computes the variance of the samples within a moving window of length ℓ:
//
//	σ² = (∑xᵢ² - (∑xᵢ)²/ℓ)/ℓ
//
// for the estimation of the population variance, or as:
//
//	σ² = (∑xᵢ² - (∑xᵢ)²/ℓ)/(ℓ-1)
//
// for the unbiased estimation of the sample variance, i={0,…,ℓ-1}.
type Variance struct {
	mu sync.RWMutex
	core.LineIndicator
	window           []float64
	windowSum        float64
	windowSquaredSum float64
	windowLength     int
	windowCount      int
	lastIndex        int
	primed           bool
	unbiased         bool
}

// NewVariance returns an instnce of the Variance indicator created using supplied parameters.
func NewVariance(p *VarianceParams) (*Variance, error) {
	const (
		invalid = "invalid variance parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "var.%c(%d%s)"
		minlen  = 2
	)

	length := p.Length
	if length < minlen {
		return nil, fmt.Errorf(fmts, invalid, "length should be greater than 1")
	}

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

	// Build mnemonic using resolved components — defaults are omitted by ComponentTripleMnemonic.
	var c byte = 'p'
	if p.IsUnbiased {
		c = 's'
	}

	mnemonic := fmt.Sprintf(fmtn, c, length, core.ComponentTripleMnemonic(bc, qc, tc))

	var desc string
	if p.IsUnbiased {
		desc = "Unbiased estimation of the sample variance " + mnemonic
	} else {
		desc = "Estimation of the population variance " + mnemonic
	}

	v := &Variance{
		window:       make([]float64, length),
		windowLength: length,
		lastIndex:    length - 1,
		unbiased:     p.IsUnbiased,
	}

	v.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, v.Update)

	return v, nil
}

// IsPrimed indicates whether an indicator is primed.
func (v *Variance) IsPrimed() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return v.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the variance.
func (v *Variance) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.Variance,
		Mnemonic:    v.LineIndicator.Mnemonic,
		Description: v.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(VarianceValue),
				Type:        outputs.ScalarType,
				Mnemonic:    v.LineIndicator.Mnemonic,
				Description: v.LineIndicator.Description,
			},
		},
	}
}

// Update updates the value of the variance, σ², given the next sample.
//
// Depending on the isUnbiased, the value is the unbiased sample variance or the population variance.
//
// The indicator is not primed during the first ℓ-1 updates.
//
//nolint:funlen
func (v *Variance) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	var value float64

	temp := sample
	wlen := float64(v.windowLength)

	v.mu.Lock()
	defer v.mu.Unlock()

	//nolint:nestif
	if v.primed {
		v.windowSum += temp
		temp *= temp
		v.windowSquaredSum += temp
		temp = v.window[0]
		v.windowSum -= temp
		temp *= temp
		v.windowSquaredSum -= temp

		if v.unbiased {
			temp = v.windowSum
			temp *= temp
			temp /= wlen
			value = v.windowSquaredSum - temp
			value /= float64(v.lastIndex)
		} else {
			temp = v.windowSum / wlen
			temp *= temp
			value = v.windowSquaredSum/wlen - temp
		}

		for i := 0; i < v.lastIndex; i++ {
			v.window[i] = v.window[i+1]
		}

		v.window[v.lastIndex] = sample
	} else {
		v.windowSum += temp
		v.window[v.windowCount] = temp
		temp *= temp
		v.windowSquaredSum += temp

		v.windowCount++
		if v.windowLength == v.windowCount {
			v.primed = true
			if v.unbiased {
				temp = v.windowSum
				temp *= temp
				temp /= wlen
				value = v.windowSquaredSum - temp
				value /= float64(v.lastIndex)
			} else {
				temp = v.windowSum / wlen
				temp *= temp
				value = v.windowSquaredSum/wlen - temp
			}
		} else {
			return math.NaN()
		}
	}

	return value
}
