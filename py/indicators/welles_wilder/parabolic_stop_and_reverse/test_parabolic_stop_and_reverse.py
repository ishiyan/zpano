import math
import unittest
from datetime import datetime, timedelta

from py.indicators.welles_wilder.parabolic_stop_and_reverse.parabolic_stop_and_reverse import ParabolicStopAndReverse
from py.indicators.welles_wilder.parabolic_stop_and_reverse.params import ParabolicStopAndReverseParams
from py.indicators.welles_wilder.parabolic_stop_and_reverse.output import ParabolicStopAndReverseOutput
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.scalar import Scalar

from .test_testdata import test_expected


def wilder_highs():
    return [
        51.12,
        52.35, 52.1, 51.8, 52.1, 52.5, 52.8, 52.5, 53.5, 53.5, 53.8, 54.2, 53.4, 53.5,
        54.4, 55.2, 55.7, 57, 57.5, 58, 57.7, 58, 57.5, 57, 56.7, 57.5,
        56.70, 56.00, 56.20, 54.80, 55.50, 54.70, 54.00, 52.50, 51.00, 51.50, 51.70, 53.00,
    ]


def wilder_lows():
    return [
        50.0,
        51.5, 51, 50.5, 51.25, 51.7, 51.85, 51.5, 52.3, 52.5, 53, 53.5, 52.5, 52.1, 53,
        54, 55, 56, 56.5, 57, 56.5, 57.3, 56.7, 56.3, 56.2, 56,
        55.50, 55.00, 54.90, 54.00, 54.50, 53.80, 53.00, 51.50, 50.00, 50.50, 50.20, 51.50,
    ]


def test_highs():
    return [
        93.25, 94.94, 96.375, 96.19, 96, 94.72, 95, 93.72, 92.47, 92.75,
        96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,
        90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,
        87.155, 87.25, 87.375, 88.97, 90, 89.845, 86.97, 85.94, 84.75, 85.47,
        84.47, 88.5, 89.47, 90, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,
        90, 88.875, 89, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,
        90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92, 91.94,
        89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,
        104.5, 105.5, 106.125, 107.94, 106.25, 107, 108.75, 110.94, 110.94, 114.22,
        123, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113, 118.315,
        116.87, 116.75, 113.87, 114.62, 115.31, 116, 121.69, 119.87, 120.87, 116.75,
        116.5, 116, 118.31, 121.5, 122, 121.44, 125.75, 127.75, 124.19, 124.44,
        125.75, 124.69, 125.31, 132, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,
        138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,
        125.12, 126.5, 128.69, 126.62, 126.69, 126, 123.12, 121.87, 124, 127,
        124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,
        124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126, 128.5, 126.75, 129.75,
        132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135, 132.38, 131.44, 130.88,
        129.63, 127.25, 127.81, 125, 126.81, 124.75, 122.81, 122.25, 121.06, 120,
        123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109,
        108.5, 113.06, 93, 94.62, 95.12, 96, 95.56, 95.31, 99, 98.81,
        96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,
        95, 94.87, 98.25, 105.12, 108.44, 109.87, 105, 106, 104.94, 104.5,
        104.44, 106.31, 112.87, 116.5, 119.19, 121, 122.12, 111.94, 112.75, 110.19,
        107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110, 110.75, 110.5,
        110.5, 109.5,
    ]


def test_lows():
    return [
        90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92, 89.75, 89.44, 90.625,
        92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,
        87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,
        84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,
        82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89, 88.565, 90.095,
        89, 86.47, 84, 83.315, 82, 83.25, 84.75, 85.28, 87.19, 88.44,
        88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92, 90.53, 89.97, 88.815,
        86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,
        101.72, 101.72, 103.155, 105.69, 103.655, 104, 105.53, 108.53, 108.75, 107.75,
        117, 118, 116, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,
        112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,
        112.56, 114, 114.75, 118.87, 119, 119.75, 122.62, 123, 121.75, 121.56,
        123.12, 122.19, 122.75, 124.37, 128, 129.5, 130.81, 130.63, 132.13, 133.88,
        135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,
        122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121,
        121.12, 118.94, 119.81, 121, 122, 124.5, 126.56, 123.5, 121.25, 121.06,
        122.31, 121, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128,
        128.38, 130.69, 131.63, 134.38, 132, 131.94, 131.94, 129.56, 123.75, 126,
        126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120, 119.62, 115.5, 116.75,
        119.06, 119.06, 115.06, 111.06, 113.12, 110, 105, 104.69, 103.87, 104.69,
        105.44, 107, 89, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,
        93.69, 93.5, 90, 90.19, 90.5, 92.12, 94.12, 94.87, 93, 93.87,
        93, 92.62, 93.56, 98.37, 104.44, 106, 101.81, 104.12, 103.37, 102.12,
        102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,
        104.5, 105.75, 108.62, 107.75, 108.06, 108, 108.19, 108.12, 109.06, 108.75,
        108.56, 106.62,
    ]


