// lvt runs local CSV/JSON Lars von Thienen research workflows.
package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"zpano/indicators/larsvonthienen/annualcyclemodel"
	"zpano/indicators/larsvonthienen/cyclescanner"
	"zpano/indicators/larsvonthienen/dynamiccycleexplorer"
	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/harmonicenergysuperpositionwave"
)

var workflows = []string{"complete-cycle-scanner", "dominant-market-vibration", "dynamic-cycle-explorer", "cycle-phase-score", "phasing-oscillator", "annual-cycle-pattern-decoder", "harmonic-energy-superposition-wave"}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(arguments []string) error {
	if len(arguments) < 2 {
		return fmt.Errorf("usage: lvt <workflow> <analyze|fit|build|replay|phase> [options]\n%s", strings.Join(workflows, "\n"))
	}
	slug, command := arguments[0], arguments[1]
	known := false
	for _, name := range workflows {
		known = known || slug == name
	}
	if !known {
		return fmt.Errorf("unknown workflow %s", slug)
	}
	args := []string{}
	for i := 2; i < len(arguments); i++ {
		if arguments[i] == "--models" {
			if i+6 >= len(arguments) {
				return errors.New("--models requires six paths")
			}
			args = append(args, "--models", strings.Join(arguments[i+1:i+7], ","))
			i += 6
		} else {
			args = append(args, arguments[i])
		}
	}
	fs := flag.NewFlagSet(slug+" "+command, flag.ContinueOnError)
	input := fs.String("input", "", "date,close CSV")
	calendarPath := fs.String("calendar", "", "date-column CSV")
	modelPath := fs.String("model", "", "saved model JSON")
	modelsPath := fs.String("models", "", "six model paths")
	historyPath := fs.String("history", "", "historical date,close CSV")
	jsonOut := fs.String("json-out", "", "report JSON")
	modelOut := fs.String("model-out", "", "model JSON")
	waveOut := fs.String("wave-out", "", "date,value CSV")
	seriesOut := fs.String("series-out", "", "historical series CSV")
	startText := fs.String("start", "", "inclusive date")
	endText := fs.String("end", "", "inclusive date")
	asOfText := fs.String("as-of", time.Now().UTC().Format("2006-01-02"), "fit date")
	year := fs.Int("year", 0, "completed source year")
	targetYear := fs.Int("target-year", 0, "target year")
	confirm := fs.Bool("confirm-complete-year", false, "assert complete source-year sessions")
	confirmYears := fs.Bool("confirm-complete-years", false, "assert complete historical sessions")
	norm := fs.String("normalization", "rms", "rms or none")
	policy := fs.String("calendar-policy", "extend", "extend or strict")
	invert := fs.Bool("invert", false, "invert frozen average")
	rolling := fs.Bool("rolling", false, "causal rolling calculation")
	window := fs.Int("window", 200, "rolling window")
	minPeriod := fs.Int("min-period", 5, "minimum period")
	maxDefault := 80
	if slug == "annual-cycle-pattern-decoder" {
		maxDefault = 125
	}
	maxPeriod := fs.Int("max-period", maxDefault, "maximum period")
	smoothing := fs.Float64("smoothing", 1600, "HP lambda")
	threshold := fs.Float64("genuine-threshold", 49, "strict Bartels threshold")
	phaseCycles := fs.Float64("phase-cycles", 2, "recent phase cycles")
	topK := fs.Int("top-k", 3, "rank count")
	distance := fs.Int("peak-distance", 2, "period suppression distance")
	segments := fs.Int("max-segments", 5, "Bartels segments")
	pool := fs.Int("candidate-pool", 10, "annual pool <=16")
	maxComponents := fs.Int("max-components", 7, "annual subset <=7")
	horizon := fs.Int("horizon", 0, "static bars-ahead projection")
	shortMin := fs.Int("short-min", 5, "short range")
	shortMax := fs.Int("short-max", 30, "short range")
	longMin := fs.Int("long-min", 31, "long range")
	longMax := fs.Int("long-max", 80, "long range")
	radians := fs.Float64("radians", math.NaN(), "sine phase")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	if command == "phase" && slug == "cycle-phase-score" {
		if math.IsNaN(*radians) || math.IsInf(*radians, 0) {
			return errors.New("phase must be finite")
		}
		return printJSON(map[string]any{"value": harmoniccycle.PhaseState(*radians), "status": phaseStatus(*radians)})
	}
	asOf, err := parseDate(*asOfText)
	if err != nil {
		return err
	}
	annualParams := &annualcyclemodel.Params{MinPeriod: *minPeriod, MaxPeriod: *maxPeriod, Smoothing: *smoothing, GenuineThreshold: *threshold, CandidatePool: *pool, MaxComponents: *maxComponents, PeakDistance: *distance}
	if slug == "annual-cycle-pattern-decoder" {
		var model annualcyclemodel.Model
		if command == "fit" {
			if *modelOut == "" || *waveOut == "" {
				return errors.New("fit requires --model-out and --wave-out")
			}
			records, err := readPrices(*input)
			if err != nil {
				return err
			}
			var calendar []time.Time
			if *calendarPath != "" {
				calendar, err = readCalendar(*calendarPath)
				if err != nil {
					return err
				}
			}
			model, err = annualcyclemodel.DecodeAnnual(records, *year, asOf, calendar, *confirm, annualParams)
			if err != nil {
				return err
			}
			if err = saveJSON(*modelOut, model); err != nil {
				return err
			}
		} else if command == "replay" {
			if err = loadJSON(*modelPath, &model); err != nil {
				return err
			}
			if err = model.Validate(); err != nil {
				return err
			}
		} else {
			return errors.New("annual workflow supports fit or replay")
		}
		if len(model.Dates) != model.SampleCount {
			return errors.New("replay requires model dates")
		}
		values := make([]float64, model.SampleCount)
		for i := range values {
			values[i] = annualcyclemodel.AnnualWave(model, float64(i))
		}
		if err = saveWave(*waveOut, model.Dates, values); err != nil {
			return err
		}
		return printJSON(map[string]any{"year": model.Year, "correlation": model.Correlation, "components": len(model.Components), "sample_count": len(values)})
	}
	if slug == "harmonic-energy-superposition-wave" {
		var model harmonicenergysuperpositionwave.Model
		var dates []time.Time
		var models []annualcyclemodel.Model
		if command == "build" {
			if *modelOut == "" || *waveOut == "" {
				return errors.New("build requires --model-out and --wave-out")
			}
			dates, err = readCalendar(*calendarPath)
			if err != nil {
				return err
			}
			for _, day := range dates {
				if day.Year() != *targetYear {
					return errors.New("target calendar must match target year")
				}
			}
			if (*modelsPath == "") == (*historyPath == "") {
				return errors.New("provide --models or --history")
			}
			if *modelsPath != "" {
				for _, path := range strings.Split(*modelsPath, ",") {
					var annual annualcyclemodel.Model
					if err = loadJSON(path, &annual); err != nil {
						return err
					}
					models = append(models, annual)
				}
			} else {
				records, err := readPrices(*historyPath)
				if err != nil {
					return err
				}
				for _, lag := range harmonicenergysuperpositionwave.MetonicLags {
					annual, err := annualcyclemodel.DecodeAnnual(records, *targetYear-lag, asOf, nil, *confirmYears, annualcyclemodel.DefaultParams())
					if err != nil {
						return err
					}
					models = append(models, annual)
				}
			}
		} else if command == "replay" {
			if err = loadJSON(*modelPath, &model); err != nil {
				return err
			}
			if err = model.Validate(); err != nil {
				return err
			}
			models = model.AnnualModels
			*targetYear = model.TargetYear
			*norm = model.Normalization
			*policy = model.CalendarPolicy
			*invert = model.Invert
			for _, d := range model.Dates {
				day, err := parseDate(d)
				if err != nil {
					return err
				}
				dates = append(dates, day)
			}
		} else {
			return errors.New("superposition workflow supports build or replay")
		}
		values, built, err := harmonicenergysuperpositionwave.SuperpositionValues(models, *targetYear, len(dates), *norm, *policy, *invert)
		if err != nil {
			return err
		}
		built.Dates = dateStrings(dates)
		if command == "build" {
			if err = saveJSON(*modelOut, built); err != nil {
				return err
			}
		}
		if err = saveWave(*waveOut, built.Dates, values); err != nil {
			return err
		}
		return printJSON(map[string]any{"target_year": built.TargetYear, "source_years": built.SourceYears, "sample_count": len(values)})
	}
	if command != "analyze" {
		return errors.New("streaming workflows support analyze")
	}
	if *horizon < 0 {
		return errors.New("horizon must be nonnegative")
	}
	records, err := readPrices(*input)
	if err != nil {
		return err
	}
	var start, end time.Time
	if *startText != "" {
		start, err = parseDate(*startText)
		if err != nil {
			return err
		}
	}
	if *endText != "" {
		end, err = parseDate(*endText)
		if err != nil {
			return err
		}
	}
	if !start.IsZero() && !end.IsZero() && start.After(end) {
		return errors.New("start must not exceed end")
	}
	selected := []annualcyclemodel.Record{}
	for _, r := range records {
		if (start.IsZero() || !r.Day.Before(start)) && (end.IsZero() || !r.Day.After(end)) {
			selected = append(selected, r)
		}
	}
	records = selected
	if len(records) < 6 {
		return errors.New("selected range needs >=6 prices")
	}
	prices := make([]float64, len(records))
	dates := make([]time.Time, len(records))
	for i, r := range records {
		prices[i] = r.Close
		dates[i] = r.Day
	}
	n := len(prices)
	if !*rolling {
		*window = n
	}
	config := cyclescanner.ScanConfig{Window: *window, MinPeriod: *minPeriod, MaxPeriod: *maxPeriod, Smoothing: *smoothing, GenuineThreshold: *threshold, PhaseCycles: *phaseCycles, TopK: *topK, PeakDistance: *distance, MaxSegments: *segments}
	if err = config.Validate(); err != nil {
		return err
	}
	scan := cyclescanner.EmptyScan(config)
	tail := prices
	if len(tail) > config.Window {
		tail = tail[len(tail)-config.Window:]
	}
	if len(tail) == config.Window {
		scan = cyclescanner.ScanCycles(tail, config)
	}
	report := scannerReport(scan, config, dates[max(0, n-config.Window):])
	series := []map[string]any{}
	shortConfig, longConfig := config, config
	shortConfig.MinPeriod = *shortMin
	shortConfig.MaxPeriod = *shortMax
	longConfig.MinPeriod = *longMin
	longConfig.MaxPeriod = *longMax
	shortConfig.TopK = 1
	longConfig.TopK = 1
	if slug == "phasing-oscillator" {
		if *shortMax >= *longMin {
			return errors.New("phasing ranges must be disjoint")
		}
		if err = shortConfig.Validate(); err != nil {
			return err
		}
		if err = longConfig.Validate(); err != nil {
			return err
		}
	}
	calculate := func(values []float64, seen int) (map[string]any, cyclescanner.ScanOutput, cyclescanner.ScanOutput) {
		current := cyclescanner.EmptyScan(config)
		if len(values) >= config.Window {
			values = values[len(values)-config.Window:]
			current = cyclescanner.ScanCycles(values, config)
		}
		short, long := cyclescanner.EmptyScan(shortConfig), cyclescanner.EmptyScan(longConfig)
		output := map[string]any{}
		cycle := current.Cycles[0]
		switch slug {
		case "complete-cycle-scanner":
			output = flattenScan(current, config)
		case "dominant-market-vibration":
			output = clean(cycle).(map[string]any)
		case "cycle-phase-score":
			output = map[string]any{"value": harmoniccycle.PhaseState(cycle.Phase), "period": cycle.Period, "phase": cycle.Phase, "status": phaseStatus(cycle.Phase)}
		case "dynamic-cycle-explorer":
			highlighter, index, fitted := math.NaN(), math.NaN(), math.NaN()
			if math.IsNaN(cycle.Period) == false {
				length := int(cycle.Period)
				delay := length / 2
				center := len(values) - 1 - delay
				highlighter = dynamiccycleexplorer.CenteredMean(values, center, max(2, int(math.Floor(cycle.Period/2+0.5)))) - dynamiccycleexplorer.CenteredMean(values, center, length)
				if !math.IsNaN(highlighter) {
					index = float64(seen - 1 - delay)
					fitted = cyclescanner.ModeledValue(cycle, -float64(delay))
				}
			}
			output = map[string]any{"value": cycle.Value, "period": cycle.Period, "phase": cycle.Phase, "eta_top": cycle.EtaTop, "eta_bottom": cycle.EtaBottom, "highlighter": highlighter, "highlighter_index": index, "fitted_at_highlighter": fitted}
		case "phasing-oscillator":
			if len(values) >= config.Window {
				short = cyclescanner.ScanCycles(values, shortConfig)
				long = cyclescanner.ScanCycles(values, longConfig)
			}
			a, b := harmoniccycle.PhaseState(short.Cycles[0].Phase), harmoniccycle.PhaseState(long.Cycles[0].Phase)
			output = map[string]any{"value": a + b, "short_score": a, "long_score": b, "short_period": short.Cycles[0].Period, "long_period": long.Cycles[0].Period}
		}
		return output, short, long
	}
	output, short, long := calculate(prices, n)
	report["output"] = output
	if *rolling || slug == "phasing-oscillator" {
		for i := range prices {
			row, _, _ := calculate(prices[:i+1], i+1)
			row["date"] = dates[i].Format("2006-01-02")
			series = append(series, row)
		}
	} else {
		cycle := scan.Cycles[0]
		for i, d := range dates {
			row := map[string]any{"date": d.Format("2006-01-02"), "fitted": cyclescanner.ModeledValue(cycle, float64(i-(n-1)))}
			if slug == "dynamic-cycle-explorer" {
				value := math.NaN()
				if !math.IsNaN(cycle.Period) {
					length := int(cycle.Period)
					value = dynamiccycleexplorer.CenteredMean(prices, i, max(2, int(math.Floor(float64(length)/2+0.5)))) - dynamiccycleexplorer.CenteredMean(prices, i, length)
				}
				row["highlighter"] = value
			}
			series = append(series, row)
		}
	}
	projection := []map[string]any{}
	for h := 1; h <= *horizon; h++ {
		value := cyclescanner.ModeledValue(scan.Cycles[0], float64(h))
		row := map[string]any{"bars_ahead": h}
		if slug == "cycle-phase-score" {
			c := scan.Cycles[0]
			value = harmoniccycle.PhaseState(c.Phase + 2*math.Pi*float64(h)/c.Period)
		}
		if slug == "phasing-oscillator" {
			a, b := short.Cycles[0], long.Cycles[0]
			x, y := harmoniccycle.PhaseState(a.Phase+2*math.Pi*float64(h)/a.Period), harmoniccycle.PhaseState(b.Phase+2*math.Pi*float64(h)/b.Period)
			value = x + y
			row["short_score"] = x
			row["long_score"] = y
		}
		row["value"] = value
		projection = append(projection, row)
	}
	report["projection"] = projection
	if slug == "phasing-oscillator" {
		report["short_scan"] = scannerReport(short, shortConfig, dates[max(0, n-config.Window):])
		report["long_scan"] = scannerReport(long, longConfig, dates[max(0, n-config.Window):])
		report["phasing_config"] = map[string]any{"short": clean(shortConfig), "long": clean(longConfig)}
	}
	if *seriesOut != "" {
		if err = saveSeries(*seriesOut, series); err != nil {
			return err
		}
	}
	if *jsonOut != "" {
		if err = saveJSON(*jsonOut, report); err != nil {
			return err
		}
	}
	return printJSON(report)
}

