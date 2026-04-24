/**
 * iconf — command-line chart configuration generator (TypeScript).
 *
 * Reads a JSON settings file containing indicator definitions, creates
 * indicator instances via the factory, runs embedded bar data through them,
 * and writes chart configuration files in JSON and TypeScript formats.
 *
 * Usage:  npx tsx cmd/iconf/main.ts <settings.json> <output-name>
 */

import { readFileSync, writeFileSync } from 'node:fs';
import { basename } from 'node:path';
import { Bar } from '../../entities/bar.js';
import { Scalar } from '../../entities/scalar.js';
import { type Indicator } from '../../indicators/core/indicator.js';
import { IndicatorIdentifier } from '../../indicators/core/indicator-identifier.js';
import { type IndicatorOutput } from '../../indicators/core/indicator-output.js';
import { Band } from '../../indicators/core/outputs/band.js';
import { Heatmap } from '../../indicators/core/outputs/heatmap.js';
import { Shape } from '../../indicators/core/outputs/shape/shape.js';
import { Pane } from '../../indicators/core/pane.js';
import { descriptorOf } from '../../indicators/core/descriptors.js';
import { createIndicator } from '../../indicators/factory/factory.js';

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

interface SettingsEntry {
    identifier: string;
    params: Record<string, unknown>;
}

/** Map camelCase JSON identifier string to IndicatorIdentifier enum value. */
function resolveIdentifier(name: string): IndicatorIdentifier {
    let pascal = name.charAt(0).toUpperCase() + name.slice(1);
    if (pascal === 'KaufmanAdaptiveMovingAverageMovingAverage') {
        pascal = 'KaufmanAdaptiveMovingAverage';
    }
    const value = (IndicatorIdentifier as unknown as Record<string, number>)[pascal];
    if (value === undefined) {
        throw new Error(`Unknown indicator identifier: "${name}" (tried "${pascal}")`);
    }
    return value as IndicatorIdentifier;
}

// ---------------------------------------------------------------------------
// Chart configuration types (mirrors Go structs / TS template types)
// ---------------------------------------------------------------------------

interface Configuration {
    width: string;
    navigationPane?: NavigationPane;
    heightNavigationPane: number;
    timeAnnotationFormat: string;
    axisLeft: boolean;
    axisRight: boolean;
    margin: Margin;
    ohlcv: OhlcvData;
    pricePane: PaneData;
    indicatorPanes: PaneData[];
    crosshair: boolean;
    volumeInPricePane: boolean;
    menuVisible: boolean;
    downloadSvgVisible: boolean;
}

interface NavigationPane {
    height: number;
    hasLine: boolean;
    hasArea: boolean;
    hasTimeAxis: boolean;
    timeTicks: number;
}

interface Margin { left: number; top: number; right: number; bottom: number; }

interface OhlcvData {
    name: string;
    data: OhlcvBar[];
    candlesticks: boolean;
}

interface OhlcvBar {
    time: string;
    open: number;
    high: number;
    low: number;
    close: number;
    volume: number;
}

interface PaneData {
    height: string;
    valueFormat: string;
    valueMarginPercentageFactor: number;
    heatmap?: HeatmapDataItem;
    bands: BandDataItem[];
    lineAreas: LineAreaDataItem[];
    horizontals: HorizontalDataItem[];
    lines: LineDataItem[];
    arrows: ArrowDataItem[];
}

interface LineDataItem {
    name: string;
    data: ScalarVal[];
    indicator: number;
    output: number;
    color: string;
    width: number;
    dash: string;
    interpolation: string;
}

interface ScalarVal { time: string; value: number; }

interface BandDataItem {
    name: string;
    data: BandVal[];
    indicator: number;
    output: number;
    color: string;
    legendColor: string;
    interpolation: string;
}

interface BandVal { time: string; upper: number; lower: number; }

interface LineAreaDataItem {
    name: string;
    data: ScalarVal[];
    indicator: number;
    output: number;
    value: number;
    color: string;
    legendColor: string;
    interpolation: string;
}

interface HeatmapDataItem {
    name: string;
    data: HeatmapVal[];
    indicator: number;
    output: number;
    gradient: string;
    invertGradient: boolean;
}

interface HeatmapVal {
    time: string;
    first: number;
    last: number;
    result: number;
    min: number;
    max: number;
    values: number[];
}

interface HorizontalDataItem { value: number; color: string; width: number; dash: string; }
interface ArrowDataItem { name: string; down: boolean; time: string; indicator: number; output: number; color: string; }

