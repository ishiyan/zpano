"""Tests for Ehlers' Center of Gravity Oscillator."""

from __future__ import annotations

from datetime import datetime, timedelta
from fractions import Fraction
import math
import unittest

from ....entities.bar import Bar
from ....entities.bar_component import BarComponent
from ....entities.quote import Quote
from ....entities.quote_component import QuoteComponent
from ....entities.scalar import Scalar
from ....entities.trade import Trade
from ....entities.trade_component import TradeComponent
from ...core.identifier import Identifier
from ...core.outputs.shape import Shape
from ...factory.factory import create_indicator
from . import (
    CenterOfGravityOscillator,
    CenterOfGravityOscillatorOutput,
    CenterOfGravityOscillatorParams,
    center_of_gravity_oscillator_mnemonic,
    center_of_gravity_oscillator_trigger_mnemonic,
    default_params,
)
from . import test_testdata as testdata

REFERENCE_CASES: tuple[tuple[int, bool, list[float], list[float]], ...] = (
    (1, False, testdata.EXPECTED_L1_C0_CG, testdata.EXPECTED_L1_C0_TR),
    (1, True, testdata.EXPECTED_L1_C1_CG, testdata.EXPECTED_L1_C1_TR),
    (2, False, testdata.EXPECTED_L2_C0_CG, testdata.EXPECTED_L2_C0_TR),
    (2, True, testdata.EXPECTED_L2_C1_CG, testdata.EXPECTED_L2_C1_TR),
    (3, False, testdata.EXPECTED_L3_C0_CG, testdata.EXPECTED_L3_C0_TR),
    (3, True, testdata.EXPECTED_L3_C1_CG, testdata.EXPECTED_L3_C1_TR),
    (5, False, testdata.EXPECTED_L5_C0_CG, testdata.EXPECTED_L5_C0_TR),
    (5, True, testdata.EXPECTED_L5_C1_CG, testdata.EXPECTED_L5_C1_TR),
    (8, False, testdata.EXPECTED_L8_C0_CG, testdata.EXPECTED_L8_C0_TR),
    (8, True, testdata.EXPECTED_L8_C1_CG, testdata.EXPECTED_L8_C1_TR),
    (10, False, testdata.EXPECTED_L10_C0_CG, testdata.EXPECTED_L10_C0_TR),
    (10, True, testdata.EXPECTED_L10_C1_CG, testdata.EXPECTED_L10_C1_TR),
    (14, False, testdata.EXPECTED_L14_C0_CG, testdata.EXPECTED_L14_C0_TR),
    (14, True, testdata.EXPECTED_L14_C1_CG, testdata.EXPECTED_L14_C1_TR),
    (20, False, testdata.EXPECTED_L20_C0_CG, testdata.EXPECTED_L20_C0_TR),
    (20, True, testdata.EXPECTED_L20_C1_CG, testdata.EXPECTED_L20_C1_TR),
    (32, False, testdata.EXPECTED_L32_C0_CG, testdata.EXPECTED_L32_C0_TR),
    (32, True, testdata.EXPECTED_L32_C1_CG, testdata.EXPECTED_L32_C1_TR),
    (64, False, testdata.EXPECTED_L64_C0_CG, testdata.EXPECTED_L64_C0_TR),
    (64, True, testdata.EXPECTED_L64_C1_CG, testdata.EXPECTED_L64_C1_TR),
    (126, False, testdata.EXPECTED_L126_C0_CG, testdata.EXPECTED_L126_C0_TR),
    (126, True, testdata.EXPECTED_L126_C1_CG, testdata.EXPECTED_L126_C1_TR),
    (252, False, testdata.EXPECTED_L252_C0_CG, testdata.EXPECTED_L252_C0_TR),
    (252, True, testdata.EXPECTED_L252_C1_CG, testdata.EXPECTED_L252_C1_TR),
)


def _source_outputs(
    prices: list[float], length: int, centered: bool
) -> list[tuple[float, float]]:
    previous = 0.0
    results: list[tuple[float, float]] = []
    for step in range(len(prices)):
        if step < length - 1:
            results.append((math.nan, math.nan))
            continue
        trigger = previous if step >= length else math.nan
        numerator = 0.0
        denominator = 0.0
        for lag in range(length):
            numerator = numerator + (1 + lag) * prices[step - lag]
            denominator = denominator + prices[step - lag]
        if denominator != 0.0:
            previous = -numerator / denominator
            if centered:
                previous = previous + (length + 1) / 2.0
        results.append((previous, trigger))
    return results


