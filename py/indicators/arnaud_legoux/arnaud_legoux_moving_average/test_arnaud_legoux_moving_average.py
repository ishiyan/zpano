import math
import unittest
from datetime import datetime

from py.indicators.arnaud_legoux.arnaud_legoux_moving_average.arnaud_legoux_moving_average import ArnaudLegouxMovingAverage
from py.indicators.arnaud_legoux.arnaud_legoux_moving_average.params import ArnaudLegouxMovingAverageParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_W9_S6_O0_85,
    EXPECTED_W9_S6_O0_5,
    EXPECTED_W10_S6_O0_85,
    EXPECTED_W5_S6_O0_9,
    EXPECTED_W1_S6_O0_85,
    EXPECTED_W3_S6_O0_85,
    EXPECTED_W21_S6_O0_85,
    EXPECTED_W50_S6_O0_85,
    EXPECTED_W9_S6_O0,
    EXPECTED_W9_S6_O1,
    EXPECTED_W9_S2_O0_85,
    EXPECTED_W9_S20_O0_85,
    EXPECTED_W9_S0_5_O0_85,
    EXPECTED_W15_S4_O0_7,
)


class TestArnaudLegouxMovingAverage(unittest.TestCase):

    def _run_test(self, params, expected):
        alma = ArnaudLegouxMovingAverage(params)
        for i, val in enumerate(INPUT_CLOSE):
            result = alma.update(val)
            exp = expected[i]
            if math.isnan(exp):
                self.assertTrue(math.isnan(result), f"expected NaN at index {i}, got {result}")
            else:
                self.assertAlmostEqual(result, exp, places=13,
                                       msg=f"mismatch at index {i}")

    def test_default_w9_s6_o0_85(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=9, sigma=6.0, offset=0.85), EXPECTED_W9_S6_O0_85)

    def test_w9_s6_o0_5(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=9, sigma=6.0, offset=0.5), EXPECTED_W9_S6_O0_5)

    def test_w10_s6_o0_85(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=10, sigma=6.0, offset=0.85), EXPECTED_W10_S6_O0_85)

    def test_w5_s6_o0_9(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=5, sigma=6.0, offset=0.9), EXPECTED_W5_S6_O0_9)

    def test_w1_s6_o0_85(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=1, sigma=6.0, offset=0.85), EXPECTED_W1_S6_O0_85)

    def test_w3_s6_o0_85(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=3, sigma=6.0, offset=0.85), EXPECTED_W3_S6_O0_85)

    def test_w21_s6_o0_85(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=21, sigma=6.0, offset=0.85), EXPECTED_W21_S6_O0_85)

    def test_w50_s6_o0_85(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=50, sigma=6.0, offset=0.85), EXPECTED_W50_S6_O0_85)

    def test_w9_s6_o0(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=9, sigma=6.0, offset=0.0), EXPECTED_W9_S6_O0)

    def test_w9_s6_o1(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=9, sigma=6.0, offset=1.0), EXPECTED_W9_S6_O1)

    def test_w9_s2_o0_85(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=9, sigma=2.0, offset=0.85), EXPECTED_W9_S2_O0_85)

    def test_w9_s20_o0_85(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=9, sigma=20.0, offset=0.85), EXPECTED_W9_S20_O0_85)

    def test_w9_s0_5_o0_85(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=9, sigma=0.5, offset=0.85), EXPECTED_W9_S0_5_O0_85)

    def test_w15_s4_o0_7(self):
        self._run_test(ArnaudLegouxMovingAverageParams(window=15, sigma=4.0, offset=0.7), EXPECTED_W15_S4_O0_7)

    def test_is_primed(self):
        alma = ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(window=9, sigma=6.0, offset=0.85))
        for i in range(8):
            alma.update(INPUT_CLOSE[i])
            self.assertFalse(alma.is_primed())
        alma.update(INPUT_CLOSE[8])
        self.assertTrue(alma.is_primed())

    def test_is_primed_window_1(self):
        alma = ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(window=1, sigma=6.0, offset=0.85))
        self.assertFalse(alma.is_primed())
        alma.update(INPUT_CLOSE[0])
        self.assertTrue(alma.is_primed())

    def test_nan_passthrough(self):
        alma = ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(window=9, sigma=6.0, offset=0.85))
        result = alma.update(math.nan)
        self.assertTrue(math.isnan(result))

    def test_metadata(self):
        alma = ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(window=9, sigma=6.0, offset=0.85))
        meta = alma.metadata()
        self.assertEqual(meta.identifier, Identifier.ARNAUD_LEGOUX_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "alma(9, 6.0, 0.85)")
        self.assertEqual(meta.description, "Arnaud Legoux moving average alma(9, 6.0, 0.85)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "alma(9, 6.0, 0.85)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError) as ctx:
            ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(window=0))
        self.assertIn("window should be greater than 0", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(window=-1))
        self.assertIn("window should be greater than 0", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(sigma=0.0))
        self.assertIn("sigma should be greater than 0", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(offset=-0.1))
        self.assertIn("offset should be between 0 and 1", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(offset=1.1))
        self.assertIn("offset should be between 0 and 1", str(ctx.exception))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)

        # update_bar
        alma = ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(window=1))
        output = alma.update_bar(Bar(t, 3.0, 3.0, 3.0, 5.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 5.0, places=13)

        # update_scalar
        alma = ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(window=1))
        output = alma.update_scalar(Scalar(t, 7.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertAlmostEqual(output[0].value, 7.0, places=13)

        # update_quote
        alma = ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(window=1))
        output = alma.update_quote(Quote(t, 3.0, 5.0, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertAlmostEqual(output[0].value, 4.0, places=13)

        # update_trade
        alma = ArnaudLegouxMovingAverage(ArnaudLegouxMovingAverageParams(window=1))
        output = alma.update_trade(Trade(t, 9.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertAlmostEqual(output[0].value, 9.0, places=13)


if __name__ == "__main__":
    unittest.main()
