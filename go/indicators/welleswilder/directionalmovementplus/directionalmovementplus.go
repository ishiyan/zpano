package directionalmovementplus

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
)

const (
	dmpMnemonic    = "+dm"
	dmpDescription = "Directional Movement Plus"
)

// DirectionalMovementPlus is Welles Wilder's Directional Movement Plus indicator.
//
// The directional movement was developed in 1978 by Welles Wilder as an indication of trend strength.
//
// The calculation of the directional movement (+DM and −DM) is as follows:
//   - UpMove = today's high − yesterday's high
//   - DownMove = yesterday's low − today's low
//   - if UpMove > DownMove and UpMove > 0, then +DM = UpMove, else +DM = 0
//
// When the length is greater than 1, Wilder's smoothing method is applied:
//
//	Today's +DM(n) = Previous +DM(n) − Previous +DM(n)/n + Today's +DM(1)
//
// The indicator is not primed during the first length updates.
//
// Reference:
//
// Wilder, J. Welles Jr. (1978). New Concepts in Technical Trading Systems.
type DirectionalMovementPlus struct {
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

// NewDirectionalMovementPlus returns a new instance of the Directional Movement Plus indicator.
func NewDirectionalMovementPlus(length int) (*DirectionalMovementPlus, error) {
	if length < 1 {
		return nil, fmt.Errorf("invalid length %d: must be >= 1", length)
	}

	return &DirectionalMovementPlus{
		length:      length,
		noSmoothing: length == 1,
		value:       math.NaN(),
	}, nil
}

// Length returns the length parameter.
func (s *DirectionalMovementPlus) Length() int {
	return s.length
}

// IsPrimed indicates whether the indicator is primed.
func (s *DirectionalMovementPlus) IsPrimed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.primed
}

// Metadata describes the output data of the indicator.
func (s *DirectionalMovementPlus) Metadata() core.Metadata {
	return core.BuildMetadata(
		core.DirectionalMovementPlus,
		dmpMnemonic,
		dmpDescription,
		[]core.OutputText{
			{Mnemonic: dmpMnemonic, Description: dmpDescription},
		},
	)
}

// Update updates the Directional Movement Plus given the next bar's high and low values.
func (s *DirectionalMovementPlus) Update(high, low float64) float64 {
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
			deltaPlus := high - s.previousHigh
			deltaMinus := s.previousLow - low

			if deltaPlus > 0 && deltaPlus > deltaMinus {
				s.value = deltaPlus
			} else {
				s.value = 0
			}
		} else {
			if s.count > 0 {
				deltaPlus := high - s.previousHigh
				deltaMinus := s.previousLow - low

				if deltaPlus > 0 && deltaPlus > deltaMinus {
					s.value = deltaPlus
				} else {
					s.value = 0
				}

				s.primed = true
			}

			s.count++
		}
	} else {
		if s.primed {
			deltaPlus := high - s.previousHigh
			deltaMinus := s.previousLow - low

			if deltaPlus > 0 && deltaPlus > deltaMinus {
				s.accumulator += -s.accumulator/float64(s.length) + deltaPlus
			} else {
				s.accumulator += -s.accumulator / float64(s.length)
			}

			s.value = s.accumulator
		} else {
			if s.count > 0 && s.length >= s.count {
				deltaPlus := high - s.previousHigh
				deltaMinus := s.previousLow - low

				if s.length > s.count {
					if deltaPlus > 0 && deltaPlus > deltaMinus {
						s.accumulator += deltaPlus
					}
				} else {
					if deltaPlus > 0 && deltaPlus > deltaMinus {
						s.accumulator += -s.accumulator/float64(s.length) + deltaPlus
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

// UpdateSample updates the Directional Movement Plus using a single sample value
// as a substitute for high and low.
func (s *DirectionalMovementPlus) UpdateSample(sample float64) float64 {
	return s.Update(sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (s *DirectionalMovementPlus) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (s *DirectionalMovementPlus) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (s *DirectionalMovementPlus) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (s *DirectionalMovementPlus) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: s.Update(v, v)}

	return output
}
