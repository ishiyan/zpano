"""Tests for signal composition."""
from __future__ import annotations

import unittest

from py.signals.compose import signal_and, signal_or, signal_not, signal_strength


class TestSignalAnd(unittest.TestCase):

    def test_all_high(self):
        self.assertAlmostEqual(signal_and(0.9, 0.8, 0.95), 0.9 * 0.8 * 0.95)

    def test_one_zero(self):
        self.assertAlmostEqual(signal_and(0.9, 0.0, 0.8), 0.0)

    def test_all_one(self):
        self.assertAlmostEqual(signal_and(1.0, 1.0, 1.0), 1.0)

    def test_two_args(self):
        self.assertAlmostEqual(signal_and(0.6, 0.7), 0.42)


class TestSignalOr(unittest.TestCase):

    def test_both_high(self):
        result = signal_or(0.8, 0.9)
        self.assertAlmostEqual(result, 0.8 + 0.9 - 0.8 * 0.9)

    def test_one_zero(self):
        self.assertAlmostEqual(signal_or(0.0, 0.7), 0.7)

    def test_both_zero(self):
        self.assertAlmostEqual(signal_or(0.0, 0.0), 0.0)

    def test_both_one(self):
        self.assertAlmostEqual(signal_or(1.0, 1.0), 1.0)

    def test_greater_than_either(self):
        """OR result >= max of inputs."""
        a, b = 0.6, 0.7
        self.assertGreaterEqual(signal_or(a, b), max(a, b))


class TestSignalNot(unittest.TestCase):

    def test_zero(self):
        self.assertAlmostEqual(signal_not(0.0), 1.0)

    def test_one(self):
        self.assertAlmostEqual(signal_not(1.0), 0.0)

    def test_half(self):
        self.assertAlmostEqual(signal_not(0.5), 0.5)

    def test_complement(self):
        for v in [0.0, 0.3, 0.5, 0.7, 1.0]:
            self.assertAlmostEqual(signal_not(v), 1.0 - v)


class TestSignalStrength(unittest.TestCase):

    def test_above_threshold(self):
        self.assertEqual(signal_strength(0.8, 0.5), 0.8)

    def test_below_threshold(self):
        self.assertEqual(signal_strength(0.3, 0.5), 0.0)

    def test_at_threshold(self):
        self.assertEqual(signal_strength(0.5, 0.5), 0.5)

    def test_just_below(self):
        self.assertEqual(signal_strength(0.499, 0.5), 0.0)

    def test_default_threshold(self):
        self.assertEqual(signal_strength(0.6), 0.6)
        self.assertEqual(signal_strength(0.4), 0.0)


if __name__ == '__main__':
    unittest.main()
