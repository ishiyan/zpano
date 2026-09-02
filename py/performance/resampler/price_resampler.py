from __future__ import annotations

import math
from collections.abc import Callable
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import Optional


@dataclass(frozen=True)
class PriceReturn:
    """
    A regular-period return produced by ``PriceResampler``.

    Attributes:
        timestamp:
            Nominal end timestamp of the return period.

        return_:
            Portfolio/asset return over the period.

        benchmark_return:
            Benchmark return over the same period, if available.
    """
    timestamp: datetime
    return_: float
    benchmark_return: Optional[float] = None


class PriceResampler:
    """
    Convert irregular timestamped prices into regular-period returns.

    ``PriceResampler`` is a streaming preprocessing component. It is
    deliberately unaware of statistical measures, annualization, risk-free
    rates, or target returns.

    Observations are grouped according to ``boundary``. The last observed
    price in each period represents that period's closing price.

    A return is calculated between consecutive period closing prices::

        return = current_close / previous_close - 1

    No interpolation is performed.

    Example::

        resampler = PriceResampler(MinuteBoundary())

        resampler.update(
            datetime(2026, 8, 31, 10, 0, 10),
            100.0,
        )

        resampler.update(
            datetime(2026, 8, 31, 10, 0, 45),
            101.0,
        )

        result = resampler.update(
            datetime(2026, 8, 31, 10, 1, 20),
            102.0,
        )

    The first period closes at 10:01:00 with a price of 101.
    The observation at 10:01:20 belongs to the next period and therefore
    causes the first period to be completed. No return is emitted yet,
    because there is no previous period close.

    When the following period closes, a return is available.

    Args:
        boundary:
            Calendar boundary used to group observations.

        on_return:
            Optional callback invoked whenever a ``PriceReturn`` is produced.
    """

    def __init__(
        self,
        boundary: PeriodBoundary,
        on_return: Optional[Callable[[PriceReturn], None]] = None,
    ) -> None:
        self.boundary = boundary
        self._on_return = on_return

        self._period_start: Optional[datetime] = None
        self._period_end: Optional[datetime] = None

        self._price: Optional[float] = None
        self._benchmark_price: Optional[float] = None

        self._previous_period_price: Optional[float] = None
        self._previous_period_benchmark_price: Optional[float] = None

        self._last_timestamp: Optional[datetime] = None

    def reset(self) -> None:
        """
        Reset all accumulated state.
        """
        self._period_start = None
        self._period_end = None

        self._price = None
        self._benchmark_price = None

        self._previous_period_price = None
        self._previous_period_benchmark_price = None

        self._last_timestamp = None

    def update(self, timestamp: datetime, price: float, benchmark_price: Optional[float] = None) -> Optional[PriceReturn]:
        """
        Add one timestamped price observation.

        Args:
            timestamp:
                Timestamp of the observation.

            price:
                Asset or portfolio price.

            benchmark_price:
                Optional benchmark price at the same timestamp.

        Returns:
            The ``PriceReturn`` produced when this observation begins a
            new period, otherwise ``None``.

        Raises:
            ValueError:
                If timestamps are not chronological or a supplied price
                is not finite and strictly positive.
        """
        self._validate_price(price)

        if benchmark_price is not None:
            self._validate_price(benchmark_price)

        if (self._last_timestamp is not None and timestamp < self._last_timestamp):
            raise ValueError("timestamps must be supplied in chronological order")

        self._last_timestamp = timestamp
        period_start = self.boundary.period_start(timestamp)

        # First observation.
        if self._period_start is None:
            self._start_period(period_start, price, benchmark_price)
            return None

        # Observation belongs to the current period.
        if period_start == self._period_start:
            self._price = price
            if benchmark_price is not None:
                self._benchmark_price = benchmark_price
            return None

        # A new period has started, so the previous period is complete.
        result = self._complete_period()

        self._start_period(period_start, price, benchmark_price)
        return result

    def flush(self) -> Optional[PriceReturn]:
        """
        Complete the current period and emit its return.

        This is useful when the input stream ends without another
        observation starting a subsequent period.

        The returned period is potentially partial in the sense that the
        data stream may have ended before the nominal period boundary.
        """
        if self._period_start is None or self._price is None:
            return None

        result = self._make_return(timestamp=self._period_end, price=self._price, benchmark_price=self._benchmark_price)

        self._previous_period_price = self._price
        self._previous_period_benchmark_price = self._benchmark_price

        self._period_start = None
        self._period_end = None
        self._price = None
        self._benchmark_price = None
        return result

    def _start_period(self, period_start: datetime, price: float, benchmark_price: Optional[float]) -> None:
        self._period_start = period_start
        self._period_end = self.boundary.next_period_start(period_start)
        self._price = price
        self._benchmark_price = benchmark_price

    def _complete_period(self) -> Optional[PriceReturn]:
        if self._price is None or self._period_end is None:
            return None
        return self._make_return(timestamp=self._period_end, price=self._price, benchmark_price=self._benchmark_price)

    def _make_return(self, timestamp: datetime, price: float, benchmark_price: Optional[float]) -> Optional[PriceReturn]:
        if self._previous_period_price is None:
            return None

        return_ = price / self._previous_period_price - 1.0
        benchmark_return = None
        if (benchmark_price is not None and self._previous_period_benchmark_price is not None):
            benchmark_return = (benchmark_price / self._previous_period_benchmark_price - 1.0)

        result = PriceReturn(timestamp=timestamp, return_=return_, benchmark_return=benchmark_return)
        if self._on_return is not None:
            self._on_return(result)
        return result

    @staticmethod
    def _validate_price(price: float) -> None:
        if not math.isfinite(price):
            raise ValueError("price must be finite")
        if price <= 0:
            raise ValueError("price must be positive")
