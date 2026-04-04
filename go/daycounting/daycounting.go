package daycounting

import (
	"time"
)

// Wikipedia
// https://en.wikipedia.org/wiki/Day_count_convention
//
// ISDA 2006 Definitions, Section 4.16 page 11
// https://web.archive.org/web/20140913145444/http://www.hsbcnet.com/gbm/attachments/standalone/2006-isda-definitions.pdf
//
// For Excel YEARFRAC function see
// https://support.microsoft.com/en-us/office/yearfrac-function-3844141e-c76d-4143-82b6-208454ddc6a8
//
// Excel YEARFRAC function:
// Basis Optional: The type of day count basis to use.
// 0: US (NASD) 30/360 (default is not set)
// 1: Actual/actual
// 2: Actual/360
// 3: Actual/365
// 4: European 30/360
//
// Day counting methods are listed in the ISO 20022, see
// https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
//
// Source code
// https://github.com/devind-team/devind_yearfrac
// https://github.com/hcnn/d30360s
// https://github.com/hcnn/d30360e2
// https://github.com/hcnn/d30360e3
// https://github.com/hcnn/d30360p
// https://github.com/hcnn/d30360u
// https://github.com/hcnn/d30360m
// https://github.com/hcnn/d30360n
// https://github.com/hcnn/d30365
// https://github.com/hcnn/act365n
// https://github.com/hcnn/act365f
// https://github.com/hcnn/act360
// https://github.com/hcnn/act_isda
// https://github.com/hcnn/act_afb
// https://github.com/AnatolyBuga/yearfrac

// IsLeapYear returns true if the given year is a leap year.
func IsLeapYear(y int) bool {
	return y%4 == 0 && (y%100 != 0 || y%400 == 0)
}

// DateToJD converts a date to Julian Day number.
//
// Algorithm adapted from
// Press, W. H., Teukolsky, S. A., Vetterling, W. T., & Flannery, B. P. (2007).
// Numerical Recipes: The Art of Scientific Computing (3rd ed.). Cambridge University Press.
func DateToJD(year, month, day int) int {
	a := (14 - month) / 12
	y := year + 4800 - a
	m := month + (12 * a) - 3

	jd := day + (153*m+2)/5 + y*365
	jd += y/4 - y/100 + y/400 - 32045
	return jd
}

// JDToDate converts a Julian Day number to a date (year, month, day).
//
// Algorithm adapted from
// Press, W. H., Teukolsky, S. A., Vetterling, W. T., & Flannery, B. P. (2007).
// Numerical Recipes: The Art of Scientific Computing (3rd ed.). Cambridge University Press.
func JDToDate(jd int) (year, month, day int) {
	a := jd + 32044
	b := (4*a + 3) / 146097
	c := a - (b*146097)/4

	d := (4*c + 3) / 1461
	e := c - (d*1461)/4
	m := (5*e + 2) / 153
	m2 := m / 10

	day = e + 1 - (153*m+2)/5
	month = m + 3 - 12*m2
	year = b*100 + d - 4800 + m2

	return
}

// Eur30360 calculates the day count fraction using the 30/360 European method.
//
// Source:
//
//	https://github.com/hcnn/d30360s
//
// Synonyms:
//   - 30/360 ICMA
//   - 30/360 Eurobond Basis
//   - ISDA-2006
//   - 30S/360 Special German
//
// ISO 20022:
//
//		A011
//		https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
//
//	Method whereby interest is calculated based on a 30-day month
//	and a 360-day year.
//
//	Accrued interest to a value date on the last day of a month
//	shall be the same as to the 30th calendar day of the same month,
//	except for February.
//
//	This means that a 31st is assumed to be a 30th and the 28 Feb
//	(or 29 Feb for a leap year) is assumed to be a 28th (or 29th).
//
//	It is the most commonly used 30/360 method for non-US straight
//	and convertible bonds issued before 01/01/1999.
func Eur30360(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	diffDays := float64(360*(y2-y1)+30*(m2-m1)) + df2 - df1

	d2Adj := min(d2, 30)
	d1Adj := min(d1, 30)

	diffDays += float64(d2Adj - d1Adj)

	if fracDays {
		return diffDays
	}
	return diffDays / 360
}

