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
import { ZeroLagErrorCorrectingExponentialMovingAverageOutput } from './output';
import {
    ZeroLagErrorCorrectingExponentialMovingAverageLengthParams,
    ZeroLagErrorCorrectingExponentialMovingAverageParams,
    ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams,
    defaultLengthParams,
    defaultSmoothingFactorParams,
    equivalentLengthFromAlpha,
    zeroLagErrorCorrectingExponentialMovingAverageEmaMnemonic,
    zeroLagErrorCorrectingExponentialMovingAverageMnemonic,
} from './params';

export {
    ZeroLagErrorCorrectingExponentialMovingAverageOutput,
    zeroLagErrorCorrectingExponentialMovingAverageEmaMnemonic,
    zeroLagErrorCorrectingExponentialMovingAverageMnemonic,
};

const LEAST_ERROR_SENTINEL = 1_000_000.0;

function checkSample(sample: unknown): number {
    if (typeof sample !== 'number' || sample === Number.POSITIVE_INFINITY || sample === Number.NEGATIVE_INFINITY) {
        throw new Error('price must be a finite real or NaN');
    }
    return sample;
}

function checkEnumRange(value: unknown, minValue: number, maxValue: number, label: string): void {
    if (typeof value !== 'number' || !Number.isInteger(value) || value < minValue || value > maxValue) {
        throw new Error(`invalid ${label}`);
    }
}

/**
 * Ehlers and Way's Zero-Lag Error-Correcting Exponential Moving Average (ZECEMA).
 *
 * Follows the November 2010 TASC EasyLanguage listing ("Zero Lag (Well, Almost)"):
 * starting from explicit zero EMA and EC states, each accepted finite sample updates
 * the baseline EMA once and searches integer-tenths gains `index / 10` for inclusive
 * integers `[-gainLimit, +gainLimit]` against the frozen previous-bar EC using strict
 * improvement (`errorValue < leastError`) from the literal `1_000_000` sentinel.
 * Both EC (`Value`) and baseline EMA (`Ema`) outputs are available from the first
 * accepted finite observation.
 *
 * Reference:
 *
 * John F. Ehlers and Ric Way, "Zero Lag (Well, Almost)", TASC, Nov 2010, v28.11, pp. 30-35.
 */
export class ZeroLagErrorCorrectingExponentialMovingAverage implements Indicator {
    private readonly configuredLength: number;
    private readonly configuredGainLimit: number;
    private readonly configuredAlpha: number;
    private readonly complement: number;
    private readonly mnemonic: string;
    private readonly description: string;
    private readonly mnemonicEma: string;
    private readonly descriptionEma: string;
    private readonly barComponentFunc: (bar: Bar) => number;
    private readonly quoteComponentFunc: (quote: Quote) => number;
    private readonly tradeComponentFunc: (trade: Trade) => number;

    private count = 0;
    private emaState = 0;
    private ecState = 0;
    private bestGainState = 0;
    private leastErrorState = 0;
    private currentValue = Number.NaN;
    private currentEma = Number.NaN;

    public static default(): ZeroLagErrorCorrectingExponentialMovingAverage {
        return new ZeroLagErrorCorrectingExponentialMovingAverage(defaultLengthParams());
    }

    public static fromLength(
        params?: ZeroLagErrorCorrectingExponentialMovingAverageLengthParams,
    ): ZeroLagErrorCorrectingExponentialMovingAverage {
        return new ZeroLagErrorCorrectingExponentialMovingAverage(params ?? defaultLengthParams());
    }

    public static fromSmoothingFactor(
        params?: Partial<ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams>,
    ): ZeroLagErrorCorrectingExponentialMovingAverage {
        const defaults = defaultSmoothingFactorParams();
        if (params !== undefined && (params === null || typeof params !== 'object')) {
            throw new Error('invalid zero-lag error-correcting exponential moving average parameters');
        }
        const raw = (params ?? {}) as Record<string, unknown>;
        if ('length' in raw && raw.length !== undefined) {
            throw new Error('cannot specify both length and smoothingFactor');
        }
        if ('gainStep' in raw) {
            throw new Error('gainStep is not supported; gain step is fixed at 0.1');
        }
        const merged: ZeroLagErrorCorrectingExponentialMovingAverageSmoothingFactorParams = {
            smoothingFactor:
                params?.smoothingFactor === undefined ? defaults.smoothingFactor : params.smoothingFactor,
            gainLimit: params?.gainLimit === undefined ? defaults.gainLimit : params.gainLimit,
            barComponent: params?.barComponent,
            quoteComponent: params?.quoteComponent,
            tradeComponent: params?.tradeComponent,
        };
        return new ZeroLagErrorCorrectingExponentialMovingAverage(merged);
    }

