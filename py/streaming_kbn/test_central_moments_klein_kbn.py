import math
import statistics
import unittest

from .central_moments_klein_kbn import CentralMomentsKleinKBN

# Bacon, Carl R., Practical Portfolio Performance Measurement and
# Attribution, 2nd ed., Wiley, 2008, p. 65 (portfolio returns).
BACON = [
    0.003, 0.026, 0.011, -0.010,
    0.015, 0.025, 0.016, 0.067,
    -0.014, 0.040, -0.005, 0.081,
    0.040, -0.037, -0.061, 0.017,
    -0.049, -0.022, 0.070, 0.058,
    -0.065, 0.024, -0.005, -0.009]

# Reference values for BACON, computed with exact rational arithmetic
# (fractions.Fraction on the binary float inputs, square roots with
# 50-digit decimal.Decimal), rounded to the nearest float.  The names
# follow scipy.stats: skew(bias=...), kurtosis(bias=..., fisher=...).
MEAN = 0.009000000000000001
VARIANCE_DDOF_0 = 0.0014989166666666668
VARIANCE_DDOF_1 = 0.0015640869565217393
STD_DDOF_0 = 0.03871584516275819
STD_DDOF_1 = 0.039548539246370897
SKEW_BIASED = -0.08256245520856804            # skew(bias=True)
SKEW_UNBIASED = -0.08817174934967535          # skew(bias=False)
KURT_BIASED_FISHER = -0.5675462058921257      # kurtosis(bias=True, fisher=True)
KURT_BIASED_PEARSON = 2.4324537941078743      # kurtosis(bias=True, fisher=False)
KURT_UNBIASED_FISHER = -0.40766032118608714   # kurtosis(bias=False, fisher=True)
KURT_UNBIASED_PEARSON = 2.592339678813913     # kurtosis(bias=False, fisher=False)

# Same statistics for [1e4 + x for x in BACON], exact for the shifted floats.
OFFSET_SKEW_BIASED = -0.08256245521966786
OFFSET_KURT_BIASED_FISHER = -0.5675462058934164


def feed(m: CentralMomentsKleinKBN, data) -> CentralMomentsKleinKBN:
    for x in data:
        m.update(x)
    return m


