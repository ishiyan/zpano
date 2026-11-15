"""Tests for Ehlers' Fractal Adaptive Moving Average (FRAMA)."""

from __future__ import annotations

from datetime import datetime, timedelta
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
    FractalAdaptiveMovingAverage,
    FractalAdaptiveMovingAverageOutput,
    FractalAdaptiveMovingAverageParams,
    default_params,
    fractal_adaptive_moving_average_dimension_mnemonic,
    fractal_adaptive_moving_average_mnemonic,
    test_testdata as testdata,
)

REFERENCE_CASES = (
    (
        FractalAdaptiveMovingAverageParams(16, 0.01, 1),
        testdata.EXPECTED_L16_A0_01_C1_VALUE,
        testdata.EXPECTED_L16_A0_01_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(2, 0.01, 1),
        testdata.EXPECTED_L2_A0_01_C1_VALUE,
        testdata.EXPECTED_L2_A0_01_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(4, 0.01, 1),
        testdata.EXPECTED_L4_A0_01_C1_VALUE,
        testdata.EXPECTED_L4_A0_01_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(8, 0.01, 1),
        testdata.EXPECTED_L8_A0_01_C1_VALUE,
        testdata.EXPECTED_L8_A0_01_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(20, 0.01, 1),
        testdata.EXPECTED_L20_A0_01_C1_VALUE,
        testdata.EXPECTED_L20_A0_01_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(30, 0.01, 1),
        testdata.EXPECTED_L30_A0_01_C1_VALUE,
        testdata.EXPECTED_L30_A0_01_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(64, 0.01, 1),
        testdata.EXPECTED_L64_A0_01_C1_VALUE,
        testdata.EXPECTED_L64_A0_01_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(128, 0.01, 1),
        testdata.EXPECTED_L128_A0_01_C1_VALUE,
        testdata.EXPECTED_L128_A0_01_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(250, 0.01, 1),
        testdata.EXPECTED_L250_A0_01_C1_VALUE,
        testdata.EXPECTED_L250_A0_01_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(252, 0.01, 1),
        testdata.EXPECTED_L252_A0_01_C1_VALUE,
        testdata.EXPECTED_L252_A0_01_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(254, 0.01, 1),
        testdata.EXPECTED_L254_A0_01_C1_VALUE,
        testdata.EXPECTED_L254_A0_01_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(16, 0.001, 1),
        testdata.EXPECTED_L16_A0_001_C1_VALUE,
        testdata.EXPECTED_L16_A0_001_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(16, 0.1, 1),
        testdata.EXPECTED_L16_A0_1_C1_VALUE,
        testdata.EXPECTED_L16_A0_1_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(16, 1.0, 1),
        testdata.EXPECTED_L16_A1_C1_VALUE,
        testdata.EXPECTED_L16_A1_C1_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(16, 0.01, 0),
        testdata.EXPECTED_L16_A0_01_C0_VALUE,
        testdata.EXPECTED_L16_A0_01_C0_FDIM,
    ),
    (
        FractalAdaptiveMovingAverageParams(30, 0.01, 0),
        testdata.EXPECTED_L30_A0_01_C0_VALUE,
        testdata.EXPECTED_L30_A0_01_C0_FDIM,
    ),
)


