# Day Count Conventions

See the [Wikipedia: Day Count Convention](https://en.wikipedia.org/wiki/Day_count_convention) for the introductionary overview.

The most common standartization documents are:

- [ISDA 2006 Definitions (Section 4.16)](https://web.archive.org/web/20140913145444/http://www.hsbcnet.com/gbm/attachments/standalone/2006-isda-definitions.pdf)
- [ISO 20022 (Field 22F Day Count Convention codes)](https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm).

The ISO 20022 codes corresponnd to the following conventions:
 
- A001: 30/360 US
- A002: 30/365
- A004: Actual/360
- A005: Actual/365 Fixed
- A008: Actual/Actual ISDA
- A010: Actual/Actual AFB
- A011: 30/360 European
- A012: 30E2/360
- A013: 30E3/360
- A014: Actual/365 Non-Leap

Excel implements the [YEARFRAC function](https://support.microsoft.com/en-us/office/yearfrac-function-3844141e-c76d-4143-82b6-208454ddc6a8) which has five basis modes equivalent to:

- Basis 0: has no exact equivalent, the closest is 30/360 US EOM
- Basis 1: Actual/Actual Excel-compatible
- Basis 2: Actual/360
- Basis 3: Actual/365 Fixed
- Basis 4: 30/360 European

## Convention summary table

| Category | Name, Synonyms | ISO 20022 Code | Excel Basis Equivalent | EOM Adjustment | Feb 29 Handling | Denominator Basis |
|----------|----------------|----------------|------------------------|-----------|----------------|-----------------|
| Simple | Raw seconds / Gregorian year | (none) | (none) | None | Leap year irrelevant except denominator is fixed | 31556952 seconds |
| 30/360 | 30/360 US, 30U/360, 30/360 ISDA, 30/360 Bond Basis, 30/360 U.S. Municipal, American Basic Rule | A001 | Basis 0 variant | Conditional (31st→30th) | Feb 28/29 treated as 28/29 depending on start | 360 |
| 30/360 | 30/360 US EOM | (none) | Closest to basis 0 | End-of-month logic incl. Feb | Feb 28/29→30 if rule triggered | 360 |
| 30/360 | 30/360 US NASD | (none) | Basis 0 variant | 31st adjustments, special final-day mapping | Feb 28 not forced unless EOM interplay | 360 |
| 30/360 | 30/360 EU, 30/360 ICMA, 30/360 Eurobond Basis, ISDA-2006 | A011 | Basis 4 | 31st→30th universally | Feb 28/29 unaffected unless >=28 rule | 360 |
| 30/360 | 30E2/360, Eurobond basis model 2 | A012 | (none) | Complex Feb leap adjustments | Feb 28/29 conditional matching prior date | 360 |
| 30/360 | 30E3/360, Eurobond basis model 3 | A013 | (none) | Feb >=28 → 30 | Feb >=28 forced to 30 | 360 |
| 30/360 | 30E+/360, 30/360 EU Plus | (none) | (none) | 31→32 for end date (shifts forward) | Feb rules follow base logic | 360 |
| Similar to US 30/360 but /365 | 30/365 | A002 | (none) | Hybrid | Feb treatment like US variant | 365 |
| Actual | Actual/360, Act/360, A/360, French | A004 | Basis 2 | None | Pure actual days | 360 |
| Actual | Actual/365 Fixed, Act/365 Fixed, A/365 Fixed, A/365F, English | A005 | Basis 3 | None | Pure actual days | 365 fixed |
| Actual | Actual/365 Non-Leap, Actual/365NL | A014 | (none) | None | Subtract each Feb 29 encountered | 365 (after removing leap days) |
| Actual/Actual | Actual/Actual Excel-compatible | (none) | Basis 1 | Special year boundary logic | Feb 29 may alter denominator (366) | Variable (365 or 366 or average) |
| Actual/Actual | Actual/Actual ISDA, Act/365 ISDA, Actual/365 ISDA, Act/365 ISDA | A008 | (none) | Split periods | Leap year segments each with 365 or 366 | Piecewise sum |
| Actual/Actual | Actual/Actual AFB, Actual/Actual FBF | A010 | (none) | Split with month rules | Leap year only if Feb 29 within specific range | Piecewise sum |

## Leap year handling summary

| Method | Leap Year Influence |
|--------|---------------------|
| RAW | Only via actual seconds; denominator constant |
| 30/360 family | Mostly through special Feb 28/29 rules; denominator fixed 360 (except 30/365) |
| Actual/360 | Adds day; denominator fixed 360 |
| Actual/365 Fixed | Adds day to numerator; denominator fixed 365 |
| Actual/365 Non-Leap | Removes Feb 29 from numerator |
| Actual/Actual Excel-compatible | Chooses 365 vs 366 or averages across span |
| Actual/Actual ISDA | Segment denominators per year (365/366) |
| Actual/Actual AFB | Conditional leap year inclusion based on month boundaries |

## Choosing a convention

| Scenario | Recommended convention | Rationale |
|----------|------------------------|-----------|
| US corporate bond accrual | 30/360 US | Market standard synthetic month method |
| Eurobond legacy instrument | 30/360 EU | ICMA / Excel basis 4 compatibility |
| Excel reconciliation (Actual/Actual) | Actual/Actual Excel-compatible | Matches spreadsheet outputs |
| Money market (short-term) | Actual/360 | Common discount/interest basis |
| Long-term swap leg (fixed) | Actual/365 Fixed | Stable denominator; widely adopted |
| Need to exclude leap distortion | Actual/365 Non-Leap | Avoids Feb 29 numerator inflation |
| ISDA-defined derivative accrual | Actual/Actual ISDA | Adheres to legal definitions |
| French banking legacy | Actual/Actual AFB | Regional compliance |
| Intraday rate analytics | RAW | High-resolution physical time support |

## Accuracy and edge cases

- End-of-month transitions: 30/360 variants can produce unintuitive equal accrual for periods ending on 30th vs 31st.
- February spans: Differences between EU models (M2/M3/Plus) are most pronounced for intervals bracketing Feb 28/29 and leap years.
- Leap year spanning for Actual/Actual can diverge materially between Excel vs ISDA vs AFB.
- NASD and EU+ models using artificial day 32 shift can change expected stub calculations—ensure downstream systems accept such normalization.
- Intraday fractions: For very small intervals (seconds), Actual conventions are preferable over synthetic 30/360.

## Intraday fractions

All methods accept `df1`, `df2` representing fractional day portion:

```text
df = (hour*3600 + minute*60 + second) / 86400
```

Fraction is added to numerator as `+ df2 - df1`. This preserves ordering and supports time-of-day precision.

## RAW

The use cases are: intraday analytics, performance measurement, pricing where market conventions are not required. This method ignores leap year variability in numerator (only affects actual seconds) and is simpler than Actual/365. It is precise for short intervals but not acceptable for coupon accrual standards.

The computation logic is:

1. Compute `diffSeconds = (t2 - t1)`.
2. If day fraction: divide by 86400 (seconds in a day).
3. Else: divide by 31556952 (seconds in average Gregorian year).

## 30/360 Family

Typical use cases are bond accrual and particularly fixed coupon instruments.

All 30/360 variants construct a synthetic day difference:

```text
BaseDays = 360*(y2 - y1) + 30*(m2 - m1)
Adjustment = (AdjustedD2 - AdjustedD1)
TotalDays = BaseDays + Adjustment + (df2 - df1)
YearFraction = TotalDays / 360 (except 30/365 which divides by 365)
```

They differ by rules transforming `d1` and `d2` (and sometimes February edge cases).

### 30/360 US (ISO 20022 code A001)

See [Github: 30U/360 daycount method](https://github.com/hcnn/d30360u).
Typically used in US corporate and municipal bonds.

- Adjust `d1` → min(d1, 30).
- Adjust `d2` → if `d2 == 31` and `d1 >= 30`, then 30; else keep.
- Feb end: If start is 30/31, then a 31st becomes 30. Feb end not forcibly changed unless interacting with start rule.
- Difference vs European is that European always converts 31→30; US is conditional.

### 30/360 US EOM

See [Github: 30U/360 EOM daycount method](https://github.com/hcnn/d30360m).

- Implements explicit end-of-month recognition including February.
- Rules combine conditions labeled `rule2` (start in Feb with day ≥28), `rule3` (end also Feb day ≥28), `rule4` (end at 31 with start ≥30).
- Results closer (but not identical) to Excel basis 0.

### 30/360 US NASD

See [Github: 30U/360 NASD daycount method](https://github.com/hcnn/d30360n).

- If `d2 == 31`: if `d1 < 30` → set `d2` to 32 (push forward one); else → 30.
- `d1` truncated to 30 if >30.
- Produces behavior similar to Excel's handling (basis 0) with modified treatment of month ends.

The key distinction of this method is the unique use of 32 to simulate next-day shift for certain start-day scenarios.

### 30/360 EU (ISO 20022 code A011)

See [Github: 30E/360 daycount method](https://github.com/hcnn/d30360s).
Used in Eurobonds (pre-Euro era), ICMA standard settlement accrual.

- Universal truncation: any `d1 > 30` →30; any `d2 > 30` →30.
- No leap-year-dependent branching.
- Simpler, symmetric adjustments.

The difference vs US is deterministic 31→30 for both start and end regardless of start day.

### 30/360 EU M2 (ISO 20022 code A012)

See [Github: 30E2/360 daycount method](https://github.com/hcnn/d30360e2).
Might be used in specialized contracts needing fidelity in `Feb` transitions.

- Adds leap-year aware handling when February is the end month.
- If end is Feb 28/29 in leap year and start day is 29/30≥, adjust `d2` to 29 or 30 accordingly.
- Purpose: More nuanced Feb normalization.

### 30/360 EU M3 (ISO 20022 code A013)

See [Github: 30E3/360 daycount method](https://github.com/hcnn/d30360e3).
Used in iche instruments; provides uniformity across leap/non-leap years.

- Forces Feb days ≥28 to become 30 on both start and end.
- Symmetric aggressive normalization for February.

### 30/360 EU+

See [Github: 30E+/360 daycount method](https://github.com/hcnn/d30360p).
Used in structured products requiring forward projection for coupon stub calculation.

- Unique rule: if `d2 == 31` then `d2Adj = 32` (forward shift), start day truncated at 30.
- Differs from NASD by lacking conditional dependence on `d1 < 30` for the 32 shift.

### 30/365 (ISO 20022 code A002)

See [Github: 30/365 daycount method](https://github.com/hcnn/d30365).
Used by some UK markets or bespoke OTC contracts.

- Same adjustment pattern as 30/360 US but final division by 365.
- Introduced for instruments referencing a 365-day accrual assumption while maintaining synthetic 30-day months.
- Hybrid between synthetic month approach and 365 denominator, reducing annual accrual vs /360 basis.

## Actual-Day Methods

All Actual methods use real day counts derived from Julian Day. Common formula elements are:

- `ActualDays = JD(date2) - JD(date1)` (or `date2.Sub(date1).Hours()/24` for Excel-compatible path).
- Add `(df2 - df1)` for intraday fractions.
- Denominator depends on method (fixed 360/365, conditional 365/366, or piecewise sums).

### Actual/360 (ISO 20022 code A004)

See [Github: Actual/360 daycount method](https://github.com/hcnn/act360).
Used in money market instruments (e.g., commercial paper, short-term deposits) especially USD/EUR.

- YearFraction = ActualDays / 360.
- No leap-year special case; leap day just adds 1 to numerator.
- Difference vs Actual/365 Fixed is the larger fraction for same period (since denominator is smaller).

### Actual/365 Fixed (ISO 20022 code A005)

See [Github: Actual/365 Fixed daycount method](https://github.com/hcnn/act365f).
Used in  UK markets, some derivatives legs.

- YearFraction = ActualDays / 365.
- Feb 29 included only in numerator; denominator unchanged.
- Difference vs Actual/365 Non-Leap: non-Leap adjusts numerator by removing leap days.

### Actual/365 Non-Leap (ISO 20022 code A014)

See [Github: Actual/365 Non-Leap daycount method](https://github.com/hcnn/act365n).
Used in instruments needing year length stability excluding leap distortion (some legacy UK loan agreements).

- Counts actual days then subtracts each Feb 29 encountered between start and end.
- Denominator stays 365.
- Produces lower accrual than Actual/365 Fixed over spans including Feb 29.

Algorithm details:

 1. `diffDays = JD2 - JD1 + df2 - df1`
 2. Count leap years where Feb 29 is fully inside interval.
 3. `diffDays -= leapYears`.
 4. Divide by 365 (or return days if `dayFrac`).

### Actual/Actual Excel-compatible

Used for interoperability with spreadsheet models and valuation spreadsheets.

- Two regimes:
  - If dates appear within ≤1 year (same year or y1+1==y2 with (m1 > m2 or same month and d1 ≥ d2)): choose denominator 365 or 366 based on leap occurrence.
  - Else: Use average days per year across the span (total days between Jan 1 of first year and Jan 1 of year after last year, divided by number of years spanned).
- Uses `time.Sub` in hours for fidelity.
- Difference vs ISDA/AFB: Excel chooses single or averaged denominator; ISDA/AFB split periods by year.

### Actual/Actual ISDA (ISO 20022 code A008)

See [Github: Actual/Actual ISDA daycount method](https://github.com/hcnn/act_isda).
Used in bonds, swaps referencing ISDA standards (e.g., certain fixed-floating accrual calculations).

- Splits interval into at most three segments: remainder of first year, whole intervening years, part of final year.
- Year-by-year denominators: 365 or 366 depending on leap year.
- YearFraction = Sum( segmentDays / segmentYearLength ).
- Difference vs Excel: piecewise vs averaged denominator.

Algorithm steps:

 1. If in same year: direct actual days / (365 or 366).
 2. Else compute `diffA` (days from start date to Dec 31 inclusive), `diffB` (days from Jan 1 to end date), plus whole intervening years.
 3. Sum with appropriate denominators.

### Actual/Actual AFB (ISO 20022 code A010)

See [Github: Actual/Actual AFB daycount method](https://github.com/hcnn/act_afb).
Used by French and some continental European instruments.

- Similar segmentation to ISDA but leap year treatment differs: leap day only counts toward denominator if located before March 1 for the starting year and after Feb for ending year.
- Provides region-specific French banking treatment.
- Difference vs ISDA: conditioned leap-year inclusion reduces effect of leap day placement.

## Worked numeric examples

The following tables show computed `YearFrac` and `DayFrac` values for three illustrative date ranges across all implemented conventions. These were generated by running `go run ./examples/compare` inside the package directory. Use them to understand relative magnitudes and leap year effects.

### Same-year (leap year) interval

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

### Cross-leap-year interval

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
- ACT/365 Non-Leap subtracts the leap day (364 numerator vs 365 fixed denominator).
- ACT/ACT Excel vs ISDA/AFB show subtle denominator segmentation differences (<0.0005 spread).

### Multi-year interval (leap year inside)

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

### Interpreting differences

- Larger fractions usually indicate a smaller denominator (e.g., ACT/360) or inclusion of leap day.
- When comparing yields or accrued interest across instruments, ensure matching conventions to avoid systematic bias.
- For reporting, prefer Actual methods unless a contractual 30/360 specification applies.

## References

1. [Wikipedia: Day Count Convention](https://en.wikipedia.org/wiki/Day_count_convention)
2. [ISO 20022 (Field 22F Day Count Convention codes)](https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
3. [ISDA 2006 Definitions (Section 4.16)](https://web.archive.org/web/20140913145444/http://www.hsbcnet.com/gbm/attachments/standalone/2006-isda-definitions.pdf)
4. [Excel YEARFRAC function](https://support.microsoft.com/en-us/office/yearfrac-function-3844141e-c76d-4143-82b6-208454ddc6a8)
5. [Github: 30E/360 daycount method](https://github.com/hcnn/d30360s)
6. [Github: 30E2/360 daycount method](https://github.com/hcnn/d30360e2)
7. [Github: 30E3/360 daycount method](https://github.com/hcnn/d30360e3)
8. [Github: 30E+/360 daycount method](https://github.com/hcnn/d30360p)
9. [Github: 30U/360 daycount method](https://github.com/hcnn/d30360u)
10. [Github: 30U/360 EOM daycount method](https://github.com/hcnn/d30360m)
11. [Github: 30U/360 NASD daycount method](https://github.com/hcnn/d30360n)
12. [Github: 30/365 daycount method](https://github.com/hcnn/d30365)
13. [Github: Actual/365 Non-Leap daycount method](https://github.com/hcnn/act365n)
14. [Github: Actual/365 Fixed daycount method](https://github.com/hcnn/act365f)
15. [Github: Actual/360 daycount method](https://github.com/hcnn/act360)
16. [Github: Actual/Actual ISDA daycount method](https://github.com/hcnn/act_isda)
17. [Github: Actual/Actual AFB daycount method](https://github.com/hcnn/act_afb)
18. [Github: YEARFRAC for Rust](https://github.com/AnatolyBuga/yearfrac)
19. [Github: YEARFRAC](https://github.com/devind-team/devind_yearfrac)
