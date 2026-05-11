"""Tests for crossover signals."""
from __future__ import annotations

import unittest

from py.signals.crossover import (
    mu_crosses_above, mu_crosses_below,
    mu_line_crosses_above, mu_line_crosses_below,
)


class TestCrossesAbove(unittest.TestCase):

    def test_clear_cross_above(self):
        """prev well below, curr well above → near 1.0."""
        result = mu_crosses_above(25.0, 35.0, 30.0, width=0.0)
        self.assertAlmostEqual(result, 1.0, places=10)

    def test_no_cross_both_above(self):
        """Both above threshold → near 0.0 (was not below)."""
        result = mu_crosses_above(35.0, 40.0, 30.0, width=0.0)
        self.assertAlmostEqual(result, 0.0, places=10)

    def test_no_cross_both_below(self):
        """Both below threshold → near 0.0 (is not above)."""
        result = mu_crosses_above(25.0, 28.0, 30.0, width=0.0)
        self.assertAlmostEqual(result, 0.0, places=10)

    def test_cross_down_not_up(self):
        """prev above, curr below → 0.0 (wrong direction)."""
        result = mu_crosses_above(35.0, 25.0, 30.0, width=0.0)
        self.assertAlmostEqual(result, 0.0, places=10)

    def test_fuzzy_near_threshold(self):
        """prev just below, curr just above → moderate membership."""
        result = mu_crosses_above(29.0, 31.0, 30.0, width=5.0)
        self.assertGreater(result, 0.1)
        self.assertLess(result, 0.9)

    def test_at_threshold(self):
        """Both at threshold → 0.5 * 0.5 = 0.25."""
        result = mu_crosses_above(30.0, 30.0, 30.0, width=0.0)
        self.assertAlmostEqual(result, 0.25, places=10)


class TestCrossesBelow(unittest.TestCase):

    def test_clear_cross_below(self):
        result = mu_crosses_below(35.0, 25.0, 30.0, width=0.0)
        self.assertAlmostEqual(result, 1.0, places=10)

    def test_no_cross_both_below(self):
        result = mu_crosses_below(25.0, 20.0, 30.0, width=0.0)
        self.assertAlmostEqual(result, 0.0, places=10)

    def test_symmetry(self):
        """crosses_below(a, b, t) == crosses_above(b, a, t)."""
        cb = mu_crosses_below(35.0, 25.0, 30.0, width=2.0)
        ca = mu_crosses_above(25.0, 35.0, 30.0, width=2.0)
        self.assertAlmostEqual(cb, ca, places=10)


class TestLineCrossesAbove(unittest.TestCase):

    def test_golden_cross(self):
        """Fast was below slow, now above → bullish crossover."""
        result = mu_line_crosses_above(
            prev_fast=49.0, curr_fast=51.0,
            prev_slow=50.0, curr_slow=50.0,
            width=0.0,
        )
        self.assertAlmostEqual(result, 1.0, places=10)

    def test_no_cross(self):
        """Fast stays above slow → 0.0."""
        result = mu_line_crosses_above(
            prev_fast=52.0, curr_fast=53.0,
            prev_slow=50.0, curr_slow=50.0,
            width=0.0,
        )
        self.assertAlmostEqual(result, 0.0, places=10)

    def test_fuzzy_near_cross(self):
        """Lines are close → moderate membership."""
        result = mu_line_crosses_above(
            prev_fast=49.5, curr_fast=50.5,
            prev_slow=50.0, curr_slow=50.0,
            width=2.0,
        )
        self.assertGreater(result, 0.0)
        self.assertLess(result, 1.0)


class TestLineCrossesBelow(unittest.TestCase):

    def test_death_cross(self):
        """Fast was above slow, now below → bearish crossover."""
        result = mu_line_crosses_below(
            prev_fast=51.0, curr_fast=49.0,
            prev_slow=50.0, curr_slow=50.0,
            width=0.0,
        )
        self.assertAlmostEqual(result, 1.0, places=10)

    def test_no_cross(self):
        result = mu_line_crosses_below(
            prev_fast=48.0, curr_fast=47.0,
            prev_slow=50.0, curr_slow=50.0,
            width=0.0,
        )
        self.assertAlmostEqual(result, 0.0, places=10)


if __name__ == '__main__':
    unittest.main()
