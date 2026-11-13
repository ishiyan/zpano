import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { formatNumber } from '../corona/corona';

export interface CoronaSignalToNoiseRatioParams {
    rasterLength?: number;
    maxRasterValue?: number;
    minParameterValue?: number;
    maxParameterValue?: number;
    highPassFilterCutoff?: number;
    minimalPeriod?: number;
    maximalPeriod?: number;
    barComponent?: BarComponent;
    quoteComponent?: QuoteComponent;
    tradeComponent?: TradeComponent;
}

export function defaultParams(): CoronaSignalToNoiseRatioParams {
    return {
        rasterLength: 50,
        maxRasterValue: 20,
        minParameterValue: 1,
        maxParameterValue: 11,
        highPassFilterCutoff: 30,
        minimalPeriod: 6,
        maximalPeriod: 30,
    };
}

export function snrFlags(params?: CoronaSignalToNoiseRatioParams): string {
    const minPeriod = params?.minimalPeriod ?? 6;
    const maxPeriod = params?.maximalPeriod ?? 30;
    const parts: string[] = [];
    if (minPeriod !== 6) parts.push(`, pmin=${minPeriod}`);
    if (maxPeriod !== 30) parts.push(`, pmax=${maxPeriod}`);
    return parts.join('');
}

export function coronaSignalToNoiseRatioMnemonic(params?: CoronaSignalToNoiseRatioParams): string {
    const rasterLength = params?.rasterLength ?? 50;
    const maxRaster = params?.maxRasterValue ?? 20;
    const minParam = params?.minParameterValue ?? 1;
    const maxParam = params?.maxParameterValue ?? 11;
    const hpCutoff = params?.highPassFilterCutoff ?? 30;
    const bc = params?.barComponent ?? DefaultBarComponent;
    const qc = params?.quoteComponent ?? DefaultQuoteComponent;
    const tc = params?.tradeComponent ?? DefaultTradeComponent;
    const comp = componentTripleMnemonic(bc, qc, tc);
    const flags = snrFlags(params);
    return `csnr(${rasterLength}, ${formatNumber(maxRaster)}, ${formatNumber(minParam)}, ${formatNumber(maxParam)}, ${hpCutoff}${flags}${comp})`;
}
