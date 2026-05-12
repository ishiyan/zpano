// Package candlestickpatterns provides streaming candlestick pattern recognition
// with fuzzy logic support.
//
// Usage:
//
//	cp := candlestickpatterns.New(nil)
//	for _, bar := range bars {
//	    cp.Update(bar.Open, bar.High, bar.Low, bar.Close)
//	    result := cp.AbandonedBaby()  // continuous float in [-100, +100]
//	}
//
// Each pattern method returns a continuous confidence value in [-100, +100],
// where positive values indicate bullish signals and negative values indicate
// bearish signals. The magnitude reflects the fuzzy confidence of the match.
// Use [fuzzy.AlphaCut] to convert to crisp {-100, 0, +100} if needed.
//
// The engine inspects the most recent N bars (stored in a ring buffer)
// and the incrementally maintained running totals for each criterion, giving
// O(1) per bar after the warmup period.
package candlestickpatterns

import (
	"zpano/candlestickpatterns/core"
	"zpano/candlestickpatterns/patterns"
)

// Re-export core types so callers can use the root package.
type (
	OHLC              = core.OHLC
	Criterion         = core.Criterion
	CriterionState    = core.CriterionState
	Options           = core.Options
	RangeEntity       = core.RangeEntity
	PatternIdentifier = core.PatternIdentifier
	PatternInfo       = core.PatternInfo
)

// Re-export RangeEntity constants.
const (
	RealBody RangeEntity = core.RealBody
	HighLow  RangeEntity = core.HighLow
	Shadows  RangeEntity = core.Shadows
)

// Re-export pattern count.
const PatternCount = core.PatternCount

// Re-export default criteria.
var (
	DefaultLongBody        = core.DefaultLongBody
	DefaultVeryLongBody    = core.DefaultVeryLongBody
	DefaultShortBody       = core.DefaultShortBody
	DefaultDojiBody        = core.DefaultDojiBody
	DefaultLongShadow      = core.DefaultLongShadow
	DefaultVeryLongShadow  = core.DefaultVeryLongShadow
	DefaultShortShadow     = core.DefaultShortShadow
	DefaultVeryShortShadow = core.DefaultVeryShortShadow
	DefaultNear            = core.DefaultNear
	DefaultFar             = core.DefaultFar
	DefaultEqual           = core.DefaultEqual
)

// Re-export pattern registry.
var PatternRegistry = core.PatternRegistry

// CandlestickPatterns is the candlestick pattern recognition engine.
//
// Provides streaming bar-by-bar evaluation of 61 Japanese candlestick patterns.
// Call Update(open, high, low, close) for each new bar, then call any pattern
// method to get the result for the current bar.
//
// Pattern methods return a continuous float in [-100, +100]:
//
//	positive: bullish signal, negative: bearish signal, near zero: no match.
//	The magnitude reflects the fuzzy confidence of the match.
//	Hikkake and HikkakeModified may return intermediate values for
//	unconfirmed signals.
type CandlestickPatterns struct {
	engine *core.CandlestickPatterns
}

// New creates a new CandlestickPatterns engine with the given options.
// Pass nil for default options.
func New(opts *Options) *CandlestickPatterns {
	return &CandlestickPatterns{engine: core.New(opts)}
}

// Update feeds a new OHLC bar into the engine.
// After calling this, all pattern methods reflect the state including this bar.
func (cp *CandlestickPatterns) Update(o, h, l, c float64) {
	cp.engine.UpdateBar(o, h, l, c)
	// Reset and update stateful patterns.
	cp.engine.HikmodConfirmed = false
	cp.engine.HikmodLastSignal = 0
	cp.engine.HikkakeModifiedUpdate()
}

// Count returns the number of bars fed so far.
func (cp *CandlestickPatterns) Count() int {
	return cp.engine.Count
}

// Evaluate evaluates a single pattern by its identifier.
func (cp *CandlestickPatterns) Evaluate(id PatternIdentifier) float64 {
	return patternDispatch[id](cp)
}

// ---------------------------------------------------------------------------
// Pattern methods — each delegates to the standalone function in patterns/.
// Names preserve the original API (PatternDoji, PatternEngulfing, etc.).
// ---------------------------------------------------------------------------

