package core

// PatternIdentifier identifies each of the 61 candlestick patterns.
// Values are sequential starting at 0, sorted alphabetically.
type PatternIdentifier int

const (
	AbandonedBaby                     PatternIdentifier = iota // 0
	AdvanceBlock                                               // 1
	BeltHold                                                   // 2
	Breakaway                                                  // 3
	ClosingMarubozu                                            // 4
	ConcealingBabySwallow                                      // 5
	Counterattack                                              // 6
	DarkCloudCover                                             // 7
	Doji                                                       // 8
	DojiStar                                                   // 9
	DragonflyDoji                                              // 10
	Engulfing                                                  // 11
	EveningDojiStar                                            // 12
	EveningStar                                                // 13
	GravestoneDoji                                             // 14
	Hammer                                                     // 15
	HangingMan                                                 // 16
	Harami                                                     // 17
	HaramiCross                                                // 18
	HighWave                                                   // 19
	Hikkake                                                    // 20
	HikkakeModified                                            // 21
	HomingPigeon                                               // 22
	IdenticalThreeCrows                                        // 23
	InNeck                                                     // 24
	InvertedHammer                                             // 25
	Kicking                                                    // 26
	KickingByLength                                            // 27
	LadderBottom                                               // 28
	LongLeggedDoji                                             // 29
	LongLine                                                   // 30
	Marubozu                                                   // 31
	MatchingLow                                                // 32
	MatHold                                                    // 33
	MorningDojiStar                                            // 34
	MorningStar                                                // 35
	OnNeck                                                     // 36
	Piercing                                                   // 37
	RickshawMan                                                // 38
	RisingFallingThreeMethods                                  // 39
	SeparatingLines                                            // 40
	ShootingStar                                               // 41
	ShortLine                                                  // 42
	SpinningTop                                                // 43
	Stalled                                                    // 44
	StickSandwich                                              // 45
	Takuri                                                     // 46
	TasukiGap                                                  // 47
	ThreeBlackCrows                                            // 48
	ThreeInside                                                // 49
	ThreeLineStrike                                            // 50
	ThreeOutside                                               // 51
	ThreeStarsInTheSouth                                       // 52
	ThreeWhiteSoldiers                                         // 53
	Thrusting                                                  // 54
	Tristar                                                    // 55
	TwoCrows                                                   // 56
	UniqueThreeRiver                                           // 57
	UpDownGapSideBySideWhiteLines                              // 58
	UpsideGapTwoCrows                                          // 59
	XSideGapThreeMethods                                       // 60
)

// PatternCount is the total number of patterns.
const PatternCount = 61

// MethodName returns the snake_case method name (matching the Python convention).
func (p PatternIdentifier) MethodName() string {
	if int(p) < len(patternNames) {
		return patternNames[p]
	}
	return ""
}

var patternNames = [PatternCount]string{
	"abandoned_baby",
	"advance_block",
	"belt_hold",
	"breakaway",
	"closing_marubozu",
	"concealing_baby_swallow",
	"counterattack",
	"dark_cloud_cover",
	"doji",
	"doji_star",
	"dragonfly_doji",
	"engulfing",
	"evening_doji_star",
	"evening_star",
	"gravestone_doji",
	"hammer",
	"hanging_man",
	"harami",
	"harami_cross",
	"high_wave",
	"hikkake",
	"hikkake_modified",
	"homing_pigeon",
	"identical_three_crows",
	"in_neck",
	"inverted_hammer",
	"kicking",
	"kicking_by_length",
	"ladder_bottom",
	"long_legged_doji",
	"long_line",
	"marubozu",
	"matching_low",
	"mat_hold",
	"morning_doji_star",
	"morning_star",
	"on_neck",
	"piercing",
	"rickshaw_man",
	"rising_falling_three_methods",
	"separating_lines",
	"shooting_star",
	"short_line",
	"spinning_top",
	"stalled",
	"stick_sandwich",
	"takuri",
	"tasuki_gap",
	"three_black_crows",
	"three_inside",
	"three_line_strike",
	"three_outside",
	"three_stars_in_the_south",
	"three_white_soldiers",
	"thrusting",
	"tristar",
	"two_crows",
	"unique_three_river",
	"up_down_gap_side_by_side_white_lines",
	"upside_gap_two_crows",
	"x_side_gap_three_methods",
}
