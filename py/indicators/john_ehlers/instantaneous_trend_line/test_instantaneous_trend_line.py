"""Tests for Instantaneous Trend Line indicator."""
import math
import unittest

from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.scalar import Scalar
from py.entities.trade import Trade
from py.indicators.john_ehlers.instantaneous_trend_line.instantaneous_trend_line import (
    InstantaneousTrendLine,
)
from py.indicators.john_ehlers.instantaneous_trend_line.output import InstantaneousTrendLineOutput
from py.indicators.john_ehlers.instantaneous_trend_line.params import (
    InstantaneousTrendLineLengthParams,
    InstantaneousTrendLineSmoothingFactorParams,
)

from .test_testdata import (
    _test_time,
    _test_input,
    _test_expected_trend_line,
    _test_expected_trigger,
    _test_input_high,
    _test_input_low,
)


def _create_default():
    params = InstantaneousTrendLineSmoothingFactorParams(smoothing_factor=0.07)
    return InstantaneousTrendLine.from_smoothing_factor(params)


class TestInstantaneousTrendLine(unittest.TestCase):
    """Tests for InstantaneousTrendLine indicator."""

    def test_update_trend_line(self):
        """Test trend line values against Excel reference data."""
        itl = _create_default()
        inp = _test_input()
        exp = _test_expected_trend_line()
        lprimed = 4

        for i in range(lprimed):
            v = itl.update(inp[i])
            self.assertTrue(math.isnan(v), f"[{i}] expected NaN, got {v}")

        for i in range(lprimed, len(inp)):
            v = itl.update(inp[i])
            self.assertAlmostEqual(v, exp[i], delta=1e-8, msg=f"[{i}]")

        self.assertTrue(math.isnan(itl.update(math.nan)))

    def test_update_trigger_line(self):
        """Test trigger line values against Excel reference data."""
        itl = _create_default()
        inp = _test_input()
        exp_trig = _test_expected_trigger()
        lprimed = 4

        for i in range(lprimed):
            itl.update(inp[i])

        for i in range(lprimed, len(inp)):
            itl.update(inp[i])
            self.assertAlmostEqual(itl._trigger_line, exp_trig[i], delta=1e-8, msg=f"[{i}]")

    def test_update_entity_scalar(self):
        """Test scalar entity update."""
        itl = _create_default()
        t = _test_time()
        inp = _test_input()
        exp_trend = _test_expected_trend_line()
        exp_trig = _test_expected_trigger()

        for i in range(len(inp)):
            s = Scalar(time=t, value=inp[i])
            out = itl.update_scalar(s)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_trend[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_trend[i], delta=1e-8, msg=f"[{i}] trend")
                self.assertAlmostEqual(out[1].value, exp_trig[i], delta=1e-8, msg=f"[{i}] trigger")

    def test_update_entity_bar(self):
        """Test bar entity update (median price)."""
        itl = _create_default()
        t = _test_time()
        inp_high = _test_input_high()
        inp_low = _test_input_low()
        exp_trend = _test_expected_trend_line()
        exp_trig = _test_expected_trigger()

        for i in range(len(inp_high)):
            b = Bar(time=t, open=0, high=inp_high[i], low=inp_low[i], close=0, volume=0)
            out = itl.update_bar(b)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_trend[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_trend[i], delta=1e-8, msg=f"[{i}] trend")
                self.assertAlmostEqual(out[1].value, exp_trig[i], delta=1e-8, msg=f"[{i}] trigger")

    def test_update_entity_quote(self):
        """Test quote entity update (mid price)."""
        itl = _create_default()
        t = _test_time()
        inp_high = _test_input_high()
        inp_low = _test_input_low()
        exp_trend = _test_expected_trend_line()
        exp_trig = _test_expected_trigger()

        for i in range(len(inp_high)):
            q = Quote(time=t, bid_price=inp_low[i], ask_price=inp_high[i], bid_size=0, ask_size=0)
            out = itl.update_quote(q)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_trend[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_trend[i], delta=1e-8, msg=f"[{i}] trend")
                self.assertAlmostEqual(out[1].value, exp_trig[i], delta=1e-8, msg=f"[{i}] trigger")

    def test_update_entity_trade(self):
        """Test trade entity update."""
        itl = _create_default()
        t = _test_time()
        inp = _test_input()
        exp_trend = _test_expected_trend_line()
        exp_trig = _test_expected_trigger()

        for i in range(len(inp)):
            tr = Trade(time=t, price=inp[i], volume=0)
            out = itl.update_trade(tr)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_trend[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_trend[i], delta=1e-8, msg=f"[{i}] trend")
                self.assertAlmostEqual(out[1].value, exp_trig[i], delta=1e-8, msg=f"[{i}] trigger")

    def test_is_primed(self):
        """Test priming behavior."""
        itl = _create_default()
        inp = _test_input()
        lprimed = 4

        self.assertFalse(itl.is_primed())

        for i in range(lprimed):
            itl.update(inp[i])
            self.assertFalse(itl.is_primed(), f"[{i+1}]")

        for i in range(lprimed, len(inp)):
            itl.update(inp[i])
            self.assertTrue(itl.is_primed(), f"[{i+1}]")

    def test_metadata(self):
        """Test metadata output."""
        itl = _create_default()
        meta = itl.metadata()

        self.assertEqual(meta.mnemonic, "iTrend(28, hl/2)")
        self.assertEqual(meta.description, "Instantaneous Trend Line iTrend(28, hl/2)")
        self.assertEqual(len(meta.outputs), 2)

        self.assertEqual(meta.outputs[0].kind, int(InstantaneousTrendLineOutput.VALUE))
        self.assertEqual(meta.outputs[0].mnemonic, "iTrend(28, hl/2)")
        self.assertEqual(meta.outputs[1].kind, int(InstantaneousTrendLineOutput.TRIGGER))
        self.assertEqual(meta.outputs[1].mnemonic, "iTrendTrigger(28, hl/2)")

    def test_create_validation(self):
        """Test parameter validation."""
        with self.assertRaises(ValueError):
            InstantaneousTrendLine.from_length(InstantaneousTrendLineLengthParams(length=0))
        with self.assertRaises(ValueError):
            InstantaneousTrendLine.from_length(InstantaneousTrendLineLengthParams(length=-8))
        with self.assertRaises(ValueError):
            InstantaneousTrendLine.from_smoothing_factor(InstantaneousTrendLineSmoothingFactorParams(smoothing_factor=-0.0001))
        with self.assertRaises(ValueError):
            InstantaneousTrendLine.from_smoothing_factor(InstantaneousTrendLineSmoothingFactorParams(smoothing_factor=1.0001))


if __name__ == '__main__':
    unittest.main()
