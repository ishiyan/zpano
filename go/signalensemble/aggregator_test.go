package signalensemble

import (
	"math"
	"testing"
)

func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// ── TestAggregatorValidation ───────────────────────────────────────────

func TestAggregatorValidation(t *testing.T) {
	t.Run("n_signals_zero", func(t *testing.T) {
		_, err := NewAggregator(AggregatorParams{NSignals: 0, Method: Equal, FeedbackDelay: 1})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("n_signals_negative", func(t *testing.T) {
		_, err := NewAggregator(AggregatorParams{NSignals: -1, Method: Equal, FeedbackDelay: 1})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("feedback_delay_zero", func(t *testing.T) {
		_, err := NewAggregator(AggregatorParams{NSignals: 2, Method: Equal, FeedbackDelay: 0})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("fixed_requires_weights", func(t *testing.T) {
		_, err := NewAggregator(AggregatorParams{NSignals: 2, Method: Fixed, FeedbackDelay: 1})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("fixed_weights_wrong_length", func(t *testing.T) {
		_, err := NewAggregator(AggregatorParams{NSignals: 2, Method: Fixed, FeedbackDelay: 1, Weights: []float64{1.0}})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("fixed_weights_zero_sum", func(t *testing.T) {
		_, err := NewAggregator(AggregatorParams{NSignals: 2, Method: Fixed, FeedbackDelay: 1, Weights: []float64{0.0, 0.0}})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("inverse_variance_window_too_small", func(t *testing.T) {
		_, err := NewAggregator(AggregatorParams{NSignals: 2, Method: InverseVariance, FeedbackDelay: 1, Window: 1})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("exponential_decay_alpha_zero", func(t *testing.T) {
		_, err := NewAggregator(AggregatorParams{NSignals: 2, Method: ExponentialDecay, FeedbackDelay: 1, Alpha: 0})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("exponential_decay_alpha_negative", func(t *testing.T) {
		_, err := NewAggregator(AggregatorParams{NSignals: 2, Method: ExponentialDecay, FeedbackDelay: 1, Alpha: -0.1})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("multiplicative_weights_eta_zero", func(t *testing.T) {
		_, err := NewAggregator(AggregatorParams{NSignals: 2, Method: MultiplicativeWeights, FeedbackDelay: 1, Eta: 0})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("bayesian_prior_wrong_length", func(t *testing.T) {
		_, err := NewAggregator(AggregatorParams{NSignals: 2, Method: Bayesian, FeedbackDelay: 1, Prior: []float64{1.0}})
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("blend_wrong_signal_count", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: Equal, FeedbackDelay: 1})
		_, err := agg.Blend([]float64{0.5, 0.5})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// ── TestFixedWeights ───────────────────────────────────────────────────

func TestFixedWeights(t *testing.T) {
	t.Run("basic_blend", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: Fixed, FeedbackDelay: 1, Weights: []float64{0.5, 0.3, 0.2}})
		result, _ := agg.Blend([]float64{1.0, 0.0, 0.0})
		if !almostEqual(result, 0.5, 1e-13) {
			t.Fatalf("expected 0.5, got %v", result)
		}
	})
	t.Run("weights_normalized", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: Fixed, FeedbackDelay: 1, Weights: []float64{2.0, 8.0}})
		w := agg.Weights()
		if !almostEqual(w[0], 0.2, 1e-13) || !almostEqual(w[1], 0.8, 1e-13) {
			t.Fatalf("expected [0.2, 0.8], got %v", w)
		}
	})
	t.Run("update_is_noop", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: Fixed, FeedbackDelay: 1, Weights: []float64{0.6, 0.4}})
		agg.Blend([]float64{0.8, 0.2})
		agg.Blend([]float64{0.7, 0.3})
		wBefore := agg.Weights()
		agg.Update(0.9)
		wAfter := agg.Weights()
		for i := range wBefore {
			if !almostEqual(wBefore[i], wAfter[i], 1e-15) {
				t.Fatalf("weights changed after update")
			}
		}
	})
	t.Run("blend_all_ones", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: Fixed, FeedbackDelay: 1, Weights: []float64{0.5, 0.3, 0.2}})
		result, _ := agg.Blend([]float64{1.0, 1.0, 1.0})
		if !almostEqual(result, 1.0, 1e-13) {
			t.Fatalf("expected 1.0, got %v", result)
		}
	})
	t.Run("blend_all_zeros", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: Fixed, FeedbackDelay: 1, Weights: []float64{0.5, 0.3, 0.2}})
		result, _ := agg.Blend([]float64{0.0, 0.0, 0.0})
		if !almostEqual(result, 0.0, 1e-13) {
			t.Fatalf("expected 0.0, got %v", result)
		}
	})
	t.Run("count_increments", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: Fixed, FeedbackDelay: 1, Weights: []float64{0.5, 0.5}})
		if agg.Count() != 0 {
			t.Fatal("expected 0")
		}
		agg.Blend([]float64{0.5, 0.5})
		if agg.Count() != 1 {
			t.Fatal("expected 1")
		}
		agg.Blend([]float64{0.5, 0.5})
		if agg.Count() != 2 {
			t.Fatal("expected 2")
		}
	})
}

// ── TestEqualWeights ───────────────────────────────────────────────────

func TestEqualWeights(t *testing.T) {
	t.Run("basic_blend", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: Equal, FeedbackDelay: 1})
		result, _ := agg.Blend([]float64{0.9, 0.3, 0.6})
		if !almostEqual(result, 0.6, 1e-13) {
			t.Fatalf("expected 0.6, got %v", result)
		}
	})
	t.Run("single_signal", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 1, Method: Equal, FeedbackDelay: 1})
		result, _ := agg.Blend([]float64{0.7})
		if !almostEqual(result, 0.7, 1e-13) {
			t.Fatalf("expected 0.7, got %v", result)
		}
	})
	t.Run("weights_are_uniform", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 4, Method: Equal, FeedbackDelay: 1})
		for _, w := range agg.Weights() {
			if !almostEqual(w, 0.25, 1e-13) {
				t.Fatalf("expected 0.25, got %v", w)
			}
		}
	})
	t.Run("update_is_noop", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: Equal, FeedbackDelay: 1})
		agg.Blend([]float64{0.8, 0.2})
		agg.Blend([]float64{0.7, 0.3})
		wBefore := agg.Weights()
		agg.Update(0.5)
		wAfter := agg.Weights()
		for i := range wBefore {
			if !almostEqual(wBefore[i], wAfter[i], 1e-15) {
				t.Fatal("weights changed")
			}
		}
	})
}

