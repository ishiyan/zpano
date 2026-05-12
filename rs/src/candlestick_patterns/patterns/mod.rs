pub mod abandoned_baby;
pub mod advance_block;
pub mod belt_hold;
pub mod breakaway;
pub mod closing_marubozu;
pub mod concealing_baby_swallow;
pub mod counterattack;
pub mod dark_cloud_cover;
pub mod doji;
pub mod doji_star;
pub mod dragonfly_doji;
pub mod engulfing;
pub mod evening_doji_star;
pub mod evening_star;
pub mod gravestone_doji;
pub mod hammer;
pub mod hanging_man;
pub mod harami;
pub mod harami_cross;
pub mod high_wave;
pub mod hikkake;
pub mod hikkake_modified;
pub mod homing_pigeon;
pub mod identical_three_crows;
pub mod in_neck;
pub mod inverted_hammer;
pub mod kicking;
pub mod kicking_by_length;
pub mod ladder_bottom;
pub mod long_legged_doji;
pub mod long_line;
pub mod marubozu;
pub mod matching_low;
pub mod mat_hold;
pub mod morning_doji_star;
pub mod morning_star;
pub mod on_neck;
pub mod piercing;
pub mod rickshaw_man;
pub mod rising_falling_three_methods;
pub mod separating_lines;
pub mod shooting_star;
pub mod short_line;
pub mod spinning_top;
pub mod stalled;
pub mod stick_sandwich;
pub mod takuri;
pub mod tasuki_gap;
pub mod three_black_crows;
pub mod three_inside;
pub mod three_line_strike;
pub mod three_outside;
pub mod three_stars_in_the_south;
pub mod three_white_soldiers;
pub mod thrusting;
pub mod tristar;
pub mod two_crows;
pub mod unique_three_river;
pub mod up_down_gap_side_by_side_white_lines;
pub mod upside_gap_two_crows;
pub mod x_side_gap_three_methods;

// Test data lives alongside pattern implementations.
#[cfg(test)]
#[derive(Debug)]
pub struct TestCase {
    pub opens: [f64; 20],
    pub highs: [f64; 20],
    pub lows: [f64; 20],
    pub closes: [f64; 20],
    pub expected: i32,
}

#[cfg(test)]
mod testdata_abandoned_baby;
#[cfg(test)]
mod testdata_advance_block;
#[cfg(test)]
mod testdata_belt_hold;
#[cfg(test)]
mod testdata_breakaway;
#[cfg(test)]
mod testdata_closing_marubozu;
#[cfg(test)]
mod testdata_concealing_baby_swallow;
#[cfg(test)]
mod testdata_counterattack;
#[cfg(test)]
mod testdata_dark_cloud_cover;
#[cfg(test)]
mod testdata_doji;
#[cfg(test)]
mod testdata_doji_star;
#[cfg(test)]
mod testdata_dragonfly_doji;
#[cfg(test)]
mod testdata_engulfing;
#[cfg(test)]
mod testdata_evening_doji_star;
#[cfg(test)]
mod testdata_evening_star;
#[cfg(test)]
mod testdata_gravestone_doji;
#[cfg(test)]
mod testdata_hammer;
#[cfg(test)]
mod testdata_hanging_man;
#[cfg(test)]
mod testdata_harami;
#[cfg(test)]
mod testdata_harami_cross;
#[cfg(test)]
mod testdata_high_wave;
#[cfg(test)]
mod testdata_hikkake;
#[cfg(test)]
mod testdata_hikkake_modified;
#[cfg(test)]
mod testdata_homing_pigeon;
#[cfg(test)]
mod testdata_identical_three_crows;
#[cfg(test)]
mod testdata_in_neck;
#[cfg(test)]
mod testdata_inverted_hammer;
#[cfg(test)]
mod testdata_kicking;
#[cfg(test)]
mod testdata_kicking_by_length;
#[cfg(test)]
mod testdata_ladder_bottom;
#[cfg(test)]
mod testdata_long_legged_doji;
#[cfg(test)]
mod testdata_long_line;
#[cfg(test)]
mod testdata_marubozu;
#[cfg(test)]
mod testdata_mat_hold;
#[cfg(test)]
mod testdata_matching_low;
#[cfg(test)]
mod testdata_morning_doji_star;
#[cfg(test)]
mod testdata_morning_star;
#[cfg(test)]
mod testdata_on_neck;
#[cfg(test)]
mod testdata_piercing;
#[cfg(test)]
mod testdata_rickshaw_man;
#[cfg(test)]
mod testdata_rising_falling_three_methods;
#[cfg(test)]
mod testdata_separating_lines;
#[cfg(test)]
mod testdata_shooting_star;
#[cfg(test)]
mod testdata_short_line;
#[cfg(test)]
mod testdata_spinning_top;
#[cfg(test)]
mod testdata_stalled;
#[cfg(test)]
mod testdata_stick_sandwich;
#[cfg(test)]
mod testdata_takuri;
#[cfg(test)]
mod testdata_tasuki_gap;
#[cfg(test)]
mod testdata_three_black_crows;
#[cfg(test)]
mod testdata_three_inside;
#[cfg(test)]
mod testdata_three_line_strike;
#[cfg(test)]
mod testdata_three_outside;
#[cfg(test)]
mod testdata_three_stars_in_the_south;
#[cfg(test)]
mod testdata_three_white_soldiers;
#[cfg(test)]
mod testdata_thrusting;
#[cfg(test)]
mod testdata_tristar;
#[cfg(test)]
mod testdata_two_crows;
#[cfg(test)]
mod testdata_unique_three_river;
#[cfg(test)]
mod testdata_up_down_gap_side_by_side_white_lines;
#[cfg(test)]
mod testdata_upside_gap_two_crows;
#[cfg(test)]
mod testdata_x_side_gap_three_methods;

