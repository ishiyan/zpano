/**
 * Aggregator for weighted blending of multiple signal sources.
 *
 * Combines n independent signal sources (each producing values in [0, 1])
 * into a single blended confidence using one of seven aggregation methods.
 * Supports delayed feedback and online weight learning.
 */

import { AggregationMethod } from './aggregation-method';
import { ErrorMetric } from './error-metric';

/** Parameters for constructing an Aggregator. */
export interface AggregatorParams {
    /** Number of signal sources (>= 1). */
    nSignals: number;
    /** Aggregation method to use. Defaults to EQUAL. */
    method?: AggregationMethod;
    /** Number of bars between signal observation and outcome availability (>= 1). */
    feedbackDelay?: number;
    /** Required for FIXED method. Normalized to sum to 1.0. */
    weights?: number[];
    /** Rolling window size for INVERSE_VARIANCE and RANK_BASED (>= 2). */
    window?: number;
    /** Decay rate for EXPONENTIAL_DECAY (0 < alpha <= 1). */
    alpha?: number;
    /** Learning rate for MULTIPLICATIVE_WEIGHTS (> 0). */
    eta?: number;
    /** Prior weights for BAYESIAN. Defaults to uniform. */
    prior?: number[];
    /** Error metric for INVERSE_VARIANCE and RANK_BASED. */
    errorMetric?: ErrorMetric;
}

/** History entry for warmup: signals at bar T and outcome for bar T. */
export interface HistoryEntry {
    signals: number[];
    outcome: number;
}

/**
 * Weighted signal ensemble aggregator.
 *
 * Blends multiple independent signal sources into a single confidence
 * value in [0, 1]. Adaptive methods update weights based on observed
 * outcomes with a configurable feedback delay.
 */
export class Aggregator {
    private _n: number;
    private _method: AggregationMethod;
    private _feedbackDelay: number;
    private _count: number = 0;
    private _weights: number[];
    private _ring: number[][] = [];
    private _ringCapacity: number;

    // Method-specific state.
    private _errors?: number[][];       // for INVERSE_VARIANCE and RANK_BASED
    private _errorMetric?: ErrorMetric;
    private _ema?: number[];            // for EXPONENTIAL_DECAY
    private _alpha?: number;
    private _logWeights?: number[];     // for MULTIPLICATIVE_WEIGHTS
    private _eta?: number;
    private _logPosterior?: number[];   // for BAYESIAN
    private _window?: number;

    constructor(params: AggregatorParams) {
        const {
            nSignals,
            method = AggregationMethod.EQUAL,
            feedbackDelay = 1,
            weights,
            window = 50,
            alpha = 0.1,
            eta = 0.5,
            prior,
            errorMetric = ErrorMetric.ABSOLUTE,
        } = params;

        if (nSignals < 1) {
            throw new Error(`n_signals must be >= 1, got ${nSignals}`);
        }
        if (feedbackDelay < 1) {
            throw new Error(`feedback_delay must be >= 1, got ${feedbackDelay}`);
        }

        this._n = nSignals;
        this._method = method;
        this._feedbackDelay = feedbackDelay;
        this._ringCapacity = feedbackDelay + 1;

        // Initialize method-specific state and weights.
        if (method === AggregationMethod.FIXED) {
            if (weights === undefined) {
                throw new Error('FIXED method requires weights');
            }
            if (weights.length !== nSignals) {
                throw new Error(`weights length ${weights.length} != n_signals ${nSignals}`);
            }
            const s = weights.reduce((a, b) => a + b, 0);
            if (s <= 0) {
                throw new Error('weights must sum to a positive value');
            }
            this._weights = weights.map(w => w / s);

        } else if (method === AggregationMethod.EQUAL) {
            this._weights = Array(nSignals).fill(1.0 / nSignals);

        } else if (method === AggregationMethod.INVERSE_VARIANCE) {
            if (window < 2) {
                throw new Error(`window must be >= 2, got ${window}`);
            }
            this._window = window;
            this._errorMetric = errorMetric;
            this._errors = [];
            for (let i = 0; i < nSignals; i++) {
                this._errors.push([]);
            }
            this._weights = Array(nSignals).fill(1.0 / nSignals);

        } else if (method === AggregationMethod.EXPONENTIAL_DECAY) {
            if (alpha <= 0 || alpha > 1) {
                throw new Error(`alpha must be in (0, 1], got ${alpha}`);
            }
            this._alpha = alpha;
            this._ema = Array(nSignals).fill(0.5); // neutral prior
            this._weights = Array(nSignals).fill(1.0 / nSignals);

        } else if (method === AggregationMethod.MULTIPLICATIVE_WEIGHTS) {
            if (eta <= 0) {
                throw new Error(`eta must be > 0, got ${eta}`);
            }
            this._eta = eta;
            this._logWeights = Array(nSignals).fill(0.0); // uniform in log-space
            this._weights = Array(nSignals).fill(1.0 / nSignals);

        } else if (method === AggregationMethod.RANK_BASED) {
            if (window < 2) {
                throw new Error(`window must be >= 2, got ${window}`);
            }
            this._window = window;
            this._errorMetric = errorMetric;
            this._errors = [];
            for (let i = 0; i < nSignals; i++) {
                this._errors.push([]);
            }
            this._weights = Array(nSignals).fill(1.0 / nSignals);

        } else if (method === AggregationMethod.BAYESIAN) {
            let normalizedPrior: number[];
            if (prior !== undefined) {
                if (prior.length !== nSignals) {
                    throw new Error(`prior length ${prior.length} != n_signals ${nSignals}`);
                }
                const s = prior.reduce((a, b) => a + b, 0);
                if (s <= 0) {
                    throw new Error('prior must sum to a positive value');
                }
                normalizedPrior = prior.map(p => p / s);
            } else {
                normalizedPrior = Array(nSignals).fill(1.0 / nSignals);
            }
            this._logPosterior = normalizedPrior.map(p => Math.log(p));
            this._weights = [...normalizedPrior];

        } else {
            throw new Error(`unknown method: ${method}`);
        }
    }

