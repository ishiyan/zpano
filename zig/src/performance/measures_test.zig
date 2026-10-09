//! Tests for `Measures` (port of Python `performance/test_measures.py`),
//! part 1: helpers, data, edge cases and measures up to the Rachev ratio.
//! Part 2 lives in `measures_test_2.zig`.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const ta = testing.allocator;

const measures_mod = @import("measures.zig");
pub const Measures = measures_mod.Measures;
const RawMomentsKleinKBN = @import("raw_moments_klein_kbn").RawMomentsKleinKBN;
pub const rd = @import("reference_data/reference_data.zig");

test {
    _ = @import("measures_test_2.zig");
}

// ── Data ─────────────────────────────────────────────────────────────────
//
// 'Portfolio bacon' dataset from the PerformanceAnalytics R package: Bacon,
// Practical portfolio performance measurement and attribution, 2nd ed.
// (2008), p. 65 (portfolio) and p. 66 (benchmark).

pub const bacon_portfolio_returns = [_]f64{
    0.003,  0.026,  0.011,  -0.010,
    0.015,  0.025,  0.016,  0.067,
    -0.014, 0.040,  -0.005, 0.081,
    0.040,  -0.037, -0.061, 0.017,
    -0.049, -0.022, 0.070,  0.058,
    -0.065, 0.024,  -0.005, -0.009,
};
pub const bacon_benchmark_returns = [_]f64{
    0.002,  0.025,  0.018,  -0.011,
    0.014,  0.018,  0.014,  0.065,
    -0.015, 0.042,  -0.006, 0.083,
    0.039,  -0.038, -0.062, 0.015,
    -0.048, 0.021,  0.060,  0.056,
    -0.067, 0.019,  -0.003, 0.000,
};
pub const bacon_portfolio_len = bacon_portfolio_returns.len;

// Extended Bacon 2023 (3rd edition) portfolio data.
pub const bacon_2023_portfolio_returns = [_]f64{
    0.003,  0.026, 0.011,  -0.009, 0.014,  0.024, 0.015,  0.066,  -0.014, 0.039,
    -0.005, 0.081, 0.040,  -0.037, -0.061, 0.014, -0.049, -0.021, 0.062,  0.058,
    -0.064, 0.017, -0.004, -0.002, -0.021, 0.011, 0.047,  0.024,  0.033,  -0.007,
    0.047,  0.006, 0.010,  -0.002, 0.034,  0.010,
};
pub const bacon_2023_drawdown_continuous = [_]f64{
    0,       0, 0, -0.0090, 0,       0, 0, 0,       -0.0140, 0,
    -0.0050, 0, 0, 0,       -0.0960, 0, 0, -0.0690, 0,       0,
    -0.0640, 0, 0, 0,       -0.0270, 0, 0, 0,       0,       -0.0070,
    0,       0, 0, -0.0020, 0,       0,
};
pub const bacon_2023_drawdown_continuous_without_zeroes = [_]f64{
    -0.0090, -0.0140, -0.0050, -0.0960, -0.0690,
    -0.0640, -0.0270, -0.0070, -0.0020,
};
pub const bacon_2023_drawdown_from_peak = [_]f64{
    0,       0,       0,       -0.0090, 0,       0,       0,       0,       -0.0140, 0,
    -0.0050, 0,       0,       -0.0370, -0.0957, -0.0831, -0.1280, -0.1463, -0.0934, -0.0408,
    -0.1022, -0.0869, -0.0906, -0.0924, -0.1115, -0.1017, -0.0595, -0.0369, -0.0051, -0.0121,
    0,       0,       0,       -0.0020, 0,       0,
};
pub const bacon_2023_portfolio_len = bacon_2023_portfolio_returns.len;

pub const sqrt2: f64 = 1.4142135623730950488016887242097;
const nan = math.nan(f64);

// ── Assertion helpers (Python assertFloatEqual / assertSeriesEqual) ─────

/// Tolerance options: `places` as in assertAlmostEqual (|a-e| < 0.5e-places),
/// `delta` absolute tolerance, `rel_tol` (with `delta` as abs_tol) as in
/// math.isclose, `skip` leading series elements, `report` prints failures.
pub const Tol = struct {
    places: u32 = 15,
    delta: ?f64 = null,
    rel_tol: ?f64 = null,
    skip: usize = 0,
    report: bool = true,
};

pub fn floatEqual(actual: f64, expected: f64, tol: Tol) bool {
    if (math.isNan(expected)) return math.isNan(actual);
    if (math.isInf(expected)) return math.isInf(actual) and ((actual > 0) == (expected > 0));
    if (math.isNan(actual) or math.isInf(actual)) return false;
    const diff = @abs(actual - expected);
    if (tol.rel_tol) |rt| {
        const abs_tol = tol.delta orelse 0.0;
        return diff <= @max(rt * @max(@abs(actual), @abs(expected)), abs_tol);
    }
    if (actual == expected) return true;
    if (tol.delta) |d| return diff <= d;
    return diff < 0.5 * math.pow(f64, 10.0, -@as(f64, @floatFromInt(tol.places)));
}

pub fn expectFloat(actual: f64, expected: f64, tol: Tol, comptime fmt: []const u8, args: anytype) !void {
    if (!floatEqual(actual, expected, tol)) {
        if (tol.report) std.debug.print(fmt ++ ": expected {d}, actual {d}\n", args ++ .{ expected, actual });
        return error.TestExpectedApproxEq;
    }
}

pub fn expectSeries(actual: []const f64, expected: []const f64, tol: Tol, comptime fmt: []const u8, args: anytype) !void {
    if (actual.len != expected.len) {
        if (tol.report) std.debug.print(fmt ++ ": series length {d} != {d}\n", args ++ .{ actual.len, expected.len });
        return error.TestExpectedEqual;
    }
    for (actual, expected, 0..) |a, e, i| {
        if (i < tol.skip) continue;
        if (!floatEqual(a, e, tol)) {
            if (tol.report) std.debug.print(fmt ++ " step {d}: expected {d}, actual {d}\n", args ++ .{ i, e, a });
            return error.TestExpectedApproxEq;
        }
    }
}

pub fn expectInvalid(r: anytype) !void {
    if (r) |_| {
        return error.TestExpectedError;
    } else |e| {
        try testing.expectEqual(error.InvalidArgument, e);
    }
}

// ── Numeric helpers for the expected values ──────────────────────────────

/// Python built-in `sum()` over floats (Neumaier, CPython >= 3.12).
pub fn pySum(values: []const f64) f64 {
    var f_result: f64 = 0.0;
    var c: f64 = 0.0;
    for (values) |x| {
        const t = f_result + x;
        if (@abs(f_result) >= @abs(x)) c += (f_result - t) + x else c += (x - t) + f_result;
        f_result = t;
    }
    if (c != 0 and math.isFinite(c)) f_result += c;
    return f_result;
}

/// math.fsum (Shewchuk partials).
pub fn fsum(values: []const f64) f64 {
    var partials: [128]f64 = undefined;
    var np: usize = 0;
    for (values) |x0| {
        var x = x0;
        var i: usize = 0;
        for (partials[0..np]) |y0| {
            var y = y0;
            if (@abs(x) < @abs(y)) std.mem.swap(f64, &x, &y);
            const hi = x + y;
            const lo = y - (hi - x);
            if (lo != 0) {
                partials[i] = lo;
                i += 1;
            }
            x = hi;
        }
        np = i;
        partials[np] = x;
        np += 1;
    }
    var total: f64 = 0;
    var k = np;
    while (k > 0) {
        k -= 1;
        total += partials[k];
    }
    return total;
}

pub fn prod1p(values: []const f64) f64 {
    var p: f64 = 1;
    for (values) |x| p *= 1 + x;
    return p;
}

pub fn fmean(values: []const f64) f64 {
    return fsum(values) / @as(f64, @floatFromInt(values.len));
}

fn sumSqDev(values: []const f64) f64 {
    const m = fmean(values);
    var buf: [512]f64 = undefined;
    for (values, 0..) |x, i| buf[i] = (x - m) * (x - m);
    return fsum(buf[0..values.len]);
}

/// statistics.pstdev
pub fn pstdev(values: []const f64) f64 {
    return @sqrt(sumSqDev(values) / @as(f64, @floatFromInt(values.len)));
}

/// statistics.stdev
pub fn stdev(values: []const f64) f64 {
    return @sqrt(sumSqDev(values) / @as(f64, @floatFromInt(values.len - 1)));
}

