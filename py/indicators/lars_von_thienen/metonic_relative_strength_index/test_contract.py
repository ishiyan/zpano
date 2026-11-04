import unittest
from py.entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT
from py.entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT
from py.entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT
from py.indicators.core.component_triple_mnemonic import component_triple_mnemonic
from py.indicators.lars_von_thienen.metonic_relative_strength_index import MetonicRelativeStrengthIndex, MetonicRelativeStrengthIndexParams

class TestComponentContract(unittest.TestCase):
    def test_mnemonics(self):
        for bc,qc,tc in [(None,None,None),(BarComponent.OPEN,None,None),(None,QuoteComponent.BID,None),(None,None,TradeComponent.VOLUME),(BarComponent.OPEN,QuoteComponent.BID,None),(BarComponent.OPEN,None,TradeComponent.VOLUME),(None,QuoteComponent.BID,TradeComponent.VOLUME)]:
            with self.subTest(components=(bc,qc,tc)):
                params=MetonicRelativeStrengthIndexParams(bar_component=bc,quote_component=qc,trade_component=tc)
                indicator=MetonicRelativeStrengthIndex(params)
                suffix=component_triple_mnemonic(bc if bc is not None else DEFAULT_BAR_COMPONENT,qc if qc is not None else DEFAULT_QUOTE_COMPONENT,tc if tc is not None else DEFAULT_TRADE_COMPONENT)
                self.assertTrue(indicator.metadata().mnemonic.endswith(suffix+")"))

    def test_production_calendar_delays(self):
        from .test_testdata import INPUT_CLOSE
        import math
        indicator=MetonicRelativeStrengthIndex(MetonicRelativeStrengthIndexParams(length=1))
        for i in range(6945):
            actual=indicator.update(INPUT_CLOSE[i%len(INPUT_CLOSE)]).value
            if i<=6939:
                self.assertTrue(math.isnan(actual))
            else:
                expected=sum(100.0 if INPUT_CLOSE[(i-lag)%len(INPUT_CLOSE)]>=INPUT_CLOSE[(i-lag-1)%len(INPUT_CLOSE)] else 0.0 for lag in (6939,4016,2922))/3.0
                self.assertAlmostEqual(actual,expected,places=13)
