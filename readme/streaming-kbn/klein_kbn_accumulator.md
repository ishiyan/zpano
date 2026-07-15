# Klein second-order Kahan-Babuška-Neumaier (KBN) compensated summation

## Domain: Floating-Point Summation

Adding a sequence of floating-point numbers naively accumulates round-off error because each addition rounds the result to the available significand. Low-order bits of the smaller operand are lost whenever the sum becomes large relative to the addend.

**Naive sum** `s += x`: worst-case relative error grows as `O(ε n)`, and RMS error as `O(ε √n)` (a random walk) for zero-mean inputs [[Higham93](#ref-higham93)]. The bound is proportional to the **condition number** of the summation problem:

```pseudocode
cond = Σ|xᵢ| / |Σxᵢ|
```

A large condition number means the sum is intrinsically sensitive to round-off, regardless of the algorithm.

**Example (Peters)** — `[1.0, 1e100, 1.0, -1e100]`:

| Method | Result |
| --- | --- |
| Exact | 2.0 |
| Naive | 0.0 |
| Kahan | 0.0 |
| **KBN** | **2.0** |

Naive and standard Kahan both return 0.0 because the `1.0` additions are completely lost when `1e100` dominates the significand.

## Algorithm Progression

### Naive Summation

```pseudocode
s = 0
for x in input:
    s += x
return s
```

Error grows linearly with the number of terms and can be catastrophic for ill-conditioned sums.

### Kahan Compensated Summation

William Kahan (1965) [[Kahan65](#ref-kahan65)] introduced a running compensation term `c` that captures the low-order bits lost in each addition:

```pseudocode
s = 0
c = 0
for x in input:
    y = x - c       // reinstate previous loss
    t = s + y
    c = (t - s) - y // capture what s lost when t was rounded
    s = t
return s
```

This reduces the error bound to `O(ε + nε²)`, effectively independent of `n` for practical sizes. However, when `|s|` and `|x|` differ hugely (e.g. `1e100 + 1.0`), Kahan still loses the small addend because the subtraction `(t - s) - y` itself rounds to zero.

### Kahan-Babuška-Neumaier (KBN)

Neumaier (1974) [[Neumaier74](#ref-neumaier74)] improved the algorithm by branching on **which operand is larger**, ensuring the correction is always computed from the smaller operand's perspective:

```pseudocode
s = 0
c = 0
for x in input:
    t = s + x
    if |s| >= |x|:
        c += (s - t) + x    // x lost low-order bits
    else:
        c += (x - t) + s    // s lost low-order bits
    s = t
return s + c
```

The correction is only applied once at the end (`s + c`), unlike Kahan which applies it each iteration.

The branch ensures that the term `(big - (big + small))` — which is exact in floating-point [[2Sum](#ref-2sum)] — is always computed with the small operand as the last subtraction, preserving the error. Peters' example yields the correct 2.0.

### Klein Second-Order KBN (Double-Compensated)

Klein (2006) [[Klein06](#ref-klein06)] generalised KBN to arbitrary order. The second-order variant applies the same KBN trick to **the correction term itself**, maintaining two compensation levels:

```pseudocode
s  = 0
cs = 0     // first-level compensation
ccs = 0    // second-level compensation

for x in input:
    // Level 1: KBN on sum + x
    t = s + x
    if |s| >= |x|: c = (s - t) + x
    else:          c = (x - t) + s
    s = t

    // Level 2: KBN on cs + c
    t = cs + c
    if |cs| >= |c|: cc = (cs - t) + c
    else:           cc = (c - t) + cs
    cs = t
    ccs += cc

return s + (cs + ccs)
```

This double compensation further reduces residual error when the first-level correction `c` itself loses bits during accumulation.

## Implementation: `KleinKBNAccumulator`

The `KleinKBNAccumulator` class implements Klein's second-order KBN algorithm. It maintains three state variables:

| Variable | Purpose |
| --- | --- |
| `_sum` | Primary sum (level 1 result) |
| `_cs` | First-level KBN correction accumulator |
| `_ccs` | Second-level KBN correction (applied to `_cs`) |

The corrected value at any point is `_sum + _cs + _ccs`.

### Methods

- **`update(x)`** — Adds `x` using two-level KBN compensation.
- **`revert(x)`** — Removes `x` by calling `update(-x)`. Provides a symmetrical inverse for rolling-window use, mirroring the `revert(x)` pattern in `Variance` and `RunningVariance`.
- **`set(x)`** — Overwrites the accumulator with `x` (resets both compensation terms to zero).
- **`reset()`** — Resets all state to zero.
- **`value`** — Returns `_sum + _cs + _ccs`.

### Usage

```python
kbn = KleinKBNAccumulator()
for x in data:
    kbn.update(x)
result = kbn.value
```

For rolling windows:

```python
kbn.update(x)   # add new value
kbn.revert(y)   # remove old value
```

## References

- <a id="ref-higham93"></a> Higham, N. J. (1993). "The accuracy of floating point summation". *SIAM Journal on Scientific Computing*, 14(4), 783–799. [doi:10.1137/0914050](https://doi.org/10.1137/0914050)
- <a id="ref-kahan65"></a> Kahan, W. (1965). "Further remarks on reducing truncation errors". *Communications of the ACM*, 8(1), 40. [doi:10.1145/363707.363723](https://doi.org/10.1145/363707.363723)
- <a id="ref-neumaier74"></a> Neumaier, A. (1974). "Rundungsfehleranalyse einiger Verfahren zur Summation endlicher Summen". *Zeitschrift für Angewandte Mathematik und Mechanik*, 54(1), 39–51. [doi:10.1002/zamm.19740540106](https://doi.org/10.1002/zamm.19740540106)
- <a id="ref-klein06"></a> Klein, A. (2006). "A generalized Kahan–Babuška-Summation-Algorithm". *Computing*, 76(3–4), 279–293. [doi:10.1007/s00607-005-0139-x](https://doi.org/10.1007/s00607-005-0139-x)
- <a id="ref-wikipedia"></a> Wikipedia. [Kahan summation algorithm](https://en.wikipedia.org/wiki/Kahan_summation_algorithm).
- <a id="ref-2sum"></a> Wikipedia. [2Sum](https://en.wikipedia.org/wiki/2Sum).
- <a id="ref-kuiperzone"></a> Kuiperzone. [Compensated-Accumulators](https://github.com/kuiperzone/Compensated-Accumulators).
- <a id="ref-numpy-issue"></a> NumPy issue #8786 — [Badly conditioned sum](https://github.com/numpy/numpy/issues/8786).
- <a id="ref-peters"></a> Peters' example discussed in the [CPython `math.fsum` implementation](https://github.com/python/cpython/blob/main/Modules/mathmodule.c).
