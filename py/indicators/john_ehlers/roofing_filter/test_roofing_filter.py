"""Tests for the RoofingFilter indicator."""

import math
import unittest
from datetime import datetime

from py.indicators.john_ehlers.roofing_filter.roofing_filter import RoofingFilter
from py.indicators.john_ehlers.roofing_filter.params import RoofingFilterParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade


def test_input():
    return [
        1065.25, 1065.25, 1063.75, 1059.25, 1059.25, 1057.75, 1054, 1056.25, 1058.5, 1059.5,
        1064.75, 1063, 1062.5, 1065, 1061.5, 1058.25, 1058.25, 1061.75, 1062, 1061.25,
        1062.5, 1066.5, 1066.5, 1069.25, 1074.75, 1075, 1076, 1078, 1079.25, 1079.75,
        1078, 1078.75, 1078.25, 1076.5, 1075.75, 1075.75, 1075, 1073.25, 1071, 1083,
        1082.25, 1084, 1085.75, 1085.25, 1085.75, 1087.25, 1089, 1089, 1090, 1095,
        1097.25, 1097.25, 1099, 1098.25, 1093.75, 1095, 1097.25, 1099.25, 1097.5, 1096,
        1095, 1094, 1095.75, 1095.75, 1093.75, 1100.5, 1102.25, 1102, 1102.75, 1105.75,
        1108.25, 1109.5, 1107.25, 1102.5, 1104.75, 1099.25, 1102.75, 1099.5, 1096.75, 1098.25,
        1095.25, 1097, 1097.75, 1100.5, 1099.5, 1101.75, 1101.75, 1102.75, 1099.75, 1097,
        1100.75, 1105.75, 1104.5, 1108.5, 1111.25, 1112.25, 1110, 1109.75, 1108.25, 1106,
    ]


def test_expected_71():
    """1-pole HP, no zero-mean (shortest=10, longest=48)."""
    return [
        0, 0, 0, -0.53, -1.62, -2.72, -4.03, -5.09, -5.05, -4.09,
        -2.20, -0.05, 1.29, 2.14, 2.39, 1.46, -0.05, -0.90, -0.80, -0.41,
        0.03, 0.99, 2.30, 3.60, 5.39, 7.33, 8.69, 9.52, 10.00, 10.11,
        9.59, 8.58, 7.46, 6.12, 4.61, 3.26, 2.16, 1.12, -0.11, 0.12,
        2.14, 4.27, 6.08, 7.22, 7.54, 7.48, 7.46, 7.43, 7.29, 7.64,
        8.69, 9.68, 10.26, 10.32, 9.23, 7.38, 5.98, 5.47, 5.30, 4.74,
        3.77, 2.58, 1.66, 1.28, 0.92, 1.21, 2.62, 4.12, 5.14, 5.97,
        6.95, 7.94, 8.26, 7.16, 5.36, 3.27, 1.36, 0.07, -1.34, -2.48,
        -3.29, -3.79, -3.61, -2.72, -1.53, -0.40, 0.67, 1.49, 1.70, 0.89,
        0.04, 0.47, 1.66, 3.05, 4.81, 6.48, 7.28, 7.00, 5.99, 4.62,
    ]


def test_expected_72():
    """1-pole HP, zero-mean (shortest=10, longest=48)."""
    return [
        0, 0, 0, -0.50, -1.46, -2.31, -3.26, -3.85, -3.34, -2.02,
        -0.01, 2.01, 3.02, 3.45, 3.26, 1.99, 0.33, -0.52, -0.35, 0.05,
        0.46, 1.31, 2.37, 3.30, 4.57, 5.84, 6.39, 6.38, 6.05, 5.41,
        4.26, 2.79, 1.39, -0.04, -1.45, -2.54, -3.26, -3.83, -4.51, -3.74,
        -1.39, 0.78, 2.39, 3.16, 3.07, 2.64, 2.29, 1.98, 1.61, 1.74,
        2.51, 3.13, 3.29, 2.94, 1.56, -0.37, -1.64, -1.91, -1.84, -2.14,
        -2.79, -3.56, -3.98, -3.85, -3.72, -2.99, -1.29, 0.27, 1.19, 1.83,
        2.53, 3.14, 3.06, 1.65, -0.25, -2.17, -3.70, -4.46, -5.23, -5.66,
        -5.72, -5.49, -4.65, -3.23, -1.72, -0.45, 0.61, 1.30, 1.34, 0.42,
        -0.43, 0.02, 1.14, 2.30, 3.67, 4.79, 4.95, 4.08, 2.63, 0.80,
    ]


