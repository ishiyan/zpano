"""Tests for Cyber Cycle indicator."""
import math
import unittest

from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.scalar import Scalar
from py.entities.trade import Trade
from py.indicators.john_ehlers.cyber_cycle.cyber_cycle import CyberCycle
from py.indicators.john_ehlers.cyber_cycle.output import CyberCycleOutput
from py.indicators.john_ehlers.cyber_cycle.params import (
    CyberCycleLengthParams,
    CyberCycleSmoothingFactorParams,
)

from .test_testdata import (
    _test_time,
    _test_input,
    _test_expected_cycle,
    _test_expected_signal,
    _test_input_high,
    _test_input_low,
)


def _create_default():
    params = CyberCycleSmoothingFactorParams(smoothing_factor=0.07, signal_lag=9)
    return CyberCycle.from_smoothing_factor(params)


class TestCyberCycle(unittest.TestCase):
    """Tests for CyberCycle indicator."""

    def test_update_cycle_value(self):
        """Test cycle values against Excel reference data."""
        cc = _create_default()
        inp = _test_input()
        exp = _test_expected_cycle()
        lprimed = 7

        for i in range(lprimed):
            v = cc.update(inp[i])
            self.assertTrue(math.isnan(v), f"[{i}] expected NaN, got {v}")

        for i in range(lprimed, len(inp)):
            v = cc.update(inp[i])
            self.assertAlmostEqual(v, exp[i], delta=1e-8, msg=f"[{i}]")

        self.assertTrue(math.isnan(cc.update(math.nan)))

    def test_update_signal(self):
        """Test signal values against Excel reference data."""
        cc = _create_default()
        inp = _test_input()
        exp_signal = _test_expected_signal()
        lprimed = 7

        for i in range(lprimed):
            cc.update(inp[i])

        for i in range(lprimed, len(inp)):
            cc.update(inp[i])
            self.assertAlmostEqual(cc._signal, exp_signal[i], delta=1e-8, msg=f"[{i}]")

    def test_update_entity_scalar(self):
        """Test scalar entity update."""
        cc = _create_default()
        t = _test_time()
        inp = _test_input()
        exp_cycle = _test_expected_cycle()
        exp_signal = _test_expected_signal()

        for i in range(len(inp)):
            s = Scalar(time=t, value=inp[i])
            out = cc.update_scalar(s)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_cycle[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_cycle[i], delta=1e-8, msg=f"[{i}] cycle")
                self.assertAlmostEqual(out[1].value, exp_signal[i], delta=1e-8, msg=f"[{i}] signal")

    def test_update_entity_bar(self):
        """Test bar entity update (median price)."""
        cc = _create_default()
        t = _test_time()
        inp_high = _test_input_high()
        inp_low = _test_input_low()
        exp_cycle = _test_expected_cycle()
        exp_signal = _test_expected_signal()

        for i in range(len(inp_high)):
            b = Bar(time=t, open=0, high=inp_high[i], low=inp_low[i], close=0, volume=0)
            out = cc.update_bar(b)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_cycle[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_cycle[i], delta=1e-8, msg=f"[{i}] cycle")
                self.assertAlmostEqual(out[1].value, exp_signal[i], delta=1e-8, msg=f"[{i}] signal")

    def test_update_entity_quote(self):
        """Test quote entity update (mid price)."""
        cc = _create_default()
        t = _test_time()
        inp_high = _test_input_high()
        inp_low = _test_input_low()
        exp_cycle = _test_expected_cycle()
        exp_signal = _test_expected_signal()

        for i in range(len(inp_high)):
            q = Quote(time=t, bid_price=inp_low[i], ask_price=inp_high[i], bid_size=0, ask_size=0)
            out = cc.update_quote(q)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_cycle[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_cycle[i], delta=1e-8, msg=f"[{i}] cycle")
                self.assertAlmostEqual(out[1].value, exp_signal[i], delta=1e-8, msg=f"[{i}] signal")

    def test_update_entity_trade(self):
        """Test trade entity update."""
        cc = _create_default()
        t = _test_time()
        inp = _test_input()
        exp_cycle = _test_expected_cycle()
        exp_signal = _test_expected_signal()

        for i in range(len(inp)):
            tr = Trade(time=t, price=inp[i], volume=0)
            out = cc.update_trade(tr)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_cycle[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_cycle[i], delta=1e-8, msg=f"[{i}] cycle")
                self.assertAlmostEqual(out[1].value, exp_signal[i], delta=1e-8, msg=f"[{i}] signal")

    def test_is_primed(self):
        """Test priming behavior."""
        cc = _create_default()
        inp = _test_input()
        lprimed = 7

        self.assertFalse(cc.is_primed())

        for i in range(lprimed):
            cc.update(inp[i])
            self.assertFalse(cc.is_primed(), f"[{i+1}]")

        for i in range(lprimed, len(inp)):
            cc.update(inp[i])
            self.assertTrue(cc.is_primed(), f"[{i+1}]")

    def test_metadata(self):
        """Test metadata output."""
        cc = _create_default()
        meta = cc.metadata()

        self.assertEqual(meta.mnemonic, "cc(28, hl/2)")
        self.assertEqual(meta.description, "Cyber Cycle cc(28, hl/2)")
        self.assertEqual(len(meta.outputs), 2)

        self.assertEqual(meta.outputs[0].kind, int(CyberCycleOutput.VALUE))
        self.assertEqual(meta.outputs[0].mnemonic, "cc(28, hl/2)")
        self.assertEqual(meta.outputs[1].kind, int(CyberCycleOutput.SIGNAL))
        self.assertEqual(meta.outputs[1].mnemonic, "ccSignal(28, hl/2)")

    def test_create_validation(self):
        """Test parameter validation."""
        with self.assertRaises(ValueError):
            CyberCycle.from_length(CyberCycleLengthParams(length=0, signal_lag=1))
        with self.assertRaises(ValueError):
            CyberCycle.from_length(CyberCycleLengthParams(length=-8, signal_lag=1))
        with self.assertRaises(ValueError):
            CyberCycle.from_length(CyberCycleLengthParams(length=1, signal_lag=0))
        with self.assertRaises(ValueError):
            CyberCycle.from_length(CyberCycleLengthParams(length=1, signal_lag=-8))
        with self.assertRaises(ValueError):
            CyberCycle.from_smoothing_factor(CyberCycleSmoothingFactorParams(smoothing_factor=-0.0001, signal_lag=8))
        with self.assertRaises(ValueError):
            CyberCycle.from_smoothing_factor(CyberCycleSmoothingFactorParams(smoothing_factor=1.0001, signal_lag=8))
        with self.assertRaises(ValueError):
            CyberCycle.from_smoothing_factor(CyberCycleSmoothingFactorParams(smoothing_factor=0.07, signal_lag=0))


if __name__ == '__main__':
    unittest.main()
