/// Weighted signal ensemble aggregator.
///
/// Blends multiple independent signal sources into a single confidence
/// value in [0, 1]. Adaptive methods update weights based on observed
/// outcomes with a configurable feedback delay.
const std = @import("std");
const math = std.math;

const method_mod = @import("method.zig");
const error_metric_mod = @import("error_metric.zig");

pub const AggregationMethod = method_mod.AggregationMethod;
pub const ErrorMetric = error_metric_mod.ErrorMetric;

/// Parameters for constructing an Aggregator.
pub const AggregatorParams = struct {
    /// Number of signal sources (>= 1).
    n_signals: usize,
    /// Aggregation method to use. Defaults to .equal.
    method: AggregationMethod = .equal,
    /// Number of bars between signal observation and outcome availability (>= 1).
    feedback_delay: usize = 1,
    /// Required for .fixed method. Normalized to sum to 1.0.
    weights: ?[]const f64 = null,
    /// Rolling window size for .inverse_variance and .rank_based (>= 2).
    window: usize = 50,
    /// Decay rate for .exponential_decay (0 < alpha <= 1).
    alpha: f64 = 0.1,
    /// Learning rate for .multiplicative_weights (> 0).
    eta: f64 = 0.5,
    /// Prior weights for .bayesian. Defaults to uniform.
    prior: ?[]const f64 = null,
    /// Error metric for .inverse_variance and .rank_based.
    error_metric: ErrorMetric = .absolute,
};

pub const AggregatorError = error{
    InvalidNSignals,
    InvalidFeedbackDelay,
    MissingWeights,
    WeightsLengthMismatch,
    WeightsZeroSum,
    InvalidWindow,
    InvalidAlpha,
    InvalidEta,
    PriorLengthMismatch,
    PriorZeroSum,
    SignalCountMismatch,
};

