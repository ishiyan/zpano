import math
import unittest
from datetime import datetime

from py.indicators.manfred_dürschner.new_moving_average.new_moving_average import NewMovingAverage
from py.indicators.manfred_dürschner.new_moving_average.params import NewMovingAverageParams, MAType
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_SEC_4_PRI_AUTO_LWMA,
    EXPECTED_SEC_8_PRI_AUTO_LWMA,
    EXPECTED_SEC_16_PRI_AUTO_LWMA,
    EXPECTED_PRI_16_SEC_8_LWMA,
    EXPECTED_PRI_32_SEC_8_LWMA,
    EXPECTED_PRI_64_SEC_8_LWMA,
    EXPECTED_PRI_8_SEC_4_LWMA,
    EXPECTED_PRI_16_SEC_4_LWMA,
    EXPECTED_PRI_32_SEC_4_LWMA,
    EXPECTED_SEC_8_SMA,
    EXPECTED_SEC_8_EMA,
    EXPECTED_SEC_8_SMMA,
)


class TestNewMovingAverage(unittest.TestCase):

    def _run_test(self, params, expected):
        nma = NewMovingAverage(params)
        for i, val in enumerate(INPUT_CLOSE):
            result = nma.update(val)
            exp = expected[i]
            if math.isnan(exp):
                self.assertTrue(math.isnan(result), f"expected NaN at index {i}, got {result}")
            else:
                self.assertAlmostEqual(result, exp, places=13,
                                       msg=f"mismatch at index {i}")

    def test_sec_4_pri_auto_lwma(self):
        self._run_test(
            NewMovingAverageParams(primary_period=0, secondary_period=4, ma_type=MAType.LWMA),
            EXPECTED_SEC_4_PRI_AUTO_LWMA)

    def test_sec_8_pri_auto_lwma(self):
        self._run_test(
            NewMovingAverageParams(primary_period=0, secondary_period=8, ma_type=MAType.LWMA),
            EXPECTED_SEC_8_PRI_AUTO_LWMA)

    def test_sec_16_pri_auto_lwma(self):
        self._run_test(
            NewMovingAverageParams(primary_period=0, secondary_period=16, ma_type=MAType.LWMA),
            EXPECTED_SEC_16_PRI_AUTO_LWMA)

    def test_pri_16_sec_8_lwma(self):
        self._run_test(
            NewMovingAverageParams(primary_period=16, secondary_period=8, ma_type=MAType.LWMA),
            EXPECTED_PRI_16_SEC_8_LWMA)

    def test_pri_32_sec_8_lwma(self):
        self._run_test(
            NewMovingAverageParams(primary_period=32, secondary_period=8, ma_type=MAType.LWMA),
            EXPECTED_PRI_32_SEC_8_LWMA)

    def test_pri_64_sec_8_lwma(self):
        self._run_test(
            NewMovingAverageParams(primary_period=64, secondary_period=8, ma_type=MAType.LWMA),
            EXPECTED_PRI_64_SEC_8_LWMA)

    def test_pri_8_sec_4_lwma(self):
        self._run_test(
            NewMovingAverageParams(primary_period=8, secondary_period=4, ma_type=MAType.LWMA),
            EXPECTED_PRI_8_SEC_4_LWMA)

    def test_pri_16_sec_4_lwma(self):
        self._run_test(
            NewMovingAverageParams(primary_period=16, secondary_period=4, ma_type=MAType.LWMA),
            EXPECTED_PRI_16_SEC_4_LWMA)

    def test_pri_32_sec_4_lwma(self):
        self._run_test(
            NewMovingAverageParams(primary_period=32, secondary_period=4, ma_type=MAType.LWMA),
            EXPECTED_PRI_32_SEC_4_LWMA)

    def test_sec_8_sma(self):
        self._run_test(
            NewMovingAverageParams(primary_period=0, secondary_period=8, ma_type=MAType.SMA),
            EXPECTED_SEC_8_SMA)

    def test_sec_8_ema(self):
        self._run_test(
            NewMovingAverageParams(primary_period=0, secondary_period=8, ma_type=MAType.EMA),
            EXPECTED_SEC_8_EMA)

    def test_sec_8_smma(self):
        self._run_test(
            NewMovingAverageParams(primary_period=0, secondary_period=8, ma_type=MAType.SMMA),
            EXPECTED_SEC_8_SMMA)

    def test_is_primed(self):
        # Default: pri=0, sec=8 -> pri becomes 32, warmup = 32 + 8 - 2 = 38 NaNs
        nma = NewMovingAverage(NewMovingAverageParams())
        for i in range(38):
            nma.update(INPUT_CLOSE[i])
            self.assertFalse(nma.is_primed(), f"should not be primed at index {i}")
        nma.update(INPUT_CLOSE[38])
        self.assertTrue(nma.is_primed())

    def test_nan_passthrough(self):
        nma = NewMovingAverage(NewMovingAverageParams())
        result = nma.update(math.nan)
        self.assertTrue(math.isnan(result))

    def test_metadata(self):
        nma = NewMovingAverage(NewMovingAverageParams())
        meta = nma.metadata()
        self.assertEqual(meta.identifier, Identifier.NEW_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "nma(32, 8, 3)")
        self.assertEqual(meta.description, "New moving average nma(32, 8, 3)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "nma(32, 8, 3)")

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        params = NewMovingAverageParams(primary_period=4, secondary_period=2, ma_type=MAType.SMA)

        # update_bar - feed enough to prime (4 + 2 - 2 = 4 warmup, primes on 5th)
        nma = NewMovingAverage(params)
        for i in range(4):
            nma.update_bar(Bar(t, INPUT_CLOSE[i], INPUT_CLOSE[i], INPUT_CLOSE[i], INPUT_CLOSE[i], 1.0))
        output = nma.update_bar(Bar(t, INPUT_CLOSE[4], INPUT_CLOSE[4], INPUT_CLOSE[4], INPUT_CLOSE[4], 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertFalse(math.isnan(output[0].value))

        # update_scalar
        nma = NewMovingAverage(params)
        for i in range(4):
            nma.update_scalar(Scalar(t, INPUT_CLOSE[i]))
        output = nma.update_scalar(Scalar(t, INPUT_CLOSE[4]))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)

        # update_quote
        nma = NewMovingAverage(params)
        for i in range(4):
            nma.update_quote(Quote(t, INPUT_CLOSE[i], INPUT_CLOSE[i], 1.0, 1.0))
        output = nma.update_quote(Quote(t, INPUT_CLOSE[4], INPUT_CLOSE[4], 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)

        # update_trade
        nma = NewMovingAverage(params)
        for i in range(4):
            nma.update_trade(Trade(t, INPUT_CLOSE[i], 1.0))
        output = nma.update_trade(Trade(t, INPUT_CLOSE[4], 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)


if __name__ == "__main__":
    unittest.main()
