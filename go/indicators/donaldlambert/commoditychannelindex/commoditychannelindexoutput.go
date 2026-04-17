//nolint:dupl
package commoditychannelindex

import (
	"bytes"
	"fmt"
)

// CommodityChannelIndexOutput describes the outputs of the indicator.
type CommodityChannelIndexOutput int

const (
	// The scalar value of the commodity channel index.
	CommodityChannelIndexValue CommodityChannelIndexOutput = iota + 1
	commodityChannelIndexLast
)

const (
	commodityChannelIndexOutputValue   = "value"
	commodityChannelIndexOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o CommodityChannelIndexOutput) String() string {
	switch o {
	case CommodityChannelIndexValue:
		return commodityChannelIndexOutputValue
	default:
		return commodityChannelIndexOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o CommodityChannelIndexOutput) IsKnown() bool {
	return o >= CommodityChannelIndexValue && o < commodityChannelIndexLast
}

// MarshalJSON implements the Marshaler interface.
func (o CommodityChannelIndexOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown commodity channel index output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == commodityChannelIndexOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *CommodityChannelIndexOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown commodity channel index output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case commodityChannelIndexOutputValue:
		*o = CommodityChannelIndexValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
