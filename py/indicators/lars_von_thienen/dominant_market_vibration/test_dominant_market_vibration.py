import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.dominant_market_vibration.dominant_market_vibration import DominantMarketVibration
from py.indicators.lars_von_thienen.dominant_market_vibration.params import DominantMarketVibrationParams
from py.indicators.lars_von_thienen.dominant_market_vibration import test_testdata as data


class TestDominantMarketVibration(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'genuine_threshold': 49.0, 'max_period': 80, 'max_segments': 5, 'min_period': 5, 'peak_distance': 2, 'phase_cycles': 2.0, 'smoothing': 1600.0, 'top_k': 3, 'window': 200}, {'PERIOD': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_PERIOD, 'AMPLITUDE': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_AMPLITUDE, 'PHASE': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_PHASE, 'STRENGTH': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_STRENGTH, 'GENUINE_PERCENT': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_SEGMENTS, 'FIT_AMPLITUDE': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_FIT_AMPLITUDE, 'FIT_SAMPLES': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_FIT_SAMPLES, 'VALUE': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_VALUE, 'SLOPE': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_SLOPE, 'ETA_TOP': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_TOP, 'ETA_BOTTOM': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_BOTTOM}),
            ({'genuine_threshold': 49.0, 'max_period': 12, 'max_segments': 5, 'min_period': 12, 'peak_distance': 2, 'phase_cycles': 0.0, 'smoothing': 1600.0, 'top_k': 1, 'window': 60}, {'PERIOD': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_PERIOD, 'AMPLITUDE': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_AMPLITUDE, 'PHASE': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_PHASE, 'STRENGTH': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_STRENGTH, 'GENUINE_PERCENT': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_SEGMENTS, 'FIT_AMPLITUDE': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_FIT_AMPLITUDE, 'FIT_SAMPLES': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_FIT_SAMPLES, 'VALUE': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_VALUE, 'SLOPE': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_SLOPE, 'ETA_TOP': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_TOP, 'ETA_BOTTOM': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_BOTTOM}),
            ({'genuine_threshold': 49.0, 'max_period': 30, 'max_segments': 5, 'min_period': 5, 'peak_distance': 2, 'phase_cycles': 2.0, 'smoothing': 0.0, 'top_k': 2, 'window': 100}, {'PERIOD': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_PERIOD, 'AMPLITUDE': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_AMPLITUDE, 'PHASE': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_PHASE, 'STRENGTH': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_STRENGTH, 'GENUINE_PERCENT': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_SEGMENTS, 'FIT_AMPLITUDE': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_FIT_AMPLITUDE, 'FIT_SAMPLES': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_FIT_SAMPLES, 'VALUE': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_VALUE, 'SLOPE': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_SLOPE, 'ETA_TOP': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_TOP, 'ETA_BOTTOM': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_BOTTOM}),
            ({'genuine_threshold': 100.0, 'max_period': 30, 'max_segments': 5, 'min_period': 5, 'peak_distance': 2, 'phase_cycles': 2.0, 'smoothing': 1600.0, 'top_k': 2, 'window': 100}, {'PERIOD': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_PERIOD, 'AMPLITUDE': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_AMPLITUDE, 'PHASE': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_PHASE, 'STRENGTH': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_STRENGTH, 'GENUINE_PERCENT': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_GENUINE_PERCENT, 'SEGMENTS': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_SEGMENTS, 'FIT_AMPLITUDE': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_FIT_AMPLITUDE, 'FIT_SAMPLES': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_FIT_SAMPLES, 'VALUE': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_VALUE, 'SLOPE': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_SLOPE, 'ETA_TOP': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_TOP, 'ETA_BOTTOM': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_BOTTOM}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = DominantMarketVibration(DominantMarketVibrationParams(**parameters))
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
        indicator = DominantMarketVibration()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = DominantMarketVibration()
        self.assertTrue(indicator.metadata().mnemonic.startswith("dmv("))
        self.assertEqual(len(indicator.metadata().outputs), 9)
