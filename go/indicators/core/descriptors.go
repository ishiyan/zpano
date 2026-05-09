package core

import (
	"zpano/indicators/core/outputs/shape"
)

// descriptors is the static registry of taxonomic descriptors for all implemented indicators.
//
// Output Kind values mirror each indicator package's Output enumeration, which starts
// at iota+1; they are written here as integer literals with a comment naming the constant.
//
//nolint:funlen,maintidx
var descriptors = map[Identifier]Descriptor{

	// ── common ────────────────────────────────────────────────────────────

	AbsolutePriceOscillator: {
		Identifier: AbsolutePriceOscillator, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	ExponentialMovingAverage: {
		Identifier: ExponentialMovingAverage, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},
	LinearRegression: {
		Identifier: LinearRegression, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 2 /* Forecast */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 3 /* Intercept */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 4 /* SlopeRad */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
			{Kind: 5 /* SlopeDeg */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
		},
	},
	Momentum: {
		Identifier: Momentum, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	PearsonsCorrelationCoefficient: {
		Identifier: PearsonsCorrelationCoefficient, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Correlation, Pane: Own}},
	},
	RateOfChange: {
		Identifier: RateOfChange, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	RateOfChangePercent: {
		Identifier: RateOfChangePercent, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	RateOfChangeRatio: {
		Identifier: RateOfChangeRatio, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	SimpleMovingAverage: {
		Identifier: SimpleMovingAverage, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},
	StandardDeviation: {
		Identifier: StandardDeviation, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Volatility, Pane: Own}},
	},
	TriangularMovingAverage: {
		Identifier: TriangularMovingAverage, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},
	Variance: {
		Identifier: Variance, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Volatility, Pane: Own}},
	},
	WeightedMovingAverage: {
		Identifier: WeightedMovingAverage, Family: "Common",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},

	// ── arnaud legoux ──────────────────────────────────────────────────────

	ArnaudLegouxMovingAverage: {
		Identifier: ArnaudLegouxMovingAverage, Family: "Arnaud Legoux",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},

	// ── donald lambert ─────────────────────────────────────────────────────

	CommodityChannelIndex: {
		Identifier: CommodityChannelIndex, Family: "Donald Lambert",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own}},
	},

	// ── gene quong ─────────────────────────────────────────────────────────

	MoneyFlowIndex: {
		Identifier: MoneyFlowIndex, Family: "Gene Quong",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: AggregateBarVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own}},
	},

	// ── george lane ────────────────────────────────────────────────────────

	Stochastic: {
		Identifier: Stochastic, Family: "George Lane",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* FastK */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 2 /* SlowK */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 3 /* SlowD */, Shape: shape.Scalar, Role: Signal, Pane: Own},
		},
	},

	// ── gerald appel ───────────────────────────────────────────────────────

	MovingAverageConvergenceDivergence: {
		Identifier: MovingAverageConvergenceDivergence, Family: "Gerald Appel",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* MACD */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
			{Kind: 2 /* Signal */, Shape: shape.Scalar, Role: Signal, Pane: Own},
			{Kind: 3 /* Histogram */, Shape: shape.Scalar, Role: Histogram, Pane: Own},
		},
	},
	PercentagePriceOscillator: {
		Identifier: PercentagePriceOscillator, Family: "Gerald Appel",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},

	// ── igor livshin ───────────────────────────────────────────────────────

	BalanceOfPower: {
		Identifier: BalanceOfPower, Family: "Igor Livshin",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own}},
	},

	// ── jack hutson ────────────────────────────────────────────────────────

	TripleExponentialMovingAverageOscillator: {
		Identifier: TripleExponentialMovingAverageOscillator, Family: "Jack Hutson",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},

	// ── john bollinger ─────────────────────────────────────────────────────

	BollingerBands: {
		Identifier: BollingerBands, Family: "John Bollinger",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Lower */, Shape: shape.Scalar, Role: Envelope, Pane: Price},
			{Kind: 2 /* Middle */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 3 /* Upper */, Shape: shape.Scalar, Role: Envelope, Pane: Price},
			{Kind: 4 /* BandWidth */, Shape: shape.Scalar, Role: Volatility, Pane: Own},
			{Kind: 5 /* PercentBand */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 6 /* Band */, Shape: shape.Band, Role: Envelope, Pane: Price},
		},
	},
	BollingerBandsTrend: {
		Identifier: BollingerBandsTrend, Family: "John Bollinger",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},

	// ── john ehlers ────────────────────────────────────────────────────────

	AutoCorrelationIndicator: {
		Identifier: AutoCorrelationIndicator, Family: "John Ehlers",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Heatmap, Role: Correlation, Pane: Own}},
	},
	AutoCorrelationPeriodogram: {
		Identifier: AutoCorrelationPeriodogram, Family: "John Ehlers",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Heatmap, Role: Spectrum, Pane: Own}},
	},
	CenterOfGravityOscillator: {
		Identifier: CenterOfGravityOscillator, Family: "John Ehlers",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
			{Kind: 2 /* Trigger */, Shape: shape.Scalar, Role: Signal, Pane: Own},
		},
	},
	CombBandPassSpectrum: {
		Identifier: CombBandPassSpectrum, Family: "John Ehlers",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Heatmap, Role: Spectrum, Pane: Own}},
	},
	CoronaSignalToNoiseRatio: {
		Identifier: CoronaSignalToNoiseRatio, Family: "John Ehlers",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Heatmap, Role: Spectrum, Pane: Own},
			{Kind: 2 /* SignalToNoiseRatio */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
		},
	},
	CoronaSpectrum: {
		Identifier: CoronaSpectrum, Family: "John Ehlers",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Heatmap, Role: Spectrum, Pane: Own},
			{Kind: 2 /* DominantCycle */, Shape: shape.Scalar, Role: CyclePeriod, Pane: Own},
			{Kind: 3 /* DominantCycleMedian */, Shape: shape.Scalar, Role: CyclePeriod, Pane: Own},
		},
	},
	CoronaSwingPosition: {
		Identifier: CoronaSwingPosition, Family: "John Ehlers",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Heatmap, Role: Spectrum, Pane: Own},
			{Kind: 2 /* SwingPosition */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
		},
	},
	CoronaTrendVigor: {
		Identifier: CoronaTrendVigor, Family: "John Ehlers",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Heatmap, Role: Spectrum, Pane: Own},
			{Kind: 2 /* TrendVigor */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
		},
	},
	CyberCycle: {
		Identifier: CyberCycle, Family: "John Ehlers",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
			{Kind: 2 /* Signal */, Shape: shape.Scalar, Role: Signal, Pane: Own},
		},
	},
	DiscreteFourierTransformSpectrum: {
		Identifier: DiscreteFourierTransformSpectrum, Family: "John Ehlers",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Heatmap, Role: Spectrum, Pane: Own}},
	},
	DominantCycle: {
		Identifier: DominantCycle, Family: "John Ehlers",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* RawPeriod */, Shape: shape.Scalar, Role: CyclePeriod, Pane: Own},
			{Kind: 2 /* Period */, Shape: shape.Scalar, Role: CyclePeriod, Pane: Own},
			{Kind: 3 /* Phase */, Shape: shape.Scalar, Role: CyclePhase, Pane: Own},
		},
	},
	FractalAdaptiveMovingAverage: {
		Identifier: FractalAdaptiveMovingAverage, Family: "John Ehlers",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 2 /* Fdim */, Shape: shape.Scalar, Role: FractalDimension, Pane: Own},
		},
	},
	HilbertTransformerInstantaneousTrendLine: {
		Identifier: HilbertTransformerInstantaneousTrendLine, Family: "John Ehlers",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 2 /* DominantCyclePeriod */, Shape: shape.Scalar, Role: CyclePeriod, Pane: Own},
		},
	},
	InstantaneousTrendLine: {
		Identifier: InstantaneousTrendLine, Family: "John Ehlers",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 2 /* Trigger */, Shape: shape.Scalar, Role: Signal, Pane: Price},
		},
	},
	MesaAdaptiveMovingAverage: {
		Identifier: MesaAdaptiveMovingAverage, Family: "John Ehlers",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value (MAMA) */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 2 /* Fama */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 3 /* Band */, Shape: shape.Band, Role: Envelope, Pane: Price},
		},
	},
	RoofingFilter: {
		Identifier: RoofingFilter, Family: "John Ehlers",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	SineWave: {
		Identifier: SineWave, Family: "John Ehlers",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
			{Kind: 2 /* Lead */, Shape: shape.Scalar, Role: Signal, Pane: Own},
			{Kind: 3 /* Band */, Shape: shape.Band, Role: Envelope, Pane: Own},
			{Kind: 4 /* DominantCyclePeriod */, Shape: shape.Scalar, Role: CyclePeriod, Pane: Own},
			{Kind: 5 /* DominantCyclePhase */, Shape: shape.Scalar, Role: CyclePhase, Pane: Own},
		},
	},
	SuperSmoother: {
		Identifier: SuperSmoother, Family: "John Ehlers",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},
	TrendCycleMode: {
		Identifier: TrendCycleMode, Family: "John Ehlers",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: RegimeFlag, Pane: Own},
			{Kind: 2 /* IsTrendMode */, Shape: shape.Scalar, Role: RegimeFlag, Pane: Own},
			{Kind: 3 /* IsCycleMode */, Shape: shape.Scalar, Role: RegimeFlag, Pane: Own},
			{Kind: 4 /* InstantaneousTrendLine */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 5 /* SineWave */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
			{Kind: 6 /* SineWaveLead */, Shape: shape.Scalar, Role: Signal, Pane: Own},
			{Kind: 7 /* DominantCyclePeriod */, Shape: shape.Scalar, Role: CyclePeriod, Pane: Own},
			{Kind: 8 /* DominantCyclePhase */, Shape: shape.Scalar, Role: CyclePhase, Pane: Own},
		},
	},
	ZeroLagErrorCorrectingExponentialMovingAverage: {
		Identifier: ZeroLagErrorCorrectingExponentialMovingAverage, Family: "John Ehlers",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},
	ZeroLagExponentialMovingAverage: {
		Identifier: ZeroLagExponentialMovingAverage, Family: "John Ehlers",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},

	// ── joseph granville ───────────────────────────────────────────────────

	OnBalanceVolume: {
		Identifier: OnBalanceVolume, Family: "Joseph Granville",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: AggregateBarVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: VolumeFlow, Pane: Own}},
	},

	// ── larry williams ─────────────────────────────────────────────────────

	UltimateOscillator: {
		Identifier: UltimateOscillator, Family: "Larry Williams",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own}},
	},
	WilliamsPercentR: {
		Identifier: WilliamsPercentR, Family: "Larry Williams",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own}},
	},

	// ── manfred durschner ──────────────────────────────────────────────────

	NewMovingAverage: {
		Identifier: NewMovingAverage, Family: "Manfred Dürschner",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},

	// ── marc chaikin ───────────────────────────────────────────────────────

	AdvanceDecline: {
		Identifier: AdvanceDecline, Family: "Marc Chaikin",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: AggregateBarVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: VolumeFlow, Pane: Own}},
	},
	AdvanceDeclineOscillator: {
		Identifier: AdvanceDeclineOscillator, Family: "Marc Chaikin",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: AggregateBarVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: VolumeFlow, Pane: Own}},
	},

	// ── mark jurik ─────────────────────────────────────────────────────────

	JurikAdaptiveRelativeTrendStrengthIndex: {
		Identifier: JurikAdaptiveRelativeTrendStrengthIndex, Family: "Mark Jurik",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	JurikAdaptiveZeroLagVelocity: {
		Identifier: JurikAdaptiveZeroLagVelocity, Family: "Mark Jurik",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	JurikCommodityChannelIndex: {
		Identifier: JurikCommodityChannelIndex, Family: "Mark Jurik",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	JurikCompositeFractalBehaviorIndex: {
		Identifier: JurikCompositeFractalBehaviorIndex, Family: "Mark Jurik",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	JurikDirectionalMovementIndex: {
		Identifier: JurikDirectionalMovementIndex, Family: "Mark Jurik",
		Adaptivity: Adaptive, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Bipolar */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
			{Kind: 2 /* Plus */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
			{Kind: 3 /* Minus */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
		},
	},
	JurikFractalAdaptiveZeroLagVelocity: {
		Identifier: JurikFractalAdaptiveZeroLagVelocity, Family: "Mark Jurik",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	JurikMovingAverage: {
		Identifier: JurikMovingAverage, Family: "Mark Jurik",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* MovingAverage */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},
	JurikRelativeTrendStrengthIndex: {
		Identifier: JurikRelativeTrendStrengthIndex, Family: "Mark Jurik",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	JurikTurningPointOscillator: {
		Identifier: JurikTurningPointOscillator, Family: "Mark Jurik",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},
	JurikWaveletSampler: {
		Identifier: JurikWaveletSampler, Family: "Mark Jurik",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},
	JurikZeroLagVelocity: {
		Identifier: JurikZeroLagVelocity, Family: "Mark Jurik",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Oscillator, Pane: Own}},
	},

	// ── patrick mulloy ─────────────────────────────────────────────────────

	DoubleExponentialMovingAverage: {
		Identifier: DoubleExponentialMovingAverage, Family: "Patrick Mulloy",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},
	TripleExponentialMovingAverage: {
		Identifier: TripleExponentialMovingAverage, Family: "Patrick Mulloy",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},

	// ── perry kaufman ──────────────────────────────────────────────────────

	KaufmanAdaptiveMovingAverage: {
		Identifier: KaufmanAdaptiveMovingAverage, Family: "Perry Kaufman",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},

	// ── tim tillson ────────────────────────────────────────────────────────

	T2ExponentialMovingAverage: {
		Identifier: T2ExponentialMovingAverage, Family: "Tim Tillson",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},
	T3ExponentialMovingAverage: {
		Identifier: T3ExponentialMovingAverage, Family: "Tim Tillson",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Smoother, Pane: Price}},
	},

	// ── tushar chande ──────────────────────────────────────────────────────

	Aroon: {
		Identifier: Aroon, Family: "Tushar Chande",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Up */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 2 /* Down */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 3 /* Osc */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
		},
	},
	ChandeMomentumOscillator: {
		Identifier: ChandeMomentumOscillator, Family: "Tushar Chande",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own}},
	},
	StochasticRelativeStrengthIndex: {
		Identifier: StochasticRelativeStrengthIndex, Family: "Tushar Chande",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* FastK */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 2 /* FastD */, Shape: shape.Scalar, Role: Signal, Pane: Own},
		},
	},

	// ── vladimir kravchuk ──────────────────────────────────────────────────

	AdaptiveTrendAndCycleFilter: {
		Identifier: AdaptiveTrendAndCycleFilter, Family: "Vladimir Kravchuk",
		Adaptivity: Adaptive, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Fatl */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 2 /* Satl */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 3 /* Rftl */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 4 /* Rstl */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 5 /* Rbci */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
			{Kind: 6 /* Ftlm */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
			{Kind: 7 /* Stlm */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
			{Kind: 8 /* Pcci */, Shape: shape.Scalar, Role: Oscillator, Pane: Own},
		},
	},

	// ── welles wilder ──────────────────────────────────────────────────────

	AverageDirectionalMovementIndex: {
		Identifier: AverageDirectionalMovementIndex, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 2 /* DirectionalMovementIndex */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 3 /* DirectionalIndicatorPlus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 4 /* DirectionalIndicatorMinus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 5 /* DirectionalMovementPlus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 6 /* DirectionalMovementMinus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 7 /* AverageTrueRange */, Shape: shape.Scalar, Role: Volatility, Pane: Own},
			{Kind: 8 /* TrueRange */, Shape: shape.Scalar, Role: Volatility, Pane: Own},
		},
	},
	AverageDirectionalMovementIndexRating: {
		Identifier: AverageDirectionalMovementIndexRating, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 2 /* AverageDirectionalMovementIndex */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 3 /* DirectionalMovementIndex */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 4 /* DirectionalIndicatorPlus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 5 /* DirectionalIndicatorMinus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 6 /* DirectionalMovementPlus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 7 /* DirectionalMovementMinus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 8 /* AverageTrueRange */, Shape: shape.Scalar, Role: Volatility, Pane: Own},
			{Kind: 9 /* TrueRange */, Shape: shape.Scalar, Role: Volatility, Pane: Own},
		},
	},
	AverageTrueRange: {
		Identifier: AverageTrueRange, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Volatility, Pane: Own}},
	},
	DirectionalIndicatorMinus: {
		Identifier: DirectionalIndicatorMinus, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 2 /* DirectionalMovementMinus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 3 /* AverageTrueRange */, Shape: shape.Scalar, Role: Volatility, Pane: Own},
			{Kind: 4 /* TrueRange */, Shape: shape.Scalar, Role: Volatility, Pane: Own},
		},
	},
	DirectionalIndicatorPlus: {
		Identifier: DirectionalIndicatorPlus, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 2 /* DirectionalMovementPlus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 3 /* AverageTrueRange */, Shape: shape.Scalar, Role: Volatility, Pane: Own},
			{Kind: 4 /* TrueRange */, Shape: shape.Scalar, Role: Volatility, Pane: Own},
		},
	},
	DirectionalMovementIndex: {
		Identifier: DirectionalMovementIndex, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{
			{Kind: 1 /* Value */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own},
			{Kind: 2 /* DirectionalIndicatorPlus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 3 /* DirectionalIndicatorMinus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 4 /* DirectionalMovementPlus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 5 /* DirectionalMovementMinus */, Shape: shape.Scalar, Role: Directional, Pane: Own},
			{Kind: 6 /* AverageTrueRange */, Shape: shape.Scalar, Role: Volatility, Pane: Own},
			{Kind: 7 /* TrueRange */, Shape: shape.Scalar, Role: Volatility, Pane: Own},
		},
	},
	DirectionalMovementMinus: {
		Identifier: DirectionalMovementMinus, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Directional, Pane: Own}},
	},
	DirectionalMovementPlus: {
		Identifier: DirectionalMovementPlus, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Directional, Pane: Own}},
	},
	NormalizedAverageTrueRange: {
		Identifier: NormalizedAverageTrueRange, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Volatility, Pane: Own}},
	},
	ParabolicStopAndReverse: {
		Identifier: ParabolicStopAndReverse, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Overlay, Pane: Price}},
	},
	RelativeStrengthIndex: {
		Identifier: RelativeStrengthIndex, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: BoundedOscillator, Pane: Own}},
	},
	TrueRange: {
		Identifier: TrueRange, Family: "Welles Wilder",
		Adaptivity: Static, InputRequirement: BarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Scalar, Role: Volatility, Pane: Own}},
	},

	// ── custom ────────────────────────────────────────────────────────────

	GoertzelSpectrum: {
		Identifier: GoertzelSpectrum, Family: "Custom",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Heatmap, Role: Spectrum, Pane: Own}},
	},
	MaximumEntropySpectrum: {
		Identifier: MaximumEntropySpectrum, Family: "Custom",
		Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
		Outputs: []OutputDescriptor{{Kind: 1 /* Value */, Shape: shape.Heatmap, Role: Spectrum, Pane: Own}},
	},
}
