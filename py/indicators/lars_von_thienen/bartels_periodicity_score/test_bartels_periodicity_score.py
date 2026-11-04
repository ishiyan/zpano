import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.bartels_periodicity_score.bartels_periodicity_score import BartelsPeriodicityScore
from py.indicators.lars_von_thienen.bartels_periodicity_score.params import BartelsPeriodicityScoreParams
from py.indicators.lars_von_thienen.bartels_periodicity_score import test_testdata as data


class TestBartelsPeriodicityScore(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'period': 20, 'segments': 5, 'smoothing': 0.0}, {'SCORE': data.EXPECTED_P20_M5_L0_SCORE, 'GENUINE_PERCENT': data.EXPECTED_P20_M5_L0_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_P20_M5_L0_SEGMENTS}),
            ({'period': 3, 'segments': 2, 'smoothing': 0.0}, {'SCORE': data.EXPECTED_P3_M2_L0_SCORE, 'GENUINE_PERCENT': data.EXPECTED_P3_M2_L0_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_P3_M2_L0_SEGMENTS}),
            ({'period': 10, 'segments': 5, 'smoothing': 0.0}, {'SCORE': data.EXPECTED_P10_M5_L0_SCORE, 'GENUINE_PERCENT': data.EXPECTED_P10_M5_L0_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_P10_M5_L0_SEGMENTS}),
            ({'period': 20, 'segments': 2, 'smoothing': 0.0}, {'SCORE': data.EXPECTED_P20_M2_L0_SCORE, 'GENUINE_PERCENT': data.EXPECTED_P20_M2_L0_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_P20_M2_L0_SEGMENTS}),
            ({'period': 20, 'segments': 5, 'smoothing': 1600.0}, {'SCORE': data.EXPECTED_P20_M5_L1600_SCORE, 'GENUINE_PERCENT': data.EXPECTED_P20_M5_L1600_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_P20_M5_L1600_SEGMENTS}),
            ({'period': 40, 'segments': 5, 'smoothing': 0.0}, {'SCORE': data.EXPECTED_P40_M5_L0_SCORE, 'GENUINE_PERCENT': data.EXPECTED_P40_M5_L0_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_P40_M5_L0_SEGMENTS}),
            ({'period': 42, 'segments': 6, 'smoothing': 0.0}, {'SCORE': data.EXPECTED_P42_M6_L0_SCORE, 'GENUINE_PERCENT': data.EXPECTED_P42_M6_L0_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_P42_M6_L0_SEGMENTS}),
            ({'period': 60, 'segments': 5, 'smoothing': 0.0}, {'SCORE': data.EXPECTED_P60_M5_L0_SCORE, 'GENUINE_PERCENT': data.EXPECTED_P60_M5_L0_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_P60_M5_L0_SEGMENTS}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = BartelsPeriodicityScore(BartelsPeriodicityScoreParams(**parameters))
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
        indicator = BartelsPeriodicityScore()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = BartelsPeriodicityScore()
        self.assertTrue(indicator.metadata().mnemonic.startswith("bps("))
        self.assertEqual(len(indicator.metadata().outputs), 3)
