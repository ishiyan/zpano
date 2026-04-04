import { } from 'jasmine';

import { Ratios } from './ratios';
import { Periodicity } from './periodicity';
import { DayCountConvention } from '../daycounting';

// Test data from 'Portfolio bacon' dataset from PerformanceAnalytics R package
const baconDatesPrevious = [
    new Date(2024, 5, 30), new Date(2024, 6, 1), new Date(2024, 6, 2), new Date(2024, 6, 3),
    new Date(2024, 6, 4), new Date(2024, 6, 5), new Date(2024, 6, 6), new Date(2024, 6, 7),
    new Date(2024, 6, 8), new Date(2024, 6, 9), new Date(2024, 6, 10), new Date(2024, 6, 11),
    new Date(2024, 6, 12), new Date(2024, 6, 13), new Date(2024, 6, 14), new Date(2024, 6, 15),
    new Date(2024, 6, 16), new Date(2024, 6, 17), new Date(2024, 6, 18), new Date(2024, 6, 19),
    new Date(2024, 6, 20), new Date(2024, 6, 21), new Date(2024, 6, 22), new Date(2024, 6, 23)
];

const baconDates = [
    new Date(2024, 6, 1), new Date(2024, 6, 2), new Date(2024, 6, 3), new Date(2024, 6, 4),
    new Date(2024, 6, 5), new Date(2024, 6, 6), new Date(2024, 6, 7), new Date(2024, 6, 8),
    new Date(2024, 6, 9), new Date(2024, 6, 10), new Date(2024, 6, 11), new Date(2024, 6, 12),
    new Date(2024, 6, 13), new Date(2024, 6, 14), new Date(2024, 6, 15), new Date(2024, 6, 16),
    new Date(2024, 6, 17), new Date(2024, 6, 18), new Date(2024, 6, 19), new Date(2024, 6, 20),
    new Date(2024, 6, 21), new Date(2024, 6, 22), new Date(2024, 6, 23), new Date(2024, 6, 24)
];

const baconPortfolioReturns = [
    0.003, 0.026, 0.011, -0.010,
    0.015, 0.025, 0.016, 0.067,
    -0.014, 0.040, -0.005, 0.081,
    0.040, -0.037, -0.061, 0.017,
    -0.049, -0.022, 0.070, 0.058,
    -0.065, 0.024, -0.005, -0.009
];

const baconBenchmarkReturns = [
    0.002, 0.025, 0.018, -0.011,
    0.014, 0.018, 0.014, 0.065,
    -0.015, 0.042, -0.006, 0.083,
    0.039, -0.038, -0.062, 0.015,
    -0.048, 0.021, 0.060, 0.056,
    -0.067, 0.019, -0.003, 0.000
];

const baconPortfolioLen = baconPortfolioReturns.length;

const SQRT2 = 1.4142135623730950488016887242097;

describe('Kurtosis', () => {
    const expectedKurtosisExcess = [
        null, -2.00000000000000000, -1.50000000000000000,
        -1.17592035552795000, -0.94669079980875600, -0.96028723389787100,
        -0.57793300076120100, 0.78641242115027200, 0.59954237086621500,
        -0.01187577489273160, 0.07517391430462480, -0.27406990671095100,
        -0.38022416153835900, -0.31560370425738600, -0.16235155227201600,
        0.02528905226985100, -0.33285099821964000, -0.37425348407483000,
        -0.58502674157514900, -0.69334606360953100, -0.77381631285861200,
        -0.68208349704651200, -0.61779722177118000, -0.56754620589212500
    ];

    it('should calculate kurtosis matching R PerformanceAnalytics', () => {
        const ratios = new Ratios(
            Periodicity.DAILY,
            0,
            0,
            DayCountConvention.RAW
        );
        ratios.reset();

        for (let i = 0; i < baconPortfolioLen; i++) {
            ratios.addReturn(
                baconPortfolioReturns[i],
                baconBenchmarkReturns[i],
                1,
                baconDatesPrevious[i],
                baconDates[i]
            );
            const expected = expectedKurtosisExcess[i];
            if (expected === null) {
                expect(ratios.kurtosis).toBeNull();
            } else {
                expect(ratios.kurtosis).toBeCloseTo(expected, 13);
            }
        }
    });
});

