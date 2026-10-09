import unittest
from unittest.mock import PropertyMock, patch
from datetime import datetime
import inspect
import math
import random
import statistics

from ..streaming_kbn import RawMomentsKleinKBN
from ..performance import Measures

from . import reference_data as rd

# From the repository root, run the public API tests with:
#   python3 -m unittest py.performance.test_measures
# Run the separate private core tests with:
#   python3 -m unittest discover -s py/performance/core -t . -p 'test_*.py'
# Neither command requires third-party Python packages.
   
"""
Test input data from the 'Portfolio bacon' dataset
from the PerformanceAnalytics R package.

https://www.rdocumentation.org/packages/PerformanceAnalytics/versions/2.0.4/topics/portfolio_bacon

The data is taken from the page 65 (portfolio) and page 66 (benchmark)
of the:

Carl R Bacon
**Practical portfolio performance : measurement and attribution**
2nd ed, Wiley finance series, 2008
ISBN 978-0-470-05928-9

Hence the name 'bacon portfolio'.

Note we use data from the second (not third) edition because it is included in the R package.

if(!require('PerformanceAnalytics')) {
    install.packages('PerformanceAnalytics')
    library('PerformanceAnalytics')
}
data(portfolio_bacon)
head(portfolio_bacon, 100)
write.csv(portfolio_bacon)
"""
bacon_dates_previous = [
    datetime(2024,6,30), datetime(2024,7,1), datetime(2024,7,2), datetime(2024,7,3),
    datetime(2024,7,4), datetime(2024,7,5), datetime(2024,7,6), datetime(2024,7,7),
    datetime(2024,7,8), datetime(2024,7,9), datetime(2024,7,10), datetime(2024,7,11),
    datetime(2024,7,12), datetime(2024,7,13), datetime(2024,7,14), datetime(2024,7,15),
    datetime(2024,7,16), datetime(2024,7,17), datetime(2024,7,18), datetime(2024,7,19),
    datetime(2024,7,20), datetime(2024,7,21), datetime(2024,7,22), datetime(2024,7,23),
]
bacon_dates = [
    datetime(2024,7,1), datetime(2024,7,2), datetime(2024,7,3),datetime(2024,7,4),
    datetime(2024,7,5), datetime(2024,7,6), datetime(2024,7,7),datetime(2024,7,8),
    datetime(2024,7,9), datetime(2024,7,10), datetime(2024,7,11),datetime(2024,7,12),
    datetime(2024,7,13), datetime(2024,7,14), datetime(2024,7,15),datetime(2024,7,16),
    datetime(2024,7,17), datetime(2024,7,18), datetime(2024,7,19),datetime(2024,7,20),
    datetime(2024,7,21), datetime(2024,7,22), datetime(2024,7,23),datetime(2024,7,24),
]
bacon_dates_previous_monthly = [
    datetime(2024,6,1), datetime(2024,7,1), datetime(2024,8,1), datetime(2024,9,1),
    datetime(2024,10,1), datetime(2024,11,1), datetime(2024,12,1), datetime(2025,1,1),
    datetime(2025,2,1), datetime(2025,3,1), datetime(2025,4,1), datetime(2025,5,1),
    datetime(2025,6,1), datetime(2025,7,1), datetime(2025,8,1), datetime(2025,9,1),
    datetime(2025,10,1), datetime(2025,11,1), datetime(2025,12,1), datetime(2026,1,1),
    datetime(2026,2,1), datetime(2026,3,1), datetime(2026,4,1), datetime(2026,5,1),
]
bacon_dates_monthly = [
    datetime(2024,7,1), datetime(2024,8,1), datetime(2024,9,1), datetime(2024,10,1),
    datetime(2024,11,1), datetime(2024,12,1), datetime(2025,1,1),datetime(2025,2,1),
    datetime(2025,3,1), datetime(2025,4,1), datetime(2025,5,1), datetime(2025,6,1),
    datetime(2025,7,1), datetime(2025,8,1), datetime(2025,9,1), datetime(2025,10,1),
    datetime(2025,11,1), datetime(2025,12,1), datetime(2026,1,1), datetime(2026,2,1),
    datetime(2026,3,1), datetime(2026,4,1), datetime(2026,5,1), datetime(2026,6,1),
]
bacon_portfolio_returns = [
    0.003, 0.026, 0.011,-0.010,
    0.015, 0.025, 0.016, 0.067,
    -0.014,0.040,-0.005, 0.081,
    0.040,-0.037,-0.061, 0.017,
    -0.049,-0.022,0.070, 0.058,
    -0.065,0.024,-0.005,-0.009,
]
bacon_benchmark_returns = [
    0.002, 0.025, 0.018,-0.011,
    0.014, 0.018, 0.014, 0.065,
    -0.015,0.042,-0.006, 0.083,
    0.039,-0.038,-0.062, 0.015,
    -0.048,0.021, 0.060, 0.056,
    -0.067,0.019,-0.003, 0.000,
]
bacon_portfolio_len = len(bacon_portfolio_returns)

"""
Extended Bacon 2023 (3rd edition) portfolio data

Bacon, Carl R.,
Practical portfolio performance measurement and attribution
Third edition. Hoboken NJ, Wiley, 2023
ISBN 9781119831945
"""
bacon_2023_dates_previous = [
    datetime(2024,6,30), datetime(2024,7,1), datetime(2024,7,2), datetime(2024,7,3),
    datetime(2024,7,4), datetime(2024,7,5), datetime(2024,7,6), datetime(2024,7,7),
    datetime(2024,7,8), datetime(2024,7,9), datetime(2024,7,10), datetime(2024,7,11),
    datetime(2024,7,12), datetime(2024,7,13), datetime(2024,7,14), datetime(2024,7,15),
    datetime(2024,7,16), datetime(2024,7,17), datetime(2024,7,18), datetime(2024,7,19),
    datetime(2024,7,20), datetime(2024,7,21), datetime(2024,7,22), datetime(2024,7,23),
    datetime(2024,7,24), datetime(2024,7,25), datetime(2024,7,26), datetime(2024,7,27),
    datetime(2024,7,28), datetime(2024,7,29), datetime(2024,7,30), datetime(2024,7,31),
    datetime(2024,8,1), datetime(2024,8,2), datetime(2024,8,3), datetime(2024,8,4)]
bacon_2023_dates = [
    datetime(2024,7,1), datetime(2024,7,2), datetime(2024,7,3),datetime(2024,7,4),
    datetime(2024,7,5), datetime(2024,7,6), datetime(2024,7,7),datetime(2024,7,8),
    datetime(2024,7,9), datetime(2024,7,10), datetime(2024,7,11),datetime(2024,7,12),
    datetime(2024,7,13), datetime(2024,7,14), datetime(2024,7,15),datetime(2024,7,16),
    datetime(2024,7,17), datetime(2024,7,18), datetime(2024,7,19),datetime(2024,7,20),
    datetime(2024,7,21), datetime(2024,7,22), datetime(2024,7,23),datetime(2024,7,24),
    datetime(2024,7,25), datetime(2024,7,26), datetime(2024,7,27),datetime(2024,7,28),
    datetime(2024,7,29), datetime(2024,7,30), datetime(2024,7,31),datetime(2024,8,1),
    datetime(2024,8,2), datetime(2024,8,3), datetime(2024,8,4),datetime(2024,8,5)]
bacon_2023_portfolio_returns = [
    0.003, 0.026, 0.011, -0.009, 0.014, 0.024, 0.015, 0.066, -0.014, 0.039,
    -0.005, 0.081, 0.040, -0.037, -0.061, 0.014, -0.049, -0.021, 0.062, 0.058,
    -0.064, 0.017, -0.004, -0.002, -0.021, 0.011, 0.047, 0.024, 0.033, -0.007,
    0.047, 0.006, 0.010, -0.002, 0.034, 0.010]
bacon_2023_drawdown_continuous = [
    0, 0, 0, -0.0090, 0, 0, 0, 0, -0.0140, 0,
    -0.0050, 0, 0, 0, -0.0960, 0, 0, -0.0690, 0, 0,
    -0.0640, 0, 0, 0, -0.0270, 0, 0, 0, 0, -0.0070,
    0, 0, 0, -0.0020, 0, 0]
bacon_2023_drawdown_continuous_without_zeroes = [
    -0.0090, -0.0140, -0.0050, -0.0960, -0.0690,
    -0.0640, -0.0270, -0.0070, -0.0020]
bacon_2023_drawdown_from_peak = [
    0, 0, 0, -0.0090, 0, 0, 0, 0, -0.0140, 0,
    -0.0050, 0, 0, -0.0370, -0.0957, -0.0831, -0.1280, -0.1463, -0.0934, -0.0408,
    -0.1022, -0.0869, -0.0906, -0.0924, -0.1115, -0.1017, -0.0595, -0.0369, -0.0051, -0.0121,
    0, 0, 0, -0.0020, 0, 0]
bacon_2023_portfolio_len = len(bacon_2023_portfolio_returns)

"""
Reference data in this file were generated using the R package
PerformanceAnalytics in the online R interpreter:.

Package:
    https://github.com/braverock/PerformanceAnalytics
    
Generation script in online interpreter:
    https://www.datacamp.com/datalab/w/28c21593-21e6-47d9-8e72-acebdd3be32c/edit

The script uses the built-in `portfolio_bacon` (second edition) dataset.

To regenerate the data, see ./reference-data/<name>.R.
"""

def assertFloatEqual(testcase: unittest.TestCase, actual, expected, places=15, delta=None,
                     prefix="", rel_tol=None):
    if math.isnan(expected):
        testcase.assertTrue(math.isnan(actual),
            msg=f'{prefix}: expected NaN, actual {actual}')
    elif math.isinf(expected) and expected > 0:
        testcase.assertTrue(math.isinf(actual) and actual > 0,
            msg=f'{prefix}: expected +Inf, actual {actual}')
    elif math.isinf(expected) and expected < 0:
        testcase.assertTrue(math.isinf(actual) and actual < 0,
            msg=f'{prefix}: expected -Inf, actual {actual}')
    else:
        testcase.assertFalse(math.isnan(actual),
            msg=f'{prefix}: expected {expected}, got {actual}')
        testcase.assertFalse(math.isinf(actual),
            msg=f'{prefix}: expected {expected}, got {actual}')
        if rel_tol is not None:
            testcase.assertTrue(math.isclose(actual, expected, rel_tol=rel_tol,
                                             abs_tol=delta or 0.0),
                                msg=f'{prefix}: expected {expected}, got {actual}')
        elif delta is None:
            testcase.assertAlmostEqual(actual, expected, places=places,
                msg=f'{prefix}: expected {expected}, got {actual}')
        else:
            testcase.assertAlmostEqual(actual, expected, delta=delta,
                msg=f'{prefix}: expected {expected}, got {actual}')

def assertSeriesEqual(testcase, actual, expected, places=15, delta=None, prefix="",
                      skip=0, rel_tol=None):
    actual = list(actual)
    expected = list(expected)
    testcase.assertEqual(len(actual), len(expected), msg=f"{prefix}: series length")
    for i, (a, e) in enumerate(zip(actual, expected)):
        if i >= skip:
            assertFloatEqual(testcase, a, e, places=places, delta=delta,
                             prefix=f"{prefix} step {i}", rel_tol=rel_tol)

def periods_per_annum(daily: bool = False, monthly: bool = False) -> float:
    if daily and monthly:
        raise ValueError("Only one of daily or monthly can be True")
    return 252 if daily else (12 if monthly else 1)

def run_stream_callback(callback,
    daily: bool = False, monthly: bool = False,
    annual_risk_free_rate:float = 0, annual_target_return: float = 0,
    returns = bacon_portfolio_returns, benchmark_returns = bacon_benchmark_returns,
    rolling_window_size = 0, start = 0, **kwargs) -> list:
    measures = Measures(
        periods_per_annum=periods_per_annum(daily=daily, monthly=monthly),
        annual_risk_free_rate=annual_risk_free_rate,
        annual_target_return=annual_target_return,
        rolling_window_size=rolling_window_size)
    measures.reset()
    results = []
    for i in range(start, len(returns)):
        measures.add_return(
            ret=returns[i],
            ret_bench=benchmark_returns[i])
        results.append(callback(measures))
    return results

