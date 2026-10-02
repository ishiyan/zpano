import math
import unittest

from .raw_moments_klein_kbn import RawMomentsKleinKBN

# https://github.com/medo64/Medo/blob/main/tests/Tests.Medo/Math/WelfordVariance.cs
# https://github.com/andrewuhl/RollingWindow/blob/master/src/RollingWindow.cpp
# https://github.com/ajcr/rolling/blob/master/rolling/similarity.py

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
# 50-digit decimal.Decimal), rounded to the nearest float.
EXPECTED = {
    'mean': 0.009000000000000001,
    'variance_ddof_0': 0.0014989166666666668,
    'variance_ddof_1': 0.0015640869565217393,
    'standard_deviation_ddof_0': 0.03871584516275819,
    'standard_deviation_ddof_1': 0.039548539246370897,
    'skewness_moment': -0.08256245520856804,          # scipy skew(bias=True)
    'skewness_fisher': -0.08817174934967535,          # scipy skew(bias=False)
    'skewness_sample': -0.09398413873544505,          # R PerformanceAnalytics "sample"
    'kurtosis_moment': 2.4324537941078743,            # scipy kurtosis(bias=True, fisher=False)
    'kurtosis_excess': -0.5675462058921257,           # scipy kurtosis(bias=True, fisher=True)
    'kurtosis_sample_excess': -0.40766032118608714,   # scipy kurtosis(bias=False, fisher=True)
    'kurtosis_sample': 2.592339678813913,             # scipy kurtosis(bias=False, fisher=False)
    'kurtosis_sample_corrected': 3.027404613878848,   # R PerformanceAnalytics "sample"
    'x1_sum': 0.21600000000000003,
    'x2_sum': 0.037918,
    'x3_sum': 0.0008738040000000003,
    'x4_sum': 0.00014466403,
    'x1': 0.009000000000000001,
    'x2': 0.0015799166666666668,
    'x3': 3.640850000000001e-05,
    'x4': 6.0276679166666674e-06,
}


def feed(m: RawMomentsKleinKBN, data) -> RawMomentsKleinKBN:
    for x in data:
        m.update(x)
    return m


