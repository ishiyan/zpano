package parabolicstopandreverse

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

const (
	defaultAccelerationInit = 0.02
	defaultAccelerationStep = 0.02
	defaultAccelerationMax  = 0.20
)

// ParabolicStopAndReverse is Welles Wilder's Parabolic Stop And Reverse (SAR) indicator.
//
// The Parabolic SAR provides potential entry and exit points. It places dots above or below
// the price to indicate the direction of the trend. When the dots are below the price, it
// signals a long (upward) trend; when above, it signals a short (downward) trend.
//
// This is the "extended" version (SAREXT) which supports separate acceleration factor
// parameters for long and short directions, an optional start value to force the initial
// direction, and a percent offset on reversal. The output is signed: positive values
// indicate long positions, negative values indicate short positions.
//
// Algorithm overview (from Welles Wilder / TA-Lib):
//
// The implementation of SAR has been somewhat open to interpretation since Wilder
// (the original author) did not define a precise algorithm on how to bootstrap it.
//
// Initial trade direction:
//   - If StartValue == 0 (auto): the direction is determined by comparing +DM and -DM
//     between the first and second bars. If -DM > +DM the initial direction is short,
//     otherwise long. Ties default to long.
//   - If StartValue > 0: force long at the specified SAR value.
//   - If StartValue < 0: force short at abs(StartValue) as the initial SAR value.
//
// Initial extreme point and SAR:
//   - For auto mode: the first bar's high/low is used as the initial SAR, and the second
//     bar's high (long) or low (short) is the initial extreme point. This is the same
//     approach used by Metastock.
//   - For forced mode: the SAR is set to the specified start value.
//
// On each subsequent bar the SAR is updated by the acceleration factor (AF) times the
// difference between the extreme point (EP) and the current SAR. The AF starts at the
// initial value and increases by the step value each time a new EP is reached, up to a
// maximum. When a reversal occurs (price penetrates the SAR), the position flips, the
// SAR is reset to the EP, and the AF is reset.
//
// Reference:
//
// Wilder, J. Welles. "New Concepts in Technical Trading Systems", 1978.
type ParabolicStopAndReverse struct {
	mu sync.RWMutex
	core.LineIndicator

	// Parameters (resolved from defaults).
	startValue      float64
	offsetOnReverse float64
	afInitLong      float64
	afStepLong      float64
	afMaxLong       float64
	afInitShort     float64
	afStepShort     float64
	afMaxShort      float64

	// State.
	count    int     // number of bars received
	isLong   bool    // current direction
	sar      float64 // current SAR value
	ep       float64 // extreme point
	afLong   float64 // current acceleration factor (long)
	afShort  float64 // current acceleration factor (short)
	prevHigh float64 // previous bar's high
	prevLow  float64 // previous bar's low
	newHigh  float64 // current bar's high
	newLow   float64 // current bar's low
	primed   bool
}

