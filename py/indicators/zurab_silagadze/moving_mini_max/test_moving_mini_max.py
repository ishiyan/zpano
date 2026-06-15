import math
import unittest
from datetime import datetime

from py.indicators.zurab_silagadze.moving_mini_max.moving_mini_max import MovingMiniMax
from py.indicators.zurab_silagadze.moving_mini_max.params import MovingMiniMaxParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from . import test_testdata as td

TOLERANCE = 1e-9


def _assert_series(test, actual, expected, label):
    test.assertEqual(len(actual), len(expected), f"{label}: length mismatch ({len(actual)} vs {len(expected)})")
    for i in range(len(expected)):
        delta = TOLERANCE * max(1.0, abs(expected[i]))
        test.assertAlmostEqual(actual[i], expected[i], delta=delta,
                               msg=f"{label}[{i}]: expected {expected[i]}, got {actual[i]}")


def _assert_levels(test, actual, expected, label):
    test.assertEqual(len(actual), len(expected), f"{label}: length mismatch ({len(actual)} vs {len(expected)})")
    for i, (price, offset, strength) in enumerate(actual):
        exp = expected[i]
        test.assertAlmostEqual(price, exp['price'], delta=TOLERANCE * max(1.0, abs(exp['price'])),
                               msg=f"{label}[{i}].price")
        test.assertEqual(offset, exp['offset'], f"{label}[{i}].offset")
        test.assertAlmostEqual(strength, exp['strength'], delta=TOLERANCE * max(1.0, abs(exp['strength'])),
                               msg=f"{label}[{i}].strength")


def _run_last(inputs, m, n, num_extrema):
    """Feed the price series and return the last primed update tuple."""
    ind = MovingMiniMax(MovingMiniMaxParams(m=m, n=n, num_extrema=num_extrema))
    last = None
    for p in inputs:
        result = ind.update(p)
        if ind.is_primed():
            last = result
    return last


