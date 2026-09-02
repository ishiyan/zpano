# Generating reference data from PerformanceAnalytics R package

The data were produced with the R scripts using the [datacamp.com](https://www.datacamp.com/datalab/w/28c21593-21e6-47d9-8e72-acebdd3be32c/edit) online R interpreter

Every script begins with loading the packages, `PerformanceAnalytics` and the optional `RobStatTM`, followed by the test data loading.

We use the monthly [portfolio_bacon.csv](https://github.com/braverock/PerformanceAnalytics/blob/master/data/portfolio_bacon.csv) as the test data.

Depending on a script, we may optionally convert monthly dates to daily dates or to yearly dates.
This is needed because some PerformanceAnalytic functions derive the scaling (periods per annum) from the two first returns.

Overall, the common preambule is

````R
# Load necessary libraries
if(!require('PerformanceAnalytics')) {
    install.packages('PerformanceAnalytics')
    library('PerformanceAnalytics')
}
# Load optional libraries
if(!require('RobStatTM')) {
    install.packages('RobStatTM')
    library('RobStatTM')
}

# Load the data
data(portfolio_bacon)
# head(portfolio_bacon, 100)
# write.csv(portfolio_bacon)

periodicity <- "monthly" # "daily", "monthly", "yearly"

if (periodicity == "daily") {
    # make daily returns from monthly ones
    dates <- as.Date(index(portfolio_bacon))
    #print(paste("dates",dates))

    # Generate a sequence of daily dates starting from the first date in the dataset
    start_date <- dates[1]
    end_date <- dates[length(dates)]
    #print(paste("start=",start_date,"end=", end_date))
    daily_dates <- seq.Date(from = start_date, to = end_date, by = "day")

    # Ensure the number of daily dates matches the number of rows in the dataset
    if (length(daily_dates) >= nrow(portfolio_bacon)) {
        daily_dates <- daily_dates[1:nrow(portfolio_bacon)]
    } else {
        stop("The generated sequence of daily dates is shorter than the number of rows in the dataset.")
    }

    index(portfolio_bacon) <- daily_dates
}

if (periodicity == "yearly") {
    # make yearly return dates from the monthly ones
    dates <- as.Date(index(portfolio_bacon))
    #print(paste("dates",dates))

    # Generate a sequence of yearly dates starting from the first date in the dataset
    start_date <- dates[1]
    end_date <- dates[length(dates)]
    #print(paste("start=",start_date,"end=", end_date))
    yearly_dates <- seq.Date(from = start_date, by = "year", length.out = nrow(portfolio_bacon))

    # Ensure the number of yearly dates matches the number of rows in the dataset
    if (length(yearly_dates) > nrow(portfolio_bacon)) {
        yearly_dates <- yearly_dates[1:nrow(portfolio_bacon)]
    } else if (length(yearly_dates) < nrow(portfolio_bacon)) {
        stop("Not enough yearly dates to match the number of rows in the dataset.")
    }

    index(portfolio_bacon) <- yearly_dates
}

portfolio_length <- nrow(portfolio_bacon)
````

This preambule is followed by one or more data generation pieces which are listed per function below.

- autocorrelation_penalty (no data)
- [cumulative_geometric_return](#cumulative_geometric_return)
- [geometric_mean_return](#geometric_mean_return)
- compound_annual_growth_rate (no data)
- [skewness](#skewness), skewness_moment, skewness_fisher, skewness_sample
- [kurtosis](#kurtosis), kurtosis_excess, kurtosis_moment, kurtosis_sample_excess, kurtosis_sample_corrected, kurtosis_sample
- [skewness_kurtosis_ratio](#skewness_kurtosis_ratio)
- jarque_bera_normality_test_statistic (no data)
- is_normal_distribution (no data)
- [var](#var) var_historical, var_gaussian, var_cornish_fisher
- [es](#es) es_historical, es_gaussian, es_cornish_fisher
- reward_to_var_ratio_historical, reward_to_var_ratio_gaussian, reward_to_var_ratio_cornish_fisher (no data)
- reward_to_es_ratio_historical, reward_to_es_ratio_gaussian, reward_to_es_ratio_cornish_fisher (no data)
- mean_absolute_deviation_ratio (no data)
- [upside_potential_ratio](#upside_potential_ratio), upside_potential_ratio_subset
- [upside_frequency](#upside_frequency)
- [upside_risk](#upside_risk), upside_risk_subset, upside_variance, upside_variance_subset, upside_potential, upside_potential_subset
- [semi_deviation](#semi_deviation)
- [downside_deviation](#downside_deviation), downside_deviation_subset
- [downside_frequency](#downside_frequency)
- [downside_potential](#downside_potential)
- [sharpe_ratio](#sharpe_ratio), sharpe_ratio_var_historical, sharpe_ratio_var_gaussian, sharpe_ratio_var_cornish_fisher, sharpe_ratio_es_historical, sharpe_ratio_es_gaussian, sharpe_ratio_es_cornish_fisher
- [downside_sharpe_ratio](#downside_sharpe_ratio)
- [adjusted_sharpe_ratio](#adjusted_sharpe_ratio)
- adjusted_sharpe_ratio_skew_only (no data)
- [probabilistic_sharpe_ratio](#probabilistic_sharpe_ratio), probabilistic_sharpe_ratio_full, probabilistic_sharpe_ratio_symmetric, probabilistic_sharpe_ratio_gaussian
- [sortino_ratio](#sortino_ratio), sortino_ratio_sqrt2
- sortino_satchell_ratio (no data)
- [omega_ratio](#omega_ratio)
- [omega_sharpe_ratio](#omega_sharpe_ratio)
- [omega_excess_return](#omega_excess_return)
- [kappa_1_ratio](#kappa_1_ratio), kappa_2_ratio, kappa_3_ratio, kappa_4_ratio
- [bernardo_ledoit_ratio](#bernardo_ledoit_ratio), gain_loss_ratio
- [d_ratio](#d_ratio)
- mean_non_zero_return (no data)
- mean_win_return (no data)
- mean_loss_return (no data)
- win_rate (no data)
- loss_rate (no data)
- [volatility_skewness](#volatility_skewness), variability_skewness
- farinelli_tibiletti_ratio (no data)
- [rachev_ratio](#rachev_ratio)
- [drawdowns_cumulative](#drawdowns_cumulative)
- [min_drawdowns_cumulative](#min_drawdowns_cumulative), worst_drawdowns_cumulative
- [drawdowns_high_watermark](#drawdowns_high_watermark)
- [calmar_ratio](#calmar_ratio)
- [sterling_ratio](#sterling_ratio)
- [burke_ratio](#burke_ratio), burke_ratio_modified
- [pain_index](#pain_index)
- [pain_ratio](#pain_ratio)
- [ulcer_index](#ulcer_index)
- [martin_ratio](#martin_ratio)
- [drawdown_average](#drawdown_average)
- [drawdown_average_length](#drawdown_average_length)
- [drawdown_average_peak_to_trough](#drawdown_average_peak_to_trough)
- [drawdown_average_recovery](#drawdown_average_recovery)
- [drawdown_deviation](#drawdown_deviation)
- [cdar_average](#cdar_average), cdar_discrete
- [cdar_beta](#cdar_beta)
- [cdar_alpha](#cdar_alpha)
- reward_to_conditional_drawdown (no data)
- [sfm_risk_premium](#sfm_risk_premium)
- [sfm_alpha](#sfm_alpha)
- [sfm_beta](#sfm_beta)
- [sfm_beta_bull](#sfm_beta_bull)
- [sfm_beta_bear](#sfm_beta_bear)
- [timing_ratio](#timing_ratio)
- [sfm_r2](#sfm_r2)
- [jensen_alpha](#jensen_alpha), jensen_alpha_annualized
- [fama_beta](#fama_beta)
- [modigliani](#modigliani)
- [tracking_error](#tracking_error)
- [active_premium](#active_premium)
- [information_ratio](#information_ratio)
- information_ratio_modified (no data)
- [systematic_risk](#systematic_risk)
- [treynor_ratio](#treynor_ratio), treynor_ratio_modified
- [specific_risk](#specific_risk)
- [total_risk](#total_risk)
- [appraisal_ratio](#appraisal_ratio) , jensen_alpha_modified, jensen_alpha_alternative
- [m_squared](#m_squared)
- [m_squared_excess](#m_squared_excess)
- [m_squared_sortino](#m_squared_sortino)
- [prospect_ratio](#prospect_ratio)
- tail_ratio (no data)
- [kelly_ratio](#kelly_ratio), kelly_ratio_full
- [hurst_exponent](#hurst_exponent)
- bias_ratio (no data)
- [upside_capture_ratio](#upside_capture_ratio)
- [downside_capture_ratio](#downside_capture_ratio)
- overall_capture_ratio (no data)
- [up_number_ratio](#up_number_ratio)
- [down_number_ratio](#down_number_ratio)
- [up_percentage_ratio](#up_percentage_ratio)
- [down_percentage_ratio]($down_percentage_ratio)

- NetSelectivity
- Return.calculate
- Return.convert
- sharpe annualized? geometric?
- VaR/ES VaR.gpd, VaR.lognormal

## cumulative_geometric_return {#cumulative_geometric_return}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/Return.cumulative.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("Return.cumulative"))
for(geometric in c(TRUE, FALSE)) {
    print(geometric)
    for (i in 1:portfolio_length) {
        result <- Return.cumulative(portfolio_bacon[1:i,], geometric=geometric)
        write.csv(result)
    }
}
```

## geometric_mean_return {#geometric_mean_return}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/mean.utils.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("mean.utils"))
for (i in 1:portfolio_length) {
    result <- mean.geometric(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## skewness {#skewness}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/skewness.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("skewness"))
for (method in c("moment", "fisher", "sample")) {
    print(paste("method:", method))
    for (i in 1:portfolio_length) {
        result <- skewness(portfolio_bacon[1:i, 1], method=method)
        write.csv(result)
    }
}
```

## kurtosis {#kurtosis}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/kurtosis.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("kurtosis"))
for (method in c("excess", "moment", "fisher", "sample", "sample_excess")) {
    print(paste("method: ", method))
    for (i in 1:portfolio_length) {
        result <- kurtosis(portfolio_bacon[1:i, ], method=method)
        write.csv(result)
    }
}
```

## skewness_kurtosis_ratio {#skewness_kurtosis_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/SkewnessKurtosisRatio.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("SkewnessKurtosisRatio"))
for (i in 1:portfolio_length) {
    result <- SkewnessKurtosisRatio(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## var {#var}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/VaR.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("VaR"))
for (method in c("modified", "gaussian", "historical")) { # modified, gaussian, historical, kernel, gpd, lognormal, montecarlo
    for (p in c(0.90, 0.95, 0.975, 0.99, 0.995, 0.999)) {
            print(paste("method:", method, "p:", p))
        for (i in 1:portfolio_length) {
            result <- VaR(portfolio_bacon[1:i, 1], method=method, p=p, invert=FALSE)
            write.csv(result)
        }
    }
}
```

## es {#es}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/ES.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("ES"))
for (method in c("modified", "gaussian", "historical")) { # modified, gaussian, historical
    for (p in c(0.90, 0.95, 0.975, 0.99, 0.995, 0.999)) {
        print(paste("method:", method, "p:", p))
        for (i in 1:portfolio_length) {
            result <- ES(portfolio_bacon[1:i, 1], method=method, p=p, invert=FALSE)
            write.csv(result)
        }
    }
}
```

## upside_potential_ratio {#upside_potential_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpsidePotentialRatio.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("UpsidePotentialRatio"))
for(method in c("full", "subset")) {
    for(mar in seq(0, 0.001, by=0.001)) {
        print(paste("method", method, "mar", mar))
        for (i in 1:portfolio_length) {
            result <- UpsidePotentialRatio(portfolio_bacon[1:i,1], method=method, MAR=mar)
            write.csv(result)
        }
    }
    for(mar in seq(0.005, 0.01, by=0.005)) {
        print(paste("method", method, "mar", mar))
        for (i in 1:portfolio_length) {
            result <- UpsidePotentialRatio(portfolio_bacon[1:i,1], method=method, MAR=mar)
            write.csv(result)
        }
    }
    for(mar in seq(0.05, 0.3, by=0.05)) {
        print(paste("method", method, "mar", mar))
        for (i in 1:portfolio_length) {
            result <- UpsidePotentialRatio(portfolio_bacon[1:i,1], method=method, MAR=mar)
            write.csv(result)
        }
    }
}
```

## upside_frequency {#upside_frequency}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpsideFrequency.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("UpsideFrequency"))
for(mar in seq(0, 0.001, by=0.001)) {
    print(paste("mar:", mar))
    for (i in 1:portfolio_length) {
        result <- UpsideFrequency(portfolio_bacon[1:i,1], MAR=mar)
        write.csv(result)
    }
}
for(mar in seq(0.005, 0.01, by=0.005)) {
    print(paste("mar:", mar))
    for (i in 1:portfolio_length) {
        result <- UpsideFrequency(portfolio_bacon[1:i,1], MAR=mar)
        write.csv(result)
    }
}
for(mar in seq(0.05, 0.3, by=0.05)) {
    print(paste("mar:", mar))
    for (i in 1:portfolio_length) {
        result <- UpsideFrequency(portfolio_bacon[1:i,1], MAR=mar)
        write.csv(result)
    }
}
```

## upside_risk {#upside_risk}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpsideRisk.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("UpsideRisk"))
for(method in c("full", "subset")) { # c("full", "subset")
    for(stat in c("risk", "variance", "potential")) { # c("risk", "variance", "potential")
        for(mar in c(0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.3)) {
            print(paste("method", method, "stat", stat, "MAR", mar))
            for (i in 1:portfolio_length) {
                result <- UpsideRisk(portfolio_bacon[1:i,1], method=method, stat=stat, MAR=mar)
                write.csv(result)
            }
        }
    }
}
```

## semi_deviation {#semi_deviation}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/SemiDeviation.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("SemiDeviation"))
for (i in 1:portfolio_length) {
    result <- SemiDeviation.R(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## downside_deviation {#downside_deviation}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/DownsideDeviation.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("DownsideDeviation"))
for(method in c("full", "subset")) {
    for(mar in seq(0, 0.001, by=0.001)) {
        print(paste("method", method, "mar", mar))
        for (i in 1:portfolio_length) {
            result <- DownsideDeviation(portfolio_bacon[1:i,1], method=method, MAR=mar)
            write.csv(result)
        }
    }
    for(mar in seq(0.005, 0.01, by=0.005)) {
        print(paste("method", method, "mar", mar))
        for (i in 1:portfolio_length) {
            result <- DownsideDeviation(portfolio_bacon[1:i,1], method=method, MAR=mar)
            write.csv(result)
        }
    }
    for(mar in seq(0.05, 0.3, by=0.05)) {
        print(paste("method", method, "mar", mar))
        for (i in 1:portfolio_length) {
            result <- DownsideDeviation(portfolio_bacon[1:i,1], method=method, MAR=mar)
            write.csv(result)
        }
    }
}
```

## downside_frequency {#downside_frequency}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/DownsideFrequency.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("DownsideFrequency"))
for(mar in seq(0, 0.001, by=0.001)) {
    print(paste("mar:", mar))
    for (i in 1:portfolio_length) {
        result <- DownsideFrequency(portfolio_bacon[1:i,1], MAR=mar)
        write.csv(result)
    }
}
for(mar in seq(0.005, 0.01, by=0.005)) {
    print(paste("mar:", mar))
    for (i in 1:portfolio_length) {
        result <- DownsideFrequency(portfolio_bacon[1:i,1], MAR=mar)
        write.csv(result)
    }
}
for(mar in seq(0.05, 0.3, by=0.05)) {
    print(paste("mar:", mar))
    for (i in 1:portfolio_length) {
        result <- DownsideFrequency(portfolio_bacon[1:i,1], MAR=mar)
        write.csv(result)
    }
}
```

## downside_potential {#downside_potential}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/DownsideDeviation.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("DownsidePotential"))
for(mar in seq(0, 0.001, by=0.001)) {
    print(paste("mar:", mar))
    for (i in 1:portfolio_length) {
        result <- DownsidePotential(portfolio_bacon[1:i,1], MAR=mar)
        write.csv(result)
    }
}
for(mar in seq(0.005, 0.01, by=0.005)) {
    print(paste("mar:", mar))
    for (i in 1:portfolio_length) {
        result <- DownsidePotential(portfolio_bacon[1:i,1], MAR=mar)
        write.csv(result)
    }
}
for(mar in seq(0.05, 0.3, by=0.05)) {
    print(paste("mar:", mar))
    for (i in 1:portfolio_length) {
        result <- DownsidePotential(portfolio_bacon[1:i,1], MAR=mar)
        write.csv(result)
    }
}
```

## sharpe_ratio {#sharpe_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/SharpeRatio.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("SharpeRatio"))
# "SemiSD" is DownsideSharpeRatio and produces the same output, skipping it.
for(fun in c("StdDev")) { # c("StdDev", "VaR", "ES", "SemiSD")
    for(rf in seq(0, 0.001, by=0.001)) {
        print(paste("FUN", fun, "Rf", rf))
        for (i in 1:portfolio_length) {
            result <- SharpeRatio(portfolio_bacon[1:i,1], FUN=fun, Rf=rf)
            write.csv(result)
        }
    }
    for(rf in seq(0.005, 0.01, by=0.005)) {
        print(paste("FUN", fun, "Rf", rf))
        for (i in 1:portfolio_length) {
            result <- SharpeRatio(portfolio_bacon[1:i,1], FUN=fun, Rf=rf)
            write.csv(result)
        }
    }
    for(rf in seq(0.05, 0.3, by=0.05)) {
        print(paste("FUN", fun, "Rf", rf))
        for (i in 1:portfolio_length) {
            result <- SharpeRatio(portfolio_bacon[1:i,1], FUN=fun, Rf=rf)
            write.csv(result)
        }
    }
}
for(method in c("historical", "gaussian", "modified")) { # c("historical", "gaussian", "modified")
    for(fun in c("VaR", "ES")) { # c("StdDev", "VaR", "ES", "SemiSD")
        for (p in c(0.90, 0.95, 0.99, 0.999)) {
            for(rf in seq(0, 0.001, by=0.001)) {
                print(paste("method", method, "FUN", fun, "p", p, "Rf", rf))
                for (i in 1:portfolio_length) {
                    result <- SharpeRatio(portfolio_bacon[1:i,1], FUN=fun, Rf=rf, p=p, method=method)
                    write.csv(result)
                }
            }
            for(rf in seq(0.01, 0.01, by=0.005)) {
                print(paste("method", method, "FUN", fun, "p", p, "Rf", rf))
                for (i in 1:portfolio_length) {
                    result <- SharpeRatio(portfolio_bacon[1:i,1], FUN=fun, Rf=rf, p=p, method=method)
                    write.csv(result)
                }
            }
            for(rf in seq(0.1, 0.1, by=0.05)) {
                print(paste("method", method, "FUN", fun, "p", p, "Rf", rf))
                for (i in 1:portfolio_length) {
                    result <- SharpeRatio(portfolio_bacon[1:i,1], FUN=fun, Rf=rf, p=p, method=method)
                    write.csv(result)
                }
            }
        }
    }
}
```

## downside_sharpe_ratio {#downside_sharpe_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/DownsideSharpeRatio.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("DownsideSharpeRatio"))
for(rf in seq(0, 0.001, by=0.001)) {
    print(paste("rf:", rf))
    for (i in 1:portfolio_length) {
        result <- DownsideSharpeRatio(portfolio_bacon[1:i,1], rf=rf)
        write.csv(result)
    }
}
for(rf in seq(0.005, 0.01, by=0.005)) {
    print(paste("rf:", rf))
    for (i in 1:portfolio_length) {
        result <- DownsideSharpeRatio(portfolio_bacon[1:i,1], rf=rf)
        write.csv(result)
    }
}
for(rf in seq(0.05, 0.3, by=0.05)) {
    print(paste("rf:", rf))
    for (i in 1:portfolio_length) {
        result <- DownsideSharpeRatio(portfolio_bacon[1:i,1], rf=rf)
        write.csv(result)
    }
}
```

## adjusted_sharpe_ratio {#adjusted_sharpe_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/AdjustedSharpeRatio.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("AdjustedSharpeRatio"))
for(rf in seq(0, 0.001, by=0.001)) {
    print(paste("Rf", rf))
    for (i in 1:portfolio_length) {
        result <- AdjustedSharpeRatio(portfolio_bacon[1:i,1], Rf=rf)
        write.csv(result)
    }
}
for(rf in seq(0.005, 0.01, by=0.005)) {
    print(paste("Rf", rf))
    for (i in 1:portfolio_length) {
        result <- AdjustedSharpeRatio(portfolio_bacon[1:i,1], Rf=rf)
        write.csv(result)
    }
}
for(rf in seq(0.05, 0.3, by=0.05)) {
    print(paste("Rf", rf))
    for (i in 1:portfolio_length) {
        result <- AdjustedSharpeRatio(portfolio_bacon[1:i,1], Rf=rf)
        write.csv(result)
    }
}
```

## probabilistic_sharpe_ratio {#probabilistic_sharpe_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/AdjustedSharpeRatio.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("AdjustedSharpeRatio"))
for(rf in seq(0, 0.001, by=0.001)) {
    print(paste("Rf", rf))
    for (i in 1:portfolio_length) {
        result <- AdjustedSharpeRatio(portfolio_bacon[1:i,1], Rf=rf)
        write.csv(result)
    }
}
for(rf in seq(0.005, 0.01, by=0.005)) {
    print(paste("Rf", rf))
    for (i in 1:portfolio_length) {
        result <- AdjustedSharpeRatio(portfolio_bacon[1:i,1], Rf=rf)
        write.csv(result)
    }
}
for(rf in seq(0.05, 0.3, by=0.05)) {
    print(paste("Rf", rf))
    for (i in 1:portfolio_length) {
        result <- AdjustedSharpeRatio(portfolio_bacon[1:i,1], Rf=rf)
        write.csv(result)
    }
}
```

## sortino_ratio {#sortino_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/SortinoRatio.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("SortinoRatio"))
for(mar in seq(0, 0.3, by=0.05)) {
    for (i in 1:portfolio_length) {
        result <- SortinoRatio(portfolio_bacon[1:i,], MAR=mar)
        write.csv(result)
    }
}
```

## omega_ratio {#omega_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/Omega.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("Omega"))
for (l in seq(0.0, 0.1, by=0.02)) { # l is actually MAR in our implementation
    for(rf in c(0)) { # results doesn't change when rf is changed
        print(paste("L: ", l, "Risk-free rate: ", rf))
        for (i in 1:portfolio_length) {
            result <- Omega(portfolio_bacon[1:i, 1], L=l,
                Rf=rf, method="simple", output="point")
            write.csv(result)
        }
    }
}
```

## omega_sharpe_ratio {#omega_sharpe_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/OmegaSharpeRatio.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("OmegaSharpeRatio"))
for(mar in seq(0, 0.3, by=0.05)) {
    print(paste("MAR", mar))
    for (i in 1:portfolio_length) {
        result <- OmegaSharpeRatio(portfolio_bacon[1:i,1], MAR=mar)
        write.csv(result)
    }
}
```

## omega_excess_return {#omega_excess_return}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/OmegaExcessReturn.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("OmegaExcessReturn"))
print(paste("with benchmaek", "MAR", mar))
# First sample is used to determine periodicity
for (i in 2:portfolio_length) {
    result <- OmegaExcessReturn(portfolio_bacon[1:i,1], Rb=portfolio_bacon[1:i,2], MAR=mar)
    write.csv(result)
}
print(paste("with self", "MAR", mar))
# First sample is used to determine periodicity
for (i in 2:portfolio_length) {
    result <- OmegaExcessReturn(portfolio_bacon[1:i,1], Rb=portfolio_bacon[1:i,1], MAR=mar)
    write.csv(result)
}
```

## kappa_ratio {#kappa_1_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/Kappa.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("Kappa"))
for (l in c(1,2,3,4)) {
    for(mar in seq(0, 0.3, by=0.05)) {
        print(paste("L: ", l, "MAR: ", mar))
        for (i in 1:portfolio_length) {
            result <- Kappa(portfolio_bacon[1:i, ], MAR=mar, l=l)
            write.csv(result)
        }
    }
}
```

## bernardo_ledoit_ratio {#bernardo_ledoit_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/BernadoLedoitratio.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("BernadoLedoitratio"))
for (i in 1:portfolio_length) {
    result <- BernardoLedoitRatio(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## d_ratio {#d_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/DRatio.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("DRatio"))
for (i in 1:portfolio_length) {
    result <- DRatio(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## volatility_skewness {#volatility_skewness}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/VolatilitySkewness.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("VolatilitySkewness"))
for (stat in c("volatility", "variability")) {
    for(mar in c(0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.3)) {
        print(paste("stat", stat, "MAR", mar))
        for (i in 1:portfolio_length) {
            result <- VolatilitySkewness(portfolio_bacon[1:i, 1], stat=stat, MAR=mar)
            write.csv(result)
        }
    }
}
```

## rachev_ratio {#rachev_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/RachevRatio.R)

```R
################################################################
# Use daily dates
################################################################
print(paste("RachevRatio"))
for (alpha in c(0.05, 0.1)) { # c(0.05, 0.1)
    for (beta in c(0.05, 0.1)) { # c(0.05, 0.1)
        for (rf in c(0.0, 0.005, 0.01, 0.05, 0.1)) {
            print(paste("alpha:", alpha, "beta:", beta, "Rf:", rf))
            for (i in 1:portfolio_length) {
                result <- RachevRatio(portfolio_bacon[1:i, 1], alpha=alpha, beta=beta, Rf=rf)
                write.csv(result)
            }
        }
    }
}
```

## drawdowns_cumulative {#drawdowns_cumulative}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/Drawdowns.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("Drawdowns"))
for (i in 1:portfolio_length) {
    result <- Drawdowns(portfolio_bacon[1:i, 1], geometric = TRUE)
    write.csv(result)
}
```

## min_drawdowns_cumulative {#min_drawdowns_cumulative}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/maxDrawdown.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("maxDrawdown"))
print(paste("invert", FALSE))
for (i in 1:portfolio_length) {
    result <- maxDrawdown(portfolio_bacon[1:i, 1], geometric = TRUE, invert=FALSE)
    write.csv(result)
}
print(paste("invert", TRUE))
for (i in 1:portfolio_length) {
    result <- maxDrawdown(portfolio_bacon[1:i, 1], geometric = TRUE, invert=TRUE)
    write.csv(result)
}
```

## drawdowns_high_watermark {#drawdowns_high_watermark}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/DrawdownPeak.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("DrawdownPeak"))
for (i in 1:portfolio_length) {
    result <- DrawdownPeak(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## calmar_ratio {#calmar_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/CalmarRatio.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("CalmarRatio"))
for (i in 1:portfolio_length) {
    result <- CalmarRatio(portfolio_bacon[1:i, 1], scale=1)
    write.csv(result)
}
```

## sterling_ratio {#sterling_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/SterlingRatio.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("SterlingRatio"))
for (excess in seq(0.0, 0.1, by=0.02)) {
    print(paste("Excess", excess))
    for (i in 1:portfolio_length) {
        result <- SterlingRatio(portfolio_bacon[1:i, ], excess=excess, scale=1)
        write.csv(result)
    }
}
```

## burke_ratio {#burke_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/BurkeRatio.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("BurkeRatio"))
for (modified in c(FALSE,TRUE)) {
    for (rf in seq(0.0, 0.1, by=0.02)) {
        print(paste("modified: ", modified, "Rf: ", rf))
        # first value (i=1) is always None
        for (i in 2:portfolio_length) {
            result <- BurkeRatio(portfolio_bacon[1:i, 1], Rf=rf, modified=modified)
            write.csv(result)
        }
    }
}
```

## pain_index {#pain_index}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/PainIndex.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("PainIndex"))
for (i in 1:portfolio_length) {
    result <- PainIndex(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## pain_ratio {#pain_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/PainRatio.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("PainRatio"))
for (rf in seq(0.0, 0.1, by=0.02)) {
    print(paste("Rf: ", rf))
    # first value (i=1) is always None because R code calculates periodicity
    for (i in 2:portfolio_length) {
        result <- PainRatio(portfolio_bacon[1:i, 1], Rf=rf)
        write.csv(result)
    }
}
```

## ulcer_index {#ulcer_index}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UlcerIndex.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("UlcerIndex"))
for (i in 1:portfolio_length) {
    result <- UlcerIndex(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## martin_ratio {#martin_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/MartinRatio.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("MartinRatio"))
for (rf in seq(0.0, 0.1, by=0.02)) {
    print(paste("Rf: ", rf))
    # first value (i=1) is always None because R code calculates periodicity
    for (i in 2:portfolio_length) {
        result <- MartinRatio(portfolio_bacon[1:i, ], Rf=rf)
        write.csv(result)
    }
}
```

## drawdown_average {#drawdown_average}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/maxDrawdown.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("AverageDrawdown"))
for (i in 1:portfolio_length) {
    result <- AverageDrawdown(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## drawdown_average_length {#drawdown_average_length}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/maxDrawdown.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("AverageLength"))
for (i in 1:portfolio_length) {
    result <- AverageLength(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## drawdown_average_peak_to_trough {#drawdown_average_peak_to_trough}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/maxDrawdown.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("AveragePeakToTrough"))
for (i in 1:portfolio_length) {
    result <- AveragePeakToTrough(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## drawdown_average_recovery {#drawdown_average_recovery}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/maxDrawdown.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("AverageRecovery"))
for (i in 1:portfolio_length) {
    result <- AverageRecovery(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## drawdown_deviation {#drawdown_deviation}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/maxDrawdown.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("DrawdownDeviation"))
for (i in 1:portfolio_length) {
    result <- DrawdownDeviation(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## cdar {#cdar_average}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/maxDrawdown.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("CDD"))
for (method in c("discrete", "average")) {
    for (geometric in c(TRUE)) {
        for (p in c(0.90, 0.95, 0.975, 0.99, 0.995, 0.999)) {
            print(paste("geometric:", geometric, "method", method "p:", p))
            for (i in 1:portfolio_length) {
                result <- CDD(portfolio_bacon[1:i, 1], method=method, p=p, geometric=geometric, invert=TRUE)
                write.csv(result)
            }
        }
    }
}
```

## cdar_beta {#cdar_beta}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/CDaR.beta.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("CDaR.beta"))
for (geometric in c(TRUE, FALSE)) {
    for (p in c(0.90, 0.95, 0.975, 0.99, 0.995, 0.999)) {
        print(paste("geometric:", geometric, "p:", p))
        for (i in 1:portfolio_length) {
            result <- CDaR.beta(portfolio_bacon[1:i, 1], Rm=portfolio_bacon[1:i, 2], p=p, geometric=geometric)
            write.csv(result)
        }
    }
}
```

## cdar_alpha {#cdar_alpha}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/CDaR.alpha.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("CDaR.alpha"))
for (geometric in c(TRUE, FALSE)) {
    for (p in c(0.90, 0.95, 0.975, 0.99, 0.995, 0.999)) {
        print(paste("geometric:", geometric, "p:", p))
        for (i in 1:portfolio_length) {
            result <- CDaR.alpha(portfolio_bacon[1:i, 1], Rm=portfolio_bacon[1:i, 2], p=p, geometric=geometric)
            write.csv(result)
        }
    }
}
```

## sfm_risk_premium {#sfm_risk_premium}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/CAPM.utils.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("CAPM.RiskPremium"))
for (rf in c(-0.01, 0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- CAPM.RiskPremium(Ra=portfolio_bacon[1:i, 1], Rf=rf)
        write.csv(result)
    }
}
```

## sfm_alpha {#sfm_alpha}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/CAPM.alpha.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("CAPM.alpha"))
for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("method=LS Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- CAPM.alpha(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, digits=15, method="LS")
        write.csv(result)
    }
}
```

## sfm_beta {#sfm_beta}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/CAPM.beta.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("CAPM.beta"))
for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("method=LS Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- CAPM.beta(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, digits=15, method="LS")
        write.csv(result)
    }
}
```

## sfm_beta_bull {#sfm_beta_bull}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/CAPM.beta.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("CAPM.beta.bull"))
# c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)
# Rf >= 0.002 causes error in lm.fit()
for (rf in c(0, 0.001, 0.0015, 0.0019)) {
    print(paste("method=LS Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- CAPM.beta.bull(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, digits=15, method="LS")
        write.csv(result)
    }
}
```

## sfm_beta_bear {#sfm_beta_bear}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/CAPM.beta.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("CAPM.beta.bear"))
# c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)
# Rf >= 0.002 causes error in lm.fit()
for (rf in c(0, 0.001, 0.0015, 0.0019)) {
    print(paste("method=LS Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- CAPM.beta.bear(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, digits=15, method="LS")
        write.csv(result)
    }
}
```

## timing_ratio {#timing_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/CAPM.beta.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("TimingRatio"))
for (rf in c(0.0, 0.005, 0.01, 0.05, 0.1)) {
    print(paste("Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- TimingRaatio(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf)
        write.csv(result)
    }
}
```

## sfm_r2 {#sfm_r2}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/table.CAPM.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("table.CAPM"))
for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05)) {
    print(paste("Rf:", rf))
    for (i in 16:portfolio_length) {
        print(paste("i", i, "Rf:", rf))
        result <- table.CAPM(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, digits=15)
        write.csv(result)
    }
}
```

## jensen_alpha {#jensen_alpha}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/CAPM.jensenAlpha.R)

```R
################################################################
# Use both yearly and daily dates
################################################################
print(paste("CAPM.jensenAlpha, not annualized -- run on YEARLY dates "))
for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("method=LS Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- CAPM.jensenAlpha(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, digits=15, method="LS")
        write.csv(result)
    }
}
print(paste("CAPM.jensenAlpha, annualized -- run on DAILY dates "))
for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("not annualized, method=LS Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- CAPM.jensenAlpha(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, digits=15, method="LS")
        write.csv(result)
    }
}
```

## fama_beta {#fama_beta}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/FamaBeta.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("FamaBeta"))
for (i in 1:portfolio_length) {
    result <- FamaBeta(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], digits=15)
    write.csv(result)
}
```

## modigliani {#modigliani}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/Modigliani.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("Modigliani"))
for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- Modigliani(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, digits=15)
        write.csv(result)
    }
}
```

## tracking_error {#tracking_error}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/TrackingError.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("TrackingError"))
for (i in 1:portfolio_length) {
    result <- TrackingError(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2])
    write.csv(result)
}
```

## active_premium {#active_premium}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/ActivePremium.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("ActivePremium"))
for (i in 1:portfolio_length) {
    result <- ActivePremium(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2])
    write.csv(result)
}
```

## information_ratio {#information_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/InformationRatio.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("InformationRatio"))
for (i in 2:portfolio_length) {
    result <- InformationRatio(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2])
    write.csv(result)
}
```

## systematic_risk {#systematic_risk}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/SystematicRisk.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("SystematicRisk"))
for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- SystematicRisk(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf)
        write.csv(result)
    }
}
```

## treynor_ratio {#treynor_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/TreynorRatio.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("TreynorRatio"))
for (modified in c(FALSE, TRUE)) {
    for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
        print(paste("modified:", modified, "Rf:", rf))
        for (i in 1:portfolio_length) {
            result <- TreynorRatio(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, modified=FALSE)
            write.csv(result)
        }
    }
}
```

## specific_risk {#specific_risk}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/SpecificRisk.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("SpecificRisk"))
for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- SpecificRisk(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf)
        write.csv(result)
    }
}
```

## total_risk {#total_risk}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/TotalRisk.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("TotalRisk"))
for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("Rf:", rf))
    for (i in 1:portfolio_length) {
        result <- TotalRisk(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf)
        write.csv(result)
    }
}
```

## appraisal_ratio {#appraisal_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/AppraisalRatio.R)

```R
################################################################
# Use all three yearly/monthly/daily dates
################################################################
print(paste("AppraisalRatio"))
for (method in c("appraisal", "modified", "alternative")) { # c("appraisal", "modified", "alternative")
    for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
        print(paste("periodicity:", periodicity, "method:", method, "Rf:", rf))
        for (i in 1:portfolio_length) {
            result <- AppraisalRatio(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, method=method, digits=17)
            write.csv(result)
        }
    }
}
```

## m_squared {#m_squared}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/MSquared.R)

```R
################################################################
# Use all three yearly/monthly/daily dates
################################################################
print(paste("MSquared"))
for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("periodicity:", periodicity, "Rf:", rf))
    for (i in 2:portfolio_length) {
        result <- MSquared(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, digits=17)
        write.csv(result)
    }
}
```

## m_squared_excess {#m_squared_excess}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/MSquaredExcess.R)

```R
################################################################
# Use all three yearly/monthly/daily dates
################################################################
print(paste("MSquaredExcess"))
for (method in c("geometric", "arithmetic")) { # c("geometric", "arithmetic")
    for (rf in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
        print(paste("periodicity:", periodicity, "Rf:", rf))
        for (i in 2:portfolio_length) {
            result <- MSquaredExcess(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, Method=method, digits=17)
            write.csv(result)
        }
    }
}
```

## m_squared_sortino {#m_squared_sortino}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/M2Sortino.R)

```R
################################################################
# Use all three yearly/monthly/daily dates
################################################################
print(paste("M2Sortino"))
for (mar in c(0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("periodicity:", periodicity, "MAR:", mar))
    for (i in 1:portfolio_length) {
        result <- M2Sortino(Ra=portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], MAR=mar, digits=17)
        write.csv(result)
    }
}
```

## prospect_ratio {#prospect_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/ProspectRatio.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("ProspectRatio"))
for (mar in c(-0.01, 0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
    print(paste("periodicity:", periodicity, "MAR:", mar))
    for (i in 1:portfolio_length) {
        result <- ProspectRatio(portfolio_bacon[1:i, 1], MAR=mar, digits=17)
        write.csv(result)
    }
}
```

## kelly_ratio {#kelly_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/KellyRatio.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("KellyRatio"))
for (method in c("half", "full") {
    for (mar in c(-0.01, 0.0, 0.001, 0.005, 0.01, 0.05, 0.1, 0.3)) {
        print(paste("periodicity:", periodicity, "method:", method, "Rf:", rf))
        for (i in 1:portfolio_length) {
            result <- KellyRatio(portfolio_bacon[1:i, 1], Rf=rf, method=method, digits=17)
            write.csv(result)
        }
    }
}
```

## hurst_exponent {#hurst_exponent}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/HurstIndex.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("HurstIndex"))
for (i in 1:portfolio_length) {
    result <- HurstIndex(portfolio_bacon[1:i, 1])
    write.csv(result)
}
```

## upside_capture_ratio {#upside_capture_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("UpDownRatios"))
for (geometric in c(FALSE, TRUE)) {
    print(paste("geometric:", geometric, "method: Capture side: Up"))
    # first value (i=1) is None, it is used to detect periodicity
    for (i in 2:portfolio_length) {
        result <- UpDownRatios(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], method="Capture", side="Up", geometric=geometric)
        write.csv(result)
    }
}
```

## downside_capture_ratio {#downside_capture_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("UpDownRatios"))
for (geometric in c(FALSE, TRUE)) {
    print(paste("geometric:", geometric, "method: Capture side: Down"))
    # first 3 values (i=1...3) produce errors, so we assume they are None.
    for (i in 4:portfolio_length) {
        result <- UpDownRatios(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], method="Capture", side="Down", geometric=geometric)
        write.csv(result)
    }
}
```

## up_number_ratio {#up_number_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("UpDownRatios"))
print(paste("method: Number side: Up"))
for (i in 1:portfolio_length) {
    result <- UpDownRatios(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], method="Number", side="Up")
    write.csv(result)
}
```

## down_number_ratio {#down_number_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("UpDownRatios"))
print(paste("method: Number side: Down"))
for (i in 1:portfolio_length) {
    result <- UpDownRatios(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], method="Number", side="Down")
    write.csv(result)
}
```

## up_percentage_ratio {#up_percentage_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("UpDownRatios"))
print(paste("method: Percent side: Up"))
for (i in 1:portfolio_length) {
    result <- UpDownRatios(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], method="Percent", side="Up")
    write.csv(result)
}
```

## down_percentage_ratio {#down_percentage_ratio}

[R source](https://github.com/braverock/PerformanceAnalytics/blob/master/R/UpDownRatios.R)

```R
################################################################
# Use yearly dates
################################################################
print(paste("UpDownRatios"))
print(paste("method: Percent side: Down"))
for (i in 1:portfolio_length) {
    result <- UpDownRatios(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], method="Percent", side="Down")
    write.csv(result)
}
```
