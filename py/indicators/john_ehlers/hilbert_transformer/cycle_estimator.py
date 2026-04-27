"""Cycle estimator abstract base class."""
from abc import ABC, abstractmethod


class CycleEstimator(ABC):
    """Describes a common Hilbert transformer cycle estimator functionality."""

    @abstractmethod
    def smoothing_length(self) -> int:
        """Returns the underlying WMA smoothing length in samples."""

    @abstractmethod
    def smoothed(self) -> float:
        """Returns the current WMA-smoothed value."""

    @abstractmethod
    def detrended(self) -> float:
        """Returns the current detrended value."""

    @abstractmethod
    def quadrature(self) -> float:
        """Returns the current Quadrature component value."""

    @abstractmethod
    def in_phase(self) -> float:
        """Returns the current InPhase component value."""

    @abstractmethod
    def period(self) -> float:
        """Returns the current period value."""

    @abstractmethod
    def count(self) -> int:
        """Returns the current count value."""

    @abstractmethod
    def primed(self) -> bool:
        """Indicates whether an instance is primed."""

    @abstractmethod
    def min_period(self) -> int:
        """Returns the minimal cycle period."""

    @abstractmethod
    def max_period(self) -> int:
        """Returns the maximal cycle period."""

    @abstractmethod
    def alpha_ema_quadrature_in_phase(self) -> float:
        """Returns alpha for EMA smoothing of in-phase and quadrature components."""

    @abstractmethod
    def alpha_ema_period(self) -> float:
        """Returns alpha for EMA smoothing of the instantaneous period."""

    @abstractmethod
    def warm_up_period(self) -> int:
        """Returns the number of updates before the estimator is primed."""

    @abstractmethod
    def update(self, sample: float) -> None:
        """Updates the estimator given the next sample value."""
