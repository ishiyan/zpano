package jurikrelativetrendstrengthindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// JurikRelativeTrendStrengthIndex computes the Jurik RSX indicator, see http://jurikres.com/.
// RSX is a noise-free version of RSI based on triple-smoothed EMA of momentum and absolute momentum.
type JurikRelativeTrendStrengthIndex struct {
	mu sync.RWMutex
	core.LineIndicator
	primed   bool
	paramLen int

	// State variables lifted from the reference batch loop.
	f0  int
	f88 int
	f90 int

	f8  float64
	f10 float64
	f18 float64
	f20 float64
	f28 float64
	f30 float64
	f38 float64
	f40 float64
	f48 float64
	f50 float64
	f58 float64
	f60 float64
	f68 float64
	f70 float64
	f78 float64
	f80 float64
}

// NewJurikRelativeTrendStrengthIndex returns an instance of the indicator created using supplied parameters.
func NewJurikRelativeTrendStrengthIndex(p *JurikRelativeTrendStrengthIndexParams) (*JurikRelativeTrendStrengthIndex, error) {
	return newJurikRelativeTrendStrengthIndex(p.Length,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

func newJurikRelativeTrendStrengthIndex(length int,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*JurikRelativeTrendStrengthIndex, error) {
	const (
		invalid = "invalid jurik relative trend strength index parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "jrsx(%d%s)"
		minlen  = 2
	)

	var (
		mnemonic  string
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if length < minlen {
		return nil, fmt.Errorf(fmts, invalid, "length should be at least 2")
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

	mnemonic = fmt.Sprintf(fmtn, length, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Jurik relative trend strength index " + mnemonic

	rsx := &JurikRelativeTrendStrengthIndex{paramLen: length}

	rsx.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, rsx.Update)

	return rsx, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *JurikRelativeTrendStrengthIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *JurikRelativeTrendStrengthIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.JurikRelativeTrendStrengthIndex,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the RSX indicator given the next sample.
func (s *JurikRelativeTrendStrengthIndex) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	const (
		hundred = 100.0
		fifty   = 50.0
		oneFive = 1.5
		half    = 0.5
		minLen  = 5
		eps     = 1e-10
	)

	length := s.paramLen

	if s.f90 == 0 {
		// First call: initialize.
		s.f90 = 1
		s.f0 = 0

		if length-1 >= minLen {
			s.f88 = length - 1
		} else {
			s.f88 = minLen
		}

		s.f8 = hundred * sample
		s.f18 = 3.0 / float64(length+2)
		s.f20 = 1 - s.f18
	} else {
		if s.f88 <= s.f90 {
			s.f90 = s.f88 + 1
		} else {
			s.f90++
		}

		s.f10 = s.f8
		s.f8 = hundred * sample
		v8 := s.f8 - s.f10

		s.f28 = s.f20*s.f28 + s.f18*v8
		s.f30 = s.f18*s.f28 + s.f20*s.f30
		vC := s.f28*oneFive - s.f30*half

		s.f38 = s.f20*s.f38 + s.f18*vC
		s.f40 = s.f18*s.f38 + s.f20*s.f40
		v10 := s.f38*oneFive - s.f40*half

		s.f48 = s.f20*s.f48 + s.f18*v10
		s.f50 = s.f18*s.f48 + s.f20*s.f50
		v14 := s.f48*oneFive - s.f50*half

		s.f58 = s.f20*s.f58 + s.f18*math.Abs(v8)
		s.f60 = s.f18*s.f58 + s.f20*s.f60
		v18 := s.f58*oneFive - s.f60*half

		s.f68 = s.f20*s.f68 + s.f18*v18
		s.f70 = s.f18*s.f68 + s.f20*s.f70
		v1C := s.f68*oneFive - s.f70*half

		s.f78 = s.f20*s.f78 + s.f18*v1C
		s.f80 = s.f18*s.f78 + s.f20*s.f80
		v20 := s.f78*oneFive - s.f80*half

		if s.f88 >= s.f90 && s.f8 != s.f10 {
			s.f0 = 1
		}

		if s.f88 == s.f90 && s.f0 == 0 {
			s.f90 = 0
		}

		if s.f88 < s.f90 && v20 > eps {
			v4 := (v14/v20 + 1) * fifty
			if v4 > hundred {
				v4 = hundred
			}

			if v4 < 0 {
				v4 = 0
			}

			s.primed = true

			return v4
		}
	}

	// During warmup or when denominator is too small.
	if s.f88 < s.f90 {
		s.primed = true
	}

	if !s.primed {
		return math.NaN()
	}

	return fifty
}
