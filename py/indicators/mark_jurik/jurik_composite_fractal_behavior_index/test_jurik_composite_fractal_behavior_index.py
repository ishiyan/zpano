import math
import unittest

from .jurik_composite_fractal_behavior_index import JurikCompositeFractalBehaviorIndex
from .params import JurikCompositeFractalBehaviorIndexParams
from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_TYPE_1_SMOOTH_2, EXPECTED_TYPE_1_SMOOTH_10, EXPECTED_TYPE_1_SMOOTH_50,
    EXPECTED_TYPE_2_SMOOTH_2, EXPECTED_TYPE_2_SMOOTH_10, EXPECTED_TYPE_2_SMOOTH_50,
    EXPECTED_TYPE_3_SMOOTH_2, EXPECTED_TYPE_3_SMOOTH_10, EXPECTED_TYPE_3_SMOOTH_50,
    EXPECTED_TYPE_4_SMOOTH_2, EXPECTED_TYPE_4_SMOOTH_10, EXPECTED_TYPE_4_SMOOTH_50,
)

EPSILON = 1e-13

TEST_CASES = [
    (1, 2, EXPECTED_TYPE_1_SMOOTH_2),
    (1, 10, EXPECTED_TYPE_1_SMOOTH_10),
    (1, 50, EXPECTED_TYPE_1_SMOOTH_50),
    (2, 2, EXPECTED_TYPE_2_SMOOTH_2),
    (2, 10, EXPECTED_TYPE_2_SMOOTH_10),
    (2, 50, EXPECTED_TYPE_2_SMOOTH_50),
    (3, 2, EXPECTED_TYPE_3_SMOOTH_2),
    (3, 10, EXPECTED_TYPE_3_SMOOTH_10),
    (3, 50, EXPECTED_TYPE_3_SMOOTH_50),
    (4, 2, EXPECTED_TYPE_4_SMOOTH_2),
    (4, 10, EXPECTED_TYPE_4_SMOOTH_10),
    (4, 50, EXPECTED_TYPE_4_SMOOTH_50),
]


class TestJurikCompositeFractalBehaviorIndex(unittest.TestCase):
    def test_cfb_values(self):
        for fractal_type, smooth, expected in TEST_CASES:
            with self.subTest(fractal_type=fractal_type, smooth=smooth):
                params = JurikCompositeFractalBehaviorIndexParams(
                    fractal_type=fractal_type, smooth=smooth
                )
                ind = JurikCompositeFractalBehaviorIndex(params)

                for i in range(len(INPUT_CLOSE)):
                    result = ind.update(INPUT_CLOSE[i])

                    # Skip last bar: reference aux loop stops at len-2.
                    if i == len(INPUT_CLOSE) - 1:
                        continue

                    if math.isnan(expected[i]) and math.isnan(result):
                        continue

                    self.assertFalse(
                        math.isnan(expected[i]),
                        f"type={fractal_type}, smooth={smooth}, bar={i}: expected should not be NaN",
                    )
                    self.assertFalse(
                        math.isnan(result),
                        f"type={fractal_type}, smooth={smooth}, bar={i}: result should not be NaN",
                    )
                    self.assertAlmostEqual(
                        result,
                        expected[i],
                        places=13,
                        msg=f"type={fractal_type}, smooth={smooth}, bar={i}: mismatch",
                    )


if __name__ == "__main__":
    unittest.main()
