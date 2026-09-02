import unittest
import math

from .cumulative_return import CumulativeReturn


class TestCumulativeReturn(unittest.TestCase):

    def test_annualized_return_definition(self):
        returns = [0.10, -0.05, 0.03, 0.08]
        periods_per_year = 12

        cr = CumulativeReturn(window_size=0)
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
        cr = CumulativeReturn(0)
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

        cr = CumulativeReturn(0)
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
        cr = CumulativeReturn(0)

        for _ in range(60):
            cr.update(0.01)

        expected = 1.01**12 - 1
        actual = cr.annualized_geometric_mean_return(12)
        self.assertAlmostEqual(actual, expected, places=15)

    def test_empty(self):
        """
        Empty accumulator
        """
        cr = CumulativeReturn(0)
        actual = math.isnan(cr.annualized_geometric_mean_return(12))
        self.assertTrue(actual)

    def test_zero_returns(self):
        """
        If every return is zero, the annualized geometric mean return must be
        exactly zero because `log1p(0) == 0` and the sum remains exactly zero `expm1(0) == 0`.
        """
        cr = CumulativeReturn(0)

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

        cr = CumulativeReturn(window_size=0)
        for r in returns:
            cr.update(r)

        expected = expected = (1 + cr.geometric_mean_return) ** periods_per_year - 1
        actual = cr.annualized_geometric_mean_return(periods_per_year)
        self.assertAlmostEqual(actual, expected, places=13)
