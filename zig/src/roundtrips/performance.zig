const std = @import("std");
const math = std.math;
const conventions = @import("conventions");
const fractional = @import("fractional");
const side_mod = @import("side");
const rt_mod = @import("roundtrip");
const exec_mod = @import("execution");

const DayCountConvention = conventions.DayCountConvention;
const DateTime = fractional.DateTime;
const RoundtripSide = side_mod.RoundtripSide;
const Roundtrip = rt_mod.Roundtrip;
const Execution = exec_mod.Execution;
const OrderSide = exec_mod.OrderSide;

const Allocator = std.mem.Allocator;
const ArrayList = std.ArrayList;

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

fn sliceMean(s: []const f64) f64 {
    if (s.len == 0) return 0;
    var sum: f64 = 0;
    for (s) |v| sum += v;
    return sum / @as(f64, @floatFromInt(s.len));
}

fn sliceStdPop(s: []const f64) f64 {
    if (s.len == 0) return 0;
    const m = sliceMean(s);
    var sum: f64 = 0;
    for (s) |v| {
        const d = v - m;
        sum += d * d;
    }
    return @sqrt(sum / @as(f64, @floatFromInt(s.len)));
}

fn maxConsecutive(bools: []const bool) usize {
    var max_streak: usize = 0;
    var current: usize = 0;
    for (bools) |b| {
        if (b) {
            current += 1;
            if (current > max_streak) max_streak = current;
        } else {
            current = 0;
        }
    }
    return max_streak;
}

fn divOrZero(a: f64, b: usize) f64 {
    if (b > 0) return a / @as(f64, @floatFromInt(b));
    return 0.0;
}

fn minSlice(s: []const f64) f64 {
    if (s.len == 0) return 0;
    var m = s[0];
    for (s[1..]) |v| {
        if (v < m) m = v;
    }
    return m;
}

fn maxSlice(s: []const f64) f64 {
    if (s.len == 0) return 0;
    var m = s[0];
    for (s[1..]) |v| {
        if (v > m) m = v;
    }
    return m;
}

// ---------------------------------------------------------------------------
// RoundtripPerformance
// ---------------------------------------------------------------------------

const PnlField = enum { gross, net };
const PnlSign = enum { positive, negative, any };

