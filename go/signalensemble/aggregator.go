package signalensemble

import (
	"fmt"
	"math"
	"sort"
)

// HistoryEntry holds a single signal/outcome pair for warmup replay.
type HistoryEntry struct {
	Signals []float64
	Outcome float64
}

// AggregatorParams holds configuration for constructing an Aggregator.
type AggregatorParams struct {
	NSignals      int               // Number of signal sources (>= 1).
	Method        AggregationMethod // Aggregation method to use. Defaults to Equal.
	FeedbackDelay int               // Number of bars between signal observation and outcome availability (>= 1).
	Weights       []float64   // Required for Fixed method. Normalized to sum to 1.0.
	Window        int         // Rolling window for InverseVariance and RankBased (>= 2).
	Alpha         float64     // Decay rate for ExponentialDecay (0 < alpha <= 1).
	Eta           float64     // Learning rate for MultiplicativeWeights (> 0).
	Prior         []float64   // Prior weights for Bayesian. Defaults to uniform.
	ErrorMetric   ErrorMetric // Error metric for InverseVariance and RankBased.
}

// DefaultParams returns AggregatorParams with sensible defaults.
func DefaultParams() AggregatorParams {
	return AggregatorParams{
		NSignals:      1,
		Method:        Equal,
		FeedbackDelay: 1,
		Window:        50,
		Alpha:         0.1,
		Eta:           0.5,
		ErrorMetric:   Absolute,
	}
}

// rollingWindow is a fixed-capacity circular buffer of float64 values.
type rollingWindow struct {
	data []float64
	cap  int
}

func newRollingWindow(capacity int) rollingWindow {
	return rollingWindow{cap: capacity}
}

func (rw *rollingWindow) append(v float64) {
	if len(rw.data) < rw.cap {
		rw.data = append(rw.data, v)
	} else {
		copy(rw.data, rw.data[1:])
		rw.data[len(rw.data)-1] = v
	}
}

func (rw *rollingWindow) len() int { return len(rw.data) }

// Aggregator blends multiple independent signal sources into a single
// confidence value in [0, 1]. Adaptive methods update weights based on
// observed outcomes with a configurable feedback delay.
type Aggregator struct {
	n             int
	method        AggregationMethod
	feedbackDelay int
	count         int
	weights       []float64
	ring          [][]float64
	ringCapacity  int
	// Method-specific state.
	errors      []rollingWindow // for InverseVariance and RankBased
	errorMetric ErrorMetric
	ema         []float64 // for ExponentialDecay
	alpha       float64
	logWeights  []float64 // for MultiplicativeWeights
	eta         float64
	logPosterior []float64 // for Bayesian
}

// NewAggregator creates an Aggregator with the given parameters.
func NewAggregator(params AggregatorParams) (*Aggregator, error) {
	if params.NSignals < 1 {
		return nil, fmt.Errorf("n_signals must be >= 1, got %d", params.NSignals)
	}
	if params.FeedbackDelay < 1 {
		return nil, fmt.Errorf("feedback_delay must be >= 1, got %d", params.FeedbackDelay)
	}

	n := params.NSignals
	a := &Aggregator{
		n:             n,
		method:        params.Method,
		feedbackDelay: params.FeedbackDelay,
		ringCapacity:  params.FeedbackDelay + 1,
	}

	// Initialize method-specific state and weights.
	switch params.Method {
	case Fixed:
		if params.Weights == nil {
			return nil, fmt.Errorf("FIXED method requires weights")
		}
		if len(params.Weights) != n {
			return nil, fmt.Errorf("weights length %d != n_signals %d", len(params.Weights), n)
		}
		s := 0.0
		for _, w := range params.Weights {
			s += w
		}
		if s <= 0 {
			return nil, fmt.Errorf("weights must sum to a positive value")
		}
		a.weights = make([]float64, n)
		for i, w := range params.Weights {
			a.weights[i] = w / s
		}

	case Equal:
		a.weights = uniformWeights(n)

	case InverseVariance:
		if params.Window < 2 {
			return nil, fmt.Errorf("window must be >= 2, got %d", params.Window)
		}
		a.errorMetric = params.ErrorMetric
		a.errors = make([]rollingWindow, n)
		for i := range a.errors {
			a.errors[i] = newRollingWindow(params.Window)
		}
		a.weights = uniformWeights(n)

	case ExponentialDecay:
		if params.Alpha <= 0 || params.Alpha > 1 {
			return nil, fmt.Errorf("alpha must be in (0, 1], got %v", params.Alpha)
		}
		a.alpha = params.Alpha
		a.ema = make([]float64, n)
		for i := range a.ema {
			a.ema[i] = 0.5 // neutral prior
		}
		a.weights = uniformWeights(n)

	case MultiplicativeWeights:
		if params.Eta <= 0 {
			return nil, fmt.Errorf("eta must be > 0, got %v", params.Eta)
		}
		a.eta = params.Eta
		a.logWeights = make([]float64, n) // uniform in log-space
		a.weights = uniformWeights(n)

	case RankBased:
		if params.Window < 2 {
			return nil, fmt.Errorf("window must be >= 2, got %d", params.Window)
		}
		a.errorMetric = params.ErrorMetric
		a.errors = make([]rollingWindow, n)
		for i := range a.errors {
			a.errors[i] = newRollingWindow(params.Window)
		}
		a.weights = uniformWeights(n)

	case Bayesian:
		normalizedPrior := make([]float64, n)
		if params.Prior != nil {
			if len(params.Prior) != n {
				return nil, fmt.Errorf("prior length %d != n_signals %d", len(params.Prior), n)
			}
			s := 0.0
			for _, p := range params.Prior {
				s += p
			}
			if s <= 0 {
				return nil, fmt.Errorf("prior must sum to a positive value")
			}
			for i, p := range params.Prior {
				normalizedPrior[i] = p / s
			}
		} else {
			for i := range normalizedPrior {
				normalizedPrior[i] = 1.0 / float64(n)
			}
		}
		a.logPosterior = make([]float64, n)
		for i, p := range normalizedPrior {
			a.logPosterior[i] = math.Log(p)
		}
		a.weights = make([]float64, n)
		copy(a.weights, normalizedPrior)

	default:
		return nil, fmt.Errorf("unknown method: %d", params.Method)
	}

	return a, nil
}

