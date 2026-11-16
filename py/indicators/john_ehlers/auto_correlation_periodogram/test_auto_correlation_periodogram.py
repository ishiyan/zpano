"""Tests for Ehlers' Autocorrelation Periodogram indicator."""

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
from py.indicators.core.component_triple_mnemonic import component_triple_mnemonic
from py.indicators.core.descriptors import descriptor_of
from py.indicators.core.identifier import Identifier
from py.indicators.core.outputs.shape import Shape
from py.indicators.factory.factory import create_indicator
from py.indicators.john_ehlers.auto_correlation_periodogram import (
    AutoCorrelationPeriodogram,
    AutoCorrelationPeriodogramOutput,
    AutoCorrelationPeriodogramParams,
    AutoCorrelationPeriodogramSource,
    PeriodogramOutput,
    auto_correlation_periodogram_dominant_cycle_mnemonic,
    auto_correlation_periodogram_mnemonic,
    default_params,
)
from py.indicators.john_ehlers.auto_correlation_periodogram import (
    test_testdata as data,
)

FIXTURE_TOLERANCE = 1e-11
ORACLE_TOLERANCE = 2e-10
EXACT_TOLERANCE = 1e-13

REFERENCE_CASES: tuple[tuple[str, AutoCorrelationPeriodogramParams], ...] = tuple(
    (
        f"V{0 if source == AutoCorrelationPeriodogramSource.ARTICLE else 1}"
        f"_L{first}_H{last}_A{averaging}_E{1 if enhanced else 0}",
        AutoCorrelationPeriodogramParams(
            source=source,
            min_period=first,
            max_period=last,
            averaging_length=averaging,
            enhance_resolution=enhanced,
        ),
    )
    for source in (
        AutoCorrelationPeriodogramSource.ARTICLE,
        AutoCorrelationPeriodogramSource.BOOK,
    )
    for first, last in ((8, 24), (8, 48), (10, 48), (16, 64))
    for averaging in (0, 3, 5, 10)
    for enhanced in (False, True)
)


def _assert_close(
    case: unittest.TestCase,
    actual: float,
    expected: float,
    label: str,
    tolerance: float = EXACT_TOLERANCE,
) -> None:
    if math.isnan(expected):
        case.assertTrue(math.isnan(actual), f"{label}: expected NaN, got {actual!r}")
        return
    case.assertFalse(math.isnan(actual), f"{label}: expected {expected!r}, got NaN")
    bound = max(tolerance, abs(expected) * tolerance)
    case.assertAlmostEqual(actual, expected, delta=bound, msg=label)