pub const Period = enum { yearly, monthly, daily };

/// Annual rate for a periodic rate as in the Python tests:
/// yearly as-is, monthly (1+r)**12-1, daily (1+r)**252-1.
pub fn annualize(rate: f64, period: Period) f64 {
    return switch (period) {
        .yearly => rate,
        .monthly => math.pow(f64, 1 + rate, 12) - 1,
        .daily => math.pow(f64, 1 + rate, 252) - 1,
    };
}

// ── Streaming helpers (run_stream_method / run_stream_property) ──────────

pub const Cfg = struct {
    daily: bool = false,
    monthly: bool = false,
    rf: f64 = 0,
    mar: f64 = 0,
    returns: []const f64 = &bacon_portfolio_returns,
    bench: []const f64 = &bacon_benchmark_returns,
    window: usize = 0,
    start: usize = 0,
};

pub fn periodsPerAnnum(daily: bool, monthly: bool) f64 {
    return if (daily) 252 else if (monthly) 12 else 1;
}

pub fn makeMeasures(cfg: Cfg) !Measures {
    var m = try Measures.init(ta, periodsPerAnnum(cfg.daily, cfg.monthly), cfg.rf, cfg.mar, cfg.window);
    m.reset();
    return m;
}

pub fn addBacon(m: *Measures, returns: []const f64, bench: []const f64) !void {
    for (returns, bench) |r, b| try m.addReturn(r, b);
}

/// Calls a measure method with extra arguments, unwrapping error unions.
pub fn call(m: *const Measures, comptime method: anytype, args: anytype) !Unwrapped(@TypeOf(@call(.auto, method, .{m} ++ args))) {
    const r = @call(.auto, method, .{m} ++ args);
    return if (@typeInfo(@TypeOf(r)) == .error_union) try r else r;
}

fn Unwrapped(comptime T: type) type {
    return switch (@typeInfo(T)) {
        .error_union => |eu| eu.payload,
        else => T,
    };
}

/// Streams the configured returns and collects `method(args...)` after
/// every observation. Caller frees.
pub fn runStream(cfg: Cfg, comptime method: anytype, args: anytype) ![]f64 {
    var m = try makeMeasures(cfg);
    defer m.deinit();
    const out = try ta.alloc(f64, cfg.returns.len - cfg.start);
    errdefer ta.free(out);
    for (cfg.start..cfg.returns.len) |i| {
        try m.addReturn(cfg.returns[i], cfg.bench[i]);
        out[i - cfg.start] = try call(&m, method, args);
    }
    return out;
}

/// Streams and collects list-valued measures. Free with `freeLists`.
pub fn runStreamLists(cfg: Cfg, comptime method: anytype, args: anytype) ![][]f64 {
    var m = try makeMeasures(cfg);
    defer m.deinit();
    const out = try ta.alloc([]f64, cfg.returns.len - cfg.start);
    var filled: usize = 0;
    errdefer {
        for (out[0..filled]) |l| ta.free(l);
        ta.free(out);
    }
    for (cfg.start..cfg.returns.len) |i| {
        try m.addReturn(cfg.returns[i], cfg.bench[i]);
        out[i - cfg.start] = try call(&m, method, args);
        filled += 1;
    }
    return out;
}

pub fn freeLists(lists: [][]f64) void {
    for (lists) |l| ta.free(l);
    ta.free(lists);
}

/// Streams with a callback (run_stream_callback).
pub fn runStreamCallback(cfg: Cfg, ctx: f64, comptime f: fn (*const Measures, f64) f64) ![]f64 {
    var m = try makeMeasures(cfg);
    defer m.deinit();
    const out = try ta.alloc(f64, cfg.returns.len - cfg.start);
    errdefer ta.free(out);
    for (cfg.start..cfg.returns.len) |i| {
        try m.addReturn(cfg.returns[i], cfg.bench[i]);
        out[i - cfg.start] = f(&m, ctx);
    }
    return out;
}

/// Stream a measure over a table keyed by a periodic rate (rf or MAR),
/// annualized per `period`.
pub fn checkByRate(
    entries: []const rd.Entry(f64, []const f64),
    period: Period,
    is_mar: bool,
    comptime method: anytype,
    args: anytype,
    tol: Tol,
    comptime label: []const u8,
) !void {
    for (entries) |e| {
        const rate = annualize(e.key, period);
        const cfg: Cfg = .{
            .daily = period == .daily,
            .monthly = period == .monthly,
            .rf = if (is_mar) 0 else rate,
            .mar = if (is_mar) rate else 0,
        };
        const actual = try runStream(cfg, method, args);
        defer ta.free(actual);
        try expectSeries(actual, e.value, tol, label ++ " ({s}, rate {d})", .{ @tagName(period), e.key });
    }
}

// ── Public measures (Python inspect-based public_measures()) ────────────

pub const Value = union(enum) {
    f: f64,
    b: bool,
    list: []f64,

    pub fn deinit(self: Value) void {
        if (self == .list) ta.free(self.list);
    }
};