def run_stream_method(method_name: str,
    daily: bool = False, monthly: bool = False,
    annual_risk_free_rate:float = 0, annual_target_return: float = 0,
    returns = bacon_portfolio_returns, benchmark_returns = bacon_benchmark_returns,
    rolling_window_size = 0, start = 0, **kwargs) -> list:
    measures = Measures(
        periods_per_annum=periods_per_annum(daily=daily, monthly=monthly),
        annual_risk_free_rate=annual_risk_free_rate,
        annual_target_return=annual_target_return,
        rolling_window_size=rolling_window_size)
    measures.reset()
    results = []
    for i in range(start, len(returns)):
        measures.add_return(
            ret=returns[i],
            ret_bench=benchmark_returns[i])
        results.append(getattr(measures, method_name)(**kwargs))
    return results

def run_stream_property(property_name: str,
    daily: bool = False, monthly: bool = False,
    annual_risk_free_rate:float = 0, annual_target_return: float = 0,
    returns = bacon_portfolio_returns, benchmark_returns = bacon_benchmark_returns,
    rolling_window_size = 0, start = 0, **kwargs) -> list:
    measures = Measures(
        periods_per_annum=periods_per_annum(daily=daily, monthly=monthly),
        annual_risk_free_rate=annual_risk_free_rate,
        annual_target_return=annual_target_return,
        rolling_window_size=rolling_window_size)
    measures.reset()
    results = []
    for i in range(start, len(returns)):
        measures.add_return(
            ret=returns[i],
            ret_bench=benchmark_returns[i])
        results.append(getattr(measures, property_name))
    return results

def make_measures(rolling_window_size=0, annual_rf: float = 0.0, annual_mar: float = 0.0,
    daily: bool = False, monthly: bool = False) -> Measures:
    measures = Measures(
        periods_per_annum=periods_per_annum(daily=daily, monthly=monthly),
        annual_risk_free_rate=annual_rf,
        annual_target_return=annual_mar,
        rolling_window_size=rolling_window_size)
    measures.reset()
    return measures

def add_bacon(measures: Measures, start: int = 0, count: int = bacon_portfolio_len,
    returns = bacon_portfolio_returns, benchmark_returns = bacon_benchmark_returns) -> None:
    for i in range(start, count):
        measures.add_return(
            ret=returns[i],
            ret_bench=benchmark_returns[i])

SQRT2 = 1.4142135623730950488016887242097

def public_measures() -> list[str]:
    """Names of all public properties and methods of Measures, except mutators."""
    names = []
    for name, member in inspect.getmembers(Measures):
        if name.startswith('_') or name in ('reset', 'add_return'):
            continue
        if isinstance(member, property) or inspect.isfunction(member):
            names.append(name)
    return names

def evaluate(measures: Measures, name: str):
    """Value of a property, or of a method called with its default arguments."""
    member = getattr(type(measures), name)
    return getattr(measures, name) if isinstance(member, property) else getattr(measures, name)()

class TestSeriesAssertions(unittest.TestCase):
    def test_rejects_truncated_series(self):
        with self.assertRaises(AssertionError):
            assertSeriesEqual(self, [1.0], [1.0, 2.0])
        with self.assertRaises(AssertionError):
            assertSeriesEqual(self, [1.0, 2.0], [1.0])

    def test_accepts_generator(self):
        assertSeriesEqual(self, (x for x in [1.0, 2.0]), [1.0, 2.0])

    def test_relative_tolerance_for_large_reference_values(self):
        assertSeriesEqual(self, [1e12 + 0.1], [1e12], rel_tol=1e-12)
        with self.assertRaises(AssertionError):
            assertSeriesEqual(self, [1e12 + 2], [1e12], rel_tol=1e-12)

class TestEdgeCases(unittest.TestCase):
    def test_empty(self):
        """No property or method raises before the first return."""
        m = make_measures()
        for name in public_measures():
            with self.subTest(name=name):
                evaluate(m, name)

    def test_single_return(self):
        m = make_measures()
        m.add_return(0.01, 0.02)
        for name in public_measures():
            with self.subTest(name=name):
                evaluate(m, name)
        self.assertTrue(math.isnan(m.sharpe_ratio))
        self.assertAlmostEqual(m.cumulative_geometric_return, 0.01, places=15)

    def test_first_negative_return_is_a_drawdown(self):
        """
        As in PerformanceAnalytics Drawdowns(), the high-water mark starts
        at the initial equity 1.
        """
        m = make_measures()
        m.add_return(-0.05, -0.02)
        m.add_return(0.02, 0.01)
        assertSeriesEqual(self, m.drawdowns_high_watermark, [-0.05, -0.031], places=15)
        assertSeriesEqual(self, m.drawdowns_cumulative, [-0.05, -0.031], places=15)
        self.assertAlmostEqual(m.worst_drawdowns_cumulative, 0.05, places=15)
        self.assertAlmostEqual(m.pain_index, (0.05 + 0.031) / 2, places=15)
        self.assertAlmostEqual(m.drawdown_average, 0.05, places=15)

    def test_long_daily_series(self):
        """Every measure works with more observations than periods per annum."""
        rng = random.Random(1)
        m = Measures(periods_per_annum=252.0)
        for _ in range(300):
            m.add_return(rng.gauss(0.0005, 0.01), rng.gauss(0.0004, 0.01))
        for name in public_measures():
            with self.subTest(name=name):
                evaluate(m, name)
        self.assertTrue(math.isfinite(m.autocorrelation_penalty))

class TestAutocorrelationPenalty(unittest.TestCase):
    """
    Since we don't have data for it, we'll test only its metamorphic properties.
    """
    def test_metamorphic_properties(self):
        # Constant returns
        returns = list(0.01 for _ in range(bacon_2023_portfolio_len))
        actual = run_stream_property("autocorrelation_penalty", daily=True,
                                     returns=returns, benchmark_returns=returns)
        assertFloatEqual(self, actual[bacon_2023_portfolio_len-1], 1.0, places=15,
                        prefix="autocorrelation penalty (constant)")

        # Too few observations
        assertFloatEqual(self, actual[0], 1.0, places=15, prefix="autocorrelation penalty (len=0)")
        assertFloatEqual(self, actual[1], 1.0, places=15, prefix="autocorrelation penalty (len=1)")

        # Positive autocorrelation
        returns = list(0.01 * i for i in range(bacon_2023_portfolio_len))
        actual = run_stream_property("autocorrelation_penalty", daily=True,
                                     returns=returns, benchmark_returns=returns)
        assertFloatEqual(self, actual[bacon_2023_portfolio_len-1], 2.722393904531189, places=15,
                        prefix="autocorrelation penalty (positive)")

        # Negative autocorrelation
        returns = list(0.01 if i % 2 == 0 else -0.01 for i in range(bacon_2023_portfolio_len))
        actual = run_stream_property("autocorrelation_penalty", daily=True,
                                     returns=returns, benchmark_returns=returns)
        assertFloatEqual(self, actual[bacon_2023_portfolio_len-1], 0.16903085094570597, places=15,
                        prefix="autocorrelation penalty (negative)")

        # Scale and translation invariance
        expected = run_stream_property("autocorrelation_penalty", daily=True)
        for scale in (4.2, -4.2):
            for shift in (0.042, -0.042):
                transformed = list(scale * r + shift for r in bacon_portfolio_returns)
                actual = run_stream_property("autocorrelation_penalty", daily=True,
                    returns=transformed, benchmark_returns=transformed)
                assertSeriesEqual(self, actual, expected, places=15,
                    prefix=f'autocorrelation penalty (transform) scale {scale} shift {shift}')

