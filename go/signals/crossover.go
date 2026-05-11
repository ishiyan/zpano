package signals

import "zpano/fuzzy"

// MuCrossesAbove returns the degree to which a value crossed above threshold from below.
// Computed as mu_below(prev, threshold) * mu_above(curr, threshold).
func MuCrossesAbove(prevValue, currValue, threshold, width float64, shape fuzzy.MembershipShape) float64 {
	wasBelow := fuzzy.MuLess(prevValue, threshold, width, shape)
	isAbove := fuzzy.MuGreater(currValue, threshold, width, shape)
	return fuzzy.TProduct(wasBelow, isAbove)
}

// MuCrossesBelow returns the degree to which a value crossed below threshold from above.
// Computed as mu_above(prev, threshold) * mu_below(curr, threshold).
func MuCrossesBelow(prevValue, currValue, threshold, width float64, shape fuzzy.MembershipShape) float64 {
	wasAbove := fuzzy.MuGreater(prevValue, threshold, width, shape)
	isBelow := fuzzy.MuLess(currValue, threshold, width, shape)
	return fuzzy.TProduct(wasAbove, isBelow)
}

// MuLineCrossesAbove returns the degree to which a fast line crossed above a slow line.
// Reduces to a threshold crossing of (fast - slow) crossing above zero.
func MuLineCrossesAbove(prevFast, currFast, prevSlow, currSlow, width float64, shape fuzzy.MembershipShape) float64 {
	prevDiff := prevFast - prevSlow
	currDiff := currFast - currSlow
	return MuCrossesAbove(prevDiff, currDiff, 0.0, width, shape)
}

// MuLineCrossesBelow returns the degree to which a fast line crossed below a slow line.
func MuLineCrossesBelow(prevFast, currFast, prevSlow, currSlow, width float64, shape fuzzy.MembershipShape) float64 {
	prevDiff := prevFast - prevSlow
	currDiff := currFast - currSlow
	return MuCrossesBelow(prevDiff, currDiff, 0.0, width, shape)
}
