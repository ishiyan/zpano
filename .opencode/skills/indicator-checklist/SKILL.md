# Indicator Checklist

Compact verification rules for indicator implementations. Three categories, mechanically verifiable.

## 1. Import Patterns

| Language | Entities import | Core import |
|----------|----------------|-------------|
| **Go** | `"zpano/entities"` | `"zpano/indicators/core"` |
| **TS** | `from '../../entities/bar'` (relative, per-entity) | `from '../../core/line-indicator'` (relative, per-file) |
| **Python** | `from ....entities.bar import Bar` (relative, per-entity) | `from ...core.line_indicator import LineIndicator` (relative, per-file) |
| **Zig** | `@import("entities")` (single barrel, build.zig module) | `@import("../../core/indicator.zig")` (relative file paths) |
| **Rust** | `use crate::entities::bar::Bar` (absolute crate paths) | `use crate::indicators::core::indicator::Indicator` (absolute crate paths) |

**Rules:**
- Zig: NEVER import individual entity modules (`@import("bar")`). Always use the `entities` barrel.
- Go: NEVER import `"zpano/entities/bar"` — the `entities` package is flat.
- Python: entities are 4 levels up (`....entities`), core is 3 levels up (`...core`).
- Rust: Always `use crate::` prefix. Never `use super::` for cross-module imports.
- TS: Never barrel imports from entities (`../../entities`). Import each entity file individually.

## 2. Naming Conventions

| Aspect | Go | TS | Python | Zig | Rust |
|--------|----|----|--------|-----|------|
| **Directory** | `simplemovingaverage/` | `simple-moving-average/` | `simple_moving_average/` | `simple_moving_average/` | `simple_moving_average/` |
| **File** | `simplemovingaverage.go` | `simple-moving-average.ts` | `simple_moving_average.py` | `simple_moving_average.zig` | `simple_moving_average.rs` |
| **Type** | `SimpleMovingAverage` | `SimpleMovingAverage` | `SimpleMovingAverage` | `SimpleMovingAverage` | `SimpleMovingAverage` |
| **Params type** | (fields on struct) | `SimpleMovingAverageParams` | `SimpleMovingAverageParams` | (fields on struct or params) | `SimpleMovingAverageParams` |
| **Output enum** | `SimpleMovingAverageOutput` | `SimpleMovingAverageOutput` | `SimpleMovingAverageOutput` | `SimpleMovingAverageOutput` | `SimpleMovingAverageOutput` |
| **Output members** | `Value = 1` (iota+1) | `VALUE = 1` | `VALUE = 1` | `value = 1` | `Value = 1` |
| **Mnemonic fn** | `Mnemonic(params)` method | `simpleMovingAverageMnemonic(params)` | `simple_moving_average_mnemonic(params)` | `mnemonic(params)` | `mnemonic(params)` |
| **Test file** | `simplemovingaverage_test.go` | `simple-moving-average.spec.ts` | `test_simple_moving_average.py` | (bottom of source file) | (bottom of source file via `#[cfg(test)]`) |

## 3. Structural Patterns

| Aspect | Go | TS | Python | Zig | Rust |
|--------|----|----|--------|-----|------|
| **Base** | Embed `core.LineIndicator` | `extends LineIndicator` | `class X(Indicator):` with LineIndicator mixin | `LineIndicator` field + delegation | Implement `Indicator` trait, use `LineIndicator` helper |
| **Concurrency** | `sync.RWMutex` field, lock in all methods | None | None | None | None |
| **Component sentinel** | `if bc == 0` (zero-value) | `if (bc === undefined)` | `if bc is None` | `if (bc == null)` on `?BarComponent` | `if bc.is_none()` on `Option<BarComponent>` |
| **Params file** | None (inline) | `./params.ts` separate | `./params.py` separate | Inline or `./params.zig` | Inline in same file |
| **Constructor validation** | Return `error` | `throw new Error(...)` | `raise ValueError(...)` | Return `error` | `panic!` or return `Result` |
| **Update method** | `Update(sample float64)` | `update(sample: number)` | `update(self, sample: float)` | `update(self, sample: f64)` | `update(&mut self, sample: f64)` |
| **Primed check** | `IsPrimed() bool` | `get isPrimed(): boolean` | `@property is_primed` | `is_primed() bool` | `is_primed(&self) -> bool` |
| **Output values** | `Value() float64` (per-output getter) | `get value(): number` | `@property value` | `value() f64` | `value(&self) -> f64` |
| **Descriptor** | Registered in `core/descriptor.go` | Registered in `core/descriptor.ts` | Registered in `core/descriptor.py` | Registered in `core/descriptor.zig` | Registered in `core/descriptor.rs` |
| **Factory** | Entry in `factory/factory.go` | Entry in `factory/factory.ts` | Entry in `factory/factory.py` | Entry in `factory/factory.zig` | Entry in `factory/factory.rs` |

## Quick Verification Commands

```bash
# Go: check an indicator compiles and tests pass
cd go && go test ./indicators/common/simplemovingaverage -v

# TS: check build + specific spec
cd ts && npm run build && npx jasmine --filter="SimpleMovingAverage"

# Python: run indicator test
PYTHONPATH=. python3 -m unittest py.indicators.common.simple_moving_average.test_simple_moving_average

# Zig: run all indicator tests (no filter available via build)
cd zig && zig build test --summary all

# Rust: run specific indicator test
cd rs && cargo test --lib simple_moving_average
```
