import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { formatNumber } from '../corona/corona';

export interface CoronaTrendVigorParams {
    rasterLength?: number;
    maxRasterValue?: number;
    minParameterValue?: number;
    maxParameterValue?: number;
    highPassFilterCutoff?: number;
    minimalPeriod?: number;
    maximalPeriod?: number;
    lookbackOffset?: number;
    barComponent?: BarComponent;
    quoteComponent?: QuoteComponent;
    tradeComponent?: TradeComponent;
}

export function defaultParams(): CoronaTrendVigorParams {
    return {
        rasterLength: 50,
        maxRasterValue: 20,
        minParameterValue: -10,
        maxParameterValue: 10,
        highPassFilterCutoff: 30,
        minimalPeriod: 6,
        maximalPeriod: 30,
        lookbackOffset: 0,
    };
}

export function trendFlags(params?: CoronaTrendVigorParams): string {
    const minPeriod = params?.minimalPeriod ?? 6;
    const maxPeriod = params?.maximalPeriod ?? 30;
    const lookbackOffset = params?.lookbackOffset ?? 0;
    const parts: string[] = [];
    if (minPeriod !== 6) parts.push(`, pmin=${minPeriod}`);
    if (maxPeriod !== 30) parts.push(`, pmax=${maxPeriod}`);
    if (lookbackOffset !== 0) parts.push(`, offset=${lookbackOffset}`);
    return parts.join('');
}

export function coronaTrendVigorMnemonic(params?: CoronaTrendVigorParams): string {
    const rasterLength = params?.rasterLength ?? 50;
    const maxRaster = params?.maxRasterValue ?? 20;
    const minParam = params?.minParameterValue ?? -10;
    const maxParam = params?.maxParameterValue ?? 10;
    const hpCutoff = params?.highPassFilterCutoff ?? 30;
    const bc = params?.barComponent ?? DefaultBarComponent;
    const qc = params?.quoteComponent ?? DefaultQuoteComponent;
    const tc = params?.tradeComponent ?? DefaultTradeComponent;
    const comp = componentTripleMnemonic(bc, qc, tc);
    const flags = trendFlags(params);
    return `ctv(${rasterLength}, ${formatNumber(maxRaster)}, ${formatNumber(minParam)}, ${formatNumber(maxParam)}, ${hpCutoff}${flags}${comp})`;
}
