import math
import unittest

from .jurik_adaptive_zero_lag_velocity import JurikAdaptiveZeroLagVelocity
from .params import JurikAdaptiveZeroLagVelocityParams
from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_LO_2_HI_15, EXPECTED_LO_2_HI_30, EXPECTED_LO_2_HI_60,
    EXPECTED_LO_5_HI_15, EXPECTED_LO_5_HI_30, EXPECTED_LO_5_HI_60,
    EXPECTED_LO_10_HI_15, EXPECTED_LO_10_HI_30, EXPECTED_LO_10_HI_60,
    EXPECTED_SENS_0_5, EXPECTED_SENS_2_5, EXPECTED_SENS_5_0,
    EXPECTED_PERIOD_1_5, EXPECTED_PERIOD_10_0, EXPECTED_PERIOD_30_0,
)


TEST_CASES = [
    ("lo=2, hi=15", JurikAdaptiveZeroLagVelocityParams(lo_length=2, hi_length=15, sensitivity=1.0, period=3.0), EXPECTED_LO_2_HI_15),
    ("lo=2, hi=30", JurikAdaptiveZeroLagVelocityParams(lo_length=2, hi_length=30, sensitivity=1.0, period=3.0), EXPECTED_LO_2_HI_30),
    ("lo=2, hi=60", JurikAdaptiveZeroLagVelocityParams(lo_length=2, hi_length=60, sensitivity=1.0, period=3.0), EXPECTED_LO_2_HI_60),
    ("lo=5, hi=15", JurikAdaptiveZeroLagVelocityParams(lo_length=5, hi_length=15, sensitivity=1.0, period=3.0), EXPECTED_LO_5_HI_15),
    ("lo=5, hi=30", JurikAdaptiveZeroLagVelocityParams(lo_length=5, hi_length=30, sensitivity=1.0, period=3.0), EXPECTED_LO_5_HI_30),
    ("lo=5, hi=60", JurikAdaptiveZeroLagVelocityParams(lo_length=5, hi_length=60, sensitivity=1.0, period=3.0), EXPECTED_LO_5_HI_60),
    ("lo=10, hi=15", JurikAdaptiveZeroLagVelocityParams(lo_length=10, hi_length=15, sensitivity=1.0, period=3.0), EXPECTED_LO_10_HI_15),
    ("lo=10, hi=30", JurikAdaptiveZeroLagVelocityParams(lo_length=10, hi_length=30, sensitivity=1.0, period=3.0), EXPECTED_LO_10_HI_30),
    ("lo=10, hi=60", JurikAdaptiveZeroLagVelocityParams(lo_length=10, hi_length=60, sensitivity=1.0, period=3.0), EXPECTED_LO_10_HI_60),
    ("sens=0.5", JurikAdaptiveZeroLagVelocityParams(lo_length=5, hi_length=30, sensitivity=0.5, period=3.0), EXPECTED_SENS_0_5),
    ("sens=2.5", JurikAdaptiveZeroLagVelocityParams(lo_length=5, hi_length=30, sensitivity=2.5, period=3.0), EXPECTED_SENS_2_5),
    ("sens=5.0", JurikAdaptiveZeroLagVelocityParams(lo_length=5, hi_length=30, sensitivity=5.0, period=3.0), EXPECTED_SENS_5_0),
    ("period=1.5", JurikAdaptiveZeroLagVelocityParams(lo_length=5, hi_length=30, sensitivity=1.0, period=1.5), EXPECTED_PERIOD_1_5),
    ("period=10.0", JurikAdaptiveZeroLagVelocityParams(lo_length=5, hi_length=30, sensitivity=1.0, period=10.0), EXPECTED_PERIOD_10_0),
    ("period=30.0", JurikAdaptiveZeroLagVelocityParams(lo_length=5, hi_length=30, sensitivity=1.0, period=30.0), EXPECTED_PERIOD_30_0),
]


class TestJurikAdaptiveZeroLagVelocity(unittest.TestCase):
    def test_javel_values(self):
        for name, params, expected in TEST_CASES:
            with self.subTest(name=name):
                ind = JurikAdaptiveZeroLagVelocity(params)

                for i in range(len(INPUT_CLOSE)):
                    result = ind.update(INPUT_CLOSE[i])

                    if math.isnan(expected[i]):
                        self.assertTrue(
                            math.isnan(result),
                            f"{name}, bar={i}: expected NaN, got {result}",
                        )
                        continue

                    self.assertFalse(
                        math.isnan(result),
                        f"{name}, bar={i}: expected {expected[i]}, got NaN",
                    )
                    self.assertAlmostEqual(
                        result,
                        expected[i],
                        places=5,
                        msg=f"{name}, bar={i}: mismatch",
                    )


if __name__ == "__main__":
    unittest.main()
