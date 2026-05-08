package jurikfractaladaptivezerolagvelocity

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// Scale sets for different fractal types.
var scaleSets = map[int][]int{
	1: {2, 3, 4, 6, 8, 12, 16, 24},
	2: {2, 3, 4, 6, 8, 12, 16, 24, 32, 48},
	3: {2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96},
	4: {2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192},
}

var weightsEven = []float64{2, 3, 6, 12, 24, 48, 96}
var weightsOdd = []float64{4, 8, 16, 32, 64, 128, 256}

// cfbAux is the streaming state for a single CFB channel at one depth.
type cfbAux struct {
	depth     int
	bar       int
	intA      []float64
	intAIdx   int
	src       []float64
	srcIdx    int
	jrc04     float64
	jrc05     float64
	jrc06     float64
	prevSample float64
	firstCall  bool
}

func newCfbAux(depth int) *cfbAux {
	return &cfbAux{
		depth:     depth,
		intA:      make([]float64, depth),
		src:       make([]float64, depth+2),
		firstCall: true,
	}
}

func (a *cfbAux) update(sample float64) float64 {
	a.bar++
	depth := a.depth
	srcSize := depth + 2

	a.src[a.srcIdx] = sample
	a.srcIdx = (a.srcIdx + 1) % srcSize

	if a.firstCall {
		a.firstCall = false
		a.prevSample = sample

		return 0.0
	}

	intAVal := math.Abs(sample - a.prevSample)
	a.prevSample = sample

	oldIntA := a.intA[a.intAIdx]
	a.intA[a.intAIdx] = intAVal
	a.intAIdx = (a.intAIdx + 1) % depth

	refBar := a.bar - 1
	if refBar < depth {
		return 0.0
	}

	if refBar <= depth*2 {
		// Recompute from scratch.
		a.jrc04 = 0.0
		a.jrc05 = 0.0
		a.jrc06 = 0.0

		curIntAPos := (a.intAIdx - 1 + depth) % depth
		curSrcPos := (a.srcIdx - 1 + srcSize) % srcSize

		for j := 0; j < depth; j++ {
			intAPos := (curIntAPos - j + depth) % depth
			srcPos := (curSrcPos - j - 1 + srcSize*2) % srcSize

			a.jrc04 += a.intA[intAPos]
			a.jrc05 += float64(depth-j) * a.intA[intAPos]
			a.jrc06 += a.src[srcPos]
		}
	} else {
		// Incremental update.
		a.jrc05 = a.jrc05 - a.jrc04 + intAVal*float64(depth)
		a.jrc04 = a.jrc04 - oldIntA + intAVal

		curSrcPos := (a.srcIdx - 1 + srcSize) % srcSize
		srcBarMinus1 := (curSrcPos - 1 + srcSize) % srcSize
		srcBarMinusDepthMinus1 := (curSrcPos - depth - 1 + srcSize*2) % srcSize

		a.jrc06 = a.jrc06 - a.src[srcBarMinusDepthMinus1] + a.src[srcBarMinus1]
	}

	curSrcPos := (a.srcIdx - 1 + srcSize) % srcSize
	jrc08 := math.Abs(float64(depth)*a.src[curSrcPos] - a.jrc06)

	if a.jrc05 == 0.0 {
		return 0.0
	}

	return jrc08 / a.jrc05
}

// cfb computes the Composite Fractal Behavior weighted dominant cycle.
type cfb struct {
	scales      []int
	numChannels int
	auxs        []*cfbAux
	auxWindows  [][]float64
	auxWinIdx   int
	er23        []float64
	smooth      int
	bar         int
	cfbValue    float64
}

func newCfb(fractalType, smooth int) *cfb {
	scales := scaleSets[fractalType]
	n := len(scales)
	auxs := make([]*cfbAux, n)

	for i, d := range scales {
		auxs[i] = newCfbAux(d)
	}

	auxWindows := make([][]float64, n)
	for i := range auxWindows {
		auxWindows[i] = make([]float64, smooth)
	}

	return &cfb{
		scales:      scales,
		numChannels: n,
		auxs:        auxs,
		auxWindows:  auxWindows,
		er23:        make([]float64, n),
		smooth:      smooth,
	}
}

