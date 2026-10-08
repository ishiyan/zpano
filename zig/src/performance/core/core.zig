//! Core building blocks for the streaming performance measures
//! (mirrors Python `performance/core/__init__.py`):
//!
//! - SFMRegression: single factor model alpha, beta, bull/bear beta, r².
//! - percentile (NumPy "linear"), percentileInPlace.
//! - normCdf, normPdf, normPpf (Acklam), erf.
//! - varHistorical, varGaussian, varCornishFisher.
//! - esHistorical, esGaussian, esCornishFisher.
//! - probabilisticSharpeRatio.
//! - CumulativeReturn: streaming geometric returns.
//! - Capture: upside/downside capture, number and percentage ratios.
//! - WinLoss: winning/losing return sums, means and counts.
//! - ContinuousDrawdownRuns (+ ddPercent): Burke-type continuous drawdowns.
//! - HighWaterMarkDrawdown: rolling high-water-mark drawdowns.
//! - DrawdownEpisode, DrawdownEpisodes: drawdown episode tracker.
//! - PartialMoments: partial moments about a threshold.
//! - RawPartialMoments: raw partial moment sums about zero.

const std = @import("std");

pub const sfm_regression = @import("sfm_regression.zig");
pub const percentile_mod = @import("percentile.zig");
pub const norm = @import("norm.zig");
pub const var_mod = @import("var.zig");
pub const es = @import("es.zig");
pub const probabilistic_sharpe_ratio = @import("probabilistic_sharpe_ratio.zig");
pub const cumulative_return = @import("cumulative_return.zig");
pub const capture = @import("capture.zig");
pub const win_loss = @import("win_loss.zig");
pub const continuous_drawdown_runs = @import("continuous_drawdown_runs.zig");
pub const high_watermark_drawdown = @import("high_watermark_drawdown.zig");
pub const drawdown_episodes = @import("drawdown_episodes.zig");
pub const partial_moments = @import("partial_moments.zig");
pub const partial_moments_raw = @import("partial_moments_raw.zig");

pub const SFMRegression = sfm_regression.SFMRegression;
pub const percentile = percentile_mod.percentile;
pub const percentileInPlace = percentile_mod.percentileInPlace;
pub const normCdf = norm.normCdf;
pub const normPdf = norm.normPdf;
pub const normPpf = norm.normPpf;
pub const erf = norm.erf;
pub const varHistorical = var_mod.varHistorical;
pub const varGaussian = var_mod.varGaussian;
pub const varCornishFisher = var_mod.varCornishFisher;
pub const esHistorical = es.esHistorical;
pub const esGaussian = es.esGaussian;
pub const esCornishFisher = es.esCornishFisher;
pub const probabilisticSharpeRatio = probabilistic_sharpe_ratio.probabilisticSharpeRatio;
pub const CumulativeReturn = cumulative_return.CumulativeReturn;
pub const Capture = capture.Capture;
pub const WinLoss = win_loss.WinLoss;
pub const ContinuousDrawdownRuns = continuous_drawdown_runs.ContinuousDrawdownRuns;
pub const ddPercent = continuous_drawdown_runs.ddPercent;
pub const HighWaterMarkDrawdown = high_watermark_drawdown.HighWaterMarkDrawdown;
pub const DrawdownEpisode = drawdown_episodes.DrawdownEpisode;
pub const DrawdownEpisodes = drawdown_episodes.DrawdownEpisodes;
pub const PartialMoments = partial_moments.PartialMoments;
pub const RawPartialMoments = partial_moments_raw.RawPartialMoments;

test {
    std.testing.refAllDecls(@This());
    _ = @import("fifo_buffer.zig");
    _ = @import("sfm_regression.zig");
    _ = @import("percentile.zig");
    _ = @import("norm.zig");
    _ = @import("var.zig");
    _ = @import("es.zig");
    _ = @import("probabilistic_sharpe_ratio.zig");
    _ = @import("cumulative_return.zig");
    _ = @import("capture.zig");
    _ = @import("win_loss.zig");
    _ = @import("continuous_drawdown_runs.zig");
    _ = @import("high_watermark_drawdown.zig");
    _ = @import("drawdown_episodes.zig");
    _ = @import("partial_moments.zig");
    _ = @import("partial_moments_raw.zig");
}
