//nolint:testpackage
package data

import (
	"testing"
	"time"
)

func TestQuoteComponentFunc(t *testing.T) {
	t.Parallel()

	q := Quote{
		Time: time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{}),
		Ask:  1., Bid: 2., AskSize: 3., BidSize: 4.,
	}

	tests := []struct {
		c QuoteComponent
		r float64
		e bool
	}{
		{QuoteBidPrice, 2., false},
		{QuoteAskPrice, 1., false},
		{QuoteBidSize, 4., false},
		{QuoteAskSize, 3., false},
		{QuoteMidPrice, (1. + 2.) / 2., false},
		{QuoteWeightedPrice, (1.*3. + 2.*4.) / (3. + 4.), false},
		{QuoteWeightedMidPrice, (1.*4. + 2.*3.) / (3. + 4.), false},
		{QuoteSpreadBp, 10000 * 2. * (1. - 2.) / (1. + 2.), false},
		{quoteLast, 0, true},
		{QuoteComponent(0), 0, true},
		{QuoteComponent(9999), 0, true},
		{QuoteComponent(-9999), 0, true},
	}

	for _, tt := range tests {
		f, e := QuoteComponentFunc(tt.c)
		eAct := e != nil
		eExp := tt.e

		if eExp != eAct {
			t.Errorf("QuoteComponentFunc('%v') error: expected '%v', actual '%v'", tt.c, eExp, eAct)
		}

		if eExp && f != nil {
			t.Errorf("QuoteComponentFunc('%v') expected function, actual nil", tt.c)
		}

		if f == nil {
			continue
		}

		rAct := f(&q)
		rExp := tt.r

		if rExp != rAct {
			t.Errorf("QuoteComponentFunc('%v') result: expected '%v', actual '%v'", tt.c, rExp, rAct)
		}
	}
}

func TestQuoteComponentString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		c    QuoteComponent
		text string
	}{
		{QuoteBidPrice, quoteBid},
		{QuoteAskPrice, quoteAsk},
		{QuoteBidSize, quoteBidSize},
		{QuoteAskSize, quoteAskSize},
		{QuoteMidPrice, quoteMid},
		{QuoteWeightedPrice, quoteWeighted},
		{QuoteWeightedMidPrice, quoteWeightedMid},
		{QuoteSpreadBp, quoteSpreadBp},
		{quoteLast, unknown},
		{QuoteComponent(0), unknown},
		{QuoteComponent(9999), unknown},
		{QuoteComponent(-9999), unknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.c.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.c, exp, act)
		}
	}
}

func TestQuoteComponentIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		c       QuoteComponent
		boolean bool
	}{
		{QuoteBidPrice, true},
		{QuoteAskPrice, true},
		{QuoteBidSize, true},
		{QuoteAskSize, true},
		{QuoteMidPrice, true},
		{QuoteWeightedPrice, true},
		{QuoteWeightedMidPrice, true},
		{QuoteSpreadBp, true},
		{quoteLast, false},
		{QuoteComponent(0), false},
		{QuoteComponent(9999), false},
		{QuoteComponent(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.c.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.c, exp, act)
		}
	}
}

func TestQuoteComponentMarshalJSON(t *testing.T) {
	t.Parallel()

	var nilstr string
	tests := []struct {
		c         QuoteComponent
		json      string
		succeeded bool
	}{
		{QuoteBidPrice, dqs + quoteBid + dqs, true},
		{QuoteAskPrice, dqs + quoteAsk + dqs, true},
		{QuoteBidSize, dqs + quoteBidSize + dqs, true},
		{QuoteAskSize, dqs + quoteAskSize + dqs, true},
		{QuoteMidPrice, dqs + quoteMid + dqs, true},
		{QuoteWeightedPrice, dqs + quoteWeighted + dqs, true},
		{QuoteWeightedMidPrice, dqs + quoteWeightedMid + dqs, true},
		{QuoteSpreadBp, dqs + quoteSpreadBp + dqs, true},
		{quoteLast, nilstr, false},
		{QuoteComponent(9999), nilstr, false},
		{QuoteComponent(-9999), nilstr, false},
		{QuoteComponent(0), nilstr, false},
	}

	for _, tt := range tests {
		exp := tt.json
		bs, err := tt.c.MarshalJSON()

		if err != nil && tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected success '%v', got error %v", tt.c, exp, err)

			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected error, got success", tt.c)

			continue
		}

		act := string(bs)
		if exp != act {
			t.Errorf("'%v'.MarshalJSON(): expected '%v', actual '%v'", tt.c, exp, act)
		}
	}
}

func TestQuoteComponentUnmarshalJSON(t *testing.T) {
	t.Parallel()

	var zero QuoteComponent
	tests := []struct {
		c         QuoteComponent
		json      string
		succeeded bool
	}{
		{QuoteBidPrice, dqs + quoteBid + dqs, true},
		{QuoteAskPrice, dqs + quoteAsk + dqs, true},
		{QuoteBidSize, dqs + quoteBidSize + dqs, true},
		{QuoteAskSize, dqs + quoteAskSize + dqs, true},
		{QuoteMidPrice, dqs + quoteMid + dqs, true},
		{QuoteWeightedPrice, dqs + quoteWeighted + dqs, true},
		{QuoteWeightedMidPrice, dqs + quoteWeightedMid + dqs, true},
		{QuoteSpreadBp, dqs + quoteSpreadBp + dqs, true},
		{zero, dqs + unknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.c
		bs := []byte(tt.json)

		var c QuoteComponent

		err := c.UnmarshalJSON(bs)
		if err != nil && tt.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected success '%v', got error %v", tt.json, exp, err)

			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("MarshalJSON('%v'): expected error, got success", tt.json)

			continue
		}

		if exp != c {
			t.Errorf("MarshalJSON('%v'): expected '%v', actual '%v'", tt.json, exp, c)
		}
	}
}
