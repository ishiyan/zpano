---
name: signal-ensemble-architecture
description: Architecture, design decisions, and implementation reference for the zpano signal ensemble library. Load when implementing ensemble methods, porting across languages, or understanding the weighted signal blending system.
---

# Signal Ensemble Architecture

Architecture, algorithms, and implementation reference for the signal ensemble library in zpano. This package provides adaptive weighted blending of multiple independent signal sources, with delayed feedback and online weight learning.

> **Related skills:**
> - `fuzzy-architecture` — underlying fuzzy logic primitives (membership functions, operators, defuzzification)
> - `signals-architecture` — logical signal composition via t-norms/s-norms (threshold, crossover, band, histogram, compose)

## Motivation

### Logical Composition vs Ensemble Voting

The `signals/` module provides **logical composition** of conditions:

```python
# "Buy when RSI oversold AND MACD crosses up" — logical conjunction
buy = signal_and(mu_oversold(rsi, 30), mu_crosses_above(prev_macd, curr_macd, 0))
```

Here, `signal_and` uses the product t-norm: if either condition fails (mu ~ 0), the entire signal collapses to zero. This models **trading rules** — structured conditions where all parts must hold.

**Ensemble voting** solves a different problem: combining multiple **independent estimators** of the same quantity. Each estimator (signal source) produces a confidence in [0, 1], and the ensemble blends them proportionally:

```python
# Five independent buy-confidence estimators, weighted by reliability
confidence = ensemble.blend([rsi_signal, macd_signal, bb_signal, vol_signal, pattern_signal])
```

Here, a single failed estimator doesn't kill the output — it drags the weighted average down proportionally. This models **consensus** across uncorrelated techniques.

| Aspect | Logical Composition (`signals/`) | Ensemble Voting (`signal_ensemble/`) |
|--------|----------------------------------|--------------------------------------|
| Operator | Product t-norm (fuzzy AND) | Weighted average |
| One zero input | Output = 0 (veto) | Output reduced proportionally |
| Purpose | Express trading rules | Blend independent estimators |
| State | Stateless | Stateful (adaptive weights) |
| Learning | None | Online weight updates from feedback |

### Connection to Sugeno Fuzzy Inference

The ensemble's weighted-average defuzzification is mathematically identical to **Sugeno zero-order** fuzzy inference:

```
Sugeno:   output = sum(w_i * z_i) / sum(w_i)
Ensemble: output = sum(signal_i * weight_i) / sum(weight_i)
```

Where `w_i` are rule firing strengths (our signals) and `z_i` are zero-order consequent constants (our weights). Many practitioners arrive at this formula empirically without the fuzzy logic framing — it's such a natural approach to combining expert opinions that it's been independently rediscovered many times.

The key difference from a full Sugeno system: we don't have a formal rule base with antecedent membership functions. Our "rules" are the signal sources themselves, which may be computed by any technique (indicators, ML models, pattern recognition, etc.). The ensemble only sees their [0, 1] outputs.

## Architecture

### Dependency Graph

```
signal_ensemble/        (standalone — zero dependencies on other zpano modules)
    |
    v
(consumers)             (trading systems, strategy engines, backtesting)
```

The signal ensemble module has **zero dependencies** on `fuzzy/`, `signals/`, `indicators/`, or any other zpano module. It is pure math operating on `float` values in [0, 1]. The caller decides what produces the input signals.

### Module Position

```
py/fuzzy/                  # Membership functions, operators, defuzzify
py/signals/                # Logical signal interpretation (threshold, crossover, band, compose)
py/signal_ensemble/        # Weighted blending of multiple signal sources
py/candlestick_patterns/   # Candlestick pattern recognition engine
py/indicators/             # Technical analysis indicators
```

### Cross-Language Folder Names

| Language | Module folder |
|----------|---------------|
| Python | `py/signal_ensemble/` |
| Go | `go/signalensemble/` |
| TypeScript | `ts/signal-ensemble/` |
| Zig | `zig/src/signal_ensemble/` |
| Rust | `rs/src/signal_ensemble/` |

## Delayed Feedback Design

### The Problem

When a signal fires, the outcome isn't immediately known:

```
Bar T:    signals computed, ensemble produces blended confidence
Bar T+1:  earliest trade execution (you couldn't buy at T, only at T+1)
Bar T+2:  earliest 1-bar return from the trade
Bar T+N:  outcome over the desired evaluation horizon
```

The ensemble needs to **pair predictions with outcomes** across a temporal gap.

