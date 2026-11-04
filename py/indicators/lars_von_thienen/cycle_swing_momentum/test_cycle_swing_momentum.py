import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.cycle_swing_momentum.cycle_swing_momentum import CycleSwingMomentum
from py.indicators.lars_von_thienen.cycle_swing_momentum.params import CycleSwingMomentumParams
from py.indicators.lars_von_thienen.cycle_swing_momentum import test_testdata as data


class TestCycleSwingMomentum(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'window': 50, 'fast_smoothing': 160.0, 'slow_smoothing': 1600.0, 'band_memory': 34, 'leveling': 10.0}, {'VALUE': data.EXPECTED_W50_F160_S1600_M34_Q10_VALUE, 'LOWER': data.EXPECTED_W50_F160_S1600_M34_Q10_LOWER, 'UPPER': data.EXPECTED_W50_F160_S1600_M34_Q10_UPPER, 'SCORE': data.EXPECTED_W50_F160_S1600_M34_Q10_SCORE}),
            ({'window': 6, 'fast_smoothing': 16.0, 'slow_smoothing': 160.0, 'band_memory': 5, 'leveling': 10.0}, {'VALUE': data.EXPECTED_W6_F16_S160_M5_Q10_VALUE, 'LOWER': data.EXPECTED_W6_F16_S160_M5_Q10_LOWER, 'UPPER': data.EXPECTED_W6_F16_S160_M5_Q10_UPPER, 'SCORE': data.EXPECTED_W6_F16_S160_M5_Q10_SCORE}),
            ({'window': 20, 'fast_smoothing': 160.0, 'slow_smoothing': 1600.0, 'band_memory': 20, 'leveling': 10.0}, {'VALUE': data.EXPECTED_W20_F160_S1600_M20_Q10_VALUE, 'LOWER': data.EXPECTED_W20_F160_S1600_M20_Q10_LOWER, 'UPPER': data.EXPECTED_W20_F160_S1600_M20_Q10_UPPER, 'SCORE': data.EXPECTED_W20_F160_S1600_M20_Q10_SCORE}),
            ({'window': 100, 'fast_smoothing': 160.0, 'slow_smoothing': 1600.0, 'band_memory': 34, 'leveling': 10.0}, {'VALUE': data.EXPECTED_W100_F160_S1600_M34_Q10_VALUE, 'LOWER': data.EXPECTED_W100_F160_S1600_M34_Q10_LOWER, 'UPPER': data.EXPECTED_W100_F160_S1600_M34_Q10_UPPER, 'SCORE': data.EXPECTED_W100_F160_S1600_M34_Q10_SCORE}),
            ({'window': 50, 'fast_smoothing': 0.0, 'slow_smoothing': 1600.0, 'band_memory': 34, 'leveling': 10.0}, {'VALUE': data.EXPECTED_W50_F0_S1600_M34_Q10_VALUE, 'LOWER': data.EXPECTED_W50_F0_S1600_M34_Q10_LOWER, 'UPPER': data.EXPECTED_W50_F0_S1600_M34_Q10_UPPER, 'SCORE': data.EXPECTED_W50_F0_S1600_M34_Q10_SCORE}),
            ({'window': 50, 'fast_smoothing': 1600.0, 'slow_smoothing': 16000.0, 'band_memory': 34, 'leveling': 10.0}, {'VALUE': data.EXPECTED_W50_F1600_S16000_M34_Q10_VALUE, 'LOWER': data.EXPECTED_W50_F1600_S16000_M34_Q10_LOWER, 'UPPER': data.EXPECTED_W50_F1600_S16000_M34_Q10_UPPER, 'SCORE': data.EXPECTED_W50_F1600_S16000_M34_Q10_SCORE}),
            ({'window': 50, 'fast_smoothing': 160.0, 'slow_smoothing': 1600.0, 'band_memory': 34, 'leveling': 5.0}, {'VALUE': data.EXPECTED_W50_F160_S1600_M34_Q5_VALUE, 'LOWER': data.EXPECTED_W50_F160_S1600_M34_Q5_LOWER, 'UPPER': data.EXPECTED_W50_F160_S1600_M34_Q5_UPPER, 'SCORE': data.EXPECTED_W50_F160_S1600_M34_Q5_SCORE}),
            ({'window': 252, 'fast_smoothing': 160.0, 'slow_smoothing': 1600.0, 'band_memory': 34, 'leveling': 10.0}, {'VALUE': data.EXPECTED_W252_F160_S1600_M34_Q10_VALUE, 'LOWER': data.EXPECTED_W252_F160_S1600_M34_Q10_LOWER, 'UPPER': data.EXPECTED_W252_F160_S1600_M34_Q10_UPPER, 'SCORE': data.EXPECTED_W252_F160_S1600_M34_Q10_SCORE}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = CycleSwingMomentum(CycleSwingMomentumParams(**parameters))
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
        indicator = CycleSwingMomentum()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = CycleSwingMomentum()
        self.assertTrue(indicator.metadata().mnemonic.startswith("csmo("))
        self.assertEqual(len(indicator.metadata().outputs), 5)
