import math
import unittest
from datetime import datetime

from py.indicators.don_mak.instantaneous_sine_wave_period.instantaneous_sine_wave_period import InstantaneousSineWavePeriod
from py.indicators.don_mak.instantaneous_sine_wave_period.params import InstantaneousSineWavePeriodParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_S0_PERIOD, EXPECTED_S0_OMEGA, EXPECTED_S0_VELOCITY, EXPECTED_S0_ACCELERATION,
    EXPECTED_S3_PERIOD, EXPECTED_S3_OMEGA, EXPECTED_S3_VELOCITY, EXPECTED_S3_ACCELERATION,
    EXPECTED_S6_PERIOD, EXPECTED_S6_OMEGA, EXPECTED_S6_VELOCITY, EXPECTED_S6_ACCELERATION,
    EXPECTED_S12_PERIOD, EXPECTED_S12_OMEGA, EXPECTED_S12_VELOCITY, EXPECTED_S12_ACCELERATION,
)

TOLERANCE = 1e-9

# (smoothing, period, omega, velocity, acceleration)
COMBOS = [
    (0, EXPECTED_S0_PERIOD, EXPECTED_S0_OMEGA, EXPECTED_S0_VELOCITY, EXPECTED_S0_ACCELERATION),
    (3, EXPECTED_S3_PERIOD, EXPECTED_S3_OMEGA, EXPECTED_S3_VELOCITY, EXPECTED_S3_ACCELERATION),
    (6, EXPECTED_S6_PERIOD, EXPECTED_S6_OMEGA, EXPECTED_S6_VELOCITY, EXPECTED_S6_ACCELERATION),
    (12, EXPECTED_S12_PERIOD, EXPECTED_S12_OMEGA, EXPECTED_S12_VELOCITY, EXPECTED_S12_ACCELERATION),
]


def _assert_series(test, actual, expected, label):
    test.assertEqual(len(actual), len(expected), f"{label}: length mismatch")
    for i in range(len(expected)):
        if math.isnan(expected[i]):
            test.assertTrue(math.isnan(actual[i]), f"{label}[{i}]: expected NaN, got {actual[i]}")
        else:
            test.assertAlmostEqual(actual[i], expected[i], delta=TOLERANCE,
                                   msg=f"{label}[{i}]: expected {expected[i]}, got {actual[i]}")


class TestInstantaneousSineWavePeriodData(unittest.TestCase):
    """Test ISWP against the reference test data for all smoothing combos."""

    def test_combos(self):
        for smoothing, exp_period, exp_omega, exp_velocity, exp_acceleration in COMBOS:
            with self.subTest(smoothing=smoothing):
                ind = InstantaneousSineWavePeriod(InstantaneousSineWavePeriodParams(smoothing=smoothing))
                periods, omegas, velocities, accelerations = [], [], [], []
                for c in INPUT_CLOSE:
                    period, omega, velocity, acceleration, _amp, _phase, _dc = ind.update(c)
                    periods.append(period)
                    omegas.append(omega)
                    velocities.append(velocity)
                    accelerations.append(acceleration)
                _assert_series(self, periods, exp_period, f"period(S{smoothing})")
                _assert_series(self, omegas, exp_omega, f"omega(S{smoothing})")
                _assert_series(self, velocities, exp_velocity, f"velocity(S{smoothing})")
                _assert_series(self, accelerations, exp_acceleration, f"acceleration(S{smoothing})")


class TestInstantaneousSineWavePeriodMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        ind = InstantaneousSineWavePeriod(InstantaneousSineWavePeriodParams())
        self.assertEqual(ind.metadata().mnemonic, "iswp(0,4.00,50.00,20.00,0.01)")

    def test_custom_mnemonic(self):
        ind = InstantaneousSineWavePeriod(InstantaneousSineWavePeriodParams(smoothing=6))
        self.assertEqual(ind.metadata().mnemonic, "iswp(6,4.00,50.00,20.00,0.01)")


class TestInstantaneousSineWavePeriodMetadata(unittest.TestCase):
    def test_default_metadata(self):
        ind = InstantaneousSineWavePeriod(InstantaneousSineWavePeriodParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.INSTANTANEOUS_SINE_WAVE_PERIOD)
        self.assertEqual(meta.mnemonic, "iswp(0,4.00,50.00,20.00,0.01)")
        self.assertEqual(len(meta.outputs), 7)


class TestInstantaneousSineWavePeriodUpdateScalar(unittest.TestCase):
    def test_update_scalar(self):
        ind = InstantaneousSineWavePeriod(InstantaneousSineWavePeriodParams())
        tm = datetime(2021, 4, 1)
        out = None
        for c in INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=c))
        self.assertEqual(len(out), 7)
        # period output ordering check against the last finite value
        last = len(INPUT_CLOSE) - 1
        if math.isnan(EXPECTED_S0_PERIOD[last]):
            self.assertTrue(math.isnan(out[0].value))
        else:
            self.assertAlmostEqual(out[0].value, EXPECTED_S0_PERIOD[last], delta=TOLERANCE)


class TestInstantaneousSineWavePeriodInvalidParams(unittest.TestCase):
    def test_smoothing_negative(self):
        with self.assertRaises(ValueError):
            InstantaneousSineWavePeriod(InstantaneousSineWavePeriodParams(smoothing=-1))

    def test_min_period_zero(self):
        with self.assertRaises(ValueError):
            InstantaneousSineWavePeriod(InstantaneousSineWavePeriodParams(min_period=0.0))

    def test_max_le_min(self):
        with self.assertRaises(ValueError):
            InstantaneousSineWavePeriod(InstantaneousSineWavePeriodParams(min_period=10.0, max_period=10.0))

    def test_error_threshold_zero(self):
        with self.assertRaises(ValueError):
            InstantaneousSineWavePeriod(InstantaneousSineWavePeriodParams(error_threshold=0.0))

    def test_dx_zero(self):
        with self.assertRaises(ValueError):
            InstantaneousSineWavePeriod(InstantaneousSineWavePeriodParams(dx=0.0))


if __name__ == '__main__':
    unittest.main()
