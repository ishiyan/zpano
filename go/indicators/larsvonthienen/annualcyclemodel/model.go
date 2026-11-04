// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package annualcyclemodel

import (
	"fmt"
	"math"
	"sort"
	"time"

	"zpano/indicators/larsvonthienen/cyclescanner"
	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/hodrickprescottsmoothinganddetrending"
)

// Params configures a completed-year fit. No target-year prices enter its model.
type Params struct {
	MinPeriod, MaxPeriod                       int
	Smoothing, GenuineThreshold                float64
	CandidatePool, MaxComponents, PeakDistance int
}

// DefaultParams returns the documented annual defaults.
func DefaultParams() *Params { return &Params{5, 125, 1600, 49, 10, 7, 2} }

// Component has sine phase at trading ordinal zero, not the scanner endpoint.
type Component struct {
	Period         float64 `json:"period"`
	Amplitude      float64 `json:"amplitude"`
	Phase          float64 `json:"phase"`
	Strength       float64 `json:"strength"`
	GenuinePercent float64 `json:"genuine_percent"`
}

// Model is the version-one portable frozen annual model.
type Model struct {
	Schema                 string         `json:"schema"`
	Version                int            `json:"version"`
	ResearchReconstruction bool           `json:"research_reconstruction"`
	Year                   int            `json:"year"`
	SampleCount            int            `json:"sample_count"`
	Origin                 string         `json:"origin"`
	PhaseConvention        string         `json:"phase_convention"`
	Components             []Component    `json:"components"`
	Correlation            float64        `json:"correlation"`
	SubsetsChecked         int            `json:"subsets_checked"`
	CandidateCount         int            `json:"candidate_count"`
	Config                 map[string]any `json:"config"`
	Dates                  []string       `json:"dates,omitempty"`
	StartDate              string         `json:"start_date,omitempty"`
	EndDate                string         `json:"end_date,omitempty"`
	Coverage               string         `json:"coverage,omitempty"`
	FittedAsOf             string         `json:"fitted_as_of,omitempty"`
}

// Pearson computes signed centered population correlation.
func Pearson(left, right []float64) float64 {
	if len(left) < 1 || len(left) != len(right) {
		panic("correlation requires nonempty equal-length vectors")
	}
	a, b := harmoniccycle.Sum(left)/float64(len(left)), harmoniccycle.Sum(right)/float64(len(right))
	xx, yy, xy := make([]float64, len(left)), make([]float64, len(left)), make([]float64, len(left))
	for i := range left {
		x, y := left[i]-a, right[i]-b
		xx[i] = x * x
		yy[i] = y * y
		xy[i] = x * y
	}
	x, y := harmoniccycle.Sum(xx), harmoniccycle.Sum(yy)
	if x == 0 || y == 0 {
		return math.NaN()
	}
	return max(-1, min(1, harmoniccycle.Sum(xy)/math.Sqrt(x*y)))
}

// ChooseSubset exhaustively checks a bounded pool with signed Pearson and deterministic ties.
func ChooseSubset(target []float64, waves [][]float64, periods []float64, maxComponents int) ([]int, float64, int, error) {
	if len(waves) < 1 || len(waves) > 16 || len(waves) != len(periods) || len(target) < 6 || maxComponents < 1 || maxComponents > 7 {
		return nil, 0, 0, fmt.Errorf("invalid annual subset inputs")
	}
	n := len(target)
	mean := harmoniccycle.Sum(target) / float64(n)
	tc := make([]float64, n)
	energyValues := make([]float64, n)
	for i, value := range target {
		tc[i] = value - mean
		energyValues[i] = tc[i] * tc[i]
	}
	energy := harmoniccycle.Sum(energyValues)
	if energy == 0 {
		return nil, 0, 0, fmt.Errorf("constant target has no correlation")
	}
	centered := make([][]float64, len(waves))
	for i, wave := range waves {
		if len(wave) != n {
			return nil, 0, 0, fmt.Errorf("annual waves must share target length")
		}
		mean := harmoniccycle.Sum(wave) / float64(n)
		centered[i] = make([]float64, n)
		for j, value := range wave {
			centered[i][j] = value - mean
		}
	}
	cov := make([]float64, len(waves))
	gram := make([][]float64, len(waves))
	for i, left := range centered {
		products := make([]float64, n)
		for k := range n {
			products[k] = tc[k] * left[k]
		}
		cov[i] = harmoniccycle.Sum(products)
		gram[i] = make([]float64, len(waves))
		for j, right := range centered {
			for k := range n {
				products[k] = left[k] * right[k]
			}
			gram[i][j] = harmoniccycle.Sum(products)
		}
	}
	best := []int{}
	bestScore := math.Inf(-1)
	bestPeriods := []float64{}
	checked := 0
	for count := 1; count <= min(maxComponents, len(waves)); count++ {
		var visit func(int, []int)
		visit = func(start int, subset []int) {
			if len(subset) < count {
				for i := start; i <= len(waves)-(count-len(subset)); i++ {
					visit(i+1, append(append([]int{}, subset...), i))
				}
				return
			}
			checked++
			numerators := make([]float64, count)
			denominators := make([]float64, 0, count*count)
			key := make([]float64, count)
			for k, i := range subset {
				numerators[k] = cov[i]
				key[k] = periods[i]
				for _, j := range subset {
					denominators = append(denominators, gram[i][j])
				}
			}
			denominator := harmoniccycle.Sum(denominators)
			if denominator <= 0 {
				return
			}
			score := max(-1, min(1, harmoniccycle.Sum(numerators)/math.Sqrt(energy*denominator)))
			sort.Float64s(key)
			smaller := len(best) > count
			if len(best) == count {
				for i := range key {
					if key[i] != bestPeriods[i] {
						smaller = key[i] < bestPeriods[i]
						break
					}
				}
			}
			if score > bestScore+1e-12 || math.Abs(score-bestScore) <= 1e-12 && smaller {
				best = append([]int{}, subset...)
				bestScore = score
				bestPeriods = key
			}
		}
		visit(0, nil)
	}
	if len(best) < 1 || bestScore <= 0 {
		return nil, 0, checked, fmt.Errorf("no positive-correlation annual subset")
	}
	return best, bestScore, checked, nil
}
func modulo(value float64) float64 {
	value = math.Mod(value, 2*math.Pi)
	if value < 0 {
		value += 2 * math.Pi
	}
	return value
}

