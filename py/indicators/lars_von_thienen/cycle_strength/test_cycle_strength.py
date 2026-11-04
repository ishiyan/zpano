import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.cycle_strength.cycle_strength import CycleStrength
from py.indicators.lars_von_thienen.cycle_strength.params import CycleStrengthParams
from py.indicators.lars_von_thienen.cycle_strength import test_testdata as data


class TestCycleStrength(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'window': 100, 'period': 20.0, 'smoothing': 1600.0}, {'STRENGTH': data.EXPECTED_W100_P20_L1600_STRENGTH, 'AMPLITUDE': data.EXPECTED_W100_P20_L1600_AMPLITUDE, 'PHASE': data.EXPECTED_W100_P20_L1600_PHASE}),
            ({'window': 100, 'period': 20.0, 'smoothing': 0.0}, {'STRENGTH': data.EXPECTED_W100_P20_L0_STRENGTH, 'AMPLITUDE': data.EXPECTED_W100_P20_L0_AMPLITUDE, 'PHASE': data.EXPECTED_W100_P20_L0_PHASE}),
            ({'window': 60, 'period': 10.0, 'smoothing': 1600.0}, {'STRENGTH': data.EXPECTED_W60_P10_L1600_STRENGTH, 'AMPLITUDE': data.EXPECTED_W60_P10_L1600_AMPLITUDE, 'PHASE': data.EXPECTED_W60_P10_L1600_PHASE}),
            ({'window': 200, 'period': 40.0, 'smoothing': 1600.0}, {'STRENGTH': data.EXPECTED_W200_P40_L1600_STRENGTH, 'AMPLITUDE': data.EXPECTED_W200_P40_L1600_AMPLITUDE, 'PHASE': data.EXPECTED_W200_P40_L1600_PHASE}),
            ({'window': 100, 'period': 20.5, 'smoothing': 1600.0}, {'STRENGTH': data.EXPECTED_W100_P20_5_L1600_STRENGTH, 'AMPLITUDE': data.EXPECTED_W100_P20_5_L1600_AMPLITUDE, 'PHASE': data.EXPECTED_W100_P20_5_L1600_PHASE}),
            ({'window': 100, 'period': 20.0, 'smoothing': 16000.0}, {'STRENGTH': data.EXPECTED_W100_P20_L16000_STRENGTH, 'AMPLITUDE': data.EXPECTED_W100_P20_L16000_AMPLITUDE, 'PHASE': data.EXPECTED_W100_P20_L16000_PHASE}),
            ({'window': 20, 'period': 40.0, 'smoothing': 1600.0}, {'STRENGTH': data.EXPECTED_W20_P40_L1600_STRENGTH, 'AMPLITUDE': data.EXPECTED_W20_P40_L1600_AMPLITUDE, 'PHASE': data.EXPECTED_W20_P40_L1600_PHASE}),
            ({'window': 252, 'period': 20.0, 'smoothing': 1600.0}, {'STRENGTH': data.EXPECTED_W252_P20_L1600_STRENGTH, 'AMPLITUDE': data.EXPECTED_W252_P20_L1600_AMPLITUDE, 'PHASE': data.EXPECTED_W252_P20_L1600_PHASE}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = CycleStrength(CycleStrengthParams(**parameters))
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
        indicator = CycleStrength()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = CycleStrength()
        self.assertTrue(indicator.metadata().mnemonic.startswith("cstr("))
        self.assertEqual(len(indicator.metadata().outputs), 3)
