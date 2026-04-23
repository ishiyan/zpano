/** Identifies the chart pane an indicator output is drawn on. */
export enum Pane {
  /** The primary price pane. */
  Price = 1,

  /** A dedicated sub-pane for this indicator. */
  Own,

  /** Drawn on the parent indicator's pane. */
  OverlayOnParent
}
