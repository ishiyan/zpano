import math
import unittest

from .klein_kbn_summator import KleinKBNSummator


class TestKleinKBNSummator(unittest.TestCase):

    def test_empty(self):
        s = KleinKBNSummator()
        self.assertEqual(s.n, 0)
        self.assertEqual(s.value, 0.0)
        self.assertTrue(math.isnan(s.mean))

    def test_update(self):
        s = KleinKBNSummator()
        for x in [1.0, 2.0, 3.0, 4.0]:
            s.update(x)
        self.assertEqual(s.n, 4)
        self.assertEqual(s.value, 10.0)
        self.assertEqual(s.mean, 2.5)

    def test_zero_is_counted(self):
        s = KleinKBNSummator()
        for x in [0.0, 3.0, 0.0]:
            s.update(x)
        self.assertEqual(s.n, 3)
        self.assertEqual(s.value, 3.0)
        self.assertEqual(s.mean, 1.0)

    def test_compensated(self):
        # Peters' example: naive summation yields 0.0.
        s = KleinKBNSummator()
        for x in [1.0, 1e100, 1.0, -1e100]:
            s.update(x)
        self.assertEqual(s.value, 2.0)
        self.assertEqual(s.mean, 0.5)

    def test_revert(self):
        s = KleinKBNSummator()
        for x in [1.0, 2.0, 0.0, 4.0]:
            s.update(x)
        s.revert(0.0)
        self.assertEqual(s.n, 3)
        self.assertEqual(s.value, 7.0)
        s.revert(1.0)  # not the most recent value
        self.assertEqual(s.n, 2)
        self.assertEqual(s.value, 6.0)
        self.assertEqual(s.mean, 3.0)

    def test_rolling_window(self):
        data = [0.003, 0.026, 0.011, -0.010, 0.015, 0.025, 0.016, 0.067]
        w = 3
        s = KleinKBNSummator()
        for i, x in enumerate(data):
            s.update(x)
            if i >= w:
                s.revert(data[i - w])
            window = data[max(0, i - w + 1):i + 1]
            self.assertEqual(s.n, len(window))
            self.assertAlmostEqual(s.value, math.fsum(window), places=17)

    def test_revert_to_empty(self):
        s = KleinKBNSummator()
        s.update(5.0)
        s.revert(5.0)
        self.assertEqual(s.n, 0)
        self.assertEqual(s.value, 0.0)
        self.assertTrue(math.isnan(s.mean))

    def test_revert_empty_raises(self):
        s = KleinKBNSummator()
        with self.assertRaises(ValueError):
            s.revert(1.0)

    def test_revert_to_empty_clears_compensation(self):
        for final_zero in (False, True):
            with self.subTest(final_zero=final_zero):
                s = KleinKBNSummator()
                for x in [0.1, 1e16, 1e32, 1e48]:
                    s.update(x)
                if final_zero:
                    s.update(0.0)
                for x in [1e16, 0.1, 1e32, 1e48]:
                    s.revert(x)
                if final_zero:
                    self.assertEqual(s.n, 1)
                    s.revert(0.0)
                self.assertEqual(s.n, 0)
                self.assertEqual(s.value, 0.0)
                self.assertTrue(math.isnan(s.mean))
                s.update(3.0)
                self.assertEqual(s.n, 1)
                self.assertEqual(s.value, 3.0)
                self.assertEqual(s.mean, 3.0)

    def test_reset(self):
        s = KleinKBNSummator()
        for x in [1.0, 2.0]:
            s.update(x)
        s.reset()
        self.assertEqual(s.n, 0)
        self.assertEqual(s.value, 0.0)
        self.assertTrue(math.isnan(s.mean))
        s.update(3.0)
        self.assertEqual(s.mean, 3.0)


if __name__ == '__main__':
    unittest.main()
