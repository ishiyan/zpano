"""Hikkake pattern (3-candle) with optional confirmation."""
from __future__ import annotations


def hikkake(self) -> int:
    """Hikkake: a three-candle pattern with stateful confirmation.

    TA-Lib behavior:
    - Detection bar: outputs +100 (bullish) or -100 (bearish)
    - Confirmation bar (within 3 bars of detection): outputs +200 or -200
    - If a new hikkake is detected on the same bar as a confirmation,
      the new hikkake takes priority.

    Must have:
    - first and second candle: inside bar (2nd lower high, higher low)
    - third candle: lower high AND lower low (bull) or higher high AND
      higher low (bear)

    Confirmation: close > high of 2nd candle (bull) or close < low of
    2nd candle (bear) within 3 bars.

    Returns:
        +100/-100 for initial detection, +200/-200 for confirmation,
        0 for no pattern.
    """
    if not self._enough(3):
        return 0

    # Check for new hikkake pattern at current bar.
    o1, h1, l1, c1 = self._bar(3)
    o2, h2, l2, c2 = self._bar(2)
    o3, h3, l3, c3 = self._bar(1)

    # Inside bar check.
    if h2 < h1 and l2 > l1:
        # Bullish: 3rd has lower high AND lower low.
        if h3 < h2 and l3 < l2:
            return 100
        # Bearish: 3rd has higher high AND higher low.
        if h3 > h2 and l3 > l2:
            return -100

    # No new pattern — check for confirmation of a recent hikkake.
    # Look back 1-3 bars for a hikkake pattern.
    for lookback in range(1, 4):
        n = 3 + lookback
        if not self._enough(n):
            break

        p1o, p1h, p1l, p1c = self._bar(n)        # 1st of pattern
        p2o, p2h, p2l, p2c = self._bar(n - 1)     # inside bar (2nd)
        p3o, p3h, p3l, p3c = self._bar(n - 2)     # breakout bar (3rd)

        # Must be a valid hikkake at that position.
        if not (p2h < p1h and p2l > p1l):
            continue

        if p3h < p2h and p3l < p2l:
            pattern_result = 100  # bullish
        elif p3h > p2h and p3l > p2l:
            pattern_result = -100  # bearish
        else:
            continue

        # Check that no intervening bar already confirmed or re-detected.
        # If there's a newer hikkake between the pattern and current bar,
        # the older one is superseded.
        superseded = False
        for gap in range(1, lookback):
            gb = n - 2 - gap  # bars between breakout and current
            if gb < 1:
                break
            # Check if there's a newer hikkake at this position
            if self._enough(gb + 2):
                ga, gah, gal, gac = self._bar(gb + 2)
                gbo, gbh, gbl, gbc = self._bar(gb + 1)
                gco, gch, gcl, gcc = self._bar(gb)
                if (gbh < gah and gbl > gal and
                    ((gch < gbh and gcl < gbl) or
                     (gch > gbh and gcl > gbl))):
                    superseded = True
                    break
            # Check if confirmation already happened
            if self._enough(gb):
                _, _, _, cc_gap = self._bar(gb)
                if pattern_result > 0 and cc_gap > p2h:
                    superseded = True
                    break
                if pattern_result < 0 and cc_gap < p2l:
                    superseded = True
                    break

        if superseded:
            continue

        # Current bar confirms?
        _, _, _, cc = self._bar(1)
        if pattern_result > 0 and cc > p2h:
            return 200
        if pattern_result < 0 and cc < p2l:
            return -200

    return 0