/// Every public property and method of `Measures` (except mutators),
/// methods called with their Python default arguments.
pub const public_measures = .{
    .{ "active_premium", Measures.activePremium, .{} },
    .{ "adjusted_sharpe_ratio", Measures.adjustedSharpeRatio, .{} },
    .{ "adjusted_sharpe_ratio_skew_only", Measures.adjustedSharpeRatioSkewOnly, .{} },
    .{ "appraisal_ratio", Measures.appraisalRatio, .{} },
    .{ "autocorrelation_penalty", Measures.autocorrelationPenalty, .{} },
    .{ "bernardo_ledoit_ratio", Measures.bernardoLedoitRatio, .{} },
    .{ "bias_ratio", Measures.biasRatio, .{@as(f64, 1.0)} },
    .{ "burke_ratio", Measures.burkeRatio, .{} },
    .{ "burke_ratio_modified", Measures.burkeRatioModified, .{} },
    .{ "calmar_ratio", Measures.calmarRatio, .{} },
    .{ "cdar_alpha", Measures.cdarAlpha, .{@as(f64, 0.95)} },
    .{ "cdar_average", Measures.cdarAverage, .{@as(f64, 0.95)} },
    .{ "cdar_beta", Measures.cdarBeta, .{@as(f64, 0.95)} },
    .{ "cdar_discrete", Measures.cdarDiscrete, .{@as(f64, 0.95)} },
    .{ "compound_annual_growth_rate", Measures.compoundAnnualGrowthRate, .{} },
    .{ "cumulative_geometric_return", Measures.cumulativeGeometricReturn, .{} },
    .{ "d_ratio", Measures.dRatio, .{} },
    .{ "down_number_ratio", Measures.downNumberRatio, .{} },
    .{ "down_percentage_ratio", Measures.downPercentageRatio, .{} },
    .{ "downside_capture_ratio", Measures.downsideCaptureRatio, .{true} },
    .{ "downside_deviation", Measures.downsideDeviation, .{} },
    .{ "downside_deviation_subset", Measures.downsideDeviationSubset, .{} },
    .{ "downside_frequency", Measures.downsideFrequency, .{} },
    .{ "downside_potential", Measures.downsidePotential, .{} },
    .{ "downside_sharpe_ratio", Measures.downsideSharpeRatio, .{} },
    .{ "drawdown_average", Measures.drawdownAverage, .{} },
    .{ "drawdown_average_length", Measures.drawdownAverageLength, .{} },
    .{ "drawdown_average_peak_to_trough", Measures.drawdownAveragePeakToTrough, .{} },
    .{ "drawdown_average_recovery", Measures.drawdownAverageRecovery, .{} },
    .{ "drawdown_deviation", Measures.drawdownDeviation, .{} },
    .{ "drawdowns_continuous_runs", Measures.drawdownsContinuousRuns, .{@as(?usize, null)} },
    .{ "drawdowns_cumulative", Measures.drawdownsCumulative, .{} },
    .{ "drawdowns_high_watermark", Measures.drawdownsHighWatermark, .{} },
    .{ "es_cornish_fisher", Measures.esCornishFisher, .{@as(f64, 0.95)} },
    .{ "es_gaussian", Measures.esGaussian, .{@as(f64, 0.95)} },
    .{ "es_historical", Measures.esHistorical, .{@as(f64, 0.95)} },
    .{ "fama_beta", Measures.famaBeta, .{} },
    .{ "farinelli_tibiletti_ratio", Measures.farinelliTibilettiRatio, .{ @as(u32, 2), @as(u32, 2) } },
    .{ "gain_loss_ratio", Measures.gainLossRatio, .{} },
    .{ "gain_to_pain_ratio", Measures.gainToPainRatio, .{} },
    .{ "geometric_mean_return", Measures.geometricMeanReturn, .{} },
    .{ "hurst_exponent", Measures.hurstExponent, .{} },
    .{ "information_ratio", Measures.informationRatio, .{} },
    .{ "information_ratio_modified", Measures.informationRatioModified, .{} },
    .{ "is_normal_distribution", Measures.isNormalDistribution, .{@as(f64, 0.95)} },
    .{ "jarque_bera_normality_test_statistic", Measures.jarqueBeraNormalityTestStatistic, .{} },
    .{ "jensen_alpha", Measures.jensenAlpha, .{} },
    .{ "jensen_alpha_alternative", Measures.jensenAlphaAlternative, .{} },
    .{ "jensen_alpha_modified", Measures.jensenAlphaModified, .{} },
    .{ "k_ratio", Measures.kRatio, .{} },
    .{ "kappa_1_ratio", Measures.kappa1Ratio, .{} },
    .{ "kappa_2_ratio", Measures.kappa2Ratio, .{} },
    .{ "kappa_3_ratio", Measures.kappa3Ratio, .{} },
    .{ "kappa_4_ratio", Measures.kappa4Ratio, .{} },
    .{ "kelly_ratio", Measures.kellyRatio, .{} },
    .{ "kelly_ratio_full", Measures.kellyRatioFull, .{} },
    .{ "kurtosis", Measures.kurtosis, .{} },
    .{ "kurtosis_excess", Measures.kurtosisExcess, .{} },
    .{ "kurtosis_moment", Measures.kurtosisMoment, .{} },
    .{ "kurtosis_sample", Measures.kurtosisSample, .{} },
    .{ "kurtosis_sample_corrected", Measures.kurtosisSampleCorrected, .{} },
    .{ "kurtosis_sample_excess", Measures.kurtosisSampleExcess, .{} },
    .{ "loss_rate", Measures.lossRate, .{} },
    .{ "m_squared", Measures.mSquared, .{} },
    .{ "m_squared_excess", Measures.mSquaredExcess, .{} },
    .{ "m_squared_sortino", Measures.mSquaredSortino, .{} },
    .{ "martin_ratio", Measures.martinRatio, .{} },
    .{ "mean_absolute_deviation_ratio", Measures.meanAbsoluteDeviationRatio, .{} },
    .{ "mean_loss_return", Measures.meanLossReturn, .{} },
    .{ "mean_non_zero_return", Measures.meanNonZeroReturn, .{} },
    .{ "mean_win_return", Measures.meanWinReturn, .{} },
    .{ "min_drawdowns_cumulative", Measures.minDrawdownsCumulative, .{} },
    .{ "modigliani", Measures.modigliani, .{} },
    .{ "omega_excess_return", Measures.omegaExcessReturn, .{} },
    .{ "omega_ratio", Measures.omegaRatio, .{} },
    .{ "omega_sharpe_ratio", Measures.omegaSharpeRatio, .{} },
    .{ "overall_capture_ratio", Measures.overallCaptureRatio, .{true} },
    .{ "pain_index", Measures.painIndex, .{} },
    .{ "pain_ratio", Measures.painRatio, .{} },
    .{ "probabilistic_sharpe_ratio", Measures.probabilisticSharpeRatio, .{@as(f64, 0.0)} },
    .{ "probabilistic_sharpe_ratio_full", Measures.probabilisticSharpeRatioFull, .{@as(f64, 0.0)} },
    .{ "probabilistic_sharpe_ratio_gaussian", Measures.probabilisticSharpeRatioGaussian, .{@as(f64, 0.0)} },
    .{ "probabilistic_sharpe_ratio_symmetric", Measures.probabilisticSharpeRatioSymmetric, .{@as(f64, 0.0)} },
    .{ "prospect_ratio", Measures.prospectRatio, .{@as(f64, 2.25)} },
    .{ "prospect_ratio_performance_analytics", Measures.prospectRatioPerformanceAnalytics, .{} },
    .{ "rachev_ratio", Measures.rachevRatio, .{ @as(f64, 0.1), @as(f64, 0.1) } },
    .{ "reward_to_conditional_drawdown", Measures.rewardToConditionalDrawdown, .{@as(f64, 0.95)} },
    .{ "reward_to_es_ratio_cornish_fisher", Measures.rewardToEsRatioCornishFisher, .{@as(f64, 0.95)} },
    .{ "reward_to_es_ratio_gaussian", Measures.rewardToEsRatioGaussian, .{@as(f64, 0.95)} },
    .{ "reward_to_es_ratio_historical", Measures.rewardToEsRatioHistorical, .{@as(f64, 0.95)} },
    .{ "reward_to_var_ratio_cornish_fisher", Measures.rewardToVarRatioCornishFisher, .{@as(f64, 0.95)} },
    .{ "reward_to_var_ratio_gaussian", Measures.rewardToVarRatioGaussian, .{@as(f64, 0.95)} },
    .{ "reward_to_var_ratio_historical", Measures.rewardToVarRatioHistorical, .{@as(f64, 0.95)} },
    .{ "semi_deviation", Measures.semiDeviation, .{} },
    .{ "sfm_alpha", Measures.sfmAlpha, .{} },
    .{ "sfm_beta", Measures.sfmBeta, .{} },
    .{ "sfm_beta_bear", Measures.sfmBetaBear, .{} },
    .{ "sfm_beta_bull", Measures.sfmBetaBull, .{} },
    .{ "sfm_r2", Measures.sfmR2, .{} },
    .{ "sfm_risk_premium", Measures.sfmRiskPremium, .{} },
    .{ "sharpe_ratio", Measures.sharpeRatio, .{} },
    .{ "sharpe_ratio_es_cornish_fisher", Measures.sharpeRatioEsCornishFisher, .{@as(f64, 0.95)} },
    .{ "sharpe_ratio_es_gaussian", Measures.sharpeRatioEsGaussian, .{@as(f64, 0.95)} },
    .{ "sharpe_ratio_es_historical", Measures.sharpeRatioEsHistorical, .{@as(f64, 0.95)} },
    .{ "sharpe_ratio_var_cornish_fisher", Measures.sharpeRatioVarCornishFisher, .{@as(f64, 0.95)} },
    .{ "sharpe_ratio_var_gaussian", Measures.sharpeRatioVarGaussian, .{@as(f64, 0.95)} },
    .{ "sharpe_ratio_var_historical", Measures.sharpeRatioVarHistorical, .{@as(f64, 0.95)} },
    .{ "skewness", Measures.skewness, .{} },
    .{ "skewness_fisher", Measures.skewnessFisher, .{} },
    .{ "skewness_kurtosis_ratio", Measures.skewnessKurtosisRatio, .{} },
    .{ "skewness_moment", Measures.skewnessMoment, .{} },
    .{ "skewness_sample", Measures.skewnessSample, .{} },
    .{ "sortino_ratio", Measures.sortinoRatio, .{} },
    .{ "sortino_ratio_sqrt2", Measures.sortinoRatioSqrt2, .{} },
    .{ "sortino_satchell_ratio", Measures.sortinoSatchellRatio, .{} },
    .{ "specific_risk", Measures.specificRisk, .{} },
    .{ "sterling_ratio", Measures.sterlingRatio, .{@as(f64, 0.1)} },
    .{ "systematic_risk", Measures.systematicRisk, .{} },
    .{ "tail_ratio", Measures.tailRatio, .{@as(f64, 0.95)} },
    .{ "timing_ratio", Measures.timingRatio, .{} },
    .{ "total_risk", Measures.totalRisk, .{} },
    .{ "tracking_error", Measures.trackingError, .{} },
    .{ "treynor_ratio", Measures.treynorRatio, .{} },
    .{ "treynor_ratio_modified", Measures.treynorRatioModified, .{} },
    .{ "ulcer_index", Measures.ulcerIndex, .{} },
    .{ "up_number_ratio", Measures.upNumberRatio, .{} },
    .{ "up_percentage_ratio", Measures.upPercentageRatio, .{} },
    .{ "upside_capture_ratio", Measures.upsideCaptureRatio, .{true} },
    .{ "upside_frequency", Measures.upsideFrequency, .{} },
    .{ "upside_potential", Measures.upsidePotential, .{} },
    .{ "upside_potential_ratio", Measures.upsidePotentialRatio, .{} },
    .{ "upside_potential_ratio_subset", Measures.upsidePotentialRatioSubset, .{} },
    .{ "upside_potential_subset", Measures.upsidePotentialSubset, .{} },
    .{ "upside_risk", Measures.upsideRisk, .{} },
    .{ "upside_risk_subset", Measures.upsideRiskSubset, .{} },
    .{ "upside_variance", Measures.upsideVariance, .{} },
    .{ "upside_variance_subset", Measures.upsideVarianceSubset, .{} },
    .{ "var_cornish_fisher", Measures.varCornishFisher, .{@as(f64, 0.95)} },
    .{ "var_gaussian", Measures.varGaussian, .{@as(f64, 0.95)} },
    .{ "var_historical", Measures.varHistorical, .{@as(f64, 0.95)} },
    .{ "variability_skewness", Measures.variabilitySkewness, .{} },
    .{ "volatility_skewness", Measures.volatilitySkewness, .{} },
    .{ "win_rate", Measures.winRate, .{} },
    .{ "worst_drawdowns_cumulative", Measures.worstDrawdownsCumulative, .{} },
};

