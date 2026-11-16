import {} from 'jasmine';

import { Bar } from '../../../entities/bar';
import { BarComponent, DefaultBarComponent } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { DefaultQuoteComponent, QuoteComponent } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { DefaultTradeComponent, TradeComponent } from '../../../entities/trade-component';
import { Adaptivity } from '../../core/adaptivity';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { descriptorOf } from '../../core/descriptors';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Heatmap } from '../../core/outputs/heatmap';
import { Shape } from '../../core/outputs/shape/shape';
import { createIndicator } from '../../factory/factory';
import {
    AutoCorrelationPeriodogram,
    AutoCorrelationPeriodogramParams,
    AutoCorrelationPeriodogramSource,
    PeriodogramOutput,
    autoCorrelationPeriodogramDominantCycleMnemonic,
    autoCorrelationPeriodogramMnemonic,
    defaultParams,
} from './auto-correlation-periodogram';
import { AutoCorrelationPeriodogramOutput } from './output';
import * as testdata from './testdata';

const FIXTURE_TOLERANCE = 1e-11;
const ORACLE_TOLERANCE = 2e-10;
const EXACT_TOLERANCE = 1e-13;

interface ReferenceCase {
    readonly name: string;
    readonly params: Required<
        Pick<
            AutoCorrelationPeriodogramParams,
            'source' | 'minPeriod' | 'maxPeriod' | 'averagingLength' | 'enhanceResolution'
        >
    >;
    readonly expectedHeatmap: number[][];
    readonly expectedMin: number[];
    readonly expectedMax: number[];
    readonly expectedDC: number[];
}

