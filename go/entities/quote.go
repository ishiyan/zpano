package data

import (
	"fmt"
	"time"
)

// Quote represents [price, size] pairs for the bid & ask.
type Quote struct {
	Time    time.Time `json:"time"`    // The date and time.
	Bid     float64   `json:"bid"`     // The bid price.
	Ask     float64   `json:"ask"`     // The ask price.
	BidSize float64   `json:"bidSize"` // The bid size.
	AskSize float64   `json:"askSize"` // The ask size.
}

// Mid is the mid-price, calculated as
//   (ask + bid) / 2.
func (q *Quote) Mid() float64 {
	return (q.Ask + q.Bid) / 2 //znolint:gomnd
}

// Weighted is a weighted price, calculated as
//   (ask*askSize + bid*bidSize) / (askSize + bidSize).
func (q *Quote) Weighted() float64 {
	size := q.AskSize + q.BidSize
	switch size {
	case 0:
		return 0
	default:
		return (q.Ask*q.AskSize + q.Bid*q.BidSize) / size
	}
}

// WeightedMid is a weighted mid-price (sometimes called micro-price), calculated as
//   (ask * bidSize + bid * askSize) / (askSize + bidSize).
// The reasoning is as follows. When there are many buyers, they repeatedly hit the
// ask price,lowering the available ask size. So we weight the bid less heavily, and the
// weighted mid-price rises. Likewise, when sellers arrive, the weighted mid-price falls.
func (q *Quote) WeightedMid() float64 {
	size := q.AskSize + q.BidSize
	switch size {
	case 0:
		return 0
	default:
		return (q.Ask*q.BidSize + q.Bid*q.AskSize) / size
	}
}

// SpreadBp is a spread in basis points (100 basis points = 1%), calculated as
//   10000 * (ask - bid) / mid.
func (q *Quote) SpreadBp() float64 {
	mid := q.Ask + q.Bid
	switch mid {
	case 0:
		return 0
	default:
		return 20000 * (q.Ask - q.Bid) / mid
	}
}

// String implements the Stringer interface.
func (q *Quote) String() string {
	return fmt.Sprintf("Quote(%v, %f, %f, %f, %f)", q.Time.Format(timeFmtDateTime), q.Bid, q.Ask, q.BidSize, q.AskSize)
}