pub const public_measures_count = public_measures.len;

/// Evaluates every public measure (Python `evaluate`). Free with
/// `freeValues`.
pub fn evaluateAll(m: *const Measures) ![public_measures_count]Value {
    var out: [public_measures_count]Value = undefined;
    var filled: usize = 0;
    errdefer for (out[0..filled]) |v| v.deinit();
    inline for (public_measures, 0..) |pm, i| {
        const v = try call(m, pm[1], pm[2]);
        out[i] = switch (@TypeOf(v)) {
            f64 => .{ .f = v },
            bool => .{ .b = v },
            []f64 => .{ .list = v },
            else => @compileError("unexpected measure type"),
        };
        filled += 1;
    }
    return out;
}

pub fn freeValues(values: []const Value) void {
    for (values) |v| v.deinit();
}

// ── Tests ────────────────────────────────────────────────────────────────

test "TestSeriesAssertions test_rejects_truncated_series" {
    try testing.expectError(error.TestExpectedEqual, expectSeries(&.{1.0}, &.{ 1.0, 2.0 }, .{ .report = false }, "", .{}));
    try testing.expectError(error.TestExpectedEqual, expectSeries(&.{ 1.0, 2.0 }, &.{1.0}, .{ .report = false }, "", .{}));
}

test "TestSeriesAssertions test_accepts_generator" {
    var gen: [2]f64 = undefined;
    for (&gen, 0..) |*g, i| g.* = @floatFromInt(i + 1);
    try expectSeries(&gen, &.{ 1.0, 2.0 }, .{}, "", .{});
}

test "TestSeriesAssertions test_relative_tolerance_for_large_reference_values" {
    try expectSeries(&.{1e12 + 0.1}, &.{1e12}, .{ .rel_tol = 1e-12 }, "", .{});
    try testing.expectError(error.TestExpectedApproxEq, expectSeries(&.{1e12 + 2}, &.{1e12}, .{ .rel_tol = 1e-12, .report = false }, "", .{}));
}

test "TestEdgeCases test_empty" {
    var m = try makeMeasures(.{});
    defer m.deinit();
    const values = try evaluateAll(&m);
    freeValues(&values);
}

test "TestEdgeCases test_single_return" {
    var m = try makeMeasures(.{});
    defer m.deinit();
    try m.addReturn(0.01, 0.02);
    const values = try evaluateAll(&m);
    freeValues(&values);
    try testing.expect(math.isNan(m.sharpeRatio()));
    try expectFloat(m.cumulativeGeometricReturn(), 0.01, .{}, "cumulative geometric return", .{});
}

test "TestEdgeCases test_first_negative_return_is_a_drawdown" {
    var m = try makeMeasures(.{});
    defer m.deinit();
    try m.addReturn(-0.05, -0.02);
    try m.addReturn(0.02, 0.01);
    const hwm = try m.drawdownsHighWatermark();
    defer ta.free(hwm);
    try expectSeries(hwm, &.{ -0.05, -0.031 }, .{}, "drawdowns high watermark", .{});
    const cum = try m.drawdownsCumulative();
    defer ta.free(cum);
    try expectSeries(cum, &.{ -0.05, -0.031 }, .{}, "drawdowns cumulative", .{});
    try expectFloat(m.worstDrawdownsCumulative(), 0.05, .{}, "worst", .{});
    try expectFloat(m.painIndex(), (0.05 + 0.031) / 2.0, .{}, "pain index", .{});
    try expectFloat(m.drawdownAverage(), 0.05, .{}, "drawdown average", .{});
}

test "TestEdgeCases test_long_daily_series" {
    var prng = std.Random.DefaultPrng.init(1);
    const rnd = prng.random();
    var m = try Measures.init(ta, 252.0, 0, 0, 0);
    defer m.deinit();
    for (0..300) |_| {
        const r = rnd.floatNorm(f64) * 0.01 + 0.0005;
        const b = rnd.floatNorm(f64) * 0.01 + 0.0004;
        try m.addReturn(r, b);
    }
    const values = try evaluateAll(&m);
    freeValues(&values);
    try testing.expect(math.isFinite(m.autocorrelationPenalty()));
}

test "TestAutocorrelationPenalty test_metamorphic_properties" {
    const n = bacon_2023_portfolio_len;
    var returns: [n]f64 = undefined;

    // Constant returns
    for (&returns) |*r| r.* = 0.01;
    {
        const actual = try runStream(.{ .daily = true, .returns = &returns, .bench = &returns }, Measures.autocorrelationPenalty, .{});
        defer ta.free(actual);
        try expectFloat(actual[n - 1], 1.0, .{}, "autocorrelation penalty (constant)", .{});
        // Too few observations
        try expectFloat(actual[0], 1.0, .{}, "autocorrelation penalty (len=0)", .{});
        try expectFloat(actual[1], 1.0, .{}, "autocorrelation penalty (len=1)", .{});
    }

    // Positive autocorrelation
    for (&returns, 0..) |*r, i| r.* = 0.01 * @as(f64, @floatFromInt(i));
    {
        const actual = try runStream(.{ .daily = true, .returns = &returns, .bench = &returns }, Measures.autocorrelationPenalty, .{});
        defer ta.free(actual);
        try expectFloat(actual[n - 1], 2.722393904531189, .{}, "autocorrelation penalty (positive)", .{});
    }

    // Negative autocorrelation
    for (&returns, 0..) |*r, i| r.* = if (i % 2 == 0) 0.01 else -0.01;
    {
        const actual = try runStream(.{ .daily = true, .returns = &returns, .bench = &returns }, Measures.autocorrelationPenalty, .{});
        defer ta.free(actual);
        try expectFloat(actual[n - 1], 0.16903085094570597, .{}, "autocorrelation penalty (negative)", .{});
    }

    // Scale and translation invariance
    const expected = try runStream(.{ .daily = true }, Measures.autocorrelationPenalty, .{});
    defer ta.free(expected);
    for ([_]f64{ 4.2, -4.2 }) |scale| {
        for ([_]f64{ 0.042, -0.042 }) |shift| {
            var transformed: [bacon_portfolio_len]f64 = undefined;
            for (&transformed, bacon_portfolio_returns) |*t, r| t.* = scale * r + shift;
            const actual = try runStream(.{ .daily = true, .returns = &transformed, .bench = &transformed }, Measures.autocorrelationPenalty, .{});
            defer ta.free(actual);
            try expectSeries(actual, expected, .{}, "autocorrelation penalty (transform) scale {d} shift {d}", .{ scale, shift });
        }
    }
}

test "TestCumulativeGeometricReturn test_matches_performance_analytics_output" {
    const expected = rd.cumulative_geometric_return.expected_values;
    inline for (.{ Period.yearly, Period.monthly, Period.daily }) |p| {
        const actual = try runStream(.{ .daily = p == .daily, .monthly = p == .monthly }, Measures.cumulativeGeometricReturn, .{});
        defer ta.free(actual);
        try expectSeries(actual, expected, .{ .places = 14 }, "cumulative geometric return ({s})", .{@tagName(p)});
    }
}

test "TestGeometricMeanReturn test_matches_performance_analytics_output" {
    const expected = rd.geometric_mean_return.expected_values_geometric;
    inline for (.{ Period.yearly, Period.monthly, Period.daily }) |p| {
        const actual = try runStream(.{ .daily = p == .daily, .monthly = p == .monthly }, Measures.geometricMeanReturn, .{});
        defer ta.free(actual);
        try expectSeries(actual, expected, .{ .places = 15 }, "geometric mean return ({s})", .{@tagName(p)});
    }
}

