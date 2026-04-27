"""Tests for the AutoCorrelation Periodogram indicator."""

import math
import unittest
import datetime

from ....entities.bar import Bar
from ....entities.bar_component import BarComponent
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ...core.identifier import Identifier
from .auto_correlation_periodogram import AutoCorrelationPeriodogram
from .params import Params


def _test_time():
    return datetime.datetime(2021, 4, 1, tzinfo=datetime.timezone.utc)


def _test_input():
    return [
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


_TOLERANCE = 1e-12
_MINMAX_TOL = 1e-10

_SNAPSHOTS = [
    {
        'i': 120, 'value_min': 0, 'value_max': 0.587896156859073,
        'spots': [
            (0, 0.005178925655054),
            (9, 0.002809468053362),
            (19, 0.003090504825683),
            (28, 0.000546592441807),
            (38, 0.000986630131293),
        ],
    },
    {
        'i': 150, 'value_min': 0, 'value_max': 0.176709059744408,
        'spots': [
            (0, 0.026721265051811),
            (9, 0.092237395343329),
            (19, 0.000025883994424),
            (28, 0.023733698891158),
            (38, 0.028877871176187),
        ],
    },
    {
        'i': 200, 'value_min': 0, 'value_max': 0.691497315922981,
        'spots': [
            (0, 0.000000000000000),
            (9, 0.664761990766748),
            (19, 0.004025474045612),
            (28, 0.017601099472114),
            (38, 0.055697077605643),
        ],
    },
    {
        'i': 250, 'value_min': 0, 'value_max': 0.133271103774289,
        'spots': [
            (0, 0.042309937132732),
            (9, 0.001944553293214),
            (19, 0.003966252606748),
            (28, 0.029986716389868),
            (38, 0.052695592829157),
        ],
    },
]


class TestAutoCorrelationPeriodogramUpdate(unittest.TestCase):
    def test_update(self):
        inp = _test_input()
        t0 = _test_time()
        x = AutoCorrelationPeriodogram(Params())

        si = 0
        for i, val in enumerate(inp):
            t = t0 + datetime.timedelta(minutes=i)
            h = x.update(val, t)

            self.assertIsNotNone(h)
            self.assertEqual(h.parameter_first, 10)
            self.assertEqual(h.parameter_last, 48)
            self.assertEqual(h.parameter_resolution, 1)

            if not x.is_primed():
                self.assertTrue(h.is_empty(), f"[{i}] expected empty before priming")
                continue

            self.assertEqual(len(h.values), 39, f"[{i}] values len")

            if si < len(_SNAPSHOTS) and _SNAPSHOTS[si]['i'] == i:
                snap = _SNAPSHOTS[si]
                self.assertAlmostEqual(h.value_min, snap['value_min'], delta=_MINMAX_TOL,
                                       msg=f"[{i}] ValueMin")
                self.assertAlmostEqual(h.value_max, snap['value_max'], delta=_MINMAX_TOL,
                                       msg=f"[{i}] ValueMax")
                for bin_idx, exp_val in snap['spots']:
                    self.assertAlmostEqual(h.values[bin_idx], exp_val, delta=_TOLERANCE,
                                           msg=f"[{i}] Values[{bin_idx}]")
                si += 1

        self.assertEqual(si, len(_SNAPSHOTS), "did not hit all snapshots")


class TestAutoCorrelationPeriodogramSyntheticSine(unittest.TestCase):
    def test_synthetic_sine(self):
        period = 20.0
        bars = 600
        t0 = _test_time()

        x = AutoCorrelationPeriodogram(Params(
            disable_automatic_gain_control=True,
            fixed_normalization=True,
        ))

        last = None
        for i in range(bars):
            sample = 100 + math.sin(2 * math.pi * i / period)
            last = x.update(sample, t0 + datetime.timedelta(minutes=i))

        self.assertIsNotNone(last)
        self.assertFalse(last.is_empty())

        peak_bin = 0
        for i in range(len(last.values)):
            if last.values[i] > last.values[peak_bin]:
                peak_bin = i

        expected_bin = int(period - last.parameter_first)
        self.assertEqual(peak_bin, expected_bin,
                         f"peak bin: expected {expected_bin} (period {period}), "
                         f"got {peak_bin} (period {last.parameter_first + peak_bin})")


class TestAutoCorrelationPeriodogramNaN(unittest.TestCase):
    def test_nan_input(self):
        x = AutoCorrelationPeriodogram(Params())
        h = x.update(math.nan, _test_time())
        self.assertTrue(h.is_empty())
        self.assertFalse(x.is_primed())


class TestAutoCorrelationPeriodogramMetadata(unittest.TestCase):
    def test_metadata(self):
        x = AutoCorrelationPeriodogram(Params())
        md = x.metadata()
        mn = "acp(10, 48, hl/2)"
        self.assertEqual(md.identifier, Identifier.AUTO_CORRELATION_PERIODOGRAM)
        self.assertEqual(md.mnemonic, mn)
        self.assertEqual(md.description, "Autocorrelation periodogram " + mn)
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].mnemonic, mn)


