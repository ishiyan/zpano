/** Describes parameters to create an instance of the Advance-Decline indicator. */
export interface AdvanceDeclineParams {
  // Advance-Decline requires HLCV bar data and has no configurable parameters.
}

export function defaultParams(): AdvanceDeclineParams {
    return {};
}
