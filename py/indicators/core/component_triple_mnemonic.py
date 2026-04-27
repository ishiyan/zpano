"""Component triple mnemonic utility for building indicator mnemonic suffixes."""

from ...entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_mnemonic
from ...entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_mnemonic
from ...entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_mnemonic


def component_triple_mnemonic(
    bc: BarComponent, qc: QuoteComponent, tc: TradeComponent,
) -> str:
    """Builds a mnemonic suffix string from bar, quote and trade components.

    A component equal to its default value is omitted from the mnemonic.
    For example, if bar component is MEDIAN (non-default), the result is ", hl/2".
    If bar component is CLOSE (default), it is omitted.
    """
    s = ""

    if bc != DEFAULT_BAR_COMPONENT:
        s += ", " + bar_component_mnemonic(bc)

    if qc != DEFAULT_QUOTE_COMPONENT:
        s += ", " + quote_component_mnemonic(qc)

    if tc != DEFAULT_TRADE_COMPONENT:
        s += ", " + trade_component_mnemonic(tc)

    return s
