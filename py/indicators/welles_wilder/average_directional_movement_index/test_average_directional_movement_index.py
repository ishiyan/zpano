"""Tests for Welles Wilder's Average Directional Movement Index indicator."""

import math
import unittest

from .average_directional_movement_index import AverageDirectionalMovementIndex
from .params import AverageDirectionalMovementIndexParams
from .output import AverageDirectionalMovementIndexOutput
from ...core.identifier import Identifier
from ....entities.scalar import Scalar

from .test_testdata import (
    _HIGH,
    _LOW,
    _CLOSE,
    _EXPECTED_ADX14,
)


# fmt: off
# fmt: on


class TestAverageDirectionalMovementIndex(unittest.TestCase):
    """Tests for the Average Directional Movement Index indicator."""

    def test_constructor_valid(self):
        adx = AverageDirectionalMovementIndex(AverageDirectionalMovementIndexParams(length=14))
        self.assertFalse(adx.is_primed())

    def test_constructor_invalid_length(self):
        with self.assertRaises(ValueError):
            AverageDirectionalMovementIndex(AverageDirectionalMovementIndexParams(length=0))
        with self.assertRaises(ValueError):
            AverageDirectionalMovementIndex(AverageDirectionalMovementIndexParams(length=-8))

    def test_is_primed(self):
        adx = AverageDirectionalMovementIndex(AverageDirectionalMovementIndexParams(length=14))

        # ADX primes at index 27 (DX primes at 14, then need 14 DX values for SMA).
        for i in range(27):
            adx.update(_CLOSE[i], _HIGH[i], _LOW[i])
            self.assertFalse(adx.is_primed(), f"[{i}] should not be primed yet")

        adx.update(_CLOSE[27], _HIGH[27], _LOW[27])
        self.assertTrue(adx.is_primed(), "[27] should be primed")

    def test_update_accuracy(self):
        adx = AverageDirectionalMovementIndex(AverageDirectionalMovementIndexParams(length=14))

        for i in range(len(_HIGH)):
            act = adx.update(_CLOSE[i], _HIGH[i], _LOW[i])
            exp = _EXPECTED_ADX14[i]

            if math.isnan(exp):
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
                continue

            self.assertFalse(math.isnan(act), f"[{i}] expected {exp}, got NaN")
            self.assertAlmostEqual(act, exp, delta=1e-8, msg=f"[{i}]")

    def test_nan_passthrough(self):
        adx = AverageDirectionalMovementIndex(AverageDirectionalMovementIndexParams(length=14))

        self.assertTrue(math.isnan(adx.update(math.nan, 1, 1)), "NaN close")
        self.assertTrue(math.isnan(adx.update(1, math.nan, 1)), "NaN high")
        self.assertTrue(math.isnan(adx.update(1, 1, math.nan)), "NaN low")
        self.assertTrue(math.isnan(adx.update_sample(math.nan)), "NaN sample")

    def test_metadata(self):
        adx = AverageDirectionalMovementIndex(AverageDirectionalMovementIndexParams(length=14))
        meta = adx.metadata()

        self.assertEqual(Identifier.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX, meta.identifier)
        self.assertEqual("adx", meta.mnemonic)
        self.assertEqual("Average Directional Movement Index", meta.description)
        self.assertEqual(8, len(meta.outputs))
        self.assertEqual(int(AverageDirectionalMovementIndexOutput.VALUE), meta.outputs[0].kind)
        self.assertEqual("adx", meta.outputs[0].mnemonic)
        self.assertEqual("Average Directional Movement Index", meta.outputs[0].description)
        self.assertEqual(int(AverageDirectionalMovementIndexOutput.DIRECTIONAL_MOVEMENT_INDEX), meta.outputs[1].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexOutput.DIRECTIONAL_INDICATOR_PLUS), meta.outputs[2].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexOutput.DIRECTIONAL_INDICATOR_MINUS), meta.outputs[3].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexOutput.DIRECTIONAL_MOVEMENT_PLUS), meta.outputs[4].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexOutput.DIRECTIONAL_MOVEMENT_MINUS), meta.outputs[5].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexOutput.AVERAGE_TRUE_RANGE), meta.outputs[6].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexOutput.TRUE_RANGE), meta.outputs[7].kind)

    def test_update_scalar(self):
        adx = AverageDirectionalMovementIndex(AverageDirectionalMovementIndexParams(length=14))

        for i in range(27):
            adx.update(_CLOSE[i], _HIGH[i], _LOW[i])

        s = Scalar(time=0, value=_HIGH[27])
        result = adx.update_scalar(s)

        self.assertEqual(1, len(result))
        self.assertIsInstance(result[0], Scalar)
        self.assertEqual(0, result[0].time)


if __name__ == "__main__":
    unittest.main()
