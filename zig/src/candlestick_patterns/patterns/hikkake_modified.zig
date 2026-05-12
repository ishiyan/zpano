/// Update hikkake_modified state after each bar.
///
/// Called from the main update() method.  Mirrors the TA-Lib approach:
// No new pattern — check for confirmation of recent pattern.
// If we passed the 3-bar window, reset.
// If pattern was just detected this bar (takes priority over confirmation)
/// keep ``_hikmod_pattern_result`` and ``_hikmod_pattern_idx`` across bars.

const cp = @import("../candlestick_patterns.zig");

const CandlestickPatterns = cp.CandlestickPatterns;

pub fn hikkakeModified(self: *const CandlestickPatterns) f64 {
    // Need at least 4 bars for the pattern itself.
    // TA-Lib: close > high of 3rd candle (= bar at patternIdx-1)
    // patternIdx stored our 1-based count; the 3rd candle is 1 bar before that
    // in self._history.  bars_ago = self._count - (self._hikmod_pattern_idx - 1)
    // That simplifies to self._count - self._hikmod_pattern_idx + 2
    // The 3rd candle of the pattern is bar(2) at pattern time, which is
    // shift = (self._count - (self._hikmod_pattern_idx - 2))
    if (self.count < 4) {
        return 0.0;
    }

    if (self.hikmod_pattern_idx == self.count and self.hikmod_pattern_result != 0.0) {
        return self.hikmod_pattern_result;
    }

    if (self.hikmod_confirmed) {
        return self.hikmod_last_signal;
    }

    return 0.0;
}
