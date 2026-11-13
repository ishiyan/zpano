"""Tests for the CoronaTrendVigor indicator."""

from __future__ import annotations

import datetime
import math
import unittest

from py.entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from py.entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from py.entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from py.indicators.core.component_triple_mnemonic import component_triple_mnemonic
from py.indicators.core.identifier import Identifier
from py.indicators.core.outputs.shape import Shape
from py.indicators.factory.factory import create_indicator
from py.indicators.john_ehlers.corona_trend_vigor import (
    CoronaTrendVigor,
    CoronaTrendVigorOutput,
    CoronaTrendVigorParams,
    corona_trend_vigor_mnemonic,
    default_params,
)
from py.indicators.john_ehlers.corona_trend_vigor import test_testdata as data

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


class TestCoronaTrendVigor(unittest.TestCase):
    def test_reference_cases(self):
        for hp in (30, 48):
            for min_period, max_period in ((6, 30), (6, 20), (10, 30), (10, 48)):
                for raster_length in (50, 25):
                    suffix = f"H{hp}_PMIN{min_period}_PMAX{max_period}_R{raster_length}"
                    with self.subTest(case=suffix):
                        params = CoronaTrendVigorParams(
                            raster_length=raster_length,
                            high_pass_filter_cutoff=hp,
                            minimal_period=min_period,
                            maximal_period=max_period,
                        )
                        indicator = CoronaTrendVigor(params)
                        step = 20.0 / raster_length
                        expected_first = -10.0 + step
                        expected_res = 1.0 / step
                        base_time = datetime.datetime(2020, 1, 1)

                        expected_heatmap = getattr(data, f"EXPECTED_{suffix}_HEATMAP")
                        expected_min = getattr(data, f"EXPECTED_{suffix}_HEATMAP_MIN")
                        expected_max = getattr(data, f"EXPECTED_{suffix}_HEATMAP_MAX")
                        expected_tv = getattr(data, f"EXPECTED_{suffix}_TREND_VIGOR")
                        expected_norm = getattr(data, f"EXPECTED_{suffix}_NORMALIZED_VIGOR")
                        expected_dc = getattr(data, f"EXPECTED_{suffix}_DOMINANT_CYCLE")
                        expected_dcm = getattr(data, f"EXPECTED_{suffix}_DOMINANT_CYCLE_MEDIAN")
                        expected_bp = getattr(data, f"EXPECTED_{suffix}_BAND_PASS")
                        expected_quad = getattr(data, f"EXPECTED_{suffix}_QUADRATURE")
                        expected_amp = getattr(data, f"EXPECTED_{suffix}_CYCLE_AMPLITUDE")
                        expected_disp = getattr(data, f"EXPECTED_{suffix}_TREND_DISPLACEMENT")
                        expected_ratio = getattr(data, f"EXPECTED_{suffix}_RAW_RATIO")
                        expected_lag = getattr(data, f"EXPECTED_{suffix}_LOOKBACK")
                        expected_width = getattr(data, f"EXPECTED_{suffix}_WIDTH")
                        expected_has_est = getattr(data, f"EXPECTED_{suffix}_HAS_CYCLE_ESTIMATE")

                        for i, price in enumerate(data.INPUT_CLOSE):
                            bar_time = base_time + datetime.timedelta(days=i)
                            snap = indicator.update(price, bar_time)
                            _assert_close(self, snap.heatmap.parameter_first, expected_first, f"{suffix}.first")
                            _assert_close(self, snap.heatmap.parameter_last, 10.0, f"{suffix}.last")
                            _assert_close(self, snap.heatmap.parameter_resolution, expected_res, f"{suffix}.res")

                            row_is_empty = math.isnan(expected_tv[i])
                            self.assertEqual(indicator.is_primed(), not row_is_empty)
                            if row_is_empty:
                                self.assertTrue(snap.heatmap.is_empty())
                            else:
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
                            _assert_close(self, snap.trend_vigor, expected_tv[i], f"{suffix}[{i}].tv")
                            _assert_close(self, snap.normalized_vigor, expected_norm[i], f"{suffix}[{i}].norm")
                            _assert_close(self, snap.dominant_cycle, expected_dc[i], f"{suffix}[{i}].dc")
                            _assert_close(self, snap.dominant_cycle_median, expected_dcm[i], f"{suffix}[{i}].dcm")
                            _assert_close(self, snap.band_pass, expected_bp[i], f"{suffix}[{i}].bp")
                            _assert_close(self, snap.quadrature, expected_quad[i], f"{suffix}[{i}].quad")
                            _assert_close(self, snap.cycle_amplitude, expected_amp[i], f"{suffix}[{i}].amp")
                            _assert_close(self, snap.trend_displacement, expected_disp[i], f"{suffix}[{i}].disp")
                            _assert_close(self, snap.raw_ratio, expected_ratio[i], f"{suffix}[{i}].ratio")
                            _assert_close(self, snap.lookback, expected_lag[i], f"{suffix}[{i}].lag")
                            _assert_close(self, snap.width, expected_width[i], f"{suffix}[{i}].width")
                            _assert_close(self, snap.has_cycle_estimate, expected_has_est[i], f"{suffix}[{i}].has_est")

                        indicator.reset()
                        self.assertFalse(indicator.is_primed())

    def test_metadata_mnemonics_and_validation(self):
        indicator = CoronaTrendVigor(default_params())
        meta = indicator.metadata()
        self.assertEqual(meta.identifier, Identifier.CORONA_TREND_VIGOR)
        self.assertEqual(meta.mnemonic, "ctv(50, 20, -10, 10, 30)")
        self.assertEqual(len(meta.outputs), 2)
        self.assertEqual(meta.outputs[CoronaTrendVigorOutput.VALUE].shape, Shape.HEATMAP)
        self.assertEqual(meta.outputs[CoronaTrendVigorOutput.TREND_VIGOR].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[CoronaTrendVigorOutput.TREND_VIGOR].mnemonic, "ctv-tv(30)")

        flagged = CoronaTrendVigor(
            CoronaTrendVigorParams(
                minimal_period=10,
                maximal_period=48,
                lookback_offset=-1,
                bar_component=BarComponent.MEDIAN,
            )
        )
        self.assertEqual(
            flagged.metadata().mnemonic,
            "ctv(50, 20, -10, 10, 30, pmin=10, pmax=48, offset=-1, hl/2)",
        )
        self.assertEqual(
            flagged.metadata().outputs[1].mnemonic,
            "ctv-tv(30, pmin=10, pmax=48, offset=-1, hl/2)",
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
            p = CoronaTrendVigorParams(bar_component=bc, quote_component=qc, trade_component=tc)
            inst = CoronaTrendVigor(p)
            resolved_bc = bc if bc is not None else DEFAULT_BAR_COMPONENT
            resolved_qc = qc if qc is not None else DEFAULT_QUOTE_COMPONENT
            resolved_tc = tc if tc is not None else DEFAULT_TRADE_COMPONENT
            suffix = component_triple_mnemonic(resolved_bc, resolved_qc, resolved_tc)
            expected = f"ctv(50, 20, -10, 10, 30{suffix})"
            self.assertEqual(inst.metadata().mnemonic, expected)
            self.assertEqual(corona_trend_vigor_mnemonic(p), expected)

        from_factory = create_indicator(Identifier.CORONA_TREND_VIGOR, {"lookback_offset": -1})
        self.assertEqual(from_factory.metadata().mnemonic, "ctv(50, 20, -10, 10, 30, offset=-1)")

        invalid_cases = (
            CoronaTrendVigorParams(raster_length=1),
            CoronaTrendVigorParams(max_raster_value=6.0),
            CoronaTrendVigorParams(min_parameter_value=0.0, max_parameter_value=0.0),
            CoronaTrendVigorParams(high_pass_filter_cutoff=4),
            CoronaTrendVigorParams(minimal_period=5),
            CoronaTrendVigorParams(minimal_period=10, maximal_period=10),
            CoronaTrendVigorParams(lookback_offset=1),
            CoronaTrendVigorParams(lookback_offset=-2),
        )
        for bad in invalid_cases:
            with self.assertRaises(ValueError):
                CoronaTrendVigor(bad)


if __name__ == "__main__":
    unittest.main()
