// Package main implements iconf, a command-line tool that creates chart
// configuration files from indicator settings.
//
// It reads a JSON settings file containing indicator definitions, creates
// indicator instances via the factory, runs embedded bar data through them,
// and writes chart configuration files in JSON and TypeScript formats.
//
// Usage:
//
//	iconf <settings.json> <output-name>
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
	"zpano/indicators/core/outputs/shape"
	"zpano/indicators/factory"
)

// settingsEntry represents a single indicator entry in the settings JSON file.
type settingsEntry struct {
	Identifier core.Identifier `json:"identifier"`
	Params     json.RawMessage `json:"params"`
}

// --- Chart configuration structs (mirror the TS template types) ---

// Configuration is the top-level chart configuration.
type Configuration struct {
	Width                string          `json:"width"`
	NavigationPane       *NavigationPane `json:"navigationPane,omitempty"`
	HeightNavigationPane int             `json:"heightNavigationPane,omitempty"`
	TimeAnnotationFormat string          `json:"timeAnnotationFormat,omitempty"`
	AxisLeft             bool            `json:"axisLeft"`
	AxisRight            bool            `json:"axisRight"`
	Margin               Margin          `json:"margin"`
	Ohlcv                OhlcvData       `json:"ohlcv"`
	PricePane            Pane            `json:"pricePane"`
	IndicatorPanes       []Pane          `json:"indicatorPanes"`
	Crosshair            bool            `json:"crosshair"`
	VolumeInPricePane    bool            `json:"volumeInPricePane"`
	MenuVisible          bool            `json:"menuVisible"`
	DownloadSvgVisible   bool            `json:"downloadSvgVisible"`
}

// NavigationPane describes a navigation pane.
type NavigationPane struct {
	Height      int  `json:"height"`
	HasLine     bool `json:"hasLine"`
	HasArea     bool `json:"hasArea"`
	HasTimeAxis bool `json:"hasTimeAxis"`
	TimeTicks   int  `json:"timeTicks"`
}

// Margin describes chart margins.
type Margin struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
}

// OhlcvData describes the OHLCV data source.
type OhlcvData struct {
	Name         string     `json:"name"`
	Data         []OhlcvBar `json:"data"`
	Candlesticks bool       `json:"candlesticks"`
}

