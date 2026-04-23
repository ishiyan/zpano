import { IndicatorIdentifier } from './indicator-identifier.js';
import { IndicatorMetadata } from './indicator-metadata.js';
import { descriptorOf } from './descriptors.js';

/**
 * Per-output mnemonic and description used when building an IndicatorMetadata
 * from the descriptor registry.
 */
export interface OutputText {
  mnemonic: string;
  description: string;
}

/**
 * Constructs an IndicatorMetadata for the indicator with the given identifier
 * by joining the registry's per-output kind and shape with the supplied
 * per-output mnemonic and description.
 *
 * texts must be in the same order and length as the descriptor's outputs.
 * Throws if no descriptor is registered for the identifier or if the length
 * of texts does not match the descriptor's outputs.
 */
export function buildMetadata(
  identifier: IndicatorIdentifier,
  mnemonic: string,
  description: string,
  texts: OutputText[]
): IndicatorMetadata {
  const d = descriptorOf(identifier);
  if (!d) {
    throw new Error(`buildMetadata: no descriptor registered for identifier ${IndicatorIdentifier[identifier]}`);
  }
  if (texts.length !== d.outputs.length) {
    throw new Error(
      `buildMetadata: identifier ${IndicatorIdentifier[identifier]} has ${d.outputs.length} outputs in descriptor but ${texts.length} texts were supplied`
    );
  }
  return {
    identifier,
    mnemonic,
    description,
    outputs: texts.map((t, i) => ({
      kind: d.outputs[i].kind,
      shape: d.outputs[i].shape,
      mnemonic: t.mnemonic,
      description: t.description,
    })),
  };
}
