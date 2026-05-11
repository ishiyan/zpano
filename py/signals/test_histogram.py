"""Tests for histogram signals."""
from __future__ import annotations

import unittest

from py.signals.histogram import mu_turns_positive, mu_turns_negative


class TestTurnsPositive(unittest.TestCase):

    def test_clear_turn_positive(self):
        """prev negative, curr positive → 1.0."""
        self.assertAlmostEqual(mu_turns_positive(-5.0, 5.0, width=0.0), 1.0, places=10)

    def test_stays_positive(self):
        """Both positive → 0.0 (was not non-positive)."""
        self.assertAlmostEqual(mu_turns_positive(3.0, 5.0, width=0.0), 0.0, places=10)

    def test_stays_negative(self):
        """Both negative → 0.0 (is not positive)."""
        self.assertAlmostEqual(mu_turns_positive(-5.0, -3.0, width=0.0), 0.0, places=10)

    def test_turns_more_negative(self):
        """prev positive, curr negative → 0.0 (wrong direction)."""
        self.assertAlmostEqual(mu_turns_positive(5.0, -5.0, width=0.0), 0.0, places=10)

    def test_from_zero(self):
        """prev=0, curr positive → 0.5 * 1.0 = 0.5."""
        result = mu_turns_positive(0.0, 5.0, width=0.0)
        self.assertAlmostEqual(result, 0.5, places=10)

    def test_fuzzy_near_zero(self):
        """Small values near zero → partial membership with width."""
        result = mu_turns_positive(-0.5, 0.5, width=2.0)
        self.assertGreater(result, 0.1)
        self.assertLess(result, 0.95)

    def test_fuzzy_width_makes_softer(self):
        """Larger width → less extreme membership for same values."""
        narrow = mu_turns_positive(-1.0, 1.0, width=0.5)
        wide = mu_turns_positive(-1.0, 1.0, width=10.0)
        self.assertGreater(narrow, wide)


class TestTurnsNegative(unittest.TestCase):

    def test_clear_turn_negative(self):
        self.assertAlmostEqual(mu_turns_negative(5.0, -5.0, width=0.0), 1.0, places=10)

    def test_stays_negative(self):
        self.assertAlmostEqual(mu_turns_negative(-5.0, -3.0, width=0.0), 0.0, places=10)

    def test_stays_positive(self):
        self.assertAlmostEqual(mu_turns_negative(3.0, 5.0, width=0.0), 0.0, places=10)

    def test_symmetry(self):
        """turns_negative(a, b) == turns_positive(-a, -b)."""
        tn = mu_turns_negative(3.0, -3.0, width=1.0)
        tp = mu_turns_positive(-3.0, 3.0, width=1.0)
        self.assertAlmostEqual(tn, tp, places=10)


if __name__ == '__main__':
    unittest.main()
