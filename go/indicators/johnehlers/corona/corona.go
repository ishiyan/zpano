package corona

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sync"
	"time"

	"zpano/indicators/core/outputs"
)

const (
	// FirstBankSample is the 1-based accepted sample count at which bandpass channels begin updating.
	FirstBankSample = 13

	// TwoPi is 2 * pi at full binary64 precision.
	TwoPi float64 = 2.0 * math.Pi
)

// SqrtThreeOverTwo is sqrt(3) / 2 used for the 60-degree leading signal.
var SqrtThreeOverTwo = math.Sqrt(3.0) / 2.0

var firWeights = [6]float64{1.0, 2.0, 3.0, 3.0, 2.0, 1.0}

// ValidateInteger checks an integer parameter against its minimum bound.
func ValidateInteger(name string, value, minimum int) error {
	if value < minimum {
		return fmt.Errorf("%s must be >= %d", name, minimum)
	}
	return nil
}

// ValidateFinite checks that a float64 parameter is finite.
func ValidateFinite(name string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("%s must be finite", name)
	}
	return nil
}

// ValidPrice returns (false, nil) for NaN, (false, err) for +/-Inf, and (true, nil) for finite prices.
func ValidPrice(price float64) (bool, error) {
	if math.IsNaN(price) {
		return false, nil
	}
	if math.IsInf(price, 0) {
		return false, errors.New("price must be finite or NaN")
	}
	return true, nil
}

func checked(value float64) (float64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, errors.New("calculation overflow; use finite, reasonably scaled prices")
	}
	return value, nil
}

// Clip clamps value into [lower, upper].
func Clip(value, lower, upper float64) float64 {
	return max(lower, min(upper, value))
}

// FormatNumber formats a float64 cleanly for mnemonics without "-0".
func FormatNumber(value float64) string {
	if value == 0 {
		value = 0
	}
	return fmt.Sprintf("%g", value)
}

// RollingAppend appends value to history (retaining at most length items) and returns (history, min, max).
func RollingAppend(history []float64, value float64, length int) ([]float64, float64, float64) {
	if len(history) < length {
		history = append(history, value)
	} else {
		copy(history, history[1:])
		history[len(history)-1] = value
	}
	minimum := history[0]
	maximum := history[0]
	for _, item := range history[1:] {
		if item < minimum {
			minimum = item
		}
		if item > maximum {
			maximum = item
		}
	}
	return history, minimum, maximum
}

// Filter holds the state of a single half-period bandpass channel.
type Filter struct {
	InPhase1    float64
	InPhase2    float64
	Quadrature1 float64
	Quadrature2 float64
	Real1       float64
	Real2       float64
	Imaginary1  float64
	Imaginary2  float64
	Power       float64
	Decibels    float64
}

// CoronaSnapshot owns all 18 diagnostic fields for one Corona engine update.
type CoronaSnapshot struct {
	Parameters          Params
	SampleCount         int
	Accepted            bool
	Primed              bool
	Price               float64
	Time                time.Time
	HighPass            float64
	Filtered            float64
	Bandwidth           float64
	DominantCycle       float64
	DominantCycleMedian float64
	MaximumPower        float64
	NormalizationPower  float64
	SelectedBins        int
	HasCycleEstimate    bool
	Power               []float64
	RelativePower       []float64
	Decibels            []float64
}

// Snapshot aliases [CoronaSnapshot].
type Snapshot = CoronaSnapshot

func cloneSnapshot(snapshot CoronaSnapshot) CoronaSnapshot {
	snapshot.Power = slices.Clone(snapshot.Power)
	snapshot.RelativePower = slices.Clone(snapshot.RelativePower)
	snapshot.Decibels = slices.Clone(snapshot.Decibels)
	return snapshot
}

// Corona implements the shared Ehlers Corona channelized receiver.
type Corona struct {
	mu               sync.RWMutex
	parameters       Params
	periods          []float64
	betas            []float64
	pole             float64
	sampleCount      int
	pricePrevious    float64
	hpHistory        [6]float64
	filteredPrevious float64
	bank             []Filter
	dc               float64
	dcHistory        [5]float64
	hasLastSnapshot  bool
	lastSnapshot     CoronaSnapshot
}

