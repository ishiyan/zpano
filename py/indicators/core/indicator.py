"""Indicator interface — abstract base class for all indicators."""

from abc import ABC, abstractmethod

from ...entities.bar import Bar
from ...entities.quote import Quote
from ...entities.trade import Trade
from ...entities.scalar import Scalar
from .metadata import Metadata
from .output import Output


class Indicator(ABC):
    """Describes common indicator functionality."""

    @abstractmethod
    def is_primed(self) -> bool:
        """Indicates whether an indicator is primed."""
        ...

    @abstractmethod
    def metadata(self) -> Metadata:
        """Describes the output data of an indicator."""
        ...

    @abstractmethod
    def update_scalar(self, sample: Scalar) -> Output:
        """Updates an indicator given the next scalar sample."""
        ...

    @abstractmethod
    def update_bar(self, sample: Bar) -> Output:
        """Updates an indicator given the next bar sample."""
        ...

    @abstractmethod
    def update_quote(self, sample: Quote) -> Output:
        """Updates an indicator given the next quote sample."""
        ...

    @abstractmethod
    def update_trade(self, sample: Trade) -> Output:
        """Updates an indicator given the next trade sample."""
        ...


def update_scalars(ind: Indicator, samples: list[Scalar]) -> list[Output]:
    """Updates the indicator given a list of scalar samples."""
    return [ind.update_scalar(s) for s in samples]


def update_bars(ind: Indicator, samples: list[Bar]) -> list[Output]:
    """Updates the indicator given a list of bar samples."""
    return [ind.update_bar(s) for s in samples]


def update_quotes(ind: Indicator, samples: list[Quote]) -> list[Output]:
    """Updates the indicator given a list of quote samples."""
    return [ind.update_quote(s) for s in samples]


def update_trades(ind: Indicator, samples: list[Trade]) -> list[Output]:
    """Updates the indicator given a list of trade samples."""
    return [ind.update_trade(s) for s in samples]
