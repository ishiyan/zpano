# Day Count Conventions in `daycounting`

This document describes all day count conventions implemented in the Go package `daycounting`. It explains:

- What each convention is used for in financial markets
- How each algorithm works (step-by-step adjustments)
- Differences between similar conventions
- ISO 20022 / Excel YEARFRAC mapping where applicable
- Treatment of leap years, February endpoints, and end-of-month (EOM) behavior
- Intraday (fraction-of-day) handling

Sources referenced (summarized, not reproduced):

- ISDA 2006 Definitions (Section 4.16)
- ISO 20022 (Field 22F Day Count Convention codes)
- Wikipedia: Day count convention
- Excel YEARFRAC documentation
- Linked open-source implementations (see comments inside `fractional.go`)

> NOTE: Explanations below are descriptive and paraphrased for clarity; no proprietary text is copied verbatim.

## 1. Core Calculation Interface

The functions exposed by this package:

- `Frac(date1, date2, convention, dayFrac)` – returns either day fraction (if `dayFrac=true`) or year fraction.
- `YearFrac(date1, date2, convention)` – year fraction wrapper.
- `DayFrac(date1, date2, convention)` – day fraction wrapper.

Time-of-day components are converted into fractional days (`df1`, `df2`) and added/subtracted inside each algorithm so all methods support intraday precision.

Julian Day conversion (`DateToJD`, `JDToDate`) is used for exact day differences in Actual-based methods.

Leap year detection via `IsLeapYear(year)` influences denominators (365 vs 366) or date adjustments.

## 2. Convention Summary Table

| Identifier | Name / Synonyms | ISO 20022 Code | Excel Basis Equivalent | Category | EOM Adjustment | Feb 29 Handling | Denominator Basis |
|------------|-----------------|----------------|------------------------|----------|----------------|-----------------|------------------|
| `RAW` | Raw (seconds / Gregorian year) | (none) | (none) | Simple | None | Leap year irrelevant except denominator is fixed | 31556952 seconds |
| `THIRTY_360_US` | 30/360 US, 30U/360, Bond Basis | A001 | (None - not basis 0) | 30/360 | Conditional (31st→30th) | Feb 28/29 treated as 28/29 depending on start | 360 |
| `THIRTY_360_US_EOM` | 30/360 US EOM | (none) | Closest to Excel basis 0 | 30/360 | End-of-month logic incl. Feb | Feb 28/29→30 if rule triggered | 360 |
| `THIRTY_360_US_NASD` | 30/360 NASD | (none) | Excel basis 0 variant | 30/360 | 31st adjustments, special final-day mapping | Feb 28 not forced unless EOM interplay | 360 |
| `THIRTY_360_EU` | 30/360 Eurobond/ICMA | A011 | Excel basis 4 | 30/360 | 31st→30th universally | Feb 28/29 unaffected unless >=28 rule | 360 |
| `THIRTY_360_EU_M2` | 30E2/360 | A012 | (None) | 30/360 | Complex Feb leap adjustments | Feb 28/29 conditional matching prior date | 360 |
| `THIRTY_360_EU_M3` | 30E3/360 | A013 | (None) | 30/360 | Feb >=28 → 30 | Feb >=28 forced to 30 | 360 |
| `THIRTY_360_EU_PLUS` | 30E+/360 | (none) | (None) | 30/360 | 31→32 for end date (shifts forward) | Feb rules follow base logic | 360 |
| `THIRTY_365` | 30/365 | A002 | (None) | Hybrid | Similar to US 30/360 but /365 | Feb treatment like US variant | 365 |
| `ACT_360` | Actual/360 | A004 | Excel basis 2 | Actual | None | Pure actual days | 360 |
| `ACT_365_FIXED` | Actual/365 Fixed | A005 | Excel basis 3 | Actual | None | Pure actual days | 365 fixed |
| `ACT_365_NONLEAP` | Actual/365 Non-Leap, Actual/365NL | A014 | (None) | Actual | None | Subtract each Feb 29 encountered | 365 (after removing leap days) |
| `ACT_ACT_EXCEL` | Actual/Actual Excel-compatible | (none) | Excel basis 1 | Actual/Actual | Special year boundary logic | Feb 29 may alter denominator (366) | Variable (365 or 366 or average) |
| `ACT_ACT_ISDA` | Actual/Actual ISDA, Act/365 ISDA | A008 | (None) | Actual/Actual | Split periods | Leap year segments each with 365 or 366 | Piecewise sum |
| `ACT_ACT_AFB` | Actual/Actual AFB | A010 | (None) | Actual/Actual | Split with month rules | Leap year only if Feb 29 within specific range | Piecewise sum |

