# Primitives

## Welfords algorithm

https://natural-blogarithm.com/post/variance-welford-vs-numpy/

The idea of Welford’s algorithm is to decompose the calculations of the mean and the variance into a form that can easily be updated whenever a new observation comes in.
For the mean this transformation is produced by simply rewriting the formula for $\bar{x}_n$ as the weighted average between the old mean $\bar{x}_{n-1}$ and the new observation $x_n$:

$$\bar{x}_n = \frac{\sum_{i=1}^n x_i}{n}  = \frac{\sum_{i=1}^{n-1} x_i}{n} + \frac{x_n}{n} = \frac{(n-1)\bar{x}_{n-1} + x_n}{n} = \bar{x}_{n-1} + \frac{x_n - \bar{x}_{n-1}}{n}$$

The decomposition for the variance is slightly more complicated. Only looking at the enumerator (the sum of squares of differences) of the formula for $s^2_n$ from above we can derive it in the following way:

$$\begin{aligned}
&\sum_{i=1}^n (x_i - \bar{x}_n)^2  =  \sum_{i=1}^n (x_i - \underbrace{(\bar{x}_{n-1} + \frac{x_n - \bar{x}_{n-1}}{n})}_{\text{decomposition of }\bar{x}_n})^2 = \sum_{i=1}^n ((x_i - \bar{x}_{n-1}) - (\frac{x_n - \bar{x}_{n-1}}{n}))^2\\\\
=&\sum_{i=1}^n (x_i - \bar{x}_{n-1})^2 - \underbrace{2 \sum_{i=1}^n (x_i - \bar{x}_{n-1})(\frac{x_n - \bar{x}_{n-1}}{n})}_{=2 \frac{(x_n - \bar{x}_{n-1})}{n}\sum_{i=1}^n(x_i - \bar{x}_{n-1})} + \underbrace{\sum_{i=1}^n (\frac{x_n - \bar{x}_{n-1}}{n})^2}_{=n \cdot (\frac{x_n - \bar{x}_{n-1}}{n})^2 = \frac{(x_n - \bar{x}_{n-1})^2}{n}} \\\\
=&\underbrace{\sum_{i=1}^{n-1} (x_i - \bar{x}_{n-1})^2}_{=(n-2) \cdot s_{n-1}^2} + \underbrace{(x_n - \bar{x}_{n-1})^2 +  \frac{(x_n - \bar{x}_{n-1})^2}{n}}_{=\frac{(n+1)(x_n - \bar{x}_{n-1})^2}{n}} - 2 \frac{(x_n - \bar{x}_{n-1})}{n}\underbrace{\sum_{i=1}^n(x_i - \bar{x}_{n-1})}_{=(n \cdot \bar{x}_n - n\cdot\bar{x}_{n-1})} \\\\
=&(n-2) \cdot s_{n-1}^2 + \frac{(n+1)(x_n - \bar{x}_{n-1})^2}{n} - 2 \frac{(x_n - \bar{x}_{n-1})}{n}\underbrace{(n \cdot \bar{x}_n - n\cdot\bar{x}_{n-1})}_{=n \cdot (\bar{x}_{n-1} + \frac{x_n - \bar{x}_{n-1}}{n} - \bar{x}_{n-1})} \\\\
=&(n-2) \cdot s_{n-1}^2 + \frac{(n+1)(x_n - \bar{x}_{n-1})^2}{n} - 2 \frac{(x_n - \bar{x}_{n-1})^2}{n} \\\\
=&(n-2) \cdot s_{n-1}^2 + \frac{(n-1)(x_n - \bar{x}_{n-1})^2}{n}
\end{aligned}$$

## Foo

https://search.brave.com/ask?q=Can+Kahan+summation+be+combined+with+Welford%27s%3F&conversation=095142bcaeec0ae61de89c46fdaef9ab641a

Yes, Kahan summation can be combined with Welford’s algorithm
to improve numerical stability, especially when dealing with large datasets or values that differ significantly in magnitude
Why Combine Them?
• Welford’s algorithm is already numerically stable for computing running mean and variance by avoiding catastrophic cancellation through incremental updates.
• However, it still involves summations that can accumulate round-off errors, particularly when:
◦ Values differ greatly in magnitude (e.g., 1 + 1e-12)
◦ The dataset is extremely large (millions or billions of samples)
In such cases, Kahan summation (or its improved variants like Kahan–Babuška–Neumaier) can be used to compensate for floating-point rounding errors during the accumulation steps in Welford’s method.

How to Combine Them
Welford’s algorithm maintains:
• Running mean:
• [ \mu_n = \mu_{n-1} + \frac{x_n - \mu_{n-1}}{n} ]
• Running sum of squared differences:
• [ M_{2,n} = M_{2,n-1} + (x_n - \mu_{n-1})(x_n - \mu_n) ]
To integrate Kahan summation:
• Use Kahan-compensated summation when updating cumulative quantities like M₂ or when computing intermediate sums.
• This is most beneficial when summing terms like (x_n - μₙ₋₁)(x_n - μₙ) over many iterations.

