import math
import unittest

from ...streaming_kbn import RawMomentsKleinKBN
from .probabilistic_sharpe_ratio import probabilistic_sharpe_ratio


class TestProbabilisticSharpeRatio(unittest.TestCase):
    def test_four_moment_assumptions(self):
        values = [-0.03, -0.01, 0.02, 0.04, 0.07]
        moments = RawMomentsKleinKBN()
        for value in values:
            moments.update(value)
        mean = math.fsum(values) / len(values)
        mu2 = math.fsum((x - mean) ** 2 for x in values) / len(values)
        skew = math.fsum((x - mean) ** 3 for x in values) / len(values) / mu2 ** 1.5
        kurt = math.fsum((x - mean) ** 4 for x in values) / len(values) / mu2 ** 2
        sr = 0.75
        reference_sr = 0.1

        for zero_skewness in (False, True):
            for normal_kurtosis in (False, True):
                with self.subTest(zero_skewness=zero_skewness,
                                  normal_kurtosis=normal_kurtosis):
                    s = 0.0 if zero_skewness else skew
                    k = 3.0 if normal_kurtosis else kurt
                    denominator = math.sqrt(1 - sr * s + sr * sr * (k - 1) / 4)
                    z = (sr - reference_sr) * math.sqrt(len(values) - 1) / denominator
                    expected = 0.5 * (1 + math.erf(z / math.sqrt(2)))
                    actual = probabilistic_sharpe_ratio(
                        moments, sr, reference_sr, zero_skewness, normal_kurtosis)
                    self.assertAlmostEqual(actual, expected, places=13)

    def test_unavailable_sharpe_or_moments(self):
        moments = RawMomentsKleinKBN()
        moments.update(0.01)
        self.assertTrue(math.isnan(probabilistic_sharpe_ratio(moments, math.nan)))
        self.assertTrue(math.isnan(probabilistic_sharpe_ratio(moments, 0.5)))
        self.assertTrue(math.isnan(probabilistic_sharpe_ratio(
            moments, 0.5, zero_skewness=True, normal_kurtosis=False)))


if __name__ == '__main__':
    unittest.main()
