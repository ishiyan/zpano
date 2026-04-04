package daycounting

import (
	"fmt"
	"time"

	"portf_py/daycounting/conventions"
)

const (
	secondsInGregorianYear = 31556952
	secondsInDay           = 60 * 60 * 24
)

// Frac calculates the fraction between two dates using a specified day count convention.
//
// If dayFrac is true, returns fraction in days; if false, returns fraction in years.
func Frac(dateTime1, dateTime2 time.Time, method conventions.DayCountConvention, dayFrac bool) (float64, error) {
	dt1 := dateTime1
	dt2 := dateTime2

	if dateTime1.After(dateTime2) {
		dt1, dt2 = dt2, dt1
	}

	if method == conventions.RAW {
		diffSeconds := dt2.Sub(dt1).Seconds()
		if dayFrac {
			return diffSeconds / secondsInDay, nil
		}
		return diffSeconds / secondsInGregorianYear, nil
	}

	y1 := dt1.Year()
	m1 := int(dt1.Month())
	d1 := dt1.Day()

	y2 := dt2.Year()
	m2 := int(dt2.Month())
	d2 := dt2.Day()

	// Time as a fraction of the day
	tm1 := (float64(dt1.Hour())*3600 + float64(dt1.Minute())*60 + float64(dt1.Second())) / 86400
	tm2 := (float64(dt2.Hour())*3600 + float64(dt2.Minute())*60 + float64(dt2.Second())) / 86400

	switch method {
	case conventions.THIRTY_360_US:
		return US30360(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.THIRTY_360_US_EOM:
		return US30360Eom(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.THIRTY_360_US_NASD:
		return US30360Nasd(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.THIRTY_360_EU:
		return Eur30360(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.THIRTY_360_EU_M2:
		return Eur30360Model2(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.THIRTY_360_EU_M3:
		return Eur30360Model3(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.THIRTY_360_EU_PLUS:
		return Eur30360Plus(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.THIRTY_365:
		return Thirty365(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.ACT_360:
		return Act360(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.ACT_365_FIXED:
		return Act365Fixed(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.ACT_365_NONLEAP:
		return Act365Nonleap(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.ACT_ACT_EXCEL:
		return ActActExcel(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.ACT_ACT_ISDA:
		return ActActIsda(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	case conventions.ACT_ACT_AFB:
		return ActActAfb(y1, m1, d1, y2, m2, d2, tm1, tm2, dayFrac), nil
	default:
		return 0, fmt.Errorf("unknown day count convention: %d", method)
	}
}

// YearFrac calculates the year fraction between two dates using a specified day count convention.
func YearFrac(dateTime1, dateTime2 time.Time, method conventions.DayCountConvention) (float64, error) {
	return Frac(dateTime1, dateTime2, method, false)
}

// DayFrac calculates the day fraction between two dates using a specified day count convention.
func DayFrac(dateTime1, dateTime2 time.Time, method conventions.DayCountConvention) (float64, error) {
	return Frac(dateTime1, dateTime2, method, true)
}