    /**
     * Blend signal sources into a single confidence value.
     *
     * @param signals Signal values in [0, 1], one per source.
     * @returns Blended confidence in [0, 1].
     */
    blend(signals: number[]): number {
        if (signals.length !== this._n) {
            throw new Error(`expected ${this._n} signals, got ${signals.length}`);
        }

        let output = 0;
        for (let i = 0; i < this._n; i++) {
            output += this._weights[i] * signals[i];
        }

        // Append to ring buffer (deque with maxlen).
        this._ring.push([...signals]);
        if (this._ring.length > this._ringCapacity) {
            this._ring.shift();
        }
        this._count++;
        return output;
    }

    /**
     * Provide outcome feedback for weight adaptation.
     *
     * For stateless methods (FIXED, EQUAL), this is a no-op.
     * For adaptive methods, pairs the outcome with the buffered signals
     * from feedbackDelay bars ago and updates weights.
     *
     * @param outcome Observed outcome in [0, 1].
     */
    update(outcome: number): void {
        if (this._method === AggregationMethod.FIXED || this._method === AggregationMethod.EQUAL) {
            return;
        }

        if (this._ring.length < this._feedbackDelay + 1) {
            return;
        }

        // Retrieve signals from feedbackDelay bars ago.
        const idx = this._ring.length - 1 - this._feedbackDelay;
        const pastSignals = this._ring[idx];

        if (this._method === AggregationMethod.INVERSE_VARIANCE) {
            this._updateInverseVariance(pastSignals, outcome);
        } else if (this._method === AggregationMethod.EXPONENTIAL_DECAY) {
            this._updateExponentialDecay(pastSignals, outcome);
        } else if (this._method === AggregationMethod.MULTIPLICATIVE_WEIGHTS) {
            this._updateMultiplicativeWeights(pastSignals, outcome);
        } else if (this._method === AggregationMethod.RANK_BASED) {
            this._updateRankBased(pastSignals, outcome);
        } else if (this._method === AggregationMethod.BAYESIAN) {
            this._updateBayesian(pastSignals, outcome);
        }
    }

    /**
     * Replay historical data through blend() + update().
     *
     * Each entry contains (signals at bar T, outcome for bar T).
     * The method handles the feedback delay internally.
     *
     * @param history Historical signal/outcome pairs.
     */
    warmup(history: HistoryEntry[]): void {
        const outcomes: number[] = [];
        for (const entry of history) {
            this.blend(entry.signals);
            outcomes.push(entry.outcome);
            const idx = outcomes.length - 1 - this._feedbackDelay;
            if (idx >= 0) {
                this.update(outcomes[idx]);
            }
        }
    }

    /** Current normalized weights (read-only copy). */
    get weights(): number[] {
        return [...this._weights];
    }

    /** Total number of blend() calls. */
    get count(): number {
        return this._count;
    }

    // ── Private update methods ──────────────────────────────────────────

    /** Compute per-signal error using the configured metric. */
    private _computeError(signal: number, outcome: number): number {
        const diff = signal - outcome;
        if (this._errorMetric === ErrorMetric.ABSOLUTE) {
            return Math.abs(diff);
        } else { // SQUARED
            return diff * diff;
        }
    }

