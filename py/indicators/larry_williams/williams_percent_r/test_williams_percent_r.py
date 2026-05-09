import math
import unittest
from datetime import datetime

from py.indicators.larry_williams.williams_percent_r.williams_percent_r import WilliamsPercentR
from py.indicators.larry_williams.williams_percent_r.output import WilliamsPercentROutput
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.scalar import Scalar
from py.entities.quote import Quote
from py.entities.trade import Trade

from .test_testdata import (
    TEST_HIGH,
    TEST_LOW,
    TEST_CLOSE,
    EXPECTED_14,
    EXPECTED_2,
)


# Standard test data (252 entries) — same H/L/C as Go tests.
class TestWilliamsPercentRUpdate14(unittest.TestCase):
    """Test Williams %R with period=14."""

    def test_values(self):
        w = WilliamsPercentR(14)
        for i in range(len(TEST_CLOSE)):
            act = w.update(TEST_CLOSE[i], TEST_HIGH[i], TEST_LOW[i])
            exp = EXPECTED_14[i]
            if exp is None:
                self.assertTrue(math.isnan(act),
                                f"[{i}] expected NaN, got {act}")
            else:
                self.assertFalse(math.isnan(act),
                                 f"[{i}] expected {exp}, got NaN")
                self.assertAlmostEqual(act, exp, delta=1e-6,
                                       msg=f"[{i}] expected {exp}, got {act}")


class TestWilliamsPercentRUpdate2(unittest.TestCase):
    """Test Williams %R with period=2."""

    def test_values(self):
        w = WilliamsPercentR(2)
        for i in range(len(TEST_CLOSE)):
            act = w.update(TEST_CLOSE[i], TEST_HIGH[i], TEST_LOW[i])
            exp = EXPECTED_2[i]
            if exp is None:
                self.assertTrue(math.isnan(act),
                                f"[{i}] expected NaN, got {act}")
            else:
                self.assertFalse(math.isnan(act),
                                 f"[{i}] expected {exp}, got NaN")
                self.assertAlmostEqual(act, exp, delta=1e-6,
                                       msg=f"[{i}] expected {exp}, got {act}")


class TestWilliamsPercentRNaN(unittest.TestCase):
    """Test NaN passthrough."""

    def test_nan_close(self):
        w = WilliamsPercentR(14)
        self.assertTrue(math.isnan(w.update(math.nan, 1, 1)))

    def test_nan_high(self):
        w = WilliamsPercentR(14)
        self.assertTrue(math.isnan(w.update(1, math.nan, 1)))

    def test_nan_low(self):
        w = WilliamsPercentR(14)
        self.assertTrue(math.isnan(w.update(1, 1, math.nan)))


class TestWilliamsPercentRIsPrimed(unittest.TestCase):
    """Test priming behavior."""

    def test_priming(self):
        w = WilliamsPercentR(14)
        self.assertFalse(w.is_primed())

        for i in range(13):
            w.update(TEST_CLOSE[i], TEST_HIGH[i], TEST_LOW[i])
            self.assertFalse(w.is_primed(), f"[{i}] should not be primed")

        w.update(TEST_CLOSE[13], TEST_HIGH[13], TEST_LOW[13])
        self.assertTrue(w.is_primed(), "should be primed after 14th update")

        w.update(TEST_CLOSE[14], TEST_HIGH[14], TEST_LOW[14])
        self.assertTrue(w.is_primed(), "should remain primed")


class TestWilliamsPercentRUpdateSample(unittest.TestCase):
    """Test update_sample (H=L=C all same → %R = 0 when primed)."""

    def test_constant_sample(self):
        w = WilliamsPercentR(14)
        for i in range(13):
            v = w.update_sample(9.0)
            self.assertTrue(math.isnan(v), f"[{i}] expected NaN")

        v = w.update_sample(9.0)
        self.assertEqual(v, 0.0)


class TestWilliamsPercentREntities(unittest.TestCase):
    """Test entity update methods."""

    def _make_primed(self):
        w = WilliamsPercentR(14)
        for i in range(14):
            w.update(TEST_CLOSE[i], TEST_HIGH[i], TEST_LOW[i])
        return w

    def test_update_bar(self):
        w = self._make_primed()
        b = Bar(time=datetime(2021, 4, 1), open=0, high=TEST_HIGH[14],
                low=TEST_LOW[14], close=TEST_CLOSE[14], volume=0)
        out = w.update_bar(b)
        self.assertEqual(len(out), 1)
        self.assertIsInstance(out[0], Scalar)
        self.assertEqual(out[0].time, datetime(2021, 4, 1))

    def test_update_scalar(self):
        w = self._make_primed()
        s = Scalar(time=datetime(2021, 4, 1), value=100)
        out = w.update_scalar(s)
        self.assertEqual(len(out), 1)
        self.assertIsInstance(out[0], Scalar)

    def test_update_quote(self):
        w = self._make_primed()
        q = Quote(time=datetime(2021, 4, 1), bid_price=99, ask_price=101,
                  bid_size=1, ask_size=1)
        out = w.update_quote(q)
        self.assertEqual(len(out), 1)
        self.assertIsInstance(out[0], Scalar)

    def test_update_trade(self):
        w = self._make_primed()
        t = Trade(time=datetime(2021, 4, 1), price=100, volume=1)
        out = w.update_trade(t)
        self.assertEqual(len(out), 1)
        self.assertIsInstance(out[0], Scalar)


class TestWilliamsPercentRMetadata(unittest.TestCase):
    """Test metadata output."""

    def test_metadata(self):
        w = WilliamsPercentR(14)
        m = w.metadata()
        self.assertEqual(m.identifier, Identifier.WILLIAMS_PERCENT_R)
        self.assertEqual(m.mnemonic, "willr")
        self.assertEqual(m.description, "Williams %R")
        self.assertEqual(len(m.outputs), 1)
        self.assertEqual(m.outputs[0].mnemonic, "willr")
        self.assertEqual(m.outputs[0].description, "Williams %R")


if __name__ == '__main__':
    unittest.main()