describe('Sharpe Ratio', () => {
    const expectedStdDev: Record<number, (number | null)[]> = {
        0: [
            null, 0.8915694197569510, 1.1419253390798400,
            0.4977924836999790, 0.6680426571226850, 0.8511810078441020,
            0.9735918376312110, 0.8462916062735410, 0.6475912629068400,
            0.7524743687246650, 0.6702597534059590, 0.7244562693337180,
            0.7945207458232130, 0.5805910371128360, 0.3566360956461000,
            0.3758075293232440, 0.2578994439571370, 0.2131725662300710,
            0.2880753096781920, 0.3448210835211740, 0.2337747541463060,
            0.2546053055676570, 0.2430648040410730, 0.2275684556623890
        ],
        0.05: [
            null, -2.1828078897497800, -3.1402946824695500,
            -2.8208240742998800, -3.0433054380033400, -2.7967375972020500,
            -2.9887005248213900, -1.3662354689514000, -1.4489272141296900,
            -1.3494093427967500, -1.4483773981646000, -0.9801467173338540,
            -0.9561181856516630, -0.9946559628057110, -1.0011155375243300,
            -1.0290804307636500, -1.0706734491553900, -1.1284729554976500,
            -0.9967676208113950, -0.9275814386971810, -0.9577955946576800,
            -0.9630722427994000, -0.9992664166133000, -1.0367007424619900
        ]
    };

    it('should calculate Sharpe ratio matching R (rf=0)', () => {
        const rf = 0;
        const ratios = new Ratios(
            Periodicity.DAILY,
            Math.pow(1 + rf, 252) - 1,
            0,
            DayCountConvention.RAW
        );
        ratios.reset();

        for (let i = 0; i < baconPortfolioLen; i++) {
            ratios.addReturn(
                baconPortfolioReturns[i],
                baconBenchmarkReturns[i],
                1,
                baconDatesPrevious[i],
                baconDates[i]
            );
            const expected = expectedStdDev[rf][i];
            const actual = ratios.sharpeRatio();
            if (expected === null) {
                expect(actual).toBeNull();
            } else {
                expect(actual).toBeCloseTo(expected, 13);
            }
        }
    });

    it('should calculate Sharpe ratio matching R (rf=0.05)', () => {
        const rf = 0.05;
        const ratios = new Ratios(
            Periodicity.DAILY,
            Math.pow(1 + rf, 252) - 1,
            0,
            DayCountConvention.RAW
        );
        ratios.reset();

        for (let i = 0; i < baconPortfolioLen; i++) {
            ratios.addReturn(
                baconPortfolioReturns[i],
                baconBenchmarkReturns[i],
                1,
                baconDatesPrevious[i],
                baconDates[i]
            );
            const expected = expectedStdDev[rf][i];
            const actual = ratios.sharpeRatio();
            if (expected === null) {
                expect(actual).toBeNull();
            } else {
                expect(actual).toBeCloseTo(expected, 13);
            }
        }
    });

    it('should ignore risk-free rate when requested', () => {
        for (const rf of [0, 0.05]) {
            const ratios = new Ratios(
                Periodicity.DAILY,
                Math.pow(1 + rf, 252) - 1,
                0,
                DayCountConvention.RAW
            );
            ratios.reset();

            for (let i = 0; i < baconPortfolioLen; i++) {
                ratios.addReturn(
                    baconPortfolioReturns[i],
                    baconBenchmarkReturns[i],
                    1,
                    baconDatesPrevious[i],
                    baconDates[i]
                );
                const expected = expectedStdDev[0][i];
                const actual = ratios.sharpeRatio(true);
                if (expected === null) {
                    expect(actual).toBeNull();
                } else {
                    expect(actual).toBeCloseTo(expected, 13);
                }
            }
        }
    });
});

