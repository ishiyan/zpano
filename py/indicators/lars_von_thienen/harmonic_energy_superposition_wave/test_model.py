import unittest, math
from . import test_testdata as data
from ..annual_cycle_model.model import fit_annual_values, annual_wave, AnnualWaveReader
from ..harmonic_energy_superposition_wave.model import superposition_values, SuperpositionWaveReader

class TestPreparedModels(unittest.TestCase):
    @staticmethod
    def models():
        return [fit_annual_values([(1+0.1*j)*data.INPUT_CLOSE[(i+7*j)%252] for i in range(252)],2021-lag,max_period=80) for j,lag in enumerate((8,11,19,27,30,38))]
    def test_case_0(self):
        models = self.models()
        reader = SuperpositionWaveReader(models,2021,252,"none","strict",False)
        expected = data.EXPECTED_N0_C0_I0_WAVE
        for reference in expected:
            self.assertAlmostEqual(reader.update(),reference,delta=max(1e-13,abs(reference)*5e-15))
        reader.reset()
        self.assertAlmostEqual(reader.update(),expected[0],delta=max(1e-13,abs(expected[0])*5e-15))
    def test_case_1(self):
        models = self.models()
        reader = SuperpositionWaveReader(models,2021,252,"rms","strict",False)
        expected = data.EXPECTED_N1_C0_I0_WAVE
        for reference in expected:
            self.assertAlmostEqual(reader.update(),reference,delta=max(1e-13,abs(reference)*5e-15))
        reader.reset()
        self.assertAlmostEqual(reader.update(),expected[0],delta=max(1e-13,abs(expected[0])*5e-15))
    def test_case_2(self):
        models = self.models()
        reader = SuperpositionWaveReader(models,2021,252,"rms","strict",True)
        expected = data.EXPECTED_N1_C0_I1_WAVE
        for reference in expected:
            self.assertAlmostEqual(reader.update(),reference,delta=max(1e-13,abs(reference)*5e-15))
        reader.reset()
        self.assertAlmostEqual(reader.update(),expected[0],delta=max(1e-13,abs(expected[0])*5e-15))
    def test_case_3(self):
        models = self.models()
        for model in models: model["sample_count"] = 250
        reader = SuperpositionWaveReader(models,2021,252,"rms","extend",False)
        expected = data.EXPECTED_N1_C1_I0_WAVE
        for reference in expected:
            self.assertAlmostEqual(reader.update(),reference,delta=max(1e-13,abs(reference)*5e-15))
        reader.reset()
        self.assertAlmostEqual(reader.update(),expected[0],delta=max(1e-13,abs(expected[0])*5e-15))
