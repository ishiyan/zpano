import { Shape } from './outputs/shape/shape.js';
import { Pane } from './pane.js';
import { Role } from './role.js';

/** Classifies a single indicator output for charting / discovery. */
export interface OutputDescriptor {
  /** Integer representation of the output enumeration of a related indicator. */
  kind: number;

  /** The data shape of this output. */
  shape: Shape;

  /** The semantic role of this output. */
  role: Role;

  /** The chart pane on which this output is drawn. */
  pane: Pane;
}