describe('Sortino Ratio', () => {
    const expectedMar: Record<number, (number | null)[]> = {
        0: [
            null, null, null,
            1.5, 2.01246117974981, 2.85773803324704,
            3.25049446787935, 5.40936687607709, 2.69307029756515,
            3.29008543386979, 2.92819766175444, 4.10863007844407,
            4.56665101160337, 1.67730613630736, 0.691483512929973,
            0.727302390567925, 0.452770753672167, 0.370054264368203,
            0.536498400203865, 0.665303673385798, 0.401733515514418,
            0.438224836666163, 0.418857174247308, 0.392372028795065
        ],
        0.05: [
            -1, -0.951329033501053, -0.967821008377905,
            -0.955961761235827, -0.959422032420532, -0.950640505399932,
            -0.95521850710367, -0.835987494907806, -0.84620916319764,
            -0.825850705880606, -0.841892559996059, -0.739594446201381,
            -0.729168016460068, -0.735413445987151, -0.731283824494091,
            -0.739823509257131, -0.750430484501361, -0.766429130761335,
            -0.726278292165206, -0.700204514919608, -0.709303305303401,
            -0.71078905810419, -0.723223919287678, -0.735374254070636
        ]
    };

    it('should calculate Sortino ratio matching R (mar=0)', () => {
        const mar = 0;
        const marAnnual = Math.pow(1 + mar, 252) - 1;
        const ratios = new Ratios(
            Periodicity.DAILY,
            0,
            marAnnual,
            DayCountConvention.RAW
        );
        ratios.reset();

        for (let i = 0; i < baconPortfolioLen; i++) {
            ratios.addReturn(
                baconPortfolioReturns[i],
                baconBenchmarkReturns[i],
                1,
                baconDatesPrevious[i],
                baconDates[i]
            );
            const expected = expectedMar[mar][i];
            const actual = ratios.sortinoRatio();
            if (expected === null) {
                expect(actual).toBeNull();
            } else {
                expect(actual).toBeCloseTo(expected, 13);
            }
        }
    });

    it('should calculate Sortino ratio matching R (mar=0.05)', () => {
        const mar = 0.05;
        const marAnnual = Math.pow(1 + mar, 252) - 1;
        const ratios = new Ratios(
            Periodicity.DAILY,
            0,
            marAnnual,
            DayCountConvention.RAW
        );
        ratios.reset();

        for (let i = 0; i < baconPortfolioLen; i++) {
            ratios.addReturn(
                baconPortfolioReturns[i],
                baconBenchmarkReturns[i],
                1,
                baconDatesPrevious[i],
                baconDates[i]
            );
            const expected = expectedMar[mar][i];
            const actual = ratios.sortinoRatio();
            if (expected === null) {
                expect(actual).toBeNull();
            } else {
                expect(actual).toBeCloseTo(expected, 13);
            }
        }
    });

    it('should calculate Jack Schwager version with sqrt(2)', () => {
        const mar = 0;
        const marAnnual = Math.pow(1 + mar, 252) - 1;
        const ratios = new Ratios(
            Periodicity.DAILY,
            0,
            marAnnual,
            DayCountConvention.RAW
        );
        ratios.reset();

        for (let i = 0; i < baconPortfolioLen; i++) {
            ratios.addReturn(
                baconPortfolioReturns[i],
                baconBenchmarkReturns[i],
                1,
                baconDatesPrevious[i],
                baconDates[i]
            );
            const expected = expectedMar[mar][i];
            const actual = ratios.sortinoRatio(false, true);
            if (expected === null) {
                expect(actual).toBeNull();
            } else {
                const expectedDivided = expected / SQRT2;
                expect(actual).toBeCloseTo(expectedDivided, 13);
            }
        }
    });
});

