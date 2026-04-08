package main

import (
	"fmt"
	"time"

	"zpano/daycounting"
	"zpano/daycounting/conventions"
)

// datePair holds two dates and a label
type datePair struct {
	label string
	start time.Time
	end   time.Time
}

func main() {
	pairs := []datePair{
		{label: "SameYear (Leap)", start: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), end: time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)},
		{label: "CrossLeapYear", start: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), end: time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)},
		{label: "MultiYear", start: time.Date(2023, 6, 30, 0, 0, 0, 0, time.UTC), end: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)},
	}

	convs := []struct {
		name string
		c    conventions.DayCountConvention
	}{
		{"RAW", conventions.RAW},
		{"30/360 US", conventions.THIRTY_360_US},
		{"30/360 US EOM", conventions.THIRTY_360_US_EOM},
		{"30/360 US NASD", conventions.THIRTY_360_US_NASD},
		{"30/360 EU", conventions.THIRTY_360_EU},
		{"30E2/360", conventions.THIRTY_360_EU_M2},
		{"30E3/360", conventions.THIRTY_360_EU_M3},
		{"30E+/360", conventions.THIRTY_360_EU_PLUS},
		{"30/365", conventions.THIRTY_365},
		{"ACT/360", conventions.ACT_360},
		{"ACT/365 Fixed", conventions.ACT_365_FIXED},
		{"ACT/365 NonLeap", conventions.ACT_365_NONLEAP},
		{"ACT/ACT Excel", conventions.ACT_ACT_EXCEL},
		{"ACT/ACT ISDA", conventions.ACT_ACT_ISDA},
		{"ACT/ACT AFB", conventions.ACT_ACT_AFB},
	}

	fmt.Printf("Day Count Comparison (Year Fractions)\n")
	fmt.Printf("Generated: %s\n\n", time.Now().Format(time.RFC3339))

	for _, p := range pairs {
		fmt.Printf("== %s: %s -> %s ==\n", p.label, p.start.Format("2006-01-02"), p.end.Format("2006-01-02"))
		fmt.Printf("%-18s %-16s %-16s\n", "Convention", "YearFrac", "DayFrac")
		fmt.Printf("%-18s %-16s %-16s\n", "---------", "--------", "-------")
		for _, cv := range convs {
			yf, err := daycounting.YearFrac(p.start, p.end, cv.c)
			if err != nil {
				fmt.Printf("%-18s ERROR            ERROR\n", cv.name)
				continue
			}
			df, err := daycounting.DayFrac(p.start, p.end, cv.c)
			if err != nil {
				fmt.Printf("%-18s %-16.12f ERROR\n", cv.name, yf)
				continue
			}
			fmt.Printf("%-18s %-16.12f %-16.8f\n", cv.name, yf, df)
		}
		fmt.Println()
	}

	fmt.Println("Notes:")
	fmt.Println("- ACT/365 NonLeap subtracts leap days from numerator.")
	fmt.Println("- RAW uses actual seconds divided by 31,556,952 (Gregorian average year).")
	fmt.Println("- 30/360 variants construct synthetic months; differences lie in end-date adjustments.")
	fmt.Println("- ACT/ACT variants differ in segmentation and denominators.")
}
