import math
import random
import unittest

from .capture import Capture


def nan_div(a: float, b: float) -> float:
    return a / b if b != 0 else math.nan


def reference(pairs) -> dict:
    """
    Naive capture ratios following the PerformanceAnalytics conventions:
    upside periods have benchmark > 0; downside capture and down-number
    use benchmark <= 0; down-percentage uses benchmark < 0.
    """
    up = [(a, b) for a, b in pairs if b > 0]
    dn = [(a, b) for a, b in pairs if b <= 0]
    dn_strict = [(a, b) for a, b in pairs if b < 0]
    return {
        'upside_capture_ratio_geometric': nan_div(
            math.prod(1 + a for a, _ in up) - 1, math.prod(1 + b for _, b in up) - 1),
        'upside_capture_ratio_arithmetic': nan_div(
            sum(a for a, _ in up), sum(b for _, b in up)),
        'downside_capture_ratio_geometric': nan_div(
            math.prod(1 + a for a, _ in dn) - 1, math.prod(1 + b for _, b in dn) - 1),
        'downside_capture_ratio_arithmetic': nan_div(
            sum(a for a, _ in dn), sum(b for _, b in dn)),
        'up_number_ratio': nan_div(sum(1 for a, _ in up if a > 0), len(up)),
        'down_number_ratio': nan_div(sum(1 for a, _ in dn if a < 0), len(dn)),
        'up_percentage_ratio': nan_div(sum(1 for a, b in up if a > b), len(up)),
        'down_percentage_ratio': nan_div(sum(1 for a, b in dn_strict if a > b), len(dn_strict)),
    }


class TestCapture(unittest.TestCase):

    def assert_matches(self, c: Capture, pairs, places=13, msg=''):
        for name, expected in reference(pairs).items():
            actual = getattr(c, name)
            if math.isnan(expected):
                self.assertTrue(math.isnan(actual), msg=f'{msg} {name}: {actual}')
            else:
                self.assertAlmostEqual(actual, expected, places=places, msg=f'{msg} {name}')

    def test_empty(self):
        c = Capture()
        for name in reference([]):
            self.assertTrue(math.isnan(getattr(c, name)), msg=name)

    def test_hand_computed(self):
        pairs = [(0.02, 0.01), (-0.01, -0.02), (0.03, 0.04), (-0.03, 0.0), (0.01, -0.01)]
        c = Capture()
        for a, b in pairs:
            c.update(a, b)
        # Up periods (b > 0): (0.02, 0.01), (0.03, 0.04)
        self.assertAlmostEqual(c.upside_capture_ratio_arithmetic, 0.05 / 0.05, places=15)
        self.assertAlmostEqual(c.upside_capture_ratio_geometric,
                               (1.02 * 1.03 - 1) / (1.01 * 1.04 - 1), places=14)
        self.assertEqual(c.up_number_ratio, 1.0)
        self.assertEqual(c.up_percentage_ratio, 0.5)
        # Down periods (b <= 0) include the zero-benchmark period.
        self.assertAlmostEqual(c.downside_capture_ratio_arithmetic,
                               (-0.01 - 0.03 + 0.01) / (-0.02 + 0.0 - 0.01), places=15)
        self.assertAlmostEqual(c.down_number_ratio, 2 / 3, places=15)
        # Down-percentage only counts strictly negative benchmark periods.
        self.assertEqual(c.down_percentage_ratio, 1.0)

    def test_matches_reference(self):
        rng = random.Random(42)
        pairs = [(rng.choice([0.0, rng.gauss(0.0, 0.03)]), rng.choice([0.0, rng.gauss(0.0, 0.03)]))
                 for _ in range(200)]
        c = Capture()
        for a, b in pairs:
            c.update(a, b)
        self.assert_matches(c, pairs, places=12)

    def test_rolling_window_matches_reference(self):
        rng = random.Random(7)
        pairs = [(rng.choice([0.0, rng.gauss(0.0, 0.03)]), rng.choice([0.0, rng.gauss(0.0, 0.03)]))
                 for _ in range(120)]
        w = 9
        c = Capture()
        for i, (a, b) in enumerate(pairs):
            if i >= w:
                c.revert(*pairs[i - w])
            c.update(a, b)
            self.assert_matches(c, pairs[max(0, i - w + 1):i + 1], places=12, msg=f'step {i}')

    def test_reset(self):
        c = Capture()
        c.update(0.01, 0.02)
        c.update(-0.01, -0.02)
        c.reset()
        for name in reference([]):
            self.assertTrue(math.isnan(getattr(c, name)), msg=name)


if __name__ == "__main__":
    unittest.main()
