"""Output metadata describing a single indicator output."""

from .shape import Shape


class OutputMetadata:
    """Describes a single indicator output."""

    __slots__ = ('kind', 'shape', 'mnemonic', 'description')

    def __init__(self, kind: int, shape: Shape, mnemonic: str, description: str) -> None:
        self.kind = kind
        self.shape = shape
        self.mnemonic = mnemonic
        self.description = description

    def __repr__(self) -> str:
        return (f"OutputMetadata(kind={self.kind}, shape={self.shape.name}, "
                f"mnemonic={self.mnemonic!r}, description={self.description!r})")
