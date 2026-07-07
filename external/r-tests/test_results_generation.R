# https://www.rdocumentation.org/packages/PerformanceAnalytics/versions/2.0.4/topics/portfolio_bacon
# https://www.datacamp.com/datalab/w/28c21593-21e6-47d9-8e72-acebdd3be32c/edit
# https://www.datacamp.com/datalab/w/28c21593-21e6-47d9-8e72-acebdd3be32c/edit#9c513e64-2a46-4c6b-9b59-9567e18e8229
# https://www.datacamp.com/datalab/w/28c21593-21e6-47d9-8e72-acebdd3be32c/edit

if(!require('PerformanceAnalytics')) {
    install.packages('PerformanceAnalytics')
    library('PerformanceAnalytics')
}

data(portfolio_bacon)
head(portfolio_bacon, 100)
write.csv(portfolio_bacon)

portfolio_length <- nrow(portfolio_bacon)
print(portfolio_length)

################################################################
# TEST DATA GENERATION FOR MISSING RATIOS (❌)
################################################################

# =========================================================================
# TrackingError (TrackingError.R → tracking_error)
# Data frequency expected: Monthly (auto-detected by periodicity())
# =========================================================================
for (i in 1:portfolio_length) {
    result <- TrackingError(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2])
    write.csv(result)
}
for (i in 1:portfolio_length) {
    result <- TrackingError(portfolio_bacon[1:i, 2], Rb=portfolio_bacon[1:i, 1])
    write.csv(result)
}

# =========================================================================
# CAPM.beta (CAPM.beta.R → beta)
# Data frequency expected: Monthly (Rf should be monthly rate)
# =========================================================================
for (rf in seq(0, 0.3, by=0.05)) {
    print(paste("Rf: ", rf))
    for (i in 1:portfolio_length) {
        result <- CAPM.beta(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf)
        write.csv(result)
    }
}
for (rf in seq(0, 0.3, by=0.05)) {
    print(paste("Rf: ", rf))
    for (i in 1:portfolio_length) {
        result <- CAPM.beta(portfolio_bacon[1:i, 2], Rb=portfolio_bacon[1:i, 1], Rf=rf)
        write.csv(result)
    }
}

# =========================================================================
# CAPM.jensenAlpha (CAPM.jensenAlpha.R → alpha)
# Data frequency expected: Monthly (Rf should be monthly rate)
# =========================================================================
for (rf in seq(0, 0.3, by=0.05)) {
    print(paste("Rf: ", rf))
    for (i in 1:portfolio_length) {
        result <- CAPM.jensenAlpha(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf)
        write.csv(result)
    }
}
for (rf in seq(0, 0.3, by=0.05)) {
    print(paste("Rf: ", rf))
    for (i in 1:portfolio_length) {
        result <- CAPM.jensenAlpha(portfolio_bacon[1:i, 2], Rb=portfolio_bacon[1:i, 1], Rf=rf)
        write.csv(result)
    }
}

# =========================================================================
# TreynorRatio (TreynorRatio.R → treynor_ratio)
# Data frequency expected: Monthly (auto-detected by periodicity())
# Parameters: modified (FALSE/TRUE)
# =========================================================================
for (modified in c(FALSE, TRUE)) {
    for (rf in seq(0, 0.3, by=0.05)) {
        print(paste("modified: ", modified, "Rf: ", rf))
        for (i in 1:portfolio_length) {
            result <- TreynorRatio(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, modified=modified)
            write.csv(result)
        }
    }
}
for (modified in c(FALSE, TRUE)) {
    for (rf in seq(0, 0.3, by=0.05)) {
        print(paste("modified: ", modified, "Rf: ", rf))
        for (i in 1:portfolio_length) {
            result <- TreynorRatio(portfolio_bacon[1:i, 2], Rb=portfolio_bacon[1:i, 1], Rf=rf, modified=modified)
            write.csv(result)
        }
    }
}

