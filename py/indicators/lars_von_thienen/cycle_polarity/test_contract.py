import unittest
from py.entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from py.entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from py.entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from py.indicators.core.component_triple_mnemonic import component_triple_mnemonic
from py.indicators.lars_von_thienen.cycle_polarity import CyclePolarity, CyclePolarityParams

class TestComponentContract(unittest.TestCase):
    def test_mnemonics(self):
        for bc,qc,tc in [(None,None,None),(BarComponent.OPEN,None,None),(None,QuoteComponent.BID,None),(None,None,TradeComponent.VOLUME),(BarComponent.OPEN,QuoteComponent.BID,None),(BarComponent.OPEN,None,TradeComponent.VOLUME),(None,QuoteComponent.BID,TradeComponent.VOLUME)]:
            with self.subTest(components=(bc,qc,tc)):
                params=CyclePolarityParams(bar_component=bc,quote_component=qc,trade_component=tc)
                indicator=CyclePolarity(params)
                suffix=component_triple_mnemonic(bc if bc is not None else DEFAULT_BAR_COMPONENT,qc if qc is not None else DEFAULT_QUOTE_COMPONENT,tc if tc is not None else DEFAULT_TRADE_COMPONENT)
                self.assertTrue(indicator.metadata().mnemonic.endswith(suffix+")"))
