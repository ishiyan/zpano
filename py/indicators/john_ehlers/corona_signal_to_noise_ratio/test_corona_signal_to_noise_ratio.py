"""Tests for the CoronaSignalToNoiseRatio indicator."""

from __future__ import annotations

import datetime
import math
import unittest

from py.entities.bar import Bar
from py.entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from py.entities.quote import Quote
from py.entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from py.entities.scalar import Scalar
from py.entities.trade import Trade
from py.entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from py.indicators.core.component_triple_mnemonic import component_triple_mnemonic
from py.indicators.core.identifier import Identifier
from py.indicators.core.outputs.shape import Shape
from py.indicators.factory.factory import create_indicator
from py.indicators.john_ehlers.corona_signal_to_noise_ratio import (
    CoronaSignalToNoiseRatio,
    CoronaSignalToNoiseRatioOutput,
    CoronaSignalToNoiseRatioParams,
    corona_signal_to_noise_ratio_mnemonic,
    default_params,
)
from py.indicators.john_ehlers.corona_signal_to_noise_ratio import test_testdata as data

TOLERANCE = 5e-12


def _assert_close(testcase: unittest.TestCase, actual: float, expected: float, label: str) -> None:
    if math.isnan(expected):
        testcase.assertTrue(math.isnan(actual), f"{label}: expected NaN, got {actual}")
    elif math.isinf(expected):
        testcase.assertEqual(actual, expected, f"{label}: expected {expected}, got {actual}")
    else:
        testcase.assertFalse(math.isnan(actual), f"{label}: expected {expected}, got NaN")
        bound = max(TOLERANCE, abs(expected) * TOLERANCE)
        testcase.assertAlmostEqual(actual, expected, delta=bound, msg=label)


