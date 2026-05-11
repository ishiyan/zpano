/**
 * Band signals.
 *
 * Fuzzy membership for price/value relative to dynamic bands
 * (e.g., Bollinger Bands, Keltner Channels, Donchian).
 */
import { MembershipShape, muGreater, muLess, tProduct } from '../fuzzy/index.ts';

/** Degree to which value is above the upper band. */
export function muAboveBand(
    value: number, upperBand: number,
    width: number = 0.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    return muGreater(value, upperBand, width, shape);
}

/** Degree to which value is below the lower band. */
export function muBelowBand(
    value: number, lowerBand: number,
    width: number = 0.0, shape: MembershipShape = MembershipShape.SIGMOID
): number {
    return muLess(value, lowerBand, width, shape);
}

/**
 * Degree to which value is inside the band channel.
 * Computed as mu_above(value, lower) * mu_below(value, upper)
 * using the band spread as the transition width for both sides.
 */
export function muBetweenBands(
    value: number, lowerBand: number, upperBand: number,
    shape: MembershipShape = MembershipShape.SIGMOID
): number {
    if (upperBand <= lowerBand) return 0.0;
    const spread = upperBand - lowerBand;
    // Width = half the spread — gives a smooth transition at each band edge.
    const width = spread * 0.5;
    const aboveLower = muGreater(value, lowerBand, width, shape);
    const belowUpper = muLess(value, upperBand, width, shape);
    return tProduct(aboveLower, belowUpper);
}
