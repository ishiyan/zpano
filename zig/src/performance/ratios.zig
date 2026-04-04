const std = @import("std");
const math = std.math;
const conventions = @import("conventions");
const fractional = @import("fractional");
const pmod = @import("periodicity");

const DayCountConvention = conventions.DayCountConvention;
const DateTime = fractional.DateTime;
const Periodicity = pmod.Periodicity;

const Allocator = std.mem.Allocator;
const ArrayList = std.ArrayList;

const sqrt2: f64 = 1.4142135623730950488016887242097;

/// Ratios accumulates portfolio returns incrementally and computes
/// various financial performance ratios at each step.
pub const Ratios = struct {
    periodicity_val: Periodicity,
    periods_per_annum: u16,
    days_per_period: f64,
    risk_free_rate: f64,
    required_return: f64,
    day_count_convention: DayCountConvention,
    rolling_window: ?u32,
    min_periods: ?u32,

    fractional_periods: ArrayList(f64),
    returns: ArrayList(f64),
    sample_count: usize,

    logret_sum: f64,
    drawdowns_cumulative: ArrayList(f64),
    drawdowns_cumulative_min: f64,
    drawdowns_peaks: ArrayList(f64),
    drawdowns_peaks_peak: usize,
    drawdown_continuous: ArrayList(f64),
    drawdown_continuous_final: ArrayList(f64),
    drawdown_continuous_finalized: bool,
    drawdown_continuous_peak: usize,
    drawdown_continuous_inside: bool,
    cumulative_return_plus1: f64,
    cumulative_return_plus1_max: f64,
    cumulative_return_geometric_mean: ?f64,
    returns_mean: ?f64,
    returns_std: ?f64,
    returns_autocorr_penalty: f64,
    excess_mean: ?f64,
    excess_std: ?f64,
    excess_autocorr_penalty: f64,
    required_mean: ?f64,
    required_lpm1: ?f64,
    required_lpm2: ?f64,
    required_lpm3: ?f64,
    required_hpm1: ?f64,
    required_hpm2: ?f64,
    required_hpm3: ?f64,
    required_autocorr_penalty: f64,
    avg_return: ?f64,
    avg_win: ?f64,
    avg_loss: ?f64,
    win_rate: ?f64,
    total_duration: f64,

    reset_called: bool,

    allocator: Allocator,

    /// Creates a new Ratios instance with the specified parameters.
    /// Annual rates are de-annualized to per-period rates based on the periodicity.
    /// rolling_window, if non-null, limits computations to the last N returns.
    /// min_periods, if non-null and > 0, causes all ratio methods to return null
    /// until at least that many samples have been added.
    pub fn init(
        allocator: Allocator,
        p: Periodicity,
        annual_risk_free_rate: f64,
        annual_target_return: f64,
        day_count_convention: DayCountConvention,
        rolling_window: ?u32,
        min_periods: ?u32,
    ) Ratios {
        const ppa = p.periodsPerAnnum();
        const dpp = p.daysPerPeriod();

        var rfr = annual_risk_free_rate;
        if (annual_risk_free_rate != 0 and ppa != 1) {
            rfr = math.pow(f64, 1.0 + annual_risk_free_rate, 1.0 / @as(f64, @floatFromInt(ppa))) - 1.0;
        }

        var rr = annual_target_return;
        if (annual_target_return != 0 and ppa != 1) {
            rr = math.pow(f64, 1.0 + annual_target_return, 1.0 / @as(f64, @floatFromInt(ppa))) - 1.0;
        }

        // Treat null or <=0 min_periods as no minimum
        const mp: ?u32 = if (min_periods != null and min_periods.? > 0) min_periods else null;

        var r = Ratios{
            .periodicity_val = p,
            .periods_per_annum = ppa,
            .days_per_period = dpp,
            .risk_free_rate = rfr,
            .required_return = rr,
            .day_count_convention = day_count_convention,
            .rolling_window = rolling_window,
            .min_periods = mp,
            .fractional_periods = ArrayList(f64).empty,
            .returns = ArrayList(f64).empty,
            .sample_count = 0,
            .logret_sum = 0,
            .drawdowns_cumulative = ArrayList(f64).empty,
            .drawdowns_cumulative_min = math.inf(f64),
            .drawdowns_peaks = ArrayList(f64).empty,
            .drawdowns_peaks_peak = 0,
            .drawdown_continuous = ArrayList(f64).empty,
            .drawdown_continuous_final = ArrayList(f64).empty,
            .drawdown_continuous_finalized = false,
            .drawdown_continuous_peak = 1,
            .drawdown_continuous_inside = false,
            .cumulative_return_plus1 = 1,
            .cumulative_return_plus1_max = -math.inf(f64),
            .total_duration = 0,
            .cumulative_return_geometric_mean = null,
            .returns_mean = null,
            .returns_std = null,
            .returns_autocorr_penalty = 1,
            .excess_mean = null,
            .excess_std = null,
            .excess_autocorr_penalty = 1,
            .required_mean = null,
            .required_lpm1 = null,
            .required_lpm2 = null,
            .required_lpm3 = null,
            .required_hpm1 = null,
            .required_hpm2 = null,
            .required_hpm3 = null,
            .required_autocorr_penalty = 1,
            .avg_return = null,
            .avg_win = null,
            .avg_loss = null,
            .win_rate = null,
            .reset_called = false,
            .allocator = allocator,
        };
        _ = &r;
        return r;
    }

    /// Frees all allocated memory.
    pub fn deinit(self: *Ratios) void {
        self.fractional_periods.deinit(self.allocator);
        self.returns.deinit(self.allocator);
        self.drawdowns_cumulative.deinit(self.allocator);
        self.drawdowns_peaks.deinit(self.allocator);
        self.drawdown_continuous.deinit(self.allocator);
        self.drawdown_continuous_final.deinit(self.allocator);
    }

    /// Resets all internal state for accumulation.
    pub fn reset(self: *Ratios) void {
        self.fractional_periods.clearRetainingCapacity();
        self.returns.clearRetainingCapacity();
        self.sample_count = 0;
        self.logret_sum = 0;
        self.drawdowns_cumulative.clearRetainingCapacity();
        self.drawdowns_cumulative_min = math.inf(f64);
        self.drawdowns_peaks.clearRetainingCapacity();
        self.drawdowns_peaks_peak = 0;
        self.drawdown_continuous.clearRetainingCapacity();
        self.drawdown_continuous_final.clearRetainingCapacity();
        self.drawdown_continuous_finalized = false;
        self.drawdown_continuous_peak = 1;
        self.drawdown_continuous_inside = false;
        self.cumulative_return_plus1 = 1;
        self.cumulative_return_plus1_max = -math.inf(f64);
        self.total_duration = 0;
        self.cumulative_return_geometric_mean = null;
        self.returns_mean = null;
        self.returns_std = null;
        self.returns_autocorr_penalty = 1;
        self.excess_mean = null;
        self.excess_std = null;
        self.excess_autocorr_penalty = 1;
        self.required_mean = null;
        self.required_lpm1 = null;
        self.required_lpm2 = null;
        self.required_lpm3 = null;
        self.required_hpm1 = null;
        self.required_hpm2 = null;
        self.required_hpm3 = null;
        self.required_autocorr_penalty = 1;
        self.avg_return = null;
        self.avg_win = null;
        self.avg_loss = null;
        self.win_rate = null;
        self.reset_called = true;
    }

    /// Adds a new return observation and updates all internal state.
    pub fn addReturn(
        self: *Ratios,
        return_val: f64,
        return_benchmark: f64,
        value: f64,
        time_start: DateTime,
        time_end: DateTime,
    ) !void {
        _ = return_benchmark;
        _ = value;

        var fractional_period: f64 = undefined;
        if (self.periodicity_val == .annual) {
            fractional_period = fractional.yearFrac(time_start, time_end, self.day_count_convention) catch return;
        } else {
            const fp_val = fractional.dayFrac(time_start, time_end, self.day_count_convention) catch return;
            fractional_period = fp_val / self.days_per_period;
        }

        try self.fractional_periods.append(self.allocator, fractional_period);
        if (fractional_period == 0) {
            return;
        }
        self.total_duration += fractional_period;
        self.sample_count += 1;

        // Normalized return
        const ret = return_val / fractional_period;
        try self.returns.append(self.allocator, ret);

        // Window slice: use last rolling_window returns, or all if not set
        const all = self.returns.items;
        const w: []const f64 = if (self.rolling_window) |rw| blk: {
            const n: usize = @intCast(rw);
            break :blk if (all.len > n) all[all.len - n ..] else all;
        } else all;
        const l = w.len;
        const lf: f64 = @floatFromInt(l);

        // Returns mean
        const mean = sliceMean(w);
        self.returns_mean = mean;

        // Returns std (ddof=1, sample)
        if (l > 1) {
            self.returns_std = sliceStdDdof1(w, mean);
        } else {
            self.returns_std = null;
        }

        self.returns_autocorr_penalty = autocorrPenalty(w);

        // Average return, win rate, avg win, avg loss
        const non_zero = countNonZero(w);
        if (non_zero > 0) {
            self.avg_return = meanNonZero(w);

            const pos_count = countPositive(w);
            self.win_rate = @as(f64, @floatFromInt(pos_count)) / @as(f64, @floatFromInt(non_zero));

            if (pos_count > 0) {
                self.avg_win = meanPositive(w);
            } else {
                self.avg_win = null;
            }

            const neg_count = countNegative(w);
            if (neg_count > 0) {
                self.avg_loss = meanNegative(w);
            } else {
                self.avg_loss = null;
            }
        } else {
            self.avg_return = null;
            self.win_rate = null;
            self.avg_win = null;
            self.avg_loss = null;
        }

        // Excess returns (returns less risk-free rate)
        if (self.risk_free_rate == 0) {
            self.excess_mean = self.returns_mean;
            self.excess_std = self.returns_std;
            self.excess_autocorr_penalty = self.returns_autocorr_penalty;
        } else {
            var excess_sum: f64 = 0;
            for (w) |v| {
                excess_sum += v - self.risk_free_rate;
            }
            const em = excess_sum / lf;
            self.excess_mean = em;
            if (l > 1) {
                var sumsq: f64 = 0;
                for (w) |v| {
                    const d = (v - self.risk_free_rate) - em;
                    sumsq += d * d;
                }
                self.excess_std = @sqrt(sumsq / @as(f64, @floatFromInt(l - 1)));
            } else {
                self.excess_std = null;
            }
            self.excess_autocorr_penalty = 1; // stub
        }

        // Lower partial moments for the raw returns (less required return)
        var lpm1_sum: f64 = 0;
        var lpm2_sum: f64 = 0;
        var lpm3_sum: f64 = 0;
        for (w) |v| {
            var diff: f64 = undefined;
            if (self.required_return == 0) {
                diff = -v;
            } else {
                diff = self.required_return - v;
            }
            if (diff < 0) diff = 0;
            lpm1_sum += diff;
            lpm2_sum += math.pow(f64, diff, 2);
            lpm3_sum += math.pow(f64, diff, 3);
        }
        self.required_lpm1 = lpm1_sum / lf;
        self.required_lpm2 = lpm2_sum / lf;
        self.required_lpm3 = lpm3_sum / lf;

        // Higher partial moments for the raw returns (less required return)
        if (self.required_return == 0) {
            self.required_mean = self.returns_mean;
            self.required_autocorr_penalty = self.returns_autocorr_penalty;
        } else {
            var rm_sum: f64 = 0;
            for (w) |v| {
                rm_sum += v - self.required_return;
            }
            self.required_mean = rm_sum / lf;
            self.required_autocorr_penalty = 1; // stub
        }

        var hpm1_sum: f64 = 0;
        var hpm2_sum: f64 = 0;
        var hpm3_sum: f64 = 0;
        for (w) |v| {
            var diff: f64 = undefined;
            if (self.required_return == 0) {
                diff = v;
            } else {
                diff = v - self.required_return;
            }
            if (diff < 0) diff = 0;
            hpm1_sum += diff;
            hpm2_sum += math.pow(f64, diff, 2);
            hpm3_sum += math.pow(f64, diff, 3);
        }
        self.required_hpm1 = hpm1_sum / lf;
        self.required_hpm2 = hpm2_sum / lf;
        self.required_hpm3 = hpm3_sum / lf;

        // Cumulative returns — recompute from window
        const w_start = all.len - l;
        var logret_sum_val: f64 = 0;
        for (0..l) |j| {
            const fp_j = self.fractional_periods.items[w_start + j];
            if (fp_j != 0) {
                logret_sum_val += @log(w[j] + 1.0);
            }
        }
        self.logret_sum = logret_sum_val;
        const cmr = @as(f64, @exp(logret_sum_val));
        self.cumulative_return_plus1 = cmr;
        if (l >= 1) {
            self.cumulative_return_geometric_mean = math.pow(f64, cmr, 1.0 / lf) - 1.0;
        }
        self.cumulative_return_plus1_max = -math.inf(f64);
        var cumr: f64 = 1.0;
        for (0..l) |j| {
            cumr *= (w[j] + 1.0);
            if (cumr > self.cumulative_return_plus1_max) {
                self.cumulative_return_plus1_max = cumr;
            }
        }

        // Drawdowns from peaks to valleys (cumulative returns) — recompute from window
        self.drawdowns_cumulative.clearRetainingCapacity();
        self.drawdowns_cumulative_min = math.inf(f64);
        cumr = 1.0;
        var cumr_max: f64 = -math.inf(f64);
        for (0..l) |j| {
            cumr *= (w[j] + 1.0);
            if (cumr > cumr_max) {
                cumr_max = cumr;
            }
            const dd = cumr / cumr_max - 1.0;
            try self.drawdowns_cumulative.append(self.allocator, dd);
            if (self.drawdowns_cumulative_min > dd) {
                self.drawdowns_cumulative_min = dd;
            }
        }

        // Drawdown peaks (used in pain index, ulcer index) — recompute from window
        self.drawdowns_peaks.clearRetainingCapacity();
        self.drawdowns_peaks_peak = 0;
        for (0..l) |j| {
            var dd_peak: f64 = 1.0;
            var k: usize = self.drawdowns_peaks_peak + 1;
            while (k <= j) : (k += 1) {
                dd_peak *= (1.0 + w[k] * 0.01);
            }
            if (dd_peak > 1.0) {
                self.drawdowns_peaks_peak = j;
                try self.drawdowns_peaks.append(self.allocator, 0);
            } else {
                try self.drawdowns_peaks.append(self.allocator, (dd_peak - 1.0) * 100.0);
            }
        }

        // Drawdown continuous (used in Burke ratio) — recompute from window
        self.drawdown_continuous.clearRetainingCapacity();
        self.drawdown_continuous_final.clearRetainingCapacity();
        self.drawdown_continuous_finalized = false;
        self.drawdown_continuous_peak = 1;
        self.drawdown_continuous_inside = false;
        for (1..l) |j| {
            if (w[j] < 0) {
                if (!self.drawdown_continuous_inside) {
                    self.drawdown_continuous_inside = true;
                    self.drawdown_continuous_peak = j - 1;
                }
                try self.drawdown_continuous.append(self.allocator, 0);
            } else {
                if (self.drawdown_continuous_inside) {
                    var dd_c: f64 = 1.0;
                    const j1 = self.drawdown_continuous_peak + 1;
                    var k: usize = j1;
                    while (k < j) : (k += 1) {
                        dd_c *= (1.0 + w[k] * 0.01);
                    }
                    try self.drawdown_continuous.append(self.allocator, (dd_c - 1.0) * 100.0);
                    self.drawdown_continuous_inside = false;
                } else {
                    try self.drawdown_continuous.append(self.allocator, 0);
                }
            }
        }
    }

    /// Returns the cumulative geometric return.
    pub fn cumulativeReturn(self: *const Ratios) f64 {
        return self.cumulative_return_plus1 - 1.0;
    }

    /// Returns the drawdowns from peaks to valleys on cumulative geometric returns.
    pub fn drawdownsCumulative(self: *const Ratios) []const f64 {
        return self.drawdowns_cumulative.items;
    }

    /// Returns the minimum (most negative) cumulative drawdown.
    pub fn minDrawdownsCumulative(self: *const Ratios) f64 {
        return self.drawdowns_cumulative_min;
    }

    /// Returns the absolute value of the worst cumulative drawdown.
    pub fn worstDrawdownsCumulative(self: *const Ratios) f64 {
        return @abs(self.drawdowns_cumulative_min);
    }

    /// Returns the drawdowns from peaks (used in pain/ulcer indices).
    pub fn drawdownsPeaks(self: *const Ratios) []const f64 {
        return self.drawdowns_peaks.items;
    }

    /// Returns drawdowns on continuous uninterrupted losing regions.
    pub fn drawdownsContinuous(self: *Ratios, peaks_only: bool, max_peaks: usize, allocator: Allocator) ![]f64 {
        self.finalizeContinuousDrawdown();
        if (!peaks_only) {
            const result = try allocator.alloc(f64, self.drawdown_continuous_final.items.len);
            @memcpy(result, self.drawdown_continuous_final.items);
            return result;
        }
        // Filter non-zero
        var drawdowns = ArrayList(f64).empty;
        defer drawdowns.deinit(allocator);
        for (self.drawdown_continuous_final.items) |v| {
            if (v != 0) {
                try drawdowns.append(allocator, v);
            }
        }
        if (max_peaks > 0 and drawdowns.items.len > 0) {
            std.mem.sort(f64, drawdowns.items, {}, std.sort.asc(f64));
            const count = @min(drawdowns.items.len, max_peaks);
            const result = try allocator.alloc(f64, count);
            @memcpy(result, drawdowns.items[0..count]);
            return result;
        }
        const result = try allocator.alloc(f64, drawdowns.items.len);
        @memcpy(result, drawdowns.items);
        return result;
    }

    fn finalizeContinuousDrawdown(self: *Ratios) void {
        if (self.drawdown_continuous_finalized) {
            return;
        }
        const w = self.windowReturns();
        self.drawdown_continuous_final.clearRetainingCapacity();
        // Copy existing drawdown_continuous into final
        self.drawdown_continuous_final.appendSlice(self.allocator, self.drawdown_continuous.items) catch return;

        if (self.drawdown_continuous_inside) {
            var dd_c: f64 = 1.0;
            const j1 = self.drawdown_continuous_peak + 1;
            var j: usize = j1;
            while (j < w.len) : (j += 1) {
                dd_c *= (1.0 + w[j] * 0.01);
            }
            self.drawdown_continuous_final.append(self.allocator, (dd_c - 1.0) * 100.0) catch return;
        } else {
            self.drawdown_continuous_final.append(self.allocator, 0) catch return;
        }
        self.drawdown_continuous_finalized = true;
    }

    /// Returns true if enough samples have been added to satisfy min_periods.
    fn isPrimed(self: *const Ratios) bool {
        if (self.min_periods) |mp| {
            return self.sample_count >= mp;
        }
        return true;
    }

    /// Returns the windowed slice of returns.
    fn windowReturns(self: *const Ratios) []const f64 {
        const all = self.returns.items;
        if (self.rolling_window) |rw| {
            const n: usize = @intCast(rw);
            if (all.len > n) return all[all.len - n ..];
        }
        return all;
    }

    /// Returns the population skewness of the returns.
    pub fn skew(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const w = self.windowReturns();
        if (w.len < 2) {
            return null;
        }
        return populationSkewness(w);
    }

    /// Returns the population excess kurtosis of the returns.
    pub fn kurtosis(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const w = self.windowReturns();
        if (w.len < 2) {
            return null;
        }
        return populationExcessKurtosis(w);
    }

    /// Calculates the ex-post Sharpe ratio.
    pub fn sharpeRatio(self: *const Ratios, ignore_risk_free_rate: bool, autocorrelation_penalty: bool) ?f64 {
        if (!self.isPrimed()) return null;
        if (ignore_risk_free_rate) {
            const rm = self.returns_mean orelse return null;
            const rs = self.returns_std orelse return null;
            if (rs == 0) return null;
            var denom = rs;
            if (autocorrelation_penalty) {
                denom *= self.returns_autocorr_penalty;
            }
            return rm / denom;
        }
        const em = self.excess_mean orelse return null;
        const es = self.excess_std orelse return null;
        if (es == 0) return null;
        var denom = es;
        if (autocorrelation_penalty) {
            denom *= self.excess_autocorr_penalty;
        }
        return em / denom;
    }

    /// Calculates the Sortino ratio.
    pub fn sortinoRatio(self: *const Ratios, autocorrelation_penalty: bool, divide_by_sqrt2: bool) ?f64 {
        if (!self.isPrimed()) return null;
        const rm = self.required_mean orelse return null;
        const lpm2 = self.required_lpm2 orelse return null;
        if (lpm2 == 0) return null;
        var denom = @sqrt(lpm2);
        if (autocorrelation_penalty) {
            denom *= self.required_autocorr_penalty;
        }
        if (divide_by_sqrt2) {
            denom *= sqrt2;
        }
        return rm / denom;
    }

    /// Calculates the Omega ratio.
    pub fn omegaRatio(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const rm = self.required_mean orelse return null;
        const lpm1 = self.required_lpm1 orelse return null;
        if (lpm1 == 0) return null;
        return rm / lpm1 + 1.0;
    }

    /// Calculates the Kappa ratio of a given order.
    pub fn kappaRatio(self: *const Ratios, order: u32) ?f64 {
        if (!self.isPrimed()) return null;
        const rm = self.required_mean orelse return null;
        switch (order) {
            1 => {
                const lpm1 = self.required_lpm1 orelse return null;
                if (lpm1 == 0) return null;
                return rm / lpm1;
            },
            2 => {
                const lpm2 = self.required_lpm2 orelse return null;
                if (lpm2 == 0) return null;
                return rm / @sqrt(lpm2);
            },
            3 => {
                const lpm3 = self.required_lpm3 orelse return null;
                if (lpm3 == 0) return null;
                return rm / std.math.cbrt(lpm3);
            },
            else => {
                const w = self.windowReturns();
                const l = w.len;
                if (l == 0) return null;
                const lf: f64 = @floatFromInt(l);
                var lpm_sum: f64 = 0;
                for (w) |v| {
                    var diff: f64 = undefined;
                    if (self.required_return == 0) {
                        diff = -v;
                    } else {
                        diff = self.required_return - v;
                    }
                    if (diff < 0) diff = 0;
                    lpm_sum += math.pow(f64, diff, @as(f64, @floatFromInt(order)));
                }
                const lpm = lpm_sum / lf;
                if (lpm == 0) return null;
                return rm / math.pow(f64, lpm, 1.0 / @as(f64, @floatFromInt(order)));
            },
        }
    }

    /// Calculates the Kappa ratio of order 3.
    pub fn kappa3Ratio(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const rm = self.required_mean orelse return null;
        const lpm3 = self.required_lpm3 orelse return null;
        if (lpm3 == 0) return null;
        return rm / std.math.cbrt(lpm3);
    }

    /// Calculates the Bernardo-Ledoit ratio.
    pub fn bernardoLedoitRatio(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const w = self.windowReturns();
        const l = w.len;
        if (l < 1) return null;
        const lf: f64 = @floatFromInt(l);

        // LPM1 with threshold=0
        var lpm1_sum: f64 = 0;
        for (w) |v| {
            const neg = -v;
            if (neg > 0) lpm1_sum += neg;
        }
        const lpm1 = lpm1_sum / lf;
        if (lpm1 == 0) return null;

        // HPM1 with threshold=0
        var hpm1_sum: f64 = 0;
        for (w) |v| {
            if (v > 0) hpm1_sum += v;
        }
        const hpm1 = hpm1_sum / lf;
        return hpm1 / lpm1;
    }

    /// Calculates the upside potential ratio.
    pub fn upsidePotentialRatio(self: *const Ratios, full: bool) ?f64 {
        if (!self.isPrimed()) return null;
        if (full) {
            const hpm1 = self.required_hpm1 orelse return null;
            const lpm2 = self.required_lpm2 orelse return null;
            if (lpm2 == 0) return null;
            return hpm1 / @sqrt(lpm2);
        }
        // Subset version
        const w = self.windowReturns();
        var below_count: usize = 0;
        var below_lpm2_sum: f64 = 0;
        for (w) |v| {
            if (v < self.required_return) {
                const diff = v - self.required_return;
                below_lpm2_sum += diff * diff;
                below_count += 1;
            }
        }
        if (below_count < 1) return null;
        const lpm2 = below_lpm2_sum / @as(f64, @floatFromInt(below_count));
        if (lpm2 == 0) return null;

        var above_sum: f64 = 0;
        var above_count: usize = 0;
        for (w) |v| {
            if (v > self.required_return) {
                above_sum += v - self.required_return;
                above_count += 1;
            }
        }
        if (above_count == 0) return null;
        const hpm1 = above_sum / @as(f64, @floatFromInt(above_count));
        return hpm1 / @sqrt(lpm2);
    }

    /// Returns the compound (annual) growth rate (CAGR).
    pub fn compoundGrowthRate(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        return self.cumulative_return_geometric_mean;
    }

    /// Calculates the Calmar ratio.
    pub fn calmarRatio(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const wdd = self.worstDrawdownsCumulative();
        if (wdd == 0) return null;
        const gm = self.cumulative_return_geometric_mean orelse return null;
        return gm / wdd;
    }

    /// Calculates the Sterling ratio with the given annual excess rate.
    pub fn sterlingRatio(self: *const Ratios, annual_excess_rate: f64) ?f64 {
        if (!self.isPrimed()) return null;
        var excess_rate = annual_excess_rate;
        if (annual_excess_rate != 0 and self.periods_per_annum != 1) {
            excess_rate = math.pow(f64, 1.0 + annual_excess_rate, 1.0 / @as(f64, @floatFromInt(self.periods_per_annum))) - 1.0;
        }
        const wdd = self.worstDrawdownsCumulative() + excess_rate;
        if (wdd == 0) return null;
        const gm = self.cumulative_return_geometric_mean orelse return null;
        return gm / wdd;
    }

    /// Calculates the Burke ratio.
    pub fn burkeRatio(self: *Ratios, modified: bool) ?f64 {
        if (!self.isPrimed()) return null;
        const gm = self.cumulative_return_geometric_mean orelse return null;
        const rate = gm - self.risk_free_rate;

        self.finalizeContinuousDrawdown();
        const all_drawdowns = self.drawdown_continuous_final.items;

        // Filter to non-zero peaks only (matching Go's DrawdownsContinuous(true, 0))
        var non_zero_count: usize = 0;
        var sum_sq: f64 = 0;
        for (all_drawdowns) |d| {
            if (d != 0) {
                sum_sq += d * d;
                non_zero_count += 1;
            }
        }
        if (non_zero_count < 1) return null;
        const sqrt_sum_sq = @sqrt(sum_sq);
        if (sqrt_sum_sq == 0) return null;

        var burke = rate / sqrt_sum_sq;
        if (modified) {
            burke *= @sqrt(@as(f64, @floatFromInt(self.windowReturns().len)));
        }
        return burke;
    }

    /// Helper to get continuous drawdown peaks without allocation.
    fn drawdownsContinuousPeaksOnly(self: *Ratios) ?[]const f64 {
        self.finalizeContinuousDrawdown();
        return self.drawdown_continuous_final.items;
    }

    /// Calculates the pain index.
    pub fn painIndex(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const l = self.drawdowns_peaks.items.len;
        if (l < 1) return null;
        return -sliceSum(self.drawdowns_peaks.items) / @as(f64, @floatFromInt(l));
    }

    /// Calculates the pain ratio.
    pub fn painRatio(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const gm = self.cumulative_return_geometric_mean orelse return null;
        const rate = gm - self.risk_free_rate;
        const l = self.drawdowns_peaks.items.len;
        if (l < 1) return null;
        const pi = -sliceSum(self.drawdowns_peaks.items) / @as(f64, @floatFromInt(l));
        if (pi == 0) return null;
        return rate / pi;
    }

    /// Calculates the ulcer index.
    pub fn ulcerIndex(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const l = self.drawdowns_peaks.items.len;
        if (l < 1) return null;
        var sum_sq: f64 = 0;
        for (self.drawdowns_peaks.items) |d| {
            sum_sq += d * d;
        }
        return @sqrt(sum_sq / @as(f64, @floatFromInt(l)));
    }

    /// Calculates the Martin (Ulcer) ratio.
    pub fn martinRatio(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const gm = self.cumulative_return_geometric_mean orelse return null;
        const rate = gm - self.risk_free_rate;
        const l = self.drawdowns_peaks.items.len;
        if (l < 1) return null;
        var sum_sq: f64 = 0;
        for (self.drawdowns_peaks.items) |d| {
            sum_sq += d * d;
        }
        const ui = @sqrt(sum_sq / @as(f64, @floatFromInt(l)));
        if (ui == 0) return null;
        return rate / ui;
    }

    /// Returns Jack Schwager's gain-to-pain ratio.
    pub fn gainToPainRatio(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const lpm1 = self.required_lpm1 orelse return null;
        if (lpm1 == 0) return null;
        const rm = self.returns_mean orelse return null;
        return rm / lpm1;
    }

    /// Calculates the risk of ruin.
    pub fn riskOfRuin(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const wr = self.win_rate orelse return null;
        return math.pow(f64, (1.0 - wr) / (1.0 + wr), @as(f64, @floatFromInt(self.windowReturns().len)));
    }

    /// Calculates the return/risk ratio.
    pub fn riskReturnRatio(self: *const Ratios) ?f64 {
        if (!self.isPrimed()) return null;
        const rm = self.returns_mean orelse return null;
        const rs = self.returns_std orelse return null;
        if (rs == 0) return null;
        return rm / rs;
    }
};

