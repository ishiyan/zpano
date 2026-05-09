import math
import unittest
from datetime import datetime

from py.indicators.common.linear_regression.linear_regression import LinearRegression
from py.indicators.common.linear_regression.params import LinearRegressionParams
from py.indicators.common.linear_regression.output import LinearRegressionOutput
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar
from py.entities.bar_component import BarComponent
from py.entities.quote_component import QuoteComponent
from py.entities.trade_component import TradeComponent

from .test_testdata import (
    INPUT,
    EXPECTED_VALUE,
    EXPECTED_FORECAST,
    EXPECTED_INTERCEPT,
    EXPECTED_SLOPE_RAD,
    EXPECTED_SLOPE_DEG,
)

class TestLinearRegression(unittest.TestCase):

    def test_update_period14_value_output(self):
        """Test the Value output for period 14 over all 252 rows."""
        lr = LinearRegression(LinearRegressionParams(length=14))

        # First 13 samples (indices 0-12) should return NaN.
        for i in range(13):
            result = lr.update(INPUT[i])
            self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")

        # Indices 13-251 should match expected values.
        for i in range(13, len(INPUT)):
            result = lr.update(INPUT[i])
            self.assertAlmostEqual(result, EXPECTED_VALUE[i], delta=1e-4,
                                   msg=f"[{i}] Value mismatch")

        # NaN input should return NaN.
        self.assertTrue(math.isnan(lr.update(math.nan)))

    def test_update_entity_all_5_outputs(self):
        """Test all 5 outputs via update_scalar for period 14 over all 252 rows."""
        lr = LinearRegression(LinearRegressionParams(length=14))
        tm = datetime(2021, 4, 1)

        # Feed first 12 samples via update.
        for i in range(12):
            lr.update(INPUT[i])

        # Feed index 12 via update_scalar — should get 5 NaN outputs.
        out = lr.update_scalar(Scalar(tm, INPUT[12]))
        self.assertEqual(len(out), 5)
        for j in range(5):
            self.assertTrue(math.isnan(out[j].value), f"output[{j}] expected NaN")

        # Feed indices 13-251 via update_scalar and verify all 5 outputs.
        for i in range(13, len(INPUT)):
            out = lr.update_scalar(Scalar(tm, INPUT[i]))
            self.assertEqual(len(out), 5)
            self.assertAlmostEqual(out[0].value, EXPECTED_VALUE[i], delta=1e-4,
                                   msg=f"[{i}] Value")
            self.assertAlmostEqual(out[1].value, EXPECTED_FORECAST[i], delta=1e-4,
                                   msg=f"[{i}] Forecast")
            self.assertAlmostEqual(out[2].value, EXPECTED_INTERCEPT[i], delta=1e-4,
                                   msg=f"[{i}] Intercept")
            self.assertAlmostEqual(out[3].value, EXPECTED_SLOPE_RAD[i], delta=1e-4,
                                   msg=f"[{i}] SlopeRad")
            self.assertAlmostEqual(out[4].value, EXPECTED_SLOPE_DEG[i], delta=1e-4,
                                   msg=f"[{i}] SlopeDeg")

    def test_update_bar(self):
        """Test update_bar returns 5 outputs."""
        lr = LinearRegression(LinearRegressionParams(length=14))
        tm = datetime(2021, 4, 1)

        # Prime with 14 samples.
        for i in range(14):
            lr.update(INPUT[i])

        bar = Bar(tm, 0.0, 0.0, 0.0, INPUT[14], 0.0)
        out = lr.update_bar(bar)
        self.assertEqual(len(out), 5)
        self.assertEqual(out[0].time, tm)

    def test_update_quote(self):
        """Test update_quote returns 5 outputs."""
        lr = LinearRegression(LinearRegressionParams(length=14))
        tm = datetime(2021, 4, 1)

        for i in range(14):
            lr.update(INPUT[i])

        quote = Quote(tm, INPUT[14], INPUT[14], 1.0, 1.0)
        out = lr.update_quote(quote)
        self.assertEqual(len(out), 5)

    def test_update_trade(self):
        """Test update_trade returns 5 outputs."""
        lr = LinearRegression(LinearRegressionParams(length=14))
        tm = datetime(2021, 4, 1)

        for i in range(14):
            lr.update(INPUT[i])

        trade = Trade(tm, INPUT[14], 100.0)
        out = lr.update_trade(trade)
        self.assertEqual(len(out), 5)

    def test_is_primed_length14(self):
        """Test is_primed for length=14."""
        lr = LinearRegression(LinearRegressionParams(length=14))

        self.assertFalse(lr.is_primed())

        for i in range(13):
            lr.update(INPUT[i])
            self.assertFalse(lr.is_primed(), f"[{i}] should not be primed")

        for i in range(13, len(INPUT)):
            lr.update(INPUT[i])
            self.assertTrue(lr.is_primed(), f"[{i}] should be primed")

    def test_is_primed_length2(self):
        """Test is_primed for length=2."""
        lr = LinearRegression(LinearRegressionParams(length=2))

        self.assertFalse(lr.is_primed())
        lr.update(INPUT[0])
        self.assertFalse(lr.is_primed())
        lr.update(INPUT[1])
        self.assertTrue(lr.is_primed())

    def test_metadata(self):
        """Test metadata output."""
        lr = LinearRegression(LinearRegressionParams(length=14))
        meta = lr.metadata()

        self.assertEqual(meta.identifier, Identifier.LINEAR_REGRESSION)
        self.assertEqual(meta.mnemonic, "linreg(14)")
        self.assertEqual(meta.description, "Linear Regression linreg(14)")
        self.assertEqual(len(meta.outputs), 5)
        self.assertEqual(meta.outputs[0].kind, int(LinearRegressionOutput.VALUE))
        self.assertEqual(meta.outputs[1].kind, int(LinearRegressionOutput.FORECAST))
        self.assertEqual(meta.outputs[2].kind, int(LinearRegressionOutput.INTERCEPT))
        self.assertEqual(meta.outputs[3].kind, int(LinearRegressionOutput.SLOPE_RAD))
        self.assertEqual(meta.outputs[4].kind, int(LinearRegressionOutput.SLOPE_DEG))

    def test_invalid_length(self):
        """Test that length < 2 raises ValueError."""
        with self.assertRaises(ValueError):
            LinearRegression(LinearRegressionParams(length=1))
        with self.assertRaises(ValueError):
            LinearRegression(LinearRegressionParams(length=0))

    def test_custom_components_mnemonic(self):
        """Test mnemonic with custom components."""
        lr = LinearRegression(LinearRegressionParams(
            length=14,
            bar_component=BarComponent.MEDIAN,
            quote_component=QuoteComponent.MID,
            trade_component=TradeComponent.PRICE,
        ))
        self.assertEqual(lr._mnemonic, "linreg(14, hl/2)")


if __name__ == '__main__':
    unittest.main()