    public static fromParams(
        params?: ZeroLagErrorCorrectingExponentialMovingAverageParams,
    ): ZeroLagErrorCorrectingExponentialMovingAverage {
        return new ZeroLagErrorCorrectingExponentialMovingAverage(params);
    }

    public constructor(params?: ZeroLagErrorCorrectingExponentialMovingAverageParams) {
        if (params !== undefined && (params === null || typeof params !== 'object')) {
            throw new Error('invalid zero-lag error-correcting exponential moving average parameters');
        }

        const raw = (params ?? defaultLengthParams()) as Record<string, unknown>;
        if ('gainStep' in raw) {
            throw new Error('gainStep is not supported; gain step is fixed at 0.1');
        }

        const hasLength = 'length' in raw && raw.length !== undefined;
        const hasSmoothingFactor = 'smoothingFactor' in raw && raw.smoothingFactor !== undefined;
        if (hasLength && hasSmoothingFactor) {
            throw new Error('cannot specify both length and smoothingFactor');
        }

        const gainLimitValue = raw.gainLimit === undefined ? 50 : raw.gainLimit;
        if (
            typeof gainLimitValue !== 'number' ||
            !Number.isInteger(gainLimitValue) ||
            !Number.isSafeInteger(gainLimitValue) ||
            gainLimitValue < 0
        ) {
            throw new Error('gainLimit must be a nonnegative safe integer in tenths');
        }

        if (raw.barComponent !== undefined) {
            checkEnumRange(raw.barComponent, BarComponent.Open, BarComponent.Average, 'barComponent');
        }
        if (raw.quoteComponent !== undefined) {
            checkEnumRange(raw.quoteComponent, QuoteComponent.Bid, QuoteComponent.SpreadBp, 'quoteComponent');
        }
        if (raw.tradeComponent !== undefined) {
            checkEnumRange(raw.tradeComponent, TradeComponent.Price, TradeComponent.Volume, 'tradeComponent');
        }

        const barComp = (raw.barComponent as BarComponent | undefined) ?? DefaultBarComponent;
        const quoteComp = (raw.quoteComponent as QuoteComponent | undefined) ?? DefaultQuoteComponent;
        const tradeComp = (raw.tradeComponent as TradeComponent | undefined) ?? DefaultTradeComponent;

        let resolvedLength: number;
        let resolvedAlpha: number;
        let resolvedComplement: number;
        let normalizedParams: ZeroLagErrorCorrectingExponentialMovingAverageParams;

        if (hasSmoothingFactor) {
            const sf = raw.smoothingFactor;
            if (typeof sf !== 'number' || !Number.isFinite(sf) || !(sf > 0 && sf <= 1)) {
                throw new Error('smoothingFactor must be in (0, 1]');
            }
            resolvedAlpha = sf;
            resolvedComplement = 1 - resolvedAlpha;
            if (!(resolvedComplement >= 0 && resolvedComplement < 1)) {
                throw new Error('smoothingFactor must yield representably stable EMA coefficients');
            }
            resolvedLength = equivalentLengthFromAlpha(resolvedAlpha);
            normalizedParams = {
                smoothingFactor: resolvedAlpha,
                gainLimit: gainLimitValue,
                barComponent: raw.barComponent as BarComponent | undefined,
                quoteComponent: raw.quoteComponent as QuoteComponent | undefined,
                tradeComponent: raw.tradeComponent as TradeComponent | undefined,
            };
        } else {
            const lengthValue = raw.length === undefined ? 20 : raw.length;
            if (typeof lengthValue !== 'number' || !Number.isInteger(lengthValue) || lengthValue < 1) {
                throw new Error('length must be an integer >= 1');
            }
            resolvedAlpha = 2 / (lengthValue + 1);
            resolvedComplement = 1 - resolvedAlpha;
            if (
                !Number.isFinite(resolvedAlpha) ||
                !(resolvedAlpha > 0 && resolvedAlpha <= 1) ||
                !(resolvedComplement >= 0 && resolvedComplement < 1)
            ) {
                throw new Error('length must yield representably stable EMA coefficients');
            }
            resolvedLength = lengthValue;
            normalizedParams = {
                length: resolvedLength,
                gainLimit: gainLimitValue,
                barComponent: raw.barComponent as BarComponent | undefined,
                quoteComponent: raw.quoteComponent as QuoteComponent | undefined,
                tradeComponent: raw.tradeComponent as TradeComponent | undefined,
            };
        }

        this.configuredLength = resolvedLength;
        this.configuredGainLimit = gainLimitValue;
        this.configuredAlpha = resolvedAlpha;
        this.complement = resolvedComplement;
        this.barComponentFunc = barComponentValue(barComp);
        this.quoteComponentFunc = quoteComponentValue(quoteComp);
        this.tradeComponentFunc = tradeComponentValue(tradeComp);

        this.mnemonic = zeroLagErrorCorrectingExponentialMovingAverageMnemonic(normalizedParams);
        this.mnemonicEma = zeroLagErrorCorrectingExponentialMovingAverageEmaMnemonic(normalizedParams);
        this.description = `Zero-lag Error-Correcting Exponential Moving Average ${this.mnemonic}`;
        this.descriptionEma = `Zero-lag Error-Correcting Exponential Moving Average baseline EMA ${this.mnemonicEma}`;

        this.reset();
    }

