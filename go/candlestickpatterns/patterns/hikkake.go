package patterns

import (
	"zpano/candlestickpatterns/core"
)

// Hikkake: a three-candle pattern with stateful confirmation.
//
// TA-Lib behavior:
// - Detection bar: outputs +100.0 (bullish) or -100.0 (bearish)
// - Confirmation bar (within 3 bars of detection): outputs +200.0 or -200.0
// - If a new hikkake is detected on the same bar as a confirmation,
//   the new hikkake takes priority.
//
// Must have:
// - first and second candle: inside bar (2nd lower high, higher low)
// - third candle: lower high AND lower low (bull) or higher high AND
//   higher low (bear)
//
// Confirmation: close > high of 2nd candle (bull) or close < low of
// 2nd candle (bear) within 3 bars.
func Hikkake(cp *core.CandlestickPatterns) float64 {
	if !cp.Enough(3) {
		return 0.0
	}

	b1 := cp.Bar(3)
	b2 := cp.Bar(2)
	b3 := cp.Bar(1)

	// Inside bar check.
	if b2.H < b1.H && b2.L > b1.L {
		// Bullish: 3rd has lower high AND lower low.
		if b3.H < b2.H && b3.L < b2.L {
			return 100.0
		}
		// Bearish: 3rd has higher high AND higher low.
		if b3.H > b2.H && b3.L > b2.L {
			return -100.0
		}
	}

	// No new pattern — check for confirmation of a recent hikkake.
	// Look back 1-3 bars for a hikkake pattern.
	for lookback := 1; lookback <= 3; lookback++ {
		n := 3 + lookback
		if !cp.Enough(n) {
			break
		}

		p1 := cp.Bar(n)     // 1st of pattern
		p2 := cp.Bar(n - 1) // inside bar (2nd)
		p3 := cp.Bar(n - 2) // breakout bar (3rd)

		// Must be a valid hikkake at that position.
		if !(p2.H < p1.H && p2.L > p1.L) {
			continue
		}

		patternResult := 0.0
		if p3.H < p2.H && p3.L < p2.L {
			patternResult = 100.0 // bullish
		} else if p3.H > p2.H && p3.L > p2.L {
			patternResult = -100.0 // bearish
		} else {
			continue
		}

		// Check that no intervening bar already confirmed or re-detected.
		// If there's a newer hikkake between the pattern and current bar,
		// the older one is superseded.
		superseded := false
		for gap := 1; gap < lookback; gap++ {
			gb := n - 2 - gap
			if gb < 1 {
				break
			}
			if cp.Enough(gb + 2) {
				ga := cp.Bar(gb + 2)
				gbo := cp.Bar(gb + 1)
				gc := cp.Bar(gb)
				if gbo.H < ga.H && gbo.L > ga.L &&
					((gc.H < gbo.H && gc.L < gbo.L) ||
						(gc.H > gbo.H && gc.L > gbo.L)) {
					superseded = true
					break
				}
			}
			if cp.Enough(gb) {
				ccGap := cp.Bar(gb)
				if patternResult > 0 && ccGap.C > p2.H {
					superseded = true
					break
				}
				if patternResult < 0 && ccGap.C < p2.L {
					superseded = true
					break
				}
			}
		}

		if superseded {
			continue
		}

		// Current bar confirms?
		cc := cp.Bar(1)
		if patternResult > 0 && cc.C > p2.H {
			return 200.0
		}
		if patternResult < 0 && cc.C < p2.L {
			return -200.0
		}
	}

	return 0.0
}
