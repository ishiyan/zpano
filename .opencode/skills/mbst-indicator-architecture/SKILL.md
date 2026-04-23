---
name: mbst-indicator-architecture
description: Architecture reference for the MBST C# trading indicators library (Mbst.Trading.Indicators). Load when converting MBST indicators to zpano or understanding the MBST source code.
---

# MBST Indicator Architecture Reference

This document describes the architecture of the **MBST C# trading indicators library**
(`Mbst.Trading.Indicators` namespace). It is the source codebase from which indicators are
converted to the zpano multi-language library.

Source files are located in `mbst-to-convert/`.

---

## Table of Contents

1. [Namespace & Assembly](#namespace--assembly)
2. [Type Hierarchy](#type-hierarchy)
3. [Root Interface: IIndicator](#root-interface-iindicator)
4. [Abstract Base: Indicator](#abstract-base-indicator)
5. [Line Indicators](#line-indicators)
6. [Band Indicators](#band-indicators)
7. [Heatmap Indicators](#heatmap-indicators)
8. [Drawing Indicators](#drawing-indicators)
9. [Facade Pattern](#facade-pattern)
10. [Data Types](#data-types)
11. [OhlcvComponent Enum](#ohlcvcomponent-enum)
12. [Serialization & Annotations](#serialization--annotations)
13. [Overwritable History Interfaces](#overwritable-history-interfaces)
14. [Color Interpolation](#color-interpolation)
15. [Files to Ignore During Conversion](#files-to-ignore-during-conversion)

---

## Namespace & Assembly

All types live in:

```csharp
namespace Mbst.Trading.Indicators
```

Source abstractions are in `mbst-to-convert/Abstractions/` (18 files).
Concrete indicators are in `mbst-to-convert/<author>/<indicator>/`.

---

## Type Hierarchy

```
IIndicator                              (root interface)
├── ILineIndicator                      (single scalar output)
│   ├── LineIndicator (abstract)        (standard implementation)
│   └── LineIndicatorFacade (sealed)    (proxy for multi-output sources)
├── IBandIndicator                      (two-value output)
│   ├── BandIndicator (abstract)        (standard implementation)
│   └── BandIndicatorFacade (sealed)    (proxy for multi-output sources)
├── IHeatmapIndicator                   (brush/intensity array output)
│   └── HeatmapIndicatorFacade (sealed) (proxy for multi-output sources)
├── IDrawingIndicator                   (WPF rendering — ignore)
│
Indicator (abstract)                    (shared base for all concrete indicators)
├── LineIndicator
└── BandIndicator
```

---

## Root Interface: IIndicator

```csharp
public interface IIndicator
{
    string Name { get; }         // Identifies the indicator type
    string Moniker { get; }      // Identifies a parameterized instance (like zpano Mnemonic)
    string Description { get; }  // Describes the indicator
    bool IsPrimed { get; }       // Whether the indicator has enough data
    void Reset();                // Resets indicator state
}
```

**Mapping to zpano:** `Name` -> `Metadata().Type`, `Moniker` -> `Metadata().Mnemonic`,
`Description` -> `Metadata().Description`. `Reset()` is dropped (zpano indicators are
immutable; create a new instance instead).

---

## Abstract Base: Indicator

```csharp
[DataContract]
public abstract class Indicator : IIndicator
{
    [DataMember] protected string name;
    [DataMember] protected string moniker;
    [DataMember] protected string description;
    [DataMember] protected bool primed;
    [DataMember] protected OhlcvComponent ohlcvComponent;
    [DataMember] protected readonly object updateLock = new object();

    protected Indicator(string name, string description,
        OhlcvComponent ohlcvComponent = OhlcvComponent.ClosingPrice)
    {
        this.name = name;
        this.description = description;
        this.ohlcvComponent = ohlcvComponent;
    }

    public bool IsPrimed { get { lock (updateLock) { return primed; } } }
    public abstract void Reset();
}
```

Key points:
- **Default OHLCV component** is `ClosingPrice` (most indicators). Some override this
  (e.g., CenterOfGravityOscillator defaults to `MedianPrice`).
- **Thread safety** via `updateLock` — zpano drops this (single-threaded update model).
- **`primed` field** set by concrete subclasses when enough samples are received.
- All `[DataContract]`/`[DataMember]` annotations are for WCF serialization — ignore.

---

## Line Indicators

### ILineIndicator

```csharp
public interface ILineIndicator : IIndicator
{
    double Update(double sample);    // Core algorithm — scalar in, scalar out
    Scalar Update(Scalar sample);    // Delegates to Update(double)
    Scalar Update(Ohlcv sample);     // Extracts component, delegates to Update(double)
}
```

### LineIndicator (abstract)

```csharp
[DataContract]
public abstract class LineIndicator : Indicator, ILineIndicator
{
    protected LineIndicator(string name, string description,
        OhlcvComponent ohlcvComponent = OhlcvComponent.ClosingPrice)
        : base(name, description, ohlcvComponent) {}

    public abstract double Update(double sample);                    // IMPLEMENT THIS

    public virtual Scalar Update(double sample, DateTime dateTime)   // Convenience
    { return new Scalar(dateTime, Update(sample)); }

    public virtual Scalar Update(Scalar scalar)                      // Delegates
    { return new Scalar(scalar.Time, Update(scalar.Value)); }

    public virtual Scalar Update(Ohlcv ohlcv)                       // Delegates + component extraction
    { return new Scalar(ohlcv.Time, Update(ohlcv.Component(ohlcvComponent))); }

    public override string ToString() { ... }                        // Debug string — ignore
}
```

**Only `Update(double)` contains the algorithm.** All other overloads are boilerplate
delegation. In zpano, this maps to `LineIndicator` embedding (Go) or
`extends LineIndicator` (TS), which auto-generates the entity update methods.

---

## Band Indicators

### IBandIndicator

```csharp
public interface IBandIndicator : IIndicator
{
    Band Update(double sample, DateTime dateTime);  // Core algorithm
    Band Update(Scalar sample);                     // Delegates
    Band Update(Ohlcv sample);                      // Delegates + component extraction
}
```

### BandIndicator (abstract)

```csharp
[DataContract]
public abstract class BandIndicator : Indicator, IBandIndicator
{
    protected BandIndicator(string name, string description,
        OhlcvComponent ohlcvComponent = OhlcvComponent.ClosingPrice)
        : base(name, description, ohlcvComponent) {}

    public abstract Band Update(double sample, DateTime dateTime);   // IMPLEMENT THIS

    public virtual Band Update(Scalar sample)                        // Delegates
    { return Update(sample.Value, sample.Time); }

    public virtual Band Update(Ohlcv sample)                         // Delegates + component extraction
    { return Update(sample.Component(ohlcvComponent), sample.Time); }
}
```

Note: `BandIndicator.Update(double, DateTime)` takes a `DateTime` parameter that
`LineIndicator.Update(double)` does not. In zpano, band outputs are emitted as
`outputs.Band{Upper, Lower}` entries in the `Output` array; see "Band Output
Semantics When Wrapping MBST's `Band`" in the conversion skill for the
upper/lower assignment convention.

---

## Heatmap Indicators

### IHeatmapIndicator

```csharp
public interface IHeatmapIndicator : IIndicator
{
    Heatmap Update(Scalar sample);
    Heatmap Update(Ohlcv sample);
    double MinParameterValue { get; }   // Ordinate bounds for rendering
    double MaxParameterValue { get; }
}
```

There is **no abstract `HeatmapIndicator` base class** — only the facade exists.
Concrete heatmap indicators implement `IHeatmapIndicator` directly.

---

## Drawing Indicators

```csharp
public interface IDrawingIndicator : IIndicator
{
    void Update(double value, DateTime dateTime);
    void Update(Scalar sample);
    void Update(Ohlcv sample);
    void ValueBounds(ref double lower, ref double upper);
    void Render(Dispatcher dispatcher, Func<DrawingContext> renderOpen, ...);
}
```

**Entirely WPF/UI-specific.** Ignore completely during conversion — zpano does not
have rendering indicators.

---

## Facade Pattern

MBST exposes individual outputs of multi-output indicators via **facade classes**:

- `LineIndicatorFacade` — wraps a `Func<double> getValue` to expose one scalar output
- `BandIndicatorFacade` — wraps `Func<double> getFirstValue` + `Func<double> getSecondValue`
- `HeatmapIndicatorFacade` — wraps `Func<Brush> getBrush`

### How Facades Work

A multi-output indicator (e.g., one that computes both a "value" and a "trigger") creates
facade instances in its constructor:

```csharp
// Inside the indicator constructor:
ValueIndicator = new LineIndicatorFacade("Value", moniker, description,
    () => IsPrimed, () => value);
TriggerIndicator = new LineIndicatorFacade("Trigger", moniker, description,
    () => IsPrimed, () => valuePrevious);
```

Consumers call `Update()` on the *source* indicator, then read from the facade.
The facade's own `Update()` ignores its input and just returns `getValue()`.

### Why Zpano Drops Facades

Zpano's `Output` array replaces facades entirely. Multi-output indicators return all
outputs in their `Output` / `IndicatorOutput` array (e.g., `output[Value]` and
`output[Trigger]`), indexed by the per-indicator output enum.

**When converting: ignore all facade creation, facade properties, and facade tests.**

---

## Data Types

### Scalar

```csharp
public class Scalar { DateTime Time; double Value; }
```

Maps to zpano `entities.Scalar` (Go) / `Scalar` (TS).

### Band

```csharp
public class Band
{
    DateTime Time;
    double FirstValue;   // default NaN
    double SecondValue;  // default NaN
    bool IsEmpty => double.IsNaN(FirstValue) || double.IsNaN(SecondValue);
}
```

Maps to zpano `outputs.Band` (Go) / `Band` (TS). Note zpano Band uses `Upper`/`Lower`
naming, not `FirstValue`/`SecondValue`.

### Heatmap

```csharp
public sealed class Heatmap
{
    Brush Brush;           // WPF Brush — will need redesign for zpano
    double[] Intensity;
    DateTime Time;
    bool IsEmpty => null == Brush;
}
```

The `Brush` field is WPF-specific. Zpano heatmap representation will differ.

### Ohlcv

Referenced but not in Abstractions — represents a bar with Open, High, Low, Close,
Volume properties plus a `Component(OhlcvComponent)` method that extracts the requested
component value.

---

## OhlcvComponent Enum

The `OhlcvComponent` enum is referenced but its definition is not in the Abstractions
folder. Known members from usage:

| MBST OhlcvComponent | Description | Zpano BarComponent |
|---------------------|-------------|--------------------|
| `ClosingPrice` | Close price (default) | `Close` (default) |
| `OpeningPrice` | Open price | `Open` |
| `HighPrice` | High price | `High` |
| `LowPrice` | Low price | `Low` |
| `MedianPrice` | (High + Low) / 2 | `Median` |
| `TypicalPrice` | (High + Low + Close) / 3 | `Typical` |
| `WeightedPrice` | (High + Low + 2*Close) / 4 | `Weighted` |
| `Volume` | Volume | `Volume` |

**Key difference from zpano:** MBST uses a single `OhlcvComponent` for all entity types
(bars only). Zpano uses a triple: `BarComponent`, `QuoteComponent`, `TradeComponent` —
supporting bars, quotes, and trades as separate entity types.

---

## Serialization & Annotations

MBST uses WCF `DataContract` serialization:

```csharp
[DataContract]   // on classes
[DataMember]     // on fields
```

**Ignore all of these during conversion.** Zpano does not serialize indicator state.

---

## Overwritable History Interfaces

Three interfaces extend the base indicator interfaces with history mutation:

- `ILineIndicatorWithOverwritableHistory : ILineIndicator`
- `IBandIndicatorWithOverwritableHistory : IBandIndicator`
- `IHeatmapIndicatorWithOverwritableHistory : IHeatmapIndicator`

These allow overwriting previously emitted values (for indicators that refine past
outputs). **Ignore during conversion** — zpano does not support history overwriting.

---

## Color Interpolation

Two files for WPF heatmap color rendering:

- `ColorInterpolationType.cs` — enum: `Linear`, `Quadratic`, `Cubic`, `InverseQuadratic`, `InverseCubic`
- `ColorInterpolation.cs` — static utility for interpolating WPF colors/brushes

**Entirely WPF-specific. Ignore during conversion.**

---

## Files to Ignore During Conversion

| File | Reason |
|------|--------|
| `IDrawingIndicator.cs` | WPF rendering — no zpano equivalent |
| `ILineIndicatorWithOverwritableHistory.cs` | History overwriting — not supported |
| `IBandIndicatorWithOverwritableHistory.cs` | History overwriting — not supported |
| `IHeatmapIndicatorWithOverwritableHistory.cs` | History overwriting — not supported |
| `ColorInterpolation.cs` | WPF color utilities |
| `ColorInterpolationType.cs` | WPF color enum |
| `LineIndicatorFacade.cs` | Facade pattern — replaced by Output array |
| `BandIndicatorFacade.cs` | Facade pattern — replaced by Output array |
| `HeatmapIndicatorFacade.cs` | Facade pattern — replaced by Output array |

The 9 files above (half of the 18 Abstractions files) can be completely ignored.
The remaining 9 files define the core type hierarchy that informs conversion decisions.
