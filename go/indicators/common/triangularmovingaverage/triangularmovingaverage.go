package triangularmovingaverage

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

// TriangularMovingAverage computes the triangular moving average (TRIMA) like a weighted moving average.
// Instead of the WMA who put more weight on the latest sample, the TRIMA puts more weight on the data
// in the middle of the window.
//
// Using algebra, it can be demonstrated that the TRIMA is equivalent to doing a SMA of a SMA.
// The following explain the rules.
//
//	➊ When the period π is even, TRIMA(x,π) = SMA(SMA(x,π/2), (π/2)+1).
//	➋ When the period π is odd, TRIMA(x,π) = SMA(SMA(x,(π+1)/2), (π+1)/2).
//
// The SMA of a SMA is the algorithm generally found in books.
//
// TradeStation deviate from the generally accepted implementation by making the TRIMA to be as follows:
//
//	TRIMA(x,π) = SMA(SMA(x, (int)(π/2)+1), (int)(π/2)+1).
//
// This formula is done regardless if the period is even or odd. In other words:
//
//	➊ A period of 4 becomes TRIMA(x,4) = SMA(SMA(x,3), 3).
//	➋ A period of 5 becomes TRIMA(x,5) = SMA(SMA(x,3), 3).
//	➌ A period of 6 becomes TRIMA(x,6) = SMA(SMA(x,4), 4).
//	➍ A period of 7 becomes TRIMA(x,7) = SMA(SMA(x,4), 4).
//
// The Metastock implementation is the same as the generally accepted one.
//
// To optimize speed, this implementation uses a better algorithm than the usual SMA of a SMA.
// The calculation from one TRIMA value to the next is done by doing 4 little adjustments.
//
// The following show a TRIMA 4-period:
//
//	TRIMA at time δ: ((1*α)+(2*β)+(2*γ)+(1*δ)) / 6
//	TRIMA at time ε: ((1*β)+(2*γ)+(2*δ)+(1*ε)) / 6
//
// To go from TRIMA δ to ε, the following is done:
//
//	➊ α and β are subtract from the numerator.
//	➋ δ is added to the numerator.
//	➌ ε is added to the numerator.
//	➍ TRIMA is calculated by doing numerator / 6.
//	➎ Sequence is repeated for the next output.
type TriangularMovingAverage struct {
	mu sync.RWMutex
	core.LineIndicator
	factor           float64
	numerator        float64
	numeratorSub     float64
	numeratorAdd     float64
	window           []float64
	windowLength     int
	windowLengthHalf int
	windowCount      int
	isOdd            bool
	primed           bool
}

// NewTriangularMovingAverage returns an instnce of the indicator created using supplied parameters.
func NewTriangularMovingAverage(p *TriangularMovingAverageParams) (*TriangularMovingAverage, error) { //nolint:funlen
	const (
		invalid = "invalid triangular moving average parameters"
		fmts    = "%s: %s"
		fmtw    = "%s: %w"
		fmtn    = "trima(%d%s)"
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
		factor    float64
		err       error
		barFunc   entities.BarFunc
		quoteFunc entities.QuoteFunc
		tradeFunc entities.TradeFunc
	)

	lengthHalf := length >> 1
	l := 1 + lengthHalf
	isOdd := length%2 == 1 //nolint:gomnd

	if isOdd {
		// Let period = 5 and l=(int)(period/2), then the formula for a "triangular" series is:
		// 1+2+3+2+1 = l*(l+1) + l+1 = (l+1)*(l+1) = 3*3 = 9.
		factor = 1. / float64(l*l) //nolint:gomnd
	} else {
		// Let period = 6 and l=(int)(period/2), then  the formula for a "triangular" series is:
		// 1+2+3+3+2+1 = l*(l+1) = 3*4 = 12.
		factor = 1. / float64(lengthHalf*l) //nolint:gomnd
		lengthHalf--
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
	mnemonic := fmt.Sprintf(fmtn, length, core.ComponentTripleMnemonic(bc, qc, tc))
	desc := "Triangular moving average " + mnemonic

	trima := &TriangularMovingAverage{
		factor:           factor,
		window:           make([]float64, length),
		windowLength:     length,
		windowLengthHalf: lengthHalf,
		isOdd:            isOdd,
	}

	trima.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, trima.Update)

	return trima, nil
}

// IsPrimed indicates whether an indicator is primed.
func (s *TriangularMovingAverage) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes an output data of the indicator.
// It always has a single scalar output -- the calculated value of the triangular moving average.
func (s *TriangularMovingAverage) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.TriangularMovingAverage,
		s.LineIndicator.Mnemonic,
		s.LineIndicator.Description,
		[]core.OutputText{
			{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
		},
	)
}

// Update updates the value of the moving average given the next sample.
//
// The indicator is not primed during the first ℓ-1 updates.
func (s *TriangularMovingAverage) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return sample
	}

	temp := sample

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.primed {
		s.numerator -= s.numeratorSub
		s.numeratorSub -= s.window[0]

		j := s.windowLength - 1
		for i := 0; i < j; i++ {
			s.window[i] = s.window[i+1]
		}

		s.window[j] = temp
		temp = s.window[s.windowLengthHalf]
		s.numeratorSub += temp

		if s.isOdd { // The logic for an odd length.
			s.numerator += s.numeratorAdd
			s.numeratorAdd -= temp
		} else { // The logic for an even length.
			s.numeratorAdd -= temp
			s.numerator += s.numeratorAdd
		}

		temp = sample
		s.numeratorAdd += temp
		s.numerator += temp
	} else {
		s.window[s.windowCount] = temp
		s.windowCount++

		if s.windowLength > s.windowCount {
			return math.NaN()
		}

		for i := s.windowLengthHalf; i >= 0; i-- {
			s.numeratorSub += s.window[i]
			s.numerator += s.numeratorSub
		}

		for i := s.windowLengthHalf + 1; i < s.windowLength; i++ {
			s.numeratorAdd += s.window[i]
			s.numerator += s.numeratorAdd
		}

		s.primed = true
	}

	return s.numerator * s.factor
}
