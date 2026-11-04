import math
import unittest
from datetime import date, timedelta
from py.indicators.lars_von_thienen.hodrick_prescott_smoothing_and_detrending.hodrick_prescott_smoothing_and_detrending import HodrickPrescottSmoothingAndDetrending
from py.indicators.lars_von_thienen.hodrick_prescott_smoothing_and_detrending.params import HodrickPrescottSmoothingAndDetrendingParams
from py.indicators.lars_von_thienen.hodrick_prescott_smoothing_and_detrending import test_testdata as data


class TestHodrickPrescottSmoothingAndDetrending(unittest.TestCase):
    def test_prepared_reference_cases(self):
        cases = [
            ({'window': 50, 'smoothing': 1600.0}, {'TREND': data.EXPECTED_W50_L1600_TREND, 'CYCLE': data.EXPECTED_W50_L1600_CYCLE}),
            ({'window': 6, 'smoothing': 1600.0}, {'TREND': data.EXPECTED_W6_L1600_TREND, 'CYCLE': data.EXPECTED_W6_L1600_CYCLE}),
            ({'window': 20, 'smoothing': 0.0}, {'TREND': data.EXPECTED_W20_L0_TREND, 'CYCLE': data.EXPECTED_W20_L0_CYCLE}),
            ({'window': 20, 'smoothing': 1.0}, {'TREND': data.EXPECTED_W20_L1_TREND, 'CYCLE': data.EXPECTED_W20_L1_CYCLE}),
            ({'window': 50, 'smoothing': 160.0}, {'TREND': data.EXPECTED_W50_L160_TREND, 'CYCLE': data.EXPECTED_W50_L160_CYCLE}),
            ({'window': 100, 'smoothing': 1600.0}, {'TREND': data.EXPECTED_W100_L1600_TREND, 'CYCLE': data.EXPECTED_W100_L1600_CYCLE}),
            ({'window': 100, 'smoothing': 100000.0}, {'TREND': data.EXPECTED_W100_L100000_TREND, 'CYCLE': data.EXPECTED_W100_L100000_CYCLE}),
            ({'window': 252, 'smoothing': 1600.0}, {'TREND': data.EXPECTED_W252_L1600_TREND, 'CYCLE': data.EXPECTED_W252_L1600_CYCLE}),
            ({'window': 253, 'smoothing': 1600.0}, {'TREND': data.EXPECTED_W253_L1600_TREND, 'CYCLE': data.EXPECTED_W253_L1600_CYCLE}),
        ]
        for parameters, expected in cases:
            with self.subTest(parameters=parameters):
                indicator = HodrickPrescottSmoothingAndDetrending(HodrickPrescottSmoothingAndDetrendingParams(**parameters))
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
        indicator = HodrickPrescottSmoothingAndDetrending()
        indicator.update(math.nan)
        self.assertFalse(indicator.is_primed())
        indicator.reset()
        self.assertFalse(indicator.is_primed())
        with self.assertRaises(ValueError):
            indicator.update(math.inf)

    def test_registered_metadata(self):
        indicator = HodrickPrescottSmoothingAndDetrending()
        self.assertTrue(indicator.metadata().mnemonic.startswith("hpsd("))
        self.assertEqual(len(indicator.metadata().outputs), 2)
