import unittest, math
from . import test_testdata as data
from ..annual_cycle_model.model import fit_annual_values, annual_wave, AnnualWaveReader
from ..harmonic_energy_superposition_wave.model import superposition_values, SuperpositionWaveReader

class TestPreparedModels(unittest.TestCase):
    def test_case_0(self):
        model = fit_annual_values(data.INPUT_CLOSE,2020,**{'candidate_pool': 10, 'genuine_threshold': 49.0, 'max_components': 7, 'max_period': 125, 'min_period': 5, 'peak_distance': 2, 'smoothing': 1600.0})
        reader = AnnualWaveReader(model)
        expected = data.EXPECTED_P05_P1125_L1600_G49_C10_K7_D2_WAVE
        self.assertAlmostEqual(model["correlation"],data.EXPECTED_P05_P1125_L1600_G49_C10_K7_D2_CORRELATION[0],places=13)
        self.assertEqual(len(model["components"]),data.EXPECTED_P05_P1125_L1600_G49_C10_K7_D2_COMPONENT_COUNT[0])
        for reference in expected:
            self.assertAlmostEqual(reader.update(),reference,delta=max(1e-13,abs(reference)*5e-15))
        reader.reset()
        self.assertAlmostEqual(reader.update(),expected[0],delta=max(1e-13,abs(expected[0])*5e-15))
    def test_case_1(self):
        model = fit_annual_values(data.INPUT_CLOSE,2020,**{'candidate_pool': 10, 'genuine_threshold': 49.0, 'max_components': 1, 'max_period': 125, 'min_period': 5, 'peak_distance': 2, 'smoothing': 1600.0})
        reader = AnnualWaveReader(model)
        expected = data.EXPECTED_P05_P1125_L1600_G49_C10_K1_D2_WAVE
        self.assertAlmostEqual(model["correlation"],data.EXPECTED_P05_P1125_L1600_G49_C10_K1_D2_CORRELATION[0],places=13)
        self.assertEqual(len(model["components"]),data.EXPECTED_P05_P1125_L1600_G49_C10_K1_D2_COMPONENT_COUNT[0])
        for reference in expected:
            self.assertAlmostEqual(reader.update(),reference,delta=max(1e-13,abs(reference)*5e-15))
        reader.reset()
        self.assertAlmostEqual(reader.update(),expected[0],delta=max(1e-13,abs(expected[0])*5e-15))
    def test_case_2(self):
        model = fit_annual_values(data.INPUT_CLOSE,2020,**{'candidate_pool': 7, 'genuine_threshold': 49.0, 'max_components': 3, 'max_period': 80, 'min_period': 5, 'peak_distance': 2, 'smoothing': 1600.0})
        reader = AnnualWaveReader(model)
        expected = data.EXPECTED_P05_P180_L1600_G49_C7_K3_D2_WAVE
        self.assertAlmostEqual(model["correlation"],data.EXPECTED_P05_P180_L1600_G49_C7_K3_D2_CORRELATION[0],places=13)
        self.assertEqual(len(model["components"]),data.EXPECTED_P05_P180_L1600_G49_C7_K3_D2_COMPONENT_COUNT[0])
        for reference in expected:
            self.assertAlmostEqual(reader.update(),reference,delta=max(1e-13,abs(reference)*5e-15))
        reader.reset()
        self.assertAlmostEqual(reader.update(),expected[0],delta=max(1e-13,abs(expected[0])*5e-15))
    def test_case_3(self):
        model = fit_annual_values(data.INPUT_CLOSE,2020,**{'candidate_pool': 10, 'genuine_threshold': 49.0, 'max_components': 7, 'max_period': 80, 'min_period': 5, 'peak_distance': 2, 'smoothing': 16000.0})
        reader = AnnualWaveReader(model)
        expected = data.EXPECTED_P05_P180_L16000_G49_C10_K7_D2_WAVE
        self.assertAlmostEqual(model["correlation"],data.EXPECTED_P05_P180_L16000_G49_C10_K7_D2_CORRELATION[0],places=13)
        self.assertEqual(len(model["components"]),data.EXPECTED_P05_P180_L16000_G49_C10_K7_D2_COMPONENT_COUNT[0])
        for reference in expected:
            self.assertAlmostEqual(reader.update(),reference,delta=max(1e-13,abs(reference)*5e-15))
        reader.reset()
        self.assertAlmostEqual(reader.update(),expected[0],delta=max(1e-13,abs(expected[0])*5e-15))
