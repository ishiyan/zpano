//! Tests for `Measures` (port of Python `performance/test_measures.py`),
//! part 2: drawdowns, CDaR, single-factor model, benchmark, capture,
//! miscellaneous, documented formulas and rolling windows.

const std = @import("std");
const math = std.math;
const testing = std.testing;
const ta = testing.allocator;

const t = @import("measures_test.zig");
const Measures = t.Measures;
const rd = t.rd;
const Tol = t.Tol;
const Cfg = t.Cfg;
const Period = t.Period;
const expectFloat = t.expectFloat;
const expectSeries = t.expectSeries;
const expectInvalid = t.expectInvalid;
const runStream = t.runStream;
const runStreamLists = t.runStreamLists;
const freeLists = t.freeLists;
const makeMeasures = t.makeMeasures;
const addBacon = t.addBacon;
const checkByRate = t.checkByRate;
const annualize = t.annualize;
const fsum = t.fsum;
const pySum = t.pySum;
const prod1p = t.prod1p;
const fmean = t.fmean;
const pstdev = t.pstdev;
const stdev = t.stdev;
const bp = &t.bacon_portfolio_returns;
const bb = &t.bacon_benchmark_returns;
const n_bacon = t.bacon_portfolio_len;
const nan = math.nan(f64);

fn fl(x: usize) f64 {
    return @floatFromInt(x);
}

test "TestDrawdownsCumulative test_matches_performance_analytics_output" {
    const actual = try runStreamLists(.{}, Measures.drawdownsCumulative, .{});
    defer freeLists(actual);
    for (rd.drawdowns_cumulative.expected_values_by_index) |e| {
        const i: usize = @intFromFloat(e.key);
        try expectSeries(actual[i], e.value, .{}, "drawdowns cumulative (i {d})", .{i});
    }
}

test "TestDrawdownsCumulative test_matches_bacon_2023_output" {
    const actual = try runStreamLists(.{ .returns = &t.bacon_2023_portfolio_returns, .bench = &t.bacon_2023_portfolio_returns }, Measures.drawdownsCumulative, .{});
    defer freeLists(actual);
    try expectSeries(actual[actual.len - 1], &t.bacon_2023_drawdown_from_peak, .{ .places = 4 }, "drawdowns cumulative (bacon 2023)", .{});
}

test "TestMinDrawdownsCumulative test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.minDrawdownsCumulative, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.min_drawdowns_cumulative.expected_values, .{}, "min drawdowns cumulative", .{});
}

test "TestWorstDrawdownsCumulative test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.worstDrawdownsCumulative, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.min_drawdowns_cumulative.expected_values_inverted, .{}, "worst drawdowns cumulative", .{});
}

test "TestDrawdownsHighWatermark test_matches_performance_analytics_output" {
    const actual = try runStreamLists(.{}, Measures.drawdownsHighWatermark, .{});
    defer freeLists(actual);
    for (rd.drawdowns_high_watermark.expected_values_by_index) |e| {
        const i: usize = @intFromFloat(e.key);
        try expectSeries(actual[i], e.value, .{}, "drawdowns high_watermark (i {d})", .{i});
    }
}

test "TestDrawdownsContinuousRuns test_matches_bacon_2023_output" {
    // The book's four-decimal drawdown values are approximate; the core run
    // tracker has separate exact rolling-window tests.
    const actual = try runStreamLists(.{ .returns = &t.bacon_2023_portfolio_returns, .bench = &t.bacon_2023_portfolio_returns }, Measures.drawdownsContinuousRuns, .{@as(?usize, null)});
    defer freeLists(actual);
    try expectSeries(actual[actual.len - 1], &t.bacon_2023_drawdown_continuous_without_zeroes, .{ .delta = 0.002 }, "drawdowns continuous runs (bacon 2023)", .{});
}

test "TestDrawdownsContinuousRuns max_runs truncates worst first" {
    // Not in Python: exercises the max_runs branch (sorted, truncated).
    var m = try makeMeasures(.{});
    defer m.deinit();
    try addBacon(&m, &t.bacon_2023_portfolio_returns, &t.bacon_2023_portfolio_returns);
    const all = try m.drawdownsContinuousRuns(null);
    defer ta.free(all);
    std.mem.sort(f64, all, {}, std.sort.asc(f64));
    const worst3 = try m.drawdownsContinuousRuns(3);
    defer ta.free(worst3);
    try testing.expectEqualSlices(f64, all[0..3], worst3);
    const many = try m.drawdownsContinuousRuns(100);
    defer ta.free(many);
    try testing.expectEqualSlices(f64, all, many);
}

test "TestCalmarRatio test_matches_performance_analytics_output" {
    inline for (.{ false, true }) |daily| {
        const actual = try runStream(.{ .daily = daily }, Measures.calmarRatio, .{});
        defer ta.free(actual);
        try expectSeries(actual, rd.calmar_ratio.expected_values, .{ .places = 13 }, "calmar ratio (daily {})", .{daily});
    }
}

test "TestSterlingRatio test_matches_performance_analytics_output" {
    inline for (.{ false, true }) |daily| {
        for (rd.sterling_ratio.expected_values_by_excess) |e| {
            const actual = try runStream(.{ .daily = daily }, Measures.sterlingRatio, .{e.key});
            defer ta.free(actual);
            try expectSeries(actual, e.value, .{ .places = 13 }, "sterling ratio (daily {}, excess {d})", .{ daily, e.key });
        }
    }
}

test "TestBurkeRatio test_matches_performance_analytics_output" {
    try checkByRate(rd.burke_ratio.expected_values_by_rf, .yearly, false, Measures.burkeRatio, .{}, .{ .places = 11 }, "burke ratio");
}

test "TestBurkeRatioModified test_matches_performance_analytics_output" {
    try checkByRate(rd.burke_ratio.expected_values_by_rf_modified, .yearly, false, Measures.burkeRatioModified, .{}, .{ .places = 11 }, "burke ratio modified");
}

