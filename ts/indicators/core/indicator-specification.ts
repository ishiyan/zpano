import { IndicatorIdentifier } from './indicator-identifier';

/** Contains all info needed to create an indicator. */
export interface IndicatorSpecification {
    /** Identifies an indicator. */
    identifier: IndicatorIdentifier;

    /**
     * Describes parameters to create an indicator.
     *
     *  The concrete type is defined by the related indicator, which in turn is defined by the __identifier__ field.
     */
    parameters: any;

    /** Describes requested kinds of indicator outputs to calculate. */
    outputs: number[];
}
