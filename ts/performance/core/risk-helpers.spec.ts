import { RawMomentsKleinKbn } from '../../streaming-kbn';
import { esCornishFisher, esGaussian, esHistorical } from './es';
import { varCornishFisher, varGaussian, varHistorical } from './var';

describe('risk helpers (VaR / ES)', () => {

    it('historical quantile, tail and risk-free rate', () => {
        const returns = [-0.2, -0.1, 0.0, 0.1];
        expect(varHistorical(returns, 0.0, 0.75)).toBeCloseTo(0.125, 15);
        expect(esHistorical(returns, 0.0, 0.75)).toBeCloseTo(0.2, 15);
        expect(varHistorical(returns, 0.01, 0.75)).toBeCloseTo(0.135, 15);
        expect(esHistorical(returns, 0.01, 0.75)).toBeCloseTo(0.21, 15);
        expect(returns).toEqual([-0.2, -0.1, 0.0, 0.1]);
    });

    it('empty historical inputs', () => {
        for (const returns of [null, undefined, []]) {
            expect(varHistorical(returns)).withContext(`${returns}`).toBeNaN();
            expect(esHistorical(returns)).withContext(`${returns}`).toBeNaN();
        }
    });

    it('Cornish-Fisher falls back for one sample', () => {
        const moments = new RawMomentsKleinKbn();
        moments.update(0.02);
        expect(varGaussian(moments)).toBeCloseTo(-0.02, 15);
        expect(esGaussian(moments)).toBeCloseTo(-0.02, 15);
        expect(varCornishFisher(moments)).toBe(varGaussian(moments));
        expect(esCornishFisher(moments)).toBe(esGaussian(moments));
    });
});
