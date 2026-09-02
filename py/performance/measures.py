from typing import List
import collections
import math

from ..streaming_kbn import RawMomentsKleinKBN, KleinKBNAccumulator
from . import core

_SQRT2 = 1.4142135623730950488016887242097

PERIODS_PER_ANNUM_YEAR = 1
PERIODS_PER_ANNUM_QUARTER = 4
PERIODS_PER_ANNUM_MONTH = 12
PERIODS_PER_ANNUM_WEEK = 52
PERIODS_PER_ANNUM_DAY = 252

PERIODS_PER_ANNUM_MINUTE_US_EQUITIES = 98280
"""390 regular-session minutes/day by 252 trading days/year"""

PERIODS_PER_ANNUM_MINUTE_CRYPTO = 525600
"""1440 minutes/day by 365 days/year"""


class Measures:
    """
   Streaming calculation of time-series performance and risk measures.

    ``Measures`` operates on return observations that have
    a common, explicitly defined observation period. It does not require
    timestamps and does not perform time-based resampling.

    ``periods_per_annum`` defines the annualization convention used by
    annualized measures. It is the number of return observations that are
    assumed to represent one year.

    Common conventions include:

        daily equity returns:
            periods_per_annum = 252

        weekly returns:
            periods_per_annum = 52

        monthly returns:
            periods_per_annum = 12

        quarterly returns:
            periods_per_annum = 4

        annual returns:
            periods_per_annum = 1

        one-minute US equity-session returns:
            periods_per_annum = 98_280

    Intraday ``periods_per_annum`` values are conventions rather than
    universal constants. For example, the annualization factor for
    one-minute returns depends on the trading calendar and session hours.

    The class is intended for regularly spaced periodic return series.
    Irregular event-based observations, such as tick returns, should
    generally be transformed into a suitable time-based return series
    before being passed to ``Measures``. The class does not infer a
    periodicity from timestamps.

    Args:
        periods_per_annum:
            Number of return periods per annum used for annualization.
            This value must be positive. It also determines the conversion
            of the annual risk-free rate and annual target return to their
            corresponding periodic rates.

            For example, with ``periods_per_annum=12``, an annual rate ``r``
            is converted to a monthly rate as:

                (1 + r) ** (1 / 12) - 1

        annual_risk_free_rate:
            Annual risk-free rate expressed as a decimal.

            The rate is converted to the periodic risk-free rate implied by
            ``periods_per_annum`` and is used by measures that require a
            risk-free rate.

            Default: 0.0

        annual_target_return:
            Annual target return, also known as the Minimum Acceptable
            Return (MAR), expressed as a decimal.

            The rate is converted to the periodic target return implied by
            ``periods_per_annum`` and is used by measures that require a
            target return or MAR.

            Default: 0.0

        rolling_window_size:
            Number of most recent return observations used by rolling
            measures.

            A value of zero or ``None`` specifies an unbounded running
            window, in which all observations are retained.

            A positive value specifies the maximum number of observations
            retained in the rolling window.

            Default: 0
     """
    def __init__(self,
        periods_per_annum: float = 252.0,
        annual_risk_free_rate: float = 0.,
        annual_target_return: float = 0.,
        rolling_window_size: int = 0):

        if periods_per_annum <= 0:
            raise ValueError("periods_per_annum must be positive")
        self.periods_per_annum = periods_per_annum
        self._sqrt_periods_per_annum = math.sqrt(periods_per_annum)

        self._annual_risk_free_rate = annual_risk_free_rate
        self.risk_free_rate = annual_risk_free_rate \
            if annual_risk_free_rate == 0 or periods_per_annum == 1 \
            else ((1 + annual_risk_free_rate) ** (1/periods_per_annum) - 1)

        self.target_return = annual_target_return \
            if annual_target_return == 0 or periods_per_annum == 1 \
            else ((1 + annual_target_return) ** (1/periods_per_annum) - 1)

        self._rolling_window_size = 0 if rolling_window_size is None or rolling_window_size < 0 else rolling_window_size
        maxlen = self._rolling_window_size if self._rolling_window_size > 0 else None
        self._returns = collections.deque(maxlen=maxlen)
        self._returns_benchmark = collections.deque(maxlen=maxlen)

        self._win_loss: core.WinLoss = core.WinLoss()
        self._capture: core.Capture = core.Capture()

        # In the following RawMomentKleinKBN constructors we use
        # `ddof=1, bias=True, fisher=True`.
        # This matches scipy's default behavior for kurtosis.

        self._returns_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)
        self._excess_returns_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)
        self._benchmark_returns_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)
        self._benchmark_excess_returns_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)

        #self._excess_covariance: core.CovarianceBullBear = core.CovarianceBullBear(ddof=1, threshold=self.risk_free_rate)
        self._sfm_regression: core.SFMRegression = core.SFMRegression(risk_free_rate=self.risk_free_rate)
        self._covariance: core.Covariance = core.Covariance(ddof=1, threshold=0)
        self._active_returns_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)
        self._active_returns_cumulative: core.CumulativeReturn = core.CumulativeReturn(window_size=self._rolling_window_size)

        self._target_returns_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)
        self._target_partial_moments: core.PartialMoments = core.PartialMoments(threshold=self.target_return)
        self._raw_partial_moments: core.RawPartialMoments = core.RawPartialMoments()
        self._benchmark_target_partial_moments: core.PartialMoments = core.PartialMoments(threshold=self.target_return)

        self._cumulative_return: core.CumulativeReturn = core.CumulativeReturn(window_size=self._rolling_window_size)
        self._cumulative_excess_return: core.CumulativeReturn = core.CumulativeReturn(window_size=self._rolling_window_size)
        self._benchmark_cumulative_return: core.CumulativeReturn = core.CumulativeReturn(window_size=self._rolling_window_size)

        self._drawdowns_cumulative = collections.deque(maxlen=maxlen)
        self._drawdowns_cumulative_minmax:core.MinMax = core.MinMax(window_size=self._rolling_window_size)
        self._drawdown_continuous_runs: core.ContinuousDrawdownRuns = core.ContinuousDrawdownRuns()
        self._drawdown_high_watermark: core.HighWaterMarkDrawdown = core.HighWaterMarkDrawdown(window_size=self._rolling_window_size)
        self._drawdown_high_watermark_benchmark: core.HighWaterMarkDrawdown = core.HighWaterMarkDrawdown(window_size=self._rolling_window_size)
        self._drawdown_episodes: core.DrawdownEpisodes = core.DrawdownEpisodes()
        self._drawdown_episodes_benchmark: core.DrawdownEpisodes = core.DrawdownEpisodes()

    def reset(self):
        """
        Reset all accumulated return data and derived streaming state.
    
        After ``reset()``, the instance behaves as if no returns have been
        added. Configuration parameters such as ``periods_per_annum``,
        ``annual_risk_free_rate``, ``annual_target_return``, and
        ``rolling_window_size`` are preserved.
        """
        #if self._rolling_window_size > 0: # ASSUME WE DON'T STORE RETURNS IF ROLLING WINDOW IS SET
        self._returns.clear()
        self._returns_benchmark.clear()

        self._win_loss.reset()
        self._capture.reset()

        self._returns_kbn.reset()
        self._excess_returns_kbn.reset()
        self._benchmark_returns_kbn.reset()
        self._benchmark_excess_returns_kbn.reset()

        #self._excess_covariance.reset()
        self._sfm_regression.reset()
        self._covariance.reset()
        self._active_returns_kbn.reset()
        self._active_returns_cumulative.reset()

        self._target_returns_kbn.reset()
        self._target_partial_moments.reset()
        self._raw_partial_moments.reset()
        self._benchmark_target_partial_moments.reset()

        self._cumulative_return.reset()
        self._cumulative_excess_return.reset()
        self._benchmark_cumulative_return.reset()

        self._drawdowns_cumulative.clear()
        self._drawdowns_cumulative_minmax.reset()
        self._drawdown_continuous_runs.reset()
        self._drawdown_high_watermark.reset()
        self._drawdown_high_watermark_benchmark.reset()
        self._drawdown_episodes.reset()
        self._drawdown_episodes_benchmark.reset()

    def add_return(self,ret: float, ret_bench: float):
        """
        Add one periodic portfolio and benchmark return observation.

        The supplied returns must already represent the same observation
        period and must use the periodicity implied by ``periods_per_annum``.

        No timestamp-based normalization or rescaling is performed. In
        particular, a monthly return is treated as one monthly observation
        regardless of the number of calendar days between its timestamps.

        The benchmark return is used by measures that compare portfolio
        performance with a benchmark, such as beta, alpha, tracking error,
        information ratio, and systematic-risk measures.

        Args:
            ret:
                Portfolio return for the current observation period, expressed
                as a decimal.

            ret_bench:
                Benchmark return for the same observation period, expressed
                as a decimal.

        Raises:
            ValueError:
                If either return is outside the valid domain required by a
                particular calculation.
        """
        if self._rolling_window_size > 0 and len(self._returns) == self._rolling_window_size:
            ret_old = self._returns.popleft()
            ret_bench_old = self._returns_benchmark.popleft()
            self._returns_kbn.revert(ret_old)
            # Excess returns (returns less risk-free rate)
            self._excess_returns_kbn.revert(ret_old - self.risk_free_rate)
            # Target returns (returns less target return)
            self._target_returns_kbn.revert(ret_old - self.target_return)
            self._target_partial_moments.revert(ret_old)
            self._raw_partial_moments.revert(ret_old)
            self._benchmark_target_partial_moments.revert(ret_bench_old)
            self._win_loss.revert(ret_old)
            self._capture.revert(ret_old, ret_bench_old)
            self._cumulative_return.revert(ret_old)
            self._cumulative_excess_return.revert(ret_old - self.risk_free_rate)
            self._benchmark_cumulative_return.revert(ret_bench_old)
            # Benchmarks
            self._benchmark_returns_kbn.revert(ret_bench_old)
            self._benchmark_excess_returns_kbn.revert(ret_bench_old - self.risk_free_rate)
            self._active_returns_kbn.revert(ret_old - ret_bench_old)
            self._active_returns_cumulative.revert(ret_old - ret_bench_old)
            self._covariance.revert(ret_old, ret_bench_old)
            self._sfm_regression.revert(ret_old, ret_bench_old)
            #self._excess_covariance.revert(ret_old, ret_bench_old)
            # Drawdowns
            # Note high watermark drawdown and drawdown episodes classes have no revert()
            self._drawdowns_cumulative.popleft()
            self._drawdown_continuous_runs.revert(ret_old) # Burke

        self._returns_kbn.update(ret)
        # Excess returns (returns less risk-free rate)
        ret_excess = ret - self.risk_free_rate
        self._excess_returns_kbn.update(ret_excess)
        # Target returns (returns less target return)
        self._target_returns_kbn.update(ret - self.target_return)
        self._target_partial_moments.update(ret)
        self._raw_partial_moments.update(ret)
        self._benchmark_target_partial_moments.update(ret_bench)
        self._win_loss.update(ret)
        self._capture.update(ret, ret_bench)
        # Benchmarks
        self._benchmark_returns_kbn.update(ret_bench)
        ret_bench_excess = ret_bench - self.risk_free_rate
        self._benchmark_excess_returns_kbn.update(ret_bench_excess)
        self._active_returns_kbn.update(ret - ret_bench)
        self._active_returns_cumulative.update(ret - ret_bench)
        self._covariance.update(ret, ret_bench)
        self._sfm_regression.update(ret, ret_bench)
        #self._excess_covariance.update(ret, ret_bench)

        self._returns.append(ret)
        self._returns_benchmark.append(ret_bench)

        # Cumulative return
        self._cumulative_return.update(ret)
        self._cumulative_excess_return.update(ret_excess)
        self._benchmark_cumulative_return.update(ret_bench)

        # Drawdowns from peaks to valleys, operates on cumulative returns
        dd = self._cumulative_return.geometric_return_plus_1 / self._cumulative_return.geometric_return_plus_1_max - 1
        self._drawdowns_cumulative.append(dd)
        self._drawdowns_cumulative_minmax.update(dd)

        # Drawdown calculation used in Burke
        self._drawdown_continuous_runs.update(ret)

        # High-water-mark drawdown and drawdown episodes
        hwm = self._drawdown_high_watermark
        recalculated = hwm.update(ret)
        if recalculated:
            self._drawdown_episodes.recalculate(hwm.drawdowns)
        else:
            self._drawdown_episodes.update(hwm.drawdown)
        hwm = self._drawdown_high_watermark_benchmark
        recalculated = hwm.update(ret_bench)
        if recalculated:
            self._drawdown_episodes_benchmark.recalculate(hwm.drawdowns)
        else:
            self._drawdown_episodes_benchmark.update(hwm.drawdown)

    @property
    def autocorrelation_penalty(self):
        """
        The autocorrelation penalty factor for serially correlated returns.

        The penalty factor is based on the variance inflation adjustment proposed
        by Andrew W. Lo ("The Statistics of Sharpe Ratios", 2002). It quantifies
        the extent to which positive serial correlation inflates risk-adjusted
        performance measures such as the Sharpe or Sortino ratio.

        The penalty factor is computed as::

            $$\\sqrt{1 + 2 * \\sum_{k=1}^{q-1}{((1 - k/q) * rho_k}}$$

        where ``rho_k`` is the sample autocorrelation at lag ``k`` and ``q`` is
        the annualization period (e.g. 252 for daily returns, 52 for weekly,
        12 for monthly, 4 for quarterly).

        Divide a Sharpe or Sortino ratio by this value to obtain an
        autocorrelation-adjusted ratio:

            adjusted_ratio = ratio / autocorrelation_penalty

        Returns
            A multiplicative penalty factor. A value of 1.0 indicates no detected
            serial correlation. Values greater than 1.0 indicate positive
            autocorrelation, which reduces the adjusted ratio. Values less than
            1.0 may occur for negatively autocorrelated returns.

        Notes
            The adjustment assumes equally spaced return observations and is intended
            for annualized risk-adjusted performance measures. If there are fewer than
            two observations or the return variance is zero, 1.0 is returned.

            The calculation depends on periods per annum.
        """
        n = self._returns_kbn.n
        if n < 2 or self._returns is None:
            return 1.0
        mean = self._returns_kbn.mean
        denom = self._returns_kbn.variance_ddof_0 * n
        if math.isnan(denom) or denom == 0:
            return 1.0

        # Lo's recommended aggregation period
        # dayly 252, weekly 52, monthly 12, quarterly 4
        q = min(self.periods_per_annum, n - 1)

        s = 0.0
        for k in range(1, q):
            numer = 0.0
            for t in range(k, n):
                numer += (self._returns[t] - mean) * (self._returns[t-k] - mean)
            rho = numer / denom
            s += (1.0 - k / q) * rho

        return math.sqrt(max(0.0, 1.0 + 2.0 * s))
    
    @property
    def cumulative_geometric_return(self):
        """
        Cumulative geometric return.
        """
        return self._cumulative_return.cumulative_geometric_return

    @property
    def geometric_mean_return(self):
        """
        The geometric mean of the returns, or geometric mean return per observation.

        It is the constant per-period return that would produce the same cumulative growth.
        """
        return self._cumulative_return.geometric_mean_return

    @property
    def compound_annual_growth_rate(self):
        """
        Compound Annual Growth Rate (CAGR) of the returns.

        Computed by annualizing the geometric mean return using
        periods per annum parameter.

        The calculation depends on periods per annum.
        """
        return self._cumulative_return.annualized_geometric_mean_return(self.periods_per_annum)

    @property
    def skewness(self):
        """
        The skewness of the distribution of returns is the
        degreeof asymmetry of a distribution around its mean.

        This "sci-py default" skewness is calculated as 'moment'
        (bias=True) skewness.
        """
        return self._returns_kbn.skewness

    @property
    def skewness_moment(self):
        """
        The 'moment' skewness of the distribution of returns,
        calculated as

        $$g_1=\\frac{\\mu_3}{\\mu_2^{3/2}}$$

        The skewness of the distribution of returns is the
        degreeof asymmetry of a distribution around its mean.
        """
        return self._returns_kbn.skewness_moment

    @property
    def skewness_fisher(self):
        """
        The 'fisher' skewness of the distribution of returns,
        calculated as

        $$g_1\\cdot \\frac{\\sqrt{n(n-1)}}{n-2}$$

        where

        $$g_1=\\frac{\\mu_3}{\\mu_2^{3/2}}$$

        The skewness of the distribution of returns is the
        degreeof asymmetry of a distribution around its mean.
        """
        return self._returns_kbn.skewness_fisher

    @property
    def skewness_sample(self):
        """
        The 'sample' skewness of the distribution of returns,
        calculated as

        $$g_1\\cdot \\frac{n^2}{(n-1)(n-2)}$$

        where

        $$g_1=\\frac{\\mu_3}{\\mu_2^{3/2}}$$

        The skewness of the distribution of returns is the
        degreeof asymmetry of a distribution around its mean.
        """
        return self._returns_kbn.skewness_sample

    @property
    def kurtosis(self):
        """
        The kurtosis of the distribution of returns is
        the degree to which a distribution peak compared
        to a normal distribution.

        This "sci-py default" is calculated as biased excess kurtosis.
        """
        return self._returns_kbn.kurtosis

    @property
    def kurtosis_excess(self):
        """
        The biased excess kurtosisof the distribution of returns,
        also known as 'excess'. Calculated as

        $$\\beta_2-3$$

        where

        $$\\beta_2=\\frac{\\mu_4}{\\mu_2^2}$$

        The kurtosis of the distribution of returns is
        the degree to which a distribution peak compared
        to a normal distribution.
        """
        return self._returns_kbn.kurtosis_excess

    @property
    def kurtosis_moment(self):
        """
        The biased Pearson (population) kurtosis of the distribution of returns,
        also known as 'moment'. Calculated as

        $$\\beta_2=\\frac{\\mu_4}{\\mu_2^2}$$

        The kurtosis of the distribution of returns is
        the degree to which a distribution peak compared
        to a normal distribution.
        """
        return self._returns_kbn.kurtosis_moment

    @property
    def kurtosis_sample_excess(self):
        """
        The unbiased excess kurtosis of the distribution of returns,
        also known as 'sample excess'. Calculated as

        $$\\frac{(n^2-1)\\beta_2-3(n - 1)^2}{(n - 2)(n - 3)}$$

        where

        $$\\beta_2=\\frac{\\mu_4}{\\mu_2^2}$$

        The kurtosis of the distribution of returns is
        the degree to which a distribution peak compared
        to a normal distribution.
        """
        return self._returns_kbn.kurtosis_sample_excess

    @property
    def kurtosis_sample_corrected(self):
        """
        The variant of unbiased Pearson kurtosis of the distribution of returns
        compatible with `PerformanceAnalytics` R package. Calculated as

        $$\\frac{(n^2 - 1)\\beta_2}{(n-2)(n-3)}$$

        where

        $$\\beta_2=\\frac{\\mu_4}{\\mu_2^2}$$

        The kurtosis of the distribution of returns is
        the degree to which a distribution peak compared
        to a normal distribution.
        """
        return self._returns_kbn.kurtosis_sample_corrected

    @property
    def kurtosis_sample(self):
        """
        The unbiased Pearson kurtosis of the distribution of returns
        also known as 'sample excess'. Calculated as

        $$\\frac{(n^2-1)\\beta_2-3(n - 1)^2}{(n - 2)(n - 3)}+3$$

        where

        $$\\beta_2=\\frac{\\mu_4}{\\mu_2^2}$$

        The kurtosis of the distribution of returns is
        the degree to which a distribution peak compared
        to a normal distribution.
        """
        return self._returns_kbn.kurtosis_sample

    @property
    def skewness_kurtosis_ratio(self) -> float:
        """
        Skewness-Kurtosis Ratio of the return distribution is
        the ratio of sample skewness to sample kurtosis.

        Positive values indicate that positive asymmetry dominates
        relative to tail heaviness, while negative values indicate
        downside asymmetry.

        This is a descriptive statistic of the return distribution
        rather than a standard risk-adjusted performance measure.

        It is used in conjunction with the Sharpe ratio to rank portfolios.
        Higher is better.
        """
        s = self._returns_kbn.skewness_moment
        k = self._returns_kbn.kurtosis_moment
        return s / k if k != 0 else math.nan

    @property
    def jarque_bera_normality_test_statistic(self) -> float:
        """
        Jarque–Bera normality test statistic.

        Tests the null hypothesis that the sample was drawn from a normal
        distribution using sample skewness and Fisher excess kurtosis.

        JB = n/6 * (skewness^2 + excess_kurtosis^2/4)

        Under the null hypothesis and for sufficiently large sample sizes,
        JB is asymptotically distributed as χ² with 2 degrees of freedom.

        Typical critical values:
            α = 0.10 : 4.61 (90% confidence)
            α = 0.05 : 5.99 (95% confidence)
            α = 0.01 : 9.21 (99% confidence)

        Returns only the test statistic; p-values can be obtained from the
        χ²(2) distribution.
        
        If BJ > critical value, reject normality hypothesis.
        """
        # Use population skewness and excess kurtosis (not sample corrected)
        # These match the scipy.stats skew and kurtosis with bias=True, fisher=True
        s = self._returns_kbn.skewness_moment
        k = self._returns_kbn.kurtosis_excess
        if math.isnan(s) or math.isnan(k):
            return math.nan
        # Bera-Jarque formula (Equation 5.17 from Bacon 3rd ed)
        n = self._returns_kbn.n
        return  (n / 6) * (s*s + (k*k) / 4)

    def is_normal_distribution(self, confidence: float = 0.95) -> bool:
        """
        Test the null hypothesis that returns are normally distributed
        using the Jarque–Bera test.

        Args:
            confidence: Confidence level in (0, 1): e.g., 0.95 or 0.99

        Returns:
            True  : normality cannot be rejected.
            False : normality is rejected  or insufficient data to perform the test.
        """
        bj = self.jarque_bera_normality_test_statistic
        if math.isnan(bj):
            return False
        if confidence <= 0 or confidence >= 1:
            raise ValueError("confidence must be between 0 and 1")

        # Pure Python inverse CDF for Chi-Squared distribution with 2 degrees of freedom.
        #
        # Formula: -2 * ln(1 - p)
        #
        # Implementing scipy.stats.chi2.ppf (Percent Point Function, or inverse CDF)
        # in pure Python without SciPy is possible but complex. It requires:
        # - Implementing the Incomplete Gamma Function (to calculate the CDF).
        # - Using a root-finding algorithm (like bisection or Newton-Raphson) to invert the CDF. 
        #
        # For 2 degrees of freedom, there is a mathematical shortcut:
        # the Chi-Squared distribution with df=2 is equivalent to an Exponential distribution
        # with λ=0.5. The inverse CDF has a closed-form solution: chi2.ppf(p,2)=−2ln(1−p)
        #
        # Verification:
        # −2ln(1 − 0.95) -> 5.991464547107982
        # −2ln(1 − 0.99) -> 9.210340371976184
        critical = -2.0 * math.log1p(-confidence)
        return bj <= critical

    def var_historical(self, confidence: float = 0.95) -> float:
        """
        Estimate historical Value at Risk (VaR) from observed returns.

        Historical VaR is a non-parametric estimate of the loss threshold that is
        exceeded with probability 1 - confidence, based solely on the empirical
        distribution of observed returns. It is computed as the negative lower-tail
        percentile of the return distribution:

            VaR = -percentile(returns, 1 - confidence)

        Args:
            confidence: Confidence level in the range (0, 1), for example
                0.95 for 95% VaR.

        Returns:
            The estimated loss as a positive value. If the lower-tail percentile
            is positive (all observed returns exceed the loss threshold), the
            result is negative, indicating "inverse risk". Callers that prefer
            missing values may replace negative results with ``math.nan``.
        """
        return core.var_historical(self._returns, risk_free_rate=0.0, confidence=confidence)

    def var_gaussian(self, confidence: float = 0.95) -> float:
        """
        Estimate Gaussian (parametric) Value at Risk (VaR).

        Gaussian VaR assumes returns follow a normal distribution and estimates
        the loss threshold exceeded with probability 1 - confidence from the
        sample mean and standard deviation:

            VaR = -(mean + z * std)

        where z is the standard normal quantile corresponding to
        1 - confidence.

        Args:
            confidence: Confidence level in the range (0, 1), for example
                0.95 for 95% VaR.

        Returns:
            The estimated loss as a positive value. If the estimate is
            negative (the lower-tail quantile is positive), the result
            indicates "inverse risk". Callers that prefer missing values
            may replace negative results with ``math.nan``.
        """
        return core.var_gaussian(returns_kbn=self._returns_kbn, confidence=confidence)

    def var_cornish_fisher(self, confidence: float = 0.95) -> float:
        """
        Estimate modified Cornish-Fisher Value at Risk (VaR).

        Cornish-Fisher VaR extends Gaussian VaR by adjusting the normal
        quantile for the observed skewness and excess kurtosis of returns.
        This generally provides a better approximation for non-normal return
        distributions while retaining a closed-form estimate.

        Args:
            confidence: Confidence level in the range (0, 1), for example
                0.95 for 95% VaR.

        Returns:
            The estimated loss as a positive value. If the estimate is
            negative (the lower-tail quantile is positive), the result
            indicates "inverse risk". Callers that prefer missing values
            may replace negative results with ``math.nan``.
        """
        return core.var_cornish_fisher(returns_kbn=self._returns_kbn, confidence=confidence)

    def es_historical(self, confidence: float = 0.95) -> float:
        """
        Estimate historical Expected Shortfall (ES) from observed returns.

        Historical ES is a non-parametric estimate of the average loss occurring
        in the worst (1 - confidence) fraction of observations. It is computed
        as the negative mean of all returns less than or equal to the historical
        VaR threshold.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% ES.

        Returns:
            Estimated expected loss beyond the VaR threshold. The result may be
            negative when all observed returns are positive, indicating
            "inverse risk".
        """
        return core.es_historical(self._returns, risk_free_rate=0.0, confidence=confidence)

    def es_gaussian(self, confidence: float = 0.95) -> float:
        """
        Estimate Gaussian Expected Shortfall (ES).

        Gaussian ES assumes returns follow a normal distribution characterized
        by the sample mean and standard deviation. It estimates the expected
        loss conditional on losses exceeding the Gaussian VaR threshold.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% ES.

        Returns:
            Estimated expected loss. The result may be negative for
            predominantly positive returns.
        """
        return core.es_gaussian(returns_kbn=self._returns_kbn, confidence=confidence)

    def es_cornish_fisher(self, confidence: float = 0.95) -> float:
        """
        Estimate Cornish-Fisher Expected Shortfall (ES).

        This method extends Gaussian ES by adjusting for sample skewness and
        excess kurtosis using the Cornish-Fisher expansion, producing a
        parametric estimate that better reflects asymmetric and heavy-tailed
        return distributions.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% ES.

        Returns:
            Estimated expected loss using the operational Cornish-Fisher
            approximation.
        """
        return core.es_cornish_fisher(returns_kbn=self._returns_kbn, confidence=confidence)

    def reward_to_var_ratio_historical(self, confidence: float = 0.95) -> float:
        """
        Calculate the reward-to-historical-VaR ratio.

        The reward-to-VaR ratio measures mean excess return relative to
        historical Value at Risk (VaR). The numerator is the mean return
        in excess of the risk-free rate, while the denominator is the
        historical VaR at the specified confidence level.

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% VaR.

        Returns:
            Mean excess return divided by historical VaR, or ``math.nan``
            if the VaR is zero.
        """
        # Use excess return over risk-free rate
        denom = self.var_historical(confidence=confidence)
        return self._excess_returns.mean / denom if denom != 0 else math.nan

    def reward_to_var_ratio_gaussian(self, confidence: float = 0.95) -> float:
        """
        Calculate the reward-to-Gaussian-VaR ratio.

        The reward-to-VaR ratio measures mean excess return relative to
        Gaussian Value at Risk (VaR). The numerator is the mean return
        in excess of the risk-free rate, while the denominator is the
        Gaussian VaR estimated under a normal-return assumption at the
        specified confidence level.

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% VaR.

        Returns:
            Mean excess return divided by Gaussian VaR, or ``math.nan``
            if the VaR is zero.
        """
        # Use excess return over risk-free rate
        denom = self.var_gaussian(confidence=confidence)
        return self._excess_returns.mean / denom if denom != 0 else math.nan

    def reward_to_var_ratio_cornish_fisher(self, confidence: float = 0.95) -> float:
        """
        Calculate the reward-to-Cornish-Fisher-VaR ratio.

        The reward-to-VaR ratio measures mean excess return relative to
        Cornish-Fisher Value at Risk (VaR). The numerator is the mean
        return in excess of the risk-free rate, while the denominator is
        the Cornish-Fisher VaR, which adjusts the Gaussian estimate for
        sample skewness and excess kurtosis.

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% VaR.

        Returns:
            Mean excess return divided by Cornish-Fisher VaR, or
            ``math.nan`` if the VaR is zero.
        """
        # Use excess return over risk-free rate
        denom = self.var_cornish_fisher(confidence=confidence)
        return self._excess_returns.mean / denom if denom != 0 else math.nan

    def reward_to_es_ratio_historical(self, confidence: float = 0.95) -> float:
        """
        Calculate the reward-to-historical-ES ratio.

        The reward-to-ES ratio measures mean excess return relative to
        historical Expected Shortfall (ES). The numerator is the mean
        return in excess of the risk-free rate, while the denominator is
        the historical ES at the specified confidence level.

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% ES.

        Returns:
            Mean excess return divided by historical ES, or ``math.nan``
            if the ES is zero.
        """
        # Use excess return over risk-free rate
        denom = self.es_historical(confidence=confidence)
        return self._excess_returns.mean / denom if denom != 0 else math.nan

    def reward_to_es_ratio_gaussian(self, confidence: float = 0.95) -> float:
        """
        Calculate the reward-to-Gaussian-ES ratio.

        The reward-to-ES ratio measures mean excess return relative to
        Gaussian Expected Shortfall (ES). The numerator is the mean
        return in excess of the risk-free rate, while the denominator is
        the Gaussian ES estimated under a normal-return assumption at
        the specified confidence level.

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% ES.

        Returns:
            Mean excess return divided by Gaussian ES, or ``math.nan``
            if the ES is zero.
        """
        # Use excess return over risk-free rate
        denom = self.es_gaussian(confidence=confidence)
        return self._excess_returns.mean / denom if denom != 0 else math.nan

    def reward_to_es_ratio_cornish_fisher(self, confidence: float = 0.95) -> float:
        """
        Calculate the reward-to-Cornish-Fisher-ES ratio.

        The reward-to-ES ratio measures mean excess return relative to
        Cornish-Fisher Expected Shortfall (ES). The numerator is the mean
        return in excess of the risk-free rate, while the denominator is
        the Cornish-Fisher ES, which adjusts the Gaussian estimate for
        sample skewness and excess kurtosis.

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% ES.

        Returns:
            Mean excess return divided by Cornish-Fisher ES, or ``math.nan``
            if the ES is zero.
        """
        # Use excess return over risk-free rate
        denom = self.es_cornish_fisher(confidence=confidence)
        return self._excess_returns.mean / denom if denom != 0 else math.nan

    @property
    def mean_absolute_deviation_ratio(self) -> float:
        """
        Estimate the mean absolute deviation ratio.

        The mean absolute deviation (MAD) ratio measures the mean return
        relative to the average absolute deviation of returns from their
        sample mean. It is calculated as:

            MAD Ratio = Mean Return / Mean Absolute Deviation

        The mean absolute deviation is the average of the absolute
        differences between each return and the sample mean. Unlike
        standard deviation, MAD does not square deviations and therefore
        is less sensitive to extreme observations.

        Returns:
            The mean return divided by the mean absolute deviation, or
            ``math.nan`` if no returns have been observed or if the mean
            absolute deviation is zero.
        """
        n = self._returns_kbn.n
        w = self._returns
        if w is None or n < 1:
            return math.nan
        mean = self._returns_kbn.mean

        # Calculate Mean Absolute Deviation: sum(|x - mean|) / n
        # Use Klein second-order Kahan-Babuška-Neumaier (KBN)
        # floating-point accumulator.
        sum: KleinKBNAccumulator = KleinKBNAccumulator()
        for x in w:
            sum.update(abs(x - mean))

        mad = sum.value / n
        return mean / mad if mad > 0 else math.nan

    @property
    def upside_potential_ratio(self) -> float:
        """
        Estimate the upside potential ratio.

        The upside potential ratio measures upside potential relative to
        downside risk. It is calculated as the first-order upper partial
        moment about the minimum acceptable return (MAR), divided by the
        square root of the second-order lower partial moment about MAR.

        The numerator captures the average upside above MAR, while the
        denominator measures downside volatility relative to MAR. Unlike
        ratios based on total volatility, this measure focuses separately
        on favorable returns and downside risk.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The upside potential ratio, or ``math.nan`` if the required
            partial moments are unavailable or downside risk is zero.
        """
        hpm1 = self._target_partial_moments.higher_partial_moment_1
        lpm2 = self._target_partial_moments.lower_partial_moment_2
        if math.isnan(hpm1) or math.isnan(lpm2) or lpm2 == 0:
            return math.nan
        return hpm1 / math.sqrt(lpm2)

    @property
    def upside_potential_ratio_subset(self) -> float:
        """
        Estimate the subset upside potential ratio.

        The subset upside potential ratio is calculated like
        ``upside_potential_ratio``, except that both the upside potential
        and downside risk are normalized using only observations that
        contribute to their respective partial moments.

        The numerator is the average excess return among observations
        above the minimum acceptable return (MAR). The denominator is the
        square root of the average squared shortfall among observations
        below MAR.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The subset upside potential ratio, or ``math.nan`` if there are
            no upside or downside observations, or if downside risk is zero.
        """
        n1 = self._target_partial_moments.upper_excess_count
        n2 = self._target_partial_moments.lower_excess_count
        if n1 == 0 or n2 == 0:
            return math.nan
        hpm1 = self._target_partial_moments.upper_excess_moment_1_sum / n1
        lpm2 = self._target_partial_moments.lower_excess_moment_2_sum / n2
        if math.isnan(hpm1) or math.isnan(lpm2) or lpm2 == 0:
            return math.nan
        return hpm1 / math.sqrt(lpm2)

    @property
    def upside_frequency(self) -> float:
        """
        Estimate the upside frequency.

        Upside frequency is the proportion of observed returns above the
        minimum acceptable return (MAR). It measures how often the
        investment exceeds the target return.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            A value between 0 and 1, where 0 indicates that no returns
            exceed MAR and 1 indicates that every return does.
        """
        return self._target_partial_moments.upside_frequency

    @property
    def upside_potential(self) -> float:
        """
        Estimate the upside potential.

        Upside potential is the first-order upper partial moment about the
        minimum acceptable return (MAR). It measures the average upside
        relative to MAR, with observations below MAR contributing zero.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The average upside above MAR, or ``math.nan`` if no returns
            have been observed.
        """
        return self._target_partial_moments.higher_partial_moment_1

    @property
    def upside_potential_subset(self) -> float:
        """
        Estimate the subset upside potential.

        Subset upside potential is the average excess return above the
        minimum acceptable return (MAR), normalized only by observations
        that exceed MAR. Consequently, it measures the average magnitude
        of upside observations only.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The subset upside potential, or ``0.0`` when no observations
            exceed MAR.
        """
        n = self._target_partial_moments.upper_excess_count
        return self._target_partial_moments.upper_excess_moment_1_sum / n if n > 0 else 0

    @property
    def upside_variance(self) -> float:
        """
        Estimate the upside variance.

        Upside variance is the second-order upper partial moment about the
        minimum acceptable return (MAR). It measures the squared magnitude
        of returns above MAR, normalized by the total number of observations.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The upside variance, or ``math.nan`` if no returns have been
            observed.
        """
        return self._target_partial_moments.higher_partial_moment_2

    @property
    def upside_variance_subset(self) -> float:
        """
        Estimate the subset upside variance.

        Subset upside variance is the second-order upper partial moment
        about the minimum acceptable return (MAR), normalized only by
        observations that exceed MAR. Consequently, it measures the
        average squared magnitude of upside observations only.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The subset upside variance, or ``0.0`` when no observations
            exceed MAR.
        """
        n = self._target_partial_moments.upper_excess_count
        return self._target_partial_moments.upper_excess_moment_2_sum / n if n > 0 else 0

    @property
    def upside_risk(self) -> float:
        """
        Estimate the upside risk.

        Upside risk is the standard deviation corresponding to the upside
        variance. It is the square root of the second-order upper partial
        moment about the minimum acceptable return (MAR).

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The upside risk, or ``math.nan`` if upside variance is
            unavailable.
        """
        variance = self.upside_variance
        return math.nan if math.isnan(variance) else math.sqrt(variance)

    @property
    def upside_risk_subset(self) -> float:
        """
        Estimate the subset upside risk.

        Subset upside risk is the standard deviation corresponding to the
        subset upside variance. It is calculated using only observations
        that exceed the minimum acceptable return (MAR).

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The subset upside risk, or ``math.nan`` if the subset upside
            variance is unavailable.
        """
        variance = self.upside_variance_subset
        return math.nan if math.isnan(variance) else math.sqrt(variance)

    @property
    def semi_deviation(self) -> float:
        """
        Estimate the semi-deviation of returns.

        Semi-deviation measures downside volatility by considering only
        returns below the sample mean. It is the square root of the
        second-order lower partial moment about the mean and is commonly
        used as the denominator of the Sortino (Downside Sharpe) ratio.

        Returns:
            The sample semi-deviation, or ``math.nan`` if no returns
            have been observed.
        """
        n = self._returns_kbn.n
        returns = self._returns
        if returns is None or n == 0:
            return math.nan

        mean = self._returns_kbn.mean

        # Accumulate squared negative deviations from the mean using a
        # compensated floating-point summation.
        sum_squared: KleinKBNAccumulator = KleinKBNAccumulator()
        for r in returns:
            deviation  = r - mean
            if deviation  < 0:
                sum_squared.update(deviation * deviation )
        # Divide by "full" window length
        return math.sqrt(sum_squared.value / n)

    @property
    def downside_deviation(self) -> float:
        """
        Estimate the downside deviation of returns.

        Downside deviation measures downside volatility relative to the
        minimum acceptable return (MAR). It is the square root of the
        second-order lower partial moment about MAR, normalized by the
        total number of observations.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The downside deviation, or ``math.nan`` if no returns
            have been observed.
        """
        denom = self._target_partial_moments.total_count
        if denom == 0:
            return math.nan
        return math.sqrt(self._target_partial_moments.lower_excess_moment_2_sum / denom)

    @property
    def downside_deviation_subset(self) -> float:
        """
        Estimate the subset downside deviation of returns.

        This statistic is identical to ``downside_deviation`` except that
        only returns below the minimum acceptable return (MAR) contribute
        to the normalization. Consequently, it measures the average
        magnitude of downside observations only.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The subset downside deviation. Returns ``0.0`` when no
            observations fall below MAR.
        """
        denom = self._target_partial_moments.lower_excess_count     
        if denom == 0:
            return 0
        return math.sqrt(self._target_partial_moments.lower_excess_moment_2_sum / denom)

    @property
    def downside_frequency(self) -> float:
        """
        Estimate the downside frequency.

        Downside frequency is the proportion of observed returns below
        the minimum acceptable return (MAR). It measures how often the
        investment fails to achieve the target return.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            A value between 0 and 1, where 0 indicates that no returns
            fall below MAR and 1 indicates that every return does.
        """
        return self._target_partial_moments.downside_frequency

    @property
    def downside_potential(self) -> float:
        """
        Estimate the downside potential.

        Downside potential is the average shortfall below the minimum
        acceptable return (MAR). Unlike downside deviation, it measures
        the average magnitude of losses relative to MAR rather than their
        squared magnitude.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The average downside shortfall, or ``math.nan`` if no returns
            have been observed.
        """
        return self._target_partial_moments.downside_potential

    @property
    def sharpe_ratio(self) -> float:
        """
        The Sharpe ratio is the mean excess return divided by the sample
        standard deviation (ddof=1) of excess returns:

        $$\\mathrm{SR} = \\frac{\\overline{R - R_f}}{\\sigma_{R-R_f}}\\,$$

        The risk-free rate is specified when constructing the ``Measures`` object.

        Returns:
            The Sharpe ratio of excess returns, or ``math.nan`` if fewer
            than two observations are available or the standard deviation
            is zero.
        """
        std = self._excess_returns_kbn.standard_deviation_ddof_1
        if math.isnan(std) or std == 0:
            return math.nan
        return self._excess_returns_kbn.mean / std

    def sharpe_ratio_var_historical(self, confidence: float = 0.95) -> float:
        """
        Modified Sharpe Ratio using historical Value-at-Risk as the measure of risk.

        The denominator is the empirical lower-tail VaR of excess returns
        at the specified confidence level:
        $$\\frac{\\overline{R-R_f}}{\\operatorname{VaR}_{\\text{hist}}(R-R_f)}\\,$$

        Internally, VaR is computed on losses from excess returns to ensure
        a positive risk measure.

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% VaR.

        Returns:
            The modified Sharpe ratio, or ``math.nan`` if insufficient data
            or if the VaR is zero.
        """
        if self._excess_returns_kbn.n < 2:
            return math.nan
        denom = core.var_historical(self._returns, risk_free_rate=self.risk_free_rate, confidence=confidence)
        if math.isnan(denom) or denom == 0:
            return math.nan
        return self._excess_returns_kbn.mean / denom

    def sharpe_ratio_var_gaussian(self, confidence: float = 0.95) -> float:
        """
        Modified Sharpe Ratio using Gaussian Value-at-Risk as the measure of risk.

        The denominator is the parametric Gaussian VaR of excess returns
        at the specified confidence level:
        $$\\frac{\\overline{R-R_f}}{\\operatorname{VaR}_{\\mathcal N}(R-R_f)}\\,$$

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% VaR.

        Returns:
            The modified Sharpe ratio, or ``math.nan`` if insufficient data
            or if the VaR is zero.
        """
        if self._excess_returns_kbn.n < 2:
            return math.nan
        denom = core.var_gaussian(returns_kbn=self._excess_returns_kbn, confidence=confidence)
        if math.isnan(denom) or denom == 0:
            return math.nan
        return self._excess_returns_kbn.mean / denom

    def sharpe_ratio_var_cornish_fisher(self, confidence: float = 0.95) -> float:
        """
        Modified Sharpe Ratio using Cornish-Fisher Value-at-Risk as the measure of risk.

        The denominator is the Cornish–Fisher VaR of excess returns, which
        adjusts the Gaussian quantile for sample skewness and excess kurtosis:
        $$\\frac{\\overline{R-R_f}}{\\operatorname{VaR}_{\\text{CF}}(R-R_f)}\\,$$

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% VaR.

        Returns:
            The modified Sharpe ratio, or ``math.nan`` if insufficient data
            or if the VaR is zero.
        """
        if self._excess_returns_kbn.n < 2:
            return math.nan
        denom = core.var_cornish_fisher(returns_kbn=self._excess_returns_kbn, confidence=confidence)
        if math.isnan(denom) or denom == 0:
            return math.nan
        return self._excess_returns_kbn.mean / denom

    def sharpe_ratio_es_historical(self, confidence: float = 0.95) -> float:
        """
        Modified Sharpe Ratio using historical Expected Shortfall (ES) as the measure of risk.

        The denominator is the empirical ES (a.k.a. Conditional VaR) of
        excess returns at the specified confidence level:
        $$\\frac{\\overline{R-R_f}}{\\operatorname{ES}_{\\text{hist}}(R-R_f)}\\,$$

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% ES.

        Returns:
            The modified Sharpe ratio, or ``math.nan`` if insufficient data
            or if the ES is zero.
        """
        if self._excess_returns_kbn.n < 2:
            return math.nan
        denom = core.es_historical(self._returns, risk_free_rate=self.risk_free_rate, confidence=confidence)
        if math.isnan(denom) or denom == 0:
            return math.nan
        return self._excess_returns_kbn.mean / denom

    def sharpe_ratio_es_gaussian(self, confidence: float = 0.95) -> float:
        """
        Modified Sharpe Ratio using Gaussian Expected Shortfall (ES) as the measure of risk.

        The denominator is the parametric Gaussian ES of excess returns at
        the specified confidence level:
        $$\\frac{\\overline{R-R_f}}{\\operatorname{ES}_{\\mathcal N}(R-R_f)}\\,$$

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% ES.

        Returns:
            The modified Sharpe ratio, or ``math.nan`` if insufficient data
            or if the ES is zero.
        """
        if self._excess_returns_kbn.n < 2:
            return math.nan
        denom = core.es_gaussian(returns_kbn=self._excess_returns_kbn, confidence=confidence)
        if math.isnan(denom) or denom == 0:
            return math.nan
        return self._excess_returns_kbn.mean / denom

    def sharpe_ratio_es_cornish_fisher(self, confidence: float = 0.95) -> float:
        """
        Modified Sharpe Ratio using Cornish-Fisher Expected Shortfall (ES) as the measure of risk.

        The denominator is the Cornish–Fisher ES of excess returns, which
        adjusts for sample skewness and excess kurtosis:
        $$\\frac{\\overline{R-R_f}}{\\operatorname{ES}_{\\text{CF}}(R-R_f)}\\,$$

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            confidence:
                Confidence level in the range (0, 1), for example
                0.95 for 95% ES.

        Returns:
            The modified Sharpe ratio, or ``math.nan`` if insufficient data
            or if the ES is zero.
        """
        if self._excess_returns_kbn.n < 2:
            return math.nan
        denom = core.es_cornish_fisher(returns_kbn=self._excess_returns_kbn, confidence=confidence)
        if math.isnan(denom) or denom == 0:
            return math.nan
        return self._excess_returns_kbn.mean / denom

    @property
    def downside_sharpe_ratio(self) -> float:
        """
        Estimate the symmetric downside-risk Sharpe ratio (Ziemba, 2005).

        Mean excess return divided by sqrt(2) times the semideviation of
        returns relative to their sample mean:
        $$\\mathrm{DSR} = \\frac{\\overline{R - R_f}}{\\sqrt{2}\\,\\operatorname{SemiSD}(R)}\\,$$
        where
        $$\\operatorname{SemiSD}(R) = \\sqrt{\\frac{1}{n}\\sum_{t=1}^n \\bigl(\\min(R_t - \\bar R, 0)\\bigr)^2}\\,$$

        The risk-free rate is specified when constructing the ``Measures`` object.

        Returns:
            The downside Sharpe ratio, ``math.nan`` if no returns are available,
            or ``\\+\\/\\-math.inf`` if semideviation is zero (sign follows the
            mean excess return).

        Notes:
            Reference: Ziemba, W. T. (2005). The symmetric downside-risk Sharpe ratio.
            The Journal of Portfolio Management, 32(1), 108–122.
        """
        # Semi-deviation (downside deviation with MAR = mean)
        semi_dev = self.semi_deviation
        if math.isnan(semi_dev):
            return math.nan
        if semi_dev == 0:
            return -math.inf if self._excess_returns_kbn.mean < 0 else math.inf
        # Use excess return over risk-free rate
        return self._excess_returns_kbn.mean / (_SQRT2 * semi_dev)

    @property
    def adjusted_sharpe_ratio(self) -> float:
        """
        Estimate the adjusted Sharpe ratio (Pezier & White, 2006).

        This adjustment accounts for skewness and excess kurtosis of the
        return distribution:
        $$\\mathrm{ASR}=\\mathrm{SR}\\Bigl(1+\\tfrac{S}{6}\\,\\mathrm{SR}-\\tfrac{\\kappa}{24}\\,\\mathrm{SR}^2\\Bigr)$$
        where ``S`` is sample skewness and ``κ`` is excess kurtosis.

        The risk-free rate is specified when constructing the ``Measures`` object.

        Returns:
            The adjusted Sharpe ratio, or ``math.nan`` if moments or SR cannot
            be estimated.

        Notes:
            Moments are computed from raw returns; SR is computed from excess returns.
            Reference: Pezier, J., & White, A. (2006). The relative merits of
            risk measures. ICMA Centre Discussion Papers.
        """
        skewness = self._returns_kbn.skewness_moment
        kurtosis = self._returns_kbn.kurtosis_excess
        if math.isnan(skewness) or math.isnan(kurtosis):
            return math.nan
        # Use standard Sharpe ratio (annualized logic not needed here as we work with period returns)
        sr = self.sharpe_ratio
        if math.isnan(sr):
            return math.nan
        # Adjusted Sharpe formula
        return sr * (1 + skewness * sr / 6 - kurtosis * sr * sr / 24)

    @property
    def adjusted_sharpe_ratio_skew_only(self) -> float:
        """
        Estimate the skewness-only adjusted Sharpe ratio.

        This variant adjusts the Sharpe ratio only for skewness and ignores
        kurtosis. Using excess kurtosis κ, the full adjustment is
        ASR = SR * (1 + (S/6)*SR - (κ/24)*SR^2). Setting κ = 0 yields:
        $$\\mathrm{ASR}_{\\text{skew-only}}=\\mathrm{SR}\\Bigl(1+\\tfrac{S}{6}\\,\\mathrm{SR}\\Bigr)$$

        The risk-free rate is specified when constructing the ``Measures`` object.

        Returns:
            The skew-only adjusted Sharpe ratio, or ``math.nan`` if skewness
            or SR cannot be estimated.
        """
        s = self._returns_kbn.skewness_moment
        if math.isnan(s):
            return math.nan
        sr = self.sharpe_ratio
        if math.isnan(sr):
            return math.nan
        return sr * (1 + (s * sr) / 6)

    def probabilistic_sharpe_ratio(self, reference_sr: float = 0.0):
        """
        Probabilistic Sharpe Ratio (PSR) with sample skewness and normal kurtosis.

        Computes the probability that the true Sharpe ratio exceeds a
        reference threshold using the skewness/kurtosis-adjusted sampling
        distribution:
        $$
        \\mathrm{PSR}(\\mathrm{SR}^\\*) = \\Phi\\!\\left(
        \\frac{(\\widehat{\\mathrm{SR}} - \\mathrm{SR}^\\*)\\sqrt{n-1}}
        {\\sqrt{\\,1 - S\\,\\widehat{\\mathrm{SR}} + \\tfrac{K-1}{4}\\,\\widehat{\\mathrm{SR}}^2\\,}}
        \\right)
        $$
        where ``S`` is sample skewness and ``K`` is set to 3 (normal kurtosis).

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            reference_sr:
                Benchmark Sharpe ratio ``SR*`` against which to evaluate.

        Returns:
            The PSR in [0, 1], or ``math.nan``` if it cannot be estimated.

        Notes:
            Based on Bailey & López de Prado (2014) and Opdyke (2007).
        """
        return core.probabilistic_sharpe_ratio(self._returns_kbn, self.sharpe_ratio,
            reference_sr=reference_sr, zero_skewness=False, normal_kurtosis=True)

    def probabilistic_sharpe_ratio_full(self, reference_sr: float = 0.0):
        """
        Probabilistic Sharpe Ratio (PSR) with sample skewness and sample kurtosis.

        Uses both sample skewness and sample kurtosis in the PSR formula:
        $$
        \\mathrm{PSR}(\\mathrm{SR}^\\*) = \\Phi\\!\\left(
        \\frac{(\\widehat{\\mathrm{SR}} - \\mathrm{SR}^\\*)\\sqrt{n-1}}
        {\\sqrt{\\,1 - S\\,\\widehat{\\mathrm{SR}} + \\tfrac{K-1}{4}\\,\\widehat{\\mathrm{SR}}^2\\,}}
        \\right)
        $$
        where ``S`` and ``K`` are the sample skewness and sample kurtosis (not
        excess kurtosis) of excess returns.

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            reference_sr:
                Benchmark Sharpe ratio ``SR*`` against which to evaluate.

        Returns:
            The PSR in [0, 1], or ``math.nan`` if it cannot be estimated.

        Notes:
            Based on Bailey & López de Prado (2014) and Opdyke (2007).
        """
        return core.probabilistic_sharpe_ratio(self._returns_kbn, self.sharpe_ratio,
            reference_sr=reference_sr, zero_skewness=False, normal_kurtosis=False)

    def probabilistic_sharpe_ratio_symmetric(self, reference_sr: float = 0.0):
        """
        Probabilistic Sharpe Ratio (PSR) with zero skewness and sample kurtosis.

        Assumes symmetric excess returns ``S=0`` but retains sample kurtosis:
        $$
        \\mathrm{PSR}(\\mathrm{SR}^\\*) = \\Phi\\!\\left(
        \\frac{(\\widehat{\\mathrm{SR}} - \\mathrm{SR}^\\*)\\sqrt{n-1}}
        {\\sqrt{\\,1 - S\\,\\widehat{\\mathrm{SR}} + \\tfrac{K-1}{4}\\,\\widehat{\\mathrm{SR}}^2\\,}}
        \\right)
        $$
        with ``S=0`` and ``K`` the sample kurtosis.

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            reference_sr:
                Benchmark Sharpe ratio ``SR*`` against which to evaluate.

        Returns:
            The PSR in [0, 1], or ``math.nan`` if it cannot be estimated.

        Notes:
            Based on Bailey & López de Prado (2014) and Opdyke (2007).
        """
        return core.probabilistic_sharpe_ratio(self._returns_kbn, self.sharpe_ratio,
            reference_sr=reference_sr, zero_skewness=True, normal_kurtosis=False)

    def probabilistic_sharpe_ratio_gaussian(self, reference_sr: float = 0.0):
        """
        Probabilistic Sharpe Ratio (PSR) under Gaussian moments.

        Assumes zero skewness and normal kurtosis ``K=3``:
        $$
        \\mathrm{PSR}(\\mathrm{SR}^\\*) = \\Phi\\!\\left(
        \\frac{(\\widehat{\\mathrm{SR}} - \\mathrm{SR}^\\*)\\sqrt{n-1}}
        {\\sqrt{\\,1 - S\\,\\widehat{\\mathrm{SR}} + \\tfrac{K-1}{4}\\,\\widehat{\\mathrm{SR}}^2\\,}}
        \\right)
        $$
        with ``S=0`` and ``K=3``.

        The risk-free rate is specified when constructing the ``Measures`` object.

        Args:
            reference_sr:
                Benchmark Sharpe ratio ``SR*`` against which to evaluate.

        Returns:
            The PSR in [0, 1], or ```math.nan`` if it cannot be estimated.

        Notes:
            Based on Bailey & López de Prado (2014) and Opdyke (2007).
        """
        return core.probabilistic_sharpe_ratio(self._returns_kbn, self.sharpe_ratio,
            reference_sr=reference_sr, zero_skewness=True, normal_kurtosis=True)

    @property
    def sortino_ratio(self) -> float:
        """
        Estimate the Sortino ratio.

        The Sortino ratio measures the arithmetic mean excess return
        relative to downside risk. The numerator is the mean return in
        excess of the minimum acceptable return (MAR), while the
        denominator is the square root of the second-order lower partial
        moment (LPM2) about MAR.

        Unlike the Sharpe ratio, the Sortino ratio does not use the
        risk-free rate as its return threshold. Instead, downside risk is
        measured relative to the MAR, so only returns below the target
        contribute to the denominator.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The arithmetic mean excess return over MAR divided by
            downside deviation, or ``math.nan`` if the downside partial
            moment is unavailable or zero.
        """
        lpm2 = self._target_partial_moments.lower_partial_moment_2
        if math.isnan(lpm2) or lpm2 == 0:
            return math.nan
        return self._target_returns_kbn.mean / math.sqrt(lpm2)

    @property
    def sortino_ratio_sqrt2(self) -> float:
        """
        Estimate the adjusted Sortino ratio proposed by Jack Schwager.

        The adjusted Sortino ratio is the standard Sortino ratio divided
        by the square root of two:

            Adjusted Sortino Ratio = Sortino Ratio / sqrt(2)

        This normalization is intended to make the Sortino ratio more
        directly comparable with the Sharpe ratio. Under symmetric return
        distributions, downside deviation is related to standard
        deviation by a factor of approximately ``sqrt(2)``, so dividing
        the Sortino ratio by ``sqrt(2)`` provides a corresponding
        normalization.

        The underlying Sortino ratio uses the arithmetic mean return in
        excess of the minimum acceptable return (MAR) and the second-order
        lower partial moment (LPM2) about MAR.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The Sortino ratio divided by ``sqrt(2)``.
        """
        return self.sortino_ratio / _SQRT2

    @property
    def sortino_satchell_ratio(self) -> float:
        """
        Estimate the Sortino-Satchell ratio.

        The Sortino-Satchell ratio measures the arithmetic mean excess
        return relative to downside deviation. The numerator is the mean
        return in excess of the minimum acceptable return (MAR), while
        the denominator is the square root of the second-order lower
        partial moment (LPM2) about MAR.

        Unlike the standard Sortino ratio, which may use a compounded
        or geometric excess return depending on its implementation, this
        measure uses the arithmetic mean of excess returns.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The arithmetic mean excess return divided by downside
            deviation, or ``math.nan`` if the downside partial moment is
            unavailable or zero.
        """
        lpm2 = self._target_partial_moments.lower_partial_moment_2
        if math.isnan(lpm2) or lpm2 == 0:
            return math.nan
        return self._target_returns_kbn.mean / math.sqrt(lpm2)

    @property
    def omega_ratio(self) -> float:
        """
        Estimate the Omega ratio relative to the minimum acceptable return.

        This implementation uses the simple empirical method, calculating
        Omega directly from the first-order partial moments of the observed
        returns.

        The Omega ratio measures the probability-weighted magnitude of
        returns above the minimum acceptable return (MAR) relative to the
        probability-weighted magnitude of returns below MAR. It can be
        expressed as the ratio of the first-order upper and lower partial
        moments about MAR:

            Omega(MAR) = UPM1(MAR) / LPM1(MAR)

        Equivalently, it can be calculated as one plus the mean excess
        return over MAR divided by the first-order lower partial moment.

        Unlike ratios based on the risk-free rate, this implementation
        uses the MAR (target return) specified when constructing the
        ``Measures`` object.

        Returns:
            The Omega ratio relative to MAR, or ``math.nan`` if the
            lower partial moment is unavailable or zero.
        """
        lpm1 = self._target_partial_moments.lower_partial_moment_1
        if math.isnan(lpm1) or lpm1 == 0:
            return math.nan
        return self._target_returns_kbn.mean / lpm1 + 1

    @property
    def omega_sharpe_ratio(self) -> float:
        """
    Estimate the Omega-Sharpe ratio relative to the minimum acceptable return.

        The Omega-Sharpe ratio is a Sharpe-like transformation of the Omega
        ratio. It is defined as the Omega ratio minus one:

            Omega-Sharpe Ratio = Omega(MAR) - 1

        Equivalently, using first-order upper and lower partial moments about
        MAR:

            Omega-Sharpe Ratio = UPM1(MAR) / LPM1(MAR) - 1

        This measure compares upside potential with downside potential,
        without using second-order moments or standard deviation.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The Omega-Sharpe ratio, or ``math.nan`` if the Omega ratio is
            unavailable.
        """
        return self.omega_ratio - 1

    @property
    def omega_excess_return(self) -> float:
        """
        Estimate the Omega excess return relative to the minimum acceptable return.

        Omega excess return is an annualized downside-risk-adjusted return.
        It is calculated as the annualized portfolio return less three times
        the product of the portfolio and benchmark downside deviations:

            Omega Excess Return = Rp - 3 * sigma_Dp * sigma_Db

        where ``Rp`` is the annualized portfolio return, ``sigma_Dp`` is the
        annualized portfolio downside deviation relative to MAR, and
        ``sigma_Db`` is the annualized benchmark downside deviation relative to MAR.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        The calculation depends on periods per annum.

        Returns:
            The annualized downside-risk-adjusted return, or ``math.nan`` if
            the required return or downside-risk measures are unavailable.
        """
        def benchmark_downside_deviation() ->float:
            if self._returns_benchmark is None or len(self._returns_benchmark) == 0:
                return math.nan
            lower_excess_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)
            for r in self._returns_benchmark:
                excess = r - self.target_return
                if excess < 0:
                    lower_excess_kbn.update(-excess)
            return math.sqrt(lower_excess_kbn.x2_sum / len(self._returns_benchmark))
        
        period = self.periods_per_annum

        # Annualized portfolio return
        rp = self._cumulative_return.geometric_mean_return
        if math.isnan(rp):
            return math.nan
        rp_annual = (1 + rp) ** period - 1

        sqrt_period = math.sqrt(period)

        # Annualized portfolio downside (full) deviation
        sigma_d_ann = self.downside_deviation * sqrt_period
        
        # Annualized benchmark downside (full) deviation
        sigma_d_ann_bench = benchmark_downside_deviation() * sqrt_period
        
        return rp_annual - 3 * sigma_d_ann * sigma_d_ann_bench

    @property
    def kappa_1_ratio(self):
        """
        Estimate the Kappa ratio of order 1.

        The Kappa-1 ratio measures the mean excess return over the
        minimum acceptable return (MAR) relative to the first-order
        lower partial moment (LPM1) about MAR:

            Kappa_1 = E[R - MAR] / LPM1(MAR)

        Because LPM1 measures the average downside shortfall in return
        units, Kappa-1 expresses excess return relative to average
        downside potential.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The Kappa-1 ratio, or ``math.nan`` if the lower partial
            moment is unavailable or zero.
        """
        lpm = self._target_partial_moments.lower_partial_moment_1
        if math.isnan(lpm) or lpm == 0:
            return math.nan
        return self._target_returns_kbn.mean / lpm

    @property
    def kappa_2_ratio(self):
        """
        Estimate the Kappa ratio of order 2.

        The Kappa-2 ratio measures the mean excess return over the
        minimum acceptable return (MAR) relative to the square root
        of the second-order lower partial moment (LPM2) about MAR:

            Kappa_2 = E[R - MAR] / sqrt(LPM2(MAR))

        Kappa-2 is equivalent to the standard Sortino ratio when both
        measures use the same MAR and the same definition of LPM2.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The Kappa-2 ratio, or ``math.nan`` if the lower partial
            moment is unavailable or zero.
        """
        lpm = self._target_partial_moments.lower_partial_moment_2
        if math.isnan(lpm) or lpm == 0:
            return math.nan
        return self._target_returns_kbn.mean / math.sqrt(lpm)

    @property
    def kappa_3_ratio(self):
        """
        Estimate the Kappa ratio of order 3.

        The Kappa-3 ratio measures the mean excess return over the
        minimum acceptable return (MAR) relative to the cube root of
        the third-order lower partial moment (LPM3) about MAR:

            Kappa_3 = E[R - MAR] / LPM3(MAR)^(1/3)

        Compared with Kappa-1 and Kappa-2, Kappa-3 places greater
        emphasis on the magnitude of large downside deviations through
        the higher-order lower partial moment.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The Kappa-3 ratio, or ``math.nan`` if the lower partial
            moment is unavailable or zero.
        """
        lpm = self._target_partial_moments.lower_partial_moment_3
        if math.isnan(lpm) or lpm == 0:
            return math.nan
        return self._target_returns_kbn.mean / (lpm ** (1/3))

    @property
    def kappa_4_ratio(self):
        """
        Estimate the Kappa ratio of order 4.

        The Kappa-4 ratio measures the mean excess return over the
        minimum acceptable return (MAR) relative to the fourth root of
        the fourth-order lower partial moment (LPM4) about MAR:

            Kappa_4 = E[R - MAR] / LPM4(MAR)^(1/4)

        The fourth-order lower partial moment places substantially
        greater emphasis on large downside deviations, making Kappa-4
        more sensitive to severe downside outcomes than lower-order
        Kappa ratios.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The Kappa-4 ratio, or ``math.nan`` if the lower partial
            moment is unavailable or zero.
        """
        lpm = self._target_partial_moments.lower_partial_moment_4
        if math.isnan(lpm) or lpm == 0:
            return math.nan
        return self._target_returns_kbn.mean / (lpm ** (1/4))

    def prospect_ratio(self, lambda_loss: float = 2.25) -> float:
        """
        Prospect Ratio based on Watanabe's formulation described by Bacon.

        The Prospect Ratio is a Sharpe/Sortino-type performance measure that
        incorporates investor preferences toward gains and losses in the
        numerator. Positive returns contribute normally, while negative
        returns are weighted by the loss-aversion coefficient ``lambda_loss``.

        Formula::

            prospect_ratio =
                ((sum_positive + lambda_loss * sum_negative) / n - target_return) / downside_deviation

        where ``sum_positive`` is the sum of positive returns,
        ``sum_negative`` is the signed sum of negative returns, ``n`` is the
        number of observations, and ``target_return`` is the configured
        minimum acceptable return (MAR).

        ``lambda_loss`` controls the investor's assumed sensitivity to losses
        relative to gains:

            * ``lambda_loss = 0``:
              no loss aversion; negative returns receive no additional
              penalty in the numerator.

            * ``lambda_loss > 0``:
              loss-averse behavior; because ``sum_negative`` is negative,
              increasing ``lambda_loss`` makes the Prospect Ratio smaller
              when losses are present.

            * ``lambda_loss = 1``:
              losses and gains receive equal linear weighting.

            * ``lambda_loss = 2.25``:
              the default value proposed by Watanabe, based on empirical
              research suggesting that investors dislike losses approximately
              two and a quarter times as much as they enjoy equivalent gains.

            * ``lambda_loss < 0``:
              gain-seeking behavior in the terminology used by Bacon; the
              negative-return component increases rather than decreases the
              numerator.

        The value of ``lambda_loss`` is an investor-preference parameter,
        not a property estimated from the return series. The default value
        of 2.25 is appropriate when following Watanabe's empirical
        parameterization. A value of 1 can be useful as a neutral reference
        point, while 0 removes the contribution of losses from the
        preference-adjusted return. Investors or strategy researchers may
        choose a different positive value to reflect their own degree of
        loss aversion or to perform sensitivity analysis.

        The result is not annualized.

        Args:
            lambda_loss: Loss-aversion coefficient applied to negative
                returns. Defaults to 2.25.

        Returns:
            float: Prospect Ratio, or ``math.nan`` when downside deviation
            is unavailable or zero, or when there are no observations.
        """
        ddev = self.downside_deviation
        if math.isnan(ddev) or ddev == 0:
            return math.nan
    
        pm = self._raw_partial_moments
        n = pm.count
        if n == 0:
            return math.nan
    
        prospect_return = (pm.sum_positive + lambda_loss * pm.sum_negative) / n
    
        return (prospect_return - self.target_return) / ddev

    @property
    def prospect_ratio_performance_analytics(self) -> float:
        """
        PerformanceAnalytics Prospect Ratio.
    
        Calculates the Prospect Ratio using the implementation of
        ``ProspectRatio`` in R's ``PerformanceAnalytics`` package.
    
        The implementation applies the loss-aversion coefficient of 2.25 to
        negative returns and subtracts the target return from the resulting
        sum before dividing by the number of observations and downside
        deviation.
    
        Formula::
    
            prospect_ratio_perfan =
                (sum_positive + 2.25 * sum_negative - target_return) / (n * downside_deviation)
    
        where ``sum_positive`` is the sum of positive returns,
        ``sum_negative`` is the signed sum of negative returns, and ``n`` is
        the number of observations.
    
        This follows the ``PerformanceAnalytics`` source implementation,
        which differs from the prospect-ratio formula documented by Bacon
        when ``target_return`` is non-zero.
    
        The result is not annualized.
    
        Returns:
            float: PerformanceAnalytics Prospect Ratio, or ``math.nan``
            when downside deviation is unavailable or zero.
        """
        lambda_loss = 2.25
        ddev = self.downside_deviation
        if math.isnan(ddev) or ddev == 0:
            return math.nan
    
        pm = self._raw_partial_moments
        n = pm.count
        if n == 0:
            return math.nan
    
        return (pm.sum_positive + lambda_loss * pm.sum_negative - self.target_return) / (ddev * n)

    @property
    def bernardo_ledoit_ratio(self):
        """
        Estimate the Bernardo-Ledoit ratio.

        The Bernardo-Ledoit ratio compares the first-order upper and lower
        partial moments about zero:

            Bernardo-Ledoit Ratio = HPM1(0) / LPM1(0)

        Equivalently, because the same normalization factor cancels from
        numerator and denominator, it is the ratio of the total positive
        returns to the absolute total negative returns:

            Bernardo-Ledoit Ratio = sum(R > 0) / abs(sum(R < 0))

        Thus, with partial moments calculated relative to zero, this measure
        is mathematically equivalent to ``gain_loss_ratio``.

        Unlike the Kappa ratios and the MAR-based Omega ratio, the
        Bernardo-Ledoit ratio does not use the ``Measures`` object's MAR.
        Its reference level is always zero.

        A value greater than 1 indicates that the total magnitude of positive
        returns exceeds the total magnitude of negative returns. A value below
        1 indicates the opposite, while 1 indicates equal total gains and
        losses.

        Returns:
            The Bernardo-Ledoit ratio, or ``math.nan`` if the first-order
            lower partial moment is zero.
        """
        lpm_1 = self._raw_partial_moments.lower_partial_moment_1
        hpm_1 = self._raw_partial_moments.higher_partial_moment_1
        return hpm_1 / lpm_1 if lpm_1 != 0 else math.nan

    @property
    def d_ratio(self) -> float:
        """
        D-Ratio.
    
        Measures downside loss relative to upside gain while accounting for
        the frequency of negative and positive returns:
    
            D-Ratio = (n_d * sum(max(-R_t, 0))) / (n_u * sum(max(R_t, 0)))
    
        where:
    
            n_d = number of negative returns
            n_u = number of positive returns
    
        D-Ratio is related to the Bernardo-Ledoit ratio. Specifically:
    
            D-Ratio = (n_d / n_u) / Bernardo-Ledoit
    
        Unlike the Bernardo-Ledoit ratio, which compares the total upside
        and downside magnitudes, D-Ratio also accounts for how frequently
        positive and negative returns occur.
    
        Lower values indicate better performance. A value of zero indicates
        that there are no negative returns, while infinity indicates that
        there are no positive returns.
    
        Zero returns are excluded from both the positive and negative return
        counts and sums.
    
        Returns:
            D-Ratio, or ``math.inf`` when there are no positive returns.
        """
        n_up = self._win_loss.winning_returns_count
        if n_up == 0:
            return math.inf  # No positive returns
        n_down = self._win_loss.losing_returns_count
        if n_down == 0:
            return 0.0   # No negative returns

        sum_up = self._win_loss.winning_returns_sum
        sum_down = self._win_loss.losing_returns_sum                
        return (-n_down * sum_down) / (n_up * sum_up)

    @property
    def gain_loss_ratio(self) -> float:
        """
        Estimate the gain-loss ratio.

        The gain-loss ratio compares the total magnitude of positive returns
        with the total magnitude of negative returns:

            Gain-Loss Ratio = sum(R > 0) / abs(sum(R < 0))

        This is mathematically equivalent to the Bernardo-Ledoit ratio when
        the latter is calculated using first-order partial moments about zero.
        The two measures differ only in their formulation: the
        Bernardo-Ledoit ratio is expressed using partial moments, while this
        measure is expressed directly as the sums of gains and losses.

        The reference level is zero; the ``Measures`` object's MAR does not
        affect this measure.

        A value greater than 1 indicates that the total magnitude of positive
        returns exceeds the total magnitude of negative returns. A value below
        1 indicates the opposite, while 1 indicates equal total gains and
        losses.

        Returns:
            The gain-loss ratio, or ``math.nan`` if the total magnitude of
            negative returns is zero.
        """
        sum_losses = abs(self._win_loss.losing_returns_sum)        
        return self._win_loss.winning_returns_sum / sum_losses if sum_losses != 0 else math.nan

    @property
    def mean_non_zero_return(self) -> float:
        """
        Arithmetic mean of non-zero returns.
    
        Computes the mean return after excluding observations equal to zero:
    
            mean_non_zero_return = sum(R_t) / N_nonzero
    
        Zero returns are not included in either the numerator or denominator.
    
        Returns:
            The arithmetic mean of non-zero returns, or ``math.nan`` if there
            are no non-zero returns.
        """
        return self._win_loss.non_zero_returns_mean

    @property
    def mean_win_return(self) -> float:
        """
        Arithmetic mean of winning returns.
    
        Computes the average return conditional on a positive return:
    
            mean_win_return = sum(R_t | R_t > 0) / N_win
    
        Only strictly positive returns are included.
    
        Returns:
            The arithmetic mean of winning returns, or ``math.nan`` if there
            are no positive returns.
        """
        return self._win_loss.winning_returns_mean

    @property
    def mean_loss_return(self) -> float:
        """
        Arithmetic mean of losing returns.
    
        Computes the average return conditional on a negative return:
    
            mean_loss_return = sum(R_t | R_t < 0) / N_loss
    
        Only strictly negative returns are included. The result is therefore
        negative.
    
        Returns:
            The arithmetic mean of losing returns, or ``math.nan`` if there
            are no negative returns.
        """
        return self._win_loss.losing_returns_mean

    @property
    def win_rate(self) -> float:
        """
        Win rate among non-zero returns.
    
        Computes the proportion of non-zero returns that are positive:
    
            win_rate = N_win / N_nonzero
    
        Zero returns are excluded from the denominator. Thus, the measure
        represents the probability of a winning return conditional on the
        return being non-zero.
    
        Returns:
            The proportion of non-zero returns that are positive, in the
            range [0, 1], or ``math.nan`` if there are no non-zero returns.
        """
        non_zero_count = self._win_loss.non_zero_returns_count
        if non_zero_count <= 0:
            return math.nan
        return self._win_loss.winning_returns_count / non_zero_count

    @property
    def loss_rate(self) -> float:
        """
        Loss rate among non-zero returns.
    
        Computes the proportion of non-zero returns that are negative:
    
            loss_rate = N_loss / N_nonzero
    
        Zero returns are excluded from the denominator. Thus, the measure
        represents the probability of a losing return conditional on the
        return being non-zero.
    
        Returns:
            The proportion of non-zero returns that are negative, in the
            range [0, 1], or ``math.nan`` if there are no non-zero returns.
        """
        non_zero_count = self._win_loss.non_zero_returns_count
        if non_zero_count <= 0:
            return math.nan
        return self._win_loss.losing_returns_count / non_zero_count

    @property
    def variability_skewness(self) -> float:
        """
        Estimate the upside-to-downside variability skewness relative to MAR.

        Variability skewness compares the second-order upper and lower
        partial moments about the minimum acceptable return (MAR):

            Variability Skewness = UPM2(MAR) / LPM2(MAR)

        The upper partial moment (UPM2) measures the average squared
        deviations of returns above MAR, while the lower partial moment
        (LPM2) measures the average squared deviations of returns below
        MAR. Both partial moments use the full normalization by the total
        number of observations.

        A value greater than 1 indicates that upside variability exceeds
        downside variability. A value of 1 indicates equal upside and
        downside variability, while a value below 1 indicates greater
        downside variability.

        This measure is analogous to the Omega ratio, but uses second-order
        partial moments rather than first-order partial moments.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The ratio of upside to downside second-order partial moments,
            or ``math.nan`` if either partial moment is unavailable or the
            downside partial moment is zero.
        """
        up_moment = self._target_partial_moments.higher_partial_moment_2
        down_moment = self._target_partial_moments.lower_partial_moment_2

        if math.isnan(up_moment) or math.isnan(down_moment) or down_moment == 0:
            return math.nan

        return up_moment / down_moment

    @property
    def volatility_skewness(self) -> float:
        """
        Estimate the upside-to-downside volatility skewness relative to MAR.

        Volatility skewness compares the square roots of the second-order
        upper and lower partial moments about the minimum acceptable return
        (MAR):

            Volatility Skewness = sqrt(UPM2(MAR) / LPM2(MAR))

        It is the square root of ``variability_skewness`` and compares
        upside and downside volatility in return units rather than squared
        return units.

        A value greater than 1 indicates greater upside volatility than
        downside volatility. A value of 1 indicates equal upside and
        downside volatility, while a value below 1 indicates greater
        downside volatility.

        This measure is analogous to the Omega ratio, but uses second-order
        partial moments rather than first-order partial moments.

        The MAR (target return) is specified when constructing the
        ``Measures`` object.

        Returns:
            The ratio of upside to downside volatility, or ``math.nan`` if
            the underlying variability skewness is unavailable.
        """
        var_skew = self.variability_skewness
        return math.sqrt(var_skew) if not math.isnan(var_skew) else math.nan

    def farinelli_tibiletti_ratio(self, upper_order: int = 2, lower_order: int = 2) -> float:
        """
        Estimate the Farinelli-Tibiletti ratio.
    
        The Farinelli-Tibiletti ratio is a generalized upside-to-downside
        performance measure based on upper and lower partial moments:
    
            FT(u, l) = UPM_u(MAR)^(1/u) / LPM_l(MAR)^(1/l)
    
        where ``UPM_u`` is the upper partial moment of order ``u`` and
        ``LPM_l`` is the lower partial moment of order ``l`, both calculated
        relative to the minimum acceptable return (MAR).
    
        The partial moments use the full-observation normalization, dividing
        by the total number of observations rather than only by the number
        of observations on the corresponding side of MAR.
    
        Important special cases include:
    
            FT(1, 1) = Omega ratio
            FT(1, 2) = Upside Potential Ratio
            FT(2, 2) = Volatility Skewness
    
        Consequently, the Farinelli-Tibiletti ratio provides a generalized
        framework encompassing several upside- and downside-based performance
        measures.
    
        The MAR (target return) is specified when constructing the
        ``Measures`` object.
    
        Args:
            upper_order:
                Order of the upper partial moment. Must be 1, 2, 3, or 4.
            lower_order:
                Order of the lower partial moment. Must be 1, 2, 3, or 4.
    
        Returns:
            The Farinelli-Tibiletti ratio.
    
        Raises:
            ValueError:
                If either partial-moment order is not 1, 2, 3, or 4.
        """
        if upper_order not in (1, 2, 3, 4):
            raise ValueError("upper_order must be 1, 2, 3, or 4")
        if lower_order not in (1, 2, 3, 4):
            raise ValueError("lower_order must be 1, 2, 3, or 4")

        if lower_order == 1:
            denom = self._target_partial_moments.lower_partial_moment_1
        elif lower_order == 2:
            denom = self._target_partial_moments.lower_partial_moment_2
            denom = math.sqrt(denom)
        elif lower_order == 3:
            denom = self._target_partial_moments.lower_partial_moment_3
            denom = denom ** (1.0/3)
        else:
            denom = self._target_partial_moments.lower_partial_moment_4
            denom = denom ** (1.0/4)

        if upper_order == 1:
            num = self._target_partial_moments.higher_partial_moment_1
        elif upper_order == 2:
            num = self._target_partial_moments.higher_partial_moment_2
            num = math.sqrt(num)
        elif upper_order == 3:
            num = self._target_partial_moments.higher_partial_moment_3
            num = num ** (1.0/3)
        else:
            num = self._target_partial_moments.higher_partial_moment_4
            num = num ** (1.0/4)

        return num / denom

    def rachev_ratio(self, alpha: float = 0.1, beta: float = 0.1) -> float:
        """
        Estimate the Rachev ratio.
    
        The Rachev ratio compares the expected return in the upper tail of
        the return distribution with the expected loss in the lower tail:
    
            Rachev Ratio = ES_upper / ES_lower
    
        where ``ES_lower`` is the magnitude of the average return in the
        lower ``alpha`` tail and ``ES_upper`` is the average return in the
        upper ``beta`` tail.
    
        The implementation follows the non-parametric definition used by
        PerformanceAnalytics. The lower-tail threshold is determined by the
        ``alpha`` quantile, while the upper-tail threshold is determined by
        the observation at position ``floor((1 - beta) * n)`` in the sorted
        sample.
    
        Note that, despite the ``rf`` parameter exposed by the
        PerformanceAnalytics implementation, the reference calculation does
        not adjust returns by the risk-free rate. This implementation
        therefore operates directly on the observed returns.
    
        Args:
            alpha:
                Lower-tail probability. For example, ``0.10`` uses the
                10% lower tail.
            beta:
                Upper-tail probability. For example, ``0.10`` uses the
                10% upper tail.
    
        Returns:
            The Rachev ratio, or ``math.nan`` if the required tails cannot
            be calculated or the lower-tail expected loss is zero.
        """
        n = self._returns_kbn.n
        returns = self._returns
    
        if returns is None or n < 2:
            return math.nan
    
        if not 0 < alpha < 1:
            raise ValueError("alpha must be between 0 and 1")
    
        if not 0 < beta < 1:
            raise ValueError("beta must be between 0 and 1")
    
        # Lower-tail VaR and Expected Shortfall.
        lower_var = core.percentile(returns, alpha)
        lower_tail = [r for r in returns if r <= lower_var]
    
        if not lower_tail:
            return math.nan
    
        es_lower = -sum(lower_tail) / len(lower_tail)
    
        # Upper-tail VaR and Expected Shortfall.
        sorted_returns = sorted(returns)
    
        # Performance Analytics uses:
        #
        #   n.upper <- floor((1-beta) * n)
        #   VaR.hat.upper <- sorted.returns[n.upper]
        #
        # R is 1-based, so convert the position to a Python index.
        upper_position = math.floor((1.0 - beta) * n)
    
        if upper_position < 1:
            upper_position = 1
        elif upper_position > n:
            upper_position = n
    
        upper_var = sorted_returns[upper_position - 1]
        upper_tail = [r for r in returns if r >= upper_var]
    
        if not upper_tail or es_lower == 0:
            return math.nan
    
        es_upper = sum(upper_tail) / len(upper_tail)
    
        return es_upper / es_lower

    @property
    def drawdowns_cumulative(self) -> List[float]:
        """
        Drawdown series of cumulative geometric returns.
    
        Each value measures the percentage decline of cumulative portfolio
        wealth from the highest cumulative wealth previously reached:
    
            D_t = W_t / max(W_1, ..., W_t) - 1
    
        where ``W_t`` is cumulative wealth at observation ``t``.
    
        Drawdowns are expressed as decimal returns and are non-positive.
        A value of ``0`` indicates that a new cumulative high has been
        reached. Negative values represent declines from the previous
        cumulative high.
    
        In a rolling-window calculation, the drawdown series corresponds
        to the observations currently contained in the rolling window.
    
        This drawdown representation is used for peak-to-valley measures,
        including maximum drawdown and ratios based on maximum drawdown
        such as the Calmar and Sterling ratios.
    
        Returns:
            List[float]:
                Cumulative geometric-return drawdowns for the current
                observation history or rolling window.
        """
        return list(self._drawdowns_cumulative)

    @property
    def min_drawdowns_cumulative(self) -> float:
        """
        Minimum cumulative geometric-return peak-to-valley  drawdown.
    
        This is the most negative value in the cumulative drawdown series:
    
            min(D_t)
    
        where each ``D_t`` measures the decline of cumulative wealth from
        its previous cumulative high.
    
        Because drawdowns are non-positive, this value represents the
        worst peak-to-valley decline in the observation history or current
        rolling window.
    
        Returns:
            float:
                The minimum cumulative drawdown. Returns ``nan`` when no
                drawdown observations are available.
        """
        return self._drawdowns_cumulative_minmax.min

    @property
    def worst_drawdowns_cumulative(self) -> float:
        """
        Magnitude of the worst cumulative geometric-return peak-to-valley drawdown.
    
        This is the absolute value of the minimum cumulative drawdown:
    
            |min(D_t)|
    
        where ``D_t`` is the drawdown of cumulative geometric wealth from
        its previous high.
    
        Unlike ``min_drawdowns_cumulative``, which is non-positive, this
        property returns the drawdown magnitude as a non-negative value.
        It is therefore convenient as the denominator of ratios such as
        the Calmar and Sterling ratios.
    
        Returns:
            float:
                Absolute magnitude of the worst cumulative drawdown.
                Returns ``nan`` when no drawdown observations are available.
        """
        return abs(self._drawdowns_cumulative_minmax.min)

    @property
    def drawdowns_high_watermark(self) -> List[float]:
        """
        Drawdown series measured from the high-water mark.
    
        For each observation, the drawdown is the percentage distance of
        cumulative portfolio wealth below the highest wealth level reached
        within the applicable observation history or rolling window:
    
            D_t = W_t / H_t - 1
    
        where
    
            H_t = max(W_1, ..., W_t)
    
        is the high-water mark.
    
        Drawdowns are non-positive. A value of ``0`` indicates that the
        portfolio is at a high-water mark, while a negative value indicates
        that it is below its high-water mark.
    
        The complete drawdown series is retained because several
        downside-risk measures depend on the depth of drawdowns over time,
        rather than only on the single worst drawdown.
    
        This representation is used by the Pain Index, Pain Ratio,
        Ulcer Index, and Martin Ratio.
    
        Returns:
            List[float]:
                High-water-mark drawdowns for the current observation
                history or rolling window.
        """
        return list(self._drawdown_high_watermark.drawdowns)

    def drawdowns_continuous_runs(self, max_runs: int = None) -> List[float]:
        """
        Return drawdowns for continuous losing-return runs.
    
        A continuous drawdown is the compounded loss over a maximal
        uninterrupted sequence of negative returns.
    
        For a losing run consisting of returns ``r_a, ..., r_b``:
    
            DD_j = prod(1 + r_i) - 1
    
        for ``i = a, ..., b``.
    
        Each continuous losing run produces exactly one drawdown value.
        Non-negative returns separate consecutive losing runs.
    
        For example, the returns::
    
            [-0.01, -0.02, +0.01, -0.03, -0.04]
    
        contain two continuous losing runs:
    
            (0.99 * 0.98) - 1 = -0.0298
            (0.97 * 0.96) - 1 = -0.0688
    
        The resulting drawdowns are therefore approximately::
    
            [-0.0298, -0.0688]
    
        Drawdowns are non-positive and are expressed using the same
        normalized return convention as the input returns.
    
        This representation is used by the Burke ratio, whose denominator
        is based on the square root of the sum of squared continuous
        drawdowns.
    
        Args:
            max_runs:
                Maximum number of losing runs to return. Runs are sorted
                from most negative to least negative before truncation.
                ``None`` or a non-positive value returns all unsorted runs.
    
        Returns:
            List[float]:
                Continuous compounded drawdowns, one value per losing run,
                optionally restricted to the worst ``max_runs`` episodes.
        """
        # One (negative) value per losing run
        drawdowns = self._drawdown_continuous_runs.drawdowns
        if len(drawdowns) < 1:
            return []
        if max_runs is not None and max_runs > 0:
            # Ascending: most negative (worst) first
            drawdowns = sorted(drawdowns)
            # Slice to keep only the worst 'max_peaks'
            drawdowns = drawdowns[:max_runs]
        return drawdowns

    @property
    def calmar_ratio(self) -> float:
        """
        Calmar ratio based on annualized return and maximum drawdown.
    
        The Calmar ratio is calculated as:
    
            Calmar = CAGR / |MDD|
    
        where ``CAGR`` is the geometric mean return of the portfolio and
        ``MDD`` is the worst cumulative peak-to-valley drawdown.
    
        The drawdown is calculated from cumulative geometric returns, not
        from individual-period returns.
    
        Because drawdowns are stored as non-positive values, the magnitude
        of the worst drawdown is obtained as:
    
            |min(D_t)|
    
        A zero maximum drawdown produces ``nan`` because the ratio would
        otherwise require division by zero.
    
        Returns:
            float:
                Calmar ratio.
        """
        wdd = self.worst_drawdowns_cumulative
        if wdd == 0:
            return math.nan
        cagr = self._cumulative_return.geometric_mean_return
        if math.isnan(cagr):
            return math.nan
        return cagr / wdd

    def sterling_ratio(self, excess: float = 0.1) -> float:
        """
        Sterling ratio based on annualized return and maximum drawdown.
    
        The Sterling ratio is calculated as:
    
            Sterling = CAGR / (|MDD| + R_e)
    
        where ``CAGR`` is the geometric mean return, ``MDD`` is the worst
        cumulative peak-to-valley drawdown, and ``R_e`` is the specified
        excess adjustment.
    
        ``excess`` is an additional drawdown allowance used in the Sterling
        denominator. Both ``excess`` and worst cumulative peak-to-valley
        drawdown are in percentage points [0,1].
        The traditional Sterling ratio uses an ``excess`` of 0.1 (10%).
    
        Args:
            excess:
                Excess added to the magnitude of maximum drawdown, [0, 1].
                A value of ``0`` applies no adjustment.
                Default: 0.1 (10%)
    
        Returns:
            float:
                Sterling ratio.
        """
        wdd = self.worst_drawdowns_cumulative + excess
        if wdd == 0:
            return math.nan
        cagr = self._cumulative_return.geometric_mean_return
        if math.isnan(cagr):
            return math.nan
        return cagr / wdd

    @property
    def burke_ratio(self) -> float:
        """
        Burke ratio based on continuous losing-return drawdowns.
    
        The Burke ratio measures return relative to the severity of
        continuous losing episodes.
    
        It is calculated as:
    
            Burke = (R_p - R_f) / sqrt(sum_j DD_j^2)
    
        where ``R_p`` is the geometric mean return of the portfolio,
        ``R_f`` is the configured risk-free rate, and ``DD_j`` is the
        compounded drawdown of the ``j``-th continuous losing run.
    
        A continuous losing run is a maximal sequence of consecutive
        negative returns. Its drawdown is the compounded return over the
        entire run:
    
            DD_j = prod_{i in run_j}(1 + r_i) - 1
    
        The denominator therefore penalizes the severity of individual
        losing episodes and, unlike maximum drawdown, incorporates all
        continuous losing runs.
    
        The risk-free rate is the rate configured when the ``Measures``
        instance is constructed.
    
        Returns:
            float:
                Burke ratio. Returns ``nan`` when the return is undefined
                or when there are no non-zero continuous drawdowns.
        """
        rate = self._cumulative_return.geometric_mean_return - self.risk_free_rate
        if math.isnan(rate):
            return math.nan
        sqrt_sum_drawdowns_squared = self._drawdown_continuous_runs.sqrt_sum_drawdowns_squared
        if sqrt_sum_drawdowns_squared == 0:
            return math.nan
        return rate / sqrt_sum_drawdowns_squared

    @property
    def burke_ratio_modified(self) -> float:
        """
        Modified Burke ratio.
    
        The modified Burke ratio scales the Burke ratio by the square root
        of the number of return observations:
    
            Modified Burke =
                Burke * sqrt(n)
    
        where ``n`` is the number of observations in the current return
        sample.
    
        The underlying Burke ratio uses continuous losing-return
        drawdowns, with each maximal sequence of consecutive negative
        returns treated as a single compounded losing episode.
    
        The risk-free rate is the rate configured when the ``Measures``
        instance is constructed.
    
        The modified form therefore preserves the Burke ratio's
        episode-based downside measure while applying a sample-size
        adjustment.
    
        Returns:
            float:
                Modified Burke ratio. Returns ``nan`` when the underlying
                Burke ratio is undefined.
        """
        burke = self.burke_ratio
        if math.isnan(burke):
            return math.nan
        return burke * math.sqrt(self._returns_kbn.n)

    @property
    def pain_index(self) -> float:
        """
        Average depth below the high-water mark.
    
        The Pain Index is the arithmetic mean of the magnitudes of
        high-water-mark drawdowns:
    
            Pain Index = - (1/n) * sum_t D_t
    
        where ``D_t <= 0`` is the drawdown from the high-water mark at
        observation ``t``.
    
        Because drawdowns are non-positive, the implementation negates
        their mean so that the Pain Index is expressed as a non-negative
        measure of average drawdown depth.
    
        Unlike maximum drawdown, which considers only the single worst
        observation, the Pain Index incorporates the entire drawdown
        history and therefore reflects both the depth and persistence of
        being below the high-water mark.
    
        Returns:
            float:
                Mean high-water-mark drawdown magnitude.
        """
        # By calculation, all values are <= 0, so we don't need abs()
        return -self._drawdown_high_watermark.drawdowns_mean

    @property
    def pain_ratio(self) -> float:
        """
        Pain ratio based on return relative to average drawdown.

        The Pain Ratio is calculated as:

            Pain Ratio = (R_p - R_f) / Pain Index

        where ``R_p`` is the geometric mean portfolio return, ``R_f`` is
        the configured risk-free rate, and the Pain Index is the average
        magnitude of the high-water-mark drawdowns.
    
        Unlike the Calmar ratio, which uses only the worst drawdown, the
        Pain Ratio uses the average depth of drawdowns over time.

        The risk-free rate is the rate configured when the ``Measures``
        instance is constructed.

        Returns:
            float:
                Pain Ratio. Returns ``nan`` when the return is undefined
                or when the Pain Index is zero.
        """
        rate = self._cumulative_return.geometric_mean_return - self.risk_free_rate
        if math.isnan(rate):
            return math.nan
        pain_index = self.pain_index
        return rate / pain_index if pain_index != 0 else math.nan

    @property
    def ulcer_index(self) -> float:
        """
        Root-mean-square high-water-mark drawdown.

        The Ulcer Index measures the depth and persistence of drawdowns
        by taking the root mean square of the high-water-mark drawdown
        series:

            UI = sqrt((1/n) * sum_t D_t^2)

        where ``D_t <= 0`` is the high-water-mark drawdown at observation
        ``t``.

        Squaring the drawdowns gives greater weight to deeper drawdowns.
        Consequently, the Ulcer Index reflects both how far the portfolio
        falls below its high-water mark and how long it remains there.

        Unlike maximum drawdown, the Ulcer Index does not depend only on
        the single worst drawdown.

        Returns:
            float:
                Ulcer Index, expressed in the same percentage/return units
                as the underlying drawdowns.
        """
        return math.sqrt(self._drawdown_high_watermark.drawdowns_squared_mean)

    @property
    def martin_ratio(self) -> float:
        """
        Martin ratio based on the Ulcer Index.

        The Martin ratio is calculated as:

            Martin = (R_p - R_f) / UI

        where ``R_p`` is the geometric mean portfolio return, ``R_f`` is
        the configured risk-free rate, and ``UI`` is the Ulcer Index.

        The Ulcer Index is calculated from the complete high-water-mark
        drawdown series and therefore penalizes both the depth and
        persistence of drawdowns.

        Unlike the Calmar ratio, which uses only maximum drawdown, the
        Martin ratio incorporates the entire history of underwater
        observations.

        The risk-free rate is the rate configured when the ``Measures``
        instance is constructed.

        Returns:
            float:
                Martin ratio. Returns ``nan`` when the return is undefined
                or when the Ulcer Index is zero.
        """
        rate = self._cumulative_return.geometric_mean_return - self.risk_free_rate
        if math.isnan(rate):
            return math.nan
        ulcer_index = self.ulcer_index
        return rate / ulcer_index if ulcer_index != 0 else math.nan

    @property
    def drawdown_average(self) -> float:
        """
        Average Drawdown (ADD).

        Computes the mean magnitude of the observed discrete drawdown
        episodes.

        Returns:
            Average drawdown as a positive decimal value, or zero
            when no drawdowns are present.
        """
        return self._drawdown_episodes.average_episode_drawdown

    @property
    def drawdown_average_length(self) -> float:
        """
        Average drawdown episode length.

        Computes the average number of observations from the beginning of
        each drawdown episode through its trough to its recovery.

        Returns:
            Average drawdown length in observations, or zero when
            no drawdowns are present.
        """
        return self._drawdown_episodes.average_episode_length

    @property
    def drawdown_average_peak_to_trough(self) -> float:
        """
        Average drawdown peak-to-trough period.

        Computes the average number of observations required for observed
        drawdowns to go from their peak to trough.

        Returns:
            Average recovery period in observations, or zero when
            no drawdowns are present.
        """
        return self._drawdown_episodes.average_episode_peak_to_trough

    @property
    def drawdown_average_recovery(self) -> float:
        """
        Average drawdown recovery period.

        Computes the average number of observations required for observed
        drawdowns to recover from their trough to the preceding high-water
        mark.

        Returns:
            Average recovery period in observations, or zero when
            no drawdowns are present.
        """
        return self._drawdown_episodes.average_episode_recovery

    @property
    def drawdown_deviation(self) -> float:
        """
        Drawdown deviation.

        Computes the root-mean-square magnitude of the observed discrete
        drawdown episodes, normalized by the total number of return
        observations.

        DD = sqrt(sum[j=1,2,...,d](D_j^2/n)) where
            D_j = jth drawdown over the entire period
            d = total number of drawdowns in entire period
            n = number of observations

        Returns:
            Drawdown deviation as a positive decimal value, or zero
            when no drawdowns are present.
        """
        return math.sqrt(self._drawdown_episodes.average_episode_drawdown_squared)
    
    def cdar_average(self, confidence: float) -> float:
        """
        Conditional Drawdown at Risk (CDaR).

        CDaR measures the mean severity of drawdowns in the worst
        ``1 - confidence`` tail of the drawdown distribution.

        Calculates the mean of the continuous drawdown observations in
        the worst tail. This corresponds to the continuous-path method
        in PerformanceAnalytics.

        Args:
            confidence:
                Confidence level of the drawdown tail. For example, ``0.95``
                selects the worst 5% of drawdowns.

        Returns:
            CDaR as a positive drawdown magnitude, or zero when it
            cannot be calculated.
        """
        if not 0.0 < confidence < 1.0:
            raise ValueError("confidence must be between 0 and 1")

        drawdowns = self._drawdown_high_watermark.drawdowns
        if not drawdowns:
            return 0

        q = core.percentile(drawdowns, 1.0 - confidence)
        if q >= 0.0:
            return 0

        tail_sum = KleinKBNAccumulator()
        tail_len = 0
        for dd in drawdowns:
            if dd <= q:
                tail_len += 1
                tail_sum.update(dd)

        return -tail_sum.value / tail_len if tail_len > 0 else 0

    def cdar_discrete(self, confidence: float) -> float:
        """
        Conditional Drawdown at Risk (CDaR).

        CDaR measures the mean severity of drawdowns in the worst
        ``1 - confidence`` tail of the drawdown distribution.

        Calculates the expected shortfall of the discrete high-water-mark
        drawdown episodes. This is the default and corresponds to the
        discrete method in PerformanceAnalytics.

        Args:
            confidence:
                Confidence level of the drawdown tail. For example, ``0.95``
                selects the worst 5% of drawdowns.

        Returns:
            CDaR as a positive drawdown magnitude, or zero when it
            cannot be calculated.
        """
        if not 0.0 < confidence < 1.0:
            raise ValueError("confidence must be between 0 and 1")

        depths = self._drawdown_episodes.depths
        if not depths:
            return 0

        q = core.percentile(depths, 1.0 - confidence)

        tail_sum = KleinKBNAccumulator()
        tail_len = 0
        for depth in depths:
            if depth <= q:
                tail_len += 1
                tail_sum.update(depth)

        return -tail_sum.value / tail_len if tail_len > 0 else 0

    def cdar_beta(self, confidence: float = 0.95) -> float:
        """
        Conditional Drawdown Beta (CDaR Beta).
    
        Measures the sensitivity of portfolio returns to the benchmark's
        worst drawdown episodes.
    
        The implementation follows the legacy PerformanceAnalytics
        ``CDaR.beta`` definition:
    
        1. Identify discrete benchmark drawdown episodes using geometric
           drawdown chaining.
        2. Select the worst ``1 - confidence`` fraction of drawdown episodes.
        3. For each selected episode, calculate the portfolio return from
           the beginning of the drawdown through its trough.
        4. Aggregate those portfolio returns geometrically.
        5. Divide their sum by the number of selected drawdowns times the
           benchmark CDD quantile.
    
        Note
        ----
        This follows the behavior of the supplied PerformanceAnalytics R
        implementation. In particular, the legacy ``CDD()`` implementation returns
        the drawdown quantile rather than the conditional mean despite its
        documentation describing CDD as a conditional measure.
    
        Args:
            confidence:
                Confidence level for the drawdown tail. For example, ``0.95``
                selects drawdowns at or below the 5th percentile.
    
        Returns:
            CDaR Beta, or ``math.nan`` when it cannot be calculated.
        """
        w = self._returns
        if w is None:
            return math.nan

        episodes = self._drawdown_episodes_benchmark.episodes
        depths = self._drawdown_episodes_benchmark.depths
        if not depths:
            return math.nan

        q = core.percentile(depths, 1.0 - confidence)
        if  q == 0.0:
            return math.nan

        sum_ret = KleinKBNAccumulator()
        ret = KleinKBNAccumulator()
        tail_len = 0
        for episode in episodes:
            if episode.depth <= q:
                tail_len += 1
                ret.reset()
                for i in range(episode.from_idx, episode.trough_idx + 1):
                    ret.update(math.log1p(w[i]))
                sum_ret.update(math.expm1(ret.value))

        return sum_ret.value / (tail_len * q) if tail_len != 0 else math.nan

    def cdar_alpha(self, confidence: float = 0.95) -> float:
        """
        Conditional Drawdown Alpha (CDaR Alpha).
    
        Measures the difference between the portfolio's annualized return
        and the annualized return predicted by its Conditional Drawdown Beta
        relative to the benchmark.

        Uses geometric chaining for the annualized return calculation.

        CDaR Alpha is analogous to CAPM alpha, but the beta is estimated
        using the portfolio's performance during the benchmark's worst
        drawdown periods.

        The implementation follows the PerformanceAnalytics ``CDaR.alpha``
        definition.
    
        Args:
            confidence:
                Confidence level for the benchmark drawdown tail. For example,
                ``0.95`` selects the worst 5% of benchmark drawdown episodes.
    
        Returns:
            Annualized CDaR Alpha, or ``math.nan`` when it cannot be calculated.
        """
        beta = self.cdar_beta(confidence=confidence)    
        if math.isnan(beta):
            return math.nan
    
        period = self.periods_per_annum
    
        # PerformanceAnalytics::CDaR.alpha uses:
        #
        #   Rm_expected_annualized = (1 + mean(Rm))^12 - 1
        #   R_expected_annualized  = (1 + mean(R))^12 - 1
        #
        # for geometric = TRUE
        #
        # Note that the R implementation assumes monthly input data and
        # hard-codes 12 rather than using a general annualization scale.
    
        r_mean = self._returns_kbn.mean
        b_mean = self._benchmark_returns_kbn.mean
    
        r_annual = (1.0 + r_mean) ** period - 1.0
        b_annual = (1.0 + b_mean) ** period - 1.0
    
        return r_annual - beta * b_annual

    def reward_to_conditional_drawdown(self, confidence: float = 0.95) -> float:
        """
        Reward to Conditional Drawdown Ratio.
        
        CAGR divided by Conditional Drawdown at Risk (CDaR).
        Uses historical CDaR as the average of worst (1-confidence) drawdowns.
        
        Args:
            confidence: Confidence level for conditional drawdown
            
        Returns:
            Reward-to-CDaR or math.nan
        """
        cagr = self._cumulative_return.geometric_mean_return
        if math.isnan(cagr):
            return math.nan
        
        # Get drawdowns from cumulative returns
        dd = self.drawdowns_cumulative
        if len(dd) < 1:
            return math.nan
        
        # Conditional drawdown: average of worst (1-confidence) drawdowns
        n_tail = max(1, int(len(dd) * (1 - confidence)))
        sorted_dd = sorted(dd)  # Most negative first
        sorted_tail = sorted_dd[:n_tail]
        # Positive number
        cdar = -sum(sorted_tail) / len(sorted_tail)

        return cagr / cdar if cdar != 0 else math.nan

    @property
    def sfm_risk_premium(self) -> float:
        """
        Single-Factor Model (SFM) Risk Premium.

        The arithmetic mean of the strategy's periodic excess returns over
        the risk-free rate.

        The risk premium is calculated as::

            sfm_risk_premium = mean(Rp - Rf)

        where ``Rp`` is the strategy return and ``Rf`` is the risk-free
        return for each observation.

        The result is expressed in the same periodicity as the input
        returns and is not annualized.

        The risk-free rate is configured when the ``Measures``
        instance is constructed.

        Returns:
            float: Arithmetic mean excess return, or ``math.nan`` when
            no valid excess-return observations are available.
        """
        return self._excess_returns_kbn.mean

    @property
    def sfm_alpha(self) -> float:
        """
        Single-Factor Model (SFM) Alpha coefficient.
        
        Alpha is the intercept of the least-squares regression of portfolio
        excess returns on benchmark excess returns:
        
            R_p - R_f = alpha + beta * (R_b - R_f) + epsilon
        
        Equivalently:
        
            alpha = mean(R_p - R_f) - beta * mean(R_b - R_f)
        
        The benchmark may be any return series; it is not required to be a
        market portfolio. Consequently, this property is described as
        Single-Factor Model alpha rather than assuming a CAPM interpretation.
        
        The risk-free rate is configured when the ``Measures`` instance is
        constructed and is used to calculate both portfolio and benchmark
        excess returns.
        
        Returns:
            The estimated SFM intercept. Returns ``math.nan`` if the benchmark
            beta cannot be estimated.
        """
        #return self._excess_covariance.alpha
        return self._sfm_regression.alpha

    @property
    def sfm_beta(self) -> float:
        """
        Single-Factor Model (SFM) Beta coefficient.
    
        Beta is the least-squares slope from the regression of portfolio
        excess returns on benchmark excess returns:
    
            R_p - R_f = alpha + beta * (R_b - R_f) + epsilon
    
        It is calculated as:
    
            beta = Cov(R_p - R_f, R_b - R_f)
                   -------------------------
                   Var(R_b - R_f)
    
        The benchmark may be any return series. When the benchmark represents
        the market portfolio, this coefficient is commonly interpreted as
        CAPM beta.
    
        The risk-free rate is configured when the ``Measures`` instance is
        constructed and is used to calculate both portfolio and benchmark
        excess returns.
    
        Returns:
            The estimated SFM beta, or ``math.nan`` if the benchmark excess
            returns have zero variance.
        """
        #return self._excess_covariance.beta
        return self._sfm_regression.beta

    @property
    def sfm_beta_bull(self) -> float:
        """
        Single-Factor Model (SFM) Bull Beta coefficient.
    
        Bull beta is the least-squares slope from the regression of portfolio
        excess returns on benchmark excess returns, restricted to observations
        for which the benchmark excess return is positive:
    
            R_b - R_f > 0
    
        The conditional regression is:
    
            R_p - R_f = alpha_bull + beta_bull * (R_b - R_f) + epsilon
    
        Bull beta measures the portfolio's sensitivity to the benchmark during
        periods in which the benchmark return exceeds the risk-free rate.
    
        The risk-free rate is configured when the ``Measures`` instance is
        constructed and is used to calculate both portfolio and benchmark
        excess returns.
    
        Returns:
            The estimated bull-market SFM beta, or ``math.nan`` if the
            conditional benchmark excess returns have zero variance.
        """
        #return self._excess_covariance.beta_bull
        return self._sfm_regression.beta_bull

    @property
    def sfm_beta_bear(self) -> float:
        """
        Single-Factor Model (SFM) Bear Beta coefficient.
    
        Bear beta is the least-squares slope from the regression of portfolio
        excess returns on benchmark excess returns, restricted to observations
        for which the benchmark excess return is negative:
    
            R_b - R_f < 0
    
        The conditional regression is:
    
            R_p - R_f = alpha_bear + beta_bear * (R_b - R_f) + epsilon
    
        Bear beta measures the portfolio's sensitivity to the benchmark during
        periods in which the benchmark return is below the risk-free rate.
    
        The risk-free rate is configured when the ``Measures`` instance is
        constructed and is used to calculate both portfolio and benchmark
        excess returns.
    
        Returns:
            The estimated bear-market SFM beta, or ``math.nan`` if the
            conditional benchmark excess returns have zero variance.
        """
        #return self._excess_covariance.beta_bear
        return self._sfm_regression.beta_bear

    @property
    def timing_ratio(self) -> float:
        """
        Timing Ratio based on conditional Single-Factor Model betas.
    
        The Timing Ratio compares the portfolio's sensitivity to the benchmark
        in rising and falling benchmark markets:
    
            Timing Ratio = beta_bull / beta_bear
    
        where ``beta_bull`` is estimated using observations for which the
        benchmark excess return is positive and ``beta_bear`` is estimated
        using observations for which the benchmark excess return is negative.
    
        A value greater than 1 indicates relatively greater benchmark
        sensitivity in rising markets than in falling markets. This is
        generally interpreted as favorable asymmetry for an investor seeking
        greater participation in rising markets and lower participation in
        falling markets.
    
        The risk-free rate is configured when the ``Measures`` instance is
        constructed and determines the benchmark excess return used to
        classify observations as bull or bear.
    
        Returns:
            The ratio of bull-market beta to bear-market beta, or ``math.nan``
            if the bear beta is zero or either conditional beta cannot be
            estimated.
        """
        #denom = self._excess_covariance.beta_bear
        #return self._excess_covariance.beta_bull / denom if denom != 0 else math.nan
        denom = self._sfm_regression.beta_bear
        return self._sfm_regression.beta_bull / denom if denom != 0 else math.nan

    @property
    def sfm_r2(self) -> float:
        """
        Coefficient of determination (R²) of the Single-Factor Model.
    
        R² measures the proportion of the variance in portfolio excess
        returns explained by the linear relationship with benchmark excess
        returns.
    
        For a single-factor OLS regression with an intercept:
    
            R^2 = Corr(R_p - R_f, R_b - R_f)^2
    
        where the correlation is calculated over the same observations used
        by the regression.
    
        The risk-free rate is configured when the ``Measures`` instance is
        constructed and is used to calculate both portfolio and benchmark
        excess returns.
    
        Returns:
            The coefficient of determination in the range [0, 1], or
            ``math.nan`` if it cannot be estimated.
        """
        #cov = self._excess_covariance.value
        #var_r = self._excess_returns_kbn.variance
        #var_b = self._benchmark_excess_returns_kbn.variance
        #denom = var_r * var_b
        #return (cov * cov) / denom if denom > 0 else 0
        return self._sfm_regression.r2

    @property
    def jensen_alpha(self) -> float:
        """
        Annualized Jensen's Alpha.

        Measures the portfolio's annualized geometric return in excess of
        the annualized return predicted by the single-factor model:

            Jensen Alpha = R_p_ann - (R_f_ann + Beta * (R_b_ann - R_f_ann))
        where:

            R_p_ann  = annualized portfolio geometric return
            R_b_ann  = annualized benchmark geometric return
            R_f_ann  = annualized risk-free rate
            Beta     = SFM beta

        Unlike ``sfm_alpha``, which is the intercept of an arithmetic
        regression of excess returns, Jensen's Alpha is calculated from
        geometrically compounded returns.

        The result is expressed as an annualized return, regardless of the
        frequency of the input observations. Annualization uses the
        ``periods_per_annum`` configured for the ``Measures`` instance.

        The risk-free rate is configured when the ``Measures`` instance is
        constructed.

        Returns:
            Annualized Jensen's Alpha.
        """
        rf = self._annual_risk_free_rate
        mean = self._cumulative_return.annualized_geometric_mean_return( \
            periods_per_year=self.periods_per_annum)
        mean_b = self._benchmark_cumulative_return.annualized_geometric_mean_return( \
            periods_per_year=self.periods_per_annum)
        return mean - (rf +  self.sfm_beta * (mean_b - rf))

    @property
    def fama_beta(self) -> float:
        """
        Fama Beta.
    
        Measures the portfolio's total risk relative to the benchmark's total
        risk:
    
            Fama Beta = sigma_P / sigma_B
    
        where:
    
            sigma_P = portfolio population standard deviation
            sigma_B = benchmark population standard deviation
    
        Unlike ``sfm_beta``, which measures systematic risk through the
        portfolio's covariance with the benchmark, Fama Beta compares their
        total volatility. It is used in Fama's decomposition of portfolio
        performance and can be used to assess the loss of diversification.
    
        Portfolio and benchmark returns are assumed to have matching
        periodicity. Consequently, their standard deviations are not
        annualized; annualization factors would be identical and cancel in
        the ratio.
    
        A value greater than 1 indicates that the portfolio has greater total
        volatility than the benchmark, while a value less than 1 indicates
        lower total volatility.
    
        Returns:
            The ratio of portfolio to benchmark population standard deviation,
            or ``math.nan`` if the benchmark standard deviation is zero.
        """
        sigma = self._returns_kbn.standard_deviation_ddof_0
        sigma_b = self._benchmark_returns_kbn.standard_deviation_ddof_0
        
        return sigma / sigma_b if sigma_b != 0 else math.nan

    @property
    def modigliani(self) -> float:
        """
        Modigliani–Modigliani measure (M²).
    
        Measures the portfolio's risk-adjusted return after scaling the
        portfolio to have the same total risk, measured by standard
        deviation, as the benchmark.
    
        The measure is defined as:
    
            M^2 = R_f + Sharpe * sigma_b
    
        or equivalently:
    
            M^2 = R_f + (R_p - R_f) * (sigma_b / sigma_p)
    
        where:
    
            R_p     = portfolio return
            R_f     = risk-free rate
            sigma_p = portfolio standard deviation
            sigma_b = benchmark standard deviation
    
        M^2 can be interpreted as the return that a hypothetically leveraged
        or deleveraged version of the portfolio would have earned if its
        standard deviation were equal to that of the benchmark.
    
        Unlike the Sharpe ratio, which expresses excess return per unit of
        total risk, the Modigliani–Modigliani measure is expressed in return
        units. This makes the risk-adjusted performance directly comparable
        with portfolio and benchmark returns.
    
        The risk-free rate is configured when the ``Measures`` instance is
        constructed.
    
        Returns:
            The portfolio's risk-adjusted return scaled to the benchmark's
            standard deviation, or ``math.nan`` if the Sharpe ratio is
            undefined.
        """
        sigma = self._excess_returns_kbn.standard_deviation_ddof_0
        if sigma == 0:
            return math.nan
        sigma_b = self._benchmark_returns_kbn.standard_deviation_ddof_0
        
        # M^2 = Rf + (Rp - Rf) * (sugma_b / sigma)
        return self.risk_free_rate + self._excess_returns_kbn.mean * sigma_b / sigma

    @property
    def tracking_error(self) -> float:
        """
        Annualized Tracking Error.

        Measures the volatility of the portfolio's active returns relative
        to the benchmark:

            Tracking Error = std(R_P - R_B) * sqrt(periods_per_annum)

        where ``std`` is the sample standard deviation (``ddof=1``),
        ``periods_per_annum`` specifies the number of return observations
        per year.

        Tracking Error measures the variability of the portfolio's active
        performance. A lower value indicates that the portfolio's returns
        tend to track the benchmark more closely.

        Returns:
            Annualized standard deviation of active returns.
        """
        return self._active_returns_kbn.standard_deviation_ddof_1 * self._sqrt_periods_per_annum

    @property
    def active_premium(self) -> float:
        """
        Annualized Active Premium.

        Measures the difference between the portfolio and benchmark
        annualized geometric returns:

            Active Premium = G_P,ann - G_B,ann

        where ``G_P,ann`` and ``G_B,ann`` are the annualized geometric
        returns of the portfolio and benchmark. The portfolio and benchmark
        returns are annualized separately before taking their difference.

        A positive value indicates that the portfolio outperformed the
        benchmark on an annualized geometric-return basis.

        Returns:
            The difference between portfolio and benchmark annualized
            geometric mean returns.
        """
        mean = self._cumulative_return.annualized_geometric_mean_return( \
            periods_per_year=self.periods_per_annum)
        mean_b = self._benchmark_cumulative_return.annualized_geometric_mean_return( \
            periods_per_year=self.periods_per_annum)
        return mean - mean_b

    @property
    def information_ratio(self) -> float:
        """
        Annualized Information Ratio.
    
        Measures annualized active return relative to annualized Tracking
        Error:

            Information Ratio = Active Premium / Tracking Error

        Here, Active Premium is the difference between portfolio and
        benchmark annualized geometric mean returns, while Tracking Error is the
        annualized standard deviation of active returns.

        A higher positive value indicates greater active return relative
        to the variability of active performance.

        Returns:
            The annualized Information Ratio, or ``math.nan`` if annualized
            Tracking Error is zero.
        """
        te = self.tracking_error
        return self.active_premium / te if te != 0 else math.nan

    @property
    def information_ratio_modified(self) -> float:
        """
        Modified Information Ratio (Israelson).

        A sign-adjusted version of the Information Ratio that avoids the
        undesirable interpretation of a higher Information Ratio when
        active return is negative and tracking error increases.

        The conventional Information Ratio is::

            IR = active_premium / tracking_error

        The modified Information Ratio is::

            MIR = IR,   if active_premium > 0
                  -IR,  otherwise

        Thus, positive active performance retains the conventional
        Information Ratio, while negative active performance is reported
        with the opposite sign so that increasing tracking error cannot
        make a negatively performing strategy appear less unfavorable.

        Returns:
            float: Modified Information Ratio, or ``math.nan`` when the
            active premium or Information Ratio is unavailable.
        """
        # Active premium = mean(Rp - Rb)
        excess = self._active_returns_kbn.mean
        ir = self.information_ratio
        if math.isnan(excess) or math.isnan(ir):
            return math.nan
        return ir if excess > 0 else -ir

    @property
    def systematic_risk(self) -> float:
        """
        Annualized systematic risk.

        Measures the volatility of the strategy's systematic component
        attributable to its exposure to benchmark excess returns in the
        single-factor model::

            Rp - Rf = alpha + beta * (Rb - Rf) + epsilon

        Systematic risk is calculated as the absolute SFM beta multiplied
        by the annualized standard deviation of benchmark excess returns::

            systematic_risk = abs(beta) * sigma(Rb - Rf)

        The beta sign is retained by ``sfm_beta`` to describe the direction
        of benchmark exposure, but systematic risk is non-negative because
        risk measures the magnitude rather than the direction of exposure.

        Returns:
            float: Annualized systematic risk, or ``math.nan`` when the
            benchmark excess-return volatility or beta is unavailable.
        """
        beta = self.sfm_beta
        if math.isnan(beta):
            return math.nan
        benchmark_risk = self._benchmark_excess_returns_kbn.standard_deviation_ddof_1
        if math.isnan(benchmark_risk):
            return math.nan
        return abs(beta) * benchmark_risk * self._sqrt_periods_per_annum

    @property
    def treynor_ratio(self) -> float:
        """
        Annualized Treynor Ratio.

        Measures annualized geometric excess return per unit of systematic
        risk:

            Treynor Ratio = G_ann(R_P - R_f) / beta

        where ``G_ann(R_P - R_f)`` is the annualized geometric return of
        the portfolio's per-period excess returns and ``beta`` is the
        portfolio's SFM beta relative to the benchmark.

        Unlike the Sharpe ratio, which uses total volatility as the risk
        measure, the Treynor Ratio uses systematic risk represented by beta.

        Returns:
            The annualized Treynor Ratio, or ``math.nan`` if beta is zero.
        """
        beta = self.sfm_beta
        if beta == 0:
            return math.nan
        return self._cumulative_excess_return.annualized_geometric_mean_return( \
            periods_per_year=self.periods_per_annum) / beta

    @property
    def treynor_ratio_modified(self) -> float:
        """
        Modified Treynor Ratio.

        Measures the annualized geometric excess return earned per unit
        of annualized systematic risk.

        Unlike the conventional Treynor Ratio, which divides excess return
        by beta, the modified ratio divides excess return by systematic
        risk, defined as benchmark volatility multiplied by beta.

        Formula::

            MTR = annualized_geometric_mean(excess_return) / systematic_risk

        where::

            systematic_risk = beta * benchmark_standard_deviation_annualized

        The excess return is annualized using the configured
        ``periods_per_annum``.

        Returns:
            float: Modified Treynor Ratio, or ``math.nan`` when systematic
            risk cannot be calculated or is zero.
        """
        sr = self.systematic_risk
        if sr == 0:
            return math.nan

        return self._cumulative_excess_return.annualized_geometric_mean_return( \
            periods_per_year=self.periods_per_annum) / sr

    @property
    def specific_risk(self) -> float:
        """
        Annualized specific (residual) risk.

        Measures the volatility of the residual return that remains after
        accounting for the strategy's systematic benchmark exposure.

        For the single-factor model::

            Rp = alpha + beta * Rb + epsilon

        the residual is::

            epsilon = Rp - (alpha + beta * Rb)

        Specific risk is the annualized standard deviation of the residuals.
        It represents idiosyncratic or strategy-specific risk not explained
        by the benchmark factor.

        Returns:
            float: Annualized specific risk, or ``math.nan`` when the
            strategy and benchmark data or regression parameters are
            unavailable.
        """
        r_p = self._returns
        r_b = self._returns_benchmark
        if r_p is None or r_b is None:
            return math.nan
        beta = self.sfm_beta
        if math.isnan(beta):
            return math.nan
        alpha = self.sfm_alpha
        if math.isnan(alpha):
            return math.nan
        epsilon_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=0, bias=True, fisher=True)
        rf = self.risk_free_rate
        for i, r in enumerate(r_p):
            #epsilon_kbn.update(r - r_b[i] * beta - alpha)
            epsilon_kbn.update(r - rf - alpha - beta *(r_b[i] - rf))
        return epsilon_kbn.standard_deviation_ddof_0 * self._sqrt_periods_per_annum

    @property
    def total_risk(self) -> float:
        """
        Annualized total risk from systematic and specific risk.

        Total risk combines the systematic and specific components of the
        single-factor return model.

        Under the standard regression decomposition, the components are
        orthogonal and therefore their variances add::

            total_risk = sqrt(systematic_risk**2 + specific_risk**2)

        Both systematic and specific risk are annualized before being
        combined.

        Returns:
            float: Annualized total risk, or ``math.nan`` when either
            component cannot be calculated.
        """
        syr = self.systematic_risk
        if math.isnan(syr):
            return math.nan
        spr = self.specific_risk
        if math.isnan(spr):
            return math.nan
        return math.sqrt(syr * syr + spr * spr)

    @property
    def appraisal_ratio(self) -> float:
        """
        Appraisal Ratio.

        Measures Jensen's alpha earned per unit of specific (residual) risk.

        Specific risk is the portion of strategy risk not explained by the
        systematic benchmark exposure. The ratio is therefore a measure of
        the efficiency with which the strategy generates benchmark-adjusted
        alpha relative to its idiosyncratic risk.

        Formula::

            appraisal_ratio = jensen_alpha / specific_risk

        A higher positive value indicates more alpha generated for each unit
        of specific risk.

        Returns:
            float: Appraisal Ratio, or ``math.nan`` when Jensen's alpha or
            specific risk is unavailable, or when specific risk is zero.
        """
        alpha = self.jensen_alpha
        if math.isnan(alpha):
            return math.nan
        spr = self.specific_risk
        return alpha / spr if spr != 0 else math.nan

    @property
    def jensen_alpha_modified(self) -> float:
        """
        Modified Jensen's Alpha.

        Measures Jensen's alpha per unit of systematic benchmark exposure.

        Unlike conventional Jensen's alpha, which is expressed as a return,
        the modified measure normalizes alpha by the strategy's SFM beta.

        Formula::

            modified_jensen_alpha = jensen_alpha / beta

        The measure can be interpreted as the alpha generated per unit of
        systematic benchmark exposure.

        Returns:
            float: Modified Jensen's alpha, or ``math.nan`` when Jensen's
            alpha is unavailable or beta is zero.
        """
        alpha = self.jensen_alpha
        if math.isnan(alpha):
            return math.nan
        beta = self.fama_beta
        return alpha / beta if beta != 0 else math.nan

    @property
    def jensen_alpha_alternative(self) -> float:
        """
        Alternative Jensen's Alpha.

        Measures Jensen's alpha per unit of systematic risk.

        Systematic risk represents the volatility attributable to the
        strategy's exposure to the benchmark/factor.

        Formula::

            alternative_jensen_alpha = jensen_alpha / systematic_risk

        Unlike ``jensen_alpha_modified``, which divides alpha by beta,
        this measure normalizes alpha by the magnitude of systematic risk
        expressed in return units.

        Returns:
            float: Alternative Jensen's alpha, or ``math.nan`` when
            Jensen's alpha or systematic risk is unavailable, or when
            systematic risk is zero.
        """
        alpha = self.jensen_alpha
        if math.isnan(alpha):
            return math.nan
        spr = self.systematic_risk
        return alpha / spr if spr != 0 else math.nan

    @property
    def m_squared(self) -> float:
        """
        Estimate the M-squared (M²) risk-adjusted return.

        M-squared transforms the portfolio return to the return that the
        portfolio would have earned if its volatility were equal to the
        benchmark volatility, while retaining the portfolio's risk-free
        rate.

        The calculation uses geometric annualized portfolio return and
        annualized population standard deviations:

            M² = R_P + (R_P - R_f) * (σ_B / σ_P) - (R_P - R_f)
               = (R_P - R_f) * σ_B / σ_P + R_f

        where:

            R_P = annualized geometric portfolio return
            R_f = annual risk-free rate
            σ_P = annualized population standard deviation of portfolio
                  returns
            σ_B = annualized population standard deviation of benchmark
                  returns

        Returns:
            The M-squared return, or ``math.nan`` if the required return
            or volatility statistics are unavailable or portfolio
            volatility is zero.
        """
        p_ret = self._cumulative_return.annualized_geometric_mean_return(self.periods_per_annum)
        if math.isnan(p_ret):
            return math.nan
        p_std = self._returns_kbn.standard_deviation_ddof_0 * self._sqrt_periods_per_annum
        if math.isnan(p_std) or p_std == 0:
            return math.nan
        b_std = self._benchmark_returns_kbn.standard_deviation_ddof_0 * self._sqrt_periods_per_annum
        if math.isnan(b_std):
            return math.nan

        return (p_ret - self._annual_risk_free_rate) * b_std / p_std + self._annual_risk_free_rate

    @property
    def m_squared_excess(self) -> float:
        """
        Estimate the geometric excess M-squared return.

        M-squared excess measures the geometric annualized return of the
        M-squared portfolio relative to the annualized geometric return
        of the benchmark.

        The geometric formulation is:

            M²_excess = (1 + M²) / (1 + R_B) - 1

        where:

            M² = M-squared return
            R_B = annualized geometric benchmark return

        Returns:
            The geometric M-squared excess return, or ``math.nan`` if
            M-squared or the benchmark return is unavailable.
        """
        m_sq = self.m_squared
        if math.isnan(m_sq):
            return math.nan
        b_ret = self._benchmark_cumulative_return.annualized_geometric_mean_return(self.periods_per_annum)
        if math.isnan(b_ret):
            return math.nan
        return (1.0 + m_sq) / (1.0 + b_ret) - 1.0

    @property
    def m_squared_sortino(self) -> float:
        """
        Estimate M-squared using downside risk (M² Sortino).

        M² Sortino transforms the portfolio return to the return that the
        portfolio would have earned if its downside risk were equal to
        the benchmark's downside risk.

        The calculation is:

            M²_S = R_P + SR_P * (σ_DM - σ_D)

        where:

            R_P   = annualized geometric portfolio return
            SR_P  = portfolio Sortino ratio
            σ_DM  = annualized benchmark downside deviation
            σ_D   = annualized portfolio downside deviation

        Downside deviation is measured relative to the configured MAR.

        Returns:
            The M² Sortino return, or ``math.nan`` if the required
            statistics are unavailable.
        """
        sortino = self.sortino_ratio
        if math.isnan(sortino):
            return math.nan
        p_ret = self._cumulative_return.annualized_geometric_mean_return(self.periods_per_annum)
        if math.isnan(p_ret):
            return math.nan
        p_dd = self.downside_deviation
        if math.isnan(p_dd):
            return math.nan

        b_count = self._benchmark_target_partial_moments.total_count
        if b_count == 0:
            return math.nan
        b_dd = math.sqrt(self._benchmark_target_partial_moments.lower_excess_moment_2_sum / b_count)
        if math.isnan(b_dd):
            return math.nan

        return p_ret + sortino * self._sqrt_periods_per_annum * (b_dd - p_dd)

    def tail_ratio(self, cutoff: float = 0.95) -> float: # NOT IN R
        """
        Tail Ratio.

        Measures the relative magnitude of the empirical upper and lower
        tails of the return distribution.

        The Tail Ratio is calculated as the selected upper percentile divided
        by the absolute value of the corresponding lower percentile::

            tail_ratio = Q(cutoff) / abs(Q(1 - cutoff))

        For the default ``cutoff=0.95`` this is the ratio of the 95th
        percentile return to the absolute value of the 5th percentile return.

        The measure describes tail asymmetry:

            * ``tail_ratio > 1``:
              the upper tail is larger in magnitude than the lower tail.

            * ``tail_ratio = 1``:
              the selected upper and lower tails have equal magnitude.

            * ``tail_ratio < 1``:
              the lower tail is larger in magnitude than the upper tail.

        A Tail Ratio above one does not necessarily indicate better
        performance, nor does a ratio below one necessarily indicate poor
        performance. The measure describes the shape of the return
        distribution and should be interpreted together with return,
        volatility, downside risk, and drawdown measures.

        The choice of ``cutoff`` determines which part of the distribution
        is examined. For example, ``0.90`` compares the 90th and 10th
        percentiles, while ``0.95`` compares the 95th and 5th percentiles.
        More extreme cutoffs focus more strongly on tail behavior but can
        be less statistically stable for shorter return histories.

        The result is not annualized.

        Args:
            cutoff: Upper-tail percentile expressed as a fraction in the
                range ``(0.5, 1)``. The corresponding lower-tail percentile
                is ``1 - cutoff``. Defaults to ``0.95``.

        Returns:
            float: Tail Ratio, or ``math.nan`` when there are insufficient
            observations or the lower percentile is zero.
        """
        if not 0.5 < cutoff < 1.0:
            raise ValueError("cutoff must be between 0.5 and 1.0")
        w = self._returns
        if w is None or len(w) < 2:
            return math.nan

        right_tail = core.percentile(w, cutoff)
        left_tail = core.percentile(w, 1 - cutoff)
        return right_tail / abs(left_tail) if left_tail != 0 else math.nan

    @property
    def kelly_ratio_full(self) -> float:
        """
        Full Kelly criterion fraction based on the mean and variance of
        excess returns.

        The full Kelly fraction is the theoretically optimal constant
        leverage or bet size under the classical Kelly framework:

            f* = E[R - Rf] / Var(R - Rf)

        where:

            R  = portfolio return,
            Rf = risk-free rate configured for the ``Measures`` instance.

        The ratio therefore represents the fraction of available capital
        that the classical Kelly criterion would allocate to the strategy
        under its assumptions. A value of ``1.0`` corresponds to 100%
        capital allocation, while values greater than 1 imply leverage.

        The calculation uses the arithmetic mean and variance of the
        observed excess returns. It is the "full Kelly" allocation;
        practical implementations often use a fraction of Kelly to reduce
        sensitivity to estimation error and model assumptions.

        Returns:
            float:
                Full Kelly leverage/bet-size fraction, or ``math.nan`` when
                the variance of excess returns is zero.
        """
        mean_excess = self._excess_returns_kbn.mean
        var_excess = self._excess_returns_kbn.variance        
        return mean_excess / var_excess if var_excess != 0 else math.nan

    @property
    def kelly_ratio(self) -> float:
        """
        Half-Kelly criterion fraction based on excess-return mean and variance.
    
        The Half-Kelly fraction is one-half of the full Kelly allocation:
    
            f*_{1/2} = 1/2 * E[R - Rf] / Var(R - Rf)
    
        where:
    
            R  = portfolio return,
            Rf = risk-free rate configured for the ``Measures`` instance.
    
        Half-Kelly is commonly used as a more conservative alternative to
        full Kelly because the theoretically optimal Kelly fraction can be
        highly sensitive to estimation error in the expected return and
        variance.
    
        A value of ``1.0`` corresponds to 100% capital allocation, while
        values greater than 1 imply leverage.
    
        Returns:
            float:
                Half-Kelly leverage/bet-size fraction, or ``math.nan`` when
                the variance of excess returns is zero.
        """
        return self.kelly_ratio_full / 2

    @property
    def hurst_exponent(self) -> float:
        """
        Hurst Exponent (Rescaled Range Analysis).

        Estimates the long-term dependence of the return series using
        rescaled range (R/S) analysis.

        The estimate is calculated as::

            H = log(R / S) / log(N)

        where ``N`` is the number of observations, ``S`` is the standard
        deviation of returns, and ``R`` is the range of the cumulative
        demeaned return series::

            R = max(Z_t) - min(Z_t)
            Z_t = sum(R_i - mean(R))

        The Hurst Exponent is commonly interpreted as follows:

            * ``H > 0.5``:
              persistent behavior, where positive (negative) returns tend
              to be followed by returns of the same sign. This is often
              associated with trending or long-memory behavior.

            * ``H ~= 0.5``:
              behavior consistent with a random walk and little evidence
              of long-term dependence.

            * ``H < 0.5``:
              anti-persistent behavior, where observations tend to reverse
              direction. This is often associated with mean-reverting
              behavior.

        The result is a single-sample R/S estimate. Unlike multi-scale
        Hurst estimation methods, it does not estimate the slope of a
        regression of log(R/S) against log(N) over multiple time scales.

        The Hurst Exponent is a statistical property of the return series,
        not a risk-adjusted performance measure. Values above or below 0.5
        should not by themselves be interpreted as evidence that a trading
        strategy is profitable or that a particular trading rule will work.

        The estimate is not annualized.

        Returns:
            float:
                Estimated Hurst Exponent, or ``math.nan`` when there are
                insufficient observations, zero return volatility, or the
                rescaled range is undefined.
        """
        n = self._returns_kbn.n
        w = self._returns
        if w is None or n < 2:
            return math.nan

        mean = self._returns_kbn.mean
        std = self._returns_kbn.standard_deviation_ddof_1
        if std == 0:
            return math.nan
        cum_sum: KleinKBNAccumulator = KleinKBNAccumulator()
        cum_min = math.inf
        cum_max = -math.inf
        for x in w:
            cum_sum.update(x - mean) # Demean
            val = cum_sum.value
            if cum_min > val:
                cum_min = val
            if cum_max < val:
                cum_max = val
        delta = cum_max - cum_min
        rescaled_range = delta / std
        if rescaled_range <= 0:
            return math.nan
        # Hurst exponent
        return math.log(rescaled_range) / math.log(n)

    def bias_ratio(self, std_dev_multiplier: float = 1.0) -> float: # RECALCULATES WINDOW
        """
        Bias Ratio.

        Measures the asymmetry of returns close to zero. The Bias Ratio was
        introduced by Adil Abdulali (2006) as a diagnostic for detecting
        stale pricing and return smoothing, particularly in illiquid assets
        and alternative investment strategies.

        The ratio compares the number of small non-negative returns with the
        number of small negative returns, where "small" is defined relative
        to the current standard deviation::

            threshold = std_dev_multiplier * std_dev
            bias_ratio = count(0 <= R <= threshold) / (1 + count(-threshold <= R < 0))

        The ``+1`` in the denominator prevents division by zero.

        A ratio below one indicates that small negative returns occur more
        frequently than small non-negative returns. Values substantially
        above one may indicate an unusually large concentration of returns
        immediately above zero, which can be a warning sign of stale pricing,
        return smoothing, or other distortions in the observed return series.

        The Bias Ratio is a distributional diagnostic rather than a
        risk-adjusted return measure. It does not measure profitability or
        risk directly and should therefore be interpreted together with
        other distribution, volatility, and liquidity diagnostics.

        The threshold is based on the standard deviation of the current
        return window. Consequently, changing ``std_dev_multiplier`` changes
        which observations are classified as being close to zero.

        Because the threshold depends on the current standard deviation,
        the complete return window must be recalculated whenever the ratio
        is requested. This property is therefore not a streaming calculation.

        The result is not annualized.

        Args:
            std_dev_multiplier:
                Number of standard deviations used to define the interval
                around zero. For example, ``1.0`` compares returns within
                one standard deviation of zero, while ``2.0`` uses a
                two-standard-deviation interval. Must be positive.
                Defaults to ``1.0``.

        Returns:
            float:
                Bias Ratio, or ``math.nan`` when there are insufficient
                observations, the standard deviation is undefined or zero,
                or the multiplier is invalid.
        """
        if std_dev_multiplier <= 0:
            raise ValueError("std_dev_multiplier must be positive")
        w = self._returns
        if w is None:
            return math.nan

        std = self._returns_kbn.standard_deviation_ddof_1
        if math.isnan(std) or std == 0:
            return math.nan
        threshold = std_dev_multiplier * std

        count_positive = 0
        count_negative = 0
        for x in w:
            if 0 <= x <= threshold:
                count_positive += 1
            elif -threshold <= x < 0:
                count_negative += 1

        return count_positive / (1 + count_negative)

    @property
    def k_ratio(self) -> float: # RECALCULATES WINDOW
        """
        K-Ratio (Lars Kestner).

        Measures the consistency and statistical strength of the growth of an
        investment's equity curve. The K-Ratio is calculated by regressing the
        cumulative log return series against time and standardizing the slope
        of that regression::

            E_t = sum(log(1 + R_i))
            K = slope / (SE(slope) * sqrt(N))

        where ``E_t`` is the cumulative log return at observation ``t``,
        ``slope`` is the estimated slope of the linear regression of ``E_t``
        on time, ``SE(slope)`` is its standard error, and ``N`` is the number
        of observations.

        The K-Ratio therefore measures whether the equity curve exhibits a
        strong and consistent upward trend. A high positive value indicates
        a steadily rising equity curve, while a value near zero indicates
        little evidence of a consistent trend. A negative value indicates
        a declining equity curve.

        Unlike the Sharpe Ratio, which evaluates return relative to the
        volatility of individual returns, the K-Ratio evaluates the
        consistency of the accumulated equity curve over time. It is therefore
        primarily a measure of equity-curve quality and path consistency.

        The measure is not annualized.

        This calculation is performed over the current return window and is
        recalculated from the cumulative log-return series. It is therefore
        not a streaming calculation.
    
        Returns:
            float:
                K-Ratio, or ``math.nan`` when there are insufficient
                observations or the regression standard error cannot be
                calculated.
        """
        n = self._returns_kbn.n
        w = self._returns
        if w is None or n < 3:
            return math.nan
        
        # Build equity curve (cumulative log returns)
        # equity[i] = sum(log1p(w[0])...log1p(w[i]))
        equity = []
        cum_sum = 0.0
        for x in w:
            # Use math.log1p for precision with small returns
            cum_sum += math.log1p(x)
            equity.append(cum_sum)
        
        # Linear regression statistics: equity = a + b * t
        # We need: sum(t), sum(t^2), sum(equity), sum(t*equity)
        # t is simply 0, 1, 2, ..., n-1
        sum_t = 0.0
        sum_t2 = 0.0
        sum_eq = 0.0
        sum_te = 0.0
        for i, eq_val in enumerate(equity):
            t_val = float(i)
            sum_t += t_val
            sum_t2 += t_val * t_val
            sum_eq += eq_val
            sum_te += t_val * eq_val

        # Means
        t_mean = sum_t / n
        equity_mean = sum_eq / n

        # S_tt = sum((t - t_mean)^2) = sum(t^2) - n * t_mean^2
        S_tt = sum_t2 - n * (t_mean ** 2)

        # S_te = sum((t - t_mean) * (eq - eq_mean)) = sum(t*eq) - n * t_mean * eq_mean
        S_te = sum_te - n * t_mean * equity_mean
        if S_tt == 0:
            return math.nan

        # Slope (b)
        slope = S_te / S_tt
    
        # Intercept (a) = eq_mean - b * t_mean
        intercept = equity_mean - slope * t_mean

        # Residuals and Variance
        # residual_var = sum((eq - (a + b*t))^2) / (n - 2)
        sum_sq_residuals = 0.0
        for i, eq_val in enumerate(equity):
            t_val = float(i)
            predicted = intercept + slope * t_val
            residual = eq_val - predicted
            sum_sq_residuals += residual * residual
        if n <= 2:
            return math.nan
        residual_var = sum_sq_residuals / (n - 2)

        # Standard Error of Slope
        # se_slope = sqrt(residual_var / S_tt)
        if residual_var < 0: # Handle floating point noise
            residual_var = 0.0
        se_slope = math.sqrt(residual_var / S_tt)
        if se_slope == 0:
            return math.nan

        # K-Ratio = slope / (se_slope * sqrt(n))
        return  slope / (se_slope * math.sqrt(n))

    @property
    def gain_to_pain_ratio(self):
        """
        Jack Schwager's Gain-to-Pain Ratio.

        GPR is the net return divided by the total magnitude of
        negative returns:

            GPR = mean(R) / LPM_1(0)

        where ``LPM_1(0)`` is the first-order lower partial moment
        about zero.

        Equivalently:

            GPR = sum(R) / sum(max(-R, 0))

        The ratio measures how much net return is generated for
        each unit of loss incurred. Higher values indicate more
        favorable return generation relative to losses.

        The calculation uses zero as the loss threshold and is
        therefore independent of the configured target return
        (MAR) and risk-free rate.

        See here for more info: https://archive.is/wip/2rwFW
        """
        lpm1 = self._raw_partial_moments.lower_partial_moment_1
        if math.isnan(lpm1) or lpm1 == 0:
            return math.nan
        return self._returns_kbn.mean / lpm1

    def upside_capture_ratio(self, geometric: bool = True) -> float:
        """
        Upside Capture Ratio.

        Measures how much of the benchmark's positive-market performance the
        investment captured during periods when the benchmark return was
        positive.

        The geometric version compares cumulative compounded returns::

            upside_capture =
                cumulative_return_asset_up / cumulative_return_benchmark_up

        The arithmetic version compares the sums (equivalently, the means)
        of periodic returns during the same periods::

            upside_capture =
                sum(asset_returns_up) / sum(benchmark_returns_up)

        A value of ``1.0`` indicates that the investment matched the
        benchmark's upside performance. Values above ``1.0`` indicate that
        the investment captured more upside than the benchmark, while values
        below ``1.0`` indicate that it captured less.

        Higher values are generally better.

        Args:
            geometric:
                If ``True``, use compounded geometric returns. If ``False``,
                use arithmetic returns. Defaults to ``True``.

        Returns:
            float:
                Upside Capture Ratio, or ``math.nan`` when the benchmark has
                no positive-return periods or its aggregate upside return is
                zero.
        """
        if geometric == True:
            return self._capture.upside_capture_ratio_geometric
        return self._capture.upside_capture_ratio_arithmetic

    def downside_capture_ratio(self, geometric: bool = True) -> float:
        """
        Downside Capture Ratio.

        Measures how much of the benchmark's negative-market performance the
        investment experienced during periods when the benchmark was
        non-positive.

        The geometric version compares cumulative compounded returns::

            downside_capture =
                cumulative_return_asset_down / cumulative_return_benchmark_down

        The arithmetic version compares the sums (equivalently, the means)
        of periodic returns during the same periods::

            downside_capture =
                sum(asset_returns_down) / sum(benchmark_returns_down)

        A value of ``1.0`` indicates that the investment matched the
        benchmark's downside performance. Values below ``1.0`` indicate that
        the investment suffered less downside than the benchmark, while
        values above ``1.0`` indicate greater downside.

        Lower values are generally better.

        Args:
            geometric:
                If ``True``, use compounded geometric returns. If ``False``,
                use arithmetic returns. Defaults to ``True``.

        Returns:
            float:
                Downside Capture Ratio, or ``math.nan`` when the benchmark
                has no qualifying downside periods or its aggregate downside
                return is zero.
        """
        if geometric == True:
            return self._capture.downside_capture_ratio_geometric
        return self._capture.downside_capture_ratio_arithmetic

    def overall_capture_ratio(self, geometric: bool = True) -> float:
        """
        Overall Capture Ratio.

        Measures the investment's relative ability to capture benchmark
        upside while limiting benchmark downside. It is calculated as the
        ratio of the Upside Capture Ratio to the Downside Capture Ratio::

            overall_capture = upside_capture / downside_capture

        A higher value generally indicates a more favorable combination of
        upside participation and downside protection.

        The calculation uses either geometric or arithmetic capture ratios,
        as selected by ``geometric``.

        Args:
            geometric:
                If ``True``, use geometric capture ratios. If ``False``, use
                arithmetic capture ratios. Defaults to ``True``.

        Returns:
            float:
                Overall Capture Ratio, or ``math.nan`` when either component
                is undefined or the Downside Capture Ratio is zero.
        """
        up = self.upside_capture_ratio(geometric=geometric)
        if up is None:
            return None
        down = self.downside_capture_ratio(geometric=geometric)
        if down is None:
            return None
        return up / down if down != 0 else None

    @property
    def up_number_ratio(self) -> float:
        """
        Up Number Ratio.

        Measures the frequency with which the investment had a positive
        return during periods when the benchmark had a positive return.

        The ratio is calculated as::

            up_number_ratio =
                number(asset_return > 0 and benchmark_return > 0)
                / number(benchmark_return > 0)

        A value of ``1.0`` means that the investment was positive in every
        period in which the benchmark was positive. Higher values indicate
        more frequent participation in benchmark advances.

        Unlike the Upside Capture Ratio, this measure considers only the
        direction of returns and does not consider their magnitude.

        Higher values are generally better.

        Returns:
            float:
                Up Number Ratio, or ``math.nan`` when the benchmark has no
                positive-return periods.
        """
        return self._capture.up_number_ratio

    @property
    def down_number_ratio(self) -> float:
        """
        Down Number Ratio.

        Measures the frequency with which the investment had a negative
        return during periods when the benchmark had a negative return.

        The ratio is calculated as::

            down_number_ratio =
                number(asset_return < 0 and benchmark_return < 0)
                / number(benchmark_return < 0)

        A value of ``0.0`` means that the investment had no negative returns
        during benchmark-down periods, while ``1.0`` means that it was
        negative in every such period.

        Unlike the Downside Capture Ratio, this measure considers only the
        direction of returns and does not consider their magnitude.

        Lower values are generally better.

        Returns:
            float:
                Down Number Ratio, or ``math.nan`` when the benchmark has no
                negative-return periods.
        """
        return self._capture.down_number_ratio

    @property
    def up_percentage_ratio(self) -> float:
        """
        Up Percentage Ratio.

        Measures the frequency with which the investment outperformed the
        benchmark during periods when the benchmark had a positive return.

        The ratio is calculated as::

            up_percentage_ratio =
                number(asset_return > benchmark_return
                       and benchmark_return > 0)
                / number(benchmark_return > 0)

        A value of ``1.0`` means that the investment outperformed the
        benchmark in every benchmark-up period. A value of ``0.0`` means
        that it never outperformed the benchmark during such periods.

        Higher values are generally better.

        Returns:
            float:
                Up Percentage Ratio, or ``math.nan`` when the benchmark has
                no positive-return periods.
        """
        return self._capture.up_percentage_ratio

    @property
    def down_percentage_ratio(self) -> float:
        """
        Down Percentage Ratio.

        Measures the frequency with which the investment outperformed the
        benchmark during periods when the benchmark had a negative return.

        The ratio is calculated as::

            down_percentage_ratio =
                number(asset_return > benchmark_return
                       and benchmark_return < 0)
                / number(benchmark_return < 0)

        Because the benchmark is negative, outperforming it means that the
        investment's return was less negative or positive. A value of
        ``1.0`` means that the investment outperformed the benchmark in
        every benchmark-down period.

        Higher values are generally better.

        Returns:
            float:
                Down Percentage Ratio, or ``math.nan`` when the benchmark has
                no negative-return periods.
        """
        return self._capture.down_percentage_ratio