// ---------- helper functions ----------

fn sliceSum(s: []const f64) f64 {
    var sum: f64 = 0;
    for (s) |v| {
        sum += v;
    }
    return sum;
}

fn sliceMean(s: []const f64) f64 {
    if (s.len == 0) return 0;
    return sliceSum(s) / @as(f64, @floatFromInt(s.len));
}

fn sliceStdDdof1(s: []const f64, mean: f64) f64 {
    if (s.len < 2) return 0;
    var sum: f64 = 0;
    for (s) |v| {
        const d = v - mean;
        sum += d * d;
    }
    return @sqrt(sum / @as(f64, @floatFromInt(s.len - 1)));
}

fn countNonZero(s: []const f64) usize {
    var count: usize = 0;
    for (s) |v| {
        if (v != 0) count += 1;
    }
    return count;
}

fn countPositive(s: []const f64) usize {
    var count: usize = 0;
    for (s) |v| {
        if (v > 0) count += 1;
    }
    return count;
}

fn countNegative(s: []const f64) usize {
    var count: usize = 0;
    for (s) |v| {
        if (v < 0) count += 1;
    }
    return count;
}

fn meanNonZero(s: []const f64) f64 {
    var sum: f64 = 0;
    var count: usize = 0;
    for (s) |v| {
        if (v != 0) {
            sum += v;
            count += 1;
        }
    }
    if (count == 0) return 0;
    return sum / @as(f64, @floatFromInt(count));
}

