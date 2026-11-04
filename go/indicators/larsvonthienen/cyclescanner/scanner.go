// Numerical source: Lars von Thienen / WhenToTrade (2017), CC BY 4.0.
// Research adaptations are documented in description.md.
package cyclescanner

import (
	"fmt"
	"math"
	"sort"

	"zpano/indicators/larsvonthienen/bartelsperiodicityscore"
	"zpano/indicators/larsvonthienen/harmoniccycle"
	"zpano/indicators/larsvonthienen/hodrickprescottsmoothinganddetrending"
)

// ScanConfig records the explicit reconstructed detection rules.
type ScanConfig struct {
	Window, MinPeriod, MaxPeriod             int
	Smoothing, GenuineThreshold, PhaseCycles float64
	TopK, PeakDistance, MaxSegments          int
}

// DefaultConfig returns the documented scanner defaults.
func DefaultConfig() ScanConfig { return ScanConfig{200, 5, 80, 1600, 49, 2, 3, 2, 5} }

// Validate checks the scanner's parameter contract.
func (s ScanConfig) Validate() error {
	if s.Window < 6 || s.MinPeriod < 3 || s.MaxPeriod < s.MinPeriod || s.TopK < 1 || s.PeakDistance < 1 || s.MaxSegments < 2 {
		return fmt.Errorf("invalid scanner integer parameters")
	}
	for name, value := range map[string]float64{"smoothing": s.Smoothing, "genuine_threshold": s.GenuineThreshold, "phase_cycles": s.PhaseCycles} {
		if err := harmoniccycle.Real(value, name, 0); err != nil {
			return err
		}
	}
	if s.GenuineThreshold > 100 || s.PhaseCycles != 0 && s.PhaseCycles < 1 {
		return fmt.Errorf("invalid scanner threshold or phase cycles")
	}
	return nil
}

// RankedCycle separates full-window detection amplitude from recent-fit amplitude.
// Phase is sine phase at the analysis endpoint.
type RankedCycle struct{ Period, Amplitude, Phase, Strength, GenuinePercent, Segments, FitAmplitude, FitSamples, Value, Slope, EtaTop, EtaBottom float64 }

// ScanOutput owns a complete spectral/evidence snapshot and padded ranked cycles.
type ScanOutput struct {
	Amplitude, GenuinePercent []float64
	Cycles                    []RankedCycle
	AcceptedCount             float64
}

// EmptyCycle returns unavailable rank fields.
func EmptyCycle() RankedCycle {
	nan := math.NaN()
	return RankedCycle{nan, nan, nan, nan, nan, nan, nan, nan, nan, nan, nan, nan}
}

// EmptyScan creates a fresh unavailable snapshot.
func EmptyScan(config ScanConfig) ScanOutput {
	count := config.MaxPeriod - config.MinPeriod + 1
	output := ScanOutput{Amplitude: make([]float64, count), GenuinePercent: make([]float64, count), Cycles: make([]RankedCycle, config.TopK), AcceptedCount: math.NaN()}
	for i := range count {
		output.Amplitude[i] = math.NaN()
		output.GenuinePercent[i] = math.NaN()
	}
	for i := range output.Cycles {
		output.Cycles[i] = EmptyCycle()
	}
	return output
}

// ModeledValue evaluates a recent endpoint-phase fit at an offset in bars.
func ModeledValue(cycle RankedCycle, offset float64) float64 {
	return cycle.FitAmplitude * math.Sin(cycle.Phase+2*math.Pi*offset/cycle.Period)
}
func modulo(value float64) float64 {
	value = math.Mod(value, 2*math.Pi)
	if value < 0 {
		value += 2 * math.Pi
	}
	return value
}

