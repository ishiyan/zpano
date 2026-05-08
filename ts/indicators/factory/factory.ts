/**
 * Factory that maps an {@link IndicatorIdentifier} and a plain-object parameter
 * bag to a fully constructed {@link Indicator} instance, so callers don't need
 * to import every indicator module directly.
 *
 * For indicators with **Length / SmoothingFactor** constructor variants the
 * factory auto-detects which to use: if the params object contains a
 * `smoothingFactor` key (or `fastLimitSmoothingFactor`/`slowLimitSmoothingFactor`
 * for MAMA, `fastestSmoothingFactor`/`slowestSmoothingFactor` for KAMA) the
 * SmoothingFactor variant is used; otherwise the Length variant.
 *
 * For indicators with **default / fromParams** static factories: if params is
 * `undefined` or an empty object `{}`, `default()` is called; otherwise
 * `fromParams(params)`.
 *
 * @module
 */

import { type Indicator } from '../core/indicator.js';
import { IndicatorIdentifier } from '../core/indicator-identifier.js';

// ── common ──────────────────────────────────────────────────────────────────
import { SimpleMovingAverage } from '../common/simple-moving-average/simple-moving-average.js';
import { defaultParams as defaultSmaParams } from '../common/simple-moving-average/params.js';
import { WeightedMovingAverage } from '../common/weighted-moving-average/weighted-moving-average.js';
import { defaultParams as defaultWmaParams } from '../common/weighted-moving-average/params.js';
import { TriangularMovingAverage } from '../common/triangular-moving-average/triangular-moving-average.js';
import { defaultParams as defaultTrimaParams } from '../common/triangular-moving-average/params.js';
import { ExponentialMovingAverage } from '../common/exponential-moving-average/exponential-moving-average.js';
import { defaultLengthParams as defaultEmaLengthParams } from '../common/exponential-moving-average/params.js';
import { Momentum } from '../common/momentum/momentum.js';
import { defaultParams as defaultMomentumParams } from '../common/momentum/params.js';
import { RateOfChange } from '../common/rate-of-change/rate-of-change.js';
import { defaultParams as defaultRocParams } from '../common/rate-of-change/params.js';
import { RateOfChangePercent } from '../common/rate-of-change-percent/rate-of-change-percent.js';
import { defaultParams as defaultRocpParams } from '../common/rate-of-change-percent/params.js';
import { RateOfChangeRatio } from '../common/rate-of-change-ratio/rate-of-change-ratio.js';
import { defaultParams as defaultRocrParams } from '../common/rate-of-change-ratio/params.js';
import { Variance } from '../common/variance/variance.js';
import { defaultParams as defaultVarianceParams } from '../common/variance/params.js';
import { StandardDeviation } from '../common/standard-deviation/standard-deviation.js';
import { defaultParams as defaultStdDevParams } from '../common/standard-deviation/params.js';
import { LinearRegression } from '../common/linear-regression/linear-regression.js';
import { defaultParams as defaultLinregParams } from '../common/linear-regression/params.js';
import { PearsonsCorrelationCoefficient } from '../common/pearsons-correlation-coefficient/pearsons-correlation-coefficient.js';
import { defaultParams as defaultCorrelParams } from '../common/pearsons-correlation-coefficient/params.js';
import { AbsolutePriceOscillator } from '../common/absolute-price-oscillator/absolute-price-oscillator.js';
import { defaultParams as defaultApoParams } from '../common/absolute-price-oscillator/params.js';

