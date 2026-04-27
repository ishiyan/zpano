#!/usr/bin/env python3
"""iconf — command-line chart configuration generator.

Reads a JSON settings file containing indicator definitions,
creates indicator instances via the factory, runs embedded bar data through them,
and writes chart configuration files in JSON and TypeScript formats.

Usage: python -m py.cmd.iconf <settings.json> <output-name>
"""

from __future__ import annotations

import json
import math
import os
import re
import sys
from datetime import datetime, timedelta, timezone

from py.entities.bar import Bar
from py.entities.scalar import Scalar
from py.indicators.core.descriptors import descriptor_of
from py.indicators.core.indicator import Indicator
from py.indicators.core.outputs.band import Band
from py.indicators.core.outputs.heatmap import Heatmap
from py.indicators.core.outputs.shape import Shape
from py.indicators.core.pane import Pane
from py.indicators.factory.factory import create_indicator

# Reuse shared infrastructure from icalc
from py.cmd.icalc.main import _IDENTIFIER_MAP, _convert_params, _test_bars


# ---------------------------------------------------------------------------
# Default colors for cycling
# ---------------------------------------------------------------------------

_LINE_COLORS = [
    "#e41a1c", "#377eb8", "#4daf4a", "#984ea3", "#ff7f00",
    "#a65628", "#f781bf", "#999999", "#66c2a5", "#fc8d62",
]

_BAND_COLORS = [
    "rgba(0,255,0,0.3)", "rgba(0,0,255,0.3)", "rgba(255,0,0,0.3)",
    "rgba(128,0,128,0.3)", "rgba(0,128,128,0.3)",
]

_TIME_FMT = "%Y-%m-%d"


# ---------------------------------------------------------------------------
# Output placement
# ---------------------------------------------------------------------------

class _OutputPlacement:
    __slots__ = ('indicator_idx', 'output_idx', 'mnemonic', 'shape', 'pane')

    def __init__(self, indicator_idx: int, output_idx: int, mnemonic: str,
                 shape: Shape, pane: Pane) -> None:
        self.indicator_idx = indicator_idx
        self.output_idx = output_idx
        self.mnemonic = mnemonic
        self.shape = shape
        self.pane = pane


# ---------------------------------------------------------------------------
# Build chart configuration
# ---------------------------------------------------------------------------