pub const RoundtripPerformance = struct {
    initial_balance: f64,
    annual_risk_free_rate: f64,
    annual_target_return: f64,
    day_count_convention: DayCountConvention,

    roundtrips: ArrayList(Roundtrip),
    returns_on_investments: ArrayList(f64),
    sortino_downside_returns: ArrayList(f64),
    returns_on_investments_annual: ArrayList(f64),
    sortino_downside_returns_annual: ArrayList(f64),

    first_time: ?DateTime,
    last_time: ?DateTime,
    max_net_pnl: f64,
    max_drawdown: f64,
    max_drawdown_percent: f64,

    total_commission: f64,
    gross_winning_commission: f64,
    gross_loosing_commission: f64,
    net_winning_commission: f64,
    net_loosing_commission: f64,
    gross_winning_long_commission: f64,
    gross_loosing_long_commission: f64,
    net_winning_long_commission: f64,
    net_loosing_long_commission: f64,
    gross_winning_short_commission: f64,
    gross_loosing_short_commission: f64,
    net_winning_short_commission: f64,
    net_loosing_short_commission: f64,

    net_pnl: f64,
    gross_pnl: f64,
    gross_winning_pnl: f64,
    gross_loosing_pnl: f64,
    net_winning_pnl: f64,
    net_loosing_pnl: f64,
    gross_long_pnl: f64,
    gross_short_pnl: f64,
    net_long_pnl: f64,
    net_short_pnl: f64,
    gross_long_winning_pnl: f64,
    gross_long_loosing_pnl: f64,
    net_long_winning_pnl: f64,
    net_long_loosing_pnl: f64,
    gross_short_winning_pnl: f64,
    gross_short_loosing_pnl: f64,
    net_short_winning_pnl: f64,
    net_short_loosing_pnl: f64,

    total_count: usize,
    long_count: usize,
    short_count: usize,
    gross_winning_count: usize,
    gross_loosing_count: usize,
    net_winning_count: usize,
    net_loosing_count: usize,
    gross_long_winning_count: usize,
    gross_long_loosing_count: usize,
    net_long_winning_count: usize,
    net_long_loosing_count: usize,
    gross_short_winning_count: usize,
    gross_short_loosing_count: usize,
    net_short_winning_count: usize,
    net_short_loosing_count: usize,

    duration_sec: f64,
    duration_sec_long: f64,
    duration_sec_short: f64,
    duration_sec_gross_winning: f64,
    duration_sec_gross_loosing: f64,
    duration_sec_net_winning: f64,
    duration_sec_net_loosing: f64,
    duration_sec_gross_long_winning: f64,
    duration_sec_gross_long_loosing: f64,
    duration_sec_net_long_winning: f64,
    duration_sec_net_long_loosing: f64,
    duration_sec_gross_short_winning: f64,
    duration_sec_gross_short_loosing: f64,
    duration_sec_net_short_winning: f64,
    duration_sec_net_short_loosing: f64,
    total_duration_annualized: f64,

    total_mae: f64,
    total_mfe: f64,
    total_eff: f64,
    total_eff_entry: f64,
    total_eff_exit: f64,

    roi_mean: ?f64,
    roi_std: ?f64,
    roi_tdd: ?f64,
    roiann_mean: ?f64,
    roiann_std: ?f64,
    roiann_tdd: ?f64,

    allocator: Allocator,

    pub fn init(allocator: Allocator, initial_balance: f64, annual_risk_free_rate: f64, annual_target_return: f64, day_count_convention: DayCountConvention) RoundtripPerformance {
        return .{
            .initial_balance = initial_balance,
            .annual_risk_free_rate = annual_risk_free_rate,
            .annual_target_return = annual_target_return,
            .day_count_convention = day_count_convention,
            .roundtrips = .empty,
            .returns_on_investments = .empty,
            .sortino_downside_returns = .empty,
            .returns_on_investments_annual = .empty,
            .sortino_downside_returns_annual = .empty,
            .first_time = null,
            .last_time = null,
            .max_net_pnl = 0,
            .max_drawdown = 0,
            .max_drawdown_percent = 0,
            .total_commission = 0,
            .gross_winning_commission = 0,
            .gross_loosing_commission = 0,
            .net_winning_commission = 0,
            .net_loosing_commission = 0,
            .gross_winning_long_commission = 0,
            .gross_loosing_long_commission = 0,
            .net_winning_long_commission = 0,
            .net_loosing_long_commission = 0,
            .gross_winning_short_commission = 0,
            .gross_loosing_short_commission = 0,
            .net_winning_short_commission = 0,
            .net_loosing_short_commission = 0,
            .net_pnl = 0,
            .gross_pnl = 0,
            .gross_winning_pnl = 0,
            .gross_loosing_pnl = 0,
            .net_winning_pnl = 0,
            .net_loosing_pnl = 0,
            .gross_long_pnl = 0,
            .gross_short_pnl = 0,
            .net_long_pnl = 0,
            .net_short_pnl = 0,
            .gross_long_winning_pnl = 0,
            .gross_long_loosing_pnl = 0,
            .net_long_winning_pnl = 0,
            .net_long_loosing_pnl = 0,
            .gross_short_winning_pnl = 0,
            .gross_short_loosing_pnl = 0,
            .net_short_winning_pnl = 0,
            .net_short_loosing_pnl = 0,
            .total_count = 0,
            .long_count = 0,
            .short_count = 0,
            .gross_winning_count = 0,
            .gross_loosing_count = 0,
            .net_winning_count = 0,
            .net_loosing_count = 0,
            .gross_long_winning_count = 0,
            .gross_long_loosing_count = 0,
            .net_long_winning_count = 0,
            .net_long_loosing_count = 0,
            .gross_short_winning_count = 0,
            .gross_short_loosing_count = 0,
            .net_short_winning_count = 0,
            .net_short_loosing_count = 0,
            .duration_sec = 0,
            .duration_sec_long = 0,
            .duration_sec_short = 0,
            .duration_sec_gross_winning = 0,
            .duration_sec_gross_loosing = 0,
            .duration_sec_net_winning = 0,
            .duration_sec_net_loosing = 0,
            .duration_sec_gross_long_winning = 0,
            .duration_sec_gross_long_loosing = 0,
            .duration_sec_net_long_winning = 0,
            .duration_sec_net_long_loosing = 0,
            .duration_sec_gross_short_winning = 0,
            .duration_sec_gross_short_loosing = 0,
            .duration_sec_net_short_winning = 0,
            .duration_sec_net_short_loosing = 0,
            .total_duration_annualized = 0,
            .total_mae = 0,
            .total_mfe = 0,
            .total_eff = 0,
            .total_eff_entry = 0,
            .total_eff_exit = 0,
            .roi_mean = null,
            .roi_std = null,
            .roi_tdd = null,
            .roiann_mean = null,
            .roiann_std = null,
            .roiann_tdd = null,
            .allocator = allocator,
        };
    }

    pub fn deinit(self: *RoundtripPerformance) void {
        self.roundtrips.deinit(self.allocator);
        self.returns_on_investments.deinit(self.allocator);
        self.sortino_downside_returns.deinit(self.allocator);
        self.returns_on_investments_annual.deinit(self.allocator);
        self.sortino_downside_returns_annual.deinit(self.allocator);
    }

    pub fn reset(self: *RoundtripPerformance) void {
        self.roundtrips.clearRetainingCapacity();
        self.returns_on_investments.clearRetainingCapacity();
        self.sortino_downside_returns.clearRetainingCapacity();
        self.returns_on_investments_annual.clearRetainingCapacity();
        self.sortino_downside_returns_annual.clearRetainingCapacity();
        self.first_time = null;
        self.last_time = null;
        self.max_net_pnl = 0;
        self.max_drawdown = 0;
        self.max_drawdown_percent = 0;
        self.total_commission = 0;
        self.gross_winning_commission = 0;
        self.gross_loosing_commission = 0;
        self.net_winning_commission = 0;
        self.net_loosing_commission = 0;
        self.gross_winning_long_commission = 0;
        self.gross_loosing_long_commission = 0;
        self.net_winning_long_commission = 0;
        self.net_loosing_long_commission = 0;
        self.gross_winning_short_commission = 0;
        self.gross_loosing_short_commission = 0;
        self.net_winning_short_commission = 0;
        self.net_loosing_short_commission = 0;
        self.net_pnl = 0;
        self.gross_pnl = 0;
        self.gross_winning_pnl = 0;
        self.gross_loosing_pnl = 0;
        self.net_winning_pnl = 0;
        self.net_loosing_pnl = 0;
        self.gross_long_pnl = 0;
        self.gross_short_pnl = 0;
        self.net_long_pnl = 0;
        self.net_short_pnl = 0;
        self.gross_long_winning_pnl = 0;
        self.gross_long_loosing_pnl = 0;
        self.net_long_winning_pnl = 0;
        self.net_long_loosing_pnl = 0;
        self.gross_short_winning_pnl = 0;
        self.gross_short_loosing_pnl = 0;
        self.net_short_winning_pnl = 0;
        self.net_short_loosing_pnl = 0;
        self.total_count = 0;
        self.long_count = 0;
        self.short_count = 0;
        self.gross_winning_count = 0;
        self.gross_loosing_count = 0;
        self.net_winning_count = 0;
        self.net_loosing_count = 0;
        self.gross_long_winning_count = 0;
        self.gross_long_loosing_count = 0;
        self.net_long_winning_count = 0;
        self.net_long_loosing_count = 0;
        self.gross_short_winning_count = 0;
        self.gross_short_loosing_count = 0;
        self.net_short_winning_count = 0;
        self.net_short_loosing_count = 0;
        self.duration_sec = 0;
        self.duration_sec_long = 0;
        self.duration_sec_short = 0;
        self.duration_sec_gross_winning = 0;
        self.duration_sec_gross_loosing = 0;
        self.duration_sec_net_winning = 0;
        self.duration_sec_net_loosing = 0;
        self.duration_sec_gross_long_winning = 0;
        self.duration_sec_gross_long_loosing = 0;
        self.duration_sec_net_long_winning = 0;
        self.duration_sec_net_long_loosing = 0;
        self.duration_sec_gross_short_winning = 0;
        self.duration_sec_gross_short_loosing = 0;
        self.duration_sec_net_short_winning = 0;
        self.duration_sec_net_short_loosing = 0;
        self.total_duration_annualized = 0;
        self.total_mae = 0;
        self.total_mfe = 0;
        self.total_eff = 0;
        self.total_eff_entry = 0;
        self.total_eff_exit = 0;
        self.roi_mean = null;
        self.roi_std = null;
        self.roi_tdd = null;
        self.roiann_mean = null;
        self.roiann_std = null;
        self.roiann_tdd = null;
    }

    pub fn addRoundtrip(self: *RoundtripPerformance, rt: Roundtrip) !void {
        try self.roundtrips.append(self.allocator, rt);
        self.total_count += 1;
        const comm = rt.commission;
        self.total_commission += comm;
        const secs = rt.duration_seconds;
        self.duration_sec += secs;
        self.total_mae += rt.maximum_adverse_excursion;
        self.total_mfe += rt.maximum_favorable_excursion;
        self.total_eff += rt.total_efficiency;
        self.total_eff_entry += rt.entry_efficiency;
        self.total_eff_exit += rt.exit_efficiency;

        const net_pnl_val = rt.net_pnl;
        self.net_pnl += net_pnl_val;
        if (net_pnl_val > 0) {
            self.net_winning_count += 1;
            self.net_winning_pnl += net_pnl_val;
            self.net_winning_commission += comm;
            self.duration_sec_net_winning += secs;
        } else if (net_pnl_val < 0) {
            self.net_loosing_count += 1;
            self.net_loosing_pnl += net_pnl_val;
            self.net_loosing_commission += comm;
            self.duration_sec_net_loosing += secs;
        }

        const gross_pnl_val = rt.gross_pnl;
        self.gross_pnl += gross_pnl_val;
        if (gross_pnl_val > 0) {
            self.gross_winning_count += 1;
            self.gross_winning_pnl += gross_pnl_val;
            self.gross_winning_commission += comm;
            self.duration_sec_gross_winning += secs;
        } else if (gross_pnl_val < 0) {
            self.gross_loosing_count += 1;
            self.gross_loosing_pnl += gross_pnl_val;
            self.gross_loosing_commission += comm;
            self.duration_sec_gross_loosing += secs;
        }

        if (rt.side == .long) {
            self.gross_long_pnl += gross_pnl_val;
            self.net_long_pnl += net_pnl_val;
            self.long_count += 1;
            self.duration_sec_long += secs;
            if (gross_pnl_val > 0) {
                self.gross_long_winning_count += 1;
                self.gross_long_winning_pnl += gross_pnl_val;
                self.gross_winning_long_commission += comm;
                self.duration_sec_gross_long_winning += secs;
            } else if (gross_pnl_val < 0) {
                self.gross_long_loosing_count += 1;
                self.gross_long_loosing_pnl += gross_pnl_val;
                self.gross_loosing_long_commission += comm;
                self.duration_sec_gross_long_loosing += secs;
            }
            // CRITICAL: uses gross_pnl_val, not net_pnl_val (intentional quirk)
            if (net_pnl_val > 0) {
                self.net_long_winning_count += 1;
                self.net_long_winning_pnl += gross_pnl_val;
                self.net_winning_long_commission += comm;
                self.duration_sec_net_long_winning += secs;
            } else if (net_pnl_val < 0) {
                self.net_long_loosing_count += 1;
                self.net_long_loosing_pnl += gross_pnl_val;
                self.net_loosing_long_commission += comm;
                self.duration_sec_net_long_loosing += secs;
            }
        } else {
            self.gross_short_pnl += gross_pnl_val;
            self.net_short_pnl += net_pnl_val;
            self.short_count += 1;
            self.duration_sec_short += secs;
            if (gross_pnl_val > 0) {
                self.gross_short_winning_count += 1;
                self.gross_short_winning_pnl += gross_pnl_val;
                self.gross_winning_short_commission += comm;
                self.duration_sec_gross_short_winning += secs;
            } else if (gross_pnl_val < 0) {
                self.gross_short_loosing_count += 1;
                self.gross_short_loosing_pnl += gross_pnl_val;
                self.gross_loosing_short_commission += comm;
                self.duration_sec_gross_short_loosing += secs;
            }
            // CRITICAL: uses gross_pnl_val, not net_pnl_val (intentional quirk)
            if (net_pnl_val > 0) {
                self.net_short_winning_count += 1;
                self.net_short_winning_pnl += gross_pnl_val;
                self.net_winning_short_commission += comm;
                self.duration_sec_net_short_winning += secs;
            } else if (net_pnl_val < 0) {
                self.net_short_loosing_count += 1;
                self.net_short_loosing_pnl += gross_pnl_val;
                self.net_loosing_short_commission += comm;
                self.duration_sec_net_short_loosing += secs;
            }
        }

        // Update first/last times and duration
        const entry_dt = rt.entryTime();
        const exit_dt = rt.exitTime();
        var changed = false;
        if (self.first_time == null or self.first_time.?.toTotalSeconds() > entry_dt.toTotalSeconds()) {
            self.first_time = entry_dt;
            changed = true;
        }
        if (self.last_time == null or self.last_time.?.toTotalSeconds() < exit_dt.toTotalSeconds()) {
            self.last_time = exit_dt;
            changed = true;
        }
        if (changed) {
            if (self.first_time) |ft| {
                if (self.last_time) |lt| {
                    if (fractional.yearFrac(ft, lt, self.day_count_convention)) |yf| {
                        self.total_duration_annualized = yf;
                    } else |_| {}
                }
            }
        }

        const roi = net_pnl_val / (rt.quantity * rt.entry_price);
        try self.returns_on_investments.append(self.allocator, roi);
        self.roi_mean = sliceMean(self.returns_on_investments.items);
        self.roi_std = sliceStdPop(self.returns_on_investments.items);

        const downside = roi - self.annual_risk_free_rate;
        if (downside < 0) {
            try self.sortino_downside_returns.append(self.allocator, downside);
            var sum_sq: f64 = 0;
            for (self.sortino_downside_returns.items) |v| sum_sq += v * v;
            self.roi_tdd = @sqrt(sum_sq / @as(f64, @floatFromInt(self.sortino_downside_returns.items.len)));
        }

        // Calculate annualized ROI
        if (fractional.yearFrac(entry_dt, exit_dt, self.day_count_convention)) |yf| {
            if (yf != 0) {
                const roiann = roi / yf;
                try self.returns_on_investments_annual.append(self.allocator, roiann);
                self.roiann_mean = sliceMean(self.returns_on_investments_annual.items);
                self.roiann_std = sliceStdPop(self.returns_on_investments_annual.items);

                const downside_ann = roiann - self.annual_risk_free_rate;
                if (downside_ann < 0) {
                    try self.sortino_downside_returns_annual.append(self.allocator, downside_ann);
                    var sum_sq: f64 = 0;
                    for (self.sortino_downside_returns_annual.items) |v| sum_sq += v * v;
                    self.roiann_tdd = @sqrt(sum_sq / @as(f64, @floatFromInt(self.sortino_downside_returns_annual.items.len)));
                }
            }
        } else |_| {}

        // Calculate max drawdown
        if (self.max_net_pnl < self.net_pnl) {
            self.max_net_pnl = self.net_pnl;
        }
        const dd = self.max_net_pnl - self.net_pnl;
        if (self.max_drawdown < dd) {
            self.max_drawdown = dd;
            self.max_drawdown_percent = self.max_drawdown / (self.initial_balance + self.max_net_pnl);
        }
    }

    // --- ROI statistics ---
    pub fn getRoiMean(self: *const RoundtripPerformance) ?f64 {
        return self.roi_mean;
    }
    pub fn getRoiStd(self: *const RoundtripPerformance) ?f64 {
        return self.roi_std;
    }
    pub fn getRoiTdd(self: *const RoundtripPerformance) ?f64 {
        return self.roi_tdd;
    }
    pub fn getRoiannMean(self: *const RoundtripPerformance) ?f64 {
        return self.roiann_mean;
    }
    pub fn getRoiannStd(self: *const RoundtripPerformance) ?f64 {
        return self.roiann_std;
    }
    pub fn getRoiannTdd(self: *const RoundtripPerformance) ?f64 {
        return self.roiann_tdd;
    }

    // --- Risk-adjusted ratios ---
    pub fn sharpeRatio(self: *const RoundtripPerformance) ?f64 {
        const m = self.roi_mean orelse return null;
        const s = self.roi_std orelse return null;
        if (s == 0) return null;
        return m / s;
    }

    pub fn sharpeRatioAnnual(self: *const RoundtripPerformance) ?f64 {
        const m = self.roiann_mean orelse return null;
        const s = self.roiann_std orelse return null;
        if (s == 0) return null;
        return m / s;
    }

    pub fn sortinoRatio(self: *const RoundtripPerformance) ?f64 {
        const m = self.roi_mean orelse return null;
        const t = self.roi_tdd orelse return null;
        if (t == 0) return null;
        return (m - self.annual_risk_free_rate) / t;
    }

    pub fn sortinoRatioAnnual(self: *const RoundtripPerformance) ?f64 {
        const m = self.roiann_mean orelse return null;
        const t = self.roiann_tdd orelse return null;
        if (t == 0) return null;
        return (m - self.annual_risk_free_rate) / t;
    }

    pub fn calmarRatio(self: *const RoundtripPerformance) ?f64 {
        const m = self.roi_mean orelse return null;
        if (self.max_drawdown_percent == 0) return null;
        return m / self.max_drawdown_percent;
    }

    pub fn calmarRatioAnnual(self: *const RoundtripPerformance) ?f64 {
        const m = self.roiann_mean orelse return null;
        if (self.max_drawdown_percent == 0) return null;
        return m / self.max_drawdown_percent;
    }

    // --- Rate of return ---
    pub fn rateOfReturn(self: *const RoundtripPerformance) ?f64 {
        if (self.initial_balance == 0) return null;
        return self.net_pnl / self.initial_balance;
    }

    pub fn rateOfReturnAnnual(self: *const RoundtripPerformance) ?f64 {
        if (self.total_duration_annualized == 0 or self.initial_balance == 0) return null;
        return (self.net_pnl / self.initial_balance) / self.total_duration_annualized;
    }

    pub fn recoveryFactor(self: *const RoundtripPerformance) ?f64 {
        const rorann = self.rateOfReturnAnnual() orelse return null;
        if (self.max_drawdown_percent == 0) return null;
        return rorann / self.max_drawdown_percent;
    }

    // --- Profit ratios ---
    pub fn grossProfitRatio(self: *const RoundtripPerformance) ?f64 {
        if (self.gross_loosing_pnl == 0) return null;
        return @abs(self.gross_winning_pnl / self.gross_loosing_pnl);
    }

    pub fn netProfitRatio(self: *const RoundtripPerformance) ?f64 {
        if (self.net_loosing_pnl == 0) return null;
        return @abs(self.net_winning_pnl / self.net_loosing_pnl);
    }

    pub fn grossProfitLongRatio(self: *const RoundtripPerformance) ?f64 {
        if (self.gross_long_loosing_pnl == 0) return null;
        return @abs(self.gross_long_winning_pnl / self.gross_long_loosing_pnl);
    }

    pub fn netProfitLongRatio(self: *const RoundtripPerformance) ?f64 {
        if (self.net_long_loosing_pnl == 0) return null;
        return @abs(self.net_long_winning_pnl / self.net_long_loosing_pnl);
    }

    pub fn grossProfitShortRatio(self: *const RoundtripPerformance) ?f64 {
        if (self.gross_short_loosing_pnl == 0) return null;
        return @abs(self.gross_short_winning_pnl / self.gross_short_loosing_pnl);
    }

    pub fn netProfitShortRatio(self: *const RoundtripPerformance) ?f64 {
        if (self.net_short_loosing_pnl == 0) return null;
        return @abs(self.net_short_winning_pnl / self.net_short_loosing_pnl);
    }

    // --- Counts ---
    pub fn totalCount(self: *const RoundtripPerformance) usize {
        return self.total_count;
    }
    pub fn longCount(self: *const RoundtripPerformance) usize {
        return self.long_count;
    }
    pub fn shortCount(self: *const RoundtripPerformance) usize {
        return self.short_count;
    }
    pub fn grossWinningCount(self: *const RoundtripPerformance) usize {
        return self.gross_winning_count;
    }
    pub fn grossLoosingCount(self: *const RoundtripPerformance) usize {
        return self.gross_loosing_count;
    }
    pub fn netWinningCount(self: *const RoundtripPerformance) usize {
        return self.net_winning_count;
    }
    pub fn netLoosingCount(self: *const RoundtripPerformance) usize {
        return self.net_loosing_count;
    }
    pub fn grossLongWinningCount(self: *const RoundtripPerformance) usize {
        return self.gross_long_winning_count;
    }
    pub fn grossLongLoosingCount(self: *const RoundtripPerformance) usize {
        return self.gross_long_loosing_count;
    }
    pub fn netLongWinningCount(self: *const RoundtripPerformance) usize {
        return self.net_long_winning_count;
    }
    pub fn netLongLoosingCount(self: *const RoundtripPerformance) usize {
        return self.net_long_loosing_count;
    }
    pub fn grossShortWinningCount(self: *const RoundtripPerformance) usize {
        return self.gross_short_winning_count;
    }
    pub fn grossShortLoosingCount(self: *const RoundtripPerformance) usize {
        return self.gross_short_loosing_count;
    }
    pub fn netShortWinningCount(self: *const RoundtripPerformance) usize {
        return self.net_short_winning_count;
    }
    pub fn netShortLoosingCount(self: *const RoundtripPerformance) usize {
        return self.net_short_loosing_count;
    }

    // --- Win/loss ratios ---
    pub fn grossWinningRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.gross_winning_count)), self.total_count);
    }
    pub fn grossLoosingRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.gross_loosing_count)), self.total_count);
    }
    pub fn netWinningRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.net_winning_count)), self.total_count);
    }
    pub fn netLoosingRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.net_loosing_count)), self.total_count);
    }
    pub fn grossLongWinningRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.gross_long_winning_count)), self.long_count);
    }
    pub fn grossLongLoosingRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.gross_long_loosing_count)), self.long_count);
    }
    pub fn netLongWinningRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.net_long_winning_count)), self.long_count);
    }
    pub fn netLongLoosingRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.net_long_loosing_count)), self.long_count);
    }
    pub fn grossShortWinningRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.gross_short_winning_count)), self.short_count);
    }
    pub fn grossShortLoosingRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.gross_short_loosing_count)), self.short_count);
    }
    pub fn netShortWinningRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.net_short_winning_count)), self.short_count);
    }
    pub fn netShortLoosingRatio(self: *const RoundtripPerformance) f64 {
        return divOrZero(@as(f64, @floatFromInt(self.net_short_loosing_count)), self.short_count);
    }

    // --- PnL totals ---
    pub fn totalGrossPnl(self: *const RoundtripPerformance) f64 {
        return self.gross_pnl;
    }
    pub fn totalNetPnl(self: *const RoundtripPerformance) f64 {
        return self.net_pnl;
    }
    pub fn winningGrossPnl(self: *const RoundtripPerformance) f64 {
        return self.gross_winning_pnl;
    }
    pub fn loosingGrossPnl(self: *const RoundtripPerformance) f64 {
        return self.gross_loosing_pnl;
    }
    pub fn winningNetPnl(self: *const RoundtripPerformance) f64 {
        return self.net_winning_pnl;
    }
    pub fn loosingNetPnl(self: *const RoundtripPerformance) f64 {
        return self.net_loosing_pnl;
    }
    pub fn winningGrossLongPnl(self: *const RoundtripPerformance) f64 {
        return self.gross_long_winning_pnl;
    }
    pub fn loosingGrossLongPnl(self: *const RoundtripPerformance) f64 {
        return self.gross_long_loosing_pnl;
    }
    pub fn winningNetLongPnl(self: *const RoundtripPerformance) f64 {
        return self.net_long_winning_pnl;
    }
    pub fn loosingNetLongPnl(self: *const RoundtripPerformance) f64 {
        return self.net_long_loosing_pnl;
    }
    pub fn winningGrossShortPnl(self: *const RoundtripPerformance) f64 {
        return self.gross_short_winning_pnl;
    }
    pub fn loosingGrossShortPnl(self: *const RoundtripPerformance) f64 {
        return self.gross_short_loosing_pnl;
    }
    pub fn winningNetShortPnl(self: *const RoundtripPerformance) f64 {
        return self.net_short_winning_pnl;
    }
    pub fn loosingNetShortPnl(self: *const RoundtripPerformance) f64 {
        return self.net_short_loosing_pnl;
    }

    // --- Average PnL ---
    pub fn averageGrossPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.gross_pnl, self.total_count);
    }
    pub fn averageNetPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.net_pnl, self.total_count);
    }
    pub fn averageGrossLongPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.gross_long_pnl, self.long_count);
    }
    pub fn averageNetLongPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.net_long_pnl, self.long_count);
    }
    pub fn averageGrossShortPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.gross_short_pnl, self.short_count);
    }
    pub fn averageNetShortPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.net_short_pnl, self.short_count);
    }
    pub fn averageWinningGrossPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.gross_winning_pnl, self.gross_winning_count);
    }
    pub fn averageLoosingGrossPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.gross_loosing_pnl, self.gross_loosing_count);
    }
    pub fn averageWinningNetPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.net_winning_pnl, self.net_winning_count);
    }
    pub fn averageLoosingNetPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.net_loosing_pnl, self.net_loosing_count);
    }
    pub fn averageWinningGrossLongPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.gross_long_winning_pnl, self.gross_long_winning_count);
    }
    pub fn averageLoosingGrossLongPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.gross_long_loosing_pnl, self.gross_long_loosing_count);
    }
    pub fn averageWinningNetLongPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.net_long_winning_pnl, self.net_long_winning_count);
    }
    pub fn averageLoosingNetLongPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.net_long_loosing_pnl, self.net_long_loosing_count);
    }
    pub fn averageWinningGrossShortPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.gross_short_winning_pnl, self.gross_short_winning_count);
    }
    pub fn averageLoosingGrossShortPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.gross_short_loosing_pnl, self.gross_short_loosing_count);
    }
    pub fn averageWinningNetShortPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.net_short_winning_pnl, self.net_short_winning_count);
    }
    pub fn averageLoosingNetShortPnl(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.net_short_loosing_pnl, self.net_short_loosing_count);
    }

    // --- Average win/loss ratios ---
    pub fn averageGrossWinningLoosingRatio(self: *const RoundtripPerformance) f64 {
        const l = self.averageLoosingGrossPnl();
        if (l != 0) return self.averageWinningGrossPnl() / l;
        return 0.0;
    }
    pub fn averageNetWinningLoosingRatio(self: *const RoundtripPerformance) f64 {
        const l = self.averageLoosingNetPnl();
        if (l != 0) return self.averageWinningNetPnl() / l;
        return 0.0;
    }
    pub fn averageGrossWinningLoosingLongRatio(self: *const RoundtripPerformance) f64 {
        const l = self.averageLoosingGrossLongPnl();
        if (l != 0) return self.averageWinningGrossLongPnl() / l;
        return 0.0;
    }
    pub fn averageNetWinningLoosingLongRatio(self: *const RoundtripPerformance) f64 {
        const l = self.averageLoosingNetLongPnl();
        if (l != 0) return self.averageWinningNetLongPnl() / l;
        return 0.0;
    }
    pub fn averageGrossWinningLoosingShortRatio(self: *const RoundtripPerformance) f64 {
        const l = self.averageLoosingGrossShortPnl();
        if (l != 0) return self.averageWinningGrossShortPnl() / l;
        return 0.0;
    }
    pub fn averageNetWinningLoosingShortRatio(self: *const RoundtripPerformance) f64 {
        const l = self.averageLoosingNetShortPnl();
        if (l != 0) return self.averageWinningNetShortPnl() / l;
        return 0.0;
    }

    // --- Profit PnL ratios ---
    pub fn grossProfitPnlRatio(self: *const RoundtripPerformance) f64 {
        if (self.gross_pnl != 0) return self.gross_winning_pnl / self.gross_pnl;
        return 0.0;
    }
    pub fn netProfitPnlRatio(self: *const RoundtripPerformance) f64 {
        if (self.net_pnl != 0) return self.net_winning_pnl / self.net_pnl;
        return 0.0;
    }
    pub fn grossProfitPnlLongRatio(self: *const RoundtripPerformance) f64 {
        if (self.gross_long_pnl != 0) return self.gross_long_winning_pnl / self.gross_long_pnl;
        return 0.0;
    }
    pub fn netProfitPnlLongRatio(self: *const RoundtripPerformance) f64 {
        if (self.net_long_pnl != 0) return self.net_long_winning_pnl / self.net_long_pnl;
        return 0.0;
    }
    pub fn grossProfitPnlShortRatio(self: *const RoundtripPerformance) f64 {
        if (self.gross_short_pnl != 0) return self.gross_short_winning_pnl / self.gross_short_pnl;
        return 0.0;
    }
    pub fn netProfitPnlShortRatio(self: *const RoundtripPerformance) f64 {
        if (self.net_short_pnl != 0) return self.net_short_winning_pnl / self.net_short_pnl;
        return 0.0;
    }

    // --- Average duration ---
    pub fn averageDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec, self.total_count);
    }
    pub fn averageGrossWinningDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_gross_winning, self.gross_winning_count);
    }
    pub fn averageGrossLoosingDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_gross_loosing, self.gross_loosing_count);
    }
    pub fn averageNetWinningDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_net_winning, self.net_winning_count);
    }
    pub fn averageNetLoosingDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_net_loosing, self.net_loosing_count);
    }
    pub fn averageLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_long, self.long_count);
    }
    pub fn averageShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_short, self.short_count);
    }
    pub fn averageGrossWinningLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_gross_long_winning, self.gross_long_winning_count);
    }
    pub fn averageGrossLoosingLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_gross_long_loosing, self.gross_long_loosing_count);
    }
    pub fn averageNetWinningLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_net_long_winning, self.net_long_winning_count);
    }
    pub fn averageNetLoosingLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_net_long_loosing, self.net_long_loosing_count);
    }
    pub fn averageGrossWinningShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_gross_short_winning, self.gross_short_winning_count);
    }
    pub fn averageGrossLoosingShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_gross_short_loosing, self.gross_short_loosing_count);
    }
    pub fn averageNetWinningShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_net_short_winning, self.net_short_winning_count);
    }
    pub fn averageNetLoosingShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.duration_sec_net_short_loosing, self.net_short_loosing_count);
    }

    // --- Min/max duration helpers ---
    fn filterDurationSeconds(self: *const RoundtripPerformance, allocator: Allocator, comptime filter: fn (Roundtrip) bool) ![]f64 {
        var result = ArrayList(f64).empty;
        for (self.roundtrips.items) |r| {
            if (filter(r)) {
                try result.append(allocator, r.duration_seconds);
            }
        }
        return result.items;
    }

    fn filterDurationSecondsDyn(self: *const RoundtripPerformance, allocator: Allocator, side_filter: ?RoundtripSide, pnl_field: PnlField, pnl_sign: PnlSign) !ArrayList(f64) {
        var result = ArrayList(f64).empty;
        for (self.roundtrips.items) |r| {
            if (side_filter) |sf| {
                if (r.side != sf) continue;
            }
            switch (pnl_field) {
                .gross => switch (pnl_sign) {
                    .positive => if (!(r.gross_pnl > 0)) continue,
                    .negative => if (!(r.gross_pnl < 0)) continue,
                    .any => {},
                },
                .net => switch (pnl_sign) {
                    .positive => if (!(r.net_pnl > 0)) continue,
                    .negative => if (!(r.net_pnl < 0)) continue,
                    .any => {},
                },
            }
            try result.append(allocator, r.duration_seconds);
        }
        return result;
    }

    pub fn minimumDurationSeconds(self: *const RoundtripPerformance) f64 {
        var result = ArrayList(f64).empty;
        defer result.deinit(self.allocator);
        for (self.roundtrips.items) |r| {
            result.append(self.allocator, r.duration_seconds) catch return 0;
        }
        return minSlice(result.items);
    }
    pub fn maximumDurationSeconds(self: *const RoundtripPerformance) f64 {
        var result = ArrayList(f64).empty;
        defer result.deinit(self.allocator);
        for (self.roundtrips.items) |r| {
            result.append(self.allocator, r.duration_seconds) catch return 0;
        }
        return maxSlice(result.items);
    }

    fn filteredMinDuration(self: *const RoundtripPerformance, side_filter: ?RoundtripSide, pnl_field: PnlField, pnl_sign: PnlSign) f64 {
        var result = self.filterDurationSecondsDyn(self.allocator, side_filter, pnl_field, pnl_sign) catch return 0;
        defer result.deinit(self.allocator);
        return minSlice(result.items);
    }
    fn filteredMaxDuration(self: *const RoundtripPerformance, side_filter: ?RoundtripSide, pnl_field: PnlField, pnl_sign: PnlSign) f64 {
        var result = self.filterDurationSecondsDyn(self.allocator, side_filter, pnl_field, pnl_sign) catch return 0;
        defer result.deinit(self.allocator);
        return maxSlice(result.items);
    }

    pub fn minimumLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(.long, .gross, .any);
    }
    pub fn maximumLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(.long, .gross, .any);
    }
    pub fn minimumShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(.short, .gross, .any);
    }
    pub fn maximumShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(.short, .gross, .any);
    }
    pub fn minimumGrossWinningDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(null, .gross, .positive);
    }
    pub fn maximumGrossWinningDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(null, .gross, .positive);
    }
    pub fn minimumGrossLoosingDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(null, .gross, .negative);
    }
    pub fn maximumGrossLoosingDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(null, .gross, .negative);
    }
    pub fn minimumNetWinningDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(null, .net, .positive);
    }
    pub fn maximumNetWinningDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(null, .net, .positive);
    }
    pub fn minimumNetLoosingDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(null, .net, .negative);
    }
    pub fn maximumNetLoosingDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(null, .net, .negative);
    }
    pub fn minimumGrossWinningLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(.long, .gross, .positive);
    }
    pub fn maximumGrossWinningLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(.long, .gross, .positive);
    }
    pub fn minimumGrossLoosingLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(.long, .gross, .negative);
    }
    pub fn maximumGrossLoosingLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(.long, .gross, .negative);
    }
    pub fn minimumNetWinningLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(.long, .net, .positive);
    }
    pub fn maximumNetWinningLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(.long, .net, .positive);
    }
    pub fn minimumNetLoosingLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(.long, .net, .negative);
    }
    pub fn maximumNetLoosingLongDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(.long, .net, .negative);
    }
    pub fn minimumGrossWinningShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(.short, .gross, .positive);
    }
    pub fn maximumGrossWinningShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(.short, .gross, .positive);
    }
    pub fn minimumGrossLoosingShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(.short, .gross, .negative);
    }
    pub fn maximumGrossLoosingShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(.short, .gross, .negative);
    }
    pub fn minimumNetWinningShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(.short, .net, .positive);
    }
    pub fn maximumNetWinningShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(.short, .net, .positive);
    }
    pub fn minimumNetLoosingShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMinDuration(.short, .net, .negative);
    }
    pub fn maximumNetLoosingShortDurationSeconds(self: *const RoundtripPerformance) f64 {
        return self.filteredMaxDuration(.short, .net, .negative);
    }

    // --- MAE / MFE / efficiency ---
    pub fn averageMaximumAdverseExcursion(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.total_mae, self.total_count);
    }
    pub fn averageMaximumFavorableExcursion(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.total_mfe, self.total_count);
    }
    pub fn averageEntryEfficiency(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.total_eff_entry, self.total_count);
    }
    pub fn averageExitEfficiency(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.total_eff_exit, self.total_count);
    }
    pub fn averageTotalEfficiency(self: *const RoundtripPerformance) f64 {
        return divOrZero(self.total_eff, self.total_count);
    }

    // Filtered average helper for MAE/MFE/efficiency
    fn filteredAvg(self: *const RoundtripPerformance, comptime field: []const u8, side_filter: ?RoundtripSide, pnl_field: PnlField, pnl_sign: enum { positive, negative }, count: usize) f64 {
        if (count == 0) return 0.0;
        var sum: f64 = 0;
        for (self.roundtrips.items) |r| {
            if (side_filter) |sf| {
                if (r.side != sf) continue;
            }
            const passes = switch (pnl_field) {
                .gross => switch (pnl_sign) {
                    .positive => r.gross_pnl > 0,
                    .negative => r.gross_pnl < 0,
                },
                .net => switch (pnl_sign) {
                    .positive => r.net_pnl > 0,
                    .negative => r.net_pnl < 0,
                },
            };
            if (passes) {
                sum += @field(r, field);
            }
        }
        return sum / @as(f64, @floatFromInt(count));
    }

    pub fn averageMaximumAdverseExcursionGrossWinning(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("maximum_adverse_excursion", null, .gross, .positive, self.gross_winning_count);
    }
    pub fn averageMaximumAdverseExcursionGrossLoosing(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("maximum_adverse_excursion", null, .gross, .negative, self.gross_loosing_count);
    }
    pub fn averageMaximumAdverseExcursionNetWinning(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("maximum_adverse_excursion", null, .net, .positive, self.net_winning_count);
    }
    pub fn averageMaximumAdverseExcursionNetLoosing(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("maximum_adverse_excursion", null, .net, .negative, self.net_loosing_count);
    }
    pub fn averageMaximumFavorableExcursionGrossWinning(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("maximum_favorable_excursion", null, .gross, .positive, self.gross_winning_count);
    }
    pub fn averageMaximumFavorableExcursionGrossLoosing(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("maximum_favorable_excursion", null, .gross, .negative, self.gross_loosing_count);
    }
    pub fn averageMaximumFavorableExcursionNetWinning(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("maximum_favorable_excursion", null, .net, .positive, self.net_winning_count);
    }
    pub fn averageMaximumFavorableExcursionNetLoosing(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("maximum_favorable_excursion", null, .net, .negative, self.net_loosing_count);
    }
    pub fn averageEntryEfficiencyGrossWinning(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("entry_efficiency", null, .gross, .positive, self.gross_winning_count);
    }
    pub fn averageEntryEfficiencyGrossLoosing(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("entry_efficiency", null, .gross, .negative, self.gross_loosing_count);
    }
    pub fn averageEntryEfficiencyNetWinning(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("entry_efficiency", null, .net, .positive, self.net_winning_count);
    }
    pub fn averageEntryEfficiencyNetLoosing(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("entry_efficiency", null, .net, .negative, self.net_loosing_count);
    }
    pub fn averageExitEfficiencyGrossWinning(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("exit_efficiency", null, .gross, .positive, self.gross_winning_count);
    }
    pub fn averageExitEfficiencyGrossLoosing(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("exit_efficiency", null, .gross, .negative, self.gross_loosing_count);
    }
    pub fn averageExitEfficiencyNetWinning(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("exit_efficiency", null, .net, .positive, self.net_winning_count);
    }
    pub fn averageExitEfficiencyNetLoosing(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("exit_efficiency", null, .net, .negative, self.net_loosing_count);
    }
    pub fn averageTotalEfficiencyGrossWinning(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("total_efficiency", null, .gross, .positive, self.gross_winning_count);
    }
    pub fn averageTotalEfficiencyGrossLoosing(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("total_efficiency", null, .gross, .negative, self.gross_loosing_count);
    }
    pub fn averageTotalEfficiencyNetWinning(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("total_efficiency", null, .net, .positive, self.net_winning_count);
    }
    pub fn averageTotalEfficiencyNetLoosing(self: *const RoundtripPerformance) f64 {
        return self.filteredAvg("total_efficiency", null, .net, .negative, self.net_loosing_count);
    }

    // --- Consecutive streaks ---
    pub fn maxConsecutiveGrossWinners(self: *const RoundtripPerformance) usize {
        var bools: [256]bool = undefined;
        const n = @min(self.roundtrips.items.len, 256);
        for (self.roundtrips.items[0..n], 0..) |r, i| bools[i] = r.gross_pnl > 0;
        return maxConsecutive(bools[0..n]);
    }
    pub fn maxConsecutiveGrossLoosers(self: *const RoundtripPerformance) usize {
        var bools: [256]bool = undefined;
        const n = @min(self.roundtrips.items.len, 256);
        for (self.roundtrips.items[0..n], 0..) |r, i| bools[i] = r.gross_pnl < 0;
        return maxConsecutive(bools[0..n]);
    }
    pub fn maxConsecutiveNetWinners(self: *const RoundtripPerformance) usize {
        var bools: [256]bool = undefined;
        const n = @min(self.roundtrips.items.len, 256);
        for (self.roundtrips.items[0..n], 0..) |r, i| bools[i] = r.net_pnl > 0;
        return maxConsecutive(bools[0..n]);
    }
    pub fn maxConsecutiveNetLoosers(self: *const RoundtripPerformance) usize {
        var bools: [256]bool = undefined;
        const n = @min(self.roundtrips.items.len, 256);
        for (self.roundtrips.items[0..n], 0..) |r, i| bools[i] = r.net_pnl < 0;
        return maxConsecutive(bools[0..n]);
    }
};

