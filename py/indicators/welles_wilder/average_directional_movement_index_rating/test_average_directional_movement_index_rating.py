"""Tests for Welles Wilder's Average Directional Movement Index Rating indicator."""

import math
import unittest

from .average_directional_movement_index_rating import AverageDirectionalMovementIndexRating
from .params import AverageDirectionalMovementIndexRatingParams
from .output import AverageDirectionalMovementIndexRatingOutput
from ...core.identifier import Identifier
from ....entities.scalar import Scalar

from .test_testdata import (
    _HIGH,
    _LOW,
    _CLOSE,
    _EXPECTED_ADXR14,
)


# fmt: off
# fmt: on


class TestAverageDirectionalMovementIndexRating(unittest.TestCase):
    """Tests for the Average Directional Movement Index Rating indicator."""

    def test_constructor_valid(self):
        adxr = AverageDirectionalMovementIndexRating(
            AverageDirectionalMovementIndexRatingParams(length=14)
        )
        self.assertFalse(adxr.is_primed())

    def test_constructor_invalid_length(self):
        with self.assertRaises(ValueError):
            AverageDirectionalMovementIndexRating(
                AverageDirectionalMovementIndexRatingParams(length=0)
            )
        with self.assertRaises(ValueError):
            AverageDirectionalMovementIndexRating(
                AverageDirectionalMovementIndexRatingParams(length=-8)
            )

    def test_is_primed(self):
        adxr = AverageDirectionalMovementIndexRating(
            AverageDirectionalMovementIndexRatingParams(length=14)
        )

        # ADX primes at index 27. ADXR needs (length-1)=13 more ADX values after that,
        # so ADXR primes at index 40.
        for i in range(40):
            adxr.update(_CLOSE[i], _HIGH[i], _LOW[i])
            self.assertFalse(adxr.is_primed(), f"[{i}] should not be primed yet")

        adxr.update(_CLOSE[40], _HIGH[40], _LOW[40])
        self.assertTrue(adxr.is_primed(), "[40] should be primed")

    def test_update_accuracy(self):
        adxr = AverageDirectionalMovementIndexRating(
            AverageDirectionalMovementIndexRatingParams(length=14)
        )

        for i in range(len(_HIGH)):
            act = adxr.update(_CLOSE[i], _HIGH[i], _LOW[i])
            exp = _EXPECTED_ADXR14[i]

            if math.isnan(exp):
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
                continue

            self.assertFalse(math.isnan(act), f"[{i}] expected {exp}, got NaN")
            self.assertAlmostEqual(act, exp, delta=1e-8, msg=f"[{i}]")

    def test_nan_passthrough(self):
        adxr = AverageDirectionalMovementIndexRating(
            AverageDirectionalMovementIndexRatingParams(length=14)
        )

        self.assertTrue(math.isnan(adxr.update(math.nan, 1, 1)), "NaN close")
        self.assertTrue(math.isnan(adxr.update(1, math.nan, 1)), "NaN high")
        self.assertTrue(math.isnan(adxr.update(1, 1, math.nan)), "NaN low")
        self.assertTrue(math.isnan(adxr.update_sample(math.nan)), "NaN sample")

    def test_metadata(self):
        adxr = AverageDirectionalMovementIndexRating(
            AverageDirectionalMovementIndexRatingParams(length=14)
        )
        meta = adxr.metadata()

        self.assertEqual(Identifier.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX_RATING, meta.identifier)
        self.assertEqual("adxr", meta.mnemonic)
        self.assertEqual("Average Directional Movement Index Rating", meta.description)
        self.assertEqual(9, len(meta.outputs))
        self.assertEqual(int(AverageDirectionalMovementIndexRatingOutput.VALUE), meta.outputs[0].kind)
        self.assertEqual("adxr", meta.outputs[0].mnemonic)
        self.assertEqual("Average Directional Movement Index Rating", meta.outputs[0].description)
        self.assertEqual(int(AverageDirectionalMovementIndexRatingOutput.AVERAGE_DIRECTIONAL_MOVEMENT_INDEX), meta.outputs[1].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexRatingOutput.DIRECTIONAL_MOVEMENT_INDEX), meta.outputs[2].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexRatingOutput.DIRECTIONAL_INDICATOR_PLUS), meta.outputs[3].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexRatingOutput.DIRECTIONAL_INDICATOR_MINUS), meta.outputs[4].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexRatingOutput.DIRECTIONAL_MOVEMENT_PLUS), meta.outputs[5].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexRatingOutput.DIRECTIONAL_MOVEMENT_MINUS), meta.outputs[6].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexRatingOutput.AVERAGE_TRUE_RANGE), meta.outputs[7].kind)
        self.assertEqual(int(AverageDirectionalMovementIndexRatingOutput.TRUE_RANGE), meta.outputs[8].kind)

    def test_update_scalar(self):
        adxr = AverageDirectionalMovementIndexRating(
            AverageDirectionalMovementIndexRatingParams(length=14)
        )

        for i in range(40):
            adxr.update(_CLOSE[i], _HIGH[i], _LOW[i])

        s = Scalar(time=0, value=_HIGH[40])
        result = adxr.update_scalar(s)

        self.assertEqual(1, len(result))
        self.assertIsInstance(result[0], Scalar)
        self.assertEqual(0, result[0].time)


if __name__ == "__main__":
    unittest.main()