// NewCorona creates and validates a Corona engine. Passing nil uses [DefaultParams].
func NewCorona(params *Params) (*Corona, error) {
	source := DefaultParams()
	if params != nil {
		copied := *params
		source = &copied
	}
	if err := ValidateInteger("HighPassFilterCutoff", source.HighPassFilterCutoff, 5); err != nil {
		return nil, err
	}
	if err := ValidateInteger("MinimalPeriod", source.MinimalPeriod, 6); err != nil {
		return nil, err
	}
	if err := ValidateInteger("MaximalPeriod", source.MaximalPeriod, source.MinimalPeriod+1); err != nil {
		return nil, err
	}
	if err := ValidateFinite("DecibelsLowerThreshold", source.DecibelsLowerThreshold); err != nil {
		return nil, err
	}
	if err := ValidateFinite("DecibelsUpperThreshold", source.DecibelsUpperThreshold); err != nil {
		return nil, err
	}
	if !(0.0 <= source.DecibelsLowerThreshold && source.DecibelsLowerThreshold < source.DecibelsUpperThreshold && source.DecibelsUpperThreshold > 0.0) {
		return nil, errors.New("require 0 <= DecibelsLowerThreshold < DecibelsUpperThreshold")
	}

	angle := TwoPi / float64(source.HighPassFilterCutoff)
	pole := (1.0 - math.Sin(angle)) / math.Cos(angle)
	if math.IsNaN(pole) || math.IsInf(pole, 0) || !(0.0 < pole && pole < 1.0) {
		return nil, errors.New("high-pass coefficient must have a stable positive pole")
	}

	bankLength := 2*(source.MaximalPeriod-source.MinimalPeriod) + 1
	periods := make([]float64, bankLength)
	betas := make([]float64, bankLength)
	for i := range bankLength {
		period := float64(2*source.MinimalPeriod+i) / 2.0
		periods[i] = period
		betas[i] = math.Cos(TwoPi / period)
	}

	engine := &Corona{
		parameters: *source,
		periods:    periods,
		betas:      betas,
		pole:       pole,
	}
	engine.resetLocked()
	return engine, nil
}

func (c *Corona) resetLocked() {
	c.sampleCount = 0
	c.pricePrevious = 0
	c.hpHistory = [6]float64{}
	c.filteredPrevious = 0
	c.bank = make([]Filter, len(c.periods))
	c.dc = 0
	c.dcHistory = [5]float64{}
	c.hasLastSnapshot = false
	c.lastSnapshot = CoronaSnapshot{}
}

// Reset discards streaming history while retaining validated parameters.
func (c *Corona) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.resetLocked()
}

// IsPrimed reports whether at least [FirstBankSample] accepted samples have been processed.
func (c *Corona) IsPrimed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sampleCount >= FirstBankSample
}

// SampleCount returns the number of accepted samples processed so far.
func (c *Corona) SampleCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sampleCount
}

// Parameters returns the validated engine configuration.
func (c *Corona) Parameters() Params {
	return c.parameters
}

// FilterBankLength returns the number of half-period bandpass channels.
func (c *Corona) FilterBankLength() int {
	return len(c.periods)
}

// MinimalPeriodTimesTwo returns 2 * MinimalPeriod.
func (c *Corona) MinimalPeriodTimesTwo() int {
	return 2 * c.parameters.MinimalPeriod
}

// MaximalPeriodTimesTwo returns 2 * MaximalPeriod.
func (c *Corona) MaximalPeriodTimesTwo() int {
	return 2 * c.parameters.MaximalPeriod
}

// LastSnapshot returns a copy of the most recent accepted snapshot if one exists.
func (c *Corona) LastSnapshot() (CoronaSnapshot, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.hasLastSnapshot {
		return CoronaSnapshot{}, false
	}
	return cloneSnapshot(c.lastSnapshot), true
}

