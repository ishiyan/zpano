//! Streaming high-water-mark drawdown episode tracker.

use std::borrow::Borrow;

use crate::streaming_kbn::KleinKbnAccumulator;

/// A single high-water-mark drawdown episode.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct DrawdownEpisode {
    /// Maximum drawdown depth, expressed as a decimal return.
    pub depth: f64,
    /// Index of the first underwater observation.
    pub from_idx: usize,
    /// Index of the deepest observation.
    pub trough_idx: usize,
    /// Recovery observation index if the episode recovered; otherwise
    /// the last observation index currently available.
    pub to_idx: usize,
    /// True if the high-water mark was recovered at `to_idx`. False if
    /// the series ended while the drawdown was still open.
    pub recovered: bool,
}

impl DrawdownEpisode {
    /// Creates an episode (Python positional constructor order).
    pub fn new(depth: f64, from_idx: usize, trough_idx: usize, to_idx: usize, recovered: bool) -> Self {
        Self { depth, from_idx, trough_idx, to_idx, recovered }
    }
}

/// Streaming high-water-mark drawdown episode tracker.
///
/// Consumes drawdown observations produced by
/// [`HighWaterMarkDrawdown`](super::HighWaterMarkDrawdown) and maintains
/// the corresponding drawdown episodes incrementally.
///
/// Drawdowns are expected as decimals, for example `-0.025` for a 2.5%
/// drawdown and `0.0` for an observation at a high-water mark.
///
/// Indices are positions in the sequence of observations passed to
/// `update()` since the last `reset()`/`recalculate()`. For a rolling
/// window, call `recalculate()` with the window's drawdowns whenever they
/// are recomputed, so indices refer to the window.
///
/// A drawdown episode begins with the first negative drawdown and remains
/// open until a non-negative drawdown is observed. The episode is
/// considered recovered when the drawdown reaches zero or becomes positive.
#[derive(Debug, Clone, Default)]
pub struct DrawdownEpisodes {
    episodes: Vec<DrawdownEpisode>,
    // Current open drawdown episode: (from, trough).
    current: Option<(usize, usize)>,
    current_depth: f64,
    // Number of observations processed.
    count: usize,
    // Running depth aggregates.
    sum_depth: KleinKbnAccumulator,
    sum_depth_squared: KleinKbnAccumulator,
    // Running episode length/peak-to-trough/recovery aggregates.
    sum_length: usize,
    sum_peak_to_trough: usize,
    sum_recovery: usize,
}

impl DrawdownEpisodes {
    /// Creates an empty tracker.
    pub fn new() -> Self {
        Self::default()
    }

    /// Resets the episode tracker to its initial empty state.
    pub fn reset(&mut self) {
        self.episodes.clear();
        self.current = None;
        self.current_depth = 0.0;
        self.count = 0;
        self.sum_depth.reset();
        self.sum_depth_squared.reset();
        self.sum_length = 0;
        self.sum_peak_to_trough = 0;
        self.sum_recovery = 0;
    }

    /// Adds one drawdown observation, expressed as a decimal.
    ///
    /// Drawdowns must be non-positive, although non-negative values are
    /// accepted and treated as recovery or high-water-mark observations.
    pub fn update(&mut self, drawdown: f64) {
        let idx = self.count;
        self.count += 1;

        if drawdown < 0.0 {
            // We are underwater.
            match self.current {
                None => {
                    // Start a new drawdown episode.
                    self.current = Some((idx, idx));
                    self.current_depth = drawdown;
                }
                Some((from, _)) if drawdown < self.current_depth => {
                    // New trough within the current episode.
                    self.current = Some((from, idx));
                    self.current_depth = drawdown;
                }
                _ => {}
            }
            return;
        }

        // We are at or above the high-water mark.
        if let Some((from, trough)) = self.current {
            // Close the current drawdown episode.
            let depth = self.current_depth;
            self.episodes.push(DrawdownEpisode::new(depth, from, trough, idx, true));
            self.sum_depth.update(depth);
            self.sum_depth_squared.update(depth * depth);
            self.sum_length += idx - from + 1;
            self.sum_peak_to_trough += trough - from + 1;
            self.sum_recovery += idx - trough + 1;
            self.current = None;
            self.current_depth = 0.0;
        }
    }

