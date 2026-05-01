package jurikzerolagvelocity

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// velAux1 computes the linear regression slope over a window of Depth+1 points.
// Reference: JVELaux1(SrcA, Depth).
type velAux1 struct {
	depth int
	win   []float64 // ring buffer of size depth+1
	idx   int
	bar   int

	// Precomputed constants.
	jrc04 float64 // depth+1
	jrc05 float64 // jrc04*(jrc04+1)/2
	jrc06 float64 // jrc05*(2*jrc04+1)/3
	jrc07 float64 // jrc05^3 - jrc06^2
}

func newVelAux1(depth int) *velAux1 {
	size := depth + 1
	jrc04 := float64(size)
	jrc05 := jrc04 * (jrc04 + 1) / 2
	jrc06 := jrc05 * (2*jrc04 + 1) / 3
	jrc07 := jrc05*jrc05*jrc05 - jrc06*jrc06

	return &velAux1{
		depth: depth,
		win:   make([]float64, size),
		jrc04: jrc04,
		jrc05: jrc05,
		jrc06: jrc06,
		jrc07: jrc07,
	}
}

func (a *velAux1) update(sample float64) float64 {
	size := a.depth + 1
	a.win[a.idx] = sample
	a.idx = (a.idx + 1) % size
	a.bar++

	// Reference starts output at Bar=Depth (0-indexed). Our bar is 1-based.
	// refBar = bar-1. Output when refBar >= depth, i.e., bar > depth.
	if a.bar <= a.depth {
		return 0
	}

	// Compute jrc08 and jrc09 over the window.
	// Reference: for jrc10 = 0..depth: SrcA[Bar-jrc10] * (jrc04 - jrc10)
	// In ring: most recent is at (idx-1+size)%size, going back.
	var jrc08, jrc09 float64

	for j := 0; j <= a.depth; j++ {
		pos := (a.idx - 1 - j + size*2) % size
		w := a.jrc04 - float64(j)
		jrc08 += a.win[pos] * w
		jrc09 += a.win[pos] * w * w
	}

	return (jrc09*a.jrc05 - jrc08*a.jrc06) / a.jrc07
}

// JurikZeroLagVelocity computes the Jurik VEL indicator.
type JurikZeroLagVelocity struct {
	mu sync.RWMutex
	core.LineIndicator
	primed     bool
	paramDepth int

	aux1 *velAux1
	aux3 *velAux3State
	bar  int
}

// aux3State is a cleaner implementation that buffers history properly.
type velAux3State struct {
	// Constants.
	length int     // 30
	eps    float64 // 0.0001
	decay  int     // 3
	beta   float64 // 0.86 - 0.55/sqrt(3)
	alpha  float64 // 1 - exp(-ln(4)/3/2)
	maxWin int     // length + 1 = 31

	// Ring buffers.
	srcRing [100]float64 // JR41
	devRing [100]float64 // JR40
	srcIdx  int          // JR25
	devIdx  int          // JR18

	// State.
	jr08  float64 // velocity estimate
	jr09  float64 // sum of values in window
	jr10  float64 // weighted sum
	jr11  int     // current window size (grows to maxWin)
	jr12  float64 // variance denominator
	jr13  float64 // (window+1)/2
	jr14  float64 // (window-1)/2
	jr19  float64 // running deviation sum
	jr20  float64 // smoothed deviation
	jr21  float64 // current output (adaptive MA)
	jr21a float64
	jr21b float64
	jr22  float64
	jr23  float64

	bar       int
	initDone  bool
	history   []float64 // buffer for first 30 samples
	histCount int
}

func newVelAux3State() *velAux3State {
	length := 30
	decay := 3

	return &velAux3State{
		length:  length,
		eps:     0.0001,
		decay:   decay,
		beta:    0.86 - 0.55/math.Sqrt(float64(decay)),
		alpha:   1 - math.Exp(-math.Log(4)/float64(decay)/2),
		maxWin:  length + 1,
		history: make([]float64, 0, length),
	}
}