class TestCoronaSignalToNoiseRatio(unittest.TestCase):
    def test_reference_cases(self):
        for hp in (30, 48):
            for min_period, max_period in ((6, 30), (6, 20), (10, 30), (10, 48)):
                for raster_length in (50, 25):
                    suffix = f"H{hp}_PMIN{min_period}_PMAX{max_period}_R{raster_length}"
                    with self.subTest(case=suffix):
                        params = CoronaSignalToNoiseRatioParams(
                            raster_length=raster_length,
                            high_pass_filter_cutoff=hp,
                            minimal_period=min_period,
                            maximal_period=max_period,
                        )
                        indicator = CoronaSignalToNoiseRatio(params)
                        step = 10.0 / raster_length
                        expected_first = 1.0 + step
                        expected_res = 1.0 / step
                        base_time = datetime.datetime(2020, 1, 1)

                        expected_heatmap = getattr(data, f"EXPECTED_{suffix}_HEATMAP")
                        expected_min = getattr(data, f"EXPECTED_{suffix}_HEATMAP_MIN")
                        expected_max = getattr(data, f"EXPECTED_{suffix}_HEATMAP_MAX")
                        expected_snr = getattr(data, f"EXPECTED_{suffix}_SIGNAL_TO_NOISE_RATIO")
                        expected_phys = getattr(data, f"EXPECTED_{suffix}_PHYSICAL_DECIBELS")
                        expected_cal = getattr(data, f"EXPECTED_{suffix}_CALIBRATED_DECIBELS")
                        expected_pos = getattr(data, f"EXPECTED_{suffix}_NORMALIZED_POSITION")
                        expected_sig = getattr(data, f"EXPECTED_{suffix}_SIGNAL")
                        expected_noise = getattr(data, f"EXPECTED_{suffix}_NOISE")
                        expected_med_range = getattr(data, f"EXPECTED_{suffix}_MEDIAN_BAR_RANGE")
                        expected_width = getattr(data, f"EXPECTED_{suffix}_WIDTH")
                        expected_dc = getattr(data, f"EXPECTED_{suffix}_DOMINANT_CYCLE")
                        expected_dcm = getattr(data, f"EXPECTED_{suffix}_DOMINANT_CYCLE_MEDIAN")
                        expected_has_est = getattr(data, f"EXPECTED_{suffix}_HAS_CYCLE_ESTIMATE")
                        expected_finite = getattr(data, f"EXPECTED_{suffix}_RATIO_IS_FINITE")

                        for i, price in enumerate(data.INPUT_CLOSE):
                            bar_time = base_time + datetime.timedelta(days=i)
                            snap = indicator.update(
                                price,
                                bar_time,
                                high=data.INPUT_HIGH[i],
                                low=data.INPUT_LOW[i],
                            )
                            _assert_close(self, snap.heatmap.parameter_first, expected_first, f"{suffix}.first")
                            _assert_close(self, snap.heatmap.parameter_last, 11.0, f"{suffix}.last")
                            _assert_close(self, snap.heatmap.parameter_resolution, expected_res, f"{suffix}.res")

                            if i < 12:
                                self.assertFalse(indicator.is_primed())
                                self.assertTrue(snap.heatmap.is_empty())
                            else:
                                self.assertTrue(indicator.is_primed())
                                self.assertEqual(len(snap.heatmap.values), raster_length)
                                for bin_index in range(raster_length):
                                    _assert_close(
                                        self,
                                        snap.heatmap.values[bin_index],
                                        expected_heatmap[i][bin_index],
                                        f"{suffix}[{i}].heatmap[{bin_index}]",
                                    )

                            _assert_close(self, snap.heatmap.value_min, expected_min[i], f"{suffix}[{i}].min")
                            _assert_close(self, snap.heatmap.value_max, expected_max[i], f"{suffix}[{i}].max")
                            _assert_close(self, snap.signal_to_noise_ratio, expected_snr[i], f"{suffix}[{i}].snr")
                            _assert_close(self, snap.physical_decibels, expected_phys[i], f"{suffix}[{i}].phys")
                            _assert_close(self, snap.calibrated_decibels, expected_cal[i], f"{suffix}[{i}].cal")
                            _assert_close(self, snap.normalized_position, expected_pos[i], f"{suffix}[{i}].pos")
                            _assert_close(self, snap.signal, expected_sig[i], f"{suffix}[{i}].signal")
                            _assert_close(self, snap.noise, expected_noise[i], f"{suffix}[{i}].noise")
                            _assert_close(self, snap.median_bar_range, expected_med_range[i], f"{suffix}[{i}].med_range")
                            _assert_close(self, snap.width, expected_width[i], f"{suffix}[{i}].width")
                            _assert_close(self, snap.dominant_cycle, expected_dc[i], f"{suffix}[{i}].dc")
                            _assert_close(self, snap.dominant_cycle_median, expected_dcm[i], f"{suffix}[{i}].dcm")
                            _assert_close(self, snap.has_cycle_estimate, expected_has_est[i], f"{suffix}[{i}].has_est")
                            _assert_close(self, snap.ratio_is_finite, expected_finite[i], f"{suffix}[{i}].finite")

                        indicator.reset()
                        self.assertFalse(indicator.is_primed())

    def test_metadata_mnemonics_entities_and_negative_min_param(self):
        indicator = CoronaSignalToNoiseRatio(default_params())
        meta = indicator.metadata()
        self.assertEqual(meta.identifier, Identifier.CORONA_SIGNAL_TO_NOISE_RATIO)
        self.assertEqual(meta.mnemonic, "csnr(50, 20, 1, 11, 30)")
        self.assertEqual(len(meta.outputs), 2)
        self.assertEqual(meta.outputs[CoronaSignalToNoiseRatioOutput.VALUE].shape, Shape.HEATMAP)
        self.assertEqual(meta.outputs[CoronaSignalToNoiseRatioOutput.SIGNAL_TO_NOISE_RATIO].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[CoronaSignalToNoiseRatioOutput.SIGNAL_TO_NOISE_RATIO].mnemonic, "csnr-snr(30)")

        neg_bounds = CoronaSignalToNoiseRatio(
            CoronaSignalToNoiseRatioParams(min_parameter_value=-5.0, max_parameter_value=5.0)
        )
        self.assertEqual(neg_bounds.metadata().mnemonic, "csnr(50, 20, -5, 5, 30)")

        flagged = CoronaSignalToNoiseRatio(
            CoronaSignalToNoiseRatioParams(
                minimal_period=10, maximal_period=48, bar_component=BarComponent.MEDIAN
            )
        )
        self.assertEqual(
            flagged.metadata().mnemonic, "csnr(50, 20, 1, 11, 30, pmin=10, pmax=48, hl/2)"
        )
        self.assertEqual(
            flagged.metadata().outputs[1].mnemonic, "csnr-snr(30, pmin=10, pmax=48, hl/2)"
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
        for bc, qc, tc in combinations:
            p = CoronaSignalToNoiseRatioParams(bar_component=bc, quote_component=qc, trade_component=tc)
            inst = CoronaSignalToNoiseRatio(p)
            resolved_bc = bc if bc is not None else DEFAULT_BAR_COMPONENT
            resolved_qc = qc if qc is not None else DEFAULT_QUOTE_COMPONENT
            resolved_tc = tc if tc is not None else DEFAULT_TRADE_COMPONENT
            suffix = component_triple_mnemonic(resolved_bc, resolved_qc, resolved_tc)
            expected = f"csnr(50, 20, 1, 11, 30{suffix})"
            self.assertEqual(inst.metadata().mnemonic, expected)
            self.assertEqual(corona_signal_to_noise_ratio_mnemonic(p), expected)

        # Scalar, Quote, and Trade adapters pass zero range without raising.
        base_time = datetime.datetime(2020, 1, 1)
        bar_ind = CoronaSignalToNoiseRatio()
        scalar_ind = CoronaSignalToNoiseRatio()
        quote_ind = CoronaSignalToNoiseRatio()
        trade_ind = CoronaSignalToNoiseRatio()
        for i in range(15):
            bar_time = base_time + datetime.timedelta(days=i)
            price = data.INPUT_CLOSE[i]
            out_bar = bar_ind.update_bar(
                Bar(bar_time, data.INPUT_OPEN[i], data.INPUT_HIGH[i], data.INPUT_LOW[i], price, 1000.0)
            )
            out_scalar = scalar_ind.update_scalar(Scalar(bar_time, price))
            out_quote = quote_ind.update_quote(Quote(bar_time, price - 0.5, price + 0.5, 10.0, 10.0))
            out_trade = trade_ind.update_trade(Trade(bar_time, price, 10.0))
            self.assertEqual(len(out_bar), 2)
            self.assertEqual(len(out_scalar), 2)
            self.assertEqual(len(out_quote), 2)
            self.assertEqual(len(out_trade), 2)
        self.assertEqual(scalar_ind.update_scalar(Scalar(base_time, 100.0))[1].value, 11.0)

        from_factory = create_indicator(
            Identifier.CORONA_SIGNAL_TO_NOISE_RATIO, {"min_parameter_value": 0.0}
        )
        self.assertEqual(from_factory.metadata().mnemonic, "csnr(50, 20, 0, 11, 30)")

    def test_validation(self):
        invalid_cases = (
            CoronaSignalToNoiseRatioParams(raster_length=1),
            CoronaSignalToNoiseRatioParams(max_raster_value=6.0),
            CoronaSignalToNoiseRatioParams(min_parameter_value=11.0, max_parameter_value=11.0),
            CoronaSignalToNoiseRatioParams(high_pass_filter_cutoff=4),
            CoronaSignalToNoiseRatioParams(minimal_period=5),
            CoronaSignalToNoiseRatioParams(minimal_period=10, maximal_period=10),
        )
        for bad in invalid_cases:
            with self.assertRaises(ValueError):
                CoronaSignalToNoiseRatio(bad)


if __name__ == "__main__":
    unittest.main()