test "TestPainIndex test_matches_performance_analytics_output" {
    // R's drawdown series differs slightly from the high-water-mark series
    // used here. TestDocumentedFormulas checks the exact formula.
    inline for (.{ false, true }) |daily| {
        const actual = try runStream(.{ .daily = daily }, Measures.painIndex, .{});
        defer ta.free(actual);
        try expectSeries(actual, rd.pain_index.expected_values, .{ .delta = 0.00098 }, "pain index (daily {})", .{daily});
    }
}

test "TestPainRatio test_matches_performance_analytics_output" {
    for (rd.pain_ratio.expected_values_by_rf) |e| {
        const tol: Tol = .{ .delta = if (e.key < 0.04) 0.016 else 0.111 };
        const a1 = try runStream(.{ .rf = e.key }, Measures.painRatio, .{});
        defer ta.free(a1);
        try expectSeries(a1, e.value, tol, "pain ratio (yearly, Rf {d})", .{e.key});
        const a2 = try runStream(.{ .daily = true, .rf = annualize(e.key, .daily) }, Measures.painRatio, .{});
        defer ta.free(a2);
        try expectSeries(a2, e.value, tol, "pain ratio (daily, Rf {d})", .{e.key});
    }
}

test "TestUlcerIndex test_matches_performance_analytics_output" {
    inline for (.{ false, true }) |daily| {
        const actual = try runStream(.{ .daily = daily }, Measures.ulcerIndex, .{});
        defer ta.free(actual);
        try expectSeries(actual, rd.ulcer_index.expected_values, .{ .delta = 0.00192 }, "ulcer index (daily {})", .{daily});
    }
}

test "TestMartinRatio test_matches_performance_analytics_output" {
    for (rd.martin_ratio.expected_values_by_rf) |e| {
        const a1 = try runStream(.{ .rf = e.key }, Measures.martinRatio, .{});
        defer ta.free(a1);
        try expectSeries(a1, e.value, .{ .delta = 0.0630 }, "martin ratio (yearly, Rf {d})", .{e.key});
        const a2 = try runStream(.{ .daily = true, .rf = annualize(e.key, .daily) }, Measures.martinRatio, .{});
        defer ta.free(a2);
        try expectSeries(a2, e.value, .{ .delta = 0.0630 }, "martin ratio (daily, Rf {d})", .{e.key});
    }
}

fn checkYearlyDaily(comptime method: anytype, expected: []const f64, comptime label: []const u8) !void {
    inline for (.{ false, true }) |daily| {
        const actual = try runStream(.{ .daily = daily }, method, .{});
        defer ta.free(actual);
        try expectSeries(actual, expected, .{}, label ++ " (daily {})", .{daily});
    }
}

test "TestDrawdownAverage test_matches_performance_analytics_output" {
    try checkYearlyDaily(Measures.drawdownAverage, rd.drawdown_average.expected_values, "drawdown average");
}

test "TestDrawdownAverageLength test_matches_performance_analytics_output" {
    try checkYearlyDaily(Measures.drawdownAverageLength, rd.drawdown_average_length.expected_values, "drawdown average length");
}

test "TestDrawdownAveragePeakToTrough test_matches_performance_analytics_output" {
    try checkYearlyDaily(Measures.drawdownAveragePeakToTrough, rd.drawdown_average_peak_to_trough.expected_values, "drawdown average peak-to-trough");
}

test "TestDrawdownAverageRecovery test_matches_performance_analytics_output" {
    try checkYearlyDaily(Measures.drawdownAverageRecovery, rd.drawdown_average_recovery.expected_values, "drawdown average recovery");
}

test "TestDrawdownDeviation test_matches_performance_analytics_output" {
    try checkYearlyDaily(Measures.drawdownDeviation, rd.drawdown_deviation.expected_values, "drawdown deviation");
}

test "TestCDaRAverage test_matches_performance_analytics_output" {
    // R and this implementation select the continuous drawdown tail
    // differently. TestDocumentedFormulas checks our linear quantile.
    inline for (.{ false, true }) |daily| {
        for (rd.cdar.expected_values_by_p_average_geometric_inverted) |e| {
            const actual = try runStream(.{ .daily = daily }, Measures.cdarAverage, .{e.key});
            defer ta.free(actual);
            try expectSeries(actual, e.value, .{ .delta = 0.02938 }, "CDaR average geometric (daily {}) p {d}", .{ daily, e.key });
        }
    }
}

test "TestCDaRDiscrete test_matches_performance_analytics_output" {
    inline for (.{ false, true }) |daily| {
        for (rd.cdar.expected_values_by_p_discrete_geometric_inverted) |e| {
            const actual = try runStream(.{ .daily = daily }, Measures.cdarDiscrete, .{e.key});
            defer ta.free(actual);
            try expectSeries(actual, e.value, .{}, "CDaR discrete geometric (daily {}) p {d}", .{ daily, e.key });
        }
    }
}

test "TestCDaRBeta test_discrete_tail_selection" {
    var m = try makeMeasures(.{});
    defer m.deinit();
    const pairs = [_][2]f64{ .{ 0.1, 0.1 }, .{ -0.05, -0.1 }, .{ 0.2, 0.2 }, .{ -0.1, -0.2 }, .{ 0.3, 0.3 }, .{ -0.15, -0.3 } };
    for (pairs) |p| try m.addReturn(p[0], p[1]);
    // At 50% confidence, two of three episodes are selected. The
    // denominator is the second-worst depth, -0.2.
    try expectFloat(try m.cdarBeta(0.5), (-0.15 - 0.1) / (2 * -0.2), .{ .places = 14 }, "cdar beta 0.5", .{});
    try expectFloat(try m.cdarBeta(0.8), -0.15 / -0.3, .{ .places = 14 }, "cdar beta 0.8", .{});
}

