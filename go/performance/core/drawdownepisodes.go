package core

import (
	"zpano/streamingkbn"
)

// DrawdownEpisode is a single high-water-mark drawdown episode.
type DrawdownEpisode struct {
	// Depth is the maximum drawdown depth, expressed as a decimal return.
	Depth float64
	// FromIdx is the index of the first underwater observation.
	FromIdx int
	// TroughIdx is the index of the deepest observation.
	TroughIdx int
	// ToIdx is the recovery observation index if the episode recovered;
	// otherwise the last observation index currently available.
	ToIdx int
	// Recovered is true if the high-water mark was recovered at ToIdx,
	// false if the series ended while the drawdown was still open.
	Recovered bool
}

// DrawdownEpisodes is a streaming high-water-mark drawdown episode tracker.
//
// It consumes drawdown observations produced by [HighWaterMarkDrawdown]
// and maintains the corresponding drawdown episodes incrementally.
//
// Drawdowns are expected as decimals, for example -0.025 for a 2.5%
// drawdown and 0.0 for an observation at a high-water mark.
//
// Indices are positions in the sequence of observations passed to Update
// since the last Reset/Recalculate. For a rolling window, call
// Recalculate with the window's drawdowns whenever they are recomputed,
// so indices refer to the window.
//
// A drawdown episode begins with the first negative drawdown and remains
// open until a non-negative drawdown is observed. The episode is
// considered recovered when the drawdown reaches zero or becomes
// positive.
//
// The zero value is ready to use.
type DrawdownEpisodes struct {
	episodes []DrawdownEpisode

	// Current open drawdown episode (valid when open is true).
	open          bool
	currentFrom   int
	currentTrough int
	currentDepth  float64

	// Number of observations processed.
	count int

	// Running depth aggregates.
	sumDepth        streamingkbn.KleinKBNAccumulator
	sumDepthSquared streamingkbn.KleinKBNAccumulator
	// Running episode length/peak-to-trough/recovery aggregates.
	sumLength       int
	sumPeakToTrough int
	sumRecovery     int
}

// NewDrawdownEpisodes returns a new empty DrawdownEpisodes.
func NewDrawdownEpisodes() *DrawdownEpisodes {
	return &DrawdownEpisodes{}
}

// Reset resets the episode tracker to its initial empty state.
func (d *DrawdownEpisodes) Reset() {
	d.episodes = d.episodes[:0]
	d.open = false
	d.currentFrom = 0
	d.currentTrough = 0
	d.currentDepth = 0.0
	d.count = 0
	d.sumDepth.Reset()
	d.sumDepthSquared.Reset()
	d.sumLength = 0
	d.sumPeakToTrough = 0
	d.sumRecovery = 0
}

// Update adds one drawdown observation, expressed as a decimal.
// Drawdowns must be non-positive, although non-negative values are
// accepted and treated as recovery or high-water-mark observations.
func (d *DrawdownEpisodes) Update(drawdown float64) {
	idx := d.count
	d.count++

	if drawdown < 0.0 {
		// We are underwater.
		if !d.open {
			// Start a new drawdown episode.
			d.open = true
			d.currentFrom = idx
			d.currentTrough = idx
			d.currentDepth = drawdown
		} else if drawdown < d.currentDepth {
			// New trough within the current episode.
			d.currentTrough = idx
			d.currentDepth = drawdown
		}
		return
	}

	// We are at or above the high-water mark.
	if d.open {
		// Close the current drawdown episode.
		d.episodes = append(d.episodes, DrawdownEpisode{
			Depth: d.currentDepth, FromIdx: d.currentFrom,
			TroughIdx: d.currentTrough, ToIdx: idx, Recovered: true,
		})
		d.sumDepth.Update(d.currentDepth)
		d.sumDepthSquared.Update(d.currentDepth * d.currentDepth)
		d.sumLength += idx - d.currentFrom + 1
		d.sumPeakToTrough += d.currentTrough - d.currentFrom + 1
		d.sumRecovery += idx - d.currentTrough + 1
		d.open = false
		d.currentFrom = 0
		d.currentTrough = 0
		d.currentDepth = 0.0
	}
}

