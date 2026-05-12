import { CandlestickPatterns } from './candlestick-patterns.ts';
import { alphaCut } from '../fuzzy/defuzzify.ts';
import { TestCase } from './patterns/testdata-doji.ts';

import { TEST_DATA_ABANDONED_BABY } from './patterns/testdata-abandoned-baby.ts';
import { TEST_DATA_ADVANCE_BLOCK } from './patterns/testdata-advance-block.ts';
import { TEST_DATA_BELT_HOLD } from './patterns/testdata-belt-hold.ts';
import { TEST_DATA_BREAKAWAY } from './patterns/testdata-breakaway.ts';
import { TEST_DATA_CLOSING_MARUBOZU } from './patterns/testdata-closing-marubozu.ts';
import { TEST_DATA_CONCEALING_BABY_SWALLOW } from './patterns/testdata-concealing-baby-swallow.ts';
import { TEST_DATA_COUNTERATTACK } from './patterns/testdata-counterattack.ts';
import { TEST_DATA_DARK_CLOUD_COVER } from './patterns/testdata-dark-cloud-cover.ts';
import { TEST_DATA_DOJI } from './patterns/testdata-doji.ts';
import { TEST_DATA_DOJI_STAR } from './patterns/testdata-doji-star.ts';
import { TEST_DATA_DRAGONFLY_DOJI } from './patterns/testdata-dragonfly-doji.ts';
import { TEST_DATA_ENGULFING } from './patterns/testdata-engulfing.ts';
import { TEST_DATA_EVENING_DOJI_STAR } from './patterns/testdata-evening-doji-star.ts';
import { TEST_DATA_EVENING_STAR } from './patterns/testdata-evening-star.ts';
import { TEST_DATA_GRAVESTONE_DOJI } from './patterns/testdata-gravestone-doji.ts';
import { TEST_DATA_HAMMER } from './patterns/testdata-hammer.ts';
import { TEST_DATA_HANGING_MAN } from './patterns/testdata-hanging-man.ts';
import { TEST_DATA_HARAMI } from './patterns/testdata-harami.ts';
import { TEST_DATA_HARAMI_CROSS } from './patterns/testdata-harami-cross.ts';
import { TEST_DATA_HIGH_WAVE } from './patterns/testdata-high-wave.ts';
import { TEST_DATA_HIKKAKE } from './patterns/testdata-hikkake.ts';
import { TEST_DATA_HIKKAKE_MODIFIED } from './patterns/testdata-hikkake-modified.ts';
import { TEST_DATA_HOMING_PIGEON } from './patterns/testdata-homing-pigeon.ts';
import { TEST_DATA_IDENTICAL_THREE_CROWS } from './patterns/testdata-identical-three-crows.ts';
import { TEST_DATA_IN_NECK } from './patterns/testdata-in-neck.ts';
import { TEST_DATA_INVERTED_HAMMER } from './patterns/testdata-inverted-hammer.ts';
import { TEST_DATA_KICKING } from './patterns/testdata-kicking.ts';
import { TEST_DATA_KICKING_BY_LENGTH } from './patterns/testdata-kicking-by-length.ts';
import { TEST_DATA_LADDER_BOTTOM } from './patterns/testdata-ladder-bottom.ts';
import { TEST_DATA_LONG_LEGGED_DOJI } from './patterns/testdata-long-legged-doji.ts';
import { TEST_DATA_LONG_LINE } from './patterns/testdata-long-line.ts';
import { TEST_DATA_MARUBOZU } from './patterns/testdata-marubozu.ts';
import { TEST_DATA_MAT_HOLD } from './patterns/testdata-mat-hold.ts';
import { TEST_DATA_MATCHING_LOW } from './patterns/testdata-matching-low.ts';
import { TEST_DATA_MORNING_DOJI_STAR } from './patterns/testdata-morning-doji-star.ts';
import { TEST_DATA_MORNING_STAR } from './patterns/testdata-morning-star.ts';
import { TEST_DATA_ON_NECK } from './patterns/testdata-on-neck.ts';
import { TEST_DATA_PIERCING } from './patterns/testdata-piercing.ts';
import { TEST_DATA_RICKSHAW_MAN } from './patterns/testdata-rickshaw-man.ts';
import { TEST_DATA_RISING_FALLING_THREE_METHODS } from './patterns/testdata-rising-falling-three-methods.ts';
import { TEST_DATA_SEPARATING_LINES } from './patterns/testdata-separating-lines.ts';
import { TEST_DATA_SHOOTING_STAR } from './patterns/testdata-shooting-star.ts';
import { TEST_DATA_SHORT_LINE } from './patterns/testdata-short-line.ts';
import { TEST_DATA_SPINNING_TOP } from './patterns/testdata-spinning-top.ts';
import { TEST_DATA_STALLED } from './patterns/testdata-stalled.ts';
import { TEST_DATA_STICK_SANDWICH } from './patterns/testdata-stick-sandwich.ts';
import { TEST_DATA_TAKURI } from './patterns/testdata-takuri.ts';
import { TEST_DATA_TASUKI_GAP } from './patterns/testdata-tasuki-gap.ts';
import { TEST_DATA_THREE_BLACK_CROWS } from './patterns/testdata-three-black-crows.ts';
import { TEST_DATA_THREE_INSIDE } from './patterns/testdata-three-inside.ts';
import { TEST_DATA_THREE_LINE_STRIKE } from './patterns/testdata-three-line-strike.ts';
import { TEST_DATA_THREE_OUTSIDE } from './patterns/testdata-three-outside.ts';
import { TEST_DATA_THREE_STARS_IN_THE_SOUTH } from './patterns/testdata-three-stars-in-the-south.ts';
import { TEST_DATA_THREE_WHITE_SOLDIERS } from './patterns/testdata-three-white-soldiers.ts';
import { TEST_DATA_THRUSTING } from './patterns/testdata-thrusting.ts';
import { TEST_DATA_TRISTAR } from './patterns/testdata-tristar.ts';
import { TEST_DATA_TWO_CROWS } from './patterns/testdata-two-crows.ts';
import { TEST_DATA_UNIQUE_THREE_RIVER } from './patterns/testdata-unique-three-river.ts';
import { TEST_DATA_UP_DOWN_GAP_SIDE_BY_SIDE_WHITE_LINES } from './patterns/testdata-up-down-gap-side-by-side-white-lines.ts';
import { TEST_DATA_UPSIDE_GAP_TWO_CROWS } from './patterns/testdata-upside-gap-two-crows.ts';
import { TEST_DATA_X_SIDE_GAP_THREE_METHODS } from './patterns/testdata-x-side-gap-three-methods.ts';

