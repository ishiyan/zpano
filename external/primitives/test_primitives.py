# To run tests, ensure you are in the root directory of the repository and run:
# python -m unittest external.primitives.test_primitives.TestCompareStats

import math
import unittest
import random
import numpy as np
from scipy import stats

from .primitives import KahanWelfordVariance, Variance, RunningVariance, RawMoments, CentralMoments
from .raw_moments_klein_kbn import RawMomentsKleinKBN
from .central_moments_klein_kbn import CentralMomentsKleinKBN

##########################################################
# Mean and Variance with update/revert methods
##########################################################

class TestVariance(unittest.TestCase):

    def setUp(self):
        self.data = [
            99.99999978143265, 99.99999989071631, 99.99999994535816,
            99.99999997267908, 99.99999998633952, 99.99999999316977,
            99.99999999829245, 99.99999999957309]

    def test_update(self):
        variance = Variance()
        for x in self.data:
            variance.update(x)
        v = variance.variance
        self.assertIs(0 < v, True, msg=f'Variance {v} is not positive')

    def test_revert(self):
        for _ in range(5):
            X = [random.random() for _ in range(20)]

            variance1 = Variance()
            variance2 = Variance()

            for x in X[:10]:
                variance1.update(x)
                variance2.update(x)

            for x in X[10:]:
                variance2.update(x)
            for x in X[10:]:
                variance2.revert(x)

            v1 = variance1.variance
            v2 = variance2.variance
            self.assertAlmostEqual(v1, v2, places=15)

    # https://nbviewer.org/github/changyaochen/changyaochen.github.io/blob/master/assets/notebooks/welford.ipynb#ns
    def test_numerical_stability(self):
        np.random.seed(42)
        d = np.random.normal(size=int(1e6))
        variance = Variance()
        for x in d:
            variance.update(x)
        #self.assertAlmostEqual(variance.variance, 1.00037611618, places=15)
        self.assertAlmostEqual(variance.variance, 1.00037611618, places=5)

class TestRollingVariance(unittest.TestCase):

    def setUp(self):
        self.data = [9., 7., 3., 2., 6., 1., 8., 5., 4.]
        
    def test_window_2(self):
        rv = RunningVariance(ddof=1, window_size=2)
        for x in self.data:
            rv.update(x)
            v = rv.variance
        self.assertAlmostEqual(rv.variance, 0.5, places=15)
        
    def test_window_none(self):
        rv = RunningVariance(ddof=1, window_size=None)
        for x in self.data:
            rv.update(x)
            v = rv.variance
        self.assertAlmostEqual(rv.variance, 7.5, places=14)
        
    def test_window_zero(self):
        rv = RunningVariance(ddof=1, window_size=0)
        for x in self.data:
            rv.update(x)
            v = rv.variance
        self.assertAlmostEqual(rv.variance, 7.5, places=14)
        
    def test_mean_variance(self):
        rv = RunningVariance(ddof=1, window_size=None)
        for x in range(100):
            rv.update(x)
        self.assertAlmostEqual(rv.mean, 49.5, places=14)
        self.assertAlmostEqual(rv.variance, 841.6666666666666, places=14)
        self.assertAlmostEqual(rv.variance**0.5, 29.0114919759, places=10)
        
    def test_mean_variance_2(self):
        rv = RunningVariance(ddof=1, window_size=None)
        for x in range(101):
            rv.update(x-1)
        self.assertAlmostEqual(rv.mean, 50.0, places=14)
        self.assertAlmostEqual(rv.variance, 841.6666666666666, places=14)

##########################################################
# Various versios of stats
##########################################################

