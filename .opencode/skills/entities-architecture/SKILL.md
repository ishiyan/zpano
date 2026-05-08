# Skill: entities-architecture

# Entities Architecture & Cross-Language Reference

This document describes the design decisions, data types, and conventions for the
**entities** module across all five languages. It is the reference for porting
entity code and for indicator authors who consume entity types.

> **How to use this document:** Do not attempt to apply all rules at once.
> 1. Identify your target language (Go, TS, Python, Zig, or Rust).
> 2. Read only that language's column in each table.
> 3. The cross-language conventions section defines the rules that must hold
>    across all ports — consult it only when porting between languages.

## Scope

The rules in this document apply **only to the `entities/` folder** within each
language directory (`go/entities/`, `ts/entities/`, `py/entities/`,
`zig/src/entities/`, `rs/src/entities/`). The `indicators/` folder has its own
architecture skill.

## Module Contents

The entities module defines four data types for financial trading data, three
component enums for field extraction, and supporting utilities.

### Module Root / Barrel Files

Each language provides a root file that re-exports all entity types:

| Language | Root file | Pattern |
|----------|-----------|---------|
| **Go** | *(implicit — package-level exports)* | `entities.Bar`, `entities.Quote`, etc. |
| **TypeScript** | *(no barrel — import from individual files)* | `import { Bar } from './bar'` |
| **Python** | `__init__.py` | Re-exports all 4 types + 3 component modules |
| **Zig** | `entities.zig` | `pub const bar = @import("bar")` + convenience aliases (`Bar`, `Scalar`, etc.) |
| **Rust** | `mod.rs` | `pub mod bar; pub mod quote;` etc. (no type aliases) |

The Zig `entities.zig` barrel mirrors the `indicators/indicators.zig` pattern:
CLIs import `const entities = @import("entities")` and use `entities.Bar`,
`entities.Scalar`, etc.

### Entity Types

| Type | Fields | Computed Methods |
|------|--------|-----------------|
| **Bar** | time, open, high, low, close, volume | `isRising`, `isFalling`, `median` (HL/2), `typical` (HLC/3), `weighted` (HLCC/4), `average` (OHLC/4) |
| **Quote** | time, bid_price, ask_price, bid_size, ask_size | `mid` ((bid+ask)/2), `weighted` ((bid*bs+ask*as)/(bs+as)), `weightedMid` ((bid*as+ask*bs)/(bs+as)), `spreadBp` (20000*(ask-bid)/(ask+bid)) |
| **Trade** | time, price, volume | *(none)* |
| **Scalar** | time, value | *(none)* |

**Zero-denominator guards:** Quote's `weighted`, `weightedMid`, and `spreadBp`
return `0.0` when the denominator is zero.

**IsRising/IsFalling:** `isRising` is `open < close` (strictly less, not <=).
`isFalling` is `close < open`. A bar where `open == close` is neither rising
nor falling.

### Component Enums

Each component enum selects which field/computation to extract from its entity
type. Component enums are used by indicators to support configurable input
sources (e.g., an SMA can operate on close, median, volume, etc.).

| Enum | Values | Default | Default Mnemonic |
|------|--------|---------|-----------------|
| **BarComponent** | Open, High, Low, Close, Volume, Median, Typical, Weighted, Average (9) | Close | `c` |
| **QuoteComponent** | Bid, Ask, BidSize, AskSize, Mid, Weighted, WeightedMid, SpreadBp (8) | Mid | `ba/2` |
| **TradeComponent** | Price, Volume (2) | Price | `p` |

Each component enum exports:
1. **Default constant** (`DefaultBarComponent`, etc.)
2. **Component value function** — returns a function/closure that extracts the
   named component from an entity instance.
3. **Component mnemonic function** — returns a short string code for display.

### Component Display Methods

Each component has two display representations:

| Purpose | Method | Example (BarComponent.Close) | Unknown value |
|---------|--------|------------------------------|---------------|
| **Full name** (JSON, debug) | `String()` / `str()` | `"close"` | Go: `"unknown"`, others: N/A (exhaustive) |
| **Short code** (chart labels) | `Mnemonic()` / mnemonic function | `"c"` | Go: `"unknown"`, Py: `"??"` |

### Language-Specific Types (Do Not Port)

The following types exist only in Go or TypeScript. They are either dead code
or indicators-module infrastructure that will be handled separately when/if
those modules are ported. **Do not port these to py/zig/rs as part of entities.**

**Dead code** (unused anywhere — not by indicators, cmd, or tests of other modules):

- **`Temporal`** interface (Go) — `DateTime() time.Time`, never implemented by
  any entity struct, never referenced outside its definition file.