    /// Rebuilds all episodes from drawdown observations history.
    ///
    /// `drawdowns` is any iterable of `f64` or `&f64`, e.g. the
    /// `&VecDeque<f64>` returned by `HighWaterMarkDrawdown::drawdowns()`.
    pub fn recalculate<I>(&mut self, drawdowns: I)
    where
        I: IntoIterator,
        I::Item: Borrow<f64>,
    {
        self.reset();
        for dd in drawdowns {
            self.update(*dd.borrow());
        }
    }

    /// Drawdown episodes currently known to the tracker.
    ///
    /// If the latest drawdown episode is still open, it is included using
    /// the last processed observation as `to_idx` and with `recovered=false`.
    pub fn episodes(&self) -> Vec<DrawdownEpisode> {
        let mut episodes = self.episodes.clone();
        if let Some((from, trough)) = self.current {
            episodes.push(DrawdownEpisode::new(self.current_depth, from, trough, self.count - 1, false));
        }
        episodes
    }

    /// Drawdown episode depths currently known to the tracker, including
    /// the depth so far of an open episode.
    pub fn depths(&self) -> Vec<f64> {
        let mut depths: Vec<f64> = self.episodes.iter().map(|e| e.depth).collect();
        if self.current.is_some() {
            depths.push(self.current_depth);
        }
        depths
    }

    /// The mean magnitude of the observed discrete episode drawdowns
    /// (0.0 when there are none).
    pub fn average_episode_drawdown(&self) -> f64 {
        let mut sum_depth = self.sum_depth.value();
        let mut count = self.episodes.len();
        if self.current.is_some() {
            sum_depth += self.current_depth;
            count += 1;
        }
        if count > 0 { -sum_depth / count as f64 } else { 0.0 }
    }

    /// The sum of squared episode depths divided by the number of
    /// observations (not the number of episodes), as in
    /// PerformanceAnalytics `DrawdownDeviation` (0.0 when empty).
    pub fn average_episode_drawdown_squared(&self) -> f64 {
        let mut sum_depth_squared = self.sum_depth_squared.value();
        let count = self.count;
        if count == 0 {
            return 0.0;
        }
        if self.current.is_some() {
            sum_depth_squared += self.current_depth * self.current_depth;
        }
        sum_depth_squared / count as f64
    }

    /// The mean length of the observed discrete drawdown episodes
    /// (0.0 when there are none).
    pub fn average_episode_length(&self) -> f64 {
        let mut sum_length = self.sum_length;
        let mut count = self.episodes.len();
        if let Some((from, _)) = self.current {
            sum_length += self.count - from;
            count += 1;
        }
        if count > 0 { sum_length as f64 / count as f64 } else { 0.0 }
    }

    /// The mean peak-to-trough length of the observed discrete drawdown
    /// episodes (0.0 when there are none).
    pub fn average_episode_peak_to_trough(&self) -> f64 {
        let mut sum_peak_to_trough = self.sum_peak_to_trough;
        let mut count = self.episodes.len();
        if let Some((from, trough)) = self.current {
            sum_peak_to_trough += trough - from + 1;
            count += 1;
        }
        if count > 0 { sum_peak_to_trough as f64 / count as f64 } else { 0.0 }
    }

