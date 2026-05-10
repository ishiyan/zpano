"""Hikkake Modified pattern (4-candle) with stateful confirmation."""
from __future__ import annotations

from ..core.primitives import real_body


def _hikkake_modified_update(self) -> None:
    """Update hikkake_modified state after each bar.

    Called from the main update() method.  Mirrors the TA-Lib approach:
    keep ``_hikmod_pattern_result`` and ``_hikmod_pattern_idx`` across bars.
    """
    # Need at least 4 bars for the pattern itself.
    if self._count < 4:
        return

    # The bar indices (1-based from end): bar(1)=current, bar(2)=prev, …
    o1, h1, l1, c1 = self._bar(4)  # i-3
    o2, h2, l2, c2 = self._bar(3)  # i-2
    o3, h3, l3, c3 = self._bar(2)  # i-1
    o4, h4, l4, c4 = self._bar(1)  # i

    # Check for new pattern (overwrites any previous unconfirmed pattern).
    if (h2 < h1 and l2 > l1 and     # 2nd inside 1st
            h3 < h2 and l3 > l2):    # 3rd inside 2nd
        near_avg = self._avg(self._near, 3)  # at i-2
        # Bullish: 4th breaks low, 2nd close near its low
        if (h4 < h3 and l4 < l3 and
                c2 <= l2 + near_avg):
            self._hikmod_pattern_result = 100
            self._hikmod_pattern_idx = self._count  # current bar index (1-based count)
            return
        # Bearish: 4th breaks high, 2nd close near its high
        if (h4 > h3 and l4 > l3 and
                c2 >= h2 - near_avg):
            self._hikmod_pattern_result = -100
            self._hikmod_pattern_idx = self._count
            return

    # No new pattern — check for confirmation of recent pattern.
    if (self._hikmod_pattern_result != 0 and
            self._count <= self._hikmod_pattern_idx + 3):
        # TA-Lib: close > high of 3rd candle (= bar at patternIdx-1)
        # patternIdx stored our 1-based count; the 3rd candle is 1 bar before that
        # in self._history.  bars_ago = self._count - (self._hikmod_pattern_idx - 1)
        bars_ago = self._count - self._hikmod_pattern_idx + 1 + 1
        # That simplifies to self._count - self._hikmod_pattern_idx + 2
        # The 3rd candle of the pattern is bar(2) at pattern time, which is
        # shift = (self._count - (self._hikmod_pattern_idx - 2))
        shift_3rd = self._count - self._hikmod_pattern_idx + 2
        _, h_3rd, l_3rd, _ = self._bar(shift_3rd)

        if self._hikmod_pattern_result > 0 and c4 > h_3rd:
            # confirmed — consume the pattern
            self._hikmod_last_signal = 200
            self._hikmod_pattern_result = 0
            self._hikmod_pattern_idx = 0
            self._hikmod_confirmed = True
            return
        if self._hikmod_pattern_result < 0 and c4 < l_3rd:
            self._hikmod_last_signal = -200
            self._hikmod_pattern_result = 0
            self._hikmod_pattern_idx = 0
            self._hikmod_confirmed = True
            return

    # If we passed the 3-bar window, reset.
    if (self._hikmod_pattern_result != 0 and
            self._count > self._hikmod_pattern_idx + 3):
        self._hikmod_pattern_result = 0
        self._hikmod_pattern_idx = 0


def hikkake_modified(self) -> int:
    """Hikkake Modified: a four-candle pattern with near criterion.

    Returns:
        +100/-100 for detection, +200/-200 for confirmation, 0 otherwise.
    """
    if self._count < 4:
        return 0

    # If pattern was just detected this bar (takes priority over confirmation)
    if (self._hikmod_pattern_idx == self._count and
            self._hikmod_pattern_result != 0):
        return self._hikmod_pattern_result

    # If just confirmed this bar
    if self._hikmod_confirmed:
        return self._hikmod_last_signal

    return 0