def test_expected_73():
    """2-pole HP (shortest=40, longest=80)."""
    return [
        0, 0, 0, -0.03, -0.10, -0.17, -0.28, -0.37, -0.38, -0.27,
        0.03, 0.52, 1.13, 1.85, 2.62, 3.37, 4.04, 4.69, 5.35, 6.00,
        6.63, 7.29, 7.99, 8.71, 9.52, 10.42, 11.34, 12.27, 13.19, 14.07,
        14.85, 15.49, 15.99, 16.32, 16.45, 16.40, 16.19, 15.82, 15.27, 14.69,
        14.20, 13.79, 13.45, 13.16, 12.90, 12.66, 12.45, 12.26, 12.07, 11.93,
        11.88, 11.88, 11.91, 11.94, 11.88, 11.69, 11.41, 11.11, 10.76, 10.33,
        9.80, 9.17, 8.47, 7.75, 6.99, 6.26, 5.64, 5.14, 4.71, 4.37,
        4.16, 4.07, 4.02, 3.93, 3.77, 3.50, 3.13, 2.68, 2.13, 1.49,
        0.79, 0.05, -0.67, -1.31, -1.86, -2.31, -2.65, -2.89, -3.06, -3.24,
        -3.40, -3.46, -3.39, -3.21, -2.88, -2.41, -1.89, -1.37, -0.91, -0.51,
    ]


def create_1pole():
    return RoofingFilter.create(RoofingFilterParams(
        shortest_cycle_period=10, longest_cycle_period=48))


def create_1pole_zero_mean():
    return RoofingFilter.create(RoofingFilterParams(
        shortest_cycle_period=10, longest_cycle_period=48, has_zero_mean=True))


def create_2pole():
    return RoofingFilter.create(RoofingFilterParams(
        shortest_cycle_period=40, longest_cycle_period=80, has_two_pole_highpass_filter=True))


class TestRoofingFilterUpdate1Pole(unittest.TestCase):
    def test_update(self):
        skip_rows = 30
        tolerance = 0.5
        inp = test_input()
        expected = test_expected_71()
        rf = create_1pole()

        for i in range(len(inp)):
            act = rf.update(inp[i])
            if i < 3:
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
                continue
            if i < skip_rows:
                continue
            self.assertAlmostEqual(act, expected[i], delta=tolerance,
                                   msg=f"[{i}] expected {expected[i]}, got {act}")

    def test_nan_passthrough(self):
        rf = create_1pole()
        self.assertTrue(math.isnan(rf.update(math.nan)))


class TestRoofingFilterUpdate1PoleZeroMean(unittest.TestCase):
    def test_update(self):
        skip_rows = 30
        tolerance = 0.5
        inp = test_input()
        expected = test_expected_72()
        rf = create_1pole_zero_mean()

        for i in range(len(inp)):
            act = rf.update(inp[i])
            if i < 4:
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
                continue
            if i < skip_rows:
                continue
            self.assertAlmostEqual(act, expected[i], delta=tolerance,
                                   msg=f"[{i}] expected {expected[i]}, got {act}")


class TestRoofingFilterUpdate2Pole(unittest.TestCase):
    def test_update(self):
        skip_rows = 30
        tolerance = 0.5
        inp = test_input()
        expected = test_expected_73()
        rf = create_2pole()

        for i in range(len(inp)):
            act = rf.update(inp[i])
            if i < 4:
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
                continue
            if i < skip_rows:
                continue
            self.assertAlmostEqual(act, expected[i], delta=tolerance,
                                   msg=f"[{i}] expected {expected[i]}, got {act}")


