"""Tests for the Jurik Moving Average (JMA) indicator."""

import math
import unittest
from datetime import datetime

from py.indicators.mark_jurik.jurik_moving_average.jurik_moving_average import JurikMovingAverage
from py.indicators.mark_jurik.jurik_moving_average.params import JurikMovingAverageParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade

from .test_testdata import (
    INPUT,
    LENGTH_20_PHASE_MIN100_OUTPUT,
    LENGTH_20_PHASE_MIN30_OUTPUT,
    LENGTH_20_PHASE_0_OUTPUT,
    LENGTH_20_PHASE_30_OUTPUT,
    LENGTH_20_PHASE_100_OUTPUT,
    LENGTH_2_PHASE_1_OUTPUT,
    LENGTH_5_PHASE_1_OUTPUT,
    LENGTH_10_PHASE_1_OUTPUT,
    LENGTH_25_PHASE_1_OUTPUT,
    LENGTH_50_PHASE_1_OUTPUT,
    LENGTH_75_PHASE_1_OUTPUT,
    LENGTH_100_PHASE_1_OUTPUT,
)


TOLERANCE = 1e-13

N = float('nan')
LEN_PRIMED = 30  # First 30 values (indices 0-29) are NaN

TEST_CASES = [
    (20, -100, LENGTH_20_PHASE_MIN100_OUTPUT),
    (20, -30, LENGTH_20_PHASE_MIN30_OUTPUT),
    (20, 0, LENGTH_20_PHASE_0_OUTPUT),
    (20, 30, LENGTH_20_PHASE_30_OUTPUT),
    (20, 100, LENGTH_20_PHASE_100_OUTPUT),
    (2, 1, LENGTH_2_PHASE_1_OUTPUT),
    (5, 1, LENGTH_5_PHASE_1_OUTPUT),
    (10, 1, LENGTH_10_PHASE_1_OUTPUT),
    (25, 1, LENGTH_25_PHASE_1_OUTPUT),
    (50, 1, LENGTH_50_PHASE_1_OUTPUT),
    (75, 1, LENGTH_75_PHASE_1_OUTPUT),
    (100, 1, LENGTH_100_PHASE_1_OUTPUT),
]


def close_enough(exp: float, got: float) -> bool:
    if math.isnan(exp):
        return math.isnan(got)
    return abs(exp - got) <= TOLERANCE


class TestJurikMovingAverage(unittest.TestCase):

    def test_update(self):
        for length, phase, expected in TEST_CASES:
            if not expected:
                continue
            with self.subTest(length=length, phase=phase):
                x = JurikMovingAverage(JurikMovingAverageParams(length=length, phase=phase))
                for i, sample in enumerate(INPUT):
                    result = x.update(sample)
                    self.assertTrue(
                        close_enough(expected[i], result),
                        f"[{i}] length={length}, phase={phase}: "
                        f"expected {expected[i]}, got {result}")

    def test_is_primed(self):
        x = JurikMovingAverage(JurikMovingAverageParams(length=20, phase=0))
        self.assertFalse(x.is_primed())
        primed_at = -1
        for i, sample in enumerate(INPUT):
            x.update(sample)
            if x.is_primed() and primed_at < 0:
                primed_at = i
        self.assertEqual(primed_at, LEN_PRIMED)

    def test_nan_input(self):
        """After priming, NaN input should return NaN."""
        x = JurikMovingAverage(JurikMovingAverageParams(length=20, phase=0))
        # Prime the indicator
        for sample in INPUT:
            x.update(sample)
        result = x.update(math.nan)
        self.assertTrue(math.isnan(result))

    def test_metadata(self):
        x = JurikMovingAverage(JurikMovingAverageParams(length=20, phase=0))
        md = x.metadata()
        self.assertEqual(md.identifier, Identifier.JURIK_MOVING_AVERAGE)
        self.assertEqual(len(md.outputs), 1)

    def test_update_entities(self):
        tm = datetime(2021, 4, 1)
        inp_val = 100.0

        def make_primed():
            x = JurikMovingAverage(JurikMovingAverageParams(length=20, phase=0))
            for sample in INPUT[:100]:
                x.update(sample)
            return x

        # Scalar
        x = make_primed()
        result = x.update_scalar(Scalar(time=tm, value=inp_val))
        self.assertEqual(len(result), 1)
        self.assertIsInstance(result[0], Scalar)
        self.assertEqual(result[0].time, tm)

        # Bar
        x = make_primed()
        result = x.update_bar(Bar(time=tm, open=inp_val, high=inp_val, low=inp_val, close=inp_val, volume=0))
        self.assertEqual(len(result), 1)
        self.assertIsInstance(result[0], Scalar)
        self.assertEqual(result[0].time, tm)

        # Quote
        x = make_primed()
        result = x.update_quote(Quote(time=tm, bid_price=inp_val, ask_price=inp_val, bid_size=1, ask_size=1))
        self.assertEqual(len(result), 1)
        self.assertIsInstance(result[0], Scalar)
        self.assertEqual(result[0].time, tm)

        # Trade
        x = make_primed()
        result = x.update_trade(Trade(time=tm, price=inp_val, volume=1))
        self.assertEqual(len(result), 1)
        self.assertIsInstance(result[0], Scalar)
        self.assertEqual(result[0].time, tm)


if __name__ == '__main__':
    unittest.main()