pub const Aggregator = struct {
    allocator: std.mem.Allocator,
    n: usize,
    method: AggregationMethod,
    feedback_delay: usize,
    count: usize,
    weights: []f64,

    // Ring buffer of signal vectors.
    ring: [][]f64,
    ring_len: usize,
    ring_start: usize,
    ring_capacity: usize,

    // Method-specific state.
    state: MethodState,

    const MethodState = union(enum) {
        none: void,
        inverse_variance: InverseVarianceState,
        exponential_decay: ExponentialDecayState,
        multiplicative_weights: MultiplicativeWeightsState,
        rank_based: RankBasedState,
        bayesian: BayesianState,
    };

    const InverseVarianceState = struct {
        errors: []RollingWindow,
        error_metric: ErrorMetric,
        window: usize,
    };

    const ExponentialDecayState = struct {
        ema: []f64,
        alpha: f64,
    };

    const MultiplicativeWeightsState = struct {
        log_weights: []f64,
        eta: f64,
    };

    const RankBasedState = struct {
        errors: []RollingWindow,
        error_metric: ErrorMetric,
        window: usize,
    };

    const BayesianState = struct {
        log_posterior: []f64,
    };

    /// Fixed-capacity ring buffer for rolling error windows.
    const RollingWindow = struct {
        data: []f64,
        len: usize,
        start: usize,
        capacity: usize,

        fn init(allocator: std.mem.Allocator, capacity: usize) !RollingWindow {
            const data = try allocator.alloc(f64, capacity);
            return .{ .data = data, .len = 0, .start = 0, .capacity = capacity };
        }

        fn deinit(self: *RollingWindow, allocator: std.mem.Allocator) void {
            allocator.free(self.data);
        }

        fn append(self: *RollingWindow, value: f64) void {
            if (self.len < self.capacity) {
                self.data[self.len] = value;
                self.len += 1;
            } else {
                self.data[self.start] = value;
                self.start = (self.start + 1) % self.capacity;
            }
        }

        fn get(self: *const RollingWindow, index: usize) f64 {
            return self.data[(self.start + index) % self.capacity];
        }

        fn sum(self: *const RollingWindow) f64 {
            var s: f64 = 0.0;
            for (0..self.len) |i| {
                s += self.get(i);
            }
            return s;
        }

        fn populationVariance(self: *const RollingWindow) f64 {
            const n: f64 = @floatFromInt(self.len);
            const mean = self.sum() / n;
            var v: f64 = 0.0;
            for (0..self.len) |i| {
                const diff = self.get(i) - mean;
                v += diff * diff;
            }
            return v / n;
        }
    };

    /// Create a new Aggregator.
    pub fn init(allocator: std.mem.Allocator, params: AggregatorParams) (AggregatorError || std.mem.Allocator.Error)!Aggregator {
        if (params.n_signals < 1) return AggregatorError.InvalidNSignals;
        if (params.feedback_delay < 1) return AggregatorError.InvalidFeedbackDelay;

        const n = params.n_signals;
        const ring_capacity = params.feedback_delay + 1;

        // Validate method-specific params before any allocation.
        switch (params.method) {
            .fixed => {
                const w = params.weights orelse return AggregatorError.MissingWeights;
                if (w.len != n) return AggregatorError.WeightsLengthMismatch;
                var s: f64 = 0.0;
                for (w) |v| s += v;
                if (s <= 0.0) return AggregatorError.WeightsZeroSum;
            },
            .inverse_variance, .rank_based => {
                if (params.window < 2) return AggregatorError.InvalidWindow;
            },
            .exponential_decay => {
                if (params.alpha <= 0.0 or params.alpha > 1.0) return AggregatorError.InvalidAlpha;
            },
            .multiplicative_weights => {
                if (params.eta <= 0.0) return AggregatorError.InvalidEta;
            },
            .bayesian => {
                if (params.prior) |prior| {
                    if (prior.len != n) return AggregatorError.PriorLengthMismatch;
                    var s: f64 = 0.0;
                    for (prior) |v| s += v;
                    if (s <= 0.0) return AggregatorError.PriorZeroSum;
                }
            },
            .equal => {},
        }

        // Allocate ring buffer slots.
        const ring = try allocator.alloc([]f64, ring_capacity);
        for (ring) |*slot| {
            slot.* = try allocator.alloc(f64, n);
        }

        // Allocate weights.
        const weights = try allocator.alloc(f64, n);

        var state: MethodState = .{ .none = {} };

        switch (params.method) {
            .fixed => {
                const w = params.weights.?;
                var s: f64 = 0.0;
                for (w) |v| s += v;
                for (0..n) |i| weights[i] = w[i] / s;
            },
            .equal => {
                const inv_n: f64 = 1.0 / @as(f64, @floatFromInt(n));
                for (0..n) |i| weights[i] = inv_n;
            },
            .inverse_variance => {
                const inv_n: f64 = 1.0 / @as(f64, @floatFromInt(n));
                for (0..n) |i| weights[i] = inv_n;
                const errors = try allocator.alloc(RollingWindow, n);
                for (errors) |*e| {
                    e.* = try RollingWindow.init(allocator, params.window);
                }
                state = .{ .inverse_variance = .{
                    .errors = errors,
                    .error_metric = params.error_metric,
                    .window = params.window,
                } };
            },
            .exponential_decay => {
                const inv_n: f64 = 1.0 / @as(f64, @floatFromInt(n));
                for (0..n) |i| weights[i] = inv_n;
                const ema = try allocator.alloc(f64, n);
                for (ema) |*e| e.* = 0.5; // neutral prior
                state = .{ .exponential_decay = .{
                    .ema = ema,
                    .alpha = params.alpha,
                } };
            },
            .multiplicative_weights => {
                const inv_n: f64 = 1.0 / @as(f64, @floatFromInt(n));
                for (0..n) |i| weights[i] = inv_n;
                const log_w = try allocator.alloc(f64, n);
                for (log_w) |*lw| lw.* = 0.0; // uniform in log-space
                state = .{ .multiplicative_weights = .{
                    .log_weights = log_w,
                    .eta = params.eta,
                } };
            },
            .rank_based => {
                const inv_n: f64 = 1.0 / @as(f64, @floatFromInt(n));
                for (0..n) |i| weights[i] = inv_n;
                const errors = try allocator.alloc(RollingWindow, n);
                for (errors) |*e| {
                    e.* = try RollingWindow.init(allocator, params.window);
                }
                state = .{ .rank_based = .{
                    .errors = errors,
                    .error_metric = params.error_metric,
                    .window = params.window,
                } };
            },
            .bayesian => {
                if (params.prior) |prior| {
                    var s: f64 = 0.0;
                    for (prior) |v| s += v;
                    const log_post = try allocator.alloc(f64, n);
                    for (0..n) |i| {
                        weights[i] = prior[i] / s;
                        log_post[i] = @log(prior[i] / s);
                    }
                    state = .{ .bayesian = .{ .log_posterior = log_post } };
                } else {
                    const inv_n: f64 = 1.0 / @as(f64, @floatFromInt(n));
                    for (0..n) |i| weights[i] = inv_n;
                    const log_post = try allocator.alloc(f64, n);
                    for (log_post) |*lp| lp.* = @log(inv_n);
                    state = .{ .bayesian = .{ .log_posterior = log_post } };
                }
            },
        }

        return .{
            .allocator = allocator,
            .n = n,
            .method = params.method,
            .feedback_delay = params.feedback_delay,
            .count = 0,
            .weights = weights,
            .ring = ring,
            .ring_len = 0,
            .ring_start = 0,
            .ring_capacity = ring_capacity,
            .state = state,
        };
    }

    /// Free all allocated memory.
    pub fn deinit(self: *Aggregator) void {
        // Free method-specific state.
        switch (self.state) {
            .none => {},
            .inverse_variance => |*s| {
                for (s.errors) |*e| e.deinit(self.allocator);
                self.allocator.free(s.errors);
            },
            .exponential_decay => |s| {
                self.allocator.free(s.ema);
            },
            .multiplicative_weights => |s| {
                self.allocator.free(s.log_weights);
            },
            .rank_based => |*s| {
                for (s.errors) |*e| e.deinit(self.allocator);
                self.allocator.free(s.errors);
            },
            .bayesian => |s| {
                self.allocator.free(s.log_posterior);
            },
        }

        // Free ring buffer.
        for (self.ring) |slot| {
            self.allocator.free(slot);
        }
        self.allocator.free(self.ring);

        // Free weights.
        self.allocator.free(self.weights);
    }

    /// Blend signal sources into a single confidence value.
    /// `signals` must contain one value per source in [0, 1].
    /// Returns the blended confidence in [0, 1].
    pub fn blend(self: *Aggregator, signals: []const f64) AggregatorError!f64 {
        if (signals.len != self.n) return AggregatorError.SignalCountMismatch;

        var output: f64 = 0.0;
        for (0..self.n) |i| {
            output += self.weights[i] * signals[i];
        }

        // Append to ring buffer.
        const write_pos = (self.ring_start + self.ring_len) % self.ring_capacity;
        if (self.ring_len < self.ring_capacity) {
            @memcpy(self.ring[write_pos], signals);
            self.ring_len += 1;
        } else {
            @memcpy(self.ring[self.ring_start], signals);
            self.ring_start = (self.ring_start + 1) % self.ring_capacity;
        }

        self.count += 1;
        return output;
    }

    /// Provide outcome feedback for weight adaptation.
    /// For stateless methods (.fixed, .equal), this is a no-op.
    /// For adaptive methods, pairs the outcome with the buffered signals
    /// from feedback_delay bars ago and updates weights.
    /// `outcome` is the observed outcome in [0, 1].
    pub fn update(self: *Aggregator, outcome: f64) void {
        if (self.method == .fixed or self.method == .equal) return;
        if (self.ring_len < self.feedback_delay + 1) return;

        // Retrieve signals from feedback_delay bars ago.
        const idx = self.ring_len - 1 - self.feedback_delay;
        const past_signals = self.ring[(self.ring_start + idx) % self.ring_capacity];

        switch (self.state) {
            .inverse_variance => |*s| self.updateInverseVariance(s, past_signals, outcome),
            .exponential_decay => |*s| self.updateExponentialDecay(s, past_signals, outcome),
            .multiplicative_weights => |*s| self.updateMultiplicativeWeights(s, past_signals, outcome),
            .rank_based => |*s| self.updateRankBased(s, past_signals, outcome),
            .bayesian => |*s| self.updateBayesian(s, past_signals, outcome),
            .none => {},
        }
    }

    /// Replay historical data through blend() + update().
    /// Each entry contains signals at bar T and the outcome for bar T.
    /// The method handles the feedback delay internally.
    pub fn warmup(self: *Aggregator, history: []const HistoryEntry) AggregatorError!void {
        var outcomes_len: usize = 0;
        // We reuse a stack buffer for outcomes (up to 4096); beyond that, just skip delayed updates.
        var outcomes_buf: [4096]f64 = undefined;

        for (history) |entry| {
            _ = try self.blend(entry.signals);
            if (outcomes_len < outcomes_buf.len) {
                outcomes_buf[outcomes_len] = entry.outcome;
            }
            outcomes_len += 1;
            if (outcomes_len > self.feedback_delay) {
                const oidx = outcomes_len - 1 - self.feedback_delay;
                if (oidx < outcomes_buf.len) {
                    self.update(outcomes_buf[oidx]);
                }
            }
        }
    }

    /// Get a copy of current weights into the provided buffer.
    pub fn getWeights(self: *const Aggregator, buffer: []f64) void {
        @memcpy(buffer[0..self.n], self.weights);
    }

    // ── Private update methods ──────────────────────────────────────────

    /// Compute per-signal error using the configured metric.
    fn computeError(error_metric: ErrorMetric, signal: f64, outcome: f64) f64 {
        const diff = signal - outcome;
        return switch (error_metric) {
            .absolute => @abs(diff),
            .squared => diff * diff,
        };
    }

    /// Update weights using inverse-variance of prediction errors.
    fn updateInverseVariance(self: *Aggregator, s: *InverseVarianceState, signals: []const f64, outcome: f64) void {
        const epsilon: f64 = 1e-15;

        for (0..self.n) |i| {
            const err = computeError(s.error_metric, signals[i], outcome);
            s.errors[i].append(err);
        }

        // Need at least 2 errors per signal to compute variance.
        for (0..self.n) |i| {
            if (s.errors[i].len < 2) return;
        }

        var total: f64 = 0.0;
        for (0..self.n) |i| {
            const v = s.errors[i].populationVariance();
            const raw = 1.0 / @max(v, epsilon);
            self.weights[i] = raw;
            total += raw;
        }
        for (0..self.n) |i| {
            self.weights[i] /= total;
        }
    }

    /// Update weights using EMA of accuracy.
    fn updateExponentialDecay(self: *Aggregator, s: *ExponentialDecayState, signals: []const f64, outcome: f64) void {
        for (0..self.n) |i| {
            const err = @abs(signals[i] - outcome);
            const accuracy = 1.0 - err;
            s.ema[i] = s.alpha * accuracy + (1.0 - s.alpha) * s.ema[i];
        }

        // Normalize, clamping negative EMAs to 0.
        var total: f64 = 0.0;
        for (0..self.n) |i| {
            const clamped = @max(s.ema[i], 0.0);
            self.weights[i] = clamped;
            total += clamped;
        }
        if (total > 0.0) {
            for (0..self.n) |i| {
                self.weights[i] /= total;
            }
        } else {
            const inv_n: f64 = 1.0 / @as(f64, @floatFromInt(self.n));
            for (0..self.n) |i| {
                self.weights[i] = inv_n;
            }
        }
    }

    /// Update weights using the Hedge algorithm in log-space.
    fn updateMultiplicativeWeights(self: *Aggregator, s: *MultiplicativeWeightsState, signals: []const f64, outcome: f64) void {
        for (0..self.n) |i| {
            const loss = @abs(signals[i] - outcome);
            s.log_weights[i] -= s.eta * loss;
        }

        // Softmax normalization (log-sum-exp trick).
        var max_log: f64 = s.log_weights[0];
        for (1..self.n) |i| {
            if (s.log_weights[i] > max_log) max_log = s.log_weights[i];
        }

        var total: f64 = 0.0;
        for (0..self.n) |i| {
            const exp_w = @exp(s.log_weights[i] - max_log);
            self.weights[i] = exp_w;
            total += exp_w;
        }
        for (0..self.n) |i| {
            self.weights[i] /= total;
        }
    }

    /// Update weights using rank of rolling accuracy.
    fn updateRankBased(self: *Aggregator, s: *RankBasedState, signals: []const f64, outcome: f64) void {
        for (0..self.n) |i| {
            const err = computeError(s.error_metric, signals[i], outcome);
            s.errors[i].append(err);
        }

        // Need at least 1 error per signal.
        for (0..self.n) |i| {
            if (s.errors[i].len < 1) return;
        }

        // Compute mean accuracy per signal, then rank.
        // Use weights buffer temporarily for accuracies, then overwrite with ranks.
        for (0..self.n) |i| {
            const mean_error = s.errors[i].sum() / @as(f64, @floatFromInt(s.errors[i].len));
            self.weights[i] = 1.0 - mean_error;
        }

        // Rank with ties (in-place).
        rankWithTies(self.weights[0..self.n]);

        var total: f64 = 0.0;
        for (0..self.n) |i| total += self.weights[i];
        if (total > 0.0) {
            for (0..self.n) |i| self.weights[i] /= total;
        } else {
            const inv_n: f64 = 1.0 / @as(f64, @floatFromInt(self.n));
            for (0..self.n) |i| self.weights[i] = inv_n;
        }
    }

    /// Update weights using Bayesian model averaging (Bernoulli likelihood).
    fn updateBayesian(self: *Aggregator, s: *BayesianState, signals: []const f64, outcome: f64) void {
        const epsilon: f64 = 1e-15;

        for (0..self.n) |i| {
            // Clamp signal to [epsilon, 1 - epsilon] to avoid log(0).
            const sig = @max(epsilon, @min(1.0 - epsilon, signals[i]));
            const log_lik = outcome * @log(sig) + (1.0 - outcome) * @log(1.0 - sig);
            s.log_posterior[i] += log_lik;
        }

        // Softmax normalization.
        var max_log: f64 = s.log_posterior[0];
        for (1..self.n) |i| {
            if (s.log_posterior[i] > max_log) max_log = s.log_posterior[i];
        }

        var total: f64 = 0.0;
        for (0..self.n) |i| {
            const exp_w = @exp(s.log_posterior[i] - max_log);
            self.weights[i] = exp_w;
            total += exp_w;
        }
        for (0..self.n) |i| {
            self.weights[i] /= total;
        }
    }
};