// Blend blends signal sources into a single confidence value.
// signals must contain one value per source in [0, 1].
// Returns the blended confidence in [0, 1].
func (a *Aggregator) Blend(signals []float64) (float64, error) {
	if len(signals) != a.n {
		return 0, fmt.Errorf("expected %d signals, got %d", a.n, len(signals))
	}

	output := 0.0
	for i := 0; i < a.n; i++ {
		output += a.weights[i] * signals[i]
	}

	// Append to ring buffer (deque with maxlen).
	cp := make([]float64, a.n)
	copy(cp, signals)
	if len(a.ring) < a.ringCapacity {
		a.ring = append(a.ring, cp)
	} else {
		copy(a.ring, a.ring[1:])
		a.ring[len(a.ring)-1] = cp
	}
	a.count++
	return output, nil
}

// Update provides outcome feedback for weight adaptation.
// For stateless methods (Fixed, Equal), this is a no-op.
// For adaptive methods, pairs the outcome with the buffered signals
// from feedbackDelay bars ago and updates weights.
// outcome is the observed outcome in [0, 1].
func (a *Aggregator) Update(outcome float64) {
	if a.method == Fixed || a.method == Equal {
		return
	}
	if len(a.ring) < a.feedbackDelay+1 {
		return
	}

	// Retrieve signals from feedback_delay bars ago.
	idx := len(a.ring) - 1 - a.feedbackDelay
	pastSignals := a.ring[idx]

	switch a.method {
	case InverseVariance:
		a.updateInverseVariance(pastSignals, outcome)
	case ExponentialDecay:
		a.updateExponentialDecay(pastSignals, outcome)
	case MultiplicativeWeights:
		a.updateMultiplicativeWeights(pastSignals, outcome)
	case RankBased:
		a.updateRankBased(pastSignals, outcome)
	case Bayesian:
		a.updateBayesian(pastSignals, outcome)
	}
}

// Warmup replays historical data through Blend() + Update().
// Each entry contains signals at bar T and the outcome for bar T.
// The method handles the feedback delay internally.
func (a *Aggregator) Warmup(history []HistoryEntry) error {
	outcomes := make([]float64, 0, len(history))
	for _, entry := range history {
		_, err := a.Blend(entry.Signals)
		if err != nil {
			return err
		}
		outcomes = append(outcomes, entry.Outcome)
		idx := len(outcomes) - 1 - a.feedbackDelay
		if idx >= 0 {
			a.Update(outcomes[idx])
		}
	}
	return nil
}

// Weights returns a copy of the current normalized weights.
func (a *Aggregator) Weights() []float64 {
	w := make([]float64, a.n)
	copy(w, a.weights)
	return w
}

// Count returns the total number of Blend() calls.
func (a *Aggregator) Count() int {
	return a.count
}

// ── Private helpers ────────────────────────────────────────────────────

func uniformWeights(n int) []float64 {
	w := make([]float64, n)
	v := 1.0 / float64(n)
	for i := range w {
		w[i] = v
	}
	return w
}

// computeError computes per-signal error using the configured metric.
func (a *Aggregator) computeError(signal, outcome float64) float64 {
	diff := signal - outcome
	if a.errorMetric == Absolute {
		return math.Abs(diff)
	}
	return diff * diff
}

