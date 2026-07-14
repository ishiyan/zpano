import unittest
import random
import numpy as np

from .primitives import Variance, RunningVariance

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
