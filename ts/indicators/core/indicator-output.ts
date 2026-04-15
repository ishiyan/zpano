import { Scalar } from '../../entities/scalar';
import { Band } from './outputs/band';
import { Heatmap } from './outputs/heatmap';

/** Defines indicator output. */
export type IndicatorOutput = (Scalar | Band | Heatmap)[];