// feed processes one sample from the aux1 output stream.
// barIdx is the 0-based index in the full aux1 output array.
func (s *velAux3State) feed(sample float64, barIdx int) float64 {
	// aux3 in reference starts at Bar=30 (0-indexed). Before that, output is 0.
	if barIdx < s.length {
		s.history = append(s.history, sample)
		return 0
	}

	s.bar++

	// One-time initialization at first frame (Bar==JR01 in reference).
	if !s.initDone {
		s.initDone = true

		// Count consecutive equal values to determine JR26.
		jr28 := 0.0
		for j := 1; j <= s.length-1; j++ {
			if s.history[len(s.history)-j] == s.history[len(s.history)-j-1] {
				jr28++
			}
		}

		jr26 := 0
		if jr28 < float64(s.length-1) {
			jr26 = barIdx - s.length
		} else {
			jr26 = barIdx
		}

		s.jr11 = int(math.Trunc(math.Min(1+float64(barIdx-jr26), float64(s.maxWin))))

		// JR21 = SrcA[Bar-1]
		s.jr21 = s.history[len(s.history)-1]
		// JR03=3, compute constants.
		// JR08 = (SrcA[Bar] - SrcA[Bar-3]) / 3
		jr07 := 3
		s.jr08 = (sample - s.history[len(s.history)-jr07]) / float64(jr07)

		// Fill source ring with historical values: SrcA[Bar-JR15] for JR15=JR11-1..1.
		for jr15 := s.jr11 - 1; jr15 >= 1; jr15-- {
			if s.srcIdx <= 0 {
				s.srcIdx = 100
			}
			s.srcIdx--
			s.srcRing[s.srcIdx] = s.history[len(s.history)-jr15]
		}

		s.history = nil // free memory
		// Fall through to common code below.
	}

	// Common code (executes every frame including first).
	// Push current value to source ring.
	if s.srcIdx <= 0 {
		s.srcIdx = 100
	}
	s.srcIdx--
	s.srcRing[s.srcIdx] = sample

	if s.jr11 <= s.length {
		// Growing phase (JR11 <= JR01=30).
		if s.bar == 1 {
			// First frame: JR21 = SrcA[Bar] (overrides the SrcA[Bar-1] set in init).
			s.jr21 = sample
		} else {
			s.jr21 = math.Sqrt(s.alpha)*sample + (1-math.Sqrt(s.alpha))*s.jr21a
		}

		if s.bar > 2 {
			s.jr08 = (s.jr21 - s.jr21b) / 2
		} else {
			s.jr08 = 0
		}

		s.jr11++
	} else if s.jr11 <= s.maxWin {
		// Transition phase (JR01 < JR11 <= JR06=31): recompute from scratch.
		s.jr12 = float64(s.jr11*(s.jr11+1)*(s.jr11-1)) / 12
		s.jr13 = float64(s.jr11+1) / 2
		s.jr14 = float64(s.jr11-1) / 2

		s.jr09 = 0.0
		s.jr10 = 0.0

		for jr15 := s.jr11 - 1; jr15 >= 0; jr15-- {
			jr24 := (s.srcIdx + jr15) % 100
			s.jr09 += s.srcRing[jr24]
			s.jr10 += s.srcRing[jr24] * (s.jr14 - float64(jr15))
		}

		jr16 := s.jr10 / s.jr12
		jr17 := (s.jr09 / float64(s.jr11)) - (jr16 * s.jr13)

		s.jr19 = 0.0
		for jr15 := s.jr11 - 1; jr15 >= 0; jr15-- {
			jr17 += jr16
			jr24 := (s.srcIdx + jr15) % 100
			s.jr19 += math.Abs(s.srcRing[jr24] - jr17)
		}

		s.jr20 = (s.jr19 / float64(s.jr11)) * math.Pow(float64(s.maxWin)/float64(s.jr11), 0.25)
		s.jr11++

		// Adaptive step.
		s.jr20 = math.Max(s.eps, s.jr20)
		s.jr22 = sample - (s.jr21 + s.jr08*s.beta)
		s.jr23 = 1 - math.Exp(-math.Abs(s.jr22)/s.jr20/float64(s.decay))
		s.jr08 = s.jr23*s.jr22 + s.jr08*s.beta
		s.jr21 += s.jr08
	} else {
		// Steady state (JR11 > JR06).
		jr24out := (s.srcIdx + s.maxWin) % 100
		s.jr10 = s.jr10 - s.jr09 + s.srcRing[jr24out]*s.jr13 + sample*s.jr14
		s.jr09 = s.jr09 - s.srcRing[jr24out] + sample

		// Deviation ring update.
		if s.devIdx <= 0 {
			s.devIdx = s.maxWin
		}
		s.devIdx--
		s.jr19 -= s.devRing[s.devIdx]

		jr16 := s.jr10 / s.jr12
		jr17 := (s.jr09 / float64(s.maxWin)) + (jr16 * s.jr14)
		s.devRing[s.devIdx] = math.Abs(sample - jr17)
		s.jr19 = math.Max(s.eps, s.jr19+s.devRing[s.devIdx])
		s.jr20 += ((s.jr19 / float64(s.maxWin)) - s.jr20) * s.alpha

		// Adaptive step.
		s.jr20 = math.Max(s.eps, s.jr20)
		s.jr22 = sample - (s.jr21 + s.jr08*s.beta)
		s.jr23 = 1 - math.Exp(-math.Abs(s.jr22)/s.jr20/float64(s.decay))
		s.jr08 = s.jr23*s.jr22 + s.jr08*s.beta
		s.jr21 += s.jr08
	}

	s.jr21b = s.jr21a
	s.jr21a = s.jr21

	return s.jr21
}

// NewJurikZeroLagVelocity returns an instance of the indicator created using supplied parameters.
func NewJurikZeroLagVelocity(p *JurikZeroLagVelocityParams) (*JurikZeroLagVelocity, error) {
	return newJurikZeroLagVelocity(p.Depth, p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

func newJurikZeroLagVelocity(depth int,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*JurikZeroLagVelocity, error) {
	const (
		invalid = "invalid jurik zero lag velocity parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "vel(%d%s)"
	)

	var (
		mnemonic  string
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if depth < 2 {
		return nil, fmt.Errorf(fmts, invalid, "depth should be at least 2")
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

	mnemonic = fmt.Sprintf(fmtn, depth, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Jurik zero lag velocity " + mnemonic

	vel := &JurikZeroLagVelocity{
		paramDepth: depth,
		aux1:       newVelAux1(depth),
		aux3:       newVelAux3State(),
	}

	vel.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, vel.Update)

	return vel, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *JurikZeroLagVelocity) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *JurikZeroLagVelocity) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.JurikZeroLagVelocity,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the VEL indicator given the next sample.
func (s *JurikZeroLagVelocity) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Stage 1: compute linear regression slope.
	aux1Val := s.aux1.update(sample)

	// Stage 2: feed into adaptive smoother.
	// barIdx is 0-based index in the full stream.
	barIdx := s.bar
	s.bar++

	result := s.aux3.feed(aux1Val, barIdx)

	// Output is 0 during warmup → NaN.
	if barIdx < s.aux3.length {
		return math.NaN()
	}

	if !s.primed {
		s.primed = true
	}

	return result
}
