"""Tests for Ehlers' RoofingFilter indicator."""

from datetime import datetime
import math
import unittest

from py.entities.bar import Bar
from py.entities.bar_component import BarComponent
from py.entities.quote import Quote
from py.entities.quote_component import QuoteComponent
from py.entities.scalar import Scalar
from py.entities.trade import Trade
from py.entities.trade_component import TradeComponent
from py.indicators.core.identifier import Identifier
from py.indicators.core.outputs.shape import Shape
from py.indicators.factory.factory import create_indicator
from py.indicators.john_ehlers.roofing_filter import (
    RoofingFilter,
    RoofingFilterOutput,
    RoofingFilterParams,
    default_params,
    roofing_filter_mnemonic,
)
from py.indicators.john_ehlers.roofing_filter import test_testdata as td


REFERENCE_CASES = (
    (RoofingFilterParams(10, 48, 2, False), td.EXPECTED_LP10_HP48_PO2_ZM0, 4),
    (RoofingFilterParams(40, 80, 2, False), td.EXPECTED_LP40_HP80_PO2_ZM0, 4),
    (RoofingFilterParams(2, 3, 2, False), td.EXPECTED_LP2_HP3_PO2_ZM0, 4),
    (RoofingFilterParams(5, 24, 2, False), td.EXPECTED_LP5_HP24_PO2_ZM0, 4),
    (RoofingFilterParams(10.5, 48, 2, False), td.EXPECTED_LP10_5_HP48_PO2_ZM0, 4),
    (RoofingFilterParams(10, 11, 2, False), td.EXPECTED_LP10_HP11_PO2_ZM0, 4),
    (RoofingFilterParams(20, 80, 2, False), td.EXPECTED_LP20_HP80_PO2_ZM0, 4),
    (RoofingFilterParams(40, 252, 2, False), td.EXPECTED_LP40_HP252_PO2_ZM0, 4),
    (RoofingFilterParams(10, 48, 1, False), td.EXPECTED_LP10_HP48_PO1_ZM0, 3),
    (RoofingFilterParams(2, 4, 1, False), td.EXPECTED_LP2_HP4_PO1_ZM0, 3),
    (RoofingFilterParams(10, 48, 1, True), td.EXPECTED_LP10_HP48_PO1_ZM1, 4),
    (RoofingFilterParams(2, 4, 1, True), td.EXPECTED_LP2_HP4_PO1_ZM1, 4),
)


class TestRoofingFilterFixtures(unittest.TestCase):
    def test_all_twelve_reference_cases(self) -> None:
        self.assertEqual(len(td.INPUT_CLOSE), 252)
        for params, expected, expected_warmup in REFERENCE_CASES:
            with self.subTest(mnemonic=roofing_filter_mnemonic(params)):
                indicator = RoofingFilter(params)
                self.assertEqual(indicator.warmup_bars(), expected_warmup)
                self.assertFalse(indicator.is_primed())
                self.assertTrue(math.isnan(indicator.value()))
                self.assertEqual(len(expected), 252)

                for index, sample in enumerate(td.INPUT_CLOSE):
                    actual = indicator.update(sample)
                    target = expected[index]
                    if index < expected_warmup:
                        self.assertFalse(indicator.is_primed(), f"bar {index} unexpectedly primed")
                        self.assertTrue(math.isnan(actual), f"bar {index} expected NaN, got {actual}")
                        self.assertTrue(math.isnan(indicator.value()))
                    else:
                        self.assertTrue(indicator.is_primed(), f"bar {index} expected primed")
                        self.assertFalse(math.isnan(target), f"bar {index} fixture unexpectedly NaN")
                        bound = max(1e-11, abs(target) * 1e-11)
                        self.assertAlmostEqual(
                            actual, target, delta=bound, msg=f"bar {index}: {actual} vs {target}"
                        )
                        self.assertAlmostEqual(indicator.value(), actual, delta=1e-15)

    def test_nan_skips_state_and_reset_reproduces_start(self) -> None:
        for params, _, _ in REFERENCE_CASES:
            with self.subTest(mnemonic=roofing_filter_mnemonic(params)):
                first = RoofingFilter(params)
                second = RoofingFilter(params)
                for sample in td.INPUT_CLOSE:
                    self.assertTrue(math.isnan(first.update(math.nan)))
                    left = first.update(sample)
                    right = second.update(sample)
                    self.assertEqual(first.is_primed(), second.is_primed())
                    if math.isnan(right):
                        self.assertTrue(math.isnan(left))
                    else:
                        self.assertAlmostEqual(left, right, delta=1e-15)

                first.reset()
                self.assertFalse(first.is_primed())
                self.assertTrue(math.isnan(first.value()))
                third = RoofingFilter(params)
                for index, sample in enumerate(td.INPUT_CLOSE):
                    left = first.update(sample)
                    right = third.update(sample)
                    if math.isnan(right):
                        self.assertTrue(math.isnan(left), f"bar {index} expected NaN after reset")
                    else:
                        self.assertAlmostEqual(left, right, delta=1e-15)


