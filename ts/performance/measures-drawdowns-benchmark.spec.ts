import { Measures } from './measures';
import {
    SeededRandom, addBacon, assertAlmostEqual, assertEqual, assertFloatEqual, assertSeriesEqual,
    assertThrows, assertTrue, bacon2023DrawdownContinuousWithoutZeroes, bacon2023DrawdownFromPeak,
    bacon2023PortfolioReturns, baconBenchmarkReturns, baconPortfolioLen, baconPortfolioReturns,
    evaluate, fmean, fsum, isClose, makeMeasures, mar, prod, pstdev, pySum, publicMeasures, rf,
    runStreamMethod, runStreamProperty, stdev,
} from './measures-test-helpers';

import * as rdActivePremium from './reference-data/active-premium';
import * as rdAppraisalRatio from './reference-data/appraisal-ratio';
import * as rdBiasRatio from './reference-data/bias-ratio';
import * as rdBurkeRatio from './reference-data/burke-ratio';
import * as rdCalmarRatio from './reference-data/calmar-ratio';
import * as rdCdar from './reference-data/cdar';
import * as rdCdarAlpha from './reference-data/cdar-alpha';
import * as rdCdarBeta from './reference-data/cdar-beta';
import * as rdDownNumberRatio from './reference-data/down-number-ratio';
import * as rdDownPercentageRatio from './reference-data/down-percentage-ratio';
import * as rdDownsideCaptureRatio from './reference-data/downside-capture-ratio';
import * as rdDrawdownAverage from './reference-data/drawdown-average';
import * as rdDrawdownAverageLength from './reference-data/drawdown-average-length';
import * as rdDrawdownAveragePeakToTrough from './reference-data/drawdown-average-peak-to-trough';
import * as rdDrawdownAverageRecovery from './reference-data/drawdown-average-recovery';
import * as rdDrawdownDeviation from './reference-data/drawdown-deviation';
import * as rdDrawdownsCumulative from './reference-data/drawdowns-cumulative';
import * as rdDrawdownsHighWatermark from './reference-data/drawdowns-high-watermark';
import * as rdFamaBeta from './reference-data/fama-beta';
import * as rdGainToPainRatio from './reference-data/gain-to-pain-ratio';
import * as rdHurstExponent from './reference-data/hurst-exponent';
import * as rdInformationRatio from './reference-data/information-ratio';
import * as rdJensenAlpha from './reference-data/jensen-alpha';
import * as rdJensenAlphaAlternative from './reference-data/jensen-alpha-alternative';
import * as rdJensenAlphaModified from './reference-data/jensen-alpha-modified';
import * as rdKRatio from './reference-data/k-ratio';
import * as rdKellyRatio from './reference-data/kelly-ratio';
import * as rdMSquared from './reference-data/m-squared';
import * as rdMSquaredExcess from './reference-data/m-squared-excess';
import * as rdMSquaredSortino from './reference-data/m-squared-sortino';
import * as rdMartinRatio from './reference-data/martin-ratio';
import * as rdMinDrawdownsCumulative from './reference-data/min-drawdowns-cumulative';
import * as rdModigliani from './reference-data/modigliani';
import * as rdOverallCaptureRatio from './reference-data/overall-capture-ratio';
import * as rdPainIndex from './reference-data/pain-index';
import * as rdPainRatio from './reference-data/pain-ratio';
import * as rdSfmAlpha from './reference-data/sfm-alpha';
import * as rdSfmBeta from './reference-data/sfm-beta';
import * as rdSfmBetaBear from './reference-data/sfm-beta-bear';
import * as rdSfmBetaBull from './reference-data/sfm-beta-bull';
import * as rdSfmR2 from './reference-data/sfm-r2';
import * as rdSfmRiskPremium from './reference-data/sfm-risk-premium';
import * as rdSpecificRisk from './reference-data/specific-risk';
import * as rdSterlingRatio from './reference-data/sterling-ratio';
import * as rdSystematicRisk from './reference-data/systematic-risk';
import * as rdTailRatio from './reference-data/tail-ratio';
import * as rdTimingRatio from './reference-data/timing-ratio';
import * as rdTotalRisk from './reference-data/total-risk';
import * as rdTrackingError from './reference-data/tracking-error';
import * as rdTreynorRatio from './reference-data/treynor-ratio';
import * as rdTreynorRatioModified from './reference-data/treynor-ratio-modified';
import * as rdUlcerIndex from './reference-data/ulcer-index';
import * as rdUpNumberRatio from './reference-data/up-number-ratio';
import * as rdUpPercentageRatio from './reference-data/up-percentage-ratio';
import * as rdUpsideCaptureRatio from './reference-data/upside-capture-ratio';

type Series = ReadonlyMap<number, readonly number[]>;

/** Periodic rate compounded to an annual rate, Python `(1 + r) ** periods - 1`. */
function annualize(r: number, periods: number): number {
    return (1 + r) ** periods - 1;
}

/**
 * Yearly run with the periodic rate, then a daily run with the rate
 * compounded over 252 periods (the common SFM test pattern).
 */
function checkYearlyAndDailyRf(name: string, data: Series, places: number, label: string,
    skip: (r: number) => number = () => 0): void {
    for (const [r, expected] of data) {
        const actual = runStreamProperty(name, rf(r));
        assertSeriesEqual(actual, expected, { places, skip: skip(r), prefix: `${label} (yearly, Rf ${r})` });
    }
    for (const [r, expected] of data) {
        const actual = runStreamProperty(name, rf(annualize(r, 252), { daily: true }));
        assertSeriesEqual(actual, expected, { places, skip: skip(r), prefix: `${label} (daily, Rf ${r})` });
    }
}