    /** Restores zero startup states and NaN public outputs while preserving configuration. */
    public reset(): void {
        this.count = 0;
        this.emaState = 0;
        this.ecState = 0;
        this.bestGainState = 0;
        this.leastErrorState = 0;
        this.currentValue = Number.NaN;
        this.currentEma = Number.NaN;
    }

    /** Returns the configured or equivalent EMA length. */
    public length(): number {
        return this.configuredLength;
    }

    /** Returns the configured integer gain limit in tenths. */
    public gainLimit(): number {
        return this.configuredGainLimit;
    }

    /** Returns the baseline EMA smoothing factor alpha. */
    public alpha(): number {
        return this.configuredAlpha;
    }

    /** Returns the number of accepted finite observations processed since construction or reset. */
    public samplesSeen(): number {
        return this.count;
    }

    /** Returns the warmup bar count (0 for ZECEMA's explicit zero-seeded recursion). */
    public warmupBars(): number {
        return 0;
    }

    /** Indicates whether at least one finite observation has been accepted. */
    public isPrimed(): boolean {
        return this.count > 0;
    }

    /** Returns the latest accepted EC value (NaN before priming). */
    public value(): number {
        return this.currentValue;
    }

    /** Returns the latest accepted baseline EMA value (NaN before priming). */
    public ema(): number {
        return this.currentEma;
    }

    /** Returns the selected gain from the latest accepted observation (NaN before priming). */
    public bestGain(): number {
        return this.count > 0 ? this.bestGainState : Number.NaN;
    }

    /** Returns the literal least-error diagnostic from the latest accepted observation (NaN before priming). */
    public leastError(): number {
        return this.count > 0 ? this.leastErrorState : Number.NaN;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.ZeroLagErrorCorrectingExponentialMovingAverage,
            this.mnemonic,
            this.description,
            [
                { mnemonic: this.mnemonic, description: this.description },
                { mnemonic: this.mnemonicEma, description: this.descriptionEma },
            ],
        );
    }

    /** Consumes one scalar price sample and returns the EC value. */
    public update(sample: number): number {
        const [ec] = this.step(sample);
        return ec;
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
        const [ec, ema] = this.step(sample);
        return [
            new Scalar({ time: stamp, value: ec }),
            new Scalar({ time: stamp, value: ema }),
        ];
    }

    private step(sample: number): [number, number] {
        const price = checkSample(sample);
        if (Number.isNaN(price)) {
            return [Number.NaN, Number.NaN];
        }

        const alpha = this.configuredAlpha;
        const complement = this.complement;
        const previousEc = this.ecState;

        const ema = alpha * price + complement * this.emaState;
        if (!Number.isFinite(ema)) {
            throw new Error('baseline EMA arithmetic overflowed');
        }

        let leastError = LEAST_ERROR_SENTINEL;
        let bestGain = this.bestGainState;
        for (let index = -this.configuredGainLimit; index <= this.configuredGainLimit; index++) {
            const gain = index / 10.0;
            const candidate = alpha * (ema + gain * (price - previousEc)) + complement * previousEc;
            const errorValue = Math.abs(price - candidate);
            if (!Number.isFinite(candidate) || !Number.isFinite(errorValue)) {
                throw new Error('gain-search arithmetic overflowed');
            }
            if (errorValue < leastError) {
                leastError = errorValue;
                bestGain = gain;
            }
        }

        const ec = alpha * (ema + bestGain * (price - previousEc)) + complement * previousEc;
        if (!Number.isFinite(ec)) {
            throw new Error('selected EC arithmetic overflowed');
        }

        this.emaState = ema;
        this.ecState = ec;
        this.bestGainState = bestGain;
        this.leastErrorState = leastError;
        this.count += 1;
        this.currentValue = ec;
        this.currentEma = ema;
        return [ec, ema];
    }
}
