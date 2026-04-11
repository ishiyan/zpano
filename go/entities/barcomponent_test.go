//nolint:testpackage
package data

import (
	"testing"
	"time"
)

func TestBarComponentFunc(t *testing.T) {
	t.Parallel()

	b := Bar{
		Time: time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{}),
		Open: 2., High: 4., Low: 1., Close: 3., Volume: 5.,
	}

	tests := []struct {
		c BarComponent
		r float64
		e bool
	}{
		{BarOpenPrice, 2., false},
		{BarHighPrice, 4., false},
		{BarLowPrice, 1., false},
		{BarClosePrice, 3., false},
		{BarVolume, 5., false},
		{BarMedianPrice, (1. + 4.) / 2., false},
		{BarTypicalPrice, (1. + 4. + 3.) / 3., false},
		{BarWeightedPrice, (1. + 4. + 3. + 3.) / 4., false},
		{BarAveragePrice, (1. + 4. + 3. + 2.) / 4., false},
		{barLast, 0, true},
		{BarComponent(0), 0, true},
		{BarComponent(9999), 0, true},
		{BarComponent(-9999), 0, true},
	}

	for _, tt := range tests {
		f, e := BarComponentFunc(tt.c)
		eAct := e != nil
		eExp := tt.e

		if eExp != eAct {
			t.Errorf("BarComponentFunc('%v') error: expected '%v', actual '%v'", tt.c, eExp, eAct)
		}

		if eExp && f != nil {
			t.Errorf("BarComponentFunc('%v') expected function, actual nil", tt.c)
		}

		if f == nil {
			continue
		}

		rAct := f(&b)
		rExp := tt.r

		if rExp != rAct {
			t.Errorf("BarComponentFunc('%v') result: expected '%v', actual '%v'", tt.c, rExp, rAct)
		}
	}
}

func TestBarComponentString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		c    BarComponent
		text string
	}{
		{BarOpenPrice, barOpen},
		{BarHighPrice, barHigh},
		{BarLowPrice, barLow},
		{BarClosePrice, barClose},
		{BarVolume, barVolume},
		{BarMedianPrice, barMedian},
		{BarTypicalPrice, barTypical},
		{BarWeightedPrice, barWeighted},
		{BarAveragePrice, barAverage},
		{barLast, unknown},
		{BarComponent(0), unknown},
		{BarComponent(9999), unknown},
		{BarComponent(-9999), unknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.c.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.c, exp, act)
		}
	}
}

func TestBarComponentIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		c       BarComponent
		boolean bool
	}{
		{BarOpenPrice, true},
		{BarHighPrice, true},
		{BarLowPrice, true},
		{BarClosePrice, true},
		{BarVolume, true},
		{BarMedianPrice, true},
		{BarTypicalPrice, true},
		{BarWeightedPrice, true},
		{BarAveragePrice, true},
		{barLast, false},
		{BarComponent(0), false},
		{BarComponent(9999), false},
		{BarComponent(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.c.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.c, exp, act)
		}
	}
}

func TestBarComponentMarshalJSON(t *testing.T) {
	t.Parallel()

	var nilstr string
	tests := []struct {
		c         BarComponent
		json      string
		succeeded bool
	}{
		{BarOpenPrice, dqs + barOpen + dqs, true},
		{BarHighPrice, dqs + barHigh + dqs, true},
		{BarLowPrice, dqs + barLow + dqs, true},
		{BarClosePrice, dqs + barClose + dqs, true},
		{BarVolume, dqs + barVolume + dqs, true},
		{BarMedianPrice, dqs + barMedian + dqs, true},
		{BarTypicalPrice, dqs + barTypical + dqs, true},
		{BarWeightedPrice, dqs + barWeighted + dqs, true},
		{BarAveragePrice, dqs + barAverage + dqs, true},
		{barLast, nilstr, false},
		{BarComponent(9999), nilstr, false},
		{BarComponent(-9999), nilstr, false},
		{BarComponent(0), nilstr, false},
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

func TestBarComponentUnmarshalJSON(t *testing.T) {
	t.Parallel()

	var zero BarComponent
	tests := []struct {
		c         BarComponent
		json      string
		succeeded bool
	}{
		{BarOpenPrice, dqs + barOpen + dqs, true},
		{BarHighPrice, dqs + barHigh + dqs, true},
		{BarLowPrice, dqs + barLow + dqs, true},
		{BarClosePrice, dqs + barClose + dqs, true},
		{BarVolume, dqs + barVolume + dqs, true},
		{BarMedianPrice, dqs + barMedian + dqs, true},
		{BarTypicalPrice, dqs + barTypical + dqs, true},
		{BarWeightedPrice, dqs + barWeighted + dqs, true},
		{BarAveragePrice, dqs + barAverage + dqs, true},
		{zero, dqs + unknown + dqs, false},
		{zero, dqs + "foobar" + dqs, false},
	}

	for _, tt := range tests {
		exp := tt.c
		bs := []byte(tt.json)

		var c BarComponent

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
