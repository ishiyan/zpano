package corona

//nolint: gofumpt
import (
	"fmt"
	"math"
	"sort"
)

const (
	defaultHighPassFilterCutoff   = 30
	defaultMinimalPeriod          = 6
	defaultMaximalPeriod          = 30
	defaultDecibelsLowerThreshold = 6.
	defaultDecibelsUpperThreshold = 20.

	highPassFilterBufferSize = 6
	// Firs coefficients {1, 2, 3, 3, 2, 1} / 12. Index 0 = oldest, index 5 = current.
	firCoefSum = 12.

	deltaLowerThreshold = 0.1
	deltaFactor         = -0.015
	deltaSummand        = 0.5

	dominantCycleBufferSize = 5
	// Median index: 5 sorted values → take index 2 (middle).
	dominantCycleMedianIndex = 2

	decibelsSmoothingAlpha    = 0.33
	decibelsSmoothingOneMinus = 0.67

	// 1 - 0.99 * normalized amplitude, clamped away from zero by 0.01.
	normalizedAmplitudeFactor = 0.99
	decibelsFloor             = 0.01
	decibelsGain              = 10.
)

// Filter holds the per-bin state of a single bandpass filter in the bank.
type Filter struct {
	InPhase            float64
	InPhasePrevious    float64
	Quadrature         float64
	QuadraturePrevious float64
	Real               float64
	RealPrevious       float64
	Imaginary          float64
	ImaginaryPrevious  float64

	// AmplitudeSquared is the |Real + j·Imaginary|² of the most recent update.
	AmplitudeSquared float64

	// Decibels is the smoothed dB value of the most recent update.
	Decibels float64
}

// Corona is the shared spectral-analysis engine. It is not an indicator on its
// own; it is consumed by the CoronaSpectrum, CoronaSignalToNoiseRatio,
// CoronaSwingPosition and CoronaTrendVigor indicators.
//
// Call Update(sample) once per bar. Read IsPrimed, DominantCycle,
// DominantCycleMedian, MaximalAmplitudeSquared and the FilterBank slice.
//
// Corona is not safe for concurrent use.
type Corona struct {
	// Configuration (immutable after construction).
	minimalPeriod          int
	maximalPeriod          int
	minimalPeriodTimesTwo  int
	maximalPeriodTimesTwo  int
	filterBankLength       int
	decibelsLowerThreshold float64
	decibelsUpperThreshold float64

	// High-pass filter coefficients.
	alpha            float64
	halfOnePlusAlpha float64

	// Pre-calculated cos(4π/n) for n = MinimalPeriodTimesTwo..MaximalPeriodTimesTwo.
	preCalculatedBeta []float64

	// HP ring buffer (oldest at index 0, current at index 5).
	highPassBuffer [highPassFilterBufferSize]float64

	// Previous raw sample and previous smoothed HP (for momentum = smoothHP - smoothHPprev).
	samplePrevious   float64
	smoothHPPrevious float64

	// Filter bank.
	filterBank []Filter

	// Running maximum of amplitude-squared across all filter bins for the most
	// recent Update. Reset to zero each bar, per MBST.
	maximalAmplitudeSquared float64

	// 5-sample ring buffer (oldest first) used for the dominant cycle median.
	// Populated with math.MaxFloat64 sentinels on construction, matching MBST.
	dominantCycleBuffer [dominantCycleBufferSize]float64

	// Sample counter (number of Update calls seen).
	sampleCount int

	// Most recent dominant cycle estimate and its 5-sample median.
	dominantCycle       float64
	dominantCycleMedian float64

	primed bool
}

