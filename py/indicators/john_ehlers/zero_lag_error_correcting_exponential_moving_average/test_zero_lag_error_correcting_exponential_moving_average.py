"""Tests for ZeroLagErrorCorrectingExponentialMovingAverage."""

from __future__ import annotations

from datetime import datetime, timedelta
import math
import unittest

from py.entities.bar import Bar
from py.entities.bar_component import BarComponent
from py.entities.quote import Quote
from py.entities.quote_component import QuoteComponent
from py.entities.scalar import Scalar
from py.entities.trade import Trade
from py.entities.trade_component import TradeComponent
from py.indicators.core.adaptivity import Adaptivity
from py.indicators.core.descriptors import descriptor_of
from py.indicators.core.identifier import Identifier
from py.indicators.core.outputs.shape import Shape
from py.indicators.factory.factory import create_indicator
from py.indicators.john_ehlers.zero_lag_error_correcting_exponential_moving_average import (
    ZeroLagErrorCorrectingExponentialMovingAverage,
    ZeroLagErrorCorrectingExponentialMovingAverageLengthParams,
    ZeroLagErrorCorrectingExponentialMovingAverageOutput,
    ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
    default_length_params,
    default_smoothing_factor_params,
    zero_lag_error_correcting_exponential_moving_average_ema_mnemonic,
    zero_lag_error_correcting_exponential_moving_average_mnemonic,
)
from py.indicators.john_ehlers.zero_lag_error_correcting_exponential_moving_average import (
    test_testdata as testdata,
)


def article_batch(
    prices: list[float], length: int, gain_limit: int
) -> list[tuple[float, float, float, float]]:
    alpha = 2.0 / (length + 1)
    ema = [0.0] * (len(prices) + 1)
    ec = [0.0] * (len(prices) + 1)
    best_gain = 0.0
    result: list[tuple[float, float, float, float]] = []
    for step, price in enumerate(prices, 1):
        ema[step] = alpha * price + (1.0 - alpha) * ema[step - 1]
        least_error = 1_000_000.0
        for integer_gain in range(-gain_limit, gain_limit + 1):
            gain = integer_gain / 10.0
            trial = (
                alpha * (ema[step] + gain * (price - ec[step - 1]))
                + (1.0 - alpha) * ec[step - 1]
            )
            error_value = price - trial
            if abs(error_value) < least_error:
                least_error = abs(error_value)
                best_gain = gain
        ec[step] = (
            alpha * (ema[step] + best_gain * (price - ec[step - 1]))
            + (1.0 - alpha) * ec[step - 1]
        )
        result.append((ec[step], ema[step], best_gain, least_error))
    return result


REFERENCE_CASES = (
    (20, 50, testdata.EXPECTED_L20_G50_EC, testdata.EXPECTED_L20_G50_EMA),
    (12, 50, testdata.EXPECTED_L12_G50_EC, testdata.EXPECTED_L12_G50_EMA),
    (32, 22, testdata.EXPECTED_L32_G22_EC, testdata.EXPECTED_L32_G22_EMA),
    (20, 0, testdata.EXPECTED_L20_G0_EC, testdata.EXPECTED_L20_G0_EMA),
    (20, 1, testdata.EXPECTED_L20_G1_EC, testdata.EXPECTED_L20_G1_EMA),
    (20, 10, testdata.EXPECTED_L20_G10_EC, testdata.EXPECTED_L20_G10_EMA),
    (20, 22, testdata.EXPECTED_L20_G22_EC, testdata.EXPECTED_L20_G22_EMA),
    (20, 100, testdata.EXPECTED_L20_G100_EC, testdata.EXPECTED_L20_G100_EMA),
    (1, 50, testdata.EXPECTED_L1_G50_EC, testdata.EXPECTED_L1_G50_EMA),
    (2, 50, testdata.EXPECTED_L2_G50_EC, testdata.EXPECTED_L2_G50_EMA),
    (5, 50, testdata.EXPECTED_L5_G50_EC, testdata.EXPECTED_L5_G50_EMA),
    (10, 50, testdata.EXPECTED_L10_G50_EC, testdata.EXPECTED_L10_G50_EMA),
    (40, 50, testdata.EXPECTED_L40_G50_EC, testdata.EXPECTED_L40_G50_EMA),
    (100, 50, testdata.EXPECTED_L100_G50_EC, testdata.EXPECTED_L100_G50_EMA),
    (252, 50, testdata.EXPECTED_L252_G50_EC, testdata.EXPECTED_L252_G50_EMA),
    (20, 49, testdata.EXPECTED_L20_G49_EC, testdata.EXPECTED_L20_G49_EMA),
)