test "TestCDaRBeta test_matches_performance_analytics_output" {
    inline for (.{ false, true }) |daily| {
        for (rd.cdar_beta.expected_values_by_p_geometric) |e| {
            const actual = try runStream(.{ .daily = daily }, Measures.cdarBeta, .{e.key});
            defer ta.free(actual);
            try expectSeries(actual, e.value, .{ .places = 13 }, "CDaR beta geometric (daily {}) p {d}", .{ daily, e.key });
        }
    }
}

const ones_001 = [_]f64{0.01} ** n_bacon;
const ones_002 = [_]f64{0.02} ** n_bacon;
const zeros = [_]f64{0.0} ** n_bacon;

test "TestCDaRBeta test_mathematical_properties" {
    // With one selected episode, identical portfolio and benchmark returns
    // give the same numerator and denominator.
    {
        var m = try makeMeasures(.{});
        defer m.deinit();
        for ([_]f64{ 0.05, -0.1 }) |r| try m.addReturn(r, r);
        try expectFloat(try m.cdarBeta(0.95), 1.0, .{}, "CDaR beta (one episode) identity", .{});
    }

    var measures = try makeMeasures(.{});
    defer measures.deinit();
    try addBacon(&measures, bp, bp);
    try expectFloat(try measures.cdarBeta(0.95), 1.0, .{ .places = 14 }, "CDaR beta (Bacon) identity", .{});

    // No drawdowns
    {
        var m = try makeMeasures(.{});
        defer m.deinit();
        try addBacon(&m, &ones_001, &ones_002);
        try expectFloat(try m.cdarBeta(0.95), nan, .{}, "CDaR beta (geometric) no drawdowns", .{});
    }

    // Zero returns (portfolio = 0)
    {
        var m = try makeMeasures(.{});
        defer m.deinit();
        try addBacon(&m, &zeros, bb);
        try expectFloat(try m.cdarBeta(0.95), 0.0, .{}, "CDaR beta (geometric) zero returns", .{});
    }

    for ([_]f64{ 0.0, 1.0 }) |confidence| try expectInvalid(measures.cdarBeta(confidence));
}

test "TestCDaRAlpha test_matches_performance_analytics_output" {
    // PerformanceAnalytics hardcodes 12 periods when annualizing the means;
    // this implementation uses periods_per_annum, so use monthly data.
    for (rd.cdar_alpha.expected_values_by_p_geometric) |e| {
        const actual = try runStream(.{ .monthly = true }, Measures.cdarAlpha, .{e.key});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .places = 14 }, "CDaR alpha geometric (monthly) p {d}", .{e.key});
    }
}

test "TestCDaRAlpha test_mathematical_properties" {
    // Identity (portfolio == benchmark).
    {
        var m = try makeMeasures(.{ .monthly = true });
        defer m.deinit();
        try addBacon(&m, bp, bp);
        try expectFloat(try m.cdarAlpha(0.95), 0.0, .{ .places = 14 }, "CDaR alpha (geometric) identity", .{});
    }
    // No drawdowns
    {
        var m = try makeMeasures(.{});
        defer m.deinit();
        try addBacon(&m, &ones_001, &ones_002);
        try expectFloat(try m.cdarAlpha(0.95), nan, .{}, "CDaR alpha (geometric) no drawdowns", .{});
    }
    // Zero returns (portfolio = 0)
    {
        var m = try makeMeasures(.{});
        defer m.deinit();
        try addBacon(&m, &zeros, bb);
        try expectFloat(try m.cdarAlpha(0.95), 0.0, .{}, "CDaR alpha (geometric) zero returns", .{});
    }
}

test "TestRewardToConditionalDrawdown test_definition" {
    for ([_]f64{ 0.8, 0.95 }) |confidence| {
        var m = try makeMeasures(.{});
        defer m.deinit();
        for (0..n_bacon) |i| {
            try m.addReturn(bp[i], bb[i]);
            const dd = try m.drawdownsHighWatermark();
            defer ta.free(dd);
            std.mem.sort(f64, dd, {}, std.sort.asc(f64));
            const k: usize = @max(1, @as(usize, @intFromFloat(fl(dd.len) * (1 - confidence))));
            const tail = dd[0..k];
            const cdar = -pySum(tail) / fl(tail.len);
            const expected = if (cdar != 0) m.geometricMeanReturn() / cdar else nan;
            try expectFloat(try m.rewardToConditionalDrawdown(confidence), expected, .{}, "confidence {d} step {d}", .{ confidence, i });
        }
    }
}

fn checkYearlyAndDailyByRf(entries: []const rd.Entry(f64, []const f64), comptime method: anytype, tol: Tol, comptime label: []const u8) !void {
    try checkByRate(entries, .yearly, false, method, .{}, tol, label);
    try checkByRate(entries, .daily, false, method, .{}, tol, label);
}

test "TestSfmRiskPremium test_matches_performance_analytics_output" {
    try checkYearlyAndDailyByRf(rd.sfm_risk_premium.expected_values_by_rf_perfan, Measures.sfmRiskPremium, .{ .places = 14 }, "SFM risk premium");
}

test "TestSfmAlpha test_matches_performance_analytics_output" {
    try checkYearlyAndDailyByRf(rd.sfm_alpha.expected_values_by_rf_perfan, Measures.sfmAlpha, .{ .places = 14 }, "SFM alpha");
}

test "TestSfmBeta test_matches_performance_analytics_output" {
    try checkYearlyAndDailyByRf(rd.sfm_beta.expected_values_by_rf_perfan, Measures.sfmBeta, .{ .places = 14 }, "SFM beta");
}

test "TestSfmBetaBull test_matches_performance_analytics_output" {
    try checkYearlyAndDailyByRf(rd.sfm_beta_bull.expected_values_by_rf_perfan, Measures.sfmBetaBull, .{ .places = 14 }, "SFM beta bull");
}

test "TestSfmBetaBear test_matches_reference_implementation_output" {
    try checkYearlyAndDailyByRf(rd.sfm_beta_bear.expected_values_by_rf_reference, Measures.sfmBetaBear, .{ .places = 14 }, "SFM beta bear");
}