describe('Omega Ratio', () => {
    const expectedLossThreshold: Record<number, (number | null)[]> = {
        0: [
            null, null, null,
            4.000000000000000, 5.500000000000000, 8.000000000000000,
            9.600000000000000, 16.300000000000000, 6.791666666666670,
            8.458333333333330, 7.000000000000000, 9.793103448275860,
            11.172413793103400, 4.909090909090910, 2.551181102362210,
            2.685039370078740, 1.937500000000000, 1.722222222222220,
            2.075757575757580, 2.368686868686870, 1.783269961977190,
            1.874524714828900, 1.839552238805970, 1.779783393501810
        ],
        0.02: [
            0.00000000000000000, 0.35294117647058800, 0.23076923076923100,
            0.10714285714285700, 0.09836065573770490, 0.18032786885245900,
            0.16923076923076900, 0.89230769230769200, 0.58585858585858600,
            0.78787878787878800, 0.62903225806451600, 1.12096774193548000,
            1.28225806451613000, 0.87845303867403300, 0.60687022900763400,
            0.60000000000000000, 0.47604790419161700, 0.42287234042553200,
            0.55585106382978700, 0.65691489361702100, 0.53579175704989200,
            0.54446854663774400, 0.51646090534979400, 0.48737864077669900
        ]
    };

    it('should calculate Omega ratio matching R (threshold=0)', () => {
        const l = 0;
        const lAnnual = Math.pow(1 + l, 252) - 1;
        const ratios = new Ratios(
            Periodicity.DAILY,
            0,
            lAnnual,
            DayCountConvention.RAW
        );
        ratios.reset();

        for (let i = 0; i < baconPortfolioLen; i++) {
            ratios.addReturn(
                baconPortfolioReturns[i],
                baconBenchmarkReturns[i],
                1,
                baconDatesPrevious[i],
                baconDates[i]
            );
            const expected = expectedLossThreshold[l][i];
            const actual = ratios.omegaRatio();
            if (expected === null) {
                expect(actual).toBeNull();
            } else {
                expect(actual).toBeCloseTo(expected, 13);
            }
        }
    });

    it('should calculate Omega ratio matching R (threshold=0.02)', () => {
        const l = 0.02;
        const lAnnual = Math.pow(1 + l, 252) - 1;
        const ratios = new Ratios(
            Periodicity.DAILY,
            0,
            lAnnual,
            DayCountConvention.RAW
        );
        ratios.reset();

        for (let i = 0; i < baconPortfolioLen; i++) {
            ratios.addReturn(
                baconPortfolioReturns[i],
                baconBenchmarkReturns[i],
                1,
                baconDatesPrevious[i],
                baconDates[i]
            );
            const expected = expectedLossThreshold[l][i];
            const actual = ratios.omegaRatio();
            if (expected === null) {
                expect(actual).toBeNull();
            } else {
                expect(actual).toBeCloseTo(expected, 13);
            }
        }
    });
});

describe('Kappa Ratio', () => {
    const expectedOrder1Mar0 = [
        null, null, null,
        3.0000000000000000, 4.5000000000000000, 7.0000000000000000,
        8.6000000000000000, 15.300000000000000, 5.7916666666666700,
        7.4583333333333300, 6.0000000000000000, 8.7931034482758600,
        10.172413793103400, 3.9090909090909090, 1.5511811023622000,
        1.6850393700787400, 0.9375000000000000, 0.7222222222222220,
        1.0757575757575800, 1.3686868686868700, 0.7832699619771860,
        0.8745247148288970, 0.8395522388059700, 0.7797833935018050
    ];

    it('should calculate Kappa ratio order 1 matching R', () => {
        const mar = 0;
        const marAnnual = Math.pow(1 + mar, 252) - 1;
        const ratios = new Ratios(
            Periodicity.DAILY,
            0,
            marAnnual,
            DayCountConvention.RAW
        );
        ratios.reset();

        for (let i = 0; i < baconPortfolioLen; i++) {
            ratios.addReturn(
                baconPortfolioReturns[i],
                baconBenchmarkReturns[i],
                1,
                baconDatesPrevious[i],
                baconDates[i]
            );
            const expected = expectedOrder1Mar0[i];
            const actual = ratios.kappaRatio(1);
            if (expected === null) {
                expect(actual).toBeNull();
            } else {
                expect(actual).toBeCloseTo(expected, 13);
            }
        }
    });
});

// ===== Rolling Window / Min Periods tests =====

function makeRatios(rollingWindow: number | null = null, minPeriods: number | null = null): Ratios {
    return new Ratios(Periodicity.DAILY, 0, 0, DayCountConvention.RAW, rollingWindow, minPeriods);
}

function addBacon(r: Ratios, i: number): void {
    r.addReturn(
        baconPortfolioReturns[i],
        baconBenchmarkReturns[i],
        1,
        baconDatesPrevious[i],
        baconDates[i]
    );
}

function addAllBacon(r: Ratios): void {
    for (let i = 0; i < 24; i++) {
        addBacon(r, i);
    }
}

