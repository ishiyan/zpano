import math
import unittest

from .jurik_relative_trend_strength_index import JurikRelativeTrendStrengthIndex
from .params import JurikRelativeTrendStrengthIndexParams
from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_LENGTH_2, EXPECTED_LENGTH_3, EXPECTED_LENGTH_4,
    EXPECTED_LENGTH_5, EXPECTED_LENGTH_6, EXPECTED_LENGTH_7,
    EXPECTED_LENGTH_8, EXPECTED_LENGTH_9, EXPECTED_LENGTH_10,
    EXPECTED_LENGTH_11, EXPECTED_LENGTH_12, EXPECTED_LENGTH_13,
    EXPECTED_LENGTH_14, EXPECTED_LENGTH_15,
)

EPSILON = 1e-13

TEST_CASES = [
    (2, EXPECTED_LENGTH_2),
    (3, EXPECTED_LENGTH_3),
    (4, EXPECTED_LENGTH_4),
    (5, EXPECTED_LENGTH_5),
    (6, EXPECTED_LENGTH_6),
    (7, EXPECTED_LENGTH_7),
    (8, EXPECTED_LENGTH_8),
    (9, EXPECTED_LENGTH_9),
    (10, EXPECTED_LENGTH_10),
    (11, EXPECTED_LENGTH_11),
    (12, EXPECTED_LENGTH_12),
    (13, EXPECTED_LENGTH_13),
    (14, EXPECTED_LENGTH_14),
    (15, EXPECTED_LENGTH_15),
]


class TestJurikRelativeTrendStrengthIndex(unittest.TestCase):
    def test_rsx_values(self):
        for length, expected in TEST_CASES:
            with self.subTest(length=length):
                params = JurikRelativeTrendStrengthIndexParams(length=length)
                ind = JurikRelativeTrendStrengthIndex(params)

                for i in range(len(INPUT_CLOSE)):
                    result = ind.update(INPUT_CLOSE[i])

                    if math.isnan(expected[i]) and math.isnan(result):
                        continue

                    self.assertFalse(
                        math.isnan(expected[i]),
                        f"length={length}, bar={i}: expected should not be NaN",
                    )
                    self.assertFalse(
                        math.isnan(result),
                        f"length={length}, bar={i}: result should not be NaN",
                    )
                    self.assertAlmostEqual(
                        result,
                        expected[i],
                        places=13,
                        msg=f"length={length}, bar={i}: mismatch",
                    )


if __name__ == "__main__":
    unittest.main()
