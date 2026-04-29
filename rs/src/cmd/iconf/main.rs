/// iconf — command-line chart configuration generator.
///
/// Reads a JSON settings file containing indicator definitions,
/// creates indicator instances via the factory, runs embedded bar data through them,
/// and writes chart configuration files in JSON and TypeScript formats.
///
/// Usage: iconf <settings.json> <output-name>

use portf::entities::bar::Bar;
use portf::entities::scalar::Scalar;
use portf::indicators::core::descriptor::descriptor_of;
use portf::indicators::core::identifier::Identifier;
use portf::indicators::core::indicator::{Indicator, Output};
use portf::indicators::core::outputs::band::Band;
use portf::indicators::core::outputs::heatmap::Heatmap;
use portf::indicators::core::outputs::shape::Shape;
use portf::indicators::core::pane::Pane;
use portf::indicators::factory::{create_indicator, JsonValue};

use std::env;
use std::fs;
use std::process;

// ---------------------------------------------------------------------------
// Colors
// ---------------------------------------------------------------------------

const LINE_COLORS: &[&str] = &[
    "#e41a1c", "#377eb8", "#4daf4a", "#984ea3", "#ff7f00",
    "#a65628", "#f781bf", "#999999", "#66c2a5", "#fc8d62",
];

const BAND_COLORS: &[&str] = &[
    "rgba(0,255,0,0.3)", "rgba(0,0,255,0.3)", "rgba(255,0,0,0.3)",
    "rgba(128,0,128,0.3)", "rgba(0,128,128,0.3)",
];

// ---------------------------------------------------------------------------
// Output placement
// ---------------------------------------------------------------------------

struct OutputPlacement {
    indicator_idx: usize,
    output_idx: usize,
    mnemonic: String,
    shape: Shape,
    pane: Pane,
}

// ---------------------------------------------------------------------------
// Simple JSON builder (no serde)
// ---------------------------------------------------------------------------

/// Minimal JSON value type for building configuration output.
#[derive(Clone)]
enum JVal {
    Null,
    Bool(bool),
    Int(i64),
    Float(f64),
    Str(String),
    Array(Vec<JVal>),
    Object(Vec<(String, JVal)>),
}

impl JVal {
    fn to_json(&self, indent: usize, current: usize) -> String {
        match self {
            JVal::Null => "null".to_string(),
            JVal::Bool(b) => if *b { "true" } else { "false" }.to_string(),
            JVal::Int(n) => format!("{}", n),
            JVal::Float(f) => format_float(*f),
            JVal::Str(s) => format!("\"{}\"", escape_json(s)),
            JVal::Array(arr) => {
                if arr.is_empty() {
                    return "[]".to_string();
                }
                let inner_indent = current + indent;
                let pad_inner = " ".repeat(inner_indent);
                let pad_close = " ".repeat(current);
                let mut out = "[\n".to_string();
                for (i, v) in arr.iter().enumerate() {
                    out.push_str(&pad_inner);
                    out.push_str(&v.to_json(indent, inner_indent));
                    if i < arr.len() - 1 {
                        out.push(',');
                    }
                    out.push('\n');
                }
                out.push_str(&pad_close);
                out.push(']');
                out
            }
            JVal::Object(pairs) => {
                if pairs.is_empty() {
                    return "{}".to_string();
                }
                let inner_indent = current + indent;
                let pad_inner = " ".repeat(inner_indent);
                let pad_close = " ".repeat(current);
                let mut out = "{\n".to_string();
                for (i, (k, v)) in pairs.iter().enumerate() {
                    out.push_str(&pad_inner);
                    out.push_str(&format!("\"{}\": ", escape_json(k)));
                    out.push_str(&v.to_json(indent, inner_indent));
                    if i < pairs.len() - 1 {
                        out.push(',');
                    }
                    out.push('\n');
                }
                out.push_str(&pad_close);
                out.push('}');
                out
            }
        }
    }

    fn to_compact_json(&self) -> String {
        match self {
            JVal::Null => "null".to_string(),
            JVal::Bool(b) => if *b { "true" } else { "false" }.to_string(),
            JVal::Int(n) => format!("{}", n),
            JVal::Float(f) => format_float(*f),
            JVal::Str(s) => format!("\"{}\"", escape_json(s)),
            JVal::Array(arr) => {
                let items: Vec<String> = arr.iter().map(|v| v.to_compact_json()).collect();
                format!("[{}]", items.join(","))
            }
            JVal::Object(pairs) => {
                let items: Vec<String> = pairs.iter()
                    .map(|(k, v)| format!("\"{}\":{}", escape_json(k), v.to_compact_json()))
                    .collect();
                format!("{{{}}}", items.join(","))
            }
        }
    }
}

fn escape_json(s: &str) -> String {
    let mut out = String::with_capacity(s.len());
    for c in s.chars() {
        match c {
            '"' => out.push_str("\\\""),
            '\\' => out.push_str("\\\\"),
            '\n' => out.push_str("\\n"),
            '\r' => out.push_str("\\r"),
            '\t' => out.push_str("\\t"),
            c if (c as u32) < 0x20 => out.push_str(&format!("\\u{:04x}", c as u32)),
            _ => out.push(c),
        }
    }
    out
}

fn format_float(f: f64) -> String {
    if f == (f as i64) as f64 && f.abs() < 1e15 {
        format!("{}", f as i64)
    } else {
        format!("{}", f)
    }
}

fn jobj(pairs: Vec<(&str, JVal)>) -> JVal {
    JVal::Object(pairs.into_iter().map(|(k, v)| (k.to_string(), v)).collect())
}

// ---------------------------------------------------------------------------
// Date formatting (same algorithm as icalc)
// ---------------------------------------------------------------------------

fn epoch_to_date_string(epoch: i64) -> String {
    let days = epoch / 86400;
    let z = days + 719468;
    let era = if z >= 0 { z } else { z - 146096 } / 146097;
    let doe = (z - era * 146097) as u32;
    let yoe = (doe - doe / 1460 + doe / 36524 - doe / 146096) / 365;
    let y = (yoe as i64) + era * 400;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = if mp < 10 { mp + 3 } else { mp - 9 };
    let y = if m <= 2 { y + 1 } else { y };
    format!("{:04}-{:02}-{:02}", y, m, d)
}

// ---------------------------------------------------------------------------
// Build configuration
// ---------------------------------------------------------------------------