// ── TestInverseVariance ────────────────────────────────────────────────

func TestInverseVariance(t *testing.T) {
	t.Run("initial_weights_uniform", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: InverseVariance, FeedbackDelay: 1, Window: 10})
		for _, w := range agg.Weights() {
			if !almostEqual(w, 1.0/3, 1e-13) {
				t.Fatalf("expected 1/3, got %v", w)
			}
		}
	})
	t.Run("accurate_signal_gets_higher_weight", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: InverseVariance, FeedbackDelay: 1, Window: 10})
		outcomes := []float64{0.5, 0.6, 0.4, 0.55, 0.45, 0.5, 0.6, 0.4, 0.55, 0.45}
		for i, outcome := range outcomes {
			sign := 1.0
			if i%2 == 0 {
				sign = 1.0
			} else {
				sign = -1.0
			}
			s0 := outcome + 0.01*sign
			s1 := 0.9
			if i%2 != 0 {
				s1 = 0.1
			}
			agg.Blend([]float64{s0, s1})
			agg.Update(outcome)
		}
		w := agg.Weights()
		if w[0] <= w[1] {
			t.Fatalf("expected w[0] > w[1], got %v", w)
		}
	})
	t.Run("squared_error_metric", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: InverseVariance, FeedbackDelay: 1, Window: 10, ErrorMetric: Squared})
		for i := 0; i < 5; i++ {
			agg.Blend([]float64{0.5, 0.5})
			agg.Update(0.5)
		}
		w := agg.Weights()
		if !almostEqual(w[0], w[1], 1e-10) {
			t.Fatalf("expected equal weights, got %v", w)
		}
	})
}