// Eur30360Model2 calculates the day count fraction using the 30E2/360 method.
//
// Source:
//
//	https://github.com/hcnn/d30360e2
//
// Synonyms:
//   - 30E2/360
//   - Eurobond basis model 2
//
// ISO 20022:
//
//	A012
//	https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
//
// Method whereby interest is calculated based on a 30-day month and
// a 360-day year.
//
// Accrued interest to a value date on the last day of a month shall
// be the same as to the 30th calendar day of the same month, except
// for the last day of February whose day of the month value shall
// be adapted to the value of the first day of the interest period
// if the latter is higher and if the period is one of a regular
// schedule.
//
// This means that a 31st is assumed to be a 30th and the 28th Feb
// of a non-leap year is assumed to be equivalent to a 29th Feb
// when the first day of the interest period is a 29th, or to a 30th
// Feb when the first day of the interest period is a 30th or a 31st.
//
// The 29th Feb of a leap year is assumed to be equivalent to a 30th
// Feb when the first day of the interest period is a 30th or a 31st.
//
// Similarly, if the coupon period starts on the last day of February,
// it is assumed to produce only one day of interest in February as if
// it was starting on a 30th Feb when the end of the period is a 30th
// or a 31st, or two days of interest in February when the end of the
// period is a 29th, or 3 days of interest in February when it is the
// 28th Feb of a non-leap year and the end of the period is before the
// 29th.
func Eur30360Model2(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	diffDays := float64(360*(y2-y1)+30*(m2-m1)) + df2 - df1
	leap1 := IsLeapYear(y1)
	d2Adj := d2

	if leap1 && m2 == 2 && d2 == 28 {
		if d1 == 29 {
			d2Adj = 29
		} else if d1 >= 30 {
			d2Adj = 30
		}
	} else if leap1 && m2 == 2 && d2 == 29 {
		if d1 >= 30 {
			d2Adj = 30
		}
	} else if d2 > 30 {
		d2Adj = 30
	}

	d1Adj := min(d1, 30)
	diffDays += float64(d2Adj - d1Adj)

	if fracDays {
		return diffDays
	}
	return diffDays / 360
}

// Eur30360Model3 calculates the day count fraction using the 30E3/360 method.
//
// Source:
//
//	https://github.com/hcnn/d30360e3
//
// Synonyms:
//   - 30E3/360
//   - Eurobond basis model 3
//
// ISO 20022:
//
//	A013
//	https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
//
// Method whereby interest is calculated based on a 30-day month
// and a 360-day year.
//
// Accrued interest to a value date on the last day of a month
// shall be the same as to the 30th calendar day of the same month.
//
// This means that a 31st is assumed to be a 30th and the 28 Feb
// (or 29 Feb for a leap year) is assumed to be equivalent to a
// 30 Feb.
//
// It is a variation of the 30E/360 (or Eurobond basis) method
// where the last day of February is always assumed to be a 30th,
// even if it is the last day of the maturity coupon period.
func Eur30360Model3(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	diffDays := float64(360*(y2-y1)+30*(m2-m1)) + df2 - df1

	d2Adj := d2
	if m2 == 2 && d2 >= 28 {
		d2Adj = 30
	} else if d2 > 30 {
		d2Adj = 30
	}

	d1Adj := d1
	if m1 == 2 && d1 >= 28 {
		d1Adj = 30
	} else if d1 > 30 {
		d1Adj = 30
	}

	diffDays += float64(d2Adj - d1Adj)

	if fracDays {
		return diffDays
	}
	return diffDays / 360
}

// Eur30360Plus calculates the day count fraction using the 30E+/360 method.
//
// Source:
//
//	https://github.com/hcnn/d30360p
//
// Synonyms:
//   - 30E+/360
func Eur30360Plus(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	diffDays := float64(360*(y2-y1)+30*(m2-m1)) + df2 - df1

	d2Adj := d2
	if d2 == 31 {
		d2Adj = 32
	}

	d1Adj := d1
	if d1 > 30 {
		d1Adj = 30
	}

	diffDays += float64(d2Adj - d1Adj)

	if fracDays {
		return diffDays
	}
	return diffDays / 360
}

