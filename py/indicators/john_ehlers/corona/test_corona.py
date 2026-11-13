"""Engine contract and edge tests for the shared Corona spectral analysis engine."""

from __future__ import annotations

import math
import unittest

from py.indicators.john_ehlers.corona import (
    AdaptiveBandPass,
    Corona,
    CoronaParams,
    Raster,
    default_params,
)
from py.indicators.john_ehlers.corona.corona import TWO_PI
from py.indicators.john_ehlers.corona_signal_to_noise_ratio import (
    CoronaSignalToNoiseRatio,
    CoronaSignalToNoiseRatioParams,
)
from py.indicators.john_ehlers.corona_spectrum import (
    CoronaSpectrum,
    CoronaSpectrumParams,
)
from py.indicators.john_ehlers.corona_spectrum import test_testdata as spectrum_data
from py.indicators.john_ehlers.corona_swing_position import (
    CoronaSwingPosition,
    CoronaSwingPositionParams,
)
from py.indicators.john_ehlers.corona_trend_vigor import (
    CoronaTrendVigor,
    CoronaTrendVigorParams,
)


def _same_values(actual, expected, tolerance: float = 5e-12) -> bool:
    if isinstance(expected, (tuple, list)):
        return (
            isinstance(actual, (tuple, list))
            and len(actual) == len(expected)
            and all(_same_values(a, b, tolerance) for a, b in zip(actual, expected))
        )
    if hasattr(expected, "values") and hasattr(actual, "values"):
        return (
            actual.time == expected.time
            and math.isclose(actual.parameter_first, expected.parameter_first, rel_tol=tolerance, abs_tol=tolerance)
            and math.isclose(actual.parameter_last, expected.parameter_last, rel_tol=tolerance, abs_tol=tolerance)
            and math.isclose(actual.parameter_resolution, expected.parameter_resolution, rel_tol=tolerance, abs_tol=tolerance)
            and _same_values(actual.value_min, expected.value_min, tolerance)
            and _same_values(actual.value_max, expected.value_max, tolerance)
            and _same_values(actual.values, expected.values, tolerance)
        )
    if not isinstance(expected, (int, float)):
        return actual == expected
    if math.isnan(expected):
        return math.isnan(actual)
    if math.isinf(expected):
        return actual == expected
    return math.isclose(actual, expected, rel_tol=tolerance, abs_tol=tolerance)


