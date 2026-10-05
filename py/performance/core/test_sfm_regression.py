import math
import unittest

from .sfm_regression import SFMRegression


class TestSFMRegression(unittest.TestCase):
    def test_full_bull_and_bear_fits_use_excess_returns(self):
        regression = SFMRegression(risk_free_rate=0.01)
        for excess_benchmark in (-0.04, -0.02, 0.0, 0.02, 0.05):
            benchmark = 0.01 + excess_benchmark
            portfolio = 0.01 + 0.005 + 2 * excess_benchmark
            regression.update(portfolio, benchmark)

        self.assertAlmostEqual(regression.alpha, 0.005, places=14)
        self.assertAlmostEqual(regression.beta, 2.0, places=14)
        self.assertAlmostEqual(regression.beta_bull, 2.0, places=14)
        self.assertAlmostEqual(regression.beta_bear, 2.0, places=14)
        self.assertAlmostEqual(regression.r2, 1.0, places=14)

        # Removing an older bear observation leaves too few bear points
        # for a slope, while the full and bull fits remain defined.
        regression.revert(0.01 + 0.005 + 2 * -0.04, 0.01 - 0.04)
        self.assertAlmostEqual(regression.beta, 2.0, places=14)
        self.assertAlmostEqual(regression.beta_bull, 2.0, places=14)
        self.assertTrue(math.isnan(regression.beta_bear))

    def test_reset_and_zero_excess_benchmark(self):
        regression = SFMRegression(risk_free_rate=0.01)
        regression.update(0.02, 0.01)
        self.assertTrue(math.isnan(regression.beta_bull))
        self.assertTrue(math.isnan(regression.beta_bear))
        regression.reset()
        for name in ('alpha', 'beta', 'beta_bull', 'beta_bear', 'r2'):
            with self.subTest(name=name):
                self.assertTrue(math.isnan(getattr(regression, name)))


if __name__ == '__main__':
    unittest.main()
