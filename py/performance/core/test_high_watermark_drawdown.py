import math
import unittest

from .high_watermark_drawdown import HighWaterMarkDrawdown

def expected_drawdowns(returns):
    """
    Independent reference implementation of chronological
    high-water-mark drawdowns.

    Returns are percentages.
    """
    equity = 1.0
    peak = 1.0
    result = []
    for ret in returns:
        equity *= 1.0 + ret * 0.01
        if equity >= peak:
            peak = equity
            dd = 0.0
        else:
            dd = (equity / peak - 1.0) * 100.0
        result.append(dd)
    return result

def expected_rolling_drawdowns(returns):
    """
    Independent reference implementation for HWM drawdowns
    within a rolling window.

    Returns are percentages.

    Important:
        The peak is a running high-water mark, not the maximum
        equity value of the complete window.
    """
    equity = 1.0
    peak = 1.0
    result = []
    for ret in returns:
        equity *= 1.0 + ret * 0.01
        if equity >= peak:
            peak = equity
            dd = 0.0
        else:
            dd = (equity / peak - 1.0) * 100.0
        result.append(dd)
    return result

def assertDrawdownsAlmostEqual(testcase: unittest.TestCase, actual, expected, places=14):
    testcase.assertEqual(len(actual), len(expected))
    for a, e in zip(actual, expected):
        testcase.assertAlmostEqual(a, e, places=places)

