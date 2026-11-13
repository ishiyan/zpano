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
    acceptSnapshot,
    formatNumber,
    validateFinite,
    validateInteger,
} from '../corona/corona';
import { CoronaParams } from '../corona/params';
import { CoronaSpectrumParams, coronaSpectrumMnemonic, defaultParams } from './params';

export { CoronaSpectrumParams, coronaSpectrumMnemonic, defaultParams };

export interface CoronaSpectrumSnapshot {
    readonly heatmap: Heatmap;
    readonly dominantCycle: number;
    readonly dominantCycleMedian: number;
    readonly maximumPower: number;
    readonly normalizationPower: number;
    readonly selectedBins: number;
    readonly hasCycleEstimate: number;
    readonly power: number[];
    readonly relativePower: number[];
}

export class CoronaSpectrum implements Indicator {
    private readonly engine: Corona;
    private readonly engineParams: Required<CoronaParams>;
    private sampleCount = 0;
    private readonly minParameterValue: number;
    private readonly maxParameterValue: number;
    private readonly parameterResolution: number;
    private readonly mnemonicValue: string;
    private readonly descriptionValue: string;
    private readonly mnemonicDC: string;
    private readonly descriptionDC: string;
    private readonly mnemonicDCM: string;
    private readonly descriptionDCM: string;
    private readonly barComponentFunc: (bar: Bar) => number;
    private readonly quoteComponentFunc: (quote: Quote) => number;
    private readonly tradeComponentFunc: (trade: Trade) => number;

    public static default(): CoronaSpectrum {
        return new CoronaSpectrum(defaultParams());
    }

    public static fromParams(params: CoronaSpectrumParams): CoronaSpectrum {
        return new CoronaSpectrum(params);
    }

    public constructor(params?: CoronaSpectrumParams) {
        const minRaster = validateFinite('minRasterValue', params?.minRasterValue ?? 6);
        const maxRaster = validateFinite('maxRasterValue', params?.maxRasterValue ?? 20);
        const minParamRaw = validateFinite('minParameterValue', params?.minParameterValue ?? 6);
        const maxParamRaw = validateFinite('maxParameterValue', params?.maxParameterValue ?? 30);
        const hpCutoff = validateInteger('highPassFilterCutoff', params?.highPassFilterCutoff ?? 30, 5);

        const minPeriod = Math.ceil(minParamRaw);
        const maxPeriod = Math.floor(maxParamRaw);
        if (minPeriod < 6) {
            throw new Error('invalid corona spectrum parameters: minParameterValue must round up to >= 6');
        }
        if (maxPeriod <= minPeriod) {
            throw new Error('invalid corona spectrum parameters: maxParameterValue must exceed minParameterValue');
        }

        const bc: BarComponent = params?.barComponent ?? DefaultBarComponent;
        const qc = params?.quoteComponent ?? DefaultQuoteComponent;
        const tc = params?.tradeComponent ?? DefaultTradeComponent;
        this.barComponentFunc = barComponentValue(bc);
        this.quoteComponentFunc = quoteComponentValue(qc);
        this.tradeComponentFunc = tradeComponentValue(tc);

        this.engine = new Corona({
            highPassFilterCutoff: hpCutoff,
            minimalPeriod: minPeriod,
            maximalPeriod: maxPeriod,
            decibelsLowerThreshold: minRaster,
            decibelsUpperThreshold: maxRaster,
        });
        this.engineParams = this.engine.parameters;
        this.minParameterValue = minPeriod;
        this.maxParameterValue = maxPeriod;
        this.parameterResolution = 2.0;

        const comp = componentTripleMnemonic(bc, qc, tc);
        this.mnemonicValue = `cspect(${formatNumber(minRaster)}, ${formatNumber(maxRaster)}, ${minPeriod}, ${maxPeriod}, ${hpCutoff}${comp})`;
        this.mnemonicDC = `cspect-dc(${hpCutoff}${comp})`;
        this.mnemonicDCM = `cspect-dcm(${hpCutoff}${comp})`;
        this.descriptionValue = 'Corona spectrum ' + this.mnemonicValue;
        this.descriptionDC = 'Corona spectrum dominant cycle ' + this.mnemonicDC;
        this.descriptionDCM = 'Corona spectrum dominant cycle median ' + this.mnemonicDCM;
    }

    public reset(): void {
        this.engine.reset();
        this.sampleCount = 0;
    }

    public isPrimed(): boolean {
        return this.sampleCount >= FIRST_BANK_SAMPLE;
    }

    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.CoronaSpectrum,
            this.mnemonicValue,
            this.descriptionValue,
            [
                { mnemonic: this.mnemonicValue, description: this.descriptionValue },
                { mnemonic: this.mnemonicDC, description: this.descriptionDC },
                { mnemonic: this.mnemonicDCM, description: this.descriptionDCM },
            ],
        );
    }

    private emptySnapshot(time: Date): CoronaSpectrumSnapshot {
        const nanSlice = new Array<number>(this.engine.filterBankLength).fill(Number.NaN);
        return {
            heatmap: Heatmap.newEmptyHeatmap(
                time,
                this.minParameterValue,
                this.maxParameterValue,
                this.parameterResolution,
            ),
            dominantCycle: Number.NaN,
            dominantCycleMedian: Number.NaN,
            maximumPower: Number.NaN,
            normalizationPower: Number.NaN,
            selectedBins: Number.NaN,
            hasCycleEstimate: Number.NaN,
            power: nanSlice,
            relativePower: nanSlice.slice(),
        };
    }

    public update(sample: number, time: Date = new Date(0)): CoronaSpectrumSnapshot {
        return this.updateFromSnapshot(this.engine.update(sample, time));
    }

    public updateFromSnapshot(snapshot: CoronaSnapshot): CoronaSpectrumSnapshot {
        if (!acceptSnapshot(snapshot, this.engineParams, this.sampleCount + 1)) {
            return this.emptySnapshot(snapshot.time);
        }
        this.sampleCount = snapshot.sampleCount;
        if (!snapshot.primed) {
            return this.emptySnapshot(snapshot.time);
        }

        const values = snapshot.decibels.slice();
        let minimum = values[0];
        let maximum = values[0];
        for (let i = 1; i < values.length; i++) {
            if (values[i] < minimum) minimum = values[i];
            if (values[i] > maximum) maximum = values[i];
        }
        const heatmap = Heatmap.newHeatmap(
            snapshot.time,
            this.minParameterValue,
            this.maxParameterValue,
            this.parameterResolution,
            minimum,
            maximum,
            values,
        );
        return {
            heatmap,
            dominantCycle: snapshot.dominantCycle,
            dominantCycleMedian: snapshot.dominantCycleMedian,
            maximumPower: snapshot.maximumPower,
            normalizationPower: snapshot.normalizationPower,
            selectedBins: snapshot.selectedBins,
            hasCycleEstimate: snapshot.hasCycleEstimate ? 1.0 : 0.0,
            power: snapshot.power.slice(),
            relativePower: snapshot.relativePower.slice(),
        };
    }

    private wrapEntity(time: Date, snapshot: CoronaSpectrumSnapshot): IndicatorOutput {
        const dcScalar = new Scalar();
        dcScalar.time = time;
        dcScalar.value = snapshot.dominantCycle;

        const dcmScalar = new Scalar();
        dcmScalar.time = time;
        dcmScalar.value = snapshot.dominantCycleMedian;

        return [snapshot.heatmap, dcScalar, dcmScalar];
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
