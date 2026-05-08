import math
import unittest

from .jurik_turning_point_oscillator import JurikTurningPointOscillator
from .params import JurikTurningPointOscillatorParams
from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_LEN_5, EXPECTED_LEN_7, EXPECTED_LEN_10, EXPECTED_LEN_14,
    EXPECTED_LEN_20, EXPECTED_LEN_28, EXPECTED_LEN_40, EXPECTED_LEN_60,
    EXPECTED_LEN_80,
)

TEST_CASES = [
    (5, EXPECTED_LEN_5),
    (7, EXPECTED_LEN_7),
    (10, EXPECTED_LEN_10),
    (14, EXPECTED_LEN_14),
    (20, EXPECTED_LEN_20),
    (28, EXPECTED_LEN_28),
    (40, EXPECTED_LEN_40),
    (60, EXPECTED_LEN_60),
    (80, EXPECTED_LEN_80),
]


class TestJurikTurningPointOscillator(unittest.TestCase):
    def test_jtpo_values(self):
        for length, expected in TEST_CASES:
            with self.subTest(length=length):
                params = JurikTurningPointOscillatorParams(length=length)
                ind = JurikTurningPointOscillator(params)

                for i in range(len(INPUT_CLOSE)):
                    result = ind.update(INPUT_CLOSE[i])

                    if math.isnan(expected[i]):
                        self.assertTrue(
                            math.isnan(result),
                            f"length={length}, bar={i}: expected NaN, got {result}",
                        )
                        continue

                    self.assertFalse(
                        math.isnan(result),
                        f"length={length}, bar={i}: expected {expected[i]}, got NaN",
                    )
                    self.assertAlmostEqual(
                        result,
                        expected[i],
                        places=5,
                        msg=f"length={length}, bar={i}: mismatch",
                    )

    def test_invalid_length(self):
        with self.assertRaises(ValueError):
            JurikTurningPointOscillator(JurikTurningPointOscillatorParams(length=1))


if __name__ == "__main__":
    unittest.main()