fn build_configuration(indicators: &mut [Box<dyn Indicator>], bars: &[Bar]) -> JVal {
    // Collect output placements from descriptor registry.
    let mut placements: Vec<OutputPlacement> = Vec::new();
    for (i, ind) in indicators.iter().enumerate() {
        let meta = ind.metadata();
        match descriptor_of(meta.identifier) {
            None => {
                for (j, om) in meta.outputs.iter().enumerate() {
                    placements.push(OutputPlacement {
                        indicator_idx: i,
                        output_idx: j,
                        mnemonic: om.mnemonic.clone(),
                        shape: om.shape,
                        pane: Pane::Own,
                    });
                }
            }
            Some(desc) => {
                for (j, od) in desc.outputs.iter().enumerate() {
                    let mnemonic = if j < meta.outputs.len() {
                        meta.outputs[j].mnemonic.clone()
                    } else {
                        format!("out[{}]", j)
                    };
                    placements.push(OutputPlacement {
                        indicator_idx: i,
                        output_idx: j,
                        mnemonic,
                        shape: od.shape,
                        pane: od.pane,
                    });
                }
            }
        }
    }

    // Run bars through indicators, collecting all outputs.
    // all_outputs[bar_idx][indicator_idx] = output list
    let mut all_outputs: Vec<Vec<Output>> = Vec::with_capacity(bars.len());
    for bar in bars {
        let mut row: Vec<Output> = Vec::with_capacity(indicators.len());
        for ind in indicators.iter_mut() {
            row.push(ind.update_bar(bar));
        }
        all_outputs.push(row);
    }

    // Build OHLCV data.
    let ohlcv_data: Vec<JVal> = bars.iter().map(|bar| {
        jobj(vec![
            ("time", JVal::Str(epoch_to_date_string(bar.time))),
            ("open", JVal::Float(bar.open)),
            ("high", JVal::Float(bar.high)),
            ("low", JVal::Float(bar.low)),
            ("close", JVal::Float(bar.close)),
            ("volume", JVal::Float(bar.volume)),
        ])
    }).collect();

    // Group placements: price pane vs own panes.
    let mut price_lines: Vec<JVal> = Vec::new();
    let mut price_bands: Vec<JVal> = Vec::new();
    let mut indicator_panes: Vec<JVal> = Vec::new();
    let mut own_pane_map: Vec<(usize, usize)> = Vec::new(); // (indicator_idx, panes_index)

    let mut color_idx = 0usize;
    let mut band_color_idx = 0usize;

    for p in &placements {
        if p.pane == Pane::Price && p.shape == Shape::Scalar {
            let line = build_line_data(p, &all_outputs, bars, LINE_COLORS[color_idx % LINE_COLORS.len()]);
            color_idx += 1;
            price_lines.push(line);
        } else if p.pane == Pane::Price && p.shape == Shape::Band {
            let band = build_band_data(p, &all_outputs, bars, BAND_COLORS[band_color_idx % BAND_COLORS.len()]);
            band_color_idx += 1;
            price_bands.push(band);
        } else {
            // Own pane (or fallback).
            let pane_idx = match own_pane_map.iter().find(|(k, _)| *k == p.indicator_idx) {
                Some((_, idx)) => *idx,
                None => {
                    let idx = indicator_panes.len();
                    own_pane_map.push((p.indicator_idx, idx));
                    // Push a placeholder — we'll replace it at the end.
                    indicator_panes.push(JVal::Null);
                    idx
                }
            };

            // We need to build panes incrementally. Use a side structure.
            // Actually, let's collect into intermediate structures first, then build JVal at the end.
            let _ = pane_idx; // handled below
        }
    }

    // Redo own pane building with proper intermediate structures.
    struct PaneData {
        height: String,
        value_format: String,
        value_margin_percentage_factor: f64,
        bands: Vec<JVal>,
        lines: Vec<JVal>,
        heatmap: Option<JVal>,
    }

    let mut own_panes: Vec<PaneData> = Vec::new();
    let mut own_pane_map2: Vec<(usize, usize)> = Vec::new();
    color_idx = 0;
    band_color_idx = 0;
    price_lines.clear();
    price_bands.clear();

    for p in &placements {
        if p.pane == Pane::Price && p.shape == Shape::Scalar {
            let line = build_line_data(p, &all_outputs, bars, LINE_COLORS[color_idx % LINE_COLORS.len()]);
            color_idx += 1;
            price_lines.push(line);
        } else if p.pane == Pane::Price && p.shape == Shape::Band {
            let band = build_band_data(p, &all_outputs, bars, BAND_COLORS[band_color_idx % BAND_COLORS.len()]);
            band_color_idx += 1;
            price_bands.push(band);
        } else {
            let pane_idx = match own_pane_map2.iter().find(|(k, _)| *k == p.indicator_idx) {
                Some((_, idx)) => *idx,
                None => {
                    let idx = own_panes.len();
                    own_pane_map2.push((p.indicator_idx, idx));
                    own_panes.push(PaneData {
                        height: "60".to_string(),
                        value_format: ",.2f".to_string(),
                        value_margin_percentage_factor: 0.01,
                        bands: Vec::new(),
                        lines: Vec::new(),
                        heatmap: None,
                    });
                    idx
                }
            };

            let pane = &mut own_panes[pane_idx];

            if p.shape == Shape::Heatmap {
                let hm = build_heatmap_data(p, &all_outputs, bars);
                pane.heatmap = Some(hm);
                pane.height = "120".to_string();
                pane.value_margin_percentage_factor = 0.0;
            } else if p.shape == Shape::Band {
                let band = build_band_data(p, &all_outputs, bars, BAND_COLORS[band_color_idx % BAND_COLORS.len()]);
                band_color_idx += 1;
                pane.bands.push(band);
            } else {
                // Scalar or unknown → line.
                let line = build_line_data(p, &all_outputs, bars, LINE_COLORS[color_idx % LINE_COLORS.len()]);
                color_idx += 1;
                pane.lines.push(line);
            }
        }
    }

    // Convert own panes to JVal.
    let indicator_panes_jval: Vec<JVal> = own_panes.into_iter().map(|pd| {
        let mut pairs: Vec<(&str, JVal)> = vec![
            ("height", JVal::Str(pd.height)),
            ("valueFormat", JVal::Str(pd.value_format)),
            ("valueMarginPercentageFactor", JVal::Float(pd.value_margin_percentage_factor)),
        ];
        if let Some(hm) = pd.heatmap {
            pairs.push(("heatmap", hm));
        }
        pairs.push(("bands", JVal::Array(pd.bands)));
        pairs.push(("lineAreas", JVal::Array(Vec::new())));
        pairs.push(("horizontals", JVal::Array(Vec::new())));
        pairs.push(("lines", JVal::Array(pd.lines)));
        pairs.push(("arrows", JVal::Array(Vec::new())));
        jobj(pairs)
    }).collect();

    let price_pane = jobj(vec![
        ("height", JVal::Str("30%".to_string())),
        ("valueFormat", JVal::Str(",.2f".to_string())),
        ("valueMarginPercentageFactor", JVal::Float(0.01)),
        ("bands", JVal::Array(price_bands)),
        ("lineAreas", JVal::Array(Vec::new())),
        ("horizontals", JVal::Array(Vec::new())),
        ("lines", JVal::Array(price_lines)),
        ("arrows", JVal::Array(Vec::new())),
    ]);

    jobj(vec![
        ("width", JVal::Str("100%".to_string())),
        ("navigationPane", jobj(vec![
            ("height", JVal::Int(30)),
            ("hasLine", JVal::Bool(true)),
            ("hasArea", JVal::Bool(false)),
            ("hasTimeAxis", JVal::Bool(true)),
            ("timeTicks", JVal::Int(0)),
        ])),
        ("heightNavigationPane", JVal::Int(30)),
        ("timeAnnotationFormat", JVal::Str("%Y-%m-%d".to_string())),
        ("axisLeft", JVal::Bool(true)),
        ("axisRight", JVal::Bool(false)),
        ("margin", jobj(vec![
            ("left", JVal::Int(0)),
            ("top", JVal::Int(10)),
            ("right", JVal::Int(20)),
            ("bottom", JVal::Int(0)),
        ])),
        ("ohlcv", jobj(vec![
            ("name", JVal::Str("TEST".to_string())),
            ("data", JVal::Array(ohlcv_data)),
            ("candlesticks", JVal::Bool(true)),
        ])),
        ("pricePane", price_pane),
        ("indicatorPanes", JVal::Array(indicator_panes_jval)),
        ("crosshair", JVal::Bool(false)),
        ("volumeInPricePane", JVal::Bool(true)),
        ("menuVisible", JVal::Bool(true)),
        ("downloadSvgVisible", JVal::Bool(true)),
    ])
}

