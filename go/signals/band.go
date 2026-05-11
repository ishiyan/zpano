package signals

import "zpano/fuzzy"

// MuAboveBand returns the degree to which value is above the upper band.
func MuAboveBand(value, upperBand, width float64, shape fuzzy.MembershipShape) float64 {
	return fuzzy.MuGreater(value, upperBand, width, shape)
}

// MuBelowBand returns the degree to which value is below the lower band.
func MuBelowBand(value, lowerBand, width float64, shape fuzzy.MembershipShape) float64 {
	return fuzzy.MuLess(value, lowerBand, width, shape)
}

// MuBetweenBands returns the degree to which value is inside the band channel.
// Computed as mu_above(value, lower) * mu_below(value, upper)
// using the band spread as the transition width for both sides.
func MuBetweenBands(value, lowerBand, upperBand float64, shape fuzzy.MembershipShape) float64 {
	if upperBand <= lowerBand {
		return 0.0
	}
	spread := upperBand - lowerBand
	// Width = half the spread — gives a smooth transition at each band edge.
	width := spread * 0.5
	aboveLower := fuzzy.MuGreater(value, lowerBand, width, shape)
	belowUpper := fuzzy.MuLess(value, upperBand, width, shape)
	return fuzzy.TProduct(aboveLower, belowUpper)
}
