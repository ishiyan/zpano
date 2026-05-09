"""Tests for the GoertzelSpectrum indicator."""

import unittest
import datetime

from py.indicators.custom.goertzel_spectrum.goertzel_spectrum import GoertzelSpectrum
from py.indicators.custom.goertzel_spectrum.params import Params, default_params
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade

from .test_testdata import (
    _test_time,
    _INPUT,
    _TOLERANCE,
    _MIN_MAX_TOL,
    _SNAPSHOTS,
)


# 252-entry TA-Lib MAMA reference series.
# Snapshots from Go reference.
class TestGoertzelSpectrumUpdate(unittest.TestCase):
    def test_update(self):
        t0 = _test_time()
        x = GoertzelSpectrum(Params())
        si = 0

        for i, val in enumerate(_INPUT):
            t = t0 + datetime.timedelta(minutes=i)
            h = x.update(val, t)
            self.assertIsNotNone(h)
            self.assertEqual(h.parameter_first, 2)
            self.assertEqual(h.parameter_last, 64)
            self.assertEqual(h.parameter_resolution, 1)

            if not x.is_primed():
                self.assertTrue(h.is_empty(),
                                f"[{i}] expected empty heatmap before priming")
                continue

            self.assertEqual(len(h.values), 63, f"[{i}] expected 63 values")

            if si < len(_SNAPSHOTS) and _SNAPSHOTS[si]['i'] == i:
                snap = _SNAPSHOTS[si]
                self.assertAlmostEqual(
                    h.value_min, snap['value_min'], delta=_MIN_MAX_TOL,
                    msg=f"[{i}] value_min")
                self.assertAlmostEqual(
                    h.value_max, snap['value_max'], delta=_MIN_MAX_TOL,
                    msg=f"[{i}] value_max")
                for idx, exp in snap['spots']:
                    self.assertAlmostEqual(
                        h.values[idx], exp, delta=_TOLERANCE,
                        msg=f"[{i}] values[{idx}]")
                si += 1

        self.assertEqual(si, len(_SNAPSHOTS), "did not hit all snapshots")


class TestGoertzelSpectrumPriming(unittest.TestCase):
    def test_primes_at_bar_63(self):
        x = GoertzelSpectrum(Params())
        self.assertFalse(x.is_primed())
        t0 = _test_time()
        primed_at = -1

        for i, val in enumerate(_INPUT):
            x.update(val, t0 + datetime.timedelta(minutes=i))
            if x.is_primed() and primed_at < 0:
                primed_at = i

        self.assertEqual(primed_at, 63)


class TestGoertzelSpectrumNaN(unittest.TestCase):
    def test_nan_input(self):
        x = GoertzelSpectrum(Params())
        h = x.update(float('nan'), _test_time())
        self.assertTrue(h.is_empty())
        self.assertFalse(x.is_primed())


class TestGoertzelSpectrumMetadata(unittest.TestCase):
    def test_metadata(self):
        x = GoertzelSpectrum(Params())
        md = x.metadata()
        mn = "gspect(64, 2, 64, 1, hl/2)"
        self.assertEqual(md.identifier, Identifier.GOERTZEL_SPECTRUM)
        self.assertEqual(md.mnemonic, mn)
        self.assertEqual(md.description, "Goertzel spectrum " + mn)
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].mnemonic, mn)


class TestGoertzelSpectrumMnemonicFlags(unittest.TestCase):
    def test_default(self):
        x = GoertzelSpectrum(Params())
        self.assertEqual(x.mnemonic, "gspect(64, 2, 64, 1, hl/2)")

    def test_first_order(self):
        x = GoertzelSpectrum(Params(is_first_order=True))
        self.assertEqual(x.mnemonic, "gspect(64, 2, 64, 1, fo, hl/2)")

    def test_no_sdc(self):
        x = GoertzelSpectrum(Params(disable_spectral_dilation_compensation=True))
        self.assertEqual(x.mnemonic, "gspect(64, 2, 64, 1, no-sdc, hl/2)")

    def test_no_agc(self):
        x = GoertzelSpectrum(Params(disable_automatic_gain_control=True))
        self.assertEqual(x.mnemonic, "gspect(64, 2, 64, 1, no-agc, hl/2)")

    def test_agc_override(self):
        x = GoertzelSpectrum(Params(automatic_gain_control_decay_factor=0.8))
        self.assertEqual(x.mnemonic, "gspect(64, 2, 64, 1, agc=0.8, hl/2)")

    def test_no_fn(self):
        x = GoertzelSpectrum(Params(fixed_normalization=True))
        self.assertEqual(x.mnemonic, "gspect(64, 2, 64, 1, no-fn, hl/2)")

    def test_all_flags(self):
        x = GoertzelSpectrum(Params(
            is_first_order=True,
            disable_spectral_dilation_compensation=True,
            disable_automatic_gain_control=True,
            fixed_normalization=True,
        ))
        self.assertEqual(x.mnemonic, "gspect(64, 2, 64, 1, fo, no-sdc, no-agc, no-fn, hl/2)")


class TestGoertzelSpectrumValidation(unittest.TestCase):
    def test_length_lt_2(self):
        with self.assertRaises(ValueError):
            GoertzelSpectrum(Params(length=1, min_period=2, max_period=64, spectrum_resolution=1))

    def test_min_period_lt_2(self):
        with self.assertRaises(ValueError):
            GoertzelSpectrum(Params(length=64, min_period=1, max_period=64, spectrum_resolution=1))

    def test_max_period_le_min(self):
        with self.assertRaises(ValueError):
            GoertzelSpectrum(Params(length=64, min_period=10, max_period=10, spectrum_resolution=1))

    def test_max_period_gt_2_length(self):
        with self.assertRaises(ValueError):
            GoertzelSpectrum(Params(length=16, min_period=2, max_period=64, spectrum_resolution=1))

    def test_agc_decay_le_0(self):
        with self.assertRaises(ValueError):
            GoertzelSpectrum(Params(automatic_gain_control_decay_factor=-0.1))

    def test_agc_decay_ge_1(self):
        with self.assertRaises(ValueError):
            GoertzelSpectrum(Params(automatic_gain_control_decay_factor=1.0))


class TestGoertzelSpectrumUpdateEntity(unittest.TestCase):
    def _prime(self, x):
        t0 = _test_time()
        for i in range(70):
            x.update(_INPUT[i % len(_INPUT)], t0)

    def test_update_scalar(self):
        x = GoertzelSpectrum(Params())
        self._prime(x)
        t = _test_time()
        result = x.update_scalar(Scalar(t, 100.0))
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, t)

    def test_update_bar(self):
        x = GoertzelSpectrum(Params())
        self._prime(x)
        t = _test_time()
        result = x.update_bar(Bar(t, 100.0, 100.0, 100.0, 100.0, 0.0))
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, t)

    def test_update_quote(self):
        x = GoertzelSpectrum(Params())
        self._prime(x)
        t = _test_time()
        result = x.update_quote(Quote(t, 100.0, 100.0, 0.0, 0.0))
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, t)

    def test_update_trade(self):
        x = GoertzelSpectrum(Params())
        self._prime(x)
        t = _test_time()
        result = x.update_trade(Trade(t, 100.0, 0.0))
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, t)


if __name__ == '__main__':
    unittest.main()
