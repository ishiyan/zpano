"""Tests for band signals."""
from __future__ import annotations

import unittest

from py.signals.band import mu_above_band, mu_below_band, mu_between_bands


class TestAboveBand(unittest.TestCase):

    def test_well_above(self):
        self.assertAlmostEqual(mu_above_band(110.0, 100.0, width=5.0), 1.0, places=2)

    def test_well_below(self):
        self.assertAlmostEqual(mu_above_band(90.0, 100.0, width=5.0), 0.0, places=2)

    def test_at_band(self):
        self.assertAlmostEqual(mu_above_band(100.0, 100.0, width=5.0), 0.5, places=10)

    def test_crisp(self):
        self.assertEqual(mu_above_band(100.1, 100.0, width=0.0), 1.0)
        self.assertEqual(mu_above_band(99.9, 100.0, width=0.0), 0.0)


class TestBelowBand(unittest.TestCase):

    def test_well_below(self):
        self.assertAlmostEqual(mu_below_band(85.0, 90.0, width=5.0), 1.0, places=2)

    def test_well_above(self):
        self.assertAlmostEqual(mu_below_band(100.0, 90.0, width=5.0), 0.0, places=2)

    def test_at_band(self):
        self.assertAlmostEqual(mu_below_band(90.0, 90.0, width=5.0), 0.5, places=10)


class TestBetweenBands(unittest.TestCase):

    def test_centered(self):
        """Value in the middle of bands → high membership."""
        result = mu_between_bands(100.0, 90.0, 110.0)
        self.assertGreater(result, 0.8)

    def test_at_upper_band(self):
        """Value at upper band → reduced membership."""
        result = mu_between_bands(110.0, 90.0, 110.0)
        self.assertLess(result, 0.6)

    def test_at_lower_band(self):
        """Value at lower band → reduced membership."""
        result = mu_between_bands(90.0, 90.0, 110.0)
        self.assertLess(result, 0.6)

    def test_outside_above(self):
        """Value well above upper band → near 0."""
        result = mu_between_bands(130.0, 90.0, 110.0)
        self.assertLess(result, 0.1)

    def test_outside_below(self):
        """Value well below lower band → near 0."""
        result = mu_between_bands(70.0, 90.0, 110.0)
        self.assertLess(result, 0.1)

    def test_degenerate_bands(self):
        """upper <= lower → 0.0."""
        self.assertEqual(mu_between_bands(100.0, 110.0, 90.0), 0.0)
        self.assertEqual(mu_between_bands(100.0, 100.0, 100.0), 0.0)

    def test_monotonic_from_center(self):
        """Membership decreases as value moves away from center."""
        center = mu_between_bands(100.0, 90.0, 110.0)
        edge = mu_between_bands(108.0, 90.0, 110.0)
        outside = mu_between_bands(115.0, 90.0, 110.0)
        self.assertGreater(center, edge)
        self.assertGreater(edge, outside)


if __name__ == '__main__':
    unittest.main()
