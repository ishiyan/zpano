import math
import unittest

from .linear_regression_klein_kbn import LinearRegressionKleinKBN

# Bacon, Carl R., Practical Portfolio Performance Measurement and
# Attribution, 2nd ed., Wiley, 2008, p. 65 (portfolio) and p. 66 (benchmark).
PORTFOLIO = [
    0.003, 0.026, 0.011, -0.010,
    0.015, 0.025, 0.016, 0.067,
    -0.014, 0.040, -0.005, 0.081,
    0.040, -0.037, -0.061, 0.017,
    -0.049, -0.022, 0.070, 0.058,
    -0.065, 0.024, -0.005, -0.009]
BENCHMARK = [
    0.002, 0.025, 0.018, -0.011,
    0.014, 0.018, 0.014, 0.065,
    -0.015, 0.042, -0.006, 0.083,
    0.039, -0.038, -0.062, 0.015,
    -0.048, 0.021, 0.060, 0.056,
    -0.067, 0.019, -0.003, 0.000]

# Reference values for x = BENCHMARK, y = PORTFOLIO, computed with exact
# rational arithmetic (fractions.Fraction on the binary float inputs,
# square roots with 50-digit decimal.Decimal), rounded to the nearest float.
SLOPE = 0.9988502086225746
INTERCEPT = -0.001030120844918352
CORRELATION = 0.9693858148753051
CO_MOMENT = 0.033844
COVARIANCE = 0.0014101666666666668
VARIANCE_X = 0.0014117899305555557
VARIANCE_Y = 0.0014989166666666668


def feed(reg: LinearRegressionKleinKBN, xs, ys) -> LinearRegressionKleinKBN:
    for x, y in zip(xs, ys):
        reg.update(x, y)
    return reg