// OhlcvBar is a bar for the chart (time as ISO string).
type OhlcvBar struct {
	Time   string  `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// Pane describes a chart pane.
type Pane struct {
	Height                      string           `json:"height"`
	ValueFormat                 string           `json:"valueFormat"`
	ValueMarginPercentageFactor float64          `json:"valueMarginPercentageFactor"`
	Heatmap                     *HeatmapData     `json:"heatmap,omitempty"`
	Bands                       []BandData       `json:"bands"`
	LineAreas                   []LineAreaData   `json:"lineAreas"`
	Horizontals                 []HorizontalData `json:"horizontals"`
	Lines                       []LineData       `json:"lines"`
	Arrows                      []ArrowData      `json:"arrows"`
}

// LineData describes a line on a pane.
type LineData struct {
	Name          string      `json:"name"`
	Data          []ScalarVal `json:"data"`
	Indicator     int         `json:"indicator"`
	Output        int         `json:"output"`
	Color         string      `json:"color"`
	Width         float64     `json:"width"`
	Dash          string      `json:"dash"`
	Interpolation string      `json:"interpolation"`
}

// ScalarVal is a time+value pair for chart data.
type ScalarVal struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

// BandData describes a band on a pane.
type BandData struct {
	Name          string    `json:"name"`
	Data          []BandVal `json:"data"`
	Indicator     int       `json:"indicator"`
	Output        int       `json:"output"`
	Color         string    `json:"color"`
	LegendColor   string    `json:"legendColor"`
	Interpolation string    `json:"interpolation"`
}

// BandVal is a time+upper+lower pair for chart data.
type BandVal struct {
	Time  string  `json:"time"`
	Upper float64 `json:"upper"`
	Lower float64 `json:"lower"`
}

// LineAreaData describes a line area on a pane.
type LineAreaData struct {
	Name          string      `json:"name"`
	Data          []ScalarVal `json:"data"`
	Indicator     int         `json:"indicator"`
	Output        int         `json:"output"`
	Value         float64     `json:"value"`
	Color         string      `json:"color"`
	LegendColor   string      `json:"legendColor"`
	Interpolation string      `json:"interpolation"`
}

// HeatmapData describes a heatmap on a pane.
type HeatmapData struct {
	Name           string       `json:"name"`
	Data           []HeatmapVal `json:"data"`
	Indicator      int          `json:"indicator"`
	Output         int          `json:"output"`
	Gradient       string       `json:"gradient"`
	InvertGradient bool         `json:"invertGradient"`
}

// HeatmapVal is a heatmap column for chart data.
type HeatmapVal struct {
	Time   string    `json:"time"`
	First  float64   `json:"first"`
	Last   float64   `json:"last"`
	Result float64   `json:"result"`
	Min    float64   `json:"min"`
	Max    float64   `json:"max"`
	Values []float64 `json:"values"`
}

// HorizontalData describes a horizontal line.
type HorizontalData struct {
	Value float64 `json:"value"`
	Color string  `json:"color"`
	Width float64 `json:"width"`
	Dash  string  `json:"dash"`
}

// ArrowData describes an arrow on a pane.
type ArrowData struct {
	Name      string `json:"name"`
	Down      bool   `json:"down"`
	Time      string `json:"time"`
	Indicator int    `json:"indicator"`
	Output    int    `json:"output"`
	Color     string `json:"color"`
}

// --- Default colors for cycling ---

var lineColors = []string{
	"#e41a1c", "#377eb8", "#4daf4a", "#984ea3", "#ff7f00",
	"#a65628", "#f781bf", "#999999", "#66c2a5", "#fc8d62",
}

var bandColors = []string{
	"rgba(0,255,0,0.3)", "rgba(0,0,255,0.3)", "rgba(255,0,0,0.3)",
	"rgba(128,0,128,0.3)", "rgba(0,128,128,0.3)",
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: iconf <settings.json> <output-name>\n")
		os.Exit(1)
	}

	settingsPath := os.Args[1]
	outputName := strings.TrimSuffix(strings.TrimSuffix(os.Args[2], ".json"), ".ts")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading settings file: %v\n", err)
		os.Exit(1)
	}

	var entries []settingsEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing settings file: %v\n", err)
		os.Exit(1)
	}

	indicators := make([]core.Indicator, 0, len(entries))
	for _, e := range entries {
		ind, err := factory.New(e.Identifier, string(e.Params))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating indicator %s: %v\n", e.Identifier, err)
			os.Exit(1)
		}
		indicators = append(indicators, ind)
	}

	bars := testBars()
	cfg := buildConfiguration(indicators, bars)

	// Write JSON.
	jsonPath := outputName + ".json"
	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", jsonPath, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s\n", jsonPath)

	// Write TypeScript.
	tsPath := outputName + ".ts"
	tsData := buildTypeScript(cfg, filepath.Base(outputName))
	if err := os.WriteFile(tsPath, []byte(tsData), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", tsPath, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s\n", tsPath)
}

// outputPlacement describes where to place an indicator output.
type outputPlacement struct {
	indicatorIdx int
	outputIdx    int
	mnemonic     string
	shape        shape.Shape
	pane         core.Pane
}

func buildConfiguration(indicators []core.Indicator, bars []entities.Bar) *Configuration {
	const timeFmt = "2006-01-02"

	// Collect output placements from descriptor registry.
	var placements []outputPlacement
	for i, ind := range indicators {
		meta := ind.Metadata()
		desc, ok := core.DescriptorOf(meta.Identifier)
		if !ok {
			// Fallback: treat all outputs as Own/Scalar.
			for j, om := range meta.Outputs {
				placements = append(placements, outputPlacement{
					indicatorIdx: i, outputIdx: j,
					mnemonic: om.Mnemonic, shape: om.Shape, pane: core.Own,
				})
			}
			continue
		}
		for j, od := range desc.Outputs {
			mnemonic := fmt.Sprintf("out[%d]", j)
			if j < len(meta.Outputs) {
				mnemonic = meta.Outputs[j].Mnemonic
			}
			placements = append(placements, outputPlacement{
				indicatorIdx: i, outputIdx: j,
				mnemonic: mnemonic, shape: od.Shape, pane: od.Pane,
			})
		}
	}

	// Run bars through indicators, collecting all outputs.
	// allOutputs[barIdx][indicatorIdx] = core.Output
	allOutputs := make([][]core.Output, len(bars))
	for i, bar := range bars {
		allOutputs[i] = make([]core.Output, len(indicators))
		for j, ind := range indicators {
			allOutputs[i][j] = ind.UpdateBar(&bar)
		}
	}

	// Build OHLCV data.
	ohlcvBars := make([]OhlcvBar, len(bars))
	for i, bar := range bars {
		ohlcvBars[i] = OhlcvBar{
			Time: bar.Time.Format(timeFmt), Open: bar.Open, High: bar.High,
			Low: bar.Low, Close: bar.Close, Volume: bar.Volume,
		}
	}

	// Group placements: price pane vs own panes.
	// For "Own" pane outputs, group by indicator index so all outputs of the same
	// indicator that share Pane=Own go into the same indicator pane.
	type ownPaneKey struct{ indicatorIdx int }
	ownPaneMap := make(map[ownPaneKey]int) // key -> indicatorPanes index
	var indicatorPanes []Pane

	pricePane := Pane{
		Height: "30%", ValueFormat: ",.2f", ValueMarginPercentageFactor: 0.01,
		Bands: []BandData{}, LineAreas: []LineAreaData{},
		Horizontals: []HorizontalData{}, Lines: []LineData{}, Arrows: []ArrowData{},
	}

	colorIdx := 0
	bandColorIdx := 0

	for _, p := range placements {
		switch {
		case p.pane == core.Price && p.shape == shape.Scalar:
			line := buildLineData(p, allOutputs, bars, timeFmt, lineColors[colorIdx%len(lineColors)])
			colorIdx++
			pricePane.Lines = append(pricePane.Lines, line)

		case p.pane == core.Price && p.shape == shape.Band:
			band := buildBandData(p, allOutputs, bars, timeFmt, bandColors[bandColorIdx%len(bandColors)])
			bandColorIdx++
			pricePane.Bands = append(pricePane.Bands, band)

		default:
			// Own pane (or fallback).
			key := ownPaneKey{p.indicatorIdx}
			idx, exists := ownPaneMap[key]
			if !exists {
				idx = len(indicatorPanes)
				ownPaneMap[key] = idx
				indicatorPanes = append(indicatorPanes, Pane{
					Height: "60", ValueFormat: ",.2f", ValueMarginPercentageFactor: 0.01,
					Bands: []BandData{}, LineAreas: []LineAreaData{},
					Horizontals: []HorizontalData{}, Lines: []LineData{}, Arrows: []ArrowData{},
				})
			}
			pane := &indicatorPanes[idx]

			switch p.shape {
			case shape.Heatmap:
				hm := buildHeatmapData(p, allOutputs, bars, timeFmt)
				pane.Heatmap = &hm
				pane.Height = "120"
				pane.ValueMarginPercentageFactor = 0

			case shape.Band:
				band := buildBandData(p, allOutputs, bars, timeFmt, bandColors[bandColorIdx%len(bandColors)])
				bandColorIdx++
				pane.Bands = append(pane.Bands, band)

			default: // Scalar or unknown → line.
				line := buildLineData(p, allOutputs, bars, timeFmt, lineColors[colorIdx%len(lineColors)])
				colorIdx++
				pane.Lines = append(pane.Lines, line)
			}
		}
	}

	return &Configuration{
		Width: "100%",
		NavigationPane: &NavigationPane{
			Height: 30, HasLine: true, HasArea: false, HasTimeAxis: true, TimeTicks: 0,
		},
		HeightNavigationPane: 30,
		TimeAnnotationFormat: "%Y-%m-%d",
		AxisLeft:             true,
		AxisRight:            false,
		Margin:               Margin{Left: 0, Top: 10, Right: 20, Bottom: 0},
		Ohlcv:                OhlcvData{Name: "TEST", Data: ohlcvBars, Candlesticks: true},
		PricePane:            pricePane,
		IndicatorPanes:       indicatorPanes,
		Crosshair:            false,
		VolumeInPricePane:    true,
		MenuVisible:          true,
		DownloadSvgVisible:   true,
	}
}

func buildLineData(p outputPlacement, allOutputs [][]core.Output, bars []entities.Bar, timeFmt, color string) LineData {
	data := make([]ScalarVal, 0, len(bars))
	for i := range bars {
		out := allOutputs[i][p.indicatorIdx]
		if p.outputIdx < len(out) {
			if s, ok := out[p.outputIdx].(entities.Scalar); ok && !math.IsNaN(s.Value) {
				data = append(data, ScalarVal{Time: s.Time.Format(timeFmt), Value: s.Value})
			}
		}
	}
	return LineData{
		Name: p.mnemonic, Data: data, Indicator: p.indicatorIdx, Output: p.outputIdx,
		Color: color, Width: 1, Dash: "", Interpolation: "natural",
	}
}

func buildBandData(p outputPlacement, allOutputs [][]core.Output, bars []entities.Bar, timeFmt, color string) BandData {
	data := make([]BandVal, 0, len(bars))
	for i := range bars {
		out := allOutputs[i][p.indicatorIdx]
		if p.outputIdx < len(out) {
			if b, ok := out[p.outputIdx].(*outputs.Band); ok && !b.IsEmpty() {
				data = append(data, BandVal{Time: b.Time.Format(timeFmt), Upper: b.Upper, Lower: b.Lower})
			}
		}
	}
	return BandData{
		Name: p.mnemonic, Data: data, Indicator: p.indicatorIdx, Output: p.outputIdx,
		Color: color, LegendColor: color, Interpolation: "natural",
	}
}

func buildHeatmapData(p outputPlacement, allOutputs [][]core.Output, bars []entities.Bar, timeFmt string) HeatmapData {
	data := make([]HeatmapVal, 0, len(bars))
	for i := range bars {
		out := allOutputs[i][p.indicatorIdx]
		if p.outputIdx < len(out) {
			if h, ok := out[p.outputIdx].(*outputs.Heatmap); ok && !h.IsEmpty() {
				data = append(data, HeatmapVal{
					Time: h.Time.Format(timeFmt), First: h.ParameterFirst, Last: h.ParameterLast,
					Result: h.ParameterResolution, Min: h.ValueMin, Max: h.ValueMax, Values: h.Values,
				})
			}
		}
	}
	return HeatmapData{
		Name: p.mnemonic, Data: data, Indicator: p.indicatorIdx, Output: p.outputIdx,
		Gradient: "Viridis", InvertGradient: false,
	}
}

// --- TypeScript generation ---

func buildTypeScript(cfg *Configuration, baseName string) string {
	var sb strings.Builder

	sb.WriteString("// Auto-generated chart configuration.\n")
	sb.WriteString("// eslint-disable-next-line\n")
	sb.WriteString("import { Configuration } from '../ohlcv-chart/template/configuration';\n")
	sb.WriteString("import { Scalar, Band, Heatmap, Bar } from '../ohlcv-chart/template/types';\n\n")

	// Emit OHLCV data array.
	ohlcvVar := sanitizeVarName(baseName) + "Ohlcv"
	sb.WriteString(fmt.Sprintf("export const %s: Bar[] = ", ohlcvVar))
	writeJSONArray(&sb, cfg.Ohlcv.Data)
	sb.WriteString(";\n\n")

	// Track variable names for data references.
	lineVars := make(map[*LineData]varRef)
	bandVars := make(map[*BandData]varRef)
	heatmapVars := make(map[*HeatmapData]varRef)
	lineAreaVars := make(map[*LineAreaData]varRef)

	varCounter := 0
	emitScalarArray := func(sb *strings.Builder, mnemonic string, data []ScalarVal) string {
		varCounter++
		vn := sanitizeVarName(fmt.Sprintf("%s_%s_%d", baseName, mnemonic, varCounter))
		sb.WriteString(fmt.Sprintf("export const %s: Scalar[] = ", vn))
		writeJSONArray(sb, data)
		sb.WriteString(";\n\n")
		return vn
	}
	emitBandArray := func(sb *strings.Builder, mnemonic string, data []BandVal) string {
		varCounter++
		vn := sanitizeVarName(fmt.Sprintf("%s_%s_%d", baseName, mnemonic, varCounter))
		sb.WriteString(fmt.Sprintf("export const %s: Band[] = ", vn))
		writeJSONArray(sb, data)
		sb.WriteString(";\n\n")
		return vn
	}
	emitHeatmapArray := func(sb *strings.Builder, mnemonic string, data []HeatmapVal) string {
		varCounter++
		vn := sanitizeVarName(fmt.Sprintf("%s_%s_%d", baseName, mnemonic, varCounter))
		sb.WriteString(fmt.Sprintf("export const %s: Heatmap[] = ", vn))
		writeJSONArray(sb, data)
		sb.WriteString(";\n\n")
		return vn
	}

	// Emit price pane data arrays.
	for i := range cfg.PricePane.Lines {
		l := &cfg.PricePane.Lines[i]
		vn := emitScalarArray(&sb, l.Name, l.Data)
		lineVars[l] = varRef{vn, "Scalar[]"}
	}
	for i := range cfg.PricePane.Bands {
		b := &cfg.PricePane.Bands[i]
		vn := emitBandArray(&sb, b.Name, b.Data)
		bandVars[b] = varRef{vn, "Band[]"}
	}
	for i := range cfg.PricePane.LineAreas {
		la := &cfg.PricePane.LineAreas[i]
		vn := emitScalarArray(&sb, la.Name, la.Data)
		lineAreaVars[la] = varRef{vn, "Scalar[]"}
	}

	// Emit indicator pane data arrays.
	for pi := range cfg.IndicatorPanes {
		pane := &cfg.IndicatorPanes[pi]
		for i := range pane.Lines {
			l := &pane.Lines[i]
			vn := emitScalarArray(&sb, l.Name, l.Data)
			lineVars[l] = varRef{vn, "Scalar[]"}
		}
		for i := range pane.Bands {
			b := &pane.Bands[i]
			vn := emitBandArray(&sb, b.Name, b.Data)
			bandVars[b] = varRef{vn, "Band[]"}
		}
		for i := range pane.LineAreas {
			la := &pane.LineAreas[i]
			vn := emitScalarArray(&sb, la.Name, la.Data)
			lineAreaVars[la] = varRef{vn, "Scalar[]"}
		}
		if pane.Heatmap != nil {
			hm := pane.Heatmap
			vn := emitHeatmapArray(&sb, hm.Name, hm.Data)
			heatmapVars[hm] = varRef{vn, "Heatmap[]"}
		}
	}

	// Emit configuration object.
	sb.WriteString(fmt.Sprintf("export const %sConfig: Configuration = {\n", sanitizeVarName(baseName)))
	sb.WriteString(fmt.Sprintf("  width: %q,\n", cfg.Width))

	if cfg.NavigationPane != nil {
		sb.WriteString("  navigationPane: {\n")
		sb.WriteString(fmt.Sprintf("    height: %d, hasLine: %t, hasArea: %t, hasTimeAxis: %t, timeTicks: %d,\n",
			cfg.NavigationPane.Height, cfg.NavigationPane.HasLine, cfg.NavigationPane.HasArea,
			cfg.NavigationPane.HasTimeAxis, cfg.NavigationPane.TimeTicks))
		sb.WriteString("  },\n")
	}

	sb.WriteString(fmt.Sprintf("  heightNavigationPane: %d,\n", cfg.HeightNavigationPane))
	sb.WriteString(fmt.Sprintf("  timeAnnotationFormat: %q,\n", cfg.TimeAnnotationFormat))
	sb.WriteString(fmt.Sprintf("  axisLeft: %t,\n", cfg.AxisLeft))
	sb.WriteString(fmt.Sprintf("  axisRight: %t,\n", cfg.AxisRight))
	sb.WriteString(fmt.Sprintf("  margin: { left: %d, top: %d, right: %d, bottom: %d },\n",
		cfg.Margin.Left, cfg.Margin.Top, cfg.Margin.Right, cfg.Margin.Bottom))
	sb.WriteString(fmt.Sprintf("  ohlcv: { name: %q, data: %s, candlesticks: %t },\n",
		cfg.Ohlcv.Name, ohlcvVar, cfg.Ohlcv.Candlesticks))

	// Price pane.
	sb.WriteString("  pricePane: {\n")
	writeTSPane(&sb, &cfg.PricePane, lineVars, bandVars, lineAreaVars, heatmapVars, "    ")
	sb.WriteString("  },\n")

	// Indicator panes.
	sb.WriteString("  indicatorPanes: [\n")
	for i := range cfg.IndicatorPanes {
		sb.WriteString("    {\n")
		writeTSPane(&sb, &cfg.IndicatorPanes[i], lineVars, bandVars, lineAreaVars, heatmapVars, "      ")
		sb.WriteString("    },\n")
	}
	sb.WriteString("  ],\n")

	sb.WriteString(fmt.Sprintf("  crosshair: %t,\n", cfg.Crosshair))
	sb.WriteString(fmt.Sprintf("  volumeInPricePane: %t,\n", cfg.VolumeInPricePane))
	sb.WriteString(fmt.Sprintf("  menuVisible: %t,\n", cfg.MenuVisible))
	sb.WriteString(fmt.Sprintf("  downloadSvgVisible: %t,\n", cfg.DownloadSvgVisible))
	sb.WriteString("};\n")

	return sb.String()
}

type varRef struct{ varName, typeName string }

func writeTSPane(sb *strings.Builder, pane *Pane,
	lineVars map[*LineData]varRef, bandVars map[*BandData]varRef,
	lineAreaVars map[*LineAreaData]varRef, heatmapVars map[*HeatmapData]varRef,
	indent string,
) {
	sb.WriteString(fmt.Sprintf("%sheight: %q, valueFormat: %q, valueMarginPercentageFactor: %g,\n",
		indent, pane.Height, pane.ValueFormat, pane.ValueMarginPercentageFactor))

	if pane.Heatmap != nil {
		hm := pane.Heatmap
		vr := heatmapVars[hm]
		sb.WriteString(fmt.Sprintf("%sheatmap: { name: %q, data: %s, indicator: %d, output: %d, gradient: %q, invertGradient: %t },\n",
			indent, hm.Name, vr.varName, hm.Indicator, hm.Output, hm.Gradient, hm.InvertGradient))
	}

	// Bands.
	sb.WriteString(indent + "bands: [")
	if len(pane.Bands) > 0 {
		sb.WriteString("\n")
		for i := range pane.Bands {
			b := &pane.Bands[i]
			vr := bandVars[b]
			sb.WriteString(fmt.Sprintf("%s  { name: %q, data: %s, indicator: %d, output: %d, color: %q, legendColor: %q, interpolation: %q },\n",
				indent, b.Name, vr.varName, b.Indicator, b.Output, b.Color, b.LegendColor, b.Interpolation))
		}
		sb.WriteString(indent)
	}
	sb.WriteString("],\n")

	// LineAreas.
	sb.WriteString(indent + "lineAreas: [")
	if len(pane.LineAreas) > 0 {
		sb.WriteString("\n")
		for i := range pane.LineAreas {
			la := &pane.LineAreas[i]
			vr := lineAreaVars[la]
			sb.WriteString(fmt.Sprintf("%s  { name: %q, data: %s, indicator: %d, output: %d, value: %g, color: %q, legendColor: %q, interpolation: %q },\n",
				indent, la.Name, vr.varName, la.Indicator, la.Output, la.Value, la.Color, la.LegendColor, la.Interpolation))
		}
		sb.WriteString(indent)
	}
	sb.WriteString("],\n")

	// Horizontals.
	sb.WriteString(indent + "horizontals: [")
	if len(pane.Horizontals) > 0 {
		sb.WriteString("\n")
		for _, h := range pane.Horizontals {
			sb.WriteString(fmt.Sprintf("%s  { value: %g, color: %q, width: %g, dash: %q },\n",
				indent, h.Value, h.Color, h.Width, h.Dash))
		}
		sb.WriteString(indent)
	}
	sb.WriteString("],\n")

	// Lines.
	sb.WriteString(indent + "lines: [")
	if len(pane.Lines) > 0 {
		sb.WriteString("\n")
		for i := range pane.Lines {
			l := &pane.Lines[i]
			vr := lineVars[l]
			sb.WriteString(fmt.Sprintf("%s  { name: %q, data: %s, indicator: %d, output: %d, color: %q, width: %g, dash: %q, interpolation: %q },\n",
				indent, l.Name, vr.varName, l.Indicator, l.Output, l.Color, l.Width, l.Dash, l.Interpolation))
		}
		sb.WriteString(indent)
	}
	sb.WriteString("],\n")

	// Arrows.
	sb.WriteString(indent + "arrows: [],\n")
}

func writeJSONArray(sb *strings.Builder, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		sb.WriteString("[]")
		return
	}
	sb.Write(data)
}

func sanitizeVarName(s string) string {
	var sb strings.Builder
	capitalize := false
	for i, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
			if capitalize {
				if c >= 'a' && c <= 'z' {
					c = c - 'a' + 'A'
				}
				capitalize = false
			}
			sb.WriteRune(c)
		case c >= '0' && c <= '9':
			if i == 0 {
				sb.WriteRune('_')
			}
			sb.WriteRune(c)
			capitalize = false
		default:
			capitalize = true
		}
	}
	return sb.String()
}

// testBars returns 252 embedded test bars constructed from the standard TA-Lib
// daily reference data used across the indicator unit tests.
func testBars() []entities.Bar {
	highs := testHighs()
	lows := testLows()
	closes := testCloses()
	volumes := testVolumes()

	bars := make([]entities.Bar, len(closes))
	baseTime := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	for i := range closes {
		openPrice := closes[0]
		if i > 0 {
			openPrice = closes[i-1]
		}

		bars[i] = entities.Bar{
			Time:   baseTime.AddDate(0, 0, i),
			Open:   openPrice,
			High:   highs[i],
			Low:    lows[i],
			Close:  closes[i],
			Volume: volumes[i],
		}
	}

	return bars
}

func testHighs() []float64 {
	return []float64{
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
	}
}

func testLows() []float64 {
	return []float64{
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
	}
}

func testCloses() []float64 {
	return []float64{
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
	}
}

func testVolumes() []float64 {
	return []float64{
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
	}
}
