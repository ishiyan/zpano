import math
import random
import unittest

from .high_watermark_drawdown import HighWaterMarkDrawdown

def expected_drawdowns(returns):
    """
    Independent reference implementation of chronological
    high-water-mark drawdowns, matching R PerformanceAnalytics
    Drawdowns(): the high-water mark starts at the initial equity 1.

    Returns and drawdowns are decimals.
    """
    equity = 1.0
    peak = 1.0
    result = []
    for ret in returns:
        equity *= 1.0 + ret
        if equity >= peak:
            peak = equity
            dd = 0.0
        else:
            dd = equity / peak - 1.0
        result.append(dd)
    return result

def expected_rolling_drawdowns(returns, window_size, i):
    """
    Drawdowns of the rolling window ending at index i:
    a fresh calculation over the returns in the window.
    """
    lo = 0 if window_size <= 0 else max(0, i - window_size + 1)
    return expected_drawdowns(returns[lo:i + 1])

def assertDrawdownsAlmostEqual(testcase: unittest.TestCase, actual, expected, places=14):
    testcase.assertEqual(len(actual), len(expected))
    for a, e in zip(actual, expected):
        testcase.assertAlmostEqual(a, e, places=places)

def assertState(testcase: unittest.TestCase, accumulator, expected_drawdowns, places=14):
    assertDrawdownsAlmostEqual(testcase, accumulator.drawdowns, expected_drawdowns,
                               places=places)
    testcase.assertEqual(accumulator.drawdowns_count, len(expected_drawdowns))
    if expected_drawdowns:
        testcase.assertAlmostEqual(accumulator.drawdown, expected_drawdowns[-1],
                                   places=places)
        testcase.assertAlmostEqual(accumulator.maximum_drawdown, min(expected_drawdowns),
                                   places=places)
        expected_mean = (sum(expected_drawdowns) / len(expected_drawdowns))
        expected_squared_mean = (sum(x * x for x in expected_drawdowns) / len(expected_drawdowns))
        testcase.assertAlmostEqual(accumulator.drawdowns_mean, expected_mean,
                                   places=places)
        testcase.assertAlmostEqual(accumulator.drawdowns_squared_mean, expected_squared_mean,
                                   places=places)
    else:
        testcase.assertTrue(math.isnan(accumulator.drawdown))
        testcase.assertTrue(math.isnan(accumulator.maximum_drawdown))
        testcase.assertTrue(math.isnan(accumulator.drawdowns_mean))
        testcase.assertTrue(math.isnan(accumulator.drawdowns_squared_mean))

def feed(accumulator, returns):
    for ret in returns:
        accumulator.update(ret)
    return accumulator

