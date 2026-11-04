# Dynamic Cycle Explorer

Research prototype master: **streaming-and-cli**. Full porting mathematics is included
here, without requiring the original books.


## Candidate behavior and centered highlighter

Part 1 chapter 6 credits Brian J. Millard's *Channels and Cycles:
A Tribute to J.M. Hurst* for the smoothed detrend/Cycle Highlighter.
Subtract centered means of one full wavelength and half a wavelength:

$$H_c=C_{L_s}(c)-C_{L_\ell}(c),\quad
L_\ell=P,\quad L_s=\max(2,\lfloor P/2+1/2\rfloor).$$

Half-up rounding is this master's convention. For odd $L=2q+1$,

$$C_L(c)=\frac1L\sum_{j=-q}^{q}x_{c+j}.$$

For even $L=2q$, average two adjacent length-$L$ means:

$$C_L(c)=\frac{\frac12x_{c-q}+\sum_{j=-q+1}^{q-1}x_{c+j}
+\frac12x_{c+q}}L.$$

Even support spans $L+1$ observations. At bar $t$, emit only center
$c=t-\lfloor P/2\rfloor$, approximately half a cycle behind.
Missing support is NaN, with no padding or extrapolation.

DynamicCycleExplorer emits current fitted value, period, endpoint phase,
ETA top/bottom, delayed highlighter, its zero-based accepted-bar index,
and fitted wave evaluated at that center. Highlighter uses raw prices,
independently of HP spectral preprocessing.

At each update the highest-strength accepted cycle supplies $P$.
A changing $P$ changes the delay: reported center indices may skip,
repeat, or move backward. Honor the returned index. Earlier emitted
snapshots are not rewritten.

Batch CLI instead uses final selected $P$ to highlight every supported
historical center and overlay the final fitted wave; unsupported edges
are empty. This is retrospective. --rolling exports causal snapshots.
--horizon supplies ideal last-model projections in bars.
An API-equivalent dynamic selector or refined estimator is not claimed.

## Configuration parameters

| Parameter | Default | Range | Meaning |
|---|---:|---|---|
| window | 200 | integer >=6 | Rolling accepted-price buffer; batch uses range length. |
| min_period | 5 | integer >=3, <=max_period | First wavelength. |
| max_period | 80 | integer >=min_period | Last wavelength; two segments needed to pass. |
| smoothing | 1600.0 | finite real >=0 | HP penalty; zero disables preprocessing. |
| genuine_threshold | 49.0 | finite real [0,100] | Strict greater-than genuine-percent gate. |
| phase_cycles | 2.0 | 0 or finite real >=1 | Recent-fit multiplier; zero full window. |
| top_k | 3 | integer >=1 | Fixed ranked slots, NaN when unused. |
| peak_distance | 2 | integer >=1 | Minimum kept-period separation. |
| max_segments | 5 | integer >=2 | Maximum latest Bartels segments. |

Numeric periods reject booleans. Defaults are research settings, not asserted platform defaults.

## Numerical authorities and reconstruction boundary

Part 1 chapter 3 and the Cycle Scanner whitepaper disclose detrending,
spectral detection, validation, and strength ranking. The supplied HP, GDFT
and Bartels C++ extracts provide actual arithmetic. The WTT scanner,
explorer and phasing C# scripts call remote APIs; they do not disclose
their detection engines. Part 1 chapter 6 describes centered moving-average
highlighting, chapter 9 phase states, and Part 2 chapter 3 annual subset
selection and historical superposition.

This master is a **working research reconstruction**, without asserting
numeric equivalence to the WTT platform. Missing dynamic HP adjustments,
peak rules, recent-phase estimation, state boundaries, annual processing
and normalization choices are resolved explicitly below. A CLI changes
how a calculation is used; it cannot recover undisclosed mathematics.

### Hodrick–Prescott detrending

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

Here $\lambda>0$ uses the entire current window's residual $y=x-\tau$.
**$\lambda=0$ is a control meaning raw prices**, without demeaning, rather
than the zero HP residual. A large raw baseline can leak DC into off-bin
wavelengths; exact sine tests use coherent sampling and a zero baseline.
A rolling HP endpoint is causal, but historical residuals inside the current
window can change at the next update.

