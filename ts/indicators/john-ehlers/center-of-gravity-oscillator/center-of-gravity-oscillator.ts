import { Bar } from '../../../entities/bar';
import { BarComponent, DefaultBarComponent, barComponentValue } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, QuoteComponent, quoteComponentValue } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, TradeComponent, tradeComponentValue } from '../../../entities/trade-component';
import { buildMetadata } from '../../core/build-metadata';
import { Indicator } from '../../core/indicator';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorOutput } from '../../core/indicator-output';
import { CenterOfGravityOscillatorOutput } from './output';
import {
    CenterOfGravityOscillatorParams,
    centerOfGravityOscillatorMnemonic,
    centerOfGravityOscillatorTriggerMnemonic,
    defaultParams,
} from './params';

export { CenterOfGravityOscillatorOutput };

function checkRealNumber(value: unknown, name: string, allowNan = false): number {
    if (typeof value !== 'number') {
        throw new Error(`${name} must be a real number`);
    }
    if (!Number.isFinite(value) && !(allowNan && Number.isNaN(value))) {
        throw new Error(`${name} must be finite${allowNan ? ' or NaN' : ''}`);
    }
    return value;
}

function checkEnumRange(value: unknown, minValue: number, maxValue: number, label: string): void {
    if (typeof value !== 'number' || !Number.isInteger(value) || value < minValue || value > maxValue) {
        throw new Error(`invalid ${label}`);
    }
}

/**
 * Ehlers' Center of Gravity Oscillator (COG).
 *
 * Supports both the May 2002 TASC / Mesa convention (`centered = false`,
 * `-Num / Denom`) and the 2004 book convention (`centered = true`,
 * `-Num / Denom + (length + 1) / 2`), exposing the oscillator value on
 * observation `length` and its one-observation delayed trigger on
 * observation `length + 1`.
 */
export class CenterOfGravityOscillator implements Indicator {
    private readonly windowLength: number;
    private readonly isCentered: boolean;
    private readonly offset: number;
    private readonly mnemonic: string;
    private readonly description: string;
    private readonly mnemonicTrig: string;
    private readonly descriptionTrig: string;
    private readonly barComponentFunc: (bar: Bar) => number;
    private readonly quoteComponentFunc: (quote: Quote) => number;
    private readonly tradeComponentFunc: (trade: Trade) => number;

    private window: number[];
    private next = 0;
    private count = 0;
    private lastValue = 0;
    private currentValue = Number.NaN;
    private currentTrigger = Number.NaN;

    public static default(): CenterOfGravityOscillator {
        return new CenterOfGravityOscillator(defaultParams());
    }

    public static fromParams(params?: CenterOfGravityOscillatorParams): CenterOfGravityOscillator {
        return new CenterOfGravityOscillator(params);
    }

    public constructor(params?: CenterOfGravityOscillatorParams) {
        const length = params === undefined || params.length === undefined ? 10 : params.length;
        if (typeof length !== 'number' || !Number.isInteger(length) || length < 1) {
            throw new Error('length should be a positive integer');
        }

        const centered = params === undefined || params.centered === undefined ? false : params.centered;
        if (typeof centered !== 'boolean') {
            throw new Error('centered must be a boolean');
        }

        if (params !== undefined && params.barComponent !== undefined) {
            checkEnumRange(params.barComponent, BarComponent.Open, BarComponent.Average, 'barComponent');
        }
        if (params !== undefined && params.quoteComponent !== undefined) {
            checkEnumRange(params.quoteComponent, QuoteComponent.Bid, QuoteComponent.SpreadBp, 'quoteComponent');
        }
        if (params !== undefined && params.tradeComponent !== undefined) {
            checkEnumRange(params.tradeComponent, TradeComponent.Price, TradeComponent.Volume, 'tradeComponent');
        }

        const barComp = params?.barComponent ?? DefaultBarComponent;
        const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
        const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;

        this.windowLength = length;
        this.isCentered = centered;
        this.offset = centered ? (length + 1) / 2 : 0;
        this.window = new Array<number>(length).fill(0);

        this.barComponentFunc = barComponentValue(barComp);
        this.quoteComponentFunc = quoteComponentValue(quoteComp);
        this.tradeComponentFunc = tradeComponentValue(tradeComp);

        const normalizedParams: CenterOfGravityOscillatorParams = {
            length,
            centered,
            barComponent: params?.barComponent,
            quoteComponent: params?.quoteComponent,
            tradeComponent: params?.tradeComponent,
        };
        this.mnemonic = centerOfGravityOscillatorMnemonic(normalizedParams);
        this.mnemonicTrig = centerOfGravityOscillatorTriggerMnemonic(normalizedParams);
        this.description = `Center of Gravity oscillator ${this.mnemonic}`;
        this.descriptionTrig = `Center of Gravity trigger ${this.mnemonicTrig}`;

        this.reset();
    }

