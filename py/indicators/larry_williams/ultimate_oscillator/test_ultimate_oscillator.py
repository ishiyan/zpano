import math
import unittest

from py.indicators.larry_williams.ultimate_oscillator.ultimate_oscillator import UltimateOscillator
from py.indicators.larry_williams.ultimate_oscillator.params import UltimateOscillatorParams
from py.entities.scalar import Scalar
from py.indicators.core.identifier import Identifier

from .test_testdata import (
    TEST_INPUT_HIGH,
    TEST_INPUT_LOW,
    TEST_INPUT_CLOSE,
    TEST_EXPECTED,
)


class TestUltimateOscillator(unittest.TestCase):

    def test_default_params_full_data(self):
        params = UltimateOscillatorParams()
        ind = UltimateOscillator(params)

        for i in range(len(TEST_INPUT_HIGH)):
            result = ind.update(TEST_INPUT_CLOSE[i], TEST_INPUT_HIGH[i], TEST_INPUT_LOW[i])

            if math.isnan(TEST_EXPECTED[i]):
                self.assertTrue(math.isnan(result),
                                f"index {i}: expected NaN, got {result}")
            else:
                self.assertAlmostEqual(result, TEST_EXPECTED[i], delta=1e-4,
                                       msg=f"index {i}: expected {TEST_EXPECTED[i]}, got {result}")

    def test_is_primed(self):
        params = UltimateOscillatorParams()
        ind = UltimateOscillator(params)

        for i in range(28):
            ind.update(TEST_INPUT_CLOSE[i], TEST_INPUT_HIGH[i], TEST_INPUT_LOW[i])
            self.assertFalse(ind.is_primed(), f"expected not primed at index {i}")

        ind.update(TEST_INPUT_CLOSE[28], TEST_INPUT_HIGH[28], TEST_INPUT_LOW[28])
        self.assertTrue(ind.is_primed(), "expected primed at index 28")

    def test_nan_input(self):
        params = UltimateOscillatorParams()
        ind = UltimateOscillator(params)

        result = ind.update(math.nan, 1.0, 1.0)
        self.assertTrue(math.isnan(result), f"expected NaN for NaN close, got {result}")

        result = ind.update(1.0, math.nan, 1.0)
        self.assertTrue(math.isnan(result), f"expected NaN for NaN high, got {result}")

        result = ind.update(1.0, 1.0, math.nan)
        self.assertTrue(math.isnan(result), f"expected NaN for NaN low, got {result}")

    def test_metadata(self):
        params = UltimateOscillatorParams()
        ind = UltimateOscillator(params)

        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.ULTIMATE_OSCILLATOR)
        self.assertEqual(meta.mnemonic, "ultosc(7, 14, 28)")

    def test_invalid_params(self):
        with self.assertRaises(ValueError):
            params = UltimateOscillatorParams(length1=1)
            UltimateOscillator(params)


if __name__ == '__main__':
    unittest.main()
