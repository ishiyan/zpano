package directionalmovementminus

import (
	"fmt"
	"math"
	"sync"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
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
func (d *DirectionalMovementMinus) Length() int {
	return d.length
}

// IsPrimed indicates whether the indicator is primed.
func (d *DirectionalMovementMinus) IsPrimed() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.primed
}

// Metadata describes the output data of the indicator.
func (d *DirectionalMovementMinus) Metadata() core.Metadata {
	return core.Metadata{
		Type:        core.DirectionalMovementMinus,
		Mnemonic:    dmmMnemonic,
		Description: dmmDescription,
		Outputs: []outputs.Metadata{
			{
				Kind:        int(DirectionalMovementMinusValue),
				Type:        outputs.ScalarType,
				Mnemonic:    dmmMnemonic,
				Description: dmmDescription,
			},
		},
	}
}

// Update updates the Directional Movement Minus given the next bar's high and low values.
func (d *DirectionalMovementMinus) Update(high, low float64) float64 {
	if math.IsNaN(high) || math.IsNaN(low) {
		return math.NaN()
	}

	if high < low {
		high, low = low, high
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.noSmoothing {
		if d.primed {
			deltaMinus := d.previousLow - low
			deltaPlus := high - d.previousHigh

			if deltaMinus > 0 && deltaPlus < deltaMinus {
				d.value = deltaMinus
			} else {
				d.value = 0
			}
		} else {
			if d.count > 0 {
				deltaMinus := d.previousLow - low
				deltaPlus := high - d.previousHigh

				if deltaMinus > 0 && deltaPlus < deltaMinus {
					d.value = deltaMinus
				} else {
					d.value = 0
				}

				d.primed = true
			}

			d.count++
		}
	} else {
		if d.primed {
			deltaMinus := d.previousLow - low
			deltaPlus := high - d.previousHigh

			if deltaMinus > 0 && deltaPlus < deltaMinus {
				d.accumulator += -d.accumulator/float64(d.length) + deltaMinus
			} else {
				d.accumulator += -d.accumulator / float64(d.length)
			}

			d.value = d.accumulator
		} else {
			if d.count > 0 && d.length >= d.count {
				deltaMinus := d.previousLow - low
				deltaPlus := high - d.previousHigh

				if d.length > d.count {
					if deltaMinus > 0 && deltaPlus < deltaMinus {
						d.accumulator += deltaMinus
					}
				} else {
					if deltaMinus > 0 && deltaPlus < deltaMinus {
						d.accumulator += -d.accumulator/float64(d.length) + deltaMinus
					} else {
						d.accumulator += -d.accumulator / float64(d.length)
					}

					d.value = d.accumulator
					d.primed = true
				}
			}

			d.count++
		}
	}

	d.previousLow = low
	d.previousHigh = high

	return d.value
}

// UpdateSample updates the Directional Movement Minus using a single sample value
// as a substitute for high and low.
func (d *DirectionalMovementMinus) UpdateSample(sample float64) float64 {
	return d.Update(sample, sample)
}

// UpdateScalar updates the indicator given the next scalar sample.
func (d *DirectionalMovementMinus) UpdateScalar(sample *entities.Scalar) core.Output {
	v := sample.Value

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(v, v)}

	return output
}

// UpdateBar updates the indicator given the next bar sample.
func (d *DirectionalMovementMinus) UpdateBar(sample *entities.Bar) core.Output {
	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(sample.High, sample.Low)}

	return output
}

// UpdateQuote updates the indicator given the next quote sample.
func (d *DirectionalMovementMinus) UpdateQuote(sample *entities.Quote) core.Output {
	v := (sample.Bid + sample.Ask) / 2 //nolint:mnd

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(v, v)}

	return output
}

// UpdateTrade updates the indicator given the next trade sample.
func (d *DirectionalMovementMinus) UpdateTrade(sample *entities.Trade) core.Output {
	v := sample.Price

	output := make([]any, 1)
	output[0] = entities.Scalar{Time: sample.Time, Value: d.Update(v, v)}

	return output
}
