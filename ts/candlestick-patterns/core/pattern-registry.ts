/**
 * PatternInfo contains descriptive metadata for a single candlestick pattern.
 */
export interface PatternInfo {
    /** Human-readable English name. */
    mnemonic: string;
    /** Traditional Japanese name in kanji, or empty when no established Japanese term exists. */
    kanji: string;
    /** Phonetic reading in hiragana, or empty when kanji is empty. */
    reading: string;
    /** Multi-sentence explanation of the pattern, its structure, and market interpretation. */
    description: string;
}

/** Maps PatternIdentifier to PatternInfo for all 61 patterns. */
export const PATTERN_REGISTRY: PatternInfo[] = [
    { // AbandonedBaby (0)
        mnemonic: 'abandoned baby',
        kanji: '捨て子線',
        reading: 'すてごせん',
        description: 'A three-candle reversal pattern. The first candle continues the ' +
            'prevailing trend, the second is a doji that gaps away from the ' +
            'first (shadows do not overlap), and the third candle gaps in the ' +
            'opposite direction and closes well into the first candle\'s body. ' +
            'A bullish abandoned baby appears in a downtrend; a bearish one ' +
            'appears in an uptrend. It is one of the rarest and most reliable ' +
            'reversal signals.',
    },
    { // AdvanceBlock (1)
        mnemonic: 'advance block',
        kanji: '',
        reading: '',
        description: 'A three-candle bearish reversal pattern that appears during an ' +
            'uptrend. Three consecutive white candles with progressively ' +
            'smaller real bodies and increasingly long upper shadows indicate ' +
            'that buying pressure is weakening. Each candle opens within the ' +
            'prior candle\'s body. The diminishing bodies and growing shadows ' +
            'warn of an impending reversal.',
    },
    { // BeltHold (2)
        mnemonic: 'belt hold',
        kanji: '寄り付き坊主',
        reading: 'よりつきぼうず',
        description: 'A single-candle pattern with a long real body and no shadow on ' +
            'the opening side. A bullish belt hold opens at the low of the ' +
            'session and closes near the high, suggesting strong buying from ' +
            'the open. A bearish belt hold opens at the high and closes near ' +
            'the low. The pattern is more significant when it appears after a ' +
            'sustained trend in the opposite direction.',
    },
    { // Breakaway (3)
        mnemonic: 'breakaway',
        kanji: '',
        reading: '',
        description: 'A five-candle reversal pattern. The first candle is a long body ' +
            'in the direction of the trend, followed by three small-body ' +
            'candles that continue the trend with progressively less momentum. ' +
            'The fifth candle is a long body in the opposite direction that ' +
            'closes within the gap created between the first and second ' +
            'candles. It signals that the prior trend has exhausted itself.',
    },
    { // ClosingMarubozu (4)
        mnemonic: 'closing marubozu',
        kanji: '寄り切り坊主',
        reading: 'よりきりぼうず',
        description: 'A single-candle pattern with no shadow on the closing side. A ' +
            'white closing marubozu has no upper shadow, meaning the close ' +
            'equals the high, indicating sustained buying pressure through ' +
            'the session\'s end. A black closing marubozu has no lower shadow, ' +
            'with the close at the low. It is a weaker signal than the full ' +
            'marubozu but still shows conviction in the closing direction.',
    },
    { // ConcealingBabySwallow (5)
        mnemonic: 'concealing baby swallow',
        kanji: '',
        reading: '',
        description: 'A four-candle bullish reversal pattern appearing in a downtrend. ' +
            'The first two candles are black marubozu (no shadows), showing ' +
            'strong selling. The third candle gaps down and has an upper ' +
            'shadow that penetrates the prior candle\'s body, forming a ' +
            '"baby" candle. The fourth black candle fully engulfs the ' +
            'third, including its shadow. This concealment of the baby ' +
            'signals that the downtrend may be ending.',
    },
    { // Counterattack (6)
        mnemonic: 'counterattack',
        kanji: '出会い線',
        reading: 'であいせん',
        description: 'A two-candle reversal pattern where the second candle opens ' +
            'sharply in the direction of the prevailing trend but then ' +
            'reverses to close at approximately the same level as the prior ' +
            'candle\'s close. The opposing colors of the two candles, combined ' +
            'with the matching closes, suggest that the dominant side has met ' +
            'equal and opposite force.',
    },
    { // DarkCloudCover (7)
        mnemonic: 'dark cloud cover',
        kanji: '被せ線',
        reading: 'かぶせせん',
        description: 'A two-candle bearish reversal pattern. The first candle is a ' +
            'long white body, followed by a black candle that opens above ' +
            'the prior candle\'s high and closes below the midpoint of the ' +
            'first candle\'s body. The deeper the penetration, the more ' +
            'bearish the implication. It is the bearish counterpart of ' +
            'the piercing pattern.',
    },
    { // Doji (8)
        mnemonic: 'doji',
        kanji: '同事',
        reading: 'どうじ',
        description: 'A single-candle pattern where the open and close are virtually ' +
            'equal, producing a very small or nonexistent real body. The ' +
            'doji indicates indecision between buyers and sellers. Its ' +
            'significance depends on context: after a long white candle it ' +
            'warns of a potential top, and after a long black candle it hints ' +
            'at a bottom. The shadow lengths further characterize the type ' +
            'of doji (long-legged, dragonfly, gravestone).',
    },
    { // DojiStar (9)
        mnemonic: 'doji star',
        kanji: '同事星',
        reading: 'どうじぼし',
        description: 'A two-candle reversal pattern. The first candle is a long body ' +
            'in the direction of the trend, followed by a doji that gaps ' +
            'away from the first candle\'s body. The gap and the doji\'s ' +
            'indecision signal a possible trend change. A bearish doji star ' +
            'forms in an uptrend; a bullish one forms in a downtrend. It is ' +
            'often the precursor to a morning or evening star.',
    },
    { // DragonflyDoji (10)
        mnemonic: 'dragonfly doji',
        kanji: '蜻蛉同事',
        reading: 'とんぼどうじ',
        description: 'A doji with a long lower shadow and no upper shadow, resembling ' +
            'a dragonfly. The open, close, and high are all at or near the ' +
            'same level. It shows that sellers drove prices down sharply ' +
            'during the session but buyers reclaimed all losses by the ' +
            'close. At the bottom of a downtrend it is a strong bullish ' +
            'reversal signal.',
    },
    { // Engulfing (11)
        mnemonic: 'engulfing',
        kanji: '抱き線',
        reading: 'だきせん',
        description: 'A two-candle reversal pattern where the second candle\'s real ' +
            'body completely engulfs the first candle\'s real body. A bullish ' +
            'engulfing has a white candle engulfing a black candle and ' +
            'appears at the end of a downtrend. A bearish engulfing has a ' +
            'black candle engulfing a white candle at the end of an uptrend. ' +
            'Volume on the engulfing candle adds to the signal\'s strength.',
    },
    { // EveningDojiStar (12)
        mnemonic: 'evening doji star',
        kanji: '宵の明星同事',
        reading: 'よいのみょうじょうどうじ',
        description: 'A three-candle bearish reversal pattern and a stronger form of ' +
            'the evening star. The first candle is a long white body, the ' +
            'second is a doji that gaps above the first, and the third is ' +
            'a long black body that closes well into the first candle\'s ' +
            'body. The doji star underscores the market\'s indecision at ' +
            'the top, making this more significant than a standard evening ' +
            'star.',
    },
    { // EveningStar (13)
        mnemonic: 'evening star',
        kanji: '宵の明星',
        reading: 'よいのみょうじょう',
        description: 'A three-candle bearish reversal pattern. The first candle is a ' +
            'long white body continuing the uptrend. The second is a small ' +
            'body that gaps above the first (the "star"). The third is a ' +
            'long black body that closes into the first candle\'s body. The ' +
            'gap and the small body reveal faltering momentum, and the third ' +
            'candle confirms the reversal.',
    },
    { // GravestoneDoji (14)
        mnemonic: 'gravestone doji',
        kanji: '塔婆同事',
        reading: 'とうばどうじ',
        description: 'A doji with a long upper shadow and no lower shadow, resembling ' +
            'a gravestone. The open, close, and low are all at or near the ' +
            'same level. It shows that buyers pushed prices higher but ' +
            'sellers drove them back down to the open by the close. At the ' +
            'top of an uptrend it is a bearish reversal signal.',
    },
    { // Hammer (15)
        mnemonic: 'hammer',
        kanji: '鎚',
        reading: 'つち',
        description: 'A single-candle bullish reversal pattern at the bottom of a ' +
            'downtrend. It has a small real body at the upper end of the ' +
            'trading range, a long lower shadow at least twice the body ' +
            'length, and little or no upper shadow. The color of the body ' +
            'is less important than the shape. The long lower shadow shows ' +
            'that sellers drove prices down but buyers reclaimed most of the ' +
            'ground, signaling a potential bottom.',
    },
    { // HangingMan (16)
        mnemonic: 'hanging man',
        kanji: '首吊り線',
        reading: 'くびつりせん',
        description: 'A single-candle bearish reversal pattern at the top of an ' +
            'uptrend. It has the same shape as a hammer \u2014 small body, long ' +
            'lower shadow, minimal upper shadow \u2014 but appears after a rise. ' +
            'The long lower shadow reveals intra-session selling pressure ' +
            'that was mostly recovered, but the pattern warns that sellers ' +
            'are becoming active. Confirmation on the next candle is ' +
            'recommended.',
    },
    { // Harami (17)
        mnemonic: 'harami',
        kanji: '孕み線',
        reading: 'はらみせん',
        description: 'A two-candle reversal pattern where the second candle\'s real ' +
            'body is contained within the first candle\'s real body. A ' +
            'bullish harami has a small white (or any color) body within a ' +
            'prior long black body; a bearish harami has a small body within ' +
            'a prior long white body. The pattern indicates that the prior ' +
            'trend\'s momentum is waning. The name means "pregnant" in ' +
            'Japanese, with the first candle as the mother.',
    },
    { // HaramiCross (18)
        mnemonic: 'harami cross',
        kanji: '孕み寄せ線',
        reading: 'はらみよせせん',
        description: 'A stronger variant of the harami where the second candle is a ' +
            'doji rather than a small body. The doji, fully contained within ' +
            'the prior long candle, amplifies the indecision signal. A ' +
            'bullish harami cross at a bottom or a bearish harami cross at ' +
            'a top is considered more significant than a standard harami.',
    },
    { // HighWave (19)
        mnemonic: 'high wave',
        kanji: '高波',
        reading: 'たかなみ',
        description: 'A single-candle pattern characterized by a small real body and ' +
            'very long upper and lower shadows. The extreme shadow lengths ' +
            'relative to the body indicate that the market moved sharply in ' +
            'both directions but ultimately settled near the open. It ' +
            'reflects extreme indecision and, in context, can signal a ' +
            'major turning point.',
    },
    { // Hikkake (20)
        mnemonic: 'hikkake',
        kanji: '',
        reading: '',
        description: 'A multi-candle pattern based on a failed inside bar breakout. ' +
            'An inside bar (lower high, higher low than the prior candle) is ' +
            'followed by a candle that breaks one side of the inside bar\'s ' +
            'range. If within three subsequent bars the price closes beyond ' +
            'the opposite side of the inside bar, the breakout is confirmed ' +
            'as a false move, and a trade in the opposite direction is ' +
            'signaled. The pattern captures trapped traders.',
    },
    { // HikkakeModified (21)
        mnemonic: 'hikkake modified',
        kanji: '',
        reading: '',
        description: 'A refined version of the hikkake with a double inside bar ' +
            'setup. The second candle is inside the first, and the third is ' +
            'inside the second, creating a tighter consolidation. ' +
            'Additionally, the second candle\'s close must be near its ' +
            'extreme (near the low for bullish, near the high for bearish). ' +
            'Confirmation follows the same rule: a close beyond the third ' +
            'candle\'s range within three bars.',
    },
    { // HomingPigeon (22)
        mnemonic: 'homing pigeon',
        kanji: '',
        reading: '',
        description: 'A two-candle bullish reversal pattern. Both candles are black, ' +
            'but the second has a smaller body that is contained within the ' +
            'first candle\'s body. Unlike a harami, both candles must be ' +
            'bearish. The shrinking body suggests that selling pressure is ' +
            'diminishing and the downtrend may be ending.',
    },
    { // IdenticalThreeCrows (23)
        mnemonic: 'identical three crows',
        kanji: '同事三羽',
        reading: 'どうじさんば',
        description: 'A three-candle bearish continuation pattern. Three consecutive ' +
            'black candles, each with very short lower shadows, where each ' +
            'candle opens at approximately the same price as the prior ' +
            'candle\'s close (the "identical" opening). The relentless ' +
            'selling with no gaps between closes and opens signals strong ' +
            'bearish conviction.',
    },
    { // InNeck (24)
        mnemonic: 'in-neck',
        kanji: '入り首線',
        reading: 'いりくびせん',
        description: 'A two-candle bearish continuation pattern. The first candle is ' +
            'a long black body followed by a small white body that opens ' +
            'below the prior candle\'s low and closes at or just barely into ' +
            'the prior candle\'s body. The weak rally confirms that sellers ' +
            'remain in control.',
    },
    { // InvertedHammer (25)
        mnemonic: 'inverted hammer',
        kanji: '逆鎚',
        reading: 'ぎゃくつち',
        description: 'A single-candle bullish reversal pattern at the bottom of a ' +
            'downtrend. It has a small real body at the lower end of the ' +
            'range, a long upper shadow, and little or no lower shadow \u2014 ' +
            'the inverted form of the hammer. Although buyers were unable ' +
            'to maintain the rally, the pattern shows buying interest is ' +
            'emerging. Confirmation on the next candle is recommended.',
    },
    { // Kicking (26)
        mnemonic: 'kicking',
        kanji: '',
        reading: '',
        description: 'A two-candle pattern composed of two marubozu of opposite ' +
            'color that gap apart. A bullish kicking has a black marubozu ' +
            'followed by a white marubozu that gaps above it. A bearish ' +
            'kicking has a white marubozu followed by a black marubozu ' +
            'that gaps below. The marubozu bodies and the gap show ' +
            'extreme conviction in the new direction. It is one of the ' +
            'most powerful candlestick signals.',
    },
    { // KickingByLength (27)
        mnemonic: 'kicking by length',
        kanji: '',
        reading: '',
        description: 'Same structure as the kicking pattern \u2014 two opposite-colored ' +
            'marubozu with a gap \u2014 but the signal direction is determined ' +
            'by which marubozu has the longer real body rather than by the ' +
            'gap direction. The longer candle is considered the dominant ' +
            'force.',
    },
    { // LadderBottom (28)
        mnemonic: 'ladder bottom',
        kanji: '',
        reading: '',
        description: 'A five-candle bullish reversal pattern. The first three candles ' +
            'are consecutive black candles with progressively lower opens ' +
            'and closes, forming a descending "ladder." The fourth candle ' +
            'is black with a notable upper shadow, hinting at buying ' +
            'interest. The fifth candle is white, opens above the fourth\'s ' +
            'body, and closes above the fourth\'s high, confirming the ' +
            'reversal.',
    },
    { // LongLeggedDoji (29)
        mnemonic: 'long-legged doji',
        kanji: '足長同事',
        reading: 'あしながどうじ',
        description: 'A doji with exceptionally long upper and lower shadows, ' +
            'reflecting extreme indecision. The market moved significantly ' +
            'in both directions during the session but opened and closed at ' +
            'nearly the same price. It often signals a major turning point, ' +
            'especially at market tops.',
    },
    { // LongLine (30)
        mnemonic: 'long line',
        kanji: '大陽線・大陰線',
        reading: 'だいようせん・だいいんせん',
        description: 'A single candle with a long real body \u2014 white (bullish) or ' +
            'black (bearish). The long body shows that one side dominated ' +
            'the session. A long white line reflects strong buying; a long ' +
            'black line reflects strong selling. It is the building block ' +
            'for many multi-candle patterns.',
    },
    { // Marubozu (31)
        mnemonic: 'marubozu',
        kanji: '丸坊主',
        reading: 'まるぼうず',
        description: 'A single candle with no shadows at all \u2014 the open equals one ' +
            'extreme and the close equals the other. A white marubozu (open ' +
            'at the low, close at the high) is the strongest bullish candle; ' +
            'a black marubozu (open at the high, close at the low) is the ' +
            'strongest bearish candle. The absence of shadows indicates ' +
            'total dominance by one side throughout the session.',
    },
    { // MatchingLow (32)
        mnemonic: 'matching low',
        kanji: '毛抜き底',
        reading: 'けぬきぞこ',
        description: 'A two-candle bullish reversal pattern. Two consecutive black ' +
            'candles close at the same or nearly the same price. The ' +
            'matching closes establish a support level, suggesting that ' +
            'sellers were unable to push prices lower on the second attempt. ' +
            'It is more significant after a sustained downtrend.',
    },
    { // MatHold (33)
        mnemonic: 'mat hold',
        kanji: '',
        reading: '',
        description: 'A five-candle bullish continuation pattern. The first candle is ' +
            'a long white body, followed by a gap up. The next three candles ' +
            'are small-bodied and drift lower but stay above the first ' +
            'candle\'s body, forming a "mat." The fifth candle is a long ' +
            'white body that closes at a new high. The pattern shows that ' +
            'the pullback was orderly and the uptrend remains intact.',
    },
    { // MorningDojiStar (34)
        mnemonic: 'morning doji star',
        kanji: '明けの明星同事',
        reading: 'あけのみょうじょうどうじ',
        description: 'A three-candle bullish reversal pattern and a stronger form of ' +
            'the morning star. The first candle is a long black body, the ' +
            'second is a doji that gaps below the first, and the third is ' +
            'a long white body that closes well into the first candle\'s ' +
            'body. The doji at the trough emphasizes the turning point, ' +
            'making this more reliable than a standard morning star.',
    },
    { // MorningStar (35)
        mnemonic: 'morning star',
        kanji: '明けの明星',
        reading: 'あけのみょうじょう',
        description: 'A three-candle bullish reversal pattern. The first candle is a ' +
            'long black body continuing the downtrend. The second is a small ' +
            'body that gaps below the first (the "star"). The third is a ' +
            'long white body that closes into the first candle\'s body. The ' +
            'gap and small body signal fading selling pressure, confirmed by ' +
            'the strong third candle.',
    },
    { // OnNeck (36)
        mnemonic: 'on-neck',
        kanji: '当て首線',
        reading: 'あてくびせん',
        description: 'A two-candle bearish continuation pattern. The first candle is ' +
            'a long black body followed by a small white body that opens ' +
            'below the prior candle\'s low and closes at approximately the ' +
            'prior candle\'s low (not into the body). The bounce only reaches ' +
            'the low, confirming continued bearish dominance.',
    },
    { // Piercing (37)
        mnemonic: 'piercing',
        kanji: '切り込み線',
        reading: 'きりこみせん',
        description: 'A two-candle bullish reversal pattern. The first candle is a ' +
            'long black body, followed by a white candle that opens below ' +
            'the prior candle\'s low and closes above the midpoint of the ' +
            'first candle\'s body. The deeper the penetration, the more ' +
            'bullish the signal. It is the bullish counterpart of the dark ' +
            'cloud cover.',
    },
    { // RickshawMan (38)
        mnemonic: 'rickshaw man',
        kanji: '人力車夫',
        reading: 'じんりきしゃふ',
        description: 'A form of long-legged doji where the body is near the center ' +
            'of the candle\'s range, with approximately equal upper and lower ' +
            'shadows. It represents complete equilibrium between buyers and ' +
            'sellers and signals extreme market indecision. At key support ' +
            'or resistance levels it can presage a reversal.',
    },
    { // RisingFallingThreeMethods (39)
        mnemonic: 'rising/falling three methods',
        kanji: '上げ三法・下げ三法',
        reading: 'あげさんぽう・さげさんぽう',
        description: 'A five-candle continuation pattern. In the rising form, a long ' +
            'white candle is followed by three small declining candles that ' +
            'stay within the first candle\'s range, then a final long white ' +
            'candle closes above the first candle\'s high. The falling form ' +
            'is the mirror. The three small candles represent a brief rest ' +
            'within the prevailing trend.',
    },
    { // SeparatingLines (40)
        mnemonic: 'separating lines',
        kanji: '振り分け線',
        reading: 'ふりわけせん',
        description: 'A two-candle continuation pattern where both candles open at ' +
            'the same price but move in opposite directions. A bullish ' +
            'separating line has a black candle followed by a white candle ' +
            'opening at the same level; a bearish version has a white then ' +
            'black candle. The shared opening and divergent closes reaffirm ' +
            'the prevailing trend\'s direction.',
    },
    { // ShootingStar (41)
        mnemonic: 'shooting star',
        kanji: '流れ星',
        reading: 'ながれぼし',
        description: 'A single-candle bearish reversal pattern at the top of an ' +
            'uptrend. It has a small real body at the lower end of the ' +
            'range, a long upper shadow at least twice the body length, ' +
            'and little or no lower shadow \u2014 the inverted form of the ' +
            'hanging man. Buyers pushed prices higher but sellers drove ' +
            'them back down, warning that the uptrend may be ending.',
    },
    { // ShortLine (42)
        mnemonic: 'short line',
        kanji: '小陽線・小陰線',
        reading: 'しょうようせん・しょういんせん',
        description: 'A single candle with a small real body \u2014 white (mildly ' +
            'bullish) or black (mildly bearish). The small body indicates ' +
            'a narrow trading range with limited conviction. It is the ' +
            'opposite of the long line and often appears as part of larger ' +
            'multi-candle patterns where small bodies signal hesitation.',
    },
    { // SpinningTop (43)
        mnemonic: 'spinning top',
        kanji: 'コマ',
        reading: 'こま',
        description: 'A single-candle indecision pattern with a small real body and ' +
            'upper and lower shadows that are longer than the body. It shows ' +
            'that neither buyers nor sellers gained a decisive advantage. ' +
            'After a long white or black candle, a spinning top warns of ' +
            'possible trend exhaustion.',
    },
    { // Stalled (44)
        mnemonic: 'stalled pattern',
        kanji: '',
        reading: '',
        description: 'A three-candle bearish reversal pattern, also called a ' +
            'deliberation pattern. Three consecutive white candles where ' +
            'the third has a notably small body (and may gap up from the ' +
            'second) indicate that the uptrend is stalling. The shrinking ' +
            'third body shows diminishing buying enthusiasm even as prices ' +
            'make new highs.',
    },
    { // StickSandwich (45)
        mnemonic: 'stick sandwich',
        kanji: '',
        reading: '',
        description: 'A three-candle bullish reversal pattern. Two black candles with ' +
            'equal (or nearly equal) closes "sandwich" a white candle in ' +
            'between. The matching closes of the two black candles establish ' +
            'a support level, and the intervening white candle shows buying ' +
            'interest, suggesting the downtrend may reverse.',
    },
    { // Takuri (46)
        mnemonic: 'takuri',
        kanji: 'たくり線',
        reading: 'たくりせん',
        description: 'A single-candle bullish reversal pattern similar to the hammer ' +
            'but with a very long lower shadow (at least three times the ' +
            'body). It appears at the bottom of a downtrend. The extremely ' +
            'long lower shadow shows an aggressive sell-off that was ' +
            'completely recovered, providing a strong reversal signal. ' +
            '"Takuri" means "groping for the bottom" in Japanese.',
    },
    { // TasukiGap (47)
        mnemonic: 'tasuki gap',
        kanji: 'たすき',
        reading: 'たすき',
        description: 'A three-candle continuation pattern. In the bullish form, two ' +
            'white candles with an upward gap are followed by a black candle ' +
            'that opens within the second candle\'s body and closes within ' +
            'the gap but does not fill it. The unfilled gap confirms ' +
            'continuation. The bearish form is the mirror. The gap acts as ' +
            'support (or resistance).',
    },
    { // ThreeBlackCrows (48)
        mnemonic: 'three black crows',
        kanji: '三羽烏',
        reading: 'さんばがらす',
        description: 'A three-candle bearish reversal pattern. Three consecutive long ' +
            'black candles, each opening within the prior candle\'s body and ' +
            'closing at or near its low. The pattern signals a dramatic ' +
            'shift in sentiment from bullish to bearish. Volume typically ' +
            'increases across the three candles.',
    },
    { // ThreeInside (49)
        mnemonic: 'three inside up/down',
        kanji: 'はらみ確認',
        reading: 'はらみかくにん',
        description: 'A three-candle reversal pattern that confirms a harami. The ' +
            'first two candles form a harami (second body inside first), ' +
            'and the third candle closes in the reversal direction \u2014 above ' +
            'the first candle\'s close for bullish (three inside up) or below ' +
            'it for bearish (three inside down). The third candle provides ' +
            'the confirmation that the harami alone lacks.',
    },
    { // ThreeLineStrike (50)
        mnemonic: 'three-line strike',
        kanji: '三本連続線',
        reading: 'さんぼんれんぞくせん',
        description: 'A four-candle continuation pattern. Three consecutive candles ' +
            'in the direction of the trend are followed by a fourth candle ' +
            'that opens further in the trend direction but then reverses ' +
            'and closes beyond the first candle\'s open, "striking" through ' +
            'all three lines. Despite the dramatic fourth candle, the ' +
            'pattern is typically a continuation rather than a reversal.',
    },
    { // ThreeOutside (51)
        mnemonic: 'three outside up/down',
        kanji: '抱き確認',
        reading: 'だきかくにん',
        description: 'A three-candle reversal pattern that confirms an engulfing ' +
            'pattern. The first two candles form a bullish or bearish ' +
            'engulfing, and the third candle continues in the reversal ' +
            'direction \u2014 closing higher for bullish (three outside up) or ' +
            'lower for bearish (three outside down). The third candle adds ' +
            'confirmation to the already strong engulfing signal.',
    },
    { // ThreeStarsInTheSouth (52)
        mnemonic: 'three stars in the south',
        kanji: '南の三つ星',
        reading: 'みなみのみつぼし',
        description: 'A three-candle bullish reversal pattern. Three consecutive ' +
            'black candles with progressively shorter bodies, higher lows, ' +
            'and shrinking shadows. Each candle shows weakening selling ' +
            'pressure. The diminishing range and the "stars" (small bodies) ' +
            'migrating upward within the downtrend signal that sellers are ' +
            'losing control.',
    },
    { // ThreeWhiteSoldiers (53)
        mnemonic: 'three white soldiers',
        kanji: '赤三兵',
        reading: 'あかさんぺい',
        description: 'A three-candle bullish reversal pattern. Three consecutive long ' +
            'white candles, each opening within the prior candle\'s body and ' +
            'closing at or near its high. The pattern signals a strong shift ' +
            'from bearish to bullish sentiment. Short or absent upper ' +
            'shadows strengthen the signal. It is the bullish counterpart ' +
            'of three black crows.',
    },
    { // Thrusting (54)
        mnemonic: 'thrusting',
        kanji: '差し込み線',
        reading: 'さしこみせん',
        description: 'A two-candle bearish continuation pattern. The first candle is ' +
            'a long black body followed by a white candle that opens below ' +
            'the prior candle\'s low and closes into the prior candle\'s body ' +
            'but below its midpoint. The weak rally (weaker than a piercing ' +
            'pattern) suggests that sellers still dominate.',
    },
    { // Tristar (55)
        mnemonic: 'tri-star',
        kanji: '三つ星',
        reading: 'みつぼし',
        description: 'A three-candle reversal pattern composed of three consecutive ' +
            'doji candles. The second doji gaps away from the first and ' +
            'third, forming a star. A bullish tri-star has the second doji ' +
            'gapping below (morning star formation); a bearish tri-star has ' +
            'it gapping above (evening star formation). It is very rare and ' +
            'signals extreme indecision at a turning point.',
    },
    { // TwoCrows (56)
        mnemonic: 'two crows',
        kanji: '二羽の烏',
        reading: 'にわのからす',
        description: 'A three-candle bearish reversal pattern. The first candle is a ' +
            'long white body in an uptrend. The second candle is a small ' +
            'black body that gaps above the first candle\'s close. The third ' +
            'candle is a black body that opens above the second\'s open but ' +
            'closes within the first candle\'s body. The two "crows" ' +
            '(black candles) nesting atop the white candle warn of a top.',
    },
    { // UniqueThreeRiver (57)
        mnemonic: 'unique three river bottom',
        kanji: '',
        reading: '',
        description: 'A three-candle bullish reversal pattern. The first candle is ' +
            'a long black body. The second is a black harami with a long ' +
            'lower shadow that sets a new low. The third is a small white ' +
            'body that closes below the second candle\'s close. The new low ' +
            'on the second candle followed by the failure to sustain it ' +
            'suggests exhaustion of selling.',
    },
    { // UpDownGapSideBySideWhiteLines (58)
        mnemonic: 'up/down-gap side-by-side white lines',
        kanji: '',
        reading: '',
        description: 'A three-candle continuation pattern. Two consecutive white ' +
            'candles of approximately equal size and equal opening prices ' +
            'appear after a gap from the first candle. In the bullish ' +
            'version the gap is upward; in the bearish version the gap is ' +
            'downward. The matching pair of white candles holding the gap ' +
            'confirms the trend\'s continuation.',
    },
    { // UpsideGapTwoCrows (59)
        mnemonic: 'upside gap two crows',
        kanji: '上放れ二羽烏',
        reading: 'うわばなれにわがらす',
        description: 'A three-candle bearish reversal pattern. The first candle is a ' +
            'long white body in an uptrend. The second is a small black body ' +
            'that gaps above the first\'s close. The third is a larger black ' +
            'body that engulfs the second but still closes above the first ' +
            'candle\'s close. The two black candles "cawing" above the ' +
            'uptrend warn that sellers are gaining strength.',
    },
    { // XSideGapThreeMethods (60)
        mnemonic: 'upside/downside gap three methods',
        kanji: '',
        reading: '',
        description: 'A three-candle continuation pattern. Two candles of the same ' +
            'color establish a gap in the direction of the trend. The third ' +
            'candle is the opposite color, opens within the second candle\'s ' +
            'body, and closes within the first candle\'s body, partially ' +
            'filling the gap. Despite the partial fill, the gap is not ' +
            'closed, confirming the prevailing trend\'s continuation.',
    },
];