class TestMovingMiniMaxData(unittest.TestCase):
    """Test MMM against the reference test data for all parameter combos."""

    def _check(self, m, n, e, exp_up, exp_down, exp_res, exp_sup):
        label = f"m{m}_n{n}_e{e}"
        last = _run_last(td.INPUT_CLOSE, m, n, e)
        self.assertIsNotNone(last, f"{label}: no primed output")
        _, _, resistances, supports, up_dist, dn_dist = last
        _assert_series(self, up_dist, exp_up, f"{label} UP")
        _assert_series(self, dn_dist, exp_down, f"{label} DOWN")
        _assert_levels(self, resistances, exp_res, f"{label} RES")
        _assert_levels(self, supports, exp_sup, f"{label} SUP")

    def test_m3_n50_e1(self):
        self._check(3, 50, 1, td.EXPECTED_M3_N50_E1_UP, td.EXPECTED_M3_N50_E1_DOWN,
                    td.EXPECTED_M3_N50_E1_RESISTANCES, td.EXPECTED_M3_N50_E1_SUPPORTS)

    def test_m3_n50_e3(self):
        self._check(3, 50, 3, td.EXPECTED_M3_N50_E3_UP, td.EXPECTED_M3_N50_E3_DOWN,
                    td.EXPECTED_M3_N50_E3_RESISTANCES, td.EXPECTED_M3_N50_E3_SUPPORTS)

    def test_m3_n100_e1(self):
        self._check(3, 100, 1, td.EXPECTED_M3_N100_E1_UP, td.EXPECTED_M3_N100_E1_DOWN,
                    td.EXPECTED_M3_N100_E1_RESISTANCES, td.EXPECTED_M3_N100_E1_SUPPORTS)

    def test_m3_n100_e3(self):
        self._check(3, 100, 3, td.EXPECTED_M3_N100_E3_UP, td.EXPECTED_M3_N100_E3_DOWN,
                    td.EXPECTED_M3_N100_E3_RESISTANCES, td.EXPECTED_M3_N100_E3_SUPPORTS)

    def test_m3_n252_e1(self):
        self._check(3, 252, 1, td.EXPECTED_M3_N252_E1_UP, td.EXPECTED_M3_N252_E1_DOWN,
                    td.EXPECTED_M3_N252_E1_RESISTANCES, td.EXPECTED_M3_N252_E1_SUPPORTS)

    def test_m3_n252_e3(self):
        self._check(3, 252, 3, td.EXPECTED_M3_N252_E3_UP, td.EXPECTED_M3_N252_E3_DOWN,
                    td.EXPECTED_M3_N252_E3_RESISTANCES, td.EXPECTED_M3_N252_E3_SUPPORTS)

    def test_m5_n50_e1(self):
        self._check(5, 50, 1, td.EXPECTED_M5_N50_E1_UP, td.EXPECTED_M5_N50_E1_DOWN,
                    td.EXPECTED_M5_N50_E1_RESISTANCES, td.EXPECTED_M5_N50_E1_SUPPORTS)

    def test_m5_n50_e3(self):
        self._check(5, 50, 3, td.EXPECTED_M5_N50_E3_UP, td.EXPECTED_M5_N50_E3_DOWN,
                    td.EXPECTED_M5_N50_E3_RESISTANCES, td.EXPECTED_M5_N50_E3_SUPPORTS)

    def test_m5_n100_e1(self):
        self._check(5, 100, 1, td.EXPECTED_M5_N100_E1_UP, td.EXPECTED_M5_N100_E1_DOWN,
                    td.EXPECTED_M5_N100_E1_RESISTANCES, td.EXPECTED_M5_N100_E1_SUPPORTS)

    def test_m5_n100_e3(self):
        self._check(5, 100, 3, td.EXPECTED_M5_N100_E3_UP, td.EXPECTED_M5_N100_E3_DOWN,
                    td.EXPECTED_M5_N100_E3_RESISTANCES, td.EXPECTED_M5_N100_E3_SUPPORTS)

    def test_m5_n252_e1(self):
        self._check(5, 252, 1, td.EXPECTED_M5_N252_E1_UP, td.EXPECTED_M5_N252_E1_DOWN,
                    td.EXPECTED_M5_N252_E1_RESISTANCES, td.EXPECTED_M5_N252_E1_SUPPORTS)

    def test_m5_n252_e3(self):
        self._check(5, 252, 3, td.EXPECTED_M5_N252_E3_UP, td.EXPECTED_M5_N252_E3_DOWN,
                    td.EXPECTED_M5_N252_E3_RESISTANCES, td.EXPECTED_M5_N252_E3_SUPPORTS)

    def test_m10_n50_e1(self):
        self._check(10, 50, 1, td.EXPECTED_M10_N50_E1_UP, td.EXPECTED_M10_N50_E1_DOWN,
                    td.EXPECTED_M10_N50_E1_RESISTANCES, td.EXPECTED_M10_N50_E1_SUPPORTS)

    def test_m10_n50_e3(self):
        self._check(10, 50, 3, td.EXPECTED_M10_N50_E3_UP, td.EXPECTED_M10_N50_E3_DOWN,
                    td.EXPECTED_M10_N50_E3_RESISTANCES, td.EXPECTED_M10_N50_E3_SUPPORTS)

    def test_m10_n100_e1(self):
        self._check(10, 100, 1, td.EXPECTED_M10_N100_E1_UP, td.EXPECTED_M10_N100_E1_DOWN,
                    td.EXPECTED_M10_N100_E1_RESISTANCES, td.EXPECTED_M10_N100_E1_SUPPORTS)

    def test_m10_n100_e3(self):
        self._check(10, 100, 3, td.EXPECTED_M10_N100_E3_UP, td.EXPECTED_M10_N100_E3_DOWN,
                    td.EXPECTED_M10_N100_E3_RESISTANCES, td.EXPECTED_M10_N100_E3_SUPPORTS)

    def test_m10_n252_e1(self):
        self._check(10, 252, 1, td.EXPECTED_M10_N252_E1_UP, td.EXPECTED_M10_N252_E1_DOWN,
                    td.EXPECTED_M10_N252_E1_RESISTANCES, td.EXPECTED_M10_N252_E1_SUPPORTS)

    def test_m10_n252_e3(self):
        self._check(10, 252, 3, td.EXPECTED_M10_N252_E3_UP, td.EXPECTED_M10_N252_E3_DOWN,
                    td.EXPECTED_M10_N252_E3_RESISTANCES, td.EXPECTED_M10_N252_E3_SUPPORTS)

    def test_m20_n50_e1(self):
        self._check(20, 50, 1, td.EXPECTED_M20_N50_E1_UP, td.EXPECTED_M20_N50_E1_DOWN,
                    td.EXPECTED_M20_N50_E1_RESISTANCES, td.EXPECTED_M20_N50_E1_SUPPORTS)

    def test_m20_n50_e3(self):
        self._check(20, 50, 3, td.EXPECTED_M20_N50_E3_UP, td.EXPECTED_M20_N50_E3_DOWN,
                    td.EXPECTED_M20_N50_E3_RESISTANCES, td.EXPECTED_M20_N50_E3_SUPPORTS)

    def test_m20_n100_e1(self):
        self._check(20, 100, 1, td.EXPECTED_M20_N100_E1_UP, td.EXPECTED_M20_N100_E1_DOWN,
                    td.EXPECTED_M20_N100_E1_RESISTANCES, td.EXPECTED_M20_N100_E1_SUPPORTS)

    def test_m20_n100_e3(self):
        self._check(20, 100, 3, td.EXPECTED_M20_N100_E3_UP, td.EXPECTED_M20_N100_E3_DOWN,
                    td.EXPECTED_M20_N100_E3_RESISTANCES, td.EXPECTED_M20_N100_E3_SUPPORTS)

    def test_m20_n252_e1(self):
        self._check(20, 252, 1, td.EXPECTED_M20_N252_E1_UP, td.EXPECTED_M20_N252_E1_DOWN,
                    td.EXPECTED_M20_N252_E1_RESISTANCES, td.EXPECTED_M20_N252_E1_SUPPORTS)

    def test_m20_n252_e3(self):
        self._check(20, 252, 3, td.EXPECTED_M20_N252_E3_UP, td.EXPECTED_M20_N252_E3_DOWN,
                    td.EXPECTED_M20_N252_E3_RESISTANCES, td.EXPECTED_M20_N252_E3_SUPPORTS)