describe('Measures', () => {

    describe('DrawdownsCumulative', () => {
        it('matches PerformanceAnalytics output', () => {
            const actual = runStreamProperty<number[]>('drawdownsCumulative');
            for (const [i, expected] of rdDrawdownsCumulative.EXPECTED_VALUES_BY_INDEX) {
                assertSeriesEqual(actual[i], expected, { places: 15, prefix: `drawdowns cumulative (i ${i})` });
            }
        });

        it('matches Bacon 2023 output', () => {
            const actual = runStreamProperty<number[]>('drawdownsCumulative',
                { returns: bacon2023PortfolioReturns, benchmarkReturns: bacon2023PortfolioReturns });
            assertSeriesEqual(actual[actual.length - 1], bacon2023DrawdownFromPeak,
                { places: 4, prefix: 'drawdowns cumulative (bacon 2023)' });
        });
    });

    describe('MinDrawdownsCumulative', () => {
        it('matches PerformanceAnalytics output', () => {
            const actual = runStreamProperty('minDrawdownsCumulative');
            assertSeriesEqual(actual, rdMinDrawdownsCumulative.EXPECTED_VALUES,
                { places: 15, prefix: 'min drawdowns cumulative' });
        });
    });

    describe('WorstDrawdownsCumulative', () => {
        it('matches PerformanceAnalytics output', () => {
            const actual = runStreamProperty('worstDrawdownsCumulative');
            assertSeriesEqual(actual, rdMinDrawdownsCumulative.EXPECTED_VALUES_INVERTED,
                { places: 15, prefix: 'worst drawdowns cumulative' });
        });
    });

    describe('DrawdownsHighWatermark', () => {
        it('matches PerformanceAnalytics output', () => {
            const actual = runStreamProperty<number[]>('drawdownsHighWatermark');
            for (const [i, expected] of rdDrawdownsHighWatermark.EXPECTED_VALUES_BY_INDEX) {
                assertSeriesEqual(actual[i], expected, { places: 15, prefix: `drawdowns high_watermark (i ${i})` });
            }
        });
    });

    describe('DrawdownsContinuousRuns', () => {
        it('matches Bacon 2023 output', () => {
            // The book's four-decimal drawdown values are approximate; the core
            // run tracker has separate exact rolling-window tests.
            const all = runStreamMethod<number[]>('drawdownsContinuousRuns',
                { returns: bacon2023PortfolioReturns, benchmarkReturns: bacon2023PortfolioReturns });
            const actual = all[all.length - 1];
            assertSeriesEqual(actual, bacon2023DrawdownContinuousWithoutZeroes,
                { delta: 0.002, prefix: 'drawdowns continuous runs (bacon 2023)' });
        });
    });

    describe('CalmarRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            const expected = rdCalmarRatio.EXPECTED_VALUES;
            let actual = runStreamProperty('calmarRatio');
            assertSeriesEqual(actual, expected, { places: 13, prefix: 'calmar ratio (yearly)' });
            actual = runStreamProperty('calmarRatio', { daily: true });
            assertSeriesEqual(actual, expected, { places: 13, prefix: 'calmar ratio (daily)' });
        });
    });

    describe('SterlingRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [excess, expected] of rdSterlingRatio.EXPECTED_VALUES_BY_EXCESS) {
                const actual = runStreamMethod('sterlingRatio', {}, excess);
                assertSeriesEqual(actual, expected, { places: 13, prefix: `sterling ratio (yearly, excess ${excess})` });
            }
            for (const [excess, expected] of rdSterlingRatio.EXPECTED_VALUES_BY_EXCESS) {
                const actual = runStreamMethod('sterlingRatio', { daily: true }, excess);
                assertSeriesEqual(actual, expected, { places: 13, prefix: `sterling ratio (daily, excess ${excess})` });
            }
        });
    });

    describe('BurkeRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            // Generated reference data is yearly.
            for (const [r, expected] of rdBurkeRatio.EXPECTED_VALUES_BY_RF) {
                const actual = runStreamProperty('burkeRatio', rf(r));
                assertSeriesEqual(actual, expected, { places: 11, prefix: `burke ratio (yearly, Rf ${r})` });
            }
        });
    });

    describe('BurkeRatioModified', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdBurkeRatio.EXPECTED_VALUES_BY_RF_MODIFIED) {
                const actual = runStreamProperty('burkeRatioModified', rf(r));
                assertSeriesEqual(actual, expected, { places: 11, prefix: `burke ratio modified (Rf ${r})` });
            }
        });
    });

    describe('PainIndex', () => {
        it('matches PerformanceAnalytics output', () => {
            // R's drawdown series differs slightly from the high-water-mark
            // series used here. DocumentedFormulas checks the exact formula.
            const expected = rdPainIndex.EXPECTED_VALUES;
            let actual = runStreamProperty('painIndex');
            assertSeriesEqual(actual, expected, { delta: 0.00098, prefix: 'pain index (yearly)' });
            actual = runStreamProperty('painIndex', { daily: true });
            assertSeriesEqual(actual, expected, { delta: 0.00098, prefix: 'pain index (daily)' });
        });
    });

    describe('PainRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdPainRatio.EXPECTED_VALUES_BY_RF) {
                const delta = r < 0.04 ? 0.016 : 0.111;
                let actual = runStreamProperty('painRatio', rf(r));
                assertSeriesEqual(actual, expected, { delta, prefix: `pain ratio (yearly, Rf ${r})` });
                actual = runStreamProperty('painRatio', rf(annualize(r, 252), { daily: true }));
                assertSeriesEqual(actual, expected, { delta, prefix: `pain ratio (daily, Rf ${r})` });
            }
        });
    });

    describe('UlcerIndex', () => {
        it('matches PerformanceAnalytics output', () => {
            // The drawdown convention differs from the R fixture; the precise
            // root-mean-square definition is checked in DocumentedFormulas.
            const expected = rdUlcerIndex.EXPECTED_VALUES;
            let actual = runStreamProperty('ulcerIndex');
            assertSeriesEqual(actual, expected, { delta: 0.00192, prefix: 'ulcer index (yearly)' });
            actual = runStreamProperty('ulcerIndex', { daily: true });
            assertSeriesEqual(actual, expected, { delta: 0.00192, prefix: 'ulcer index (daily)' });
        });
    });

    describe('MartinRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdMartinRatio.EXPECTED_VALUES_BY_RF) {
                let actual = runStreamProperty('martinRatio', rf(r));
                assertSeriesEqual(actual, expected, { delta: 0.0630, prefix: `martin ratio (yearly, Rf ${r})` });
                actual = runStreamProperty('martinRatio', rf(annualize(r, 252), { daily: true }));
                assertSeriesEqual(actual, expected, { delta: 0.0630, prefix: `martin ratio (daily, Rf ${r})` });
            }
        });
    });

    const episodeCases: readonly [string, string, readonly number[]][] = [
        ['DrawdownAverage', 'drawdownAverage', rdDrawdownAverage.EXPECTED_VALUES],
        ['DrawdownAverageLength', 'drawdownAverageLength', rdDrawdownAverageLength.EXPECTED_VALUES],
        ['DrawdownAveragePeakToTrough', 'drawdownAveragePeakToTrough', rdDrawdownAveragePeakToTrough.EXPECTED_VALUES],
        ['DrawdownAverageRecovery', 'drawdownAverageRecovery', rdDrawdownAverageRecovery.EXPECTED_VALUES],
        ['DrawdownDeviation', 'drawdownDeviation', rdDrawdownDeviation.EXPECTED_VALUES],
    ];
    for (const [suite, name, expected] of episodeCases) {
        describe(suite, () => {
            it('matches PerformanceAnalytics output', () => {
                let actual = runStreamProperty(name);
                assertSeriesEqual(actual, expected, { places: 15, prefix: `${name} (yearly)` });
                actual = runStreamProperty(name, { daily: true });
                assertSeriesEqual(actual, expected, { places: 15, prefix: `${name} (daily)` });
            });
        });
    }

    describe('CDaRAverage', () => {
        it('matches PerformanceAnalytics output', () => {
            // R and this implementation select the continuous drawdown tail
            // differently. DocumentedFormulas checks our linear quantile.
            for (const [p, expected] of rdCdar.EXPECTED_VALUES_BY_P_AVERAGE_GEOMETRIC_INVERTED) {
                const actual = runStreamMethod('cdarAverage', {}, p);
                assertSeriesEqual(actual, expected, { delta: 0.02938, prefix: `CDaR average geometric (yearly) p ${p}` });
            }
            for (const [p, expected] of rdCdar.EXPECTED_VALUES_BY_P_AVERAGE_GEOMETRIC_INVERTED) {
                const actual = runStreamMethod('cdarAverage', { daily: true }, p);
                assertSeriesEqual(actual, expected, { delta: 0.02938, prefix: `CDaR average geometric (daily) p ${p}` });
            }
        });
    });

    describe('CDaRDiscrete', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [p, expected] of rdCdar.EXPECTED_VALUES_BY_P_DISCRETE_GEOMETRIC_INVERTED) {
                const actual = runStreamMethod('cdarDiscrete', {}, p);
                assertSeriesEqual(actual, expected, { places: 15, prefix: `CDaR discrete geometric (yearly) p ${p}` });
            }
            for (const [p, expected] of rdCdar.EXPECTED_VALUES_BY_P_DISCRETE_GEOMETRIC_INVERTED) {
                const actual = runStreamMethod('cdarDiscrete', { daily: true }, p);
                assertSeriesEqual(actual, expected, { places: 15, prefix: `CDaR discrete geometric (daily) p ${p}` });
            }
        });
    });

    describe('CDaRBeta', () => {
        it('discrete tail selection', () => {
            const m = makeMeasures();
            for (const [portfolio, benchmark] of [[0.1, 0.1], [-0.05, -0.1], [0.2, 0.2], [-0.1, -0.2],
                [0.3, 0.3], [-0.15, -0.3]]) {
                m.addReturn(portfolio, benchmark);
            }
            // At 50% confidence, two of three episodes are selected. The
            // denominator is the second-worst depth, -0.2.
            assertAlmostEqual(m.cdarBeta(0.5), (-0.15 - 0.1) / (2 * -0.2), { places: 14 });
            assertAlmostEqual(m.cdarBeta(0.8), -0.15 / -0.3, { places: 14 });
        });

        it('matches PerformanceAnalytics output', () => {
            for (const [p, expected] of rdCdarBeta.EXPECTED_VALUES_BY_P_GEOMETRIC) {
                const actual = runStreamMethod('cdarBeta', {}, p);
                assertSeriesEqual(actual, expected, { places: 13, prefix: `CDaR beta geometric (yearly) p ${p}` });
            }
            for (const [p, expected] of rdCdarBeta.EXPECTED_VALUES_BY_P_GEOMETRIC) {
                const actual = runStreamMethod('cdarBeta', { daily: true }, p);
                assertSeriesEqual(actual, expected, { places: 13, prefix: `CDaR beta geometric (daily) p ${p}` });
            }
        });

        it('mathematical properties', () => {
            // With one selected episode, identical portfolio and benchmark
            // returns give the same numerator and denominator.
            let measures = makeMeasures();
            for (const ret of [0.05, -0.1]) {
                measures.addReturn(ret, ret);
            }
            assertFloatEqual(measures.cdarBeta(), 1.0, { places: 15, prefix: 'CDaR beta (one episode) identity' });

            measures = makeMeasures();
            addBacon(measures, 0, baconPortfolioLen, baconPortfolioReturns, baconPortfolioReturns);
            assertFloatEqual(measures.cdarBeta(), 1.0, { places: 14, prefix: 'CDaR beta (Bacon) identity' });

            // No drawdowns
            let transformed = makeMeasures();
            addBacon(transformed, 0, baconPortfolioLen,
                new Array(baconPortfolioLen).fill(0.01), new Array(baconPortfolioLen).fill(0.02));
            assertFloatEqual(transformed.cdarBeta(), NaN, { prefix: 'CDaR beta (geometric) no drawdowns' });

            // Zero returns (portfolio = 0)
            transformed = makeMeasures();
            addBacon(transformed, 0, baconPortfolioLen, new Array(baconPortfolioLen).fill(0));
            assertFloatEqual(transformed.cdarBeta(), 0.0, { places: 15, prefix: 'CDaR beta (geometric) zero returns' });

            for (const confidence of [0.0, 1.0]) {
                assertThrows(() => measures.cdarBeta(confidence));
            }
        });
    });

    describe('CDaRAlpha', () => {
        it('matches PerformanceAnalytics output', () => {
            // PerformanceAnalytics hardcodes a period of 12 when annualizing the
            // means; we use periodsPerAnnum, so monthly runs match the reference.
            for (const [p, expected] of rdCdarAlpha.EXPECTED_VALUES_BY_P_GEOMETRIC) {
                const actual = runStreamMethod('cdarAlpha', { monthly: true }, p);
                assertSeriesEqual(actual, expected, { places: 14, prefix: `CDaR alpha geometric (monthly) p ${p}` });
            }
        });

        it('mathematical properties', () => {
            // Identity (portfolio == benchmark).
            const measures = makeMeasures(0, 0, 0, false, true);
            addBacon(measures, 0, baconPortfolioLen, baconPortfolioReturns, baconPortfolioReturns);
            assertFloatEqual(measures.cdarAlpha(), 0.0, { places: 14, prefix: 'CDaR alpha (geometric) identity' });

            // No drawdowns
            let transformed = makeMeasures();
            addBacon(transformed, 0, baconPortfolioLen,
                new Array(baconPortfolioLen).fill(0.01), new Array(baconPortfolioLen).fill(0.02));
            assertFloatEqual(transformed.cdarAlpha(), NaN, { prefix: 'CDaR alpha (geometric) no drawdowns' });

            // Zero returns (portfolio = 0)
            transformed = makeMeasures();
            addBacon(transformed, 0, baconPortfolioLen, new Array(baconPortfolioLen).fill(0));
            assertFloatEqual(transformed.cdarAlpha(), 0.0, { places: 15, prefix: 'CDaR alpha (geometric) zero returns' });
        });
    });

    describe('RewardToConditionalDrawdown', () => {
        it('definition', () => {
            // Geometric mean return divided by the mean magnitude of the worst
            // max(1, int(n * (1 - confidence))) drawdowns.
            for (const confidence of [0.8, 0.95]) {
                const m = makeMeasures();
                for (let i = 0; i < baconPortfolioLen; i++) {
                    m.addReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i]);
                    const dd = m.drawdownsHighWatermark.sort((a, b) => a - b);
                    const tail = dd.slice(0, Math.max(1, Math.trunc(dd.length * (1 - confidence))));
                    const cdar = -pySum(tail) / tail.length;
                    const expected = cdar !== 0 ? m.geometricMeanReturn / cdar : NaN;
                    assertFloatEqual(m.rewardToConditionalDrawdown(confidence), expected,
                        { places: 15, prefix: `confidence ${confidence} step ${i}` });
                }
            }
        });
    });

    describe('SfmRiskPremium', () => {
        it('matches PerformanceAnalytics output', () => {
            checkYearlyAndDailyRf('sfmRiskPremium', rdSfmRiskPremium.EXPECTED_VALUES_BY_RF_PERFAN, 14, 'SFM risk premium');
        });
    });

    describe('SfmAlpha', () => {
        it('matches PerformanceAnalytics output', () => {
            checkYearlyAndDailyRf('sfmAlpha', rdSfmAlpha.EXPECTED_VALUES_BY_RF_PERFAN, 14, 'SFM alpha');
        });
    });

    describe('SfmBeta', () => {
        it('matches PerformanceAnalytics output', () => {
            checkYearlyAndDailyRf('sfmBeta', rdSfmBeta.EXPECTED_VALUES_BY_RF_PERFAN, 14, 'SFM beta');
        });
    });

    describe('SfmBetaBull', () => {
        it('matches PerformanceAnalytics output', () => {
            checkYearlyAndDailyRf('sfmBetaBull', rdSfmBetaBull.EXPECTED_VALUES_BY_RF_PERFAN, 14, 'SFM beta bull');
        });
    });

    describe('SfmBetaBear', () => {
        it('matches reference implementation output', () => {
            checkYearlyAndDailyRf('sfmBetaBear', rdSfmBetaBear.EXPECTED_VALUES_BY_RF_REFERENCE, 14, 'SFM beta bear');
        });
    });

    describe('TimingRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            checkYearlyAndDailyRf('timingRatio', rdTimingRatio.EXPECTED_VALUES_BY_RF_PERFAN, 14, 'timing ratio');
        });
    });

    describe('SfmR2', () => {
        it('matches PerformanceAnalytics output', () => {
            checkYearlyAndDailyRf('sfmR2', rdSfmR2.EXPECTED_VALUES_BY_RF_PERFAN, 14, 'SFM R^2',
                r => r < 0.05 ? 15 : 18);
        });
    });

    describe('JensenAlpha', () => {
        it('high daily risk-free rate definition', () => {
            // The R fixtures lose precision after compounding a 10% or 30%
            // periodic rate over 252 periods. Check the documented formula.
            for (const r of [0.1, 0.3]) {
                const annualRf = annualize(r, 252);
                const m = new Measures(252, annualRf);
                for (let i = 0; i < baconPortfolioLen; i++) {
                    m.addReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i]);
                    if (i === 0) {
                        continue;
                    }
                    const pAnn = prod(baconPortfolioReturns.slice(0, i + 1).map(x => 1 + x)) ** (252 / (i + 1)) - 1;
                    const bAnn = prod(baconBenchmarkReturns.slice(0, i + 1).map(x => 1 + x)) ** (252 / (i + 1)) - 1;
                    const expected = pAnn - (m.sfmBeta * bAnn + (1 - m.sfmBeta) * annualRf);
                    assertFloatEqual(m.jensenAlpha, expected,
                        { delta: 1e-9, relTol: 1e-12, prefix: `Jensen alpha rf=${r} step=${i}` });
                    if (m.sfmBeta !== 0) {
                        assertFloatEqual(m.jensenAlphaModified, expected / m.sfmBeta,
                            { delta: 1e-9, relTol: 1e-12, prefix: `Jensen alpha modified rf=${r} step=${i}` });
                    }
                }
            }
        });

        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdJensenAlpha.EXPECTED_VALUES_BY_RF_DAILY_PERFAN) {
                const actual = runStreamProperty('jensenAlpha', rf(annualize(r, 252), { daily: true }));
                if (r <= 0.05) {
                    // At higher periodic rates the R reference loses precision
                    // through cancellation; the formula test above covers them.
                    assertSeriesEqual(actual, expected,
                        { delta: 9e-10, relTol: 1e-10, prefix: `Jensen alpha (daily, Rf ${r})` });
                }
            }
            for (const [r, expected] of rdJensenAlpha.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN) {
                const actual = runStreamProperty('jensenAlpha', rf(annualize(r, 12), { monthly: true }));
                assertSeriesEqual(actual, expected, { places: 12, prefix: `Jensen alpha (monthly, Rf ${r})` });
            }
            for (const [r, expected] of rdJensenAlpha.EXPECTED_VALUES_BY_RF_YEARLY_PERFAN) {
                const actual = runStreamProperty('jensenAlpha', rf(r));
                assertSeriesEqual(actual, expected, { places: 14, prefix: `Jensen alpha (yearly, Rf ${r})` });
            }
        });
    });

    describe('FamaBeta', () => {
        it('matches PerformanceAnalytics output', () => {
            const expected = rdFamaBeta.EXPECTED_VALUES_PERFAN;
            let actual = runStreamProperty('famaBeta');
            assertSeriesEqual(actual, expected, { places: 14, prefix: 'fama beta (yearly)' });
            actual = runStreamProperty('famaBeta', { daily: true });
            assertSeriesEqual(actual, expected, { places: 14, prefix: 'fama beta (daily)' });
        });
    });

    describe('Modigliani', () => {
        it('matches PerformanceAnalytics output', () => {
            checkYearlyAndDailyRf('modigliani', rdModigliani.EXPECTED_VALUES_BY_RF_PERFAN, 15, 'Modigliani-Modigliani');
        });
    });

    describe('TrackingError', () => {
        it('matches PerformanceAnalytics output', () => {
            let actual = runStreamProperty('trackingError', { daily: true });
            assertSeriesEqual(actual, rdTrackingError.EXPECTED_VALUES_DAILY_PERFAN, { places: 15, prefix: 'tracking error (daily)' });
            actual = runStreamProperty('trackingError', { monthly: true });
            assertSeriesEqual(actual, rdTrackingError.EXPECTED_VALUES_MONTHLY_PERFAN, { places: 15, prefix: 'tracking error (monthly)' });
            actual = runStreamProperty('trackingError');
            assertSeriesEqual(actual, rdTrackingError.EXPECTED_VALUES_ANNUAL_PERFAN, { places: 15, prefix: 'tracking error (yearly)' });
        });
    });

    describe('ActivePremium', () => {
        it('matches PerformanceAnalytics output', () => {
            // We skip the first element because PerformanceAnalytics uses it
            // to determine periodicity.
            let actual = runStreamProperty('activePremium', { daily: true });
            assertSeriesEqual(actual, rdActivePremium.EXPECTED_VALUES_DAILY_PERFAN,
                { places: 11, skip: 1, prefix: 'active premium (daily)' });
            actual = runStreamProperty('activePremium', { monthly: true });
            assertSeriesEqual(actual, rdActivePremium.EXPECTED_VALUES_MONTHLY_PERFAN,
                { places: 14, skip: 1, prefix: 'active premium (monthly)' });
            actual = runStreamProperty('activePremium');
            assertSeriesEqual(actual, rdActivePremium.EXPECTED_VALUES_ANNUAL_PERFAN,
                { places: 15, skip: 1, prefix: 'active premium (yearly)' });
        });
    });

    describe('InformationRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            let actual = runStreamProperty('informationRatio', { daily: true });
            assertSeriesEqual(actual, rdInformationRatio.EXPECTED_VALUES_DAILY_PERFAN,
                { places: 10, skip: 2, prefix: 'information ratio (daily)' });
            actual = runStreamProperty('informationRatio', { monthly: true });
            assertSeriesEqual(actual, rdInformationRatio.EXPECTED_VALUES_MONTHLY_PERFAN,
                { places: 12, skip: 2, prefix: 'information ratio (monthly)' });
            actual = runStreamProperty('informationRatio');
            assertSeriesEqual(actual, rdInformationRatio.EXPECTED_VALUES_ANNUAL_PERFAN,
                { places: 13, skip: 2, prefix: 'information ratio (yearly)' });
        });
    });

    describe('InformationRatioModified', () => {
        it('sign rule', () => {
            // Equals informationRatio when the geometric active premium is positive,
            // otherwise its negation.
            const m = makeMeasures(0, 0, 0, false, true);
            for (let i = 0; i < baconPortfolioLen; i++) {
                m.addReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i]);
                const active = m.activePremium;
                const ir = m.informationRatio;
                const expected = Number.isNaN(ir) ? NaN : (active > 0 ? ir : -ir);
                assertFloatEqual(m.informationRatioModified, expected, { places: 15, prefix: `step ${i}` });
            }
        });
    });

    /** Daily, monthly and yearly runs of the annualized SFM measures. */
    function checkDailyMonthlyAnnual(name: string, label: string,
        daily: Series, dailyPlaces: number, monthly: Series, monthlyPlaces: number,
        annual: Series, annualPlaces: number): void {
        for (const [r, expected] of daily) {
            const actual = runStreamProperty(name, rf(annualize(r, 252), { daily: true }));
            assertSeriesEqual(actual, expected, { places: dailyPlaces, prefix: `${label} (daily, Rf ${r})` });
        }
        for (const [r, expected] of monthly) {
            const actual = runStreamProperty(name, rf(annualize(r, 12), { monthly: true }));
            assertSeriesEqual(actual, expected, { places: monthlyPlaces, prefix: `${label} (monthly, Rf ${r})` });
        }
        for (const [r, expected] of annual) {
            const actual = runStreamProperty(name, rf(r));
            assertSeriesEqual(actual, expected, { places: annualPlaces, prefix: `${label} (yearly, Rf ${r})` });
        }
    }

    describe('SystematicRisk', () => {
        it('matches PerformanceAnalytics output', () => {
            checkDailyMonthlyAnnual('systematicRisk', 'systematic risk',
                rdSystematicRisk.EXPECTED_VALUES_BY_RF_DAILY_PERFAN, 14,
                rdSystematicRisk.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN, 15,
                rdSystematicRisk.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN, 15);
        });
    });

    describe('TreynorRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            checkDailyMonthlyAnnual('treynorRatio', 'treynor ratio',
                rdTreynorRatio.EXPECTED_VALUES_BY_RF_DAILY_PERFAN, 10,
                rdTreynorRatio.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN, 13,
                rdTreynorRatio.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN, 14);
        });
    });

    describe('TreynorRatioModified', () => {
        it('matches PerformanceAnalytics output', () => {
            checkDailyMonthlyAnnual('treynorRatioModified', 'treynor ratio modified',
                rdTreynorRatioModified.EXPECTED_VALUES_BY_RF_DAILY_PERFAN, 10,
                rdTreynorRatioModified.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN, 12,
                rdTreynorRatioModified.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN, 12);
        });
    });

    describe('SpecificRisk', () => {
        it('matches PerformanceAnalytics output', () => {
            checkDailyMonthlyAnnual('specificRisk', 'specific risk',
                rdSpecificRisk.EXPECTED_VALUES_BY_RF_DAILY_PERFAN, 14,
                rdSpecificRisk.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN, 15,
                rdSpecificRisk.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN, 15);
        });
    });

    describe('TotalRisk', () => {
        it('matches PerformanceAnalytics output', () => {
            checkDailyMonthlyAnnual('totalRisk', 'total_risk',
                rdTotalRisk.EXPECTED_VALUES_BY_RF_DAILY_PERFAN, 14,
                rdTotalRisk.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN, 14,
                rdTotalRisk.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN, 15);
        });
    });

    describe('AppraisalRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdAppraisalRatio.EXPECTED_VALUES_BY_RF_DAILY_PERFAN) {
                const actual = runStreamProperty('appraisalRatio', rf(annualize(r, 252), { daily: true }));
                if (r < 0.1) {
                    assertSeriesEqual(actual, expected, { delta: r < 0.05 ? 1e-8 : 1e-4, skip: 2,
                        prefix: `appraisal ratio (daily, Rf ${r})` });
                }
            }
            for (const [r, expected] of rdAppraisalRatio.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN) {
                const actual = runStreamProperty('appraisalRatio', rf(annualize(r, 12), { monthly: true }));
                assertSeriesEqual(actual, expected, { delta: r < 0.05 ? 1e-11 : 1e-9, skip: 2,
                    prefix: `appraisal ratio (monthly, Rf ${r})` });
            }
            for (const [r, expected] of rdAppraisalRatio.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN) {
                const actual = runStreamProperty('appraisalRatio', rf(r));
                assertSeriesEqual(actual, expected, { delta: r < 0.05 ? 1e-13 : 1e-11, skip: 2,
                    prefix: `appraisal ratio (yearly, Rf ${r})` });
            }
        });
    });

    describe('JensenAlphaModified', () => {
        it('definition', () => {
            const m = makeMeasures();
            for (let i = 0; i < baconPortfolioLen; i++) {
                m.addReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i]);
                const beta = m.sfmBeta;
                const expected = beta !== 0 ? m.jensenAlpha / beta : NaN;
                assertFloatEqual(m.jensenAlphaModified, expected, { places: 14, prefix: `Jensen alpha modified n=${i + 1}` });
            }
        });

        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdJensenAlphaModified.EXPECTED_VALUES_BY_RF_DAILY_PERFAN) {
                const actual = runStreamProperty('jensenAlphaModified', rf(annualize(r, 252), { daily: true }));
                if (r < 0.05) {
                    // Annualizing 24 daily observations amplifies the reference's
                    // floating-point error, even when the formula agrees.
                    assertSeriesEqual(actual, expected, { delta: 1e-8, prefix: `Jensen alpha modified (daily, Rf ${r})` });
                }
            }
            for (const [r, expected] of rdJensenAlphaModified.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN) {
                const actual = runStreamProperty('jensenAlphaModified', rf(annualize(r, 12), { monthly: true }));
                assertSeriesEqual(actual, expected, { delta: 1e-10, prefix: `Jensen alpha modified (monthly, Rf ${r})` });
            }
            for (const [r, expected] of rdJensenAlphaModified.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN) {
                const actual = runStreamProperty('jensenAlphaModified', rf(r));
                assertSeriesEqual(actual, expected, { delta: 1e-13, prefix: `Jensen alpha modified (yearly, Rf ${r})` });
            }
        });
    });

    describe('JensenAlphaAlternative', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdJensenAlphaAlternative.EXPECTED_VALUES_BY_RF_DAILY_PERFAN) {
                const actual = runStreamProperty('jensenAlphaAlternative', rf(annualize(r, 252), { daily: true }));
                if (r < 0.3) {
                    assertSeriesEqual(actual, expected, { delta: r < 0.1 ? 0.1826 : 0.707,
                        prefix: `Jensen alpha alternative (daily, Rf ${r})` });
                }
            }
            for (const [r, expected] of rdJensenAlphaAlternative.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN) {
                const actual = runStreamProperty('jensenAlphaAlternative', rf(annualize(r, 12), { monthly: true }));
                assertSeriesEqual(actual, expected, { places: r < 0.3 ? 10 : 9,
                    prefix: `Jensen alpha alternative (monthly, Rf ${r})` });
            }
            for (const [r, expected] of rdJensenAlphaAlternative.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN) {
                const actual = runStreamProperty('jensenAlphaAlternative', rf(r));
                assertSeriesEqual(actual, expected, { places: 12, prefix: `Jensen alpha alternative (yearly, Rf ${r})` });
            }
        });
    });

    describe('MSquared', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdMSquared.EXPECTED_VALUES_BY_RF_DAILY_PERFAN) {
                const actual = runStreamProperty('mSquared', rf(annualize(r, 252), { daily: true }));
                if (r < 0.05) {
                    assertSeriesEqual(actual, expected, { delta: r < 0.01 ? 0.1849 : 0.82956,
                        prefix: `M squared (daily, Rf ${r})` });
                }
            }
            for (const [r, expected] of rdMSquared.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN) {
                const actual = runStreamProperty('mSquared', rf(annualize(r, 12), { monthly: true }));
                assertSeriesEqual(actual, expected, { delta: r < 0.05 ? 0.00861 : 1.621,
                    prefix: `M squared (monthly, Rf ${r})` });
            }
            for (const [r, expected] of rdMSquared.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN) {
                const actual = runStreamProperty('mSquared', rf(r));
                assertSeriesEqual(actual, expected, { places: 14, prefix: `M squared (yearly, Rf ${r})` });
            }
        });
    });

    describe('MSquaredExcess', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [r, expected] of rdMSquaredExcess.EXPECTED_VALUES_BY_RF_DAILY_PERFAN) {
                const actual = runStreamProperty('mSquaredExcess', rf(annualize(r, 252), { daily: true }));
                if (r < 0.05) {
                    assertSeriesEqual(actual, expected, { delta: r < 0.01 ? 0.02244 : 0.101,
                        prefix: `M squared excess (daily, Rf ${r})` });
                }
            }
            for (const [r, expected] of rdMSquaredExcess.EXPECTED_VALUES_BY_RF_MONTHLY_PERFAN) {
                const actual = runStreamProperty('mSquaredExcess', rf(annualize(r, 12), { monthly: true }));
                assertSeriesEqual(actual, expected, { delta: r < 0.05 ? 0.007782 : 1.466,
                    prefix: `M squared excess (monthly, Rf ${r})` });
            }
            for (const [r, expected] of rdMSquaredExcess.EXPECTED_VALUES_BY_RF_ANNUAL_PERFAN) {
                const actual = runStreamProperty('mSquaredExcess', rf(r));
                assertSeriesEqual(actual, expected, { places: 15, prefix: `M squared excess (yearly, Rf ${r})` });
            }
        });
    });

    describe('MSquaredSortino', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [m, expected] of rdMSquaredSortino.EXPECTED_VALUES_BY_MAR_DAILY_PERFAN) {
                const actual = runStreamProperty('mSquaredSortino', mar(annualize(m, 252), { daily: true }));
                assertSeriesEqual(actual, expected, { places: 11, skip: 3, prefix: `M squared Sortino (daily, MAR ${m})` });
            }
            for (const [m, expected] of rdMSquaredSortino.EXPECTED_VALUES_BY_MAR_MONTHLY_PERFAN) {
                const actual = runStreamProperty('mSquaredSortino', mar(annualize(m, 12), { monthly: true }));
                assertSeriesEqual(actual, expected, { places: 14, skip: 3, prefix: `M squared Sortino (monthly, MAR ${m})` });
            }
            for (const [m, expected] of rdMSquaredSortino.EXPECTED_VALUES_BY_MAR_ANNUAL_PERFAN) {
                const actual = runStreamProperty('mSquaredSortino', mar(m));
                assertSeriesEqual(actual, expected, { places: 15, skip: 3, prefix: `M squared Sortino (yearly, MAR ${m})` });
            }
        });
    });

    describe('TailRatio', () => {
        it('matches reference implementation output', () => {
            for (const [cutoff, expected] of rdTailRatio.EXPECTED_VALUES_BY_CUTOFF_REFERENCE) {
                const actual = runStreamMethod('tailRatio', {}, cutoff);
                assertSeriesEqual(actual, expected, { places: 15, prefix: `tail ratio (yearly, cutoff ${cutoff})` });
            }
        });
    });

    function checkKelly(name: string, data: Series, label: string): void {
        for (const [r, expected] of data) {
            const actual = runStreamProperty(name, rf(annualize(r, 252), { daily: true }));
            assertSeriesEqual(actual, expected, { places: 11, prefix: `${label} (daily, Rf ${r})` });
        }
        for (const [r, expected] of data) {
            const actual = runStreamProperty(name, rf(annualize(r, 12), { monthly: true }));
            assertSeriesEqual(actual, expected, { places: 11, prefix: `${label} (monthly, Rf ${r})` });
        }
        for (const [r, expected] of data) {
            const actual = runStreamProperty(name, rf(r));
            assertSeriesEqual(actual, expected, { places: 11, prefix: `${label} (yearly, Rf ${r})` });
        }
    }

    describe('KellyRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            checkKelly('kellyRatio', rdKellyRatio.EXPECTED_VALUES_BY_RF_PERFAN, 'Kelly ratio');
        });
    });

    describe('KellyRatioFull', () => {
        it('matches PerformanceAnalytics output', () => {
            checkKelly('kellyRatioFull', rdKellyRatio.EXPECTED_VALUES_BY_RF_FULL_PERFAN, 'Kelly ratio full');
        });
    });

    describe('HurstExponent', () => {
        it('matches PerformanceAnalytics output', () => {
            const actual = runStreamProperty('hurstExponent');
            assertSeriesEqual(actual, rdHurstExponent.EXPECTED_VALUES_PERFAN, { places: 14, prefix: 'Hurst exponent' });
        });
    });

    describe('BiasRatio', () => {
        it('matches reference implementation output', () => {
            for (const [mult, expected] of rdBiasRatio.EXPECTED_VALUES_BY_MULT_REFERENCE) {
                const actual = runStreamMethod('biasRatio', {}, mult);
                assertSeriesEqual(actual, expected, { places: 15, prefix: `bias ratio (yearly, std_dev_multiplier ${mult})` });
            }
        });
    });

    describe('KRatio', () => {
        it('matches reference implementation output', () => {
            const actual = runStreamProperty('kRatio');
            assertSeriesEqual(actual, rdKRatio.EXPECTED_VALUES_REFERENCE, { places: 14, prefix: 'K-ratio' });
        });
    });

    describe('GainToPainRatio', () => {
        it('matches reference implementation output', () => {
            const actual = runStreamProperty('gainToPainRatio');
            assertSeriesEqual(actual, rdGainToPainRatio.EXPECTED_VALUES_REFERENCE, { places: 15, prefix: 'Gain-to-pain ratio' });
        });
    });

    describe('UpsideCaptureRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [geom, expected] of rdUpsideCaptureRatio.EXPECTED_VALUES_BY_GEOMETRIC_PERFAN) {
                const actual = runStreamMethod('upsideCaptureRatio', {}, geom);
                assertSeriesEqual(actual, expected, { places: 13, skip: 1, prefix: `Upside capture ratio (yearly, geometric ${geom})` });
            }
        });
    });

    describe('DownsideCaptureRatio', () => {
        it('matches PerformanceAnalytics output', () => {
            for (const [geom, expected] of rdDownsideCaptureRatio.EXPECTED_VALUES_BY_GEOMETRIC_PERFAN) {
                const actual = runStreamMethod('downsideCaptureRatio', {}, geom);
                assertSeriesEqual(actual, expected, { places: 14, prefix: `Downside capture ratio (yearly, geometric ${geom})` });
            }
        });
    });

    describe('OverallCaptureRatio', () => {
        it('matches reference implementation output', () => {
            for (const [geom, expected] of rdOverallCaptureRatio.EXPECTED_VALUES_BY_GEOMETRIC_REFERENCE) {
                const actual = runStreamMethod('overallCaptureRatio', {}, geom);
                assertSeriesEqual(actual, expected, { places: 13, prefix: `Overall capture ratio (yearly, geometric ${geom})` });
            }
        });
    });

    const ratioCases: readonly [string, string, readonly number[], string][] = [
        ['UpNumberRatio', 'upNumberRatio', rdUpNumberRatio.EXPECTED_VALUES_PERFAN, 'Up number ratio (yearly)'],
        ['DownNumberRatio', 'downNumberRatio', rdDownNumberRatio.EXPECTED_VALUES_PERFAN, 'Down number ratio (yearly)'],
        ['UpPercentageRatio', 'upPercentageRatio', rdUpPercentageRatio.EXPECTED_VALUES_PERFAN, 'Up percentage ratio (yearly)'],
        ['DownPercentageRatio', 'downPercentageRatio', rdDownPercentageRatio.EXPECTED_VALUES_PERFAN, 'Down percentage ratio (yearly)'],
    ];
    for (const [suite, name, expected, prefix] of ratioCases) {
        describe(suite, () => {
            it('matches PerformanceAnalytics output', () => {
                assertSeriesEqual(runStreamProperty(name), expected, { places: 15, prefix });
            });
        });
    }

    describe('DocumentedFormulas', () => {
        it('drawdown risk and ratios', () => {
            for (const r of [0.0, 0.05]) {
                const m = makeMeasures(0, r);
                for (let i = 0; i < baconPortfolioLen; i++) {
                    m.addReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i]);
                    const drawdowns = m.drawdownsHighWatermark;
                    const n = i + 1;
                    const pain = -fsum(drawdowns) / n;
                    const ulcer = Math.sqrt(fsum(drawdowns.map(x => x * x)) / n);
                    const geometricReturn = prod(baconPortfolioReturns.slice(0, n).map(x => 1 + x)) ** (1 / n) - 1;
                    assertAlmostEqual(m.painIndex, pain, { places: 14 });
                    assertAlmostEqual(m.ulcerIndex, ulcer, { places: 14 });
                    if (pain > 0) {
                        assertAlmostEqual(m.painRatio, (geometricReturn - r) / pain, { places: 12 });
                    }
                    if (ulcer > 0) {
                        assertAlmostEqual(m.martinRatio, (geometricReturn - r) / ulcer, { places: 12 });
                    }
                }
            }
        });

        it('cdar average and alpha', () => {
            const m = makeMeasures(0, 0, 0, false, true);
            for (let k = 0; k < baconPortfolioLen; k++) {
                m.addReturn(baconPortfolioReturns[k], baconBenchmarkReturns[k]);
                const drawdowns = m.drawdownsHighWatermark.sort((a, b) => a - b);
                for (const confidence of [0.9, 0.95]) {
                    const position = (1 - confidence) * (drawdowns.length - 1);
                    const lo = Math.trunc(position);
                    const q = drawdowns[lo] + (position - lo) * (
                        drawdowns[Math.min(lo + 1, drawdowns.length - 1)] - drawdowns[lo]);
                    const tail = drawdowns.filter(d => d <= q);
                    const expectedCdar = q < 0 ? -fsum(tail) / tail.length : 0.0;
                    assertAlmostEqual(m.cdarAverage(confidence), expectedCdar, { places: 14 });
                    const beta = m.cdarBeta(confidence);
                    if (Number.isFinite(beta)) {
                        const portfolioMean = fmean(baconPortfolioReturns.slice(0, drawdowns.length));
                        const benchmarkMean = fmean(baconBenchmarkReturns.slice(0, drawdowns.length));
                        const expectedAlpha = (1 + portfolioMean) ** 12 - 1
                            - beta * ((1 + benchmarkMean) ** 12 - 1);
                        assertAlmostEqual(m.cdarAlpha(confidence), expectedAlpha, { places: 13 });
                    }
                }
            }
        });

        it('m squared and Jensen alpha alternative', () => {
            for (const [periods, periodicRf] of [[1, 0.05], [12, 0.05], [252, 0.01]]) {
                const annualRf = annualize(periodicRf, periods);
                const m = new Measures(periods, annualRf);
                for (let i = 0; i < baconPortfolioLen; i++) {
                    m.addReturn(baconPortfolioReturns[i], baconBenchmarkReturns[i]);
                    const n = i + 1;
                    if (n < 2) {
                        continue;
                    }
                    const portfolio = baconPortfolioReturns.slice(0, n);
                    const benchmark = baconBenchmarkReturns.slice(0, n);
                    const pAnn = prod(portfolio.map(x => 1 + x)) ** (periods / n) - 1;
                    const bAnn = prod(benchmark.map(x => 1 + x)) ** (periods / n) - 1;
                    const scale = pstdev(benchmark) / pstdev(portfolio);
                    const expectedM2 = pAnn * scale + annualRf * (1 - scale);
                    assertFloatEqual(m.mSquared, expectedM2,
                        { delta: 1e-10, relTol: 1e-11, prefix: `M squared periods=${periods} n=${n}` });
                    const expectedExcess = (1 + expectedM2) / (1 + bAnn) - 1;
                    assertFloatEqual(m.mSquaredExcess, expectedExcess,
                        { delta: 1e-10, relTol: 1e-11, prefix: `M squared excess periods=${periods} n=${n}` });
                    const systematicRisk = Math.abs(m.sfmBeta) * stdev(benchmark) * Math.sqrt(periods);
                    if (systematicRisk > 0 && Number.isFinite(systematicRisk)) {
                        assertFloatEqual(m.jensenAlphaAlternative, m.jensenAlpha / systematicRisk,
                            { delta: 1e-10, relTol: 1e-11, prefix: `Jensen alpha alternative periods=${periods} n=${n}` });
                    }
                }
            }
        });

        it('m squared equal volatility with extreme rate', () => {
            const m = new Measures(252, 1.3 ** 252 - 1);
            for (const [portfolio, benchmark] of [[0.125, 0.25], [0.375, 0.5]]) {
                m.addReturn(portfolio, benchmark);
            }
            // Binary-exact inputs give exactly equal portfolio and benchmark
            // volatility, so the annual risk-free terms must cancel.
            const expected = ((1.125 * 1.375) ** 126) - 1;
            assertTrue(isClose(m.mSquared, expected, 1e-14));
        });
    });

    describe('RollingWindow', () => {
        it('rolling matches fresh', () => {
            // At every step, including while the window is still filling, every
            // public getter and method (with default arguments) of a
            // rolling-window instance equals that of a fresh instance fed only
            // the returns in the window.
            const rng = new SeededRandom(42);
            const randomReturns = Array.from({ length: 150 }, () => rng.gauss(0.002, 0.03));
            const randomBenchmark = Array.from({ length: 150 }, () => rng.gauss(0.001, 0.025));
            const configs = [
                { window: 10, periodsPerAnnum: 1, annualRf: 0.0, annualMar: 0.0,
                    returns: baconPortfolioReturns, benchmark: baconBenchmarkReturns },
                { window: 30, periodsPerAnnum: 12, annualRf: 0.05, annualMar: 0.03,
                    returns: randomReturns, benchmark: randomBenchmark },
            ];
            const names = publicMeasures();
            for (const cfg of configs) {
                const window = cfg.window;
                const create = () => new Measures(cfg.periodsPerAnnum, cfg.annualRf, cfg.annualMar);
                const rolling = new Measures(cfg.periodsPerAnnum, cfg.annualRf, cfg.annualMar, window);
                const { returns, benchmark } = cfg;
                for (let i = 0; i < returns.length; i++) {
                    rolling.addReturn(returns[i], benchmark[i]);
                    const fresh = create();
                    for (let j = Math.max(0, i - window + 1); j < i + 1; j++) {
                        fresh.addReturn(returns[j], benchmark[j]);
                    }
                    for (const name of names) {
                        const actual = evaluate(rolling, name);
                        const expected = evaluate(fresh, name);
                        const prefix = `window ${window} step ${i} ${name}`;
                        if (Array.isArray(expected)) {
                            assertSeriesEqual(actual as number[], expected, { places: 12, prefix });
                            assertEqual((actual as number[]).length, expected.length, prefix);
                        } else if (typeof expected === 'boolean') {
                            assertEqual(actual, expected, prefix);
                        } else {
                            const e = expected as number;
                            const delta = Number.isFinite(e) ? 1e-12 * Math.max(1.0, Math.abs(e)) : undefined;
                            assertFloatEqual(actual as number, e, { delta, prefix });
                        }
                    }
                }
            }
        });
    });
});
