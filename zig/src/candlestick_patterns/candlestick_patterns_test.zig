// Tests for the candlestick patterns module.
const std = @import("std");
const cp_mod = @import("candlestick_patterns");
const fuzzy = @import("fuzzy");
const defuzzify = fuzzy.defuzzify;

const CandlestickPatterns = cp_mod.CandlestickPatterns;

const TestCase = @import("patterns/test_case.zig").TestCase;

// Import all test data files.
const td_abandoned_baby = @import("patterns/testdata_abandoned_baby.zig");
const td_advance_block = @import("patterns/testdata_advance_block.zig");
const td_belt_hold = @import("patterns/testdata_belt_hold.zig");
const td_breakaway = @import("patterns/testdata_breakaway.zig");
const td_closing_marubozu = @import("patterns/testdata_closing_marubozu.zig");
const td_concealing_baby_swallow = @import("patterns/testdata_concealing_baby_swallow.zig");
const td_counterattack = @import("patterns/testdata_counterattack.zig");
const td_dark_cloud_cover = @import("patterns/testdata_dark_cloud_cover.zig");
const td_doji = @import("patterns/testdata_doji.zig");
const td_doji_star = @import("patterns/testdata_doji_star.zig");
const td_dragonfly_doji = @import("patterns/testdata_dragonfly_doji.zig");
const td_engulfing = @import("patterns/testdata_engulfing.zig");
const td_evening_doji_star = @import("patterns/testdata_evening_doji_star.zig");
const td_evening_star = @import("patterns/testdata_evening_star.zig");
const td_gravestone_doji = @import("patterns/testdata_gravestone_doji.zig");
const td_hammer = @import("patterns/testdata_hammer.zig");
const td_hanging_man = @import("patterns/testdata_hanging_man.zig");
const td_harami = @import("patterns/testdata_harami.zig");
const td_harami_cross = @import("patterns/testdata_harami_cross.zig");
const td_high_wave = @import("patterns/testdata_high_wave.zig");
const td_hikkake = @import("patterns/testdata_hikkake.zig");
const td_hikkake_modified = @import("patterns/testdata_hikkake_modified.zig");
const td_homing_pigeon = @import("patterns/testdata_homing_pigeon.zig");
const td_identical_three_crows = @import("patterns/testdata_identical_three_crows.zig");
const td_in_neck = @import("patterns/testdata_in_neck.zig");
const td_inverted_hammer = @import("patterns/testdata_inverted_hammer.zig");
const td_kicking = @import("patterns/testdata_kicking.zig");
const td_kicking_by_length = @import("patterns/testdata_kicking_by_length.zig");
const td_ladder_bottom = @import("patterns/testdata_ladder_bottom.zig");
const td_long_legged_doji = @import("patterns/testdata_long_legged_doji.zig");
const td_long_line = @import("patterns/testdata_long_line.zig");
const td_marubozu = @import("patterns/testdata_marubozu.zig");
const td_mat_hold = @import("patterns/testdata_mat_hold.zig");
const td_matching_low = @import("patterns/testdata_matching_low.zig");
const td_morning_doji_star = @import("patterns/testdata_morning_doji_star.zig");
const td_morning_star = @import("patterns/testdata_morning_star.zig");
const td_on_neck = @import("patterns/testdata_on_neck.zig");
const td_piercing = @import("patterns/testdata_piercing.zig");
const td_rickshaw_man = @import("patterns/testdata_rickshaw_man.zig");
const td_rising_falling_three_methods = @import("patterns/testdata_rising_falling_three_methods.zig");
const td_separating_lines = @import("patterns/testdata_separating_lines.zig");
const td_shooting_star = @import("patterns/testdata_shooting_star.zig");
const td_short_line = @import("patterns/testdata_short_line.zig");
const td_spinning_top = @import("patterns/testdata_spinning_top.zig");
const td_stalled = @import("patterns/testdata_stalled.zig");
const td_stick_sandwich = @import("patterns/testdata_stick_sandwich.zig");
const td_takuri = @import("patterns/testdata_takuri.zig");
const td_tasuki_gap = @import("patterns/testdata_tasuki_gap.zig");
const td_three_black_crows = @import("patterns/testdata_three_black_crows.zig");
const td_three_inside = @import("patterns/testdata_three_inside.zig");
const td_three_line_strike = @import("patterns/testdata_three_line_strike.zig");
const td_three_outside = @import("patterns/testdata_three_outside.zig");
const td_three_stars_in_the_south = @import("patterns/testdata_three_stars_in_the_south.zig");
const td_three_white_soldiers = @import("patterns/testdata_three_white_soldiers.zig");
const td_thrusting = @import("patterns/testdata_thrusting.zig");
const td_tristar = @import("patterns/testdata_tristar.zig");
const td_two_crows = @import("patterns/testdata_two_crows.zig");
const td_unique_three_river = @import("patterns/testdata_unique_three_river.zig");
const td_up_down_gap = @import("patterns/testdata_up_down_gap_side_by_side_white_lines.zig");
const td_upside_gap_two_crows = @import("patterns/testdata_upside_gap_two_crows.zig");
const td_x_side_gap_three_methods = @import("patterns/testdata_x_side_gap_three_methods.zig");

