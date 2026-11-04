import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.cycle_polarity.cycle_polarity import CyclePolarity
from py.indicators.lars_von_thienen.cycle_polarity.params import CyclePolarityParams
from py.indicators.lars_von_thienen.cycle_polarity import test_testdata as data


class TestCyclePolarity(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'window': 100, 'short_period': 20.0, 'long_period': 40.0, 'smoothing': 1600.0, 'tolerance': 0.0}, {'POLARITY': data.EXPECTED_W100_S20_L40_HP1600_T0_POLARITY, 'SHORT_VALUE': data.EXPECTED_W100_S20_L40_HP1600_T0_SHORT_VALUE, 'LONG_VALUE': data.EXPECTED_W100_S20_L40_HP1600_T0_LONG_VALUE, 'SHORT_SLOPE': data.EXPECTED_W100_S20_L40_HP1600_T0_SHORT_SLOPE, 'LONG_SLOPE': data.EXPECTED_W100_S20_L40_HP1600_T0_LONG_SLOPE}),
            ({'window': 100, 'short_period': 20.0, 'long_period': 40.0, 'smoothing': 0.0, 'tolerance': 0.0}, {'POLARITY': data.EXPECTED_W100_S20_L40_HP0_T0_POLARITY, 'SHORT_VALUE': data.EXPECTED_W100_S20_L40_HP0_T0_SHORT_VALUE, 'LONG_VALUE': data.EXPECTED_W100_S20_L40_HP0_T0_LONG_VALUE, 'SHORT_SLOPE': data.EXPECTED_W100_S20_L40_HP0_T0_SHORT_SLOPE, 'LONG_SLOPE': data.EXPECTED_W100_S20_L40_HP0_T0_LONG_SLOPE}),
            ({'window': 60, 'short_period': 10.0, 'long_period': 20.0, 'smoothing': 1600.0, 'tolerance': 0.0}, {'POLARITY': data.EXPECTED_W60_S10_L20_HP1600_T0_POLARITY, 'SHORT_VALUE': data.EXPECTED_W60_S10_L20_HP1600_T0_SHORT_VALUE, 'LONG_VALUE': data.EXPECTED_W60_S10_L20_HP1600_T0_LONG_VALUE, 'SHORT_SLOPE': data.EXPECTED_W60_S10_L20_HP1600_T0_SHORT_SLOPE, 'LONG_SLOPE': data.EXPECTED_W60_S10_L20_HP1600_T0_LONG_SLOPE}),
            ({'window': 200, 'short_period': 40.0, 'long_period': 80.0, 'smoothing': 1600.0, 'tolerance': 0.0}, {'POLARITY': data.EXPECTED_W200_S40_L80_HP1600_T0_POLARITY, 'SHORT_VALUE': data.EXPECTED_W200_S40_L80_HP1600_T0_SHORT_VALUE, 'LONG_VALUE': data.EXPECTED_W200_S40_L80_HP1600_T0_LONG_VALUE, 'SHORT_SLOPE': data.EXPECTED_W200_S40_L80_HP1600_T0_SHORT_SLOPE, 'LONG_SLOPE': data.EXPECTED_W200_S40_L80_HP1600_T0_LONG_SLOPE}),
            ({'window': 100, 'short_period': 17.5, 'long_period': 31.0, 'smoothing': 1600.0, 'tolerance': 0.0}, {'POLARITY': data.EXPECTED_W100_S17_5_L31_HP1600_T0_POLARITY, 'SHORT_VALUE': data.EXPECTED_W100_S17_5_L31_HP1600_T0_SHORT_VALUE, 'LONG_VALUE': data.EXPECTED_W100_S17_5_L31_HP1600_T0_LONG_VALUE, 'SHORT_SLOPE': data.EXPECTED_W100_S17_5_L31_HP1600_T0_SHORT_SLOPE, 'LONG_SLOPE': data.EXPECTED_W100_S17_5_L31_HP1600_T0_LONG_SLOPE}),
            ({'window': 100, 'short_period': 20.0, 'long_period': 40.0, 'smoothing': 1600.0, 'tolerance': 0.1}, {'POLARITY': data.EXPECTED_W100_S20_L40_HP1600_T0_1_POLARITY, 'SHORT_VALUE': data.EXPECTED_W100_S20_L40_HP1600_T0_1_SHORT_VALUE, 'LONG_VALUE': data.EXPECTED_W100_S20_L40_HP1600_T0_1_LONG_VALUE, 'SHORT_SLOPE': data.EXPECTED_W100_S20_L40_HP1600_T0_1_SHORT_SLOPE, 'LONG_SLOPE': data.EXPECTED_W100_S20_L40_HP1600_T0_1_LONG_SLOPE}),
            ({'window': 100, 'short_period': 20.0, 'long_period': 40.0, 'smoothing': 1600.0, 'tolerance': 1000.0}, {'POLARITY': data.EXPECTED_W100_S20_L40_HP1600_T1000_POLARITY, 'SHORT_VALUE': data.EXPECTED_W100_S20_L40_HP1600_T1000_SHORT_VALUE, 'LONG_VALUE': data.EXPECTED_W100_S20_L40_HP1600_T1000_LONG_VALUE, 'SHORT_SLOPE': data.EXPECTED_W100_S20_L40_HP1600_T1000_SHORT_SLOPE, 'LONG_SLOPE': data.EXPECTED_W100_S20_L40_HP1600_T1000_LONG_SLOPE}),
            ({'window': 253, 'short_period': 20.0, 'long_period': 40.0, 'smoothing': 1600.0, 'tolerance': 0.0}, {'POLARITY': data.EXPECTED_W253_S20_L40_HP1600_T0_POLARITY, 'SHORT_VALUE': data.EXPECTED_W253_S20_L40_HP1600_T0_SHORT_VALUE, 'LONG_VALUE': data.EXPECTED_W253_S20_L40_HP1600_T0_LONG_VALUE, 'SHORT_SLOPE': data.EXPECTED_W253_S20_L40_HP1600_T0_SHORT_SLOPE, 'LONG_SLOPE': data.EXPECTED_W253_S20_L40_HP1600_T0_LONG_SLOPE}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = CyclePolarity(CyclePolarityParams(**parameters))
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
        indicator = CyclePolarity()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = CyclePolarity()
        self.assertTrue(indicator.metadata().mnemonic.startswith("cpol("))
        self.assertEqual(len(indicator.metadata().outputs), 5)
