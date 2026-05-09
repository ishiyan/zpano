import math
import unittest

from py.indicators.welles_wilder.directional_movement_index.directional_movement_index import DirectionalMovementIndex
from py.indicators.welles_wilder.directional_movement_index.params import DirectionalMovementIndexParams
from py.indicators.welles_wilder.directional_movement_index.output import DirectionalMovementIndexOutput
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    TEST_INPUT_HIGH,
    TEST_INPUT_LOW,
    TEST_INPUT_CLOSE,
    TEST_EXPECTED_DX14,
)


# TA-Lib test data (252 entries), extracted programmatically from DirectionalMovementIndexTest.cs.
# Expected DX14 (length=14), 252 entries.
class TestDirectionalMovementIndex(unittest.TestCase):

    def test_constructor_valid(self):
        dx = DirectionalMovementIndex(DirectionalMovementIndexParams(14))
        self.assertFalse(dx.is_primed())

    def test_constructor_invalid_length(self):
        with self.assertRaises(ValueError):
            DirectionalMovementIndex(DirectionalMovementIndexParams(0))
        with self.assertRaises(ValueError):
            DirectionalMovementIndex(DirectionalMovementIndexParams(-8))

    def test_is_primed(self):
        dx = DirectionalMovementIndex(DirectionalMovementIndexParams(14))
        for i in range(14):
            dx.update(TEST_INPUT_CLOSE[i], TEST_INPUT_HIGH[i], TEST_INPUT_LOW[i])
            self.assertFalse(dx.is_primed(), f"[{i}] should not be primed yet")

        dx.update(TEST_INPUT_CLOSE[14], TEST_INPUT_HIGH[14], TEST_INPUT_LOW[14])
        self.assertTrue(dx.is_primed(), "[14] should be primed")

    def test_update(self):
        dx = DirectionalMovementIndex(DirectionalMovementIndexParams(14))
        for i in range(len(TEST_INPUT_HIGH)):
            act = dx.update(TEST_INPUT_CLOSE[i], TEST_INPUT_HIGH[i], TEST_INPUT_LOW[i])
            exp = TEST_EXPECTED_DX14[i]

            if math.isnan(exp):
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
                continue

            self.assertFalse(math.isnan(act), f"[{i}] expected {exp}, got NaN")
            self.assertAlmostEqual(act, exp, delta=1e-8, msg=f"[{i}]")

    def test_nan_passthrough(self):
        dx = DirectionalMovementIndex(DirectionalMovementIndexParams(14))
        self.assertTrue(math.isnan(dx.update(float('nan'), 1, 1)))
        self.assertTrue(math.isnan(dx.update(1, float('nan'), 1)))
        self.assertTrue(math.isnan(dx.update(1, 1, float('nan'))))

    def test_metadata(self):
        dx = DirectionalMovementIndex(DirectionalMovementIndexParams(14))
        meta = dx.metadata()
        self.assertEqual(Identifier.DIRECTIONAL_MOVEMENT_INDEX, meta.identifier)
        self.assertEqual("dx", meta.mnemonic)
        self.assertEqual(7, len(meta.outputs))

    def test_update_scalar(self):
        dx = DirectionalMovementIndex(DirectionalMovementIndexParams(14))
        for i in range(14):
            dx.update(TEST_INPUT_CLOSE[i], TEST_INPUT_HIGH[i], TEST_INPUT_LOW[i])

        s = Scalar(time=None, value=TEST_INPUT_HIGH[14])
        result = dx.update_scalar(s)
        self.assertEqual(1, len(result))


if __name__ == '__main__':
    unittest.main()
