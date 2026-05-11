"""Unit tests for fuzzy membership functions."""
from __future__ import annotations

import unittest

from .membership import mu_less, mu_less_equal, mu_greater, mu_greater_equal, \
    mu_near, mu_direction, MembershipShape


class TestMuLess(unittest.TestCase):
    """Tests for mu_less and mu_less_equal."""

    # -- Sigmoid shape (default) -----------------------------------------

    def test_crossover_at_threshold(self):
        """μ = 0.5 exactly at x == threshold."""
        self.assertAlmostEqual(mu_less(10.0, 10.0, 2.0), 0.5, places=10)

    def test_well_below_threshold(self):
        """μ ≈ 1.0 when x is well below threshold."""
        val = mu_less(8.0, 10.0, 2.0)
        self.assertGreater(val, 0.99)

    def test_well_above_threshold(self):
        """μ ≈ 0.0 when x is well above threshold."""
        val = mu_less(12.0, 10.0, 2.0)
        self.assertLess(val, 0.01)

    def test_monotonically_decreasing(self):
        """μ should decrease as x increases."""
        vals = [mu_less(x, 10.0, 2.0) for x in [8.0, 9.0, 10.0, 11.0, 12.0]]
        for i in range(len(vals) - 1):
            self.assertGreater(vals[i], vals[i + 1])

    def test_symmetry(self):
        """Sigmoid is symmetric around the threshold."""
        below = mu_less(9.0, 10.0, 2.0)
        above = mu_less(11.0, 10.0, 2.0)
        self.assertAlmostEqual(below + above, 1.0, places=10)

    # -- Linear shape ----------------------------------------------------

    def test_linear_crossover(self):
        self.assertAlmostEqual(mu_less(10.0, 10.0, 4.0, MembershipShape.LINEAR), 0.5)

    def test_linear_below_range(self):
        self.assertEqual(mu_less(7.0, 10.0, 4.0, MembershipShape.LINEAR), 1.0)

    def test_linear_above_range(self):
        self.assertEqual(mu_less(13.0, 10.0, 4.0, MembershipShape.LINEAR), 0.0)

    def test_linear_midpoint(self):
        self.assertAlmostEqual(mu_less(9.0, 10.0, 4.0, MembershipShape.LINEAR), 0.75)

    # -- Crisp (width=0) -------------------------------------------------

    def test_crisp_below(self):
        self.assertEqual(mu_less(9.0, 10.0, 0.0), 1.0)

    def test_crisp_above(self):
        self.assertEqual(mu_less(11.0, 10.0, 0.0), 0.0)

    def test_crisp_at_threshold(self):
        self.assertEqual(mu_less(10.0, 10.0, 0.0), 0.5)

    # -- mu_less_equal is identical for continuous values -----------------

    def test_less_equal_same_as_less(self):
        self.assertEqual(mu_less_equal(9.5, 10.0, 2.0),
                         mu_less(9.5, 10.0, 2.0))


class TestMuGreater(unittest.TestCase):
    """Tests for mu_greater and mu_greater_equal."""

    def test_complement_of_less(self):
        """mu_greater = 1 - mu_less."""
        for x in [8.0, 9.0, 10.0, 11.0, 12.0]:
            self.assertAlmostEqual(
                mu_greater(x, 10.0, 2.0) + mu_less(x, 10.0, 2.0),
                1.0, places=10)

    def test_crossover(self):
        self.assertAlmostEqual(mu_greater(10.0, 10.0, 2.0), 0.5, places=10)

    def test_well_above(self):
        self.assertGreater(mu_greater(12.0, 10.0, 2.0), 0.99)

    def test_well_below(self):
        self.assertLess(mu_greater(8.0, 10.0, 2.0), 0.01)

    def test_greater_equal_complement(self):
        self.assertAlmostEqual(
            mu_greater_equal(9.5, 10.0, 2.0) + mu_less_equal(9.5, 10.0, 2.0),
            1.0, places=10)


