import unittest

from .percentile import percentile

class TestPercentile(unittest.TestCase):

    def check_window(self, window, description, expected):
        q_values = (0, 0.01, 0.05, 0.1, 0.2, 0.25, 0.3, 0.4, 0.5, 0.6, 0.7, 0.75, 0.8, 0.9, 0.95, 0.99, 1)

        for i, (q, expected) in enumerate(zip(q_values, expected)):
            actual = percentile(window, q)
            self.assertAlmostEqual(actual, expected, places=14,
                msg=f'{description} step {i} q {q}: expected {expected}, got {actual}')

        # Property test: monotonicity
        last = float("-inf")
        for q in range(101):
            q *= 0.01
            current = percentile(window, q)
            self.assertGreaterEqual(current, last,
                msg=f'{description} step {i} q {q}: monotonicity: last {last}, current {current}')
            last = current

        # Property test: affine invariance
        # Percentiles satisfy
        # $$P_q(aX+b)=aP_q(X)+b$, a\gt 0$
        scaled = [3*x + 7 for x in window]
        for q in range(101):
            q *= 0.01
            act = 3 * percentile(window, q) + 7
            exp = percentile(scaled, q)
            self.assertAlmostEqual(act, exp, places=13,
                msg=f'{description} step {i} q {q}: affine invariance: expected {exp}, actual {act}')

    def test_reference_dataset(self):
        win = (-5.453279550656607, -3.6648332058049427, 5.947309146654682, 3.525093415019491,
            -2.1778089879618197, -3.34372144267231, 1.9661750717437965, -6.265316287925733,
            3.4551208802924265, 8.836057305398743, -5.03508570740858, 8.977623036666365,
            3.3447490620074483, -8.082041288117757, -1.1632066766437443, 7.729598386550354,
            3.949069997640443, -3.4705427185977573, 4.67856326660133, -5.597300889090276,
            -8.368108609155838, -6.802087978499049, -3.197996300905894, -0.6961369259589816,
            -4.671579434184581, 6.315528068496139, -6.13411221421011, -7.410618476455994,
            -8.166704969101282, 1.971360273298263, 7.094838087480028, 2.0324248338742628,
            8.63976722271967, 4.495627221840401, 7.211026347865847, 8.586756031506326,
            0.9237201816470613, 8.75345917535514, -0.10024119842351453, -4.5245363502002505,
            -0.9644258505047869, 3.3007784679906056)

        exp = (-8.368108609155838, -8.28553311673347, -8.048470147534669, -6.748410809441717,
            -5.369640782007001, -4.634818663188499, -3.6065460596427874, -1.7719680634345873,
            0.41173949161177337, 2.793437014344066, 3.821877022854157, 4.632829255411098,
            6.241884284127849, 8.501040267010728, 8.747774577723366, 8.91958108684664,
            8.977623036666365)

        self.check_window(win, "reference dataset", exp)

    def test_algorithm(self):
        ex = (42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42, 42)
        self.check_window([42], "one element", ex)

        ex = (10.0, 10.1, 10.5, 11.0, 12.0, 12.5, 13, 14, 15, 16, 17, 17.5, 18, 19, 19.5, 19.9, 20)
        self.check_window((10, 20), "two elements", ex)

        ex = (1.0, 1.04, 1.2, 1.4, 1.8, 2.0, 2.2, 2.6, 3.0, 3.4, 3.8, 4.0, 4.2, 4.6, 4.8, 4.96, 5.0) 
        self.check_window((1, 2, 3, 4, 5), "sorted odd number of elements", ex)
        self.check_window((5, 2, 1, 4, 3), "unsorted odd number of elements", ex)
        self.check_window((5, 4, 3, 2, 1), "reverse sorted odd number of elements", ex)

        ex = (1.0, 1.03, 1.15, 1.3, 1.6, 1.75, 1.9, 2.2, 2.5, 2.8, 3.1, 3.25, 3.4, 3.7, 3.85, 3.97, 4.0) 
        self.check_window((1, 2, 3, 4), "sorted even number of elements", ex)
        self.check_window((3, 2, 4, 1), "unsorted even number of elements", ex)
        self.check_window((4, 3, 2, 1), "reverse sorted even number of elements", ex)

        ex = (1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.2, 1.6, 2.0, 2.0, 2.0, 2.0, 2.2, 2.6, 2.8, 2.96, 3.0) 
        self.check_window((1, 1, 2, 2, 3), "duplicate elements", ex)

        ex = (1.0, 1.04, 1.2, 1.4, 1.8, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.6, 3.80, 4.4, 4.88, 5.0) 
        self.check_window((1, 2, 2, 2, 5), "more duplicate elements", ex)

        ex = (2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0, 2.0) 
        self.check_window((2, 2, 2, 2, 2), "equal elements", ex)

        ex = (-10.0, -9.8, -9.0, -8.0, -6.0, -5.0, -4.0, -2.0, 0.0, 2.0, 4.0, 5.0, 6.0, 8.0, 9.0, 9.8, 10.0) 
        self.check_window((-10, -5, 0, 5, 10), "negative elements", ex)

        ex = (-2.71828, -2.6372748, -2.313254, -1.908228, -1.098176, -0.69315, -0.439078, 0.069066,
            0.57721, 0.91201, 1.24681, 1.41421, 1.759686, 2.450638, 2.796114, 3.0724948, 3.14159) 
        self.check_window((3.14159, -2.71828, 0.57721, 1.41421, -0.69315), "floating-point elements", ex)

        # Q out of range
        with self.assertRaises(ValueError):
            percentile([42], -0.01)
        with self.assertRaises(ValueError):
            percentile([42], 1.01)

        # Empty or None window
        with self.assertRaises(ValueError):
            percentile([], 42)
