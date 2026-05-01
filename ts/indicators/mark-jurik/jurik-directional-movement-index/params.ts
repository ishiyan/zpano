/** Parameters for the Jurik Directional Movement Index indicator. */
export interface JurikDirectionalMovementIndexParams {
    /** JMA smoothing length (minimum 1). */
    length: number;
}

export function defaultParams(): JurikDirectionalMovementIndexParams {
    return { length: 14 };
}
