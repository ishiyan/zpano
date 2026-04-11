"""An [open, high, low, close, volume] bar."""

from datetime import datetime


class Bar:
    """Represents an OHLCV price bar."""

    def __init__(self, time: datetime, open: float, high: float, low: float, \
                 close: float, volume: float):
        self.time = time
        self.open = open
        self.high = high
        self.low = low
        self.close = close
        self.volume = volume

    def is_rising(self) -> bool:
        """Indicates whether this is a rising bar (open < close)."""
        return self.open < self.close

    def is_falling(self) -> bool:
        """Indicates whether this is a falling bar (close < open)."""
        return self.close < self.open

    def median(self) -> float:
        """The median price: (low + high) / 2."""
        return (self.low + self.high) / 2

    def typical(self) -> float:
        """The typical price: (low + high + close) / 3."""
        return (self.low + self.high + self.close) / 3

    def weighted(self) -> float:
        """The weighted price: (low + high + 2*close) / 4."""
        return (self.low + self.high + self.close + self.close) / 4

    def average(self) -> float:
        """The average price: (low + high + open + close) / 4."""
        return (self.low + self.high + self.open + self.close) / 4

    def __repr__(self) -> str:
        return f"Bar({self.time}, {self.open}, {self.high}, {self.low}, " \
               f"{self.close}, {self.volume})"
