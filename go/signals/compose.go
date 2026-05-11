package signals

import "zpano/fuzzy"

// SignalAnd combines signals with product t-norm (fuzzy AND).
// All signals must be high for the result to be high.
func SignalAnd(signals ...float64) float64 {
	return fuzzy.TProductAll(signals...)
}

// SignalOr combines two signals with probabilistic s-norm (fuzzy OR).
// Result is high when either signal is high. Equivalent to a + b - a*b.
func SignalOr(a, b float64) float64 {
	return fuzzy.SProbabilistic(a, b)
}

// SignalNot negates a signal (fuzzy complement). Returns 1 - signal.
func SignalNot(signal float64) float64 {
	return fuzzy.FNot(signal)
}

// SignalStrength filters weak signals below minStrength to zero.
// Signals at or above the threshold pass through unchanged.
func SignalStrength(signal, minStrength float64) float64 {
	if signal >= minStrength {
		return signal
	}
	return 0.0
}
