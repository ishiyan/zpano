//nolint:testpackage
package outputs

import (
	"math"
	"testing"
	"time"
)

func TestLevelsNew(t *testing.T) {
	t.Parallel()

	check := func(name string, exp, act any) {
		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	tm := testLevelsTime()

	t.Run("new initialized levels", func(t *testing.T) {
		t.Parallel()

		entries := []Level{
			NewLevel(105.5, 3, 0.8),
			NewLevel(102.0, 1, 0.6),
		}
		l := NewLevels(tm, entries)
		check("Time", tm, l.Time)
		check("len(Levels)", 2, len(l.Levels))
		check("Levels[0].Value", 105.5, l.Levels[0].Value)
		check("Levels[0].Offset", 3, l.Levels[0].Offset)
		check("Levels[0].Strength", 0.8, l.Levels[0].Strength)
	})

	t.Run("new value level has NaN strength", func(t *testing.T) {
		t.Parallel()

		lv := NewValueLevel(42.0)
		check("Value", 42.0, lv.Value)
		check("Offset", 0, lv.Offset)
		check("Strength is NaN", true, math.IsNaN(lv.Strength))
	})

	t.Run("new empty levels", func(t *testing.T) {
		t.Parallel()

		l := NewEmptyLevels(tm)
		check("Time", tm, l.Time)
		check("len(Levels)", 0, len(l.Levels))
	})
}

func TestLevelsIsEmpty(t *testing.T) {
	t.Parallel()

	check := func(condition string, exp, act any) {
		if exp != act {
			t.Errorf("(%s): IsEmpty is incorrect: expected %v, actual %v", condition, exp, act)
		}
	}

	l := NewEmptyLevels(testLevelsTime())
	check("empty", true, l.IsEmpty())

	l = NewLevels(testLevelsTime(), []Level{NewValueLevel(1.)})
	check("one level", false, l.IsEmpty())

	l = NewLevels(testLevelsTime(), nil)
	check("nil levels", true, l.IsEmpty())
}

func TestLevelsString(t *testing.T) {
	t.Parallel()

	l := NewLevels(testLevelsTime(), []Level{NewLevel(1., 2, 0.5), NewLevel(2., 0, 0.9)})
	expected := "{2021-04-01 00:00:00, [(1.000000, 2, 0.500000) (2.000000, 0, 0.900000)]}"

	if actual := l.String(); actual != expected {
		t.Errorf("expected %s, actual %s", expected, actual)
	}

	le := NewEmptyLevels(testLevelsTime())
	expectedEmpty := "{2021-04-01 00:00:00, []}"

	if actual := le.String(); actual != expectedEmpty {
		t.Errorf("expected %s, actual %s", expectedEmpty, actual)
	}
}

func testLevelsTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}