// NewCorona creates a new Corona engine using the provided parameters.
// Zero-valued fields in p are replaced with the Ehlers defaults.
// Pass nil to use all defaults.
func NewCorona(p *Params) (*Corona, error) {
	var cfg Params
	if p != nil {
		cfg = *p
	}
	applyDefaults(&cfg)

	if err := verifyParameters(&cfg); err != nil {
		return nil, err
	}

	c := &Corona{
		minimalPeriod:          cfg.MinimalPeriod,
		maximalPeriod:          cfg.MaximalPeriod,
		minimalPeriodTimesTwo:  cfg.MinimalPeriod * 2,
		maximalPeriodTimesTwo:  cfg.MaximalPeriod * 2,
		decibelsLowerThreshold: cfg.DecibelsLowerThreshold,
		decibelsUpperThreshold: cfg.DecibelsUpperThreshold,
	}

	c.filterBankLength = c.maximalPeriodTimesTwo - c.minimalPeriodTimesTwo + 1
	c.filterBank = make([]Filter, c.filterBankLength)

	// MBST initializes the dominant cycle buffer with MaxValue sentinels, so the
	// median is meaningful only once the buffer is fully populated with real
	// samples. We reproduce that behaviour exactly.
	for i := range c.dominantCycleBuffer {
		c.dominantCycleBuffer[i] = math.MaxFloat64
	}
	c.dominantCycle = math.MaxFloat64
	c.dominantCycleMedian = math.MaxFloat64

	// High-pass filter coefficients.
	phi := 2. * math.Pi / float64(cfg.HighPassFilterCutoff)
	c.alpha = (1. - math.Sin(phi)) / math.Cos(phi)
	c.halfOnePlusAlpha = 0.5 * (1. + c.alpha)

	// Pre-calculate β = cos(4π / n) for each half-period index n.
	c.preCalculatedBeta = make([]float64, c.filterBankLength)
	for index := 0; index < c.filterBankLength; index++ {
		n := c.minimalPeriodTimesTwo + index
		c.preCalculatedBeta[index] = math.Cos(4. * math.Pi / float64(n))
	}

	return c, nil
}

// applyDefaults fills zero-valued Params fields with defaults in-place.
func applyDefaults(p *Params) {
	if p.HighPassFilterCutoff <= 0 {
		p.HighPassFilterCutoff = defaultHighPassFilterCutoff
	}
	if p.MinimalPeriod <= 0 {
		p.MinimalPeriod = defaultMinimalPeriod
	}
	if p.MaximalPeriod <= 0 {
		p.MaximalPeriod = defaultMaximalPeriod
	}
	if p.DecibelsLowerThreshold == 0 {
		p.DecibelsLowerThreshold = defaultDecibelsLowerThreshold
	}
	if p.DecibelsUpperThreshold == 0 {
		p.DecibelsUpperThreshold = defaultDecibelsUpperThreshold
	}
}

func verifyParameters(p *Params) error {
	const (
		invalid = "invalid corona parameters"
		fmts    = "%s: %s"
	)

	if p.HighPassFilterCutoff < 2 {
		return fmt.Errorf(fmts, invalid, "HighPassFilterCutoff should be >= 2")
	}
	if p.MinimalPeriod < 2 {
		return fmt.Errorf(fmts, invalid, "MinimalPeriod should be >= 2")
	}
	if p.MaximalPeriod <= p.MinimalPeriod {
		return fmt.Errorf(fmts, invalid, "MaximalPeriod should be > MinimalPeriod")
	}
	if p.DecibelsLowerThreshold < 0 {
		return fmt.Errorf(fmts, invalid, "DecibelsLowerThreshold should be >= 0")
	}
	if p.DecibelsUpperThreshold <= p.DecibelsLowerThreshold {
		return fmt.Errorf(fmts, invalid, "DecibelsUpperThreshold should be > DecibelsLowerThreshold")
	}

	return nil
}

// MinimalPeriod returns the minimum cycle period covered by the filter bank.
func (c *Corona) MinimalPeriod() int { return c.minimalPeriod }

// MaximalPeriod returns the maximum cycle period covered by the filter bank.
func (c *Corona) MaximalPeriod() int { return c.maximalPeriod }

// MinimalPeriodTimesTwo returns the minimum filter-bank half-period index.
func (c *Corona) MinimalPeriodTimesTwo() int { return c.minimalPeriodTimesTwo }

