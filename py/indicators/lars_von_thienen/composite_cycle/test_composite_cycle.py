import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.composite_cycle.composite_cycle import CompositeCycle
from py.indicators.lars_von_thienen.composite_cycle.params import CompositeCycleParams
from py.indicators.lars_von_thienen.composite_cycle import test_testdata as data


class TestCompositeCycle(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'window': 100, 'periods': [20.0, 40.0], 'smoothing': 1600.0}, {'VALUE': data.EXPECTED_W100_P20_40_L1600_VALUE, 'SLOPE': data.EXPECTED_W100_P20_40_L1600_SLOPE, 'NEXT_VALUE': data.EXPECTED_W100_P20_40_L1600_NEXT_VALUE}),
            ({'window': 100, 'periods': [20.0], 'smoothing': 0.0}, {'VALUE': data.EXPECTED_W100_P20_L0_VALUE, 'SLOPE': data.EXPECTED_W100_P20_L0_SLOPE, 'NEXT_VALUE': data.EXPECTED_W100_P20_L0_NEXT_VALUE}),
            ({'window': 100, 'periods': [20.0, 40.0], 'smoothing': 0.0}, {'VALUE': data.EXPECTED_W100_P20_40_L0_VALUE, 'SLOPE': data.EXPECTED_W100_P20_40_L0_SLOPE, 'NEXT_VALUE': data.EXPECTED_W100_P20_40_L0_NEXT_VALUE}),
            ({'window': 120, 'periods': [10.0, 20.0, 40.0], 'smoothing': 1600.0}, {'VALUE': data.EXPECTED_W120_P10_20_40_L1600_VALUE, 'SLOPE': data.EXPECTED_W120_P10_20_40_L1600_SLOPE, 'NEXT_VALUE': data.EXPECTED_W120_P10_20_40_L1600_NEXT_VALUE}),
            ({'window': 100, 'periods': [17.5, 31.0], 'smoothing': 1600.0}, {'VALUE': data.EXPECTED_W100_P17_5_31_L1600_VALUE, 'SLOPE': data.EXPECTED_W100_P17_5_31_L1600_SLOPE, 'NEXT_VALUE': data.EXPECTED_W100_P17_5_31_L1600_NEXT_VALUE}),
            ({'window': 200, 'periods': [20.0, 40.0], 'smoothing': 1600.0}, {'VALUE': data.EXPECTED_W200_P20_40_L1600_VALUE, 'SLOPE': data.EXPECTED_W200_P20_40_L1600_SLOPE, 'NEXT_VALUE': data.EXPECTED_W200_P20_40_L1600_NEXT_VALUE}),
            ({'window': 100, 'periods': [20.0, 40.0], 'smoothing': 160.0}, {'VALUE': data.EXPECTED_W100_P20_40_L160_VALUE, 'SLOPE': data.EXPECTED_W100_P20_40_L160_SLOPE, 'NEXT_VALUE': data.EXPECTED_W100_P20_40_L160_NEXT_VALUE}),
            ({'window': 253, 'periods': [20.0, 40.0], 'smoothing': 1600.0}, {'VALUE': data.EXPECTED_W253_P20_40_L1600_VALUE, 'SLOPE': data.EXPECTED_W253_P20_40_L1600_SLOPE, 'NEXT_VALUE': data.EXPECTED_W253_P20_40_L1600_NEXT_VALUE}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = CompositeCycle(CompositeCycleParams(**parameters))
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
        indicator = CompositeCycle()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = CompositeCycle()
        self.assertTrue(indicator.metadata().mnemonic.startswith("ccyc("))
        self.assertEqual(len(indicator.metadata().outputs), 3)
