package entities

import (
	"bytes"
	"errors"
	"fmt"
)

// QuoteComponent describes a component of the Quote type.
type QuoteComponent int

// QuoteFunc defines a function to get a component value from the Quote type.
type QuoteFunc func(q *Quote) float64

// DefaultQuoteComponent is the default quote component used when no explicit component is specified.
const DefaultQuoteComponent = QuoteMidPrice

const (
	// QuoteBidPrice is the bid price component.
	QuoteBidPrice QuoteComponent = iota + 1

	// QuoteAskPrice is the ask price component.
	QuoteAskPrice

	// QuoteBidSize is the bid size component.
	QuoteBidSize

	// QuoteAskSize is the ask size component.
	QuoteAskSize

	// QuoteMidPrice is the med-price component, calculated as
	//   (ask + bid) / 2.
	QuoteMidPrice

	// QuoteWeightedPrice is the weighted price component, calculated as
	//   (ask*askSize + bid*bidSize) / (askSize + bidSize).
	QuoteWeightedPrice

	// QuoteWeightedMidPrice is the weighted mid-price component (sometimes called micro-price), calculated as
	//   (ask*bidSize + bid*askSize) / (askSize + bidSize).
	QuoteWeightedMidPrice

	// QuoteSpreadBp is the spread in basis points (100 basis points = 1%) component, calculated as
	//   10000 * (ask - bid) / mid.
	QuoteSpreadBp
	quoteLast
)

const (
	quoteBid         = "bid"
	quoteAsk         = "ask"
	quoteBidSize     = "bidSize"
	quoteAskSize     = "askSize"
	quoteMid         = "mid"
	quoteWeighted    = "weighted"
	quoteWeightedMid = "weightedMid"
	quoteSpreadBp    = "spreadBp"
)

const (
	quoteMnemonicBid         = "b"
	quoteMnemonicAsk         = "a"
	quoteMnemonicBidSize     = "bs"
	quoteMnemonicAskSize     = "as"
	quoteMnemonicMid         = "ba/2"
	quoteMnemonicWeighted    = "(bbs+aas)/(bs+as)"
	quoteMnemonicWeightedMid = "(bas+abs)/(bs+as)"
	quoteMnemonicSpreadBp    = "spread bp"
)

var errUnknownQuoteComponent = errors.New("unknown quote component")

// QuoteComponentFunc returns a QuoteFunc function to get a component value from the Quote type.
func QuoteComponentFunc(c QuoteComponent) (QuoteFunc, error) {
	switch c {
	case QuoteBidPrice:
		return func(q *Quote) float64 { return q.Bid }, nil
	case QuoteAskPrice:
		return func(q *Quote) float64 { return q.Ask }, nil
	case QuoteBidSize:
		return func(q *Quote) float64 { return q.BidSize }, nil
	case QuoteAskSize:
		return func(q *Quote) float64 { return q.AskSize }, nil
	case QuoteMidPrice:
		return func(q *Quote) float64 { return q.Mid() }, nil
	case QuoteWeightedPrice:
		return func(q *Quote) float64 { return q.Weighted() }, nil
	case QuoteWeightedMidPrice:
		return func(q *Quote) float64 { return q.WeightedMid() }, nil
	case QuoteSpreadBp:
		return func(q *Quote) float64 { return q.SpreadBp() }, nil
	default:
		return nil, fmt.Errorf("%d: %w", int(c), errUnknownQuoteComponent)
	}
}

// String implements the Stringer interface.
func (s QuoteComponent) String() string {
	switch s {
	case QuoteBidPrice:
		return quoteBid
	case QuoteAskPrice:
		return quoteAsk
	case QuoteBidSize:
		return quoteBidSize
	case QuoteAskSize:
		return quoteAskSize
	case QuoteMidPrice:
		return quoteMid
	case QuoteWeightedPrice:
		return quoteWeighted
	case QuoteWeightedMidPrice:
		return quoteWeightedMid
	case QuoteSpreadBp:
		return quoteSpreadBp
	default:
		return unknown
	}
}

// Mnemonic returns a short mnemonic code for the quote component.
func (s QuoteComponent) Mnemonic() string {
	switch s {
	case QuoteBidPrice:
		return quoteMnemonicBid
	case QuoteAskPrice:
		return quoteMnemonicAsk
	case QuoteBidSize:
		return quoteMnemonicBidSize
	case QuoteAskSize:
		return quoteMnemonicAskSize
	case QuoteMidPrice:
		return quoteMnemonicMid
	case QuoteWeightedPrice:
		return quoteMnemonicWeighted
	case QuoteWeightedMidPrice:
		return quoteMnemonicWeightedMid
	case QuoteSpreadBp:
		return quoteMnemonicSpreadBp
	default:
		return unknown
	}
}

// IsKnown determines if this quote component is known.
func (s QuoteComponent) IsKnown() bool {
	return s >= QuoteBidPrice && s < quoteLast
}

// MarshalJSON implements the Marshaler interface.
func (s QuoteComponent) MarshalJSON() ([]byte, error) {
	str := s.String()
	if str == unknown {
		return nil, fmt.Errorf(marshalErrFmt, str, errUnknownQuoteComponent)
	}

	const extra = 2 // Two bytes for quotes.

	b := make([]byte, 0, len(str)+extra)
	b = append(b, dqc)
	b = append(b, str...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (s *QuoteComponent) UnmarshalJSON(data []byte) error {
	d := bytes.Trim(data, dqs)
	str := string(d)

	switch str {
	case quoteBid:
		*s = QuoteBidPrice
	case quoteAsk:
		*s = QuoteAskPrice
	case quoteBidSize:
		*s = QuoteBidSize
	case quoteAskSize:
		*s = QuoteAskSize
	case quoteMid:
		*s = QuoteMidPrice
	case quoteWeighted:
		*s = QuoteWeightedPrice
	case quoteWeightedMid:
		*s = QuoteWeightedMidPrice
	case quoteSpreadBp:
		*s = QuoteSpreadBp
	default:
		return fmt.Errorf(unmarshalErrFmt, str, errUnknownQuoteComponent)
	}

	return nil
}