class TestHighWaterMarkDrawdown(unittest.TestCase):

    # ------------------------------------------------------------------
    # Expanding-window tests
    # ------------------------------------------------------------------

    def test_expanding_empty(self):
        """An expanding accumulator starts empty."""
        acc = HighWaterMarkDrawdown(window_size=0)
        assertState(self, acc, [])

    def test_expanding_all_positive_returns(self):
        """
        Returns: +10%, +5%, +20%

        Every observation creates a new high-water mark, therefore
        every drawdown is zero.
        """
        acc = feed(HighWaterMarkDrawdown(window_size=0), [0.10, 0.05, 0.20])
        assertState(self, acc, [0.0, 0.0, 0.0])

    def test_expanding_first_negative_return(self):
        """
        Returns: -5%, -2%, +10%

        The high-water mark starts at the initial equity 1.0, so the
        first negative return is already a drawdown (as in R).

            0.95     -> DD = -5%
            0.931    -> DD = -6.9%
            1.0241   -> new HWM, DD = 0
        """
        acc = feed(HighWaterMarkDrawdown(window_size=0), [-0.05, -0.02, 0.10])
        assertState(self, acc, [-0.05, -0.069, 0.0])

    def test_expanding_simple_drawdown_and_recovery(self):
        """
        Returns: +10%, -5%, +10%

        Equity:

            1.1000    -> HWM, DD = 0
            1.0450    -> DD = -5%
            1.1495    -> new HWM, DD = 0
        """
        acc = feed(HighWaterMarkDrawdown(window_size=0), [0.10, -0.05, 0.10])
        assertState(self, acc, [0.0, -0.05, 0.0])

    def test_expanding_compounded_drawdown(self):
        """
        Returns: +10%, -10%, -10%

        Starting from 1.10:

            1.10 * 0.90 = 0.99
            0.99 * 0.90 = 0.891

        Relative to the 1.10 peak: 0.891 / 1.10 - 1 = -19%
        """
        acc = feed(HighWaterMarkDrawdown(window_size=0), [0.10, -0.10, -0.10])
        assertState(self, acc, [0.0, -0.10, -0.19])

    def test_expanding_new_high_water_mark_resets_drawdown(self):
        """
        Returns: +10%, -5%, +6%, -2%

        After +6%, equity is 1.1077 > 1.1, a new high-water mark,
        so the final -2% is measured from that new peak.
        """
        acc = feed(HighWaterMarkDrawdown(window_size=0), [0.10, -0.05, 0.06, -0.02])
        assertState(self, acc, [0.0, -0.05, 0.0, -0.02])

    def test_expanding_reset(self):
        acc = feed(HighWaterMarkDrawdown(window_size=0), [0.10, -0.05, -0.02])
        self.assertGreater(acc.drawdowns_count, 0)
        acc.reset()
        assertState(self, acc, [])

        # The accumulator can be reused, starting from equity 1.0 again.
        acc.update(-0.05)
        assertState(self, acc, [-0.05])

    def test_expanding_matches_reference(self):
        rng = random.Random(42)
        returns = [rng.gauss(0.0, 0.03) for _ in range(200)]
        acc = feed(HighWaterMarkDrawdown(window_size=0), returns)
        assertState(self, acc, expected_drawdowns(returns), places=13)

    def test_zero_size_means_expanding(self):
        returns = [0.10, -0.05, -0.02, 0.05]
        acc = feed(HighWaterMarkDrawdown(window_size=0), returns)
        assertState(self, acc, expected_drawdowns(returns))

    def test_negative_size_means_expanding(self):
        """
        The constructor normalizes non-positive window sizes to zero,
        so a negative window size also means expanding mode.
        """
        returns = [0.10, -0.05, -0.02]
        acc = feed(HighWaterMarkDrawdown(window_size=-10), returns)
        assertState(self, acc, expected_drawdowns(returns))

    # ------------------------------------------------------------------
    # Rolling-window tests: the window equals a fresh calculation over
    # its returns, starting from the equity just before the window.
    # ------------------------------------------------------------------

    def test_rolling_window_peak_eviction(self):
        """
        Window size = 3.

        [+10%, -5%, -2%] -> [0, -5%, -6.9%]

        Slide to [-5%, -2%, +3%]: the +10% peak leaves the window.
        Fresh calculation:

            0.95     -> -5%
            0.931    -> -6.9%
            0.95893  -> -4.107%
        """
        acc = feed(HighWaterMarkDrawdown(window_size=3), [0.10, -0.05, -0.02])
        assertState(self, acc, [0.0, -0.05, -0.069])
        acc.update(0.03)
        assertState(self, acc, [-0.05, -0.069, -0.04107])

    def test_rolling_window_evicted_peak_followed_by_new_peak(self):
        """
        Window size = 3.

        [+5%, -2%, +10%] -> [0, -2%, 0]

        Slide to [-2%, +10%, -3%]:

            0.98     -> -2%
            1.078    -> 0 (new HWM)
            1.04566  -> -3%
        """
        acc = feed(HighWaterMarkDrawdown(window_size=3), [0.05, -0.02, 0.10])
        assertState(self, acc, [0.0, -0.02, 0.0])
        acc.update(-0.03)
        assertState(self, acc, [-0.02, 0.0, -0.03])

    def test_rolling_window_peak_eviction_recomputes_drawdowns(self):
        """
        Window size = 3.

        [+10%, -5%, -5%] -> [0, -5%, -9.75%]

        Slide to [-5%, -5%, +1%]:

            0.95      -> -5%
            0.9025    -> -9.75%
            0.911525  -> -8.8475%
        """
        acc = feed(HighWaterMarkDrawdown(window_size=3), [0.10, -0.05, -0.05])
        assertState(self, acc, [0.0, -0.05, -0.0975])
        acc.update(0.01)
        assertState(self, acc, [-0.05, -0.0975, -0.088475])

    def test_rolling_window_all_negative_returns(self):
        returns = [-0.01, -0.02, -0.03, -0.04]
        acc = HighWaterMarkDrawdown(window_size=3)
        for i, ret in enumerate(returns):
            acc.update(ret)
            assertState(self, acc, expected_rolling_drawdowns(returns, 3, i))

    def test_rolling_window_size_one(self):
        returns = [0.10, -0.05, -0.02, 0.03, -0.04]
        acc = HighWaterMarkDrawdown(window_size=1)
        for i, ret in enumerate(returns):
            acc.update(ret)
            assertState(self, acc, [min(ret, 0.0)])

    def test_rolling_window_matches_fresh_calculation(self):
        rng = random.Random(7)
        for window_size in (2, 3, 4, 7, 20):
            for _ in range(10):
                returns = [rng.gauss(0.0, 0.03) for _ in range(80)]
                acc = HighWaterMarkDrawdown(window_size=window_size)
                for i, ret in enumerate(returns):
                    acc.update(ret)
                    with self.subTest(window_size=window_size, step=i):
                        assertState(self, acc,
                            expected_rolling_drawdowns(returns, window_size, i), places=13)

    def test_rolling_window_recompute_flag(self):
        """
        Evicting an observation can only change the remaining high-water
        marks if its return was negative (its equity is below the window's
        starting equity).
        """
        acc = HighWaterMarkDrawdown(window_size=2)
        self.assertFalse(acc.update(0.10))
        self.assertFalse(acc.update(-0.05))
        # Evicts +10%: the remaining peaks don't change.
        self.assertFalse(acc.update(-0.02))
        # Evicts -5% and the new first observation (-2%) is also below
        # the old starting equity: recompute.
        self.assertTrue(acc.update(0.01))
        assertState(self, acc, [-0.02, -0.0102])
        # Evicts -2%; the new first observation (equity 1.0343) is still
        # below the old starting equity (1.045): recompute.
        self.assertTrue(acc.update(0.03))
        assertState(self, acc, [0.0, 0.0])
        # Evicts +1%: its equity is above the window's starting equity,
        # so the remaining peaks don't change.
        self.assertFalse(acc.update(-0.01))
        assertState(self, acc, [0.0, -0.01])

if __name__ == "__main__":
    unittest.main()