class TestLinearRegressionKleinKBN(unittest.TestCase):

    def assert_all_nan(self, reg):
        for name in ('slope', 'intercept', 'correlation'):
            with self.subTest(name=name):
                self.assertTrue(math.isnan(getattr(reg, name)))

    def test_bacon(self):
        reg = feed(LinearRegressionKleinKBN(), BENCHMARK, PORTFOLIO)
        self.assertEqual(reg.n, len(PORTFOLIO))
        self.assertAlmostEqual(reg.slope, SLOPE, places=14)
        self.assertAlmostEqual(reg.intercept, INTERCEPT, places=15)
        self.assertAlmostEqual(reg.correlation, CORRELATION, places=14)
        self.assertAlmostEqual(reg.co_moment, CO_MOMENT, places=16)
        self.assertAlmostEqual(reg.covariance, COVARIANCE, places=16)
        self.assertAlmostEqual(reg.variance_x, VARIANCE_X, places=16)
        self.assertAlmostEqual(reg.variance_y, VARIANCE_Y, places=16)

    def test_perfect_fit(self):
        reg = LinearRegressionKleinKBN()
        for x in range(5):
            reg.update(x, 2 * x + 1)
        self.assertAlmostEqual(reg.slope, 2.0, places=13)
        self.assertAlmostEqual(reg.intercept, 1.0, places=13)
        self.assertAlmostEqual(reg.correlation, 1.0, places=15)
        self.assertAlmostEqual(reg.mean_x, 2.0, places=15)
        self.assertAlmostEqual(reg.mean_y, 5.0, places=15)
        self.assertAlmostEqual(reg.variance_x, 2.0, places=15)
        self.assertAlmostEqual(reg.co_moment, 20.0, places=13)
        self.assertAlmostEqual(reg.covariance, 4.0, places=13)

    def test_negative_correlation(self):
        reg = LinearRegressionKleinKBN()
        for x in range(5):
            reg.update(x, -2.0 * x + 1.0)
        self.assertAlmostEqual(reg.slope, -2.0, places=13)
        self.assertAlmostEqual(reg.intercept, 1.0, places=13)
        self.assertAlmostEqual(reg.correlation, -1.0, places=15)
        self.assertAlmostEqual(reg.covariance, -4.0, places=13)

    def test_constant_y(self):
        reg = LinearRegressionKleinKBN()
        for x in range(5):
            reg.update(x, 3.0)
        self.assertAlmostEqual(reg.slope, 0.0, places=13)
        self.assertAlmostEqual(reg.intercept, 3.0, places=13)
        self.assertEqual(reg.covariance, 0.0)
        self.assertTrue(math.isnan(reg.correlation))

    def test_constant_x(self):
        reg = LinearRegressionKleinKBN()
        for y in range(5):
            reg.update(3.0, y)
        self.assertEqual(reg.covariance, 0.0)
        self.assert_all_nan(reg)

    def test_empty(self):
        reg = LinearRegressionKleinKBN()
        self.assertEqual(reg.n, 0)
        self.assertEqual(reg.co_moment, 0.0)
        self.assertTrue(math.isnan(reg.covariance))
        self.assertTrue(math.isnan(reg.variance_x))
        self.assert_all_nan(reg)

    def test_single_point(self):
        reg = LinearRegressionKleinKBN()
        reg.update(1.0, 2.0)
        self.assertEqual(reg.covariance, 0.0)
        self.assert_all_nan(reg)

    def test_two_points(self):
        reg = LinearRegressionKleinKBN()
        reg.update(0.0, 1.0)
        reg.update(2.0, 5.0)
        self.assertAlmostEqual(reg.slope, 2.0, places=13)
        self.assertAlmostEqual(reg.intercept, 1.0, places=13)
        self.assertAlmostEqual(reg.correlation, 1.0, places=13)

    def test_revert_most_recent(self):
        reg = feed(LinearRegressionKleinKBN(), BENCHMARK + [0.5], PORTFOLIO + [-0.5])
        reg.revert(0.5, -0.5)
        self.assertEqual(reg.n, len(PORTFOLIO))
        self.assertAlmostEqual(reg.slope, SLOPE, places=13)
        self.assertAlmostEqual(reg.intercept, INTERCEPT, places=14)
        self.assertAlmostEqual(reg.correlation, CORRELATION, places=13)
        self.assertAlmostEqual(reg.covariance, COVARIANCE, places=15)

    def test_revert_oldest(self):
        reg = feed(LinearRegressionKleinKBN(), [0.5] + BENCHMARK, [-0.5] + PORTFOLIO)
        reg.revert(0.5, -0.5)
        self.assertAlmostEqual(reg.slope, SLOPE, places=13)
        self.assertAlmostEqual(reg.intercept, INTERCEPT, places=14)
        self.assertAlmostEqual(reg.correlation, CORRELATION, places=13)
        self.assertAlmostEqual(reg.covariance, COVARIANCE, places=15)

    def test_revert_to_single(self):
        reg = LinearRegressionKleinKBN()
        reg.update(1.0, 2.0)
        reg.update(3.0, 4.0)
        reg.revert(3.0, 4.0)
        self.assertEqual(reg.n, 1)
        self.assertAlmostEqual(reg.mean_x, 1.0, places=15)
        self.assertAlmostEqual(reg.mean_y, 2.0, places=15)
        self.assertAlmostEqual(reg.co_moment, 0.0, places=15)
        self.assert_all_nan(reg)

    def test_revert_to_empty(self):
        reg = LinearRegressionKleinKBN()
        reg.update(1.0, 2.0)
        reg.revert(1.0, 2.0)
        self.assertEqual(reg.n, 0)
        self.assert_all_nan(reg)

    def test_revert_empty_raises(self):
        reg = LinearRegressionKleinKBN()
        with self.assertRaises(ValueError):
            reg.revert(1.0, 2.0)

    def test_rolling_window(self):
        names = ('slope', 'intercept', 'correlation', 'covariance', 'co_moment',
                 'variance_x', 'variance_y')
        w = 6
        reg = LinearRegressionKleinKBN()
        for i, (x, y) in enumerate(zip(BENCHMARK, PORTFOLIO)):
            reg.update(x, y)
            if i >= w:
                reg.revert(BENCHMARK[i - w], PORTFOLIO[i - w])
            lo = max(0, i - w + 1)
            ref = feed(LinearRegressionKleinKBN(), BENCHMARK[lo:i + 1], PORTFOLIO[lo:i + 1])
            self.assertEqual(reg.n, ref.n)
            for name in names:
                with self.subTest(step=i, name=name):
                    actual, expected = getattr(reg, name), getattr(ref, name)
                    if math.isnan(expected):
                        self.assertTrue(math.isnan(actual))
                    else:
                        self.assertAlmostEqual(actual, expected, places=13)

    def test_reset(self):
        reg = LinearRegressionKleinKBN()
        for x in range(5):
            reg.update(x, 2 * x + 1)
        reg.reset()
        self.assertEqual(reg.n, 0)
        self.assert_all_nan(reg)
        reg.update(0.0, 1.0)
        reg.update(1.0, 3.0)
        self.assertAlmostEqual(reg.slope, 2.0, places=13)


if __name__ == '__main__':
    unittest.main()