// ── TestExponentialDecay ───────────────────────────────────────────────

func TestExponentialDecay(t *testing.T) {
	t.Run("initial_weights_uniform", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: ExponentialDecay, FeedbackDelay: 1, Alpha: 0.1})
		for _, w := range agg.Weights() {
			if !almostEqual(w, 1.0/3, 1e-13) {
				t.Fatalf("expected 1/3, got %v", w)
			}
		}
	})
	t.Run("good_signal_weight_increases", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: ExponentialDecay, FeedbackDelay: 1, Alpha: 0.3})
		for i := 0; i < 20; i++ {
			agg.Blend([]float64{0.8, 0.2})
			agg.Update(0.8)
		}
		w := agg.Weights()
		if w[0] <= w[1] {
			t.Fatalf("expected w[0] > w[1], got %v", w)
		}
	})
	t.Run("alpha_one", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: ExponentialDecay, FeedbackDelay: 1, Alpha: 1.0})
		agg.Blend([]float64{0.9, 0.1})
		agg.Blend([]float64{0.9, 0.1})
		agg.Update(0.9)
		w := agg.Weights()
		if !almostEqual(w[0], 1.0/1.2, 1e-13) {
			t.Fatalf("expected %v, got %v", 1.0/1.2, w[0])
		}
		if !almostEqual(w[1], 0.2/1.2, 1e-13) {
			t.Fatalf("expected %v, got %v", 0.2/1.2, w[1])
		}
	})
}

// ── TestMultiplicativeWeights ──────────────────────────────────────────

func TestMultiplicativeWeights(t *testing.T) {
	t.Run("initial_weights_uniform", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: MultiplicativeWeights, FeedbackDelay: 1, Eta: 0.5})
		for _, w := range agg.Weights() {
			if !almostEqual(w, 1.0/3, 1e-13) {
				t.Fatalf("expected 1/3, got %v", w)
			}
		}
	})
	t.Run("best_signal_converges", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: MultiplicativeWeights, FeedbackDelay: 1, Eta: 0.5})
		for i := 0; i < 50; i++ {
			agg.Blend([]float64{0.8, 0.2, 0.3})
			agg.Update(0.8)
		}
		w := agg.Weights()
		if w[0] <= 0.5 {
			t.Fatalf("expected w[0] > 0.5, got %v", w[0])
		}
		if w[0] <= w[1] || w[0] <= w[2] {
			t.Fatalf("expected w[0] dominant, got %v", w)
		}
	})
	t.Run("high_eta_faster_convergence", func(t *testing.T) {
		var results [2]float64
		for idx, eta := range []float64{0.1, 1.0} {
			agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: MultiplicativeWeights, FeedbackDelay: 1, Eta: eta})
			for i := 0; i < 10; i++ {
				agg.Blend([]float64{0.9, 0.1})
				agg.Update(0.9)
			}
			results[idx] = agg.Weights()[0]
		}
		if results[1] <= results[0] {
			t.Fatalf("expected higher eta → higher w[0], got %v vs %v", results[1], results[0])
		}
	})
}

// ── TestRankBased ──────────────────────────────────────────────────────

