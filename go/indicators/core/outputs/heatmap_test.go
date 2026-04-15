//nolint:testpackage
package outputs

import (
	"math"
	"testing"
	"time"
)

func TestHeatmapNew(t *testing.T) {
	t.Parallel()

	const (
		p1 = 1.
		p2 = 2.
		p3 = 3.
		p4 = 4.
		p5 = 5.
		p6 = 6.
		p7 = 7.
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkNaN := func(name string, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("%s is incorrect: expected NaN, actual %v", name, act)
		}
	}

	time := testBandTime()

	t.Run("new initialized heatmap", func(t *testing.T) {
		t.Parallel()

		h := NewHeatmap(time, p1, p2, p3, p4, p5, []float64{p6, p7})
		check("Time", time, h.Time)
		check("ParameterFirst", p1, h.ParameterFirst)
		check("ParameterLast", p2, h.ParameterLast)
		check("ParameterResolution", p3, h.ParameterResolution)
		check("ValueMin", p4, h.ValueMin)
		check("ValueMax", p5, h.ValueMax)
		check("Values length", 2, len(h.Values))
		check("Values [0]", p6, h.Values[0])
		check("Values [1]", p7, h.Values[1])
	})

	t.Run("new empty heatmap", func(t *testing.T) {
		t.Parallel()

		h := NewEmptyHeatmap(time, p1, p2, p3)
		check("Time", time, h.Time)
		check("ParameterFirst", p1, h.ParameterFirst)
		check("ParameterLast", p2, h.ParameterLast)
		check("ParameterResolution", p3, h.ParameterResolution)
		checkNaN("ValueMin", h.ValueMin)
		checkNaN("ValueMax", h.ValueMax)
		check("Values length", 0, len(h.Values))
	})
}

func TestHeatmapIsEmpty(t *testing.T) {
	t.Parallel()

	h := testHeatmapCreate()
	if h.IsEmpty() {
		t.Error("expected not empty, actual is empty")
	}

	h.Values = nil
	if !h.IsEmpty() {
		t.Error("expected empty (Values is nil), actual not empty")
	}

	h.Values = []float64{}
	if !h.IsEmpty() {
		t.Error("expected empty (Values length is 0), actual not empty")
	}
}

func TestHeatmapString(t *testing.T) {
	t.Parallel()

	h := testHeatmapCreate()
	expected := "{2021-04-01 00:00:00, (1.000000, 2.000000, 3.000000), (4.000000, 5.000000), [6 7]}"

	if actual := h.String(); actual != expected {
		t.Errorf("expected %s, actual %s", expected, actual)
	}
}

func testHeatmapCreate() Heatmap {
	return Heatmap{
		Time: testHeatmapTime(), ParameterFirst: 1., ParameterLast: 2., ParameterResolution: 3,
		ValueMin: 4., ValueMax: 5., Values: []float64{6., 7.},
	}
}

func testHeatmapTime() time.Time { return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{}) }
