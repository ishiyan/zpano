import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';

/** Describes parameters to create an instance of Ehlers' CenterOfGravityOscillator indicator. */
export interface CenterOfGravityOscillatorParams {
    /** Observation window length (integer >= 1). Default: 10. */
    length?: number;

    /** Zero-centered 2004 book baseline (-Num/Denom + (length+1)/2) when true; May 2002 article baseline (-Num/Denom) when false. Default: false. */
    centered?: boolean;

    /** Bar component used when updating with a bar sample; undefined defaults to Close and is omitted from the mnemonic. */
    barComponent?: BarComponent;

    /** Quote component used when updating with a quote sample; undefined defaults to Mid and is omitted from the mnemonic. */
    quoteComponent?: QuoteComponent;

    /** Trade component used when updating with a trade sample; undefined defaults to Price and is omitted from the mnemonic. */
    tradeComponent?: TradeComponent;
}

export type Params = CenterOfGravityOscillatorParams;

export function defaultParams(): CenterOfGravityOscillatorParams {
    return {
        length: 10,
        centered: false,
    };
}

function formatMnemonic(prefix: string, params?: CenterOfGravityOscillatorParams): string {
    const length = params === undefined || params.length === undefined ? 10 : params.length;
    const centered = params === undefined || params.centered === undefined ? false : params.centered;
    const barComp = params?.barComponent ?? DefaultBarComponent;
    const quoteComp = params?.quoteComponent ?? DefaultQuoteComponent;
    const tradeComp = params?.tradeComponent ?? DefaultTradeComponent;
    const centeredTag = centered ? ', centered' : '';
    const suffix = componentTripleMnemonic(barComp, quoteComp, tradeComp);
    return `${prefix}(${length}${centeredTag}${suffix})`;
}

export function centerOfGravityOscillatorMnemonic(params?: CenterOfGravityOscillatorParams): string {
    return formatMnemonic('cog', params);
}

export function centerOfGravityOscillatorTriggerMnemonic(params?: CenterOfGravityOscillatorParams): string {
    return formatMnemonic('cogTrig', params);
}