// FitAnnualValues fits a numerical completed year; DecodeAnnual additionally enforces dates and coverage.
func FitAnnualValues(prices []float64, year int, p *Params) (Model, error) {
	if p == nil {
		p = DefaultParams()
	}
	if year < 1 || p.CandidatePool < 1 || p.CandidatePool > 16 || p.MaxComponents < 1 || p.MaxComponents > 7 {
		return Model{}, fmt.Errorf("invalid annual bounds")
	}
	config := cyclescanner.ScanConfig{Window: len(prices), MinPeriod: p.MinPeriod, MaxPeriod: p.MaxPeriod, Smoothing: p.Smoothing, GenuineThreshold: p.GenuineThreshold, PhaseCycles: 0, TopK: p.CandidatePool, PeakDistance: p.PeakDistance, MaxSegments: 5}
	if err := config.Validate(); err != nil {
		return Model{}, err
	}
	scan := cyclescanner.ScanCycles(prices, config)
	cycles := make([]cyclescanner.RankedCycle, 0, p.CandidatePool)
	for _, cycle := range scan.Cycles {
		if !math.IsNaN(cycle.Period) {
			cycles = append(cycles, cycle)
		}
	}
	if len(cycles) < 1 {
		return Model{}, fmt.Errorf("no accepted annual cycles")
	}
	n := len(prices)
	waves := make([][]float64, len(cycles))
	periods := make([]float64, len(cycles))
	for i, cycle := range cycles {
		periods[i] = cycle.Period
		waves[i] = make([]float64, n)
		for j := range n {
			waves[i][j] = cyclescanner.ModeledValue(cycle, float64(j-(n-1)))
		}
	}
	subset, correlation, checked, err := ChooseSubset(hodrickprescottsmoothinganddetrending.CycleData(prices, p.Smoothing), waves, periods, p.MaxComponents)
	if err != nil {
		return Model{}, err
	}
	components := make([]Component, len(subset))
	for i, index := range subset {
		cycle := cycles[index]
		components[i] = Component{cycle.Period, cycle.FitAmplitude, modulo(cycle.Phase - 2*math.Pi*float64(n-1)/cycle.Period), cycle.Strength, cycle.GenuinePercent}
	}
	sort.Slice(components, func(i, j int) bool { return components[i].Period < components[j].Period })
	return Model{Schema: "lvt-annual-cycle-model", Version: 1, ResearchReconstruction: true, Year: year, SampleCount: n, Origin: "first-observed-trading-day, zero-based ordinal", PhaseConvention: "sine radians at ordinal zero", Components: components, Correlation: correlation, SubsetsChecked: checked, CandidateCount: len(cycles), Config: map[string]any{"window": n, "min_period": p.MinPeriod, "max_period": p.MaxPeriod, "smoothing": p.Smoothing, "genuine_threshold": p.GenuineThreshold, "phase_cycles": 0, "top_k": p.CandidatePool, "peak_distance": p.PeakDistance, "max_segments": 5, "candidate_pool": p.CandidatePool, "max_components": p.MaxComponents}}, nil
}