/// History entry for warmup: signals at bar T and outcome for bar T.
pub const HistoryEntry = struct {
    signals: []const f64,
    outcome: f64,
};

/// Rank values from 1 (worst) to n (best), averaging ties.
/// Operates in-place: values are replaced with their ranks.
fn rankWithTies(values: []f64) void {
    const n = values.len;
    if (n == 0) return;

    // We need sorted indices. Use a simple insertion sort (n is typically small).
    // Stack-allocate index array for small n, heap not needed since max_signals is bounded.
    var indices_buf: [256]usize = undefined;
    const indices = indices_buf[0..n];
    for (0..n) |i| indices[i] = i;

    // Copy original values since we'll overwrite them.
    var values_buf: [256]f64 = undefined;
    const orig = values_buf[0..n];
    @memcpy(orig, values);

    // Insertion sort by value.
    for (1..n) |i_outer| {
        const key_idx = indices[i_outer];
        const key_val = orig[key_idx];
        var j: usize = i_outer;
        while (j > 0 and orig[indices[j - 1]] > key_val) {
            indices[j] = indices[j - 1];
            j -= 1;
        }
        indices[j] = key_idx;
    }

    // Assign ranks with tie averaging.
    var i: usize = 0;
    while (i < n) {
        // Find the end of the tie group.
        var j: usize = i + 1;
        while (j < n and orig[indices[j]] == orig[indices[i]]) {
            j += 1;
        }
        // Average rank for this group (1-based).
        const avg_rank: f64 = (@as(f64, @floatFromInt(i + 1)) + @as(f64, @floatFromInt(j))) / 2.0;
        for (i..j) |k| {
            values[indices[k]] = avg_rank;
        }
        i = j;
    }
}