class TestCoronaEngine(unittest.TestCase):
    def test_startup_and_empty_selection_holds(self):
        engine = Corona(default_params())
        self.assertEqual(engine.filter_bank_length, 49)
        self.assertEqual(engine.minimal_period_times_two, 12)
        self.assertEqual(engine.maximal_period_times_two, 60)

        snapshots = [engine.update(price) for price in spectrum_data.INPUT_CLOSE]
        self.assertTrue(all(not s.primed and s.maximum_power == 0.0 for s in snapshots[:12]))
        self.assertTrue(snapshots[12].primed)

        holds = 0
        for previous_snapshot, current_snapshot in zip(snapshots, snapshots[1:]):
            if current_snapshot.selected_bins == 0:
                holds += 1
                self.assertEqual(
                    current_snapshot.dominant_cycle, previous_snapshot.dominant_cycle
                )
                self.assertFalse(current_snapshot.has_cycle_estimate)
        self.assertGreater(holds, 0)
        self.assertAlmostEqual(snapshots[-1].dominant_cycle, 11.5, delta=5e-12)

    def test_shared_and_standalone_consumers_match(self):
        engine = Corona()
        standalone = (
            CoronaSpectrum(),
            CoronaSwingPosition(),
            CoronaTrendVigor(),
            CoronaSignalToNoiseRatio(),
        )
        shared = (
            CoronaSpectrum(),
            CoronaSwingPosition(),
            CoronaTrendVigor(),
            CoronaSignalToNoiseRatio(),
        )
        for i, price in enumerate(spectrum_data.INPUT_CLOSE):
            snapshot = engine.update(price)
            for consumer_a, consumer_b in zip(standalone, shared):
                if isinstance(consumer_a, CoronaSignalToNoiseRatio):
                    high = spectrum_data.INPUT_HIGH[i]
                    low = spectrum_data.INPUT_LOW[i]
                    out_a = consumer_a.update(price, high=high, low=low)
                    out_b = consumer_b.update_from_snapshot(snapshot, high=high, low=low)
                else:
                    out_a = consumer_a.update(price)
                    out_b = consumer_b.update_from_snapshot(snapshot)
                self.assertTrue(_same_values(tuple(out_a), tuple(out_b)))

        self.assertEqual(engine.sample_count, 252)
        with self.assertRaises(ValueError):
            shared[0].update_from_snapshot(snapshot)

        incompatible = CoronaSpectrum(CoronaSpectrumParams(high_pass_filter_cutoff=48))
        fresh_engine = Corona()
        first_snapshot = fresh_engine.update(spectrum_data.INPUT_CLOSE[0])
        with self.assertRaises(ValueError):
            incompatible.update_from_snapshot(first_snapshot)

    def test_nan_skips_and_snapshot_owns_arrays(self):
        factories = (
            ("spectrum", CoronaSpectrum),
            ("swing", CoronaSwingPosition),
            ("trend", CoronaTrendVigor),
            ("snr", CoronaSignalToNoiseRatio),
        )
        for name, cls in factories:
            with self.subTest(indicator=name):
                first = cls()
                second = cls()
                kwargs = {"bar_range": 2.0} if name == "snr" else {}
                for i, price in enumerate(spectrum_data.INPUT_CLOSE[:50]):
                    if i == 25:
                        missing = first.update(math.nan, **kwargs)
                        self.assertTrue(missing.heatmap.is_empty())
                    out_first = first.update(price, **kwargs)
                    out_second = second.update(price, **kwargs)
                    self.assertTrue(_same_values(tuple(out_first), tuple(out_second)))
                out_first.heatmap.values[0] = -999.0
                out_first = first.update(102.0, **kwargs)
                out_second = second.update(102.0, **kwargs)
                self.assertTrue(_same_values(tuple(out_first), tuple(out_second)))

    def test_reject_bad_parameters_and_nonfinite_before_mutation(self):
        bad_configs = (
            CoronaParams(minimal_period=2),
            CoronaParams(minimal_period=0),
            CoronaParams(minimal_period=True),  # type: ignore[arg-type]
            CoronaParams(minimal_period=6.0),  # type: ignore[arg-type]
            CoronaParams(minimal_period=6, maximal_period=6),
            CoronaParams(high_pass_filter_cutoff=4),
            CoronaParams(high_pass_filter_cutoff=0),
            CoronaParams(decibels_lower_threshold=math.nan),
            CoronaParams(decibels_lower_threshold=-1.0),
            CoronaParams(decibels_lower_threshold=20.0, decibels_upper_threshold=20.0),
            CoronaParams(decibels_upper_threshold=math.inf),
        )
        for config in bad_configs:
            with self.assertRaises(ValueError):
                Corona(config)

        # Explicit 0.0 lower threshold is valid (no zero-means-default substitution).
        zero_lower = Corona(CoronaParams(decibels_lower_threshold=0.0))
        self.assertEqual(zero_lower.parameters.decibels_lower_threshold, 0.0)

        engine = Corona()
        for price in spectrum_data.INPUT_CLOSE[:30]:
            engine.update(price)
        before = engine.last_snapshot
        for bad_price in (math.inf, -math.inf, 1e308):
            with self.assertRaises(ValueError):
                engine.update(bad_price)
            self.assertIs(engine.last_snapshot, before)
            self.assertEqual(engine.sample_count, 30)

        engine.reset()
        self.assertEqual(engine.sample_count, 0)
        self.assertIsNone(engine.last_snapshot)

    def test_full_cycle_trend_cancels_sinusoid_and_preserves_linear_displacement(self):
        consumer = CoronaTrendVigor()
        engine = Corona()
        for i in range(120):
            price = 100.0 + 0.5 * i + 3.0 * math.sin(TWO_PI * i / 20.0)
            snapshot = engine.update(price)._replace(
                dominant_cycle=20.0, dominant_cycle_median=20.0
            )
            output = consumer.update_from_snapshot(snapshot)
            if i < 20:
                self.assertTrue(output.heatmap.is_empty())
            else:
                self.assertEqual(output.lookback, 20.0)
                self.assertAlmostEqual(output.trend_displacement, 10.0, delta=1e-12)
        self.assertTrue(consumer.is_primed())
        with self.assertRaises(ValueError):
            CoronaTrendVigor(CoronaTrendVigorParams(lookback_offset=0.0))  # type: ignore[arg-type]

    def test_phase_lead_uses_current_bandpass_and_full_precision_constant(self):
        bandpass = AdaptiveBandPass()
        previous_bp = 0.0
        for i in range(100):
            cycle = bandpass.update(100.0 + math.sin(TWO_PI * i / 20.0), 20.0)
            if i >= 2:
                self.assertAlmostEqual(
                    cycle.quadrature,
                    (cycle.band_pass - previous_bp) / (TWO_PI / 20.0),
                    delta=1e-14,
                )
                self.assertAlmostEqual(
                    cycle.lead60,
                    0.5 * cycle.band_pass + (math.sqrt(3.0) / 2.0) * cycle.quadrature,
                    delta=1e-14,
                )
            previous_bp = cycle.band_pass

    def test_constant_and_zero_prices_have_explicit_diagnostics(self):
        factories = (
            ("spectrum", CoronaSpectrum),
            ("swing", CoronaSwingPosition),
            ("trend", CoronaTrendVigor),
            ("snr", CoronaSignalToNoiseRatio),
        )
        for name, cls in factories:
            for price in (0.0, 100.0):
                instance = cls()
                for _ in range(100):
                    result = (
                        instance.update(price, bar_range=2.0)
                        if name == "snr"
                        else instance.update(price)
                    )
                self.assertFalse(result.heatmap.is_empty())
                self.assertEqual(result.has_cycle_estimate, 0.0)
                if name == "swing":
                    self.assertEqual(result.normalized_position, 0.5)
                    self.assertEqual(result.swing_position, 0.0)
                    self.assertTrue(all(v == 20.0 for v in result.heatmap.values))
                if name == "snr":
                    if price > 0.0:
                        self.assertEqual(result.physical_decibels, -math.inf)
                    else:
                        self.assertTrue(math.isnan(result.physical_decibels))
                    self.assertEqual(result.signal_to_noise_ratio, 1.0)

    def test_snr_physical_amplitude_ratio_and_range_validation(self):
        consumer = CoronaSignalToNoiseRatio()
        engine = Corona()
        for _ in range(350):
            snapshot = engine.update(100.0)._replace(maximum_power=4.0)
            output = consumer.update_from_snapshot(snapshot, bar_range=2.0)
        self.assertAlmostEqual(output.signal, 4.0, delta=1e-12)
        self.assertAlmostEqual(output.noise, 2.0, delta=1e-12)
        self.assertAlmostEqual(output.physical_decibels, 20.0 * math.log10(2.0), delta=1e-12)
        self.assertAlmostEqual(
            output.signal_to_noise_ratio, 1.0 + output.physical_decibels + 3.5, delta=1e-12
        )

        zero_noise = CoronaSignalToNoiseRatio()
        no_signal = CoronaSignalToNoiseRatio()
        engine = Corona()
        for _ in range(50):
            snapshot = engine.update(100.0)
            zero_out = zero_noise.update_from_snapshot(
                snapshot._replace(maximum_power=4.0), bar_range=0.0
            )
            none_out = no_signal.update_from_snapshot(snapshot, bar_range=0.0)
        self.assertEqual(zero_out.physical_decibels, math.inf)
        self.assertEqual(zero_out.signal_to_noise_ratio, 11.0)
        self.assertTrue(math.isnan(none_out.physical_decibels))
        self.assertEqual(none_out.signal_to_noise_ratio, 1.0)

        instance = CoronaSignalToNoiseRatio()
        for kwargs in (
            {},
            {"high": 99.0, "low": 101.0},
            {"bar_range": -1.0},
            {"high": math.inf, "low": 100.0},
            {"high": 101.0, "low": 99.0, "bar_range": 2.0},
        ):
            with self.assertRaises(ValueError):
                instance.update(100.0, **kwargs)
            self.assertFalse(instance.is_primed())

        self.assertTrue(instance.update(100.0, high=math.nan, low=99.0).heatmap.is_empty())

    def test_raster_geometry_marker_endpoint_and_half_up_rounding(self):
        raster = Raster(50, -5.0, 5.0, 20.0)
        raster.update(1.0, 0.1, "swing")
        column = raster.column(None)
        self.assertAlmostEqual(column.parameter_first, -4.8, delta=1e-14)
        self.assertAlmostEqual(column.parameter_last, 5.0, delta=1e-14)
        self.assertAlmostEqual(column.parameter_resolution, 5.0, delta=1e-14)
        self.assertEqual(column.values[-1], 0.0)
        self.assertGreater(column.values[0], 0.0)

        raster = Raster(50, -5.0, 5.0, 20.0)
        raster.values = [8.0] * 50
        raster.update(0.01, 0.1, "swing")
        self.assertEqual(raster.values[0], 4.0)


if __name__ == "__main__":
    unittest.main()