// US30360 calculates the day count fraction using the 30/360 US method.
//
// Source:
//
//	https://github.com/hcnn/d30360u
//
// Synonyms:
//   - 30/360 ISDA
//   - 30U/360
//   - 30/360 US
//   - 30/360 Bond Basis
//   - 30/360 U.S. Municipal
//   - American Basic Rule
//
// ISO 20022:
//
//	A001
//	https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
//
// Method whereby interest is calculated based on a 30-day month
// and a 360-day year.
//
// Accrued interest to a value date on the last day of a month shall
// be the same as to the 30th calendar day of the same month, except
// for February, and provided that the interest period started on a
// 30th or a 31st.
//
// This means that a 31st is assumed to be a 30th if the period started
// on a 30th or a 31st and the 28 Feb (or 29 Feb for a leap year) is
// assumed to be a 28th (or 29th).
//
// It is the most commonly used 30/360 method for US straight and
// convertible bonds.
func US30360(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	diffDays := float64(360*(y2-y1)+30*(m2-m1)) + df2 - df1

	d2Adj := d2
	if d2 == 31 && d1 >= 30 {
		d2Adj = 30
	}

	d1Adj := min(d1, 30)
	diffDays += float64(d2Adj - d1Adj)

	if fracDays {
		return diffDays
	}
	return diffDays / 360
}

// US30360Eom calculates the day count fraction using the 30/360 US EOM method.
//
// Source:
//
//	https://github.com/hcnn/d30360m
//
// Synonyms:
//   - 30/360 US EOM
func US30360Eom(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	diffDays := float64(360*(y2-y1)+30*(m2-m1)) + df2 - df1

	rule2 := m1 == 2 && d1 >= 28
	rule3 := rule2 && m2 == 2 && d2 >= 28
	rule4 := d2 == 31 && d1 >= 30

	d1Adj := d1
	if rule2 {
		d1Adj = 30
	} else if d1 > 30 {
		d1Adj = 30
	}

	d2Adj := d2
	if rule4 || rule3 {
		d2Adj = 30
	}

	diffDays += float64(d2Adj - d1Adj)

	if fracDays {
		return diffDays
	}
	return diffDays / 360
}

// US30360Nasd calculates the day count fraction using the 30/360 NASD method.
//
// Source:
//
//	https://github.com/hcnn/d30360n
//
// Synonyms:
//   - 30/360 NASD
func US30360Nasd(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	diffDays := float64(360*(y2-y1)+30*(m2-m1)) + df2 - df1

	d2Adj := d2
	if d2 == 31 {
		if d1 < 30 {
			d2Adj = 32
		} else {
			d2Adj = 30
		}
	}

	d1Adj := min(d1, 30)
	diffDays += float64(d2Adj - d1Adj)

	if fracDays {
		return diffDays
	}
	return diffDays / 360
}

// Thirty365 calculates the day count fraction using the 30/365 method.
//
// Source:
//
//	https://github.com/hcnn/d30365
//
// Synonyms:
//   - 30/365
//
// ISO 20022:
//
//	A002
//	https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
//
// Method whereby interest is calculated based on a 30-day month
// in a way similar to the 30/360 (basic rule) and a 365-day year.
//
// Accrued interest to a value date on the last day of a month shall
// be the same as to the 30th calendar day of the same month, except
// for February.
//
// This means that a 31st is assumed to be a 30th and the 28 Feb (or
// 29 Feb for a leap year) is assumed to be a 28th (or 29th).
func Thirty365(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	diffDays := float64(360*(y2-y1)+30*(m2-m1)) + df2 - df1

	d2Adj := d2
	if d2 == 31 && d1 >= 30 {
		d2Adj = 30
	}

	d1Adj := min(d1, 30)
	diffDays += float64(d2Adj - d1Adj)

	if fracDays {
		return diffDays
	}
	return diffDays / 365
}