// MaximalPeriodTimesTwo returns the maximum filter-bank half-period index.
func (c *Corona) MaximalPeriodTimesTwo() int { return c.maximalPeriodTimesTwo }

// FilterBankLength returns the number of filter bins (MaxPeriod*2 - MinPeriod*2 + 1).
func (c *Corona) FilterBankLength() int { return c.filterBankLength }

// FilterBank returns a read-only view of the filter bank. Do not mutate.
func (c *Corona) FilterBank() []Filter { return c.filterBank }

// IsPrimed reports whether the engine has seen enough samples to produce
// meaningful output (sampleCount >= MinimalPeriodTimesTwo).
func (c *Corona) IsPrimed() bool { return c.primed }

// DominantCycle returns the most recent weighted-center-of-gravity estimate of
// the dominant cycle period. Returns MinimalPeriod (as a float) before
// priming.
func (c *Corona) DominantCycle() float64 { return c.dominantCycle }

// DominantCycleMedian returns the 5-sample median of the most recent dominant
// cycle estimates.
func (c *Corona) DominantCycleMedian() float64 { return c.dominantCycleMedian }

// MaximalAmplitudeSquared returns the maximum amplitude-squared observed across
// the filter bank for the most recently processed sample. Matches MBST: reset
// to zero at the start of every Update call, not a running maximum across time.
func (c *Corona) MaximalAmplitudeSquared() float64 { return c.maximalAmplitudeSquared }

