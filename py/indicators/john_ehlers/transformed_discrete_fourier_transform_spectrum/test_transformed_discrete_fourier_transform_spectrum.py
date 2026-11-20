"""Tests for Ehlers' Transformed Discrete Fourier Transform Spectrum indicator."""

from __future__ import annotations

import datetime
import math
import unittest

from py.entities.bar import Bar
from py.entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from py.entities.quote import Quote
from py.entities.quote_component import DEFAULT_QUOTE_COMPONENT, QuoteComponent
from py.entities.scalar import Scalar
from py.entities.trade import Trade
from py.entities.trade_component import DEFAULT_TRADE_COMPONENT, TradeComponent
from py.indicators.core.adaptivity import Adaptivity
from py.indicators.core.descriptors import descriptor_of
from py.indicators.core.identifier import Identifier
from py.indicators.core.input_requirement import InputRequirement
from py.indicators.core.output import Output
from py.indicators.core.outputs.heatmap import Heatmap
from py.indicators.core.outputs.shape import Shape
from py.indicators.core.pane import Pane
from py.indicators.core.role import Role
from py.indicators.core.volume_usage import VolumeUsage
from py.indicators.factory.factory import create_indicator
from py.indicators.john_ehlers.transformed_discrete_fourier_transform_spectrum import (
    SpectrumOutput,
    SpectrumState,
    TransformedDiscreteFourierTransformSpectrum,
    TransformedDiscreteFourierTransformSpectrumOutput,
    TransformedDiscreteFourierTransformSpectrumParams,
    default_params,
    test_testdata as data,
    transformed_discrete_fourier_transform_spectrum_dominant_cycle_mnemonic,
    transformed_discrete_fourier_transform_spectrum_mnemonic,
)

FIXTURE_TOLERANCE = 1e-13
ORACLE_TOLERANCE = 3e-11

