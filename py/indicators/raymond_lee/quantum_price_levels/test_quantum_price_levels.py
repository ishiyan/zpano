import math
import unittest
from datetime import datetime

from py.indicators.raymond_lee.quantum_price_levels.quantum_price_levels import QuantumPriceLevels
from py.indicators.raymond_lee.quantum_price_levels.params import QuantumPriceLevelsParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from . import test_testdata as td

TOLERANCE = 1e-9


def _assert_series(test, actual, expected, label):
    test.assertEqual(len(actual), len(expected), f"{label}: length mismatch ({len(actual)} vs {len(expected)})")
    for i in range(len(expected)):
        delta = TOLERANCE * max(1.0, abs(expected[i]))
        test.assertAlmostEqual(actual[i], expected[i], delta=delta,
                               msg=f"{label}[{i}]: expected {expected[i]}, got {actual[i]}")


def _run_last(inputs, lookback=None, num_levels=21, num_bins=100, scale_factor=0.21):
    """Feed the price series and return the last (lambda, sigma, nqpr, resistances, supports)."""
    if lookback is None:
        lookback = len(inputs) - 1
    ind = QuantumPriceLevels(QuantumPriceLevelsParams(
        lookback=lookback, num_levels=num_levels, num_bins=num_bins, scale_factor=scale_factor))
    last = None
    for p in inputs:
        result = ind.update(p)
        if ind.is_primed():
            last = result
    return last


def _values(levels_list):
    return list(levels_list)


class TestQuantumPriceLevelsData(unittest.TestCase):
    """Test QPL against the reference test data for all parameter combos."""

    def _check(self, label, last, exp_nqpr, exp_upper, exp_lower):
        self.assertIsNotNone(last, f"{label}: no primed output")
        _, _, nqpr, resistances, supports = last
        _assert_series(self, nqpr, exp_nqpr, f"{label} NQPR")
        _assert_series(self, resistances, exp_upper, f"{label} UPPER")
        _assert_series(self, supports, exp_lower, f"{label} LOWER")

    # ── batch combos on the 252-bar input (lookback = len-1) ──────────────

    def test_default(self):
        self._check("default", _run_last(td.INPUT_CLOSE),
                    td.EXPECTED_NQPR, td.EXPECTED_UPPER, td.EXPECTED_LOWER)

    def test_factor_010(self):
        self._check("F0_10", _run_last(td.INPUT_CLOSE, scale_factor=0.10),
                    td.EXPECTED_NQPR_F0_10, td.EXPECTED_UPPER_F0_10, td.EXPECTED_LOWER_F0_10)

    def test_factor_042(self):
        self._check("F0_42", _run_last(td.INPUT_CLOSE, scale_factor=0.42),
                    td.EXPECTED_NQPR_F0_42, td.EXPECTED_UPPER_F0_42, td.EXPECTED_LOWER_F0_42)

    def test_bins_50(self):
        self._check("B50", _run_last(td.INPUT_CLOSE, num_bins=50),
                    td.EXPECTED_NQPR_B50, td.EXPECTED_UPPER_B50, td.EXPECTED_LOWER_B50)

    def test_bins_50_factor_010(self):
        self._check("B50_F0_10", _run_last(td.INPUT_CLOSE, num_bins=50, scale_factor=0.10),
                    td.EXPECTED_NQPR_B50_F0_10, td.EXPECTED_UPPER_B50_F0_10, td.EXPECTED_LOWER_B50_F0_10)

    def test_bins_50_factor_042(self):
        self._check("B50_F0_42", _run_last(td.INPUT_CLOSE, num_bins=50, scale_factor=0.42),
                    td.EXPECTED_NQPR_B50_F0_42, td.EXPECTED_UPPER_B50_F0_42, td.EXPECTED_LOWER_B50_F0_42)

    def test_levels_5(self):
        self._check("L5", _run_last(td.INPUT_CLOSE, num_levels=5),
                    td.EXPECTED_NQPR_L5, td.EXPECTED_UPPER_L5, td.EXPECTED_LOWER_L5)

    def test_levels_10(self):
        self._check("L10", _run_last(td.INPUT_CLOSE, num_levels=10),
                    td.EXPECTED_NQPR_L10, td.EXPECTED_UPPER_L10, td.EXPECTED_LOWER_L10)

    def test_all_non_default(self):
        self._check("L10_B50_F0_42", _run_last(td.INPUT_CLOSE, num_levels=10, num_bins=50, scale_factor=0.42),
                    td.EXPECTED_NQPR_L10_B50_F0_42, td.EXPECTED_UPPER_L10_B50_F0_42, td.EXPECTED_LOWER_L10_B50_F0_42)

    def test_long_2k(self):
        self._check("2K", _run_last(td.INPUT_PRICES_2048),
                    td.EXPECTED_NQPR_2K, td.EXPECTED_UPPER_2K, td.EXPECTED_LOWER_2K)

    # ── reference-price combos: validate via re-projection of NQPR ────────

    def _check_ref(self, label, ref, exp_nqpr, exp_upper, exp_lower):
        last = _run_last(td.INPUT_CLOSE)
        self.assertIsNotNone(last)
        _, _, nqpr, _, _ = last
        _assert_series(self, nqpr, exp_nqpr, f"{label} NQPR")
        _assert_series(self, [ref * m for m in nqpr], exp_upper, f"{label} UPPER")
        _assert_series(self, [ref / m for m in nqpr], exp_lower, f"{label} LOWER")

    def test_ref_50(self):
        self._check_ref("R50_0", 50.0, td.EXPECTED_NQPR_R50_0, td.EXPECTED_UPPER_R50_0, td.EXPECTED_LOWER_R50_0)

    def test_ref_1000(self):
        self._check_ref("R1000_0", 1000.0, td.EXPECTED_NQPR_R1000_0, td.EXPECTED_UPPER_R1000_0, td.EXPECTED_LOWER_R1000_0)

    def test_ref_1_2345(self):
        self._check_ref("R1_2345", 1.2345, td.EXPECTED_NQPR_R1_2345, td.EXPECTED_UPPER_R1_2345, td.EXPECTED_LOWER_R1_2345)

    # ── streaming combos (sliding window) on the 252-bar input ────────────

    def test_streaming_100(self):
        self._check("S100", _run_last(td.INPUT_CLOSE, lookback=100),
                    td.EXPECTED_NQPR_S100, td.EXPECTED_UPPER_S100, td.EXPECTED_LOWER_S100)

    def test_streaming_150_bins_50(self):
        self._check("S150_B50", _run_last(td.INPUT_CLOSE, lookback=150, num_bins=50),
                    td.EXPECTED_NQPR_S150_B50, td.EXPECTED_UPPER_S150_B50, td.EXPECTED_LOWER_S150_B50)

    def test_streaming_200_factor_042(self):
        self._check("S200_F0_42", _run_last(td.INPUT_CLOSE, lookback=200, scale_factor=0.42),
                    td.EXPECTED_NQPR_S200_F0_42, td.EXPECTED_UPPER_S200_F0_42, td.EXPECTED_LOWER_S200_F0_42)


