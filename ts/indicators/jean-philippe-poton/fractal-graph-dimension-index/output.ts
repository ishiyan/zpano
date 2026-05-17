/** Enumerates outputs of the Fractal Graph Dimension Index indicator. */
export enum FractalGraphDimensionIndexOutput {

    /** The fractal graph dimension value. */
    FgdiValue = 0,

    /** The upper band (fgdi + stddev). */
    Upper = 1,

    /** The lower band (fgdi - stddev). */
    Lower = 2,

    /** The standard deviation of the dimension estimate. */
    Stddev = 3,

    /** The lower/upper band pair. */
    Band = 4,
}
