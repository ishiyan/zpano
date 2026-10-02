import math
import random
import unittest

from .klein_kbn_accumulator import KleinKBNAccumulator


def naive_sum(data) -> float:
    s = 0.0
    for x in data:
        s += x
    return s


def kbn_sum(data) -> float:
    kbn = KleinKBNAccumulator()
    for x in data:
        kbn.update(x)
    return kbn.value


class TestKleinKBNAccumulator(unittest.TestCase):

    def setUp(self):
        # https://en.wikipedia.org/wiki/Kahan_summation_algorithm
        # A simple example due to Peters: summing [1.0, +1e100, 1.0, -1e100]
        # in double precision, Kahan's algorithm yields 0.0, whereas
        # Neumaier's algorithm yields the correct value 2.0.
        self.peters_data = [1.0, +1e100, 1.0, -1e100]

        # https://github.com/numpy/numpy/issues/8786
        # A badly conditioned sum, condition number ~2.188e+14.
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

        # A sequence where the second-level correction is non-zero at more
        # than one step, so it must be accumulated (ccs += cc), not
        # overwritten (ccs = cc).  Overwriting yields -1.0000000000000001e-16.
        self.klein_data = [1e-16, -1e16, 1.0, 1e-16, -1.0, -1e-16, -1e-32, 1e16]
        self.klein_expected = 9.999999999999999e-17  # math.fsum(klein_data)

    def test_initial_value_is_zero(self):
        self.assertEqual(KleinKBNAccumulator().value, 0.0)

    def test_peters(self):
        self.assertEqual(naive_sum(self.peters_data), 0.0)
        self.assertEqual(kbn_sum(self.peters_data), 2.0)

    def test_numpy_issue(self):
        self.assertAlmostEqual(kbn_sum(self.numpy_data), self.numpy_expected, places=16)
        self.assertNotAlmostEqual(naive_sum(self.numpy_data), self.numpy_expected, places=3)

    def test_second_level_correction_is_accumulated(self):
        self.assertEqual(kbn_sum(self.klein_data), self.klein_expected)

    def test_matches_fsum_on_mixed_magnitudes(self):
        rng = random.Random(42)
        for _ in range(200):
            data = [rng.uniform(-1.0, 1.0) * 10.0 ** rng.choice([-8, 0, 8])
                    for _ in range(100)]
            self.assertEqual(kbn_sum(data), math.fsum(data))

    def test_better_accuracy_than_naive(self):
        # Add and then subtract the same values, so the exact sum is 0.
        rng = random.Random(42)
        data = [rng.uniform(0.0, 1e7) for _ in range(100000)]
        data += [-x for x in data]
        k = kbn_sum(data)
        v = naive_sum(data)
        self.assertEqual(k, 0.0)
        self.assertNotEqual(v, 0.0)

    def test_update_zero_keeps_compensation(self):
        kbn = KleinKBNAccumulator()
        for x in self.klein_data:
            kbn.update(x)
        kbn.update(0.0)
        self.assertEqual(kbn.value, self.klein_expected)

    def test_revert(self):
        kbn = KleinKBNAccumulator()
        kbn.update(1.5)
        kbn.update(2.5)
        kbn.revert(2.5)
        self.assertEqual(kbn.value, 1.5)
        kbn.revert(1.5)
        self.assertEqual(kbn.value, 0.0)

    def test_revert_not_most_recent(self):
        kbn = KleinKBNAccumulator()
        for x in self.peters_data:
            kbn.update(x)
        kbn.revert(1e100)  # not the most recent value
        self.assertEqual(kbn.value, 2.0 - 1e100)
        kbn.revert(-1e100)
        self.assertEqual(kbn.value, 2.0)

    def test_revert_restores_compensated_sum(self):
        kbn = KleinKBNAccumulator()
        for x in self.numpy_data:
            kbn.update(x)
        kbn.update(1e20)
        kbn.revert(1e20)
        self.assertAlmostEqual(kbn.value, self.numpy_expected, places=16)

    def test_set(self):
        kbn = KleinKBNAccumulator()
        for x in self.peters_data:
            kbn.update(x)
        kbn.set(5.0)
        self.assertEqual(kbn.value, 5.0)
        # Compensation terms are cleared, so only the new values count.
        kbn.update(1e100)
        kbn.update(-1e100)
        self.assertEqual(kbn.value, 5.0)

    def test_reset(self):
        kbn = KleinKBNAccumulator()
        for x in self.peters_data:
            kbn.update(x)
        kbn.reset()
        self.assertEqual(kbn.value, 0.0)
        kbn.update(1.5)
        self.assertEqual(kbn.value, 1.5)


if __name__ == '__main__':
    unittest.main()