const REFERENCE_CASES: readonly ReferenceCase[] = [
    {
        name: 'V0_L8_H24_A0_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 24, averagingLength: 0, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L8_H24_A0_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H24_A0_E0_MIN,
        expectedMax: testdata.expectedV0_L8_H24_A0_E0_MAX,
        expectedDC: testdata.expectedV0_L8_H24_A0_E0_DC,
    },
    {
        name: 'V0_L8_H24_A0_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 24, averagingLength: 0, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L8_H24_A0_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H24_A0_E1_MIN,
        expectedMax: testdata.expectedV0_L8_H24_A0_E1_MAX,
        expectedDC: testdata.expectedV0_L8_H24_A0_E1_DC,
    },
    {
        name: 'V0_L8_H24_A3_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 24, averagingLength: 3, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L8_H24_A3_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H24_A3_E0_MIN,
        expectedMax: testdata.expectedV0_L8_H24_A3_E0_MAX,
        expectedDC: testdata.expectedV0_L8_H24_A3_E0_DC,
    },
    {
        name: 'V0_L8_H24_A3_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 24, averagingLength: 3, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L8_H24_A3_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H24_A3_E1_MIN,
        expectedMax: testdata.expectedV0_L8_H24_A3_E1_MAX,
        expectedDC: testdata.expectedV0_L8_H24_A3_E1_DC,
    },
    {
        name: 'V0_L8_H24_A5_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 24, averagingLength: 5, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L8_H24_A5_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H24_A5_E0_MIN,
        expectedMax: testdata.expectedV0_L8_H24_A5_E0_MAX,
        expectedDC: testdata.expectedV0_L8_H24_A5_E0_DC,
    },
    {
        name: 'V0_L8_H24_A5_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 24, averagingLength: 5, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L8_H24_A5_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H24_A5_E1_MIN,
        expectedMax: testdata.expectedV0_L8_H24_A5_E1_MAX,
        expectedDC: testdata.expectedV0_L8_H24_A5_E1_DC,
    },
    {
        name: 'V0_L8_H24_A10_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 24, averagingLength: 10, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L8_H24_A10_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H24_A10_E0_MIN,
        expectedMax: testdata.expectedV0_L8_H24_A10_E0_MAX,
        expectedDC: testdata.expectedV0_L8_H24_A10_E0_DC,
    },
    {
        name: 'V0_L8_H24_A10_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 24, averagingLength: 10, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L8_H24_A10_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H24_A10_E1_MIN,
        expectedMax: testdata.expectedV0_L8_H24_A10_E1_MAX,
        expectedDC: testdata.expectedV0_L8_H24_A10_E1_DC,
    },
    {
        name: 'V0_L8_H48_A0_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 48, averagingLength: 0, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L8_H48_A0_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H48_A0_E0_MIN,
        expectedMax: testdata.expectedV0_L8_H48_A0_E0_MAX,
        expectedDC: testdata.expectedV0_L8_H48_A0_E0_DC,
    },
    {
        name: 'V0_L8_H48_A0_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 48, averagingLength: 0, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L8_H48_A0_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H48_A0_E1_MIN,
        expectedMax: testdata.expectedV0_L8_H48_A0_E1_MAX,
        expectedDC: testdata.expectedV0_L8_H48_A0_E1_DC,
    },
    {
        name: 'V0_L8_H48_A3_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 48, averagingLength: 3, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L8_H48_A3_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H48_A3_E0_MIN,
        expectedMax: testdata.expectedV0_L8_H48_A3_E0_MAX,
        expectedDC: testdata.expectedV0_L8_H48_A3_E0_DC,
    },
    {
        name: 'V0_L8_H48_A3_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 48, averagingLength: 3, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L8_H48_A3_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H48_A3_E1_MIN,
        expectedMax: testdata.expectedV0_L8_H48_A3_E1_MAX,
        expectedDC: testdata.expectedV0_L8_H48_A3_E1_DC,
    },
    {
        name: 'V0_L8_H48_A5_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 48, averagingLength: 5, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L8_H48_A5_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H48_A5_E0_MIN,
        expectedMax: testdata.expectedV0_L8_H48_A5_E0_MAX,
        expectedDC: testdata.expectedV0_L8_H48_A5_E0_DC,
    },
    {
        name: 'V0_L8_H48_A5_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 48, averagingLength: 5, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L8_H48_A5_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H48_A5_E1_MIN,
        expectedMax: testdata.expectedV0_L8_H48_A5_E1_MAX,
        expectedDC: testdata.expectedV0_L8_H48_A5_E1_DC,
    },
    {
        name: 'V0_L8_H48_A10_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 48, averagingLength: 10, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L8_H48_A10_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H48_A10_E0_MIN,
        expectedMax: testdata.expectedV0_L8_H48_A10_E0_MAX,
        expectedDC: testdata.expectedV0_L8_H48_A10_E0_DC,
    },
    {
        name: 'V0_L8_H48_A10_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 8, maxPeriod: 48, averagingLength: 10, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L8_H48_A10_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L8_H48_A10_E1_MIN,
        expectedMax: testdata.expectedV0_L8_H48_A10_E1_MAX,
        expectedDC: testdata.expectedV0_L8_H48_A10_E1_DC,
    },
    {
        name: 'V0_L10_H48_A0_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 10, maxPeriod: 48, averagingLength: 0, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L10_H48_A0_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L10_H48_A0_E0_MIN,
        expectedMax: testdata.expectedV0_L10_H48_A0_E0_MAX,
        expectedDC: testdata.expectedV0_L10_H48_A0_E0_DC,
    },
    {
        name: 'V0_L10_H48_A0_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 10, maxPeriod: 48, averagingLength: 0, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L10_H48_A0_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L10_H48_A0_E1_MIN,
        expectedMax: testdata.expectedV0_L10_H48_A0_E1_MAX,
        expectedDC: testdata.expectedV0_L10_H48_A0_E1_DC,
    },
    {
        name: 'V0_L10_H48_A3_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 10, maxPeriod: 48, averagingLength: 3, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L10_H48_A3_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L10_H48_A3_E0_MIN,
        expectedMax: testdata.expectedV0_L10_H48_A3_E0_MAX,
        expectedDC: testdata.expectedV0_L10_H48_A3_E0_DC,
    },
    {
        name: 'V0_L10_H48_A3_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 10, maxPeriod: 48, averagingLength: 3, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L10_H48_A3_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L10_H48_A3_E1_MIN,
        expectedMax: testdata.expectedV0_L10_H48_A3_E1_MAX,
        expectedDC: testdata.expectedV0_L10_H48_A3_E1_DC,
    },
    {
        name: 'V0_L10_H48_A5_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 10, maxPeriod: 48, averagingLength: 5, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L10_H48_A5_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L10_H48_A5_E0_MIN,
        expectedMax: testdata.expectedV0_L10_H48_A5_E0_MAX,
        expectedDC: testdata.expectedV0_L10_H48_A5_E0_DC,
    },
    {
        name: 'V0_L10_H48_A5_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 10, maxPeriod: 48, averagingLength: 5, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L10_H48_A5_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L10_H48_A5_E1_MIN,
        expectedMax: testdata.expectedV0_L10_H48_A5_E1_MAX,
        expectedDC: testdata.expectedV0_L10_H48_A5_E1_DC,
    },
    {
        name: 'V0_L10_H48_A10_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 10, maxPeriod: 48, averagingLength: 10, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L10_H48_A10_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L10_H48_A10_E0_MIN,
        expectedMax: testdata.expectedV0_L10_H48_A10_E0_MAX,
        expectedDC: testdata.expectedV0_L10_H48_A10_E0_DC,
    },
    {
        name: 'V0_L10_H48_A10_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 10, maxPeriod: 48, averagingLength: 10, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L10_H48_A10_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L10_H48_A10_E1_MIN,
        expectedMax: testdata.expectedV0_L10_H48_A10_E1_MAX,
        expectedDC: testdata.expectedV0_L10_H48_A10_E1_DC,
    },
    {
        name: 'V0_L16_H64_A0_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 16, maxPeriod: 64, averagingLength: 0, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L16_H64_A0_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L16_H64_A0_E0_MIN,
        expectedMax: testdata.expectedV0_L16_H64_A0_E0_MAX,
        expectedDC: testdata.expectedV0_L16_H64_A0_E0_DC,
    },
    {
        name: 'V0_L16_H64_A0_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 16, maxPeriod: 64, averagingLength: 0, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L16_H64_A0_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L16_H64_A0_E1_MIN,
        expectedMax: testdata.expectedV0_L16_H64_A0_E1_MAX,
        expectedDC: testdata.expectedV0_L16_H64_A0_E1_DC,
    },
    {
        name: 'V0_L16_H64_A3_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 16, maxPeriod: 64, averagingLength: 3, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L16_H64_A3_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L16_H64_A3_E0_MIN,
        expectedMax: testdata.expectedV0_L16_H64_A3_E0_MAX,
        expectedDC: testdata.expectedV0_L16_H64_A3_E0_DC,
    },
    {
        name: 'V0_L16_H64_A3_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 16, maxPeriod: 64, averagingLength: 3, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L16_H64_A3_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L16_H64_A3_E1_MIN,
        expectedMax: testdata.expectedV0_L16_H64_A3_E1_MAX,
        expectedDC: testdata.expectedV0_L16_H64_A3_E1_DC,
    },
    {
        name: 'V0_L16_H64_A5_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 16, maxPeriod: 64, averagingLength: 5, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L16_H64_A5_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L16_H64_A5_E0_MIN,
        expectedMax: testdata.expectedV0_L16_H64_A5_E0_MAX,
        expectedDC: testdata.expectedV0_L16_H64_A5_E0_DC,
    },
    {
        name: 'V0_L16_H64_A5_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 16, maxPeriod: 64, averagingLength: 5, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L16_H64_A5_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L16_H64_A5_E1_MIN,
        expectedMax: testdata.expectedV0_L16_H64_A5_E1_MAX,
        expectedDC: testdata.expectedV0_L16_H64_A5_E1_DC,
    },
    {
        name: 'V0_L16_H64_A10_E0',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 16, maxPeriod: 64, averagingLength: 10, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV0_L16_H64_A10_E0_HEATMAP,
        expectedMin: testdata.expectedV0_L16_H64_A10_E0_MIN,
        expectedMax: testdata.expectedV0_L16_H64_A10_E0_MAX,
        expectedDC: testdata.expectedV0_L16_H64_A10_E0_DC,
    },
    {
        name: 'V0_L16_H64_A10_E1',
        params: { source: AutoCorrelationPeriodogramSource.Article, minPeriod: 16, maxPeriod: 64, averagingLength: 10, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV0_L16_H64_A10_E1_HEATMAP,
        expectedMin: testdata.expectedV0_L16_H64_A10_E1_MIN,
        expectedMax: testdata.expectedV0_L16_H64_A10_E1_MAX,
        expectedDC: testdata.expectedV0_L16_H64_A10_E1_DC,
    },
    {
        name: 'V1_L8_H24_A0_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 24, averagingLength: 0, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L8_H24_A0_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H24_A0_E0_MIN,
        expectedMax: testdata.expectedV1_L8_H24_A0_E0_MAX,
        expectedDC: testdata.expectedV1_L8_H24_A0_E0_DC,
    },
    {
        name: 'V1_L8_H24_A0_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 24, averagingLength: 0, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L8_H24_A0_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H24_A0_E1_MIN,
        expectedMax: testdata.expectedV1_L8_H24_A0_E1_MAX,
        expectedDC: testdata.expectedV1_L8_H24_A0_E1_DC,
    },
    {
        name: 'V1_L8_H24_A3_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 24, averagingLength: 3, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L8_H24_A3_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H24_A3_E0_MIN,
        expectedMax: testdata.expectedV1_L8_H24_A3_E0_MAX,
        expectedDC: testdata.expectedV1_L8_H24_A3_E0_DC,
    },
    {
        name: 'V1_L8_H24_A3_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 24, averagingLength: 3, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L8_H24_A3_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H24_A3_E1_MIN,
        expectedMax: testdata.expectedV1_L8_H24_A3_E1_MAX,
        expectedDC: testdata.expectedV1_L8_H24_A3_E1_DC,
    },
    {
        name: 'V1_L8_H24_A5_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 24, averagingLength: 5, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L8_H24_A5_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H24_A5_E0_MIN,
        expectedMax: testdata.expectedV1_L8_H24_A5_E0_MAX,
        expectedDC: testdata.expectedV1_L8_H24_A5_E0_DC,
    },
    {
        name: 'V1_L8_H24_A5_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 24, averagingLength: 5, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L8_H24_A5_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H24_A5_E1_MIN,
        expectedMax: testdata.expectedV1_L8_H24_A5_E1_MAX,
        expectedDC: testdata.expectedV1_L8_H24_A5_E1_DC,
    },
    {
        name: 'V1_L8_H24_A10_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 24, averagingLength: 10, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L8_H24_A10_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H24_A10_E0_MIN,
        expectedMax: testdata.expectedV1_L8_H24_A10_E0_MAX,
        expectedDC: testdata.expectedV1_L8_H24_A10_E0_DC,
    },
    {
        name: 'V1_L8_H24_A10_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 24, averagingLength: 10, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L8_H24_A10_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H24_A10_E1_MIN,
        expectedMax: testdata.expectedV1_L8_H24_A10_E1_MAX,
        expectedDC: testdata.expectedV1_L8_H24_A10_E1_DC,
    },
    {
        name: 'V1_L8_H48_A0_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 48, averagingLength: 0, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L8_H48_A0_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H48_A0_E0_MIN,
        expectedMax: testdata.expectedV1_L8_H48_A0_E0_MAX,
        expectedDC: testdata.expectedV1_L8_H48_A0_E0_DC,
    },
    {
        name: 'V1_L8_H48_A0_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 48, averagingLength: 0, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L8_H48_A0_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H48_A0_E1_MIN,
        expectedMax: testdata.expectedV1_L8_H48_A0_E1_MAX,
        expectedDC: testdata.expectedV1_L8_H48_A0_E1_DC,
    },
    {
        name: 'V1_L8_H48_A3_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 48, averagingLength: 3, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L8_H48_A3_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H48_A3_E0_MIN,
        expectedMax: testdata.expectedV1_L8_H48_A3_E0_MAX,
        expectedDC: testdata.expectedV1_L8_H48_A3_E0_DC,
    },
    {
        name: 'V1_L8_H48_A3_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 48, averagingLength: 3, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L8_H48_A3_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H48_A3_E1_MIN,
        expectedMax: testdata.expectedV1_L8_H48_A3_E1_MAX,
        expectedDC: testdata.expectedV1_L8_H48_A3_E1_DC,
    },
    {
        name: 'V1_L8_H48_A5_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 48, averagingLength: 5, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L8_H48_A5_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H48_A5_E0_MIN,
        expectedMax: testdata.expectedV1_L8_H48_A5_E0_MAX,
        expectedDC: testdata.expectedV1_L8_H48_A5_E0_DC,
    },
    {
        name: 'V1_L8_H48_A5_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 48, averagingLength: 5, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L8_H48_A5_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H48_A5_E1_MIN,
        expectedMax: testdata.expectedV1_L8_H48_A5_E1_MAX,
        expectedDC: testdata.expectedV1_L8_H48_A5_E1_DC,
    },
    {
        name: 'V1_L8_H48_A10_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 48, averagingLength: 10, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L8_H48_A10_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H48_A10_E0_MIN,
        expectedMax: testdata.expectedV1_L8_H48_A10_E0_MAX,
        expectedDC: testdata.expectedV1_L8_H48_A10_E0_DC,
    },
    {
        name: 'V1_L8_H48_A10_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 8, maxPeriod: 48, averagingLength: 10, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L8_H48_A10_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L8_H48_A10_E1_MIN,
        expectedMax: testdata.expectedV1_L8_H48_A10_E1_MAX,
        expectedDC: testdata.expectedV1_L8_H48_A10_E1_DC,
    },
    {
        name: 'V1_L10_H48_A0_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 10, maxPeriod: 48, averagingLength: 0, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L10_H48_A0_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L10_H48_A0_E0_MIN,
        expectedMax: testdata.expectedV1_L10_H48_A0_E0_MAX,
        expectedDC: testdata.expectedV1_L10_H48_A0_E0_DC,
    },
    {
        name: 'V1_L10_H48_A0_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 10, maxPeriod: 48, averagingLength: 0, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L10_H48_A0_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L10_H48_A0_E1_MIN,
        expectedMax: testdata.expectedV1_L10_H48_A0_E1_MAX,
        expectedDC: testdata.expectedV1_L10_H48_A0_E1_DC,
    },
    {
        name: 'V1_L10_H48_A3_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 10, maxPeriod: 48, averagingLength: 3, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L10_H48_A3_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L10_H48_A3_E0_MIN,
        expectedMax: testdata.expectedV1_L10_H48_A3_E0_MAX,
        expectedDC: testdata.expectedV1_L10_H48_A3_E0_DC,
    },
    {
        name: 'V1_L10_H48_A3_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 10, maxPeriod: 48, averagingLength: 3, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L10_H48_A3_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L10_H48_A3_E1_MIN,
        expectedMax: testdata.expectedV1_L10_H48_A3_E1_MAX,
        expectedDC: testdata.expectedV1_L10_H48_A3_E1_DC,
    },
    {
        name: 'V1_L10_H48_A5_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 10, maxPeriod: 48, averagingLength: 5, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L10_H48_A5_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L10_H48_A5_E0_MIN,
        expectedMax: testdata.expectedV1_L10_H48_A5_E0_MAX,
        expectedDC: testdata.expectedV1_L10_H48_A5_E0_DC,
    },
    {
        name: 'V1_L10_H48_A5_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 10, maxPeriod: 48, averagingLength: 5, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L10_H48_A5_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L10_H48_A5_E1_MIN,
        expectedMax: testdata.expectedV1_L10_H48_A5_E1_MAX,
        expectedDC: testdata.expectedV1_L10_H48_A5_E1_DC,
    },
    {
        name: 'V1_L10_H48_A10_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 10, maxPeriod: 48, averagingLength: 10, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L10_H48_A10_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L10_H48_A10_E0_MIN,
        expectedMax: testdata.expectedV1_L10_H48_A10_E0_MAX,
        expectedDC: testdata.expectedV1_L10_H48_A10_E0_DC,
    },
    {
        name: 'V1_L10_H48_A10_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 10, maxPeriod: 48, averagingLength: 10, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L10_H48_A10_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L10_H48_A10_E1_MIN,
        expectedMax: testdata.expectedV1_L10_H48_A10_E1_MAX,
        expectedDC: testdata.expectedV1_L10_H48_A10_E1_DC,
    },
    {
        name: 'V1_L16_H64_A0_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 16, maxPeriod: 64, averagingLength: 0, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L16_H64_A0_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L16_H64_A0_E0_MIN,
        expectedMax: testdata.expectedV1_L16_H64_A0_E0_MAX,
        expectedDC: testdata.expectedV1_L16_H64_A0_E0_DC,
    },
    {
        name: 'V1_L16_H64_A0_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 16, maxPeriod: 64, averagingLength: 0, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L16_H64_A0_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L16_H64_A0_E1_MIN,
        expectedMax: testdata.expectedV1_L16_H64_A0_E1_MAX,
        expectedDC: testdata.expectedV1_L16_H64_A0_E1_DC,
    },
    {
        name: 'V1_L16_H64_A3_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 16, maxPeriod: 64, averagingLength: 3, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L16_H64_A3_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L16_H64_A3_E0_MIN,
        expectedMax: testdata.expectedV1_L16_H64_A3_E0_MAX,
        expectedDC: testdata.expectedV1_L16_H64_A3_E0_DC,
    },
    {
        name: 'V1_L16_H64_A3_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 16, maxPeriod: 64, averagingLength: 3, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L16_H64_A3_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L16_H64_A3_E1_MIN,
        expectedMax: testdata.expectedV1_L16_H64_A3_E1_MAX,
        expectedDC: testdata.expectedV1_L16_H64_A3_E1_DC,
    },
    {
        name: 'V1_L16_H64_A5_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 16, maxPeriod: 64, averagingLength: 5, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L16_H64_A5_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L16_H64_A5_E0_MIN,
        expectedMax: testdata.expectedV1_L16_H64_A5_E0_MAX,
        expectedDC: testdata.expectedV1_L16_H64_A5_E0_DC,
    },
    {
        name: 'V1_L16_H64_A5_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 16, maxPeriod: 64, averagingLength: 5, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L16_H64_A5_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L16_H64_A5_E1_MIN,
        expectedMax: testdata.expectedV1_L16_H64_A5_E1_MAX,
        expectedDC: testdata.expectedV1_L16_H64_A5_E1_DC,
    },
    {
        name: 'V1_L16_H64_A10_E0',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 16, maxPeriod: 64, averagingLength: 10, enhanceResolution: false },
        expectedHeatmap: testdata.expectedV1_L16_H64_A10_E0_HEATMAP,
        expectedMin: testdata.expectedV1_L16_H64_A10_E0_MIN,
        expectedMax: testdata.expectedV1_L16_H64_A10_E0_MAX,
        expectedDC: testdata.expectedV1_L16_H64_A10_E0_DC,
    },
    {
        name: 'V1_L16_H64_A10_E1',
        params: { source: AutoCorrelationPeriodogramSource.Book, minPeriod: 16, maxPeriod: 64, averagingLength: 10, enhanceResolution: true },
        expectedHeatmap: testdata.expectedV1_L16_H64_A10_E1_HEATMAP,
        expectedMin: testdata.expectedV1_L16_H64_A10_E1_MIN,
        expectedMax: testdata.expectedV1_L16_H64_A10_E1_MAX,
        expectedDC: testdata.expectedV1_L16_H64_A10_E1_DC,
    },
];

