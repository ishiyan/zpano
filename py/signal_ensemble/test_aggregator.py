"""Tests for signal ensemble aggregator."""

import math
import unittest

from py.signal_ensemble.method import AggregationMethod
from py.signal_ensemble.error_metric import ErrorMetric
from py.signal_ensemble.aggregator import Aggregator


class TestAggregatorValidation(unittest.TestCase):
    """Constructor validation tests."""

    def test_n_signals_zero(self) -> None:
        with self.assertRaises(ValueError):
            Aggregator(n_signals=0, method=AggregationMethod.EQUAL)

    def test_n_signals_negative(self) -> None:
        with self.assertRaises(ValueError):
            Aggregator(n_signals=-1, method=AggregationMethod.EQUAL)

    def test_feedback_delay_zero(self) -> None:
        with self.assertRaises(ValueError):
            Aggregator(n_signals=2, feedback_delay=0)

    def test_fixed_requires_weights(self) -> None:
        with self.assertRaises(ValueError):
            Aggregator(n_signals=2, method=AggregationMethod.FIXED)

    def test_fixed_weights_wrong_length(self) -> None:
        with self.assertRaises(ValueError):
            Aggregator(n_signals=2, method=AggregationMethod.FIXED, weights=[1.0])

    def test_fixed_weights_zero_sum(self) -> None:
        with self.assertRaises(ValueError):
            Aggregator(n_signals=2, method=AggregationMethod.FIXED, weights=[0.0, 0.0])

    def test_inverse_variance_window_too_small(self) -> None:
        with self.assertRaises(ValueError):
            Aggregator(n_signals=2, method=AggregationMethod.INVERSE_VARIANCE, window=1)

    def test_exponential_decay_alpha_zero(self) -> None:
        with self.assertRaises(ValueError):
            Aggregator(n_signals=2, method=AggregationMethod.EXPONENTIAL_DECAY, alpha=0)

    def test_exponential_decay_alpha_negative(self) -> None:
        with self.assertRaises(ValueError):
            Aggregator(n_signals=2, method=AggregationMethod.EXPONENTIAL_DECAY, alpha=-0.1)

    def test_multiplicative_weights_eta_zero(self) -> None:
        with self.assertRaises(ValueError):
            Aggregator(n_signals=2, method=AggregationMethod.MULTIPLICATIVE_WEIGHTS, eta=0)

    def test_bayesian_prior_wrong_length(self) -> None:
        with self.assertRaises(ValueError):
            Aggregator(n_signals=2, method=AggregationMethod.BAYESIAN, prior=[1.0])

    def test_blend_wrong_signal_count(self) -> None:
        agg = Aggregator(n_signals=3, method=AggregationMethod.EQUAL)
        with self.assertRaises(ValueError):
            agg.blend([0.5, 0.5])


class TestFixedWeights(unittest.TestCase):
    """Tests for FIXED aggregation method."""

    def test_basic_blend(self) -> None:
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.FIXED,
            weights=[0.5, 0.3, 0.2],
        )
        result = agg.blend([1.0, 0.0, 0.0])
        self.assertAlmostEqual(result, 0.5, places=13)

    def test_weights_normalized(self) -> None:
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.FIXED,
            weights=[2.0, 8.0],
        )
        w = agg.weights
        self.assertAlmostEqual(w[0], 0.2, places=13)
        self.assertAlmostEqual(w[1], 0.8, places=13)

    def test_update_is_noop(self) -> None:
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.FIXED,
            weights=[0.6, 0.4],
        )
        agg.blend([0.8, 0.2])
        agg.blend([0.7, 0.3])
        w_before = agg.weights
        agg.update(0.9)
        self.assertEqual(agg.weights, w_before)

    def test_blend_all_ones(self) -> None:
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.FIXED,
            weights=[0.5, 0.3, 0.2],
        )
        result = agg.blend([1.0, 1.0, 1.0])
        self.assertAlmostEqual(result, 1.0, places=13)

    def test_blend_all_zeros(self) -> None:
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.FIXED,
            weights=[0.5, 0.3, 0.2],
        )
        result = agg.blend([0.0, 0.0, 0.0])
        self.assertAlmostEqual(result, 0.0, places=13)

    def test_count_increments(self) -> None:
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.FIXED,
            weights=[0.5, 0.5],
        )
        self.assertEqual(agg.count, 0)
        agg.blend([0.5, 0.5])
        self.assertEqual(agg.count, 1)
        agg.blend([0.5, 0.5])
        self.assertEqual(agg.count, 2)


