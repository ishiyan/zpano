"""Unit tests for fuzzy defuzzification utilities."""
from __future__ import annotations

import unittest

from .defuzzify import alpha_cut


class TestAlphaCut(unittest.TestCase):
    """Tests for the alpha_cut function."""

    # -- Basic behavior --------------------------------------------------

    def test_strong_bearish(self):
        """High confidence bearish → -100."""
        self.assertEqual(alpha_cut(-87.3), -100)

    def test_weak_bearish(self):
        """Low confidence bearish → 0 (filtered out)."""
        self.assertEqual(alpha_cut(-32.1), 0)

    def test_strong_bullish(self):
        self.assertEqual(alpha_cut(92.5), 100)

    def test_weak_bullish(self):
        self.assertEqual(alpha_cut(15.0), 0)

    def test_zero(self):
        self.assertEqual(alpha_cut(0.0), 0)

    # -- Confirmation level (200) ----------------------------------------

    def test_strong_confirmation(self):
        """High confidence confirmation → 200."""
        self.assertEqual(alpha_cut(156.8), 200)

    def test_negative_confirmation(self):
        self.assertEqual(alpha_cut(-180.0), -200)

    # -- Alpha threshold tuning ------------------------------------------

    def test_high_alpha_filters_more(self):
        """alpha=0.9 filters out 87% confidence."""
        self.assertEqual(alpha_cut(-87.3, alpha=0.9), 0)

    def test_high_alpha_passes_strong(self):
        self.assertEqual(alpha_cut(-95.0, alpha=0.9), -100)

    def test_low_alpha_passes_more(self):
        """alpha=0.1 passes even weak signals."""
        self.assertEqual(alpha_cut(-15.0, alpha=0.1), -100)

    def test_alpha_zero_passes_all(self):
        """alpha=0 passes everything except exact 0."""
        self.assertEqual(alpha_cut(-1.0, alpha=0.0), -100)

    # -- Boundary cases --------------------------------------------------

    def test_exactly_at_threshold(self):
        """Confidence exactly at alpha → passes."""
        self.assertEqual(alpha_cut(50.0, alpha=0.5), 100)

    def test_just_below_threshold(self):
        """Confidence just below alpha → filtered."""
        self.assertEqual(alpha_cut(49.9, alpha=0.5), 0)

    def test_exactly_100(self):
        self.assertEqual(alpha_cut(100.0), 100)

    def test_exactly_minus_100(self):
        self.assertEqual(alpha_cut(-100.0), -100)

    # -- Scale parameter -------------------------------------------------

    def test_custom_scale(self):
        """With scale=50, -40 is 80% confident → passes at alpha=0.5."""
        self.assertEqual(alpha_cut(-40.0, alpha=0.5, scale=50.0), -50)

    def test_invalid_scale(self):
        self.assertEqual(alpha_cut(-87.3, scale=0.0), 0)

    # -- Return type -----------------------------------------------------

    def test_returns_int(self):
        result = alpha_cut(-87.3)
        self.assertIsInstance(result, int)


if __name__ == '__main__':
    unittest.main()