// Act365Nonleap calculates the day count fraction using the Actual/365 Non-Leap method.
//
// Source:
//
//	https://github.com/hcnn/act365n
//
// Synonyms:
//   - Actual/365NL
//   - Actual/365 Non-Leap
//
// ISO 20022:
//
//	A014
//	https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
//
// Method whereby interest is calculated based on the actual
// number of accrued days in the interest period, excluding
// any leap day from the count, and a 365-day year.
func Act365Nonleap(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	diffDays := float64(DateToJD(y2, m2, d2)-DateToJD(y1, m1, d1)) + df2 - df1

	leapYears := 0
	if IsLeapYear(y1) && m1 <= 2 {
		leapYears++
	}
	if y1 != y2 && IsLeapYear(y2) && m2 >= 3 {
		leapYears++
	}
	if y1+1 < y2 {
		for now := y1 + 1; now < y2; now++ {
			if IsLeapYear(now) {
				leapYears++
			}
		}
	}

	diffDays -= float64(leapYears)

	if fracDays {
		return diffDays
	}
	return diffDays / 365
}

// Act365Fixed calculates the day count fraction using the Actual/365 Fixed method.
//
// Source:
//
//	https://github.com/hcnn/act365f
//
// Synonyms:
//   - Actual/365 Fixed
//   - Act/365 Fixed
//   - A/365 Fixed
//   - A/365F
//   - English
//
// ISO 20022:
//
//	A005
//	https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
//
// Method whereby interest is calculated based on the actual
// number of accrued days in the interest period and a 365-day year.
func Act365Fixed(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	diffDays := float64(DateToJD(y2, m2, d2)-DateToJD(y1, m1, d1)) + df2 - df1

	if fracDays {
		return diffDays
	}
	return diffDays / 365
}

// Act360 calculates the day count fraction using the Actual/360 method.
//
// Source:
//
//	https://github.com/hcnn/act360
//
// Synonyms:
//   - Actual/360
//   - Act/360
//   - A/360
//   - French
//
// ISO 20022:
//
//	A004
//	https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
//
// Method whereby interest is calculated based on the actual
// number of accrued days in the interest period and a 360-day year.
func Act360(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	diffDays := float64(DateToJD(y2, m2, d2)-DateToJD(y1, m1, d1)) + df2 - df1

	if fracDays {
		return diffDays
	}
	return diffDays / 360
}

func feb29Between(date1, date2 time.Time, y1, y2 int) bool {
	// Check each year in the range
	for y := y1; y <= y2; y++ {
		if IsLeapYear(y) {
			leapDay := time.Date(y, time.February, 29, 0, 0, 0, 0, time.UTC)
			if !date1.After(leapDay) && !leapDay.After(date2) {
				return true
			}
		}
	}
	return false
}

func appearsLeYear(y1, m1, d1, y2, m2, d2 int) bool {
	// Returns true if date1 and date2 "appear" to be 1 year or less apart.
	// This compares the values of year, month, and day directly to each other.
	// Requires date1 <= date2; returns boolean. Used by basis 1.
	if y1 == y2 {
		return true
	}
	if y1+1 == y2 && (m1 > m2 || (m1 == m2 && d1 >= d2)) {
		return true
	}
	return false
}

