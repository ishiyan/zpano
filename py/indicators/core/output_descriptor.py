"""Output descriptor classifying a single indicator output for charting/discovery."""

from .outputs.shape import Shape
from .role import Role
from .pane import Pane


class OutputDescriptor:
    """Classifies a single indicator output for charting / discovery."""

    __slots__ = ('kind', 'shape', 'role', 'pane')

    def __init__(self, kind: int, shape: Shape, role: Role, pane: Pane) -> None:
        self.kind = kind
        self.shape = shape
        self.role = role
        self.pane = pane

    def __repr__(self) -> str:
        return (f"OutputDescriptor(kind={self.kind}, shape={self.shape.name}, "
                f"role={self.role.name}, pane={self.pane.name})")
