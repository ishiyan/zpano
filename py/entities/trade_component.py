"""Trade component enum and utilities."""

from enum import IntEnum
from typing import Callable

from .trade import Trade


class TradeComponent(IntEnum):
    """Enumerates components of a Trade."""
    PRICE = 0
    VOLUME = 1


def trade_component_value(component: TradeComponent) -> Callable[[Trade], float]:
    """Returns a function that extracts the given component value from a Trade."""
    if component == TradeComponent.PRICE:
        return lambda t: t.price
    elif component == TradeComponent.VOLUME:
        return lambda t: t.volume
    else:
        return lambda t: t.price


def trade_component_mnemonic(component: TradeComponent) -> str:
    """Returns the mnemonic string for the given trade component."""
    _mnemonics = {
        TradeComponent.PRICE: 'p',
        TradeComponent.VOLUME: 'v',
    }
    return _mnemonics.get(component, '??')
