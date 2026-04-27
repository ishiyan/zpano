"""Tests for Kaufman Adaptive Moving Average indicator."""

import math
import unittest
from datetime import datetime

from .kaufman_adaptive_moving_average import KaufmanAdaptiveMovingAverage
from .params import (KaufmanAdaptiveMovingAverageLengthParams,
                     KaufmanAdaptiveMovingAverageSmoothingFactorParams)
from .output import KaufmanAdaptiveMovingAverageOutput
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent
from ....entities.quote_component import QuoteComponent
from ....entities.trade_component import TradeComponent
from ...core.identifier import Identifier


# Data taken from TA-Lib test_KAMA.xls, Close, C5..C256, 252 entries.
_INPUT = [
    91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
    96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
    88.375000, 87.625000, 84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000,
    85.250000, 87.125000, 85.815000, 88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000,
    83.375000, 85.500000, 89.190000, 89.440000, 91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000,
    89.030000, 88.815000, 84.280000, 83.500000, 82.690000, 84.750000, 85.655000, 86.190000, 88.940000, 89.280000,
    88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000, 93.155000, 91.720000, 90.000000, 89.690000,
    88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000, 104.940000, 106.000000, 102.500000,
    102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000,
    112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000,
    110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
    116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000,
    124.750000, 123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000,
    132.250000, 131.000000, 132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000,
    136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000,
    125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000, 122.190000, 119.310000,
    123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000, 124.440000,
    122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
    132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000,
    130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000,
    117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
    107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
    94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000,
    95.000000, 95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000,
    105.000000, 104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000,
    113.370000, 109.000000, 109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000,
    108.000000, 108.620000, 109.750000, 109.810000, 109.000000, 108.750000, 107.870000,
]

# Expected KAMA values from TA-Lib test_KAMA.xls, J5..J256.
_EXPECTED = [
    math.nan, math.nan, math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan, math.nan, math.nan,
    92.6574744421924, 92.7783471257434, 93.0592520064115, 92.9356368995325, 92.9000149644911,
    92.8990048732289, 92.8229942018608, 92.7516051928620, 92.7414384525517, 92.6960363223993,
    92.3934372123882, 91.9139380062599, 90.7658162726830, 90.0740111936089, 89.3620815288014,
    87.6656280861040, 87.4895131032692, 87.4974604839614, 87.4487997113532, 87.4134797590652,
    87.3586513546248, 87.3571985565411, 87.3428271277309, 87.4342339727455, 87.4790967331831,
    87.4478089486627, 87.4341052772180, 87.2779545841798, 87.1866387951289, 87.0799098978843,
    86.9861110535034, 86.9549433796085, 87.0479997922396, 87.0668566957271, 87.2090146571776,
    87.4600776240503, 87.8014795040326, 87.8826076877600, 88.2803844203263, 88.5454141018648,
    88.5859031486005, 88.5965040436874, 88.2719621445720, 87.8163354339468, 86.8611444903465,
    86.6741610056912, 86.5906930013157, 86.5766752991618, 86.6296450514704, 86.6650208354184,
    86.6783504731998, 86.6895963952268, 87.6981988794437, 88.5095835057360, 89.9508715587081,
    90.9585930437125, 91.4794679492180, 91.5092409530174, 91.4856744284233, 91.4717808315536,
    91.4557387469302, 91.1940009725015, 89.4266294004067, 88.8455374050859, 88.3697094609281,
    88.5930899916723, 89.1316678888979, 90.8601116442358, 93.2091460910382, 94.0581656977510,
    94.9201636069605, 96.8889752566530, 99.4062425239817, 101.1201449462390, 102.3769237660390,
    102.6006738368170, 103.3003850710980, 103.6578508957870, 104.0764855627630, 106.4159093020280,
    112.1346727325330, 113.5057358502340, 114.2548283428500, 115.0085673230990, 115.3491682211620,
    115.4744042357010, 115.4586954188130, 115.4033778968360, 115.3819703222920, 115.4596680866820,
    115.4927139908920, 115.5083211482970, 115.3016588863670, 115.2382416224770, 115.1532481002890,
    115.1580191296150, 115.3257950434630, 115.3602912952500, 115.4272550190370, 115.4236654978450,
    115.4094918992810, 115.4100431369950, 115.4265778341240, 115.7744740794160, 116.0930627623780,
    116.3101967717570, 116.6603109196670, 117.3487018143020, 117.8153888221880, 118.4531290804430,
    119.3499419409230, 119.8086689971510, 120.6175024210070, 122.0458817467430, 123.9704416533650,
    125.8138480326600, 126.3738969105690, 127.6872486354350, 129.2393432164220, 131.6880947713340,
    133.5239638088170, 135.0004207395880, 135.6288233403940, 135.7374059656390, 135.8007904215550,
    135.7583248045180, 135.5543718432480, 135.2569852680960, 133.6204824276490, 131.3192797761920,
    128.7932379609940, 128.4062405870340, 128.4039316032540, 128.0791656483760, 127.8414201748350,
    127.1988985844810, 126.5381546649790, 125.6607070438540, 125.6440698902700, 125.6229493897650,
    125.5972771029140, 125.1856884028260, 125.1156207098550, 124.9914050152240, 124.9677440635400,
    125.0508437113440, 125.3554407671800, 125.3059272985400, 125.2940386783170, 125.2530757692210,
    125.2419747210570, 125.1887237516160, 125.1656598262800, 125.1342643444030, 125.1261708430550,
    125.0293527295390, 125.0082100078360, 125.1058124672220, 125.1321388339230, 125.5284397017590,
    126.2554117345480, 126.9803557764160, 128.5646940398630, 129.8559054638140, 130.0995104273400,
    130.5156892070650, 130.6273781337970, 130.6136632314180, 130.5821372483140, 130.5780360175850,
    130.4619826221790, 130.2592097652620, 129.0901503140520, 128.7592330158310, 128.3218396854650,
    127.9194919253990, 127.1326782278630, 126.7107330400510, 126.1909025410680, 125.5077119513560,
    125.3652360592940, 125.0689417277010, 124.6785367307510, 123.1715118076970, 122.3246069304410,
    120.4996045001390, 118.0226226271800, 116.5389084881180, 115.7700047414230, 114.4762055991300,
    112.8691910705370, 111.7330463494810, 105.8813879559000, 103.7386265802100, 101.7705073498860,
    100.9556429673090, 100.0740835866110, 99.5051792798608, 99.4197548401710, 99.2260466472373,
    98.8377738185378, 98.4351675572326, 98.3887252314702, 98.0891751313173, 98.0708172638065,
    98.0047820815841, 97.9717872707032, 97.9587393847739, 97.9160266616328, 97.8272391679346,
    97.8109932013579, 97.7811643727499, 97.7968786191168, 98.8421055702164, 100.3972096134300,
    101.1278312905150, 101.3486183367770, 101.7632588756100, 101.9699249107700, 102.0803180404650,
    102.2131955779830, 102.6495717799380, 104.1660350536590, 105.9174582846280, 107.1295132390960,
    109.3610815395210, 109.7246822740860, 109.7071337912410, 109.7068748325140, 109.6867591775540,
    109.6319778699710, 109.6221417907160, 109.6271816752350, 109.5930223785590, 109.6314010730650,
    109.3937985883840, 109.3445353771140, 109.3487688924230, 109.3510517081720, 109.3489501843720,
    109.3310159853090, 109.2940150671190,
]


