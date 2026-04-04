# Day Counting TypeScript Module

This directory contains TypeScript conversions of Python day counting modules from `py/daycounting/`.

## Files Converted

### Source Files
- **conventions.ts** - Converted from `conventions.py`
  - Enum `DayCountConvention` defining various day count conventions
  - Function `fromString()` to parse string conventions

- **daycounting.ts** - Converted from `daycounting.py`
  - Core day counting calculation functions
  - Julian day conversion utilities
  - Various 30/360 and Actual/Actual methods

- **fractional.ts** - Converted from `fractional.py`
  - High-level API functions `yearFrac()` and `dayFrac()`
  - Integrates all day counting methods

### Test Files (Jasmine)
- **conventions_spec.ts** - Converted from `test_daycounting_conventions.py`
- **daycounting_spec.ts** - Converted from `test_daycounting.py`
- **fractional_spec.ts** - Converted from `test_daycounting_fractional.py`

## Key Differences from Python

1. **Enums**: Python's `Enum` class converted to TypeScript `enum`
2. **Type Annotations**: Python type hints converted to TypeScript types
3. **Date Handling**: JavaScript `Date` objects instead of Python `datetime`
   - Note: JavaScript months are 0-indexed (0=January), Python uses 1-indexed
4. **Default Parameters**: Both languages support default parameters similarly
5. **Maps vs Dicts**: Python dictionaries converted to TypeScript `Map` or `Record`

## Setup for Running Tests

### Install Dependencies

First, you'll need to set up a Node.js/TypeScript project. If you don't have a `package.json` yet:

```bash
npm init -y
```

### Install TypeScript and Jasmine

```bash
npm install --save-dev typescript @types/node
npm install --save-dev jasmine @types/jasmine
```

### Configure TypeScript

Create a `tsconfig.json` in the `ts` directory:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "lib": ["ES2020"],
    "outDir": "./dist",
    "rootDir": "./",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "moduleResolution": "node"
  },
  "include": ["**/*.ts"],
  "exclude": ["node_modules", "dist"]
}
```

### Configure Jasmine

Initialize Jasmine:

```bash
npx jasmine init
```

Create `spec/support/jasmine.json`:

```json
{
  "spec_dir": "daycounting",
  "spec_files": [
    "**/*_spec.ts"
  ],
  "helpers": [
    "helpers/**/*.js"
  ],
  "stopSpecOnExpectationFailure": false,
  "random": false
}
```

### Add NPM Scripts

Add to your `package.json`:

```json
{
  "scripts": {
    "build": "tsc",
    "test": "npm run build && jasmine",
    "test:watch": "nodemon --exec 'npm test' --watch daycounting"
  }
}
```

### Running Tests

```bash
# Build TypeScript files
npm run build

# Run tests
npm test
```

## Usage Example

```typescript
import { yearFrac, dayFrac } from './daycounting/fractional';
import { DayCountConvention } from './daycounting/conventions';

// Calculate year fraction using different conventions
const date1 = new Date(2020, 0, 1); // January 1, 2020
const date2 = new Date(2021, 0, 1); // January 1, 2021

// Using Actual/Actual Excel convention
const years = yearFrac(date1, date2, DayCountConvention.ACT_ACT_EXCEL);
console.log(`Years: ${years}`); // 1.0

// Using 30/360 European convention
const years360 = yearFrac(date1, date2, DayCountConvention.THIRTY_360_EU);
console.log(`Years (30/360 EU): ${years360}`);

// Get day fraction
const days = dayFrac(date1, date2, DayCountConvention.RAW);
console.log(`Days: ${days}`); // 366
```

## API Reference

### DayCountConvention Enum

Available conventions:
- `RAW` - Raw seconds difference
- `THIRTY_360_US` - 30/360 US (ISDA)
- `THIRTY_360_US_EOM` - 30/360 US End-of-Month
- `THIRTY_360_US_NASD` - 30/360 NASD
- `THIRTY_360_EU` - 30/360 European (Excel basis 4)
- `THIRTY_360_EU_M2` - 30E2/360 Eurobond model 2
- `THIRTY_360_EU_M3` - 30E3/360 Eurobond model 3
- `THIRTY_360_EU_PLUS` - 30E+/360
- `THIRTY_365` - 30/365
- `ACT_360` - Actual/360 (Excel basis 2)
- `ACT_365_FIXED` - Actual/365 Fixed (Excel basis 3)
- `ACT_365_NONLEAP` - Actual/365 Non-Leap
- `ACT_ACT_EXCEL` - Actual/Actual Excel (basis 1)
- `ACT_ACT_ISDA` - Actual/Actual ISDA
- `ACT_ACT_AFB` - Actual/Actual AFB

### Functions

#### `fromString(convention: string): DayCountConvention`
Converts a string to a DayCountConvention enum value.

#### `yearFrac(dateTime1: Date, dateTime2: Date, method?: DayCountConvention): number`
Calculates the year fraction between two dates using the specified convention.

#### `dayFrac(dateTime1: Date, dateTime2: Date, method?: DayCountConvention): number`
Calculates the day fraction between two dates using the specified convention.

## Notes

- The TypeScript code maintains the same Excel compatibility as the Python version
- Test files use Jasmine instead of Python's unittest
- All numeric precision tests have been preserved with appropriate rounding
- The code has been typed for better IDE support and compile-time safety

## References

- [ISO 20022 Day Count Conventions](https://www.iso20022.org/15022/uhb/mt565-16-field-22f.htm)
- [Excel YEARFRAC Function](https://support.microsoft.com/en-us/office/yearfrac-function-3844141e-c76d-4143-82b6-208454ddc6a8)
- [ISDA 2006 Definitions](https://web.archive.org/web/20140913145444/http://www.hsbcnet.com/gbm/attachments/standalone/2006-isda-definitions.pdf)
