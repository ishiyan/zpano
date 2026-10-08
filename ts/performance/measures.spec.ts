import { RawMomentsKleinKbn } from '../streaming-kbn';

import { Measures } from './measures';
import {
    SQRT2, SeededRandom, addBacon, assertAlmostEqual, assertAssertionFails, assertEqual,
    assertFalse, assertFloatEqual, assertSeriesEqual, assertThrows, assertTrue,
    bacon2023PortfolioLen, bacon2023PortfolioReturns, baconBenchmarkReturns, baconPortfolioLen,
    baconPortfolioReturns, evaluate, fsum, makeMeasures, mar, prod, pySum, publicMeasures,
    PUBLIC_METHODS, PUBLIC_PROPERTIES, rf, runStreamCallback, runStreamMethod, runStreamProperty,
} from './measures-test-helpers';

import * as rdBernardoLedoitRatio from './reference-data/bernardo-ledoit-ratio';
import * as rdCumulativeGeometricReturn from './reference-data/cumulative-geometric-return';
import * as rdDRatio from './reference-data/d-ratio';
import * as rdDownsideDeviation from './reference-data/downside-deviation';
import * as rdDownsideFrequency from './reference-data/downside-frequency';
import * as rdDownsidePotential from './reference-data/downside-potential';
import * as rdDownsideSharpeRatio from './reference-data/downside-sharpe-ratio';
import * as rdAdjustedSharpeRatio from './reference-data/adjusted-sharpe-ratio';
import * as rdEs from './reference-data/es';
import * as rdGeometricMeanReturn from './reference-data/geometric-mean-return';
import * as rdJarqueBera from './reference-data/jarque-bera-normality-test-statistic';
import * as rdKappaRatio from './reference-data/kappa-ratio';
import * as rdKurtosis from './reference-data/kurtosis';
import * as rdLossRate from './reference-data/loss-rate';
import * as rdMeanLossReturn from './reference-data/mean-loss-return';
import * as rdMeanNonZeroReturn from './reference-data/mean-non-zero-return';
import * as rdMeanWinReturn from './reference-data/mean-win-return';
import * as rdOmegaExcessReturn from './reference-data/omega-excess-return';
import * as rdOmegaRatio from './reference-data/omega-ratio';
import * as rdOmegaSharpeRatio from './reference-data/omega-sharpe-ratio';
import * as rdProbabilisticSharpeRatio from './reference-data/probabilistic-sharpe-ratio';
import * as rdProspectRatio from './reference-data/prospect-ratio';
import * as rdRachevRatio from './reference-data/rachev-ratio';
import * as rdSemiDeviation from './reference-data/semi-deviation';
import * as rdSharpeRatio from './reference-data/sharpe-ratio';
import * as rdSkewness from './reference-data/skewness';
import * as rdSkewnessKurtosisRatio from './reference-data/skewness-kurtosis-ratio';
import * as rdSortinoRatio from './reference-data/sortino-ratio';
import * as rdUpsideFrequency from './reference-data/upside-frequency';
import * as rdUpsidePotentialRatio from './reference-data/upside-potential-ratio';
import * as rdUpsideRisk from './reference-data/upside-risk';
import * as rdVar from './reference-data/var';
import * as rdVolatilitySkewness from './reference-data/volatility-skewness';
import * as rdWinRate from './reference-data/win-rate';

/** snake_case → camelCase, for names built from reference-data keys. */
function camel(name: string): string {
    return name.replace(/_([a-z0-9])/g, (_, c: string) => c.toUpperCase());
}