func parseDate(text string) (time.Time, error) { return time.Parse("2006-01-02", text) }
func readCSV(path string) ([][]string, error) {
	if path == "" {
		return nil, errors.New("missing CSV path")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return csv.NewReader(f).ReadAll()
}
func readPrices(path string) ([]annualcyclemodel.Record, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, errors.New("empty price CSV")
	}
	dateCol, closeCol := -1, -1
	for i, h := range rows[0] {
		if h == "date" {
			dateCol = i
		}
		if h == "close" {
			closeCol = i
		}
	}
	if dateCol < 0 || closeCol < 0 {
		return nil, errors.New("CSV requires date,close columns")
	}
	records := []annualcyclemodel.Record{}
	var previous time.Time
	for _, row := range rows[1:] {
		day, err := parseDate(row[dateCol])
		if err != nil {
			return nil, err
		}
		price, err := strconv.ParseFloat(row[closeCol], 64)
		if err != nil || math.IsNaN(price) || math.IsInf(price, 0) {
			return nil, errors.New("CSV closes must be finite")
		}
		if !day.After(previous) {
			return nil, errors.New("dates must increase strictly")
		}
		records = append(records, annualcyclemodel.Record{Day: day, Close: price})
		previous = day
	}
	return records, nil
}
func readCalendar(path string) ([]time.Time, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, errors.New("empty calendar CSV")
	}
	column := -1
	for i, h := range rows[0] {
		if h == "date" {
			column = i
		}
	}
	if column < 0 {
		return nil, errors.New("calendar requires date column")
	}
	var dates []time.Time
	var previous time.Time
	for _, row := range rows[1:] {
		day, err := parseDate(row[column])
		if err != nil {
			return nil, err
		}
		if !day.After(previous) {
			return nil, errors.New("dates must increase strictly")
		}
		dates = append(dates, day)
		previous = day
	}
	return dates, nil
}
func dateStrings(dates []time.Time) []string {
	output := make([]string, len(dates))
	for i, d := range dates {
		output[i] = d.Format("2006-01-02")
	}
	return output
}
func snake(text string) string {
	var b strings.Builder
	for i, c := range text {
		if unicode.IsUpper(c) && i > 0 {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(c))
	}
	return b.String()
}

