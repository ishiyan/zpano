// Package conventions provides day count convention definitions and utilities.
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
package conventions
