package core

// Candlestick primitives: color, body, shadow, gap, and enclosure tests.
//
// These are pure functions operating on OHLC values.

// ---------------------------------------------------------------------------
// Color
// ---------------------------------------------------------------------------

// IsWhite returns true when a candlestick is white (bullish): close >= open.
func IsWhite(o, c float64) bool {
	return c >= o
}

// IsBlack returns true when a candlestick is black (bearish): close < open.
func IsBlack(o, c float64) bool {
	return c < o
}

// ---------------------------------------------------------------------------
// Real body
// ---------------------------------------------------------------------------

// RealBodyLen returns the absolute length of the real body.
func RealBodyLen(o, c float64) float64 {
	if c >= o {
		return c - o
	}
	return o - c
}

// WhiteRealBody returns the length of the real body of a white candlestick (close - open).
func WhiteRealBody(o, c float64) float64 {
	return c - o
}

// BlackRealBody returns the length of the real body of a black candlestick (open - close).
func BlackRealBody(o, c float64) float64 {
	return o - c
}

// ---------------------------------------------------------------------------
// Shadows
// ---------------------------------------------------------------------------

// UpperShadow returns the length of the upper shadow.
func UpperShadow(o, h, c float64) float64 {
	if c >= o {
		return h - c
	}
	return h - o
}

// LowerShadow returns the length of the lower shadow.
func LowerShadow(o, l, c float64) float64 {
	if c >= o {
		return o - l
	}
	return c - l
}

// WhiteUpperShadow returns the length of the upper shadow of a white candlestick.
func WhiteUpperShadow(h, c float64) float64 {
	return h - c
}

// BlackUpperShadow returns the length of the upper shadow of a black candlestick.
func BlackUpperShadow(o, h float64) float64 {
	return h - o
}

// WhiteLowerShadow returns the length of the lower shadow of a white candlestick.
func WhiteLowerShadow(o, l float64) float64 {
	return o - l
}

// BlackLowerShadow returns the length of the lower shadow of a black candlestick.
func BlackLowerShadow(l, c float64) float64 {
	return c - l
}

// ---------------------------------------------------------------------------
// Gap tests
// ---------------------------------------------------------------------------

// IsRealBodyGapUp returns true when max(open1, close1) < min(open2, close2).
func IsRealBodyGapUp(o1, c1, o2, c2 float64) bool {
	return max(o1, c1) < min(o2, c2)
}

// IsRealBodyGapDown returns true when min(open1, close1) > max(open2, close2).
func IsRealBodyGapDown(o1, c1, o2, c2 float64) bool {
	return min(o1, c1) > max(o2, c2)
}

// IsHighLowGapUp returns true when high of first candle < low of second candle.
func IsHighLowGapUp(h1, l2 float64) bool {
	return h1 < l2
}

// IsHighLowGapDown returns true when low of first candle > high of second candle.
func IsHighLowGapDown(l1, h2 float64) bool {
	return l1 > h2
}

// ---------------------------------------------------------------------------
// Enclosure tests
// ---------------------------------------------------------------------------

// IsRealBodyEnclosesRealBody returns true when the real body of candle 1
// completely encloses the real body of candle 2.
func IsRealBodyEnclosesRealBody(o1, c1, o2, c2 float64) bool {
	var min1, max1 float64
	if c1 > o1 {
		min1, max1 = o1, c1
	} else {
		min1, max1 = c1, o1
	}
	var min2, max2 float64
	if c2 > o2 {
		min2, max2 = o2, c2
	} else {
		min2, max2 = c2, o2
	}
	return max1 > max2 && min1 < min2
}

// IsRealBodyEnclosesOpen returns true when the real body of candle 1
// encloses the open of candle 2.
func IsRealBodyEnclosesOpen(o1, c1, o2 float64) bool {
	if o1 > c1 {
		return o2 < o1 && o2 > c1
	}
	return o2 > o1 && o2 < c1
}

// IsRealBodyEnclosesClose returns true when the real body of candle 1
// encloses the close of candle 2.
func IsRealBodyEnclosesClose(o1, c1, c2 float64) bool {
	if o1 > c1 {
		return c2 < o1 && c2 > c1
	}
	return c2 > o1 && c2 < c1
}

// ---------------------------------------------------------------------------
// Misc comparisons
// ---------------------------------------------------------------------------

// IsHighExceedsClose returns true when high of candle 1 > close of candle 2.
func IsHighExceedsClose(h1, c2 float64) bool {
	return h1 > c2
}

// IsOpensWithin returns true when candle 1 opens within the real body of candle 2
// (with optional tolerance).
func IsOpensWithin(o1, o2, c2, tolerance float64) bool {
	return o1 >= min(o2, c2)-tolerance && o1 <= max(o2, c2)+tolerance
}

// ---------------------------------------------------------------------------
// Range value for a single candle (used by Criterion)
// ---------------------------------------------------------------------------

// CandleRangeValue computes the range value of a candle for a given RangeEntity type.
func CandleRangeValue(entity RangeEntity, o, h, l, c float64) float64 {
	switch entity {
	case RealBody:
		if c >= o {
			return c - o
		}
		return o - c
	case HighLow:
		return h - l
	default:
		// SHADOWS: average of upper and lower shadow
		if c >= o {
			return (h - c + o - l) / 2.0
		}
		return (h - o + c - l) / 2.0
	}
}