func TestRankBased(t *testing.T) {
	t.Run("initial_weights_uniform", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: RankBased, FeedbackDelay: 1, Window: 10})
		for _, w := range agg.Weights() {
			if !almostEqual(w, 1.0/3, 1e-13) {
				t.Fatalf("expected 1/3, got %v", w)
			}
		}
	})
	t.Run("rank_ordering", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: RankBased, FeedbackDelay: 1, Window: 10})
		for i := 0; i < 15; i++ {
			agg.Blend([]float64{0.7, 0.5, 0.2})
			agg.Update(0.7)
		}
		w := agg.Weights()
		if w[0] <= w[1] || w[1] <= w[2] {
			t.Fatalf("expected w[0] > w[1] > w[2], got %v", w)
		}
	})
	t.Run("ties_get_average_rank", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: RankBased, FeedbackDelay: 1, Window: 10})
		for i := 0; i < 5; i++ {
			agg.Blend([]float64{0.5, 0.5})
			agg.Update(0.5)
		}
		w := agg.Weights()
		if !almostEqual(w[0], w[1], 1e-13) {
			t.Fatalf("expected equal weights, got %v", w)
		}
	})
}

// ── TestBayesian ───────────────────────────────────────────────────────

func TestBayesian(t *testing.T) {
	t.Run("uniform_prior", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: Bayesian, FeedbackDelay: 1})
		for _, w := range agg.Weights() {
			if !almostEqual(w, 1.0/3, 1e-13) {
				t.Fatalf("expected 1/3, got %v", w)
			}
		}
	})
	t.Run("custom_prior", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: Bayesian, FeedbackDelay: 1, Prior: []float64{0.5, 0.3, 0.2}})
		w := agg.Weights()
		if !almostEqual(w[0], 0.5, 1e-13) || !almostEqual(w[1], 0.3, 1e-13) || !almostEqual(w[2], 0.2, 1e-13) {
			t.Fatalf("expected [0.5, 0.3, 0.2], got %v", w)
		}
	})
	t.Run("good_predictor_dominates", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: Bayesian, FeedbackDelay: 1})
		for i := 0; i < 20; i++ {
			agg.Blend([]float64{0.9, 0.1})
			agg.Update(0.9)
		}
		w := agg.Weights()
		if w[0] <= 0.9 {
			t.Fatalf("expected w[0] > 0.9, got %v", w[0])
		}
	})
	t.Run("evidence_overrides_prior", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: Bayesian, FeedbackDelay: 1, Prior: []float64{0.1, 0.9}})
		for i := 0; i < 50; i++ {
			agg.Blend([]float64{0.8, 0.2})
			agg.Update(0.8)
		}
		w := agg.Weights()
		if w[0] <= w[1] {
			t.Fatalf("expected w[0] > w[1], got %v", w)
		}
	})
}

// ── TestDelayedFeedback ────────────────────────────────────────────────

func TestDelayedFeedback(t *testing.T) {
	t.Run("delay_1", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: ExponentialDecay, FeedbackDelay: 1, Alpha: 1.0})
		agg.Blend([]float64{0.9, 0.1})
		agg.Blend([]float64{0.5, 0.5})
		agg.Update(0.9)
		w := agg.Weights()
		if !almostEqual(w[0], 1.0/1.2, 1e-13) {
			t.Fatalf("expected %v, got %v", 1.0/1.2, w[0])
		}
		if !almostEqual(w[1], 0.2/1.2, 1e-13) {
			t.Fatalf("expected %v, got %v", 0.2/1.2, w[1])
		}
	})
	t.Run("delay_2", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: ExponentialDecay, FeedbackDelay: 2, Alpha: 1.0})
		agg.Blend([]float64{0.9, 0.1})
		agg.Blend([]float64{0.5, 0.5})
		agg.Update(0.9)
		for _, w := range agg.Weights() {
			if !almostEqual(w, 0.5, 1e-13) {
				t.Fatalf("expected 0.5, got %v", w)
			}
		}
		agg.Blend([]float64{0.3, 0.7})
		agg.Update(0.9)
		w := agg.Weights()
		if !almostEqual(w[0], 1.0/1.2, 1e-13) {
			t.Fatalf("expected %v, got %v", 1.0/1.2, w[0])
		}
		if !almostEqual(w[1], 0.2/1.2, 1e-13) {
			t.Fatalf("expected %v, got %v", 0.2/1.2, w[1])
		}
	})
	t.Run("update_without_enough_history", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: ExponentialDecay, FeedbackDelay: 3, Alpha: 0.5})
		agg.Blend([]float64{0.5, 0.5})
		wBefore := agg.Weights()
		agg.Update(0.5)
		wAfter := agg.Weights()
		for i := range wBefore {
			if !almostEqual(wBefore[i], wAfter[i], 1e-15) {
				t.Fatal("weights changed")
			}
		}
	})
}