test "TestTimingRatio test_matches_performance_analytics_output" {
    try checkYearlyAndDailyByRf(rd.timing_ratio.expected_values_by_rf_perfan, Measures.timingRatio, .{ .places = 14 }, "timing ratio");
}

test "TestSfmR2 test_matches_performance_analytics_output" {
    for ([_]Period{ .yearly, .daily }) |p| {
        for (rd.sfm_r2.expected_values_by_rf_perfan) |e| {
            const actual = try runStream(.{ .daily = p == .daily, .rf = annualize(e.key, p) }, Measures.sfmR2, .{});
            defer ta.free(actual);
            try expectSeries(actual, e.value, .{ .places = 14, .skip = if (e.key < 0.05) 15 else 18 }, "SFM R^2 ({s}, Rf {d})", .{ @tagName(p), e.key });
        }
    }
}

test "TestJensenAlpha test_high_daily_risk_free_rate_definition" {
    // The R fixtures lose precision after compounding a 10% or 30% periodic
    // rate over 252 periods. Check the documented formula.
    for ([_]f64{ 0.1, 0.3 }) |rf| {
        const annual_rf = annualize(rf, .daily);
        var m = try Measures.init(ta, 252, annual_rf, 0, 0);
        defer m.deinit();
        for (0..n_bacon) |i| {
            try m.addReturn(bp[i], bb[i]);
            if (i == 0) continue;
            const p_ann = math.pow(f64, prod1p(bp[0 .. i + 1]), 252.0 / fl(i + 1)) - 1;
            const b_ann = math.pow(f64, prod1p(bb[0 .. i + 1]), 252.0 / fl(i + 1)) - 1;
            const beta = m.sfmBeta();
            const expected = p_ann - (beta * b_ann + (1 - beta) * annual_rf);
            try expectFloat(m.jensenAlpha(), expected, .{ .delta = 1e-9, .rel_tol = 1e-12 }, "Jensen alpha rf={d} step={d}", .{ rf, i });
            if (beta != 0) {
                try expectFloat(m.jensenAlphaModified(), expected / beta, .{ .delta = 1e-9, .rel_tol = 1e-12 }, "Jensen alpha modified rf={d} step={d}", .{ rf, i });
            }
        }
    }
}

test "TestJensenAlpha test_matches_performance_analytics_output" {
    for (rd.jensen_alpha.expected_values_by_rf_daily_perfan) |e| {
        // At higher periodic rates the R reference loses precision through
        // cancellation; the formula test above covers them.
        if (e.key > 0.05) continue;
        const actual = try runStream(.{ .daily = true, .rf = annualize(e.key, .daily) }, Measures.jensenAlpha, .{});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{ .delta = 9e-10, .rel_tol = 1e-10 }, "Jensen alpha (daily, Rf {d})", .{e.key});
    }
    try checkByRate(rd.jensen_alpha.expected_values_by_rf_monthly_perfan, .monthly, false, Measures.jensenAlpha, .{}, .{ .places = 12 }, "Jensen alpha");
    try checkByRate(rd.jensen_alpha.expected_values_by_rf_yearly_perfan, .yearly, false, Measures.jensenAlpha, .{}, .{ .places = 14 }, "Jensen alpha");
}

test "TestFamaBeta test_matches_performance_analytics_output" {
    inline for (.{ false, true }) |daily| {
        const actual = try runStream(.{ .daily = daily }, Measures.famaBeta, .{});
        defer ta.free(actual);
        try expectSeries(actual, rd.fama_beta.expected_values_perfan, .{ .places = 14 }, "fama beta (daily {})", .{daily});
    }
}

test "TestModigliani test_matches_performance_analytics_output" {
    try checkYearlyAndDailyByRf(rd.modigliani.expected_values_by_rf_perfan, Measures.modigliani, .{}, "Modigliani-Modigliani");
}

test "TestTrackingError test_matches_performance_analytics_output" {
    const cases = .{
        .{ Period.daily, rd.tracking_error.expected_values_daily_perfan },
        .{ Period.monthly, rd.tracking_error.expected_values_monthly_perfan },
        .{ Period.yearly, rd.tracking_error.expected_values_annual_perfan },
    };
    inline for (cases) |c| {
        const actual = try runStream(.{ .daily = c[0] == .daily, .monthly = c[0] == .monthly }, Measures.trackingError, .{});
        defer ta.free(actual);
        try expectSeries(actual, c[1], .{}, "tracking error ({s})", .{@tagName(c[0])});
    }
}

test "TestActivePremium test_matches_performance_analytics_output" {
    // Skip the first element: PerformanceAnalytics uses it to determine
    // periodicity.
    const cases = .{
        .{ Period.daily, rd.active_premium.expected_values_daily_perfan, 11 },
        .{ Period.monthly, rd.active_premium.expected_values_monthly_perfan, 14 },
        .{ Period.yearly, rd.active_premium.expected_values_annual_perfan, 15 },
    };
    inline for (cases) |c| {
        const actual = try runStream(.{ .daily = c[0] == .daily, .monthly = c[0] == .monthly }, Measures.activePremium, .{});
        defer ta.free(actual);
        try expectSeries(actual, c[1], .{ .places = c[2], .skip = 1 }, "active premium ({s})", .{@tagName(c[0])});
    }
}

test "TestInformationRatio test_matches_performance_analytics_output" {
    const cases = .{
        .{ Period.daily, rd.information_ratio.expected_values_daily_perfan, 10 },
        .{ Period.monthly, rd.information_ratio.expected_values_monthly_perfan, 12 },
        .{ Period.yearly, rd.information_ratio.expected_values_annual_perfan, 13 },
    };
    inline for (cases) |c| {
        const actual = try runStream(.{ .daily = c[0] == .daily, .monthly = c[0] == .monthly }, Measures.informationRatio, .{});
        defer ta.free(actual);
        try expectSeries(actual, c[1], .{ .places = c[2], .skip = 2 }, "information ratio ({s})", .{@tagName(c[0])});
    }
}