class TestEqualWeights(unittest.TestCase):
    """Tests for EQUAL aggregation method."""

    def test_basic_blend(self) -> None:
        agg = Aggregator(n_signals=3, method=AggregationMethod.EQUAL)
        result = agg.blend([0.9, 0.3, 0.6])
        self.assertAlmostEqual(result, 0.6, places=13)

    def test_single_signal(self) -> None:
        agg = Aggregator(n_signals=1, method=AggregationMethod.EQUAL)
        result = agg.blend([0.7])
        self.assertAlmostEqual(result, 0.7, places=13)

    def test_weights_are_uniform(self) -> None:
        agg = Aggregator(n_signals=4, method=AggregationMethod.EQUAL)
        for w in agg.weights:
            self.assertAlmostEqual(w, 0.25, places=13)

    def test_update_is_noop(self) -> None:
        agg = Aggregator(n_signals=2, method=AggregationMethod.EQUAL)
        agg.blend([0.8, 0.2])
        agg.blend([0.7, 0.3])
        w_before = agg.weights
        agg.update(0.5)
        self.assertEqual(agg.weights, w_before)


class TestInverseVariance(unittest.TestCase):
    """Tests for INVERSE_VARIANCE aggregation method."""

    def test_initial_weights_uniform(self) -> None:
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.INVERSE_VARIANCE,
            feedback_delay=1,
            window=10,
        )
        for w in agg.weights:
            self.assertAlmostEqual(w, 1.0 / 3, places=13)

    def test_accurate_signal_gets_higher_weight(self) -> None:
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.INVERSE_VARIANCE,
            feedback_delay=1,
            window=10,
        )
        # Signal 0 is always close to outcome, signal 1 is erratic.
        outcomes = [0.5, 0.6, 0.4, 0.55, 0.45, 0.5, 0.6, 0.4, 0.55, 0.45]
        for i, outcome in enumerate(outcomes):
            # Signal 0 = outcome + tiny noise, signal 1 = random.
            s0 = outcome + 0.01 * ((-1) ** i)
            s1 = 0.9 if i % 2 == 0 else 0.1
            agg.blend([s0, s1])
            agg.update(outcome)

        w = agg.weights
        self.assertGreater(w[0], w[1])

    def test_squared_error_metric(self) -> None:
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.INVERSE_VARIANCE,
            feedback_delay=1,
            window=10,
            error_metric=ErrorMetric.SQUARED,
        )
        for i in range(5):
            agg.blend([0.5, 0.5])
            agg.update(0.5)
        # Both signals identical — weights should be equal.
        w = agg.weights
        self.assertAlmostEqual(w[0], w[1], places=10)


class TestExponentialDecay(unittest.TestCase):
    """Tests for EXPONENTIAL_DECAY aggregation method."""

    def test_initial_weights_uniform(self) -> None:
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.EXPONENTIAL_DECAY,
            feedback_delay=1,
        )
        for w in agg.weights:
            self.assertAlmostEqual(w, 1.0 / 3, places=13)

    def test_good_signal_weight_increases(self) -> None:
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.EXPONENTIAL_DECAY,
            feedback_delay=1,
            alpha=0.3,
        )
        # Signal 0 is always accurate, signal 1 always wrong.
        for _ in range(20):
            agg.blend([0.8, 0.2])
            agg.update(0.8)

        w = agg.weights
        self.assertGreater(w[0], w[1])

    def test_alpha_one(self) -> None:
        """Alpha=1 means only the most recent observation matters."""
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.EXPONENTIAL_DECAY,
            feedback_delay=1,
            alpha=1.0,
        )
        agg.blend([0.9, 0.1])
        agg.blend([0.9, 0.1])
        agg.update(0.9)
        w = agg.weights
        # Signal 0 accuracy = 1 - |0.9-0.9| = 1.0
        # Signal 1 accuracy = 1 - |0.1-0.9| = 0.2
        self.assertAlmostEqual(w[0], 1.0 / 1.2, places=13)
        self.assertAlmostEqual(w[1], 0.2 / 1.2, places=13)


