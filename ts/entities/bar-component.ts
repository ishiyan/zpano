import { Bar } from './bar';

/** Enumerates price components of a _Bar_. */
export enum BarComponent {
  /** The opening price. */
  Open,

  /** The highest price. */
  High,

  /** The lowest price. */
  Low,

  /** The closing price. */
  Close,

  /** The volume. */
  Volume,

  /** The median price, calculated as _(high + low) / 2_. */
  Median,

  /** The typical price, calculated as _(high + low + close) / 3_. */
  Typical,

  /** The weighted price, calculated as _(high + low + 2*close) / 4_. */
  Weighted,

  /** The average price, calculated as _(open + high + low + close) / 4_. */
  Average,
}

/** The default bar component used when none is specified. */
export const DefaultBarComponent = BarComponent.Close;

/** Function for calculating a price component of a _Bar_. */
export const barComponentValue = (component: BarComponent): (bar: Bar) => number => {
  switch (component) {
    case BarComponent.Open:
      return (b: Bar) => b.open;
    case BarComponent.High:
      return (b: Bar) => b.high;
    case BarComponent.Low:
      return (b: Bar) => b.low;
    case BarComponent.Close:
      return (b: Bar) => b.close;
    case BarComponent.Volume:
      return (b: Bar) => b.volume;
    case BarComponent.Median:
      return (b: Bar) => b.median();
    case BarComponent.Typical:
      return (b: Bar) => b.typical();
    case BarComponent.Weighted:
      return (b: Bar) => b.weighted();
    case BarComponent.Average:
      return (b: Bar) => b.average();
    default: // Default to close.
      return (b: Bar) => b.close;
  }
};

/** The mnemonic of a price component of a _Bar_. */
export const barComponentMnemonic = (component: BarComponent): string => {
  switch (component) {
    case BarComponent.Open:
      return 'o';
    case BarComponent.High:
      return 'h';
    case BarComponent.Low:
      return 'l';
    case BarComponent.Close:
      return 'c';
    case BarComponent.Volume:
      return 'v';
    case BarComponent.Median:
      return 'hl/2';
    case BarComponent.Typical:
      return 'hlc/3';
    case BarComponent.Weighted:
      return 'hlcc/4';
    case BarComponent.Average:
      return 'ohlc/4';
    default:
      return '??';
  }
};