// ---------------------------------------------------------------------------
// Default colors
// ---------------------------------------------------------------------------

const lineColors = [
    '#e41a1c', '#377eb8', '#4daf4a', '#984ea3', '#ff7f00',
    '#a65628', '#f781bf', '#999999', '#66c2a5', '#fc8d62',
];

const bandColors = [
    'rgba(0,255,0,0.3)', 'rgba(0,0,255,0.3)', 'rgba(255,0,0,0.3)',
    'rgba(128,0,128,0.3)', 'rgba(0,128,128,0.3)',
];

// ---------------------------------------------------------------------------
// Output placement
// ---------------------------------------------------------------------------

interface OutputPlacement {
    indicatorIdx: number;
    outputIdx: number;
    mnemonic: string;
    shape: Shape;
    pane: Pane;
}

// ---------------------------------------------------------------------------
// Configuration builder
// ---------------------------------------------------------------------------

function formatDate(d: Date): string {
    return d.toISOString().slice(0, 10);
}

function buildConfiguration(indicators: Indicator[], bars: Bar[]): Configuration {
    // Collect output placements from descriptor registry.
    const placements: OutputPlacement[] = [];
    for (let i = 0; i < indicators.length; i++) {
        const meta = indicators[i].metadata();
        const desc = descriptorOf(meta.identifier);
        if (!desc) {
            // Fallback: treat all outputs as Own/Scalar.
            for (let j = 0; j < meta.outputs.length; j++) {
                placements.push({
                    indicatorIdx: i, outputIdx: j,
                    mnemonic: meta.outputs[j].mnemonic, shape: meta.outputs[j].shape, pane: Pane.Own,
                });
            }
            continue;
        }
        for (let j = 0; j < desc.outputs.length; j++) {
            const mnemonic = j < meta.outputs.length ? meta.outputs[j].mnemonic : `out[${j}]`;
            placements.push({
                indicatorIdx: i, outputIdx: j,
                mnemonic, shape: desc.outputs[j].shape, pane: desc.outputs[j].pane,
            });
        }
    }

    // Run bars through indicators, collecting all outputs.
    const allOutputs: IndicatorOutput[][] = [];
    for (let i = 0; i < bars.length; i++) {
        const row: IndicatorOutput[] = [];
        for (const ind of indicators) {
            row.push(ind.updateBar(bars[i]));
        }
        allOutputs.push(row);
    }

    // Build OHLCV data.
    const ohlcvBars: OhlcvBar[] = bars.map(b => ({
        time: formatDate(b.time), open: b.open, high: b.high,
        low: b.low, close: b.close, volume: b.volume,
    }));

    // Group placements: price pane vs own panes.
    const ownPaneMap = new Map<number, number>(); // indicatorIdx -> indicatorPanes index
    const indicatorPanes: PaneData[] = [];

    const pricePane: PaneData = {
        height: '30%', valueFormat: ',.2f', valueMarginPercentageFactor: 0.01,
        bands: [], lineAreas: [], horizontals: [], lines: [], arrows: [],
    };

    let colorIdx = 0;
    let bandColorIdx = 0;

    for (const p of placements) {
        if (p.pane === Pane.Price && p.shape === Shape.Scalar) {
            pricePane.lines.push(buildLineData(p, allOutputs, bars, lineColors[colorIdx % lineColors.length]));
            colorIdx++;
        } else if (p.pane === Pane.Price && p.shape === Shape.Band) {
            pricePane.bands.push(buildBandData(p, allOutputs, bars, bandColors[bandColorIdx % bandColors.length]));
            bandColorIdx++;
        } else {
            // Own pane (or fallback).
            let idx = ownPaneMap.get(p.indicatorIdx);
            if (idx === undefined) {
                idx = indicatorPanes.length;
                ownPaneMap.set(p.indicatorIdx, idx);
                indicatorPanes.push({
                    height: '60', valueFormat: ',.2f', valueMarginPercentageFactor: 0.01,
                    bands: [], lineAreas: [], horizontals: [], lines: [], arrows: [],
                });
            }
            const pane = indicatorPanes[idx];

            if (p.shape === Shape.Heatmap) {
                pane.heatmap = buildHeatmapData(p, allOutputs, bars);
                pane.height = '120';
                pane.valueMarginPercentageFactor = 0;
            } else if (p.shape === Shape.Band) {
                pane.bands.push(buildBandData(p, allOutputs, bars, bandColors[bandColorIdx % bandColors.length]));
                bandColorIdx++;
            } else {
                pane.lines.push(buildLineData(p, allOutputs, bars, lineColors[colorIdx % lineColors.length]));
                colorIdx++;
            }
        }
    }

    return {
        width: '100%',
        navigationPane: { height: 30, hasLine: true, hasArea: false, hasTimeAxis: true, timeTicks: 0 },
        heightNavigationPane: 30,
        timeAnnotationFormat: '%Y-%m-%d',
        axisLeft: true,
        axisRight: false,
        margin: { left: 0, top: 10, right: 20, bottom: 0 },
        ohlcv: { name: 'TEST', data: ohlcvBars, candlesticks: true },
        pricePane,
        indicatorPanes,
        crosshair: false,
        volumeInPricePane: true,
        menuVisible: true,
        downloadSvgVisible: true,
    };
}

