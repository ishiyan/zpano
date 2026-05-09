"""Tests for Center of Gravity Oscillator indicator."""
import math
import unittest

from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.scalar import Scalar
from py.entities.trade import Trade
from py.indicators.john_ehlers.center_of_gravity_oscillator.center_of_gravity_oscillator import (
    CenterOfGravityOscillator,
)
from py.indicators.john_ehlers.center_of_gravity_oscillator.output import CenterOfGravityOscillatorOutput
from py.indicators.john_ehlers.center_of_gravity_oscillator.params import (
    CenterOfGravityOscillatorParams,
)

from .test_testdata import (
    _test_time,
    _test_input,
    _test_expected_cog,
    _test_expected_trigger,
    _test_input_high,
    _test_input_low,
)


def _create_default():
    params = CenterOfGravityOscillatorParams(length=10)
    return CenterOfGravityOscillator.create(params)


class TestCenterOfGravityOscillator(unittest.TestCase):
    """Tests for CenterOfGravityOscillator indicator."""

    def test_update_cog_value(self):
        """Test COG values against Excel reference data."""
        cog = _create_default()
        inp = _test_input()
        exp = _test_expected_cog()
        lprimed = 10

        for i in range(lprimed):
            v = cog.update(inp[i])
            self.assertTrue(math.isnan(v), f"[{i}] expected NaN, got {v}")

        for i in range(lprimed, len(inp)):
            v = cog.update(inp[i])
            self.assertAlmostEqual(v, exp[i], delta=1e-8, msg=f"[{i}]")

        self.assertTrue(math.isnan(cog.update(math.nan)))

    def test_update_trigger(self):
        """Test trigger values against Excel reference data."""
        cog = _create_default()
        inp = _test_input()
        exp_trig = _test_expected_trigger()
        lprimed = 10

        for i in range(lprimed):
            cog.update(inp[i])

        for i in range(lprimed, len(inp)):
            cog.update(inp[i])
            self.assertAlmostEqual(cog._value_previous, exp_trig[i], delta=1e-8, msg=f"[{i}]")

    def test_update_entity_scalar(self):
        """Test scalar entity update."""
        cog = _create_default()
        t = _test_time()
        inp = _test_input()
        exp_cog = _test_expected_cog()
        exp_trig = _test_expected_trigger()

        for i in range(len(inp)):
            s = Scalar(time=t, value=inp[i])
            out = cog.update_scalar(s)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_cog[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_cog[i], delta=1e-8, msg=f"[{i}] cog")
                self.assertAlmostEqual(out[1].value, exp_trig[i], delta=1e-8, msg=f"[{i}] trigger")

    def test_update_entity_bar(self):
        """Test bar entity update (median price)."""
        cog = _create_default()
        t = _test_time()
        inp_high = _test_input_high()
        inp_low = _test_input_low()
        exp_cog = _test_expected_cog()
        exp_trig = _test_expected_trigger()

        for i in range(len(inp_high)):
            b = Bar(time=t, open=0, high=inp_high[i], low=inp_low[i], close=0, volume=0)
            out = cog.update_bar(b)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_cog[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_cog[i], delta=1e-8, msg=f"[{i}] cog")
                self.assertAlmostEqual(out[1].value, exp_trig[i], delta=1e-8, msg=f"[{i}] trigger")

    def test_update_entity_quote(self):
        """Test quote entity update (mid price)."""
        cog = _create_default()
        t = _test_time()
        inp_high = _test_input_high()
        inp_low = _test_input_low()
        exp_cog = _test_expected_cog()
        exp_trig = _test_expected_trigger()

        for i in range(len(inp_high)):
            q = Quote(time=t, bid_price=inp_low[i], ask_price=inp_high[i], bid_size=0, ask_size=0)
            out = cog.update_quote(q)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_cog[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_cog[i], delta=1e-8, msg=f"[{i}] cog")
                self.assertAlmostEqual(out[1].value, exp_trig[i], delta=1e-8, msg=f"[{i}] trigger")

    def test_update_entity_trade(self):
        """Test trade entity update."""
        cog = _create_default()
        t = _test_time()
        inp = _test_input()
        exp_cog = _test_expected_cog()
        exp_trig = _test_expected_trigger()

        for i in range(len(inp)):
            tr = Trade(time=t, price=inp[i], volume=0)
            out = cog.update_trade(tr)
            self.assertEqual(len(out), 2, f"[{i}] output length")
            if math.isnan(exp_cog[i]):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                self.assertTrue(math.isnan(out[1].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, exp_cog[i], delta=1e-8, msg=f"[{i}] cog")
                self.assertAlmostEqual(out[1].value, exp_trig[i], delta=1e-8, msg=f"[{i}] trigger")

    def test_is_primed(self):
        """Test priming behavior."""
        cog = _create_default()
        inp = _test_input()
        lprimed = 10

        self.assertFalse(cog.is_primed())

        for i in range(lprimed):
            cog.update(inp[i])
            self.assertFalse(cog.is_primed(), f"[{i+1}]")

        for i in range(lprimed, len(inp)):
            cog.update(inp[i])
            self.assertTrue(cog.is_primed(), f"[{i+1}]")

    def test_metadata(self):
        """Test metadata output."""
        cog = _create_default()
        meta = cog.metadata()

        self.assertEqual(meta.mnemonic, "cog(10, hl/2)")
        self.assertEqual(meta.description, "Center of Gravity oscillator cog(10, hl/2)")
        self.assertEqual(len(meta.outputs), 2)

        self.assertEqual(meta.outputs[0].kind, int(CenterOfGravityOscillatorOutput.VALUE))
        self.assertEqual(meta.outputs[0].mnemonic, "cog(10, hl/2)")
        self.assertEqual(meta.outputs[1].kind, int(CenterOfGravityOscillatorOutput.TRIGGER))
        self.assertEqual(meta.outputs[1].mnemonic, "cogTrig(10, hl/2)")

    def test_create_validation(self):
        """Test parameter validation."""
        with self.assertRaises(ValueError):
            CenterOfGravityOscillator.create(CenterOfGravityOscillatorParams(length=0))
        with self.assertRaises(ValueError):
            CenterOfGravityOscillator.create(CenterOfGravityOscillatorParams(length=-8))


if __name__ == '__main__':
    unittest.main()
