"""Smoke tests for the Corona spectral analysis engine."""

import math
import unittest

from .corona import Corona
from .params import CoronaParams

from .test_testdata import _talib_input


class TestCoronaDefaultSmoke(unittest.TestCase):
    """Smoke tests matching Go's corona_smoke_test.go."""

    def test_filter_bank_length(self):
        c = Corona()
        self.assertEqual(c.filter_bank_length, 49)

    def test_half_period_range(self):
        c = Corona()
        self.assertEqual(c.minimal_period_times_two, 12)
        self.assertEqual(c.maximal_period_times_two, 60)

    def test_primes_at_correct_sample(self):
        c = Corona()
        inp = _talib_input()
        primed_at = -1
        for i, v in enumerate(inp):
            c.update(v)
            if c.is_primed() and primed_at < 0:
                primed_at = i
        self.assertGreaterEqual(primed_at, 0, "engine never primed over 252 samples")
        # primedAt is 0-based index; Go checks primedAt+1 == MinimalPeriodTimesTwo
        self.assertEqual(primed_at + 1, c.minimal_period_times_two)

    def test_final_values_finite_and_in_range(self):
        c = Corona()
        for v in _talib_input():
            c.update(v)

        dc = c.dominant_cycle
        dcm = c.dominant_cycle_median
        self.assertFalse(math.isnan(dc) or math.isinf(dc), f"DominantCycle = {dc}")
        self.assertFalse(math.isnan(dcm) or math.isinf(dcm), f"DominantCycleMedian = {dcm}")

        min_p = float(c.minimal_period)
        max_p = float(c.maximal_period)
        self.assertGreaterEqual(dc, min_p)
        self.assertLessEqual(dc, max_p)
        self.assertGreaterEqual(dcm, min_p)
        self.assertLessEqual(dcm, max_p)

    def test_maximal_amplitude_squared_positive(self):
        c = Corona()
        for v in _talib_input():
            c.update(v)
        m = c.maximal_amplitude_squared
        self.assertGreater(m, 0)
        self.assertFalse(math.isnan(m) or math.isinf(m))

    def test_dc_exceeds_minimal_period(self):
        c = Corona()
        saw_above_min = False
        for v in _talib_input():
            c.update(v)
            if c.is_primed() and c.dominant_cycle > float(c.minimal_period):
                saw_above_min = True
        self.assertTrue(saw_above_min,
                        "DominantCycle never exceeded MinimalPeriod across 252 samples")


class TestCoronaNaN(unittest.TestCase):
    """NaN input is a no-op."""

    def test_nan_preserves_state(self):
        c = Corona()
        for v in _talib_input()[:20]:
            c.update(v)
        self.assertTrue(c.is_primed())

        dc_before = c.dominant_cycle
        dcm_before = c.dominant_cycle_median

        result = c.update(float('nan'))
        self.assertTrue(result)
        self.assertEqual(c.dominant_cycle, dc_before)
        self.assertEqual(c.dominant_cycle_median, dcm_before)


class TestCoronaInvalidParams(unittest.TestCase):
    """Parameter validation."""

    def test_cutoff_too_small(self):
        with self.assertRaises(ValueError):
            Corona(CoronaParams(high_pass_filter_cutoff=1))

    def test_min_too_small(self):
        with self.assertRaises(ValueError):
            Corona(CoronaParams(minimal_period=1))

    def test_max_le_min(self):
        with self.assertRaises(ValueError):
            Corona(CoronaParams(minimal_period=10, maximal_period=10))

    def test_negative_db_lower(self):
        with self.assertRaises(ValueError):
            Corona(CoronaParams(decibels_lower_threshold=-1))

    def test_db_upper_le_lower(self):
        with self.assertRaises(ValueError):
            Corona(CoronaParams(decibels_lower_threshold=6, decibels_upper_threshold=6))


if __name__ == '__main__':
    unittest.main()
