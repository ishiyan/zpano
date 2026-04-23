package core

import (
	"bytes"
	"errors"
	"fmt"
)

// InputRequirement classifies the minimum input data type an indicator consumes.
type InputRequirement int

const (
	// ScalarInput denotes an indicator that consumes a scalar time series (e.g., prices).
	ScalarInput InputRequirement = iota + 1

	// QuoteInput denotes an indicator that consumes level-1 quotes.
	QuoteInput

	// BarInput denotes an indicator that consumes OHLCV bars.
	BarInput

	// TradeInput denotes an indicator that consumes individual trades.
	TradeInput
	inputRequirementLast
)

const (
	inputRequirementUnknown = "unknown"
	inputRequirementScalar  = "scalar"
	inputRequirementQuote   = "quote"
	inputRequirementBar     = "bar"
	inputRequirementTrade   = "trade"
)

var errUnknownInputRequirement = errors.New("unknown indicator input requirement")

// String implements the Stringer interface.
func (s InputRequirement) String() string {
	switch s {
	case ScalarInput:
		return inputRequirementScalar
	case QuoteInput:
		return inputRequirementQuote
	case BarInput:
		return inputRequirementBar
	case TradeInput:
		return inputRequirementTrade
	default:
		return inputRequirementUnknown
	}
}

// IsKnown determines if this input requirement is known.
func (s InputRequirement) IsKnown() bool {
	return s >= ScalarInput && s < inputRequirementLast
}

// MarshalJSON implements the Marshaler interface.
func (s InputRequirement) MarshalJSON() ([]byte, error) {
	str := s.String()
	if str == inputRequirementUnknown {
		return nil, fmt.Errorf("cannot marshal '%s': %w", str, errUnknownInputRequirement)
	}

	const extra = 2

	b := make([]byte, 0, len(str)+extra)
	b = append(b, '"')
	b = append(b, str...)
	b = append(b, '"')

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (s *InputRequirement) UnmarshalJSON(data []byte) error {
	d := bytes.Trim(data, "\"")
	str := string(d)

	switch str {
	case inputRequirementScalar:
		*s = ScalarInput
	case inputRequirementQuote:
		*s = QuoteInput
	case inputRequirementBar:
		*s = BarInput
	case inputRequirementTrade:
		*s = TradeInput
	default:
		return fmt.Errorf("cannot unmarshal '%s': %w", str, errUnknownInputRequirement)
	}

	return nil
}
