import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.cyclic_smoothed_relative_strength_index.cyclic_smoothed_relative_strength_index import CyclicSmoothedRelativeStrengthIndex
from py.indicators.lars_von_thienen.cyclic_smoothed_relative_strength_index.params import CyclicSmoothedRelativeStrengthIndexParams
from py.indicators.lars_von_thienen.cyclic_smoothed_relative_strength_index import test_testdata as data


class TestCyclicSmoothedRelativeStrengthIndex(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'dominant_cycle': 20, 'vibration': 10, 'leveling': 10.0}, {'VALUE': data.EXPECTED_D20_V10_Q10_VALUE, 'LOWER': data.EXPECTED_D20_V10_Q10_LOWER, 'UPPER': data.EXPECTED_D20_V10_Q10_UPPER}),
            ({'dominant_cycle': 10, 'vibration': 10, 'leveling': 10.0}, {'VALUE': data.EXPECTED_D10_V10_Q10_VALUE, 'LOWER': data.EXPECTED_D10_V10_Q10_LOWER, 'UPPER': data.EXPECTED_D10_V10_Q10_UPPER}),
            ({'dominant_cycle': 21, 'vibration': 10, 'leveling': 10.0}, {'VALUE': data.EXPECTED_D21_V10_Q10_VALUE, 'LOWER': data.EXPECTED_D21_V10_Q10_LOWER, 'UPPER': data.EXPECTED_D21_V10_Q10_UPPER}),
            ({'dominant_cycle': 40, 'vibration': 10, 'leveling': 10.0}, {'VALUE': data.EXPECTED_D40_V10_Q10_VALUE, 'LOWER': data.EXPECTED_D40_V10_Q10_LOWER, 'UPPER': data.EXPECTED_D40_V10_Q10_UPPER}),
            ({'dominant_cycle': 100, 'vibration': 10, 'leveling': 10.0}, {'VALUE': data.EXPECTED_D100_V10_Q10_VALUE, 'LOWER': data.EXPECTED_D100_V10_Q10_LOWER, 'UPPER': data.EXPECTED_D100_V10_Q10_UPPER}),
            ({'dominant_cycle': 20, 'vibration': 3, 'leveling': 10.0}, {'VALUE': data.EXPECTED_D20_V3_Q10_VALUE, 'LOWER': data.EXPECTED_D20_V3_Q10_LOWER, 'UPPER': data.EXPECTED_D20_V3_Q10_UPPER}),
            ({'dominant_cycle': 20, 'vibration': 6, 'leveling': 10.0}, {'VALUE': data.EXPECTED_D20_V6_Q10_VALUE, 'LOWER': data.EXPECTED_D20_V6_Q10_LOWER, 'UPPER': data.EXPECTED_D20_V6_Q10_UPPER}),
            ({'dominant_cycle': 20, 'vibration': 10, 'leveling': 5.0}, {'VALUE': data.EXPECTED_D20_V10_Q5_VALUE, 'LOWER': data.EXPECTED_D20_V10_Q5_LOWER, 'UPPER': data.EXPECTED_D20_V10_Q5_UPPER}),
            ({'dominant_cycle': 20, 'vibration': 10, 'leveling': 20.0}, {'VALUE': data.EXPECTED_D20_V10_Q20_VALUE, 'LOWER': data.EXPECTED_D20_V10_Q20_LOWER, 'UPPER': data.EXPECTED_D20_V10_Q20_UPPER}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = CyclicSmoothedRelativeStrengthIndex(CyclicSmoothedRelativeStrengthIndexParams(**parameters))
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
        indicator = CyclicSmoothedRelativeStrengthIndex()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = CyclicSmoothedRelativeStrengthIndex()
        self.assertTrue(indicator.metadata().mnemonic.startswith("crsi("))
        self.assertEqual(len(indicator.metadata().outputs), 4)
