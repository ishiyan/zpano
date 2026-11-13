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
    AdaptiveBandPass,
    Corona,
    CoronaSnapshot,
    Raster,
    acceptSnapshot,
    checked,
    clip,
    formatNumber,
} from '../corona/corona';
import { CoronaParams } from '../corona/params';
import { CoronaTrendVigorParams, coronaTrendVigorMnemonic, defaultParams, trendFlags } from './params';

export { CoronaTrendVigorParams, coronaTrendVigorMnemonic, defaultParams };

export interface CoronaTrendVigorSnapshot {
    readonly heatmap: Heatmap;
    readonly trendVigor: number;
    readonly normalizedVigor: number;
    readonly dominantCycle: number;
    readonly dominantCycleMedian: number;
    readonly bandPass: number;
    readonly quadrature: number;
    readonly cycleAmplitude: number;
    readonly trendDisplacement: number;
    readonly rawRatio: number;
    readonly lookback: number;
    readonly width: number;
    readonly hasCycleEstimate: number;
}

export class CoronaTrendVigor implements Indicator {
    private readonly engine: Corona;
    private readonly engineParams: Required<CoronaParams>;
    private readonly bandpass: AdaptiveBandPass;
    private readonly raster: Raster;
    private readonly lookbackOffset: number;
    private sampleCount = 0;
    private prices: number[] = [];
    private ratio = 0.0;
    private ready = false;
    private readonly mnemonicValue: string;
    private readonly descriptionValue: string;
    private readonly mnemonicTV: string;
    private readonly descriptionTV: string;
    private readonly barComponentFunc: (bar: Bar) => number;
    private readonly quoteComponentFunc: (quote: Quote) => number;
    private readonly tradeComponentFunc: (trade: Trade) => number;

    public static default(): CoronaTrendVigor {
        return new CoronaTrendVigor(defaultParams());
    }

    public static fromParams(params: CoronaTrendVigorParams): CoronaTrendVigor {
        return new CoronaTrendVigor(params);
    }

    public constructor(params?: CoronaTrendVigorParams) {
        const rasterLength = params?.rasterLength ?? 50;
        const maxRaster = params?.maxRasterValue ?? 20;
        const minParam = params?.minParameterValue ?? -10;
        const maxParam = params?.maxParameterValue ?? 10;
        const hpCutoff = params?.highPassFilterCutoff ?? 30;
        const minPeriod = params?.minimalPeriod ?? 6;
        const maxPeriod = params?.maximalPeriod ?? 30;
        const lookbackOffset = params?.lookbackOffset ?? 0;

        if (typeof lookbackOffset !== 'number' || !Number.isInteger(lookbackOffset) || (lookbackOffset !== 0 && lookbackOffset !== -1)) {
            throw new Error('invalid corona trend vigor parameters: lookbackOffset must be 0 or -1');
        }

        this.raster = new Raster(rasterLength, minParam, maxParam, maxRaster);
        this.engine = new Corona({
            highPassFilterCutoff: hpCutoff,
            minimalPeriod: minPeriod,
            maximalPeriod: maxPeriod,
            decibelsLowerThreshold: 6.0,
            decibelsUpperThreshold: this.raster.ceiling,
        });
        this.engineParams = this.engine.parameters;
        this.bandpass = new AdaptiveBandPass();
        this.lookbackOffset = lookbackOffset;

        const bc: BarComponent = params?.barComponent ?? DefaultBarComponent;
        const qc = params?.quoteComponent ?? DefaultQuoteComponent;
        const tc = params?.tradeComponent ?? DefaultTradeComponent;
        this.barComponentFunc = barComponentValue(bc);
        this.quoteComponentFunc = quoteComponentValue(qc);
        this.tradeComponentFunc = tradeComponentValue(tc);

        const comp = componentTripleMnemonic(bc, qc, tc);
        const flags = trendFlags({ minimalPeriod: minPeriod, maximalPeriod: maxPeriod, lookbackOffset });
        this.mnemonicValue = `ctv(${this.raster.length}, ${formatNumber(this.raster.ceiling)}, ${formatNumber(this.raster.lower)}, ${formatNumber(this.raster.upper)}, ${hpCutoff}${flags}${comp})`;
        this.mnemonicTV = `ctv-tv(${hpCutoff}${flags}${comp})`;
        this.descriptionValue = 'Corona trend vigor ' + this.mnemonicValue;
        this.descriptionTV = 'Corona trend vigor scalar ' + this.mnemonicTV;
    }

