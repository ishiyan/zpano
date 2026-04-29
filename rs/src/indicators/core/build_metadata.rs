use super::descriptor::descriptor_of;
use super::identifier::Identifier;
use super::metadata::Metadata;
use super::outputs::metadata::OutputMetadata;

/// Per-output text supplied by the indicator implementation.
pub struct OutputText {
    pub mnemonic: String,
    pub description: String,
}

/// Constructs a `Metadata` by joining the descriptor registry's per-output
/// kind and shape with the supplied per-output mnemonic and description.
///
/// Panics if no descriptor is registered for the identifier or if the
/// number of texts does not match the descriptor's output count.
pub fn build_metadata(
    identifier: Identifier,
    mnemonic: &str,
    description: &str,
    texts: &[OutputText],
) -> Metadata {
    let d = descriptor_of(identifier)
        .unwrap_or_else(|| panic!("build_metadata: no descriptor registered for {:?}", identifier));

    assert_eq!(
        texts.len(),
        d.outputs.len(),
        "build_metadata: identifier {:?} has {} outputs in descriptor but {} texts were supplied",
        identifier,
        d.outputs.len(),
        texts.len(),
    );

    let outputs = d
        .outputs
        .iter()
        .zip(texts.iter())
        .map(|(od, t)| OutputMetadata {
            kind: od.kind,
            shape: od.shape,
            mnemonic: t.mnemonic.to_string(),
            description: t.description.to_string(),
        })
        .collect();

    Metadata {
        identifier,
        mnemonic: mnemonic.to_string(),
        description: description.to_string(),
        outputs,
    }
}