function buildLineData(p: OutputPlacement, allOutputs: IndicatorOutput[][], bars: Bar[], color: string): LineDataItem {
    const data: ScalarVal[] = [];
    for (let i = 0; i < bars.length; i++) {
        const out = allOutputs[i][p.indicatorIdx];
        if (p.outputIdx < out.length) {
            const val = out[p.outputIdx];
            if (val instanceof Scalar && !Number.isNaN(val.value)) {
                data.push({ time: formatDate(val.time), value: val.value });
            }
        }
    }
    return {
        name: p.mnemonic, data, indicator: p.indicatorIdx, output: p.outputIdx,
        color, width: 1, dash: '', interpolation: 'natural',
    };
}

function buildBandData(p: OutputPlacement, allOutputs: IndicatorOutput[][], bars: Bar[], color: string): BandDataItem {
    const data: BandVal[] = [];
    for (let i = 0; i < bars.length; i++) {
        const out = allOutputs[i][p.indicatorIdx];
        if (p.outputIdx < out.length) {
            const val = out[p.outputIdx];
            if (val instanceof Band && !Number.isNaN(val.lower) && !Number.isNaN(val.upper)) {
                data.push({ time: formatDate(val.time), upper: val.upper, lower: val.lower });
            }
        }
    }
    return {
        name: p.mnemonic, data, indicator: p.indicatorIdx, output: p.outputIdx,
        color, legendColor: color, interpolation: 'natural',
    };
}

function buildHeatmapData(p: OutputPlacement, allOutputs: IndicatorOutput[][], bars: Bar[]): HeatmapDataItem {
    const data: HeatmapVal[] = [];
    for (let i = 0; i < bars.length; i++) {
        const out = allOutputs[i][p.indicatorIdx];
        if (p.outputIdx < out.length) {
            const val = out[p.outputIdx];
            if (val instanceof Heatmap && !val.isEmpty()) {
                data.push({
                    time: formatDate(val.time), first: val.parameterFirst, last: val.parameterLast,
                    result: val.parameterResolution, min: val.valueMin, max: val.valueMax, values: val.values,
                });
            }
        }
    }
    return {
        name: p.mnemonic, data, indicator: p.indicatorIdx, output: p.outputIdx,
        gradient: 'Viridis', invertGradient: false,
    };
}

// ---------------------------------------------------------------------------
// TypeScript generation
// ---------------------------------------------------------------------------

