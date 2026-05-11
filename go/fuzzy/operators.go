package fuzzy

// TProduct computes the product t-norm: a * b.
// All conditions contribute proportionally. The default choice.
func TProduct(a, b float64) float64 {
	return a * b
}

// TMin computes the minimum t-norm (Zadeh): min(a, b).
// Dominated by the weakest condition.
func TMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// TLukasiewicz computes the Łukasiewicz t-norm: max(0, a + b - 1).
// Very strict — both conditions must have high membership.
func TLukasiewicz(a, b float64) float64 {
	v := a + b - 1.0
	if v < 0.0 {
		return 0.0
	}
	return v
}

// SProbabilistic computes the probabilistic sum: a + b - a*b.
// Dual of the product t-norm.
func SProbabilistic(a, b float64) float64 {
	return a + b - a*b
}

// SMax computes the maximum s-norm (Zadeh): max(a, b).
// Dual of the minimum t-norm.
func SMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// FNot computes the standard fuzzy negation: 1 - a.
func FNot(a float64) float64 {
	return 1.0 - a
}

// TProductAll computes the product t-norm over multiple arguments.
// Returns 1.0 for zero arguments (identity element of product).
func TProductAll(args ...float64) float64 {
	result := 1.0
	for _, a := range args {
		result *= a
	}
	return result
}

// TMinAll computes the minimum t-norm over multiple arguments.
// Returns 1.0 for zero arguments (identity element of min).
func TMinAll(args ...float64) float64 {
	if len(args) == 0 {
		return 1.0
	}
	result := args[0]
	for _, a := range args[1:] {
		if a < result {
			result = a
		}
	}
	return result
}
