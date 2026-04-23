//nolint:testpackage
package shape

import (
	"testing"
)

func TestShapeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s    Shape
		text string
	}{
		{Scalar, scalar},
		{Band, band},
		{Heatmap, heatmap},
		{Polyline, polyline},
		{last, unknown},
		{Shape(0), unknown},
		{Shape(9999), unknown},
		{Shape(-9999), unknown},
	}

	for _, tt := range tests {
		exp := tt.text
		act := tt.s.String()

		if exp != act {
			t.Errorf("'%v'.String(): expected '%v', actual '%v'", tt.s, exp, act)
		}
	}
}

func TestShapeIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s       Shape
		boolean bool
	}{
		{Scalar, true},
		{Band, true},
		{Heatmap, true},
		{Polyline, true},
		{last, false},
		{Shape(0), false},
		{Shape(9999), false},
		{Shape(-9999), false},
	}

	for _, tt := range tests {
		exp := tt.boolean
		act := tt.s.IsKnown()

		if exp != act {
			t.Errorf("'%v'.IsKnown(): expected '%v', actual '%v'", tt.s, exp, act)
		}
	}
}

func TestShapeMarshalJSON(t *testing.T) {
	t.Parallel()

	var nilstr string
	tests := []struct {
		s         Shape
		json      string
		succeeded bool
	}{
		{Scalar, "\"scalar\"", true},
		{Band, "\"band\"", true},
		{Heatmap, "\"heatmap\"", true},
		{Polyline, "\"polyline\"", true},
		{last, nilstr, false},
		{Shape(9999), nilstr, false},
		{Shape(-9999), nilstr, false},
		{Shape(0), nilstr, false},
	}

	for _, tt := range tests {
		exp := tt.json
		bs, err := tt.s.MarshalJSON()

		if err != nil && tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected success '%v', got error %v", tt.s, exp, err)

			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("'%v'.MarshalJSON(): expected error, got success", tt.s)

			continue
		}

		act := string(bs)
		if exp != act {
			t.Errorf("'%v'.MarshalJSON(): expected '%v', actual '%v'", tt.s, exp, act)
		}
	}
}

func TestShapeUnmarshalJSON(t *testing.T) {
	t.Parallel()

	var zero Shape
	tests := []struct {
		s         Shape
		json      string
		succeeded bool
	}{
		{Scalar, "\"scalar\"", true},
		{Band, "\"band\"", true},
		{Heatmap, "\"heatmap\"", true},
		{Polyline, "\"polyline\"", true},
		{zero, "\"unknown\"", false},
		{zero, "\"foobar\"", false},
	}

	for _, tt := range tests {
		exp := tt.s
		bs := []byte(tt.json)

		var act Shape

		err := act.UnmarshalJSON(bs)
		if err != nil && tt.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected success '%v', got error %v", tt.json, exp, err)

			continue
		}

		if err == nil && !tt.succeeded {
			t.Errorf("UnmarshalJSON('%v'): expected error, got success", tt.json)

			continue
		}

		if exp != act {
			t.Errorf("UnmarshalJSON('%v'): expected '%v', actual '%v'", tt.json, exp, act)
		}
	}
}
