# Composite Cycle

This folder is the standalone **research prototype master reference implementation** of `CompositeCycle`. Its numeric authority is `composite-cycle.py`; this document contains the theoretical rationale, complete disclosed equations, explicit reconstruction choices, and porting contract. No source book, sibling folder, framework, NumPy, pandas, network service, or third-party package is required to run or port the numerical code.


## Principles and source contract

Part 1's Cycle Plotter and composite examples sum individual detected sine
waves to reconstruct or project cyclic behavior. The physical analogy is
superposition: two waves can reinforce or cancel depending on their phase.
The formula specifies a model generator, not intrinsically a price-input
indicator. This master therefore offers the pure `composite_value(models,u)`
helper and a price-stream wrapper with **configured periods** and rolling
GDFT fits. The fitting wrapper is an explicit adaptation; no proprietary
automatic peak selection is claimed.

This is the ordinary composite cycle. It is not Part 2's multi-year Harmonic
Energy Superposition model, which decodes complete historical years and
aligns different annual waves to a future target year.

## Complete mathematical calculation

The Hodrick–Prescott decomposition writes $x_i=\tau_i+c_i$ and chooses trend
$\tau$ by minimizing the complete disclosed smoothness objective:

$$\min_{\tau_0,\ldots,\tau_{N-1}}
\sum_{i=0}^{N-1}(x_i-\tau_i)^2+
\lambda\sum_{i=0}^{N-3}(\tau_i-2\tau_{i+1}+\tau_{i+2})^2.$$

Let $D$ be the second-difference matrix, whose row is $[1,-2,1]$. Then

$$A\tau=x,\qquad A=I+\lambda D^TD,\qquad c=x-\tau.$$

For $N\ge6$, the diagonal and two upper diagonals of this symmetric
pentadiagonal matrix, indexed from zero, are

$$a_0=a_{N-1}=1+\lambda,\quad a_1=a_{N-2}=1+5\lambda,
\quad a_i=1+6\lambda\ (2\le i\le N-3),$$

$$b_0=b_{N-2}=-2\lambda,\quad b_i=-4\lambda\ (1\le i\le N-3),
\quad b_{N-1}=0,$$

$$c_i=\lambda\ (0\le i\le N-3),\quad c_{N-2}=c_{N-1}=0.$$

Here $c_i$ denotes a matrix diagonal coefficient, not the cyclical residual.
The code follows the supplied forward elimination and backward substitution
in `hp_filter.cpp`, retaining its arithmetic order. For an independent
derivation, a banded $LDL^T$ solve gives the same mathematical solution:

$$\ell_{i,i-2}=c_{i-2}/d_{i-2},$$
$$\ell_{i,i-1}=
\frac{b_{i-1}-\ell_{i,i-2}d_{i-2}\ell_{i-1,i-2}}{d_{i-1}},$$
$$d_i=a_i-\ell_{i,i-1}^{2}d_{i-1}-\ell_{i,i-2}^{2}d_{i-2}.$$

Terms with negative indices are omitted. Solve $Ly=x$, then $Dz=y$, then
$L^T\tau=z$. This supplementary derivation explains why only two preceding
states are required; a port should preserve the **actual master's** forward
loop to reproduce its floating-point results. The objective has a unique
solution for finite $\lambda\ge0$. At $\lambda=0$, $\tau=x$ and the residual
is zero. A constant or linear sequence has zero second differences and is
preserved mathematically, up to numerical roundoff.

The physical analogy is a flexible ruler fitted to observations: the first
term holds it near prices; the second resists changes of slope. Increasing
$\lambda$ makes the ruler harder to bend. This analogy explains smoothing,
not evidence that a financial series contains a real periodic force. The
book and whitepaper explicitly acknowledge that HP filtering can create
apparent cycles and require subsequent validation.


Positive smoothing yields $y=x-\tau$; zero disables preprocessing.

For configured wavelength $P>2$ and an oldest-first $N$-sample window $y_j$,
define $\omega=2\pi/P$, $s=\sin\omega$, $c=\cos\omega$, and $b=2c$.
Reset the two delay states before **each window and each candidate period**:

$$q_{-2}=q_{-1}=0,\qquad q_j=bq_{j-1}-q_{j-2}+y_j.$$

After the last sample, let $Q_1=q_{N-1}$ and $Q_2=q_{N-2}$. The supplied
source computes the final complex value and explicitly rotates its reference:

$$r_1=Q_1-Q_2\cos\omega,\quad i_1=Q_2\sin\omega,$$

$$C=\cos((N-1)\omega),\quad S=\sin((N-1)\omega),$$

$$R=r_1C+i_1S,\qquad I=i_1C-r_1S.$$

These equal the real and imaginary parts of the negative-exponent transform

$$X(\omega)=\sum_{j=0}^{N-1}y_j e^{-i\omega j}.$$

The published amplitude, **sine** phase, and trough offset are

$$A_P=\frac{2\sqrt{R^2+I^2}}N,\qquad
\phi_P=\frac\pi2+\operatorname{atan2}(I,R),$$

$$o_P=P\left(\frac34-\frac{\phi_P}{2\pi}\right).$$

The returned phase references the **first sample of the window**. It is not
an end-of-window phase or the proprietary scanner's separately fitted recent
phase. It remains in the source's unwrapped interval $[-\pi/2,3\pi/2]$.
The modeled wave and slope at offset $u$ are

$$w_P(u)=A_P\sin(2\pi u/P+\phi_P),$$
$$w'_P(u)=A_P(2\pi/P)\cos(2\pi u/P+\phi_P).$$

Use $u=N-1$ for the current observation and $u=N-1+h$ for an $h$-bar
projection. The source's $o_P$ is an unwrapped offset from the window start;
the next trough's distance from the current endpoint is

$$\operatorname{ETA}_{\min}=(o_P-(N-1))\bmod P.$$

Amplitude zero has no meaningful phase. This master returns amplitude zero
and phase/offset NaN in the exact zero-vector case, instead of manufacturing
a phase from `atan2(0,0)`. A zero-amplitude modeled wave still evaluates to
zero. No epsilon-based energy gate, taper/window weighting, mean subtraction,
least-squares amplitude correction, or spectral normalization is added.

The physical analogy is a receiver measuring a chosen tone in a short noisy
record, historically used for telephone signaling. It does not require the
frequency to occupy an integer FFT bin. Rectangular-window leakage still
occurs: non-integer cycles, trends, and nearby components affect amplitude
and phase. The $2/N$ amplitude is exact for isolated, coherently sampled
sinusoids away from DC and Nyquist; it is not generally an exact regression
coefficient for arbitrary fractional periods.


For the configured ordered, nonempty set of distinct periods $P_k$, fit each
to the same rolling data and construct

$$C(u)=\sum_{k=1}^{K}A_k\sin(2\pi u/P_k+\phi_k),$$

$$C'(u)=\sum_{k=1}^{K}A_k\frac{2\pi}{P_k}\cos(2\pi u/P_k+\phi_k).$$

Current output is $C(W-1)$, analytic slope is $C'(W-1)$, and `next_value`
is $C(W)$. Read-only `project(h)` returns $C(W-1+h)$ for nonnegative real
$h$. The `models` tuple snapshots the most recently accepted fit, using the
first sample of **that window** as phase origin. `project` never inserts
predictions into observed history.

The single-frequency coefficients are separately estimated projections,
not a joint multivariate least-squares fit. Leakage/cross-talk between
nearby/incommensurate periods can cause imperfect reconstruction and
double counting. Do not add fitted price trend to this cyclic output
implicitly. Refit all models on each accepted price; past emitted outputs
remain unchanged while the current forecast can change.

## Input and output contract

`update(price)` accepts one chronological scalar observation. A missing NaN returns unavailable output without advancing state. Infinity, booleans, malformed parameters and input-contract violations raise `ValueError`. Reset preserves configuration and clears every numerical history.