class TestRawMomentsKleinKBN(unittest.TestCase):

    def test_simple_update(self):
        m = feed(RawMomentsKleinKBN(ddof=0), [1.0, 2.0, 3.0, 4.0])
        self.assertEqual(m.n, 4)
        self.assertAlmostEqual(m.mean, 2.5, places=15)
        self.assertAlmostEqual(m.variance, 1.25, places=15)
        self.assertAlmostEqual(m.skewness, 0.0, places=14)
        self.assertAlmostEqual(m.kurtosis, -1.36, places=13)

    def test_bacon_all_properties(self):
        m = feed(RawMomentsKleinKBN(), BACON)
        for name, expected in EXPECTED.items():
            with self.subTest(name=name):
                self.assertAlmostEqual(getattr(m, name), expected, places=14)

    def test_dispatch(self):
        cases = [
            (True, True, 'skewness_moment', 'kurtosis_excess'),
            (True, False, 'skewness_moment', 'kurtosis_moment'),
            (False, True, 'skewness_fisher', 'kurtosis_sample_excess'),
            (False, False, 'skewness_fisher', 'kurtosis_sample'),
        ]
        for bias, fisher, skew, kurt in cases:
            with self.subTest(bias=bias, fisher=fisher):
                m = feed(RawMomentsKleinKBN(bias=bias, fisher=fisher), BACON)
                self.assertEqual(m.skewness, getattr(m, skew))
                self.assertEqual(m.kurtosis, getattr(m, kurt))
                self.assertAlmostEqual(m.skewness, EXPECTED[skew], places=14)
                self.assertAlmostEqual(m.kurtosis, EXPECTED[kurt], places=13)

    def test_ddof(self):
        for ddof in (0, 1):
            m = feed(RawMomentsKleinKBN(ddof=ddof), BACON)
            self.assertEqual(m.variance, getattr(m, f'variance_ddof_{ddof}'))
            self.assertEqual(m.standard_deviation, getattr(m, f'standard_deviation_ddof_{ddof}'))
        m = feed(RawMomentsKleinKBN(ddof=1), [1.0, 2.0, 3.0])
        self.assertAlmostEqual(m.variance, 1.0, places=15)
        self.assertAlmostEqual(m.standard_deviation, 1.0, places=15)

    def test_kurtosis_sample_corrected_difference(self):
        # kurtosis_sample_corrected - kurtosis_sample = (9n-15) / ((n-2)(n-3))
        m = feed(RawMomentsKleinKBN(), BACON)
        n = len(BACON)
        self.assertAlmostEqual(m.kurtosis_sample_corrected - m.kurtosis_sample,
                               (9 * n - 15) / ((n - 2) * (n - 3)), places=14)

    def test_empty(self):
        m = RawMomentsKleinKBN()
        self.assertEqual(m.n, 0)
        self.assertEqual(m.mean, 0.0)
        for name in ('variance', 'standard_deviation', 'skewness', 'kurtosis',
                     'x1', 'x2', 'x3', 'x4'):
            with self.subTest(name=name):
                self.assertTrue(math.isnan(getattr(m, name)))
        self.assertEqual(m.x1_sum, 0.0)

    def test_minimum_sample_sizes(self):
        data = [1.0, 2.0, 4.0, 8.0]
        minimum_n = {
            'skewness_moment': 2,
            'skewness_fisher': 3,
            'skewness_sample': 3,
            'kurtosis_moment': 2,
            'kurtosis_excess': 2,
            'kurtosis_sample_excess': 4,
            'kurtosis_sample': 4,
            'kurtosis_sample_corrected': 4,
        }
        m = RawMomentsKleinKBN()
        for i, x in enumerate(data):
            m.update(x)
            n = i + 1
            for name, min_n in minimum_n.items():
                with self.subTest(name=name, n=n):
                    self.assertEqual(math.isnan(getattr(m, name)), n < min_n)

    def test_constant_data(self):
        m = feed(RawMomentsKleinKBN(ddof=0), [0.1] * 5)
        self.assertAlmostEqual(m.mean, 0.1, places=16)
        self.assertAlmostEqual(m.variance, 0.0, places=16)
        self.assertTrue(math.isnan(m.skewness))
        self.assertTrue(math.isnan(m.kurtosis))

    def test_scale_invariance(self):
        # The cancellation threshold is relative, so tiny values work.
        m = feed(RawMomentsKleinKBN(), [x * 1e-6 for x in BACON])
        self.assertAlmostEqual(m.skewness_moment, EXPECTED['skewness_moment'], places=13)
        self.assertAlmostEqual(m.kurtosis_excess, EXPECTED['kurtosis_excess'], places=13)

    def test_revert_partial(self):
        data = [10.0, 18.0, 5.0, 12.0, 7.0]
        m_full = feed(RawMomentsKleinKBN(ddof=0), data)
        m_part = feed(RawMomentsKleinKBN(ddof=0), data[:4])
        m_full.revert(data[4])
        self.assertEqual(m_full.n, 4)
        self.assertAlmostEqual(m_full.mean, m_part.mean, places=15)
        self.assertAlmostEqual(m_full.variance, m_part.variance, places=15)
        self.assertAlmostEqual(m_full.skewness, m_part.skewness, places=14)
        self.assertAlmostEqual(m_full.kurtosis, m_part.kurtosis, places=13)

    def test_revert_not_most_recent(self):
        m = feed(RawMomentsKleinKBN(), BACON + [0.5])
        m.revert(0.5)
        m2 = feed(RawMomentsKleinKBN(), [0.5] + BACON)
        m2.revert(0.5)  # the oldest sample
        for name in ('mean', 'variance', 'skewness', 'kurtosis'):
            with self.subTest(name=name):
                self.assertAlmostEqual(getattr(m, name), EXPECTED_BY_DISPATCH[name], places=13)
                self.assertAlmostEqual(getattr(m2, name), EXPECTED_BY_DISPATCH[name], places=13)

    def test_revert_to_empty(self):
        m = feed(RawMomentsKleinKBN(ddof=0), BACON)
        for x in BACON:
            m.revert(x)
        self.assertEqual(m.n, 0)
        self.assertEqual(m.mean, 0.0)
        self.assertEqual(m.x1_sum, 0.0)
        self.assertTrue(math.isnan(m.variance))
        feed(m, [1.0, 2.0, 3.0, 4.0])
        self.assertAlmostEqual(m.variance, 1.25, places=15)

    def test_revert_empty_raises(self):
        m = RawMomentsKleinKBN()
        with self.assertRaises(ValueError):
            m.revert(1.0)

    def test_rolling_window(self):
        names = ('mean', 'variance', 'standard_deviation', 'skewness', 'kurtosis',
                 'skewness_sample', 'kurtosis_sample_corrected', 'x1', 'x2', 'x3', 'x4')
        w = 5
        m = RawMomentsKleinKBN(ddof=1, bias=False, fisher=True)
        for i, x in enumerate(BACON):
            m.update(x)
            if i >= w:
                m.revert(BACON[i - w])
            ref = feed(RawMomentsKleinKBN(ddof=1, bias=False, fisher=True),
                       BACON[max(0, i - w + 1):i + 1])
            self.assertEqual(m.n, ref.n)
            for name in names:
                with self.subTest(step=i, name=name):
                    actual, expected = getattr(m, name), getattr(ref, name)
                    if math.isnan(expected):
                        self.assertTrue(math.isnan(actual))
                    else:
                        self.assertAlmostEqual(actual, expected, places=13)

    def test_standard_deviation_is_real_after_revert(self):
        m = RawMomentsKleinKBN(ddof=0)
        for x in [0.1, 0.1, 0.7]:
            m.update(x)
        m.revert(0.7)
        self.assertIsInstance(m.standard_deviation, float)
        self.assertGreaterEqual(m.variance, 0.0)
        self.assertAlmostEqual(m.standard_deviation, 0.0, places=15)

    def test_variance_getter_has_no_side_effects(self):
        m = feed(RawMomentsKleinKBN(ddof=0), BACON)
        v = m.variance
        _ = m.standard_deviation
        self.assertEqual(m.variance, v)
        self.assertEqual(m.n, len(BACON))

    def test_reset(self):
        m = feed(RawMomentsKleinKBN(), BACON)
        m.reset()
        self.assertEqual(m.n, 0)
        self.assertEqual(m.mean, 0.0)
        self.assertEqual(m.x4_sum, 0.0)
        self.assertTrue(math.isnan(m.variance))
        feed(m, [1.0, 2.0, 3.0])
        self.assertAlmostEqual(m.variance, 1.0, places=15)


# Defaults are ddof=1, bias=True, fisher=True.
EXPECTED_BY_DISPATCH = {
    'mean': EXPECTED['mean'],
    'variance': EXPECTED['variance_ddof_1'],
    'skewness': EXPECTED['skewness_moment'],
    'kurtosis': EXPECTED['kurtosis_excess'],
}


if __name__ == '__main__':
    unittest.main()
