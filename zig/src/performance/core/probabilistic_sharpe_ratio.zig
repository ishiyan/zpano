//! Probabilistic Sharpe Ratio (Bailey and López de Prado).

const std = @import("std");
const math = std.math;
const testing = std.testing;
const RawMomentsKleinKBN = @import("raw_moments_klein_kbn").RawMomentsKleinKBN;

const norm = @import("norm.zig");

/// Probabilistic Sharpe Ratio: the probability that the true Sharpe ratio
/// exceeds `reference_sr`, given the observed Sharpe ratio `sr`:
///
///     PSR = Φ((sr − sr*)·√(n − 1) / √(1 − sr·γ₃ + sr²·(γ₄ − 1)/4))
///
/// γ₃ is the population skewness (0 when `zero_skewness`), γ₄ the
/// population (Pearson) kurtosis (3 when `normal_kurtosis`).
/// Python defaults: reference_sr = 0.0, zero_skewness = false,
/// normal_kurtosis = true.
///
/// NaN when `sr`, a required moment, or the result is unavailable, or the
/// denominator is zero. Where Python's math.sqrt would raise ValueError
/// (negative radicand, or n = 0) this returns NaN instead.
pub fn probabilisticSharpeRatio(returns_kbn: *const RawMomentsKleinKBN, sr: f64, reference_sr: f64, zero_skewness: bool, normal_kurtosis: bool) f64 {
    if (math.isNan(sr)) return math.nan(f64);
    var skewness: f64 = undefined;
    if (zero_skewness) {
        skewness = 0;
    } else {
        skewness = returns_kbn.skewnessMoment(); // or _sample
        if (math.isNan(skewness)) return math.nan(f64);
    }
    var kurtosis: f64 = undefined;
    if (normal_kurtosis) {
        kurtosis = 3; // excess kurtosis = 0, so K = 3
    } else {
        kurtosis = returns_kbn.kurtosisExcess();
        if (math.isNan(kurtosis)) return math.nan(f64);
        kurtosis += 3; // convert to regular kurtosis
    }

    const denom = @sqrt(1 - sr * skewness + (sr * sr) * (kurtosis - 1) / 4);
    if (denom == 0) return math.nan(f64);

    const n: f64 = @floatFromInt(returns_kbn.n());
    const z = (sr - reference_sr) * @sqrt(n - 1) / denom;
    return norm.normCdf(z);
}

// ── Tests ──────────────────────────────────────────────────────────────────

fn expectAlmostEqual(expected: f64, actual: f64, places: u5) !void {
    if (expected == actual) return;
    const tol = 0.5 * math.pow(f64, 10.0, -@as(f64, @floatFromInt(places)));
    if (!(@abs(expected - actual) <= tol)) {
        std.debug.print("expected {d}, got {d} (places {d})\n", .{ expected, actual, places });
        return error.TestExpectedApproxEq;
    }
}

test "four moment assumptions" {
    const values = [_]f64{ -0.03, -0.01, 0.02, 0.04, 0.07 };
    var moments = RawMomentsKleinKBN{};
    for (values) |v| moments.update(v);
    const len: f64 = @floatFromInt(values.len);
    var sum: f64 = 0;
    for (values) |x| sum += x;
    const mean = sum / len;
    var s2: f64 = 0;
    var s3: f64 = 0;
    var s4: f64 = 0;
    for (values) |x| {
        const dev = x - mean;
        s2 += dev * dev;
        s3 += dev * dev * dev;
        s4 += dev * dev * dev * dev;
    }
    const mu2 = s2 / len;
    const skew = s3 / len / math.pow(f64, mu2, 1.5);
    const kurt = s4 / len / (mu2 * mu2);
    const sr = 0.75;
    const reference_sr = 0.1;

    for ([_]bool{ false, true }) |zero_skewness| {
        for ([_]bool{ false, true }) |normal_kurtosis| {
            const s: f64 = if (zero_skewness) 0.0 else skew;
            const k: f64 = if (normal_kurtosis) 3.0 else kurt;
            const denominator = @sqrt(1 - sr * s + sr * sr * (k - 1) / 4);
            const z = (sr - reference_sr) * @sqrt(len - 1) / denominator;
            const expected = 0.5 * (1 + norm.erf(z / @sqrt(2.0)));
            const actual = probabilisticSharpeRatio(&moments, sr, reference_sr, zero_skewness, normal_kurtosis);
            try expectAlmostEqual(expected, actual, 13);
        }
    }
}

test "unavailable sharpe or moments" {
    var moments = RawMomentsKleinKBN{};
    moments.update(0.01);
    try testing.expect(math.isNan(probabilisticSharpeRatio(&moments, math.nan(f64), 0.0, false, true)));
    try testing.expect(math.isNan(probabilisticSharpeRatio(&moments, 0.5, 0.0, false, true)));
    try testing.expect(math.isNan(probabilisticSharpeRatio(&moments, 0.5, 0.0, true, false)));
}