def _build_configuration(indicators: list[Indicator], bars: list[Bar]) -> dict:
    # Collect output placements from descriptor registry.
    placements: list[_OutputPlacement] = []
    for i, ind in enumerate(indicators):
        meta = ind.metadata()
        desc = descriptor_of(meta.identifier)
        if desc is None:
            # Fallback: treat all outputs as Own/Scalar.
            for j, om in enumerate(meta.outputs):
                placements.append(_OutputPlacement(
                    i, j, om.mnemonic, om.shape, Pane.OWN))
            continue
        for j, od in enumerate(desc.outputs):
            mnemonic = f"out[{j}]"
            if j < len(meta.outputs):
                mnemonic = meta.outputs[j].mnemonic
            placements.append(_OutputPlacement(
                i, j, mnemonic, od.shape, od.pane))

    # Run bars through indicators, collecting all outputs.
    # all_outputs[bar_idx][indicator_idx] = output list
    all_outputs: list[list[list]] = []
    for bar in bars:
        row: list[list] = []
        for ind in indicators:
            row.append(ind.update_bar(bar))
        all_outputs.append(row)

    # Build OHLCV data.
    ohlcv_bars = []
    for bar in bars:
        ohlcv_bars.append({
            "time": bar.time.strftime(_TIME_FMT),
            "open": bar.open,
            "high": bar.high,
            "low": bar.low,
            "close": bar.close,
            "volume": bar.volume,
        })

    # Group placements: price pane vs own panes.
    price_pane = {
        "height": "30%", "valueFormat": ",.2f", "valueMarginPercentageFactor": 0.01,
        "bands": [], "lineAreas": [], "horizontals": [], "lines": [], "arrows": [],
    }

    indicator_panes: list[dict] = []
    own_pane_map: dict[int, int] = {}  # indicator_idx -> indicatorPanes index

    color_idx = 0
    band_color_idx = 0

    for p in placements:
        if p.pane == Pane.PRICE and p.shape == Shape.SCALAR:
            line = _build_line_data(p, all_outputs, bars, _LINE_COLORS[color_idx % len(_LINE_COLORS)])
            color_idx += 1
            price_pane["lines"].append(line)

        elif p.pane == Pane.PRICE and p.shape == Shape.BAND:
            band = _build_band_data(p, all_outputs, bars, _BAND_COLORS[band_color_idx % len(_BAND_COLORS)])
            band_color_idx += 1
            price_pane["bands"].append(band)

        else:
            # Own pane (or fallback).
            key = p.indicator_idx
            if key not in own_pane_map:
                own_pane_map[key] = len(indicator_panes)
                indicator_panes.append({
                    "height": "60", "valueFormat": ",.2f", "valueMarginPercentageFactor": 0.01,
                    "bands": [], "lineAreas": [], "horizontals": [], "lines": [], "arrows": [],
                })
            pane = indicator_panes[own_pane_map[key]]

            if p.shape == Shape.HEATMAP:
                hm = _build_heatmap_data(p, all_outputs, bars)
                pane["heatmap"] = hm
                pane["height"] = "120"
                pane["valueMarginPercentageFactor"] = 0

            elif p.shape == Shape.BAND:
                band = _build_band_data(p, all_outputs, bars, _BAND_COLORS[band_color_idx % len(_BAND_COLORS)])
                band_color_idx += 1
                pane["bands"].append(band)

            else:  # Scalar or unknown → line.
                line = _build_line_data(p, all_outputs, bars, _LINE_COLORS[color_idx % len(_LINE_COLORS)])
                color_idx += 1
                pane["lines"].append(line)

    return {
        "width": "100%",
        "navigationPane": {
            "height": 30, "hasLine": True, "hasArea": False, "hasTimeAxis": True, "timeTicks": 0,
        },
        "heightNavigationPane": 30,
        "timeAnnotationFormat": "%Y-%m-%d",
        "axisLeft": True,
        "axisRight": False,
        "margin": {"left": 0, "top": 10, "right": 20, "bottom": 0},
        "ohlcv": {"name": "TEST", "data": ohlcv_bars, "candlesticks": True},
        "pricePane": price_pane,
        "indicatorPanes": indicator_panes,
        "crosshair": False,
        "volumeInPricePane": True,
        "menuVisible": True,
        "downloadSvgVisible": True,
    }


def _build_line_data(p: _OutputPlacement, all_outputs: list[list[list]],
                     bars: list[Bar], color: str) -> dict:
    data = []
    for i, bar in enumerate(bars):
        out = all_outputs[i][p.indicator_idx]
        if p.output_idx < len(out):
            val = out[p.output_idx]
            if isinstance(val, Scalar) and not math.isnan(val.value):
                data.append({"time": val.time.strftime(_TIME_FMT), "value": val.value})
    return {
        "name": p.mnemonic, "data": data, "indicator": p.indicator_idx, "output": p.output_idx,
        "color": color, "width": 1, "dash": "", "interpolation": "natural",
    }


def _build_band_data(p: _OutputPlacement, all_outputs: list[list[list]],
                     bars: list[Bar], color: str) -> dict:
    data = []
    for i, bar in enumerate(bars):
        out = all_outputs[i][p.indicator_idx]
        if p.output_idx < len(out):
            val = out[p.output_idx]
            if isinstance(val, Band) and not val.is_empty():
                data.append({"time": val.time.strftime(_TIME_FMT), "upper": val.upper, "lower": val.lower})
    return {
        "name": p.mnemonic, "data": data, "indicator": p.indicator_idx, "output": p.output_idx,
        "color": color, "legendColor": color, "interpolation": "natural",
    }