fn meanPositive(s: []const f64) f64 {
    var sum: f64 = 0;
    var count: usize = 0;
    for (s) |v| {
        if (v > 0) {
            sum += v;
            count += 1;
        }
    }
    if (count == 0) return 0;
    return sum / @as(f64, @floatFromInt(count));
}

fn meanNegative(s: []const f64) f64 {
    var sum: f64 = 0;
    var count: usize = 0;
    for (s) |v| {
        if (v < 0) {
            sum += v;
            count += 1;
        }
    }
    if (count == 0) return 0;
    return sum / @as(f64, @floatFromInt(count));
}

fn populationSkewness(s: []const f64) f64 {
    const n: f64 = @floatFromInt(s.len);
    const mean = sliceMean(s);
    var m2: f64 = 0;
    var m3: f64 = 0;
    for (s) |v| {
        const d = v - mean;
        m2 += d * d;
        m3 += d * d * d;
    }
    m2 /= n;
    m3 /= n;
    if (m2 == 0) return 0;
    return m3 / math.pow(f64, m2, 1.5);
}

fn populationExcessKurtosis(s: []const f64) f64 {
    const n: f64 = @floatFromInt(s.len);
    const mean = sliceMean(s);
    var m2: f64 = 0;
    var m4: f64 = 0;
    for (s) |v| {
        const d = v - mean;
        const d2 = d * d;
        m2 += d2;
        m4 += d2 * d2;
    }
    m2 /= n;
    m4 /= n;
    if (m2 == 0) return 0;
    return m4 / (m2 * m2) - 3.0;
}