describe('Min Periods', () => {
    it('should return null before threshold', () => {
        const r = makeRatios(null, 5);
        for (let i = 0; i < 4; i++) {
            addBacon(r, i);
            expect(r.sharpeRatio(true)).toBeNull();
            expect(r.sortinoRatio()).toBeNull();
        }
        addBacon(r, 4);
        expect(r.sharpeRatio(true)).not.toBeNull();
    });

    it('should treat zero minPeriods as null (always primed)', () => {
        const r = makeRatios(null, 0);
        addBacon(r, 0);
        addBacon(r, 1);
        expect(r.sharpeRatio(true)).not.toBeNull();
    });

    it('should treat negative minPeriods as null (always primed)', () => {
        const r = makeRatios(null, -1);
        addBacon(r, 0);
        addBacon(r, 1);
        expect(r.sharpeRatio(true)).not.toBeNull();
    });

    it('should match baseline with minPeriods=1', () => {
        const rBase = makeRatios(null, null);
        addAllBacon(rBase);

        const rMp = makeRatios(null, 1);
        addAllBacon(rMp);

        expect(rMp.sharpeRatio(true)).toBeCloseTo(rBase.sharpeRatio(true)!, 15);
        expect(rMp.omegaRatio()).toBeCloseTo(rBase.omegaRatio()!, 15);
    });
});

describe('Rolling Window', () => {
    it('should match fresh instance fed last N returns', () => {
        const rRolling = makeRatios(10, null);
        addAllBacon(rRolling);

        const rFresh = makeRatios(null, null);
        for (let i = 14; i < 24; i++) {
            addBacon(rFresh, i);
        }

        const eps = 13;
        expect(rRolling.sharpeRatio(true)).toBeCloseTo(rFresh.sharpeRatio(true)!, eps);
        expect(rRolling.omegaRatio()).toBeCloseTo(rFresh.omegaRatio()!, eps);
        expect(rRolling.skew).toBeCloseTo(rFresh.skew!, eps);
        expect(rRolling.kurtosis).toBeCloseTo(rFresh.kurtosis!, eps);
        expect(rRolling.painIndex()).toBeCloseTo(rFresh.painIndex()!, eps);
        expect(rRolling.ulcerIndex()).toBeCloseTo(rFresh.ulcerIndex()!, eps);
        expect(rRolling.riskOfRuin).toBeCloseTo(rFresh.riskOfRuin!, eps);
    });

    it('should behave like expanding when null', () => {
        const r1 = makeRatios(null, null);
        addAllBacon(r1);

        const r2 = new Ratios(Periodicity.DAILY, 0, 0, DayCountConvention.RAW);
        addAllBacon(r2);

        expect(r1.sharpeRatio(true)).toBeCloseTo(r2.sharpeRatio(true)!, 15);
    });

    it('should behave like expanding when window > data length', () => {
        const rLarge = makeRatios(100, null);
        addAllBacon(rLarge);

        const rBase = makeRatios(null, null);
        addAllBacon(rBase);

        expect(rLarge.sharpeRatio(true)).toBeCloseTo(rBase.sharpeRatio(true)!, 15);
    });
});

describe('Rolling Window with Min Periods', () => {
    it('should respect min_periods before window fills', () => {
        const r = makeRatios(10, 5);
        for (let i = 0; i < 4; i++) {
            addBacon(r, i);
            expect(r.sharpeRatio(true)).toBeNull();
        }
        addBacon(r, 4);
        expect(r.sharpeRatio(true)).not.toBeNull();

        addAllBacon(r);

        // After full dataset, window=10: should match fresh last 10
        // But we added 24 + 24 = 48 returns total. Let's make a clean test:
    });

    it('should match fresh after full dataset', () => {
        const r = makeRatios(10, 5);
        addAllBacon(r);

        const rFresh = makeRatios(null, null);
        for (let i = 14; i < 24; i++) {
            addBacon(rFresh, i);
        }

        expect(r.sharpeRatio(true)).toBeCloseTo(rFresh.sharpeRatio(true)!, 13);
    });

    it('should handle minPeriods > window', () => {
        const r = makeRatios(5, 10);
        for (let i = 0; i < 9; i++) {
            addBacon(r, i);
            expect(r.sharpeRatio(true)).toBeNull();
        }
        addBacon(r, 9);
        expect(r.sharpeRatio(true)).not.toBeNull();
    });
});
