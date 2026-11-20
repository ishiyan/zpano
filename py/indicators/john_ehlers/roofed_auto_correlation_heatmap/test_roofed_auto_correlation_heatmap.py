"""Tests for Ehlers' 2013 Roofed Autocorrelation Heatmap indicator."""

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
from py.indicators.john_ehlers.roofed_auto_correlation_heatmap import (
    AutocorrelationState,
    RoofedAutoCorrelationHeatmap,
    RoofedAutoCorrelationHeatmapOutput,
    RoofedAutoCorrelationHeatmapParams,
    default_params,
    roofed_auto_correlation_heatmap_mnemonic,
    test_testdata as data,
)

FIXTURE_TOLERANCE = 1e-12
ORACLE_TOLERANCE = 1e-13

REFERENCE_CASES: tuple[
    tuple[
        str,
        RoofedAutoCorrelationHeatmapParams,
        list[list[float]],
        list[float],
        list[float],
    ],
    ...,
] = (
    (
        "L03_L148_A0_HP48_LP10",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=3,
            max_lag=48,
            averaging_length=0,
            highpass_period=48.0,
            smoothing_period=10.0,
        ),
        data.EXPECTED_L03_L148_A0_HP48_LP10_HEATMAP,
        data.EXPECTED_L03_L148_A0_HP48_LP10_MIN,
        data.EXPECTED_L03_L148_A0_HP48_LP10_MAX,
    ),
    (
        "L03_L148_A2_HP48_LP10",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=3,
            max_lag=48,
            averaging_length=2,
            highpass_period=48.0,
            smoothing_period=10.0,
        ),
        data.EXPECTED_L03_L148_A2_HP48_LP10_HEATMAP,
        data.EXPECTED_L03_L148_A2_HP48_LP10_MIN,
        data.EXPECTED_L03_L148_A2_HP48_LP10_MAX,
    ),
    (
        "L03_L148_A3_HP48_LP10",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=3,
            max_lag=48,
            averaging_length=3,
            highpass_period=48.0,
            smoothing_period=10.0,
        ),
        data.EXPECTED_L03_L148_A3_HP48_LP10_HEATMAP,
        data.EXPECTED_L03_L148_A3_HP48_LP10_MIN,
        data.EXPECTED_L03_L148_A3_HP48_LP10_MAX,
    ),
    (
        "L03_L148_A20_HP48_LP10",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=3,
            max_lag=48,
            averaging_length=20,
            highpass_period=48.0,
            smoothing_period=10.0,
        ),
        data.EXPECTED_L03_L148_A20_HP48_LP10_HEATMAP,
        data.EXPECTED_L03_L148_A20_HP48_LP10_MIN,
        data.EXPECTED_L03_L148_A20_HP48_LP10_MAX,
    ),
    (
        "L03_L148_A100_HP48_LP10",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=3,
            max_lag=48,
            averaging_length=100,
            highpass_period=48.0,
            smoothing_period=10.0,
        ),
        data.EXPECTED_L03_L148_A100_HP48_LP10_HEATMAP,
        data.EXPECTED_L03_L148_A100_HP48_LP10_MIN,
        data.EXPECTED_L03_L148_A100_HP48_LP10_MAX,
    ),
    (
        "L03_L148_A1_HP48_LP10",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=3,
            max_lag=48,
            averaging_length=1,
            highpass_period=48.0,
            smoothing_period=10.0,
        ),
        data.EXPECTED_L03_L148_A1_HP48_LP10_HEATMAP,
        data.EXPECTED_L03_L148_A1_HP48_LP10_MIN,
        data.EXPECTED_L03_L148_A1_HP48_LP10_MAX,
    ),
    (
        "L03_L148_A0_HP48_LP10_5",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=3,
            max_lag=48,
            averaging_length=0,
            highpass_period=48.0,
            smoothing_period=10.5,
        ),
        data.EXPECTED_L03_L148_A0_HP48_LP10_5_HEATMAP,
        data.EXPECTED_L03_L148_A0_HP48_LP10_5_MIN,
        data.EXPECTED_L03_L148_A0_HP48_LP10_5_MAX,
    ),
    (
        "L03_L148_A0_HP48_LP20",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=3,
            max_lag=48,
            averaging_length=0,
            highpass_period=48.0,
            smoothing_period=20.0,
        ),
        data.EXPECTED_L03_L148_A0_HP48_LP20_HEATMAP,
        data.EXPECTED_L03_L148_A0_HP48_LP20_MIN,
        data.EXPECTED_L03_L148_A0_HP48_LP20_MAX,
    ),
    (
        "L03_L148_A0_HP80_LP10",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=3,
            max_lag=48,
            averaging_length=0,
            highpass_period=80.0,
            smoothing_period=10.0,
        ),
        data.EXPECTED_L03_L148_A0_HP80_LP10_HEATMAP,
        data.EXPECTED_L03_L148_A0_HP80_LP10_MIN,
        data.EXPECTED_L03_L148_A0_HP80_LP10_MAX,
    ),
    (
        "L03_L124_A0_HP24_LP10",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=3,
            max_lag=24,
            averaging_length=0,
            highpass_period=24.0,
            smoothing_period=10.0,
        ),
        data.EXPECTED_L03_L124_A0_HP24_LP10_HEATMAP,
        data.EXPECTED_L03_L124_A0_HP24_LP10_MIN,
        data.EXPECTED_L03_L124_A0_HP24_LP10_MAX,
    ),
    (
        "L012_L112_A12_HP48_LP10",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=12,
            max_lag=12,
            averaging_length=12,
            highpass_period=48.0,
            smoothing_period=10.0,
        ),
        data.EXPECTED_L012_L112_A12_HP48_LP10_HEATMAP,
        data.EXPECTED_L012_L112_A12_HP48_LP10_MIN,
        data.EXPECTED_L012_L112_A12_HP48_LP10_MAX,
    ),
    (
        "L00_L10_A3_HP48_LP10",
        RoofedAutoCorrelationHeatmapParams(
            min_lag=0,
            max_lag=0,
            averaging_length=3,
            highpass_period=48.0,
            smoothing_period=10.0,
        ),
        data.EXPECTED_L00_L10_A3_HP48_LP10_HEATMAP,
        data.EXPECTED_L00_L10_A3_HP48_LP10_MIN,
        data.EXPECTED_L00_L10_A3_HP48_LP10_MAX,
    ),
)


