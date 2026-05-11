"""Unit tests for fuzzy operators."""
from __future__ import annotations

import unittest

from .operators import (
    t_product, t_min, t_lukasiewicz,
    s_probabilistic, s_max,
    f_not,
    t_product_all, t_min_all,
)


class TestTNorms(unittest.TestCase):
    """Tests for t-norm (fuzzy AND) operators."""

    # -- Product ---------------------------------------------------------

    def test_product_basic(self):
        self.assertAlmostEqual(t_product(0.8, 0.6), 0.48)

    def test_product_identity(self):
        """T(a, 1) = a."""
        self.assertAlmostEqual(t_product(0.7, 1.0), 0.7)

    def test_product_annihilator(self):
        """T(a, 0) = 0."""
        self.assertAlmostEqual(t_product(0.7, 0.0), 0.0)

    def test_product_commutativity(self):
        self.assertAlmostEqual(t_product(0.3, 0.8), t_product(0.8, 0.3))

    # -- Minimum ---------------------------------------------------------

    def test_min_basic(self):
        self.assertEqual(t_min(0.8, 0.6), 0.6)

    def test_min_identity(self):
        self.assertEqual(t_min(0.7, 1.0), 0.7)

    def test_min_annihilator(self):
        self.assertEqual(t_min(0.7, 0.0), 0.0)

    # -- Łukasiewicz -----------------------------------------------------

    def test_lukasiewicz_both_high(self):
        self.assertAlmostEqual(t_lukasiewicz(0.9, 0.8), 0.7)

    def test_lukasiewicz_one_low(self):
        self.assertAlmostEqual(t_lukasiewicz(0.3, 0.5), 0.0)

    def test_lukasiewicz_clamp(self):
        """Result never negative."""
        self.assertEqual(t_lukasiewicz(0.1, 0.2), 0.0)

    def test_lukasiewicz_identity(self):
        self.assertAlmostEqual(t_lukasiewicz(0.7, 1.0), 0.7)


class TestSNorms(unittest.TestCase):
    """Tests for s-norm (fuzzy OR) operators."""

    def test_probabilistic_basic(self):
        self.assertAlmostEqual(s_probabilistic(0.8, 0.6), 0.92)

    def test_probabilistic_identity(self):
        """S(a, 0) = a."""
        self.assertAlmostEqual(s_probabilistic(0.7, 0.0), 0.7)

    def test_probabilistic_annihilator(self):
        """S(a, 1) = 1."""
        self.assertAlmostEqual(s_probabilistic(0.7, 1.0), 1.0)

    def test_max_basic(self):
        self.assertEqual(s_max(0.8, 0.6), 0.8)

    def test_max_identity(self):
        self.assertEqual(s_max(0.7, 0.0), 0.7)


class TestNegation(unittest.TestCase):
    """Tests for fuzzy negation."""

    def test_not_basic(self):
        self.assertAlmostEqual(f_not(0.3), 0.7)

    def test_not_zero(self):
        self.assertAlmostEqual(f_not(0.0), 1.0)

    def test_not_one(self):
        self.assertAlmostEqual(f_not(1.0), 0.0)

    def test_double_negation(self):
        self.assertAlmostEqual(f_not(f_not(0.4)), 0.4, places=10)


class TestVariadic(unittest.TestCase):
    """Tests for variadic t-norm helpers."""

    def test_product_all_three(self):
        self.assertAlmostEqual(t_product_all(0.8, 0.6, 0.5), 0.24)

    def test_product_all_single(self):
        self.assertAlmostEqual(t_product_all(0.7), 0.7)

    def test_product_all_empty(self):
        """Identity element: product of nothing = 1."""
        self.assertAlmostEqual(t_product_all(), 1.0)

    def test_min_all_three(self):
        self.assertEqual(t_min_all(0.8, 0.6, 0.9), 0.6)

    def test_min_all_empty(self):
        self.assertEqual(t_min_all(), 1.0)

    def test_product_all_five(self):
        """Five conditions at μ=0.9 → 0.9⁵ ≈ 0.59."""
        result = t_product_all(0.9, 0.9, 0.9, 0.9, 0.9)
        self.assertAlmostEqual(result, 0.9 ** 5, places=10)


class TestDuality(unittest.TestCase):
    """Verify De Morgan's law: T(a,b) = 1 - S(1-a, 1-b)."""

    def test_product_probabilistic_duality(self):
        a, b = 0.7, 0.4
        lhs = t_product(a, b)
        rhs = f_not(s_probabilistic(f_not(a), f_not(b)))
        self.assertAlmostEqual(lhs, rhs, places=10)

    def test_min_max_duality(self):
        a, b = 0.7, 0.4
        lhs = t_min(a, b)
        rhs = f_not(s_max(f_not(a), f_not(b)))
        self.assertAlmostEqual(lhs, rhs, places=10)


if __name__ == '__main__':
    unittest.main()
