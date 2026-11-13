"""Tests for Ehlers' SuperSmoother indicator."""

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
from py.indicators.john_ehlers.super_smoother import (
    SuperSmoother,
    SuperSmootherOutput,
    SuperSmootherParams,
    default_params,
    super_smoother_mnemonic,
)
from py.indicators.john_ehlers.super_smoother import test_testdata as testdata


REFERENCE_CASES = (
    (SuperSmootherParams(2), testdata.EXPECTED_P2, 2),
    (SuperSmootherParams(3), testdata.EXPECTED_P3, 2),
    (SuperSmootherParams(5), testdata.EXPECTED_P5, 2),
    (SuperSmootherParams(8), testdata.EXPECTED_P8, 2),
    (SuperSmootherParams(10), testdata.EXPECTED_P10, 2),
    (SuperSmootherParams(10.5), testdata.EXPECTED_P10_5, 2),
    (SuperSmootherParams(20), testdata.EXPECTED_P20, 2),
    (SuperSmootherParams(40), testdata.EXPECTED_P40, 2),
    (SuperSmootherParams(80), testdata.EXPECTED_P80, 2),
    (SuperSmootherParams(252), testdata.EXPECTED_P252, 2),
)


class TestSuperSmootherFixtures(unittest.TestCase):
    def test_all_ten_reference_cases(self) -> None:
        self.assertEqual(len(testdata.INPUT_CLOSE), 252)
        for params, expected, expected_warmup in REFERENCE_CASES:
            with self.subTest(mnemonic=super_smoother_mnemonic(params)):
                indicator = SuperSmoother(params)
                self.assertEqual(indicator.warmup_bars(), expected_warmup)
                self.assertFalse(indicator.is_primed())
                self.assertTrue(math.isnan(indicator.value()))
                self.assertEqual(len(expected), 252)

                for index, sample in enumerate(testdata.INPUT_CLOSE):
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
            with self.subTest(mnemonic=super_smoother_mnemonic(params)):
                first = SuperSmoother(params)
                second = SuperSmoother(params)
                for sample in testdata.INPUT_CLOSE:
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
                third = SuperSmoother(params)
                for index, sample in enumerate(testdata.INPUT_CLOSE):
                    left = first.update(sample)
                    right = third.update(sample)
                    if math.isnan(right):
                        self.assertTrue(math.isnan(left), f"bar {index} expected NaN after reset")
                    else:
                        self.assertAlmostEqual(left, right, delta=1e-15)


class TestSuperSmootherInvariants(unittest.TestCase):
    def test_constant_input_and_priming(self) -> None:
        indicator = SuperSmoother()
        outputs = [indicator.update(7.0) for _ in range(500)]
        first_finite = next(index for index, item in enumerate(outputs) if math.isfinite(item))
        self.assertEqual(first_finite, indicator.warmup_bars())
        self.assertAlmostEqual(outputs[-1], 7.0, delta=1e-12)

    def test_nyquist_alternating_zero(self) -> None:
        indicator = SuperSmoother()
        outputs = [indicator.update(1.0 if index % 2 == 0 else -1.0) for index in range(500)]
        self.assertAlmostEqual(outputs[-1], 0.0, delta=1e-12)


