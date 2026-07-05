# Risk/Reward Ratios

In the ./external/performance/ folder I collected downloaded sources inspireing me to write code in py/performance/ .
These are:

- directory: ./external/performance/stuart/ , short name `stuart`
  The article and code [Measures of Risk-adjusted Return](https://web.archive.org/web/20260209060500/http://www.turingfinance.com/computational-investing-with-python-week-one/) from Stuart Gordon Reid's [turingfinance](https://web.archive.org/web/20251113012344/http://www.turingfinance.com/) blog. He also has a [Github gist](https://gist.github.com/StuartGordonReid/67a1ec4fbc8a84c0e856) with another version of his code.
  The markdown version of the article is measures-of-risk-adjusted-return.md and the 3 variants of code are:
  - ./external/performance/stuart/stuart.py (copied by hand from the article)
  - ./external/performance/stuart/RiskAdjustedMetrics.py (code form Github gist)
  I was fascinated with explanation of the topic. Next, I discovered the source from the next bullet.
- directory: ./external/performance/ranaroussi/ , short name `ranaroussi`
  The [quantstats](https://github.com/ranaroussi/quantstats) Git repo by Ran Aroussi. I found it to implement a different set of measures, mainly targeting the trading strategies, which I was interested in.
  The downloaded files are:
  - ./external/performance/ranaroussi/stats.py (all-in-one code)
  - ./external/performance/ranaroussi/utils.py (utilities)
  Searching for the most authorative source, I found one in the next bullet.
- directory: ./external/performance/braverock/ , short name `braverock`
  The [PerformanceAnalytics](https://github.com/braverock/PerformanceAnalytics) Git repo implements a comprehencive set of measures in R.
  It is the most complete and scientific-looking from all I could find on the internet.
  The code is in ./external/performance/braverock/R/ folder , the documentation is in ./external/performance/braverock/man/ folder.
  Reading comments in the code, I found that very many of them refer to the book shown in the next bullet.
- directory: ./external/performance/bacon3/ , short name `bacon3`
  The startibg point of the downloaded markdown version of the book is ./external/performance/bacon3/toc.md
  > Carl Bacon, *Practical portfolio performance measurement and attribution*,
  > third edition 2023 Wiley
  > ISBN:9781119831945
- directory: ./py/performance/, short name `my`
  - ./py/performance/ratios.py contains my im plementation of some functionality inspired by the sources above
  - ./py/performance/test_performance_ratios.py contains tests with master data created from `braverock` code in R environment (https://www.datacamp.com/datalab/w/28c21593-21e6-47d9-8e72-acebdd3be32c/edit) using readme/performance/performance_analytics.R

Below I collected the list of implemented functions for `stuart`, `ranaroussi`, `braverock` and `my`.

## `stuart`: stuart.py has the following functions

vol(returns) # the standard deviation of returns
beta(returns, market) # the division in Beta is done with the variance of the market and not standard deviation
lpm(returns, threshold, order) # lower partial moment of the returns
hpm(returns, threshold, order) # higher partial moment of the returns
var(returns, alpha) # the historical simulation var of the returns
cvar(returns, alpha) # the conditional VaR of the returns
dd(returns, tau) # draw-down given time period tau
max_dd(returns) # maximum draw-down for any tau in (0, T) where T is the length of the return series
average_dd(returns, periods) # average maximum drawdown over n periods
average_dd_squared(returns, periods) # average maximum drawdown squared over n periods
treynor_ratio(er, returns, market, rf)
sharpe_ratio(er, returns, rf)
information_ratio(returns, benchmark)
modigliani_ratio(er, returns, benchmark, rf)
excess_var(er, returns, rf, alpha)
conditional_sharpe_ratio(er, returns, rf, alpha)
omega_ratio(er, returns, rf, target=0)
sortino_ratio(er, returns, rf, target=0)
kappa_three_ratio(er, returns, rf, target=0)
gain_loss_ratio(returns, target=0)
upside_potential_ratio(returns, target=0)
calmar_ratio(er, returns, rf)
sterling_ration(er, returns, rf, periods)
burke_ratio(er, returns, rf, periods)

## `ranaroussi` stats.py has the following functions

pct_rank(prices, window=60) # percentile rank of prices over a rolling window
compsum(returns) # rolling compounded returns (cumulative product)
comp(returns) # total compounded returns (final cumulative return)
distribution(returns, compounded=True) # return distributions across different time periods
expected_return(returns, aggregate=None, compounded=True) # expected return (geometric mean) for a given period
geometric_mean(returns, aggregate=None, compounded=True) # geometric mean of returns
ghpr(returns, aggregate=None, compounded=True)t # Geometric Holding Period Return.
outliers(returns, quantile=0.95) # return outlier returns above a specified quantile
remove_outliers(returns, quantile=0.95) # Remove outlier returns above a specified quantile
best(returns, aggregate=None, compounded=True) # the best (highest) return for a given period
worst(returns, aggregate=None, compounded=True) # the worst (lowest) return for a given period
consecutive_wins(returns, aggregate=None, compounded=True) # maximum number of consecutive winning periods
consecutive_losses(returns, aggregate=None, compounded=True) # maximum number of consecutive losing periods
exposure(returns) # market exposure time as percentage of periods with non-zero returns
win_rate(returns, aggregate=None, compounded=True) # percentage of profitable periods
avg_return(returns, aggregate=None, compounded=True) # average return per period (excluding zero returns)
avg_win(returns, aggregate=None, compounded=True) # mean of positive returns
avg_loss(returns, aggregate=None, compounded=True) # mean of negative returns
volatility(returns, periods=252, annualize=True) # standard deviation of returns
rolling_volatility(returns, rolling_period=126, periods_per_year=252) # rolling volatility over a specified window
implied_volatility(returns, periods=252, annualize=True) # implied volatility using log returns
autocorr_penalty(returns) # autocorrelation penalty for risk-adjusted metrics
sharpe(returns, rf=0.0, periods=252, annualize=True, smart=False) # Sharpe ratio of excess returns
smart_sharpe(returns, rf=0.0, periods=252, annualize=True) # Sharpe with autocorrelation penalty
rolling_sharpe(returns, rf=0.0, rolling_period=126, annualize=True, periods_per_year=252) # Sharpe ratio over a specified window
sortino(returns, rf=0.0, periods=252, annualize=True, smart=False) # Sortino ratio of excess returns
smart_sortino(returns, rf=0.0, periods=252, annualize=True) # Sortino with autocorrelation penalty
rolling_sortino(returns, rf=0.0, rolling_period=126, annualize=True, periods_per_year=252) # Sortino ratio over a specified window
adjusted_sortino(returns, rf=0.0, periods, annualize=True, smart=False) # Jack Schwager's adjusted Sortino ratio
probabilistic_ratio(series, rf=0.0, base="sharpe", periods=252, annualize=False, smart=False): # probabilistic ratio for a given base metric
probabilistic_sharpe_ratio(series, rf=0.0, periods=252, annualize=False, smart=False) # Probabilistic Sharpe Ratio (PSR)
probabilistic_sortino_ratio(series, rf=0.0, periods=252, annualize=False, smart=False) # Probabilistic Sortino Ratio
probabilistic_adjusted_sortino_ratio(series, rf=0.0, periods=252, annualize=False, smart=False) # Probabilistic Adjusted Sortino Ratio
treynor_ratio(returns, benchmark, periods=252.0, rf=0.0)
omega(returns, rf=0.0, required_return=0.0, periods=252)
gain_to_pain_ratio(returns, rf=0, resolution="D") # Jack Schwager's Gain-to-Pain Ratio (GPR)
cagr(returns, rf=0.0, compounded=True, periods=252) # Compound Annual Growth Rate (CAGR) of excess returns
rar(returns, rf=0.0) # Risk-Adjusted Return (RAR)
skew(returns) # returns' skewness
kurtosis(returns) # returns' kurtosis
calmar(returns, periods=252) # Calmar ratio (CAGR / Maximum Drawdown)
ulcer_index(returns) # Ulcer Index (downside risk measurement)
ulcer_performance_index(returns, rf=0.0) # Ulcer Performance Index (UPI)
serenity_index(returns, rf=0.0) # Serenity Index
risk_of_ruin(returns) # probability of losing all capital
value_at_risk(returns, sigma=1.0, confidence=0.95) # daily Value at Risk (VaR)
conditional_value_at_risk(returns, sigma=1.0, confidence=0.95) # ==expected_shortfall==cvar, Conditional Value at Risk (CVaR), also known as Expected Shortfall
tail_ratio(returns, cutoff=0.95) # ratio between right and left tails
payoff_ratio(returns) # == win_loss_ratio,  average win / average loss
profit_ratio(returns) # win ratio / loss ratio
profit_factor(returns) # total wins / total losses
cpc_index(returns) # Profit Factor * Win Rate * Win-Loss Ratio
common_sense_ratio(returns) # Profit Factor * Tail Ratio
outlier_win_ratio(returns, quantile=0.99) # ratio of the 99th percentile of returns to the mean positive return
outlier_loss_ratio(returns, quantile=0.01) # ratio of the 1st percentile of returns to the mean negative return
recovery_factor(returns, rf=0.0) # total returns / maximum drawdown
risk_return_ratio(returns) # mean return / standard deviation
max_drawdown(prices) # maximum drawdown from peak to trough
to_drawdown_series(returns) # Convert returns series to drawdown series
kelly_criterion(returns) # recommended maximum amount of capital that should be allocated to the given strategy
r_squared(returns, benchmark) # ==r2,  coefficient of determination versus benchmark
information_ratio(returns, benchmark) # Information Ratio
greeks(returns, benchmark, periods=252) # alpha and beta relative to benchmark
rolling_greeks(returns, benchmark, periods=252) # alpha and beta over time
compare(returns, benchmark, aggregate=None, compounded=True, round_vals=None, prepare_returns=True) # Compare returns to benchmark across different time periods
monthly_returns(returns, eoy=True, compounded=True) # monthly returns in a pivot table format
drawdown_details(drawdown) # drawdown statistics for each drawdown period: start date/valley (max) date/end date/days/max drawdown/99th percentile drawdown (excludes outliers)
montecarlo(returns, sims=1000, bust=None, goal=None, seed=None) # simulation by shuffling returns
montecarlo_sharpe(returns, sims=1000, rf=0.0, periods=252, seed=None) # Distribution of Sharpe ratios across Monte Carlo simulations
montecarlo_drawdown(returns, sims=1000, seed=None) # Distribution of maximum drawdowns across Monte Carlo simulations
montecarlo_cagr(returns, sims=1000, seed=None) # Distribution of CAGR across Monte Carlo simulations

## `braverock`: has the following functions

R source file names are in comments, I skipped all `chart.*.R`, `check*.R`, `table*.R` and other utility source files.
If you see `... # R/CalmarRatio.R ...`, then the file is located in `./external/performance/braverock/R/CalmarRatio.R`

ActivePremium(Ra, Rb, scale=NA) # R/ActivePremium.R return on an investment's annualized return minus the benchmark's annualized return
AdjustedSharpeRatio(R, Rf=0) # R/AdjustedSharpeRatio.R was introduced by Pezier and White to adjust for skewness and kurtosis by incorporating a penalty factor for negative skewness and excess kurtosis
AppraisalRatio(Ra, Rb, Rf=0, method=c("appraisal", "modified", "alternative")) # R/AppraisalRatio.R is the Jensen's alpha adjusted for specific risk. The numerator is divided by specific risk instead of total risk.
BernardoLedoitRatio(R) # R/BernadoLedoitratio.R take the sum of the subset of returns that are above 0 and we divide it by the opposite of the sum of the subset of returns that are below 0
BurkeRatio(R, Rf=0, modified=FALSE) # R/BurkeRatio.R modified Burke ratio is the Burke ratio multiplied by the square root of the number of datas
CalmarRatio(R, scale=NA) # R/CalmarRatio.R the ratio of annualized return over the absolute value of the maximum drawdown
SterlingRatio(R, scale=NA, excess=.1) # R/CalmarRatio.R adds an excess risk measure to the maximum drawdown
CAPM.alpha(Ra, Rb, Rf=0, ..., digits=3, benchmarkCols=T, method="LS", family="mopt", warning=T) # R/CAPM.alpha.R a wrapper for calculating a single factor model (CAPM) alpha
CAPM.beta(Ra, Rb, Rf=0, ..., digits=3, benchmarkCols=T, method="LS", family="mopt", warning=T) # R/CAPM.beta.R a wrapper for calculating a CAPM beta
CAPM.beta.bull(Ra, Rb, Rf=0, ..., digits=3, benchmarkCols=T, method="LS", family="mopt") # R/CAPM.beta.R a wrapper for calculating a conditional CAPM beta for up markets
CAPM.beta.bear(Ra, Rb, Rf=0, ..., digits=3, benchmarkCols=T, method="LS", family="mopt") # R/CAPM.beta.R a wrapper for calculating a conditional CAPM beta for down markets
TimingRatio(Ra, Rb, Rf=0, ...) # R/CAPM.beta.R the ratio of the two conditional CAPM betas (up and down)
CAPM.dynamic(Ra, Rb, Rf=0, Z, lags=1, ...) # R/CAPM.dynamic.R time-varying conditional single factor model beta
CAPM.epsilon(Ra, Rb, Rf=0, ...) # R/CAPM.epsilon.R an error term measuring the vertical distance between the return predicted by the equation and the real result
CAPM.jensenAlpha(Ra, Rb, Rf=0, ..., method="LS", family="mopt", series=FALSE) # R/CAPM.jensenAlpha.R the intercept of the regression equation and is in effect the exess return adjusted for systematic risk
CAPM.CML.slope(Rb, Rf=0 ) # R/CAPM.utils.R the slope of the Capital Market Line for looking at how a particular asset compares to the CML
CAPM.CML(Ra, Rb, Rf=0) # R/CAPM.utils.R capital market line
CAPM.RiskPremium(Ra, Rf=0) # R/CAPM.utils.R the measure of how much the asset's performance differs from the risk free rate
CAPM.SML.slope(Rb, Rf=0)# R/CAPM.utils.R the slope of the Security Market Line for looking at how a particular asset compares to the SML created by the benchmark
CDaR.alpha(R, Rm, p=0.95, weights=NULL, geometric=TRUE, type=NULL, ...) # R/CDAR.alpha.R conditional drawdown alpha
CDaR.beta(R, Rm, p=0.95, weights=NULL, geometric=TRUE, type=NULL, ...) # R/CDAR.beta.R conditional drawdown beta
DownsideDeviation(R, MAR=0, method=c("full","subset"), ..., potential=FALSE, SE=FALSE, SE.control=NULL) # R/DownsideDeviation.R similar to semi deviation, eliminates positive returns
DownsidePotential(R, MAR=0) # R/DownsideDeviation.R downside potential
DownsideFrequency(R, MAR=0, ...) # R/DownsideFrequency.R downside frequency of the return distribution
DownsideSharpeRatio(R, rf=0, SE=FALSE, SE.control=NULL, ...) # R/DownsideSharpeRatio.R the ratio of the mean excess return to the square root of lower semivariance
DRatio(R, ...) # R/DRatio.R similar to the Bernado Ledoit ratio but inverted and taking into account the frequency of positive and negative returns
DrawdownPeak(R, ...) # R/DrawdownPeak.R for each return its drawdown since the previous peak
Drawdowns(R, geometric=TRUE, ...) # R/Drawdowns.R the drawdown levels in a timeseries
CVaR <- ES(R=NULL, p=0.95, ..., method=c("modified", "gaussian", "historical"), clean=c("none", "boudt", "geltner", "locScaleRob"), portfolio_method=c("single", "component"), weights=NULL, mu=NULL, sigma=NULL, m3=NULL, m4=NULL, invert=TRUE, operational=TRUE, SE=FALSE, SE.control=NULL, p.tr=0.97, init=c(1.00, 0.3), nsim=10000) # R/ES.R Expected Shortfall or Conditional Value-at-Risk(CVaR)
FamaBeta(Ra, Rb, ...) # R/FamaBeta.R a beta used to calculate the loss of diversification
findDrawdowns(R, geometric=TRUE, ...) # R/findDrawdowns.R finds the starting period, the ending period, and the amount and length of the drawdown
Frequency(R, ...) # R/Frequency.R the period of the return distribution (ie 12 if monthly return, 4 if quarterly return)
HerfindahlIndex(Ra, ...) # R/HerfindahlIndex.R no description
HurstIndex(R, ...) # R/HurstIndex.R  measures whether returns are mean reverting, totally random, or persistent
InformationRatio(Ra, Rb, scale=NA, ...) # R/InformationRatio.R the Active Premium divided by the Tracking Error
Kappa(R, MAR, l, ...) # R/Kappa.R a generalized downside risk-adjusted performance measure
KellyRatio(R, Rf=0, method="half") # R/KellyRatio.R Kelly criterion ratio (leverage or bet size) for a strategy
kurtosis(x, na.rm=FALSE, method=c("excess", "moment", "fisher", "sample", "sample_excess"), ...) # R/kurtosis.R kurtosis of a univariate distribution
lpm(R, n=2, threshold=0, about_mean=FALSE, SE=FALSE, SE.control=NULL, ...) # R/lpm.R a lower partial moment for a time series
M2Sortino(Ra, Rb, MAR=0, ...) # R/M2Sortino.R a M^2 calculated for downside risk instead of total risk
MarketTiming(Ra, Rb, Rf=0, method=c("TM", "HM"), ...) # R/MarketTiming.R estimate Treynor-Mazuy or Merton-Henriksson market timing model
MartinRatio(R, Rf=0, ...) # R/MartinRatio.R divide the difference of the portfolio return and the risk free rate by the Ulcer index
maxDrawdown(R, weights=NULL, geometric=TRUE, invert=TRUE, ...) # R/maxDrawdown.R the maximum drawdown from peak equity
mean.geometric(x, ...) # R/mean.utils.R the mean geometric return for a return series
mean.arithmetic(x, SE=FALSE, SE.control=NULL, ...) # R/mean.utils.R the mean arithmetic return for a return series
mean.stderr(x, ...) # R/mean.utils.R the standard error of the mean for a return series
mean.LCL(x, ci=0.95, ...) # R/mean.utils.R a lower bound for the confidence interval given
mean.UCL(x, ci=0.95, ...) # R/mean.utils.R an upper bound for the confidence interval given
MeanAbsoluteDeviation(R, ...) # R/MeanAbsoluteDeviation.R the sum of the absolute value of the difference between the returns and the mean of the returns and divide it by the number of returns
MinTrackRecord(R=NULL, Rf=0, refSR, p=0.95, weights=NULL, n=NULL, sr=NULL, sk=NULL, kr=NULL, ignore_skewness=FALSE, ignore_kurtosis=TRUE) # R/MinTRL.R how long should a track record be in order to have a p-level statistical confidence that its Sharpe ratio is above a given threshold
MM.NCE(R, as.mat=TRUE, ...) # R/MM.NCE.R the nearest comoment estimators as in Boudt, Cornilly and Verdonck (2020)
Modigliani(Ra, Rb, Rf=0, ...) # R/Modigliani.R the portfolio return adjusted upward or downward to match the benchmark's standard deviation
MSquared(Ra, Rb, Rf=0, ...) # R/MSquared.R M squared of the return distribution
MSquaredExcess(Ra, Rb, Rf=0, Method=c("geometric", "arithmetic"), ...) # R/MSquaredExcess.R the quantity above the standard M, there is a geometric excess return and an arithmetic excess return
NetSelectivity(Ra, Rb, Rf=0, ...) # R/NetSelectivity.R net selectivity of the return distribution
Omega(R, L=0, method=c("simple", "interp", "binomial", "blackscholes"), output=c("point", "full"), Rf=0, SE=FALSE, SE.control=NULL, ...) # R/Omega.R capture all of the higher moments of the returns distribution
OmegaExcessReturn(Ra, Rb, MAR=0, ...) # R/OmegaExcessReturn.R multiply the downside variance of the style benchmark by 3 times the style beta
OmegaSharpeRatio(R, MAR=0, ...) # R/OmegaSharpeRatio.R a conversion of the omega ratio to a ranking statistic in familiar form to the Sharpe ratio
PainIndex(R, ...) # R/PainIndex.R the mean value of the drawdowns over the entire analysis period
PainRatio(R, Rf=0, ...) # R/PainRatio.R divide the difference of the portfolio return and the risk free rate by the Pain index
ProbSharpeRatio(R=NULL, Rf=0, refSR, p=0.95, weights=NULL, n=NULL, sr=NULL, sk=NULL, kr=NULL, ignore_skewness=FALSE, ignore_kurtosis=TRUE) # R/ProbSharpeRatio.R probabilistic sharpe ratio
ProspectRatio(R, MAR, ...) # R/ProspectRatio.R a ratio used to penalise loss since most people feel loss greater than gain
RachevRatio(R, alpha=0.1, beta=0.1, rf=0, SE=FALSE, SE.control=NULL, ...) # R/RachevRatio.R a non-parametric estimator of the upper tail reward potential relative to the lower tail risk in a non-Gaussian setting
Return.annualized.excess(Rp, Rb, scale=NA, geometric=TRUE ) # R/Return.annualized.excess.R an annualized excess return
Return.annualized(R, scale=NA, geometric=TRUE, na.rm=TRUE) # R/Return.annualized.R an annualized return
Return.calculate(prices, method=c("discrete","log","difference")) # R/Return.calculate.R calculate simple or compound returns from prices
Return.convert(R, destinationType=c("discrete", "log", "difference", "level"), seedValue=NULL, initial=TRUE) # R/Return.convert.R coredata content from one type of return to another
Return.cumulative(R, geometric=TRUE) # R/Return.cumulative.R a compounded (geometric) cumulative return
Return.excess(R, Rf=0) # R/Return.excess.R the returns of an asset in excess of the given risk free rate
Return.Geltner(Ra, ...) # R/Return.Geltner.R calculate Geltner liquidity-adjusted return series
Return.locScaleRob(R, alpha.robust=0.05, normal.efficiency=0.99, ...) # R/Return.locScaleRob.R returns the data after passing through a robust location and scale filter
Return.portfolio <- Return.rebalancing(R, weights=NULL, wealth.index=FALSE, contribution=FALSE, geometric=TRUE, rebalance_on=c(NA, "years", "quarters", "months", "weeks", "days"), value=1, verbose=FALSE, ..., rebal_cost=0, full_investment=FALSE) # R/Return.portfolio.R
Return.relative(Ra, Rb, ...) # R/Return.relative.R calculate the relative return of one asset to another
RPESE.control(estimator=c("Mean", "SD", "VaR", "ES", "SR", "DSR", "SoR", "ESratio", "VaRratio", "SoR", "LPM", "OmegaRatio", "SemiSD", "RachevRatio"), se.method=NULL, cleanOutliers=NULL, fitting.method=NULL, freq.include=NULL, freq.par=NULL, a=NULL, b=NULL) # R/RPESE.control.R computation of Standard Errors for Risk and Performance estimators
Selectivity(Ra, Rb, Rf=0, ...) # R/Selectivity.R the same as Jensen's alpha
SemiDeviation(R, SE=FALSE, SE.control=NULL, ...) # R/SemiDeviation.R a wrapper of DownsideDeviation with MAR=mean(x)
SFM.coefficients(Ra, Rb, Rf=0, subset=TRUE, ..., method="Robust", family="mopt", digits=3, benchmarkCols=T, Model=F, warning=T) # R/SFM.coefficients.R calculate single factor model alpha and beta coefficients
SFM.fit.models(Ra, Rb, Rf=0, family="mopt", which.plots=NULL, plots=TRUE) # R/SFM.fit.models.R compare SFM estimated using robust estimators with that estimated by OLS
SharpeRatio.annualized(R, Rf=0, scale=NA, geometric=FALSE, ...) # R/SharpeRatio.annualized.R annualized Sharpe Ratio
SharpeRatio(R, Rf=0, p=0.95, FUN=c("StdDev", "VaR", "ES", "SemiSD"), weights=NULL, annualize=FALSE, geometric=FALSE, SE=FALSE, SE.control=NULL, ...) # R/SharpeRatio.R a traditional or modified Sharpe Ratio of return over StdDev or VaR or ES
skewness(x, na.rm=FALSE, method=c("moment", "fisher", "sample"), ...) # R/skewness.R skewness of a univariate distribution
SkewnessKurtosisRatio(R, ...) # R/SkewnessKurtosisRatio.R the division of Skewness by Kurtosis
SmoothingIndex(R, neg.thetas=FALSE, MAorder=2, verbose=FALSE, ...) # R/SmoothingIndex.R calculate Normalized Getmansky Smoothing Index
sortDrawdowns(runs) # R/sortDrawdowns.R gives the drawdowns in order of worst to best
SortinoRatio(R, MAR=0, ..., weights=NULL, SE=FALSE, SE.control=NULL) # R/SortinoRatio.R calculate Sortino Ratio of performance over downside risk
SpecificRisk(Ra, Rb, Rf=0,  ...) # R/SpecificRisk.R the standard deviation of the error term in the regression equation
StdDev.annualized <- sd.annualized <- sd.multiperiod(x, scale=NA, ..., sample_method=c("unbiased", "ML")) # R/StdDev.annualized.R calculate a multiperiod or annualized Standard Deviation
StdDev(R, ..., clean=c("none", "boudt", "geltner", "locScaleRob"), portfolio_method=c("single", "component"), weights=NULL, mu=NULL, sigma=NULL, use="everything", method=c("pearson", "kendall", "spearman"), sample_method=c("unbiased", "ML"), SE=FALSE, SE.control=NULL) # R/StdDev.R for univariate and multivariate series, also calculates component contribution to standard deviation of a portfolio
SystematicRisk(Ra, Rb, Rf=0, scale=NA, ...) # R/SystematicRisk.R the product of beta by market risk
TotalRisk(Ra, Rb, Rf=0, ...) # R/TotalRisk.R the square of total risk is the sum of the square of systematic risk and the square of specific risk
TrackingError(Ra, Rb, scale=NA) # R/TrackingError.R a measure of the unexplained portion of performance relative to a benchmark
TreynorRatio(Ra, Rb, Rf=0, scale=NA, modified=FALSE) # R/TreynorRatio.R calculate Treynor Ratio or modified Treynor Ratio of excess return over CAPM beta
UlcerIndex(R, ...) # R/UlcerIndex.R similar to drawdown deviation except that the impact of the duration of drawdowns is incorporated
UpDownRatios(Ra, Rb, method=c("Capture", "Number", "Percent"), side=c("Up", "Down"), geometric=TRUE) # R/UpDownRatios.R metrics on up and down markets for the benchmark asset
UpsideFrequency(R, MAR=0, ...) # R/UpsideFrequency.R the subset of returns that are more than MAR and divide the length of this subset by the total number of returns
UpsidePotentialRatio(R, MAR=0, method=c("subset", "full")) # R/UpsidePotentialRatio.R calculate Upside Potential Ratio of upside performance over downside risk
UpsideRisk(R, MAR=0, method=c("full","subset"), stat=c("risk","variance","potential"), ...) # R/UpsideRisk.R similar of semideviation taking the return above the MAR instead of using the mean return or zero
VaR.backtest(R, VaR, p=0.95) # R/VaR.backtest.R a simple binomial backtest for a VaR model
VaR.gpd(R, p, SE=FALSE, ...) # R/VaR.gpd.R Value at Risk and Expected Shortfall via Generalized Pareto Distribution (GPD)
VaR.lognormal(R, p, ...) # R/VaR.lognormal.R Value at Risk and Expected Shortfall via Lognormal Distribution
VaR.Marginal(R, p=0.95, method=c("modified","gaussian","historical"), weightingvector=NULL) # R/VaR.Marginal.R marginal VaR
VaR.montecarlo(R, p, nsim=10000, ...) # R/VaR.montecarlo.R Value at Risk and Expected Shortfall via Monte Carlo Simulation
VaR(R=NULL, p=0.95, ..., method=c("modified", "gaussian", "historical", "kernel", "gpd", "lognormal", "montecarlo"), clean=c("none", "boudt", "geltner", "locScaleRob"), portfolio_method=c("single", "component", "marginal"), weights=NULL, mu=NULL, sigma=NULL, m3=NULL, m4=NULL, invert=TRUE, SE=FALSE, SE.control=NULL, p.tr=0.97, init=c(1.00, 0.3), nsim=10000) # R/VaR.R calculate various VaR measures
VolatilitySkewness(R, MAR=0, stat=c("volatility", "variability"), ...) # R/VolatilitySkewness.R a similar measure to omega but using the second partial moment

## My own initial implementation

Located in ./py/performance/ratios.py file

cumulative_return()
drawdowns_cumulative()
min_drawdowns_cumulative()
worst_drawdowns_cumulative()
drawdowns_peaks() -> array
drawdowns_continuous(peaks_only=False, max_peaks: int=None) -> array
skew()
kurtosis()
sharpe_ratio(ignore_risk_free_rate=False, autocorrelation_penalty=False)
sortino_ratio(autocorrelation_penalty=False, divide_by_sqrt2=False)
omega_ratio()
kappa_ratio(order: int=3)
kappa3_ratio(order: int=3)
bernardo_ledoit_ratio()
upside_potential_ratio(full =True)
compound_growth_rate()
calmar_ratio()
sterling_ratio(annual_excess_rate: float=0)
burke_ratio(modified=False)
pain_index()
pain_ratio()
ulcer_index()
martin_ratio()
gain_to_pain_ratio()
risk_of_ruin()
risk_return_ratio()

In `bacon3` ch5-13.md file there is a `Table 5.27 — Periodic table of risk measures`.
Please make an overview of functionality I'm missing comparing to this table and plan implementation in 1 phases:

- phase 1: single return series
- phase 2: multiple return series, i.e. benchmarks and portfolios

Write the plan into ./external/to-implement.md