# =========================================================================
# AppraisalRatio (AppraisalRatio.R → appraisal_ratio)
# Data frequency expected: Monthly (Rf should be monthly rate)
# Parameters: method ("appraisal", "modified", "alternative")
# =========================================================================
for (method in c("appraisal", "modified", "alternative")) {
    for (rf in seq(0, 0.3, by=0.05)) {
        print(paste("method: ", method, "Rf: ", rf))
        for (i in 1:portfolio_length) {
            result <- AppraisalRatio(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf, method=method)
            write.csv(result)
        }
    }
}
for (method in c("appraisal", "modified", "alternative")) {
    for (rf in seq(0, 0.3, by=0.05)) {
        print(paste("method: ", method, "Rf: ", rf))
        for (i in 1:portfolio_length) {
            result <- AppraisalRatio(portfolio_bacon[1:i, 2], Rb=portfolio_bacon[1:i, 1], Rf=rf, method=method)
            write.csv(result)
        }
    }
}

# =========================================================================
# FamaBeta (FamaBeta.R → fama_beta)
# Data frequency expected: Monthly (uses Frequency() internally)
# =========================================================================
for (i in 1:portfolio_length) {
    result <- FamaBeta(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2])
    write.csv(result)
}
for (i in 1:portfolio_length) {
    result <- FamaBeta(portfolio_bacon[1:i, 2], Rb=portfolio_bacon[1:i, 1])
    write.csv(result)
}

# =========================================================================
# DownsideDeviation (DownsideDeviation.R → downside_deviation)
# Data frequency expected: Monthly (MAR should be monthly rate)
# Parameters: MAR, method ("full", "subset")
# =========================================================================
for (method in c("full", "subset")) {
    for (mar in seq(0, 0.1, by=0.02)) {
        print(paste("method: ", method, "MAR: ", mar))
        for (i in 1:portfolio_length) {
            result <- DownsideDeviation(portfolio_bacon[1:i, ], MAR=mar, method=method)
            write.csv(result)
        }
    }
}

# =========================================================================
# DownsidePotential (DownsideDeviation.R → downside_potential)
# Data frequency expected: Monthly (MAR should be monthly rate)
# =========================================================================
for (mar in seq(0, 0.1, by=0.02)) {
    print(paste("MAR: ", mar))
    for (i in 1:portfolio_length) {
        result <- DownsidePotential(portfolio_bacon[1:i, ], MAR=mar)
        write.csv(result)
    }
}

# =========================================================================
# SemiDeviation (SemiDeviation.R → semi_deviation)
# Data frequency expected: Monthly
# =========================================================================
for (i in 1:portfolio_length) {
    result <- SemiDeviation(portfolio_bacon[1:i, ])
    write.csv(result)
}

# =========================================================================
# VolatilitySkewness (VolatilitySkewness.R → volatility_skewness / variability)
# Data frequency expected: Monthly (MAR should be monthly rate)
# Parameters: MAR, stat ("volatility", "variability")
# =========================================================================
for (stat in c("volatility", "variability")) {
    for (mar in seq(0, 0.1, by=0.02)) {
        print(paste("stat: ", stat, "MAR: ", mar))
        for (i in 1:portfolio_length) {
            result <- VolatilitySkewness(portfolio_bacon[1:i, ], MAR=mar, stat=stat)
            write.csv(result)
        }
    }
}

# =========================================================================
# DownsideFrequency (DownsideFrequency.R → downside_frequency)
# Data frequency expected: Monthly (MAR should be monthly rate)
# =========================================================================
for (mar in seq(0, 0.1, by=0.02)) {
    print(paste("MAR: ", mar))
    for (i in 1:portfolio_length) {
        result <- DownsideFrequency(portfolio_bacon[1:i, ], MAR=mar)
        write.csv(result)
    }
}