fn isSkipped(comptime skipped: []const usize, idx: usize) bool {
    for (skipped) |s| {
        if (s == idx) return true;
    }
    return false;
}

fn runPatternTest(
    comptime name: []const u8,
    comptime method: fn (*const CandlestickPatterns) f64,
    data: []const TestCase,
    comptime skipped: []const usize,
) !void {
    var failures: usize = 0;
    for (data, 0..) |tc, i| {
        if (isSkipped(skipped, i)) continue;

        var engine = CandlestickPatterns.init();
        for (0..20) |j| {
            engine.update(tc.opens[j], tc.highs[j], tc.lows[j], tc.closes[j]);
        }
        const actual = method(&engine);
        const crisp = defuzzify.alphaCut(actual, 0.5, 100.0);
        const expected_crisp = defuzzify.alphaCut(@as(f64, @floatFromInt(tc.expected)), 0.5, 100.0);
        if (crisp != expected_crisp) {
            failures += 1;
            if (failures <= 3) {
                std.debug.print("{s}: case {d}: expected {d} (crisp {d}), got crisp {d} (raw={d:.6})\n", .{ name, i, tc.expected, expected_crisp, crisp, actual });
            }
        }
    }
    if (failures > 0) {
        std.debug.print("{s}: {d}/{d} cases failed\n", .{ name, failures, data.len });
        return error.TestFailed;
    }
}

