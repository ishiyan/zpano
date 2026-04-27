"""BuildMetadata utility for constructing indicator Metadata from the descriptor registry."""

from .identifier import Identifier
from .metadata import Metadata
from .outputs.metadata import OutputMetadata
from .descriptors import descriptor_of


class OutputText:
    """Provides the per-output mnemonic and description for BuildMetadata."""

    __slots__ = ('mnemonic', 'description')

    def __init__(self, mnemonic: str, description: str) -> None:
        self.mnemonic = mnemonic
        self.description = description


def build_metadata(identifier: Identifier, mnemonic: str, description: str,
                   texts: list[OutputText]) -> Metadata:
    """Constructs a Metadata for the indicator with the given identifier.

    Joins the registry's per-output Kind and Shape with the supplied
    per-output mnemonic and description.

    texts must be in the same order and length as the descriptor's Outputs.
    Raises ValueError if no descriptor is registered or if lengths don't match.
    """
    d = descriptor_of(identifier)
    if d is None:
        raise ValueError(
            f"build_metadata: no descriptor registered for identifier {identifier.name}")

    if len(texts) != len(d.outputs):
        raise ValueError(
            f"build_metadata: identifier {identifier.name} has {len(d.outputs)} outputs "
            f"in descriptor but {len(texts)} texts were supplied")

    out = []
    for i, t in enumerate(texts):
        out.append(OutputMetadata(
            kind=d.outputs[i].kind,
            shape=d.outputs[i].shape,
            mnemonic=t.mnemonic,
            description=t.description,
        ))

    return Metadata(
        identifier=identifier,
        mnemonic=mnemonic,
        description=description,
        outputs=out,
    )
