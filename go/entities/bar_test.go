//nolint:testpackage
package data

import (
	"testing"
	"time"
)

const (
	barFmt        = "expected %v, actual %v"
	barFmtRising  = "rising: expected %v, actual %v"
	barFmtFalling = "falling: expected %v, actual %v"
	barFmtFlat    = "flat: expected %v, actual %v"
)

func TestBarMedian(t *testing.T) {
	t.Parallel()

	b := bar(0, 3, 2, 0, 0)
	//nolint:ifshort
	exp := (b.Low + b.High) / 2

	if act := b.Median(); act != exp {
		t.Errorf(barFmt, exp, act)
	}
}

func TestBarTypical(t *testing.T) {
	t.Parallel()

	b := bar(0, 4, 2, 3, 0)
	exp := (b.Low + b.High + b.Close) / 3

	if act := b.Typical(); act != exp {
		t.Errorf(barFmt, exp, act)
	}
}

func TestBarWeighted(t *testing.T) {
	t.Parallel()

	b := bar(0, 4, 2, 3, 0)
	exp := (b.Low + b.High + b.Close + b.Close) / 4

	if act := b.Weighted(); act != exp {
		t.Errorf(barFmt, exp, act)
	}
}

func TestBarAverage(t *testing.T) {
	t.Parallel()

	b := bar(3, 5, 2, 4, 0)
	exp := (b.Low + b.High + b.Open + b.Close) / 4

	if act := b.Average(); act != exp {
		t.Errorf(barFmt, exp, act)
	}
}

func TestBarIsRising(t *testing.T) {
	t.Parallel()

	b := bar(2, 0, 0, 3, 0)
	exp := true
	act := b.IsRising()

	if act != exp {
		t.Errorf(barFmtRising, exp, act)
	}

	b = bar(3, 0, 0, 2, 0)
	exp = false
	act = b.IsRising()

	if act != exp {
		t.Errorf(barFmtFalling, exp, act)
	}

	b = bar(0, 0, 0, 0, 0)
	act = b.IsRising()

	if act != exp {
		t.Errorf(barFmtFlat, exp, act)
	}
}

func TestBarIsFalling(t *testing.T) {
	t.Parallel()

	b := bar(2, 0, 0, 3, 0)
	exp := false
	act := b.IsFalling()

	if act != exp {
		t.Errorf(barFmtRising, exp, act)
	}

	b = bar(3, 0, 0, 2, 0)
	exp = true
	act = b.IsFalling()

	if act != exp {
		t.Errorf(barFmtFalling, exp, act)
	}

	b = bar(0, 0, 0, 0, 0)
	exp = false
	act = b.IsFalling()

	if act != exp {
		t.Errorf(barFmtFlat, exp, act)
	}
}

func TestBarString(t *testing.T) {
	t.Parallel()

	b := bar(2, 3, 4, 5, 6)
	exp := "{2021-04-01 00:00:00, 2.000000, 3.000000, 4.000000, 5.000000, 6.000000}"

	if act := b.String(); act != exp {
		t.Errorf(barFmt, exp, act)
	}
}

func bar(o, h, l, c, v float64) Bar {
	return Bar{
		Time: time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{}),
		Open: o, High: h, Low: l, Close: c, Volume: v,
	}
}
