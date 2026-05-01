package jurikcompositefractalbehaviorindex

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// depthSets defines the fractal depths used for each FractalType (1–4).
var depthSets = [4][]int{
	{2, 3, 4, 6, 8, 12, 16, 24},                           // Type 1: JCFB24
	{2, 3, 4, 6, 8, 12, 16, 24, 32, 48},                   // Type 2: JCFB48
	{2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96},           // Type 3: JCFB96
	{2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192}, // Type 4: JCFB192
}

// weights for the weighted composite: indices 0..13 map to depths [2,4,3,8,6,16,12,32,24,64,48,128,96,256].
// Even indices get weights: 2, 3, 6, 12, 24, 48, 96
// Odd indices get weights: 4, 8, 16, 32, 64, 128, 256
var weightsEven = []float64{2, 3, 6, 12, 24, 48, 96}
var weightsOdd = []float64{4, 8, 16, 32, 64, 128, 256}

// cfbAux is the streaming state for a single JCFBaux(depth) instance.
type cfbAux struct {
	depth int
	bar   int

	// Ring buffer for IntA values (abs differences). Size = depth.
	intA    []float64
	intAIdx int

	// Ring buffer for source values. Size = depth+1 (need SrcA[Bar-depth] through SrcA[Bar]).
	src    []float64
	srcIdx int

	// Running sums.
	jrc04 float64 // sum of |diff| over last depth bars
	jrc05 float64 // weighted sum
	jrc06 float64 // sum of source values over last depth bars (SrcA[Bar-1]..SrcA[Bar-depth])

	prevSample float64
	firstCall  bool
}

func newCFBAux(depth int) *cfbAux {
	return &cfbAux{
		depth:     depth,
		intA:      make([]float64, depth),
		src:       make([]float64, depth+2),
		firstCall: true,
	}
}

// update processes one new sample and returns the JCFBaux value for this bar.
func (a *cfbAux) update(sample float64) float64 {
	a.bar++

	// Store sample in source ring buffer (size = depth+2).
	srcSize := a.depth + 2
	a.src[a.srcIdx] = sample
	a.srcIdx = (a.srcIdx + 1) % srcSize

	// Compute IntA = |sample - prevSample| (from second call onward).
	var intAVal float64
	if a.firstCall {
		a.firstCall = false
		a.prevSample = sample
		return 0
	}

	intAVal = math.Abs(sample - a.prevSample)
	a.prevSample = sample

	// Store in intA ring buffer (size = depth).
	oldIntA := a.intA[a.intAIdx]
	a.intA[a.intAIdx] = intAVal
	a.intAIdx = (a.intAIdx + 1) % a.depth

	// Reference outputs from Bar=Depth to Bar=len-2 (0-indexed).
	// Our bar counter is 1-based: bar=1 = SrcA[0]. Ref Bar = bar-1.
	// Output when refBar >= Depth, i.e., bar >= Depth+1.
	refBar := a.bar - 1
	if refBar < a.depth {
		return 0
	}

	// Reference: if Bar <= Depth*2, recompute from scratch; else incremental.
	if refBar <= a.depth*2 {
		a.jrc04 = 0
		a.jrc05 = 0
		a.jrc06 = 0

		// Sum IntA[Bar], IntA[Bar-1], ..., IntA[Bar-Depth+1] and corresponding weighted/src values.
		// In ring: latest intA at (intAIdx-1+depth)%depth, going back.
		curIntAPos := (a.intAIdx - 1 + a.depth) % a.depth
		// SrcA[Bar-jrc07-1]: src values at positions 1..depth back from current.
		// Current sample is at (srcIdx-1+srcSize)%srcSize in the src ring.
		curSrcPos := (a.srcIdx - 1 + srcSize) % srcSize

		for j := 0; j < a.depth; j++ {
			intAPos := (curIntAPos - j + a.depth) % a.depth
			intAV := a.intA[intAPos]

			// SrcA[Bar-j-1] = source j+1 positions back from current.
			srcPos := (curSrcPos - j - 1 + srcSize*2) % srcSize
			srcV := a.src[srcPos]

			a.jrc04 += intAV
			a.jrc05 += float64(a.depth-j) * intAV
			a.jrc06 += srcV
		}
	} else {
		// Incremental update.
		a.jrc05 = a.jrc05 - a.jrc04 + intAVal*float64(a.depth)
		a.jrc04 = a.jrc04 - oldIntA + intAVal

		// SrcA[Bar-1] and SrcA[Bar-Depth-1].
		curSrcPos := (a.srcIdx - 1 + srcSize) % srcSize
		srcBarMinus1 := (curSrcPos - 1 + srcSize) % srcSize
		srcBarMinusDepthMinus1 := (curSrcPos - a.depth - 1 + srcSize) % srcSize

		a.jrc06 = a.jrc06 - a.src[srcBarMinusDepthMinus1] + a.src[srcBarMinus1]
	}

	// jrc08 = |Depth * SrcA[Bar] - jrc06|
	curSrcPos := (a.srcIdx - 1 + srcSize) % srcSize
	jrc08 := math.Abs(float64(a.depth)*a.src[curSrcPos] - a.jrc06)

	if a.jrc05 == 0 {
		return 0
	}

	return jrc08 / a.jrc05
}