func (c *cfb) update(sample float64) float64 {
	c.bar++
	refBar := c.bar - 1

	auxValues := make([]float64, c.numChannels)
	for i, aux := range c.auxs {
		auxValues[i] = aux.update(sample)
	}

	if refBar == 0 {
		return 0.0
	}

	smooth := c.smooth
	n := c.numChannels

	if refBar <= smooth {
		winPos := c.auxWinIdx
		for i := 0; i < n; i++ {
			c.auxWindows[i][winPos] = auxValues[i]
		}

		c.auxWinIdx = (c.auxWinIdx + 1) % smooth

		for i := 0; i < n; i++ {
			s := 0.0
			for j := 0; j < refBar; j++ {
				pos := (c.auxWinIdx - 1 - j + smooth*2) % smooth
				s += c.auxWindows[i][pos]
			}

			c.er23[i] = s / float64(refBar)
		}
	} else {
		winPos := c.auxWinIdx
		for i := 0; i < n; i++ {
			oldVal := c.auxWindows[i][winPos]
			c.auxWindows[i][winPos] = auxValues[i]
			c.er23[i] += (auxValues[i] - oldVal) / float64(smooth)
		}

		c.auxWinIdx = (c.auxWinIdx + 1) % smooth
	}

	if refBar > 5 {
		er22 := make([]float64, n)

		// Odd-indexed channels (descending).
		er15 := 1.0
		for idx := n - 1; idx >= 1; idx -= 2 {
			er22[idx] = er15 * c.er23[idx]
			er15 *= (1 - er22[idx])
		}

		// Even-indexed channels (descending).
		er16 := 1.0
		for idx := n - 2; idx >= 0; idx -= 2 {
			er22[idx] = er16 * c.er23[idx]
			er16 *= (1 - er22[idx])
		}

		// Weighted sum.
		var er17, er18 float64
		for idx := 0; idx < n; idx++ {
			sq := er22[idx] * er22[idx]
			er18 += sq

			if idx%2 == 0 {
				er17 += sq * weightsEven[idx/2]
			} else {
				er17 += sq * weightsOdd[idx/2]
			}
		}

		if er18 == 0.0 {
			c.cfbValue = 0.0
		} else {
			c.cfbValue = er17 / er18
		}
	}

	return c.cfbValue
}

// velSmooth is the adaptive smoother (Stage 2) with fixed period=3.0.
type velSmooth struct {
	jrc03       float64
	jrc06       int
	jrc07       int
	emaFactor   float64
	damping     float64
	eps2        float64
	bufferSize  int
	buffer      []float64
	idx         int
	length      int
	velocity    float64
	position    float64
	smoothedMAD float64
	madInit     bool
	initialized bool
}

func newVelSmooth(period float64) *velSmooth {
	const eps2 = 0.0001

	jrc03 := math.Min(500.0, math.Max(eps2, period))
	jrc06 := int(math.Max(31, math.Ceil(2*period)))
	jrc07 := int(math.Min(30, math.Ceil(period)))
	emaFactor := 1.0 - math.Exp(-math.Log(4.0)/(period/2.0))
	damping := 0.86 - 0.55/math.Sqrt(jrc03)

	return &velSmooth{
		jrc03:      jrc03,
		jrc06:      jrc06,
		jrc07:      jrc07,
		emaFactor:  emaFactor,
		damping:    damping,
		eps2:       eps2,
		bufferSize: 1001,
		buffer:     make([]float64, 1001),
	}
}

func (v *velSmooth) update(value float64) float64 {
	v.buffer[v.idx] = value
	v.idx = (v.idx + 1) % v.bufferSize
	v.length++

	if v.length > v.bufferSize {
		v.length = v.bufferSize
	}

	length := v.length

	if !v.initialized {
		v.initialized = true
		v.position = value
		v.velocity = 0.0
		v.smoothedMAD = 0.0

		return v.position
	}

	// Linear regression over capped window.
	n := length
	if n > v.jrc06 {
		n = v.jrc06
	}

	var sx, sy, sxy, sx2 float64

	for i := 0; i < n; i++ {
		idx := (v.idx - 1 - i + v.bufferSize) % v.bufferSize
		x := float64(i)
		y := v.buffer[idx]
		sx += x
		sy += y
		sxy += x * y
		sx2 += x * x
	}

	fn := float64(n)
	var slope float64

	if n > 1 {
		slope = (fn*sxy - sx*sy) / (fn*sx2 - sx*sx)
	}

	intercept := (sy - slope*sx) / fn

	// MAD from regression residuals.
	var mad float64
	for i := 0; i < n; i++ {
		idx := (v.idx - 1 - i + v.bufferSize) % v.bufferSize
		predicted := intercept + slope*float64(i)
		mad += math.Abs(v.buffer[idx] - predicted)
	}

	mad /= fn

	// Scale MAD.
	scaledMAD := mad * 1.2 * math.Pow(float64(v.jrc06)/fn, 0.25)

	// Smooth MAD with EMA.
	if !v.madInit {
		v.smoothedMAD = scaledMAD
		if scaledMAD > 0 {
			v.madInit = true
		}
	} else {
		v.smoothedMAD += (scaledMAD - v.smoothedMAD) * v.emaFactor
	}

	smoothedMAD := math.Max(v.eps2, v.smoothedMAD)

	// Adaptive velocity/position dynamics.
	predictionError := value - v.position
	responseFactor := 1.0 - math.Exp(-math.Abs(predictionError)/(smoothedMAD*v.jrc03))
	v.velocity = responseFactor*predictionError + v.velocity*v.damping
	v.position += v.velocity

	return v.position
}

