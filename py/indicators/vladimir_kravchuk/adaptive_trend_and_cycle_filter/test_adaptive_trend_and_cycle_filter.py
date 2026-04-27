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


TOLERANCE = 1e-10

INPUT = [
    92.0000, 93.1725, 95.3125, 94.8450, 94.4075, 94.1100, 93.5000, 91.7350, 90.9550, 91.6875,
    94.5000, 97.9700, 97.5775, 90.7825, 89.0325, 92.0950, 91.1550, 89.7175, 90.6100, 91.0000,
    88.9225, 87.5150, 86.4375, 83.8900, 83.0025, 82.8125, 82.8450, 86.7350, 86.8600, 87.5475,
    85.7800, 86.1725, 86.4375, 87.2500, 88.9375, 88.2050, 85.8125, 84.5950, 83.6575, 84.4550,
    83.5000, 86.7825, 88.1725, 89.2650, 90.8600, 90.7825, 91.8600, 90.3600, 89.8600, 90.9225,
    89.5000, 87.6725, 86.5000, 84.2825, 82.9075, 84.2500, 85.6875, 86.6100, 88.2825, 89.5325,
    89.5000, 88.0950, 90.6250, 92.2350, 91.6725, 92.5925, 93.0150, 91.1725, 90.9850, 90.3775,
    88.2500, 86.9075, 84.0925, 83.1875, 84.2525, 97.8600, 99.8750, 103.2650, 105.9375, 103.5000,
    103.1100, 103.6100, 104.6400, 106.8150, 104.9525, 105.5000, 107.1400, 109.7350, 109.8450, 110.9850,
    120.0000, 119.8750, 117.9075, 119.4075, 117.9525, 117.2200, 115.6425, 113.1100, 111.7500, 114.5175,
    114.7450, 115.4700, 112.5300, 112.0300, 113.4350, 114.2200, 119.5950, 117.9650, 118.7150, 115.0300,
    114.5300, 115.0000, 116.5300, 120.1850, 120.5000, 120.5950, 124.1850, 125.3750, 122.9700, 123.0000,
    124.4350, 123.4400, 124.0300, 128.1850, 129.6550, 130.8750, 132.3450, 132.0650, 133.8150, 135.6600,
    137.0350, 137.4700, 137.3450, 136.3150, 136.4400, 136.2850, 129.0950, 128.3100, 126.0000, 124.0300,
    123.9350, 125.0300, 127.2500, 125.6200, 125.5300, 123.9050, 120.6550, 119.9650, 120.7800, 124.0000,
    122.7800, 120.7200, 121.7800, 122.4050, 123.2500, 126.1850, 127.5600, 126.5650, 123.0600, 122.7150,
    123.5900, 122.3100, 122.4650, 123.9650, 123.9700, 124.1550, 124.4350, 127.0000, 125.5000, 128.8750,
    130.5350, 132.3150, 134.0650, 136.0350, 133.7800, 132.7500, 133.4700, 130.9700, 127.5950, 128.4400,
    127.9400, 125.8100, 124.6250, 122.7200, 124.0900, 123.2200, 121.4050, 120.9350, 118.2800, 118.3750,
    121.1550, 120.9050, 117.1250, 113.0600, 114.9050, 112.4350, 107.9350, 105.9700, 106.3700, 106.8450,
    106.9700, 110.0300, 91.0000, 93.5600, 93.6200, 95.3100, 94.1850, 94.7800, 97.6250, 97.5900,
    95.2500, 94.7200, 92.2200, 91.5650, 92.2200, 93.8100, 95.5900, 96.1850, 94.6250, 95.1200,
    94.0000, 93.7450, 95.9050, 101.7450, 106.4400, 107.9350, 103.4050, 105.0600, 104.1550, 103.3100,
    103.3450, 104.8400, 110.4050, 114.5000, 117.3150, 118.2500, 117.1850, 109.7500, 109.6550, 108.5300,
    106.2200, 107.7200, 109.8400, 109.0950, 109.0900, 109.1550, 109.3150, 109.0600, 109.9050, 109.6250,
    109.5300, 108.0600,
]

N = float('nan')

# Snapshots: (index, fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci)
SNAPSHOTS = [
    (0, N, N, N, N, N, N, N, N),
    (38, 84.9735715498821, N, N, N, N, N, N, -1.3160715498821),
    (39, 84.4518660416872, N, N, N, N, N, N, 0.0031339583128),
    (43, 88.2793028340854, N, 84.9781981272507, N, N, 3.3011047068347, N, 0.9856971659146),
    (44, 90.3071933727095, N, 85.3111711946473, N, N, 4.9960221780622, N, 0.5528066272905),
    (55, 83.5737547263234, N, 87.6545375029340, N, -701.3930208567576, -4.0807827766106, N, 0.6762452736766),
    (56, 84.2004074439195, N, 86.4101353078987, N, -596.7632782263086, -2.2097278639792, N, 1.4870925560805),
    (64, 91.3026041176860, 89.8909098632724, 89.2605446508615, N, 260.0958399205915, 2.0420594668245, N, 0.3698958823140),
    (65, 91.9122247829182, 90.3013166280409, 90.0608560382592, N, 271.4055284612814, 1.8513687446590, N, 0.6802752170818),
    (90, 115.0676036598003, 109.5130909788342, 106.9904903948140, 91.0255929287335, 648.4101282691054, 8.0771132649863, 18.4874980501007, 4.9323963401997),
    (91, 117.8447026727287, 111.5377810965825, 108.9908122410267, 91.4218609612485, 750.5214819459538, 8.8538904317020, 20.1159201353340, 2.0302973272713),
    (100, 112.8634350429428, 119.4023289602100, 115.8265249211198, 97.7871686087879, -617.3149799371608, -2.9630898781769, 21.6151603514221, 1.8815649570572),
    (150, 121.5097808704445, 124.0945687443045, 123.2003217712845, 127.9357790331669, -268.9358266646477, -1.6905409008400, -3.8412102888624, 1.2702191295555),
    (200, 106.1833142820738, 109.8912725552509, 109.8071754394800, 127.4173713354640, -592.7380669351005, -3.6238611574062, -17.5260987802131, 0.7866857179262),
    (251, 108.1030068950443, 114.1981767327412, 110.1319723971535, 102.4461386298790, -312.3373212974634, -2.0289655021092, 11.7520381028621, -0.0430068950443),
]


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