### Solution: Ring Buffer with Configurable Delay

The `Aggregator` stores recent signal vectors in a ring buffer. When `update(outcome)` is called, it retrieves the signal vector from `feedback_delay` bars ago and updates weights based on how well each signal predicted that outcome.

```python
agg = Aggregator(n_signals=3, method=Method.INVERSE_VARIANCE, feedback_delay=2)

# Bar 1: blend only (no outcome available yet)
c1 = agg.blend([0.8, 0.3, 0.6])   # signals buffered, weights = initial

# Bar 2: blend only (still no outcome for Bar 1)
c2 = agg.blend([0.7, 0.5, 0.4])   # signals buffered, weights = initial

# Bar 3: blend, then provide outcome for Bar 1
c3 = agg.blend([0.9, 0.2, 0.7])
agg.update(outcome=0.8)            # paired with Bar 1 signals [0.8, 0.3, 0.6]
                                    # weights updated based on per-signal accuracy

# Bar 4: blend with updated weights, then provide outcome for Bar 2
c4 = agg.blend([0.6, 0.4, 0.8])
agg.update(outcome=0.3)            # paired with Bar 2 signals [0.7, 0.5, 0.4]
```

### Tolerance for Missing Updates

`update()` is not mandatory every bar. If the caller skips a bar (e.g., no trade was taken, no outcome to report), the corresponding buffered signals are never paired and silently age out of the ring buffer. Weights remain unchanged for that step.

For stateless methods (`FIXED`, `EQUAL`), `update()` is always a no-op.

### Warmup Lifecycle

For morning intraday sessions, the caller can warm up on yesterday's data:

```python
agg = Aggregator(n_signals=5, method=Method.INVERSE_VARIANCE, feedback_delay=2)

# Warm up on yesterday's 390 one-minute bars
# Each tuple: (signals_at_bar_T, outcome_for_bar_T)
agg.warmup(yesterdays_data)

# Weights are now calibrated — start live session
for bar in live_stream:
    signals = compute_signals(bar)
    confidence = agg.blend(signals)
    if have_outcome:
        agg.update(outcome)
```

`warmup(history)` replays `(signals, outcome)` pairs through `blend()` + `update()`, handling the feedback delay internally. The result is identical to having processed those bars live. Each tuple contains the signals observed at bar T and the outcome that evaluates bar T's prediction. The method buffers signals via `blend()` and, once enough bars have accumulated to satisfy the delay, feeds outcomes via `update()` paired with the correct historical signals.

## Aggregation Methods

### Enum: `AggregationMethod`

```python
class AggregationMethod(IntEnum):
    FIXED = 0                   # User-supplied static weights
    EQUAL = 1                   # Uniform 1/n weights
    INVERSE_VARIANCE = 2        # Weight by 1/variance of errors
    EXPONENTIAL_DECAY = 3       # EMA of accuracy
    MULTIPLICATIVE_WEIGHTS = 4  # Hedge algorithm (online learning)
    RANK_BASED = 5              # Weight by rank of rolling accuracy
    BAYESIAN = 6                # Bayesian model averaging
```

### Enum: `ErrorMetric`

```python
class ErrorMetric(IntEnum):
    ABSOLUTE = 0   # |signal_i - outcome|
    SQUARED = 1    # (signal_i - outcome)^2
```

Used by `INVERSE_VARIANCE` and `RANK_BASED` methods. Default: `ABSOLUTE`.

### Method 1: Fixed Weights (Stateless)

**Use case**: Weights determined externally (e.g., from offline optimization, domain expertise).

**Parameters**: `weights: list[float]` (required, normalized to sum to 1.0 in constructor)

**Algorithm**:
```
blend(signals) = sum(weights[i] * signals[i])
update(outcome) = no-op
```

**State**: None (immutable after construction).

### Method 2: Equal Weights (Stateless)

**Use case**: Baseline / no-information prior. Surprisingly competitive (DeMiguel et al. 2009).

**Parameters**: None.

**Algorithm**:
```
blend(signals) = sum(signals[i]) / n
update(outcome) = no-op
```

**State**: None.

### Method 3: Inverse-Variance (Adaptive)

**Use case**: Weight signals by reliability — signals with lower prediction error variance get higher weight.

**Parameters**: `window: int = 50` (rolling window size), `error_metric: ErrorMetric = ABSOLUTE`

