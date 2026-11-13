import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { formatNumber } from '../corona/corona';

export interface CoronaSpectrumParams {
    minRasterValue?: number;
    maxRasterValue?: number;
    minParameterValue?: number;
    maxParameterValue?: number;
    highPassFilterCutoff?: number;
    barComponent?: BarComponent;
    quoteComponent?: QuoteComponent;
    tradeComponent?: TradeComponent;
}

export function defaultParams(): CoronaSpectrumParams {
    return {
        minRasterValue: 6,
        maxRasterValue: 20,
        minParameterValue: 6,
        maxParameterValue: 30,
        highPassFilterCutoff: 30,
    };
}

export function coronaSpectrumMnemonic(params?: CoronaSpectrumParams): string {
    const minRaster = params?.minRasterValue ?? 6;
    const maxRaster = params?.maxRasterValue ?? 20;
    const minPeriod = Math.ceil(params?.minParameterValue ?? 6);
    const maxPeriod = Math.floor(params?.maxParameterValue ?? 30);
    const hpCutoff = params?.highPassFilterCutoff ?? 30;
    const bc = params?.barComponent ?? DefaultBarComponent;
    const qc = params?.quoteComponent ?? DefaultQuoteComponent;
    const tc = params?.tradeComponent ?? DefaultTradeComponent;
    const comp = componentTripleMnemonic(bc, qc, tc);
    return `cspect(${formatNumber(minRaster)}, ${formatNumber(maxRaster)}, ${minPeriod}, ${maxPeriod}, ${hpCutoff}${comp})`;
}
