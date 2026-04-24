/** Parameters of the Ultimate Oscillator indicator. */
export interface UltimateOscillatorParams {
  /** First time period (default 7). Minimum 2. */
  length1?: number;
  /** Second time period (default 14). Minimum 2. */
  length2?: number;
  /** Third time period (default 28). Minimum 2. */
  length3?: number;
}

export function defaultParams(): UltimateOscillatorParams {
    return { length1: 7, length2: 14, length3: 28 };
}