func (cp *CandlestickPatterns) AbandonedBaby() float64              { return patterns.AbandonedBaby(cp.engine) }
func (cp *CandlestickPatterns) AdvanceBlock() float64               { return patterns.AdvanceBlock(cp.engine) }
func (cp *CandlestickPatterns) BeltHold() float64                   { return patterns.BeltHold(cp.engine) }
func (cp *CandlestickPatterns) Breakaway() float64                  { return patterns.Breakaway(cp.engine) }
func (cp *CandlestickPatterns) ClosingMarubozu() float64            { return patterns.ClosingMarubozu(cp.engine) }
func (cp *CandlestickPatterns) ConcealingBabySwallow() float64      { return patterns.ConcealingBabySwallow(cp.engine) }
func (cp *CandlestickPatterns) Counterattack() float64              { return patterns.Counterattack(cp.engine) }
func (cp *CandlestickPatterns) DarkCloudCover() float64             { return patterns.DarkCloudCover(cp.engine) }
func (cp *CandlestickPatterns) PatternDoji() float64                { return patterns.Doji(cp.engine) }
func (cp *CandlestickPatterns) DojiStar() float64                   { return patterns.DojiStar(cp.engine) }
func (cp *CandlestickPatterns) DragonflyDoji() float64              { return patterns.DragonflyDoji(cp.engine) }
func (cp *CandlestickPatterns) PatternEngulfing() float64           { return patterns.Engulfing(cp.engine) }
func (cp *CandlestickPatterns) EveningDojiStar() float64            { return patterns.EveningDojiStar(cp.engine) }
func (cp *CandlestickPatterns) EveningStar() float64                { return patterns.EveningStar(cp.engine) }
func (cp *CandlestickPatterns) GravestoneDoji() float64             { return patterns.GravestoneDoji(cp.engine) }
func (cp *CandlestickPatterns) PatternHammer() float64              { return patterns.Hammer(cp.engine) }
func (cp *CandlestickPatterns) HangingMan() float64                 { return patterns.HangingMan(cp.engine) }
func (cp *CandlestickPatterns) PatternHarami() float64              { return patterns.Harami(cp.engine) }
func (cp *CandlestickPatterns) HaramiCross() float64                { return patterns.HaramiCross(cp.engine) }
func (cp *CandlestickPatterns) HighWave() float64                   { return patterns.HighWave(cp.engine) }
func (cp *CandlestickPatterns) PatternHikkake() float64             { return patterns.Hikkake(cp.engine) }
func (cp *CandlestickPatterns) HikkakeModified() float64            { return patterns.HikkakeModified(cp.engine) }
func (cp *CandlestickPatterns) HomingPigeon() float64               { return patterns.HomingPigeon(cp.engine) }
func (cp *CandlestickPatterns) IdenticalThreeCrows() float64        { return patterns.IdenticalThreeCrows(cp.engine) }
func (cp *CandlestickPatterns) InNeck() float64                     { return patterns.InNeck(cp.engine) }
func (cp *CandlestickPatterns) InvertedHammer() float64             { return patterns.InvertedHammer(cp.engine) }
func (cp *CandlestickPatterns) PatternKicking() float64             { return patterns.Kicking(cp.engine) }
func (cp *CandlestickPatterns) PatternKickingByLength() float64     { return patterns.KickingByLength(cp.engine) }
func (cp *CandlestickPatterns) LadderBottom() float64               { return patterns.LadderBottom(cp.engine) }
func (cp *CandlestickPatterns) LongLeggedDoji() float64             { return patterns.LongLeggedDoji(cp.engine) }
func (cp *CandlestickPatterns) LongLine() float64                   { return patterns.LongLine(cp.engine) }
func (cp *CandlestickPatterns) PatternMarubozu() float64            { return patterns.Marubozu(cp.engine) }
func (cp *CandlestickPatterns) MatchingLow() float64                { return patterns.MatchingLow(cp.engine) }
func (cp *CandlestickPatterns) MatHold() float64                    { return patterns.MatHold(cp.engine) }
func (cp *CandlestickPatterns) MorningDojiStar() float64            { return patterns.MorningDojiStar(cp.engine) }
func (cp *CandlestickPatterns) MorningStar() float64                { return patterns.MorningStar(cp.engine) }
func (cp *CandlestickPatterns) OnNeck() float64                     { return patterns.OnNeck(cp.engine) }
func (cp *CandlestickPatterns) PatternPiercing() float64            { return patterns.Piercing(cp.engine) }
func (cp *CandlestickPatterns) RickshawMan() float64                { return patterns.RickshawMan(cp.engine) }
func (cp *CandlestickPatterns) RisingFallingThreeMethods() float64  { return patterns.RisingFallingThreeMethods(cp.engine) }
func (cp *CandlestickPatterns) SeparatingLines() float64            { return patterns.SeparatingLines(cp.engine) }
func (cp *CandlestickPatterns) ShootingStar() float64               { return patterns.ShootingStar(cp.engine) }
func (cp *CandlestickPatterns) ShortLine() float64                  { return patterns.ShortLine(cp.engine) }
func (cp *CandlestickPatterns) SpinningTop() float64                { return patterns.SpinningTop(cp.engine) }
func (cp *CandlestickPatterns) Stalled() float64                    { return patterns.Stalled(cp.engine) }
func (cp *CandlestickPatterns) StickSandwich() float64              { return patterns.StickSandwich(cp.engine) }
func (cp *CandlestickPatterns) PatternTakuri() float64              { return patterns.Takuri(cp.engine) }
func (cp *CandlestickPatterns) TasukiGap() float64                  { return patterns.TasukiGap(cp.engine) }
func (cp *CandlestickPatterns) ThreeBlackCrows() float64            { return patterns.ThreeBlackCrows(cp.engine) }
func (cp *CandlestickPatterns) ThreeInside() float64                { return patterns.ThreeInside(cp.engine) }
func (cp *CandlestickPatterns) ThreeLineStrike() float64            { return patterns.ThreeLineStrike(cp.engine) }
func (cp *CandlestickPatterns) ThreeOutside() float64               { return patterns.ThreeOutside(cp.engine) }
func (cp *CandlestickPatterns) ThreeStarsInTheSouth() float64       { return patterns.ThreeStarsInTheSouth(cp.engine) }
func (cp *CandlestickPatterns) ThreeWhiteSoldiers() float64         { return patterns.ThreeWhiteSoldiers(cp.engine) }
func (cp *CandlestickPatterns) PatternThrusting() float64           { return patterns.Thrusting(cp.engine) }
func (cp *CandlestickPatterns) PatternTristar() float64             { return patterns.Tristar(cp.engine) }
func (cp *CandlestickPatterns) TwoCrows() float64                   { return patterns.TwoCrows(cp.engine) }
func (cp *CandlestickPatterns) UniqueThreeRiver() float64           { return patterns.UniqueThreeRiver(cp.engine) }
func (cp *CandlestickPatterns) UpDownGapSideBySideWhiteLines() float64 { return patterns.UpDownGapSideBySideWhiteLines(cp.engine) }
func (cp *CandlestickPatterns) UpsideGapTwoCrows() float64          { return patterns.UpsideGapTwoCrows(cp.engine) }
func (cp *CandlestickPatterns) XSideGapThreeMethods() float64       { return patterns.XSideGapThreeMethods(cp.engine) }