// ── Tests ───────────────────────────────────────────────────────────────

const testing = std.testing;

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

// ── Validation tests ────────────────────────────────────────────────────

test "validation: n_signals zero" {
    const result = Aggregator.init(testing.allocator, .{ .n_signals = 0 });
    try testing.expectError(AggregatorError.InvalidNSignals, result);
}

test "validation: feedback_delay zero" {
    const result = Aggregator.init(testing.allocator, .{ .n_signals = 2, .feedback_delay = 0 });
    try testing.expectError(AggregatorError.InvalidFeedbackDelay, result);
}

test "validation: fixed requires weights" {
    const result = Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .fixed });
    try testing.expectError(AggregatorError.MissingWeights, result);
}

test "validation: fixed weights wrong length" {
    const w = [_]f64{1.0};
    const result = Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .fixed, .weights = &w });
    try testing.expectError(AggregatorError.WeightsLengthMismatch, result);
}

test "validation: fixed weights zero sum" {
    const w = [_]f64{ 0.0, 0.0 };
    const result = Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .fixed, .weights = &w });
    try testing.expectError(AggregatorError.WeightsZeroSum, result);
}

test "validation: inverse_variance window too small" {
    const result = Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .inverse_variance, .window = 1 });
    try testing.expectError(AggregatorError.InvalidWindow, result);
}

