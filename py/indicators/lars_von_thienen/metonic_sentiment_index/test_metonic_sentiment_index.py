import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.metonic_sentiment_index.metonic_sentiment_index import MetonicSentimentIndex
from py.indicators.lars_von_thienen.metonic_sentiment_index.params import MetonicSentimentIndexParams
from py.indicators.lars_von_thienen.metonic_sentiment_index import test_testdata as data


class TestMetonicSentimentIndex(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'length': 13, 'delay1': 6939, 'delay2': 4016, 'delay3': 2922, 'mode': 0}, {'VALUE': data.EXPECTED_L13_D16939_D24016_D32922_M0_VALUE, 'DAILY_SENTIMENT': data.EXPECTED_L13_D16939_D24016_D32922_M0_DAILY_SENTIMENT, 'SMOOTHED_CLOSE': data.EXPECTED_L13_D16939_D24016_D32922_M0_SMOOTHED_CLOSE, 'SMOOTHED_OPEN': data.EXPECTED_L13_D16939_D24016_D32922_M0_SMOOTHED_OPEN}),
            ({'length': 13, 'delay1': 6939, 'delay2': 4016, 'delay3': 2922, 'mode': 1}, {'VALUE': data.EXPECTED_L13_D16939_D24016_D32922_M1_VALUE, 'DAILY_SENTIMENT': data.EXPECTED_L13_D16939_D24016_D32922_M1_DAILY_SENTIMENT, 'SMOOTHED_CLOSE': data.EXPECTED_L13_D16939_D24016_D32922_M1_SMOOTHED_CLOSE, 'SMOOTHED_OPEN': data.EXPECTED_L13_D16939_D24016_D32922_M1_SMOOTHED_OPEN}),
            ({'length': 13, 'delay1': 19, 'delay2': 11, 'delay3': 8, 'mode': 0}, {'VALUE': data.EXPECTED_L13_D119_D211_D38_M0_VALUE, 'DAILY_SENTIMENT': data.EXPECTED_L13_D119_D211_D38_M0_DAILY_SENTIMENT, 'SMOOTHED_CLOSE': data.EXPECTED_L13_D119_D211_D38_M0_SMOOTHED_CLOSE, 'SMOOTHED_OPEN': data.EXPECTED_L13_D119_D211_D38_M0_SMOOTHED_OPEN}),
            ({'length': 13, 'delay1': 19, 'delay2': 11, 'delay3': 8, 'mode': 1}, {'VALUE': data.EXPECTED_L13_D119_D211_D38_M1_VALUE, 'DAILY_SENTIMENT': data.EXPECTED_L13_D119_D211_D38_M1_DAILY_SENTIMENT, 'SMOOTHED_CLOSE': data.EXPECTED_L13_D119_D211_D38_M1_SMOOTHED_CLOSE, 'SMOOTHED_OPEN': data.EXPECTED_L13_D119_D211_D38_M1_SMOOTHED_OPEN}),
            ({'length': 1, 'delay1': 19, 'delay2': 11, 'delay3': 8, 'mode': 0}, {'VALUE': data.EXPECTED_L1_D119_D211_D38_M0_VALUE, 'DAILY_SENTIMENT': data.EXPECTED_L1_D119_D211_D38_M0_DAILY_SENTIMENT, 'SMOOTHED_CLOSE': data.EXPECTED_L1_D119_D211_D38_M0_SMOOTHED_CLOSE, 'SMOOTHED_OPEN': data.EXPECTED_L1_D119_D211_D38_M0_SMOOTHED_OPEN}),
            ({'length': 10, 'delay1': 19, 'delay2': 11, 'delay3': 8, 'mode': 1}, {'VALUE': data.EXPECTED_L10_D119_D211_D38_M1_VALUE, 'DAILY_SENTIMENT': data.EXPECTED_L10_D119_D211_D38_M1_DAILY_SENTIMENT, 'SMOOTHED_CLOSE': data.EXPECTED_L10_D119_D211_D38_M1_SMOOTHED_CLOSE, 'SMOOTHED_OPEN': data.EXPECTED_L10_D119_D211_D38_M1_SMOOTHED_OPEN}),
            ({'length': 20, 'delay1': 100, 'delay2': 60, 'delay3': 40, 'mode': 0}, {'VALUE': data.EXPECTED_L20_D1100_D260_D340_M0_VALUE, 'DAILY_SENTIMENT': data.EXPECTED_L20_D1100_D260_D340_M0_DAILY_SENTIMENT, 'SMOOTHED_CLOSE': data.EXPECTED_L20_D1100_D260_D340_M0_SMOOTHED_CLOSE, 'SMOOTHED_OPEN': data.EXPECTED_L20_D1100_D260_D340_M0_SMOOTHED_OPEN}),
            ({'length': 13, 'delay1': 8, 'delay2': 19, 'delay3': 11, 'mode': 1}, {'VALUE': data.EXPECTED_L13_D18_D219_D311_M1_VALUE, 'DAILY_SENTIMENT': data.EXPECTED_L13_D18_D219_D311_M1_DAILY_SENTIMENT, 'SMOOTHED_CLOSE': data.EXPECTED_L13_D18_D219_D311_M1_SMOOTHED_CLOSE, 'SMOOTHED_OPEN': data.EXPECTED_L13_D18_D219_D311_M1_SMOOTHED_OPEN}),
            ({'length': 1, 'delay1': 251, 'delay2': 100, 'delay3': 50, 'mode': 0}, {'VALUE': data.EXPECTED_L1_D1251_D2100_D350_M0_VALUE, 'DAILY_SENTIMENT': data.EXPECTED_L1_D1251_D2100_D350_M0_DAILY_SENTIMENT, 'SMOOTHED_CLOSE': data.EXPECTED_L1_D1251_D2100_D350_M0_SMOOTHED_CLOSE, 'SMOOTHED_OPEN': data.EXPECTED_L1_D1251_D2100_D350_M0_SMOOTHED_OPEN}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = MetonicSentimentIndex(MetonicSentimentIndexParams(**parameters))
                for index, price in enumerate(data.INPUT_CLOSE):
                    result = indicator.update(price, time=date(2000, 1, 1) + timedelta(days=index), open_price=data.INPUT_OPEN[index])
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
        indicator = MetonicSentimentIndex()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = MetonicSentimentIndex()
        self.assertTrue(indicator.metadata().mnemonic.startswith("msent("))
        self.assertEqual(len(indicator.metadata().outputs), 4)
