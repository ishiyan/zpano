"""Tests for all 61 candlestick patterns against TA-Lib reference data.

Usage:
    python -m unittest py.candlestick_patterns.patterns.test_patterns -v
"""
from __future__ import annotations

import unittest

from ..candlestick_patterns import CandlestickPatterns

from .test_data_abandoned_baby import TEST_DATA_ABANDONED_BABY
from .test_data_advance_block import TEST_DATA_ADVANCE_BLOCK
from .test_data_belt_hold import TEST_DATA_BELT_HOLD
from .test_data_breakaway import TEST_DATA_BREAKAWAY
from .test_data_closing_marubozu import TEST_DATA_CLOSING_MARUBOZU
from .test_data_concealing_baby_swallow import TEST_DATA_CONCEALING_BABY_SWALLOW
from .test_data_counterattack import TEST_DATA_COUNTERATTACK
from .test_data_dark_cloud_cover import TEST_DATA_DARK_CLOUD_COVER
from .test_data_doji import TEST_DATA_DOJI
from .test_data_doji_star import TEST_DATA_DOJI_STAR
from .test_data_dragonfly_doji import TEST_DATA_DRAGONFLY_DOJI
from .test_data_engulfing import TEST_DATA_ENGULFING
from .test_data_evening_doji_star import TEST_DATA_EVENING_DOJI_STAR
from .test_data_evening_star import TEST_DATA_EVENING_STAR
from .test_data_gravestone_doji import TEST_DATA_GRAVESTONE_DOJI
from .test_data_hammer import TEST_DATA_HAMMER
from .test_data_hanging_man import TEST_DATA_HANGING_MAN
from .test_data_harami import TEST_DATA_HARAMI
from .test_data_harami_cross import TEST_DATA_HARAMI_CROSS
from .test_data_high_wave import TEST_DATA_HIGH_WAVE
from .test_data_hikkake import TEST_DATA_HIKKAKE
from .test_data_hikkake_modified import TEST_DATA_HIKKAKE_MODIFIED
from .test_data_homing_pigeon import TEST_DATA_HOMING_PIGEON
from .test_data_identical_three_crows import TEST_DATA_IDENTICAL_THREE_CROWS
from .test_data_in_neck import TEST_DATA_IN_NECK
from .test_data_inverted_hammer import TEST_DATA_INVERTED_HAMMER
from .test_data_kicking import TEST_DATA_KICKING
from .test_data_kicking_by_length import TEST_DATA_KICKING_BY_LENGTH
from .test_data_ladder_bottom import TEST_DATA_LADDER_BOTTOM
from .test_data_long_legged_doji import TEST_DATA_LONG_LEGGED_DOJI
from .test_data_long_line import TEST_DATA_LONG_LINE
from .test_data_marubozu import TEST_DATA_MARUBOZU
from .test_data_mat_hold import TEST_DATA_MAT_HOLD
from .test_data_matching_low import TEST_DATA_MATCHING_LOW
from .test_data_morning_doji_star import TEST_DATA_MORNING_DOJI_STAR
from .test_data_morning_star import TEST_DATA_MORNING_STAR
from .test_data_on_neck import TEST_DATA_ON_NECK
from .test_data_piercing import TEST_DATA_PIERCING
from .test_data_rickshaw_man import TEST_DATA_RICKSHAW_MAN
from .test_data_rising_falling_three_methods import TEST_DATA_RISING_FALLING_THREE_METHODS
from .test_data_separating_lines import TEST_DATA_SEPARATING_LINES
from .test_data_shooting_star import TEST_DATA_SHOOTING_STAR
from .test_data_short_line import TEST_DATA_SHORT_LINE
from .test_data_spinning_top import TEST_DATA_SPINNING_TOP
from .test_data_stalled import TEST_DATA_STALLED
from .test_data_stick_sandwich import TEST_DATA_STICK_SANDWICH
from .test_data_takuri import TEST_DATA_TAKURI
from .test_data_tasuki_gap import TEST_DATA_TASUKI_GAP
from .test_data_three_black_crows import TEST_DATA_THREE_BLACK_CROWS
from .test_data_three_inside import TEST_DATA_THREE_INSIDE
from .test_data_three_line_strike import TEST_DATA_THREE_LINE_STRIKE
from .test_data_three_outside import TEST_DATA_THREE_OUTSIDE
from .test_data_three_stars_in_the_south import TEST_DATA_THREE_STARS_IN_THE_SOUTH
from .test_data_three_white_soldiers import TEST_DATA_THREE_WHITE_SOLDIERS
from .test_data_thrusting import TEST_DATA_THRUSTING
from .test_data_tristar import TEST_DATA_TRISTAR
from .test_data_two_crows import TEST_DATA_TWO_CROWS
from .test_data_unique_three_river import TEST_DATA_UNIQUE_THREE_RIVER
from .test_data_up_down_gap_side_by_side_white_lines import TEST_DATA_UP_DOWN_GAP_SIDE_BY_SIDE_WHITE_LINES
from .test_data_upside_gap_two_crows import TEST_DATA_UPSIDE_GAP_TWO_CROWS
from .test_data_x_side_gap_three_methods import TEST_DATA_X_SIDE_GAP_THREE_METHODS