function assertClose(
    actual: number,
    expected: number,
    label: string,
    tolerance: number = EXACT_TOLERANCE,
): void {
    if (Number.isNaN(expected)) {
        if (!Number.isNaN(actual)) {
            fail(`${label}: expected NaN, got ${actual}`);
        }
        return;
    }
    if (!Number.isFinite(actual)) {
        fail(`${label}: expected ${expected}, got ${actual}`);
        return;
    }
    const bound = Math.max(tolerance, Math.abs(expected) * tolerance);
    if (Math.abs(actual - expected) > bound) {
        fail(
            `${label}: expected ${expected}, got ${actual} (diff ${Math.abs(actual - expected)} > ${bound})`,
        );
    }
}

interface OracleStep {
    readonly highPass: number;
    readonly filtered: number;
    readonly correlations: number[];
    readonly powers: number[];
    readonly maximum: number;
    readonly dominant: number;
    readonly display: number[];
    readonly hasDominant: boolean;
}

function radians(degrees: number): number {
    return (degrees * Math.PI) / 180.0;
}

function oldAt(series: readonly number[], index: number): number {
    return index >= 0 ? series[index] : 0.0;
}

function sourceReference(
    prices: readonly number[],
    configuration: ReferenceCase['params'],
): OracleStep[] {
    const book = configuration.source === AutoCorrelationPeriodogramSource.Book;
    const hpHistory: number[] = [];
    const filteredHistory: number[] = [];
    const correlations = new Array<number>(configuration.maxPeriod + 1).fill(0.0);
    const binTotal = configuration.maxPeriod - configuration.minPeriod + 1;
    const powers = new Array<number>(binTotal).fill(0.0);
    let maximum = 0.0;
    let dominant = 0.0;
    let hasDominant = false;
    const steps: OracleStep[] = [];

    const ssAngle = radians((Math.SQRT2 * 180.0) / configuration.minPeriod);
    const decay = Math.exp((-Math.SQRT2 * Math.PI) / configuration.minPeriod);
    const c2 = 2.0 * decay * Math.cos(ssAngle);
    const c3 = -decay * decay;
    const c1 = 1.0 - c2 - c3;

    const articleAngle = radians(360.0 / configuration.maxPeriod);
    const articlePole = (1.0 - Math.sin(articleAngle)) / Math.cos(articleAngle);
    const bookAngle = radians(((1.0 / Math.SQRT2) * 360.0) / configuration.maxPeriod);
    const alpha = (Math.cos(bookAngle) + Math.sin(bookAngle) - 1.0) / Math.cos(bookAngle);

    for (let stepIndex = 0; stepIndex < prices.length; stepIndex++) {
        const price = prices[stepIndex];
        const highValue = book
            ? (1.0 - alpha / 2.0) *
                  (1.0 - alpha / 2.0) *
                  (price - 2.0 * oldAt(prices, stepIndex - 1) + oldAt(prices, stepIndex - 2)) +
              2.0 * (1.0 - alpha) * oldAt(hpHistory, stepIndex - 1) -
              (1.0 - alpha) * (1.0 - alpha) * oldAt(hpHistory, stepIndex - 2)
            : 0.5 * (1.0 + articlePole) * (price - oldAt(prices, stepIndex - 1)) +
              articlePole * oldAt(hpHistory, stepIndex - 1);
        hpHistory.push(highValue);

        const filteredValue =
            (c1 * (highValue + oldAt(hpHistory, stepIndex - 1))) / 2.0 +
            c2 * oldAt(filteredHistory, stepIndex - 1) +
            c3 * oldAt(filteredHistory, stepIndex - 2);
        filteredHistory.push(filteredValue);

        for (let lag = 0; lag <= configuration.maxPeriod; lag++) {
            const windowSize =
                configuration.averagingLength !== 0 ? configuration.averagingLength : lag;
            let sx = 0.0;
            let sy = 0.0;
            let sxx = 0.0;
            let syy = 0.0;
            let sxy = 0.0;
            for (let offset = 0; offset < windowSize; offset++) {
                const xSample = oldAt(filteredHistory, stepIndex - offset);
                const ySample = oldAt(filteredHistory, stepIndex - lag - offset);
                sx += xSample;
                sy += ySample;
                sxx += xSample * xSample;
                sxy += xSample * ySample;
                syy += ySample * ySample;
            }
            const denominator =
                (windowSize * sxx - sx * sx) * (windowSize * syy - sy * sy);
            if (denominator > 0.0) {
                correlations[lag] =
                    (windowSize * sxy - sx * sy) / Math.sqrt(denominator);
            }
        }

        for (let binIndex = 0; binIndex < binTotal; binIndex++) {
            const period = configuration.minPeriod + binIndex;
            let cosinePart = 0.0;
            let sinePart = 0.0;
            for (let lag = 3; lag <= configuration.maxPeriod; lag++) {
                const angle = radians((360.0 * lag) / period);
                cosinePart += correlations[lag] * Math.cos(angle);
                sinePart += correlations[lag] * Math.sin(angle);
            }
            const squaredSum = cosinePart * cosinePart + sinePart * sinePart;
            powers[binIndex] = 0.2 * squaredSum * squaredSum + 0.8 * powers[binIndex];
        }

        maximum = book ? 0.995 * maximum : 0.0;
        for (const powerValue of powers) {
            if (powerValue > maximum) {
                maximum = powerValue;
            }
        }
        const normalized = powers.map((powerValue) =>
            maximum > 0.0 ? powerValue / maximum : 0.0,
        );
        const weights =
            !book && configuration.enhanceResolution
                ? normalized.map((normValue) => normValue * normValue * normValue)
                : normalized.slice();
        const peak = Math.max(...weights);
        let spx = 0.0;
        let sp = 0.0;
        for (let binIndex = 0; binIndex < binTotal; binIndex++) {
            const weight = weights[binIndex];
            if (
                (book && weight >= 0.5) ||
                (!book && peak >= 0.25 && weight >= 0.25)
            ) {
                spx += (configuration.minPeriod + binIndex) * weight;
                sp += weight;
            }
        }
        let candidate = sp !== 0.0 ? spx / sp : 0.0;
        if ((book && sp === 0.0) || (!book && sp < 0.25)) {
            candidate = dominant;
        } else if (sp !== 0.0) {
            hasDominant = true;
        }
        dominant = candidate;
        const display =
            book && configuration.enhanceResolution
                ? normalized.map((normValue) => normValue * normValue)
                : weights.slice();
        steps.push({
            highPass: highValue,
            filtered: filteredValue,
            correlations: correlations.slice(),
            powers: powers.slice(),
            maximum,
            dominant,
            display,
            hasDominant,
        });
    }

    return steps;
}