# =========================================================================
# HurstIndex (HurstIndex.R → hurst_index)
# Data frequency expected: Monthly
# =========================================================================
for (i in 1:portfolio_length) {
    result <- HurstIndex(portfolio_bacon[1:i, ])
    write.csv(result)
}

# =========================================================================
# VaR (VaR.R → var)
# Data frequency expected: Monthly (no auto-annualization for single method)
# Parameters: p (0.95, 0.99), method ("modified", "gaussian", "historical")
# =========================================================================
for (method in c("modified", "gaussian", "historical")) {
    for (p in c(0.95, 0.99)) {
        print(paste("method: ", method, "p: ", p))
        for (i in 1:portfolio_length) {
            result <- VaR(portfolio_bacon[1:i, ], p=p, method=method, portfolio_method="single")
            write.csv(result)
        }
    }
}

# =========================================================================
# ES (ES.R → cvar)
# Data frequency expected: Monthly
# Parameters: p (0.95, 0.99), method ("modified", "gaussian", "historical"), invert=TRUE, operational=TRUE
# =========================================================================
for (method in c("modified", "gaussian", "historical")) {
    for (p in c(0.95, 0.99)) {
        print(paste("method: ", method, "p: ", p))
        for (i in 1:portfolio_length) {
            result <- ES(portfolio_bacon[1:i, ], p=p, method=method, portfolio_method="single", invert=TRUE, operational=TRUE)
            write.csv(result)
        }
    }
}

# =========================================================================
# Modigliani (Modigliani.R / MSquared.R → modigliani_modigliani)
# Data frequency expected: Monthly (Rf should be monthly rate)
# =========================================================================
for (rf in seq(0, 0.3, by=0.05)) {
    print(paste("Rf: ", rf))
    for (i in 1:portfolio_length) {
        result <- Modigliani(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], Rf=rf)
        write.csv(result)
    }
}
for (rf in seq(0, 0.3, by=0.05)) {
    print(paste("Rf: ", rf))
    for (i in 1:portfolio_length) {
        result <- Modigliani(portfolio_bacon[1:i, 2], Rb=portfolio_bacon[1:i, 1], Rf=rf)
        write.csv(result)
    }
}

# =========================================================================
# OmegaSharpeRatio (OmegaSharpeRatio.R → omega_sharpe_ratio)
# Data frequency expected: Monthly (MAR should be monthly rate)
# =========================================================================
for (mar in seq(0, 0.1, by=0.02)) {
    print(paste("MAR: ", mar))
    for (i in 1:portfolio_length) {
        result <- OmegaSharpeRatio(portfolio_bacon[1:i, ], MAR=mar)
        write.csv(result)
    }
}

# =========================================================================
# RachevRatio (RachevRatio.R → rachev_ratio)
# Data frequency expected: Monthly (rf should be monthly rate)
# Parameters: alpha (0.05, 0.1, 0.15), beta (0.05, 0.1, 0.15), rf
# =========================================================================
for (alpha in c(0.05, 0.1, 0.15)) {
    for (beta in c(0.05, 0.1, 0.15)) {
        for (rf in seq(0, 0.1, by=0.02)) {
            print(paste("alpha: ", alpha, "beta: ", beta, "rf: ", rf))
            for (i in 1:portfolio_length) {
                result <- RachevRatio(portfolio_bacon[1:i, ], alpha=alpha, beta=beta, rf=rf)
                write.csv(result)
            }
        }
    }
}