| Output | Meaning and porting representation |
|---|---|
| `value` | Sum of configured fitted sine waves at current endpoint; price units; NaN means unavailable. |
| `slope` | Analytic derivative of the current composite wave; price units per bar; NaN means unavailable. |
| `next_value` | One-bar-ahead value of current fitted composite; price units; forecast, not observation; NaN means unavailable. |

Named numeric outputs can map to zpano scalar plots; lower/upper map to `Band`, and pivot offsets map to `Levels`/`Polyline`. They intentionally carry numbers rather than colors. The zero-lag EMA returns a scalar float. Output snapshots do not share mutable recursive state with later updates.

All recurrence/state operations use Python binary64 floats and full-precision `math.pi`. Porting requires binary64 equivalents and the documented iteration order. Numerical overflow is rejected; after an arithmetic failure reset before reusing the instance. Parameter/input validation and skipped NaNs themselves do not mutate state. Priming means data availability, not settled IIR transients, statistical significance, or predictive performance.

## Configuration

| Parameter | Shortcut | Default | Valid range | Purpose |
|---|---|---|---|---|
| `window` | `W` | `100` | integer ≥6 | Accepted prices in each rolling analysis |
| `periods` | `P` | `(20.0, 40.0)` | nonempty ordered sequence of distinct finite reals >2 | Configured wavelengths, in supplied summation order |
| `smoothing` | `L` | `1600.0` | finite real ≥0 | HP penalty; zero disables preprocessing in spectral/Bartels wrappers |

Values that are finite but too extreme for resolvable binary64 arithmetic can be rejected during calculation. There is no undocumented automatic retuning; construct another configured instance and replay required history when changing parameters.

## Calculation flow

```mermaid
flowchart TD
    A[Input observation] --> B{Finite valid input?}
    B -->|NaN| X[Unavailable output; no state advance]
    B -->|Invalid| E[Raise ValueError]
    B -->|Valid| N0[Retain W prices]
    N0 --> N1[HP residual or raw input]
    N1 --> N2[GDFT for configured periods]
    N2 --> N3[Sum sine values and analytic slopes]
    N3 --> N4[Evaluate one-bar projection without feedback]
    N4 --> N5[Emit composite and snapshot models]
```

Startup NaNs and any output-specific readiness gates described above are applied while preserving the documented internal calculations. Rolling analysis recomputes the fixed window and emits only the current snapshot. It never overwrites previously emitted stream observations.

## Reference parameter combinations

The following combinations are generated automatically; the number is below the requested limit of 64. Every numeric output gets one 252-element array per combination in each language.

| Case | Parameters | What it tests |
|---|---|---|
| 1 | `window=100, periods=(20.0, 40.0), smoothing=1600.0` | Default two-period composite fitted to HP residual. |
| 2 | `window=100, periods=(20.0,), smoothing=0.0` | Single-wave composite degenerates to its configured cycle. |
| 3 | `window=100, periods=(20.0, 40.0), smoothing=0.0` | Two configured waves fitted without detrending. |
| 4 | `window=120, periods=(10.0, 20.0, 40.0), smoothing=1600.0` | Three harmonic periods. |
| 5 | `window=100, periods=(17.5, 31.0), smoothing=1600.0` | Fractional and incommensurate configured periods. |
| 6 | `window=200, periods=(20.0, 40.0), smoothing=1600.0` | Longer analysis window. |
| 7 | `window=100, periods=(20.0, 40.0), smoothing=160.0` | Weak detrending penalty. |
| 8 | `window=253, periods=(20.0, 40.0), smoothing=1600.0` | No fitted composite before an overlong window is full. |

## Generating and porting fixtures

From this folder run:

```sh
python3 composite-cycle.py
python3 composite-cycle.py --generate-reference-data
```

The first command demonstrates chronological updates for every configuration. The second regenerates all expected fixtures and verifies the saved Python decimals. The example embeds the exact 252 shared `INPUT_CLOSE` values; auxiliary OHLC is read from the copied local `test_testdata.py`. All five input fixture files are preserved, including their matching OHLC arrays.