test "validation: exponential_decay alpha zero" {
    const result = Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .exponential_decay, .alpha = 0.0 });
    try testing.expectError(AggregatorError.InvalidAlpha, result);
}

test "validation: multiplicative_weights eta zero" {
    const result = Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .multiplicative_weights, .eta = 0.0 });
    try testing.expectError(AggregatorError.InvalidEta, result);
}

test "validation: bayesian prior wrong length" {
    const p = [_]f64{1.0};
    const result = Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .bayesian, .prior = &p });
    try testing.expectError(AggregatorError.PriorLengthMismatch, result);
}

test "validation: blend wrong signal count" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .equal });
    defer agg.deinit();
    const signals = [_]f64{ 0.5, 0.5 };
    try testing.expectError(AggregatorError.SignalCountMismatch, agg.blend(&signals));
}

// ── Fixed weights tests ─────────────────────────────────────────────────

test "fixed: basic blend" {
    const w = [_]f64{ 0.5, 0.3, 0.2 };
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .fixed, .weights = &w });
    defer agg.deinit();
    const signals = [_]f64{ 1.0, 0.0, 0.0 };
    const result = try agg.blend(&signals);
    try testing.expect(almostEqual(result, 0.5, 1e-13));
}

test "fixed: weights normalized" {
    const w = [_]f64{ 2.0, 8.0 };
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .fixed, .weights = &w });
    defer agg.deinit();
    try testing.expect(almostEqual(agg.weights[0], 0.2, 1e-13));
    try testing.expect(almostEqual(agg.weights[1], 0.8, 1e-13));
}

test "fixed: update is noop" {
    const w = [_]f64{ 0.6, 0.4 };
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .fixed, .weights = &w });
    defer agg.deinit();
    const s1 = [_]f64{ 0.8, 0.2 };
    _ = try agg.blend(&s1);
    const s2 = [_]f64{ 0.7, 0.3 };
    _ = try agg.blend(&s2);
    const w0_before = agg.weights[0];
    agg.update(0.9);
    try testing.expect(almostEqual(agg.weights[0], w0_before, 1e-15));
}

test "fixed: blend all ones" {
    const w = [_]f64{ 0.5, 0.3, 0.2 };
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .fixed, .weights = &w });
    defer agg.deinit();
    const signals = [_]f64{ 1.0, 1.0, 1.0 };
    const result = try agg.blend(&signals);
    try testing.expect(almostEqual(result, 1.0, 1e-13));
}

test "fixed: blend all zeros" {
    const w = [_]f64{ 0.5, 0.3, 0.2 };
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .fixed, .weights = &w });
    defer agg.deinit();
    const signals = [_]f64{ 0.0, 0.0, 0.0 };
    const result = try agg.blend(&signals);
    try testing.expect(almostEqual(result, 0.0, 1e-13));
}

