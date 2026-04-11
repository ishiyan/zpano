"""Bar component enum and utilities."""

from enum import IntEnum
from typing import Callable

from .bar import Bar


class BarComponent(IntEnum):
    """Enumerates components of a Bar."""
    OPEN = 0
    HIGH = 1
    LOW = 2
    CLOSE = 3
    VOLUME = 4
    MEDIAN = 5
    TYPICAL = 6
    WEIGHTED = 7
    AVERAGE = 8


def bar_component_value(component: BarComponent) -> Callable[[Bar], float]:
    """Returns a function that extracts the given component value from a Bar."""
    if component == BarComponent.OPEN:
        return lambda b: b.open
    elif component == BarComponent.HIGH:
        return lambda b: b.high
    elif component == BarComponent.LOW:
        return lambda b: b.low
    elif component == BarComponent.CLOSE:
        return lambda b: b.close
    elif component == BarComponent.VOLUME:
        return lambda b: b.volume
    elif component == BarComponent.MEDIAN:
        return lambda b: b.median()
    elif component == BarComponent.TYPICAL:
        return lambda b: b.typical()
    elif component == BarComponent.WEIGHTED:
        return lambda b: b.weighted()
    elif component == BarComponent.AVERAGE:
        return lambda b: b.average()
    else:
        return lambda b: b.close


def bar_component_mnemonic(component: BarComponent) -> str:
    """Returns the mnemonic string for the given bar component."""
    _mnemonics = {
        BarComponent.OPEN: 'o',
        BarComponent.HIGH: 'h',
        BarComponent.LOW: 'l',
        BarComponent.CLOSE: 'c',
        BarComponent.VOLUME: 'v',
        BarComponent.MEDIAN: 'hl/2',
        BarComponent.TYPICAL: 'hlc/3',
        BarComponent.WEIGHTED: 'hlcc/4',
        BarComponent.AVERAGE: 'ohlc/4',
    }
    return _mnemonics.get(component, '??')
