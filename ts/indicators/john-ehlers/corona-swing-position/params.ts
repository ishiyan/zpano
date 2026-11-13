import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { formatNumber } from '../corona/corona';

export interface CoronaSwingPositionParams {
    rasterLength?: number;
    maxRasterValue?: number;
    minParameterValue?: number;
    maxParameterValue?: number;
    highPassFilterCutoff?: number;
    minimalPeriod?: number;
    maximalPeriod?: number;
    leadWindow?: number;
    positionWindow?: number;
    barComponent?: BarComponent;
    quoteComponent?: QuoteComponent;
    tradeComponent?: TradeComponent;
}

export function defaultParams(): CoronaSwingPositionParams {
    return {
        rasterLength: 50,
        maxRasterValue: 20,
        minParameterValue: -5,
        maxParameterValue: 5,
        highPassFilterCutoff: 30,
        minimalPeriod: 6,
        maximalPeriod: 30,
        leadWindow: 51,
        positionWindow: 21,
    };
}

export function swingFlags(params?: CoronaSwingPositionParams): string {
    const minPeriod = params?.minimalPeriod ?? 6;
    const maxPeriod = params?.maximalPeriod ?? 30;
    const leadWindow = params?.leadWindow ?? 51;
    const positionWindow = params?.positionWindow ?? 21;
    const parts: string[] = [];
    if (minPeriod !== 6) parts.push(`, pmin=${minPeriod}`);
    if (maxPeriod !== 30) parts.push(`, pmax=${maxPeriod}`);
    if (leadWindow !== 51) parts.push(`, lead=${leadWindow}`);
    if (positionWindow !== 21) parts.push(`, pos=${positionWindow}`);
    return parts.join('');
}

export function coronaSwingPositionMnemonic(params?: CoronaSwingPositionParams): string {
    const rasterLength = params?.rasterLength ?? 50;
    const maxRaster = params?.maxRasterValue ?? 20;
    const minParam = params?.minParameterValue ?? -5;
    const maxParam = params?.maxParameterValue ?? 5;
    const hpCutoff = params?.highPassFilterCutoff ?? 30;
    const bc = params?.barComponent ?? DefaultBarComponent;
    const qc = params?.quoteComponent ?? DefaultQuoteComponent;
    const tc = params?.tradeComponent ?? DefaultTradeComponent;
    const comp = componentTripleMnemonic(bc, qc, tc);
    const flags = swingFlags(params);
    return `cswing(${rasterLength}, ${formatNumber(maxRaster)}, ${formatNumber(minParam)}, ${formatNumber(maxParam)}, ${hpCutoff}${flags}${comp})`;
}