// ===========================================================================
// Tests
// ===========================================================================

const testing = std.testing;

fn almostEqual(a: f64, b: f64, epsilon: f64) bool {
    return @abs(a - b) < epsilon;
}

fn optAlmostEqual(a: ?f64, b: f64, epsilon: f64) bool {
    if (a) |v| return almostEqual(v, b, epsilon);
    return false;
}

// Test roundtrip helpers
fn makeRt1() Roundtrip {
    return Roundtrip.init(
        .{ .side = .buy, .price = 50.0, .commission_per_unit = 0.01, .unrealized_price_high = 56.0, .unrealized_price_low = 48.0, .year = 2024, .month = 1, .day = 1, .hour = 9, .minute = 30, .second = 0 },
        .{ .side = .sell, .price = 55.0, .commission_per_unit = 0.02, .unrealized_price_high = 57.0, .unrealized_price_low = 49.0, .year = 2024, .month = 1, .day = 5, .hour = 16, .minute = 0, .second = 0 },
        100.0,
    );
}

fn makeRt2() Roundtrip {
    return Roundtrip.init(
        .{ .side = .sell, .price = 80.0, .commission_per_unit = 0.03, .unrealized_price_high = 85.0, .unrealized_price_low = 72.0, .year = 2024, .month = 2, .day = 1, .hour = 10, .minute = 0, .second = 0 },
        .{ .side = .buy, .price = 72.0, .commission_per_unit = 0.02, .unrealized_price_high = 83.0, .unrealized_price_low = 70.0, .year = 2024, .month = 2, .day = 10, .hour = 15, .minute = 30, .second = 0 },
        200.0,
    );
}

