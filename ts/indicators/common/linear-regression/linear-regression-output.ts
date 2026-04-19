/** Enumerates outputs of the LinearRegression indicator. */
export enum LinearRegressionOutput {
  /** The regression value at the last bar of the window: b + m*(period-1). */
  Value = 0,
  /** The time series forecast (one bar ahead): b + m*period. */
  Forecast = 1,
  /** The y-intercept of the regression line: b. */
  Intercept = 2,
  /** The slope of the regression line: m. */
  SlopeRad = 3,
  /** The slope expressed in degrees: atan(m)*180/pi. */
  SlopeDeg = 4,
}
