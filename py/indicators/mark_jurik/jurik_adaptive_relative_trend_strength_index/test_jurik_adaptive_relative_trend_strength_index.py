import math
import unittest

from .jurik_adaptive_relative_trend_strength_index import JurikAdaptiveRelativeTrendStrengthIndex
from .params import JurikAdaptiveRelativeTrendStrengthIndexParams
from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_LO_2_HI_15, EXPECTED_LO_2_HI_30, EXPECTED_LO_2_HI_60,
    EXPECTED_LO_5_HI_15, EXPECTED_LO_5_HI_30, EXPECTED_LO_5_HI_60,
    EXPECTED_LO_10_HI_15, EXPECTED_LO_10_HI_30, EXPECTED_LO_10_HI_60,
)

EPSILON = 1e-6

TEST_CASES = [
    (2, 15, EXPECTED_LO_2_HI_15),
    (2, 30, EXPECTED_LO_2_HI_30),
    (2, 60, EXPECTED_LO_2_HI_60),
    (5, 15, EXPECTED_LO_5_HI_15),
    (5, 30, EXPECTED_LO_5_HI_30),
    (5, 60, EXPECTED_LO_5_HI_60),
    (10, 15, EXPECTED_LO_10_HI_15),
    (10, 30, EXPECTED_LO_10_HI_30),
    (10, 60, EXPECTED_LO_10_HI_60),
]


class TestJurikAdaptiveRelativeTrendStrengthIndex(unittest.TestCase):
    def test_jarsx_values(self):
        for lo, hi, expected in TEST_CASES:
            with self.subTest(lo_length=lo, hi_length=hi):
                params = JurikAdaptiveRelativeTrendStrengthIndexParams(
                    lo_length=lo, hi_length=hi)
                ind = JurikAdaptiveRelativeTrendStrengthIndex(params)

                for i in range(len(INPUT_CLOSE)):
                    result = ind.update(INPUT_CLOSE[i])

                    if math.isnan(expected[i]):
                        self.assertTrue(
                            math.isnan(result),
                            f"lo={lo}, hi={hi}, bar={i}: expected NaN, got {result}",
                        )
                        continue

                    self.assertFalse(
                        math.isnan(result),
                        f"lo={lo}, hi={hi}, bar={i}: expected {expected[i]}, got NaN",
                    )
                    self.assertAlmostEqual(
                        result,
                        expected[i],
                        places=5,
                        msg=f"lo={lo}, hi={hi}, bar={i}: mismatch",
                    )


if __name__ == "__main__":
    unittest.main()
