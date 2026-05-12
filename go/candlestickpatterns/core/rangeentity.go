// Package core provides the internal types and helpers for the candlestick
// pattern recognition engine. Pattern implementations in the patterns/
// subpackage operate on the exported Engine struct defined here.
package core

// RangeEntity identifies which part of a candlestick is used for range comparisons.
type RangeEntity int

const (
	// RealBody identifies the length of the real body of a candlestick.
	RealBody RangeEntity = iota
	// HighLow identifies the length of the high-low range of a candlestick.
	HighLow
	// Shadows identifies the length of the shadows of a candlestick.
	Shadows
)
