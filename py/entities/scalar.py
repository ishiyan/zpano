"""A scalar (time and value) entity."""

from datetime import datetime


class Scalar:
    """Represents a timestamped scalar value."""

    def __init__(self, time: datetime, value: float):
        self.time = time
        self.value = value

    def __repr__(self) -> str:
        return f"Scalar({self.time}, {self.value})"
