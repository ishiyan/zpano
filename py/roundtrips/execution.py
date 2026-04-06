from datetime import datetime
from enum import Enum


class OrderSide(Enum):
    """Enumerates the sides of an order."""

    BUY = 'buy'
    """A buy order."""

    SELL = 'sell'
    """A sell order."""

    def is_sell(self):
        return self == OrderSide.SELL


class Execution:
    """Represents an order execution (fill).

    Parameters
    ----------
    side : OrderSide
        The side of the order.
    price : float
        The execution price.
    commission_per_unit : float
        The commission per unit of quantity.
    unrealized_price_high : float
        The highest unrealized price during the execution period.
    unrealized_price_low : float
        The lowest unrealized price during the execution period.
    datetime : datetime
        The date and time of the execution.
    """

    def __init__(self, *,
                 side: OrderSide,
                 price: float,
                 commission_per_unit: float,
                 unrealized_price_high: float,
                 unrealized_price_low: float,
                 dt: datetime):
        self.side = side
        self.price = price
        self.commission_per_unit = commission_per_unit
        self.unrealized_price_high = unrealized_price_high
        self.unrealized_price_low = unrealized_price_low
        self.datetime = dt
