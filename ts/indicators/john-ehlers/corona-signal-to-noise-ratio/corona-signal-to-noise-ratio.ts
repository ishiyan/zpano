import { Bar } from '../../../entities/bar';
import { BarComponent, DefaultBarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { Indicator } from '../../core/indicator';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { Heatmap } from '../../core/outputs/heatmap';
import {
    Corona,
    CoronaSnapshot,
    FIRST_BANK_SAMPLE,
    Raster,
    acceptSnapshot,
    checked,
    clip,
    formatNumber,
    validPrice,
} from '../corona/corona';
import { CoronaParams } from '../corona/params';
import {
    CoronaSignalToNoiseRatioParams,
    coronaSignalToNoiseRatioMnemonic,
    defaultParams,
    snrFlags,
} from './params';

export { CoronaSignalToNoiseRatioParams, coronaSignalToNoiseRatioMnemonic, defaultParams };

export interface CoronaSignalToNoiseRatioSnapshot {
    readonly heatmap: Heatmap;
    readonly signalToNoiseRatio: number;
    readonly physicalDecibels: number;
    readonly calibratedDecibels: number;
    readonly normalizedPosition: number;
    readonly signal: number;
    readonly noise: number;
    readonly medianBarRange: number;
    readonly width: number;
    readonly dominantCycle: number;
    readonly dominantCycleMedian: number;
    readonly hasCycleEstimate: number;
    readonly ratioIsFinite: number;
}

function validateHighLow(low: number, high: number): number {
    if (typeof low !== 'number' || typeof high !== 'number') {
        throw new Error('high and low must be numbers');
    }
    if (Number.isNaN(low) || Number.isNaN(high)) {
        return Number.NaN;
    }
    if (!Number.isFinite(low) || !Number.isFinite(high)) {
        throw new Error('high and low must be finite');
    }
    if (high < low) {
        throw new Error('high must be >= low');
    }
    return checked(high - low);
}

function validateBarRange(barRange: number): number {
    if (typeof barRange !== 'number') {
        throw new Error('barRange must be a number');
    }
    if (Number.isNaN(barRange)) {
        return Number.NaN;
    }
    if (!Number.isFinite(barRange) || barRange < 0.0) {
        throw new Error('barRange must be finite and >= 0');
    }
    return barRange;
}

export class CoronaSignalToNoiseRatio implements Indicator {
    private readonly engine: Corona;
    private readonly engineParams: Required<CoronaParams>;
    private readonly raster: Raster;
    private sampleCount = 0;
    private averagePrice = 0.0;
    private signal = 0.0;
    private noise = 0.0;
    private ranges: number[] = [0.0, 0.0, 0.0, 0.0, 0.0];
    private readonly mnemonicValue: string;
    private readonly descriptionValue: string;
    private readonly mnemonicSNR: string;
    private readonly descriptionSNR: string;
    private readonly barComponentFunc: (bar: Bar) => number;
    private readonly quoteComponentFunc: (quote: Quote) => number;
    private readonly tradeComponentFunc: (trade: Trade) => number;

    public static default(): CoronaSignalToNoiseRatio {
        return new CoronaSignalToNoiseRatio(defaultParams());
    }

    public static fromParams(params: CoronaSignalToNoiseRatioParams): CoronaSignalToNoiseRatio {
        return new CoronaSignalToNoiseRatio(params);
    }

    public constructor(params?: CoronaSignalToNoiseRatioParams) {
        const rasterLength = params?.rasterLength ?? 50;
        const maxRaster = params?.maxRasterValue ?? 20;
        const minParam = params?.minParameterValue ?? 1;
        const maxParam = params?.maxParameterValue ?? 11;
        const hpCutoff = params?.highPassFilterCutoff ?? 30;
        const minPeriod = params?.minimalPeriod ?? 6;
        const maxPeriod = params?.maximalPeriod ?? 30;

        this.raster = new Raster(rasterLength, minParam, maxParam, maxRaster);
        this.engine = new Corona({
            highPassFilterCutoff: hpCutoff,
            minimalPeriod: minPeriod,
            maximalPeriod: maxPeriod,
            decibelsLowerThreshold: 6.0,
            decibelsUpperThreshold: this.raster.ceiling,
        });
        this.engineParams = this.engine.parameters;

        const bc: BarComponent = params?.barComponent ?? DefaultBarComponent;
        const qc = params?.quoteComponent ?? DefaultQuoteComponent;
        const tc = params?.tradeComponent ?? DefaultTradeComponent;
        this.barComponentFunc = barComponentValue(bc);
        this.quoteComponentFunc = quoteComponentValue(qc);
        this.tradeComponentFunc = tradeComponentValue(tc);

        const comp = componentTripleMnemonic(bc, qc, tc);
        const flags = snrFlags({ minimalPeriod: minPeriod, maximalPeriod: maxPeriod });
        this.mnemonicValue = `csnr(${this.raster.length}, ${formatNumber(this.raster.ceiling)}, ${formatNumber(this.raster.lower)}, ${formatNumber(this.raster.upper)}, ${hpCutoff}${flags}${comp})`;
        this.mnemonicSNR = `csnr-snr(${hpCutoff}${flags}${comp})`;
        this.descriptionValue = 'Corona signal to noise ratio ' + this.mnemonicValue;
        this.descriptionSNR = 'Corona signal to noise ratio scalar ' + this.mnemonicSNR;
    }

    public reset(): void {
        this.engine.reset();
        this.raster.reset();
        this.sampleCount = 0;
        this.averagePrice = 0.0;
        this.signal = 0.0;
        this.noise = 0.0;
        this.ranges = [0.0, 0.0, 0.0, 0.0, 0.0];
    }

    public isPrimed(): boolean {
        return this.sampleCount >= FIRST_BANK_SAMPLE;
    }

    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.CoronaSignalToNoiseRatio,
            this.mnemonicValue,
            this.descriptionValue,
            [
                { mnemonic: this.mnemonicValue, description: this.descriptionValue },
                { mnemonic: this.mnemonicSNR, description: this.descriptionSNR },
            ],
        );
    }

    private emptySnapshot(time: Date): CoronaSignalToNoiseRatioSnapshot {
        return {
            heatmap: this.raster.empty(time),
            signalToNoiseRatio: Number.NaN,
            physicalDecibels: Number.NaN,
            calibratedDecibels: Number.NaN,
            normalizedPosition: Number.NaN,
            signal: Number.NaN,
            noise: Number.NaN,
            medianBarRange: Number.NaN,
            width: Number.NaN,
            dominantCycle: Number.NaN,
            dominantCycleMedian: Number.NaN,
            hasCycleEstimate: Number.NaN,
            ratioIsFinite: Number.NaN,
        };
    }

    private updateFromSnapshotRangeInternal(
        snapshot: CoronaSnapshot,
        barRange: number,
    ): CoronaSignalToNoiseRatioSnapshot {
        if (!acceptSnapshot(snapshot, this.engineParams, this.sampleCount + 1)) {
            return this.emptySnapshot(snapshot.time);
        }
        const rangeValue = validateBarRange(barRange);
        if (Number.isNaN(rangeValue)) {
            throw new Error('accepted shared snapshot requires a finite bar range');
        }
        this.sampleCount = snapshot.sampleCount;

        this.ranges = [this.ranges[1], this.ranges[2], this.ranges[3], this.ranges[4], rangeValue];
        const sortedRanges = this.ranges.slice().sort((a, b) => a - b);
        const medianRange = sortedRanges[2];

        if (snapshot.sampleCount === 1) {
            this.averagePrice = snapshot.price;
        } else {
            this.averagePrice = checked(0.1 * snapshot.price + 0.9 * this.averagePrice);
            if (this.averagePrice !== 0.0 || snapshot.maximumPower > 0.0) {
                this.signal = checked(0.2 * Math.sqrt(snapshot.maximumPower) + 0.9 * this.signal);
            }
            if (this.averagePrice !== 0.0 && snapshot.sampleCount > 5) {
                this.noise = checked(0.1 * medianRange + 0.9 * this.noise);
            }
        }

        let physicalDecibels: number;
        if (this.signal > 0.0 && this.noise > 0.0) {
            physicalDecibels = 20.0 * (Math.log10(this.signal) - Math.log10(this.noise));
        } else if (this.signal > 0.0) {
            physicalDecibels = Number.POSITIVE_INFINITY;
        } else if (this.noise > 0.0) {
            physicalDecibels = Number.NEGATIVE_INFINITY;
        } else {
            physicalDecibels = Number.NaN;
        }

        const calibratedDecibels = physicalDecibels + 3.5;
        const position = Number.isNaN(calibratedDecibels)
            ? 0.0
            : clip(calibratedDecibels, 0.0, 10.0) / 10.0;
        const width = Math.max(0.0, 0.2 - 0.4 * position);
        this.raster.update(position, width, 'snr', position > 0.5);

        if (!snapshot.primed) {
            return this.emptySnapshot(snapshot.time);
        }

        const mapped = this.raster.lower + (this.raster.upper - this.raster.lower) * position;
        return {
            heatmap: this.raster.column(snapshot.time),
            signalToNoiseRatio: mapped,
            physicalDecibels,
            calibratedDecibels,
            normalizedPosition: position,
            signal: this.signal,
            noise: this.noise,
            medianBarRange: medianRange,
            width,
            dominantCycle: snapshot.dominantCycle,
            dominantCycleMedian: snapshot.dominantCycleMedian,
            hasCycleEstimate: snapshot.hasCycleEstimate ? 1.0 : 0.0,
            ratioIsFinite: Number.isFinite(physicalDecibels) ? 1.0 : 0.0,
        };
    }

    public update(
        sample: number,
        sampleLow: number,
        sampleHigh: number,
        time: Date = new Date(0),
    ): CoronaSignalToNoiseRatioSnapshot {
        if (!validPrice(sample)) {
            return this.emptySnapshot(time);
        }
        const rangeValue = validateHighLow(sampleLow, sampleHigh);
        if (Number.isNaN(rangeValue)) {
            return this.emptySnapshot(time);
        }
        return this.updateFromSnapshotRangeInternal(this.engine.update(sample, time), rangeValue);
    }

    public updateRange(
        sample: number,
        barRange: number,
        time: Date = new Date(0),
    ): CoronaSignalToNoiseRatioSnapshot {
        if (!validPrice(sample)) {
            return this.emptySnapshot(time);
        }
        const rangeValue = validateBarRange(barRange);
        if (Number.isNaN(rangeValue)) {
            return this.emptySnapshot(time);
        }
        return this.updateFromSnapshotRangeInternal(this.engine.update(sample, time), rangeValue);
    }

    public updateFromSnapshot(
        snapshot: CoronaSnapshot,
        sampleLow: number,
        sampleHigh: number,
    ): CoronaSignalToNoiseRatioSnapshot {
        if (!snapshot.accepted) {
            return this.updateFromSnapshotRangeInternal(snapshot, 0.0);
        }
        const rangeValue = validateHighLow(sampleLow, sampleHigh);
        return this.updateFromSnapshotRangeInternal(snapshot, rangeValue);
    }

    public updateFromSnapshotRange(
        snapshot: CoronaSnapshot,
        barRange: number,
    ): CoronaSignalToNoiseRatioSnapshot {
        return this.updateFromSnapshotRangeInternal(snapshot, barRange);
    }

    private wrapEntity(time: Date, snapshot: CoronaSignalToNoiseRatioSnapshot): IndicatorOutput {
        const snrScalar = new Scalar();
        snrScalar.time = time;
        snrScalar.value = snapshot.signalToNoiseRatio;
        return [snapshot.heatmap, snrScalar];
    }

    public updateScalar(sample: Scalar): IndicatorOutput {
        return this.wrapEntity(sample.time, this.update(sample.value, sample.value, sample.value, sample.time));
    }

    public updateBar(sample: Bar): IndicatorOutput {
        return this.wrapEntity(
            sample.time,
            this.update(this.barComponentFunc(sample), sample.low, sample.high, sample.time),
        );
    }

    public updateQuote(sample: Quote): IndicatorOutput {
        const value = this.quoteComponentFunc(sample);
        return this.wrapEntity(sample.time, this.update(value, value, value, sample.time));
    }

    public updateTrade(sample: Trade): IndicatorOutput {
        const value = this.tradeComponentFunc(sample);
        return this.wrapEntity(sample.time, this.update(value, value, value, sample.time));
    }
}