fn makeRt3() Roundtrip {
    return Roundtrip.init(
        .{ .side = .buy, .price = 60.0, .commission_per_unit = 0.005, .unrealized_price_high = 62.0, .unrealized_price_low = 53.0, .year = 2024, .month = 3, .day = 1, .hour = 9, .minute = 30, .second = 0 },
        .{ .side = .sell, .price = 54.0, .commission_per_unit = 0.005, .unrealized_price_high = 61.0, .unrealized_price_low = 52.0, .year = 2024, .month = 3, .day = 3, .hour = 16, .minute = 0, .second = 0 },
        150.0,
    );
}

fn makeRt4() Roundtrip {
    return Roundtrip.init(
        .{ .side = .sell, .price = 40.0, .commission_per_unit = 0.01, .unrealized_price_high = 42.0, .unrealized_price_low = 39.0, .year = 2024, .month = 4, .day = 1, .hour = 10, .minute = 0, .second = 0 },
        .{ .side = .buy, .price = 45.0, .commission_per_unit = 0.01, .unrealized_price_high = 46.0, .unrealized_price_low = 38.0, .year = 2024, .month = 4, .day = 5, .hour = 15, .minute = 0, .second = 0 },
        300.0,
    );
}

fn makeRt5() Roundtrip {
    return Roundtrip.init(
        .{ .side = .buy, .price = 100.0, .commission_per_unit = 0.02, .unrealized_price_high = 112.0, .unrealized_price_low = 98.0, .year = 2024, .month = 5, .day = 1, .hour = 9, .minute = 0, .second = 0 },
        .{ .side = .sell, .price = 110.0, .commission_per_unit = 0.02, .unrealized_price_high = 115.0, .unrealized_price_low = 99.0, .year = 2024, .month = 5, .day = 15, .hour = 16, .minute = 0, .second = 0 },
        50.0,
    );
}

