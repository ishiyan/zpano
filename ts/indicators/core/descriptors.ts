import { Adaptivity } from './adaptivity.js';
import { Descriptor } from './descriptor.js';
import { IndicatorIdentifier } from './indicator-identifier.js';
import { InputRequirement } from './input-requirement.js';
import { OutputDescriptor } from './output-descriptor.js';
import { Shape } from './outputs/shape/shape.js';
import { Pane } from './pane.js';
import { Role } from './role.js';
import { VolumeUsage } from './volume-usage.js';

// Helper to keep entries terse. Output Kind values are 0-based, mirroring each
// indicator package's TypeScript Output enumeration.
function out(kind: number, shape: Shape, role: Role, pane: Pane): OutputDescriptor {
  return { kind, shape, role, pane };
}

function desc(
  identifier: IndicatorIdentifier,
  family: string,
  adaptivity: Adaptivity,
  inputRequirement: InputRequirement,
  volumeUsage: VolumeUsage,
  outputs: OutputDescriptor[]
): Descriptor {
  return { identifier, family, adaptivity, inputRequirement, volumeUsage, outputs };
}

const S = Shape;
const R = Role;
const P = Pane;
const A = Adaptivity;
const I = InputRequirement;
const V = VolumeUsage;

/** Static registry of taxonomic descriptors for all implemented indicators. */
const descriptors: Map<IndicatorIdentifier, Descriptor> = new Map<IndicatorIdentifier, Descriptor>([
  [IndicatorIdentifier.SimpleMovingAverage, desc(IndicatorIdentifier.SimpleMovingAverage, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.WeightedMovingAverage, desc(IndicatorIdentifier.WeightedMovingAverage, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.TriangularMovingAverage, desc(IndicatorIdentifier.TriangularMovingAverage, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.ExponentialMovingAverage, desc(IndicatorIdentifier.ExponentialMovingAverage, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.DoubleExponentialMovingAverage, desc(IndicatorIdentifier.DoubleExponentialMovingAverage, 'Patrick Mulloy', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.TripleExponentialMovingAverage, desc(IndicatorIdentifier.TripleExponentialMovingAverage, 'Patrick Mulloy', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.T2ExponentialMovingAverage, desc(IndicatorIdentifier.T2ExponentialMovingAverage, 'Tim Tillson', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.T3ExponentialMovingAverage, desc(IndicatorIdentifier.T3ExponentialMovingAverage, 'Tim Tillson', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.KaufmanAdaptiveMovingAverage, desc(IndicatorIdentifier.KaufmanAdaptiveMovingAverage, 'Perry Kaufman', A.Adaptive, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.JurikMovingAverage, desc(IndicatorIdentifier.JurikMovingAverage, 'Mark Jurik', A.Adaptive, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.MesaAdaptiveMovingAverage, desc(IndicatorIdentifier.MesaAdaptiveMovingAverage, 'John Ehlers', A.Adaptive, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.Smoother, P.Price),
    out(1, S.Scalar, R.Smoother, P.Price),
    out(2, S.Band, R.Envelope, P.Price),
  ])],
  [IndicatorIdentifier.FractalAdaptiveMovingAverage, desc(IndicatorIdentifier.FractalAdaptiveMovingAverage, 'John Ehlers', A.Adaptive, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.Smoother, P.Price),
    out(1, S.Scalar, R.FractalDimension, P.Own),
  ])],
  [IndicatorIdentifier.DominantCycle, desc(IndicatorIdentifier.DominantCycle, 'John Ehlers', A.Adaptive, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.CyclePeriod, P.Own),
    out(1, S.Scalar, R.CyclePeriod, P.Own),
    out(2, S.Scalar, R.CyclePhase, P.Own),
  ])],
  [IndicatorIdentifier.Momentum, desc(IndicatorIdentifier.Momentum, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.RateOfChange, desc(IndicatorIdentifier.RateOfChange, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.RateOfChangePercent, desc(IndicatorIdentifier.RateOfChangePercent, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.RelativeStrengthIndex, desc(IndicatorIdentifier.RelativeStrengthIndex, 'Welles Wilder', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.BoundedOscillator, P.Own)])],
  [IndicatorIdentifier.ChandeMomentumOscillator, desc(IndicatorIdentifier.ChandeMomentumOscillator, 'Tushar Chande', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.BoundedOscillator, P.Own)])],
  [IndicatorIdentifier.BollingerBands, desc(IndicatorIdentifier.BollingerBands, 'John Bollinger', A.Static, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.Envelope, P.Price),
    out(1, S.Scalar, R.Smoother, P.Price),
    out(2, S.Scalar, R.Envelope, P.Price),
    out(3, S.Scalar, R.Volatility, P.Own),
    out(4, S.Scalar, R.BoundedOscillator, P.Own),
    out(5, S.Band, R.Envelope, P.Price),
  ])],
  [IndicatorIdentifier.Variance, desc(IndicatorIdentifier.Variance, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Volatility, P.Own)])],
  [IndicatorIdentifier.StandardDeviation, desc(IndicatorIdentifier.StandardDeviation, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Volatility, P.Own)])],
  [IndicatorIdentifier.GoertzelSpectrum, desc(IndicatorIdentifier.GoertzelSpectrum, 'Custom', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Heatmap, R.Spectrum, P.Own)])],
  [IndicatorIdentifier.CenterOfGravityOscillator, desc(IndicatorIdentifier.CenterOfGravityOscillator, 'John Ehlers', A.Static, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.Oscillator, P.Own),
    out(1, S.Scalar, R.Signal, P.Own),
  ])],
  [IndicatorIdentifier.CyberCycle, desc(IndicatorIdentifier.CyberCycle, 'John Ehlers', A.Static, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.Oscillator, P.Own),
    out(1, S.Scalar, R.Signal, P.Own),
  ])],
  [IndicatorIdentifier.InstantaneousTrendLine, desc(IndicatorIdentifier.InstantaneousTrendLine, 'John Ehlers', A.Static, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.Smoother, P.Price),
    out(1, S.Scalar, R.Signal, P.Price),
  ])],
  [IndicatorIdentifier.SuperSmoother, desc(IndicatorIdentifier.SuperSmoother, 'John Ehlers', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.ZeroLagExponentialMovingAverage, desc(IndicatorIdentifier.ZeroLagExponentialMovingAverage, 'John Ehlers', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.ZeroLagErrorCorrectingExponentialMovingAverage, desc(IndicatorIdentifier.ZeroLagErrorCorrectingExponentialMovingAverage, 'John Ehlers', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Smoother, P.Price)])],
  [IndicatorIdentifier.RoofingFilter, desc(IndicatorIdentifier.RoofingFilter, 'John Ehlers', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.TrueRange, desc(IndicatorIdentifier.TrueRange, 'Welles Wilder', A.Static, I.BarInput, V.NoVolume, [out(0, S.Scalar, R.Volatility, P.Own)])],
  [IndicatorIdentifier.AverageTrueRange, desc(IndicatorIdentifier.AverageTrueRange, 'Welles Wilder', A.Static, I.BarInput, V.NoVolume, [out(0, S.Scalar, R.Volatility, P.Own)])],
  [IndicatorIdentifier.NormalizedAverageTrueRange, desc(IndicatorIdentifier.NormalizedAverageTrueRange, 'Welles Wilder', A.Static, I.BarInput, V.NoVolume, [out(0, S.Scalar, R.Volatility, P.Own)])],
  [IndicatorIdentifier.DirectionalMovementMinus, desc(IndicatorIdentifier.DirectionalMovementMinus, 'Welles Wilder', A.Static, I.BarInput, V.NoVolume, [out(0, S.Scalar, R.Directional, P.Own)])],
  [IndicatorIdentifier.DirectionalMovementPlus, desc(IndicatorIdentifier.DirectionalMovementPlus, 'Welles Wilder', A.Static, I.BarInput, V.NoVolume, [out(0, S.Scalar, R.Directional, P.Own)])],
  [IndicatorIdentifier.DirectionalIndicatorMinus, desc(IndicatorIdentifier.DirectionalIndicatorMinus, 'Welles Wilder', A.Static, I.BarInput, V.NoVolume, [
    out(0, S.Scalar, R.Directional, P.Own),
    out(1, S.Scalar, R.Directional, P.Own),
    out(2, S.Scalar, R.Volatility, P.Own),
    out(3, S.Scalar, R.Volatility, P.Own),
  ])],
  [IndicatorIdentifier.DirectionalIndicatorPlus, desc(IndicatorIdentifier.DirectionalIndicatorPlus, 'Welles Wilder', A.Static, I.BarInput, V.NoVolume, [
    out(0, S.Scalar, R.Directional, P.Own),
    out(1, S.Scalar, R.Directional, P.Own),
    out(2, S.Scalar, R.Volatility, P.Own),
    out(3, S.Scalar, R.Volatility, P.Own),
  ])],
  [IndicatorIdentifier.DirectionalMovementIndex, desc(IndicatorIdentifier.DirectionalMovementIndex, 'Welles Wilder', A.Static, I.BarInput, V.NoVolume, [
    out(0, S.Scalar, R.BoundedOscillator, P.Own),
    out(1, S.Scalar, R.Directional, P.Own),
    out(2, S.Scalar, R.Directional, P.Own),
    out(3, S.Scalar, R.Directional, P.Own),
    out(4, S.Scalar, R.Directional, P.Own),
    out(5, S.Scalar, R.Volatility, P.Own),
    out(6, S.Scalar, R.Volatility, P.Own),
  ])],
  [IndicatorIdentifier.AverageDirectionalMovementIndex, desc(IndicatorIdentifier.AverageDirectionalMovementIndex, 'Welles Wilder', A.Static, I.BarInput, V.NoVolume, [
    out(0, S.Scalar, R.BoundedOscillator, P.Own),
    out(1, S.Scalar, R.BoundedOscillator, P.Own),
    out(2, S.Scalar, R.Directional, P.Own),
    out(3, S.Scalar, R.Directional, P.Own),
    out(4, S.Scalar, R.Directional, P.Own),
    out(5, S.Scalar, R.Directional, P.Own),
    out(6, S.Scalar, R.Volatility, P.Own),
    out(7, S.Scalar, R.Volatility, P.Own),
  ])],
  [IndicatorIdentifier.AverageDirectionalMovementIndexRating, desc(IndicatorIdentifier.AverageDirectionalMovementIndexRating, 'Welles Wilder', A.Static, I.BarInput, V.NoVolume, [
    out(0, S.Scalar, R.BoundedOscillator, P.Own),
    out(1, S.Scalar, R.BoundedOscillator, P.Own),
    out(2, S.Scalar, R.BoundedOscillator, P.Own),
    out(3, S.Scalar, R.Directional, P.Own),
    out(4, S.Scalar, R.Directional, P.Own),
    out(5, S.Scalar, R.Directional, P.Own),
    out(6, S.Scalar, R.Directional, P.Own),
    out(7, S.Scalar, R.Volatility, P.Own),
    out(8, S.Scalar, R.Volatility, P.Own),
  ])],
  [IndicatorIdentifier.WilliamsPercentR, desc(IndicatorIdentifier.WilliamsPercentR, 'Larry Williams', A.Static, I.BarInput, V.NoVolume, [out(0, S.Scalar, R.BoundedOscillator, P.Own)])],
  [IndicatorIdentifier.PercentagePriceOscillator, desc(IndicatorIdentifier.PercentagePriceOscillator, 'Gerald Appel', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.AbsolutePriceOscillator, desc(IndicatorIdentifier.AbsolutePriceOscillator, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.CommodityChannelIndex, desc(IndicatorIdentifier.CommodityChannelIndex, 'Donald Lambert', A.Static, I.BarInput, V.NoVolume, [out(0, S.Scalar, R.BoundedOscillator, P.Own)])],
  [IndicatorIdentifier.MoneyFlowIndex, desc(IndicatorIdentifier.MoneyFlowIndex, 'Gene Quong', A.Static, I.BarInput, V.AggregateBarVolume, [out(0, S.Scalar, R.BoundedOscillator, P.Own)])],
  [IndicatorIdentifier.OnBalanceVolume, desc(IndicatorIdentifier.OnBalanceVolume, 'Joseph Granville', A.Static, I.BarInput, V.AggregateBarVolume, [out(0, S.Scalar, R.VolumeFlow, P.Own)])],
  [IndicatorIdentifier.BalanceOfPower, desc(IndicatorIdentifier.BalanceOfPower, 'Igor Livshin', A.Static, I.BarInput, V.NoVolume, [out(0, S.Scalar, R.BoundedOscillator, P.Own)])],
  [IndicatorIdentifier.RateOfChangeRatio, desc(IndicatorIdentifier.RateOfChangeRatio, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.PearsonsCorrelationCoefficient, desc(IndicatorIdentifier.PearsonsCorrelationCoefficient, 'Common', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Correlation, P.Own)])],
  [IndicatorIdentifier.LinearRegression, desc(IndicatorIdentifier.LinearRegression, 'Common', A.Static, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.Smoother, P.Price),
    out(1, S.Scalar, R.Smoother, P.Price),
    out(2, S.Scalar, R.Smoother, P.Price),
    out(3, S.Scalar, R.Oscillator, P.Own),
    out(4, S.Scalar, R.Oscillator, P.Own),
  ])],
  [IndicatorIdentifier.UltimateOscillator, desc(IndicatorIdentifier.UltimateOscillator, 'Larry Williams', A.Static, I.BarInput, V.NoVolume, [out(0, S.Scalar, R.BoundedOscillator, P.Own)])],
  [IndicatorIdentifier.StochasticRelativeStrengthIndex, desc(IndicatorIdentifier.StochasticRelativeStrengthIndex, 'Tushar Chande', A.Static, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.BoundedOscillator, P.Own),
    out(1, S.Scalar, R.Signal, P.Own),
  ])],
  [IndicatorIdentifier.Stochastic, desc(IndicatorIdentifier.Stochastic, 'George Lane', A.Static, I.BarInput, V.NoVolume, [
    out(0, S.Scalar, R.BoundedOscillator, P.Own),
    out(1, S.Scalar, R.BoundedOscillator, P.Own),
    out(2, S.Scalar, R.Signal, P.Own),
  ])],
  [IndicatorIdentifier.Aroon, desc(IndicatorIdentifier.Aroon, 'Tushar Chande', A.Static, I.BarInput, V.NoVolume, [
    out(0, S.Scalar, R.BoundedOscillator, P.Own),
    out(1, S.Scalar, R.BoundedOscillator, P.Own),
    out(2, S.Scalar, R.Oscillator, P.Own),
  ])],
  [IndicatorIdentifier.AdvanceDecline, desc(IndicatorIdentifier.AdvanceDecline, 'Marc Chaikin', A.Static, I.BarInput, V.AggregateBarVolume, [out(0, S.Scalar, R.VolumeFlow, P.Own)])],
  [IndicatorIdentifier.AdvanceDeclineOscillator, desc(IndicatorIdentifier.AdvanceDeclineOscillator, 'Marc Chaikin', A.Static, I.BarInput, V.AggregateBarVolume, [out(0, S.Scalar, R.VolumeFlow, P.Own)])],
  [IndicatorIdentifier.ParabolicStopAndReverse, desc(IndicatorIdentifier.ParabolicStopAndReverse, 'Welles Wilder', A.Static, I.BarInput, V.NoVolume, [out(0, S.Scalar, R.Overlay, P.Price)])],
  [IndicatorIdentifier.TripleExponentialMovingAverageOscillator, desc(IndicatorIdentifier.TripleExponentialMovingAverageOscillator, 'Jack Hutson', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.BollingerBandsTrend, desc(IndicatorIdentifier.BollingerBandsTrend, 'John Bollinger', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.MovingAverageConvergenceDivergence, desc(IndicatorIdentifier.MovingAverageConvergenceDivergence, 'Gerald Appel', A.Static, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.Oscillator, P.Own),
    out(1, S.Scalar, R.Signal, P.Own),
    out(2, S.Scalar, R.Histogram, P.Own),
  ])],
  [IndicatorIdentifier.SineWave, desc(IndicatorIdentifier.SineWave, 'John Ehlers', A.Adaptive, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.Oscillator, P.Own),
    out(1, S.Scalar, R.Signal, P.Own),
    out(2, S.Band, R.Envelope, P.Own),
    out(3, S.Scalar, R.CyclePeriod, P.Own),
    out(4, S.Scalar, R.CyclePhase, P.Own),
  ])],
  [IndicatorIdentifier.HilbertTransformerInstantaneousTrendLine, desc(IndicatorIdentifier.HilbertTransformerInstantaneousTrendLine, 'John Ehlers', A.Adaptive, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.Smoother, P.Price),
    out(1, S.Scalar, R.CyclePeriod, P.Own),
  ])],
  [IndicatorIdentifier.TrendCycleMode, desc(IndicatorIdentifier.TrendCycleMode, 'John Ehlers', A.Adaptive, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.RegimeFlag, P.Own),
    out(1, S.Scalar, R.RegimeFlag, P.Own),
    out(2, S.Scalar, R.RegimeFlag, P.Own),
    out(3, S.Scalar, R.Smoother, P.Price),
    out(4, S.Scalar, R.Oscillator, P.Own),
    out(5, S.Scalar, R.Signal, P.Own),
    out(6, S.Scalar, R.CyclePeriod, P.Own),
    out(7, S.Scalar, R.CyclePhase, P.Own),
  ])],
  [IndicatorIdentifier.CoronaSpectrum, desc(IndicatorIdentifier.CoronaSpectrum, 'John Ehlers', A.Adaptive, I.ScalarInput, V.NoVolume, [
    out(0, S.Heatmap, R.Spectrum, P.Own),
    out(1, S.Scalar, R.CyclePeriod, P.Own),
    out(2, S.Scalar, R.CyclePeriod, P.Own),
  ])],
  [IndicatorIdentifier.CoronaSignalToNoiseRatio, desc(IndicatorIdentifier.CoronaSignalToNoiseRatio, 'John Ehlers', A.Adaptive, I.ScalarInput, V.NoVolume, [
    out(0, S.Heatmap, R.Spectrum, P.Own),
    out(1, S.Scalar, R.BoundedOscillator, P.Own),
  ])],
  [IndicatorIdentifier.CoronaSwingPosition, desc(IndicatorIdentifier.CoronaSwingPosition, 'John Ehlers', A.Adaptive, I.ScalarInput, V.NoVolume, [
    out(0, S.Heatmap, R.Spectrum, P.Own),
    out(1, S.Scalar, R.BoundedOscillator, P.Own),
  ])],
  [IndicatorIdentifier.CoronaTrendVigor, desc(IndicatorIdentifier.CoronaTrendVigor, 'John Ehlers', A.Adaptive, I.ScalarInput, V.NoVolume, [
    out(0, S.Heatmap, R.Spectrum, P.Own),
    out(1, S.Scalar, R.Oscillator, P.Own),
  ])],
  [IndicatorIdentifier.AdaptiveTrendAndCycleFilter, desc(IndicatorIdentifier.AdaptiveTrendAndCycleFilter, 'Vladimir Kravchuk', A.Adaptive, I.ScalarInput, V.NoVolume, [
    out(0, S.Scalar, R.Smoother, P.Price),
    out(1, S.Scalar, R.Smoother, P.Price),
    out(2, S.Scalar, R.Smoother, P.Price),
    out(3, S.Scalar, R.Smoother, P.Price),
    out(4, S.Scalar, R.Smoother, P.Price),
    out(5, S.Scalar, R.Oscillator, P.Own),
    out(6, S.Scalar, R.Oscillator, P.Own),
    out(7, S.Scalar, R.Oscillator, P.Own),
  ])],
  [IndicatorIdentifier.MaximumEntropySpectrum, desc(IndicatorIdentifier.MaximumEntropySpectrum, 'Custom', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Heatmap, R.Spectrum, P.Own)])],
  [IndicatorIdentifier.DiscreteFourierTransformSpectrum, desc(IndicatorIdentifier.DiscreteFourierTransformSpectrum, 'John Ehlers', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Heatmap, R.Spectrum, P.Own)])],
  [IndicatorIdentifier.CombBandPassSpectrum, desc(IndicatorIdentifier.CombBandPassSpectrum, 'John Ehlers', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Heatmap, R.Spectrum, P.Own)])],
  [IndicatorIdentifier.AutoCorrelationIndicator, desc(IndicatorIdentifier.AutoCorrelationIndicator, 'John Ehlers', A.Static, I.ScalarInput, V.NoVolume, [out(0, S.Heatmap, R.Correlation, P.Own)])],
  [IndicatorIdentifier.AutoCorrelationPeriodogram, desc(IndicatorIdentifier.AutoCorrelationPeriodogram, 'John Ehlers', A.Adaptive, I.ScalarInput, V.NoVolume, [out(0, S.Heatmap, R.Spectrum, P.Own)])],
  [IndicatorIdentifier.JurikRelativeTrendStrengthIndex, desc(IndicatorIdentifier.JurikRelativeTrendStrengthIndex, 'Mark Jurik', A.Static, I.ScalarInput, V.NoVolume, [out(1, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.JurikCompositeFractalBehaviorIndex, desc(IndicatorIdentifier.JurikCompositeFractalBehaviorIndex, 'Mark Jurik', A.Static, I.ScalarInput, V.NoVolume, [out(1, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.JurikZeroLagVelocity, desc(IndicatorIdentifier.JurikZeroLagVelocity, 'Mark Jurik', A.Adaptive, I.ScalarInput, V.NoVolume, [out(1, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.JurikDirectionalMovementIndex, desc(IndicatorIdentifier.JurikDirectionalMovementIndex, 'Mark Jurik', A.Adaptive, I.BarInput, V.NoVolume, [
    out(1, S.Scalar, R.Oscillator, P.Own),
    out(2, S.Scalar, R.Oscillator, P.Own),
    out(3, S.Scalar, R.Oscillator, P.Own),
  ])],
  [IndicatorIdentifier.JurikAdaptiveRelativeTrendStrengthIndex, desc(IndicatorIdentifier.JurikAdaptiveRelativeTrendStrengthIndex, 'Mark Jurik', A.Adaptive, I.ScalarInput, V.NoVolume, [out(1, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.JurikAdaptiveZeroLagVelocity, desc(IndicatorIdentifier.JurikAdaptiveZeroLagVelocity, 'Mark Jurik', A.Adaptive, I.ScalarInput, V.NoVolume, [out(1, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.JurikCommodityChannelIndex, desc(IndicatorIdentifier.JurikCommodityChannelIndex, 'Mark Jurik', A.Adaptive, I.ScalarInput, V.NoVolume, [out(1, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.JurikFractalAdaptiveZeroLagVelocity, desc(IndicatorIdentifier.JurikFractalAdaptiveZeroLagVelocity, 'Mark Jurik', A.Adaptive, I.ScalarInput, V.NoVolume, [out(1, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.JurikTurningPointOscillator, desc(IndicatorIdentifier.JurikTurningPointOscillator, 'Mark Jurik', A.Static, I.ScalarInput, V.NoVolume, [out(1, S.Scalar, R.Oscillator, P.Own)])],
  [IndicatorIdentifier.JurikWaveletSampler, desc(IndicatorIdentifier.JurikWaveletSampler, 'Mark Jurik', A.Static, I.ScalarInput, V.NoVolume, [out(1, S.Scalar, R.Smoother, P.Price)])],
]);

/**
 * Returns the taxonomic descriptor for the given indicator identifier, or
 * `undefined` if no descriptor is registered for the identifier.
 */
export function descriptorOf(id: IndicatorIdentifier): Descriptor | undefined {
  return descriptors.get(id);
}

/** Returns a copy of the full descriptor registry. */
export function getDescriptors(): Map<IndicatorIdentifier, Descriptor> {
  return new Map(descriptors);
}
