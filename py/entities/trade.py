"""A trade (price and volume) entity."""

from datetime import datetime


class Trade:
    """Represents a trade (time and sales)."""

    def __init__(self, time: datetime, price: float, volume: float):
        self.time = time
        self.price = price
        self.volume = volume

    def __repr__(self) -> str:
        return f"Trade({self.time}, {self.price}, {self.volume})"