fn makeRt6() Roundtrip {
    return Roundtrip.init(
        .{ .side = .sell, .price = 90.0, .commission_per_unit = 0.015, .unrealized_price_high = 92.0, .unrealized_price_low = 84.0, .year = 2024, .month = 6, .day = 1, .hour = 10, .minute = 0, .second = 0 },
        .{ .side = .buy, .price = 82.0, .commission_per_unit = 0.015, .unrealized_price_high = 93.0, .unrealized_price_low = 80.0, .year = 2024, .month = 6, .day = 20, .hour = 15, .minute = 0, .second = 0 },
        100.0,
    );
}

fn makeAllRts() [6]Roundtrip {
    return .{ makeRt1(), makeRt2(), makeRt3(), makeRt4(), makeRt5(), makeRt6() };
}

fn addAllRts(perf: *RoundtripPerformance) !void {
    const rts = makeAllRts();
    for (&rts) |*r| {
        try perf.addRoundtrip(r.*);
    }
}

// --- Init tests ---

test "perf init default initial balance" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expect(almostEqual(perf.initial_balance, 100000.0, 1e-13));
}

test "perf init default annual risk free rate" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expect(almostEqual(perf.annual_risk_free_rate, 0.0, 1e-13));
}

test "perf init total count zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expectEqual(@as(usize, 0), perf.totalCount());
}