// NewParabolicStopAndReverse returns an instance of the indicator created using supplied parameters.
func NewParabolicStopAndReverse(p *ParabolicStopAndReverseParams) (*ParabolicStopAndReverse, error) {
	const invalid = "invalid parabolic stop and reverse parameters"

	// Resolve defaults.
	afInitLong := p.AccelerationInitLong
	if afInitLong == 0 {
		afInitLong = defaultAccelerationInit
	}

	afStepLong := p.AccelerationLong
	if afStepLong == 0 {
		afStepLong = defaultAccelerationStep
	}

	afMaxLong := p.AccelerationMaxLong
	if afMaxLong == 0 {
		afMaxLong = defaultAccelerationMax
	}

	afInitShort := p.AccelerationInitShort
	if afInitShort == 0 {
		afInitShort = defaultAccelerationInit
	}

	afStepShort := p.AccelerationShort
	if afStepShort == 0 {
		afStepShort = defaultAccelerationStep
	}

	afMaxShort := p.AccelerationMaxShort
	if afMaxShort == 0 {
		afMaxShort = defaultAccelerationMax
	}

	// Validate: acceleration factors must be positive.
	if afInitLong < 0 || afStepLong < 0 || afMaxLong < 0 {
		return nil, fmt.Errorf("%s: long acceleration factors must be non-negative", invalid)
	}

	if afInitShort < 0 || afStepShort < 0 || afMaxShort < 0 {
		return nil, fmt.Errorf("%s: short acceleration factors must be non-negative", invalid)
	}

	if p.OffsetOnReverse < 0 {
		return nil, fmt.Errorf("%s: offset on reverse must be non-negative", invalid)
	}

	// Clamp: init and step cannot exceed max.
	if afInitLong > afMaxLong {
		afInitLong = afMaxLong
	}

	if afStepLong > afMaxLong {
		afStepLong = afMaxLong
	}

	if afInitShort > afMaxShort {
		afInitShort = afMaxShort
	}

	if afStepShort > afMaxShort {
		afStepShort = afMaxShort
	}

	mnemonic := "sar()"
	desc := "Parabolic Stop And Reverse " + mnemonic

	barFunc, _ := entities.BarComponentFunc(entities.BarClosePrice)
	quoteFunc, _ := entities.QuoteComponentFunc(entities.DefaultQuoteComponent)
	tradeFunc, _ := entities.TradeComponentFunc(entities.DefaultTradeComponent)

	s := &ParabolicStopAndReverse{
		startValue:      p.StartValue,
		offsetOnReverse: p.OffsetOnReverse,
		afInitLong:      afInitLong,
		afStepLong:      afStepLong,
		afMaxLong:       afMaxLong,
		afInitShort:     afInitShort,
		afStepShort:     afStepShort,
		afMaxShort:      afMaxShort,
		afLong:          afInitLong,
		afShort:         afInitShort,
	}

	s.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, s.Update)

	return s, nil
}

// IsPrimed indicates whether the indicator has received enough data to produce valid output.
func (s *ParabolicStopAndReverse) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *ParabolicStopAndReverse) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.ParabolicStopAndReverse,
		Mnemonic:    s.LineIndicator.Mnemonic,
		Description: s.LineIndicator.Description,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(ParabolicStopAndReverseValue),
				Type:        outputs.ScalarType,
				Mnemonic:    s.LineIndicator.Mnemonic,
				Description: s.LineIndicator.Description,
			},
		},
	}
}

// Update updates the indicator with the given scalar sample.
// For scalar updates, high and low are the same, so SAR behaves as if there is no range.
func (s *ParabolicStopAndReverse) Update(sample float64) float64 {
	if math.IsNaN(sample) {
		return math.NaN()
	}

	return s.UpdateHL(sample, sample)
}

// UpdateHL updates the indicator with the given high and low values.
func (s *ParabolicStopAndReverse) UpdateHL(high, low float64) float64 {
	if math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.count++

	// First bar: store high/low, no output yet.
	if s.count == 1 {
		s.newHigh = high
		s.newLow = low

		return math.NaN()
	}

	// Second bar: initialize SAR, EP, and direction.
	if s.count == 2 {
		prevHigh := s.newHigh
		prevLow := s.newLow

		if s.startValue == 0 {
			// Auto-detect direction using MINUS_DM logic.
			minusDM := prevLow - low
			plusDM := high - prevHigh

			if minusDM < 0 {
				minusDM = 0
			}

			if plusDM < 0 {
				plusDM = 0
			}

			s.isLong = minusDM <= plusDM

			if s.isLong {
				s.ep = high
				s.sar = prevLow
			} else {
				s.ep = low
				s.sar = prevHigh
			}
		} else if s.startValue > 0 {
			s.isLong = true
			s.ep = high
			s.sar = s.startValue
		} else {
			s.isLong = false
			s.ep = low
			s.sar = math.Abs(s.startValue)
		}

		// Set newHigh/newLow for the "cheat" to prepare iteration.
		s.newHigh = high
		s.newLow = low
		s.primed = true

		// Fall through to the main loop logic below.
	}

	// Main SAR calculation (bars 2+).
	if s.count >= 2 {
		s.prevLow = s.newLow
		s.prevHigh = s.newHigh
		s.newLow = low
		s.newHigh = high

		if s.count == 2 {
			// On the second call, prevLow/prevHigh are already set above,
			// and newLow/newHigh are the current bar. But the TaLib algorithm
			// reads from todayIdx which starts at startIdx (=1), and the first
			// iteration reads todayIdx's values as newHigh/newLow, then increments.
			// Since we already set prevHigh/prevLow = bar[0] values and
			// newHigh/newLow = bar[1] values in the init above, we need to
			// re-assign to match: the loop iteration uses the SAME bar as init.
			s.prevLow = s.newLow
			s.prevHigh = s.newHigh
		}

		if s.isLong {
			return s.updateLong()
		}

		return s.updateShort()
	}

	return math.NaN()
}