fn cagrDefinition(m: *const Measures, periods: f64) f64 {
    const w = m.returns.slice();
    const growth = prod1p(w);
    return math.pow(f64, growth, periods / @as(f64, @floatFromInt(w.len))) - 1;
}

test "TestCompoundAnnualGrowthRate test_annualized_return_definition" {
    const cases = [_]struct { p: Period, ppa: f64, places: u32 }{
        .{ .p = .yearly, .ppa = 1, .places = 15 },
        .{ .p = .monthly, .ppa = 12, .places = 14 },
        .{ .p = .daily, .ppa = 252, .places = 11 },
    };
    for (cases) |c| {
        const cfg: Cfg = .{ .daily = c.p == .daily, .monthly = c.p == .monthly };
        const expected = try runStreamCallback(cfg, c.ppa, cagrDefinition);
        defer ta.free(expected);
        const actual = try runStream(cfg, Measures.compoundAnnualGrowthRate, .{});
        defer ta.free(actual);
        try expectSeries(actual, expected, .{ .places = c.places }, "compound annual growth rate ({s})", .{@tagName(c.p)});
    }
}

fn eql(a: []const u8, b: []const u8) bool {
    return std.mem.eql(u8, a, b);
}

test "TestSkewness test_matches_performance_analytics_output" {
    for (rd.skewness.expected_values_by_method) |e| {
        const actual = if (eql(e.key, "moment"))
            try runStream(.{}, Measures.skewnessMoment, .{})
        else if (eql(e.key, "fisher"))
            try runStream(.{}, Measures.skewnessFisher, .{})
        else
            try runStream(.{}, Measures.skewnessSample, .{});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = 14 }, "skewness_{s}", .{e.key});
        if (eql(e.key, "moment")) {
            const s = try runStream(.{}, Measures.skewness, .{});
            defer ta.free(s);
            try expectSeries(s, e.value, .{ .places = 14 }, "skewness", .{});
        }
    }
}

test "TestSkewness test_raw_moments_klein_kbn" {
    for (rd.skewness.expected_values_by_method) |e| {
        const bias = !eql(e.key, "fisher");
        var kbn: RawMomentsKleinKBN = .{ .ddof = 1, .bias = bias, .fisher = true };
        for (0..bacon_portfolio_len) |i| {
            kbn.update(bacon_portfolio_returns[i]);
            if (eql(e.key, "moment")) {
                try expectFloat(kbn.skewnessMoment(), kbn.skewness(), .{}, "step {d} skewness_{s} vs. skewness", .{ i, e.key });
            } else if (eql(e.key, "fisher")) {
                try expectFloat(kbn.skewnessFisher(), kbn.skewness(), .{}, "step {d} skewness_{s} vs. skewness", .{ i, e.key });
            } else {
                try expectFloat(kbn.skewnessSample(), e.value[i], .{ .places = 14 }, "step {d} skewness_{s}", .{ i, e.key });
            }
        }
    }
}

test "TestKurtosis test_matches_performance_analytics_output" {
    for (rd.kurtosis.expected_values_by_method) |e| {
        const actual = if (eql(e.key, "excess"))
            try runStream(.{}, Measures.kurtosisExcess, .{})
        else if (eql(e.key, "moment"))
            try runStream(.{}, Measures.kurtosisMoment, .{})
        else if (eql(e.key, "sample_corrected"))
            try runStream(.{}, Measures.kurtosisSampleCorrected, .{})
        else
            try runStream(.{}, Measures.kurtosisSampleExcess, .{});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = 13 }, "kurtosis_{s}", .{e.key});
        if (eql(e.key, "excess")) {
            const k = try runStream(.{}, Measures.kurtosis, .{});
            defer ta.free(k);
            try expectSeries(k, e.value, .{ .places = 13 }, "kurtosis", .{});
        }
    }
}

test "TestKurtosis test_raw_moments_klein_kbn" {
    for (rd.kurtosis.expected_values_by_method) |e| {
        var bias = true;
        var fisher = true;
        if (eql(e.key, "moment")) {
            fisher = false;
        } else if (eql(e.key, "sample_corrected")) {
            bias = false;
            fisher = false;
        } else if (eql(e.key, "sample_excess")) {
            bias = false;
        }
        var kbn: RawMomentsKleinKBN = .{ .ddof = 1, .bias = bias, .fisher = fisher };
        for (0..bacon_portfolio_len) |i| {
            kbn.update(bacon_portfolio_returns[i]);
            const actual = if (eql(e.key, "excess"))
                kbn.kurtosisExcess()
            else if (eql(e.key, "moment"))
                kbn.kurtosisMoment()
            else if (eql(e.key, "sample_corrected"))
                kbn.kurtosisSampleCorrected()
            else
                kbn.kurtosisSampleExcess();
            const dispatched = if (eql(e.key, "sample_corrected")) kbn.kurtosisSample() else actual;
            try expectFloat(dispatched, kbn.kurtosis(), .{}, "step {d} kurtosis_{s} / kurtosis", .{ i, e.key });
            try expectFloat(actual, e.value[i], .{ .places = 13 }, "step {d} kurtosis_{s}", .{ i, e.key });
        }
    }
}

test "TestSkewnessKurtosisRatio test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.skewnessKurtosisRatio, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.skewness_kurtosis_ratio.expected_values, .{ .places = 14 }, "skewness-kurtosis ratio", .{});
}

test "TestJarqueBeraNrmalityTestStatistic test_matches_bacon3_output" {
    const actual = try runStream(.{ .returns = &bacon_2023_portfolio_returns, .bench = &bacon_2023_portfolio_returns }, Measures.jarqueBeraNormalityTestStatistic, .{});
    defer ta.free(actual);
    // Chapter 5, exhibit 5.4
    try expectFloat(actual[bacon_2023_portfolio_len - 1], 0.34, .{ .places = 2 }, "Jarque-Bera normality (bacon3)", .{});
}

test "TestJarqueBeraNrmalityTestStatistic test_matches_scipy_output" {
    const actual = try runStream(.{}, Measures.jarqueBeraNormalityTestStatistic, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.jarque_bera_normality_test_statistic.expected_values, .{ .places = 14 }, "Jarque-Bera normality (scipy)", .{});
}

test "TestIsNormalDistribution test_mocked_jb" {
    // Python patches the JB property with a PropertyMock; here the decision
    // rule used by isNormalDistribution is tested directly.
    const isNormalFromJb = measures_mod.isNormalFromJb;
    // Normality accepted: 5.0 < 5.991...
    try testing.expect(try isNormalFromJb(5.0, 0.95));
    // Normality rejected
    try testing.expect(!try isNormalFromJb(10.0, 0.95));
    // NaN statistic
    try testing.expect(!try isNormalFromJb(nan, 0.95));
    // Invalid confidence
    try expectInvalid(isNormalFromJb(0.0, 1.0));
    try expectInvalid(isNormalFromJb(0.0, 0.0));
    // Custom confidence
    try testing.expect(try isNormalFromJb(8.0, 0.99));
    try testing.expect(!try isNormalFromJb(10.0, 0.99));

    // The method validates confidence even when JB is NaN.
    var m = try Measures.init(ta, 1, 0, 0, 0);
    defer m.deinit();
    try testing.expect(!try m.isNormalDistribution(0.95));
    try expectInvalid(m.isNormalDistribution(1.0));
}

test "TestVarCornishFisher test_matches_performance_analytics_output" {
    for (rd.@"var".expected_values_by_p_cornish_fisher) |e| {
        const actual = try runStream(.{}, Measures.varCornishFisher, .{e.key});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = 9 }, "var cornish-fisher p {d}", .{e.key});
    }
}

test "TestVarGaussian test_matches_performance_analytics_output" {
    for (rd.@"var".expected_values_by_p_gaussian) |e| {
        const actual = try runStream(.{}, Measures.varGaussian, .{e.key});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = 9 }, "var gaussian p {d}", .{e.key});
    }
}

test "TestVarHistorical test_matches_performance_analytics_output" {
    for (rd.@"var".expected_values_by_p_historical) |e| {
        const actual = try runStream(.{}, Measures.varHistorical, .{e.key});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = if (e.key == 0.999) 4 else 15 }, "var historical p {d}", .{e.key});
    }
}

