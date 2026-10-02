import math
import random
import unittest

from .cumulative_return import CumulativeReturn


class TestCumulativeReturn(unittest.TestCase):

    def test_annualized_return_definition(self):
        returns = [0.10, -0.05, 0.03, 0.08]
        periods_per_year = 12

        cr = CumulativeReturn()
        for r in returns:
            cr.update(r)

        growth = math.prod(1 + r for r in returns)
        expected = growth ** (periods_per_year / len(returns)) - 1
        actual = cr.annualized_geometric_mean_return(periods_per_year)
        self.assertAlmostEqual(actual, expected, places=15)

    def test_one_month_return(self):
        """
        One monthly return of 1%, expected (1.01)^12-1.
        """
        cr = CumulativeReturn()
        cr.update(0.01)

        expected = 1.01**12 - 1
        actual = cr.annualized_geometric_mean_return(12)
        self.assertAlmostEqual(actual, expected, places=15)

    def test_yearly_returns(self):
        """
        If the observation itself is yearly, then the annualized
        geometric meanreturn equals the reometric mean return.
        """
        returns = [0.12, -0.04, 0.08]

        cr = CumulativeReturn()
        for r in returns:
            cr.update(r)

        expected = cr.geometric_mean_return
        actual = cr.annualized_geometric_mean_return(1)
        self.assertAlmostEqual(actual, expected, places=15)

    def test_constant_monthly_return(self):
        """
        If every monthly return is exactly $r$,
        $$((1+r)^n)^{12/n}=(1+r)^12$$.

        Notice the number of observations cancels completely.
        This is an excellent invariant.
        """
        cr = CumulativeReturn()

        for _ in range(60):
            cr.update(0.01)

        expected = 1.01**12 - 1
        actual = cr.annualized_geometric_mean_return(12)
        self.assertAlmostEqual(actual, expected, places=15)

    def test_empty(self):
        """
        Empty accumulator
        """
        cr = CumulativeReturn()
        actual = math.isnan(cr.annualized_geometric_mean_return(12))
        self.assertTrue(actual)

    def test_zero_returns(self):
        """
        If every return is zero, the annualized geometric mean return must be
        exactly zero because `log1p(0) == 0` and the sum remains exactly zero `expm1(0) == 0`.
        """
        cr = CumulativeReturn()

        for _ in range(100):
            cr.update(0.0)

        expected = 0
        actual = cr.annualized_geometric_mean_return(252)
        self.assertAlmostEqual(actual, expected, places=15)


    def test_consistency_with_geometric_mean_return(self):
        """
        Tests $1+annualized=(1+geometric mean)^p$
        """
        returns = [0.0010, -0.0005, 0.0003, 0.0008]
        periods_per_year = 252

        cr = CumulativeReturn()
        for r in returns:
            cr.update(r)

        expected = (1 + cr.geometric_mean_return) ** periods_per_year - 1
        actual = cr.annualized_geometric_mean_return(periods_per_year)
        self.assertAlmostEqual(actual, expected, places=13)

    def test_rolling_window_matches_fresh_calculation(self):
        rng = random.Random(42)
        returns = [rng.choice([0.0, rng.gauss(0.0, 0.03)]) for _ in range(100)]
        window_size = 7
        cr = CumulativeReturn()
        for i, r in enumerate(returns):
            if i >= window_size:
                cr.revert(returns[i - window_size])
            cr.update(r)
            window = returns[max(0, i - window_size + 1):i + 1]
            growth = math.prod(1 + x for x in window)
            self.assertEqual(cr.count, len(window))
            self.assertAlmostEqual(cr.cumulative_geometric_return, growth - 1, places=14)
            self.assertAlmostEqual(cr.geometric_mean_return,
                                   growth ** (1 / len(window)) - 1, places=14)

    def test_revert_empty_raises(self):
        cr = CumulativeReturn()
        with self.assertRaises(ValueError):
            cr.revert(0.01)

    def test_reset(self):
        cr = CumulativeReturn()
        for r in (0.1, -0.2):
            cr.update(r)
        cr.reset()
        self.assertEqual(cr.count, 0)
        self.assertEqual(cr.cumulative_geometric_return, 0.0)
        self.assertTrue(math.isnan(cr.geometric_mean_return))
