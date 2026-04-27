"""Tests for the Discrete Fourier Transform Spectrum indicator."""

import math
import unittest
import datetime

from ....entities.bar import Bar
from ....entities.bar_component import BarComponent
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ...core.identifier import Identifier
from .discrete_fourier_transform_spectrum import DiscreteFourierTransformSpectrum
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
        'i': 47, 'value_min': 0, 'value_max': 1,
        'spots': [
            (0, 0.092590275198666),
            (9, 0.027548278511410),
            (19, 0.157582436454944),
            (28, 0.818873586056752),
            (38, 0.994469163657834),
        ],
    },
    {
        'i': 60, 'value_min': 0, 'value_max': 1,
        'spots': [
            (0, 0.051096698360059),
            (9, 0.058377239289306),
            (19, 0.248187339190831),
            (28, 0.136096222727665),
            (38, 0.059069344951534),
        ],
    },
    {
        'i': 100, 'value_min': 0, 'value_max': 1,
        'spots': [
            (0, 0.150638336509665),
            (9, 0.046435987728045),
            (19, 0.103832850895319),
            (28, 0.402801132284104),
            (38, 1.000000000000000),
        ],
    },
    {
        'i': 150, 'value_min': 0, 'value_max': 0.5272269971142493,
        'spots': [
            (0, 0.000000000000000),
            (9, 0.091857989427651),
            (19, 0.219395988856534),
            (28, 0.516960894560452),
            (38, 0.468682020733700),
        ],
    },
    {
        'i': 200, 'value_min': 0, 'value_max': 0.6015223942655807,
        'spots': [
            (0, 0.107853213261092),
            (9, 0.164118955219278),
            (19, 0.306440501928972),
            (28, 0.569768020155262),
            (38, 0.585690371992475),
        ],
    },
]


class TestDiscreteFourierTransformSpectrumUpdate(unittest.TestCase):
    def test_update(self):
        inp = _test_input()
        t0 = _test_time()
        x = DiscreteFourierTransformSpectrum(Params())

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


class TestDiscreteFourierTransformSpectrumPrimesAtBar47(unittest.TestCase):
    def test_primes_at_47(self):
        x = DiscreteFourierTransformSpectrum(Params())
        self.assertFalse(x.is_primed())

        inp = _test_input()
        t0 = _test_time()
        primed_at = -1

        for i, val in enumerate(inp):
            x.update(val, t0 + datetime.timedelta(minutes=i))
            if x.is_primed() and primed_at < 0:
                primed_at = i

        self.assertEqual(primed_at, 47)


class TestDiscreteFourierTransformSpectrumNaN(unittest.TestCase):
    def test_nan_input(self):
        x = DiscreteFourierTransformSpectrum(Params())
        h = x.update(math.nan, _test_time())
        self.assertTrue(h.is_empty())
        self.assertFalse(x.is_primed())


class TestDiscreteFourierTransformSpectrumSyntheticSine(unittest.TestCase):
    def test_synthetic_sine(self):
        period = 16.0
        bars = 200
        t0 = _test_time()

        x = DiscreteFourierTransformSpectrum(Params(
            disable_spectral_dilation_compensation=True,
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
        self.assertEqual(peak_bin, expected_bin)


class TestDiscreteFourierTransformSpectrumMetadata(unittest.TestCase):
    def test_metadata(self):
        x = DiscreteFourierTransformSpectrum(Params())
        md = x.metadata()
        mn = "dftps(48, 10, 48, 1, hl/2)"
        self.assertEqual(md.identifier, Identifier.DISCRETE_FOURIER_TRANSFORM_SPECTRUM)
        self.assertEqual(md.mnemonic, mn)
        self.assertEqual(md.description, "Discrete Fourier transform spectrum " + mn)
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].mnemonic, mn)


