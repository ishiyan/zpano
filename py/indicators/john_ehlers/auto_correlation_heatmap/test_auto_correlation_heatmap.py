"""Tests for Ehlers' February 2025 Autocorrelation Heatmap indicator."""

from __future__ import annotations

import datetime
import math
import unittest

from py.entities.bar import Bar
from py.entities.bar_component import BarComponent
from py.entities.quote import Quote
from py.entities.quote_component import QuoteComponent
from py.entities.scalar import Scalar
from py.entities.trade import Trade
from py.entities.trade_component import TradeComponent
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
from py.indicators.john_ehlers.auto_correlation_heatmap import (
    AutoCorrelationHeatmap,
    AutoCorrelationHeatmapOutput,
    AutoCorrelationHeatmapParams,
    AutocorrelationState,
    auto_correlation_heatmap_mnemonic,
    default_params,
    test_testdata as data,
)

FIXTURE_TOLERANCE = 1e-13
ORACLE_TOLERANCE = 1e-13

REFERENCE_CASES: tuple[
    tuple[
        str,
        AutoCorrelationHeatmapParams,
        list[list[float]],
        list[float],
        list[float],
    ],
    ...,
] = (
    (
        "L20_P20_H99",
        AutoCorrelationHeatmapParams(length=20, smoothing_period=20.0, max_lag=99),
        data.EXPECTED_L20_P20_H99_HEATMAP,
        data.EXPECTED_L20_P20_H99_MIN,
        data.EXPECTED_L20_P20_H99_MAX,
    ),
    (
        "L2_P20_H99",
        AutoCorrelationHeatmapParams(length=2, smoothing_period=20.0, max_lag=99),
        data.EXPECTED_L2_P20_H99_HEATMAP,
        data.EXPECTED_L2_P20_H99_MIN,
        data.EXPECTED_L2_P20_H99_MAX,
    ),
    (
        "L3_P20_H99",
        AutoCorrelationHeatmapParams(length=3, smoothing_period=20.0, max_lag=99),
        data.EXPECTED_L3_P20_H99_HEATMAP,
        data.EXPECTED_L3_P20_H99_MIN,
        data.EXPECTED_L3_P20_H99_MAX,
    ),
    (
        "L5_P20_H99",
        AutoCorrelationHeatmapParams(length=5, smoothing_period=20.0, max_lag=99),
        data.EXPECTED_L5_P20_H99_HEATMAP,
        data.EXPECTED_L5_P20_H99_MIN,
        data.EXPECTED_L5_P20_H99_MAX,
    ),
    (
        "L40_P20_H99",
        AutoCorrelationHeatmapParams(length=40, smoothing_period=20.0, max_lag=99),
        data.EXPECTED_L40_P20_H99_HEATMAP,
        data.EXPECTED_L40_P20_H99_MIN,
        data.EXPECTED_L40_P20_H99_MAX,
    ),
    (
        "L100_P20_H99",
        AutoCorrelationHeatmapParams(length=100, smoothing_period=20.0, max_lag=99),
        data.EXPECTED_L100_P20_H99_HEATMAP,
        data.EXPECTED_L100_P20_H99_MIN,
        data.EXPECTED_L100_P20_H99_MAX,
    ),
    (
        "L20_P10_H99",
        AutoCorrelationHeatmapParams(length=20, smoothing_period=10.0, max_lag=99),
        data.EXPECTED_L20_P10_H99_HEATMAP,
        data.EXPECTED_L20_P10_H99_MIN,
        data.EXPECTED_L20_P10_H99_MAX,
    ),
    (
        "L20_P10_5_H99",
        AutoCorrelationHeatmapParams(length=20, smoothing_period=10.5, max_lag=99),
        data.EXPECTED_L20_P10_5_H99_HEATMAP,
        data.EXPECTED_L20_P10_5_H99_MIN,
        data.EXPECTED_L20_P10_5_H99_MAX,
    ),
    (
        "L20_P40_H99",
        AutoCorrelationHeatmapParams(length=20, smoothing_period=40.0, max_lag=99),
        data.EXPECTED_L20_P40_H99_HEATMAP,
        data.EXPECTED_L20_P40_H99_MIN,
        data.EXPECTED_L20_P40_H99_MAX,
    ),
    (
        "L20_P20_H32",
        AutoCorrelationHeatmapParams(length=20, smoothing_period=20.0, max_lag=32),
        data.EXPECTED_L20_P20_H32_HEATMAP,
        data.EXPECTED_L20_P20_H32_MIN,
        data.EXPECTED_L20_P20_H32_MAX,
    ),
    (
        "L2_P20_H32",
        AutoCorrelationHeatmapParams(length=2, smoothing_period=20.0, max_lag=32),
        data.EXPECTED_L2_P20_H32_HEATMAP,
        data.EXPECTED_L2_P20_H32_MIN,
        data.EXPECTED_L2_P20_H32_MAX,
    ),
    (
        "L20_P20_H0",
        AutoCorrelationHeatmapParams(length=20, smoothing_period=20.0, max_lag=0),
        data.EXPECTED_L20_P20_H0_HEATMAP,
        data.EXPECTED_L20_P20_H0_MIN,
        data.EXPECTED_L20_P20_H0_MAX,
    ),
)