def _article_batch(
    prices: list[float],
    length: int,
    slowest: float = 0.01,
    mode: int = 0,
    highs: list[float] | None = None,
    lows: list[float] | None = None,
) -> list[tuple[float, float]]:
    high_series = prices if highs is None else highs
    low_series = prices if lows is None else lows
    half = length // 2
    scale = math.log(slowest) if mode == 1 else -4.6
    dimension = 0.0
    filtered = [0.0] * len(prices)
    outputs: list[tuple[float, float]] = []
    for index, price in enumerate(prices):
        if index + 1 >= length:
            recent_highs = [high_series[index - offset] for offset in range(half)]
            recent_lows = [low_series[index - offset] for offset in range(half)]
            older_highs = [high_series[index - offset] for offset in range(half, length)]
            older_lows = [low_series[index - offset] for offset in range(half, length)]
            all_highs = [high_series[index - offset] for offset in range(length)]
            all_lows = [low_series[index - offset] for offset in range(length)]
            n1 = (max(recent_highs) - min(recent_lows)) / half
            n2 = (max(older_highs) - min(older_lows)) / half
            n3 = (max(all_highs) - min(all_lows)) / length
            if n1 > 0.0 and n2 > 0.0 and n3 > 0.0:
                dimension = (math.log(n1 + n2) - math.log(n3)) / math.log(2.0)
            raw_alpha = 1.0 if dimension <= 1.0 else math.exp(scale * (dimension - 1.0))
            alpha = max(slowest, min(1.0, raw_alpha))
            filtered[index] = price if index + 1 < length + 1 else alpha * price + (1.0 - alpha) * filtered[index - 1]
            outputs.append((filtered[index], dimension))
        else:
            filtered[index] = price
            outputs.append((math.nan, math.nan))
    return outputs


class TestFractalAdaptiveMovingAverageFixtures(unittest.TestCase):
    def test_all_sixteen_reference_cases(self) -> None:
        self.assertEqual(len(testdata.INPUT_CLOSE), 252)
        for params, expected_values, expected_dimensions in REFERENCE_CASES:
            with self.subTest(mnemonic=fractal_adaptive_moving_average_mnemonic(params)):
                indicator = FractalAdaptiveMovingAverage(params)
                warmup = params.length - 1
                self.assertEqual(indicator.warmup_bars(), warmup)
                self.assertEqual(indicator.samples_seen(), 0)
                self.assertFalse(indicator.is_primed())
                self.assertTrue(math.isnan(indicator.value()))
                self.assertTrue(math.isnan(indicator.fractal_dimension()))
                self.assertTrue(math.isnan(indicator.alpha()))
                self.assertEqual(len(expected_values), 252)
                self.assertEqual(len(expected_dimensions), 252)

                for index, sample in enumerate(testdata.INPUT_CLOSE):
                    actual_value = indicator.update(sample)
                    actual_dimension = indicator.fractal_dimension()
                    target_value = expected_values[index]
                    target_dimension = expected_dimensions[index]

                    if index < warmup:
                        self.assertFalse(indicator.is_primed(), f"bar {index} unexpectedly primed")
                        self.assertTrue(math.isnan(actual_value), f"bar {index} expected NaN value")
                        self.assertTrue(math.isnan(actual_dimension), f"bar {index} expected NaN dimension")
                        self.assertTrue(math.isnan(target_value))
                        self.assertTrue(math.isnan(target_dimension))
                    else:
                        self.assertTrue(indicator.is_primed(), f"bar {index} expected primed")
                        value_bound = max(1e-12, abs(target_value) * 1e-12)
                        dim_bound = max(1e-12, abs(target_dimension) * 1e-12)
                        self.assertAlmostEqual(
                            actual_value, target_value, delta=value_bound, msg=f"bar {index} value"
                        )
                        self.assertAlmostEqual(
                            actual_dimension,
                            target_dimension,
                            delta=dim_bound,
                            msg=f"bar {index} dimension",
                        )
                        self.assertAlmostEqual(indicator.value(), actual_value, delta=1e-15)

    def test_streaming_matches_independent_batch_oracle(self) -> None:
        for params, _, _ in REFERENCE_CASES:
            stream = FractalAdaptiveMovingAverage(params)
            expected = _article_batch(
                testdata.INPUT_CLOSE,
                params.length,
                params.slowest_smoothing_factor,
                params.coefficient_mode,
            )
            for index, sample in enumerate(testdata.INPUT_CLOSE):
                actual_value = stream.update(sample)
                actual_dimension = stream.fractal_dimension()
                target_value, target_dimension = expected[index]
                if math.isnan(target_value):
                    self.assertTrue(math.isnan(actual_value))
                    self.assertTrue(math.isnan(actual_dimension))
                else:
                    self.assertAlmostEqual(actual_value, target_value, delta=1e-12)
                    self.assertAlmostEqual(actual_dimension, target_dimension, delta=1e-12)

    def test_midpoint_and_bar_ranges_match_article_batch(self) -> None:
        midpoints = [
            (high + low) / 2.0
            for high, low in zip(testdata.INPUT_HIGH, testdata.INPUT_LOW)
        ]
        stream = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(16, 0.01, 0)
        )
        expected = _article_batch(
            midpoints,
            16,
            0.01,
            0,
            highs=testdata.INPUT_HIGH,
            lows=testdata.INPUT_LOW,
        )
        for index, (midpoint, high, low) in enumerate(
            zip(midpoints, testdata.INPUT_HIGH, testdata.INPUT_LOW)
        ):
            actual_value = stream.update(midpoint, high, low)
            actual_dimension = stream.fractal_dimension()
            target_value, target_dimension = expected[index]
            if math.isnan(target_value):
                self.assertTrue(math.isnan(actual_value))
                self.assertTrue(math.isnan(actual_dimension))
            else:
                self.assertAlmostEqual(actual_value, target_value, delta=1e-12)
                self.assertAlmostEqual(actual_dimension, target_dimension, delta=1e-12)