test "fixed: count increments" {
    const w = [_]f64{ 0.5, 0.5 };
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .fixed, .weights = &w });
    defer agg.deinit();
    try testing.expectEqual(@as(usize, 0), agg.count);
    const s = [_]f64{ 0.5, 0.5 };
    _ = try agg.blend(&s);
    try testing.expectEqual(@as(usize, 1), agg.count);
    _ = try agg.blend(&s);
    try testing.expectEqual(@as(usize, 2), agg.count);
}

// ── Equal weights tests ─────────────────────────────────────────────────

test "equal: basic blend" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .equal });
    defer agg.deinit();
    const signals = [_]f64{ 0.9, 0.3, 0.6 };
    const result = try agg.blend(&signals);
    try testing.expect(almostEqual(result, 0.6, 1e-13));
}

test "equal: single signal" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 1, .method = .equal });
    defer agg.deinit();
    const signals = [_]f64{0.7};
    const result = try agg.blend(&signals);
    try testing.expect(almostEqual(result, 0.7, 1e-13));
}

test "equal: weights are uniform" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 4, .method = .equal });
    defer agg.deinit();
    for (0..4) |i| {
        try testing.expect(almostEqual(agg.weights[i], 0.25, 1e-13));
    }
}

test "equal: update is noop" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .equal });
    defer agg.deinit();
    const s1 = [_]f64{ 0.8, 0.2 };
    _ = try agg.blend(&s1);
    const s2 = [_]f64{ 0.7, 0.3 };
    _ = try agg.blend(&s2);
    const w0_before = agg.weights[0];
    agg.update(0.5);
    try testing.expect(almostEqual(agg.weights[0], w0_before, 1e-15));
}

// ── Inverse-variance tests ──────────────────────────────────────────────

test "inverse_variance: initial weights uniform" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .inverse_variance, .window = 10 });
    defer agg.deinit();
    for (0..3) |i| {
        try testing.expect(almostEqual(agg.weights[i], 1.0 / 3.0, 1e-13));
    }
}

test "inverse_variance: accurate signal gets higher weight" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .inverse_variance, .window = 10, .feedback_delay = 1 });
    defer agg.deinit();

    const outcomes = [_]f64{ 0.5, 0.6, 0.4, 0.55, 0.45, 0.5, 0.6, 0.4, 0.55, 0.45 };
    for (0..outcomes.len) |i| {
        const outcome = outcomes[i];
        const s0 = outcome + 0.01 * if (i % 2 == 0) @as(f64, 1.0) else @as(f64, -1.0);
        const s1: f64 = if (i % 2 == 0) 0.9 else 0.1;
        const signals = [_]f64{ s0, s1 };
        _ = try agg.blend(&signals);
        agg.update(outcome);
    }

    try testing.expect(agg.weights[0] > agg.weights[1]);
}

test "inverse_variance: squared error metric" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .inverse_variance, .window = 10, .feedback_delay = 1, .error_metric = .squared });
    defer agg.deinit();

    for (0..5) |_| {
        const signals = [_]f64{ 0.5, 0.5 };
        _ = try agg.blend(&signals);
        agg.update(0.5);
    }
    try testing.expect(almostEqual(agg.weights[0], agg.weights[1], 1e-10));
}

// ── Exponential decay tests ─────────────────────────────────────────────

test "exponential_decay: initial weights uniform" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .exponential_decay });
    defer agg.deinit();
    for (0..3) |i| {
        try testing.expect(almostEqual(agg.weights[i], 1.0 / 3.0, 1e-13));
    }
}

test "exponential_decay: good signal weight increases" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .exponential_decay, .feedback_delay = 1, .alpha = 0.3 });
    defer agg.deinit();

    for (0..20) |_| {
        const signals = [_]f64{ 0.8, 0.2 };
        _ = try agg.blend(&signals);
        agg.update(0.8);
    }

    try testing.expect(agg.weights[0] > agg.weights[1]);
}

test "exponential_decay: alpha one" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .exponential_decay, .feedback_delay = 1, .alpha = 1.0 });
    defer agg.deinit();

    const s = [_]f64{ 0.9, 0.1 };
    _ = try agg.blend(&s);
    _ = try agg.blend(&s);
    agg.update(0.9);
    // Signal 0 accuracy = 1.0, signal 1 accuracy = 0.2
    try testing.expect(almostEqual(agg.weights[0], 1.0 / 1.2, 1e-13));
    try testing.expect(almostEqual(agg.weights[1], 0.2 / 1.2, 1e-13));
}

// ── Multiplicative weights tests ────────────────────────────────────────

test "multiplicative_weights: initial weights uniform" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .multiplicative_weights });
    defer agg.deinit();
    for (0..3) |i| {
        try testing.expect(almostEqual(agg.weights[i], 1.0 / 3.0, 1e-13));
    }
}

test "multiplicative_weights: best signal converges" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .multiplicative_weights, .feedback_delay = 1, .eta = 0.5 });
    defer agg.deinit();

    for (0..50) |_| {
        const signals = [_]f64{ 0.8, 0.2, 0.3 };
        _ = try agg.blend(&signals);
        agg.update(0.8);
    }

    try testing.expect(agg.weights[0] > 0.5);
    try testing.expect(agg.weights[0] > agg.weights[1]);
    try testing.expect(agg.weights[0] > agg.weights[2]);
}