// ActActExcel calculates the day count fraction using Excel's Actual/Actual (basis 1) method.
//
// Cannot find it in ISO 20022.
//
// Found it on github (https://github.com/AnatolyBuga/yearfrac)
// and verified it with Excel.
//
// Other actual/actual methods from ISO 20022 produce
// different figures compared to Excel.
func ActActExcel(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	date1 := time.Date(y1, time.Month(m1), d1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(y2, time.Month(m2), d2, 0, 0, 0, 0, time.UTC)

	if appearsLeYear(y1, m1, d1, y2, m2, d2) {
		var yearDays float64
		if y1 == y2 && IsLeapYear(y1) {
			yearDays = 366 // leap year
		} else if feb29Between(date1, date2, y1, y2) || (m2 == 2 && d2 == 29) {
			yearDays = 366 // leap year feb29
		} else {
			yearDays = 365 // leap year else
		}
		df := date2.Sub(date1).Hours() / 24
		if fracDays {
			return df + df2 - df1
		}
		return (df + df2 - df1) / yearDays
	} else {
		yearStart1 := time.Date(y1, time.January, 1, 0, 0, 0, 0, time.UTC)
		yearStart2 := time.Date(y2+1, time.January, 1, 0, 0, 0, 0, time.UTC)
		yearDays := yearStart2.Sub(yearStart1).Hours() / 24
		avgYearDays := yearDays / float64(y2-y1+1)
		df := date2.Sub(date1).Hours() / 24
		if fracDays {
			return df + df2 - df1
		}
		return (df + df2 - df1) / avgYearDays
	}
}

// ActActIsda calculates the day count fraction using the Actual/Actual ISDA method.
//
// Source:
//
//	https://github.com/hcnn/act_isda
//
// Synonyms:
//   - Actual/Actual ISDA
//   - Act/Act ISDA
//   - Actual/365 ISDA
//   - Act/365 ISDA
//
// ISO 20022:
//
//	A008
//	https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
//
// Method whereby interest is calculated based on the actual number
// of accrued days of the interest period that fall on a normal year,
// divided by 365, added to the actual number of days of the interest
// period that fall on a leap year, divided by 366.
func ActActIsda(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	if y1 == y2 {
		denom := 365.0
		if IsLeapYear(y2) {
			denom = 366.0
		}
		diffDays := float64(DateToJD(y2, m2, d2)-DateToJD(y1, m1, d1)) + df2 - df1
		if fracDays {
			return diffDays
		}
		return diffDays / denom
	}

	denomA := 365.0
	if IsLeapYear(y1) {
		denomA = 366.0
	}
	diffA := float64(DateToJD(y1, 12, 31) - DateToJD(y1, m1, d1) + 1)

	denomB := 365.0
	if IsLeapYear(y2) {
		denomB = 366.0
	}
	diffB := float64(DateToJD(y2, m2, d2) - DateToJD(y2, 1, 1))

	if fracDays {
		diff := diffA - df1 + diffB + df2
		for year := y1 + 1; year < y2; year++ {
			if IsLeapYear(year) {
				diff += 366
			} else {
				diff += 365
			}
		}
		return diff
	}

	return (diffA-df1)/denomA + (diffB+df2)/denomB + float64(y2-y1-1)
}

// ActActAfb calculates the day count fraction using the Actual/Actual AFB method.
//
// Source:
//
//	https://github.com/hcnn/act_afb
//
// Synonyms:
//   - Actual/Actual AFB
//   - Actual/Actual FBF
//
// ISO 20022:
//
//	A010
//	https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm
//
// Method whereby interest is calculated based on the actual
// number of accrued days and a 366-day year (if 29 Feb falls
// in the coupon period) or a 365-day year (if 29 Feb does not
// fall in the coupon period).
//
// If a coupon period is longer than one year, it is split by
// repetitively separating full year sub-periods counting backwards
// from the end of the coupon period (a year backwards from a 28 Feb
// being 29 Feb, if it exists).
//
// The first of the sub-periods starts on the start date of the
// accrued interest period and thus is possibly shorter than a year.
//
// Then the interest computation is operated separately on each
// sub-period and the intermediate results are summed up.
func ActActAfb(y1, m1, d1, y2, m2, d2 int, df1, df2 float64, fracDays bool) float64 {
	if y1 == y2 {
		denom := 365.0
		if m1 < 3 && IsLeapYear(y1) {
			denom = 366.0
		}
		diffDays := float64(DateToJD(y2, m2, d2)-DateToJD(y1, m1, d1)) + df2 - df1
		if fracDays {
			return diffDays
		}
		return diffDays / denom
	}

	denomA := 365.0
	if m1 < 3 && IsLeapYear(y1) {
		denomA = 366.0
	}
	diffA := float64(DateToJD(y1, 12, 31) - DateToJD(y1, m1, d1) + 1)

	denomB := 365.0
	if m2 >= 3 && IsLeapYear(y2) {
		denomB = 366.0
	}
	diffB := float64(DateToJD(y2, m2, d2) - DateToJD(y2, 1, 1))

	if fracDays {
		diff := diffA - df1 + diffB + df2
		for year := y1 + 1; year < y2; year++ {
			if IsLeapYear(year) {
				diff += 366
			} else {
				diff += 365
			}
		}
		return diff
	}

	return (diffA-df1)/denomA + (diffB+df2)/denomB + float64(y2-y1-1)
}
