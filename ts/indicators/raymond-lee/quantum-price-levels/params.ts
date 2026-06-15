import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';

/** Describes parameters to create an instance of the Quantum Price Levels indicator. */
export interface QuantumPriceLevelsParams {
    /**
     * The number of price-return ratios maintained in the sliding window.
     *
     * Priming requires lookback+1 prices. The value should be >= 2. The default value is 2048.
     */
    lookback?: number;

    /**
     * The number of quantum energy levels to compute (n = 0..numLevels-1).
     *
     * The value should be >= 1. The default value is 21.
     */
    numLevels?: number;

    /**
     * The number of histogram bins for the wavefunction distribution.
     *
     * The value should be >= 2. The default value is 100.
     */
    numBins?: number;

    /**
     * The empirical scaling constant in the NQPR formula (1 + scaleFactor*sigma*QPR).
     *
     * The value should be > 0. The default value is 0.21.
     */
    scaleFactor?: number;

    /**
     * A component of a bar to use when updating the indicator with a bar sample.
     *
     * If _undefined_, the bar component will have a default value (ClosePrice)
     * and will not be shown in the indicator mnemonic.
     */
    barComponent?: BarComponent;

    /**
     * A component of a quote to use when updating the indicator with a quote sample.
     *
     * If _undefined_, the quote component will have a default value and will not be shown in the indicator mnemonic.
     */
    quoteComponent?: QuoteComponent;

    /**
     * A component of a trade to use when updating the indicator with a trade sample.
     *
     * If _undefined_, the trade component will have a default value and will not be shown in the indicator mnemonic.
     */
    tradeComponent?: TradeComponent;
}

export function defaultParams(): QuantumPriceLevelsParams {
    return { lookback: 2048, numLevels: 21, numBins: 100, scaleFactor: 0.21 };
}