def _source_reference(
    prices: list[float], configuration: AutoCorrelationPeriodogramParams
) -> list[tuple[float, float, list[float], list[float], float, float, list[float], bool]]:
    book = configuration.source == AutoCorrelationPeriodogramSource.BOOK
    assert configuration.min_period is not None
    min_p = configuration.min_period
    max_p = configuration.max_period
    hp_history: list[float] = []
    filtered_history: list[float] = []
    correlations = [0.0] * (max_p + 1)
    powers = [0.0] * (max_p - min_p + 1)
    maximum = 0.0
    dominant = 0.0
    has_dominant = False
    records = []

    ss_angle = math.radians(math.sqrt(2.0) * 180.0 / min_p)
    decay = math.exp(-math.sqrt(2.0) * math.pi / min_p)
    c2 = 2.0 * decay * math.cos(ss_angle)
    c3 = -decay * decay
    c1 = 1.0 - c2 - c3
    article_angle = math.radians(360.0 / max_p)
    article_pole = (1.0 - math.sin(article_angle)) / math.cos(article_angle)
    book_angle = math.radians((1.0 / math.sqrt(2.0)) * 360.0 / max_p)
    alpha = (math.cos(book_angle) + math.sin(book_angle) - 1.0) / math.cos(book_angle)

    def old(series: list[float], index: int) -> float:
        return series[index] if index >= 0 else 0.0

    for step_index, price in enumerate(prices):
        if book:
            high_value = (
                (1.0 - alpha / 2.0)
                * (1.0 - alpha / 2.0)
                * (
                    price
                    - 2.0 * old(prices, step_index - 1)
                    + old(prices, step_index - 2)
                )
                + 2.0 * (1.0 - alpha) * old(hp_history, step_index - 1)
                - (1.0 - alpha) * (1.0 - alpha) * old(hp_history, step_index - 2)
            )
        else:
            high_value = 0.5 * (1.0 + article_pole) * (
                price - old(prices, step_index - 1)
            ) + article_pole * old(hp_history, step_index - 1)
        hp_history.append(high_value)

        filtered_value = (
            c1 * (high_value + old(hp_history, step_index - 1)) / 2.0
            + c2 * old(filtered_history, step_index - 1)
            + c3 * old(filtered_history, step_index - 2)
        )
        filtered_history.append(filtered_value)

        for lag in range(max_p + 1):
            window_size = configuration.averaging_length or lag
            sx = sy = sxx = syy = sxy = 0.0
            for offset in range(window_size):
                x_sample = old(filtered_history, step_index - offset)
                y_sample = old(filtered_history, step_index - lag - offset)
                sx += x_sample
                sy += y_sample
                sxx += x_sample * x_sample
                sxy += x_sample * y_sample
                syy += y_sample * y_sample
            denominator = (window_size * sxx - sx * sx) * (
                window_size * syy - sy * sy
            )
            if denominator > 0.0:
                correlations[lag] = (window_size * sxy - sx * sy) / math.sqrt(
                    denominator
                )

        for bin_index, period in enumerate(range(min_p, max_p + 1)):
            cosine_part = sine_part = 0.0
            for lag in range(3, max_p + 1):
                angle = math.radians(360.0 * lag / period)
                cosine_part += correlations[lag] * math.cos(angle)
                sine_part += correlations[lag] * math.sin(angle)
            squared_sum = cosine_part * cosine_part + sine_part * sine_part
            powers[bin_index] = 0.2 * squared_sum * squared_sum + 0.8 * powers[bin_index]

        maximum = 0.995 * maximum if book else 0.0
        for power_value in powers:
            if power_value > maximum:
                maximum = power_value
        normalized = [
            power_value / maximum if maximum > 0.0 else 0.0 for power_value in powers
        ]
        weights = (
            [norm_value**3 for norm_value in normalized]
            if not book and configuration.enhance_resolution
            else normalized
        )
        spx = sp = 0.0
        peak = max(weights)
        for bin_index, weight in enumerate(weights):
            if (book and weight >= 0.5) or (
                not book and peak >= 0.25 and weight >= 0.25
            ):
                spx += (min_p + bin_index) * weight
                sp += weight
        candidate = spx / sp if sp != 0.0 else 0.0
        if (book and sp == 0.0) or (not book and sp < 0.25):
            candidate = dominant
        elif sp != 0.0:
            has_dominant = True
        dominant = candidate
        display = (
            [norm_value**2 for norm_value in normalized]
            if book and configuration.enhance_resolution
            else weights[:]
        )
        records.append(
            (
                high_value,
                filtered_value,
                correlations[:],
                powers[:],
                maximum,
                dominant,
                display,
                has_dominant,
            )
        )
    return records