def _batch_source(
    closes: list[float],
    length: int,
    smoothing_period: float,
    max_lag: int,
) -> list[tuple[float, list[float] | None]]:
    """Independent literal TASC/MESA batch oracle with 1-based Corr[lag+1]."""
    filtered_series = [0.0] * len(closes)
    correlations = [0.0] * 101
    angle = math.sqrt(2.0) * math.pi / smoothing_period
    attenuation = math.exp(-angle)
    coeff_c2 = 2.0 * attenuation * math.cos(angle)
    coeff_c3 = -attenuation * attenuation
    coeff_c1 = (1.0 + coeff_c2 - coeff_c3) / 4.0
    outputs: list[tuple[float, list[float] | None]] = []
    for bar_index, price in enumerate(closes):
        if bar_index < 3:
            filtered_series[bar_index] = price
        else:
            filtered_series[bar_index] = (
                (1.0 - coeff_c1) * price
                + (2.0 * coeff_c1 - coeff_c2) * closes[bar_index - 1]
                - (coeff_c1 + coeff_c3) * closes[bar_index - 2]
                + coeff_c2 * filtered_series[bar_index - 1]
                + coeff_c3 * filtered_series[bar_index - 2]
            )
        for lag in range(max_lag + 1):
            sum_x = 0.0
            sum_y = 0.0
            sum_xx = 0.0
            sum_xy = 0.0
            sum_yy = 0.0
            for offset in range(length):
                sample_x = (
                    filtered_series[bar_index - offset]
                    if bar_index - offset >= 0
                    else 0.0
                )
                sample_y = (
                    filtered_series[bar_index - lag - offset]
                    if bar_index - lag - offset >= 0
                    else 0.0
                )
                sum_x += sample_x
                sum_y += sample_y
                sum_xx += sample_x * sample_x
                sum_xy += sample_x * sample_y
                sum_yy += sample_y * sample_y
            var_x = length * sum_xx - sum_x * sum_x
            var_y = length * sum_yy - sum_y * sum_y
            if var_x > 0.0 and var_y > 0.0:
                correlations[lag + 1] = (length * sum_xy - sum_x * sum_y) / math.sqrt(
                    var_x * var_y
                )
        if bar_index + 1 < max(4, length + max_lag):
            outputs.append((filtered_series[bar_index], None))
        else:
            outputs.append((filtered_series[bar_index], correlations[1 : max_lag + 2]))
    return outputs


