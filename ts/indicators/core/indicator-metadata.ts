import { OutputMetadata } from './outputs/output-metadata';
import { IndicatorType } from './indicator-type';

/** Describes a type and requested outputs of an indicator. */
export interface IndicatorMetadata {
    /** Identifies a type this indicator. */
    type: IndicatorType;

    /** A short name (mnemonic) of this indicator. */
    mnemonic: string;

    /** A description of this indicator. */
    description: string;

    /** An array of metadata for individual requested outputs. */
    outputs: OutputMetadata[];
}