**Algorithm**:
```
error_i = |signal_i - outcome|           (ABSOLUTE)
          (signal_i - outcome)^2         (SQUARED)

# Rolling window of errors per signal
errors_i.append(error_i)                 # deque, maxlen = window

# Variance of error window
var_i = variance(errors_i)               # population variance

# Raw weight = inverse variance (with floor to avoid division by zero)
raw_i = 1.0 / max(var_i, epsilon)

# Normalize
weights[i] = raw_i / sum(raw_j for all j)
```

**State per signal**: `deque[float]` of errors, maxlen = window.

**Initial weights**: `1/n` until at least 2 errors have been collected per signal.

**Properties**: Signals that consistently predict well (low variance) get high weight. Signals with erratic accuracy get downweighted. The rolling window allows adaptation to regime changes.

### Method 4: Exponential Decay (Adaptive)

**Use case**: Recent performance matters more than distant history. Adapts faster than inverse-variance to regime changes.

**Parameters**: `alpha: float = 0.1` (decay rate, higher = faster adaptation)

**Algorithm**:
```
accuracy_i = 1.0 - error_i              # where error uses ABSOLUTE metric
ema_i = alpha * accuracy_i + (1 - alpha) * ema_i

# Normalize
weights[i] = max(ema_i, 0) / sum(max(ema_j, 0) for all j)
```

**State per signal**: `float` — EMA of accuracy.

**Initial state**: `ema_i = 0.5` (neutral prior).

**Properties**: Exponential memory — a signal that was great yesterday but terrible today will be rapidly downweighted. The `alpha` parameter controls the effective memory length (~`1/alpha` bars). No explicit window needed.

### Method 5: Multiplicative Weights / Hedge (Online Learning)

**Use case**: Theoretical guarantees from online learning theory. Converges to the best signal in hindsight (regret-bounded).

**Parameters**: `eta: float = 0.5` (learning rate)

**Algorithm**:
```
loss_i = |signal_i - outcome|            # per-signal loss

# Update log-weights
log_w_i -= eta * loss_i

# Normalize via softmax
max_log = max(log_w_j for all j)
weights[i] = exp(log_w_i - max_log) / sum(exp(log_w_j - max_log) for all j)
```

**State per signal**: `float` — log-weight.

**Initial state**: `log_w_i = 0.0` (uniform in log-space = uniform weights).

**Properties**: From the "prediction with expert advice" literature (Freund & Schapire, 1997). The multiplicative update exponentially penalizes poor performers. The regret bound guarantees that over T rounds, the ensemble's cumulative loss is at most `O(sqrt(T * ln(n)))` worse than the best single signal. The `eta` parameter trades off between adaptation speed and stability — higher `eta` reacts faster but is noisier.

**Implementation note**: Work in log-space to avoid numerical underflow. Subtract `max_log` before exp() to avoid overflow (log-sum-exp trick).

### Method 6: Rank-Based (Adaptive)

**Use case**: Robust to outliers. Weights depend only on relative ordering of signal accuracy, not magnitudes.

**Parameters**: `window: int = 50` (rolling window), `error_metric: ErrorMetric = ABSOLUTE`

**Algorithm**:
```
accuracy_i = 1.0 - mean(errors_i)       # mean accuracy over rolling window

# Rank signals by accuracy (best = highest rank = n, worst = 1)
# Ties get the average rank.
ranks = rank_by_accuracy(accuracies)

# Normalize
weights[i] = ranks[i] / sum(ranks[j] for all j)
```

**State per signal**: `deque[float]` of errors, maxlen = window.

**Initial weights**: `1/n` until at least 1 error per signal.

**Properties**: A signal with accuracy 0.99 vs 0.90 gets adjacent ranks, not 10x the weight. This prevents a single dominant signal from monopolizing the blend. Useful when error distributions are heavy-tailed or when you want diversification.

### Method 7: Bayesian Model Averaging (Adaptive)

**Use case**: Principled probabilistic weighting. Each signal is treated as a "model" with a posterior probability proportional to how well it predicts observed outcomes.

**Parameters**: `prior: list[float] = None` (defaults to uniform `[1/n, ..., 1/n]`)

**Algorithm**:
```
# Likelihood of outcome given signal_i (Bernoulli model):
# P(outcome | signal_i) = signal_i^outcome * (1 - signal_i)^(1 - outcome)
#
# In log-space:
# log P = outcome * log(signal_i) + (1 - outcome) * log(1 - signal_i)
# (with signals clamped to [epsilon, 1-epsilon] to avoid log(0))

log_lik_i = outcome * log(clamp(signal_i)) + (1 - outcome) * log(clamp(1 - signal_i))

# Update log-posterior
log_posterior_i += log_lik_i

# Normalize via softmax
weights[i] = softmax(log_posteriors)[i]
```