fn autocorrPenalty(returns: []const f64) f64 {
    _ = returns;
    return 1;
}

// ============== Tests ==============

const testing = std.testing;

const epsilon: f64 = 1e-13;

fn almostEqual(a: f64, b: f64, tol: f64) bool {
    return @abs(a - b) < tol;
}

fn assertNullableFloat(comptime step: []const u8, expected: ?f64, actual: ?f64) !void {
    return assertNullableFloatEps(step, expected, actual, epsilon);
}

fn assertNullableFloatEps(comptime step: []const u8, expected: ?f64, actual: ?f64, eps: f64) !void {
    if (expected == null) {
        if (actual != null) {
            return error.ExpectedNullGotValue;
        }
        return;
    }
    if (actual == null) {
        return error.ExpectedValueGotNull;
    }
    if (!almostEqual(actual.?, expected.?, eps)) {
        std.debug.print("FAIL {s}: expected {d:.16}, got {d:.16}, diff={e}\n", .{ step, expected.?, actual.?, @abs(actual.? - expected.?) });
        return error.ValueMismatch;
    }
}

// Bacon dataset
const bacon_dates_previous = [_]DateTime{
    .{ .year = 2024, .month = 6, .day = 30 }, .{ .year = 2024, .month = 7, .day = 1 },
    .{ .year = 2024, .month = 7, .day = 2 },  .{ .year = 2024, .month = 7, .day = 3 },
    .{ .year = 2024, .month = 7, .day = 4 },  .{ .year = 2024, .month = 7, .day = 5 },
    .{ .year = 2024, .month = 7, .day = 6 },  .{ .year = 2024, .month = 7, .day = 7 },
    .{ .year = 2024, .month = 7, .day = 8 },  .{ .year = 2024, .month = 7, .day = 9 },
    .{ .year = 2024, .month = 7, .day = 10 }, .{ .year = 2024, .month = 7, .day = 11 },
    .{ .year = 2024, .month = 7, .day = 12 }, .{ .year = 2024, .month = 7, .day = 13 },
    .{ .year = 2024, .month = 7, .day = 14 }, .{ .year = 2024, .month = 7, .day = 15 },
    .{ .year = 2024, .month = 7, .day = 16 }, .{ .year = 2024, .month = 7, .day = 17 },
    .{ .year = 2024, .month = 7, .day = 18 }, .{ .year = 2024, .month = 7, .day = 19 },
    .{ .year = 2024, .month = 7, .day = 20 }, .{ .year = 2024, .month = 7, .day = 21 },
    .{ .year = 2024, .month = 7, .day = 22 }, .{ .year = 2024, .month = 7, .day = 23 },
};

const bacon_dates = [_]DateTime{
    .{ .year = 2024, .month = 7, .day = 1 },  .{ .year = 2024, .month = 7, .day = 2 },
    .{ .year = 2024, .month = 7, .day = 3 },  .{ .year = 2024, .month = 7, .day = 4 },
    .{ .year = 2024, .month = 7, .day = 5 },  .{ .year = 2024, .month = 7, .day = 6 },
    .{ .year = 2024, .month = 7, .day = 7 },  .{ .year = 2024, .month = 7, .day = 8 },
    .{ .year = 2024, .month = 7, .day = 9 },  .{ .year = 2024, .month = 7, .day = 10 },
    .{ .year = 2024, .month = 7, .day = 11 }, .{ .year = 2024, .month = 7, .day = 12 },
    .{ .year = 2024, .month = 7, .day = 13 }, .{ .year = 2024, .month = 7, .day = 14 },
    .{ .year = 2024, .month = 7, .day = 15 }, .{ .year = 2024, .month = 7, .day = 16 },
    .{ .year = 2024, .month = 7, .day = 17 }, .{ .year = 2024, .month = 7, .day = 18 },
    .{ .year = 2024, .month = 7, .day = 19 }, .{ .year = 2024, .month = 7, .day = 20 },
    .{ .year = 2024, .month = 7, .day = 21 }, .{ .year = 2024, .month = 7, .day = 22 },
    .{ .year = 2024, .month = 7, .day = 23 }, .{ .year = 2024, .month = 7, .day = 24 },
};

const bacon_portfolio_returns = [_]f64{
    0.003,  0.026,  0.011,  -0.010,
    0.015,  0.025,  0.016,  0.067,
    -0.014, 0.040,  -0.005, 0.081,
    0.040,  -0.037, -0.061, 0.017,
    -0.049, -0.022, 0.070,  0.058,
    -0.065, 0.024,  -0.005, -0.009,
};

