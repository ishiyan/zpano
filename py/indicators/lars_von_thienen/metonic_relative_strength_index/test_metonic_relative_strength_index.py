import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.metonic_relative_strength_index.metonic_relative_strength_index import MetonicRelativeStrengthIndex
from py.indicators.lars_von_thienen.metonic_relative_strength_index.params import MetonicRelativeStrengthIndexParams
from py.indicators.lars_von_thienen.metonic_relative_strength_index import test_testdata as data


class TestMetonicRelativeStrengthIndex(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'length': 13, 'delay1': 6939, 'delay2': 4016, 'delay3': 2922}, {'VALUE': data.EXPECTED_L13_D16939_D24016_D32922_VALUE, 'LINEAR_RSI': data.EXPECTED_L13_D16939_D24016_D32922_LINEAR_RSI}),
            ({'length': 13, 'delay1': 19, 'delay2': 11, 'delay3': 8}, {'VALUE': data.EXPECTED_L13_D119_D211_D38_VALUE, 'LINEAR_RSI': data.EXPECTED_L13_D119_D211_D38_LINEAR_RSI}),
            ({'length': 1, 'delay1': 19, 'delay2': 11, 'delay3': 8}, {'VALUE': data.EXPECTED_L1_D119_D211_D38_VALUE, 'LINEAR_RSI': data.EXPECTED_L1_D119_D211_D38_LINEAR_RSI}),
            ({'length': 14, 'delay1': 19, 'delay2': 11, 'delay3': 8}, {'VALUE': data.EXPECTED_L14_D119_D211_D38_VALUE, 'LINEAR_RSI': data.EXPECTED_L14_D119_D211_D38_LINEAR_RSI}),
            ({'length': 50, 'delay1': 100, 'delay2': 60, 'delay3': 40}, {'VALUE': data.EXPECTED_L50_D1100_D260_D340_VALUE, 'LINEAR_RSI': data.EXPECTED_L50_D1100_D260_D340_LINEAR_RSI}),
            ({'length': 13, 'delay1': 8, 'delay2': 19, 'delay3': 11}, {'VALUE': data.EXPECTED_L13_D18_D219_D311_VALUE, 'LINEAR_RSI': data.EXPECTED_L13_D18_D219_D311_LINEAR_RSI}),
            ({'length': 13, 'delay1': 238, 'delay2': 100, 'delay3': 50}, {'VALUE': data.EXPECTED_L13_D1238_D2100_D350_VALUE, 'LINEAR_RSI': data.EXPECTED_L13_D1238_D2100_D350_LINEAR_RSI}),
            ({'length': 13, 'delay1': 239, 'delay2': 100, 'delay3': 50}, {'VALUE': data.EXPECTED_L13_D1239_D2100_D350_VALUE, 'LINEAR_RSI': data.EXPECTED_L13_D1239_D2100_D350_LINEAR_RSI}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = MetonicRelativeStrengthIndex(MetonicRelativeStrengthIndexParams(**parameters))
                for index, price in enumerate(data.INPUT_CLOSE):
                    result = indicator.update(price, time=date(2000, 1, 1) + timedelta(days=index))
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
        indicator = MetonicRelativeStrengthIndex()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = MetonicRelativeStrengthIndex()
        self.assertTrue(indicator.metadata().mnemonic.startswith("mrsi("))
        self.assertEqual(len(indicator.metadata().outputs), 2)