class TestAutoCorrelationHeatmap(unittest.TestCase):
    """Unit tests for AutoCorrelationHeatmap."""

    def _assert_close_or_nan(
        self, actual: float, expected: float, tolerance: float
    ) -> None:
        if math.isnan(expected):
            self.assertTrue(math.isnan(actual))
        else:
            self.assertFalse(math.isnan(actual))
            self.assertAlmostEqual(actual, expected, delta=tolerance)

    def _assert_same_entity_output(self, first: Output, second: Output) -> None:
        self.assertEqual(len(first), 1)
        self.assertEqual(len(second), 1)
        first_heatmap = first[0]
        second_heatmap = second[0]
        self.assertIsInstance(first_heatmap, Heatmap)
        self.assertIsInstance(second_heatmap, Heatmap)
        self.assertEqual(first_heatmap.time, second_heatmap.time)
        self.assertEqual(first_heatmap.parameter_first, second_heatmap.parameter_first)
        self.assertEqual(first_heatmap.parameter_last, second_heatmap.parameter_last)
        self.assertEqual(
            first_heatmap.parameter_resolution, second_heatmap.parameter_resolution
        )
        self._assert_close_or_nan(
            first_heatmap.value_min, second_heatmap.value_min, FIXTURE_TOLERANCE
        )
        self._assert_close_or_nan(
            first_heatmap.value_max, second_heatmap.value_max, FIXTURE_TOLERANCE
        )
        self.assertEqual(len(first_heatmap.values), len(second_heatmap.values))
        for actual_item, expected_item in zip(
            first_heatmap.values, second_heatmap.values
        ):
            self._assert_close_or_nan(actual_item, expected_item, FIXTURE_TOLERANCE)

    def test_all_reference_cases_against_saved_fixtures(self) -> None:
        self.assertEqual(len(data.INPUT_CLOSE), 252)
        for (
            name,
            case_params,
            expected_heatmap,
            expected_min,
            expected_max,
        ) in REFERENCE_CASES:
            with self.subTest(case=name):
                indicator = AutoCorrelationHeatmap(case_params)
                warmup = max(4, case_params.length + case_params.max_lag) - 1
                self.assertEqual(indicator.warmup_bars, warmup)
                bin_count = case_params.max_lag + 1

                for bar_index, price in enumerate(data.INPUT_CLOSE):
                    result = indicator.update(price)
                    self.assertEqual(result.parameter_first, 0.0)
                    self.assertEqual(result.parameter_last, float(case_params.max_lag))
                    self.assertEqual(result.parameter_resolution, 1.0)
                    self.assertEqual(indicator.is_primed(), bar_index >= warmup)

                    expected_row = expected_heatmap[bar_index]
                    self.assertEqual(len(expected_row), bin_count)
                    if bar_index < warmup:
                        self.assertTrue(result.is_empty())
                        self.assertTrue(math.isnan(result.value_min))
                        self.assertTrue(math.isnan(result.value_max))
                        self.assertTrue(math.isnan(expected_min[bar_index]))
                        self.assertTrue(math.isnan(expected_max[bar_index]))
                        for item in expected_row:
                            self.assertTrue(math.isnan(item))
                    else:
                        self.assertFalse(result.is_empty())
                        self.assertEqual(len(result.values), bin_count)
                        self._assert_close_or_nan(
                            result.value_min, expected_min[bar_index], FIXTURE_TOLERANCE
                        )
                        self._assert_close_or_nan(
                            result.value_max, expected_max[bar_index], FIXTURE_TOLERANCE
                        )
                        for lag_index in range(bin_count):
                            self._assert_close_or_nan(
                                result.values[lag_index],
                                expected_row[lag_index],
                                FIXTURE_TOLERANCE,
                            )

    def test_all_reference_cases_against_independent_batch_oracle(self) -> None:
        for name, case_params, _, _, _ in REFERENCE_CASES:
            with self.subTest(case=name):
                indicator = AutoCorrelationHeatmap(case_params)
                oracle = _batch_source(
                    data.INPUT_CLOSE,
                    case_params.length,
                    case_params.smoothing_period,
                    case_params.max_lag,
                )
                for bar_index, (price, (filtered, bins)) in enumerate(
                    zip(data.INPUT_CLOSE, oracle)
                ):
                    actual = indicator.update(price)
                    self.assertEqual(indicator.samples_seen, bar_index + 1)
                    self.assertAlmostEqual(
                        indicator.state.filtered, filtered, delta=ORACLE_TOLERANCE
                    )
                    if bins is None:
                        self.assertTrue(actual.is_empty())
                    else:
                        self.assertEqual(len(actual.values), len(bins))
                        for actual_bin, expected_bin in zip(actual.values, bins):
                            self.assertAlmostEqual(
                                actual_bin, expected_bin, delta=ORACLE_TOLERANCE
                            )
                        self.assertAlmostEqual(
                            actual.value_min, min(bins), delta=ORACLE_TOLERANCE
                        )
                        self.assertAlmostEqual(
                            actual.value_max, max(bins), delta=ORACLE_TOLERANCE
                        )

    def test_nan_skip_overflow_rollback_and_reset(self) -> None:
        params = AutoCorrelationHeatmapParams(
            length=5, smoothing_period=10.0, max_lag=12
        )
        subject = AutoCorrelationHeatmap(params)
        clean = AutoCorrelationHeatmap(params)

        for bar_index, price in enumerate(data.INPUT_CLOSE[:50]):
            before = subject.state
            skipped = subject.update(math.nan)
            self.assertTrue(skipped.is_empty())
            self.assertEqual(subject.state, before)

            for invalid_sample in (math.inf, -math.inf, 1e308):
                with self.assertRaises(ValueError):
                    subject.update(invalid_sample)
                self.assertEqual(subject.state, before)

            actual = subject.update(price)
            expected = clean.update(price)
            self.assertEqual(subject.is_primed(), bar_index + 1 >= 17)
            self.assertEqual(len(actual.values), len(expected.values))
            for actual_item, expected_item in zip(actual.values, expected.values):
                self.assertAlmostEqual(
                    actual_item, expected_item, delta=FIXTURE_TOLERANCE
                )

        subject.reset()
        clean.reset()
        self.assertEqual(subject.state, clean.state)
        self.assertEqual(subject.samples_seen, 0)
        self.assertFalse(subject.is_primed())
        for price in data.INPUT_CLOSE[:50]:
            actual = subject.update(price)
            expected = clean.update(price)
            self.assertEqual(actual.values, expected.values)

    def test_analytical_sine_pattern_and_two_point_limit(self) -> None:
        sine = [math.sin(2.0 * math.pi * step / 20.0) for step in range(1000)]
        indicator = AutoCorrelationHeatmap(
            AutoCorrelationHeatmapParams(length=20, smoothing_period=20.0, max_lag=40)
        )
        column = Heatmap.empty(None, 0.0, 40.0, 1.0)
        for price in sine:
            column = indicator.update(price)
        ideal = [math.cos(2.0 * math.pi * lag / 20.0) for lag in range(41)]
        for actual_item, expected_item in zip(column.values, ideal):
            self.assertAlmostEqual(actual_item, expected_item, delta=1e-11)

        short = AutoCorrelationHeatmap(
            AutoCorrelationHeatmapParams(length=2, smoothing_period=20.0, max_lag=20)
        )
        filtered_history: list[float] = []
        for bar_index, price in enumerate(sine):
            short_column = short.update(price)
            filtered_history.insert(0, short.state.filtered)
            if short.is_primed() and bar_index > 300:
                for lag, actual_item in enumerate(short_column.values):
                    direction = (filtered_history[0] - filtered_history[1]) * (
                        filtered_history[lag] - filtered_history[lag + 1]
                    )
                    expected_sign = 1.0 if direction > 0.0 else -1.0
                    self.assertAlmostEqual(actual_item, expected_sign, delta=2e-9)

    def test_degenerate_variance_retains_previous_and_zero_start(self) -> None:
        indicator = AutoCorrelationHeatmap(
            AutoCorrelationHeatmapParams(length=3, smoothing_period=20.0, max_lag=2)
        )
        result = Heatmap.empty(None, 0.0, 2.0, 1.0)
        for _ in range(30):
            result = indicator.update(0.0)
        self.assertEqual(result.values, [0.0, 0.0, 0.0])

        indicator._correlations = [0.25, -0.5, 0.75]
        retained = indicator.update(0.0)
        self.assertEqual(retained.values, [0.25, -0.5, 0.75])
        self.assertEqual(retained.value_min, -0.5)
        self.assertEqual(retained.value_max, 0.75)

    def test_snapshots_common_lag_independence_and_validation(self) -> None:
        stamp = datetime.datetime(2026, 2, 1, tzinfo=datetime.timezone.utc)
        wide = AutoCorrelationHeatmap(
            AutoCorrelationHeatmapParams(length=5, smoothing_period=20.0, max_lag=99)
        )
        compact = AutoCorrelationHeatmap(
            AutoCorrelationHeatmapParams(length=5, smoothing_period=20.0, max_lag=32)
        )
        zero_lag = AutoCorrelationHeatmap(
            AutoCorrelationHeatmapParams(length=5, smoothing_period=20.0, max_lag=0)
        )

        for price in data.INPUT_CLOSE:
            wide_col = wide.update(price, stamp)
            compact_col = compact.update(price, stamp)
            zero_lag.update(price, stamp)
            self.assertIsInstance(wide.state, AutocorrelationState)
            self.assertEqual(wide.state.correlations[:33], compact.state.correlations)
            self.assertEqual(wide.state.correlations[0], zero_lag.state.correlations[0])
            self.assertEqual(wide_col.time, stamp)
            if not wide_col.is_empty():
                self.assertEqual(wide_col.values[:33], compact_col.values)
                snapshot = wide.state
                wide_col.values[:] = [1234.0] * 100
                self.assertEqual(wide.state, snapshot)

        invalid_params = (
            AutoCorrelationHeatmapParams(length=1),
            AutoCorrelationHeatmapParams(length=True),  # type: ignore[arg-type]
            AutoCorrelationHeatmapParams(length=2.5),  # type: ignore[arg-type]
            AutoCorrelationHeatmapParams(max_lag=-1),
            AutoCorrelationHeatmapParams(max_lag=100),
            AutoCorrelationHeatmapParams(max_lag=True),  # type: ignore[arg-type]
            AutoCorrelationHeatmapParams(smoothing_period=0.5),
            AutoCorrelationHeatmapParams(smoothing_period=math.nan),
            AutoCorrelationHeatmapParams(smoothing_period=math.inf),
            AutoCorrelationHeatmapParams(smoothing_period=True),  # type: ignore[arg-type]
        )
        for bad in invalid_params:
            with self.subTest(bad=bad), self.assertRaises(ValueError):
                AutoCorrelationHeatmap(bad)

        long_window = AutoCorrelationHeatmap(AutoCorrelationHeatmapParams(length=300))
        for price in data.INPUT_CLOSE:
            self.assertTrue(long_window.update(price).is_empty())
        self.assertFalse(long_window.is_primed())

    def test_mnemonics_metadata_and_descriptor(self) -> None:
        cases = (
            (default_params(), "ach(20, 20, 99)"),
            (
                AutoCorrelationHeatmapParams(
                    length=20, smoothing_period=10.5, max_lag=0
                ),
                "ach(20, 10.5, 0)",
            ),
            (
                AutoCorrelationHeatmapParams(bar_component=BarComponent.MEDIAN),
                "ach(20, 20, 99, hl/2)",
            ),
            (
                AutoCorrelationHeatmapParams(quote_component=QuoteComponent.BID),
                "ach(20, 20, 99, b)",
            ),
            (
                AutoCorrelationHeatmapParams(trade_component=TradeComponent.VOLUME),
                "ach(20, 20, 99, v)",
            ),
            (
                AutoCorrelationHeatmapParams(
                    bar_component=BarComponent.TYPICAL,
                    quote_component=QuoteComponent.ASK,
                ),
                "ach(20, 20, 99, hlc/3, a)",
            ),
            (
                AutoCorrelationHeatmapParams(
                    bar_component=BarComponent.OPEN,
                    quote_component=QuoteComponent.WEIGHTED,
                    trade_component=TradeComponent.VOLUME,
                ),
                "ach(20, 20, 99, o, (bbs+aas)/(bs+as), v)",
            ),
        )
        for case_params, want_mnemonic in cases:
            with self.subTest(mnemonic=want_mnemonic):
                self.assertEqual(
                    auto_correlation_heatmap_mnemonic(case_params), want_mnemonic
                )
                instance = AutoCorrelationHeatmap(case_params)
                meta = instance.metadata()
                self.assertEqual(meta.identifier, Identifier.AUTO_CORRELATION_HEATMAP)
                self.assertEqual(meta.mnemonic, want_mnemonic)
                self.assertEqual(
                    meta.description, f"Autocorrelation heatmap {want_mnemonic}"
                )
                self.assertEqual(len(meta.outputs), 1)
                self.assertEqual(
                    meta.outputs[0].kind, int(AutoCorrelationHeatmapOutput.VALUE)
                )
                self.assertEqual(meta.outputs[0].shape, Shape.HEATMAP)
                self.assertEqual(meta.outputs[0].mnemonic, want_mnemonic)

        desc = descriptor_of(Identifier.AUTO_CORRELATION_HEATMAP)
        self.assertIsNotNone(desc)
        assert desc is not None
        self.assertEqual(desc.family, "John Ehlers")
        self.assertEqual(desc.adaptivity, Adaptivity.STATIC)
        self.assertEqual(desc.input_requirement, InputRequirement.SCALAR_INPUT)
        self.assertEqual(desc.volume_usage, VolumeUsage.NO_VOLUME)
        self.assertEqual(len(desc.outputs), 1)
        self.assertEqual(desc.outputs[0].shape, Shape.HEATMAP)
        self.assertEqual(desc.outputs[0].role, Role.CORRELATION)
        self.assertEqual(desc.outputs[0].pane, Pane.OWN)

    def test_entity_updates_and_factory(self) -> None:
        scalar_ind = AutoCorrelationHeatmap()
        bar_ind = AutoCorrelationHeatmap()
        quote_ind = AutoCorrelationHeatmap()
        trade_ind = AutoCorrelationHeatmap()
        open_ind = AutoCorrelationHeatmap(
            AutoCorrelationHeatmapParams(bar_component=BarComponent.OPEN)
        )
        open_ref = AutoCorrelationHeatmap()

        base_time = datetime.datetime(2026, 1, 1)
        for bar_index in range(len(data.INPUT_CLOSE)):
            timestamp = base_time + datetime.timedelta(days=bar_index)
            close_price = data.INPUT_CLOSE[bar_index]
            open_price = data.INPUT_OPEN[bar_index]
            high_price = data.INPUT_HIGH[bar_index]
            low_price = data.INPUT_LOW[bar_index]

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

        default_created = create_indicator(Identifier.AUTO_CORRELATION_HEATMAP, {})
        self.assertEqual(default_created.metadata().mnemonic, "ach(20, 20, 99)")

        custom_created = create_indicator(
            Identifier.AUTO_CORRELATION_HEATMAP,
            {
                "length": 20,
                "smoothingPeriod": 10.5,
                "maxLag": 0,
                "barComponent": 5,
            },
        )
        self.assertEqual(custom_created.metadata().mnemonic, "ach(20, 10.5, 0, hl/2)")

        for bad_params in (
            {"unknownKey": 1},
            {"length": None},
            {"length": 1},
            {"smoothingPeriod": 0.5},
            {"maxLag": -1},
            {"maxLag": 100},
            {"smoothingPeriod": 10.0, "smoothing_period": 10.0},
        ):
            with self.assertRaises(ValueError):
                create_indicator(Identifier.AUTO_CORRELATION_HEATMAP, bad_params)


if __name__ == "__main__":
    unittest.main()
