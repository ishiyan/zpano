import math
import unittest
from datetime import datetime

from py.indicators.tushar_chande.chande_momentum_oscillator.chande_momentum_oscillator import ChandeMomentumOscillator
from py.indicators.tushar_chande.chande_momentum_oscillator.params import ChandeMomentumOscillatorParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade

from .test_testdata import BOOK_INPUT, BOOK_OUTPUT, INPUT

# Book data: Chande & Kroll, The New Technical Trader, p.96, Table 5.1
class TestChandeMomentumOscillator(unittest.TestCase):

    def test_book_length_10(self):
        """Book data: length=10, Chande & Kroll Table 5.1."""
        cmo = ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=10))
        for i in range(10):
            v = cmo.update(BOOK_INPUT[i])
            self.assertTrue(math.isnan(v), f"[{i}] expected NaN")
        for i in range(10, len(BOOK_INPUT)):
            v = cmo.update(BOOK_INPUT[i])
            self.assertAlmostEqual(v, BOOK_OUTPUT[i - 10], delta=1e-13)

    def test_nan_passthrough(self):
        cmo = ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=10))
        # Feed book data to prime
        for i in range(len(BOOK_INPUT)):
            cmo.update(BOOK_INPUT[i])
        # NaN passthrough after primed
        v = cmo.update(math.nan)
        self.assertTrue(math.isnan(v))

    def test_is_primed_length_1(self):
        cmo = ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=1))
        self.assertFalse(cmo.is_primed())
        cmo.update(INPUT[0])
        self.assertFalse(cmo.is_primed())
        cmo.update(INPUT[1])
        self.assertTrue(cmo.is_primed())

    def test_is_primed_length_2(self):
        cmo = ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=2))
        self.assertFalse(cmo.is_primed())
        for i in range(2):
            cmo.update(INPUT[i])
            self.assertFalse(cmo.is_primed())
        cmo.update(INPUT[2])
        self.assertTrue(cmo.is_primed())

    def test_is_primed_length_10(self):
        cmo = ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=10))
        self.assertFalse(cmo.is_primed())
        for i in range(10):
            cmo.update(INPUT[i])
            self.assertFalse(cmo.is_primed())
        cmo.update(INPUT[10])
        self.assertTrue(cmo.is_primed())

    def test_entity_scalar(self):
        t = datetime(2021, 4, 1)
        cmo = ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=2))
        cmo.update(0.0)
        cmo.update(0.0)
        out = cmo.update_scalar(Scalar(t, 3.0))
        self.assertEqual(len(out), 1)
        self.assertAlmostEqual(out[0].value, 100.0, delta=1e-13)

    def test_entity_bar(self):
        t = datetime(2021, 4, 1)
        cmo = ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=2))
        cmo.update(0.0)
        cmo.update(0.0)
        out = cmo.update_bar(Bar(t, 0, 0, 0, 3.0, 0))
        self.assertEqual(len(out), 1)
        self.assertAlmostEqual(out[0].value, 100.0, delta=1e-13)

    def test_entity_quote(self):
        t = datetime(2021, 4, 1)
        cmo = ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=2))
        cmo.update(0.0)
        cmo.update(0.0)
        out = cmo.update_quote(Quote(t, 3.0, 3.0, 0, 0))
        self.assertEqual(len(out), 1)
        self.assertAlmostEqual(out[0].value, 100.0, delta=1e-13)

    def test_entity_trade(self):
        t = datetime(2021, 4, 1)
        cmo = ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=2))
        cmo.update(0.0)
        cmo.update(0.0)
        out = cmo.update_trade(Trade(t, 3.0, 0))
        self.assertEqual(len(out), 1)
        self.assertAlmostEqual(out[0].value, 100.0, delta=1e-13)

    def test_metadata(self):
        cmo = ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=5))
        meta = cmo.metadata()
        self.assertEqual(meta.identifier, Identifier.CHANDE_MOMENTUM_OSCILLATOR)
        self.assertEqual(meta.mnemonic, "cmo(5)")
        self.assertEqual(meta.description, "Chande Momentum Oscillator cmo(5)")
        self.assertEqual(len(meta.outputs), 1)

    def test_invalid_params(self):
        with self.assertRaises(ValueError):
            ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=0))
        with self.assertRaises(ValueError):
            ChandeMomentumOscillator(ChandeMomentumOscillatorParams(length=-1))


if __name__ == "__main__":
    unittest.main()
