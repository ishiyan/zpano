//nolint:testpackage
package data

import (
	"testing"
	"time"
)

func TestQuoteMid(t *testing.T) {
	t.Parallel()

	q := Quote{Time: time.Time{}, Bid: 3.0, Ask: 2.0, BidSize: 0, AskSize: 0}
	expected := (q.Ask + q.Bid) / 2

	if actual := q.Mid(); actual != expected {
		t.Errorf("expected %f, actual %f", expected, actual)
	}
}

func TestQuoteWeighted(t *testing.T) {
	t.Parallel()

	q := Quote{Time: time.Time{}, Bid: 3.0, Ask: 2.0, BidSize: 5.0, AskSize: 4.0}
	expected := (q.Ask*q.AskSize + q.Bid*q.BidSize) / (q.AskSize + q.BidSize)
	actual := q.Weighted()

	if actual != expected {
		t.Errorf("expected %f, actual %f", expected, actual)
	}

	q = Quote{Time: time.Time{}, Bid: 3.0, Ask: 2.0, BidSize: 0, AskSize: 0}
	expected = 0.0
	actual = q.Weighted()

	if actual != expected {
		t.Errorf("zero size: expected %f, actual %f", expected, actual)
	}
}

func TestQuoteWeightedMid(t *testing.T) {
	t.Parallel()

	q := Quote{Time: time.Time{}, Bid: 3.0, Ask: 2.0, BidSize: 5.0, AskSize: 4.0}
	expected := (q.Ask*q.BidSize + q.Bid*q.AskSize) / (q.AskSize + q.BidSize)
	actual := q.WeightedMid()

	if actual != expected {
		t.Errorf("expected %f, actual %f", expected, actual)
	}

	q = Quote{Time: time.Time{}, Bid: 3.0, Ask: 2.0, BidSize: 0, AskSize: 0}
	expected = 0.0
	actual = q.WeightedMid()

	if actual != expected {
		t.Errorf("zero size: expected %f, actual %f", expected, actual)
	}
}

func TestQuoteSpreadBp(t *testing.T) {
	t.Parallel()

	q := Quote{Time: time.Time{}, Bid: 3.0, Ask: 2.0, BidSize: 0, AskSize: 0}
	expected := 20000 * (q.Ask - q.Bid) / (q.Ask + q.Bid)
	actual := q.SpreadBp()

	if actual != expected {
		t.Errorf("expected %f, actual %f", expected, actual)
	}

	q = Quote{Time: time.Time{}, Bid: 0, Ask: 0, BidSize: 0, AskSize: 0}
	expected = 0.0
	actual = q.SpreadBp()

	if actual != expected {
		t.Errorf("zero mid: expected %f, actual %f", expected, actual)
	}
}

func TestQuoteString(t *testing.T) {
	t.Parallel()

	q := Quote{
		Time: time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{}),
		Ask:  2.0, Bid: 3.0, AskSize: 4.0, BidSize: 5.0,
	}
	expected := "Quote(2021-04-01 00:00:00, 3.000000, 2.000000, 5.000000, 4.000000)"

	if actual := q.String(); actual != expected {
		t.Errorf("expected %s, actual %s", expected, actual)
	}
}