const bacon_benchmark_returns = [_]f64{
    0.002,  0.025,  0.018,  -0.011,
    0.014,  0.018,  0.014,  0.065,
    -0.015, 0.042,  -0.006, 0.083,
    0.039,  -0.038, -0.062, 0.015,
    -0.048, 0.021,  0.060,  0.056,
    -0.067, 0.019,  -0.003, 0.000,
};

const bacon_len: usize = bacon_portfolio_returns.len;

fn newRatiosWithRF(allocator: Allocator, rf: f64) Ratios {
    const rf_annual = math.pow(f64, 1.0 + rf, 252.0) - 1.0;
    var r = Ratios.init(allocator, .daily, rf_annual, 0, .raw, null, null);
    r.reset();
    return r;
}

fn newRatiosWithMAR(allocator: Allocator, mar: f64) Ratios {
    const mar_annual = math.pow(f64, 1.0 + mar, 252.0) - 1.0;
    var r = Ratios.init(allocator, .daily, 0, mar_annual, .raw, null, null);
    r.reset();
    return r;
}

fn newRatiosWithRFandMAR(allocator: Allocator, rf: f64, mar: f64) Ratios {
    const rf_annual = math.pow(f64, 1.0 + rf, 252.0) - 1.0;
    const mar_annual = math.pow(f64, 1.0 + mar, 252.0) - 1.0;
    var r = Ratios.init(allocator, .daily, rf_annual, mar_annual, .raw, null, null);
    r.reset();
    return r;
}

fn addBaconReturn(r: *Ratios, i: usize) !void {
    try r.addReturn(
        bacon_portfolio_returns[i],
        bacon_benchmark_returns[i],
        1,
        bacon_dates_previous[i],
        bacon_dates[i],
    );
}

// ---------- Kurtosis Test ----------

test "kurtosis conformance to R PerformanceAnalytics" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,                 -2.00000000000000000, -1.50000000000000000,
        -1.17592035552795000, -0.94669079980875600, -0.96028723389787100,
        -0.57793300076120100, 0.78641242115027200,  0.59954237086621500,
        -0.01187577489273160, 0.07517391430462480,  -0.27406990671095100,
        -0.38022416153835900, -0.31560370425738600, -0.16235155227201600,
        0.02528905226985100,  -0.33285099821964000, -0.37425348407483000,
        -0.58502674157514900, -0.69334606360953100, -0.77381631285861200,
        -0.68208349704651200, -0.61779722177118000, -0.56754620589212500,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.kurtosis();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

// ---------- Sharpe Ratio Tests ----------

test "sharpe ratio rf=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,               0.8915694197569510, 1.1419253390798400,
        0.4977924836999790, 0.6680426571226850, 0.8511810078441020,
        0.9735918376312110, 0.8462916062735410, 0.6475912629068400,
        0.7524743687246650, 0.6702597534059590, 0.7244562693337180,
        0.7945207458232130, 0.5805910371128360, 0.3566360956461000,
        0.3758075293232440, 0.2578994439571370, 0.2131725662300710,
        0.2880753096781920, 0.3448210835211740, 0.2337747541463060,
        0.2546053055676570, 0.2430648040410730, 0.2275684556623890,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.sharpeRatio(false, false);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "sharpe ratio rf=0.05" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0.05);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,                -2.1828078897497800, -3.1402946824695500,
        -2.8208240742998800, -3.0433054380033400, -2.7967375972020500,
        -2.9887005248213900, -1.3662354689514000, -1.4489272141296900,
        -1.3494093427967500, -1.4483773981646000, -0.9801467173338540,
        -0.9561181856516630, -0.9946559628057110, -1.0011155375243300,
        -1.0290804307636500, -1.0706734491553900, -1.1284729554976500,
        -0.9967676208113950, -0.9275814386971810, -0.9577955946576800,
        -0.9630722427994000, -0.9992664166133000, -1.0367007424619900,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.sharpeRatio(false, false);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "sharpe ratio rf=0.10" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0.10);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,               -5.257185199256510, -7.422514704018940,
        -6.139440632299740, -6.754653533129370, -6.444656202248210,
        -6.950992887274000, -3.578762544176350, -3.545445691166220,
        -3.451293054318160, -3.567014549735160, -2.684749704001430,
        -2.706757117126540, -2.569902962724260, -2.358867170694770,
        -2.433968390850540, -2.399246342267910, -2.470118477225370,
        -2.281610551300980, -2.199983960915530, -2.149365943461670,
        -2.180749791166460, -2.241597637267670, -2.300969940586380,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.sharpeRatio(false, false);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "sharpe ratio ignore risk-free rate" {
    const allocator = testing.allocator;

    const expected_rf0 = [_]?f64{
        null,               0.8915694197569510, 1.1419253390798400,
        0.4977924836999790, 0.6680426571226850, 0.8511810078441020,
        0.9735918376312110, 0.8462916062735410, 0.6475912629068400,
        0.7524743687246650, 0.6702597534059590, 0.7244562693337180,
        0.7945207458232130, 0.5805910371128360, 0.3566360956461000,
        0.3758075293232440, 0.2578994439571370, 0.2131725662300710,
        0.2880753096781920, 0.3448210835211740, 0.2337747541463060,
        0.2546053055676570, 0.2430648040410730, 0.2275684556623890,
    };

    for ([_]f64{ 0, 0.05, 0.10 }) |rf| {
        var ratios = newRatiosWithRF(allocator, rf);
        defer ratios.deinit();

        for (0..bacon_len) |i| {
            try addBaconReturn(&ratios, i);
            const actual = ratios.sharpeRatio(true, false);
            if (expected_rf0[i] == null) {
                try testing.expect(actual == null);
            } else {
                try testing.expect(actual != null);
                try testing.expect(almostEqual(actual.?, expected_rf0[i].?, epsilon));
            }
        }
    }
}

// ---------- Sortino Ratio Tests ----------