// ---------------------------------------------------------------------------
// Build line/band/heatmap data
// ---------------------------------------------------------------------------

fn build_line_data(p: &OutputPlacement, all_outputs: &[Vec<Output>], bars: &[Bar], color: &str) -> JVal {
    let mut data: Vec<JVal> = Vec::new();
    for (i, _bar) in bars.iter().enumerate() {
        let out = &all_outputs[i][p.indicator_idx];
        if p.output_idx < out.len() {
            if let Some(s) = out[p.output_idx].downcast_ref::<Scalar>() {
                if !s.value.is_nan() {
                    data.push(jobj(vec![
                        ("time", JVal::Str(epoch_to_date_string(s.time))),
                        ("value", JVal::Float(s.value)),
                    ]));
                }
            }
        }
    }
    jobj(vec![
        ("name", JVal::Str(p.mnemonic.clone())),
        ("data", JVal::Array(data)),
        ("indicator", JVal::Int(p.indicator_idx as i64)),
        ("output", JVal::Int(p.output_idx as i64)),
        ("color", JVal::Str(color.to_string())),
        ("width", JVal::Int(1)),
        ("dash", JVal::Str(String::new())),
        ("interpolation", JVal::Str("natural".to_string())),
    ])
}

fn build_band_data(p: &OutputPlacement, all_outputs: &[Vec<Output>], bars: &[Bar], color: &str) -> JVal {
    let mut data: Vec<JVal> = Vec::new();
    for (i, _bar) in bars.iter().enumerate() {
        let out = &all_outputs[i][p.indicator_idx];
        if p.output_idx < out.len() {
            if let Some(b) = out[p.output_idx].downcast_ref::<Band>() {
                if !b.is_empty() {
                    data.push(jobj(vec![
                        ("time", JVal::Str(epoch_to_date_string(b.time))),
                        ("upper", JVal::Float(b.upper)),
                        ("lower", JVal::Float(b.lower)),
                    ]));
                }
            }
        }
    }
    jobj(vec![
        ("name", JVal::Str(p.mnemonic.clone())),
        ("data", JVal::Array(data)),
        ("indicator", JVal::Int(p.indicator_idx as i64)),
        ("output", JVal::Int(p.output_idx as i64)),
        ("color", JVal::Str(color.to_string())),
        ("legendColor", JVal::Str(color.to_string())),
        ("interpolation", JVal::Str("natural".to_string())),
    ])
}

fn build_heatmap_data(p: &OutputPlacement, all_outputs: &[Vec<Output>], bars: &[Bar]) -> JVal {
    let mut data: Vec<JVal> = Vec::new();
    for (i, _bar) in bars.iter().enumerate() {
        let out = &all_outputs[i][p.indicator_idx];
        if p.output_idx < out.len() {
            if let Some(hm) = out[p.output_idx].downcast_ref::<Heatmap>() {
                if !hm.is_empty() {
                    let values_jval: Vec<JVal> = hm.values.iter().map(|v| JVal::Float(*v)).collect();
                    data.push(jobj(vec![
                        ("time", JVal::Str(epoch_to_date_string(hm.time))),
                        ("first", JVal::Float(hm.parameter_first)),
                        ("last", JVal::Float(hm.parameter_last)),
                        ("result", JVal::Float(hm.parameter_resolution)),
                        ("min", JVal::Float(hm.value_min)),
                        ("max", JVal::Float(hm.value_max)),
                        ("values", JVal::Array(values_jval)),
                    ]));
                }
            }
        }
    }
    jobj(vec![
        ("name", JVal::Str(p.mnemonic.clone())),
        ("data", JVal::Array(data)),
        ("indicator", JVal::Int(p.indicator_idx as i64)),
        ("output", JVal::Int(p.output_idx as i64)),
        ("gradient", JVal::Str("Viridis".to_string())),
        ("invertGradient", JVal::Bool(false)),
    ])
}

// ---------------------------------------------------------------------------
// TypeScript generation
// ---------------------------------------------------------------------------

