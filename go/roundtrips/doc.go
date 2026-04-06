// Package roundtrips implements trading round-trip tracking with comprehensive
// performance statistics.
//
// A round-trip represents a complete trade cycle: opening a position (entry)
// and closing it (exit). The package computes 19 properties per round-trip
// including PnL, MAE/MFE, and efficiency metrics, plus 100+ aggregate
// performance statistics via RoundtripPerformance.
package roundtrips