## 3. Detailed Algorithm Explanations

### 3.1 RAW
**Use Case:** Intraday analytics, performance measurement, pricing where market conventions are not required. 
**Logic:**
1. Compute `diffSeconds = (t2 - t1)`.
2. If day fraction: divide by 86400.
3. Else: divide by 31556952 (average Gregorian year length).
**Differences:** Ignores leap year variability in numerator (only affects actual seconds). Simpler than Actual/365.
**Pros/Cons:** Precise for short intervals; not acceptable for coupon accrual standards.

### 3.2 30/360 Family (General Pattern)
All 30/360 variants construct a synthetic day difference:
```
BaseDays = 360*(y2 - y1) + 30*(m2 - m1)
Adjustment = (AdjustedD2 - AdjustedD1)
TotalDays = BaseDays + Adjustment + (df2 - df1)
YearFraction = TotalDays / 360 (except 30/365 which divides by 365)
```
They differ by rules transforming `d1` and `d2` (and sometimes February edge cases).
Use case: Bond accrual, particularly fixed coupon instruments.

#### 3.2.1 THIRTY_360_US (A001)
- Adjust `d1` → min(d1, 30).
- Adjust `d2` → if `d2 == 31` and `d1 >= 30`, then 30; else keep.
- Feb end: If start is 30/31, then a 31st becomes 30. Feb end not forcibly changed unless interacting with start rule.
**Usage:** US corporate and municipal bonds.
**Difference vs European:** European always converts 31→30; US is conditional.

#### 3.2.2 THIRTY_360_US_EOM
- Implements explicit end-of-month recognition including February.
- Rules combine conditions labeled `rule2` (start in Feb with day ≥28), `rule3` (end also Feb day ≥28), `rule4` (end at 31 with start ≥30).
- Results closer (but not identical) to Excel basis 0.
**Usage:** Approximation when Excel compatibility desired but pure NASD logic differs from requirements.

#### 3.2.3 THIRTY_360_US_NASD
- If `d2 == 31`: if `d1 < 30` → set `d2` to 32 (push forward one); else → 30.
- `d1` truncated to 30 if >30.
- Produces behavior similar to Excel's handling (basis 0) with modified treatment of month ends.
**Key Distinction:** Unique use of 32 to simulate next-day shift for certain start-day scenarios.

#### 3.2.4 THIRTY_360_EU (A011)
- Universal truncation: any `d1 > 30` →30; any `d2 > 30` →30.
- No leap-year-dependent branching.
- Simpler, symmetric adjustments.
**Usage:** Eurobonds (pre-Euro era), ICMA standard settlement accrual.
**Difference vs US:** Deterministic 31→30 for both start and end regardless of start day.

#### 3.2.5 THIRTY_360_EU_M2 (A012)
- Adds leap-year aware handling when February is the end month.
- If end is Feb 28/29 in leap year and start day is 29/30≥, adjust `d2` to 29 or 30 accordingly.
- Purpose: More nuanced Feb normalization.
**Usage:** Specialized contracts needing fidelity in Feb transitions.

#### 3.2.6 THIRTY_360_EU_M3 (A013)
- Forces Feb days ≥28 to become 30 on both start and end.
- Symmetric aggressive normalization for February.
**Usage:** Niche instruments; provides uniformity across leap/non-leap years.

#### 3.2.7 THIRTY_360_EU_PLUS
- Unique rule: if `d2 == 31` then `d2Adj = 32` (forward shift), start day truncated at 30.
- Differs from NASD by lacking conditional dependence on `d1 < 30` for the 32 shift.
**Usage:** Structured products requiring forward projection for coupon stub calculation.

#### 3.2.8 THIRTY_365 (A002)
- Same adjustment pattern as THIRTY_360_US but final division by 365.
- Introduced for instruments referencing a 365-day accrual assumption while maintaining synthetic 30-day months.
**Usage:** Some UK markets or bespoke OTC contracts.
**Difference:** Hybrid between synthetic month approach and 365 denominator, reducing annual accrual vs /360 basis.

### 3.3 Actual-Day Methods
All Actual methods use real day counts derived from Julian Day or `time.Sub` difference.