fn sanitize_var_name(s: &str) -> String {
    let mut result = String::new();
    let mut capitalize = false;
    for (i, c) in s.chars().enumerate() {
        if c.is_ascii_alphabetic() {
            if capitalize {
                result.push(c.to_ascii_uppercase());
                capitalize = false;
            } else {
                result.push(c);
            }
        } else if c.is_ascii_digit() {
            if i == 0 {
                result.push('_');
            }
            result.push(c);
            capitalize = false;
        } else {
            capitalize = true;
        }
    }
    result
}

fn ts_bool(val: bool) -> &'static str {
    if val { "true" } else { "false" }
}

fn ts_num(val: f64) -> String {
    if val == (val as i64) as f64 && val.abs() < 1e15 {
        format!("{}", val as i64)
    } else {
        format!("{}", val)
    }
}

fn build_typescript(cfg: &JVal, base_name: &str) -> String {
    // We need to traverse the JVal tree. Extract what we need via helper methods.
    let cfg_pairs = match cfg {
        JVal::Object(p) => p,
        _ => return String::new(),
    };

    let get = |key: &str| -> &JVal {
        cfg_pairs.iter().find(|(k, _)| k == key).map(|(_, v)| v).unwrap_or(&JVal::Null)
    };

    let get_str = |jv: &JVal| -> String {
        if let JVal::Str(s) = jv { s.clone() } else { String::new() }
    };

    let get_bool = |jv: &JVal| -> bool {
        if let JVal::Bool(b) = jv { *b } else { false }
    };

    let get_int = |jv: &JVal| -> i64 {
        if let JVal::Int(n) = jv { *n } else { 0 }
    };

    let get_float = |jv: &JVal| -> f64 {
        match jv {
            JVal::Float(f) => *f,
            JVal::Int(n) => *n as f64,
            _ => 0.0,
        }
    };

    let get_obj_field = |obj: &JVal, key: &str| -> JVal {
        if let JVal::Object(pairs) = obj {
            pairs.iter().find(|(k, _)| k == key).map(|(_, v)| v.clone()).unwrap_or(JVal::Null)
        } else {
            JVal::Null
        }
    };

    fn get_array_ref(jv: &JVal) -> &[JVal] {
        if let JVal::Array(arr) = jv { arr } else { &[] }
    }

    let mut lines: Vec<String> = Vec::new();
    let mut var_counter = 0usize;

    lines.push("// Auto-generated chart configuration.".to_string());
    lines.push("// eslint-disable-next-line".to_string());
    lines.push("import { Configuration } from '../ohlcv-chart/template/configuration';".to_string());
    lines.push("import { Scalar, Band, Heatmap, Bar } from '../ohlcv-chart/template/types';".to_string());
    lines.push(String::new());

    // Emit OHLCV data.
    let ohlcv_var = format!("{}Ohlcv", sanitize_var_name(base_name));
    let ohlcv = get("ohlcv");
    let ohlcv_data = get_obj_field(ohlcv, "data");
    lines.push(format!("export const {}: Bar[] = {};", ohlcv_var, ohlcv_data.to_compact_json()));
    lines.push(String::new());

    let mut emit_data_array = |type_name: &str, mnemonic: &str, data: &JVal| -> String {
        var_counter += 1;
        let vn = sanitize_var_name(&format!("{}_{}_{}", base_name, mnemonic, var_counter));
        lines.push(format!("export const {}: {}[] = {};", vn, type_name, data.to_compact_json()));
        lines.push(String::new());
        vn
    };

    // Emit price pane data arrays.
    let price_pane = get("pricePane");
    let price_lines_arr = get_obj_field(price_pane, "lines");
    let price_bands_arr = get_obj_field(price_pane, "bands");
    let price_line_areas_arr = get_obj_field(price_pane, "lineAreas");

    let mut price_line_vars: Vec<String> = Vec::new();
    for line in get_array_ref(&price_lines_arr) {
        let name = get_str(&get_obj_field(line, "name"));
        let data = get_obj_field(line, "data");
        let vn = emit_data_array("Scalar", &name, &data);
        price_line_vars.push(vn);
    }

    let mut price_band_vars: Vec<String> = Vec::new();
    for band in get_array_ref(&price_bands_arr) {
        let name = get_str(&get_obj_field(band, "name"));
        let data = get_obj_field(band, "data");
        let vn = emit_data_array("Band", &name, &data);
        price_band_vars.push(vn);
    }

    let mut price_linearea_vars: Vec<String> = Vec::new();
    for la in get_array_ref(&price_line_areas_arr) {
        let name = get_str(&get_obj_field(la, "name"));
        let data = get_obj_field(la, "data");
        let vn = emit_data_array("Scalar", &name, &data);
        price_linearea_vars.push(vn);
    }

    // Emit indicator pane data arrays.
    let ind_panes_arr = get("indicatorPanes");
    let ind_panes = get_array_ref(ind_panes_arr);

    let mut ind_line_vars: Vec<Vec<String>> = Vec::new();
    let mut ind_band_vars: Vec<Vec<String>> = Vec::new();
    let mut ind_linearea_vars: Vec<Vec<String>> = Vec::new();
    let mut ind_heatmap_vars: Vec<Option<String>> = Vec::new();

    for pane in ind_panes {
        let pane_lines = get_obj_field(pane, "lines");
        let mut lvars: Vec<String> = Vec::new();
        for line in get_array_ref(&pane_lines) {
            let name = get_str(&get_obj_field(line, "name"));
            let data = get_obj_field(line, "data");
            let vn = emit_data_array("Scalar", &name, &data);
            lvars.push(vn);
        }
        ind_line_vars.push(lvars);

        let pane_bands = get_obj_field(pane, "bands");
        let mut bvars: Vec<String> = Vec::new();
        for band in get_array_ref(&pane_bands) {
            let name = get_str(&get_obj_field(band, "name"));
            let data = get_obj_field(band, "data");
            let vn = emit_data_array("Band", &name, &data);
            bvars.push(vn);
        }
        ind_band_vars.push(bvars);

        let pane_la = get_obj_field(pane, "lineAreas");
        let mut lavars: Vec<String> = Vec::new();
        for la in get_array_ref(&pane_la) {
            let name = get_str(&get_obj_field(la, "name"));
            let data = get_obj_field(la, "data");
            let vn = emit_data_array("Scalar", &name, &data);
            lavars.push(vn);
        }
        ind_linearea_vars.push(lavars);

        let hm = get_obj_field(pane, "heatmap");
        let hm_var = if let JVal::Object(_) = &hm {
            let name = get_str(&get_obj_field(&hm, "name"));
            let data = get_obj_field(&hm, "data");
            Some(emit_data_array("Heatmap", &name, &data))
        } else {
            None
        };
        ind_heatmap_vars.push(hm_var);
    }

    // Emit configuration object.
    let config_var = format!("{}Config", sanitize_var_name(base_name));
    lines.push(format!("export const {}: Configuration = {{", config_var));
    lines.push(format!("  width: \"{}\",", get_str(get("width"))));

    let nav = get("navigationPane");
    if let JVal::Object(_) = nav {
        lines.push(format!(
            "  navigationPane: {{ height: {}, hasLine: {}, hasArea: {}, hasTimeAxis: {}, timeTicks: {} }},",
            get_int(&get_obj_field(nav, "height")),
            ts_bool(get_bool(&get_obj_field(nav, "hasLine"))),
            ts_bool(get_bool(&get_obj_field(nav, "hasArea"))),
            ts_bool(get_bool(&get_obj_field(nav, "hasTimeAxis"))),
            get_int(&get_obj_field(nav, "timeTicks")),
        ));
    }

    lines.push(format!("  heightNavigationPane: {},", get_int(get("heightNavigationPane"))));
    lines.push(format!("  timeAnnotationFormat: \"{}\",", get_str(get("timeAnnotationFormat"))));
    lines.push(format!("  axisLeft: {},", ts_bool(get_bool(get("axisLeft")))));
    lines.push(format!("  axisRight: {},", ts_bool(get_bool(get("axisRight")))));

    let margin = get("margin");
    lines.push(format!(
        "  margin: {{ left: {}, top: {}, right: {}, bottom: {} }},",
        get_int(&get_obj_field(margin, "left")),
        get_int(&get_obj_field(margin, "top")),
        get_int(&get_obj_field(margin, "right")),
        get_int(&get_obj_field(margin, "bottom")),
    ));

    lines.push(format!(
        "  ohlcv: {{ name: \"{}\", data: {}, candlesticks: {} }},",
        get_str(&get_obj_field(ohlcv, "name")),
        ohlcv_var,
        ts_bool(get_bool(&get_obj_field(ohlcv, "candlesticks"))),
    ));

    // Price pane.
    lines.push("  pricePane: {".to_string());
    write_ts_pane(&mut lines, price_pane, &price_line_vars, &price_band_vars, &price_linearea_vars, None, "    ");
    lines.push("  },".to_string());

    // Indicator panes.
    lines.push("  indicatorPanes: [".to_string());
    for (pi, pane) in ind_panes.iter().enumerate() {
        lines.push("    {".to_string());
        write_ts_pane(
            &mut lines, pane,
            &ind_line_vars[pi], &ind_band_vars[pi], &ind_linearea_vars[pi],
            ind_heatmap_vars[pi].as_deref(), "      ",
        );
        lines.push("    },".to_string());
    }
    lines.push("  ],".to_string());

    lines.push(format!("  crosshair: {},", ts_bool(get_bool(get("crosshair")))));
    lines.push(format!("  volumeInPricePane: {},", ts_bool(get_bool(get("volumeInPricePane")))));
    lines.push(format!("  menuVisible: {},", ts_bool(get_bool(get("menuVisible")))));
    lines.push(format!("  downloadSvgVisible: {},", ts_bool(get_bool(get("downloadSvgVisible")))));
    lines.push("};".to_string());

    lines.join("\n") + "\n"
}

