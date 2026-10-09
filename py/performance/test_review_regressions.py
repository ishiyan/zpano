"""Regression coverage for the cross-language performance review fixes."""
import math
import unittest

from .measures import Measures


class TestReviewRegressions(unittest.TestCase):
    def test_gain_to_pain_is_independent_of_repetitions_and_window_length(self):
        for window in (0, 2, 4):
            m = Measures(periods_per_annum=1, rolling_window_size=window)
            self.assertTrue(math.isnan(m.gain_to_pain_ratio))
            for _ in range(4):
                m.add_return(0.1, 0)
                m.add_return(-0.05, 0)
                self.assertAlmostEqual(m.gain_to_pain_ratio, 1.0, places=13)
            m.reset()
            m.add_return(0.1, 0)
            self.assertTrue(math.isnan(m.gain_to_pain_ratio))

    def test_modified_information_ratio_uses_geometric_active_premium(self):
        m = Measures(periods_per_annum=1)
        m.add_return(0.5, 0.05)
        m.add_return(-0.3, 0.05)
        # Arithmetic active mean is +0.05; geometric active premium is negative.
        self.assertLess(m.active_premium, 0)
        self.assertAlmostEqual(m.information_ratio_modified,
                               0.04473320734100739, places=13)

    def test_invalid_arguments_raise_before_insufficient_data_fallback(self):
        m = Measures(periods_per_annum=1)
        for populated in (False, True):
            if populated:
                m.add_return(-0.1, 0)
                m.add_return(0.2, 0)
            for confidence in (-1e20, -1, 0, 1, 2, math.nan, math.inf, -math.inf):
                with self.subTest(populated=populated, confidence=confidence):
                    with self.assertRaisesRegex(ValueError, 'confidence must be between 0 and 1'):
                        m.is_normal_distribution(confidence)
                    with self.assertRaisesRegex(ValueError, 'confidence must be between 0 and 1'):
                        m.reward_to_conditional_drawdown(confidence)
            for multiplier in (0, -1, math.nan, -math.inf):
                with self.assertRaisesRegex(ValueError, 'std_dev_multiplier must be positive'):
                    m.bias_ratio(multiplier)
        self.assertTrue(math.isnan(Measures(1).reward_to_conditional_drawdown(0.95)))

    def test_documented_continuous_runs_and_zero_benchmark_down_number(self):
        m = Measures(periods_per_annum=1)
        for ret in (-0.01, -0.02, 0.01, -0.03, -0.04):
            m.add_return(ret, 0)
        for actual, expected in zip(m.drawdowns_continuous_runs(), (-0.029998, -0.069988)):
            self.assertAlmostEqual(actual, expected, places=13)
        self.assertAlmostEqual(m.down_number_ratio, 0.8, places=13)