class TestAutoCorrelationPeriodogramMnemonicFlags(unittest.TestCase):
    def test_default(self):
        x = AutoCorrelationPeriodogram(Params())
        self.assertEqual(x.mnemonic, "acp(10, 48, hl/2)")

    def test_average_override(self):
        x = AutoCorrelationPeriodogram(Params(averaging_length=5))
        self.assertEqual(x.mnemonic, "acp(10, 48, average=5, hl/2)")

    def test_no_sqr(self):
        x = AutoCorrelationPeriodogram(Params(disable_spectral_squaring=True))
        self.assertEqual(x.mnemonic, "acp(10, 48, no-sqr, hl/2)")

    def test_no_smooth(self):
        x = AutoCorrelationPeriodogram(Params(disable_smoothing=True))
        self.assertEqual(x.mnemonic, "acp(10, 48, no-smooth, hl/2)")

    def test_no_agc(self):
        x = AutoCorrelationPeriodogram(Params(disable_automatic_gain_control=True))
        self.assertEqual(x.mnemonic, "acp(10, 48, no-agc, hl/2)")

    def test_agc_override(self):
        x = AutoCorrelationPeriodogram(Params(automatic_gain_control_decay_factor=0.8))
        self.assertEqual(x.mnemonic, "acp(10, 48, agc=0.8, hl/2)")

    def test_no_fn(self):
        x = AutoCorrelationPeriodogram(Params(fixed_normalization=True))
        self.assertEqual(x.mnemonic, "acp(10, 48, no-fn, hl/2)")

    def test_all_flags(self):
        x = AutoCorrelationPeriodogram(Params(
            averaging_length=5,
            disable_spectral_squaring=True,
            disable_smoothing=True,
            disable_automatic_gain_control=True,
            fixed_normalization=True,
        ))
        self.assertEqual(x.mnemonic, "acp(10, 48, average=5, no-sqr, no-smooth, no-agc, no-fn, hl/2)")


class TestAutoCorrelationPeriodogramValidation(unittest.TestCase):
    def test_min_period_lt_2(self):
        with self.assertRaises(ValueError):
            AutoCorrelationPeriodogram(Params(min_period=1, max_period=48, averaging_length=3))

    def test_max_period_le_min(self):
        with self.assertRaises(ValueError):
            AutoCorrelationPeriodogram(Params(min_period=10, max_period=10, averaging_length=3))

    def test_averaging_length_lt_1(self):
        with self.assertRaises(ValueError):
            AutoCorrelationPeriodogram(Params(averaging_length=-1))

    def test_agc_decay_le_0(self):
        with self.assertRaises(ValueError):
            AutoCorrelationPeriodogram(Params(automatic_gain_control_decay_factor=-0.1))

    def test_agc_decay_ge_1(self):
        with self.assertRaises(ValueError):
            AutoCorrelationPeriodogram(Params(automatic_gain_control_decay_factor=1.0))


class TestAutoCorrelationPeriodogramUpdateEntity(unittest.TestCase):
    def _prime(self, x):
        inp = _test_input()
        t = _test_time()
        for i in range(100):
            x.update(inp[i % len(inp)], t)

    def test_update_scalar(self):
        t = _test_time()
        x = AutoCorrelationPeriodogram(Params())
        self._prime(x)
        s = Scalar(time=t, value=100.0)
        out = x.update_scalar(s)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)

    def test_update_bar(self):
        t = _test_time()
        x = AutoCorrelationPeriodogram(Params())
        self._prime(x)
        b = Bar(time=t, open=100, high=100, low=100, close=100, volume=0)
        out = x.update_bar(b)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)

    def test_update_quote(self):
        t = _test_time()
        x = AutoCorrelationPeriodogram(Params())
        self._prime(x)
        q = Quote(time=t, bid_price=100, ask_price=100, bid_size=1, ask_size=1)
        out = x.update_quote(q)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)

    def test_update_trade(self):
        t = _test_time()
        x = AutoCorrelationPeriodogram(Params())
        self._prime(x)
        r = Trade(time=t, price=100, volume=1)
        out = x.update_trade(r)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)


if __name__ == '__main__':
    unittest.main()
