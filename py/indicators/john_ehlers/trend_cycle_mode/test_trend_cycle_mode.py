"""Tests for TrendCycleMode indicator."""

import math
import unittest
from datetime import datetime

from py.indicators.john_ehlers.trend_cycle_mode.trend_cycle_mode import TrendCycleMode
from py.indicators.john_ehlers.trend_cycle_mode.params import TrendCycleModeParams
from py.indicators.john_ehlers.trend_cycle_mode.output import TrendCycleModeOutput
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
    TCM_INPUT,
    TCM_EXPECTED_PERIOD,
    TCM_EXPECTED_PHASE,
    TCM_EXPECTED_SINE,
    TCM_EXPECTED_SINE_LEAD,
    TCM_EXPECTED_ITL,
    TCM_EXPECTED_VALUE,
    TOLERANCE,
    SKIP,
    SETTLE_SKIP,
)


def _phase_diff(expected, actual):
    """Compare phases modulo 360."""
    d = (expected - actual) % 360.0
    if d > 180:
        d -= 360
    elif d < -180:
        d += 360
    return abs(d)


class TestTrendCycleModeUpdate(unittest.TestCase):

    def test_reference_period(self):
        x = TrendCycleMode.create_default()
        for i in range(SKIP, len(TCM_INPUT)):
            _, _, _, _, _, _, period, _ = x.update(TCM_INPUT[i])
            if math.isnan(period) or i < SETTLE_SKIP:
                continue
            self.assertAlmostEqual(TCM_EXPECTED_PERIOD[i], period, delta=TOLERANCE,
                                   msg=f"[{i}] period")

    def test_reference_phase(self):
        x = TrendCycleMode.create_default()
        for i in range(SKIP, len(TCM_INPUT)):
            _, _, _, _, _, _, _, phase = x.update(TCM_INPUT[i])
            if math.isnan(phase) or math.isnan(TCM_EXPECTED_PHASE[i]) or i < SETTLE_SKIP:
                continue
            d = _phase_diff(TCM_EXPECTED_PHASE[i], phase)
            self.assertLessEqual(d, TOLERANCE, msg=f"[{i}] phase: expected {TCM_EXPECTED_PHASE[i]}, actual {phase}")

    def test_reference_sine_wave(self):
        x = TrendCycleMode.create_default()
        for i in range(SKIP, len(TCM_INPUT)):
            _, _, _, _, sine, _, _, _ = x.update(TCM_INPUT[i])
            if math.isnan(sine) or math.isnan(TCM_EXPECTED_SINE[i]) or i < SETTLE_SKIP:
                continue
            self.assertAlmostEqual(TCM_EXPECTED_SINE[i], sine, delta=TOLERANCE,
                                   msg=f"[{i}] sine")

    def test_reference_sine_wave_lead(self):
        x = TrendCycleMode.create_default()
        for i in range(SKIP, len(TCM_INPUT)):
            _, _, _, _, _, sine_lead, _, _ = x.update(TCM_INPUT[i])
            if math.isnan(sine_lead) or math.isnan(TCM_EXPECTED_SINE_LEAD[i]) or i < SETTLE_SKIP:
                continue
            self.assertAlmostEqual(TCM_EXPECTED_SINE_LEAD[i], sine_lead, delta=TOLERANCE,
                                   msg=f"[{i}] sineLead")

    def test_reference_itl(self):
        x = TrendCycleMode.create_default()
        for i in range(SKIP, len(TCM_INPUT)):
            _, _, _, itl, _, _, _, _ = x.update(TCM_INPUT[i])
            if math.isnan(itl) or math.isnan(TCM_EXPECTED_ITL[i]) or i < SETTLE_SKIP:
                continue
            self.assertAlmostEqual(TCM_EXPECTED_ITL[i], itl, delta=TOLERANCE,
                                   msg=f"[{i}] itl")

    def test_reference_value(self):
        x = TrendCycleMode.create_default()
        limit = len(TCM_EXPECTED_VALUE)
        for i in range(SKIP, len(TCM_INPUT)):
            value, _, _, _, _, _, _, _ = x.update(TCM_INPUT[i])
            if i >= limit:
                continue
            # MBST known mismatches.
            if i == 70 or i == 71:
                continue
            if math.isnan(value) or math.isnan(TCM_EXPECTED_VALUE[i]):
                continue
            self.assertAlmostEqual(TCM_EXPECTED_VALUE[i], value, delta=TOLERANCE,
                                   msg=f"[{i}] value")

    def test_trend_cycle_complementary(self):
        x = TrendCycleMode.create_default()
        for i in range(SKIP, len(TCM_INPUT)):
            value, trend, cycle, _, _, _, _, _ = x.update(TCM_INPUT[i])
            if math.isnan(value):
                continue
            self.assertEqual(trend + cycle, 1.0, msg=f"[{i}] trend+cycle")
            if value > 0:
                self.assertEqual(trend, 1.0, msg=f"[{i}] value>0 but trend!=1")
            if value < 0:
                self.assertEqual(trend, 0.0, msg=f"[{i}] value<0 but trend!=0")

    def test_nan_input(self):
        x = TrendCycleMode.create_default()
        result = x.update(math.nan)
        for i, v in enumerate(result):
            self.assertTrue(math.isnan(v), msg=f"output[{i}] should be NaN")