**State per signal**: `float` — cumulative log-posterior.

**Initial state**: `log_posterior_i = log(prior_i)`.

**Properties**: The Bernoulli likelihood treats each signal as a probability estimate and each outcome as a binary observation. A signal that consistently predicts outcomes well accumulates high log-likelihood. The posterior naturally balances prior beliefs with observed evidence.

**Clamping**: Signals are clamped to `[1e-15, 1 - 1e-15]` before taking log to prevent `-inf`.

**Convergence**: The posterior concentrates on the best-predicting signal(s) over time. With uniform prior, this is equivalent to maximum likelihood weighting. Custom priors allow injecting domain knowledge (e.g., "I trust the RSI signal more a priori").

## API Design

### `Aggregator` Class

```python
class Aggregator:
    def __init__(self,
        n_signals: int,
        method: AggregationMethod = AggregationMethod.EQUAL,
        feedback_delay: int = 1,
        *,
        # Method-specific parameters (keyword-only):
        weights: list[float] | None = None,        # FIXED: required
        window: int = 50,                           # INVERSE_VARIANCE, RANK_BASED
        alpha: float = 0.1,                         # EXPONENTIAL_DECAY
        eta: float = 0.5,                           # MULTIPLICATIVE_WEIGHTS
        prior: list[float] | None = None,           # BAYESIAN: optional
        error_metric: ErrorMetric = ErrorMetric.ABSOLUTE,  # INVERSE_VARIANCE, RANK_BASED
    )

    def blend(self, signals: list[float]) -> float
    def update(self, outcome: float) -> None
    def warmup(self, history: list[tuple[list[float], float]]) -> None

    @property
    def weights(self) -> list[float]

    @property
    def count(self) -> int
```

### Constructor Validation

- `n_signals >= 1`
- `feedback_delay >= 1`
- `FIXED` requires `weights` with length `n_signals`; normalized to sum to 1.0
- `window >= 2` for `INVERSE_VARIANCE` and `RANK_BASED`
- `0 < alpha <= 1` for `EXPONENTIAL_DECAY`
- `eta > 0` for `MULTIPLICATIVE_WEIGHTS`
- `prior` length must equal `n_signals` if provided; normalized to sum to 1.0

### `blend(signals) -> float`

1. Validate `len(signals) == n_signals`
2. Compute `output = sum(weights[i] * signals[i])` (or `/ sum(weights)` if not pre-normalized)
3. Append `signals` to ring buffer
4. Increment `count`
5. Return output in [0, 1]

### `update(outcome) -> None`

1. If method is `FIXED` or `EQUAL`: return (no-op)
2. If ring buffer has fewer than `feedback_delay + 1` entries: return (not enough history)
3. Retrieve signals from `feedback_delay` bars ago
4. Compute per-signal error: `error_i = metric(signals[i], outcome)`
5. Update method-specific state
6. Recompute and normalize weights

### `warmup(history) -> None`

```python
def warmup(self, history: list[tuple[list[float], float]]) -> None:
    outcomes = []
    for signals, outcome in history:
        self.blend(signals)
        outcomes.append(outcome)
        idx = len(outcomes) - 1 - self._feedback_delay
        if idx >= 0:
            self.update(outcomes[idx])
```

### `weights` (property)

Returns a copy of the current normalized weights. Read-only.

### `count` (property)

Returns the total number of `blend()` calls.

## Internal State

### Shared State (all methods)

```python
_n: int                              # number of signals
_method: AggregationMethod
_feedback_delay: int
_weights: list[float]                # current normalized weights, sum = 1.0
_ring: deque[list[float]]            # buffered signal vectors, maxlen = feedback_delay + 1
_count: int                          # total blend() calls
```

### Per-Method State

| Method | Additional fields |
|--------|-------------------|
| FIXED | (none — weights set once in constructor) |
| EQUAL | (none — weights always 1/n) |
| INVERSE_VARIANCE | `_errors: list[deque[float]]` (one deque per signal, maxlen = window), `_error_metric`, `_window` |
| EXPONENTIAL_DECAY | `_ema: list[float]` (one EMA per signal), `_alpha` |
| MULTIPLICATIVE_WEIGHTS | `_log_weights: list[float]` (one per signal), `_eta` |
| RANK_BASED | `_errors: list[deque[float]]` (one deque per signal, maxlen = window), `_error_metric`, `_window` |
| BAYESIAN | `_log_posterior: list[float]` (one per signal) |