test "multiplicative_weights: high eta faster convergence" {
    var results: [2]f64 = undefined;
    const etas = [_]f64{ 0.1, 1.0 };
    for (0..2) |e| {
        var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .multiplicative_weights, .feedback_delay = 1, .eta = etas[e] });
        defer agg.deinit();
        for (0..10) |_| {
            const signals = [_]f64{ 0.9, 0.1 };
            _ = try agg.blend(&signals);
            agg.update(0.9);
        }
        results[e] = agg.weights[0];
    }
    try testing.expect(results[1] > results[0]);
}

// ── Rank-based tests ────────────────────────────────────────────────────

test "rank_based: initial weights uniform" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .rank_based, .window = 10 });
    defer agg.deinit();
    for (0..3) |i| {
        try testing.expect(almostEqual(agg.weights[i], 1.0 / 3.0, 1e-13));
    }
}

test "rank_based: rank ordering" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .rank_based, .feedback_delay = 1, .window = 10 });
    defer agg.deinit();

    for (0..15) |_| {
        const signals = [_]f64{ 0.7, 0.5, 0.2 };
        _ = try agg.blend(&signals);
        agg.update(0.7);
    }

    try testing.expect(agg.weights[0] > agg.weights[1]);
    try testing.expect(agg.weights[1] > agg.weights[2]);
}

test "rank_based: ties get average rank" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .rank_based, .feedback_delay = 1, .window = 10 });
    defer agg.deinit();

    for (0..5) |_| {
        const signals = [_]f64{ 0.5, 0.5 };
        _ = try agg.blend(&signals);
        agg.update(0.5);
    }

    try testing.expect(almostEqual(agg.weights[0], agg.weights[1], 1e-13));
}

// ── Bayesian tests ──────────────────────────────────────────────────────

test "bayesian: uniform prior" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .bayesian });
    defer agg.deinit();
    for (0..3) |i| {
        try testing.expect(almostEqual(agg.weights[i], 1.0 / 3.0, 1e-13));
    }
}

test "bayesian: custom prior" {
    const p = [_]f64{ 0.5, 0.3, 0.2 };
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .bayesian, .prior = &p });
    defer agg.deinit();
    try testing.expect(almostEqual(agg.weights[0], 0.5, 1e-13));
    try testing.expect(almostEqual(agg.weights[1], 0.3, 1e-13));
    try testing.expect(almostEqual(agg.weights[2], 0.2, 1e-13));
}

test "bayesian: good predictor dominates" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .bayesian, .feedback_delay = 1 });
    defer agg.deinit();

    for (0..20) |_| {
        const signals = [_]f64{ 0.9, 0.1 };
        _ = try agg.blend(&signals);
        agg.update(0.9);
    }

    try testing.expect(agg.weights[0] > 0.9);
}

test "bayesian: evidence overrides prior" {
    const p = [_]f64{ 0.1, 0.9 };
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .bayesian, .feedback_delay = 1, .prior = &p });
    defer agg.deinit();

    for (0..50) |_| {
        const signals = [_]f64{ 0.8, 0.2 };
        _ = try agg.blend(&signals);
        agg.update(0.8);
    }

    try testing.expect(agg.weights[0] > agg.weights[1]);
}

// ── Delayed feedback tests ──────────────────────────────────────────────

test "delayed_feedback: delay 1" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .exponential_decay, .feedback_delay = 1, .alpha = 1.0 });
    defer agg.deinit();

    const s1 = [_]f64{ 0.9, 0.1 };
    _ = try agg.blend(&s1);
    const s2 = [_]f64{ 0.5, 0.5 };
    _ = try agg.blend(&s2);
    agg.update(0.9);

    try testing.expect(almostEqual(agg.weights[0], 1.0 / 1.2, 1e-13));
    try testing.expect(almostEqual(agg.weights[1], 0.2 / 1.2, 1e-13));
}

test "delayed_feedback: delay 2" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .exponential_decay, .feedback_delay = 2, .alpha = 1.0 });
    defer agg.deinit();

    const s1 = [_]f64{ 0.9, 0.1 };
    _ = try agg.blend(&s1);
    const s2 = [_]f64{ 0.5, 0.5 };
    _ = try agg.blend(&s2);
    // Not enough history (need 3 entries).
    agg.update(0.9);
    for (0..2) |i| {
        try testing.expect(almostEqual(agg.weights[i], 0.5, 1e-13));
    }

    const s3 = [_]f64{ 0.3, 0.7 };
    _ = try agg.blend(&s3);
    agg.update(0.9); // pairs with s1

    try testing.expect(almostEqual(agg.weights[0], 1.0 / 1.2, 1e-13));
    try testing.expect(almostEqual(agg.weights[1], 0.2 / 1.2, 1e-13));
}