class TestMuNear(unittest.TestCase):
    """Tests for mu_near (bell-shaped membership)."""

    # -- Sigmoid (Gaussian bell) -----------------------------------------

    def test_peak_at_target(self):
        self.assertAlmostEqual(mu_near(10.0, 10.0, 2.0), 1.0, places=10)

    def test_falls_off(self):
        """μ should be small at distance = width from target."""
        val = mu_near(12.0, 10.0, 2.0)
        self.assertLess(val, 0.05)

    def test_symmetric(self):
        below = mu_near(9.0, 10.0, 2.0)
        above = mu_near(11.0, 10.0, 2.0)
        self.assertAlmostEqual(below, above, places=10)

    def test_monotonic_from_center(self):
        """μ decreases as distance from target increases."""
        vals = [mu_near(10.0 + d, 10.0, 2.0) for d in [0, 0.5, 1.0, 1.5, 2.0]]
        for i in range(len(vals) - 1):
            self.assertGreater(vals[i], vals[i + 1])

    # -- Linear (triangular) ---------------------------------------------

    def test_linear_peak(self):
        self.assertAlmostEqual(mu_near(10.0, 10.0, 2.0, MembershipShape.LINEAR), 1.0)

    def test_linear_at_boundary(self):
        self.assertEqual(mu_near(12.0, 10.0, 2.0, MembershipShape.LINEAR), 0.0)

    def test_linear_midpoint(self):
        self.assertAlmostEqual(mu_near(11.0, 10.0, 2.0, MembershipShape.LINEAR), 0.5)

    # -- Crisp (width=0) -------------------------------------------------

    def test_crisp_exact(self):
        self.assertEqual(mu_near(10.0, 10.0, 0.0), 1.0)

    def test_crisp_any_distance(self):
        self.assertEqual(mu_near(10.1, 10.0, 0.0), 0.0)


class TestMuDirection(unittest.TestCase):
    """Tests for mu_direction (fuzzy candle direction)."""

    def test_large_white_body(self):
        """Large bullish bar → direction ≈ +1."""
        d = mu_direction(100.0, 110.0, 5.0)
        self.assertGreater(d, 0.95)

    def test_large_black_body(self):
        """Large bearish bar → direction ≈ -1."""
        d = mu_direction(110.0, 100.0, 5.0)
        self.assertLess(d, -0.95)

    def test_doji(self):
        """Doji (close ≈ open) → direction ≈ 0."""
        d = mu_direction(100.0, 100.0, 5.0)
        self.assertAlmostEqual(d, 0.0, places=10)

    def test_tiny_white_body(self):
        """Small bullish bar → direction slightly positive."""
        d = mu_direction(100.0, 100.1, 5.0)
        self.assertGreater(d, 0.0)
        self.assertLess(d, 0.1)

    def test_antisymmetric(self):
        """direction(o, c) = -direction(c, o) when body_avg is same."""
        d1 = mu_direction(100.0, 105.0, 5.0)
        d2 = mu_direction(105.0, 100.0, 5.0)
        self.assertAlmostEqual(d1, -d2, places=10)

    def test_zero_body_avg_white(self):
        """body_avg=0 → crisp +1 for white."""
        self.assertEqual(mu_direction(100.0, 101.0, 0.0), 1.0)

    def test_zero_body_avg_black(self):
        """body_avg=0 → crisp -1 for black."""
        self.assertEqual(mu_direction(101.0, 100.0, 0.0), -1.0)

    def test_zero_body_avg_doji(self):
        """body_avg=0, doji → crisp +1 (c >= o)."""
        self.assertEqual(mu_direction(100.0, 100.0, 0.0), 1.0)

    def test_range_bounded(self):
        """Direction should always be in [-1, +1]."""
        for o, c, avg in [(0, 1000, 1), (1000, 0, 1), (50, 50, 100)]:
            d = mu_direction(o, c, avg)
            self.assertGreaterEqual(d, -1.0)
            self.assertLessEqual(d, 1.0)


class TestEdgeCases(unittest.TestCase):
    """Edge cases and numerical robustness."""

    def test_very_large_x(self):
        """No overflow for extreme values."""
        val = mu_less(1e10, 0.0, 1.0)
        self.assertEqual(val, 0.0)

    def test_very_small_x(self):
        val = mu_less(-1e10, 0.0, 1.0)
        self.assertEqual(val, 1.0)

    def test_tiny_width(self):
        """Very small width → approaches crisp."""
        val = mu_less(9.999, 10.0, 0.001)
        self.assertGreater(val, 0.99)

    def test_huge_width(self):
        """Very large width → everything near 0.5."""
        val = mu_less(0.0, 10.0, 1000.0)
        self.assertGreater(val, 0.49)
        self.assertLess(val, 0.60)


if __name__ == '__main__':
    unittest.main()