## Design Decisions

### D1: Input/Output Range — [0, 1]

All inputs and outputs are in [0, 1], consistent with the `signals/` module.

**Rationale**: Signals represent confidence/probability, not direction. Direction is handled by the caller (e.g., separate buy and sell ensembles, or sign convention outside the ensemble). Keeping [0, 1] simplifies the math (Bernoulli likelihood in Bayesian, accuracy = 1 - error, etc.) and avoids ambiguity about whether negative values mean "bearish confidence" or "anti-confidence."

### D2: Delayed Feedback via Ring Buffer

The ensemble buffers signal vectors internally and pairs them with outcomes when `update()` is called.

**Rationale**: The alternative (caller manages pairing) adds bookkeeping burden to every consumer. Since the delay is fixed per ensemble instance, the ring buffer is simple and correct. The maxlen = `feedback_delay + 1` keeps memory bounded.

### D3: Warmup Produces Identical Results to Live Replay

`warmup()` is syntactic sugar over sequential `blend()` + `update()` calls.

**Rationale**: No "batch mode" with different math. The morning warmup and live session use the same code path. This avoids subtle divergences and makes the method easy to reason about.

### D4: ErrorMetric as Configurable Enum

`INVERSE_VARIANCE` and `RANK_BASED` accept an `ErrorMetric` parameter (ABSOLUTE or SQUARED).

**Rationale**: Absolute error is simpler and more robust to outliers. Squared error penalizes large misses more, which may be desirable when large errors are especially costly. Making it configurable lets the caller choose without code changes.

### D5: Softmax Normalization in Log-Space

`MULTIPLICATIVE_WEIGHTS` and `BAYESIAN` work in log-space and normalize via softmax with the log-sum-exp trick.

**Rationale**: Direct weight multiplication causes underflow after many updates. Log-space is numerically stable. The max-subtraction trick (`exp(log_w - max_log)`) prevents overflow in exp().

### D6: Bernoulli Likelihood for Bayesian

The Bayesian method models `P(outcome | signal)` as a Bernoulli distribution: `signal^outcome * (1-signal)^(1-outcome)`.

**Rationale**: With signals and outcomes both in [0, 1], Bernoulli is the natural choice — it treats the signal as a probability estimate and the outcome as a soft observation. A signal of 0.8 that sees outcome 0.9 gets high likelihood; a signal of 0.2 seeing outcome 0.9 gets low likelihood. This is the same loss function used in logistic regression (cross-entropy).

### D7: Tolerance for Missing Updates

`update()` is optional. If skipped, weights stay unchanged and the corresponding buffered signals are silently dropped from the ring buffer.

**Rationale**: In live trading, not every bar produces an actionable signal or a measurable outcome. Requiring `update()` every bar would force the caller to fabricate outcomes. Tolerating gaps is more practical.

### D8: Zero Dependencies

The signal ensemble module depends on no other zpano module.

**Rationale**: The math is simple (weighted averages, variance, softmax). Importing from `fuzzy/` would add coupling for no benefit. The module should work in any context where the caller has [0, 1] signal values, regardless of how they were produced.

## Cross-Language Naming Conventions

| Concept | Python | Go | TypeScript | Zig | Rust |
|---------|--------|----|------------|-----|------|
| Class/struct | `Aggregator` | `Aggregator` | `Aggregator` | `Aggregator` | `Aggregator` |
| Method enum | `AggregationMethod` | `AggregationMethod` | `AggregationMethod` | `AggregationMethod` | `AggregationMethod` |
| Enum member | `INVERSE_VARIANCE` | `InverseVariance` | `INVERSE_VARIANCE` | `inverse_variance` | `InverseVariance` |
| Error metric enum | `ErrorMetric` | `ErrorMetric` | `ErrorMetric` | `ErrorMetric` | `ErrorMetric` |
| blend method | `blend(signals)` | `Blend(signals)` | `blend(signals)` | `blend(signals)` | `blend(signals)` |
| update method | `update(outcome)` | `Update(outcome)` | `update(outcome)` | `update(outcome)` | `update(outcome)` |
| warmup method | `warmup(history)` | `Warmup(history)` | `warmup(history)` | `warmup(history)` | `warmup(history)` |
| weights property | `weights` | `Weights()` | `weights` | `getWeights(buf)` | `weights()` |
| File: aggregator | `aggregator.py` | `aggregator.go` | `aggregator.ts` | `aggregator.zig` | `aggregator.rs` |
| File: method enum | `method.py` | `method.go` | `aggregation-method.ts` | `method.zig` | `method.rs` |
| File: error metric | `error_metric.py` | `errormetric.go` | `error-metric.ts` | `error_metric.zig` | `error_metric.rs` |
| File: tests | `test_aggregator.py` | `aggregator_test.go` | `aggregator.spec.ts` | (inline) | (inline) |