// ScanCycles runs batch HP, GDFT, latest-raw-segment validation and strength ranking.
func ScanCycles(prices []float64, config ScanConfig) ScanOutput {
	if err := config.Validate(); err != nil {
		panic(err)
	}
	if len(prices) < 6 {
		panic("scan requires >= 6 prices")
	}
	for _, value := range prices {
		harmoniccycle.Sample(value)
		if math.IsNaN(value) {
			panic("scan batch data must be finite")
		}
	}
	residual := hodrickprescottsmoothinganddetrending.CycleData(prices, config.Smoothing)
	output := EmptyScan(config)
	count := len(output.Amplitude)
	models := make([]harmoniccycle.CycleModel, count)
	evidence := make([]bartelsperiodicityscore.BartelsOutput, count)
	for i := range count {
		period := config.MinPeriod + i
		models[i] = harmoniccycle.Goertzel(residual, float64(period))
		segments := min(len(prices)/period, config.MaxSegments)
		evidence[i] = bartelsperiodicityscore.BartelsScore(prices[len(prices)-segments*period:], period, config.MaxSegments, 2)
		output.Amplitude[i] = models[i].Amplitude
		output.GenuinePercent[i] = evidence[i].GenuinePercent
	}
	candidates := make([]int, 0, count)
	for i, model := range models {
		left, right := math.Inf(-1), math.Inf(-1)
		if i > 0 {
			left = output.Amplitude[i-1]
		}
		if i+1 < count {
			right = output.Amplitude[i+1]
		}
		score := evidence[i]
		if model.Amplitude > 0 && model.Amplitude > left && model.Amplitude >= right && !math.IsNaN(score.GenuinePercent) && score.GenuinePercent > config.GenuineThreshold {
			candidates = append(candidates, i)
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		left, right := models[candidates[i]], models[candidates[j]]
		a, b := left.Amplitude/left.Period, right.Amplitude/right.Period
		if a == b {
			return left.Period < right.Period
		}
		return a > b
	})
	accepted := make([]RankedCycle, 0, len(candidates))
	for _, i := range candidates {
		model, score := models[i], evidence[i]
		suppressed := false
		for _, prior := range accepted {
			if math.Abs(model.Period-prior.Period) < float64(config.PeakDistance) {
				suppressed = true
				break
			}
		}
		if suppressed {
			continue
		}
		fitN := len(prices)
		if config.PhaseCycles != 0 {
			fitN = min(fitN, max(6, int(math.Ceil(config.PhaseCycles*model.Period))))
		}
		fit := model
		if fitN != len(prices) {
			fit = harmoniccycle.Goertzel(hodrickprescottsmoothinganddetrending.CycleData(prices[len(prices)-fitN:], config.Smoothing), model.Period)
		}
		phase := modulo(fit.Phase + 2*math.Pi*float64(fitN-1)/model.Period)
		omega := 2 * math.Pi / model.Period
		value := fit.Amplitude * math.Sin(phase)
		slope := fit.Amplitude * omega * math.Cos(phase)
		etaTop := modulo(math.Pi/2-phase) / omega
		etaBottom := modulo(3*math.Pi/2-phase) / omega
		accepted = append(accepted, RankedCycle{model.Period, model.Amplitude, phase, model.Amplitude / model.Period, score.GenuinePercent, score.Segments, fit.Amplitude, float64(fitN), value, slope, etaTop, etaBottom})
	}
	copy(output.Cycles, accepted)
	output.AcceptedCount = float64(len(accepted))
	return output
}

// RollingScanner retains a bounded oldest-first window. Wrappers provide synchronization.
type RollingScanner struct {
	Config      ScanConfig
	Prices      []float64
	SamplesSeen int
}

// NewRollingScanner validates and creates an internal scanner.
func NewRollingScanner(config ScanConfig) (*RollingScanner, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &RollingScanner{Config: config, Prices: make([]float64, 0, config.Window)}, nil
}

// Reset discards history while preserving the configuration.
func (s *RollingScanner) Reset() { s.Prices = s.Prices[:0]; s.SamplesSeen = 0 }

// IsPrimed reports full-window readiness, independent of acceptance.
func (s *RollingScanner) IsPrimed() bool { return len(s.Prices) == s.Config.Window }

// Update calculates before committing the next finite sample.
func (s *RollingScanner) Update(price float64) ScanOutput {
	harmoniccycle.Sample(price)
	if math.IsNaN(price) {
		return EmptyScan(s.Config)
	}
	prices := harmoniccycle.Push(append(make([]float64, 0, s.Config.Window), s.Prices...), price, s.Config.Window)
	output := EmptyScan(s.Config)
	if len(prices) == s.Config.Window {
		output = ScanCycles(prices, s.Config)
	}
	s.Prices = prices
	s.SamplesSeen++
	return output
}