test "TestInformationRatioModified test_sign_rule" {
    var m = try makeMeasures(.{ .monthly = true });
    defer m.deinit();
    var diffs: [n_bacon]f64 = undefined;
    for (0..n_bacon) |i| {
        try m.addReturn(bp[i], bb[i]);
        diffs[i] = bp[i] - bb[i];
        const active = fsum(diffs[0 .. i + 1]);
        const ir = m.informationRatio();
        const expected = if (math.isNan(ir)) nan else if (active > 0) ir else -ir;
        try expectFloat(m.informationRatioModified(), expected, .{}, "step {d}", .{i});
    }
}

fn checkDailyMonthlyAnnual(
    daily: []const rd.Entry(f64, []const f64),
    monthly: []const rd.Entry(f64, []const f64),
    annual: []const rd.Entry(f64, []const f64),
    comptime method: anytype,
    tols: [3]Tol,
    comptime label: []const u8,
) !void {
    try checkByRate(daily, .daily, false, method, .{}, tols[0], label);
    try checkByRate(monthly, .monthly, false, method, .{}, tols[1], label);
    try checkByRate(annual, .yearly, false, method, .{}, tols[2], label);
}

test "TestSystematicRisk test_matches_performance_analytics_output" {
    const r = rd.systematic_risk;
    try checkDailyMonthlyAnnual(r.expected_values_by_rf_daily_perfan, r.expected_values_by_rf_monthly_perfan, r.expected_values_by_rf_annual_perfan, Measures.systematicRisk, .{ .{ .places = 14 }, .{ .places = 15 }, .{ .places = 15 } }, "systematic risk");
}

test "TestTreynorRatio test_matches_performance_analytics_output" {
    const r = rd.treynor_ratio;
    try checkDailyMonthlyAnnual(r.expected_values_by_rf_daily_perfan, r.expected_values_by_rf_monthly_perfan, r.expected_values_by_rf_annual_perfan, Measures.treynorRatio, .{ .{ .places = 10 }, .{ .places = 13 }, .{ .places = 14 } }, "treynor ratio");
}

test "TestTreynorRatioModified test_matches_performance_analytics_output" {
    const r = rd.treynor_ratio_modified;
    try checkDailyMonthlyAnnual(r.expected_values_by_rf_daily_perfan, r.expected_values_by_rf_monthly_perfan, r.expected_values_by_rf_annual_perfan, Measures.treynorRatioModified, .{ .{ .places = 10 }, .{ .places = 12 }, .{ .places = 12 } }, "treynor ratio modified");
}

test "TestSpecificRisk test_matches_performance_analytics_output" {
    const r = rd.specific_risk;
    try checkDailyMonthlyAnnual(r.expected_values_by_rf_daily_perfan, r.expected_values_by_rf_monthly_perfan, r.expected_values_by_rf_annual_perfan, Measures.specificRisk, .{ .{ .places = 14 }, .{ .places = 15 }, .{ .places = 15 } }, "specific risk");
}

test "TestTotalRisk test_matches_performance_analytics_output" {
    const r = rd.total_risk;
    try checkDailyMonthlyAnnual(r.expected_values_by_rf_daily_perfan, r.expected_values_by_rf_monthly_perfan, r.expected_values_by_rf_annual_perfan, Measures.totalRisk, .{ .{ .places = 14 }, .{ .places = 14 }, .{ .places = 15 } }, "total risk");
}

fn checkRateSeries(e: rd.Entry(f64, []const f64), period: Period, comptime method: anytype, tol: Tol, comptime label: []const u8) !void {
    const actual = try runStream(.{ .daily = period == .daily, .monthly = period == .monthly, .rf = annualize(e.key, period) }, method, .{});
    defer ta.free(actual);
    try expectSeries(actual, e.value, tol, label ++ " ({s}, Rf {d})", .{ @tagName(period), e.key });
}

test "TestAppraisalRatio test_matches_performance_analytics_output" {
    for (rd.appraisal_ratio.expected_values_by_rf_daily_perfan) |e| {
        if (e.key < 0.1) try checkRateSeries(e, .daily, Measures.appraisalRatio, .{ .delta = if (e.key < 0.05) 1e-8 else 1e-4, .skip = 2 }, "appraisal ratio");
    }
    for (rd.appraisal_ratio.expected_values_by_rf_monthly_perfan) |e| {
        try checkRateSeries(e, .monthly, Measures.appraisalRatio, .{ .delta = if (e.key < 0.05) 1e-11 else 1e-9, .skip = 2 }, "appraisal ratio");
    }
    for (rd.appraisal_ratio.expected_values_by_rf_annual_perfan) |e| {
        try checkRateSeries(e, .yearly, Measures.appraisalRatio, .{ .delta = if (e.key < 0.05) 1e-13 else 1e-11, .skip = 2 }, "appraisal ratio");
    }
}

test "TestJensenAlphaModified test_definition" {
    var m = try makeMeasures(.{});
    defer m.deinit();
    for (0..n_bacon) |i| {
        try m.addReturn(bp[i], bb[i]);
        const beta = m.sfmBeta();
        const expected = if (beta != 0) m.jensenAlpha() / beta else nan;
        try expectFloat(m.jensenAlphaModified(), expected, .{ .places = 14 }, "Jensen alpha modified n={d}", .{i + 1});
    }
}

test "TestJensenAlphaModified test_matches_performance_analytics_output" {
    for (rd.jensen_alpha_modified.expected_values_by_rf_daily_perfan) |e| {
        // Annualizing 24 daily observations amplifies the reference's
        // floating-point error, even when the formula agrees.
        if (e.key < 0.05) try checkRateSeries(e, .daily, Measures.jensenAlphaModified, .{ .delta = 1e-8 }, "Jensen alpha modified");
    }
    for (rd.jensen_alpha_modified.expected_values_by_rf_monthly_perfan) |e| {
        try checkRateSeries(e, .monthly, Measures.jensenAlphaModified, .{ .delta = 1e-10 }, "Jensen alpha modified");
    }
    for (rd.jensen_alpha_modified.expected_values_by_rf_annual_perfan) |e| {
        try checkRateSeries(e, .yearly, Measures.jensenAlphaModified, .{ .delta = 1e-13 }, "Jensen alpha modified");
    }
}