// clean maps unavailable numerical fields to standard JSON nulls.
func clean(value any) any { return cleanValue(reflect.ValueOf(value)) }
func cleanValue(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return cleanValue(v.Elem())
	}
	switch v.Kind() {
	case reflect.Float64, reflect.Float32:
		x := v.Float()
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return nil
		}
		return x
	case reflect.Struct:
		m := map[string]any{}
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			key := strings.Split(f.Tag.Get("json"), ",")[0]
			if key == "-" {
				continue
			}
			if key == "" {
				key = snake(f.Name)
			}
			if strings.Contains(f.Tag.Get("json"), "omitempty") && v.Field(i).IsZero() {
				continue
			}
			m[key] = cleanValue(v.Field(i))
		}
		return m
	case reflect.Map:
		m := map[string]any{}
		iter := v.MapRange()
		for iter.Next() {
			m[iter.Key().String()] = cleanValue(iter.Value())
		}
		return m
	case reflect.Slice, reflect.Array:
		a := make([]any, v.Len())
		for i := range a {
			a[i] = cleanValue(v.Index(i))
		}
		return a
	default:
		return v.Interface()
	}
}
func saveJSON(path string, value any) error {
	data, err := json.MarshalIndent(clean(value), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
func printJSON(value any) error {
	data, err := json.Marshal(clean(value))
	if err == nil {
		fmt.Println(string(data))
	}
	return err
}
func loadJSON(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}
func saveWave(path string, dates []string, values []float64) error {
	if path == "" || len(dates) != len(values) {
		return errors.New("wave-out and matching calendar required")
	}
	rows := make([][]string, len(values)+1)
	rows[0] = []string{"date", "value"}
	for i, v := range values {
		rows[i+1] = []string{dates[i], strconv.FormatFloat(v, 'g', 15, 64)}
	}
	return writeCSV(path, rows)
}
func writeCSV(path string, rows [][]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	return writer.WriteAll(rows)
}
func saveSeries(path string, series []map[string]any) error {
	if len(series) == 0 {
		return errors.New("empty series")
	}
	keys := []string{"date"}
	for key := range series[0] {
		if key != "date" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys[1:])
	rows := [][]string{keys}
	for _, r := range series {
		row := []string{}
		for _, k := range keys {
			v := clean(r[k])
			text := ""
			if v != nil {
				switch x := v.(type) {
				case float64:
					text = strconv.FormatFloat(x, 'g', 15, 64)
				default:
					text = fmt.Sprint(x)
				}
			}
			row = append(row, text)
		}
		rows = append(rows, row)
	}
	return writeCSV(path, rows)
}
func phaseStatus(phase float64) string {
	if math.IsNaN(phase) {
		return "unavailable"
	}
	angle := math.Mod(phase, 2*math.Pi)
	if angle < 0 {
		angle += 2 * math.Pi
	}
	if angle < math.Pi/2 {
		return "rising toward top"
	}
	if angle < math.Pi {
		return "falling from top"
	}
	if angle < 3*math.Pi/2 {
		return "falling toward bottom"
	}
	return "rising from bottom"
}

func flattenScan(scan cyclescanner.ScanOutput, c cyclescanner.ScanConfig) map[string]any {
	output := map[string]any{"accepted_count": scan.AcceptedCount}
	for i := range scan.Amplitude {
		output["amplitude_p"+strconv.Itoa(c.MinPeriod+i)] = scan.Amplitude[i]
		output["genuine_p"+strconv.Itoa(c.MinPeriod+i)] = scan.GenuinePercent[i]
	}
	for i, cycle := range scan.Cycles {
		for field, value := range clean(cycle).(map[string]any) {
			output["rank"+strconv.Itoa(i+1)+"_"+field] = value
		}
	}
	return output
}

func scannerReport(scan cyclescanner.ScanOutput, c cyclescanner.ScanConfig, dates []time.Time) map[string]any {
	periods := make([]int, c.MaxPeriod-c.MinPeriod+1)
	for i := range periods {
		periods[i] = c.MinPeriod + i
	}
	return map[string]any{
		"schema": "lvt-cycle-scan", "version": 1, "research_reconstruction": true,
		"start_date": dates[0].Format("2006-01-02"), "end_date": dates[len(dates)-1].Format("2006-01-02"),
		"sample_count": len(dates), "config": clean(c),
		"phase_convention": "sine radians at analysis endpoint",
		"validation_input": "raw prices, most recent complete segments",
		"accepted_count":   scan.AcceptedCount, "cycles": clean(scan.Cycles),
		"spectrum": map[string]any{"periods": periods, "amplitude": scan.Amplitude, "genuine_percent": scan.GenuinePercent},
	}
}