// ── TestWarmup ─────────────────────────────────────────────────────────

func TestWarmup(t *testing.T) {
	t.Run("warmup_equals_live_replay", func(t *testing.T) {
		history := []HistoryEntry{
			{[]float64{0.8, 0.3}, 0.7},
			{[]float64{0.6, 0.5}, 0.5},
			{[]float64{0.9, 0.2}, 0.8},
			{[]float64{0.7, 0.4}, 0.6},
			{[]float64{0.5, 0.6}, 0.4},
		}

		agg1, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: ExponentialDecay, FeedbackDelay: 1, Alpha: 0.2})
		agg1.Warmup(history)

		agg2, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: ExponentialDecay, FeedbackDelay: 1, Alpha: 0.2})
		outcomes := []float64{}
		for _, entry := range history {
			agg2.Blend(entry.Signals)
			outcomes = append(outcomes, entry.Outcome)
			idx := len(outcomes) - 1 - 1
			if idx >= 0 {
				agg2.Update(outcomes[idx])
			}
		}

		if agg1.Count() != agg2.Count() {
			t.Fatalf("counts differ: %d vs %d", agg1.Count(), agg2.Count())
		}
		w1, w2 := agg1.Weights(), agg2.Weights()
		for i := range w1 {
			if !almostEqual(w1[i], w2[i], 1e-13) {
				t.Fatalf("weights differ at %d: %v vs %v", i, w1[i], w2[i])
			}
		}
	})
	t.Run("warmup_with_delay_2", func(t *testing.T) {
		history := []HistoryEntry{
			{[]float64{0.8, 0.3, 0.5}, 0.7},
			{[]float64{0.6, 0.5, 0.4}, 0.5},
			{[]float64{0.9, 0.2, 0.6}, 0.8},
			{[]float64{0.7, 0.4, 0.3}, 0.6},
			{[]float64{0.5, 0.6, 0.7}, 0.4},
			{[]float64{0.4, 0.7, 0.2}, 0.3},
		}

		agg1, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: MultiplicativeWeights, FeedbackDelay: 2, Eta: 0.3})
		agg1.Warmup(history)

		agg2, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: MultiplicativeWeights, FeedbackDelay: 2, Eta: 0.3})
		outcomes := []float64{}
		for _, entry := range history {
			agg2.Blend(entry.Signals)
			outcomes = append(outcomes, entry.Outcome)
			idx := len(outcomes) - 1 - 2
			if idx >= 0 {
				agg2.Update(outcomes[idx])
			}
		}

		if agg1.Count() != agg2.Count() {
			t.Fatalf("counts differ")
		}
		w1, w2 := agg1.Weights(), agg2.Weights()
		for i := range w1 {
			if !almostEqual(w1[i], w2[i], 1e-13) {
				t.Fatalf("weights differ at %d", i)
			}
		}
	})
	t.Run("warmup_bayesian", func(t *testing.T) {
		history := make([]HistoryEntry, 20)
		for i := range history {
			history[i] = HistoryEntry{[]float64{0.9, 0.1}, 0.9}
		}

		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: Bayesian, FeedbackDelay: 1})
		agg.Warmup(history)

		w := agg.Weights()
		if w[0] <= 0.9 {
			t.Fatalf("expected w[0] > 0.9, got %v", w[0])
		}
	})
}

// ── TestWeightsProperty ────────────────────────────────────────────────

