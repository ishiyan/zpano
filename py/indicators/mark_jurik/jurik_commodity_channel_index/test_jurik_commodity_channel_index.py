import math
import unittest

from .jurik_commodity_channel_index import JurikCommodityChannelIndex
from .params import JurikCommodityChannelIndexParams
from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_LEN_10, EXPECTED_LEN_14, EXPECTED_LEN_20,
    EXPECTED_LEN_30, EXPECTED_LEN_40, EXPECTED_LEN_50,
    EXPECTED_LEN_60, EXPECTED_LEN_80, EXPECTED_LEN_100,
)

TEST_CASES = [
    (10, EXPECTED_LEN_10),
    (14, EXPECTED_LEN_14),
    (20, EXPECTED_LEN_20),
    (30, EXPECTED_LEN_30),
    (40, EXPECTED_LEN_40),
    (50, EXPECTED_LEN_50),
    (60, EXPECTED_LEN_60),
    (80, EXPECTED_LEN_80),
    (100, EXPECTED_LEN_100),
]


class TestJurikCommodityChannelIndex(unittest.TestCase):
    def test_jcci_values(self):
        for length, expected in TEST_CASES:
            with self.subTest(length=length):
                params = JurikCommodityChannelIndexParams(length=length)
                ind = JurikCommodityChannelIndex(params)

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


if __name__ == "__main__":
    unittest.main()
