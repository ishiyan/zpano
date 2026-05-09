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

from .test_testdata import (
    TEST_TIME,
    TEST_INPUT,
    TOLERANCE,
    MINMAX_TOL,
    SNAPSHOTS,
)


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
