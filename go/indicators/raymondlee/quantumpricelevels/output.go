//nolint:dupl
package quantumpricelevels

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Lambda is the anharmonic coefficient of the quantum potential well.
	Lambda Output = iota + 1

	// ReturnStdDev is the population standard deviation of the price-return ratios in the window.
	ReturnStdDev

	// NormalizedMultipliers is the normalized QPR multipliers (1 + scaleFactor*sigma*QPR(n)), one per level.
	NormalizedMultipliers

	// Resistances is the resistance price levels above the current price (price * NQPR(n)).
	Resistances

	// Supports is the support price levels below the current price (price / NQPR(n)).
	Supports

	outputLast
)

const (
	lambdaStr       = "lambda"
	returnStdDevStr = "returnStdDev"
	nqprStr         = "normalizedMultipliers"
	resistancesStr  = "resistances"
	supportsStr     = "supports"
	unknownStr      = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Lambda:
		return lambdaStr
	case ReturnStdDev:
		return returnStdDevStr
	case NormalizedMultipliers:
		return nqprStr
	case Resistances:
		return resistancesStr
	case Supports:
		return supportsStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Lambda && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown quantum price levels output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == unknownStr {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *Output) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown quantum price levels output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case lambdaStr:
		*o = Lambda
	case returnStdDevStr:
		*o = ReturnStdDev
	case nqprStr:
		*o = NormalizedMultipliers
	case resistancesStr:
		*o = Resistances
	case supportsStr:
		*o = Supports
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