def _build_heatmap_data(p: _OutputPlacement, all_outputs: list[list[list]],
                        bars: list[Bar]) -> dict:
    data = []
    for i, bar in enumerate(bars):
        out = all_outputs[i][p.indicator_idx]
        if p.output_idx < len(out):
            val = out[p.output_idx]
            if isinstance(val, Heatmap) and not val.is_empty():
                data.append({
                    "time": val.time.strftime(_TIME_FMT),
                    "first": val.parameter_first, "last": val.parameter_last,
                    "result": val.parameter_resolution,
                    "min": val.value_min, "max": val.value_max,
                    "values": val.values,
                })
    return {
        "name": p.mnemonic, "data": data, "indicator": p.indicator_idx, "output": p.output_idx,
        "gradient": "Viridis", "invertGradient": False,
    }


# ---------------------------------------------------------------------------
# TypeScript generation
# ---------------------------------------------------------------------------

def _sanitize_var_name(s: str) -> str:
    """Convert a string to a valid camelCase JS variable name."""
    result = []
    capitalize = False
    for i, c in enumerate(s):
        if c.isalpha():
            if capitalize:
                c = c.upper()
                capitalize = False
            result.append(c)
        elif c.isdigit():
            if i == 0:
                result.append('_')
            result.append(c)
            capitalize = False
        else:
            capitalize = True
    return ''.join(result)


def _json_compact(obj) -> str:
    """Serialize to compact JSON (no spaces)."""
    return json.dumps(obj, separators=(',', ':'))