REFERENCE_CASES: tuple[
    tuple[
        str,
        TransformedDiscreteFourierTransformSpectrumParams,
        list[list[float]],
        list[float],
        list[float],
        list[float],
    ],
    ...,
] = (
    (
        "P8_Q50_W50_HP40",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=8, max_period=50, window_length=50, highpass_period=40.0
        ),
        data.EXPECTED_P8_Q50_W50_HP40_HEATMAP,
        data.EXPECTED_P8_Q50_W50_HP40_MIN,
        data.EXPECTED_P8_Q50_W50_HP40_MAX,
        data.EXPECTED_P8_Q50_W50_HP40_DOMINANT_CYCLE,
    ),
    (
        "P8_Q50_W1_HP40",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=8, max_period=50, window_length=1, highpass_period=40.0
        ),
        data.EXPECTED_P8_Q50_W1_HP40_HEATMAP,
        data.EXPECTED_P8_Q50_W1_HP40_MIN,
        data.EXPECTED_P8_Q50_W1_HP40_MAX,
        data.EXPECTED_P8_Q50_W1_HP40_DOMINANT_CYCLE,
    ),
    (
        "P8_Q50_W20_HP40",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=8, max_period=50, window_length=20, highpass_period=40.0
        ),
        data.EXPECTED_P8_Q50_W20_HP40_HEATMAP,
        data.EXPECTED_P8_Q50_W20_HP40_MIN,
        data.EXPECTED_P8_Q50_W20_HP40_MAX,
        data.EXPECTED_P8_Q50_W20_HP40_DOMINANT_CYCLE,
    ),
    (
        "P8_Q50_W48_HP40",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=8, max_period=50, window_length=48, highpass_period=40.0
        ),
        data.EXPECTED_P8_Q50_W48_HP40_HEATMAP,
        data.EXPECTED_P8_Q50_W48_HP40_MIN,
        data.EXPECTED_P8_Q50_W48_HP40_MAX,
        data.EXPECTED_P8_Q50_W48_HP40_DOMINANT_CYCLE,
    ),
    (
        "P8_Q50_W64_HP40",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=8, max_period=50, window_length=64, highpass_period=40.0
        ),
        data.EXPECTED_P8_Q50_W64_HP40_HEATMAP,
        data.EXPECTED_P8_Q50_W64_HP40_MIN,
        data.EXPECTED_P8_Q50_W64_HP40_MAX,
        data.EXPECTED_P8_Q50_W64_HP40_DOMINANT_CYCLE,
    ),
    (
        "P8_Q50_W252_HP40",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=8, max_period=50, window_length=252, highpass_period=40.0
        ),
        data.EXPECTED_P8_Q50_W252_HP40_HEATMAP,
        data.EXPECTED_P8_Q50_W252_HP40_MIN,
        data.EXPECTED_P8_Q50_W252_HP40_MAX,
        data.EXPECTED_P8_Q50_W252_HP40_DOMINANT_CYCLE,
    ),
    (
        "P8_Q50_W50_HP20",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=8, max_period=50, window_length=50, highpass_period=20.0
        ),
        data.EXPECTED_P8_Q50_W50_HP20_HEATMAP,
        data.EXPECTED_P8_Q50_W50_HP20_MIN,
        data.EXPECTED_P8_Q50_W50_HP20_MAX,
        data.EXPECTED_P8_Q50_W50_HP20_DOMINANT_CYCLE,
    ),
    (
        "P8_Q50_W50_HP80",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=8, max_period=50, window_length=50, highpass_period=80.0
        ),
        data.EXPECTED_P8_Q50_W50_HP80_HEATMAP,
        data.EXPECTED_P8_Q50_W50_HP80_MIN,
        data.EXPECTED_P8_Q50_W50_HP80_MAX,
        data.EXPECTED_P8_Q50_W50_HP80_DOMINANT_CYCLE,
    ),
    (
        "P8_Q50_W50_HP40_5",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=8, max_period=50, window_length=50, highpass_period=40.5
        ),
        data.EXPECTED_P8_Q50_W50_HP40_5_HEATMAP,
        data.EXPECTED_P8_Q50_W50_HP40_5_MIN,
        data.EXPECTED_P8_Q50_W50_HP40_5_MAX,
        data.EXPECTED_P8_Q50_W50_HP40_5_DOMINANT_CYCLE,
    ),
    (
        "P10_Q30_W50_HP40",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=10, max_period=30, window_length=50, highpass_period=40.0
        ),
        data.EXPECTED_P10_Q30_W50_HP40_HEATMAP,
        data.EXPECTED_P10_Q30_W50_HP40_MIN,
        data.EXPECTED_P10_Q30_W50_HP40_MAX,
        data.EXPECTED_P10_Q30_W50_HP40_DOMINANT_CYCLE,
    ),
    (
        "P8_Q80_W50_HP40",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=8, max_period=80, window_length=50, highpass_period=40.0
        ),
        data.EXPECTED_P8_Q80_W50_HP40_HEATMAP,
        data.EXPECTED_P8_Q80_W50_HP40_MIN,
        data.EXPECTED_P8_Q80_W50_HP40_MAX,
        data.EXPECTED_P8_Q80_W50_HP40_DOMINANT_CYCLE,
    ),
    (
        "P20_Q20_W50_HP40",
        TransformedDiscreteFourierTransformSpectrumParams(
            min_period=20, max_period=20, window_length=50, highpass_period=40.0
        ),
        data.EXPECTED_P20_Q20_W50_HP40_HEATMAP,
        data.EXPECTED_P20_Q20_W50_HP40_MIN,
        data.EXPECTED_P20_Q20_W50_HP40_MAX,
        data.EXPECTED_P20_Q20_W50_HP40_DOMINANT_CYCLE,
    ),
)