test "perf init roi mean none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expectEqual(@as(?f64, null), perf.getRoiMean());
}

test "perf init roi std none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expectEqual(@as(?f64, null), perf.getRoiStd());
}

test "perf init roi tdd none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expectEqual(@as(?f64, null), perf.getRoiTdd());
}

test "perf init sharpe ratio none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expectEqual(@as(?f64, null), perf.sharpeRatio());
}

test "perf init sortino ratio none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expectEqual(@as(?f64, null), perf.sortinoRatio());
}

test "perf init calmar ratio none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expectEqual(@as(?f64, null), perf.calmarRatio());
}

test "perf init empty roundtrips list" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expectEqual(@as(usize, 0), perf.roundtrips.items.len);
}

test "perf init total gross pnl zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expect(almostEqual(perf.totalGrossPnl(), 0.0, 1e-13));
}

test "perf init total net pnl zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expect(almostEqual(perf.totalNetPnl(), 0.0, 1e-13));
}

test "perf init max drawdown zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expect(almostEqual(perf.max_drawdown, 0.0, 1e-13));
}

test "perf init average net pnl zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expect(almostEqual(perf.averageNetPnl(), 0.0, 1e-13));
}

// --- Reset tests ---

test "perf reset total count zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try perf.addRoundtrip(makeRt3());
    perf.reset();
    try testing.expectEqual(@as(usize, 0), perf.totalCount());
}

test "perf reset total net pnl zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try perf.addRoundtrip(makeRt3());
    perf.reset();
    try testing.expect(almostEqual(perf.totalNetPnl(), 0.0, 1e-13));
}

test "perf reset roi mean none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try perf.addRoundtrip(makeRt3());
    perf.reset();
    try testing.expectEqual(@as(?f64, null), perf.getRoiMean());
}

test "perf reset roundtrips list empty" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try perf.addRoundtrip(makeRt3());
    perf.reset();
    try testing.expectEqual(@as(usize, 0), perf.roundtrips.items.len);
}