// updateInverseVariance updates weights using inverse-variance of prediction errors.
func (a *Aggregator) updateInverseVariance(signals []float64, outcome float64) {
	const epsilon = 1e-15

	for i := 0; i < a.n; i++ {
		e := a.computeError(signals[i], outcome)
		a.errors[i].append(e)
	}

	// Need at least 2 errors per signal to compute variance.
	for i := 0; i < a.n; i++ {
		if a.errors[i].len() < 2 {
			return
		}
	}

	raw := make([]float64, a.n)
	for i := 0; i < a.n; i++ {
		data := a.errors[i].data
		n := float64(len(data))
		mean := 0.0
		for _, e := range data {
			mean += e
		}
		mean /= n
		variance := 0.0
		for _, e := range data {
			d := e - mean
			variance += d * d
		}
		variance /= n // population variance
		raw[i] = 1.0 / math.Max(variance, epsilon)
	}

	total := 0.0
	for _, r := range raw {
		total += r
	}
	for i := range a.weights {
		a.weights[i] = raw[i] / total
	}
}

// updateExponentialDecay updates weights using EMA of accuracy.
func (a *Aggregator) updateExponentialDecay(signals []float64, outcome float64) {
	for i := 0; i < a.n; i++ {
		error_ := math.Abs(signals[i] - outcome)
		accuracy := 1.0 - error_
		a.ema[i] = a.alpha*accuracy + (1.0-a.alpha)*a.ema[i]
	}

	// Normalize, clamping negative EMAs to 0.
	clamped := make([]float64, a.n)
	total := 0.0
	for i, e := range a.ema {
		c := math.Max(e, 0.0)
		clamped[i] = c
		total += c
	}
	if total > 0 {
		for i := range a.weights {
			a.weights[i] = clamped[i] / total
		}
	} else {
		a.weights = uniformWeights(a.n)
	}
}

// updateMultiplicativeWeights updates weights using the Hedge algorithm in log-space.
func (a *Aggregator) updateMultiplicativeWeights(signals []float64, outcome float64) {
	for i := 0; i < a.n; i++ {
		loss := math.Abs(signals[i] - outcome)
		a.logWeights[i] -= a.eta * loss
	}

	// Softmax normalization (log-sum-exp trick).
	maxLog := a.logWeights[0]
	for _, lw := range a.logWeights[1:] {
		if lw > maxLog {
			maxLog = lw
		}
	}
	total := 0.0
	expWeights := make([]float64, a.n)
	for i, lw := range a.logWeights {
		expWeights[i] = math.Exp(lw - maxLog)
		total += expWeights[i]
	}
	for i := range a.weights {
		a.weights[i] = expWeights[i] / total
	}
}

// updateRankBased updates weights using rank of rolling accuracy.
func (a *Aggregator) updateRankBased(signals []float64, outcome float64) {
	for i := 0; i < a.n; i++ {
		e := a.computeError(signals[i], outcome)
		a.errors[i].append(e)
	}

	// Need at least 1 error per signal.
	for i := 0; i < a.n; i++ {
		if a.errors[i].len() < 1 {
			return
		}
	}

	// Compute mean accuracy per signal.
	accuracies := make([]float64, a.n)
	for i := 0; i < a.n; i++ {
		data := a.errors[i].data
		mean := 0.0
		for _, e := range data {
			mean += e
		}
		mean /= float64(len(data))
		accuracies[i] = 1.0 - mean
	}

	// Rank by accuracy (best = highest rank = n, worst = 1).
	// Ties get the average rank.
	ranks := rankWithTies(accuracies)

	total := 0.0
	for _, r := range ranks {
		total += r
	}
	if total > 0 {
		for i := range a.weights {
			a.weights[i] = ranks[i] / total
		}
	} else {
		a.weights = uniformWeights(a.n)
	}
}

// rankWithTies ranks values from 1 (worst) to n (best), averaging ties.
func rankWithTies(values []float64) []float64 {
	n := len(values)
	// Sort indices by value.
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(a, b int) bool {
		return values[indices[a]] < values[indices[b]]
	})

	ranks := make([]float64, n)
	i := 0
	for i < n {
		// Find the end of the tie group.
		j := i + 1
		for j < n && values[indices[j]] == values[indices[i]] {
			j++
		}
		// Average rank for this group (1-based).
		avgRank := float64(i+1+j) / 2.0
		for k := i; k < j; k++ {
			ranks[indices[k]] = avgRank
		}
		i = j
	}
	return ranks
}

// updateBayesian updates weights using Bayesian model averaging (Bernoulli likelihood).
func (a *Aggregator) updateBayesian(signals []float64, outcome float64) {
	const epsilon = 1e-15

	for i := 0; i < a.n; i++ {
		// Clamp signal to [epsilon, 1 - epsilon] to avoid log(0).
		s := math.Max(epsilon, math.Min(1.0-epsilon, signals[i]))
		logLik := outcome*math.Log(s) + (1.0-outcome)*math.Log(1.0-s)
		a.logPosterior[i] += logLik
	}

	// Softmax normalization.
	maxLog := a.logPosterior[0]
	for _, lp := range a.logPosterior[1:] {
		if lp > maxLog {
			maxLog = lp
		}
	}
	total := 0.0
	expWeights := make([]float64, a.n)
	for i, lp := range a.logPosterior {
		expWeights[i] = math.Exp(lp - maxLog)
		total += expWeights[i]
	}
	for i := range a.weights {
		a.weights[i] = expWeights[i] / total
	}
}
