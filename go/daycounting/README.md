# Day counting

This package provides day counting calculation functions implementing the following conventions.

- **RAW**: Simple seconds-based calculation
- **30/360 variants**: US, US EOM, US NASD, EU (Eurobond), EU Model 2, EU Model 3, EU Plus
- **30/365**: 30-day month, 365-day year
- **Actual/360**: Actual days, 360-day year
- **Actual/365**: Fixed and Non-Leap variants
- **Actual/Actual**: Excel, ISDA, and AFB variants

## Structure

```text
daycounting/
├── doc.go                  # Package docs
├── conventions/
│   ├── conventions.go      # Day count conventions
│   └── conventions_test.go # Tests for conventions
├── daycounting.go          # Day counting functions
├── daycounting_test.go     # Tests for day counting functions
├── fractional.go           # Fraction calculation functions
├── fractional_test.go      # Tests for fractional functions
└── examples/
    └── compare
        └── main.go         # Prints day count comparison in markdown format
```

## Core functions

### Convention parsing

```go
import "portf_py/daycounting/conventions"

// Parse string to convention
conv, err := conventions.FromString("act/365 fixed")
// Returns conventions.ACT_365_FIXED
```

### Year fraction calculation

```go
import (
    "portf_py/daycounting"
    "portf_py/daycounting/conventions"
    "time"
)

// Calculate year fraction between two dates
dt1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
dt2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

yearFrac, err := daycounting.YearFrac(dt1, dt2, conventions.ACT_365_FIXED)
// Returns ~1.002739...
```

### Day fraction calculation

```go
// Calculate day fraction between two dates
dayFrac, err := daycounting.DayFrac(dt1, dt2, conventions.ACT_365_FIXED)
// Returns number of days
```

### Direct calculation functions

```go
result := daycounting.Eur30360(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.Eur30360Model2(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.Eur30360Model3(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.Eur30360Plus(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.US30360(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.US30360Eom(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.US30360Nasd(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.Thirty365(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.Act365Nonleap(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.Act365Fixed(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.Act360(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.ActActExcel(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.ActActIsda(2020, 1, 1, 2021, 1, 1, 0, 0, false)
result := daycounting.ActActAfb(2020, 1, 1, 2021, 1, 1, 0, 0, false)
```

## Running tests

```bash
# Test all packages
go test ./daycounting/...

# Test conventions
go test portf_py/daycounting/conventions -v

# Test daycounting
go test portf_py/daycounting -v

# Run with coverage
go test ./daycounting/... -cover

# Run benchmarks
go test ./daycounting -bench=. -benchmem
```

## Examples

### Basic usage

```go
package main

import (
    "fmt"
    "log"
    "time"

    "portf_py/daycounting"
    "portf_py/daycounting/conventions"
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

### Using string conventions

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

### Bond calculation example

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