// TryUpdate advances the engine by one observation and returns an immutable snapshot or error.
func (c *Corona) TryUpdate(price float64, t time.Time) (CoronaSnapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ok, err := ValidPrice(price)
	if err != nil {
		return CoronaSnapshot{}, err
	}
	if !ok {
		if c.hasLastSnapshot {
			skipped := cloneSnapshot(c.lastSnapshot)
			skipped.Accepted = false
			skipped.Price = price
			skipped.Time = t
			return skipped, nil
		}
		bankLength := len(c.periods)
		return CoronaSnapshot{
			Parameters:          c.parameters,
			SampleCount:         0,
			Accepted:            false,
			Primed:              false,
			Price:               price,
			Time:                t,
			HighPass:            0,
			Filtered:            0,
			Bandwidth:           0.5,
			DominantCycle:       0,
			DominantCycleMedian: float64(c.parameters.MinimalPeriod),
			MaximumPower:        0,
			NormalizationPower:  0.01,
			SelectedBins:        0,
			HasCycleEstimate:    false,
			Power:               make([]float64, bankLength),
			RelativePower:       make([]float64, bankLength),
			Decibels:            make([]float64, bankLength),
		}, nil
	}

	count := c.sampleCount + 1
	previous := c.pricePrevious
	if count == 1 {
		previous = price
	}
	hp, err := checked(0.5*(1.0+c.pole)*(price-previous) + c.pole*c.hpHistory[5])
	if err != nil {
		return CoronaSnapshot{}, err
	}
	var hpHistory [6]float64
	copy(hpHistory[:5], c.hpHistory[1:])
	hpHistory[5] = hp

	firSum := 0.0
	for i := range 6 {
		firSum += hpHistory[i] * firWeights[i]
	}
	filtered, err := checked(firSum / 12.0)
	if err != nil {
		return CoronaSnapshot{}, err
	}
	if count < 7 {
		filtered, err = checked(price - previous)
		if err != nil {
			return CoronaSnapshot{}, err
		}
	}
	if count == 1 {
		filtered = 0.0
	}
	bandwidth := max(0.1, 0.5-0.015*float64(count))

	bank := make([]Filter, len(c.periods))
	for i, period := range c.periods {
		old := c.bank[i]
		if count < FirstBankSample {
			bank[i] = old
			continue
		}
		beta := c.betas[i]
		gamma := 1.0 / math.Cos(2.0*TwoPi*bandwidth/period)
		alpha := gamma - math.Sqrt(gamma*gamma-1.0)
		inPhase := filtered
		quadrature, qErr := checked((period / TwoPi) * (filtered - c.filteredPrevious))
		if qErr != nil {
			return CoronaSnapshot{}, qErr
		}
		realPart, rErr := checked(0.5*(1.0-alpha)*(inPhase-old.InPhase2) + beta*(1.0+alpha)*old.Real1 - alpha*old.Real2)
		if rErr != nil {
			return CoronaSnapshot{}, rErr
		}
		imaginaryPart, iErr := checked(0.5*(1.0-alpha)*(quadrature-old.Quadrature2) + beta*(1.0+alpha)*old.Imaginary1 - alpha*old.Imaginary2)
		if iErr != nil {
			return CoronaSnapshot{}, iErr
		}
		power, pErr := checked(realPart*realPart + imaginaryPart*imaginaryPart)
		if pErr != nil {
			return CoronaSnapshot{}, pErr
		}
		bank[i] = Filter{
			InPhase1:    inPhase,
			InPhase2:    old.InPhase1,
			Quadrature1: quadrature,
			Quadrature2: old.Quadrature1,
			Real1:       realPart,
			Real2:       old.Real1,
			Imaginary1:  imaginaryPart,
			Imaginary2:  old.Imaginary1,
			Power:       power,
			Decibels:    old.Decibels,
		}
	}

	maximum := 0.0
	for _, channel := range bank {
		if channel.Power > maximum {
			maximum = channel.Power
		}
	}
	firstPower := bank[0].Power
	if firstPower == 0.0 {
		firstPower = 0.01
	}
	normalizer := max(firstPower, maximum)

	numerator := 0.0
	denominator := 0.0
	selected := 0
	lowerThreshold := c.parameters.DecibelsLowerThreshold
	upperThreshold := c.parameters.DecibelsUpperThreshold
	powers := make([]float64, len(bank))
	relativePowers := make([]float64, len(bank))
	decibelsSlice := make([]float64, len(bank))

	for i, period := range c.periods {
		channel := bank[i]
		relative := channel.Power / normalizer
		raw := channel.Decibels
		if relative > 0.0 {
			raw = 10.0 * math.Log10((1.0-0.99*relative)/0.01)
		}
		decibels := min(upperThreshold, 0.33*raw+0.67*channel.Decibels)
		channel.Decibels = decibels
		bank[i] = channel
		powers[i] = channel.Power
		relativePowers[i] = relative
		decibelsSlice[i] = decibels

		if decibels <= lowerThreshold {
			weight := upperThreshold - decibels
			numerator += period * weight
			denominator += weight
			selected++
		}
	}

	dc := c.dc
	if denominator > 0.0 {
		dc = numerator / denominator
	}
	var dcHistory [5]float64
	copy(dcHistory[:4], c.dcHistory[1:])
	dcHistory[4] = dc

	sortedDC := dcHistory
	slices.Sort(sortedDC[:])
	median := max(float64(c.parameters.MinimalPeriod), sortedDC[2])

	snapshot := CoronaSnapshot{
		Parameters:          c.parameters,
		SampleCount:         count,
		Accepted:            true,
		Primed:              count >= FirstBankSample,
		Price:               price,
		Time:                t,
		HighPass:            hp,
		Filtered:            filtered,
		Bandwidth:           bandwidth,
		DominantCycle:       dc,
		DominantCycleMedian: median,
		MaximumPower:        maximum,
		NormalizationPower:  normalizer,
		SelectedBins:        selected,
		HasCycleEstimate:    maximum > 0.0 && selected > 0,
		Power:               powers,
		RelativePower:       relativePowers,
		Decibels:            decibelsSlice,
	}

	c.sampleCount = count
	c.pricePrevious = price
	c.hpHistory = hpHistory
	c.filteredPrevious = filtered
	c.bank = bank
	c.dc = dc
	c.dcHistory = dcHistory
	c.hasLastSnapshot = true
	c.lastSnapshot = cloneSnapshot(snapshot)
	return snapshot, nil
}

