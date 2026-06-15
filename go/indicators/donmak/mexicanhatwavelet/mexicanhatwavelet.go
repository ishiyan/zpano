package mexicanhatwavelet

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// Preset dilation values (a_f) for the three standard bands (Table 5.2).
const (
	dilationHigh = 1.483  // omega_0 = 1.3558 rad, period ~ 4.63 bars
	dilationMid  = 4.048  // omega_0 = 0.4670 rad, period ~ 13.45 bars
	dilationLow  = 15.97  // omega_0 = 0.1156 rad, period ~ 54.35 bars
)

// dilationFromPeriod computes dilation a_f from a desired center period in bars (Eq 5.11).
func dilationFromPeriod(period float64) (float64, error) {
	omega0 := 2.0 * math.Pi / period
	twoOverA := 1.091*omega0 - 0.071*omega0*omega0

	if twoOverA <= 0.0 {
		return 0, fmt.Errorf(
			"invalid mexican hat wavelet parameters: period is too large for the fitting formula (2/a <= 0)")
	}

	return 2.0 / twoOverA, nil
}

// computeCoefficients computes normalized Mexican Hat wavelet FIR coefficients
// for dilation a_f. psi(t) = (1 - 2*t^2) * exp(-t^2); h(n) = psi(n / a_f) for
// n = 0..K, K = 4 * round(a_f), normalized by 0.488 + 0.646*a_f + 0.0001*a_f^2.
func computeCoefficients(aF float64) []float64 {
	k := 4 * int(math.RoundToEven(aF))
	if k < 1 {
		k = 1
	}

	norm := 0.488 + 0.646*aF + 0.0001*aF*aF

	coeffs := make([]float64, 0, k+1)
	for n := 0; n <= k; n++ {
		t := float64(n) / aF
		t2 := t * t
		hN := (1.0 - 2.0*t2) * math.Exp(-t2)
		coeffs = append(coeffs, hN/norm)
	}

	return coeffs
}

// MexicanHatWavelet is Don Mak's Mexican Hat Wavelet (MHW) bandpass filter.
//
// It is a causal bandpass FIR filter derived from the Mexican Hat wavelet (the
// second derivative of a Gaussian), decomposing price into frequency bands with
// zero phase shift at the filter's center frequency.
//
// Reference:
//
// Mak, Don K. (2006). Mathematical Techniques in Financial Market Trading. Ch 5.
type MexicanHatWavelet struct {
	mu sync.RWMutex

	core.LineIndicator

	coefficients []float64
	numTaps      int

	buffer []float64
	count  int

	primed bool
}

// NewMexicanHatWavelet returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen,cyclop,gocognit
func NewMexicanHatWavelet(p *Params) (*MexicanHatWavelet, error) {
	const (
		invalid = "invalid mexican hat wavelet parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
	)

	var (
		aF  float64
		cfg string
	)

	switch p.Band {
	case BandHigh:
		aF, cfg = dilationHigh, "high"
	case BandMid:
		aF, cfg = dilationMid, "mid"
	case BandLow:
		aF, cfg = dilationLow, "low"
	case BandCustom:
		hasDilation := p.Dilation != 0.0
		hasPeriod := p.Period != 0.0

		if hasDilation && hasPeriod {
			return nil, fmt.Errorf(fmts, invalid, "provide only one of dilation or period, not both")
		}

		if !hasDilation && !hasPeriod {
			return nil, fmt.Errorf(fmts, invalid, "band=custom requires either dilation or period")
		}

		if hasPeriod {
			if p.Period <= 2.0 {
				return nil, fmt.Errorf(fmts, invalid, "period must be > 2")
			}

			var err error
			if aF, err = dilationFromPeriod(p.Period); err != nil {
				return nil, err
			}

			cfg = fmt.Sprintf("p%.2f", p.Period)
		} else {
			if p.Dilation <= 0.0 {
				return nil, fmt.Errorf(fmts, invalid, "dilation must be > 0")
			}

			aF = p.Dilation
			cfg = fmt.Sprintf("d%.2f", p.Dilation)
		}
	default:
		return nil, fmt.Errorf(fmts, invalid, "unknown band")
	}

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

	mnemonic := fmt.Sprintf("mhw(%s%s)", cfg, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Mexican hat wavelet " + mnemonic

	coefficients := computeCoefficients(aF)

	mhw := &MexicanHatWavelet{
		coefficients: coefficients,
		numTaps:      len(coefficients),
		buffer:       make([]float64, len(coefficients)),
	}

	mhw.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, mhw.Update)

	return mhw, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *MexicanHatWavelet) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *MexicanHatWavelet) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.MexicanHatWavelet,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample value and returns the filter output.
func (s *MexicanHatWavelet) Update(sample float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Shift buffer right and insert the new price at position 0.
	for i := s.numTaps - 1; i > 0; i-- {
		s.buffer[i] = s.buffer[i-1]
	}

	s.buffer[0] = sample
	s.count++

	if s.count < s.numTaps {
		s.primed = false

		return math.NaN()
	}

	// FIR convolution: y = sum(coefficients[k] * buffer[k]).
	y := 0.0
	for k := 0; k < s.numTaps; k++ {
		y += s.coefficients[k] * s.buffer[k]
	}

	s.primed = true

	return y
}