class TestCumulativeGeometricReturn(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.cumulative_geometric_return.EXPECTED_VALUES
        actual = run_stream_property("cumulative_geometric_return")
        assertSeriesEqual(self, actual, expected, places=14,
                              prefix="cumulative geometric return (yearly)")

        actual = run_stream_property("cumulative_geometric_return", monthly=True)
        assertSeriesEqual(self, actual, expected, places=14,
                              prefix="cumulative geometric return (monthly)")

        actual = run_stream_property("cumulative_geometric_return", daily=True)
        assertSeriesEqual(self, actual, expected, places=14,
                              prefix="cumulative geometric return (daily)")

class TestGeometricMeanReturn(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.geometric_mean_return.EXPECTED_VALUES_GEOMETRIC
        actual = run_stream_property("geometric_mean_return")
        assertSeriesEqual(self, actual, expected, places=15,
                              prefix="geometric mean return (yearly)")

        actual = run_stream_property("geometric_mean_return", monthly=True)
        assertSeriesEqual(self, actual, expected, places=15,
                              prefix="geometric mean return (monthly)")

        actual = run_stream_property("geometric_mean_return", daily=True)
        assertSeriesEqual(self, actual, expected, places=15,
                              prefix="geometric mean return (daily)")

class TestCompoundAnnualGrowthRate(unittest.TestCase):
    def test_annualized_return_definition(self):
        def calculate(measures: Measures, periods_per_annum: float) -> float:
            growth = math.prod(1 + r for r in measures._returns)
            return growth ** (periods_per_annum / len(measures._returns)) - 1

        expected = run_stream_callback(lambda r: calculate(r, 1))
        actual = run_stream_property("compound_annual_growth_rate")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix="compound annual growth rate (yearly)")

        expected = run_stream_callback(lambda r: calculate(r, 12), monthly=True)
        actual = run_stream_property("compound_annual_growth_rate", monthly=True)
        assertSeriesEqual(self, actual, expected, places=14,
                          prefix="compound annual growth rate (monthly)")

        expected = run_stream_callback(lambda r: calculate(r, 252), daily=True)
        actual = run_stream_property("compound_annual_growth_rate", daily=True)
        assertSeriesEqual(self, actual, expected, places=11,
                          prefix="compound annual growth rate (daily)")

class TestSkewness(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        We implement hardcoded `bias=True` in RawMomentsKleinKBN,
        so we test only `moment` method here.

        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for method, expected in rd.skewness.EXPECTED_VALUES_BY_METHOD.items():
            name = f'skewness_{method}'
            actual = run_stream_property(name)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'{name}')
            if method == "moment":
                actual = run_stream_property("skewness")
                assertSeriesEqual(self, actual, expected, places=14,
                                  prefix="skewness")

    def test_raw_moments_klein_kbn(self):
        """
        The mapping between Performance Analytics package naming
        conventions and parameters of RawMomentsKleinKBN:

        - *moment*: bias=True
        - *fisher*: bias=False
        - *sample*: using dedicated `skewness_sample` property
        """
        for method, expected in rd.skewness.EXPECTED_VALUES_BY_METHOD.items():
            bias = False if method == 'fisher' else True
            kbn = RawMomentsKleinKBN(ddof=1, bias=bias, fisher=True)
            for i in range(bacon_portfolio_len):
                kbn.update(bacon_portfolio_returns[i])
                if method == 'moment':
                    actual = kbn.skewness_moment
                    assertFloatEqual(self, actual, kbn.skewness, places=15,
                                     prefix=f'step {i} skewness_{method} vs. skewness')
                elif method == 'fisher':
                    actual = kbn.skewness_fisher
                    assertFloatEqual(self, actual, kbn.skewness, places=15,
                                     prefix=f'step {i} skewness_{method} vs. skewness')
                else: # method == 'sample'
                    actual = kbn.skewness_sample
                    assertFloatEqual(self, actual, expected[i], places=14,
                                     prefix=f'step {i} skewness_{method}')

class TestKurtosis(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        We implement hardcoded `bias=True, fisher=True` in RawMomentsKleinKBN,
        so the `kurtosis` is always `excess` here.

        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for method, expected in rd.kurtosis.EXPECTED_VALUES_BY_METHOD.items():
            name = f'kurtosis_{method}'
            actual = run_stream_property(name)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'{name}')
            if method == "excess":
                actual = run_stream_property("kurtosis")
                assertSeriesEqual(self, actual, expected, places=13,
                                  prefix="kurtosis")

    def test_raw_moments_klein_kbn(self):
        """
        The mapping between R package naming conventions and parameters of RawMomentsKleinKBN:

        - *excess* or *fisher*: bias=True, fisher=True
        - *moment* or *pearson*`: bias=True, fisher=False
        - *sample_excess*: bias=False, fisher=True
        - *sample*: no parameter combination, use `kurtosis_sample_corrected`;
          `kurtosis` with bias=False, fisher=False returns `kurtosis_sample`
          (scipy-compatible G2+3), which differs from the R value.

        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for method, expected in rd.kurtosis.EXPECTED_VALUES_BY_METHOD.items():
            # Default to 'excess'.
            bias=True
            fisher=True
            if method  == 'moment':
                fisher=False
            elif method  == 'sample_corrected':
                bias=False
                fisher=False
            elif method  == 'sample_excess':
                bias=False
            kbn  = RawMomentsKleinKBN(ddof=1, bias=bias, fisher=fisher)
            for i in range(bacon_portfolio_len):
                kbn.update(bacon_portfolio_returns[i])
                if method == 'excess':
                    actual = kbn.kurtosis_excess
                elif method  == 'moment':
                    actual = kbn.kurtosis_moment
                elif method  == 'sample_corrected':
                    actual = kbn.kurtosis_sample_corrected
                else: # method  == 'sample_excess'
                    actual = kbn.kurtosis_sample_excess
                dispatched = kbn.kurtosis_sample if method == 'sample_corrected' else actual
                assertFloatEqual(self, dispatched, kbn.kurtosis, places=15,
                                     prefix=f'step {i} kurtosis_{method} / kurtosis')
                assertFloatEqual(self, actual, expected[i], places=13,
                                     prefix=f'step {i} kurtosis_{method}')

class TestSkewnessKurtosisRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        expected = rd.skewness_kurtosis_ratio.EXPECTED_VALUES
        actual = run_stream_property("skewness_kurtosis_ratio")
        assertSeriesEqual(self, actual, expected, places=14,
                          prefix=f'skewness-kurtosis ratio')

class TestJarqueBeraNrmalityTestStatistic(unittest.TestCase):
    """
    Reference values were generated by scipy.

    Calculation doesn't depend on periods per annum, so we use yearly default.
    """
    def test_matches_bacon3_output(self):
        actual = run_stream_property("jarque_bera_normality_test_statistic",
           returns=bacon_2023_portfolio_returns, benchmark_returns=bacon_2023_portfolio_returns)
        expected = 0.34 # Chapter 5, exhibit 5.4
        assertFloatEqual(self, actual[bacon_2023_portfolio_len-1], expected, places=2,
                         prefix="Jarque-Bera normality (bacon3)")

    def test_matches_scipy_output(self):
        expected = rd.jarque_bera_normality_test_statistic.EXPECTED_VALUES
        actual = run_stream_property("jarque_bera_normality_test_statistic")
        assertSeriesEqual(self, actual, expected, places=14,
                          prefix='Jarque-Bera normality (scipy)')

class TestIsNormalDistribution(unittest.TestCase):
    def test_mocked_jb(self):
        ratios = Measures(periods_per_annum=1,
            annual_risk_free_rate=0, annual_target_return=0)
        with patch.object(type(ratios), "jarque_bera_normality_test_statistic",
            new_callable=PropertyMock) as jb:

                    # Test normality accepted
                    # Since 5.0 < 5.991..., the method should return True.
                    jb.return_value = 5.0
                    actual = ratios.is_normal_distribution()
                    self.assertTrue(actual, msg=f'normality accepted')

                    # Test normality rejected
                    jb.return_value = 10.0
                    actual = ratios.is_normal_distribution()
                    self.assertFalse(actual, msg=f'normality rejected')

                    # Test NaN statistic
                    jb.return_value = math.nan
                    actual = ratios.is_normal_distribution()
                    self.assertFalse(actual, msg=f'NaN statistic')

                    # Test invalid confidence
                    jb.return_value = 0.0
                    with self.assertRaises(ValueError):
                        ratios.is_normal_distribution(1.0)
                    with self.assertRaises(ValueError):
                        ratios.is_normal_distribution(0.0)

                    # Test custom confidence
                    jb.return_value = 8.0
                    actual = ratios.is_normal_distribution(0.99)
                    self.assertTrue(actual, msg=f'custom confidence 8')
                    jb.return_value = 10.0
                    actual = ratios.is_normal_distribution(0.99)
                    self.assertFalse(actual, msg=f'custom confidence 10')

class TestVarCornishFisher(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, expected in rd.var.EXPECTED_VALUES_BY_P_CORNISH_FISHER.items():
            actual = run_stream_method("var_cornish_fisher", confidence=p)
            assertSeriesEqual(self, actual, expected, places=9,
                              prefix=f'var cornish-fisher p {p}')

class TestVarGaussian(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, expected in rd.var.EXPECTED_VALUES_BY_P_GAUSSIAN.items():
            actual = run_stream_method("var_gaussian", confidence=p)
            assertSeriesEqual(self, actual, expected, places=9,
                              prefix=f'var gaussian p {p}')

class TestVarHistorical(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, expected in rd.var.EXPECTED_VALUES_BY_P_HISTORICAL.items():
            actual = run_stream_method("var_historical", confidence=p)
            assertSeriesEqual(self, actual, expected, places=4 if p == 0.999 else 15,
                              prefix=f'var historical p {p}')

class TestEsCornishFisher(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, expected in rd.es.EXPECTED_VALUES_BY_P_CORNISH_FISHER.items():
            actual = run_stream_method("es_cornish_fisher", confidence=p)
            assertSeriesEqual(self, actual, expected, places=9 if p < 0.995 else (8 if p < 0.999 else 7),
                              prefix=f'es cornish-fisher p {p}')

class TestEsGaussian(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, expected in rd.es.EXPECTED_VALUES_BY_P_GAUSSIAN.items():
            actual = run_stream_method("es_gaussian", confidence=p)
            assertSeriesEqual(self, actual, expected, places=9 if p < 0.995 else 8,
                              prefix=f'es gaussian p {p}')

class TestEsHistorical(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, expected in rd.es.EXPECTED_VALUES_BY_P_HISTORICAL.items():
            actual = run_stream_method("es_historical", confidence=p)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'es historical p {p}')

class TestRewardToVarEsRatios(unittest.TestCase):
    """
    reward_to_var_ratio_* and reward_to_es_ratio_* divide the mean excess
    return by the VaR/ES of the raw returns.
    """
    NAMES = ['var_historical', 'var_gaussian', 'var_cornish_fisher',
             'es_historical', 'es_gaussian', 'es_cornish_fisher']

    def test_zero_risk_free_rate_equals_sharpe_variants(self):
        """With a zero risk-free rate, they equal sharpe_ratio_var_* / sharpe_ratio_es_*."""
        for name in self.NAMES:
            reward = run_stream_method(f'reward_to_{name.split("_")[0]}_ratio_{name.split("_", 1)[1]}')
            sharpe = run_stream_method(f'sharpe_ratio_{name}')
            assertSeriesEqual(self, reward, sharpe, places=14, prefix=f'reward_to {name}', skip=1)

    def test_definition(self):
        annual_rf = 0.05
        for name in self.NAMES:
            reward_name = f'reward_to_{name.split("_")[0]}_ratio_{name.split("_", 1)[1]}'
            for confidence in (0.9, 0.95):
                m = make_measures(annual_rf=annual_rf, monthly=True)
                for i in range(bacon_portfolio_len):
                    m.add_return(bacon_portfolio_returns[i], bacon_benchmark_returns[i])
                    excess_mean = math.fsum(r - m.risk_free_rate
                        for r in bacon_portfolio_returns[:i + 1]) / (i + 1)
                    denom = getattr(m, name)(confidence=confidence)
                    expected = excess_mean / denom if denom != 0 else math.nan
                    actual = getattr(m, reward_name)(confidence=confidence)
                    assertFloatEqual(self, actual, expected, places=14,
                        prefix=f'{reward_name} confidence {confidence} step {i}')

class TestMeanAbsoluteDeviationRatio(unittest.TestCase):
    def test_exact_values(self):
        """
        mean / (sum|r - mean| / n) on the Bacon portfolio returns, computed
        with exact rational arithmetic (fractions.Fraction).
        """
        expected = [
            math.nan, 1.2608695652173914, 1.5789473684210527, 0.6818181818181818,
            0.9, 1.129032258064516, 1.308695652173913, 1.2618556701030927,
            0.9623076923076923, 1.0358796296296295, 0.9166666666666666, 0.96045197740113,
            1.0325794291868604, 0.7659033078880407, 0.480644111906311, 0.5178463399879009,
            0.3424072265625, 0.276536312849162, 0.3738222796970257, 0.4428828239908482,
            0.29573420836751435, 0.3247753530166881, 0.3100659077291792, 0.289544235924933]
        actual = run_stream_property("mean_absolute_deviation_ratio")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix="mean absolute deviation ratio")

class TestUpsidePotentialRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.upside_potential_ratio.EXPECTED_VALUES_BY_MAR_FULL.items():
            actual = run_stream_property("upside_potential_ratio", annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'dupside potential ratio (full) MAR {mar}')

class TestUpsidePotentialRatioSubset(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.upside_potential_ratio.EXPECTED_VALUES_BY_MAR_SUBSET.items():
            actual = run_stream_property("upside_potential_ratio_subset", annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'dupside potential ratio (subset) MAR {mar}')

class TestUpsideFrequency(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.upside_frequency.EXPECTED_VALUES_BY_MAR.items():
            actual = run_stream_property("upside_frequency", annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'upside frequency MAR {mar}')

class TestUpsidePotential(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.upside_risk.EXPECTED_VALUES_BY_MAR_POTENTIAL_FULL.items():
            actual = run_stream_property("upside_potential", annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'upside potential (full) MAR {mar}')
        for mar, expected in rd.upside_risk.EXPECTED_VALUES_BY_MAR_POTENTIAL_FULL.items():
            annual_mar = (1 + mar) ** 252 - 1
            actual = run_stream_property("upside_potential",
                                         daily=True, annual_target_return=annual_mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'upside potential (full) daily MAR {mar}')

class TestUpsidePotentialSubset(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.upside_risk.EXPECTED_VALUES_BY_MAR_POTENTIAL_SUBSET.items():
            actual = run_stream_property("upside_potential_subset",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'upside potential (subset) MAR {mar}')

class TestUpsideVariance(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.upside_risk.EXPECTED_VALUES_BY_MAR_VARIANCE_FULL.items():
            actual = run_stream_property("upside_variance",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'upside variance (full) MAR {mar}')

class TestUpsideVarianceSubset(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.upside_risk.EXPECTED_VALUES_BY_MAR_VARIANCE_SUBSET.items():
            actual = run_stream_property("upside_variance_subset",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'upside variance (subset) MAR {mar}')

class TestUpsideRisk(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.upside_risk.EXPECTED_VALUES_BY_MAR_RISK_FULL.items():
            actual = run_stream_property("upside_risk",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'upside risk (full) MAR {mar}')

class TestUpsideRiskSubset(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.upside_risk.EXPECTED_VALUES_BY_MAR_RISK_SUBSET.items():
            actual = run_stream_property("upside_risk_subset",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'upside risk (subset) yearly MAR {mar}')
        for mar, expected in rd.upside_risk.EXPECTED_VALUES_BY_MAR_RISK_SUBSET.items():
            annual_mar = (1 + mar) ** 252 - 1
            actual = run_stream_property("upside_risk_subset",
                                         daily=True, annual_target_return=annual_mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'upside risk (subset) daily MAR {mar}')

class TestSemiDeviation(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        actual = run_stream_property("semi_deviation")
        assertSeriesEqual(self, actual, rd.semi_deviation.EXPECTED_VALUES, places=15,
            prefix='semi-deviation')

class TestDownsideDeviation(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.downside_deviation.EXPECTED_VALUES_BY_MAR_FULL.items():
            actual = run_stream_property("downside_deviation",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'downside deviation MAR {mar}')

class TestDownsideDeviationSubset(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.downside_deviation.EXPECTED_VALUES_BY_MAR_SUBSET.items():
            actual = run_stream_property("downside_deviation_subset",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'downside deviation subset MAR {mar}')

class TestDownsideFrequency(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.downside_frequency.EXPECTED_VALUES_BY_MAR.items():
            actual = run_stream_property("downside_frequency",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'downside frequency MAR {mar}')

class TestDownsidePotential(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.downside_potential.EXPECTED_VALUES_BY_MAR.items():
            actual = run_stream_property("downside_potential",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'downside potential MAR {mar}')

class TestSharpeRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for rf, expected in rd.sharpe_ratio.EXPECTED_VALUES_BY_RF_STDEV.items():
            actual = run_stream_property("sharpe_ratio",
                                         annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=13 if rf < 0.25 else 12,
                              prefix=f'Sharpe ratio (stdev) Rf {rf}')

class TestSharpeRatioVarHistorical(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, rf_pack in rd.sharpe_ratio.EXPECTED_VALUES_BY_P_RF_VAR_HISTORICAL.items():
            for rf, expected in rf_pack.items():
                actual = run_stream_method("sharpe_ratio_var_historical",
                                           annual_risk_free_rate=rf, confidence=p)
                assertSeriesEqual(self, actual, expected, places=12,
                                  prefix=f'Sharpe ratio (VaR historical) conf {p} Rf {rf}')

class TestSharpeRatioVarGaussian(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, rf_pack in rd.sharpe_ratio.EXPECTED_VALUES_BY_P_RF_VAR_GAUSSIAN.items():
            for rf, expected in rf_pack.items():
                actual = run_stream_method("sharpe_ratio_var_gaussian",
                                           annual_risk_free_rate=rf, confidence=p)
                assertSeriesEqual(self, actual, expected, places=5,
                                  prefix=f'Sharpe ratio (VaR Gaussian) conf {p} Rf {rf}')

class TestSharpeRatioVarCornishFisher(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, rf_pack in rd.sharpe_ratio.EXPECTED_VALUES_BY_P_RF_VAR_CORNISH_FISHER.items():
            for rf, expected in rf_pack.items():
                actual = run_stream_method("sharpe_ratio_var_cornish_fisher",
                                           annual_risk_free_rate=rf, confidence=p)
                assertSeriesEqual(self, actual, expected, places=6,
                                  prefix=f'Sharpe ratio (VaR Cornish-Fisher) conf {p} Rf {rf}')

class TestSharpeRatioEsHistorical(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, rf_pack in rd.sharpe_ratio.EXPECTED_VALUES_BY_P_RF_ES_HISTORICAL.items():
            for rf, expected in rf_pack.items():
                actual = run_stream_method("sharpe_ratio_es_historical",
                                           annual_risk_free_rate=rf, confidence=p)
                self.assertEqual(len(actual), len(expected))
                for i, (a, e) in enumerate(zip(actual, expected)):
                    prefix = f'Sharpe ratio (ES historical) conf {p} Rf {rf} step {i}'
                    if p == 0.9 and rf == 0.001 and i == 10:
                        # The quantile is exactly the second-worst return.
                        # This implementation includes both tied-to-tail
                        # observations; the R reference includes only one.
                        excess_mean = math.fsum(r - rf for r in bacon_portfolio_returns[:11]) / 11
                        self.assertAlmostEqual(a, excess_mean / 0.013, places=13,
                                               msg=prefix)
                    else:
                        assertFloatEqual(self, a, e, delta=1e-12, prefix=prefix)

class TestSharpeRatioEsGaussian(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, rf_pack in rd.sharpe_ratio.EXPECTED_VALUES_BY_P_RF_ES_GAUSSIAN.items():
            for rf, expected in rf_pack.items():
                actual = run_stream_method("sharpe_ratio_es_gaussian",
                                           annual_risk_free_rate=rf, confidence=p)
                assertSeriesEqual(self, actual, expected, places=7,
                                  prefix=f'Sharpe ratio (ES Gaussian) conf {p} Rf {rf}')

class TestSharpeRatioEsCornishFisher(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for p, rf_pack in rd.sharpe_ratio.EXPECTED_VALUES_BY_P_RF_ES_CORNISH_FISHER.items():
            for rf, expected in rf_pack.items():
                actual = run_stream_method("sharpe_ratio_es_cornish_fisher",
                                           annual_risk_free_rate=rf, confidence=p)
                assertSeriesEqual(self, actual, expected, places=5,
                                  prefix=f'Sharpe ratio (ES Cornish-Fisher) conf {p} Rf {rf}')

class TestDownsideSharpeRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for rf, expected in rd.downside_sharpe_ratio.EXPECTED_VALUES_BY_RF.items():
            actual = run_stream_property("downside_sharpe_ratio",
                                         annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'downside Sharpe ratio Rf {rf}')

class TestAdjustedSharpeRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for rf, expected in rd.adjusted_sharpe_ratio.EXPECTED_VALUES_BY_RF.items():
            actual = run_stream_property("adjusted_sharpe_ratio",
                                         annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=12 if rf < 0.2 else 11,
                              prefix=f'adjusted Sharpe ratio (stdev) Rf {rf}')

class TestAdjustedSharpeRatioSkewOnly(unittest.TestCase):
    def test_scale_and_translation_invariance(self):
        """
        Since we don't have data for it, we'll test only its metamorphic properties.
        """
        rf = 0.0042
        expected = run_stream_property("adjusted_sharpe_ratio_skew_only", annual_risk_free_rate=rf)

        for scale in (4.2, -4.2):
            for shift in (0.042, -0.042):
                transformed = list(scale * r + shift for r in bacon_portfolio_returns)
                ratios = Measures(
                    periods_per_annum=1,
                    annual_risk_free_rate=scale * rf + shift,
                    annual_target_return=0)
                ratios.reset()
                for i in range(bacon_portfolio_len):
                    ratios.add_return(
                        ret=transformed[i],
                        ret_bench=transformed[i])
                    a = ratios.adjusted_sharpe_ratio_skew_only
                    assertFloatEqual(self, a if scale > 0 else -a, expected[i],
                        places=14, prefix=f"ASR skew-only (scale {scale} shift {shift}) step {i}")

class TestProbabilisticSharpeRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for ref_sr, rf_pack in rd.probabilistic_sharpe_ratio.EXPECTED_VALUES_BY_REFSR_RF.items():
            for rf, expected in rf_pack.items():
                actual = run_stream_method("probabilistic_sharpe_ratio",
                                           annual_risk_free_rate=rf, reference_sr=ref_sr)
                assertSeriesEqual(self, actual, expected, places=14,
                                  prefix=f'probabilistic Sharpe ratio reference_sr {ref_sr} Rf {rf}')

class TestProbabilisticSharpeRatioFull(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for ref_sr, rf_pack in rd.probabilistic_sharpe_ratio.EXPECTED_VALUES_BY_REFSR_RF_FULL.items():
            for rf, expected in rf_pack.items():
                actual = run_stream_method("probabilistic_sharpe_ratio_full",
                                           annual_risk_free_rate=rf, reference_sr=ref_sr)
                assertSeriesEqual(self, actual, expected, places=14,
                                  prefix=f'probabilistic Sharpe ratio (full) reference_sr {ref_sr} Rf {rf}')

class TestProbabilisticSharpeRatioSymmetric(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for ref_sr, rf_pack in rd.probabilistic_sharpe_ratio.EXPECTED_VALUES_BY_REFSR_RF_SYMMETRIC.items():
            for rf, expected in rf_pack.items():
                actual = run_stream_method("probabilistic_sharpe_ratio_symmetric",
                                           annual_risk_free_rate=rf, reference_sr=ref_sr)
                assertSeriesEqual(self, actual, expected, places=14,
                                  prefix=f'probabilistic Sharpe ratio (symmetric) reference_sr {ref_sr} Rf {rf}')

class TestProbabilisticSharpeRatioGaussian(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for ref_sr, rf_pack in rd.probabilistic_sharpe_ratio.EXPECTED_VALUES_BY_REFSR_RF_GAUSSIAN.items():
            for rf, expected in rf_pack.items():
                actual = run_stream_method("probabilistic_sharpe_ratio_gaussian",
                                           annual_risk_free_rate=rf, reference_sr=ref_sr)
                assertSeriesEqual(self, actual, expected, places=14,
                                  prefix=f'probabilistic Sharpe ratio (Gaussian) reference_sr {ref_sr} Rf {rf}')

class TestSortinoRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.sortino_ratio.EXPECTED_VALUES_BY_MAR.items():
            actual = run_stream_property("sortino_ratio",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'sortino ratio MAR {mar}')

    def test_jack_schager_sqrt2_version(self):
        for mar, expected in rd.sortino_ratio.EXPECTED_VALUES_BY_MAR.items():
            expected = [x / math.sqrt(2) for x in expected]
            actual = run_stream_property("sortino_ratio_sqrt2",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'sortino ratio (sqrt2) MAR {mar}')

class TestSortinoSatchellRatio(unittest.TestCase):
    def test_should_be_computable(self):
        actual = run_stream_property("sortino_satchell_ratio")
        assertFloatEqual(self, actual[len(actual)-1], 0.3923720287950653, places=15,
                              prefix=f'Sortino-Satchell ratio')

class TestOmegaRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.omega_ratio.EXPECTED_VALUES_BY_MAR.items():
            actual = run_stream_property("omega_ratio",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'omega ratio MAR {mar}')

class TestOmegaSharpeRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.omega_sharpe_ratio.EXPECTED_VALUES_BY_MAR.items():
            actual = run_stream_property("omega_sharpe_ratio",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'omega Dharpe ratio MAR {mar}')

class TestOmegaExcessReturn(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for mar, expected in rd.omega_excess_return.EXPECTED_VALUES_BY_MAR_WITH_BENCHMARK.items():
            annual_mar = (1 + mar) ** 252 - 1
            actual = run_stream_property("omega_excess_return", daily=True,
                                         annual_target_return=annual_mar)
            assertSeriesEqual(self, actual, expected, places=11,
                              prefix=f'omega excess return (with benchmak) MAR {mar}')
        for mar, expected in rd.omega_excess_return.EXPECTED_VALUES_BY_MAR_WITH_SELF.items():
            annual_mar = (1 + mar) ** 252 - 1
            actual = run_stream_property("omega_excess_return", daily=True,
                                         annual_target_return=annual_mar,
                                         benchmark_returns=bacon_portfolio_returns)
            assertSeriesEqual(self, actual, expected, places=11,
                              prefix=f'omega excess return (with self) MAR {mar}')

class TestKappaRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Generated data is yearly
        """
        for mar, expected in rd.kappa_ratio.EXPECTED_VALUES_BY_MAR_ORDER_1.items():
            actual = run_stream_property("kappa_1_ratio",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'kappa 1 ratio MAR {mar}')
        for mar, expected in rd.kappa_ratio.EXPECTED_VALUES_BY_MAR_ORDER_2.items():
            actual = run_stream_property("kappa_2_ratio",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'kappa 2 ratio MAR {mar}')
        for mar, expected in rd.kappa_ratio.EXPECTED_VALUES_BY_MAR_ORDER_3.items():
            actual = run_stream_property("kappa_3_ratio",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'kappa 3 ratio MAR {mar}')
        for mar, expected in rd.kappa_ratio.EXPECTED_VALUES_BY_MAR_ORDER_4.items():
            actual = run_stream_property("kappa_4_ratio",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'kappa 4 ratio MAR {mar}')

class TestProspectRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for mar, expected in rd.prospect_ratio.EXPECTED_VALUES_BY_MAR_PERFAN.items():
            actual = run_stream_property("prospect_ratio_performance_analytics",
                 annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'Prospect ratio PerformanceAnalytics version (yearly, MAR {mar})')

    def test_matches_reference_implementation_output(self):
        for mar, expected in rd.prospect_ratio.EXPECTED_VALUES_BY_MAR_REFERENCE.items():
            actual = run_stream_method("prospect_ratio",
                 annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'Prospect ratio (yearly, MAR {mar})')

class TestBernardoLedoitRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        expected = rd.bernardo_ledoit_ratio.EXPECTED_VALUES
        actual = run_stream_property("bernardo_ledoit_ratio")
        assertSeriesEqual(self, actual, expected, places=13, prefix="Bernado-Ledoit ratio")

class TestDRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        expected = rd.d_ratio.EXPECTED_VALUES_PERFAN
        actual = run_stream_property("d_ratio")
        assertSeriesEqual(self, actual, expected, places=15,
                        prefix="d-ratio")

class TestGainLossRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        expected = rd.bernardo_ledoit_ratio.EXPECTED_VALUES
        actual = run_stream_property("gain_loss_ratio")
        assertSeriesEqual(self, actual, expected, places=13,
                          prefix="gain-loss ratio")

class TestMeanNonZeroReturn(unittest.TestCase):
    def test_calculated_by_hand(self):
        expected = rd.mean_non_zero_return.EXPECTED_VALUES
        actual = run_stream_property("mean_non_zero_return")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix="mean non-zero return")