// patternDispatch is the dispatch table: PatternIdentifier → method.
var patternDispatch [PatternCount]func(*CandlestickPatterns) float64

func init() {
	patternDispatch = [PatternCount]func(*CandlestickPatterns) float64{
		(*CandlestickPatterns).AbandonedBaby,
		(*CandlestickPatterns).AdvanceBlock,
		(*CandlestickPatterns).BeltHold,
		(*CandlestickPatterns).Breakaway,
		(*CandlestickPatterns).ClosingMarubozu,
		(*CandlestickPatterns).ConcealingBabySwallow,
		(*CandlestickPatterns).Counterattack,
		(*CandlestickPatterns).DarkCloudCover,
		(*CandlestickPatterns).PatternDoji,
		(*CandlestickPatterns).DojiStar,
		(*CandlestickPatterns).DragonflyDoji,
		(*CandlestickPatterns).PatternEngulfing,
		(*CandlestickPatterns).EveningDojiStar,
		(*CandlestickPatterns).EveningStar,
		(*CandlestickPatterns).GravestoneDoji,
		(*CandlestickPatterns).PatternHammer,
		(*CandlestickPatterns).HangingMan,
		(*CandlestickPatterns).PatternHarami,
		(*CandlestickPatterns).HaramiCross,
		(*CandlestickPatterns).HighWave,
		(*CandlestickPatterns).PatternHikkake,
		(*CandlestickPatterns).HikkakeModified,
		(*CandlestickPatterns).HomingPigeon,
		(*CandlestickPatterns).IdenticalThreeCrows,
		(*CandlestickPatterns).InNeck,
		(*CandlestickPatterns).InvertedHammer,
		(*CandlestickPatterns).PatternKicking,
		(*CandlestickPatterns).PatternKickingByLength,
		(*CandlestickPatterns).LadderBottom,
		(*CandlestickPatterns).LongLeggedDoji,
		(*CandlestickPatterns).LongLine,
		(*CandlestickPatterns).PatternMarubozu,
		(*CandlestickPatterns).MatchingLow,
		(*CandlestickPatterns).MatHold,
		(*CandlestickPatterns).MorningDojiStar,
		(*CandlestickPatterns).MorningStar,
		(*CandlestickPatterns).OnNeck,
		(*CandlestickPatterns).PatternPiercing,
		(*CandlestickPatterns).RickshawMan,
		(*CandlestickPatterns).RisingFallingThreeMethods,
		(*CandlestickPatterns).SeparatingLines,
		(*CandlestickPatterns).ShootingStar,
		(*CandlestickPatterns).ShortLine,
		(*CandlestickPatterns).SpinningTop,
		(*CandlestickPatterns).Stalled,
		(*CandlestickPatterns).StickSandwich,
		(*CandlestickPatterns).PatternTakuri,
		(*CandlestickPatterns).TasukiGap,
		(*CandlestickPatterns).ThreeBlackCrows,
		(*CandlestickPatterns).ThreeInside,
		(*CandlestickPatterns).ThreeLineStrike,
		(*CandlestickPatterns).ThreeOutside,
		(*CandlestickPatterns).ThreeStarsInTheSouth,
		(*CandlestickPatterns).ThreeWhiteSoldiers,
		(*CandlestickPatterns).PatternThrusting,
		(*CandlestickPatterns).PatternTristar,
		(*CandlestickPatterns).TwoCrows,
		(*CandlestickPatterns).UniqueThreeRiver,
		(*CandlestickPatterns).UpDownGapSideBySideWhiteLines,
		(*CandlestickPatterns).UpsideGapTwoCrows,
		(*CandlestickPatterns).XSideGapThreeMethods,
	}
}
