import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.zero_lag_exponential_moving_average.zero_lag_exponential_moving_average import LarsVonThienenZeroLagExponentialMovingAverage
from py.indicators.lars_von_thienen.zero_lag_exponential_moving_average.params import LarsVonThienenZeroLagExponentialMovingAverageParams
from py.indicators.lars_von_thienen.zero_lag_exponential_moving_average import test_testdata as data


class TestLarsVonThienenZeroLagExponentialMovingAverage(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'length': 13}, {'VALUE': data.EXPECTED_L13_VALUE}),
            ({'length': 1}, {'VALUE': data.EXPECTED_L1_VALUE}),
            ({'length': 2}, {'VALUE': data.EXPECTED_L2_VALUE}),
            ({'length': 3}, {'VALUE': data.EXPECTED_L3_VALUE}),
            ({'length': 10}, {'VALUE': data.EXPECTED_L10_VALUE}),
            ({'length': 20}, {'VALUE': data.EXPECTED_L20_VALUE}),
            ({'length': 50}, {'VALUE': data.EXPECTED_L50_VALUE}),
            ({'length': 252}, {'VALUE': data.EXPECTED_L252_VALUE}),
            ({'length': 600}, {'VALUE': data.EXPECTED_L600_VALUE}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = LarsVonThienenZeroLagExponentialMovingAverage(LarsVonThienenZeroLagExponentialMovingAverageParams(**parameters))
                for index, price in enumerate(data.INPUT_CLOSE):
                    result = indicator.update(price)
                    actual = {"VALUE": result}
                    self.assertEqual(set(actual), set(expected))
                    for field, values in expected.items():
                        value = values[index]
                        if math.isnan(value):
                            self.assertTrue(math.isnan(actual[field]), (index, field))
                        else:
                            self.assertAlmostEqual(actual[field], value,
                                delta=max(1e-13, abs(value) * 5e-15), msg=str((index, field)))

    def test_reset_and_missing_input(self):
        indicator = LarsVonThienenZeroLagExponentialMovingAverage()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = LarsVonThienenZeroLagExponentialMovingAverage()
        self.assertTrue(indicator.metadata().mnemonic.startswith("lzlema("))
        self.assertEqual(len(indicator.metadata().outputs), 1)
