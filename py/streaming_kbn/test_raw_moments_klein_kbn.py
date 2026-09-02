import math
import unittest
import numpy as np
from scipy import stats

from .raw_moments_klein_kbn import RawMomentsKleinKBN

# https://github.com/medo64/Medo/blob/main/tests/Tests.Medo/Math/WelfordVariance.cs
# https://github.com/andrewuhl/RollingWindow/blob/master/src/RollingWindow.cpp
# https://github.com/ajcr/rolling/blob/master/rolling/similarity.py

class TestRawMomentsKleinKBN(unittest.TestCase):

    def test_simple_update(self):
        m = RawMomentsKleinKBN(ddof=0)
        for x in [1.0, 2.0, 3.0, 4.0]:
            m.update(x)
        self.assertAlmostEqual(m.mean, 2.5, places=15)
        self.assertAlmostEqual(m.variance, 1.25, places=15)
        self.assertAlmostEqual(m.skewness, 0.0, places=14)
        self.assertAlmostEqual(m.kurtosis, -1.36, places=13)

    def test_compare_scipy(self):
        data = [0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067,
                -0.014, 0.040, -0.005, 0.081, 0.040, -0.037, -0.061, 0.017,
                -0.049, -0.022, 0.070, 0.058, -0.065, 0.024, -0.005, -0.009]
        m = RawMomentsKleinKBN(ddof=0, bias=True, fisher=True)
        for x in data:
            m.update(x)
        sp_m = stats.tmean(data)
        sp_v = stats.tvar(data, ddof=0)
        sp_s = stats.skew(data, bias=True)
        sp_k = stats.kurtosis(data, bias=True, fisher=True)

        self.assertAlmostEqual(m.mean, sp_m, places=15)
        self.assertAlmostEqual(m.variance, sp_v, places=14)
        self.assertAlmostEqual(m.skewness, sp_s, places=14)
        self.assertAlmostEqual(m.kurtosis, sp_k, places=13)

    def test_compare_scipy_bias_false(self):
        data = [0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067,
                -0.014, 0.040, -0.005, 0.081, 0.040, -0.037, -0.061, 0.017,
                -0.049, -0.022, 0.070, 0.058, -0.065, 0.024, -0.005, -0.009]
        m = RawMomentsKleinKBN(ddof=0, bias=False, fisher=True)
        for x in data:
            m.update(x)
        sp_s = stats.skew(data, bias=False)
        sp_k = stats.kurtosis(data, bias=False, fisher=True)

        self.assertAlmostEqual(m.skewness, sp_s, places=14)
        self.assertAlmostEqual(m.kurtosis, sp_k, places=13)

    def test_ddof(self):
        m = RawMomentsKleinKBN(ddof=1)
        for x in [1.0, 2.0, 3.0]:
            m.update(x)
        self.assertAlmostEqual(m.variance, 1.0, places=15)

    def test_standard_deviation(self):
        m = RawMomentsKleinKBN(ddof=0)
        for x in [1.0, 2.0, 3.0, 4.0]:
            m.update(x)
        self.assertAlmostEqual(m.standard_deviation, m.variance**0.5, places=15)

    def test_revert_roundtrip(self):
        data = [1.0, 2.0, 3.0, 4.0, 5.0]
        m = RawMomentsKleinKBN(ddof=0)
        for x in data:
            m.update(x)
        for x in reversed(data[1:]):
            m.revert(x)
        self.assertEqual(m.n, 1)
        self.assertAlmostEqual(m.mean, 1.0, places=15)

    def test_revert_partial(self):
        data = [10.0, 18.0, 5.0, 12.0, 7.0]
        m_full = RawMomentsKleinKBN(ddof=0)
        m_part = RawMomentsKleinKBN(ddof=0)
        for x in data:
            m_full.update(x)
        for x in data[:4]:
            m_part.update(x)
        m_full.revert(data[4])
        self.assertAlmostEqual(m_full.mean, m_part.mean, places=15)
        self.assertAlmostEqual(m_full.variance, m_part.variance, places=15)
        self.assertAlmostEqual(m_full.skewness, m_part.skewness, places=14)
        self.assertAlmostEqual(m_full.kurtosis, m_part.kurtosis, places=13)

    def test_reset(self):
        m = RawMomentsKleinKBN()
        m.update(10.0)
        m.reset()
        self.assertEqual(m.n, 0)
        self.assertTrue(math.isnan(m.variance))