class TestFractalAdaptiveMovingAverageInvariants(unittest.TestCase):
    def test_priming_final_price_seed_and_first_adaptive_update(self) -> None:
        stream = FractalAdaptiveMovingAverage(FractalAdaptiveMovingAverageParams(length=16))
        previous_seed = 0.0
        for index, price in enumerate(testdata.INPUT_CLOSE):
            actual_value = stream.update(price)
            self.assertEqual(stream.samples_seen(), index + 1)
            self.assertEqual(stream.is_primed(), index >= 15)
            self.assertEqual(stream.warmup_bars(), 15)
            self.assertEqual(math.isnan(actual_value), index < 15)
            self.assertEqual(math.isnan(stream.fractal_dimension()), index < 15)
            if index == 15:
                self.assertEqual(actual_value, price)
                previous_seed = actual_value
            if index == 16:
                current_alpha = stream.alpha()
                self.assertAlmostEqual(
                    actual_value,
                    current_alpha * price + (1.0 - current_alpha) * previous_seed,
                    delta=1e-15,
                )

    def test_long_lookback_priming_boundaries(self) -> None:
        for length, finite_count in ((128, 125), (250, 3), (252, 1), (254, 0)):
            stream = FractalAdaptiveMovingAverage(
                FractalAdaptiveMovingAverageParams(length=length)
            )
            values = [stream.update(price) for price in testdata.INPUT_CLOSE]
            self.assertEqual(sum(math.isfinite(item) for item in values), finite_count)
            if length == 252:
                self.assertEqual(values[-1], testdata.INPUT_CLOSE[-1])

    def test_raw_dimension_of_ramp_is_below_one(self) -> None:
        expected_dimension = math.log2(28.0 / 15.0)
        for offset in (100.0, 0.0, -100.0):
            stream = FractalAdaptiveMovingAverage(
                FractalAdaptiveMovingAverageParams(length=16)
            )
            for step in range(40):
                actual_value = stream.update(offset + step)
                if step >= 15:
                    self.assertAlmostEqual(
                        stream.fractal_dimension(), expected_dimension, places=13
                    )
                    self.assertLess(stream.fractal_dimension(), 1.0)
                    self.assertEqual(stream.alpha(), 1.0)
                    self.assertEqual(actual_value, offset + step)

    def test_gap_dimension_is_not_clamped_or_exponentiated_unsafely(self) -> None:
        stream = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(length=4)
        )
        actual_value = math.nan
        for price in (0.0, 1.0, 100.0, 101.0):
            actual_value = stream.update(price)
        self.assertAlmostEqual(
            stream.fractal_dimension(), math.log2(4.0 / 101.0), places=13
        )
        self.assertLess(stream.fractal_dimension(), 0.0)
        self.assertEqual(stream.alpha(), 1.0)
        self.assertEqual(actual_value, 101.0)

    def test_constant_and_minimum_length_guard(self) -> None:
        for level in (-123.0, 0.0, 123.0):
            stream = FractalAdaptiveMovingAverage(
                FractalAdaptiveMovingAverageParams(length=4)
            )
            for _ in range(20):
                actual_value = stream.update(level)
                if stream.is_primed():
                    self.assertEqual(actual_value, level)
                    self.assertEqual(stream.fractal_dimension(), 0.0)

        minimum = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(length=2)
        )
        for price in testdata.INPUT_CLOSE:
            actual_value = minimum.update(price)
            if minimum.is_primed():
                self.assertEqual(minimum.fractal_dimension(), 0.0)
                self.assertEqual(actual_value, price)

    def test_invalid_half_boxes_hold_previous_dimension(self) -> None:
        stream = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(length=4)
        )
        for price in (0.0, 1.0, 0.0, 1.0):
            stream.update(price)
        self.assertEqual(stream.fractal_dimension(), 2.0)
        stream.update(1.0)
        self.assertEqual(stream.fractal_dimension(), 2.0)
        for _ in range(10):
            stream.update(1.0)
            self.assertEqual(stream.fractal_dimension(), 2.0)

    def test_alpha_endpoints_and_full_precision_difference(self) -> None:
        precise = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(length=4, coefficient_mode=1)
        )
        article = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(length=4, coefficient_mode=0)
        )
        for _ in range(5):
            precise.update(1.0, 2.0, 0.0)
            article.update(1.0, 2.0, 0.0)
        self.assertEqual(precise.fractal_dimension(), 2.0)
        self.assertEqual(article.fractal_dimension(), 2.0)
        self.assertAlmostEqual(precise.alpha(), 0.01, places=15)
        self.assertEqual(article.alpha(), math.exp(-4.6))
        self.assertGreater(article.alpha(), precise.alpha())

        unit = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(slowest_smoothing_factor=1.0)
        )
        for price in testdata.INPUT_CLOSE:
            actual_value = unit.update(price)
            if unit.is_primed():
                self.assertEqual(actual_value, price)
                self.assertEqual(unit.alpha(), 1.0)

    def test_scalar_matches_explicit_equal_high_low_and_affine_invariance(self) -> None:
        scalar_stream = FractalAdaptiveMovingAverage()
        bar_stream = FractalAdaptiveMovingAverage()
        transformed_stream = FractalAdaptiveMovingAverage()
        for price in testdata.INPUT_CLOSE:
            left_val = scalar_stream.update(price)
            right_val = bar_stream.update(price, price, price)
            transformed_val = transformed_stream.update(3.0 * price - 1000.0)
            if math.isnan(left_val):
                self.assertTrue(math.isnan(right_val))
                self.assertTrue(math.isnan(transformed_val))
            else:
                self.assertEqual(left_val, right_val)
                self.assertEqual(
                    scalar_stream.fractal_dimension(), bar_stream.fractal_dimension()
                )
                self.assertAlmostEqual(
                    scalar_stream.fractal_dimension(),
                    transformed_stream.fractal_dimension(),
                    places=11,
                )
                self.assertAlmostEqual(
                    3.0 * left_val - 1000.0, transformed_val, places=10
                )

    def test_nan_gap_and_reset(self) -> None:
        stream = FractalAdaptiveMovingAverage()
        control = FractalAdaptiveMovingAverage()
        for price in testdata.INPUT_CLOSE[:80]:
            left_val = stream.update(price)
            right_val = control.update(price)
            if math.isnan(right_val):
                self.assertTrue(math.isnan(left_val))
            else:
                self.assertEqual(left_val, right_val)
                self.assertEqual(
                    stream.fractal_dimension(), control.fractal_dimension()
                )
            seen = stream.samples_seen()
            saved_val = stream.value()
            saved_dim = stream.fractal_dimension()
            self.assertTrue(math.isnan(stream.update(math.nan)))
            self.assertTrue(math.isnan(stream.update(price, math.nan, price)))
            self.assertEqual(stream.samples_seen(), seen)
            if math.isnan(saved_val):
                self.assertTrue(math.isnan(stream.value()))
                self.assertTrue(math.isnan(stream.fractal_dimension()))
            else:
                self.assertEqual(stream.value(), saved_val)
                self.assertEqual(stream.fractal_dimension(), saved_dim)

        stream.reset()
        self.assertEqual(stream.samples_seen(), 0)
        self.assertFalse(stream.is_primed())
        self.assertTrue(math.isnan(stream.value()))
        self.assertTrue(math.isnan(stream.fractal_dimension()))
        fresh = FractalAdaptiveMovingAverage()
        for price in testdata.INPUT_CLOSE:
            left_val = stream.update(price)
            right_val = fresh.update(price)
            if math.isnan(right_val):
                self.assertTrue(math.isnan(left_val))
            else:
                self.assertEqual(left_val, right_val)
                self.assertEqual(
                    stream.fractal_dimension(), fresh.fractal_dimension()
                )


