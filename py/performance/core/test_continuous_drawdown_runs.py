import math
import unittest

from .continuous_drawdown_runs import ContinuousDrawdownRuns

def dd(*returns):
    """Compounded drawdown of one continuous losing run."""
    compounded = math.prod(1.0 + r * 0.01 for r in returns)
    return (compounded - 1.0) * 100.0

def assertDrawdownsAlmostEqual(testcase: unittest.TestCase, actual, expected,
                               places=12, prefix=""):
    testcase.assertEqual(len(actual), len(expected))
    for i, (a, e) in enumerate(zip(actual, expected)):
        testcase.assertAlmostEqual(a, e, places=places,
                                   msg=f'{prefix} step {i} (actual {a}, expected {e})')

def assertState(testcase: unittest.TestCase, accumulator, expected_drawdowns,
                expected_run_count=None, places=12, prefix=""):
    assertDrawdownsAlmostEqual(testcase, accumulator.drawdowns,
                               expected_drawdowns, places=places, prefix=prefix)
    if expected_run_count is not None:
        testcase.assertEqual(accumulator.run_count, expected_run_count, msg=prefix)
    expected_sum_sq = sum(x * x for x in expected_drawdowns)
    testcase.assertAlmostEqual(accumulator.sum_drawdowns_squared,
                               expected_sum_sq, places=places, msg=prefix)
    testcase.assertAlmostEqual(accumulator.sqrt_sum_drawdowns_squared,
                               math.sqrt(expected_sum_sq), places=places, msg=prefix)

