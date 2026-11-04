# SPDX-License-Identifier: MPL-2.0
# Source: Detrended Rhythm Oscillator, © WhenToTrade / Lars von Thienen.
import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.detrended_rhythm_oscillators.detrended_rhythm_oscillators import DetrendedRhythmOscillators
from py.indicators.lars_von_thienen.detrended_rhythm_oscillators.params import DetrendedRhythmOscillatorsParams
from py.indicators.lars_von_thienen.detrended_rhythm_oscillators import test_testdata as data


class TestDetrendedRhythmOscillators(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'fast_length': 3, 'slow_length': 50, 'zig_period': 30, 'cycle_memory': 15, 'detrend': True}, {'VALUE': data.EXPECTED_F3_S50_Z30_M15_D1_VALUE, 'DIRECTION': data.EXPECTED_F3_S50_Z30_M15_D1_DIRECTION, 'CYCLE_LENGTH': data.EXPECTED_F3_S50_Z30_M15_D1_CYCLE_LENGTH, 'PIVOT_VALUE': data.EXPECTED_F3_S50_Z30_M15_D1_PIVOT_VALUE, 'PIVOT_OFFSET': data.EXPECTED_F3_S50_Z30_M15_D1_PIVOT_OFFSET}),
            ({'fast_length': 3, 'slow_length': 50, 'zig_period': 5, 'cycle_memory': 15, 'detrend': True}, {'VALUE': data.EXPECTED_F3_S50_Z5_M15_D1_VALUE, 'DIRECTION': data.EXPECTED_F3_S50_Z5_M15_D1_DIRECTION, 'CYCLE_LENGTH': data.EXPECTED_F3_S50_Z5_M15_D1_CYCLE_LENGTH, 'PIVOT_VALUE': data.EXPECTED_F3_S50_Z5_M15_D1_PIVOT_VALUE, 'PIVOT_OFFSET': data.EXPECTED_F3_S50_Z5_M15_D1_PIVOT_OFFSET}),
            ({'fast_length': 3, 'slow_length': 50, 'zig_period': 10, 'cycle_memory': 15, 'detrend': True}, {'VALUE': data.EXPECTED_F3_S50_Z10_M15_D1_VALUE, 'DIRECTION': data.EXPECTED_F3_S50_Z10_M15_D1_DIRECTION, 'CYCLE_LENGTH': data.EXPECTED_F3_S50_Z10_M15_D1_CYCLE_LENGTH, 'PIVOT_VALUE': data.EXPECTED_F3_S50_Z10_M15_D1_PIVOT_VALUE, 'PIVOT_OFFSET': data.EXPECTED_F3_S50_Z10_M15_D1_PIVOT_OFFSET}),
            ({'fast_length': 2, 'slow_length': 20, 'zig_period': 10, 'cycle_memory': 15, 'detrend': True}, {'VALUE': data.EXPECTED_F2_S20_Z10_M15_D1_VALUE, 'DIRECTION': data.EXPECTED_F2_S20_Z10_M15_D1_DIRECTION, 'CYCLE_LENGTH': data.EXPECTED_F2_S20_Z10_M15_D1_CYCLE_LENGTH, 'PIVOT_VALUE': data.EXPECTED_F2_S20_Z10_M15_D1_PIVOT_VALUE, 'PIVOT_OFFSET': data.EXPECTED_F2_S20_Z10_M15_D1_PIVOT_OFFSET}),
            ({'fast_length': 5, 'slow_length': 100, 'zig_period': 30, 'cycle_memory': 15, 'detrend': True}, {'VALUE': data.EXPECTED_F5_S100_Z30_M15_D1_VALUE, 'DIRECTION': data.EXPECTED_F5_S100_Z30_M15_D1_DIRECTION, 'CYCLE_LENGTH': data.EXPECTED_F5_S100_Z30_M15_D1_CYCLE_LENGTH, 'PIVOT_VALUE': data.EXPECTED_F5_S100_Z30_M15_D1_PIVOT_VALUE, 'PIVOT_OFFSET': data.EXPECTED_F5_S100_Z30_M15_D1_PIVOT_OFFSET}),
            ({'fast_length': 3, 'slow_length': 50, 'zig_period': 5, 'cycle_memory': 3, 'detrend': True}, {'VALUE': data.EXPECTED_F3_S50_Z5_M3_D1_VALUE, 'DIRECTION': data.EXPECTED_F3_S50_Z5_M3_D1_DIRECTION, 'CYCLE_LENGTH': data.EXPECTED_F3_S50_Z5_M3_D1_CYCLE_LENGTH, 'PIVOT_VALUE': data.EXPECTED_F3_S50_Z5_M3_D1_PIVOT_VALUE, 'PIVOT_OFFSET': data.EXPECTED_F3_S50_Z5_M3_D1_PIVOT_OFFSET}),
            ({'fast_length': 3, 'slow_length': 50, 'zig_period': 10, 'cycle_memory': 15, 'detrend': False}, {'VALUE': data.EXPECTED_F3_S50_Z10_M15_D0_VALUE, 'DIRECTION': data.EXPECTED_F3_S50_Z10_M15_D0_DIRECTION, 'CYCLE_LENGTH': data.EXPECTED_F3_S50_Z10_M15_D0_CYCLE_LENGTH, 'PIVOT_VALUE': data.EXPECTED_F3_S50_Z10_M15_D0_PIVOT_VALUE, 'PIVOT_OFFSET': data.EXPECTED_F3_S50_Z10_M15_D0_PIVOT_OFFSET}),
            ({'fast_length': 3, 'slow_length': 300, 'zig_period': 30, 'cycle_memory': 15, 'detrend': True}, {'VALUE': data.EXPECTED_F3_S300_Z30_M15_D1_VALUE, 'DIRECTION': data.EXPECTED_F3_S300_Z30_M15_D1_DIRECTION, 'CYCLE_LENGTH': data.EXPECTED_F3_S300_Z30_M15_D1_CYCLE_LENGTH, 'PIVOT_VALUE': data.EXPECTED_F3_S300_Z30_M15_D1_PIVOT_VALUE, 'PIVOT_OFFSET': data.EXPECTED_F3_S300_Z30_M15_D1_PIVOT_OFFSET}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = DetrendedRhythmOscillators(DetrendedRhythmOscillatorsParams(**parameters))
                for index, price in enumerate(data.INPUT_CLOSE):
                    result = indicator.update(price)
                    actual = {field.upper(): value for field, value in result._asdict().items()}
                    self.assertEqual(set(actual), set(expected))
                    for field, values in expected.items():
                        value = values[index]
                        if math.isnan(value):
                            self.assertTrue(math.isnan(actual[field]), (index, field))
                        else:
                            self.assertAlmostEqual(actual[field], value,
                                delta=max(1e-13, abs(value) * 5e-15), msg=str((index, field)))

    def test_reset_and_missing_input(self):
        indicator = DetrendedRhythmOscillators()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = DetrendedRhythmOscillators()
        self.assertTrue(indicator.metadata().mnemonic.startswith("dro("))
        self.assertEqual(len(indicator.metadata().outputs), 5)