class TestQuantumPriceLevelsScalars(unittest.TestCase):
    def test_lambda_sigma_default(self):
        last = _run_last(td.INPUT_CLOSE)
        lambda_, sigma, _, _, _ = last
        self.assertAlmostEqual(lambda_, 9.739608012591481e-01, delta=1e-9)
        self.assertAlmostEqual(sigma, 2.662021797593086e-02, delta=1e-9)


class TestQuantumPriceLevelsMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        ind = QuantumPriceLevels(QuantumPriceLevelsParams())
        self.assertEqual(ind.metadata().mnemonic, "qpl(2048,21,100,0.21)")


class TestQuantumPriceLevelsMetadata(unittest.TestCase):
    def test_metadata(self):
        ind = QuantumPriceLevels(QuantumPriceLevelsParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.QUANTUM_PRICE_LEVELS)
        self.assertEqual(len(meta.outputs), 5)


class TestQuantumPriceLevelsUpdateScalar(unittest.TestCase):
    def test_update_scalar_outputs(self):
        ind = QuantumPriceLevels(QuantumPriceLevelsParams(lookback=100))
        tm = datetime(2021, 4, 1)
        out = None
        for p in td.INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=p))
        self.assertEqual(len(out), 5)
        # outputs 0,1 are scalars; 2,3,4 are Levels
        self.assertEqual(len(out[3].levels), 21)


class TestQuantumPriceLevelsInvalidParams(unittest.TestCase):
    def test_lookback_too_small(self):
        with self.assertRaises(ValueError):
            QuantumPriceLevels(QuantumPriceLevelsParams(lookback=1))

    def test_num_levels_too_small(self):
        with self.assertRaises(ValueError):
            QuantumPriceLevels(QuantumPriceLevelsParams(num_levels=0))

    def test_num_bins_too_small(self):
        with self.assertRaises(ValueError):
            QuantumPriceLevels(QuantumPriceLevelsParams(num_bins=1))

    def test_scale_factor_non_positive(self):
        with self.assertRaises(ValueError):
            QuantumPriceLevels(QuantumPriceLevelsParams(scale_factor=0.0))


if __name__ == '__main__':
    unittest.main()