#[cfg(test)]
pub use testdata_abandoned_baby::TEST_DATA_ABANDONED_BABY;
#[cfg(test)]
pub use testdata_advance_block::TEST_DATA_ADVANCE_BLOCK;
#[cfg(test)]
pub use testdata_belt_hold::TEST_DATA_BELT_HOLD;
#[cfg(test)]
pub use testdata_breakaway::TEST_DATA_BREAKAWAY;
#[cfg(test)]
pub use testdata_closing_marubozu::TEST_DATA_CLOSING_MARUBOZU;
#[cfg(test)]
pub use testdata_concealing_baby_swallow::TEST_DATA_CONCEALING_BABY_SWALLOW;
#[cfg(test)]
pub use testdata_counterattack::TEST_DATA_COUNTERATTACK;
#[cfg(test)]
pub use testdata_dark_cloud_cover::TEST_DATA_DARK_CLOUD_COVER;
#[cfg(test)]
pub use testdata_doji::TEST_DATA_DOJI;
#[cfg(test)]
pub use testdata_doji_star::TEST_DATA_DOJI_STAR;
#[cfg(test)]
pub use testdata_dragonfly_doji::TEST_DATA_DRAGONFLY_DOJI;
#[cfg(test)]
pub use testdata_engulfing::TEST_DATA_ENGULFING;
#[cfg(test)]
pub use testdata_evening_doji_star::TEST_DATA_EVENING_DOJI_STAR;
#[cfg(test)]
pub use testdata_evening_star::TEST_DATA_EVENING_STAR;
#[cfg(test)]
pub use testdata_gravestone_doji::TEST_DATA_GRAVESTONE_DOJI;
#[cfg(test)]
pub use testdata_hammer::TEST_DATA_HAMMER;
#[cfg(test)]
pub use testdata_hanging_man::TEST_DATA_HANGING_MAN;
#[cfg(test)]
pub use testdata_harami::TEST_DATA_HARAMI;
#[cfg(test)]
pub use testdata_harami_cross::TEST_DATA_HARAMI_CROSS;
#[cfg(test)]
pub use testdata_high_wave::TEST_DATA_HIGH_WAVE;
#[cfg(test)]
pub use testdata_hikkake::TEST_DATA_HIKKAKE;
#[cfg(test)]
pub use testdata_hikkake_modified::TEST_DATA_HIKKAKE_MODIFIED;
#[cfg(test)]
pub use testdata_homing_pigeon::TEST_DATA_HOMING_PIGEON;
#[cfg(test)]
pub use testdata_identical_three_crows::TEST_DATA_IDENTICAL_THREE_CROWS;
#[cfg(test)]
pub use testdata_in_neck::TEST_DATA_IN_NECK;
#[cfg(test)]
pub use testdata_inverted_hammer::TEST_DATA_INVERTED_HAMMER;
#[cfg(test)]
pub use testdata_kicking::TEST_DATA_KICKING;
#[cfg(test)]
pub use testdata_kicking_by_length::TEST_DATA_KICKING_BY_LENGTH;
#[cfg(test)]
pub use testdata_ladder_bottom::TEST_DATA_LADDER_BOTTOM;
#[cfg(test)]
pub use testdata_long_legged_doji::TEST_DATA_LONG_LEGGED_DOJI;
#[cfg(test)]
pub use testdata_long_line::TEST_DATA_LONG_LINE;
#[cfg(test)]
pub use testdata_marubozu::TEST_DATA_MARUBOZU;
#[cfg(test)]
pub use testdata_mat_hold::TEST_DATA_MAT_HOLD;
#[cfg(test)]
pub use testdata_matching_low::TEST_DATA_MATCHING_LOW;
#[cfg(test)]
pub use testdata_morning_doji_star::TEST_DATA_MORNING_DOJI_STAR;
#[cfg(test)]
pub use testdata_morning_star::TEST_DATA_MORNING_STAR;
#[cfg(test)]
pub use testdata_on_neck::TEST_DATA_ON_NECK;
#[cfg(test)]
pub use testdata_piercing::TEST_DATA_PIERCING;
#[cfg(test)]
pub use testdata_rickshaw_man::TEST_DATA_RICKSHAW_MAN;
#[cfg(test)]
pub use testdata_rising_falling_three_methods::TEST_DATA_RISING_FALLING_THREE_METHODS;
#[cfg(test)]
pub use testdata_separating_lines::TEST_DATA_SEPARATING_LINES;
#[cfg(test)]
pub use testdata_shooting_star::TEST_DATA_SHOOTING_STAR;
#[cfg(test)]
pub use testdata_short_line::TEST_DATA_SHORT_LINE;
#[cfg(test)]
pub use testdata_spinning_top::TEST_DATA_SPINNING_TOP;
#[cfg(test)]
pub use testdata_stalled::TEST_DATA_STALLED;
#[cfg(test)]
pub use testdata_stick_sandwich::TEST_DATA_STICK_SANDWICH;
#[cfg(test)]
pub use testdata_takuri::TEST_DATA_TAKURI;
#[cfg(test)]
pub use testdata_tasuki_gap::TEST_DATA_TASUKI_GAP;
#[cfg(test)]
pub use testdata_three_black_crows::TEST_DATA_THREE_BLACK_CROWS;
#[cfg(test)]
pub use testdata_three_inside::TEST_DATA_THREE_INSIDE;
#[cfg(test)]
pub use testdata_three_line_strike::TEST_DATA_THREE_LINE_STRIKE;
#[cfg(test)]
pub use testdata_three_outside::TEST_DATA_THREE_OUTSIDE;
#[cfg(test)]
pub use testdata_three_stars_in_the_south::TEST_DATA_THREE_STARS_IN_THE_SOUTH;
#[cfg(test)]
pub use testdata_three_white_soldiers::TEST_DATA_THREE_WHITE_SOLDIERS;
#[cfg(test)]
pub use testdata_thrusting::TEST_DATA_THRUSTING;
#[cfg(test)]
pub use testdata_tristar::TEST_DATA_TRISTAR;
#[cfg(test)]
pub use testdata_two_crows::TEST_DATA_TWO_CROWS;
#[cfg(test)]
pub use testdata_unique_three_river::TEST_DATA_UNIQUE_THREE_RIVER;
#[cfg(test)]
pub use testdata_up_down_gap_side_by_side_white_lines::TEST_DATA_UP_DOWN_GAP_SIDE_BY_SIDE_WHITE_LINES;
#[cfg(test)]
pub use testdata_upside_gap_two_crows::TEST_DATA_UPSIDE_GAP_TWO_CROWS;
#[cfg(test)]
pub use testdata_x_side_gap_three_methods::TEST_DATA_X_SIDE_GAP_THREE_METHODS;