Example: Kahan-Enhanced Welford in Python
class KahanWelford:
    def __init__(self):
        self.n = 0
        self.mean = 0.0
        self.M2 = 0.0
        # Kahan compensation terms
        self.c_mean = 0.0
        self.c_M2 = 0.0

    def update(self, x):
        self.n += 1
        delta = x - self.mean
        # Update mean with Kahan compensation
        delta_kahan = delta / self.n - self.c_mean
        new_mean = self.mean + delta_kahan
        self.c_mean = (new_mean - self.mean) - delta_kahan
        self.mean = new_mean

        delta2 = x - self.mean
        term = delta * delta2
        # Update M2 with Kahan compensation
        term_kahan = term - self.c_M2
        new_M2 = self.M2 + term_kahan
        self.c_M2 = (new_M2 - self.M2) - term_kahan
        self.M2 = new_M2

    def variance(self):
        return self.M2 / (self.n - 1) if self.n > 1 else 0.0

When Is It Useful?
• High-precision applications: cryptography, scientific simulations, financial modeling.
• Streaming data over long durations: e.g., health monitoring of hardware random number generators.
• Single-precision (float32) environments: where rounding errors are more pronounced.
Trade-offs
• Slight performance overhead: Kahan summation adds a few extra operations per update.
• Increased code complexity: but often worth it for critical applications.
https://github.com/online-ml/watermill.rs/tree/master/src
https://github.com/bagol1000/fastwindow/blob/main/src/rolling_moments.cpp
https://github.com/breuderink/epsilon/blob/main/src/transform.c
https://github.com/kuiperzone/Compensated-Accumulators/blob/master/CompensatedAccumulators.Test/CompensatedSumTest.cs
https://github.com/kuiperzone/Compensated-Accumulators/tree/master/CompensatedAccumulators
https://github.com/JuliaMath/KahanSummation.jl/tree/master
https://github.com/JuliaMath/KahanSummation.jl/blob/master/test/runtests.jl
https://quantsrus.github.io/post/klein_sum_and_monte_carlo_statistics/
https://github.com/numpy/numpy/issues/8786

Welford’s Algorithm is primarily used for computing rolling variance and standard deviation because it offers numerical stability and O(1) time complexity by updating statistics incrementally as the window slides. Libraries like pandas and the R roll package utilize this method to avoid the precision loss associated with direct summation.
Kahan Summation is a technique designed to reduce floating-point error accumulation during rolling sums and means. It compensates for lost low-order bits by maintaining a running correction variable, ensuring higher precision in sliding window calculations where small errors can compound.

Implementation in Rolling Statistics Libraries

Library/Context
Algorithm
Purpose

pandas
Welford’s
Used for .rolling().std() and .rolling().var() for numerical stability.

pandas
Kahan Summation
Proposed/used for .rolling().sum() to minimize precision errors in sliding windows.

R roll
Welford’s
Default online algorithm for variance/std to ensure speed and accuracy in time-series data.

ajcr/rolling
Welford’s
Used for variance and standard deviation calculations.

rust-rolling-stats
Welford’s
Uses Welford's Online Algorithm for mean, variance, and standard deviation.


For rolling means, Kahan summation is often preferred over naive summation to maintain accuracy, while Welford’s is reserved for variance-related metrics due to its superior handling of squared differences




In the folder external\primitives\ I have  the primitives.py and test_primitives.py where I try to implement streaming version of simple stats calculation: mean, variance, skewness, kurtosis.

I want this calculator also to operate with rolling window, so I made update(x) and revert(x) methods for that.
I try different approaches to achieve o(1) complexity and numerical stability.

So far I like MomentsKlein and Klein variants, but they give different results compared to each other and to scipy in TestCompareStats.test_bacon():

Klein mean: 0.009000000000000001
MomentsKlein mean: 0.009000000000000001
SciPy mean: 0.009000000000000003

Klein variance: 0.0014989166666666666
MomentsKlein variance: 0.0015640869565217393
SciPy variance: 0.0014989166666666668

Klein skewness: -0.08817174934967528
MomentsKlein skewness: -0.08256245520856803
SciPy skewness: -0.08256245520856835

Klein kurtosis: -0.40766032118608714
MomentsKlein kurtosis: -0.5675462058921261
SciPy kurtosis: -0.5675462058921252

to run the test, you should be in the root directory of this repo and run
python -m unittest external.primitives.test_primitives.TestCompareStats

Please analyze both implementations and help me with:
- adding ddof in MomentsKlein __init__ and fixing calculations to use ddof
- When mean/variance/skewness/kurtosis from MomentsKlein or Klein are "almost equal" to scipy, why there are some small differences?
- why variance in MomentsKlein is different from Klein and scipy?
- why skewness in Klein is different from MomentsKlein and scipy?
- why kurtosis in Klein is different from MomentsKlein and scipy?
- which method do you think is better, MomentsKleim or Klein, or maybe something else?
