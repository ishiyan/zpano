"""A price quote (bid/ask price and size pair)."""

from datetime import datetime


class Quote:
    """Represents a bid/ask quote."""

    def __init__(self, time: datetime, bid_price: float, ask_price: float, \
                 bid_size: float, ask_size: float):
        self.time = time
        self.bid_price = bid_price
        self.ask_price = ask_price
        self.bid_size = bid_size
        self.ask_size = ask_size

    def mid(self) -> float:
        """The mid-price: (ask_price + bid_price) / 2."""
        return (self.ask_price + self.bid_price) / 2

    def weighted(self) -> float:
        """The weighted price: (ask*askSize + bid*bidSize) / (askSize + bidSize).

        Returns 0 if total size is 0.
        """
        size = self.ask_size + self.bid_size
        if size == 0:
            return 0.0
        return (self.ask_price * self.ask_size + self.bid_price * self.bid_size) / size

    def weighted_mid(self) -> float:
        """The weighted mid-price (micro-price): (ask*bidSize + bid*askSize) / (askSize + bidSize).

        Returns 0 if total size is 0.
        """
        size = self.ask_size + self.bid_size
        if size == 0:
            return 0.0
        return (self.ask_price * self.bid_size + self.bid_price * self.ask_size) / size

    def spread_bp(self) -> float:
        """The spread in basis points: 20000 * (ask - bid) / (ask + bid).

        Returns 0 if mid is 0.
        """
        mid = self.ask_price + self.bid_price
        if mid == 0:
            return 0.0
        return 20000 * (self.ask_price - self.bid_price) / mid

    def __repr__(self) -> str:
        return f"Quote({self.time}, {self.bid_price}, {self.ask_price}, " \
               f"{self.bid_size}, {self.ask_size})"
