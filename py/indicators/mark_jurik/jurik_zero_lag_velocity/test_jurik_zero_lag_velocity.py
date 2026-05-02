import math
import unittest

from .jurik_zero_lag_velocity import JurikZeroLagVelocity
from .params import JurikZeroLagVelocityParams
from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_DEPTH_2, EXPECTED_DEPTH_3, EXPECTED_DEPTH_4,
    EXPECTED_DEPTH_5, EXPECTED_DEPTH_6, EXPECTED_DEPTH_7,
    EXPECTED_DEPTH_8, EXPECTED_DEPTH_9, EXPECTED_DEPTH_10,
    EXPECTED_DEPTH_11, EXPECTED_DEPTH_12, EXPECTED_DEPTH_13,
    EXPECTED_DEPTH_14, EXPECTED_DEPTH_15,
)

EPSILON = 1e-13

TEST_CASES = [
    (2, EXPECTED_DEPTH_2),
    (3, EXPECTED_DEPTH_3),
    (4, EXPECTED_DEPTH_4),
    (5, EXPECTED_DEPTH_5),
    (6, EXPECTED_DEPTH_6),
    (7, EXPECTED_DEPTH_7),
    (8, EXPECTED_DEPTH_8),
    (9, EXPECTED_DEPTH_9),
    (10, EXPECTED_DEPTH_10),
    (11, EXPECTED_DEPTH_11),
    (12, EXPECTED_DEPTH_12),
    (13, EXPECTED_DEPTH_13),
    (14, EXPECTED_DEPTH_14),
    (15, EXPECTED_DEPTH_15),
]


class TestJurikZeroLagVelocity(unittest.TestCase):
    def test_vel_values(self):
        for depth, expected in TEST_CASES:
            with self.subTest(depth=depth):
                params = JurikZeroLagVelocityParams(depth=depth)
                ind = JurikZeroLagVelocity(params)

                for i in range(len(INPUT_CLOSE)):
                    result = ind.update(INPUT_CLOSE[i])

                    if math.isnan(expected[i]) and math.isnan(result):
                        continue

                    self.assertFalse(
                        math.isnan(expected[i]),
                        f"depth={depth}, bar={i}: expected should not be NaN",
                    )
                    self.assertFalse(
                        math.isnan(result),
                        f"depth={depth}, bar={i}: result should not be NaN",
                    )
                    self.assertAlmostEqual(
                        result,
                        expected[i],
                        places=13,
                        msg=f"depth={depth}, bar={i}: mismatch",
                    )


if __name__ == "__main__":
    unittest.main()
