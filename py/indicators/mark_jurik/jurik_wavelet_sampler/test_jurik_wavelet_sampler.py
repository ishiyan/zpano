"""Tests for the Jurik Wavelet Sampler indicator."""

import math
import unittest
from datetime import datetime

from py.indicators.mark_jurik.jurik_wavelet_sampler.jurik_wavelet_sampler import JurikWaveletSampler
from py.indicators.mark_jurik.jurik_wavelet_sampler.params import JurikWaveletSamplerParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_INDEX_6,
    EXPECTED_INDEX_12,
    EXPECTED_INDEX_16,
)


TOLERANCE = 5


class TestJurikWaveletSampler(unittest.TestCase):
    """Tests for the Jurik wavelet sampler indicator."""

    def _run_test(self, index: int, expected: list[list]) -> None:
        params = JurikWaveletSamplerParams(index=index)
        indicator = JurikWaveletSampler(params)
        t = datetime(2020, 1, 1)

        for bar_idx, price in enumerate(INPUT_CLOSE):
            indicator.update_scalar(Scalar(t, price))
            cols = indicator.columns

            self.assertEqual(len(cols), index)
            for c in range(index):
                exp = expected[bar_idx][c]
                got = cols[c]
                if exp is None:
                    self.assertTrue(math.isnan(got),
                                    f"bar={bar_idx} col={c}: expected NaN, got {got}")
                else:
                    self.assertAlmostEqual(got, exp, places=TOLERANCE,
                                           msg=f"bar={bar_idx} col={c}: expected {exp}, got {got}")

    def test_index_6(self) -> None:
        self._run_test(6, EXPECTED_INDEX_6)

    def test_index_12(self) -> None:
        self._run_test(12, EXPECTED_INDEX_12)

    def test_index_16(self) -> None:
        self._run_test(16, EXPECTED_INDEX_16)

    def test_metadata(self) -> None:
        params = JurikWaveletSamplerParams(index=12)
        indicator = JurikWaveletSampler(params)
        meta = indicator.metadata()
        self.assertEqual(meta.identifier, Identifier.JURIK_WAVELET_SAMPLER)
        self.assertIn("jwav(12", meta.mnemonic)

    def test_invalid_index(self) -> None:
        with self.assertRaises(ValueError):
            JurikWaveletSampler(JurikWaveletSamplerParams(index=0))
        with self.assertRaises(ValueError):
            JurikWaveletSampler(JurikWaveletSamplerParams(index=19))

    def test_is_primed(self) -> None:
        params = JurikWaveletSamplerParams(index=6)
        indicator = JurikWaveletSampler(params)
        t = datetime(2020, 1, 1)

        # The last column (n=7, M=2) has dead_zone = 7 + 1 = 8
        # So primed after 9 bars (bar_count > 8 for all columns)
        for i in range(8):
            indicator.update_scalar(Scalar(t, INPUT_CLOSE[i]))
            self.assertFalse(indicator.is_primed())

        indicator.update_scalar(Scalar(t, INPUT_CLOSE[8]))
        self.assertTrue(indicator.is_primed())


if __name__ == '__main__':
    unittest.main()