fn write_ts_pane(
    lines: &mut Vec<String>,
    pane: &JVal,
    line_vars: &[String],
    band_vars: &[String],
    linearea_vars: &[String],
    heatmap_var: Option<&str>,
    indent: &str,
) {
    let get_obj_field = |obj: &JVal, key: &str| -> JVal {
        if let JVal::Object(pairs) = obj {
            pairs.iter().find(|(k, _)| k == key).map(|(_, v)| v.clone()).unwrap_or(JVal::Null)
        } else {
            JVal::Null
        }
    };

    let get_str = |jv: &JVal| -> String {
        if let JVal::Str(s) = jv { s.clone() } else { String::new() }
    };

    let get_int = |jv: &JVal| -> i64 {
        if let JVal::Int(n) = jv { *n } else { 0 }
    };

    let get_float = |jv: &JVal| -> f64 {
        match jv {
            JVal::Float(f) => *f,
            JVal::Int(n) => *n as f64,
            _ => 0.0,
        }
    };

    let get_bool = |jv: &JVal| -> bool {
        if let JVal::Bool(b) = jv { *b } else { false }
    };

    let get_array = |jv: &JVal| -> Vec<JVal> {
        if let JVal::Array(arr) = jv { arr.clone() } else { Vec::new() }
    };

    let height = get_str(&get_obj_field(pane, "height"));
    let value_format = get_str(&get_obj_field(pane, "valueFormat"));
    let vmpf = get_float(&get_obj_field(pane, "valueMarginPercentageFactor"));

    lines.push(format!(
        "{}height: \"{}\", valueFormat: \"{}\", valueMarginPercentageFactor: {},",
        indent, height, value_format, ts_num(vmpf)
    ));

    // Heatmap.
    if let Some(hm_var) = heatmap_var {
        let hm = get_obj_field(pane, "heatmap");
        if let JVal::Object(_) = &hm {
            lines.push(format!(
                "{}heatmap: {{ name: \"{}\", data: {}, indicator: {}, output: {}, gradient: \"{}\", invertGradient: {} }},",
                indent,
                get_str(&get_obj_field(&hm, "name")),
                hm_var,
                get_int(&get_obj_field(&hm, "indicator")),
                get_int(&get_obj_field(&hm, "output")),
                get_str(&get_obj_field(&hm, "gradient")),
                ts_bool(get_bool(&get_obj_field(&hm, "invertGradient"))),
            ));
        }
    }

    // Bands.
    let bands = get_array(&get_obj_field(pane, "bands"));
    if !bands.is_empty() {
        lines.push(format!("{}bands: [", indent));
        for (i, b) in bands.iter().enumerate() {
            let vn = &band_vars[i];
            lines.push(format!(
                "{}  {{ name: \"{}\", data: {}, indicator: {}, output: {}, color: \"{}\", legendColor: \"{}\", interpolation: \"{}\" }},",
                indent,
                get_str(&get_obj_field(b, "name")),
                vn,
                get_int(&get_obj_field(b, "indicator")),
                get_int(&get_obj_field(b, "output")),
                get_str(&get_obj_field(b, "color")),
                get_str(&get_obj_field(b, "legendColor")),
                get_str(&get_obj_field(b, "interpolation")),
            ));
        }
        lines.push(format!("{}],", indent));
    } else {
        lines.push(format!("{}bands: [],", indent));
    }

    // LineAreas.
    let line_areas = get_array(&get_obj_field(pane, "lineAreas"));
    if !line_areas.is_empty() {
        lines.push(format!("{}lineAreas: [", indent));
        for (i, la) in line_areas.iter().enumerate() {
            let vn = &linearea_vars[i];
            lines.push(format!(
                "{}  {{ name: \"{}\", data: {}, indicator: {}, output: {}, value: {}, color: \"{}\", legendColor: \"{}\", interpolation: \"{}\" }},",
                indent,
                get_str(&get_obj_field(la, "name")),
                vn,
                get_int(&get_obj_field(la, "indicator")),
                get_int(&get_obj_field(la, "output")),
                ts_num(get_float(&get_obj_field(la, "value"))),
                get_str(&get_obj_field(la, "color")),
                get_str(&get_obj_field(la, "legendColor")),
                get_str(&get_obj_field(la, "interpolation")),
            ));
        }
        lines.push(format!("{}],", indent));
    } else {
        lines.push(format!("{}lineAreas: [],", indent));
    }

    // Horizontals.
    let horizontals = get_array(&get_obj_field(pane, "horizontals"));
    if !horizontals.is_empty() {
        lines.push(format!("{}horizontals: [", indent));
        for h in &horizontals {
            lines.push(format!(
                "{}  {{ value: {}, color: \"{}\", width: {}, dash: \"{}\" }},",
                indent,
                ts_num(get_float(&get_obj_field(h, "value"))),
                get_str(&get_obj_field(h, "color")),
                ts_num(get_float(&get_obj_field(h, "width"))),
                get_str(&get_obj_field(h, "dash")),
            ));
        }
        lines.push(format!("{}],", indent));
    } else {
        lines.push(format!("{}horizontals: [],", indent));
    }

    // Lines.
    let pane_lines = get_array(&get_obj_field(pane, "lines"));
    if !pane_lines.is_empty() {
        lines.push(format!("{}lines: [", indent));
        for (i, l) in pane_lines.iter().enumerate() {
            let vn = &line_vars[i];
            lines.push(format!(
                "{}  {{ name: \"{}\", data: {}, indicator: {}, output: {}, color: \"{}\", width: {}, dash: \"{}\", interpolation: \"{}\" }},",
                indent,
                get_str(&get_obj_field(l, "name")),
                vn,
                get_int(&get_obj_field(l, "indicator")),
                get_int(&get_obj_field(l, "output")),
                get_str(&get_obj_field(l, "color")),
                ts_num(get_float(&get_obj_field(l, "width"))),
                get_str(&get_obj_field(l, "dash")),
                get_str(&get_obj_field(l, "interpolation")),
            ));
        }
        lines.push(format!("{}],", indent));
    } else {
        lines.push(format!("{}lines: [],", indent));
    }

    // Arrows.
    lines.push(format!("{}arrows: [],", indent));
}