// Update advances the engine by one observation and panics on non-finite (+/-Inf) or overflow input.
func (c *Corona) Update(price float64, t time.Time) CoronaSnapshot {
	snapshot, err := c.TryUpdate(price, t)
	if err != nil {
		panic(err)
	}
	return snapshot
}

// AcceptSnapshot validates that snapshot matches expectedParams and advances monotonically.
func AcceptSnapshot(snapshot CoronaSnapshot, expectedParams Params, expectedNextCount int) (bool, error) {
	if snapshot.Parameters != expectedParams {
		return false, errors.New("snapshot must use this consumer's Corona parameters")
	}
	if !snapshot.Accepted {
		return false, nil
	}
	if snapshot.SampleCount != expectedNextCount {
		return false, errors.New("consume every accepted snapshot exactly once, in order from sample 1")
	}
	return true, nil
}

// AdaptiveCycle holds the output of [AdaptiveBandPass].
type AdaptiveCycle struct {
	BandPass   float64
	Quadrature float64
	Amplitude  float64
	Lead60     float64
}

// AdaptiveBandPass implements the two-pole bandpass filter tuned to the dominant cycle median.
type AdaptiveBandPass struct {
	prices     [2]float64
	priceCount int
	bp1        float64
	bp2        float64
}

// NewAdaptiveBandPass creates a zero-initialized [AdaptiveBandPass].
func NewAdaptiveBandPass() *AdaptiveBandPass {
	return &AdaptiveBandPass{}
}

// Reset clears the filter history.
func (a *AdaptiveBandPass) Reset() {
	a.prices = [2]float64{}
	a.priceCount = 0
	a.bp1 = 0
	a.bp2 = 0
}

// TryUpdate advances the adaptive bandpass filter by one sample.
func (a *AdaptiveBandPass) TryUpdate(price, period float64) (AdaptiveCycle, error) {
	if a.priceCount < 2 {
		a.prices[a.priceCount] = price
		a.priceCount++
		return AdaptiveCycle{}, nil
	}
	omega := TwoPi / period
	beta := math.Cos(omega)
	gamma := 1.0 / math.Cos(2.0*omega*0.1)
	alpha := gamma - math.Sqrt(gamma*gamma-1.0)
	bp, err := checked(0.5*(1.0-alpha)*(price-a.prices[0]) + beta*(1.0+alpha)*a.bp1 - alpha*a.bp2)
	if err != nil {
		return AdaptiveCycle{}, err
	}
	quadrature, err := checked((bp - a.bp1) / omega)
	if err != nil {
		return AdaptiveCycle{}, err
	}
	amplitude, err := checked(math.Sqrt(bp*bp + quadrature*quadrature))
	if err != nil {
		return AdaptiveCycle{}, err
	}
	lead60, err := checked(0.5*bp + SqrtThreeOverTwo*quadrature)
	if err != nil {
		return AdaptiveCycle{}, err
	}
	a.prices[0] = a.prices[1]
	a.prices[1] = price
	a.bp2 = a.bp1
	a.bp1 = bp
	return AdaptiveCycle{
		BandPass:   bp,
		Quadrature: quadrature,
		Amplitude:  amplitude,
		Lead60:     lead60,
	}, nil
}

// Update advances the adaptive bandpass filter and panics on overflow.
func (a *AdaptiveBandPass) Update(price, period float64) AdaptiveCycle {
	cycle, err := a.TryUpdate(price, period)
	if err != nil {
		panic(err)
	}
	return cycle
}

