import { Bar } from './bar';
import { Trade } from './trade';
import { Quote } from './quote';
import { Scalar } from './scalar';

/** A temporal entity is one of _Bar_, _Scalar_, _Trade_ or _Quote_. */
export type TemporalEntity = Bar | Scalar | Quote | Trade;

