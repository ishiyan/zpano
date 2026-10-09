//! Streaming calculation of time-series performance and risk measures
//! (port of Python `performance/measures.py`).
//!
//! `Measures` operates on return observations that have a common,
//! explicitly defined observation period. It does not require timestamps
//! and does not perform time-based resampling.
//!
//! `periods_per_annum` defines the annualization convention used by
//! annualized measures: the number of return observations assumed to
//! represent one year (daily equity 252, weekly 52, monthly 12,
//! quarterly 4, annual 1, one-minute US equity session 98_280).
//!
//! Conventions of this port:
//!
//! - Python properties and methods become camelCase methods; Python
//!   default arguments are documented in `///` comments and must be passed
//!   explicitly.
//! - NaN is returned exactly where Python returns `math.nan`, 0.0 where
//!   Python returns 0, and ±inf where Python returns ±inf.
//! - List-valued measures return a new `[]f64` allocated with the
//!   `Measures` allocator; the caller owns it and frees it with the same
//!   allocator.
//! - `error.InvalidArgument` is returned exactly where Python raises
//!   `ValueError` directly (`init`, `isNormalDistribution`,
//!   `farinelliTibilettiRatio`, `rachevRatio`, `cdarAverage`,
//!   `cdarDiscrete`, `cdarBeta`, `cdarAlpha`, `tailRatio`, `biasRatio`,
//!   `rewardToConditionalDrawdown`).
//! - VaR/ES based measures (`var*`, `es*`, `rewardTo*Ratio*`,
//!   `sharpeRatioVar*`, `sharpeRatioEs*`), whose Python versions raise only
//!   indirectly through core helpers, return NaN exactly where Python would
//!   raise (historical: 1 - confidence ∉ [0, 1] or NaN; Gaussian and
//!   Cornish-Fisher: 1 - confidence ∉ (0, 1)) instead of an error.
//! - Measures that need scratch memory (sorting, percentiles, episode
//!   lists) use the stored allocator and may return `error.OutOfMemory`.

const std = @import("std");
const math = std.math;
const Allocator = std.mem.Allocator;

const RawMomentsKleinKBN = @import("raw_moments_klein_kbn").RawMomentsKleinKBN;
const KleinKBNAccumulator = @import("klein_kbn_accumulator").KleinKBNAccumulator;

const core = @import("core/core.zig");
const FifoBuffer = @import("core/fifo_buffer.zig").FifoBuffer;

const sqrt2: f64 = 1.4142135623730950488016887242097;
const nan = math.nan(f64);

pub const periods_per_annum_year: f64 = 1;
pub const periods_per_annum_quarter: f64 = 4;
pub const periods_per_annum_month: f64 = 12;
pub const periods_per_annum_week: f64 = 52;
pub const periods_per_annum_day: f64 = 252;
/// 390 regular-session minutes/day by 252 trading days/year.
pub const periods_per_annum_minute_us_equities: f64 = 98280;
/// 1440 minutes/day by 365 days/year.
pub const periods_per_annum_minute_crypto: f64 = 525600;

/// Errors of the measures that raise `ValueError` in Python and may also
/// allocate scratch memory.
pub const Error = error{InvalidArgument} || Allocator.Error;

/// Errors of `Measures.addReturn`. `error.EmptyRevert` cannot occur when the
/// API is used as intended (it would indicate corrupted internal state).
pub const AddReturnError = Allocator.Error || error{EmptyRevert};

/// The Jarque–Bera normality decision rule used by
/// `Measures.isNormalDistribution`: true when `jb <= -2·ln(1 - confidence)`
/// (the χ²(2) critical value).
///
/// Returns false when `jb` is NaN and confidence is valid.
/// Returns `error.InvalidArgument` unless confidence is in (0, 1), including NaN.
pub fn isNormalFromJb(jb: f64, confidence: f64) error{InvalidArgument}!bool {
    if (!(0 < confidence and confidence < 1)) return error.InvalidArgument;
    if (math.isNan(jb)) return false;
    // For 2 degrees of freedom chi2.ppf(p, 2) = -2 ln(1 - p).
    const critical = -2.0 * math.log1p(-confidence);
    return jb <= critical;
}

/// Python's built-in `sum()` over floats (CPython ≥ 3.12: Neumaier
/// compensated summation).
fn pySum(values: []const f64) f64 {
    var f_result: f64 = 0.0;
    var c: f64 = 0.0;
    for (values) |x| {
        const t = f_result + x;
        if (@abs(f_result) >= @abs(x)) {
            c += (f_result - t) + x;
        } else {
            c += (x - t) + f_result;
        }
        f_result = t;
    }
    if (c != 0 and math.isFinite(c)) f_result += c;
    return f_result;
}

inline fn fl(x: usize) f64 {
    return @floatFromInt(x);
}

/// Maps a core historical VaR/ES result: InvalidArgument → NaN.
fn historicalOrNan(r: core.var_mod.HistoricalError!f64) Allocator.Error!f64 {
    return r catch |e| switch (e) {
        error.OutOfMemory => error.OutOfMemory,
        error.InvalidArgument => nan,
    };
}

/// Maps a core parametric VaR/ES result: InvalidArgument → NaN.
fn parametricOrNan(r: core.var_mod.ParametricError!f64) f64 {
    return r catch nan;
}

/// Percentile with an argument that is known to be valid; only
/// `error.OutOfMemory` can surface.
fn percentileValid(allocator: Allocator, window: []const f64, q: f64) Allocator.Error!f64 {
    return core.percentile(allocator, window, q) catch |e| switch (e) {
        error.OutOfMemory => error.OutOfMemory,
        error.InvalidArgument => unreachable,
    };
}

fn newMoments() RawMomentsKleinKBN {
    // ddof=1, bias=True, fisher=True matches scipy's default behavior for kurtosis.
    return .{ .ddof = 1, .bias = true, .fisher = true };
}