def _run_pattern(test_data, pattern_method_name):
    """Feed each test case into a fresh CandlestickPatterns and check the last bar's result."""
    results = []
    for i, (opens, highs, lows, closes, expected) in enumerate(test_data):
        cp = CandlestickPatterns()
        for o, h, l, c in zip(opens, highs, lows, closes):
            cp.update(o, h, l, c)
        method = getattr(cp, pattern_method_name)
        actual = method()
        results.append((i, expected, actual))
    return results


class TestCandlestickPatterns(unittest.TestCase):
    """Test all 61 candlestick patterns against TA-Lib reference data."""

    def _assert_pattern(self, test_data, pattern_name):
        if not test_data:
            self.skipTest(f'No test data for {pattern_name}')
        results = _run_pattern(test_data, pattern_name)
        failures = [(i, exp, act) for i, exp, act in results if exp != act]
        if failures:
            msgs = [f'  case {i}: expected {exp}, got {act}' for i, exp, act in failures]
            self.fail(
                f'{pattern_name}: {len(failures)}/{len(results)} cases failed:\n'
                + '\n'.join(msgs)
            )

    def test_abandoned_baby(self):
        self._assert_pattern(TEST_DATA_ABANDONED_BABY, 'abandoned_baby')

    def test_advance_block(self):
        self._assert_pattern(TEST_DATA_ADVANCE_BLOCK, 'advance_block')

    def test_belt_hold(self):
        self._assert_pattern(TEST_DATA_BELT_HOLD, 'belt_hold')

    def test_breakaway(self):
        self._assert_pattern(TEST_DATA_BREAKAWAY, 'breakaway')

    def test_closing_marubozu(self):
        self._assert_pattern(TEST_DATA_CLOSING_MARUBOZU, 'closing_marubozu')

    def test_concealing_baby_swallow(self):
        self._assert_pattern(TEST_DATA_CONCEALING_BABY_SWALLOW, 'concealing_baby_swallow')

    def test_counterattack(self):
        self._assert_pattern(TEST_DATA_COUNTERATTACK, 'counterattack')

    def test_dark_cloud_cover(self):
        self._assert_pattern(TEST_DATA_DARK_CLOUD_COVER, 'dark_cloud_cover')

    def test_doji(self):
        self._assert_pattern(TEST_DATA_DOJI, 'doji')

    def test_doji_star(self):
        self._assert_pattern(TEST_DATA_DOJI_STAR, 'doji_star')

    def test_dragonfly_doji(self):
        self._assert_pattern(TEST_DATA_DRAGONFLY_DOJI, 'dragonfly_doji')

    def test_engulfing(self):
        self._assert_pattern(TEST_DATA_ENGULFING, 'engulfing')

    def test_evening_doji_star(self):
        self._assert_pattern(TEST_DATA_EVENING_DOJI_STAR, 'evening_doji_star')

    def test_evening_star(self):
        self._assert_pattern(TEST_DATA_EVENING_STAR, 'evening_star')

    def test_gravestone_doji(self):
        self._assert_pattern(TEST_DATA_GRAVESTONE_DOJI, 'gravestone_doji')

    def test_hammer(self):
        self._assert_pattern(TEST_DATA_HAMMER, 'hammer')

    def test_hanging_man(self):
        self._assert_pattern(TEST_DATA_HANGING_MAN, 'hanging_man')

    def test_harami(self):
        self._assert_pattern(TEST_DATA_HARAMI, 'harami')

    def test_harami_cross(self):
        self._assert_pattern(TEST_DATA_HARAMI_CROSS, 'harami_cross')

    def test_high_wave(self):
        self._assert_pattern(TEST_DATA_HIGH_WAVE, 'high_wave')

    def test_hikkake(self):
        self._assert_pattern(TEST_DATA_HIKKAKE, 'hikkake')

    def test_hikkake_modified(self):
        self._assert_pattern(TEST_DATA_HIKKAKE_MODIFIED, 'hikkake_modified')

    def test_homing_pigeon(self):
        self._assert_pattern(TEST_DATA_HOMING_PIGEON, 'homing_pigeon')

    def test_identical_three_crows(self):
        self._assert_pattern(TEST_DATA_IDENTICAL_THREE_CROWS, 'identical_three_crows')

    def test_in_neck(self):
        self._assert_pattern(TEST_DATA_IN_NECK, 'in_neck')

    def test_inverted_hammer(self):
        self._assert_pattern(TEST_DATA_INVERTED_HAMMER, 'inverted_hammer')

    def test_kicking(self):
        self._assert_pattern(TEST_DATA_KICKING, 'kicking')

    def test_kicking_by_length(self):
        self._assert_pattern(TEST_DATA_KICKING_BY_LENGTH, 'kicking_by_length')

    def test_ladder_bottom(self):
        self._assert_pattern(TEST_DATA_LADDER_BOTTOM, 'ladder_bottom')

    def test_long_legged_doji(self):
        self._assert_pattern(TEST_DATA_LONG_LEGGED_DOJI, 'long_legged_doji')

    def test_long_line(self):
        self._assert_pattern(TEST_DATA_LONG_LINE, 'long_line')

    def test_marubozu(self):
        self._assert_pattern(TEST_DATA_MARUBOZU, 'marubozu')

    def test_mat_hold(self):
        self._assert_pattern(TEST_DATA_MAT_HOLD, 'mat_hold')

    def test_matching_low(self):
        self._assert_pattern(TEST_DATA_MATCHING_LOW, 'matching_low')

    def test_morning_doji_star(self):
        self._assert_pattern(TEST_DATA_MORNING_DOJI_STAR, 'morning_doji_star')

    def test_morning_star(self):
        self._assert_pattern(TEST_DATA_MORNING_STAR, 'morning_star')

    def test_on_neck(self):
        self._assert_pattern(TEST_DATA_ON_NECK, 'on_neck')

    def test_piercing(self):
        self._assert_pattern(TEST_DATA_PIERCING, 'piercing')

    def test_rickshaw_man(self):
        self._assert_pattern(TEST_DATA_RICKSHAW_MAN, 'rickshaw_man')

    def test_rising_falling_three_methods(self):
        self._assert_pattern(TEST_DATA_RISING_FALLING_THREE_METHODS, 'rising_falling_three_methods')

    def test_separating_lines(self):
        self._assert_pattern(TEST_DATA_SEPARATING_LINES, 'separating_lines')

    def test_shooting_star(self):
        self._assert_pattern(TEST_DATA_SHOOTING_STAR, 'shooting_star')

    def test_short_line(self):
        self._assert_pattern(TEST_DATA_SHORT_LINE, 'short_line')

    def test_spinning_top(self):
        self._assert_pattern(TEST_DATA_SPINNING_TOP, 'spinning_top')

    def test_stalled(self):
        self._assert_pattern(TEST_DATA_STALLED, 'stalled')

    def test_stick_sandwich(self):
        self._assert_pattern(TEST_DATA_STICK_SANDWICH, 'stick_sandwich')

    def test_takuri(self):
        self._assert_pattern(TEST_DATA_TAKURI, 'takuri')

    def test_tasuki_gap(self):
        self._assert_pattern(TEST_DATA_TASUKI_GAP, 'tasuki_gap')

    def test_three_black_crows(self):
        self._assert_pattern(TEST_DATA_THREE_BLACK_CROWS, 'three_black_crows')

    def test_three_inside(self):
        self._assert_pattern(TEST_DATA_THREE_INSIDE, 'three_inside')

    def test_three_line_strike(self):
        self._assert_pattern(TEST_DATA_THREE_LINE_STRIKE, 'three_line_strike')

    def test_three_outside(self):
        self._assert_pattern(TEST_DATA_THREE_OUTSIDE, 'three_outside')

    def test_three_stars_in_the_south(self):
        self._assert_pattern(TEST_DATA_THREE_STARS_IN_THE_SOUTH, 'three_stars_in_the_south')

    def test_three_white_soldiers(self):
        self._assert_pattern(TEST_DATA_THREE_WHITE_SOLDIERS, 'three_white_soldiers')

    def test_thrusting(self):
        self._assert_pattern(TEST_DATA_THRUSTING, 'thrusting')

    def test_tristar(self):
        self._assert_pattern(TEST_DATA_TRISTAR, 'tristar')

    def test_two_crows(self):
        self._assert_pattern(TEST_DATA_TWO_CROWS, 'two_crows')

    def test_unique_three_river(self):
        self._assert_pattern(TEST_DATA_UNIQUE_THREE_RIVER, 'unique_three_river')

    def test_up_down_gap_side_by_side_white_lines(self):
        self._assert_pattern(TEST_DATA_UP_DOWN_GAP_SIDE_BY_SIDE_WHITE_LINES, 'up_down_gap_side_by_side_white_lines')

    def test_upside_gap_two_crows(self):
        self._assert_pattern(TEST_DATA_UPSIDE_GAP_TWO_CROWS, 'upside_gap_two_crows')

    def test_x_side_gap_three_methods(self):
        self._assert_pattern(TEST_DATA_X_SIDE_GAP_THREE_METHODS, 'x_side_gap_three_methods')


if __name__ == '__main__':
    unittest.main()