def _build_typescript(cfg: dict, base_name: str) -> str:
    lines: list[str] = []
    w = lines.append

    w("// Auto-generated chart configuration.")
    w("// eslint-disable-next-line")
    w("import { Configuration } from '../ohlcv-chart/template/configuration';")
    w("import { Scalar, Band, Heatmap, Bar } from '../ohlcv-chart/template/types';")
    w("")

    # Emit OHLCV data array.
    ohlcv_var = _sanitize_var_name(base_name) + "Ohlcv"
    w(f"export const {ohlcv_var}: Bar[] = {_json_compact(cfg['ohlcv']['data'])};")
    w("")

    # Track variable names for data references.
    # We'll use (pane_type, pane_idx, data_type, item_idx) -> var_name
    var_refs: dict[str, str] = {}
    var_counter = 0

    def emit_scalar_array(mnemonic: str, data: list) -> str:
        nonlocal var_counter
        var_counter += 1
        vn = _sanitize_var_name(f"{base_name}_{mnemonic}_{var_counter}")
        w(f"export const {vn}: Scalar[] = {_json_compact(data)};")
        w("")
        return vn

    def emit_band_array(mnemonic: str, data: list) -> str:
        nonlocal var_counter
        var_counter += 1
        vn = _sanitize_var_name(f"{base_name}_{mnemonic}_{var_counter}")
        w(f"export const {vn}: Band[] = {_json_compact(data)};")
        w("")
        return vn

    def emit_heatmap_array(mnemonic: str, data: list) -> str:
        nonlocal var_counter
        var_counter += 1
        vn = _sanitize_var_name(f"{base_name}_{mnemonic}_{var_counter}")
        w(f"export const {vn}: Heatmap[] = {_json_compact(data)};")
        w("")
        return vn

    # Emit price pane data arrays and track refs.
    price_line_vars: list[str] = []
    for line in cfg["pricePane"]["lines"]:
        vn = emit_scalar_array(line["name"], line["data"])
        price_line_vars.append(vn)

    price_band_vars: list[str] = []
    for band in cfg["pricePane"]["bands"]:
        vn = emit_band_array(band["name"], band["data"])
        price_band_vars.append(vn)

    price_linearea_vars: list[str] = []
    for la in cfg["pricePane"]["lineAreas"]:
        vn = emit_scalar_array(la["name"], la["data"])
        price_linearea_vars.append(vn)

    # Emit indicator pane data arrays.
    ind_line_vars: list[list[str]] = []
    ind_band_vars: list[list[str]] = []
    ind_linearea_vars: list[list[str]] = []
    ind_heatmap_vars: list[str | None] = []

    for pane in cfg["indicatorPanes"]:
        lvars: list[str] = []
        for line in pane["lines"]:
            vn = emit_scalar_array(line["name"], line["data"])
            lvars.append(vn)
        ind_line_vars.append(lvars)

        bvars: list[str] = []
        for band in pane["bands"]:
            vn = emit_band_array(band["name"], band["data"])
            bvars.append(vn)
        ind_band_vars.append(bvars)

        lavars: list[str] = []
        for la in pane["lineAreas"]:
            vn = emit_scalar_array(la["name"], la["data"])
            lavars.append(vn)
        ind_linearea_vars.append(lavars)

        hm_var = None
        if "heatmap" in pane:
            hm = pane["heatmap"]
            hm_var = emit_heatmap_array(hm["name"], hm["data"])
        ind_heatmap_vars.append(hm_var)

    # Emit configuration object.
    config_var = _sanitize_var_name(base_name) + "Config"
    w(f"export const {config_var}: Configuration = {{")
    w(f'  width: "{cfg["width"]}",')

    nav = cfg.get("navigationPane")
    if nav:
        w("  navigationPane: {")
        w(f"    height: {nav['height']}, hasLine: {_ts_bool(nav['hasLine'])}, "
          f"hasArea: {_ts_bool(nav['hasArea'])}, hasTimeAxis: {_ts_bool(nav['hasTimeAxis'])}, "
          f"timeTicks: {nav['timeTicks']},")
        w("  },")

    w(f"  heightNavigationPane: {cfg['heightNavigationPane']},")
    w(f'  timeAnnotationFormat: "{cfg["timeAnnotationFormat"]}",')
    w(f"  axisLeft: {_ts_bool(cfg['axisLeft'])},")
    w(f"  axisRight: {_ts_bool(cfg['axisRight'])},")
    m = cfg["margin"]
    w(f"  margin: {{ left: {m['left']}, top: {m['top']}, right: {m['right']}, bottom: {m['bottom']} }},")
    w(f'  ohlcv: {{ name: "{cfg["ohlcv"]["name"]}", data: {ohlcv_var}, candlesticks: {_ts_bool(cfg["ohlcv"]["candlesticks"])} }},')

    # Price pane.
    w("  pricePane: {")
    _write_ts_pane(lines, cfg["pricePane"], price_line_vars, price_band_vars, price_linearea_vars, None, "    ")
    w("  },")

    # Indicator panes.
    w("  indicatorPanes: [")
    for pi, pane in enumerate(cfg["indicatorPanes"]):
        w("    {")
        _write_ts_pane(lines, pane, ind_line_vars[pi], ind_band_vars[pi], ind_linearea_vars[pi], ind_heatmap_vars[pi], "      ")
        w("    },")
    w("  ],")

    w(f"  crosshair: {_ts_bool(cfg['crosshair'])},")
    w(f"  volumeInPricePane: {_ts_bool(cfg['volumeInPricePane'])},")
    w(f"  menuVisible: {_ts_bool(cfg['menuVisible'])},")
    w(f"  downloadSvgVisible: {_ts_bool(cfg['downloadSvgVisible'])},")
    w("};")

    return '\n'.join(lines) + '\n'


def _ts_bool(val: bool) -> str:
    return "true" if val else "false"


def _ts_num(val: float) -> str:
    if val == int(val):
        return str(int(val))
    return str(val)