class TestMultiplicativeWeights(unittest.TestCase):
    """Tests for MULTIPLICATIVE_WEIGHTS (Hedge) aggregation method."""

    def test_initial_weights_uniform(self) -> None:
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.MULTIPLICATIVE_WEIGHTS,
            feedback_delay=1,
        )
        for w in agg.weights:
            self.assertAlmostEqual(w, 1.0 / 3, places=13)

    def test_best_signal_converges(self) -> None:
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.MULTIPLICATIVE_WEIGHTS,
            feedback_delay=1,
            eta=0.5,
        )
        # Signal 0 is perfect, signals 1 and 2 are bad.
        for _ in range(50):
            agg.blend([0.8, 0.2, 0.3])
            agg.update(0.8)

        w = agg.weights
        self.assertGreater(w[0], 0.5)
        self.assertGreater(w[0], w[1])
        self.assertGreater(w[0], w[2])

    def test_high_eta_faster_convergence(self) -> None:
        """Higher eta means faster adaptation."""
        results = []
        for eta in [0.1, 1.0]:
            agg = Aggregator(
                n_signals=2,
                method=AggregationMethod.MULTIPLICATIVE_WEIGHTS,
                feedback_delay=1,
                eta=eta,
            )
            for _ in range(10):
                agg.blend([0.9, 0.1])
                agg.update(0.9)
            results.append(agg.weights[0])

        # Higher eta should give higher weight to the good signal after same rounds.
        self.assertGreater(results[1], results[0])


class TestRankBased(unittest.TestCase):
    """Tests for RANK_BASED aggregation method."""

    def test_initial_weights_uniform(self) -> None:
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.RANK_BASED,
            feedback_delay=1,
            window=10,
        )
        for w in agg.weights:
            self.assertAlmostEqual(w, 1.0 / 3, places=13)

    def test_rank_ordering(self) -> None:
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.RANK_BASED,
            feedback_delay=1,
            window=10,
        )
        # Signal 0: best, signal 1: medium, signal 2: worst.
        for _ in range(15):
            agg.blend([0.7, 0.5, 0.2])
            agg.update(0.7)

        w = agg.weights
        self.assertGreater(w[0], w[1])
        self.assertGreater(w[1], w[2])

    def test_ties_get_average_rank(self) -> None:
        """When two signals have identical accuracy, they should get equal weight."""
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.RANK_BASED,
            feedback_delay=1,
            window=10,
        )
        for _ in range(5):
            agg.blend([0.5, 0.5])
            agg.update(0.5)

        w = agg.weights
        self.assertAlmostEqual(w[0], w[1], places=13)


class TestBayesian(unittest.TestCase):
    """Tests for BAYESIAN aggregation method."""

    def test_uniform_prior(self) -> None:
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.BAYESIAN,
            feedback_delay=1,
        )
        for w in agg.weights:
            self.assertAlmostEqual(w, 1.0 / 3, places=13)

    def test_custom_prior(self) -> None:
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.BAYESIAN,
            feedback_delay=1,
            prior=[0.5, 0.3, 0.2],
        )
        w = agg.weights
        self.assertAlmostEqual(w[0], 0.5, places=13)
        self.assertAlmostEqual(w[1], 0.3, places=13)
        self.assertAlmostEqual(w[2], 0.2, places=13)

    def test_good_predictor_dominates(self) -> None:
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.BAYESIAN,
            feedback_delay=1,
        )
        # Signal 0 predicts well, signal 1 poorly.
        for _ in range(20):
            agg.blend([0.9, 0.1])
            agg.update(0.9)

        w = agg.weights
        self.assertGreater(w[0], 0.9)

    def test_evidence_overrides_prior(self) -> None:
        """Even with unfavorable prior, strong evidence should win."""
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.BAYESIAN,
            feedback_delay=1,
            prior=[0.1, 0.9],  # prior favors signal 1
        )
        # But signal 0 is the accurate one.
        for _ in range(50):
            agg.blend([0.8, 0.2])
            agg.update(0.8)

        w = agg.weights
        self.assertGreater(w[0], w[1])


