"""LineIndicator base class for indicators with a single numeric input/output."""

from typing import Callable

from ...entities.bar import Bar
from ...entities.quote import Quote
from ...entities.trade import Trade
from ...entities.scalar import Scalar
from .output import Output


class LineIndicator:
    """Provides update_scalar, update_bar, update_quote and update_trade
    methods for indicators that take a single numeric input and produce
    output via a core update function.

    Concrete indicators should inherit from this class and pass their
    update method to the constructor.
    """

    def __init__(
        self,
        mnemonic: str,
        description: str,
        bar_func: Callable[[Bar], float],
        quote_func: Callable[[Quote], float],
        trade_func: Callable[[Trade], float],
        update_fn: Callable[[float], float],
    ) -> None:
        self.mnemonic = mnemonic
        self.description = description
        self._bar_func = bar_func
        self._quote_func = quote_func
        self._trade_func = trade_func
        self._update_fn = update_fn

    def update_scalar(self, sample: Scalar) -> Output:
        """Updates the indicator given the next scalar sample."""
        value = self._update_fn(sample.value)
        return [Scalar(sample.time, value)]

    def update_bar(self, sample: Bar) -> Output:
        """Updates the indicator given the next bar sample."""
        return self.update_scalar(Scalar(sample.time, self._bar_func(sample)))

    def update_quote(self, sample: Quote) -> Output:
        """Updates the indicator given the next quote sample."""
        return self.update_scalar(Scalar(sample.time, self._quote_func(sample)))

    def update_trade(self, sample: Trade) -> Output:
        """Updates the indicator given the next trade sample."""
        return self.update_scalar(Scalar(sample.time, self._trade_func(sample)))