class TestParabolicStopAndReverse(unittest.TestCase):

    def test_252_bar(self):
        """252-bar test with default params."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams())
        highs = test_highs()
        lows = test_lows()
        expected = test_expected()

        for i in range(len(highs)):
            result = sar.update_hl(highs[i], lows[i])

            if math.isnan(expected[i]):
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")
                continue

            self.assertAlmostEqual(result, expected[i], delta=1e-6,
                                   msg=f"[{i}] expected {expected[i]:.10f}, got {result:.10f}")

    def test_wilder_spot_checks(self):
        """Wilder's original SAR data with spot checks."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams())
        highs = wilder_highs()
        lows = wilder_lows()
        results = []

        for i in range(len(highs)):
            results.append(sar.update_hl(highs[i], lows[i]))

        spot_checks = [
            (0, 50.00),
            (1, 50.047),
            (4, 50.182),
            (35, 52.93),
            (36, 50.00),
        ]

        for out_index, expected in spot_checks:
            actual = abs(results[out_index + 1])  # +1 because results[0] = NaN
            self.assertAlmostEqual(actual, expected, delta=1e-3,
                                   msg=f"Wilder spot check output[{out_index}]: "
                                       f"expected {expected:.4f}, got {actual:.4f}")

    def test_is_primed(self):
        """IsPrimed transitions: not primed -> not primed after 1 bar -> primed after 2."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams())

        self.assertFalse(sar.is_primed())

        sar.update_hl(93.25, 90.75)
        self.assertFalse(sar.is_primed())

        sar.update_hl(94.94, 91.405)
        self.assertTrue(sar.is_primed())

    def test_metadata(self):
        """Metadata: identifier, mnemonic, output count."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams())
        meta = sar.metadata()

        self.assertEqual(meta.identifier, Identifier.PARABOLIC_STOP_AND_REVERSE)
        self.assertEqual(meta.mnemonic, "sar()")
        self.assertEqual(len(meta.outputs), 1)

    def test_constructor_validation_defaults(self):
        """Default params are valid."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams())
        self.assertIsNotNone(sar)

    def test_constructor_validation_negative_long_init(self):
        """Negative long init acceleration raises ValueError."""
        with self.assertRaises(ValueError):
            ParabolicStopAndReverse(ParabolicStopAndReverseParams(acceleration_init_long=-0.01))

    def test_constructor_validation_negative_short_step(self):
        """Negative short step acceleration raises ValueError."""
        with self.assertRaises(ValueError):
            ParabolicStopAndReverse(ParabolicStopAndReverseParams(acceleration_short=-0.01))

    def test_constructor_validation_negative_offset(self):
        """Negative offset on reverse raises ValueError."""
        with self.assertRaises(ValueError):
            ParabolicStopAndReverse(ParabolicStopAndReverseParams(offset_on_reverse=-0.01))

    def test_constructor_validation_custom_valid(self):
        """Custom valid params are accepted."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams(
            acceleration_init_long=0.01,
            acceleration_long=0.01,
            acceleration_max_long=0.10,
            acceleration_init_short=0.03,
            acceleration_short=0.03,
            acceleration_max_short=0.30,
        ))
        self.assertIsNotNone(sar)

    def test_constructor_validation_positive_start_value(self):
        """Positive start value is accepted."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams(start_value=100.0))
        self.assertIsNotNone(sar)

    def test_constructor_validation_negative_start_value(self):
        """Negative start value is accepted."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams(start_value=-100.0))
        self.assertIsNotNone(sar)

    def test_update_bar(self):
        """Feed two bars via entity update; first is NaN, second is valid."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams())
        now = datetime.now()

        bar1 = Bar(time=now, open=91, high=93.25, low=90.75, close=91.5, volume=1000)
        out1 = sar.update_bar(bar1)
        scalar1 = out1[0]
        self.assertTrue(math.isnan(scalar1.value), f"expected NaN for first bar, got {scalar1.value}")

        bar2 = Bar(time=now + timedelta(minutes=1), open=92, high=94.94, low=91.405, close=94.815, volume=1000)
        out2 = sar.update_bar(bar2)
        scalar2 = out2[0]
        self.assertFalse(math.isnan(scalar2.value), f"expected valid value for second bar, got NaN")

    def test_nan_handling(self):
        """NaN input returns NaN but doesn't corrupt state."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams())

        sar.update_hl(93.25, 90.75)
        sar.update_hl(94.94, 91.405)

        result = sar.update_hl(math.nan, 92.0)
        self.assertTrue(math.isnan(result), f"expected NaN for NaN input, got {result}")

        result = sar.update_hl(96.375, 94.25)
        self.assertFalse(math.isnan(result), "expected valid output after NaN, got NaN")

    def test_forced_start_long(self):
        """With start_value=85.0, second bar should be positive (long)."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams(start_value=85.0))
        highs = test_highs()
        lows = test_lows()

        result = sar.update_hl(highs[0], lows[0])
        self.assertTrue(math.isnan(result), f"expected NaN for first bar, got {result}")

        result = sar.update_hl(highs[1], lows[1])
        self.assertGreater(result, 0, f"expected positive (long) SAR with forced long start, got {result}")

    def test_forced_start_short(self):
        """With start_value=-100.0, second bar should be negative (short)."""
        sar = ParabolicStopAndReverse(ParabolicStopAndReverseParams(start_value=-100.0))
        highs = test_highs()
        lows = test_lows()

        result = sar.update_hl(highs[0], lows[0])
        self.assertTrue(math.isnan(result), f"expected NaN for first bar, got {result}")

        result = sar.update_hl(highs[1], lows[1])
        self.assertLess(result, 0, f"expected negative (short) SAR with forced short start, got {result}")


if __name__ == '__main__':
    unittest.main()
