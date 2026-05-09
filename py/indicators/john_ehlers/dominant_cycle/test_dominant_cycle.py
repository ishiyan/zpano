"""Tests for the DominantCycle indicator."""

import math
import unittest
from datetime import datetime

from py.indicators.john_ehlers.dominant_cycle.dominant_cycle import DominantCycle
from py.indicators.john_ehlers.dominant_cycle.params import DominantCycleParams
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_type import CycleEstimatorType
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_params import CycleEstimatorParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade

from .test_testdata import (
    test_input,
    test_expected_period,
    test_expected_phase,
    TOLERANCE,
)


def phase_diff(a: float, b: float) -> float:
    """Shortest signed angular difference between two angles."""
    d = (a - b) % 360
    if d > 180:
        d -= 360
    elif d <= -180:
        d += 360
    return d


def create_default():
    return DominantCycle.create_default()


def create_alpha(alpha: float, estimator_type: CycleEstimatorType):
    params = DominantCycleParams(
        estimator_type=estimator_type,
        estimator_params=CycleEstimatorParams(
            smoothing_length=4, alpha_ema_quadrature_in_phase=0.2,
            alpha_ema_period=0.2, warm_up_period=0),
        alpha_ema_period_additional=alpha,
    )
    return DominantCycle.create(params)


class TestDominantCycleUpdate(unittest.TestCase):
    def test_reference_period(self):
        skip = 9
        settle_skip = 177
        inp = test_input()
        exp_period = test_expected_period()
        dc = create_default()

        for i in range(skip, len(inp)):
            _, period, _ = dc.update(inp[i])
            if math.isnan(period) or i < settle_skip:
                continue
            self.assertAlmostEqual(period, exp_period[i], delta=TOLERANCE,
                                   msg=f"[{i}] period: expected {exp_period[i]}, got {period}")

    def test_reference_phase(self):
        skip = 9
        settle_skip = 177
        inp = test_input()
        exp_phase = test_expected_phase()
        dc = create_default()

        for i in range(skip, len(inp)):
            _, _, phase = dc.update(inp[i])
            if math.isnan(phase) or i < settle_skip:
                continue
            if math.isnan(exp_phase[i]):
                continue
            self.assertAlmostEqual(phase_diff(exp_phase[i], phase), 0.0, delta=TOLERANCE,
                                   msg=f"[{i}] phase: expected {exp_phase[i]}, got {phase}")

    def test_nan_input(self):
        dc = create_default()
        raw, per, pha = dc.update(math.nan)
        self.assertTrue(math.isnan(raw))
        self.assertTrue(math.isnan(per))
        self.assertTrue(math.isnan(pha))


class TestDominantCycleIsPrimed(unittest.TestCase):
    def test_primes_within_sequence(self):
        inp = test_input()
        dc = create_default()
        self.assertFalse(dc.is_primed())
        primed_at = -1
        for i in range(len(inp)):
            dc.update(inp[i])
            if dc.is_primed() and primed_at < 0:
                primed_at = i
        self.assertGreaterEqual(primed_at, 0, "should become primed")
        self.assertTrue(dc.is_primed())


class TestDominantCycleUpdateEntity(unittest.TestCase):
    def test_update_scalar(self):
        dc = create_default()
        for _ in range(30):
            dc.update(100.0)
        t = datetime(2021, 4, 1)
        out = dc.update_scalar(Scalar(time=t, value=100.0))
        self.assertEqual(len(out), 3)
        for o in out:
            self.assertIsInstance(o, Scalar)
            self.assertEqual(o.time, t)

    def test_update_bar(self):
        dc = create_default()
        for _ in range(30):
            dc.update(100.0)
        t = datetime(2021, 4, 1)
        out = dc.update_bar(Bar(time=t, open=100.0, high=100.0, low=100.0, close=100.0, volume=0.0))
        self.assertEqual(len(out), 3)

    def test_update_quote(self):
        dc = create_default()
        for _ in range(30):
            dc.update(100.0)
        t = datetime(2021, 4, 1)
        out = dc.update_quote(Quote(time=t, bid_price=100.0, ask_price=100.0, bid_size=1.0, ask_size=1.0))
        self.assertEqual(len(out), 3)

    def test_update_trade(self):
        dc = create_default()
        for _ in range(30):
            dc.update(100.0)
        t = datetime(2021, 4, 1)
        out = dc.update_trade(Trade(time=t, price=100.0, volume=0.0))
        self.assertEqual(len(out), 3)


class TestDominantCycleMetadata(unittest.TestCase):
    def test_default(self):
        dc = create_default()
        m = dc.metadata()
        self.assertEqual(m.identifier, Identifier.DOMINANT_CYCLE)
        self.assertEqual(m.mnemonic, "dcp(0.330)")
        self.assertEqual(m.description, "Dominant cycle period dcp(0.330)")
        self.assertEqual(len(m.outputs), 3)
        self.assertEqual(m.outputs[0].mnemonic, "dcp-raw(0.330)")
        self.assertEqual(m.outputs[1].mnemonic, "dcp(0.330)")
        self.assertEqual(m.outputs[2].mnemonic, "dcph(0.330)")

    def test_phase_accumulator(self):
        dc = create_alpha(0.5, CycleEstimatorType.PHASE_ACCUMULATOR)
        m = dc.metadata()
        self.assertEqual(m.mnemonic, "dcp(0.500, pa(4, 0.200, 0.200))")


class TestDominantCycleConstruction(unittest.TestCase):
    def test_invalid_alpha_zero(self):
        with self.assertRaises(ValueError):
            DominantCycle.create(DominantCycleParams(alpha_ema_period_additional=0.0))

    def test_invalid_alpha_gt_one(self):
        with self.assertRaises(ValueError):
            DominantCycle.create(DominantCycleParams(alpha_ema_period_additional=1.00000001))


class TestDominantCycleSmoothedPrice(unittest.TestCase):
    def test_nan_before_primed(self):
        dc = create_default()
        self.assertTrue(math.isnan(dc.smoothed_price()))

    def test_finite_after_primed(self):
        inp = test_input()
        dc = create_default()
        for v in inp:
            dc.update(v)
            if dc.is_primed():
                self.assertFalse(math.isnan(dc.smoothed_price()))
                break


class TestDominantCycleMaxPeriod(unittest.TestCase):
    def test_max_period(self):
        dc = create_default()
        self.assertEqual(dc.max_period(), 50)


if __name__ == '__main__':
    unittest.main()