describe('AutoCorrelationPeriodogram', () => {
    it('matches all 64 reference fixtures', () => {
        expect(REFERENCE_CASES.length).toBe(64);
        const baseTime = new Date(Date.UTC(2020, 0, 1));
        for (const tc of REFERENCE_CASES) {
            const indicator = new AutoCorrelationPeriodogram(tc.params);
            const bins = tc.params.maxPeriod - tc.params.minPeriod + 1;
            for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
                const sampleTime = new Date(baseTime.getTime() + barIndex * 60000);
                const result = indicator.update(testdata.testInput[barIndex], sampleTime);
                expect(result.heatmap.time.getTime()).toBe(sampleTime.getTime());
                expect(result.heatmap.parameterFirst).toBe(tc.params.minPeriod);
                expect(result.heatmap.parameterLast).toBe(tc.params.maxPeriod);
                expect(result.heatmap.parameterResolution).toBe(1.0);

                const primed = barIndex >= indicator.warmupBars;
                expect(indicator.isPrimed()).toBe(primed);
                assertClose(
                    result.dominantCycle,
                    tc.expectedDC[barIndex],
                    `${tc.name}.dc[${barIndex}]`,
                    FIXTURE_TOLERANCE,
                );
                assertClose(
                    result.heatmap.valueMin,
                    tc.expectedMin[barIndex],
                    `${tc.name}.min[${barIndex}]`,
                    FIXTURE_TOLERANCE,
                );
                assertClose(
                    result.heatmap.valueMax,
                    tc.expectedMax[barIndex],
                    `${tc.name}.max[${barIndex}]`,
                    FIXTURE_TOLERANCE,
                );

                if (!primed) {
                    expect(result.heatmap.isEmpty()).toBeTrue();
                    continue;
                }

                expect(result.heatmap.values.length).toBe(bins);
                for (let binIndex = 0; binIndex < bins; binIndex++) {
                    assertClose(
                        result.heatmap.values[binIndex],
                        tc.expectedHeatmap[barIndex][binIndex],
                        `${tc.name}.hm[${barIndex}][${binIndex}]`,
                        FIXTURE_TOLERANCE,
                    );
                }
            }
        }
    });

    it('matches independent degree-based source transcription across all 64 cases', () => {
        for (const tc of REFERENCE_CASES) {
            const reference = sourceReference(testdata.testInput, tc.params);
            const indicator = new AutoCorrelationPeriodogram(tc.params);
            for (let barIndex = 0; barIndex < testdata.testInput.length; barIndex++) {
                const result = indicator.update(testdata.testInput[barIndex]);
                const expected = reference[barIndex];
                const state = indicator.state;
                assertClose(state.highPass, expected.highPass, `${tc.name}.hp`, ORACLE_TOLERANCE);
                assertClose(
                    state.filtered,
                    expected.filtered,
                    `${tc.name}.filtered`,
                    ORACLE_TOLERANCE,
                );
                assertClose(
                    state.normalizationMaximum,
                    expected.maximum,
                    `${tc.name}.max`,
                    ORACLE_TOLERANCE,
                );
                assertClose(
                    state.internalDominantCycle,
                    expected.dominant,
                    `${tc.name}.dom`,
                    ORACLE_TOLERANCE,
                );
                for (let lag = 0; lag < expected.correlations.length; lag++) {
                    assertClose(
                        state.correlations[lag],
                        expected.correlations[lag],
                        `${tc.name}.corr`,
                        ORACLE_TOLERANCE,
                    );
                }
                for (let binIndex = 0; binIndex < expected.powers.length; binIndex++) {
                    assertClose(
                        state.smoothedPower[binIndex],
                        expected.powers[binIndex],
                        `${tc.name}.pow`,
                        ORACLE_TOLERANCE,
                    );
                }

                const primed = barIndex >= indicator.warmupBars;
                if (primed) {
                    for (let binIndex = 0; binIndex < expected.display.length; binIndex++) {
                        assertClose(
                            result.heatmap.values[binIndex],
                            expected.display[binIndex],
                            `${tc.name}.disp`,
                            ORACLE_TOLERANCE,
                        );
                    }
                } else {
                    expect(result.heatmap.isEmpty()).toBeTrue();
                }
                const expectedDC = primed && expected.hasDominant ? expected.dominant : Number.NaN;
                assertClose(result.dominantCycle, expectedDC, `${tc.name}.dc`, ORACLE_TOLERANCE);
            }
        }
    });

    it('preserves enhancement placement and book centroid hold', () => {
        const runSeries = (
            source: AutoCorrelationPeriodogramSource,
            enhanceResolution: boolean,
        ): PeriodogramOutput[] => {
            const indicator = new AutoCorrelationPeriodogram({ source, enhanceResolution });
            return testdata.testInput.map((sample) => indicator.update(sample));
        };

        const articlePlain = runSeries(AutoCorrelationPeriodogramSource.Article, false);
        const articleEnhanced = runSeries(AutoCorrelationPeriodogramSource.Article, true);
        const bookPlain = runSeries(AutoCorrelationPeriodogramSource.Book, false);
        const bookEnhanced = runSeries(AutoCorrelationPeriodogramSource.Book, true);

        for (let barIndex = 50; barIndex < 252; barIndex++) {
            assertClose(
                bookPlain[barIndex].dominantCycle,
                bookEnhanced[barIndex].dominantCycle,
                'book.dc',
            );
            expect(bookEnhanced[barIndex].heatmap.values).toEqual(
                bookPlain[barIndex].heatmap.values.map((cell) => cell * cell),
            );
            expect(articleEnhanced[barIndex].heatmap.values).toEqual(
                articlePlain[barIndex].heatmap.values.map((cell) => cell * cell * cell),
            );
            assertClose(articlePlain[barIndex].heatmap.valueMax, 1.0, 'article.max');
        }
        expect(articlePlain[251].dominantCycle).not.toBe(articleEnhanced[251].dominantCycle);

        let holds = 0;
        for (let barIndex = 51; barIndex < 252; barIndex++) {
            if (bookPlain[barIndex].heatmap.valueMax < 0.5) {
                holds++;
                assertClose(
                    bookPlain[barIndex].dominantCycle,
                    bookPlain[barIndex - 1].dominantCycle,
                    'book.hold',
                );
            }
        }
        expect(holds).toBeGreaterThan(0);
    });

    it('handles zero power, sine peak at 20, skipped NaN, overflow atomicity, and reset', () => {
        for (const source of [
            AutoCorrelationPeriodogramSource.Article,
            AutoCorrelationPeriodogramSource.Book,
        ]) {
            const zeroInd = new AutoCorrelationPeriodogram({ source });
            let zeroResult!: PeriodogramOutput;
            for (let stepIndex = 0; stepIndex < 100; stepIndex++) {
                zeroResult = zeroInd.update(0.0);
            }
            expect(zeroInd.isPrimed()).toBeTrue();
            expect(Number.isNaN(zeroResult.dominantCycle)).toBeTrue();
            expect(zeroResult.heatmap.values.every((cell) => cell === 0.0)).toBeTrue();
        }

        const sineInd = AutoCorrelationPeriodogram.default();
        let sineOut!: PeriodogramOutput;
        for (let stepIndex = 0; stepIndex < 1000; stepIndex++) {
            sineOut = sineInd.update(100.0 + Math.sin((2.0 * Math.PI * stepIndex) / 20.0));
        }
        const maxCell = Math.max(...sineOut.heatmap.values);
        const peakPeriod =
            sineOut.heatmap.parameterFirst + sineOut.heatmap.values.indexOf(maxCell);
        expect(peakPeriod).toBe(20.0);
        expect(sineOut.dominantCycle).toBeGreaterThan(18.0);
        expect(sineOut.dominantCycle).toBeLessThan(22.0);

        const indicator = AutoCorrelationPeriodogram.default();
        const firstRun = testdata.testInput.map((sample) => indicator.update(sample));
        const savedState = indicator.state;
        const savedOutput = indicator.output;
        const savedCount = indicator.samplesSeen;

        const skipped = indicator.update(Number.NaN);
        expect(skipped.heatmap.isEmpty()).toBeTrue();
        expect(Number.isNaN(skipped.dominantCycle)).toBeTrue();
        expect(indicator.output).toBe(savedOutput);
        expect(indicator.samplesSeen).toBe(savedCount);

        for (const badSample of [
            Number.POSITIVE_INFINITY,
            Number.NEGATIVE_INFINITY,
            1e308,
            true as unknown as number,
            'bad' as unknown as number,
        ]) {
            expect(() => indicator.update(badSample)).toThrow();
            expect(indicator.samplesSeen).toBe(savedCount);
            assertClose(indicator.state.highPass, savedState.highPass, 'overflow.hp', 0);
            assertClose(indicator.state.filtered, savedState.filtered, 'overflow.filt', 0);
        }

        indicator.reset();
        const secondRun = testdata.testInput.map((sample) => indicator.update(sample));
        for (let barIndex = 0; barIndex < firstRun.length; barIndex++) {
            expect(secondRun[barIndex].heatmap.values).toEqual(
                firstRun[barIndex].heatmap.values,
            );
            assertClose(
                secondRun[barIndex].dominantCycle,
                firstRun[barIndex].dominantCycle,
                'replay.dc',
                0,
            );
        }

        const snapshot = secondRun[251].heatmap.values.slice();
        secondRun[251].heatmap.values[0] = -999.0;
        expect(indicator.output.heatmap.values).toEqual(snapshot);
    });

    it('validates metadata, mnemonics, entity updates, factory, and invalid params', () => {
        const descriptor = descriptorOf(IndicatorIdentifier.AutoCorrelationPeriodogram);
        expect(descriptor).toBeDefined();
        expect(descriptor!.adaptivity).toBe(Adaptivity.Adaptive);
        expect(descriptor!.outputs.length).toBe(2);

        const defaultInd = AutoCorrelationPeriodogram.default();
        const meta = defaultInd.metadata();
        expect(meta.identifier).toBe(IndicatorIdentifier.AutoCorrelationPeriodogram);
        expect(meta.mnemonic).toBe('acp(8, 48, 3)');
        expect(meta.outputs.length).toBe(2);
        expect(meta.outputs[0].kind).toBe(AutoCorrelationPeriodogramOutput.Value);
        expect(meta.outputs[0].shape).toBe(Shape.Heatmap);
        expect(meta.outputs[0].mnemonic).toBe('acp(8, 48, 3)');
        expect(meta.outputs[1].kind).toBe(AutoCorrelationPeriodogramOutput.DominantCycle);
        expect(meta.outputs[1].shape).toBe(Shape.Scalar);
        expect(meta.outputs[1].mnemonic).toBe('acp-dc(8, 48, 3)');

        const bookEnh = AutoCorrelationPeriodogram.fromParams({
            source: AutoCorrelationPeriodogramSource.Book,
            enhanceResolution: true,
        });
        expect(bookEnh.minPeriod).toBe(10);
        expect(bookEnh.metadata().mnemonic).toBe('acp(10, 48, 3, book, enh)');
        expect(bookEnh.metadata().outputs[1].mnemonic).toBe('acp-dc(10, 48, 3, book)');

        expect(autoCorrelationPeriodogramMnemonic({ enhanceResolution: true })).toBe(
            'acp(8, 48, 3, enh)',
        );
        expect(
            autoCorrelationPeriodogramDominantCycleMnemonic({ enhanceResolution: true }),
        ).toBe('acp-dc(8, 48, 3, enh)');

        const combos: Array<
            [BarComponent | undefined, QuoteComponent | undefined, TradeComponent | undefined]
        > = [
            [undefined, undefined, undefined],
            [BarComponent.Open, undefined, undefined],
            [undefined, QuoteComponent.Bid, undefined],
            [undefined, undefined, TradeComponent.Volume],
            [BarComponent.Median, QuoteComponent.Ask, undefined],
            [BarComponent.Median, undefined, TradeComponent.Volume],
            [undefined, QuoteComponent.Ask, TradeComponent.Volume],
            [BarComponent.Open, QuoteComponent.Bid, TradeComponent.Volume],
        ];
        for (const [barComp, quoteComp, tradeComp] of combos) {
            const params = {
                ...defaultParams(),
                barComponent: barComp,
                quoteComponent: quoteComp,
                tradeComponent: tradeComp,
            };
            const instance = new AutoCorrelationPeriodogram(params);
            const suffix = componentTripleMnemonic(
                barComp ?? DefaultBarComponent,
                quoteComp ?? DefaultQuoteComponent,
                tradeComp ?? DefaultTradeComponent,
            );
            expect(instance.metadata().mnemonic).toBe(`acp(8, 48, 3${suffix})`);
            expect(instance.metadata().outputs[1].mnemonic).toBe(
                `acp-dc(8, 48, 3${suffix})`,
            );
        }

        const baseTime = new Date(Date.UTC(2020, 0, 1));
        const barInd = AutoCorrelationPeriodogram.default();
        const scalarInd = AutoCorrelationPeriodogram.default();
        const quoteInd = AutoCorrelationPeriodogram.default();
        const tradeInd = AutoCorrelationPeriodogram.default();
        for (let stepIndex = 0; stepIndex < 60; stepIndex++) {
            const sampleTime = new Date(baseTime.getTime() + stepIndex * 86400000);
            const price = testdata.testInput[stepIndex];
            const bar = new Bar();
            bar.time = sampleTime;
            bar.open = testdata.testOpen[stepIndex];
            bar.high = testdata.testHigh[stepIndex];
            bar.low = testdata.testLow[stepIndex];
            bar.close = price;
            bar.volume = 1000;
            const scalar = new Scalar();
            scalar.time = sampleTime;
            scalar.value = price;
            const quote = new Quote();
            quote.time = sampleTime;
            quote.bidPrice = price - 0.5;
            quote.askPrice = price + 0.5;
            quote.bidSize = 10;
            quote.askSize = 10;
            const trade = new Trade();
            trade.time = sampleTime;
            trade.price = price;
            trade.volume = 10;

            const outBar = barInd.updateBar(bar);
            const outScalar = scalarInd.updateScalar(scalar);
            const outQuote = quoteInd.updateQuote(quote);
            const outTrade = tradeInd.updateTrade(trade);
            expect(outBar.length).toBe(2);
            expect((outBar[0] as Heatmap).values).toEqual((outScalar[0] as Heatmap).values);
            assertClose(
                (outBar[1] as Scalar).value,
                (outScalar[1] as Scalar).value,
                `bar_vs_scalar[${stepIndex}]`,
            );
            assertClose(
                (outBar[1] as Scalar).value,
                (outQuote[1] as Scalar).value,
                `bar_vs_quote[${stepIndex}]`,
            );
            assertClose(
                (outBar[1] as Scalar).value,
                (outTrade[1] as Scalar).value,
                `bar_vs_trade[${stepIndex}]`,
            );
        }

        const factoryDefault = createIndicator(IndicatorIdentifier.AutoCorrelationPeriodogram);
        expect(factoryDefault.metadata().mnemonic).toBe('acp(8, 48, 3)');
        const factoryBook = createIndicator(IndicatorIdentifier.AutoCorrelationPeriodogram, {
            source: 'book',
            enhanceResolution: true,
            barComponent: BarComponent.Median,
        });
        expect(factoryBook.metadata().mnemonic).toBe('acp(10, 48, 3, book, enh, hl/2)');
        expect(factoryBook.metadata().outputs[1].mnemonic).toBe(
            'acp-dc(10, 48, 3, book, hl/2)',
        );

        for (const badParams of [
            { source: 'literal370' },
            { source: 1 },
            { minPeriod: 0 },
            { minPeriod: 2 },
            { minPeriod: 8.5 },
            { minPeriod: true },
            { minPeriod: null },
            { maxPeriod: 4 },
            { maxPeriod: 8 },
            { source: 'book', maxPeriod: 10 },
            { averagingLength: 1 },
            { averagingLength: -1 },
            { averagingLength: 2.5 },
            { enhanceResolution: 1 },
            { disableAutomaticGainControl: true },
            { automaticGainControlDecayFactor: 0.995 },
            { fixedNormalization: true },
            { unknownKey: 1 },
        ]) {
            expect(() => new AutoCorrelationPeriodogram(badParams as never)).toThrow();
            expect(() =>
                createIndicator(IndicatorIdentifier.AutoCorrelationPeriodogram, badParams as never),
            ).toThrow();
        }
    });
});