// ---------------------------------------------------------------------------
// JSON value to string (for passing params to factory)
// ---------------------------------------------------------------------------

fn json_value_to_string(val: &JsonValue) -> String {
    match val {
        JsonValue::Null => "null".to_string(),
        JsonValue::Bool(b) => b.to_string(),
        JsonValue::Number(n) => {
            if *n == (*n as i64) as f64 && n.abs() < 1e15 {
                format!("{}", *n as i64)
            } else {
                format!("{}", n)
            }
        }
        JsonValue::Str(s) => format!("{:?}", s),
        JsonValue::Array(arr) => {
            let items: Vec<String> = arr.iter().map(json_value_to_string).collect();
            format!("[{}]", items.join(","))
        }
        JsonValue::Object(pairs) => {
            let items: Vec<String> = pairs.iter()
                .map(|(k, v)| format!("{:?}:{}", k, json_value_to_string(v)))
                .collect();
            format!("{{{}}}", items.join(","))
        }
    }
}

// ---------------------------------------------------------------------------
// Embedded test data (reused from icalc)
// ---------------------------------------------------------------------------

const BASE_TIME: i64 = 1577923200; // 2020-01-02 00:00:00 UTC
const DAY_SECS: i64 = 86400;

fn test_bars() -> Vec<Bar> {
    let highs = test_highs();
    let lows = test_lows();
    let closes = test_closes();
    let volumes = test_volumes();

    let mut bars = Vec::with_capacity(closes.len());
    for i in 0..closes.len() {
        let open_price = if i == 0 { closes[0] } else { closes[i - 1] };
        let t = BASE_TIME + (i as i64) * DAY_SECS;
        bars.push(Bar::new(t, open_price, highs[i], lows[i], closes[i], volumes[i]));
    }
    bars
}

fn test_highs() -> Vec<f64> {
    vec![
        93.25, 94.94, 96.375, 96.19, 96.0, 94.72, 95.0, 93.72, 92.47, 92.75,
        96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,
        90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,
        87.155, 87.25, 87.375, 88.97, 90.0, 89.845, 86.97, 85.94, 84.75, 85.47,
        84.47, 88.5, 89.47, 90.0, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,
        90.0, 88.875, 89.0, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,
        90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92.0, 91.94,
        89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,
        104.5, 105.5, 106.125, 107.94, 106.25, 107.0, 108.75, 110.94, 110.94, 114.22,
        123.0, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113.0, 118.315,
        116.87, 116.75, 113.87, 114.62, 115.31, 116.0, 121.69, 119.87, 120.87, 116.75,
        116.5, 116.0, 118.31, 121.5, 122.0, 121.44, 125.75, 127.75, 124.19, 124.44,
        125.75, 124.69, 125.31, 132.0, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,
        138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,
        125.12, 126.5, 128.69, 126.62, 126.69, 126.0, 123.12, 121.87, 124.0, 127.0,
        124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,
        124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126.0, 128.5, 126.75, 129.75,
        132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135.0, 132.38, 131.44, 130.88,
        129.63, 127.25, 127.81, 125.0, 126.81, 124.75, 122.81, 122.25, 121.06, 120.0,
        123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109.0,
        108.5, 113.06, 93.0, 94.62, 95.12, 96.0, 95.56, 95.31, 99.0, 98.81,
        96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,
        95.0, 94.87, 98.25, 105.12, 108.44, 109.87, 105.0, 106.0, 104.94, 104.5,
        104.44, 106.31, 112.87, 116.5, 119.19, 121.0, 122.12, 111.94, 112.75, 110.19,
        107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110.0, 110.75, 110.5,
        110.5, 109.5,
    ]
}

