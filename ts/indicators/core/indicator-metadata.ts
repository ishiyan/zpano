import { OutputMetadata } from './outputs/output-metadata';
import { IndicatorIdentifier } from './indicator-identifier';

/** Describes an indicator and its outputs. */
export interface IndicatorMetadata {
    /** Identifies this indicator. */
    identifier: IndicatorIdentifier;

    /** A short name (mnemonic) of this indicator. */
    mnemonic: string;

    /** A description of this indicator. */
    description: string;

    /** An array of metadata for individual requested outputs. */
    outputs: OutputMetadata[];
}
