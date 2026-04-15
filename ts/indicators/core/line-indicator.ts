import { Bar } from '../../entities/bar';
import { BarComponent, DefaultBarComponent, barComponentValue } from '../../entities/bar-component';
import { Quote } from '../../entities/quote';
import { QuoteComponent, DefaultQuoteComponent, quoteComponentValue } from '../../entities/quote-component';
import { Scalar } from '../../entities/scalar';
import { Trade } from '../../entities/trade';
import { TradeComponent, DefaultTradeComponent, tradeComponentValue } from '../../entities/trade-component';
import { IndicatorMetadata } from './indicator-metadata';
import { IndicatorOutput } from './indicator-output';
import { Indicator } from './indicator';

/** Implements Indicator interface for a line indicator. */
export abstract class LineIndicator implements Indicator {

  protected barComponentFunc!: ((bar: Bar) => number);
  protected quoteComponentFunc!: ((quote: Quote) => number);
  protected tradeComponentFunc!: ((trade: Trade) => number);

  protected mnemonic!: string;
  protected description!: string;
  protected primed!: boolean;

  protected set barComponent(component: BarComponent | undefined) {
    if (component === undefined) {
      component = DefaultBarComponent;
    }
    this.barComponentFunc = barComponentValue(component);
  }

  protected set quoteComponent(component: QuoteComponent | undefined) {
    if (component === undefined) {
      component = DefaultQuoteComponent;
    }
    this.quoteComponentFunc = quoteComponentValue(component);
  }

  protected set tradeComponent(component: TradeComponent | undefined) {
    if (component === undefined) {
      component = DefaultTradeComponent;
    }
    this.tradeComponentFunc = tradeComponentValue(component);
  }

  /** Indicates whether an indicator is primed. */
  public isPrimed(): boolean { return this.primed; }

  /** Describes a requested output data of an indicator. */
  public abstract metadata(): IndicatorMetadata;

  /** Updates the value of the indicator given the next sample. */
  public abstract update(sample: number): number;

  /** Updates an indicator given the next scalar sample. */
  public updateScalar(sample: Scalar): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(sample.value);
    return [scalar];
  }

  /** Updates an indicator given the next bar sample. */
  public updateBar(sample: Bar): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(this.barComponentFunc(sample));
    return [scalar];
  }

  /** Updates an indicator given the next quote sample. */
  public updateQuote(sample: Quote): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(this.quoteComponentFunc(sample));
    return [scalar];
  }

  /** Updates an indicator given the next trade sample. */
  public updateTrade(sample: Trade): IndicatorOutput {
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = this.update(this.tradeComponentFunc(sample));
    return [scalar];
  }

  /** Updates the value of the line indicator given an array of scalar samples. */
  public updateScalars(array: Scalar[]): Scalar[] {
    const scalars: Scalar[] = [];
    for (const element of array) {
      const scalar = new Scalar();
      scalar.time = element.time;
      scalar.value = this.update(element.value);
      scalars.push(scalar);
    }

    return scalars;
  }

  /** Updates the value of the line indicator given an array of bar samples. */
  public updateBars(bars: Bar[]): Scalar[] {
    const scalars: Scalar[] = [];
    for (const bar of bars) {
      const scalar = new Scalar();
      scalar.time = bar.time;
      scalar.value = this.update(this.barComponentFunc(bar));
      scalars.push(scalar);
    }

    return scalars;
  }

  /** Updates the value of the line indicator given an array of quote samples. */
  public updateQuotes(quotes: Quote[]): Scalar[] {
    const scalars: Scalar[] = [];
    for (const quote of quotes) {
      const scalar = new Scalar();
      scalar.time = quote.time;
      scalar.value = this.update(this.quoteComponentFunc(quote));
      scalars.push(scalar);
    }

    return scalars;
  }

  /** Updates the value of the line indicator given an array of trade samples. */
  public updateTrades(trades: Trade[]): Scalar[] {
    const scalars: Scalar[] = [];
    for (const trade of trades) {
      const scalar = new Scalar();
      scalar.time = trade.time;
      scalar.value = this.update(this.tradeComponentFunc(trade));
      scalars.push(scalar);
    }

    return scalars;
  }
}
