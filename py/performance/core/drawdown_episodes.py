import collections
from dataclasses import dataclass
import math

from ...streaming_kbn import KleinKBNAccumulator


@dataclass(frozen=True)
class DrawdownEpisode:
    """
    A single high-water-mark drawdown episode.

    Attributes:
        depth:
            Maximum drawdown depth, expressed as a decimal return.
        from_idx:
            Index of the first underwater observation.
        trough_idx:
            Index of the deepest observation.
        to_idx:
            Recovery observation index if the episode recovered;
            otherwise the last observation index currently available.
        recovered:
            True if the high-water mark was recovered at ``to_idx``.
            False if the series ended while the drawdown was still open.
    """
    depth: float
    from_idx: int
    trough_idx: int
    to_idx: int
    recovered: bool

class DrawdownEpisodes:
    """
    Streaming high-water-mark drawdown episode tracker.

    The class consumes drawdown observations produced by
    ``HighWaterMarkDrawdown`` and maintains the corresponding
    drawdown episodes incrementally.

    Drawdowns are expected as percent returns, for example ``-2.5``
    for a 2.5% drawdown and ``0.0`` for an observation at a
    high-water mark.

    A drawdown episode begins with the first negative drawdown and
    remains open until a non-negative drawdown is observed.

    The episode is considered recovered when the drawdown reaches
    zero or becomes positive.
    """

    def __init__(self) -> None:
        self._episodes: list[DrawdownEpisode] = []

        # Current open drawdown episode.
        self._current_from: int | None = None
        self._current_trough: int | None = None
        self._current_depth: float = 0.0

        # Number of observations processed.
        self._count: int = 0

        # Running depth aggregates.
        self._sum_depth: KleinKBNAccumulator = KleinKBNAccumulator()
        self._sum_depth_squared: KleinKBNAccumulator = KleinKBNAccumulator()
        # Running episode length/peak-to-trough/recovery aggregates
        self._sum_length: int = 0
        self._sum_peak_to_trough: int = 0
        self._sum_recovery: int = 0

    def reset(self) -> None:
        """Reset the episode tracker to its initial empty state."""
        self._episodes.clear()
        self._current_from = None
        self._current_trough = None
        self._current_depth = 0.0
        self._count = 0
        self._sum_depth.reset()
        self._sum_depth_squared.reset()
        self._sum_length = 0
        self._sum_peak_to_trough = 0
        self._sum_recovery = 0

    def update(self, drawdown: float) -> None:
        """
        Add one drawdown observation.

        Args:
            drawdown:
                Drawdown at the current observation, expressed as a
                percent return. Drawdowns must be non-positive, although
                non-negative values are accepted and treated as recovery
                or high-water-mark observations.
        """
        idx = self._count
        self._count += 1

        # Convert to decimals.
        #drawdown *= 0.01

        if drawdown < 0.0:
            # We are underwater.
            if self._current_from is None:
                # Start a new drawdown episode.
                self._current_from = idx
                self._current_trough = idx
                self._current_depth = drawdown
            elif drawdown < self._current_depth:
                # New trough within the current episode.
                self._current_trough = idx
                self._current_depth = drawdown
            return

        # We are at or above the high-water mark.
        if self._current_from is not None:
            # Close the current drawdown episode.
            self._episodes.append(DrawdownEpisode(depth=self._current_depth,
                from_idx=self._current_from, trough_idx=self._current_trough,
                to_idx=idx, recovered=True))
            self._sum_depth.update(self._current_depth)
            self._sum_depth_squared.update(self._current_depth * self._current_depth)
            self._sum_length += idx - self._current_from + 1
            self._sum_peak_to_trough += self._current_trough - self._current_from + 1
            self._sum_recovery += idx - self._current_trough + 1
            self._current_from = None
            self._current_trough = None
            self._current_depth = 0.0

    def recalculate(self, drawdowns: list[float]) -> None:
        """
        Rebuild all episodes from drawdown observations history.

        Args:
            drawdowns:
                Drawdown observations.
        """
        self.reset()
        for dd in drawdowns:
            self.update(dd)

    @property
    def episodes(self) -> list[DrawdownEpisode]:
        """
        Drawdown episodes currently known to the tracker.

        If the latest drawdown episode is still open, it is included
        using the last processed observation as ``to_idx`` and with
        ``recovered=False``.
        """
        episodes = list(self._episodes)
        if self._current_from is not None:
            episodes.append(DrawdownEpisode(depth=self._current_depth,
                from_idx=self._current_from, trough_idx=self._current_trough,
                to_idx=self._count - 1, recovered=False))
        return episodes

    @property
    def depths(self) -> list[float]:
        """
        Drawdown episode depths currently known to the tracker.

        If the latest drawdown episode is still open, it is included
        using the last processed observation as ``to_idx`` and with
        ``recovered=False``.
        """
        depths = list(episode.depth for episode in self._episodes)
        if self._current_from is not None:
            depths.append(self._current_depth)
        return depths

    @property
    def average_episode_drawdown(self) -> float:
        """
        The mean magnitude of the observed discrete episode drawdowns.
        """
        sum_depth = self._sum_depth.value
        count = len(self._episodes)
        if self._current_from is not None:
            sum_depth += self._current_depth
            count += 1
        return -sum_depth / count if count > 0 else 0.0

    @property
    def average_episode_drawdown_squared(self) -> float:
        """
        The mean of the observed discrete squared episode drawdowns.
        """
        sum_depth_squared = self._sum_depth_squared.value
        count = self._count
        if count == 0:
            return 0.0
        if self._current_from is not None:
            sum_depth_squared += self._current_depth * self._current_depth
        return sum_depth_squared / count

    @property
    def average_episode_length(self) -> float:
        """
        The mean length of the observed discrete drawdown episodes.
        """
        sum_length = self._sum_length
        count = len(self._episodes)
        if self._current_from is not None:
            sum_length += self._count - self._current_from
            count += 1
        return float(sum_length) / count if count > 0 else 0.0

    @property
    def average_episode_peak_to_trough(self) -> float:
        """
        The mean peak-to-trough length of the observed discrete drawdown episodes.
        """
        sum_peak_to_trough = self._sum_peak_to_trough
        count = len(self._episodes)
        if self._current_from is not None:
            sum_peak_to_trough += self._current_trough - self._current_from + 1
            count += 1
        return float(sum_peak_to_trough) / count if count > 0 else 0.0

    @property
    def average_episode_recovery(self) -> float:
        """
        The mean recovery length of the observed discrete drawdown episodes.
        """
        sum_recovery = self._sum_recovery
        count = len(self._episodes)
        if self._current_from is not None:
            sum_recovery += self._count - self._current_trough
            count += 1
        return float(sum_recovery) / count if count > 0 else 0.0