class TestCompareStats(unittest.TestCase):

    def setUp(self):
        # https://en.wikipedia.org/wiki/Kahan_summation_algorithm
        # For many sequences of numbers, both algorithms agree,
        # but a simple example due to Peters[11] shows how they can differ: summing 
        # [1.0, +1e100, 1.0, -1e100] in double precision, Kahan's algorithm yields 0.0,
        # whereas Neumaier's algorithm yields the correct value 2.0.
        self.peters_data = [1.0, +1e100, 1.0, -1e100]
        self.peters_expected = 0.5

        # https://github.com/numpy/numpy/issues/8786
        #  A badly conditioned sum, condition number ~2.188e+14
        self.numpy_data = [
            -0.41253261766461263,
            41287272281118.43,
            -1.4727977348624173e-14,
            5670.3302557520055,
            2.119245229045646e-11,
            -0.003679264134906428,
            -6.892634568678797e-14,
            -0.0006984744181630712,
            -4054136.048352595,
            -1003.101760720037,
            -1.4436349910427172e-17,
            -41287268231649.57]
        self.numpy_expected = -0.377392919181026/12.0

        self.bacon_data = [
            0.003, 0.026, 0.011,-0.010,
            0.015, 0.025, 0.016, 0.067,
            -0.014,0.040,-0.005, 0.081,
            0.040,-0.037,-0.061, 0.017,
            -0.049,-0.022,0.070, 0.058,
            -0.065,0.024,-0.005,-0.009]
        self.expected_kurtosis = [
            None, -2.00000000000000000, -1.50000000000000000,
            -1.17592035552795000, -0.94669079980875600, -0.96028723389787100,
            -0.57793300076120100, 0.78641242115027200, 0.59954237086621500,
            -0.01187577489273160, 0.07517391430462480, -0.27406990671095100,
            -0.38022416153835900, -0.31560370425738600, -0.16235155227201600,
            0.02528905226985100, -0.33285099821964000, -0.37425348407483000,
            -0.58502674157514900, -0.69334606360953100, -0.77381631285861200,
            -0.68208349704651200, -0.61779722177118000,  -0.56754620589212500]


    def test_peters(self):
        kbn = RawMomentsKleinKBN(ddof=0)
        kwv = KahanWelfordVariance(ddof=0)
        st2 = RawMoments(ddof=0)
        mom = CentralMoments(ddof=0)
        var = Variance(ddof=0)
        for x in self.peters_data:
            kbn.update(x)
            kwv.update(x)
            st2.update(x)
            mom.update(x)
            var.update(x)
        k_m = kbn.mean
        k_v = kbn.variance
        k_s = kbn.skewness
        k_k = kbn.kurtosis
        kwv_m = kwv.mean
        kwv_v = kwv.variance
        st2_m = st2.mean
        st2_v = st2.variance
        st2_s = st2.skewness
        st2_k = st2.kurtosis
        mom_m = mom.mean
        mom_v = mom.variance
        mom_s = mom.skewness
        mom_k = mom.kurtosis
        var_m = var.mean
        var_v = var.variance
        np_m = np.mean(self.peters_data)
        np_v = np.var(self.peters_data, ddof=0)
        sp_m = stats.tmean(self.peters_data)
        sp_v = stats.tvar(self.peters_data, ddof=0)
        sp_s = stats.skew(self.peters_data, bias=False)
        sp_k = stats.kurtosis(self.peters_data, bias=False)

        print(f'\nExact mean (Peters): {self.peters_expected}')
        print(f'KBN mean: {k_m} (error: {abs(k_m - self.peters_expected)})')
        print(f'Kahan-Welford mean: {kwv_m} (error: {abs(kwv_m - self.peters_expected)})')
        print(f'RawMoments mean: {st2_m} (error: {abs(st2_m - self.peters_expected)})')
        print(f'CentralMoments mean: {mom_m} (error: {abs(mom_m - self.peters_expected)})')
        print(f'Variance mean: {var_m} (error: {abs(var_m - self.peters_expected)})')
        print(f'NumPy mean: {np_m} (error: {abs(np_m - self.peters_expected)})')
        print(f'SciPy mean: {sp_m} (error: {abs(sp_m - self.peters_expected)})')
        print(f'KBN variance: {k_v}')
        print(f'Kahan-Welford variance: {kwv_v}')
        print(f'RawMoments variance: {st2_v}')
        print(f'CentralMoments variance: {mom_v}')
        print(f'Variance variance: {var_v}')
        print(f'NumPy variance: {np_v}')
        print(f'SciPy variance: {sp_v}')
        print(f'KBN skewness: {k_s}')
        print(f'RawMoments skewness: {st2_s}')
        print(f'CentralMoments skewness: {mom_s}')
        print(f'NumPy skewness: {sp_s}')
        print(f'SciPy skewness: {sp_s}')
        print(f'KBN kurtosis: {k_k}')
        print(f'RawMoments kurtosis: {st2_k}')
        print(f'CentralMoments kurtosis: {mom_k}')
        print(f'NumPy kurtosis: {sp_k}')
        print(f'SciPy kurtosis: {sp_k}\n')

        self.assertAlmostEqual(k_m, 0.5, places=16, msg=f'RawMomentsKleinKBN mean {k_m} is not equal to 0.5')
        #self.assertAlmostEqual(k_v, 1.0, places=16, msg=f'RawMomentsKleinKBN variance {k_v} is not equal to 1.0')
        #self.assertAlmostEqual(k_s, 1.0, places=16, msg=f'RawMomentsKleinKBN skewness {k_s} is not equal to 1.0')
        #self.assertAlmostEqual(k_k, 1.0, places=16, msg=f'RawMomentsKleinKBN kurtosis {k_k} is not equal to 1.0')

    def test_numpy(self):
        kbn = RawMomentsKleinKBN(ddof=0)
        kwv = KahanWelfordVariance(ddof=0)
        st2 = RawMoments(ddof=0)
        mom = CentralMoments(ddof=0)
        var = Variance(ddof=0)
        for x in self.numpy_data:
            kbn.update(x)
            kwv.update(x)
            st2.update(x)
            mom.update(x)
            var.update(x)
        k_m = kbn.mean
        k_v = kbn.variance
        k_s = kbn.skewness
        k_k = kbn.kurtosis
        kwv_m = kwv.mean
        kwv_v = kwv.variance
        st2_m = st2.mean
        st2_v = st2.variance
        st2_s = st2.skewness
        st2_k = st2.kurtosis
        mom_m = mom.mean
        mom_v = mom.variance
        mom_s = mom.skewness
        mom_k = mom.kurtosis
        var_m = var.mean
        var_v = var.variance
        np_m = np.mean(self.numpy_data)
        np_v = np.var(self.numpy_data, ddof=0)
        sp_m = stats.tmean(self.numpy_data)
        sp_v = stats.tvar(self.numpy_data, ddof=0)
        sp_s = stats.skew(self.numpy_data, bias=False)
        sp_k = stats.kurtosis(self.numpy_data, bias=False)

        print(f'\nExact mean (NumPy): {self.numpy_expected}')
        print(f'KBN mean: {k_m} (error: {abs(k_m - self.numpy_expected)})')
        print(f'Kahan-Welford mean: {kwv_m} (error: {abs(kwv_m - self.numpy_expected)})')
        print(f'RawMoments mean: {st2_m} (error: {abs(st2_m - self.numpy_expected)})')
        print(f'CentralMoments mean: {mom_m} (error: {abs(mom_m - self.numpy_expected)})')
        print(f'Variance mean: {var_m} (error: {abs(var_m - self.numpy_expected)})')
        print(f'NumPy mean: {np_m} (error: {abs(np_m - self.numpy_expected)})')
        print(f'SciPy mean: {sp_m} (error: {abs(sp_m - self.numpy_expected)})')
        print(f'KBN variance: {k_v}')
        print(f'Kahan-Welford variance: {kwv_v}')
        print(f'RawMoments variance: {st2_v}')
        print(f'CentralMoments variance: {mom_v}')
        print(f'Variance variance: {var_v}')
        print(f'NumPy variance: {np_v}')
        print(f'SciPy variance: {sp_v}')
        print(f'KBN skewness: {k_s}')
        print(f'RawMoments skewness: {st2_s}')
        print(f'CentralMoments skewness: {mom_s}')
        print(f'NumPy skewness: {sp_s}')
        print(f'SciPy skewness: {sp_s}')
        print(f'KBN kurtosis: {k_k}')
        print(f'RawMoments kurtosis: {st2_k}')
        print(f'CentralMoments kurtosis: {mom_k}')
        print(f'NumPy kurtosis: {sp_k}')
        print(f'SciPy kurtosis: {sp_k}\n')

        self.assertAlmostEqual(k_m, self.numpy_expected, places=16, msg=f'RawMomentsKleinKBN mean {k_m} is not equal to {self.numpy_expected}')
        #self.assertAlmostEqual(k_v, 1.0, places=16, msg=f'RawMomentsKleinKBN variance {k_v} is not equal to 1.0')
        #self.assertAlmostEqual(k_s, 1.0, places=16, msg=f'RawMomentsKleinKBN skewness {k_s} is not equal to 1.0')
        #self.assertAlmostEqual(k_k, 1.0, places=16, msg=f'RawMomentsKleinKBN kurtosis {k_k} is not equal to 1.0')


    def test_bacon(self):
        kbn = RawMomentsKleinKBN(ddof=0)
        st2 = RawMoments(ddof=0)
        mom = CentralMoments()
        mok = CentralMomentsKleinKBN(ddof=0)
        var = Variance(ddof=0)
        for x in self.bacon_data:
            kbn.update(x)
            st2.update(x)
            mom.update(x)
            var.update(x)
            mok.update(x)
        k_m = kbn.mean
        k_v = kbn.variance
        k_s = kbn.skewness
        k_k = kbn.kurtosis
        st2_m = st2.mean
        st2_v = st2.variance
        st2_s = st2.skewness
        st2_k = st2.kurtosis
        mom_m = mom.mean
        mom_v = mom.variance
        mom_s = mom.skewness
        mom_k = mom.kurtosis
        mok_m = mok.mean
        mok_v = mok.variance
        mok_s = mok.skewness
        mok_k = mok.kurtosis
        var_m = var.mean
        var_v = var.variance
        sp_m = stats.tmean(self.bacon_data)
        sp_v = stats.tvar(self.bacon_data, ddof=0)
        sp_s = stats.skew(self.bacon_data, bias=True)
        sp_k = stats.kurtosis(self.bacon_data, bias=True, fisher=True)

        print(f'\nExact kurtosis (Bacon): -0.56754620589212500')
        print(f'KBN mean: {k_m}')
        print(f'RawMoments mean: {st2_m}')
        print(f'CentralMoments mean: {mom_m}')
        print(f'CentralMomentsKleinKBN mean: {mok_m}')
        print(f'Variance mean: {var_m}')
        print(f'SciPy mean: {sp_m}')
        print(f'KBN variance: {k_v}')
        print(f'RawMoments variance: {st2_v}')
        print(f'CentralMoments variance: {mom_v}')
        print(f'CentralMomentsKleinKBN variance: {mok_v}')
        print(f'Variance variance: {var_v}')
        print(f'SciPy variance: {sp_v}')
        print(f'KBN skewness: {k_s}')
        print(f'RawMoments skewness: {st2_s}')
        print(f'CentralMoments skewness: {mom_s}')
        print(f'CentralMomentsKleinKBN skewness: {mok_s}')
        print(f'SciPy skewness: {sp_s}')
        print(f'KBN kurtosis: {k_k}')
        print(f'RawMoments kurtosis: {st2_k}')
        print(f'CentralMoments kurtosis: {mom_k}')
        print(f'CentralMomentsKleinKBN kurtosis: {mok_k}')
        print(f'SciPy kurtosis: {sp_k}\n')

        self.assertAlmostEqual(mok_m, sp_m, places=15)
        self.assertAlmostEqual(mok_v, sp_v, places=15)
        self.assertAlmostEqual(mok_s, sp_s, places=14)
        self.assertAlmostEqual(mok_k, sp_k, places=13)

        self.assertAlmostEqual(k_s, sp_s, places=14)
        self.assertAlmostEqual(st2_s, sp_s, places=14)
        self.assertAlmostEqual(k_k, sp_k, places=13)
        self.assertAlmostEqual(st2_k, sp_k, places=13)

    def test_bacon_bias_false(self):
        kbn = RawMomentsKleinKBN(ddof=0, bias=False, fisher=True)
        st2 = RawMoments(ddof=0, bias=False, fisher=True)
        mom = CentralMoments(ddof=0, bias=False, fisher=True)
        mok = CentralMomentsKleinKBN(ddof=0, bias=False, fisher=True)
        for x in self.bacon_data:
            kbn.update(x)
            st2.update(x)
            mom.update(x)
            mok.update(x)
        sp_s = stats.skew(self.bacon_data, bias=False)
        sp_k = stats.kurtosis(self.bacon_data, bias=False, fisher=True)

        self.assertAlmostEqual(kbn.skewness, sp_s, places=14)
        self.assertAlmostEqual(st2.skewness, sp_s, places=14)
        self.assertAlmostEqual(mom.skewness, sp_s, places=14)
        self.assertAlmostEqual(mok.skewness, sp_s, places=14)
        self.assertAlmostEqual(kbn.kurtosis, sp_k, places=13)
        self.assertAlmostEqual(st2.kurtosis, sp_k, places=13)
        self.assertAlmostEqual(mom.kurtosis, sp_k, places=13)
        self.assertAlmostEqual(mok.kurtosis, sp_k, places=13)

    def test_revert_lifo_simple(self):
        data = [10.0, 18.0, 5.0]
        mom_full = CentralMoments(ddof=0)
        mok_full = CentralMomentsKleinKBN(ddof=0)
        for x in data:
            mom_full.update(x)
            mok_full.update(x)

        mom_first2 = CentralMoments(ddof=0)
        mok_first2 = CentralMomentsKleinKBN(ddof=0)
        for x in data[:2]:
            mom_first2.update(x)
            mok_first2.update(x)

        mom_full.revert(data[2])
        mok_full.revert(data[2])

        self.assertAlmostEqual(mom_full.mean, mom_first2.mean, places=15)
        self.assertAlmostEqual(mom_full.variance, mom_first2.variance, places=15)
        self.assertAlmostEqual(mom_full.skewness, mom_first2.skewness, places=14)
        self.assertAlmostEqual(mom_full.kurtosis, mom_first2.kurtosis, places=13)

        self.assertAlmostEqual(mok_full.mean, mok_first2.mean, places=15)
        self.assertAlmostEqual(mok_full.variance, mok_first2.variance, places=15)
        self.assertAlmostEqual(mok_full.skewness, mok_first2.skewness, places=14)
        self.assertAlmostEqual(mok_full.kurtosis, mok_first2.kurtosis, places=13)

    def test_revert_lifo_bacon(self):
        mom_full = CentralMoments(ddof=0)
        mom_part = CentralMoments(ddof=0)
        mok_full = CentralMomentsKleinKBN(ddof=0)
        mok_part = CentralMomentsKleinKBN(ddof=0)
        for x in self.bacon_data:
            mom_full.update(x)
            mok_full.update(x)
        for x in self.bacon_data[:-1]:
            mom_part.update(x)
            mok_part.update(x)

        mom_full.revert(self.bacon_data[-1])
        mok_full.revert(self.bacon_data[-1])

        self.assertAlmostEqual(mom_full.mean, mom_part.mean, places=15)
        self.assertAlmostEqual(mom_full.variance, mom_part.variance, places=15)
        self.assertAlmostEqual(mom_full.skewness, mom_part.skewness, places=13)
        self.assertAlmostEqual(mom_full.kurtosis, mom_part.kurtosis, places=12)

        self.assertAlmostEqual(mok_full.mean, mok_part.mean, places=15)
        self.assertAlmostEqual(mok_full.variance, mok_part.variance, places=15)
        self.assertAlmostEqual(mok_full.skewness, mok_part.skewness, places=13)
        self.assertAlmostEqual(mok_full.kurtosis, mok_part.kurtosis, places=12)

    def test_revert_lifo_roundtrip(self):
        mom = CentralMoments(ddof=0)
        mok = CentralMomentsKleinKBN(ddof=0)
        for x in self.bacon_data:
            mom.update(x)
            mok.update(x)
        for x in reversed(self.bacon_data):
            mom.revert(x)
            mok.revert(x)

        self.assertEqual(mom.n, 0)
        self.assertEqual(mom.mean, 0.0)
        self.assertEqual(mom.variance, 0.0)
        self.assertEqual(mok.n, 0)
        self.assertEqual(mok.mean, 0.0)
        self.assertEqual(mok.variance, 0.0)

