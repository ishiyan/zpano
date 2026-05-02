import math
import unittest

from .jurik_directional_movement_index import JurikDirectionalMovementIndex
from .params import JurikDirectionalMovementIndexParams
from .test_testdata import EXPECTED_DMX, INPUT_CLOSE, INPUT_HIGH, INPUT_LOW


class TestJurikDirectionalMovementIndex(unittest.TestCase):
    def test_dmx_values(self):
        for config in EXPECTED_DMX:
            length = config["length"]
            expected_bipolar = config["expected_bipolar"]
            expected_plus = config["expected_plus"]
            expected_minus = config["expected_minus"]

            with self.subTest(length=length):
                params = JurikDirectionalMovementIndexParams(length=length)
                ind = JurikDirectionalMovementIndex(params)

                for i in range(len(INPUT_CLOSE)):
                    result = ind.update(INPUT_HIGH[i], INPUT_LOW[i], INPUT_CLOSE[i])
                    bipolar, plus, minus = result

                    if i < 41:
                        continue

                    self.assertFalse(
                        math.isnan(expected_bipolar[i]),
                        f"length={length}, bar={i}: expected bipolar should not be NaN after warmup",
                    )
                    self.assertFalse(
                        math.isnan(bipolar),
                        f"length={length}, bar={i}: actual bipolar should not be NaN after warmup",
                    )
                    self.assertAlmostEqual(
                        bipolar,
                        expected_bipolar[i],
                        places=10,
                        msg=f"length={length}, bar={i}: bipolar mismatch",
                    )

                    self.assertFalse(
                        math.isnan(expected_plus[i]),
                        f"length={length}, bar={i}: expected plus should not be NaN after warmup",
                    )
                    self.assertFalse(
                        math.isnan(plus),
                        f"length={length}, bar={i}: actual plus should not be NaN after warmup",
                    )
                    self.assertAlmostEqual(
                        plus,
                        expected_plus[i],
                        places=10,
                        msg=f"length={length}, bar={i}: plus mismatch",
                    )

                    if len(expected_minus) > 0:
                        self.assertFalse(
                            math.isnan(expected_minus[i]),
                            f"length={length}, bar={i}: expected minus should not be NaN after warmup",
                        )
                        self.assertFalse(
                            math.isnan(minus),
                            f"length={length}, bar={i}: actual minus should not be NaN after warmup",
                        )
                        self.assertAlmostEqual(
                            minus,
                            expected_minus[i],
                            places=10,
                            msg=f"length={length}, bar={i}: minus mismatch",
                        )


if __name__ == "__main__":
    unittest.main()
