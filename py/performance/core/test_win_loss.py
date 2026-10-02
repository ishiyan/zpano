import math
import random
import unittest

from .win_loss import WinLoss


def reference(returns) -> dict:
    def mean(values):
        return sum(values) / len(values) if values else math.nan
    wins = [r for r in returns if r > 0]
    losses = [r for r in returns if r < 0]
    non_zero = [r for r in returns if r != 0]
    return {
        'non_zero_returns_count': len(non_zero),
        'non_zero_returns_mean': mean(non_zero),
        'winning_returns_count': len(wins),
        'winning_returns_sum': sum(wins),
        'winning_returns_mean': mean(wins),
        'losing_returns_count': len(losses),
        'losing_returns_sum': sum(losses),
        'losing_returns_mean': mean(losses),
    }


class TestWinLoss(unittest.TestCase):

    def assert_matches(self, wl: WinLoss, returns, places=15, msg=''):
        for name, expected in reference(returns).items():
            actual = getattr(wl, name)
            if isinstance(expected, float) and math.isnan(expected):
                self.assertTrue(math.isnan(actual), msg=f'{msg} {name}: {actual}')
            else:
                self.assertAlmostEqual(actual, expected, places=places, msg=f'{msg} {name}')

    def test_empty(self):
        self.assert_matches(WinLoss(), [])

    def test_hand_computed(self):
        wl = WinLoss()
        for r in [0.02, 0.0, -0.01, 0.04, 0.0, -0.03]:
            wl.update(r)
        self.assertEqual(wl.non_zero_returns_count, 4)
        self.assertAlmostEqual(wl.non_zero_returns_mean, 0.02 / 4, places=16)
        self.assertEqual(wl.winning_returns_count, 2)
        self.assertAlmostEqual(wl.winning_returns_mean, 0.03, places=16)
        self.assertEqual(wl.losing_returns_count, 2)
        self.assertAlmostEqual(wl.losing_returns_mean, -0.02, places=16)

    def test_rolling_window_matches_reference(self):
        rng = random.Random(42)
        returns = [rng.choice([0.0, rng.gauss(0.0, 0.03)]) for _ in range(120)]
        w = 6
        wl = WinLoss()
        for i, r in enumerate(returns):
            if i >= w:
                wl.revert(returns[i - w])
            wl.update(r)
            self.assert_matches(wl, returns[max(0, i - w + 1):i + 1], msg=f'step {i}')

    def test_reset(self):
        wl = WinLoss()
        for r in (0.01, -0.02):
            wl.update(r)
        wl.reset()
        self.assert_matches(wl, [])


if __name__ == "__main__":
    unittest.main()
