package directionalmovementminus

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

const (
	dmmMnemonic    = "-dm"
	dmmDescription = "Directional Movement Minus"
)

// DirectionalMovementMinus is Welles Wilder's Directional Movement Minus indicator.
//
// The directional movement was developed in 1978 by Welles Wilder as an indication of trend strength.
//
// The calculation of the directional movement (+DM and −DM) is as follows:
//   - UpMove = today's high − yesterday's high
//   - DownMove = yesterday's low − today's low
//   - if DownMove > UpMove and DownMove > 0, then −DM = DownMove, else −DM = 0
//
// When the length is greater than 1, Wilder's smoothing method is applied:
//
//	Today's −DM(n) = Previous −DM(n) − Previous −DM(n)/n + Today's −DM(1)
//
// The indicator is not primed during the first length updates.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type DirectionalMovementMinus struct {
	mu           sync.RWMutex
	length       int
	noSmoothing  bool
	count        int
	previousHigh float64
	previousLow  float64
	value        float64
	accumulator  float64
	primed       bool
}

// NewDirectionalMovementMinus returns a new instance of the Directional Movement Minus indicator.
func NewDirectionalMovementMinus(length int) (*DirectionalMovementMinus, error) {
	if length < 1 {
		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
	}

	return &DirectionalMovementMinus{
		length:      length,
		noSmoothing: length == 1,
		value:       math.NaN(),
	}, nil
}

// Length returns the length parameter.
func (s *DirectionalMovementMinus) Length() int {
	return s.length
}

// IsPrimed indicates whether the indicator is primed.
func (s *DirectionalMovementMinus) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *DirectionalMovementMinus) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.DirectionalMovementMinus,
		dmmMnemonic,
		dmmDescription,
		[]core.OutputText{
			{Mnemonic: dmmMnemonic, Description: dmmDescription},
		},
	)
}

// Update updates the Directional Movement Minus given the next bar's high and low values.
func (s *DirectionalMovementMinus) Update(high, low float64) float64 {
	if math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	if high < low {
		high, low = low, high
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.noSmoothing {
		if s.primed {
			deltaMinus := s.previousLow - low
			deltaPlus := high - s.previousHigh

			if deltaMinus > 0 && deltaPlus < deltaMinus {
				s.value = deltaMinus
			} else {
				s.value = 0
			}
		} else {
			if s.count > 0 {
				deltaMinus := s.previousLow - low
				deltaPlus := high - s.previousHigh

				if deltaMinus > 0 && deltaPlus < deltaMinus {
					s.value = deltaMinus
				} else {
					s.value = 0
				}

				s.primed = true
			}

			s.count++
		}
	} else {
		if s.primed {
			deltaMinus := s.previousLow - low
			deltaPlus := high - s.previousHigh

			if deltaMinus > 0 && deltaPlus < deltaMinus {
				s.accumulator += -s.accumulator/float64(s.length) + deltaMinus
			} else {
				s.accumulator += -s.accumulator / float64(s.length)
			}

			s.value = s.accumulator
		} else {
			if s.count > 0 && s.length >= s.count {
				deltaMinus := s.previousLow - low
				deltaPlus := high - s.previousHigh

				if s.length > s.count {
					if deltaMinus > 0 && deltaPlus < deltaMinus {
						s.accumulator += deltaMinus
					}
				} else {
					if deltaMinus > 0 && deltaPlus < deltaMinus {
						s.accumulator += -s.accumulator/float64(s.length) + deltaMinus
					} else {
						s.accumulator += -s.accumulator / float64(s.length)
					}

					s.value = s.accumulator
					s.primed = true
				}
			}

			s.count++
		}
	}

	s.previousLow = low
	s.previousHigh = high

	return s.value
}

// UpdateSample updates the Directional Movement Minus using a single sample value
// as a substitute for high and low.
func (s *DirectionalMovementMinus) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *DirectionalMovementMinus) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *DirectionalMovementMinus) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *DirectionalMovementMinus) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *DirectionalMovementMinus) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}
