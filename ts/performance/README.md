# Performance Metrics Module

This module provides comprehensive performance and risk metrics for portfolio analysis, converted from Python to TypeScript.

## Features

- **Periodicity**: Enum for different time periodicities (DAILY, WEEKLY, MONTHLY, QUARTERLY, ANNUAL)
- **Ratios**: Comprehensive class for calculating various performance ratios

## Available Ratios

### Risk-Adjusted Returns
- **Sharpe Ratio**: Risk-adjusted return using standard deviation
- **Sortino Ratio**: Downside risk-adjusted return using downside deviation
- **Omega Ratio**: Probability-weighted ratio of gains to losses
- **Kappa Ratio**: Generalized downside risk-adjusted performance
- **Calmar Ratio**: Return relative to maximum drawdown
- **Sterling Ratio**: Return relative to average drawdown
- **Burke Ratio**: Return relative to square root of sum of squared drawdowns

### Drawdown Metrics
- **Martin Ratio**: Return relative to Ulcer Index
- **Pain Ratio**: Return relative to Pain Index
- **Maximum Drawdown**: Largest peak-to-trough decline
- **Average Drawdown**: Mean of all drawdowns

### Statistical Metrics
- **Skewness**: Distribution asymmetry
- **Kurtosis**: Distribution tail heaviness
- **Upside Potential Ratio**: Upside potential relative to downside risk
- **Upside Risk**: Risk of returns above target
- **Downside Risk**: Risk of returns below target

## Installation

```bash
npm install
```

## Building

```bash
npm run build
```

## Testing

Run the Jasmine test suite:

```bash
npm test
```

Run with coverage:

```bash
npm run test:coverage
```

## Usage

```typescript
import { Ratios, Periodicity } from '@portf_py/performance';
import { DayCountConvention } from '../daycounting';

// Create a ratios calculator
const ratios = new Ratios(
    Periodicity.DAILY,
    0.02,  // Annual risk-free rate (2%)
    0,     // Target return (MAR)
    DayCountConvention.RAW
);

// Reset before calculating
ratios.reset();

// Add returns sequentially
for (let i = 0; i < returns.length; i++) {
    ratios.addReturn(
        portfolioReturns[i],
        benchmarkReturns[i],
        1,  // weight
        previousDates[i],
        currentDates[i]
    );
}

// Calculate various ratios
const sharpe = ratios.sharpeRatio();
const sortino = ratios.sortinoRatio();
const omega = ratios.omegaRatio();
const calmar = ratios.calmarRatio();
const maxDD = ratios.maxDrawdown;
const kurtosis = ratios.kurtosis;
```

## Implementation Notes

### Statistical Functions

This module implements custom statistical functions to avoid external dependencies:

- **Mean**: Simple arithmetic mean
- **Standard Deviation**: Sample standard deviation (n-1 divisor)
- **Skewness**: Fisher-Pearson coefficient of skewness
- **Kurtosis**: Excess kurtosis adjusted for sample size

### Date Handling

JavaScript Date objects are used with 0-indexed months (January = 0), different from Python's 1-indexed months.

### Precision

All calculations maintain high precision and are tested against R PerformanceAnalytics package outputs with 13+ decimal places of accuracy.

## Test Data

Tests use the "Portfolio Bacon" dataset from the R PerformanceAnalytics package for validation:
- 24 daily portfolio and benchmark returns
- Expected values computed from R package functions
- Tolerance of 1e-13 for floating-point comparisons

## License

MIT
