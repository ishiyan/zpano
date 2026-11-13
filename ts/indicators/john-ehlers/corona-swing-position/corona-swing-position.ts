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
    FIRST_BANK_SAMPLE,
    Raster,
    acceptSnapshot,
    clip,
    formatNumber,
    rollingAppend,
    validateInteger,
} from '../corona/corona';
import { CoronaParams } from '../corona/params';
import { CoronaSwingPositionParams, coronaSwingPositionMnemonic, defaultParams, swingFlags } from './params';

export { CoronaSwingPositionParams, coronaSwingPositionMnemonic, defaultParams };

export interface CoronaSwingPositionSnapshot {
    readonly heatmap: Heatmap;
    readonly swingPosition: number;
    readonly normalizedPosition: number;
    readonly dominantCycle: number;
    readonly dominantCycleMedian: number;
    readonly bandPass: number;
    readonly quadrature: number;
    readonly cycleAmplitude: number;
    readonly lead60: number;
    readonly positionSpan: number;
    readonly width: number;
    readonly hasCycleEstimate: number;
}

export class CoronaSwingPosition implements Indicator {
    private readonly engine: Corona;
    private readonly engineParams: Required<CoronaParams>;
    private readonly bandpass: AdaptiveBandPass;
    private readonly raster: Raster;
    private readonly leadWindow: number;
    private readonly positionWindow: number;
    private sampleCount = 0;
    private leads: number[] = [];
    private positions: number[] = [];
    private readonly mnemonicValue: string;
    private readonly descriptionValue: string;
    private readonly mnemonicSP: string;
    private readonly descriptionSP: string;
    private readonly barComponentFunc: (bar: Bar) => number;
    private readonly quoteComponentFunc: (quote: Quote) => number;
    private readonly tradeComponentFunc: (trade: Trade) => number;

    public static default(): CoronaSwingPosition {
        return new CoronaSwingPosition(defaultParams());
    }

    public static fromParams(params: CoronaSwingPositionParams): CoronaSwingPosition {
        return new CoronaSwingPosition(params);
    }

    public constructor(params?: CoronaSwingPositionParams) {
        const rasterLength = params?.rasterLength ?? 50;
        const maxRaster = params?.maxRasterValue ?? 20;
        const minParam = params?.minParameterValue ?? -5;
        const maxParam = params?.maxParameterValue ?? 5;
        const hpCutoff = params?.highPassFilterCutoff ?? 30;
        const minPeriod = params?.minimalPeriod ?? 6;
        const maxPeriod = params?.maximalPeriod ?? 30;
        const leadWindow = validateInteger('leadWindow', params?.leadWindow ?? 51, 2);
        const positionWindow = validateInteger('positionWindow', params?.positionWindow ?? 21, 2);

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
        this.leadWindow = leadWindow;
        this.positionWindow = positionWindow;

        const bc: BarComponent = params?.barComponent ?? DefaultBarComponent;
        const qc = params?.quoteComponent ?? DefaultQuoteComponent;
        const tc = params?.tradeComponent ?? DefaultTradeComponent;
        this.barComponentFunc = barComponentValue(bc);
        this.quoteComponentFunc = quoteComponentValue(qc);
        this.tradeComponentFunc = tradeComponentValue(tc);

        const comp = componentTripleMnemonic(bc, qc, tc);
        const flags = swingFlags({ minimalPeriod: minPeriod, maximalPeriod: maxPeriod, leadWindow, positionWindow });
        this.mnemonicValue = `cswing(${this.raster.length}, ${formatNumber(this.raster.ceiling)}, ${formatNumber(this.raster.lower)}, ${formatNumber(this.raster.upper)}, ${hpCutoff}${flags}${comp})`;
        this.mnemonicSP = `cswing-sp(${hpCutoff}${flags}${comp})`;
        this.descriptionValue = 'Corona swing position ' + this.mnemonicValue;
        this.descriptionSP = 'Corona swing position scalar ' + this.mnemonicSP;
    }

    public reset(): void {
        this.engine.reset();
        this.bandpass.reset();
        this.raster.reset();
        this.sampleCount = 0;
        this.leads = [];
        this.positions = [];
    }

    public isPrimed(): boolean {
        return this.sampleCount >= FIRST_BANK_SAMPLE;
    }

    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.CoronaSwingPosition,
            this.mnemonicValue,
            this.descriptionValue,
            [
                { mnemonic: this.mnemonicValue, description: this.descriptionValue },
                { mnemonic: this.mnemonicSP, description: this.descriptionSP },
            ],
        );
    }

    private emptySnapshot(time: Date): CoronaSwingPositionSnapshot {
        return {
            heatmap: this.raster.empty(time),
            swingPosition: Number.NaN,
            normalizedPosition: Number.NaN,
            dominantCycle: Number.NaN,
            dominantCycleMedian: Number.NaN,
            bandPass: Number.NaN,
            quadrature: Number.NaN,
            cycleAmplitude: Number.NaN,
            lead60: Number.NaN,
            positionSpan: Number.NaN,
            width: Number.NaN,
            hasCycleEstimate: Number.NaN,
        };
    }

    public update(sample: number, time: Date = new Date(0)): CoronaSwingPositionSnapshot {
        return this.updateFromSnapshot(this.engine.update(sample, time));
    }

    public updateFromSnapshot(snapshot: CoronaSnapshot): CoronaSwingPositionSnapshot {
        if (!acceptSnapshot(snapshot, this.engineParams, this.sampleCount + 1)) {
            return this.emptySnapshot(snapshot.time);
        }
        const cycle = this.bandpass.update(snapshot.price, snapshot.dominantCycleMedian);
        this.sampleCount = snapshot.sampleCount;

        const [lowest, highest] = rollingAppend(this.leads, cycle.lead60, this.leadWindow);
        const flat = highest === lowest;
        const position = clip(flat ? 0.5 : (cycle.lead60 - lowest) / (highest - lowest), 0.0, 1.0);
        const [lowPosition, highPosition] = rollingAppend(this.positions, position, this.positionWindow);
        const span = highPosition - lowPosition;
        let width = span > 0.85 ? 0.01 : 0.15 * span;
        if (width === 0.0) {
            width = 0.01;
        }

        this.raster.update(position, width, 'swing', span > 0.8 || flat);
        if (!snapshot.primed) {
            return this.emptySnapshot(snapshot.time);
        }

        const mapped = this.raster.lower + (this.raster.upper - this.raster.lower) * position;
        return {
            heatmap: this.raster.column(snapshot.time),
            swingPosition: mapped,
            normalizedPosition: position,
            dominantCycle: snapshot.dominantCycle,
            dominantCycleMedian: snapshot.dominantCycleMedian,
            bandPass: cycle.bandPass,
            quadrature: cycle.quadrature,
            cycleAmplitude: cycle.amplitude,
            lead60: cycle.lead60,
            positionSpan: span,
            width,
            hasCycleEstimate: snapshot.hasCycleEstimate ? 1.0 : 0.0,
        };
    }

    private wrapEntity(time: Date, snapshot: CoronaSwingPositionSnapshot): IndicatorOutput {
        const spScalar = new Scalar();
        spScalar.time = time;
        spScalar.value = snapshot.swingPosition;
        return [snapshot.heatmap, spScalar];
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