#### Common Formula Elements:
- `ActualDays = JD(date2) - JD(date1)` (or `date2.Sub(date1).Hours()/24` for Excel-compatible path).
- Add `(df2 - df1)` for intraday fractions.
- Denominator depends on method (fixed 360/365, conditional 365/366, or piecewise sums).

#### 3.3.1 ACT_360 (A004)
- YearFraction = ActualDays / 360.
- No leap-year special case; leap day just adds 1 to numerator.
**Usage:** Money market instruments (e.g., commercial paper, short-term deposits) especially USD/EUR.
**Difference vs ACT_365_FIXED:** Larger fraction for same period (since denominator is smaller).

#### 3.3.2 ACT_365_FIXED (A005)
- YearFraction = ActualDays / 365.
- Feb 29 included only in numerator; denominator unchanged.
**Usage:** UK markets, some derivatives legs.
**Difference vs ACT_365_NONLEAP:** Non-Leap adjusts numerator by removing leap days.

#### 3.3.3 ACT_365_NONLEAP (A014)
- Counts actual days then subtracts each Feb 29 encountered between start and end.
- Denominator stays 365.
**Algorithm Details:**
 1. `diffDays = JD2 - JD1 + df2 - df1`
 2. Count leap years where Feb 29 is fully inside interval.
 3. `diffDays -= leapYears`.
 4. Divide by 365 (or return days if `dayFrac`).
**Usage:** Instruments needing year length stability excluding leap distortion (some legacy UK loan agreements).
**Difference:** Produces lower accrual than ACT_365_FIXED over spans including Feb 29.

#### 3.3.4 ACT_ACT_EXCEL
- Two regimes:
  - If dates appear within ≤1 year (same year or y1+1==y2 with (m1 > m2 or same month and d1 ≥ d2)): choose denominator 365 or 366 based on leap occurrence.
  - Else: Use average days per year across the span (total days between Jan 1 of first year and Jan 1 of year after last year, divided by number of years spanned).
- Uses `time.Sub` in hours for fidelity.
**Usage:** Interoperability with spreadsheet models; valuation spreadsheets.
**Difference vs ISDA/AFB:** Excel chooses single or averaged denominator; ISDA/AFB split periods by year.

#### 3.3.5 ACT_ACT_ISDA (A008)
- Splits interval into at most three segments: remainder of first year, whole intervening years, part of final year.
- Year-by-year denominators: 365 or 366 depending on leap year.
- YearFraction = Sum( segmentDays / segmentYearLength ).
**Algorithm Steps:**
 1. If in same year: direct actual days / (365 or 366).
 2. Else compute `diffA` (days from start date to Dec 31 inclusive), `diffB` (days from Jan 1 to end date), plus whole intervening years.
 3. Sum with appropriate denominators.
**Usage:** Bonds, swaps referencing ISDA standards (e.g., certain fixed-floating accrual calculations).
**Difference vs Excel:** Piecewise vs averaged denominator.

#### 3.3.6 ACT_ACT_AFB (A010)
- Similar segmentation to ISDA but leap year treatment differs: leap day only counts toward denominator if located before March 1 for the starting year and after Feb for ending year.
- Provides region-specific French banking treatment.
**Usage:** French and some continental European instruments.
**Difference vs ISDA:** Conditioned leap-year inclusion reduces effect of leap day placement.

### 3.4 Intraday Fractions
All methods accept `df1`, `df2` representing fractional day portion:
```
df = (hour*3600 + minute*60 + second) / 86400
```
Fraction added to numerator as `+ df2 - df1`. This preserves ordering and supports time-of-day precision in interest accrual or rate compounding.

### 3.5 Julian Day Support
Used primarily in Actual methods for integral day calculations immune to DST/timezone complications. Ensures stable day counts independent of local offsets.

### 3.6 Leap Year Handling Summary
| Method | Leap Year Influence |
|--------|---------------------|
| RAW | Only via actual seconds; denominator constant |
| 30/360 family | Mostly through special Feb 28/29 rules; denominator fixed 360 (except 30/365) |
| ACT_360 | Adds day; denominator fixed 360 |
| ACT_365_FIXED | Adds day to numerator; denominator fixed 365 |
| ACT_365_NONLEAP | Removes Feb 29 from numerator |
| ACT_ACT_EXCEL | Chooses 365 vs 366 or averages across span |
| ACT_ACT_ISDA | Segment denominators per year (365/366) |
| ACT_ACT_AFB | Conditional leap year inclusion based on month boundaries |

