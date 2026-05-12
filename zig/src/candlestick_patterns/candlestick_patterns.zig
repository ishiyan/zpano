//! Streaming candlestick pattern recognition engine with fuzzy logic support.
//!
//! Each pattern method returns a continuous confidence value in [-100, +100],
//! where positive values indicate bullish signals and negative values indicate
//! bearish signals. The magnitude reflects the fuzzy confidence of the match.
//! Use `defuzzify.alphaCut` to convert to crisp {-100, 0, +100} if needed.
//!
//! The engine inspects the most recent N bars (stored in a ring buffer)
//! and the incrementally maintained running totals for each criterion, giving
//! O(1) per bar after the warmup period.

const std = @import("std");
const fuzzy = @import("fuzzy");
const membership = fuzzy.membership;
const operators = fuzzy.operators;
const defuzzify = fuzzy.defuzzify;

const MembershipShape = membership.MembershipShape;

// Re-export core types from the core/ submodules.
pub const core = @import("core/core.zig");

pub const RangeEntity = core.RangeEntity;
pub const OHLC = core.OHLC;
pub const Criterion = core.Criterion;
pub const CriterionState = core.CriterionState;
pub const PatternIdentifier = core.PatternIdentifier;
pub const pattern_count = core.pattern_count;

// Re-export primitives.
pub const isWhite = core.isWhite;
pub const isBlack = core.isBlack;
pub const realBodyLen = core.realBodyLen;
pub const upperShadow = core.upperShadow;
pub const lowerShadow = core.lowerShadow;
pub const isRealBodyGapUp = core.isRealBodyGapUp;
pub const isRealBodyGapDown = core.isRealBodyGapDown;
pub const isHighLowGapUp = core.isHighLowGapUp;
pub const isHighLowGapDown = core.isHighLowGapDown;
pub const candleRangeValue = core.candleRangeValue;

// Re-export defaults.
pub const default_long_body = core.default_long_body;
pub const default_very_long_body = core.default_very_long_body;
pub const default_short_body = core.default_short_body;
pub const default_doji_body = core.default_doji_body;
pub const default_long_shadow = core.default_long_shadow;
pub const default_very_long_shadow = core.default_very_long_shadow;
pub const default_short_shadow = core.default_short_shadow;
pub const default_very_short_shadow = core.default_very_short_shadow;
pub const default_near = core.default_near;
pub const default_far = core.default_far;
pub const default_equal = core.default_equal;

// ---------------------------------------------------------------------------
// CandlestickPatterns
// ---------------------------------------------------------------------------

// Minimum history size: 5-candle patterns + 10 default criterion period + 5 margin.
const min_history: usize = 20;

