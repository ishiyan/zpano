"""Tests for the Comb Band-Pass Spectrum indicator."""

import math
import unittest
import datetime

from py.indicators.john_ehlers.comb_band_pass_spectrum.comb_band_pass_spectrum import CombBandPassSpectrum
from py.indicators.john_ehlers.comb_band_pass_spectrum.params import Params
from py.indicators.john_ehlers.comb_band_pass_spectrum.output import Output
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade


TEST_TIME = datetime.datetime(2021, 4, 1, tzinfo=datetime.timezone.utc)

TEST_INPUT = [
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

TOLERANCE = 1e-12
MINMAX_TOL = 1e-10

SNAPSHOTS = [
    {
        'i': 47, 'value_min': 0, 'value_max': 0.351344643038070,
        'spots': [(0, 0.004676953354739), (9, 0.032804657174884),
                  (19, 0.298241001617233), (28, 0.269179028265479),
                  (38, 0.145584088643502)],
    },
    {
        'i': 60, 'value_min': 0, 'value_max': 0.233415131482019,
        'spots': [(0, 0.003611349016608), (9, 0.021460554913141),
                  (19, 0.159313027547382), (28, 0.219799344776603),
                  (38, 0.171081964194873)],
    },
    {
        'i': 100, 'value_min': 0, 'value_max': 0.064066532878879,
        'spots': [(0, 0.015789490651889), (9, 0.030957048077702),
                  (19, 0.004154893462836), (28, 0.042739584630981),
                  (38, 0.048070192646483)],
    },
    {
        'i': 150, 'value_min': 0, 'value_max': 0.044774991014571,
        'spots': [(0, 0.010977897375080), (9, 0.022161976000123),
                  (19, 0.005434298746720), (28, 0.041109264147755),
                  (38, 0.000028252306207)],
    },
    {
        'i': 200, 'value_min': 0, 'value_max': 0.056007975310479,
        'spots': [(0, 0.002054905622165), (9, 0.042579171063316),
                  (19, 0.003278307476910), (28, 0.033557809407585),
                  (38, 0.018072829155854)],
    },
]


class TestCombBandPassSpectrum(unittest.TestCase):

    def test_update(self):
        x = CombBandPassSpectrum(Params())
        si = 0
        for i, v in enumerate(TEST_INPUT):
            t = TEST_TIME + datetime.timedelta(minutes=i)
            h = x.update(v, t)
            self.assertIsNotNone(h)
            self.assertEqual(h.parameter_first, 10)
            self.assertEqual(h.parameter_last, 48)
            self.assertEqual(h.parameter_resolution, 1)

            if not x.is_primed():
                self.assertTrue(h.is_empty(), f"[{i}] expected empty before priming")
                continue

            self.assertEqual(len(h.values), 39, f"[{i}] values length")

            if si < len(SNAPSHOTS) and SNAPSHOTS[si]['i'] == i:
                snap = SNAPSHOTS[si]
                self.assertAlmostEqual(h.value_min, snap['value_min'], delta=MINMAX_TOL)
                self.assertAlmostEqual(h.value_max, snap['value_max'], delta=MINMAX_TOL)
                for idx, expected in snap['spots']:
                    self.assertAlmostEqual(
                        h.values[idx], expected, delta=TOLERANCE,
                        msg=f"[{i}] Values[{idx}]")
                si += 1

        self.assertEqual(si, len(SNAPSHOTS), "did not hit all snapshots")

    def test_primes_at_bar_47(self):
        x = CombBandPassSpectrum(Params())
        self.assertFalse(x.is_primed())
        primed_at = -1
        for i, v in enumerate(TEST_INPUT):
            x.update(v, TEST_TIME + datetime.timedelta(minutes=i))
            if x.is_primed() and primed_at < 0:
                primed_at = i
        self.assertEqual(primed_at, 47)

    def test_nan_input(self):
        x = CombBandPassSpectrum(Params())
        h = x.update(math.nan, TEST_TIME)
        self.assertTrue(h.is_empty())
        self.assertFalse(x.is_primed())

    def test_synthetic_sine(self):
        period = 20.0
        bars = 400
        x = CombBandPassSpectrum(Params(
            disable_spectral_dilation_compensation=True,
            disable_automatic_gain_control=True,
            fixed_normalization=True,
        ))
        last = None
        for i in range(bars):
            sample = 100 + math.sin(2 * math.pi * i / period)
            last = x.update(sample, TEST_TIME + datetime.timedelta(minutes=i))
        self.assertFalse(last.is_empty())
        peak_bin = max(range(len(last.values)), key=lambda j: last.values[j])
        expected_bin = int(period - last.parameter_first)
        self.assertEqual(peak_bin, expected_bin)

    def test_metadata(self):
        x = CombBandPassSpectrum(Params())
        md = x.metadata()
        mn = "cbps(10, 48, hl/2)"
        self.assertEqual(md.identifier, Identifier.COMB_BAND_PASS_SPECTRUM)
        self.assertEqual(md.mnemonic, mn)
        self.assertEqual(md.description, "Comb band-pass spectrum " + mn)
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].mnemonic, mn)

    def test_mnemonic_flags(self):
        cases = [
            (Params(), "cbps(10, 48, hl/2)"),
            (Params(bandwidth=0.5), "cbps(10, 48, bw=0.5, hl/2)"),
            (Params(disable_spectral_dilation_compensation=True), "cbps(10, 48, no-sdc, hl/2)"),
            (Params(disable_automatic_gain_control=True), "cbps(10, 48, no-agc, hl/2)"),
            (Params(automatic_gain_control_decay_factor=0.8), "cbps(10, 48, agc=0.8, hl/2)"),
            (Params(fixed_normalization=True), "cbps(10, 48, no-fn, hl/2)"),
            (Params(bandwidth=0.5, disable_spectral_dilation_compensation=True,
                    disable_automatic_gain_control=True, fixed_normalization=True),
             "cbps(10, 48, bw=0.5, no-sdc, no-agc, no-fn, hl/2)"),
        ]
        for p, expected_mn in cases:
            with self.subTest(mn=expected_mn):
                x = CombBandPassSpectrum(p)
                self.assertEqual(x.mnemonic, expected_mn)

    def test_validation(self):
        cases = [
            (Params(min_period=1, max_period=48, bandwidth=0.3), "MinPeriod should be >= 2"),
            (Params(min_period=10, max_period=10, bandwidth=0.3), "MaxPeriod should be > MinPeriod"),
            (Params(bandwidth=-0.1), "Bandwidth should be in (0, 1)"),
            (Params(bandwidth=1.0), "Bandwidth should be in (0, 1)"),
            (Params(automatic_gain_control_decay_factor=-0.1),
             "AutomaticGainControlDecayFactor should be in (0, 1)"),
            (Params(automatic_gain_control_decay_factor=1.0),
             "AutomaticGainControlDecayFactor should be in (0, 1)"),
        ]
        for p, msg_part in cases:
            with self.subTest(msg=msg_part):
                with self.assertRaises(ValueError) as ctx:
                    CombBandPassSpectrum(p)
                self.assertIn(msg_part, str(ctx.exception))

    def test_update_entity(self):
        for i in range(60):
            pass  # prime below

        def _prime(x):
            for i in range(60):
                x.update(TEST_INPUT[i % len(TEST_INPUT)], TEST_TIME)

        with self.subTest("scalar"):
            x = CombBandPassSpectrum(Params())
            _prime(x)
            s = Scalar(time=TEST_TIME, value=100.0)
            out = x.update_scalar(s)
            self.assertEqual(len(out), 1)

        with self.subTest("bar"):
            x = CombBandPassSpectrum(Params())
            _prime(x)
            b = Bar(time=TEST_TIME, open=100, high=100, low=100, close=100, volume=0)
            out = x.update_bar(b)
            self.assertEqual(len(out), 1)

        with self.subTest("quote"):
            x = CombBandPassSpectrum(Params())
            _prime(x)
            q = Quote(time=TEST_TIME, bid_price=100, ask_price=100, bid_size=1, ask_size=1)
            out = x.update_quote(q)
            self.assertEqual(len(out), 1)

        with self.subTest("trade"):
            x = CombBandPassSpectrum(Params())
            _prime(x)
            tr = Trade(time=TEST_TIME, price=100, volume=1)
            out = x.update_trade(tr)
            self.assertEqual(len(out), 1)

    def test_output_enum(self):
        self.assertEqual(Output.VALUE, 0)


if __name__ == '__main__':
    unittest.main()
