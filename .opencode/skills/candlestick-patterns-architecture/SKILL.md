---
name: candlestick-patterns-architecture
description: Architecture, folder layout, naming conventions, and design patterns for the zpano candlestick pattern recognition library. Load when creating new patterns, porting patterns across languages, or understanding the codebase structure.
---

# Candlestick Patterns Architecture

This document describes the design decisions, folder layout, and naming conventions for the **zpano** candlestick pattern recognition library. It covers the fuzzy-logic-based engine, the 61 pattern implementations, and the streaming criterion system.

> **Related skill:** The `fuzzy-architecture` skill covers the standalone fuzzy logic primitives (membership functions, operators, defuzzify) that this module depends on. Load it for details on sigmoid/linear membership, t-norms, s-norms, and alpha-cut.

## Overview

The candlestick patterns module recognizes 61 Japanese candlestick patterns using **fuzzy logic** instead of crisp boolean conditions. Each pattern returns a continuous value in [-100, +100] (or [-200, +200] for confirmation signals), representing the degree of pattern membership. An `alpha_cut` function converts these to crisp signals for trading decisions.

### Key Design Principles

1. **Fuzzy over crisp.** Traditional pattern recognition uses hard thresholds (body > average = "long body"). This library uses sigmoid membership functions with configurable transition widths, producing graded pattern strength.
2. **Streaming / incremental.** The engine processes one OHLC bar at a time via `update(o, h, l, c)`. All criterion averages are maintained as running totals with O(1) updates.
3. **One file per pattern.** Each of the 61 patterns lives in its own file under `patterns/`. Pattern functions take the engine as receiver/parameter and use its fuzzy helpers.
4. **Cross-language parity.** All five languages (Python, Go, TypeScript, Zig, Rust) produce identical results. Tests use `alpha_cut` comparison to handle floating-point edge cases.

## Module Dependencies

```
fuzzy/              (standalone — membership functions, operators, defuzzify)
  |
  v
candlestick_patterns/
  core/             (primitives, criteria, defaults, identifiers, registry)
  patterns/         (61 pattern files + test data)
  engine            (CandlestickPatterns / CandlestickPatternsEngine)
```

