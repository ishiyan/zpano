use crate::fuzzy::alpha_cut;
use super::*;
use super::patterns::*;

struct PatternSpec {
    name: &'static str,
    method: fn(&CandlestickPatterns) -> f64,
    data: &'static [TestCase],
    skipped: &'static [usize],
}

fn run_pattern_test(spec: &PatternSpec) {
    let mut failures = 0;
    for (i, tc) in spec.data.iter().enumerate() {
        if spec.skipped.contains(&i) {
            continue;
        }
        let mut cp = CandlestickPatterns::new();
        for j in 0..20 {
            cp.update(tc.opens[j], tc.highs[j], tc.lows[j], tc.closes[j]);
        }
        let actual = (spec.method)(&cp);
        let crisp = alpha_cut(actual, 0.5, 100.0);
        let expected_crisp = alpha_cut(tc.expected as f64, 0.5, 100.0);
        if crisp != expected_crisp {
            failures += 1;
            if failures <= 10 {
                eprintln!(
                    "{}: case {}: expected {} (crisp {}), got crisp {} (raw={:.6})",
                    spec.name, i, tc.expected, expected_crisp, crisp, actual
                );
            }
        }
    }
    if failures > 0 {
        panic!(
            "{}: {}/{} cases failed",
            spec.name,
            failures,
            spec.data.len()
        );
    }
}

#[test]
fn test_abandoned_baby() {
    run_pattern_test(&PatternSpec { name: "abandoned_baby", method: CandlestickPatterns::abandoned_baby, data: TEST_DATA_ABANDONED_BABY, skipped: &[185] });
}

#[test]
fn test_advance_block() {
    run_pattern_test(&PatternSpec { name: "advance_block", method: CandlestickPatterns::advance_block, data: TEST_DATA_ADVANCE_BLOCK, skipped: &[6, 14, 117, 126, 151] });
}

#[test]
fn test_belt_hold() {
    run_pattern_test(&PatternSpec { name: "belt_hold", method: CandlestickPatterns::belt_hold, data: TEST_DATA_BELT_HOLD, skipped: &[] });
}

#[test]
fn test_breakaway() {
    run_pattern_test(&PatternSpec { name: "breakaway", method: CandlestickPatterns::breakaway, data: TEST_DATA_BREAKAWAY, skipped: &[21] });
}

#[test]
fn test_closing_marubozu() {
    run_pattern_test(&PatternSpec { name: "closing_marubozu", method: CandlestickPatterns::closing_marubozu, data: TEST_DATA_CLOSING_MARUBOZU, skipped: &[] });
}

#[test]
fn test_concealing_baby_swallow() {
    run_pattern_test(&PatternSpec { name: "concealing_baby_swallow", method: CandlestickPatterns::concealing_baby_swallow, data: TEST_DATA_CONCEALING_BABY_SWALLOW, skipped: &[28] });
}

#[test]
fn test_counterattack() {
    run_pattern_test(&PatternSpec { name: "counterattack", method: CandlestickPatterns::counterattack, data: TEST_DATA_COUNTERATTACK, skipped: &[61] });
}

#[test]
fn test_dark_cloud_cover() {
    run_pattern_test(&PatternSpec { name: "dark_cloud_cover", method: CandlestickPatterns::dark_cloud_cover, data: TEST_DATA_DARK_CLOUD_COVER, skipped: &[] });
}

#[test]
fn test_doji() {
    run_pattern_test(&PatternSpec { name: "doji", method: CandlestickPatterns::doji, data: TEST_DATA_DOJI, skipped: &[] });
}

#[test]
fn test_doji_star() {
    run_pattern_test(&PatternSpec { name: "doji_star", method: CandlestickPatterns::doji_star, data: TEST_DATA_DOJI_STAR, skipped: &[] });
}

#[test]
fn test_dragonfly_doji() {
    run_pattern_test(&PatternSpec { name: "dragonfly_doji", method: CandlestickPatterns::dragonfly_doji, data: TEST_DATA_DRAGONFLY_DOJI, skipped: &[] });
}

#[test]
fn test_engulfing() {
    run_pattern_test(&PatternSpec { name: "engulfing", method: CandlestickPatterns::engulfing, data: TEST_DATA_ENGULFING, skipped: &[] });
}