test "perf reset returns on investments empty" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try perf.addRoundtrip(makeRt3());
    perf.reset();
    try testing.expectEqual(@as(usize, 0), perf.returns_on_investments.items.len);
}

test "perf reset max drawdown zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try perf.addRoundtrip(makeRt3());
    perf.reset();
    try testing.expect(almostEqual(perf.max_drawdown, 0.0, 1e-13));
}

// --- Single long winner tests ---

test "perf single winner total count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(usize, 1), perf.totalCount());
}

test "perf single winner long count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(usize, 1), perf.longCount());
}

test "perf single winner short count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(usize, 0), perf.shortCount());
}

test "perf single winner gross winning count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(usize, 1), perf.grossWinningCount());
}

test "perf single winner total gross pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expect(almostEqual(perf.totalGrossPnl(), 500.0, 1e-13));
}

test "perf single winner total net pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expect(almostEqual(perf.totalNetPnl(), 497.0, 1e-13));
}

test "perf single winner total commission" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expect(almostEqual(perf.total_commission, 3.0, 1e-13));
}

test "perf single winner roi mean" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expect(optAlmostEqual(perf.getRoiMean(), 0.0994, 1e-13));
}

test "perf single winner roi std zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expect(optAlmostEqual(perf.getRoiStd(), 0.0, 1e-13));
}

test "perf single winner roi tdd none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(?f64, null), perf.getRoiTdd());
}

test "perf single winner sharpe ratio none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(?f64, null), perf.sharpeRatio());
}

test "perf single winner sortino ratio none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(?f64, null), perf.sortinoRatio());
}

test "perf single winner calmar ratio none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(?f64, null), perf.calmarRatio());
}

test "perf single winner max drawdown zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expect(almostEqual(perf.max_drawdown, 0.0, 1e-13));
}

test "perf single winner rate of return" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expect(optAlmostEqual(perf.rateOfReturn(), 0.00497, 1e-13));
}

test "perf single winner gross winning ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expect(almostEqual(perf.grossWinningRatio(), 1.0, 1e-13));
}

test "perf single winner net winning ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expect(almostEqual(perf.netWinningRatio(), 1.0, 1e-13));
}

test "perf single winner gross profit ratio none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(?f64, null), perf.grossProfitRatio());
}

test "perf single winner net profit ratio none" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(?f64, null), perf.netProfitRatio());
}

test "perf single winner average mae" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    const rt1 = makeRt1();
    try perf.addRoundtrip(rt1);
    try testing.expect(almostEqual(perf.averageMaximumAdverseExcursion(), rt1.maximum_adverse_excursion, 1e-13));
}

test "perf single winner average duration" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expect(almostEqual(perf.averageDurationSeconds(), 369000.0, 1e-13));
}

test "perf single winner max consecutive gross winners" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(usize, 1), perf.maxConsecutiveGrossWinners());
}

test "perf single winner max consecutive gross loosers" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(usize, 0), perf.maxConsecutiveGrossLoosers());
}

// --- Single loser tests ---

test "perf single loser total net pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt3());
    try testing.expect(almostEqual(perf.totalNetPnl(), -901.5, 1e-13));
}

test "perf single loser max drawdown" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt3());
    try testing.expect(almostEqual(perf.max_drawdown, 901.5, 1e-13));
}

test "perf single loser max drawdown percent" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt3());
    try testing.expect(almostEqual(perf.max_drawdown_percent, 0.009015, 1e-13));
}

test "perf single loser calmar ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt3());
    try testing.expect(optAlmostEqual(perf.calmarRatio(), -11.11111111111111, 1e-10));
}

test "perf single loser roi mean" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt3());
    try testing.expect(optAlmostEqual(perf.getRoiMean(), -0.10016666666666667, 1e-13));
}

test "perf single loser roi tdd" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt3());
    try testing.expect(optAlmostEqual(perf.getRoiTdd(), 0.10016666666666667, 1e-13));
}

test "perf single loser sortino ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt3());
    try testing.expect(optAlmostEqual(perf.sortinoRatio(), -1.0, 1e-13));
}

test "perf single loser gross loosing count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt3());
    try testing.expectEqual(@as(usize, 1), perf.grossLoosingCount());
}

test "perf single loser net loosing count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt3());
    try testing.expectEqual(@as(usize, 1), perf.netLoosingCount());
}

// --- Multiple mixed roundtrips tests ---

test "perf multi total count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 6), perf.totalCount());
}

test "perf multi long count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 3), perf.longCount());
}

test "perf multi short count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 3), perf.shortCount());
}

test "perf multi gross winning count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 4), perf.grossWinningCount());
}

test "perf multi gross loosing count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 2), perf.grossLoosingCount());
}

test "perf multi net winning count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 4), perf.netWinningCount());
}

test "perf multi net loosing count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 2), perf.netLoosingCount());
}

test "perf multi gross long winning count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 2), perf.grossLongWinningCount());
}

test "perf multi gross long loosing count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 1), perf.grossLongLoosingCount());
}

test "perf multi net long winning count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 2), perf.netLongWinningCount());
}

test "perf multi net long loosing count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 1), perf.netLongLoosingCount());
}

test "perf multi gross short winning count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 2), perf.grossShortWinningCount());
}

test "perf multi gross short loosing count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 1), perf.grossShortLoosingCount());
}

test "perf multi net short winning count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 2), perf.netShortWinningCount());
}

test "perf multi net short loosing count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 1), perf.netShortLoosingCount());
}

// --- PnL totals ---

test "perf multi total gross pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.totalGrossPnl(), 1000.0, 1e-13));
}

test "perf multi total net pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.totalNetPnl(), 974.5, 1e-13));
}

test "perf multi winning gross pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.winningGrossPnl(), 3400.0, 1e-13));
}

test "perf multi loosing gross pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.loosingGrossPnl(), -2400.0, 1e-13));
}

test "perf multi winning net pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.winningNetPnl(), 3382.0, 1e-13));
}

test "perf multi loosing net pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.loosingNetPnl(), -2407.5, 1e-13));
}

test "perf multi winning gross long pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.winningGrossLongPnl(), 1000.0, 1e-13));
}

test "perf multi loosing gross long pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.loosingGrossLongPnl(), -900.0, 1e-13));
}

test "perf multi winning gross short pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.winningGrossShortPnl(), 2400.0, 1e-13));
}

test "perf multi loosing gross short pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.loosingGrossShortPnl(), -1500.0, 1e-13));
}

// --- Commission ---

test "perf multi total commission" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.total_commission, 25.5, 1e-13));
}

test "perf multi gross winning commission" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.gross_winning_commission, 18.0, 1e-13));
}

test "perf multi gross loosing commission" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.gross_loosing_commission, 7.5, 1e-13));
}

test "perf multi net winning commission" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.net_winning_commission, 18.0, 1e-13));
}

test "perf multi net loosing commission" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.net_loosing_commission, 7.5, 1e-13));
}

// --- Average PnL ---

test "perf multi average gross pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageGrossPnl(), 1000.0 / 6.0, 1e-13));
}

test "perf multi average net pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageNetPnl(), 974.5 / 6.0, 1e-13));
}

test "perf multi average winning gross pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageWinningGrossPnl(), 3400.0 / 4.0, 1e-13));
}

test "perf multi average loosing gross pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageLoosingGrossPnl(), -2400.0 / 2.0, 1e-13));
}

test "perf multi average winning net pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageWinningNetPnl(), 3382.0 / 4.0, 1e-13));
}

test "perf multi average loosing net pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageLoosingNetPnl(), -2407.5 / 2.0, 1e-13));
}

test "perf multi average gross long pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageGrossLongPnl(), 100.0 / 3.0, 1e-13));
}

test "perf multi average gross short pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageGrossShortPnl(), 300.0, 1e-13));
}

// --- Win/loss ratios ---

test "perf multi gross winning ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.grossWinningRatio(), 4.0 / 6.0, 1e-13));
}

test "perf multi gross loosing ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.grossLoosingRatio(), 2.0 / 6.0, 1e-13));
}

test "perf multi net winning ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.netWinningRatio(), 4.0 / 6.0, 1e-13));
}

test "perf multi net loosing ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.netLoosingRatio(), 2.0 / 6.0, 1e-13));
}

test "perf multi gross long winning ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.grossLongWinningRatio(), 2.0 / 3.0, 1e-13));
}

test "perf multi gross short winning ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.grossShortWinningRatio(), 2.0 / 3.0, 1e-13));
}

// --- Profit ratios ---

test "perf multi gross profit ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.grossProfitRatio(), 1.4166666666666667, 1e-13));
}

test "perf multi net profit ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.netProfitRatio(), 1.4047767393561785, 1e-13));
}

test "perf multi gross profit long ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.grossProfitLongRatio(), 1.1111111111111112, 1e-13));
}