- **`TemporalEntityKind`** (TS) — string enum (`'bar'`, `'quote'`, `'trade'`,
  `'scalar'`), never imported by indicators or cmd.
- **`TemporalEntity`** (TS) — union type `Bar | Scalar | Quote | Trade`,
  never imported by indicators or cmd.

**Indicators infrastructure** (used by Go indicators/cmd, port with indicators):

- **JSON marshal/unmarshal** on component enums — supports the JSON-based
  factory and `iconf` CLI tool. Every indicator `Output` enum, `Identifier`,
  and core enums also implement this pattern.
- **`IsKnown()`** on component enums — used by indicators for iteration and
  validation (`for id := ...; id.IsKnown(); id++`). Unnecessary in languages
  with exhaustive pattern matching (Zig, Rust).

## Cross-Language Conventions

The rules below are grouped by concern. When implementing in a single language,
read only your language's column in each table.

### File Layout

| Language | Entity files | Component files | Barrel / Root | Test files |
|----------|-------------|-----------------|---------------|------------|
| **Go** | `bar.go`, `quote.go`, `trade.go`, `scalar.go` | `barcomponent.go`, `quotecomponent.go`, `tradecomponent.go` | *(package-level)* | `*_test.go` |
| **TypeScript** | `bar.ts`, `quote.ts`, `trade.ts`, `scalar.ts` | `bar-component.ts`, `quote-component.ts`, `trade-component.ts` | *(none)* | `*.spec.ts` |
| **Python** | `bar.py`, `quote.py`, `trade.py`, `scalar.py` | `bar_component.py`, `quote_component.py`, `trade_component.py` | `__init__.py` | `test_entities.py` |
| **Zig** | `bar.zig`, `quote.zig`, `trade.zig`, `scalar.zig` | `bar_component.zig`, `quote_component.zig`, `trade_component.zig` | `entities.zig` | Tests at bottom of source |
| **Rust** | `bar.rs`, `quote.rs`, `trade.rs`, `scalar.rs` | `bar_component.rs`, `quote_component.rs`, `trade_component.rs` | `mod.rs` | `#[cfg(test)]` at bottom |

### Field Naming

| Field | Go | TypeScript | Python | Zig | Rust |
|-------|-----|-----------|--------|-----|------|
| Time | `Time time.Time` | `time: Date` | `time: datetime` | `time: i64` (epoch) | `time: i64` (epoch) |
| Bid price | `Bid float64` | `bidPrice: number` | `bid_price: float` | `bid_price: f64` | `bid_price: f64` |
| Ask price | `Ask float64` | `askPrice: number` | `ask_price: float` | `ask_price: f64` | `ask_price: f64` |

Go uses `time.Time`, TypeScript uses `Date`, Python uses `datetime.datetime`.
Zig and Rust use a raw `i64` epoch timestamp (not the `DateTime` struct from
the daycounting module, which is used only there).
Go uses short field names for all exported struct fields (e.g., `Bid`, `Ask`,
`Open`, `High`, `Volume`) because the enclosing type name already provides
context (e.g., `Quote.Bid` is unambiguous). All other languages use descriptive
compound names (`bid_price`, `ask_price`, `bid_size`, `ask_size`).

### Enum Numbering