func TestWeightsProperty(t *testing.T) {
	t.Run("weights_returns_copy", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: Equal, FeedbackDelay: 1})
		w := agg.Weights()
		w[0] = 999.0
		if !almostEqual(agg.Weights()[0], 0.5, 1e-13) {
			t.Fatal("internal weights were modified")
		}
	})
	t.Run("weights_sum_to_one", func(t *testing.T) {
		methods := []struct {
			method  AggregationMethod
			weights []float64
		}{
			{Fixed, []float64{0.3, 0.7}},
			{Equal, nil},
			{InverseVariance, nil},
			{ExponentialDecay, nil},
			{MultiplicativeWeights, nil},
			{RankBased, nil},
			{Bayesian, nil},
		}
		for _, m := range methods {
			p := AggregatorParams{NSignals: 2, FeedbackDelay: 1, Method: m.method, Window: 10, Alpha: 0.1, Eta: 0.5}
			if m.weights != nil {
				p.Weights = m.weights
			}
			agg, err := NewAggregator(p)
			if err != nil {
				t.Fatalf("failed for method %d: %v", m.method, err)
			}
			s := 0.0
			for _, w := range agg.Weights() {
				s += w
			}
			if !almostEqual(s, 1.0, 1e-13) {
				t.Fatalf("weights don't sum to 1 for method %d: %v", m.method, s)
			}
		}
	})
}

// ── TestEdgeCases ──────────────────────────────────────────────────────

func TestEdgeCases(t *testing.T) {
	t.Run("single_signal", func(t *testing.T) {
		methods := []struct {
			method  AggregationMethod
			weights []float64
		}{
			{Fixed, []float64{1.0}},
			{Equal, nil},
			{InverseVariance, nil},
			{ExponentialDecay, nil},
			{MultiplicativeWeights, nil},
			{RankBased, nil},
			{Bayesian, nil},
		}
		for _, m := range methods {
			p := AggregatorParams{NSignals: 1, FeedbackDelay: 1, Method: m.method, Window: 10, Alpha: 0.1, Eta: 0.5}
			if m.weights != nil {
				p.Weights = m.weights
			}
			agg, err := NewAggregator(p)
			if err != nil {
				t.Fatalf("failed for method %d: %v", m.method, err)
			}
			result, _ := agg.Blend([]float64{0.73})
			if !almostEqual(result, 0.73, 1e-13) {
				t.Fatalf("single signal failed for method %d: got %v", m.method, result)
			}
		}
	})
	t.Run("many_signals", func(t *testing.T) {
		n := 100
		agg, _ := NewAggregator(AggregatorParams{NSignals: n, Method: Equal, FeedbackDelay: 1})
		signals := make([]float64, n)
		for i := range signals {
			signals[i] = 0.5
		}
		result, _ := agg.Blend(signals)
		if !almostEqual(result, 0.5, 1e-13) {
			t.Fatalf("expected 0.5, got %v", result)
		}
	})
	t.Run("extreme_signals", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: Equal, FeedbackDelay: 1})
		result, _ := agg.Blend([]float64{0.0, 1.0})
		if !almostEqual(result, 0.5, 1e-13) {
			t.Fatalf("expected 0.5, got %v", result)
		}
	})
	t.Run("bayesian_extreme_signals", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 2, Method: Bayesian, FeedbackDelay: 1})
		agg.Blend([]float64{0.0, 1.0})
		agg.Blend([]float64{0.0, 1.0})
		agg.Update(1.0)
		w := agg.Weights()
		s := 0.0
		for _, ww := range w {
			s += ww
		}
		if !almostEqual(s, 1.0, 1e-13) {
			t.Fatalf("weights don't sum to 1: %v", s)
		}
	})
	t.Run("inverse_variance_identical_signals", func(t *testing.T) {
		agg, _ := NewAggregator(AggregatorParams{NSignals: 3, Method: InverseVariance, FeedbackDelay: 1, Window: 10})
		for i := 0; i < 5; i++ {
			agg.Blend([]float64{0.5, 0.5, 0.5})
			agg.Update(0.5)
		}
		w := agg.Weights()
		for i := 0; i < 3; i++ {
			if !almostEqual(w[i], 1.0/3, 1e-10) {
				t.Fatalf("expected 1/3, got %v", w[i])
			}
		}
	})
}