// ── mark-jurik ──────────────────────────────────────────────────────────────
import { JurikMovingAverage } from '../mark-jurik/jurik-moving-average/jurik-moving-average.js';
import { defaultParams as defaultJmaParams } from '../mark-jurik/jurik-moving-average/params.js';
import { JurikCompositeFractalBehaviorIndex } from '../mark-jurik/jurik-composite-fractal-behavior-index/jurik-composite-fractal-behavior-index.js';
import { defaultParams as defaultCfbParams } from '../mark-jurik/jurik-composite-fractal-behavior-index/params.js';
import { JurikZeroLagVelocity } from '../mark-jurik/jurik-zero-lag-velocity/jurik-zero-lag-velocity.js';
import { defaultParams as defaultVelParams } from '../mark-jurik/jurik-zero-lag-velocity/params.js';
import { JurikRelativeTrendStrengthIndex } from '../mark-jurik/jurik-relative-trend-strength-index/jurik-relative-trend-strength-index.js';
import { defaultParams as defaultRsxParams } from '../mark-jurik/jurik-relative-trend-strength-index/params.js';
import { JurikDirectionalMovementIndex } from '../mark-jurik/jurik-directional-movement-index/jurik-directional-movement-index.js';
import { defaultParams as defaultDmxParams } from '../mark-jurik/jurik-directional-movement-index/params.js';
import { JurikTurningPointOscillator } from '../mark-jurik/jurik-turning-point-oscillator/jurik-turning-point-oscillator.js';
import { defaultParams as defaultJtpoParams } from '../mark-jurik/jurik-turning-point-oscillator/params.js';
import { JurikAdaptiveRelativeTrendStrengthIndex } from '../mark-jurik/jurik-adaptive-relative-trend-strength-index/jurik-adaptive-relative-trend-strength-index.js';
import { defaultParams as defaultJarsxParams } from '../mark-jurik/jurik-adaptive-relative-trend-strength-index/params.js';
import { JurikAdaptiveZeroLagVelocity } from '../mark-jurik/jurik-adaptive-zero-lag-velocity/jurik-adaptive-zero-lag-velocity.js';
import { defaultParams as defaultJavelParams } from '../mark-jurik/jurik-adaptive-zero-lag-velocity/params.js';
import { JurikCommodityChannelIndex } from '../mark-jurik/jurik-commodity-channel-index/jurik-commodity-channel-index.js';
import { defaultParams as defaultJccxParams } from '../mark-jurik/jurik-commodity-channel-index/params.js';
import { JurikFractalAdaptiveZeroLagVelocity } from '../mark-jurik/jurik-fractal-adaptive-zero-lag-velocity/jurik-fractal-adaptive-zero-lag-velocity.js';
import { defaultParams as defaultJvelcfbParams } from '../mark-jurik/jurik-fractal-adaptive-zero-lag-velocity/params.js';
import { JurikWaveletSampler } from '../mark-jurik/jurik-wavelet-sampler/jurik-wavelet-sampler.js';
import { defaultParams as defaultWavParams } from '../mark-jurik/jurik-wavelet-sampler/params.js';

// ── patrick-mulloy ──────────────────────────────────────────────────────────
import { DoubleExponentialMovingAverage } from '../patrick-mulloy/double-exponential-moving-average/double-exponential-moving-average.js';
import { defaultLengthParams as defaultDemaLengthParams } from '../patrick-mulloy/double-exponential-moving-average/params.js';
import { TripleExponentialMovingAverage } from '../patrick-mulloy/triple-exponential-moving-average/triple-exponential-moving-average.js';
import { defaultLengthParams as defaultTemaLengthParams } from '../patrick-mulloy/triple-exponential-moving-average/params.js';

// ── tim-tillson ─────────────────────────────────────────────────────────────
import { T2ExponentialMovingAverage } from '../tim-tillson/t2-exponential-moving-average/t2-exponential-moving-average.js';
import { defaultLengthParams as defaultT2LengthParams, defaultSmoothingFactorParams as defaultT2SfParams } from '../tim-tillson/t2-exponential-moving-average/params.js';
import { T3ExponentialMovingAverage } from '../tim-tillson/t3-exponential-moving-average/t3-exponential-moving-average.js';
import { defaultLengthParams as defaultT3LengthParams, defaultSmoothingFactorParams as defaultT3SfParams } from '../tim-tillson/t3-exponential-moving-average/params.js';

// ── perry-kaufman ───────────────────────────────────────────────────────────
import { KaufmanAdaptiveMovingAverage } from '../perry-kaufman/kaufman-adaptive-moving-average/kaufman-adaptive-moving-average.js';
import { defaultLengthParams as defaultKamaLengthParams, defaultSmoothingFactorParams as defaultKamaSfParams } from '../perry-kaufman/kaufman-adaptive-moving-average/params.js';