/// Streaming calculation of time-series performance and risk measures.
///
/// Create with `init`, free with `deinit`. Feed one (portfolio,
/// benchmark) return pair per period with `addReturn`.
///
/// Public configuration fields: `periods_per_annum`, `risk_free_rate`
/// (periodic), `target_return` (periodic). All other fields are internal
/// state and must not be modified.
pub const Measures = struct {
    allocator: Allocator,

    /// Number of return periods per annum used for annualization.
    periods_per_annum: f64,
    /// Periodic risk-free rate implied by the annual rate and
    /// `periods_per_annum`.
    risk_free_rate: f64,
    /// Periodic target return (MAR) implied by the annual target return
    /// and `periods_per_annum`.
    target_return: f64,

    // ── internal state ───────────────────────────────────────────────────
    sqrt_periods_per_annum: f64,
    annual_risk_free_rate: f64,
    rolling_window_size: usize,

    returns: FifoBuffer(f64) = .{},
    returns_benchmark: FifoBuffer(f64) = .{},

    win_loss: core.WinLoss = .{},
    capture: core.Capture = .{},

    returns_kbn: RawMomentsKleinKBN,
    excess_returns_kbn: RawMomentsKleinKBN,
    benchmark_returns_kbn: RawMomentsKleinKBN,
    benchmark_excess_returns_kbn: RawMomentsKleinKBN,

    sfm_regression: core.SFMRegression,
    active_returns_kbn: RawMomentsKleinKBN,

    target_returns_kbn: RawMomentsKleinKBN,
    target_partial_moments: core.PartialMoments,
    raw_partial_moments: core.RawPartialMoments = .{},
    benchmark_target_partial_moments: core.PartialMoments,

    cumulative_return: core.CumulativeReturn = .{},
    cumulative_excess_return: core.CumulativeReturn = .{},
    benchmark_cumulative_return: core.CumulativeReturn = .{},

    drawdown_continuous_runs: core.ContinuousDrawdownRuns,
    drawdown_high_watermark: core.HighWaterMarkDrawdown,
    drawdown_high_watermark_benchmark: core.HighWaterMarkDrawdown,
    drawdown_episodes: core.DrawdownEpisodes,
    drawdown_episodes_benchmark: core.DrawdownEpisodes,

    const Self = @This();

    /// Creates an empty instance.
    ///
    /// Args (Python defaults):
    ///     periods_per_annum: number of return periods per annum (252.0);
    ///         must be positive, else `error.InvalidArgument`. Also converts
    ///         the annual rates to periodic ones: (1 + r)^(1/P) - 1 (used
    ///         as-is when the annual rate is 0 or P == 1).
    ///     annual_risk_free_rate: annual risk-free rate as a decimal (0.0).
    ///     annual_target_return: annual target return / MAR (0.0).
    ///     rolling_window_size: number of most recent observations used by
    ///         rolling measures; 0 means an unbounded window (0).
    pub fn init(
        allocator: Allocator,
        periods_per_annum: f64,
        annual_risk_free_rate: f64,
        annual_target_return: f64,
        rolling_window_size: usize,
    ) error{InvalidArgument}!Self {
        if (periods_per_annum <= 0) return error.InvalidArgument;

        const rf = if (annual_risk_free_rate == 0 or periods_per_annum == 1)
            annual_risk_free_rate
        else
            math.pow(f64, 1 + annual_risk_free_rate, 1 / periods_per_annum) - 1;

        const target = if (annual_target_return == 0 or periods_per_annum == 1)
            annual_target_return
        else
            math.pow(f64, 1 + annual_target_return, 1 / periods_per_annum) - 1;

        const hwm_window: i64 = math.cast(i64, rolling_window_size) orelse math.maxInt(i64);

        return .{
            .allocator = allocator,
            .periods_per_annum = periods_per_annum,
            .risk_free_rate = rf,
            .target_return = target,
            .sqrt_periods_per_annum = @sqrt(periods_per_annum),
            .annual_risk_free_rate = annual_risk_free_rate,
            .rolling_window_size = rolling_window_size,
            .returns_kbn = newMoments(),
            .excess_returns_kbn = newMoments(),
            .benchmark_returns_kbn = newMoments(),
            .benchmark_excess_returns_kbn = newMoments(),
            .sfm_regression = core.SFMRegression.init(rf),
            .active_returns_kbn = newMoments(),
            .target_returns_kbn = newMoments(),
            .target_partial_moments = core.PartialMoments.init(target),
            .benchmark_target_partial_moments = core.PartialMoments.init(target),
            .drawdown_continuous_runs = core.ContinuousDrawdownRuns.init(allocator),
            .drawdown_high_watermark = core.HighWaterMarkDrawdown.init(allocator, hwm_window),
            .drawdown_high_watermark_benchmark = core.HighWaterMarkDrawdown.init(allocator, hwm_window),
            .drawdown_episodes = core.DrawdownEpisodes.init(allocator),
            .drawdown_episodes_benchmark = core.DrawdownEpisodes.init(allocator),
        };
    }

    /// Frees all memory owned by the instance.
    pub fn deinit(self: *Self) void {
        self.returns.deinit(self.allocator);
        self.returns_benchmark.deinit(self.allocator);
        self.drawdown_continuous_runs.deinit();
        self.drawdown_high_watermark.deinit();
        self.drawdown_high_watermark_benchmark.deinit();
        self.drawdown_episodes.deinit();
        self.drawdown_episodes_benchmark.deinit();
    }

    /// Reset all accumulated return data and derived streaming state.
    ///
    /// Configuration (periods per annum, rates, rolling window size) is
    /// preserved.
    pub fn reset(self: *Self) void {
        self.returns.clear();
        self.returns_benchmark.clear();

        self.win_loss.reset();
        self.capture.reset();

        self.returns_kbn.reset();
        self.excess_returns_kbn.reset();
        self.benchmark_returns_kbn.reset();
        self.benchmark_excess_returns_kbn.reset();

        self.sfm_regression.reset();
        self.active_returns_kbn.reset();

        self.target_returns_kbn.reset();
        self.target_partial_moments.reset();
        self.raw_partial_moments.reset();
        self.benchmark_target_partial_moments.reset();

        self.cumulative_return.reset();
        self.cumulative_excess_return.reset();
        self.benchmark_cumulative_return.reset();

        self.drawdown_continuous_runs.reset();
        self.drawdown_high_watermark.reset();
        self.drawdown_high_watermark_benchmark.reset();
        self.drawdown_episodes.reset();
        self.drawdown_episodes_benchmark.reset();
    }

    /// Add one periodic portfolio and benchmark return observation.
    ///
    /// The returns must already represent the same observation period and
    /// use the periodicity implied by `periods_per_annum`. Inputs are not
    /// validated (a return <= -1 yields NaN/inf from log1p where Python
    /// raises).
    ///
    /// May return `error.OutOfMemory`; the instance should then be `reset`
    /// before further use.
    pub fn addReturn(self: *Self, ret: f64, ret_bench: f64) AddReturnError!void {
        const evicted = self.rolling_window_size > 0 and self.returns.len() == self.rolling_window_size;
        try self.returns.ensureUnusedCapacity(self.allocator, 1);
        try self.returns_benchmark.ensureUnusedCapacity(self.allocator, 1);
        if (evicted) {
            const ret_old = self.returns.popFront();
            const ret_bench_old = self.returns_benchmark.popFront();
            try self.returns_kbn.revert(ret_old);
            // Excess returns (returns less risk-free rate)
            try self.excess_returns_kbn.revert(ret_old - self.risk_free_rate);
            // Target returns (returns less target return)
            try self.target_returns_kbn.revert(ret_old - self.target_return);
            try self.target_partial_moments.revert(ret_old);
            try self.raw_partial_moments.revert(ret_old);
            try self.benchmark_target_partial_moments.revert(ret_bench_old);
            try self.win_loss.revert(ret_old);
            try self.capture.revert(ret_old, ret_bench_old);
            try self.cumulative_return.revert(ret_old);
            try self.cumulative_excess_return.revert(ret_old - self.risk_free_rate);
            try self.benchmark_cumulative_return.revert(ret_bench_old);
            // Benchmarks
            try self.benchmark_returns_kbn.revert(ret_bench_old);
            try self.benchmark_excess_returns_kbn.revert(ret_bench_old - self.risk_free_rate);
            try self.active_returns_kbn.revert(ret_old - ret_bench_old);
            try self.sfm_regression.revert(ret_old, ret_bench_old);
            // Drawdowns: high watermark drawdown and drawdown episodes have
            // no revert(), they evict the oldest observation themselves.
            try self.drawdown_continuous_runs.revert(ret_old); // Burke
        }

        self.returns_kbn.update(ret);
        // Excess returns (returns less risk-free rate)
        const ret_excess = ret - self.risk_free_rate;
        self.excess_returns_kbn.update(ret_excess);
        // Target returns (returns less target return)
        self.target_returns_kbn.update(ret - self.target_return);
        self.target_partial_moments.update(ret);
        self.raw_partial_moments.update(ret);
        self.benchmark_target_partial_moments.update(ret_bench);
        self.win_loss.update(ret);
        self.capture.update(ret, ret_bench);
        // Benchmarks
        self.benchmark_returns_kbn.update(ret_bench);
        const ret_bench_excess = ret_bench - self.risk_free_rate;
        self.benchmark_excess_returns_kbn.update(ret_bench_excess);
        self.active_returns_kbn.update(ret - ret_bench);
        self.sfm_regression.update(ret, ret_bench);

        try self.returns.append(self.allocator, ret);
        try self.returns_benchmark.append(self.allocator, ret_bench);

        // Cumulative return
        self.cumulative_return.update(ret);
        self.cumulative_excess_return.update(ret_excess);
        self.benchmark_cumulative_return.update(ret_bench);

        // Drawdown calculation used in Burke
        try self.drawdown_continuous_runs.update(ret);

        // High-water-mark drawdown and drawdown episodes. Drawdown episodes
        // have no revert(): when an observation leaves the rolling window
        // (or the window's drawdowns were recomputed), the episodes are
        // rebuilt from the window's drawdowns, so that episode indices
        // refer to positions in the window.
        var hwm = &self.drawdown_high_watermark;
        var recalculated = try hwm.update(ret);
        if (recalculated or evicted) {
            try self.drawdown_episodes.recalculate(hwm.drawdowns());
        } else {
            try self.drawdown_episodes.update(hwm.drawdown());
        }
        hwm = &self.drawdown_high_watermark_benchmark;
        recalculated = try hwm.update(ret_bench);
        if (recalculated or evicted) {
            try self.drawdown_episodes_benchmark.recalculate(hwm.drawdowns());
        } else {
            try self.drawdown_episodes_benchmark.update(hwm.drawdown());
        }
    }

    // ── Autocorrelation ──────────────────────────────────────────────────

    /// The autocorrelation penalty factor for serially correlated returns
    /// (Lo, "The Statistics of Sharpe Ratios", 2002):
    ///
    ///     sqrt(1 + 2 * sum_{k=1}^{q-1} (1 - k/q) * rho_k)
    ///
    /// where rho_k is the sample autocorrelation at lag k and
    /// q = int(min(periods_per_annum, n - 1)). Divide a Sharpe or Sortino
    /// ratio by it to obtain an autocorrelation-adjusted ratio.
    ///
    /// 1.0 when there are fewer than two observations or the return
    /// variance is zero. O(n·q).
    pub fn autocorrelationPenalty(self: *const Self) f64 {
        const n = self.returns_kbn.n();
        if (n < 2) return 1.0;
        const mean = self.returns_kbn.mean();
        const denom = self.returns_kbn.varianceDdof0() * fl(n);
        if (math.isNan(denom) or denom == 0) return 1.0;

        // Lo's recommended aggregation period: Python int(min(P, n - 1)).
        const nm1 = fl(n - 1);
        const lim = if (nm1 < self.periods_per_annum) nm1 else self.periods_per_annum;
        if (math.isNan(lim)) return nan;
        const q: usize = @intFromFloat(lim);

        const w = self.returns.slice();
        var s: f64 = 0.0;
        var k: usize = 1;
        while (k < q) : (k += 1) {
            var numer: f64 = 0.0;
            var t: usize = k;
            while (t < n) : (t += 1) {
                numer += (w[t] - mean) * (w[t - k] - mean);
            }
            const rho = numer / denom;
            s += (1.0 - fl(k) / fl(q)) * rho;
        }

        const v = 1.0 + 2.0 * s;
        return @sqrt(if (v > 0.0) v else 0.0);
    }

    // ── Return and growth ────────────────────────────────────────────────

    /// Cumulative geometric return, prod(1 + r) - 1 (0.0 when empty).
    pub fn cumulativeGeometricReturn(self: *const Self) f64 {
        return self.cumulative_return.cumulativeGeometricReturn();
    }

    /// The geometric mean return per observation, prod(1 + r)^(1/n) - 1.
    pub fn geometricMeanReturn(self: *const Self) f64 {
        return self.cumulative_return.geometricMeanReturn();
    }

    /// Compound Annual Growth Rate (CAGR): the geometric mean return
    /// annualized with `periods_per_annum`.
    pub fn compoundAnnualGrowthRate(self: *const Self) f64 {
        return self.cumulative_return.annualizedGeometricMeanReturn(self.periods_per_annum);
    }

    // ── Moments and normality ────────────────────────────────────────────

    /// The "scipy default" skewness, calculated as 'moment' (bias=True)
    /// skewness.
    pub fn skewness(self: *const Self) f64 {
        return self.returns_kbn.skewness();
    }

    /// The 'moment' skewness g1 = mu3 / mu2^(3/2).
    pub fn skewnessMoment(self: *const Self) f64 {
        return self.returns_kbn.skewnessMoment();
    }

    /// The 'fisher' skewness g1 * sqrt(n(n-1)) / (n-2).
    pub fn skewnessFisher(self: *const Self) f64 {
        return self.returns_kbn.skewnessFisher();
    }

    /// The 'sample' skewness g1 * n² / ((n-1)(n-2)).
    pub fn skewnessSample(self: *const Self) f64 {
        return self.returns_kbn.skewnessSample();
    }

    /// The "scipy default" kurtosis: biased excess kurtosis.
    pub fn kurtosis(self: *const Self) f64 {
        return self.returns_kbn.kurtosis();
    }

    /// The biased excess kurtosis beta2 - 3, beta2 = mu4 / mu2².
    pub fn kurtosisExcess(self: *const Self) f64 {
        return self.returns_kbn.kurtosisExcess();
    }

    /// The biased Pearson (population) kurtosis beta2 = mu4 / mu2².
    pub fn kurtosisMoment(self: *const Self) f64 {
        return self.returns_kbn.kurtosisMoment();
    }

    /// The unbiased excess kurtosis
    /// ((n²-1)·beta2 - 3(n-1)²) / ((n-2)(n-3)).
    pub fn kurtosisSampleExcess(self: *const Self) f64 {
        return self.returns_kbn.kurtosisSampleExcess();
    }

    /// The PerformanceAnalytics 'sample' kurtosis
    /// (n²-1)·beta2 / ((n-2)(n-3)).
    pub fn kurtosisSampleCorrected(self: *const Self) f64 {
        return self.returns_kbn.kurtosisSampleCorrected();
    }

    /// The unbiased Pearson kurtosis
    /// ((n²-1)·beta2 - 3(n-1)²) / ((n-2)(n-3)) + 3.
    pub fn kurtosisSample(self: *const Self) f64 {
        return self.returns_kbn.kurtosisSample();
    }

    /// Ratio of the population ('moment') skewness g1 to the population
    /// (Pearson, non-excess) kurtosis beta2. Higher is better.
    pub fn skewnessKurtosisRatio(self: *const Self) f64 {
        const s = self.returns_kbn.skewnessMoment();
        const k = self.returns_kbn.kurtosisMoment();
        return if (k != 0) s / k else nan;
    }

    /// Jarque–Bera normality test statistic
    /// JB = n/6 * (skewness² + excess_kurtosis²/4), using the population
    /// skewness and excess kurtosis. Asymptotically χ²(2) under normality.
    pub fn jarqueBeraNormalityTestStatistic(self: *const Self) f64 {
        const s = self.returns_kbn.skewnessMoment();
        const k = self.returns_kbn.kurtosisExcess();
        if (math.isNan(s) or math.isNan(k)) return nan;
        const n = fl(self.returns_kbn.n());
        return (n / 6) * (s * s + (k * k) / 4);
    }

    /// Jarque–Bera test of the null hypothesis that returns are normally
    /// distributed (Python default confidence = 0.95).
    ///
    /// True: normality cannot be rejected. False: normality is rejected or
    /// there is insufficient data (JB is NaN). `error.InvalidArgument`
    /// when confidence ∉ (0, 1), including NaN. See `isNormalFromJb`.
    pub fn isNormalDistribution(self: *const Self, confidence: f64) error{InvalidArgument}!bool {
        return isNormalFromJb(self.jarqueBeraNormalityTestStatistic(), confidence);
    }

    // ── VaR, ES, reward-to-risk ──────────────────────────────────────────

    /// Historical VaR, -percentile(returns, 1 - confidence)
    /// (Python default confidence = 0.95). NaN when empty or when
    /// 1 - confidence ∉ [0, 1] (Python raises). Sorts a scratch copy.
    pub fn varHistorical(self: *const Self, confidence: f64) Allocator.Error!f64 {
        return historicalOrNan(core.varHistorical(self.allocator, self.returns.slice(), 0.0, confidence));
    }

    /// Gaussian VaR, -(mean + z·sigma) (Python default confidence = 0.95).
    /// NaN when undefined or when 1 - confidence ∉ (0, 1) (Python raises).
    pub fn varGaussian(self: *const Self, confidence: f64) f64 {
        return parametricOrNan(core.varGaussian(&self.returns_kbn, confidence));
    }

    /// Modified Cornish-Fisher VaR (Python default confidence = 0.95).
    /// NaN when undefined or when 1 - confidence ∉ (0, 1) (Python raises).
    pub fn varCornishFisher(self: *const Self, confidence: f64) f64 {
        return parametricOrNan(core.varCornishFisher(&self.returns_kbn, confidence));
    }

    /// Historical Expected Shortfall (Python default confidence = 0.95).
    /// NaN when empty or when 1 - confidence ∉ [0, 1] (Python raises).
    /// Sorts a scratch copy.
    pub fn esHistorical(self: *const Self, confidence: f64) Allocator.Error!f64 {
        return historicalOrNan(core.esHistorical(self.allocator, self.returns.slice(), 0.0, confidence));
    }

    /// Gaussian Expected Shortfall (Python default confidence = 0.95).
    /// NaN when undefined or when 1 - confidence ∉ (0, 1) (Python raises).
    pub fn esGaussian(self: *const Self, confidence: f64) f64 {
        return parametricOrNan(core.esGaussian(&self.returns_kbn, confidence));
    }

    /// Cornish-Fisher Expected Shortfall (Python default confidence = 0.95).
    /// NaN when undefined or when 1 - confidence ∉ (0, 1) (Python raises).
    pub fn esCornishFisher(self: *const Self, confidence: f64) f64 {
        return parametricOrNan(core.esCornishFisher(&self.returns_kbn, confidence));
    }

    /// Mean excess return divided by the historical VaR of the raw returns
    /// (Python default confidence = 0.95).
    pub fn rewardToVarRatioHistorical(self: *const Self, confidence: f64) Allocator.Error!f64 {
        const denom = try self.varHistorical(confidence);
        return if (denom != 0) self.excess_returns_kbn.mean() / denom else nan;
    }

    /// Mean excess return divided by the Gaussian VaR of the raw returns
    /// (Python default confidence = 0.95).
    pub fn rewardToVarRatioGaussian(self: *const Self, confidence: f64) f64 {
        const denom = self.varGaussian(confidence);
        return if (denom != 0) self.excess_returns_kbn.mean() / denom else nan;
    }

    /// Mean excess return divided by the Cornish-Fisher VaR of the raw
    /// returns (Python default confidence = 0.95).
    pub fn rewardToVarRatioCornishFisher(self: *const Self, confidence: f64) f64 {
        const denom = self.varCornishFisher(confidence);
        return if (denom != 0) self.excess_returns_kbn.mean() / denom else nan;
    }

    /// Mean excess return divided by the historical ES of the raw returns
    /// (Python default confidence = 0.95).
    pub fn rewardToEsRatioHistorical(self: *const Self, confidence: f64) Allocator.Error!f64 {
        const denom = try self.esHistorical(confidence);
        return if (denom != 0) self.excess_returns_kbn.mean() / denom else nan;
    }

    /// Mean excess return divided by the Gaussian ES of the raw returns
    /// (Python default confidence = 0.95).
    pub fn rewardToEsRatioGaussian(self: *const Self, confidence: f64) f64 {
        const denom = self.esGaussian(confidence);
        return if (denom != 0) self.excess_returns_kbn.mean() / denom else nan;
    }

    /// Mean excess return divided by the Cornish-Fisher ES of the raw
    /// returns (Python default confidence = 0.95).
    pub fn rewardToEsRatioCornishFisher(self: *const Self, confidence: f64) f64 {
        const denom = self.esCornishFisher(confidence);
        return if (denom != 0) self.excess_returns_kbn.mean() / denom else nan;
    }

    /// Mean return divided by the mean absolute deviation
    /// sum(|r - mean|) / n. NaN when empty or MAD is 0. O(n).
    pub fn meanAbsoluteDeviationRatio(self: *const Self) f64 {
        const n = self.returns_kbn.n();
        if (n < 1) return nan;
        const mean = self.returns_kbn.mean();
        var sum: KleinKBNAccumulator = .{};
        for (self.returns.slice()) |x| sum.update(@abs(x - mean));
        const mad = sum.value() / fl(n);
        return if (mad > 0) mean / mad else nan;
    }

    // ── Upside / downside partial moments ────────────────────────────────

    /// Upside potential ratio HPM1 / sqrt(LPM2) about the target return.
    pub fn upsidePotentialRatio(self: *const Self) f64 {
        const hpm1 = self.target_partial_moments.higherPartialMoment1();
        const lpm2 = self.target_partial_moments.lowerPartialMoment2();
        if (math.isNan(hpm1) or math.isNan(lpm2) or lpm2 == 0) return nan;
        return hpm1 / @sqrt(lpm2);
    }

    /// Subset upside potential ratio: like `upsidePotentialRatio` but the
    /// moments are averaged over the observations above / below target.
    pub fn upsidePotentialRatioSubset(self: *const Self) f64 {
        const n1 = self.target_partial_moments.upperExcessCount();
        const n2 = self.target_partial_moments.lowerExcessCount();
        if (n1 == 0 or n2 == 0) return nan;
        const hpm1 = self.target_partial_moments.upperExcessMoment1Sum() / fl(n1);
        const lpm2 = self.target_partial_moments.lowerExcessMoment2Sum() / fl(n2);
        if (math.isNan(hpm1) or math.isNan(lpm2) or lpm2 == 0) return nan;
        return hpm1 / @sqrt(lpm2);
    }

    /// Proportion of returns above the target return.
    pub fn upsideFrequency(self: *const Self) f64 {
        return self.target_partial_moments.upsideFrequency();
    }

    /// First-order upper partial moment about the target (HPM1).
    pub fn upsidePotential(self: *const Self) f64 {
        return self.target_partial_moments.higherPartialMoment1();
    }

    /// Average excess return above target over the upside observations
    /// (0.0 when there are none).
    pub fn upsidePotentialSubset(self: *const Self) f64 {
        const n = self.target_partial_moments.upperExcessCount();
        return if (n > 0) self.target_partial_moments.upperExcessMoment1Sum() / fl(n) else 0.0;
    }

    /// Second-order upper partial moment about the target (HPM2).
    pub fn upsideVariance(self: *const Self) f64 {
        return self.target_partial_moments.higherPartialMoment2();
    }

    /// Second-order upper partial moment over the upside observations
    /// (0.0 when there are none).
    pub fn upsideVarianceSubset(self: *const Self) f64 {
        const n = self.target_partial_moments.upperExcessCount();
        return if (n > 0) self.target_partial_moments.upperExcessMoment2Sum() / fl(n) else 0.0;
    }

    /// Upside risk sqrt(HPM2).
    pub fn upsideRisk(self: *const Self) f64 {
        const variance = self.upsideVariance();
        return if (math.isNan(variance)) nan else @sqrt(variance);
    }

    /// Subset upside risk sqrt(upsideVarianceSubset).
    pub fn upsideRiskSubset(self: *const Self) f64 {
        const variance = self.upsideVarianceSubset();
        return if (math.isNan(variance)) nan else @sqrt(variance);
    }

    /// Semi-deviation: sqrt(sum over r < mean of (r - mean)² / n). O(n).
    pub fn semiDeviation(self: *const Self) f64 {
        const n = self.returns_kbn.n();
        if (n == 0) return nan;
        const mean = self.returns_kbn.mean();
        var sum_squared: KleinKBNAccumulator = .{};
        for (self.returns.slice()) |r| {
            const deviation = r - mean;
            if (deviation < 0) sum_squared.update(deviation * deviation);
        }
        // Divide by "full" window length
        return @sqrt(sum_squared.value() / fl(n));
    }

    /// Downside deviation sqrt(LPM2) about the target, divided by the full
    /// number of observations.
    pub fn downsideDeviation(self: *const Self) f64 {
        const denom = self.target_partial_moments.totalCount();
        if (denom == 0) return nan;
        return @sqrt(self.target_partial_moments.lowerExcessMoment2Sum() / fl(denom));
    }

    /// Downside deviation divided by the number of observations below
    /// target (0.0 when there are none).
    pub fn downsideDeviationSubset(self: *const Self) f64 {
        const denom = self.target_partial_moments.lowerExcessCount();
        if (denom == 0) return 0.0;
        return @sqrt(self.target_partial_moments.lowerExcessMoment2Sum() / fl(denom));
    }

    /// Proportion of returns below the target return.
    pub fn downsideFrequency(self: *const Self) f64 {
        return self.target_partial_moments.downsideFrequency();
    }

    /// Average shortfall below the target return (LPM1).
    pub fn downsidePotential(self: *const Self) f64 {
        return self.target_partial_moments.downsidePotential();
    }

    // ── Sharpe family ────────────────────────────────────────────────────

    /// Sharpe ratio: mean excess return / sample (ddof=1) standard
    /// deviation of excess returns.
    pub fn sharpeRatio(self: *const Self) f64 {
        const s = self.excess_returns_kbn.standardDeviationDdof1();
        if (math.isNan(s) or s == 0) return nan;
        return self.excess_returns_kbn.mean() / s;
    }

    /// Modified Sharpe ratio with the historical VaR of excess returns
    /// (Python default confidence = 0.95).
    pub fn sharpeRatioVarHistorical(self: *const Self, confidence: f64) Allocator.Error!f64 {
        if (self.excess_returns_kbn.n() < 2) return nan;
        const denom = try historicalOrNan(core.varHistorical(self.allocator, self.returns.slice(), self.risk_free_rate, confidence));
        if (math.isNan(denom) or denom == 0) return nan;
        return self.excess_returns_kbn.mean() / denom;
    }

    /// Modified Sharpe ratio with the Gaussian VaR of excess returns
    /// (Python default confidence = 0.95).
    pub fn sharpeRatioVarGaussian(self: *const Self, confidence: f64) f64 {
        if (self.excess_returns_kbn.n() < 2) return nan;
        const denom = parametricOrNan(core.varGaussian(&self.excess_returns_kbn, confidence));
        if (math.isNan(denom) or denom == 0) return nan;
        return self.excess_returns_kbn.mean() / denom;
    }

    /// Modified Sharpe ratio with the Cornish-Fisher VaR of excess returns
    /// (Python default confidence = 0.95).
    pub fn sharpeRatioVarCornishFisher(self: *const Self, confidence: f64) f64 {
        if (self.excess_returns_kbn.n() < 2) return nan;
        const denom = parametricOrNan(core.varCornishFisher(&self.excess_returns_kbn, confidence));
        if (math.isNan(denom) or denom == 0) return nan;
        return self.excess_returns_kbn.mean() / denom;
    }

    /// Modified Sharpe ratio with the historical ES of excess returns
    /// (Python default confidence = 0.95).
    pub fn sharpeRatioEsHistorical(self: *const Self, confidence: f64) Allocator.Error!f64 {
        if (self.excess_returns_kbn.n() < 2) return nan;
        const denom = try historicalOrNan(core.esHistorical(self.allocator, self.returns.slice(), self.risk_free_rate, confidence));
        if (math.isNan(denom) or denom == 0) return nan;
        return self.excess_returns_kbn.mean() / denom;
    }

    /// Modified Sharpe ratio with the Gaussian ES of excess returns
    /// (Python default confidence = 0.95).
    pub fn sharpeRatioEsGaussian(self: *const Self, confidence: f64) f64 {
        if (self.excess_returns_kbn.n() < 2) return nan;
        const denom = parametricOrNan(core.esGaussian(&self.excess_returns_kbn, confidence));
        if (math.isNan(denom) or denom == 0) return nan;
        return self.excess_returns_kbn.mean() / denom;
    }

    /// Modified Sharpe ratio with the Cornish-Fisher ES of excess returns
    /// (Python default confidence = 0.95).
    pub fn sharpeRatioEsCornishFisher(self: *const Self, confidence: f64) f64 {
        if (self.excess_returns_kbn.n() < 2) return nan;
        const denom = parametricOrNan(core.esCornishFisher(&self.excess_returns_kbn, confidence));
        if (math.isNan(denom) or denom == 0) return nan;
        return self.excess_returns_kbn.mean() / denom;
    }

    /// Symmetric downside-risk Sharpe ratio (Ziemba, 2005): mean excess
    /// return / (sqrt(2) · semi-deviation). ±inf when the semi-deviation is
    /// 0 (-inf if the mean excess return is negative). O(n).
    pub fn downsideSharpeRatio(self: *const Self) f64 {
        const semi_dev = self.semiDeviation();
        if (math.isNan(semi_dev)) return nan;
        if (semi_dev == 0) {
            return if (self.excess_returns_kbn.mean() < 0) -math.inf(f64) else math.inf(f64);
        }
        return self.excess_returns_kbn.mean() / (sqrt2 * semi_dev);
    }

    /// Adjusted Sharpe ratio (Pezier & White, 2006):
    /// SR · (1 + S·SR/6 - K·SR²/24), S skewness, K excess kurtosis.
    pub fn adjustedSharpeRatio(self: *const Self) f64 {
        const s = self.returns_kbn.skewnessMoment();
        const k = self.returns_kbn.kurtosisExcess();
        if (math.isNan(s) or math.isNan(k)) return nan;
        const sr = self.sharpeRatio();
        if (math.isNan(sr)) return nan;
        return sr * (1 + s * sr / 6 - k * sr * sr / 24);
    }

    /// Skewness-only adjusted Sharpe ratio SR · (1 + S·SR/6).
    pub fn adjustedSharpeRatioSkewOnly(self: *const Self) f64 {
        const s = self.returns_kbn.skewnessMoment();
        if (math.isNan(s)) return nan;
        const sr = self.sharpeRatio();
        if (math.isNan(sr)) return nan;
        return sr * (1 + (s * sr) / 6);
    }

    /// Probabilistic Sharpe Ratio with sample skewness and normal kurtosis
    /// (Python default reference_sr = 0.0).
    pub fn probabilisticSharpeRatio(self: *const Self, reference_sr: f64) f64 {
        return core.probabilisticSharpeRatio(&self.returns_kbn, self.sharpeRatio(), reference_sr, false, true);
    }

    /// Probabilistic Sharpe Ratio with sample skewness and sample kurtosis
    /// (Python default reference_sr = 0.0).
    pub fn probabilisticSharpeRatioFull(self: *const Self, reference_sr: f64) f64 {
        return core.probabilisticSharpeRatio(&self.returns_kbn, self.sharpeRatio(), reference_sr, false, false);
    }

    /// Probabilistic Sharpe Ratio with zero skewness and sample kurtosis
    /// (Python default reference_sr = 0.0).
    pub fn probabilisticSharpeRatioSymmetric(self: *const Self, reference_sr: f64) f64 {
        return core.probabilisticSharpeRatio(&self.returns_kbn, self.sharpeRatio(), reference_sr, true, false);
    }

    /// Probabilistic Sharpe Ratio under Gaussian moments (S = 0, K = 3)
    /// (Python default reference_sr = 0.0).
    pub fn probabilisticSharpeRatioGaussian(self: *const Self, reference_sr: f64) f64 {
        return core.probabilisticSharpeRatio(&self.returns_kbn, self.sharpeRatio(), reference_sr, true, true);
    }

    // ── Sortino, Omega, Kappa and friends ────────────────────────────────

    /// Sortino ratio: mean(r - T) / sqrt(LPM2).
    pub fn sortinoRatio(self: *const Self) f64 {
        const lpm2 = self.target_partial_moments.lowerPartialMoment2();
        if (math.isNan(lpm2) or lpm2 == 0) return nan;
        return self.target_returns_kbn.mean() / @sqrt(lpm2);
    }

    /// Jack Schwager's adjusted Sortino ratio: Sortino / sqrt(2).
    pub fn sortinoRatioSqrt2(self: *const Self) f64 {
        return self.sortinoRatio() / sqrt2;
    }

    /// Sortino-Satchell ratio: mean(r - T) / sqrt(LPM2).
    pub fn sortinoSatchellRatio(self: *const Self) f64 {
        const lpm2 = self.target_partial_moments.lowerPartialMoment2();
        if (math.isNan(lpm2) or lpm2 == 0) return nan;
        return self.target_returns_kbn.mean() / @sqrt(lpm2);
    }

    /// Omega ratio: mean(r - T) / LPM1 + 1.
    pub fn omegaRatio(self: *const Self) f64 {
        const lpm1 = self.target_partial_moments.lowerPartialMoment1();
        if (math.isNan(lpm1) or lpm1 == 0) return nan;
        return self.target_returns_kbn.mean() / lpm1 + 1;
    }

    /// Omega-Sharpe ratio: Omega - 1.
    pub fn omegaSharpeRatio(self: *const Self) f64 {
        return self.omegaRatio() - 1;
    }

    /// Omega excess return (annualized):
    /// ((1 + Gm)^P - 1) - 3 · sigma_d · sigma_d_bench, with annualized
    /// portfolio and benchmark downside deviations. The benchmark downside
    /// deviation is recomputed over the window (O(n)).
    pub fn omegaExcessReturn(self: *const Self) f64 {
        const period = self.periods_per_annum;

        // Annualized portfolio return
        const rp = self.cumulative_return.geometricMeanReturn();
        if (math.isNan(rp)) return nan;
        const rp_annual = math.pow(f64, 1 + rp, period) - 1;

        const sqrt_period = @sqrt(period);

        // Annualized portfolio downside (full) deviation
        const sigma_d_ann = self.downsideDeviation() * sqrt_period;

        // Annualized benchmark downside (full) deviation
        const sigma_d_ann_bench = self.benchmarkDownsideDeviation() * sqrt_period;

        return rp_annual - 3 * sigma_d_ann * sigma_d_ann_bench;
    }

    fn benchmarkDownsideDeviation(self: *const Self) f64 {
        const rb = self.returns_benchmark.slice();
        if (rb.len == 0) return nan;
        var lower_excess_kbn = newMoments();
        for (rb) |r| {
            const excess = r - self.target_return;
            if (excess < 0) lower_excess_kbn.update(-excess);
        }
        return @sqrt(lower_excess_kbn.x2Sum() / fl(rb.len));
    }

    /// Kappa ratio of order 1: mean(r - T) / LPM1.
    pub fn kappa1Ratio(self: *const Self) f64 {
        const lpm = self.target_partial_moments.lowerPartialMoment1();
        if (math.isNan(lpm) or lpm == 0) return nan;
        return self.target_returns_kbn.mean() / lpm;
    }

    /// Kappa ratio of order 2: mean(r - T) / sqrt(LPM2).
    pub fn kappa2Ratio(self: *const Self) f64 {
        const lpm = self.target_partial_moments.lowerPartialMoment2();
        if (math.isNan(lpm) or lpm == 0) return nan;
        return self.target_returns_kbn.mean() / @sqrt(lpm);
    }

    /// Kappa ratio of order 3: mean(r - T) / LPM3^(1/3).
    pub fn kappa3Ratio(self: *const Self) f64 {
        const lpm = self.target_partial_moments.lowerPartialMoment3();
        if (math.isNan(lpm) or lpm == 0) return nan;
        return self.target_returns_kbn.mean() / math.pow(f64, lpm, 1.0 / 3.0);
    }

    /// Kappa ratio of order 4: mean(r - T) / LPM4^(1/4).
    pub fn kappa4Ratio(self: *const Self) f64 {
        const lpm = self.target_partial_moments.lowerPartialMoment4();
        if (math.isNan(lpm) or lpm == 0) return nan;
        return self.target_returns_kbn.mean() / math.pow(f64, lpm, 1.0 / 4.0);
    }

    /// Prospect ratio (Watanabe, as described by Bacon):
    /// ((Σr⁺ + λ·Σr⁻)/n - T) / downside deviation
    /// (Python default lambda_loss = 2.25).
    pub fn prospectRatio(self: *const Self, lambda_loss: f64) f64 {
        const ddev = self.downsideDeviation();
        if (math.isNan(ddev) or ddev == 0) return nan;
        const pm = &self.raw_partial_moments;
        const n = pm.count();
        if (n == 0) return nan;
        const prospect_return = (pm.sumPositive() + lambda_loss * pm.sumNegative()) / fl(n);
        return (prospect_return - self.target_return) / ddev;
    }

    /// PerformanceAnalytics Prospect ratio:
    /// (Σr⁺ + 2.25·Σr⁻ - T) / (downside deviation · n).
    pub fn prospectRatioPerformanceAnalytics(self: *const Self) f64 {
        const lambda_loss = 2.25;
        const ddev = self.downsideDeviation();
        if (math.isNan(ddev) or ddev == 0) return nan;
        const pm = &self.raw_partial_moments;
        const n = pm.count();
        if (n == 0) return nan;
        return (pm.sumPositive() + lambda_loss * pm.sumNegative() - self.target_return) / (ddev * fl(n));
    }

    /// Bernardo-Ledoit ratio: Σmax(r, 0) / Σmax(-r, 0).
    pub fn bernardoLedoitRatio(self: *const Self) f64 {
        const lpm_1 = self.raw_partial_moments.lowerPartialMoment1();
        const hpm_1 = self.raw_partial_moments.higherPartialMoment1();
        return if (lpm_1 != 0) hpm_1 / lpm_1 else nan;
    }

    /// D-ratio: (-n_down · Σr⁻) / (n_up · Σr⁺). +inf when there are no
    /// positive returns (including the empty case), 0.0 when there are no
    /// negative returns.
    pub fn dRatio(self: *const Self) f64 {
        const n_up = self.win_loss.winningReturnsCount();
        if (n_up == 0) return math.inf(f64); // No positive returns
        const n_down = self.win_loss.losingReturnsCount();
        if (n_down == 0) return 0.0; // No negative returns
        const sum_up = self.win_loss.winningReturnsSum();
        const sum_down = self.win_loss.losingReturnsSum();
        return (-fl(n_down) * sum_down) / (fl(n_up) * sum_up);
    }

    /// Gain-loss ratio: Σ positive returns / |Σ negative returns|.
    pub fn gainLossRatio(self: *const Self) f64 {
        const sum_losses = @abs(self.win_loss.losingReturnsSum());
        return if (sum_losses != 0) self.win_loss.winningReturnsSum() / sum_losses else nan;
    }

    // ── Win / loss ───────────────────────────────────────────────────────

    /// Arithmetic mean of non-zero returns.
    pub fn meanNonZeroReturn(self: *const Self) f64 {
        return self.win_loss.nonZeroReturnsMean();
    }

    /// Arithmetic mean of positive returns.
    pub fn meanWinReturn(self: *const Self) f64 {
        return self.win_loss.winningReturnsMean();
    }

    /// Arithmetic mean of negative returns.
    pub fn meanLossReturn(self: *const Self) f64 {
        return self.win_loss.losingReturnsMean();
    }

    /// Proportion of non-zero returns that are positive.
    pub fn winRate(self: *const Self) f64 {
        const non_zero_count = self.win_loss.nonZeroReturnsCount();
        if (non_zero_count == 0) return nan;
        return fl(self.win_loss.winningReturnsCount()) / fl(non_zero_count);
    }

    /// Proportion of non-zero returns that are negative.
    pub fn lossRate(self: *const Self) f64 {
        const non_zero_count = self.win_loss.nonZeroReturnsCount();
        if (non_zero_count == 0) return nan;
        return fl(self.win_loss.losingReturnsCount()) / fl(non_zero_count);
    }

    /// Variability skewness HPM2 / LPM2 about the target.
    pub fn variabilitySkewness(self: *const Self) f64 {
        const up_moment = self.target_partial_moments.higherPartialMoment2();
        const down_moment = self.target_partial_moments.lowerPartialMoment2();
        if (math.isNan(up_moment) or math.isNan(down_moment) or down_moment == 0) return nan;
        return up_moment / down_moment;
    }

    /// Volatility skewness sqrt(HPM2 / LPM2) about the target.
    pub fn volatilitySkewness(self: *const Self) f64 {
        const var_skew = self.variabilitySkewness();
        return if (!math.isNan(var_skew)) @sqrt(var_skew) else nan;
    }

    /// Farinelli-Tibiletti ratio HPM_p^(1/p) / LPM_q^(1/q) about the target
    /// (Python defaults upper_order = 2, lower_order = 2).
    ///
    /// `error.InvalidArgument` unless both orders are in 1..4.
    pub fn farinelliTibilettiRatio(self: *const Self, upper_order: u32, lower_order: u32) error{InvalidArgument}!f64 {
        if (upper_order < 1 or upper_order > 4) return error.InvalidArgument;
        if (lower_order < 1 or lower_order > 4) return error.InvalidArgument;
        const pm = &self.target_partial_moments;

        const denom = switch (lower_order) {
            1 => pm.lowerPartialMoment1(),
            2 => @sqrt(pm.lowerPartialMoment2()),
            3 => math.pow(f64, pm.lowerPartialMoment3(), 1.0 / 3.0),
            else => math.pow(f64, pm.lowerPartialMoment4(), 1.0 / 4.0),
        };
        const num = switch (upper_order) {
            1 => pm.higherPartialMoment1(),
            2 => @sqrt(pm.higherPartialMoment2()),
            3 => math.pow(f64, pm.higherPartialMoment3(), 1.0 / 3.0),
            else => math.pow(f64, pm.higherPartialMoment4(), 1.0 / 4.0),
        };

        if (math.isNan(num) or math.isNan(denom) or denom == 0) return nan;
        return num / denom;
    }

    /// Rachev ratio (PerformanceAnalytics non-parametric estimate): upper
    /// tail ES at level beta divided by lower tail ES at level alpha
    /// (Python defaults alpha = 0.1, beta = 0.1). O(n log n).
    ///
    /// NaN when there are fewer than two observations (checked first);
    /// `error.InvalidArgument` when alpha or beta ∉ (0, 1).
    pub fn rachevRatio(self: *const Self, alpha: f64, beta: f64) Error!f64 {
        const n = self.returns_kbn.n();
        const returns = self.returns.slice();
        if (n < 2) return nan;
        if (!(0 < alpha and alpha < 1)) return error.InvalidArgument;
        if (!(0 < beta and beta < 1)) return error.InvalidArgument;

        const sorted_returns = try self.allocator.dupe(f64, returns);
        defer self.allocator.free(sorted_returns);
        std.mem.sort(f64, sorted_returns, {}, std.sort.asc(f64));

        const tail = try self.allocator.alloc(f64, returns.len);
        defer self.allocator.free(tail);

        // Lower-tail VaR and Expected Shortfall.
        const lower_var = try percentileValid(self.allocator, returns, alpha);
        var len: usize = 0;
        for (returns) |r| {
            if (r <= lower_var) {
                tail[len] = r;
                len += 1;
            }
        }
        if (len == 0) return nan;
        const es_lower = -pySum(tail[0..len]) / fl(len);

        // Upper-tail VaR and Expected Shortfall. PerformanceAnalytics uses
        // n.upper <- floor((1-beta) * n); VaR.hat.upper <- sorted[n.upper]
        // (1-based).
        var upper_position: f64 = @floor((1.0 - beta) * fl(n));
        if (upper_position < 1) {
            upper_position = 1;
        } else if (upper_position > fl(n)) {
            upper_position = fl(n);
        }
        const upper_var = sorted_returns[@as(usize, @intFromFloat(upper_position)) - 1];
        len = 0;
        for (returns) |r| {
            if (r >= upper_var) {
                tail[len] = r;
                len += 1;
            }
        }
        if (len == 0 or es_lower == 0) return nan;
        const es_upper = pySum(tail[0..len]) / fl(len);
        return es_upper / es_lower;
    }

    // ── Drawdowns ────────────────────────────────────────────────────────

    /// Drawdown series of cumulative geometric returns (PerformanceAnalytics
    /// `Drawdowns()`), one value per observation in the window.
    ///
    /// Returns a new slice owned by the caller (free with the `Measures`
    /// allocator).
    pub fn drawdownsCumulative(self: *const Self) Allocator.Error![]f64 {
        return self.allocator.dupe(f64, self.drawdown_high_watermark.drawdowns());
    }

    /// Minimum (most negative) cumulative drawdown (NaN when empty). O(n).
    pub fn minDrawdownsCumulative(self: *const Self) f64 {
        return self.drawdown_high_watermark.maximumDrawdown();
    }

    /// Magnitude of the worst cumulative drawdown (NaN when empty). O(n).
    pub fn worstDrawdownsCumulative(self: *const Self) f64 {
        return @abs(self.drawdown_high_watermark.maximumDrawdown());
    }

    /// Drawdown series measured from the high-water mark (same values as
    /// `drawdownsCumulative`).
    ///
    /// Returns a new slice owned by the caller (free with the `Measures`
    /// allocator).
    pub fn drawdownsHighWatermark(self: *const Self) Allocator.Error![]f64 {
        return self.allocator.dupe(f64, self.drawdown_high_watermark.drawdowns());
    }

    /// Drawdowns for continuous losing-return runs, one (negative) value
    /// per run in R's percent convention (Burke), oldest first. When
    /// `max_runs` is positive, the runs are sorted worst first and
    /// truncated to `max_runs` (Python default max_runs = None → null).
    ///
    /// Returns a new slice owned by the caller (free with the `Measures`
    /// allocator).
    pub fn drawdownsContinuousRuns(self: *const Self, max_runs: ?usize) Allocator.Error![]f64 {
        const drawdowns = try self.drawdown_continuous_runs.drawdowns(self.allocator);
        if (drawdowns.len < 1) return drawdowns;
        if (max_runs) |mr| {
            if (mr > 0) {
                // Ascending: most negative (worst) first
                std.mem.sort(f64, drawdowns, {}, std.sort.asc(f64));
                if (mr < drawdowns.len) {
                    if (self.allocator.resize(drawdowns, mr)) return drawdowns[0..mr];
                    defer self.allocator.free(drawdowns);
                    return self.allocator.dupe(f64, drawdowns[0..mr]);
                }
            }
        }
        return drawdowns;
    }

    /// Calmar ratio: geometric mean return / |maximum drawdown|
    /// (not annualized).
    pub fn calmarRatio(self: *const Self) f64 {
        const wdd = self.worstDrawdownsCumulative();
        if (wdd == 0) return nan;
        const cagr = self.cumulative_return.geometricMeanReturn();
        if (math.isNan(cagr)) return nan;
        return cagr / wdd;
    }

    /// Sterling ratio: geometric mean return / (|maximum drawdown| + excess)
    /// (Python default excess = 0.1).
    pub fn sterlingRatio(self: *const Self, excess: f64) f64 {
        const wdd = self.worstDrawdownsCumulative() + excess;
        if (wdd == 0) return nan;
        const cagr = self.cumulative_return.geometricMeanReturn();
        if (math.isNan(cagr)) return nan;
        return cagr / wdd;
    }

    /// Burke ratio: (Gm - rf) / sqrt(Σ continuous drawdown²).
    pub fn burkeRatio(self: *const Self) f64 {
        const rate = self.cumulative_return.geometricMeanReturn() - self.risk_free_rate;
        if (math.isNan(rate)) return nan;
        const d = self.drawdown_continuous_runs.sqrtSumDrawdownsSquared();
        if (d == 0) return nan;
        return rate / d;
    }

    /// Modified Burke ratio: Burke ratio · sqrt(n).
    pub fn burkeRatioModified(self: *const Self) f64 {
        const burke = self.burkeRatio();
        if (math.isNan(burke)) return nan;
        return burke * @sqrt(fl(self.returns_kbn.n()));
    }

    /// Pain index: average depth below the high-water mark, -mean(D).
    pub fn painIndex(self: *const Self) f64 {
        // By calculation, all values are <= 0, so we don't need abs()
        return -self.drawdown_high_watermark.drawdownsMean();
    }

    /// Pain ratio: (Gm - rf) / pain index.
    pub fn painRatio(self: *const Self) f64 {
        const rate = self.cumulative_return.geometricMeanReturn() - self.risk_free_rate;
        if (math.isNan(rate)) return nan;
        const pain_index = self.painIndex();
        return if (pain_index != 0) rate / pain_index else nan;
    }

    /// Ulcer index: root-mean-square high-water-mark drawdown.
    pub fn ulcerIndex(self: *const Self) f64 {
        return @sqrt(self.drawdown_high_watermark.drawdownsSquaredMean());
    }

    /// Martin ratio: (Gm - rf) / ulcer index.
    pub fn martinRatio(self: *const Self) f64 {
        const rate = self.cumulative_return.geometricMeanReturn() - self.risk_free_rate;
        if (math.isNan(rate)) return nan;
        const ulcer_index = self.ulcerIndex();
        return if (ulcer_index != 0) rate / ulcer_index else nan;
    }

    /// Average drawdown episode depth magnitude (0.0 when none).
    pub fn drawdownAverage(self: *const Self) f64 {
        return self.drawdown_episodes.averageEpisodeDrawdown();
    }

    /// Average drawdown episode length in observations (0.0 when none).
    pub fn drawdownAverageLength(self: *const Self) f64 {
        return self.drawdown_episodes.averageEpisodeLength();
    }

    /// Average drawdown peak-to-trough length (0.0 when none).
    pub fn drawdownAveragePeakToTrough(self: *const Self) f64 {
        return self.drawdown_episodes.averageEpisodePeakToTrough();
    }

    /// Average drawdown recovery length (0.0 when none).
    pub fn drawdownAverageRecovery(self: *const Self) f64 {
        return self.drawdown_episodes.averageEpisodeRecovery();
    }

    /// Drawdown deviation sqrt(Σ depth² / n_observations) (0.0 when none).
    pub fn drawdownDeviation(self: *const Self) f64 {
        return @sqrt(self.drawdown_episodes.averageEpisodeDrawdownSquared());
    }

    /// Conditional Drawdown at Risk over the continuous drawdown path: the
    /// mean magnitude of the drawdowns at or below their (1 - confidence)
    /// linear percentile (Python default confidence = 0.95). 0.0 when
    /// empty or the percentile is non-negative.
    ///
    /// `error.InvalidArgument` when confidence ∉ (0, 1).
    pub fn cdarAverage(self: *const Self, confidence: f64) Error!f64 {
        if (!(0.0 < confidence and confidence < 1.0)) return error.InvalidArgument;

        const drawdowns = self.drawdown_high_watermark.drawdowns();
        if (drawdowns.len == 0) return 0.0;

        const q = try percentileValid(self.allocator, drawdowns, 1.0 - confidence);
        if (q >= 0.0) return 0.0;

        var tail_sum: KleinKBNAccumulator = .{};
        var tail_len: usize = 0;
        for (drawdowns) |dd| {
            if (dd <= q) {
                tail_len += 1;
                tail_sum.update(dd);
            }
        }
        return if (tail_len > 0) -tail_sum.value() / fl(tail_len) else 0.0;
    }

    /// Conditional Drawdown at Risk over discrete drawdown episode depths
    /// (PerformanceAnalytics default; Python default confidence = 0.95).
    /// 0.0 when there are no episodes.
    ///
    /// `error.InvalidArgument` when confidence ∉ (0, 1).
    pub fn cdarDiscrete(self: *const Self, confidence: f64) Error!f64 {
        if (!(0.0 < confidence and confidence < 1.0)) return error.InvalidArgument;

        const depths = try self.drawdown_episodes.depths(self.allocator);
        defer self.allocator.free(depths);
        if (depths.len == 0) return 0.0;

        const q = try percentileValid(self.allocator, depths, 1.0 - confidence);

        var tail_sum: KleinKBNAccumulator = .{};
        var tail_len: usize = 0;
        for (depths) |depth| {
            if (depth <= q) {
                tail_len += 1;
                tail_sum.update(depth);
            }
        }
        return if (tail_len > 0) -tail_sum.value() / fl(tail_len) else 0.0;
    }

    /// Conditional Drawdown Beta: the mean compounded portfolio return over
    /// the peak-to-trough periods of the worst benchmark drawdown episodes,
    /// divided by the tail threshold depth (Python default
    /// confidence = 0.95). NaN when there are no benchmark episodes or the
    /// threshold depth is zero.
    ///
    /// `error.InvalidArgument` when confidence ∉ (0, 1).
    pub fn cdarBeta(self: *const Self, confidence: f64) Error!f64 {
        if (!(0.0 < confidence and confidence < 1.0)) return error.InvalidArgument;

        const w = self.returns.slice();
        const episodes = try self.drawdown_episodes_benchmark.episodes(self.allocator);
        defer self.allocator.free(episodes);
        if (episodes.len == 0) return nan;

        const depths = try self.drawdown_episodes_benchmark.depths(self.allocator);
        defer self.allocator.free(depths);
        std.mem.sort(f64, depths, {}, std.sort.asc(f64));

        const ceil_count = @ceil(fl(depths.len) * (1.0 - confidence));
        const tail_count: usize = if (ceil_count < 1) 1 else @intFromFloat(ceil_count);
        const q = depths[tail_count - 1];
        if (q == 0.0) return nan;

        var sum_ret: KleinKBNAccumulator = .{};
        var ret: KleinKBNAccumulator = .{};
        var tail_len: usize = 0;
        for (episodes) |episode| {
            if (episode.depth <= q) {
                tail_len += 1;
                ret.reset();
                var i = episode.from_idx;
                while (i <= episode.trough_idx) : (i += 1) ret.update(math.log1p(w[i]));
                sum_ret.update(math.expm1(ret.value()));
            }
        }
        return if (tail_len != 0) sum_ret.value() / (fl(tail_len) * q) else nan;
    }

    /// Conditional Drawdown Alpha: ((1 + mean(r))^P - 1) -
    /// cdarBeta · ((1 + mean(b))^P - 1) (Python default confidence = 0.95).
    /// NaN when the CDaR beta is NaN.
    ///
    /// `error.InvalidArgument` when confidence ∉ (0, 1).
    pub fn cdarAlpha(self: *const Self, confidence: f64) Error!f64 {
        const beta = try self.cdarBeta(confidence);
        if (math.isNan(beta)) return nan;
        const period = self.periods_per_annum;
        const r_mean = self.returns_kbn.mean();
        const b_mean = self.benchmark_returns_kbn.mean();
        const r_annual = math.pow(f64, 1.0 + r_mean, period) - 1.0;
        const b_annual = math.pow(f64, 1.0 + b_mean, period) - 1.0;
        return r_annual - beta * b_annual;
    }

    /// Reward to conditional drawdown: geometric mean return divided by the
    /// mean magnitude of the worst max(1, int(n · (1 - confidence)))
    /// drawdowns (Python default confidence = 0.95). NaN when empty.
    /// Returns `error.InvalidArgument` unless confidence is in (0, 1).
    pub fn rewardToConditionalDrawdown(self: *const Self, confidence: f64) Error!f64 {
        if (!(0 < confidence and confidence < 1)) return error.InvalidArgument;
        const cagr = self.cumulative_return.geometricMeanReturn();
        if (math.isNan(cagr)) return nan;

        const dd = try self.allocator.dupe(f64, self.drawdown_high_watermark.drawdowns());
        defer self.allocator.free(dd);
        if (dd.len < 1) return nan;

        const t = fl(dd.len) * (1 - confidence);
        const n_tail: usize = if (t >= fl(dd.len)) dd.len else if (t < 1) 1 else @intFromFloat(t);
        std.mem.sort(f64, dd, {}, std.sort.asc(f64)); // Most negative first
        const sorted_tail = dd[0..n_tail];
        // Positive number
        const cdar = -pySum(sorted_tail) / fl(sorted_tail.len);
        return if (cdar != 0) cagr / cdar else nan;
    }

    // ── Single-factor model / benchmark ──────────────────────────────────

    /// SFM risk premium: mean periodic excess return over the risk-free
    /// rate.
    pub fn sfmRiskPremium(self: *const Self) f64 {
        return self.excess_returns_kbn.mean();
    }

    /// SFM alpha: intercept of the regression of r - rf on b - rf.
    pub fn sfmAlpha(self: *const Self) f64 {
        return self.sfm_regression.alpha();
    }

    /// SFM beta: slope of the regression of r - rf on b - rf.
    pub fn sfmBeta(self: *const Self) f64 {
        return self.sfm_regression.beta();
    }

    /// SFM bull beta: slope over the periods with b - rf > 0.
    pub fn sfmBetaBull(self: *const Self) f64 {
        return self.sfm_regression.betaBull();
    }

    /// SFM bear beta: slope over the periods with b - rf < 0.
    pub fn sfmBetaBear(self: *const Self) f64 {
        return self.sfm_regression.betaBear();
    }

    /// Timing ratio: bull beta / bear beta.
    pub fn timingRatio(self: *const Self) f64 {
        const denom = self.sfm_regression.betaBear();
        return if (denom != 0) self.sfm_regression.betaBull() / denom else nan;
    }

    /// Coefficient of determination (R²) of the single-factor model.
    pub fn sfmR2(self: *const Self) f64 {
        return self.sfm_regression.r2();
    }

    /// Annualized Jensen's alpha:
    /// Gann_p - (beta · Gann_b + (1 - beta) · rf_annual).
    pub fn jensenAlpha(self: *const Self) f64 {
        const rf = self.annual_risk_free_rate;
        const mean = self.cumulative_return.annualizedGeometricMeanReturn(self.periods_per_annum);
        const mean_b = self.benchmark_cumulative_return.annualizedGeometricMeanReturn(self.periods_per_annum);
        const beta = self.sfmBeta();
        // This form retains mean_b when beta == 1, even for a huge rf.
        return mean - (beta * mean_b + (1.0 - beta) * rf);
    }

    /// Fama beta: population standard deviation of r / that of b.
    pub fn famaBeta(self: *const Self) f64 {
        const sigma = self.returns_kbn.standardDeviationDdof0();
        const sigma_b = self.benchmark_returns_kbn.standardDeviationDdof0();
        return if (sigma_b != 0) sigma / sigma_b else nan;
    }

    /// Modigliani–Modigliani measure (periodic): rf + mean(e) · sigma_b /
    /// sigma_e (population standard deviations).
    pub fn modigliani(self: *const Self) f64 {
        const sigma = self.excess_returns_kbn.standardDeviationDdof0();
        if (sigma == 0) return nan;
        const sigma_b = self.benchmark_returns_kbn.standardDeviationDdof0();
        return self.risk_free_rate + self.excess_returns_kbn.mean() * sigma_b / sigma;
    }

    /// Annualized tracking error: sample standard deviation of r - b ·
    /// sqrt(P).
    pub fn trackingError(self: *const Self) f64 {
        return self.active_returns_kbn.standardDeviationDdof1() * self.sqrt_periods_per_annum;
    }

    /// Annualized active premium: Gann_p - Gann_b.
    pub fn activePremium(self: *const Self) f64 {
        const mean = self.cumulative_return.annualizedGeometricMeanReturn(self.periods_per_annum);
        const mean_b = self.benchmark_cumulative_return.annualizedGeometricMeanReturn(self.periods_per_annum);
        return mean - mean_b;
    }

    /// Annualized information ratio: active premium / tracking error.
    pub fn informationRatio(self: *const Self) f64 {
        const te = self.trackingError();
        return if (te != 0) self.activePremium() / te else nan;
    }

    /// Information ratio when the annualized geometric active premium is
    /// positive, and its negation otherwise.
    pub fn informationRatioModified(self: *const Self) f64 {
        const excess = self.activePremium();
        const ir = self.informationRatio();
        if (math.isNan(excess) or math.isNan(ir)) return nan;
        return if (excess > 0) ir else -ir;
    }

    /// Annualized systematic risk: |beta| · sample sigma(b - rf) · sqrt(P).
    pub fn systematicRisk(self: *const Self) f64 {
        const beta = self.sfmBeta();
        if (math.isNan(beta)) return nan;
        const benchmark_risk = self.benchmark_excess_returns_kbn.standardDeviationDdof1();
        if (math.isNan(benchmark_risk)) return nan;
        return @abs(beta) * benchmark_risk * self.sqrt_periods_per_annum;
    }

    /// Annualized Treynor ratio: annualized geometric excess return / beta.
    pub fn treynorRatio(self: *const Self) f64 {
        const beta = self.sfmBeta();
        if (beta == 0) return nan;
        return self.cumulative_excess_return.annualizedGeometricMeanReturn(self.periods_per_annum) / beta;
    }

    /// Modified Treynor ratio: annualized geometric excess return /
    /// systematic risk.
    pub fn treynorRatioModified(self: *const Self) f64 {
        const sr = self.systematicRisk();
        if (sr == 0) return nan;
        return self.cumulative_excess_return.annualizedGeometricMeanReturn(self.periods_per_annum) / sr;
    }

    /// Annualized specific (residual) risk: population standard deviation
    /// of r - rf - alpha - beta·(b - rf), times sqrt(P). O(n).
    pub fn specificRisk(self: *const Self) f64 {
        const beta = self.sfmBeta();
        if (math.isNan(beta)) return nan;
        const alpha = self.sfmAlpha();
        if (math.isNan(alpha)) return nan;
        var epsilon_kbn: RawMomentsKleinKBN = .{ .ddof = 0, .bias = true, .fisher = true };
        const rf = self.risk_free_rate;
        const r_b = self.returns_benchmark.slice();
        for (self.returns.slice(), 0..) |r, i| {
            epsilon_kbn.update(r - rf - alpha - beta * (r_b[i] - rf));
        }
        return epsilon_kbn.standardDeviationDdof0() * self.sqrt_periods_per_annum;
    }

    /// Annualized total risk sqrt(systematic² + specific²). O(n).
    pub fn totalRisk(self: *const Self) f64 {
        const syr = self.systematicRisk();
        if (math.isNan(syr)) return nan;
        const spr = self.specificRisk();
        if (math.isNan(spr)) return nan;
        return @sqrt(syr * syr + spr * spr);
    }

    /// Appraisal ratio: Jensen's alpha / specific risk. O(n).
    pub fn appraisalRatio(self: *const Self) f64 {
        const alpha = self.jensenAlpha();
        if (math.isNan(alpha)) return nan;
        const spr = self.specificRisk();
        return if (spr != 0) alpha / spr else nan;
    }

    /// Modified Jensen's alpha: Jensen's alpha / beta.
    pub fn jensenAlphaModified(self: *const Self) f64 {
        const alpha = self.jensenAlpha();
        if (math.isNan(alpha)) return nan;
        const beta = self.sfmBeta();
        return if (beta != 0) alpha / beta else nan;
    }

    /// Alternative Jensen's alpha: Jensen's alpha / systematic risk.
    pub fn jensenAlphaAlternative(self: *const Self) f64 {
        const alpha = self.jensenAlpha();
        if (math.isNan(alpha)) return nan;
        const spr = self.systematicRisk();
        return if (spr != 0) alpha / spr else nan;
    }

    /// M-squared (annualized, geometric):
    /// Gann_p · s + rf_annual · (1 - s), s = sigma_b / sigma_p (population).
    pub fn mSquared(self: *const Self) f64 {
        const p_ret = self.cumulative_return.annualizedGeometricMeanReturn(self.periods_per_annum);
        if (math.isNan(p_ret)) return nan;
        const p_std = self.returns_kbn.standardDeviationDdof0() * self.sqrt_periods_per_annum;
        if (math.isNan(p_std) or p_std == 0) return nan;
        const b_std = self.benchmark_returns_kbn.standardDeviationDdof0() * self.sqrt_periods_per_annum;
        if (math.isNan(b_std)) return nan;
        const scale = b_std / p_std;
        // Keep p_ret when scale == 1 instead of subtracting two large rf terms.
        return p_ret * scale + self.annual_risk_free_rate * (1.0 - scale);
    }

    /// Geometric excess M-squared: (1 + M²) / (1 + Gann_b) - 1.
    pub fn mSquaredExcess(self: *const Self) f64 {
        const m_sq = self.mSquared();
        if (math.isNan(m_sq)) return nan;
        const b_ret = self.benchmark_cumulative_return.annualizedGeometricMeanReturn(self.periods_per_annum);
        if (math.isNan(b_ret)) return nan;
        return (1.0 + m_sq) / (1.0 + b_ret) - 1.0;
    }

    /// M² Sortino: Gann_p + Sortino · sqrt(P) · (dd_b - dd_p), with
    /// downside deviations about the target.
    pub fn mSquaredSortino(self: *const Self) f64 {
        const sortino = self.sortinoRatio();
        if (math.isNan(sortino)) return nan;
        const p_ret = self.cumulative_return.annualizedGeometricMeanReturn(self.periods_per_annum);
        if (math.isNan(p_ret)) return nan;
        const p_dd = self.downsideDeviation();
        if (math.isNan(p_dd)) return nan;
        const b_count = self.benchmark_target_partial_moments.totalCount();
        if (b_count == 0) return nan;
        const b_dd = @sqrt(self.benchmark_target_partial_moments.lowerExcessMoment2Sum() / fl(b_count));
        if (math.isNan(b_dd)) return nan;
        return p_ret + sortino * self.sqrt_periods_per_annum * (b_dd - p_dd);
    }

    // ── Miscellaneous ────────────────────────────────────────────────────

    /// Tail ratio (not in R): percentile(cutoff) / |percentile(1 - cutoff)|
    /// (Python default cutoff = 0.95). NaN with fewer than two
    /// observations.
    ///
    /// `error.InvalidArgument` unless 0.5 < cutoff < 1.
    pub fn tailRatio(self: *const Self, cutoff: f64) Error!f64 {
        if (!(0.5 < cutoff and cutoff < 1.0)) return error.InvalidArgument;
        const w = self.returns.slice();
        if (w.len < 2) return nan;
        const right_tail = try percentileValid(self.allocator, w, cutoff);
        const left_tail = try percentileValid(self.allocator, w, 1 - cutoff);
        return if (left_tail != 0) right_tail / @abs(left_tail) else nan;
    }

    /// Full Kelly fraction: mean(e) / Var(e) (sample variance, ddof=1).
    pub fn kellyRatioFull(self: *const Self) f64 {
        const mean_excess = self.excess_returns_kbn.mean();
        const var_excess = self.excess_returns_kbn.variance();
        return if (var_excess != 0) mean_excess / var_excess else nan;
    }

    /// Half-Kelly fraction: kellyRatioFull / 2.
    pub fn kellyRatio(self: *const Self) f64 {
        return self.kellyRatioFull() / 2;
    }

    /// Hurst exponent from single-scale rescaled range analysis:
    /// ln(R/S) / ln(n). O(n).
    pub fn hurstExponent(self: *const Self) f64 {
        const n = self.returns_kbn.n();
        if (n < 2) return nan;
        const mean = self.returns_kbn.mean();
        const s = self.returns_kbn.standardDeviationDdof1();
        if (s == 0) return nan;
        var cum_sum: KleinKBNAccumulator = .{};
        var cum_min = math.inf(f64);
        var cum_max = -math.inf(f64);
        for (self.returns.slice()) |x| {
            cum_sum.update(x - mean); // Demean
            const val = cum_sum.value();
            if (cum_min > val) cum_min = val;
            if (cum_max < val) cum_max = val;
        }
        const delta = cum_max - cum_min;
        const rescaled_range = delta / s;
        if (rescaled_range <= 0) return nan;
        return @log(rescaled_range) / @log(fl(n));
    }

    /// Bias ratio: count(0 <= r <= k·sigma) / (1 + count(-k·sigma <= r < 0))
    /// (Python default std_dev_multiplier = 1.0). O(n).
    ///
    /// `error.InvalidArgument` when std_dev_multiplier is nonpositive or NaN.
    pub fn biasRatio(self: *const Self, std_dev_multiplier: f64) error{InvalidArgument}!f64 {
        if (!(std_dev_multiplier > 0)) return error.InvalidArgument;
        const s = self.returns_kbn.standardDeviationDdof1();
        if (math.isNan(s) or s == 0) return nan;
        const threshold = std_dev_multiplier * s;
        var count_positive: usize = 0;
        var count_negative: usize = 0;
        for (self.returns.slice()) |x| {
            if (0 <= x and x <= threshold) {
                count_positive += 1;
            } else if (-threshold <= x and x < 0) {
                count_negative += 1;
            }
        }
        return fl(count_positive) / fl(1 + count_negative);
    }

    /// K-ratio (Lars Kestner): slope of the regression of the cumulative
    /// log equity curve on time divided by (its standard error · sqrt(n)).
    /// NaN with fewer than three observations. O(n), plain sums.
    pub fn kRatio(self: *const Self) f64 {
        const n_u = self.returns_kbn.n();
        if (n_u < 3) return nan;
        const n = fl(n_u);
        const w = self.returns.slice();

        // Equity curve (cumulative log returns), linear regression sums.
        var cum_sum: f64 = 0.0;
        var sum_t: f64 = 0.0;
        var sum_t2: f64 = 0.0;
        var sum_eq: f64 = 0.0;
        var sum_te: f64 = 0.0;
        for (w, 0..) |x, i| {
            cum_sum += math.log1p(x);
            const t_val = fl(i);
            sum_t += t_val;
            sum_t2 += t_val * t_val;
            sum_eq += cum_sum;
            sum_te += t_val * cum_sum;
        }

        const t_mean = sum_t / n;
        const equity_mean = sum_eq / n;
        const s_tt = sum_t2 - n * (t_mean * t_mean);
        const s_te = sum_te - n * t_mean * equity_mean;
        if (s_tt == 0) return nan;

        const slope = s_te / s_tt;
        const intercept = equity_mean - slope * t_mean;

        var sum_sq_residuals: f64 = 0.0;
        cum_sum = 0.0;
        for (w, 0..) |x, i| {
            cum_sum += math.log1p(x);
            const predicted = intercept + slope * fl(i);
            const residual = cum_sum - predicted;
            sum_sq_residuals += residual * residual;
        }
        var residual_var = sum_sq_residuals / (n - 2);
        if (residual_var < 0) residual_var = 0.0; // Handle floating point noise
        const se_slope = @sqrt(residual_var / s_tt);
        if (se_slope == 0) return nan;
        return slope / (se_slope * @sqrt(n));
    }

    /// Jack Schwager's gain-to-pain ratio: Σr / Σmax(-r, 0).
    /// NaN when there are no losses.
    pub fn gainToPainRatio(self: *const Self) f64 {
        const lpm1 = self.raw_partial_moments.lowerPartialMoment1();
        if (math.isNan(lpm1) or lpm1 == 0) return nan;
        return self.returns_kbn.x1Sum() / lpm1;
    }

    // ── Capture ──────────────────────────────────────────────────────────

    /// Upside capture ratio over periods with benchmark > 0
    /// (Python default geometric = true).
    pub fn upsideCaptureRatio(self: *const Self, geometric: bool) f64 {
        if (geometric) return self.capture.upsideCaptureRatioGeometric();
        return self.capture.upsideCaptureRatioArithmetic();
    }

    /// Downside capture ratio over periods with benchmark <= 0
    /// (Python default geometric = true).
    pub fn downsideCaptureRatio(self: *const Self, geometric: bool) f64 {
        if (geometric) return self.capture.downsideCaptureRatioGeometric();
        return self.capture.downsideCaptureRatioArithmetic();
    }

    /// Overall capture ratio: upside / downside capture
    /// (Python default geometric = true).
    pub fn overallCaptureRatio(self: *const Self, geometric: bool) f64 {
        const up = self.upsideCaptureRatio(geometric);
        const down = self.downsideCaptureRatio(geometric);
        if (math.isNan(up) or math.isNan(down) or down == 0) return nan;
        return up / down;
    }

    /// Up number ratio: share of benchmark-up periods with a positive
    /// portfolio return.
    pub fn upNumberRatio(self: *const Self) f64 {
        return self.capture.upNumberRatio();
    }

    /// Down number ratio: share of benchmark-down (b <= 0) periods with a
    /// negative portfolio return.
    pub fn downNumberRatio(self: *const Self) f64 {
        return self.capture.downNumberRatio();
    }

    /// Up percentage ratio: share of benchmark-up periods in which the
    /// portfolio outperformed the benchmark.
    pub fn upPercentageRatio(self: *const Self) f64 {
        return self.capture.upPercentageRatio();
    }

    /// Down percentage ratio: share of benchmark-down (b < 0) periods in
    /// which the portfolio outperformed the benchmark.
    pub fn downPercentageRatio(self: *const Self) f64 {
        return self.capture.downPercentageRatio();
    }
};