# =========================================================================
# CDAR.alpha (CDAR.alpha.R → cdar_alpha)
# Data frequency expected: Monthly (annualizes with ^12 internally)
# Parameters: p (0.95, 0.99), type (NULL, "average", "max"), geometric (TRUE)
# =========================================================================
for (p in c(0.95, 0.99)) {
    for (type in c(NULL, "average", "max")) {
        type_str <- ifelse(is.null(type), "NULL", type)
        print(paste("p: ", p, "type: ", type_str))
        for (i in 2:portfolio_length) {  # first value may be NA
            result <- CDaR.alpha(portfolio_bacon[1:i, 1], Rm=portfolio_bacon[1:i, 2], p=p, type=type, geometric=TRUE)
            write.csv(result)
        }
    }
}
for (p in c(0.95, 0.99)) {
    for (type in c(NULL, "average", "max")) {
        type_str <- ifelse(is.null(type), "NULL", type)
        print(paste("p: ", p, "type: ", type_str))
        for (i in 2:portfolio_length) {  # first value may be NA
            result <- CDaR.alpha(portfolio_bacon[1:i, 2], Rm=portfolio_bacon[1:i, 1], p=p, type=type, geometric=TRUE)
            write.csv(result)
        }
    }
}

# =========================================================================
# CDAR.beta (CDAR.beta.R → cdar_beta)
# Data frequency expected: Monthly (annualizes with ^12 internally)
# Parameters: p (0.95, 0.99), type (NULL, "average", "max"), geometric (TRUE)
# =========================================================================
for (p in c(0.95, 0.99)) {
    for (type in c(NULL, "average", "max")) {
        type_str <- ifelse(is.null(type), "NULL", type)
        print(paste("p: ", p, "type: ", type_str))
        for (i in 2:portfolio_length) {  # first value may be NA
            result <- CDaR.beta(portfolio_bacon[1:i, 1], Rm=portfolio_bacon[1:i, 2], p=p, type=type, geometric=TRUE)
            write.csv(result)
        }
    }
}
for (p in c(0.95, 0.99)) {
    for (type in c(NULL, "average", "max")) {
        type_str <- ifelse(is.null(type), "NULL", type)
        print(paste("p: ", p, "type: ", type_str))
        for (i in 2:portfolio_length) {  # first value may be NA
            result <- CDaR.beta(portfolio_bacon[1:i, 2], Rm=portfolio_bacon[1:i, 1], p=p, type=type, geometric=TRUE)
            write.csv(result)
        }
    }
}

# =========================================================================
# UpDownRatios (UpDownRatios.R → upside_capture, downside_capture, capture_ratio)
# Data frequency expected: Monthly
# Parameters: method ("Capture", "Number", "Percent"), side ("Up", "Down"), geometric (TRUE)
# =========================================================================
for (method in c("Capture", "Number", "Percent")) {
    for (side in c("Up", "Down")) {
        print(paste("method: ", method, "side: ", side))
        for (i in 1:portfolio_length) {
            result <- UpDownRatios(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], method=method, side=side, geometric=TRUE)
            write.csv(result)
        }
    }
}
for (method in c("Capture", "Number", "Percent")) {
    for (side in c("Up", "Down")) {
        print(paste("method: ", method, "side: ", side))
        for (i in 1:portfolio_length) {
            result <- UpDownRatios(portfolio_bacon[1:i, 2], Rb=portfolio_bacon[1:i, 1], method=method, side=side, geometric=TRUE)
            write.csv(result)
        }
    }
}

# =========================================================================
# OmegaExcessReturn (OmegaExcessReturn.R → omega_excess_return)
# Data frequency expected: Monthly (MAR should be monthly rate)
# =========================================================================
for (mar in seq(0, 0.1, by=0.02)) {
    print(paste("MAR: ", mar))
    for (i in 1:portfolio_length) {
        result <- OmegaExcessReturn(portfolio_bacon[1:i, 1], Rb=portfolio_bacon[1:i, 2], MAR=mar)
        write.csv(result)
    }
}
for (mar in seq(0, 0.1, by=0.02)) {
    print(paste("MAR: ", mar))
    for (i in 1:portfolio_length) {
        result <- OmegaExcessReturn(portfolio_bacon[1:i, 2], Rb=portfolio_bacon[1:i, 1], MAR=mar)
        write.csv(result)
    }
}