class TestMovingMiniMaxScalars(unittest.TestCase):
    def test_up_down_latest(self):
        last = _run_last(td.INPUT_CLOSE, 3, 50, 1)
        up, down, _, _, up_dist, dn_dist = last
        self.assertAlmostEqual(up, up_dist[-1], delta=1e-12)
        self.assertAlmostEqual(down, dn_dist[-1], delta=1e-12)


class TestMovingMiniMaxMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        ind = MovingMiniMax(MovingMiniMaxParams())
        self.assertEqual(ind.metadata().mnemonic, "mmm(5,50,3)")


class TestMovingMiniMaxMetadata(unittest.TestCase):
    def test_metadata(self):
        ind = MovingMiniMax(MovingMiniMaxParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.MOVING_MINI_MAX)
        self.assertEqual(len(meta.outputs), 6)


class TestMovingMiniMaxUpdateScalar(unittest.TestCase):
    def test_update_scalar_outputs(self):
        ind = MovingMiniMax(MovingMiniMaxParams(m=5, n=50, num_extrema=3))
        tm = datetime(2021, 4, 1)
        out = None
        for p in td.INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=p))
        self.assertEqual(len(out), 6)
        self.assertEqual(len(out[4].points), 50)
        self.assertEqual(len(out[5].points), 50)


class TestMovingMiniMaxInvalidParams(unittest.TestCase):
    def test_m_too_small(self):
        with self.assertRaises(ValueError):
            MovingMiniMax(MovingMiniMaxParams(m=0))

    def test_n_too_small(self):
        with self.assertRaises(ValueError):
            MovingMiniMax(MovingMiniMaxParams(m=5, n=10))

    def test_num_extrema_too_small(self):
        with self.assertRaises(ValueError):
            MovingMiniMax(MovingMiniMaxParams(num_extrema=0))


if __name__ == '__main__':
    unittest.main()
