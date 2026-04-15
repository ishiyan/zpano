//nolint:testpackage
package outputs

import (
	"math"
	"testing"
	"time"
)

func TestBandNew(t *testing.T) {
	t.Parallel()

	const (
		p1 = 1.
		p2 = 2.
	)

	check := func(name string, exp, act any) {
		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkNaN := func(name string, act float64) {
		if !math.IsNaN(act) {
			t.Errorf("%s is incorrect: expected NaN, actual %v", name, act)
		}
	}

	time := testBandTime()

	t.Run("new initialized band, lower < upper", func(t *testing.T) {
		t.Parallel()

		b := NewBand(time, p1, p2)
		check("Time", time, b.Time)
		check("Lower", p1, b.Lower)
		check("Upper", p2, b.Upper)
	})

	t.Run("new initialized band, lower > upper", func(t *testing.T) {
		t.Parallel()

		b := NewBand(time, p2, p1)
		check("Time", time, b.Time)
		check("Lower", p1, b.Lower)
		check("Upper", p2, b.Upper)
	})

	t.Run("new empty band", func(t *testing.T) {
		t.Parallel()

		b := NewEmptyBand(time)
		check("Time", time, b.Time)
		checkNaN("Lower", b.Lower)
		checkNaN("Upper", b.Upper)
	})
}

func TestBandIsEmpty(t *testing.T) {
	t.Parallel()

	check := func(condition string, exp, act any) {
		if exp != act {
			t.Errorf("(%s): IsEmpty is incorrect: expected %v, actual %v", condition, exp, act)
		}
	}

	b := testBandCreate()
	check("Lower and Upper not NaN", false, b.IsEmpty())

	b.Lower = math.NaN()
	check("Lower is NaN", true, b.IsEmpty())

	b.Upper = math.NaN()
	check("Lower and Upper are NaN", true, b.IsEmpty())

	b.Lower = 1.
	check("Upper is NaN", true, b.IsEmpty())
}

func TestBandString(t *testing.T) {
	t.Parallel()

	b := testBandCreate()
	expected := "{2021-04-01 00:00:00, 1.000000, 2.000000}"

	if actual := b.String(); actual != expected {
		t.Errorf("expected %s, actual %s", expected, actual)
	}
}

func testBandCreate() Band {
	return Band{Time: testBandTime(), Lower: 1., Upper: 2.}
}

func testBandTime() time.Time { return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{}) }
