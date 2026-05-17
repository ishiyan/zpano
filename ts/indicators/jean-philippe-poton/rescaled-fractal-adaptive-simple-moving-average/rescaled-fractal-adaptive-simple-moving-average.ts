import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { RescaledFractalAdaptiveSimpleMovingAverageParams } from './params';

/** Function to calculate mnemonic of a __RescaledFractalAdaptiveSimpleMovingAverage__ indicator. */
export const rescaledFractalAdaptiveSimpleMovingAverageMnemonic = (params: RescaledFractalAdaptiveSimpleMovingAverageParams): string =>
    'rsfrasma('.concat(
        params.period.toString(), ',',
        params.normalSpeed.toString(), ',',
        params.priceScale.toFixed(1),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent), ')');

/** Rescaled Fractal Adaptive Simple Moving Average (RSFRASMA) line indicator. */
export class RescaledFractalAdaptiveSimpleMovingAverage extends LineIndicator {
    private _closes: Array<number>;
    private _period: number;
    private _normalSpeed: number;
    private _priceScale: number;
    private _nIter: number;
    private _blockSizes: number[];
    private _blockCounts: number[];

    /**
     * Constructs an instance given parameters.
     * Period must be a power of 2, >= 4.
     * NormalSpeed must be >= 1.
     **/
    public constructor(params: RescaledFractalAdaptiveSimpleMovingAverageParams) {
        super();
        const period = Math.floor(params.period);
        const normalSpeed = Math.floor(params.normalSpeed);
        const priceScale = params.priceScale || 1.0;

        if (period < 4) {
            throw new Error('period should be greater than 3');
        }

        if ((period & (period - 1)) !== 0) {
            throw new Error('period must be a power of 2');
        }

        if (normalSpeed < 1) {
            throw new Error('normal_speed should be greater than 0');
        }

        this.mnemonic = rescaledFractalAdaptiveSimpleMovingAverageMnemonic(params);
        this.description = 'Rescaled fractal adaptive simple moving average ' + this.mnemonic;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
        this._closes = [];
        this._period = period;
        this._normalSpeed = normalSpeed;
        this._priceScale = priceScale;
        this.primed = false;

        // Precompute R/S parameters.
        const k0 = period / 4;
        let nIter = 0;

        if (k0 >= 2) {
            nIter = Math.floor(Math.log(k0) / Math.log(2));
        }

        this._nIter = nIter;
        this._blockSizes = new Array(nIter + 1).fill(0);
        this._blockCounts = new Array(nIter + 1).fill(0);

        for (let u = 1; u <= nIter; u++) {
            this._blockSizes[u] = 1 << (u + 1);
            this._blockCounts[u] = period / this._blockSizes[u];
        }
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.RescaledFractalAdaptiveSimpleMovingAverage,
            this.mnemonic,
            this.description,
            [{ mnemonic: this.mnemonic, description: this.description }],
        );
    }

    /** Updates the value of the indicator given the next sample. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        const period = this._period;
        const priceScale = this._priceScale;

        this._closes.push(sample);
        const nCloses = this._closes.length;

        if (nCloses <= period) {
            return Number.NaN;
        }

        if (!this.primed) {
            this.primed = true;
        }

        const pos = nCloses - 1;

        // R/S analysis.
        const nIter = this._nIter;
        let sumx = 0.0;
        let sumy = 0.0;
        let sumx2 = 0.0;
        let sumxy = 0.0;
        let validScales = 0;

        for (let u = 1; u <= nIter; u++) {
            const blockSize = this._blockSizes[u];
            const nBlocksU = this._blockCounts[u];

            if (nBlocksU < 1) {
                continue;
            }

            let rsSum = 0.0;
            let t = 0;
            let blockCount = 0;

            while (t <= period - blockSize) {
                // Block mean.
                let mu = 0.0;

                for (let j = 1; j <= blockSize; j++) {
                    mu += priceScale * this._closes[pos - (t + j)];
                }

                mu /= blockSize;

                // Population std.
                let sumSq = 0.0;

                for (let j = 1; j <= blockSize; j++) {
                    const diff = priceScale * this._closes[pos - (t + j)] - mu;
                    sumSq += diff * diff;
                }

                let std = Math.sqrt(sumSq / blockSize);

                if (std <= 0.0) {
                    std = 0.1;
                }

                // Cumulative deviations and range.
                let cumDev = 0.0;
                let wMax = 0.0;
                let wMin = 9999999999.0;

                for (let k = 1; k <= blockSize; k++) {
                    cumDev += priceScale * this._closes[pos - (t + k)] - mu;

                    if (cumDev > wMax) { wMax = cumDev; }
                    if (cumDev < wMin) { wMin = cumDev; }
                }

                if (wMax < 0.0) { wMax = 0.0; }
                if (wMin > 0.0) { wMin = 0.0; }

                const rVal = wMax - wMin;
                rsSum += rVal / std;
                t += blockSize;
                blockCount++;
            }

            // Average R/S for this scale.
            let rsAvg = 1.0;

            if (blockCount > 0) {
                rsAvg = rsSum / blockCount;
            }

            if (rsAvg <= 0.0) {
                rsAvg = 1e-10;
            }

            const log2D = Math.log(blockSize) / Math.log(2);
            const log2Rs = Math.log(rsAvg) / Math.log(2);

            sumx += log2D;
            sumy += log2Rs;
            sumx2 += log2D * log2D;
            sumxy += log2D * log2Rs;
            validScales++;
        }

        // Linear regression slope = Hurst exponent.
        let h = 0.5;

        if (validScales >= 2) {
            const h1 = validScales * sumxy - sumx * sumy;
            let h2 = validScales * sumx2 - sumx * sumx;

            if (h2 <= 0.0) {
                h2 = 0.1;
            }

            h = h1 / h2;
        }

        // Guard H.
        if (2.0 * h <= 0.0) {
            h = 0.001;
        }

        const alpha = 1.0 / (2.0 * h);
        let spd = Math.round(this._normalSpeed * alpha);

        if (spd < 1) {
            spd = 1;
        }

        // Compute SMA with adapted speed.
        let smaStart = pos - spd + 1;

        if (smaStart < 0) {
            smaStart = 0;
        }

        let total = 0.0;
        const count = pos - smaStart + 1;

        for (let i = smaStart; i <= pos; i++) {
            total += this._closes[i];
        }

        return total / count;
    }
}