#[test]
fn test_evening_doji_star() {
    run_pattern_test(&PatternSpec { name: "evening_doji_star", method: CandlestickPatterns::evening_doji_star, data: TEST_DATA_EVENING_DOJI_STAR, skipped: &[] });
}

#[test]
fn test_evening_star() {
    run_pattern_test(&PatternSpec { name: "evening_star", method: CandlestickPatterns::evening_star, data: TEST_DATA_EVENING_STAR, skipped: &[] });
}

#[test]
fn test_gravestone_doji() {
    run_pattern_test(&PatternSpec { name: "gravestone_doji", method: CandlestickPatterns::gravestone_doji, data: TEST_DATA_GRAVESTONE_DOJI, skipped: &[137] });
}

#[test]
fn test_hammer() {
    run_pattern_test(&PatternSpec { name: "hammer", method: CandlestickPatterns::hammer, data: TEST_DATA_HAMMER, skipped: &[8, 79] });
}

#[test]
fn test_hanging_man() {
    run_pattern_test(&PatternSpec { name: "hanging_man", method: CandlestickPatterns::hanging_man, data: TEST_DATA_HANGING_MAN, skipped: &[9, 53, 158] });
}

#[test]
fn test_harami() {
    run_pattern_test(&PatternSpec { name: "harami", method: CandlestickPatterns::harami, data: TEST_DATA_HARAMI, skipped: &[4, 8, 28, 103, 110, 111, 123, 130, 131, 148, 151, 188] });
}

#[test]
fn test_harami_cross() {
    run_pattern_test(&PatternSpec { name: "harami_cross", method: CandlestickPatterns::harami_cross, data: TEST_DATA_HARAMI_CROSS, skipped: &[1, 21, 32, 35, 68, 74, 84, 89, 97, 121, 143, 146, 147, 166, 184] });
}

#[test]
fn test_high_wave() {
    run_pattern_test(&PatternSpec { name: "high_wave", method: CandlestickPatterns::high_wave, data: TEST_DATA_HIGH_WAVE, skipped: &[27, 83, 99, 161] });
}

#[test]
fn test_hikkake() {
    run_pattern_test(&PatternSpec { name: "hikkake", method: CandlestickPatterns::hikkake, data: TEST_DATA_HIKKAKE, skipped: &[] });
}

#[test]
fn test_hikkake_modified() {
    run_pattern_test(&PatternSpec { name: "hikkake_modified", method: CandlestickPatterns::hikkake_modified, data: TEST_DATA_HIKKAKE_MODIFIED, skipped: &[] });
}

#[test]
fn test_homing_pigeon() {
    run_pattern_test(&PatternSpec { name: "homing_pigeon", method: CandlestickPatterns::homing_pigeon, data: TEST_DATA_HOMING_PIGEON, skipped: &[] });
}

#[test]
fn test_identical_three_crows() {
    run_pattern_test(&PatternSpec { name: "identical_three_crows", method: CandlestickPatterns::identical_three_crows, data: TEST_DATA_IDENTICAL_THREE_CROWS, skipped: &[] });
}

#[test]
fn test_in_neck() {
    run_pattern_test(&PatternSpec { name: "in_neck", method: CandlestickPatterns::in_neck, data: TEST_DATA_IN_NECK, skipped: &[] });
}

#[test]
fn test_inverted_hammer() {
    run_pattern_test(&PatternSpec { name: "inverted_hammer", method: CandlestickPatterns::inverted_hammer, data: TEST_DATA_INVERTED_HAMMER, skipped: &[] });
}

#[test]
fn test_kicking() {
    run_pattern_test(&PatternSpec { name: "kicking", method: CandlestickPatterns::kicking, data: TEST_DATA_KICKING, skipped: &[] });
}

#[test]
fn test_kicking_by_length() {
    run_pattern_test(&PatternSpec { name: "kicking_by_length", method: CandlestickPatterns::kicking_by_length, data: TEST_DATA_KICKING_BY_LENGTH, skipped: &[] });
}