### Goertzel detection and sine model

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

The preceding unwrapped phase belongs to the primitive Goertzel helper.
The scanner phase instead belongs to the **analysis endpoint**, as specified
in the recent-fit equations below.

### Bartels validation on original prices

Each complete cycle contributes a harmonic vector. Stable amplitudes/phases
reinforce the mean vector; inconsistent vectors cancel. This is the supplied
harmonic periodicity score, not the similarly named rank test.

For integer period $P$, choose $m=\min(\lfloor N/P\rfloor,M)$ using the
most recent $mP$ **raw prices**, oldest first. Require $m\ge2$. With
zero-based index $x=jP+k$:

$$\theta_x=((360/P)(x+1)/180)\pi,$$
$$a_j=\sum_{k=0}^{P-1}x_{jP+k}\sin\theta_{jP+k},\qquad
b_j=\sum_{k=0}^{P-1}x_{jP+k}\cos\theta_{jP+k}.$$
$$\bar a=\frac1m\sum_j a_j,\quad \bar b=\frac1m\sum_j b_j,\quad
E=\frac1m\sum_j(a_j^2+b_j^2),$$
$$A_{\rm mean}=\sqrt{\bar a^2+\bar b^2},\quad A_{\rm rms}=\sqrt E,\quad
a_{\rm ref}=A_{\rm rms}/\sqrt m,$$
$$r=A_{\rm mean}/a_{\rm ref},\qquad B=e^{-r^2},\qquad G=100(1-B).$$

Accept $G>T$, strictly, default $T=49$. Missing segments/zero energy give
NaN and are rejected. The negative exponential avoids reciprocal overflow.
$G$ is the source display convention, **not a calibrated posterior probability**.
For $m$ perfectly equal aligned vectors, $G=100(1-e^{-m})$; at $m=5$
the ceiling is 99.326%, and at $m=2$ it is 86.466%. Threshold 100 rejects all.

Raw-price validation is this master's explicit cross-check against HP-created
apparent cycles. No additional detrending/per-segment mean removal is added.
DC sums cancel mathematically; finite precision and trends still affect them.

### Peak selection, ranking and suppression

Sweep every integer wavelength in the closed interval:

$$\mathcal P=\{P_0,P_0+1,\ldots,P_1\}.$$

For full-window GDFT amplitudes $A_i$, a peak satisfies

$$A_i>0,\qquad A_i>A_{i-1},\qquad A_i\ge A_{i+1}.$$

An absent boundary neighbor is $-\infty$. Boundary peaks are one-sided;
a one-period grid is valid; a plateau chooses its shortest member.
These are **research choices**, not disclosed WTT rules.

After Bartels rejection, sort by decreasing strength

$$S_P=A_P/P.$$

Exact strength ties choose shorter $P$. Visit in that order, keeping a cycle
only if it is at least $D$ bars from every previously kept period.
Suppression follows validation/ranking. Return top $K$ cycles; unused slots
are NaN. accepted_count includes unsuppressed cycles beyond the returned $K$.

Strength has price units per bar, distinct from instantaneous analytic slope.
The whitepaper example $A=300,P=110$ gives $S=2.72727\ldots$, while
$A=200,P=60$ gives $S=3.33333\ldots$; the smaller-amplitude cycle ranks first.

### Recent-phase reconstruction and projections

The whitepaper specifies a shorter-window phase run, without its exact
length or estimator. This master uses

$$R_P=\begin{cases}N,&F=0,\\
\min(N,\max(6,\lceil FP\rceil)),&F\ge1.\end{cases}$$

Recompute HP on the latest $R_P$ raw prices and rerun GDFT at detected $P$.
If $R_P=N$, reuse the full fit. Do not slice the full-window HP residual:
its endpoint conditions differ. For recent-fit amplitude $\widehat A_P$
and first-sample sine phase $\widehat\phi_P$:

