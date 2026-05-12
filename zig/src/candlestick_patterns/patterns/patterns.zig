//! Barrel re-export for all 60 standalone pattern functions.
//
//! Each pattern is implemented in its own file as a standalone pub fn
//! taking *const CandlestickPatterns. The CandlestickPatterns struct
//! delegates to these via one-liner methods.

pub const abandoned_baby = @import("abandoned_baby.zig");
pub const advance_block = @import("advance_block.zig");
pub const belt_hold = @import("belt_hold.zig");
pub const breakaway = @import("breakaway.zig");
pub const closing_marubozu = @import("closing_marubozu.zig");
pub const concealing_baby_swallow = @import("concealing_baby_swallow.zig");
pub const counterattack = @import("counterattack.zig");
pub const dark_cloud_cover = @import("dark_cloud_cover.zig");
pub const doji = @import("doji.zig");
pub const doji_star = @import("doji_star.zig");
pub const dragonfly_doji = @import("dragonfly_doji.zig");
pub const engulfing = @import("engulfing.zig");
pub const evening_doji_star = @import("evening_doji_star.zig");
pub const evening_star = @import("evening_star.zig");
pub const gravestone_doji = @import("gravestone_doji.zig");
pub const hammer = @import("hammer.zig");
pub const hanging_man = @import("hanging_man.zig");
pub const harami = @import("harami.zig");
pub const harami_cross = @import("harami_cross.zig");
pub const high_wave = @import("high_wave.zig");
pub const hikkake = @import("hikkake.zig");
pub const hikkake_modified = @import("hikkake_modified.zig");
pub const homing_pigeon = @import("homing_pigeon.zig");
pub const identical_three_crows = @import("identical_three_crows.zig");
pub const in_neck = @import("in_neck.zig");
pub const inverted_hammer = @import("inverted_hammer.zig");
pub const kicking = @import("kicking.zig");
pub const kicking_by_length = @import("kicking_by_length.zig");
pub const ladder_bottom = @import("ladder_bottom.zig");
pub const long_legged_doji = @import("long_legged_doji.zig");
pub const long_line = @import("long_line.zig");
pub const marubozu = @import("marubozu.zig");
pub const matching_low = @import("matching_low.zig");
pub const mat_hold = @import("mat_hold.zig");
pub const morning_doji_star = @import("morning_doji_star.zig");
pub const morning_star = @import("morning_star.zig");
pub const on_neck = @import("on_neck.zig");
pub const piercing = @import("piercing.zig");
pub const rickshaw_man = @import("rickshaw_man.zig");
pub const rising_falling_three_methods = @import("rising_falling_three_methods.zig");
pub const separating_lines = @import("separating_lines.zig");
pub const shooting_star = @import("shooting_star.zig");
pub const short_line = @import("short_line.zig");
pub const spinning_top = @import("spinning_top.zig");
pub const stalled = @import("stalled.zig");
pub const stick_sandwich = @import("stick_sandwich.zig");
pub const takuri = @import("takuri.zig");
pub const tasuki_gap = @import("tasuki_gap.zig");
pub const three_black_crows = @import("three_black_crows.zig");
pub const three_inside = @import("three_inside.zig");
pub const three_line_strike = @import("three_line_strike.zig");
pub const three_outside = @import("three_outside.zig");
pub const three_stars_in_the_south = @import("three_stars_in_the_south.zig");
pub const three_white_soldiers = @import("three_white_soldiers.zig");
pub const thrusting = @import("thrusting.zig");
pub const tristar = @import("tristar.zig");
pub const two_crows = @import("two_crows.zig");
pub const unique_three_river = @import("unique_three_river.zig");
pub const up_down_gap_side_by_side_white_lines = @import("up_down_gap_side_by_side_white_lines.zig");
pub const upside_gap_two_crows = @import("upside_gap_two_crows.zig");
pub const x_side_gap_three_methods = @import("x_side_gap_three_methods.zig");