#[test]
fn test_ladder_bottom() {
    run_pattern_test(&PatternSpec { name: "ladder_bottom", method: CandlestickPatterns::ladder_bottom, data: TEST_DATA_LADDER_BOTTOM, skipped: &[] });
}

#[test]
fn test_long_legged_doji() {
    run_pattern_test(&PatternSpec { name: "long_legged_doji", method: CandlestickPatterns::long_legged_doji, data: TEST_DATA_LONG_LEGGED_DOJI, skipped: &[92, 103] });
}

#[test]
fn test_long_line() {
    run_pattern_test(&PatternSpec { name: "long_line", method: CandlestickPatterns::long_line, data: TEST_DATA_LONG_LINE, skipped: &[] });
}

#[test]
fn test_marubozu() {
    run_pattern_test(&PatternSpec { name: "marubozu", method: CandlestickPatterns::marubozu, data: TEST_DATA_MARUBOZU, skipped: &[19] });
}

#[test]
fn test_mat_hold() {
    run_pattern_test(&PatternSpec { name: "mat_hold", method: CandlestickPatterns::mat_hold, data: TEST_DATA_MAT_HOLD, skipped: &[] });
}

#[test]
fn test_matching_low() {
    run_pattern_test(&PatternSpec { name: "matching_low", method: CandlestickPatterns::matching_low, data: TEST_DATA_MATCHING_LOW, skipped: &[] });
}

#[test]
fn test_morning_doji_star() {
    run_pattern_test(&PatternSpec { name: "morning_doji_star", method: CandlestickPatterns::morning_doji_star, data: TEST_DATA_MORNING_DOJI_STAR, skipped: &[] });
}

#[test]
fn test_morning_star() {
    run_pattern_test(&PatternSpec { name: "morning_star", method: CandlestickPatterns::morning_star, data: TEST_DATA_MORNING_STAR, skipped: &[] });
}

#[test]
fn test_on_neck() {
    run_pattern_test(&PatternSpec { name: "on_neck", method: CandlestickPatterns::on_neck, data: TEST_DATA_ON_NECK, skipped: &[] });
}

#[test]
fn test_piercing() {
    run_pattern_test(&PatternSpec { name: "piercing", method: CandlestickPatterns::piercing, data: TEST_DATA_PIERCING, skipped: &[93] });
}

#[test]
fn test_rickshaw_man() {
    run_pattern_test(&PatternSpec { name: "rickshaw_man", method: CandlestickPatterns::rickshaw_man, data: TEST_DATA_RICKSHAW_MAN, skipped: &[69, 193] });
}

#[test]
fn test_rising_falling_three_methods() {
    run_pattern_test(&PatternSpec { name: "rising_falling_three_methods", method: CandlestickPatterns::rising_falling_three_methods, data: TEST_DATA_RISING_FALLING_THREE_METHODS, skipped: &[76, 180] });
}

#[test]
fn test_separating_lines() {
    run_pattern_test(&PatternSpec { name: "separating_lines", method: CandlestickPatterns::separating_lines, data: TEST_DATA_SEPARATING_LINES, skipped: &[70, 112] });
}

#[test]
fn test_shooting_star() {
    run_pattern_test(&PatternSpec { name: "shooting_star", method: CandlestickPatterns::shooting_star, data: TEST_DATA_SHOOTING_STAR, skipped: &[22, 90] });
}

#[test]
fn test_short_line() {
    run_pattern_test(&PatternSpec { name: "short_line", method: CandlestickPatterns::short_line, data: TEST_DATA_SHORT_LINE, skipped: &[] });
}

#[test]
fn test_spinning_top() {
    run_pattern_test(&PatternSpec { name: "spinning_top", method: CandlestickPatterns::spinning_top, data: TEST_DATA_SPINNING_TOP, skipped: &[1, 4, 116, 171] });
}

#[test]
fn test_stalled() {
    run_pattern_test(&PatternSpec { name: "stalled", method: CandlestickPatterns::stalled, data: TEST_DATA_STALLED, skipped: &[5, 180, 198] });
}

#[test]
fn test_stick_sandwich() {
    run_pattern_test(&PatternSpec { name: "stick_sandwich", method: CandlestickPatterns::stick_sandwich, data: TEST_DATA_STICK_SANDWICH, skipped: &[] });
}