func (s *ParabolicStopAndReverse) updateLong() float64 {
	// Switch to short if the low penetrates the SAR value.
	if s.newLow <= s.sar {
		s.isLong = false
		s.sar = s.ep

		if s.sar < s.prevHigh {
			s.sar = s.prevHigh
		}

		if s.sar < s.newHigh {
			s.sar = s.newHigh
		}

		if s.offsetOnReverse != 0.0 {
			s.sar += s.sar * s.offsetOnReverse
		}

		result := -s.sar

		// Reset short AF and set EP.
		s.afShort = s.afInitShort
		s.ep = s.newLow

		// Calculate the new SAR.
		s.sar = s.sar + s.afShort*(s.ep-s.sar)

		if s.sar < s.prevHigh {
			s.sar = s.prevHigh
		}

		if s.sar < s.newHigh {
			s.sar = s.newHigh
		}

		return result
	}

	// No switch — output the current SAR.
	result := s.sar

	// Adjust AF and EP.
	if s.newHigh > s.ep {
		s.ep = s.newHigh
		s.afLong += s.afStepLong

		if s.afLong > s.afMaxLong {
			s.afLong = s.afMaxLong
		}
	}

	// Calculate the new SAR.
	s.sar = s.sar + s.afLong*(s.ep-s.sar)

	if s.sar > s.prevLow {
		s.sar = s.prevLow
	}

	if s.sar > s.newLow {
		s.sar = s.newLow
	}

	return result
}

func (s *ParabolicStopAndReverse) updateShort() float64 {
	// Switch to long if the high penetrates the SAR value.
	if s.newHigh >= s.sar {
		s.isLong = true
		s.sar = s.ep

		if s.sar > s.prevLow {
			s.sar = s.prevLow
		}

		if s.sar > s.newLow {
			s.sar = s.newLow
		}

		if s.offsetOnReverse != 0.0 {
			s.sar -= s.sar * s.offsetOnReverse
		}

		result := s.sar

		// Reset long AF and set EP.
		s.afLong = s.afInitLong
		s.ep = s.newHigh

		// Calculate the new SAR.
		s.sar = s.sar + s.afLong*(s.ep-s.sar)

		if s.sar > s.prevLow {
			s.sar = s.prevLow
		}

		if s.sar > s.newLow {
			s.sar = s.newLow
		}

		return result
	}

	// No switch — output the negated SAR.
	result := -s.sar

	// Adjust AF and EP.
	if s.newLow < s.ep {
		s.ep = s.newLow
		s.afShort += s.afStepShort

		if s.afShort > s.afMaxShort {
			s.afShort = s.afMaxShort
		}
	}

	// Calculate the new SAR.
	s.sar = s.sar + s.afShort*(s.ep-s.sar)

	if s.sar < s.prevHigh {
		s.sar = s.prevHigh
	}

	if s.sar < s.newHigh {
		s.sar = s.newHigh
	}

	return result
}

// UpdateBar updates the indicator given the next bar sample.
// This shadows LineIndicator.UpdateBar to extract high and low from the bar.
func (s *ParabolicStopAndReverse) UpdateBar(sample *entities.Bar) core.Output {
	value := s.UpdateHL(sample.High, sample.Low)

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: value}

	return output
}
