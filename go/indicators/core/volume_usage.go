package core

import (
	"bytes"
	"errors"
	"fmt"
)

// VolumeUsage classifies how an indicator uses volume information.
type VolumeUsage int

const (
	// NoVolume denotes an indicator that does not use volume.
	NoVolume VolumeUsage = iota + 1

	// AggregateBarVolume denotes an indicator that consumes per-bar aggregated volume.
	AggregateBarVolume

	// PerTradeVolume denotes an indicator that consumes per-trade volume.
	PerTradeVolume

	// QuoteLiquidityVolume denotes an indicator that consumes quote-side liquidity (bid/ask sizes).
	QuoteLiquidityVolume
	volumeUsageLast
)

const (
	volumeUsageUnknown        = "unknown"
	volumeUsageNone           = "none"
	volumeUsageAggregateBar   = "aggregateBar"
	volumeUsagePerTrade       = "perTrade"
	volumeUsageQuoteLiquidity = "quoteLiquidity"
)

var errUnknownVolumeUsage = errors.New("unknown indicator volume usage")

// String implements the Stringer interface.
func (s VolumeUsage) String() string {
	switch s {
	case NoVolume:
		return volumeUsageNone
	case AggregateBarVolume:
		return volumeUsageAggregateBar
	case PerTradeVolume:
		return volumeUsagePerTrade
	case QuoteLiquidityVolume:
		return volumeUsageQuoteLiquidity
	default:
		return volumeUsageUnknown
	}
}

// IsKnown determines if this volume usage is known.
func (s VolumeUsage) IsKnown() bool {
	return s >= NoVolume && s < volumeUsageLast
}

// MarshalJSON implements the Marshaler interface.
func (s VolumeUsage) MarshalJSON() ([]byte, error) {
	str := s.String()
	if str == volumeUsageUnknown {
		return nil, fmt.Errorf("cannot marshal '%s': %w", str, errUnknownVolumeUsage)
	}

	const extra = 2

	b := make([]byte, 0, len(str)+extra)
	b = append(b, '"')
	b = append(b, str...)
	b = append(b, '"')

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (s *VolumeUsage) UnmarshalJSON(data []byte) error {
	d := bytes.Trim(data, "\"")
	str := string(d)

	switch str {
	case volumeUsageNone:
		*s = NoVolume
	case volumeUsageAggregateBar:
		*s = AggregateBarVolume
	case volumeUsagePerTrade:
		*s = PerTradeVolume
	case volumeUsageQuoteLiquidity:
		*s = QuoteLiquidityVolume
	default:
		return fmt.Errorf("cannot unmarshal '%s': %w", str, errUnknownVolumeUsage)
	}

	return nil
}
