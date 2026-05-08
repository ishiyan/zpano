import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { NewMovingAverageParams, MAType } from './params';

// ── Streaming MA helpers ────────────────────────────────────────────────────

interface StreamingMA {
    update(sample: number): number;
}

class StreamingSMA implements StreamingMA {
    private readonly period: number;
    private readonly buffer: number[];
    private bufferIndex: number;
    private bufferCount: number;
    private sum: number;
    private _primed: boolean;

    constructor(period: number) {
        this.period = period;
        this.buffer = new Array(period).fill(0);
        this.bufferIndex = 0;
        this.bufferCount = 0;
        this.sum = 0;
        this._primed = false;
    }

    update(sample: number): number {
        if (Number.isNaN(sample)) return sample;
        const period = this.period;
        if (this._primed) {
            this.sum -= this.buffer[this.bufferIndex];
        }
        this.buffer[this.bufferIndex] = sample;
        this.sum += sample;
        this.bufferIndex = (this.bufferIndex + 1) % period;
        if (!this._primed) {
            this.bufferCount++;
            if (this.bufferCount < period) return NaN;
            this._primed = true;
        }
        return this.sum / period;
    }
}

class StreamingEMA implements StreamingMA {
    private readonly period: number;
    private readonly multiplier: number;
    private count: number;
    private sum: number;
    private value: number;
    private _primed: boolean;

    constructor(period: number) {
        this.period = period;
        this.multiplier = 2.0 / (period + 1);
        this.count = 0;
        this.sum = 0;
        this.value = NaN;
        this._primed = false;
    }

    update(sample: number): number {
        if (Number.isNaN(sample)) return sample;
        if (!this._primed) {
            this.count++;
            this.sum += sample;
            if (this.count < this.period) return NaN;
            this.value = this.sum / this.period;
            this._primed = true;
            return this.value;
        }
        this.value = (sample - this.value) * this.multiplier + this.value;
        return this.value;
    }
}

class StreamingSMMA implements StreamingMA {
    private readonly period: number;
    private count: number;
    private sum: number;
    private value: number;
    private _primed: boolean;

    constructor(period: number) {
        this.period = period;
        this.count = 0;
        this.sum = 0;
        this.value = NaN;
        this._primed = false;
    }

    update(sample: number): number {
        if (Number.isNaN(sample)) return sample;
        if (!this._primed) {
            this.count++;
            this.sum += sample;
            if (this.count < this.period) return NaN;
            this.value = this.sum / this.period;
            this._primed = true;
            return this.value;
        }
        this.value = (this.value * (this.period - 1) + sample) / this.period;
        return this.value;
    }
}

class StreamingLWMA implements StreamingMA {
    private readonly period: number;
    private readonly buffer: number[];
    private bufferIndex: number;
    private bufferCount: number;
    private readonly weightSum: number;
    private _primed: boolean;

    constructor(period: number) {
        this.period = period;
        this.buffer = new Array(period).fill(0);
        this.bufferIndex = 0;
        this.bufferCount = 0;
        this.weightSum = period * (period + 1) / 2.0;
        this._primed = false;
    }

    update(sample: number): number {
        if (Number.isNaN(sample)) return sample;
        const period = this.period;
        this.buffer[this.bufferIndex] = sample;
        this.bufferIndex = (this.bufferIndex + 1) % period;
        if (!this._primed) {
            this.bufferCount++;
            if (this.bufferCount < period) return NaN;
            this._primed = true;
        }
        let result = 0;
        let index = this.bufferIndex;
        for (let i = 0; i < period; i++) {
            result += (i + 1) * this.buffer[index];
            index = (index + 1) % period;
        }
        return result / this.weightSum;
    }
}

function createStreamingMA(maType: MAType, period: number): StreamingMA {
    switch (maType) {
        case MAType.SMA: return new StreamingSMA(period);
        case MAType.EMA: return new StreamingEMA(period);
        case MAType.SMMA: return new StreamingSMMA(period);
        case MAType.LWMA: return new StreamingLWMA(period);
        default: throw new Error(`unknown MA type: ${maType}`);
    }
}

// ── NMA indicator ───────────────────────────────────────────────────────────

/** New Moving Average (NMA) by Manfred Dürschner. */
export class NewMovingAverage extends LineIndicator {
    private readonly alpha: number;
    private readonly maPrimary: StreamingMA;
    private readonly maSecondary: StreamingMA;

    constructor(params: NewMovingAverageParams) {
        super();

        let primaryPeriod = params.primary_period ?? 0;
        const secondaryPeriod = params.secondary_period ?? 8;
        const maType = params.ma_type ?? MAType.LWMA;

        // Enforce Nyquist constraint.
        if (primaryPeriod < 4) primaryPeriod = 4;
        let sec = secondaryPeriod;
        if (sec < 2) sec = 2;
        if (primaryPeriod < sec * 2) primaryPeriod = sec * 4;

        // Compute alpha.
        const nyquistRatio = Math.floor(primaryPeriod / sec);
        this.alpha = nyquistRatio * (primaryPeriod - 1) / (primaryPeriod - nyquistRatio);

        this.mnemonic = `nma(${primaryPeriod}, ${sec}, ${maType}${componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent)})`;
        this.description = `New moving average ${this.mnemonic}`;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
        this.primed = false;

        this.maPrimary = createStreamingMA(maType, primaryPeriod);
        this.maSecondary = createStreamingMA(maType, sec);
    }

    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.NewMovingAverage,
            this.mnemonic,
            this.description,
            [{ mnemonic: this.mnemonic, description: this.description }],
        );
    }

    public update(sample: number): number {
        if (Number.isNaN(sample)) return sample;

        const ma1Value = this.maPrimary.update(sample);
        if (Number.isNaN(ma1Value)) return NaN;

        const ma2Value = this.maSecondary.update(ma1Value);
        if (Number.isNaN(ma2Value)) return NaN;

        this.primed = true;

        const alpha = this.alpha;
        return (1.0 + alpha) * ma1Value - alpha * ma2Value;
    }
}