$$\psi_P=(\widehat\phi_P+2\pi(R_P-1)/P)\bmod2\pi,$$
$$v_P=\widehat A_P\sin\psi_P,\qquad
v'_P=\widehat A_P(2\pi/P)\cos\psi_P,$$
$$\operatorname{ETA}_{\rm top}=((\pi/2-\psi_P)\bmod2\pi)/(2\pi/P),$$
$$\operatorname{ETA}_{\rm bottom}=((3\pi/2-\psi_P)\bmod2\pi)/(2\pi/P),$$
$$w_P(h)=\widehat A_P\sin(\psi_P+2\pi h/P).$$

$h=0$ is the last observed bar, negative $h$ reconstructs history, positive
$h$ projects the frozen ideal wave. ETA zero means currently at the ideal
extremum. ETAs are in bars, not dates or guaranteed price turning points.
Ranking retains full-window $A_P/P$, never the recent-fit amplitude.
Zero recent-fit amplitude makes phase-dependent fields unavailable.

## Input, startup and portability contract

All numerical work uses standard-library binary64 arithmetic and full
precision math.pi. Preserve operation order in the source loops.
math.fsum is used for correlation/composite sums; ports may use a carefully
accumulated sum with appropriate rounding tolerance. No FFT, matrix library,
market API, NumPy or pandas is required.

The five live wrappers implement `update(price)`, `reset()`, `is_primed`,
and `warmup_bars=N-1`. Equally spaced accepted bars determine phase, not
elapsed time. A finite price advances one bar. NaN emits unavailable
fields and does not advance history, phase or delayed indices. Infinity,
booleans and invalid parameters raise ValueError. Reset after overflow.
The model readers implement `update()` without prices: playback does not fit.

Priming at $N$ accepted prices differs from having an accepted cycle.
After priming, ranked fields can remain NaN while the spectrum is finite
and accepted_count is zero. Periods beyond $N/2$ lack two full segments
and cannot pass validation even if their spectral bins exist.

Each folder includes the **complete** portable `cycle_core.py` and its
same-basename entry point, loading only that colocated core from any
working directory. No parent directory or sibling is required for conversion.
The parent core is the maintenance authority; verification checks all seven
copies byte-for-byte.

Live output uses named tuples and a zpano-shaped Heatmap with amplitude
values, integer period axis and time=None. Attach timestamps externally.
Ranked fields/model waves map to Polyline; fixed phasing thresholds to
Levels/Band. Delayed highlighter values must use their reported indices.

## Processing cost

For $B=P_1-P_0+1$ bins and $N$ observations, detection and validation cost
$O(NB)$ per full rolling update, with $O(N+B+K)$ storage. Each selected
recent fit costs $O(R_P)$. This deliberately transparent reference recomputes
windows and is a research master, rather than an optimized incremental DFT.
Annual selection adds the bounded subset enumeration described below.

## Calculation flow

~~~mermaid
flowchart TD
    N0["Price buffer"]
    N1["Scanner selects cycle"]
    N0 --> N1
    N2["Current model and ETAs"]
    N1 --> N2
    N3["Centered means at delayed center"]
    N2 --> N3
    N4["Current output and delayed highlighter index"]
    N3 --> N4
~~~

## Reference configurations

| Case | Parameters | What it tests |
|---|---|---|
| 1 | `window=200`, `min_period=5`, `max_period=80`, `smoothing=1600.0`, `genuine_threshold=49.0`, `phase_cycles=2.0`, `top_k=3`, `peak_distance=2`, `max_segments=5` | Default rolling scanner with raw-price validation and recent phase. |
| 2 | `window=60`, `min_period=12`, `max_period=12`, `smoothing=1600.0`, `genuine_threshold=49.0`, `phase_cycles=0.0`, `top_k=1`, `peak_distance=2`, `max_segments=5` | Short window, one-period grid, full-window phase origin. |
| 3 | `window=100`, `min_period=5`, `max_period=30`, `smoothing=0.0`, `genuine_threshold=49.0`, `phase_cycles=2.0`, `top_k=2`, `peak_distance=2`, `max_segments=5` | No detrending, smaller grid, two-cycle recent fit. |
| 4 | `window=100`, `min_period=5`, `max_period=30`, `smoothing=1600.0`, `genuine_threshold=100.0`, `phase_cycles=2.0`, `top_k=2`, `peak_distance=2`, `max_segments=5` | 100-percent threshold rejects every cycle but preserves spectrum. |