// ── john-ehlers ─────────────────────────────────────────────────────────────
import { MesaAdaptiveMovingAverage } from '../john-ehlers/mesa-adaptive-moving-average/mesa-adaptive-moving-average.js';
import { defaultLengthParams as defaultMamaLengthParams, defaultSmoothingFactorParams as defaultMamaSfParams } from '../john-ehlers/mesa-adaptive-moving-average/params.js';
import { FractalAdaptiveMovingAverage } from '../john-ehlers/fractal-adaptive-moving-average/fractal-adaptive-moving-average.js';
import { defaultParams as defaultFramaParams } from '../john-ehlers/fractal-adaptive-moving-average/params.js';
import { DominantCycle } from '../john-ehlers/dominant-cycle/dominant-cycle.js';
import { SuperSmoother } from '../john-ehlers/super-smoother/super-smoother.js';
import { defaultParams as defaultSsParams } from '../john-ehlers/super-smoother/params.js';
import { CenterOfGravityOscillator } from '../john-ehlers/center-of-gravity-oscillator/center-of-gravity-oscillator.js';
import { defaultParams as defaultCogParams } from '../john-ehlers/center-of-gravity-oscillator/params.js';
import { CyberCycle } from '../john-ehlers/cyber-cycle/cyber-cycle.js';
import { defaultLengthParams as defaultCcLengthParams } from '../john-ehlers/cyber-cycle/length-params.js';
import { defaultSmoothingFactorParams as defaultCcSfParams } from '../john-ehlers/cyber-cycle/smoothing-factor-params.js';
import { InstantaneousTrendLine } from '../john-ehlers/instantaneous-trend-line/instantaneous-trend-line.js';
import { defaultLengthParams as defaultItlLengthParams } from '../john-ehlers/instantaneous-trend-line/length-params.js';
import { defaultSmoothingFactorParams as defaultItlSfParams } from '../john-ehlers/instantaneous-trend-line/smoothing-factor-params.js';
import { ZeroLagExponentialMovingAverage } from '../john-ehlers/zero-lag-exponential-moving-average/zero-lag-exponential-moving-average.js';
import { defaultParams as defaultZemaParams } from '../john-ehlers/zero-lag-exponential-moving-average/params.js';
import { ZeroLagErrorCorrectingExponentialMovingAverage } from '../john-ehlers/zero-lag-error-correcting-exponential-moving-average/zero-lag-error-correcting-exponential-moving-average.js';
import { defaultParams as defaultZecemaParams } from '../john-ehlers/zero-lag-error-correcting-exponential-moving-average/params.js';
import { RoofingFilter } from '../john-ehlers/roofing-filter/roofing-filter.js';
import { defaultParams as defaultRoofParams } from '../john-ehlers/roofing-filter/params.js';
import { SineWave } from '../john-ehlers/sine-wave/sine-wave.js';
import { HilbertTransformerInstantaneousTrendLine } from '../john-ehlers/hilbert-transformer-instantaneous-trend-line/hilbert-transformer-instantaneous-trend-line.js';
import { TrendCycleMode } from '../john-ehlers/trend-cycle-mode/trend-cycle-mode.js';
import { CoronaSpectrum } from '../john-ehlers/corona-spectrum/corona-spectrum.js';
import { CoronaSignalToNoiseRatio } from '../john-ehlers/corona-signal-to-noise-ratio/corona-signal-to-noise-ratio.js';
import { CoronaSwingPosition } from '../john-ehlers/corona-swing-position/corona-swing-position.js';
import { CoronaTrendVigor } from '../john-ehlers/corona-trend-vigor/corona-trend-vigor.js';
import { AutoCorrelationIndicator } from '../john-ehlers/auto-correlation-indicator/auto-correlation-indicator.js';
import { AutoCorrelationPeriodogram } from '../john-ehlers/auto-correlation-periodogram/auto-correlation-periodogram.js';
import { CombBandPassSpectrum } from '../john-ehlers/comb-band-pass-spectrum/comb-band-pass-spectrum.js';
import { DiscreteFourierTransformSpectrum } from '../john-ehlers/discrete-fourier-transform-spectrum/discrete-fourier-transform-spectrum.js';

// ── welles-wilder ───────────────────────────────────────────────────────────
import { TrueRange } from '../welles-wilder/true-range/true-range.js';
import { AverageTrueRange } from '../welles-wilder/average-true-range/average-true-range.js';
import { NormalizedAverageTrueRange } from '../welles-wilder/normalized-average-true-range/normalized-average-true-range.js';
import { DirectionalMovementPlus } from '../welles-wilder/directional-movement-plus/directional-movement-plus.js';
import { DirectionalMovementMinus } from '../welles-wilder/directional-movement-minus/directional-movement-minus.js';
import { DirectionalIndicatorPlus } from '../welles-wilder/directional-indicator-plus/directional-indicator-plus.js';
import { DirectionalIndicatorMinus } from '../welles-wilder/directional-indicator-minus/directional-indicator-minus.js';
import { DirectionalMovementIndex } from '../welles-wilder/directional-movement-index/directional-movement-index.js';
import { AverageDirectionalMovementIndex } from '../welles-wilder/average-directional-movement-index/average-directional-movement-index.js';
import { AverageDirectionalMovementIndexRating } from '../welles-wilder/average-directional-movement-index-rating/average-directional-movement-index-rating.js';
import { RelativeStrengthIndex } from '../welles-wilder/relative-strength-index/relative-strength-index.js';
import { defaultParams as defaultRsiParams } from '../welles-wilder/relative-strength-index/params.js';
import { ParabolicStopAndReverse } from '../welles-wilder/parabolic-stop-and-reverse/parabolic-stop-and-reverse.js';

