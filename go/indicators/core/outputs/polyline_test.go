//nolint:testpackage
package outputs

import (
	"testing"
	"time"
)

func TestPolylineNew(t *testing.T) {
	t.Parallel()

	check := func(name string, exp, act any) {
		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	tm := testPolylineTime()

	t.Run("new initialized polyline", func(t *testing.T) {
		t.Parallel()

		points := []Point{{Offset: 3, Value: 10.}, {Offset: 1, Value: 20.}, {Offset: 0, Value: 15.}}
		p := NewPolyline(tm, points)
		check("Time", tm, p.Time)
		check("len(Points)", 3, len(p.Points))
		check("Points[0].Offset", 3, p.Points[0].Offset)
		check("Points[0].Value", 10., p.Points[0].Value)
		check("Points[2].Offset", 0, p.Points[2].Offset)
		check("Points[2].Value", 15., p.Points[2].Value)
	})

	t.Run("new empty polyline", func(t *testing.T) {
		t.Parallel()

		p := NewEmptyPolyline(tm)
		check("Time", tm, p.Time)
		check("len(Points)", 0, len(p.Points))
	})
}

func TestPolylineIsEmpty(t *testing.T) {
	t.Parallel()

	check := func(condition string, exp, act any) {
		if exp != act {
			t.Errorf("(%s): IsEmpty is incorrect: expected %v, actual %v", condition, exp, act)
		}
	}

	p := NewEmptyPolyline(testPolylineTime())
	check("empty", true, p.IsEmpty())

	p = NewPolyline(testPolylineTime(), []Point{{Offset: 0, Value: 1.}})
	check("one point", false, p.IsEmpty())

	p = NewPolyline(testPolylineTime(), nil)
	check("nil points", true, p.IsEmpty())
}

func TestPolylineString(t *testing.T) {
	t.Parallel()

	p := NewPolyline(testPolylineTime(), []Point{{Offset: 2, Value: 1.}, {Offset: 0, Value: 2.}})
	expected := "{2021-04-01 00:00:00, [(2, 1.000000) (0, 2.000000)]}"

	if actual := p.String(); actual != expected {
		t.Errorf("expected %s, actual %s", expected, actual)
	}

	pe := NewEmptyPolyline(testPolylineTime())
	expectedEmpty := "{2021-04-01 00:00:00, []}"

	if actual := pe.String(); actual != expectedEmpty {
		t.Errorf("expected %s, actual %s", expectedEmpty, actual)
	}
}

func testPolylineTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}
