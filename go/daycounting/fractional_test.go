package daycounting

import (
	"testing"
	"time"

	"portf_py/daycounting/conventions"
)

const (
	secondsInLeapYear    = 31622400
	secondsInNonLeapYear = 31536000
)

func TestYearFracRAW(t *testing.T) {
	t.Run("leap year", func(t *testing.T) {
		dt1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		dt2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		result, err := YearFrac(dt1, dt2, conventions.RAW)
		if err != nil {
			t.Fatalf("YearFrac returned error: %v", err)
		}
		expected := float64(secondsInLeapYear) / float64(secondsInGregorianYear)
		if !almostEqual(result, expected, 1e-15) {
			t.Errorf("YearFrac(2020, 2021, RAW) = %v, want %v", result, expected)
		}
	})

	t.Run("non-leap year", func(t *testing.T) {
		dt1 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		dt2 := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		result, err := YearFrac(dt1, dt2, conventions.RAW)
		if err != nil {
			t.Fatalf("YearFrac returned error: %v", err)
		}
		expected := float64(secondsInNonLeapYear) / float64(secondsInGregorianYear)
		if !almostEqual(result, expected, 1e-15) {
			t.Errorf("YearFrac(2021, 2022, RAW) = %v, want %v", result, expected)
		}
	})
}

func TestYearFracInvalidMethod(t *testing.T) {
	dt1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	// Test with invalid convention value
	_, err := YearFrac(dt1, dt2, conventions.DayCountConvention(999))
	if err == nil {
		t.Error("YearFrac with invalid method should return error")
	}
}

