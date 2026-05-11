package signals

import "zpano/fuzzy"

// MuAbove returns the degree to which value is above threshold.
// At value == threshold the membership is 0.5. The transition
// zone spans threshold ± width/2.
func MuAbove(value, threshold, width float64, shape fuzzy.MembershipShape) float64 {
	return fuzzy.MuGreater(value, threshold, width, shape)
}

// MuBelow returns the degree to which value is below threshold.
// Complement of MuAbove.
func MuBelow(value, threshold, width float64, shape fuzzy.MembershipShape) float64 {
	return fuzzy.MuLess(value, threshold, width, shape)
}

// MuOverbought returns the degree of overbought condition.
// Convenience alias for MuAbove with default level 70.
func MuOverbought(value, level, width float64, shape fuzzy.MembershipShape) float64 {
	return fuzzy.MuGreater(value, level, width, shape)
}

// MuOversold returns the degree of oversold condition.
// Convenience alias for MuBelow with default level 30.
func MuOversold(value, level, width float64, shape fuzzy.MembershipShape) float64 {
	return fuzzy.MuLess(value, level, width, shape)
}