// Update feeds the next sample to the engine.
//
// The sample is assumed to be a raw price value (typically (High+Low)/2).
// NaN samples are treated as a no-op that leaves state unchanged and returns
// without marking progress.
//
// Returns true once IsPrimed has been reached and the current bar's outputs
// are meaningful.
func (c *Corona) Update(sample float64) bool {
	if math.IsNaN(sample) {
		return c.primed
	}

	c.sampleCount++

	// First sample: MBST stores it as the prior-sample reference and returns
	// with no further processing. Preserve that behaviour bit-for-bit.
	if c.sampleCount == 1 {
		c.samplePrevious = sample

		return false
	}

	// Step 1: High-pass filter.
	// HP[new] = α · HP[previous] + halfOnePlusAlpha · (sample - samplePrevious)
	hp := c.alpha*c.highPassBuffer[highPassFilterBufferSize-1] +
		c.halfOnePlusAlpha*(sample-c.samplePrevious)
	c.samplePrevious = sample

	// Shift buffer left: buffer[0] drops, newest goes to buffer[5].
	for i := 0; i < highPassFilterBufferSize-1; i++ {
		c.highPassBuffer[i] = c.highPassBuffer[i+1]
	}
	c.highPassBuffer[highPassFilterBufferSize-1] = hp

	// Step 2: 6-tap FIR smoothing with coefficients {1, 2, 3, 3, 2, 1} / 12.
	// buffer[0] is oldest → weight 1.
	smoothHP := (c.highPassBuffer[0] +
		2*c.highPassBuffer[1] +
		3*c.highPassBuffer[2] +
		3*c.highPassBuffer[3] +
		2*c.highPassBuffer[4] +
		c.highPassBuffer[5]) / firCoefSum

	// Step 3: Momentum = current smoothHP − previous smoothHP.
	momentum := smoothHP - c.smoothHPPrevious
	c.smoothHPPrevious = smoothHP

	// Step 4: Adaptive delta.
	delta := deltaFactor*float64(c.sampleCount) + deltaSummand
	if delta < deltaLowerThreshold {
		delta = deltaLowerThreshold
	}

	// Step 5: Filter-bank update. Per MBST, MaximalAmplitudeSquared is reset
	// to zero each bar and becomes the max across the current bar's bank.
	c.maximalAmplitudeSquared = 0
	for index := 0; index < c.filterBankLength; index++ {
		n := c.minimalPeriodTimesTwo + index
		nf := float64(n)

		gamma := 1. / math.Cos(8.*math.Pi*delta/nf)
		a := gamma - math.Sqrt(gamma*gamma-1.)

		quadrature := momentum * (nf / (4. * math.Pi))
		inPhase := smoothHP

		halfOneMinA := 0.5 * (1. - a)
		beta := c.preCalculatedBeta[index]
		betaOnePlusA := beta * (1. + a)

		f := &c.filterBank[index]

		real := halfOneMinA*(inPhase-f.InPhasePrevious) + betaOnePlusA*f.Real - a*f.RealPrevious
		imag := halfOneMinA*(quadrature-f.QuadraturePrevious) + betaOnePlusA*f.Imaginary - a*f.ImaginaryPrevious

		ampSq := real*real + imag*imag

		// Shift state: previous <- current, current <- new.
		f.InPhasePrevious = f.InPhase
		f.InPhase = inPhase
		f.QuadraturePrevious = f.Quadrature
		f.Quadrature = quadrature
		f.RealPrevious = f.Real
		f.Real = real
		f.ImaginaryPrevious = f.Imaginary
		f.Imaginary = imag
		f.AmplitudeSquared = ampSq

		if ampSq > c.maximalAmplitudeSquared {
			c.maximalAmplitudeSquared = ampSq
		}
	}

	// Step 6: dB normalization and dominant-cycle weighted average.
	// MBST resets DominantCycle to 0 before the loop; if no bin qualifies it
	// remains 0 and is subsequently clamped up to MinimalPeriod.
	var numerator, denominator float64
	c.dominantCycle = 0
	for index := 0; index < c.filterBankLength; index++ {
		f := &c.filterBank[index]

		decibels := 0.
		if c.maximalAmplitudeSquared > 0 {
			normalized := f.AmplitudeSquared / c.maximalAmplitudeSquared
			if normalized > 0 {
				// dB = 10 · log10( (1 − 0.99·norm) / 0.01 )
				arg := (1. - normalizedAmplitudeFactor*normalized) / decibelsFloor
				if arg > 0 {
					decibels = decibelsGain * math.Log10(arg)
				}
			}
		}

		// EMA smoothing: dB = 0.33·new + 0.67·old.
		decibels = decibelsSmoothingAlpha*decibels + decibelsSmoothingOneMinus*f.Decibels
		if decibels > c.decibelsUpperThreshold {
			decibels = c.decibelsUpperThreshold
		}
		f.Decibels = decibels

		// Only bins with dB at or below the lower threshold contribute.
		if decibels <= c.decibelsLowerThreshold {
			n := float64(c.minimalPeriodTimesTwo + index)
			adjusted := c.decibelsUpperThreshold - decibels
			numerator += n * adjusted
			denominator += adjusted
		}
	}

	// Compute DC as 0.5 · num / denom. The factor 0.5 converts the
	// half-period index space (12..60) back to period space (6..30).
	if denominator != 0 {
		c.dominantCycle = 0.5 * numerator / denominator
	}
	if c.dominantCycle < float64(c.minimalPeriod) {
		c.dominantCycle = float64(c.minimalPeriod)
	}

	// Step 7: 5-sample median of dominant cycle.
	// Shift left and push current; median is sorted[index 2] per MBST.
	for i := 0; i < dominantCycleBufferSize-1; i++ {
		c.dominantCycleBuffer[i] = c.dominantCycleBuffer[i+1]
	}
	c.dominantCycleBuffer[dominantCycleBufferSize-1] = c.dominantCycle

	var sorted [dominantCycleBufferSize]float64
	sorted = c.dominantCycleBuffer
	sort.Float64s(sorted[:])
	c.dominantCycleMedian = sorted[dominantCycleMedianIndex]
	if c.dominantCycleMedian < float64(c.minimalPeriod) {
		c.dominantCycleMedian = float64(c.minimalPeriod)
	}

	if c.sampleCount < c.minimalPeriodTimesTwo {
		return false
	}
	c.primed = true

	return true
}
