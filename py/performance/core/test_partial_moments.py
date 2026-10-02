import math
import random
import unittest

from .partial_moments import PartialMoments
from .partial_moments_raw import RawPartialMoments


def mean(values) -> float:
    return sum(values) / len(values) if values else math.nan


def reference(returns, threshold) -> dict:
    """Naive partial moments about the threshold."""
    lower = [threshold - r for r in returns if r < threshold]
    upper = [r - threshold for r in returns if r > threshold]
    n = len(returns)
    expected = {
        'total_count': n,
        'lower_excess_count': len(lower),
        'upper_excess_count': len(upper),
        'downside_frequency': len(lower) / n if n else math.nan,
        'upside_frequency': len(upper) / n if n else math.nan,
        'downside_potential': mean([max(threshold - r, 0.0) for r in returns]),
        'upper_excess_moment_1_sum': sum(upper),
        'upper_excess_moment_2_sum': sum(x * x for x in upper),
        'lower_excess_moment_2_sum': sum(x * x for x in lower),
    }
    for k in (1, 2, 3, 4):
        expected[f'lower_partial_moment_{k}'] = mean([max(threshold - r, 0.0) ** k for r in returns])
        expected[f'higher_partial_moment_{k}'] = mean([max(r - threshold, 0.0) ** k for r in returns])
        expected[f'upper_excess_moment_{k}'] = mean([x ** k for x in upper])
        expected[f'lower_excess_moment_{k}'] = mean([x ** k for x in lower])
    return expected


def raw_reference(returns) -> dict:
    """Naive raw partial moments (sums, threshold 0)."""
    return {
        'count': len(returns),
        'lower_partial_moment_1': sum(max(-r, 0.0) for r in returns),
        'higher_partial_moment_1': sum(max(r, 0.0) for r in returns),
        'count_negative': sum(1 for r in returns if r < 0),
        'sum_negative': sum(r for r in returns if r < 0),
        'count_positive': sum(1 for r in returns if r > 0),
        'sum_positive': sum(r for r in returns if r > 0),
    }


class AssertMixin:
    def assert_matches(self, obj, expected: dict, places=14, msg=''):
        for name, e in expected.items():
            a = getattr(obj, name)
            if isinstance(e, float) and math.isnan(e):
                self.assertTrue(math.isnan(a), msg=f'{msg} {name}: {a}')
            else:
                self.assertAlmostEqual(a, e, places=places, msg=f'{msg} {name}')


def random_returns(seed: int, n: int, threshold: float) -> list:
    # Mix in values exactly equal to the threshold and to zero.
    rng = random.Random(seed)
    return [rng.choice([threshold, 0.0, rng.gauss(0.0, 0.03), rng.gauss(0.0, 0.03)])
            for _ in range(n)]


class TestPartialMoments(AssertMixin, unittest.TestCase):

    def test_empty(self):
        pm = PartialMoments(threshold=0.01)
        self.assertEqual(pm.total_count, 0)
        self.assertTrue(math.isnan(pm.downside_frequency))
        self.assertTrue(math.isnan(pm.upside_frequency))
        self.assertTrue(math.isnan(pm.lower_partial_moment_2))

    def test_hand_computed(self):
        pm = PartialMoments(threshold=0.01)
        for r in [0.03, -0.01, 0.01, 0.00]:
            pm.update(r)
        # Shortfalls below 1%: 0.02, 0.01; excesses above: 0.02.
        self.assertAlmostEqual(pm.lower_partial_moment_1, (0.02 + 0.01) / 4, places=16)
        self.assertAlmostEqual(pm.lower_partial_moment_2, (0.0004 + 0.0001) / 4, places=16)
        self.assertAlmostEqual(pm.higher_partial_moment_1, 0.02 / 4, places=16)
        self.assertEqual(pm.lower_excess_count, 2)
        self.assertEqual(pm.upper_excess_count, 1)
        self.assertEqual(pm.downside_frequency, 0.5)
        self.assertEqual(pm.upside_frequency, 0.25)

    def test_matches_reference(self):
        for threshold in (0.0, 0.005):
            returns = random_returns(42, 200, threshold)
            pm = PartialMoments(threshold=threshold)
            for r in returns:
                pm.update(r)
            self.assert_matches(pm, reference(returns, threshold), msg=f'threshold {threshold}')

    def test_rolling_window_matches_reference(self):
        threshold = 0.005
        returns = random_returns(7, 120, threshold)
        w = 8
        pm = PartialMoments(threshold=threshold)
        for i, r in enumerate(returns):
            if i >= w:
                pm.revert(returns[i - w])
            pm.update(r)
            window = returns[max(0, i - w + 1):i + 1]
            self.assert_matches(pm, reference(window, threshold), places=13, msg=f'step {i}')

    def test_reset(self):
        pm = PartialMoments(threshold=0.0)
        for r in (0.01, -0.02):
            pm.update(r)
        pm.reset()
        self.assertEqual(pm.total_count, 0)
        self.assertEqual(pm.lower_excess_count, 0)
        self.assertEqual(pm.upper_excess_count, 0)


class TestRawPartialMoments(AssertMixin, unittest.TestCase):

    def test_empty(self):
        self.assert_matches(RawPartialMoments(), raw_reference([]))

    def test_matches_reference(self):
        returns = random_returns(42, 200, 0.0)
        pm = RawPartialMoments()
        for r in returns:
            pm.update(r)
        self.assert_matches(pm, raw_reference(returns))

    def test_rolling_window_matches_reference(self):
        returns = random_returns(7, 120, 0.0)
        w = 8
        pm = RawPartialMoments()
        for i, r in enumerate(returns):
            if i >= w:
                pm.revert(returns[i - w])
            pm.update(r)
            self.assert_matches(pm, raw_reference(returns[max(0, i - w + 1):i + 1]),
                                msg=f'step {i}')

    def test_reset(self):
        pm = RawPartialMoments()
        for r in (0.01, -0.02):
            pm.update(r)
        pm.reset()
        self.assert_matches(pm, raw_reference([]))


if __name__ == "__main__":
    unittest.main()