describe('Measures', () => {

    describe('SeriesAssertions', () => {
        it('rejects truncated series', () => {
            assertAssertionFails(() => assertSeriesEqual([1.0], [1.0, 2.0]));
            assertAssertionFails(() => assertSeriesEqual([1.0, 2.0], [1.0]));
        });

        it('accepts generator', () => {
            function* gen() {
                yield 1.0;
                yield 2.0;
            }
            assertSeriesEqual(gen(), [1.0, 2.0]);
        });

        it('relative tolerance for large reference values', () => {
            assertSeriesEqual([1e12 + 0.1], [1e12], { relTol: 1e-12 });
            assertAssertionFails(() => assertSeriesEqual([1e12 + 2], [1e12], { relTol: 1e-12 }));
        });
    });

    describe('EdgeCases', () => {
        it('public measures list is complete', () => {
            // Not in Python: checks the explicit list against the class, since
            // TypeScript has no equivalent of inspect.getmembers().
            const proto = Measures.prototype as unknown as Record<string, unknown>;
            const names = Object.getOwnPropertyNames(Measures.prototype)
                .filter(n => n !== 'constructor' && n !== 'reset' && n !== 'addReturn' && !n.startsWith('_'))
                .sort();
            expect(publicMeasures()).toEqual(names);
            expect(names.length).toBe(144);
            for (const n of PUBLIC_PROPERTIES) {
                const d = Object.getOwnPropertyDescriptor(Measures.prototype, n);
                expect(d?.get).withContext(n).toBeDefined();
            }
            for (const n of PUBLIC_METHODS) {
                expect(typeof proto[n]).withContext(n).toBe('function');
            }
        });

        it('empty', () => {
            // No property or method throws before the first return.
            const m = makeMeasures();
            for (const name of publicMeasures()) {
                expect(() => evaluate(m, name)).withContext(name).not.toThrow();
            }
        });

        it('single return', () => {
            const m = makeMeasures();
            m.addReturn(0.01, 0.02);
            for (const name of publicMeasures()) {
                expect(() => evaluate(m, name)).withContext(name).not.toThrow();
            }
            assertTrue(Number.isNaN(m.sharpeRatio));
            assertAlmostEqual(m.cumulativeGeometricReturn, 0.01, { places: 15 });
        });

        it('first negative return is a drawdown', () => {
            // As in PerformanceAnalytics Drawdowns(), the high-water mark
            // starts at the initial equity 1.
            const m = makeMeasures();
            m.addReturn(-0.05, -0.02);
            m.addReturn(0.02, 0.01);
            assertSeriesEqual(m.drawdownsHighWatermark, [-0.05, -0.031], { places: 15 });
            assertSeriesEqual(m.drawdownsCumulative, [-0.05, -0.031], { places: 15 });
            assertAlmostEqual(m.worstDrawdownsCumulative, 0.05, { places: 15 });
            assertAlmostEqual(m.painIndex, (0.05 + 0.031) / 2, { places: 15 });
            assertAlmostEqual(m.drawdownAverage, 0.05, { places: 15 });
        });

        it('long daily series', () => {
            // Every measure works with more observations than periods per annum.
            const rng = new SeededRandom(1);
            const m = new Measures(252.0);
            for (let i = 0; i < 300; i++) {
                const r = rng.gauss(0.0005, 0.01);
                const b = rng.gauss(0.0004, 0.01);
                m.addReturn(r, b);
            }
            for (const name of publicMeasures()) {
                expect(() => evaluate(m, name)).withContext(name).not.toThrow();
            }
            assertTrue(Number.isFinite(m.autocorrelationPenalty));
        });

        it('constructor rejects non-positive periods per annum', () => {
            // Not in Python as a separate test: the documented ValueError.
            expect(() => new Measures(0)).toThrowError('periods_per_annum must be positive');
            expect(() => new Measures(-1)).toThrowError('periods_per_annum must be positive');
        });
    });

    describe('AutocorrelationPenalty', () => {
        it('metamorphic properties', () => {
            // Constant returns
            let returns = Array.from({ length: bacon2023PortfolioLen }, () => 0.01);
            let actual = runStreamProperty('autocorrelationPenalty',
                { daily: true, returns, benchmarkReturns: returns });
            assertFloatEqual(actual[bacon2023PortfolioLen - 1], 1.0,
                { places: 15, prefix: 'autocorrelation penalty (constant)' });

            // Too few observations
            assertFloatEqual(actual[0], 1.0, { places: 15, prefix: 'autocorrelation penalty (len=0)' });
            assertFloatEqual(actual[1], 1.0, { places: 15, prefix: 'autocorrelation penalty (len=1)' });

            // Positive autocorrelation
            returns = Array.from({ length: bacon2023PortfolioLen }, (_, i) => 0.01 * i);
            actual = runStreamProperty('autocorrelationPenalty',
                { daily: true, returns, benchmarkReturns: returns });
            assertFloatEqual(actual[bacon2023PortfolioLen - 1], 2.722393904531189,
                { places: 15, prefix: 'autocorrelation penalty (positive)' });

            // Negative autocorrelation
            returns = Array.from({ length: bacon2023PortfolioLen }, (_, i) => i % 2 === 0 ? 0.01 : -0.01);
            actual = runStreamProperty('autocorrelationPenalty',
                { daily: true, returns, benchmarkReturns: returns });
            assertFloatEqual(actual[bacon2023PortfolioLen - 1], 0.16903085094570597,
                { places: 15, prefix: 'autocorrelation penalty (negative)' });

            // Scale and translation invariance
            const expected = runStreamProperty('autocorrelationPenalty', { daily: true });
            for (const scale of [4.2, -4.2]) {
                for (const shift of [0.042, -0.042]) {
                    const transformed = baconPortfolioReturns.map(r => scale * r + shift);
                    actual = runStreamProperty('autocorrelationPenalty',
                        { daily: true, returns: transformed, benchmarkReturns: transformed });
                    assertSeriesEqual(actual, expected, { places: 15,
                        prefix: `autocorrelation penalty (transform) scale ${scale} shift ${shift}` });
                }
            }
        });
    });

    describe('CumulativeGeometricReturn', () => {
        it('matches PerformanceAnalytics output', () => {
            const expected = rdCumulativeGeometricReturn.EXPECTED_VALUES;
            let actual = runStreamProperty('cumulativeGeometricReturn');
            assertSeriesEqual(actual, expected, { places: 14, prefix: 'cumulative geometric return (yearly)' });
            actual = runStreamProperty('cumulativeGeometricReturn', { monthly: true });
            assertSeriesEqual(actual, expected, { places: 14, prefix: 'cumulative geometric return (monthly)' });
            actual = runStreamProperty('cumulativeGeometricReturn', { daily: true });
            assertSeriesEqual(actual, expected, { places: 14, prefix: 'cumulative geometric return (daily)' });
        });
    });

    describe('GeometricMeanReturn', () => {
        it('matches PerformanceAnalytics output', () => {
            const expected = rdGeometricMeanReturn.EXPECTED_VALUES_GEOMETRIC;
            let actual = runStreamProperty('geometricMeanReturn');
            assertSeriesEqual(actual, expected, { places: 15, prefix: 'geometric mean return (yearly)' });
            actual = runStreamProperty('geometricMeanReturn', { monthly: true });
            assertSeriesEqual(actual, expected, { places: 15, prefix: 'geometric mean return (monthly)' });
            actual = runStreamProperty('geometricMeanReturn', { daily: true });
            assertSeriesEqual(actual, expected, { places: 15, prefix: 'geometric mean return (daily)' });
        });
    });

    describe('CompoundAnnualGrowthRate', () => {
        it('annualized return definition', () => {
            const calculate = (measures: Measures, ppa: number): number => {
                const w = (measures as unknown as { _returns: number[] })._returns;
                const growth = prod(w.map(r => 1 + r));
                return growth ** (ppa / w.length) - 1;
            };

            let expected = runStreamCallback(r => calculate(r, 1));
            let actual = runStreamProperty('compoundAnnualGrowthRate');
            assertSeriesEqual(actual, expected, { places: 15, prefix: 'compound annual growth rate (yearly)' });

            expected = runStreamCallback(r => calculate(r, 12), { monthly: true });
            actual = runStreamProperty('compoundAnnualGrowthRate', { monthly: true });
            assertSeriesEqual(actual, expected, { places: 14, prefix: 'compound annual growth rate (monthly)' });

            expected = runStreamCallback(r => calculate(r, 252), { daily: true });
            actual = runStreamProperty('compoundAnnualGrowthRate', { daily: true });
            assertSeriesEqual(actual, expected, { places: 11, prefix: 'compound annual growth rate (daily)' });
        });
    });

    describe('Skewness', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [method, expected] of rdSkewness.EXPECTED_VALUES_BY_METHOD) {
                const name = camel(`skewness_${method}`);
                let actual = runStreamProperty(name);
                assertSeriesEqual(actual, expected, { places: 14, prefix: name });
                if (method === 'moment') {
                    actual = runStreamProperty('skewness');
                    assertSeriesEqual(actual, expected, { places: 14, prefix: 'skewness' });
                }
            }
        });

        it('raw moments KleinKbn', () => {
            for (const [method, expected] of rdSkewness.EXPECTED_VALUES_BY_METHOD) {
                const bias = method === 'fisher' ? false : true;
                const kbn = new RawMomentsKleinKbn(1, bias, true);
                for (let i = 0; i < baconPortfolioLen; i++) {
                    kbn.update(baconPortfolioReturns[i]);
                    if (method === 'moment') {
                        assertFloatEqual(kbn.skewnessMoment, kbn.skewness,
                            { places: 15, prefix: `step ${i} skewness_${method} vs. skewness` });
                    } else if (method === 'fisher') {
                        assertFloatEqual(kbn.skewnessFisher, kbn.skewness,
                            { places: 15, prefix: `step ${i} skewness_${method} vs. skewness` });
                    } else { // method === 'sample'
                        assertFloatEqual(kbn.skewnessSample, expected[i],
                            { places: 14, prefix: `step ${i} skewness_${method}` });
                    }
                }
            }
        });
    });

    describe('Kurtosis', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [method, expected] of rdKurtosis.EXPECTED_VALUES_BY_METHOD) {
                const name = camel(`kurtosis_${method}`);
                let actual = runStreamProperty(name);
                assertSeriesEqual(actual, expected, { places: 13, prefix: name });
                if (method === 'excess') {
                    actual = runStreamProperty('kurtosis');
                    assertSeriesEqual(actual, expected, { places: 13, prefix: 'kurtosis' });
                }
            }
        });

        it('raw moments KleinKbn', () => {
            for (const [method, expected] of rdKurtosis.EXPECTED_VALUES_BY_METHOD) {
                // Default to 'excess'.
                let bias = true;
                let fisher = true;
                if (method === 'moment') {
                    fisher = false;
                } else if (method === 'sample_corrected') {
                    bias = false;
                    fisher = false;
                } else if (method === 'sample_excess') {
                    bias = false;
                }
                const kbn = new RawMomentsKleinKbn(1, bias, fisher);
                for (let i = 0; i < baconPortfolioLen; i++) {
                    kbn.update(baconPortfolioReturns[i]);
                    let actual: number;
                    if (method === 'excess') {
                        actual = kbn.kurtosisExcess;
                    } else if (method === 'moment') {
                        actual = kbn.kurtosisMoment;
                    } else if (method === 'sample_corrected') {
                        actual = kbn.kurtosisSampleCorrected;
                    } else { // method === 'sample_excess'
                        actual = kbn.kurtosisSampleExcess;
                    }
                    const dispatched = method === 'sample_corrected' ? kbn.kurtosisSample : actual;
                    assertFloatEqual(dispatched, kbn.kurtosis,
                        { places: 15, prefix: `step ${i} kurtosis_${method} / kurtosis` });
                    assertFloatEqual(actual, expected[i],
                        { places: 13, prefix: `step ${i} kurtosis_${method}` });
                }
            }
        });
    });

    describe('SkewnessKurtosisRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            const expected = rdSkewnessKurtosisRatio.EXPECTED_VALUES;
            const actual = runStreamProperty('skewnessKurtosisRatio');
            assertSeriesEqual(actual, expected, { places: 14, prefix: 'skewness-kurtosis ratio' });
        });
    });

    describe('JarqueBeraNormalityTestStatistic', () => {
        it('matches Bacon3 output', () => {
            const actual = runStreamProperty('jarqueBeraNormalityTestStatistic',
                { returns: bacon2023PortfolioReturns, benchmarkReturns: bacon2023PortfolioReturns });
            const expected = 0.34; // Chapter 5, exhibit 5.4
            assertFloatEqual(actual[bacon2023PortfolioLen - 1], expected,
                { places: 2, prefix: 'Jarque-Bera normality (bacon3)' });
        });

        it('matches scipy output', () => {
            const expected = rdJarqueBera.EXPECTED_VALUES;
            const actual = runStreamProperty('jarqueBeraNormalityTestStatistic');
            assertSeriesEqual(actual, expected, { places: 14, prefix: 'Jarque-Bera normality (scipy)' });
        });
    });

    describe('IsNormalDistribution', () => {
        it('mocked JB', () => {
            // Python patches the property with a PropertyMock; Jasmine's
            // spyOnProperty replaces the getter in the same way.
            const ratios = new Measures(1, 0, 0);
            const jb = spyOnProperty(ratios, 'jarqueBeraNormalityTestStatistic', 'get');

            // Test normality accepted: since 5.0 < 5.991..., the method should return true.
            jb.and.returnValue(5.0);
            assertTrue(ratios.isNormalDistribution(), 'normality accepted');

            // Test normality rejected
            jb.and.returnValue(10.0);
            assertFalse(ratios.isNormalDistribution(), 'normality rejected');

            // Test NaN statistic
            jb.and.returnValue(NaN);
            assertFalse(ratios.isNormalDistribution(), 'NaN statistic');

            // Test invalid confidence
            jb.and.returnValue(0.0);
            assertThrows(() => ratios.isNormalDistribution(1.0));
            assertThrows(() => ratios.isNormalDistribution(0.0));

            // Test custom confidence
            jb.and.returnValue(8.0);
            assertTrue(ratios.isNormalDistribution(0.99), 'custom confidence 8');
            jb.and.returnValue(10.0);
            assertFalse(ratios.isNormalDistribution(0.99), 'custom confidence 10');
        });
    });

    describe('VarCornishFisher', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, expected] of rdVar.EXPECTED_VALUES_BY_P_CORNISH_FISHER) {
                const actual = runStreamMethod('varCornishFisher', {}, p);
                assertSeriesEqual(actual, expected, { places: 9, prefix: `var cornish-fisher p ${p}` });
            }
        });
    });

    describe('VarGaussian', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, expected] of rdVar.EXPECTED_VALUES_BY_P_GAUSSIAN) {
                const actual = runStreamMethod('varGaussian', {}, p);
                assertSeriesEqual(actual, expected, { places: 9, prefix: `var gaussian p ${p}` });
            }
        });
    });

    describe('VarHistorical', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, expected] of rdVar.EXPECTED_VALUES_BY_P_HISTORICAL) {
                const actual = runStreamMethod('varHistorical', {}, p);
                assertSeriesEqual(actual, expected,
                    { places: p === 0.999 ? 4 : 15, prefix: `var historical p ${p}` });
            }
        });
    });

    describe('EsCornishFisher', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, expected] of rdEs.EXPECTED_VALUES_BY_P_CORNISH_FISHER) {
                const actual = runStreamMethod('esCornishFisher', {}, p);
                assertSeriesEqual(actual, expected,
                    { places: p < 0.995 ? 9 : (p < 0.999 ? 8 : 7), prefix: `es cornish-fisher p ${p}` });
            }
        });
    });

    describe('EsGaussian', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, expected] of rdEs.EXPECTED_VALUES_BY_P_GAUSSIAN) {
                const actual = runStreamMethod('esGaussian', {}, p);
                assertSeriesEqual(actual, expected, { places: p < 0.995 ? 9 : 8, prefix: `es gaussian p ${p}` });
            }
        });
    });

    describe('EsHistorical', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, expected] of rdEs.EXPECTED_VALUES_BY_P_HISTORICAL) {
                const actual = runStreamMethod('esHistorical', {}, p);
                assertSeriesEqual(actual, expected, { places: 15, prefix: `es historical p ${p}` });
            }
        });
    });

    describe('RewardToVarEsRatios', () => {
        // rewardToVarRatio* and rewardToEsRatio* divide the mean excess
        // return by the VaR/ES of the raw returns.
        const NAMES: readonly [string, string, string][] = [
            // [VaR/ES method, reward method, Sharpe method]
            ['varHistorical', 'rewardToVarRatioHistorical', 'sharpeRatioVarHistorical'],
            ['varGaussian', 'rewardToVarRatioGaussian', 'sharpeRatioVarGaussian'],
            ['varCornishFisher', 'rewardToVarRatioCornishFisher', 'sharpeRatioVarCornishFisher'],
            ['esHistorical', 'rewardToEsRatioHistorical', 'sharpeRatioEsHistorical'],
            ['esGaussian', 'rewardToEsRatioGaussian', 'sharpeRatioEsGaussian'],
            ['esCornishFisher', 'rewardToEsRatioCornishFisher', 'sharpeRatioEsCornishFisher'],
        ];

        it('zero risk-free rate equals Sharpe variants', () => {
            for (const [name, rewardName, sharpeName] of NAMES) {
                const reward = runStreamMethod(rewardName);
                const sharpe = runStreamMethod(sharpeName);
                assertSeriesEqual(reward, sharpe, { places: 14, prefix: `reward_to ${name}`, skip: 1 });
            }
        });

        it('definition', () => {
            const annualRf = 0.05;
            for (const [name, rewardName] of NAMES) {
                for (const confidence of [0.9, 0.95]) {
                    const m = makeMeasures(0, annualRf, 0, false, true);
                    const methods = m as unknown as Record<string, (c: number) => number>;
                    for (let i = 0; i < baconPortfolioLen; i++) {
                        m.addReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i]);
                        const excessMean = fsum(baconPortfolioReturns.slice(0, i + 1)
                            .map(r => r - m.riskFreeRate)) / (i + 1);
                        const denom = methods[name].call(m, confidence);
                        const expected = denom !== 0 ? excessMean / denom : NaN;
                        const actual = methods[rewardName].call(m, confidence);
                        assertFloatEqual(actual, expected,
                            { places: 14, prefix: `${rewardName} confidence ${confidence} step ${i}` });
                    }
                }
            }
        });
    });

    describe('MeanAbsoluteDeviationRatio', () => {
        it('exact values', () => {
            // mean / (sum|r - mean| / n) on the Bacon portfolio returns,
            // computed with exact rational arithmetic (fractions.Fraction).
            const expected = [
                NaN, 1.2608695652173914, 1.5789473684210527, 0.6818181818181818,
                0.9, 1.129032258064516, 1.308695652173913, 1.2618556701030927,
                0.9623076923076923, 1.0358796296296295, 0.9166666666666666, 0.96045197740113,
                1.0325794291868604, 0.7659033078880407, 0.480644111906311, 0.5178463399879009,
                0.3424072265625, 0.276536312849162, 0.3738222796970257, 0.4428828239908482,
                0.29573420836751435, 0.3247753530166881, 0.3100659077291792, 0.289544235924933];
            const actual = runStreamProperty('meanAbsoluteDeviationRatio');
            assertSeriesEqual(actual, expected, { places: 15, prefix: 'mean absolute deviation ratio' });
        });
    });

    describe('UpsidePotentialRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdUpsidePotentialRatio.EXPECTED_VALUES_BY_MAR_FULL) {
                const actual = runStreamProperty('upsidePotentialRatio', mar(m));
                assertSeriesEqual(actual, expected, { places: 14, prefix: `dupside potential ratio (full) MAR ${m}` });
            }
        });
    });

    describe('UpsidePotentialRatioSubset', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdUpsidePotentialRatio.EXPECTED_VALUES_BY_MAR_SUBSET) {
                const actual = runStreamProperty('upsidePotentialRatioSubset', mar(m));
                assertSeriesEqual(actual, expected, { places: 14, prefix: `dupside potential ratio (subset) MAR ${m}` });
            }
        });
    });

    describe('UpsideFrequency', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdUpsideFrequency.EXPECTED_VALUES_BY_MAR) {
                const actual = runStreamProperty('upsideFrequency', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `upside frequency MAR ${m}` });
            }
        });
    });

    describe('UpsidePotential', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdUpsideRisk.EXPECTED_VALUES_BY_MAR_POTENTIAL_FULL) {
                const actual = runStreamProperty('upsidePotential', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `upside potential (full) MAR ${m}` });
            }
            for (const [m, expected] of rdUpsideRisk.EXPECTED_VALUES_BY_MAR_POTENTIAL_FULL) {
                const annualMar = (1 + m) ** 252 - 1;
                const actual = runStreamProperty('upsidePotential', mar(annualMar, { daily: true }));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `upside potential (full) daily MAR ${m}` });
            }
        });
    });

    describe('UpsidePotentialSubset', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdUpsideRisk.EXPECTED_VALUES_BY_MAR_POTENTIAL_SUBSET) {
                const actual = runStreamProperty('upsidePotentialSubset', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `upside potential (subset) MAR ${m}` });
            }
        });
    });

    describe('UpsideVariance', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdUpsideRisk.EXPECTED_VALUES_BY_MAR_VARIANCE_FULL) {
                const actual = runStreamProperty('upsideVariance', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `upside variance (full) MAR ${m}` });
            }
        });
    });

    describe('UpsideVarianceSubset', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdUpsideRisk.EXPECTED_VALUES_BY_MAR_VARIANCE_SUBSET) {
                const actual = runStreamProperty('upsideVarianceSubset', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `upside variance (subset) MAR ${m}` });
            }
        });
    });

    describe('UpsideRisk', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdUpsideRisk.EXPECTED_VALUES_BY_MAR_RISK_FULL) {
                const actual = runStreamProperty('upsideRisk', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `upside risk (full) MAR ${m}` });
            }
        });
    });

    describe('UpsideRiskSubset', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdUpsideRisk.EXPECTED_VALUES_BY_MAR_RISK_SUBSET) {
                const actual = runStreamProperty('upsideRiskSubset', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `upside risk (subset) yearly MAR ${m}` });
            }
            for (const [m, expected] of rdUpsideRisk.EXPECTED_VALUES_BY_MAR_RISK_SUBSET) {
                const annualMar = (1 + m) ** 252 - 1;
                const actual = runStreamProperty('upsideRiskSubset', mar(annualMar, { daily: true }));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `upside risk (subset) daily MAR ${m}` });
            }
        });
    });

    describe('SemiDeviation', () => {
        it('matches PerformanceAnalytics output', () => {
            const actual = runStreamProperty('semiDeviation');
            assertSeriesEqual(actual, rdSemiDeviation.EXPECTED_VALUES, { places: 15, prefix: 'semi-deviation' });
        });
    });

    describe('DownsideDeviation', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdDownsideDeviation.EXPECTED_VALUES_BY_MAR_FULL) {
                const actual = runStreamProperty('downsideDeviation', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `downside deviation MAR ${m}` });
            }
        });
    });

    describe('DownsideDeviationSubset', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdDownsideDeviation.EXPECTED_VALUES_BY_MAR_SUBSET) {
                const actual = runStreamProperty('downsideDeviationSubset', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `downside deviation subset MAR ${m}` });
            }
        });
    });

    describe('DownsideFrequency', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdDownsideFrequency.EXPECTED_VALUES_BY_MAR) {
                const actual = runStreamProperty('downsideFrequency', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `downside frequency MAR ${m}` });
            }
        });
    });

    describe('DownsidePotential', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdDownsidePotential.EXPECTED_VALUES_BY_MAR) {
                const actual = runStreamProperty('downsidePotential', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `downside potential MAR ${m}` });
            }
        });
    });

    describe('SharpeRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdSharpeRatio.EXPECTED_VALUES_BY_RF_STDEV) {
                const actual = runStreamProperty('sharpeRatio', rf(r));
                assertSeriesEqual(actual, expected,
                    { places: r < 0.25 ? 13 : 12, prefix: `Sharpe ratio (stdev) Rf ${r}` });
            }
        });
    });

    describe('SharpeRatioVarHistorical', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, rfPack] of rdSharpeRatio.EXPECTED_VALUES_BY_P_RF_VAR_HISTORICAL) {
                for (const [r, expected] of rfPack) {
                    const actual = runStreamMethod('sharpeRatioVarHistorical', rf(r), p);
                    assertSeriesEqual(actual, expected,
                        { places: 12, prefix: `Sharpe ratio (VaR historical) conf ${p} Rf ${r}` });
                }
            }
        });
    });

    describe('SharpeRatioVarGaussian', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, rfPack] of rdSharpeRatio.EXPECTED_VALUES_BY_P_RF_VAR_GAUSSIAN) {
                for (const [r, expected] of rfPack) {
                    const actual = runStreamMethod('sharpeRatioVarGaussian', rf(r), p);
                    assertSeriesEqual(actual, expected,
                        { places: 5, prefix: `Sharpe ratio (VaR Gaussian) conf ${p} Rf ${r}` });
                }
            }
        });
    });

    describe('SharpeRatioVarCornishFisher', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, rfPack] of rdSharpeRatio.EXPECTED_VALUES_BY_P_RF_VAR_CORNISH_FISHER) {
                for (const [r, expected] of rfPack) {
                    const actual = runStreamMethod('sharpeRatioVarCornishFisher', rf(r), p);
                    assertSeriesEqual(actual, expected,
                        { places: 6, prefix: `Sharpe ratio (VaR Cornish-Fisher) conf ${p} Rf ${r}` });
                }
            }
        });
    });

    describe('SharpeRatioEsHistorical', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, rfPack] of rdSharpeRatio.EXPECTED_VALUES_BY_P_RF_ES_HISTORICAL) {
                for (const [r, expected] of rfPack) {
                    const actual = runStreamMethod('sharpeRatioEsHistorical', rf(r), p);
                    assertEqual(actual.length, expected.length);
                    for (let i = 0; i < actual.length; i++) {
                        const a = actual[i];
                        const e = expected[i];
                        const prefix = `Sharpe ratio (ES historical) conf ${p} Rf ${r} step ${i}`;
                        if (p === 0.9 && r === 0.001 && i === 10) {
                            // The quantile is exactly the second-worst return.
                            // This implementation includes both tied-to-tail
                            // observations; the R reference includes only one.
                            const excessMean = fsum(baconPortfolioReturns.slice(0, 11).map(x => x - r)) / 11;
                            assertAlmostEqual(a, excessMean / 0.013, { places: 13, msg: prefix });
                        } else {
                            assertFloatEqual(a, e, { delta: 1e-12, prefix });
                        }
                    }
                }
            }
        });
    });

    describe('SharpeRatioEsGaussian', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, rfPack] of rdSharpeRatio.EXPECTED_VALUES_BY_P_RF_ES_GAUSSIAN) {
                for (const [r, expected] of rfPack) {
                    const actual = runStreamMethod('sharpeRatioEsGaussian', rf(r), p);
                    assertSeriesEqual(actual, expected,
                        { places: 7, prefix: `Sharpe ratio (ES Gaussian) conf ${p} Rf ${r}` });
                }
            }
        });
    });

    describe('SharpeRatioEsCornishFisher', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, rfPack] of rdSharpeRatio.EXPECTED_VALUES_BY_P_RF_ES_CORNISH_FISHER) {
                for (const [r, expected] of rfPack) {
                    const actual = runStreamMethod('sharpeRatioEsCornishFisher', rf(r), p);
                    assertSeriesEqual(actual, expected,
                        { places: 5, prefix: `Sharpe ratio (ES Cornish-Fisher) conf ${p} Rf ${r}` });
                }
            }
        });
    });

    describe('DownsideSharpeRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdDownsideSharpeRatio.EXPECTED_VALUES_BY_RF) {
                const actual = runStreamProperty('downsideSharpeRatio', rf(r));
                assertSeriesEqual(actual, expected, { places: 13, prefix: `downside Sharpe ratio Rf ${r}` });
            }
        });
    });

    describe('AdjustedSharpeRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdAdjustedSharpeRatio.EXPECTED_VALUES_BY_RF) {
                const actual = runStreamProperty('adjustedSharpeRatio', rf(r));
                assertSeriesEqual(actual, expected,
                    { places: r < 0.2 ? 12 : 11, prefix: `adjusted Sharpe ratio (stdev) Rf ${r}` });
            }
        });
    });

    describe('AdjustedSharpeRatioSkewOnly', () => {
        it('scale and translation invariance', () => {
            // Since we don't have data for it, we test only its metamorphic properties.
            const r = 0.0042;
            const expected = runStreamProperty('adjustedSharpeRatioSkewOnly', rf(r));

            for (const scale of [4.2, -4.2]) {
                for (const shift of [0.042, -0.042]) {
                    const transformed = baconPortfolioReturns.map(x => scale * x + shift);
                    const ratios = new Measures(1, scale * r + shift, 0);
                    ratios.reset();
                    for (let i = 0; i < baconPortfolioLen; i++) {
                        ratios.addReturn(transformed[i], transformed[i]);
                        const a = ratios.adjustedSharpeRatioSkewOnly;
                        assertFloatEqual(scale > 0 ? a : -a, expected[i],
                            { places: 14, prefix: `ASR skew-only (scale ${scale} shift ${shift}) step ${i}` });
                    }
                }
            }
        });
    });

    describe('ProbabilisticSharpeRatio', () => {
        const cases: readonly [string, string, ReadonlyMap<number, ReadonlyMap<number, readonly number[]>>][] = [
            ['probabilisticSharpeRatio', '', rdProbabilisticSharpeRatio.EXPECTED_VALUES_BY_REFSR_RF],
            ['probabilisticSharpeRatioFull', ' (full)', rdProbabilisticSharpeRatio.EXPECTED_VALUES_BY_REFSR_RF_FULL],
            ['probabilisticSharpeRatioSymmetric', ' (symmetric)',
                rdProbabilisticSharpeRatio.EXPECTED_VALUES_BY_REFSR_RF_SYMMETRIC],
            ['probabilisticSharpeRatioGaussian', ' (Gaussian)',
                rdProbabilisticSharpeRatio.EXPECTED_VALUES_BY_REFSR_RF_GAUSSIAN],
        ];
        for (const [name, label, data] of cases) {
            it(`${name}: matches PerformanceAnalytics output`, () => {
                for (const [refSr, rfPack] of data) {
                    for (const [r, expected] of rfPack) {
                        const actual = runStreamMethod(name, rf(r), refSr);
                        assertSeriesEqual(actual, expected, { places: 14,
                            prefix: `probabilistic Sharpe ratio${label} reference_sr ${refSr} Rf ${r}` });
                    }
                }
            });
        }
    });

    describe('SortinoRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdSortinoRatio.EXPECTED_VALUES_BY_MAR) {
                const actual = runStreamProperty('sortinoRatio', mar(m));
                assertSeriesEqual(actual, expected, { places: 14, prefix: `sortino ratio MAR ${m}` });
            }
        });

        it('Jack Schwager sqrt2 version', () => {
            for (const [m, exp] of rdSortinoRatio.EXPECTED_VALUES_BY_MAR) {
                const expected = exp.map(x => x / Math.sqrt(2));
                const actual = runStreamProperty('sortinoRatioSqrt2', mar(m));
                assertSeriesEqual(actual, expected, { places: 14, prefix: `sortino ratio (sqrt2) MAR ${m}` });
            }
        });
    });

    describe('SortinoSatchellRatio', () => {
        it('should be computable', () => {
            const actual = runStreamProperty('sortinoSatchellRatio');
            assertFloatEqual(actual[actual.length - 1], 0.3923720287950653,
                { places: 15, prefix: 'Sortino-Satchell ratio' });
        });
    });

    describe('OmegaRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdOmegaRatio.EXPECTED_VALUES_BY_MAR) {
                const actual = runStreamProperty('omegaRatio', mar(m));
                assertSeriesEqual(actual, expected, { places: 13, prefix: `omega ratio MAR ${m}` });
            }
        });
    });

    describe('OmegaSharpeRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdOmegaSharpeRatio.EXPECTED_VALUES_BY_MAR) {
                const actual = runStreamProperty('omegaSharpeRatio', mar(m));
                assertSeriesEqual(actual, expected, { places: 13, prefix: `omega Dharpe ratio MAR ${m}` });
            }
        });
    });

    describe('OmegaExcessReturn', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdOmegaExcessReturn.EXPECTED_VALUES_BY_MAR_WITH_BENCHMARK) {
                const annualMar = (1 + m) ** 252 - 1;
                const actual = runStreamProperty('omegaExcessReturn', mar(annualMar, { daily: true }));
                assertSeriesEqual(actual, expected,
                    { places: 11, prefix: `omega excess return (with benchmak) MAR ${m}` });
            }
            for (const [m, expected] of rdOmegaExcessReturn.EXPECTED_VALUES_BY_MAR_WITH_SELF) {
                const annualMar = (1 + m) ** 252 - 1;
                const actual = runStreamProperty('omegaExcessReturn',
                    mar(annualMar, { daily: true, benchmarkReturns: baconPortfolioReturns }));
                assertSeriesEqual(actual, expected,
                    { places: 11, prefix: `omega excess return (with self) MAR ${m}` });
            }
        });
    });

    describe('KappaRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            // Generated data is yearly.
            const cases: readonly [string, ReadonlyMap<number, readonly number[]>, number][] = [
                ['kappa1Ratio', rdKappaRatio.EXPECTED_VALUES_BY_MAR_ORDER_1, 13],
                ['kappa2Ratio', rdKappaRatio.EXPECTED_VALUES_BY_MAR_ORDER_2, 14],
                ['kappa3Ratio', rdKappaRatio.EXPECTED_VALUES_BY_MAR_ORDER_3, 14],
                ['kappa4Ratio', rdKappaRatio.EXPECTED_VALUES_BY_MAR_ORDER_4, 14],
            ];
            for (const [name, data, places] of cases) {
                for (const [m, expected] of data) {
                    const actual = runStreamProperty(name, mar(m));
                    assertSeriesEqual(actual, expected, { places, prefix: `${name} MAR ${m}` });
                }
            }
        });
    });

    describe('ProspectRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdProspectRatio.EXPECTED_VALUES_BY_MAR_PERFAN) {
                const actual = runStreamProperty('prospectRatioPerformanceAnalytics', mar(m));
                assertSeriesEqual(actual, expected,
                    { places: 13, prefix: `Prospect ratio PerformanceAnalytics version (yearly, MAR ${m})` });
            }
        });

        it('matches reference implementation output', () => {
            for (const [m, expected] of rdProspectRatio.EXPECTED_VALUES_BY_MAR_REFERENCE) {
                const actual = runStreamMethod('prospectRatio', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `Prospect ratio (yearly, MAR ${m})` });
            }
        });
    });

    describe('BernardoLedoitRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            const expected = rdBernardoLedoitRatio.EXPECTED_VALUES;
            const actual = runStreamProperty('bernardoLedoitRatio');
            assertSeriesEqual(actual, expected, { places: 13, prefix: 'Bernado-Ledoit ratio' });
        });
    });

    describe('DRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            const expected = rdDRatio.EXPECTED_VALUES_PERFAN;
            const actual = runStreamProperty('dRatio');
            assertSeriesEqual(actual, expected, { places: 15, prefix: 'd-ratio' });
        });
    });

    describe('GainLossRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            const expected = rdBernardoLedoitRatio.EXPECTED_VALUES;
            const actual = runStreamProperty('gainLossRatio');
            assertSeriesEqual(actual, expected, { places: 13, prefix: 'gain-loss ratio' });
        });
    });

    describe('MeanNonZeroReturn', () => {
        it('calculated by hand', () => {
            const actual = runStreamProperty('meanNonZeroReturn');
            assertSeriesEqual(actual, rdMeanNonZeroReturn.EXPECTED_VALUES, { places: 15, prefix: 'mean non-zero return' });
        });
    });

    describe('MeanWinReturn', () => {
        it('calculated by hand', () => {
            const actual = runStreamProperty('meanWinReturn');
            assertSeriesEqual(actual, rdMeanWinReturn.EXPECTED_VALUES, { places: 15, prefix: 'mean win return' });
        });
    });

    describe('MeanLossReturn', () => {
        it('calculated by hand', () => {
            const actual = runStreamProperty('meanLossReturn');
            assertSeriesEqual(actual, rdMeanLossReturn.EXPECTED_VALUES, { places: 15, prefix: 'mean loss return' });
        });
    });

    describe('WinRate', () => {
        it('calculated by hand', () => {
            const actual = runStreamProperty('winRate');
            assertSeriesEqual(actual, rdWinRate.EXPECTED_VALUES, { places: 15, prefix: 'win rate' });
        });
    });

    describe('LossRate', () => {
        it('calculated by hand', () => {
            const actual = runStreamProperty('lossRate');
            assertSeriesEqual(actual, rdLossRate.EXPECTED_VALUES, { places: 15, prefix: 'loss rate' });
        });
    });

    describe('VolatilitySkewness', () => {
        function baconReference(expected: readonly number[]): readonly number[] {
            // The MAR 0.2 and 0.3 fixtures contain three extra trailing zeros.
            // Compare the 24 observations that correspond to the Bacon inputs.
            assertTrue(expected.slice(baconPortfolioLen).every(x => x === 0));
            return expected.slice(0, baconPortfolioLen);
        }

        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdVolatilitySkewness.EXPECTED_VALUES_BY_MAR_VARIABILITY) {
                const actual = runStreamProperty('volatilitySkewness', mar(m));
                assertSeriesEqual(actual, baconReference(expected), { places: 13, prefix: `volatility skewness MAR ${m}` });
            }
            for (const [m, expected] of rdVolatilitySkewness.EXPECTED_VALUES_BY_MAR_VOLATILITY) {
                const actual = runStreamProperty('variabilitySkewness', mar(m));
                assertSeriesEqual(actual, baconReference(expected), { places: 13, prefix: `variability skewness MAR ${m}` });
            }
        });
    });

    describe('FarinelliTibilettiRatio', () => {
        it('should be computable', () => {
            const m = 0.005;

            const verify = (upper: number, lower: number, relatedProperty: string,
                transform: (r: number) => number, places: number): void => {
                const expected = runStreamProperty(relatedProperty, mar(m));
                const actual = runStreamMethod('farinelliTibilettiRatio', mar(m), upper, lower);
                assertSeriesEqual(actual.map(transform), expected, { places,
                    prefix: `Farinelli-Tibiletti ratio (u ${upper}, l ${lower}) vs ${relatedProperty}` });
            };

            verify(1, 1, 'omegaRatio', r => r, 14);
            verify(1, 1, 'kappa1Ratio', r => r - 1, 14);
            verify(1, 2, 'upsidePotentialRatio', r => r, 15);
            verify(2, 2, 'volatilitySkewness', r => r, 14);
            verify(2, 2, 'variabilitySkewness', r => r * r, 13);

            const verifyManual = (upper: number, lower: number, places: number): void => {
                const pmax = (x: number): number => x > 0.0 ? x : 0.0; // Python max(x, 0.0)
                const upm = pySum(baconPortfolioReturns.map(r => pmax(r - m) ** upper)) / baconPortfolioLen;
                const lpm = pySum(baconPortfolioReturns.map(r => pmax(m - r) ** lower)) / baconPortfolioLen;
                const expected = upm ** (1.0 / upper) / lpm ** (1.0 / lower);
                const actual = runStreamMethod('farinelliTibilettiRatio', mar(m), upper, lower);
                assertFloatEqual(actual[actual.length - 1], expected, { places,
                    prefix: `Farinelli-Tibiletti ratio (u ${upper}, l ${lower}) vs manual calculation` });
            };

            for (const i of [1, 2, 3, 4]) {
                for (const j of [1, 2, 3, 4]) {
                    verifyManual(i, j, 15);
                }
            }
        });
    });

    describe('RachevRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            let alpha = 0.05;
            for (const [beta, bundle] of rdRachevRatio.EXPECTED_VALUES_BY_BETA_RF_ALFA_0_05) {
                for (const [r, expected] of bundle) {
                    const actual = runStreamMethod('rachevRatio', rf(r), alpha, beta);
                    assertSeriesEqual(actual, expected,
                        { places: 14, prefix: `Rachev ratio (alpha ${alpha} beta ${beta} Rf ${r})` });
                }
            }
            alpha = 0.1;
            for (const [beta, bundle] of rdRachevRatio.EXPECTED_VALUES_BY_BETA_RF_ALFA_0_1) {
                for (const [r, expected] of bundle) {
                    const actual = runStreamMethod('rachevRatio', rf(r), alpha, beta);
                    assertSeriesEqual(actual, expected,
                        { places: 14, prefix: `Rachev ratio (alpha ${alpha} beta ${beta} Rf ${r})` });
                }
            }
        });
    });

    describe('Errors', () => {
        // Not separate Python tests: the documented ValueError messages and
        // check ordering of the fallible methods.
        it('throw with the Python messages', () => {
            const m = makeMeasures();
            addBacon(m);
            expect(() => m.isNormalDistribution(1.5)).toThrowError('confidence must be between 0 and 1');
            expect(() => m.farinelliTibilettiRatio(0, 2)).toThrowError('upper_order must be 1, 2, 3, or 4');
            expect(() => m.farinelliTibilettiRatio(2, 5)).toThrowError('lower_order must be 1, 2, 3, or 4');
            expect(() => m.rachevRatio(0, 0.1)).toThrowError('alpha must be between 0 and 1');
            expect(() => m.rachevRatio(0.1, 1)).toThrowError('beta must be between 0 and 1');
            expect(() => m.cdarAverage(1)).toThrowError('confidence must be between 0 and 1');
            expect(() => m.cdarDiscrete(0)).toThrowError('confidence must be between 0 and 1');
            expect(() => m.cdarBeta(1)).toThrowError('confidence must be between 0 and 1');
            expect(() => m.cdarAlpha(0)).toThrowError('confidence must be between 0 and 1');
            expect(() => m.tailRatio(0.5)).toThrowError('cutoff must be between 0.5 and 1.0');
            expect(() => m.biasRatio(0)).toThrowError('std_dev_multiplier must be positive');
        });

        it('check ordering', () => {
            const m = makeMeasures();
            // JB is NaN before enough data: false before validating confidence.
            expect(m.isNormalDistribution(2)).toBeFalse();
            // Rachev checks n < 2 before validating alpha and beta.
            m.addReturn(0.01, 0.01);
            expect(m.rachevRatio(0, 0)).toBeNaN();
        });
    });
});
