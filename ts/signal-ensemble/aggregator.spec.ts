import { AggregationMethod } from './aggregation-method';
import { ErrorMetric } from './error-metric';
import { Aggregator, HistoryEntry } from './aggregator';

describe('Aggregator', () => {

    describe('Validation', () => {
        it('should reject n_signals zero', () => {
            expect(() => new Aggregator({ nSignals: 0, method: AggregationMethod.EQUAL })).toThrowError();
        });

        it('should reject n_signals negative', () => {
            expect(() => new Aggregator({ nSignals: -1, method: AggregationMethod.EQUAL })).toThrowError();
        });

        it('should reject feedback_delay zero', () => {
            expect(() => new Aggregator({ nSignals: 2, feedbackDelay: 0 })).toThrowError();
        });

        it('should require weights for FIXED', () => {
            expect(() => new Aggregator({ nSignals: 2, method: AggregationMethod.FIXED })).toThrowError();
        });

        it('should reject wrong length weights for FIXED', () => {
            expect(() => new Aggregator({ nSignals: 2, method: AggregationMethod.FIXED, weights: [1.0] })).toThrowError();
        });

        it('should reject zero-sum weights for FIXED', () => {
            expect(() => new Aggregator({ nSignals: 2, method: AggregationMethod.FIXED, weights: [0.0, 0.0] })).toThrowError();
        });

        it('should reject window < 2 for INVERSE_VARIANCE', () => {
            expect(() => new Aggregator({ nSignals: 2, method: AggregationMethod.INVERSE_VARIANCE, window: 1 })).toThrowError();
        });

        it('should reject alpha zero for EXPONENTIAL_DECAY', () => {
            expect(() => new Aggregator({ nSignals: 2, method: AggregationMethod.EXPONENTIAL_DECAY, alpha: 0 })).toThrowError();
        });

        it('should reject alpha negative for EXPONENTIAL_DECAY', () => {
            expect(() => new Aggregator({ nSignals: 2, method: AggregationMethod.EXPONENTIAL_DECAY, alpha: -0.1 })).toThrowError();
        });

        it('should reject eta zero for MULTIPLICATIVE_WEIGHTS', () => {
            expect(() => new Aggregator({ nSignals: 2, method: AggregationMethod.MULTIPLICATIVE_WEIGHTS, eta: 0 })).toThrowError();
        });

        it('should reject wrong length prior for BAYESIAN', () => {
            expect(() => new Aggregator({ nSignals: 2, method: AggregationMethod.BAYESIAN, prior: [1.0] })).toThrowError();
        });

        it('should reject wrong signal count in blend', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.EQUAL });
            expect(() => agg.blend([0.5, 0.5])).toThrowError();
        });
    });

    describe('Fixed weights', () => {
        it('should blend with basic weights', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.FIXED, weights: [0.5, 0.3, 0.2] });
            const result = agg.blend([1.0, 0.0, 0.0]);
            expect(result).toBeCloseTo(0.5, 13);
        });

        it('should normalize weights', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.FIXED, weights: [2.0, 8.0] });
            const w = agg.weights;
            expect(w[0]).toBeCloseTo(0.2, 13);
            expect(w[1]).toBeCloseTo(0.8, 13);
        });

        it('should not change weights on update', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.FIXED, weights: [0.6, 0.4] });
            agg.blend([0.8, 0.2]);
            agg.blend([0.7, 0.3]);
            const wBefore = agg.weights;
            agg.update(0.9);
            expect(agg.weights).toEqual(wBefore);
        });

        it('should blend all ones to 1.0', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.FIXED, weights: [0.5, 0.3, 0.2] });
            const result = agg.blend([1.0, 1.0, 1.0]);
            expect(result).toBeCloseTo(1.0, 13);
        });

        it('should blend all zeros to 0.0', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.FIXED, weights: [0.5, 0.3, 0.2] });
            const result = agg.blend([0.0, 0.0, 0.0]);
            expect(result).toBeCloseTo(0.0, 13);
        });

        it('should increment count', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.FIXED, weights: [0.5, 0.5] });
            expect(agg.count).toBe(0);
            agg.blend([0.5, 0.5]);
            expect(agg.count).toBe(1);
            agg.blend([0.5, 0.5]);
            expect(agg.count).toBe(2);
        });
    });

    describe('Equal weights', () => {
        it('should blend to average', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.EQUAL });
            const result = agg.blend([0.9, 0.3, 0.6]);
            expect(result).toBeCloseTo(0.6, 13);
        });

        it('should pass through single signal', () => {
            const agg = new Aggregator({ nSignals: 1, method: AggregationMethod.EQUAL });
            const result = agg.blend([0.7]);
            expect(result).toBeCloseTo(0.7, 13);
        });

        it('should have uniform weights', () => {
            const agg = new Aggregator({ nSignals: 4, method: AggregationMethod.EQUAL });
            for (const w of agg.weights) {
                expect(w).toBeCloseTo(0.25, 13);
            }
        });

        it('should not change weights on update', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.EQUAL });
            agg.blend([0.8, 0.2]);
            agg.blend([0.7, 0.3]);
            const wBefore = agg.weights;
            agg.update(0.5);
            expect(agg.weights).toEqual(wBefore);
        });
    });

    describe('Inverse variance', () => {
        it('should start with uniform weights', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.INVERSE_VARIANCE, feedbackDelay: 1, window: 10 });
            for (const w of agg.weights) {
                expect(w).toBeCloseTo(1.0 / 3, 13);
            }
        });

        it('should give higher weight to accurate signal', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.INVERSE_VARIANCE, feedbackDelay: 1, window: 10 });
            const outcomes = [0.5, 0.6, 0.4, 0.55, 0.45, 0.5, 0.6, 0.4, 0.55, 0.45];
            for (let i = 0; i < outcomes.length; i++) {
                const outcome = outcomes[i];
                const s0 = outcome + 0.01 * ((-1) ** i);
                const s1 = i % 2 === 0 ? 0.9 : 0.1;
                agg.blend([s0, s1]);
                agg.update(outcome);
            }
            const w = agg.weights;
            expect(w[0]).toBeGreaterThan(w[1]);
        });

        it('should work with squared error metric', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.INVERSE_VARIANCE, feedbackDelay: 1, window: 10, errorMetric: ErrorMetric.SQUARED });
            for (let i = 0; i < 5; i++) {
                agg.blend([0.5, 0.5]);
                agg.update(0.5);
            }
            const w = agg.weights;
            expect(w[0]).toBeCloseTo(w[1], 10);
        });
    });

    describe('Exponential decay', () => {
        it('should start with uniform weights', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.EXPONENTIAL_DECAY, feedbackDelay: 1 });
            for (const w of agg.weights) {
                expect(w).toBeCloseTo(1.0 / 3, 13);
            }
        });

        it('should increase weight for good signal', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.EXPONENTIAL_DECAY, feedbackDelay: 1, alpha: 0.3 });
            for (let i = 0; i < 20; i++) {
                agg.blend([0.8, 0.2]);
                agg.update(0.8);
            }
            const w = agg.weights;
            expect(w[0]).toBeGreaterThan(w[1]);
        });

        it('should use only latest observation with alpha=1', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.EXPONENTIAL_DECAY, feedbackDelay: 1, alpha: 1.0 });
            agg.blend([0.9, 0.1]);
            agg.blend([0.9, 0.1]);
            agg.update(0.9);
            const w = agg.weights;
            expect(w[0]).toBeCloseTo(1.0 / 1.2, 13);
            expect(w[1]).toBeCloseTo(0.2 / 1.2, 13);
        });
    });

    describe('Multiplicative weights', () => {
        it('should start with uniform weights', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.MULTIPLICATIVE_WEIGHTS, feedbackDelay: 1 });
            for (const w of agg.weights) {
                expect(w).toBeCloseTo(1.0 / 3, 13);
            }
        });

        it('should converge to best signal', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.MULTIPLICATIVE_WEIGHTS, feedbackDelay: 1, eta: 0.5 });
            for (let i = 0; i < 50; i++) {
                agg.blend([0.8, 0.2, 0.3]);
                agg.update(0.8);
            }
            const w = agg.weights;
            expect(w[0]).toBeGreaterThan(0.5);
            expect(w[0]).toBeGreaterThan(w[1]);
            expect(w[0]).toBeGreaterThan(w[2]);
        });

        it('should converge faster with higher eta', () => {
            const results: number[] = [];
            for (const eta of [0.1, 1.0]) {
                const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.MULTIPLICATIVE_WEIGHTS, feedbackDelay: 1, eta });
                for (let i = 0; i < 10; i++) {
                    agg.blend([0.9, 0.1]);
                    agg.update(0.9);
                }
                results.push(agg.weights[0]);
            }
            expect(results[1]).toBeGreaterThan(results[0]);
        });
    });

    describe('Rank based', () => {
        it('should start with uniform weights', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.RANK_BASED, feedbackDelay: 1, window: 10 });
            for (const w of agg.weights) {
                expect(w).toBeCloseTo(1.0 / 3, 13);
            }
        });

        it('should order weights by accuracy', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.RANK_BASED, feedbackDelay: 1, window: 10 });
            for (let i = 0; i < 15; i++) {
                agg.blend([0.7, 0.5, 0.2]);
                agg.update(0.7);
            }
            const w = agg.weights;
            expect(w[0]).toBeGreaterThan(w[1]);
            expect(w[1]).toBeGreaterThan(w[2]);
        });

        it('should give equal weight to tied signals', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.RANK_BASED, feedbackDelay: 1, window: 10 });
            for (let i = 0; i < 5; i++) {
                agg.blend([0.5, 0.5]);
                agg.update(0.5);
            }
            const w = agg.weights;
            expect(w[0]).toBeCloseTo(w[1], 13);
        });
    });

    describe('Bayesian', () => {
        it('should start with uniform prior', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.BAYESIAN, feedbackDelay: 1 });
            for (const w of agg.weights) {
                expect(w).toBeCloseTo(1.0 / 3, 13);
            }
        });

        it('should use custom prior', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.BAYESIAN, feedbackDelay: 1, prior: [0.5, 0.3, 0.2] });
            const w = agg.weights;
            expect(w[0]).toBeCloseTo(0.5, 13);
            expect(w[1]).toBeCloseTo(0.3, 13);
            expect(w[2]).toBeCloseTo(0.2, 13);
        });

        it('should let good predictor dominate', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.BAYESIAN, feedbackDelay: 1 });
            for (let i = 0; i < 20; i++) {
                agg.blend([0.9, 0.1]);
                agg.update(0.9);
            }
            const w = agg.weights;
            expect(w[0]).toBeGreaterThan(0.9);
        });

        it('should let evidence override prior', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.BAYESIAN, feedbackDelay: 1, prior: [0.1, 0.9] });
            for (let i = 0; i < 50; i++) {
                agg.blend([0.8, 0.2]);
                agg.update(0.8);
            }
            const w = agg.weights;
            expect(w[0]).toBeGreaterThan(w[1]);
        });
    });

    describe('Delayed feedback', () => {
        it('should pair with previous blend for delay=1', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.EXPONENTIAL_DECAY, feedbackDelay: 1, alpha: 1.0 });
            agg.blend([0.9, 0.1]);
            agg.blend([0.5, 0.5]);
            agg.update(0.9);
            const w = agg.weights;
            expect(w[0]).toBeCloseTo(1.0 / 1.2, 13);
            expect(w[1]).toBeCloseTo(0.2 / 1.2, 13);
        });

        it('should pair with blend from 2 bars ago for delay=2', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.EXPONENTIAL_DECAY, feedbackDelay: 2, alpha: 1.0 });
            agg.blend([0.9, 0.1]);
            agg.blend([0.5, 0.5]);
            agg.update(0.9);
            for (const w of agg.weights) {
                expect(w).toBeCloseTo(0.5, 13);
            }
            agg.blend([0.3, 0.7]);
            agg.update(0.9);
            const w = agg.weights;
            expect(w[0]).toBeCloseTo(1.0 / 1.2, 13);
            expect(w[1]).toBeCloseTo(0.2 / 1.2, 13);
        });

        it('should be noop without enough history', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.EXPONENTIAL_DECAY, feedbackDelay: 3, alpha: 0.5 });
            agg.blend([0.5, 0.5]);
            const wBefore = agg.weights;
            agg.update(0.5);
            expect(agg.weights).toEqual(wBefore);
        });
    });

    describe('Warmup', () => {
        it('should equal live replay', () => {
            const history: HistoryEntry[] = [
                { signals: [0.8, 0.3], outcome: 0.7 },
                { signals: [0.6, 0.5], outcome: 0.5 },
                { signals: [0.9, 0.2], outcome: 0.8 },
                { signals: [0.7, 0.4], outcome: 0.6 },
                { signals: [0.5, 0.6], outcome: 0.4 },
            ];

            const agg1 = new Aggregator({ nSignals: 2, method: AggregationMethod.EXPONENTIAL_DECAY, feedbackDelay: 1, alpha: 0.2 });
            agg1.warmup(history);

            const agg2 = new Aggregator({ nSignals: 2, method: AggregationMethod.EXPONENTIAL_DECAY, feedbackDelay: 1, alpha: 0.2 });
            const outcomes: number[] = [];
            for (const entry of history) {
                agg2.blend(entry.signals);
                outcomes.push(entry.outcome);
                const idx = outcomes.length - 1 - 1;
                if (idx >= 0) {
                    agg2.update(outcomes[idx]);
                }
            }

            expect(agg1.count).toBe(agg2.count);
            const w1 = agg1.weights;
            const w2 = agg2.weights;
            for (let i = 0; i < w1.length; i++) {
                expect(w1[i]).toBeCloseTo(w2[i], 13);
            }
        });

        it('should equal live replay with delay=2', () => {
            const history: HistoryEntry[] = [
                { signals: [0.8, 0.3, 0.5], outcome: 0.7 },
                { signals: [0.6, 0.5, 0.4], outcome: 0.5 },
                { signals: [0.9, 0.2, 0.6], outcome: 0.8 },
                { signals: [0.7, 0.4, 0.3], outcome: 0.6 },
                { signals: [0.5, 0.6, 0.7], outcome: 0.4 },
                { signals: [0.4, 0.7, 0.2], outcome: 0.3 },
            ];

            const agg1 = new Aggregator({ nSignals: 3, method: AggregationMethod.MULTIPLICATIVE_WEIGHTS, feedbackDelay: 2, eta: 0.3 });
            agg1.warmup(history);

            const agg2 = new Aggregator({ nSignals: 3, method: AggregationMethod.MULTIPLICATIVE_WEIGHTS, feedbackDelay: 2, eta: 0.3 });
            const outcomes: number[] = [];
            for (const entry of history) {
                agg2.blend(entry.signals);
                outcomes.push(entry.outcome);
                const idx = outcomes.length - 1 - 2;
                if (idx >= 0) {
                    agg2.update(outcomes[idx]);
                }
            }

            expect(agg1.count).toBe(agg2.count);
            const w1 = agg1.weights;
            const w2 = agg2.weights;
            for (let i = 0; i < w1.length; i++) {
                expect(w1[i]).toBeCloseTo(w2[i], 13);
            }
        });

        it('should produce calibrated Bayesian weights', () => {
            const history: HistoryEntry[] = [];
            for (let i = 0; i < 20; i++) {
                history.push({ signals: [0.9, 0.1], outcome: 0.9 });
            }

            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.BAYESIAN, feedbackDelay: 1 });
            agg.warmup(history);
            const w = agg.weights;
            expect(w[0]).toBeGreaterThan(0.9);
        });
    });

    describe('Weights property', () => {
        it('should return a copy', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.EQUAL });
            const w = agg.weights;
            w[0] = 999.0;
            expect(agg.weights[0]).toBeCloseTo(0.5, 13);
        });

        it('should sum to one for all methods', () => {
            const methods = [
                AggregationMethod.FIXED,
                AggregationMethod.EQUAL,
                AggregationMethod.INVERSE_VARIANCE,
                AggregationMethod.EXPONENTIAL_DECAY,
                AggregationMethod.MULTIPLICATIVE_WEIGHTS,
                AggregationMethod.RANK_BASED,
                AggregationMethod.BAYESIAN,
            ];
            for (const method of methods) {
                const params: any = { nSignals: 2, method };
                if (method === AggregationMethod.FIXED) {
                    params.weights = [0.3, 0.7];
                }
                const agg = new Aggregator(params);
                const sum = agg.weights.reduce((a, b) => a + b, 0);
                expect(sum).toBeCloseTo(1.0, 13);
            }
        });
    });

    describe('Edge cases', () => {
        it('should return signal itself for single signal', () => {
            const methods = [
                AggregationMethod.FIXED,
                AggregationMethod.EQUAL,
                AggregationMethod.INVERSE_VARIANCE,
                AggregationMethod.EXPONENTIAL_DECAY,
                AggregationMethod.MULTIPLICATIVE_WEIGHTS,
                AggregationMethod.RANK_BASED,
                AggregationMethod.BAYESIAN,
            ];
            for (const method of methods) {
                const params: any = { nSignals: 1, method };
                if (method === AggregationMethod.FIXED) {
                    params.weights = [1.0];
                }
                const agg = new Aggregator(params);
                const result = agg.blend([0.73]);
                expect(result).toBeCloseTo(0.73, 13);
            }
        });

        it('should handle many signals', () => {
            const n = 100;
            const agg = new Aggregator({ nSignals: n, method: AggregationMethod.EQUAL });
            const signals = Array(n).fill(0.5);
            const result = agg.blend(signals);
            expect(result).toBeCloseTo(0.5, 13);
        });

        it('should handle extreme signals', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.EQUAL });
            const result = agg.blend([0.0, 1.0]);
            expect(result).toBeCloseTo(0.5, 13);
        });

        it('should handle Bayesian with extreme signals', () => {
            const agg = new Aggregator({ nSignals: 2, method: AggregationMethod.BAYESIAN, feedbackDelay: 1 });
            agg.blend([0.0, 1.0]);
            agg.blend([0.0, 1.0]);
            agg.update(1.0);
            const w = agg.weights;
            const sum = w.reduce((a, b) => a + b, 0);
            expect(sum).toBeCloseTo(1.0, 13);
        });

        it('should give equal weights for identical signals with inverse variance', () => {
            const agg = new Aggregator({ nSignals: 3, method: AggregationMethod.INVERSE_VARIANCE, feedbackDelay: 1, window: 10 });
            for (let i = 0; i < 5; i++) {
                agg.blend([0.5, 0.5, 0.5]);
                agg.update(0.5);
            }
            const w = agg.weights;
            for (let i = 0; i < 3; i++) {
                expect(w[i]).toBeCloseTo(1.0 / 3, 10);
            }
        });
    });
});