class TestKaufmanAdaptiveMovingAverage(unittest.TestCase):
    """Tests for KAMA indicator."""

    def test_values_length(self):
        """Test KAMA(10, 2, 30) with 252 close prices from TA-Lib."""
        params = KaufmanAdaptiveMovingAverageLengthParams(
            efficiency_ratio_length=10, fastest_length=2, slowest_length=30)
        kama = KaufmanAdaptiveMovingAverage.from_length(params)

        for i in range(10):
            result = kama.update(_INPUT[i])
            self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")

        for i in range(10, len(_INPUT)):
            result = kama.update(_INPUT[i])
            self.assertAlmostEqual(_EXPECTED[i], result, delta=1e-8,
                                   msg=f"[{i}] expected {_EXPECTED[i]}, got {result}")

        # NaN passthrough
        self.assertTrue(math.isnan(kama.update(math.nan)))

    def test_is_primed(self):
        """Priming requires erLength=10 samples."""
        params = KaufmanAdaptiveMovingAverageLengthParams(
            efficiency_ratio_length=10, fastest_length=2, slowest_length=30)
        kama = KaufmanAdaptiveMovingAverage.from_length(params)

        self.assertFalse(kama.is_primed())

        for i in range(10):
            kama.update(_INPUT[i])
            self.assertFalse(kama.is_primed(), f"[{i+1}] should not be primed")

        for i in range(10, len(_INPUT)):
            kama.update(_INPUT[i])
            self.assertTrue(kama.is_primed(), f"[{i+1}] should be primed")

    def test_metadata_length(self):
        """Metadata for length-based construction."""
        params = KaufmanAdaptiveMovingAverageLengthParams(
            efficiency_ratio_length=10, fastest_length=2, slowest_length=30)
        kama = KaufmanAdaptiveMovingAverage.from_length(params)
        m = kama.metadata()

        self.assertEqual(Identifier.KAUFMAN_ADAPTIVE_MOVING_AVERAGE, m.identifier)
        self.assertEqual("kama(10, 2, 30)", m.mnemonic)
        self.assertEqual("Kaufman adaptive moving average kama(10, 2, 30)", m.description)
        self.assertEqual(1, len(m.outputs))
        self.assertEqual("kama(10, 2, 30)", m.outputs[0].mnemonic)

    def test_metadata_alpha(self):
        """Metadata for smoothing-factor-based construction."""
        params = KaufmanAdaptiveMovingAverageSmoothingFactorParams(
            efficiency_ratio_length=10,
            fastest_smoothing_factor=0.666666666,
            slowest_smoothing_factor=0.064516129)
        kama = KaufmanAdaptiveMovingAverage.from_smoothing_factor(params)
        m = kama.metadata()

        self.assertEqual("kama(10, 0.6667, 0.0645)", m.mnemonic)
        self.assertEqual("Kaufman adaptive moving average kama(10, 0.6667, 0.0645)", m.description)

    def test_metadata_non_default_bar_component(self):
        """Mnemonic includes non-default bar component."""
        params = KaufmanAdaptiveMovingAverageLengthParams(
            efficiency_ratio_length=10, fastest_length=2, slowest_length=30,
            bar_component=BarComponent.MEDIAN)
        kama = KaufmanAdaptiveMovingAverage.from_length(params)
        m = kama.metadata()

        self.assertEqual("kama(10, 2, 30, hl/2)", m.mnemonic)

    def test_metadata_non_default_quote_component(self):
        """Mnemonic includes non-default quote component via alpha constructor."""
        params = KaufmanAdaptiveMovingAverageSmoothingFactorParams(
            efficiency_ratio_length=10,
            fastest_smoothing_factor=2.0 / 3.0,
            slowest_smoothing_factor=2.0 / 31.0,
            quote_component=QuoteComponent.BID)
        kama = KaufmanAdaptiveMovingAverage.from_smoothing_factor(params)
        m = kama.metadata()

        self.assertEqual("kama(10, 0.6667, 0.0645, b)", m.mnemonic)

    def test_invalid_params_length(self):
        """Invalid length params raise ValueError."""
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_length(
                KaufmanAdaptiveMovingAverageLengthParams(
                    efficiency_ratio_length=1, fastest_length=2, slowest_length=30))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_length(
                KaufmanAdaptiveMovingAverageLengthParams(
                    efficiency_ratio_length=10, fastest_length=1, slowest_length=30))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_length(
                KaufmanAdaptiveMovingAverageLengthParams(
                    efficiency_ratio_length=10, fastest_length=2, slowest_length=1))

    def test_invalid_params_alpha(self):
        """Invalid alpha params raise ValueError."""
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_smoothing_factor(
                KaufmanAdaptiveMovingAverageSmoothingFactorParams(
                    efficiency_ratio_length=1))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_smoothing_factor(
                KaufmanAdaptiveMovingAverageSmoothingFactorParams(
                    fastest_smoothing_factor=-0.00000001))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_smoothing_factor(
                KaufmanAdaptiveMovingAverageSmoothingFactorParams(
                    fastest_smoothing_factor=1.00000001))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_smoothing_factor(
                KaufmanAdaptiveMovingAverageSmoothingFactorParams(
                    slowest_smoothing_factor=-0.00000001))
        with self.assertRaises(ValueError):
            KaufmanAdaptiveMovingAverage.from_smoothing_factor(
                KaufmanAdaptiveMovingAverageSmoothingFactorParams(
                    slowest_smoothing_factor=1.00000001))

    def test_update_entity(self):
        """Entity update methods return Scalar output."""
        params = KaufmanAdaptiveMovingAverageLengthParams(
            efficiency_ratio_length=10, fastest_length=2, slowest_length=30)
        kama = KaufmanAdaptiveMovingAverage.from_length(params)
        t = datetime(2021, 4, 1)

        # Prime with 10 zeros
        for _ in range(10):
            kama.update(0.0)

        inp = 3.0
        expected = 1.3333333333333328

        # Scalar
        kama2 = KaufmanAdaptiveMovingAverage.from_length(
            KaufmanAdaptiveMovingAverageLengthParams(
                efficiency_ratio_length=10, fastest_length=2, slowest_length=30))
        for _ in range(10):
            kama2.update(0.0)
        out = kama2.update_scalar(Scalar(time=t, value=inp))
        self.assertEqual(1, len(out))
        self.assertEqual(expected, out[0].value)

        # Bar
        kama3 = KaufmanAdaptiveMovingAverage.from_length(
            KaufmanAdaptiveMovingAverageLengthParams(
                efficiency_ratio_length=10, fastest_length=2, slowest_length=30))
        for _ in range(10):
            kama3.update(0.0)
        out = kama3.update_bar(Bar(time=t, open=0, high=0, low=0, close=inp, volume=0))
        self.assertEqual(1, len(out))
        self.assertEqual(expected, out[0].value)

        # Quote
        kama4 = KaufmanAdaptiveMovingAverage.from_length(
            KaufmanAdaptiveMovingAverageLengthParams(
                efficiency_ratio_length=10, fastest_length=2, slowest_length=30))
        for _ in range(10):
            kama4.update(0.0)
        out = kama4.update_quote(Quote(time=t, bid_price=inp, ask_price=inp, bid_size=0, ask_size=0))
        self.assertEqual(1, len(out))
        self.assertEqual(expected, out[0].value)

        # Trade
        kama5 = KaufmanAdaptiveMovingAverage.from_length(
            KaufmanAdaptiveMovingAverageLengthParams(
                efficiency_ratio_length=10, fastest_length=2, slowest_length=30))
        for _ in range(10):
            kama5.update(0.0)
        out = kama5.update_trade(Trade(time=t, price=inp, volume=0))
        self.assertEqual(1, len(out))
        self.assertEqual(expected, out[0].value)


if __name__ == '__main__':
    unittest.main()
