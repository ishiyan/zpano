"""Tests for Kaufman Adaptive Moving Average indicator."""

import math
import unittest
from datetime import datetime

from .kaufman_adaptive_moving_average import KaufmanAdaptiveMovingAverage
from .params import (KaufmanAdaptiveMovingAverageLengthParams,
                     KaufmanAdaptiveMovingAverageSmoothingFactorParams)
from .output import KaufmanAdaptiveMovingAverageOutput
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent
from ...core.identifier import Identifier

from .test_testdata import _INPUT, _EXPECTED


# Data taken from TA-Lib test_KAMA.xls, Close, C5..C256, 252 entries.
# Expected KAMA values from TA-Lib test_KAMA.xls, J5..J256.
class TestKaufmanAdaptiveMovingAverage(unittest.TestCase):
    """Tests for KAMA indicator."""

    def test_values_length(self):
        """Test KAMA(10, 2, 30) with 252 close prices from TA-Lib."""
        params = KaufmanAdaptiveMovingAverageLengthParams(
            efficiency_ratio_length=10, fastest_length=2, slowest_length=30)
        kama = KaufmanAdaptiveMovingAverage.from_length(params)

        for i in range(10):
            result = kama.update(_INPUT[i])
            self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")

        for i in range(10, len(_INPUT)):
            result = kama.update(_INPUT[i])
            self.assertAlmostEqual(_EXPECTED[i], result, delta=1e-8,
                                   msg=f"[{i}] expected {_EXPECTED[i]}, got {result}")

        # NaN passthrough
        self.assertTrue(math.isnan(kama.update(math.nan)))

    def test_is_primed(self):
        """Priming requires erLength=10 samples."""
        params = KaufmanAdaptiveMovingAverageLengthParams(
            efficiency_ratio_length=10, fastest_length=2, slowest_length=30)
        kama = KaufmanAdaptiveMovingAverage.from_length(params)

        self.assertFalse(kama.is_primed())

        for i in range(10):
            kama.update(_INPUT[i])
            self.assertFalse(kama.is_primed(), f"[{i+1}] should not be primed")

        for i in range(10, len(_INPUT)):
            kama.update(_INPUT[i])
            self.assertTrue(kama.is_primed(), f"[{i+1}] should be primed")

    def test_metadata_length(self):
        """Metadata for length-based construction."""
        params = KaufmanAdaptiveMovingAverageLengthParams(
            efficiency_ratio_length=10, fastest_length=2, slowest_length=30)
        kama = KaufmanAdaptiveMovingAverage.from_length(params)
        m = kama.metadata()

        self.assertEqual(Identifier.KAUFMAN_ADAPTIVE_MOVING_AVERAGE, m.identifier)
        self.assertEqual("kama(10, 2, 30)", m.mnemonic)
        self.assertEqual("Kaufman adaptive moving average kama(10, 2, 30)", m.description)
        self.assertEqual(1, len(m.outputs))
        self.assertEqual("kama(10, 2, 30)", m.outputs[0].mnemonic)

    def test_metadata_alpha(self):
        """Metadata for smoothing-factor-based construction."""
        params = KaufmanAdaptiveMovingAverageSmoothingFactorParams(
            efficiency_ratio_length=10,
            fastest_smoothing_factor=0.666666666,
            slowest_smoothing_factor=0.064516129)
        kama = KaufmanAdaptiveMovingAverage.from_smoothing_factor(params)
        m = kama.metadata()

        self.assertEqual("kama(10, 0.6667, 0.0645)", m.mnemonic)
        self.assertEqual("Kaufman adaptive moving average kama(10, 0.6667, 0.0645)", m.description)

    def test_metadata_non_default_bar_component(self):
        """Mnemonic includes non-default bar component."""
        params = KaufmanAdaptiveMovingAverageLengthParams(
            efficiency_ratio_length=10, fastest_length=2, slowest_length=30,
            bar_component=BarComponent.MEDIAN)
        kama = KaufmanAdaptiveMovingAverage.from_length(params)
        m = kama.metadata()

        self.assertEqual("kama(10, 2, 30, hl/2)", m.mnemonic)

    def test_metadata_non_default_quote_component(self):
        """Mnemonic includes non-default quote component via alpha constructor."""
        params = KaufmanAdaptiveMovingAverageSmoothingFactorParams(
            efficiency_ratio_length=10,
            fastest_smoothing_factor=2.0 / 3.0,
            slowest_smoothing_factor=2.0 / 31.0,
            quote_component=QuoteComponent.BID)
        kama = KaufmanAdaptiveMovingAverage.from_smoothing_factor(params)
        m = kama.metadata()

        self.assertEqual("kama(10, 0.6667, 0.0645, b)", m.mnemonic)

    def test_invalid_params_length(self):
        """Invalid length params raise ValueError."""
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_length(
                KaufmanAdaptiveMovingAverageLengthParams(
                    efficiency_ratio_length=1, fastest_length=2, slowest_length=30))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_length(
                KaufmanAdaptiveMovingAverageLengthParams(
                    efficiency_ratio_length=10, fastest_length=1, slowest_length=30))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_length(
                KaufmanAdaptiveMovingAverageLengthParams(
                    efficiency_ratio_length=10, fastest_length=2, slowest_length=1))

    def test_invalid_params_alpha(self):
        """Invalid alpha params raise ValueError."""
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_smoothing_factor(
                KaufmanAdaptiveMovingAverageSmoothingFactorParams(
                    efficiency_ratio_length=1))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_smoothing_factor(
                KaufmanAdaptiveMovingAverageSmoothingFactorParams(
                    fastest_smoothing_factor=-0.00000001))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_smoothing_factor(
                KaufmanAdaptiveMovingAverageSmoothingFactorParams(
                    fastest_smoothing_factor=1.00000001))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_smoothing_factor(
                KaufmanAdaptiveMovingAverageSmoothingFactorParams(
                    slowest_smoothing_factor=-0.00000001))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_smoothing_factor(
                KaufmanAdaptiveMovingAverageSmoothingFactorParams(
                    slowest_smoothing_factor=1.00000001))

    def test_update_entity(self):
        """Entity update methods return Scalar output."""
        params = KaufmanAdaptiveMovingAverageLengthParams(
            efficiency_ratio_length=10, fastest_length=2, slowest_length=30)
        kama = KaufmanAdaptiveMovingAverage.from_length(params)
        t = datetime(2021, 4, 1)

        # Prime with 10 zeros
        for _ in range(10):
            kama.update(0.0)

        inp = 3.0
        expected = 1.3333333333333328

        # Scalar
        kama2 = KaufmanAdaptiveMovingAverage.from_length(
            KaufmanAdaptiveMovingAverageLengthParams(
                efficiency_ratio_length=10, fastest_length=2, slowest_length=30))
        for _ in range(10):
            kama2.update(0.0)
        out = kama2.update_scalar(Scalar(time=t, value=inp))
        self.assertEqual(1, len(out))
        self.assertEqual(expected, out[0].value)

        # Bar
        kama3 = KaufmanAdaptiveMovingAverage.from_length(
            KaufmanAdaptiveMovingAverageLengthParams(
                efficiency_ratio_length=10, fastest_length=2, slowest_length=30))
        for _ in range(10):
            kama3.update(0.0)
        out = kama3.update_bar(Bar(time=t, open=0, high=0, low=0, close=inp, volume=0))
        self.assertEqual(1, len(out))
        self.assertEqual(expected, out[0].value)

        # Quote
        kama4 = KaufmanAdaptiveMovingAverage.from_length(
            KaufmanAdaptiveMovingAverageLengthParams(
                efficiency_ratio_length=10, fastest_length=2, slowest_length=30))
        for _ in range(10):
            kama4.update(0.0)
        out = kama4.update_quote(Quote(time=t, bid_price=inp, ask_price=inp, bid_size=0, ask_size=0))
        self.assertEqual(1, len(out))
        self.assertEqual(expected, out[0].value)

        # Trade
        kama5 = KaufmanAdaptiveMovingAverage.from_length(
            KaufmanAdaptiveMovingAverageLengthParams(
                efficiency_ratio_length=10, fastest_length=2, slowest_length=30))
        for _ in range(10):
            kama5.update(0.0)
        out = kama5.update_trade(Trade(time=t, price=inp, volume=0))
        self.assertEqual(1, len(out))
        self.assertEqual(expected, out[0].value)


if __name__ == '__main__':
    unittest.main()
