import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.phasing_oscillator.phasing_oscillator import PhasingOscillator
from py.indicators.lars_von_thienen.phasing_oscillator.params import PhasingOscillatorParams
from py.indicators.lars_von_thienen.phasing_oscillator import test_testdata as data


class TestPhasingOscillator(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'genuine_threshold': 49.0, 'long_max': 80, 'long_min': 31, 'phase_cycles': 2.0, 'short_max': 30, 'short_min': 5, 'smoothing': 1600.0, 'window': 200}, {'VALUE': data.EXPECTED_W200_S05_S130_L031_L180_HP1600_G49_F2_VALUE, 'SHORT_SCORE': data.EXPECTED_W200_S05_S130_L031_L180_HP1600_G49_F2_SHORT_SCORE, 'LONG_SCORE': data.EXPECTED_W200_S05_S130_L031_L180_HP1600_G49_F2_LONG_SCORE, 'SHORT_PERIOD': data.EXPECTED_W200_S05_S130_L031_L180_HP1600_G49_F2_SHORT_PERIOD, 'LONG_PERIOD': data.EXPECTED_W200_S05_S130_L031_L180_HP1600_G49_F2_LONG_PERIOD}),
            ({'genuine_threshold': 49.0, 'long_max': 40, 'long_min': 16, 'phase_cycles': 0.0, 'short_max': 15, 'short_min': 5, 'smoothing': 1600.0, 'window': 100}, {'VALUE': data.EXPECTED_W100_S05_S115_L016_L140_HP1600_G49_F0_VALUE, 'SHORT_SCORE': data.EXPECTED_W100_S05_S115_L016_L140_HP1600_G49_F0_SHORT_SCORE, 'LONG_SCORE': data.EXPECTED_W100_S05_S115_L016_L140_HP1600_G49_F0_LONG_SCORE, 'SHORT_PERIOD': data.EXPECTED_W100_S05_S115_L016_L140_HP1600_G49_F0_SHORT_PERIOD, 'LONG_PERIOD': data.EXPECTED_W100_S05_S115_L016_L140_HP1600_G49_F0_LONG_PERIOD}),
            ({'genuine_threshold': 49.0, 'long_max': 40, 'long_min': 16, 'phase_cycles': 2.0, 'short_max': 15, 'short_min': 5, 'smoothing': 0.0, 'window': 100}, {'VALUE': data.EXPECTED_W100_S05_S115_L016_L140_HP0_G49_F2_VALUE, 'SHORT_SCORE': data.EXPECTED_W100_S05_S115_L016_L140_HP0_G49_F2_SHORT_SCORE, 'LONG_SCORE': data.EXPECTED_W100_S05_S115_L016_L140_HP0_G49_F2_LONG_SCORE, 'SHORT_PERIOD': data.EXPECTED_W100_S05_S115_L016_L140_HP0_G49_F2_SHORT_PERIOD, 'LONG_PERIOD': data.EXPECTED_W100_S05_S115_L016_L140_HP0_G49_F2_LONG_PERIOD}),
            ({'genuine_threshold': 100.0, 'long_max': 80, 'long_min': 31, 'phase_cycles': 2.0, 'short_max': 30, 'short_min': 5, 'smoothing': 1600.0, 'window': 200}, {'VALUE': data.EXPECTED_W200_S05_S130_L031_L180_HP1600_G100_F2_VALUE, 'SHORT_SCORE': data.EXPECTED_W200_S05_S130_L031_L180_HP1600_G100_F2_SHORT_SCORE, 'LONG_SCORE': data.EXPECTED_W200_S05_S130_L031_L180_HP1600_G100_F2_LONG_SCORE, 'SHORT_PERIOD': data.EXPECTED_W200_S05_S130_L031_L180_HP1600_G100_F2_SHORT_PERIOD, 'LONG_PERIOD': data.EXPECTED_W200_S05_S130_L031_L180_HP1600_G100_F2_LONG_PERIOD}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = PhasingOscillator(PhasingOscillatorParams(**parameters))
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
        indicator = PhasingOscillator()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = PhasingOscillator()
        self.assertTrue(indicator.metadata().mnemonic.startswith("phosc("))
        self.assertEqual(len(indicator.metadata().outputs), 5)