    /** Restores startup state while preserving configuration. */
    public reset(): void {
        this.window = new Array<number>(this.windowLength).fill(0);
        this.next = 0;
        this.count = 0;
        this.lastValue = 0;
        this.currentValue = Number.NaN;
        this.currentTrigger = Number.NaN;
    }

    /** Returns the configured observation window length. */
    public length(): number {
        return this.windowLength;
    }

    /** Returns whether the zero-centered book baseline is enabled. */
    public centered(): boolean {
        return this.isCentered;
    }

    /** Returns the latest accepted oscillator value (NaN before priming). */
    public value(): number {
        return this.currentValue;
    }

    /** Returns the latest accepted trigger value (NaN until observation length+1). */
    public trigger(): number {
        return this.currentTrigger;
    }

    /** Indicates whether at least length finite observations have filled the window. */
    public isPrimed(): boolean {
        return this.count >= this.windowLength;
    }

    /** Indicates whether a preceding full-window oscillator value exists. */
    public isTriggerPrimed(): boolean {
        return this.count > this.windowLength;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.CenterOfGravityOscillator,
            this.mnemonic,
            this.description,
            [
                { mnemonic: this.mnemonic, description: this.description },
                { mnemonic: this.mnemonicTrig, description: this.descriptionTrig },
            ],
        );
    }

    /** Consumes one chronological scalar price and returns the CG value. */
    public update(sample: number): number {
        const price = checkRealNumber(sample, 'price', true);
        if (Number.isNaN(price)) {
            return Number.NaN;
        }

        const trigger = this.count >= this.windowLength ? this.lastValue : Number.NaN;

        this.window[this.next] = price;
        this.next = (this.next + 1) % this.windowLength;
        if (this.count <= this.windowLength) {
            this.count += 1;
        }
        if (this.count < this.windowLength) {
            return Number.NaN;
        }

        let numerator = 0;
        let denominator = 0;
        for (let lag = 0; lag < this.windowLength; lag++) {
            const index = (this.next + this.windowLength - 1 - lag) % this.windowLength;
            const item = this.window[index];
            numerator += (1 + lag) * item;
            denominator += item;
        }

        if (denominator !== 0) {
            this.lastValue = -numerator / denominator + this.offset;
        }
        this.currentValue = this.lastValue;
        this.currentTrigger = trigger;
        return this.currentValue;
    }

    /** Updates the indicator given the next scalar sample. */
    public updateScalar(sample: Scalar): IndicatorOutput {
        return this.updateEntity(sample.time, sample.value);
    }

    /** Updates the indicator given the next bar sample. */
    public updateBar(sample: Bar): IndicatorOutput {
        return this.updateEntity(sample.time, this.barComponentFunc(sample));
    }

    /** Updates the indicator given the next quote sample. */
    public updateQuote(sample: Quote): IndicatorOutput {
        return this.updateEntity(sample.time, this.quoteComponentFunc(sample));
    }

    /** Updates the indicator given the next trade sample. */
    public updateTrade(sample: Trade): IndicatorOutput {
        return this.updateEntity(sample.time, this.tradeComponentFunc(sample));
    }

    private updateEntity(stamp: Date, sample: number): IndicatorOutput {
        const cog = this.update(sample);
        const trig = Number.isNaN(cog) ? Number.NaN : this.currentTrigger;
        return [
            new Scalar({ time: stamp, value: cog }),
            new Scalar({ time: stamp, value: trig }),
        ];
    }
}