// ── john-bollinger ──────────────────────────────────────────────────────────
import { BollingerBands } from '../john-bollinger/bollinger-bands/bollinger-bands.js';
import { defaultParams as defaultBbParams } from '../john-bollinger/bollinger-bands/params.js';
import { BollingerBandsTrend } from '../john-bollinger/bollinger-bands-trend/bollinger-bands-trend.js';
import { defaultParams as defaultBbtrendParams } from '../john-bollinger/bollinger-bands-trend/params.js';

// ── gerald-appel ────────────────────────────────────────────────────────────
import { PercentagePriceOscillator } from '../gerald-appel/percentage-price-oscillator/percentage-price-oscillator.js';
import { defaultParams as defaultPpoParams } from '../gerald-appel/percentage-price-oscillator/params.js';
import { MovingAverageConvergenceDivergence } from '../gerald-appel/moving-average-convergence-divergence/moving-average-convergence-divergence.js';

// ── tushar-chande ───────────────────────────────────────────────────────────
import { ChandeMomentumOscillator } from '../tushar-chande/chande-momentum-oscillator/chande-momentum-oscillator.js';
import { defaultParams as defaultCmoParams } from '../tushar-chande/chande-momentum-oscillator/params.js';
import { StochasticRelativeStrengthIndex } from '../tushar-chande/stochastic-relative-strength-index/stochastic-relative-strength-index.js';
import { defaultParams as defaultStochRsiParams } from '../tushar-chande/stochastic-relative-strength-index/params.js';
import { Aroon } from '../tushar-chande/aroon/aroon.js';
import { defaultParams as defaultAroonParams } from '../tushar-chande/aroon/params.js';

// ── donald-lambert ──────────────────────────────────────────────────────────
import { CommodityChannelIndex } from '../donald-lambert/commodity-channel-index/commodity-channel-index.js';
import { defaultParams as defaultCciParams } from '../donald-lambert/commodity-channel-index/params.js';

// ── gene-quong ──────────────────────────────────────────────────────────────
import { MoneyFlowIndex } from '../gene-quong/money-flow-index/money-flow-index.js';
import { defaultParams as defaultMfiParams } from '../gene-quong/money-flow-index/params.js';

// ── george-lane ─────────────────────────────────────────────────────────────
import { Stochastic } from '../george-lane/stochastic/stochastic.js';
import { defaultParams as defaultStochParams } from '../george-lane/stochastic/params.js';

// ── joseph-granville ────────────────────────────────────────────────────────
import { OnBalanceVolume } from '../joseph-granville/on-balance-volume/on-balance-volume.js';

// ── igor-livshin ────────────────────────────────────────────────────────────
import { BalanceOfPower } from '../igor-livshin/balance-of-power/balance-of-power.js';

// ── marc-chaikin ────────────────────────────────────────────────────────────
import { AdvanceDecline } from '../marc-chaikin/advance-decline/advance-decline.js';
import { AdvanceDeclineOscillator } from '../marc-chaikin/advance-decline-oscillator/advance-decline-oscillator.js';
import { defaultParams as defaultAdoscParams } from '../marc-chaikin/advance-decline-oscillator/params.js';

// ── larry-williams ──────────────────────────────────────────────────────────
import { WilliamsPercentR } from '../larry-williams/williams-percent-r/williams-percent-r.js';
import { UltimateOscillator } from '../larry-williams/ultimate-oscillator/ultimate-oscillator.js';

// ── jack-hutson ─────────────────────────────────────────────────────────────
import { TripleExponentialMovingAverageOscillator } from '../jack-hutson/triple-exponential-moving-average-oscillator/triple-exponential-moving-average-oscillator.js';
import { defaultParams as defaultTrixParams } from '../jack-hutson/triple-exponential-moving-average-oscillator/params.js';

