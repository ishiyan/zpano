// ---------------------------------------------------------------------------
// CriterionState
// ---------------------------------------------------------------------------
//
// Maintains a running total for a single Criterion over a sliding window.

const Criterion = @import("criterion.zig").Criterion;

pub const max_ring_size: usize = 20;

/// Maintains a running total for a single Criterion over a sliding window.
///
/// The window covers the `average_period` bars ending at a configurable offset
/// from the current bar. Each pattern decides which offset to use when querying.
pub const CriterionState = struct {
    criterion: Criterion,
    ring: [max_ring_size]f64 = [_]f64{0.0} ** max_ring_size,
    ring_size: usize,
    ring_start: usize = 0,
    ring_len: usize = 0,
    total: f64 = 0.0,

    pub fn init(c: Criterion, max_shift: usize) CriterionState {
        // Ring must hold period + max_shift entries so totalAt works for all shifts.
        var ring_size: usize = 0;
        if (c.average_period > 0) {
            ring_size = c.average_period + max_shift;
        }
        return CriterionState{
            .criterion = c,
            .ring_size = ring_size,
        };
    }

    /// Adds the contribution of a new bar and evicts the oldest if the window is full.
    pub fn push(self: *CriterionState, o: f64, h: f64, l: f64, c: f64) void {
        if (self.ring_size == 0) return;
        const val = self.criterion.candleContribution(o, h, l, c);
        if (self.ring_len == self.ring_size) {
            // Evict oldest
            self.total -= self.ring[self.ring_start];
            self.ring[self.ring_start] = val;
            self.ring_start = (self.ring_start + 1) % self.ring_size;
        } else {
            const idx = (self.ring_start + self.ring_len) % self.ring_size;
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
    pub fn totalAt(self: *const CriterionState, shift: usize) f64 {
        if (self.ring_size == 0 or self.criterion.average_period == 0) return 0.0;
        const period = self.criterion.average_period;
        const n = self.ring_len;
        if (shift >= n) return 0.0;
        const end = n - shift;
        if (end < period) return 0.0;
        const start = end - period;
        var total: f64 = 0.0;
        for (start..end) |i| {
            total += self.ring[(self.ring_start + i) % self.ring_size];
        }
        return total;
    }

    /// Computes the average criterion value.
    ///
    /// shift: How many bars back from the end the window should end.
    /// o, h, l, c: OHLC of the reference candle (used when average_period == 0).
    pub fn avg(self: *const CriterionState, shift: usize, o: f64, h: f64, l: f64, c: f64) f64 {
        return self.criterion.averageValueFromTotal(
            self.totalAt(shift),
            o, h, l, c,
        );
    }
};