test "TestJensenAlphaAlternative test_matches_performance_analytics_output" {
    for (rd.jensen_alpha_alternative.expected_values_by_rf_daily_perfan) |e| {
        if (e.key < 0.3) try checkRateSeries(e, .daily, Measures.jensenAlphaAlternative, .{ .delta = if (e.key < 0.1) 0.1826 else 0.707 }, "Jensen alpha alternative");
    }
    for (rd.jensen_alpha_alternative.expected_values_by_rf_monthly_perfan) |e| {
        try checkRateSeries(e, .monthly, Measures.jensenAlphaAlternative, .{ .places = if (e.key < 0.3) 10 else 9 }, "Jensen alpha alternative");
    }
    for (rd.jensen_alpha_alternative.expected_values_by_rf_annual_perfan) |e| {
        try checkRateSeries(e, .yearly, Measures.jensenAlphaAlternative, .{ .places = 12 }, "Jensen alpha alternative");
    }
}

test "TestMSquared test_matches_performance_analytics_output" {
    for (rd.m_squared.expected_values_by_rf_daily_perfan) |e| {
        if (e.key < 0.05) try checkRateSeries(e, .daily, Measures.mSquared, .{ .delta = if (e.key < 0.01) 0.1849 else 0.82956 }, "M squared");
    }
    for (rd.m_squared.expected_values_by_rf_monthly_perfan) |e| {
        try checkRateSeries(e, .monthly, Measures.mSquared, .{ .delta = if (e.key < 0.05) 0.00861 else 1.621 }, "M squared");
    }
    for (rd.m_squared.expected_values_by_rf_annual_perfan) |e| {
        try checkRateSeries(e, .yearly, Measures.mSquared, .{ .places = 14 }, "M squared");
    }
}

test "TestMSquaredExcess test_matches_performance_analytics_output" {
    for (rd.m_squared_excess.expected_values_by_rf_daily_perfan) |e| {
        if (e.key < 0.05) try checkRateSeries(e, .daily, Measures.mSquaredExcess, .{ .delta = if (e.key < 0.01) 0.02244 else 0.101 }, "M squared excess");
    }
    for (rd.m_squared_excess.expected_values_by_rf_monthly_perfan) |e| {
        try checkRateSeries(e, .monthly, Measures.mSquaredExcess, .{ .delta = if (e.key < 0.05) 0.007782 else 1.466 }, "M squared excess");
    }
    for (rd.m_squared_excess.expected_values_by_rf_annual_perfan) |e| {
        try checkRateSeries(e, .yearly, Measures.mSquaredExcess, .{}, "M squared excess");
    }
}

test "TestMSquaredSortino test_matches_performance_analytics_output" {
    const r = rd.m_squared_sortino;
    try checkByRate(r.expected_values_by_mar_daily_perfan, .daily, true, Measures.mSquaredSortino, .{}, .{ .places = 11, .skip = 3 }, "M squared Sortino");
    try checkByRate(r.expected_values_by_mar_monthly_perfan, .monthly, true, Measures.mSquaredSortino, .{}, .{ .places = 14, .skip = 3 }, "M squared Sortino");
    try checkByRate(r.expected_values_by_mar_annual_perfan, .yearly, true, Measures.mSquaredSortino, .{}, .{ .places = 15, .skip = 3 }, "M squared Sortino");
}

test "TestTailRatio test_matches_reference_implementation_output" {
    for (rd.tail_ratio.expected_values_by_cutoff_reference) |e| {
        const actual = try runStream(.{}, Measures.tailRatio, .{e.key});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{}, "tail ratio (yearly, cutoff {d})", .{e.key});
    }
    var m = try makeMeasures(.{});
    defer m.deinit();
    try expectInvalid(m.tailRatio(0.5));
    try expectInvalid(m.tailRatio(1.0));
}

test "TestKellyRatio test_matches_performance_analytics_output" {
    inline for (.{ Period.daily, Period.monthly, Period.yearly }) |p| {
        try checkByRate(rd.kelly_ratio.expected_values_by_rf_perfan, p, false, Measures.kellyRatio, .{}, .{ .places = 11 }, "Kelly ratio");
    }
}

test "TestKellyRatioFull test_matches_performance_analytics_output" {
    inline for (.{ Period.daily, Period.monthly, Period.yearly }) |p| {
        try checkByRate(rd.kelly_ratio.expected_values_by_rf_full_perfan, p, false, Measures.kellyRatioFull, .{}, .{ .places = 11 }, "Kelly ratio full");
    }
}

test "TestHurstExponent test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.hurstExponent, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.hurst_exponent.expected_values_perfan, .{ .places = 14 }, "Hurst exponent", .{});
}

test "TestBiasRatio test_matches_reference_implementation_output" {
    for (rd.bias_ratio.expected_values_by_mult_reference) |e| {
        const actual = try runStream(.{}, Measures.biasRatio, .{e.key});
        defer ta.free(actual);
        try expectSeries(actual, e.value, .{}, "bias ratio (yearly, std_dev_multiplier {d})", .{e.key});
    }
    var m = try makeMeasures(.{});
    defer m.deinit();
    try expectInvalid(m.biasRatio(0.0));
}

test "TestKRatio test_matches_reference_implementation_output" {
    const actual = try runStream(.{}, Measures.kRatio, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.k_ratio.expected_values_reference, .{ .places = 14 }, "K-ratio", .{});
}

test "TestGainToPainRatio test_matches_reference_implementation_output" {
    const actual = try runStream(.{}, Measures.gainToPainRatio, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.gain_to_pain_ratio.expected_values_reference, .{}, "Gain-to-pain ratio", .{});
}