test "delayed_feedback: update without enough history" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .exponential_decay, .feedback_delay = 3, .alpha = 0.5 });
    defer agg.deinit();

    const s = [_]f64{ 0.5, 0.5 };
    _ = try agg.blend(&s);
    const w0_before = agg.weights[0];
    agg.update(0.5);
    try testing.expect(almostEqual(agg.weights[0], w0_before, 1e-15));
}

// ── Warmup tests ────────────────────────────────────────────────────────

test "warmup: equals live replay" {
    const s1 = [_]f64{ 0.8, 0.3 };
    const s2 = [_]f64{ 0.6, 0.5 };
    const s3 = [_]f64{ 0.9, 0.2 };
    const s4 = [_]f64{ 0.7, 0.4 };
    const s5 = [_]f64{ 0.5, 0.6 };
    const history = [_]HistoryEntry{
        .{ .signals = &s1, .outcome = 0.7 },
        .{ .signals = &s2, .outcome = 0.5 },
        .{ .signals = &s3, .outcome = 0.8 },
        .{ .signals = &s4, .outcome = 0.6 },
        .{ .signals = &s5, .outcome = 0.4 },
    };

    // Method 1: warmup.
    var agg1 = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .exponential_decay, .feedback_delay = 1, .alpha = 0.2 });
    defer agg1.deinit();
    try agg1.warmup(&history);

    // Method 2: manual replay.
    var agg2 = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .exponential_decay, .feedback_delay = 1, .alpha = 0.2 });
    defer agg2.deinit();
    const outcomes = [_]f64{ 0.7, 0.5, 0.8, 0.6, 0.4 };
    for (0..history.len) |i| {
        _ = try agg2.blend(history[i].signals);
        if (i >= 1) {
            agg2.update(outcomes[i - 1]);
        }
    }

    try testing.expectEqual(agg1.count, agg2.count);
    for (0..2) |i| {
        try testing.expect(almostEqual(agg1.weights[i], agg2.weights[i], 1e-13));
    }
}

test "warmup: bayesian calibrated" {
    const s = [_]f64{ 0.9, 0.1 };
    var entries: [20]HistoryEntry = undefined;
    for (0..20) |i| {
        entries[i] = .{ .signals = &s, .outcome = 0.9 };
    }

    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .bayesian, .feedback_delay = 1 });
    defer agg.deinit();
    try agg.warmup(&entries);

    try testing.expect(agg.weights[0] > 0.9);
}

// ── Edge case tests ─────────────────────────────────────────────────────

test "edge: single signal all methods" {
    const methods = [_]AggregationMethod{ .fixed, .equal, .inverse_variance, .exponential_decay, .multiplicative_weights, .rank_based, .bayesian };
    for (methods) |m| {
        const w = [_]f64{1.0};
        var agg = try Aggregator.init(testing.allocator, .{
            .n_signals = 1,
            .method = m,
            .weights = if (m == .fixed) &w else null,
        });
        defer agg.deinit();
        const signals = [_]f64{0.73};
        const result = try agg.blend(&signals);
        try testing.expect(almostEqual(result, 0.73, 1e-13));
    }
}

test "edge: extreme signals" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .equal });
    defer agg.deinit();
    const signals = [_]f64{ 0.0, 1.0 };
    const result = try agg.blend(&signals);
    try testing.expect(almostEqual(result, 0.5, 1e-13));
}

test "edge: bayesian extreme signals" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 2, .method = .bayesian, .feedback_delay = 1 });
    defer agg.deinit();
    const signals = [_]f64{ 0.0, 1.0 };
    _ = try agg.blend(&signals);
    _ = try agg.blend(&signals);
    agg.update(1.0);
    var w_sum: f64 = 0.0;
    for (0..2) |i| w_sum += agg.weights[i];
    try testing.expect(almostEqual(w_sum, 1.0, 1e-13));
}

test "edge: inverse_variance identical signals" {
    var agg = try Aggregator.init(testing.allocator, .{ .n_signals = 3, .method = .inverse_variance, .feedback_delay = 1, .window = 10 });
    defer agg.deinit();

    for (0..5) |_| {
        const signals = [_]f64{ 0.5, 0.5, 0.5 };
        _ = try agg.blend(&signals);
        agg.update(0.5);
    }

    for (0..3) |i| {
        try testing.expect(almostEqual(agg.weights[i], 1.0 / 3.0, 1e-10));
    }
}

test "edge: weights sum to one all methods" {
    const methods = [_]AggregationMethod{ .fixed, .equal, .inverse_variance, .exponential_decay, .multiplicative_weights, .rank_based, .bayesian };
    for (methods) |m| {
        const w = [_]f64{ 0.3, 0.7 };
        var agg = try Aggregator.init(testing.allocator, .{
            .n_signals = 2,
            .method = m,
            .weights = if (m == .fixed) &w else null,
        });
        defer agg.deinit();
        var w_sum: f64 = 0.0;
        for (0..2) |i| w_sum += agg.weights[i];
        try testing.expect(almostEqual(w_sum, 1.0, 1e-13));
    }
}