Python identifiers follow `EXPECTED_<shortcut><value>_..._<OUTPUT>`. Integral-valued floats use their shortest integer spelling, fractional dots become underscores, and a period tuple is encoded in order (for example `P20_40`). Multiple outputs have explicit suffixes. Spectrum suffixes `A20`, `PH20`, and `TO20` mean amplitude, first-window sine phase, and trough offset for the 20-bar bin; `DOM`, `AMP`, and `PHASE` mean the selected candidate fields.

Before each array, the first comment repeats “What it tests” and identifies its output; the next records full parameter names/values. Arrays contain **15 significant decimal digits** and NaNs at each unavailable output position. Calculations are never rounded internally.

| Language | Identifier and shape | Missing value |
|---|---|---|
| Python | `EXPECTED_... = [...]` | `math.nan` |
| Go | `var expected... = []float64{...}` | `math.NaN()` |
| TypeScript | `export const expected...: number[] = [...]` | `NaN` |
| Rust | `pub fn expected_...() -> Vec<f64>` inside copied test module | `f64::NAN` |
| Zig | `pub fn expected...() [252]f64` | `nan` from `std.math.nan(f64)` |

The Go package is the folder name with hyphens/underscores removed. Expected Rust suffixes are lowercase; other language parameter shortcuts remain uppercase. Correct language spellings (`math.NaN()`, `f64::NAN`) are used in place of typographical variants in the request.

The generator first saves Python arrays, reads them back, checks length/field set/decimal precision, and only then copies those saved values into the other languages. Generated blocks are replaceable/idempotent and preserve the input arrays. Ported implementations should compare NaN masks exactly and finite binary64 results with an appropriate absolute/relative tolerance; 15-digit fixtures are rounded decimal references, not promises of cross-platform bit identity.

The parent-folder verification scripts check all saved language values and compile/load the fixtures. Numerical tests additionally use independent DFT sums, HP normal equations, known harmonic vectors, analytic wave superposition, calendar boundaries, and reset/missing-input behavior. These checks validate the specified algorithms; they do not validate the books’ trading-performance claims.

## References and BibTeX

1. Lars von Thienen, *Decoding the Hidden Market Rhythm, Part 1: Dynamic
   Cycles*, 2017. Numerical sections are linked to the supplied chapter/code
   archive: [chapter 3](../../2017-decoding-the-hidden-market-rhythm-1/ch03.md),
   [chapter 4](../../2017-decoding-the-hidden-market-rhythm-1/ch04.md),
   [chapter 6](../../2017-decoding-the-hidden-market-rhythm-1/ch06.md), and
   [downloaded source directory](../../2017-decoding-the-hidden-market-rhythm-1/code/).
2. Lars von Thienen, *Decoding the Hidden Market Rhythm, Part 2: Metonic
   Cycles*, second edition/release 2017, [local complete book](../../2017-decoding-the-hidden-market-rhythm-2/content.md),
   especially §§10.1–10.3, chapter 11, and §§12.3–12.5.2.
3. Lars von Thienen, Foundation for the Study of Cycles,
   [*The Cycle Scanner Algorithm*, whitepaper](../../the-cycle-scanner-algorithm-whitepaper/content.md).
   The local document supplies no publication year; none is invented.
4. Robert J. Hodrick and Edward C. Prescott, “Postwar U.S. Business Cycles:
   An Empirical Investigation,” *Journal of Money, Credit and Banking*,
   29(1), 1997, pp. 1–16. The original working-paper formulation dates to 1980.
5. Gerald Goertzel, “An Algorithm for the Evaluation of Finite Trigonometric
   Series,” *The American Mathematical Monthly*, 65(1), 1958, pp. 34–35.
6. Julius Bartels, “Random fluctuations, persistence, and quasi-persistence
   in geophysical and cosmical periodicities,” *Terrestrial Magnetism and
   Atmospheric Electricity*, 40(1), 1935. Cited by the supplied whitepaper.
7. Charles E. Armstrong, “Applying the Bartels Test of Significance to a Time
   Series Cycle Analysis,” 1973, as cited by the whitepaper. Bibliographic
   detail beyond the supplied citation is not asserted.
