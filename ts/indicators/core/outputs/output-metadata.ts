import { OutputType } from './output-type';

/** Describes a single indicator output. */
export interface OutputMetadata {
    /**
     * An identification of this indicator output.
     *
     * It is an integer representation of provided outputs enumeration of a related indicator.
     */
    kind: number;

    /** Identifies a data type of this indicator output. */
    type: OutputType;

    /** A short name (mnemonic) of this indicator output. */
    mnemonic: string;

    /** A description of this indicator output. */
    description: string;
}