class TestSuperSmootherValidationAndMetadata(unittest.TestCase):
    def test_invalid_configuration_and_infinite_sample(self) -> None:
        invalid_cases = (
            SuperSmootherParams(shortest_cycle_period=1.0),
            SuperSmootherParams(shortest_cycle_period=1.999),
            SuperSmootherParams(shortest_cycle_period=0.0),
            SuperSmootherParams(shortest_cycle_period=-2.0),
            SuperSmootherParams(shortest_cycle_period=math.nan),
            SuperSmootherParams(shortest_cycle_period=math.inf),
            SuperSmootherParams(shortest_cycle_period=-math.inf),
            SuperSmootherParams(shortest_cycle_period=True),  # type: ignore[arg-type]
        )
        for invalid_params in invalid_cases:
            with self.assertRaises(ValueError):
                SuperSmoother(invalid_params)

        indicator = SuperSmoother()
        for bad_sample in (math.inf, -math.inf):
            with self.assertRaises(ValueError):
                indicator.update(bad_sample)
        self.assertFalse(indicator.is_primed())

    def test_metadata_and_mnemonics(self) -> None:
        indicator = SuperSmoother(default_params())
        meta = indicator.metadata()
        self.assertEqual(meta.identifier, Identifier.SUPER_SMOOTHER)
        self.assertEqual(meta.mnemonic, "ss(10)")
        self.assertEqual(meta.description, "Super Smoother ss(10)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].kind, int(SuperSmootherOutput.VALUE))
        self.assertEqual(meta.outputs[0].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[0].mnemonic, "ss(10)")
        self.assertEqual(meta.outputs[0].description, "Super Smoother ss(10)")

        mnemonic_cases = (
            (SuperSmootherParams(), "ss(10)"),
            (SuperSmootherParams(10.5), "ss(10.5)"),
            (SuperSmootherParams(bar_component=BarComponent.MEDIAN), "ss(10, hl/2)"),
            (SuperSmootherParams(quote_component=QuoteComponent.BID), "ss(10, b)"),
            (SuperSmootherParams(trade_component=TradeComponent.VOLUME), "ss(10, v)"),
            (
                SuperSmootherParams(
                    bar_component=BarComponent.OPEN,
                    quote_component=QuoteComponent.ASK,
                ),
                "ss(10, o, a)",
            ),
            (
                SuperSmootherParams(
                    bar_component=BarComponent.TYPICAL,
                    trade_component=TradeComponent.VOLUME,
                ),
                "ss(10, hlc/3, v)",
            ),
            (
                SuperSmootherParams(
                    quote_component=QuoteComponent.WEIGHTED,
                    trade_component=TradeComponent.VOLUME,
                ),
                "ss(10, (bbs+aas)/(bs+as), v)",
            ),
        )
        for params, expected_mnemonic in mnemonic_cases:
            self.assertEqual(super_smoother_mnemonic(params), expected_mnemonic)
            self.assertEqual(SuperSmoother(params).metadata().mnemonic, expected_mnemonic)

    def test_entity_updates_and_factory(self) -> None:
        bar_indicator = SuperSmoother.create_default()
        scalar_indicator = SuperSmoother()
        timestamp = datetime(2024, 1, 2)

        for index in range(10):
            close_price = testdata.INPUT_CLOSE[index]
            bar = Bar(
                time=timestamp,
                open=testdata.INPUT_OPEN[index],
                high=testdata.INPUT_HIGH[index],
                low=testdata.INPUT_LOW[index],
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

        quote_indicator = SuperSmoother()
        quote_out = []
        for _ in range(5):
            quote_out = quote_indicator.update_quote(
                Quote(time=timestamp, bid_price=99.0, ask_price=101.0, bid_size=1.0, ask_size=1.0)
            )
        self.assertEqual(len(quote_out), 1)
        self.assertFalse(math.isnan(quote_out[0].value))

        trade_indicator = SuperSmoother()
        trade_out = []
        for _ in range(5):
            trade_out = trade_indicator.update_trade(
                Trade(time=timestamp, price=100.0, volume=50.0)
            )
        self.assertEqual(len(trade_out), 1)
        self.assertFalse(math.isnan(trade_out[0].value))

        created_default = create_indicator(Identifier.SUPER_SMOOTHER)
        self.assertEqual(created_default.metadata().mnemonic, "ss(10)")

        created_custom = create_indicator(
            Identifier.SUPER_SMOOTHER,
            {"shortest_cycle_period": 10.5},
        )
        self.assertEqual(created_custom.metadata().mnemonic, "ss(10.5)")


if __name__ == "__main__":
    unittest.main()
