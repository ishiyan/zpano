package signals

import "zpano/fuzzy"

// MuTurnsPositive returns the degree to which a histogram turned from non-positive to positive.
// Equivalent to MuCrossesAbove(prev, curr, threshold=0, width).
func MuTurnsPositive(prevValue, currValue, width float64, shape fuzzy.MembershipShape) float64 {
	wasNonpositive := fuzzy.MuLess(prevValue, 0.0, width, shape)
	isPositive := fuzzy.MuGreater(currValue, 0.0, width, shape)
	return fuzzy.TProduct(wasNonpositive, isPositive)
}

// MuTurnsNegative returns the degree to which a histogram turned from non-negative to negative.
// Equivalent to MuCrossesBelow(prev, curr, threshold=0, width).
func MuTurnsNegative(prevValue, currValue, width float64, shape fuzzy.MembershipShape) float64 {
	wasNonnegative := fuzzy.MuGreater(prevValue, 0.0, width, shape)
	isNegative := fuzzy.MuLess(currValue, 0.0, width, shape)
	return fuzzy.TProduct(wasNonnegative, isNegative)
}