## Package Structure

### Python (`py/signal_ensemble/`)
```
__init__.py                # Barrel exports: Aggregator, AggregationMethod, ErrorMetric
method.py                  # AggregationMethod enum
error_metric.py            # ErrorMetric enum
aggregator.py              # Aggregator class
test_aggregator.py         # Tests
```

### Go (`go/signalensemble/`)
```
doc.go                     # Package documentation
method.go                  # AggregationMethod enum
errormetric.go             # ErrorMetric enum
aggregator.go              # Aggregator struct + methods
aggregator_test.go         # Tests
```

### TypeScript (`ts/signal-ensemble/`)
```
index.ts                   # Barrel re-exports
method.ts                  # AggregationMethod enum
error-metric.ts            # ErrorMetric enum
aggregator.ts              # Aggregator class
aggregator.spec.ts         # Tests
```

### Zig (`zig/src/signal_ensemble/`)
```
signal_ensemble.zig        # Barrel re-exports (pub const from sub-modules)
method.zig                 # AggregationMethod enum
error_metric.zig           # ErrorMetric enum
aggregator.zig             # Aggregator struct + methods + inline tests
```

### Rust (`rs/src/signal_ensemble/`)
```
mod.rs                     # Module root with pub use re-exports
method.rs                  # AggregationMethod enum
error_metric.rs            # ErrorMetric enum
aggregator.rs              # Aggregator struct + methods + inline tests
```

## Usage Examples

### Basic: Equal-Weight Ensemble

```python
from py.signal_ensemble import Aggregator, AggregationMethod

agg = Aggregator(n_signals=3, method=AggregationMethod.EQUAL)

# Each bar: compute signals from different techniques
confidence = agg.blend([rsi_signal, macd_signal, bb_signal])
# confidence = (rsi_signal + macd_signal + bb_signal) / 3
```

### Adaptive: Inverse-Variance with Morning Warmup

```python
from py.signal_ensemble import Aggregator, AggregationMethod, ErrorMetric

agg = Aggregator(
    n_signals=5,
    method=AggregationMethod.INVERSE_VARIANCE,
    feedback_delay=2,
    window=100,
    error_metric=ErrorMetric.SQUARED,
)

# Warm up on yesterday's data (390 one-minute bars)
# Each entry: (signals_at_bar_T, outcome_for_bar_T)
agg.warmup(yesterdays_data)

# Live session
for bar in live_stream:
    signals = [compute_rsi(bar), compute_macd(bar), compute_bb(bar),
               compute_vol(bar), compute_pattern(bar)]
    confidence = agg.blend(signals)

    if bar.index >= 2:  # have outcome for bar T-2
        outcome = compute_outcome(bar_minus_2)
        agg.update(outcome)

    if confidence > 0.6:
        execute_buy(size=confidence)  # proportional sizing
```

### Online Learning: Multiplicative Weights

```python
agg = Aggregator(
    n_signals=4,
    method=AggregationMethod.MULTIPLICATIVE_WEIGHTS,
    feedback_delay=1,
    eta=0.3,
)

# Over time, the weights converge toward the best signal
for bar in stream:
    confidence = agg.blend(signals)
    agg.update(outcome)

# After 1000 bars:
print(agg.weights)  # e.g., [0.45, 0.30, 0.20, 0.05]
# Signal 0 dominated — it was the most accurate
```

### Bayesian with Informative Prior

```python
agg = Aggregator(
    n_signals=3,
    method=AggregationMethod.BAYESIAN,
    feedback_delay=1,
    prior=[0.5, 0.3, 0.2],  # I trust signal 0 more a priori
)

# Posterior updates with each observation
# If signal 2 turns out to be best, posterior will eventually override the prior
```

## Cross-Language Porting Guide

Lessons learned from the five-language implementation. This section complements the architecture above with practical conversion notes.

### Method-Specific State Modeling