test "sortino ratio mar=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,              null,              null,
        1.5,               2.01246117974981,  2.85773803324704,
        3.25049446787935,  5.40936687607709,  2.69307029756515,
        3.29008543386979,  2.92819766175444,  4.10863007844407,
        4.56665101160337,  1.67730613630736,  0.691483512929973,
        0.727302390567925, 0.452770753672167, 0.370054264368203,
        0.536498400203865, 0.665303673385798, 0.401733515514418,
        0.438224836666163, 0.418857174247308, 0.392372028795065,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.sortinoRatio(false, false);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "sortino ratio mar=0.05" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0.05);
    defer ratios.deinit();

    const expected = [_]?f64{
        -1,                 -0.951329033501053, -0.967821008377905,
        -0.955961761235827, -0.959422032420532, -0.950640505399932,
        -0.95521850710367,  -0.835987494907806, -0.84620916319764,
        -0.825850705880606, -0.841892559996059, -0.739594446201381,
        -0.729168016460068, -0.735413445987151, -0.731283824494091,
        -0.739823509257131, -0.750430484501361, -0.766429130761335,
        -0.726278292165206, -0.700204514919608, -0.709303305303401,
        -0.71078905810419,  -0.723223919287678, -0.735374254070636,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.sortinoRatio(false, false);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "sortino ratio mar=0.10" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0.10);
    defer ratios.deinit();

    const expected = [_]?f64{
        -1,                 -0.991075392350217, -0.994004065367307,
        -0.990197182430257, -0.991346575643354, -0.990116442284817,
        -0.991246179848335, -0.967496714088971, -0.966414074414246,
        -0.964235550350565, -0.966082469415414, -0.94189872360901,
        -0.942394085487388, -0.936339897881547, -0.925395562084343,
        -0.929178181525619, -0.927078590338157, -0.930569106181964,
        -0.919801251226696, -0.914287704738154, -0.910539461748751,
        -0.912598194566486, -0.916559222172636, -0.920182172840589,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.sortinoRatio(false, false);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "sortino ratio sqrt2 version" {
    const allocator = testing.allocator;

    const expected_mar0 = [_]?f64{
        null,              null,              null,
        1.5,               2.01246117974981,  2.85773803324704,
        3.25049446787935,  5.40936687607709,  2.69307029756515,
        3.29008543386979,  2.92819766175444,  4.10863007844407,
        4.56665101160337,  1.67730613630736,  0.691483512929973,
        0.727302390567925, 0.452770753672167, 0.370054264368203,
        0.536498400203865, 0.665303673385798, 0.401733515514418,
        0.438224836666163, 0.418857174247308, 0.392372028795065,
    };

    for ([_]f64{ 0, 0.05, 0.10 }) |mar| {
        var ratios = newRatiosWithMAR(allocator, mar);
        defer ratios.deinit();

        var expected_for_mar: [24]?f64 = undefined;
        if (mar == 0) {
            expected_for_mar = expected_mar0;
        } else if (mar == 0.05) {
            expected_for_mar = [_]?f64{
                -1,                 -0.951329033501053, -0.967821008377905,
                -0.955961761235827, -0.959422032420532, -0.950640505399932,
                -0.95521850710367,  -0.835987494907806, -0.84620916319764,
                -0.825850705880606, -0.841892559996059, -0.739594446201381,
                -0.729168016460068, -0.735413445987151, -0.731283824494091,
                -0.739823509257131, -0.750430484501361, -0.766429130761335,
                -0.726278292165206, -0.700204514919608, -0.709303305303401,
                -0.71078905810419,  -0.723223919287678, -0.735374254070636,
            };
        } else {
            expected_for_mar = [_]?f64{
                -1,                 -0.991075392350217, -0.994004065367307,
                -0.990197182430257, -0.991346575643354, -0.990116442284817,
                -0.991246179848335, -0.967496714088971, -0.966414074414246,
                -0.964235550350565, -0.966082469415414, -0.94189872360901,
                -0.942394085487388, -0.936339897881547, -0.925395562084343,
                -0.929178181525619, -0.927078590338157, -0.930569106181964,
                -0.919801251226696, -0.914287704738154, -0.910539461748751,
                -0.912598194566486, -0.916559222172636, -0.920182172840589,
            };
        }

        for (0..bacon_len) |i| {
            try addBaconReturn(&ratios, i);
            const actual = ratios.sortinoRatio(false, true);
            if (expected_for_mar[i] == null) {
                try testing.expect(actual == null);
            } else {
                const expected_divided = expected_for_mar[i].? / sqrt2;
                try testing.expect(actual != null);
                try testing.expect(almostEqual(actual.?, expected_divided, epsilon));
            }
        }
    }
}

// ---------- Omega Ratio Tests ----------

test "omega ratio threshold=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,               null,               null,
        4.000000000000000,  5.500000000000000,  8.000000000000000,
        9.600000000000000,  16.300000000000000, 6.791666666666670,
        8.458333333333330,  7.000000000000000,  9.793103448275860,
        11.172413793103400, 4.909090909090910,  2.551181102362210,
        2.685039370078740,  1.937500000000000,  1.722222222222220,
        2.075757575757580,  2.368686868686870,  1.783269961977190,
        1.874524714828900,  1.839552238805970,  1.779783393501810,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.omegaRatio();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "omega ratio threshold=0.02" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0.02);
    defer ratios.deinit();

    const expected = [_]?f64{
        0.00000000000000000, 0.35294117647058800, 0.23076923076923100,
        0.10714285714285700, 0.09836065573770490, 0.18032786885245900,
        0.16923076923076900, 0.89230769230769200, 0.58585858585858600,
        0.78787878787878800, 0.62903225806451600, 1.12096774193548000,
        1.28225806451613000, 0.87845303867403300, 0.60687022900763400,
        0.60000000000000000, 0.47604790419161700, 0.42287234042553200,
        0.55585106382978700, 0.65691489361702100, 0.53579175704989200,
        0.54446854663774400, 0.51646090534979400, 0.48737864077669900,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.omegaRatio();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "omega ratio threshold=0.04" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0.04);
    defer ratios.deinit();

    const expected = [_]?f64{
        0.00000000000000000, 0.00000000000000000, 0.00000000000000000,
        0.00000000000000000, 0.00000000000000000, 0.00000000000000000,
        0.00000000000000000, 0.13917525773195900, 0.10887096774193500,
        0.10887096774193500, 0.09215017064846420, 0.23208191126279900,
        0.23208191126279900, 0.18378378378378400, 0.14437367303609300,
        0.13765182186234800, 0.11663807890223000, 0.10542635658914700,
        0.15193798449612400, 0.17984496124031000, 0.15466666666666700,
        0.15143603133159300, 0.14303329223181300, 0.13488372093023300,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.omegaRatio();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "omega ratio threshold=0.06" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0.06);
    defer ratios.deinit();

    const expected = [_]?f64{
        0.00000000000000000, 0.00000000000000000, 0.00000000000000000,
        0.00000000000000000, 0.00000000000000000, 0.00000000000000000,
        0.00000000000000000, 0.02095808383233530, 0.01715686274509810,
        0.01635514018691590, 0.01419878296146050, 0.05679513184584180,
        0.05458089668615990, 0.04590163934426230, 0.03830369357045140,
        0.03617571059431530, 0.03171007927519820, 0.02901554404145080,
        0.03937823834196890, 0.03929679420889350, 0.03479853479853480,
        0.03368794326241140, 0.03185247275775360, 0.03011093502377180,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.omegaRatio();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

// ---------- Kappa Ratio Tests ----------

test "kappa ratio order=1 mar=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,               null,               null,
        3.0000000000000000, 4.5000000000000000, 7.0000000000000000,
        8.6000000000000000, 15.300000000000000, 5.7916666666666700,
        7.4583333333333300, 6.0000000000000000, 8.7931034482758600,
        10.172413793103400, 3.9090909090909090, 1.5511811023622000,
        1.6850393700787400, 0.9375000000000000, 0.7222222222222220,
        1.0757575757575800, 1.3686868686868700, 0.7832699619771860,
        0.8745247148288970, 0.8395522388059700, 0.7797833935018050,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.kappaRatio(1);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "kappa ratio order=2 mar=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,               null,               null,
        1.5000000000000000, 2.0124611797498100, 2.8577380332470400,
        3.2504944678793500, 5.4093668760770900, 2.6930702975651500,
        3.2900854338697900, 2.9281976617544400, 4.1086300784440700,
        4.5666510116033700, 1.6773061363073600, 0.6914835129299730,
        0.7273023905679250, 0.4527707536721670, 0.3700542643682030,
        0.5364984002038650, 0.6653036733857980, 0.4017335155144180,
        0.4382248366661630, 0.4188571742473080, 0.3923720287950650,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.kappaRatio(2);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "kappa ratio order=3 mar=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,               null,               null,
        1.1905507889761500, 1.5389783520090300, 2.1199740249708300,
        2.3501725959775100, 3.8250000000000000, 2.0689080079822300,
        2.4835586338430600, 2.2408934899599800, 3.0989871337864400,
        3.3988098734763700, 1.1713241279859900, 0.4942094486331960,
        0.5142481946830330, 0.3389522803724070, 0.2803047018509310,
        0.4027354737116720, 0.4951749226471700, 0.3070994714658920,
        0.3324074590706010, 0.3156667962042520, 0.2944582876612480,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.kappaRatio(3);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "kappa3 matches kappa order=3" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,               null,               null,
        1.1905507889761500, 1.5389783520090300, 2.1199740249708300,
        2.3501725959775100, 3.8250000000000000, 2.0689080079822300,
        2.4835586338430600, 2.2408934899599800, 3.0989871337864400,
        3.3988098734763700, 1.1713241279859900, 0.4942094486331960,
        0.5142481946830330, 0.3389522803724070, 0.2803047018509310,
        0.4027354737116720, 0.4951749226471700, 0.3070994714658920,
        0.3324074590706010, 0.3156667962042520, 0.2944582876612480,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.kappa3Ratio();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "kappa ratio order=4 mar=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,               null,               null,
        1.0606601717798200, 1.3458139030991000, 1.8259320100855000,
        1.9983654900858500, 3.2164287883454600, 1.8033735333115700,
        2.1458818396425000, 1.9358196995813000, 2.6577517731212100,
        2.8955073548113600, 0.9572325404178820, 0.4101535241803780,
        0.4244948756831930, 0.2893122764499430, 0.2395667425326850,
        0.3426567344926290, 0.4195093677307130, 0.2646839611869220,
        0.2853879958557920, 0.2700286247384100, 0.2510732850424660,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.kappaRatio(4);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "kappa ratio order=1 mar=0.05" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0.05);
    defer ratios.deinit();

    const expected = [_]?f64{
        -1.0000000000000000, -1.0000000000000000, -1.0000000000000000,
        -1.0000000000000000, -1.0000000000000000, -1.0000000000000000,
        -1.0000000000000000, -0.9356060606060610, -0.9481707317073170,
        -0.9497041420118340, -0.9567430025445290, -0.8778625954198470,
        -0.8808933002481390, -0.9020408163265300, -0.9201331114808650,
        -0.9242902208201890, -0.9345156889495230, -0.9403726708074530,
        -0.9155279503105590, -0.9055900621118010, -0.9173913043478260,
        -0.9196617336152220, -0.9240759240759240, -0.9283018867924530,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.kappaRatio(1);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

// ---------- Bernardo-Ledoit Ratio Test ----------

test "bernardo-ledoit ratio" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,              null,              null,
        4.000000000000000, 5.500000000000000, 8.000000000000000,
        9.600000000000000, 16.30000000000000, 6.791666666666670,
        8.458333333333330, 7.000000000000000, 9.793103448275860,
        11.17241379310340, 4.909090909090910, 2.551181102362200,
        2.685039370078740, 1.937500000000000, 1.722222222222220,
        2.075757575757580, 2.368686868686870, 1.783269961977190,
        1.874524714828900, 1.839552238805970, 1.779783393501800,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.bernardoLedoitRatio();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

// ---------- Upside Potential Ratio Test ----------

test "upside potential ratio full=true mar=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithMAR(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,               null,               null,
        2.0000000000000000, 2.4596747752497700, 3.2659863237109000,
        3.6284589408885800, 5.7629202666703600, 3.1580608525404200,
        3.7312142071260700, 3.4162306053801800, 4.5758860481494800,
        5.0155760263033600, 2.1063844502464600, 1.1372622243112200,
        1.1589257718862700, 0.9357262242558120, 0.8824370919549460,
        1.0352152229285800, 1.1513927041252400, 0.9146263047391370,
        0.9393254107670360, 0.9177626084618780, 0.8955528249813290,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.upsidePotentialRatio(true);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

// ---------- Cumulative Return Test ----------

test "cumulative return" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const expected = [_]f64{
        0.00299999999999989, 0.02907799999999990, 0.04039785799999970,
        0.02999387941999990, 0.04544378761129960, 0.07157988230158210,
        0.08872516041840740, 0.16166974616644100, 0.14540636972011000,
        0.19122262450891500, 0.18526651138637000, 0.28127309880866600,
        0.33252402276101300, 0.28322063391885500, 0.20494417524980500,
        0.22542822622905200, 0.16538224314382800, 0.13974383379466400,
        0.21952590216029100, 0.29025840448558800, 0.20639160819402400,
        0.23534500679068100, 0.22916828175672800, 0.21810576722091700,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.cumulativeReturn();
        try testing.expect(almostEqual(actual, expected[i], epsilon));
    }
}

