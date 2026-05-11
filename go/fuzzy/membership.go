// Package fuzzy provides fuzzy logic primitives for membership, operators,
// and defuzzification.
package fuzzy

import "math"

// MembershipShape selects the transition curve for membership functions.
type MembershipShape int

const (
	// Sigmoid is a smooth logistic curve. Default for most applications.
	Sigmoid MembershipShape = iota
	// Linear is a piecewise-linear ramp (trapezoidal/triangular).
	Linear
)

// Steepness constant for sigmoid shape.
// k = sigmoidK / width gives ≈0.997 at threshold ± width/2.
const sigmoidK = 12.0

// sigmoid computes 1 / (1 + exp(k * (x - threshold))).
// Returns the "less-than" membership: high when x << threshold,
// low when x >> threshold, exactly 0.5 at x == threshold.
func sigmoid(x, threshold, k float64) float64 {
	exponent := k * (x - threshold)
	// Clamp to avoid overflow in exp().
	if exponent > 500.0 {
		return 0.0
	}
	if exponent < -500.0 {
		return 1.0
	}
	return 1.0 / (1.0 + math.Exp(exponent))
}

// MuLess returns the degree to which x is less than threshold.
//
// At threshold: μ = 0.5.
// At threshold - width/2: μ ≈ 0.997 (sigmoid) or 1.0 (linear).
// At threshold + width/2: μ ≈ 0.003 (sigmoid) or 0.0 (linear).
//
// When width = 0 (crisp): 1.0 if x < threshold, 0.5 if x == threshold,
// 0.0 if x > threshold.
func MuLess(x, threshold, width float64, shape MembershipShape) float64 {
	if width <= 0.0 {
		if x < threshold {
			return 1.0
		}
		if x > threshold {
			return 0.0
		}
		return 0.5
	}

	if shape == Linear {
		half := width * 0.5
		if x <= threshold-half {
			return 1.0
		}
		if x >= threshold+half {
			return 0.0
		}
		return (threshold + half - x) / width
	}
	// sigmoid
	return sigmoid(x, threshold, sigmoidK/width)
}

// MuLessEqual returns the degree to which x ≤ threshold.
// Identical to MuLess for continuous values — the distinction is conceptual.
func MuLessEqual(x, threshold, width float64, shape MembershipShape) float64 {
	return MuLess(x, threshold, width, shape)
}

// MuGreater returns the degree to which x > threshold. Complement of MuLess.
func MuGreater(x, threshold, width float64, shape MembershipShape) float64 {
	return 1.0 - MuLess(x, threshold, width, shape)
}

// MuGreaterEqual returns the degree to which x ≥ threshold. Complement of MuLessEqual.
func MuGreaterEqual(x, threshold, width float64, shape MembershipShape) float64 {
	return 1.0 - MuLessEqual(x, threshold, width, shape)
}

// MuNear returns a bell-shaped membership: degree to which x ≈ target.
//
// μ = 1.0 at x == target.
// μ ≈ 0 at |x - target| ≥ width.
//
// For sigmoid shape: Gaussian bell exp(-k * (x - target)²).
// For linear shape: triangular peak at target with base 2 * width.
func MuNear(x, target, width float64, shape MembershipShape) float64 {
	if width <= 0.0 {
		if x == target {
			return 1.0
		}
		return 0.0
	}

	if shape == Linear {
		dist := math.Abs(x - target)
		if dist >= width {
			return 0.0
		}
		return 1.0 - dist/width
	}
	// sigmoid → Gaussian bell
	// σ chosen so that μ ≈ 0.003 at |x - target| = width.
	sigma := width / 2.41
	d := (x - target) / sigma
	return math.Exp(-d * d)
}

// MuDirection returns a fuzzy candle direction ∈ [-1, +1].
//
// +1 = fully bullish (large white body).
//
//	0 = neutral (doji-like).
//
// -1 = fully bearish (large black body).
//
// Uses tanh(steepness * (c - o) / bodyAvg).
// When bodyAvg ≤ 0: returns +1.0 if c ≥ o, else -1.0 (crisp).
func MuDirection(o, c, bodyAvg, steepness float64) float64 {
	if bodyAvg <= 0.0 {
		if c >= o {
			return 1.0
		}
		return -1.0
	}
	return math.Tanh(steepness * (c - o) / bodyAvg)
}