fn checkByGeometric(entries: []const rd.Entry(bool, []const f64), comptime method: anytype, tol: Tol, comptime label: []const u8) !void {
    for (entries) |e| {
        const actual = try runStream(.{}, method, .{e.key});
        defer ta.free(actual);
        try expectSeries(actual, e.value, tol, label ++ " (yearly, geometric {})", .{e.key});
    }
}

test "TestUpsideCaptureRatio test_matches_performance_analytics_output" {
    try checkByGeometric(rd.upside_capture_ratio.expected_values_by_geometric_perfan, Measures.upsideCaptureRatio, .{ .places = 13, .skip = 1 }, "Upside capture ratio");
}

test "TestDownsideCaptureRatio test_matches_performance_analytics_output" {
    try checkByGeometric(rd.downside_capture_ratio.expected_values_by_geometric_perfan, Measures.downsideCaptureRatio, .{ .places = 14 }, "Downside capture ratio");
}

test "TestOverallCaptureRatio test_matches_reference_implementation_output" {
    try checkByGeometric(rd.overall_capture_ratio.expected_values_by_geometric_reference, Measures.overallCaptureRatio, .{ .places = 13 }, "Overall capture ratio");
}

test "TestUpNumberRatio test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.upNumberRatio, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.up_number_ratio.expected_values_perfan, .{}, "Up number ratio (yearly)", .{});
}

test "TestDownNumberRatio test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.downNumberRatio, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.down_number_ratio.expected_values_perfan, .{}, "Down number ratio (yearly)", .{});
}

test "TestUpPercentageRatio test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.upPercentageRatio, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.up_percentage_ratio.expected_values_perfan, .{}, "Up percentage ratio (yearly)", .{});
}

test "TestDownPercentageRatio test_matches_performance_analytics_output" {
    const actual = try runStream(.{}, Measures.downPercentageRatio, .{});
    defer ta.free(actual);
    try expectSeries(actual, rd.down_percentage_ratio.expected_values_perfan, .{}, "Down percentage ratio (yearly)", .{});
}

test "TestDocumentedFormulas test_drawdown_risk_and_ratios" {
    for ([_]f64{ 0.0, 0.05 }) |rf| {
        var m = try makeMeasures(.{ .rf = rf });
        defer m.deinit();
        for (0..n_bacon) |i| {
            try m.addReturn(bp[i], bb[i]);
            const drawdowns = try m.drawdownsHighWatermark();
            defer ta.free(drawdowns);
            const n = fl(i + 1);
            const pain = -fsum(drawdowns) / n;
            var sq: [n_bacon]f64 = undefined;
            for (drawdowns, 0..) |x, j| sq[j] = x * x;
            const ulcer = @sqrt(fsum(sq[0..drawdowns.len]) / n);
            const geometric_return = math.pow(f64, prod1p(bp[0 .. i + 1]), 1 / n) - 1;
            try expectFloat(m.painIndex(), pain, .{ .places = 14 }, "pain index rf {d} step {d}", .{ rf, i });
            try expectFloat(m.ulcerIndex(), ulcer, .{ .places = 14 }, "ulcer index rf {d} step {d}", .{ rf, i });
            if (pain > 0) try expectFloat(m.painRatio(), (geometric_return - rf) / pain, .{ .places = 12 }, "pain ratio rf {d} step {d}", .{ rf, i });
            if (ulcer > 0) try expectFloat(m.martinRatio(), (geometric_return - rf) / ulcer, .{ .places = 12 }, "martin ratio rf {d} step {d}", .{ rf, i });
        }
    }
}

test "TestDocumentedFormulas test_cdar_average_and_alpha" {
    var m = try makeMeasures(.{ .monthly = true });
    defer m.deinit();
    for (0..n_bacon) |step| {
        try m.addReturn(bp[step], bb[step]);
        const drawdowns = try m.drawdownsHighWatermark();
        defer ta.free(drawdowns);
        std.mem.sort(f64, drawdowns, {}, std.sort.asc(f64));
        const len = drawdowns.len;
        for ([_]f64{ 0.9, 0.95 }) |confidence| {
            const position = (1 - confidence) * fl(len - 1);
            const lo: usize = @intFromFloat(position);
            const q = drawdowns[lo] + (position - fl(lo)) * (drawdowns[@min(lo + 1, len - 1)] - drawdowns[lo]);
            var tail: [n_bacon]f64 = undefined;
            var k: usize = 0;
            for (drawdowns) |d| {
                if (d <= q) {
                    tail[k] = d;
                    k += 1;
                }
            }
            const expected_cdar = if (q < 0) -fsum(tail[0..k]) / fl(k) else 0.0;
            try expectFloat(try m.cdarAverage(confidence), expected_cdar, .{ .places = 14 }, "cdar average step {d} conf {d}", .{ step, confidence });
            const beta = try m.cdarBeta(confidence);
            if (math.isFinite(beta)) {
                const portfolio_mean = fmean(bp[0..len]);
                const benchmark_mean = fmean(bb[0..len]);
                const expected_alpha = math.pow(f64, 1 + portfolio_mean, 12) - 1 - beta * (math.pow(f64, 1 + benchmark_mean, 12) - 1);
                try expectFloat(try m.cdarAlpha(confidence), expected_alpha, .{ .places = 13 }, "cdar alpha step {d} conf {d}", .{ step, confidence });
            }
        }
    }
}