# =========================================================================
# skewness (skewness.R → skew)
# Data frequency expected: Monthly
# Parameters: method ("moment", "fisher", "sample")
# =========================================================================
for (method in c("moment", "fisher", "sample")) {
    print(paste("method: ", method))
    for (i in 1:portfolio_length) {
        result <- skewness(portfolio_bacon[1:i, ], method=method)
        write.csv(result)
    }
}

# =========================================================================
# SkewnessKurtosisRatio (SkewnessKurtosisRatio.R → skewness_kurtosis_ratio, bera_jarque_statistic)
# Data frequency expected: Monthly
# Also: is_normal_distribution for bera_jarque_statistic with confidence
# =========================================================================
for (i in 1:portfolio_length) {
    result <- SkewnessKurtosisRatio(portfolio_bacon[1:i, ])
    write.csv(result)
}
# Bera-Jarque statistic via is_normal_distribution
for (confidence in c(0.95, 0.99)) {
    print(paste("confidence: ", confidence))
    for (i in 1:portfolio_length) {
        bj_stat <- BeraJarqueStatistic(portfolio_bacon[1:i, ])  # if exists, or compute manually
        # The function is_normal_distribution returns boolean, we need the statistic
        # Use skewness and kurtosis directly
        # Actually SkewnessKurtosisRatio function uses moment method for both
        write.csv(bj_stat)
    }
}

# =========================================================================
# MeanAbsoluteDeviation (MeanAbsoluteDeviation.R → mad_ratio)
# Data frequency expected: Monthly
# =========================================================================
for (i in 1:portfolio_length) {
    result <- MeanAbsoluteDeviation(portfolio_bacon[1:i, ])
    write.csv(result)
}

# =========================================================================
# KellyRatio (KellyRatio.R → kelly_ratio)
# Data frequency expected: Monthly (Rf should be monthly rate)
# Parameters: Rf, method ("half", "full")
# =========================================================================
for (method in c("half", "full")) {
    for (rf in seq(0, 0.3, by=0.05)) {
        print(paste("method: ", method, "Rf: ", rf))
        for (i in 1:portfolio_length) {
            result <- KellyRatio(portfolio_bacon[1:i, ], Rf=rf, method=method)
            write.csv(result)
        }
    }
}

# =========================================================================
# ProbSharpeRatio (ProbSharpeRatio.R → probabilistic_sharpe_ratio)
# Data frequency expected: Monthly (Rf should be monthly rate)
# Parameters: Rf, refSR (0.1 to 0.5), p (0.95), ignore_skewness (FALSE), ignore_kurtosis (TRUE)
# =========================================================================
for (rf in seq(0, 0.1, by=0.02)) {
    for (refSR in seq(0.1, 0.5, by=0.1)) {
        print(paste("rf: ", rf, "refSR: ", refSR))
        for (i in 2:portfolio_length) {  # needs at least 2 observations
            result <- ProbSharpeRatio(portfolio_bacon[1:i, ], Rf=rf, refSR=refSR, p=0.95, ignore_skewness=FALSE, ignore_kurtosis=TRUE)
            write.csv(result$sr_prob)
            write.csv(result$sr_confidence_interval)
        }
    }
}

# =========================================================================
# ProspectRatio (ProspectRatio.R → prospect_ratio)
# Data frequency expected: Monthly (MAR should be monthly rate)
# =========================================================================
for (mar in seq(0, 0.1, by=0.02)) {
    print(paste("MAR: ", mar))
    for (i in 1:portfolio_length) {
        result <- ProspectRatio(portfolio_bacon[1:i, ], MAR=mar)
        write.csv(result)
    }
}

# =========================================================================
# DRatio (DRatio.R → d_ratio)
# Data frequency expected: Monthly
# =========================================================================
for (i in 1:portfolio_length) {
    result <- DRatio(portfolio_bacon[1:i, ])
    write.csv(result)
}