test "perf multi gross profit short ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.grossProfitShortRatio(), 1.6, 1e-13));
}

// --- Profit PnL ratios ---

test "perf multi gross profit pnl ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.grossProfitPnlRatio(), 3.4, 1e-13));
}

test "perf multi net profit pnl ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.netProfitPnlRatio(), 3382.0 / 974.5, 1e-13));
}

// --- Average win/loss ratio ---

test "perf multi average gross winning loosing ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageGrossWinningLoosingRatio(), 850.0 / -1200.0, 1e-13));
}

test "perf multi average net winning loosing ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageNetWinningLoosingRatio(), 845.5 / -1203.75, 1e-13));
}

// --- ROI statistics ---

test "perf multi roi mean" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.getRoiMean(), 0.026877314814814812, 1e-13));
}

test "perf multi roi std" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.getRoiStd(), 0.0991356544050762, 1e-13));
}

test "perf multi roi tdd" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.getRoiTdd(), 0.11354208715518468, 1e-13));
}

test "perf multi roiann mean" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.getRoiannMean(), -1.7233887909446202, 1e-12));
}

test "perf multi roiann std" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.getRoiannStd(), 8.73138705463156, 1e-12));
}

test "perf multi roiann tdd" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.getRoiannTdd(), 13.751365296707874, 1e-12));
}

// --- Risk-adjusted ratios ---

test "perf multi sharpe ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.sharpeRatio(), 0.27111653194916085, 1e-13));
}

test "perf multi sharpe ratio annual" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.sharpeRatioAnnual(), -0.1973785814512082, 1e-12));
}

test "perf multi sortino ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.sortinoRatio(), 0.23671675841293985, 1e-13));
}

test "perf multi sortino ratio annual" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.sortinoRatioAnnual(), -0.1253249225629404, 1e-12));
}

test "perf multi calmar ratio" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.calmarRatio(), 1.139698624091381, 1e-12));
}

test "perf multi calmar ratio annual" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.calmarRatioAnnual(), -73.07812731097131, 1e-10));
}

// --- Rate of return ---

test "perf multi rate of return" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.rateOfReturn(), 0.009745, 1e-13));
}

test "perf multi rate of return annual" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.rateOfReturnAnnual(), 0.020786693247353695, 1e-12));
}

test "perf multi recovery factor" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(optAlmostEqual(perf.recoveryFactor(), 0.8814335009522727, 1e-12));
}

// --- Drawdown ---

test "perf multi max net pnl" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.max_net_pnl, 2087.0, 1e-13));
}

test "perf multi max drawdown" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.max_drawdown, 2407.5, 1e-13));
}

test "perf multi max drawdown percent" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.max_drawdown_percent, 2407.5 / (100000.0 + 2087.0), 1e-13));
}

// --- Duration ---

test "perf multi average duration seconds" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageDurationSeconds(), 770100.0, 1e-13));
}

test "perf multi average long duration seconds" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageLongDurationSeconds(), 600000.0, 1e-13));
}

test "perf multi average short duration seconds" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageShortDurationSeconds(), 940200.0, 1e-13));
}

test "perf multi average gross winning duration seconds" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageGrossWinningDurationSeconds(), 1015200.0, 1e-13));
}

test "perf multi average gross loosing duration seconds" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.averageGrossLoosingDurationSeconds(), 279900.0, 1e-13));
}

test "perf multi minimum duration seconds" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.minimumDurationSeconds(), 196200.0, 1e-13));
}

test "perf multi maximum duration seconds" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.maximumDurationSeconds(), 1659600.0, 1e-13));
}

test "perf multi minimum long duration seconds" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.minimumLongDurationSeconds(), 196200.0, 1e-13));
}

test "perf multi maximum long duration seconds" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.maximumLongDurationSeconds(), 1234800.0, 1e-13));
}

test "perf multi minimum short duration seconds" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.minimumShortDurationSeconds(), 363600.0, 1e-13));
}

test "perf multi maximum short duration seconds" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expect(almostEqual(perf.maximumShortDurationSeconds(), 1659600.0, 1e-13));
}

// --- MAE / MFE / Efficiency ---

test "perf multi average mae" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    const rts = makeAllRts();
    var sum: f64 = 0;
    for (&rts) |*r| sum += r.maximum_adverse_excursion;
    try testing.expect(almostEqual(perf.averageMaximumAdverseExcursion(), sum / 6.0, 1e-13));
}

test "perf multi average mfe" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    const rts = makeAllRts();
    var sum: f64 = 0;
    for (&rts) |*r| sum += r.maximum_favorable_excursion;
    try testing.expect(almostEqual(perf.averageMaximumFavorableExcursion(), sum / 6.0, 1e-13));
}

test "perf multi average entry efficiency" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    const rts = makeAllRts();
    var sum: f64 = 0;
    for (&rts) |*r| sum += r.entry_efficiency;
    try testing.expect(almostEqual(perf.averageEntryEfficiency(), sum / 6.0, 1e-13));
}

test "perf multi average exit efficiency" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    const rts = makeAllRts();
    var sum: f64 = 0;
    for (&rts) |*r| sum += r.exit_efficiency;
    try testing.expect(almostEqual(perf.averageExitEfficiency(), sum / 6.0, 1e-13));
}

test "perf multi average total efficiency" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    const rts = makeAllRts();
    var sum: f64 = 0;
    for (&rts) |*r| sum += r.total_efficiency;
    try testing.expect(almostEqual(perf.averageTotalEfficiency(), sum / 6.0, 1e-13));
}

// --- Consecutive ---

test "perf multi max consecutive gross winners" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 2), perf.maxConsecutiveGrossWinners());
}

test "perf multi max consecutive gross loosers" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 2), perf.maxConsecutiveGrossLoosers());
}

test "perf multi max consecutive net winners" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 2), perf.maxConsecutiveNetWinners());
}

test "perf multi max consecutive net loosers" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 2), perf.maxConsecutiveNetLoosers());
}

// --- Time tracking ---

test "perf multi first time" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    const ft = perf.first_time.?;
    try testing.expectEqual(@as(i32, 2024), ft.year);
    try testing.expectEqual(@as(u8, 1), ft.month);
    try testing.expectEqual(@as(u8, 1), ft.day);
    try testing.expectEqual(@as(u8, 9), ft.hour);
    try testing.expectEqual(@as(u8, 30), ft.minute);
}

test "perf multi last time" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    const lt = perf.last_time.?;
    try testing.expectEqual(@as(i32, 2024), lt.year);
    try testing.expectEqual(@as(u8, 6), lt.month);
    try testing.expectEqual(@as(u8, 20), lt.day);
    try testing.expectEqual(@as(u8, 15), lt.hour);
    try testing.expectEqual(@as(u8, 0), lt.minute);
}

// --- Edge cases ---

test "perf edge zero initial balance rate of return none" {
    var perf = RoundtripPerformance.init(testing.allocator, 0.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expectEqual(@as(?f64, null), perf.rateOfReturn());
}

test "perf edge no roundtrips average gross pnl zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expect(almostEqual(perf.averageGrossPnl(), 0.0, 1e-13));
}

test "perf edge no roundtrips average net pnl zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expect(almostEqual(perf.averageNetPnl(), 0.0, 1e-13));
}

test "perf edge no roundtrips gross winning ratio zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expect(almostEqual(perf.grossWinningRatio(), 0.0, 1e-13));
}

test "perf edge no roundtrips average duration zero" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expect(almostEqual(perf.averageDurationSeconds(), 0.0, 1e-13));
}

test "perf edge sharpe none single point" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(?f64, null), perf.sharpeRatio());
}

test "perf edge rate of return annual none zero duration" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try testing.expectEqual(@as(?f64, null), perf.rateOfReturnAnnual());
}

test "perf edge recovery factor none no drawdown" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try perf.addRoundtrip(makeRt1());
    try testing.expectEqual(@as(?f64, null), perf.recoveryFactor());
}

// --- Incremental update tests ---

test "perf incremental roi list length" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    const rts = makeAllRts();
    for (&rts, 0..) |*r, i| {
        try perf.addRoundtrip(r.*);
        try testing.expectEqual(i + 1, perf.returns_on_investments.items.len);
    }
}

test "perf incremental roi values" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    const expected = [_]f64{
        0.0994,
        0.099375,
        -0.10016666666666667,
        -0.1255,
        0.0996,
        0.08855555555555556,
    };
    for (expected, 0..) |exp, i| {
        try testing.expect(almostEqual(perf.returns_on_investments.items[i], exp, 1e-13));
    }
}

test "perf incremental sortino downside count" {
    var perf = RoundtripPerformance.init(testing.allocator, 100000.0, 0.0, 0.0, .raw);
    defer perf.deinit();
    try addAllRts(&perf);
    try testing.expectEqual(@as(usize, 2), perf.sortino_downside_returns.items.len);
}
