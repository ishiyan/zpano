# Indicator Checklist

Verification rules for indicator implementations. Focus only on the column
corresponding to the language you are currently implementing or verifying.
Rules are mechanically verifiable — check each one after conversion/implementation.

---

## 1. Directory & File Layout

| Aspect | Go | TS | Python | Zig | Rust |
|--------|----|----|--------|-----|------|
| **Author dir** | `lowercasejoined/` | `kebab-case/` | `snake_case/` | `snake_case/` | `snake_case/` |
| **Indicator dir** | `lowercasejoined/` | `kebab-case/` | `snake_case/` | `snake_case/` | `snake_case/` |
| **Main file** | `<indicator>.go` | `<indicator>.ts` | `<indicator>.py` | `<indicator>.zig` | `<indicator>.rs` |
| **Test file** | `<indicator>_test.go` | `<indicator>.spec.ts` | `test_<indicator>.py` | bottom of source | `#[cfg(test)]` in source |
| **Test data** | `testdata_test.go` | `testdata.ts` | `test_testdata.py` | bottom of source / `testdata.zig` | inline in test module |
| **Params file** | `params.go` (no prefix) | `params.ts` (no prefix) | inline or separate | single file | inline |
| **Output file** | `output.go` (no prefix) | inline | inline | inline | inline |
| **Output test** | `output_test.go` | inline | inline | inline | inline |

**Rules:**
- Indicator must be **exactly two levels** below `indicators/`: `indicators/<group>/<indicator>/`.
- Use **full descriptive names**, never short mnemonics (`simple-moving-average/` not `sma/`).
- Four reserved folder names (`core/`, `common/`, `custom/`, `factory/`) must never be used as author names.
- **Go:** author folder must contain `doc.go` with package declaration and godoc comment.
- **Python:** every directory (author, indicator) must contain `__init__.py`. Indicator-level `__init__.py` must re-export class, output enum, params, and default_params.
- **Rust:** indicator folder must contain `mod.rs` re-exporting `mod <name>; pub use <name>::*;`. Register in group's `mod.rs`.
- **Zig:** register in barrel file `zig/src/indicators/indicators.zig` (both `pub const` export and comptime test inclusion).
- Auxiliary files (params, output) drop the indicator-name prefix. Only the main file and test keep it.

---

## 2. Import Patterns

| Language | Entities import | Core import |
|----------|----------------|-------------|
| **Go** | `"zpano/entities"` | `"zpano/indicators/core"` |
| **TS** | `from '../../entities/bar'` (relative, per-entity) | `from '../../core/line-indicator'` (relative, per-file) |
| **Python** | `from ....entities.bar import Bar` (relative, per-entity) | `from ...core.line_indicator import LineIndicator` (relative, per-file) |
| **Zig** | `@import("entities")` (single barrel, build.zig module) | `@import("../../core/indicator.zig")` (relative file paths) |
| **Rust** | `use crate::entities::bar::Bar` (absolute crate paths) | `use crate::indicators::core::indicator::Indicator` (absolute crate paths) |

**Rules:**
- **Go:** NEVER import `"zpano/entities/bar"` — the `entities` package is flat.
- **TS:** Never barrel imports from entities (`../../entities`). Import each entity file individually.
- **Python:** Entities are 4 levels up (`....entities`), core is 3 levels up (`...core`).
- **Zig:** NEVER import individual entity modules (`@import("bar")`). Always use the `entities` barrel.
- **Rust:** Always `use crate::` prefix. Never `use super::` for cross-module imports.

---

## 3. Naming Conventions