# =========================================================================
# LPM/HPM (lpm.R → lpm, hpm)
# Data frequency expected: Monthly (threshold should be monthly rate)
# Parameters: n (moment order 1,2,3), threshold (0, 0.01, 0.02), about_mean (FALSE)
# Note: lpm function in PerformanceAnalytics uses about_mean parameter
# =========================================================================
for (n in c(1, 2, 3)) {
    for (threshold in c(0, 0.01, 0.02)) {
        print(paste("n: ", n, "threshold: ", threshold))
        for (i in 1:portfolio_length) {
            result <- lpm(portfolio_bacon[1:i, ], n=n, threshold=threshold, about_mean=FALSE)
            write.csv(result)
        }
    }
}

################################################################
# ORIGINAL TEST DATA GENERATION (from initial.R)
################################################################

for(rf in seq(0, 0.3, by=0.05)) {
  print(paste("Risk-free rate: ", rf))
  for (i in 1:portfolio_length) {
    result <- AdjustedSharpeRatio(portfolio_bacon[1:i, ], Rf=rf)
    write.csv(result)
  }
}

for (fun in c("StdDev", "VaR", "ES", "SemiSD")) {
  for(rf in seq(0, 0.3, by=0.05)) {
    for (i in 1:portfolio_length) {
      result <- SharpeRatio(portfolio_bacon[1:i, ], Rf=rf, FUN=fun)
      write.csv(result)
    }
  }
}

for(rf in seq(0, 0.3, by=0.05)) {
  for (i in 1:portfolio_length) {
    result <- DownsideSharpeRatio(portfolio_bacon[1:i, ], Rf=rf)
    write.csv(result)
  }
}

for (i in 1:portfolio_length) {
    result <- BernardoLedoitRatio(portfolio_bacon[1:i, ])
    write.csv(result)
}

for (modified in c(TRUE, FALSE)) {
  for(rf in seq(0, 0.3, by=0.05)) {
    for (i in 2:portfolio_length) {
      result <- BurkeRatio(portfolio_bacon[1:i, ], Rf=rf, modified=modified)
      write.csv(result)
    }
  }
}

for(rf in seq(0, 0.3, by=0.05)) {
  for (i in portfolio_length:portfolio_length) {
    result <- Return.excess(portfolio_bacon[1:i,], Rf=rf)
    write.csv(result)
  }
}

for(geometric in c(TRUE, FALSE)) {
  print(geometric)
  for (i in 1:portfolio_length) {
    result <- Return.cumulative(portfolio_bacon[1:i,], geometric=geometric)
    write.csv(result)
  }
}

for(mar in seq(0, 0.3, by=0.05)) {
  for (i in 1:portfolio_length) {
    result <- SortinoRatio(portfolio_bacon[1:i,], MAR=mar)
    write.csv(result)
  }
}

for (l in seq(0.0, 0.1, by=0.02)) {
  for(rf in c(0)) {
    print(paste("L: ", l, "Risk-free rate: ", rf))
    for (i in 1:portfolio_length) {
      result <- Omega(portfolio_bacon[1:i, ], L=l, Rf=rf, method="simple", output="point")
      write.csv(result)
    }
  }
}

for (l in c(1,2,3,4)) {
  for(mar in seq(0, 0.3, by=0.05)) {
    print(paste("L: ", l, "MAR: ", mar))
    for (i in 1:portfolio_length) {
      result <- Kappa(portfolio_bacon[1:i, ], MAR=mar, l=l)
      write.csv(result)
    }
  }
}

for (i in 1:portfolio_length) {
  result <- InformationRatio(portfolio_bacon[1:i,1], Rb=portfolio_bacon[1:i,2])
  write.csv(result)
}
for (i in 1:portfolio_length) {
  result <- InformationRatio(portfolio_bacon[1:i,2], Rb=portfolio_bacon[1:i,1])
  write.csv(result)
}

for (method in c("excess", "moment", "fisher", "sample", "sample_excess")) {
  print(paste("method: ", method))
  for (i in 1:portfolio_length) {
    result <- kurtosis(portfolio_bacon[1:i, ], method=method)
    write.csv(result)
  }
}

