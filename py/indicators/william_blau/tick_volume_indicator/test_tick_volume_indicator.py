import math
import unittest
from datetime import datetime

from py.indicators.william_blau.tick_volume_indicator.tick_volume_indicator import TickVolumeIndicator
from py.indicators.william_blau.tick_volume_indicator.params import TickVolumeIndicatorParams
from py.indicators.core.identifier import Identifier
from py.indicators.core.outputs.shape import Shape
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.scalar import Scalar
from py.entities.trade import Trade

from .test_testdata import (
    INPUT_OPEN, INPUT_HIGH, INPUT_LOW, INPUT_CLOSE,
    INPUT_UPTICKS, INPUT_DOWNTICKS,
    EXPECTED_R12_S12_U1,
    EXPECTED_R25_S13_U1,
    EXPECTED_R32_S32_U5,
    EXPECTED_R1_S1_U1,
    EXPECTED_R32_S5_U1,
    EXPECTED_R12_S12_U5,
    EXPECTED_R20_S5_U3,
    EXPECTED_R5_S5_U5,
    EXPECTED_R32_S32_U1,
    EXPECTED_R10_S10_U1,
    EXPECTED_R50_S25_U1,
    EXPECTED_R12_S26_U9,
    EXPECTED_R3_S3_U3,
    EXPECTED_R7_S4_U2,
    EXPECTED_R64_S1_U1,
    EXPECTED_R12_S12_U3,
)

TOLERANCE = 1e-10

# (r, s, u, expected)
COMBOS = [
    (12, 12, 1, EXPECTED_R12_S12_U1),
    (25, 13, 1, EXPECTED_R25_S13_U1),
    (32, 32, 5, EXPECTED_R32_S32_U5),
    (1, 1, 1, EXPECTED_R1_S1_U1),
    (32, 5, 1, EXPECTED_R32_S5_U1),
    (12, 12, 5, EXPECTED_R12_S12_U5),
    (20, 5, 3, EXPECTED_R20_S5_U3),
    (5, 5, 5, EXPECTED_R5_S5_U5),
    (32, 32, 1, EXPECTED_R32_S32_U1),
    (10, 10, 1, EXPECTED_R10_S10_U1),
    (50, 25, 1, EXPECTED_R50_S25_U1),
    (12, 26, 9, EXPECTED_R12_S26_U9),
    (3, 3, 3, EXPECTED_R3_S3_U3),
    (7, 4, 2, EXPECTED_R7_S4_U2),
    (64, 1, 1, EXPECTED_R64_S1_U1),
    (12, 12, 3, EXPECTED_R12_S12_U3),
]