class TestAutoCorrelationPeriodogram(unittest.TestCase):
    def test_all_64_reference_fixtures(self) -> None:
        self.assertEqual(len(REFERENCE_CASES), 64)
        base_time = datetime.datetime(2020, 1, 1, tzinfo=datetime.timezone.utc)
        for suffix, params in REFERENCE_CASES:
            with self.subTest(case=suffix):
                indicator = AutoCorrelationPeriodogram(params)
                expected_dc = getattr(data, f"EXPECTED_{suffix}_DC")
                expected_heatmap = getattr(data, f"EXPECTED_{suffix}_HEATMAP")
                expected_min = getattr(data, f"EXPECTED_{suffix}_MIN")
                expected_max = getattr(data, f"EXPECTED_{suffix}_MAX")
                assert params.min_period is not None
                bins = params.max_period - params.min_period + 1

                for bar_index, sample in enumerate(data.INPUT_CLOSE):
                    bar_time = base_time + datetime.timedelta(minutes=bar_index)
                    result = indicator.update(sample, bar_time)
                    self.assertEqual(result.heatmap.time, bar_time)
                    self.assertEqual(
                        result.heatmap.parameter_first, float(params.min_period)
                    )
                    self.assertEqual(
                        result.heatmap.parameter_last, float(params.max_period)
                    )
                    self.assertEqual(result.heatmap.parameter_resolution, 1.0)

                    primed = bar_index >= indicator.warmup_bars
                    self.assertEqual(indicator.is_primed(), primed)
                    _assert_close(
                        self,
                        result.dominant_cycle,
                        expected_dc[bar_index],
                        f"{suffix}.dc[{bar_index}]",
                        FIXTURE_TOLERANCE,
                    )
                    _assert_close(
                        self,
                        result.heatmap.value_min,
                        expected_min[bar_index],
                        f"{suffix}.min[{bar_index}]",
                        FIXTURE_TOLERANCE,
                    )
                    _assert_close(
                        self,
                        result.heatmap.value_max,
                        expected_max[bar_index],
                        f"{suffix}.max[{bar_index}]",
                        FIXTURE_TOLERANCE,
                    )

                    if not primed:
                        self.assertTrue(result.heatmap.is_empty())
                        self.assertTrue(
                            all(math.isnan(cell) for cell in expected_heatmap[bar_index])
                        )
                        continue

                    self.assertEqual(len(result.heatmap.values), bins)
                    for bin_index in range(bins):
                        _assert_close(
                            self,
                            result.heatmap.values[bin_index],
                            expected_heatmap[bar_index][bin_index],
                            f"{suffix}.hm[{bar_index}][{bin_index}]",
                            FIXTURE_TOLERANCE,
                        )

    def test_independent_degree_based_source_transcription(self) -> None:
        for suffix, params in REFERENCE_CASES:
            with self.subTest(case=suffix):
                reference = _source_reference(data.INPUT_CLOSE, params)
                indicator = AutoCorrelationPeriodogram(params)
                for bar_index, sample in enumerate(data.INPUT_CLOSE):
                    result = indicator.update(sample)
                    (
                        hp_ref,
                        filt_ref,
                        corr_ref,
                        power_ref,
                        max_ref,
                        dom_ref,
                        disp_ref,
                        has_dom_ref,
                    ) = reference[bar_index]
                    state = indicator.state
                    _assert_close(
                        self, state.high_pass, hp_ref, "hp", ORACLE_TOLERANCE
                    )
                    _assert_close(
                        self, state.filtered, filt_ref, "filtered", ORACLE_TOLERANCE
                    )
                    _assert_close(
                        self,
                        state.normalization_maximum,
                        max_ref,
                        "max",
                        ORACLE_TOLERANCE,
                    )
                    _assert_close(
                        self,
                        state.internal_dominant_cycle,
                        dom_ref,
                        "dom",
                        ORACLE_TOLERANCE,
                    )
                    for lag, expected_corr in enumerate(corr_ref):
                        _assert_close(
                            self,
                            state.correlations[lag],
                            expected_corr,
                            "corr",
                            ORACLE_TOLERANCE,
                        )
                    for bin_index, expected_pow in enumerate(power_ref):
                        _assert_close(
                            self,
                            state.smoothed_power[bin_index],
                            expected_pow,
                            "power",
                            ORACLE_TOLERANCE,
                        )

                    primed = bar_index >= indicator.warmup_bars
                    if primed:
                        for bin_index, expected_disp in enumerate(disp_ref):
                            _assert_close(
                                self,
                                result.heatmap.values[bin_index],
                                expected_disp,
                                "disp",
                                ORACLE_TOLERANCE,
                            )
                    else:
                        self.assertTrue(result.heatmap.is_empty())
                    expected_dc = dom_ref if primed and has_dom_ref else math.nan
                    _assert_close(
                        self, result.dominant_cycle, expected_dc, "dc", ORACLE_TOLERANCE
                    )

    def test_enhancement_placement_and_book_hold(self) -> None:
        records: dict[tuple[str, bool], list[PeriodogramOutput]] = {}
        for source in ("article", "book"):
            for enhanced in (False, True):
                indicator = AutoCorrelationPeriodogram(
                    AutoCorrelationPeriodogramParams(
                        source=source, enhance_resolution=enhanced
                    )
                )
                records[source, enhanced] = [
                    indicator.update(sample) for sample in data.INPUT_CLOSE
                ]
        for bar_index in range(50, 252):
            plain_book = records["book", False][bar_index]
            enh_book = records["book", True][bar_index]
            self.assertEqual(plain_book.dominant_cycle, enh_book.dominant_cycle)
            self.assertEqual(
                enh_book.heatmap.values,
                [cell * cell for cell in plain_book.heatmap.values],
            )

            plain_art = records["article", False][bar_index]
            enh_art = records["article", True][bar_index]
            self.assertEqual(
                enh_art.heatmap.values,
                [cell * cell * cell for cell in plain_art.heatmap.values],
            )
            self.assertEqual(plain_art.heatmap.value_max, 1.0)

        self.assertNotEqual(
            records["article", False][-1].dominant_cycle,
            records["article", True][-1].dominant_cycle,
        )
        book_series = records["book", False]
        holds = [
            step for step in range(51, 252) if book_series[step].heatmap.value_max < 0.5
        ]
        self.assertTrue(holds)
        for step in holds:
            self.assertEqual(
                book_series[step].dominant_cycle, book_series[step - 1].dominant_cycle
            )

    def test_zero_power_and_retained_undefined_correlation(self) -> None:
        for source in ("article", "book"):
            indicator = AutoCorrelationPeriodogram(
                AutoCorrelationPeriodogramParams(source=source)
            )
            for _ in range(100):
                result = indicator.update(0.0)
            self.assertTrue(indicator.is_primed())
            self.assertTrue(math.isnan(result.dominant_cycle))
            self.assertEqual(
                result.heatmap.values, [0.0] * len(result.heatmap.values)
            )
            indicator._correlations[12] = 0.75
            indicator.update(0.0)
            self.assertEqual(indicator.state.correlations[12], 0.75)

    def test_sine_peak_and_warmup_lengths(self) -> None:
        self.assertEqual(AutoCorrelationPeriodogram().warmup_bars, 50)
        self.assertEqual(
            AutoCorrelationPeriodogram(
                AutoCorrelationPeriodogramParams(averaging_length=0)
            ).warmup_bars,
            95,
        )
        indicator = AutoCorrelationPeriodogram()
        for step_index in range(1000):
            result = indicator.update(
                100.0 + math.sin(2.0 * math.pi * step_index / 20.0)
            )
        peak = result.heatmap.parameter_first + result.heatmap.values.index(
            max(result.heatmap.values)
        )
        self.assertEqual(peak, 20.0)
        self.assertTrue(18.0 < result.dominant_cycle < 22.0)

    def test_skipped_nan_overflow_reset_and_immutability(self) -> None:
        indicator = AutoCorrelationPeriodogram()
        first_run = [indicator.update(sample) for sample in data.INPUT_CLOSE]
        saved_state = indicator.state
        saved_output = indicator.output
        saved_count = indicator.samples_seen

        skipped = indicator.update(math.nan)
        self.assertTrue(skipped.heatmap.is_empty())
        self.assertTrue(math.isnan(skipped.dominant_cycle))
        self.assertIs(indicator.state, saved_state)
        self.assertIs(indicator.output, saved_output)
        self.assertEqual(indicator.samples_seen, saved_count)

        for bad_sample in (math.inf, -math.inf, 1e308, True, "bad"):
            with self.assertRaises(ValueError):
                indicator.update(bad_sample)  # type: ignore[arg-type]
            self.assertIs(indicator.state, saved_state)
            self.assertEqual(indicator.samples_seen, saved_count)

        indicator.reset()
        second_run = [indicator.update(sample) for sample in data.INPUT_CLOSE]
        for first_item, second_item in zip(first_run, second_run):
            self.assertEqual(first_item.heatmap.values, second_item.heatmap.values)
            _assert_close(
                self, first_item.dominant_cycle, second_item.dominant_cycle, "replay.dc"
            )

        snapshot = second_run[-1].heatmap.values[:]
        second_run[-1].heatmap.values[0] = -999.0
        self.assertEqual(indicator.output.heatmap.values, snapshot)

    def test_metadata_mnemonics_entities_and_factory(self) -> None:
        descriptor = descriptor_of(Identifier.AUTO_CORRELATION_PERIODOGRAM)
        self.assertIsNotNone(descriptor)
        assert descriptor is not None
        self.assertEqual(descriptor.adaptivity, Adaptivity.ADAPTIVE)
        self.assertEqual(len(descriptor.outputs), 2)

        default_ind = AutoCorrelationPeriodogram.create_default()
        meta = default_ind.metadata()
        self.assertEqual(meta.identifier, Identifier.AUTO_CORRELATION_PERIODOGRAM)
        self.assertEqual(meta.mnemonic, "acp(8, 48, 3)")
        self.assertEqual(len(meta.outputs), 2)
        self.assertEqual(meta.outputs[0].kind, AutoCorrelationPeriodogramOutput.VALUE)
        self.assertEqual(meta.outputs[0].shape, Shape.HEATMAP)
        self.assertEqual(meta.outputs[0].mnemonic, "acp(8, 48, 3)")
        self.assertEqual(
            meta.outputs[1].kind, AutoCorrelationPeriodogramOutput.DOMINANT_CYCLE
        )
        self.assertEqual(meta.outputs[1].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[1].mnemonic, "acp-dc(8, 48, 3)")

        book_enh_params = AutoCorrelationPeriodogramParams(
            source=AutoCorrelationPeriodogramSource.BOOK, enhance_resolution=True
        )
        book_enh = AutoCorrelationPeriodogram.create(book_enh_params)
        self.assertEqual(book_enh.min_period, 10)
        self.assertEqual(book_enh.metadata().mnemonic, "acp(10, 48, 3, book, enh)")
        self.assertEqual(
            book_enh.metadata().outputs[1].mnemonic, "acp-dc(10, 48, 3, book)"
        )

        art_enh_params = AutoCorrelationPeriodogramParams(enhance_resolution=True)
        self.assertEqual(
            auto_correlation_periodogram_mnemonic(art_enh_params),
            "acp(8, 48, 3, enh)",
        )
        self.assertEqual(
            auto_correlation_periodogram_dominant_cycle_mnemonic(art_enh_params),
            "acp-dc(8, 48, 3, enh)",
        )

        combinations = (
            (None, None, None),
            (BarComponent.OPEN, None, None),
            (None, QuoteComponent.BID, None),
            (None, None, TradeComponent.VOLUME),
            (BarComponent.MEDIAN, QuoteComponent.ASK, None),
            (BarComponent.MEDIAN, None, TradeComponent.VOLUME),
            (None, QuoteComponent.ASK, TradeComponent.VOLUME),
            (BarComponent.OPEN, QuoteComponent.BID, TradeComponent.VOLUME),
        )
        for bar_comp, quote_comp, trade_comp in combinations:
            params = default_params()
            params.bar_component = bar_comp
            params.quote_component = quote_comp
            params.trade_component = trade_comp
            instance = AutoCorrelationPeriodogram(params)
            suffix = component_triple_mnemonic(
                bar_comp if bar_comp is not None else DEFAULT_BAR_COMPONENT,
                quote_comp if quote_comp is not None else DEFAULT_QUOTE_COMPONENT,
                trade_comp if trade_comp is not None else DEFAULT_TRADE_COMPONENT,
            )
            self.assertEqual(instance.metadata().mnemonic, f"acp(8, 48, 3{suffix})")
            self.assertEqual(
                instance.metadata().outputs[1].mnemonic, f"acp-dc(8, 48, 3{suffix})"
            )

        base_time = datetime.datetime(2020, 1, 1, tzinfo=datetime.timezone.utc)
        bar_ind = AutoCorrelationPeriodogram()
        scalar_ind = AutoCorrelationPeriodogram()
        quote_ind = AutoCorrelationPeriodogram()
        trade_ind = AutoCorrelationPeriodogram()
        for step_index in range(60):
            bar_time = base_time + datetime.timedelta(days=step_index)
            price = data.INPUT_CLOSE[step_index]
            out_bar = bar_ind.update_bar(
                Bar(
                    bar_time,
                    data.INPUT_OPEN[step_index],
                    data.INPUT_HIGH[step_index],
                    data.INPUT_LOW[step_index],
                    price,
                    1000.0,
                )
            )
            out_scalar = scalar_ind.update_scalar(Scalar(bar_time, price))
            out_quote = quote_ind.update_quote(
                Quote(bar_time, price - 0.5, price + 0.5, 10.0, 10.0)
            )
            out_trade = trade_ind.update_trade(Trade(bar_time, price, 10.0))
            self.assertEqual(len(out_bar), 2)
            _assert_close(
                self,
                out_bar[1].value,
                out_scalar[1].value,
                f"bar_vs_scalar[{step_index}]",
            )
            _assert_close(
                self,
                out_bar[1].value,
                out_quote[1].value,
                f"bar_vs_quote[{step_index}]",
            )
            _assert_close(
                self,
                out_bar[1].value,
                out_trade[1].value,
                f"bar_vs_trade[{step_index}]",
            )
            self.assertEqual(out_bar[0].values, out_scalar[0].values)

        factory_default = create_indicator(Identifier.AUTO_CORRELATION_PERIODOGRAM)
        self.assertEqual(factory_default.metadata().mnemonic, "acp(8, 48, 3)")
        factory_book = create_indicator(
            Identifier.AUTO_CORRELATION_PERIODOGRAM,
            {
                "source": "book",
                "enhanceResolution": True,
                "barComponent": BarComponent.MEDIAN,
            },
        )
        self.assertEqual(
            factory_book.metadata().mnemonic, "acp(10, 48, 3, book, enh, hl/2)"
        )
        self.assertEqual(
            factory_book.metadata().outputs[1].mnemonic,
            "acp-dc(10, 48, 3, book, hl/2)",
        )

    def test_invalid_parameters_and_factory_rejection(self) -> None:
        invalid_params = (
            AutoCorrelationPeriodogramParams(source="literal370"),
            AutoCorrelationPeriodogramParams(source=1),  # type: ignore[arg-type]
            AutoCorrelationPeriodogramParams(source=True),  # type: ignore[arg-type]
            AutoCorrelationPeriodogramParams(min_period=0),
            AutoCorrelationPeriodogramParams(min_period=2),
            AutoCorrelationPeriodogramParams(min_period=True),  # type: ignore[arg-type]
            AutoCorrelationPeriodogramParams(min_period=8.0),  # type: ignore[arg-type]
            AutoCorrelationPeriodogramParams(max_period=4),
            AutoCorrelationPeriodogramParams(max_period=8),
            AutoCorrelationPeriodogramParams(
                source="book", max_period=10
            ),
            AutoCorrelationPeriodogramParams(max_period=True),  # type: ignore[arg-type]
            AutoCorrelationPeriodogramParams(averaging_length=1),
            AutoCorrelationPeriodogramParams(averaging_length=-1),
            AutoCorrelationPeriodogramParams(averaging_length=False),  # type: ignore[arg-type]
            AutoCorrelationPeriodogramParams(enhance_resolution=1),  # type: ignore[arg-type]
            AutoCorrelationPeriodogramParams(bar_component=True),  # type: ignore[arg-type]
        )
        for bad_params in invalid_params:
            with self.subTest(bad_params=bad_params), self.assertRaises(ValueError):
                AutoCorrelationPeriodogram(bad_params)

        invalid_dicts = (
            {"minPeriod": 0},
            {"minPeriod": 2},
            {"minPeriod": 8.0},
            {"minPeriod": None},
            {"minPeriod": 10, "min_period": 10},
            {"maxPeriod": 8},
            {"averagingLength": 1},
            {"averagingLength": -1},
            {"source": "literal370"},
            {"source": 0},
            {"enhanceResolution": 1},
            {"disableAutomaticGainControl": True},
            {"automaticGainControlDecayFactor": 0.995},
            {"fixedNormalization": True},
            {"unknownField": 1},
        )
        for bad_dict in invalid_dicts:
            with self.subTest(bad_dict=bad_dict), self.assertRaises(ValueError):
                create_indicator(Identifier.AUTO_CORRELATION_PERIODOGRAM, bad_dict)


if __name__ == "__main__":
    unittest.main()