function sanitizeVarName(s: string): string {
    let result = '';
    let capitalize = false;
    for (let i = 0; i < s.length; i++) {
        const c = s[i];
        if ((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
            if (capitalize) {
                result += c.toUpperCase();
                capitalize = false;
            } else {
                result += c;
            }
        } else if (c >= '0' && c <= '9') {
            if (i === 0) result += '_';
            result += c;
            capitalize = false;
        } else {
            capitalize = true;
        }
    }
    return result;
}

function buildTypeScript(cfg: Configuration, baseName: string): string {
    const lines: string[] = [];
    const w = (s: string) => lines.push(s);

    w("// Auto-generated chart configuration.");
    w("// eslint-disable-next-line");
    w("import { Configuration } from '../ohlcv-chart/template/configuration';");
    w("import { Scalar, Band, Heatmap, Bar } from '../ohlcv-chart/template/types';");
    w("");

    // Emit OHLCV data array.
    const ohlcvVar = sanitizeVarName(baseName) + 'Ohlcv';
    w(`export const ${ohlcvVar}: Bar[] = ${JSON.stringify(cfg.ohlcv.data)};`);
    w("");

    // Track variable names for data references.
    const lineVarMap = new Map<LineDataItem, string>();
    const bandVarMap = new Map<BandDataItem, string>();
    const heatmapVarMap = new Map<HeatmapDataItem, string>();
    const lineAreaVarMap = new Map<LineAreaDataItem, string>();

    let varCounter = 0;

    function emitScalarArray(mnemonic: string, data: ScalarVal[]): string {
        varCounter++;
        const vn = sanitizeVarName(`${baseName}_${mnemonic}_${varCounter}`);
        w(`export const ${vn}: Scalar[] = ${JSON.stringify(data)};`);
        w("");
        return vn;
    }

    function emitBandArray(mnemonic: string, data: BandVal[]): string {
        varCounter++;
        const vn = sanitizeVarName(`${baseName}_${mnemonic}_${varCounter}`);
        w(`export const ${vn}: Band[] = ${JSON.stringify(data)};`);
        w("");
        return vn;
    }

    function emitHeatmapArray(mnemonic: string, data: HeatmapVal[]): string {
        varCounter++;
        const vn = sanitizeVarName(`${baseName}_${mnemonic}_${varCounter}`);
        w(`export const ${vn}: Heatmap[] = ${JSON.stringify(data)};`);
        w("");
        return vn;
    }

    // Emit price pane data arrays.
    for (const l of cfg.pricePane.lines) {
        lineVarMap.set(l, emitScalarArray(l.name, l.data));
    }
    for (const b of cfg.pricePane.bands) {
        bandVarMap.set(b, emitBandArray(b.name, b.data));
    }
    for (const la of cfg.pricePane.lineAreas) {
        lineAreaVarMap.set(la, emitScalarArray(la.name, la.data));
    }

    // Emit indicator pane data arrays.
    for (const pane of cfg.indicatorPanes) {
        for (const l of pane.lines) {
            lineVarMap.set(l, emitScalarArray(l.name, l.data));
        }
        for (const b of pane.bands) {
            bandVarMap.set(b, emitBandArray(b.name, b.data));
        }
        for (const la of pane.lineAreas) {
            lineAreaVarMap.set(la, emitScalarArray(la.name, la.data));
        }
        if (pane.heatmap) {
            heatmapVarMap.set(pane.heatmap, emitHeatmapArray(pane.heatmap.name, pane.heatmap.data));
        }
    }

    // Emit configuration object.
    const cfgVar = sanitizeVarName(baseName) + 'Config';
    w(`export const ${cfgVar}: Configuration = {`);
    w(`  width: ${JSON.stringify(cfg.width)},`);

    if (cfg.navigationPane) {
        const np = cfg.navigationPane;
        w(`  navigationPane: {`);
        w(`    height: ${np.height}, hasLine: ${np.hasLine}, hasArea: ${np.hasArea}, hasTimeAxis: ${np.hasTimeAxis}, timeTicks: ${np.timeTicks},`);
        w(`  },`);
    }

    w(`  heightNavigationPane: ${cfg.heightNavigationPane},`);
    w(`  timeAnnotationFormat: ${JSON.stringify(cfg.timeAnnotationFormat)},`);
    w(`  axisLeft: ${cfg.axisLeft},`);
    w(`  axisRight: ${cfg.axisRight},`);
    w(`  margin: { left: ${cfg.margin.left}, top: ${cfg.margin.top}, right: ${cfg.margin.right}, bottom: ${cfg.margin.bottom} },`);
    w(`  ohlcv: { name: ${JSON.stringify(cfg.ohlcv.name)}, data: ${ohlcvVar}, candlesticks: ${cfg.ohlcv.candlesticks} },`);

    // Price pane.
    w(`  pricePane: {`);
    writeTSPane(lines, cfg.pricePane, lineVarMap, bandVarMap, lineAreaVarMap, heatmapVarMap, '    ');
    w(`  },`);

    // Indicator panes.
    w(`  indicatorPanes: [`);
    for (const pane of cfg.indicatorPanes) {
        w(`    {`);
        writeTSPane(lines, pane, lineVarMap, bandVarMap, lineAreaVarMap, heatmapVarMap, '      ');
        w(`    },`);
    }
    w(`  ],`);

    w(`  crosshair: ${cfg.crosshair},`);
    w(`  volumeInPricePane: ${cfg.volumeInPricePane},`);
    w(`  menuVisible: ${cfg.menuVisible},`);
    w(`  downloadSvgVisible: ${cfg.downloadSvgVisible},`);
    w(`};`);

    return lines.join('\n') + '\n';
}

function writeTSPane(
    lines: string[], pane: PaneData,
    lineVarMap: Map<LineDataItem, string>, bandVarMap: Map<BandDataItem, string>,
    lineAreaVarMap: Map<LineAreaDataItem, string>, heatmapVarMap: Map<HeatmapDataItem, string>,
    indent: string,
): void {
    const w = (s: string) => lines.push(s);

    w(`${indent}height: ${JSON.stringify(pane.height)}, valueFormat: ${JSON.stringify(pane.valueFormat)}, valueMarginPercentageFactor: ${pane.valueMarginPercentageFactor},`);

    if (pane.heatmap) {
        const hm = pane.heatmap;
        const vn = heatmapVarMap.get(hm)!;
        w(`${indent}heatmap: { name: ${JSON.stringify(hm.name)}, data: ${vn}, indicator: ${hm.indicator}, output: ${hm.output}, gradient: ${JSON.stringify(hm.gradient)}, invertGradient: ${hm.invertGradient} },`);
    }

    // Bands.
    if (pane.bands.length === 0) {
        w(`${indent}bands: [],`);
    } else {
        w(`${indent}bands: [`);
        for (const b of pane.bands) {
            const vn = bandVarMap.get(b)!;
            w(`${indent}  { name: ${JSON.stringify(b.name)}, data: ${vn}, indicator: ${b.indicator}, output: ${b.output}, color: ${JSON.stringify(b.color)}, legendColor: ${JSON.stringify(b.legendColor)}, interpolation: ${JSON.stringify(b.interpolation)} },`);
        }
        w(`${indent}],`);
    }

    // LineAreas.
    if (pane.lineAreas.length === 0) {
        w(`${indent}lineAreas: [],`);
    } else {
        w(`${indent}lineAreas: [`);
        for (const la of pane.lineAreas) {
            const vn = lineAreaVarMap.get(la)!;
            w(`${indent}  { name: ${JSON.stringify(la.name)}, data: ${vn}, indicator: ${la.indicator}, output: ${la.output}, value: ${la.value}, color: ${JSON.stringify(la.color)}, legendColor: ${JSON.stringify(la.legendColor)}, interpolation: ${JSON.stringify(la.interpolation)} },`);
        }
        w(`${indent}],`);
    }

    // Horizontals.
    if (pane.horizontals.length === 0) {
        w(`${indent}horizontals: [],`);
    } else {
        w(`${indent}horizontals: [`);
        for (const h of pane.horizontals) {
            w(`${indent}  { value: ${h.value}, color: ${JSON.stringify(h.color)}, width: ${h.width}, dash: ${JSON.stringify(h.dash)} },`);
        }
        w(`${indent}],`);
    }

    // Lines.
    if (pane.lines.length === 0) {
        w(`${indent}lines: [],`);
    } else {
        w(`${indent}lines: [`);
        for (const l of pane.lines) {
            const vn = lineVarMap.get(l)!;
            w(`${indent}  { name: ${JSON.stringify(l.name)}, data: ${vn}, indicator: ${l.indicator}, output: ${l.output}, color: ${JSON.stringify(l.color)}, width: ${l.width}, dash: ${JSON.stringify(l.dash)}, interpolation: ${JSON.stringify(l.interpolation)} },`);
        }
        w(`${indent}],`);
    }

    // Arrows.
    w(`${indent}arrows: [],`);
}

// ---------------------------------------------------------------------------
// Test bar data (252-entry TA-Lib reference set)
// ---------------------------------------------------------------------------

function testBars(): Bar[] {
    const highs = testHighs();
    const lows = testLows();
    const closes = testCloses();
    const volumes = testVolumes();

    const bars: Bar[] = [];
    const baseTime = new Date('2020-01-02T00:00:00Z');

    for (let i = 0; i < closes.length; i++) {
        const openPrice = i > 0 ? closes[i - 1] : closes[0];
        const bar = new Bar();
        bar.time = new Date(baseTime.getTime() + i * 86_400_000);
        bar.open = openPrice;
        bar.high = highs[i];
        bar.low = lows[i];
        bar.close = closes[i];
        bar.volume = volumes[i];
        bars.push(bar);
    }

    return bars;
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

function main(): void {
    const args = process.argv.slice(2);
    if (args.length < 2) {
        process.stderr.write('usage: iconf <settings.json> <output-name>\n');
        process.exit(1);
    }

    const settingsPath = args[0];
    let outputName = args[1];
    // Strip extension if provided.
    outputName = outputName.replace(/\.(json|ts)$/, '');

    const data = readFileSync(settingsPath, 'utf-8');
    const entries: SettingsEntry[] = JSON.parse(data);

    const indicators: Indicator[] = [];
    for (const e of entries) {
        const id = resolveIdentifier(e.identifier);
        const params = Object.keys(e.params).length === 0 ? undefined : e.params;
        try {
            indicators.push(createIndicator(id, params as Record<string, any>));
        } catch (err) {
            process.stderr.write(`error creating indicator ${e.identifier}: ${err}\n`);
            process.exit(1);
        }
    }

    const bars = testBars();
    const cfg = buildConfiguration(indicators, bars);

    // Write JSON.
    const jsonPath = outputName + '.json';
    writeFileSync(jsonPath, JSON.stringify(cfg, null, 2), 'utf-8');
    console.log(`wrote ${jsonPath}`);

    // Write TypeScript.
    const tsPath = outputName + '.ts';
    const tsData = buildTypeScript(cfg, basename(outputName));
    writeFileSync(tsPath, tsData, 'utf-8');
    console.log(`wrote ${tsPath}`);
}

// ---------------------------------------------------------------------------
// Embedded test data (TA-Lib test_data.c)
// ---------------------------------------------------------------------------

function testHighs(): number[] {
    return [
        93.25, 94.94, 96.375, 96.19, 96, 94.72, 95, 93.72, 92.47, 92.75,
        96.25, 99.625, 99.125, 92.75, 91.315, 93.25, 93.405, 90.655, 91.97, 92.25,
        90.345, 88.5, 88.25, 85.5, 84.44, 84.75, 84.44, 89.405, 88.125, 89.125,
        87.155, 87.25, 87.375, 88.97, 90, 89.845, 86.97, 85.94, 84.75, 85.47,
        84.47, 88.5, 89.47, 90, 92.44, 91.44, 92.97, 91.72, 91.155, 91.75,
        90, 88.875, 89, 85.25, 83.815, 85.25, 86.625, 87.94, 89.375, 90.625,
        90.75, 88.845, 91.97, 93.375, 93.815, 94.03, 94.03, 91.815, 92, 91.94,
        89.75, 88.75, 86.155, 84.875, 85.94, 99.375, 103.28, 105.375, 107.625, 105.25,
        104.5, 105.5, 106.125, 107.94, 106.25, 107, 108.75, 110.94, 110.94, 114.22,
        123, 121.75, 119.815, 120.315, 119.375, 118.19, 116.69, 115.345, 113, 118.315,
        116.87, 116.75, 113.87, 114.62, 115.31, 116, 121.69, 119.87, 120.87, 116.75,
        116.5, 116, 118.31, 121.5, 122, 121.44, 125.75, 127.75, 124.19, 124.44,
        125.75, 124.69, 125.31, 132, 131.31, 132.25, 133.88, 133.5, 135.5, 137.44,
        138.69, 139.19, 138.5, 138.13, 137.5, 138.88, 132.13, 129.75, 128.5, 125.44,
        125.12, 126.5, 128.69, 126.62, 126.69, 126, 123.12, 121.87, 124, 127,
        124.44, 122.5, 123.75, 123.81, 124.5, 127.87, 128.56, 129.63, 124.87, 124.37,
        124.87, 123.62, 124.06, 125.87, 125.19, 125.62, 126, 128.5, 126.75, 129.75,
        132.69, 133.94, 136.5, 137.69, 135.56, 133.56, 135, 132.38, 131.44, 130.88,
        129.63, 127.25, 127.81, 125, 126.81, 124.75, 122.81, 122.25, 121.06, 120,
        123.25, 122.75, 119.19, 115.06, 116.69, 114.87, 110.87, 107.25, 108.87, 109,
        108.5, 113.06, 93, 94.62, 95.12, 96, 95.56, 95.31, 99, 98.81,
        96.81, 95.94, 94.44, 92.94, 93.94, 95.5, 97.06, 97.5, 96.25, 96.37,
        95, 94.87, 98.25, 105.12, 108.44, 109.87, 105, 106, 104.94, 104.5,
        104.44, 106.31, 112.87, 116.5, 119.19, 121, 122.12, 111.94, 112.75, 110.19,
        107.94, 109.69, 111.06, 110.44, 110.12, 110.31, 110.44, 110, 110.75, 110.5,
        110.5, 109.5,
    ];
}

function testLows(): number[] {
    return [
        90.75, 91.405, 94.25, 93.5, 92.815, 93.5, 92, 89.75, 89.44, 90.625,
        92.75, 96.315, 96.03, 88.815, 86.75, 90.94, 88.905, 88.78, 89.25, 89.75,
        87.5, 86.53, 84.625, 82.28, 81.565, 80.875, 81.25, 84.065, 85.595, 85.97,
        84.405, 85.095, 85.5, 85.53, 87.875, 86.565, 84.655, 83.25, 82.565, 83.44,
        82.53, 85.065, 86.875, 88.53, 89.28, 90.125, 90.75, 89, 88.565, 90.095,
        89, 86.47, 84, 83.315, 82, 83.25, 84.75, 85.28, 87.19, 88.44,
        88.25, 87.345, 89.28, 91.095, 89.53, 91.155, 92, 90.53, 89.97, 88.815,
        86.75, 85.065, 82.03, 81.5, 82.565, 96.345, 96.47, 101.155, 104.25, 101.75,
        101.72, 101.72, 103.155, 105.69, 103.655, 104, 105.53, 108.53, 108.75, 107.75,
        117, 118, 116, 118.5, 116.53, 116.25, 114.595, 110.875, 110.5, 110.72,
        112.62, 114.19, 111.19, 109.44, 111.56, 112.44, 117.5, 116.06, 116.56, 113.31,
        112.56, 114, 114.75, 118.87, 119, 119.75, 122.62, 123, 121.75, 121.56,
        123.12, 122.19, 122.75, 124.37, 128, 129.5, 130.81, 130.63, 132.13, 133.88,
        135.38, 135.75, 136.19, 134.5, 135.38, 133.69, 126.06, 126.87, 123.5, 122.62,
        122.75, 123.56, 125.81, 124.62, 124.37, 121.81, 118.19, 118.06, 117.56, 121,
        121.12, 118.94, 119.81, 121, 122, 124.5, 126.56, 123.5, 121.25, 121.06,
        122.31, 121, 120.87, 122.06, 122.75, 122.69, 122.87, 125.5, 124.25, 128,
        128.38, 130.69, 131.63, 134.38, 132, 131.94, 131.94, 129.56, 123.75, 126,
        126.25, 124.37, 121.44, 120.44, 121.37, 121.69, 120, 119.62, 115.5, 116.75,
        119.06, 119.06, 115.06, 111.06, 113.12, 110, 105, 104.69, 103.87, 104.69,
        105.44, 107, 89, 92.5, 92.12, 94.62, 92.81, 94.25, 96.25, 96.37,
        93.69, 93.5, 90, 90.19, 90.5, 92.12, 94.12, 94.87, 93, 93.87,
        93, 92.62, 93.56, 98.37, 104.44, 106, 101.81, 104.12, 103.37, 102.12,
        102.25, 103.37, 107.94, 112.5, 115.44, 115.5, 112.25, 107.56, 106.56, 106.87,
        104.5, 105.75, 108.62, 107.75, 108.06, 108, 108.19, 108.12, 109.06, 108.75,
        108.56, 106.62,
    ];
}

function testCloses(): number[] {
    return [
        91.5, 94.815, 94.375, 95.095, 93.78, 94.625, 92.53, 92.75, 90.315, 92.47,
        96.125, 97.25, 98.5, 89.875, 91, 92.815, 89.155, 89.345, 91.625, 89.875,
        88.375, 87.625, 84.78, 83, 83.5, 81.375, 84.44, 89.25, 86.375, 86.25,
        85.25, 87.125, 85.815, 88.97, 88.47, 86.875, 86.815, 84.875, 84.19, 83.875,
        83.375, 85.5, 89.19, 89.44, 91.095, 90.75, 91.44, 89, 91, 90.5,
        89.03, 88.815, 84.28, 83.5, 82.69, 84.75, 85.655, 86.19, 88.94, 89.28,
        88.625, 88.5, 91.97, 91.5, 93.25, 93.5, 93.155, 91.72, 90, 89.69,
        88.875, 85.19, 83.375, 84.875, 85.94, 97.25, 99.875, 104.94, 106, 102.5,
        102.405, 104.595, 106.125, 106, 106.065, 104.625, 108.625, 109.315, 110.5, 112.75,
        123, 119.625, 118.75, 119.25, 117.94, 116.44, 115.19, 111.875, 110.595, 118.125,
        116, 116, 112, 113.75, 112.94, 116, 120.5, 116.62, 117, 115.25,
        114.31, 115.5, 115.87, 120.69, 120.19, 120.75, 124.75, 123.37, 122.94, 122.56,
        123.12, 122.56, 124.62, 129.25, 131, 132.25, 131, 132.81, 134, 137.38,
        137.81, 137.88, 137.25, 136.31, 136.25, 134.63, 128.25, 129, 123.87, 124.81,
        123, 126.25, 128.38, 125.37, 125.69, 122.25, 119.37, 118.5, 123.19, 123.5,
        122.19, 119.31, 123.31, 121.12, 123.37, 127.37, 128.5, 123.87, 122.94, 121.75,
        124.44, 122, 122.37, 122.94, 124, 123.19, 124.56, 127.25, 125.87, 128.86,
        132, 130.75, 134.75, 135, 132.38, 133.31, 131.94, 130, 125.37, 130.13,
        127.12, 125.19, 122, 125, 123, 123.5, 120.06, 121, 117.75, 119.87,
        122, 119.19, 116.37, 113.5, 114.25, 110, 105.06, 107, 107.87, 107,
        107.12, 107, 91, 93.94, 93.87, 95.5, 93, 94.94, 98.25, 96.75,
        94.81, 94.37, 91.56, 90.25, 93.94, 93.62, 97, 95, 95.87, 94.06,
        94.62, 93.75, 98, 103.94, 107.87, 106.06, 104.5, 105, 104.19, 103.06,
        103.42, 105.27, 111.87, 116, 116.62, 118.28, 113.37, 109, 109.7, 109.25,
        107, 109.19, 110, 109.2, 110.12, 108, 108.62, 109.75, 109.81, 109,
        108.75, 107.87,
    ];
}

function testVolumes(): number[] {
    return [
        4077500, 4955900, 4775300, 4155300, 4593100, 3631300, 3382800, 4954200, 4500000, 3397500,
        4204500, 6321400, 10203600, 19043900, 11692000, 9553300, 8920300, 5970900, 5062300, 3705600,
        5865600, 5603000, 5811900, 8483800, 5995200, 5408800, 5430500, 6283800, 5834800, 4515500,
        4493300, 4346100, 3700300, 4600200, 4557200, 4323600, 5237500, 7404100, 4798400, 4372800,
        3872300, 10750800, 5804800, 3785500, 5014800, 3507700, 4298800, 4842500, 3952200, 3304700,
        3462000, 7253900, 9753100, 5953000, 5011700, 5910800, 4916900, 4135000, 4054200, 3735300,
        2921900, 2658400, 4624400, 4372200, 5831600, 4268600, 3059200, 4495500, 3425000, 3630800,
        4168100, 5966900, 7692800, 7362500, 6581300, 19587700, 10378600, 9334700, 10467200, 5671400,
        5645000, 4518600, 4519500, 5569700, 4239700, 4175300, 4995300, 4776600, 4190000, 6035300,
        12168900, 9040800, 5780300, 4320800, 3899100, 3221400, 3455500, 4304200, 4703900, 8316300,
        10553900, 6384800, 7163300, 7007800, 5114100, 5263800, 6666100, 7398400, 5575000, 4852300,
        4298100, 4900500, 4887700, 6964800, 4679200, 9165000, 6469800, 6792000, 4423800, 5231900,
        4565600, 6235200, 5225900, 8261400, 5912500, 3545600, 5714500, 6653900, 6094500, 4799200,
        5050800, 5648900, 4726300, 5585600, 5124800, 7630200, 14311600, 8793600, 8874200, 6966600,
        5525500, 6515500, 5291900, 5711700, 4327700, 4568000, 6859200, 5757500, 7367000, 6144100,
        4052700, 5849700, 5544700, 5032200, 4400600, 4894100, 5140000, 6610900, 7585200, 5963100,
        6045500, 8443300, 6464700, 6248300, 4357200, 4774700, 6216900, 6266900, 5584800, 5284500,
        7554500, 7209500, 8424800, 5094500, 4443600, 4591100, 5658400, 6094100, 14862200, 7544700,
        6985600, 8093000, 7590000, 7451300, 7078000, 7105300, 8778800, 6643900, 10563900, 7043100,
        6438900, 8057700, 14240000, 17872300, 7831100, 8277700, 15017800, 14183300, 13921100, 9683000,
        9187300, 11380500, 69447300, 26673600, 13768400, 11371600, 9872200, 9450500, 11083300, 9552800,
        11108400, 10374200, 16701900, 13741900, 8523600, 9551900, 8680500, 7151700, 9673100, 6264700,
        8541600, 8358000, 18720800, 19683100, 13682500, 10668100, 9710600, 3113100, 5682000, 5763600,
        5340000, 6220800, 14680500, 9933000, 11329500, 8145300, 16644700, 12593800, 7138100, 7442300,
        9442300, 7123600, 7680600, 4839800, 4775500, 4008800, 4533600, 3741100, 4084800, 2685200,
        3438000, 2870500,
    ];
}

// ---------------------------------------------------------------------------
// Run
// ---------------------------------------------------------------------------

main();
