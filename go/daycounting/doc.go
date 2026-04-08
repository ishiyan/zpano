// Package daycounting provides day count calculation functions for financial applications.
//
// This package implements various day count conventions used in fixed income securities,
// derivatives, and other financial instruments to calculate accrued interest, year fractions,
// and day fractions between two dates.
//
// # Key Features
//
//   - Multiple day count conventions (30/360, Actual/360, Actual/365, Actual/Actual variants)
//   - Julian Date conversion utilities for precise date calculations
//   - Leap year detection and handling
//   - Year fraction and day fraction calculations
//   - Support for intraday time fractions
//   - Excel-compatible calculation methods
//   - ISO 20022 compliant conventions
//
// ```text
// go/daycounting/
// ├── conventions/
// │   ├── conventions.go      # Day count convention definitions
// │   └── conventions_test.go # Tests for conventions
// ├── daycounting.go          # Core day counting functions
// ├── daycounting_test.go     # Tests for day counting functions
// ├── fractional.go           # High-level fraction calculation API
// └── fractional_test.go      # Tests for fractional functions
// ```
//
// ## 1. Core Calculation Interface
//
// The functions exposed by this package:
//
//   - `Frac(date1, date2, convention, dayFrac)` – returns either day fraction (if `dayFrac=true`) or year fraction.
//   - `YearFrac(date1, date2, convention)` – year fraction wrapper.
//   - `DayFrac(date1, date2, convention)` – day fraction wrapper.
//
// Time-of-day components are converted into fractional days (`df1`, `df2`) and added/subtracted inside each algorithm so all methods support intraday precision.
//
// Julian Day conversion (`DateToJD`, `JDToDate`) is used for exact day differences in Actual-based methods.
//
// Leap year detection via `IsLeapYear(year)` influences denominators (365 vs 366) or date adjustments.
//
// # Day Count Conventions
//
// The package supports the following convention families:
//
//   - 30/360 variants (US, European, NASD, with various adjustments)
//   - Actual/360 (French method)
//   - Actual/365 (Fixed and Non-Leap variants)
//   - Actual/Actual (Excel, ISDA, AFB variants)
//   - 30/365 method
//   - RAW method (for intraday calculations)
//
// # Julian Date Support
//
// Julian Date (JD) functions enable precise date arithmetic:
//
//   - DateToJD: Converts calendar dates to Julian Day numbers
//   - JDToDate: Converts Julian Day numbers back to calendar dates
//
// Julian Dates provide a continuous count of days since a reference point,
// facilitating accurate day count calculations across different conventions.
//
// # Leap Year Handling
//
// The package includes leap year detection (IsLeapYear) which is critical for:
//
//   - Accurate day counting in Actual/Actual methods
//   - Proper handling of February 29th in various conventions
//   - Year fraction denominator adjustments (365 vs 366 days)
//
// # Year and Day Fractions
//
// Core calculation functions:
//
//   - Frac: General-purpose fraction calculator with convention selection
//   - YearFrac: Returns time difference as a fraction of a year
//   - DayFrac: Returns time difference as a number of days (with fractional component)
//
// These functions support intraday precision by accounting for hours, minutes, and seconds.
//
// # Usage Example
//
//	import (
//	    "time"
//	    "zpano/daycounting"
//	    "zpano/daycounting/conventions"
//	)
//
//	date1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
//	date2 := time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)
//
//	// Calculate year fraction using Actual/360 convention
//	yearFrac, _ := daycounting.YearFrac(date1, date2, conventions.ACT_360)
//
//	// Calculate day fraction using 30/360 US convention
//	dayFrac, _ := daycounting.DayFrac(date1, date2, conventions.THIRTY_360_US)
//
// # Standards Compliance
//
// The package implements methods documented in:
//
//   - ISO 20022 Day Count Convention codes
//   - ISDA 2006 Definitions (Section 4.16)
//   - Excel YEARFRAC function compatibility
//   - Various market-specific conventions (NASD, ICMA, AFB)
//
// For detailed information on specific conventions, see the conventions subpackage.
package daycounting