def _write_ts_pane(lines: list[str], pane: dict, line_vars: list[str],
                   band_vars: list[str], linearea_vars: list[str],
                   heatmap_var: str | None, indent: str) -> None:
    w = lines.append
    w(f'{indent}height: "{pane["height"]}", valueFormat: "{pane["valueFormat"]}", '
      f'valueMarginPercentageFactor: {_ts_num(pane["valueMarginPercentageFactor"])},')

    if heatmap_var and "heatmap" in pane:
        hm = pane["heatmap"]
        w(f'{indent}heatmap: {{ name: "{hm["name"]}", data: {heatmap_var}, '
          f'indicator: {hm["indicator"]}, output: {hm["output"]}, '
          f'gradient: "{hm["gradient"]}", invertGradient: {_ts_bool(hm["invertGradient"])} }},')

    # Bands.
    if pane["bands"]:
        w(f"{indent}bands: [")
        for i, b in enumerate(pane["bands"]):
            vn = band_vars[i]
            w(f'{indent}  {{ name: "{b["name"]}", data: {vn}, indicator: {b["indicator"]}, '
              f'output: {b["output"]}, color: "{b["color"]}", legendColor: "{b["legendColor"]}", '
              f'interpolation: "{b["interpolation"]}" }},')
        w(f"{indent}],")
    else:
        w(f"{indent}bands: [],")

    # LineAreas.
    if pane["lineAreas"]:
        w(f"{indent}lineAreas: [")
        for i, la in enumerate(pane["lineAreas"]):
            vn = linearea_vars[i]
            w(f'{indent}  {{ name: "{la["name"]}", data: {vn}, indicator: {la["indicator"]}, '
              f'output: {la["output"]}, value: {_ts_num(la["value"])}, color: "{la["color"]}", '
              f'legendColor: "{la["legendColor"]}", interpolation: "{la["interpolation"]}" }},')
        w(f"{indent}],")
    else:
        w(f"{indent}lineAreas: [],")

    # Horizontals.
    if pane["horizontals"]:
        w(f"{indent}horizontals: [")
        for h in pane["horizontals"]:
            w(f'{indent}  {{ value: {_ts_num(h["value"])}, color: "{h["color"]}", '
              f'width: {_ts_num(h["width"])}, dash: "{h["dash"]}" }},')
        w(f"{indent}],")
    else:
        w(f"{indent}horizontals: [],")

    # Lines.
    if pane["lines"]:
        w(f"{indent}lines: [")
        for i, l in enumerate(pane["lines"]):
            vn = line_vars[i]
            w(f'{indent}  {{ name: "{l["name"]}", data: {vn}, indicator: {l["indicator"]}, '
              f'output: {l["output"]}, color: "{l["color"]}", width: {_ts_num(l["width"])}, '
              f'dash: "{l["dash"]}", interpolation: "{l["interpolation"]}" }},')
        w(f"{indent}],")
    else:
        w(f"{indent}lines: [],")

    # Arrows.
    w(f"{indent}arrows: [],")


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    if len(sys.argv) < 3:
        print("usage: python -m py.cmd.iconf <settings.json> <output-name>", file=sys.stderr)
        sys.exit(1)

    settings_path = sys.argv[1]
    output_name = sys.argv[2]
    # Strip .json or .ts suffix
    if output_name.endswith('.json'):
        output_name = output_name[:-5]
    if output_name.endswith('.ts'):
        output_name = output_name[:-3]

    with open(settings_path) as f:
        entries = json.load(f)

    indicators: list[Indicator] = []
    for e in entries:
        id_str = e['identifier']
        ident = _IDENTIFIER_MAP.get(id_str)
        if ident is None:
            print(f"error: unknown indicator identifier: {id_str}", file=sys.stderr)
            sys.exit(1)

        params = _convert_params(e.get('params', {}))
        ind = create_indicator(ident, params if params else None)
        indicators.append(ind)

    bars = _test_bars()
    cfg = _build_configuration(indicators, bars)

    # Write JSON.
    json_path = output_name + ".json"
    with open(json_path, 'w') as f:
        json.dump(cfg, f, indent=2)
    print(f"wrote {json_path}")

    # Write TypeScript.
    ts_path = output_name + ".ts"
    base_name = os.path.basename(output_name)
    ts_data = _build_typescript(cfg, base_name)
    with open(ts_path, 'w') as f:
        f.write(ts_data)
    print(f"wrote {ts_path}")


if __name__ == '__main__':
    main()