test "TestDocumentedFormulas test_m_squared_and_jensen_alpha_alternative" {
    const cases = [_][2]f64{ .{ 1, 0.05 }, .{ 12, 0.05 }, .{ 252, 0.01 } };
    for (cases) |c| {
        const periods = c[0];
        const annual_rf = math.pow(f64, 1 + c[1], periods) - 1;
        var m = try Measures.init(ta, periods, annual_rf, 0, 0);
        defer m.deinit();
        for (0..n_bacon) |i| {
            try m.addReturn(bp[i], bb[i]);
            const n = i + 1;
            if (n < 2) continue;
            const portfolio = bp[0..n];
            const benchmark = bb[0..n];
            const p_ann = math.pow(f64, prod1p(portfolio), periods / fl(n)) - 1;
            const b_ann = math.pow(f64, prod1p(benchmark), periods / fl(n)) - 1;
            const scale = pstdev(benchmark) / pstdev(portfolio);
            const expected_m2 = p_ann * scale + annual_rf * (1 - scale);
            const tol: Tol = .{ .delta = 1e-10, .rel_tol = 1e-11 };
            try expectFloat(m.mSquared(), expected_m2, tol, "M squared periods={d} n={d}", .{ periods, n });
            const expected_excess = (1 + expected_m2) / (1 + b_ann) - 1;
            try expectFloat(m.mSquaredExcess(), expected_excess, tol, "M squared excess periods={d} n={d}", .{ periods, n });
            const systematic_risk = @abs(m.sfmBeta()) * stdev(benchmark) * @sqrt(periods);
            if (systematic_risk > 0 and math.isFinite(systematic_risk)) {
                try expectFloat(m.jensenAlphaAlternative(), m.jensenAlpha() / systematic_risk, tol, "Jensen alpha alternative periods={d} n={d}", .{ periods, n });
            }
        }
    }
}

test "TestDocumentedFormulas test_m_squared_equal_volatility_with_extreme_rate" {
    var m = try Measures.init(ta, 252, math.pow(f64, 1.3, 252) - 1, 0, 0);
    defer m.deinit();
    try m.addReturn(0.125, 0.25);
    try m.addReturn(0.375, 0.5);
    // Binary-exact inputs give exactly equal portfolio and benchmark
    // volatility, so the annual risk-free terms must cancel.
    const expected = math.pow(f64, 1.125 * 1.375, 126) - 1;
    try expectFloat(m.mSquared(), expected, .{ .rel_tol = 1e-14 }, "M squared extreme rate", .{});
}

test "TestRollingWindow test_rolling_matches_fresh" {
    // At every step, including while the window is still filling, every
    // public measure of a rolling-window instance equals that of a fresh
    // instance fed only the returns in the window.
    var prng = std.Random.DefaultPrng.init(42);
    const rnd = prng.random();
    var random_returns: [150]f64 = undefined;
    var random_benchmark: [150]f64 = undefined;
    for (&random_returns) |*r| r.* = rnd.floatNorm(f64) * 0.03 + 0.002;
    for (&random_benchmark) |*r| r.* = rnd.floatNorm(f64) * 0.025 + 0.001;

    const Config = struct { window: usize, ppa: f64, rf: f64, mar: f64, returns: []const f64, bench: []const f64 };
    const configs = [_]Config{
        .{ .window = 10, .ppa = 1, .rf = 0.0, .mar = 0.0, .returns = bp, .bench = bb },
        .{ .window = 30, .ppa = 12, .rf = 0.05, .mar = 0.03, .returns = &random_returns, .bench = &random_benchmark },
    };
    for (configs) |cfg| {
        var rolling = try Measures.init(ta, cfg.ppa, cfg.rf, cfg.mar, cfg.window);
        defer rolling.deinit();
        for (0..cfg.returns.len) |i| {
            try rolling.addReturn(cfg.returns[i], cfg.bench[i]);
            var fresh = try Measures.init(ta, cfg.ppa, cfg.rf, cfg.mar, 0);
            defer fresh.deinit();
            const from = if (i + 1 > cfg.window) i + 1 - cfg.window else 0;
            for (from..i + 1) |j| try fresh.addReturn(cfg.returns[j], cfg.bench[j]);

            const actual = try t.evaluateAll(&rolling);
            defer t.freeValues(&actual);
            const expected = try t.evaluateAll(&fresh);
            defer t.freeValues(&expected);
            inline for (t.public_measures, 0..) |pm, k| {
                const a = actual[k];
                const e = expected[k];
                switch (e) {
                    .list => |el| try expectSeries(a.list, el, .{ .places = 12 }, "window {d} step {d} " ++ pm[0], .{ cfg.window, i }),
                    .b => |eb| try testing.expectEqual(eb, a.b),
                    .f => |ef| {
                        const tol: Tol = if (math.isFinite(ef)) .{ .delta = 1e-12 * @max(1.0, @abs(ef)) } else .{};
                        try expectFloat(a.f, ef, tol, "window {d} step {d} " ++ pm[0], .{ cfg.window, i });
                    },
                }
            }
        }
    }
}

test "Measures init rejects non-positive periods_per_annum" {
    try expectInvalid(Measures.init(ta, 0, 0, 0, 0));
    try expectInvalid(Measures.init(ta, -12, 0, 0, 0));
}

test "VaR/ES confidence outside the helper domain maps to NaN" {
    // Not in Python: where Python raises ValueError from the core helpers,
    // the Zig port returns NaN; at the domain edges it returns Python's value.
    var m = try makeMeasures(.{});
    defer m.deinit();
    try addBacon(&m, bp, &[_]f64{0.0} ** n_bacon);
    try expectFloat(try m.varHistorical(1.0), 0.065, .{}, "var historical 1.0", .{});
    try expectFloat(try m.varHistorical(0.0), -0.081, .{}, "var historical 0.0", .{});
    try expectFloat(try m.esHistorical(1.0), 0.065, .{}, "es historical 1.0", .{});
    try expectFloat(try m.sharpeRatioVarHistorical(1.0), 0.13846153846153847, .{}, "sharpe var historical 1.0", .{});
    try testing.expect(math.isNan(try m.varHistorical(1.5)));
    try testing.expect(math.isNan(try m.esHistorical(1.5)));
    try testing.expect(math.isNan(try m.rewardToVarRatioHistorical(math.nan(f64))));
    try testing.expect(math.isNan(m.varGaussian(1.0)));
    try testing.expect(math.isNan(m.varGaussian(1.5)));
    try testing.expect(math.isNan(m.esCornishFisher(1.5)));
    try testing.expect(math.isNan(m.sharpeRatioEsGaussian(0.0)));
}