The biggest structural divergence across languages is how method-specific state is represented:

| Language | Pattern | Details |
|----------|---------|---------|
| **Python** | Flat attributes | `self._ema`, `self._log_weights`, etc. set in `__init__` per-method branch. Unused attrs simply don't exist. |
| **Go** | Flat struct fields | All method-specific fields live on `Aggregator` struct. Unused fields stay zero-valued. Comment `// Method-specific state.` groups them. |
| **TypeScript** | Optional private fields | `private _ema?: number[]`, `private _logWeights?: number[]`. Declared with `?` suffix, set only by relevant method branch. |
| **Zig** | Tagged union (`MethodState`) | `union(enum)` with variants: `.none`, `.inverse_variance`, `.exponential_decay`, etc. Each variant holds only its own fields. Most type-safe approach. |
| **Rust** | Enum (`MethodState`) | `enum` with struct variants. Same pattern as Zig but with Rust ownership semantics. Match arms destructure the active variant. |

**Key insight**: Go and Python use "flat" state (all fields on one struct/object, unused fields are zero/absent). Zig and Rust use sum types (tagged union/enum) that make illegal states unrepresentable. TypeScript is in between (optional fields). When porting, don't force one language's pattern onto another — use the idiomatic representation.

### Ring Buffer Implementation

Each language uses a different ring buffer strategy:

| Language | Implementation |
|----------|---------------|
| **Python** | `collections.deque(maxlen=capacity)` — built-in, automatic eviction |
| **Go** | Slice with manual `copy(ring, ring[1:])` shift — simple but O(n) per append |
| **TypeScript** | Array with `push()` + `shift()` when over capacity — same O(n) approach |
| **Zig** | Pre-allocated slice array with `ring_start`/`ring_len` indices — true O(1) circular buffer, no allocations after init |
| **Rust** | `VecDeque<Vec<f64>>` — built-in double-ended queue with O(1) amortized push/pop |

**Key insight**: Python and Rust have built-in ring buffer types. Go and TypeScript use simple slice/array shifting (acceptable for the small sizes involved — `feedback_delay + 1` is typically small). Zig uses a manual circular buffer with start/len tracking for zero-allocation operation after init.

### Rolling Error Window

The per-signal error history window has similar diversity:

| Language | Implementation |
|----------|---------------|
| **Python** | `collections.deque(maxlen=window)` |
| **Go** | Custom `rollingWindow` struct with slice + manual shift |
| **TypeScript** | Array with `push()` + `shift()` |
| **Zig** | Custom `RollingWindow` struct with pre-allocated slice, `start`/`len` indices, and helper methods (`sum`, `populationVariance`, `get`) |
| **Rust** | Custom `RollingWindow` struct with `Vec<f64>`, `start`/`len`/`capacity` fields |

### Zig-Specific Patterns

1. **Validate-then-allocate**: All parameter validation happens in a first `switch` before any `allocator.alloc()` calls. This avoids partial allocation cleanup on validation errors — a critical pattern in Zig where there's no RAII.

2. **Stack-allocated buffers in `rankWithTies`**: Uses `var indices_buf: [256]usize = undefined` and `var values_buf: [256]f64 = undefined` on the stack instead of heap allocation. Safe because `n_signals` is bounded in practice. Avoids allocator dependency in a hot path.

3. **`@memcpy` for signal buffering**: Ring buffer slots are pre-allocated slices; `@memcpy(self.ring[write_pos], signals)` copies signal data without allocation.

4. **Warmup stack buffer**: `var outcomes_buf: [4096]f64 = undefined` avoids heap allocation for the outcomes list during warmup (up to 4096 bars, sufficient for intraday sessions).

5. **Error set**: Custom `AggregatorError` with specific variants per validation failure, combined with `std.mem.Allocator.Error` via error union in `init()`.

### Rust-Specific Patterns

1. **Borrow checker and `update()`**: The `update` method needs `&mut self.state` and `&mut self.weights` simultaneously. Solution: private update methods are `fn update_*(signals, outcome, state_fields, n, weights)` — static methods that receive decomposed borrows, avoiding the need for `&mut self` in the inner methods.

2. **`Default` trait for `AggregatorParams`**: Enables `..Default::default()` pattern for partial construction, matching Go's zero-value-is-useful philosophy.

3. **`VecDeque` for ring buffer**: Standard library provides O(1) push/pop at both ends, perfect for the sliding window pattern.