| Aspect | Go | TS | Python | Zig | Rust |
|--------|----|----|--------|-----|------|
| **Type** | `SimpleMovingAverage` | `SimpleMovingAverage` | `SimpleMovingAverage` | `SimpleMovingAverage` | `SimpleMovingAverage` |
| **Params type** | (fields on struct) | `SimpleMovingAverageParams` | `SimpleMovingAverageParams` | (fields on struct) | `SimpleMovingAverageParams` |
| **Output enum** | `Output` (bare) | `SimpleMovingAverageOutput` | `SimpleMovingAverageOutput` | `SimpleMovingAverageOutput` | `SimpleMovingAverageOutput` |
| **Output members** | `Value` (bare, iota) | `VALUE` (0-based) | `VALUE` (IntEnum 0-based) | `value` (enum(u8) 1-based) | `Value` (#[repr(u8)] 1-based) |
| **Mnemonic fn** | `Mnemonic(params)` method | `<camelCase>Mnemonic(params)` | `<snake_case>_mnemonic(params)` | `mnemonic(params)` | `mnemonic(params)` |
| **Identifier** | `PascalCase` | `PascalCase` | `UPPER_SNAKE_CASE` | `snake_case` | `PascalCase` |

**Rules:**
- Go output type/consts use **bare names** (`Output`, `Value`) — package provides scoping; long-form would stutter.
- Multi-output Go: strip indicator-name prefix, keep descriptive suffix.
- Banned abbreviations in local variables: `idx`, `tmp`, `res`, `sig`, `val`, `prev`, `avg`, `mult`, `buf`, `param`, `hist` — use long forms.

---

## 4. Structural Patterns

| Aspect | Go | TS | Python | Zig | Rust |
|--------|----|----|--------|-----|------|
| **Base** | Embed `core.LineIndicator` | `extends LineIndicator` | `class X(Indicator)` + LineIndicator mixin | `LineIndicator` field + delegation | Implement `Indicator` trait, use `LineIndicator` helper |
| **Concurrency** | `sync.RWMutex` field named `mu` | None | None | None | None |
| **Component sentinel** | `if bc == 0` (zero-value) | `if (bc === undefined)` | `if bc is None` | `if (bc == null)` on `?BarComponent` | `if bc.is_none()` on `Option<BarComponent>` |
| **Constructor validation** | Return `error` | `throw new Error(...)` | `raise ValueError(...)` | Return `error` | `panic!` or return `Result` |
| **Update method** | `Update(sample float64)` | `update(sample: number)` | `update(self, sample: float)` | `update(self, sample: f64)` | `update(&mut self, sample: f64)` |
| **Primed check** | `IsPrimed() bool` | `get isPrimed(): boolean` | `@property is_primed` | `is_primed() bool` | `is_primed(&self) -> bool` |
| **Output values** | `Value() float64` (per-output getter) | `get value(): number` | `@property value` | `value() f64` | `value(&self) -> f64` |

---

## 5. Component Sentinel & DefaultParams

**Component sentinel resolution order:**
1. Resolve zero/undefined/None/null → default component.
2. Build component functions from resolved values.
3. Build mnemonic using `ComponentTripleMnemonic` with resolved values (Go) or raw values (TS — handles internally).
4. Initialize `LineIndicator` with mnemonic, description, component functions.

**DefaultParams rules:**
- Every indicator must export `DefaultParams()` (Go) / `defaultParams()` (TS) / `default_params()` (Python) / struct defaults (Zig) / `Default::default()` (Rust).
- Dual-variant indicators: export both `DefaultLengthParams()` and `DefaultSmoothingFactorParams()`.
- No-params indicators: still export `DefaultParams()` returning empty struct/object.
- Component fields are **omitted** from defaults (zero/undefined/None/null already means "use default").

---

## 6. Mnemonic Format

- Prefix: lowercase, typically 2-7 chars, digits and leading `+`/`-` allowed.
- Format: `<prefix>(<param1>, <param2>, ...<componentSuffix>)`.
- Configurable-component indicators **must** append `ComponentTripleMnemonic(bc, qc, tc)`.
- Hardcoded-component indicators omit the suffix (e.g., `bop`, `ad`, `psar(0.02,0.2)`).
- Default components are **omitted** from mnemonic.
- Same prefix in all 5 languages (cross-language parity).

---

## 7. Metadata

- Must use `BuildMetadata` / `buildMetadata` / `build_metadata` helper — never construct `Metadata` literals directly.
- Metadata `identifier` field must be the correct enum variant.
- Per-output metadata carries `kind` (from output enum value), `shape` (from descriptor registry), `mnemonic`, `description`.
- Outputs count and order must match the descriptor row.

---

## 8. Registration Checklist

### 8a. Identifier (all 5 languages)

Register in `core/identifier` in all 5 languages simultaneously.

| Language | File | Naming | Numbering |
|----------|------|--------|-----------|
| Go | `go/indicators/core/identifier.go` | `PascalCase` | `iota + 1` (1-based) |
| TS | `ts/indicators/core/indicator-identifier.ts` | `PascalCase` | auto (0-based) |
| Python | `py/indicators/core/identifier.py` | `UPPER_SNAKE_CASE` | explicit (0-based) |
| Zig | `zig/src/indicators/core/identifier.zig` | `snake_case` | explicit (0-based) |
| Rust | `rs/src/indicators/core/identifier.rs` | `PascalCase` | explicit (0-based) |

**Grouping rules:**
- Identifiers are grouped by author with `// ──` / `# ──` comment dividers.
- "common" first, author groups alphabetical, "custom" last.
- **Append** new member at end of its author group — do not re-sort alphabetically.
- Assign the next sequential number.
- If new author group needed, insert in alphabetical order between existing groups.
- **Go:** also update private string-constant block, `String()` switch, `UnmarshalJSON()` switch,
  and all four test tables in `identifier_test.go` (String, IsKnown, MarshalJSON, UnmarshalJSON).
- **Zig:** also update `asStr` and `fromStr` match tables.
- **Rust:** also update `as_str` and `from_str` match arms.

### 8b. Descriptor (all 5 languages)

Register in `core/descriptors.{go,ts,py,zig,rs}`. Missing descriptor causes `BuildMetadata` to panic at runtime.
Outputs order must match the `OutputText[]` passed by `metadata()`.

**Grouping rules (same as identifiers):**
- Descriptor entries are grouped by author with `// ──` / `# ──` comment dividers.
- "common" first, author groups alphabetical, "custom" last.
- **Append** new entry at end of its author group — do not re-sort alphabetically within a group.
- If a new author group is needed, insert a new divider in alphabetical order between existing groups.

### 8c. Factory (all 5 languages)

Register in `factory/factory.{go,ts,py,zig,rs}`.

### 8d. icalc settings.json

- Add entry with default params using **camelCase** identifier string.
- Must run icalc in all 5 languages to verify no crash.
- **Python-specific:** also add entry to `_IDENTIFIER_MAP` dict in `py/cmd/icalc/main.py`.

---

## 9. Output Enum Conventions

| Language | Start value | Member style | Type annotation |
|----------|-------------|--------------|-----------------|
| Go | `iota` (0-based per const block) | `PascalCase` bare | `Output` (bare type) |
| TS | 0 | `UPPER_SNAKE_CASE` | `<Name>Output` |
| Python | 0 (`IntEnum`) | `UPPER_SNAKE_CASE` | `<Name>Output` |
| Zig | 1 (`enum(u8)`) | `snake_case` | `<Name>Output` |
| Rust | 1 (`#[repr(u8)]`) | `PascalCase` | `<Name>Output`, derive `Debug, Clone, Copy, PartialEq, Eq` |

---

## 10. Test Conventions

- Mnemonic sub-tests must cover: all components zero, each component set individually, and combinations of two.
- **Go:** metadata test must check `Identifier`, `Outputs[0].Shape`, `Outputs[0].Mnemonic`.
- **Go:** `sync.RWMutex` field must always be named `mu`, always `RWMutex` (never plain `Mutex`).
- **Go:** receiver naming: compound type → `s`; simple type → first letter lowercased.
- **Python:** use `delta=` parameter in `assertAlmostEqual`, absolute imports in tests.
- **Rust:** tolerance typically `1e-8`; test data as `const` arrays; NaN checks with `.is_nan()`.

---

## 11. Language-Specific Rules

### Go
- `LineIndicator` must be assigned **after** struct creation (method reference requires existing pointer).
- All methods on a type must use the same receiver name.

### Zig
- `fixSlices()` must be implemented for indicators storing mnemonic/description in owned `[N]u8` buffers.
- Mnemonic buffer: `[64]u8`, description buffer: `[128]u8`, built via `std.fmt.bufPrint`.
- `OutputArray` is stack-based (max 9 outputs, no heap).
- `getMetadata` takes `*Metadata` out-pointer (not return value).
- vtable pattern: `const vtable = Indicator.GenVTable(Self)` + `pub fn indicator()`.

### Rust
- Component functions stored as `fn(&Bar) -> f64` function pointers (not closures).
- Extract sample first, then call `self.update()` — cannot use closure delegation (borrow checker).
- Metadata stored on struct, constructed at init time.
- `Output` type is `Vec<Box<dyn Any>>`.

---

## 12. Cross-Language Parity

- Same identifier enum set across all 5 languages (same count, same names).
- Same mnemonic prefix in all 5 languages.
- Same descriptor row in all 5 languages.
- Same factory case in all 5 languages.
- Same `settings.json` entry exercised in all 5 languages.
- Results must match to **13+ decimal places** across implementations.

---

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