// RasterKind selects the blending recurrence for [Raster].
type RasterKind uint8

const (
	RasterSwing RasterKind = iota + 1
	RasterTrend
	RasterSNR
)

// Raster implements the Corona intensity raster for Swing, Trend, and SNR wrappers.
type Raster struct {
	length     int
	lower      float64
	upper      float64
	ceiling    float64
	step       float64
	first      float64
	resolution float64
	values     []float64
}

// NewRaster validates and constructs a [Raster].
func NewRaster(length int, lower, upper, ceiling float64) (*Raster, error) {
	if err := ValidateInteger("RasterLength", length, 2); err != nil {
		return nil, err
	}
	if err := ValidateFinite("MinParameterValue", lower); err != nil {
		return nil, err
	}
	if err := ValidateFinite("MaxParameterValue", upper); err != nil {
		return nil, err
	}
	if err := ValidateFinite("MaxRasterValue", ceiling); err != nil {
		return nil, err
	}
	if upper <= lower || ceiling <= 0.0 {
		return nil, errors.New("require MaxParameterValue > MinParameterValue and MaxRasterValue > 0")
	}
	step := (upper - lower) / float64(length)
	return &Raster{
		length:     length,
		lower:      lower,
		upper:      upper,
		ceiling:    ceiling,
		step:       step,
		first:      lower + step,
		resolution: 1.0 / step,
		values:     make([]float64, length),
	}, nil
}

// Reset zeroes all raster cells.
func (r *Raster) Reset() {
	clear(r.values)
}

// Length returns the number of raster cells.
func (r *Raster) Length() int { return r.length }

// Lower returns the lower scalar display bound.
func (r *Raster) Lower() float64 { return r.lower }

// Upper returns the upper scalar display bound.
func (r *Raster) Upper() float64 { return r.upper }

// Ceiling returns the maximum raster intensity ceiling.
func (r *Raster) Ceiling() float64 { return r.ceiling }

// Step returns the ordinate step per cell.
func (r *Raster) Step() float64 { return r.step }

// First returns the ordinate of the first sampled cell (lower + step).
func (r *Raster) First() float64 { return r.first }

// Resolution returns bins per unit (1 / step) matching [outputs.Heatmap].
func (r *Raster) Resolution() float64 { return r.resolution }

// ValuesCopy returns a defensive copy of the current raster cells.
func (r *Raster) ValuesCopy() []float64 {
	return slices.Clone(r.values)
}

// SetValues overwrites the raster cells (used in unit tests).
func (r *Raster) SetValues(values []float64) {
	copy(r.values, values)
}

// Empty returns an empty [outputs.Heatmap] with this raster's axis geometry.
func (r *Raster) Empty(t time.Time) *outputs.Heatmap {
	return outputs.NewEmptyHeatmap(t, r.first, r.upper, r.resolution)
}

// Column returns a populated [outputs.Heatmap] with a fresh copy of values.
func (r *Raster) Column(t time.Time) *outputs.Heatmap {
	copied := slices.Clone(r.values)
	minimum := copied[0]
	maximum := copied[0]
	for _, value := range copied[1:] {
		if value < minimum {
			minimum = value
		}
		if value > maximum {
			maximum = value
		}
	}
	return outputs.NewHeatmap(t, r.first, r.upper, r.resolution, minimum, maximum, copied)
}

// Update updates all raster cells in place using explicit half-up marker rounding.
func (r *Raster) Update(position, width float64, kind RasterKind, suppress bool) {
	marker := int(math.Floor(float64(r.length)*position + 0.5))
	var exponent, scale, historyScale float64
	switch kind {
	case RasterSwing:
		exponent, scale, historyScale = 0.95, 0.5, 0.5
	case RasterTrend:
		exponent, scale, historyScale = 0.85, 0.8, 0.2
	case RasterSNR:
		exponent, scale, historyScale = 0.8, 0.5, 1.0
	}
	for i, old := range r.values {
		j := i + 1
		var value float64
		switch {
		case suppress:
			value = r.ceiling
		case j == marker:
			value = 0.5 * old
		case width <= 0.0:
			value = r.ceiling
		default:
			distance := math.Abs(r.ceiling*position-r.ceiling*float64(j)/float64(r.length)) / width
			if kind == RasterSNR && j > marker && distance <= 1.0 {
				value = r.ceiling
			} else {
				value = scale * (math.Pow(distance, exponent) + historyScale*old)
			}
		}
		r.values[i] = Clip(value, 0.0, r.ceiling)
	}
}
