import { RawMomentsKleinKbn } from '../../streaming-kbn';
import { normCdf } from './norm';
import { probabilisticSharpeRatio } from './probabilistic-sharpe-ratio';

describe('probabilisticSharpeRatio', () => {

    /** Compensated (Neumaier) sum, like Python's built-in sum of floats. */
    function pySum(xs: readonly number[]): number {
        let s = 0.0;
        let c = 0.0;
        for (const x of xs) {
            const t = s + x;
            if (Math.abs(s) >= Math.abs(x)) {
                c += (s - t) + x;
            } else {
                c += (x - t) + s;
            }
            s = t;
        }
        return s + c;
    }

    it('four moment assumptions', () => {
        const values = [-0.03, -0.01, 0.02, 0.04, 0.07];
        const moments = new RawMomentsKleinKbn();
        for (const value of values) {
            moments.update(value);
        }
        const n = values.length;
        const mean = pySum(values) / n;
        const mu2 = pySum(values.map(x => (x - mean) ** 2)) / n;
        const skew = pySum(values.map(x => (x - mean) ** 3)) / n / mu2 ** 1.5;
        const kurt = pySum(values.map(x => (x - mean) ** 4)) / n / mu2 ** 2;
        const sr = 0.75;
        const referenceSr = 0.1;

        for (const zeroSkewness of [false, true]) {
            for (const normalKurtosis of [false, true]) {
                const s = zeroSkewness ? 0.0 : skew;
                const k = normalKurtosis ? 3.0 : kurt;
                const denominator = Math.sqrt(1 - sr * s + sr * sr * (k - 1) / 4);
                const z = (sr - referenceSr) * Math.sqrt(n - 1) / denominator;
                const expected = normCdf(z);
                const actual = probabilisticSharpeRatio(moments, sr, referenceSr, zeroSkewness, normalKurtosis);
                expect(actual).withContext(`zeroSkewness=${zeroSkewness} normalKurtosis=${normalKurtosis}`)
                    .toBeCloseTo(expected, 13);
            }
        }
    });

    it('unavailable Sharpe or moments', () => {
        const moments = new RawMomentsKleinKbn();
        moments.update(0.01);
        expect(probabilisticSharpeRatio(moments, NaN)).toBeNaN();
        expect(probabilisticSharpeRatio(moments, 0.5)).toBeNaN();
        expect(probabilisticSharpeRatio(moments, 0.5, 0.0, true, false)).toBeNaN();
    });
});
