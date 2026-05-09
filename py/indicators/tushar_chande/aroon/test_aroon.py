import math
import unittest

from py.indicators.tushar_chande.aroon.aroon import Aroon
from py.indicators.tushar_chande.aroon.params import AroonParams
from py.entities.scalar import Scalar
from py.indicators.core.identifier import Identifier

from .test_testdata import INPUT_HIGH, INPUT_LOW, EXPECTED


class TestAroon(unittest.TestCase):

    def test_length14_full_data(self):
        ind = Aroon(AroonParams(length=14))
        for i in range(252):
            up, down, osc = ind.update(INPUT_HIGH[i], INPUT_LOW[i])
            exp_up, exp_down, exp_osc = EXPECTED[i]
            if math.isnan(exp_up):
                self.assertTrue(math.isnan(up), f"[{i}] Up: expected NaN, got {up}")
                self.assertTrue(math.isnan(down), f"[{i}] Down: expected NaN, got {down}")
                self.assertTrue(math.isnan(osc), f"[{i}] Osc: expected NaN, got {osc}")
            else:
                self.assertAlmostEqual(up, exp_up, delta=1e-6,
                                       msg=f"[{i}] Up: expected {exp_up}, got {up}")
                self.assertAlmostEqual(down, exp_down, delta=1e-6,
                                       msg=f"[{i}] Down: expected {exp_down}, got {down}")
                self.assertAlmostEqual(osc, exp_osc, delta=1e-6,
                                       msg=f"[{i}] Osc: expected {exp_osc}, got {osc}")

    def test_is_primed(self):
        ind = Aroon(AroonParams(length=14))
        self.assertFalse(ind.is_primed())
        for i in range(14):
            ind.update(INPUT_HIGH[i], INPUT_LOW[i])
            self.assertFalse(ind.is_primed(), f"[{i}] expected not primed")
        ind.update(INPUT_HIGH[14], INPUT_LOW[14])
        self.assertTrue(ind.is_primed())

    def test_nan_input(self):
        ind = Aroon(AroonParams(length=14))
        up, down, osc = ind.update(math.nan, 1.0)
        self.assertTrue(math.isnan(up))
        self.assertTrue(math.isnan(down))
        self.assertTrue(math.isnan(osc))

    def test_metadata(self):
        ind = Aroon(AroonParams(length=14))
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.AROON)
        self.assertEqual(meta.mnemonic, "aroon(14)")
        self.assertEqual(len(meta.outputs), 3)

    def test_invalid_params(self):
        with self.assertRaises(ValueError):
            Aroon(AroonParams(length=1))
        with self.assertRaises(ValueError):
            Aroon(AroonParams(length=0))
        with self.assertRaises(ValueError):
            Aroon(AroonParams(length=-1))


if __name__ == "__main__":
    unittest.main()