fn test_lows() -> Vec<f64> {
    vec![
        90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92.0, 89.75, 89.44, 90.625,
        92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,
        87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,
        84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,
        82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89.0, 88.565, 90.095,
        89.0, 86.47, 84.0, 83.315, 82.0, 83.25, 84.75, 85.28, 87.19, 88.44,
        88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92.0, 90.53, 89.97, 88.815,
        86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,
        101.72, 101.72, 103.155, 105.69, 103.655, 104.0, 105.53, 108.53, 108.75, 107.75,
        117.0, 118.0, 116.0, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,
        112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,
        112.56, 114.0, 114.75, 118.87, 119.0, 119.75, 122.62, 123.0, 121.75, 121.56,
        123.12, 122.19, 122.75, 124.37, 128.0, 129.5, 130.81, 130.63, 132.13, 133.88,
        135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,
        122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121.0,
        121.12, 118.94, 119.81, 121.0, 122.0, 124.5, 126.56, 123.5, 121.25, 121.06,
        122.31, 121.0, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128.0,
        128.38, 130.69, 131.63, 134.38, 132.0, 131.94, 131.94, 129.56, 123.75, 126.0,
        126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120.0, 119.62, 115.5, 116.75,
        119.06, 119.06, 115.06, 111.06, 113.12, 110.0, 105.0, 104.69, 103.87, 104.69,
        105.44, 107.0, 89.0, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,
        93.69, 93.5, 90.0, 90.19, 90.5, 92.12, 94.12, 94.87, 93.0, 93.87,
        93.0, 92.62, 93.56, 98.37, 104.44, 106.0, 101.81, 104.12, 103.37, 102.12,
        102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,
        104.5, 105.75, 108.62, 107.75, 108.06, 108.0, 108.19, 108.12, 109.06, 108.75,
        108.56, 106.62,
    ]
}

fn test_closes() -> Vec<f64> {
    vec![
        91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,
        96.125, 97.25, 98.5, 89.875, 91.0, 92.815, 89.155, 89.345, 91.625, 89.875,
        88.375, 87.625, 84.78, 83.0, 83.5, 81.375, 84.44, 89.25, 86.375, 86.25,
        85.25, 87.125, 85.815, 88.97, 88.47, 86.875, 86.815, 84.875, 84.19, 83.875,
        83.375, 85.5, 89.19, 89.44, 91.095, 90.75, 91.44, 89.0, 91.0, 90.5,
        89.03, 88.815, 84.28, 83.5, 82.69, 84.75, 85.655, 86.19, 88.94, 89.28,
        88.625, 88.5, 91.97, 91.5, 93.25, 93.5, 93.155, 91.72, 90.0, 89.69,
        88.875, 85.19, 83.375, 84.875, 85.94, 97.25, 99.875, 104.94, 106.0, 102.5,
        102.405, 104.595, 106.125, 106.0, 106.065, 104.625, 108.625, 109.315, 110.5, 112.75,
        123.0, 119.625, 118.75, 119.25, 117.94, 116.44, 115.19, 111.875, 110.595, 118.125,
        116.0, 116.0, 112.0, 113.75, 112.94, 116.0, 120.5, 116.62, 117.0, 115.25,
        114.31, 115.5, 115.87, 120.69, 120.19, 120.75, 124.75, 123.37, 122.94, 122.56,
        123.12, 122.56, 124.62, 129.25, 131.0, 132.25, 131.0, 132.81, 134.0, 137.38,
        137.81, 137.88, 137.25, 136.31, 136.25, 134.63, 128.25, 129.0, 123.87, 124.81,
        123.0, 126.25, 128.38, 125.37, 125.69, 122.25, 119.37, 118.5, 123.19, 123.5,
        122.19, 119.31, 123.31, 121.12, 123.37, 127.37, 128.5, 123.87, 122.94, 121.75,
        124.44, 122.0, 122.37, 122.94, 124.0, 123.19, 124.56, 127.25, 125.87, 128.86,
        132.0, 130.75, 134.75, 135.0, 132.38, 133.31, 131.94, 130.0, 125.37, 130.13,
        127.12, 125.19, 122.0, 125.0, 123.0, 123.5, 120.06, 121.0, 117.75, 119.87,
        122.0, 119.19, 116.37, 113.5, 114.25, 110.0, 105.06, 107.0, 107.87, 107.0,
        107.12, 107.0, 91.0, 93.94, 93.87, 95.5, 93.0, 94.94, 98.25, 96.75,
        94.81, 94.37, 91.56, 90.25, 93.94, 93.62, 97.0, 95.0, 95.87, 94.06,
        94.62, 93.75, 98.0, 103.94, 107.87, 106.06, 104.5, 105.0, 104.19, 103.06,
        103.42, 105.27, 111.87, 116.0, 116.62, 118.28, 113.37, 109.0, 109.7, 109.25,
        107.0, 109.19, 110.0, 109.2, 110.12, 108.0, 108.62, 109.75, 109.81, 109.0,
        108.75, 107.87,
    ]
}

