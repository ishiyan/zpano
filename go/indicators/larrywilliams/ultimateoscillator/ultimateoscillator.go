package ultimateoscillator

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

const (
	ultoscDescription = "Ultimate Oscillator"
	defaultLength1    = 7
	defaultLength2    = 14
	defaultLength3    = 28
	minLength         = 2
	weight1           = 4.0
	weight2           = 2.0
	weight3           = 1.0
	totalWeight       = weight1 + weight2 + weight3 // 7.0
	hundred           = 100.0
)

// UltimateOscillator is Larry Williams' Ultimate Oscillator indicator.
//
// The Ultimate Oscillator combines three different time periods into a single
// oscillator that measures buying pressure relative to true range. The three
// periods are weighted 4:2:1 (shortest:medium:longest).
//
// The indicator requires bar data (high, low, close) and does not use a single
// bar component. For scalar, quote, and trade updates, the single value is used
// as a substitute for high, low, and close.
//
// Reference:
//
// Williams, Larry (1985). "The Ultimate Oscillator". Technical Analysis of Stocks & Commodities.
type UltimateOscillator struct {
	mu sync.RWMutex

	// Sorted periods: p1 <= p2 <= p3.
	p1, p2, p3 int

	// Previous close for true range / true low calculation.
	previousClose float64

	// Circular buffers for buying pressure and true range values.
	// Size = p3 (longest period).
	bpBuffer    []float64
	trBuffer    []float64
	bufferIndex int

	// Running sums for each period window.
	bpSum1, bpSum2, bpSum3 float64
	trSum1, trSum2, trSum3 float64

	// Count of values received (excluding the first bar which only sets previousClose).
	count  int
	primed bool

	mnemonic string
}

// NewUltimateOscillator returns a new instance of the Ultimate Oscillator indicator.
func NewUltimateOscillator(p *UltimateOscillatorParams) (*UltimateOscillator, error) {
	l1 := p.Length1
	if l1 == 0 {
		l1 = defaultLength1
	}

	l2 := p.Length2
	if l2 == 0 {
		l2 = defaultLength2
	}

	l3 := p.Length3
	if l3 == 0 {
		l3 = defaultLength3
	}

	if l1 < minLength {
		return nil, fmt.Errorf("length1 must be >= %d, got %d", minLength, l1)
	}

	if l2 < minLength {
		return nil, fmt.Errorf("length2 must be >= %d, got %d", minLength, l2)
	}

	if l3 < minLength {
		return nil, fmt.Errorf("length3 must be >= %d, got %d", minLength, l3)
	}

	// Sort the three periods ascending.
	s1, s2, s3 := sortThree(l1, l2, l3)

	mnemonic := fmt.Sprintf("ultosc(%d, %d, %d)", l1, l2, l3)

	return &UltimateOscillator{
		p1:            s1,
		p2:            s2,
		p3:            s3,
		previousClose: math.NaN(),
		bpBuffer:      make([]float64, s3),
		trBuffer:      make([]float64, s3),
		mnemonic:      mnemonic,
	}, nil
}

// sortThree returns three ints sorted ascending.
func sortThree(a, b, c int) (int, int, int) {
	if a > b {
		a, b = b, a
	}

	if b > c {
		b, c = c, b
	}

	if a > b {
		a, b = b, a
	}

	return a, b, c
}

// IsPrimed indicates whether the indicator is primed.
func (s *UltimateOscillator) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *UltimateOscillator) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.UltimateOscillator,
		s.mnemonic,
		ultoscDescription+" "+s.mnemonic,
		[]core.OutputText{
			{Mnemonic: s.mnemonic, Description: ultoscDescription + " " + s.mnemonic},
		},
	)
}

// Update updates the Ultimate Oscillator given the next bar's close, high, and low values.
func (s *UltimateOscillator) Update(close, high, low float64) float64 {
	if math.IsNaN(close) || math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// First bar: just store close, return NaN.
	if math.IsNaN(s.previousClose) {
		s.previousClose = close

		return math.NaN()
	}

	// Calculate buying pressure and true range.
	trueLow := math.Min(low, s.previousClose)
	bp := close - trueLow

	tr := high - low
	if d := math.Abs(s.previousClose - high); d > tr {
		tr = d
	}

	if d := math.Abs(s.previousClose - low); d > tr {
		tr = d
	}

	s.previousClose = close

	s.count++

	// Remove trailing values BEFORE storing the new value in the circular buffer,
	// because for p3 the old index equals bufferIndex (the buffer wraps exactly).
	if s.count > s.p1 {
		oldIndex := (s.bufferIndex - s.p1 + s.p3) % s.p3
		s.bpSum1 -= s.bpBuffer[oldIndex]
		s.trSum1 -= s.trBuffer[oldIndex]
	}

	if s.count > s.p2 {
		oldIndex := (s.bufferIndex - s.p2 + s.p3) % s.p3
		s.bpSum2 -= s.bpBuffer[oldIndex]
		s.trSum2 -= s.trBuffer[oldIndex]
	}

	if s.count > s.p3 {
		oldIndex := (s.bufferIndex - s.p3 + s.p3) % s.p3
		s.bpSum3 -= s.bpBuffer[oldIndex]
		s.trSum3 -= s.trBuffer[oldIndex]
	}

	// Add to running sums.
	s.bpSum1 += bp
	s.bpSum2 += bp
	s.bpSum3 += bp
	s.trSum1 += tr
	s.trSum2 += tr
	s.trSum3 += tr

	// Store in circular buffer (after subtraction so p3 trailing reads the old value).
	s.bpBuffer[s.bufferIndex] = bp
	s.trBuffer[s.bufferIndex] = tr

	// Advance buffer index.
	s.bufferIndex = (s.bufferIndex + 1) % s.p3

	// Need at least p3 values (the longest period) to produce output.
	if s.count < s.p3 {
		return math.NaN()
	}

	s.primed = true

	// Calculate output.
	var output float64

	if s.trSum1 != 0 {
		output += weight1 * (s.bpSum1 / s.trSum1)
	}

	if s.trSum2 != 0 {
		output += weight2 * (s.bpSum2 / s.trSum2)
	}

	if s.trSum3 != 0 {
		output += weight3 * (s.bpSum3 / s.trSum3)
	}

	return hundred * (output / totalWeight)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *UltimateOscillator) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *UltimateOscillator) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.Close, sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *UltimateOscillator) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *UltimateOscillator) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v, v)}

	return output
}