| Language | Start value | "Not set" sentinel | Default resolution |
|----------|------------|-------------------|-------------------|
| **Go** | `iota + 1` (1-based) | `0` (zero value) | `if bc == 0 { bc = DefaultBarComponent }` |
| **TypeScript** | `0` | `undefined` | `=== undefined` check in setter |
| **Python** | `0` (IntEnum) | `None` (Optional) | `if bc is None: bc = DEFAULT_BAR_COMPONENT` |
| **Zig** | `0` (enum(u8)) | `null` (?T optional) | `bc orelse default_bar_component` |
| **Rust** | `0` (#[repr(u8)]) | `None` (Option\<T\>) | `bc.unwrap_or(DEFAULT_BAR_COMPONENT)` |

**Critical for indicator porting:** Go's zero-sentinel pattern (`if bc == 0`)
maps to Optional/Option/?T in other languages. Do NOT change existing 0-based
enums to 1-based. The indicator params structs use the language-idiomatic
optional type for component fields.

### Component Value Function

| Language | Signature | Unknown handling |
|----------|-----------|-----------------|
| **Go** | `BarComponentFunc(c BarComponent) (BarFunc, error)` | Returns error |
| **TypeScript** | `barComponentValue(c: BarComponent): (b: Bar) => number` | Returns close (default fallback) |
| **Python** | `bar_component_value(c: BarComponent) -> Callable[[Bar], float]` | Returns close (default fallback) |
| **Zig** | `componentValue(c: BarComponent) *const fn(Bar) f64` | Exhaustive switch, no unknown case |
| **Rust** | `component_value(c: BarComponent) -> fn(&Bar) -> f64` | Exhaustive match, no unknown case |

### Default Constants

| Language | Bar | Quote | Trade |
|----------|-----|-------|-------|
| **Go** | `DefaultBarComponent = BarClosePrice` | `DefaultQuoteComponent = QuoteMidPrice` | `DefaultTradeComponent = TradePrice` |
| **TypeScript** | `DefaultBarComponent = BarComponent.CLOSE` | `DefaultQuoteComponent = QuoteComponent.MID` | `DefaultTradeComponent = TradeComponent.PRICE` |
| **Python** | `DEFAULT_BAR_COMPONENT = BarComponent.CLOSE` | `DEFAULT_QUOTE_COMPONENT = QuoteComponent.MID` | `DEFAULT_TRADE_COMPONENT = TradeComponent.PRICE` |
| **Zig** | `default_bar_component: BarComponent = .close` | `default_quote_component: QuoteComponent = .mid` | `default_trade_component: TradeComponent = .price` |
| **Rust** | `DEFAULT_BAR_COMPONENT: BarComponent = BarComponent::Close` | `DEFAULT_QUOTE_COMPONENT: QuoteComponent = QuoteComponent::Mid` | `DEFAULT_TRADE_COMPONENT: TradeComponent = TradeComponent::Price` |

### Mnemonic Tables

**BarComponent:**

| Value | String | Mnemonic |
|-------|--------|----------|
| Open | `open` | `o` |
| High | `high` | `h` |
| Low | `low` | `l` |
| Close | `close` | `c` |
| Volume | `volume` | `v` |
| Median | `median` | `hl/2` |
| Typical | `typical` | `hlc/3` |
| Weighted | `weighted` | `hlcc/4` |
| Average | `average` | `ohlc/4` |

**QuoteComponent:**

| Value | String | Mnemonic |
|-------|--------|----------|
| Bid | `bid` | `b` |
| Ask | `ask` | `a` |
| BidSize | `bidSize` | `bs` |
| AskSize | `askSize` | `as` |
| Mid | `mid` | `ba/2` |
| Weighted | `weighted` | `(bbs+aas)/(bs+as)` |
| WeightedMid | `weightedMid` | `(bas+abs)/(bs+as)` |
| SpreadBp | `spreadBp` | `spread bp` |

**TradeComponent:**

| Value | String | Mnemonic |
|-------|--------|----------|
| Price | `price` | `p` |
| Volume | `volume` | `v` |

## Design Decisions Log

| Decision | Rationale |
|----------|-----------|
| Keep 0-based enums in py/zig/rust, use Optional for "not set" | Avoids breaking changes to existing code and tests. Each language has a natural "absent" type (`None`, `null`, `None`) that is cleaner than Go's zero-sentinel. |
| Go uses 1-based enums (iota+1) with zero = "not set" | Idiomatic Go pattern for optional enum fields in structs where zero value serves as sentinel. |
| TS uses `undefined` for "not set" (not `null`) | TS convention: `undefined` means "not provided", `null` means "intentionally empty". Component params are optional object properties. |
| No Temporal interface in py/zig/rust | Dead code in Go — never implemented or referenced. No reason to port. |
| No JSON marshal in py/zig/rust | Go indicators infrastructure. Port alongside the indicators module, not entities. |
| `String()` vs `Mnemonic()` separation | `String()` returns full words for serialization/debug. `Mnemonic()` returns short codes for chart labels. Different audiences, different needs. |
| Unknown component: Go returns `"unknown"`, Py returns `"??"` | Documented convention per AGENTS.md. Zig/Rust use exhaustive switches so unknown is impossible. |
| Quote fields: Go `Bid`/`Ask` vs others `bid_price`/`ask_price` | Go relies on type context for clarity. Other languages use descriptive names per their conventions. Both are correct. |

## Indicator Integration Points

When porting indicators, the key entity features consumed are:

1. **Component function types** (`BarFunc`, `QuoteFunc`, `TradeFunc`) — stored
   in the `LineIndicator` base to route `updateBar`/`updateQuote`/`updateTrade`
   calls to the core `update(sample)` method.
2. **Default component constants** — used in indicator constructors to resolve
   "not set" components to defaults.
3. **Component mnemonic functions** — used by `componentTripleMnemonic` to build
   indicator mnemonics like `sma(14, hl/2)`.
4. **Entity structs** — passed to `updateBar`/`updateQuote`/`updateTrade` methods.

The indicator `core/` module imports entities for these four purposes. No other
entity features (JSON, Temporal, TemporalEntityKind) are consumed by indicators.
