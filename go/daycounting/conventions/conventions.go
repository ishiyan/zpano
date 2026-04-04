package conventions

import (
	"fmt"
	"strings"
)

// DayCountConvention represents different day count conventions used in financial calculations.
type DayCountConvention int

const (
	// RAW takes the difference in seconds between two dates and divides
	// it by the number of seconds in a Gregorian year (31556952).
	//
	// This is what has most sense for intraday periods or when
	// we are not concerned with the calculation of the interest
	// accrual between coupon payment dates.
	//
	// Strings: 'raw'
	RAW DayCountConvention = iota

	// THIRTY_360_US is 30/360 (ISDA) or 30/360 (American Basic Rule)
	//
	// This is NOT the same as the "US (NASD) 30/360" (basis 0)
	// in Excel YEARFRAC function.
	//
	// Use THIRTY_360_US_EOM for the closest match.
	//
	// Coded as A001 in ISO 20022
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Strings: '30/360 us', '30u/360'
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
	THIRTY_360_US

	// THIRTY_360_US_EOM is 30/360 US End-Of-Month
	//
	// This is NOT the same as the "US (NASD) 30/360" (basis 0)
	// in Excel YEARFRAC function.
	//
	// Although the results are not completely the same,
	// this is the closest match.
	//
	// This method is not listed in ISO 20022.
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Found it on github (https://github.com/hcnn/d30360m)
	//
	// Strings: '30/360 us eom', '30u/360 eom'
	THIRTY_360_US_EOM

	// THIRTY_360_US_NASD is 30/360 NASD
	//
	// This is NOT the same as the "US (NASD) 30/360" (basis 0)
	// in Excel YEARFRAC function.
	//
	// Use THIRTY_360_US_EOM for the closest match.
	//
	// This method is not listed in ISO 20022.
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Found it on github (https://github.com/hcnn/d30360n)
	//
	// Strings: '30/360 us nasd', '30u/360 nasd'
	THIRTY_360_US_NASD

	// THIRTY_360_EU is 30/360 Eurobond Basis or 30/360 ICMA
	//
	// This is the same as the "Eur 30/360" (basis 4)
	// in Excel YEARFRAC function.
	//
	// Coded as A011 in ISO 20022
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Strings: '30/360 eu', '30e/360'
	//
	// Method whereby interest is calculated based on a 30-day month
	// and a 360-day year.
	//
	// Accrued interest to a value date on the last day of a month
	// shall be the same as to the 30th calendar day of the same month,
	// except for February.
	//
	// This means that a 31st is assumed to be a 30th and the 28 Feb
	// (or 29 Feb for a leap year) is assumed to be a 28th (or 29th).
	//
	// It is the most commonly used 30/360 method for non-US straight
	// and convertible bonds issued before 01/01/1999.
	THIRTY_360_EU

	// THIRTY_360_EU_M2 is 30E2/360 or Eurobond basis model 2
	//
	// This is NOT the same as the "Eur 30/360" (basis 4)
	// in Excel YEARFRAC function.
	//
	// Use THIRTY_360_EU if you want excel-compatible results.
	//
	// Coded as A012 in ISO 20022
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Strings: '30/360 eu2', '30e2/360'
	THIRTY_360_EU_M2

	// THIRTY_360_EU_M3 is 30E3/360 or Eurobond basis model 3
	//
	// This is NOT the same as the "Eur 30/360" (basis 4)
	// in Excel YEARFRAC function.
	//
	// Use THIRTY_360_EU if you want excel-compatible results.
	//
	// Coded as A013 in ISO 20022
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Strings: '30/360 eu3', '30e3/360'
	THIRTY_360_EU_M3

	// THIRTY_360_EU_PLUS is 30E+/360
	//
	// This is NOT the same as the "Eur 30/360" (basis 4)
	// in Excel YEARFRAC function.
	//
	// Use THIRTY_360_EU if you want excel-compatible results.
	//
	// This method is not listed in ISO 20022.
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Found it on github (https://github.com/hcnn/d30360p)
	//
	// Strings: '30/360 eu+', '30e+/360'
	THIRTY_360_EU_PLUS

	// THIRTY_365 is 30/365
	//
	// There is no related basis in Excel YEARFRAC function.
	//
	// Coded as A002 in ISO 20022
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Strings: '30/365'
	THIRTY_365

	// ACT_360 is Actual/360
	//
	// This is the same as the "Actual/360" (basis 2)
	// in Excel YEARFRAC function.
	//
	// Coded as A004 in ISO 20022
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Strings: 'act/360'
	ACT_360

	// ACT_365_FIXED is Actual/365 Fixed
	//
	// This is the same as the "Actual/365" (basis 3)
	// in Excel YEARFRAC function.
	//
	// Coded as A005 in ISO 20022
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Strings: 'act/365 fixed'
	ACT_365_FIXED

	// ACT_365_NONLEAP is Actual/365 Non-Leap
	//
	// This is NOT the same as the "Actual/365" (basis 3)
	// in Excel YEARFRAC function.
	//
	// Use ACT_365_FIXED if you want excel-compatible results.
	//
	// Coded as A014 in ISO 20022
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Strings: 'act/365 nonleap'
	ACT_365_NONLEAP

	// ACT_ACT_EXCEL is Excel-compatible Actual/Actual (basis 1) method.
	//
	// This method is not listed in ISO 20022.
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Other actual/actual methods from ISO 20022 produce
	// different figures compared to Excel.
	//
	// Found it on github (https://github.com/AnatolyBuga/yearfrac)
	// and verified it with Excel.
	//
	// Strings: 'act/act excel'
	ACT_ACT_EXCEL

	// ACT_ACT_ISDA is Actual/Actual ISDA or Actual/365 ISDA
	//
	// This is NOT the same as the "Actual/Actual" (basis 1)
	// in Excel YEARFRAC function.
	//
	// Use ACT_ACT_EXCEL if you want excel-compatible results.
	//
	// Coded as A008 in ISO 20022
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Strings: 'act/act isda', 'act/365 isda'
	ACT_ACT_ISDA

	// ACT_ACT_AFB is Actual/Actual AFB
	//
	// This is NOT the same as the "Actual/Actual" (basis 1)
	// in Excel YEARFRAC function.
	//
	// Use ACT_ACT_EXCEL if you want excel-compatible results.
	//
	// Coded as A010 in ISO 20022
	// (https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
	//
	// Strings: 'act/act afb'
	ACT_ACT_AFB
)

var conventionMap = map[string]DayCountConvention{
	"raw":             RAW,
	"30/360 us":       THIRTY_360_US,
	"30/360 us eom":   THIRTY_360_US_EOM,
	"30/360 us nasd":  THIRTY_360_US_NASD,
	"30/360 eu":       THIRTY_360_EU,
	"30/360 eu2":      THIRTY_360_EU_M2,
	"30/360 eu3":      THIRTY_360_EU_M3,
	"30/360 eu+":      THIRTY_360_EU_PLUS,
	"30/365":          THIRTY_365,
	"act/360":         ACT_360,
	"act/365 fixed":   ACT_365_FIXED,
	"act/365 nonleap": ACT_365_NONLEAP,
	"act/act excel":   ACT_ACT_EXCEL,
	"act/act isda":    ACT_ACT_ISDA,
	"act/act afb":     ACT_ACT_AFB,
}

// FromString converts a string representation to a DayCountConvention.
// The comparison is case-insensitive.
// Returns an error if the string does not match any known convention.
func FromString(convention string) (DayCountConvention, error) {
	normalized := strings.ToLower(convention)

	if conv, ok := conventionMap[normalized]; ok {
		return conv, nil
	}

	// Build list of valid conventions for error message
	validConventions := make([]string, 0, len(conventionMap))
	for key := range conventionMap {
		validConventions = append(validConventions, key)
	}

	return RAW, fmt.Errorf("day count convention '%s' must be one of: %v", convention, validConventions)
}