class TestContinuousDrawdownRuns(unittest.TestCase):
    # ------------------------------------------------------------------
    # Expanding-window tests
    # ------------------------------------------------------------------

    def test_expanding_single_losing_run(self):
        """
        Expanding window:

            -1%, -2%, -3%

        All three returns belong to one continuous losing run.
        """
        acc = ContinuousDrawdownRuns()
        acc.update(-1.0)
        assertState(self, acc, [dd(-1.0)], expected_run_count=1)
        acc.update(-2.0)
        assertState(self, acc, [dd(-1.0, -2.0)], expected_run_count=1)
        acc.update(-3.0)
        assertState(self, acc, [dd(-1.0, -2.0, -3.0)], expected_run_count=1)

    def test_expanding_multiple_losing_runs(self):
        """
        Expanding window:

            -1%, -2%, +1%, -3%, -4%, +2%, -5%

        Expected runs:

            [-1%, -2%]
            [-3%, -4%]
            [-5%]
        """
        acc = ContinuousDrawdownRuns()
        returns = [-1.0, -2.0, 1.0, -3.0, -4.0, 2.0, -5.0]
        for ret in returns:
            acc.update(ret)
        expected = [dd(-1.0, -2.0), dd(-3.0, -4.0), dd(-5.0)]
        assertState(self, acc, expected, expected_run_count=3)

    def test_expanding_non_negative_returns_are_separators(self):
        """
        Zero and positive returns must separate losing runs.
        """
        acc = ContinuousDrawdownRuns()
        for ret in [-2.0, 0.0, -3.0, 1.0, -4.0]:
            acc.update(ret)
        expected = [dd(-2.0), dd(-3.0), dd(-4.0)]
        assertState(self, acc, expected, expected_run_count=3)

    def test_expanding_positive_return_does_not_create_drawdown(self):
        acc = ContinuousDrawdownRuns()
        for ret in [2.0, 3.0, 0.0, 5.0]:
            acc.update(ret)
        assertState(self, acc, [], expected_run_count=0)

    def test_reset(self):
        acc = ContinuousDrawdownRuns()
        for ret in [-1.0, -2.0, 1.0, -3.0]:
            acc.update(ret)
        self.assertGreater(acc.run_count, 0)

        acc.reset()
        assertState(self, acc,[], expected_run_count=0)

        # It must also be possible to use it again after reset.
        acc.update(-4.0)
        assertState(self, acc, [dd(-4.0)], expected_run_count=1)

    # ------------------------------------------------------------------
    # Rolling-window tests
    # ------------------------------------------------------------------

    def test_rolling_window_eviction_from_front_of_losing_run(self):
        """
        Window size = 3

        Full sequence:
            -1%, -2%, -3%

        Initial window:
            [-1%, -2%, -3%]

        Then slide to:
            [-2%, -3%]

        The first return is removed from the front of the
        continuous run, so the run must be recomputed as -2%, -3%.
        """
        acc = ContinuousDrawdownRuns()
        for ret in [-1.0, -2.0, -3.0]:
            acc.update(ret)
        assertState(self, acc, [dd(-1.0, -2.0, -3.0)], expected_run_count=1)
        acc.revert(-1.0)
        acc.update(-4.0)
        assertState(self, acc, [dd(-2.0, -3.0, -4.0)], expected_run_count=1)

    def test_rolling_window_eviction_of_entire_losing_run(self):
        """
        Window size = 3

        Initial:
            [-1%, -2%, +1%]

        Runs:
            [-1%, -2%]

        Slide:
            [-2%, +1%, -3%]

        The old -1% is removed, leaving [-2%] as the first run.
        """
        acc = ContinuousDrawdownRuns()
        for ret in [-1.0, -2.0, 1.0]:
            acc.update(ret)
        assertState(self, acc, [dd(-1.0, -2.0)], expected_run_count=1)
        acc.revert(-1.0)
        acc.update(-3.0)
        assertState(self, acc, [dd(-2.0), dd(-3.0)], expected_run_count=2)

    def test_rolling_window_eviction_of_separator(self):
        """
        Window size = 3

        Initial:
            [-1%, +1%, -2%]

        Runs:
            [-1%]
            [-2%]

        Slide:
            [+1%, -2%, -3%]

        The positive separator is removed. The remaining -2%, -3%
        must form one continuous run.
        """
        acc = ContinuousDrawdownRuns()
        for ret in [-1.0, 1.0, -2.0]:
            acc.update(ret)
        assertState(self, acc, [dd(-1.0), dd(-2.0)], expected_run_count=2)
        acc.revert(-1.0)
        acc.update(-3.0)
        assertState(self, acc, [dd(-2.0, -3.0)], expected_run_count=1)

    def test_rolling_window_multiple_runs(self):
        """
        Window size = 5

        Initial:
            [-1%, -2%, +1%, -3%, -4%]

        Runs:
            [-1%, -2%]
            [-3%, -4%]

        Slide 1:
            [-2%, +1%, -3%, -4%, +2%]

        Runs:
            [-2%]
            [-3%, -4%]

        Slide 2:
            [+1%, -3%, -4%, +2%, -5%]

        Runs:
            [-3%, -4%]
            [-5%]
        """
        acc = ContinuousDrawdownRuns()
        initial = [-1.0, -2.0, 1.0, -3.0, -4.0]
        for ret in initial:
            acc.update(ret)
        assertState(self, acc, [dd(-1.0, -2.0), dd(-3.0, -4.0)], expected_run_count=2)

        # Slide 1: remove -1%, add +2%
        acc.revert(-1.0)
        acc.update(2.0)
        assertState(self, acc, [dd(-2.0), dd(-3.0, -4.0)], expected_run_count=2)

        # Slide 2: remove -2%, add -5%
        acc.revert(-2.0)
        acc.update(-5.0)
        assertState(self, acc, [dd(-3.0, -4.0), dd(-5.0)], expected_run_count=2)

    def test_rolling_window_new_return_extends_existing_run(self):
        """
        Initial:
            [+1%, -2%, -3%]

        Slide:
            [-2%, -3%, -4%]

        The new -4% extends the existing right-most run.
        """
        acc = ContinuousDrawdownRuns()
        for ret in [1.0, -2.0, -3.0]:
            acc.update(ret)
        assertState(self, acc, [dd(-2.0, -3.0)], expected_run_count=1)
        acc.revert(1.0)
        acc.update(-4.0)
        assertState(self, acc, [dd(-2.0, -3.0, -4.0)], expected_run_count=1)

    def test_rolling_window_new_negative_starts_new_run_after_separator(self):
        """
        Initial:
            [-2%, +1%, -3%]

        Slide:
            [+1%, -3%, -4%]

        The new -4% extends the -3% run rather than starting
        a separate run.
        """
        acc = ContinuousDrawdownRuns()
        for ret in [-2.0, 1.0, -3.0]:
            acc.update(ret)
        assertState(self, acc, [dd(-2.0), dd(-3.0)], expected_run_count=2)
        acc.revert(-2.0)
        acc.update(-4.0)
        assertState(self, acc, [dd(-3.0, -4.0)], expected_run_count=1)

    def test_revert_then_update_order_is_required(self):
        """
        Explicitly exercise the intended rolling-window protocol:

            revert(old)
            update(new)

        The resulting state must correspond exactly to the new window.
        """
        acc = ContinuousDrawdownRuns()
        window = [-1.0, -2.0, 2.0]
        for ret in window:
            acc.update(ret)
        assertState(self, acc, [dd(-1.0, -2.0)], expected_run_count=1)

        # New window: [-2%, +2%, -3%]
        acc.revert(-1.0)
        acc.update(-3.0)
        assertState(self, acc, [dd(-2.0), dd(-3.0)], expected_run_count=2)

    # ------------------------------------------------------------------
    # Numerical consistency
    # ------------------------------------------------------------------

    def test_sqrt_sum_drawdowns_squared(self):
        """
        Check the Burke denominator directly.

        Runs:
            [-1%, -2%]
            [-3%]
        """
        acc = ContinuousDrawdownRuns()
        for ret in [-1.0, -2.0, 1.0, -3.0]:
            acc.update(ret)

        dd1 = dd(-1.0, -2.0)
        dd2 = dd(-3.0)

        expected_sum_sq = dd1 ** 2 + dd2 ** 2
        expected_sqrt = math.sqrt(expected_sum_sq)

        self.assertAlmostEqual(acc.sum_drawdowns_squared, expected_sum_sq, places=12)
        self.assertAlmostEqual(acc.sqrt_sum_drawdowns_squared, expected_sqrt, places=12)

if __name__ == "__main__":
    unittest.main()
