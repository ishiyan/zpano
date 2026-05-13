"""Aggregator for weighted blending of multiple signal sources.

Combines n independent signal sources (each producing values in [0, 1])
into a single blended confidence using one of seven aggregation methods.
Supports delayed feedback and online weight learning.
"""

import math
from collections import deque

from .method import AggregationMethod
from .error_metric import ErrorMetric


class Aggregator:
    """Weighted signal ensemble aggregator.

    Blends multiple independent signal sources into a single confidence
    value in [0, 1]. Adaptive methods update weights based on observed
    outcomes with a configurable feedback delay.

    Parameters
    ----------
    n_signals : int
        Number of signal sources (>= 1).
    method : AggregationMethod
        Aggregation method to use.
    feedback_delay : int
        Number of bars between signal observation and outcome availability (>= 1).
    weights : list[float] | None
        Required for FIXED method. Normalized to sum to 1.0.
    window : int
        Rolling window size for INVERSE_VARIANCE and RANK_BASED (>= 2).
    alpha : float
        Decay rate for EXPONENTIAL_DECAY (0 < alpha <= 1).
    eta : float
        Learning rate for MULTIPLICATIVE_WEIGHTS (> 0).
    prior : list[float] | None
        Prior weights for BAYESIAN. Defaults to uniform.
    error_metric : ErrorMetric
        Error metric for INVERSE_VARIANCE and RANK_BASED.
    """

    def __init__(
        self,
        n_signals: int,
        method: AggregationMethod = AggregationMethod.EQUAL,
        feedback_delay: int = 1,
        *,
        weights: list[float] | None = None,
        window: int = 50,
        alpha: float = 0.1,
        eta: float = 0.5,
        prior: list[float] | None = None,
        error_metric: ErrorMetric = ErrorMetric.ABSOLUTE,
    ) -> None:
        if n_signals < 1:
            raise ValueError(f"n_signals must be >= 1, got {n_signals}")
        if feedback_delay < 1:
            raise ValueError(f"feedback_delay must be >= 1, got {feedback_delay}")

        self._n = n_signals
        self._method = method
        self._feedback_delay = feedback_delay
        self._count = 0
        self._ring: deque[list[float]] = deque(maxlen=feedback_delay + 1)

        # Initialize method-specific state and weights.
        if method == AggregationMethod.FIXED:
            if weights is None:
                raise ValueError("FIXED method requires weights")
            if len(weights) != n_signals:
                raise ValueError(
                    f"weights length {len(weights)} != n_signals {n_signals}"
                )
            s = sum(weights)
            if s <= 0:
                raise ValueError("weights must sum to a positive value")
            self._weights = [w / s for w in weights]

        elif method == AggregationMethod.EQUAL:
            self._weights = [1.0 / n_signals] * n_signals

        elif method == AggregationMethod.INVERSE_VARIANCE:
            if window < 2:
                raise ValueError(f"window must be >= 2, got {window}")
            self._window = window
            self._error_metric = error_metric
            self._errors: list[deque[float]] = [
                deque(maxlen=window) for _ in range(n_signals)
            ]
            self._weights = [1.0 / n_signals] * n_signals

        elif method == AggregationMethod.EXPONENTIAL_DECAY:
            if alpha <= 0 or alpha > 1:
                raise ValueError(f"alpha must be in (0, 1], got {alpha}")
            self._alpha = alpha
            self._ema = [0.5] * n_signals  # neutral prior
            self._weights = [1.0 / n_signals] * n_signals

        elif method == AggregationMethod.MULTIPLICATIVE_WEIGHTS:
            if eta <= 0:
                raise ValueError(f"eta must be > 0, got {eta}")
            self._eta = eta
            self._log_weights = [0.0] * n_signals  # uniform in log-space
            self._weights = [1.0 / n_signals] * n_signals

        elif method == AggregationMethod.RANK_BASED:
            if window < 2:
                raise ValueError(f"window must be >= 2, got {window}")
            self._window = window
            self._error_metric = error_metric
            self._errors: list[deque[float]] = [
                deque(maxlen=window) for _ in range(n_signals)
            ]
            self._weights = [1.0 / n_signals] * n_signals

        elif method == AggregationMethod.BAYESIAN:
            if prior is not None:
                if len(prior) != n_signals:
                    raise ValueError(
                        f"prior length {len(prior)} != n_signals {n_signals}"
                    )
                s = sum(prior)
                if s <= 0:
                    raise ValueError("prior must sum to a positive value")
                normalized_prior = [p / s for p in prior]
            else:
                normalized_prior = [1.0 / n_signals] * n_signals
            self._log_posterior = [math.log(p) for p in normalized_prior]
            self._weights = list(normalized_prior)

        else:
            raise ValueError(f"unknown method: {method}")

    def blend(self, signals: list[float]) -> float:
        """Blend signal sources into a single confidence value.

        Parameters
        ----------
        signals : list[float]
            Signal values in [0, 1], one per source.

        Returns
        -------
        float
            Blended confidence in [0, 1].
        """
        if len(signals) != self._n:
            raise ValueError(
                f"expected {self._n} signals, got {len(signals)}"
            )

        output = sum(self._weights[i] * signals[i] for i in range(self._n))
        self._ring.append(list(signals))
        self._count += 1
        return output

    def update(self, outcome: float) -> None:
        """Provide outcome feedback for weight adaptation.

        For stateless methods (FIXED, EQUAL), this is a no-op.
        For adaptive methods, pairs the outcome with the buffered signals
        from feedback_delay bars ago and updates weights.

        Parameters
        ----------
        outcome : float
            Observed outcome in [0, 1].
        """
        if self._method in (AggregationMethod.FIXED, AggregationMethod.EQUAL):
            return

        if len(self._ring) < self._feedback_delay + 1:
            return

        # Retrieve signals from feedback_delay bars ago.
        idx = len(self._ring) - 1 - self._feedback_delay
        past_signals = self._ring[idx]

        if self._method == AggregationMethod.INVERSE_VARIANCE:
            self._update_inverse_variance(past_signals, outcome)
        elif self._method == AggregationMethod.EXPONENTIAL_DECAY:
            self._update_exponential_decay(past_signals, outcome)
        elif self._method == AggregationMethod.MULTIPLICATIVE_WEIGHTS:
            self._update_multiplicative_weights(past_signals, outcome)
        elif self._method == AggregationMethod.RANK_BASED:
            self._update_rank_based(past_signals, outcome)
        elif self._method == AggregationMethod.BAYESIAN:
            self._update_bayesian(past_signals, outcome)

    def warmup(self, history: list[tuple[list[float], float]]) -> None:
        """Replay historical data through blend() + update().

        Each tuple contains (signals_at_bar_T, outcome_for_bar_T).
        The method handles the feedback delay internally.

        Parameters
        ----------
        history : list[tuple[list[float], float]]
            Historical signal/outcome pairs.
        """
        outcomes: list[float] = []
        for signals, outcome in history:
            self.blend(signals)
            outcomes.append(outcome)
            idx = len(outcomes) - 1 - self._feedback_delay
            if idx >= 0:
                self.update(outcomes[idx])

    @property
    def weights(self) -> list[float]:
        """Current normalized weights (read-only copy)."""
        return list(self._weights)

    @property
    def count(self) -> int:
        """Total number of blend() calls."""
        return self._count

    # ── Private update methods ──────────────────────────────────────────

    def _compute_error(self, signal: float, outcome: float) -> float:
        """Compute per-signal error using the configured metric."""
        diff = signal - outcome
        if self._error_metric == ErrorMetric.ABSOLUTE:
            return abs(diff)
        else:  # SQUARED
            return diff * diff

    def _update_inverse_variance(
        self, signals: list[float], outcome: float
    ) -> None:
        """Update weights using inverse-variance of prediction errors."""
        epsilon = 1e-15

        for i in range(self._n):
            error = self._compute_error(signals[i], outcome)
            self._errors[i].append(error)

        # Need at least 2 errors to compute variance.
        if any(len(self._errors[i]) < 2 for i in range(self._n)):
            return

        raw = []
        for i in range(self._n):
            errors = self._errors[i]
            n = len(errors)
            mean = sum(errors) / n
            var = sum((e - mean) ** 2 for e in errors) / n  # population variance
            raw.append(1.0 / max(var, epsilon))

        total = sum(raw)
        self._weights = [r / total for r in raw]

    def _update_exponential_decay(
        self, signals: list[float], outcome: float
    ) -> None:
        """Update weights using EMA of accuracy."""
        for i in range(self._n):
            error = abs(signals[i] - outcome)
            accuracy = 1.0 - error
            self._ema[i] = self._alpha * accuracy + (1.0 - self._alpha) * self._ema[i]

        # Normalize, clamping negative EMAs to 0.
        clamped = [max(e, 0.0) for e in self._ema]
        total = sum(clamped)
        if total > 0:
            self._weights = [c / total for c in clamped]
        else:
            self._weights = [1.0 / self._n] * self._n

    def _update_multiplicative_weights(
        self, signals: list[float], outcome: float
    ) -> None:
        """Update weights using the Hedge algorithm in log-space."""
        for i in range(self._n):
            loss = abs(signals[i] - outcome)
            self._log_weights[i] -= self._eta * loss

        # Softmax normalization (log-sum-exp trick).
        max_log = max(self._log_weights)
        exp_weights = [math.exp(lw - max_log) for lw in self._log_weights]
        total = sum(exp_weights)
        self._weights = [e / total for e in exp_weights]

    def _update_rank_based(
        self, signals: list[float], outcome: float
    ) -> None:
        """Update weights using rank of rolling accuracy."""
        for i in range(self._n):
            error = self._compute_error(signals[i], outcome)
            self._errors[i].append(error)

        # Need at least 1 error per signal.
        if any(len(self._errors[i]) < 1 for i in range(self._n)):
            return

        # Compute mean accuracy per signal.
        accuracies = []
        for i in range(self._n):
            mean_error = sum(self._errors[i]) / len(self._errors[i])
            accuracies.append(1.0 - mean_error)

        # Rank by accuracy (best = highest rank = n, worst = 1).
        # Ties get the average rank.
        ranks = self._rank_with_ties(accuracies)

        total = sum(ranks)
        if total > 0:
            self._weights = [r / total for r in ranks]
        else:
            self._weights = [1.0 / self._n] * self._n

    @staticmethod
    def _rank_with_ties(values: list[float]) -> list[float]:
        """Rank values from 1 (worst) to n (best), averaging ties."""
        n = len(values)
        # Sort indices by value.
        sorted_indices = sorted(range(n), key=lambda i: values[i])
        ranks = [0.0] * n

        i = 0
        while i < n:
            # Find the end of the tie group.
            j = i + 1
            while j < n and values[sorted_indices[j]] == values[sorted_indices[i]]:
                j += 1
            # Average rank for this group (1-based).
            avg_rank = (i + 1 + j) / 2.0
            for k in range(i, j):
                ranks[sorted_indices[k]] = avg_rank
            i = j

        return ranks

    def _update_bayesian(
        self, signals: list[float], outcome: float
    ) -> None:
        """Update weights using Bayesian model averaging (Bernoulli likelihood)."""
        epsilon = 1e-15

        for i in range(self._n):
            # Clamp signal to [epsilon, 1 - epsilon] to avoid log(0).
            s = max(epsilon, min(1.0 - epsilon, signals[i]))
            log_lik = outcome * math.log(s) + (1.0 - outcome) * math.log(1.0 - s)
            self._log_posterior[i] += log_lik

        # Softmax normalization.
        max_log = max(self._log_posterior)
        exp_weights = [math.exp(lp - max_log) for lp in self._log_posterior]
        total = sum(exp_weights)
        self._weights = [e / total for e in exp_weights]
