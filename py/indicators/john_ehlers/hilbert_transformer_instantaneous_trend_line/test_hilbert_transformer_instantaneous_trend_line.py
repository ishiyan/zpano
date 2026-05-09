"""Tests for the HilbertTransformerInstantaneousTrendLine indicator."""

import math
import unittest

from py.indicators.john_ehlers.hilbert_transformer_instantaneous_trend_line.hilbert_transformer_instantaneous_trend_line import HilbertTransformerInstantaneousTrendLine
from py.indicators.john_ehlers.hilbert_transformer_instantaneous_trend_line.params import HilbertTransformerInstantaneousTrendLineParams
from py.indicators.john_ehlers.hilbert_transformer_instantaneous_trend_line.output import HilbertTransformerInstantaneousTrendLineOutput
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_type import CycleEstimatorType
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_params import CycleEstimatorParams
from py.entities.bar_component import BarComponent
from py.entities.quote_component import QuoteComponent
from py.entities.trade_component import TradeComponent

from .test_testdata import (
    _TIME,
    _INPUT,
    _EXPECTED_PERIOD,
    _EXPECTED_VALUE,
    _TOLERANCE,
    _SKIP,
    _SETTLE_SKIP,
)


# Input data from TA-Lib test_MAMA.xsl, 252 entries.
# Expected period data from test_MAMA.xsl, smoothed, 252 entries.
# Expected value data from MBST InstantaneousTrendLineTest.cs, 252 entries.
class TestHilbertTransformerInstantaneousTrendLineUpdate(unittest.TestCase):

    def test_reference_value(self):
        x = HilbertTransformerInstantaneousTrendLine.create_default()
        for i in range(_SKIP, len(_INPUT)):
            value, _ = x.update(_INPUT[i])
            if math.isnan(value) or i < _SETTLE_SKIP:
                continue
            if math.isnan(_EXPECTED_VALUE[i]):
                continue
            self.assertAlmostEqual(value, _EXPECTED_VALUE[i], delta=_TOLERANCE,
                                   msg=f"[{i}] value")

    def test_reference_period(self):
        x = HilbertTransformerInstantaneousTrendLine.create_default()
        for i in range(_SKIP, len(_INPUT)):
            _, period = x.update(_INPUT[i])
            if math.isnan(period) or i < _SETTLE_SKIP:
                continue
            self.assertAlmostEqual(period, _EXPECTED_PERIOD[i], delta=_TOLERANCE,
                                   msg=f"[{i}] period")

    def test_nan_input(self):
        x = HilbertTransformerInstantaneousTrendLine.create_default()
        value, period = x.update(math.nan)
        self.assertTrue(math.isnan(value))
        self.assertTrue(math.isnan(period))


class TestHilbertTransformerInstantaneousTrendLineIsPrimed(unittest.TestCase):

    def test_primes_within_input(self):
        x = HilbertTransformerInstantaneousTrendLine.create_default()
        self.assertFalse(x.is_primed())
        primed_at = -1
        for i, v in enumerate(_INPUT):
            x.update(v)
            if x.is_primed() and primed_at < 0:
                primed_at = i
        self.assertGreaterEqual(primed_at, 0)
        self.assertTrue(x.is_primed())


class TestHilbertTransformerInstantaneousTrendLineMetadata(unittest.TestCase):

    def test_metadata(self):
        x = HilbertTransformerInstantaneousTrendLine.create_default()
        m = x.metadata()
        mnemonic = "htitl(0.330, 4, 1.000, hl/2)"
        mnemonic_dcp = "dcp(0.330, hl/2)"
        self.assertEqual(m.identifier, Identifier.HILBERT_TRANSFORMER_INSTANTANEOUS_TREND_LINE)
        self.assertEqual(m.mnemonic, mnemonic)
        self.assertEqual(m.description, f"Hilbert transformer instantaneous trend line {mnemonic}")
        self.assertEqual(len(m.outputs), 2)
        self.assertEqual(m.outputs[0].mnemonic, mnemonic)
        self.assertEqual(m.outputs[1].mnemonic, mnemonic_dcp)


class TestHilbertTransformerInstantaneousTrendLineUpdateEntity(unittest.TestCase):

    def _prime(self):
        x = HilbertTransformerInstantaneousTrendLine.create_default()
        for i in range(200):
            x.update(_INPUT[i % len(_INPUT)])
        return x

    def test_update_scalar(self):
        x = self._prime()
        out = x.update_scalar(Scalar(time=_TIME, value=100.0))
        self.assertEqual(len(out), 2)
        self.assertIsInstance(out[0], Scalar)
        self.assertIsInstance(out[1], Scalar)

    def test_update_bar(self):
        x = self._prime()
        out = x.update_bar(Bar(time=_TIME, open=100, high=100, low=100, close=100, volume=0))
        self.assertEqual(len(out), 2)

    def test_update_quote(self):
        x = self._prime()
        out = x.update_quote(Quote(time=_TIME, bid_price=100, ask_price=100, bid_size=1, ask_size=1))
        self.assertEqual(len(out), 2)

    def test_update_trade(self):
        x = self._prime()
        out = x.update_trade(Trade(time=_TIME, price=100, volume=1))
        self.assertEqual(len(out), 2)


