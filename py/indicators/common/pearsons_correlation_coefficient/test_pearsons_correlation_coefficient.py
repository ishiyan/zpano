import math
import unittest
from datetime import datetime

from py.indicators.common.pearsons_correlation_coefficient.pearsons_correlation_coefficient import PearsonsCorrelationCoefficient
from py.indicators.common.pearsons_correlation_coefficient.params import PearsonsCorrelationCoefficientParams
from py.indicators.common.pearsons_correlation_coefficient.output import PearsonsCorrelationCoefficientOutput
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar
from py.entities.bar_component import BarComponent

from .test_testdata import HIGH, LOW, EXCEL_EXPECTED

class TestPearsonsCorrelationCoefficient(unittest.TestCase):

    def test_talib_spot_checks_period20(self):
        """TA-Lib spot checks for period=20."""
        c = PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=20))

        for i in range(19):
            result = c.update_pair(HIGH[i], LOW[i])
            if i < 19:
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN")

        # Index 19 is first value (already fed above at i=19 in range(19)... wait, range(19) is 0-18.
        # We need to feed index 19 separately.
        result = c.update_pair(HIGH[19], LOW[19])
        self.assertAlmostEqual(result, 0.9401569, delta=1e-4, msg="[19] spot check")

        result = c.update_pair(HIGH[20], LOW[20])
        self.assertAlmostEqual(result, 0.9471812, delta=1e-4, msg="[20] spot check")

        # Feed remaining up to 251
        for i in range(21, len(HIGH)):
            result = c.update_pair(HIGH[i], LOW[i])

        self.assertAlmostEqual(result, 0.8866901, delta=1e-4, msg="[251] spot check")

    def test_excel_verification_period20(self):
        """Excel verification for period=20 at 1e-10 tolerance."""
        c = PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=20))

        for i in range(19):
            result = c.update_pair(HIGH[i], LOW[i])
            self.assertTrue(math.isnan(result), f"[{i}] expected NaN")

        for i in range(19, len(HIGH)):
            result = c.update_pair(HIGH[i], LOW[i])
            self.assertAlmostEqual(result, EXCEL_EXPECTED[i], delta=1e-10,
                                   msg=f"[{i}] Excel verification")

    def test_nan_input(self):
        """NaN input returns NaN."""
        c = PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=20))
        self.assertTrue(math.isnan(c.update_pair(math.nan, 1.0)))
        self.assertTrue(math.isnan(c.update_pair(1.0, math.nan)))

    def test_update_scalar_constant(self):
        """correl(x,x) with constant value returns 0 (zero variance)."""
        tm = datetime(2021, 4, 1)
        c = PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=2))
        c.update(3.0)
        c.update(3.0)
        out = c.update_scalar(Scalar(tm, 3.0))
        self.assertEqual(len(out), 1)
        self.assertAlmostEqual(out[0].value, 0.0, delta=1e-10)
        self.assertEqual(out[0].time, tm)

    def test_update_bar_high_low(self):
        """update_bar extracts High as X and Low as Y."""
        tm = datetime(2021, 4, 1)
        c = PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=2))
        c.update_pair(10, 5)
        c.update_pair(20, 10)
        bar = Bar(tm, 0.0, 10.0, 5.0, 0.0, 0.0)
        out = c.update_bar(bar)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, tm)
        self.assertFalse(math.isnan(out[0].value))

    def test_update_quote(self):
        """update_quote returns 1 output."""
        tm = datetime(2021, 4, 1)
        c = PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=2))
        c.update(3.0)
        c.update(3.0)
        out = c.update_quote(Quote(tm, 3.0, 3.0, 1.0, 1.0))
        self.assertEqual(len(out), 1)

    def test_update_trade(self):
        """update_trade returns 1 output."""
        tm = datetime(2021, 4, 1)
        c = PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=2))
        c.update(3.0)
        c.update(3.0)
        out = c.update_trade(Trade(tm, 3.0, 100.0))
        self.assertEqual(len(out), 1)

    def test_is_primed_length1(self):
        """is_primed for length=1."""
        c = PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=1))
        self.assertFalse(c.is_primed())
        c.update_pair(HIGH[0], LOW[0])
        self.assertTrue(c.is_primed())

    def test_is_primed_length20(self):
        """is_primed for length=20."""
        c = PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=20))
        self.assertFalse(c.is_primed())

        for i in range(19):
            c.update_pair(HIGH[i], LOW[i])
            self.assertFalse(c.is_primed(), f"[{i}]")

        c.update_pair(HIGH[19], LOW[19])
        self.assertTrue(c.is_primed())

    def test_metadata(self):
        """Test metadata."""
        c = PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=20))
        meta = c.metadata()

        self.assertEqual(meta.identifier, Identifier.PEARSONS_CORRELATION_COEFFICIENT)
        self.assertEqual(meta.mnemonic, "correl(20)")
        self.assertEqual(meta.description, "Pearsons Correlation Coefficient correl(20)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].kind, int(PearsonsCorrelationCoefficientOutput.VALUE))

    def test_invalid_length(self):
        """length < 1 raises ValueError."""
        with self.assertRaises(ValueError):
            PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=0))
        with self.assertRaises(ValueError):
            PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(length=-1))

    def test_custom_components_mnemonic(self):
        """Mnemonic with custom bar component."""
        c = PearsonsCorrelationCoefficient(PearsonsCorrelationCoefficientParams(
            length=20,
            bar_component=BarComponent.MEDIAN,
        ))
        self.assertEqual(c.mnemonic, "correl(20, hl/2)")


if __name__ == '__main__':
    unittest.main()
