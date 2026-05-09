import math
import unittest

from py.indicators.george_lane.stochastic.stochastic import Stochastic
from py.indicators.george_lane.stochastic.params import StochasticParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.scalar import Scalar

from .test_testdata import (
    TEST_HIGH,
    TEST_LOW,
    TEST_CLOSE,
    TEST_TIME,
)


# Standard test data (252 entries) — same H/L/C as Go tests.
class TestStochastic(unittest.TestCase):
    """Tests for the Stochastic Oscillator indicator."""

    def test_5_sma3_sma4_single_value(self):
        """fastK=5, slowK=3/SMA, slowD=4/SMA. First primed at index 9."""
        ind = Stochastic(StochasticParams(fast_k_length=5, slow_k_length=3, slow_d_length=4))
        for i in range(9):
            ind.update(TEST_CLOSE[i], TEST_HIGH[i], TEST_LOW[i])
        _, slow_k, slow_d = ind.update(TEST_CLOSE[9], TEST_HIGH[9], TEST_LOW[9])
        self.assertAlmostEqual(slow_k, 38.139, delta=1e-2)
        self.assertAlmostEqual(slow_d, 36.725, delta=1e-2)
        self.assertTrue(ind.is_primed())

    def test_5_sma3_sma3_first_value(self):
        """fastK=5, slowK=3/SMA, slowD=3/SMA. First primed at index 8."""
        ind = Stochastic(StochasticParams(fast_k_length=5, slow_k_length=3, slow_d_length=3))
        for i in range(8):
            ind.update(TEST_CLOSE[i], TEST_HIGH[i], TEST_LOW[i])
        _, slow_k, slow_d = ind.update(TEST_CLOSE[8], TEST_HIGH[8], TEST_LOW[8])
        self.assertAlmostEqual(slow_k, 24.0128, delta=1e-2)
        self.assertAlmostEqual(slow_d, 36.254, delta=1e-2)
        self.assertTrue(ind.is_primed())

    def test_5_sma3_sma3_last_value(self):
        """fastK=5, slowK=3/SMA, slowD=3/SMA. Last values."""
        ind = Stochastic(StochasticParams(fast_k_length=5, slow_k_length=3, slow_d_length=3))
        slow_k = slow_d = 0.0
        for i in range(252):
            _, slow_k, slow_d = ind.update(TEST_CLOSE[i], TEST_HIGH[i], TEST_LOW[i])
        self.assertAlmostEqual(slow_k, 30.194, delta=1e-2)
        self.assertAlmostEqual(slow_d, 43.69, delta=1e-2)

    def test_5_sma3_sma4_last_value(self):
        """fastK=5, slowK=3/SMA, slowD=4/SMA. Last values."""
        ind = Stochastic(StochasticParams(fast_k_length=5, slow_k_length=3, slow_d_length=4))
        slow_k = slow_d = 0.0
        for i in range(252):
            _, slow_k, slow_d = ind.update(TEST_CLOSE[i], TEST_HIGH[i], TEST_LOW[i])
        self.assertAlmostEqual(slow_k, 30.194, delta=1e-2)
        self.assertAlmostEqual(slow_d, 46.641, delta=1e-2)

    def test_is_primed(self):
        """Priming at index 8 for fastK=5, slowK=3/SMA, slowD=3/SMA."""
        ind = Stochastic(StochasticParams(fast_k_length=5, slow_k_length=3, slow_d_length=3))
        self.assertFalse(ind.is_primed())
        for i in range(8):
            ind.update(TEST_CLOSE[i], TEST_HIGH[i], TEST_LOW[i])
            self.assertFalse(ind.is_primed(), f"[{i}] should not be primed")
        ind.update(TEST_CLOSE[8], TEST_HIGH[8], TEST_LOW[8])
        self.assertTrue(ind.is_primed(), "should be primed after index 8")

    def test_nan(self):
        """NaN input returns NaN outputs."""
        ind = Stochastic(StochasticParams(fast_k_length=5, slow_k_length=3, slow_d_length=3))
        fk, sk, sd = ind.update(math.nan, 1.0, 1.0)
        self.assertTrue(math.isnan(fk))
        self.assertTrue(math.isnan(sk))
        self.assertTrue(math.isnan(sd))

    def test_metadata(self):
        """Test metadata."""
        ind = Stochastic(StochasticParams(fast_k_length=5, slow_k_length=3, slow_d_length=3))
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.STOCHASTIC)
        self.assertEqual(meta.mnemonic, "stoch(5/SMA3/SMA3)")
        self.assertEqual(len(meta.outputs), 3)

    def test_update_bar(self):
        """Test UpdateBar entity adapter."""
        ind = Stochastic(StochasticParams(fast_k_length=5, slow_k_length=3, slow_d_length=3))
        for i in range(8):
            bar = Bar(time=TEST_TIME, open=0, high=TEST_HIGH[i], low=TEST_LOW[i],
                      close=TEST_CLOSE[i], volume=0)
            out = ind.update_bar(bar)
            self.assertTrue(math.isnan(out[2].value), f"[{i}] expected NaN slowD")
        bar = Bar(time=TEST_TIME, open=0, high=TEST_HIGH[8], low=TEST_LOW[8],
                  close=TEST_CLOSE[8], volume=0)
        out = ind.update_bar(bar)
        self.assertAlmostEqual(out[1].value, 24.0128, delta=1e-2)
        self.assertAlmostEqual(out[2].value, 36.254, delta=1e-2)

    def test_invalid_params(self):
        """Invalid parameters raise ValueError."""
        with self.assertRaises(ValueError):
            Stochastic(StochasticParams(fast_k_length=0, slow_k_length=3, slow_d_length=3))
        with self.assertRaises(ValueError):
            Stochastic(StochasticParams(fast_k_length=5, slow_k_length=0, slow_d_length=3))
        with self.assertRaises(ValueError):
            Stochastic(StochasticParams(fast_k_length=5, slow_k_length=3, slow_d_length=0))
        with self.assertRaises(ValueError):
            Stochastic(StochasticParams(fast_k_length=-1, slow_k_length=3, slow_d_length=3))


if __name__ == "__main__":
    unittest.main()
