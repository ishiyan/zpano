import math
import unittest
import numpy as np
from scipy import stats

from .central_moments_klein_kbn import CentralMomentsKleinKBN


class TestCentralMomentsKleinKBN(unittest.TestCase):

    def setUp(self):
        self.bacon_data = [
            0.003, 0.026, 0.011, -0.010,
            0.015, 0.025, 0.016, 0.067,
            -0.014, 0.040, -0.005, 0.081,
            0.040, -0.037, -0.061, 0.017,
            -0.049, -0.022, 0.070, 0.058,
            -0.065, 0.024, -0.005, -0.009]

    def test_simple_update(self):
        m = CentralMomentsKleinKBN(ddof=0)
        for x in [1.0, 2.0, 3.0, 4.0]:
            m.update(x)
        self.assertAlmostEqual(m.mean, 2.5, places=15)
        self.assertAlmostEqual(m.variance, 1.25, places=15)
        self.assertAlmostEqual(m.skewness, 0.0, places=14)
        self.assertAlmostEqual(m.kurtosis, -1.36, places=13)

    def test_compare_scipy(self):
        m = CentralMomentsKleinKBN(ddof=0, bias=True, fisher=True)
        for x in self.bacon_data:
            m.update(x)
        sp_m = stats.tmean(self.bacon_data)
        sp_v = stats.tvar(self.bacon_data, ddof=0)
        sp_s = stats.skew(self.bacon_data, bias=True)
        sp_k = stats.kurtosis(self.bacon_data, bias=True, fisher=True)

        self.assertAlmostEqual(m.mean, sp_m, places=15)
        self.assertAlmostEqual(m.variance, sp_v, places=15)
        self.assertAlmostEqual(m.skewness, sp_s, places=14)
        self.assertAlmostEqual(m.kurtosis, sp_k, places=13)

    def test_compare_scipy_bias_false(self):
        m = CentralMomentsKleinKBN(ddof=0, bias=False, fisher=True)
        for x in self.bacon_data:
            m.update(x)
        sp_s = stats.skew(self.bacon_data, bias=False)
        sp_k = stats.kurtosis(self.bacon_data, bias=False, fisher=True)

        self.assertAlmostEqual(m.skewness, sp_s, places=14)
        self.assertAlmostEqual(m.kurtosis, sp_k, places=13)

    def test_ddof(self):
        m = CentralMomentsKleinKBN(ddof=1)
        for x in [1.0, 2.0, 3.0]:
            m.update(x)
        self.assertAlmostEqual(m.variance, 1.0, places=15)

    def test_revert_lifo_simple(self):
        data = [10.0, 18.0, 5.0]
        m_full = CentralMomentsKleinKBN(ddof=0)
        m_part = CentralMomentsKleinKBN(ddof=0)
        for x in data:
            m_full.update(x)
        for x in data[:2]:
            m_part.update(x)
        m_full.revert(data[2])

        self.assertAlmostEqual(m_full.mean, m_part.mean, places=15)
        self.assertAlmostEqual(m_full.variance, m_part.variance, places=15)
        self.assertAlmostEqual(m_full.skewness, m_part.skewness, places=14)
        self.assertAlmostEqual(m_full.kurtosis, m_part.kurtosis, places=13)

    def test_revert_lifo_bacon(self):
        m_full = CentralMomentsKleinKBN(ddof=0)
        m_part = CentralMomentsKleinKBN(ddof=0)
        for x in self.bacon_data:
            m_full.update(x)
        for x in self.bacon_data[:-1]:
            m_part.update(x)
        m_full.revert(self.bacon_data[-1])

        self.assertAlmostEqual(m_full.mean, m_part.mean, places=15)
        self.assertAlmostEqual(m_full.variance, m_part.variance, places=15)
        self.assertAlmostEqual(m_full.skewness, m_part.skewness, places=13)
        self.assertAlmostEqual(m_full.kurtosis, m_part.kurtosis, places=12)

    def test_revert_lifo_roundtrip(self):
        m = CentralMomentsKleinKBN(ddof=0)
        for x in self.bacon_data:
            m.update(x)
        for x in reversed(self.bacon_data):
            m.revert(x)

        self.assertEqual(m.n, 0)
        self.assertEqual(m.mean, 0.0)
        self.assertEqual(m.variance, 0.0)

    def test_reset(self):
        m = CentralMomentsKleinKBN()
        m.update(10.0)
        m.reset()
        self.assertEqual(m.n, 0)
        self.assertEqual(m.mean, 0.0)
        self.assertEqual(m.variance, 0.0)