## 4. Choosing a Convention
| Scenario | Recommended Convention | Rationale |
|----------|-----------------------|-----------|
| US corporate bond accrual | `THIRTY_360_US` | Market standard synthetic month method |
| Eurobond legacy instrument | `THIRTY_360_EU` | ICMA / Excel basis 4 compatibility |
| Excel reconciliation (Actual/Actual) | `ACT_ACT_EXCEL` | Matches spreadsheet outputs |
| Money market (short-term) | `ACT_360` | Common discount/interest basis |
| Long-term swap leg (fixed) | `ACT_365_FIXED` | Stable denominator; widely adopted |
| Need to exclude leap distortion | `ACT_365_NONLEAP` | Avoids Feb 29 numerator inflation |
| ISDA-defined derivative accrual | `ACT_ACT_ISDA` | Adheres to legal definitions |
| French banking legacy | `ACT_ACT_AFB` | Regional compliance |
| Intraday rate analytics | `RAW` | High-resolution time support |

## 5. Accuracy and Edge Cases
- End-of-month transitions: 30/360 variants can produce unintuitive equal accrual for periods ending on 30th vs 31st.
- February spans: Differences between EU models (M2/M3/Plus) are most pronounced for intervals bracketing Feb 28/29 and leap years.
- Leap year spanning for Actual/Actual can diverge materially between Excel vs ISDA vs AFB.
- NASD and EU+ models using artificial day 32 shift can change expected stub calculations—ensure downstream systems accept such normalization.
- Intraday fractions: For very small intervals (seconds), Actual conventions are preferable over synthetic 30/360.

## 6. Performance Considerations
Algorithms are O(1) except leap year counting loops (in `Act365Nonleap` and Actual/Actual methods) which are O(Y) where Y is number of intervening years (typically negligible). Julian conversions are arithmetic-only (no allocations).

## 7. Validation and Cross-Checking
Recommended cross-check strategy:
1. Compare `ACT_ACT_EXCEL` output with spreadsheet `YEARFRAC(...,1)`.
2. Compare `THIRTY_360_EU` with `YEARFRAC(...,4)`.
3. Benchmark `ACT_360` vs simple `(ActualDays / 360)` manual calculation.
4. Verify leap year spanning examples (including Feb 29) across Actual/Actual variants to ensure expected denominator differences.

## 8. Future Extensions
Potential additions:
- ACT/ACT ICMA variant (coupon-based period splitting).
- Business day adjustment wrappers.
- Stubs and schedule generation utilities.
- Support for time zones other than UTC with explicit offset normalization.

## 9. Disclaimer
This document provides educational guidance. For contractual or legal settlement calculations, always refer to the official documentation (ISDA Definitions, ICMA handbooks, prospectuses) and validate outputs against authoritative sources.

## 10. References (Informational)
- ISDA 2006 Definitions (Section 4.16) – Day Count Fraction definitions (summarized)
- ISO 20022: Field 22F codes for day count conventions
- Excel: YEARFRAC function reference
- Open-source implementations cited in source code comments
- Wikipedia: Day count convention overview

---
Generated October 25, 2025.

## 11. Worked Numeric Examples

The following tables show computed `YearFrac` and `DayFrac` values for three illustrative date ranges across all implemented conventions. These were generated by running `go run ./examples/compare` inside the module directory (`go/`). Use them to understand relative magnitudes and leap year effects.

### 11.1 Same-Year (Leap Year) Interval

Start: 2024-01-15  End: 2024-07-15 (Leap year, no Feb 29 inside span)

| Convention | YearFrac | DayFrac |
|------------|---------:|--------:|
| RAW | 0.498299075272 | 182.00000000 |
| 30/360 US | 0.500000000000 | 180.00000000 |
| 30/360 US EOM | 0.500000000000 | 180.00000000 |
| 30/360 US NASD | 0.500000000000 | 180.00000000 |
| 30/360 EU | 0.500000000000 | 180.00000000 |
| 30E2/360 | 0.500000000000 | 180.00000000 |
| 30E3/360 | 0.500000000000 | 180.00000000 |
| 30E+/360 | 0.500000000000 | 180.00000000 |
| 30/365 | 0.493150684932 | 180.00000000 |
| ACT/360 | 0.505555555556 | 182.00000000 |
| ACT/365 Fixed | 0.498630136986 | 182.00000000 |
| ACT/365 NonLeap | 0.495890410959 | 181.00000000 |
| ACT/ACT Excel | 0.497267759563 | 182.00000000 |
| ACT/ACT ISDA | 0.497267759563 | 182.00000000 |
| ACT/ACT AFB | 0.497267759563 | 182.00000000 |

