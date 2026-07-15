import math
import unittest

from .linear_regression_klein_kbn import LinearRegressionKleinKBN


class TestLinearRegressionKleinKBN(unittest.TestCase):
    def test_perfect_fit(self):
        reg = LinearRegressionKleinKBN()
        for x in range(5):
            reg.update(x, 2 * x + 1)
        self.assertAlmostEqual(reg.slope, 2.0, places=13)
        self.assertAlmostEqual(reg.intercept, 1.0, places=13)
        self.assertAlmostEqual(reg.correlation, 1.0, places=13)

    def test_zero_correlation(self):
        reg = LinearRegressionKleinKBN()
        for x in range(5):
            reg.update(x, 0.0)
        self.assertAlmostEqual(reg.slope, 0.0, places=13)
        self.assertTrue(math.isnan(reg.correlation))

    def test_single_point(self):
        reg = LinearRegressionKleinKBN()
        reg.update(1.0, 2.0)
        self.assertTrue(math.isnan(reg.slope))
        self.assertTrue(math.isnan(reg.intercept))
        self.assertTrue(math.isnan(reg.correlation))

    def test_two_points(self):
        reg = LinearRegressionKleinKBN()
        reg.update(0.0, 1.0)
        reg.update(2.0, 5.0)
        self.assertAlmostEqual(reg.slope, 2.0, places=13)
        self.assertAlmostEqual(reg.intercept, 1.0, places=13)
        self.assertAlmostEqual(reg.correlation, 1.0, places=13)

    def test_revert_matches_single_update(self):
        reg = LinearRegressionKleinKBN()
        reg.update(1.0, 2.0)
        reg.update(3.0, 4.0)
        reg.revert(3.0, 4.0)

        ref = LinearRegressionKleinKBN()
        ref.update(1.0, 2.0)

        self.assertEqual(reg.n, ref.n)
        self.assertTrue(math.isnan(reg.slope))
        self.assertTrue(math.isnan(ref.slope))

    def test_revert_to_empty(self):
        reg = LinearRegressionKleinKBN()
        reg.update(1.0, 2.0)
        reg.revert(1.0, 2.0)
        self.assertEqual(reg.n, 0)
        self.assertTrue(math.isnan(reg.slope))
        self.assertTrue(math.isnan(reg.intercept))
        self.assertTrue(math.isnan(reg.correlation))

    def test_rolling_window(self):
        data = [(0.0, 1.0), (1.0, 3.0), (2.0, 5.0), (3.0, 7.0), (4.0, 9.0)]
        reg = LinearRegressionKleinKBN()
        for x, y in data:
            reg.update(x, y)

        reg.revert(*data[0])
        reg.revert(*data[1])
        reg.update(5.0, 11.0)
        reg.update(6.0, 13.0)

        ref = LinearRegressionKleinKBN()
        for x, y in data[2:]:
            ref.update(x, y)
        ref.update(5.0, 11.0)
        ref.update(6.0, 13.0)

        self.assertEqual(reg.n, ref.n)
        self.assertAlmostEqual(reg.slope, ref.slope, places=12)
        self.assertAlmostEqual(reg.intercept, ref.intercept, places=12)
        self.assertAlmostEqual(reg.correlation, ref.correlation, places=12)

    def test_negative_correlation(self):
        reg = LinearRegressionKleinKBN()
        for x in range(5):
            reg.update(x, -2.0 * x + 1.0)
        self.assertAlmostEqual(reg.slope, -2.0, places=13)
        self.assertAlmostEqual(reg.intercept, 1.0, places=13)
        self.assertAlmostEqual(reg.correlation, -1.0, places=13)

    def test_reset(self):
        reg = LinearRegressionKleinKBN()
        for x in range(5):
            reg.update(x, 2 * x + 1)
        reg.reset()
        self.assertEqual(reg.n, 0)
        self.assertTrue(math.isnan(reg.slope))
        self.assertTrue(math.isnan(reg.intercept))
        self.assertTrue(math.isnan(reg.correlation))
        reg.update(0.0, 1.0)
        reg.update(1.0, 3.0)
        self.assertAlmostEqual(reg.slope, 2.0, places=13)