def _listing_oracle(
    closes: list[float],
    params: RoofedAutoCorrelationHeatmapParams,
) -> list[tuple[float, float, tuple[float, ...], list[float]]]:
    """Direct full-series transcription of Book Listing 8-2 with zero prehistory."""
    prices: list[float] = []
    highpass_series: list[float] = []
    filtered_series: list[float] = []
    correlations = [0.0] * (params.max_lag + 1)

    def previous(series: list[float], lag: int) -> float:
        return series[-1 - lag] if len(series) > lag else 0.0

    theta = math.sqrt(2.0) * math.pi / params.highpass_period
    alpha = (math.cos(theta) + math.sin(theta) - 1.0) / math.cos(theta)
    attenuation = math.exp(-math.sqrt(2.0) * math.pi / params.smoothing_period)
    coeff_c2 = 2.0 * attenuation * math.cos(
        math.sqrt(2.0) * math.pi / params.smoothing_period
    )
    coeff_c3 = -attenuation * attenuation
    coeff_c1 = 1.0 - coeff_c2 - coeff_c3

    snapshots: list[tuple[float, float, tuple[float, ...], list[float]]] = []
    for price in closes:
        high = (
            (1.0 - alpha / 2.0)
            * (1.0 - alpha / 2.0)
            * (price - 2.0 * previous(prices, 0) + previous(prices, 1))
            + 2.0 * (1.0 - alpha) * previous(highpass_series, 0)
            - (1.0 - alpha) * (1.0 - alpha) * previous(highpass_series, 1)
        )
        filtered = (
            coeff_c1 * (high + previous(highpass_series, 0)) / 2.0
            + coeff_c2 * previous(filtered_series, 0)
            + coeff_c3 * previous(filtered_series, 1)
        )
        prices.append(price)
        highpass_series.append(high)
        filtered_series.append(filtered)

        for lag in range(params.max_lag + 1):
            window = params.averaging_length if params.averaging_length else lag
            sum_x = 0.0
            sum_y = 0.0
            sum_xx = 0.0
            sum_xy = 0.0
            sum_yy = 0.0
            for count in range(window):
                sample_x = previous(filtered_series, count)
                sample_y = previous(filtered_series, lag + count)
                sum_x += sample_x
                sum_y += sample_y
                sum_xx += sample_x * sample_x
                sum_xy += sample_x * sample_y
                sum_yy += sample_y * sample_y
            denominator = (window * sum_xx - sum_x * sum_x) * (
                window * sum_yy - sum_y * sum_y
            )
            if denominator > 0.0:
                correlations[lag] = (window * sum_xy - sum_x * sum_y) / math.sqrt(
                    denominator
                )
            correlations[lag] = 0.5 * (correlations[lag] + 1.0)

        snapshots.append(
            (
                high,
                filtered,
                tuple(correlations),
                correlations[params.min_lag : params.max_lag + 1],
            )
        )
    return snapshots


