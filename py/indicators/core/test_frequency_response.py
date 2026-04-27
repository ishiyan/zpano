"""Tests for frequency_response module."""

import math
import sys
import unittest

from py.indicators.core.frequency_response import (
    Component,
    FrequencyResponse,
    _direct_real_fft,
    _is_valid_signal_length,
    _prepare_filtered_signal,
    _prepare_frequency_domain,
    calculate,
)


class IdentityFilter:
    """Identity filter that returns the sample unchanged."""

    def metadata(self):
        class M:
            mnemonic = 'identity'
        return M()

    def update(self, sample: float) -> float:
        return sample


class TestFrequencyResponse(unittest.TestCase):

    def test_validate_signal_length(self):
        valid = {4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192}
        for i in range(-1, 8199):
            exp = i in valid
            act = _is_valid_signal_length(i)
            self.assertEqual(exp, act, f"isValidSignalLength({i})")

    def test_prepare_frequency_domain(self):
        length = 7
        expected = [1/7, 2/7, 3/7, 4/7, 5/7, 6/7, 7/7]
        actual = [0.0] * 7
        _prepare_frequency_domain(length, actual)
        for i in range(length):
            self.assertAlmostEqual(expected[i], actual[i], places=15,
                                   msg=f"freq[{i}]")

    def test_prepare_filtered_signal(self):
        length = 7
        warmup = 5
        expected = [1000.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0]
        f = IdentityFilter()
        actual = _prepare_filtered_signal(length, f, warmup)
        for i in range(length):
            self.assertAlmostEqual(expected[i], actual[i], places=15,
                                   msg=f"signal[{i}]")

    def test_fft_constant_signal(self):
        expected = [16.0] + [0.0] * 15
        actual = [1.0] * 16
        _direct_real_fft(actual)
        for i in range(16):
            self.assertAlmostEqual(expected[i], actual[i],
                                   delta=sys.float_info.min,
                                   msg=f"FFT[{i}]")

    def test_calculate_identity(self):
        f = IdentityFilter()
        fr = calculate(512, f, 128)
        self.assertIsNotNone(fr)
        self.assertEqual('identity', fr.label)
        self.assertEqual(255, len(fr.normalized_frequency))

    def test_calculate_invalid_length(self):
        f = IdentityFilter()
        with self.assertRaises(ValueError):
            calculate(100, f, 0)


if __name__ == '__main__':
    unittest.main()