class TestDelayedFeedback(unittest.TestCase):
    """Tests for the delayed feedback mechanism."""

    def test_delay_1(self) -> None:
        """With delay=1, update() pairs with the previous blend()."""
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.EXPONENTIAL_DECAY,
            feedback_delay=1,
            alpha=1.0,
        )
        agg.blend([0.9, 0.1])   # buffered at index 0
        agg.blend([0.5, 0.5])   # buffered at index 1
        # Now ring has 2 entries; delay=1 means we pair with index 0.
        agg.update(0.9)
        w = agg.weights
        # Signal 0 had 0.9 vs outcome 0.9 → accuracy 1.0
        # Signal 1 had 0.1 vs outcome 0.9 → accuracy 0.2
        self.assertAlmostEqual(w[0], 1.0 / 1.2, places=13)
        self.assertAlmostEqual(w[1], 0.2 / 1.2, places=13)

    def test_delay_2(self) -> None:
        """With delay=2, update() pairs with blend() from 2 bars ago."""
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.EXPONENTIAL_DECAY,
            feedback_delay=2,
            alpha=1.0,
        )
        agg.blend([0.9, 0.1])   # bar 1
        agg.blend([0.5, 0.5])   # bar 2
        # Not enough history yet (need delay+1 = 3 entries).
        agg.update(0.9)
        # Weights should still be initial (update was no-op).
        for w in agg.weights:
            self.assertAlmostEqual(w, 0.5, places=13)

        agg.blend([0.3, 0.7])   # bar 3, now ring has 3 entries
        agg.update(0.9)         # pairs with bar 1 signals [0.9, 0.1]
        w = agg.weights
        self.assertAlmostEqual(w[0], 1.0 / 1.2, places=13)
        self.assertAlmostEqual(w[1], 0.2 / 1.2, places=13)

    def test_update_without_enough_history(self) -> None:
        """update() is a no-op when not enough signals are buffered."""
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.EXPONENTIAL_DECAY,
            feedback_delay=3,
            alpha=0.5,
        )
        agg.blend([0.5, 0.5])
        w_before = agg.weights
        agg.update(0.5)  # only 1 entry, need 4
        self.assertEqual(agg.weights, w_before)


class TestWarmup(unittest.TestCase):
    """Tests for the warmup method."""

    def test_warmup_equals_live_replay(self) -> None:
        """Warmup should produce the same result as sequential blend/update."""
        history = [
            ([0.8, 0.3], 0.7),
            ([0.6, 0.5], 0.5),
            ([0.9, 0.2], 0.8),
            ([0.7, 0.4], 0.6),
            ([0.5, 0.6], 0.4),
        ]

        # Method 1: warmup.
        agg1 = Aggregator(
            n_signals=2,
            method=AggregationMethod.EXPONENTIAL_DECAY,
            feedback_delay=1,
            alpha=0.2,
        )
        agg1.warmup(history)

        # Method 2: manual replay.
        agg2 = Aggregator(
            n_signals=2,
            method=AggregationMethod.EXPONENTIAL_DECAY,
            feedback_delay=1,
            alpha=0.2,
        )
        outcomes = []
        for signals, outcome in history:
            agg2.blend(signals)
            outcomes.append(outcome)
            idx = len(outcomes) - 1 - 1  # feedback_delay = 1
            if idx >= 0:
                agg2.update(outcomes[idx])

        self.assertEqual(agg1.count, agg2.count)
        for w1, w2 in zip(agg1.weights, agg2.weights):
            self.assertAlmostEqual(w1, w2, places=13)

    def test_warmup_with_delay_2(self) -> None:
        """Warmup with feedback_delay=2 should also match live replay."""
        history = [
            ([0.8, 0.3, 0.5], 0.7),
            ([0.6, 0.5, 0.4], 0.5),
            ([0.9, 0.2, 0.6], 0.8),
            ([0.7, 0.4, 0.3], 0.6),
            ([0.5, 0.6, 0.7], 0.4),
            ([0.4, 0.7, 0.2], 0.3),
        ]

        agg1 = Aggregator(
            n_signals=3,
            method=AggregationMethod.MULTIPLICATIVE_WEIGHTS,
            feedback_delay=2,
            eta=0.3,
        )
        agg1.warmup(history)

        agg2 = Aggregator(
            n_signals=3,
            method=AggregationMethod.MULTIPLICATIVE_WEIGHTS,
            feedback_delay=2,
            eta=0.3,
        )
        outcomes = []
        for signals, outcome in history:
            agg2.blend(signals)
            outcomes.append(outcome)
            idx = len(outcomes) - 1 - 2
            if idx >= 0:
                agg2.update(outcomes[idx])

        self.assertEqual(agg1.count, agg2.count)
        for w1, w2 in zip(agg1.weights, agg2.weights):
            self.assertAlmostEqual(w1, w2, places=13)

    def test_warmup_bayesian(self) -> None:
        """Warmup on Bayesian method produces calibrated weights."""
        history = [([0.9, 0.1], 0.9)] * 10 + [([0.9, 0.1], 0.9)] * 10

        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.BAYESIAN,
            feedback_delay=1,
        )
        agg.warmup(history)

        w = agg.weights
        self.assertGreater(w[0], 0.9)