class TestFractalAdaptiveMovingAverageValidationAndEntities(unittest.TestCase):
    def test_invalid_observations_preserve_state(self) -> None:
        stream = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(length=2)
        )
        stream.update(0.0)
        invalid_calls = (
            (math.inf,),
            (-math.inf,),
            (True,),
            (None,),
            ("invalid",),
            (1.0, 2.0),
            (1.0, None, 2.0),
            (1.0, 0.0, 2.0),
            (1.0, math.inf, 0.0),
            (0.0, 1e308, -1e308),
        )
        for call_args in invalid_calls:
            saved_seen = stream.samples_seen()
            with self.subTest(call_args=call_args), self.assertRaises(ValueError):
                stream.update(*call_args)  # type: ignore[arg-type]
            self.assertEqual(stream.samples_seen(), saved_seen)
            self.assertFalse(stream.is_primed())

    def test_configuration_validation(self) -> None:
        invalid_configs = (
            FractalAdaptiveMovingAverageParams(length=1),
            FractalAdaptiveMovingAverageParams(length=17),
            FractalAdaptiveMovingAverageParams(length=16.0),  # type: ignore[arg-type]
            FractalAdaptiveMovingAverageParams(length=True),  # type: ignore[arg-type]
            FractalAdaptiveMovingAverageParams(slowest_smoothing_factor=0.0),
            FractalAdaptiveMovingAverageParams(slowest_smoothing_factor=-0.1),
            FractalAdaptiveMovingAverageParams(slowest_smoothing_factor=1.1),
            FractalAdaptiveMovingAverageParams(slowest_smoothing_factor=math.nan),
            FractalAdaptiveMovingAverageParams(slowest_smoothing_factor=math.inf),
            FractalAdaptiveMovingAverageParams(slowest_smoothing_factor=True),  # type: ignore[arg-type]
            FractalAdaptiveMovingAverageParams(coefficient_mode=2),
            FractalAdaptiveMovingAverageParams(coefficient_mode=-1),
            FractalAdaptiveMovingAverageParams(coefficient_mode=True),  # type: ignore[arg-type]
            FractalAdaptiveMovingAverageParams(
                coefficient_mode=0, slowest_smoothing_factor=0.1
            ),
        )
        for bad_params in invalid_configs:
            with self.subTest(bad_params=bad_params), self.assertRaises(ValueError):
                FractalAdaptiveMovingAverage(bad_params)

    def test_metadata_mnemonics_and_factory(self) -> None:
        default_ind = FractalAdaptiveMovingAverage.create_default()
        meta = default_ind.metadata()
        self.assertEqual(meta.identifier, Identifier.FRACTAL_ADAPTIVE_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "frama(16, 0.01)")
        self.assertEqual(meta.description, "Fractal adaptive moving average frama(16, 0.01)")
        self.assertEqual(len(meta.outputs), 2)
        self.assertEqual(meta.outputs[0].kind, FractalAdaptiveMovingAverageOutput.VALUE)
        self.assertEqual(meta.outputs[0].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[0].mnemonic, "frama(16, 0.01)")
        self.assertEqual(meta.outputs[1].kind, FractalAdaptiveMovingAverageOutput.FDIM)
        self.assertEqual(meta.outputs[1].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[1].mnemonic, "framaDim(16, 0.01)")

        article_params = FractalAdaptiveMovingAverageParams(
            length=30,
            slowest_smoothing_factor=0.01,
            coefficient_mode=0,
            bar_component=BarComponent.MEDIAN,
            quote_component=QuoteComponent.BID,
            trade_component=TradeComponent.VOLUME,
        )
        self.assertEqual(
            fractal_adaptive_moving_average_mnemonic(article_params),
            "frama(30, 0.01, mode=0, hl/2, b, v)",
        )
        self.assertEqual(
            fractal_adaptive_moving_average_dimension_mnemonic(article_params),
            "framaDim(30, 0.01, mode=0, hl/2, b, v)",
        )

        from_factory = create_indicator(
            Identifier.FRACTAL_ADAPTIVE_MOVING_AVERAGE,
            {"length": 30, "slowestSmoothingFactor": 0.01, "coefficientMode": 0},
        )
        self.assertEqual(from_factory.metadata().mnemonic, "frama(30, 0.01, mode=0)")

    def test_entity_adapters_and_nan_entity_gap(self) -> None:
        base_time = datetime(2026, 1, 1)
        bar_ind = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(
                length=4, bar_component=BarComponent.MEDIAN
            )
        )
        direct_ind = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(length=4)
        )
        for step in range(8):
            stamp = base_time + timedelta(days=step)
            high = testdata.INPUT_HIGH[step]
            low = testdata.INPUT_LOW[step]
            bar = Bar(
                time=stamp,
                open=testdata.INPUT_OPEN[step],
                high=high,
                low=low,
                close=testdata.INPUT_CLOSE[step],
                volume=1000.0,
            )
            out = bar_ind.update_bar(bar)
            expected_val = direct_ind.update((high + low) / 2.0, high, low)
            self.assertEqual(len(out), 2)
            self.assertEqual(out[0].time, stamp)
            self.assertEqual(out[1].time, stamp)
            if math.isnan(expected_val):
                self.assertTrue(math.isnan(out[0].value))
                self.assertTrue(math.isnan(out[1].value))
            else:
                self.assertAlmostEqual(out[0].value, expected_val, delta=1e-15)
                self.assertAlmostEqual(
                    out[1].value, direct_ind.fractal_dimension(), delta=1e-15
                )

        saved_val = bar_ind.value()
        saved_dim = bar_ind.fractal_dimension()
        nan_bar = Bar(
            time=base_time + timedelta(days=99),
            open=1.0,
            high=2.0,
            low=1.0,
            close=math.nan,
            volume=10.0,
        )
        nan_bar_default = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(length=4)
        )
        for step in range(4):
            nan_bar_default.update(testdata.INPUT_CLOSE[step])
        gap_out = nan_bar_default.update_bar(nan_bar)
        self.assertTrue(math.isnan(gap_out[0].value))
        self.assertTrue(math.isnan(gap_out[1].value))
        self.assertFalse(math.isnan(saved_val))
        self.assertFalse(math.isnan(saved_dim))

        quote_ind = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(length=2)
        )
        quote_ind.update_quote(
            Quote(time=base_time, bid_price=10.0, ask_price=12.0, bid_size=1.0, ask_size=1.0)
        )
        q_out = quote_ind.update_quote(
            Quote(
                time=base_time + timedelta(days=1),
                bid_price=11.0,
                ask_price=13.0,
                bid_size=1.0,
                ask_size=1.0,
            )
        )
        self.assertAlmostEqual(q_out[0].value, 12.0, delta=1e-15)

        scalar_ind = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(length=2)
        )
        scalar_ind.update_scalar(Scalar(time=base_time, value=5.0))
        s_out = scalar_ind.update_scalar(
            Scalar(time=base_time + timedelta(days=1), value=7.0)
        )
        self.assertEqual(s_out[0].value, 7.0)
        self.assertEqual(s_out[1].value, 0.0)

        trade_ind = FractalAdaptiveMovingAverage(
            FractalAdaptiveMovingAverageParams(length=2)
        )
        trade_ind.update_trade(Trade(time=base_time, price=5.0, volume=10.0))
        t_out = trade_ind.update_trade(
            Trade(time=base_time + timedelta(days=1), price=8.0, volume=20.0)
        )
        self.assertEqual(t_out[0].value, 8.0)
        self.assertEqual(t_out[1].value, 0.0)