def article_batch(
    prices: list[float],
    params: TransformedDiscreteFourierTransformSpectrumParams,
) -> list[SpectrumState]:
    """Direct chronological translation of the January 2007 article equations."""
    price = [0.0] + list(prices)
    hp = [0.0] * len(price)
    cleaned = [0.0] * len(price)
    decibels = [0.0] * (params.max_period + 1)
    dominant = 0.0
    results: list[SpectrumState] = []
    for step in range(1, len(price)):
        if step <= 5:
            hp[step] = price[step]
            cleaned[step] = price[step]
        else:
            alpha = (1.0 - math.sin(2.0 * math.pi / params.highpass_period)) / math.cos(
                2.0 * math.pi / params.highpass_period
            )
            hp[step] = (
                0.5 * (1.0 + alpha) * (price[step] - price[step - 1])
                + alpha * hp[step - 1]
            )
            cleaned[step] = (
                hp[step]
                + 2.0 * hp[step - 1]
                + 3.0 * hp[step - 2]
                + 3.0 * hp[step - 3]
                + 2.0 * hp[step - 4]
                + hp[step - 5]
            ) / 12.0
        cosine = [0.0] * (params.max_period + 1)
        sine = [0.0] * (params.max_period + 1)
        power = [0.0] * (params.max_period + 1)
        for period in range(params.min_period, params.max_period + 1):
            for lag in range(params.window_length):
                sample = cleaned[step - lag] if step - lag >= 0 else 0.0
                cosine[period] += sample * math.cos(2.0 * math.pi * lag / period)
                sine[period] += sample * math.sin(2.0 * math.pi * lag / period)
            power[period] = (
                cosine[period] * cosine[period] + sine[period] * sine[period]
            )
        maximum = power[params.min_period]
        for period in range(params.min_period, params.max_period + 1):
            if power[period] > maximum:
                maximum = power[period]
        for period in range(params.min_period, params.max_period + 1):
            if maximum > 0.0 and power[period] > 0.0:
                decibels[period] = (
                    -10.0
                    * math.log(0.01 / (1.0 - 0.99 * power[period] / maximum))
                    / math.log(10.0)
                )
            if decibels[period] > 20.0:
                decibels[period] = 20.0
        numerator = 0.0
        denominator = 0.0
        for period in range(params.min_period, params.max_period + 1):
            if decibels[period] < 3.0:
                numerator += period * (3.0 - decibels[period])
                denominator += 3.0 - decibels[period]
        if denominator != 0.0:
            dominant = numerator / denominator
        results.append(
            SpectrumState(
                hp[step],
                cleaned[step],
                maximum,
                dominant,
                tuple(cosine[params.min_period :]),
                tuple(sine[params.min_period :]),
                tuple(power[params.min_period :]),
                tuple(decibels[params.min_period :]),
            )
        )
    return results