def assertState(testcase: unittest.TestCase, accumulator, expected_drawdowns, places=14):
    assertDrawdownsAlmostEqual(testcase, accumulator.drawdowns, expected_drawdowns,
                               places=places)
    testcase.assertEqual(accumulator.drawdowns_count, len(expected_drawdowns))
    if expected_drawdowns:
        testcase.assertAlmostEqual(accumulator.maximum_drawdown, min(expected_drawdowns),
                                   places=places)
        expected_mean = (sum(expected_drawdowns) / len(expected_drawdowns))
        expected_squared_mean = (sum(x * x for x in expected_drawdowns) / len(expected_drawdowns))
        testcase.assertAlmostEqual(accumulator.drawdowns_mean, expected_mean,
                                   places=places)
        testcase.assertAlmostEqual(accumulator.drawdowns_squared_mean, expected_squared_mean,
                                   places=places)
    else:
        testcase.assertTrue(math.isnan(accumulator.maximum_drawdown))
        testcase.assertTrue(math.isnan(accumulator.drawdowns_mean))
        testcase.assertTrue(math.isnan(accumulator.drawdowns_squared_mean))

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
        Returns:

            +10%, +5%, +20%

        Every observation creates a new high-water mark, therefore
        every drawdown is zero.
        """
        acc = HighWaterMarkDrawdown(window_size=0)
        returns = [10.0, 5.0, 20.0]
        for ret in returns:
            acc.update(ret)
        expected = expected_drawdowns(returns)
        assertState(self, acc, expected)
        self.assertEqual(expected, [0.0, 0.0, 0.0])

    def test_expanding_simple_drawdown_and_recovery(self):
        """
        Returns:

            +10%, -5%, +10%

        Equity:

            1.0000
            1.1000    -> HWM, DD = 0
            1.0450    -> DD = -5%
            1.1495    -> new HWM, DD = 0
        """
        acc = HighWaterMarkDrawdown(window_size=0)
        returns = [10.0, -5.0, 10.0]
        for ret in returns:
            acc.update(ret)
        expected = expected_drawdowns(returns)
        assertState(self, acc, expected, places=13)
        self.assertAlmostEqual(expected[1], -5.0, places=13)

    def test_expanding_compounded_drawdown(self):
        """
        Returns:

            +10%, -10%, -10%

        The second and third observations form a compounded decline.

        Starting from 1.10:

            1.10 * 0.90 = 0.99
            0.99 * 0.90 = 0.891

        Relative to the 1.10 peak:

            0.891 / 1.10 - 1 = -19%
        """
        acc = HighWaterMarkDrawdown(window_size=0)
        returns = [10.0, -10.0, -10.0]
        for ret in returns:
            acc.update(ret)
        expected = expected_drawdowns(returns)
        assertState(self, acc, expected, places=12)
        self.assertAlmostEqual(acc.drawdowns[1], -10.0, places=14)
        self.assertAlmostEqual(acc.drawdowns[2], -19.0, places=14)
        self.assertAlmostEqual(acc.maximum_drawdown, -19.0, places=14)

    def test_expanding_new_high_water_mark_resets_drawdown(self):
        """
        Returns:

            +10%, -5%, +6%, -2%

        The +6% return creates a new high-water mark, so the final
        -2% return is measured from that new peak rather than from
        the original +10% peak.
        """
        acc = HighWaterMarkDrawdown(window_size=0)
        returns = [10.0, -5.0, 6.0, -2.0]
        for ret in returns:
            acc.update(ret)
        expected = expected_drawdowns(returns)
        assertState(self, acc, expected, places=13)
        self.assertAlmostEqual(acc.drawdowns[0], 0.0, places=14)
        self.assertAlmostEqual(acc.drawdowns[1], -5.0, places=14)
        self.assertAlmostEqual(acc.drawdowns[2], 0.0, places=14)

        # After +6%, equity is 1.107.
        # After -2%, equity is 1.08486.
        # 1.08486 / 1.107 - 1 = -2%.
        self.assertAlmostEqual(acc.drawdowns[3], -2.0, places=14)

    def test_expanding_reset(self):
        acc = HighWaterMarkDrawdown(window_size=0)
        for ret in [10.0, -5.0, -2.0]:
            acc.update(ret)
        self.assertGreater(acc.drawdowns_count, 0)
        acc.reset()
        assertState(self, acc, [])

        # Verify that the accumulator can be reused.
        acc.update(-5.0)
        assertState(self, acc, [0.0])

    # ------------------------------------------------------------------
    # Rolling-window tests
    # ------------------------------------------------------------------

    def test_rolling_window_peak_eviction_to_new_peak(self):
        """
        Window size = 3.

        Initial window:

            +10%, -5%, -2%

        The +10% observation remains in the window, so the original
        high-water mark remains valid.
        """
        acc = HighWaterMarkDrawdown(window_size=3)
        for ret in [10.0, -5.0, -2.0]:
            acc.update(ret)
        expected = expected_rolling_drawdowns([10.0, -5.0, -2.0])
        assertState(self, acc, expected, places=12)

        # Slide:
        #
        # old window: [+10%, -5%, -2%]
        # new window: [-5%, -2%, +3%]
        #
        # The old +10% peak leaves, so this also tests peak eviction.
        acc.update(3.0)
        expected = expected_rolling_drawdowns([-5.0, -2.0, 3.0])
        assertState(self, acc, expected, places=13)

    def test_rolling_window_without_peak_eviction(self):
        acc = HighWaterMarkDrawdown(window_size=3)

        # +5% becomes the initial peak, then +10% becomes
        # the newer peak.
        for ret in [5.0, -2.0, 10.0]:
            acc.update(ret)

        # Chronological HWM:
        #
        # +5%  ->   0%
        # -2%  ->  -2%
        # +10% ->   0%
        assertState(self, acc, [0.0, -2.0, 0.0], places=14)

        # Slide window:
        #
        # old: [+5%, -2%, +10%]
        # new: [-2%, +10%, -3%]
        #
        # +5% leaves, and it WAS a high-water mark.
        #
        # Therefore this actually IS a peak-eviction case.
        acc.update(-3.0)

        # Recomputed chronological HWM:
        #
        # -2%  ->  0%
        # +10% ->  0%
        # -3%  -> -3%
        assertState(self, acc, [0.0, 0.0, -3.0], places=14)

    def test_rolling_window_peak_eviction_recomputes_drawdowns(self):
        acc = HighWaterMarkDrawdown(window_size=3)

        # Initial window:
        #
        # [+10%, -5%, -5%]
        #
        # HWM = +10%
        #
        # DD:
        #   0%
        #  -5%
        #  -9.75%
        for ret in [10.0, -5.0, -5.0]:
            acc.update(ret)
        assertState(self, acc, [0.0, -5.0, -9.75], places=13)

        # +1% enters.
        #
        # New window:
        #
        # [-5%, -5%, +1%]
        #
        # The old +10% HWM has left.
        #
        # New chronological HWM:
        #
        # -5% -> 0%
        # -5% -> -5%
        # +1% -> -4.05%
        acc.update(1.0)
        assertState(self, acc, [0.0, -5.0, -4.05], places=13)

    def test_rolling_window_multiple_high_water_marks(self):
        """
        Window size = 4.

        Sequence:

            +10%, -5%, +5%, -2%, -3%

        The window is checked after every update against a simple
        recalculation from scratch.
        """
        acc = HighWaterMarkDrawdown(window_size=4)
        returns = [10.0, -5.0, 5.0, -2.0, -3.0]
        for i, ret in enumerate(returns):
            acc.update(ret)
            current = returns[
                max(0, i - 4 + 1):
                i + 1
            ]
            expected = expected_rolling_drawdowns(current)
            assertState(self, acc, expected, places=12)

    def test_rolling_window_new_high_water_mark(self):
        """
        Window size = 3.

        Initial:

            [+5%, -2%, +10%]

        The +10% creates a new high-water mark.

        Slide:

            [-2%, +10%, -5%]

        The +10% remains in the window and therefore remains the
        high-water mark.
        """
        acc = HighWaterMarkDrawdown(window_size=3)
        initial = [5.0, -2.0, 10.0]
        for ret in initial:
            acc.update(ret)
        assertState(self, acc, expected_rolling_drawdowns(initial), places=13)

        acc.update(-5.0)
        new_window = [-2.0, 10.0, -5.0]
        assertState(self,  acc, expected_rolling_drawdowns(new_window))

        # The final -5% is measured relative to the +10% HWM.
        self.assertAlmostEqual(acc.maximum_drawdown, -5.0, places=14)

    def test_rolling_window_peak_leaves_but_another_peak_remains(self):
        """
        Window size = 4.

        Initial:

            [+10%, -2%, +5%, -3%]

        The first +10% is the HWM.

        After adding -1%:

            [-2%, +5%, -3%, -1%]

        The +10% peak has left, so the implementation must recompute
        using the +5% observation as the new HWM.
        """
        acc = HighWaterMarkDrawdown(window_size=4)
        initial = [10.0, -2.0, 5.0, -3.0]
        for ret in initial:
            acc.update(ret)
        assertState(self, acc, expected_rolling_drawdowns(initial), places=12)

        acc.update(-1.0)
        new_window = [-2.0, 5.0, -3.0, -1.0]
        expected = expected_rolling_drawdowns(new_window)
        assertState(self, acc, expected)

    def test_rolling_window_all_negative_returns(self):
        """
        Window size = 3.

        With only negative returns, the first observation in each
        window is the high-water mark.
        """
        acc = HighWaterMarkDrawdown(window_size=3)
        returns = [-1.0, -2.0, -3.0, -4.0]
        for i, ret in enumerate(returns):
            acc.update(ret)
            current = returns[
                max(0, i - 3 + 1):
                i + 1
            ]
            expected = expected_rolling_drawdowns(current)
            assertState(self, acc, expected, places=13)

    def test_rolling_window_zero_size_means_expanding(self):
        """
        window_size=0 must behave as an expanding window.
        """
        acc = HighWaterMarkDrawdown(window_size=0)
        returns = [10.0, -5.0, -2.0, 5.0]
        for ret in returns:
            acc.update(ret)
        expected = expected_drawdowns(returns)
        assertState(self, acc, expected, places=12)
        self.assertEqual(acc.drawdowns_count, len(returns))

    def test_rolling_window_negative_size_means_expanding(self):
        """
        The constructor normalizes non-positive window sizes to zero,
        so a negative window size also means expanding mode.
        """
        acc = HighWaterMarkDrawdown(window_size=-10)
        returns = [10.0, -5.0, -2.0]
        for ret in returns:
            acc.update(ret)
        expected = expected_drawdowns(returns)
        assertState(self, acc, expected, places=12)

if __name__ == "__main__":
    unittest.main()