// AnnualWave evaluates the frozen model at any trading ordinal, including beyond the source year.
func AnnualWave(model Model, index float64) float64 {
	values := make([]float64, len(model.Components))
	for i, c := range model.Components {
		values[i] = c.Amplitude * math.Sin(2*math.Pi*index/c.Period+c.Phase)
	}
	return harmoniccycle.Sum(values)
}

// Validate checks portable schema, phases, components and any declared dates.
func (m Model) Validate() error {
	if m.Schema != "lvt-annual-cycle-model" || m.Version != 1 || m.Year < 1 || m.SampleCount < 6 || m.PhaseConvention != "sine radians at ordinal zero" || len(m.Components) < 1 || len(m.Components) > 7 {
		return fmt.Errorf("unsupported annual model schema or fields")
	}
	seen := map[float64]bool{}
	for _, c := range m.Components {
		if err := harmoniccycle.Real(c.Period, "period", 2.0000000000000004); err != nil {
			return err
		}
		if err := harmoniccycle.Real(c.Amplitude, "amplitude", 0); err != nil {
			return err
		}
		if math.IsNaN(c.Phase) || math.IsInf(c.Phase, 0) || seen[c.Period] {
			return fmt.Errorf("invalid or duplicate annual component")
		}
		seen[c.Period] = true
	}
	if m.Dates != nil {
		if len(m.Dates) != m.SampleCount {
			return fmt.Errorf("model dates must match sample_count")
		}
		var previous time.Time
		for _, text := range m.Dates {
			day, err := time.Parse("2006-01-02", text)
			if err != nil || day.Year() != m.Year || !day.After(previous) {
				return fmt.Errorf("invalid annual model dates")
			}
			previous = day
		}
	}
	if m.FittedAsOf != "" {
		day, err := time.Parse("2006-01-02", m.FittedAsOf)
		if err != nil || day.Before(time.Date(m.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)) {
			return fmt.Errorf("annual fit predates year completion")
		}
	}
	return nil
}

// Record is a completed trading observation, distinct from Metonic calendar-day input.
type Record struct {
	Day   time.Time
	Close float64
}

// DecodeAnnual checks past-year completion and exact supplied calendar or an explicit completeness assertion.
func DecodeAnnual(records []Record, year int, asOf time.Time, calendar []time.Time, confirmComplete bool, p *Params) (Model, error) {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	if year < 1 || asOf.Before(time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return Model{}, fmt.Errorf("annual decoding requires a completed year")
	}
	dates := []time.Time{}
	prices := []float64{}
	var previous time.Time
	for _, record := range records {
		if !record.Day.After(previous) {
			return Model{}, fmt.Errorf("dates must increase strictly")
		}
		previous = record.Day
		if record.Day.Year() == year {
			dates = append(dates, record.Day)
			prices = append(prices, record.Close)
		}
	}
	if len(dates) < 6 || dates[0].Month() != 1 || dates[len(dates)-1].Month() != 12 {
		return Model{}, fmt.Errorf("annual calendar must span January through December")
	}
	coverage := "caller asserted complete year; session completeness unverified"
	if calendar != nil {
		if len(calendar) != len(dates) {
			return Model{}, fmt.Errorf("observed dates must match supplied calendar")
		}
		for i, day := range dates {
			if !day.Equal(calendar[i]) {
				return Model{}, fmt.Errorf("observed dates must match supplied calendar")
			}
		}
		coverage = "exact match to supplied full-year calendar"
	} else if !confirmComplete {
		return Model{}, fmt.Errorf("provide calendar or confirm complete year")
	}
	model, err := FitAnnualValues(prices, year, p)
	if err != nil {
		return model, err
	}
	model.Dates = make([]string, len(dates))
	for i, day := range dates {
		model.Dates[i] = day.Format("2006-01-02")
	}
	model.StartDate = model.Dates[0]
	model.EndDate = model.Dates[len(dates)-1]
	model.Coverage = coverage
	model.FittedAsOf = asOf.Format("2006-01-02")
	return model, nil
}

// AnnualWaveReader plays a validated frozen model without fitting any prices.
type AnnualWaveReader struct {
	model Model
	index int
}

// NewAnnualWaveReader validates a frozen model.
func NewAnnualWaveReader(model Model) (*AnnualWaveReader, error) {
	if err := model.Validate(); err != nil {
		return nil, err
	}
	return &AnnualWaveReader{model: model}, nil
}

// Reset returns playback to ordinal zero.
func (s *AnnualWaveReader) Reset() { s.index = 0 }

// Update emits the next modeled value; annual models can extrapolate.
func (s *AnnualWaveReader) Update() float64 {
	value := AnnualWave(s.model, float64(s.index))
	s.index++
	return value
}