class TestMeanWinReturn(unittest.TestCase):
    def test_calculated_by_hand(self):
        expected = rd.mean_win_return.EXPECTED_VALUES
        actual = run_stream_property("mean_win_return")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix="mean win return")

class TestMeanLossReturn(unittest.TestCase):
    def test_calculated_by_hand(self):
        expected = rd.mean_loss_return.EXPECTED_VALUES
        actual = run_stream_property("mean_loss_return")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix="mean loss return")

class TestWinRate(unittest.TestCase):
    def test_calculated_by_hand(self):
        expected = rd.win_rate.EXPECTED_VALUES
        actual = run_stream_property("win_rate")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix="win rate")

class TestLossRate(unittest.TestCase):
    def test_calculated_by_hand(self):
        expected = rd.loss_rate.EXPECTED_VALUES
        actual = run_stream_property("loss_rate")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix="loss rate")

class TestVolatilitySkewness(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        for mar, expected in rd.volatility_skewness.EXPECTED_VALUES_BY_MAR_VOLATILITY.items():
            actual = run_stream_property("volatility_skewness",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'volatility skewness MAR {mar}')
        for mar, expected in rd.volatility_skewness.EXPECTED_VALUES_BY_MAR_VARIABILITY.items():
            actual = run_stream_property("variability_skewness",
                                         annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'variability skewness MAR {mar}')

class TestFarinelliTibilettiRatio(unittest.TestCase):
    def test_should_be_computable(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        mar = 0.005
        annual_mar = (1 + mar) ** 252 - 1

        def verify(upper: int, lower: int, related_property: str, transform, places: int):
            expected = run_stream_property(related_property,
                upper_order=upper, lower_order=lower, annual_target_return=mar)
            actual = run_stream_method("farinelli_tibiletti_ratio",
                upper_order=upper, lower_order=lower, annual_target_return=mar)
            assertSeriesEqual(self, (transform(r) for r in actual), expected, places=places,
                prefix=f'Farinelli-Tibiletti ratio (u {upper}, l {lower}) vs {related_property}')

        verify(1, 1, "omega_ratio", lambda r: r, 14)
        verify(1, 1, "kappa_1_ratio", lambda r: r - 1, 14)
        verify(1, 2, "upside_potential_ratio", lambda r: r, 15)
        verify(2, 2, "volatility_skewness", lambda r: r, 14)
        verify(2, 2, "variability_skewness", lambda r: r * r, 13)

        def verify_manual(upper: int, lower: int, places: int):
            upm = sum(max(r - mar, 0.0) ** upper for r in bacon_portfolio_returns) / bacon_portfolio_len
            lpm = sum(max(mar - r, 0.0) ** lower for r in bacon_portfolio_returns) / bacon_portfolio_len
            expected = upm ** (1.0 / upper) / lpm ** (1.0/lower)
            actual = run_stream_method("farinelli_tibiletti_ratio",
                upper_order=upper, lower_order=lower, annual_target_return=mar)
            assertFloatEqual(self, actual[len(actual) - 1], expected, places=places,
                prefix=f'Farinelli-Tibiletti ratio (u {upper}, l {lower}) vs manual calculation')

        for i in (1, 2, 3, 4):
            for j in (1, 2, 3, 4):
                verify_manual(i, j, 15)

class TestRachevRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        alpha = 0.05
        for beta, bundle in rd.rachev_ratio.EXPECTED_VALUES_BY_BETA_RF_ALFA_0_05.items():
            for rf, expected in bundle.items():
                actual = run_stream_method("rachev_ratio",
                    alpha=alpha, beta=beta, annual_risk_free_rate=rf)
                assertSeriesEqual(self, actual, expected, places=14,
                                  prefix=f'Rachev ratio (alpha {alpha} beta {beta} Rf {rf})')
        alpha = 0.1
        for beta, bundle in rd.rachev_ratio.EXPECTED_VALUES_BY_BETA_RF_ALFA_0_1.items():
            for rf, expected in bundle.items():
                actual = run_stream_method("rachev_ratio",
                    alpha=alpha, beta=beta, annual_risk_free_rate=rf)
                assertSeriesEqual(self, actual, expected, places=14,
                                  prefix=f'Rachev ratio (alpha {alpha} beta {beta} Rf {rf})')

class TestDrawdownsCumulative(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        actual = run_stream_property("drawdowns_cumulative")
        for i, expected in rd.drawdowns_cumulative.EXPECTED_VALUES_BY_INDEX.items():
            assertSeriesEqual(self, actual[i], expected, places=15,
                              prefix=f'drawdowns cumulative (i {i})')

    def test_matches_bacon_2023_output(self):
        actual = run_stream_property("drawdowns_cumulative",
            returns=bacon_2023_portfolio_returns, benchmark_returns=bacon_2023_portfolio_returns)
        assertSeriesEqual(self, actual[len(actual) - 1], bacon_2023_drawdown_from_peak, places=4,
            prefix=f'drawdowns cumulative (bacon 2023)')

class TestMinDrawdownsCumulative(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        expected = rd.min_drawdowns_cumulative.EXPECTED_VALUES
        actual = run_stream_property("min_drawdowns_cumulative")
        assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'min drawdowns cumulative')

class TestWorstDrawdownsCumulative(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        expected = rd.min_drawdowns_cumulative.EXPECTED_VALUES_INVERTED
        actual = run_stream_property("worst_drawdowns_cumulative")
        assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'worst drawdowns cumulative')

class TestDrawdownsHighWatermark(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Calculation doesn't depend on periods per annum, so we use yearly default.
        """
        actual = run_stream_property("drawdowns_high_watermark")
        for i, expected in rd.drawdowns_high_watermark.EXPECTED_VALUES_BY_INDEX.items():
            assertSeriesEqual(self, actual[i], expected, places=15,
                              prefix=f'drawdowns high_watermark (i {i})')

class TestDrawdownsContinuousRuns(unittest.TestCase):
    def test_matches_bacon_2023_output(self):
        # The book's four-decimal drawdown values are approximate; the core
        # run tracker has separate exact rolling-window tests.
        actual = run_stream_method("drawdowns_continuous_runs",
            returns=bacon_2023_portfolio_returns, benchmark_returns=bacon_2023_portfolio_returns)
        actual = actual[len(actual) - 1]
        expected = bacon_2023_drawdown_continuous_without_zeroes
        assertSeriesEqual(self, actual, expected, delta=0.002,
            prefix=f'drawdowns continuous runs (bacon 2023)')

class TestCalmarRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.calmar_ratio.EXPECTED_VALUES
        actual = run_stream_property("calmar_ratio")
        assertSeriesEqual(self, actual, expected, places=13,
                              prefix="calmar ratio (yearly)")
        actual = run_stream_property("calmar_ratio", daily=True)
        assertSeriesEqual(self, actual, expected, places=13,
                              prefix="calmar ratio (daily)")

class TestSterlingRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for excess, expected in rd.sterling_ratio.EXPECTED_VALUES_BY_EXCESS.items():
            actual = run_stream_method("sterling_ratio", excess=excess)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'sterling ratio (yearly, excess {excess})')
        for excess, expected in rd.sterling_ratio.EXPECTED_VALUES_BY_EXCESS.items():
            actual = run_stream_method("sterling_ratio", excess=excess, daily=True)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'sterling ratio (daily, excess {excess})')

class TestBurkeRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Generated reference data is yearly.
        """
        for rf, expected in rd.burke_ratio.EXPECTED_VALUES_BY_RF.items():
            actual = run_stream_property("burke_ratio",
                                         annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=11,
                              prefix=f'burke ratio (yearly, Rf {rf})')

class TestBurkeRatioModified(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        Generated reference data is yearly.
        """
        for rf, expected in rd.burke_ratio.EXPECTED_VALUES_BY_RF_MODIFIED.items():
            actual = run_stream_property("burke_ratio_modified",
                                         annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=11,
                              prefix=f'burke ratio modified (Rf {rf})')

class TestPainIndex(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        # R's drawdown series differs slightly from the high-water-mark
        # series used here. TestDocumentedFormulas checks the exact formula.
        expected = rd.pain_index.EXPECTED_VALUES
        actual = run_stream_property("pain_index")
        assertSeriesEqual(self, actual, expected, delta=0.00098,
                          prefix="pain index (yearly)")
        actual = run_stream_property("pain_index", daily=True)
        assertSeriesEqual(self, actual, expected, delta=0.00098,
                          prefix="pain index (daily)")

class TestPainRatio(unittest.TestCase):
    """Keep the R comparison; check the documented ratio separately below."""

    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.pain_ratio.EXPECTED_VALUES_BY_RF.items():
            actual = run_stream_property("pain_ratio",
                                         annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, delta=0.016 if rf < 0.04 else 0.111,
                              prefix=f'pain ratio (yearly, Rf {rf})')
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("pain_ratio", daily=True,
                                         annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, delta=0.016 if rf < 0.04 else 0.111,
                              prefix=f'pain ratio (daily, Rf {rf})')

class TestUlcerIndex(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        # The drawdown convention differs from the R fixture; the precise
        # root-mean-square definition is checked in TestDocumentedFormulas.
        expected = rd.ulcer_index.EXPECTED_VALUES
        actual = run_stream_property("ulcer_index")
        assertSeriesEqual(self, actual, expected, delta=0.00192,
                          prefix="ulcer index (yearly)")
        actual = run_stream_property("ulcer_index", daily=True)
        assertSeriesEqual(self, actual, expected, delta=0.00192,
                          prefix="ulcer index (daily)")

class TestMartinRatio(unittest.TestCase):
    """Keep the R comparison; check the documented ratio separately below."""

    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.martin_ratio.EXPECTED_VALUES_BY_RF.items():
            actual = run_stream_property("martin_ratio",
                                         annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, delta=0.0630,
                              prefix=f'martin ratio (yearly, Rf {rf})')
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("martin_ratio", daily=True,
                                         annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, delta=0.0630,
                              prefix=f'martin ratio (daily, Rf {rf})')

class TestDrawdownAverage(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.drawdown_average.EXPECTED_VALUES
        actual = run_stream_property("drawdown_average")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix=f'drawdown average (yearly)')
        actual = run_stream_property("drawdown_average", daily=True)
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix=f'drawdown average (daily)')

class TestDrawdownAverageLength(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.drawdown_average_length.EXPECTED_VALUES
        actual = run_stream_property("drawdown_average_length")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix=f'drawdown average length (yearly)')
        actual = run_stream_property("drawdown_average_length", daily=True)
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix=f'drawdown average length (daily)')

class TestDrawdownAveragePeakToTrough(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.drawdown_average_peak_to_trough.EXPECTED_VALUES
        actual = run_stream_property("drawdown_average_peak_to_trough")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix=f'drawdown average peak-to-trough (yearly)')
        actual = run_stream_property("drawdown_average_peak_to_trough", daily=True)
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix=f'drawdown average peak-to-trough (daily)')

class TestDrawdownAverageRecovery(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.drawdown_average_recovery.EXPECTED_VALUES
        actual = run_stream_property("drawdown_average_recovery")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix=f'drawdown average recovery (yearly)')
        actual = run_stream_property("drawdown_average_recovery", daily=True)
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix=f'drawdown average recovery (daily)')

class TestDrawdownDeviation(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.drawdown_deviation.EXPECTED_VALUES
        actual = run_stream_property("drawdown_deviation")
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix=f'drawdown deviation (yearly)')
        actual = run_stream_property("drawdown_deviation", daily=True)
        assertSeriesEqual(self, actual, expected, places=15,
                          prefix=f'drawdown deviation (daily)')

class TestCDaRAverage(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        # R and this implementation select the continuous drawdown tail
        # differently. TestDocumentedFormulas checks our linear quantile.
        for p, expected in rd.cdar.EXPECTED_VALUES_BY_P_AVERAGE_GEOMETRIC_INVERTED.items():
            actual = run_stream_method("cdar_average", confidence=p)
            assertSeriesEqual(self, actual, expected, delta=0.02938,
                              prefix=f'CDaR average geometric (yearly) p {p}')
        for p, expected in rd.cdar.EXPECTED_VALUES_BY_P_AVERAGE_GEOMETRIC_INVERTED.items():
            actual = run_stream_method("cdar_average", confidence=p, daily=True)
            assertSeriesEqual(self, actual, expected, delta=0.02938,
                              prefix=f'CDaR average geometric (daily) p {p}')

class TestCDaRDiscrete(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for p, expected in rd.cdar.EXPECTED_VALUES_BY_P_DISCRETE_GEOMETRIC_INVERTED.items():
            actual = run_stream_method("cdar_discrete", confidence=p)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'CDaR discrete geometric (yearly) p {p}')
        for p, expected in rd.cdar.EXPECTED_VALUES_BY_P_DISCRETE_GEOMETRIC_INVERTED.items():
            actual = run_stream_method("cdar_discrete", confidence=p, daily=True)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'CDaR discrete geometric (daily) p {p}')

class TestCDaRBeta(unittest.TestCase):
    def test_discrete_tail_selection(self):
        m = make_measures()
        for portfolio, benchmark in ((0.1, 0.1), (-0.05, -0.1),
                                     (0.2, 0.2), (-0.1, -0.2),
                                     (0.3, 0.3), (-0.15, -0.3)):
            m.add_return(portfolio, benchmark)
        # At 50% confidence, two of three episodes are selected. The
        # denominator is the second-worst depth, -0.2.
        self.assertAlmostEqual(m.cdar_beta(0.5), (-0.15 - 0.1) / (2 * -0.2),
                               places=14)
        self.assertAlmostEqual(m.cdar_beta(0.8), -0.15 / -0.3, places=14)

    def test_matches_performance_analytics_output(self):
        for p, expected in rd.cdar_beta.EXPECTED_VALUES_BY_P_GEOMETRIC.items():
            actual = run_stream_method("cdar_beta", confidence=p)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'CDaR beta geometric (yearly) p {p}')
        for p, expected in rd.cdar_beta.EXPECTED_VALUES_BY_P_GEOMETRIC.items():
            actual = run_stream_method("cdar_beta", confidence=p, daily=True)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'CDaR beta geometric (daily) p {p}')

    def test_mathematical_properties(self):
        # With one selected episode, identical portfolio and benchmark
        # returns give the same numerator and denominator.
        measures = make_measures()
        for ret in (0.05, -0.1):
            measures.add_return(ret, ret)
        assertFloatEqual(self, measures.cdar_beta(), 1.0, places=15,
                         prefix='CDaR beta (one episode) identity')

        measures = make_measures()
        add_bacon(measures, benchmark_returns=bacon_portfolio_returns)
        assertFloatEqual(self, measures.cdar_beta(), 1.0, places=14,
                         prefix='CDaR beta (Bacon) identity')

        # No drawdowns
        measures_transformed = make_measures()
        add_bacon(measures_transformed, returns=[0.01] * bacon_portfolio_len,
            benchmark_returns=[0.02] * bacon_portfolio_len)
        assertFloatEqual(self, measures_transformed.cdar_beta(), math.nan,
            prefix=f'CDaR beta (geometric) no drawdowns')

        # Zero returns (portfolio ​= 0)
        measures_transformed = make_measures()
        add_bacon(measures_transformed, returns=[0] * bacon_portfolio_len)
        assertFloatEqual(self, measures_transformed.cdar_beta(), 0.0,
            places = 15, prefix=f'CDaR beta (geometric) zero returns')

        for confidence in (0.0, 1.0):
            with self.assertRaises(ValueError):
                measures.cdar_beta(confidence)

class TestCDaRAlpha(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        """
        PerformanceAnalytics has a bug using hardcoded period of 12 when annualizing the means.
        Our implementation uses ``self.periods_per_annum`` instead.
        So we have to use monthly dates to match the reference data.
        """
        for p, expected in rd.cdar_alpha.EXPECTED_VALUES_BY_P_GEOMETRIC.items():
            actual = run_stream_method("cdar_alpha", confidence=p, monthly=True)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'CDaR alpha geometric (monthly) p {p}')

    def test_mathematical_properties(self):
        # Identity (portfolio == benchmark).
        measures = make_measures(monthly=True)
        add_bacon(measures, benchmark_returns=bacon_portfolio_returns)
        assertFloatEqual(self, measures.cdar_alpha(), 0.0, places=14,
                              prefix='CDaR alpha (geometric) identity')

        # No drawdowns
        measures_transformed = make_measures()
        add_bacon(measures_transformed, returns=[0.01] * bacon_portfolio_len,
            benchmark_returns=[0.02] * bacon_portfolio_len)
        assertFloatEqual(self, measures_transformed.cdar_alpha(), math.nan,
            prefix=f'CDaR alpha (geometric) no drawdowns')

        # Zero returns (portfolio ​= 0)
        measures_transformed = make_measures()
        add_bacon(measures_transformed, returns=[0] * bacon_portfolio_len)
        assertFloatEqual(self, measures_transformed.cdar_alpha(), 0.0,
            places = 15, prefix=f'CDaR alpha (geometric) zero returns')

class TestRewardToConditionalDrawdown(unittest.TestCase):
    def test_definition(self):
        """
        Geometric mean return divided by the mean magnitude of the worst
        max(1, int(n * (1 - confidence))) drawdowns.
        """
        for confidence in (0.8, 0.95):
            m = make_measures()
            for i in range(bacon_portfolio_len):
                m.add_return(bacon_portfolio_returns[i], bacon_benchmark_returns[i])
                dd = sorted(m.drawdowns_high_watermark)
                tail = dd[:max(1, int(len(dd) * (1 - confidence)))]
                cdar = -sum(tail) / len(tail)
                expected = m.geometric_mean_return / cdar if cdar != 0 else math.nan
                assertFloatEqual(self, m.reward_to_conditional_drawdown(confidence=confidence),
                    expected, places=15, prefix=f'confidence {confidence} step {i}')

class TestSfmRiskPremium(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.sfm_risk_premium.EXPECTED_VALUES_BY_RF_PERFAN.items():
            actual = run_stream_property("sfm_risk_premium",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'SFM risk premium (yearly, Rf {rf})')
        for rf, expected in rd.sfm_risk_premium.EXPECTED_VALUES_BY_RF_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("sfm_risk_premium", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'SFM risk premium (daily, Rf {rf})')

class TestSfmAlpha(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.sfm_alpha.EXPECTED_VALUES_BY_RF_PERFAN.items():
            actual = run_stream_property("sfm_alpha",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'SFM alpha (yearly, Rf {rf})')
        for rf, expected in rd.sfm_alpha.EXPECTED_VALUES_BY_RF_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("sfm_alpha", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'SFM alpha (daily, Rf {rf})')

class TestSfmBeta(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.sfm_beta.EXPECTED_VALUES_BY_RF_PERFAN.items():
            actual = run_stream_property("sfm_beta",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'SFM beta (yearly, Rf {rf})')
        for rf, expected in rd.sfm_beta.EXPECTED_VALUES_BY_RF_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("sfm_beta", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'SFM beta (daily, Rf {rf})')

class TestSfmBetaBull(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.sfm_beta_bull.EXPECTED_VALUES_BY_RF_PERFAN.items():
            actual = run_stream_property("sfm_beta_bull",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'SFM beta bull (yearly, Rf {rf})')
        for rf, expected in rd.sfm_beta_bull.EXPECTED_VALUES_BY_RF_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("sfm_beta_bull", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'SFM beta bull (daily, Rf {rf})')

class TestSfmBetaBear(unittest.TestCase):
    def test_matches_reference_implementation_output(self):
        for rf, expected in rd.sfm_beta_bear.EXPECTED_VALUES_BY_RF_REFERENCE.items():
            actual = run_stream_property("sfm_beta_bear",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'SFM beta bear (yearly, Rf {rf})')
        for rf, expected in rd.sfm_beta_bear.EXPECTED_VALUES_BY_RF_REFERENCE.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("sfm_beta_bear", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'SFM beta bear (daily, Rf {rf})')

class TestTimingRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.timing_ratio.EXPECTED_VALUES_BY_RF_PERFAN.items():
            actual = run_stream_property("timing_ratio",
                                         annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'timing ratio (yearly, Rf {rf})')
        for rf, expected in rd.timing_ratio.EXPECTED_VALUES_BY_RF_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("timing_ratio", daily=True,
                                         annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'timing ratio (daily, Rf {rf})')

class TestSfmR2(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.sfm_r2.EXPECTED_VALUES_BY_RF_PERFAN.items():
            actual = run_stream_property("sfm_r2",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=14, skip=15 if rf < 0.05 else 18,
                              prefix=f'SFM R^2 (yearly, Rf {rf})')
        for rf, expected in rd.sfm_r2.EXPECTED_VALUES_BY_RF_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("sfm_r2", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=14, skip=15 if rf < 0.05 else 18,
                              prefix=f'SFM R^2 (daily, Rf {rf})')

class TestJensenAlpha(unittest.TestCase):
    def test_high_daily_risk_free_rate_definition(self):
        # The R fixtures lose precision after compounding a 10% or 30%
        # periodic rate over 252 periods. Check the documented formula.
        for rf in (0.1, 0.3):
            annual_rf = (1 + rf) ** 252 - 1
            m = Measures(periods_per_annum=252, annual_risk_free_rate=annual_rf)
            for i, (ret, bench) in enumerate(zip(bacon_portfolio_returns,
                                                  bacon_benchmark_returns)):
                m.add_return(ret, bench)
                if i == 0:
                    continue
                p_ann = math.prod(1 + x for x in bacon_portfolio_returns[:i + 1]) \
                    ** (252 / (i + 1)) - 1
                b_ann = math.prod(1 + x for x in bacon_benchmark_returns[:i + 1]) \
                    ** (252 / (i + 1)) - 1
                expected = p_ann - (m.sfm_beta * b_ann +
                                    (1 - m.sfm_beta) * annual_rf)
                assertFloatEqual(self, m.jensen_alpha, expected,
                                 delta=1e-9, rel_tol=1e-12,
                                 prefix=f'Jensen alpha rf={rf} step={i}')
                if m.sfm_beta != 0:
                    assertFloatEqual(self, m.jensen_alpha_modified,
                                     expected / m.sfm_beta,
                                     delta=1e-9, rel_tol=1e-12,
                                     prefix=f'Jensen alpha modified rf={rf} step={i}')

    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.jensen_alpha.EXPECTED_VALUES_BY_RF_DAILY_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("jensen_alpha", daily=True,
                 annual_risk_free_rate=annual_rf)
            if rf <= 0.05:
                # At higher periodic rates the R reference loses precision
                # through cancellation; the formula test above covers them.
                assertSeriesEqual(self, actual, expected, delta=9e-10, rel_tol=1e-10,
                        prefix=f'Jensen alpha (daily, Rf {rf})')
        for rf, expected in rd.jensen_alpha.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("jensen_alpha", monthly=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=12,
                prefix=f'Jensen alpha (monthly, Rf {rf})')
        for rf, expected in rd.jensen_alpha.EXPECTED_VALUES_BY_RF_YEARLY_PERFAN.items():
            actual = run_stream_property("jensen_alpha",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=14,
                prefix=f'Jensen alpha (yearly, Rf {rf})')

class TestFamaBeta(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.fama_beta.EXPECTED_VALUES_PERFAN
        actual = run_stream_property("fama_beta")
        assertSeriesEqual(self, actual, expected, places=14,
                        prefix=f'fama beta (yearly)')
        actual = run_stream_property("fama_beta", daily=True)
        assertSeriesEqual(self, actual, expected, places=14,
                        prefix=f'fama beta (daily)')

class TestModigliani(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.modigliani.EXPECTED_VALUES_BY_RF_PERFAN.items():
            actual = run_stream_property("modigliani",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'Modigliani-Modigliani (yearly, Rf {rf})')
        for rf, expected in rd.modigliani.EXPECTED_VALUES_BY_RF_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("modigliani", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'Modigliani-Modigliani (daily, Rf {rf})')

class TestTrackingError(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.tracking_error.EXPECTED_VALUES_DAILY_PERFAN
        actual = run_stream_property("tracking_error", daily=True)
        assertSeriesEqual(self, actual, expected, places=15,
                        prefix="tracking error (daily)")
        expected = rd.tracking_error.EXPECTED_VALUES_MONTHLY_PERFAN
        actual = run_stream_property("tracking_error", monthly=True)
        assertSeriesEqual(self, actual, expected, places=15,
                        prefix="tracking error (monthly)")
        expected = rd.tracking_error.EXPECTED_VALUES_ANNUAL_PERFAN
        actual = run_stream_property("tracking_error")
        assertSeriesEqual(self, actual, expected, places=15,
                        prefix="tracking error (yearly)")

class TestActivePremium(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.active_premium.EXPECTED_VALUES_DAILY_PERFAN
        actual = run_stream_property("active_premium", daily=True)
        # Note we skip first element because PerformanceAnalytics code
        # uses it to determine periodicity.
        assertSeriesEqual(self, actual, expected, places=11, skip=1,
                        prefix="active premium (daily)")
        expected = rd.active_premium.EXPECTED_VALUES_MONTHLY_PERFAN
        actual = run_stream_property("active_premium", monthly=True)
        assertSeriesEqual(self, actual, expected, places=14, skip=1,
                        prefix="active premium (monthly)")
        expected = rd.active_premium.EXPECTED_VALUES_ANNUAL_PERFAN
        actual = run_stream_property("active_premium")
        assertSeriesEqual(self, actual, expected, places=15, skip=1,
                        prefix="active premium (yearly)")

class TestInformationRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.information_ratio.EXPECTED_VALUES_DAILY_PERFAN
        actual = run_stream_property("information_ratio", daily=True)
        assertSeriesEqual(self, actual, expected, places=10, skip=2,
                        prefix="information ratio (daily)")
        expected = rd.information_ratio.EXPECTED_VALUES_MONTHLY_PERFAN
        actual = run_stream_property("information_ratio", monthly=True)
        assertSeriesEqual(self, actual, expected, places=12, skip=2,
                        prefix="information ratio (monthly)")
        expected = rd.information_ratio.EXPECTED_VALUES_ANNUAL_PERFAN
        actual = run_stream_property("information_ratio")
        assertSeriesEqual(self, actual, expected, places=13, skip=2,
                        prefix="information ratio (yearly)")

class TestInformationRatioModified(unittest.TestCase):
    def test_sign_rule(self):
        """
        Equals information_ratio when the geometric active premium is positive,
        otherwise its negation.
        """
        m = make_measures(monthly=True)
        for i in range(bacon_portfolio_len):
            m.add_return(bacon_portfolio_returns[i], bacon_benchmark_returns[i])
            active = m.active_premium
            ir = m.information_ratio
            expected = math.nan if math.isnan(ir) else (ir if active > 0 else -ir)
            assertFloatEqual(self, m.information_ratio_modified, expected, places=15,
                             prefix=f'step {i}')

class TestSystematicRisk(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.systematic_risk.EXPECTED_VALUES_BY_RF_DAILY_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("systematic_risk", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'systematic risk (daily, Rf {rf})')
        for rf, expected in rd.systematic_risk.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("systematic_risk", monthly=True,
                annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'systematic risk (monthly, Rf {rf})')
        for rf, expected in rd.systematic_risk.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN.items():
            actual = run_stream_property("systematic_risk",
                annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'systematic risk (yearly, Rf {rf})')

class TestTreynorRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.treynor_ratio.EXPECTED_VALUES_BY_RF_DAILY_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("treynor_ratio", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=10,
                              prefix=f'treynor ratio (daily, Rf {rf})')
        for rf, expected in rd.treynor_ratio.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("treynor_ratio", monthly=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=13,
                            prefix=f'treynor ratio (monthly, Rf {rf})')
        for rf, expected in rd.treynor_ratio.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN.items():
            actual = run_stream_property("treynor_ratio",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=14,
                            prefix=f'treynor ratio (yearly, Rf {rf})')

class TestTreynorRatioModified(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.treynor_ratio_modified.EXPECTED_VALUES_BY_RF_DAILY_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("treynor_ratio_modified", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=10,
                              prefix=f'treynor ratio modified (daily, Rf {rf})')
        for rf, expected in rd.treynor_ratio_modified.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("treynor_ratio_modified", monthly=True,
                annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=12,
                            prefix=f'treynor ratio modified (monthly, Rf {rf})')
        for rf, expected in rd.treynor_ratio_modified.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN.items():
            actual = run_stream_property("treynor_ratio_modified",
                annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=12,
                            prefix=f'treynor ratio modified (yearly, Rf {rf})')

class TestSpecificRisk(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.specific_risk.EXPECTED_VALUES_BY_RF_DAILY_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("specific_risk", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'specific risk (daily, Rf {rf})')
        for rf, expected in rd.specific_risk.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("specific_risk", monthly=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'specific risk (monthly, Rf {rf})')
        for rf, expected in rd.specific_risk.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN.items():
            actual = run_stream_property("specific_risk",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'specific risk (yearly, Rf {rf})')

class TestTotalRisk(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.total_risk.EXPECTED_VALUES_BY_RF_DAILY_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("total_risk", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'total_risk (daily, Rf {rf})')
        for rf, expected in rd.total_risk.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("total_risk", monthly=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'total_risk (monthly, Rf {rf})')
        for rf, expected in rd.total_risk.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN.items():
            actual = run_stream_property("total_risk",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'total_risk (yearly, Rf {rf})')

class TestAppraisalRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.appraisal_ratio.EXPECTED_VALUES_BY_RF_DAILY_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("appraisal_ratio", daily=True,
                 annual_risk_free_rate=annual_rf)
            if rf < 0.1:
                assertSeriesEqual(self, actual, expected, delta=1e-8 if rf < 0.05 else 1e-4, skip=2,
                              prefix=f'appraisal ratio (daily, Rf {rf})')
        for rf, expected in rd.appraisal_ratio.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("appraisal_ratio", monthly=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, delta=1e-11 if rf < 0.05 else 1e-9, skip=2,
                              prefix=f'appraisal ratio (monthly, Rf {rf})')
        for rf, expected in rd.appraisal_ratio.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN.items():
            actual = run_stream_property("appraisal_ratio",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, delta=1e-13 if rf < 0.05 else 1e-11, skip=2,
                              prefix=f'appraisal ratio (yearly, Rf {rf})')

class TestJensenAlphaModified(unittest.TestCase):
    def test_definition(self):
        m = make_measures()
        for i, (ret, bench) in enumerate(zip(bacon_portfolio_returns,
                                              bacon_benchmark_returns)):
            m.add_return(ret, bench)
            beta = m.sfm_beta
            expected = m.jensen_alpha / beta if beta != 0 else math.nan
            assertFloatEqual(self, m.jensen_alpha_modified, expected,
                             places=14, prefix=f'Jensen alpha modified n={i + 1}')

    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.jensen_alpha_modified.EXPECTED_VALUES_BY_RF_DAILY_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("jensen_alpha_modified", daily=True,
                 annual_risk_free_rate=annual_rf)
            if rf < 0.05:
                # Annualizing 24 daily observations amplifies the reference's
                # floating-point error, even when the formula agrees.
                assertSeriesEqual(self, actual, expected, delta=1e-8,
                              prefix=f'Jensen alpha modified (daily, Rf {rf})')
        for rf, expected in rd.jensen_alpha_modified.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("jensen_alpha_modified", monthly=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, delta=1e-10,
                              prefix=f'Jensen alpha modified (monthly, Rf {rf})')
        for rf, expected in rd.jensen_alpha_modified.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN.items():
            actual = run_stream_property("jensen_alpha_modified",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, delta=1e-13,
                              prefix=f'Jensen alpha modified (yearly, Rf {rf})')

class TestJensenAlphaAlternative(unittest.TestCase):
    """Compare R and separately check the documented systematic-risk ratio."""

    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.jensen_alpha_alternative.EXPECTED_VALUES_BY_RF_DAILY_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("jensen_alpha_alternative", daily=True,
                 annual_risk_free_rate=annual_rf)
            if rf < 0.3:
                assertSeriesEqual(self, actual, expected, delta=0.1826 if rf < 0.1 else 0.707,
                              prefix=f'Jensen alpha alternative (daily, Rf {rf})')
        for rf, expected in rd.jensen_alpha_alternative.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("jensen_alpha_alternative", monthly=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=10 if rf < 0.3 else 9,
                              prefix=f'Jensen alpha alternative (monthly, Rf {rf})')
        for rf, expected in rd.jensen_alpha_alternative.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN.items():
            actual = run_stream_property("jensen_alpha_alternative",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=12,
                              prefix=f'Jensen alpha alternative (yearly, Rf {rf})')

class TestMSquared(unittest.TestCase):
    """Compare R and separately check the documented geometric formula."""

    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.m_squared.EXPECTED_VALUES_BY_RF_DAILY_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("m_squared", daily=True,
                 annual_risk_free_rate=annual_rf)
            if rf < 0.05:
                assertSeriesEqual(self, actual, expected, delta=0.1849 if rf < 0.01 else 0.82956,
                              prefix=f'M squared (daily, Rf {rf})')
        for rf, expected in rd.m_squared.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("m_squared", monthly=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, delta = 0.00861 if rf < 0.05 else 1.621,
                              prefix=f'M squared (monthly, Rf {rf})')
        for rf, expected in rd.m_squared.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN.items():
            actual = run_stream_property("m_squared",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'M squared (yearly, Rf {rf})')

class TestMSquaredExcess(unittest.TestCase):
    """Compare R and separately check the documented geometric formula."""

    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.m_squared_excess.EXPECTED_VALUES_BY_RF_DAILY_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("m_squared_excess", daily=True,
                 annual_risk_free_rate=annual_rf)
            if rf < 0.05:
                assertSeriesEqual(self, actual, expected, delta=0.02244 if rf < 0.01 else 0.101,
                              prefix=f'M squared excess (daily, Rf {rf})')
        for rf, expected in rd.m_squared_excess.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("m_squared_excess", monthly=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, delta=0.007782 if rf < 0.05 else 1.466,
                              prefix=f'M squared excess (monthly, Rf {rf})')
        for rf, expected in rd.m_squared_excess.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN.items():
            actual = run_stream_property("m_squared_excess",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'M squared excess (yearly, Rf {rf})')

class TestMSquaredSortino(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for mar, expected in rd.m_squared_sortino.EXPECTED_VALUES_BY_MAR_DAILY_PERFAN.items():
            annual_mar = (1 + mar) ** 252 - 1
            actual = run_stream_property("m_squared_sortino", daily=True,
                 annual_target_return=annual_mar)
            assertSeriesEqual(self, actual, expected, places=11, skip=3,
                              prefix=f'M squared Sortino (daily, MAR {mar})')
        for mar, expected in rd.m_squared_sortino.EXPECTED_VALUES_BY_MAR_MONTHLY_PERFAN.items():
            annual_mar = (1 + mar) ** 12 - 1
            actual = run_stream_property("m_squared_sortino", monthly=True,
                 annual_target_return=annual_mar)
            assertSeriesEqual(self, actual, expected, places=14, skip=3,
                              prefix=f'M squared Sortino (monthly, MAR {mar})')
        for mar, expected in rd.m_squared_sortino.EXPECTED_VALUES_BY_MAR_ANNUAL_PERFAN.items():
            actual = run_stream_property("m_squared_sortino",
                 annual_target_return=mar)
            assertSeriesEqual(self, actual, expected, places=15, skip=3,
                              prefix=f'M squared Sortino (yearly, MAR {mar})')

class TestTailRatio(unittest.TestCase):
    def test_matches_reference_implementation_output(self):
        for cutoff, expected in rd.tail_ratio.EXPECTED_VALUES_BY_CUTOFF_REFERENCE.items():
            actual = run_stream_method("tail_ratio", cutoff=cutoff)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'tail ratio (yearly, cutoff {cutoff})')

class TestKellyRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.kelly_ratio.EXPECTED_VALUES_BY_RF_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("kelly_ratio", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=11,
                              prefix=f'Kelly ratio (daily, Rf {rf})')
        for rf, expected in rd.kelly_ratio.EXPECTED_VALUES_BY_RF_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("kelly_ratio", monthly=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=11,
                              prefix=f'Kelly ratio (monthly, Rf {rf})')
        for rf, expected in rd.kelly_ratio.EXPECTED_VALUES_BY_RF_PERFAN.items():
            actual = run_stream_property("kelly_ratio",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=11,
                              prefix=f'Kelly ratio (yearly, Rf {rf})')

class TestKellyRatioFull(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for rf, expected in rd.kelly_ratio.EXPECTED_VALUES_BY_RF_FULL_PERFAN.items():
            annual_rf = (1 + rf) ** 252 - 1
            actual = run_stream_property("kelly_ratio_full", daily=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=11,
                              prefix=f'Kelly ratio full (daily, Rf {rf})')
        for rf, expected in rd.kelly_ratio.EXPECTED_VALUES_BY_RF_FULL_PERFAN.items():
            annual_rf = (1 + rf) ** 12 - 1
            actual = run_stream_property("kelly_ratio_full", monthly=True,
                 annual_risk_free_rate=annual_rf)
            assertSeriesEqual(self, actual, expected, places=11,
                              prefix=f'Kelly ratio full (monthly, Rf {rf})')
        for rf, expected in rd.kelly_ratio.EXPECTED_VALUES_BY_RF_FULL_PERFAN.items():
            actual = run_stream_property("kelly_ratio_full",
                 annual_risk_free_rate=rf)
            assertSeriesEqual(self, actual, expected, places=11,
                              prefix=f'Kelly ratio full (yearly, Rf {rf})')

class TestHurstExponent(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.hurst_exponent.EXPECTED_VALUES_PERFAN
        actual = run_stream_property("hurst_exponent")
        assertSeriesEqual(self, actual, expected, places=14, prefix=f'Hurst exponent')

class TestBiasRatio(unittest.TestCase):
    def test_matches_reference_implementation_output(self):
        for mult, expected in rd.bias_ratio.EXPECTED_VALUES_BY_MULT_REFERENCE.items():
            actual = run_stream_method("bias_ratio", std_dev_multiplier=mult)
            assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'bias ratio (yearly, std_dev_multiplier {mult})')

class TestKRatio(unittest.TestCase):
    def test_matches_reference_implementation_output(self):
        expected = rd.k_ratio.EXPECTED_VALUES_REFERENCE
        actual = run_stream_property("k_ratio")
        assertSeriesEqual(self, actual, expected, places=14, prefix=f'K-ratio')

class TestGainToPainRatio(unittest.TestCase):
    def test_matches_reference_implementation_output(self):
        expected = rd.gain_to_pain_ratio.EXPECTED_VALUES_REFERENCE
        actual = run_stream_property("gain_to_pain_ratio")
        assertSeriesEqual(self, actual, expected, places=15, prefix=f'Gain-to-pain ratio')

class TestUpsideCaptureRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for geom, expected in rd.upside_capture_ratio.EXPECTED_VALUES_BY_GEOMETRIC_PERFAN.items():
            actual = run_stream_method("upside_capture_ratio", geometric=geom)
            assertSeriesEqual(self, actual, expected, places=13, skip=1,
                              prefix=f'Upside capture ratio (yearly, geometric {geom})')

class TestDownsideCaptureRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        for geom, expected in rd.downside_capture_ratio.EXPECTED_VALUES_BY_GEOMETRIC_PERFAN.items():
            actual = run_stream_method("downside_capture_ratio", geometric=geom)
            assertSeriesEqual(self, actual, expected, places=14,
                              prefix=f'Downside capture ratio (yearly, geometric {geom})')

class TestOverallCaptureRatio(unittest.TestCase):
    def test_matches_reference_implementation_output(self):
        for geom, expected in rd.overall_capture_ratio.EXPECTED_VALUES_BY_GEOMETRIC_REFERENCE.items():
            actual = run_stream_method("overall_capture_ratio", geometric=geom)
            assertSeriesEqual(self, actual, expected, places=13,
                              prefix=f'Overall capture ratio (yearly, geometric {geom})')

class TestUpNumberRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.up_number_ratio.EXPECTED_VALUES_PERFAN
        actual = run_stream_property("up_number_ratio")
        assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'Up number ratio (yearly)')

class TestDownNumberRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.down_number_ratio.EXPECTED_VALUES_PERFAN
        actual = run_stream_property("down_number_ratio")
        assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'Down number ratio (yearly)')

class TestUpPercentageRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.up_percentage_ratio.EXPECTED_VALUES_PERFAN
        actual = run_stream_property("up_percentage_ratio")
        assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'Up percentage ratio (yearly)')

class TestDownPercentageRatio(unittest.TestCase):
    def test_matches_performance_analytics_output(self):
        expected = rd.down_percentage_ratio.EXPECTED_VALUES_PERFAN
        actual = run_stream_property("down_percentage_ratio")
        assertSeriesEqual(self, actual, expected, places=15,
                              prefix=f'Down percentage ratio (yearly)')

class TestDocumentedFormulas(unittest.TestCase):
    """Formula checks alongside comparisons with the R fixtures."""

    def test_drawdown_risk_and_ratios(self):
        for rf in (0.0, 0.05):
            m = make_measures(annual_rf=rf)
            for i, (ret, bench) in enumerate(zip(bacon_portfolio_returns,
                                                  bacon_benchmark_returns)):
                m.add_return(ret, bench)
                drawdowns = m.drawdowns_high_watermark
                n = i + 1
                pain = -math.fsum(drawdowns) / n
                ulcer = math.sqrt(math.fsum(x * x for x in drawdowns) / n)
                geometric_return = math.prod(1 + x for x in bacon_portfolio_returns[:n]) \
                    ** (1 / n) - 1
                self.assertAlmostEqual(m.pain_index, pain, places=14)
                self.assertAlmostEqual(m.ulcer_index, ulcer, places=14)
                if pain > 0:
                    self.assertAlmostEqual(m.pain_ratio,
                                           (geometric_return - rf) / pain, places=12)
                if ulcer > 0:
                    self.assertAlmostEqual(m.martin_ratio,
                                           (geometric_return - rf) / ulcer, places=12)

    def test_cdar_average_and_alpha(self):
        m = make_measures(monthly=True)
        for ret, bench in zip(bacon_portfolio_returns, bacon_benchmark_returns):
            m.add_return(ret, bench)
            drawdowns = sorted(m.drawdowns_high_watermark)
            for confidence in (0.9, 0.95):
                position = (1 - confidence) * (len(drawdowns) - 1)
                lo = int(position)
                q = drawdowns[lo] + (position - lo) * (
                    drawdowns[min(lo + 1, len(drawdowns) - 1)] - drawdowns[lo])
                tail = [d for d in drawdowns if d <= q]
                expected_cdar = -math.fsum(tail) / len(tail) if q < 0 else 0.0
                self.assertAlmostEqual(m.cdar_average(confidence), expected_cdar,
                                       places=14)
                beta = m.cdar_beta(confidence)
                if math.isfinite(beta):
                    portfolio_mean = statistics.fmean(bacon_portfolio_returns[:len(drawdowns)])
                    benchmark_mean = statistics.fmean(bacon_benchmark_returns[:len(drawdowns)])
                    expected_alpha = (1 + portfolio_mean) ** 12 - 1 \
                        - beta * ((1 + benchmark_mean) ** 12 - 1)
                    self.assertAlmostEqual(m.cdar_alpha(confidence), expected_alpha,
                                           places=13)

    def test_m_squared_and_jensen_alpha_alternative(self):
        for periods, periodic_rf in ((1, 0.05), (12, 0.05), (252, 0.01)):
            annual_rf = (1 + periodic_rf) ** periods - 1
            m = Measures(periods_per_annum=periods,
                         annual_risk_free_rate=annual_rf)
            for i, (ret, bench) in enumerate(zip(bacon_portfolio_returns,
                                                  bacon_benchmark_returns)):
                m.add_return(ret, bench)
                n = i + 1
                if n < 2:
                    continue
                portfolio = bacon_portfolio_returns[:n]
                benchmark = bacon_benchmark_returns[:n]
                p_ann = math.prod(1 + x for x in portfolio) ** (periods / n) - 1
                b_ann = math.prod(1 + x for x in benchmark) ** (periods / n) - 1
                p_std = statistics.pstdev(portfolio)
                b_std = statistics.pstdev(benchmark)
                scale = b_std / p_std
                expected_m2 = p_ann * scale + annual_rf * (1 - scale)
                assertFloatEqual(self, m.m_squared, expected_m2,
                                 delta=1e-10, rel_tol=1e-11,
                                 prefix=f'M squared periods={periods} n={n}')
                expected_excess = (1 + expected_m2) / (1 + b_ann) - 1
                assertFloatEqual(self, m.m_squared_excess, expected_excess,
                                 delta=1e-10, rel_tol=1e-11,
                                 prefix=f'M squared excess periods={periods} n={n}')
                systematic_risk = abs(m.sfm_beta) * statistics.stdev(benchmark) \
                    * math.sqrt(periods)
                if systematic_risk > 0 and math.isfinite(systematic_risk):
                    assertFloatEqual(self, m.jensen_alpha_alternative,
                                     m.jensen_alpha / systematic_risk,
                                     delta=1e-10, rel_tol=1e-11,
                                     prefix=f'Jensen alpha alternative periods={periods} n={n}')

    def test_m_squared_equal_volatility_with_extreme_rate(self):
        m = Measures(periods_per_annum=252,
                     annual_risk_free_rate=1.3 ** 252 - 1)
        for portfolio, benchmark in ((0.125, 0.25), (0.375, 0.5)):
            m.add_return(portfolio, benchmark)
        # Binary-exact inputs give exactly equal portfolio and benchmark
        # volatility, so the annual risk-free terms must cancel.
        expected = ((1.125 * 1.375) ** 126) - 1
        self.assertTrue(math.isclose(m.m_squared, expected, rel_tol=1e-14))


class TestRollingWindow(unittest.TestCase):
    """
    Rolling window should produce the same results as a fresh instance
    fed only the last N returns.
    """
    def test_rolling_matches_fresh(self):
        """
        At every step, including while the window is still filling, every
        public property and every method (with default arguments) of a
        rolling-window instance equals that of a fresh instance fed only
        the returns in the window.
        """
        rng = random.Random(42)
        random_returns = [rng.gauss(0.002, 0.03) for _ in range(150)]
        random_benchmark = [rng.gauss(0.001, 0.025) for _ in range(150)]
        configs = [
            dict(window=10, periods_per_annum=1, annual_rf=0.0, annual_mar=0.0,
                 returns=bacon_portfolio_returns, benchmark=bacon_benchmark_returns),
            dict(window=30, periods_per_annum=12, annual_rf=0.05, annual_mar=0.03,
                 returns=random_returns, benchmark=random_benchmark),
        ]
        names = public_measures()
        for cfg in configs:
            window = cfg['window']
            def create():
                return Measures(periods_per_annum=cfg['periods_per_annum'],
                    annual_risk_free_rate=cfg['annual_rf'],
                    annual_target_return=cfg['annual_mar'])
            rolling = Measures(periods_per_annum=cfg['periods_per_annum'],
                annual_risk_free_rate=cfg['annual_rf'],
                annual_target_return=cfg['annual_mar'],
                rolling_window_size=window)
            returns, benchmark = cfg['returns'], cfg['benchmark']
            for i in range(len(returns)):
                rolling.add_return(returns[i], benchmark[i])
                fresh = create()
                for j in range(max(0, i - window + 1), i + 1):
                    fresh.add_return(returns[j], benchmark[j])
                for name in names:
                    actual, expected = evaluate(rolling, name), evaluate(fresh, name)
                    prefix = f'window {window} step {i} {name}'
                    if isinstance(expected, list):
                        assertSeriesEqual(self, actual, expected, places=12, prefix=prefix)
                        self.assertEqual(len(actual), len(expected), msg=prefix)
                    elif isinstance(expected, bool):
                        self.assertEqual(actual, expected, msg=prefix)
                    else:
                        delta = 1e-12 * max(1.0, abs(expected)) if math.isfinite(expected) else None
                        assertFloatEqual(self, actual, expected, delta=delta, prefix=prefix)

class TestGenerateReferenceOutput(unittest.TestCase):
    def foo_test_generate_reference_output(self):
        def f(z:float):
            return "math.nan" if math.isnan(z) else z

        print("EXPECTED_VALUES_REFERENCE = {")

        for x in (0.01, 1):
            act = run_stream_property("gain_to_pain_ratio")
            print(f'    {x}: [')
            print(f'    {f(act[0])}, {f(act[1])}, {f(act[2])},')
            print(f'    {f(act[3])}, {f(act[4])}, {f(act[5])},')
            print(f'    {f(act[6])}, {f(act[7])}, {f(act[8])},')
            print(f'    {f(act[9])}, {f(act[10])}, {f(act[11])},')
            print(f'    {f(act[12])}, {f(act[13])}, {f(act[14])},')
            print(f'    {f(act[15])}, {f(act[16])}, {f(act[17])},')
            print(f'    {f(act[18])}, {f(act[19])}, {f(act[20])},')
            print(f'    {f(act[21])}, {f(act[22])}, {f(act[23])}],')
        print("}")

if __name__ == '__main__':
    unittest.main()
