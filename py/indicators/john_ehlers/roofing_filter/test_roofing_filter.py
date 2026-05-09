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

from .test_testdata import (
    test_input,
    test_expected_71,
    test_expected_72,
    test_expected_73,
)


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
