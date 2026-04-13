"""Tests for entities module."""

import unittest
from datetime import datetime

from .bar import Bar
from .bar_component import BarComponent, bar_component_value, bar_component_mnemonic, DEFAULT_BAR_COMPONENT
from .quote import Quote
from .quote_component import QuoteComponent, quote_component_value, quote_component_mnemonic, DEFAULT_QUOTE_COMPONENT
from .trade import Trade
from .trade_component import TradeComponent, trade_component_value, trade_component_mnemonic, DEFAULT_TRADE_COMPONENT


def _bar(o: float, h: float, l: float, c: float, v: float) -> Bar:
    return Bar(datetime(2021, 4, 1), o, h, l, c, v)


class TestBarMedian(unittest.TestCase):
    def test_median(self):
        b = _bar(0, 3, 2, 0, 0)
        self.assertEqual(b.median(), (b.low + b.high) / 2)


class TestBarTypical(unittest.TestCase):
    def test_typical(self):
        b = _bar(0, 4, 2, 3, 0)
        self.assertEqual(b.typical(), (b.low + b.high + b.close) / 3)


class TestBarWeighted(unittest.TestCase):
    def test_weighted(self):
        b = _bar(0, 4, 2, 3, 0)
        self.assertEqual(b.weighted(), (b.low + b.high + b.close + b.close) / 4)


class TestBarAverage(unittest.TestCase):
    def test_average(self):
        b = _bar(3, 5, 2, 4, 0)
        self.assertEqual(b.average(), (b.low + b.high + b.open + b.close) / 4)


class TestBarIsRising(unittest.TestCase):
    def test_rising(self):
        b = _bar(2, 0, 0, 3, 0)
        self.assertTrue(b.is_rising())

    def test_falling(self):
        b = _bar(3, 0, 0, 2, 0)
        self.assertFalse(b.is_rising())

    def test_flat(self):
        b = _bar(0, 0, 0, 0, 0)
        self.assertFalse(b.is_rising())


class TestBarIsFalling(unittest.TestCase):
    def test_rising(self):
        b = _bar(2, 0, 0, 3, 0)
        self.assertFalse(b.is_falling())

    def test_falling(self):
        b = _bar(3, 0, 0, 2, 0)
        self.assertTrue(b.is_falling())

    def test_flat(self):
        b = _bar(0, 0, 0, 0, 0)
        self.assertFalse(b.is_falling())


class TestBarComponentValue(unittest.TestCase):
    def setUp(self):
        self.b = Bar(datetime(2021, 4, 1), 2, 4, 1, 3, 5)

    def test_open(self):
        self.assertEqual(bar_component_value(BarComponent.OPEN)(self.b), 2)

    def test_high(self):
        self.assertEqual(bar_component_value(BarComponent.HIGH)(self.b), 4)

    def test_low(self):
        self.assertEqual(bar_component_value(BarComponent.LOW)(self.b), 1)

    def test_close(self):
        self.assertEqual(bar_component_value(BarComponent.CLOSE)(self.b), 3)

    def test_volume(self):
        self.assertEqual(bar_component_value(BarComponent.VOLUME)(self.b), 5)

    def test_median(self):
        self.assertEqual(bar_component_value(BarComponent.MEDIAN)(self.b), (1 + 4) / 2)

    def test_typical(self):
        self.assertEqual(bar_component_value(BarComponent.TYPICAL)(self.b), (1 + 4 + 3) / 3)

    def test_weighted(self):
        self.assertEqual(bar_component_value(BarComponent.WEIGHTED)(self.b), (1 + 4 + 3 + 3) / 4)

    def test_average(self):
        self.assertEqual(bar_component_value(BarComponent.AVERAGE)(self.b), (1 + 4 + 3 + 2) / 4)

    def test_default(self):
        self.assertEqual(bar_component_value(9999)(self.b), self.b.close)


class TestBarComponentMnemonic(unittest.TestCase):
    def test_all(self):
        expected = [
            (BarComponent.OPEN, 'o'),
            (BarComponent.HIGH, 'h'),
            (BarComponent.LOW, 'l'),
            (BarComponent.CLOSE, 'c'),
            (BarComponent.VOLUME, 'v'),
            (BarComponent.MEDIAN, 'hl/2'),
            (BarComponent.TYPICAL, 'hlc/3'),
            (BarComponent.WEIGHTED, 'hlcc/4'),
            (BarComponent.AVERAGE, 'ohlc/4'),
        ]
        for comp, mnemonic in expected:
            self.assertEqual(bar_component_mnemonic(comp), mnemonic)

    def test_unknown(self):
        self.assertEqual(bar_component_mnemonic(9999), '??')


class TestDefaultBarComponent(unittest.TestCase):
    def test_default(self):
        self.assertEqual(DEFAULT_BAR_COMPONENT, BarComponent.CLOSE)


class TestQuoteMid(unittest.TestCase):
    def test_mid(self):
        q = Quote(datetime(2021, 4, 1), 3.0, 2.0, 0, 0)
        self.assertEqual(q.mid(), (q.ask_price + q.bid_price) / 2)