// ---------- Drawdowns Cumulative Test ----------

test "drawdowns cumulative" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const expected_drawdowns = [_]f64{
        0.000000000000000000,  0.000000000000000000,  0.000000000000000000,
        -0.009999999999999900, 0.000000000000000000,  0.000000000000000000,
        0.000000000000000000,  0.000000000000000000,  -0.014000000000000000,
        0.000000000000000000,  -0.005000000000000120, 0.000000000000000000,
        0.000000000000000000,  -0.037000000000000100, -0.095743000000000000,
        -0.080370631000000200, -0.125432470081000000, -0.144672955739218000,
        -0.084800062640963400, -0.031718466274139200, -0.094656765966320100,
        -0.072928528349511800, -0.077563885707764200, -0.085865810736394400,
    };
    const expected_worst: f64 = 0.1446729557392180;

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
    }

    const actual = ratios.drawdownsCumulative();
    for (expected_drawdowns, 0..) |exp, i| {
        try testing.expect(almostEqual(actual[i], exp, epsilon));
    }
    try testing.expect(almostEqual(ratios.worstDrawdownsCumulative(), expected_worst, epsilon));
}

// ---------- Calmar Ratio Test ----------

test "calmar ratio" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const calmar_eps: f64 = 1e-12;

    const expected = [_]?f64{
        null,                null,                null,
        0.74155751780918500, 0.89279126631479400, 1.15889854414036000,
        1.22179559465027000, 1.89088510302246000, 1.08562360529801000,
        1.26085762243604000, 1.11225700971196000, 1.49066405029967000,
        1.59487944362032000, 0.48572827216522300, 0.13062513296618000,
        0.13355239428276700, 0.07209886266479390, 0.05041253535660620,
        0.07257832783270360, 0.08863890501902830, 0.06203631318696950,
        0.06672377010548700, 0.06228923867560830, 0.05705690600200920,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.calmarRatio();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, calmar_eps));
        }
    }
}

// ---------- Sterling Ratio Tests ----------

test "sterling ratio excess=0" {
    const allocator = testing.allocator;
    var ratios = Ratios.init(allocator, .daily, 0, 0, .raw, null, null);
    ratios.reset();
    defer ratios.deinit();

    const sterling_eps: f64 = 1e-12;

    const expected = [_]?f64{
        null,                null,                null,
        0.74155751780918500, 0.89279126631479400, 1.15889854414036000,
        1.22179559465027000, 1.89088510302246000, 1.08562360529801000,
        1.26085762243604000, 1.11225700971196000, 1.49066405029967000,
        1.59487944362032000, 0.48572827216522300, 0.13062513296618000,
        0.13355239428276700, 0.07209886266479390, 0.05041253535660620,
        0.07257832783270360, 0.08863890501902830, 0.06203631318696950,
        0.06672377010548700, 0.06228923867560830, 0.05705690600200920,
    };

    const excess_annual = math.pow(f64, 1.0 + 0.0, 252.0) - 1.0;
    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.sterlingRatio(excess_annual);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, sterling_eps));
        }
    }
}

test "sterling ratio excess=0.02" {
    const allocator = testing.allocator;
    var ratios = Ratios.init(allocator, .daily, 0, 0, .raw, null, null);
    ratios.reset();
    defer ratios.deinit();

    const sterling_eps: f64 = 1e-12;

    const expected = [_]?f64{
        0.14999999999999500, 0.72174090072224500, 0.66442920035313400,
        0.24718583926972700, 0.29759708877159600, 0.38629951471345200,
        0.40726519821675300, 0.63029503434081700, 0.44702148453447300,
        0.51917666806190000, 0.45798818046963000, 0.61380284424104300,
        0.65671506502013300, 0.31529729947567100, 0.10805355058691200,
        0.11047499102161600, 0.06218376425181400, 0.04428978919828240,
        0.06376348297771220, 0.07787345727187060, 0.05450182606873110,
        0.05861997797933450, 0.05472403303561820, 0.05012718208397050,
    };

    const excess_annual = math.pow(f64, 1.0 + 0.02, 252.0) - 1.0;
    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.sterlingRatio(excess_annual);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, sterling_eps));
        }
    }
}

// ---------- Burke Ratio Tests ----------

test "burke ratio unmodified rf=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const burke_eps: f64 = 1e-12;

    const expected = [_]?f64{
        null,                null,                null,
        0.74155751780925900, 0.89279126631488400, 1.15889854414048000,
        1.22179559465039000, 1.89088510302265000, 0.88340826476302900,
        1.02600204980225000, 0.86912185514484500, 1.16481055500805000,
        1.24624485947780000, 0.43717141205593600, 0.12556405008668100,
        0.12837789439234000, 0.08147141226635310, 0.05962926105099300,
        0.08584753824355520, 0.10484440763127000, 0.06479655403731050,
        0.06969257444720940, 0.06501838430134020, 0.05929350254553110,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.burkeRatio(false);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, burke_eps));
        }
    }
}

test "burke ratio modified rf=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const burke_eps: f64 = 1e-12;

    const expected = [_]?f64{
        null,               null,               null,
        1.4831150356185200, 1.9963419611982000, 2.8387100967984600,
        3.2325672963992100, 5.3482307151677600, 2.6502247942890900,
        3.2445033613766200, 2.8825510906130700, 4.0350221249328900,
        4.4933997426306100, 1.6357456432054900, 0.4863074748680710,
        0.5135115775693600, 0.3359152382424160, 0.2529855290778000,
        0.3742007437554000, 0.4688784450484330, 0.2969351136482720,
        0.3268871495298580, 0.3118172170272280, 0.2904776525979330,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.burkeRatio(true);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, burke_eps));
        }
    }
}

// ---------- Drawdown Peaks Test ----------

test "drawdown peaks" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const expected_peaks = [_]f64{
        0.00000000000000000,  0.00000000000000000,  0.00000000000000000,
        -0.00999999999999890, 0.00000000000000000,  0.00000000000000000,
        0.00000000000000000,  0.00000000000000000,  -0.01400000000000290,
        0.00000000000000000,  -0.00499999999999945, 0.00000000000000000,
        0.00000000000000000,  -0.03699999999999810, -0.09797742999999580,
        -0.08099408616309980, -0.12995439906088300, -0.15192580909309000,
        -0.08203215715946180, -0.02407973581061150, -0.08906408398233760,
        -0.06508545936249050, -0.07008220508951670, -0.07907589769106100,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
    }
    const actual_peaks = ratios.drawdownsPeaks();
    for (expected_peaks, 0..) |exp, i| {
        try testing.expect(almostEqual(actual_peaks[i], exp, epsilon));
    }
}

// ---------- Pain Index Test ----------

test "pain index" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        0.000000000000000000, 0.000000000000000000, 0.000000000000000000,
        0.002499999999999720, 0.001999999999999780, 0.001666666666666480,
        0.001428571428571270, 0.001249999999999860, 0.002666666666666870,
        0.002400000000000180, 0.002636363636363750, 0.002416666666666770,
        0.002230769230769330, 0.004714285714285670, 0.010931828666666300,
        0.015310719760193400, 0.022054465601410500, 0.029269540239837100,
        0.032046520077712100, 0.031648180864357100, 0.034382271489022800,
        0.035777870937816800, 0.037269363727021100, 0.039011302642189500,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.painIndex();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

// ---------- Pain Ratio Test ----------

test "pain ratio rf=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const pain_ratio_eps: f64 = 1e-12;

    const expected = [_]?f64{
        null,               null,               null,
        2.9662300712370400, 4.4639563315744200, 6.9533912648428800,
        8.5525691625527300, 15.127080824181200, 5.6995239278141000,
        7.3550027975430300, 5.9064682584701500, 8.6355710500115400,
        10.009243404789200, 3.8122309845695300, 1.1440393448276400,
        0.8351473403007020, 0.4100547525167660, 0.2491781707736400,
        0.3276524622550170, 0.4051939805814540, 0.2610350161067210,
        0.2698071401733870, 0.2417956028428860, 0.2115948629646200,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.painRatio();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, pain_ratio_eps));
        }
    }
}

// ---------- Ulcer Index Test ----------