class TestRoofedAutoCorrelationHeatmap(unittest.TestCase):
    """Unit tests for RoofedAutoCorrelationHeatmap."""

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

    def test_reference_fixtures_and_oracle(self) -> None:
        closes = data.INPUT_CLOSE
        self.assertEqual(len(closes), 252)

        for (
            case_name,
            case_params,
            expected_heatmap,
            expected_min,
            expected_max,
        ) in REFERENCE_CASES:
            with self.subTest(case=case_name):
                indicator = RoofedAutoCorrelationHeatmap(case_params)
                oracle_rows = _listing_oracle(closes, case_params)
                expected_bins = case_params.max_lag - case_params.min_lag + 1

                for bar_index, close_price in enumerate(closes):
                    timestamp = datetime.datetime(
                        2024, 1, 1, tzinfo=datetime.timezone.utc
                    ) + datetime.timedelta(days=bar_index)
                    column = indicator.update(close_price, timestamp)
                    state = indicator.state
                    oracle_high, oracle_filtered, oracle_corr, oracle_displayed = (
                        oracle_rows[bar_index]
                    )

                    self.assertAlmostEqual(
                        state.high_pass, oracle_high, delta=ORACLE_TOLERANCE
                    )
                    self.assertAlmostEqual(
                        state.filtered, oracle_filtered, delta=ORACLE_TOLERANCE
                    )
                    self.assertEqual(len(state.correlations), len(oracle_corr))
                    for actual_corr, want_corr in zip(state.correlations, oracle_corr):
                        self.assertAlmostEqual(
                            actual_corr, want_corr, delta=ORACLE_TOLERANCE
                        )

                    self.assertEqual(column.time, timestamp)
                    self.assertEqual(column.parameter_first, float(case_params.min_lag))
                    self.assertEqual(column.parameter_last, float(case_params.max_lag))
                    self.assertEqual(column.parameter_resolution, 1.0)

                    row = expected_heatmap[bar_index]
                    self.assertEqual(len(row), expected_bins)
                    if bar_index < indicator.warmup_bars:
                        self.assertFalse(indicator.is_primed())
                        self.assertTrue(column.is_empty())
                        self.assertEqual(column.values, [])
                        self.assertTrue(math.isnan(column.value_min))
                        self.assertTrue(math.isnan(column.value_max))
                        self.assertTrue(all(math.isnan(value) for value in row))
                        self.assertTrue(math.isnan(expected_min[bar_index]))
                        self.assertTrue(math.isnan(expected_max[bar_index]))
                    else:
                        self.assertTrue(indicator.is_primed())
                        self.assertFalse(column.is_empty())
                        self.assertEqual(len(column.values), expected_bins)
                        self._assert_close_or_nan(
                            column.value_min,
                            expected_min[bar_index],
                            FIXTURE_TOLERANCE,
                        )
                        self._assert_close_or_nan(
                            column.value_max,
                            expected_max[bar_index],
                            FIXTURE_TOLERANCE,
                        )
                        for actual_value, expected_value, oracle_value in zip(
                            column.values, row, oracle_displayed
                        ):
                            self._assert_close_or_nan(
                                actual_value, expected_value, FIXTURE_TOLERANCE
                            )
                            self.assertAlmostEqual(
                                actual_value, oracle_value, delta=ORACLE_TOLERANCE
                            )

    def test_lag_zero_seed_and_display_slice_invariance(self) -> None:
        seed_only = RoofedAutoCorrelationHeatmap(
            RoofedAutoCorrelationHeatmapParams(
                min_lag=0, max_lag=0, averaging_length=0
            )
        )
        self.assertEqual(seed_only.warmup_bars, 0)
        self.assertEqual(seed_only.update(100.0).values, [0.5])
        self.assertEqual(seed_only.update(100.0).values, [0.75])

        lag_zero_self = RoofedAutoCorrelationHeatmap(
            RoofedAutoCorrelationHeatmapParams(
                min_lag=0, max_lag=0, averaging_length=3
            )
        )
        self.assertEqual(lag_zero_self.warmup_bars, 2)
        outputs = [lag_zero_self.update(price) for price in data.INPUT_CLOSE]
        self.assertEqual(len(outputs[-1].values), 1)
        self.assertAlmostEqual(outputs[-1].values[0], 1.0, delta=ORACLE_TOLERANCE)

        full = RoofedAutoCorrelationHeatmap()
        compact = RoofedAutoCorrelationHeatmap(
            RoofedAutoCorrelationHeatmapParams(min_lag=10, max_lag=48)
        )
        for price in data.INPUT_CLOSE:
            full.update(price)
            compact.update(price)
        self.assertEqual(full.state, compact.state)
        self.assertEqual(full.output.values[7:], compact.output.values)

    def test_sine_wave_cosine_lag_response(self) -> None:
        indicator = RoofedAutoCorrelationHeatmap(
            RoofedAutoCorrelationHeatmapParams(
                min_lag=0, max_lag=40, averaging_length=20
            )
        )
        result = Heatmap.empty(None, 0.0, 40.0, 1.0)
        for step in range(1000):
            result = indicator.update(math.sin(2.0 * math.pi * step / 20.0))
        for lag, actual in enumerate(result.values):
            expected = 0.5 * (1.0 + math.cos(2.0 * math.pi * lag / 20.0))
            self.assertAlmostEqual(actual, expected, delta=1e-11)

    def test_nan_omission_overflow_reset_and_snapshot_ownership(self) -> None:
        clean = RoofedAutoCorrelationHeatmap()
        dirty = RoofedAutoCorrelationHeatmap()
        for bar_index, price in enumerate(data.INPUT_CLOSE):
            if bar_index in (0, 17, 95, 140):
                before_state = dirty.state
                before_seen = dirty.samples_seen
                omitted = dirty.update(math.nan)
                self.assertTrue(omitted.is_empty())
                self.assertEqual(dirty.state, before_state)
                self.assertEqual(dirty.samples_seen, before_seen)
            out_clean = clean.update(price)
            out_dirty = dirty.update(price)
            self.assertEqual(out_clean.values, out_dirty.values)
            self.assertEqual(clean.state, dirty.state)

        snapshot = clean.output
        snapshot.values[0] = -999.0
        self.assertNotEqual(clean.output.values[0], -999.0)

        before_overflow = clean.state
        before_count = clean.samples_seen
        for bad_sample in (math.inf, -math.inf, True, 1e308):
            with self.assertRaises(ValueError):
                clean.update(bad_sample)  # type: ignore[arg-type]
            self.assertEqual(clean.state, before_overflow)
            self.assertEqual(clean.samples_seen, before_count)

        first_pass = [val for val in clean.output.values]
        clean.reset()
        self.assertFalse(clean.is_primed())
        self.assertEqual(clean.samples_seen, 0)
        self.assertTrue(clean.output.is_empty())
        for price in data.INPUT_CLOSE:
            clean.update(price)
        self.assertEqual(clean.output.values, first_pass)

    def test_invalid_parameters(self) -> None:
        invalid_cases = (
            RoofedAutoCorrelationHeatmapParams(min_lag=-1),
            RoofedAutoCorrelationHeatmapParams(min_lag=5, max_lag=4),
            RoofedAutoCorrelationHeatmapParams(max_lag=True),  # type: ignore[arg-type]
            RoofedAutoCorrelationHeatmapParams(averaging_length=-1),
            RoofedAutoCorrelationHeatmapParams(highpass_period=4.0),
            RoofedAutoCorrelationHeatmapParams(smoothing_period=1.5),
            RoofedAutoCorrelationHeatmapParams(highpass_period=math.nan),
            RoofedAutoCorrelationHeatmapParams(smoothing_period=math.inf),
            RoofedAutoCorrelationHeatmapParams(highpass_period=True),  # type: ignore[arg-type]
        )
        for bad in invalid_cases:
            with self.subTest(bad=bad), self.assertRaises(ValueError):
                RoofedAutoCorrelationHeatmap(bad)

    def test_mnemonics_metadata_and_descriptor(self) -> None:
        self.assertEqual(default_params(), RoofedAutoCorrelationHeatmapParams())
        mnemonic_cases = (
            (None, "rach(3, 48, 0, 48, 10)"),
            (RoofedAutoCorrelationHeatmapParams(), "rach(3, 48, 0, 48, 10)"),
            (
                RoofedAutoCorrelationHeatmapParams(
                    min_lag=0,
                    max_lag=24,
                    averaging_length=3,
                    highpass_period=24.0,
                    smoothing_period=10.5,
                ),
                "rach(0, 24, 3, 24, 10.5)",
            ),
            (
                RoofedAutoCorrelationHeatmapParams(bar_component=BarComponent.MEDIAN),
                "rach(3, 48, 0, 48, 10, hl/2)",
            ),
            (
                RoofedAutoCorrelationHeatmapParams(quote_component=QuoteComponent.BID),
                "rach(3, 48, 0, 48, 10, b)",
            ),
            (
                RoofedAutoCorrelationHeatmapParams(
                    trade_component=TradeComponent.VOLUME
                ),
                "rach(3, 48, 0, 48, 10, v)",
            ),
            (
                RoofedAutoCorrelationHeatmapParams(
                    bar_component=BarComponent.TYPICAL,
                    quote_component=QuoteComponent.ASK,
                ),
                "rach(3, 48, 0, 48, 10, hlc/3, a)",
            ),
            (
                RoofedAutoCorrelationHeatmapParams(
                    bar_component=BarComponent.OPEN,
                    quote_component=QuoteComponent.WEIGHTED,
                    trade_component=TradeComponent.VOLUME,
                ),
                "rach(3, 48, 0, 48, 10, o, (bbs+aas)/(bs+as), v)",
            ),
        )
        for case_params, want_mnemonic in mnemonic_cases:
            with self.subTest(want=want_mnemonic):
                self.assertEqual(
                    roofed_auto_correlation_heatmap_mnemonic(case_params),
                    want_mnemonic,
                )
                instance = RoofedAutoCorrelationHeatmap(case_params)
                meta = instance.metadata()
                self.assertEqual(
                    meta.identifier, Identifier.ROOFED_AUTO_CORRELATION_HEATMAP
                )
                self.assertEqual(meta.mnemonic, want_mnemonic)
                self.assertEqual(
                    meta.description, f"Roofed autocorrelation heatmap {want_mnemonic}"
                )
                self.assertEqual(len(meta.outputs), 1)
                self.assertEqual(
                    meta.outputs[0].kind, int(RoofedAutoCorrelationHeatmapOutput.VALUE)
                )
                self.assertEqual(meta.outputs[0].shape, Shape.HEATMAP)
                self.assertEqual(meta.outputs[0].mnemonic, want_mnemonic)

        desc = descriptor_of(Identifier.ROOFED_AUTO_CORRELATION_HEATMAP)
        self.assertEqual(desc.family, "John Ehlers")
        self.assertEqual(desc.adaptivity, Adaptivity.STATIC)
        self.assertEqual(desc.input_requirement, InputRequirement.SCALAR_INPUT)
        self.assertEqual(desc.volume_usage, VolumeUsage.NO_VOLUME)
        self.assertEqual(len(desc.outputs), 1)
        self.assertEqual(desc.outputs[0].shape, Shape.HEATMAP)
        self.assertEqual(desc.outputs[0].role, Role.CORRELATION)
        self.assertEqual(desc.outputs[0].pane, Pane.OWN)

    def test_entity_routing_and_factory(self) -> None:
        scalar_ind = RoofedAutoCorrelationHeatmap()
        bar_ind = RoofedAutoCorrelationHeatmap()
        quote_ind = RoofedAutoCorrelationHeatmap()
        trade_ind = RoofedAutoCorrelationHeatmap()
        open_ind = RoofedAutoCorrelationHeatmap(
            RoofedAutoCorrelationHeatmapParams(bar_component=BarComponent.OPEN)
        )
        open_ref = RoofedAutoCorrelationHeatmap()

        for index, close_price in enumerate(data.INPUT_CLOSE):
            timestamp = datetime.datetime(
                2024, 1, 1, tzinfo=datetime.timezone.utc
            ) + datetime.timedelta(days=index)
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
            Identifier.ROOFED_AUTO_CORRELATION_HEATMAP, {}
        )
        self.assertEqual(
            default_created.metadata().mnemonic, "rach(3, 48, 0, 48, 10)"
        )

        custom_created = create_indicator(
            Identifier.ROOFED_AUTO_CORRELATION_HEATMAP,
            {
                "minLag": 0,
                "maxLag": 24,
                "averagingLength": 3,
                "highpassPeriod": 24.0,
                "smoothingPeriod": 10.5,
                "barComponent": 5,
            },
        )
        self.assertEqual(
            custom_created.metadata().mnemonic, "rach(0, 24, 3, 24, 10.5, hl/2)"
        )

        for bad_params in (
            {"unknown": 1},
            {"minLag": None},
            {"minLag": -1},
            {"minLag": 10, "maxLag": 5},
            {"averagingLength": -1},
            {"highpassPeriod": 4.0},
            {"smoothingPeriod": 1.0},
            {"smoothingPeriod": 10.0, "smoothing_period": 10.0},
        ):
            with self.assertRaises(ValueError):
                create_indicator(
                    Identifier.ROOFED_AUTO_CORRELATION_HEATMAP, bad_params
                )


if __name__ == "__main__":
    unittest.main()
