package core

import (
	"slices"
	"testing"
)

// Drawdown observations (decimals) with two recovered episodes and an
// open one at the end:
//
//	idx:   0     1     2     3     4     5     6     7     8
//	dd:    0  -0.10 -0.20 -0.05    0  -0.03    0  -0.04 -0.06
//	           |--- episode 1 ---|     |-- 2 --|     |-- 3 (open)
var episodeDrawdowns = []float64{0.0, -0.10, -0.20, -0.05, 0.0, -0.03, 0.0, -0.04, -0.06}

var expectedEpisodes = []DrawdownEpisode{
	{Depth: -0.20, FromIdx: 1, TroughIdx: 2, ToIdx: 4, Recovered: true},
	{Depth: -0.03, FromIdx: 5, TroughIdx: 5, ToIdx: 6, Recovered: true},
	{Depth: -0.06, FromIdx: 7, TroughIdx: 8, ToIdx: 8, Recovered: false},
}

func feedEpisodes(tracker *DrawdownEpisodes, drawdowns []float64) *DrawdownEpisodes {
	for _, dd := range drawdowns {
		tracker.Update(dd)
	}
	return tracker
}

func assertEpisodes(t *testing.T, actual, expected []DrawdownEpisode) {
	t.Helper()
	if !slices.Equal(actual, expected) {
		t.Errorf("episodes %+v, expected %+v", actual, expected)
	}
}

func assertExact(t *testing.T, actual, expected float64, msg string) {
	t.Helper()
	if actual != expected {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

func assertEpisodesEmpty(t *testing.T, d *DrawdownEpisodes) {
	t.Helper()
	assertEpisodes(t, d.Episodes(), []DrawdownEpisode{})
	if len(d.Depths()) != 0 {
		t.Errorf("depths %v, expected empty", d.Depths())
	}
	assertExact(t, d.AverageEpisodeDrawdown(), 0.0, "average drawdown")
	assertExact(t, d.AverageEpisodeDrawdownSquared(), 0.0, "average drawdown squared")
	assertExact(t, d.AverageEpisodeLength(), 0.0, "average length")
	assertExact(t, d.AverageEpisodePeakToTrough(), 0.0, "average peak to trough")
	assertExact(t, d.AverageEpisodeRecovery(), 0.0, "average recovery")
}

func TestDrawdownEpisodesEmpty(t *testing.T) {
	t.Parallel()
	assertEpisodesEmpty(t, NewDrawdownEpisodes())
}

func TestDrawdownEpisodesNoDrawdowns(t *testing.T) {
	t.Parallel()
	d := feedEpisodes(NewDrawdownEpisodes(), []float64{0.0, 0.0, 0.0})
	assertEpisodes(t, d.Episodes(), []DrawdownEpisode{})
	assertExact(t, d.AverageEpisodeDrawdown(), 0.0, "average drawdown")
	assertExact(t, d.AverageEpisodeDrawdownSquared(), 0.0, "average drawdown squared")
}

func TestDrawdownEpisodesEpisodes(t *testing.T) {
	t.Parallel()
	d := feedEpisodes(NewDrawdownEpisodes(), episodeDrawdowns)
	assertEpisodes(t, d.Episodes(), expectedEpisodes)
	if !slices.Equal(d.Depths(), []float64{-0.20, -0.03, -0.06}) {
		t.Errorf("depths %v", d.Depths())
	}
}

func TestDrawdownEpisodesAverages(t *testing.T) {
	t.Parallel()
	d := feedEpisodes(NewDrawdownEpisodes(), episodeDrawdowns)
	// Runtime variables so the expressions are evaluated in float64.
	a, b, c := 0.20, 0.03, 0.06
	a2, b2, c2 := 0.04, 0.0009, 0.0036
	eight, five, three := 8.0, 5.0, 3.0
	assertAlmostEqual(t, d.AverageEpisodeDrawdown(), (a+b+c)/3, places(16), "average drawdown")
	// Divided by the number of observations, not episodes.
	assertAlmostEqual(t, d.AverageEpisodeDrawdownSquared(),
		(a2+b2+c2)/float64(len(episodeDrawdowns)), places(16), "average drawdown squared")
	// Lengths 4, 2, 2 (the open episode has no recovery observation).
	assertAlmostEqual(t, d.AverageEpisodeLength(), eight/three, places(15), "average length")
	// Peak to trough 2, 1, 2.
	assertAlmostEqual(t, d.AverageEpisodePeakToTrough(), five/three, places(15), "average peak to trough")
	// Recovery 3, 2, 1.
	assertAlmostEqual(t, d.AverageEpisodeRecovery(), 2.0, places(15), "average recovery")
}

func TestDrawdownEpisodesOpenEpisodeProgress(t *testing.T) {
	t.Parallel()
	d := NewDrawdownEpisodes()
	d.Update(-0.05)
	assertEpisodes(t, d.Episodes(), []DrawdownEpisode{{-0.05, 0, 0, 0, false}})
	d.Update(-0.02) // shallower, the trough doesn't move
	assertEpisodes(t, d.Episodes(), []DrawdownEpisode{{-0.05, 0, 0, 1, false}})
	d.Update(-0.08) // new trough
	assertEpisodes(t, d.Episodes(), []DrawdownEpisode{{-0.08, 0, 2, 2, false}})
	d.Update(0.0) // recovery closes the episode
	assertEpisodes(t, d.Episodes(), []DrawdownEpisode{{-0.08, 0, 2, 3, true}})
}

func TestDrawdownEpisodesRecalculateMatchesIncremental(t *testing.T) {
	t.Parallel()
	incremental := feedEpisodes(NewDrawdownEpisodes(), episodeDrawdowns)
	rebuilt := feedEpisodes(NewDrawdownEpisodes(), []float64{-0.5, -0.7, 0.0})
	rebuilt.Recalculate(episodeDrawdowns)
	assertEpisodes(t, rebuilt.Episodes(), incremental.Episodes())
	assertExact(t, rebuilt.AverageEpisodeDrawdown(), incremental.AverageEpisodeDrawdown(), "average drawdown")
	assertExact(t, rebuilt.AverageEpisodeDrawdownSquared(), incremental.AverageEpisodeDrawdownSquared(),
		"average drawdown squared")
	assertExact(t, rebuilt.AverageEpisodeLength(), incremental.AverageEpisodeLength(), "average length")
	assertExact(t, rebuilt.AverageEpisodeRecovery(), incremental.AverageEpisodeRecovery(), "average recovery")
}

func TestDrawdownEpisodesReset(t *testing.T) {
	t.Parallel()
	d := feedEpisodes(NewDrawdownEpisodes(), episodeDrawdowns)
	d.Reset()
	assertEpisodesEmpty(t, d)
	d.Update(-0.01)
	assertEpisodes(t, d.Episodes(), []DrawdownEpisode{{-0.01, 0, 0, 0, false}})
}