class TestTrendCycleModeIsPrimed(unittest.TestCase):

    def test_primes_within_input(self):
        x = TrendCycleMode.create_default()
        self.assertFalse(x.is_primed())
        primed_at = -1
        for i in range(len(TCM_INPUT)):
            x.update(TCM_INPUT[i])
            if x.is_primed() and primed_at < 0:
                primed_at = i
        self.assertGreaterEqual(primed_at, 0)
        self.assertTrue(x.is_primed())


class TestTrendCycleModeMetadata(unittest.TestCase):

    def test_metadata(self):
        x = TrendCycleMode.create_default()
        m = x.metadata()

        mn_value = "tcm(0.330, 4, 1.000, 1.500%, hl/2)"
        mn_trend = "tcm-trend(0.330, 4, 1.000, 1.500%, hl/2)"
        mn_cycle = "tcm-cycle(0.330, 4, 1.000, 1.500%, hl/2)"
        mn_itl = "tcm-itl(0.330, 4, 1.000, 1.500%, hl/2)"
        mn_sine = "tcm-sine(0.330, 4, 1.000, 1.500%, hl/2)"
        mn_sine_lead = "tcm-sineLead(0.330, 4, 1.000, 1.500%, hl/2)"
        mn_dcp = "dcp(0.330, hl/2)"
        mn_dc_phase = "dcph(0.330, hl/2)"

        self.assertEqual(Identifier.TREND_CYCLE_MODE, m.identifier)
        self.assertEqual(mn_value, m.mnemonic)
        self.assertEqual("Trend versus cycle mode " + mn_value, m.description)
        self.assertEqual(8, len(m.outputs))

        self.assertEqual(mn_value, m.outputs[0].mnemonic)
        self.assertEqual(mn_trend, m.outputs[1].mnemonic)
        self.assertEqual(mn_cycle, m.outputs[2].mnemonic)
        self.assertEqual(mn_itl, m.outputs[3].mnemonic)
        self.assertEqual(mn_sine, m.outputs[4].mnemonic)
        self.assertEqual(mn_sine_lead, m.outputs[5].mnemonic)
        self.assertEqual(mn_dcp, m.outputs[6].mnemonic)
        self.assertEqual(mn_dc_phase, m.outputs[7].mnemonic)

        self.assertEqual("Dominant cycle period " + mn_dcp, m.outputs[6].description)
        self.assertEqual("Dominant cycle phase " + mn_dc_phase, m.outputs[7].description)


class TestTrendCycleModeUpdateEntity(unittest.TestCase):

    def _prime_and_check(self, x, update_fn):
        inp = TCM_INPUT
        for i in range(200):
            x.update(inp[i % len(inp)])
        result = update_fn()
        self.assertEqual(8, len(result))
        for i, s in enumerate(result):
            self.assertIsInstance(s, Scalar, msg=f"output[{i}] not a Scalar")

    def test_update_scalar(self):
        x = TrendCycleMode.create_default()
        tm = datetime(2021, 4, 1)
        s = Scalar(time=tm, value=100.0)
        self._prime_and_check(x, lambda: x.update_scalar(s))

    def test_update_bar(self):
        x = TrendCycleMode.create_default()
        tm = datetime(2021, 4, 1)
        b = Bar(time=tm, open=100.0, high=100.0, low=100.0, close=100.0, volume=0.0)
        self._prime_and_check(x, lambda: x.update_bar(b))

    def test_update_quote(self):
        x = TrendCycleMode.create_default()
        tm = datetime(2021, 4, 1)
        q = Quote(time=tm, bid_price=100.0, ask_price=100.0, bid_size=0.0, ask_size=0.0)
        self._prime_and_check(x, lambda: x.update_quote(q))

    def test_update_trade(self):
        x = TrendCycleMode.create_default()
        tm = datetime(2021, 4, 1)
        r = Trade(time=tm, price=100.0, volume=0.0)
        self._prime_and_check(x, lambda: x.update_trade(r))