class TestCentralMomentsKleinKBN(unittest.TestCase):

    def test_simple_update(self):
        m = feed(CentralMomentsKleinKBN(ddof=0), [1.0, 2.0, 3.0, 4.0])
        self.assertEqual(m.n, 4)
        self.assertAlmostEqual(m.mean, 2.5, places=15)
        self.assertAlmostEqual(m.variance, 1.25, places=15)
        self.assertAlmostEqual(m.skewness, 0.0, places=14)
        self.assertAlmostEqual(m.kurtosis, -1.36, places=13)

    def test_bacon_mean_variance(self):
        m0 = feed(CentralMomentsKleinKBN(ddof=0), BACON)
        m1 = feed(CentralMomentsKleinKBN(ddof=1), BACON)
        self.assertAlmostEqual(m0.mean, MEAN, places=16)
        self.assertAlmostEqual(m0.variance, VARIANCE_DDOF_0, places=16)
        self.assertAlmostEqual(m1.variance, VARIANCE_DDOF_1, places=16)
        self.assertAlmostEqual(m0.standard_deviation, STD_DDOF_0, places=15)
        self.assertAlmostEqual(m1.standard_deviation, STD_DDOF_1, places=15)

    def test_bacon_skewness_kurtosis(self):
        cases = [
            (True, True, SKEW_BIASED, KURT_BIASED_FISHER),
            (True, False, SKEW_BIASED, KURT_BIASED_PEARSON),
            (False, True, SKEW_UNBIASED, KURT_UNBIASED_FISHER),
            (False, False, SKEW_UNBIASED, KURT_UNBIASED_PEARSON),
        ]
        for bias, fisher, skew, kurt in cases:
            with self.subTest(bias=bias, fisher=fisher):
                m = feed(CentralMomentsKleinKBN(ddof=0, bias=bias, fisher=fisher), BACON)
                self.assertAlmostEqual(m.skewness, skew, places=14)
                self.assertAlmostEqual(m.kurtosis, kurt, places=13)

    def test_large_offset(self):
        # Central moments don't suffer from the cancellation of raw power sums.
        m = feed(CentralMomentsKleinKBN(ddof=0), [1e4 + x for x in BACON])
        self.assertAlmostEqual(m.mean, 1e4 + MEAN, places=11)
        self.assertAlmostEqual(m.variance, VARIANCE_DDOF_0, places=13)
        self.assertAlmostEqual(m.skewness, OFFSET_SKEW_BIASED, places=10)
        self.assertAlmostEqual(m.kurtosis, OFFSET_KURT_BIASED_FISHER, places=10)

    def test_scale_invariance(self):
        m = feed(CentralMomentsKleinKBN(ddof=0), [x * 1e-6 for x in BACON])
        self.assertAlmostEqual(m.skewness, SKEW_BIASED, places=14)
        self.assertAlmostEqual(m.kurtosis, KURT_BIASED_FISHER, places=13)

    def test_empty(self):
        m = CentralMomentsKleinKBN()
        self.assertEqual(m.n, 0)
        self.assertEqual(m.mean, 0.0)
        self.assertTrue(math.isnan(m.variance))
        self.assertTrue(math.isnan(m.standard_deviation))
        self.assertTrue(math.isnan(m.skewness))
        self.assertTrue(math.isnan(m.kurtosis))

    def test_ddof(self):
        m = feed(CentralMomentsKleinKBN(ddof=1), [1.0, 2.0, 3.0])
        self.assertAlmostEqual(m.variance, 1.0, places=15)
        self.assertAlmostEqual(m.standard_deviation, 1.0, places=15)
        m = feed(CentralMomentsKleinKBN(ddof=1), [1.0])
        self.assertTrue(math.isnan(m.variance))
        self.assertTrue(math.isnan(m.standard_deviation))

    def test_invalid_ddof(self):
        for ddof in (-1, 0.5, True):
            with self.subTest(ddof=ddof), self.assertRaises(ValueError):
                CentralMomentsKleinKBN(ddof=ddof)

    def test_minimum_sample_sizes(self):
        data = [1.0, 2.0, 4.0, 8.0]
        # (bias, fisher) -> minimum n for (skewness, kurtosis)
        cases = {
            (True, True): (2, 2),
            (True, False): (2, 2),
            (False, True): (3, 4),
            (False, False): (3, 4),
        }
        for (bias, fisher), (skew_n, kurt_n) in cases.items():
            m = CentralMomentsKleinKBN(ddof=0, bias=bias, fisher=fisher)
            for i, x in enumerate(data):
                m.update(x)
                n = i + 1
                with self.subTest(bias=bias, fisher=fisher, n=n):
                    self.assertEqual(math.isnan(m.skewness), n < skew_n)
                    self.assertEqual(math.isnan(m.kurtosis), n < kurt_n)

    def test_constant_data(self):
        m = feed(CentralMomentsKleinKBN(ddof=0), [3.0] * 5)
        self.assertEqual(m.mean, 3.0)
        self.assertEqual(m.variance, 0.0)
        self.assertEqual(m.standard_deviation, 0.0)
        self.assertTrue(math.isnan(m.skewness))
        self.assertTrue(math.isnan(m.kurtosis))

    def test_revert_lifo_simple(self):
        data = [10.0, 18.0, 5.0]
        m_full = feed(CentralMomentsKleinKBN(ddof=0), data)
        m_part = feed(CentralMomentsKleinKBN(ddof=0), data[:2])
        m_full.revert(data[2])

        self.assertEqual(m_full.n, 2)
        self.assertAlmostEqual(m_full.mean, m_part.mean, places=15)
        self.assertAlmostEqual(m_full.variance, m_part.variance, places=15)
        self.assertAlmostEqual(m_full.skewness, m_part.skewness, places=14)
        self.assertAlmostEqual(m_full.kurtosis, m_part.kurtosis, places=13)

    def test_revert_lifo_bacon(self):
        for bias, fisher in [(True, True), (False, False)]:
            with self.subTest(bias=bias, fisher=fisher):
                m_full = feed(CentralMomentsKleinKBN(ddof=0, bias=bias, fisher=fisher), BACON)
                m_part = feed(CentralMomentsKleinKBN(ddof=0, bias=bias, fisher=fisher), BACON[:-1])
                m_full.revert(BACON[-1])

                self.assertAlmostEqual(m_full.mean, m_part.mean, places=15)
                self.assertAlmostEqual(m_full.variance, m_part.variance, places=15)
                self.assertAlmostEqual(m_full.skewness, m_part.skewness, places=13)
                self.assertAlmostEqual(m_full.kurtosis, m_part.kurtosis, places=12)

    def test_revert_then_update(self):
        m = feed(CentralMomentsKleinKBN(ddof=0), BACON)
        for x in reversed(BACON[12:]):
            m.revert(x)
        feed(m, BACON[12:])
        self.assertAlmostEqual(m.mean, MEAN, places=15)
        self.assertAlmostEqual(m.variance, VARIANCE_DDOF_0, places=15)
        self.assertAlmostEqual(m.skewness, SKEW_BIASED, places=12)
        self.assertAlmostEqual(m.kurtosis, KURT_BIASED_FISHER, places=12)

    def test_revert_lifo_roundtrip(self):
        m = feed(CentralMomentsKleinKBN(ddof=0), BACON)
        for x in reversed(BACON):
            m.revert(x)
        self.assertEqual(m.n, 0)
        self.assertEqual(m.mean, 0.0)
        self.assertTrue(math.isnan(m.variance))

    def test_revert_oldest_and_middle(self):
        data = [0.0, 1.0, 2.0, 4.0, 8.0]
        m = feed(CentralMomentsKleinKBN(ddof=0), data)
        for removed in (0.0, 2.0):
            m.revert(removed)
            data.remove(removed)
            mean = statistics.fmean(data)
            mu2 = math.fsum((x - mean) ** 2 for x in data) / len(data)
            mu3 = math.fsum((x - mean) ** 3 for x in data) / len(data)
            mu4 = math.fsum((x - mean) ** 4 for x in data) / len(data)
            self.assertEqual(m.n, len(data))
            self.assertAlmostEqual(m.mean, mean, places=14)
            self.assertAlmostEqual(m.variance, mu2, places=14)
            self.assertAlmostEqual(m.skewness, mu3 / mu2 ** 1.5, places=13)
            self.assertAlmostEqual(m.kurtosis, mu4 / mu2 ** 2 - 3, places=13)

    def test_fifo_rolling_window(self):
        m = CentralMomentsKleinKBN(ddof=0)
        width = 6
        for i, x in enumerate(BACON):
            m.update(x)
            if i >= width:
                m.revert(BACON[i - width])
            window = BACON[max(0, i - width + 1):i + 1]
            self.assertEqual(m.n, len(window))
            self.assertAlmostEqual(m.mean, statistics.fmean(window), places=14)
            self.assertAlmostEqual(m.variance, statistics.pvariance(window), places=14)

    def test_revert_empty_raises(self):
        m = CentralMomentsKleinKBN()
        with self.assertRaises(ValueError):
            m.revert(1.0)

    def test_standard_deviation_is_real_after_revert(self):
        # Reverting to two equal samples can leave a tiny negative M2.
        m = CentralMomentsKleinKBN(ddof=0)
        for x in [0.1, 0.1, 0.7]:
            m.update(x)
        m.revert(0.7)
        self.assertIsInstance(m.standard_deviation, float)
        self.assertGreaterEqual(m.variance, 0.0)
        self.assertAlmostEqual(m.standard_deviation, 0.0, places=15)

    def test_reset(self):
        m = feed(CentralMomentsKleinKBN(), BACON)
        m.reset()
        self.assertEqual(m.n, 0)
        self.assertEqual(m.mean, 0.0)
        self.assertTrue(math.isnan(m.variance))
        feed(m, [1.0, 2.0, 3.0])
        self.assertAlmostEqual(m.variance, 1.0, places=15)


if __name__ == '__main__':
    unittest.main()