func TestYearFracValidMethods(t *testing.T) {
	y1, m1, d1 := 2020, 1, 1
	y2, m2, d2 := 2021, 1, 1
	dt1 := time.Date(y1, time.Month(m1), d1, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(y2, time.Month(m2), d2, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		method   conventions.DayCountConvention
		expected float64
	}{
		{"THIRTY_360_US", conventions.THIRTY_360_US, US30360(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"THIRTY_360_US_EOM", conventions.THIRTY_360_US_EOM, US30360Eom(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"THIRTY_360_US_NASD", conventions.THIRTY_360_US_NASD, US30360Nasd(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"THIRTY_360_EU", conventions.THIRTY_360_EU, Eur30360(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"THIRTY_360_EU_M2", conventions.THIRTY_360_EU_M2, Eur30360Model2(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"THIRTY_360_EU_M3", conventions.THIRTY_360_EU_M3, Eur30360Model3(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"THIRTY_360_EU_PLUS", conventions.THIRTY_360_EU_PLUS, Eur30360Plus(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"THIRTY_365", conventions.THIRTY_365, Thirty365(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"ACT_360", conventions.ACT_360, Act360(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"ACT_365_FIXED", conventions.ACT_365_FIXED, Act365Fixed(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"ACT_365_NONLEAP", conventions.ACT_365_NONLEAP, Act365Nonleap(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"ACT_ACT_EXCEL", conventions.ACT_ACT_EXCEL, ActActExcel(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"ACT_ACT_ISDA", conventions.ACT_ACT_ISDA, ActActIsda(y1, m1, d1, y2, m2, d2, 0, 0, false)},
		{"ACT_ACT_AFB", conventions.ACT_ACT_AFB, ActActAfb(y1, m1, d1, y2, m2, d2, 0, 0, false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := YearFrac(dt1, dt2, tt.method)
			if err != nil {
				t.Fatalf("YearFrac returned error: %v", err)
			}
			if !almostEqual(result, tt.expected, 1e-15) {
				t.Errorf("YearFrac with %s = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestDayFracRAW(t *testing.T) {
	dt1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	result, err := DayFrac(dt1, dt2, conventions.RAW)
	if err != nil {
		t.Fatalf("DayFrac returned error: %v", err)
	}

	expected := 1.0 // One day
	if !almostEqual(result, expected, 1e-10) {
		t.Errorf("DayFrac(1 day, RAW) = %v, want %v", result, expected)
	}
}

func TestDayFracValidMethods(t *testing.T) {
	y1, m1, d1 := 2020, 1, 1
	y2, m2, d2 := 2020, 2, 1
	dt1 := time.Date(y1, time.Month(m1), d1, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(y2, time.Month(m2), d2, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		method   conventions.DayCountConvention
		expected float64
	}{
		{"THIRTY_360_US", conventions.THIRTY_360_US, US30360(y1, m1, d1, y2, m2, d2, 0, 0, true)},
		{"THIRTY_360_EU", conventions.THIRTY_360_EU, Eur30360(y1, m1, d1, y2, m2, d2, 0, 0, true)},
		{"ACT_360", conventions.ACT_360, Act360(y1, m1, d1, y2, m2, d2, 0, 0, true)},
		{"ACT_365_FIXED", conventions.ACT_365_FIXED, Act365Fixed(y1, m1, d1, y2, m2, d2, 0, 0, true)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DayFrac(dt1, dt2, tt.method)
			if err != nil {
				t.Fatalf("DayFrac returned error: %v", err)
			}
			if !almostEqual(result, tt.expected, 1e-15) {
				t.Errorf("DayFrac with %s = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestFracSwappedDates(t *testing.T) {
	// Test that Frac handles swapped dates correctly
	dt1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	result1, err1 := YearFrac(dt1, dt2, conventions.ACT_365_FIXED)
	result2, err2 := YearFrac(dt2, dt1, conventions.ACT_365_FIXED)

	if err1 != nil || err2 != nil {
		t.Fatalf("YearFrac returned error: %v, %v", err1, err2)
	}

	if !almostEqual(result1, result2, epsilon) {
		t.Errorf("YearFrac with swapped dates should give same result: %v != %v", result1, result2)
	}
}

func TestFracWithIntraDayTimes(t *testing.T) {
	// Test with specific times of day
	dt1 := time.Date(2020, 1, 1, 9, 30, 0, 0, time.UTC)  // 9:30 AM
	dt2 := time.Date(2020, 1, 1, 15, 45, 0, 0, time.UTC) // 3:45 PM

	result, err := YearFrac(dt1, dt2, conventions.RAW)
	if err != nil {
		t.Fatalf("YearFrac returned error: %v", err)
	}

	// Should be fraction of a day
	if result >= 1.0 || result <= 0.0 {
		t.Errorf("YearFrac for intraday should be between 0 and 1, got %v", result)
	}
}

func TestDayFracEur30360(t *testing.T) {
	// Test specific example
	dt1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)

	result, err := DayFrac(dt1, dt2, conventions.THIRTY_360_EU)
	if err != nil {
		t.Fatalf("DayFrac returned error: %v", err)
	}

	expected := 30.0 // 30 days in 30/360 convention
	if !almostEqual(result, expected, epsilon) {
		t.Errorf("DayFrac Eur30360 (Jan to Feb) = %v, want %v", result, expected)
	}
}

func TestYearFracEur30360(t *testing.T) {
	// Test specific example
	dt1 := time.Date(2018, 12, 15, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2019, 3, 1, 0, 0, 0, 0, time.UTC)

	result, err := YearFrac(dt1, dt2, conventions.THIRTY_360_EU)
	if err != nil {
		t.Fatalf("YearFrac returned error: %v", err)
	}

	expected := 0.21111111
	if !almostEqual(result, expected, 1e-8) {
		t.Errorf("YearFrac Eur30360 = %v, want %v", result, expected)
	}
}

func TestActualMethods(t *testing.T) {
	// Compare actual day count methods for a leap year
	dt1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	act360, _ := YearFrac(dt1, dt2, conventions.ACT_360)
	act365, _ := YearFrac(dt1, dt2, conventions.ACT_365_FIXED)
	actActExcel, _ := YearFrac(dt1, dt2, conventions.ACT_ACT_EXCEL)
	actActIsda, _ := YearFrac(dt1, dt2, conventions.ACT_ACT_ISDA)

	// Act/360 should be larger than Act/365 (same numerator, smaller denominator)
	if act360 <= act365 {
		t.Errorf("Act/360 (%v) should be > Act/365 (%v)", act360, act365)
	}

	// ActAct methods should be ~1.0 for full year
	if !almostEqual(actActExcel, 1.0, 1e-10) {
		t.Errorf("ActActExcel for full year = %v, want ~1.0", actActExcel)
	}

	if !almostEqual(actActIsda, 1.0, 1e-10) {
		t.Errorf("ActActIsda for full year = %v, want ~1.0", actActIsda)
	}
}

func BenchmarkYearFrac(b *testing.B) {
	dt1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	b.Run("RAW", func(b *testing.B) {
		for b.Loop() {
			_, _ = YearFrac(dt1, dt2, conventions.RAW)
		}
	})

	b.Run("ACT_365_FIXED", func(b *testing.B) {
		for b.Loop() {
			_, _ = YearFrac(dt1, dt2, conventions.ACT_365_FIXED)
		}
	})

	b.Run("THIRTY_360_EU", func(b *testing.B) {
		for b.Loop() {
			_, _ = YearFrac(dt1, dt2, conventions.THIRTY_360_EU)
		}
	})

	b.Run("ACT_ACT_EXCEL", func(b *testing.B) {
		for b.Loop() {
			_, _ = YearFrac(dt1, dt2, conventions.ACT_ACT_EXCEL)
		}
	})
}