8. J. Welles Wilder Jr., *New Concepts in Technical Trading Systems*,
   Trend Research, 1978, for the explicit underlying Wilder RSI convention.
9. The [zpano output types](../../../../external/zpano/py/indicators/core/outputs/)
   and [prepared Ehlers examples](../../../john-ehlers/prepared-for-conversion/)
   inform packaging/output contracts, not replacement numerical algorithms.

The source-specific scripts and C++ extracts are numerical authorities where
named above. HP, GDFT, Bartels, cRSI and CSI extracts carry Lars von Thienen's
2017 [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) notice: retain
author/book/site attribution and indicate adaptations. The DRO source states
[MPL 2.0](https://www.mozilla.org/en-US/MPL/2.0/); its Python numerical
extraction retains that source notice. Part 2 itself states all rights
reserved; its formulas are described here as the supplied research source,
without claiming that its book inherits the gist license.

```bibtex
@book{vonThienen2017Dynamic,
  author = {von Thienen, Lars},
  title = {Decoding the Hidden Market Rhythm, Part 1: Dynamic Cycles},
  year = {2017},
  url = {https://www.whentotrade.com/},
  note = {Supplied chapter archive and accompanying numerical source extracts}
}
@book{vonThienen2017Metonic,
  author = {von Thienen, Lars},
  title = {Decoding the Hidden Market Rhythm, Part 2: Metonic Cycles},
  year = {2017},
  edition = {2},
  note = {Kindle release 2017; supplied complete local text}
}
@techreport{vonThienenCycleScanner,
  author = {von Thienen, Lars},
  title = {The Cycle Scanner Algorithm},
  institution = {Foundation for the Study of Cycles},
  note = {Supplied whitepaper; publication date not stated}
}
@article{hodrickPrescott1997,
  author = {Hodrick, Robert J. and Prescott, Edward C.},
  title = {Postwar U.S. Business Cycles: An Empirical Investigation},
  journal = {Journal of Money, Credit and Banking},
  year = {1997}, volume = {29}, number = {1}, pages = {1--16}
}
@article{goertzel1958,
  author = {Goertzel, Gerald},
  title = {An Algorithm for the Evaluation of Finite Trigonometric Series},
  journal = {The American Mathematical Monthly},
  year = {1958}, volume = {65}, number = {1}, pages = {34--35}
}
@article{bartels1935,
  author = {Bartels, Julius},
  title = {Random fluctuations, persistence, and quasi-persistence in geophysical and cosmical periodicities},
  journal = {Terrestrial Magnetism and Atmospheric Electricity},
  year = {1935}, volume = {40}, number = {1}
}
@misc{armstrong1973,
  author = {Armstrong, Charles E.},
  title = {Applying the Bartels Test of Significance to a Time Series Cycle Analysis},
  year = {1973},
  note = {Citation as supplied by the Cycle Scanner whitepaper}
}
@book{wilder1978,
  author = {Wilder, J. Welles Jr.},
  title = {New Concepts in Technical Trading Systems},
  publisher = {Trend Research}, year = {1978}
}
@misc{vonThienenNumericalExtracts2017,
  author = {von Thienen, Lars},
  title = {HP Filter, Goertzel, Bartels Score, Cyclic RSI, and Cycle Swing source extracts},
  year = {2017}, url = {https://www.whentotrade.com/},
  note = {Downloaded local code archive; source files carry CC BY 4.0}
}
@misc{vonThienenDRO,
  author = {von Thienen, Lars},
  title = {Detrended Rhythm Oscillator, Pine v5 source},
  note = {Supplied dro.pinesctipt; MPL 2.0; ZigZag helpers credit LonesomeTheBlue}
}
@misc{zpanoOutputContracts,
  title = {zpano Indicator Core Output Types},
  note = {Local external/zpano/py/indicators/core/outputs; structural compatibility reference}
}
@misc{preparedEhlersMasters,
  title = {Prepared Ehlers Streaming Research Prototype Masters},
  note = {Local inputs/john-ehlers/prepared-for-conversion; fixture and packaging examples}
}
```
