package hilberttransformer

import (
	"bytes"
	"fmt"
)

// CycleEstimatorType enumerates types of techniques to estimate
// an instantaneous period using a Hilbert transformer.
type CycleEstimatorType int

const (
	// HomodyneDiscriminator identifies an instantaneous period estimation
	// based on the homodyne discriminator technique.
	HomodyneDiscriminator CycleEstimatorType = iota + 1

	// HomodyneDiscriminatorTaLib identifies an instantaneous period estimation
	// based on the homodyne discriminator technique (TA-Lib implementation with unrolled loops).
	HomodyneDiscriminatorUnrolled

	// PhaseAccumulation identifies an instantaneous period estimation
	// based on the phase accumulation technique.
	PhaseAccumulator

	// DualDifferentiator identifies an instantaneous period estimation
	// based on the dual differentiation technique.
	DualDifferentiator
	last
)

const (
	unknown                       = "unknown"
	homodyneDiscriminator         = "homodyneDiscriminator"
	homodyneDiscriminatorUnrolled = "homodyneDiscriminatorUnrolled"
	phaseAccumulator              = "phaseAccumulator"
	dualDifferentiator            = "dualDifferentiator"
)

// String implements the Stringer interface.
func (s CycleEstimatorType) String() string {
	switch s {
	case HomodyneDiscriminator:
		return homodyneDiscriminator
	case HomodyneDiscriminatorUnrolled:
		return homodyneDiscriminatorUnrolled
	case PhaseAccumulator:
		return phaseAccumulator
	case DualDifferentiator:
		return dualDifferentiator
	default:
		return unknown
	}
}

// IsKnown determines if this cycle estimator type is known.
func (s CycleEstimatorType) IsKnown() bool {
	return s >= HomodyneDiscriminator && s < last
}

// MarshalJSON implements the Marshaler interface.
func (s CycleEstimatorType) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown cycle estimator type"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	str := s.String()
	if str == unknown {
		return nil, fmt.Errorf(errFmt, str)
	}

	b := make([]byte, 0, len(str)+extra)
	b = append(b, dqc)
	b = append(b, str...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (s *CycleEstimatorType) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown cycle estimator type"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	str := string(d)

	switch str {
	case homodyneDiscriminator:
		*s = HomodyneDiscriminator
	case homodyneDiscriminatorUnrolled:
		*s = HomodyneDiscriminatorUnrolled
	case phaseAccumulator:
		*s = PhaseAccumulator
	case dualDifferentiator:
		*s = DualDifferentiator
	default:
		return fmt.Errorf(errFmt, str)
	}

	return nil
}