Expected arrays preserve all original input arrays and use exactly the
252 INPUT_CLOSE values. Each parameter/output has its own 252-element array.
Round once, to 15 significant digits on Python serialization; then copy
the saved decimal values into Go, TypeScript, Rust and Zig.
NaNs are math.nan, math.NaN(), NaN, f64::NAN and Zig nan.
Names encode shortcuts/values, replacing decimal points with underscores,
and append the output field. Each array has the requested two comment lines.

Live references retain startup and missing-rank NaNs. CLI annual waves
are retrospective frozen-model outputs, **finite at ordinal zero because
fitting already finished**. Repeated correlation/component-count arrays
are model diagnostics, not estimates from partial live years.

Annual cases assign INPUT_CLOSE an artificial 2020 weekday calendar.
Superposition cases use six synthetic histories for target 2021:
source j has circular rotation 7*j and scale 1+0.1*j for j=0..5,
in lag order 8,11,19,27,30,38. The extension case truncates source-model
normalization intervals/dates to 250, then emits 252 target ordinals.
These are deliberate port tests, not historical market models or the book's
2008 chart. Example provenance is recorded in example-notes.md.

## Run and verify

~~~sh
python3 dynamic-cycle-explorer.py
python3 dynamic-cycle-explorer.py --generate-reference-data
~~~

No flags runs all reference configurations. The regeneration flag
idempotently saves/verifies Python first, then writes other languages.
The exact input closes are embedded in __main__.

From the parent directory run test_research_masters.py,
test_optional_masters.py and verify_references.py --native. Checks cover
direct DFT, strength/evidence/phase rules, delayed centered means, exhaustive
subsets against direct Pearson, year completeness/no-lookahead, model/CSV
round trips, normalization and calendar extension. The verifier checks
all core copies, every reference array and native Go/Rust/Zig/Node loading.
Native probes validate fixtures rather than five independently ported
indicators.

## Runnable CLI examples

~~~sh
python3 dynamic-cycle-explorer.py analyze --input example-prices.csv --json-out scan.json --series-out series.csv
python3 dynamic-cycle-explorer.py analyze --input example-prices.csv --rolling --json-out rolling.json --series-out rolling.csv
~~~

Input CSV is date,close with finite prices and increasing unique ISO dates.
--start/--end select an inclusive range; batch uses all selected prices.
--window controls rolling only. --horizon projects in bars without inventing
future sessions. JSON unavailable fields are null; CSV missing values are empty.
JSON preserves binary64 precision; CSV uses 15 significant digits.

All CLI commands are local and require no network.

## References and BibTeX

1. Lars von Thienen, *Decoding the Hidden Market Rhythm, Part 1: Dynamic
   Cycles*, 2017: [chapter 3](../../2017-decoding-the-hidden-market-rhythm-1/ch03.md),
   [chapter 6](../../2017-decoding-the-hidden-market-rhythm-1/ch06.md),
   [chapter 9](../../2017-decoding-the-hidden-market-rhythm-1/ch09.md).
2. Lars von Thienen, *Decoding the Hidden Market Rhythm, Part 2: Metonic
   Cycles*, 2017 release, [complete source](../../2017-decoding-the-hidden-market-rhythm-2/content.md),
   chapter 3, annual decoder and historical superposition.
3. Lars von Thienen, Foundation for the Study of Cycles,
   [*The Cycle Scanner Algorithm*](../../the-cycle-scanner-algorithm-whitepaper/content.md),
   Steps 1–4, especially the separately fitted shorter-window phase.
   Supplied text does not state publication year.
4. [Downloaded numerical extracts](../../2017-decoding-the-hidden-market-rhythm-1/code/):
   hp_filter.cpp, goertzel.cpp, BartelsScore.cpp. WTT scanner/explorer/phasing
   C# extracts are API clients rather than local numerical authorities.
