# Performance measures

```bibtex
@book{bacon2023,
  author    = {Bacon, Carl R.},
  title     = {Practical Portfolio Performance Measurement and Attribution},
  edition   = {3rd},
  publisher = {John Wiley \& Sons},
  year      = {2023},
  month     = {January},
  day       = {31},
  isbn      = {9781119831945},
  isbn10    = {1119831946},
  pages     = {560},
  series    = {The Wiley Finance Series}
}
@misc{braverock,
  title        = {PerformanceAnalytics: Econometric Tools for Performance and Risk Analysis},
  author       = {Brian G. Peterson and Peter Carl and Kris Boudt and Ross Bennett and Joshua Ulrich and Eric Zivot and Dries Cornilly and Eric Hung and Matthieu Lestel and Kyle Balkissoon and Diethelm Wuertz and Anthony Alexander Christidis and R. Douglas Martin and Zeheng Zenith Zhou and Justin M. Shea and Dhairya Jain and Talgat Daniyarov},
  note         = {Last accessed: July 10, 2026},
  url          = {https://github.com/braverock/PerformanceAnalytics},
  publisher    = {GitHub}
}
@Manual{braverockCran,
  title        = {PerformanceAnalytics: Econometric Tools for Performance and Risk Analysis},
  author       = {Brian G. Peterson and Peter Carl and Kris Boudt and Ross Bennett and Joshua Ulrich and Eric Zivot and Dries Cornilly and Eric Hung and Matthieu Lestel and Kyle Balkissoon and Diethelm Wuertz and Anthony Alexander Christidis and R. Douglas Martin and Zeheng Zenith Zhou and Justin M. Shea and Dhairya Jain and Talgat Daniyarov},
  year         = {2026},
  note         = {R package version 2.1.0. Last accessed: July 10, 2026},
  url          = {https://cran.r-project.org/web/packages/PerformanceAnalytics/index.html},
  doi          = {10.32614/CRAN.package.PerformanceAnalytics},
  publisher    = {Comprehensive R Archive Network (CRAN)}
}
```

## Strteaming min/max

Suppose we process a sequence of numbers $x_0,x_1,x_2,...$ one value at a time.
After each new number arrives, it can immediately determine the minimum and maximum values either:

- over all values seen so far (the running or unbounded window case in which we "remember everything"), or
- over only the last $w$ values (the sliding or bounded window case in which we "remember only the last $w$ values").

### The unbounded case

The unbounded case is very simple. After observing $x_0,...,x_n$ we store two numbers:
$$m_n=\min_{0\le i\le n}{x_i}$$
and
$$M_n=\max_{0\le i\le n}{x_i}$$
Updating is trivial:
$$
\begin{array}{lcl}
m_{n+1} &=& \min(Mmn, x_{n+1})\\
M_{n+1} &=& \max(M_n, x_{n+1})
\end{array}
$$

### Why the sliding window is harder

The interesting part is the sliding window case. Suppose the window size is $w=4$.
After receiving $x_0,x_1,x_2,x_3,x_4,x_5,...$ the current minimum should be
$\min\{x_{n-3},x_{n-2},x_{n-1},x_n\}$
The obvious solution would recompute the minimum over the window every time.
That costs $O(w)$ operations per update.
By trying a different approach, we can achieves $O(1)$ amortized time.

The key idea is that we never store values that can never become the minimum (or maximum) in the future.

Instead of storing every value in the window, the algorithm stores only values that could still become the minimum (or maximum).
This is the purpose of the two deques.

### The monotonic deque

A deque ("double-ended queue") is simply a list where one can efficiently insert or remove elements at both ends.
The algorithm always inserts at the back. It removes

- old elements from the front,
- dominated elements from the back.

We maintain two deques.

**Minimum deque** stores pairs $(i, x_i)$ where $i$ is the position in the stream.
Its defining property is
$$x_{i_1}\lt x_{i_2}\lt x_{i_3}\lt ...$$
from front to back. The smallest value is always at the front.

**Maximum deque** is similar, but
$$x_{i_1}\gt x_{i_2}\gt x_{i_3}\gt ...$$
rom front to back. The largest value is always at the front.

### Why discard elements?

This is the clever part. Suppose the minimum deque currently contains

| index | value |
| --- | --- |
| 3 | 2 |
| 6 | 5 |
| 8 | 7 |

Now suppose the next observation is $x_9=4$.
The last stored value is $7$. Since $7\ge 4$, the value $7$ is removed.
Why?
Because whenever both values are inside the window,

- $4$ is newer,
- $4$ is smalle.

The value $7$ can never again become the minimum of any future window.
It is permanently dominated. Then compare with $5$.
Again, $5\ge 4$, so $5$ is removed too. Eventually we get

| index | value |
| --- | --- |
| 3 | 2 |
| 9 | 4 |

The deque remains increasing.

### Why this works

Imagine two observations $x_i\gt x_j$ with $i\lt j$. The later observation is both

- newer,
- smaller.

The earlier one is therefore permanently irrelevant. Whenever $x_i$ is still inside the window, $x_j$ is too.
Since $x_j\lt x_i$, $x_i$ can never become the minimum.
Hence it may safely be discarded immediately.
This mathematical invariant is maintained by code.

### Example

Suppose the window size is three. The stream is $4, 2, 5, 3, 1$.

Insert 4. Minimum deque: $[(0,4)]$, minimum = 4.

Insert 2. Since $4\ge 2$, 4 can never again be a minimum. Remove it. Deque: $[(1,2)]$, minimum = 2.

Insert 5. Since $5\gt 2$, keep both. Deque: $[(1,2),(2,5)]$, minimum remains 2.

Insert 3. Now compare with the back. Since $5\ge 3$, remove 5. Deque becomes $[(1,2),(3,3)]$.

Insert 1. First remove elements that have left the window. Then compare from the back.
Both $3\ge 1$ and $2\ge 1$, so both disappear. Deque: $[(4,1)]$, minimum is immediately available.

### Why store the indices?

Suppose the window size is $w=5$. When the current index is $n$, the valid indices are $n-4,...,n$.
Anything older must disappear. We compute $cutoff = index - window_size$ and removes every stored element with $stored_index\le cutoff$.
Those observations are outside the current window. Without storing indices, there would be no way to know whether a value is still valid.

### Why does the minimum stay at the front?

The deque satisfies two invariants.

1. Indices increase
   lements are always appended at the back. Thus $i_1\lt i_2\lt ...$.
2. Values increase
   Whenever a smaller value arrives, larger values at the back are removed. Thus $x_{i_1}\gt x_{i_2}\gt ...$.
   Therefore the front element is simultaneously
   - the oldest surviving candidate,
   - the smallest surviving value.
   Hence it is exactly the minimum of the current window.

The same argument, with inequalities reversed, applies to the maximum deque.

### Complexity

Each observation is

- inserted exactly once;
- removed at most once.

Although an update may occasionally remove many elements, each element can only be removed once over the entire lifetime of the algorithm.

Therefore, for $n$ updates:

- total insertions: $n$,
- total removals: at most $n$.

So the total work is proportional to $n$, giving an amortized cost of $O(1)$ per update.

The memory usage is at most $O(w)$, where $w$ is the window size.

### A mathematical interpretation

The algorithm maintains two sets of undominated candidates.

For the minimum deque, an observation $(i,x_i)$ is discarded as soon as there exists a later observation $(j,x_j)$ with $j\gt i,\ \ \ x_j\le x_i$.
The later observation dominates the earlier one because it is both more recent and no larger.
The dominated observation can never again be the minimum of any future window.

The maximum deque uses the analogous dominance relation $j\gt i,\ \ \ x_j\ge x_i$.

Viewed this way, the deques are not arbitrary data structures but compact representations of the *Pareto-optimal candidates* under the partial order induced by "more recent" and "smaller" (or "larger").
Maintaining these candidate sets is what allows the class to answer sliding-window minimum and maximum queries in constant amortized time.

The deque therefore stores exactly the undominated candidates, ordered by time.
This elegant invariant is what allows the current minimum (or maximum) to be read directly from the front of the deque in constant time.

## Cumulative returns

### Geometric return vs geometric mean

These are often confused. Suppose returns are 0.1, 0.2,-0.05.

The cumulative geometric return
$$R_g=\prod_i{(1+r_i)}-1$$
answers the question "How much money did I make overall?"

The geometric mean
$$\bar{R_g}=\left(\prod_i^n{(1+r_i)}\right)^{1/n}-1 = \exp(\frac{1}{n}\sum_i{\ln(1+r_i)})-1$$
answers the question "What constant return each period gives the same ending wealth?"

Note that geometric mean is always less than geometric mean, being equal only when every return is identical
$$\bar{R_g} \le \bar{R_a}$$
where
$$\bar{R_a}=\frac{1}{n}\sum_i^n{r_i}$$

### Calculating streaming cumulative geometric returns

Suppose a sequence of simple returns $r_1,r_2,...,r_n,\ \ r_i\gt -1$. The cumulative growth factor is
$$G_n=\prod_{i=1}^n{1+r_i}$$,
and the cumulative geometric return is $R_n=G_n-1$.
Instead of computing the product directly, we can store
$$L_n=\sum_{i=1}^n{\log(1+r_i)}$$.
Since
$$\log\left(\prod_{i=1}^n{(1+r_i)}\right)=\sum_{i=1}^n{\log(1+r_i)}$$,
we recover the quantities of interest through
$$G_n=e^{L_n}$$, and $$R_n=e^{L_n}-1$$
Thus, internally we never multipliy growth factors; we only accumulate logarithms.

#### Streaming updates

When a new return r arrives, we compute $\ell=\log(1+r_i)$ (using `log1p(r)`, which is a numerically accurate implementation of `log(1+r)`, especially when `r` is very small).

The cumulative logarithmic return is updated as
$$L_{new}=L_{old}+\log(1+r)$$.
No previous returns need to be revisited, so each update has constant computational cost.

#### Numerically stable summation

The cumulative logarithmic return is stored in a Klein-Kahan�Babu�ka-N accumulator.

Mathematically this still represents
$$L_n=\sum_{i=1}^n{\log(1+r_i)}$$
but the accumulator uses a compensated summation algorithm (a refinement of the Kahan�Babu�ka family) to reduce the rounding errors that occur when many floating-point numbers are added.

This is particularly useful for long return series, where ordinary floating-point summation gradually accumulates numerical error.

#### Removing observations

The method `revert(ret)` subtracts the logarithmic contribution of one observation:
$$L_{new}=L_{old}-\log(1+r)$$.
This allows the class to support sliding windows, where the oldest observation is removed as a new one enters.

#### Recovering cumulative return

The property `cumulative_geometric_return` returns
$$e^{L_n}-1$$.
Instead of computing `exp(L)-1`, we use `expm1(L)`, which computes the same quantity with higher numerical accuracy when `L` is close to zero.

#### Recovering the cumulative growth factor

The property `geometric_return_plus_1` returns
$$G_n=e^{L_n}$$.
This is simply the cumulative wealth multiplier. For example,

| Growth factor | Cumulative return |
| --- | --- |
| 0.9 | -10% |
| 1.0 | 0% |
| 1.15 | 15% |

#### Geometric mean return

If $n$ observations have been processed, then
$$L_n=\sum_{i=1}^n{\log(1+r_i)}$$.
The average logarithmic return is $L_n/n$. Exponentiating gives
$$\exp\left(\frac{L_n}{n}\right)=\left( \prod_{i=1}^n{(1+r_i)} \right)^{1/n}$$,
which is the geometric mean growth factor. Subtracting one gives
$$\left( \prod_{i=1}^n{(1+r_i)} \right)^{1/n}-1$$,
the geometric mean return (or CAGR when the observations correspond to one time unit).
Again, `expm1()` is used for improved numerical accuracy.

#### Tracking extrema

We also maintains the cumulative logarithmic return after every update, $L_1,L_2,...,L_n$.
Rather than storing the cumulative growth factors
$$e^{L_1},e^{L_2},...,e^{L_n}$$,
we store only the logarithmic values.
This is mathematically sufficient because the exponential function is strictly increasing:
$$L_a\lt L_b\Longleftrightarrow e^{L_a}\lt e^{L_b}$$.
Therefore,
$$\min_i e^{L_i}=e^{min_i L_i}$$,
and similarly,
$$\max_i e^{L_i}=e^{max_i L_i}$$.
The minimum and maximum are therefore maintained in logarithmic space, and exponentiation is performed only when the corresponding properties are requested.
This avoids an exponential evaluation on every update while producing exactly the same result.

#### Computational complexity

Each update performs only:

- one logarithm,
- one compensated addition,
- one min/max update,

all of which have constant complexity.

Consequently:

- Update: $O(1)$
- Revert: $O(1)$
- Queries: $O(1)$

regardless of the number of observations already processed.

In summary, the class implements streaming cumulative geometric returns by exploiting the identity
$$\prod_{i=1}^n{(1+r_i)}=\exp\left(\sum_{i=1}^n{\log(1+r_i)}\right)$$.
All internal state is maintained in logarithmic space, where products become sums, compensated summation improves numerical accuracy, and extrema can be tracked without repeated exponentiation.
Only when a cumulative return, growth factor, or geometric mean is requested is the logarithmic quantity transformed back via the exponential function.
This separation between the internal logarithmic representation and the externally reported geometric quantities makes the algorithm both mathematically elegant and numerically robust.

### The geometric mean return and the Compound Annual Growth Rate (CAGR)

The geometric mean return is the constant per-period return that would produce the same cumulative growth.

CAGR (Compound Annual Growth Rate) is the geometric mean return per year.

Suppose you have $n$ periodic returns $r_1,...,r_n$. Then
$$G=\prod_{i=1}^n{(1+r_i)}$$
is the cumulative growth factor. The per-period geometric mean return is
$$\left(\prod_{i=1}^n{(1+r_i)}\right)^{1/n}-1$$
because
$$\exp\left(\frac{1}{n}\sum_{i=1}^n{\log(1+r_i)}\right)-1=\left(\prod_{i=1}^n{(1+r_i)}\right)^{1/n}-1$$.
If each observation is one year, then this quantity is the CAGR.
If each observation is one month, then it is the monthly geometric mean return, not the CAGR. The CAGR would be
$$\left(\prod_{i=1}^n{(1+r_i)}\right)^{12/n}-1=(1+\bar{r}_{geom.month})^{12}-1$$.
Similarly:

- Daily returns -- exponent 252 (trading days in a year),
- Weekly returns -- exponent 52,
- Quarterly returns -- exponent 4.

### Cumulative vs geometric mean

The term cumulative refers to compounding over observations:
$$\prod_{i=1}^n{(1+r_i)}-1$$
The geometric mean refers to the average per observation:
$$\left(\prod_{i=1}^n{(1+r_i)}\right)^{1/n}-1$$.
These are different concepts. A mean is inherently an average, so "cumulative mean" is mathematically contradictory.

| Property | Mathematical quantity |
| --- | --- |
| cumulative_geometric_return | $$\prod_{i=1}^n{(1+r_i)}-1$$ |
| geometric_mean_return | $$\left(\prod_{i=1}^n{(1+r_i)}\right)^{1/n}-1$$ |
| annualized_return(periods_per_year) or cagr(periods_per_year) | $$\left(\prod_{i=1}^n{(1+r_i)}\right)^{p/n}-1$$ |

where $p$ is the number of observations per year (252, 52, 12, 4, 1, ...).

The annualization is
$$(1+r_{geom})^p=\exp\left(\frac{p}{n}\sum_{i=1}^n{\log(1+r_i)}\right)-1$$.

## Skewness and Kurtosis

Thehe are the standard statistical properties of a distribution.
How do we calculate them?

### Central moments

We maintain 'streaming' raw moments ($E[X^k]$)

$$\begin{array}{lcl}
\mu_1^{'} &=& \sum{x_i}\\
\mu_2^{'} &=& \sum{x_i^2}\\
\mu_3^{'} &=& \sum{x_i^3}\\
\mu_4^{'} &=& \sum{x_i^4}
\end{array}$$

updating these sums every time we add a new $x$.
To get central moments $\mu_k$ from the raw ones $\mu_k^{'}$, we calculate

\begin{array}{lcl}
\mu_1 &=& \frac{\mu_1^{'}}{n}\\
\mu_2 &=& \frac{\mu_2^{'}}{n}-\mu_1^2\\
\mu_3 &=& \frac{\mu_3^{'}}{n}-\mu_1^3-3\mu_1 \cdot \mu_2\\
\mu_4 &=& \frac{\mu_4^{'}}{n}-\mu_1^4-6\mu_1^2\mu_2-4\mu_3\mu_1
\end{array}

where

- $\mu_1$ is the 1st central moment (mean)
- $\mu_2$ is the 2nd central moment (population variance)
- $\mu_3$ is the 3rd central moment (measures asymmetry)
- $\mu_4$ is the 4th central moments (measures tail weight)

Since these moments have different units ($\mu_r$ has units of $x^r$), statisticians usually normalize them.

Karl Pearson defined
$$\beta_1=\frac{\mu_3^2}{\mu_2^3}$$
which is a nonnegative measure of skewness, and
$$\beta_2=\frac{\mu_4}{\mu_2^2}$$
which is Pearson kurtosis.
The higher-order quantities ($\beta_3$, $\beta_4$, ...) can be defined analogously from higher central moments,
but they are rarely used outside specialized statistical literature.

The $\beta_1$ and $\beta_2$ are unsigned, people generally prefer the signed versions:
$$\gamma_1=\frac{\mu_3}{\mu_2^{3/2}}$$
which is the familiar skewness coefficient. Since $\gamma_1^2=\beta_1$, $\gamma_1$ is simply the signed square root of Pearson's $\beta_1$

Similarly,
$$\gamma_2=\beta_2-3=\frac{\mu_4}{\mu_2^2}-3$$
which is excess kurtosis. The subtraction of 3 makes the normal distribution have $\gamma_2=0$, instead of $\beta_2=3$.

Why 3? For a normal distribution, $\mu_4=3\mu_2^2$, so $\beta_2=3$.
Subtracting 3 makes the normal distribution the natural zero point

- $\gamma_2>0$: heavier tails than normal (leptokurtic)
- $\gamma_2=0$: normal (mesokurtic)
- $\gamma_2<0$: lighter tails than normal (platykurtic)

### Skewness

Skewness is the degree of asymmetry of a distribution around its mean.

There are three variants of skewness we calculate.
Two of them are determined by the `bias` boolean parameter we set during construction time.
The `skewness` property calculates either the 'moment' or the 'fisher' skewness.
The third variant, 'sample' skewness, can be only accessed via dedicated `skewness_sample` property.

The 'moment' skewness is calculated as
$$\gamma_1=\frac{\mu_3}{\mu_2^{3/2}}$$
This calculation method is used when the `bias` boolean parameter is set to `true`.
Use the `skewness_moment` property to get the 'moment' skewness regardless of the value of the `bias` parameter.

The 'fisher' skewness is calculated as
$$\gamma_1\cdot \frac{\sqrt{n(n-1)}}{n-2}$$
where $\gamma_1$ is the 'moment' skewness.
This calculation method is used when the `bias` boolean parameter is set to `false`.
Use the `skewness_fisher` property to get the 'fisher' skewness regardless of the value of the `bias` parameter.

The 'sample' skewness is calculated as
$$\gamma_1\\cdot \\frac{n^2}{(n-1)(n-2)}$$
where $\gamma_1$ is the 'moment' skewness.
It is not related to the `bias` parameter, use the `skewness_sample` property to get it.

### Kurtosis

Kurtosis is the degree to which a distribution peak compared to a normal distribution.

There are five variants of kurtosis we calculate.
Four of them are determined by the `bias` and `fisher` boolean parameters we set during construction time.
The following table shows the variants of kurtosis returned by the `kurtosis` property depending on the values of these `bias` and `fisher` parameters.

| bias | fisher | variant | dexcription |
| --- | --- | --- | --- |
| true | true | 'excess' | biased excess kurtosis |
| true | false | 'moment' | biased Pearson (population) kurtosis |
| false | true | 'sample excess' | unbiased excess kurtosis |
| false | false | 'sample' | unbiased Pearson kurtosis |

The 'moment' kurtosis is calculated as
$$\beta_2=\frac{\mu_4}{\mu_2^2}$$
Use the `kurtosis_moment` property to get it regardless of the value of the `bias` and `fisher` parameters.

The 'excess' kurtosis is calculated as
$$\beta_2-3=\frac{\mu_4}{\mu_2^2}-3$$
Use the `kurtosis_excess` property to get it regardless of the value of the `bias` and `fisher` parameters.

The 'sample excess' kurtosis is calculated as
$$\frac{(n^2-1)\beta_2-3(n - 1)^2}{(n - 2)(n - 3)}$$
Use the `kurtosis_sample_excess` property to get it regardless of the value of the `bias` and `fisher` parameters.

The 'sample' kurtosis is calculated as
$$\frac{(n^2 - 1)\beta_2}{(n-2)(n-3)}$$
This variant is compatible with `PerformanceAnalytics` R package which is not computing `unbiased Pearson kurtosis` as `unbiased excess + 3`, but using the historical sample kurtosis definition widely used in finance literature, where the correction factor is applied directly to the Pearson kurtosis.
This formula is not mathematically equivalent to the next one, used by SciPy (bias=False), SAS, and Joanes & Gill Type 2:
$$\frac{(n^2-1)\beta_2-3(n - 1)^2}{(n - 2)(n - 3)}+3$$
It differs by
$$\frac{9n-15}{(n-2)(n-3)}$$
from the R package version mentioned above.
The difference is approximately $0.44$ for $n=24$.
Use the `kurtosis_sample_corrected` property to get `PerformanceAnalytics` R package variant regardless of the value of the `bias` and `fisher` parameters.
Use the `kurtosis_sample` property to get SciPy (bias=False), SAS, and Joanes & Gill Type 2 version regardless of the value of the `bias` and `fisher` parameters.

## Skewness-Kurtosis ratio

The Skewness-Kurtosis ratio is a specialized statistical metric used to refine portfolio evaluation beyond standard mean-variance analysis.
It is calculated by dividing the skewness (asymmetry) of a return distribution by its kurtosis (tailedness).

The Skewness-Kurtosis ratio is a statistical metric calculated by dividing the skewness of a distribution by its kurtosis.
It is primarily used in finance and econometrics to rank portfolios in conjunction with the Sharpe ratio.

The ratio is defined as: $SKR=\frac{S}{K}$ where $S$ represents the skewness and $K$ represents the kurtosis (often excess kurtosis) of the asset or portfolio returns.

Interpretation: A higher ratio indicates a more favorable risk-adjusted profile, suggesting that the returns have a better balance of asymmetry relative to their tail extremity.

Application: It helps investors assess whether a portfolio�s return distribution offers adequate compensation for the skew and kurtosis (tail risk) present, complementing traditional mean-variance analysis.

The metric serves as a diagnostic tool for the shape of the return distribution:

- Higher Ratio: Generally preferred. It indicates a favorable balance where positive asymmetry (upside potential) is high relative to the extremity of the tails (risk of outliers).
- Lower or Negative Ratio: Suggests undesirable characteristics, such as negative skewness (propensity for large losses) or excessively high kurtosis (fat tails implying frequent extreme events).

Because financial returns often deviate from a normal distribution, this ratio helps investors distinguish between assets that may have similar means and variances but vastly different risk profiles regarding extreme outcomes.

The Skewness-Kurtosis ratio is not typically used in isolation but rather in conjunction with the Sharpe ratio.

- Complement to Sharpe: While the Sharpe ratio evaluates return per unit of volatility (standard deviation), it assumes a normal distribution. The Skewness-Kurtosis ratio adjusts this view by accounting for the third and fourth moments of the distribution.
- Rational Selection: Research suggests that for a rational, risk-averse investor (specifically one with a Constant Absolute Risk Averse utility function), selecting portfolios based on higher values of this ratio (alongside other metrics) can better maximize utility when returns are non-Gaussian. It helps identify portfolios that offer "lottery-like" upside (positive skew) without excessive tail risk (high kurtosis).

### GPT

A concise description could be:
Skewness-Kurtosis Ratio: The ratio of sample skewness to sample kurtosis. Positive values indicate that positive asymmetry dominates relative to tail heaviness, while negative values indicate downside asymmetry. This is a descriptive statistic of the return distribution rather than a standard risk-adjusted performance measure.

The Skewness-Kurtosis ratio is a niche metric rather than a standard performance measure like the Sharpe ratio or Sortino ratio. It is occasionally used by quantitative traders to summarize the shape of a return distribution, especially to distinguish strategies with asymmetric payoffs from those with fat tails.

The ratio is typically defined as $\frac{skewness}{kurtosis}$ or, less commonly, $\frac{skewness}{excess kurtosis}$.
Unfortunately, there is no universally accepted definition in the literature, so it's important to document exactly which version is used.

#### Intuition
Recall the meanings of the two moments:
- Skewness ($\gamma_1$) measures asymmetry.
  - Positive: occasional large gains.
  - Negative: occasional large losses.
- Kurtosis ($\gamma_2$) measures tail heaviness.
  - High kurtosis means more extreme returns, in either direction.

The ratio attempts to measure how much desirable asymmetry (positive skew) is obtained per unit of tail risk.

#### Interpretation
Suppose four trading strategies have identical mean and volatility.

| Strategy | Skewness | Kurtosis | Ratio |
| --- | --- | --- | --- |
| A | 1.0 | 3 | 0.33 |
| B | 1.0 | 10 | 0.10 |
| C | -1.0 | 3 | -0.33 |
| D | 0.5 | 2 | 0.25 |

Interpretation:

- Large positive ratio
  - positive skew
  - relatively modest tail risk
  - generally desirable
- Ratio near zero
  - either little skew
  - or skew overwhelmed by heavy tails
- Negative ratio
  - negative skew
  - downside tail events dominate
  - characteristic of many option-selling strategies

#### Why traders sometimes look at it
Many performance measures ignore higher moments.
For example, the Sharpe ratio treats

- many small gains with occasional crashes, and
- many small losses with occasional windfalls

as equally risky if their variances are equal.
The Skewness-Kurtosis ratio attempts to distinguish these situations.

For example:

- Trend-following (CTA)
  - positive skew
  - moderate kurtosis
  - relatively high ratio
- Short-volatility
  - negative skew
  - high kurtosis
  - strongly negative ratio

This aligns well with practitioners' intuition.

#### Limitations
The ratio has several important drawbacks.

- Not theoretically derived.
  Unlike the Sharpe ratio, it is not the solution to an optimization problem or utility maximization.
- Sensitive to estimation error.
  Both skewness and kurtosis are third- and fourth-order sample moments.
  They require much larger samples than the mean or variance to estimate reliably.
- Can be misleading.
  Two distributions may have the same ratio but very different risk characteristics.

For example,

| Skewness | Kurtosis | Ratio |
| --- | --- | --- |
| 0.5 | 2 | 0.25 |
| 2.5 | 10 | 0.25 |

These distributions are clearly not equally attractive.

- Undefined or unstable.
  If excess kurtosis is used, $\gamma_1/\gamma_2$, the denominator may be close to zero, producing very large values.

#### Better alternatives

For evaluating trading strategies, practitioners usually prefer metrics that incorporate higher moments more directly:

- Adjusted Sharpe Ratio (Favre & Galeano), which adjusts the Sharpe ratio using skewness and kurtosis.
- Omega Ratio, which considers the entire return distribution above and below a threshold.
- Kappa Ratios, a generalization of the Sortino ratio.
- Expected Shortfall (CVaR) combined with skewness, which provides a clearer view of downside tail risk.

## Jarque�Bera normality test statistic

The test was introduced by Carlos M. Jarque and Anil K. Bera, two econometricians.
Their influential paper is:
> Jarque, C. M., & Bera, A. K. (1980). Efficient tests for normality, homoscedasticity and serial independence of regression residuals. Economics Letters, 6(3), 255�259.

The paper developed several specification tests for regression diagnostics, with the normality test becoming by far the most widely used. Although originally proposed for regression residuals, it is now routinely applied to any sample where assessing normality is of interest, including financial returns.

Assuming:
- the biased (population moment) skewness $\gamma_1=\mu_3/\mu_2^{3/2}$
- the biased excess kurtosis $\gamma_2=\mu_4/\mu_2^2-3$
- $n$is the sample size
the classical Jarque�Bera statistic:
$$JB=\frac{n}{6}\left(\gamma_1^2+\frac{\gamma_2^2}{4}\right)$$.

The `JB` test is an asymptotic test. The approximation $JB\sim \chi_2^2$ is only good when $n$ is reasonably large.
For financial returns this is usually acceptable once you have roughly 100�200 observations, and very good with several hundred.
For very small samples (say $n\lt 50$), the test tends to reject normality too often.

### What is the intuition?

A normal distribution is completely characterized by two numbers:
- mean
- variance

Once those are fixed,
- skewness must equal 0
- excess kurtosis must equal 0

The Jarque�Bera test simply asks:
> How far are the observed skewness and kurtosis from what a normal distribution should have?

If both are close to zero, the statistic is small.
If either is far from zero, the statistic becomes large.

### Why these particular constants?

uppose the data really are normal.

For large samples, $\sqrt{n}\gamma_1$ behaves approximately like a standard normal variable with variance 6.
Likewise, $\sqrt{n}\gamma_2$ has approximate variance 24.

Equivalently, $$\frac{n\gamma_1^2}{6}$$ is approximately $\chi_1^2$ and
$$\frac{n\gamma_2^2}{24}$$ is also approximately $\chi_1^2$.
Adding two independent chi-square variables with one degree of freedom gives $\chi_2^2$.
That is exactly
$$JB=\frac{n}{6}\left(\gamma_1^2+\frac{\gamma_2^2}{4}\right)$$.
So the constants 6 and 4 are not arbitrary�they come from the asymptotic variances of the sample skewness and kurtosis under normality.

### A geometric interpretation

Think of the vector $(\gamma_1,\gamma_2}$ as a point in the plane.
For a perfectly normal distribution, the true point is $(0,0)$.
But because of sampling variability, even normally distributed data produce points scattered around the origin.

The scatter is not circular:
- skewness fluctuates with variance $6/n$,
- kurtosis fluctuates with variance $24/n$.

The JB statistic rescales each coordinate by its natural variance and computes the squared distance:
$$JB=\left(\frac{\sqrt{n}\gamma_1}{\sqrt{6}}\right)^2+\left(\frac{\sqrt{n}\gamma_2}{\sqrt{24}}\right)^2$$.
So it is essentially the squared Mahalanobis distance from the origin using the asymptotic covariance matrix.

### Why does it detect non-normality?

Suppose returns are symmetric but heavy-tailed.
Then $g_1\approx 0$, $g_2\gt 0$. Only the kurtosis term contributes.

Suppose returns are asymmetric but otherwise normal-looking.
Then $g_1\neq 0$, $g_2\approx 0$. Only the skewness term contributes.

Suppose both occur. Both terms become large, and the statistic grows rapidly.

## Value at Risk

### Lognormal VaR

Despite the name, this isn't fundamentally different from Gaussian VaR.
It starts from the assumption $\log(S_T/S_0)\sim N(\mu,\sigma)$ instead of $r\sim N(\mu,\sigma)$.

The loss threshold is then computed by exponentiating the normal quantile. Very roughly,

- Gaussian: $VaR = -(\mu + z\sigma)$
- Lognormal: $VaR = 1 - \exp(\mu + z\sigma)$

or an equivalent form depending on whether $\mu$ is arithmetic or logarithmic.

Advantages:

- Better when returns are relatively large (monthly/yearly).
- Never predicts losses greater than 100%.
- Consistent with Black-Scholes assumptions.

Disadvantages:

For daily returns, $\exp(x)\approx 1 + x$, so Gaussian and lognormal VaR are almost identical.

If your library mostly targets daily strategy returns, I wouldn't rush to add it.

### Kernel VaR

Instead of assuming a normal distribution, estimate the probability density using kernel density estimation (KDE).

Instead of

```text
      /\
     /  \
```

you estimate

```text
    /\      /\
___/  \____/  \___
```

matching the observed data.

Advantages:

- No parametric assumptions.
- Smooth estimate.
- Better than historical VaR for moderate sample sizes.

Disadvantages:

- Needs the entire sample.
- Needs bandwidth selection.
- Not naturally streaming.

### Generalized Pareto Distribution (GPD)

This is the "serious" method for estimating very rare losses.
Instead of modeling the whole return distribution, fit only the extreme tail.
For example,

```text
returns

^^^^^^^^^^^^^^^^^^^^^^
                  |
             threshold
                  |
                  v

fit only here
```

using the Generalized Pareto Distribution.

This comes from Extreme Value Theory.

Advantages:

- Excellent for 99.9% or 99.99% VaR.
- Can estimate losses never observed.

Disadvantages:

- Need enough tail observations.
- Threshold selection.
- Maximum-likelihood fitting.
- Much more mathematics.

If I were writing a professional risk library, I'd eventually implement this.

### Monte Carlo

Here you assume a model $returns tomorrow = today + random shock$.
then simulate 10000, 50000 or 100000 future paths,
then compute VaR from the simulated losses.

The quality depends entirely on the simulation model. For example

- Geometric Brownian Motion
- GARCH
- Heston
- Jump diffusion

Monte Carlo itself isn't a distribution.
It's a way of computing VaR once you've chosen a model.

## Conditional Value at Risk vs Expected Shortfall

In quantitative finance today, Expected Shortfall (ES) and Conditional Value at Risk (CVaR) are almost always treated as the same risk measure.
If you implement one, you've effectively implemented both.

There is a bit of historical nuance, though.

### Value at Risk (VaR)

VaR answers:
> How bad can things get with X% confidence?

For 95% confidence: $VaR_{95}$ = 5th percentile of returns (with sign inverted if expressing losses).

It tells you the threshold, but nothing about losses beyond it.

### Expected Shortfall (ES)

ES answers:
> If I end up in the worst 5% of cases, what is my average loss?

Mathematically, $ES_\alpha =-E[R | R\le q_\alpha]$ where $q_\alpha$ s the lower-tail quantile.

So ES is simply the mean of all returns worse than the VaR threshold.

#### Example

Returns: -15%, -12%, -10%, -8%, -5% ...

If $VaR_{95} = 8%$ then $ES_{95} = average(8%, 10%, 12%, 15%) = 11.25%$

ES is always at least as large as VaR.

### Conditional VaR (CVaR)

Originally, Conditional VaR meant exactly the same quantity:
> VaR conditioned on being in the tail.

In modern literature, $CVaR == ES$. Most papers use one or the other interchangeably.

### Why two names?

History. Banks first used VaR.

Researchers noticed VaR has mathematical shortcomings:
- not subadditive
- not coherent
- ignores tail severity

So Expected Shortfall became the preferred coherent alternative.
Meanwhile engineering and optimization literature popularized the name Conditional Value-at-Risk (CVaR).
Different communities, same formula.

### Regulatory terminology

Modern banking regulation (`Basel III` / `Basel IV`) uses Expected Shortfall.
VaR has largely been replaced by ES for market risk capital.
So today's finance literature increasingly says

- VaR
- ES

rather than

- VaR
- CVaR

## Difference between ES.CornishFisher() and operES.CornishFisher()

This is actually a very interesting story.

### Original Cornish-Fisher ES

The original formula is $ES = -\mu + \sigma * MES / \alpha$, where $MES = \Phi(h) � correction(...)$.
This comes directly from integrating the Cornish-Fisher density.
The problem is that it can produce $ES < VaR$ which is impossible.
Expected Shortfall is the average loss beyond VaR. It must satisfy $ES > VaR$ always.
Unfortunately, the Cornish-Fisher approximation is only an approximation, and for sufficiently skewed/heavy-tailed distributions it sometimes violates this inequality.

### Operational ES

Boudt, Peterson and Croux proposed $ES = -\mu - \sigma \min(-MES/\alpha, h)$ instead.
That tiny $\min(...)$ guarantees $ES\ge VaR$ for every parameter combination.
This is what PerformanceAnalytics calls `operational=TRUE` and it's why it's the default.

## Semi-deviation

Semi-deviation measures downside volatility by considering only returns below the sample mean.
It is the square root of the second-order lower partial moment about the mean and is commonly used as the denominator of the Sortino (Downside Sharpe) ratio.

## Downside Frequency

To calculate Downside Frequency, we take the subset of returns that are less than the target (or Minimum Acceptable Returns (MAR)) returns and divide the length of this subset by the total number of returns.
$$DownsideFrequency(MAR)=\frac{\#\{r_i\lt MAR\}}{n}$$
where $n$ is the number of observations of the entire series.

## DownsidePotential

To calculate Downside Potential, we take the returns that are less than the target (or Minimum Acceptable Returns (MAR)) returns and take the differences of those to the target.
We sum and divide by the total number of returns.

Mean of lower partial moments (also called shortfall).

$$DownsidePotential(R, MAR) = \sum^{n}_{t=1}\frac{min[(R_{t} - MAR), 0]} {n}$$

## Downside Deviation

Downside deviation, similar to semi deviation, eliminates positive returns when calculating risk.
To calculate it, we take the returns that are less than the target (or Minimum Acceptable Returns (MAR)) returns and take the differences of those to the target.
We sum the squares and divide by the total number of returns to get a below-target semi-variance.

Downside deviation, similar to semi deviation, eliminates positive returns when calculating risk.
Instead of using the mean return or zero, it uses the Minimum Acceptable Return as proposed by Sharpe (which may be the mean historical return or zero).
It measures the variability of underperformance below a minimum targer rate.
The downside variance is the square of the downside potential.

To calculate it, we take the subset of returns that are less than the target (or Minimum Acceptable Returns (MAR)) returns and take the differences of those to the target.
We sum the squares and divide by the total number of returns to get a below-target semi-variance.

$$DownsideDeviation(R , MAR) = \delta_{MAR} = \sqrt{\sum^{n}_{t=1}\frac{min[(R_{t} - MAR), 0]^2}{n}}$$

where $n$ is either the number of observations of the entire series or the number of observations in the subset of the series falling below the MAR.

SemiDeviation or SemiVariance is a popular alternative downside risk measure that may be used in place of standard deviation or variance.
SemiDeviation and SemiVariance are implemented as a wrapper of DownsideDeviation with MAR=mean(R).

In many functions like Markowitz optimization, semideviation may be substituted directly, and the covariance matrix may be constructed from semideviation or the vector of returns below the mean rather than from variance or the full vector of returns.

In semideviation, by convention, the value of \eqn{n} is set to the full number of observations.
In semivariance the the value of \eqn{n} is set to the subset of returns below the mean.
It should be noted that while this is the correct mathematical definition of semivariance, this result doesn't make any sense if you are also going to be using the time series of returns below the mean or below a MAR to construct a semi-covariance matrix for portfolio optimization.

Sortino recommends calculating downside deviation utilizing a continuous fitted distribution rather than the discrete distribution of observations.
This would have significant utility, especially in cases of a small number of observations.
He recommends using a lognormal distribution, or a fitted distribution based on a relevant style index, to construct the returns below the MAR to increase the confidence in the final result.
Hopefully, in the future, we'll add a fitted option to this function, and would be happy to accept a contribution of this nature.

references
- Sortino, F. and Price, L. Performance Measurement in a Downside Risk Framework. **Journal of Investing**. Fall 1994, 59-65.
- Carl Bacon, **Practical portfolio performance measurement and attribution**, second edition 2008
- Plantinga, A., van der Meer, R. and Sortino, F. The Impact of Downside Risk on Risk-Adjusted Performance of Mutual Funds in the Euronext Markets. July 19, 2001.
  Available at SSRN: \url{https://www.ssrn.com/abstract=277352}
- https://www.sortino.com/htm/performance.htm, see especially end note 10
- https://en.wikipedia.org/wiki/Semivariance

## Downside Risk Measures

Traditional volatility treats positive and negative returns equally. Downside risk measures focus only on returns below a **Minimum Acceptable Return (MAR)**, making them well suited for evaluating investment strategies where upside volatility is not considered risk.

The MAR is specified when constructing the `Ratios` object through the `annual_target_return` parameter.

### `downside_frequency`

The proportion of observed returns below MAR.

This statistic measures **how often** the strategy fails to achieve the target return.

* Range: **0** to **1**
* `0` means no observations are below MAR.
* `1` means every observation is below MAR.

### `downside_potential`

The average shortfall below MAR.

Unlike downside deviation, downside potential measures the **average magnitude** of losses relative to the target without squaring them, making it easier to interpret in the original return units.

### `downside_deviation`

The square root of the second-order lower partial moment ($LPM_2$) about MAR.

Only returns below MAR contribute to the calculation, but the sum is normalized by the **total number of observations**. This is the definition used by PerformanceAnalytics and is commonly used in downside risk measures such as the Sortino ratio.

### `downside_deviation_subset`

An alternative definition of downside deviation.

This statistic differs from `downside_deviation` only in the normalization: the sum of squared shortfalls is divided by the **number of returns below MAR** instead of the total number of observations.

Because only downside observations contribute to the denominator, `downside_deviation_subset` is generally greater than or equal to `downside_deviation`.

| Property                    | Measures                       | Denominator                |
| --------------------------- | ------------------------------ | -------------------------- |
| `downside_frequency`        | Frequency of returns below MAR | Total observations         |
| `downside_potential`        | Mean shortfall below MAR       | Total observations         |
| `downside_deviation`        | RMS shortfall below MAR        | Total observations         |
| `downside_deviation_subset` | RMS shortfall below MAR        | Downside observations only |

## Upside Frequency

Upside frequency of the return distribution.

To calculate Upside Frequency, we take the subset of returns that are more than the target (or Minimum Acceptable Returns (MAR)) returns and divide the length of this subset by the total number of returns.
$$\text{UpsideFrequency}(R , MAR) = \sum^{n}_{t=1}\frac{max[(R_{t} - MAR), 0]}{R_{t}n}$$
where $n$ is the number of observations of the entire series.

## Upside Risk, Variance and Potential

Upside risk, variance and potential of the return distribution

Upside Risk is the similar of semideviation taking the return above the Minimum Acceptable Return instead of using the mean return or zero.

To calculate it, we take the subset of returns that are more than the target (or Minimum Acceptable Returns (MAR)) returns and take the differences of those to the target.
We sum the squares and divide by the total number of returns and return the square root.

$$\text{UpsideRisk}(R , MAR) = \sqrt{\sum^{n}_{t=1}\frac{max[(R_{t} - MAR), 0]^2}{n}}$$
$$\text{UpsideVariance}(R, MAR) = \sum^{n}_{t=1}{\frac{max[(R_{t} - MAR), 0]^2}{n}}$$
$$\text{UpsidePotential}(R, MAR) = \sum^{n}_{t=1}{\frac{max[(R_{t} - MAR), 0]}{n}}$$

where $n$ is either the number of observations of the entire series or the number of observations in the subset of the series falling below the MAR.

The `subset` suffix indicates whether to use the length of the subset of the series below the MAR as the denominator.
Without this suffix, the length of the full series is used.

## Upside Potential Ratio

calculate Upside Potential Ratio of upside performance over downside risk

Sortino proposed an improvement on the Sharpe Ratio to better account for skill and excess performance by using only downside semivariance as the measure of risk.
That measure is the Sortino Ratio. Upside Potential Ratio was a further improvement, extending the measurement of only upside on the numerator, and only downside of the denominator of the ratio equation.

Sortino contends that risk should be measured in terms of not meeting the investment goal.
This gives rise to the notion of `Minimum Acceptable Return` or `MAR`.
All of Sortino's proposed measures include the MAR, and are more sensitive to downside or extreme risks than measures that use volatility(standard deviation of returns) as the measure of risk.

Choosing the MAR carefully is very important, especially when comparing disparate investment choices.
If the MAR is too low, it will not adequately capture the risks that concern the investor, and if the MAR is too high, it will unfavorably portray what may otherwise be a sound investment.
When comparing multiple investments, some papers recommend using the risk free rate as the MAR.
Practitioners may wish to choose one MAR for consistency, several standardized MAR values for reporting a range of scenarios, or a MAR customized to the objective of the investor.

$$UPR=\frac{\sum^{n}_{t=1}{(R_{t} - MAR)}}{\delta_{MAR}}$$
where $\delta_{MAR}$ is the Downside Deviation.

The numerator in Upside Potential Ratio only uses returns that exceed the MAR, and the denominator (in Downside Deviation) only uses returns that fall short of the MAR by default.
Sortino contends that this is a more accurate and balanced protrayal of return potential, wherase Sortino Ratio can reward managers most at the peak of a cycle, without adequately penalizing them for past mediocre performance.
Others have used the full series, and this is provided as an option by the suffix of the property.

## Upside Measures

Upside measures focus on returns above a **Minimum Acceptable Return (MAR)**. Unlike traditional volatility, which treats positive and negative deviations equally, upside measures quantify how frequently and by how much a strategy exceeds its target return.

The MAR is specified when constructing the `Ratios` object through the `annual_target_return` parameter.

### `upside_frequency`

The proportion of observed returns above MAR.

This statistic measures **how often** the strategy exceeds the target return.

* Range: **0** to **1**
* `0` means no observations are above MAR.
* `1` means every observation is above MAR.

### `upside_potential`

The average upside above MAR.

Upside potential measures the average magnitude of returns above the target without squaring them. Unlike `upside_potential_subset`, it is normalized by the **total number of observations**, so observations at or below MAR contribute zero to the average.

### `upside_potential_subset`

An alternative definition of upside potential.

This statistic differs from `upside_potential` only in the normalization: the sum of positive excess returns is divided by the **number of returns above MAR** instead of the total number of observations.

Because only upside observations contribute to the denominator, `upside_potential_subset` is generally greater than or equal to `upside_potential`.

### `upside_variance`

The second-order upper partial moment ($UPM_2$) about MAR.

Only returns above MAR contribute to the calculation, but the sum of squared excess returns is normalized by the **total number of observations**. This is the upside counterpart to `downside_deviation`'s underlying lower partial moment.

### `upside_variance_subset`

An alternative definition of upside variance.

This statistic differs from `upside_variance` only in the normalization: the sum of squared excess returns is divided by the **number of returns above MAR** instead of the total number of observations.

Because only upside observations contribute to the denominator, `upside_variance_subset` is generally greater than or equal to `upside_variance`.

### `upside_risk`

The square root of `upside_variance`.

It measures the magnitude of upside fluctuations relative to MAR while retaining the normalization by the **total number of observations**. In other words, it is the upside counterpart to `downside_deviation`.

### `upside_risk_subset`

The square root of `upside_variance_subset`.

It measures upside fluctuations using only observations above MAR for normalization. Consequently, it is generally greater than or equal to `upside_risk`.

### `upside_potential_ratio`

The ratio of upside potential to downside risk:
$$\text{Upside Potential Ratio}=\frac{UPM_1}{\sqrt{LPM_2}}$$

where $UPM_1$ is the first-order upper partial moment about MAR and $LPM_2$ is the second-order lower partial moment about MAR.

The ratio combines **upside potential** with **downside risk**, rewarding strategies that generate larger returns above MAR while penalizing larger shortfalls below MAR.

### `upside_potential_ratio_subset`

An alternative version of the upside potential ratio using subset normalization for both components.

The numerator uses the average excess return among observations above MAR, while the denominator uses the square root of the average squared shortfall among observations below MAR:

$$\text{Upside Potential Ratio}_{\text{subset}}=\frac{UPM_{1,\text{subset}}}{\sqrt{LPM_{2,\text{subset}}}}$$

This version therefore compares the **average magnitude of upside observations** with the **average magnitude of downside observations**, rather than diluting either measure with observations that do not contribute to the corresponding partial moment.

| Property                       | Measures                              | Denominator / Normalization       |
| ------------------------------ | ------------------------------------- | --------------------------------- |
| `upside_frequency`             | Frequency of returns above MAR        | Total observations                |
| `upside_potential`             | Mean upside above MAR                 | Total observations                |
| `upside_potential_subset`      | Mean upside above MAR                 | Upside observations only          |
| `upside_variance`              | Squared upside above MAR              | Total observations                |
| `upside_variance_subset`       | Squared upside above MAR              | Upside observations only          |
| `upside_risk`                  | RMS upside above MAR                  | Total observations                |
| `upside_risk_subset`           | RMS upside above MAR                  | Upside observations only         |
| `upside_potential_ratio`       | Upside potential relative to downside | Total observations for both LPMs |
| `upside_potential_ratio_subset`| Upside potential relative to downside | Upside/downside observations only |

## Risk-to-VaR / ES Ratios

Risk-to-VaR and Risk-to-ES ratios measure the return earned relative to a tail-risk estimate. They provide a simple way to compare the strategy's **mean excess return** with the magnitude of its potential losses.

The numerator is the mean return in excess of the risk-free rate. The denominator is either **Value at Risk (VaR)** or **Expected Shortfall (ES)** at the specified confidence level.

A higher ratio indicates more excess return relative to the estimated tail risk.

### `reward_to_var_ratio_historical`

The mean excess return divided by **historical VaR**.

Historical VaR is estimated directly from the observed return distribution without assuming a particular parametric distribution. Consequently, the ratio reflects the empirical tail behavior of the observed returns.

### `reward_to_var_ratio_gaussian`

The mean excess return divided by **Gaussian VaR**.

Gaussian VaR assumes that returns follow a normal distribution. It therefore provides a parametric risk estimate based primarily on the mean and standard deviation of returns.

This measure is simple and computationally efficient, but it may underestimate tail risk when returns are skewed or have heavier tails than the normal distribution.

### `reward_to_var_ratio_cornish_fisher`

The mean excess return divided by **Cornish-Fisher VaR**.

Cornish-Fisher VaR extends the Gaussian approach by adjusting the estimated quantile for **sample skewness and excess kurtosis**. This allows the ratio to account for some departures from normality while retaining a parametric approach.

### `reward_to_es_ratio_historical`

The mean excess return divided by **historical Expected Shortfall (ES)**.

Historical ES measures the average loss in the tail beyond the VaR threshold using the observed return distribution. Because it considers the magnitude of losses throughout the tail rather than only a single quantile, it generally provides a more comprehensive measure of tail risk than VaR.

### `reward_to_es_ratio_gaussian`

The mean excess return divided by **Gaussian Expected Shortfall (ES)**.

Gaussian ES estimates the average tail loss under a normal-return assumption. Like Gaussian VaR, it is convenient and analytically tractable but may not adequately capture skewness or heavy tails.

### `reward_to_es_ratio_cornish_fisher`

The mean excess return divided by **Cornish-Fisher Expected Shortfall (ES)**.

Cornish-Fisher ES adjusts the Gaussian tail estimate using sample skewness and excess kurtosis. It provides a parametric estimate that attempts to better reflect asymmetric and heavy-tailed return distributions.

### VaR vs. ES

**VaR** describes a loss threshold at a given confidence level: it estimates the loss that will not be exceeded with a specified probability.

**ES** goes further by estimating the average loss when that threshold is exceeded. As a result, ES takes the severity of the worst outcomes into account rather than considering only the boundary of the tail.

For this reason, reward-to-ES ratios can provide a more informative assessment of compensation for extreme downside risk, while reward-to-VaR ratios provide a simpler measure based on a tail-loss threshold.

### Historical vs. Gaussian vs. Cornish-Fisher

The three variants differ only in how the VaR or ES denominator is estimated:

| Method | Risk estimate | Distribution assumption |
| ------------------------ | -------------------------- | ------------------------- |
| `historical` | Empirical VaR / ES | None |
| `gaussian` | Parametric VaR / ES | Normal distribution |
| `cornish_fisher` | Adjusted parametric VaR / ES | Accounts for skewness and excess kurtosis |

Together, these ratios allow the same reward-to-risk concept to be evaluated under different assumptions about the return distribution.

| Property | Numerator | Denominator |
| ------------------------------ | --------------------- | ---------------------- |
| `reward_to_var_ratio_historical` | Mean excess return | Historical VaR |
| `reward_to_var_ratio_gaussian` | Mean excess return | Gaussian VaR |
| `reward_to_var_ratio_cornish_fisher` | Mean excess return | Cornish-Fisher VaR |
| `reward_to_es_ratio_historical` | Mean excess return | Historical ES |
| `reward_to_es_ratio_gaussian` | Mean excess return | Gaussian ES |
| `reward_to_es_ratio_cornish_fisher` | Mean excess return | Cornish-Fisher ES |

## Mean Absolute Deviation (MAD)

Mean Absolute Deviation (MAD) measures the average absolute distance of returns from their mean.
Unlike standard deviation, MAD does not square deviations, making it less sensitive to extreme observations.

For returns $r_1, \ldots, r_n$ with mean $\bar{r}$, MAD is defined as:
$$MAD = \frac{1}{n}\sum_{i=1}^{n}|r_i - \bar{r}|$$

The `mean_absolute_deviation_ratio` compares the mean return with this measure of dispersion:
$$\text{MAD Ratio} = \frac{\bar{r}}{MAD}$$

A higher ratio indicates greater mean return relative to the typical absolute deviation of returns.
Unlike Sharpe-style ratios, the MAD Ratio uses absolute deviations rather than squared deviations, providing a simple and more robust measure of return relative to variability.

## Sharpe Ratio

The Sharpe ratio is simply the return per unit of risk (represented by variability).
In the classic case, the unit of risk is the standard deviation of the returns.

$$\frac{\overline{R-R_f}}{\sqrt{\sigma_{(R-R_f)}}}$$

William Sharpe now recommends Information Ratio preferentially to the original Sharpe Ratio.

The higher the Sharpe ratio, the better the combined performance of "risk" and return.

As noted, the traditional Sharpe Ratio is a risk-adjusted measure of return that uses standard deviation to represent risk.

The Sharpe Ratio can be used to measure both 'excess return' (over a risk-free rate) and 'differential return' (excess return over a benchmark).

A number of papers now recommend using a "modified Sharpe" ratio using a Modified Cornish-Fisher VaR or CVaR/Expected Shortfall as the measure of Risk.

We have extended this concept to create multivariate modified Sharpe-like Ratios for standard deviation, Gaussian VaR, modified VaR, Gaussian Expected Shortfall, and modified Expected Shortfall.

Most recently, we have added Downside Sharpe Ratio (DSR), a short name for what Ziemba (2005) called the "Symmetric Downside Risk Sharpe Ratio" and is defined as the ratio of the mean return to the square root of lower semivariance:
$$\frac{\overline{R-R_f}}{\sqrt{2}SemiSD(R)}$$.

This function returns a traditional or modified Sharpe ratio for the same periodicity of the data being input (e.g., monthly data -> monthly SR).

References
Sharpe, W.F. The Sharpe Ratio,\emph{Journal of Portfolio Management},Fall 1994, 49-58.
Laurent Favre and Jose-Antonio Galeano. Mean-Modified Value-at-Risk Optimization with Hedge Funds. Journal of Alternative Investment, Fall 2002, v 5.
Ziemba, W. T. (2005). The symmetric downside-risk Sharpe ratio. The Journal of Portfolio Management, 32(1), 108-122.
Jacquier, E., Kane, A., Marcus, A. (2003). Geometric Mean or Arithmetic Mean: A Reconsideration. Financial Analysts Journal, November/December 2003, p. 46-53.

## Adjusted Sharpe ratio

Adjusted Sharpe ratio was introduced by Pezier and White (2006) to adjusts for skewness and kurtosis by incorporating a penalty factor for negative skewness and excess kurtosis.
$$Adjusted Sharpe Ratio = SR \left[1 + \frac{S}{6} SR - \frac{K - 3}{24} SR^2\right]$$
where $SR$ is the sharpe ratio with data annualized, $S$ is the skewness and $K$ is the kurtosis

References
Carl Bacon, **Practical portfolio performance measurement and attribution**, second edition 2008 p.99
Pezier, Jaques and White, Anthony. 2006. The Relative Merits of Investable  Hedge Fund Indices and of Funds of Hedge Funds in Optimal Passive Portfolios.
{https://econpapers.repec.org/paper/rdgicmadp/icma-dp2006-10.htm}

## Downside Sharpe Ratio (DSR)

The Downside Sharpe Ratio (DSR) is a short name for what Ziemba (2005) called the "Symmetric Downside Risk Sharpe Ratio" and is defined as the ratio of the mean excess return to the square root of lower semivariance:
$$DSR=\frac{\overline{(R-R_{f})}}{\sqrt{2}SemiSD(R)}$$

Ziemba, W. T. (2005). The symmetric downside-risk Sharpe ratio. The Journal of Portfolio Management, 32(1), 108-122.

## Can I apply adjustment which uses skewness and kurtosis to var_* and es_* Sharpe ratios?

### The six VaR/ES Sharpe ratios

Having

- sharpe_ratio_var_historical()
- sharpe_ratio_var_gaussian()
- sharpe_ratio_var_cornish_fisher()
- sharpe_ratio_es_historical()
- sharpe_ratio_es_gaussian()
- sharpe_ratio_es_cornish_fisher()

is perfectly reasonable.

These correspond directly to measures implemented in PerformanceAnalytics and the academic literature:
$$\frac{E\left[R-R_j\right]}{VaR}$$ or
$$\frac{E\left[R-R_j\right]}{ES}$$
where only the denominator changes. Your users can easily understand them.

### Adjusted Sharpe is something different

The adjusted Sharpe ratio (Pezier & White) is
$$SR_{adj}= SR \left[1 + \frac{S}{6} SR - \frac{K - 3}{24} SR^2\right]$$
(or the equivalent formula using excess kurtosis).

Notice that the correction is derived specifically for the standard Sharpe ratio.

It says:
> "Sharpe assumes normality. If returns are skewed/kurtotic, here's a correction."

It does not say
> "Take any reward/risk ratio and multiply it by this factor."

### A VaR ratio already accounts for non-normality

Consider
$$\frac{\mu-R_j}{\text{Modified VaR}}$$
where Modified VaR is Cornish-Fisher.
The denominator already incorporates skewness and kurtosis through the Cornish-Fisher expansion.
Applying the adjusted Sharpe correction afterward would effectively use skewness and kurtosis twice.

### ES ratios even more so

ES is already a coherent tail-risk measure.
Historical ES uses the empirical distribution.
Modified ES explicitly adjusts using skewness and kurtosis.
Adding an adjusted Sharpe multiplier has no theoretical justification.

You'll find many papers on Modified Sharpe Ratio (VaR denominator),
Conditional Sharpe Ratio (ES denominator),
STARR Ratio (Stable Tail Adjusted Return Ratio).
but essentially none defining Adjusted Modified Sharpe Ratio or Adjusted ES Sharpe Ratio because they're solving different problems.

## Probabilistic Sharpe Ratio (PSR)

Yes, the Probabilistic Sharpe Ratio (PSR) is a different concept. It belongs in the Sharpe ratio family, but not in the same subgroup as your sharpe_ratio_adjusted, sharpe_ratio_var_*, or sharpe_ratio_es_* methods.

Here's how I would classify them.

1. Plain Sharpe ratio

The classic
$$SR=\frac{\mu - R_f}{\sigma}$$

2. Alternative risk denominator

These simply replace standard deviation by another risk measure.
- sharpe_ratio_downside
- sharpe_ratio_var_historical
- sharpe_ratio_var_gaussian
- sharpe_ratio_var_cornish_fisher
- sharpe_ratio_es_historical
- sharpe_ratio_es_gaussian
- sharpe_ratio_es_cornish_fisher

Conceptually,
$$\frac{text{excess return}}}{text{alternative risk}}$$
These are all siblings.

3. Adjusted Sharpe ratio

This is not replacing the denominator.
Instead, it starts from the ordinary Sharpe ratio and then adjusts it for skewness and kurtosis.
Typical formula (Pezier & White)
$$SR_{adj}= SR \left[1 + \frac{S}{6} SR - \frac{K - 3}{24} SR^2\right]$$
This is a different idea.

4. Probabilistic Sharpe Ratio (PSR)

This is different again. It asks
> Given my observed Sharpe ratio and sample size, what is the probability that the true Sharpe ratio exceeds some benchmark?

For example

observed SR = 0.9
benchmark = 0
sample size = 150

PSR might return 0.97 meaning
> There is a 97% probability that the true Sharpe ratio is greater than zero.

So the output is a probability, not a ratio.
$$PSR = \Phi\left(\frac{(SR-SR^*)\sqrt{n-1}}{\sqrt{1-\gamma_3SR+\frac{\gamma_4-1}{4}SR^2}}\right)$$
where
$SR^*$ is the benchmark Sharpe ratio,
$\gamma_3$  is skewness,
$\gamma_4$ is kurtosis,
$n$ is sample size.

Notice that skewness and kurtosis appear again, but this is statistical inference rather than a performance adjustment.

5. Deflated Sharpe Ratio (DSR)

This is the next step. Suppose you tested 500 trading strategies.
One strategy has SR = 2.1
Is that impressive?
Maybe not�it could simply be the best out of 500 random attempts.

The Deflated Sharpe Ratio adjusts PSR for
- multiple testing
- selection bias
- non-normal returns.

It is becoming increasingly common in quantitative finance.

## Calculating sharpe

The Sharpe ratio is defined as
$$Sharpe=\frac{E\left[R-R_f\right]}{\text{Risk}(R-R_f)$$
The important point is that both the numerator and the denominator are computed from the same excess-return series.

For the classical Sharpe ratio: mean(excess returns) / std(excess returns)
For the modified VaR Sharpe ratio: mean(excess returns) / VaR(excess returns)
For the ES variant: mean(excess returns) / ES(excess returns)

## Why would one exclude skewness or kurtosis from the probabilistic Sharpe ratio?

That's an excellent question. I had the same reaction the first time I saw the `PerformanceAnalytics` interface. The answer is that "more moments" does **not** necessarily mean "more accurate."
There are several reasons.

1. Skewness and kurtosis are estimated with much higher error

The Sharpe ratio only needs estimates of

* mean
* variance

These converge relatively quickly. Skewness converges much more slowly. Kurtosis converges even more slowly. For example, with 60 monthly returns:

* mean is reasonably estimated,
* variance is reasonably estimated,
* skewness is fairly noisy,
* kurtosis is often almost meaningless.

If the true return distribution is nearly normal, adding a very noisy kurtosis estimate can actually make the PSR estimate *worse*.

2. Financial kurtosis is dominated by a few observations

Suppose your return series is

```text
0.5%
0.3%
0.4%
...
0.2%
-12%
```

That single crash can double or triple the sample kurtosis. Now the PSR changes dramatically because of one observation.
Some practitioners would rather ignore kurtosis than let one outlier dominate the result.

3. The formulas are asymptotic approximations

The Bailey & L�pez de Prado PSR formula is derived from an approximation to the sampling distribution of the Sharpe ratio.
The expansion is

* first order -> variance only
* second order -> skewness
* third order -> kurtosis

Higher-order approximations are not always better for finite samples. This is common in statistics.

4. Some users want compatibility

PerformanceAnalytics is over 15 years old. Existing reports may have been produced with

```r
ignore.kurtosis = TRUE
```

Changing the default today would change published numbers. Backward compatibility is valuable.

5. Many return distributions are close enough to normal

Suppose

```text
skew = 0.05
kurtosis = 0.12
```

The corrections are tiny. Ignoring them makes almost no practical difference.

### Why is the default `ignore_kurtosis=TRUE`?

That default puzzled me too. Looking at the PSR formula,

$$[\operatorname{Var}(\widehat{SR})\propto 1-\gamma_3 SR+\frac{\gamma_4-1}{4}SR^2]$$

the kurtosis term is multiplied by (SR^2). Unless

* Sharpe is large, **and**
* kurtosis is estimated well,

its contribution is usually small relative to its estimation error. Many authors therefore include skewness but omit kurtosis. That tends to produce a more stable estimate.

### Why would you ignore skewness too?

Mostly for comparison. If you set

```text
ignore_skewness=True
ignore_kurtosis=True
```

you recover the normal-theory PSR. That lets you quantify how much non-normality matters.

### If this were my library...

I would probably do exactly what you're doing:

```python
probabilistic_sharpe_ratio(
    reference_sr=0,
    *,
    ignore_skewness=False,
    ignore_kurtosis=True,
)
```

to match PerformanceAnalytics.

But in the documentation I'd add a note such as:

> By default, skewness is incorporated while kurtosis is ignored, matching the default behavior of the PerformanceAnalytics R package. Including higher-order moments can improve the approximation when they are estimated reliably, but sample kurtosis is often much noisier than the mean, variance, or skewness, especially for small samples.

That gives users the rationale without forcing them to know the statistical literature.

One thing I'd also be curious about is **which PSR formula PerformanceAnalytics implements**. If it's the Bailey & L�pez de Prado (2012) formula, I'd expect the code to expose these switches because the original derivation naturally separates the skewness and kurtosis terms, making it easy to include or exclude each independently.

## Sharpe ratio family essay

Below is a compact guide to a family of Sharpe-ratio-style performance measures that differ mainly in the way �risk� is defined or how estimation issues (non-normality, autocorrelation, selection bias) are handled. I first lay out notation and the canonical formulas, then comment on their use-cases and trade-offs. Where appropriate, I give the most common versions you can copy as LaTeX. Two small SVG sketches are included at the end.

1) Notation and basic objects

Let $r_t$ be simple returns observed at times $t = 1,�,n$ and let $r_f$ be the (per-period) risk-free rate or a minimum acceptable return (MAR). Define excess returns $x_t = r_t - r_f$, with sample mean and (unbiased) standard deviation
$$
\bar{x} = \frac{1}{n}\sum_{t=1}^n x_t, \qquad s = \sqrt{\frac{1}{n-1}\sum_{t=1}^n (x_t - \bar{x})^2}.
$$
When dealing with tail-based measures, it is convenient to work with losses$ L_t = -x_t$ (so that large positive $L_t$ are bad). Let $\alpha$ denote the confidence level (e.g., $\alpha$ = 0.95), $z_{\alpha} = \Phi^{-1}(\alpha)$ the $\alpha$-quantile of the standard normal, $\Phi$ the standard normal pdf, $\Phi$ the cdf, $S$ the sample skewness of $x_t$, and $K$ the sample kurtosis (conventionally, $K=3$ for a normal distribution).

Annualization note: If your data are at frequency $m$ per year (e.g., $m$=252 for daily, 12 for monthly), a common practice is $SR_{ann} \approx \sqrt{m} SR_{periodic}$, recognizing that this relies on iid returns assumptions. Tail-based measures are not homogeneous in the same way; annualization should be handled with care.

2) Standard Sharpe ratio

Formula:
$$
\mathrm{SR} \;=\; \frac{\bar{x}}{s}.
$$

Interpretation: Excess return per unit of total volatility. Assumes volatility is an adequate proxy for risk and that returns are approximately symmetric (or that you accept equal penalization of upside and downside deviations).

3) Downside Sharpe ratio

Replace total volatility with downside deviation around the MAR (often $r_f$):
$$d_{\text{MAR}} \;=\; \sqrt{\frac{1}{n}\sum_{t=1}^n \bigl(\min(x_t,0)\bigr)^2},\qquad \mathrm{DSR} \;=\; \frac{\bar{x}}{d_{\text{MAR}}}$$

Interpretation: Penalizes only negative deviations below the threshold, which can be more aligned with investor preferences if upside variability is not considered risk.

4) Adjusted Sharpe ratios for non-normality

These adjust $SR$ for skewness $S$ and kurtosis $K$ of returns (one widely used form derived from moment expansions is):
$$\mathrm{ASR}\;=\; \mathrm{SR}\left(1 + \frac{S}{6}\,\mathrm{SR} - \frac{K-3}{24}\,\mathrm{SR}^2\right)$$
and a skew-only version would keep only the first correction term:
$$\mathrm{Skew\text{-}only\ ASR}\;=\; \mathrm{SR}\left(1 + \frac{S}{6}\,\mathrm{SR}\right)$$

Interpretation: Positive skewness raises the adjusted score, while excess kurtosis (fat tails) reduces it, reflecting the fact that $SR$ overstates �quality� when tails are heavy and understates it for positively skewed payoffs.

5) Smart Sharpe ratio (autocorrelation-adjusted)

Returns from some strategies (e.g., illiquid, smoothed, overlapping-window) violate the $iid$ assumption, inflating the naive $SR$. A common fix replaces $n$ by an effective number of independent observations $n_eff$ that accounts for autocorrelation via the sum of (potentially weighted) autocorrelations $\rho_k$:
$$n_{\text{eff}} \approx \frac{n}{1 + 2\sum_{k=1}^{K} w_k \,\rho_k}, \quad \text{with } 0 \le w_k \le 1 \text{ (e.g., } w_k = 1 - k/n\text{)}$$
Then define
$$\mathrm{SSR} \;=\; \mathrm{SR} \,\sqrt{\frac{n_{\text{eff}}}{n}}$$

Interpretation: Penalizes $SR$ for return smoothing or serial dependence, yielding a more realistic signal-to-noise ratio.

6) VaR-based Sharpe ratios

General idea: Divide mean excess return by Value-at-Risk (VaR) at level $\alpha$ for losses $L_t = -x_t$:
$$\mathrm{SR}_{\mathrm{VaR}(\alpha)} \;=\; \frac{\bar{x}}{\mathrm{VaR}_\alpha(L)}$$

- Historical VaR: Empirical quantile of $L_t$ at level $\alpha$,
$$\mathrm{VaR}_\alpha^{\text{hist}} = \text{Quantile}_\alpha(L_1,\dots,L_n)$$

- Gaussian (parametric) VaR: If L \approx N(\mu_L,\sigma_L^2), then
$$\mathrm{VaR}_\alpha^{\mathcal{N}} = \mu_L + z_\alpha\,\sigma_L$$
Often one centers $L$ (i.e., uses $\mu_L=0$) and sets $\sigma_L$ equal to the sample standard deviation of $L$.

- Cornish-Fisher VaR: Adjust $z_\alpha$ for skewness/kurtosis using the Cornish-Fisher expansion. If the standardization is based on $L$s moments (skewness $S_L$ and kurtosis $K_L$),
$$z_\alpha^{\text{CF}} \approx z_\alpha + \frac{S_L}{6}(z_\alpha^2-1) + \frac{K_L - 3}{24}(z_\alpha^3-3z_\alpha) - \frac{S_L^2}{36}(2z_\alpha^3-5z_\alpha)$$
and then
$$\mathrm{VaR}_\alpha^{\text{CF}} \approx \mu_L + \sigma_L\,z_\alpha^{\text{CF}}$$

Interpretation: VaR-based $SR$s use downside quantiles as the unit of risk. Historical VaR is assumption-free but sample-dependent. Gaussian VaR is simple but fragile under non-normality. Cornish-Fisher improves the Gaussian approximation by incorporating estimated skewness and kurtosis.

7) ES-based (CVaR) Sharpe ratios

Replace VaR with Expected Shortfall (ES, a coherent risk measure):
$$\text{SR}_{\text{ES}(\alpha)} \;=\; \frac{\bar{x}}{\text{ES}_\alpha(L)},\quad\text{ES}_\alpha(L) = \mathbb{E}[\,L \mid L \ge \mathrm{VaR}_\alpha(L)\,]$$

- Historical ES:
$$\mathrm{ES}_\alpha^{\text{hist}} = \frac{1}{\#\{t: L_t \ge \mathrm{VaR}_\alpha^{\text{hist}}\}}\sum_{t: L_t \ge \mathrm{VaR}_\alpha^{\text{hist}}} L_t$$

- Gaussian ES (for $L \approx N(\mu_L,\sigma_L^2)$):
$$\mathrm{ES}_\alpha^{\mathcal{N}} = \mu_L \;+\; \sigma_L\,\frac{\phi(z_\alpha)}{1-\alpha}$$

- Cornish-Fisher ES: Combine a Cornish-Fisher VaR with a tail-expectation approximation; practical implementations often use a CF-adjusted quantile to locate the tail, then average the tail beyond that threshold (semi-parametric), or use closed-form CF-type expansions for ES (less common and more delicate).

Interpretation: ES penalizes the entire tail, not just a quantile, and is therefore more sensitive to extreme losses than VaR. ES-based SRs are often preferred in risk management contexts.

8) Probabilistic Sharpe ratio (PSR)

Given a sample Sharpe ratio $\widehat{\mathrm{SR}}$ and a reference threshold $\mathrm{SR}^\ast$, PSR is the probability that the true Sharpe ratio exceeds $\mathrm{SR}^\ast$, under assumptions about the sampling distribution of $\widehat{\mathrm{SR}}$. A widely used approximation (Bailey & Lopez de Prado) employs skewness and kurtosis corrections (Opdyke) and yields:
$$\mathrm{PSR}(\mathrm{SR}^\ast) \;=\; \Phi\!\left(\frac{(\widehat{\mathrm{SR}} - \mathrm{SR}^\ast)\,\sqrt{n-1}}{\sqrt{\,1 - S\,\widehat{\mathrm{SR}} + \frac{K-1}{4}\,\widehat{\mathrm{SR}}^2\,}}\right)$$
- If ignore_scewness = TRUE, set $S$ = 0 in the denominator.
- If ignore_kurtosis = TRUE, set $K$ = 3 (normal kurtosis) in the denominator.

Interpretation: PSR answers
>how confident should I be that the true risk-adjusted performance exceeds a benchmark?
This is especially valuable for multiple trials and non-normal returns, where naive SRs tend to be over-optimistic.

9) Practical remarks and trade-offs

- Symmetry vs. asymmetry: sharpe_ratio treats upside and downside equally; downside_sharpe_ratio, VaR- and ES-based versions emphasize downside.
- Distributional assumptions: Gaussian versions are simple but can mislead under skewness/fat tails; Cornish-Fisher partially corrects this; historical is robust but sample heavy.
- Non-iid data: smart_sharpe_ratio penalizes for autocorrelation/smoothing a common issue in illiquid assets and overlapping signals.
- Statistical significance: probabilistic_sharpe_ratio formally quantifies confidence in a reported SR relative to a hurdle.
- Target setting: In downside measures, pick MAR carefully (risk-free, inflation, investor mandate, etc.). Choice of $\alpha$ in VaR/ES controls tail sensitivity.

10) Two small SVG sketches

A) Risk measure denominator families

Copyable inline SVG (simple schematic):

<svg width="560" height="150" xmlns="http://www.w3.org/2000/svg">
  <rect x="0" y="0" width="560" height="150" fill="#ffffff" stroke="none"/>
  <text x="15" y="22" font-family="Arial" font-size="14" fill="#222">Same numerator: mean excess return, \sum[x] \approx \bar{x}</text>
  <line x1="15" y1="28" x2="545" y2="28" stroke="#bbb" stroke-width="1"/>
  <g transform="translate(15,45)">
    <rect x="0" y="0" width="160" height="90" fill="#f7fbff" stroke="#5aa0d3"/>
    <text x="10" y="20" font-family="Arial" font-size="12" fill="#0a4a78">Volatility-based</text>
    <text x="10" y="40" font-family="Arial" font-size="12" fill="#333">SR = \bar{x}/s</text>
    <text x="10" y="58" font-family="Arial" font-size="12" fill="#333">DSR = \bar{x}/d</text>
    <text x="10" y="76" font-family="Arial" font-size="12" fill="#333">Adjusted SR (S,K)</text>
  </g>
  <g transform="translate(195,45)">
    <rect x="0" y="0" width="160" height="90" fill="#fff8f2" stroke="#e08b3c"/>
    <text x="10" y="20" font-family="Arial" font-size="12" fill="#7a3f00">VaR-based</text>
    <text x="10" y="40" font-family="Arial" font-size="12" fill="#333">\bar{x}/VaR</text>
    <text x="10" y="58" font-family="Arial" font-size="12" fill="#333">hist / Gaussian</text>
    <text x="10" y="76" font-family="Arial" font-size="12" fill="#333">Cornish�Fisher</text>
  </g>
  <g transform="translate(375,45)">
    <rect x="0" y="0" width="160" height="90" fill="#f6fff6" stroke="#5aa05a"/>
    <text x="10" y="20" font-family="Arial" font-size="12" fill="#2d5a2d">ES-based</text>
    <text x="10" y="40" font-family="Arial" font-size="12" fill="#333">\bar{x}/ES</text>
    <text x="10" y="58" font-family="Arial" font-size="12" fill="#333">hist / Gaussian</text>
    <text x="10" y="76" font-family="Arial" font-size="12" fill="#333">Cornish�Fisher</text>
  </g>
</svg>

B) Probabilistic Sharpe Ratio

<svg width="560" height="180" xmlns="http://www.w3.org/2000/svg">
  <rect x="0" y="0" width="560" height="180" fill="#ffffff" stroke="none"/>
  <text x="15" y="20" font-family="Arial" font-size="14" fill="#222">Probabilistic Sharpe Ratio: P(true SR &gt; SR*)</text>
  <line x1="15" y1="26" x2="545" y2="26" stroke="#bbb" stroke-width="1"/>
  <g transform="translate(30,40)">
    <path d="M 0 80 C 60 -20, 180 -20, 240 80" fill="none" stroke="#5aa0d3" stroke-width="2"/>
    <path d="M 240 80 C 300 180, 420 180, 480 80" fill="none" stroke="#5aa0d3" stroke-width="2"/>
    <line x1="260" y1="0" x2="260" y2="160" stroke="#e08b3c" stroke-width="1" stroke-dasharray="4,4"/>
    <text x="250" y="15" font-family="Arial" font-size="12" fill="#e08b3c">SR*</text>
    <line x1="310" y1="0" x2="310" y2="160" stroke="#2d5a2d" stroke-width="1" stroke-dasharray="4,4"/>
    <text x="300" y="15" font-family="Arial" font-size="12" fill="#2d5a2d">SR^</text>
    <rect x="260" y="80" width="220" height="80" fill="rgba(90,160,83,0.15)" stroke="none"/>
    <text x="265" y="150" font-family="Arial" font-size="12" fill="#2d5a2d">Area = PSR</text>
  </g>
</svg>

Summary

- sharpe_ratio: $SR = \bar{x}/s$.
- downside_sharpe_ratio: $DSR = \bar{x}/d$ with $d$ the downside deviation.
- adjusted_sharpe_ratio: incorporates $S$ and $K$ to correct $SR$ for non-normality; one common form is $SR(1 + S SR/6 - (K-3) SR^2/24)$.
- skew_only_adjusted_sharpe_ratio: keeps the skewness term only.
- smart_sharpe_ratio: adjusts $SR$ by $\sqrt{n_eff/n}$ to downgrade smoothed/serially correlated returns.
- sharpe_ratio_var_*: replace the denominator by $VaR_\alpha$ using historical, Gaussian, or Cornish-Fisher estimates.
- sharpe_ratio_es_*: replace the denominator by $ES_]alpha$ using historical, Gaussian, or Cornish-Fisher estimates.
- probabilistic_sharpe_ratio: $PSR = \Phi( ((\widehat{SR} - SR*)\sqrt{n-1}) / \sqrt{1 - \widehat{SR} + ((K-1)/4) \widehat{SR}^2} ), with options to ignore $S$ or $K$.

Choosing among these variants depends on your beliefs about return distributions, your notion of "risk", sample size, and the degree of autocorrelation or selection bias in your backtests. As a rule of thumb: prefer ES over VaR if you care about tail severity, correct for skewness/kurtosis when non-normality is evident, and always check PSR (and related tools) when you want to assess whether an observed Sharpe ratio is statistically credible rather than illusory.

## Schwager-Sortino ratio

Jack Schwager's adjusted Sortino ratio (Sortino/$\sqrt2$), proposed by Jack Schwager for direct comparison with the Sharpe ratio.
This normalization makes the metric directly comparable to the Sharpe ratio under symmetric return distributions.

In his [Market Wizards Search Part 2: The Performance Statistics I Use](https://archive.is/2rwFW#selection-599.0-603.1607) article, Jack Schwager describes
> Sortino Ratio/$\sqrt2$ - The Sortino Ratio is a variation of the Sharpe ratio that uses only downside deviation to measure risk instead of the standard deviation, which is based on all returns. The reason for dividing by $\sqrt2$ is explained in the note below. I much prefer the Sortino ratio to the Sharpe ratio because it only penalizes downside volatility, whereas the risk measure of the Sharpe ratio doesn't distinguish between upside and downside volatility.

> Note: Since the Sortino ratio has the same numerator as the Sharpe ratio but calculates the denominator based on the squared deviations of only losing returns, instead of all returns, it will be biased to be higher than the Sharpe ratio, even for traders whose returns are negatively skewed (i.e., large losses are greater in absolute magnitude than large gains). A common mistake is to assume if that if the Sortino ratio is higher than the Sharpe ratio, it implies returns are positively skewed (i.e., large gains are greater in magnitude than large losses). Since the loss measure in the Sortino ratio will be based on summing a smaller number of deviations (i.e., only the deviations of losing returns), the Sortino ratio will almost invariably be higher than the Sharpe ratio. To allow for comparing the Sortino ratio to the Sharpe ratio, we multiply the risk measure of the Sortino ratio by the square root of 2 (which is the same as dividing the Sortino ratio by the square root of 2). Multiplying the risk measure of the Sortino ratio by the square root of 2 will equalize the risk measures of the Sharpe and Sortino ratios when upside and downside deviations are equal, which seems appropriate. The adjusted version of the Sortino ratio allows for direct comparisons of the Sharpe and Sortino ratios. Generally speaking, a higher adjusted Sortino ratio implies that the distribution of returns is right-skewed (a greater tendency for large gains than large losses). And, similarly, a lower adjusted Sortino ratio implies returns are left-skewed (a greater propensity for large losses than large gains).

> In searching for potential candidates to interview for a Market Wizards book, I look for the following performance thresholds: Sortino Ratio/$\sqrt2$ - 2.0 or higher

This is not an established or standard metric in academic literature.
It appears to be Schwager's own normalization that he uses in *Market Sense and Nonsense* and related writings.

What Schwager proposes here is not a new definition of the Sortino ratio. It is simply
$$\text{Adjusted Sortino} = \frac{\text{Sortino}}{\sqrt{2}}$$.

The rationale is statistical:

* The Sharpe ratio uses the standard deviation of **all** returns.
* The Sortino ratio uses only the **downside** deviations.
* If returns are perfectly symmetric (e.g. normally distributed around the target), then downside variance is approximately **half** of total variance:
  $$\sigma_D^2 \approx \frac{\sigma^2}{2}$$, hence $$\sigma_D \approx \frac{\sigma}{\sqrt2}$$.
* Therefore the unadjusted Sortino is expected to be roughly $\sqrt2$
  times the Sharpe ratio even when there is **no asymmetry** in returns.

Dividing by ($\sqrt2$) removes this built-in scaling difference, making the values directly comparable.

## Sortino Ratios

The Sortino ratio measures **return relative to downside risk**. Unlike the Sharpe ratio, which treats both positive and negative volatility as risk, the Sortino ratio considers only returns below a **Minimum Acceptable Return (MAR)**.

The MAR is specified when constructing the `Ratios` object through the `annual_target_return` parameter.

### `sortino_ratio`

The Sortino ratio is calculated as the arithmetic mean excess return over MAR divided by the square root of the second-order lower partial moment ($LPM_2$):

$$\text{Sortino Ratio}=\frac{\operatorname{mean}(R - MAR)}{\sqrt{LPM_2}}$$

Only returns below MAR contribute to the downside-risk denominator. A higher ratio indicates greater excess return relative to downside risk.

Unlike the Sharpe ratio, the Sortino ratio does **not** use the risk-free rate as its return threshold. The investor's required return, represented by MAR, defines what constitutes undesirable performance.

### `sortino_ratio_sqrt2`

The adjusted Sortino ratio divides the standard Sortino ratio by $\sqrt{2}$:

$$\text{Adjusted Sortino Ratio}=\frac{\text{Sortino Ratio}}{\sqrt{2}}$$

This normalization was proposed by Jack Schwager to make the Sortino ratio more directly comparable with the Sharpe ratio. For symmetric return distributions, downside deviation is related to standard deviation by a factor of approximately $\sqrt{2}$, motivating this adjustment.

### `sortino_satchell_ratio`

The Sortino-Satchell ratio also measures the arithmetic mean excess return over MAR relative to downside deviation. In the implementation, its formula is therefore equivalent to the standard `sortino_ratio`:

$$\text{Sortino-Satchell Ratio}=\frac{\operatorname{mean}(R - MAR)}{\sqrt{LPM_2}}$$

The distinction is primarily **conceptual and historical**: the ratio follows the Sortino-Satchell formulation and explicitly uses the arithmetic mean of excess returns.

### Comparison

| Property | Numerator | Denominator | Normalization |
| --- | --- | --- | --- |
| `sortino_ratio` | Mean excess return over MAR | Downside deviation | Standard |
| `sortino_ratio_sqrt2` | Mean excess return over MAR | Downside deviation | Divided by $\sqrt{2}$ |
| `sortino_satchell_ratio` | Mean excess return over MAR | Downside deviation | Standard |

All three ratios share the same basic principle: **reward performance above the required return while penalizing only downside volatility**. The main difference is the normalization applied to the standard Sortino measure.

## Autocorrelation Penalty

The **Sortino ratio with autocorrelation penalty** is not a standard, universally defined metric like the Sharpe or Sortino ratio.
It is analogous to the **Sharpe ratio with autocorrelation penalty**, where the denominator is inflated to account for serially correlated returns.

The idea comes from Andrew W. Lo's observation that Sharpe ratios are overstated when returns are positively autocorrelated.
The same logic can be applied to the Sortino ratio.

The usual approach is:

1. Compute the standard Sortino ratio: $$S = \frac{\bar r - r_f}{\sigma_D}$$,
   where (\sigma_D) is the downside deviation.

2. Estimate the autocorrelations $\rho_k$ of the return series.

3. Compute Lo's variance inflation factor: $$A(q)=\sqrt{1 + 2\sum_{k=1}^{q-1}\left(1-\frac{k}{q}\right)\rho_k}$$.

4. Penalize the ratio: $$S_{\text{adj}}=\frac{S}{A(q)}$$.

The numerator is unchanged. The downside deviation is unchanged. Only the **final ratio** is divided by the autocorrelation adjustment factor.

Why? Positive autocorrelation makes a strategy appear smoother than it really is. Examples include:

* illiquid assets,
* private equity,
* many hedge funds,
* option-selling strategies,
* NAV-smoothed funds.

These strategies often exhibit artificially high Sharpe and Sortino ratios because consecutive returns are not independent.

Is this standardized? Not really.
Unlike the Sharpe adjustment (which has a well-known theoretical basis in Lo's paper), there is **no widely accepted "autocorrelation-adjusted Sortino ratio."** Different libraries and papers implement it differently:

* some apply Lo's correction directly to the Sortino ratio;
* some recompute downside deviation using long-run variance estimators (much rarer);
* some simply report the standard Sortino alongside the autocorrelation-adjusted Sharpe.

## Omega Ratio

The important point is that **`Rf` in PerformanceAnalytics' `Omega()` is not the target return/MAR**.
It is a discount rate used in the option-pricing interpretation of Omega, and for the `"simple"` method it **cancels completely**.

### 1. What the R implementation actually calculates

For `method = "simple"`:
```r
numerator <- exp(-Rf) * mean(pmax(x - L, 0))
denominator <- exp(-Rf) * mean(pmax(L - x, 0))
omega <- numerator / denominator
```

Therefore,
$$\Omega(L)=\frac{e^{-Rf} E[(R-L)^+]}{e^{-Rf} E[(L-R)^+]}=\frac{E[(R-L)^+]}{E[(L-R)^+]}$$
The `exp(-Rf)` factor appears in **both** numerator and denominator, so it cancels:
$$\boxed{\Omega(L)=\frac{E[(R-L)^+]}{E[(L-R)^+]}}$$
Consequently, changing `Rf` cannot change the result of the simple Omega calculation.
So your observation that:
> results doesn't change when rf is changed

is exactly what we should expect.

### 2. `L` is the MAR-like threshold

The parameter that corresponds to your **target return / MAR** is actually `L`.
The R documentation says:
> `L` is the loss threshold that can be specified as zero, return from a benchmark index, or an absolute rate of return

So conceptually:

```text
PerformanceAnalytics     Your library
---------------------    ----------------
L                        annual_target_return / MAR
Rf                       no equivalent needed
```

For the simple Omega calculation, `Rf` has no effect on the result.
This means that your use of:

```python
self._target_partial_moments
```

is conceptually appropriate, provided that those partial moments are calculated relative to the same target/MAR.

### 3. Your formula is mathematically equivalent

Your implementation is:

```python
lpm1 = self._target_partial_moments.lower_partial_moment_1
return self._target_returns_kbn.mean / lpm1 + 1
```

Assuming:

```text
target_returns = returns - MAR
```

then:
$$\operatorname{mean}(R-MAR)=E[(R-MAR)^+] - E[(MAR-R)^+]$$,
Let
$$UPM_1 = E[(R-MAR)^+]$$
and
$$LPM_1 = E[(MAR-R)^+]$$.
Therefore:
$$E[R-MAR] = UPM_1-LPM_1$$
and hence:

$$\frac{E[R-MAR]}{LPM_1}+1=\# \frac{UPM_1-LPM_1}{LPM_1}+1\frac{UPM_1}{LPM_1}$$.
Thus your implementation:
$$\boxed{1+\frac{E[R-MAR]}{LPM_1}}$$
is exactly equivalent to:
$$\boxed{\frac{UPM_1}{LPM_1}}$$
which is the standard simple Omega formulation.
So **your implementation is correct**.

### 4. A useful way to describe your implementation

I'd document your `omega_ratio` as an Omega ratio **relative to MAR**, rather than mentioning risk-free rate at all.
The conceptual formula is:
$$\Omega(MAR)=\frac{E[(R-MAR)^+]}{E[(MAR-R)^+]}$$
or equivalently:
$$\Omega(MAR)=1+\frac{E[R-MAR]}{LPM_1(MAR)}$$.
That second formulation is exactly what your code implements.
One subtle but important point: **Omega is not inherently a risk-free-rate-based ratio.** Its defining threshold is (L), which can be zero, a benchmark return, or an investor-specified target/MAR. PerformanceAnalytics exposes `Rf` because of the option-pricing interpretation, but `Rf` does not alter the simple Omega ratio.
So I would keep your implementation target-based:

```python
lpm1 = self._target_partial_moments.lower_partial_moment_1
if math.isnan(lpm1) or lpm1 == 0:
    return math.nan
return self._target_returns_kbn.mean / lpm1 + 1
```

and treat it as **Omega relative to MAR**, not Omega relative to the risk-free rate.

### Implementing the interpolated method

There is an important distinction: **for the standard empirical Omega ratio, you probably don't need to implement PerformanceAnalytics' `"interp"` method at all.**
Your current implementation is the natural equivalent of its `"simple"` method:
$$\Omega(MAR)=\frac{\frac1n\sum_i \max(R_i-MAR,0)}{\frac1n\sum_i \max(MAR-R_i,0)}$$.
The interpolation method is somewhat more involved because it first constructs an **interpolated empirical CDF** and then numerically integrates the CDF above and below the target.

### What PerformanceAnalytics `"interp"` does

Conceptually, it calculates:
$$\Omega(L)=\frac{\int_L^b [1-F(r)],dr}{\int_a^L F(r),dr}$$.
It then:

1. Sorts the observed returns.
2. Constructs the empirical CDF.
3. Linearly interpolates the CDF between observed return values.
4. Numerically integrates:
   * (F(r)) below `L`
   * (1-F(r)) above `L`
5. Divides the two integrals.

So yes, it is **more complicated computationally** than your current implementation.

### But there is a subtle issue

For the empirical distribution, the integral formulation is closely related to the partial-moment formulation:
$$E[(R-L)^+]=\int_L^\infty [1-F(r)],dr$$
and
$$E[(L-R)^+]=\int_{-\infty}^L F(r),dr$$.
Therefore the standard empirical Omega can be calculated directly from the first-order partial moments�which is exactly what you're doing.
In other words, your:

```python
return self._target_returns_kbn.mean / lpm1 + 1
```

is not merely a convenient approximation. It is a direct calculation of the same mathematical quantity.

### Why does PerformanceAnalytics have `"interp"` then?

The interpolation method is useful when you want to construct **Omega as a smooth function of the target (L)**.

For example, PerformanceAnalytics can produce Omega for every point along the return distribution. The interpolation method essentially creates a piecewise-linear approximation of the empirical CDF and integrates it.

That is useful for things like an **Omega curve**:

```text
Omega
  ^
  |
  |\
  | \
  |  \
  |   \
  |    \
  +-------------> MAR
```

where you evaluate $\Omega(L)$ for many different values of (L).
For a single target MAR, your partial-moment implementation is much simpler and, mathematically, more direct.

## Omega Excess Return

Omega Excess Return is an **annualized downside-risk-adjusted return**. Unlike the Omega ratio, which compares upside and downside potential, Omega Excess Return expresses the result in the same units as return and incorporates both portfolio and benchmark downside risk.

The measure is calculated as:
$$\text{Omega Excess Return}=R_P - 3\sigma_{D,P}\sigma_{D,B}$$
where:

- $R_P$ is the annualized geometric return of the portfolio.
- $\sigma_{D,P}$ is the annualized portfolio downside deviation relative to the **Minimum Acceptable Return (MAR)**.
- $\sigma_{D,B}$ is the annualized benchmark downside deviation relative to the same MAR.

The downside deviations use the full number of observations for normalization, so only returns below MAR contribute to the downside-risk measure while all observations determine its normalization.

### Interpretation

The first term represents the portfolio's annualized return, while the second term is a **downside-risk penalty**. A higher Omega Excess Return therefore indicates a better combination of return and downside-risk characteristics.

Because the benchmark downside deviation is included, the measure is particularly useful for evaluating a portfolio **relative to a benchmark**. A portfolio with a high return but substantial downside risk may receive a lower Omega Excess Return than a portfolio with a similar return and more favorable downside characteristics.

Unlike `omega_ratio` and `omega_sharpe_ratio`, Omega Excess Return is **not a dimensionless ratio**. It is expressed in return units and can therefore be interpreted as a risk-adjusted annualized return.

The name "Omega" reflects its historical association with the family of downside-risk performance measures developed around the Omega framework. However, unlike the `omega_ratio`, this statistic is not calculated as a ratio of upper and lower partial moments.

The MAR is specified when constructing the `Ratios` object through the target-return configuration.

| Property | Measures | Result |
| --- | --- | --- |
| `omega_ratio` | Upside potential relative to downside potential | Dimensionless ratio |
| `omega_sharpe_ratio` | Omega ratio relative to 1 | Dimensionless ratio |
| `omega_excess_return` | Return less a benchmark-related downside-risk penalty | Return |

## Omega Ratio vs. Omega-Sharpe Ratio

The **Omega ratio** and **Omega-Sharpe ratio** are closely related downside-risk performance measures. Both use the **Minimum Acceptable Return (MAR)** as the reference point and distinguish between returns above and below that target.

### Omega Ratio

The Omega ratio compares the total upside potential with the total downside potential relative to MAR:
$$\Omega(MAR)=\frac{UPM_1(MAR)}{LPM_1(MAR)}$$

It can also be expressed as:
$$\Omega(MAR)=1+\frac{E[R-MAR]}{LPM_1(MAR)}$$

An Omega ratio of **1** indicates that upside and downside potential are equal. Values above 1 indicate that upside potential exceeds downside potential, while values below 1 indicate the opposite.

Because it considers first-order partial moments, Omega incorporates the magnitude of returns on both sides of MAR rather than reducing risk to a single measure such as standard deviation.

### Omega-Sharpe Ratio

The Omega-Sharpe ratio used here is simply a transformation of the Omega ratio:
$$\text{Omega-Sharpe Ratio}=\Omega(MAR)-1$$
or equivalently:
$$\text{Omega-Sharpe Ratio}=\frac{UPM_1(MAR)}{LPM_1(MAR)}-1$$
Thus, the two measures contain **exactly the same information**. The difference is their reference point.

An Omega ratio of 1 corresponds to an Omega-Sharpe ratio of 0. An Omega ratio of 1.5 corresponds to an Omega-Sharpe ratio of 0.5, and an Omega ratio of 0.8 corresponds to an Omega-Sharpe ratio of -0.2.

### Why have both?

The Omega ratio is intuitive as a **relative upside-to-downside measure**: values greater than 1 indicate more upside potential than downside potential.

The Omega-Sharpe ratio shifts the scale so that **zero represents the neutral point**. This makes it somewhat more natural to interpret as a Sharpe-like performance measure:

* **> 0** � upside potential exceeds downside potential.
* **= 0** � upside and downside potential are equal.
* **< 0** � downside potential exceeds upside potential.

Despite its name, the Omega-Sharpe ratio does **not** use standard deviation or second-order downside risk. It is based on the same first-order partial moments as the Omega ratio.

In practical terms, **Omega is the more established and directly interpretable measure**, while Omega-Sharpe is simply a shifted version that can be more convenient when comparing performance measures whose neutral value is zero.

## Kappa Ratios

The **Kappa ratios** are a family of downside-risk-adjusted performance measures that evaluate returns relative to a **Minimum Acceptable Return (MAR)**. Unlike traditional ratios such as the Sharpe ratio, Kappa ratios do not treat positive and negative volatility equally. Instead, they consider only returns that fall below the target return.

The general Kappa ratio of order $n$ is:
$$K_n =\frac{E[R-MAR]}{\left(LPM_n(MAR)\right)^{1/n}}$$
where $LPM_n$ is the **nth-order lower partial moment**. The numerator represents the mean excess return over MAR, while the denominator represents downside risk with a sensitivity determined by the order.

### Different orders

The order determines how strongly the measure penalizes large downside deviations.

**Kappa-1** uses the first-order lower partial moment:
$$K_1 = \frac{E[R-MAR]}{LPM_1}$$
It relates excess return to the average downside shortfall and is therefore closely associated with **downside potential**.

**Kappa-2** uses the second-order lower partial moment:
$$K_2 = \frac{E[R-MAR]}{\sqrt{LPM_2}}$$
This is particularly important because, when the same MAR and LPM definition are used, **Kappa-2 is equivalent to the Sortino ratio**.

**Kappa-3** and **Kappa-4** use third- and fourth-order lower partial moments.
These increasingly emphasize large downside deviations.
Consequently, they can be useful when severe losses are more important than ordinary downside fluctuations.

### Interpretation

Higher Kappa values indicate a more favorable combination of excess return and downside risk.
A positive value indicates that the mean return exceeds MAR, while a negative value indicates that it falls below MAR.

The Kappa family provides a useful way to examine a strategy from different perspectives.
Kappa-1 focuses on the average magnitude of shortfalls, while higher-order versions progressively increase the penalty assigned to severe losses.

In practice, **Kappa-2 is the most familiar member because of its relationship to the Sortino ratio**.
Kappa-3 and Kappa-4 can provide additional information about sensitivity to large downside events, although higher-order measures are generally less common because they become increasingly sensitive to extreme observations and sample size.

All Kappa ratios in this implementation use the **MAR (target return)** specified when constructing the `Ratios` object.

### Do we often need kappa ratios of order > 4 in practice

Not usually. **Kappa ratios above order 4 are uncommon in practical investment analysis.**
The usual hierarchy is:

| Ratio         | Downside measure | Practical use                                  |
| ------------- | ---------------- | ---------------------------------------------- |
| **Kappa-1**   | $LPM_1$             | Downside potential / average shortfall         |
| **Kappa-2**   | $LPM_2$             | **Very common; equivalent to Sortino**         |
| **Kappa-3**   | $LPM_3$             | Sometimes useful for emphasizing severe losses |
| **Kappa-4**   | $LPM_4#             | Mostly specialized / tail-risk analysis        |
| **Kappa > 4** | Higher-order LPM | Rare                                           |

The reason is that increasing the order makes the measure increasingly sensitive to **large downside observations**. For order (n),
$$K_n =\frac{E[R-MAR]}{\left(LPM_n(MAR)\right)^{1/n}}$$.

As $n$ increases, very large shortfalls receive disproportionately greater weight. For example:
$$LPM_2 \sim d^2,\qquad LPM_3 \sim d^3,\qquad LPM_4 \sim d^4$$.

Consequently, a single extreme loss can have a substantial influence on Kappa-4 and potentially dominate even more strongly for Kappa-5, Kappa-6, etc.

### Why Kappa-2 is the practical sweet spot

Kappa-2 has a particularly useful interpretation:
$$K_2 =\frac{E[R-MAR]}{\sqrt{LPM_2(MAR)}}$$

which is essentially the **Sortino ratio** when the same MAR and LPM definition are used.

Kappa-3 and Kappa-4 are useful when you specifically want to ask:

> "How does performance look when I penalize increasingly severe downside outcomes?"

But beyond order 4, interpretation becomes progressively less intuitive, and estimation becomes increasingly sensitive to sample size and extreme observations.

## Prospect ratio

In the original Prospect Theory of Kahneman and Tversky, the value function is typically represented as something like
$$v(x)=\begin{cases}x^\alpha, & x\ge0\\-\lambda(-x)^\beta, & x<0,\end{cases}$$

where $\alpha,\beta$ describe diminishing sensitivity, $\lambda>1$ represents **loss aversion**.
The important parameter here is $\lambda$.

The original experimental estimates are often summarized with a loss-aversion coefficient around $2.25$. In other words, a loss of a given magnitude is psychologically weighted roughly $2.25$ times as strongly as a gain of the same magnitude.

So, for a return of $+5\%$ and $-5\%$, the prospect-theory weighting would conceptually be $+5\%$ versus $2.25(-5\%)=-11.25\%$.

The latter is a much larger negative contribution.

This is the conceptual basis for the `2.25` in the implementation.

The **classic Prospect Theory value function is not simply**
$$r_+ + 2.25r_-.$$

The canonical value function has a **nonlinear power transformation**:
$$v(r)=\begin{cases}r^\alpha,&r\ge0,\\-\lambda(-r)^\beta,&r<0.\end{cases}$$

A commonly used parameterization is approximately
$$\alpha=\beta\approx0.88,\qquad\lambda\approx2.25.$$

Thus the classical Prospect Theory treatment of a return distribution would involve something more like
$$\frac{1}{n}\sum_i v(r_i)$$

not simply
$$\frac{1}{n}\left[\sum r_i^+ + 2.25\sum r_i^-\right]$$

The Prospect Ratio used in portfolio-performance literature is generally attributed to the implementation described by **Carl Bacon**, rather than being a direct canonical equation from Kahneman and Tversky.

The basic idea is
$$PR=\frac{\text{prospect-adjusted return}}{\text{downside risk}}$$

with the prospect-adjusted return constructed by giving losses greater weight than gains.

There's a subtle problem with calling this quantity simply a "Prospect Theory ratio."
The loss-aversion component is:
$$r_i^+ + \lambda r_i^-$$

while MAR is introduced separately.

If MAR is supposed to represent the **reference point** in Prospect Theory, the more theoretically natural construction would evaluate gains and losses **relative to MAR**
$$x_i=R_i-MAR$$

and then apply the value function
$$v(x_i)=\begin{cases}x_i^\alpha,&x_i\ge0\\-\lambda(-x_i)^\beta,&x_i<0.\end{cases}$$

That would be much closer to actual Prospect Theory.

But that is **not what `PerformanceAnalytics::ProspectRatio()` does**.

It separates the concepts:

- positive/negative returns are classified relative to **zero**;
- `MAR` is subtracted from the resulting average;
- downside deviation is calculated relative to MAR.

That's a portfolio-performance metric inspired by Prospect Theory, rather than a literal Prospect Theory utility calculation.

For `PerformanceAnalytics`, $\lambda=2.25$ is essentially a **fixed loss-aversion coefficient**.
It is a **behavioral-model assumption**
> A negative return is weighted approximately 2.25 times as heavily as an equally sized positive return.

Consequently, two strategies with identical conventional returns can have very different Prospect Ratios if their gain/loss distributions differ.

There isn't one universally accepted loss-aversion constant.

Different studies and experimental designs have produced different estimates. Values around 2–2.5 are commonly associated with the classic literature, but later research has found considerable variation.

There are really **three different things** that should not be conflated:
$$\boxed{\text{Kahneman--Tversky Prospect Theory}}$$

uses a nonlinear value function with a loss-aversion parameter $\lambda$;
$$\boxed{\text{Bacon's Prospect Ratio}}$$

turns the idea of loss aversion into a portfolio-performance ratio; and
$$\boxed{\text{PerformanceAnalytics::ProspectRatio}}$$

implements a particular Bacon-inspired formula using the fixed coefficient **2.25**.

Bacon's formula is
$$PR=\frac{\frac{1}{n}\sum_{i=1}^n\left[\max(r_i,0)+\lambda\min(r_i,0)\right]-r_T}{\sigma_D}$$

Expanding the sum
$$PR=\frac{\frac{\sum r_i^+ + \lambda\sum r_i^-}{n}-r_T}{\sigma_D}$$

Bacon's note is worth paying attention to:
> prospect ratio is similar to Kappa but with investor preferences expressed in the numerator rather than in the denominator

Kappa uses:
$$K_\alpha=\frac{E[R]-MAR}{LPM_\alpha(MAR)^{1/\alpha}}$$

so investor preference enters through **what constitutes downside risk** and its order.

Prospect Ratio instead modifies the return
$$\frac{E[r^+ + \lambda r^-]-MAR}{\sigma_D}$$

while retaining downside deviation as the denominator.
So conceptually

```text
Kappa
    investor preference
          ↓
    denominator
    LPM order

Prospect Ratio
    investor preference
          ↓
    numerator
    loss-aversion λ
```

That's a useful reason to have the Bacon version in the library.

One subtle point about `lambda_loss = 1`

With $\lambda=1$, $r_i^+ + r_i^- = r_i$ so the prospect-adjusted return becomes simply
$$\frac{1}{n}\sum_i r_i = \overline R$$

Consequently
$$PR_{\lambda=1}=\frac{\overline R-MAR}{\sigma_D}$$

That's essentially a **Sortino-type ratio using downside deviation**.

So you can think of the parameter as moving away from an ordinary downside-adjusted return measure
$$\lambda=1\quad\rightarrow\quad\text{ordinary return weighting}$$

toward stronger loss aversion
$$\lambda>1\quad\rightarrow\quad\text{increasing penalty for losses}$$

And
$$\lambda=2.25$$

is the particular empirical preference assumption used by Watanabe/Bacon.

That makes `lambda_loss` a particularly nice parameter for **strategy sensitivity analysis**. For example, rather than asking only "which strategy has the highest Prospect Ratio?", you can examine rankings at $\lambda=0,\;1,\;2.25,\;3,\;4$ and see whether the strategy remains attractive as the investor becomes increasingly loss-averse.

## Prospect Ratio in the landscape of performance measures

The Prospect Ratio occupies an interesting position in a performance-measure library because it sits between **Sortino/Kappa-style downside-risk measures** and **behavioral-performance measures**. It is fundamentally a downside-adjusted return measure, but unlike most classical ratios, it introduces an explicit assumption about how strongly an investor dislikes losses relative to gains.

The key formula in the Bacon/Watanabe formulation is
$$PR(\lambda)=\frac{\frac{1}{n}\sum_{i=1}^{n}\left[\max(r_i,0)+\lambda\min(r_i,0)\right]-MAR}{\sigma_D}$$

where $\lambda$ is the loss-aversion coefficient and $\sigma_D$ is downside deviation.

This gives the Prospect Ratio a particularly useful interpretation
> How attractive is the strategy's return distribution after adjusting the return itself for investor loss aversion and then dividing by downside risk?

### Where does Prospect Ratio belong?

I would put it primarily in **downside-risk / asymmetric-return family**, with a secondary connection to **behavioral measures**.

The hierarchy could look roughly like this:

```text
Return / risk measures
│
├── Total-risk measures
│   ├── Sharpe Ratio
│   ├── M²
│   └── ...
│
├── Systematic-risk measures
│   ├── Treynor Ratio
│   ├── Modified Treynor
│   └── ...
│
├── Active-risk measures
│   ├── Information Ratio
│   ├── Tracking Error
│   └── ...
│
├── Downside-risk measures
│   ├── Sortino Ratio
│   ├── Kappa Ratios
│   ├── Omega Ratio
│   └── Prospect Ratio
│
└── Behavioral / preference-sensitive measures
    └── Prospect Ratio
```

The last two branches overlap: **Prospect Ratio is fundamentally a downside measure with an explicit behavioral preference parameter.**

That makes it somewhat unusual in the library.

### Prospect Ratio and Sortino

The relationship to Sortino is particularly close. Sortino is
$$Sortino=\frac{E[R]-MAR}{LPM_2(MAR)^{1/2}}$$

It says
> Reward return above MAR, but penalize downside volatility.

Prospect Ratio says
> First modify the return according to how strongly the investor dislikes losses, then divide by downside volatility.

The difference is therefore
$$\boxed{Sortino:\quad\text{preference enters primarily through the denominator}}$$

versus
$$\boxed{Prospect:\quad\text{preference enters explicitly in the numerator}}$$

This is exactly the distinction Bacon points out in the passage you quoted.

### Prospect Ratio and Kappa

Kappa generalizes Sortino
$$K_\nu=\frac{E[R]-MAR}{LPM_\nu(MAR)^{1/\nu}}$$

Increasing the order $\nu$ makes the denominator increasingly sensitive to large losses.
For example, $$K_1,\quad K_2,\quad K_3,\quad K_4$ progressively change the **risk penalty**.

Prospect Ratio takes a different approach.
It keeps downside deviation in the denominator but changes the **return contribution**:
$$r_i\rightarrow r_i^+ + \lambda r_i^-$$

So you can think of the distinction as

| Measure        | Investor preference enters |
| -------------- | -------------------------- |
| Sortino        | downside risk              |
| Kappa          | downside-risk order        |
| Prospect Ratio | gain/loss weighting        |

This is why I think Prospect Ratio deserves its own property rather than being treated as just another Kappa variant.

### The role of $\lambda$ is particularly interesting

The parameter makes Prospect Ratio much more than a single fixed statistic. With $\lambda=1$, we have
$$r_i^+ + r_i^- = r_i$$

so
$$PR(1)=\frac{E[R]-MAR}{\sigma_D}$$

That is essentially the familiar downside-deviation-adjusted return measure.

As $\lambda$ increases, $\lambda>1$, losses receive progressively greater weight.

At the Watanabe/Bacon default $\lambda=2.25$, a loss is given 2.25 times the weight of an equivalent gain.

Therefore the Prospect Ratio lets you perform something that ordinary Sortino and Kappa don't naturally provide:
> Sensitivity analysis over investor loss aversion.

### Why this can be useful for trading strategies

Consider two strategies with the same average return and similar downside deviation.

- Strategy A: Returns are relatively balanced: $+1\%, +1\%, +1\%, -1\%, -1\%$
- Strategy B: Returns are more asymmetric: $+2\%, +2\%, +1\%, -3\%, -1\%$

Their conventional average-return statistics may look surprisingly similar.
But an investor who strongly dislikes losses will view Strategy B differently.

The Prospect Ratio makes that preference explicit.

With increasing $\lambda$, the contribution of the negative observations becomes increasingly important $\sum r^+ + \lambda\sum r^-$.

Thus the ratio can reveal that a strategy's attractive average return depends heavily on accepting occasional large losses.

### This is particularly relevant to trading strategies with asymmetric payoff distributions

Prospect Ratio can be interesting for strategies such as:

- trend-following,
- momentum,
- mean reversion,
- option-like strategies,
- strategies with frequent small wins and occasional large losses,
- strategies with frequent small losses and occasional large wins,
- tail-risk strategies,
- highly asymmetric systematic strategies.

For example, consider a strategy that produces many small profits but occasionally suffers a very large loss.

A conventional Sharpe ratio may not fully communicate the psychological attractiveness of that distribution.

The Prospect Ratio allows you to ask
> Does this strategy remain attractive when losses are valued more heavily than gains?

That is a very relevant question for an actual investor.

### A particularly useful application: strategy comparison

Suppose you have

| Strategy | Sharpe | Sortino | Prospect λ=2.25 |
| -------- | -----: | ------: | --------------: |
| A        |   1.20 |    1.70 |            1.35 |
| B        |   1.30 |    1.80 |            0.85 |

A conventional evaluation might select B.

But the Prospect Ratio tells you something important:
> B's attractive risk-adjusted return is less robust to loss aversion.

You could then examine
$$PR(0),\quad PR(1),\quad PR(2.25),\quad PR(3),\quad PR(4)$$

If Strategy A remains superior as $\lambda$ increases, it may be the more appropriate strategy for a loss-averse investor.

This makes Prospect Ratio particularly useful as a **robustness/sensitivity measure**, rather than simply another number on a leaderboard.

### It also complements your existing metrics very well

Given the metrics you've been implementing, I would see Prospect Ratio as part of a broader progression:

Sharpe
> How much excess return per unit of total volatility?

Sortino
> How much return above MAR per unit of downside volatility?

Kappa
> How does performance look when downside losses of different orders are emphasized?

Omega
> How much probability-weighted gain exists relative to probability-weighted loss around MAR?

Prospect Ratio
> How attractive is the return distribution when gains and losses are valued differently according to an investor's loss aversion?

That's a very nice progression.

### Prospect Ratio is not necessarily "better" than Sortino

I would resist treating Prospect Ratio as a superior Sortino ratio.
They answer different questions.

Sortino is comparatively objective:
$$\frac{\text{excess return}}{\text{downside deviation}}$$

Prospect Ratio introduces an explicit subjective parameter $\lambda$.
That means two investors can legitimately obtain different Prospect Ratios for the **same strategy**.

That's not a weakness—it is the point of the measure—but it means it should be interpreted differently.

For a quantitative research report, I would therefore report something like:
> Sortino = 1.42; Prospect Ratio ($\lambda=2.25$) = 0.91.

rather than replacing Sortino with Prospect Ratio.

### Where I would use it in your trading-strategy evaluation

I'd give it a **secondary but meaningful role**.

For example:

- Core performance
  - annualized return
  - Sharpe
  - Sortino
  - maximum drawdown
  - Calmar
- Benchmark-relative
  - Jensen alpha
  - Information Ratio
  - Treynor
  - Appraisal Ratio
- Distribution / downside analysis
  - Kappa 1–4
  - Omega
  - Prospect Ratio

And then use Prospect Ratio specifically to answer:
> How robust is the strategy's attractiveness to an investor who places greater psychological weight on losses?

That is a different and valuable question from "does the strategy have a good Sharpe ratio?"

Making `lambda_loss: float = 2.25` function parameter rather than baking 2.25 into the strategy object is actually very appropriate.

It allows you to treat $\lambda$\)$ as an **investor preference dimension**.

You could evaluate a strategy's Prospect Ratio curve:
$$\lambda\mapsto PR(\lambda)$$

Conceptually

```text
Prospect Ratio
     │
     │\
     │ \
     │  \
     │   \
     │    \
     └──────────── λ
       0   1  2.25  4
```

For a strategy with meaningful losses, the ratio will generally decline as $\lambda$ increases.

That curve can be more informative than a single Prospect Ratio number because it tells you **how sensitive the strategy's attractiveness is to loss aversion**.

So in this library I would regard `prospect_ratio()` as a **downside/distribution measure with an investor-preference parameter**, sitting alongside Sortino, Kappa and Omega—not as a replacement for any of them.

## Bernardo-Ledoit ratio

It looks similar to Kappa-1 because both use first-order downside quantities, but there is an important distinction: **Bernardo-Ledoit uses raw returns relative to zero, whereas your Kappa ratios use returns relative to MAR and put mean excess return in the numerator.**

### One subtle distinction from Kappa-1

It is worth keeping this distinction in your documentation because the formulas can look deceptively similar:
$$\text{Bernardo-Ledoit}=\frac{HPM_1(0)}{LPM_1(0)}$$
whereas your Kappa-1 is:
$$K_1=\frac{E[R-MAR]}{LPM_1(MAR)}$$.

So **Bernardo-Ledoit is fundamentally an upside/downside balance measure**, while **Kappa is an excess-return/downside-risk measure**. The former does not require a MAR; the latter does.

## MAR-based downside-risk-adjusted return ratios

omega_sharpe_ratio, sortino_ratio, kappa_{1-4}_ratio, sortino-satchell_ratio — **conceptually, they belong to the same family**, but there is an important distinction in how that family is organized.

I would group them as **MAR-based downside-risk-adjusted return ratios**, with several different ways of measuring downside risk.

| Ratio                    | Numerator                             | Downside measure   | Order |
| ------------------------ | ------------------------------------- | ------------------ | ----: |
| `sortino_ratio`          | Mean excess return over MAR           | $\sqrt{LPM_2}$     |     2 |
| `kappa1_ratio`           | Mean excess return over MAR           | $LPM_1$            |     1 |
| `kappa2_ratio`           | Mean excess return over MAR           | $\sqrt{LPM_2}$     |     2 |
| `kappa3_ratio`           | Mean excess return over MAR           | $LPM_3^{1/3}$      |     3 |
| `kappa4_ratio`           | Mean excess return over MAR           | $LPM_4^{1/4}$      |     4 |
| `sortino_satchell_ratio` | Mean excess return over MAR           | $\sqrt{LPM_2}$     |     2 |
| `omega_sharpe_ratio`     | Upside potential - downside potential | $LPM_1$ implicitly |     1 |

### The Kappa family is the clearest core

The cleanest way to think about your library is:
$$\boxed{K_n = \frac{E[R-MAR]}{\sqrt[n]{LPM_n(MAR)}}}$$

Then:

* **Kappa-1** = first-order downside risk
* **Kappa-2** = second-order downside risk
* **Kappa-3** = third-order downside risk
* **Kappa-4** = fourth-order downside risk

And your **Sortino ratio is effectively Kappa-2**.

So I would actually consider `sortino_ratio` to be a **named special case of the Kappa family**, rather than a completely separate family.

### Sortino-Satchell is very close

Your `sortino_satchell_ratio` has the same mathematical structure as Kappa-2:
$$\frac{E[R-MAR]}{\sqrt{LPM_2(MAR)}}$$.

Therefore, **with your implementation, it is mathematically identical to `kappa2_ratio`** if both use the same target returns and LPM definition.

That raises an important API/documentation question: if the two properties produce exactly the same number, you should explain why you expose both.
The distinction is primarily **historical/nomenclatural**, rather than mathematical in your implementation.

### Omega-Sharpe is the odd one out

`omega_sharpe_ratio` belongs to the same broad **downside-risk / target-return family**, but it isn't another Kappa ratio.
You have:
$$\Omega = \frac{UPM_1}{LPM_1}$$
and:
$$\text{Omega-Sharpe}=\Omega-1=\frac{UPM_1-LPM_1}{LPM_1}$$.

Since
$$UPM_1-LPM_1=E[R-MAR]$$,
you get:
$$\boxed{\text{Omega-Sharpe}=\frac{E[R-MAR]}{LPM_1}}$$
which is **exactly Kappa-1**, assuming the same MAR and partial-moment definitions.

So there is an even stronger relationship than it initially appears:
$$\boxed{\text{Omega-Sharpe} = \text{Kappa-1}}$$
for your implementation.

That's a very useful observation for your library.

### Your ratios form a surprisingly neat hierarchy

Under your definitions, you essentially have:

```text
                    MAR-based downside-adjusted ratios
                                  |
                 ┌────────────────┴────────────────┐
                 │                                 │
           Kappa family                    Omega-based measures
                 │                                 │
      ┌──────────┼──────────┐                      │
      │          │          │                      │
   Kappa-1    Kappa-2    Kappa-3 ...          Omega-Sharpe
      │          │                              │
      │          └── Sortino                    └── = Kappa-1
      │
      └── Omega-Sharpe
```

And:
$$\boxed{K_1 = \text{Omega-Sharpe}}$$
$$\boxed{K_2 = \text{Sortino}}$$
while `Sortino-Satchell` is also equivalent to Kappa-2 **given your implementation**.

So I would describe these not as six independent ratios, but as a **family of closely related target-return/downside-risk measures with substantial mathematical overlap**.

That is actually valuable for the documentation of your library: users can understand that changing the ratio often means changing **the downside-risk definition/order**, rather than switching to an entirely unrelated performance concept.

### And how omega ratio fit in this family?

Yes — **`omega_ratio` fits the same broad family, but it is better viewed as a neighboring branch rather than another Kappa/Sortino-style ratio.**

The key is that your `omega_ratio` can be rewritten in terms of the same quantities:

$$\Omega(MAR)=\frac{UPM_1(MAR)}{LPM_1(MAR)}$$
and, because
$$UPM_1(MAR)-LPM_1(MAR)=E[R-MAR]$$,
we get:
$$\Omega(MAR)=1+\frac{E[R-MAR]}{LPM_1(MAR)}$$.

Therefore, **with your definitions**:
$$\boxed{\Omega-1=K_1}$$
and, as we discussed,
$$\boxed{\text{Omega-Sharpe}=\Omega-1=K_1}$$.

So there is a very direct relationship:

```text
                 MAR-based performance measures
                              │
              ┌───────────────┴───────────────┐
              │                               │
        Kappa / Sortino                  Omega family
              │                               │
       ┌──────┼──────┐                        │
       │      │      │                        │
     Kappa1 Kappa2 Kappa3 ...               Omega
       │      │                               │
       │   Sortino                            │
       │                                      │
       └──────── Omega-Sharpe ────────────────┘
                 Omega - 1
```

### The important conceptual difference

**Kappa ratios** start with:
> How much mean excess return do I get per unit of downside risk?

For example:
$$K_n = \frac{E[R-MAR]}{\sqrt[n]{LPM_n}}$$.

The **Omega ratio** instead starts with:
> How much total upside potential do I have compared with total downside potential?

$$\Omega = \frac{UPM_1}{LPM_1}$$.

So Omega is more naturally an **upside/downside balance measure**, whereas Kappa is a **return/downside-risk measure**.

But at order 1 they meet mathematically:
$$K_1=\# \frac{E[R-MAR]}{LPM_1}\# \frac{UPM_1-LPM_1}{LPM_1}\Omega-1$$.

### This gives your library a particularly nice structure

I would organize the documentation conceptually like this:

**1. Kappa family — downside-risk hierarchy**

* `kappa1_ratio`
* `kappa2_ratio`
* `kappa3_ratio`
* `kappa4_ratio`

with different orders of LPM controlling sensitivity to large losses.

**2. Sortino family — named special cases**

* `sortino_ratio` → essentially **Kappa-2**
* `sortino_ratio_sqrt2` → normalized Sortino
* `sortino_satchell_ratio` → same mathematical structure as Kappa-2 in your implementation

**3. Omega family — upside/downside balance**

* `omega_ratio`
* `omega_sharpe_ratio` → **Omega − 1 = Kappa-1**

So I would **not put Omega itself inside the Kappa family**, even though it is mathematically connected to it. I'd describe it as a **closely related MAR-based partial-moment measure**.

And this relationship is worth documenting prominently because users might otherwise wonder why `omega_sharpe_ratio` and `kappa1_ratio` return exactly the same value in your implementation.

## Variability and Volatility Skewness

**Variability skewness** and **volatility skewness** are related measures that compare upside and downside dispersion relative to a **Minimum Acceptable Return (MAR)**. Unlike traditional volatility, which treats positive and negative deviations symmetrically, these measures distinguish between variability above and below the investor's target return.

### `variability_skewness`
(Bacon3):
> Although the rankings will be identical for consistency with other measures typically used with asset management, I prefer the square root of volatility skewness. To differentiate the term, I use the name variability skewness.
>
>$$\text{Variability skewness} = \frac{\text{Upside risk}}{\text{Downside risk}} = \frac{\sigma_U}{\sigma_D}$$

Variability skewness compares the second-order upper and lower partial moments:
$$\text{Variability Skewness}=\frac{UPM_2(MAR)}{LPM_2(MAR)}$$.

Because the deviations are squared, this is a comparison of **variance-like quantities**.

A value of **1** indicates equal upside and downside variability. Values above 1 indicate that returns exhibit greater variability above MAR, while values below 1 indicate greater variability below MAR.

This measure can therefore help distinguish between two strategies that have similar overall volatility but very different distributions of upside and downside variation.

### `volatility_skewness`

(Bacon3):
> A similar measure to omega but using the second partial moment is volatility skewness, (Rom and Ferguson, "A Software Developer's View: Using Post-Modern Portfolio Theory to Improve Investment Performance Measurement" (2001).) the ratio of the upside variance compared to the downside variance. Values greater than 1 would indicate positive skewness and values less than 1 would indicate negative skewness:
>
> $$\text{Volatility skewness} = \frac{\tilde{\sigma}_U^2}{\tilde{\sigma}_D^2}$$
>
> **Interpretation**
>
> This measure, in particular, rewards extreme positive events and penalises extreme negative events. For those like me who are unsure about the merits of over-rewarding extreme (potentially one-off) positive events, but are still concerned by negative extreme events, then perhaps the upside potential ratio is more appropriate.

Volatility skewness is the square root of variability skewness:
$$\text{Volatility Skewness}=\sqrt{\frac{UPM_2(MAR)}{LPM_2(MAR)}}$$.

Taking the square root puts the underlying quantities back onto the **return/volatility scale**, making the measure more directly comparable to familiar volatility measures.

Again, **1** represents equal upside and downside volatility. A value above 1 means upside volatility dominates, while a value below 1 means downside volatility dominates.

### Relationship between the two

The two measures contain the same information:
$$\boxed{\text{Volatility Skewness}=\sqrt{\text{Variability Skewness}}}$$

The distinction is therefore primarily one of **scale and interpretation**. `variability_skewness` works with squared deviations, while `volatility_skewness` works with their square roots.

Both measures are particularly useful when **upside variability is not considered harmful in the same way as downside variability**. By defining the comparison around MAR, they provide a more targeted view of the shape of an investment's return distribution than conventional symmetric volatility.

## Bernardo-Ledoit and Gain-Loss ratios

Let's write both out.

### Bernardo-Ledoit

Your implementation is:
$$BL = \frac{HPM_1}{LPM_1}$$.

For target 0:
$$HPM_1 = \frac{1}{N}\sum_{r_i>0} r_i$$
and
$$LPM_1 = \frac{1}{N}\sum_{r_i<0}|r_i|$$.

Therefore:
$$BL = \frac{\frac{1}{N}\sum_{r_i>0}r_i}{\frac{1}{N}\sum_{r_i<0}|r_i|} = \frac{\sum_{r_i>0}r_i}{\sum_{r_i<0}|r_i|}$$.

### Gain-Loss ratio

We calculate:

```python
winning_returns_sum / abs(loosing_returns_sum)
```

which is exactly:
$$GL = \frac{\sum_{r_i>0}r_i}{|\sum_{r_i<0}r_i|}$$.

Since the negative returns are negative,
$$|\sum_{r_i<0}r_i|=\sum_{r_i<0}|r_i|$$,
so:
$$\boxed{BL = GL}$$

The factor (1/N) in the partial moments simply cancels.

### So why do both exist?

They are essentially **two presentations of the same measure**.

The difference is conceptual:

* **Bernardo-Ledoit ratio** is expressed in terms of **first-order partial moments**.
* **Gain-Loss ratio** is expressed directly as **total gains divided by total losses**.

But mathematically, with target/threshold (0), they are identical.

This is actually similar to the relationship you discovered earlier between your Omega/Kappa measures. A lot of these performance ratios are different historical formulations of the same underlying quantities.

### One caveat: MAR

There is an important distinction from your other ratios.

Your Bernardo-Ledoit implementation uses:

```python
self._raw_pertial_moments
```

so it is relative to **zero**.

Your `gain_loss_ratio` also implicitly uses **zero** because it classifies returns simply as positive or negative.

Therefore they match.

If you changed Gain-Loss to classify gains/losses relative to MAR, then it would become:
$$\frac{\sum (R_i-MAR)*+}{\sum (MAR-R_i)*+}$$,
which is essentially:
$$\frac{UPM_1(MAR)}{LPM_1(MAR)}=\Omega(MAR)$$.

So you can think of these relationships as:
$$\boxed{\text{Gain-Loss}=# \text{Bernardo-Ledoit}\Omega(0)}$$
for your definitions.

And more generally:
$$\boxed{\Omega(MAR)=\frac{UPM_1(MAR)}{LPM_1(MAR)}}$$
so **Omega is essentially the MAR-generalized version of this gain/loss concept**.

I would therefore consider whether you really need both properties in the public API. If you keep both, I'd document `gain_loss_ratio` as the direct gain/loss formulation and `bernardo_ledoit_ratio` as the partial-moment formulation, explicitly noting that **they are equivalent when the partial moments are taken about zero**.

## D-ratio

D ratio is
$$D(R)=\frac{n_d\sum_t\max(-R_t,0)}{n_u\sum_t\max(R_t,0)}$$
where $n_d$ is number of negative observations, $n_u$ is number of positive observations.

The relationship to Bernardo–Ledoit is
$$BL=\frac{\sum R_t^+}{\sum |R_t^-|}$$
whereas D-Ratio is
$$D=\frac{n_d}{n_u}\frac{\sum |R_t^-|}{\sum R_t^+}$$
Thus
$$D=\frac{n_d}{n_u}\frac{1}{BL}$$

So D-Ratio is the inverse Bernardo–Ledoit ratio adjusted by the relative frequency of losing and winning observations.

### D-Ratio versus Gain/Loss Ratio

The ordinary Gain/Loss Ratio is generally:
$$GL=\frac{\text{average gain}}{\left|\text{average loss}\right|}$$
With $n_u=\text{number of gains},\qquad n_d=\text{number of losses}$,
that becomes
$$GL=\frac{\sum R^+/n_u}{|\sum R^-|/n_d}=\frac{n_d\sum R^+}{n_u|\sum R^-|}$$
Therefore
$$D=\frac{1}{GL}$$

So D-Ratio is essentially the inverse Gain/Loss Ratio.
And because Bernardo–Ledoit is
$$BL=\frac{\sum R^+}{|\sum R^-|}$$
we have
$$D=\frac{n_d}{n_u}\frac{1}{BL}$$
and
$$D=\frac{1}{GL}$$

This makes the metric much easier to understand conceptually.

### The three measures are therefore:

| Measure | Formula | Interpretation |
| --- | --- | --- |
| Bernardo–Ledoit | $$\frac{\sum R^+}{\sum R^-}$$ | Total upside vs. total downside |
| Gain/Loss | $$\frac{\text{mean gain}}{\text{mean loss}}$$ | Typical gain vs. typical loss |
| D-Ratio | $$\frac{n_d\sum R^-}{n_u\sum R^+}$$ | Inverse typical gain/loss |

One interesting consequence is that D-Ratio doesn't contain fundamentally new information relative to Gain/Loss Ratio - it is its reciprocal. Its value in PerformanceAnalytics is largely historical/conventional naming and the convenient "lower is better" orientation.

## Farinelli–Tibiletti Ratios and Their Relationship to Other Performance Measures

The **Farinelli–Tibiletti ratio** provides a useful unifying framework for several downside- and upside-sensitive performance measures. Rather than fixing the order of the upside and downside partial moments, it allows the two orders to be chosen independently.

For a minimum acceptable return (MAR), the general form is
$$FT(u,l)=\frac{UPM_u(MAR)^{1/u}}{LPM_l(MAR)^{1/l}}$$,
where $UPM_u$ is the upper partial moment of order $u$, and $LPM_l$ is the lower partial moment of order $l$.

This gives the measure considerable flexibility. Increasing the order places more emphasis on large deviations, while choosing different orders for the numerator and denominator allows upside and downside behavior to be weighted differently.

### Omega as the first-order case

The most important special case is
$$FT(1,1)=\frac{UPM_1(MAR)}{LPM_1(MAR)$$}.
This is exactly the **Omega ratio**:
$$\boxed{FT(1,1)=\Omega}$$.
Thus, Omega can be viewed as the first-order Farinelli–Tibiletti ratio, comparing total upside potential with total downside potential relative to MAR.

This also explains the relationship with the Kappa family. Since
$$UPM_1-LPM_1=E[R-MAR]$$,
we have
$$\Omega-1=# \frac{E[R-MAR]}{LPM_1}K_1##.
Consequently, under the definitions used in the library,
$$\boxed{\Omega-1=K_1=\text{Omega-Sharpe}}$$.

### Upside Potential Ratio

Another important special case is
$$FT(1,2)=\frac{UPM_1(MAR)}{\sqrt{LPM_2(MAR)}}$$.

This is the structure of the **Upside Potential Ratio**. It combines first-order upside potential with second-order downside risk.

This makes intuitive sense: the numerator measures the average magnitude of returns exceeding MAR, while the denominator penalizes downside deviations more strongly because they are squared.

### Volatility Skewness

With equal orders of two,
$$FT(2,2)=\frac{\sqrt{UPM_2(MAR)}}{\sqrt{LPM_2(MAR)}}$$.

This is the **volatility skewness** measure used in the library.

It compares upside and downside volatility rather than comparing return to downside risk. A value greater than one indicates greater volatility above MAR than below MAR, while a value below one indicates the opposite.

The corresponding variability measure,
$$\frac{UPM_2}{LPM_2}$$,
is simply the square of volatility skewness.

### A broader view of the Kappa family

The Kappa family takes a somewhat different perspective:
$$K_n = \frac{E[R-MAR]}{LPM_n(MAR)^{1/n}}$$.

It therefore keeps the numerator fixed—the mean excess return—and changes the order of downside risk.

Farinelli–Tibiletti generalizes this idea in another direction. Instead of fixing the numerator, it allows **both the upside and downside orders to vary**:
$$FT(u,l)=\frac{UPM_u^{1/u}}{LPM_l^{1/l}}$$.

This means that Kappa and Farinelli–Tibiletti ratios are related, but they answer somewhat different questions.

Kappa asks:
> How much excess return do I obtain for a given order of downside risk?

Farinelli–Tibiletti asks:
> How does upside potential of a chosen order compare with downside risk of another chosen order?

### A useful map

The relationships can be summarized as:

| Measure                | Farinelli–Tibiletti form |
| ---------------------- | ------------------------ |
| Omega                  | $FT(1,1)$                |
| Upside Potential Ratio | $FT(1,2)$                |
| Volatility Skewness    | $FT(2,2)$                |
| Variability Skewness   | $FT(2,2)^2$              |
| Kappa-1                | $\Omega-1$               |
| Omega-Sharpe           | $\Omega-1$               |
| Sortino                | Kappa-2                  |

This makes Farinelli–Tibiletti particularly useful as a **conceptual framework** for your library. Many apparently different performance ratios turn out to be different points—or simple transformations—within the same partial-moment framework.

The main advantage is not that every existing ratio should be replaced by an `F-T(u,l)` call. Named ratios such as Omega and Sortino remain valuable because they are familiar to practitioners. Rather, the Farinelli–Tibiletti formulation provides the mathematical structure that explains **why these measures are related and how they differ**.

## Testing Farinelli-Tibiletti

For the orders you identified:
$$FT(1,1)=\frac{UPM_1}{LPM_1}=\Omega$$
so compare against your `omega_ratio`.

For `(1, 2)`:
$$FT(1,2)=\frac{UPM_1}{\sqrt{LPM_2}}$$
so compare against your `upside_potential_ratio`.

For `(2, 2)`:
$$FT(2,2)=\frac{\sqrt{UPM_2}}{\sqrt{LPM_2}}$$
so compare against your `volatility_skewness`.

And, importantly, because you have the variability version:
$$FT(2,2)^2=\frac{UPM_2}{LPM_2}$$
so compare `ft * ft` with your `variability_skewness`.
Yes. There are a few more useful checks, but **the three you already have are the strongest named-ratio identities**:

* `FT(1,1)` = `omega_ratio`
* `FT(1,2)` = `upside_potential_ratio`
* `FT(2,2)` = `volatility_skewness`

Getting **14 decimal places** on all three is excellent evidence that your F–T implementation is consistent with your partial-moment calculations.

### Other combinations worth testing

You can test the relationship with your Kappa ratios indirectly.
For example:
$$FT(u,l)=\frac{UPM_u^{1/u}}{LPM_l^{1/l}}$$
doesn't directly equal Kappa, because Kappa has **mean excess return** in the numerator. However, for (u=1),
$$UPM_1 - LPM_1 = E[R-MAR]$$.
Therefore:
$$FT(1,1)-1=# \frac{UPM_1-LPM_1}{LPM_1}# \frac{E[R-MAR]}{LPM_1}K_1$$.

So you have another excellent test:
$$\boxed{FT(1,1)-1=K_1}$$
That's actually a particularly nice test because it connects **Farinelli–Tibiletti, Omega, and Kappa**.

### A second useful identity: FT(2,2)

You already test:
$$FT(2,2)=\text{volatility_skewness}$$.
But because your `variability_skewness` is the ratio of the second moments,
$$\text{variability skewness}=\frac{UPM_2}{LPM_2}$$,
you also have
$$\boxed{FT(2,2)^2=\text{variability_skewness}}$$.
This is another very strong internal consistency check.

### What about `(2,1)`, `(3,1)`, `(3,2)`, etc.?

These are perfectly legitimate F–T ratios:
$$FT(2,1)=\frac{\sqrt{UPM_2}}{LPM_1}$$
$$FT(3,1)=\frac{\sqrt[3]{UPM_3}}{LPM_1}$$
$$FT(3,2)=\frac{\sqrt[3]{UPM_3}}{\sqrt{LPM_2}}$$
and so forth.
But unless you have another independently implemented metric with exactly the same definition, **there isn't much value in testing them against another property**. You would essentially be testing the F–T implementation against itself.
You *can*, however, construct an independent reference calculation directly from the raw returns.
That is a worthwhile **independent implementation test**, particularly for `(2,1)`, `(3,1)`, `(3,2)`, and `(4,4)`.

### I'd organize your tests like this

| Test                                | Identity    | Strength |
| ----------------------------------- | ----------- | -------- |
| `FT(1,1) == Omega`                  | Direct      | ⭐⭐⭐⭐⭐    |
| `FT(1,2) == Upside Potential Ratio` | Direct      | ⭐⭐⭐⭐⭐    |
| `FT(2,2) == Volatility Skewness`    | Direct      | ⭐⭐⭐⭐⭐    |
| `FT(1,1) - 1 == Kappa1`             | Algebraic   | ⭐⭐⭐⭐⭐    |
| `FT(2,2)² == Variability Skewness`  | Algebraic   | ⭐⭐⭐⭐⭐    |
| `FT(2,1)` vs raw calculation        | Independent | ⭐⭐⭐⭐     |
| `FT(3,2)` vs raw calculation        | Independent | ⭐⭐⭐⭐     |
| `FT(4,4)` vs raw calculation        | Independent | ⭐⭐⭐⭐     |

## Rachev Ratio

The **Rachev ratio** is a tail-risk performance measure that compares the expected magnitude of favorable extreme returns with the expected magnitude of unfavorable extreme returns. Unlike measures such as the Sharpe or Sortino ratio, which summarize the entire return distribution or its downside around a target, the Rachev ratio focuses specifically on the **tails** of the distribution.

For lower-tail probability (\alpha) and upper-tail probability (\beta), it can be expressed conceptually as
$$\text{Rachev Ratio}=\frac{ES_{\text{upper}}(\beta)}{ES_{\text{lower}}(\alpha)}$$,
where (ES_{\text{upper}}) is the average return in the upper tail and (ES_{\text{lower}}) is the magnitude of the average loss in the lower tail.

### Interpreting the ratio

A Rachev ratio greater than 1 means that the expected magnitude of extreme gains exceeds the expected magnitude of extreme losses. A value below 1 indicates that the extreme downside is larger than the extreme upside.

For example, a ratio of 1.5 means that the average return in the selected upper tail is approximately 1.5 times the magnitude of the average loss in the selected lower tail.

The choice of (\alpha) and (\beta) is important. With

```text
alpha = 0.10
beta  = 0.10
```

the measure compares the lower 10% tail with the upper 10% tail. Using smaller probabilities, such as 5%, makes the measure more focused on extreme events.

### Why tail behavior matters

Two strategies can have almost identical average returns and volatility while having very different tail characteristics. One might occasionally experience catastrophic losses, while another might have relatively benign downside but occasional large gains.

The Rachev ratio is designed to reveal this difference.

This makes it particularly interesting for strategies involving **asymmetric payoff distributions**, such as option strategies, trend-following systems, alternative investments, and strategies where extreme gains and losses are important parts of the return profile.

### Rachev versus Sharpe and Sortino

The Sharpe ratio considers average excess return relative to overall volatility:
$$\frac{E[R-R_f]}{\sigma}$$.
The Sortino ratio replaces total volatility with downside deviation relative to MAR:
$$\frac{E[R-MAR]}{\sqrt{LPM_2}}$$.

The Rachev ratio goes further toward the tails. It does not primarily ask:
> How much return do I receive for my typical amount of risk?

Instead, it asks:
> When extreme outcomes occur, how large are my potential gains compared with my potential losses?

This makes it complementary to, rather than a replacement for, Sharpe and Sortino.

### An important distinction from your Omega implementation

The Rachev ratio and Omega ratio are both tail-oriented, but they use the distribution differently.

**Omega** considers *all* observations relative to a threshold:
$$\Omega(MAR)=\frac{UPM_1(MAR)}{LPM_1(MAR)}$$.

Rachev instead isolates specified tails. Thus, Omega gives a broader picture of the balance between gains and losses around MAR, whereas Rachev concentrates on **extreme upside versus extreme downside**.

This distinction is particularly useful in your library because it gives users several levels of distributional analysis:

* **Omega:** overall gain/loss potential relative to MAR.
* **Sortino:** return relative to downside risk.
* **Kappa:** return relative to increasingly higher-order downside risk.
* **Rachev:** extreme upside versus extreme downside.

### A caveat about the PerformanceAnalytics implementation

Your implementation deliberately follows the empirical definition used by **PerformanceAnalytics**, including its particular treatment of empirical tail thresholds. This matters because there are several possible ways to estimate Expected Shortfall from a finite sample, especially when the desired tail boundary falls between observations.

Therefore, the Rachev ratio should be regarded not simply as an abstract mathematical formula, but as a **sample-based tail estimator whose precise value depends on the chosen empirical convention**.

That is also why your 14-decimal-place conformance test against the R implementation is valuable: it confirms that your Python implementation reproduces the reference definition, including its finite-sample tail handling.

## Three Drawdown Concepts

Drawdown is not a single mathematical object. It depends on **what constitutes the reference point from which loss is measured** and, equally importantly, **what question the risk measure is trying to answer**.

The `Ratios` implementation deliberately maintains three different drawdown representations:

1. **Cumulative peak-to-valley drawdown** — a drawdown of the cumulative-return/equity path, primarily used to identify maximum drawdown.
2. **High-water-mark drawdown** — the drawdown experienced at every observation relative to the running high-water mark, used by Pain, Ulcer, and Martin-type measures.
3. **Continuous drawdown runs** — compounded losses during maximal consecutive runs of negative returns, used by the Burke ratio.

These measures are related, but they are **not interchangeable**. They summarize different aspects of downside experience.

### 1. Why do we need several kinds of drawdown?

Consider a portfolio whose equity evolves approximately as follows:

```text
             peak
              *
             / \
            /   \       *
           /     \     / \
          /       \   /   \
         /         \_/     \
```

There are several perfectly reasonable questions one might ask:

> **How far did the portfolio fall from a previous peak?**

That is the classical **peak-to-valley drawdown** question.

> **At each observation, how far below its high-water mark was the portfolio?**

That produces a **drawdown time series**, which is useful when we care about the *duration and persistence* of investor pain.

> **How severe were the individual episodes of consecutive losses?**

That is a different question. A sequence of `-1%, -2%, -3%` represents one continuous losing episode, and its compounded loss is approximately `-5.9%`. The **Burke ratio** wants to penalize that episode as one event rather than three unrelated observations.

These distinctions are illustrated schematically below.

![Image](https://images.openai.com/static-rsc-4/bfyU9Ru7WqbcMEJPApblxVdOOK4zlbR_ZbOatCI004EJnlfCyWH2SNlXdbRBEtPU4AnA3U157goaVTyr78urBqS4FRhbo9eMiKWH8XrO8QdU6QDkGbNv0GedEp843-oHiVXDAeS_8_ybnWoCr7Att_JWeJaY-IrW6vcsnFCEnYm7pH95-_TCaj4ab6bZ1d7q?purpose=fullsize)

![Image](https://images.openai.com/static-rsc-4/EJB4sEa-AoQMEQetQT2MVTKha5grBe-NoUGANSn6h7yzk3xfWgYYYawtx6RxkjiQXO1j5epCFLQUhHUMU8oFo6iSRoOKmiUBb7JH8CmAMaE3h9NBHAlA0WptQrXVAIxSb0c8pzD5Iax5DC6UraSO6rFoShvP8-U12krUTGF-Oh9gxl7PCmiRMCRX_no9f9F0?purpose=fullsize)

![Image](https://images.openai.com/static-rsc-4/LeiIKgw4dzBMMwMGvEZlxJZHRChxyxVc6PBCzJkyPAxLacMeUt8zJiGA_SG8sPnV8ps1ZRDj9Xm9AlwPW1RCsUTv158bmQkPt9RGiWrPMnaxpclnZbFPpCARjFkjUBou0TyztTi_QqoIBl59rgcqmlE7nw9ER_9DeT27wIlPdqxxT5dAoNB-Yk51G-IIj6Cm?purpose=fullsize)

![Image](https://images.openai.com/static-rsc-4/p0dziFyLAM3xDsaCl-83h4LAC9DQ7LXg3hToRgCQQvaaWUYmjGdjwSfzpOBEuEBBtriTCzq50HiSlbZszZjaFNOHq2JdzyPCT7djaIeOr32xyb8nJOJKapK9x5ILgrqaIWCiSrPutFsPb5X_WpvEYmVH5T9zzubJHefXKMYwCRo0CgM0XXWw7MtyZWxkkRFi?purpose=fullsize)

### 2. Cumulative peak-to-valley drawdown

The first representation is the traditional drawdown calculated from the cumulative return series.

Suppose the periodic returns are (r_1,\ldots,r_t), expressed as decimal returns. The cumulative wealth process is
$$W_t = \prod_{i=1}^{t}(1+r_i)$$.

The running maximum wealth is
$$M_t = \max_{1\leq i\leq t} W_i$$.

The drawdown at time $t$ is therefore
$$\boxed{D_t^{\mathrm{cum}}=\frac{W_t}{M_t}-1}$$
or, in percentage form,
$$\boxed{D_t^{\mathrm{cum}} = \left(\frac{W_t}{M_t}-1\right)100}$$.

Because $W_t\leq M_t$,
$$D_t^{\mathrm{cum}}\leq0$$.

The maximum drawdown is simply
$$\boxed{\mathrm{MDD} = \min_t D_t^{\mathrm{cum}}}$$.

#### Example

Suppose: $r = [10%,-5%,-10%,20%]$.

The wealth path is
$1.000 \rightarrow 1.100 \rightarrow 1.045 \rightarrow 0.9405 \rightarrow 1.1286$.

The running maximum is
$1.000,\quad 1.100,\quad 1.100,\quad 1.100,\quad 1.1286$.

Thus the drawdown series is approximately
$0%,\quad -5%,\quad -14.5%,\quad 0%$.

The maximum drawdown is therefore $-14.5%$.
This is the classic **peak-to-valley** interpretation.

#### How we calculate it

Your driver maintains cumulative return:

```python
self._cumulative_return.update(ret)
```

and then calculates:

```python
dd = self._cumulative_return.geometric_return_plus_1 / self._cumulative_return.geometric_return_plus_1_max - 1
```

Thus:
$$D_t^{\mathrm{cum}}=\frac{W_t}{M_t}-1$$.

The resulting observations are stored in:

```python
self._drawdowns_cumulative
```

and fed into:

```python
self._drawdowns_cumulative_minmax.update(dd)
```

The important characteristic is that this representation is fundamentally about the **cumulative equity trajectory and its peaks and valleys**.

#### Best use

This is the natural representation for:

* Maximum Drawdown
* identifying peak-to-valley losses
* historical drawdown analysis
* determining the worst observed loss from a previous equity high
* describing portfolio "crashes" or major equity declines

It answers:

> **What was the worst loss from an accumulated wealth peak?**

### 3. High-water-mark drawdown

The second representation is more granular.
Instead of asking only about the worst peak-to-valley event, we calculate a drawdown **at every observation**.

Let $W_t$ be cumulative wealth and $H_t = \max_{1\leq i\leq t}W_i$ be the high-water mark up to time $t$.
Then
$$\boxed{D_t^{\mathrm{HWM}}=\frac{W_t}{H_t}-1}$$.

This looks mathematically identical to the cumulative drawdown formula above, and in an **expanding window** it is indeed the same concept.

The important distinction in your architecture is what happens when the calculation is performed over a **rolling window** and how the resulting entire drawdown series is subsequently aggregated.

Your `HighWaterMarkDrawdown` class explicitly maintains:

```python
_cumlog
_dd
_peak
_sum_dd
_sum_dd2
```

so that every observation in the current window has its own HWM-relative drawdown.

#### Why store the whole HWM drawdown series?

Because many risk measures don't care only about the single worst drawdown.

They care about **how much time the investor spends underwater and how deeply underwater the portfolio tends to be**.

Your class exposes $D_1,D_2,\ldots,D_n$ and calculates mean drawdown
$$\boxed{\overline{D}=\frac{1}{n}\sum_{t=1}^{n}D_t}$$
through `drawdowns_mean`.

Because drawdowns are non-positive, this quantity is also non-positive.
Its magnitude, $-\overline D$, can be interpreted as average depth below the high-water mark.

#### Mean squared drawdown

Your class also calculates
$$\boxed{\frac{1}{n}\sum_{t=1}^{n}D_t^2}$$
through `drawdowns_squared_mean` and therefore provides the ingredients for RMS-type downside measures.

The square is important because it makes deep drawdowns disproportionately expensive $(-20)^2 = 400$ whereas $(-5)^2 = 25$.

A 20% drawdown is therefore given **16 times the squared penalty** of a 5% drawdown.

### 4. Pain Index, Ulcer Index and Martin Ratio

This is why your comment in `Ratios.add_return()` is important:

```python
# High-water-mark drawdown used in
# Pain Index, Pain Ratio, Ulcer Index, and Martin Ratio
self._drawdown_high_watermark.update(ret)
```

These measures are fundamentally concerned with the **underwater experience of the portfolio**, rather than just the single worst drawdown.

For example, the Ulcer Index can be expressed as
$$\boxed{UI =\sqrt{\frac{1}{n}\sum_{t=1}^{n}D_t^2}}$$
where $D_t$ is the HWM-relative drawdown, usually expressed in percentage units.

This makes the Ulcer Index sensitive to both **depth** of drawdowns, and **persistence** of drawdowns.

That is an important distinction from Maximum Drawdown.
Consider two portfolios:

```text
Portfolio A:
0, 0, 0, -20%, 0, 0, 0

Portfolio B:
0, -10%, -10%, -10%, -10%, -10%, 0
```

Both may have a maximum drawdown around 20–40% depending on the exact path, but the investor experience is very different.

A measure based on the whole HWM drawdown sequence can capture that difference.

### 5. Continuous drawdown runs

The third concept is substantially different.

`ContinuousDrawdownRuns` does **not** calculate an equity drawdown from a high-water mark.

Instead, it identifies maximal runs of consecutive negative returns.

For example, `-1%,-2%,+1%,-3%,-4%,+2%,-5%` contains three losing runs: `[-1%,-2%],[-3%,-4%],[-5%]`.

The positive returns are separators. For each run $r_a,r_{a+1},\ldots,r_b$, the continuous drawdown is
$$\boxed{DD_j =\left[\prod_{i=a}^{b}(1+r_i)-1\right]100}$$.
Since every $r_i<0$, $DD_j < 0$.
For example, `-1%,-2%,-3%` produces `(0.99)(0.98)(0.97)-1=-0.058906`.
The important feature is that this is **one drawdown event**, not three.

#### Why use logarithms

Your implementation represents each run internally as a logarithmic sum:

```python
logr = math.log1p(ret * 0.01)
```

and for a run
$$L_j = \sum_{i=a}^{b}\log(1+r_i)$$
Then
$$DD_j =\\operatorname{expm1}(L_j)\times100$$.
This works because
$$\sum_i\log(1+r_i)=\log\left(\prod_i(1+r_i)\right)$$.
It also makes extending and shrinking a run particularly convenient:
$$L_{\mathrm{new}}=L_{\mathrm{old}}+\log(1+r)$$.
When an old observation leaves:
$$L_{\mathrm{new}}=\#\# L_{\mathrm{old}}\log(1+r_{\mathrm{old}})$$.

That is exactly what your streaming implementation does.

### 6. Why continuous runs are appropriate for Burke

The Burke ratio is concerned with the **severity of drawdown episodes**.

Your accumulator maintains
$$\sum_j DD_j^2$$
and exposes
$$B_{\mathrm{denominator}}=\sqrt{\sum_j DD_j^2}$$.
Thus, if the continuous drawdowns are $-5%,\quad -10%,\quad -3%$,
the denominator is
$$\sqrt{5^2+10^2+3^2}=\sqrt{134}\approx11.576%$$.
This treats each continuous losing episode as one observation.

That is fundamentally different from squaring every negative return separately.
For example, `-1%,-2%,-3%` is one losing episode: $DD=-5.8906%$.
Burke therefore sees approximately $DD^2=34.70$.
It does **not** see $1^2+2^2+3^2=14$.

That difference is intentional.

The continuous-run approach captures **episode severity**, including compounding.

### 7. The three measures answer three different questions

A useful way of thinking about the architecture is:

| Drawdown                  | Reference                         | Unit of analysis  | Main question                                             |
| ------------------------- | --------------------------------- | ----------------- | --------------------------------------------------------- |
| Cumulative peak-to-valley | Cumulative equity peak            | Peak/valley event | How bad was the worst decline?                            |
| High-water-mark           | Running equity peak               | Every observation | How deeply and persistently was the portfolio underwater? |
| Continuous run            | Previous return in a negative run | Losing episode    | How severe were consecutive-loss episodes?                |

Graphically:

![Image](https://images.openai.com/static-rsc-4/_vE08pcZWLCkf29yuyLrxu3EqnBqd9qo5C_bZT-pfOx7UxL-zuhMhqEIfwAYl1SCOEYj8Mo2CWbXkqrnq1H65rZz6p0DrDF9t3kZDSmYaQQCMUbnlDpY-T1MLo8QuuW0ybNF5xXGeXxasTgB04MSHER8F8qWceHXTUG6xWbfudwy-LxoU49_HMzLNY-yYhI8?purpose=fullsize)

![Image](https://images.openai.com/static-rsc-4/bfyU9Ru7WqbcMEJPApblxVdOOK4zlbR_ZbOatCI004EJnlfCyWH2SNlXdbRBEtPU4AnA3U157goaVTyr78urBqS4FRhbo9eMiKWH8XrO8QdU6QDkGbNv0GedEp843-oHiVXDAeS_8_ybnWoCr7Att_JWeJaY-IrW6vcsnFCEnYm7pH95-_TCaj4ab6bZ1d7q?purpose=fullsize)

![Image](https://images.openai.com/static-rsc-4/LeiIKgw4dzBMMwMGvEZlxJZHRChxyxVc6PBCzJkyPAxLacMeUt8zJiGA_SG8sPnV8ps1ZRDj9Xm9AlwPW1RCsUTv158bmQkPt9RGiWrPMnaxpclnZbFPpCARjFkjUBou0TyztTi_QqoIBl59rgcqmlE7nw9ER_9DeT27wIlPdqxxT5dAoNB-Yk51G-IIj6Cm?purpose=fullsize)

### 8. Cumulative drawdown versus HWM drawdown

These two are the easiest to confuse because their mathematical formulas look almost identical.

The distinction is primarily **what information you retain and what you subsequently do with it**.

A maximum-drawdown calculation ultimately asks $\min_t D_t$, so the entire path is reduced to its worst observation.

The HWM representation retains ${D_1,D_2,\ldots,D_n}$ because downstream measures use the entire sequence.

For example, $\mathrm{MDD} = \min(D_t)$ but
$$UI =\sqrt{\frac{1}{n}\sum_tD_t^2}$$.

These are mathematically different functionals of the same underlying HWM drawdown path.

This is why it is useful for `Ratios` to maintain the HWM drawdown accumulator separately rather than trying to derive everything from `maximum_drawdown`.

### 9. The rolling-window distinction

The rolling case makes the distinction even more important.
Suppose the window is $t-N+1,\ldots,t$.
When an old high-water-mark observation leaves the window, the applicable historical peak for subsequent observations may change.
Your `HighWaterMarkDrawdown` explicitly detects this:
```python
evicted_peak = old_dd == 0.0
```
and then
```python
if evicted_peak:
    self._recompute()
```

This is necessary because the drawdown history itself has changed.

By contrast, when the leaving observation was not a peak, the current HWM remains valid and the class can update in (O(1)).

This gives the accumulator a useful optimization $O(1)\quad\text{normally}$ and $O(N)\quad\text{when the relevant peak leaves the window}$.

That is a natural trade-off for an exact rolling HWM drawdown calculation.

### 10. Why continuous runs are not interchangeable with HWM drawdowns

Consider `+10%, -5%, -5%, +10%`. The HWM drawdown sequence is approximately $0,\quad -5,\quad -9.75,\quad 0$.
This tells us that the portfolio became increasingly underwater from its high-water mark.
The continuous-run representation is $-5%,-5%$ which produces $(0.95)^2-1=-9.75%%.
So the same underlying returns produce:

* an **observation-level underwater path** for HWM measures;
* a **single losing episode** for Burke.

Neither is "more correct". They answer different questions.

### 11. A useful conceptual diagram

The three perspectives can be summarized as follows:

![Image](https://images.openai.com/static-rsc-4/bfyU9Ru7WqbcMEJPApblxVdOOK4zlbR_ZbOatCI004EJnlfCyWH2SNlXdbRBEtPU4AnA3U157goaVTyr78urBqS4FRhbo9eMiKWH8XrO8QdU6QDkGbNv0GedEp843-oHiVXDAeS_8_ybnWoCr7Att_JWeJaY-IrW6vcsnFCEnYm7pH95-_TCaj4ab6bZ1d7q?purpose=fullsize)

![Image](https://images.openai.com/static-rsc-4/ovlhDqzSxa-05X0y7BJ8EylJVMcYDdipNXA_ZpspzxhJ1RFt0MPIqag36T2nyRz0-_AykG3v5wcZu8wnrOtyTULM80pYBG-jxdQaUNV67S_bgQLzN4Qc9SsWmUyXgbHhBjUMcl86a-w5wNEPnmUWP851bPpKa2lCdfDB-Ih5Kt38-jgb9aK_BN39lPTLAbLc?purpose=fullsize)

![Image](https://images.openai.com/static-rsc-4/JcRkwpgtTZW_mdQ4b3ln4FAVk-3QgUf5G7nuwAtnHesYMaZRqjZzjXSUx1q6gP7NgQVrp-qeREtvFactD8s2jLyaHufkajMPKayB3er6Sn9FdofWjoE0MvtuYe0FYPoHqf1DTb7YuRXfnnfMwPvCWAjOHJol6RsblhowVgOeRGhuxwlt7ArhrrZQS5NWF6b3?purpose=fullsize)

And conceptually:

```text
                    RETURN HISTORY
                          │
             ┌────────────┼────────────┐
             │            │            │
             ▼            ▼            ▼
       cumulative      running      negative
        equity          HWM          returns
             │            │            │
             ▼            ▼            ▼
       peak-to-valley   DD at each   continuous
        drawdowns       observation    runs
             │            │            │
             ▼            ▼            ▼
          MDD          Pain/Ulcer     Burke
```

### 12. When should each one be used?

#### Cumulative peak-to-valley drawdown

Use it when the primary concern is **the worst historical loss from a wealth peak**.
Typical applications:

* Maximum Drawdown
* reporting portfolio downside
* comparing strategies by worst historical loss
* stress/history analysis
* investor communication

It is intuitive:
> "The strategy lost 28% from its previous high."

#### High-water-mark drawdown

Use it when you care about the **entire underwater experience**.
Typical applications:

* Pain Index
* Pain Ratio
* Ulcer Index
* Martin Ratio
* measures involving average or squared drawdown

It answers:
> "How deeply and persistently does this strategy remain below its high-water mark?"

This is particularly useful when two strategies have similar maximum drawdowns but very different recovery characteristics.

#### Continuous drawdown runs

Use it when the unit of risk should be a **losing episode**, rather than an observation.
Typical applications:

* Burke ratio
* analysis of consecutive losses
* drawdown-event severity
* comparing strategies based on clustered downside returns

It answers:
> "When the strategy starts losing, how severe is the resulting continuous losing episode?"

### 13. SVG illustration: cumulative/HWM drawdown

Here is a simple SVG illustrating an equity path and its running high-water mark:

```svg
<svg xmlns="http://www.w3.org/2000/svg"
     width="760" height="260" viewBox="0 0 760 260">
  <polyline
      points="40,210 120,150 200,170 280,110
              360,140 440,90 520,120 600,70 700,100"
      fill="none"
      stroke="black"
      stroke-width="3"/>

  <polyline
      points="40,210 120,150 200,150 280,110
              360,110 440,90 520,90 600,70 700,70"
      fill="none"
      stroke="gray"
      stroke-width="2"
      stroke-dasharray="7 5"/>

  <line x1="40" y1="220"
        x2="700" y2="220"
        stroke="black"/>

  <text x="45" y="245">time</text>
  <text x="560" y="55">equity</text>
  <text x="505" y="88">running HWM</text>
  <text x="255" y="105">drawdown</text>
</svg>
```

The solid line represents cumulative equity; the dashed line represents the running high-water mark. The vertical separation between them is the HWM drawdown.

### 14. SVG illustration: continuous losing runs

```svg
<svg xmlns="http://www.w3.org/2000/svg"
     width="760" height="260" viewBox="0 0 760 260">

  <line x1="40" y1="130"
        x2="700" y2="130"
        stroke="black"/>

  <rect x="90" y="130"
        width="95" height="55"
        fill="none"
        stroke="black"/>

  <rect x="255" y="130"
        width="145" height="75"
        fill="none"
        stroke="black"/>

  <rect x="485" y="130"
        width="80" height="40"
        fill="none"
        stroke="black"/>

  <text x="105" y="220">run 1</text>
  <text x="305" y="235">run 2</text>
  <text x="500" y="190">run 3</text>

  <text x="45" y="118">0%</text>
  <text x="45" y="150">negative returns</text>

  <text x="45" y="255">
    Each maximal consecutive negative-return run becomes one compounded DD.
  </text>
</svg>
```

The important point is that the individual negative observations inside a rectangle are **one event** from the Burke perspective.

### 15. SVG illustration: the three perspectives

```svg
<svg xmlns="http://www.w3.org/2000/svg"
     width="760" height="360" viewBox="0 0 760 360">

  <text x="30" y="30" font-size="18">
    Same return history, three different questions
  </text>

  <text x="30" y="75">
    Cumulative peak-to-valley
  </text>

  <line x1="250" y1="70"
        x2="700" y2="70"
        stroke="black"/>

  <path d="M260 70 L340 40 L420 100 L500 45 L580 120 L680 60"
        fill="none"
        stroke="black"
        stroke-width="3"/>

  <text x="30" y="165">
    HWM drawdown series
  </text>

  <line x1="250" y1="160"
        x2="700" y2="160"
        stroke="black"/>

  <path d="M260 160 L340 160 L420 190 L500 160 L580 205 L680 160"
        fill="none"
        stroke="black"
        stroke-width="3"/>

  <text x="30" y="255">
    Continuous losing runs
  </text>

  <line x1="250" y1="250"
        x2="700" y2="250"
        stroke="black"/>

  <rect x="320" y="250"
        width="75" height="35"
        fill="none"
        stroke="black"/>

  <rect x="500" y="250"
        width="110" height="55"
        fill="none"
        stroke="black"/>

  <text x="30" y="335">
    Peak-to-valley, point-in-time path, and run severity answer different risk questions.
  </text>
</svg>
```

### 16. The essential distinction

The most concise way to characterize the three implementations is:
$$\\boxed{
\begin{array}{ll}
\textbf{Cumulative drawdown:}
&
\displaystyle
D_t=\frac{W_t}{\max_{s\leq t}W_s}-1
[1.5em]
\textbf{HWM drawdown:}
&
\displaystyle
{D_t}*{t=1}^{n}
\quad\text{retained as a complete underwater series}
[1.5em]
\textbf{Continuous drawdown:}
&
\displaystyle
DD_j=
\left(\prod*{i\in R_j}(1+r_i)-1\right)
\quad
\text{for each negative-return run }R_j.
\end{array}
}
$$
The first is primarily about **peak-to-valley loss**.

The second is about **depth and persistence of being underwater**.

The third is about **severity of consecutive losing episodes**.

That is why the three representations belong together in a performance/risk library, but should **not be collapsed into one generic "drawdown" calculation**. Each preserves information required by a different family of risk measures.

## Drawdown measures

The library uses **three different notions of drawdown**. They are related, but they answer different questions about downside risk:

1. **Cumulative peak-to-valley drawdown** — how far cumulative wealth falls from a previous peak.
2. **High-water-mark drawdown** — how far the portfolio is below its running high-water mark at every observation.
3. **Continuous losing-run drawdown** — how severe each uninterrupted sequence of negative returns is.

Using several definitions is necessary because no single drawdown measure captures all aspects of downside risk. Maximum drawdown emphasizes the **worst episode**, high-water-mark measures capture **depth and persistence**, while continuous losing-run drawdowns emphasize the **severity of individual losing sequences**.

### 1. Cumulative peak-to-valley drawdown

Cumulative drawdown is based on the evolution of cumulative geometric wealth.

Let $r_t$ be the return at observation $t$, expressed as a decimal, and let cumulative wealth be
$$W_t = \prod_{i=1}^{t}(1+r_i)$$.
The cumulative high-water mark is
$$H_t = \max_{1 \le i \le t} W_i$$.
The drawdown at time $t$ is
$$D_t^{\mathrm{cum}}=\frac{W_t}{H_t}-1$$.
Thus,
$$D_t^{\mathrm{cum}} \le 0$$.
The **maximum drawdown** is the most negative observation:
$$MDD = \min_t D_t^{\mathrm{cum}}$$.

The library exposes both representations:

- `min_drawdowns_cumulative` — signed maximum drawdown, e.g. (-0.25);
- `worst_drawdowns_cumulative` — magnitude of maximum drawdown, e.g. (0.25).

#### What it measures

This definition answers:
> **What was the largest decline from a previous cumulative wealth peak to a subsequent valley?**

It is an event-oriented measure. The important observation is the worst peak-to-valley decline, rather than how long the portfolio remained underwater or how many drawdown episodes occurred.

This makes it particularly useful for measures such as:

- Calmar Ratio
- Sterling Ratio
- reporting maximum historical loss from a peak.

### 2. High-water-mark drawdown

High-water-mark drawdown retains the drawdown **at every observation**, rather than reducing the entire history to its single worst value.

At observation $t$, let $H_t = \max_{1\le i\le t} W_i$ be the running high-water mark.
The high-water-mark drawdown is
$$D_t^{\mathrm{HWM}}=\frac{W_t}{H_t}-1$$.
Equivalently,
$$D_t^{\mathrm{HWM}}=\begin{cases}0, & W_t = H_t,[4pt]\dfrac{W_t}{H_t}-1, & W_t < H_t.\end{cases}$$

The resulting series contains one drawdown observation for every return observation.

Unlike maximum drawdown, it therefore preserves information about
> how frequently and how deeply the portfolio is underwater.

#### Mean drawdown: Pain Index

The Pain Index is the mean magnitude of the high-water-mark drawdowns:
$$\mathrm{Pain}=-\frac{1}{n}\sum_{t=1}^{n}D_t^{\mathrm{HWM}}$$.

Because $D_t^{\mathrm{HWM}}\le0$, the negative sign makes the Pain Index non-negative.

The corresponding Pain Ratio is
$$\mathrm{Pain\ Ratio}=\frac{R-R_f}{\mathrm{Pain}}$$,
where $R$ is the annualized geometric return and $R_f$ is the configured risk-free rate.

#### Ulcer Index

The Ulcer Index gives greater weight to deep drawdowns by using the squared drawdown:
$$\mathrm{UI}=\sqrt{\frac{1}{n}\sum_{t=1}^{n}\left(D_t^{\mathrm{HWM}}\right)^2$$}.

The corresponding Martin Ratio is
$$\mathrm{Martin\ Ratio}=\frac{R-R_f}{\mathrm{UI}}$$.

#### What it measures _

This definition answers:
> **How deeply and persistently was the portfolio below its high-water mark?**

Two portfolios can have exactly the same maximum drawdown but very different high-water-mark drawdown histories. One might recover immediately, while the other might remain underwater for a long period.

The Pain and Ulcer measures preserve this distinction.

### 3. Continuous losing-run drawdown

Continuous drawdown is based directly on **returns**, rather than on cumulative wealth relative to a historical peak.

Consider a maximal sequence of consecutive negative returns
$r_a,r_{a+1},\ldots,r_b,\qquad r_i<0$.

Its continuous drawdown is the compounded return over that entire losing run:
$$DD_j^{\mathrm{run}}=\prod_{i=a}^{b}(1+r_i)-1$$.

Each uninterrupted losing sequence therefore produces **one drawdown value**.
A non-negative return terminates the current losing run and separates it from the next one.
For example, $-2%, -3%, -4%]$ is one continuous losing run:
$DD^{\mathrm{run}}=\# (1-0.02)(1-0.03)(1-0.04)-1=-8.64%$.

The individual losses are not treated as three independent drawdowns.

#### Burke denominator

The Burke denominator aggregates the squared continuous drawdowns:
$$B_D=\sqrt{\sum_{j=1}^{m}\left(DD_j^{\mathrm{run}}\right)^2}$$.

The Burke Ratio is then
$$\mathrm{Burke}=\frac{R-R_f}{\sqrt{\sum_{j=1}^{m}\left(DD_j^{\mathrm{run}}\right)^2}}$$.

The modified Burke Ratio additionally scales this by the square root of the number of observations:
$$\mathrm{Modified\ Burke}=\mathrm{Burke}\sqrt{n}$$.

#### What it measures _2

This definition answers:
> **How severe are the portfolio's uninterrupted losing episodes?**

It is deliberately different from a high-water-mark drawdown. A portfolio can experience several consecutive losses while still being above its historical high-water mark, and those losses can nevertheless form a meaningful continuous losing episode.

This is why the Burke calculation needs its own drawdown definition.

### A common example

Consider the following sequence of returns:
$+10%,\quad -5%,\quad -2%,\quad +8%,\quad -3%,\quad -4%$

Starting with wealth $W_0=1$, cumulative wealth is:

| (t) | Return | Wealth (W_t) |      HWM |   HWM DD |
| --: | -----: | -----------: | -------: | -------: |
|   1 | (+10%) |     1.100000 | 1.100000 |  0.0000% |
|   2 |  (-5%) |     1.045000 | 1.100000 | -5.0000% |
|   3 |  (-2%) |     1.024100 | 1.100000 | -6.9000% |
|   4 |  (+8%) |     1.106028 | 1.106028 |  0.0000% |
|   5 |  (-3%) |     1.072847 | 1.106028 | -3.0000% |
|   6 |  (-4%) |     1.029934 | 1.106028 | -6.8800% |

This single example demonstrates the three definitions.

#### Cumulative peak-to-valley drawdown

The cumulative wealth reaches a first peak of $1.10$.
It then falls to $1.0241$.
Therefore, $\frac{1.0241}{1.10}-1=-0.069=-6.9%$.

Later, wealth reaches a new peak: $1.106028$.
The final value is approximately $1.029934$, giving $\frac{1.029934}{1.106028}-1=-6.88%$.
Therefore the maximum drawdown is $MDD=-6.90%$.

The first drawdown is slightly worse than the second.

#### High-water-mark drawdown

The complete HWM drawdown series is $0,;-5%,;-6.9%,;0,;-3%,;-6.88%$.
This contains information that maximum drawdown alone loses.
For example, the portfolio spends four observations below a high-water mark, with two separate underwater periods:
$-5%,-6.9%$ and $-3%,-6.88%$.

The Pain Index therefore considers **all six observations**, while the maximum drawdown only considers the worst observation.

The Ulcer Index goes one step further and gives disproportionately greater weight to the two deeper observations:
$-6.9%,\qquad -6.88%$.

#### Continuous losing-run drawdown

Now ignore the high-water mark and look only at consecutive negative returns.
There are two losing runs:

Run 1 $-5%,-2%$ with compounded loss $(1-0.05)(1-0.02)-1=\# 0.95\times0.98-1=-6.9%$.

Run 2 $-3%,-4%$ with compounded loss $(1-0.03)(1-0.04)-1=\# 0.97\times0.96-1=-6.88%$.

Thus, $DD^{\mathrm{run}}=[-6.9%,-6.88%]$ and the Burke denominator is
$$\sqrt{(-6.9%)^2+(-6.88%)^2}$$.

Notice the interesting result:
$\text{continuous runs}=[-6.9%,-6.88%]$ while
$\text{HWM drawdowns}=[0,-5%,-6.9%,0,-3%,-6.88%]$.

The first losing run happens to end at exactly the same drawdown as the HWM drawdown because the portfolio never makes a new high during that run. But conceptually the two measures are different.

### Why three definitions?

The distinction can be summarized as follows:

| Drawdown | Unit of analysis | Main question | Typical use |
| --- | --- | --- | --- |
| **Cumulative peak-to-valley** | Peak → subsequent valley | How bad was the worst loss from a peak? | Maximum Drawdown, Calmar, Sterling |
| **High-water mark** | Every observation | How deeply and persistently was the portfolio underwater? | Pain, Ulcer, Martin |
| **Continuous losing run** | Consecutive negative returns | How severe were uninterrupted losing episodes? | Burke |

Another useful way to view them is:
$$
\boxed{
\begin{aligned}
\text{Cumulative DD}
&\rightarrow \textbf{depth of the worst event},[4pt]
\text{HWM DD}
&\rightarrow \textbf{depth + persistence of being underwater},[4pt]
\text{Continuous DD}
&\rightarrow \textbf{severity of losing sequences}.
\end{aligned}}$$

They should therefore **not be treated as interchangeable implementations of "drawdown."**

### Rolling-window interpretation

The distinction becomes even more important when `Ratios` is configured with a rolling window.
For a window of $k$ observations, the relevant history is the current window: $r_{t-k+1},\ldots,r_t$.
The three measures then behave differently.

#### Cumulative drawdown

The cumulative wealth path is evaluated over the current window, and the peak-to-valley drawdown is determined from that path.

#### High-water-mark drawdown

Every observation in the current window receives a drawdown relative to the **highest cumulative wealth within that window**.

Consequently, when the old high-water-mark observation leaves the window, the HWM drawdowns may all need to be recomputed.

This is why `HighWaterMarkDrawdown` explicitly maintains the cumulative log-equity history and performs a recomputation when the relevant peak is evicted.

#### Continuous losing runs

The window contains a collection of maximal negative-return runs. When an observation leaves the window, the first losing run may shrink or disappear, which is why `ContinuousDrawdownRuns` supports `revert()`.

This is fundamentally a **return-sequence** calculation rather than a wealth-peak calculation.

### Graphic comparison

A clean three-panel SVG using the exact three definitions we established:

- Cumulative peak-to-valley drawdown
- High-water-mark drawdown
- Continuous losing-run drawdown

It uses the common example: $+10%,−5%,−2%,+8%,−3%,−4%$ and shows the corresponding formulas and applications to Calmar/Sterling, Pain/Ulcer/Martin, and Burke.

```svg
<svg xmlns="http://www.w3.org/2000/svg" width="1800" height="900" viewBox="0 0 1800 900">
<defs><style>
.title{font:700 24px Arial;fill:white}.h{font:700 20px Arial}.b{font:16px Arial}.f{font:italic 21px Georgia}
.panel{fill:white;stroke-width:2}.axis{stroke:#222;stroke-width:2}.line{fill:none;stroke:#2856a6;stroke-width:3}.peak{fill:none;stroke:#555;stroke-width:2;stroke-dasharray:5 5}.box{fill:#fafafa;stroke-width:1.5}
</style></defs>
<rect width="1800" height="900" fill="white"/>

<!-- 1 -->
<rect x="20" y="20" width="570" height="860" rx="16" class="panel" stroke="#2856a6"/>
<rect x="20" y="20" width="570" height="55" rx="12" fill="#2856a6"/>
<text x="42" y="56" class="title">1. CUMULATIVE PEAK-TO-VALLEY</text>
<text x="42" y="108" class="b">Cumulative wealth relative to its running peak.</text>
<text x="42" y="145" class="h" fill="#2856a6">Definition</text>
<text x="42" y="178" class="f">Wₜ = ∏ᵢ₌₁ᵗ(1+rᵢ),   Hₜ = max Wᵢ</text>
<rect x="85" y="198" width="440" height="62" rx="10" class="box" stroke="#2856a6"/>
<text x="115" y="238" class="f">Dᶜᵘᵐₜ = Wₜ/Hₜ − 1 ≤ 0</text>
<text x="42" y="295" class="b">Maximum drawdown: MDD = minₜ Dᶜᵘᵐₜ</text>
<text x="42" y="323" class="b">Signed: min_drawdowns_cumulative = MDD</text>
<text x="42" y="351" class="b">Magnitude: worst_drawdowns_cumulative = |MDD|</text>
<line x1="75" y1="520" x2="540" y2="520" class="axis"/><line x1="75" y1="390" x2="75" y2="520" class="axis"/>
<polyline points="100,455 175,410 250,432 325,470 400,420 475,456 525,500" class="line"/>
<polyline points="100,410 175,410 250,410 325,410 400,410 475,410 525,420" class="peak"/>
<g fill="#2856a6"><circle cx="100" cy="455" r="5"/><circle cx="175" cy="410" r="5"/><circle cx="250" cy="432" r="5"/><circle cx="325" cy="470" r="5"/><circle cx="400" cy="420" r="5"/><circle cx="475" cy="456" r="5"/><circle cx="525" cy="500" r="5"/></g>
<text x="220" y="455" class="b" fill="#b22">−6.90%</text><text x="470" y="490" class="b" fill="#b22">−6.88%</text>
<text x="42" y="570" class="h" fill="#2856a6">Common example</text>
<text x="42" y="602" class="b">Returns: +10%, −5%, −2%, +8%, −3%, −4%</text>
<text x="42" y="634" class="b">Drawdown: 0%, −5%, −6.90%, 0%, −3%, −6.88%</text>
<rect x="70" y="680" width="470" height="95" rx="10" class="box" stroke="#2856a6"/>
<text x="100" y="720" class="f">MDD = −6.90%</text><text x="100" y="755" class="f">|MDD| = 6.90%</text>
<text x="42" y="825" class="b">Use: Calmar and Sterling ratios.</text>

<!-- 2 -->
<rect x="615" y="20" width="570" height="860" rx="16" class="panel" stroke="#2f7d32"/>
<rect x="615" y="20" width="570" height="55" rx="12" fill="#2f7d32"/>
<text x="637" y="56" class="title">2. HIGH-WATER-MARK</text>
<text x="637" y="108" class="b">Drawdown below the running high-water mark at every t.</text>
<text x="637" y="145" class="h" fill="#2f7d32">Definition</text>
<text x="637" y="178" class="f">Hₜ = max Wᵢ,   Dᴴᵂᴹₜ = Wₜ/Hₜ − 1 ≤ 0</text>
<line x1="670" y1="520" x2="1135" y2="520" class="axis"/><line x1="670" y1="390" x2="670" y2="520" class="axis"/>
<polyline points="695,455 770,410 845,432 920,470 995,420 1070,456 1120,500" class="line"/>
<polyline points="695,455 770,410 845,410 920,410 995,420 1070,420 1120,420" class="peak" stroke="#2f7d32"/>
<g fill="#2856a6"><circle cx="695" cy="455" r="5"/><circle cx="770" cy="410" r="5"/><circle cx="845" cy="432" r="5"/><circle cx="920" cy="470" r="5"/><circle cx="995" cy="420" r="5"/><circle cx="1070" cy="456" r="5"/><circle cx="1120" cy="500" r="5"/></g>
<text x="637" y="570" class="h" fill="#2f7d32">Complete HWM series</text>
<text x="637" y="605" class="f">[0, −5%, −6.90%, 0, −3%, −6.88%]</text>
<rect x="650" y="650" width="235" height="120" rx="10" class="box" stroke="#2f7d32"/>
<text x="670" y="688" class="h" fill="#2f7d32">Pain Index</text><text x="670" y="728" class="f">−(1/n)ΣDᴴᵂᴹₜ</text><text x="670" y="755" class="b">mean drawdown magnitude</text>
<rect x="900" y="650" width="235" height="120" rx="10" class="box" stroke="#2f7d32"/>
<text x="920" y="688" class="h" fill="#2f7d32">Ulcer Index</text><text x="920" y="728" class="f">√[(1/n)Σ(Dᴴᵂᴹₜ)²]</text><text x="920" y="755" class="b">weights deep losses more</text>
<text x="637" y="825" class="b">Use: Pain, Pain Ratio, Ulcer, Martin Ratio.</text>

<!-- 3 -->
<rect x="1210" y="20" width="570" height="860" rx="16" class="panel" stroke="#6a35a8"/>
<rect x="1210" y="20" width="570" height="55" rx="12" fill="#6a35a8"/>
<text x="1232" y="56" class="title">3. CONTINUOUS LOSING-RUN</text>
<text x="1232" y="108" class="b">One compounded drawdown for each uninterrupted negative run.</text>
<text x="1232" y="145" class="h" fill="#6a35a8">Definition</text>
<text x="1232" y="178" class="f">DDʳᵘⁿⱼ = ∏ᵢ₌ₐᵇ(1+rᵢ) − 1 ≤ 0</text>
<line x1="1260" y1="355" x2="1735" y2="355" class="axis"/>
<rect x="1290" y="300" width="32" height="55" fill="#4f9d55"/><rect x="1360" y="355" width="32" height="28" fill="#c94b4b"/><rect x="1430" y="355" width="32" height="15" fill="#c94b4b"/><rect x="1540" y="315" width="32" height="40" fill="#4f9d55"/><rect x="1610" y="355" width="32" height="18" fill="#c94b4b"/><rect x="1680" y="355" width="32" height="25" fill="#c94b4b"/>
<text x="1350" y="415" class="b">−5%</text><text x="1422" y="398" class="b">−2%</text><text x="1602" y="405" class="b">−3%</text><text x="1672" y="412" class="b">−4%</text>
<text x="1360" y="445" class="b" fill="#6a35a8">Run 1</text><text x="1620" y="470" class="b" fill="#6a35a8">Run 2</text>
<rect x="1250" y="500" width="490" height="135" rx="10" class="box" stroke="#6a35a8"/>
<text x="1275" y="535" class="h" fill="#6a35a8">Same example</text>
<text x="1275" y="568" class="b">Run 1: (1−.05)(1−.02)−1 = −6.90%</text>
<text x="1275" y="600" class="b">Run 2: (1−.03)(1−.04)−1 = −6.88%</text>
<rect x="1250" y="675" width="490" height="95" rx="10" class="box" stroke="#6a35a8"/>
<text x="1275" y="712" class="h" fill="#6a35a8">Burke denominator</text>
<text x="1275" y="750" class="f">√ Σⱼ(DDʳᵘⁿⱼ)²</text>
<text x="1250" y="825" class="f">Burke = (R−Rᶠ) / √Σ(DDʳᵘⁿⱼ)²</text>
</svg>
```

### Pain versus Ulcer/Martin

**Pain Index and Ulcer Index are the two different ways of aggregating the same high-water-mark drawdown series**, while Pain Ratio and Martin Ratio turn those risk measures into return-to-risk ratios.

#### 1. Pain Index and Pain Ratio

Your implementation has:
$$D_t = \frac{W_t}{H_t}-1 \le 0$$
where $H_t$ is the running high-water mark.

The **Pain Index** is simply the average depth of those drawdowns:
$$\operatorname{Pain}=-\frac{1}{n}\sum_{t=1}^{n}D_t$$
So every observation contributes **linearly**.
For example, $D=[0,-2,-4,-6,0]$ gives $\operatorname{Pain}=\frac{0+2+4+6+0}{5}=2.4%$.

The corresponding **Pain Ratio** is:
$$\operatorname{Pain\ Ratio}=\frac{R-R_f}{\operatorname{Pain}}$$

So the question answered is:
> **How much excess return did I earn per unit of average drawdown pain?**

#### 2. Ulcer Index and Martin Ratio

The **Ulcer Index** uses exactly the same $D_t$, but squares it:
$$\operatorname{Ulcer}=\sqrt{\frac{1}{n}\sum_{t=1}^{n}D_t^2}$$

Using the same example: $D=[0,-2,-4,-6,0]$ we get
$$\operatorname{Ulcer}=\sqrt{\frac{0^2+2^2+4^2+6^2+0^2}{5}}=\sqrt{11.2}\approx3.35%$$.

The corresponding **Martin Ratio** is:
$$\operatorname{Martin\ Ratio}=\frac{R-R_f}{\operatorname{Ulcer}}$$

So it answers:
> **How much excess return did I earn per unit of drawdown severity, with deep drawdowns penalized disproportionately?**

#### Why do we need both?

The key difference is **linear vs quadratic penalty**.

Suppose we have two portfolios:

| Drawdowns          |  Pain | Ulcer |
| ------------------ | ----: | ----: |
| A: (0,-2,-4,-6,0)  | 2.40% | 3.35% |
| B: (0,-1,-1,-1,-9) | 2.40% | 4.11% |

Notice something interesting: $\operatorname{Pain}_A=\operatorname{Pain}_B=2.4%$.

The Pain Index considers them **equally painful on average**.

But the Ulcer Index says B is worse, $\operatorname{Ulcer}_A\approx3.35%$ versus $\operatorname{Ulcer}_B\approx4.11%$.

Why? Because $9^2=81$ dominates the Ulcer calculation.

The Ulcer Index therefore strongly penalizes **deep drawdowns**, whereas Pain Index treats every percentage point of drawdown equally.

#### This is why Martin is different from Pain

You can think of the progression as:

$$\boxed{\text{Drawdown series}\rightarrow
\begin{cases}
\text{mean absolute depth} &\rightarrow \text{Pain}[3pt]
\text{RMS depth} &\rightarrow \text{Ulcer}
\end{cases}}$$

and then:
$$\boxed{\text{Pain}\rightarrow\text{Pain Ratio}}$$

versus

$$\boxed{\text{Ulcer}\rightarrow\text{Martin Ratio}}$$.

The important conceptual distinction is therefore **not really Pain Ratio vs Martin Ratio**. It is:
> **Pain Index vs Ulcer Index.**

The ratios merely use those two different denominators.

#### Why the Ulcer Index is particularly useful

The Ulcer Index has another desirable property: it captures both **depth and persistence**.
Consider $A=[0,-6,0,0,0]$ and $B=[0,-6,-6,-6,-6]$.
Both experienced a 6% drawdown, but B stayed underwater.
For A $\operatorname{Ulcer}_A=\sqrt{\frac{36}{5}}\approx2.68%$.
For B $\operatorname{Ulcer}_B=\sqrt{\frac{4(36)}{5}}\approx5.37%$.

So the Ulcer Index recognizes that **being 6% below the high-water mark for four periods is much worse than briefly touching −6%**.

That makes it particularly appropriate for measuring the investor experience of a strategy whose problem is not merely its worst loss, but **how deeply and how persistently it remains underwater**.

#### Why Martin ratio isn't called "Ulcer Ratio"?

The Ulcer Index was developed by Peter Martin and Byron McCann as a measure of downside risk based on percentage drawdowns from a high-water mark. Martin subsequently used the Ulcer Index in a return/risk ratio that became known as the Martin Ratio.

So "Martin" is essentially an eponym for the ratio, while "Ulcer" identifies the underlying risk measure.

In financial literature, Martin Ratio is the recognized name. Calling it "Ulcer Ratio" could make your implementation look like you're defining a new ratio rather than implementing the established Martin measure.

## Calmar and Sterling ratios

Calmar and Sterling are worth separating from Pain/Martin/Burke because they use **drawdown as a relatively simple loss-of-capital denominator**, rather than trying to characterize the whole underwater experience or individual losing episodes.

### Calmar ratio

The Calmar ratio is essentially a return-to-maximum-drawdown measure:
$$\operatorname{Calmar}=\frac{\operatorname{CAGR}}{\left|\operatorname{MDD}\right|}$$
where $MDD$ is the worst peak-to-valley drawdown over the evaluation period.

This makes Calmar particularly intuitive:
> **How much compound annual return did the strategy generate for each unit of its worst historical loss?**

For example, if $\operatorname{CAGR}=12%$ and $\operatorname{MDD}=-3$0%, then $\operatorname{Calmar}=0.$40.

The strength of Calmar is also its weakness: **one exceptionally bad episode determines the denominator**. Two strategies with identical histories except for one unusually deep drawdown can have dramatically different Calmar ratios.

PerformanceAnalytics likewise defines Calmar as annualized return divided by the absolute maximum drawdown.

### Sterling ratio

Sterling was designed as a related return-to-drawdown measure, but with a **buffer added to the drawdown denominator**.

The traditional definition is approximately:
$$\operatorname{Sterling}=\frac{\operatorname{CAGR}}{|\operatorname{DD}|+10%}$$
where the 10% is the traditional Sterling excess-risk allowance.

PerformanceAnalytics explicitly implements this convention: `excess = 0.1` is the traditional/default amount added to the maximum drawdown.

Thus, if $\operatorname{CAGR}=12%,\qquad|\operatorname{MDD}|=30%$, then $\operatorname{Sterling}=\frac{12%}{30%+10%}=0.30$.

This makes Sterling more conservative than Calmar.

#### But why isn't the Sterling "excess" MAR or (R_f)?

This is the interesting part.

**Because the traditional Sterling excess is not a return hurdle. It is a fixed risk adjustment to the drawdown denominator.**

That distinction is important.

A risk-free rate or MAR answers a question like:
> "What return do I require before considering the investment's return satisfactory?"

The Sterling 10% answers something more like:
> "How much additional allowance should I add to the observed drawdown when judging its severity?"

So they belong to different conceptual dimensions.

##### Risk-free rate

$R-R_f$ is a **reward adjustment**.
It says:
> The investor could have earned (R_f) without taking the portfolio's risk, so only the excess return should receive credit.

This is why Pain Ratio, Martin Ratio and Burke Ratio naturally use $R_f$ in the numerator.

##### MAR

MAR is a **performance hurdle**.
MAR defines what constitutes an unacceptable return.

It is therefore conceptually different from Sterling's historical 10% adjustment.

##### Sterling's excess

Sterling instead modifies $|\operatorname{MDD}|$ into $|\operatorname{MDD}|+E$ where traditionally $E=10%$.

It is a **denominator adjustment**, not a return hurdle.

##### There is also an historical reason

The 10% should not be interpreted as a universal economic constant.

It is essentially a **historical convention**. PerformanceAnalytics explicitly describes 10% as the traditional Sterling excess.

There are also multiple variants of the Sterling ratio in the literature. Some later formulations replace the fixed 10% adjustment with an excess-return/risk-free-rate treatment and/or use average drawdowns rather than maximum drawdown.

### The four ratios now form a very nice conceptual family

You can summarize your drawdown ratios as:

| Ratio        | Reward  | Drawdown risk                | What it emphasizes          |
| ------------ | ------- | ---------------------------- | --------------------------- |
| **Calmar**   | CAGR    | Maximum drawdown             | Worst loss                  |
| **Sterling** | CAGR    | Maximum drawdown + allowance | Worst loss with buffer      |
| **Pain**     | $R-R_f$ | Mean HWM drawdown            | Average underwater pain     |
| **Martin**   | $R-R_f$ | RMS HWM drawdown             | Deep + persistent drawdowns |
| **Burke**    | $R-R_f$ | RMS continuous losing runs   | Severity of losing episodes |

So **Calmar and Sterling are deliberately much simpler** than Pain, Martin and Burke.

Calmar asks:
> How bad was the worst thing that happened?

Sterling asks:
> How bad was the worst thing that happened, after allowing a fixed buffer?

Pain asks:
> How much time did I spend underwater on average?

Martin asks:
> How severe were my underwater periods, giving extra weight to deep ones?

and Burke asks:
> How severe were my individual uninterrupted losing episodes?

That distinction explains why we needs all of these measures.

## Next

Next

## Timing Ratio

The **Timing Ratio** is a conditional beta measure designed to evaluate whether a portfolio has greater sensitivity to the benchmark during **rising markets than during falling markets**. Rather than treating the portfolio's beta as constant across all market conditions, it separates observations according to the direction of the benchmark's **excess return**.

It is defined as
$$\text{Timing Ratio}=\frac{\beta_{\text{bull}}}{\beta_{\text{bear}}}$$,
where $\beta_{\text{bull}}$ is the portfolio's conditional beta when benchmark excess returns are positive, and $\beta_{\text{bear}}$ is the conditional beta when benchmark excess returns are negative.

### Conditional beta

For each regime, beta is estimated in the usual way:
$$\beta=\frac{\operatorname{Cov}(R_p^e,R_b^e)}{\operatorname{Var}(R_b^e)}$$,
where $R_p^e$ and $R_b^e$ are portfolio and benchmark returns in excess of the risk-free rate.

This distinction is important. The classification is not simply based on whether the benchmark return is positive or negative.
An observation belongs to the "bull" regime when $R_b-R_f>0$, and to the "bear" regime when $R_b-R_f<0$.

Thus, the measure evaluates the portfolio relative to the return that could have been earned from the risk-free asset.

### Interpreting the ratio

A **Timing Ratio greater than 1** indicates that the portfolio has greater benchmark sensitivity in favorable market conditions than in unfavorable conditions. This is generally the desirable pattern: the portfolio participates more strongly when the benchmark's excess return is positive and has lower sensitivity when it is negative.

For example,
$$\beta_{\text{bull}}=1.2,\qquad \beta_{\text{bear}}=0.8$$
gives
$$\text{Timing Ratio}=\frac{1.2}{0.8}=1.5$$.

The portfolio therefore exhibits substantially greater exposure to the benchmark during the positive-excess-return regime.

Conversely, a ratio below 1 indicates that downside-market sensitivity is relatively larger than upside-market sensitivity.

### Timing Ratio versus ordinary beta

Ordinary beta compresses the relationship between a portfolio and benchmark into a single number:
$$\beta=\frac{\operatorname{Cov}(R_p,R_b)}{\operatorname{Var}(R_b)}$$.

That can hide an important asymmetry. A portfolio could have an overall beta of approximately 1 while having, for example, $\beta_{\text{bull}}=1.3$ and $\beta_{\text{bear}}=0.7$.

The unconditional beta would not communicate this particularly attractive conditional behavior, whereas the Timing Ratio would:
$$\frac{1.3}{0.7}\approx1.86$$.

The Timing Ratio therefore belongs to the family of **conditional performance measures**: it asks not merely how a strategy behaves on average, but whether its market exposure changes depending on the market regime.

### Relation to downside and upside measures

This makes the Timing Ratio conceptually complementary to the downside-risk measures in your library.

Measures such as **Sortino, Kappa, Omega, and Rachev** examine the distribution of portfolio returns from different perspectives. The Timing Ratio instead examines the **relationship between portfolio and benchmark**, specifically whether that relationship is asymmetric across market conditions.

A useful way to think about the distinction is:

| Measure          | Main question                                                        |
| ---------------- | -------------------------------------------------------------------- |
| Sharpe           | How much return do I get per unit of total volatility?               |
| Sortino          | How much return do I get per unit of downside risk?                  |
| Omega            | How do gains and losses balance around MAR?                          |
| Rachev           | How do extreme gains compare with extreme losses?                    |
| **Timing Ratio** | Does my benchmark exposure favor rising rather than falling markets? |

The Timing Ratio is consequently particularly relevant for strategies that attempt to **participate in upside markets while protecting capital during adverse markets**.

### A practical caveat

The ratio should not be interpreted as proof of genuine market-timing skill by itself. Conditional betas are estimated from subsets of observations, so they can be sensitive to sample size and to the particular risk-free rate used to define the regimes.

Nevertheless, it provides a useful diagnostic: **does the portfolio's market exposure have a favorable conditional asymmetry?**

Your implementation now reproduces the PerformanceAnalytics reference calculation to 14 decimal places, including the important treatment of **excess benchmark returns for determining the two regimes**. That makes it a particularly useful addition to the library's collection of asymmetric and downside-aware performance measures.

## Kelly Ratio and the testing of trading strategies

The Kelly criterion is fundamentally different from most of the performance ratios in a trading-strategy report. Sharpe, Sortino, Calmar, Martin, and Burke ratios primarily answer:
> How attractive is the strategy's historical return relative to some measure of risk?

The Kelly criterion asks a different question:
> Given the estimated distribution of returns, how much capital should be allocated to the strategy to maximize long-run compounded wealth?

That makes Kelly particularly interesting when **testing a trading strategy**, because it connects statistical properties of the strategy's returns with a theoretically motivated position size.

### The basic Kelly equation

For a strategy whose returns are measured in excess of the risk-free rate, $R_e = R-R_f$, the classical continuous-return approximation gives the full Kelly fraction as
$$f^* =\frac{\mathbb{E}[R_e]}{\operatorname{Var}(R_e)}$$
or, using the notation of your `Ratios` class,
$$f^* =\frac{\mu_e}{\sigma_e^2}$$.

The `kelly_ratio_full` therefore represents a **capital allocation or leverage fraction**, rather than a conventional risk-adjusted performance ratio.

For example, $\mu_e=0.001,\qquad\sigma_e^2=0.0004$ gives $f^*=\frac{0.001}{0.0004}=2.5$.

The theoretical full-Kelly allocation would therefore be **2.5× capital**, i.e. 250% exposure.

That is an important result: a strategy can have an apparently attractive Sharpe ratio while its estimated Kelly fraction is so large that applying full Kelly would be extremely aggressive.

### Why excess returns rather than raw returns?

Kelly is concerned with the incremental growth obtained from taking the strategy's risk.

If a risk-free asset already provides $R_f$, the relevant return from the risky strategy is $R_e=R-R_f$.

This also distinguishes Kelly from the metrics in your library that use **MAR**.

MAR represents a **minimum acceptable return**. Kelly is not asking whether returns satisfy a target. It is trying to determine the growth-maximizing exposure to a risky opportunity relative to the alternative of earning the risk-free rate.

### Full Kelly versus Half Kelly

The theoretical optimum is called **Full Kelly**:
$$f_{\mathrm{Full}}=\frac{\mu_e}{\sigma_e^2}$$
Your second property uses:
$$f_{\mathrm{Half}}=\frac12 f_{\mathrm{Full}$$}.

Thus, `kelly_ratio_full` is the theoretical Kelly allocation, while ``kelly_ratio` is your Half-Kelly allocation.

The reason Half Kelly is interesting for strategy testing is **estimation error**.

The Kelly fraction depends directly on the estimated mean $[f^*\propto\mu_e$.

Expected return is notoriously difficult to estimate. A small overestimate of the strategy's expected return can therefore produce a large overestimate of the supposedly optimal position size.

For example, $\mu_e=0.10%,\quad \sigma^2=0.05%$ might imply $f^*=2$.

But if the true expected return is only half the estimate, the optimal fraction is approximately half as large.

Consequently, full Kelly can be **statistically fragile even when the underlying strategy is sound**.

Half Kelly sacrifices some theoretical growth in exchange for substantially less exposure to estimation error and drawdown.

### Kelly is particularly useful in strategy testing

Suppose you have three strategies:

| Strategy | CAGR | Sharpe |  MDD | Full Kelly |
| -------- | ---: | -----: | ---: | ---------: |
| A        |  12% |    1.1 | −15% |        0.8 |
| B        |  12% |    1.1 | −15% |        2.5 |
| C        |  15% |    1.0 | −25% |        3.2 |

A conventional performance analysis might conclude that A, B and C are broadly comparable.

Kelly tells a different story.

It says that the **estimated optimal exposure** differs dramatically.

This is valuable because a backtest doesn't only tell you whether a strategy made money. It can also tell you whether the strategy's statistical characteristics justify taking a large position.

### But never estimate Kelly from the whole backtest and stop there

This is probably the most important practical point.
Suppose your backtest contains 10 years of data and you calculate:
$$\hat f=\frac{\hat\mu_e}{\hat\sigma_e^2}$$.

That number is an **estimate**, not the true Kelly fraction.
The estimate of the mean is especially problematic.
If the true expected excess return is close to zero, sampling noise can make: $\hat\mu_e>0$ and therefore produce a positive Kelly fraction even when the strategy has no genuine positive edge.

This makes Kelly particularly useful as a **diagnostic**, rather than simply as an instruction to leverage the strategy.

### How I would use Kelly when testing a trading strategy

A good workflow is:

#### 1. Establish the strategy's statistical edge

First examine:

* mean return
* volatility
* Sharpe
* Sortino
* skewness
* kurtosis
* maximum drawdown
* Calmar
* Pain/Martin
* Burke

Then calculate:
$$\hat f_{\mathrm{Kelly}}=\frac{\hat\mu_e}{\hat\sigma_e^2}$$.

Kelly should be viewed as an additional dimension of the analysis, not a replacement for these measures.

#### 2. Calculate Kelly out-of-sample

Don't calculate Kelly only on the complete historical dataset.

Instead use a walk-forward procedure:
$$\text{training period}\rightarrow\hat f\rightarrow\text{out-of-sample period}$$.

For example:

```text
2010–2014     estimate Kelly
2015          trade using that Kelly

2011–2015     estimate Kelly
2016          trade using that Kelly

2012–2016     estimate Kelly
2017          trade using that Kelly

...
```

This lets you determine whether the estimated Kelly fraction is **stable through time**.

That is much more informative than one number calculated from the entire backtest.

#### 3. Compare Full Kelly and Half Kelly

A particularly useful robustness test is $0.25f^*,\quad 0.50f^*,\quad 1.00f^*$.
Then examine the resulting equity curves.
If the strategy only looks attractive at $1.0f^*$ but becomes uncomfortable at $0.5f^*$, the strategy may be highly dependent on an aggressive estimate of its edge.

Conversely, if the strategy remains attractive at $0.25f^*$ or $0.5f^$*, that provides considerably more room for estimation error.

### Kelly and drawdown are complementary

This connects particularly well with the drawdown measures you've been implementing.

Kelly says approximately:
> How much should I bet?

Calmar says:
> How bad was the worst drawdown?

Pain says:
> How much time did I spend underwater?

Ulcer/Martin says:
> How severe were those underwater periods?

Burke says:
> How severe were the individual losing runs?

These answer different questions.

A strategy can have $\text{Kelly}=2.5$ but a very unpleasant drawdown profile.

That does **not** mean the Kelly calculation is necessarily wrong. It means that applying theoretical full Kelly could produce an unacceptable investor experience.

This is one of the strongest arguments for reporting **Kelly together with drawdown measures**, rather than presenting it in isolation.

### A useful strategy-testing table

For your library, I think a strategy report could have something like:

| Dimension       | Measure    | Question                                        |
| --------------- | ---------- | ----------------------------------------------- |
| Return          | CAGR       | How quickly did capital compound?               |
| Reward/risk     | Sharpe     | Return relative to total volatility?            |
| Downside        | Sortino    | Return relative to downside volatility?         |
| Worst loss      | Calmar     | Return relative to worst drawdown?              |
| Underwater      | Martin     | Return relative to depth/severity of drawdowns? |
| Losing episodes | Burke      | Return relative to continuous losing runs?      |
| Position sizing | Full Kelly | What is theoretical growth-optimal exposure?    |
| Position sizing | Half Kelly | What is a more conservative Kelly exposure?     |

That gives you both **strategy evaluation** and **position-sizing information**.

### One major warning: Kelly assumes more than the formula suggests

The elegant expression
$$f^*=\frac{\mu_e}{\sigma_e^2}$$
can give the impression that Kelly requires only a mean and variance.

It doesn't, in a strict theoretical sense.
The exact Kelly criterion maximizes expected logarithmic wealth
$$f^*=\arg\max_f E[\log(1+fR_e)]$$.

The familiar
$$\frac{\mu_e}{\sigma_e^2}$$
formula is an approximation under particular assumptions about returns and position sizing.

This matters enormously for trading strategies because returns can exhibit:

- skewness,
- fat tails,
- serial correlation,
- volatility clustering,
- regime changes,
- nonlinear payoffs,
- leverage constraints,
- transaction costs.

A strategy with substantial negative skewness, for example, can have a Kelly fraction from the mean/variance approximation that is considerably more aggressive than would be appropriate under its actual return distribution.

For serious strategy research, therefore, I would treat your `kelly_ratio_full` as the **mean-variance Kelly estimate**, not as an unquestionable "optimal leverage."

### The practical interpretation

For your library, I would summarize Kelly like this:
> The Kelly ratio estimates the theoretically growth-optimal allocation to a strategy from its expected excess return and variance. Full Kelly provides the theoretical maximum-growth allocation under the model assumptions, while Half Kelly deliberately reduces exposure to estimation error and model uncertainty. In trading-strategy testing, Kelly is best used as a position-sizing diagnostic and robustness measure rather than as a standalone measure of strategy quality. Its interpretation should be combined with out-of-sample testing, drawdown analysis, and sensitivity to fractional-Kelly allocations.

And there is a particularly useful research question to ask of every backtest:
> Is the estimated Kelly fraction stable out of sample?

In practice, I would consider **Kelly stability across rolling/walk-forward samples** considerably more informative than the single full-sample Kelly number.

## Testing CDaR beta and alpha

0. Definition

There are several definitions of CDaR in the literature.
The tests below target the discrete drawdown-event definition corresponding to the PerformanceAnalytics implementation.

Let benchmark drawdown episodes be $D_j<0,\qquad j=1,\ldots,N$, with each episode having a `from` and `trough` observation.

For confidence $c$, define $p=1-c$ and let $q_p = Q_p(D_1,\ldots,D_N)$ be the lower-tail quantile of drawdown depths.

Select
$$\mathcal T={j:D_j\leq q_p}$$.

For each selected episode define the portfolio return from its beginning to trough:
$$P_j = \begin{cases}
\displaystyle\prod_{t=f_j}^{\tau_j}(1+r_{p,t})-1, & \text{geometric}\\
\displaystyle\sum_{t=f_j}^{\tau_j}r_{p,t}, & \text{arithmetic}
\end{cases}
$$

Then our CDaR beta is
$$\beta_D=\frac{\sum_{j\in\mathcal T}P_j}{|\mathcal T|,q_p}$$

and CDaR alpha is
$$\alpha_D=R_{p,\mathrm{ann}}-\beta_D R_{m,\mathrm{ann}}.

With that definition established, there are some excellent invariant tests.

1. Identity test: portfolio == benchmark

This is probably the **single most important test**.

If $R_p=R_m$ then every portfolio drawdown return equals the corresponding benchmark drawdown return.

Therefore, $\beta_D=1$ and $\alpha_D=0$.

2. Constant leverage and scaling test

Suppose portfolio returns are a constant multiple of benchmark returns, $r_{p,t}=k,r_{m,t}$.

For the **arithmetic** version, this gives a particularly clean property.

Within every selected drawdown, $P_j=kD_jS.

Therefore,
$$\beta_D=\frac{k\sum D_j}{Nq_p}$$.

Notice something subtle: $\beta_D \neq k$ in general.
That's because the denominator uses the **drawdown quantile**, not the mean selected drawdown.

This is actually a useful test because it prevents you from accidentally implementing the tempting but incorrect:
$$\beta=\frac{\text{mean portfolio drawdown}}{\text{mean benchmark drawdown}}$$.

However, for geometric returns, $r_{p,t}=k r_{m,t}$ does **not** imply $P_j=kD_j$.
So don't expect the same scaling property.

3. No drawdowns

If the benchmark never experiences a drawdown, $D_j=0$ for all $j$, then CDaR beta is undefined because $q_p=0$.
Therefore, $\beta_D=\mathrm{NaN}$.

4. Zero-beta portfolio

Construct a portfolio with $R_p=0$ for the whole sample.

Then $\beta_D=0$ and $\alpha_D=0$ because every selected portfolio episode return is zero.

## CDaR Alpha and Beta: Portfolio Metrics or Trading-Strategy Metrics?

Conditional Drawdown at Risk (CDaR) is a drawdown-based analogue of Conditional Value at Risk (CVaR).
Instead of asking
> How bad are the worst return observations?

CDaR asks
> How bad are the worst drawdowns?

This makes it particularly relevant to investors and strategy developers because drawdown is path-dependent: two strategies can have identical average returns and volatility while having very different sequences of losses and recovery periods.

CDaR can be used not only for portfolio evaluation but also for trading-strategy testing.
In fact, CDaR alpha and beta can be especially informative when comparing strategies whose principal difference is the shape, persistence, or severity of their drawdowns.

### What is CDaR?

Suppose an equity curve generates a sequence of drawdowns $D_1,D_2,\ldots,D_n$, where drawdowns are conventionally represented as non-positive numbers, or equivalently as positive losses depending on the implementation.

The maximum drawdown is simply $MDD = \max D_i$ when drawdown magnitude is represented positively.

CDaR goes further. For a chosen confidence level $\alpha$, it considers the tail of the drawdown distribution. Conceptually,
$$CDaR_\alpha=E[D\mid D\geq VaR_\alpha(D)]$$
for positive drawdown magnitudes.

Thus, if (CDaR_{95}) is 18%, the interpretation is roughly:
> Among the worst 5% of drawdown observations, the average drawdown is 18%.

The important point is that CDaR considers the distribution of drawdowns rather than the distribution of individual returns.
This gives it a different perspective from volatility, semideviation, Sortino ratio, or VaR.

### CDaR Beta

CDaR beta attempts to measure how strongly the drawdown behavior of one strategy or portfolio is related to the drawdown behavior of a benchmark.

The intuition is analogous to ordinary beta
$$\beta=\frac{\operatorname{Cov}(R_p,R_b)}{\operatorname{Var}(R_b)}$$.

But instead of examining ordinary returns $R$, CDaR beta focuses on the drawdown experience.

Depending on the particular definition being implemented, the exact formula can differ.
This distinction is important because there is not one universally standardized "CDaR beta" in the same way that ordinary CAPM beta has a widely accepted definition.

Conceptually, however, the question is
> When the benchmark experiences drawdowns, how sensitive is the strategy's drawdown behavior?

A CDaR beta greater than one suggests that the strategy tends to experience disproportionately large drawdowns relative to the benchmark's drawdown behavior.

A beta below one suggests less drawdown sensitivity.

A negative value, where the definition permits it, would indicate an inverse relationship between the two drawdown series.

This is potentially much more useful than ordinary beta for strategies whose risk is dominated by loss episodes rather than day-to-day return volatility.

### CDaR Alpha

CDaR alpha is analogous to alpha in a conventional performance model.

Ordinary CAPM alpha asks whether a portfolio produced more return than would be expected given its beta exposure
$$\alpha_p=R_p-R_f-\beta_p(R_b-R_f)$$.

CDaR alpha instead attempts to identify an abnormal component of drawdown-related performance after accounting for the strategy's exposure to benchmark drawdowns.
$$\text{CDaR Alpha}=\text{strategy drawdown performance}-\text{expected drawdown performance given CDaR beta}$$.

The precise mathematical form depends on the particular CDaR alpha definition.

The interpretation is therefore different from ordinary alpha:
> Does this strategy exhibit better or worse drawdown behavior than would be expected from its relationship with the benchmark?

A favorable CDaR alpha would indicate that the strategy manages drawdowns better than its benchmark exposure would suggest.

### Why CDaR is interesting for trading strategies

A trading strategy generates an equity curve, and that equity curve has a sequence of drawdowns:
$$0\rightarrow -2%\rightarrow -5%\rightarrow -1%\rightarrow 0\rightarrow -8%\rightarrow -12%\rightarrow -4%\rightarrow 0$$

The drawdown sequence contains information that ordinary return statistics lose.
Consider two strategies:

| Metric           | Strategy A | Strategy B |
| ---------------- | ---------: | ---------: |
| CAGR             |        15% |        15% |
| Volatility       |        12% |        12% |
| Sharpe           |        1.2 |        1.2 |
| Maximum drawdown |        25% |        25% |
| CDaR             |        10% |        18% |

The maximum drawdown alone says the strategies are identical. The Sharpe ratio also says they are identical.

But CDaR tells us something important:
> Strategy B spends more of its bad periods in severe drawdowns.

That may be extremely important for live trading.

A strategy that occasionally suffers one large but rapidly recovered drawdown can have the same maximum drawdown as a strategy that repeatedly remains underwater for extended periods. Their operational characteristics are very different.

### CDaR Beta in strategy testing

CDaR beta can therefore answer questions such as
> Does my strategy amplify market drawdowns?

Suppose you test a trend-following strategy against the S&P 500. You might find $\beta_{\text{ordinary}} = 0.35$ which looks attractive.
The strategy appears to have relatively little market exposure.

But suppose its drawdown behavior gives $\beta_{\text{CDaR}} = 1.15$.
That tells a very different story.

The strategy may have low return correlation with the market while nevertheless experiencing severe losses during the same adverse market regimes.

This can happen with strategies that have nonlinear exposures.

For example:

- short-volatility strategies,
- mean-reversion strategies,
- leveraged strategies,
- options strategies,
- carry strategies,
- liquidity-sensitive strategies.

Their ordinary beta can look innocuous while their tail/drawdown exposure is substantial.

CDaR beta can therefore provide another dimension of strategy diversification.

### CDaR Alpha can be even more interesting

Imagine two strategies with approximately the same CDaR beta $\beta_{CDaR,A}\approx\beta_{CDaR,B}$, but
$$CDaR_{\alpha,A}>CDaR_{\alpha,B}$$.

The implication is that Strategy A achieves better drawdown characteristics for approximately the same benchmark drawdown exposure.
This is potentially useful when selecting among strategies.

Instead of asking only
> Which strategy has the highest Sharpe ratio?
you can ask:
> Which strategy delivers the best return and drawdown characteristics relative to the drawdown risk it takes?

That is much closer to the way a portfolio manager actually experiences risk.

### CDaR versus maximum drawdown

Maximum drawdown $MDD = \max_t D_t$ is extremely intuitive.
But it has one major statistical weakness: it uses only one observation.

Suppose two strategies have $MDD_A=MDD_B=30%$.
That doesn't mean their drawdown distributions are equivalent.

Strategy A might have $5%,6%,7%,8%,30%$, while Strategy B has $15%,18%,21%,25%,30%$.
Both have a 30% maximum drawdown. But Strategy B has a much more persistent severe-drawdown profile.

CDaR captures this distinction.
This is the reason to regard CDaR as a natural companion to maximum drawdown, rather than a replacement for it.

### CDaR and drawdown duration

CDaR measures the magnitude of drawdowns, but not necessarily their duration.

Two strategies can have identical CDaR, $CDaR_{95}=20%$, yet one may recover in three months while another remains underwater for three years.

Therefore, for strategy evaluation I would consider CDaR together with:

- Maximum Drawdown
- CDaR
- Average Drawdown
- Drawdown Duration
- Maximum Drawdown Duration
- Recovery Time
- Ulcer Index
- Calmar/MAR ratio
- Sortino ratio

This gives a much more complete picture.

### One important implementation issue

There is a subtle point that is particularly relevant to your library:
> CDaR is inherently path-dependent.

You cannot generally calculate it correctly by treating independent daily returns as though they were independent drawdown observations.

You first need to construct the equity/wealth
$$W_t=W_{t-1}(1+r_t)$$
and then calculate its running high-water mark
$$H_t=\max_{s\leq t}W_s$$.
The drawdown is then
$$D_t=1-\frac{W_t}{H_t}$$.

Only after constructing this drawdown series does it make sense to calculate the CDaR distribution.

This is particularly important for rolling-window strategy testing. The beginning of a rolling window can affect the high-water mark and therefore the drawdowns observed inside the window.

### CDaR is therefore not "only a portfolio metric"

I would make the distinction this way.

Portfolio application
> How does this portfolio behave relative to a benchmark in adverse drawdown regimes?

Trading-strategy application
> How does this strategy's equity curve behave in its adverse regimes, and how does that behavior compare with a benchmark or another strategy?

Both are legitimate.
In fact, CDaR may arguably be more directly interpretable for a trading strategy than some traditional portfolio statistics because the strategy's equity curve is the object you actually have to survive operationally.

A trader ultimately doesn't experience "variance", they experience $\text{equity}\rightarrow\text{drawdown}\rightarrow\text{recovery}$.
CDaR describes an important part of that process.
Yes. Based on the implementation we've been discussing, I would add a section like this, emphasizing the **quantile/interpolation step, the drawdown series, and the distinction between CDaR alpha/beta and ordinary return-based alpha/beta**.

### Mathematical implementation of CDaR

In the implementation, CDaR is calculated from the drawdown process of the portfolio or strategy equity curve, rather than directly from the sequence of periodic returns. Let $r_t$ denote the periodic return and let $W_t$ denote the corresponding wealth process,
$$W_t = W_{t-1}(1+r_t)$$.
The running high-water mark is
$$H_t = \max_{1\leq i\leq t} W_i$$,
and the corresponding drawdown is
$$D_t = 1-\frac{W_t}{H_t}$$.

Thus $D_t\geq0$, with $D_t=0$ whenever the wealth process is at a new high-water mark. This representation is particularly useful for implementation because it preserves the path dependence of drawdowns: the same set of returns can produce a different drawdown profile when their temporal ordering is changed.

The CDaR calculation then treats the resulting drawdown observations as a distribution and focuses on its upper tail.
For a confidence level $p$, the drawdown quantile is $q_p = Q_D(p)$, where $Q_D$ is the empirical quantile function of the drawdown observations.
The implementation uses the Type-8 quantile definition, corresponding to R's `quantile(..., type = 8)` convention.
Importantly, the probability $p$ is expressed on the interval $[0,1]$. Thus, a 95% CDaR calculation uses $p=0.95$, not $p=95$.

The Type-8 quantile is based on the plotting-position parameter $m=\frac{p+1}{3}$, with the quantile interpolated between the appropriate ordered drawdown observations. This detail is important for reproducibility because different empirical quantile definitions can produce slightly different tail thresholds, particularly for relatively short samples.
Using the Type-8 definition makes the Python implementation directly comparable with the corresponding R `PerformanceAnalytics` calculations.

Once the tail threshold $q_p$ has been determined, CDaR represents the mean drawdown in the tail beyond that threshold.
In its continuous formulation,
$$CDaR_p=E[D\mid D\geq q_p]$$.

For an empirical sample, this corresponds to averaging the drawdown observations lying in the upper $1-p$ tail, with the quantile boundary handled consistently with the selected empirical-quantile convention.
Consequently, CDaR differs fundamentally from maximum drawdown: maximum drawdown depends on a single extreme observation,
$$MDD=\max_t D_t$$,
whereas CDaR incorporates information from the entire adverse tail of the drawdown distribution.

This distinction is particularly relevant in the implementation of rolling CDaR statistics.
When a rolling window is used, the drawdown process must retain the appropriate high-water-mark state and support the removal of observations leaving the window.
A drawdown is therefore not merely a transformed return observation; its value depends on the preceding wealth path.
The `HighWaterMarkDrawdown` and `ContinuousDrawdown` components provide the state required to maintain this path-dependent information while observations are added to and removed from a rolling sample.

For CDaR beta, the same principle is applied to the relationship between the strategy's and benchmark's drawdown behavior.
Rather than measuring sensitivity of periodic returns to benchmark returns, the statistic measures sensitivity in the drawdown domain.
CDaR alpha then represents the component of drawdown-related performance that is not explained by the strategy's benchmark-relative CDaR exposure.
This makes these measures conceptually analogous to conventional beta and alpha, while operating on a substantially different—and path-dependent—risk variable.

One consequence is that CDaR alpha and beta should not be interpreted as simply replacing conventional CAPM alpha and beta.
A strategy can have a relatively small ordinary market beta while having substantial drawdown sensitivity, particularly when its payoff distribution is nonlinear or its losses are concentrated in particular market regimes.
CDaR beta therefore provides complementary information about co-movement during adverse equity-curve regimes, while CDaR alpha asks whether the strategy's drawdown behavior is favorable or unfavorable relative to that exposure.

Overall, the mathematical implementation can therefore be viewed as a sequence of transformations:
$$\boxed{r_t\longrightarrow W_t\longrightarrow H_t\longrightarrow D_t\longrightarrow Q_D(p)\longrightarrow CDaR_p}$$
and, for the benchmark-relative measures,
$$\boxed{(D_t^{portfolio},D_t^{benchmark})\longrightarrow\beta_{CDaR}\longrightarrow\alpha_{CDaR}}$$

This formulation makes clear why CDaR is applicable to both portfolio analysis and trading-strategy testing.
In both cases the fundamental object is an equity/wealth process, and CDaR measures the statistical behavior of its adverse excursions from previous highs.

### Semantics

For a recovered episode:

```text
from_idx ───────── trough_idx ───────── to_idx
    │                   │                  │
  start               worst            recovery
```

For an unrecovered episode:

```text
from_idx ───────── trough_idx ───────── to_idx
    │                   │                  │
  start               worst          last observation
                                      (not recovery)
```

So:

- `from_idx`is the first underwater observation.
- `trough_idx` is the deepest observation.
- `to_idx` is the recovery observation **if recovered**, otherwise the last observation in the input.
- `recovered=True` means the `to_idx` is genuinely the recovery observation.
- `recovered=False` means the `to_idx` is simply the end of the available data.

## SFM Risk Premium

SFM RiskP remium is simply the arithmetic mean of the periodic excess returns.
$$\text{SFM Risk Premium}=\frac{1}{n}\sum_{i=1}^{n}(R_{p,i}-R_{f,i})=\overline{R_p-R_f}$$

It is **not annualized**, it has the same periodicity as the input returns, returning the mean of the supplied periodic excess returns.

It is closely related to your other properties:

```text
sfm_risk_premium
        │
        ├── numerator of Sharpe ratio
        ├── numerator of Treynor ratio
        └── component of SFM alpha
```

More specifically,
$$R_p-R_f=\alpha+\beta(R_b-R_f)+\epsilon$$

taking means gives
$$\overline{R_p-R_f}=\alpha+\beta\overline{(R_b-R_f)}$$

because the mean residual is zero in the OLS regression.

So it is a useful base quantity for the SFM family, not just an isolated one.

It answers a very simple question
> What average return did the strategy earn above the risk-free rate per observation?

### Why it is useful

For a trading strategy, it gives you the most basic measure of whether the strategy is being compensated above cash.

It is particularly useful as a component of other measures
$$Sharpe =\frac{\overline{R_p-R_f}}{\sigma_p}$$

and
$$Treynor=\frac{\overline{R_p-R_f}}{\beta}$$

It is also useful for interpreting the SFM decomposition
$$\overline{R_p-R_f}=\alpha+\beta\overline{R_b-R_f}$$.

That lets you distinguish the strategy's average excess return into

- compensation associated with benchmark exposure, and
- Jensen/SFM alpha.

### But by itself it tells you very little about risk

Suppose two strategies both have $\text{SFM Risk Premium}=0.05\%$ per day.
That could mean

- Strategy A: very stable returns, low drawdowns, low volatility.
- Strategy B: enormous volatility, huge drawdowns, occasional spectacular gains.

The risk premium alone cannot distinguish them.
For that reason, I would not use it to rank strategies.

I'd regard it as a foundation metric

```text
sfm_risk_premium
       │
       ├── Sharpe
       ├── Treynor
       ├── Jensen alpha
       └── other excess-return measures
```

Then the higher-level evaluation metrics answer progressively more interesting questions:

Risk premium
> How much excess return?

Sharpe / Sortino
> How much excess return per unit of total/downside risk?

Treynor
> How much excess return per unit of systematic exposure?

Jensen alpha
> How much return beyond that explained by systematic exposure?

Appraisal ratio
> How efficiently is alpha generated relative to specific risk?

## SFM alpha/beta architecture

The full beta calculation implements
$$\beta = \frac{\sum_i (x_i-\bar{x})(y_i-\bar{y})}{\sum_i (y_i-\bar{y})^2}$$
which is exactly the slope of the OLS regression of portfolio excess returns on benchmark excess returns.

When calculating beta we don't need the covariance/variance normalization
The `m2_ab` and `m2_bb` are both unnormalized sums of squares/cross-products:
$$M_{ab}=\sum (a_i-\bar a)(b_i-\bar b)$$
$$M_{bb}=\sum (b_i-\bar b)^2.$$
Therefore
$$\frac{M_{ab}/(n-1)}{M_{bb}/(n-1)}=\frac{M_{ab}}{M_{bb}}$$

This is also why the `ddof` setting has no effect on `beta`, although it does affect your `value` covariance properties.

The alpha formula calculates the OLS intercept
$$\alpha = \bar R_a^e-\beta\bar R_b^e$$
This implementation is statistically equivalent to
$$R_a-R_f=\alpha+\beta(R_b-R_f)+\epsilon$$

If there is no benchmark variance, then the OLS slope and therefore the OLS intercept are not identifiable.
Mathematically,
$$\operatorname{Var}(R_b^e)=0$$
means $\beta$ cannot be estimated.

Returning the portfolio excess mean is arguably a reasonable convention:
$$\alpha = E[R_a^e]$$
when $\beta$ is unavailable, but it is not the CAPM/SFM OLS alpha.

The inverse-Welford formula, or the reversal mathematicsis as follow.

Apart from those edge cases, the reversal mathematics is good.
Forward $\delta_a=a-\bar a_{\rm old}$ and $\delta_b=b-\bar b_{\rm old}$, then
$$\bar b_{\rm new}=\bar b_{\rm old}+\frac{\delta_b}{n}$$
and
$$M_{ab,new}=M_{ab,old}+\delta_a(b-\bar b_{\rm new})$$

The reverse calculation reconstructs
$$\bar a_{\rm old}=\frac{n\bar a_{\rm new}-a}{n-1}$$
and
$$M_{ab,old}=M_{ab,new}(a-\bar a_{\rm old})(b-\bar b_{\rm new})$$

Theoretically, $R_f$ can be a vector, but we support only $R_f = \text{constant}$ but not $R_{f,t}$.

In case of vector, if $R_f$ varies through time $x_t=R_{a,t}-R_{f,t}$ $y_t=R_{b,t}-R_{f,t}$, and the covariance is
$$\operatorname{Cov}(x,y)=\operatorname{Cov}(R_a-R_f,R_b-R_f)$$

Conceptually, `CovarianceBullBear` maintains three independent regressions.
- full $R_a^e = \alpha+\beta R_b^e+\epsilon$
- bull $R_a^e = \alpha^+ + \beta^+R_b^e+\epsilon^+\quad\text{when }R_b^e>04
- bear $R_a^e = \alpha^- + \beta^-R_b^e+\epsilon^-\quad\text{when }R_b^e<0$

Timing Ratio is
$$\text{Timing Ratio}=\frac{\beta^+}{\beta^-}$$

Yes. I would regard the SFM coefficients as **quite useful in backtesting**, but primarily as **diagnostic and attribution statistics**, rather than as standalone strategy-selection criteria.

## Single-Factor Model in portfolio backtesting

The Single-Factor Model (SFM) is the empirical regression
$$R_{p,t}-R_{f,t}=\alpha+\beta\left(R_{b,t}-R_{f,t}\right)+\epsilon_t$$
where $R_p$ is the strategy or portfolio return, $R_b$ is a benchmark, and $R_f$ is the risk-free rate.

Despite its historical association with the Capital Asset Pricing Model (CAPM), the regression itself does not require one to believe the full CAPM theory. It is simply a one-factor model describing the relationship between a strategy and a benchmark.

That distinction is important in backtesting. If your benchmark is the S&P 500, you can interpret the regression in a CAPM-like fashion. But if your benchmark is, say, a momentum index, sector ETF, or another strategy, calling the coefficient "CAPM beta" becomes unnecessarily restrictive. SFM is the more general and useful terminology.

### Beta: what kind of strategy have I actually built?

The most familiar coefficient is beta
$$\beta = \frac{\operatorname{Cov}(R_p^e,R_b^e)}{\operatorname{Var}(R_b^e)}$$

It answers a very practical backtesting question
> How strongly does my strategy respond to movements in the benchmark?

For example, $\beta=1.0$ means that, in the linear regression sense, a 1 percentage-point increase in benchmark excess return is associated with approximately a 1 percentage-point increase in strategy excess return.

A beta of:
- 0 means little linear exposure to the benchmark;
- 0.5 means roughly half the benchmark sensitivity;
- 1.0 means benchmark-like sensitivity;
- 1.5 means leveraged sensitivity;
- negative means the strategy tends to move opposite to the benchmark.

This is extremely useful when interpreting a backtest. Suppose you have two strategies:

|               | Strategy A | Strategy B |
| ------------- | ---------: | ---------: |
| Annual return |        12% |        12% |
| Volatility    |        14% |        11% |
| SFM beta      |       1.15 |       0.35 |
| SFM (R^2)     |       0.80 |       0.15 |

The two strategies have the same headline return, but they are doing very different things.
Strategy A is largely participating in the benchmark's risk premium.
Strategy B is much less dependent on it.
That distinction can be extremely important for portfolio construction.

### Alpha: did the strategy actually add something?

The intercept is
$$\alpha = \bar R_p^e-\beta\bar R_b^e$$

Conceptually, alpha asks
> After accounting for the strategy's linear exposure to the benchmark, is there residual return left over?

This is probably the most interesting SFM statistic from a backtesting perspective.
Imagine
$$R_p^e = \alpha+\beta R_b^e+\epsilon$$

A strategy can have a high return simply because it has a high beta.
For example, a strategy with $\beta=1.4$ might produce impressive returns during a strong bull market without actually demonstrating much independent skill.

Alpha attempts to separate those two effects.

A positive alpha means that, conditional on this particular single-factor model, the strategy has historically earned more than its benchmark exposure would predict.

But this qualification is crucial
> Positive alpha is not proof of trading skill.

It can arise from
- omitted risk factors;
- nonlinear exposures;
- volatility effects;
- regime dependence;
- data mining;
- survivorship bias;
- transaction-cost assumptions;
- an inappropriate benchmark;
- luck.

Use alpha as a diagnostic, not as a "skill detector."

### (R^2): how much of the strategy is explained by the benchmark?

$$R^2=\rho^2$$
is particularly useful alongside beta.

It answers
> How much of the variation in strategy returns is linearly explained by the benchmark?

For example, $R^2=0.80$ means that approximately 80% of the variation in the strategy's returns is explained by the one-factor linear relationship, while $R^2=0.10$ indicates a much weaker relationship.

This makes beta and (R^2) complementary:
$$\boxed{\beta=\text{magnitude of exposure}}$$
$$[\boxed{R^2=\text{strength of the linear relationship}}$$

A beta of 1.0 with $R^2=0.90$ is very different from a beta of 1.0 with $R^2=0.05$.
In the first case, the strategy behaves substantially like the benchmark.
In the second, the same beta is a relatively weak summary of the strategy's behavior.

### Bull and bear betas are particularly interesting for trading strategies

We have $\beta_{\text{bull}}$ estimated when $R_b-R_f>0$, and $\beta_{\text{bear}}$ estimated when $R_b-R_f<0$.

These answer
> Does the strategy respond differently to the benchmark in favorable and unfavorable environments?

Consider two strategies.

Strategy A has $\beta_{\text{bull}}=1.2$ and $\beta_{\text{bear}}=1.1$.
It participates strongly in both directions.

Strategy B has $\beta_{\text{bull}}=1.2$ and $\beta_{\text{bear}}=0.4$.
This is much more interesting.
The strategy captures a large portion of the benchmark's upside while having considerably less downside sensitivity.
That is exactly the sort of behavior many active strategies are designed to achieve.

$## Timing Ratio

Timing ratio summarizes this asymmetry
$$\text{Timing Ratio}=\frac{\beta_{\text{bull}}}{\beta_{\text{bear}}}$$

For Strategy B above
$$\frac{1.2}{0.4}=3$$

That's an intuitively attractive result: the strategy has three times as much benchmark sensitivity in the favorable regime as in the unfavorable regime.

This is particularly relevant for strategies involving:
- trend following;
- tactical asset allocation;
- market timing;
- defensive overlays;
- volatility targeting;
- dynamic hedging;
- regime-switching strategies.

For a simple buy-and-hold equity strategy, you generally shouldn't expect spectacular bull/bear asymmetry. For a strategy explicitly designed to manage downside exposure, however, this becomes a very useful diagnostic.

### Where to use SFM in a backtesting framework

Think of the SFM family as answering
> what is my strategy actually exposed to?
rather than
> is my strategy good?

A useful hierarchy is

$$\text{CAGR},\quad\text{Sharpe},\quad\text{Sortino},\quad\text{Omega},\quad\text{Kappa},\ldots$$
These answer
> How good were the returns relative to risk?

$$\text{Max Drawdown},\quad\text{Calmar},\quad\text{Martin},\quad\text{Pain},\quad\text{CDaR},\ldots$$
These answer
> How painful was the path?

$$\alpha,\quad\beta,\quad R^2,\quad\beta_{\rm bull},\quad\beta_{\rm bear}$$
These answer:
> What generated those returns, and how dependent are they on the benchmark?
That makes SFM a natural complement to the extensive downside-risk metrics you're already implementing.

### One important warning for backtesting

DDo not rank strategies primarily by alpha.

Suppose you test 500 strategies and select the one with the highest historical SFM alpha. You have almost certainly introduced a multiple-testing/data-mining problem.

A much better use is
> This strategy has attractive downside characteristics, and SFM analysis shows that its performance is not simply explained by its benchmark exposure.

That's a much stronger piece of evidence than
> This strategy has the highest alpha.

Calculate SFM statistics *out of sample* whenever possible.
For example

```text
Training period
    ↓
develop/select strategy
    ↓
Validation period
    ↓
evaluate
    ↓
Out-of-sample SFM
    ├── alpha
    ├── beta
    ├── R²
    ├── bull beta
    └── bear beta
```

An alpha that survives out of sample is considerably more interesting than one observed only during strategy development.

### Overall assessment

I would classify the properties approximately like this:

| Property        | Backtesting usefulness                              |
| --------------- | --------------------------------------------------- |
| `sfm_beta`      | **High** — benchmark exposure                       |
| `sfm_alpha`     | **High, but easy to misuse** — residual performance |
| `sfm_r2`        | **High** — dependence on benchmark                  |
| `sfm_beta_bull` | **High for asymmetric/dynamic strategies**          |
| `sfm_beta_bear` | **High for downside analysis**                      |
| `timing_ratio`  | **High for timing/defensive strategies**            |

The most useful insight is that SFM doesn't replace your risk-adjusted performance ratios; it explains them.

A strategy with a high Sortino ratio is interesting. A strategy with a high Sortino ratio *and* low benchmark (R^2), modest beta, positive out-of-sample alpha, and substantially lower bear beta tells a much more informative story.

## Jensen's Alpha: Mathematics and Use in Trading

Jensen's Alpha is one of the classical measures of **risk-adjusted performance**. Its purpose is simple: it asks whether a strategy earned more or less return than would be expected given its exposure to a benchmark or market factor.

For a trading strategy, this makes it more informative than looking at raw return alone. A strategy that earns 20% while taking enormous market risk is not necessarily more impressive than one earning 12% with very little systematic exposure.

### 1. The mathematical idea

Jensen's Alpha originates from the Capital Asset Pricing Model (CAPM), or more generally the single-factor model (SFM):
$$R_{p,t}-R_f=\alpha+\beta(R_{b,t}-R_f)+\epsilon_t$$
where
$R_{p,t}$ is the portfolio or strategy return,
$R_{b,t}$ is the benchmark return,
$R_f$ is the risk-free rate,
$\beta$ measures sensitivity to the benchmark,
$\alpha$ is the component of return not explained by that benchmark exposure,
$\epsilon_t$ is the residual return.

The familiar Jensen's Alpha expression is
$$\alpha_J=R_p-\left[R_f+\beta(R_b-R_f)\right]$$
The term in brackets is the return that the strategy would theoretically be expected to earn given its beta.
Thus, $\alpha_J>0$ means the strategy outperformed the beta-implied return, while $\alpha_J<0$ means it underperformed.

For example, suppose: $R_p=15%,\qquad R_b=10%,\qquad R_f=3%,\qquad\beta=1.2$.
The expected return is $3%+1.2(10%-3%)=11.4%$.
Therefore, $\alpha_J=15%-11.4%=3.6%$
The strategy generated 3.6 percentage points more return than its benchmark exposure would suggest.

### 2. Jensen's Alpha versus SFM regression alpha

There is an important subtlety for your implementation.

The regression model is usually estimated using arithmetic periodic returns
$$R_{p,t}-R_f=\alpha+\beta(R_{b,t}-R_f)+\epsilon_t$$

Consequently, the regression intercept is
$$\alpha_{\mathrm{SFM}}=\overline{R_p-R_f}\beta\overline{R_b-R_f}$$

Jensen's Alpha, however, can be calculated from compounded portfolio and benchmark returns
$$\alpha_J=R_p-[R_f+\beta(R_b-R_f)]$$
If $R_p$ and $R_b$ are geometric returns, Jensen's Alpha and the regression intercept need not be identical.

### 3. Why geometric returns are interesting

For a trading strategy, the geometric return has an important advantage: it respects compounding.
Suppose a strategy produces $+50%,\quad -50%$.

The arithmetic mean is
$$\frac{50%-50%}{2}=0%$$
But the actual compounded wealth is
$$1.5\times0.5=0.75$$
The strategy lost $25%$.
Its geometric mean return is
$$\sqrt{1.5\times0.5}-1=\sqrt{0.75}-1\approx -13.40%$$.
For evaluating the actual growth of capital, the geometric measure is therefore much more meaningful.
For $n$ observations
$$R_{p,g}=\left(\prod_{t=1}^{n}(1+R_{p,t})\right)^{1/n}-1$$.
Annualization then gives
$$R_{p,\mathrm{ann}}=\left(\prod_{t=1}^{n}(1+R_{p,t})\right)^{m/n}-1$$
where $m$ is the number of observations per year.

### 4. What Jensen's Alpha actually tells a trader

The most useful interpretation is
> How much return did the strategy generate beyond what can be attributed to its systematic benchmark exposure?

This is particularly useful when comparing strategies with different betas.
Imagine two strategies:

| Strategy | Return | Beta | Jensen Alpha |
| -------- | -----: | ---: | -----------: |
| A        |    20% |  1.5 |           1% |
| B        |    14% |  0.5 |           5% |

Strategy A has the higher raw return, but much of that return can be explained by its large market exposure.

Strategy B has the lower absolute return but considerably more return relative to its systematic risk.

For a portfolio manager, this can be a much more interesting result.

### 5. Jensen's Alpha is especially useful with benchmark-relative strategies

It works particularly well when there is a meaningful economic benchmark, like

- an equity long/short strategy vs. an equity index,
- a market-neutral strategy vs. the broad market,
- a sector strategy vs. its sector index,
- a bond strategy vs. an appropriate bond benchmark,
- a factor strategy vs. the factor it is intended to exploit.

Suppose you have a momentum strategy on the S&P 500.
A raw annual return of 18% doesn't tell you much if the S&P 500 itself returned 17%.

If the strategy has $\beta=1.1$, then its expected return under the single-factor model might be even higher than 17%.

Jensen's Alpha tells you whether the strategy actually added value *after accounting for that exposure*.

### 6. The choice of benchmark is crucial

This is probably the biggest practical limitation of Jensen's Alpha.
Alpha is not an intrinsic property of a strategy.
It is a property of
$$\text{strategy + benchmark + risk model}$$
Change the benchmark and you can change the alpha.

Consider a technology-stock strategy.
Against the S&P 500 it might have $\beta=1.3,\qquad \alpha=2%$.
Against the Nasdaq 100 it might have $\beta=0.9,\qquad \alpha=-1%$.

Neither calculation is necessarily "wrong." They answer different questions.
Therefore, when using Jensen's Alpha in strategy research, the benchmark should be economically justified rather than selected because it produces a favorable alpha.

### 7. Alpha does not measure total risk

This is another important limitation. Jensen's Alpha only adjusts for the risk captured by the model's beta.

A strategy could have $\beta\approx0$ and therefore potentially show a large positive Jensen Alpha while still having substantial

- volatility,
- drawdowns,
- tail risk,
- skewness,
- liquidity risk,
- leverage,
- nonlinear exposure.

This is especially important for trading strategies involving options, stop-losses, leverage, or nonlinear payoffs.

A beta of zero does not mean zero risk.
It merely means that the strategy's linear covariance with the chosen benchmark is approximately zero.

### 8. Alpha can therefore be misleading for trading strategies

Suppose a strategy produces many small gains and occasional enormous losses.
It might have a respectable Jensen Alpha before the large loss occurs.

But Jensen's Alpha itself does not tell you about the shape of those losses.
This is why I would never use Jensen's Alpha as a standalone strategy-selection metric.

It should be examined alongside measures such as Sharpe, Sortino, Calmar, Martin/Ulcer, Omega, Kappa and the drawdown/episode statistics you've been implementing.
The combination answers much more interesting questions
> Did the strategy outperform?
> Was the return generated independently of the market?
> How much downside risk was taken?
> How severe were the drawdowns?
> How asymmetric were the returns?

### 9. Jensen's Alpha can be particularly valuable for evaluating "alpha strategies"

There is an interesting circularity in trading.
A strategy might be advertised as an "alpha strategy" because it generates returns independent of the market.
Jensen's Alpha provides a direct test of that claim.
Suppose $\beta\approx0$ and the strategy generates $R_p=12%$.
With a risk-free rate of 3% $\alpha_J=12%-3%-0(R_b-3%)=9%$
That is potentially very interesting.
The strategy produced substantial return without requiring significant exposure to the benchmark.
But you still need to investigate whether the result is statistically robust and whether other risk factors explain it.

### 10. The single-factor limitation is important

Modern quantitative strategies rarely have only one source of systematic risk.
A strategy might have exposure to market, size, value, momentum, quality, volatility, interest rates, credit, commodities, currencies, liquidity.
A strategy could therefore appear to have positive Jensen Alpha simply because the chosen benchmark doesn't capture one of its systematic exposures.

This motivates multifactor models
$$R_p-R_f=\alpha+\beta_1F_1+\beta_2F_2+\cdots+\beta_kF_k+\epsilon$$

The corresponding alpha is the return unexplained by all included factors.
For serious strategy research, multifactor alpha is often more informative than single-factor Jensen Alpha.

### 11. What about intraday strategies?

For an intraday strategy, an annualized Jensen Alpha can sometimes be awkward.
Suppose your observations are five-minute returns. You might have thousands of observations per year.

Annualizing
$$R_{\mathrm{ann}}=\left(\prod(1+r_t)\right)^{m/n}-1$$
can produce an impressive-looking annual number even when the strategy is intended to be evaluated primarily on an intraday horizon.

Here *per-observation-period Jensen Alpha* is therefore useful
$$\alpha_{J,\mathrm{period}}=G_p-[R_f+\beta(G_b-R_f)]$$

It lets you say
> For each five-minute observation period, how much geometric return did the strategy generate beyond its benchmark-implied return?

while the annualized version answers
> What is the corresponding annualized excess return?

Those are different reporting perspectives, and both are worthwhile in a quantitative library.

### 12. Jensen's Alpha is best viewed as an attribution measure

Perhaps the most useful way to think about it is not
> How profitable is my strategy?
but
> How much of my strategy's return is left after accounting for its benchmark exposure?

That makes Jensen's Alpha particularly useful for strategy attribution and comparison.
A high-return strategy with high beta may not have much alpha.
A lower-return market-neutral strategy may have substantial alpha.
And a strategy with excellent alpha but terrible drawdowns may still be unacceptable.

So, for a library, I would regard Jensen's Alpha as a complementary return-attribution metric, rather than a primary risk-adjusted performance metric.

### In practical strategy testing

A useful hierarchy is

- Return: How much did I make?
- Beta: How much market exposure did I take?
- Jensen Alpha: How much return remains after accounting for that exposure?
- Sharpe / Sortino / Omega / Kappa: How efficiently did I generate that return relative to risk
- Drawdown / Calmar / Martin / CDaR: How painful was the path to achieving it?

That combination is much more powerful than Jensen's Alpha alone.

## Fama Beta and Modigliani–Modigliani ($M^2$)

Fama Beta and Modigliani–Modigliani ($M^2$) are complementary benchmark-relative performance measures, but they answer different questions.

Fama Beta measures relative total risk
$$\beta_F=\frac{\sigma_P}{\sigma_B}$$
where both standard deviations use `ddof=0` and the portfolio and benchmark returns have the same periodicity. A value of 1 means the strategy has the same volatility as the benchmark; 1.5 means 50% more total volatility; 0.7 means 30% less. Unlike SFM beta, Fama Beta says nothing about *which part* of the portfolio's risk is systematic—it simply compares total dispersion.

In trading, this makes Fama Beta useful for identifying whether a strategy is achieving its returns by taking substantially more or less total risk than its benchmark.

$M^2$, on the other hand, converts risk-adjusted performance back into *return units*
$$M^2=R_f+\frac{\overline{R_p-R_f}}{\sigma_P}\sigma_B$$

It can be interpreted as the return the strategy would have produced if its volatility had been scaled to equal the benchmark's volatility. This makes M² much easier to interpret than the Sharpe ratio when comparing strategies: it answers, approximately,
> What return would this strategy have generated at the benchmark's level of risk?

For example, if a strategy has a high return but substantially higher volatility than the benchmark, $M^2$ can reveal that much of its apparent performance comes from taking additional risk. Conversely, a lower-return strategy with substantially lower volatility may have a surprisingly attractive M².

Together they provide a useful pair
$$\text{Fama Beta} \rightarrow \text{How much total risk did I take?}$$
$$\M^2 \rightarrow \text{What return did I generate at benchmark risk?}$$

For trading-strategy evaluation, we can use them alongside Sharpe/Sortino and drawdown measures, rather than as standalone metrics. Fama Beta describes the strategy's risk scale, while $M^2$ translates risk-adjusted performance into something directly comparable with ordinary returns.

### 1. Where these metrics belong

We can organize them roughly as:

| Metric         | Category                    | Main idea                                           |
| -------------- | --------------------------- | --------------------------------------------------- |
| `sfm_beta`     | Single-Factor Model / CAPM  | Systematic risk relative to benchmark               |
| `sfm_alpha`    | Single-Factor Model / CAPM  | Regression alpha                                    |
| `jensen_alpha` | Risk-adjusted return / CAPM | Return unexplained by beta                          |
| `fama_beta`    | Benchmark-relative risk     | Total risk relative to benchmark risk               |
| `modigliani`   | Risk-adjusted performance   | Portfolio return normalized to benchmark volatility |
| Sharpe         | Risk-adjusted performance   | Excess return per unit of total risk                |
| Treynor        | Risk-adjusted performance   | Excess return per unit of systematic risk           |

So I would not put Fama Beta under SFM/CAPM. It is specifically a *total-risk-relative-to-benchmark* measure.

And $M^2$ is essentially a *risk-adjusted performance measure closely related to Sharpe ratio*, not an SFM measure.

### Fama Beta

Fama Betc can be defined as
$$\beta_F=\frac{\sigma_{P,\mathrm{population}}\sqrt{f_P}}{\sigma_{B,\mathrm{population}}\sqrt{f_B}}$$
One can also define Fama Beta using annualized volatility
$$\beta_F=\frac{\sigma_{P,\mathrm{ann}}}{\sigma_{B,\mathrm{ann}}}$$
If portfolio and benchmark have the same number of observations, and the same frequency, then the factors cancel
$$\beta_F=\frac{\sigma_P}{\sigma_B}$$

The difference with SFM beta is fundamental.
$$\beta_{\mathrm{SFM}}=\frac{\operatorname{Cov}(R_p,R_b)}{\operatorname{Var}(R_b)}$$
measures *systematic exposure*, while
$$\beta_F=\frac{\sigma_p}{\sigma_b}$$
measures *relative total risk*.

For example, $\beta_{\mathrm{SFM}}=0.8$ but $\beta_F=1.4$.

That would tell you something interesting: the portfolio has relatively modest *linear market exposure*, but considerably more total volatility than the benchmark.

### Modigliani-Modigliani

Starting with the formula $M^2=SR_p\sigma_b+R_f$, since
$$SR_p=\frac{R_p-R_f}{\sigma_p}$$
we get
$$M^2=R_f+\frac{R_p-R_f}{\sigma_p}\sigma_b$$
Therefore
$$M^2=R_f+(R_p-R_f)\frac{\sigma_b}{\sigma_p}$$
which is exactly the formula we calculate it.

### There is a nice mathematical relationship between $M^2$ and Fama Beta

This is actually worth highlighting because these two metrics complement each other.
Start with
$$M^2=R_f+SR_p\sigma_b$$
and
$$SR_p=\frac{R_p-R_f}{\sigma_p}$$
Therefore
$$M^2=R_f+(R_p-R_f)\frac{\sigma_b}{\sigma_p}$$
but
$$\beta_F=\frac{\sigma_p}{\sigma_b}$$
Therefore
$$M^2=R_f+\frac{R_p-R_f}{\beta_F}$$
assuming the same return and volatility conventions.

This gives Fama Beta and $M^2$ a particularly intuitive relationship.

Fama Beta asks
> How much total risk did the portfolio take relative to the benchmark?
$M^2$ asks
> What would the portfolio's return have been if we scaled its risk to the benchmark's risk?

That's why $M^2$ is sometimes easier to interpret than Sharpe.
Sharpe is expressed as *return per unit of risk*, $M^2$ converts that back into *return units*.

### $M^2$ is basically a "risk-normalized return"

Suppose $R_p=15%,\quad R_f=3%,\quad\sigma_p=20%,\quad\sigma_b=15%$.
Then $SR_p=\frac{15-3}{20}=0.60$.
$M^2$ is $3+0.60(15)=12%$.
So the portfolio's *risk-adjusted equivalent return at benchmark volatility is 12%*.

The portfolio actually earned 15%, but it took more volatility than the benchmark $\beta_F=\frac{20}{15}=1.333$.
If we scaled the portfolio down to benchmark risk, its return would be 12%.
That's an extremely intuitive interpretation.

### Where to put everything in the library

Given the collection we're building, it can be organized approximately as follows

- Single-Factor Model / CAPM: sfm_beta, sfm_alpha, jensen_alpha, jensen_alpha_annualized
- Benchmark-relative risk: fama_beta
- Risk-adjusted performance: sharpe_ratio, sortino_ratio, treynor_ratio, modigliani

There is some unavoidable overlap here because these measures come from different historical schools of performance measurement.

We would not force them into mutually exclusive mathematical categories.
Instead, categorize according to their primary interpretation.

## Tracking error, active premium, information ratio, Treynor ratio

These eight properties form a fairly coherent benchmark-relative performance group, but they answer different questions.

### Tracking Error

Tracking Error measures how much the portfolio's returns fluctuate around the benchmark
$$TE=\sigma(R_P-R_B)$$

A low Tracking Error means the strategy closely follows the benchmark; a high value means its active returns are more variable. The annualized version is conventionally
$$TE_{ann}=TE\sqrt{m}$$
where $m$ is the number of observations per year.

For a trading strategy, Tracking Error is useful for distinguishing a strategy that is genuinely behaving differently from its benchmark from one that is essentially a benchmark clone.

### Active Premium

Active Premium measures the portfolio's return advantage over the benchmark.
The non-annualized version uses the difference between the portfolio and benchmark geometric mean returns, while the annualized version uses their separately annualized geometric returns, $AP=G_P-G_B$.

A positive Active Premium means the strategy has outperformed the benchmark; a negative value means underperformance.

### Information Ratio

The Information Ratio combines the previous two concepts
$$IR=\frac{AP}{TE}$$

It asks
> How much active return am I generating for each unit of active risk?

Thus, high Active Premium is not necessarily impressive if it comes with very high Tracking Error. A strategy producing 2% active return with 1% Tracking Error has an Information Ratio of 2, whereas one producing 5% active return with 10% Tracking Error has an Information Ratio of only 0.5.

The annualized version combines annualized Active Premium and annualized Tracking Error
$$IR_{ann}=\frac{AP_{ann}}{TE_{ann}}$$

This is particularly useful when evaluating active trading strategies against a benchmark because it focuses specifically on the return and risk that are attributable to deviating from that benchmark.

### Treynor Ratio

The Treynor Ratio takes a different approach to risk. Instead of Tracking Error, it uses *systematic risk (beta)*:
$$TR=\frac{R_P-R_f}{\beta_P}$$

In this implementation, the numerator is based on the geometric mean of portfolio excess returns.
The question is therefore
> How much excess return did the strategy generate for each unit of systematic risk?

This makes Treynor useful when the strategy is part of a diversified portfolio where idiosyncratic risk can be diversified away. For a standalone trading strategy, however, Sharpe or Sortino is usually more informative because total risk matters.

### How they fit together

You can think of the four measures as two different views of active performance:

| Measure | Return | Risk | Main question |
| --- | --- | --- | --- |
| Active Premium | Portfolio vs benchmark | — | Did I beat the benchmark? |
| Tracking Error | — | Active risk | How differently did I behave?  |
| Information Ratio | Active Premium | Tracking Error | How efficiently did I generate active return? |
| Treynor Ratio | Excess return | Beta | How efficiently did I generate return per unit of systematic risk? |

And the annualized versions simply put the relevant quantities onto an annual scale.

For trading, I would therefore look at them together.

Active Premium tells you whether you added value, Tracking Error tells you how much active risk you took, and Information Ratio tells you how efficiently you converted that active risk into excess performance. Treynor Ratio provides a complementary systematic-risk perspective.

For an intraday strategy, the non-annualized measures describe the strategy at its native sampling frequency, while the annualized versions allow comparison with conventional portfolio-performance statistics.

## Information Ratio versus Modified Information Ratio

The ordinary Information Ratio is
$$IR=\frac{E[R_P-R_B]}{\sigma(R_P-R_B)}$$

or, in another terminology,
$$IR=\frac{\text{active premium}}{\text{tracking error}}$$

It is therefore a very natural measure for a strategy whose objective is to outperform a benchmark.

The problem addressed by the Israelson modification is a slightly counterintuitive property of the ordinary IR.

Suppose a strategy has negative active return $E[R_P-R_B]<0$. Then the Information Ratio is negative $IR<0$.

Now increase the tracking error while keeping the negative active return constant: $IR=\frac{-2\%}{4\%}=-0.50$. If tracking error increases to 8%, $IR=\frac{-2\%}{8\%}=-0.25$.

Numerically, the IR has improved from -0.50 to -0.25, even though the strategy has not generated any additional active return. This is the undesirable property Israelson's modification attempts to address.

The implementation therefore applies
$$MIR =\begin{cases}IR,& AP>0\\ -IR,& AP\leq0\end{cases}$$

Consequently, $-0.50\rightarrow +0.50$ rather than allowing increasing tracking error to make a losing strategy appear better.

## Treynor Ratio versus Modified Treynor Ratio

An ordinary Treynor ratio is conceptually
$$TR=\frac{R_P-R_F}{\beta}$$

It asks
> How much excess return did the strategy generate per unit of systematic exposure?

This is a systematic-risk measure. Unlike Sharpe, it does not penalize specific/idiosyncratic risk in the denominator.
The modified version instead uses
$$MTR=\frac{R_{P,\text{excess,geom}}}{\text{systematic risk}}$$

With the definition of systematic risk
$$SR_{\text{sys}}=\beta\sigma_B\sqrt{P}$$

where $P$ is periods per annum. Therefore,
$$MTR=\frac{R_{P,\text{excess,geom}}}{\beta\sigma_B\sqrt{P}}$$

This is quite different from conventional Treynor.

### The important conceptual difference

Conventional Treynor has
$$\frac{\text{return}}{\beta}$$

Modified Treynor has approximately
$$\frac{\text{return}}{\beta\sigma_B}$$

So conventional Treynor measures return per unit of beta, whereas modified version measures return per unit of systematic volatility.

That makes `treynor_ratio_modified` conceptually closer to
> return generated per unit of systematic risk expressed as standard deviation.

This is useful because beta by itself is dimensionless, while systematic risk is expressed in percentage-return units.

## Systematic, specific and total risk

I would group these three together conceptually

```text
systematic_risk
specific_risk
total_risk
```

They are the risk decomposition associated with the single-factor model. Start with
$$R_P=\alpha+\beta R_B+\epsilon$$

Here $\alpha$ is intercept, $\beta R_B$ is systematic component, $\epsilon$ is specific/residual component. The model decomposes the strategy's return into
$$\text{Return}=\text{alpha}+\text{systematic component}+\text{specific component}$$

and, under the usual regression assumptions, the variance decomposes approximately as
$$\sigma_P^2=\sigma_{\text{systematic}}^2+\sigma_{\text{specific}}^2$$

because the residual is orthogonal to the factor. Thus.
$$\sigma_{\text{total}}=\sqrt{\sigma_{\text{systematic}}^2+\sigma_{\text{specific}}^2}$$

which is exactly the relationship implemented by `total_risk`.

That makes these three properties much more than three unrelated statistics.

### Systematic risk

Systematic risk represents the portion of the strategy's volatility attributable to its exposure to the benchmark/factor.
With one benchmark,
$$\sigma_{\text{systematic}}=|\beta|\sigma_B$$

and annualized
$$\sigma_{\text{systematic,ann}}=|\beta|\sigma_{B,\text{ann}}$$

Conceptually
> How much of the strategy's volatility can be explained by its exposure to the benchmark?

For example, imagine $\beta=0.8$ and benchmark volatility is $15\%$.
Then $\sigma_{\text{systematic}}=0.8\times15\%=12\%$.

The strategy therefore has approximately 12% annualized systematic risk.

### Specific risk

Specific risk is the volatility of the regression residual
$$\epsilon_i=R_{P,i}-(\alpha+\beta R_{B,i})$$

Therefore
$$\sigma_{\text{specific}}=\sigma(\epsilon)$$

This is sometimes called

- idiosyncratic risk,
- residual risk,
- unsystematic risk,
- specific risk.

The basic question is
> How much of the strategy's return variability remains unexplained after accounting for benchmark exposure?

This can be particularly interesting for trading strategies.
A strategy might have $\beta=1$ and therefore substantial systematic risk, but very little residual risk.
Another strategy might have $\beta=0.2$ yet have huge specific risk.

The two strategies could have similar total volatility while having completely different sources of risk.

### Total risk

Total risk is simply the combination
$$\sigma_T=\sqrt{\sigma_S^2+\sigma_{\epsilon}^2}$$

where $\sigma_S=\text{systematic risk}$ and $\sigma_{\epsilon}=\text{specific risk}$.

This gives a useful three-level hierarchy

```text
                     Total Risk
                    /           \
           Systematic Risk    Specific Risk
                 |
          Benchmark exposure
```

This is a very useful conceptual complement to existing `total_risk`/`standard_deviation` metric.

### Why this matters for trading strategies

The decomposition becomes particularly valuable when analyzing
> what kind of strategy you actually built.

Suppose you have two strategies:

- Strategy A $\sigma_T=15\%$ with $\sigma_S=14\%,\qquad\sigma_{\epsilon}=5\%$. Most of its risk comes from benchmark exposure.
- Strategy B $\sigma_T=15\%$ with $\sigma_S=5\%,\qquad\sigma_{\epsilon}=14\%$.

The strategies have approximately the same total risk, but they are fundamentally different.

- Strategy A is essentially a benchmark-exposure strategy.
- Strategy B has much more idiosyncratic/strategy-specific risk.

That distinction is extremely important for portfolio construction.

### It also tells you whether alpha is really independent

Suppose a strategy has a spectacular raw return.
You might initially conclude
> This strategy generates alpha.

But regression tells you whether the return is actually associated with benchmark exposure.

For example, $R_P=\alpha+\beta R_B+\epsilon$.
If $\beta=1.2$ and $\alpha\approx0$, then much of the strategy's return may simply be compensation for taking systematic market exposure.

Conversely, a strategy with $\beta\approx0$ and positive alpha is much more interesting as a potential source of diversifying return.

This is one reason the `sfm_alpha`, `sfm_beta`, `systematic_risk`, and `specific_risk` properties work nicely together.

### Particularly useful for market-neutral strategies

This becomes even more interesting for market-neutral trading.

Imagine a market-neutral strategy with $\beta\approx0$.
Its systematic risk is correspondingly small, $\sigma_{\text{systematic}}\approx0.$
Yet it might have $\sigma_{\text{specific}}>10\%$.

That tells you that the strategy is not actually low risk. It is simply taking little market risk.
Its risk is coming from the strategy itself:

- model errors,
- security selection,
- factor exposures not included in the benchmark,
- leverage,
- execution,
- nonlinear effects,
- regime changes.

So systematic risk should never be interpreted as the risk of the strategy. It is only one component of total risk.

### Relationship with other metrics

Current collection is actually becoming quite coherent. I'd think of it as several interconnected families:

Active-performance family

- active_premium
- tracking_error
- information_ratio
- information_ratio_modified

These answer
> How well does the strategy outperform its benchmark relative to active risk?

Beta/systematic-performance family

- sfm_alpha
- sfm_beta
- treynor_ratio
- treynor_ratio_modified
- systematic_risk

These answer
> How much return am I getting for my systematic benchmark exposure?

Risk decomposition family

- systematic_risk
- specific_risk
- total_risk

These answer
> Where does the strategy's volatility come from?

## Appraisal Ratio and Modified/Alternative Jensen alpha

Conventional Jensen's alpha
$$\alpha_J=R_P-\left[R_F+\beta(R_B-R_F)\right]$$

It answers
> How much return did the strategy generate above the return predicted by its systematic market exposure?

That is already an important trading-strategy statistic.
These three measures then take this alpha and normalize it by different forms of risk/exposure.

Jensen's alpha (units: return)
> How much abnormal return did the strategy generate?
Units: **return**.

Modified Jensen's alpha $\frac{\alpha_J}{\beta}$ (units: return per beta)
> How much alpha did the strategy generate per unit of beta?

This is analogous to Treynor's use of beta as a denominator, but the numerator is Jensen's alpha rather than total excess return.

Alternative Jensen's alpha $\frac{\alpha_J}{\sigma_{\text{systematic}}}$ (units: effectively a dimensionless ratio if both numerator and denominator are returns)
> How much alpha did the strategy generate per unit of systematic volatility?

This is closer to a Sharpe-like normalization, except that only systematic risk is used.

Appraisal Ratio $\frac{\alpha_J}{\sigma_{\text{specific}}}$
> How much alpha did the strategy generate per unit of residual/idiosyncratic risk?

This is arguably the most interesting of the three for active trading.

### The risk decomposition makes the relationships clearer

The preceding group
$$\sigma_T^2=\sigma_S^2+\sigma_E^2$$

where $\sigma_T$ is total risk, $\sigma_S$ is systematic risk, $\sigma_E$ is specific risk, gives us the natural context.

You can therefore think of the metrics as asking

| Metric                   | Alpha divided by | Main question                      |
| ------------------------ | ---------------- | ---------------------------------- |
| Jensen's alpha           | nothing          | How much alpha?                    |
| Modified Jensen alpha    | \(\beta\)        | Alpha per unit of beta?            |
| Alternative Jensen alpha | systematic risk  | Alpha per unit of systematic risk? |
| Appraisal ratio          | specific risk    | Alpha per unit of specific risk?   |

### Why the Appraisal Ratio is particularly interesting

Suppose two strategies both produce $\alpha_J=3\%$, but

- Strategy A $\sigma_{\text{specific}}=4\%$ so $AR=\frac{3\%}{4\%}=0.75$.
- Strategy B $\sigma_{\text{specific}}=12\%$ so $AR=\frac{3\%}{12\%}=0.25$.

Both generated exactly the same Jensen alpha, but Strategy A generated that alpha with substantially less residual risk.

For an active manager or trading strategy, that is valuable information.

The Appraisal Ratio therefore has a particularly natural interpretation
> How efficiently does the strategy convert its idiosyncratic risk-taking into benchmark-adjusted alpha?

### Why this is useful in trading

Imagine a strategy has $\beta=1.0,\qquad\alpha=4\%,\qquad\sigma_{\text{specific}}=5\%$. Then $AR=0.8$.
That is attractive: the strategy is generating substantial alpha relative to the risk that remains unexplained by the market.

Now imagine another strategy $\beta=0.1,\qquad\alpha=4\%,\qquad\sigma_{\text{specific}}=15\%.$
The second strategy might sound impressive because it has low market exposure and the same alpha, but $AR=0.267$.

The alpha is much less efficient relative to the strategy-specific risk.
This distinction can be very useful when evaluating market-neutral or low-beta strategies.

### Modified Jensen alpha and low-beta strategies

The $\frac{\alpha}{\beta}$ can be useful for another reason.
Suppose $\alpha=3\%$.

- For Strategy A $\beta=1.0$ giving $3\%$.
- For Strategy B $\beta=0.25$ giving $12\%$.

The modified measure highlights that Strategy B generated substantial alpha relative to its systematic exposure.

However, this is also where I would be cautious.
A very small beta can make $\frac{\alpha}{\beta}$ explode.

If $\beta\rightarrow0$, the ratio becomes unstable or undefined.

Consequently, I would not use `jensen_alpha_modified` as a primary strategy-ranking statistic. It is better treated as a diagnostic.

### The same caution applies to Alternative Jensen alpha

The alternative measure
$$\frac{\alpha}{\sigma_{\text{systematic}}}$$

has essentially the same problem.
If systematic risk becomes very small, the ratio can become enormous.
That is not necessarily evidence of an extraordinary strategy.

It may simply mean
> The denominator is approaching zero.

This is particularly relevant for market-neutral strategies.

For example, $\beta\approx0$ can produce $\sigma_{\text{systematic}}\approx0$.
A small amount of alpha can then generate an enormous Alternative Jensen Alpha.
So I would regard this metric as a conditional diagnostic, rather than a standalone ranking measure.

### Appraisal Ratio has a different failure mode

The Appraisal Ratio can also become unstable when $\sigma_{\text{specific}}\rightarrow0$.
But this situation has a different interpretation.
If residual risk is almost zero, the strategy's returns are almost completely explained by the benchmark exposure plus alpha.
That can actually be meaningful, but the ratio mathematically becomes very large.
Again, the convention of returning `math.nan` when the denominator is zero is sensible.

### Relationship to Sharpe, Treynor and Information Ratio

Your growing collection can now be seen as a set of different return per risk questions.

| | |
| --- | --- | ---|
| Sharpe | $$\frac{R_P-R_F}{\sigma_T}$$ | Return per total risk. |
| Sortino | $$\frac{R_P-MAR}{\sigma_{\text{downside}}}$$ | Return per downside risk. |
| Treynor | $$\frac{R_P-R_F}{\beta}$$ |  Return per unit of systematic exposure. |
| Information Ratio | $$\frac{R_P-R_B}{TE}$$ | Active return per active risk. |
| Appraisal Ratio | $$\frac{\alpha}{\sigma_{\text{specific}}}$$ | Alpha per specific risk. |
| Alternative Jensen alpha | $$\frac{\alpha}{\sigma_{\text{systematic}}}$$ | Alpha per systematic risk. |
| Modified Jensen alpha | $$\frac{\alpha}{\beta}$$ | Alpha per unit of beta. |

That is a very useful conceptual map for your library.

### How I would use them in a trading-strategy report

I wouldn't use all three Jensen-alpha variants as equal headline metrics.
I'd make the hierarchy something like:

Primary: Jensen alpha
> Did the strategy produce abnormal return after accounting for benchmark exposure?

Primary/secondary: Appraisal Ratio
> How efficiently did it produce that alpha relative to residual risk?

Diagnostic: Modified Jensen alpha
> How much alpha was generated relative to beta?

Diagnostic: Alternative Jensen alpha
> How much alpha was generated relative to systematic volatility?

And then pair those with

- total_risk
- systematic_risk
- specific_risk

That lets you tell a much more complete story.
For example
> The strategy generated 4.2% Jensen alpha with a beta of 0.35. Its total annualized risk was 11%, of which only 3% was systematic and 10.6% was specific. The Appraisal Ratio was 0.40, indicating that the alpha came with substantial strategy-specific risk.

That is considerably more informative than simply
> Sharpe = 1.1.

### One architectural suggestion

I would keep these properties together in your documentation as a Jensen-alpha / alpha-efficiency family:

- jensen_alpha
- jensen_alpha_modified
- jensen_alpha_alternative
- appraisal_ratio

and keep the following as the risk decomposition family:

- systematic_risk
- specific_risk
- total_risk

The two families then connect through the denominators
$$\text{Jensen Alpha}\rightarrow\begin{cases}\alpha/\beta\\\alpha/\sigma_{\text{systematic}}\\\alpha/\sigma_{\text{specific}}\end{cases}$$

That is a clean API design because each derived metric has a very obvious mathematical relationship to the base `jensen_alpha`.

## Jensen alpha hierarchy

We effectively have
$$\sigma_{\rm systematic}=|\beta|\,\sigma(R_b-R_f)$$

$$\sigma_{\rm specific}=\sigma\left[(R_p-R_f)-\alpha-\beta(R_b-R_f)\right]$$

and

$$\sigma_{\rm total}=\sqrt{\sigma_{\rm systematic}^2+\sigma_{\rm specific}^2}$$

with annualization applied consistently.

That gives ya very clean interpretation of the SFM:

$$\underbrace{R_p-R_f}_{\text{strategy excess return}}=\underbrace{\alpha}_{\text{abnormal return}}+\underbrace{\beta(R_b-R_f)}_{\text{systematic component}}+\underbrace{\epsilon}_{\text{specific component}}$$.

And the corresponding risk decomposition is
$$\underbrace{\sigma_{\rm total}^2}_{\text{overall variability}}=\underbrace{\sigma_{\rm systematic}^2}_{\text{benchmark-related}}+\underbrace{\sigma_{\rm specific}^2}_{\text{strategy-specific}}$$

That is a particularly nice foundation for the other metrics, for example
$$\text{Appraisal Ratio}=\frac{\alpha}{\sigma_{\rm specific}}$$

and

$$\text{Alternative Jensen Alpha}=\frac{\alpha}{\sigma_{\rm systematic}}.$$

So the denominators now genuinely correspond to the two components of the SFM risk decomposition.

## One more important point about Jensen alpha

This also tells us how `jensen_alpha`, `sfm_alpha`, and `sfm_beta` need to relate.

The SFM regression is
$$R_p-R_f=\alpha+\beta(R_b-R_f)+\epsilon$$

The fitted alpha is therefore
$$\alpha=E[R_p-R_f]-\beta E[R_b-R_f]$$

That is exactly the conventional CAPM/Jensen-alpha formulation
$$\alpha_J=R_p-R_f-\beta(R_b-R_f)$$

at the appropriate expectation/mean level.

So if `sfm_alpha` is the intercept from this regression, it is essentially the same economic quantity as Jensen's alpha under this single-factor formulation.

That means the family becomes particularly elegant

```text
SFM model

Rp - Rf
   │
   ├── alpha ────────────────> Jensen alpha
   │                             │
   │                             ├── alpha / beta
   │                             ├── alpha / systematic risk
   │                             └── alpha / specific risk
   │
   ├── beta ─────────────────> systematic risk
   │
   └── epsilon ──────────────> specific risk
                                 │
systematic risk + specific risk ─┴─> total risk
```

And that is a strong reason to keep these properties together in your documentation.

## M-squared, M-squared Excess, and M-squared Sortino

M-squared, usually written $M^2$, is a particularly useful performance measure when the goal is to turn a risk-adjusted ratio into a return-like quantity.
The measure was developed to make risk-adjusted performance easier to interpret and compare with ordinary investment returns.
Rather than saying that one strategy has a Sharpe ratio of 1.2 and another has a Sharpe ratio of 0.9, M-squared asks a more intuitive question:
> What return would the strategy have produced if it had been scaled to the risk level of the benchmark?

This makes M-squared especially useful when evaluating trading strategies with materially different volatility.
A low-volatility strategy can have an attractive Sharpe ratio but a relatively modest raw return, while a high-volatility strategy can have a larger raw return but a poorer risk-adjusted performance.
M-squared translates those differences back into the familiar unit of return.

`PerformanceAnalytics` describes M-squared as
> a risk-adjusted return that allows portfolios with different levels of risk to be compared.

### Ordinary M-squared

The standard M-squared measure starts with the Sharpe ratio.
Let $R_P$ be an annualized portfolio/strategy return, $R_F$ be an annualized risk-free rate, $R_M$ be an annualized benchmark return, $\sigma_P$ be a portfolio standard deviation, $\sigma_M$ be a benchmark standard deviation.

The Sharpe ratio is
$$\SR_P =\frac{R_P-R_F}{\sigma_P}$$

M-squared then adjusts the portfolio return to the benchmark's volatility
$$M^2 = R_P + SR_P(\sigma_M-\sigma_P)$$

which can be rearranged to
$$M^2 = R_F+(R_P-R_F)\frac{\sigma_M}{\sigma_P}$$

The second form makes the interpretation particularly clear.
The strategy's excess return over the risk-free rate is multiplied by
$$\frac{\sigma_M}{\sigma_P}$$

In other words, we imagine scaling the strategy's risky exposure until its volatility equals that of the benchmark.

#### Example

Suppose a strategy has $R_P=12\%, \qquad R_F=2\%, \qquad \sigma_P=10\%$ and the benchmark has $\sigma_M=15\%$.
The strategy's Sharpe ratio is
$$SR_P=\frac{12\%-2\%}{10\%}=1.0$$

Its M-squared is therefore
$$M^2=12\%+1.0(15\%-10\%)=17\%$$

The interpretation is
> If the strategy had been scaled to have the benchmark's 15% volatility, its risk-adjusted return would have been approximately 17%.

That is much easier to interpret than simply saying "Sharpe = 1.0."

### M-squared as a volatility-normalized strategy return

There is an important conceptual distinction between Sharpe ratio and M-squared.
The Sharpe ratio is dimensionless:
$$SR=\frac{\text{return}}{\text{risk}}$$

M-squared has units of return, $M^2=\text{return}$,
Consequently, M-squared can be compared directly with
- the strategy's annual return,
- the benchmark's annual return,
- another strategy's annual return,
- a target return,
- a required rate of return.

This is its main practical advantage.

For trading-strategy research, this can be useful when strategies have deliberately different volatility targets. For example

| Strategy  | Return | Volatility | Sharpe | M² at benchmark risk |
| --------- | -----: | ---------: | -----: | -------------------: |
| A         |     8% |         5% |   1.20 |                  14% |
| B         |    14% |        12% |   1.00 |                  12% |
| Benchmark |    10% |        10% |      — |                  10% |

The raw returns make Strategy B look better than Strategy A. The Sharpe ratio makes Strategy A look better. M-squared puts both on the same volatility scale and therefore gives a return-like comparison.

This is particularly useful when comparing systematic trading strategies, because position sizing, leverage and volatility targeting can otherwise make raw returns difficult to compare.

### The benchmark is essential

M-squared is not an absolute measure. It is inherently relative to a benchmark risk level. The benchmark does not necessarily have to be "the market" in an economic sense. It is simply the return series whose volatility defines the target risk level.

For example, you could use

- S&P 500 for equity strategies,
- a bond index for fixed-income strategies,
- BTC for a crypto strategy,
- a portfolio benchmark,
- or another strategy whose risk level you regard as appropriate.

This is why your Python property naturally belongs conceptually with your benchmark-dependent metrics.

M-squared answers
> What would this strategy's return look like if I put it on the benchmark's volatility scale?

It does not answer
> How much did this strategy beat the benchmark?

That latter question is better addressed by active premium, information ratio, etc.

### $M^2$ Excess

M-squared itself produces a risk-adjusted return. M-squared Excess goes one step further and asks:
> How much return does the risk-adjusted strategy produce relative to the benchmark?

The geometric M-squared excess wwhich we implement here is
$$M^2_{\text{excess,geom}}=\frac{1+M^2}{1+R_B}-1$$

This is subtly different from simply subtracting returns. For example, $M^2=12\%, \qquad R_B=10\%$. Arithmetic excess $12\%-10\%=2\%$, geometric excess $\frac{1.12}{1.10}-1=1.818\%$.

The geometric definition treats the comparison as a relative compounded return, whereas the arithmetic definition treats it as a simple percentage-point difference.

#### Why M-squared Excess is useful for trading strategies

M-squared is useful when the primary question is
> How good is this strategy at a given volatility?

M-squared Excess changes the question to
> How good is this strategy at a given volatility relative to the benchmark's return?

Imagine two strategies.

| | Strategy A | Strategy B |
| --- | --- | --- |
| Return | 9% | 14% |
| Volatility | 6% | 15% |
| Sharpe | 1.17 | 0.80 |
| $M^2$ at benchmark volatility | 15% | 10% |

If the benchmark returns 11%, Strategy A has $M^2_{\text{excess,arith}}=15\%-11\%=4\%$, Strategy B hasm $M^2_{\text{excess,arith}}=10\%-11\%=-1\%$.

So despite Strategy B having the larger raw return, Strategy A has substantially better risk-adjusted performance.

This is exactly the kind of distinction that matters in systematic trading.

### $M^2$ Sortino

Ordinary M-squared uses total volatility, $\sigma_P$. That means upside and downside variation are treated identically. For many trading strategies, that is undesirable.

A strategy can have substantial upside volatility because it occasionally generates very large positive returns. Standard deviation treats those positive observations as risk even though the investor may actually welcome them.

The Sortino framework instead focuses on returns below a Minimum Acceptable Return (MAR).

The downside deviation is conceptually based on the lower partial moment
$$LPM_2(MAR)=\frac{1}{n}\sum_{i=1}^{n}\min(R_i-MAR,0)^2$$

The downside deviation is
$$D(MAR)=\sqrt{LPM_2(MAR)}$$

The Sortino ratio is then
$$Sortino=\frac{R_P-MAR}{D_P}$$

We can say the Sortino ratio uses excess return over MAR divided by downside deviation.
M-squared Sortino takes exactly the same idea behind ordinary M-squared and replaces total volatility with downside risk.
$$M_S^2=R_P+Sortino_P\left(D_M-D_P\right)$$

where $R_P$ is annualized portfolio return, $D_P$ is portfolio annualized downside risk, $D_M$ is benchmark annualized downside risk, $Sortino_P$ is portfolio Sortino ratio.

This is the direct downside-risk analogue of ordinary M-squared.

#### The important interpretation of $M^2$ Sortino

Suppose $R_P=12\%$ and the strategy has downside deviation $D_P=8\%$. Suppose the benchmark has $D_M=12\%$. The strategy's Sortino ratio is $Sortino=\frac{12\%-MAR}{8\%}$.

If the resulting Sortino ratio is 1.25, then $M_S^2=12\%+1.25(12\%-8\%)=17\%$.

The interpretation is analogous to ordinary M-squared
> $M^2$ Sortino estimates the return the strategy would produce if its downside risk were scaled to the benchmark's downside risk.

The crucial difference is that we are no longer matching total volatility. We are matching downside risk relative to MAR.

That makes M² Sortino particularly interesting for strategies with asymmetric return distributions.

#### Why $M^2$ Sortino can be better for trading strategies

Consider two strategies with identical annual returns:

- Strategy A produces many moderate positive returns and occasional large losses.
- Strategy B produces volatile returns, including many very large positive observations, but relatively few damaging losses.

Both may have similar standard deviations. Standard M-squared will therefore treat them similarly.

But their downside deviations may be very different. $M^2$ Sortino can recognize that difference.

This is important for

- trend-following strategies,
- option strategies,
- asymmetric strategies,
- long-volatility strategies,
- momentum strategies,
- strategies with positively skewed returns,
- strategies with substantial upside variability.

In these cases, standard deviation may penalize desirable upside variation, whereas downside deviation focuses specifically on failure to achieve the investor's target.

That is the same philosophical distinction underlying Sortino versus Sharpe.

#### MAR makes $M^2$ Sortino investor-dependent

There is an important consequence of using MAR. Ordinary $M^2$ primarily depends on the risk-free rate used in the Sharpe ratio.

$M^2$ Sortino depends on the Minimum Acceptable Return.

Suppose a strategy has $R_P=15\%$. If MAR is 0%, relatively few observations may qualify as downside observations.
If MAR is 8%, substantially more observations fall below the target.

Consequently, $D_P(0\%) \neq D_P(8\%)$ and therefore $M_S^2(0\%) \neq M_S^2(8\%)$. This isn't a defect. It is the purpose of the measure.

A strategy can look excellent when the objective is merely "don't lose money," but considerably less attractive when the objective is "achieve at least 8% annually."

Choosing MAR carefully is important because it defines what constitutes unacceptable performance.

### Relationship between the three measures

The three measures can be viewed as a small hierarchy
$$\boxed{\text{Sharpe}}\quad\longrightarrow\quad\boxed{M^2}$$

and
$$\boxed{\text{Sortino}}\quad\longrightarrow\quad\boxed{M_S^2}$$

while M-squared Excess adds a benchmark-return comparison
$$\boxed{M^2}\quad\longrightarrow\quad\boxed{M^2_{\text{Excess}}}$$

Conceptually:

| Measure    | Performance basis | Risk basis         | Output        |
| ---------- | ----------------- | ------------------ | ------------- |
| Sharpe     | excess return     | total volatility   | ratio         |
| $M^2$      | excess return     | total volatility   | return        |
| $M^2$ Excess  | M²             | benchmark return   | excess return |
| Sortino    | return over MAR   | downside deviation | ratio         |
| $M^2$ Sortino | return         | downside deviation | return        |

This distinction is important
> Sharpe and Sortino are ratios. M-squared measures are not.

### $M^2$ versus Information Ratio

There is also an interesting relationship with the Information Ratio. The Information Ratio is approximately
$$IR=\frac{R_P-R_B}{TE}$$

where $TE$ is tracking error.

It therefore asks:
> How much active return am I receiving per unit of active risk?

$M^2$ asks something different
> What would my return be if my total risk were normalized to the benchmark's risk?

So these measures answer different questions:

$M^2$
> How good is the strategy after putting everyone on the same volatility scale?

$M^2$ Excess
> How much does that risk-normalized return exceed the benchmark?

Information Ratio
> How efficiently does the strategy generate active return relative to its tracking error?

For a family-office or institutional strategy analysis, it can therefore be useful to report both rather than trying to choose one universal measure.

### Practical use in strategy evaluation

For a trading strategy, the three measures may be used as follows.

#### $M^2$

Use it when comparing strategies with different volatility targets.

For example, compare:

- a 5% volatility strategy,
- a 10% volatility strategy,
- a 15% volatility strategy.

$M^2$ puts them on the same benchmark-risk scale.

#### $M^2$ Excess

Use it when the benchmark's return matters as well as its risk.

It answers whether the risk-normalized strategy actually delivered something better than simply owning the benchmark.

#### $M^2$ Sortino

Use it when downside risk is more meaningful than total volatility.

This is often the most interesting version for strategies with asymmetric distributions or a clearly defined investment objective.

### An important limitation

$M^2$ the weaknesses of the underlying Sharpe or Sortino framework.

Ordinary $M^2$ assumes that standard deviation is a meaningful description of risk. That can be questionable for

- fat-tailed returns,
- highly skewed strategies,
- options,
- nonlinear strategies,
- strategies with regime changes.

$M^2$ Sortino improves the treatment of downside risk, but it does not eliminate distributional problems. Downside deviation is still a second-order measure and can therefore miss some characteristics of extreme-tail behavior.

This is why $M^2$/$M^2$ Sortino should be viewed alongside your other measures such as:

- maximum drawdown,
- Ulcer Index,
- Pain Index,
- CDaR,
- Kappa,
- Omega,
- skewness and kurtosis,
- Calmar/Martin ratios.

A strategy that has a spectacular M² but unacceptable drawdown characteristics should not automatically be considered superior.

## Tail Ratio

It adds something that your Sharpe/Sortino/Kappa/Omega family does **not** directly measure: the **relative magnitude of the empirical upper and lower tails**.

It is especially useful for trading strategies with asymmetric return distributions.

I would put it in a **distribution / tail-risk family**, rather than in the conventional risk-adjusted-return family.

Your broader organization could look like:

```text
Performance measures
│
├── Return
│   ├── arithmetic mean
│   ├── geometric mean
│   └── ...
│
├── Volatility / total risk
│   ├── standard deviation
│   ├── Sharpe
│   └── M²
│
├── Downside / lower-partial-moment
│   ├── downside deviation
│   ├── Sortino
│   ├── Kappa
│   ├── Omega
│   └── Prospect Ratio
│
├── Drawdown
│   ├── maximum drawdown
│   ├── Calmar
│   ├── Martin
│   ├── Pain
│   └── Sterling
│
├── Benchmark / factor
│   ├── beta
│   ├── Jensen alpha
│   ├── Information Ratio
│   ├── Treynor
│   └── Appraisal Ratio
│
└── Distribution / tail
    ├── skewness
    ├── kurtosis
    ├── Tail Ratio
    └── ...
```

I'd call this family **Distribution and Tail Measures**.

That is important because `tail_ratio` isn't really a "ratio of return to risk." It is a **shape statistic**.

The definition is:
$$TR_q=\frac{Q_q(R)}{|Q_{1-q}(R)|}$$

with $q=0.95$ by default. Therefore
$$TR_{0.95}=\frac{P_{95}(R)}{|P_5(R)|}$$

It compares a high positive percentile with the corresponding low negative percentile.
For example, $P_{95}=+2\%$ and $P_5=-1\%$ gives $TR=2$.

The interpretation is
> The 95th-percentile gain is approximately twice the magnitude of the 5th-percentile loss.

Conversely, $P_{95}=1\%,\qquad P_5=-2\%$ gives $TR=0.5$.

Now the lower tail is twice as large as the upper tail.

### The important interpretation of 1

The natural reference point is $TR=1$ because that means $|P_{95}|=|P_5|$. So

- $> 1$: upper tail is larger
- $= 1$: approximately symmetric tails at those quantiles
- $< 1$: lower tail is larger

This makes it quite intuitive.

### Why is it useful for trading?

Imagine two strategies with identical Sharpe ratios:

- Strategy A: many modest gains, occasional modest losses
- Strategy B: many modest gains, occasional very large losses

Their volatility-adjusted performance could be similar, but their tail behavior can be dramatically different.
Tail Ratio exposes that difference.

For a strategy with $TR=0.45$, the lower tail is more than twice as large as the upper tail at the selected percentiles.
That is a warning sign for a strategy whose risk is concentrated in adverse tail events.

Conversely, a strategy with $TR=2.0$ has substantially more upside-tail magnitude than downside-tail magnitude.

That can be desirable, although it should **not** automatically be interpreted as "low risk."

### It complements skewness rather than replacing it

This is probably the most important relationship with your existing distribution measures.

You already have skewness-related measures.

Skewness considers the **entire distribution** through a third standardized moment:
$$\gamma_1 =\frac{E[(R-\mu)^3]}{\sigma^3}$$.

Tail Ratio instead asks a much simpler empirical question:
$$\frac{\text{upper quantile magnitude}}{\text{lower quantile magnitude}}.$$

So:

Skewness
> How asymmetric is the entire distribution around its mean?

Tail Ratio
> How asymmetric are the selected upper and lower tails?

That makes Tail Ratio particularly attractive for trading because you often care much more about the **tails** than about the behavior of the distribution around its mean.

### It is also complementary to downside measures

This is where it fits nicely with your existing library.

Sortino
$$\frac{E[R]-MAR}{\sigma_D}$$

asks about **average performance relative to downside volatility**.

Kappa changes the emphasis on large downside observations through the LPM order.

Omega compares gains and losses relative to MAR across the distribution.

Prospect Ratio introduces explicit loss aversion into the numerator.

Tail Ratio does something different:
> How large is the upper tail relative to the lower tail?

So Tail Ratio is a **distributional diagnostic**, rather than a performance ratio in the same sense as Sharpe or Sortino.

### It can be particularly valuable for strategy classification

Consider several common trading profiles.

- Trend following
  Trend-following strategies can have relatively frequent small losses and fewer but larger gains. You might see $TR>1$. That is consistent with a strategy whose positive tail compensates for numerous smaller losses.
- Short-volatility / option-like strategies
  These can exhibit many small gains and occasional very large losses. You might see $TR\ll1$. That immediately tells you something that a high Sharpe ratio might conceal.
- Mean reversion
  Mean-reversion strategies can also have deceptively attractive average statistics while carrying asymmetric tail risk. Tail Ratio is useful as an additional diagnostic.

### But don't use Tail Ratio alone

There are three important limitations.

1. It ignores the center of the distribution
   Two strategies can have the same Tail Ratio but radically different returns, volatility and drawdowns.
2. It depends on the chosen percentile
   `tail_ratio(0.95)` is not the same statistic as `tail_ratio(0.90)` or `tail_ratio(0.99)`
   For relatively short trading histories, the 99th/1st percentile can also become very unstable.
3. A high Tail Ratio doesn't necessarily mean a good strategy
   Suppose $P_{95}=0.5\%,\qquad P_5=-0.1\%$. Then $TR=5$. That sounds excellent, but the strategy may have almost no return and poor overall economics.

So Tail Ratio should be considered alongside:

- CAGR/annualized return,
- Sharpe,
- Sortino,
- maximum drawdown,
- Calmar,
- skewness,
- perhaps VaR/expected shortfall.

### Overall

I think `tail_ratio` is a very worthwhile addition because it fills a gap between your existing metrics.

Your existing measures answer questions such as
> How much return do I get for risk?

while Tail Ratio asks
> What does the asymmetry of my extreme outcomes look like?

For trading strategies, that is valuable information—especially when combined with **skewness, Sortino/Kappa, drawdown, and Omega/Prospect Ratio**.

In particular, I'd consider **Tail Ratio + skewness + maximum drawdown** a very useful small diagnostic set for detecting strategies whose conventional Sharpe ratio may be hiding unfavorable tail characteristics.

## Bias Ratio

The **Bias Ratio**, introduced by Adil Abdulali, is a distributional diagnostic designed to detect an unusual concentration of returns close to zero. It compares the frequency of small positive returns with the frequency of small negative returns, using a standard-deviation-based threshold around zero.
$$BR=\frac{N(0\le R\le k\sigma)}{1+N(-k\sigma\le R<0)}$$

A low Bias Ratio suggests relatively frequent small negative observations, while an unusually high value can indicate **stale pricing, return smoothing, or illiquidity**. This makes it particularly useful when analyzing hedge funds, alternative investments, or other assets whose valuations may not be marked frequently.

Unlike Sharpe, Sortino, or Tail Ratio, the Bias Ratio is **not a measure of investment performance**. It is better viewed as a diagnostic of the *quality and shape of the observed return series*. For liquid, frequently marked trading strategies it is generally less important, but for illiquid strategies it can provide a useful warning signal that conventional risk statistics may be understating the true variability of returns.

## K-Ratio

The **K-Ratio**, developed by Lars Kestner, measures the **consistency of an investment's equity-curve growth**. Rather than comparing return with volatility like the Sharpe Ratio, it fits a linear regression to the cumulative log-return curve and evaluates the slope relative to its statistical uncertainty.

A high positive K-Ratio indicates that the equity curve has a strong and consistent upward trend, while a value near zero suggests little consistency and a negative value indicates a declining trend.

For trading strategies, the K-Ratio can therefore complement Sharpe and Sortino by answering a different question: **not simply "how much return did the strategy generate for its risk?", but "how consistently did the equity curve grow?"** It is best regarded as an equity-curve/path-quality diagnostic rather than a primary risk-adjusted performance measure.
I agree with your instinct: **K-Ratio is interesting, but I would not prioritize making it streaming**.

The main reason is that it adds a somewhat different perspective, but there is considerable overlap with measures you already have.

### What K-Ratio measures

The K-Ratio asks roughly
> How consistently does the strategy's equity curve rise over time?

You regress cumulative log wealth against time
$$E_t = \sum_{i=1}^{t}\log(1+R_i)$$

and then standardize the estimated slope
$$K=\frac{\hat\beta}{SE(\hat\beta)\sqrt{N}}$$

So it rewards an equity curve with a **strong and statistically consistent upward trend**, rather than simply a high average return.

This makes it somewhat analogous to a **t-statistic for the trend of the equity curve**.

### Why it is different from Sharpe

Sharpe asks
$$SR = \frac{\bar R-R_f}{\sigma_R}$$

K-Ratio instead asks
> Whether the accumulated wealth path has a statistically stable slope.

Consequently, two strategies could have similar Sharpe ratios but different K-Ratios if one produces a much smoother and more consistently rising equity curve.

That is genuinely useful information.

However, you already have several measures that capture related characteristics

- Sharpe: return relative to total volatility
- Sortino: return relative to downside volatility
- Calmar/Martin/Pain: return relative to drawdown characteristics
- Skewness/Tail Ratio: shape of the return distribution
- Maximum drawdown: path-dependent loss
- $M^2$: risk-adjusted return expressed in return units

So K-Ratio is more of a **path-consistency diagnostic** than a fundamental risk measure.

Certainly. For your implementation, I would include the **R/S formula** explicitly:

## Hurst Exponent

The **Hurst exponent** $H$ measures the degree of persistence or mean reversion in a time series. Using rescaled-range (R/S) analysis, this implementation estimates it as
$$H = \frac{\log(R/S)}{\log(N)}$$

where $N$ is the number of observations, $S$ is the standard deviation of returns, and $R$ is the range of the cumulative demeaned returns

$$Z_t=\sum_{i=1}^{t}(R_i-\bar R)$$
$$R=\max(Z_t)-\min(Z_t)$$

Thus, $\frac{R}{S}$ is the **rescaled range**.

The usual interpretation is:

- $H>0.5$: persistent behavior; movements tend to continue in the same direction, potentially favoring **trend-following and momentum** strategies.
- $H\approx0.5$: behavior broadly consistent with a random walk.
- $H<0.5$: anti-persistent behavior; movements tend to reverse, potentially favoring **mean-reversion** strategies.

In trading, the Hurst exponent is most useful as a **regime diagnostic** rather than as a standalone trading signal. A strategy could, for example, give greater weight to trend-following signals when $H$ is persistently above 0.5 and favor mean-reversion when it is below 0.5. However, Hurst estimates are sensitive to the estimation method and sample window, so the apparent persistence should be validated out-of-sample before being incorporated into a trading system.

I would make a fairly strong distinction between the two.

## Gain-to-Pain Ratio

Schwager's definition is
$$GPR=\frac{\sum_{t=1}^{N} R_t}{\left|\sum_{R_t<0}R_t\right|}$$

or, equivalently, using negative returns,
$$GPR=\frac{\text{net return}}{\text{total magnitude of losses}}$$

This has a very intuitive interpretation
> How much net return did the strategy generate for each unit of loss it experienced?

Consider:

- Strategy A: average return = 0.10%, average loss magnitude = 0.05%, $GPR=2.0$
- Strategy B: average return = 0.10%, average loss magnitude = 0.20%, $GPR=0.5$

The difference is immediately understandable.

It also has a nice relationship with the decomposition
$$\sum R_t=\sum R_t^+ -\sum |R_t^-|$$

Therefore
$$GPR=\frac{\text{gross gains}-\text{gross losses}}{\text{gross losses}}=\frac{\text{gross gains}}{\text{gross losses}}-1$$

So GPR is essentially the **profit factor minus one**, when profit factor is defined as gross profit divided by gross loss.

That's a useful connection to trading-system analysis.

### One caveat

GPR can become extremely large when losses are very small, and is undefined when there are no losses. So it should not be treated as a standalone quality measure.

### There is also a useful architectural distinction

Your library is becoming quite sophisticated, and I think it helps to distinguish three categories.

### Core performance/risk measures

These belong naturally in the library:

- Sharpe
- Sortino
- Omega
- Kappa
- Calmar
- Burke
- Martin
- Pain
- Ulcer
- Information Ratio
- Treynor
- Jensen Alpha
- Appraisal Ratio
- Capture ratios
- etc.

These are established performance statistics and/or are directly compatible with `PerformanceAnalytics`.

### Specialized but useful trading statistics

This is where I'd put:

- Gain-to-Pain Ratio
- K-Ratio
- Hurst exponent
- Bias Ratio
- perhaps Prospect Ratio

These aren't necessarily part of `PerformanceAnalytics`, but they provide genuinely different information.

GPR in particular has enough intuitive value that I think it earns its place.

## Capture Measures and Their Use in Evaluating Trading Strategies

Capture measures describe how an investment or trading strategy behaves relative to a benchmark during different market conditions. Rather than asking only whether the strategy has a high average return or Sharpe ratio, capture analysis asks a more practical question
> What does the strategy do when the benchmark goes up, and what does it do when the benchmark goes down?

This distinction is important because two strategies can have the same total return, volatility, or beta while having very different behavior in rising and falling markets.

The capture family consists of six related measures:

- Upside Capture Ratio
- Downside Capture Ratio
- Overall Capture Ratio
- Up Number Ratio
- Down Number Ratio
- Up Percentage Ratio
- Down Percentage Ratio

The first three primarily describe the **magnitude** of participation. The latter four describe the **frequency** of directional behavior or outperformance.

### Upside Capture Ratio

The Upside Capture Ratio measures how much of the benchmark's positive performance the strategy captures.
Only observations where the benchmark return is positive are included
$$R_{b,t} > 0$$

For the geometric version, the strategy and benchmark returns are compounded over all benchmark-positive observations
$$UCR_G=\frac{\prod_{t:R_{b,t}>0}(1+R_{a,t})-1}{\prod_{t:R_{b,t}>0}(1+R_{b,t})-1}$$

This is the default interpretation when measuring investment performance because it respects compounding.

The arithmetic version uses the sums of returns
$$UCR_A=\frac{\sum_{t:R_{b,t}>0}R_{a,t}}{\sum_{t:R_{b,t}>0}R_{b,t}}$$

The two versions answer slightly different questions. The geometric version measures participation in the **compounded benchmark gain**, while the arithmetic version measures participation in the **sum of periodic returns**.

#### Interpretation

An Upside Capture Ratio of

- $1.00$ the strategy captures approximately 100% of benchmark upside.
- $> 1.00$ the strategy participates more strongly in benchmark advances.
- $< 1.00$ the strategy captures less of the benchmark's upside.

For example, an Upside Capture of 1.20 means that, over benchmark-positive periods, the strategy captured roughly 120% of the benchmark's cumulative upside under the chosen definition.

A value below one is not necessarily bad. A market-neutral or defensive strategy may intentionally sacrifice upside participation in exchange for better downside protection.

### Downside Capture Ratio

The Downside Capture Ratio measures how much of the benchmark's downside the strategy experiences.

For PerformanceAnalytics-style capture, benchmark observations with $R_{b,t} \leq 0$ are included.

Geometric definition
$$DCR_G=\frac{\prod_{t:R_{b,t}\leq0}(1+R_{a,t})-1}{\prod_{t:R_{b,t}\leq0}(1+R_{b,t})-1}$$

Because both numerator and denominator are normally negative, the resulting ratio is generally positive.

Arithmetic definition
$$DCR_A=\frac{\sum_{t:R_{b,t}\leq0}R_{a,t}}{\sum_{t:R_{b,t}\leq0}R_{b,t}}$$

#### Interpretation

Unlike Upside Capture, **lower is generally better**.

- $1.00$ approximately the same downside participation as the benchmark.
- $< 1.00$ the strategy suffers less downside.
- $> 1.00$ the strategy suffers more downside.

For example, $UCR=1.15,\qquad DCR=0.70$ would describe a very attractive asymmetric relationship: the strategy captures more than the benchmark's upside while experiencing only about 70% of its downside.

That combination is generally much more interesting than simply having a beta below one.

### Overall Capture Ratio

The Overall Capture Ratio combines the two magnitude measures
$$OCR=\frac{UCR}{DCR}.$$

It summarizes the asymmetry between upside and downside participation.

For example, $UCR=1.10,\qquad DCR=0.80$ produces $OCR=\frac{1.10}{0.80}=1.375$.

A higher value generally indicates a more favorable upside/downside participation profile.

The Overall Capture Ratio should nevertheless not be interpreted in isolation. Two strategies can have the same Overall Capture Ratio but very different absolute behavior. For example, $(1.20,0.80)$ and $(0.60,0.40)$ both produce an Overall Capture Ratio of 1.5, but the second strategy participates much less in both market advances and declines.

Therefore, the individual Upside and Downside Capture Ratios should always be examined alongside the overall ratio.

### Up Number Ratio

Capture ratios measure **magnitude**. The Up Number Ratio instead measures **frequency**.

It asks
> When the benchmark was positive, how often was the strategy also positive?

The formula is
$$UNR=\frac{\#\{R_{a,t}>0 \;\land\; R_{b,t}>0\}}{\#\{R_{b,t}>0\}}$$

Thus, if the benchmark was positive on 100 observations and the strategy was also positive on 65 of them, $UNR=0.65$.

The strategy had positive returns on 65% of benchmark-positive observations.

This measure is deliberately different from Upside Capture. A strategy can have a low Up Number Ratio but a high Upside Capture if it makes relatively few but very large gains when the benchmark rises.

Conversely, it can have a high Up Number Ratio but mediocre Upside Capture if it is frequently positive but its gains are small.

### Down Number Ratio

The Down Number Ratio measures how frequently the strategy also loses money when the benchmark loses money.
The relevant benchmark observations are those where $R_{b,t}<0$.

The formula is
$$DNR=\frac{\#\{R_{a,t}<0 \;\land\; R_{b,t}<0\}}{\#\{R_{b,t}<0\}}$$

Here **lower is generally preferable** for a defensive strategy.

For example, $DNR=0.35$ means that the strategy was negative on only 35% of benchmark-negative observations.

Again, this is a frequency measure rather than a magnitude measure.

A strategy might lose money during only 30% of down-market observations but, when it does lose, lose substantially. Down Number alone would not reveal that problem. Downside Capture would.

This is why the two measures work particularly well together.

### Up Percentage Ratio

The Up Percentage Ratio measures how frequently the strategy **outperformed the benchmark** when the benchmark was positive.

The formula is
$$UPR=\frac{\#\{R_{a,t}>R_{b,t}\;\land\;R_{b,t}>0\}}{\#\{R_{b,t}>0\}}$$

For example, $UPR=0.60$ means that the strategy outperformed the benchmark in 60% of benchmark-positive observations.

This differs from Up Number Ratio.

Suppose the benchmark returned +2% and the strategy returned +1%.

The strategy was positive, so this counts toward **Up Number**, but it did not outperform the benchmark, so it does not count toward **Up Percentage**.

Thus

- Up Number asks whether the strategy made money.
- Up Percentage asks whether the strategy beat the benchmark.

### Down Percentage Ratio

The Down Percentage Ratio measures how frequently the strategy outperformed the benchmark during benchmark-negative observations:
$$DPR=\frac{\#\{R_{a,t}>R_{b,t}\;\land\;R_{b,t}<0\}}{\#\{R_{b,t}<0\}}$$

This is especially interesting for defensive strategies.

Suppose the benchmark falls by 5% while the strategy falls by only 2%: $R_a=-2\%,\qquad R_b=-5\%$.

The strategy still lost money, so it does **not** count as a positive observation for Down Number.

But $-2\%>-5\%$, so it **does** count as an instance of outperforming the benchmark and therefore contributes to Down Percentage.

This distinction is important. A strategy can have an excellent Down Percentage Ratio while still having a relatively high Down Number Ratio if it frequently loses money during market declines but usually loses less than the benchmark.

### Magnitude versus Frequency

The most useful way to understand the capture family is to divide it into two groups.

Magnitude measures (Upside Capture and Downside Capture) answer
> How much does the strategy participate?

Frequency measures (Up Number, Down Number, Up Percentage and Down Percentage) answer
> How often does the strategy behave in a particular way?

This distinction is extremely useful in strategy analysis. Consider two strategies:

| Measure      | Strategy A | Strategy B |
| ------------ | ---------: | ---------: |
| Up Capture   |       1.10 |       0.95 |
| Down Capture |       0.90 |       0.55 |
| Up Number    |       0.70 |       0.45 |
| Down Number  |       0.65 |       0.30 |

- Strategy A participates more strongly in both directions and behaves more consistently with the benchmark.
- Strategy B participates much less in both directions, but has considerably better downside protection.

Neither is automatically superior. Strategy A may be preferable for an aggressive mandate, while Strategy B may be preferable for capital preservation.

### Capture Measures and Beta

Capture ratios should not be confused with beta.

Beta measures the average linear sensitivity of the strategy to benchmark returns:
$$\beta =\frac{\operatorname{Cov}(R_a,R_b)}{\operatorname{Var}(R_b)}$$

Capture ratios instead explicitly condition on the **sign of benchmark returns**.

Consequently, capture analysis can reveal asymmetric behavior that beta hides.
For example, a strategy might have $\beta \approx 1$ but $UCR=1.20,\qquad DCR=0.75$.

Its average linear sensitivity could look similar to the benchmark, while its actual behavior is strongly asymmetric: it participates more in advances and less in declines.

This is one reason capture analysis is particularly useful for nonlinear strategies, tactical strategies, option strategies, and strategies with explicit defensive mechanisms.

### Using Capture Measures to Evaluate a Trading Strategy

A useful evaluation framework is to examine the four most informative quantities together:
$$\boxed{UCR,\quad DCR,\quad UNR,\quad DNR}$$

and then use the percentage measures to understand benchmark-relative consistency.

#### 1. First examine magnitude

Ask
> Does the strategy capture enough upside?

Look at **Upside Capture**.

Then ask
> How much downside does it absorb?

Look at **Downside Capture**.

A desirable defensive profile might look approximately like $UCR>1,\qquad DCR<1$.

That indicates greater upside participation combined with lower downside participation.

However, this should be treated as a desirable characteristic rather than a universal requirement.

#### 2. Then examine frequency

Next ask
> Does the strategy achieve that behavior consistently?

Look at Up Number, Down Number, Up Percentage, and Down Percentage.

For example, a strategy with $UCR=1.20$ but $UNR=0.35$ may be generating its upside participation through a small number of unusually large wins.

That may be perfectly intentional for a convex or trend-following strategy, but it tells you something very different from a strategy with $UCR=1.05,\qquad UNR=0.75$.

The first strategy is more dependent on occasional large gains; the second participates more consistently.

#### 3. Examine downside asymmetry

For most long-only or capital-preservation strategies, downside behavior deserves particular attention.

A useful combination is $DCR<1$ together with $DPR>0.5$.

That suggests the strategy not only has lower downside magnitude but also frequently performs better than the benchmark during benchmark declines.

If Down Capture is low but Down Percentage is also low, the result deserves investigation. It may mean that the strategy has a few very strong defensive observations that dominate the aggregate downside result, rather than consistently outperforming during down markets.

#### 4. Look at the complete profile

A strategy should therefore not be evaluated using a single capture statistic.

For example, $UCR=1.15$ sounds attractive. But if $DCR=1.30$, the strategy is also experiencing substantially more downside than the benchmark.

Conversely $UCR=0.70,\qquad DCR=0.40$
 describes a highly defensive strategy. It sacrifices considerable upside but protects capital effectively during declines.

Whether either profile is desirable depends on the strategy's objective.

### Capture Analysis in Backtesting

Capture measures are particularly useful when evaluating a trading strategy across different market regimes.

Instead of calculating them only over the entire backtest, calculate them over meaningful subperiods:

- bull markets,
- bear markets,
- high-volatility periods,
- low-volatility periods,
- recessionary periods,
- crisis periods,
- individual calendar years,
- rolling windows.

This reveals whether the strategy's apparent asymmetry is stable or merely a consequence of a particular historical period.

For example, a strategy may have excellent historical downside capture because of one crisis in which its hedging mechanism worked exceptionally well. Rolling capture statistics can reveal whether that behavior persists.

This is particularly important because capture measures are **conditional statistics**. They depend on how many benchmark-up and benchmark-down observations are available.

A very small number of benchmark-positive or benchmark-negative observations can make the corresponding statistic unstable.

### A Practical Strategy-Comparison Framework

For a trading strategy, I would organize the capture analysis approximately as follows.

- Return participation $UCR$: How much upside does the strategy capture?
- Downside participation $DCR$: How much downside does it absorb?
- Overall asymmetry $OCR=\frac{UCR}{DCR}$: Does the upside/downside participation appear favorable?
- Directional consistency $UNR,\quad DNR$: How frequently does the strategy move in the same direction as the benchmark?
- Relative consistency $UPR,\quad DPR$: How frequently does the strategy outperform the benchmark in up and down markets?

Taken together, these measures provide a much richer description than simply reporting beta, correlation, or total return.

### Relationship to Risk-Adjusted Performance

Capture ratios should not replace Sharpe, Sortino, Calmar, Information Ratio, or drawdown statistics. They answer a different question.

For example:

- Sharpe Ratio: How much return was earned per unit of total volatility?
- Sortino Ratio: How much return was earned relative to downside risk?
- Information Ratio: How much active return was earned per unit of tracking error?
- Beta: How sensitive was the strategy to benchmark movements on average?
- Capture Ratios: How did the strategy behave specifically during benchmark advances and declines?

This makes capture analysis particularly valuable as a **complementary diagnostic**.

A strategy can have an excellent Sharpe Ratio but poor downside capture. Another can have a modest Sharpe Ratio because it deliberately holds cash much of the time while having exceptionally good downside protection.

The capture statistics explain some of the behavior hidden inside the aggregate risk-adjusted ratios.

### Final Perspective

The greatest strength of the capture family is that it decomposes benchmark-relative behavior into two dimensions:
$$\boxed{\text{Magnitude}}\qquad\text{and}\qquad\boxed{\text{Frequency}}$$

Upside and Downside Capture answer **how much** of market movements the strategy participates in.

Up Number, Down Number, Up Percentage, and Down Percentage answer **how often** it participates or outperforms.

For trading-strategy evaluation, this produces a useful behavioral fingerprint
$$\boxed{(UCR,\ DCR,\ UNR,\ DNR,\ UPR,\ DPR)}$$

rather than a single summary statistic.

In practice, the most attractive profiles often exhibit some combination of strong upside participation, limited downside participation, and frequent relative outperformance during adverse benchmark periods. But the correct profile ultimately depends on the strategy's mandate. A market-neutral, trend-following, defensive, or highly convex strategy should not be judged against the same capture profile as a leveraged long strategy.

The capture family is therefore best viewed not as a collection of isolated ratios, but as a **conditional description of how a strategy makes and loses money relative to its benchmark across different market directions**.

If you want, I can also turn this into a **shorter library-documentation version** with a compact table of all six formulas and “higher/lower is better” guidance.

## Upside Capture Ratio

Also sometimes shortened to Up-Capture Ratio

Source (bacon2023 p. 126):
> $$\text{Up-capture indicator} = \frac{\overline{r}^{\,+}}{\overline{b}^{\,+}}$$
>
> where:
>
> - $\overline{b}^{\,+}$ = average positive benchmark return
> - $\overline{r}^{\,+}$ = average portfolio return for each period in which the benchmark return is positive
>
> The up-capture indicator divides the average portfolio return by the average benchmark return for each period in which the benchmark return is positive - the greater the value the better.

Source [braverock](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R):
> Calculate metrics on how the asset performed in up market, measured by periods when the benchmark asset was up.
>
> Up Capture Ratio: this is a measure of an investment's compound return when the benchmark was up divided by the benchmark's compound return when the benchmark was up. The greater the value, the better.

Source [braverockCran](https://cran.r-project.org/web/packages/PerformanceAnalytics/refman/PerformanceAnalytics.html#UpDownRatios)

Source [morningstar](https://www.morningstar.com/investing-terms/upside-capture-ratio)
> Upside capture ratio is a measurement of an investment's relative performance in up markets. An up market is defined as a period (months or quarters) in which market return was positive.
>
> An example of upside capture ratio: If an investment has an 80% upside capture ratio, then historically when the market was up 10%, the investment captured 80% of that and was up 8%.

Source [morningstar](https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Upside_Capture_Ratio.htm)
> Upside Capture Ratio measures a manager's performance in up markets relative to the market (benchmark) itself. It is calculated by taking the security?s upside capture return and dividing it by the benchmark?s upside capture return.
>
> Upside capture ratio,
> UC = UCR/UCR_bm * 100
> Arithmetic upside capture ratio is calculated by using arithmetic Upside Capture Return for both denominator and numerator.

Source [investopedia](https://www.investopedia.com/terms/u/up-market-capture-ratio.asp)
> The up-market capture ratio is a performance metric that measures how well an investment manager performs relative to a benchmark index during periods of market growth.
>
> A ratio above 1 points to managers outperforming the index. You can calculate the ratio by dividing the manager's returns by the returns of the index during the period of time when the market is in an upward trend.

The Upside Capture Ratio measures how much of a benchmark's positive returns an investment portfolio captures during up-market periods.
It is calculated by dividing the fund's returns during periods when the benchmark is positive by the benchmark's returns during those same periods.

A ratio above 1 indicates the fund outperformed the benchmark during market rallies.
A ratio below 1 indicates the fund underperformed the benchmark during market rallies.
A ratio near 1 suggests the fund's performance closely tracks the benchmark during up markets.
This metric is typically used alongside the Downside Capture Ratio to assess a fund manager's ability to generate gains while managing risk.  Passive index funds generally have capture ratios close to 1, while active managers may exhibit higher or lower ratios depending on their strategy.

$$\text{Upside Capture Ratio} = \frac{text{Asset Returns in Up Markets}}{text{Benchmark Returns in Up Markets}}$$

## Downside Capture Ratio

Also sometimes shortened to Down-Capture Ratio

Source (bacon2023 p. 126):
> $$\text{Down-capture indicator} = \frac{\overline{r}^{\,-}}{\overline{b}^{\,-}}$$
>
> where:
>
> - $\overline{b}^{\,-}$ = average negative benchmark return
> - $\overline{r}^{\,-}$ = average portfolio return for each period in which the benchmark return is negative
>
> The down-capture indicator divides the average portfolio return by the average benchmark return for each period in which the benchmark return is negative. Lower values are preferred.

Source [braverock](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R):
> Calculate metrics on how the asset performed in down market, measured by periods when the benchmark asset was down.
>
> Down Capture Ratio: this is a measure of an investment's compound return when the benchmark was down divided by the benchmark's compound return when the benchmark was down. The lower the value, the better.

Source [braverockCran](https://cran.r-project.org/web/packages/PerformanceAnalytics/refman/PerformanceAnalytics.html#UpDownRatios)

Source [morningstar](https://www.morningstar.com/investing-terms/downside-capture-ratio)
> Downside capture ratio is a measurement of an investment's relative performance in down markets. A down market is defined as a period (months or quarters) in which the market return was negative.
>
> For example, if an investment has an 80% downside capture ratio, then historically when the market was down 10%, the investment captured 80% of that and was down 8%.

Source [morningstar](https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Downside_Capture_Ratio.htm)
> Downside Capture Ratio measures manager's performance in down markets. A down-market is defined as those periods (months or quarters) in which market return is less than 0. In essence, it tells you what percentage of the down-market was captured by the manager. For example, if the ratio is 110%, the manager has captured 110% of the down-market and therefore underperformed the market on the downside.
>
> DCR = downside capture return of the subject
> DCR_bm = downside capture return of the benchmark
> DC = DCR/DCR_bm * 100
> Arithmetic downside capture ratio is calculated by using arithmetic Downside Capture Return for both denominator and numerator

Arithmetic downside capture ratio is calculated by using arithmetic Downside Capture Return for both denominator and numerator

Source [investopedia](https://www.investopedia.com/terms/d/down-market-capture-ratio.asp)
> The down-market capture ratio evaluates an investment's ability to limit losses compared to a benchmark during market downturns.
>
> The down-market capture ratio is a tool for assessing investment managers' performance during market declines. A ratio of less than 100 signifies outperformance of the index during downturns.
>
> It's important to consider both the down-market and up-market capture ratios to gain a comprehensive understanding of a manager's performance across different market conditions. Comparing the up-market and down-market ratios can offer insights into the manager's ability to capitalize on favorable market conditions and mitigate losses during downturns.
>
> Evaluate investment managers using these ratios as part of a broader assessment, incorporating other performance metrics and qualitative factors.

The Downside Capture Ratio (also known as the Down-Market Capture Ratio) is a performance metric that evaluates an investment manager's ability to limit losses compared to a benchmark index during market downturns.  It is calculated by dividing the fund's returns during negative market periods by the benchmark's returns during the same periods.

A ratio below 1 indicates that the investment lost less than the benchmark, signifying outperformance and effective downside protection.  Conversely, a ratio above 100 means the investment declined more than the benchmark, indicating underperformance during bearish trends.  Investors typically analyze this ratio alongside the Upside Capture Ratio to determine if a manager's defensive performance in down markets is compensated by strong gains in up markets.

$$\text{Downside Capture Ratio} = \frac{text{Asset Returns in Down Markets}}{text{Benchmark Returns in Down Markets}}$$

## Overall Capture Ratio

Source [morningstar]{https://global.morningstar.com/en-ca/personal-finance/what-are-upside-and-downside-capture-ratios}
> The term "upside/downside capture ratio" might sound wonky, but the concept is pretty straightforward. In short, the statistics show you whether a given fund has outperformed -- gained more or lost less than -- a broad market benchmark during periods of market strength and weakness, and if so, by how much.
>
> Upside capture ratios for funds are calculated by taking the fund's monthly return during months when the benchmark had a positive return and dividing it by the benchmark return during that same month. Downside capture ratios are calculated by taking the fund's monthly return during the periods of negative benchmark performance and dividing it by the benchmark return. Morningstar.ca displays the upside and downside capture ratios over one-, three-, five-, 10- and 15-year periods by calculating the geometric average for both the fund and index returns during the up and down months, respectively, over each time period.
>
> An upside capture ratio over 100 indicates a fund has generally outperformed the benchmark during periods of positive returns for the benchmark. Meanwhile, a downside capture ratio of less than 100 indicates that a fund has lost less than its benchmark in periods when the benchmark has been in the red. The benchmark used to determine the ratios is determined by the fund's category and is indicated right below the table. For some context, we also show the category average upside/downside capture ratios for those same time periods.

The Capture Ratio is a statistical metric used to evaluate an investment's performance against a benchmark index during bullish and bearish market phases.  It consists of two distinct components: the Up-Market Capture Ratio and the Down-Market Capture Ratio, which together assess a fund manager's ability to generate gains during rallies and mitigate losses during downturns.

Up-Market Capture Ratio This metric measures how much of the benchmark's gains an investment captures during periods when the market is rising.
Formula: $\text{Fund Returns during Upside} / \text{Benchmark Returns during Upside}$
Interpretation: A ratio above 1 indicates the fund outperformed the benchmark during upswings, while a ratio below 1 suggests underperformance.  Passive index funds typically have a ratio close to 100%.

Down-Market Capture Ratio This metric evaluates how much of the benchmark's losses an investment captures during periods when the market is falling.
Formula: $\text{Fund Returns during Downside} / \text{Benchmark Returns during Downside}$
Interpretation: A ratio below 1 is desirable, indicating the fund lost less than the benchmark during downturns.  A ratio above 1 means the fund experienced greater losses than the benchmark.

Overall Assessment Investors often compare both ratios to determine the quality of risk-adjusted returns. A Market Capture Ratio can be derived by dividing the Up-Market ratio by the Down-Market ratio; a value greater than 1 generally indicates that the fund�'s up-market performance compensates for its down-market performance, signaling effective management.

The Overall Capture Ratio is a performance metric that measures the asymmetry of a fund's returns by dividing its Up-Side Capture Ratio by its Down-Side Capture Ratio.  This single figure indicates whether a fund captures more of the benchmark's upside than it loses during downside periods, providing a clearer picture of risk-adjusted performance than Beta or Alpha alone.

The ratio is calculated using the following formula:

$$\text{Overall Capture Ratio}=\frac{\text{Up-Side Capture Ratio}}{\text{Down-Side Capture Ratio}}$$
Where:

- Up-Market Capture Ratio: The fund's return during months when the benchmark rose, divided by the benchmark's return.
- Down-Market Capture Ratio: The fund's return during months when the benchmark fell, divided by the benchmark's return.

- Value > 1.0: Indicates favorable asymmetry, meaning the fund captures more upside than downside.  This is generally considered a sign of skilled management or effective defensive positioning.
- Value < 1.0: Indicates unfavorable asymmetry, meaning the fund loses more relative to the benchmark in down markets than it gains in up markets.

- Benchmark Selection: The ratio is only meaningful if calculated against a style-matched benchmark (e.g., comparing a large-cap fund to the S&P 500).
- Negative Values: If a fund has a negative down-side capture (gaining when the market falls), the overall ratio becomes negative and loses standard interpretive value; in such cases, evaluate up and down captures separately.
- Complementary Metrics: Capture ratios should be used alongside other metrics like the Sharpe Ratio or Jensen's Alpha for a complete evaluation of manager performance.

The Overall Capture Ratio is a performance metric that measures a fund's ability to generate favorable asymmetry by capturing more upside than downside relative to a benchmark.  It is calculated by dividing the Up-Market Capture Ratio by the Down-Market Capture Ratio

The Overall Capture Ratio synthesizes two distinct metrics to reveal whether a manager adds value through skill or simply by taking on more risk. Unlike Beta, which measures general market sensitivity, this ratio specifically isolates performance during bull and bear markets.

- $> 1.0$ Favorable Asymmetry: The fund captures more upside than downside.  This is the ideal profile for long-term wealth creation.
- $= 1.0$ Neutral: The fund participates equally in gains and losses relative to the benchmark (similar to a passive index fund).
- $< 1.0$ Unfavorable Asymmetry: The fund suffers more in downturns than it gains in upturns relative to the benchmark.

A "good" Overall Capture Ratio depends on investment objectives, but generally:

- Above 1.0: Indicates the manager is adding value through asymmetry.
- Above 1.2: Considered excellent, suggesting the manager successfully limits losses while participating reasonably in gains.

Context Matters: A fund with 85% up-capture and 60% down-capture (Ratio: 1.42) is often superior to a fund with 110% up-capture and 105% down-capture (Ratio: 1.05), as avoiding large drawdowns mathematically aids long-term compounding more than marginal extra gains in bull markets.

## Up Number Ratio

Source (bacon2023 p. 126):
> The up-number ratio measures the percentage of returns in each measurement period in which the portfolio returns are greater than zero when the benchmark returns are greater than zero. Ideally the ratio should be 100% - the closer to 100% the better.

Source [braverock](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R):
> Calculate metrics on how the asset performed in up market, measured by periods when the benchmark was up.
>
> Up Number Ratio: this is a measure of the number of periods that the investment was up when the benchmark was up, divided by the number of periods that the benchmark was up.

Source [braverockCran](https://cran.r-project.org/web/packages/PerformanceAnalytics/refman/PerformanceAnalytics.html#UpDownRatios)

Source [morningstar](https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Up_Number_Ratio.htm)
> Up number ratio is a measure of the number of periods that the investment was up, when the benchmark was up, divided by the number of periods that the benchmark was up. The larger the ratio, the better.

The Up-Number Ratio is a performance metric defined by Morningstar as the number of periods an investment was up when the benchmark was up, divided by the total number of periods the benchmark was up.  A higher ratio indicates better performance, as it signifies the investment captured more of the benchmark's positive movements. While Investopedia and Wikipedia do not have dedicated entries for the "Up-Number Ratio," they do cover related concepts such as the Up-Market Capture Ratio, which measures the percentage of index returns captured during up-market periods, and the PEG Ratio, which compares a stock's price-to-earnings ratio to its earnings growth rate.

The Up-Number Ratio is a specific statistical measure used primarily by Morningstar to evaluate how frequently an investment participates in a benchmark's gains.
Unlike capture ratios that measure the magnitude of returns, this ratio measures the frequency of positive performance.

For example, if a benchmark had positive returns in 100 months, and the investment in question also had positive returns in 85 of those specific months, the Up-Number Ratio would be 0.85 (or 85%).
A higher ratio is always better, indicating the investment rarely misses out on the benchmark's up markets.

It is crucial to distinguish the Up-Number Ratio from the more common Up-Market Capture Ratio, which is widely covered on Investopedia and often confused with the former.

- Up-Number Ratio (Frequency): Measures how often the investment was positive when the benchmark was positive.  It ignores how much the investment gained, only caring that the return was $>0$.
- Up-Market Capture Ratio (Magnitude): Measures how much of the benchmark's return the investment captured.  It is calculated by dividing the investment's average return during up markets by the benchmark's average return during those same periods. An investment could have a low Up-Number Ratio (it was flat or down often) but a massive Up-Market Capture Ratio (when it did go up, it soared), or vice versa.

There is also a related metric called the Up-Percentage Ratio, which measures the number of periods the investment outperformed the benchmark when the benchmark was up, divided by the total up periods.
This is a stricter test than the Up-Number Ratio, which only requires the investment to be positive, not necessarily better than the benchmark.

The metric remains a proprietary or specialized statistic found primarily within Morningstar Direct and similar professional analytics platforms rather than general financial education resources.

Investors use the Up-Number Ratio to identify consistency rather than explosive growth.

- High Ratio (>0.90): Suggests the fund rarely misses a rally. This is typical of index funds or highly correlated active funds.
- Low Ratio (<0.70): Suggests the fund often sits out market rallies. This might be acceptable for a defensive strategy (e.g., holding cash or bonds during stock rallies) but is concerning for an aggressive growth fund that claims to track the market
he ratio quantifies the consistency of an investment's positive performance relative to its benchmark during bullish periods. It is calculated as:

$$\text{Up-Number Ratio}=\frac{\text{Number of periods where Investment > 0 AND Benchmark > 0}}{\text{Total number of periods where Benchmark > 0}}$$

- Numerator: Counts only the months (or periods) where both the investment and the benchmark posted positive returns.
- Denominator: Counts all months where the benchmark posted a positive return.

Interpretation: The result is a decimal or percentage. A ratio of 1.0 (or 100%) means the investment rose every single time the benchmark rose.  A ratio of 0.85 means the investment missed out on 15% of the benchmark's up-periods, even if it outperformed significantly in the periods it did rise.

## Down Number Ratio

Source (bacon2023 p. 126):
> The down-number ratio measures the percentage of returns in each measurement period in which the portfolio returns are less than zero when the benchmark returns are less than zero. Ideally, but very rarely, the ratio should be 0%. The lower the ratio the better, although for highly correlated returns, ratios close to 100% should be expected.

Source [braverock](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R):
> Calculate metrics on how the asset performed in down market, measured by periods when the benchmark was down.
>
> Down Number Ratio: this is a measure of the number of periods that the investment was down when the benchmark was down, divided by the number of periods that the benchmark was down.

Source [braverockCran](https://cran.r-project.org/web/packages/PerformanceAnalytics/refman/PerformanceAnalytics.html#UpDownRatios)

Source [morningstar](https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Down_Number_Ratio.htm)
> The down number ratio is a measure of the number of periods that the Investment was down when the Benchmark was down, divided by the number of periods that the Benchmark was down. The smaller the ratio, the better.

The Down-Number Ratio is the counterpart to the Up-Number Ratio, measuring the frequency with which an investment declines during periods when its benchmark also declines.
Like its "up" equivalent, it is a metric of consistency (frequency) rather than severity (magnitude).

The Down-Number Ratio quantifies how often a fund participates in the benchmark's losses. It is calculated as:

$$\text{Down-Number Ratio}=\frac{\text{Number of periods where Investment < 0 AND Benchmark < 0}}{\text{Total number of periods where Benchmark < 0}}$$
where
- Numerator: Counts the months where both the investment and the benchmark posted negative returns.
- Denominator: Counts all months where the benchmark posted a negative return.

Interpretation:
- Lower is Better: A ratio of 0.0 means the investment never fell when the benchmark fell (ideal downside protection).
- Higher is Worse: A ratio of 1.0 means the investment fell every single time the benchmark fell.

Example: A ratio of 0.40 indicates the fund managed to avoid losses in 60% of the benchmark's down months, even if it fell sharply in the other 40%.

Investors use the Up-Number Ratio to identify consistency rather than explosive growth.

- High Ratio (>0.90): Suggests the fund rarely misses a rally. This is typical of index funds or highly correlated active funds.
- Low Ratio (<0.70): Suggests the fund often sits out market rallies. This might be acceptable for a defensive strategy (e.g., holding cash or bonds during stock rallies) but is concerning for an aggressive growth fund that claims to track the market.

## Up Percentage Ratio

Source (bacon2023 p. 126):
> More interestingly the up-percentage ratio measures the percentage of periods in which the excess return of the portfolio against the benchmark is greater than zero in each measurement period when the benchmark return is greater than zero. In other words, how often does the portfolio manager outperform a rising market?

Source [braverock](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R):
> Calculate metrics on how the asset performed in up market, measured by periods when the benchmark was up.
>
> Up Percentage Ratio: this is a measure of the number of periods that the investment outperformed the benchmark when the benchmark was up, divided by the number of periods that the benchmark was up. Unlike the prior two metrics, in both cases a higher value is better.

Source [braverockCran](https://cran.r-project.org/web/packages/PerformanceAnalytics/refman/PerformanceAnalytics.html#UpDownRatios)

Source [morningstar](https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Up_Percentage_Ratio.htm)
> Up percentage ratio is a measure of the number of periods that the Investment outperformed the Benchmark when the Benchmark was up, divided by the number of periods that the benchmark was up. The larger the ratio, the better.

The Up Percentage Ratio is a performance metric used in finance to evaluate how often an investment outperforms its benchmark during periods when the benchmark itself is rising.

The Up Percentage Ratio measures the number of periods that an investment outperformed its benchmark when the benchmark was up, divided by the total number of periods that the benchmark was up. It specifically isolates performance during positive market conditions to determine consistency in capturing gains.

- Interpretation: A larger ratio indicates better performance. It shows the percentage of "up markets" in which the manager successfully beat the index.
- Distinction: Unlike the Up-Market Capture Ratio, which measures the magnitude of gains relative to the benchmark, the Up Percentage Ratio measures the frequency of outperformance.

The formula for the Up Percentage Ratio is:

$$\text{Up Percentage Ratio}=\frac{\text{Number of periods Investment > Benchmark (when Benchmark > 0)}}{\text{Total number of periods Benchmark > 0}}$$

Where:

- Numerator: The count of time periods (e.g., months, quarters) where the benchmark had a positive return AND the investment's return was higher than the benchmark's return.
- Denominator: The total count of time periods where the benchmark had a positive return.

While the Up-Market Capture Ratio tells you if a manager captured more growth than the index (e.g., a ratio of 120 means they captured 120% of the upside), the Up Percentage Ratio tells you how reliable that outperformance was across different time periods.

To fully assess an investment manager, the Up Percentage Ratio is often viewed alongside other capture ratios:

- Up Percentage Ratio: How often did the manager beat the benchmark when the market was up?
- Up-Market Capture Ratio: How much of the benchmark's gain did the manager capture?  (e.g., did they get 110% of the rise?)
- Up Number Ratio: How often did the investment rise when the benchmark rose? (Does not require outperformance, just positive correlation).

## Down Percentage Ratio

Source (bacon2023 p. 126):
> The down-percentage ratio measures the percentage of periods in which the excess return is greater than zero in each measurement period when the benchmark return is less than zero. In other words, how often does the portfolio manager outperform a falling market?

Source [braverock](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R):
> Calculate metrics on how the asset performed in down market, measured by periods when the benchmark was down.
>
> Down Percentage Ratio: this is a measure of the number of periods that the investment outperformed the benchmark when the benchmark was down, divided by the number of periods that the benchmark was down. Unlike the prior two metrics, in both cases a higher value is better.

Source [braverockCran](https://cran.r-project.org/web/packages/PerformanceAnalytics/refman/PerformanceAnalytics.html#UpDownRatios)

Source [morningstar](https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Down_Percentage_Ratio.htm)
> The down percentage ratio is a measure of the number of periods that the Investment outperformed the Benchmark when the Benchmark was down, divided by the number of periods that the benchmark was down. The larger the ratio, the better.

The Down Percentage Ratio is a financial metric used to evaluate an investment's performance relative to a benchmark specifically during periods when the benchmark is declining.  It measures the frequency of outperformance rather than the magnitude of returns.

The Down Percentage Ratio calculates the proportion of time an investment outperforms its benchmark when the benchmark itself is posting negative returns. Unlike capture ratios that measure the size of losses, this ratio measures the consistency of defensive performance.

The metric is defined as:
$$\text{Down Percentage Ratio}=\frac{\text{Number of periods Investment > Benchmark (when Benchmark was down}}{\text{Total number of periods Benchmark was down}}$$

A higher ratio indicates better performance, as it means the investment managed to beat the benchmark more often during market downturns.

To calculate the ratio, analysts identify all periods (typically months or quarters) where the benchmark index had negative returns. They then count how many of those specific periods the investment fund achieved a higher return (i.e., lost less or gained more) than the benchmark.

- Measurement: Frequency of outperformance
- Formula Basis: Count of periods
- Ideal Value: Higher is better (>50%)
- Interpretation: "How often did it lose less?"

Investors utilize this ratio to assess a fund manager's ability to navigate bear markets consistently.

- Ratio > 50%: Indicates the manager outperformed the benchmark in the majority of down periods, suggesting strong defensive stock selection or asset allocation.
- Ratio < 50%: Suggests the fund often falls more than the benchmark when the market drops, even if the long-term average loss (capture ratio) appears acceptable.

This metric is particularly valuable for investors prioritizing capital preservation who want to avoid funds that only look good due to a few lucky months of outperformance amidst many months of underperformance

## Morningstar custom statistics

[morningstart custom calculations](https://morningstardirect.morningstar.com/clientcomm/customcalculations.pdf)
[morningstar direct samples](https://admainnew.morningstar.com/directhelp/samples.pdf)
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Bull_Beta.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Bear_Beta.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Sortino_Ratio.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Downside_Deviation.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Information_Ratio.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Excess_Return.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Tracking_Error.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Sharpe_Ratio.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Average_Gain.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Loss_Standard_Deviation.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Downside_Capture_Return.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Upside_Capture_Return.htm
https://admainnew.morningstar.com/directhelp/Glossary/Custom_Statistics/Relative_Risk.htm
https://gladmainnew.morningstar.com/clientcomm/F.WorkbookSamples.pdf
https://morningstardirect.morningstar.com/clientcomm/CustomDatabase_EMEA.pdf
https://morningstardirect.morningstar.com/clientcomm/PerformanceReporting2012.pdf
https://morningstardirect.morningstar.com/clientcomm/CustomDatabase.pdf
https://morningstardirect.morningstar.com/clientcomm/PerformanceReports.pdf

## Misc links

[Smart Sharpe ratio](https://www.alternativesoft.com/the-difference-between-the-Sharpe-ratio-and-the-Smart-Sharpe-Ratio.html)

[portolio optimization](https://www.kenwuyang.com/en/post/portfolio-optimization-with-python/)

[bacon 3rd ed excel](https://bcs.wiley.com/he-bcs/Books?action=resource&bcsId=12534&itemId=1119831946&resourceId=49583)
