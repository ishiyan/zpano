"""Indicator metadata describing an indicator and its outputs."""

from .identifier import Identifier
from .outputs.metadata import OutputMetadata


class Metadata:
    """Describes an indicator and its outputs."""

    __slots__ = ('identifier', 'mnemonic', 'description', 'outputs')

    def __init__(self, identifier: Identifier, mnemonic: str, description: str,
                 outputs: list[OutputMetadata]) -> None:
        self.identifier = identifier
        self.mnemonic = mnemonic
        self.description = description
        self.outputs = outputs

    def __repr__(self) -> str:
        return (f"Metadata(identifier={self.identifier.name}, mnemonic={self.mnemonic!r}, "
                f"description={self.description!r}, outputs={self.outputs})")