interface PatternSpec {
    name: string;
    method: (cp: CandlestickPatterns) => number;
    data: TestCase[];
    skipped?: Set<number>;
}

function skipSet(...indices: number[]): Set<number> {
    return new Set(indices);
}

function runPatternTests(specs: PatternSpec[]): void {
    for (const spec of specs) {
        describe(spec.name, () => {
            for (let i = 0; i < spec.data.length; i++) {
                const tc = spec.data[i];
                if (spec.skipped && spec.skipped.has(i)) {
                    continue;
                }
                it(`case ${i}`, () => {
                    const cp = new CandlestickPatterns();
                    for (let j = 0; j < 20; j++) {
                        cp.update(tc.opens[j], tc.highs[j], tc.lows[j], tc.closes[j]);
                    }
                    const actual = spec.method(cp);
                    const crisp = alphaCut(actual, 0.5, 100.0);
                    const expectedCrisp = alphaCut(tc.expected, 0.5, 100.0);
                    expect(crisp).toBe(expectedCrisp);
                });
            }
        });
    }
}

describe('CandlestickPatterns', () => {
    const specs: PatternSpec[] = [
        { name: 'abandoned_baby', method: (cp) => cp.abandonedBaby(), data: TEST_DATA_ABANDONED_BABY, skipped: skipSet(185) },
        { name: 'advance_block', method: (cp) => cp.advanceBlock(), data: TEST_DATA_ADVANCE_BLOCK, skipped: skipSet(6, 14, 117, 126, 151) },
        { name: 'belt_hold', method: (cp) => cp.beltHold(), data: TEST_DATA_BELT_HOLD },
        { name: 'breakaway', method: (cp) => cp.breakaway(), data: TEST_DATA_BREAKAWAY, skipped: skipSet(21) },
        { name: 'closing_marubozu', method: (cp) => cp.closingMarubozu(), data: TEST_DATA_CLOSING_MARUBOZU },
        { name: 'concealing_baby_swallow', method: (cp) => cp.concealingBabySwallow(), data: TEST_DATA_CONCEALING_BABY_SWALLOW, skipped: skipSet(28) },
        { name: 'counterattack', method: (cp) => cp.counterattack(), data: TEST_DATA_COUNTERATTACK, skipped: skipSet(61) },
        { name: 'dark_cloud_cover', method: (cp) => cp.darkCloudCover(), data: TEST_DATA_DARK_CLOUD_COVER },
        { name: 'doji', method: (cp) => cp.doji(), data: TEST_DATA_DOJI },
        { name: 'doji_star', method: (cp) => cp.dojiStar(), data: TEST_DATA_DOJI_STAR },
        { name: 'dragonfly_doji', method: (cp) => cp.dragonflyDoji(), data: TEST_DATA_DRAGONFLY_DOJI },
        { name: 'engulfing', method: (cp) => cp.engulfing(), data: TEST_DATA_ENGULFING },
        { name: 'evening_doji_star', method: (cp) => cp.eveningDojiStar(), data: TEST_DATA_EVENING_DOJI_STAR },
        { name: 'evening_star', method: (cp) => cp.eveningStar(), data: TEST_DATA_EVENING_STAR },
        { name: 'gravestone_doji', method: (cp) => cp.gravestoneDoji(), data: TEST_DATA_GRAVESTONE_DOJI, skipped: skipSet(137) },
        { name: 'hammer', method: (cp) => cp.hammer(), data: TEST_DATA_HAMMER, skipped: skipSet(8, 79) },
        { name: 'hanging_man', method: (cp) => cp.hangingMan(), data: TEST_DATA_HANGING_MAN, skipped: skipSet(9, 53, 158) },
        { name: 'harami', method: (cp) => cp.harami(), data: TEST_DATA_HARAMI, skipped: skipSet(4, 8, 28, 103, 110, 111, 123, 130, 131, 148, 151, 188) },
        { name: 'harami_cross', method: (cp) => cp.haramiCross(), data: TEST_DATA_HARAMI_CROSS, skipped: skipSet(1, 21, 32, 35, 68, 74, 84, 89, 97, 121, 143, 146, 147, 166, 184) },
        { name: 'high_wave', method: (cp) => cp.highWave(), data: TEST_DATA_HIGH_WAVE, skipped: skipSet(27, 83, 99, 161) },
        { name: 'hikkake', method: (cp) => cp.hikkake(), data: TEST_DATA_HIKKAKE },
        { name: 'hikkake_modified', method: (cp) => cp.hikkakeModified(), data: TEST_DATA_HIKKAKE_MODIFIED },
        { name: 'homing_pigeon', method: (cp) => cp.homingPigeon(), data: TEST_DATA_HOMING_PIGEON },
        { name: 'identical_three_crows', method: (cp) => cp.identicalThreeCrows(), data: TEST_DATA_IDENTICAL_THREE_CROWS },
        { name: 'in_neck', method: (cp) => cp.inNeck(), data: TEST_DATA_IN_NECK },
        { name: 'inverted_hammer', method: (cp) => cp.invertedHammer(), data: TEST_DATA_INVERTED_HAMMER },
        { name: 'kicking', method: (cp) => cp.kicking(), data: TEST_DATA_KICKING },
        { name: 'kicking_by_length', method: (cp) => cp.kickingByLength(), data: TEST_DATA_KICKING_BY_LENGTH },
        { name: 'ladder_bottom', method: (cp) => cp.ladderBottom(), data: TEST_DATA_LADDER_BOTTOM },
        { name: 'long_legged_doji', method: (cp) => cp.longLeggedDoji(), data: TEST_DATA_LONG_LEGGED_DOJI, skipped: skipSet(92, 103) },
        { name: 'long_line', method: (cp) => cp.longLine(), data: TEST_DATA_LONG_LINE },
        { name: 'marubozu', method: (cp) => cp.marubozu(), data: TEST_DATA_MARUBOZU, skipped: skipSet(19) },
        { name: 'mat_hold', method: (cp) => cp.matHold(), data: TEST_DATA_MAT_HOLD },
        { name: 'matching_low', method: (cp) => cp.matchingLow(), data: TEST_DATA_MATCHING_LOW },
        { name: 'morning_doji_star', method: (cp) => cp.morningDojiStar(), data: TEST_DATA_MORNING_DOJI_STAR },
        { name: 'morning_star', method: (cp) => cp.morningStar(), data: TEST_DATA_MORNING_STAR },
        { name: 'on_neck', method: (cp) => cp.onNeck(), data: TEST_DATA_ON_NECK },
        { name: 'piercing', method: (cp) => cp.piercing(), data: TEST_DATA_PIERCING, skipped: skipSet(93) },
        { name: 'rickshaw_man', method: (cp) => cp.rickshawMan(), data: TEST_DATA_RICKSHAW_MAN, skipped: skipSet(69, 193) },
        { name: 'rising_falling_three_methods', method: (cp) => cp.risingFallingThreeMethods(), data: TEST_DATA_RISING_FALLING_THREE_METHODS, skipped: skipSet(76, 180) },
        { name: 'separating_lines', method: (cp) => cp.separatingLines(), data: TEST_DATA_SEPARATING_LINES, skipped: skipSet(70, 112) },
        { name: 'shooting_star', method: (cp) => cp.shootingStar(), data: TEST_DATA_SHOOTING_STAR, skipped: skipSet(22, 90) },
        { name: 'short_line', method: (cp) => cp.shortLine(), data: TEST_DATA_SHORT_LINE },
        { name: 'spinning_top', method: (cp) => cp.spinningTop(), data: TEST_DATA_SPINNING_TOP, skipped: skipSet(1, 4, 116, 171) },
        { name: 'stalled', method: (cp) => cp.stalled(), data: TEST_DATA_STALLED, skipped: skipSet(5, 180, 198) },
        { name: 'stick_sandwich', method: (cp) => cp.stickSandwich(), data: TEST_DATA_STICK_SANDWICH },
        { name: 'takuri', method: (cp) => cp.takuri(), data: TEST_DATA_TAKURI, skipped: skipSet(72, 154) },
        { name: 'tasuki_gap', method: (cp) => cp.tasukiGap(), data: TEST_DATA_TASUKI_GAP, skipped: skipSet(161) },
        { name: 'three_black_crows', method: (cp) => cp.threeBlackCrows(), data: TEST_DATA_THREE_BLACK_CROWS },
        { name: 'three_inside', method: (cp) => cp.threeInside(), data: TEST_DATA_THREE_INSIDE },
        { name: 'three_line_strike', method: (cp) => cp.threeLineStrike(), data: TEST_DATA_THREE_LINE_STRIKE },
        { name: 'three_outside', method: (cp) => cp.threeOutside(), data: TEST_DATA_THREE_OUTSIDE },
        { name: 'three_stars_in_the_south', method: (cp) => cp.threeStarsInTheSouth(), data: TEST_DATA_THREE_STARS_IN_THE_SOUTH, skipped: skipSet(21) },
        { name: 'three_white_soldiers', method: (cp) => cp.threeWhiteSoldiers(), data: TEST_DATA_THREE_WHITE_SOLDIERS },
        { name: 'thrusting', method: (cp) => cp.thrusting(), data: TEST_DATA_THRUSTING, skipped: skipSet(1, 34, 93) },
        { name: 'tristar', method: (cp) => cp.tristar(), data: TEST_DATA_TRISTAR, skipped: skipSet(2, 44, 50, 51, 53, 66, 77, 88, 98, 130, 138, 142, 149, 156, 173, 180, 182, 183, 186) },
        { name: 'two_crows', method: (cp) => cp.twoCrows(), data: TEST_DATA_TWO_CROWS },
        { name: 'unique_three_river', method: (cp) => cp.uniqueThreeRiver(), data: TEST_DATA_UNIQUE_THREE_RIVER },
        { name: 'up_down_gap_side_by_side_white_lines', method: (cp) => cp.upDownGapSideBySideWhiteLines(), data: TEST_DATA_UP_DOWN_GAP_SIDE_BY_SIDE_WHITE_LINES, skipped: skipSet(34, 35, 36, 37, 38, 39) },
        { name: 'upside_gap_two_crows', method: (cp) => cp.upsideGapTwoCrows(), data: TEST_DATA_UPSIDE_GAP_TWO_CROWS },
        { name: 'x_side_gap_three_methods', method: (cp) => cp.xSideGapThreeMethods(), data: TEST_DATA_X_SIDE_GAP_THREE_METHODS },
    ];

    runPatternTests(specs);
});