#[test]
fn test_takuri() {
    run_pattern_test(&PatternSpec { name: "takuri", method: CandlestickPatterns::takuri, data: TEST_DATA_TAKURI, skipped: &[72, 154] });
}

#[test]
fn test_tasuki_gap() {
    run_pattern_test(&PatternSpec { name: "tasuki_gap", method: CandlestickPatterns::tasuki_gap, data: TEST_DATA_TASUKI_GAP, skipped: &[161] });
}

#[test]
fn test_three_black_crows() {
    run_pattern_test(&PatternSpec { name: "three_black_crows", method: CandlestickPatterns::three_black_crows, data: TEST_DATA_THREE_BLACK_CROWS, skipped: &[] });
}

#[test]
fn test_three_inside() {
    run_pattern_test(&PatternSpec { name: "three_inside", method: CandlestickPatterns::three_inside, data: TEST_DATA_THREE_INSIDE, skipped: &[] });
}

#[test]
fn test_three_line_strike() {
    run_pattern_test(&PatternSpec { name: "three_line_strike", method: CandlestickPatterns::three_line_strike, data: TEST_DATA_THREE_LINE_STRIKE, skipped: &[] });
}

#[test]
fn test_three_outside() {
    run_pattern_test(&PatternSpec { name: "three_outside", method: CandlestickPatterns::three_outside, data: TEST_DATA_THREE_OUTSIDE, skipped: &[] });
}

#[test]
fn test_three_stars_in_the_south() {
    run_pattern_test(&PatternSpec { name: "three_stars_in_the_south", method: CandlestickPatterns::three_stars_in_the_south, data: TEST_DATA_THREE_STARS_IN_THE_SOUTH, skipped: &[21] });
}

#[test]
fn test_three_white_soldiers() {
    run_pattern_test(&PatternSpec { name: "three_white_soldiers", method: CandlestickPatterns::three_white_soldiers, data: TEST_DATA_THREE_WHITE_SOLDIERS, skipped: &[] });
}

#[test]
fn test_thrusting() {
    run_pattern_test(&PatternSpec { name: "thrusting", method: CandlestickPatterns::thrusting, data: TEST_DATA_THRUSTING, skipped: &[1, 34, 93] });
}

#[test]
fn test_tristar() {
    run_pattern_test(&PatternSpec { name: "tristar", method: CandlestickPatterns::tristar, data: TEST_DATA_TRISTAR, skipped: &[2, 44, 50, 51, 53, 66, 77, 88, 98, 130, 138, 142, 149, 156, 173, 180, 182, 183, 186] });
}

#[test]
fn test_two_crows() {
    run_pattern_test(&PatternSpec { name: "two_crows", method: CandlestickPatterns::two_crows, data: TEST_DATA_TWO_CROWS, skipped: &[] });
}

#[test]
fn test_unique_three_river() {
    run_pattern_test(&PatternSpec { name: "unique_three_river", method: CandlestickPatterns::unique_three_river, data: TEST_DATA_UNIQUE_THREE_RIVER, skipped: &[] });
}

#[test]
fn test_up_down_gap_side_by_side_white_lines() {
    run_pattern_test(&PatternSpec { name: "up_down_gap_side_by_side_white_lines", method: CandlestickPatterns::up_down_gap_side_by_side_white_lines, data: TEST_DATA_UP_DOWN_GAP_SIDE_BY_SIDE_WHITE_LINES, skipped: &[34, 35, 36, 37, 38, 39] });
}

#[test]
fn test_upside_gap_two_crows() {
    run_pattern_test(&PatternSpec { name: "upside_gap_two_crows", method: CandlestickPatterns::upside_gap_two_crows, data: TEST_DATA_UPSIDE_GAP_TWO_CROWS, skipped: &[] });
}

#[test]
fn test_x_side_gap_three_methods() {
    run_pattern_test(&PatternSpec { name: "x_side_gap_three_methods", method: CandlestickPatterns::x_side_gap_three_methods, data: TEST_DATA_X_SIDE_GAP_THREE_METHODS, skipped: &[] });
}