fn test_volumes() -> Vec<f64> {
    vec![
        4077500.0, 4955900.0, 4775300.0, 4155300.0, 4593100.0, 3631300.0, 3382800.0, 4954200.0, 4500000.0, 3397500.0,
        4204500.0, 6321400.0, 10203600.0, 19043900.0, 11692000.0, 9553300.0, 8920300.0, 5970900.0, 5062300.0, 3705600.0,
        5865600.0, 5603000.0, 5811900.0, 8483800.0, 5995200.0, 5408800.0, 5430500.0, 6283800.0, 5834800.0, 4515500.0,
        4493300.0, 4346100.0, 3700300.0, 4600200.0, 4557200.0, 4323600.0, 5237500.0, 7404100.0, 4798400.0, 4372800.0,
        3872300.0, 10750800.0, 5804800.0, 3785500.0, 5014800.0, 3507700.0, 4298800.0, 4842500.0, 3952200.0, 3304700.0,
        3462000.0, 7253900.0, 9753100.0, 5953000.0, 5011700.0, 5910800.0, 4916900.0, 4135000.0, 4054200.0, 3735300.0,
        2921900.0, 2658400.0, 4624400.0, 4372200.0, 5831600.0, 4268600.0, 3059200.0, 4495500.0, 3425000.0, 3630800.0,
        4168100.0, 5966900.0, 7692800.0, 7362500.0, 6581300.0, 19587700.0, 10378600.0, 9334700.0, 10467200.0, 5671400.0,
        5645000.0, 4518600.0, 4519500.0, 5569700.0, 4239700.0, 4175300.0, 4995300.0, 4776600.0, 4190000.0, 6035300.0,
        12168900.0, 9040800.0, 5780300.0, 4320800.0, 3899100.0, 3221400.0, 3455500.0, 4304200.0, 4703900.0, 8316300.0,
        10553900.0, 6384800.0, 7163300.0, 7007800.0, 5114100.0, 5263800.0, 6666100.0, 7398400.0, 5575000.0, 4852300.0,
        4298100.0, 4900500.0, 4887700.0, 6964800.0, 4679200.0, 9165000.0, 6469800.0, 6792000.0, 4423800.0, 5231900.0,
        4565600.0, 6235200.0, 5225900.0, 8261400.0, 5912500.0, 3545600.0, 5714500.0, 6653900.0, 6094500.0, 4799200.0,
        5050800.0, 5648900.0, 4726300.0, 5585600.0, 5124800.0, 7630200.0, 14311600.0, 8793600.0, 8874200.0, 6966600.0,
        5525500.0, 6515500.0, 5291900.0, 5711700.0, 4327700.0, 4568000.0, 6859200.0, 5757500.0, 7367000.0, 6144100.0,
        4052700.0, 5849700.0, 5544700.0, 5032200.0, 4400600.0, 4894100.0, 5140000.0, 6610900.0, 7585200.0, 5963100.0,
        6045500.0, 8443300.0, 6464700.0, 6248300.0, 4357200.0, 4774700.0, 6216900.0, 6266900.0, 5584800.0, 5284500.0,
        7554500.0, 7209500.0, 8424800.0, 5094500.0, 4443600.0, 4591100.0, 5658400.0, 6094100.0, 14862200.0, 7544700.0,
        6985600.0, 8093000.0, 7590000.0, 7451300.0, 7078000.0, 7105300.0, 8778800.0, 6643900.0, 10563900.0, 7043100.0,
        6438900.0, 8057700.0, 14240000.0, 17872300.0, 7831100.0, 8277700.0, 15017800.0, 14183300.0, 13921100.0, 9683000.0,
        9187300.0, 11380500.0, 69447300.0, 26673600.0, 13768400.0, 11371600.0, 9872200.0, 9450500.0, 11083300.0, 9552800.0,
        11108400.0, 10374200.0, 16701900.0, 13741900.0, 8523600.0, 9551900.0, 8680500.0, 7151700.0, 9673100.0, 6264700.0,
        8541600.0, 8358000.0, 18720800.0, 19683100.0, 13682500.0, 10668100.0, 9710600.0, 3113100.0, 5682000.0, 5763600.0,
        5340000.0, 6220800.0, 14680500.0, 9933000.0, 11329500.0, 8145300.0, 16644700.0, 12593800.0, 7138100.0, 7442300.0,
        9442300.0, 7123600.0, 7680600.0, 4839800.0, 4775500.0, 4008800.0, 4533600.0, 3741100.0, 4084800.0, 2685200.0,
        3438000.0, 2870500.0,
    ]
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() < 3 {
        eprintln!("usage: iconf <settings.json> <output-name>");
        process::exit(1);
    }

    let settings_path = &args[1];
    let mut output_name = args[2].clone();
    if output_name.ends_with(".json") {
        output_name = output_name[..output_name.len() - 5].to_string();
    }
    if output_name.ends_with(".ts") {
        output_name = output_name[..output_name.len() - 3].to_string();
    }

    let data = fs::read_to_string(settings_path).unwrap_or_else(|e| {
        eprintln!("error reading settings file: {}", e);
        process::exit(1);
    });

    let parsed = JsonValue::parse(&data).unwrap_or_else(|e| {
        eprintln!("error parsing settings file: {}", e);
        process::exit(1);
    });

    let entries = parsed.as_array().unwrap_or_else(|| {
        eprintln!("error: settings file must be a JSON array");
        process::exit(1);
    });

    let mut indicators: Vec<Box<dyn Indicator>> = Vec::with_capacity(entries.len());

    for entry in entries {
        let id_str = entry.get("identifier")
            .and_then(|v| v.as_str_val())
            .unwrap_or_else(|| {
                eprintln!("error: entry missing 'identifier' string");
                process::exit(1);
            });

        let ident = Identifier::from_str(id_str).unwrap_or_else(|| {
            eprintln!("error: unknown indicator identifier: {}", id_str);
            process::exit(1);
        });

        let params_json = match entry.get("params") {
            Some(p) => json_value_to_string(p),
            None => "{}".to_string(),
        };

        let ind = create_indicator(ident, &params_json).unwrap_or_else(|e| {
            eprintln!("error creating indicator {}: {}", id_str, e);
            process::exit(1);
        });

        indicators.push(ind);
    }

    let bars = test_bars();
    let cfg = build_configuration(&mut indicators, &bars);

    // Write JSON.
    let json_path = format!("{}.json", output_name);
    let json_data = cfg.to_json(2, 0);
    fs::write(&json_path, &json_data).unwrap_or_else(|e| {
        eprintln!("error writing {}: {}", json_path, e);
        process::exit(1);
    });
    println!("wrote {}", json_path);

    // Write TypeScript.
    let ts_path = format!("{}.ts", output_name);
    let base_name = std::path::Path::new(&output_name)
        .file_name()
        .and_then(|n| n.to_str())
        .unwrap_or(&output_name);
    let ts_data = build_typescript(&cfg, base_name);
    fs::write(&ts_path, &ts_data).unwrap_or_else(|e| {
        eprintln!("error writing {}: {}", ts_path, e);
        process::exit(1);
    });
    println!("wrote {}", ts_path);
}
