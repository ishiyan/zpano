package fuzzy

import (
	"testing"
)

// -- T-norms --

func TestTNorms(t *testing.T) {
	t.Run("product_basic", func(t *testing.T) {
		if !almostEqual(TProduct(0.8, 0.6), 0.48, 1e-10) {
			t.Error("expected 0.48")
		}
	})
	t.Run("product_identity", func(t *testing.T) {
		if !almostEqual(TProduct(0.7, 1.0), 0.7, 1e-10) {
			t.Error("expected 0.7")
		}
	})
	t.Run("product_annihilator", func(t *testing.T) {
		if !almostEqual(TProduct(0.7, 0.0), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("product_commutativity", func(t *testing.T) {
		if !almostEqual(TProduct(0.3, 0.8), TProduct(0.8, 0.3), 1e-10) {
			t.Error("expected commutative")
		}
	})
	t.Run("min_basic", func(t *testing.T) {
		if TMin(0.8, 0.6) != 0.6 {
			t.Error("expected 0.6")
		}
	})
	t.Run("min_identity", func(t *testing.T) {
		if TMin(0.7, 1.0) != 0.7 {
			t.Error("expected 0.7")
		}
	})
	t.Run("min_annihilator", func(t *testing.T) {
		if TMin(0.7, 0.0) != 0.0 {
			t.Error("expected 0.0")
		}
	})
	t.Run("lukasiewicz_both_high", func(t *testing.T) {
		if !almostEqual(TLukasiewicz(0.9, 0.8), 0.7, 1e-10) {
			t.Error("expected 0.7")
		}
	})
	t.Run("lukasiewicz_one_low", func(t *testing.T) {
		if !almostEqual(TLukasiewicz(0.3, 0.5), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("lukasiewicz_clamp", func(t *testing.T) {
		if TLukasiewicz(0.1, 0.2) != 0.0 {
			t.Error("expected 0.0")
		}
	})
	t.Run("lukasiewicz_identity", func(t *testing.T) {
		if !almostEqual(TLukasiewicz(0.7, 1.0), 0.7, 1e-10) {
			t.Error("expected 0.7")
		}
	})
}

// -- S-norms --

func TestSNorms(t *testing.T) {
	t.Run("probabilistic_basic", func(t *testing.T) {
		if !almostEqual(SProbabilistic(0.8, 0.6), 0.92, 1e-10) {
			t.Error("expected 0.92")
		}
	})
	t.Run("probabilistic_identity", func(t *testing.T) {
		if !almostEqual(SProbabilistic(0.7, 0.0), 0.7, 1e-10) {
			t.Error("expected 0.7")
		}
	})
	t.Run("probabilistic_annihilator", func(t *testing.T) {
		if !almostEqual(SProbabilistic(0.7, 1.0), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("max_basic", func(t *testing.T) {
		if SMax(0.8, 0.6) != 0.8 {
			t.Error("expected 0.8")
		}
	})
	t.Run("max_identity", func(t *testing.T) {
		if SMax(0.7, 0.0) != 0.7 {
			t.Error("expected 0.7")
		}
	})
}

// -- Negation --

func TestNegation(t *testing.T) {
	t.Run("not_basic", func(t *testing.T) {
		if !almostEqual(FNot(0.3), 0.7, 1e-10) {
			t.Error("expected 0.7")
		}
	})
	t.Run("not_zero", func(t *testing.T) {
		if !almostEqual(FNot(0.0), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("not_one", func(t *testing.T) {
		if !almostEqual(FNot(1.0), 0.0, 1e-10) {
			t.Error("expected 0.0")
		}
	})
	t.Run("double_negation", func(t *testing.T) {
		if !almostEqual(FNot(FNot(0.4)), 0.4, 1e-10) {
			t.Error("expected 0.4")
		}
	})
}

// -- Variadic --

func TestVariadic(t *testing.T) {
	t.Run("product_all_three", func(t *testing.T) {
		if !almostEqual(TProductAll(0.8, 0.6, 0.5), 0.24, 1e-10) {
			t.Error("expected 0.24")
		}
	})
	t.Run("product_all_single", func(t *testing.T) {
		if !almostEqual(TProductAll(0.7), 0.7, 1e-10) {
			t.Error("expected 0.7")
		}
	})
	t.Run("product_all_empty", func(t *testing.T) {
		if !almostEqual(TProductAll(), 1.0, 1e-10) {
			t.Error("expected 1.0")
		}
	})
	t.Run("min_all_three", func(t *testing.T) {
		if TMinAll(0.8, 0.6, 0.9) != 0.6 {
			t.Error("expected 0.6")
		}
	})
	t.Run("min_all_empty", func(t *testing.T) {
		if TMinAll() != 1.0 {
			t.Error("expected 1.0")
		}
	})
	t.Run("product_all_five", func(t *testing.T) {
		result := TProductAll(0.9, 0.9, 0.9, 0.9, 0.9)
		expected := 0.9 * 0.9 * 0.9 * 0.9 * 0.9
		if !almostEqual(result, expected, 1e-10) {
			t.Errorf("expected %f, got %f", expected, result)
		}
	})
}

// -- Duality --

func TestDuality(t *testing.T) {
	t.Run("product_probabilistic_duality", func(t *testing.T) {
		a, b := 0.7, 0.4
		lhs := TProduct(a, b)
		rhs := FNot(SProbabilistic(FNot(a), FNot(b)))
		if !almostEqual(lhs, rhs, 1e-10) {
			t.Error("De Morgan failed for product/probabilistic")
		}
	})
	t.Run("min_max_duality", func(t *testing.T) {
		a, b := 0.7, 0.4
		lhs := TMin(a, b)
		rhs := FNot(SMax(FNot(a), FNot(b)))
		if !almostEqual(lhs, rhs, 1e-10) {
			t.Error("De Morgan failed for min/max")
		}
	})
}
