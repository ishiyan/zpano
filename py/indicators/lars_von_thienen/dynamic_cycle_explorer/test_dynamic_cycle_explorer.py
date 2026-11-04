import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.dynamic_cycle_explorer.dynamic_cycle_explorer import DynamicCycleExplorer
from py.indicators.lars_von_thienen.dynamic_cycle_explorer.params import DynamicCycleExplorerParams
from py.indicators.lars_von_thienen.dynamic_cycle_explorer import test_testdata as data


class TestDynamicCycleExplorer(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'genuine_threshold': 49.0, 'max_period': 80, 'max_segments': 5, 'min_period': 5, 'peak_distance': 2, 'phase_cycles': 2.0, 'smoothing': 1600.0, 'top_k': 3, 'window': 200}, {'VALUE': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_VALUE, 'PERIOD': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_PERIOD, 'PHASE': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_PHASE, 'ETA_TOP': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_TOP, 'ETA_BOTTOM': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_ETA_BOTTOM, 'HIGHLIGHTER': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_HIGHLIGHTER, 'HIGHLIGHTER_INDEX': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_HIGHLIGHTER_INDEX, 'FITTED_AT_HIGHLIGHTER': data.EXPECTED_W200_P05_P180_L1600_G49_F2_K3_D2_M5_FITTED_AT_HIGHLIGHTER}),
            ({'genuine_threshold': 49.0, 'max_period': 12, 'max_segments': 5, 'min_period': 12, 'peak_distance': 2, 'phase_cycles': 0.0, 'smoothing': 1600.0, 'top_k': 1, 'window': 60}, {'VALUE': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_VALUE, 'PERIOD': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_PERIOD, 'PHASE': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_PHASE, 'ETA_TOP': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_TOP, 'ETA_BOTTOM': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_ETA_BOTTOM, 'HIGHLIGHTER': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_HIGHLIGHTER, 'HIGHLIGHTER_INDEX': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_HIGHLIGHTER_INDEX, 'FITTED_AT_HIGHLIGHTER': data.EXPECTED_W60_P012_P112_L1600_G49_F0_K1_D2_M5_FITTED_AT_HIGHLIGHTER}),
            ({'genuine_threshold': 49.0, 'max_period': 30, 'max_segments': 5, 'min_period': 5, 'peak_distance': 2, 'phase_cycles': 2.0, 'smoothing': 0.0, 'top_k': 2, 'window': 100}, {'VALUE': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_VALUE, 'PERIOD': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_PERIOD, 'PHASE': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_PHASE, 'ETA_TOP': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_TOP, 'ETA_BOTTOM': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_ETA_BOTTOM, 'HIGHLIGHTER': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_HIGHLIGHTER, 'HIGHLIGHTER_INDEX': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_HIGHLIGHTER_INDEX, 'FITTED_AT_HIGHLIGHTER': data.EXPECTED_W100_P05_P130_L0_G49_F2_K2_D2_M5_FITTED_AT_HIGHLIGHTER}),
            ({'genuine_threshold': 100.0, 'max_period': 30, 'max_segments': 5, 'min_period': 5, 'peak_distance': 2, 'phase_cycles': 2.0, 'smoothing': 1600.0, 'top_k': 2, 'window': 100}, {'VALUE': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_VALUE, 'PERIOD': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_PERIOD, 'PHASE': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_PHASE, 'ETA_TOP': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_TOP, 'ETA_BOTTOM': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_ETA_BOTTOM, 'HIGHLIGHTER': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_HIGHLIGHTER, 'HIGHLIGHTER_INDEX': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_HIGHLIGHTER_INDEX, 'FITTED_AT_HIGHLIGHTER': data.EXPECTED_W100_P05_P130_L1600_G100_F2_K2_D2_M5_FITTED_AT_HIGHLIGHTER}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = DynamicCycleExplorer(DynamicCycleExplorerParams(**parameters))
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
        indicator = DynamicCycleExplorer()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = DynamicCycleExplorer()
        self.assertTrue(indicator.metadata().mnemonic.startswith("dce("))
        self.assertEqual(len(indicator.metadata().outputs), 9)
