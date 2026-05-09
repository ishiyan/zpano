"""Tests for the MaximumEntropySpectrum indicator."""

import unittest
import datetime

from py.indicators.custom.maximum_entropy_spectrum.maximum_entropy_spectrum import MaximumEntropySpectrum
from py.indicators.custom.maximum_entropy_spectrum.params import Params, default_params
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
class TestMaximumEntropySpectrumUpdate(unittest.TestCase):
    def test_update(self):
        t0 = _test_time()
        x = MaximumEntropySpectrum(Params())
        si = 0

        for i, val in enumerate(_INPUT):
            t = t0 + datetime.timedelta(minutes=i)
            h = x.update(val, t)
            self.assertIsNotNone(h)
            self.assertEqual(h.parameter_first, 2)
            self.assertEqual(h.parameter_last, 59)
            self.assertEqual(h.parameter_resolution, 1)

            if not x.is_primed():
                self.assertTrue(h.is_empty(),
                                f"[{i}] expected empty heatmap before priming")
                continue

            self.assertEqual(len(h.values), 58, f"[{i}] expected 58 values")

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


class TestMaximumEntropySpectrumPriming(unittest.TestCase):
    def test_primes_at_bar_59(self):
        x = MaximumEntropySpectrum(Params())
        self.assertFalse(x.is_primed())
        t0 = _test_time()
        primed_at = -1

        for i, val in enumerate(_INPUT):
            x.update(val, t0 + datetime.timedelta(minutes=i))
            if x.is_primed() and primed_at < 0:
                primed_at = i

        self.assertEqual(primed_at, 59)


class TestMaximumEntropySpectrumNaN(unittest.TestCase):
    def test_nan_input(self):
        x = MaximumEntropySpectrum(Params())
        h = x.update(float('nan'), _test_time())
        self.assertTrue(h.is_empty())
        self.assertFalse(x.is_primed())


class TestMaximumEntropySpectrumMetadata(unittest.TestCase):
    def test_metadata(self):
        x = MaximumEntropySpectrum(Params())
        md = x.metadata()
        mn = "mespect(60, 30, 2, 59, 1, hl/2)"
        self.assertEqual(md.identifier, Identifier.MAXIMUM_ENTROPY_SPECTRUM)
        self.assertEqual(md.mnemonic, mn)
        self.assertEqual(md.description, "Maximum entropy spectrum " + mn)
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].mnemonic, mn)


class TestMaximumEntropySpectrumMnemonicFlags(unittest.TestCase):
    def test_default(self):
        x = MaximumEntropySpectrum(Params())
        self.assertEqual(x.mnemonic, "mespect(60, 30, 2, 59, 1, hl/2)")

    def test_no_agc(self):
        x = MaximumEntropySpectrum(Params(disable_automatic_gain_control=True))
        self.assertEqual(x.mnemonic, "mespect(60, 30, 2, 59, 1, no-agc, hl/2)")

    def test_agc_override(self):
        x = MaximumEntropySpectrum(Params(automatic_gain_control_decay_factor=0.8))
        self.assertEqual(x.mnemonic, "mespect(60, 30, 2, 59, 1, agc=0.8, hl/2)")

    def test_no_fn(self):
        x = MaximumEntropySpectrum(Params(fixed_normalization=True))
        self.assertEqual(x.mnemonic, "mespect(60, 30, 2, 59, 1, no-fn, hl/2)")

    def test_all_flags(self):
        x = MaximumEntropySpectrum(Params(
            disable_automatic_gain_control=True,
            fixed_normalization=True,
        ))
        self.assertEqual(x.mnemonic, "mespect(60, 30, 2, 59, 1, no-agc, no-fn, hl/2)")


class TestMaximumEntropySpectrumValidation(unittest.TestCase):
    def test_length_lt_2(self):
        with self.assertRaises(ValueError):
            MaximumEntropySpectrum(Params(length=1, degree=1, min_period=2, max_period=4, spectrum_resolution=1))

    def test_degree_ge_length(self):
        with self.assertRaises(ValueError):
            MaximumEntropySpectrum(Params(length=4, degree=4, min_period=2, max_period=4, spectrum_resolution=1))

    def test_min_period_lt_2(self):
        with self.assertRaises(ValueError):
            MaximumEntropySpectrum(Params(length=60, degree=30, min_period=1, max_period=59, spectrum_resolution=1))

    def test_max_period_le_min(self):
        with self.assertRaises(ValueError):
            MaximumEntropySpectrum(Params(length=60, degree=30, min_period=10, max_period=10, spectrum_resolution=1))

    def test_max_period_gt_2_length(self):
        with self.assertRaises(ValueError):
            MaximumEntropySpectrum(Params(length=10, degree=5, min_period=2, max_period=59, spectrum_resolution=1))

    def test_agc_decay_le_0(self):
        with self.assertRaises(ValueError):
            MaximumEntropySpectrum(Params(automatic_gain_control_decay_factor=-0.1))

    def test_agc_decay_ge_1(self):
        with self.assertRaises(ValueError):
            MaximumEntropySpectrum(Params(automatic_gain_control_decay_factor=1.0))


class TestMaximumEntropySpectrumUpdateEntity(unittest.TestCase):
    def _prime(self, x):
        t0 = _test_time()
        for i in range(70):
            x.update(_INPUT[i % len(_INPUT)], t0)

    def test_update_scalar(self):
        x = MaximumEntropySpectrum(Params())
        self._prime(x)
        t = _test_time()
        result = x.update_scalar(Scalar(t, 100.0))
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, t)

    def test_update_bar(self):
        x = MaximumEntropySpectrum(Params())
        self._prime(x)
        t = _test_time()
        result = x.update_bar(Bar(t, 100.0, 100.0, 100.0, 100.0, 0.0))
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, t)

    def test_update_quote(self):
        x = MaximumEntropySpectrum(Params())
        self._prime(x)
        t = _test_time()
        result = x.update_quote(Quote(t, 100.0, 100.0, 0.0, 0.0))
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, t)

    def test_update_trade(self):
        x = MaximumEntropySpectrum(Params())
        self._prime(x)
        t = _test_time()
        result = x.update_trade(Trade(t, 100.0, 0.0))
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, t)


if __name__ == '__main__':
    unittest.main()
