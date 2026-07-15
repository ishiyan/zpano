import math
import unittest
import numpy as np

from .klein_kbn_accumulator import KleinKBNAccumulator


class NaiveSum:
    def __init__(self) -> None:
        self._value = 0.0

    def reset(self) -> None:
        self._value = 0.0

    def set(self, x) -> None:
        self._value = x

    def update(self, x) -> None:
        self._value += x
    @property
    def value(self) -> float:
        return self._value


##########################################################
# KBN Accumulator Tests
##########################################################

class TestKleinKBNAccumulator(unittest.TestCase):

    def setUp(self):
        # https://en.wikipedia.org/wiki/Kahan_summation_algorithm
        # For many sequences of numbers, both algorithms agree,
        # but a simple example due to Peters[11] shows how they can differ: summing 
        # [1.0, +1e100, 1.0, -1e100] in double precision, Kahan's algorithm yields 0.0,
        # whereas Neumaier's algorithm yields the correct value 2.0.
        self.peters_data = [1.0, +1e100, 1.0, -1e100]

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
        self.numpy_expected = -0.377392919181026

    def test_peters(self):
        naive = NaiveSum()
        kbn = KleinKBNAccumulator()
        for x in self.peters_data:
            naive.update(x)
            kbn.update(x)
        v = naive.value
        k = kbn.value
        n = np.sum(self.peters_data)
        m = math.fsum(self.peters_data)
        print(f'\nExact sum (Peters): 2.0')
        print(f'fsum: {m} (error: {abs(m - 2.0)})')
        print(f'NumPy: {n} (error: {abs(n - 2.0)})')
        print(f'KBN: {k} (error: {abs(k - 2.0)})')
        print(f'Naive: {v} (error: {abs(v - 2.0)})\n')
        self.assertAlmostEqual(k, 2.0, places=15, msg=f'KBN sum {k} is not equal to 2.0')
        #self.assertAlmostEqual(v, 2.0, places=15, msg=f'Naive sum {v} is not equal to 2.0')

    def test_numpy(self):
        naive = NaiveSum()
        kbn = KleinKBNAccumulator()
        for x in self.numpy_data:
            naive.update(x)
            kbn.update(x)
        v = naive.value
        k = kbn.value
        n = np.sum(self.numpy_data)
        m = math.fsum(self.numpy_data)
        print(f'\nExact sum: {self.numpy_expected}')
        print(f'fsum: {m} (error: {abs(m - self.numpy_expected)})')
        print(f'NumPy: {n} (error: {abs(n - self.numpy_expected)})')
        print(f'KBN: {k} (error: {abs(k - self.numpy_expected)})')
        print(f'Naive: {v} (error: {abs(v - self.numpy_expected)})\n')
        self.assertAlmostEqual(k, self.numpy_expected, places=16, msg=f'KBN sum {k} is not equal to {self.numpy_expected}')
        #self.assertAlmostEqual(v, self.numpy_expected, places=16, msg=f'Naive sum {v} is not equal to {self.numpy_expected}')

    def test_better_accuracy_than_naive(self):
        spread = 1e7
        naive = NaiveSum()
        kbn = KleinKBNAccumulator()

        rng = np.random.default_rng(seed=42)
        for x in rng.uniform(size=1000000):
            x *= spread
            naive.update(x)
            kbn.update(x)

        rng = np.random.default_rng(seed=42)
        for x in rng.uniform(size=1000000):
            x *= spread
            naive.update(-x)
            kbn.update(-x)

        v = naive.value
        k = kbn.value
        self.assertTrue(abs(k) <= abs(v), msg=f'KBN sum {k} is not more accurate than naive sum {v}')

    def test_revert(self):
        kbn = KleinKBNAccumulator()
        self.assertAlmostEqual(kbn.value, 0.0, places=15)

        # update then revert should restore prior value
        kbn.update(1.5)
        kbn.update(2.5)
        expected_before = kbn.value
        kbn.revert(2.5)
        self.assertAlmostEqual(kbn.value, 1.5, places=15)
        kbn.revert(1.5)
        self.assertAlmostEqual(kbn.value, 0.0, places=15)

    def test_reset(self):
        kbn = KleinKBNAccumulator()
        kbn.update(1.5)
        kbn.reset()
        self.assertAlmostEqual(kbn.value, 0.0, places=15)

        kbn.update(1.5)
        self.assertAlmostEqual(kbn.value, 1.5, places=15)
