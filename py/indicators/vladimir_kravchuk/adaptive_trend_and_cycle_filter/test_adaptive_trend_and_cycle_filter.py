"""Tests for the Adaptive Trend and Cycle Filter (ATCF) indicator."""

import math
import unittest
from datetime import datetime

from py.indicators.vladimir_kravchuk.adaptive_trend_and_cycle_filter.adaptive_trend_and_cycle_filter import AdaptiveTrendAndCycleFilter
from py.indicators.vladimir_kravchuk.adaptive_trend_and_cycle_filter.params import AdaptiveTrendAndCycleFilterParams
from py.indicators.vladimir_kravchuk.adaptive_trend_and_cycle_filter.output import AdaptiveTrendAndCycleFilterOutput
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade

from .test_testdata import (
    TOLERANCE,
    INPUT,
    N,
    SNAPSHOTS,
)


# Snapshots: (index, fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci)
def close_enough(exp: float, got: float) -> bool:
    if math.isnan(exp):
        return math.isnan(got)
    return abs(exp - got) <= TOLERANCE


class TestAdaptiveTrendAndCycleFilter(unittest.TestCase):

    def test_update(self):
        x = AdaptiveTrendAndCycleFilter(AdaptiveTrendAndCycleFilterParams())
        si = 0
        for i, sample in enumerate(INPUT):
            fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci = x.update(sample)
            if si < len(SNAPSHOTS) and SNAPSHOTS[si][0] == i:
                s = SNAPSHOTS[si]
                names = ['fatl', 'satl', 'rftl', 'rstl', 'rbci', 'ftlm', 'stlm', 'pcci']
                vals = [fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci]
                for j, (name, val) in enumerate(zip(names, vals)):
                    self.assertTrue(
                        close_enough(s[j + 1], val),
                        f"[{i}] {name}: expected {s[j + 1]}, got {val}")
                si += 1
        self.assertEqual(si, len(SNAPSHOTS), "did not hit all snapshots")

    def test_primes_at_bar_90(self):
        x = AdaptiveTrendAndCycleFilter(AdaptiveTrendAndCycleFilterParams())
        self.assertFalse(x.is_primed())
        primed_at = -1
        for i, sample in enumerate(INPUT):
            x.update(sample)
            if x.is_primed() and primed_at < 0:
                primed_at = i
        self.assertEqual(primed_at, 90)

    def test_nan_input(self):
        x = AdaptiveTrendAndCycleFilter(AdaptiveTrendAndCycleFilterParams())
        result = x.update(math.nan)
        for v in result:
            self.assertTrue(math.isnan(v), f"expected NaN, got {v}")
        self.assertFalse(x.is_primed())

    def test_metadata(self):
        x = AdaptiveTrendAndCycleFilter(AdaptiveTrendAndCycleFilterParams())
        md = x.metadata()
        mn = "atcf()"
        self.assertEqual(md.identifier, Identifier.ADAPTIVE_TREND_AND_CYCLE_FILTER)
        self.assertEqual(md.mnemonic, mn)
        self.assertEqual(md.description, f"Adaptive trend and cycle filter {mn}")
        self.assertEqual(len(md.outputs), 8)

        expected_mnemonics = [
            "fatl()", "satl()", "rftl()", "rstl()",
            "rbci()", "ftlm()", "stlm()", "pcci()",
        ]
        for i, em in enumerate(expected_mnemonics):
            self.assertEqual(md.outputs[i].mnemonic, em)

    def test_update_entities(self):
        tm = datetime(2021, 4, 1)
        inp_val = 100.0

        # Prime the indicator first
        def make_primed():
            x = AdaptiveTrendAndCycleFilter(AdaptiveTrendAndCycleFilterParams())
            for i in range(100):
                x.update(INPUT[i])
            return x

        # Scalar
        x = make_primed()
        result = x.update_scalar(Scalar(time=tm, value=inp_val))
        self.assertEqual(len(result), 8)
        for i, s in enumerate(result):
            self.assertIsInstance(s, Scalar)
            self.assertEqual(s.time, tm)

        # Bar
        x = make_primed()
        result = x.update_bar(Bar(time=tm, open=inp_val, high=inp_val, low=inp_val, close=inp_val, volume=0))
        self.assertEqual(len(result), 8)
        for s in result:
            self.assertEqual(s.time, tm)

        # Quote
        x = make_primed()
        result = x.update_quote(Quote(time=tm, bid_price=inp_val, ask_price=inp_val, bid_size=1, ask_size=1))
        self.assertEqual(len(result), 8)
        for s in result:
            self.assertEqual(s.time, tm)

        # Trade
        x = make_primed()
        result = x.update_trade(Trade(time=tm, price=inp_val, volume=1))
        self.assertEqual(len(result), 8)
        for s in result:
            self.assertEqual(s.time, tm)


if __name__ == '__main__':
    unittest.main()