test "ulcer index" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const expected = [_]?f64{
        0.000000000000000000, 0.000000000000000000, 0.000000000000000000,
        0.004999999999999450, 0.004472135954999090, 0.004082482904638180,
        0.003779644730091860, 0.003535533905932350, 0.005734883511362320,
        0.005440588203494720, 0.005402019824271570, 0.005172040216394730,
        0.004969135507541710, 0.010987005311470400, 0.027434256917710300,
        0.033400616370435100, 0.045203959104378100, 0.056676085534212600,
        0.058286267631993700, 0.057065017555245000, 0.058983749023907400,
        0.059274727347882400, 0.059785256332054100, 0.060711532990550200,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.ulcerIndex();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

// ---------- Martin Ratio Test ----------

test "martin ratio rf=0" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithRF(allocator, 0);
    defer ratios.deinit();

    const martin_eps: f64 = 1e-12;

    const expected = [_]?f64{
        null,               null,               null,
        1.4831150356185200, 1.9963419611982000, 2.8387100967984600,
        3.2325672963992100, 5.3482307151677600, 2.6502247942890900,
        3.2445033613766200, 2.8825510906130700, 4.0350221249328900,
        4.4933997426306100, 1.6357456432054900, 0.4558695408843890,
        0.3828284707085000, 0.2000607604567100, 0.1286844429639630,
        0.1801474281465850, 0.2247200286966700, 0.1521601617470090,
        0.1628539762413560, 0.1507322845601700, 0.1359641377846650,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.martinRatio();
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, martin_eps));
        }
    }
}

// ---------- Min Periods Tests ----------

fn newRatiosWithWindow(allocator: Allocator, rolling_window: ?u32, min_periods: ?u32) Ratios {
    var r = Ratios.init(allocator, .daily, 0, 0, .raw, rolling_window, min_periods);
    r.reset();
    return r;
}

fn addBaconReturns(r: *Ratios, count: usize) !void {
    const n = if (count == 0) bacon_len else @min(count, bacon_len);
    for (0..n) |i| {
        try addBaconReturn(r, i);
    }
}

test "min_periods: sharpe nil before min_periods" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithWindow(allocator, null, 10);
    defer ratios.deinit();

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        if (i < 9) {
            try testing.expect(ratios.sharpeRatio(false, false) == null);
        } else {
            try testing.expect(ratios.sharpeRatio(false, false) != null);
        }
    }
}

test "min_periods: all ratios nil before min_periods" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithWindow(allocator, null, 10);
    defer ratios.deinit();

    try addBaconReturns(&ratios, 9);

    try testing.expect(ratios.sharpeRatio(false, false) == null);
    try testing.expect(ratios.sortinoRatio(false, false) == null);
    try testing.expect(ratios.omegaRatio() == null);
    try testing.expect(ratios.kappaRatio(1) == null);
    try testing.expect(ratios.kappaRatio(2) == null);
    try testing.expect(ratios.kappaRatio(3) == null);
    try testing.expect(ratios.kappa3Ratio() == null);
    try testing.expect(ratios.bernardoLedoitRatio() == null);
    try testing.expect(ratios.upsidePotentialRatio(true) == null);
    try testing.expect(ratios.compoundGrowthRate() == null);
    try testing.expect(ratios.calmarRatio() == null);
    try testing.expect(ratios.sterlingRatio(0) == null);
    try testing.expect(ratios.painIndex() == null);
    try testing.expect(ratios.painRatio() == null);
    try testing.expect(ratios.ulcerIndex() == null);
    try testing.expect(ratios.martinRatio() == null);
    try testing.expect(ratios.kurtosis() == null);
    try testing.expect(ratios.skew() == null);
    try testing.expect(ratios.gainToPainRatio() == null);
    try testing.expect(ratios.riskOfRuin() == null);
    try testing.expect(ratios.riskReturnRatio() == null);

    // Feed one more to hit min_periods
    try addBaconReturn(&ratios, 9);
    try testing.expect(ratios.sharpeRatio(false, false) != null);
    try testing.expect(ratios.kurtosis() != null);
}

test "min_periods: zero is ignored" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithWindow(allocator, null, 0);
    defer ratios.deinit();

    try addBaconReturn(&ratios, 0);
    // CumulativeReturn is always valid (f64)
    try testing.expect(ratios.cumulativeReturn() != 0 or bacon_portfolio_returns[0] == 0);
}

// ---------- Rolling Window Tests ----------

test "rolling window: matches fresh" {
    const allocator = testing.allocator;
    const window: u32 = 10;
    var r_rolling = newRatiosWithWindow(allocator, window, null);
    defer r_rolling.deinit();
    try addBaconReturns(&r_rolling, 0); // all

    var r_fresh = newRatiosWithWindow(allocator, null, null);
    defer r_fresh.deinit();
    const start = bacon_len - window;
    for (start..bacon_len) |i| {
        try addBaconReturn(&r_fresh, i);
    }

    const match_eps: f64 = 1e-13;
    try assertNullableFloatEps("rolling sharpe", r_fresh.sharpeRatio(false, false), r_rolling.sharpeRatio(false, false), match_eps);
    try assertNullableFloatEps("rolling sortino", r_fresh.sortinoRatio(false, false), r_rolling.sortinoRatio(false, false), match_eps);
    try testing.expect(almostEqual(r_rolling.cumulativeReturn(), r_fresh.cumulativeReturn(), match_eps));
    try assertNullableFloatEps("rolling kurtosis", r_fresh.kurtosis(), r_rolling.kurtosis(), match_eps);
    try assertNullableFloatEps("rolling omega", r_fresh.omegaRatio(), r_rolling.omegaRatio(), match_eps);
    try assertNullableFloatEps("rolling calmar", r_fresh.calmarRatio(), r_rolling.calmarRatio(), match_eps);
    try assertNullableFloatEps("rolling pain_index", r_fresh.painIndex(), r_rolling.painIndex(), match_eps);
    try assertNullableFloatEps("rolling ulcer_index", r_fresh.ulcerIndex(), r_rolling.ulcerIndex(), match_eps);
    try assertNullableFloatEps("rolling martin", r_fresh.martinRatio(), r_rolling.martinRatio(), match_eps);
    try testing.expect(almostEqual(r_rolling.worstDrawdownsCumulative(), r_fresh.worstDrawdownsCumulative(), match_eps));
}

test "rolling window: sharpe step by step" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithWindow(allocator, 10, null);
    defer ratios.deinit();

    const expected = [_]?f64{
        null,
        0.8915694197569513,
        1.1419253390798365,
        0.49779248369997886,
        0.6680426571226848,
        0.8511810078441023,
        0.9735918376312113,
        0.8462916062735413,
        0.6475912629068395,
        0.7524743687246648,
        0.6988231811021255,
        0.7111123104828202,
        0.798675261552181,
        0.6310757998776281,
        0.3386466454024338,
        0.32170438498662823,
        0.16115775541041388,
        -0.022215518961695248,
        0.14832204365045173,
        0.17865069359303465,
        0.05655715365926667,
        -0.049597686094872355,
        -0.14538530360069923,
        -0.08934238062974807,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        const actual = ratios.sharpeRatio(false, false);
        if (expected[i] == null) {
            try testing.expect(actual == null);
        } else {
            try testing.expect(actual != null);
            try testing.expect(almostEqual(actual.?, expected[i].?, epsilon));
        }
    }
}

test "rolling window: cumulative return step by step" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithWindow(allocator, 10, null);
    defer ratios.deinit();

    const expected = [_]f64{
        0.0029999999999998916, 0.029077999999999937,
        0.04039785799999973,   0.02999387941999987,
        0.045443787611299635,  0.07157988230158208,
        0.08872516041840739,   0.1616697461664407,
        0.14540636972011045,   0.19122262450891503,
        0.18172134734433754,   0.24506898292322488,
        0.2807831278339803,    0.2458526788930535,
        0.1525671581089434,    0.14357151199687346,
        0.0704099487293568,    -0.018874479983775894,
        0.06471024991618646,   0.08313792731858194,
        0.017823077430024314,  -0.035845669483492104,
        -0.07756388570776418,  -0.050743313329589035,
    };

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        try testing.expect(almostEqual(ratios.cumulativeReturn(), expected[i], epsilon));
    }
}

test "rolling window: nil is expanding" {
    const allocator = testing.allocator;
    var r_expanding = newRatiosWithWindow(allocator, null, null);
    defer r_expanding.deinit();
    try addBaconReturns(&r_expanding, 0);

    var r_none = newRatiosWithWindow(allocator, null, null);
    defer r_none.deinit();
    try addBaconReturns(&r_none, 0);

    try assertNullableFloatEps("expanding sharpe", r_expanding.sharpeRatio(false, false), r_none.sharpeRatio(false, false), epsilon);
    try testing.expect(almostEqual(r_expanding.cumulativeReturn(), r_none.cumulativeReturn(), epsilon));
}

// ---------- Rolling Window With Min Periods Tests ----------

test "rolling window with min_periods: combined" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithWindow(allocator, 10, 5);
    defer ratios.deinit();

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        if (i < 4) {
            try testing.expect(ratios.sharpeRatio(false, false) == null);
            try testing.expect(ratios.kurtosis() == null);
        } else {
            try testing.expect(ratios.kurtosis() != null);
        }
    }
}

test "rolling window with min_periods: min_periods > window" {
    const allocator = testing.allocator;
    var ratios = newRatiosWithWindow(allocator, 5, 10);
    defer ratios.deinit();

    for (0..bacon_len) |i| {
        try addBaconReturn(&ratios, i);
        if (i < 9) {
            try testing.expect(ratios.sharpeRatio(false, false) == null);
        } else {
            try testing.expect(ratios.sharpeRatio(false, false) != null);
        }
    }
}