    /// The mean recovery length of the observed discrete drawdown episodes
    /// (0.0 when there are none).
    pub fn average_episode_recovery(&self) -> f64 {
        let mut sum_recovery = self.sum_recovery;
        let mut count = self.episodes.len();
        if let Some((_, trough)) = self.current {
            sum_recovery += self.count - trough;
            count += 1;
        }
        if count > 0 { sum_recovery as f64 / count as f64 } else { 0.0 }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::performance::core::test_support::almost_equal;

    // Drawdown observations (decimals) with two recovered episodes and an
    // open one at the end:
    //
    //   idx:   0     1     2     3     4     5     6     7     8
    //   dd:    0  -0.10 -0.20 -0.05    0  -0.03    0  -0.04 -0.06
    //              |--- episode 1 ---|     |-- 2 --|     |-- 3 (open)
    const DRAWDOWNS: [f64; 9] = [0.0, -0.10, -0.20, -0.05, 0.0, -0.03, 0.0, -0.04, -0.06];

    fn episodes_expected() -> Vec<DrawdownEpisode> {
        vec![
            DrawdownEpisode { depth: -0.20, from_idx: 1, trough_idx: 2, to_idx: 4, recovered: true },
            DrawdownEpisode { depth: -0.03, from_idx: 5, trough_idx: 5, to_idx: 6, recovered: true },
            DrawdownEpisode { depth: -0.06, from_idx: 7, trough_idx: 8, to_idx: 8, recovered: false },
        ]
    }

    fn feed(mut t: DrawdownEpisodes, drawdowns: &[f64]) -> DrawdownEpisodes {
        for &dd in drawdowns {
            t.update(dd);
        }
        t
    }

    fn assert_empty(t: &DrawdownEpisodes) {
        assert_eq!(t.episodes(), vec![]);
        assert_eq!(t.depths(), Vec::<f64>::new());
        assert_eq!(t.average_episode_drawdown(), 0.0);
        assert_eq!(t.average_episode_drawdown_squared(), 0.0);
        assert_eq!(t.average_episode_length(), 0.0);
        assert_eq!(t.average_episode_peak_to_trough(), 0.0);
        assert_eq!(t.average_episode_recovery(), 0.0);
    }

    #[test]
    fn test_empty() {
        assert_empty(&DrawdownEpisodes::new());
    }

    #[test]
    fn test_no_drawdowns() {
        let t = feed(DrawdownEpisodes::new(), &[0.0, 0.0, 0.0]);
        assert_eq!(t.episodes(), vec![]);
        assert_eq!(t.average_episode_drawdown(), 0.0);
        assert_eq!(t.average_episode_drawdown_squared(), 0.0);
    }

    #[test]
    fn test_episodes() {
        let t = feed(DrawdownEpisodes::new(), &DRAWDOWNS);
        assert_eq!(t.episodes(), episodes_expected());
        assert_eq!(t.depths(), vec![-0.20, -0.03, -0.06]);
    }

    #[test]
    fn test_averages() {
        let t = feed(DrawdownEpisodes::new(), &DRAWDOWNS);
        assert!(almost_equal(t.average_episode_drawdown(), (0.20 + 0.03 + 0.06) / 3.0, 16));
        // Divided by the number of observations, not episodes.
        assert!(almost_equal(
            t.average_episode_drawdown_squared(),
            (0.04 + 0.0009 + 0.0036) / DRAWDOWNS.len() as f64,
            16
        ));
        // Lengths 4, 2, 2 (the open episode has no recovery observation).
        assert!(almost_equal(t.average_episode_length(), 8.0 / 3.0, 15));
        // Peak to trough 2, 1, 2.
        assert!(almost_equal(t.average_episode_peak_to_trough(), 5.0 / 3.0, 15));
        // Recovery 3, 2, 1.
        assert!(almost_equal(t.average_episode_recovery(), 2.0, 15));
    }

    #[test]
    fn test_open_episode_progress() {
        let mut t = DrawdownEpisodes::new();
        t.update(-0.05);
        assert_eq!(t.episodes(), vec![DrawdownEpisode::new(-0.05, 0, 0, 0, false)]);
        t.update(-0.02); // shallower, the trough doesn't move
        assert_eq!(t.episodes(), vec![DrawdownEpisode::new(-0.05, 0, 0, 1, false)]);
        t.update(-0.08); // new trough
        assert_eq!(t.episodes(), vec![DrawdownEpisode::new(-0.08, 0, 2, 2, false)]);
        t.update(0.0); // recovery closes the episode
        assert_eq!(t.episodes(), vec![DrawdownEpisode::new(-0.08, 0, 2, 3, true)]);
    }

    #[test]
    fn test_recalculate_matches_incremental() {
        let incremental = feed(DrawdownEpisodes::new(), &DRAWDOWNS);
        let mut rebuilt = feed(DrawdownEpisodes::new(), &[-0.5, -0.7, 0.0]);
        rebuilt.recalculate(DRAWDOWNS);
        assert_eq!(rebuilt.episodes(), incremental.episodes());
        assert_eq!(rebuilt.average_episode_drawdown(), incremental.average_episode_drawdown());
        assert_eq!(
            rebuilt.average_episode_drawdown_squared(),
            incremental.average_episode_drawdown_squared()
        );
        assert_eq!(rebuilt.average_episode_length(), incremental.average_episode_length());
        assert_eq!(rebuilt.average_episode_recovery(), incremental.average_episode_recovery());
    }

    #[test]
    fn test_reset() {
        let mut t = feed(DrawdownEpisodes::new(), &DRAWDOWNS);
        t.reset();
        assert_empty(&t);
        t.update(-0.01);
        assert_eq!(t.episodes(), vec![DrawdownEpisode::new(-0.01, 0, 0, 0, false)]);
    }
}