class TestZeroLagErrorCorrectingExponentialMovingAverage(unittest.TestCase):
    def assert_close_or_nan(self, actual: float, expected: float, tol: float = 2e-12) -> None:
        if math.isnan(expected):
            self.assertTrue(math.isnan(actual))
        else:
            self.assertTrue(
                math.isclose(actual, expected, rel_tol=tol, abs_tol=tol),
                f"{actual!r} != {expected!r}",
            )

    def test_all_reference_fixtures_and_article_batch(self) -> None:
        base_time = datetime(2020, 1, 1)
        for length, gain_limit, expected_ec, expected_ema in REFERENCE_CASES:
            with self.subTest(length=length, gain_limit=gain_limit):
                params = ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(
                    length=length, gain_limit=gain_limit
                )
                stream = ZeroLagErrorCorrectingExponentialMovingAverage.from_length(params)
                entity_stream = ZeroLagErrorCorrectingExponentialMovingAverage.from_length(params)
                reference = article_batch(testdata.INPUT_CLOSE, length, gain_limit)
                for bar_index, price in enumerate(testdata.INPUT_CLOSE):
                    ec = stream.update(price)
                    ema = stream.ema()
                    self.assert_close_or_nan(ec, expected_ec[bar_index])
                    self.assert_close_or_nan(ema, expected_ema[bar_index])
                    self.assert_close_or_nan(ec, reference[bar_index][0])
                    self.assert_close_or_nan(ema, reference[bar_index][1])
                    self.assertEqual(stream.best_gain(), reference[bar_index][2])
                    self.assertEqual(stream.least_error(), reference[bar_index][3])

                    stamp = base_time + timedelta(days=bar_index)
                    entity_out = entity_stream.update_scalar(Scalar(time=stamp, value=price))
                    self.assertEqual(len(entity_out), 2)
                    self.assertEqual(entity_out[0].time, stamp)
                    self.assertEqual(entity_out[1].time, stamp)
                    self.assertEqual(entity_out[0].value, ec)
                    self.assertEqual(entity_out[1].value, ema)

    def test_zero_state_startup_and_first_observation_priming(self) -> None:
        stream = ZeroLagErrorCorrectingExponentialMovingAverage()
        self.assertFalse(stream.is_primed())
        self.assertTrue(math.isnan(stream.value()))
        self.assertTrue(math.isnan(stream.ema()))
        self.assertTrue(math.isnan(stream.best_gain()))
        self.assertTrue(math.isnan(stream.least_error()))
        self.assertEqual(stream.alpha(), 2.0 / 21.0)
        self.assertNotEqual(stream.alpha(), 0.095)
        self.assertEqual(stream.warmup_bars(), 0)

        price = testdata.INPUT_CLOSE[0]
        ec = stream.update(price)
        expected_ema = (2.0 / 21.0) * price
        expected_ec = (2.0 / 21.0) * (expected_ema + 5.0 * price)
        self.assertEqual(stream.ema(), expected_ema)
        self.assertEqual(ec, expected_ec)
        self.assertEqual(stream.value(), expected_ec)
        self.assertEqual(stream.best_gain(), 5.0)
        self.assertTrue(stream.is_primed())
        self.assertEqual(stream.samples_seen(), 1)

    def test_inclusive_integer_endpoints_frozen_previous_ec_and_ties(self) -> None:
        for limit in (1, 22, 49, 50):
            stream = ZeroLagErrorCorrectingExponentialMovingAverage(
                ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(
                    length=20, gain_limit=limit
                )
            )
            stream.update(testdata.INPUT_CLOSE[0])
            self.assertEqual(stream.best_gain(), limit / 10.0)

        frozen_stream = ZeroLagErrorCorrectingExponentialMovingAverage(
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(
                length=20, gain_limit=22
            )
        )
        for price in testdata.INPUT_CLOSE[:60]:
            previous_ec, previous_ema = frozen_stream._ec_state, frozen_stream._ema_state
            alpha = frozen_stream.alpha()
            ema = alpha * price + (1.0 - alpha) * previous_ema
            candidates = [
                (
                    alpha * (ema + (index / 10.0) * (price - previous_ec))
                    + (1.0 - alpha) * previous_ec,
                    index / 10.0,
                )
                for index in range(-22, 23)
            ]
            winner, gain = min(candidates, key=lambda pair: abs(price - pair[0]))
            ec = frozen_stream.update(price)
            self.assertEqual(ec, winner)
            self.assertEqual(frozen_stream.best_gain(), gain)
            self.assertEqual(frozen_stream.ema(), ema)

        tie_stream = ZeroLagErrorCorrectingExponentialMovingAverage(
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(
                length=20, gain_limit=49
            )
        )
        for _ in range(5):
            ec = tie_stream.update(0.0)
            self.assertEqual((ec, tie_stream.ema()), (0.0, 0.0))
            self.assertEqual(tie_stream.least_error(), 0.0)
            self.assertEqual(tie_stream.best_gain(), -4.9)

    def test_gain_zero_is_two_cascaded_emas_and_baseline_independent(self) -> None:
        stream = ZeroLagErrorCorrectingExponentialMovingAverage(
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(
                length=20, gain_limit=0
            )
        )
        alpha, complement = 2.0 / 21.0, 1.0 - 2.0 / 21.0
        for step in range(50):
            ec = stream.update(1.0 if step == 0 else 0.0)
            self.assertAlmostEqual(
                stream.ema(), alpha * complement**step, delta=1e-14
            )
            self.assertAlmostEqual(
                ec, (step + 1) * alpha * alpha * complement**step, delta=1e-14
            )
            self.assertEqual(stream.best_gain(), 0.0)
        self.assertNotEqual(stream.value(), stream.ema())

        streams = [
            ZeroLagErrorCorrectingExponentialMovingAverage(
                ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(
                    length=20, gain_limit=limit
                )
            )
            for limit in (0, 1, 22, 50, 100)
        ]
        ema = 0.0
        for price in testdata.INPUT_CLOSE:
            ema = alpha * price + complement * ema
            for candidate_stream in streams:
                candidate_stream.update(price)
                self.assertEqual(candidate_stream.ema(), ema)

    def test_finite_sentinel_alpha_one_long_length_nan_reset_and_overflow(self) -> None:
        sentinel_stream = ZeroLagErrorCorrectingExponentialMovingAverage()
        sentinel_stream.update(0.0)
        self.assertEqual(sentinel_stream.best_gain(), -5.0)
        ec = sentinel_stream.update(100_000_000.0)
        self.assertEqual(sentinel_stream.best_gain(), -5.0)
        self.assertEqual(sentinel_stream.least_error(), 1_000_000.0)
        self.assertNotEqual(sentinel_stream.least_error(), abs(100_000_000.0 - ec))
        self.assertLess(ec, 0.0)
        expected = article_batch([0.0, 100_000_000.0], 20, 50)[-1]
        self.assert_close_or_nan(ec, expected[0])
        self.assert_close_or_nan(sentinel_stream.ema(), expected[1])

        alpha_one = ZeroLagErrorCorrectingExponentialMovingAverage(
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(
                length=1, gain_limit=50
            )
        )
        for price in testdata.INPUT_CLOSE:
            self.assertEqual((alpha_one.update(price), alpha_one.ema()), (price, price))

        long_stream = ZeroLagErrorCorrectingExponentialMovingAverage(
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(
                length=252, gain_limit=50
            )
        )
        for price in testdata.INPUT_CLOSE:
            self.assertTrue(math.isfinite(long_stream.update(price)))
            self.assertTrue(math.isfinite(long_stream.ema()))
            self.assertTrue(long_stream.is_primed())

        nan_stream = ZeroLagErrorCorrectingExponentialMovingAverage()
        control = ZeroLagErrorCorrectingExponentialMovingAverage()
        for price in testdata.INPUT_CLOSE[:80]:
            self.assertEqual(nan_stream.update(price), control.update(price))
            self.assertEqual(nan_stream.ema(), control.ema())
            saved = (
                nan_stream.value(),
                nan_stream.ema(),
                nan_stream.samples_seen(),
                nan_stream.best_gain(),
            )
            self.assertTrue(math.isnan(nan_stream.update(math.nan)))
            self.assertEqual(
                (
                    nan_stream.value(),
                    nan_stream.ema(),
                    nan_stream.samples_seen(),
                    nan_stream.best_gain(),
                ),
                saved,
            )
        nan_stream.reset()
        self.assertFalse(nan_stream.is_primed())
        self.assertEqual(nan_stream.samples_seen(), 0)
        fresh = ZeroLagErrorCorrectingExponentialMovingAverage()
        for price in testdata.INPUT_CLOSE:
            self.assertEqual(nan_stream.update(price), fresh.update(price))
            self.assertEqual(nan_stream.ema(), fresh.ema())

        overflow_stream = ZeroLagErrorCorrectingExponentialMovingAverage()
        overflow_stream.update(testdata.INPUT_CLOSE[0])
        for bad_sample in (math.inf, -math.inf, None, True, "invalid", 1e308):
            snapshot = (
                overflow_stream.value(),
                overflow_stream.ema(),
                overflow_stream.samples_seen(),
                overflow_stream._ema_state,
                overflow_stream._ec_state,
                overflow_stream.best_gain(),
                overflow_stream.least_error(),
            )
            with self.subTest(bad_sample=bad_sample), self.assertRaises(ValueError):
                overflow_stream.update(bad_sample)  # type: ignore[arg-type]
            self.assertEqual(
                (
                    overflow_stream.value(),
                    overflow_stream.ema(),
                    overflow_stream.samples_seen(),
                    overflow_stream._ema_state,
                    overflow_stream._ec_state,
                    overflow_stream.best_gain(),
                    overflow_stream.least_error(),
                ),
                snapshot,
            )

    def test_smoothing_factor_variant_metadata_entities_and_factory(self) -> None:
        from_length = ZeroLagErrorCorrectingExponentialMovingAverage.from_length(
            default_length_params()
        )
        from_alpha = ZeroLagErrorCorrectingExponentialMovingAverage.from_smoothing_factor(
            default_smoothing_factor_params()
        )
        self.assertEqual(from_alpha.length(), 20)
        self.assertEqual(from_alpha.metadata().outputs[0].mnemonic, "zecema(20, 0.09523810, 50)")
        self.assertEqual(from_alpha.metadata().outputs[1].mnemonic, "zecemaEma(20, 0.09523810, 50)")
        for price in testdata.INPUT_CLOSE:
            self.assertEqual(from_alpha.update(price), from_length.update(price))
            self.assertEqual(from_alpha.ema(), from_length.ema())

        meta = from_length.metadata()
        desc = descriptor_of(Identifier.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE)
        self.assertIsNotNone(desc)
        assert desc is not None
        self.assertEqual(desc.adaptivity, Adaptivity.ADAPTIVE)
        self.assertEqual(meta.identifier, Identifier.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(len(meta.outputs), 2)
        self.assertEqual(meta.outputs[0].kind, ZeroLagErrorCorrectingExponentialMovingAverageOutput.VALUE)
        self.assertEqual(meta.outputs[0].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[0].mnemonic, "zecema(20, 50)")
        self.assertEqual(meta.outputs[1].kind, ZeroLagErrorCorrectingExponentialMovingAverageOutput.EMA)
        self.assertEqual(meta.outputs[1].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[1].mnemonic, "zecemaEma(20, 50)")

        stamp = datetime(2020, 1, 1)
        for entity_out in (
            ZeroLagErrorCorrectingExponentialMovingAverage().update_bar(
                Bar(time=stamp, open=100.0, high=101.0, low=99.0, close=100.0, volume=10.0)
            ),
            ZeroLagErrorCorrectingExponentialMovingAverage().update_quote(
                Quote(time=stamp, bid_price=100.0, ask_price=100.0, bid_size=1.0, ask_size=1.0)
            ),
            ZeroLagErrorCorrectingExponentialMovingAverage().update_trade(
                Trade(time=stamp, price=100.0, volume=1.0)
            ),
        ):
            self.assertEqual(len(entity_out), 2)
            self.assertTrue(math.isfinite(entity_out[0].value))
            self.assertTrue(math.isfinite(entity_out[1].value))

        self.assertEqual(
            zero_lag_error_correcting_exponential_moving_average_mnemonic(
                ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(
                    length=32,
                    gain_limit=22,
                    bar_component=BarComponent.OPEN,
                    quote_component=QuoteComponent.ASK,
                )
            ),
            "zecema(32, 22, o, a)",
        )
        self.assertEqual(
            zero_lag_error_correcting_exponential_moving_average_ema_mnemonic(
                ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams(
                    smoothing_factor=0.2,
                    gain_limit=22,
                    bar_component=BarComponent.TYPICAL,
                    trade_component=TradeComponent.VOLUME,
                )
            ),
            "zecemaEma(9, 0.20000000, 22, hlc/3, v)",
        )

        for bad_params in (
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(length=0),
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(length=True),  # type: ignore[arg-type]
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(length=20.0),  # type: ignore[arg-type]
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(length=10**30),
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(gain_limit=-1),
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(gain_limit=5.0),  # type: ignore[arg-type]
            ZeroLagErrorCorrectingExponentialMovingAverageLengthParams(gain_limit=True),  # type: ignore[arg-type]
            ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams(smoothing_factor=0.0),
            ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams(smoothing_factor=1.1),
            ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams(smoothing_factor=1e-30),
            ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams(smoothing_factor=True),  # type: ignore[arg-type]
        ):
            with self.subTest(bad_params=bad_params), self.assertRaises(ValueError):
                ZeroLagErrorCorrectingExponentialMovingAverage(bad_params)

        factory_default = create_indicator(
            Identifier.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE
        )
        self.assertEqual(factory_default.metadata().mnemonic, "zecema(20, 50)")
        factory_alpha = create_indicator(
            Identifier.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE,
            {"smoothingFactor": 0.2, "gainLimit": 22},
        )
        self.assertEqual(factory_alpha.metadata().mnemonic, "zecema(9, 0.20000000, 22)")
        for bad_dict in (
            {"gainStep": 0.1},
            {"length": 20, "smoothingFactor": 0.2},
            {"length": 0},
            {"gainLimit": -1},
        ):
            with self.assertRaises(ValueError):
                create_indicator(
                    Identifier.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE,
                    bad_dict,
                )


if __name__ == "__main__":
    unittest.main()
