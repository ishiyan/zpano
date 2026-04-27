"""Specification contains all info needed to create an indicator."""

from .identifier import Identifier


class Specification:
    """Contains all info needed to create an indicator."""

    __slots__ = ('identifier', 'parameters', 'outputs')

    def __init__(self, identifier: Identifier, parameters: object = None,
                 outputs: list[int] | None = None) -> None:
        self.identifier = identifier
        self.parameters = parameters
        self.outputs = outputs if outputs is not None else []