class TestCenterOfGravityOscillator(unittest.TestCase):
    def _assert_number(
        self, actual: float, expected: float, *, delta: float = 1e-13
    ) -> None:
        if math.isnan(expected):
            self.assertTrue(math.isnan(actual), f"expected NaN, got {actual}")
        else:
            self.assertFalse(math.isnan(actual), f"expected {expected}, got NaN")
            self.assertAlmostEqual(actual, expected, delta=delta)

    def test_all_twenty_four_reference_cases_and_published_loop(self) -> None:
        self.assertEqual(len(testdata.INPUT_CLOSE), 252)
        base_time = datetime(2020, 1, 1)

        for length, centered, expected_values, expected_triggers in REFERENCE_CASES:
            params = CenterOfGravityOscillatorParams(length=length, centered=centered)
            with self.subTest(mnemonic=center_of_gravity_oscillator_mnemonic(params)):
                indicator = CenterOfGravityOscillator(params)
                bar_indicator = CenterOfGravityOscillator(params)
                oracle = _source_outputs(testdata.INPUT_CLOSE, length, centered)

                self.assertEqual(indicator.length(), length)
                self.assertEqual(indicator.centered(), centered)
                self.assertFalse(indicator.is_primed())
                self.assertFalse(indicator.is_trigger_primed())
                self.assertTrue(math.isnan(indicator.value()))
                self.assertTrue(math.isnan(indicator.trigger()))
                self.assertEqual(len(expected_values), 252)
                self.assertEqual(len(expected_triggers), 252)

                for index, price in enumerate(testdata.INPUT_CLOSE):
                    actual_value = indicator.update(price)
                    actual_trigger = indicator.trigger()

                    self.assertEqual(indicator.is_primed(), index >= length - 1)
                    self.assertEqual(indicator.is_trigger_primed(), index >= length)
                    self._assert_number(actual_value, expected_values[index])
                    self._assert_number(actual_trigger, expected_triggers[index])
                    self._assert_number(indicator.value(), actual_value, delta=0.0)

                    oracle_value, oracle_trigger = oracle[index]
                    self._assert_number(actual_value, oracle_value, delta=0.0)
                    self._assert_number(actual_trigger, oracle_trigger, delta=0.0)

                    stamp = base_time + timedelta(days=index)
                    bar_out = bar_indicator.update_bar(
                        Bar(
                            time=stamp,
                            open=testdata.INPUT_OPEN[index],
                            high=testdata.INPUT_HIGH[index],
                            low=testdata.INPUT_LOW[index],
                            close=price,
                            volume=1000.0,
                        )
                    )
                    self.assertEqual(len(bar_out), 2)
                    self.assertEqual(bar_out[0].time, stamp)
                    self.assertEqual(bar_out[1].time, stamp)
                    self._assert_number(bar_out[0].value, actual_value, delta=0.0)
                    self._assert_number(bar_out[1].value, actual_trigger, delta=0.0)

    def test_exact_decimal_oracle_and_per_output_availability(self) -> None:
        decimal_prices = [Fraction(str(price)) for price in testdata.INPUT_CLOSE]
        for length, centered, _, _ in REFERENCE_CASES:
            with self.subTest(length=length, centered=centered):
                indicator = CenterOfGravityOscillator(
                    CenterOfGravityOscillatorParams(length=length, centered=centered)
                )
                prior = math.nan
                for index, price in enumerate(testdata.INPUT_CLOSE):
                    actual_value = indicator.update(price)
                    actual_trigger = indicator.trigger()
                    if index < length - 1:
                        self.assertTrue(math.isnan(actual_value))
                        self.assertTrue(math.isnan(actual_trigger))
                        self.assertFalse(indicator.is_primed())
                        continue
                    chronological = decimal_prices[index - length + 1 : index + 1]
                    numerator = sum(
                        (length - age) * item for age, item in enumerate(chronological)
                    )
                    exact = -numerator / sum(chronological)
                    if centered:
                        exact += Fraction(length + 1, 2)
                    self._assert_number(actual_value, float(exact), delta=1e-12)
                    self._assert_number(actual_trigger, prior, delta=1e-12)
                    prior = float(exact)
                    self.assertTrue(indicator.is_primed())
                    self.assertEqual(indicator.is_trigger_primed(), index >= length)

    def test_newest_price_has_weight_one_and_sign_is_negative(self) -> None:
        article = CenterOfGravityOscillator(CenterOfGravityOscillatorParams(length=3))
        book = CenterOfGravityOscillator(
            CenterOfGravityOscillatorParams(length=3, centered=True)
        )
        article_value = math.nan
        book_value = math.nan
        for price in (10.0, 20.0, 30.0):
            article_value = article.update(price)
            book_value = book.update(price)
        self._assert_number(article_value, -5.0 / 3.0, delta=1e-15)
        self._assert_number(book_value, 1.0 / 3.0, delta=1e-15)
        self.assertTrue(math.isnan(article.trigger()))
        self._assert_number(article.update(60.0), -18.0 / 11.0, delta=1e-15)
        self._assert_number(article.trigger(), -5.0 / 3.0, delta=1e-15)

    def test_constant_price_length_one_scale_invariance_and_crossover(self) -> None:
        for length in (1, 2, 3, 10):
            for centered in (False, True):
                indicator = CenterOfGravityOscillator(
                    CenterOfGravityOscillatorParams(length=length, centered=centered)
                )
                baseline = 0.0 if centered else -(length + 1) / 2.0
                for step in range(length + 2):
                    actual_value = indicator.update(8.0)
                    self._assert_number(
                        actual_value,
                        math.nan if step < length - 1 else baseline,
                    )
                    self._assert_number(
                        indicator.trigger(),
                        math.nan if step < length else baseline,
                    )

        base = CenterOfGravityOscillator(
            CenterOfGravityOscillatorParams(length=5, centered=True)
        )
        scaled = CenterOfGravityOscillator(
            CenterOfGravityOscillatorParams(length=5, centered=True)
        )
        for price in testdata.INPUT_CLOSE[:25]:
            self._assert_number(base.update(price), scaled.update(3.75 * price))
            self._assert_number(base.trigger(), scaled.trigger())

        rising = CenterOfGravityOscillator(
            CenterOfGravityOscillatorParams(length=4, centered=True)
        )
        falling = CenterOfGravityOscillator(
            CenterOfGravityOscillatorParams(length=4, centered=True)
        )
        for step in range(4):
            rising.update(float(10 + step))
            falling.update(float(13 - step))
        self.assertGreater(rising.value(), 0.0)
        self.assertLess(falling.value(), 0.0)
        after_drop = rising.update(1.0)
        self.assertLess(after_drop, rising.trigger())

    def test_zero_denominator_and_subnormal_guards(self) -> None:
        for centered in (False, True):
            initial = CenterOfGravityOscillator(
                CenterOfGravityOscillatorParams(length=2, centered=centered)
            )
            initial.update(1.0)
            self.assertEqual(initial.update(-1.0), 0.0)
            self.assertTrue(math.isnan(initial.trigger()))

            subsequent = CenterOfGravityOscillator(
                CenterOfGravityOscillatorParams(length=2, centered=centered)
            )
            subsequent.update(1.0)
            previous = subsequent.update(2.0)
            held = subsequent.update(-2.0)
            self.assertEqual(held, previous)
            self.assertEqual(subsequent.trigger(), previous)
            resumed = subsequent.update(4.0)
            self.assertEqual(subsequent.trigger(), previous)
            self.assertEqual(resumed, 1.5 if centered else 0.0)

        subnormal = CenterOfGravityOscillator(CenterOfGravityOscillatorParams(length=2))
        subnormal.update(1e-309)
        self.assertAlmostEqual(subnormal.update(3e-309), -1.25, delta=1e-14)

    def test_nan_skip_infinity_rejection_and_reset(self) -> None:
        skipped = CenterOfGravityOscillator(CenterOfGravityOscillatorParams(length=3))
        uninterrupted = CenterOfGravityOscillator(
            CenterOfGravityOscillatorParams(length=3)
        )
        stamp = datetime(2026, 1, 1)
        for price in (1.0, 2.0, 3.0, 4.0, 5.0):
            self.assertTrue(math.isnan(skipped.update(math.nan)))
            gap = skipped.update_scalar(Scalar(time=stamp, value=math.nan))
            self.assertTrue(math.isnan(gap[0].value) and math.isnan(gap[1].value))
            actual = skipped.update(price)
            reference = uninterrupted.update(price)
            self._assert_number(actual, reference, delta=0.0)
            self._assert_number(skipped.trigger(), uninterrupted.trigger(), delta=0.0)
        self.assertTrue(skipped.is_trigger_primed())

        single = CenterOfGravityOscillator(CenterOfGravityOscillatorParams(length=1))
        single.update(4.0)
        for bad_sample in (math.inf, -math.inf, True, "bad"):
            with self.assertRaises(ValueError):
                single.update(bad_sample)  # type: ignore[arg-type]
        self.assertEqual(single.update(5.0), -1.0)
        self.assertEqual(single.trigger(), -1.0)

        resettable = CenterOfGravityOscillator(
            CenterOfGravityOscillatorParams(length=2, centered=True)
        )
        for price in (1.0, 2.0, 3.0):
            resettable.update(price)
        resettable.reset()
        self.assertFalse(resettable.is_primed())
        self.assertFalse(resettable.is_trigger_primed())
        self.assertTrue(math.isnan(resettable.value()))
        self.assertTrue(math.isnan(resettable.trigger()))
        self.assertTrue(math.isnan(resettable.update(1.0)))
        self.assertEqual(resettable.update(-1.0), 0.0)

    def test_validation_metadata_entities_and_factory(self) -> None:
        for bad_length in (0, -1, 1.5, 2.0, True, "10"):
            with self.subTest(length=bad_length), self.assertRaises(ValueError):
                CenterOfGravityOscillator(
                    CenterOfGravityOscillatorParams(length=bad_length)  # type: ignore[arg-type]
                )
        for bad_centered in (0, 1, None, "book"):
            with self.subTest(centered=bad_centered), self.assertRaises(ValueError):
                CenterOfGravityOscillator(
                    CenterOfGravityOscillatorParams(centered=bad_centered)  # type: ignore[arg-type]
                )
        with self.assertRaises(ValueError):
            CenterOfGravityOscillator(
                CenterOfGravityOscillatorParams(bar_component=9999)  # type: ignore[arg-type]
            )

        default_ind = CenterOfGravityOscillator.create_default()
        meta = default_ind.metadata()
        self.assertEqual(meta.identifier, Identifier.CENTER_OF_GRAVITY_OSCILLATOR)
        self.assertEqual(meta.mnemonic, "cog(10)")
        self.assertEqual(meta.description, "Center of Gravity oscillator cog(10)")
        self.assertEqual(len(meta.outputs), 2)
        self.assertEqual(meta.outputs[0].kind, CenterOfGravityOscillatorOutput.VALUE)
        self.assertEqual(meta.outputs[0].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[0].mnemonic, "cog(10)")
        self.assertEqual(meta.outputs[1].kind, CenterOfGravityOscillatorOutput.TRIGGER)
        self.assertEqual(meta.outputs[1].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[1].mnemonic, "cogTrig(10)")
        self.assertEqual(
            meta.outputs[1].description, "Center of Gravity trigger cogTrig(10)"
        )

        custom_params = CenterOfGravityOscillatorParams(
            length=14,
            centered=True,
            bar_component=BarComponent.MEDIAN,
            quote_component=QuoteComponent.BID,
            trade_component=TradeComponent.VOLUME,
        )
        self.assertEqual(
            center_of_gravity_oscillator_mnemonic(custom_params),
            "cog(14, centered, hl/2, b, v)",
        )
        self.assertEqual(
            center_of_gravity_oscillator_trigger_mnemonic(custom_params),
            "cogTrig(14, centered, hl/2, b, v)",
        )

        from_factory = create_indicator(
            Identifier.CENTER_OF_GRAVITY_OSCILLATOR,
            {"length": 8, "centered": True, "barComponent": BarComponent.MEDIAN},
        )
        self.assertEqual(from_factory.metadata().mnemonic, "cog(8, centered, hl/2)")
        with self.assertRaises(ValueError):
            create_indicator(
                Identifier.CENTER_OF_GRAVITY_OSCILLATOR, {"length": 1.5}
            )
        with self.assertRaises(ValueError):
            create_indicator(
                Identifier.CENTER_OF_GRAVITY_OSCILLATOR, {"centered": 1}
            )

        base_time = datetime(2026, 1, 1)
        bar_median = CenterOfGravityOscillator(
            CenterOfGravityOscillatorParams(
                length=4, centered=True, bar_component=BarComponent.MEDIAN
            )
        )
        quote_mid = CenterOfGravityOscillator(
            CenterOfGravityOscillatorParams(length=4, centered=True)
        )
        trade_price = CenterOfGravityOscillator(
            CenterOfGravityOscillatorParams(length=4, centered=True)
        )
        scalar_direct = CenterOfGravityOscillator(
            CenterOfGravityOscillatorParams(length=4, centered=True)
        )
        for step in range(12):
            stamp = base_time + timedelta(days=step)
            median = (testdata.INPUT_HIGH[step] + testdata.INPUT_LOW[step]) / 2.0
            bar_out = bar_median.update_bar(
                Bar(
                    time=stamp,
                    open=testdata.INPUT_OPEN[step],
                    high=testdata.INPUT_HIGH[step],
                    low=testdata.INPUT_LOW[step],
                    close=testdata.INPUT_CLOSE[step],
                    volume=1000.0,
                )
            )
            quote_out = quote_mid.update_quote(
                Quote(
                    time=stamp,
                    bid_price=testdata.INPUT_LOW[step],
                    ask_price=testdata.INPUT_HIGH[step],
                    bid_size=1.0,
                    ask_size=1.0,
                )
            )
            trade_out = trade_price.update_trade(
                Trade(time=stamp, price=median, volume=10.0)
            )
            scalar_out = scalar_direct.update_scalar(Scalar(time=stamp, value=median))
            for slot in range(2):
                self._assert_number(
                    bar_out[slot].value, scalar_out[slot].value, delta=1e-15
                )
                self._assert_number(
                    quote_out[slot].value, scalar_out[slot].value, delta=1e-15
                )
                self._assert_number(
                    trade_out[slot].value, scalar_out[slot].value, delta=1e-15
                )


if __name__ == "__main__":
    unittest.main()