class TestTickVolumeIndicatorData(unittest.TestCase):
    """Test TVI against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for r, s, u, expected in COMBOS:
            with self.subTest(r=r, s=s, u=u):
                ind = TickVolumeIndicator(TickVolumeIndicatorParams(r=r, s=s, u=u))
                for i in range(len(INPUT_UPTICKS)):
                    act = ind.update(INPUT_UPTICKS[i], INPUT_DOWNTICKS[i])

                    if math.isnan(expected[i]):
                        self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
                    else:
                        self.assertAlmostEqual(act, expected[i], delta=TOLERANCE,
                                               msg=f"[{i}] expected {expected[i]}, got {act}")


class TestTickVolumeIndicatorPassthrough(unittest.TestCase):
    """Test the all-passthrough invariant TVI(1,1,1) = 100*(up-down)/(up+down)."""

    def test_passthrough(self):
        ind = TickVolumeIndicator(TickVolumeIndicatorParams(r=1, s=1, u=1))
        self.assertAlmostEqual(ind.update(8.0, 2.0), 60.0, delta=TOLERANCE)
        self.assertAlmostEqual(ind.update(0.0, 5.0), -100.0, delta=TOLERANCE)
        # Flat market: denominator 0 -> division guard 0.0.
        self.assertAlmostEqual(ind.update(0.0, 0.0), 0.0, delta=TOLERANCE)


class TestTickVolumeIndicatorPrimed(unittest.TestCase):
    """The TVI has no NaN warm-up region: it is primed after the first update."""

    def test_primed(self):
        ind = TickVolumeIndicator(TickVolumeIndicatorParams())
        self.assertFalse(ind.is_primed())
        ind.update(INPUT_UPTICKS[0], INPUT_DOWNTICKS[0])
        self.assertTrue(ind.is_primed())

    def test_primed_tick_rule(self):
        ind = TickVolumeIndicator(TickVolumeIndicatorParams())
        self.assertFalse(ind.is_primed())
        ind.update_scalar(Scalar(time=datetime(2021, 4, 1), value=10.0))
        self.assertTrue(ind.is_primed())


class TestTickVolumeIndicatorMnemonic(unittest.TestCase):
    """Test mnemonic generation."""

    def test_default_mnemonic(self):
        ind = TickVolumeIndicator(TickVolumeIndicatorParams())
        self.assertEqual(ind.metadata().mnemonic, "tvi(12,12,1)")

    def test_custom_mnemonic(self):
        ind = TickVolumeIndicator(TickVolumeIndicatorParams(r=32, s=32, u=5))
        self.assertEqual(ind.metadata().mnemonic, "tvi(32,32,5)")


class TestTickVolumeIndicatorMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = TickVolumeIndicator(TickVolumeIndicatorParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.TICK_VOLUME_INDICATOR)
        self.assertEqual(meta.mnemonic, "tvi(12,12,1)")
        self.assertEqual(meta.description, "Tick Volume Indicator tvi(12,12,1)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].kind, 0)
        self.assertEqual(meta.outputs[0].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[0].mnemonic, "tvi(12,12,1)")
        self.assertEqual(meta.outputs[0].description, "Tick Volume Indicator tvi(12,12,1)")


class TestTickVolumeIndicatorUpdateEntities(unittest.TestCase):
    """Test the entity update methods and their upticks/downticks mapping."""

    def test_update_bar(self):
        # A bar maps up = close - low, down = high - close: the fixture proxy.
        ind = TickVolumeIndicator(TickVolumeIndicatorParams(r=12, s=12, u=1))
        tm = datetime(2021, 4, 1)
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_bar(Bar(time=tm, open=INPUT_OPEN[i], high=INPUT_HIGH[i],
                                     low=INPUT_LOW[i], close=INPUT_CLOSE[i], volume=0.0))
            self.assertEqual(len(out), 1)
            self.assertEqual(out[0].time, tm)
            self.assertAlmostEqual(out[0].value, EXPECTED_R12_S12_U1[i], delta=TOLERANCE,
                                   msg=f"[{i}] expected {EXPECTED_R12_S12_U1[i]}, got {out[0].value}")

    def test_update_scalar(self):
        # Tick rule on the value: first -> 0, up -> +100, down -> -100, flat -> 0.
        ind = TickVolumeIndicator(TickVolumeIndicatorParams(r=1, s=1, u=1))
        tm = datetime(2021, 4, 1)
        for value, expected in [(10.0, 0.0), (12.0, 100.0), (11.0, -100.0), (11.0, 0.0)]:
            out = ind.update_scalar(Scalar(time=tm, value=value))
            self.assertEqual(len(out), 1)
            self.assertEqual(out[0].time, tm)
            self.assertAlmostEqual(out[0].value, expected, delta=TOLERANCE)

    def test_update_trade(self):
        # Tick rule on the price.
        ind = TickVolumeIndicator(TickVolumeIndicatorParams(r=1, s=1, u=1))
        tm = datetime(2021, 4, 1)
        for price, expected in [(10.0, 0.0), (12.0, 100.0), (11.0, -100.0), (11.0, 0.0)]:
            out = ind.update_trade(Trade(time=tm, price=price, volume=1.0))
            self.assertEqual(len(out), 1)
            self.assertEqual(out[0].time, tm)
            self.assertAlmostEqual(out[0].value, expected, delta=TOLERANCE)

    def test_update_quote(self):
        # Tick rule on the mid price (bid + ask) / 2.
        ind = TickVolumeIndicator(TickVolumeIndicatorParams(r=1, s=1, u=1))
        tm = datetime(2021, 4, 1)
        for bid, ask, expected in [(9.0, 11.0, 0.0), (11.0, 13.0, 100.0),
                                   (10.0, 12.0, -100.0), (10.5, 11.5, 0.0)]:
            out = ind.update_quote(Quote(time=tm, bid_price=bid, ask_price=ask,
                                         bid_size=1.0, ask_size=1.0))
            self.assertEqual(len(out), 1)
            self.assertEqual(out[0].time, tm)
            self.assertAlmostEqual(out[0].value, expected, delta=TOLERANCE)

    def test_update_scalar_matches_tick_rule(self):
        # Scalar updates equal update(max(d, 0), max(-d, 0)) of the value changes.
        ind = TickVolumeIndicator(TickVolumeIndicatorParams(r=12, s=12, u=3))
        ref = TickVolumeIndicator(TickVolumeIndicatorParams(r=12, s=12, u=3))
        tm = datetime(2021, 4, 1)
        for i in range(len(INPUT_CLOSE)):
            diff = 0.0 if i == 0 else INPUT_CLOSE[i] - INPUT_CLOSE[i - 1]
            expected = ref.update(max(diff, 0.0), max(-diff, 0.0))
            out = ind.update_scalar(Scalar(time=tm, value=INPUT_CLOSE[i]))
            self.assertAlmostEqual(out[0].value, expected, delta=TOLERANCE,
                                   msg=f"[{i}] expected {expected}, got {out[0].value}")


class TestTickVolumeIndicatorInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            TickVolumeIndicator(TickVolumeIndicatorParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            TickVolumeIndicator(TickVolumeIndicatorParams(s=0))

    def test_u_too_small(self):
        with self.assertRaises(ValueError):
            TickVolumeIndicator(TickVolumeIndicatorParams(u=0))


class TestTickVolumeIndicatorDefaultParams(unittest.TestCase):
    """Test default parameters."""

    def test_default_params(self):
        from py.indicators.william_blau.tick_volume_indicator.params import default_params
        p = default_params()
        self.assertEqual((p.r, p.s, p.u), (12, 12, 1))


if __name__ == '__main__':
    unittest.main()
