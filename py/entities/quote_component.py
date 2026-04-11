"""Quote component enum and utilities."""

from enum import IntEnum
from typing import Callable

from .quote import Quote


class QuoteComponent(IntEnum):
    """Enumerates components of a Quote."""
    BID = 0
    ASK = 1
    BID_SIZE = 2
    ASK_SIZE = 3
    MID = 4
    WEIGHTED = 5
    WEIGHTED_MID = 6
    SPREAD_BP = 7


def quote_component_value(component: QuoteComponent) -> Callable[[Quote], float]:
    """Returns a function that extracts the given component value from a Quote."""
    if component == QuoteComponent.BID:
        return lambda q: q.bid_price
    elif component == QuoteComponent.ASK:
        return lambda q: q.ask_price
    elif component == QuoteComponent.BID_SIZE:
        return lambda q: q.bid_size
    elif component == QuoteComponent.ASK_SIZE:
        return lambda q: q.ask_size
    elif component == QuoteComponent.MID:
        return lambda q: q.mid()
    elif component == QuoteComponent.WEIGHTED:
        return lambda q: q.weighted()
    elif component == QuoteComponent.WEIGHTED_MID:
        return lambda q: q.weighted_mid()
    elif component == QuoteComponent.SPREAD_BP:
        return lambda q: q.spread_bp()
    else:
        return lambda q: q.mid()


def quote_component_mnemonic(component: QuoteComponent) -> str:
    """Returns the mnemonic string for the given quote component."""
    _mnemonics = {
        QuoteComponent.BID: 'b',
        QuoteComponent.ASK: 'a',
        QuoteComponent.BID_SIZE: 'bs',
        QuoteComponent.ASK_SIZE: 'as',
        QuoteComponent.MID: 'ba/2',
        QuoteComponent.WEIGHTED: '(bbs+aas)/(bs+as)',
        QuoteComponent.WEIGHTED_MID: '(bas+abs)/(bs+as)',
        QuoteComponent.SPREAD_BP: 'spread bp',
    }
    return _mnemonics.get(component, '??')