4. **`mod.rs` Style A**: Simple `pub mod` + `pub use *` glob re-exports. The module is small enough that explicit re-exports would add noise without benefit.

### Go-Specific Patterns

1. **Params struct with defaults**: `DefaultParams()` factory function returns a params struct with sensible defaults. Callers override only the fields they need.

2. **`rollingWindow` helper type**: Small unexported struct with `append()` and `len()` methods. Encapsulates the sliding window logic without importing external packages.

3. **`sort.Slice` for ranking**: Uses `sort.Slice` with a closure for index-based sorting in `rankWithTies`. Go's sort is not stable, but for ranking with tie-averaging this doesn't matter.

4. **`doc.go`**: Package-level documentation in a separate file, per Go convention.

### TypeScript-Specific Patterns

1. **Optional private fields with `?`**: `private _ema?: number[]` — TypeScript's optional property syntax is a natural fit for method-specific state.

2. **Array spread for copies**: `[...signals]` for ring buffer entries, `[...this._weights]` for the `weights` getter. Idiomatic and concise.

3. **Jasmine `describe`/`it` nesting**: Tests grouped by method with shared setup patterns.

4. **Numeric enum starting at 0**: Both `AggregationMethod` and `ErrorMetric` use `UPPER_SNAKE_CASE` members starting at 0.

### Comment Alignment Discipline

After implementing all five languages, aligning inline comments was a significant effort. Lessons:

1. **Write TS/Python first with full comments** — these are the reference implementations with the richest doc conventions (`@param`/`@returns` in TS, NumPy-style docstrings in Python).

2. **Port comments alongside code** — don't defer "comment alignment" to a separate pass. It's much harder to retrofit comments across 5 languages than to carry them forward during initial porting.

3. **Comment categories to track**:
   - **Struct/class field docs**: What each field stores and when it's used
   - **Method doc comments**: Purpose, parameters, return values, error conditions
   - **Inline comments**: Algorithm steps, why-not-what explanations, numeric constants
   - **Section dividers**: `// ── Private update methods ──` separators

4. **Language-idiomatic doc formats**:
   - Python: NumPy-style docstrings with `Parameters` / `Returns` sections
   - Go: Godoc prose in `//` comments, params/returns described inline (no `@param` tags)
   - TypeScript: JSDoc with `@param` / `@returns` tags
   - Zig: `///` doc comments, prose style, backtick-quoted param names
   - Rust: `///` rustdoc with `# Errors` / `# Arguments` sections

5. **Private method docs are easy to miss**: Internal `_update_inverse_variance`, `_compute_error`, `rank_with_ties` etc. tend to get skipped during porting because they're "private." But they're critical for maintainability — add doc comments even for unexported functions.

### Test Counts

| Language | Test count | Test framework |
|----------|-----------|----------------|
| Python | 51 | `unittest.TestCase` |
| Go | 51 | `testing` (table-driven subtests) |
| TypeScript | 51 | Jasmine 5 (`describe`/`it`) |
| Zig | 46 | Built-in `test` blocks |
| Rust | 46 | Built-in `#[test]` |

Zig and Rust have 46 tests (vs 51) because some test variations that use separate test functions in Python/Go/TS are condensed into single tests with loops in Zig/Rust.

### Common Porting Pitfalls

1. **Ring buffer off-by-one**: The ring capacity must be `feedback_delay + 1`, not `feedback_delay`. The buffer needs to hold the current bar's signals plus `feedback_delay` previous bars.

2. **Population variance, not sample variance**: Inverse-variance uses `sum(diff^2) / n`, not `/ (n-1)`. All five languages must use population variance for matching results.

3. **Clamping epsilon**: Bayesian method clamps signals to `[1e-15, 1 - 1e-15]`. This constant must be identical across languages.

4. **EMA neutral prior**: Exponential decay initializes `ema[i] = 0.5` (not 0.0 or 1.0). This ensures initial weights are uniform.

5. **Log-space uniform**: Multiplicative weights initializes `log_weights[i] = 0.0` (not `log(1/n)`). Since `exp(0) = 1` for all signals, softmax gives uniform weights. Using `log(1/n)` would also give uniform weights but is unnecessary.

6. **Warmup feedback delay**: `warmup()` must handle the delay internally — it feeds outcomes starting from index `feedback_delay`, not from index 0. The pattern is:
   ```
   for i, (signals, outcome) in enumerate(history):
       blend(signals)
       idx = i - feedback_delay
       if idx >= 0:
           update(outcomes[idx])
   ```
