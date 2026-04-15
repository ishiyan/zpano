import { IndicatorType } from './indicator-type';

/** Contains all info needed to create an indicator. */
export interface IndicatorSpecification {
    /** Identifies an indicator type. */
    type: IndicatorType;

    /**
     * Describes parameters to create an indicator.
     *
     *  The concrete type is defined by the related indicator, which in turn is defined by the __type__ field.
     */
    parameters: any;

    /** Describes requested kinds of indicator outputs to calculate. */
    outputs: number[];
}
