"""Descriptor classifies an indicator along multiple taxonomic dimensions."""

from .identifier import Identifier
from .adaptivity import Adaptivity
from .input_requirement import InputRequirement
from .volume_usage import VolumeUsage
from .output_descriptor import OutputDescriptor


class Descriptor:
    """Classifies an indicator along multiple taxonomic dimensions."""

    __slots__ = ('identifier', 'family', 'adaptivity', 'input_requirement',
                 'volume_usage', 'outputs')

    def __init__(self, identifier: Identifier, family: str, adaptivity: Adaptivity,
                 input_requirement: InputRequirement, volume_usage: VolumeUsage,
                 outputs: list[OutputDescriptor]) -> None:
        self.identifier = identifier
        self.family = family
        self.adaptivity = adaptivity
        self.input_requirement = input_requirement
        self.volume_usage = volume_usage
        self.outputs = outputs

    def __repr__(self) -> str:
        return (f"Descriptor(identifier={self.identifier.name}, family={self.family!r}, "
                f"adaptivity={self.adaptivity.name}, input_requirement={self.input_requirement.name}, "
                f"volume_usage={self.volume_usage.name}, outputs={self.outputs})")
