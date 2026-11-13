"""Tests for the CoronaSpectrum indicator."""

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
from py.indicators.john_ehlers.corona_spectrum import (
    CoronaSpectrum,
    CoronaSpectrumOutput,
    CoronaSpectrumParams,
    corona_spectrum_mnemonic,
    default_params,
)
from py.indicators.john_ehlers.corona_spectrum import test_testdata as data

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


class TestCoronaSpectrum(unittest.TestCase):
    def test_reference_cases(self):
        for hp in (30, 48):
            for min_period, max_period in ((6, 30), (6, 20), (10, 30), (10, 48)):
                for threshold in (6, 3):
                    suffix = f"H{hp}_PMIN{min_period}_PMAX{max_period}_DB{threshold}"
                    with self.subTest(case=suffix):
                        params = CoronaSpectrumParams(
                            min_raster_value=float(threshold),
                            max_raster_value=20.0,
                            min_parameter_value=float(min_period),
                            max_parameter_value=float(max_period),
                            high_pass_filter_cutoff=hp,
                        )
                        indicator = CoronaSpectrum(params)
                        cells = 2 * (max_period - min_period) + 1
                        base_time = datetime.datetime(2020, 1, 1)

                        expected_heatmap = getattr(data, f"EXPECTED_{suffix}_HEATMAP")
                        expected_min = getattr(data, f"EXPECTED_{suffix}_HEATMAP_MIN")
                        expected_max = getattr(data, f"EXPECTED_{suffix}_HEATMAP_MAX")
                        expected_dc = getattr(data, f"EXPECTED_{suffix}_DOMINANT_CYCLE")
                        expected_dcm = getattr(data, f"EXPECTED_{suffix}_DOMINANT_CYCLE_MEDIAN")
                        expected_max_power = getattr(data, f"EXPECTED_{suffix}_MAXIMUM_POWER")
                        expected_norm_power = getattr(data, f"EXPECTED_{suffix}_NORMALIZATION_POWER")
                        expected_selected = getattr(data, f"EXPECTED_{suffix}_SELECTED_BINS")
                        expected_has_est = getattr(data, f"EXPECTED_{suffix}_HAS_CYCLE_ESTIMATE")
                        expected_power = getattr(data, f"EXPECTED_{suffix}_POWER")
                        expected_rel_power = getattr(data, f"EXPECTED_{suffix}_RELATIVE_POWER")

                        for i, price in enumerate(data.INPUT_CLOSE):
                            bar_time = base_time + datetime.timedelta(days=i)
                            snap = indicator.update(price, bar_time)
                            self.assertEqual(snap.heatmap.parameter_first, float(min_period))
                            self.assertEqual(snap.heatmap.parameter_last, float(max_period))
                            self.assertEqual(snap.heatmap.parameter_resolution, 2.0)

                            if i < 12:
                                self.assertFalse(indicator.is_primed())
                                self.assertTrue(snap.heatmap.is_empty())
                                self.assertEqual(len(snap.power), cells)
                                self.assertTrue(all(math.isnan(v) for v in snap.power))
                                self.assertTrue(all(math.isnan(v) for v in snap.relative_power))
                            else:
                                self.assertTrue(indicator.is_primed())
                                self.assertEqual(len(snap.heatmap.values), cells)
                                for bin_index in range(cells):
                                    _assert_close(
                                        self,
                                        snap.heatmap.values[bin_index],
                                        expected_heatmap[i][bin_index],
                                        f"{suffix}[{i}].heatmap[{bin_index}]",
                                    )
                                    _assert_close(
                                        self,
                                        snap.power[bin_index],
                                        expected_power[i][bin_index],
                                        f"{suffix}[{i}].power[{bin_index}]",
                                    )
                                    _assert_close(
                                        self,
                                        snap.relative_power[bin_index],
                                        expected_rel_power[i][bin_index],
                                        f"{suffix}[{i}].rel_power[{bin_index}]",
                                    )

                            _assert_close(self, snap.heatmap.value_min, expected_min[i], f"{suffix}[{i}].min")
                            _assert_close(self, snap.heatmap.value_max, expected_max[i], f"{suffix}[{i}].max")
                            _assert_close(self, snap.dominant_cycle, expected_dc[i], f"{suffix}[{i}].dc")
                            _assert_close(self, snap.dominant_cycle_median, expected_dcm[i], f"{suffix}[{i}].dcm")
                            _assert_close(self, snap.maximum_power, expected_max_power[i], f"{suffix}[{i}].max_power")
                            _assert_close(self, snap.normalization_power, expected_norm_power[i], f"{suffix}[{i}].norm_power")
                            _assert_close(self, snap.selected_bins, expected_selected[i], f"{suffix}[{i}].selected")
                            _assert_close(self, snap.has_cycle_estimate, expected_has_est[i], f"{suffix}[{i}].has_est")

                        indicator.reset()
                        self.assertFalse(indicator.is_primed())

    def test_metadata_mnemonics_and_entities(self):
        indicator = CoronaSpectrum(default_params())
        meta = indicator.metadata()
        self.assertEqual(meta.identifier, Identifier.CORONA_SPECTRUM)
        self.assertEqual(meta.mnemonic, "cspect(6, 20, 6, 30, 30)")
        self.assertEqual(meta.description, "Corona spectrum cspect(6, 20, 6, 30, 30)")
        self.assertEqual(len(meta.outputs), 3)
        self.assertEqual(meta.outputs[CoronaSpectrumOutput.VALUE].shape, Shape.HEATMAP)
        self.assertEqual(meta.outputs[CoronaSpectrumOutput.VALUE].mnemonic, "cspect(6, 20, 6, 30, 30)")
        self.assertEqual(meta.outputs[CoronaSpectrumOutput.DOMINANT_CYCLE].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[CoronaSpectrumOutput.DOMINANT_CYCLE].mnemonic, "cspect-dc(30)")
        self.assertEqual(meta.outputs[CoronaSpectrumOutput.DOMINANT_CYCLE_MEDIAN].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[CoronaSpectrumOutput.DOMINANT_CYCLE_MEDIAN].mnemonic, "cspect-dcm(30)")

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
            p = CoronaSpectrumParams(bar_component=bc, quote_component=qc, trade_component=tc)
            inst = CoronaSpectrum(p)
            resolved_bc = bc if bc is not None else DEFAULT_BAR_COMPONENT
            resolved_qc = qc if qc is not None else DEFAULT_QUOTE_COMPONENT
            resolved_tc = tc if tc is not None else DEFAULT_TRADE_COMPONENT
            suffix = component_triple_mnemonic(resolved_bc, resolved_qc, resolved_tc)
            expected = f"cspect(6, 20, 6, 30, 30{suffix})"
            self.assertEqual(inst.metadata().mnemonic, expected)
            self.assertEqual(corona_spectrum_mnemonic(p), expected)

        base_time = datetime.datetime(2020, 1, 1)
        bar_ind = CoronaSpectrum()
        scalar_ind = CoronaSpectrum()
        quote_ind = CoronaSpectrum()
        trade_ind = CoronaSpectrum()
        for i in range(15):
            bar_time = base_time + datetime.timedelta(days=i)
            price = data.INPUT_CLOSE[i]
            out_bar = bar_ind.update_bar(
                Bar(bar_time, data.INPUT_OPEN[i], data.INPUT_HIGH[i], data.INPUT_LOW[i], price, 1000.0)
            )
            out_scalar = scalar_ind.update_scalar(Scalar(bar_time, price))
            out_quote = quote_ind.update_quote(Quote(bar_time, price - 0.5, price + 0.5, 10.0, 10.0))
            out_trade = trade_ind.update_trade(Trade(bar_time, price, 10.0))
            self.assertEqual(len(out_bar), 3)
            _assert_close(self, out_bar[1].value, out_scalar[1].value, f"bar vs scalar[{i}]")
            _assert_close(self, out_bar[1].value, out_quote[1].value, f"bar vs quote[{i}]")
            _assert_close(self, out_bar[1].value, out_trade[1].value, f"bar vs trade[{i}]")

        from_factory = create_indicator(Identifier.CORONA_SPECTRUM, {"min_raster_value": 0.0})
        self.assertEqual(from_factory.metadata().mnemonic, "cspect(0, 20, 6, 30, 30)")

    def test_validation(self):
        invalid_cases = (
            CoronaSpectrumParams(min_raster_value=-0.1),
            CoronaSpectrumParams(min_raster_value=20.0, max_raster_value=20.0),
            CoronaSpectrumParams(min_parameter_value=5.0),
            CoronaSpectrumParams(min_parameter_value=0.0),
            CoronaSpectrumParams(min_parameter_value=10.0, max_parameter_value=10.0),
            CoronaSpectrumParams(high_pass_filter_cutoff=4),
            CoronaSpectrumParams(high_pass_filter_cutoff=0),
        )
        for bad in invalid_cases:
            with self.assertRaises(ValueError):
                CoronaSpectrum(bad)


if __name__ == "__main__":
    unittest.main()
