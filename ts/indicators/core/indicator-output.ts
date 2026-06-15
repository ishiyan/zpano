import { Scalar } from '../../entities/scalar';
import { Band } from './outputs/band';
import { Heatmap } from './outputs/heatmap';
import { Polyline } from './outputs/polyline';
import { Levels } from './outputs/levels';

/** Defines indicator output. */
export type IndicatorOutput = (Scalar | Band | Heatmap | Polyline | Levels)[];
