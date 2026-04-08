# Day Counting Package (Go)

This package provides day count calculation functions for financial applications, converted from TypeScript to Go.

## Structure

```text
go/daycounting/
├── conventions/
│   ├── conventions.go      # Day count convention definitions
│   └── conventions_test.go # Tests for conventions
├── daycounting.go          # Core day counting functions
├── daycounting_test.go     # Tests for day counting functions
├── fractional.go           # High-level fraction calculation API
└── fractional_test.go      # Tests for fractional functions
```

## Features

### Day Count Conventions

The package supports 15 different day count conventions used in financial calculations:

- **RAW**: Simple seconds-based calculation
- **30/360 variants**: US, US EOM, US NASD, EU (Eurobond), EU Model 2, EU Model 3, EU Plus
- **30/365**: 30-day month, 365-day year
- **Actual/360**: Actual days, 360-day year
- **Actual/365**: Fixed and Non-Leap variants
- **Actual/Actual**: Excel, ISDA, and AFB variants

### Core Functions

#### Convention Parsing

```go
import "zpano/daycounting/conventions"

// Parse string to convention
conv, err := conventions.FromString("act/365 fixed")
// Returns conventions.ACT_365_FIXED
```

#### Year Fraction Calculation

```go
import (
    "zpano/daycounting"
    "zpano/daycounting/conventions"
    "time"
)

// Calculate year fraction between two dates
dt1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
dt2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

yearFrac, err := daycounting.YearFrac(dt1, dt2, conventions.ACT_365_FIXED)
// Returns ~1.002739...
```

#### Day Fraction Calculation

```go
// Calculate day fraction between two dates
dayFrac, err := daycounting.DayFrac(dt1, dt2, conventions.ACT_365_FIXED)
// Returns number of days
```

#### Low-Level Functions

```go
// Direct calculation functions
result := daycounting.Eur30360(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.US30360(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.Act360(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.ActActExcel(2020, 1, 1, 2021, 1, 1, 0, 0, false)
// ... and many more
```

## Installation

```bash
cd go
go mod download
```

## Running Tests

```bash
# Test all packages
go test ./daycounting/...

# Test conventions
go test zpano/daycounting/conventions -v

# Test daycounting
go test zpano/daycounting -v

# Run with coverage
go test ./daycounting/... -cover

# Run benchmarks
go test ./daycounting -bench=.
```

## Key Differences from TypeScript

### 1. Error Handling
Go uses explicit error returns instead of exceptions:

```go
// TypeScript: throws error
yearFrac(dt1, dt2, method)

// Go: returns error
result, err := daycounting.YearFrac(dt1, dt2, method)
if err != nil {
    // handle error
}
```

### 2. Package Structure
Go uses separate packages to avoid naming conflicts:

- `conventions` package: For `DayCountConvention` type and `FromString` function
- Main `daycounting` package: For calculation functions

### 3. Time Handling
Go uses `time.Time` from the standard library instead of JavaScript Date:

```go
dt := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
```

### 4. Function Names
Go uses PascalCase for exported functions (capitalized first letter):

```typescript
// TypeScript
eur30360(...)
actActIsda(...)
```

```go
// Go
daycounting.Eur30360(...)
daycounting.ActActIsda(...)
```

## Excel Compatibility

The package provides Excel-compatible functions:

- **Basis 0** (30/360): Use `THIRTY_360_US_EOM` (closest match)
- **Basis 1** (Actual/Actual): Use `ACT_ACT_EXCEL` (exact match)
- **Basis 2** (Actual/360): Use `ACT_360` (exact match)
- **Basis 3** (Actual/365): Use `ACT_365_FIXED` (exact match)
- **Basis 4** (European 30/360): Use `THIRTY_360_EU` (exact match)

## ISO 20022 Standards

The package implements day count conventions as specified in ISO 20022:

- A001: 30/360 US (THIRTY_360_US)
- A002: 30/365 (THIRTY_365)
- A004: Actual/360 (ACT_360)
- A005: Actual/365 Fixed (ACT_365_FIXED)
- A008: Actual/Actual ISDA (ACT_ACT_ISDA)
- A010: Actual/Actual AFB (ACT_ACT_AFB)
- A011: 30/360 European (THIRTY_360_EU)
- A012: 30E2/360 (THIRTY_360_EU_M2)
- A013: 30E3/360 (THIRTY_360_EU_M3)
- A014: Actual/365 Non-Leap (ACT_365_NONLEAP)

## Performance

The package includes benchmarks for common operations:

```bash
go test ./daycounting -bench=. -benchmem
```

Typical performance on modern hardware:
- RAW method: ~100-200 ns/op
- 30/360 methods: ~50-100 ns/op
- Actual methods: ~200-400 ns/op
- Actual/Actual methods: ~500-800 ns/op

## Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "time"

    "zpano/daycounting"
    "zpano/daycounting/conventions"
)

func main() {
    // Create dates
    dt1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
    dt2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

    // Calculate year fraction
    yf, err := daycounting.YearFrac(dt1, dt2, conventions.ACT_365_FIXED)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Year fraction: %.10f\n", yf)

    // Calculate day fraction
    df, err := daycounting.DayFrac(dt1, dt2, conventions.ACT_365_FIXED)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Day fraction: %.2f\n", df)
}
```

### Using String Conventions

```go
// Parse convention from string
conv, err := conventions.FromString("act/act excel")
if err != nil {
    log.Fatal(err)
}

yf, err := daycounting.YearFrac(dt1, dt2, conv)
if err != nil {
    log.Fatal(err)
}
```

### Bond Calculation Example

```go
// Calculate accrued interest using 30/360 European
settlementDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
lastCouponDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

fraction, err := daycounting.YearFrac(
    lastCouponDate,
    settlementDate,
    conventions.THIRTY_360_EU,
)
if err != nil {
    log.Fatal(err)
}

couponRate := 0.05 // 5%
faceValue := 1000.0
accruedInterest := faceValue * couponRate * fraction
fmt.Printf("Accrued interest: $%.2f\n", accruedInterest)
```

## References

- [Excel YEARFRAC function](https://support.microsoft.com/en-us/office/yearfrac-function-3844141e-c76d-4143-82b6-208454ddc6a8)
- [ISO 20022 Day Count Codes](https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
- [ISDA 2006 Definitions](https://web.archive.org/web/20140913145444/http://www.hsbcnet.com/gbm/attachments/standalone/2006-isda-definitions.pdf)
- [Wikipedia: Day Count Convention](https://en.wikipedia.org/wiki/Day_count_convention)

## License

MIT
