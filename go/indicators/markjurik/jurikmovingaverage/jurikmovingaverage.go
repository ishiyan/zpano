package jurikmovingaverage

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// JurikMovingAverage computes the Jurik moving average (JMA), see http://jurikres.com/.
type JurikMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	primed bool
	list   []float64
	ring   []float64
	ring2  []float64
	buffer []float64
	s28    int
	s30    int
	s38    int
	s40    int
	s48    int
	s50    int
	s70    int
	f0     int
	fD8    int
	fF0    int
	v5     int
	s8     float64
	s18    float64
	f10    float64
	f18    float64
	f38    float64
	f50    float64
	f58    float64
	f78    float64
	f88    float64
	f90    float64
	f98    float64
	fA8    float64
	fB8    float64
	fC0    float64
	fC8    float64
	fF8    float64
	v1     float64
	v2     float64
	v3     float64
}

// NewJurikMovingAverage returns an instnce of the indicator created using supplied parameters.
func NewJurikMovingAverage(p *JurikMovingAverageParams) (*JurikMovingAverage, error) {
	return newJurikMovingAverage(p.Length, p.Phase,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

//nolint:funlen
func newJurikMovingAverage(length, phase int,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*JurikMovingAverage, error) {
	const (
		invalid      = "invalid jurik moving average parameters"
		fmts         = "%s: %s"
		fmtw         = "%s: %w"
		fmtn         = "jma(%d, %d%s)"
		minlen       = 1
		two          = 2
		hundred      = 100
		onePointFive = 1.5
		pointFive    = 0.5
		pointNine    = 0.9
		epsilon      = 1e-10 // 0.00000001
	)

	var (
		mnemonic  string
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if length < minlen {
		return nil, fmt.Errorf(fmts, invalid, "length should be positive")
	}

	if phase < -hundred || phase > hundred {
		return nil, fmt.Errorf(fmts, invalid, "phase should be in range [-100, 100]")
	}

	// Resolve defaults for component functions.
	// A zero value means "use default".
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

	// Build mnemonic using resolved components — defaults are omitted by ComponentTripleMnemonic.
	mnemonic = fmt.Sprintf(fmtn, length, phase, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Jurik moving average " + mnemonic

	const (
		c128  = 128
		c11   = 11
		c62   = 62
		c28   = 63
		c30   = 64
		cInit = 1000000.0
	)

	// These slices will be automatically filled with zeroes.
	list := make([]float64, c128)
	ring := make([]float64, c128)
	ring2 := make([]float64, c11)
	buffer := make([]float64, c62)

	for i := range c30 {
		list[i] = -cInit
	}

	for i := c30; i < c128; i++ {
		list[i] = cInit
	}

	f80 := epsilon
	if length > 1 {
		f80 = (float64(length) - 1) / two
	}

	f10 := float64(phase)/hundred + onePointFive

	v1 := math.Log(math.Sqrt(f80))
	v2 := v1
	v3 := max(v2/math.Log(two)+two, 0)

	f98 := v3
	f88 := max(f98-two, pointFive)

	f78 := math.Sqrt(f80) * f98
	f90 := f78 / (f78 + 1)
	f80 *= pointNine
	f50 := f80 / (f80 + two)

	jma := &JurikMovingAverage{
		list:   list,
		ring:   ring,
		ring2:  ring2,
		buffer: buffer,
		s28:    c28,
		s30:    c30,
		f0:     1,
		f10:    f10,
		f50:    f50,
		f78:    f78,
		f88:    f88,
		f90:    f90,
		f98:    f98,
		v1:     v1,
		v2:     v2,
		v3:     v3,
	}

	jma.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, jma.Update)

	return jma, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *JurikMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the Jurik moving average.
func (s *JurikMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.JurikMovingAverage,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the Jurik moving average given the next sample.
//
// The indicator is not primed during the first 30 updates.
func (s *JurikMovingAverage) Update(sample float64) float64 { //nolint:funlen, cyclop, gocognit, gocyclo, maintidx
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	const (
		c2      = 2
		c10     = 10
		c29     = 29
		c30     = 30
		c31     = 31
		c32     = 32
		c61     = 61
		c64     = 64
		c96     = 96
		c127    = 127
		c128    = 128
		two     = 2
		epsilon = 1e-10
	)

	if s.fF0 < c61 {
		s.fF0++
		s.buffer[s.fF0] = sample
	}

	if s.fF0 <= c30 {
		return math.NaN()
	}

	s.primed = true
	if s.f0 == 0 { //nolint:nestif
		s.fD8 = 0
	} else {
		s.f0 = 0
		s.v5 = 0

		for i := 1; i < c30; i++ {
			if s.buffer[i+1] != s.buffer[i] {
				s.v5 = 1
			}
		}

		s.fD8 = s.v5 * c30
		if s.fD8 == 0 {
			s.f38 = sample
		} else {
			s.f38 = s.buffer[1]
		}

		s.f18 = s.f38
		if s.fD8 > c29 {
			s.fD8 = c29
		}
	}

	for i := s.fD8; i >= 0; i-- {
		f8 := sample
		if i != 0 {
			f8 = s.buffer[c31-i]
		}

		f28 := f8 - s.f18
		f48 := f8 - s.f38
		a28 := math.Abs(f28)
		a48 := math.Abs(f48)
		s.v2 = max(a28, a48)

		fA0 := s.v2
		v := fA0 + epsilon

		if s.s48 <= 1 {
			s.s48 = c127
		} else {
			s.s48--
		}

		if s.s50 <= 1 {
			s.s50 = c10
		} else {
			s.s50--
		}

		if s.s70 < c128 {
			s.s70++
		}

		s.s8 += v - s.ring2[s.s50]
		s.ring2[s.s50] = v
		s20 := s.s8 / float64(s.s70)

		if s.s70 > c10 {
			s20 = s.s8 / c10
		}

		var s58, s68 int

		if s.s70 > c127 { //nolint:nestif
			s10 := s.ring[s.s48]
			s.ring[s.s48] = s20
			s68 = c64
			s58 = s68

			for s68 > 1 {
				if s.list[s58] < s10 {
					s68 /= c2
					s58 += s68
				} else if s.list[s58] <= s10 {
					s68 = 1
				} else {
					s68 /= c2
					s58 -= s68
				}
			}
		} else {
			s.ring[s.s48] = s20
			if s.s28+s.s30 > c127 {
				s.s30--
				s58 = s.s30
			} else {
				s.s28++
				s58 = s.s28
			}

			s.s38 = min(s.s28, c96)
			s.s40 = max(s.s30, c32)
		}

		s68 = c64
		s60 := s68

		for s68 > 1 {
			if s.list[s60] >= s20 {
				if s.list[s60-1] <= s20 {
					s68 = 1
				} else {
					s68 /= c2
					s60 -= s68
				}
			} else {
				s68 /= c2
				s60 += s68
			}

			if s60 == c127 && s20 > s.list[c127] {
				s60 = c128
			}
		}

		if s.s70 > c127 { //nolint:nestif
			if s58 >= s60 {
				if s.s38+1 > s60 && s.s40-1 < s60 {
					s.s18 += s20
				} else if s.s40 > s60 && s.s40-1 < s58 {
					s.s18 += s.list[s.s40-1]
				}
			} else if s.s40 >= s60 {
				if s.s38+1 < s60 && s.s38+1 > s58 {
					s.s18 += s.list[s.s38+1]
				}
			} else if s.s38+2 > s60 {
				s.s18 += s20
			} else if s.s38+1 < s60 && s.s38+1 > s58 {
				s.s18 += s.list[s.s38+1]
			}

			if s58 > s60 {
				if s.s40-1 < s58 && s.s38+1 > s58 {
					s.s18 -= s.list[s58]
				} else if s.s38 < s58 && s.s38+1 > s60 {
					s.s18 -= s.list[s.s38]
				}
			} else {
				if s.s38+1 > s58 && s.s40-1 < s58 {
					s.s18 -= s.list[s58]
				} else if s.s40 > s58 && s.s40 < s60 {
					s.s18 -= s.list[s.s40]
				}
			}
		}

		if s58 <= s60 {
			if s58 >= s60 {
				s.list[s60] = s20
			} else {
				for k := s58 + 1; k <= s60-1; k++ {
					s.list[k-1] = s.list[k]
				}

				s.list[s60-1] = s20
			}
		} else {
			for k := s58 - 1; k >= s60; k-- {
				s.list[k+1] = s.list[k]
			}

			s.list[s60] = s20
		}

		if s.s70 < c128 {
			s.s18 = 0
			for k := s.s40; k <= s.s38; k++ {
				s.s18 += s.list[k]
			}
		}

		f60 := s.s18 / float64(s.s38-s.s40+1)

		if s.fF8+1 > c31 {
			s.fF8 = c31
		} else {
			s.fF8++
		}

		if s.fF8 <= c30 { //nolint:nestif
			if f28 > 0 {
				s.f18 = f8
			} else {
				s.f18 = f8 - f28*s.f90
			}

			if f48 < 0 {
				s.f38 = f8
			} else {
				s.f38 = f8 - f48*s.f90
			}

			s.fB8 = sample
			if s.fF8 != c30 {
				continue
			}

			v4 := 1
			s.fC0 = sample

			if math.Ceil(s.f78) >= 1 {
				v4 = int(math.Ceil(s.f78))
			}

			v2 := 1
			fE8 := v4

			if math.Floor(s.f78) >= 1 {
				v2 = int(math.Floor(s.f78))
			}

			f68 := 1.0
			fE0 := v2

			if fE8 != fE0 {
				v4 = fE8 - fE0
				f68 = (s.f78 - float64(fE0)) / float64(v4)
			}

			v5 := min(fE0, c29)
			v6 := min(fE8, c29)
			s.fA8 = (sample-s.buffer[s.fF0-v5])*(1-f68)/float64(fE0) +
				(sample-s.buffer[s.fF0-v6])*f68/float64(fE8)
		} else {
			p := math.Pow(fA0/f60, s.f88)
			s.v1 = min(s.f98, p)

			if s.v1 < 1 {
				s.v2 = 1
			} else {
				s.v3 = min(s.f98, p)
				s.v2 = s.v3
			}

			s.f58 = s.v2
			f70 := math.Pow(s.f90, math.Sqrt(s.f58))

			if f28 > 0 {
				s.f18 = f8
			} else {
				s.f18 = f8 - f28*f70
			}

			if f48 < 0 {
				s.f38 = f8
			} else {
				s.f38 = f8 - f48*f70
			}
		}
	}

	if s.fF8 > c30 {
		f30 := math.Pow(s.f50, s.f58)
		s.fC0 = (1-f30)*sample + f30*s.fC0
		s.fC8 = (sample-s.fC0)*(1-s.f50) + s.f50*s.fC8
		fD0 := s.f10*s.fC8 + s.fC0
		f20 := f30 * -two
		f40 := f30 * f30
		fB0 := f20 + f40 + 1
		s.fA8 = (fD0-s.fB8)*fB0 + f40*s.fA8
		s.fB8 += s.fA8
	}

	return s.fB8
}
