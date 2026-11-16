"""Tests for Ehlers' Discrete Fourier Transform Spectrum indicator."""

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
from py.indicators.john_ehlers.discrete_fourier_transform_spectrum import (
    DiscreteFourierTransformSpectrum,
    DiscreteFourierTransformSpectrumOutput,
    DiscreteFourierTransformSpectrumParams,
    SpectrumOutput,
    default_params,
    discrete_fourier_transform_spectrum_dominant_cycle_mnemonic,
    discrete_fourier_transform_spectrum_mnemonic,
)
from py.indicators.john_ehlers.discrete_fourier_transform_spectrum import (
    test_testdata as data,
)

TOLERANCE = 1e-13
ORACLE_TOLERANCE = 2e-11

REFERENCE_CASES: tuple[tuple[str, DiscreteFourierTransformSpectrumParams], ...] = (
    (
        "P010_P148_HP48_LP10_SDC1_E0_W48",
        DiscreteFourierTransformSpectrumParams(
            min_period=10,
            max_period=48,
            window_length=48,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P148_HP48_LP10_SDC0_E0_W48",
        DiscreteFourierTransformSpectrumParams(
            min_period=10,
            max_period=48,
            window_length=48,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=False,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P148_HP48_LP10_SDC1_E1_W48",
        DiscreteFourierTransformSpectrumParams(
            min_period=10,
            max_period=48,
            window_length=48,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=True,
            enhance_resolution=True,
        ),
    ),
    (
        "P010_P148_HP48_LP10_SDC0_E1_W48",
        DiscreteFourierTransformSpectrumParams(
            min_period=10,
            max_period=48,
            window_length=48,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=False,
            enhance_resolution=True,
        ),
    ),
    (
        "P010_P148_HP48_LP10_SDC1_E0_W24",
        DiscreteFourierTransformSpectrumParams(
            min_period=10,
            max_period=48,
            window_length=24,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P148_HP48_LP10_SDC1_E0_W96",
        DiscreteFourierTransformSpectrumParams(
            min_period=10,
            max_period=48,
            window_length=96,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P148_HP48_LP10_SDC1_E0_W252",
        DiscreteFourierTransformSpectrumParams(
            min_period=10,
            max_period=48,
            window_length=252,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P148_HP48_LP10_5_SDC1_E0_W48",
        DiscreteFourierTransformSpectrumParams(
            min_period=10,
            max_period=48,
            window_length=48,
            highpass_period=48.0,
            smoothing_period=10.5,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P148_HP80_LP20_SDC1_E0_W48",
        DiscreteFourierTransformSpectrumParams(
            min_period=10,
            max_period=48,
            window_length=48,
            highpass_period=80.0,
            smoothing_period=20.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P06_P124_HP24_LP6_SDC1_E0_W24",
        DiscreteFourierTransformSpectrumParams(
            min_period=6,
            max_period=24,
            window_length=24,
            highpass_period=24.0,
            smoothing_period=6.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P020_P120_HP48_LP10_SDC1_E0_W48",
        DiscreteFourierTransformSpectrumParams(
            min_period=20,
            max_period=20,
            window_length=48,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P024_P160_HP60_LP12_SDC1_E0_W60",
        DiscreteFourierTransformSpectrumParams(
            min_period=24,
            max_period=60,
            window_length=60,
            highpass_period=60.0,
            smoothing_period=12.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
)


def _assert_close(
    testcase: unittest.TestCase,
    actual: float,
    expected: float,
    label: str,
    tolerance: float = TOLERANCE,
) -> None:
    if math.isnan(expected):
        testcase.assertTrue(math.isnan(actual), f"{label}: expected NaN, got {actual}")
    else:
        testcase.assertFalse(math.isnan(actual), f"{label}: expected {expected}, got NaN")
        bound = max(tolerance, abs(expected) * tolerance)
        testcase.assertAlmostEqual(actual, expected, delta=bound, msg=label)


def _assert_same_output(
    testcase: unittest.TestCase,
    actual: SpectrumOutput,
    expected: SpectrumOutput,
    label: str,
) -> None:
    testcase.assertEqual(actual.heatmap.time, expected.heatmap.time, f"{label}.time")
    testcase.assertEqual(
        actual.heatmap.parameter_first, expected.heatmap.parameter_first, f"{label}.first"
    )
    testcase.assertEqual(
        actual.heatmap.parameter_last, expected.heatmap.parameter_last, f"{label}.last"
    )
    testcase.assertEqual(
        actual.heatmap.parameter_resolution,
        expected.heatmap.parameter_resolution,
        f"{label}.resolution",
    )
    _assert_close(testcase, actual.heatmap.value_min, expected.heatmap.value_min, f"{label}.min")
    _assert_close(testcase, actual.heatmap.value_max, expected.heatmap.value_max, f"{label}.max")
    testcase.assertEqual(len(actual.heatmap.values), len(expected.heatmap.values), f"{label}.len")
    for index, (actual_value, expected_value) in enumerate(
        zip(actual.heatmap.values, expected.heatmap.values)
    ):
        _assert_close(testcase, actual_value, expected_value, f"{label}.values[{index}]")
    _assert_close(testcase, actual.dominant_cycle, expected.dominant_cycle, f"{label}.dc")


class _ListingOracle:
    """Independent direct transcription of Book Listing 9-1."""

    def __init__(self, params: DiscreteFourierTransformSpectrumParams) -> None:
        self.params = params
        self.prices: list[float] = []
        self.hp: list[float] = []
        self.filt: list[float] = []
        self.max_power: float = 0.0
        self.dominant: float = 0.0
        self.identified: bool = False
        self.periods = range(params.min_period, params.max_period + 1)

    @staticmethod
    def _previous(series: list[float], lag: int) -> float:
        return series[-1 - lag] if len(series) > lag else 0.0

    def update(self, price: float) -> tuple[list[float], float, float, float]:
        params = self.params
        theta = math.sqrt(2.0) * math.pi / params.highpass_period
        alpha = (math.cos(theta) + math.sin(theta) - 1.0) / math.cos(theta)
        high = (
            (1.0 - alpha / 2.0)
            * (1.0 - alpha / 2.0)
            * (price - 2.0 * self._previous(self.prices, 0) + self._previous(self.prices, 1))
            + 2.0 * (1.0 - alpha) * self._previous(self.hp, 0)
            - (1.0 - alpha) * (1.0 - alpha) * self._previous(self.hp, 1)
        )
        a1 = math.exp(-math.sqrt(2.0) * math.pi / params.smoothing_period)
        c2 = 2.0 * a1 * math.cos(math.sqrt(2.0) * math.pi / params.smoothing_period)
        c3 = -a1 * a1
        c1 = 1.0 - c2 - c3
        filtered = (
            c1 * (high + self._previous(self.hp, 0)) / 2.0
            + c2 * self._previous(self.filt, 0)
            + c3 * self._previous(self.filt, 1)
        )
        self.prices.append(price)
        self.hp.append(high)
        self.filt.append(filtered)

        powers: dict[int, float] = {}
        for period in self.periods:
            comp = float(period) if params.spectral_dilation_compensation else 1.0
            cosine = 0.0
            sine = 0.0
            for n in range(params.window_length):
                value = self._previous(self.filt, n)
                cosine += value * math.cos(2.0 * math.pi * n / period) / comp
                sine += value * math.sin(2.0 * math.pi * n / period) / comp
            powers[period] = cosine * cosine + sine * sine

        self.max_power = 0.995 * self.max_power
        for period in self.periods:
            if powers[period] > self.max_power:
                self.max_power = powers[period]
        for period in self.periods:
            powers[period] = powers[period] / self.max_power if self.max_power else 0.0

        sp = 0.0
        spx = 0.0
        for period in self.periods:
            if powers[period] >= 0.5:
                spx += period * powers[period]
                sp += powers[period]
        if sp != 0.0:
            self.dominant = spx / sp
            self.identified = True
        values = [
            powers[period] ** 3 if params.enhance_resolution else powers[period]
            for period in self.periods
        ]
        return values, (self.dominant if self.identified else math.nan), high, filtered


class TestDiscreteFourierTransformSpectrum(unittest.TestCase):
    def test_reference_cases(self) -> None:
        base_time = datetime.datetime(2020, 1, 1)
        for suffix, params in REFERENCE_CASES:
            with self.subTest(case=suffix):
                indicator = DiscreteFourierTransformSpectrum(params)
                oracle = _ListingOracle(indicator.parameters)
                cells = params.max_period - params.min_period + 1

                expected_heatmap = getattr(data, f"EXPECTED_{suffix}_HEATMAP")
                expected_min = getattr(data, f"EXPECTED_{suffix}_MIN")
                expected_max = getattr(data, f"EXPECTED_{suffix}_MAX")
                expected_dc = getattr(data, f"EXPECTED_{suffix}_DOMINANT_CYCLE")

                for bar_index, price in enumerate(data.INPUT_CLOSE):
                    bar_time = base_time + datetime.timedelta(days=bar_index)
                    result = indicator.update(price, bar_time)
                    oracle_values, oracle_dc, oracle_hp, oracle_filt = oracle.update(price)

                    _assert_close(
                        self,
                        indicator.state.high_pass,
                        oracle_hp,
                        f"{suffix}[{bar_index}].hp",
                        ORACLE_TOLERANCE,
                    )
                    _assert_close(
                        self,
                        indicator.state.filtered,
                        oracle_filt,
                        f"{suffix}[{bar_index}].filt",
                        ORACLE_TOLERANCE,
                    )

                    self.assertEqual(result.heatmap.time, bar_time)
                    self.assertEqual(result.heatmap.parameter_first, float(params.min_period))
                    self.assertEqual(result.heatmap.parameter_last, float(params.max_period))
                    self.assertEqual(result.heatmap.parameter_resolution, 1.0)

                    if bar_index < indicator.warmup_bars:
                        self.assertFalse(indicator.is_primed())
                        self.assertTrue(result.heatmap.is_empty())
                        self.assertTrue(math.isnan(result.heatmap.value_min))
                        self.assertTrue(math.isnan(result.heatmap.value_max))
                        self.assertTrue(math.isnan(result.dominant_cycle))
                    else:
                        self.assertTrue(indicator.is_primed())
                        self.assertEqual(len(result.heatmap.values), cells)
                        for bin_index in range(cells):
                            _assert_close(
                                self,
                                result.heatmap.values[bin_index],
                                expected_heatmap[bar_index][bin_index],
                                f"{suffix}[{bar_index}].heatmap[{bin_index}]",
                            )
                            _assert_close(
                                self,
                                result.heatmap.values[bin_index],
                                oracle_values[bin_index],
                                f"{suffix}[{bar_index}].oracle[{bin_index}]",
                                ORACLE_TOLERANCE,
                            )
                        _assert_close(
                            self,
                            result.dominant_cycle,
                            oracle_dc,
                            f"{suffix}[{bar_index}].oracle_dc",
                            ORACLE_TOLERANCE,
                        )

                    _assert_close(
                        self,
                        result.heatmap.value_min,
                        expected_min[bar_index],
                        f"{suffix}[{bar_index}].min",
                    )
                    _assert_close(
                        self,
                        result.heatmap.value_max,
                        expected_max[bar_index],
                        f"{suffix}[{bar_index}].max",
                    )
                    _assert_close(
                        self,
                        result.dominant_cycle,
                        expected_dc[bar_index],
                        f"{suffix}[{bar_index}].dc",
                    )

                indicator.reset()
                self.assertFalse(indicator.is_primed())
                self.assertEqual(indicator.samples_seen, 0)

    def test_missing_infinity_and_overflow_preserve_state(self) -> None:
        indicator = DiscreteFourierTransformSpectrum()
        control = DiscreteFourierTransformSpectrum()
        for price in data.INPUT_CLOSE[:120]:
            indicator.update(price)
            control.update(price)
        state = indicator.state
        last_output = indicator.output
        count = indicator.samples_seen
        missing_time = datetime.datetime(2020, 6, 1)

        missing_result = indicator.update(math.nan, missing_time)
        self.assertTrue(missing_result.heatmap.is_empty())
        self.assertTrue(math.isnan(missing_result.dominant_cycle))
        self.assertEqual(missing_result.heatmap.time, missing_time)
        self.assertEqual(indicator.state, state)
        self.assertIs(indicator.output, last_output)
        self.assertEqual(indicator.samples_seen, count)

        for bad_sample in (math.inf, -math.inf, True, 1e308):
            with self.assertRaises(ValueError):
                indicator.update(bad_sample)  # type: ignore[arg-type]
            self.assertEqual(indicator.state, state)
            self.assertIs(indicator.output, last_output)
            self.assertEqual(indicator.samples_seen, count)

        actual = indicator.update(data.INPUT_CLOSE[120])
        expected = control.update(data.INPUT_CLOSE[120])
        _assert_same_output(self, actual, expected, "after_rejected_samples")

    def test_owned_snapshot_and_reset_replay(self) -> None:
        indicator = DiscreteFourierTransformSpectrum()
        outputs = [indicator.update(price) for price in data.INPUT_CLOSE]
        last_heatmap = outputs[-1].heatmap
        snapshot = list(last_heatmap.values)
        indicator.update(101.0)
        self.assertEqual(last_heatmap.values, snapshot)

        last_heatmap.values[0] = -999.0
        control = DiscreteFourierTransformSpectrum()
        for price in list(data.INPUT_CLOSE) + [101.0]:
            control.update(price)
        _assert_same_output(
            self, indicator.update(102.0), control.update(102.0), "mutated_snapshot_isolation"
        )

        indicator.reset()
        replay = [indicator.update(price) for price in data.INPUT_CLOSE]
        original = DiscreteFourierTransformSpectrum()
        reference = [original.update(price) for price in data.INPUT_CLOSE]
        for index, (actual, expected) in enumerate(zip(replay, reference)):
            _assert_same_output(self, actual, expected, f"replay[{index}]")

    def test_zero_input_and_degenerate_branch(self) -> None:
        indicator = DiscreteFourierTransformSpectrum()
        result = indicator.output
        for _ in range(indicator.warmup_bars + 4):
            result = indicator.update(0.0)
        self.assertTrue(indicator.is_primed())
        self.assertEqual(result.heatmap.values, [0.0] * 39)
        self.assertTrue(math.isnan(result.dominant_cycle))
        self.assertEqual(indicator.state.normalization_maximum, 0.0)

    def test_single_bin_and_record_boundary(self) -> None:
        boundary = DiscreteFourierTransformSpectrum(
            DiscreteFourierTransformSpectrumParams(window_length=252)
        )
        results = [boundary.update(price) for price in data.INPUT_CLOSE]
        self.assertEqual(boundary.warmup_bars, 251)
        self.assertTrue(all(item.heatmap.is_empty() for item in results[:-1]))
        self.assertFalse(results[-1].heatmap.is_empty())

        single = DiscreteFourierTransformSpectrum(
            DiscreteFourierTransformSpectrumParams(min_period=20, max_period=20)
        )
        for price in data.INPUT_CLOSE:
            single_result = single.update(price)
            if not math.isnan(single_result.dominant_cycle):
                _assert_close(self, single_result.dominant_cycle, 20.0, "single_bin_dc")

    def test_fourier_compensation_and_display_enhancement(self) -> None:
        compensated = DiscreteFourierTransformSpectrum()
        plain = DiscreteFourierTransformSpectrum(
            DiscreteFourierTransformSpectrumParams(spectral_dilation_compensation=False)
        )
        enhanced = DiscreteFourierTransformSpectrum(
            DiscreteFourierTransformSpectrumParams(enhance_resolution=True)
        )
        for index, price in enumerate(data.INPUT_CLOSE):
            comp_out = compensated.update(price)
            plain.update(price)
            enh_out = enhanced.update(price)
            self.assertEqual(compensated.state, enhanced.state)
            _assert_close(
                self, comp_out.dominant_cycle, enh_out.dominant_cycle, f"enh_dc[{index}]"
            )
            if not comp_out.heatmap.is_empty():
                for bin_index, (base_val, enh_val) in enumerate(
                    zip(comp_out.heatmap.values, enh_out.heatmap.values)
                ):
                    _assert_close(
                        self,
                        enh_val,
                        base_val * base_val * base_val,
                        f"enh_cube[{index}][{bin_index}]",
                    )
        for bin_index, period in enumerate(range(10, 49)):
            _assert_close(
                self,
                compensated.state.raw_power[bin_index],
                plain.state.raw_power[bin_index] / (period * period),
                f"comp_ratio[{period}]",
                ORACLE_TOLERANCE,
            )

    def test_sine_wave_selected_period(self) -> None:
        indicator = DiscreteFourierTransformSpectrum(
            DiscreteFourierTransformSpectrumParams(window_length=120)
        )
        result = indicator.output
        for step in range(1000):
            result = indicator.update(math.sin(2.0 * math.pi * step / 20.0))
        peak_period = (
            max(range(len(result.heatmap.values)), key=result.heatmap.values.__getitem__)
            + 10
        )
        self.assertTrue(abs(peak_period - 20) <= 1, f"peak_period={peak_period}")
        self.assertTrue(
            17.0 < result.dominant_cycle < 23.0, f"dominant_cycle={result.dominant_cycle}"
        )

    def test_metadata_mnemonics_entities_and_factory(self) -> None:
        desc = descriptor_of(Identifier.DISCRETE_FOURIER_TRANSFORM_SPECTRUM)
        self.assertEqual(desc.adaptivity, Adaptivity.STATIC)
        self.assertEqual(len(desc.outputs), 2)

        indicator = DiscreteFourierTransformSpectrum.create_default()
        meta = indicator.metadata()
        self.assertEqual(meta.identifier, Identifier.DISCRETE_FOURIER_TRANSFORM_SPECTRUM)
        self.assertEqual(meta.mnemonic, "dftps(10, 48, 48, 48, 10)")
        self.assertEqual(
            meta.description, "Discrete Fourier transform spectrum dftps(10, 48, 48, 48, 10)"
        )
        self.assertEqual(len(meta.outputs), 2)
        self.assertEqual(
            meta.outputs[DiscreteFourierTransformSpectrumOutput.VALUE].shape, Shape.HEATMAP
        )
        self.assertEqual(
            meta.outputs[DiscreteFourierTransformSpectrumOutput.VALUE].mnemonic,
            "dftps(10, 48, 48, 48, 10)",
        )
        self.assertEqual(
            meta.outputs[DiscreteFourierTransformSpectrumOutput.DOMINANT_CYCLE].shape,
            Shape.SCALAR,
        )
        self.assertEqual(
            meta.outputs[DiscreteFourierTransformSpectrumOutput.DOMINANT_CYCLE].mnemonic,
            "dftps-dc(10, 48, 48, 48, 10)",
        )
        self.assertEqual(
            meta.outputs[DiscreteFourierTransformSpectrumOutput.DOMINANT_CYCLE].description,
            "Discrete Fourier transform spectrum dominant cycle dftps-dc(10, 48, 48, 48, 10)",
        )

        flagged = DiscreteFourierTransformSpectrum(
            DiscreteFourierTransformSpectrumParams(
                smoothing_period=10.5,
                spectral_dilation_compensation=False,
                enhance_resolution=True,
            )
        )
        self.assertEqual(flagged.metadata().mnemonic, "dftps(10, 48, 48, 48, 10.5, no-sdc, enh)")
        self.assertEqual(
            flagged.metadata().outputs[DiscreteFourierTransformSpectrumOutput.DOMINANT_CYCLE].mnemonic,
            "dftps-dc(10, 48, 48, 48, 10.5, no-sdc)",
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
            params = DiscreteFourierTransformSpectrumParams(
                bar_component=bar_comp,
                quote_component=quote_comp,
                trade_component=trade_comp,
            )
            instance = DiscreteFourierTransformSpectrum(params)
            resolved_bar = bar_comp if bar_comp is not None else DEFAULT_BAR_COMPONENT
            resolved_quote = quote_comp if quote_comp is not None else DEFAULT_QUOTE_COMPONENT
            resolved_trade = trade_comp if trade_comp is not None else DEFAULT_TRADE_COMPONENT
            comp_suffix = component_triple_mnemonic(
                resolved_bar, resolved_quote, resolved_trade
            )
            expected_primary = f"dftps(10, 48, 48, 48, 10{comp_suffix})"
            expected_dc = f"dftps-dc(10, 48, 48, 48, 10{comp_suffix})"
            self.assertEqual(instance.metadata().mnemonic, expected_primary)
            self.assertEqual(
                discrete_fourier_transform_spectrum_mnemonic(params), expected_primary
            )
            self.assertEqual(
                discrete_fourier_transform_spectrum_dominant_cycle_mnemonic(params),
                expected_dc,
            )

        base_time = datetime.datetime(2020, 1, 1)
        bar_ind = DiscreteFourierTransformSpectrum(default_params())
        scalar_ind = DiscreteFourierTransformSpectrum(default_params())
        quote_ind = DiscreteFourierTransformSpectrum(default_params())
        trade_ind = DiscreteFourierTransformSpectrum(default_params())
        for index in range(60):
            bar_time = base_time + datetime.timedelta(days=index)
            price = data.INPUT_CLOSE[index]
            out_bar = bar_ind.update_bar(
                Bar(
                    bar_time,
                    data.INPUT_OPEN[index],
                    data.INPUT_HIGH[index],
                    data.INPUT_LOW[index],
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
                self, out_bar[1].value, out_scalar[1].value, f"bar_vs_scalar[{index}]"
            )
            _assert_close(
                self, out_bar[1].value, out_quote[1].value, f"bar_vs_quote[{index}]"
            )
            _assert_close(
                self, out_bar[1].value, out_trade[1].value, f"bar_vs_trade[{index}]"
            )
            self.assertEqual(out_bar[0].values, out_scalar[0].values)

        from_factory_default = create_indicator(
            Identifier.DISCRETE_FOURIER_TRANSFORM_SPECTRUM
        )
        self.assertEqual(
            from_factory_default.metadata().mnemonic, "dftps(10, 48, 48, 48, 10)"
        )
        from_factory_custom = create_indicator(
            Identifier.DISCRETE_FOURIER_TRANSFORM_SPECTRUM,
            {
                "minPeriod": 8,
                "maxPeriod": 32,
                "windowLength": 24,
                "highpassPeriod": 40.0,
                "smoothingPeriod": 8.0,
                "spectralDilationCompensation": False,
                "enhanceResolution": True,
                "barComponent": BarComponent.MEDIAN,
            },
        )
        self.assertEqual(
            from_factory_custom.metadata().mnemonic,
            "dftps(8, 32, 24, 40, 8, no-sdc, enh, hl/2)",
        )

    def test_invalid_parameters_and_factory_rejection(self) -> None:
        invalid_params = (
            DiscreteFourierTransformSpectrumParams(min_period=2),
            DiscreteFourierTransformSpectrumParams(min_period=0),
            DiscreteFourierTransformSpectrumParams(min_period=True),  # type: ignore[arg-type]
            DiscreteFourierTransformSpectrumParams(min_period=10, max_period=9),
            DiscreteFourierTransformSpectrumParams(max_period=True),  # type: ignore[arg-type]
            DiscreteFourierTransformSpectrumParams(window_length=1),
            DiscreteFourierTransformSpectrumParams(window_length=0),
            DiscreteFourierTransformSpectrumParams(window_length=48.0),  # type: ignore[arg-type]
            DiscreteFourierTransformSpectrumParams(highpass_period=4.0),
            DiscreteFourierTransformSpectrumParams(highpass_period=math.nan),
            DiscreteFourierTransformSpectrumParams(highpass_period=math.inf),
            DiscreteFourierTransformSpectrumParams(highpass_period=True),  # type: ignore[arg-type]
            DiscreteFourierTransformSpectrumParams(smoothing_period=1.0),
            DiscreteFourierTransformSpectrumParams(smoothing_period=math.nan),
            DiscreteFourierTransformSpectrumParams(smoothing_period=False),  # type: ignore[arg-type]
            DiscreteFourierTransformSpectrumParams(spectral_dilation_compensation=1),  # type: ignore[arg-type]
            DiscreteFourierTransformSpectrumParams(enhance_resolution="yes"),  # type: ignore[arg-type]
        )
        for bad_params in invalid_params:
            with self.subTest(bad_params=bad_params), self.assertRaises(ValueError):
                DiscreteFourierTransformSpectrum(bad_params)

        invalid_dicts = (
            {"length": 48},
            {"spectrumResolution": 1},
            {"disableSpectralDilationCompensation": True},
            {"disableAutomaticGainControl": True},
            {"automaticGainControlDecayFactor": 0.995},
            {"fixedNormalization": True},
            {"unknownParam": 1},
            {"minPeriod": 12, "min_period": 12},
            {"minPeriod": 10.5},
            {"windowLength": 1},
            {"spectralDilationCompensation": 0},
        )
        for bad_dict in invalid_dicts:
            with self.subTest(bad_dict=bad_dict), self.assertRaises(ValueError):
                create_indicator(
                    Identifier.DISCRETE_FOURIER_TRANSFORM_SPECTRUM, bad_dict
                )


if __name__ == "__main__":
    unittest.main()