for (method in c("full","subset")) {
  for(mar in seq(0, 0.1, by=0.02)) {
    print(paste("method: ", method, "MAR: ", mar))
    for (i in 1:portfolio_length) {
      result <- UpsidePotentialRatio(portfolio_bacon[1:i, ], MAR=mar, method=method)
      write.csv(result)
    }
  }
}

result <- Drawdowns(portfolio_bacon[,], geometric = TRUE)
write.csv(result)
result <- maxDrawdown(portfolio_bacon[,], geometric = TRUE)
write.csv(result)

for (i in 1:portfolio_length) {
    result <- CalmarRatio(portfolio_bacon[1:i, ], scale=1)
    write.csv(result)
}

for (excess in seq(0.0, 0.1, by=0.02)) {
  for (i in 1:portfolio_length) {
    result <- SterlingRatio(portfolio_bacon[1:i, ], excess=excess, scale=1)
    write.csv(result)
  }
}

################################################################
# make daily returns from monthly ones
################################################################
dates <- as.Date(index(portfolio_bacon))

# Generate a sequence of daily dates starting from the first date in the dataset
start_date <- dates[1]
end_date <- dates[length(dates)]
daily_dates <- seq.Date(from = start_date, to = end_date, by = "day")

# Ensure the number of daily dates matches the number of rows in the dataset
if (length(daily_dates) >= nrow(portfolio_bacon)) {
  daily_dates <- daily_dates[1:nrow(portfolio_bacon)]
} else {
  stop("The generated sequence of daily dates is shorter than the number of rows in the dataset.")
}

index(portfolio_bacon) <- daily_dates
head(portfolio_bacon, 100)
###################################################################

################################################################
# make yearly return dates from the monthly ones,
# because R code annualizes the returns
################################################################
dates <- as.Date(index(portfolio_bacon))

# Generate a sequence of yearly dates starting from the first date in the dataset
start_date <- dates[1]
end_date <- dates[length(dates)]
yearly_dates <- seq.Date(from = start_date, by = "year", length.out = nrow(portfolio_bacon))

# Ensure the number of yearly dates matches the number of rows in the dataset
if (length(yearly_dates) > nrow(portfolio_bacon)) {
  yearly_dates <- yearly_dates[1:nrow(portfolio_bacon)]
} else if (length(yearly_dates) < nrow(portfolio_bacon)) {
  stop("Not enough yearly dates to match the number of rows in the dataset.")
}

index(portfolio_bacon) <- yearly_dates
portfolio_length <- nrow(portfolio_bacon)
###################################################################

for (modified in c(FALSE,TRUE)) {
  for (rf in seq(0.0, 0.1, by=0.02)) {
    print(paste("modified: ", modified, "Rf: ", rf))
    for (i in 2:portfolio_length) {
      result <- BurkeRatio(portfolio_bacon[1:i, ], Rf=rf, modified=modified)
      write.csv(result)
    }
  }
}

for (rf in seq(0.0, 0.1, by=0.02)) {
  print(paste("Rf: ", rf))
  for (i in 2:portfolio_length) {
    result <- PainRatio(portfolio_bacon[1:i, ], Rf=rf)
    write.csv(result)
  }
}

for (rf in seq(0.0, 0.1, by=0.02)) {
  print(paste("Rf: ", rf))
  for (i in 2:portfolio_length) {
    result <- MartinRatio(portfolio_bacon[1:i, ], Rf=rf)
    write.csv(result)
  }
}

for (i in 1:portfolio_length) {
  result <- PainIndex(portfolio_bacon[1:i, ])
  write.csv(result)
}

for (i in 1:portfolio_length) {
  result <- UlcerIndex(portfolio_bacon[1:i, ])
  write.csv(result)
}

# https://www.datacamp.com/datalab/w/28c21593-21e6-47d9-8e72-acebdd3be32c/edit