// ── vladimir-kravchuk ───────────────────────────────────────────────────────
import { AdaptiveTrendAndCycleFilter } from '../vladimir-kravchuk/adaptive-trend-and-cycle-filter/adaptive-trend-and-cycle-filter.js';

// ── custom ──────────────────────────────────────────────────────────────────
import { GoertzelSpectrum } from '../custom/goertzel-spectrum/goertzel-spectrum.js';
import { MaximumEntropySpectrum } from '../custom/maximum-entropy-spectrum/maximum-entropy-spectrum.js';

// ── arnaud-legoux ───────────────────────────────────────────────────────────
import { ArnaudLegouxMovingAverage } from '../arnaud-legoux/arnaud-legoux-moving-average/arnaud-legoux-moving-average.js';
import { defaultParams as defaultAlmaParams } from '../arnaud-legoux/arnaud-legoux-moving-average/params.js';


// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Returns `true` when the param key is present in the object. */
function hasKey(params: Record<string, unknown>, key: string): boolean {
    return key in params;
}

/** Returns `true` when `params` is `undefined`, `null`, or `{}`. */
function isEmpty(params: Record<string, unknown> | undefined | null): boolean {
    if (params == null) return true;
    return Object.keys(params).length === 0;
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

/**
 * Creates an indicator from its identifier and a plain params object.
 *
 * @param identifier - The indicator to create.
 * @param params     - Construction parameters (shape depends on the indicator).
 *                     Pass `undefined` or `{}` for default parameters.
 * @returns A fully constructed {@link Indicator}.
 * @throws  If the identifier is not supported or params are invalid.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function createIndicator(identifier: IndicatorIdentifier, params?: Record<string, any>): Indicator {
    const p = params ?? {};

    switch (identifier) {

        // ── common ──────────────────────────────────────────────────────

        case IndicatorIdentifier.SimpleMovingAverage:
            return new SimpleMovingAverage({ ...defaultSmaParams(), ...p });

        case IndicatorIdentifier.WeightedMovingAverage:
            return new WeightedMovingAverage({ ...defaultWmaParams(), ...p });

        case IndicatorIdentifier.TriangularMovingAverage:
            return new TriangularMovingAverage({ ...defaultTrimaParams(), ...p });

        case IndicatorIdentifier.ExponentialMovingAverage:
            if (hasKey(p, 'smoothingFactor')) {
                return new ExponentialMovingAverage(p as any);
            }
            return new ExponentialMovingAverage({ ...defaultEmaLengthParams(), ...p });

        case IndicatorIdentifier.Variance:
            return new Variance({ ...defaultVarianceParams(), ...p });

        case IndicatorIdentifier.StandardDeviation:
            return new StandardDeviation({ ...defaultStdDevParams(), ...p });

        case IndicatorIdentifier.Momentum:
            return new Momentum({ ...defaultMomentumParams(), ...p });

        case IndicatorIdentifier.RateOfChange:
            return new RateOfChange({ ...defaultRocParams(), ...p });

        case IndicatorIdentifier.RateOfChangePercent:
            return new RateOfChangePercent({ ...defaultRocpParams(), ...p });

        case IndicatorIdentifier.RateOfChangeRatio:
            return new RateOfChangeRatio({ ...defaultRocrParams(), ...p });

        case IndicatorIdentifier.AbsolutePriceOscillator:
            return new AbsolutePriceOscillator({ ...defaultApoParams(), ...p });

        case IndicatorIdentifier.PearsonsCorrelationCoefficient:
            return new PearsonsCorrelationCoefficient({ ...defaultCorrelParams(), ...p });

        case IndicatorIdentifier.LinearRegression:
            return new LinearRegression({ ...defaultLinregParams(), ...p });

        // ── mark-jurik ──────────────────────────────────────────────────

        case IndicatorIdentifier.JurikMovingAverage:
            return new JurikMovingAverage({ ...defaultJmaParams(), ...p });

        case IndicatorIdentifier.JurikCompositeFractalBehaviorIndex:
            return new JurikCompositeFractalBehaviorIndex({ ...defaultCfbParams(), ...p });

        case IndicatorIdentifier.JurikZeroLagVelocity:
            return new JurikZeroLagVelocity({ ...defaultVelParams(), ...p });

        case IndicatorIdentifier.JurikRelativeTrendStrengthIndex:
            return new JurikRelativeTrendStrengthIndex({ ...defaultRsxParams(), ...p });

        case IndicatorIdentifier.JurikDirectionalMovementIndex:
            return new JurikDirectionalMovementIndex({ ...defaultDmxParams(), ...p });

        case IndicatorIdentifier.JurikTurningPointOscillator:
            return new JurikTurningPointOscillator({ ...defaultJtpoParams(), ...p });

        case IndicatorIdentifier.JurikAdaptiveRelativeTrendStrengthIndex:
            return new JurikAdaptiveRelativeTrendStrengthIndex({ ...defaultJarsxParams(), ...p });

        case IndicatorIdentifier.JurikAdaptiveZeroLagVelocity:
            return new JurikAdaptiveZeroLagVelocity({ ...defaultJavelParams(), ...p });

        case IndicatorIdentifier.JurikCommodityChannelIndex:
            return new JurikCommodityChannelIndex({ ...defaultJccxParams(), ...p });

        case IndicatorIdentifier.JurikFractalAdaptiveZeroLagVelocity:
            return new JurikFractalAdaptiveZeroLagVelocity({ ...defaultJvelcfbParams(), ...p });

        case IndicatorIdentifier.JurikWaveletSampler:
            return new JurikWaveletSampler({ ...defaultWavParams(), ...p });

        // ── patrick-mulloy ──────────────────────────────────────────────

        case IndicatorIdentifier.DoubleExponentialMovingAverage:
            if (hasKey(p, 'smoothingFactor')) {
                return new DoubleExponentialMovingAverage(p as any);
            }
            return new DoubleExponentialMovingAverage({ ...defaultDemaLengthParams(), ...p });

        case IndicatorIdentifier.TripleExponentialMovingAverage:
            if (hasKey(p, 'smoothingFactor')) {
                return new TripleExponentialMovingAverage(p as any);
            }
            return new TripleExponentialMovingAverage({ ...defaultTemaLengthParams(), ...p });

        // ── tim-tillson ─────────────────────────────────────────────────

        case IndicatorIdentifier.T2ExponentialMovingAverage:
            if (hasKey(p, 'smoothingFactor')) {
                return new T2ExponentialMovingAverage({ ...defaultT2SfParams(), ...p } as any);
            }
            return new T2ExponentialMovingAverage({ ...defaultT2LengthParams(), ...p });

        case IndicatorIdentifier.T3ExponentialMovingAverage:
            if (hasKey(p, 'smoothingFactor')) {
                return new T3ExponentialMovingAverage({ ...defaultT3SfParams(), ...p } as any);
            }
            return new T3ExponentialMovingAverage({ ...defaultT3LengthParams(), ...p });

        // ── perry-kaufman ───────────────────────────────────────────────

        case IndicatorIdentifier.KaufmanAdaptiveMovingAverage:
            if (hasKey(p, 'fastestSmoothingFactor') || hasKey(p, 'slowestSmoothingFactor')) {
                return new KaufmanAdaptiveMovingAverage({ ...defaultKamaSfParams(), ...p } as any);
            }
            return new KaufmanAdaptiveMovingAverage({ ...defaultKamaLengthParams(), ...p });

        // ── john-ehlers ─────────────────────────────────────────────────

        case IndicatorIdentifier.MesaAdaptiveMovingAverage:
            if (hasKey(p, 'fastLimitSmoothingFactor') || hasKey(p, 'slowLimitSmoothingFactor')) {
                return MesaAdaptiveMovingAverage.fromSmoothingFactor({
                    ...defaultMamaSfParams(),
                    ...p,
                } as any);
            }
            if (isEmpty(p)) {
                return MesaAdaptiveMovingAverage.default();
            }
            return MesaAdaptiveMovingAverage.fromLength({ ...defaultMamaLengthParams(), ...p } as any);

        case IndicatorIdentifier.FractalAdaptiveMovingAverage:
            return new FractalAdaptiveMovingAverage({ ...defaultFramaParams(), ...p });

        case IndicatorIdentifier.DominantCycle:
            if (isEmpty(p)) return DominantCycle.default();
            return DominantCycle.fromParams(p as any);

        case IndicatorIdentifier.SuperSmoother:
            return new SuperSmoother({ ...defaultSsParams(), ...p });

        case IndicatorIdentifier.CenterOfGravityOscillator:
            return new CenterOfGravityOscillator({ ...defaultCogParams(), ...p });

        case IndicatorIdentifier.CyberCycle:
            if (hasKey(p, 'smoothingFactor')) {
                return new CyberCycle({ ...defaultCcSfParams(), ...p } as any);
            }
            return new CyberCycle({ ...defaultCcLengthParams(), ...p });

        case IndicatorIdentifier.InstantaneousTrendLine:
            if (hasKey(p, 'smoothingFactor')) {
                return new InstantaneousTrendLine({ ...defaultItlSfParams(), ...p } as any);
            }
            return new InstantaneousTrendLine({ ...defaultItlLengthParams(), ...p });

        case IndicatorIdentifier.ZeroLagExponentialMovingAverage:
            return new ZeroLagExponentialMovingAverage({
                ...defaultZemaParams(),
                ...p,
            });

        case IndicatorIdentifier.ZeroLagErrorCorrectingExponentialMovingAverage:
            return new ZeroLagErrorCorrectingExponentialMovingAverage({
                ...defaultZecemaParams(),
                ...p,
            });

        case IndicatorIdentifier.RoofingFilter:
            return new RoofingFilter({ ...defaultRoofParams(), ...p });

        case IndicatorIdentifier.SineWave:
            if (isEmpty(p)) return SineWave.default();
            return SineWave.fromParams(p as any);

        case IndicatorIdentifier.HilbertTransformerInstantaneousTrendLine:
            if (isEmpty(p)) return HilbertTransformerInstantaneousTrendLine.default();
            return HilbertTransformerInstantaneousTrendLine.fromParams(p as any);

        case IndicatorIdentifier.TrendCycleMode:
            if (isEmpty(p)) return TrendCycleMode.default();
            return TrendCycleMode.fromParams(p as any);

        case IndicatorIdentifier.CoronaSpectrum:
            if (isEmpty(p)) return CoronaSpectrum.default();
            return CoronaSpectrum.fromParams(p as any);

        case IndicatorIdentifier.CoronaSignalToNoiseRatio:
            if (isEmpty(p)) return CoronaSignalToNoiseRatio.default();
            return CoronaSignalToNoiseRatio.fromParams(p as any);

        case IndicatorIdentifier.CoronaSwingPosition:
            if (isEmpty(p)) return CoronaSwingPosition.default();
            return CoronaSwingPosition.fromParams(p as any);

        case IndicatorIdentifier.CoronaTrendVigor:
            if (isEmpty(p)) return CoronaTrendVigor.default();
            return CoronaTrendVigor.fromParams(p as any);

        case IndicatorIdentifier.AutoCorrelationIndicator:
            if (isEmpty(p)) return AutoCorrelationIndicator.default();
            return AutoCorrelationIndicator.fromParams(p as any);

        case IndicatorIdentifier.AutoCorrelationPeriodogram:
            if (isEmpty(p)) return AutoCorrelationPeriodogram.default();
            return AutoCorrelationPeriodogram.fromParams(p as any);

        case IndicatorIdentifier.CombBandPassSpectrum:
            if (isEmpty(p)) return CombBandPassSpectrum.default();
            return CombBandPassSpectrum.fromParams(p as any);

        case IndicatorIdentifier.DiscreteFourierTransformSpectrum:
            if (isEmpty(p)) return DiscreteFourierTransformSpectrum.default();
            return DiscreteFourierTransformSpectrum.fromParams(p as any);

        // ── welles-wilder ───────────────────────────────────────────────

        case IndicatorIdentifier.TrueRange:
            return new TrueRange();

        case IndicatorIdentifier.AverageTrueRange:
            return new AverageTrueRange((p as any).length ?? 14);

        case IndicatorIdentifier.NormalizedAverageTrueRange:
            return new NormalizedAverageTrueRange((p as any).length ?? 14);

        case IndicatorIdentifier.DirectionalMovementPlus:
            return new DirectionalMovementPlus((p as any).length ?? 14);

        case IndicatorIdentifier.DirectionalMovementMinus:
            return new DirectionalMovementMinus((p as any).length ?? 14);

        case IndicatorIdentifier.DirectionalIndicatorPlus:
            return new DirectionalIndicatorPlus((p as any).length ?? 14);

        case IndicatorIdentifier.DirectionalIndicatorMinus:
            return new DirectionalIndicatorMinus((p as any).length ?? 14);

        case IndicatorIdentifier.DirectionalMovementIndex:
            return new DirectionalMovementIndex((p as any).length ?? 14);

        case IndicatorIdentifier.AverageDirectionalMovementIndex:
            return new AverageDirectionalMovementIndex((p as any).length ?? 14);

        case IndicatorIdentifier.AverageDirectionalMovementIndexRating:
            return new AverageDirectionalMovementIndexRating((p as any).length ?? 14);

        case IndicatorIdentifier.RelativeStrengthIndex:
            return new RelativeStrengthIndex({ ...defaultRsiParams(), ...p });

        case IndicatorIdentifier.ParabolicStopAndReverse:
            return new ParabolicStopAndReverse(isEmpty(p) ? undefined : p as any);

        // ── john-bollinger ──────────────────────────────────────────────

        case IndicatorIdentifier.BollingerBands:
            return new BollingerBands({ ...defaultBbParams(), ...p });

        case IndicatorIdentifier.BollingerBandsTrend:
            return new BollingerBandsTrend({
                ...defaultBbtrendParams(),
                ...p,
            });

        // ── gerald-appel ────────────────────────────────────────────────

        case IndicatorIdentifier.PercentagePriceOscillator:
            return new PercentagePriceOscillator({ ...defaultPpoParams(), ...p });

        case IndicatorIdentifier.MovingAverageConvergenceDivergence:
            return new MovingAverageConvergenceDivergence(isEmpty(p) ? undefined : p as any);

        // ── tushar-chande ───────────────────────────────────────────────

        case IndicatorIdentifier.ChandeMomentumOscillator:
            return new ChandeMomentumOscillator({ ...defaultCmoParams(), ...p });

        case IndicatorIdentifier.StochasticRelativeStrengthIndex:
            return new StochasticRelativeStrengthIndex({ ...defaultStochRsiParams(), ...p });

        case IndicatorIdentifier.Aroon:
            return new Aroon({ ...defaultAroonParams(), ...p });

        // ── donald-lambert ──────────────────────────────────────────────

        case IndicatorIdentifier.CommodityChannelIndex:
            return new CommodityChannelIndex({ ...defaultCciParams(), ...p });

        // ── gene-quong ──────────────────────────────────────────────────

        case IndicatorIdentifier.MoneyFlowIndex:
            return new MoneyFlowIndex({ ...defaultMfiParams(), ...p });

        // ── george-lane ─────────────────────────────────────────────────

        case IndicatorIdentifier.Stochastic:
            return new Stochastic({ ...defaultStochParams(), ...p });

        // ── joseph-granville ────────────────────────────────────────────

        case IndicatorIdentifier.OnBalanceVolume:
            return new OnBalanceVolume(isEmpty(p) ? undefined : p as any);

        // ── igor-livshin ────────────────────────────────────────────────

        case IndicatorIdentifier.BalanceOfPower:
            return new BalanceOfPower();

        // ── marc-chaikin ────────────────────────────────────────────────

        case IndicatorIdentifier.AdvanceDecline:
            return new AdvanceDecline();

        case IndicatorIdentifier.AdvanceDeclineOscillator:
            return new AdvanceDeclineOscillator({ ...defaultAdoscParams(), ...p });

        // ── larry-williams ──────────────────────────────────────────────

        case IndicatorIdentifier.WilliamsPercentR:
            return new WilliamsPercentR((p as any).length ?? 14);

        case IndicatorIdentifier.UltimateOscillator:
            return new UltimateOscillator(isEmpty(p) ? undefined : p as any);

        // ── jack-hutson ─────────────────────────────────────────────────

        case IndicatorIdentifier.TripleExponentialMovingAverageOscillator:
            return new TripleExponentialMovingAverageOscillator({ ...defaultTrixParams(), ...p });

        // ── vladimir-kravchuk ───────────────────────────────────────────

        case IndicatorIdentifier.AdaptiveTrendAndCycleFilter:
            if (isEmpty(p)) return AdaptiveTrendAndCycleFilter.default();
            return AdaptiveTrendAndCycleFilter.fromParams(p as any);

        // ── custom ──────────────────────────────────────────────────────

        case IndicatorIdentifier.GoertzelSpectrum:
            if (isEmpty(p)) return GoertzelSpectrum.default();
            return GoertzelSpectrum.fromParams(p as any);

        case IndicatorIdentifier.MaximumEntropySpectrum:
            if (isEmpty(p)) return MaximumEntropySpectrum.default();
            return MaximumEntropySpectrum.fromParams(p as any);

        // ── arnaud-legoux ───────────────────────────────────────────

        case IndicatorIdentifier.ArnaudLegouxMovingAverage:
            return new ArnaudLegouxMovingAverage({ ...defaultAlmaParams(), ...p });

        default:
            throw new Error(`Unsupported indicator: ${IndicatorIdentifier[identifier] ?? identifier}`);
    }
}
