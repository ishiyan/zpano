/** Describes parameters to create an instance of the Parabolic Stop And Reverse indicator. */
export interface ParabolicStopAndReverseParams {
  /**
   * Controls the initial direction and SAR value.
   *
   *  0  = Auto-detect direction using the first two bars (default).
   *  >0 = Force long at the specified SAR value.
   *  <0 = Force short at abs(startValue) as the initial SAR value.
   *
   * Default is 0.0.
   */
  startValue?: number;

  /**
   * A percent offset added/removed to the initial stop on short/long reversal.
   *
   * Default is 0.0.
   */
  offsetOnReverse?: number;

  /**
   * The initial acceleration factor for the long direction.
   *
   * Default is 0.02.
   */
  accelerationInitLong?: number;

  /**
   * The acceleration factor increment for the long direction.
   *
   * Default is 0.02.
   */
  accelerationLong?: number;

  /**
   * The maximum acceleration factor for the long direction.
   *
   * Default is 0.20.
   */
  accelerationMaxLong?: number;

  /**
   * The initial acceleration factor for the short direction.
   *
   * Default is 0.02.
   */
  accelerationInitShort?: number;

  /**
   * The acceleration factor increment for the short direction.
   *
   * Default is 0.02.
   */
  accelerationShort?: number;

  /**
   * The maximum acceleration factor for the short direction.
   *
   * Default is 0.20.
   */
  accelerationMaxShort?: number;
}

export function defaultParams(): ParabolicStopAndReverseParams {
    return {
        startValue: 0,
        offsetOnReverse: 0,
        accelerationInitLong: 0.02,
        accelerationLong: 0.02,
        accelerationMaxLong: 0.20,
        accelerationInitShort: 0.02,
        accelerationShort: 0.02,
        accelerationMaxShort: 0.20,
    };
}
