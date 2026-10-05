import collections
import math
import unittest

from ...streaming_kbn import RawMomentsKleinKBN
from .es import es_cornish_fisher, es_gaussian, es_historical
from .var import var_cornish_fisher, var_gaussian, var_historical


class TestRiskHelpers(unittest.TestCase):
    def test_historical_quantile_tail_and_risk_free_rate(self):
        returns = collections.deque([-0.2, -0.1, 0.0, 0.1])
        self.assertAlmostEqual(var_historical(returns, confidence=0.75), 0.125, places=15)
        self.assertAlmostEqual(es_historical(returns, confidence=0.75), 0.2, places=15)
        self.assertAlmostEqual(var_historical(returns, risk_free_rate=0.01,
                                             confidence=0.75), 0.135, places=15)
        self.assertAlmostEqual(es_historical(returns, risk_free_rate=0.01,
                                            confidence=0.75), 0.21, places=15)
        self.assertEqual(list(returns), [-0.2, -0.1, 0.0, 0.1])

    def test_empty_historical_inputs(self):
        for returns in (None, collections.deque()):
            with self.subTest(returns=returns):
                self.assertTrue(math.isnan(var_historical(returns)))
                self.assertTrue(math.isnan(es_historical(returns)))

    def test_cornish_fisher_falls_back_for_one_sample(self):
        moments = RawMomentsKleinKBN()
        moments.update(0.02)
        self.assertAlmostEqual(var_gaussian(moments), -0.02, places=15)
        self.assertAlmostEqual(es_gaussian(moments), -0.02, places=15)
        self.assertEqual(var_cornish_fisher(moments), var_gaussian(moments))
        self.assertEqual(es_cornish_fisher(moments), es_gaussian(moments))


if __name__ == '__main__':
    unittest.main()
