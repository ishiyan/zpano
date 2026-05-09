"""Tests for SuperSmoother indicator."""

import math
import unittest
from datetime import datetime

from py.indicators.john_ehlers.super_smoother.super_smoother import SuperSmoother
from py.indicators.john_ehlers.super_smoother.params import SuperSmootherParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.bar_component import BarComponent

from .test_testdata import _input, _expected


class TestSuperSmoother(unittest.TestCase):

    def _create(self, period=10):
        params = SuperSmootherParams(shortest_cycle_period=period)
        return SuperSmoother.create(params)

    def test_update(self):
        """Test SuperSmoother Update against Julia reference data."""
        skip_rows = 60
        tolerance = 0.5

        inp = _input()
        exp = _expected()
        ss = self._create(10)

        for i in range(len(inp)):
            act = ss.update(inp[i])

            if i < 2:
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
                continue

            if i < skip_rows:
                continue

            self.assertAlmostEqual(act, exp[i], delta=tolerance,
                                   msg=f"[{i}] expected {exp[i]}, got {act}")

        # NaN passthrough.
        self.assertTrue(math.isnan(ss.update(math.nan)))

    def test_is_primed(self):
        """Test priming behavior."""
        inp = _input()
        ss = self._create(10)

        self.assertFalse(ss.is_primed())

        for i in range(2):
            ss.update(inp[i])
            self.assertFalse(ss.is_primed(), f"[{i}] should not be primed")

        ss.update(inp[2])
        self.assertTrue(ss.is_primed(), "[2] should be primed")

    def test_update_entity(self):
        """Test entity update methods."""
        t = datetime(2021, 4, 1)
        inp_val = 100.0

        def check(output):
            self.assertEqual(len(output), 1)
            s = output[0]
            self.assertIsInstance(s, Scalar)
            self.assertEqual(s.time, t)
            self.assertFalse(math.isnan(s.value))

        # Scalar
        ss = self._create(10)
        for _ in range(3):
            ss.update(inp_val)
        check(ss.update_scalar(Scalar(time=t, value=inp_val)))

        # Bar (default component is Median = (high+low)/2)
        ss = self._create(10)
        for _ in range(3):
            ss.update(inp_val)
        b = Bar(time=t, open=0, high=inp_val, low=inp_val, close=0, volume=0)
        check(ss.update_bar(b))

        # Quote
        ss = self._create(10)
        for _ in range(3):
            ss.update(inp_val)
        q = Quote(time=t, bid_price=inp_val, ask_price=inp_val, bid_size=0, ask_size=0)
        check(ss.update_quote(q))

        # Trade
        ss = self._create(10)
        for _ in range(3):
            ss.update(inp_val)
        tr = Trade(time=t, price=inp_val, volume=0)
        check(ss.update_trade(tr))

    def test_metadata(self):
        """Test metadata output."""
        ss = self._create(10)
        m = ss.metadata()

        self.assertEqual(m.identifier, Identifier.SUPER_SMOOTHER)
        self.assertEqual(len(m.outputs), 1)
        self.assertEqual(m.outputs[0].mnemonic, "ss(10, hl/2)")
        self.assertEqual(m.outputs[0].description, "Super Smoother ss(10, hl/2)")

    def test_create_validation(self):
        """Test parameter validation."""
        with self.assertRaises(ValueError):
            SuperSmoother.create(SuperSmootherParams(shortest_cycle_period=1))
        with self.assertRaises(ValueError):
            SuperSmoother.create(SuperSmootherParams(shortest_cycle_period=0))
        with self.assertRaises(ValueError):
            SuperSmoother.create(SuperSmootherParams(shortest_cycle_period=-1))

    def test_create_default_components(self):
        """Test default component mnemonics."""
        ss = self._create(10)
        m = ss.metadata()
        # Default bar component is MEDIAN (non-default vs CLOSE), so mnemonic includes ", hl/2"
        self.assertEqual(m.outputs[0].mnemonic, "ss(10, hl/2)")

    def test_create_custom_bar_component(self):
        """Test custom bar component."""
        params = SuperSmootherParams(shortest_cycle_period=10, bar_component=BarComponent.OPEN)
        ss = SuperSmoother.create(params)
        m = ss.metadata()
        self.assertEqual(m.outputs[0].mnemonic, "ss(10, o)")


if __name__ == '__main__':
    unittest.main()
