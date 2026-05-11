package fuzzy

import "testing"

func TestAlphaCut(t *testing.T) {
	t.Run("strong_bearish", func(t *testing.T) {
		if AlphaCut(-87.3, 0.5, 100.0) != -100 {
			t.Error("expected -100")
		}
	})
	t.Run("weak_bearish", func(t *testing.T) {
		if AlphaCut(-32.1, 0.5, 100.0) != 0 {
			t.Error("expected 0")
		}
	})
	t.Run("strong_bullish", func(t *testing.T) {
		if AlphaCut(92.5, 0.5, 100.0) != 100 {
			t.Error("expected 100")
		}
	})
	t.Run("weak_bullish", func(t *testing.T) {
		if AlphaCut(15.0, 0.5, 100.0) != 0 {
			t.Error("expected 0")
		}
	})
	t.Run("zero", func(t *testing.T) {
		if AlphaCut(0.0, 0.5, 100.0) != 0 {
			t.Error("expected 0")
		}
	})
	t.Run("strong_confirmation", func(t *testing.T) {
		if AlphaCut(156.8, 0.5, 100.0) != 200 {
			t.Error("expected 200")
		}
	})
	t.Run("negative_confirmation", func(t *testing.T) {
		if AlphaCut(-180.0, 0.5, 100.0) != -200 {
			t.Error("expected -200")
		}
	})
	t.Run("high_alpha_filters_more", func(t *testing.T) {
		if AlphaCut(-87.3, 0.9, 100.0) != 0 {
			t.Error("expected 0")
		}
	})
	t.Run("high_alpha_passes_strong", func(t *testing.T) {
		if AlphaCut(-95.0, 0.9, 100.0) != -100 {
			t.Error("expected -100")
		}
	})
	t.Run("low_alpha_passes_more", func(t *testing.T) {
		if AlphaCut(-15.0, 0.1, 100.0) != -100 {
			t.Error("expected -100")
		}
	})
	t.Run("alpha_zero_passes_all", func(t *testing.T) {
		if AlphaCut(-1.0, 0.0, 100.0) != -100 {
			t.Error("expected -100")
		}
	})
	t.Run("exactly_at_threshold", func(t *testing.T) {
		if AlphaCut(50.0, 0.5, 100.0) != 100 {
			t.Error("expected 100")
		}
	})
	t.Run("just_below_threshold", func(t *testing.T) {
		if AlphaCut(49.9, 0.5, 100.0) != 0 {
			t.Error("expected 0")
		}
	})
	t.Run("exactly_100", func(t *testing.T) {
		if AlphaCut(100.0, 0.5, 100.0) != 100 {
			t.Error("expected 100")
		}
	})
	t.Run("exactly_minus_100", func(t *testing.T) {
		if AlphaCut(-100.0, 0.5, 100.0) != -100 {
			t.Error("expected -100")
		}
	})
	t.Run("custom_scale", func(t *testing.T) {
		if AlphaCut(-40.0, 0.5, 50.0) != -50 {
			t.Error("expected -50")
		}
	})
	t.Run("invalid_scale", func(t *testing.T) {
		if AlphaCut(-87.3, 0.5, 0.0) != 0 {
			t.Error("expected 0")
		}
	})
}