test "TestEsCornishFisher test_matches_performance_analytics_output" {
    for (rd.es.expected_values_by_p_cornish_fisher) |e| {
        const actual = try runStream(.{}, Measures.esCornishFisher, .{e.key});
        defer ta.free(actual);
        const places: u32 = if (e.key < 0.995) 9 else if (e.key < 0.999) 8 else 7;
        try expectSeries(actual, e.value, .{ .places = places }, "es cornish-fisher p {d}", .{e.key});
    }
}

test "TestEsGaussian test_matches_performance_analytics_output" {
    for (rd.es.expected_values_by_p_gaussian) |e| {
        const actual = try runStream(.{}, Measures.esGaussian, .{e.key});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = if (e.key < 0.995) 9 else 8 }, "es gaussian p {d}", .{e.key});
    }
}

test "TestEsHistorical test_matches_performance_analytics_output" {
    for (rd.es.expected_values_by_p_historical) |e| {
        const actual = try runStream(.{}, Measures.esHistorical, .{e.key});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = 15 }, "es historical p {d}", .{e.key});
    }
}

/// (name, risk measure, reward-to ratio, sharpe variant)
const reward_pairs = .{
    .{ "var_historical", Measures.varHistorical, Measures.rewardToVarRatioHistorical, Measures.sharpeRatioVarHistorical },
    .{ "var_gaussian", Measures.varGaussian, Measures.rewardToVarRatioGaussian, Measures.sharpeRatioVarGaussian },
    .{ "var_cornish_fisher", Measures.varCornishFisher, Measures.rewardToVarRatioCornishFisher, Measures.sharpeRatioVarCornishFisher },
    .{ "es_historical", Measures.esHistorical, Measures.rewardToEsRatioHistorical, Measures.sharpeRatioEsHistorical },
    .{ "es_gaussian", Measures.esGaussian, Measures.rewardToEsRatioGaussian, Measures.sharpeRatioEsGaussian },
    .{ "es_cornish_fisher", Measures.esCornishFisher, Measures.rewardToEsRatioCornishFisher, Measures.sharpeRatioEsCornishFisher },
};

test "TestRewardToVarEsRatios test_zero_risk_free_rate_equals_sharpe_variants" {
    inline for (reward_pairs) |p| {
        const reward = try runStream(.{}, p[2], .{@as(f64, 0.95)});
        defer ta.free(reward);
        const sharpe = try runStream(.{}, p[3], .{@as(f64, 0.95)});
        defer ta.free(sharpe);
        try expectSeries(reward, sharpe, .{ .places = 14, .skip = 1 }, "reward_to {s}", .{p[0]});
    }
}

test "TestRewardToVarEsRatios test_definition" {
    const annual_rf = 0.05;
    inline for (reward_pairs) |p| {
        for ([_]f64{ 0.9, 0.95 }) |confidence| {
            var m = try makeMeasures(.{ .rf = annual_rf, .monthly = true });
            defer m.deinit();
            var excess: [bacon_portfolio_len]f64 = undefined;
            for (0..bacon_portfolio_len) |i| {
                try m.addReturn(bacon_portfolio_returns[i], bacon_benchmark_returns[i]);
                for (0..i + 1) |j| excess[j] = bacon_portfolio_returns[j] - m.risk_free_rate;
                const excess_mean = fsum(excess[0 .. i + 1]) / @as(f64, @floatFromInt(i + 1));
                const denom = try call(&m, p[1], .{confidence});
                const expected = if (denom != 0) excess_mean / denom else nan;
                const actual = try call(&m, p[2], .{confidence});
                try expectFloat(actual, expected, .{ .places = 14 }, "reward_to {s} confidence {d} step {d}", .{ p[0], confidence, i });
            }
        }
    }
}

test "TestMeanAbsoluteDeviationRatio test_exact_values" {
    // mean / (sum|r - mean| / n), computed with exact rational arithmetic.
    const expected = [_]f64{
        nan,                 1.2608695652173914, 1.5789473684210527, 0.6818181818181818,
        0.9,                 1.129032258064516,  1.308695652173913,  1.2618556701030927,
        0.9623076923076923,  1.0358796296296295, 0.9166666666666666, 0.96045197740113,
        1.0325794291868604,  0.7659033078880407, 0.480644111906311,  0.5178463399879009,
        0.3424072265625,     0.276536312849162,  0.3738222796970257, 0.4428828239908482,
        0.29573420836751435, 0.3247753530166881, 0.3100659077291792, 0.289544235924933,
    };
    const actual = try runStream(.{}, Measures.meanAbsoluteDeviationRatio, .{});
    defer ta.free(actual);
    try expectSeries(actual, &expected, .{}, "mean absolute deviation ratio", .{});
}

test "TestUpsidePotentialRatio test_matches_performance_analytics_output" {
    try checkByRate(rd.upside_potential_ratio.expected_values_by_mar_full, .yearly, true, Measures.upsidePotentialRatio, .{}, .{ .places = 14 }, "upside potential ratio (full)");
}

test "TestUpsidePotentialRatioSubset test_matches_performance_analytics_output" {
    try checkByRate(rd.upside_potential_ratio.expected_values_by_mar_subset, .yearly, true, Measures.upsidePotentialRatioSubset, .{}, .{ .places = 14 }, "upside potential ratio (subset)");
}

test "TestUpsideFrequency test_matches_performance_analytics_output" {
    try checkByRate(rd.upside_frequency.expected_values_by_mar, .yearly, true, Measures.upsideFrequency, .{}, .{}, "upside frequency");
}

test "TestUpsidePotential test_matches_performance_analytics_output" {
    try checkByRate(rd.upside_risk.expected_values_by_mar_potential_full, .yearly, true, Measures.upsidePotential, .{}, .{}, "upside potential (full)");
    try checkByRate(rd.upside_risk.expected_values_by_mar_potential_full, .daily, true, Measures.upsidePotential, .{}, .{}, "upside potential (full)");
}

test "TestUpsidePotentialSubset test_matches_performance_analytics_output" {
    try checkByRate(rd.upside_risk.expected_values_by_mar_potential_subset, .yearly, true, Measures.upsidePotentialSubset, .{}, .{}, "upside potential (subset)");
}

test "TestUpsideVariance test_matches_performance_analytics_output" {
    try checkByRate(rd.upside_risk.expected_values_by_mar_variance_full, .yearly, true, Measures.upsideVariance, .{}, .{}, "upside variance (full)");
}

test "TestUpsideVarianceSubset test_matches_performance_analytics_output" {
    try checkByRate(rd.upside_risk.expected_values_by_mar_variance_subset, .yearly, true, Measures.upsideVarianceSubset, .{}, .{}, "upside variance (subset)");
}

test "TestUpsideRisk test_matches_performance_analytics_output" {
    try checkByRate(rd.upside_risk.expected_values_by_mar_risk_full, .yearly, true, Measures.upsideRisk, .{}, .{}, "upside risk (full)");
}

test "TestUpsideRiskSubset test_matches_performance_analytics_output" {
    try checkByRate(rd.upside_risk.expected_values_by_mar_risk_subset, .yearly, true, Measures.upsideRiskSubset, .{}, .{}, "upside risk (subset)");
    try checkByRate(rd.upside_risk.expected_values_by_mar_risk_subset, .daily, true, Measures.upsideRiskSubset, .{}, .{}, "upside risk (subset)");
}

test "TestSemiDeviation test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.semiDeviation, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.semi_deviation.expected_values, .{}, "semi-deviation", .{});
}

test "TestDownsideDeviation test_matches_performance_analytics_output" {
    try checkByRate(rd.downside_deviation.expected_values_by_mar_full, .yearly, true, Measures.downsideDeviation, .{}, .{}, "downside deviation");
}

test "TestDownsideDeviationSubset test_matches_performance_analytics_output" {
    try checkByRate(rd.downside_deviation.expected_values_by_mar_subset, .yearly, true, Measures.downsideDeviationSubset, .{}, .{}, "downside deviation subset");
}

test "TestDownsideFrequency test_matches_performance_analytics_output" {
    try checkByRate(rd.downside_frequency.expected_values_by_mar, .yearly, true, Measures.downsideFrequency, .{}, .{}, "downside frequency");
}

test "TestDownsidePotential test_matches_performance_analytics_output" {
    try checkByRate(rd.downside_potential.expected_values_by_mar, .yearly, true, Measures.downsidePotential, .{}, .{}, "downside potential");
}

