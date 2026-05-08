package jurikadaptiverelativetrendstrengthindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// JurikAdaptiveRelativeTrendStrengthIndex computes the Jurik Adaptive RSX indicator (JARSX).
// Combines adaptive length selection (volatility regime detection) with
// the RSX core (triple-cascaded lag-reduced EMA oscillator). Output in [0, 100].
type JurikAdaptiveRelativeTrendStrengthIndex struct {
	mu sync.RWMutex
	core.LineIndicator
	primed   bool
	loLength int
	hiLength int
	eps      float64

	barCount      int
	previousPrice float64

	// Rolling buffers for adaptive length.
	longBuffer [100]float64
	longIndex  int
	longSum    float64
	longCount  int
	shortBuffer [10]float64
	shortIndex  int
	shortSum    float64
	shortCount  int

	// RSX core state.
	kg     float64
	c      float64
	warmup int
	// Signal path (3 cascaded stages).
	sig1A, sig1B float64
	sig2A, sig2B float64
	sig3A, sig3B float64
	// Denominator path (3 cascaded stages).
	den1A, den1B float64
	den2A, den2B float64
	den3A, den3B float64
}

// NewJurikAdaptiveRelativeTrendStrengthIndex returns an instance of the indicator.
func NewJurikAdaptiveRelativeTrendStrengthIndex(p *JurikAdaptiveRelativeTrendStrengthIndexParams) (*JurikAdaptiveRelativeTrendStrengthIndex, error) {
	return newJurikAdaptiveRelativeTrendStrengthIndex(p.LoLength, p.HiLength, p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

func newJurikAdaptiveRelativeTrendStrengthIndex(loLength, hiLength int,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*JurikAdaptiveRelativeTrendStrengthIndex, error) {
	const (
		invalid = "invalid jurik adaptive relative trend strength index parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "jarsx(%d, %d%s)"
	)

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if loLength < 2 {
		return nil, fmt.Errorf(fmts, invalid, "lo_length should be at least 2")
	}

	if hiLength < loLength {
		return nil, fmt.Errorf(fmts, invalid, "hi_length should be at least lo_length")
	}

	if bc == 0 {
		bc = entities.DefaultBarComponent
	}

	if qc == 0 {
		qc = entities.DefaultQuoteComponent
	}

	if tc == 0 {
		tc = entities.DefaultTradeComponent
	}

	if barFunc, err = entities.BarComponentFunc(bc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if quoteFunc, err = entities.QuoteComponentFunc(qc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	if tradeFunc, err = entities.TradeComponentFunc(tc); err != nil {
		return nil, fmt.Errorf(fmtw, invalid, err)
	}

	mnemonic := fmt.Sprintf(fmtn, loLength, hiLength, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Jurik adaptive relative trend strength index " + mnemonic

	ind := &JurikAdaptiveRelativeTrendStrengthIndex{
		loLength: loLength,
		hiLength: hiLength,
		eps:      0.001,
	}

	ind.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, ind.Update)

	return ind, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *JurikAdaptiveRelativeTrendStrengthIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *JurikAdaptiveRelativeTrendStrengthIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.JurikAdaptiveRelativeTrendStrengthIndex,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample.
func (s *JurikAdaptiveRelativeTrendStrengthIndex) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	bar := s.barCount
	s.barCount++

	if bar == 0 {
		s.previousPrice = sample

		// First bar: add 0 to both buffers.
		s.longBuffer[0] = 0.0
		s.longSum = 0.0
		s.longCount = 1
		s.shortBuffer[0] = 0.0
		s.shortSum = 0.0
		s.shortCount = 1

		// Compute adaptive length from bar 0.
		avg1 := 0.0
		avg2 := 0.0
		value2 := math.Log((s.eps + avg1) / (s.eps + avg2))
		value3 := value2 / (1.0 + math.Abs(value2))
		adaptiveLength := float64(s.loLength) +
			float64(s.hiLength-s.loLength)*(1.0+value3)/2.0
		length := int(adaptiveLength)
		if length < 2 {
			length = 2
		}

		s.kg = 3.0 / float64(length+2)
		s.c = 1.0 - s.kg
		s.warmup = length - 1
		if s.warmup < 5 {
			s.warmup = 5
		}

		return math.NaN()
	}

	// Bars 1+
	oldPrice := s.previousPrice
	s.previousPrice = sample
	value1 := math.Abs(sample - oldPrice)

	// Update long rolling buffer.
	if s.longCount < 100 {
		s.longBuffer[s.longCount] = value1
		s.longSum += value1
		s.longCount++
	} else {
		s.longSum -= s.longBuffer[s.longIndex]
		s.longBuffer[s.longIndex] = value1
		s.longSum += value1
		s.longIndex = (s.longIndex + 1) % 100
	}

	// Update short rolling buffer.
	if s.shortCount < 10 {
		s.shortBuffer[s.shortCount] = value1
		s.shortSum += value1
		s.shortCount++
	} else {
		s.shortSum -= s.shortBuffer[s.shortIndex]
		s.shortBuffer[s.shortIndex] = value1
		s.shortSum += value1
		s.shortIndex = (s.shortIndex + 1) % 10
	}

	// RSX core computation.
	mom := 100.0 * (sample - oldPrice)
	absMom := math.Abs(mom)

	kg := s.kg
	c := s.c

	// Signal path — Stage 1.
	s.sig1A = c*s.sig1A + kg*mom
	s.sig1B = kg*s.sig1A + c*s.sig1B
	s1 := 1.5*s.sig1A - 0.5*s.sig1B

	// Signal path — Stage 2.
	s.sig2A = c*s.sig2A + kg*s1
	s.sig2B = kg*s.sig2A + c*s.sig2B
	s2 := 1.5*s.sig2A - 0.5*s.sig2B

	// Signal path — Stage 3.
	s.sig3A = c*s.sig3A + kg*s2
	s.sig3B = kg*s.sig3A + c*s.sig3B
	numerator := 1.5*s.sig3A - 0.5*s.sig3B

	// Denominator path — Stage 1.
	s.den1A = c*s.den1A + kg*absMom
	s.den1B = kg*s.den1A + c*s.den1B
	d1 := 1.5*s.den1A - 0.5*s.den1B

	// Denominator path — Stage 2.
	s.den2A = c*s.den2A + kg*d1
	s.den2B = kg*s.den2A + c*s.den2B
	d2 := 1.5*s.den2A - 0.5*s.den2B

	// Denominator path — Stage 3.
	s.den3A = c*s.den3A + kg*d2
	s.den3B = kg*s.den3A + c*s.den3B
	denominator := 1.5*s.den3A - 0.5*s.den3B

	// Output after warmup.
	if bar >= s.warmup {
		s.primed = true

		var value float64
		if denominator != 0.0 {
			value = (numerator/denominator + 1.0) * 50.0
		} else {
			value = 50.0
		}

		return math.Max(0.0, math.Min(100.0, value))
	}

	return math.NaN()
}
