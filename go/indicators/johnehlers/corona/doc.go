// Package corona implements the Corona spectral-analysis engine used by the
// Ehlers Corona Charts suite:
//
//   - CoronaSpectrum
//   - CoronaSignalToNoiseRatio
//   - CoronaSwingPosition
//   - CoronaTrendVigor
//
// The engine performs:
//
//  1. High-pass detrending (cutoff = highPassFilterCutoff, default 30).
//  2. 6-tap FIR smoothing with weights {1, 2, 3, 3, 2, 1} / 12.
//  3. A bank of 2-pole bandpass filters covering cycle periods
//     [minimalPeriod, maximalPeriod] (default [6, 30]) at half-period
//     resolution (internally indexed in units of 2*period so each integer
//     index represents a half-period step).
//  4. Per-bin amplitude-squared and decibel (dB) normalization with EMA
//     smoothing.
//  5. Weighted center-of-gravity estimate of the dominant cycle period and
//     a 5-sample median of that estimate.
//
// Reference:
//
//	Ehlers, John F. "Measuring Cycle Periods", Technical Analysis of Stocks
//	and Commodities, November 2008, and companion Trader Tips pieces from
//	eSignal, NinjaTrader and Wealth-Lab (same issue).
//
// The package exposes the Corona engine as a helper -- it is not registered
// as an indicator type on its own.
package corona