class TestRoofingFilterInvariants(unittest.TestCase):
    def test_constant_input_and_priming(self) -> None:
        indicator = RoofingFilter()
        outputs = [indicator.update(7.0) for _ in range(500)]
        first_finite = next(index for index, item in enumerate(outputs) if math.isfinite(item))
        self.assertEqual(first_finite, indicator.warmup_bars())
        self.assertAlmostEqual(outputs[-1], 0.0, delta=1e-12)

    def test_two_pole_ramp_rejection_and_period_four_boundary(self) -> None:
        indicator = RoofingFilter()
        outputs = [indicator.update(float(index)) for index in range(1000)]
        self.assertAlmostEqual(outputs[-1], 0.0, delta=1e-12)

        boundary = RoofingFilter(RoofingFilterParams(2.0, 4.0, 1, False))
        self.assertAlmostEqual(boundary._gain, 0.5, delta=1e-15)
        self.assertAlmostEqual(boundary._decay, 0.0, delta=1e-15)


class TestRoofingFilterValidationAndMetadata(unittest.TestCase):
    def test_invalid_configuration_and_infinite_sample(self) -> None:
        invalid_cases = (
            RoofingFilterParams(shortest_cycle_period=1.0),
            RoofingFilterParams(shortest_cycle_period=0.0),
            RoofingFilterParams(shortest_cycle_period=-2.0),
            RoofingFilterParams(shortest_cycle_period=math.nan),
            RoofingFilterParams(shortest_cycle_period=math.inf),
            RoofingFilterParams(shortest_cycle_period=10.0, longest_cycle_period=10.0),
            RoofingFilterParams(shortest_cycle_period=10.0, longest_cycle_period=5.0),
            RoofingFilterParams(shortest_cycle_period=10.0, longest_cycle_period=math.inf),
            RoofingFilterParams(shortest_cycle_period=10.0, longest_cycle_period=math.nan),
            RoofingFilterParams(highpass_poles=0),
            RoofingFilterParams(highpass_poles=3),
            RoofingFilterParams(highpass_poles=True),  # type: ignore[arg-type]
            RoofingFilterParams(highpass_poles=2, zero_mean=True),
            RoofingFilterParams(shortest_cycle_period=True),  # type: ignore[arg-type]
            RoofingFilterParams(zero_mean=1),  # type: ignore[arg-type]
        )
        for invalid_params in invalid_cases:
            with self.assertRaises(ValueError):
                RoofingFilter(invalid_params)

        indicator = RoofingFilter()
        for bad_sample in (math.inf, -math.inf):
            with self.assertRaises(ValueError):
                indicator.update(bad_sample)
        self.assertFalse(indicator.is_primed())

    def test_metadata_and_mnemonics(self) -> None:
        indicator = RoofingFilter(default_params())
        meta = indicator.metadata()
        self.assertEqual(meta.identifier, Identifier.ROOFING_FILTER)
        self.assertEqual(meta.mnemonic, "roof2hp(10, 48)")
        self.assertEqual(meta.description, "Roofing Filter roof2hp(10, 48)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].kind, int(RoofingFilterOutput.VALUE))
        self.assertEqual(meta.outputs[0].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[0].mnemonic, "roof2hp(10, 48)")
        self.assertEqual(meta.outputs[0].description, "Roofing Filter roof2hp(10, 48)")

        mnemonic_cases = (
            (RoofingFilterParams(), "roof2hp(10, 48)"),
            (RoofingFilterParams(10.5, 48.0, 2, False), "roof2hp(10.5, 48)"),
            (RoofingFilterParams(10.0, 48.0, 1, False), "roof1hp(10, 48)"),
            (RoofingFilterParams(10.0, 48.0, 1, True), "roof1hpzm(10, 48)"),
            (
                RoofingFilterParams(bar_component=BarComponent.MEDIAN),
                "roof2hp(10, 48, hl/2)",
            ),
            (
                RoofingFilterParams(quote_component=QuoteComponent.BID),
                "roof2hp(10, 48, b)",
            ),
            (
                RoofingFilterParams(trade_component=TradeComponent.VOLUME),
                "roof2hp(10, 48, v)",
            ),
            (
                RoofingFilterParams(
                    bar_component=BarComponent.OPEN,
                    quote_component=QuoteComponent.ASK,
                ),
                "roof2hp(10, 48, o, a)",
            ),
            (
                RoofingFilterParams(
                    bar_component=BarComponent.TYPICAL,
                    trade_component=TradeComponent.VOLUME,
                ),
                "roof2hp(10, 48, hlc/3, v)",
            ),
            (
                RoofingFilterParams(
                    quote_component=QuoteComponent.WEIGHTED,
                    trade_component=TradeComponent.VOLUME,
                ),
                "roof2hp(10, 48, (bbs+aas)/(bs+as), v)",
            ),
        )
        for params, expected_mnemonic in mnemonic_cases:
            self.assertEqual(roofing_filter_mnemonic(params), expected_mnemonic)
            self.assertEqual(RoofingFilter(params).metadata().mnemonic, expected_mnemonic)

    def test_entity_updates_and_factory(self) -> None:
        bar_indicator = RoofingFilter()
        scalar_indicator = RoofingFilter()
        timestamp = datetime(2024, 1, 2)
        for index in range(10):
            close_price = td.INPUT_CLOSE[index]
            bar = Bar(
                time=timestamp,
                open=td.INPUT_OPEN[index],
                high=td.INPUT_HIGH[index],
                low=td.INPUT_LOW[index],
                close=close_price,
                volume=1000.0,
            )
            bar_out = bar_indicator.update_bar(bar)
            scalar_out = scalar_indicator.update_scalar(Scalar(time=timestamp, value=close_price))
            self.assertEqual(len(bar_out), 1)
            self.assertEqual(bar_out[0].time, timestamp)
            if math.isnan(scalar_out[0].value):
                self.assertTrue(math.isnan(bar_out[0].value))
            else:
                self.assertAlmostEqual(bar_out[0].value, scalar_out[0].value, delta=1e-15)

        quote_indicator = RoofingFilter()
        for _ in range(5):
            quote_out = quote_indicator.update_quote(
                Quote(time=timestamp, bid_price=99.0, ask_price=101.0, bid_size=1.0, ask_size=1.0)
            )
        self.assertEqual(len(quote_out), 1)
        self.assertFalse(math.isnan(quote_out[0].value))

        trade_indicator = RoofingFilter()
        for _ in range(5):
            trade_out = trade_indicator.update_trade(
                Trade(time=timestamp, price=100.0, volume=50.0)
            )
        self.assertEqual(len(trade_out), 1)
        self.assertFalse(math.isnan(trade_out[0].value))

        created_default = create_indicator(Identifier.ROOFING_FILTER)
        self.assertEqual(created_default.metadata().mnemonic, "roof2hp(10, 48)")

        created_custom = create_indicator(
            Identifier.ROOFING_FILTER,
            {
                "shortest_cycle_period": 10.5,
                "longest_cycle_period": 48.0,
                "highpass_poles": 1,
                "zero_mean": True,
            },
        )
        self.assertEqual(created_custom.metadata().mnemonic, "roof1hpzm(10.5, 48)")


if __name__ == "__main__":
    unittest.main()
