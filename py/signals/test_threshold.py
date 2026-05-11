"""Tests for threshold signals."""
from __future__ import annotations

import unittest

from py.fuzzy import MembershipShape
from py.signals.threshold import mu_above, mu_below, mu_overbought, mu_oversold


class TestMuAbove(unittest.TestCase):

    def test_well_above(self):
        """Value clearly above threshold → near 1.0."""
        self.assertAlmostEqual(mu_above(80.0, 70.0, width=5.0), 1.0, places=2)

    def test_well_below(self):
        """Value clearly below threshold → near 0.0."""
        self.assertAlmostEqual(mu_above(60.0, 70.0, width=5.0), 0.0, places=2)

    def test_at_threshold(self):
        """Value at threshold → 0.5."""
        self.assertAlmostEqual(mu_above(70.0, 70.0, width=5.0), 0.5, places=10)

    def test_zero_width_above(self):
        """Crisp: value > threshold → 1.0."""
        self.assertEqual(mu_above(70.1, 70.0, width=0.0), 1.0)

    def test_zero_width_below(self):
        """Crisp: value < threshold → 0.0."""
        self.assertEqual(mu_above(69.9, 70.0, width=0.0), 0.0)

    def test_zero_width_equal(self):
        """Crisp: value == threshold → 0.5."""
        self.assertAlmostEqual(mu_above(70.0, 70.0, width=0.0), 0.5, places=10)

    def test_monotonic(self):
        """Higher value → higher membership."""
        m1 = mu_above(68.0, 70.0, width=5.0)
        m2 = mu_above(70.0, 70.0, width=5.0)
        m3 = mu_above(72.0, 70.0, width=5.0)
        self.assertLess(m1, m2)
        self.assertLess(m2, m3)

    def test_linear_shape(self):
        """Linear shape gives 0.5 at threshold, 0 and 1 at edges."""
        self.assertAlmostEqual(mu_above(70.0, 70.0, 10.0, MembershipShape.LINEAR), 0.5)
        self.assertAlmostEqual(mu_above(65.0, 70.0, 10.0, MembershipShape.LINEAR), 0.0)
        self.assertAlmostEqual(mu_above(75.0, 70.0, 10.0, MembershipShape.LINEAR), 1.0)


class TestMuBelow(unittest.TestCase):

    def test_well_below(self):
        self.assertAlmostEqual(mu_below(20.0, 30.0, width=5.0), 1.0, places=2)

    def test_well_above(self):
        self.assertAlmostEqual(mu_below(40.0, 30.0, width=5.0), 0.0, places=2)

    def test_at_threshold(self):
        self.assertAlmostEqual(mu_below(30.0, 30.0, width=5.0), 0.5, places=10)

    def test_complement_of_above(self):
        """mu_below + mu_above ≈ 1 for any value."""
        for v in [25.0, 30.0, 35.0, 50.0]:
            total = mu_below(v, 30.0, 5.0) + mu_above(v, 30.0, 5.0)
            self.assertAlmostEqual(total, 1.0, places=10)


class TestOverboughtOversold(unittest.TestCase):

    def test_overbought_high_rsi(self):
        self.assertGreater(mu_overbought(85.0), 0.95)

    def test_overbought_low_rsi(self):
        self.assertLess(mu_overbought(50.0), 0.01)

    def test_oversold_low_rsi(self):
        self.assertGreater(mu_oversold(15.0), 0.95)

    def test_oversold_high_rsi(self):
        self.assertLess(mu_oversold(50.0), 0.01)

    def test_overbought_custom_level(self):
        self.assertAlmostEqual(mu_overbought(80.0, level=80.0), 0.5, places=10)

    def test_oversold_custom_level(self):
        self.assertAlmostEqual(mu_oversold(20.0, level=20.0), 0.5, places=10)


if __name__ == '__main__':
    unittest.main()
