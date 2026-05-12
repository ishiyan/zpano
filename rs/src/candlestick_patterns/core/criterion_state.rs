// ---------------------------------------------------------------------------
// CriterionState
// ---------------------------------------------------------------------------
//
// Maintains a running total for a single Criterion over a sliding window.

use super::criterion::Criterion;

/// Maintains a running total for a single Criterion over a sliding window.
///
/// The window covers the `average_period` bars ending at a configurable offset
/// from the current bar. Each pattern decides which offset to use when querying.
pub struct CriterionState {
    pub criterion: Criterion,
    ring: Vec<f64>,
    ring_size: usize,
    ring_start: usize,
    ring_len: usize,
    total: f64,
}

impl CriterionState {
    pub fn new(c: Criterion, max_shift: usize) -> Self {
        // Ring must hold period + max_shift entries so total_at works for all shifts.
        let ring_size = if c.average_period > 0 {
            c.average_period + max_shift
        } else {
            0
        };
        Self {
            criterion: c,
            ring: vec![0.0; ring_size],
            ring_size,
            ring_start: 0,
            ring_len: 0,
            total: 0.0,
        }
    }

    /// Adds the contribution of a new bar and evicts the oldest if the window is full.
    pub fn push(&mut self, o: f64, h: f64, l: f64, c: f64) {
        if self.ring_size == 0 {
            return;
        }
        let val = self.criterion.candle_contribution(o, h, l, c);
        if self.ring_len == self.ring_size {
            // Evict oldest
            self.total -= self.ring[self.ring_start];
            self.ring[self.ring_start] = val;
            self.ring_start = (self.ring_start + 1) % self.ring_size;
        } else {
            let idx = (self.ring_start + self.ring_len) % self.ring_size;
            self.ring[idx] = val;
            self.ring_len += 1;
        }
        self.total += val;
    }

    /// Computes the running total for bars ending at `shift` bars before the current bar.
    ///
    /// For streaming we maintain a single running total for the most recent window.
    /// To support different shifts per pattern, we recompute from the ring buffer.
    /// Since the ring is at most 10 elements, this is still O(period) but with tiny constant.
    pub fn total_at(&self, shift: usize) -> f64 {
        if self.ring_size == 0 || self.criterion.average_period == 0 {
            return 0.0;
        }
        let period = self.criterion.average_period;
        let n = self.ring_len;
        if shift > n || period > n - shift {
            return 0.0;
        }
        let end = n - shift;
        let start = end - period;
        let mut total = 0.0;
        for i in start..end {
            total += self.ring[(self.ring_start + i) % self.ring_size];
        }
        total
    }

    /// Computes the average criterion value.
    ///
    /// shift: How many bars back from the end the window should end.
    /// o, h, l, c: OHLC of the reference candle (used when average_period == 0).
    pub fn avg(&self, shift: usize, o: f64, h: f64, l: f64, c: f64) -> f64 {
        self.criterion.average_value_from_total(self.total_at(shift), o, h, l, c)
    }
}
