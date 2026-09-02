from datetime import datetime
from numbers import Real
import warnings
import numpy as np
from scipy.stats import kurtosis, skew, norm

from ..daycounting import DayCountConvention, year_frac, day_frac
from .periodicity import Periodicity

_SQRT2 = 1.4142135623730950488016887242097

class Ratios:
    """
    Various financial ratios to evaluate the performance of a strategy.
    """
    def __init__(self,
        periodicity: Periodicity = Periodicity.DAILY,
        annual_risk_free_rate: float = 0.,
        annual_target_return: float = 0., # TARGET RETURN !!!!
        day_count_convention: DayCountConvention = DayCountConvention.RAW,
        rolling_window: int = None,
        min_periods: int = None):
        """
        Args:
            annual_risk_free_rate float:
                Annual risk-free rate.
                Default: 0.0
            annual_trading_days int:
                Annual trading days.
                Default: 252
            day_count_convention DayCountConvention:
                Day count convention.
                Default: DayCountConvention.RAW
        """
        self.periodicity = periodicity
        periods_per_annum = 252 if periodicity == Periodicity.DAILY \
            else 52 if periodicity == Periodicity.WEEKLY \
            else 12 if periodicity == Periodicity.MONTHLY \
            else 4 if periodicity == Periodicity.QUARTERLY \
            else 1
        self.periods_per_annum = periods_per_annum
        self.days_per_period = 1 if periodicity == Periodicity.DAILY \
            else 252 / 52 if periodicity == Periodicity.WEEKLY \
            else 252 / 12 if periodicity == Periodicity.MONTHLY \
            else 252 / 4 if periodicity == Periodicity.QUARTERLY \
            else 252

        self.risk_free_rate = annual_risk_free_rate \
            if annual_risk_free_rate == 0 or periods_per_annum == 1 \
            else ((1 + annual_risk_free_rate) ** (1/periods_per_annum) - 1)

        self.required_return = annual_target_return \
            if annual_target_return == 0 or periods_per_annum == 1 \
            else ((1 + annual_target_return) ** (1/periods_per_annum) - 1)
        self.day_count_convention = day_count_convention
        self.rolling_window = rolling_window
        self.min_periods = min_periods if min_periods is not None and min_periods > 0 \
            else None

        self.fractional_periods: np.ndarray = None
        self.returns: np.ndarray = None
        self._logret_sum: float = None
        self._drawdowns_cumulative: np.ndarray = None
        self._drawdowns_cumulative_min: float = None
        self._drawdowns_peaks: np.ndarray = None
        self._drawdowns_peaks_peak: int = None
        self._drawdown_continuous: np.ndarray = None
        self._drawdown_continuous_final: np.ndarray = None
        self._drawdown_continuous_finalized: bool = None
        self._drawdown_continuous_peak: int = None
        self._drawdown_continuous_inside: bool = None
        self._cumulative_return_plus_1: float = None
        self._cumulative_return_plus_1_100: float = None
        self._cumulative_return_geometric_mean: float = None
        self._cumulative_return_plus_1_max: float = None
        self._returns_mean: float = None
        self._returns_std: float = None
        self._returns_autocorr_penalty: float = None
        self._returns_skew: float = None
        self._returns_kurtosis: float = None
        self._rf_mean: float = None
        self._rf_std: float = None
        self._excess_mean: float = None
        self._excess_std: float = None
        self._excess_autocorr_penalty: float = None
        self._required_mean: float = None
        self._required_lpm_1: float = None
        self._required_lpm_2: float = None
        self._required_lpm_3: float = None
        self._required_hpm_1: float = None
        self._required_hpm_2: float = None
        self._required_hpm_3: float = None
        self._required_autocorr_penalty: float = None
        self._avg_return: float = None
        self._avg_win: float = None
        self._avg_loss: float = None
        self._win_rate: float = None
        self._total_duration: float = None
        self._sample_count: int = 0

        # For LPM/HPM generic calculations
        self._lpm_cache: dict = {}
        self._hpm_cache: dict = {}

        # Benchmark tracking (Phase 2)
        self.benchmark_returns: np.ndarray = None
        self._benchmark_mean: float = None
        self._benchmark_std: float = None
        self._benchmark_autocorr_penalty: float = None
        self._excess_benchmark_mean: float = None
        self._excess_benchmark_std: float = None
        self._covariance: float = None
        self._correlation: float = None
        self._beta: float = None
        self._alpha: float = None
        self._r_squared: float = None
        self._tracking_error: float = None
        self._active_premium: float = None
        self._information_ratio: float = None
        self._treynor_ratio: float = None
        self._appraisal_ratio: float = None
        self._upside_capture: float = None
        self._downside_capture: float = None
        self._capture_ratio: float = None

    def reset(self):
        self.fractional_periods = np.array([])
        self.returns = np.array([])
        self._logret_sum = 0
        self._drawdowns_cumulative = np.array([])
        self._drawdowns_cumulative_min = np.inf
        self._drawdowns_peaks = np.array([])
        self._drawdowns_peaks_peak = 0
        self._drawdown_continuous = np.array([])
        self._drawdown_continuous_final = np.array([])
        self._drawdown_continuous_finalized = False
        self._drawdown_continuous_peak = 1
        self._drawdown_continuous_inside = False
        self._cumulative_return_plus_1 = 1
        self._cumulative_return_plus_1_100 = 1
        self._cumulative_return_plus_1_max = -np.inf
        self._total_duration = 0
        self._sample_count = 0

        # Reset LPM/HPM caches
        self._lpm_cache = {}
        self._hpm_cache = {}

        # Reset benchmark tracking
        self.benchmark_returns = np.array([])
        self._benchmark_mean = None
        self._benchmark_std = None
        self._benchmark_autocorr_penalty = None
        self._excess_benchmark_mean = None
        self._excess_benchmark_std = None
        self._covariance = None
        self._correlation = None
        self._beta = None
        self._alpha = None
        self._r_squared = None
        self._tracking_error = None
        self._active_premium = None
        self._information_ratio = None
        self._treynor_ratio = None
        self._appraisal_ratio = None
        self._upside_capture = None
        self._downside_capture = None
        self._capture_ratio = None

    def add_return(self,
                   return_: float,
                   return_benchmark: float,
                   value: float,
                   time_start: datetime,
                   time_end: datetime):
        if self.periodicity == Periodicity.ANNUAL:
            fractional_period = year_frac(time_start, time_end,
                self.day_count_convention)
        else:
            fractional_period = day_frac(time_start, time_end,
                self.day_count_convention) / self.days_per_period

        self.fractional_periods = np.append(self.fractional_periods, fractional_period)
        if fractional_period == 0:
            print('Zero fractional time period, perfomance not updated')
            return
        self._total_duration += fractional_period ### DO SMTH WITH IT
        self._sample_count += 1

        # Normalized returns
        ret = return_ / fractional_period
        self.returns = np.append(self.returns, ret)

        # Window slice: use last rolling_window returns, or all if not set
        w = self.returns if self.rolling_window is None \
            else self.returns[-self.rolling_window:]
        l = len(w)

        self._returns_mean = np.mean(w)
        self._returns_std = \
            np.std(w, ddof=1) if l > 1 else None
        self._returns_autocorr_penalty = self._autocorr_penalty(w)
        # Skewness and Kurtosis for adjusted Sharpe and PSR
        if l > 2:
            self._returns_skew = skew(w, bias=True)
            self._returns_kurtosis = kurtosis(w, bias=True, fisher=True)  # excess kurtosis
        else:
            self._returns_skew = None
            self._returns_kurtosis = None

        tmp1 = w[w != 0]
        len1 = len(tmp1)
        self._avg_return = tmp1.mean() if len1 > 0 else None
        tmp2 = w[w > 0]
        len2 = len(tmp2)
        self._win_rate = len2 / len1 if len1 > 0 else None
        self._avg_win = tmp2.mean() if len2 > 0 else None
        tmp2 = w[w < 0]
        len2 = len(tmp2)
        self._avg_loss = tmp2.mean() if len2 > 0 else None

        # Excess returns (returns less risk-free rate)
        if self.risk_free_rate == 0:
            self._excess_mean = self._returns_mean
            self._excess_std = self._returns_std
            self._excess_autocorr_penalty = self._returns_autocorr_penalty
        else:
            tmp2 = w - self.risk_free_rate
            self._excess_mean = np.mean(tmp2)
            self._excess_std = np.std(tmp2, ddof=1) if l > 1 else None
            self._excess_autocorr_penalty = self._autocorr_penalty(tmp2)

        # Lower partial moments for the raw returns (less required return)
        if self.required_return == 0:
            tmp2 = -w
        else:
            tmp2 = self.required_return - w
        # Set the minimum of each to 0
        tmp2 = tmp2.clip(min=0)
        # Calculate the sum of the excess returns to the power of order
        self._required_lpm_1 = np.sum(tmp2) / l
        self._required_lpm_2 = np.sum(tmp2 ** 2) / l
        self._required_lpm_3 = np.sum(tmp2 ** 3) / l

        # Higher partial moments for the raw returns (less required return)
        if self.required_return == 0:
            tmp2 = w
            self._required_mean = self._returns_mean
            self._required_autocorr_penalty = self._returns_autocorr_penalty
        else:
            # Calculate the difference between the returns and the threshold
            tmp2 = w - self.required_return
            self._required_mean = np.mean(tmp2)
            self._required_autocorr_penalty = self._autocorr_penalty(tmp2)
        # Set the minimum of each to 0
        tmp2 = tmp2.clip(min=0)
        # Calculate the sum of the excess returns to the power of order
        self._required_hpm_1 = np.sum(tmp2) / l
        self._required_hpm_2 = np.sum(tmp2 ** 2) / l
        self._required_hpm_3 = np.sum(tmp2 ** 3) / l

        # Cumulative returns — recompute from window
        logret_sum = 0
        for j in range(len(self.returns) - l, len(self.returns)):
            fp_j = self.fractional_periods[j]
            if fp_j != 0:
                raw_ret_j = w[j - (len(self.returns) - l)]
                logret_sum += np.log(raw_ret_j + 1)
        self._logret_sum = logret_sum
        cmr = np.exp(logret_sum)
        self._cumulative_return_plus_1 = cmr
        if l >= 1:
            try:
                with warnings.catch_warnings():
                    warnings.simplefilter("error", RuntimeWarning)
                    self._cumulative_return_geometric_mean = pow(cmr, 1 / l) - 1
            except RuntimeWarning:
                pass
        self._cumulative_return_plus_1_max = -np.inf
        # Recompute running max of cumulative returns within window
        cumr = 1
        for j in range(l):
            cumr *= (w[j] + 1)
            if cumr > self._cumulative_return_plus_1_max:
                self._cumulative_return_plus_1_max = cumr

        # Drawdowns from peaks to valleys,
        # operates on cumulative returns — recompute from window.
        self._drawdowns_cumulative = np.array([])
        self._drawdowns_cumulative_min = np.inf
        cumr = 1
        cumr_max = -np.inf
        for j in range(l):
            cumr *= (w[j] + 1)
            if cumr > cumr_max:
                cumr_max = cumr
            dd = cumr / cumr_max - 1
            self._drawdowns_cumulative = np.append(self._drawdowns_cumulative, dd)
            if dd < self._drawdowns_cumulative_min:
                self._drawdowns_cumulative_min = dd

        # Different drawdown calculation used in pain index, ulcer index
        # Recompute from window
        self._drawdowns_peaks = np.array([])
        self._drawdowns_peaks_peak = 0
        for j in range(l):
            dd = 1
            for k in range(self._drawdowns_peaks_peak + 1, j + 1):
                dd *= (1 + w[k] * 0.01)
            if dd > 1:
                self._drawdowns_peaks_peak = j
                self._drawdowns_peaks = np.append(self._drawdowns_peaks, 0)
            else:
                self._drawdowns_peaks = np.append(self._drawdowns_peaks, (dd - 1) * 100)

        # Drawdown calculation used in Burke — recompute from window
        self._drawdown_continuous = np.array([])
        self._drawdown_continuous_final = np.array([])
        self._drawdown_continuous_finalized = False
        self._drawdown_continuous_peak = 1
        self._drawdown_continuous_inside = False
        for j in range(1, l):
            if w[j] < 0:
                if not self._drawdown_continuous_inside:
                    self._drawdown_continuous_inside = True
                    self._drawdown_continuous_peak = j - 1
                self._drawdown_continuous = np.append(self._drawdown_continuous, 0)
            else:
                if self._drawdown_continuous_inside:
                    dd = 1
                    j1 = self._drawdown_continuous_peak + 1
                    for k in range(j1, j):
                        dd = dd * (1 + w[k] * 0.01)
                    self._drawdown_continuous = np.append(self._drawdown_continuous, (dd - 1) * 100)
                    self._drawdown_continuous_inside = False
                else:
                    self._drawdown_continuous = np.append(self._drawdown_continuous, 0)

        # Benchmark tracking (Phase 2)
        bench_ret = return_benchmark / fractional_period
        self.benchmark_returns = np.append(self.benchmark_returns, bench_ret)
        
        # Benchmark window slice
        bw = self.benchmark_returns if self.rolling_window is None \
            else self.benchmark_returns[-self.rolling_window:]
        bl = len(bw)
        
        if bl > 0:
            self._benchmark_mean = np.mean(bw)
            self._benchmark_std = np.std(bw, ddof=1) if bl > 1 else None
            self._benchmark_autocorr_penalty = self._autocorr_penalty(bw)
            
            # Excess benchmark returns
            if self.risk_free_rate == 0:
                self._excess_benchmark_mean = self._benchmark_mean
                self._excess_benchmark_std = self._benchmark_std
            else:
                tmp_b = bw - self.risk_free_rate
                self._excess_benchmark_mean = np.mean(tmp_b)
                self._excess_benchmark_std = np.std(tmp_b, ddof=1) if bl > 1 else None
            
            # Covariance and correlation
            if bl > 1 and self._returns_std is not None and self._benchmark_std is not None \
                    and self._returns_std > 0 and self._benchmark_std > 0:
                self._covariance = np.cov(w, bw)[0, 1]
                self._correlation = self._covariance / (self._returns_std * self._benchmark_std)
                
                # Beta = Cov(Rp, Rb) / Var(Rb)
                bench_var = self._benchmark_std ** 2
                if bench_var > 0:
                    self._beta = self._covariance / bench_var
                    
                    # Alpha = Rp - (Rf + Beta * (Rb - Rf))
                    # Using arithmetic means for alpha calculation
                    rf = self.risk_free_rate
                    self._alpha = self._returns_mean - (rf + self._beta * (self._benchmark_mean - rf))
                    
                    # R-squared
                    self._r_squared = self._correlation ** 2
                    
                    # Tracking error = std(Rp - Rb)
                    active_returns = w - bw
                    self._tracking_error = np.std(active_returns, ddof=1)
                    
                    # Active premium = mean(Rp - Rb)
                    self._active_premium = np.mean(active_returns)
                    
                    # Information ratio = Active premium / Tracking error
                    if self._tracking_error is not None and self._tracking_error > 0:
                        self._information_ratio = self._active_premium / self._tracking_error
                    
                    # Treynor ratio = (Rp - Rf) / Beta
                    excess_portfolio = self._excess_mean if self._excess_mean is not None else self._returns_mean
                    if self._beta is not None and self._beta != 0 and excess_portfolio is not None:
                        self._treynor_ratio = excess_portfolio / self._beta
                    
                    # Appraisal ratio = Alpha / Specific risk (residual std)
                    # Specific risk = std(Rp - (Alpha + Beta * Rb))
                    if self._alpha is not None and self._beta is not None:
                        predicted = self._alpha + self._beta * bw
                        residuals = w - predicted
                        specific_risk = np.std(residuals, ddof=1)
                        if specific_risk > 0:
                            self._appraisal_ratio = self._alpha / specific_risk
                    
                    # Upside/Downside capture
                    up_mask = bw > 0
                    down_mask = bw < 0
                    if np.any(up_mask):
                        self._upside_capture = np.mean(w[up_mask]) / np.mean(bw[up_mask])
                    if np.any(down_mask):
                        self._downside_capture = np.mean(w[down_mask]) / np.mean(bw[down_mask])
                    if self._upside_capture is not None and self._downside_capture is not None \
                            and self._downside_capture != 0:
                        self._capture_ratio = self._upside_capture / self._downside_capture

    def _autocorr_penalty(self, returns) -> float:
        """Metric to account for auto correlation"""
        return 1

    @property
    def _is_primed(self) -> bool:
        """Returns True if enough samples have been added to satisfy min_periods."""
        if self.min_periods is None:
            return True
        return self._sample_count >= self.min_periods

    @property
    def _window_returns(self):
        """Returns the windowed slice of returns."""
        if self.rolling_window is None:
            return self.returns
        return self.returns[-self.rolling_window:]
    
    @property
    def cumulative_return(self):
        """Cumulative geometric returns"""
        return self._cumulative_return_plus_1 - 1
    
    @property
    def drawdowns_cumulative(self):
        """
        Drawdowns from peaks to valleys on cumulative geometric returns.
        """
        return self._drawdowns_cumulative
    
    @property
    def min_drawdowns_cumulative(self):
        """
        The minimum value of the drawdowns from peaks to valleys
        on cumulative geometric returns.
        """
        return self._drawdowns_cumulative_min
    
    @property
    def worst_drawdowns_cumulative(self):
        """
        The absolute value of the minimum value
        of the drawdowns from peaks to valleys
        on cumulative geometric returns.
        """
        return abs(self._drawdowns_cumulative_min)
    
    def drawdowns_peaks(self):
        """
        Drawdowns from peaks to valleys.
        """
        return self._drawdowns_peaks

    def drawdowns_continuous(self, peaks_only: bool = False, max_peaks: int = None):
        """
        Drawdowns on any continuous, uninterrupted losing return region.

        On every end of such uninterrupted negative return region,
        drawdown is the cumulative return ovr the region.

        Otherwise, drawdown is zero.
        
        Used in Burke ratio.

        Args:
            peaks_only bool:
                Return only the peaks, not zeroes.
                Default: False
            max_peaks int:
                Maximum number of peaks to return.
                Has no effect if `peaks_only` is False.
                Default: None
        """
        def finalize_calculation():
            if not self._drawdown_continuous_finalized:
                w = self._window_returns
                if self._drawdown_continuous_inside:
                    dd = 1
                    j1 = self._drawdown_continuous_peak + 1
                    for j in range(j1, len(w)):
                        dd = dd * (1 + w[j] * 0.01)
                    self._drawdown_continuous_final = np.append(self._drawdown_continuous,
                        (dd - 1) * 100)
                else:
                    self._drawdown_continuous_final = np.append(self._drawdown_continuous, 0)
                self._drawdown_continuous_finalized = True
        finalize_calculation()
        if not peaks_only:
            return self._drawdown_continuous_final
        drawdowns = self._drawdown_continuous_final[self._drawdown_continuous_final != 0]
        if max_peaks is not None:
            if len(drawdowns) > 0:
                drawdowns = np.sort(drawdowns)
                drawdowns = drawdowns[:max_peaks]
        return drawdowns

    @property
    def skew(self):
        """
        Calculates returns' skewness
        (the degree of asymmetry of a distribution around its mean)
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        return skew(w) if len(w) > 1 else None

    @property
    def kurtosis(self):
        """
        Calculates returns' kurtosis
         (the degree to which a distribution peak compared to a normal distribution)
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        return kurtosis(w) if len(w) > 1 else None

    # https://www.alternativesoft.com/the-difference-between-the-Sharpe-ratio-and-the-Smart-Sharpe-Ratio.html
    def sharpe_ratio(self,
        ignore_risk_free_rate: bool = False,
        autocorrelation_penalty: bool = False) -> float:
        """
        Ex post Sharpe ratio over excess or raw returns,
        with or without autocorrelation penalty.
        
        Args:
            ignore_risk_free_rate bool:
                Ignore the risk-free rate.
                
                If True, the ratio is calculated over raw returns.
                Sometimes this is called the "risk-return ratio".

                If False, the ratio is calculated over excess returns.
                Default: False
            autocorrelation_penalty bool:
                Apply autocorrelation penalty.
                Default: False
        """
        if not self._is_primed:
            return None
        if ignore_risk_free_rate:
            if (self._returns_mean is None) or \
                (self._returns_std is None) or (self._returns_std == 0):
                return None
            denominator = self._returns_std
            if autocorrelation_penalty:
                denominator *= self._returns_autocorr_penalty
            return self._returns_mean / denominator
        else:
            if (self._excess_mean is None) or \
                (self._excess_std is None) or (self._excess_std == 0):
                return None
            denominator = self._excess_std
            if autocorrelation_penalty:
                denominator *= self._excess_autocorr_penalty
            return self._excess_mean / denominator

    def sortino_ratio(self,
        autocorrelation_penalty: bool = False,
        divide_by_sqrt2: bool = False) -> float:
        """
        Sortino ratio over excess returns,
        with or without autocorrelation penalty.

        Excess returns are returns less the required return.

        Note that the Sortino ratio doesn't use the risk-free rate.
        
        Args:
            autocorrelation_penalty bool:
                Apply autocorrelation penalty.
                Default: False
            divide_by_sqrt2 bool:
                Divide by the square root of 2.

                This Jack Schwager's version of the Sortino ratio
                allows for direct comparisons to the Sharpe.
                
                See here for more info: https://archive.is/wip/2rwFW                
                Default: False
        """
        if not self._is_primed:
            return None
        if (self._required_mean is None) or \
            (self._required_lpm_2 is None) or (self._required_lpm_2 == 0):
            return None
        denominator = np.sqrt(self._required_lpm_2)
        if autocorrelation_penalty:
            denominator *= self._required_autocorr_penalty
        if divide_by_sqrt2:
            denominator *= _SQRT2
        return self._required_mean / denominator

    def omega_ratio(self):
        """
        Omega ratio over normalized returns
        """
        if not self._is_primed:
            return None
        #if (self._required_hpm_1 is None) or \
        if (self._required_mean is None) or \
            (self._required_lpm_1 is None) or (self._required_lpm_1 == 0):
            return None

        #return self._required_hpm_1 / self._required_lpm_1
        return self._required_mean / self._required_lpm_1 + 1

    def kappa_ratio(self, order: int = 3):
        """
        Kappa ratio over normalized returns
        """
        if not self._is_primed:
            return None
        if (self._required_mean is None):
            return None
        if order == 1:
            if (self._required_lpm_1 is None) or (self._required_lpm_1 == 0):
                return None
            return self._required_mean / self._required_lpm_1
        elif order == 2:
            if (self._required_lpm_2 is None) or (self._required_lpm_2 == 0):
                return None
            return self._required_mean / np.sqrt(self._required_lpm_2)
        elif order == 3:
            if (self._required_lpm_3 is None) or (self._required_lpm_3 == 0):
                return None
            return self._required_mean / (self._required_lpm_3 ** (1/3))
        else:
            w = self._window_returns
            if self.required_return == 0:
                tmp = -w
            else:
                tmp = self.required_return - w
            tmp = tmp.clip(min=0)
            lpm = np.sum(tmp ** order) / len(w)
            if (lpm is None) or (lpm == 0):
                return None
            return self._required_mean / (lpm ** (1/order))

    def kappa3_ratio(self, order: int = 3):
        """
        Kappa order 3 ratio over normalized returns
        """
        if not self._is_primed:
            return None
        if (self._required_mean is None) or \
            (self._required_lpm_3 is None) or (self._required_lpm_3 == 0):
            return None
        return self._required_mean / (self._required_lpm_3 ** (1/3))

    def bernardo_ledoit_ratio(self):
        """
        Bernardo and Ledoit ratio over normalized returns
        """
        if not self._is_primed:
            return None
        l = len(self._window_returns)
        if l < 1:
            return None
        tmp = -self._window_returns
        tmp = tmp.clip(min=0)
        lpm_1 = np.sum(tmp) / l
        if lpm_1 is None or lpm_1 == 0:
            return None 
        tmp = self._window_returns.clip(min=0)
        hpm_1 = np.sum(tmp) / l
        return hpm_1 / lpm_1

    def upside_potential_ratio(self, full : bool = True):
        """
        The upside-potential ratio over normalized returns
        """
        if not self._is_primed:
            return None
        if full:
            if (self._required_hpm_1 is None) or \
                (self._required_lpm_2 is None) or (self._required_lpm_2 == 0):
                return None
            return self._required_hpm_1 / np.sqrt(self._required_lpm_2)
        else:
            w = self._window_returns
            tmp = w[w < self.required_return]
            l = len(tmp)
            if l < 1:
                return None
            tmp = tmp - self.required_return
            lpm_2 = np.sum(tmp ** 2) / l
            if lpm_2 is None or lpm_2 == 0:
                return None
            tmp = w[w > self.required_return]
            if len(tmp) == 0:
                return None
            tmp = tmp - self.required_return
            #hpm_1 = np.sum(tmp) / l if l > 0 else None
            hpm_1 = np.mean(tmp)
            return hpm_1 / np.sqrt(lpm_2)

    def compound_growth_rate(self):
        """
        Compound (annual) growth rate (CAGR), or the geometric mean of the returns.
        """
        if not self._is_primed:
            return None
        return self._cumulative_return_geometric_mean
    
    def calmar_ratio(self):
        """
        Calmar ratio over normalized returns
        """
        if not self._is_primed:
            return None
        wdd = self.worst_drawdowns_cumulative
        if wdd == 0:
            return None
        cagr = self._cumulative_return_geometric_mean
        if cagr is None:
            return None
        return cagr / wdd
    
    def sterling_ratio(self, annual_excess_rate: float = 0):
        """
        Steling ratio over normalized returns

        Args:
            annual_excess_rate float:
                Annual excess rate to add to maximum drawdown.
                Default: 0.1 (10%)
        """
        #excess_rate = annual_excess_rate if self.is_annual \
        #    else ((1 + annual_excess_rate) ** (1/252) - 1)
        if not self._is_primed:
            return None
        excess_rate = annual_excess_rate \
            if annual_excess_rate == 0 or self.periods_per_annum == 1 \
            else ((1 + annual_excess_rate) ** (1/self.periods_per_annum) - 1)

        wdd = self.worst_drawdowns_cumulative + excess_rate
        if wdd == 0:
            return None
        cagr = self._cumulative_return_geometric_mean
        if cagr is None:
            return None
        return cagr / wdd

    def burke_ratio(self, modified: bool = False):
        """
        Burke ratio of the return distribution.

        Args:
            modified bool:
                Which ratio to calculate, Burke ratio or modified Burke ratio.
                Default: False
        """
        if not self._is_primed:
            return None
        rate = self._cumulative_return_geometric_mean - self.risk_free_rate
        if rate is None:
            return None
        drawdowns = self.drawdowns_continuous(peaks_only=True)
        if len(drawdowns) < 1:
            return None
        sqrt_sum_drawdowns_squared = np.sqrt(np.sum(np.square(drawdowns)))
        if sqrt_sum_drawdowns_squared == 0:
            return None
        burke = rate / sqrt_sum_drawdowns_squared
        if modified:
            burke *= np.sqrt(len(self._window_returns))
        return burke
  
    def pain_index(self):
        """
        Pain index over normalized returns
        """
        if not self._is_primed:
            return None
        l = len(self._drawdowns_peaks)
        if l < 1:
            return None
        # By calculation, all values are <= 0, so we don't need abs()
        return -np.sum(self._drawdowns_peaks) / l

    def pain_ratio(self):
        """
        Pain ratio over normalized returns
        """
        if not self._is_primed:
            return None
        rate = self._cumulative_return_geometric_mean - self.risk_free_rate
        if rate is None:
            return None
        l = len(self._drawdowns_peaks)
        if l < 1:
            return None
        # By calculation, all values are <= 0, so we don't need abs()
        pain_index = -np.sum(self._drawdowns_peaks) / l
        return (rate / pain_index) if pain_index != 0 else None

    def ulcer_index(self):
        """
        Ulcer index over normalized returns
        """
        if not self._is_primed:
            return None
        l = len(self._drawdowns_peaks)
        if l < 1:
            return None
        ulcer_index = np.sqrt(np.sum(np.square(self._drawdowns_peaks)) / l)
        return ulcer_index
    
    def martin_ratio(self):
        """
        Ulcer ratio over normalized returns
        """
        if not self._is_primed:
            return None
        rate = self._cumulative_return_geometric_mean - self.risk_free_rate
        if rate is None:
            return None
        l = len(self._drawdowns_peaks)
        if l < 1:
            return None
        ulcer_index = np.sqrt(np.sum(np.square(self._drawdowns_peaks)) / l)
        return (rate / ulcer_index) if ulcer_index != 0 else None

    ########################################

    @property
    def gain_to_pain_ratio(self):
        """
        Jack Schwager's GPR. See here for more info:
        https://archive.is/wip/2rwFW
        """
        if not self._is_primed:
            return None
        return (self._returns_mean / self._required_lpm_1) \
            if self._required_lpm_1 != 0 else None

    @property
    def risk_of_ruin(self):
        """
        Calculates the risk of ruin
        (the likelihood of losing all one's investment capital)
        """
        if not self._is_primed:
            return None
        wr = self._win_rate
        return ((1 - wr) / (1 + wr)) ** len(self._window_returns)
    

    @property
    def risk_return_ratio(self): # !!! DELETE
        """
        Calculates the return / risk ratio
        (Sharpe ratio without factoring in the risk-free rate)
        """
        if not self._is_primed:
            return None
        if (self._returns_mean is None) or \
            (self._returns_std is None) or (self._returns_std == 0):
            return None
        return self._returns_mean / self._returns_std

    # =========================================================================
    # NEW MEASURES - Phase 1 High Priority
    # =========================================================================

    def lpm(self, order: int = 2, threshold: float = 0.0) -> float:
        """
        Lower Partial Moment of order `order` below `threshold`.
        
        LPM_n = (1/N) * sum_{r_i < threshold} (threshold - r_i)^n
        
        Args:
            order: Moment order (1, 2, 3, ...)
            threshold: Minimum acceptable return (MAR) in same periodicity as returns
            
        Returns:
            LPM value or None if not enough data
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        l = len(w)
        if l < 1:
            return None
        
        # Check cache first
        cache_key = (order, threshold)
        if cache_key in self._lpm_cache:
            return self._lpm_cache[cache_key]
        
        tmp = threshold - w
        tmp = tmp.clip(min=0)
        result = np.sum(tmp ** order) / l
        self._lpm_cache[cache_key] = result
        return result

    def hpm(self, order: int = 2, threshold: float = 0.0) -> float:
        """
        Higher Partial Moment of order `order` above `threshold`.
        
        HPM_n = (1/N) * sum_{r_i > threshold} (r_i - threshold)^n
        
        Args:
            order: Moment order (1, 2, 3, ...)
            threshold: Minimum acceptable return (MAR) in same periodicity as returns
            
        Returns:
            HPM value or None if not enough data
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        l = len(w)
        if l < 1:
            return None
        
        # Check cache first
        cache_key = (order, threshold)
        if cache_key in self._hpm_cache:
            return self._hpm_cache[cache_key]
        
        tmp = w - threshold
        tmp = tmp.clip(min=0)
        result = np.sum(tmp ** order) / l
        self._hpm_cache[cache_key] = result
        return result

    def downside_deviation(self, mar: float = 0.0, method: str = "full") -> float:
        """
        Downside deviation (semi-deviation) - square root of LPM order 2.
        
        Args:
            mar: Minimum acceptable return in same periodicity as returns
            method: "full" uses N (total observations), "subset" uses count below MAR
            
        Returns:
            Downside deviation or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        l = len(w)
        if l < 1:
            return None
        
        tmp = mar - w
        tmp = tmp.clip(min=0)
        
        if method == "full":
            denom = l
        elif method == "subset":
            denom = np.sum(tmp > 0)
            if denom == 0:
                return None
        else:
            raise ValueError("method must be 'full' or 'subset'")
        
        return np.sqrt(np.sum(tmp ** 2) / denom)

    def semi_deviation(self) -> float:
        """
        Semi-deviation - downside deviation with MAR = mean return.
        """
        if not self._is_primed:
            return None
        return self.downside_deviation(mar=self._returns_mean, method="full")

    def _cornish_fisher_z(self, p: float) -> float:
        """Cornish-Fisher expansion for z-score adjustment."""
        if self._returns_skew is None or self._returns_kurtosis is None:
            return norm.ppf(p)
        
        z = norm.ppf(p)
        S = self._returns_skew
        K = self._returns_kurtosis  # excess kurtosis
        
        # Cornish-Fisher expansion
        z_cf = (z 
                + (z**2 - 1) * S / 6 
                + (z**3 - 3*z) * K / 24 
                - (2*z**3 - 5*z) * S**2 / 36)
        return z_cf

    def var_historical(self, confidence: float = 0.95) -> float:
        """
        Historical Value at Risk (VaR) at given confidence level.
        
        VaR = -quantile(returns, 1 - confidence)
        
        Args:
            confidence: Confidence level (e.g., 0.95 for 95% VaR)
            
        Returns:
            VaR as positive number (loss amount) or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 1:
            return None
        return -np.percentile(w, (1 - confidence) * 100)

    def var_gaussian(self, confidence: float = 0.95) -> float:
        """
        Gaussian (parametric) Value at Risk.
        
        VaR = -(mean + z * std) where z = norm.ppf(1 - confidence)
        
        Args:
            confidence: Confidence level
            
        Returns:
            VaR as positive number or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 2:
            return None
        mean = np.mean(w)
        std = np.std(w, ddof=1)
        if std == 0:
            return None
        z = norm.ppf(1 - confidence)
        return -(mean + z * std)

    def var_cornish_fisher(self, confidence: float = 0.95) -> float:
        """
        Modified Cornish-Fisher VaR accounting for skewness and kurtosis.
        
        Args:
            confidence: Confidence level
            
        Returns:
            VaR as positive number or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 2:
            return None
        mean = np.mean(w)
        std = np.std(w, ddof=1)
        if std == 0:
            return None
        z_cf = self._cornish_fisher_z(1 - confidence)
        return -(mean + z_cf * std)

    def var(self, confidence: float = 0.95, method: str = "modified") -> float:
        """
        Value at Risk with selectable method.
        
        Args:
            confidence: Confidence level (default 0.95)
            method: "historical", "gaussian", or "modified" (Cornish-Fisher)
            
        Returns:
            VaR as positive number or None
        """
        if method == "historical":
            return self.var_historical(confidence)
        elif method == "gaussian":
            return self.var_gaussian(confidence)
        elif method == "modified":
            return self.var_cornish_fisher(confidence)
        else:
            raise ValueError("method must be 'historical', 'gaussian', or 'modified'")

    def cvar_historical(self, confidence: float = 0.95) -> float:
        """
        Historical Conditional VaR (Expected Shortfall).
        
        CVaR = -mean(returns[returns <= -VaR])
        
        Args:
            confidence: Confidence level
            
        Returns:
            CVaR as positive number or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 1:
            return None
        var = self.var_historical(confidence)
        tail = w[w <= -var]
        if len(tail) == 0:
            return None
        return -np.mean(tail)

    def cvar_gaussian(self, confidence: float = 0.95) -> float:
        """
        Gaussian Conditional VaR (Expected Shortfall).
        
        CVaR = -mean + std * phi(z) / (1 - confidence)
        where phi is standard normal PDF, z = norm.ppf(confidence)
        
        Args:
            confidence: Confidence level
            
        Returns:
            CVaR as positive number or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 2:
            return None
        mean = np.mean(w)
        std = np.std(w, ddof=1)
        if std == 0:
            return None
        z = norm.ppf(confidence)
        phi_z = norm.pdf(z)
        return -mean + std * phi_z / (1 - confidence)

    def cvar_cornish_fisher(self, confidence: float = 0.95) -> float:
        """
        Modified Cornish-Fisher CVaR (Expected Shortfall).
        
        Uses the operational version that ensures CVaR >= VaR.
        
        Args:
            confidence: Confidence level
            
        Returns:
            CVaR as positive number or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 2:
            return None
        
        # Use numerical integration approach for modified ES
        # Approximate using adjusted quantiles
        var = self.var_cornish_fisher(confidence)
        tail = w[w <= -var]
        if len(tail) > 0:
            cvar_hist = -np.mean(tail)
            # Operational: ensure CVaR >= VaR
            return max(cvar_hist, var)
        return self.cvar_gaussian(confidence)

    def cvar(self, confidence: float = 0.95, method: str = "modified") -> float:
        """
        Conditional VaR / Expected Shortfall with selectable method.
        
        Args:
            confidence: Confidence level
            method: "historical", "gaussian", or "modified"
            
        Returns:
            CVaR as positive number or None
        """
        if method == "historical":
            return self.cvar_historical(confidence)
        elif method == "gaussian":
            return self.cvar_gaussian(confidence)
        elif method == "modified":
            return self.cvar_cornish_fisher(confidence)
        else:
            raise ValueError("method must be 'historical', 'gaussian', or 'modified'")

    def adjusted_sharpe_ratio(self, rf: float = None) -> float:
        """
        Adjusted Sharpe Ratio (Pezier & White 2006).
        
        AdjSR = SR * [1 + (S/6)*SR - ((K-3)/24)*SR^2]
        where S = skewness, K = kurtosis (not excess)
        
        Args:
            rf: Risk-free rate for period (uses self.risk_free_rate if None)
            
        Returns:
            Adjusted Sharpe Ratio or None
        """
        if not self._is_primed:
            return None
        if self._returns_skew is None or self._returns_kurtosis is None:
            return None
        
        # Use standard Sharpe ratio (annualized logic not needed here as we work with period returns)
        sr = self.sharpe_ratio(ignore_risk_free_rate=(rf == 0 or rf is None))
        if sr is None:
            return None
        
        S = self._returns_skew
        K = self._returns_kurtosis + 3  # Convert excess kurtosis to regular kurtosis
        
        # Adjusted Sharpe formula
        adj_sr = sr * (1 + (S / 6) * sr - ((K - 3) / 24) * sr**2)
        return adj_sr

    def tail_ratio(self, cutoff: float = 0.95) -> float:
        """
        Tail Ratio - ratio of right tail to left tail.
        
        TailRatio = percentile(returns, cutoff) / |percentile(returns, 1-cutoff)|
        
        Args:
            cutoff: Percentile for tail (default 0.95 for 95th/5th percentile)
            
        Returns:
            Tail ratio or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 2:
            return None
        
        right_tail = np.percentile(w, cutoff * 100)
        left_tail = np.percentile(w, (1 - cutoff) * 100)
        
        if left_tail == 0:
            return None
        return right_tail / abs(left_tail)

    def kelly_ratio(self, rf: float = None, method: str = "half") -> float:
        """
        Kelly Criterion ratio (leverage/bet size).
        
        Kelly = mean(excess) / var(excess)
        Half-Kelly = Kelly / 2 (default)
        
        Args:
            rf: Risk-free rate (uses self.risk_free_rate if None)
            method: "full" or "half" Kelly
            
        Returns:
            Kelly ratio or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 2:
            return None
        
        if rf is None:
            rf = self.risk_free_rate
        
        excess = w - rf
        mean_excess = np.mean(excess)
        var_excess = np.var(excess, ddof=1)
        
        if var_excess == 0:
            return None
        
        kelly = mean_excess / var_excess
        if method == "half":
            kelly /= 2
        return kelly

    def probabilistic_sharpe_ratio(self, ref_sr: float = 0.0, 
                                    ignore_skewness: bool = False,
                                    ignore_kurtosis: bool = True) -> float:
        """
        Probabilistic Sharpe Ratio (Marcos Lopez de Prado).
        
        PSR = Phi((SR - ref_SR) * sqrt(n-1) / sqrt(1 - SR*S + SR^2*(K-1)/4))
        
        Args:
            ref_sr: Reference Sharpe ratio (benchmark)
            ignore_skewness: If True, set skewness to 0
            ignore_kurtosis: If True, set excess kurtosis to 0 (K=3)
            
        Returns:
            PSR (probability that true SR > ref_SR) or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        n = len(w)
        if n < 2:
            return None
        
        sr = self.sharpe_ratio()
        if sr is None:
            return None
        
        S = 0 if ignore_skewness else (self._returns_skew or 0)
        K = 3 if ignore_kurtosis else ((self._returns_kurtosis or 0) + 3)
        
        denom = np.sqrt(1 - sr * S + (sr**2) * (K - 1) / 4)
        if denom == 0:
            return None
        
        z = (sr - ref_sr) * np.sqrt(n - 1) / denom
        return norm.cdf(z)

    def k_ratio(self) -> float:
        """
        K-Ratio (Lars Kestner) - slope of log equity curve regression.
        
        K = slope / (stderr_slope * sqrt(N))
        where equity curve is cumulative log returns
        
        Returns:
            K-Ratio or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        n = len(w)
        if n < 3:
            return None
        
        # Build equity curve (cumulative log returns)
        log_returns = np.log1p(w)
        equity = np.cumsum(log_returns)
        
        # Linear regression: equity = a + b * t
        t = np.arange(n, dtype=float)
        
        # Using normal equations for slope and its standard error
        t_mean = np.mean(t)
        equity_mean = np.mean(equity)
        
        S_tt = np.sum((t - t_mean)**2)
        S_te = np.sum((t - t_mean) * (equity - equity_mean))
        
        if S_tt == 0:
            return None
        
        slope = S_te / S_tt
        
        # Residuals
        residuals = equity - (slope * t + equity_mean - slope * t_mean)
        residual_var = np.sum(residuals**2) / (n - 2)
        
        # Standard error of slope
        se_slope = np.sqrt(residual_var / S_tt)
        
        if se_slope == 0:
            return None
        
        k_ratio = slope / (se_slope * np.sqrt(n))
        return k_ratio

    def sortino_satchell_ratio(self) -> float:
        """
        Sortino-Satchell Ratio (Sortino & Satchell 2001).
        
        Uses excess return over MAR in numerator and downside deviation in denominator.
        Unlike standard Sortino, this uses arithmetic mean of excess returns.
        
        Args:
            None
            
        Returns:
            Sortino-Satchell Ratio or None
        """
        if not self._is_primed:
            return None
        if (self._required_mean is None) or \
            (self._required_lpm_2 is None) or (self._required_lpm_2 == 0):
            return None
        denominator = np.sqrt(self._required_lpm_2)
        return self._required_mean / denominator

    def gain_loss_ratio(self) -> float:
        """
        Gain-Loss Ratio (Bernardo & Ledoit variant).
        
        Ratio of sum of positive returns to absolute sum of negative returns.
        GainLoss = sum(r_i > 0) / |sum(r_i < 0)|
        
        Returns:
            Gain-Loss Ratio or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 1:
            return None
        
        gains = w[w > 0]
        losses = w[w < 0]
        
        if len(gains) == 0 or len(losses) == 0:
            return None
        
        sum_gains = np.sum(gains)
        sum_losses = abs(np.sum(losses))
        
        if sum_losses == 0:
            return None
        return sum_gains / sum_losses

    def reward_to_var(self, confidence: float = 0.95, method: str = "modified") -> float:
        """
        Reward-to-VaR Ratio.
        
        Mean excess return divided by VaR.
        
        Args:
            confidence: VaR confidence level
            method: VaR method ("historical", "gaussian", "modified")
            
        Returns:
            Reward-to-VaR or None
        """
        if not self._is_primed:
            return None
        var = self.var(confidence=confidence, method=method)
        if var is None or var == 0:
            return None
        # Use excess return over risk-free rate
        excess_mean = self._excess_mean if self._excess_mean is not None else self._returns_mean
        if excess_mean is None:
            return None
        return excess_mean / var

    def downside_sharpe_ratio(self) -> float:
        """
        Downside Sharpe Ratio (Ziemba 2005).
        
        Mean excess return divided by downside deviation (sqrt(2) * semideviation).
        DSR = mean(excess) / (sqrt(2) * semi_deviation)
        
        Returns:
            Downside Sharpe Ratio or None
        """
        if not self._is_primed:
            return None
        # Use excess return over risk-free rate
        excess_mean = self._excess_mean if self._excess_mean is not None else self._returns_mean
        if excess_mean is None:
            return None
        # Semi-deviation (downside deviation with MAR = mean)
        semi_dev = self.semi_deviation()
        if semi_dev is None or semi_dev == 0:
            return None
        return excess_mean / (_SQRT2 * semi_dev)

    def reward_to_conditional_drawdown(self, confidence: float = 0.95) -> float:
        """
        Reward to Conditional Drawdown Ratio.
        
        CAGR divided by Conditional Drawdown at Risk (CDaR).
        Uses historical CDaR as the average of worst (1-confidence) drawdowns.
        
        Args:
            confidence: Confidence level for conditional drawdown
            
        Returns:
            Reward-to-CDaR or None
        """
        if not self._is_primed:
            return None
        cagr = self._cumulative_return_geometric_mean
        if cagr is None:
            return None
        
        # Get drawdowns from cumulative returns
        dd = self.drawdowns_cumulative
        if len(dd) < 1:
            return None
        
        # Conditional drawdown: average of worst (1-confidence) drawdowns
        n_tail = max(1, int(len(dd) * (1 - confidence)))
        sorted_dd = np.sort(dd)  # Most negative first
        cdar = -np.mean(sorted_dd[:n_tail])  # Positive number
        
        if cdar == 0:
            return None
        return cagr / cdar

    def modigliani_modigliani(self, benchmark_returns: np.ndarray = None) -> float:
        """
        Modigliani-Modigliani M² Measure.
        
        Risk-adjusted return of portfolio adjusted to have same risk (std dev) as benchmark.
        M² = Rf + (Rp - Rf) * (σb / σp)
        
        Args:
            benchmark_returns: Optional benchmark returns array. If None, uses stored benchmark.
            
        Returns:
            M² value or None
        """
        if not self._is_primed:
            return None
        
        # Portfolio excess return
        port_excess = self._excess_mean
        if port_excess is None:
            return None
        
        # Portfolio std dev
        port_std = self._excess_std if self._excess_std is not None else self._returns_std
        if port_std is None or port_std == 0:
            return None
        
        # If benchmark returns not provided, we can't compute M² properly
        # This requires Phase 2 (benchmark tracking)
        if benchmark_returns is None:
            # Return None - need benchmark for M²
            return None
        
        # Benchmark excess return and std
        bench_excess = np.mean(benchmark_returns)
        bench_std = np.std(benchmark_returns, ddof=1)
        if bench_std == 0:
            return None
        
        # M² = Rf + (Rp - Rf) * (σb / σp)
        # Simplified: rf + port_excess * (bench_std / port_std)
        rf = self.risk_free_rate
        m2 = rf + port_excess * (bench_std / port_std)
        return m2

    # =========================================================================
    # PHASE 1 LOW PRIORITY / NICE-TO-HAVE
    # =========================================================================

    def downside_frequency(self, mar: float = 0.0) -> float:
        """
        Downside Frequency - proportion of returns below MAR.
        
        Args:
            mar: Minimum acceptable return
            
        Returns:
            Proportion of returns below MAR (0 to 1) or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 1:
            return None
        below_mar = np.sum(w < mar)
        return below_mar / len(w)

    def downside_potential(self, mar: float = 0.0) -> float:
        """
        Downside Potential - average shortfall below MAR.
        
        DownsidePotential = mean(max(MAR - r_i, 0))
        
        Args:
            mar: Minimum acceptable return
            
        Returns:
            Downside potential (positive number) or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 1:
            return None
        shortfall = np.maximum(mar - w, 0)
        return np.mean(shortfall)

    def volatility_skewness(self, mar: float = 0.0, stat: str = "volatility") -> float:
        """
        Volatility Skewness - ratio of upside to downside volatility.
        
        Similar to Omega but using second partial moments.
        volatility: uses sqrt of second partial moments
        variability: uses second partial moments directly
        
        Uses the "full" method (divides by total n) per Bacon 3rd ed.
        
        Args:
            mar: Minimum acceptable return
            stat: "volatility" or "variability"
            
        Returns:
            Volatility skewness ratio or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        n = len(w)
        if n < 1:
            return None
        
        # Upside returns
        up = w[w > mar]
        # Downside returns
        down = w[w < mar]
        
        if len(up) == 0 or len(down) == 0:
            return None
        
        # Use total n as denominator per Bacon (full method)
        up_moment = np.sum((up - mar) ** 2) / n
        down_moment = np.sum((mar - down) ** 2) / n
        
        if down_moment == 0:
            return None
        
        if stat == "volatility":
            return np.sqrt(up_moment) / np.sqrt(down_moment)
        elif stat == "variability":
            return up_moment / down_moment
        else:
            raise ValueError("stat must be 'volatility' or 'variability'")

    def prospect_ratio(self, mar: float = 0.0, alpha: float = 0.88, 
                       lambda_loss: float = 2.25) -> float:
        """
        Prospect Ratio based on Prospect Theory (Kahneman & Tversky).
        
        Weights gains by alpha and losses by lambda_loss.
        
        Args:
            mar: Minimum acceptable return (reference point)
            alpha: Risk aversion parameter for gains (default 0.88)
            lambda_loss: Loss aversion parameter (default 2.25)
            
        Returns:
            Prospect ratio or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 1:
            return None
        
        # Prospect theory value function
        gains = w[w > mar] - mar
        losses = mar - w[w < mar]
        
        if len(gains) == 0 and len(losses) == 0:
            return None
        
        # Value function: v(x) = x^alpha for gains, -lambda * |x|^alpha for losses
        gain_value = np.mean(gains ** alpha) if len(gains) > 0 else 0
        loss_value = -lambda_loss * np.mean(losses ** alpha) if len(losses) > 0 else 0
        
        total_value = gain_value + loss_value
        if total_value == 0:
            return None
        
        # Ratio of prospect value to downside risk
        return total_value / self.downside_deviation(mar=mar)

    def farinelli_tibiletti_ratio(self, mar: float = 0.0, 
                                   u: float = 2.0, l: float = 2.0) -> float:
        """
        Farinelli-Tibiletti Ratio (Farinelli & Tibiletti 2008).
        
        Generalized measure combining upside and downside partial moments.
        
        F-T(l, u) = (HPM_u)^(1/u) / (LPM_l)^(1/l)
        
        Where:
        - HPM_u = (1/n) * sum_{r_i > mar} (r_i - mar)^u
        - LPM_l = (1/n) * sum_{r_i < mar} (mar - r_i)^l
        
        Special cases (from Bacon 3rd ed, Chapter 5):
        - u=1, l=1: Omega ratio
        - u=2, l=2: Variability skewness (upside risk / downside risk)
        - u=1, l=2: Upside potential ratio
        - u < 1, l > 1: Risk averse
        - u > 1, l < 1: Risk seeking
        
        Args:
            mar: Minimum acceptable return (target)
            u: Upside moment order (default 2.0)
            l: Downside moment order (default 2.0)
            
        Returns:
            Farinelli-Tibiletti ratio or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        n = len(w)
        if n < 1:
            return None
        
        # Upside partial moment (HPM) - divide by total n per Bacon
        up = w[w > mar]
        if len(up) == 0:
            return None
        hpm_u = np.sum((up - mar) ** u) / n
        
        # Downside partial moment (LPM) - divide by total n per Bacon
        down = w[w < mar]
        if len(down) == 0:
            return None
        lpm_l = np.sum((mar - down) ** l) / n
        
        if lpm_l == 0:
            return None
        
        return (hpm_u ** (1/u)) / (lpm_l ** (1/l))

    def bias_ratio(self, std_dev_multiple: float = 1.0) -> float:
        """
        Bias Ratio (Adil Abdulali 2006).
        
        Measures the distribution of returns near zero to detect stale pricing
        and potential smoothing/fraud in illiquid asset strategies.
        
        BR = Count(returns in [0, std_dev]) / (1 + Count(returns in [-std_dev, 0]))
        
        A bias ratio < 1 is undesirable; equity indexes typically 1-1.5;
        very high ratios may indicate smoothing or fraud.
        
        Args:
            std_dev_multiple: Number of standard deviations for "close to zero" 
                            threshold (default 1.0 = 1 sigma)
            
        Returns:
            Bias ratio or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        n = len(w)
        if n < 2:
            return None
        
        std = self._returns_std
        if std is None or std == 0:
            return None
        
        threshold = std_dev_multiple * std
        
        # Count returns in [0, threshold]
        count_positive = np.sum((w >= 0) & (w <= threshold))
        
        # Count returns in [-threshold, 0)
        count_negative = np.sum((w >= -threshold) & (w < 0))
        
        # Bias ratio = count_positive / (1 + count_negative)
        return count_positive / (1 + count_negative)

    def skewness_kurtosis_ratio(self) -> float:
        """
        Skewness-Kurtosis Ratio (S/K).
        
        Ratio of skewness to kurtosis (moment method).
        Higher is better.
        
        Returns:
            S/K ratio or None
        """
        if not self._is_primed:
            return None
        if self._returns_skew is None or self._returns_kurtosis is None:
            return None
        # Convert excess kurtosis to regular kurtosis
        K = self._returns_kurtosis + 3
        if K == 0:
            return None
        return self._returns_skew / K

    def bera_jarque_statistic(self) -> float:
        """
        Bera-Jarque (Jarque-Bera) Normality Test Statistic.
        
        Tests whether return distribution is normal by combining
        skewness and excess kurtosis.
        
        BJ = n/6 * (skewness^2 + excess_kurtosis^2/4)
        
        Critical values (chi-squared with 2 df):
        - 95% confidence: 5.99
        - 99% confidence: 9.21
        
        If BJ > critical value, reject normality hypothesis.
        
        Returns:
            Bera-Jarque statistic or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        n = len(w)
        if n < 3:
            return None
        
        # Use population skewness and excess kurtosis (not sample corrected)
        # These match the scipy.stats skew and kurtosis with bias=True, fisher=True
        skew = self._returns_skew
        excess_kurtosis = self._returns_kurtosis
        
        if skew is None or excess_kurtosis is None:
            return None
        
        # Bera-Jarque formula (Equation 5.17 from Bacon 3rd ed)
        bj = (n / 6) * (skew**2 + (excess_kurtosis**2) / 4)
        return float(bj)

    def is_normal_distribution(self, 
                   confidence: float = 0.95) -> bool:
        """
        Test if returns follow normal distribution using Bera-Jarque test.
        
        Args:
            confidence: Confidence level (0.95 or 0.99)
            
        Returns:
            True if cannot reject normality, False if reject, None if insufficient data
        """
        bj = self.bera_jarque_statistic()
        if bj is None:
            return None
        
        if confidence == 0.95:
            return bj <= 5.99
        elif confidence == 0.99:
            return bj <= 9.21
        else:
            # Use chi2 ppf for other confidence levels
            from scipy.stats import chi2
            critical = chi2.ppf(confidence, 2)
            return bj <= critical

    def mad_ratio(self) -> float:
        """
        Mean Absolute Deviation Ratio.
        
        MAD Ratio = Mean Return / Mean Absolute Deviation
        
        Returns:
            MAD ratio or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 1:
            return None
        mad = np.mean(np.abs(w - np.mean(w)))
        if mad == 0:
            return None
        return np.mean(w) / mad

    def omega_sharpe_ratio(self) -> float:
        """
        Omega-Sharpe Ratio.
        
        Conversion of Omega ratio to Sharpe-like form.
        Omega-Sharpe = (Omega - 1) * sqrt(LPM2)
        
        Returns:
            Omega-Sharpe ratio or None
        """
        if not self._is_primed:
            return None
        omega = self.omega_ratio()
        if omega is None:
            return None
        lpm2 = self._required_lpm_2
        if lpm2 is None or lpm2 == 0:
            return None
        return (omega - 1) * np.sqrt(lpm2)

    def omega_excess_return(self, mar: float = 0.0) -> float:
        """
        Omega Excess Return.
        
        Portfolio return minus 3 times portfolio downside deviation 
        times benchmark downside deviation.
        
        Args:
            mar: Minimum acceptable return
            
        Returns:
            Omega excess return or None
        """
        if not self._is_primed:
            return None
        if self._benchmark_std is None or self._benchmark_std == 0:
            return None
        
        # Annualize returns and downside deviations
        period = self.periods_per_annum
        if period <= 1:
            return None
        
        # Portfolio annualized return
        rp = self._cumulative_return_geometric_mean
        if rp is None:
            return None
        rp_annual = (1 + rp) ** period - 1
        
        # Portfolio annualized downside deviation
        sigma_d = self.downside_deviation(mar=mar)
        if sigma_d is None:
            return None
        sigma_d_annual = sigma_d * np.sqrt(period)
        
        # Benchmark annualized downside deviation
        sigma_dm = self._benchmark_std * np.sqrt(period)
        
        # Omega excess return
        return rp_annual - 3 * sigma_d_annual * sigma_dm

    def rachev_ratio(self, alpha: float = 0.1, beta: float = 0.1, rf: float = 0.0) -> float:
        """
        Rachev Ratio.
        
        Non-parametric estimator of upper tail reward potential relative to 
        lower tail risk.
        
        Rachev = ES_upper / ES_lower
        where ES_upper = Expected Shortfall of upper tail (1-beta)
        and ES_lower = Expected Shortfall of lower tail (alpha)
        
        Args:
            alpha: Lower tail probability (default 0.1)
            beta: Upper tail probability (default 0.1)
            rf: Risk-free rate (default 0)
            
        Returns:
            Rachev Ratio or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 2:
            return None
        
        # Excess returns
        excess = w - rf
        
        # Lower tail (alpha quantile)
        var_lower = -np.percentile(excess, alpha * 100)
        tail_lower = excess[excess <= -var_lower]
        if len(tail_lower) == 0:
            return None
        es_lower = -np.mean(tail_lower)
        
        # Upper tail (beta quantile)
        n_upper = max(1, int((1 - beta) * len(excess)))
        sorted_excess = np.sort(excess)
        var_upper = sorted_excess[n_upper - 1]
        tail_upper = excess[excess >= var_upper]
        if len(tail_upper) == 0:
            return None
        es_upper = np.mean(tail_upper)
        
        if es_lower == 0:
            return None
        return es_upper / es_lower

    def timing_ratio(self) -> float:
        """
        Timing Ratio.
        
        Ratio of bull market beta to bear market beta.
        TimingRatio = Beta_bull / Beta_bear
        
        Returns:
            Timing ratio or None
        """
        if not self._is_primed:
            return None
        if self._beta is None or self._beta == 0:
            return None
        
        w = self._window_returns
        bw = self.benchmark_returns if self.rolling_window is None \
            else self.benchmark_returns[-self.rolling_window:]
        bl = len(bw)
        
        if bl < 2:
            return None
        
        # Bull market: benchmark > 0
        bull_mask = bw > 0
        # Bear market: benchmark < 0
        bear_mask = bw < 0
        
        if not np.any(bull_mask) or not np.any(bear_mask):
            return None
        
        # Bull beta
        bull_cov = np.cov(w[bull_mask], bw[bull_mask])[0, 1]
        bull_var = np.var(bw[bull_mask])
        if bull_var == 0:
            return None
        beta_bull = bull_cov / bull_var
        
        # Bear beta
        bear_cov = np.cov(w[bear_mask], bw[bear_mask])[0, 1]
        bear_var = np.var(bw[bear_mask])
        if bear_var == 0:
            return None
        beta_bear = bear_cov / bear_var
        
        if beta_bear == 0:
            return None
        
        return beta_bull / beta_bear

    def d_ratio(self) -> float:
        """
        D-Ratio.
        
        Similar to Bernardo-Ledoit but inverted and accounting for frequency
        of positive and negative returns.
        
        DRatio = (nd * sum(max(-r_i, 0))) / (nu * sum(max(r_i, 0)))
        where nd = count of negative returns, nu = count of positive returns
        
        Lower is better. 0 = no negative returns, inf = no positive returns.
        
        Returns:
            D-Ratio or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < 1:
            return None
        
        pos = w[w > 0]
        neg = w[w < 0]
        
        nu = len(pos)
        nd = len(neg)
        
        if nu == 0:
            return np.inf  # No positive returns
        if nd == 0:
            return 0.0   # No negative returns
        
        sum_pos = np.sum(pos)
        sum_neg = np.sum(neg)
        
        if sum_pos == 0:
            return np.inf
        
        return (-nd * sum_neg) / (nu * sum_pos)

    def smoothing_index(self, neg_thetas: bool = False, ma_order: int = 2) -> float:
        """
        Smoothing Index (Getmansky).
        
        Normalized Herfindahl index of MA coefficients from ARIMA(0,0,q) fit.
        Lower values indicate more smoothing (less liquid).
        
        Args:
            neg_thetas: If False, remove negative coefficients
            ma_order: Order of MA process (default 2)
            
        Returns:
            Smoothing index (0 to 1, lower = more smoothing) or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        if len(w) < ma_order + 2:
            return None
        
        # Demean returns
        w_demean = w - np.mean(w)
        
        # Fit ARMA(0, q) using Yule-Walker equations (simplified MA estimation)
        # This is a simplified version - full MLE would require scipy.signal or statsmodels
        # We'll use autocorrelation method for MA coefficients
        
        n = len(w_demean)
        if n < ma_order + 2:
            return None
        
        # Compute autocorrelations
        acf = np.zeros(ma_order + 1)
        for k in range(ma_order + 1):
            if k == 0:
                acf[k] = 1.0
            else:
                acf[k] = np.sum(w_demean[k:] * w_demean[:-k]) / np.sum(w_demean ** 2)
        
        # For MA(q), solve Yule-Walker equations
        # rho_k = sum_{i=1}^q theta_i * rho_{k-i} for k=1..q
        # With theta_0 = 1
        
        if ma_order == 1:
            # MA(1): rho_1 = theta_1 / (1 + theta_1^2)
            # Solve quadratic: theta_1^2 * rho_1 - theta_1 + rho_1 = 0
            rho = acf[1]
            if abs(rho) > 0.5:  # MA(1) invertibility condition
                return None
            # Use the smaller root for invertibility
            disc = 1 - 4 * rho**2
            if disc < 0:
                return None
            theta1 = (1 - np.sqrt(disc)) / (2 * rho)
            thetas = np.array([1.0, theta1])
        elif ma_order == 2:
            # MA(2): rho_1 = theta_1 + theta_1*theta_2 / (1 + theta_1^2 + theta_2^2)
            # rho_2 = theta_2 / (1 + theta_1^2 + theta_2^2)
            # This is complex to solve analytically, use simplified approximation
            rho1, rho2 = acf[1], acf[2]
            # Approximate: assume theta_1 and theta_2 small
            # rho_1 ≈ theta_1, rho_2 ≈ theta_2
            thetas = np.array([1.0, rho1, rho2])
        else:
            # General case: use autocorrelation approximation
            thetas = np.array([1.0] + list(acf[1:ma_order+1]))
        
        # Remove negative thetas if requested
        if not neg_thetas:
            thetas = np.maximum(thetas, 0)
        
        # Normalize to sum to 1
        thetas = thetas / np.sum(thetas)
        
        # Herfindahl index
        smoothing_idx = np.sum(thetas ** 2)
        
        return float(smoothing_idx)

    def cdar_beta(self, confidence: float = 0.95, geometric: bool = True) -> float:
        """
        Conditional Drawdown Beta (CDaR Beta).
        
        Measures the sensitivity of portfolio returns to benchmark drawdowns.
        Focuses only on the worst (1-p) fraction of benchmark drawdown periods.
        
        Args:
            confidence: Confidence level for tail (default 0.95 = worst 5%)
            geometric: Use geometric chaining for returns (default True)
            
        Returns:
            CDaR Beta or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        bw = self.benchmark_returns if self.rolling_window is None \
            else self.benchmark_returns[-self.rolling_window:]
        n = len(w)
        if n < 2 or len(bw) != n:
            return None
        
        # Compute benchmark drawdowns
        dd_bench = self._compute_benchmark_drawdowns(bw, geometric)
        if len(dd_bench) < 1:
            return None
        
        # Get threshold quantile
        p = 1 - confidence  # proportion of worst drawdowns
        if p <= 0:
            q_quantile = np.min(dd_bench)
        else:
            q_quantile = np.quantile(dd_bench, p)
        
        # Select periods with drawdowns at or below quantile
        dd_indices = np.where(dd_bench <= q_quantile)[0]
        if len(dd_indices) < 1:
            return None
        
        # Sum portfolio returns over selected drawdown periods
        sum_dd_port = 0.0
        for idx in dd_indices:
            # Get the drawdown period for this index
            dd_period = self._get_drawdown_period(bw, idx, geometric)
            if dd_period is not None:
                sum_dd_port += dd_period
        
        # Average benchmark drawdown in tail
        cdd_value = np.mean(dd_bench[dd_indices])
        if cdd_value == 0:
            return None
        
        beta_dd = sum_dd_port / (len(dd_indices) * cdd_value)
        return float(beta_dd)

    def cdar_alpha(self, confidence: float = 0.95, geometric: bool = True) -> float:
        """
        Conditional Drawdown Alpha (CDaR Alpha).
        
        Alpha based on CDaR Beta - annualized portfolio return minus
        CDaR Beta times annualized benchmark return.
        
        Args:
            confidence: Confidence level for tail (default 0.95)
            geometric: Use geometric chaining for returns (default True)
            
        Returns:
            CDaR Alpha or None
        """
        if not self._is_primed:
            return None
        
        beta = self.cdar_beta(confidence=confidence, geometric=geometric)
        if beta is None:
            return None
        
        w = self._window_returns
        bw = self.benchmark_returns if self.rolling_window is None \
            else self.benchmark_returns[-self.rolling_window:]
        
        # Annualize returns (assuming monthly frequency -> 12 periods)
        period = 12 if self.periodicity == Periodicity.MONTHLY else self.periods_per_annum
        
        if geometric:
            rp_annual = (1 + np.mean(w)) ** period - 1
            rm_annual = (1 + np.mean(bw)) ** period - 1
        else:
            rp_annual = np.mean(w) * period
            rm_annual = np.mean(bw) * period
        
        alpha = rp_annual - beta * rm_annual
        return float(alpha)

    def fama_beta(self) -> float:
        """
        Fama Beta.
        
        Ratio of portfolio standard deviation to benchmark standard deviation.
        Used to calculate loss of diversification.
        
        FamaBeta = sigma_P / sigma_M
        
        Returns:
            Fama Beta or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        bw = self.benchmark_returns if self.rolling_window is None \
            else self.benchmark_returns[-self.rolling_window:]
        n = len(w)
        if n < 2 or len(bw) != n:
            return None
        
        sigma_p = np.std(w, ddof=1)
        sigma_m = np.std(bw, ddof=1)
        
        if sigma_m == 0:
            return None
        
        return sigma_p / sigma_m

    def sfm_coefficients(self, rf: float = None, method: str = "LS") -> dict:
        """
        Single Factor Model Coefficients (Alpha and Beta).
        
        OLS regression of excess portfolio returns on excess benchmark returns.
        
        Args:
            rf: Risk-free rate (uses self.risk_free_rate if None)
            method: "LS" for Least Squares (only implemented)
            
        Returns:
            Dict with 'alpha', 'beta', 'r_squared' or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        bw = self.benchmark_returns if self.rolling_window is None \
            else self.benchmark_returns[-self.rolling_window:]
        n = len(w)
        if n < 2 or len(bw) != n:
            return None
        
        if rf is None:
            rf = self.risk_free_rate
        
        # Excess returns
        xra = w - rf
        xrb = bw - rf
        
        # OLS regression
        # beta = Cov(xra, xrb) / Var(xrb)
        cov = np.cov(xra, xrb)[0, 1]
        var_b = np.var(xrb, ddof=1)
        if var_b == 0:
            return None
        
        beta = cov / var_b
        alpha = np.mean(xra) - beta * np.mean(xrb)
        r2 = (cov ** 2) / (np.var(xra, ddof=1) * var_b) if np.var(xra, ddof=1) > 0 else 0
        
        return {
            'alpha': float(alpha),
            'beta': float(beta),
            'r_squared': float(r2)
        }

    def hurst_index(self) -> float:
        """
        Hurst Index (Rescaled Range Analysis).
        
        Measures long-term memory of returns:
        - H > 0.5: Persistent (trending)
        - H = 0.5: Random walk
        - H < 0.5: Mean-reverting (anti-persistent)
        
        H = log(R/S) / log(n)
        where R = range of cumulative demeaned returns
              S = standard deviation of returns
              n = number of observations
        
        Returns:
            Hurst exponent (0 to 1) or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        n = len(w)
        if n < 2:
            return None
        
        # Demean returns
        w_demean = w - np.mean(w)
        
        # Cumulative sum of demeaned returns
        z = np.cumsum(w_demean)
        
        # Range
        r_range = np.max(z) - np.min(z)
        
        # Standard deviation
        s = np.std(w, ddof=1)
        if s == 0:
            return None
        
        # Rescaled range
        m = r_range / s
        if m <= 0:
            return None
        
        # Hurst exponent
        h = np.log(m) / np.log(n)
        return float(h)

    def _compute_benchmark_drawdowns(self, returns: np.ndarray, geometric: bool = True) -> np.ndarray:
        """Helper to compute drawdowns for benchmark."""
        if len(returns) < 1:
            return np.array([])
        
        cumr = 1.0
        cumr_max = 1.0
        drawdowns = []
        
        for ret in returns:
            if geometric:
                cumr *= (1 + ret)
            else:
                cumr += ret
            if cumr > cumr_max:
                cumr_max = cumr
            dd = cumr / cumr_max - 1
            drawdowns.append(dd)
        
        return np.array(drawdowns)

    def _get_drawdown_period(self, returns: np.ndarray, trough_idx: int, geometric: bool) -> float:
        """Get cumulative return from peak to trough for a drawdown period."""
        if trough_idx < 0 or trough_idx >= len(returns):
            return None
        
        # Find peak before trough
        cumr = 1.0
        cumr_max = 1.0
        peak_idx = 0
        
        for i, ret in enumerate(returns[:trough_idx+1]):
            if geometric:
                cumr *= (1 + ret)
            else:
                cumr += ret
            if cumr > cumr_max:
                cumr_max = cumr
                peak_idx = i
        
        # Compute return from peak to trough
        if geometric:
            cumr = 1.0
            for i in range(peak_idx, trough_idx + 1):
                cumr *= (1 + returns[i])
            return cumr - 1
        else:
            return sum(returns[peak_idx:trough_idx + 1])

    # =========================================================================
    # ADDITIONAL MEASURES FROM BACON 3RD EDITION
    # =========================================================================

    def percentile_rank(self, method: int = 3) -> float:
        """
        Percentile rank of the latest return in the window.
        
        Computes the rank of the most recent return among all returns in the window.
        Uses Bacon's preferred method 3 by default.
        
        Args:
            method: 1 to 5 (see Bacon 3rd ed, Chapter 4, Table 4.6)
                1: n/N
                2: (n-1)/N
                3: (n-1)/(N-1)  [DEFAULT - median is exactly 50%]
                4: (n-0.5)/N
                5: n/(N+1)
        
        Returns:
            Percentile rank (0 to 1) or None
        """
        if not self._is_primed:
            return None
        w = self._window_returns
        n = len(w)
        if n < 1:
            return None
        
        # Rank of the most recent return (1 = best, n = worst)
        rank = 1 + np.sum(w[:-1] > w[-1])  # Exclude the last element itself
        if rank == n:  # Worst
            rank = n
        
        if method == 1:
            return rank / n
        elif method == 2:
            return (rank - 1) / n
        elif method == 3:
            return (rank - 1) / (n - 1) if n > 1 else 0.5
        elif method == 4:
            return (rank - 0.5) / n
        elif method == 5:
            return rank / (n + 1)
        else:
            raise ValueError("method must be 1-5")

    def modified_information_ratio(self) -> float:
        """
        Modified Information Ratio (Israelson 2005).
        
        Addresses the issue that standard IR rewards higher tracking error
        when excess return is negative.
        
        Formula: MIR = IR if excess > 0 else -IR
        
        Returns:
            Modified IR or None
        """
        if not self._is_primed:
            return None
        
        ir = self.information_ratio
        if ir is None:
            return None
        
        excess = self.active_premium
        if excess is None:
            return None
        
        return ir if excess > 0 else -ir

    def information_ratio_geometric(self) -> float:
        """
        Geometric Information Ratio (Bacon 3rd ed, Chapter 5).
        
        Uses geometric mean of active returns instead of arithmetic.
        IR_G = geometric_active_return / tracking_error
        
        Returns:
            Geometric IR or None
        """
        if not self._is_primed:
            return None
        
        te = self.tracking_error
        if te is None or te == 0:
            return None
        
        w = self._window_returns
        bw = self.benchmark_returns if self.rolling_window is None \
            else self.benchmark_returns[-self.rolling_window:]
        
        if len(w) != len(bw) or len(w) < 1:
            return None
        
        active = w - bw
        geo_excess = np.prod(1 + active) ** (1 / len(active)) - 1
        
        return geo_excess / te


# =========================================================================
# PHASE 2: Benchmark-Dependent Measures
# =========================================================================

    def __init_benchmark_attrs(self):
        """Initialize benchmark tracking attributes."""
        self.benchmark_returns: np.ndarray = np.array([])
        self._benchmark_mean: float = None
        self._benchmark_std: float = None
        self._benchmark_autocorr_penalty: float = None
        self._excess_benchmark_mean: float = None
        self._excess_benchmark_std: float = None
        self._covariance: float = None
        self._correlation: float = None
        self._beta: float = None
        self._alpha: float = None
        self._r_squared: float = None
        self._tracking_error: float = None
        self._active_premium: float = None
        self._information_ratio: float = None
        self._treynor_ratio: float = None
        self._appraisal_ratio: float = None
        self._upside_capture: float = None
        self._downside_capture: float = None
        self._capture_ratio: float = None

    def reset(self):
        self.fractional_periods = np.array([])
        self.returns = np.array([])
        self.benchmark_returns = np.array([])
        self._logret_sum = 0
        self._drawdowns_cumulative = np.array([])
        self._drawdowns_cumulative_min = np.inf
        self._drawdowns_peaks = np.array([])
        self._drawdowns_peaks_peak = 0
        self._drawdown_continuous = np.array([])
        self._drawdown_continuous_final = np.array([])
        self._drawdown_continuous_finalized = False
        self._drawdown_continuous_peak = 1
        self._drawdown_continuous_inside = False
        self._cumulative_return_plus_1 = 1
        self._cumulative_return_plus_1_100 = 1
        self._cumulative_return_plus_1_max = -np.inf
        self._total_duration = 0
        self._sample_count = 0

        # Reset LPM/HPM caches
        self._lpm_cache = {}
        self._hpm_cache = {}

        # Reset benchmark attributes
        self.__init_benchmark_attrs()

    @property
    def beta(self) -> float:
        """CAPM Beta."""
        if not self._is_primed:
            return None
        return self._beta

    @property
    def alpha(self) -> float:
        """Jensen's Alpha."""
        if not self._is_primed:
            return None
        return self._alpha

    @property
    def r_squared(self) -> float:
        """R-squared (coefficient of determination)."""
        if not self._is_primed:
            return None
        return self._r_squared

    @property
    def tracking_error(self) -> float:
        """Tracking Error (std of active returns)."""
        if not self._is_primed:
            return None
        return self._tracking_error

    @property
    def active_premium(self) -> float:
        """Active Premium (mean of active returns)."""
        if not self._is_primed:
            return None
        return self._active_premium

    @property
    def information_ratio(self) -> float:
        """Information Ratio (Active Premium / Tracking Error)."""
        if not self._is_primed:
            return None
        return self._information_ratio

    @property
    def treynor_ratio(self) -> float:
        """Treynor Ratio (Excess return / Beta)."""
        if not self._is_primed:
            return None
        return self._treynor_ratio

    @property
    def appraisal_ratio(self) -> float:
        """Appraisal Ratio (Alpha / Specific Risk)."""
        if not self._is_primed:
            return None
        return self._appraisal_ratio

    @property
    def upside_capture(self) -> float:
        """Upside Capture Ratio."""
        if not self._is_primed:
            return None
        return self._upside_capture

    @property
    def downside_capture(self) -> float:
        """Downside Capture Ratio."""
        if not self._is_primed:
            return None
        return self._downside_capture

    @property
    def capture_ratio(self) -> float:
        """Capture Ratio (Upside Capture / Downside Capture)."""
        if not self._is_primed:
            return None
        return self._capture_ratio

    def modigliani_modigliani(self, benchmark_returns: np.ndarray = None) -> float:
        """
        Modigliani-Modigliani M² Measure.
        
        Risk-adjusted return of portfolio adjusted to have same risk (std dev) as benchmark.
        M² = Rf + (Rp - Rf) * (σb / σp)
        
        Args:
            benchmark_returns: Optional benchmark returns array. If None, uses stored benchmark.
            
        Returns:
            M² value or None
        """
        if not self._is_primed:
            return None
        
        # If no external benchmark provided, use internal tracked benchmark
        if benchmark_returns is None:
            if self._benchmark_std is None or self._benchmark_std == 0:
                return None
            bench_std = self._benchmark_std
            port_excess = self._excess_mean
        else:
            # Use provided benchmark
            bench_std = np.std(benchmark_returns, ddof=1)
            if bench_std == 0:
                return None
            port_excess = self._excess_mean
        
        if port_excess is None:
            return None
        
        # Portfolio std dev
        port_std = self._excess_std if self._excess_std is not None else self._returns_std
        if port_std is None or port_std == 0:
            return None
        
        # M² = Rf + (Rp - Rf) * (σb / σp)
        rf = self.risk_free_rate
        m2 = rf + port_excess * (bench_std / port_std)
        return m2