test "TestSharpeRatio test_matches_performance_analytics_output" {
    for (rd.sharpe_ratio.expected_values_by_rf_stdev) |e| {
        const actual = try runStream(.{ .rf = e.key }, Measures.sharpeRatio, .{});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = if (e.key < 0.25) 13 else 12 }, "Sharpe ratio (stdev) Rf {d}", .{e.key});
    }
}

fn checkSharpeByPRf(
    table: []const rd.Entry(f64, []const rd.Entry(f64, []const f64)),
    comptime method: anytype,
    places: u32,
    comptime label: []const u8,
) !void {
    for (table) |pe| {
        for (pe.value) |re| {
            const actual = try runStream(.{ .rf = re.key }, method, .{pe.key});
            defer ta.free(actual);
            try expectSeries(actual, re.value, .{ .places = places }, label ++ " conf {d} Rf {d}", .{ pe.key, re.key });
        }
    }
}

test "TestSharpeRatioVarHistorical test_matches_performance_analytics_output" {
    try checkSharpeByPRf(rd.sharpe_ratio.expected_values_by_p_rf_var_historical, Measures.sharpeRatioVarHistorical, 12, "Sharpe ratio (VaR historical)");
}

test "TestSharpeRatioVarGaussian test_matches_performance_analytics_output" {
    try checkSharpeByPRf(rd.sharpe_ratio.expected_values_by_p_rf_var_gaussian, Measures.sharpeRatioVarGaussian, 5, "Sharpe ratio (VaR Gaussian)");
}

test "TestSharpeRatioVarCornishFisher test_matches_performance_analytics_output" {
    try checkSharpeByPRf(rd.sharpe_ratio.expected_values_by_p_rf_var_cornish_fisher, Measures.sharpeRatioVarCornishFisher, 6, "Sharpe ratio (VaR Cornish-Fisher)");
}

test "TestSharpeRatioEsHistorical test_matches_performance_analytics_output" {
    for (rd.sharpe_ratio.expected_values_by_p_rf_es_historical) |pe| {
        for (pe.value) |re| {
            const p = pe.key;
            const rf = re.key;
            const actual = try runStream(.{ .rf = rf }, Measures.sharpeRatioEsHistorical, .{p});
            defer ta.free(actual);
            try testing.expectEqual(re.value.len, actual.len);
            for (actual, re.value, 0..) |a, e, i| {
                if (p == 0.9 and rf == 0.001 and i == 10) {
                    // The quantile is exactly the second-worst return. This
                    // implementation includes both tied-to-tail observations;
                    // the R reference includes only one.
                    var ex: [11]f64 = undefined;
                    for (&ex, 0..) |*x, j| x.* = bacon_portfolio_returns[j] - rf;
                    const excess_mean = fsum(&ex) / 11.0;
                    try expectFloat(a, excess_mean / 0.013, .{ .places = 13 }, "Sharpe ratio (ES historical) conf {d} Rf {d} step {d}", .{ p, rf, i });
                } else {
                    try expectFloat(a, e, .{ .delta = 1e-12 }, "Sharpe ratio (ES historical) conf {d} Rf {d} step {d}", .{ p, rf, i });
                }
            }
        }
    }
}

test "TestSharpeRatioEsGaussian test_matches_performance_analytics_output" {
    try checkSharpeByPRf(rd.sharpe_ratio.expected_values_by_p_rf_es_gaussian, Measures.sharpeRatioEsGaussian, 7, "Sharpe ratio (ES Gaussian)");
}

test "TestSharpeRatioEsCornishFisher test_matches_performance_analytics_output" {
    try checkSharpeByPRf(rd.sharpe_ratio.expected_values_by_p_rf_es_cornish_fisher, Measures.sharpeRatioEsCornishFisher, 5, "Sharpe ratio (ES Cornish-Fisher)");
}

test "TestDownsideSharpeRatio test_matches_performance_analytics_output" {
    try checkByRate(rd.downside_sharpe_ratio.expected_values_by_rf, .yearly, false, Measures.downsideSharpeRatio, .{}, .{ .places = 13 }, "downside Sharpe ratio");
}

test "TestAdjustedSharpeRatio test_matches_performance_analytics_output" {
    for (rd.adjusted_sharpe_ratio.expected_values_by_rf) |e| {
        const actual = try runStream(.{ .rf = e.key }, Measures.adjustedSharpeRatio, .{});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = if (e.key < 0.2) 12 else 11 }, "adjusted Sharpe ratio (stdev) Rf {d}", .{e.key});
    }
}

test "TestAdjustedSharpeRatioSkewOnly test_scale_and_translation_invariance" {
    const rf = 0.0042;
    const expected = try runStream(.{ .rf = rf }, Measures.adjustedSharpeRatioSkewOnly, .{});
    defer ta.free(expected);
    for ([_]f64{ 4.2, -4.2 }) |scale| {
        for ([_]f64{ 0.042, -0.042 }) |shift| {
            var transformed: [bacon_portfolio_len]f64 = undefined;
            for (&transformed, bacon_portfolio_returns) |*t, r| t.* = scale * r + shift;
            var m = try Measures.init(ta, 1, scale * rf + shift, 0, 0);
            defer m.deinit();
            m.reset();
            for (0..bacon_portfolio_len) |i| {
                try m.addReturn(transformed[i], transformed[i]);
                const a = m.adjustedSharpeRatioSkewOnly();
                try expectFloat(if (scale > 0) a else -a, expected[i], .{ .places = 14 }, "ASR skew-only (scale {d} shift {d}) step {d}", .{ scale, shift, i });
            }
        }
    }
}

fn checkPsr(table: []const rd.Entry(f64, []const rd.Entry(f64, []const f64)), comptime method: anytype, comptime label: []const u8) !void {
    for (table) |se| {
        for (se.value) |re| {
            const actual = try runStream(.{ .rf = re.key }, method, .{se.key});
            defer ta.free(actual);
            try expectSeries(actual, re.value, .{ .places = 14 }, label ++ " reference_sr {d} Rf {d}", .{ se.key, re.key });
        }
    }
}

test "TestProbabilisticSharpeRatio test_matches_performance_analytics_output" {
    try checkPsr(rd.probabilistic_sharpe_ratio.expected_values_by_refsr_rf, Measures.probabilisticSharpeRatio, "probabilistic Sharpe ratio");
}

test "TestProbabilisticSharpeRatioFull test_matches_performance_analytics_output" {
    try checkPsr(rd.probabilistic_sharpe_ratio.expected_values_by_refsr_rf_full, Measures.probabilisticSharpeRatioFull, "probabilistic Sharpe ratio (full)");
}

test "TestProbabilisticSharpeRatioSymmetric test_matches_performance_analytics_output" {
    try checkPsr(rd.probabilistic_sharpe_ratio.expected_values_by_refsr_rf_symmetric, Measures.probabilisticSharpeRatioSymmetric, "probabilistic Sharpe ratio (symmetric)");
}

test "TestProbabilisticSharpeRatioGaussian test_matches_performance_analytics_output" {
    try checkPsr(rd.probabilistic_sharpe_ratio.expected_values_by_refsr_rf_gaussian, Measures.probabilisticSharpeRatioGaussian, "probabilistic Sharpe ratio (Gaussian)");
}

test "TestSortinoRatio test_matches_performance_analytics_output" {
    try checkByRate(rd.sortino_ratio.expected_values_by_mar, .yearly, true, Measures.sortinoRatio, .{}, .{ .places = 14 }, "sortino ratio");
}

test "TestSortinoRatio test_jack_schager_sqrt2_version" {
    for (rd.sortino_ratio.expected_values_by_mar) |e| {
        const expected = try ta.alloc(f64, e.value.len);
        defer ta.free(expected);
        for (expected, e.value) |*x, v| x.* = v / @sqrt(@as(f64, 2.0));
        const actual = try runStream(.{ .mar = e.key }, Measures.sortinoRatioSqrt2, .{});
        defer ta.free(actual);
        try expectSeries(actual, expected, .{ .places = 14 }, "sortino ratio (sqrt2) MAR {d}", .{e.key});
    }
}

test "TestSortinoSatchellRatio test_should_be_computable" {
    const actual = try runStream(.{}, Measures.sortinoSatchellRatio, .{});
    defer ta.free(actual);
    try expectFloat(actual[actual.len - 1], 0.3923720287950653, .{}, "Sortino-Satchell ratio", .{});
}