The candlestick module depends on:
- **fuzzy/** — `mu_less`, `mu_greater`, `mu_near`, `mu_direction`, `t_product`, `t_product_all`, `t_min_all`, `f_not`, `alpha_cut`
- **entities/** — only `Bar` type for the thin wrapper class (not used by engine internals)

## Folder Layout

### Python (`py/candlestick_patterns/`)
```
candlestick_patterns.py          # Engine class (CandlestickPatterns)
core/
    __init__.py
    primitives.py                # Pure OHLC functions (real_body, shadows, gaps, enclosure)
    criterion.py                 # Criterion dataclass (entity, period, factor)
    defaults.py                  # 11 default Criterion constants
    range_entity.py              # RangeEntity enum (REAL_BODY, HIGH_LOW, SHADOWS)
    pattern_identifier.py        # PatternIdentifier enum (61 values, 0-60)
    pattern_registry.py          # PatternInfo + PATTERN_REGISTRY dict (mnemonic, kanji, reading, description)
patterns/
    doji.py, hammer.py, ...      # 61 pattern files (one function each)
    test_patterns.py             # Test harness
    test_data_*.py               # 61 test data files
```

### Go (`go/candlestickpatterns/`)
```
candlestickpatterns.go           # Thin wrapper (holds *core.CandlestickPatterns)
core/
    criterion.go                 # Criterion struct
    criterionstate.go            # CriterionState (ring buffer + running total)
    defaults.go                  # 11 default constants
    engine.go                    # CandlestickPatterns engine struct + fuzzy helpers
    ohlc.go                      # OHLC tuple type
    patternidentifier.go         # PatternIdentifier enum (iota)
    patternregistry.go           # PatternInfo + Registry map
    primitives.go                # Pure OHLC functions
    rangeentity.go               # RangeEntity enum
patterns/
    abandonedbaby.go, ...        # 61 pattern files (no underscores in filenames)
    patterns_test.go             # Single test file
    testdata_*_test.go           # 61 test data files (with _test.go suffix)
```

### TypeScript (`ts/candlestick-patterns/`)
```
candlestick-patterns.ts          # Thin wrapper class
candlestick-patterns.spec.ts     # Tests
index.ts                         # Barrel exports
core/
    criterion.ts
    defaults.ts
    engine.ts                    # CandlestickPatternsEngine + CriterionState
    index.ts                     # Barrel
    pattern-identifier.ts
    pattern-registry.ts
    primitives.ts
    range-entity.ts
patterns/
    abandoned-baby.ts, ...       # 61 pattern files (kebab-case)
    index.ts                     # Barrel
    testdata-*.ts                # 61 test data files
```

### Zig (`zig/src/candlestick_patterns/`)
```
candlestick_patterns.zig         # Engine
candlestick_patterns_test.zig    # Tests
core/
    core.zig                     # Barrel / re-exports
    criterion.zig
    criterion_state.zig
    defaults.zig
    ohlc.zig
    pattern_identifier.zig
    primitives.zig
    range_entity.zig
patterns/
    abandoned_baby.zig, ...      # 61 pattern files
    patterns.zig                 # Barrel
    test_case.zig                # TestCase struct helper
    testdata_*.zig               # 61 test data files
```

### Rust (`rs/src/candlestick_patterns/`)
```
candlestick_patterns.rs          # Engine
candlestick_patterns_test.rs     # Tests
mod.rs                           # Module root
core/
    criterion.rs
    criterion_state.rs
    defaults.rs
    mod.rs
    ohlc.rs
    pattern_identifier.rs
    pattern_registry.rs
    primitives.rs
    range_entity.rs
patterns/
    abandoned_baby.rs, ...       # 61 pattern files
    mod.rs
    testdata_*.rs                # 61 test data files
```

## Core Components

### RangeEntity

Enum with 3 values controlling which part of the candle a criterion measures:

| Value | Meaning | Calculation |
|-------|---------|-------------|
| `REAL_BODY` (0) | Candle body | `abs(close - open)` |
| `HIGH_LOW` (1) | Full range | `high - low` |
| `SHADOWS` (2) | Average shadow | `(upper_shadow + lower_shadow) / 2` |

### Criterion

A triple `(entity: RangeEntity, average_period: int, factor: float)` defining what to measure and how:

- **entity** — which candle dimension to measure
- **average_period** — number of bars for the running average (0 = use current candle only)
- **factor** — multiplier on the average (e.g., 3.0 for "very long body" = 3x the average body)

Key method: `average_value_from_total(total, o, h, l, c)`:
- When `period > 0`: returns `factor * total / period` (or `/ (period * 2)` for SHADOWS)
- When `period == 0`: returns `factor * candle_range_value(entity, o, h, l, c)`

### Default Criteria (11 constants)

| Name | Entity | Period | Factor | Meaning |
|------|--------|--------|--------|---------|
| `LONG_BODY` | REAL_BODY | 10 | 1.0 | Body >= 1x average body |
| `VERY_LONG_BODY` | REAL_BODY | 10 | 3.0 | Body >= 3x average body |
| `SHORT_BODY` | REAL_BODY | 10 | 1.0 | Body <= 1x average body |
| `DOJI_BODY` | HIGH_LOW | 10 | 0.1 | Body <= 10% of average range |
| `LONG_SHADOW` | REAL_BODY | 0 | 1.0 | Shadow >= 1x own body |
| `VERY_LONG_SHADOW` | REAL_BODY | 0 | 2.0 | Shadow >= 2x own body |
| `SHORT_SHADOW` | SHADOWS | 10 | 1.0 | Shadow <= 1x average shadow |
| `VERY_SHORT_SHADOW` | HIGH_LOW | 10 | 0.1 | Shadow <= 10% of average range |
| `NEAR` | HIGH_LOW | 5 | 0.2 | Within 20% of average range |
| `FAR` | HIGH_LOW | 5 | 0.6 | Within 60% of average range |
| `EQUAL` | HIGH_LOW | 5 | 0.05 | Within 5% of average range |

Note: `LONG_SHADOW` and `VERY_LONG_SHADOW` have `period=0`, comparing against the **current candle's own body** rather than a historical average.

### CriterionState

Maintains a sliding-window running total for one criterion:

- **ring buffer**: `deque[float]` with maxlen = `average_period + max_shift` (or None if period==0)
- **running total**: maintained incrementally via `push(o, h, l, c)`
- **`total_at(shift)`**: recomputes sum from ring for a shifted window position
- **`avg(shift, o, h, l, c)`**: delegates to `criterion.average_value_from_total()`

The `max_shift` parameter (default 5) allows patterns to look back at criterion averages from previous bars.

### Primitives

Pure functions for candle geometry (no state, no fuzzy logic):

**Color:** `is_white(o, c)` (c >= o), `is_black(o, c)` (c < o)

**Body:** `real_body(o, c)`, `white_real_body(o, c)`, `black_real_body(o, c)`

**Shadows:** `upper_shadow(o, h, c)`, `lower_shadow(o, l, c)`, plus color-specific variants (`white_upper_shadow`, `black_lower_shadow`, etc.)

**Gaps:** `is_real_body_gap_up/down(o1, c1, o2, c2)`, `is_high_low_gap_up/down(h1/l1, l2/h2)`

**Enclosure:** `is_real_body_encloses_real_body(o1, c1, o2, c2)`, `is_real_body_encloses_open/close`

**Range value:** `candle_range_value(entity, o, h, l, c)` — dispatches by RangeEntity

### PatternIdentifier

Enum with 61 values (0-60), alphabetically ordered:

```
ABANDONED_BABY=0, ADVANCE_BLOCK=1, BELT_HOLD=2, BREAKAWAY=3,
CLOSING_MARUBOZU=4, CONCEALING_BABY_SWALLOW=5, COUNTERATTACK=6,
DARK_CLOUD_COVER=7, DOJI=8, DOJI_STAR=9, DRAGONFLY_DOJI=10,
ENGULFING=11, EVENING_DOJI_STAR=12, EVENING_STAR=13, GRAVESTONE_DOJI=14,
HAMMER=15, HANGING_MAN=16, HARAMI=17, HARAMI_CROSS=18, HIGH_WAVE=19,
HIKKAKE=20, HIKKAKE_MODIFIED=21, HOMING_PIGEON=22,
IDENTICAL_THREE_CROWS=23, IN_NECK=24, INVERTED_HAMMER=25, KICKING=26,
KICKING_BY_LENGTH=27, LADDER_BOTTOM=28, LONG_LEGGED_DOJI=29,
LONG_LINE=30, MARUBOZU=31, MATCHING_LOW=32, MAT_HOLD=33,
MORNING_DOJI_STAR=34, MORNING_STAR=35, ON_NECK=36, PIERCING=37,
RICKSHAW_MAN=38, RISING_FALLING_THREE_METHODS=39, SEPARATING_LINES=40,
SHOOTING_STAR=41, SHORT_LINE=42, SPINNING_TOP=43, STALLED=44,
STICK_SANDWICH=45, TAKURI=46, TASUKI_GAP=47, THREE_BLACK_CROWS=48,
THREE_INSIDE=49, THREE_LINE_STRIKE=50, THREE_OUTSIDE=51,
THREE_STARS_IN_THE_SOUTH=52, THREE_WHITE_SOLDIERS=53, THRUSTING=54,
TRISTAR=55, TWO_CROWS=56, UNIQUE_THREE_RIVER=57,
UP_DOWN_GAP_SIDE_BY_SIDE_WHITE_LINES=58, UPSIDE_GAP_TWO_CROWS=59,
X_SIDE_GAP_THREE_METHODS=60
```

Has a `method_name` property returning the lowercase name (e.g., `"abandoned_baby"`).

### Pattern Registry

Maps each `PatternIdentifier` to a `PatternInfo` with:
- `mnemonic: str` — human-readable name (e.g., `"abandoned baby"`)
- `kanji: str | None` — Japanese kanji (e.g., `"捨て子線"`) or None for Western-origin patterns
- `reading: str | None` — hiragana reading (e.g., `"すてごせん"`) or None
- `description: str` — multi-sentence explanation of the pattern

## Engine Architecture

### CandlestickPatterns Engine

The central class that processes OHLC bars and evaluates patterns.

**Constructor parameters:**
- 11 optional `Criterion` overrides (default to the 11 constants above)
- `fuzz_ratio: float` (default 0.2) — width of fuzzy transition zones as fraction of criterion average
- `shape: MembershipShape` (default SIGMOID) — sigmoid or linear membership curves

**Internal state:**
- 11 `CriterionState` instances (one per criterion)
- `history: deque[OHLC]` — ring buffer of recent bars, maxlen = max(max_period + 10, 20)
- `count: int` — total bars processed
- Hikkake-modified state: `hikmod_pattern_result`, `hikmod_pattern_idx`, `hikmod_confirmed`, `hikmod_last_signal`

**Streaming interface:**
- `update(o, h, l, c)` — append bar to history, push all criterion states, increment count, update hikkake_modified state

**Bar access:**
- `bar(shift)` — get OHLC at shift (1 = most recent, 2 = previous, etc.)
- `has(n)` — check if >= n bars available
- `enough(n_candles, *criteria)` — check bars >= n_candles + each criterion's average_period

**Criterion averages:**
- `avg(criterion_state, shift)` — criterion average at given shift
- `avg_ref(criterion_state, shift, ref_shift)` — average window ending at `shift`, using OHLC from `ref_shift`

**Fuzzy helpers** (all delegate to fuzzy module functions):
- `mu_less(value, criterion, shift)` — degree value < criterion average
- `mu_greater(value, criterion, shift)` — degree value > criterion average
- `mu_near_value(value, target, criterion, shift)` — degree value ≈ target
- `mu_ge_raw(value, threshold, width)` — raw greater-or-equal membership
- `mu_gt_raw(value, threshold, width)` — raw greater-than membership
- `mu_lt_raw(value, threshold, width)` — raw less-than membership
- `mu_bullish(o, c, shift)` — max(0, direction) — bullish degree
- `mu_bearish(o, c, shift)` — max(0, -direction) — bearish degree
- `mu_direction_raw(o, c, shift)` — raw [-1, +1] direction using `mu_direction` with short_body avg and steepness=2.0
- `width(criterion, shift)` — `fuzz_ratio * avg`

**Pattern dispatch:**
- `evaluate(pattern_identifier) -> int` — dispatches to pattern function, returns result

### Thin Wrapper

Each language has a thin wrapper at the module root that:
- Holds a reference to the engine
- Provides `update(bar)` accepting a `Bar` entity (extracting OHLC)
- Delegates all 61 pattern methods to the engine

Go: `CandlestickPatterns` struct in `candlestickpatterns.go` holding `*core.CandlestickPatterns`
TS: `CandlestickPatterns` class in `candlestick-patterns.ts` holding `CandlestickPatternsEngine`

## Pattern Categories

### Category A — Fixed Direction
The pattern has an inherent direction (always bullish or always bearish). Returns a positive or negative value directly.

Examples: `hammer` (always bullish → positive), `shooting_star` (always bearish → negative)

### Category B — Direction from Candle Color
The pattern's direction is determined by the color (white/black) of a key candle. Uses `mu_bullish`/`mu_bearish` to set sign.

Examples: `engulfing`, `belt_hold`, `marubozu`

### Category C — Both Branches Evaluated
Both bullish and bearish interpretations are computed; the stronger one wins. Returns `bull - bear` (net direction).

Examples: `harami`, `doji_star`, `morning_star`/`evening_star`

### Special: Hikkake (Crisp + Stateful)
- `hikkake` — crisp logic (not fuzzy), returns ±100 for detection, ±200 for confirmation
- `hikkake_modified` — extends hikkake with cross-bar state tracking via `hikmod_*` fields updated in `update()`

These are the only patterns that don't use fuzzy membership.

## Pattern Function Convention

All patterns follow this signature:
- **Python:** `def pattern_name(self) -> float` (method on engine class, imported as mix-in)
- **Go:** `func PatternName(cp *CandlestickPatterns) float64` (standalone function taking engine)
- **TypeScript:** `export function patternName(cp: CandlestickPatternsEngine): number`
- **Zig:** `pub fn patternName(self: *CandlestickPatterns) f64`
- **Rust:** `pub fn pattern_name(cp: &CandlestickPatterns) -> f64` (or `&mut` for hikkake_modified)

Each pattern function:
1. Checks `enough(n_candles, ...criteria)` — returns 0 if insufficient data
2. Extracts OHLC values from `bar(shift)` for relevant candles
3. Computes fuzzy memberships using engine helpers
4. Combines memberships with t-norms (`t_product`, `t_product_all`, `t_min_all`)
5. Returns result scaled to [-100, +100]

## Test Architecture

### Test Data Format

Each pattern has a dedicated test data file exporting a list of test cases. Each case contains:
- `opens: [float, ...]` — sequence of open prices
- `highs: [float, ...]` — sequence of high prices
- `lows: [float, ...]` — sequence of low prices
- `closes: [float, ...]` — sequence of close prices
- `expected: float` — expected output from TA-Lib reference

The number of bars per case varies by pattern (1-bar patterns like doji have short sequences; 5-bar patterns like breakaway have longer ones).

### Test Comparison via Alpha-Cut

Tests do NOT compare raw fuzzy values. Instead:
1. Run pattern, get actual fuzzy value
2. Apply `alpha_cut(expected)` and `alpha_cut(actual)` — converts to crisp {-200, -100, 0, +100, +200}
3. Compare crisp values

This approach tolerates fuzzy membership differences as long as they're on the same side of the 50% threshold.

### Known Fuzzy Divergences

`_KNOWN_FUZZY_DIVERGENCES` is a set of `(pattern_name, case_index)` tuples marking borderline cases where the fuzzy membership is ~0.5 and the alpha-cut may differ from TA-Lib's crisp logic. These are skipped in all languages. ~80+ entries across patterns including hammer, hanging_man, shooting_star, takuri, harami, harami_cross, tristar, etc.

### Test Harness Pattern (all languages)

```
for each test case:
    create fresh engine with default criteria
    feed all bars via update(o, h, l, c)
    call pattern method
    apply alpha_cut to both expected and actual
    if (pattern, index) in KNOWN_DIVERGENCES: skip
    assert alpha_cut(expected) == alpha_cut(actual)
```

### Testdata File Naming

| Language | Pattern | Example |
|----------|---------|---------|
| Python | `test_data_<pattern>.py` | `test_data_abandoned_baby.py` |
| Go | `testdata<pattern>_test.go` | `testdataabandonedbaby_test.go` |
| TypeScript | `testdata-<pattern>.ts` | `testdata-abandoned-baby.ts` |
| Zig | `testdata_<pattern>.zig` | `testdata_abandoned_baby.zig` |
| Rust | `testdata_<pattern>.rs` | `testdata_abandoned_baby.rs` |

## Cross-Language Naming Conventions

| Concept | Python | Go | TypeScript | Zig | Rust |
|---------|--------|-----|-----------|-----|------|
| Engine class | `CandlestickPatterns` | `CandlestickPatterns` (core) | `CandlestickPatternsEngine` | `CandlestickPatterns` | `CandlestickPatterns` |
| Pattern function | `def doji(self)` | `func Doji(cp *CP)` | `function doji(cp)` | `fn doji(self)` | `fn doji(cp)` |
| Criterion field | `_long_body` | `longBody` | `longBody` | `long_body` | `long_body` |
| Enum member | `ABANDONED_BABY` | `AbandonedBaby` | `ABANDONED_BABY` | `abandoned_baby` | `AbandonedBaby` |
| Default constant | `DEFAULT_LONG_BODY` | `DefaultLongBody` | `DEFAULT_LONG_BODY` | `default_long_body` | `DEFAULT_LONG_BODY` |
| File naming | `snake_case.py` | `nounderscores.go` | `kebab-case.ts` | `snake_case.zig` | `snake_case.rs` |
| Pattern dir | `patterns/` | `patterns/` | `patterns/` | `patterns/` | `patterns/` |

### Fuzzy Import Patterns

| Language | Engine imports fuzzy as | Pattern files import operators as |
|----------|------------------------|----------------------------------|
| Python | `from ..fuzzy import mu_less, ...` (individual symbols) | `from ..fuzzy import t_product_all, ...` |
| Go | `import "zpano/fuzzy"` → `fuzzy.MuLess(...)` | `import "zpano/fuzzy"` → `fuzzy.TProductAll(...)` |
| TypeScript | `import { muLess, ... } from '../../fuzzy';` | `import { tProductAll, ... } from '../../fuzzy';` |
| Zig | `const fuzzy = @import("fuzzy");` then `fuzzy.membership`, `fuzzy.operators`, `fuzzy.defuzzify` | `const operators = @import("fuzzy").operators;` |
| Rust | `use crate::fuzzy::{mu_less, ...};` | `use crate::fuzzy::{t_product_all, ...};` |

Zig uses a single `fuzzy` barrel module (`src/fuzzy/fuzzy.zig`) registered in build.zig, which re-exports `membership`, `operators`, and `defuzzify` as sub-namespaces. Consumer files create local aliases: `const membership = fuzzy.membership;`.

## Fuzzy Logic Application

### Membership Width

The `width` for fuzzy membership is computed as:
```
width = fuzz_ratio * criterion_average_at_shift
```

With default `fuzz_ratio=0.2`, the transition zone is 20% of the criterion's running average. This means:
- For a criterion average of 5.0, the transition zone spans ~1.0 price units
- Membership transitions smoothly from 0 to 1 across this zone

### Sigmoid Parameters

- `k = 12 / width` — steepness of the sigmoid
- Exponent clamped at [-500, +500] to prevent overflow
- `mu_near` uses a Gaussian bell with `sigma = width / 2.41`

### Direction Membership

`mu_direction(o, c, body_avg, steepness=2.0)` returns [-1, +1]:
```
tanh(steepness * (close - open) / body_avg)
```

- `mu_bullish = max(0, direction)` — positive for white candles
- `mu_bearish = max(0, -direction)` — positive for black candles
- Uses `short_body` criterion average as `body_avg`

### Combining Memberships

Patterns combine multiple conditions using:
- `t_product(a, b)` = `a * b` — fuzzy AND (most common)
- `t_product_all(*args)` — product of all memberships
- `t_min_all(*args)` — min of all memberships (used for strict conjunction)
- `f_not(a)` = `1 - a` — fuzzy NOT

### Output Scale

Most patterns return `membership * 100.0` (or `-100.0` for bearish). The sign indicates direction; the magnitude indicates confidence.

## Adding a New Pattern

1. Create pattern file in `patterns/` following the one-function-per-file convention
2. Add `PatternIdentifier` enum value (maintain alphabetical order)
3. Add `PatternInfo` entry in the registry (mnemonic, kanji if applicable, description)
4. Register in the engine's dispatch map
5. Create test data file with TA-Lib reference cases
6. Add test method to test harness
7. Port to all five languages maintaining identical logic