class TestRoofingFilterIsPrimed(unittest.TestCase):
    def test_1pole(self):
        inp = test_input()
        rf = create_1pole()
        self.assertFalse(rf.is_primed())
        for i in range(3):
            rf.update(inp[i])
            self.assertFalse(rf.is_primed(), f"[{i}] should not be primed")
        rf.update(inp[3])
        self.assertTrue(rf.is_primed(), "[3] should be primed")

    def test_1pole_zero_mean(self):
        inp = test_input()
        rf = create_1pole_zero_mean()
        for i in range(4):
            rf.update(inp[i])
            self.assertFalse(rf.is_primed(), f"[{i}] should not be primed")
        rf.update(inp[4])
        self.assertTrue(rf.is_primed(), "[4] should be primed")

    def test_2pole(self):
        inp = test_input()
        rf = create_2pole()
        for i in range(4):
            rf.update(inp[i])
            self.assertFalse(rf.is_primed(), f"[{i}] should not be primed")
        rf.update(inp[4])
        self.assertTrue(rf.is_primed(), "[4] should be primed")


class TestRoofingFilterUpdateEntity(unittest.TestCase):
    def test_update_scalar(self):
        rf = create_1pole()
        for _ in range(4):
            rf.update(100.0)
        t = datetime(2021, 4, 1)
        s = Scalar(time=t, value=100.0)
        out = rf.update_scalar(s)
        self.assertEqual(len(out), 1)
        self.assertIsInstance(out[0], Scalar)
        self.assertEqual(out[0].time, t)
        self.assertFalse(math.isnan(out[0].value))

    def test_update_bar(self):
        rf = create_1pole()
        for _ in range(4):
            rf.update(100.0)
        t = datetime(2021, 4, 1)
        b = Bar(time=t, open=100.0, high=100.0, low=100.0, close=100.0, volume=0.0)
        out = rf.update_bar(b)
        self.assertEqual(len(out), 1)
        self.assertIsInstance(out[0], Scalar)

    def test_update_quote(self):
        rf = create_1pole()
        for _ in range(4):
            rf.update(100.0)
        t = datetime(2021, 4, 1)
        q = Quote(time=t, bid_price=100.0, ask_price=100.0, bid_size=1.0, ask_size=1.0)
        out = rf.update_quote(q)
        self.assertEqual(len(out), 1)
        self.assertIsInstance(out[0], Scalar)

    def test_update_trade(self):
        rf = create_1pole()
        for _ in range(4):
            rf.update(100.0)
        t = datetime(2021, 4, 1)
        r = Trade(time=t, price=100.0, volume=0.0)
        out = rf.update_trade(r)
        self.assertEqual(len(out), 1)
        self.assertIsInstance(out[0], Scalar)


class TestRoofingFilterMetadata(unittest.TestCase):
    def test_metadata(self):
        rf = create_1pole()
        m = rf.metadata()
        self.assertEqual(m.identifier, Identifier.ROOFING_FILTER)
        self.assertEqual(len(m.outputs), 1)
        self.assertEqual(m.outputs[0].mnemonic, "roof1hp(10, 48, hl/2)")
        self.assertEqual(m.outputs[0].description, "Roofing Filter roof1hp(10, 48, hl/2)")


class TestRoofingFilterConstruction(unittest.TestCase):
    def test_valid_params(self):
        rf = create_1pole()
        self.assertFalse(rf.is_primed())

    def test_2pole_mnemonic(self):
        rf = RoofingFilter.create(RoofingFilterParams(
            shortest_cycle_period=10, longest_cycle_period=48,
            has_two_pole_highpass_filter=True))
        m = rf.metadata()
        self.assertEqual(m.outputs[0].mnemonic, "roof2hp(10, 48, hl/2)")

    def test_zero_mean_mnemonic(self):
        rf = RoofingFilter.create(RoofingFilterParams(
            shortest_cycle_period=10, longest_cycle_period=48,
            has_zero_mean=True))
        m = rf.metadata()
        self.assertEqual(m.outputs[0].mnemonic, "roof1hpzm(10, 48, hl/2)")

    def test_shortest_too_small(self):
        with self.assertRaises(ValueError):
            RoofingFilter.create(RoofingFilterParams(
                shortest_cycle_period=1, longest_cycle_period=48))

    def test_longest_not_greater(self):
        with self.assertRaises(ValueError):
            RoofingFilter.create(RoofingFilterParams(
                shortest_cycle_period=10, longest_cycle_period=10))


if __name__ == '__main__':
    unittest.main()