class TestNewTrendCycleMode(unittest.TestCase):

    def _make_est_params(self):
        return CycleEstimatorParams(smoothing_length=4, alpha_ema_quadrature_in_phase=0.2,
                                    alpha_ema_period=0.2, warm_up_period=100)

    def test_default(self):
        x = TrendCycleMode.create_default()
        self.assertFalse(x.is_primed())

    def test_tlsl_2(self):
        p = TrendCycleModeParams(
            alpha_ema_period_additional=0.33,
            estimator_params=self._make_est_params(),
            trend_line_smoothing_length=2,
            cycle_part_multiplier=1.0,
            separation_percentage=1.5,
            bar_component=BarComponent.CLOSE,
            quote_component=QuoteComponent.MID,
            trade_component=TradeComponent.PRICE,
        )
        x = TrendCycleMode.create(p)
        self.assertAlmostEqual(x._coeff0, 2.0 / 3.0)
        self.assertAlmostEqual(x._coeff1, 1.0 / 3.0)

    def test_tlsl_3_phase_accumulator(self):
        p = TrendCycleModeParams(
            alpha_ema_period_additional=0.5,
            estimator_type=CycleEstimatorType.PHASE_ACCUMULATOR,
            estimator_params=self._make_est_params(),
            trend_line_smoothing_length=3,
            cycle_part_multiplier=0.5,
            separation_percentage=2.0,
            bar_component=BarComponent.CLOSE,
            quote_component=QuoteComponent.MID,
            trade_component=TradeComponent.PRICE,
        )
        x = TrendCycleMode.create(p)
        m = x.metadata()
        self.assertEqual("tcm(0.500, 3, 0.500, 2.000%, pa(4, 0.200, 0.200))", m.mnemonic)

    def test_alpha_le_0(self):
        p = TrendCycleModeParams(alpha_ema_period_additional=0.0, estimator_params=self._make_est_params())
        with self.assertRaises(ValueError):
            TrendCycleMode.create(p)

    def test_alpha_gt_1(self):
        p = TrendCycleModeParams(alpha_ema_period_additional=1.00000001, estimator_params=self._make_est_params())
        with self.assertRaises(ValueError):
            TrendCycleMode.create(p)

    def test_tlsl_lt_2(self):
        p = TrendCycleModeParams(estimator_params=self._make_est_params(), trend_line_smoothing_length=1)
        with self.assertRaises(ValueError):
            TrendCycleMode.create(p)

    def test_tlsl_gt_4(self):
        p = TrendCycleModeParams(estimator_params=self._make_est_params(), trend_line_smoothing_length=5)
        with self.assertRaises(ValueError):
            TrendCycleMode.create(p)

    def test_cpm_le_0(self):
        p = TrendCycleModeParams(estimator_params=self._make_est_params(), cycle_part_multiplier=0.0)
        with self.assertRaises(ValueError):
            TrendCycleMode.create(p)

    def test_cpm_gt_10(self):
        p = TrendCycleModeParams(estimator_params=self._make_est_params(), cycle_part_multiplier=10.00001)
        with self.assertRaises(ValueError):
            TrendCycleMode.create(p)

    def test_sep_le_0(self):
        p = TrendCycleModeParams(estimator_params=self._make_est_params(), separation_percentage=0.0)
        with self.assertRaises(ValueError):
            TrendCycleMode.create(p)

    def test_sep_gt_100(self):
        p = TrendCycleModeParams(estimator_params=self._make_est_params(), separation_percentage=100.00001)
        with self.assertRaises(ValueError):
            TrendCycleMode.create(p)


if __name__ == '__main__':
    unittest.main()
