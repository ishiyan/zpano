// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package harmonicenergysuperpositionwave

import (
	"fmt"
	"io"
	"math"
	"time"

	"zpano/indicators/larsvonthienen/annualcyclemodel"
	"zpano/indicators/larsvonthienen/harmoniccycle"
)

// MetonicLags fixes the six historical source years.
var MetonicLags = [6]int{8, 11, 19, 27, 30, 38}

// Scale records source-year-only normalization.
type Scale struct {
	Year        int     `json:"year"`
	Mean        float64 `json:"mean"`
	Scale       float64 `json:"scale"`
	SampleCount int     `json:"sample_count"`
}

// Model is the portable version-one superposition model.
type Model struct {
	Schema                 string                   `json:"schema"`
	Version                int                      `json:"version"`
	ResearchReconstruction bool                     `json:"research_reconstruction"`
	TargetYear             int                      `json:"target_year"`
	SourceYears            []int                    `json:"source_years"`
	Lags                   []int                    `json:"lags"`
	Normalization          string                   `json:"normalization"`
	CalendarPolicy         string                   `json:"calendar_policy"`
	Invert                 bool                     `json:"invert"`
	SampleCount            int                      `json:"sample_count"`
	Scales                 []Scale                  `json:"scales"`
	AnnualModels           []annualcyclemodel.Model `json:"annual_models"`
	Dates                  []string                 `json:"dates,omitempty"`
}

// SuperpositionValues aligns six source waves at ordinal zero, never time-stretching or reading target prices.
func SuperpositionValues(models []annualcyclemodel.Model, targetYear, count int, normalization, calendarPolicy string, invert bool) ([]float64, Model, error) {
	if targetYear < 39 || count < 1 || normalization != "none" && normalization != "rms" || calendarPolicy != "strict" && calendarPolicy != "extend" {
		return nil, Model{}, fmt.Errorf("invalid superposition configuration")
	}
	byYear := map[int]annualcyclemodel.Model{}
	for _, model := range models {
		if err := model.Validate(); err != nil {
			return nil, Model{}, err
		}
		if _, ok := byYear[model.Year]; ok {
			return nil, Model{}, fmt.Errorf("duplicate source year")
		}
		byYear[model.Year] = model
	}
	if len(byYear) != 6 {
		return nil, Model{}, fmt.Errorf("require six historical models")
	}
	model := Model{Schema: "lvt-superposition-model", Version: 1, ResearchReconstruction: true, TargetYear: targetYear, Lags: append([]int{}, MetonicLags[:]...), Normalization: normalization, CalendarPolicy: calendarPolicy, Invert: invert, SampleCount: count}
	waves := make([][]float64, 6)
	for i, lag := range MetonicLags {
		year := targetYear - lag
		annual, ok := byYear[year]
		if !ok {
			return nil, Model{}, fmt.Errorf("require target years minus 8,11,19,27,30,38")
		}
		n := annual.SampleCount
		if calendarPolicy == "strict" && count != n {
			return nil, Model{}, fmt.Errorf("strict ordinal alignment requires equal annual sample counts")
		}
		base := make([]float64, n)
		for j := range n {
			base[j] = annualcyclemodel.AnnualWave(annual, float64(j))
		}
		mean, scale := 0.0, 1.0
		if normalization == "rms" {
			mean = harmoniccycle.Sum(base) / float64(n)
			energy := make([]float64, n)
			for j, value := range base {
				energy[j] = (value - mean) * (value - mean)
			}
			scale = math.Sqrt(harmoniccycle.Sum(energy) / float64(n))
		}
		if scale == 0 {
			return nil, Model{}, fmt.Errorf("zero-energy annual wave cannot be normalized")
		}
		waves[i] = make([]float64, count)
		for j := range count {
			waves[i][j] = (annualcyclemodel.AnnualWave(annual, float64(j)) - mean) / scale
		}
		model.SourceYears = append(model.SourceYears, year)
		model.Scales = append(model.Scales, Scale{year, mean, scale, n})
		model.AnnualModels = append(model.AnnualModels, annual)
	}
	sign := 1.0
	if invert {
		sign = -1
	}
	values := make([]float64, count)
	terms := make([]float64, 6)
	for i := range count {
		for j := range 6 {
			terms[j] = waves[j][i]
		}
		values[i] = sign * harmoniccycle.Sum(terms) / 6
	}
	return values, model, nil
}

// SuperpositionWaveReader ends at the declared target count.
type SuperpositionWaveReader struct {
	values []float64
	index  int
}

// NewSuperpositionWaveReader validates models and prepares bounded playback.
func NewSuperpositionWaveReader(models []annualcyclemodel.Model, targetYear, count int, normalization, calendarPolicy string, invert bool) (*SuperpositionWaveReader, error) {
	values, _, err := SuperpositionValues(models, targetYear, count, normalization, calendarPolicy, invert)
	if err != nil {
		return nil, err
	}
	return &SuperpositionWaveReader{values: values}, nil
}

// Reset returns playback to ordinal zero.
func (s *SuperpositionWaveReader) Reset() { s.index = 0 }

// Update returns io.EOF beyond the target calendar.
func (s *SuperpositionWaveReader) Update() (float64, error) {
	if s.index >= len(s.values) {
		return 0, io.EOF
	}
	value := s.values[s.index]
	s.index++
	return value, nil
}

// Validate rejects unsupported schemas, invalid source models and inconsistent target calendars.
func (m Model) Validate() error {
	if m.Schema != "lvt-superposition-model" || m.Version != 1 {
		return fmt.Errorf("unsupported superposition model")
	}
	_, rebuilt, err := SuperpositionValues(m.AnnualModels, m.TargetYear, m.SampleCount, m.Normalization, m.CalendarPolicy, m.Invert)
	if err != nil {
		return err
	}
	if len(m.SourceYears) != 6 || len(m.Lags) != 6 {
		return fmt.Errorf("invalid source years or lags")
	}
	for i := range 6 {
		if m.SourceYears[i] != rebuilt.SourceYears[i] || m.Lags[i] != MetonicLags[i] {
			return fmt.Errorf("inconsistent source years or lags")
		}
	}
	if m.Dates != nil {
		if len(m.Dates) != m.SampleCount {
			return fmt.Errorf("dates must match sample_count")
		}
		previous := ""
		for _, text := range m.Dates {
			day, err := time.Parse("2006-01-02", text)
			if err != nil || day.Year() != m.TargetYear || text <= previous {
				return fmt.Errorf("invalid target dates")
			}
			previous = text
		}
	}
	return nil
}