# https://github.com/webclinic017/cryptoTrendFollowing/blob/main/main.py
#def annualising_factor(series):
#    return round(365 / ((np.diff(df.index.values).mean().astype(float) * 1e-9)/(24*3600)))

# https://github.com/sprksh/finance-calculator/blob/master/src/finance_calculator/calculators/ratio_calculator.py

# https://github.com/laholmes/risk-adjusted-return/blob/e0865a586d466a52ab1e549d46778e2aac0f0b0f/app.py#L336
# this one is good without pandas
# https://www.turingfinance.com/computational-investing-with-python-week-one/

# https://github.com/midas-research/profit-naacl

# https://github.com/AI-ApeX-DeV/TradeBot-USDT_BTC/blob/main/Metrics.txt


# https://www.pm-research.com/content/iijinvest/3/3/59
# Sortino, F.A.; Price, L.N. (1994). "Performance measurement in a downside risk framework". Journal of Investing. 3 (3): 50–8. doi:10.3905/joi.3.3.59
#
# "Sortino: A 'Sharper' Ratio" (PDF). Red Rock Capital. Retrieved February 16, 2014.
# http://www.redrockcapital.com/Sortino__A__Sharper__Ratio_Red_Rock_Capital.pdf
#
# https://en.wikipedia.org/wiki/Rate_of_return
# According to the CFA Institute's Global Investment Performance Standards (GIPS),[3]
#"Returns for periods of less than one year must not be annualized."
# https://www.cfainstitute.org/en/membership/professional-development/refresher-readings/gips-overview
# Overview of the Global Investment Performance Standards "GIPS Standards"
#
# https://github.com/Peter-Staadecker/Lions-and-Tigers-and-Sortinos-Oh-My/blob/main/tiingo%20analysis%20multi-yr%20monthly%20v4%20%20-%20API%20key%20blank.py
# min_acceptable_return_in_period
# ------------------ calculate Sortino ratio =  (stock return - minimum acceptable return)/ downside std dev
# note 1. returns are % growth, not growth ratio
# note 2. downside std dev counts all periods both above and below target in the std dev denominator,
#   i.e. the zero values in the downside are not thrown away. For emphasis on this point see
#   e.g. http://www.redrockcapital.com/Sortino__A__Sharper__Ratio_Red_Rock_Capital.pdf
# note 3. I use a geometric average for average growth in calculating the numerator of the Sortino ratio.
#   Some may prefer an arithmetic ratio. Based on some quick tests the difference is likely minor.
# note 4. I use std. deviation for the population, not for a sample. Again the differences are likely minor.
# note 5. There are however, major differences in ratios depending on whether the data frequency is daily, weekly
#   monthly or yearly. As a result, it seems that the ratios (up/down market capture and Sortino) are more
#   useful for comparison between stocks when measured with the same data frequency, rather than as an absolute measure.
#   The beta ratio seems least influenced by frequency.
#
# downsideStdDevAnnlzd: float = downsideStdDevPeriod * (PeriodsInYr ** 0.5)  # <----////////adjust if needed
#
#if dataFreqStr == "weekly":
#    PeriodsInYr = 52
#elif dataFreqStr == "monthly":
#    PeriodsInYr = 12
#elif dataFreqStr == "annually":
#    PeriodsInYr = 1
#elif dataFreqStr == "daily":
#    PeriodsInYr = 253  # using an average
#
# https://github.com/kunnn1/Quant-Calc/blob/main/quant_calc.py
# def calculate_sharpe_ratio(returns, risk_free_rate, trading_days=252):
#   excess_returns = returns - risk_free_rate / trading_days
#    sharpe_ratio = np.sqrt(trading_days) * excess_returns.mean() / excess_returns.std()
#    return sharpe_ratio
# def calculate_sortino_ratio(returns, risk_free_rate, trading_days=252):
#    downside_returns = returns[returns < 0]
#    excess_returns = returns - risk_free_rate / trading_days
#    sortino_ratio = np.sqrt(trading_days) * excess_returns.mean() / downside_returns.std()
#    return sortino_ratio
# def calculate_max_drawdown(returns):
#    cumulative_returns = (1 + returns).cumprod()
#    peak = cumulative_returns.cummax()
#    drawdown = (cumulative_returns - peak) / peak
#    max_drawdown = drawdown.min()
#    return max_drawdown
#
#
