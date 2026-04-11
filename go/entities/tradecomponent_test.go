//nolint:testpackage
package data

import (
	"testing"
	"time"
)

func TestTradeComponentFunc(t *testing.T) {
	t.Parallel()

	tr := Trade{
		Time:  time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{}),
		Price: 1., Volume: 2.,
	}

	tests := []struct {
		c TradeComponent
		r float64
		e bool
	}{
		{TradePrice, 1., false},
		{TradeVolume, 2., false},
		{tradeLast, 0, true},
		{TradeComponent(0), 0, true},
		{TradeComponent(9999), 0, true},
		{TradeComponent(-9999), 0, true},
	}

	for _, tt := range tests {
		f, e := TradeComponentFunc(tt.c)
		eAct := e != nil
		eExp := tt.e

		if eExp != eAct {
			t.Errorf("TradeComponentFunc('%v') error: expected '%v', actual '%v'", tt.c, eExp, eAct)
		}

		if eExp && f != nil {
			t.Errorf("TradeComponentFunc('%v') expected function, actual nil", tt.c)
		}

		if f == nil {
			continue
		}

		rAct := f(&tr)
		rExp := tt.r

		if rExp != rAct {
			t.Errorf("TradeComponentFunc('%v') result: expected '%v', actual '%v'", tt.c, rExp, rAct)
		}
	}
}

func TestTradeComponentString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		c    TradeComponent
		text string
	}{
		{TradePrice, tradePrice},
		{TradeVolume, tradeVolume},
		{tradeLast, unknown},
		{TradeComponent(0), unknown},
		{TradeComponent(9999), unknown},
		{TradeComponent(-9999), unknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.c.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.c, exp, act)
		}
	}
}

func TestTradeComponentIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		c       TradeComponent
		boolean bool
	}{
		{TradePrice, true},
		{TradeVolume, true},
		{tradeLast, false},
		{TradeComponent(0), false},
		{TradeComponent(9999), false},
		{TradeComponent(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.c.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.c, exp, act)
		}
	}
}

func TestTradeComponentMarshalJSON(t *testing.T) {
	t.Parallel()

	var nilstr string
	tests := []struct {
		c         TradeComponent
		json      string
		succeeded bool
	}{
		{TradePrice, dqs + tradePrice + dqs, true},
		{TradeVolume, dqs + tradeVolume + dqs, true},
		{tradeLast, nilstr, false},
		{TradeComponent(9999), nilstr, false},
		{TradeComponent(-9999), nilstr, false},
		{TradeComponent(0), nilstr, false},
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

func TestTradeComponentUnmarshalJSON(t *testing.T) {
	t.Parallel()

	var zero TradeComponent
	tests := []struct {
		c         TradeComponent
		json      string
		succeeded bool
	}{
		{TradePrice, dqs + tradePrice + dqs, true},
		{TradeVolume, dqs + tradeVolume + dqs, true},
		{zero, dqs + unknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.c
		bs := []byte(tt.json)

		var c TradeComponent

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
