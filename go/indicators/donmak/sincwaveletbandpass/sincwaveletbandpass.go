package sincwaveletbandpass

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// velocityTaps is the number of taps in the cubic velocity kernel.
const velocityTaps = 4

// velocityKernel is the cubic velocity kernel (PFD degree=3, order=1, smoothing=0).
var velocityKernel = [velocityTaps]float64{11.0 / 6.0, -3.0, 3.0 / 2.0, -1.0 / 3.0}

// bandParams returns the (omega0, omega1, numTaps) for a band.
func bandParams(band Band) (omega0, omega1 float64, numTaps int, ok bool) {
	switch band {
	case BandHigh:
		return math.Pi / 4, math.Pi / 8, 121, true
	case BandMid:
		return math.Pi / 8, math.Pi / 16, 121, true
	case BandLow:
		return math.Pi / 16, math.Pi / 32, 201, true
	case BandFull:
		return math.Pi / 4, math.Pi / 32, 201, true
	default:
		return 0, 0, 0, false
	}
}

// computeCoefficients computes sinc band-pass filter coefficients (difference of two sinc functions).
func computeCoefficients(omega0, omega1 float64, numTaps int) []float64 {
	coeffs := make([]float64, numTaps)
	coeffs[0] = (omega0 - omega1) / math.Pi

	for k := 1; k < numTaps; k++ {
		piK := math.Pi * float64(k)
		coeffs[k] = math.Sin(omega0*float64(k))/piK - math.Sin(omega1*float64(k))/piK
	}

	return coeffs
}

// SincWaveletBandpass is Don Mak's Sinc Wavelet Band-Pass (SWB) filter.
//
// It is a causal FIR band-pass filter derived from the sinc wavelet system,
// decomposing price into frequency bands (HIGH, MID, LOW, FULL). Optionally a
// cubic velocity kernel is applied to produce a momentum oscillator.
//
// Reference:
//
// Mak, D.K. (2003). The Science of Financial Market Trading. Ch 9, Appendix 7.
type SincWaveletBandpass struct {
	mu sync.RWMutex

	core.LineIndicator

	velocity     bool
	coefficients []float64
	numTaps      int

	priceBuffer []float64
	priceCount  int
	priceIndex  int

	velBuffer [velocityTaps]float64
	velCount  int
	velIndex  int

	primed bool
}

// NewSincWaveletBandpass returns an instance of the indicator created using supplied parameters.
//
//nolint:funlen
func NewSincWaveletBandpass(p *Params) (*SincWaveletBandpass, error) {
	const (
		invalid = "invalid sinc wavelet band-pass parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
	)

	omega0, omega1, numTaps, ok := bandParams(p.Band)
	if !ok {
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

	bandNames := map[Band]string{BandHigh: "high", BandMid: "mid", BandLow: "low", BandFull: "full"}
	cfg := bandNames[p.Band]
	if p.Velocity {
		cfg += ",v"
	}

	mnemonic := fmt.Sprintf("swb(%s%s)", cfg, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Sinc wavelet band-pass " + mnemonic

	swb := &SincWaveletBandpass{
		velocity:     p.Velocity,
		coefficients: computeCoefficients(omega0, omega1, numTaps),
		numTaps:      numTaps,
		priceBuffer:  make([]float64, numTaps),
	}

	swb.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, swb.Update)

	return swb, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *SincWaveletBandpass) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *SincWaveletBandpass) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.SincWaveletBandpass,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample value and returns the filter output.
func (s *SincWaveletBandpass) Update(sample float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store price in the ring buffer.
	s.priceBuffer[s.priceIndex] = sample
	s.priceIndex = (s.priceIndex + 1) % s.numTaps
	s.priceCount++

	if s.priceCount < s.numTaps {
		s.primed = false

		return math.NaN()
	}

	// Band-pass convolution: coefficients[k] multiplies the k-th most recent price.
	bpValue := 0.0
	idx := s.priceIndex - 1

	for k := 0; k < s.numTaps; k++ {
		bufIdx := ((idx % s.numTaps) + s.numTaps) % s.numTaps
		bpValue += s.coefficients[k] * s.priceBuffer[bufIdx]
		idx--
	}

	if !s.velocity {
		s.primed = true

		return bpValue
	}

	// Store band-pass output in the velocity ring buffer.
	s.velBuffer[s.velIndex] = bpValue
	s.velIndex = (s.velIndex + 1) % velocityTaps
	s.velCount++

	if s.velCount < velocityTaps {
		s.primed = false

		return math.NaN()
	}

	// Cubic velocity: kernel[k] multiplies the k-th most recent band-pass value.
	velValue := 0.0
	idx = s.velIndex - 1

	for k := 0; k < velocityTaps; k++ {
		bufIdx := ((idx % velocityTaps) + velocityTaps) % velocityTaps
		velValue += velocityKernel[k] * s.velBuffer[bufIdx]
		idx--
	}

	s.primed = true

	return velValue
}