test "TestOmegaRatio test_matches_performance_analytics_output" {
    try checkByRate(rd.omega_ratio.expected_values_by_mar, .yearly, true, Measures.omegaRatio, .{}, .{ .places = 13 }, "omega ratio");
}

test "TestOmegaSharpeRatio test_matches_performance_analytics_output" {
    try checkByRate(rd.omega_sharpe_ratio.expected_values_by_mar, .yearly, true, Measures.omegaSharpeRatio, .{}, .{ .places = 13 }, "omega Sharpe ratio");
}

test "TestOmegaExcessReturn test_matches_performance_analytics_output" {
    try checkByRate(rd.omega_excess_return.expected_values_by_mar_with_benchmark, .daily, true, Measures.omegaExcessReturn, .{}, .{ .places = 11 }, "omega excess return (with benchmark)");
    for (rd.omega_excess_return.expected_values_by_mar_with_self) |e| {
        const actual = try runStream(.{ .daily = true, .mar = annualize(e.key, .daily), .bench = &bacon_portfolio_returns }, Measures.omegaExcessReturn, .{});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = 11 }, "omega excess return (with self) MAR {d}", .{e.key});
    }
}

test "TestKappaRatio test_matches_performance_analytics_output" {
    try checkByRate(rd.kappa_ratio.expected_values_by_mar_order_1, .yearly, true, Measures.kappa1Ratio, .{}, .{ .places = 13 }, "kappa 1 ratio");
    try checkByRate(rd.kappa_ratio.expected_values_by_mar_order_2, .yearly, true, Measures.kappa2Ratio, .{}, .{ .places = 14 }, "kappa 2 ratio");
    try checkByRate(rd.kappa_ratio.expected_values_by_mar_order_3, .yearly, true, Measures.kappa3Ratio, .{}, .{ .places = 14 }, "kappa 3 ratio");
    try checkByRate(rd.kappa_ratio.expected_values_by_mar_order_4, .yearly, true, Measures.kappa4Ratio, .{}, .{ .places = 14 }, "kappa 4 ratio");
}

test "TestProspectRatio test_matches_performance_analytics_output" {
    try checkByRate(rd.prospect_ratio.expected_values_by_mar_perfan, .yearly, true, Measures.prospectRatioPerformanceAnalytics, .{}, .{ .places = 13 }, "Prospect ratio PerformanceAnalytics version");
}

test "TestProspectRatio test_matches_reference_implementation_output" {
    try checkByRate(rd.prospect_ratio.expected_values_by_mar_reference, .yearly, true, Measures.prospectRatio, .{@as(f64, 2.25)}, .{ .places = 15 }, "Prospect ratio");
}

test "TestBernardoLedoitRatio test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.bernardoLedoitRatio, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.bernardo_ledoit_ratio.expected_values, .{ .places = 13 }, "Bernardo-Ledoit ratio", .{});
}

test "TestDRatio test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.dRatio, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.d_ratio.expected_values_perfan, .{}, "d-ratio", .{});
}

test "TestGainLossRatio test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.gainLossRatio, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.bernardo_ledoit_ratio.expected_values, .{ .places = 13 }, "gain-loss ratio", .{});
}

test "TestMeanNonZeroReturn test_calculated_by_hand" {
    const actual = try runStream(.{}, Measures.meanNonZeroReturn, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.mean_non_zero_return.expected_values, .{}, "mean non-zero return", .{});
}

test "TestMeanWinReturn test_calculated_by_hand" {
    const actual = try runStream(.{}, Measures.meanWinReturn, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.mean_win_return.expected_values, .{}, "mean win return", .{});
}

test "TestMeanLossReturn test_calculated_by_hand" {
    const actual = try runStream(.{}, Measures.meanLossReturn, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.mean_loss_return.expected_values, .{}, "mean loss return", .{});
}

test "TestWinRate test_calculated_by_hand" {
    const actual = try runStream(.{}, Measures.winRate, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.win_rate.expected_values, .{}, "win rate", .{});
}

test "TestLossRate test_calculated_by_hand" {
    const actual = try runStream(.{}, Measures.lossRate, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.loss_rate.expected_values, .{}, "loss rate", .{});
}

test "TestVolatilitySkewness test_matches_performance_analytics_output" {
    for (rd.volatility_skewness.expected_values_by_mar_volatility) |e| {
        const actual = try runStream(.{ .mar = e.key }, Measures.volatilitySkewness, .{});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = 13 }, "volatility skewness MAR {d}", .{e.key});
    }
    for (rd.volatility_skewness.expected_values_by_mar_variability) |e| {
        const actual = try runStream(.{ .mar = e.key }, Measures.variabilitySkewness, .{});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = 13 }, "variability skewness MAR {d}", .{e.key});
    }
}

fn ftVerify(upper: u32, lower: u32, comptime related: anytype, comptime transform: fn (f64) f64, places: u32, mar: f64, comptime label: []const u8) !void {
    const expected = try runStream(.{ .mar = mar }, related, .{});
    defer ta.free(expected);
    const actual = try runStream(.{ .mar = mar }, Measures.farinelliTibilettiRatio, .{ upper, lower });
    defer ta.free(actual);
    for (actual) |*a| a.* = transform(a.*);
    try expectSeries(actual, expected, .{ .places = places }, "Farinelli-Tibiletti ratio (u {d}, l {d}) vs " ++ label, .{ upper, lower });
}

fn identity(r: f64) f64 {
    return r;
}
fn minusOne(r: f64) f64 {
    return r - 1;
}
fn square(r: f64) f64 {
    return r * r;
}

test "TestFarinelliTibilettiRatio test_should_be_computable" {
    const mar = 0.005;
    try ftVerify(1, 1, Measures.omegaRatio, identity, 14, mar, "omega_ratio");
    try ftVerify(1, 1, Measures.kappa1Ratio, minusOne, 14, mar, "kappa_1_ratio");
    try ftVerify(1, 2, Measures.upsidePotentialRatio, identity, 15, mar, "upside_potential_ratio");
    try ftVerify(2, 2, Measures.volatilitySkewness, identity, 14, mar, "volatility_skewness");
    try ftVerify(2, 2, Measures.variabilitySkewness, square, 13, mar, "variability_skewness");

    for ([_]u32{ 1, 2, 3, 4 }) |upper| {
        for ([_]u32{ 1, 2, 3, 4 }) |lower| {
            var up_terms: [bacon_portfolio_len]f64 = undefined;
            var lo_terms: [bacon_portfolio_len]f64 = undefined;
            for (bacon_portfolio_returns, 0..) |r, i| {
                up_terms[i] = math.pow(f64, @max(r - mar, 0.0), @floatFromInt(upper));
                lo_terms[i] = math.pow(f64, @max(mar - r, 0.0), @floatFromInt(lower));
            }
            const upm = pySum(&up_terms) / @as(f64, bacon_portfolio_len);
            const lpm = pySum(&lo_terms) / @as(f64, bacon_portfolio_len);
            const expected = math.pow(f64, upm, 1.0 / @as(f64, @floatFromInt(upper))) / math.pow(f64, lpm, 1.0 / @as(f64, @floatFromInt(lower)));
            const actual = try runStream(.{ .mar = mar }, Measures.farinelliTibilettiRatio, .{ upper, lower });
            defer ta.free(actual);
            try expectFloat(actual[actual.len - 1], expected, .{ .places = 15 }, "Farinelli-Tibiletti ratio (u {d}, l {d}) vs manual calculation", .{ upper, lower });
        }
    }

    // Invalid orders (Python ValueError).
    var m = try makeMeasures(.{});
    defer m.deinit();
    try expectInvalid(m.farinelliTibilettiRatio(0, 2));
    try expectInvalid(m.farinelliTibilettiRatio(2, 5));
}

test "TestRachevRatio test_matches_performance_analytics_output" {
    const tables = .{
        .{ 0.05, rd.rachev_ratio.expected_values_by_beta_rf_alfa_0_05 },
        .{ 0.1, rd.rachev_ratio.expected_values_by_beta_rf_alfa_0_1 },
    };
    inline for (tables) |t| {
        const alpha: f64 = t[0];
        for (t[1]) |be| {
            for (be.value) |re| {
                const actual = try runStream(.{ .rf = re.key }, Measures.rachevRatio, .{ alpha, be.key });
                defer ta.free(actual);
                try expectSeries(actual, re.value, .{ .places = 14 }, "Rachev ratio (alpha {d} beta {d} Rf {d})", .{ alpha, be.key, re.key });
            }
        }
    }
}