class TestQuoteWeighted(unittest.TestCase):
    def test_weighted(self):
        q = Quote(datetime(2021, 4, 1), 3.0, 2.0, 5.0, 4.0)
        expected = (q.ask_price * q.ask_size + q.bid_price * q.bid_size) / \
                   (q.ask_size + q.bid_size)
        self.assertEqual(q.weighted(), expected)

    def test_zero_size(self):
        q = Quote(datetime(2021, 4, 1), 3.0, 2.0, 0, 0)
        self.assertEqual(q.weighted(), 0.0)


class TestQuoteWeightedMid(unittest.TestCase):
    def test_weighted_mid(self):
        q = Quote(datetime(2021, 4, 1), 3.0, 2.0, 5.0, 4.0)
        expected = (q.ask_price * q.bid_size + q.bid_price * q.ask_size) / \
                   (q.ask_size + q.bid_size)
        self.assertEqual(q.weighted_mid(), expected)

    def test_zero_size(self):
        q = Quote(datetime(2021, 4, 1), 3.0, 2.0, 0, 0)
        self.assertEqual(q.weighted_mid(), 0.0)


class TestQuoteSpreadBp(unittest.TestCase):
    def test_spread_bp(self):
        q = Quote(datetime(2021, 4, 1), 3.0, 2.0, 0, 0)
        expected = 20000 * (q.ask_price - q.bid_price) / (q.ask_price + q.bid_price)
        self.assertEqual(q.spread_bp(), expected)

    def test_zero_mid(self):
        q = Quote(datetime(2021, 4, 1), 0, 0, 0, 0)
        self.assertEqual(q.spread_bp(), 0.0)


class TestQuoteComponentValue(unittest.TestCase):
    def setUp(self):
        self.q = Quote(datetime(2021, 4, 1), 2.0, 1.0, 4.0, 3.0)

    def test_bid(self):
        self.assertEqual(quote_component_value(QuoteComponent.BID)(self.q), 2)

    def test_ask(self):
        self.assertEqual(quote_component_value(QuoteComponent.ASK)(self.q), 1)

    def test_bid_size(self):
        self.assertEqual(quote_component_value(QuoteComponent.BID_SIZE)(self.q), 4)

    def test_ask_size(self):
        self.assertEqual(quote_component_value(QuoteComponent.ASK_SIZE)(self.q), 3)

    def test_mid(self):
        self.assertEqual(quote_component_value(QuoteComponent.MID)(self.q), (1 + 2) / 2)

    def test_weighted(self):
        self.assertEqual(
            quote_component_value(QuoteComponent.WEIGHTED)(self.q),
            (1 * 3 + 2 * 4) / (3 + 4))

    def test_weighted_mid(self):
        self.assertEqual(
            quote_component_value(QuoteComponent.WEIGHTED_MID)(self.q),
            (1 * 4 + 2 * 3) / (3 + 4))

    def test_spread_bp(self):
        self.assertEqual(
            quote_component_value(QuoteComponent.SPREAD_BP)(self.q),
            10000 * 2 * (1 - 2) / (1 + 2))

    def test_default(self):
        self.assertEqual(quote_component_value(9999)(self.q), self.q.mid())


class TestQuoteComponentMnemonic(unittest.TestCase):
    def test_all(self):
        expected = [
            (QuoteComponent.BID, 'b'),
            (QuoteComponent.ASK, 'a'),
            (QuoteComponent.BID_SIZE, 'bs'),
            (QuoteComponent.ASK_SIZE, 'as'),
            (QuoteComponent.MID, 'ba/2'),
            (QuoteComponent.WEIGHTED, '(bbs+aas)/(bs+as)'),
            (QuoteComponent.WEIGHTED_MID, '(bas+abs)/(bs+as)'),
            (QuoteComponent.SPREAD_BP, 'spread bp'),
        ]
        for comp, mnemonic in expected:
            self.assertEqual(quote_component_mnemonic(comp), mnemonic)

    def test_unknown(self):
        self.assertEqual(quote_component_mnemonic(9999), '??')


class TestDefaultQuoteComponent(unittest.TestCase):
    def test_default(self):
        self.assertEqual(DEFAULT_QUOTE_COMPONENT, QuoteComponent.MID)


class TestTradeComponentValue(unittest.TestCase):
    def setUp(self):
        self.t = Trade(datetime(2021, 4, 1), 1.0, 2.0)

    def test_price(self):
        self.assertEqual(trade_component_value(TradeComponent.PRICE)(self.t), 1)

    def test_volume(self):
        self.assertEqual(trade_component_value(TradeComponent.VOLUME)(self.t), 2)

    def test_default(self):
        self.assertEqual(trade_component_value(9999)(self.t), self.t.price)


class TestTradeComponentMnemonic(unittest.TestCase):
    def test_all(self):
        expected = [
            (TradeComponent.PRICE, 'p'),
            (TradeComponent.VOLUME, 'v'),
        ]
        for comp, mnemonic in expected:
            self.assertEqual(trade_component_mnemonic(comp), mnemonic)

    def test_unknown(self):
        self.assertEqual(trade_component_mnemonic(9999), '??')


class TestDefaultTradeComponent(unittest.TestCase):
    def test_default(self):
        self.assertEqual(DEFAULT_TRADE_COMPONENT, TradeComponent.PRICE)


if __name__ == '__main__':
    unittest.main()