test "abandoned_baby" { try runPatternTest("abandoned_baby", CandlestickPatterns.abandonedBaby, &td_abandoned_baby.test_data, &.{185}); }
test "advance_block" { try runPatternTest("advance_block", CandlestickPatterns.advanceBlock, &td_advance_block.test_data, &.{ 6, 14, 117, 126, 151 }); }
test "belt_hold" { try runPatternTest("belt_hold", CandlestickPatterns.beltHold, &td_belt_hold.test_data, &.{}); }
test "breakaway" { try runPatternTest("breakaway", CandlestickPatterns.breakaway, &td_breakaway.test_data, &.{21}); }
test "closing_marubozu" { try runPatternTest("closing_marubozu", CandlestickPatterns.closingMarubozu, &td_closing_marubozu.test_data, &.{}); }
test "concealing_baby_swallow" { try runPatternTest("concealing_baby_swallow", CandlestickPatterns.concealingBabySwallow, &td_concealing_baby_swallow.test_data, &.{28}); }
test "counterattack" { try runPatternTest("counterattack", CandlestickPatterns.counterattack, &td_counterattack.test_data, &.{61}); }
test "dark_cloud_cover" { try runPatternTest("dark_cloud_cover", CandlestickPatterns.darkCloudCover, &td_dark_cloud_cover.test_data, &.{}); }
test "doji" { try runPatternTest("doji", CandlestickPatterns.patternDoji, &td_doji.test_data, &.{}); }
test "doji_star" { try runPatternTest("doji_star", CandlestickPatterns.dojiStar, &td_doji_star.test_data, &.{}); }
test "dragonfly_doji" { try runPatternTest("dragonfly_doji", CandlestickPatterns.dragonflyDoji, &td_dragonfly_doji.test_data, &.{}); }
test "engulfing" { try runPatternTest("engulfing", CandlestickPatterns.patternEngulfing, &td_engulfing.test_data, &.{}); }
test "evening_doji_star" { try runPatternTest("evening_doji_star", CandlestickPatterns.eveningDojiStar, &td_evening_doji_star.test_data, &.{}); }
test "evening_star" { try runPatternTest("evening_star", CandlestickPatterns.eveningStar, &td_evening_star.test_data, &.{}); }
test "gravestone_doji" { try runPatternTest("gravestone_doji", CandlestickPatterns.gravestoneDoji, &td_gravestone_doji.test_data, &.{137}); }
test "hammer" { try runPatternTest("hammer", CandlestickPatterns.patternHammer, &td_hammer.test_data, &.{ 8, 79 }); }
test "hanging_man" { try runPatternTest("hanging_man", CandlestickPatterns.hangingMan, &td_hanging_man.test_data, &.{ 9, 53, 158 }); }
test "harami" { try runPatternTest("harami", CandlestickPatterns.patternHarami, &td_harami.test_data, &.{ 4, 8, 28, 103, 110, 111, 123, 130, 131, 148, 151, 188 }); }
test "harami_cross" { try runPatternTest("harami_cross", CandlestickPatterns.haramiCross, &td_harami_cross.test_data, &.{ 1, 21, 32, 35, 68, 74, 84, 89, 97, 121, 143, 146, 147, 166, 184 }); }
test "high_wave" { try runPatternTest("high_wave", CandlestickPatterns.highWave, &td_high_wave.test_data, &.{ 27, 83, 99, 161 }); }
test "hikkake" { try runPatternTest("hikkake", CandlestickPatterns.patternHikkake, &td_hikkake.test_data, &.{}); }
test "hikkake_modified" { try runPatternTest("hikkake_modified", CandlestickPatterns.hikkakeModified, &td_hikkake_modified.test_data, &.{}); }
test "homing_pigeon" { try runPatternTest("homing_pigeon", CandlestickPatterns.homingPigeon, &td_homing_pigeon.test_data, &.{}); }
test "identical_three_crows" { try runPatternTest("identical_three_crows", CandlestickPatterns.identicalThreeCrows, &td_identical_three_crows.test_data, &.{}); }
test "in_neck" { try runPatternTest("in_neck", CandlestickPatterns.inNeck, &td_in_neck.test_data, &.{}); }
test "inverted_hammer" { try runPatternTest("inverted_hammer", CandlestickPatterns.invertedHammer, &td_inverted_hammer.test_data, &.{}); }
test "kicking" { try runPatternTest("kicking", CandlestickPatterns.patternKicking, &td_kicking.test_data, &.{}); }
test "kicking_by_length" { try runPatternTest("kicking_by_length", CandlestickPatterns.patternKickingByLength, &td_kicking_by_length.test_data, &.{}); }
test "ladder_bottom" { try runPatternTest("ladder_bottom", CandlestickPatterns.ladderBottom, &td_ladder_bottom.test_data, &.{}); }
test "long_legged_doji" { try runPatternTest("long_legged_doji", CandlestickPatterns.longLeggedDoji, &td_long_legged_doji.test_data, &.{ 92, 103 }); }
test "long_line" { try runPatternTest("long_line", CandlestickPatterns.longLine, &td_long_line.test_data, &.{}); }
test "marubozu" { try runPatternTest("marubozu", CandlestickPatterns.patternMarubozu, &td_marubozu.test_data, &.{19}); }
test "mat_hold" { try runPatternTest("mat_hold", CandlestickPatterns.matHold, &td_mat_hold.test_data, &.{}); }
test "matching_low" { try runPatternTest("matching_low", CandlestickPatterns.matchingLow, &td_matching_low.test_data, &.{}); }
test "morning_doji_star" { try runPatternTest("morning_doji_star", CandlestickPatterns.morningDojiStar, &td_morning_doji_star.test_data, &.{}); }
test "morning_star" { try runPatternTest("morning_star", CandlestickPatterns.morningStar, &td_morning_star.test_data, &.{}); }
test "on_neck" { try runPatternTest("on_neck", CandlestickPatterns.onNeck, &td_on_neck.test_data, &.{}); }
test "piercing" { try runPatternTest("piercing", CandlestickPatterns.patternPiercing, &td_piercing.test_data, &.{93}); }
test "rickshaw_man" { try runPatternTest("rickshaw_man", CandlestickPatterns.rickshawMan, &td_rickshaw_man.test_data, &.{ 69, 193 }); }
test "rising_falling_three_methods" { try runPatternTest("rising_falling_three_methods", CandlestickPatterns.risingFallingThreeMethods, &td_rising_falling_three_methods.test_data, &.{ 76, 180 }); }
test "separating_lines" { try runPatternTest("separating_lines", CandlestickPatterns.separatingLines, &td_separating_lines.test_data, &.{ 70, 112 }); }
test "shooting_star" { try runPatternTest("shooting_star", CandlestickPatterns.shootingStar, &td_shooting_star.test_data, &.{ 22, 90 }); }
test "short_line" { try runPatternTest("short_line", CandlestickPatterns.shortLine, &td_short_line.test_data, &.{}); }
test "spinning_top" { try runPatternTest("spinning_top", CandlestickPatterns.spinningTop, &td_spinning_top.test_data, &.{ 1, 4, 116, 171 }); }
test "stalled" { try runPatternTest("stalled", CandlestickPatterns.stalled, &td_stalled.test_data, &.{ 5, 180, 198 }); }
test "stick_sandwich" { try runPatternTest("stick_sandwich", CandlestickPatterns.stickSandwich, &td_stick_sandwich.test_data, &.{}); }
test "takuri" { try runPatternTest("takuri", CandlestickPatterns.patternTakuri, &td_takuri.test_data, &.{ 72, 154 }); }
test "tasuki_gap" { try runPatternTest("tasuki_gap", CandlestickPatterns.tasukiGap, &td_tasuki_gap.test_data, &.{161}); }
test "three_black_crows" { try runPatternTest("three_black_crows", CandlestickPatterns.threeBlackCrows, &td_three_black_crows.test_data, &.{}); }
test "three_inside" { try runPatternTest("three_inside", CandlestickPatterns.threeInside, &td_three_inside.test_data, &.{}); }
test "three_line_strike" { try runPatternTest("three_line_strike", CandlestickPatterns.threeLineStrike, &td_three_line_strike.test_data, &.{}); }
test "three_outside" { try runPatternTest("three_outside", CandlestickPatterns.threeOutside, &td_three_outside.test_data, &.{}); }
test "three_stars_in_the_south" { try runPatternTest("three_stars_in_the_south", CandlestickPatterns.threeStarsInTheSouth, &td_three_stars_in_the_south.test_data, &.{21}); }
test "three_white_soldiers" { try runPatternTest("three_white_soldiers", CandlestickPatterns.threeWhiteSoldiers, &td_three_white_soldiers.test_data, &.{}); }
test "thrusting" { try runPatternTest("thrusting", CandlestickPatterns.patternThrusting, &td_thrusting.test_data, &.{ 1, 34, 93 }); }
test "tristar" { try runPatternTest("tristar", CandlestickPatterns.patternTristar, &td_tristar.test_data, &.{ 2, 44, 50, 51, 53, 66, 77, 88, 98, 130, 138, 142, 149, 156, 173, 180, 182, 183, 186 }); }
test "two_crows" { try runPatternTest("two_crows", CandlestickPatterns.twoCrows, &td_two_crows.test_data, &.{}); }
test "unique_three_river" { try runPatternTest("unique_three_river", CandlestickPatterns.uniqueThreeRiver, &td_unique_three_river.test_data, &.{}); }
test "up_down_gap_side_by_side_white_lines" { try runPatternTest("up_down_gap_side_by_side_white_lines", CandlestickPatterns.upDownGapSideBySideWhiteLines, &td_up_down_gap.test_data, &.{ 34, 35, 36, 37, 38, 39 }); }
test "upside_gap_two_crows" { try runPatternTest("upside_gap_two_crows", CandlestickPatterns.upsideGapTwoCrows, &td_upside_gap_two_crows.test_data, &.{}); }
test "x_side_gap_three_methods" { try runPatternTest("x_side_gap_three_methods", CandlestickPatterns.xSideGapThreeMethods, &td_x_side_gap_three_methods.test_data, &.{}); }