/// Candlestick pattern recognition engine.
///
/// Provides streaming bar-by-bar evaluation of 61 Japanese candlestick patterns.
/// Call update(open, high, low, close) for each new bar, then call any pattern
/// method to get the result for the current bar.
///
/// Pattern methods return a continuous float in [-100, +100]:
///   positive: bullish signal, negative: bearish signal, near zero: no match.
///   The magnitude reflects the fuzzy confidence of the match.
///   Hikkake and HikkakeModified may return intermediate values for
///   unconfirmed signals.
pub const CandlestickPatterns = struct {
    const Self = @This();

    // Fuzzy configuration.
    fuzz_ratio: f64,
    shape: MembershipShape,

    // Criteria states.
    long_body: CriterionState,
    very_long_body: CriterionState,
    short_body: CriterionState,
    doji_body: CriterionState,
    long_shadow: CriterionState,
    very_long_shadow: CriterionState,
    short_shadow: CriterionState,
    very_short_shadow: CriterionState,
    near: CriterionState,
    far: CriterionState,
    equal: CriterionState,

    // Ring buffer of recent bars: each entry is (open, high, low, close).
    // Size is the largest criterion period + 5 candles + 5 margin, floored at min_history.
    history: [min_history]OHLC,
    hist_size: usize,
    hist_start: usize,
    hist_len: usize,
    /// Number of bars fed so far.
    count: usize,

    // Hikkake modified state.
    hikmod_pattern_result: f64,
    hikmod_pattern_idx: i32,
    hikmod_confirmed: bool,
    hikmod_last_signal: f64,

    /// Creates a new CandlestickPatterns engine with default options.
    pub fn init() Self {
        return Self{
            .fuzz_ratio = 0.2,
            .shape = .sigmoid,
            .long_body = CriterionState.init(default_long_body, 5),
            .very_long_body = CriterionState.init(default_very_long_body, 5),
            .short_body = CriterionState.init(default_short_body, 5),
            .doji_body = CriterionState.init(default_doji_body, 5),
            .long_shadow = CriterionState.init(default_long_shadow, 5),
            .very_long_shadow = CriterionState.init(default_very_long_shadow, 5),
            .short_shadow = CriterionState.init(default_short_shadow, 5),
            .very_short_shadow = CriterionState.init(default_very_short_shadow, 5),
            .near = CriterionState.init(default_near, 5),
            .far = CriterionState.init(default_far, 5),
            .equal = CriterionState.init(default_equal, 5),
            .history = undefined,
            .hist_size = min_history,
            .hist_start = 0,
            .hist_len = 0,
            .count = 0,
            .hikmod_pattern_result = 0.0,
            .hikmod_pattern_idx = 0,
            .hikmod_confirmed = false,
            .hikmod_last_signal = 0.0,
        };
    }

    /// Feeds a new OHLC bar into the engine.
    ///
    /// After calling this, all pattern methods reflect the state including this bar.
    pub fn update(self: *Self, o: f64, h: f64, l: f64, c: f64) void {
        const b = OHLC{ .o = o, .h = h, .l = l, .c = c };
        if (self.hist_len == self.hist_size) {
            self.history[self.hist_start] = b;
            self.hist_start = (self.hist_start + 1) % self.hist_size;
        } else {
            const idx = (self.hist_start + self.hist_len) % self.hist_size;
            self.history[idx] = b;
            self.hist_len += 1;
        }
        self.long_body.push(o, h, l, c);
        self.very_long_body.push(o, h, l, c);
        self.short_body.push(o, h, l, c);
        self.doji_body.push(o, h, l, c);
        self.long_shadow.push(o, h, l, c);
        self.very_long_shadow.push(o, h, l, c);
        self.short_shadow.push(o, h, l, c);
        self.very_short_shadow.push(o, h, l, c);
        self.near.push(o, h, l, c);
        self.far.push(o, h, l, c);
        self.equal.push(o, h, l, c);
        self.count += 1;

        // Update stateful patterns.
        self.hikmod_confirmed = false;
        self.hikmod_last_signal = 0.0;
        self.hikkakeModifiedUpdate();
    }

    // ------------------------------------------------------------------
    // Helper: get bar at position relative to end.
    // shift=1 means the most recent bar, shift=2 the one before, etc.
    // ------------------------------------------------------------------

    /// Get OHLC of a bar. shift=1 is most recent, shift=2 is one before, etc.
    pub fn bar(self: *const Self, shift: usize) OHLC {
        const idx = (self.hist_start + self.hist_len - shift) % self.hist_size;
        return self.history[idx];
    }

    /// Checks if we have sufficient bars for a pattern requiring n_candles
    /// plus the maximum average_period of the given criteria.
    pub fn enough(self: *const Self, n_candles: usize, criteria_list: []const *const CriterionState) bool {
        if (self.hist_len < n_candles) return false;
        const avail = self.hist_len - n_candles;
        for (criteria_list) |cs| {
            if (avail < cs.criterion.average_period) return false;
        }
        return true;
    }

    // ------------------------------------------------------------------
    // Criterion average helpers (shift is from the end, 1-based)
    // ------------------------------------------------------------------

    /// Gets the criterion average value at a given shift from the most recent bar.
    ///
    /// TA-Lib convention: the average uses the `period` bars BEFORE the reference bar
    /// (excluding the reference bar itself).
    pub fn avgCS(self: *const Self, cs: *const CriterionState, shift: usize) f64 {
        const b_val = self.bar(shift);
        return cs.avg(shift, b_val.o, b_val.h, b_val.l, b_val.c);
    }

    // ------------------------------------------------------------------
    // Fuzzy membership helpers
    // ------------------------------------------------------------------

    /// Fuzzy 'value < avg': degree of membership.
    pub fn muLessCs(self: *const Self, value: f64, cs: *const CriterionState, shift: usize) f64 {
        const avg_val = self.avgCS(cs, shift);
        var w = self.fuzz_ratio * avg_val;
        if (avg_val <= 0.0) w = 0.0;
        return membership.muLess(value, avg_val, w, self.shape);
    }

    /// Fuzzy 'value > avg': degree of membership.
    pub fn muGreaterCs(self: *const Self, value: f64, cs: *const CriterionState, shift: usize) f64 {
        const avg_val = self.avgCS(cs, shift);
        var w = self.fuzz_ratio * avg_val;
        if (avg_val <= 0.0) w = 0.0;
        return membership.muGreater(value, avg_val, w, self.shape);
    }

    /// Fuzzy 'value ≈ target ± avg': degree of closeness.
    pub fn muNearValue(self: *const Self, value: f64, target: f64, cs: *const CriterionState, shift: usize) f64 {
        const avg_val = self.avgCS(cs, shift);
        var w = self.fuzz_ratio * avg_val;
        if (avg_val <= 0.0) w = 0.0;
        return membership.muNear(value, target, w, self.shape);
    }

    /// Fuzzy 'value >= threshold' with explicit width (no criterion).
    pub fn muGeRaw(self: *const Self, value: f64, threshold: f64, width: f64) f64 {
        return membership.muGreaterEqual(value, threshold, width, self.shape);
    }

    /// Fuzzy 'value > threshold' with explicit width (no criterion).
    pub fn muGtRaw(self: *const Self, value: f64, threshold: f64, width: f64) f64 {
        return membership.muGreater(value, threshold, width, self.shape);
    }

    /// Fuzzy 'value < threshold' with explicit width (no criterion).
    pub fn muLtRaw(self: *const Self, value: f64, threshold: f64, width: f64) f64 {
        return membership.muLess(value, threshold, width, self.shape);
    }

    /// Raw fuzzy direction ∈ [-1, +1].
    pub fn muDirectionRaw(self: *const Self, o: f64, c: f64, shift: usize) f64 {
        const avg_val = self.avgCS(&self.short_body, shift);
        return membership.muDirection(o, c, avg_val, 2.0);
    }

    /// Fuzzy degree of bullishness ∈ [0, 1].
    pub fn muBullish(self: *const Self, o: f64, c: f64, shift: usize) f64 {
        const d = self.muDirectionRaw(o, c, shift);
        if (d > 0.0) return d;
        return 0.0;
    }

    /// Fuzzy degree of bearishness ∈ [0, 1].
    pub fn muBearish(self: *const Self, o: f64, c: f64, shift: usize) f64 {
        const d = self.muDirectionRaw(o, c, shift);
        if (-d > 0.0) return -d;
        return 0.0;
    }

    // -----------------------------------------------------------------------
    // Pattern methods (delegated to patterns/ files)
    // -----------------------------------------------------------------------

    const patterns = @import("patterns/patterns.zig");

    pub fn abandonedBaby(self: *const Self) f64 {
        return patterns.abandoned_baby.abandonedBaby(self);
    }

    pub fn advanceBlock(self: *const Self) f64 {
        return patterns.advance_block.advanceBlock(self);
    }

    pub fn beltHold(self: *const Self) f64 {
        return patterns.belt_hold.beltHold(self);
    }

    pub fn breakaway(self: *const Self) f64 {
        return patterns.breakaway.breakaway(self);
    }

    pub fn closingMarubozu(self: *const Self) f64 {
        return patterns.closing_marubozu.closingMarubozu(self);
    }

    pub fn concealingBabySwallow(self: *const Self) f64 {
        return patterns.concealing_baby_swallow.concealingBabySwallow(self);
    }

    pub fn counterattack(self: *const Self) f64 {
        return patterns.counterattack.counterattack(self);
    }

    pub fn darkCloudCover(self: *const Self) f64 {
        return patterns.dark_cloud_cover.darkCloudCover(self);
    }

    pub fn patternDoji(self: *const Self) f64 {
        return patterns.doji.patternDoji(self);
    }

    pub fn dojiStar(self: *const Self) f64 {
        return patterns.doji_star.dojiStar(self);
    }

    pub fn dragonflyDoji(self: *const Self) f64 {
        return patterns.dragonfly_doji.dragonflyDoji(self);
    }

    pub fn patternEngulfing(self: *const Self) f64 {
        return patterns.engulfing.patternEngulfing(self);
    }

    pub fn eveningDojiStar(self: *const Self) f64 {
        return patterns.evening_doji_star.eveningDojiStar(self);
    }

    pub fn eveningStar(self: *const Self) f64 {
        return patterns.evening_star.eveningStar(self);
    }

    pub fn gravestoneDoji(self: *const Self) f64 {
        return patterns.gravestone_doji.gravestoneDoji(self);
    }

    pub fn patternHammer(self: *const Self) f64 {
        return patterns.hammer.patternHammer(self);
    }

    pub fn hangingMan(self: *const Self) f64 {
        return patterns.hanging_man.hangingMan(self);
    }

    pub fn patternHarami(self: *const Self) f64 {
        return patterns.harami.patternHarami(self);
    }

    pub fn haramiCross(self: *const Self) f64 {
        return patterns.harami_cross.haramiCross(self);
    }

    pub fn highWave(self: *const Self) f64 {
        return patterns.high_wave.highWave(self);
    }

    pub fn patternHikkake(self: *const Self) f64 {
        return patterns.hikkake.patternHikkake(self);
    }

    pub fn hikkakeModified(self: *const Self) f64 {
        return patterns.hikkake_modified.hikkakeModified(self);
    }

    pub fn homingPigeon(self: *const Self) f64 {
        return patterns.homing_pigeon.homingPigeon(self);
    }

    pub fn identicalThreeCrows(self: *const Self) f64 {
        return patterns.identical_three_crows.identicalThreeCrows(self);
    }

    pub fn inNeck(self: *const Self) f64 {
        return patterns.in_neck.inNeck(self);
    }

    pub fn invertedHammer(self: *const Self) f64 {
        return patterns.inverted_hammer.invertedHammer(self);
    }

    pub fn patternKicking(self: *const Self) f64 {
        return patterns.kicking.patternKicking(self);
    }

    pub fn patternKickingByLength(self: *const Self) f64 {
        return patterns.kicking_by_length.patternKickingByLength(self);
    }

    pub fn ladderBottom(self: *const Self) f64 {
        return patterns.ladder_bottom.ladderBottom(self);
    }

    pub fn longLeggedDoji(self: *const Self) f64 {
        return patterns.long_legged_doji.longLeggedDoji(self);
    }

    pub fn longLine(self: *const Self) f64 {
        return patterns.long_line.longLine(self);
    }

    pub fn patternMarubozu(self: *const Self) f64 {
        return patterns.marubozu.patternMarubozu(self);
    }

    pub fn matchingLow(self: *const Self) f64 {
        return patterns.matching_low.matchingLow(self);
    }

    pub fn matHold(self: *const Self) f64 {
        return patterns.mat_hold.matHold(self);
    }

    pub fn morningDojiStar(self: *const Self) f64 {
        return patterns.morning_doji_star.morningDojiStar(self);
    }

    pub fn morningStar(self: *const Self) f64 {
        return patterns.morning_star.morningStar(self);
    }

    pub fn onNeck(self: *const Self) f64 {
        return patterns.on_neck.onNeck(self);
    }

    pub fn patternPiercing(self: *const Self) f64 {
        return patterns.piercing.patternPiercing(self);
    }

    pub fn rickshawMan(self: *const Self) f64 {
        return patterns.rickshaw_man.rickshawMan(self);
    }

    pub fn risingFallingThreeMethods(self: *const Self) f64 {
        return patterns.rising_falling_three_methods.risingFallingThreeMethods(self);
    }

    pub fn separatingLines(self: *const Self) f64 {
        return patterns.separating_lines.separatingLines(self);
    }

    pub fn shootingStar(self: *const Self) f64 {
        return patterns.shooting_star.shootingStar(self);
    }

    pub fn shortLine(self: *const Self) f64 {
        return patterns.short_line.shortLine(self);
    }

    pub fn spinningTop(self: *const Self) f64 {
        return patterns.spinning_top.spinningTop(self);
    }

    pub fn stalled(self: *const Self) f64 {
        return patterns.stalled.stalled(self);
    }

    pub fn stickSandwich(self: *const Self) f64 {
        return patterns.stick_sandwich.stickSandwich(self);
    }

    pub fn patternTakuri(self: *const Self) f64 {
        return patterns.takuri.patternTakuri(self);
    }

    pub fn tasukiGap(self: *const Self) f64 {
        return patterns.tasuki_gap.tasukiGap(self);
    }

    pub fn threeBlackCrows(self: *const Self) f64 {
        return patterns.three_black_crows.threeBlackCrows(self);
    }

    pub fn threeInside(self: *const Self) f64 {
        return patterns.three_inside.threeInside(self);
    }

    pub fn threeLineStrike(self: *const Self) f64 {
        return patterns.three_line_strike.threeLineStrike(self);
    }

    pub fn threeOutside(self: *const Self) f64 {
        return patterns.three_outside.threeOutside(self);
    }

    pub fn threeStarsInTheSouth(self: *const Self) f64 {
        return patterns.three_stars_in_the_south.threeStarsInTheSouth(self);
    }

    pub fn threeWhiteSoldiers(self: *const Self) f64 {
        return patterns.three_white_soldiers.threeWhiteSoldiers(self);
    }

    pub fn patternThrusting(self: *const Self) f64 {
        return patterns.thrusting.patternThrusting(self);
    }

    pub fn patternTristar(self: *const Self) f64 {
        return patterns.tristar.patternTristar(self);
    }

    pub fn twoCrows(self: *const Self) f64 {
        return patterns.two_crows.twoCrows(self);
    }

    pub fn uniqueThreeRiver(self: *const Self) f64 {
        return patterns.unique_three_river.uniqueThreeRiver(self);
    }

    pub fn upDownGapSideBySideWhiteLines(self: *const Self) f64 {
        return patterns.up_down_gap_side_by_side_white_lines.upDownGapSideBySideWhiteLines(self);
    }

    pub fn upsideGapTwoCrows(self: *const Self) f64 {
        return patterns.upside_gap_two_crows.upsideGapTwoCrows(self);
    }

    pub fn xSideGapThreeMethods(self: *const Self) f64 {
        return patterns.x_side_gap_three_methods.xSideGapThreeMethods(self);
    }




    /// hikkakeModifiedUpdate is called from update() to track stateful hikkake_modified pattern.
    fn hikkakeModifiedUpdate(self: *Self) void {
        if (self.count < 4) return;

        const b1 = self.bar(4);
        const b2 = self.bar(3);
        const b3 = self.bar(2);
        const b4 = self.bar(1);

        // Check for new pattern.
        if (b2.h < b1.h and b2.l > b1.l and
            b3.h < b2.h and b3.l > b2.l)
        {
            const nearAvg = self.avgCS(&self.near, 3);
            // Bullish: 4th breaks low, 2nd close near its low.
            if (b4.h < b3.h and b4.l < b3.l and b2.c <= b2.l + nearAvg) {
                self.hikmod_pattern_result = 100.0;
                self.hikmod_pattern_idx = @intCast(self.count);
                return;
            }
            // Bearish: 4th breaks high, 2nd close near its high.
            if (b4.h > b3.h and b4.l > b3.l and b2.c >= b2.h - nearAvg) {
                self.hikmod_pattern_result = -100.0;
                self.hikmod_pattern_idx = @intCast(self.count);
                return;
            }
        }

        // No new pattern — check for confirmation.
        const count_i: i32 = @intCast(self.count);
        if (self.hikmod_pattern_result != 0.0 and count_i <= self.hikmod_pattern_idx + 3) {
            const shift3rd: usize = self.count - @as(usize, @intCast(self.hikmod_pattern_idx)) + 2;
            const b3rd = self.bar(shift3rd);

            if (self.hikmod_pattern_result > 0 and b4.c > b3rd.h) {
                self.hikmod_last_signal = 200.0;
                self.hikmod_pattern_result = 0.0;
                self.hikmod_pattern_idx = 0;
                self.hikmod_confirmed = true;
                return;
            }
            if (self.hikmod_pattern_result < 0 and b4.c < b3rd.l) {
                self.hikmod_last_signal = -200.0;
                self.hikmod_pattern_result = 0.0;
                self.hikmod_pattern_idx = 0;
                self.hikmod_confirmed = true;
                return;
            }
        }

        // If we passed the 3-bar window, reset.
        if (self.hikmod_pattern_result != 0.0 and count_i > self.hikmod_pattern_idx + 3) {
            self.hikmod_pattern_result = 0.0;
            self.hikmod_pattern_idx = 0;
        }
    }


    // Dispatch

    /// Evaluates a single pattern by its identifier.
    ///
    /// Returns the pattern signal: +100 bullish, -100 bearish, 0 no match.
    pub fn evaluate(self: *const Self, id: PatternIdentifier) f64 {
        return switch (id) {
            .abandoned_baby => self.abandonedBaby(),
            .advance_block => self.advanceBlock(),
            .belt_hold => self.beltHold(),
            .breakaway => self.breakaway(),
            .closing_marubozu => self.closingMarubozu(),
            .concealing_baby_swallow => self.concealingBabySwallow(),
            .counterattack => self.counterattack(),
            .dark_cloud_cover => self.darkCloudCover(),
            .doji => self.patternDoji(),
            .doji_star => self.dojiStar(),
            .dragonfly_doji => self.dragonflyDoji(),
            .engulfing => self.patternEngulfing(),
            .evening_doji_star => self.eveningDojiStar(),
            .evening_star => self.eveningStar(),
            .gravestone_doji => self.gravestoneDoji(),
            .hammer => self.patternHammer(),
            .hanging_man => self.hangingMan(),
            .harami => self.patternHarami(),
            .harami_cross => self.haramiCross(),
            .high_wave => self.highWave(),
            .hikkake => self.patternHikkake(),
            .hikkake_modified => self.hikkakeModified(),
            .homing_pigeon => self.homingPigeon(),
            .identical_three_crows => self.identicalThreeCrows(),
            .in_neck => self.inNeck(),
            .inverted_hammer => self.invertedHammer(),
            .kicking => self.patternKicking(),
            .kicking_by_length => self.patternKickingByLength(),
            .ladder_bottom => self.ladderBottom(),
            .long_legged_doji => self.longLeggedDoji(),
            .long_line => self.longLine(),
            .marubozu => self.patternMarubozu(),
            .matching_low => self.matchingLow(),
            .mat_hold => self.matHold(),
            .morning_doji_star => self.morningDojiStar(),
            .morning_star => self.morningStar(),
            .on_neck => self.onNeck(),
            .piercing => self.patternPiercing(),
            .rickshaw_man => self.rickshawMan(),
            .rising_falling_three_methods => self.risingFallingThreeMethods(),
            .separating_lines => self.separatingLines(),
            .shooting_star => self.shootingStar(),
            .short_line => self.shortLine(),
            .spinning_top => self.spinningTop(),
            .stalled => self.stalled(),
            .stick_sandwich => self.stickSandwich(),
            .takuri => self.patternTakuri(),
            .tasuki_gap => self.tasukiGap(),
            .three_black_crows => self.threeBlackCrows(),
            .three_inside => self.threeInside(),
            .three_line_strike => self.threeLineStrike(),
            .three_outside => self.threeOutside(),
            .three_stars_in_the_south => self.threeStarsInTheSouth(),
            .three_white_soldiers => self.threeWhiteSoldiers(),
            .thrusting => self.patternThrusting(),
            .tristar => self.patternTristar(),
            .two_crows => self.twoCrows(),
            .unique_three_river => self.uniqueThreeRiver(),
            .up_down_gap_side_by_side_white_lines => self.upDownGapSideBySideWhiteLines(),
            .upside_gap_two_crows => self.upsideGapTwoCrows(),
            .x_side_gap_three_methods => self.xSideGapThreeMethods(),
        };
    }
};