    /** Update weights using inverse-variance of prediction errors. */
    private _updateInverseVariance(signals: number[], outcome: number): void {
        const epsilon = 1e-15;

        for (let i = 0; i < this._n; i++) {
            const error = this._computeError(signals[i], outcome);
            this._errors![i].push(error);
            if (this._errors![i].length > this._window!) {
                this._errors![i].shift();
            }
        }

        // Need at least 2 errors to compute variance.
        for (let i = 0; i < this._n; i++) {
            if (this._errors![i].length < 2) return;
        }

        const raw: number[] = [];
        for (let i = 0; i < this._n; i++) {
            const errors = this._errors![i];
            const n = errors.length;
            let mean = 0;
            for (const e of errors) mean += e;
            mean /= n;
            let variance = 0;
            for (const e of errors) variance += (e - mean) ** 2;
            variance /= n; // population variance
            raw.push(1.0 / Math.max(variance, epsilon));
        }

        const total = raw.reduce((a, b) => a + b, 0);
        this._weights = raw.map(r => r / total);
    }

    /** Update weights using EMA of accuracy. */
    private _updateExponentialDecay(signals: number[], outcome: number): void {
        for (let i = 0; i < this._n; i++) {
            const error = Math.abs(signals[i] - outcome);
            const accuracy = 1.0 - error;
            this._ema![i] = this._alpha! * accuracy + (1.0 - this._alpha!) * this._ema![i];
        }

        // Normalize, clamping negative EMAs to 0.
        const clamped = this._ema!.map(e => Math.max(e, 0.0));
        const total = clamped.reduce((a, b) => a + b, 0);
        if (total > 0) {
            this._weights = clamped.map(c => c / total);
        } else {
            this._weights = Array(this._n).fill(1.0 / this._n);
        }
    }

    /** Update weights using the Hedge algorithm in log-space. */
    private _updateMultiplicativeWeights(signals: number[], outcome: number): void {
        for (let i = 0; i < this._n; i++) {
            const loss = Math.abs(signals[i] - outcome);
            this._logWeights![i] -= this._eta! * loss;
        }

        // Softmax normalization (log-sum-exp trick).
        const maxLog = Math.max(...this._logWeights!);
        const expWeights = this._logWeights!.map(lw => Math.exp(lw - maxLog));
        const total = expWeights.reduce((a, b) => a + b, 0);
        this._weights = expWeights.map(e => e / total);
    }

    /** Update weights using rank of rolling accuracy. */
    private _updateRankBased(signals: number[], outcome: number): void {
        for (let i = 0; i < this._n; i++) {
            const error = this._computeError(signals[i], outcome);
            this._errors![i].push(error);
            if (this._errors![i].length > this._window!) {
                this._errors![i].shift();
            }
        }

        // Need at least 1 error per signal.
        for (let i = 0; i < this._n; i++) {
            if (this._errors![i].length < 1) return;
        }

        // Compute mean accuracy per signal.
        const accuracies: number[] = [];
        for (let i = 0; i < this._n; i++) {
            const errors = this._errors![i];
            const meanError = errors.reduce((a, b) => a + b, 0) / errors.length;
            accuracies.push(1.0 - meanError);
        }

        // Rank by accuracy (best = highest rank = n, worst = 1).
        // Ties get the average rank.
        const ranks = Aggregator._rankWithTies(accuracies);
        const total = ranks.reduce((a, b) => a + b, 0);
        if (total > 0) {
            this._weights = ranks.map(r => r / total);
        } else {
            this._weights = Array(this._n).fill(1.0 / this._n);
        }
    }

    /** Rank values from 1 (worst) to n (best), averaging ties. */
    private static _rankWithTies(values: number[]): number[] {
        const n = values.length;
        // Sort indices by value.
        const sortedIndices = Array.from({ length: n }, (_, i) => i);
        sortedIndices.sort((a, b) => values[a] - values[b]);
        const ranks = new Array<number>(n);

        let i = 0;
        while (i < n) {
            // Find the end of the tie group.
            let j = i + 1;
            while (j < n && values[sortedIndices[j]] === values[sortedIndices[i]]) {
                j++;
            }
            // Average rank for this group (1-based).
            const avgRank = (i + 1 + j) / 2.0;
            for (let k = i; k < j; k++) {
                ranks[sortedIndices[k]] = avgRank;
            }
            i = j;
        }

        return ranks;
    }

    /** Update weights using Bayesian model averaging (Bernoulli likelihood). */
    private _updateBayesian(signals: number[], outcome: number): void {
        const epsilon = 1e-15;

        for (let i = 0; i < this._n; i++) {
            // Clamp signal to [epsilon, 1 - epsilon] to avoid log(0).
            const s = Math.max(epsilon, Math.min(1.0 - epsilon, signals[i]));
            const logLik = outcome * Math.log(s) + (1.0 - outcome) * Math.log(1.0 - s);
            this._logPosterior![i] += logLik;
        }

        // Softmax normalization.
        const maxLog = Math.max(...this._logPosterior!);
        const expWeights = this._logPosterior!.map(lp => Math.exp(lp - maxLog));
        const total = expWeights.reduce((a, b) => a + b, 0);
        this._weights = expWeights.map(e => e / total);
    }
}