Key observations:

- 30/360 family compresses actual 182 days into a standardized 180-day span (0.50 fraction).
- ACT methods reflect actual 182-day length; ACT/360 is largest due to 360 denominator.
- ACT/365 NonLeap reduces one day (181) as no leap day discount applies within span; effect stems from method design, not Feb 29 presence here.

### 11.2 Cross-Leap-Year Interval

Start: 2024-02-29  End: 2025-02-28 (Spans leap day origin year to non-leap year end)

| Convention | YearFrac | DayFrac |
|------------|---------:|--------:|
| RAW | 0.999336057551 | 365.00000000 |
| 30/360 US | 0.997222222222 | 359.00000000 |
| 30/360 US EOM | 1.000000000000 | 360.00000000 |
| 30/360 US NASD | 0.997222222222 | 359.00000000 |
| 30/360 EU | 0.997222222222 | 359.00000000 |
| 30E2/360 | 1.000000000000 | 360.00000000 |
| 30E3/360 | 1.000000000000 | 360.00000000 |
| 30E+/360 | 0.997222222222 | 359.00000000 |
| 30/365 | 0.983561643836 | 359.00000000 |
| ACT/360 | 1.013888888889 | 365.00000000 |
| ACT/365 Fixed | 1.000000000000 | 365.00000000 |
| ACT/365 NonLeap | 0.997260273973 | 364.00000000 |
| ACT/ACT Excel | 0.997267759563 | 365.00000000 |
| ACT/ACT ISDA | 0.997701923797 | 365.00000000 |
| ACT/ACT AFB | 0.997701923797 | 365.00000000 |

Key observations:

- Different 30/360 variants disagree on whether the span is treated as 359 or 360 synthetic days.
- ACT/365 NonLeap subtracts the leap day (364 numerator vs 365 fixed denominator).
- ACT/ACT Excel vs ISDA/AFB show subtle denominator segmentation differences (<0.0005 spread).

### 11.3 Multi-Year Interval (Leap Year Inside)

Start: 2023-06-30  End: 2025-03-01 (Includes all of 2024 leap day internally)

| Convention | YearFrac | DayFrac |
|------------|---------:|--------:|
| RAW | 1.670123274263 | 610.00000000 |
| 30/360 US | 1.669444444444 | 601.00000000 |
| 30/360 US EOM | 1.669444444444 | 601.00000000 |
| 30/360 US NASD | 1.669444444444 | 601.00000000 |
| 30/360 EU | 1.669444444444 | 601.00000000 |
| 30E2/360 | 1.669444444444 | 601.00000000 |
| 30E3/360 | 1.669444444444 | 601.00000000 |
| 30E+/360 | 1.669444444444 | 601.00000000 |
| 30/365 | 1.646575342466 | 601.00000000 |
| ACT/360 | 1.694444444444 | 610.00000000 |
| ACT/365 Fixed | 1.671232876712 | 610.00000000 |
| ACT/365 NonLeap | 1.668493150685 | 609.00000000 |
| ACT/ACT Excel | 1.669708029197 | 610.00000000 |
| ACT/ACT ISDA | 1.668493150685 | 610.00000000 |
| ACT/ACT AFB | 1.668493150685 | 610.00000000 |

Key observations:

- Synthetic 30/360 methods reduce 610 actual days to 601 synthetic days (~9 day compression).
- NonLeap removes the leap day (609 vs 610), while RAW, Fixed, and 360 retain it.
- Excel Actual/Actual slightly higher than ISDA/AFB due to averaging denominator vs piecewise segmentation.

### 11.4 Reproducing These Tables

Run:

```bash
cd go
go run ./examples/compare
```

(On Windows PowerShell: `Set-Location .\go; go run ./examples/compare`)

### 11.5 Interpreting Differences

- Larger fractions usually indicate a smaller denominator (e.g., ACT/360) or inclusion of leap day.
- When comparing yields or accrued interest across instruments, ensure matching conventions to avoid systematic bias.
- For reporting, prefer Actual methods unless a contractual 30/360 specification applies.