class TestWeightsProperty(unittest.TestCase):
    """Tests for the weights property."""

    def test_weights_returns_copy(self) -> None:
        agg = Aggregator(n_signals=2, method=AggregationMethod.EQUAL)
        w = agg.weights
        w[0] = 999.0
        # Internal weights should be unaffected.
        self.assertAlmostEqual(agg.weights[0], 0.5, places=13)

    def test_weights_sum_to_one(self) -> None:
        for method in AggregationMethod:
            kwargs: dict = {}
            if method == AggregationMethod.FIXED:
                kwargs["weights"] = [0.3, 0.7]
            agg = Aggregator(n_signals=2, method=method, **kwargs)
            self.assertAlmostEqual(sum(agg.weights), 1.0, places=13,
                                   msg=f"weights don't sum to 1 for {method.name}")


class TestEdgeCases(unittest.TestCase):
    """Edge case tests."""

    def test_single_signal(self) -> None:
        """Single signal should return the signal itself for all methods."""
        for method in AggregationMethod:
            kwargs: dict = {}
            if method == AggregationMethod.FIXED:
                kwargs["weights"] = [1.0]
            agg = Aggregator(n_signals=1, method=method, **kwargs)
            result = agg.blend([0.73])
            self.assertAlmostEqual(result, 0.73, places=13,
                                   msg=f"single signal failed for {method.name}")

    def test_many_signals(self) -> None:
        """Large number of signals should work."""
        n = 100
        agg = Aggregator(n_signals=n, method=AggregationMethod.EQUAL)
        signals = [0.5] * n
        result = agg.blend(signals)
        self.assertAlmostEqual(result, 0.5, places=13)

    def test_extreme_signals(self) -> None:
        """Signals at boundary values 0.0 and 1.0."""
        agg = Aggregator(n_signals=2, method=AggregationMethod.EQUAL)
        result = agg.blend([0.0, 1.0])
        self.assertAlmostEqual(result, 0.5, places=13)

    def test_bayesian_extreme_signals(self) -> None:
        """Bayesian should handle signals at 0.0 and 1.0 without crashing."""
        agg = Aggregator(
            n_signals=2,
            method=AggregationMethod.BAYESIAN,
            feedback_delay=1,
        )
        agg.blend([0.0, 1.0])
        agg.blend([0.0, 1.0])
        agg.update(1.0)
        # Should not raise; weights should be valid.
        w = agg.weights
        self.assertAlmostEqual(sum(w), 1.0, places=13)

    def test_inverse_variance_identical_signals(self) -> None:
        """Identical signals should produce equal weights."""
        agg = Aggregator(
            n_signals=3,
            method=AggregationMethod.INVERSE_VARIANCE,
            feedback_delay=1,
            window=10,
        )
        for _ in range(5):
            agg.blend([0.5, 0.5, 0.5])
            agg.update(0.5)

        w = agg.weights
        for i in range(3):
            self.assertAlmostEqual(w[i], 1.0 / 3, places=10)


if __name__ == "__main__":
    unittest.main()