class TestDiscreteFourierTransformSpectrumMnemonicFlags(unittest.TestCase):
    def test_default(self):
        x = DiscreteFourierTransformSpectrum(Params())
        self.assertEqual(x.mnemonic, "dftps(48, 10, 48, 1, hl/2)")

    def test_no_sdc(self):
        x = DiscreteFourierTransformSpectrum(Params(disable_spectral_dilation_compensation=True))
        self.assertEqual(x.mnemonic, "dftps(48, 10, 48, 1, no-sdc, hl/2)")

    def test_no_agc(self):
        x = DiscreteFourierTransformSpectrum(Params(disable_automatic_gain_control=True))
        self.assertEqual(x.mnemonic, "dftps(48, 10, 48, 1, no-agc, hl/2)")

    def test_agc_override(self):
        x = DiscreteFourierTransformSpectrum(Params(automatic_gain_control_decay_factor=0.8))
        self.assertEqual(x.mnemonic, "dftps(48, 10, 48, 1, agc=0.8, hl/2)")

    def test_no_fn(self):
        x = DiscreteFourierTransformSpectrum(Params(fixed_normalization=True))
        self.assertEqual(x.mnemonic, "dftps(48, 10, 48, 1, no-fn, hl/2)")

    def test_all_flags(self):
        x = DiscreteFourierTransformSpectrum(Params(
            disable_spectral_dilation_compensation=True,
            disable_automatic_gain_control=True,
            fixed_normalization=True,
        ))
        self.assertEqual(x.mnemonic, "dftps(48, 10, 48, 1, no-sdc, no-agc, no-fn, hl/2)")


class TestDiscreteFourierTransformSpectrumValidation(unittest.TestCase):
    def test_length_lt_2(self):
        with self.assertRaises(ValueError):
            DiscreteFourierTransformSpectrum(Params(length=1, min_period=10, max_period=48, spectrum_resolution=1))

    def test_min_period_lt_2(self):
        with self.assertRaises(ValueError):
            DiscreteFourierTransformSpectrum(Params(length=48, min_period=1, max_period=48, spectrum_resolution=1))

    def test_max_period_le_min(self):
        with self.assertRaises(ValueError):
            DiscreteFourierTransformSpectrum(Params(length=48, min_period=10, max_period=10, spectrum_resolution=1))

    def test_max_period_gt_2x_length(self):
        with self.assertRaises(ValueError):
            DiscreteFourierTransformSpectrum(Params(length=10, min_period=2, max_period=48, spectrum_resolution=1))

    def test_agc_decay_le_0(self):
        with self.assertRaises(ValueError):
            DiscreteFourierTransformSpectrum(Params(automatic_gain_control_decay_factor=-0.1))

    def test_agc_decay_ge_1(self):
        with self.assertRaises(ValueError):
            DiscreteFourierTransformSpectrum(Params(automatic_gain_control_decay_factor=1.0))


class TestDiscreteFourierTransformSpectrumUpdateEntity(unittest.TestCase):
    def _prime(self, x):
        inp = _test_input()
        t = _test_time()
        for i in range(60):
            x.update(inp[i % len(inp)], t)

    def test_update_scalar(self):
        t = _test_time()
        x = DiscreteFourierTransformSpectrum(Params())
        self._prime(x)
        s = Scalar(time=t, value=100.0)
        out = x.update_scalar(s)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)

    def test_update_bar(self):
        t = _test_time()
        x = DiscreteFourierTransformSpectrum(Params())
        self._prime(x)
        b = Bar(time=t, open=100, high=100, low=100, close=100, volume=0)
        out = x.update_bar(b)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)

    def test_update_quote(self):
        t = _test_time()
        x = DiscreteFourierTransformSpectrum(Params())
        self._prime(x)
        q = Quote(time=t, bid_price=100, ask_price=100, bid_size=1, ask_size=1)
        out = x.update_quote(q)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)

    def test_update_trade(self):
        t = _test_time()
        x = DiscreteFourierTransformSpectrum(Params())
        self._prime(x)
        r = Trade(time=t, price=100, volume=1)
        out = x.update_trade(r)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)


if __name__ == '__main__':
    unittest.main()