    public reset(): void {
        this.engine.reset();
        this.bandpass.reset();
        this.raster.reset();
        this.sampleCount = 0;
        this.prices = [];
        this.ratio = 0.0;
        this.ready = false;
    }

    public isPrimed(): boolean {
        return this.ready;
    }

    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.CoronaTrendVigor,
            this.mnemonicValue,
            this.descriptionValue,
            [
                { mnemonic: this.mnemonicValue, description: this.descriptionValue },
                { mnemonic: this.mnemonicTV, description: this.descriptionTV },
            ],
        );
    }

    private emptySnapshot(time: Date): CoronaTrendVigorSnapshot {
        return {
            heatmap: this.raster.empty(time),
            trendVigor: Number.NaN,
            normalizedVigor: Number.NaN,
            dominantCycle: Number.NaN,
            dominantCycleMedian: Number.NaN,
            bandPass: Number.NaN,
            quadrature: Number.NaN,
            cycleAmplitude: Number.NaN,
            trendDisplacement: Number.NaN,
            rawRatio: Number.NaN,
            lookback: Number.NaN,
            width: Number.NaN,
            hasCycleEstimate: Number.NaN,
        };
    }

    public update(sample: number, time: Date = new Date(0)): CoronaTrendVigorSnapshot {
        return this.updateFromSnapshot(this.engine.update(sample, time));
    }

    public updateFromSnapshot(snapshot: CoronaSnapshot): CoronaTrendVigorSnapshot {
        if (!acceptSnapshot(snapshot, this.engineParams, this.sampleCount + 1)) {
            return this.emptySnapshot(snapshot.time);
        }
        const cycle = this.bandpass.update(snapshot.price, snapshot.dominantCycleMedian);
        this.sampleCount = snapshot.sampleCount;

        this.prices.push(snapshot.price);
        if (this.prices.length > this.engineParams.maximalPeriod + 1) {
            this.prices.shift();
        }

        const lag = Math.max(1, Math.floor(snapshot.dominantCycleMedian) + this.lookbackOffset);
        const hasHistory = this.prices.length > lag;
        const trend = hasHistory ? snapshot.price - this.prices[this.prices.length - 1 - lag] : Number.NaN;
        const rawRatio = hasHistory && cycle.amplitude > 0.0 ? trend / cycle.amplitude : Number.NaN;
        if (hasHistory && trend !== 0.0 && cycle.amplitude > 0.0) {
            this.ratio = clip(checked(0.33 * rawRatio + 0.67 * this.ratio), -10.0, 10.0);
        }

        const vigor = 0.05 * (this.ratio + 10.0);
        let width = 0.3 <= vigor && vigor <= 0.7 ? Math.min(vigor - 0.3, 0.7 - vigor) : 0.01;
        if (width === 0.0) {
            width = 0.01;
        }

        this.raster.update(vigor, width, 'trend', vigor < 0.3 || vigor > 0.7);
        this.ready = snapshot.primed && hasHistory;
        if (!this.ready) {
            return this.emptySnapshot(snapshot.time);
        }

        const mapped = this.raster.lower + (this.raster.upper - this.raster.lower) * vigor;
        return {
            heatmap: this.raster.column(snapshot.time),
            trendVigor: mapped,
            normalizedVigor: vigor,
            dominantCycle: snapshot.dominantCycle,
            dominantCycleMedian: snapshot.dominantCycleMedian,
            bandPass: cycle.bandPass,
            quadrature: cycle.quadrature,
            cycleAmplitude: cycle.amplitude,
            trendDisplacement: trend,
            rawRatio,
            lookback: lag,
            width,
            hasCycleEstimate: snapshot.hasCycleEstimate ? 1.0 : 0.0,
        };
    }

    private wrapEntity(time: Date, snapshot: CoronaTrendVigorSnapshot): IndicatorOutput {
        const tvScalar = new Scalar();
        tvScalar.time = time;
        tvScalar.value = snapshot.trendVigor;
        return [snapshot.heatmap, tvScalar];
    }

    public updateScalar(sample: Scalar): IndicatorOutput {
        return this.wrapEntity(sample.time, this.update(sample.value, sample.time));
    }

    public updateBar(sample: Bar): IndicatorOutput {
        return this.wrapEntity(sample.time, this.update(this.barComponentFunc(sample), sample.time));
    }

    public updateQuote(sample: Quote): IndicatorOutput {
        return this.wrapEntity(sample.time, this.update(this.quoteComponentFunc(sample), sample.time));
    }

    public updateTrade(sample: Trade): IndicatorOutput {
        return this.wrapEntity(sample.time, this.update(this.tradeComponentFunc(sample), sample.time));
    }
}
