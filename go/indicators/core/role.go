package core

import (
	"bytes"
	"errors"
	"fmt"
)

// Role classifies the semantic role a single indicator output plays in analysis.
type Role int

const (
	// Smoother denotes a trend-following line that smooths price action.
	Smoother Role = iota + 1

	// Envelope denotes upper/lower channel bounds drawn around price.
	Envelope

	// Overlay denotes a generic overlay drawn on the price pane (e.g., SAR dots).
	Overlay

	// Polyline denotes a variable-length sequence of (offset, value) points.
	Polyline

	// Oscillator denotes a centered, unbounded momentum-style series.
	Oscillator

	// BoundedOscillator denotes an oscillator confined to a fixed range (e.g., 0..100).
	BoundedOscillator

	// Volatility denotes a dispersion-style measure (standard deviation, ATR, etc.).
	Volatility

	// VolumeFlow denotes an accumulation/distribution-style volume flow measure.
	VolumeFlow

	// Directional denotes a direction-of-movement measure (DI/DM family).
	Directional

	// CyclePeriod denotes a dominant cycle length output.
	CyclePeriod

	// CyclePhase denotes a dominant cycle phase/angle output.
	CyclePhase

	// FractalDimension denotes a fractal-dimension-style measure.
	FractalDimension

	// Spectrum denotes a multi-row spectral heat-map column.
	Spectrum

	// Signal denotes a derived signal line (e.g., MACD signal).
	Signal

	// Histogram denotes a bar-style difference series.
	Histogram

	// RegimeFlag denotes a discrete regime/state indicator.
	RegimeFlag

	// Correlation denotes a correlation-coefficient-style measure.
	Correlation
	roleLast
)

const (
	roleUnknown           = "unknown"
	roleSmoother          = "smoother"
	roleEnvelope          = "envelope"
	roleOverlay           = "overlay"
	rolePolyline          = "polyline"
	roleOscillator        = "oscillator"
	roleBoundedOscillator = "boundedOscillator"
	roleVolatility        = "volatility"
	roleVolumeFlow        = "volumeFlow"
	roleDirectional       = "directional"
	roleCyclePeriod       = "cyclePeriod"
	roleCyclePhase        = "cyclePhase"
	roleFractalDimension  = "fractalDimension"
	roleSpectrum          = "spectrum"
	roleSignal            = "signal"
	roleHistogram         = "histogram"
	roleRegimeFlag        = "regimeFlag"
	roleCorrelation       = "correlation"
)

var errUnknownRole = errors.New("unknown indicator role")

// String implements the Stringer interface.
func (r Role) String() string {
	switch r {
	case Smoother:
		return roleSmoother
	case Envelope:
		return roleEnvelope
	case Overlay:
		return roleOverlay
	case Polyline:
		return rolePolyline
	case Oscillator:
		return roleOscillator
	case BoundedOscillator:
		return roleBoundedOscillator
	case Volatility:
		return roleVolatility
	case VolumeFlow:
		return roleVolumeFlow
	case Directional:
		return roleDirectional
	case CyclePeriod:
		return roleCyclePeriod
	case CyclePhase:
		return roleCyclePhase
	case FractalDimension:
		return roleFractalDimension
	case Spectrum:
		return roleSpectrum
	case Signal:
		return roleSignal
	case Histogram:
		return roleHistogram
	case RegimeFlag:
		return roleRegimeFlag
	case Correlation:
		return roleCorrelation
	default:
		return roleUnknown
	}
}

// IsKnown determines if this role is known.
func (r Role) IsKnown() bool {
	return r >= Smoother && r < roleLast
}

// MarshalJSON implements the Marshaler interface.
func (r Role) MarshalJSON() ([]byte, error) {
	s := r.String()
	if s == roleUnknown {
		return nil, fmt.Errorf("cannot marshal '%s': %w", s, errUnknownRole)
	}

	const extra = 2

	b := make([]byte, 0, len(s)+extra)
	b = append(b, '"')
	b = append(b, s...)
	b = append(b, '"')

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (r *Role) UnmarshalJSON(data []byte) error {
	d := bytes.Trim(data, "\"")
	s := string(d)

	switch s {
	case roleSmoother:
		*r = Smoother
	case roleEnvelope:
		*r = Envelope
	case roleOverlay:
		*r = Overlay
	case rolePolyline:
		*r = Polyline
	case roleOscillator:
		*r = Oscillator
	case roleBoundedOscillator:
		*r = BoundedOscillator
	case roleVolatility:
		*r = Volatility
	case roleVolumeFlow:
		*r = VolumeFlow
	case roleDirectional:
		*r = Directional
	case roleCyclePeriod:
		*r = CyclePeriod
	case roleCyclePhase:
		*r = CyclePhase
	case roleFractalDimension:
		*r = FractalDimension
	case roleSpectrum:
		*r = Spectrum
	case roleSignal:
		*r = Signal
	case roleHistogram:
		*r = Histogram
	case roleRegimeFlag:
		*r = RegimeFlag
	case roleCorrelation:
		*r = Correlation
	default:
		return fmt.Errorf("cannot unmarshal '%s': %w", s, errUnknownRole)
	}

	return nil
}