// JurikFractalAdaptiveZeroLagVelocity computes the JVELCFB indicator.
type JurikFractalAdaptiveZeroLagVelocity struct {
	mu sync.RWMutex
	core.LineIndicator
	primed  bool
	loDepth int
	hiDepth int

	prices   []float64
	barCount int
	cfbInst  *cfb
	cfbMin   *float64
	cfbMax   *float64
	smooth   *velSmooth
}

// NewJurikFractalAdaptiveZeroLagVelocity returns an instance of the indicator.
func NewJurikFractalAdaptiveZeroLagVelocity(p *JurikFractalAdaptiveZeroLagVelocityParams) (*JurikFractalAdaptiveZeroLagVelocity, error) {
	return newJurikFractalAdaptiveZeroLagVelocity(p.LoDepth, p.HiDepth, p.FractalType, p.Smooth,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

func newJurikFractalAdaptiveZeroLagVelocity(loDepth, hiDepth, fractalType, smooth int,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*JurikFractalAdaptiveZeroLagVelocity, error) {
	const (
		invalid = "invalid jurik fractal adaptive zero lag velocity parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "jvelcfb(%d, %d, %d, %d%s)"
	)

	var (
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if loDepth < 2 {
		return nil, fmt.Errorf(fmts, invalid, "lo_depth should be at least 2")
	}

	if hiDepth < loDepth {
		return nil, fmt.Errorf(fmts, invalid, "hi_depth should be at least lo_depth")
	}

	if fractalType < 1 || fractalType > 4 {
		return nil, fmt.Errorf(fmts, invalid, "fractal_type should be 1-4")
	}

	if smooth < 1 {
		return nil, fmt.Errorf(fmts, invalid, "smooth should be at least 1")
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

	mnemonic := fmt.Sprintf(fmtn, loDepth, hiDepth, fractalType, smooth,
		core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Jurik fractal adaptive zero lag velocity " + mnemonic

	ind := &JurikFractalAdaptiveZeroLagVelocity{
		loDepth: loDepth,
		hiDepth: hiDepth,
		cfbInst: newCfb(fractalType, smooth),
		smooth:  newVelSmooth(3.0),
	}

	ind.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, ind.Update)

	return ind, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *JurikFractalAdaptiveZeroLagVelocity) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *JurikFractalAdaptiveZeroLagVelocity) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.JurikFractalAdaptiveZeroLagVelocity,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the indicator given the next sample.
func (s *JurikFractalAdaptiveZeroLagVelocity) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	bar := s.barCount
	s.barCount++

	s.prices = append(s.prices, sample)

	// CFB computation.
	cfbVal := s.cfbInst.update(sample)

	if bar == 0 {
		return math.NaN()
	}

	// Stochastic normalization.
	if s.cfbMin == nil {
		min := cfbVal
		max := cfbVal
		s.cfbMin = &min
		s.cfbMax = &max
	} else {
		if cfbVal < *s.cfbMin {
			*s.cfbMin = cfbVal
		}

		if cfbVal > *s.cfbMax {
			*s.cfbMax = cfbVal
		}
	}

	cfbRange := *s.cfbMax - *s.cfbMin
	var sr float64

	if cfbRange != 0.0 {
		sr = (cfbVal - *s.cfbMin) / cfbRange
	} else {
		sr = 0.5
	}

	depthF := float64(s.loDepth) + sr*float64(s.hiDepth-s.loDepth)
	depth := int(math.Round(depthF))

	// Stage 1: WLS slope.
	if bar < depth {
		return math.NaN()
	}

	n := float64(depth + 1)
	s1 := n * (n + 1) / 2.0
	s2 := s1 * (2*n + 1) / 3.0
	denom := s1*s1*s1 - s2*s2

	var sumXW, sumXW2 float64

	for i := 0; i <= depth; i++ {
		w := n - float64(i)
		p := s.prices[bar-i]
		sumXW += p * w
		sumXW2 += p * w * w
	}

	slope := (sumXW2*s1 - sumXW*s2) / denom

	// Stage 2: adaptive smoother.
	result := s.smooth.update(slope)

	if !s.primed {
		s.primed = true
	}

	return result
}