class TestTransformedDiscreteFourierTransformSpectrum(unittest.TestCase):
    """Unit tests for TransformedDiscreteFourierTransformSpectrum."""

    def _assert_close(self, actual: float, expected: float, tolerance: float) -> None:
        if math.isnan(expected):
            self.assertTrue(math.isnan(actual), f"expected NaN, got {actual!r}")
            return
        self.assertFalse(math.isnan(actual), f"expected {expected!r}, got NaN")
        bound = max(tolerance, abs(expected) * tolerance)
        self.assertAlmostEqual(actual, expected, delta=bound)

    def _assert_same_float(self, actual: float, expected: float) -> None:
        if math.isnan(expected):
            self.assertTrue(math.isnan(actual), f"expected NaN, got {actual!r}")
            return
        self.assertEqual(actual, expected)

    def _assert_same_heatmap(self, actual: Heatmap, expected: Heatmap) -> None:
        self.assertEqual(actual.time, expected.time)
        self.assertEqual(actual.parameter_first, expected.parameter_first)
        self.assertEqual(actual.parameter_last, expected.parameter_last)
        self.assertEqual(actual.parameter_resolution, expected.parameter_resolution)
        self.assertEqual(actual.is_empty(), expected.is_empty())
        self._assert_same_float(actual.value_min, expected.value_min)
        self._assert_same_float(actual.value_max, expected.value_max)
        self.assertEqual(actual.values, expected.values)

    def _assert_same_spectrum_output(
        self, actual: SpectrumOutput, expected: SpectrumOutput
    ) -> None:
        self._assert_same_heatmap(actual.heatmap, expected.heatmap)
        self._assert_same_float(actual.dominant_cycle, expected.dominant_cycle)

    def _assert_same_entity_output(self, actual: Output, expected: Output) -> None:
        self.assertEqual(len(actual), 2)
        self.assertEqual(len(expected), 2)
        self.assertIsInstance(actual[0], Heatmap)
        self.assertIsInstance(expected[0], Heatmap)
        self.assertIsInstance(actual[1], Scalar)
        self.assertIsInstance(expected[1], Scalar)
        self._assert_same_heatmap(actual[0], expected[0])
        self.assertEqual(actual[1].time, expected[1].time)
        self._assert_same_float(actual[1].value, expected[1].value)

    def test_all_reference_fixture_cases(self) -> None:
        base_time = datetime.datetime(2026, 1, 1)
        for (
            case_name,
            case_params,
            expected_heatmap,
            expected_min,
            expected_max,
            expected_dc,
        ) in REFERENCE_CASES:
            with self.subTest(case=case_name):
                indicator = TransformedDiscreteFourierTransformSpectrum.create(
                    case_params
                )
                channel_count = case_params.max_period - case_params.min_period + 1
                for index, close_price in enumerate(data.INPUT_CLOSE):
                    timestamp = base_time + datetime.timedelta(days=index)
                    result = indicator.update(close_price, timestamp)
                    self.assertEqual(result.heatmap.time, timestamp)
                    self.assertEqual(
                        result.heatmap.parameter_first, float(case_params.min_period)
                    )
                    self.assertEqual(
                        result.heatmap.parameter_last, float(case_params.max_period)
                    )
                    self.assertEqual(result.heatmap.parameter_resolution, 1.0)

                    row = expected_heatmap[index]
                    if math.isnan(expected_dc[index]):
                        self.assertFalse(indicator.is_primed())
                        self.assertTrue(result.heatmap.is_empty())
                        self.assertTrue(math.isnan(result.heatmap.value_min))
                        self.assertTrue(math.isnan(result.heatmap.value_max))
                        self.assertTrue(math.isnan(result.dominant_cycle))
                        self.assertTrue(all(math.isnan(cell) for cell in row))
                    else:
                        self.assertTrue(indicator.is_primed())
                        self.assertFalse(result.heatmap.is_empty())
                        self.assertEqual(len(result.heatmap.values), channel_count)
                        self._assert_close(
                            result.heatmap.value_min,
                            expected_min[index],
                            FIXTURE_TOLERANCE,
                        )
                        self._assert_close(
                            result.heatmap.value_max,
                            expected_max[index],
                            FIXTURE_TOLERANCE,
                        )
                        self._assert_close(
                            result.dominant_cycle,
                            expected_dc[index],
                            FIXTURE_TOLERANCE,
                        )
                        for actual_cell, expected_cell in zip(
                            result.heatmap.values, row
                        ):
                            self._assert_close(
                                actual_cell, expected_cell, FIXTURE_TOLERANCE
                            )

    def test_independent_article_oracle_across_all_cases(self) -> None:
        for case_name, case_params, _, _, _, _ in REFERENCE_CASES:
            with self.subTest(case=case_name):
                stream = TransformedDiscreteFourierTransformSpectrum(case_params)
                oracle = article_batch(data.INPUT_CLOSE, case_params)
                for step, (price, wanted) in enumerate(
                    zip(data.INPUT_CLOSE, oracle), 1
                ):
                    result = stream.update(price, step)
                    state = stream.state
                    self._assert_close(state.high_pass, wanted.high_pass, ORACLE_TOLERANCE)
                    self._assert_close(
                        state.cleaned_data, wanted.cleaned_data, ORACLE_TOLERANCE
                    )
                    self._assert_close(
                        state.maximum_power, wanted.maximum_power, ORACLE_TOLERANCE
                    )
                    self._assert_close(
                        state.dominant_cycle, wanted.dominant_cycle, ORACLE_TOLERANCE
                    )
                    for actual_vec, wanted_vec in (
                        (state.cosine_parts, wanted.cosine_parts),
                        (state.sine_parts, wanted.sine_parts),
                        (state.powers, wanted.powers),
                        (state.decibels, wanted.decibels),
                    ):
                        for actual_item, wanted_item in zip(actual_vec, wanted_vec):
                            self._assert_close(
                                actual_item, wanted_item, ORACLE_TOLERANCE
                            )
                    self.assertEqual(
                        result.heatmap.is_empty(), step <= stream.warmup_bars
                    )

    def test_startup_passthrough_and_sixth_bar_fir(self) -> None:
        stream = TransformedDiscreteFourierTransformSpectrum()
        samples = [100.0, 102.0, 101.0, 105.0, 103.0, 107.0]
        for step, sample in enumerate(samples[:5], 1):
            stream.update(sample)
            self.assertEqual(stream.samples_seen, step)
            self.assertEqual(stream.state.high_pass, sample)
            self.assertEqual(stream.state.cleaned_data, sample)
        stream.update(samples[5])
        alpha = stream.highpass_alpha
        expected_hp = 0.5 * (1.0 + alpha) * (samples[5] - samples[4]) + alpha * samples[4]
        expected_cleaned = (
            expected_hp
            + 2.0 * samples[4]
            + 3.0 * samples[3]
            + 3.0 * samples[2]
            + 2.0 * samples[1]
            + samples[0]
        ) / 12.0
        self._assert_close(stream.state.high_pass, expected_hp, ORACLE_TOLERANCE)
        self._assert_close(stream.state.cleaned_data, expected_cleaned, ORACLE_TOLERANCE)

    def test_silence_guarded_db_and_centroid_retention(self) -> None:
        stream = TransformedDiscreteFourierTransformSpectrum()
        for _ in range(70):
            result = stream.update(0.0)
        self.assertEqual(result.heatmap.values, [0.0] * 43)
        self.assertEqual(result.dominant_cycle, 29.0)

        stream._decibels = [17.0] * 43
        stream._decibels[10 - stream.min_period] = 1.0
        stream._decibels[20 - stream.min_period] = 2.5
        result = stream.update(0.0)
        self.assertEqual(result.dominant_cycle, 12.0)

        stream._decibels = [17.0] * 43
        result = stream.update(0.0)
        self.assertEqual(result.dominant_cycle, 12.0)
        self.assertEqual(result.heatmap.values, [17.0] * 43)

    def test_no_agc_and_synthetic_twenty_bar_cycle(self) -> None:
        stream = TransformedDiscreteFourierTransformSpectrum()
        saw_decay = False
        previous_maximum = 0.0
        for price in data.INPUT_CLOSE:
            stream.update(price)
            current_max = stream.state.maximum_power
            self.assertEqual(current_max, max(stream.state.powers))
            saw_decay = saw_decay or (current_max < 0.995 * previous_maximum)
            previous_maximum = current_max
        self.assertTrue(saw_decay)

        sine_stream = TransformedDiscreteFourierTransformSpectrum(
            TransformedDiscreteFourierTransformSpectrumParams(window_length=120)
        )
        for step in range(1000):
            sine_out = sine_stream.update(math.sin(2.0 * math.pi * step / 20.0))
            if step >= 900:
                self._assert_close(sine_out.dominant_cycle, 20.0, ORACLE_TOLERANCE)

    def test_missing_invalid_overflow_and_reset(self) -> None:
        stream = TransformedDiscreteFourierTransformSpectrum()
        control = TransformedDiscreteFourierTransformSpectrum()
        for price in data.INPUT_CLOSE[:80]:
            stream.update(price)
            control.update(price)
        before_state = stream.state
        before_output = stream.output
        before_count = stream.samples_seen

        missing = stream.update(math.nan, "missing")
        self.assertTrue(missing.heatmap.is_empty())
        self.assertEqual(missing.heatmap.time, "missing")
        self.assertTrue(math.isnan(missing.dominant_cycle))

        for invalid in (math.inf, -math.inf, 1e308, True, None, "1"):
            with self.assertRaises(ValueError):
                stream.update(invalid)  # type: ignore[arg-type]
            self.assertEqual(stream.state, before_state)
            self.assertEqual(stream.samples_seen, before_count)
            self.assertIs(stream.output, before_output)

        self._assert_same_spectrum_output(
            stream.update(data.INPUT_CLOSE[80]),
            control.update(data.INPUT_CLOSE[80]),
        )

        snapshot = stream.update(data.INPUT_CLOSE[81])
        control.update(data.INPUT_CLOSE[81])
        saved_values = snapshot.heatmap.values.copy()
        snapshot.heatmap.values[0] = -999.0
        self._assert_same_spectrum_output(
            stream.update(data.INPUT_CLOSE[82]),
            control.update(data.INPUT_CLOSE[82]),
        )
        self.assertNotEqual(saved_values[0], -999.0)

        stream.reset()
        replayed = TransformedDiscreteFourierTransformSpectrum()
        for price in data.INPUT_CLOSE:
            actual = stream.update(price)
            expected = replayed.update(price)
            if stream.is_primed():
                self._assert_same_spectrum_output(actual, expected)
            else:
                self.assertTrue(actual.heatmap.is_empty())

    def test_parameter_validation(self) -> None:
        invalid_cases = (
            TransformedDiscreteFourierTransformSpectrumParams(min_period=7),
            TransformedDiscreteFourierTransformSpectrumParams(min_period=20, max_period=19),
            TransformedDiscreteFourierTransformSpectrumParams(max_period=True),  # type: ignore[arg-type]
            TransformedDiscreteFourierTransformSpectrumParams(window_length=0),
            TransformedDiscreteFourierTransformSpectrumParams(window_length=3.0),  # type: ignore[arg-type]
            TransformedDiscreteFourierTransformSpectrumParams(highpass_period=4.0),
            TransformedDiscreteFourierTransformSpectrumParams(highpass_period=math.inf),
            TransformedDiscreteFourierTransformSpectrumParams(highpass_period=math.nan),
            TransformedDiscreteFourierTransformSpectrumParams(highpass_period=True),  # type: ignore[arg-type]
        )
        for invalid in invalid_cases:
            with self.subTest(params=invalid), self.assertRaises(ValueError):
                TransformedDiscreteFourierTransformSpectrum(invalid)

    def test_mnemonics_metadata_and_descriptor(self) -> None:
        cases = (
            (default_params(), "tdfts(8, 50, 50, 40)", "tdfts-dc(8, 50, 50, 40)"),
            (
                TransformedDiscreteFourierTransformSpectrumParams(
                    bar_component=DEFAULT_BAR_COMPONENT,
                    quote_component=DEFAULT_QUOTE_COMPONENT,
                    trade_component=DEFAULT_TRADE_COMPONENT,
                ),
                "tdfts(8, 50, 50, 40)",
                "tdfts-dc(8, 50, 50, 40)",
            ),
            (
                TransformedDiscreteFourierTransformSpectrumParams(
                    bar_component=BarComponent.MEDIAN
                ),
                "tdfts(8, 50, 50, 40, hl/2)",
                "tdfts-dc(8, 50, 50, 40, hl/2)",
            ),
            (
                TransformedDiscreteFourierTransformSpectrumParams(
                    quote_component=QuoteComponent.BID
                ),
                "tdfts(8, 50, 50, 40, b)",
                "tdfts-dc(8, 50, 50, 40, b)",
            ),
            (
                TransformedDiscreteFourierTransformSpectrumParams(
                    trade_component=TradeComponent.VOLUME
                ),
                "tdfts(8, 50, 50, 40, v)",
                "tdfts-dc(8, 50, 50, 40, v)",
            ),
            (
                TransformedDiscreteFourierTransformSpectrumParams(
                    min_period=10,
                    max_period=30,
                    window_length=48,
                    highpass_period=40.5,
                    bar_component=BarComponent.MEDIAN,
                    quote_component=QuoteComponent.ASK,
                ),
                "tdfts(10, 30, 48, 40.5, hl/2, a)",
                "tdfts-dc(10, 30, 48, 40.5, hl/2, a)",
            ),
            (
                TransformedDiscreteFourierTransformSpectrumParams(
                    quote_component=QuoteComponent.BID,
                    trade_component=TradeComponent.VOLUME,
                ),
                "tdfts(8, 50, 50, 40, b, v)",
                "tdfts-dc(8, 50, 50, 40, b, v)",
            ),
        )
        for case_params, want_main, want_dc in cases:
            with self.subTest(want=want_main):
                self.assertEqual(
                    transformed_discrete_fourier_transform_spectrum_mnemonic(
                        case_params
                    ),
                    want_main,
                )
                self.assertEqual(
                    transformed_discrete_fourier_transform_spectrum_dominant_cycle_mnemonic(
                        case_params
                    ),
                    want_dc,
                )
                instance = TransformedDiscreteFourierTransformSpectrum(case_params)
                meta = instance.metadata()
                self.assertEqual(
                    meta.identifier,
                    Identifier.TRANSFORMED_DISCRETE_FOURIER_TRANSFORM_SPECTRUM,
                )
                self.assertEqual(meta.mnemonic, want_main)
                self.assertEqual(len(meta.outputs), 2)
                self.assertEqual(
                    meta.outputs[0].kind,
                    int(TransformedDiscreteFourierTransformSpectrumOutput.VALUE),
                )
                self.assertEqual(meta.outputs[0].shape, Shape.HEATMAP)
                self.assertEqual(meta.outputs[0].mnemonic, want_main)
                self.assertEqual(
                    meta.outputs[1].kind,
                    int(
                        TransformedDiscreteFourierTransformSpectrumOutput.DOMINANT_CYCLE
                    ),
                )
                self.assertEqual(meta.outputs[1].shape, Shape.SCALAR)
                self.assertEqual(meta.outputs[1].mnemonic, want_dc)

        desc = descriptor_of(Identifier.TRANSFORMED_DISCRETE_FOURIER_TRANSFORM_SPECTRUM)
        self.assertIsNotNone(desc)
        assert desc is not None
        self.assertEqual(desc.family, "John Ehlers")
        self.assertEqual(desc.adaptivity, Adaptivity.STATIC)
        self.assertEqual(desc.input_requirement, InputRequirement.SCALAR_INPUT)
        self.assertEqual(desc.volume_usage, VolumeUsage.NO_VOLUME)
        self.assertEqual(len(desc.outputs), 2)
        self.assertEqual(desc.outputs[0].role, Role.SPECTRUM)
        self.assertEqual(desc.outputs[0].pane, Pane.OWN)
        self.assertEqual(desc.outputs[1].role, Role.CYCLE_PERIOD)
        self.assertEqual(desc.outputs[1].pane, Pane.OWN)

    def test_entity_updates_and_factory(self) -> None:
        scalar_ind = TransformedDiscreteFourierTransformSpectrum()
        bar_ind = TransformedDiscreteFourierTransformSpectrum()
        quote_ind = TransformedDiscreteFourierTransformSpectrum()
        trade_ind = TransformedDiscreteFourierTransformSpectrum()
        open_ind = TransformedDiscreteFourierTransformSpectrum(
            TransformedDiscreteFourierTransformSpectrumParams(
                bar_component=BarComponent.OPEN
            )
        )
        open_ref = TransformedDiscreteFourierTransformSpectrum()

        base_time = datetime.datetime(2026, 1, 1)
        for index in range(len(data.INPUT_CLOSE)):
            timestamp = base_time + datetime.timedelta(days=index)
            close_price = data.INPUT_CLOSE[index]
            open_price = data.INPUT_OPEN[index]
            high_price = data.INPUT_HIGH[index]
            low_price = data.INPUT_LOW[index]

            out_scalar = scalar_ind.update_scalar(Scalar(timestamp, close_price))
            out_bar = bar_ind.update_bar(
                Bar(timestamp, open_price, high_price, low_price, close_price, 1000.0)
            )
            out_quote = quote_ind.update_quote(
                Quote(timestamp, close_price, close_price, 10.0, 10.0)
            )
            out_trade = trade_ind.update_trade(Trade(timestamp, close_price, 10.0))
            self._assert_same_entity_output(out_scalar, out_bar)
            self._assert_same_entity_output(out_scalar, out_quote)
            self._assert_same_entity_output(out_scalar, out_trade)

            out_open_bar = open_ind.update_bar(
                Bar(timestamp, open_price, high_price, low_price, close_price, 1000.0)
            )
            out_open_scalar = open_ref.update_scalar(Scalar(timestamp, open_price))
            self._assert_same_entity_output(out_open_bar, out_open_scalar)

        default_created = create_indicator(
            Identifier.TRANSFORMED_DISCRETE_FOURIER_TRANSFORM_SPECTRUM, {}
        )
        self.assertEqual(default_created.metadata().mnemonic, "tdfts(8, 50, 50, 40)")

        custom_created = create_indicator(
            Identifier.TRANSFORMED_DISCRETE_FOURIER_TRANSFORM_SPECTRUM,
            {
                "minPeriod": 10,
                "maxPeriod": 30,
                "windowLength": 48,
                "highpassPeriod": 40.5,
                "barComponent": 5,
            },
        )
        self.assertEqual(
            custom_created.metadata().mnemonic, "tdfts(10, 30, 48, 40.5, hl/2)"
        )

        for bad_params in (
            {"unknownKey": 1},
            {"minPeriod": None},
            {"minPeriod": 7},
            {"windowLength": 0},
            {"highpassPeriod": 4.0},
            {"minPeriod": 10, "min_period": 10},
        ):
            with self.assertRaises(ValueError):
                create_indicator(
                    Identifier.TRANSFORMED_DISCRETE_FOURIER_TRANSFORM_SPECTRUM,
                    bad_params,
                )


if __name__ == "__main__":
    unittest.main()