// Recalculate rebuilds all episodes from the drawdown observations
// history.
func (d *DrawdownEpisodes) Recalculate(drawdowns []float64) {
	d.Reset()
	for _, dd := range drawdowns {
		d.Update(dd)
	}
}

// Episodes returns the drawdown episodes currently known to the tracker
// as a newly allocated slice.
//
// If the latest drawdown episode is still open, it is included using the
// last processed observation as ToIdx and with Recovered=false.
func (d *DrawdownEpisodes) Episodes() []DrawdownEpisode {
	episodes := make([]DrawdownEpisode, len(d.episodes), len(d.episodes)+1)
	copy(episodes, d.episodes)
	if d.open {
		episodes = append(episodes, DrawdownEpisode{
			Depth: d.currentDepth, FromIdx: d.currentFrom,
			TroughIdx: d.currentTrough, ToIdx: d.count - 1, Recovered: false,
		})
	}
	return episodes
}

// Depths returns the drawdown episode depths currently known to the
// tracker, including the depth so far of an open episode, as a newly
// allocated slice.
func (d *DrawdownEpisodes) Depths() []float64 {
	depths := make([]float64, len(d.episodes), len(d.episodes)+1)
	for i, e := range d.episodes {
		depths[i] = e.Depth
	}
	if d.open {
		depths = append(depths, d.currentDepth)
	}
	return depths
}

// AverageEpisodeDrawdown returns the mean magnitude of the observed
// discrete episode drawdowns (0.0 when there are none).
func (d *DrawdownEpisodes) AverageEpisodeDrawdown() float64 {
	sumDepth := d.sumDepth.Value()
	count := len(d.episodes)
	if d.open {
		sumDepth += d.currentDepth
		count++
	}
	if count > 0 {
		return -sumDepth / float64(count)
	}
	return 0.0
}

// AverageEpisodeDrawdownSquared returns the sum of squared episode depths
// divided by the number of observations (not the number of episodes), as
// in PerformanceAnalytics DrawdownDeviation (0.0 when empty).
func (d *DrawdownEpisodes) AverageEpisodeDrawdownSquared() float64 {
	sumDepthSquared := d.sumDepthSquared.Value()
	count := d.count
	if count == 0 {
		return 0.0
	}
	if d.open {
		sumDepthSquared += d.currentDepth * d.currentDepth
	}
	return sumDepthSquared / float64(count)
}

// AverageEpisodeLength returns the mean length of the observed discrete
// drawdown episodes (0.0 when there are none).
func (d *DrawdownEpisodes) AverageEpisodeLength() float64 {
	sumLength := d.sumLength
	count := len(d.episodes)
	if d.open {
		sumLength += d.count - d.currentFrom
		count++
	}
	if count > 0 {
		return float64(sumLength) / float64(count)
	}
	return 0.0
}

// AverageEpisodePeakToTrough returns the mean peak-to-trough length of
// the observed discrete drawdown episodes (0.0 when there are none).
func (d *DrawdownEpisodes) AverageEpisodePeakToTrough() float64 {
	sumPeakToTrough := d.sumPeakToTrough
	count := len(d.episodes)
	if d.open {
		sumPeakToTrough += d.currentTrough - d.currentFrom + 1
		count++
	}
	if count > 0 {
		return float64(sumPeakToTrough) / float64(count)
	}
	return 0.0
}

// AverageEpisodeRecovery returns the mean recovery length of the observed
// discrete drawdown episodes (0.0 when there are none).
func (d *DrawdownEpisodes) AverageEpisodeRecovery() float64 {
	sumRecovery := d.sumRecovery
	count := len(d.episodes)
	if d.open {
		sumRecovery += d.count - d.currentTrough
		count++
	}
	if count > 0 {
		return float64(sumRecovery) / float64(count)
	}
	return 0.0
}
