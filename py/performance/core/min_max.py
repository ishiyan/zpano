from collections import deque
import math

WINDOW_UNBOUNDED = 0

class MinMax:
    def __init__(self, window_size: int = WINDOW_UNBOUNDED) -> None:
        """
        The valid sliding `window_size` value is >= 1.
        We use zero value to indicate unbounded running window with infinite window size.
        """
        if window_size < 0:
            raise ValueError("window_size must be non-negative")

        # Zero window_size means no window, so we can just update the max value
        self._window_unbounded = window_size == WINDOW_UNBOUNDED

        if self._window_unbounded:
            self._min = math.inf
            self._max = -math.inf
        else:
            self._window_size = window_size
            self._index = 0
            # Monotonic deque for MIN (increasing order of values), stores (index, value)
            self._min_deque = deque()
            # Monotonic deque for MAX (decreasing order of values), stores (index, value)
            self._max_deque = deque()

    def reset(self) -> None:
        if self._window_unbounded:
            self._min = math.inf
            self._max = -math.inf
        else:
            self._index = 0
            self._min_deque.clear()
            self._max_deque.clear()

    def update(self, x: float) -> None:
        if self._window_unbounded:
            if self._min > x:
                self._min = x
            if self._max < x:
                self._max = x
            return

        # Remove out-of-window elements from the front
        cutoff = self._index - self._window_size
        
        while self._min_deque and self._min_deque[0][0] <= cutoff:
            self._min_deque.popleft()
        while self._max_deque and self._max_deque[0][0] <= cutoff:
            self._max_deque.popleft()

        # Maintain monotonic property for MIN (increasing)
        # Remove elements from back that are >= current value
        while self._min_deque and self._min_deque[-1][1] >= x:
            self._min_deque.pop()
        # Maintain monotonic property for MAX (decreasing)
        # Remove elements from back that are <= current value
        while self._max_deque and self._max_deque[-1][1] <= x:
            self._max_deque.pop()
            
        # Add current value with its index
        self._min_deque.append((self._index, x))
        self._max_deque.append((self._index, x))
        
        self._index += 1

    @property
    def min(self) -> float:
        """Returns the minimum value in the current window."""
        if self._window_unbounded:
            return self._min
        if not self._min_deque:
            return math.nan
        return self._min_deque[0][1]

    @property
    def max(self) -> float:
        """Returns the maximum value in the current window."""
        if self._window_unbounded:
            return self._max
        if not self._max_deque:
            return math.nan
        return self._max_deque[0][1]