// JurikCompositeFractalBehaviorIndex computes the Jurik CFB indicator, see http://jurikres.com/.
// CFB measures composite fractal behavior across multiple time depths.
type JurikCompositeFractalBehaviorIndex struct {
	mu sync.RWMutex
	core.LineIndicator
	primed       bool
	paramFractal int
	paramSmooth  int

	// Number of depth channels.
	numChannels int

	// Aux instances (one per depth channel).
	auxInstances []*cfbAux

	// Sliding window per channel for running average (er23).
	// Each is a ring buffer of size Smooth.
	auxWindows [][]float64
	auxWinIdx  int
	auxWinLen  int // current fill level (up to Smooth)

	// Running sums for er23 (the averages).
	er23 []float64

	bar  int
	er19 float64
}

// NewJurikCompositeFractalBehaviorIndex returns an instance of the indicator created using supplied parameters.
func NewJurikCompositeFractalBehaviorIndex(p *JurikCompositeFractalBehaviorIndexParams) (*JurikCompositeFractalBehaviorIndex, error) {
	return newJurikCompositeFractalBehaviorIndex(p.FractalType, p.Smooth,
		p.BarComponent, p.QuoteComponent, p.TradeComponent)
}

func newJurikCompositeFractalBehaviorIndex(fractalType, smooth int,
	bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) (*JurikCompositeFractalBehaviorIndex, error) {
	const (
		invalid = "invalid jurik composite fractal behavior index parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "cfb(%d,%d%s)"
	)

	var (
		mnemonic  string
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	if fractalType < 1 || fractalType > 4 {
		return nil, fmt.Errorf(fmts, invalid, "fractal type should be between 1 and 4")
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

	mnemonic = fmt.Sprintf(fmtn, fractalType, smooth, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Jurik composite fractal behavior index " + mnemonic

	depths := depthSets[fractalType-1]
	numCh := len(depths)

	auxInstances := make([]*cfbAux, numCh)
	for i, d := range depths {
		auxInstances[i] = newCFBAux(d)
	}

	auxWindows := make([][]float64, numCh)
	for i := range auxWindows {
		auxWindows[i] = make([]float64, smooth)
	}

	cfb := &JurikCompositeFractalBehaviorIndex{
		paramFractal: fractalType,
		paramSmooth:  smooth,
		numChannels:  numCh,
		auxInstances: auxInstances,
		auxWindows:   auxWindows,
		er23:         make([]float64, numCh),
		er19:         20,
	}

	cfb.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, cfb.Update)

	return cfb, nil
}

// IsPrimed indicates whether the indicator is primed.
func (s *JurikCompositeFractalBehaviorIndex) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *JurikCompositeFractalBehaviorIndex) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.JurikCompositeFractalBehaviorIndex,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the CFB indicator given the next sample.
func (s *JurikCompositeFractalBehaviorIndex) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.bar++

	// Feed all aux instances.
	auxValues := make([]float64, s.numChannels)
	for i, aux := range s.auxInstances {
		auxValues[i] = aux.update(sample)
	}

	// Bar 0 in reference outputs 0.0 → NaN for streaming. Reference bar 0 = our bar 1.
	if s.bar == 1 {
		return math.NaN()
	}

	// Update running averages (er23). Reference starts main loop at Bar=1.
	// Reference: if Bar <= er29 (Smooth), recompute from scratch; else sliding window.
	// In our terms: bar-1 is the reference Bar (since bar=1 → refBar=0, bar=2 → refBar=1).
	refBar := s.bar - 1

	smooth := s.paramSmooth

	if refBar <= smooth {
		// Growing window: accumulate and divide by refBar.
		// Store current aux values in window.
		winPos := s.auxWinIdx
		for i := 0; i < s.numChannels; i++ {
			s.auxWindows[i][winPos] = auxValues[i]
		}
		s.auxWinIdx = (s.auxWinIdx + 1) % smooth
		s.auxWinLen = refBar

		// Recompute sums from scratch (like reference).
		for i := 0; i < s.numChannels; i++ {
			sum := 0.0
			for j := 0; j < refBar; j++ {
				// Reference: er23[i] += er_i[Bar-er20] for er20=0..Bar-1
				// That means sum of aux values from current bar going back.
				// In our ring: positions from (auxWinIdx-1) going back refBar items.
				pos := (s.auxWinIdx - 1 - j + smooth*2) % smooth
				sum += s.auxWindows[i][pos]
			}
			s.er23[i] = sum / float64(refBar)
		}
	} else {
		// Sliding window: er23[i] += (new - old) / smooth
		winPos := s.auxWinIdx
		for i := 0; i < s.numChannels; i++ {
			oldVal := s.auxWindows[i][winPos]
			s.auxWindows[i][winPos] = auxValues[i]
			s.er23[i] += (auxValues[i] - oldVal) / float64(smooth)
		}
		s.auxWinIdx = (s.auxWinIdx + 1) % smooth
	}

	// Compute weighted composite (only when refBar > 5).
	if refBar > 5 {
		n := s.numChannels
		er22 := make([]float64, n)

		// Odd-indexed channels (descending): n-1, n-3, n-5, ...
		er15 := 1.0
		for idx := n - 1; idx >= 1; idx -= 2 {
			er22[idx] = er15 * s.er23[idx]
			er15 *= (1 - er22[idx])
		}

		// Even-indexed channels (descending): n-2, n-4, ..., 0
		er16 := 1.0
		for idx := n - 2; idx >= 0; idx -= 2 {
			er22[idx] = er16 * s.er23[idx]
			er16 *= (1 - er22[idx])
		}

		// Weighted sum.
		er17 := 0.0
		er18 := 0.0

		for idx := 0; idx < n; idx++ {
			sq := er22[idx] * er22[idx]
			er18 += sq
			// Weight: even indices use weightsEven[idx/2], odd use weightsOdd[idx/2]
			if idx%2 == 0 {
				er17 += sq * weightsEven[idx/2]
			} else {
				er17 += sq * weightsOdd[idx/2]
			}
		}

		if er18 == 0 {
			s.er19 = 0
		} else {
			s.er19 = er17 / er18
		}
	}

	if !s.primed {
		// Reference outputs 20.0 for bars 1-5 (refBar 1-5).
		if refBar > 5 {
			s.primed = true
		}
	}

	return s.er19
}