class TestNewHilbertTransformerInstantaneousTrendLine(unittest.TestCase):

    def _ce_params(self):
        return CycleEstimatorParams(smoothing_length=4, alpha_ema_quadrature_in_phase=0.2,
                                    alpha_ema_period=0.2, warm_up_period=0)

    def test_default(self):
        x = HilbertTransformerInstantaneousTrendLine.create_default()
        self.assertEqual(x._mnemonic, "htitl(0.330, 4, 1.000, hl/2)")
        self.assertFalse(x.is_primed())

    def test_tlsl_2(self):
        p = HilbertTransformerInstantaneousTrendLineParams(
            alpha_ema_period_additional=0.33,
            estimator_type=CycleEstimatorType.HOMODYNE_DISCRIMINATOR,
            estimator_params=self._ce_params(),
            trend_line_smoothing_length=2,
            cycle_part_multiplier=1.0,
            bar_component=BarComponent.CLOSE,
            quote_component=QuoteComponent.MID,
            trade_component=TradeComponent.PRICE,
        )
        x = HilbertTransformerInstantaneousTrendLine.create(p)
        self.assertEqual(x._mnemonic, "htitl(0.330, 2, 1.000)")
        self.assertAlmostEqual(x._coeff0, 2.0 / 3.0)
        self.assertAlmostEqual(x._coeff1, 1.0 / 3.0)

    def test_tlsl_3_phase_accumulator(self):
        p = HilbertTransformerInstantaneousTrendLineParams(
            alpha_ema_period_additional=0.5,
            estimator_type=CycleEstimatorType.PHASE_ACCUMULATOR,
            estimator_params=self._ce_params(),
            trend_line_smoothing_length=3,
            cycle_part_multiplier=0.5,
            bar_component=BarComponent.CLOSE,
            quote_component=QuoteComponent.MID,
            trade_component=TradeComponent.PRICE,
        )
        x = HilbertTransformerInstantaneousTrendLine.create(p)
        self.assertEqual(x._mnemonic, "htitl(0.500, 3, 0.500, pa(4, 0.200, 0.200))")

    def test_alpha_zero(self):
        p = HilbertTransformerInstantaneousTrendLineParams(
            alpha_ema_period_additional=0.0,
            estimator_params=self._ce_params(),
            trend_line_smoothing_length=4, cycle_part_multiplier=1.0)
        with self.assertRaises(ValueError):
            HilbertTransformerInstantaneousTrendLine.create(p)

    def test_alpha_above_one(self):
        p = HilbertTransformerInstantaneousTrendLineParams(
            alpha_ema_period_additional=1.00000001,
            estimator_params=self._ce_params(),
            trend_line_smoothing_length=4, cycle_part_multiplier=1.0)
        with self.assertRaises(ValueError):
            HilbertTransformerInstantaneousTrendLine.create(p)

    def test_tlsl_below_2(self):
        p = HilbertTransformerInstantaneousTrendLineParams(
            alpha_ema_period_additional=0.33,
            estimator_params=self._ce_params(),
            trend_line_smoothing_length=1, cycle_part_multiplier=1.0)
        with self.assertRaises(ValueError):
            HilbertTransformerInstantaneousTrendLine.create(p)

    def test_tlsl_above_4(self):
        p = HilbertTransformerInstantaneousTrendLineParams(
            alpha_ema_period_additional=0.33,
            estimator_params=self._ce_params(),
            trend_line_smoothing_length=5, cycle_part_multiplier=1.0)
        with self.assertRaises(ValueError):
            HilbertTransformerInstantaneousTrendLine.create(p)

    def test_cpm_zero(self):
        p = HilbertTransformerInstantaneousTrendLineParams(
            alpha_ema_period_additional=0.33,
            estimator_params=self._ce_params(),
            trend_line_smoothing_length=4, cycle_part_multiplier=0.0)
        with self.assertRaises(ValueError):
            HilbertTransformerInstantaneousTrendLine.create(p)

    def test_cpm_above_10(self):
        p = HilbertTransformerInstantaneousTrendLineParams(
            alpha_ema_period_additional=0.33,
            estimator_params=self._ce_params(),
            trend_line_smoothing_length=4, cycle_part_multiplier=10.00001)
        with self.assertRaises(ValueError):
            HilbertTransformerInstantaneousTrendLine.create(p)


if __name__ == '__main__':
    unittest.main()
