"""Tests for Ehlers' Comb Band-Pass Spectrum indicator."""

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
from py.indicators.john_ehlers.comb_band_pass_spectrum import (
    CombBandPassSpectrum,
    CombBandPassSpectrumOutput,
    CombBandPassSpectrumParams,
    SpectrumOutput,
    comb_band_pass_spectrum_dominant_cycle_mnemonic,
    comb_band_pass_spectrum_mnemonic,
    default_params,
)
from py.indicators.john_ehlers.comb_band_pass_spectrum import (
    test_testdata as data,
)

TOLERANCE = 1e-13
ORACLE_TOLERANCE = 2e-11

REFERENCE_CASES: tuple[tuple[str, CombBandPassSpectrumParams], ...] = (
    (
        "P010_P148_HP48_LP10_SDC1_E0_B0_3",
        CombBandPassSpectrumParams(
            min_period=10,
            max_period=48,
            bandwidth=0.3,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P148_HP48_LP10_SDC0_E0_B0_3",
        CombBandPassSpectrumParams(
            min_period=10,
            max_period=48,
            bandwidth=0.3,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=False,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P148_HP48_LP10_SDC1_E1_B0_3",
        CombBandPassSpectrumParams(
            min_period=10,
            max_period=48,
            bandwidth=0.3,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=True,
            enhance_resolution=True,
        ),
    ),
    (
        "P010_P148_HP48_LP10_SDC0_E1_B0_3",
        CombBandPassSpectrumParams(
            min_period=10,
            max_period=48,
            bandwidth=0.3,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=False,
            enhance_resolution=True,
        ),
    ),
    (
        "P010_P148_HP48_LP10_SDC1_E0_B0_1",
        CombBandPassSpectrumParams(
            min_period=10,
            max_period=48,
            bandwidth=0.1,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P148_HP48_LP10_SDC1_E0_B0_7",
        CombBandPassSpectrumParams(
            min_period=10,
            max_period=48,
            bandwidth=0.7,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P148_HP48_LP10_5_SDC1_E0_B0_3",
        CombBandPassSpectrumParams(
            min_period=10,
            max_period=48,
            bandwidth=0.3,
            highpass_period=48.0,
            smoothing_period=10.5,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P148_HP80_LP20_SDC1_E0_B0_3",
        CombBandPassSpectrumParams(
            min_period=10,
            max_period=48,
            bandwidth=0.3,
            highpass_period=80.0,
            smoothing_period=20.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P06_P124_HP24_LP6_SDC1_E0_B0_3",
        CombBandPassSpectrumParams(
            min_period=6,
            max_period=24,
            bandwidth=0.3,
            highpass_period=24.0,
            smoothing_period=6.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P020_P120_HP48_LP10_SDC1_E0_B0_3",
        CombBandPassSpectrumParams(
            min_period=20,
            max_period=20,
            bandwidth=0.3,
            highpass_period=48.0,
            smoothing_period=10.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P024_P160_HP60_LP12_SDC1_E0_B0_3",
        CombBandPassSpectrumParams(
            min_period=24,
            max_period=60,
            bandwidth=0.3,
            highpass_period=60.0,
            smoothing_period=12.0,
            spectral_dilation_compensation=True,
            enhance_resolution=False,
        ),
    ),
    (
        "P010_P1240_HP240_LP10_SDC1_E0_B0_3",
        CombBandPassSpectrumParams(
            min_period=10,
            max_period=240,
            bandwidth=0.3,
            highpass_period=240.0,
            smoothing_period=10.0,
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
        testcase.assertFalse(
            math.isnan(actual), f"{label}: expected {expected}, got NaN"
        )
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
        actual.heatmap.parameter_first,
        expected.heatmap.parameter_first,
        f"{label}.first",
    )
    testcase.assertEqual(
        actual.heatmap.parameter_last, expected.heatmap.parameter_last, f"{label}.last"
    )
    testcase.assertEqual(
        actual.heatmap.parameter_resolution,
        expected.heatmap.parameter_resolution,
        f"{label}.resolution",
    )
    _assert_close(
        testcase, actual.heatmap.value_min, expected.heatmap.value_min, f"{label}.min"
    )
    _assert_close(
        testcase, actual.heatmap.value_max, expected.heatmap.value_max, f"{label}.max"
    )
    testcase.assertEqual(
        len(actual.heatmap.values), len(expected.heatmap.values), f"{label}.len"
    )
    for index, (actual_value, expected_value) in enumerate(
        zip(actual.heatmap.values, expected.heatmap.values)
    ):
        _assert_close(
            testcase, actual_value, expected_value, f"{label}.values[{index}]"
        )
    _assert_close(testcase, actual.dominant_cycle, expected.dominant_cycle, f"{label}.dc")


class _ListingOracle:
    """Independent one-based direct transcription of Book Listing 10-1."""

    def __init__(self, params: CombBandPassSpectrumParams) -> None:
        self.params = params
        self.prices: list[float] = []
        self.hp: list[float] = []
        self.filt: list[float] = []
        self.max_power: float = 0.0
        self.dominant: float = 0.0
        self.identified: bool = False
        self.periods = range(params.min_period, params.max_period + 1)
        self.bp: dict[int, list[float]] = {
            period: [0.0] * (params.max_period + 1) for period in self.periods
        }

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
            * (
                price
                - 2.0 * self._previous(self.prices, 0)
                + self._previous(self.prices, 1)
            )
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
            row = self.bp[period]
            for m in range(params.max_period, 1, -1):
                row[m] = row[m - 1]
            if len(self.prices) > 12:
                beta = math.cos(2.0 * math.pi / period)
                gamma = 1.0 / math.cos(2.0 * math.pi * params.bandwidth / period)
                pole = gamma - math.sqrt(gamma * gamma - 1.0)
                row[1] = (
                    0.5 * (1.0 - pole) * (filtered - self._previous(self.filt, 2))
                    + beta * (1.0 + pole) * row[2]
                    - pole * row[3]
                )
            powers[period] = 0.0
            for m in range(1, period + 1):
                powers[period] += (row[m] / comp) * (row[m] / comp)

        self.max_power = 0.995 * self.max_power
        for period in self.periods:
            if powers[period] > self.max_power:
                self.max_power = powers[period]
        for period in self.periods:
            powers[period] = (
                powers[period] / self.max_power if self.max_power else 0.0
            )

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


class TestCombBandPassSpectrum(unittest.TestCase):
    def test_reference_cases(self) -> None:
        base_time = datetime.datetime(2020, 1, 1)
        for suffix, params in REFERENCE_CASES:
            with self.subTest(case=suffix):
                indicator = CombBandPassSpectrum(params)
                oracle = _ListingOracle(indicator.parameters)
                cells = params.max_period - params.min_period + 1

                expected_heatmap = getattr(data, f"EXPECTED_{suffix}_HEATMAP")
                expected_min = getattr(data, f"EXPECTED_{suffix}_MIN")
                expected_max = getattr(data, f"EXPECTED_{suffix}_MAX")
                expected_dc = getattr(data, f"EXPECTED_{suffix}_DOMINANT_CYCLE")

                for bar_index, price in enumerate(data.INPUT_CLOSE):
                    bar_time = base_time + datetime.timedelta(days=bar_index)
                    result = indicator.update(price, bar_time)
                    oracle_values, oracle_dc, oracle_hp, oracle_filt = oracle.update(
                        price
                    )

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
                    self.assertEqual(
                        result.heatmap.parameter_first, float(params.min_period)
                    )
                    self.assertEqual(
                        result.heatmap.parameter_last, float(params.max_period)
                    )
                    self.assertEqual(result.heatmap.parameter_resolution, 1.0)

                    if bar_index < indicator.warmup_bars:
                        self.assertFalse(indicator.is_primed())
                        self.assertTrue(result.heatmap.is_empty())
                        self.assertTrue(math.isnan(result.heatmap.value_min))
                        self.assertTrue(math.isnan(result.heatmap.value_max))
                        self.assertTrue(math.isnan(result.dominant_cycle))
                        self.assertTrue(math.isnan(expected_min[bar_index]))
                        self.assertTrue(math.isnan(expected_max[bar_index]))
                        self.assertTrue(math.isnan(expected_dc[bar_index]))
                        self.assertTrue(
                            all(math.isnan(cell) for cell in expected_heatmap[bar_index])
                        )
                    else:
                        self.assertTrue(indicator.is_primed())
                        self.assertEqual(len(result.heatmap.values), cells)
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
                        _assert_close(
                            self,
                            result.dominant_cycle,
                            oracle_dc,
                            f"{suffix}[{bar_index}].oracle_dc",
                            ORACLE_TOLERANCE,
                        )
                        for bin_index in range(cells):
                            _assert_close(
                                self,
                                result.heatmap.values[bin_index],
                                expected_heatmap[bar_index][bin_index],
                                f"{suffix}[{bar_index}][{bin_index}]",
                            )
                            _assert_close(
                                self,
                                result.heatmap.values[bin_index],
                                oracle_values[bin_index],
                                f"{suffix}[{bar_index}].oracle[{bin_index}]",
                                ORACLE_TOLERANCE,
                            )

    def test_source_gate_twelfth_vs_thirteenth_bar(self) -> None:
        indicator = CombBandPassSpectrum.create_default()
        for price in data.INPUT_CLOSE[:12]:
            indicator.update(price)
            self.assertTrue(
                all(value == 0.0 for row in indicator.band_pass_history for value in row)
            )
        indicator.update(data.INPUT_CLOSE[12])
        self.assertTrue(any(row[0] != 0.0 for row in indicator.band_pass_history))

    def test_missing_infinity_and_overflow_preserve_state(self) -> None:
        indicator = CombBandPassSpectrum.create_default()
        control = CombBandPassSpectrum.create_default()
        for price in data.INPUT_CLOSE[:65]:
            indicator.update(price)
            control.update(price)

        saved_state = indicator.state
        saved_history = indicator.band_pass_history
        saved_output = indicator.output
        saved_count = indicator.samples_seen

        nan_time = datetime.datetime(2025, 1, 1)
        missing = indicator.update(math.nan, nan_time)
        self.assertTrue(missing.heatmap.is_empty())
        self.assertEqual(missing.heatmap.time, nan_time)
        self.assertTrue(math.isnan(missing.dominant_cycle))
        self.assertEqual(indicator.state, saved_state)
        self.assertEqual(indicator.band_pass_history, saved_history)
        _assert_same_output(self, indicator.output, saved_output, "nan_preserved")
        self.assertEqual(indicator.samples_seen, saved_count)

        for invalid in (math.inf, -math.inf, True, False, 1e308):
            with self.assertRaises(ValueError):
                indicator.update(invalid)  # type: ignore[arg-type]
            self.assertEqual(indicator.state, saved_state)
            self.assertEqual(indicator.band_pass_history, saved_history)
            _assert_same_output(self, indicator.output, saved_output, "err_preserved")
            self.assertEqual(indicator.samples_seen, saved_count)

        next_a = indicator.update(data.INPUT_CLOSE[65])
        next_b = control.update(data.INPUT_CLOSE[65])
        _assert_same_output(self, next_a, next_b, "post_skip")
        self.assertEqual(indicator.state, control.state)
        self.assertEqual(indicator.band_pass_history, control.band_pass_history)

    def test_owned_snapshot_and_reset_replay(self) -> None:
        indicator = CombBandPassSpectrum.create_default()
        first_run: list[SpectrumOutput] = [
            indicator.update(price) for price in data.INPUT_CLOSE[:80]
        ]
        primed_snapshot = list(first_run[60].heatmap.values)
        first_run[60].heatmap.values[0] = -999.0

        fresh = CombBandPassSpectrum.create_default()
        for price in data.INPUT_CLOSE[:61]:
            fresh_out = fresh.update(price)
        self.assertEqual(fresh_out.heatmap.values, primed_snapshot)

        indicator.reset()
        self.assertFalse(indicator.is_primed())
        self.assertEqual(indicator.samples_seen, 0)
        self.assertTrue(indicator.output.heatmap.is_empty())
        self.assertTrue(math.isnan(indicator.output.dominant_cycle))
        second_run = [indicator.update(price) for price in data.INPUT_CLOSE[:80]]
        replay = CombBandPassSpectrum.create_default()
        for index, price in enumerate(data.INPUT_CLOSE[:80]):
            _assert_same_output(
                self, second_run[index], replay.update(price), f"replay[{index}]"
            )

    def test_zero_input_and_degenerate_branch(self) -> None:
        indicator = CombBandPassSpectrum.create_default()
        for _ in range(70):
            result = indicator.update(0.0)
        self.assertTrue(indicator.is_primed())
        self.assertEqual(result.heatmap.values, [0.0] * 39)
        self.assertEqual(result.heatmap.value_min, 0.0)
        self.assertEqual(result.heatmap.value_max, 0.0)
        self.assertTrue(math.isnan(result.dominant_cycle))

    def test_single_bin_and_record_boundary(self) -> None:
        single = CombBandPassSpectrum(
            CombBandPassSpectrumParams(min_period=20, max_period=20)
        )
        self.assertEqual(single.warmup_bars, 31)
        for price in data.INPUT_CLOSE:
            single_out = single.update(price)
            if not math.isnan(single_out.dominant_cycle):
                _assert_close(self, single_out.dominant_cycle, 20.0, "single_dc")
        self.assertEqual(len(single_out.heatmap.values), 1)

        boundary = CombBandPassSpectrum(
            CombBandPassSpectrumParams(
                min_period=10,
                max_period=240,
                highpass_period=240.0,
                smoothing_period=10.0,
            )
        )
        self.assertEqual(boundary.warmup_bars, 251)
        results = [boundary.update(price) for price in data.INPUT_CLOSE]
        self.assertTrue(all(item.heatmap.is_empty() for item in results[:-1]))
        self.assertFalse(results[-1].heatmap.is_empty())
        self.assertEqual(len(results[-1].heatmap.values), 231)

    def test_compensation_and_display_enhancement_never_changes_centroid(self) -> None:
        compensated = CombBandPassSpectrum.create_default()
        uncompensated = CombBandPassSpectrum(
            CombBandPassSpectrumParams(spectral_dilation_compensation=False)
        )
        enhanced = CombBandPassSpectrum(
            CombBandPassSpectrumParams(enhance_resolution=True)
        )
        for bar_index, price in enumerate(data.INPUT_CLOSE):
            out_comp = compensated.update(price)
            uncompensated.update(price)
            out_enh = enhanced.update(price)
            self.assertEqual(compensated.state, enhanced.state)
            _assert_close(
                self,
                out_comp.dominant_cycle,
                out_enh.dominant_cycle,
                f"enh_dc[{bar_index}]",
            )
            for index, period in enumerate(range(10, 49)):
                _assert_close(
                    self,
                    compensated.state.raw_power[index],
                    uncompensated.state.raw_power[index] / float(period * period),
                    f"sdc_raw[{bar_index}][{index}]",
                )
            if not out_comp.heatmap.is_empty():
                for index, (plain_val, enh_val) in enumerate(
                    zip(out_comp.heatmap.values, out_enh.heatmap.values)
                ):
                    _assert_close(
                        self,
                        enh_val,
                        plain_val * plain_val * plain_val,
                        f"cube[{bar_index}][{index}]",
                    )

    def test_sine_wave_selected_period(self) -> None:
        indicator = CombBandPassSpectrum.create_default()
        for step in range(1000):
            result = indicator.update(math.sin(2.0 * math.pi * step / 20.0))
        peak_period = (
            max(
                range(len(result.heatmap.values)),
                key=result.heatmap.values.__getitem__,
            )
            + 10
        )
        self.assertLessEqual(abs(peak_period - 20), 2)
        self.assertGreater(result.dominant_cycle, 17.0)
        self.assertLess(result.dominant_cycle, 23.0)

    def test_metadata_mnemonics_entities_and_factory(self) -> None:
        self.assertEqual(int(CombBandPassSpectrumOutput.VALUE), 0)
        self.assertEqual(int(CombBandPassSpectrumOutput.DOMINANT_CYCLE), 1)

        desc = descriptor_of(Identifier.COMB_BAND_PASS_SPECTRUM)
        self.assertIsNotNone(desc)
        assert desc is not None
        self.assertEqual(desc.adaptivity, Adaptivity.STATIC)
        self.assertEqual(len(desc.outputs), 2)

        indicator = CombBandPassSpectrum.create_default()
        meta = indicator.metadata()
        self.assertEqual(meta.identifier, Identifier.COMB_BAND_PASS_SPECTRUM)
        self.assertEqual(meta.mnemonic, "cbps(10, 48, 0.3, 48, 10)")
        self.assertEqual(
            meta.description,
            "Comb band-pass spectrum cbps(10, 48, 0.3, 48, 10)",
        )
        self.assertEqual(len(meta.outputs), 2)
        self.assertEqual(meta.outputs[0].shape, Shape.HEATMAP)
        self.assertEqual(meta.outputs[0].mnemonic, "cbps(10, 48, 0.3, 48, 10)")
        self.assertEqual(meta.outputs[1].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[1].mnemonic, "cbps-dc(10, 48, 0.3, 48, 10)")
        self.assertEqual(
            meta.outputs[1].description,
            "Comb band-pass spectrum dominant cycle cbps-dc(10, 48, 0.3, 48, 10)",
        )

        mnemonic_cases = (
            (
                default_params(),
                "cbps(10, 48, 0.3, 48, 10)",
                "cbps-dc(10, 48, 0.3, 48, 10)",
            ),
            (
                CombBandPassSpectrumParams(
                    bandwidth=0.5,
                    spectral_dilation_compensation=False,
                    enhance_resolution=True,
                    smoothing_period=10.5,
                ),
                "cbps(10, 48, 0.5, 48, 10.5, no-sdc, enh)",
                "cbps-dc(10, 48, 0.5, 48, 10.5, no-sdc)",
            ),
            (
                CombBandPassSpectrumParams(bar_component=BarComponent.MEDIAN),
                "cbps(10, 48, 0.3, 48, 10, hl/2)",
                "cbps-dc(10, 48, 0.3, 48, 10, hl/2)",
            ),
            (
                CombBandPassSpectrumParams(quote_component=QuoteComponent.BID),
                "cbps(10, 48, 0.3, 48, 10, b)",
                "cbps-dc(10, 48, 0.3, 48, 10, b)",
            ),
            (
                CombBandPassSpectrumParams(trade_component=TradeComponent.VOLUME),
                "cbps(10, 48, 0.3, 48, 10, v)",
                "cbps-dc(10, 48, 0.3, 48, 10, v)",
            ),
            (
                CombBandPassSpectrumParams(
                    bar_component=BarComponent.OPEN,
                    quote_component=QuoteComponent.ASK,
                ),
                "cbps(10, 48, 0.3, 48, 10, o, a)",
                "cbps-dc(10, 48, 0.3, 48, 10, o, a)",
            ),
            (
                CombBandPassSpectrumParams(
                    bar_component=BarComponent.HIGH,
                    trade_component=TradeComponent.VOLUME,
                ),
                "cbps(10, 48, 0.3, 48, 10, h, v)",
                "cbps-dc(10, 48, 0.3, 48, 10, h, v)",
            ),
            (
                CombBandPassSpectrumParams(
                    quote_component=QuoteComponent.SPREAD_BP,
                    trade_component=TradeComponent.VOLUME,
                ),
                "cbps(10, 48, 0.3, 48, 10, spread bp, v)",
                "cbps-dc(10, 48, 0.3, 48, 10, spread bp, v)",
            ),
        )
        for case_params, expected_mn, expected_dc_mn in mnemonic_cases:
            self.assertEqual(
                comb_band_pass_spectrum_mnemonic(case_params), expected_mn
            )
            self.assertEqual(
                comb_band_pass_spectrum_dominant_cycle_mnemonic(case_params),
                expected_dc_mn,
            )

        self.assertEqual(
            component_triple_mnemonic(
                DEFAULT_BAR_COMPONENT,
                DEFAULT_QUOTE_COMPONENT,
                DEFAULT_TRADE_COMPONENT,
            ),
            "",
        )

        base_time = datetime.datetime(2020, 1, 1)
        bar_ind = CombBandPassSpectrum(default_params())
        scalar_ind = CombBandPassSpectrum(default_params())
        quote_ind = CombBandPassSpectrum(default_params())
        trade_ind = CombBandPassSpectrum(default_params())
        for index in range(70):
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

        from_factory_default = create_indicator(Identifier.COMB_BAND_PASS_SPECTRUM)
        self.assertEqual(
            from_factory_default.metadata().mnemonic, "cbps(10, 48, 0.3, 48, 10)"
        )
        from_factory_custom = create_indicator(
            Identifier.COMB_BAND_PASS_SPECTRUM,
            {
                "minPeriod": 8,
                "maxPeriod": 32,
                "bandwidth": 0.4,
                "highpassPeriod": 40.0,
                "smoothingPeriod": 8.0,
                "spectralDilationCompensation": False,
                "enhanceResolution": True,
                "barComponent": BarComponent.MEDIAN,
            },
        )
        self.assertEqual(
            from_factory_custom.metadata().mnemonic,
            "cbps(8, 32, 0.4, 40, 8, no-sdc, enh, hl/2)",
        )

    def test_invalid_parameters_and_factory_rejection(self) -> None:
        invalid_params = (
            CombBandPassSpectrumParams(min_period=2),
            CombBandPassSpectrumParams(min_period=0),
            CombBandPassSpectrumParams(min_period=True),  # type: ignore[arg-type]
            CombBandPassSpectrumParams(min_period=10, max_period=9),
            CombBandPassSpectrumParams(max_period=True),  # type: ignore[arg-type]
            CombBandPassSpectrumParams(bandwidth=0.0),
            CombBandPassSpectrumParams(bandwidth=-0.1),
            CombBandPassSpectrumParams(bandwidth=2.5),
            CombBandPassSpectrumParams(min_period=3, max_period=10, bandwidth=0.75),
            CombBandPassSpectrumParams(bandwidth=math.nan),
            CombBandPassSpectrumParams(bandwidth=math.inf),
            CombBandPassSpectrumParams(bandwidth=True),  # type: ignore[arg-type]
            CombBandPassSpectrumParams(highpass_period=4.0),
            CombBandPassSpectrumParams(highpass_period=math.nan),
            CombBandPassSpectrumParams(highpass_period=math.inf),
            CombBandPassSpectrumParams(highpass_period=True),  # type: ignore[arg-type]
            CombBandPassSpectrumParams(smoothing_period=1.0),
            CombBandPassSpectrumParams(smoothing_period=math.nan),
            CombBandPassSpectrumParams(smoothing_period=False),  # type: ignore[arg-type]
            CombBandPassSpectrumParams(spectral_dilation_compensation=1),  # type: ignore[arg-type]
            CombBandPassSpectrumParams(enhance_resolution="yes"),  # type: ignore[arg-type]
            CombBandPassSpectrumParams(bar_component=True),  # type: ignore[arg-type]
        )
        for bad_params in invalid_params:
            with self.subTest(bad_params=bad_params), self.assertRaises(ValueError):
                CombBandPassSpectrum(bad_params)

        invalid_dicts = (
            {"length": 48},
            {"windowLength": 48},
            {"disableSpectralDilationCompensation": True},
            {"disableAutomaticGainControl": True},
            {"automaticGainControlDecayFactor": 0.995},
            {"fixedNormalization": True},
            {"unknownParam": 1},
            {"minPeriod": 12, "min_period": 12},
            {"minPeriod": None},
            {"barComponent": None},
            {"minPeriod": 10.5},
            {"bandwidth": 0.0},
            {"bandwidth": 2.5},
            {"spectralDilationCompensation": 0},
        )
        for bad_dict in invalid_dicts:
            with self.subTest(bad_dict=bad_dict), self.assertRaises(ValueError):
                create_indicator(Identifier.COMB_BAND_PASS_SPECTRUM, bad_dict)


if __name__ == "__main__":
    unittest.main()