5. Robert J. Hodrick and Edward C. Prescott, “Postwar U.S. Business Cycles:
   An Empirical Investigation,” *Journal of Money, Credit and Banking*,
   29(1), 1997, pp. 1–16.
6. Gerald Goertzel, “An Algorithm for the Evaluation of Finite Trigonometric
   Series,” *The American Mathematical Monthly*, 65(1), 1958, pp. 34–35.
7. Julius Bartels, “Random fluctuations, persistence, and quasi-persistence
   in geophysical and cosmical periodicities,” *Terrestrial Magnetism and
   Atmospheric Electricity*, 40(1), 1935, as cited in the whitepaper.
8. Charles E. Armstrong, “Applying the Bartels Test of Significance to a
   Time Series Cycle Analysis,” 1973, citation supplied by the whitepaper.
9. Brian J. Millard, *Channels and Cycles: A Tribute to J.M. Hurst*,
   named in Part 1 chapter 6; edition/year not established by supplied text.
10. [zpano output contracts](../../../../external/zpano/py/indicators/core/outputs/)
    and [Ehlers prepared examples](../../../john-ehlers/prepared-for-conversion/)
    inform packaging and output structure.

HP/GDFT/Bartels source arithmetic is © 2017 Lars von Thienen / WhenToTrade,
CC BY 4.0. Retain author/book/site attribution and indicate adaptations.
The book text is not presumed to inherit the code-extract license.
The pipeline/phase/annual normalization choices above are documented
research adaptations.

~~~bibtex
@book{vonThienen2017Dynamic,
  author={von Thienen, Lars},
  title={Decoding the Hidden Market Rhythm, Part 1: Dynamic Cycles},
  year={2017}, url={https://www.whentotrade.com/},
  note={Supplied chapter archive and numerical extracts}
}
@book{vonThienen2017Metonic,
  author={von Thienen, Lars},
  title={Decoding the Hidden Market Rhythm, Part 2: Metonic Cycles},
  year={2017}, edition={2},
  note={Supplied complete local text; chapter 3 annual decoding and superposition}
}
@techreport{vonThienenCycleScanner,
  author={von Thienen, Lars}, title={The Cycle Scanner Algorithm},
  institution={Foundation for the Study of Cycles},
  note={Supplied whitepaper; publication date not stated}
}
@misc{vonThienenNumericalExtracts2017,
  author={von Thienen, Lars}, year={2017},
  title={HP Filter, Goertzel and Bartels Score numerical extracts},
  url={https://www.whentotrade.com/}, note={Downloaded local sources; CC BY 4.0}
}
@article{hodrickPrescott1997,
  author={Hodrick, Robert J. and Prescott, Edward C.},
  title={Postwar U.S. Business Cycles: An Empirical Investigation},
  journal={Journal of Money, Credit and Banking},
  year={1997}, volume={29}, number={1}, pages={1--16}
}
@article{goertzel1958,
  author={Goertzel, Gerald},
  title={An Algorithm for the Evaluation of Finite Trigonometric Series},
  journal={The American Mathematical Monthly},
  year={1958}, volume={65}, number={1}, pages={34--35}
}
@article{bartels1935,
  author={Bartels, Julius},
  title={Random fluctuations, persistence, and quasi-persistence in geophysical and cosmical periodicities},
  journal={Terrestrial Magnetism and Atmospheric Electricity},
  year={1935}, volume={40}, number={1}
}
@misc{armstrong1973,
  author={Armstrong, Charles E.},
  title={Applying the Bartels Test of Significance to a Time Series Cycle Analysis},
  year={1973}, note={Bibliographic detail as supplied by the whitepaper}
}
@book{millardChannelsCycles,
  author={Millard, Brian J.}, title={Channels and Cycles: A Tribute to J.M. Hurst},
  note={Named in supplied Part 1 chapter 6; edition/year not established}
}
@misc{zpanoOutputContracts,
  title={zpano Indicator Core Output Types},
  note={Local external/zpano/py/indicators/core/outputs}
}
@misc{preparedEhlersMasters,
  title={Prepared Ehlers Streaming Research Prototype Masters},
  note={Local inputs/john-ehlers/prepared-for-conversion}
}
~~~
