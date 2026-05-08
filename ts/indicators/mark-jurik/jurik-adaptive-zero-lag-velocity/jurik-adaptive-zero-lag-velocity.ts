import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { JurikAdaptiveZeroLagVelocityParams } from './params';

/** Function to calculate mnemonic of a JurikAdaptiveZeroLagVelocity indicator. */
export const jurikAdaptiveZeroLagVelocityMnemonic = (params: JurikAdaptiveZeroLagVelocityParams): string =>
    'javel('.concat(
        params.loLength.toString(), ', ',
        params.hiLength.toString(), ', ',
        params.sensitivity.toString(), ', ',
        params.period.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
        ')');

/** Adaptive smoother (Stage 2) for JAVEL. */
class VelSmooth {
    private readonly jrc03: number;
    private readonly jrc06: number;
    private readonly jrc07: number;
    private readonly emaFactor: number;
    private readonly damping: number;
    private readonly eps2 = 0.0001;
    private readonly bufferSize = 1001;
    private readonly buffer: number[];
    private head = 0;
    private length = 0;
    private barCount = 0;
    private velocity = 0;
    private position = 0;
    private smoothedMAD = 0;
    private initialized = false;

    constructor(period: number) {
        this.jrc03 = Math.min(500.0, Math.max(this.eps2, period));
        this.jrc06 = Math.max(31, Math.ceil(2 * period));
        this.jrc07 = Math.min(30, Math.ceil(period));
        this.emaFactor = 1.0 - Math.exp(-Math.log(4.0) / (period / 2.0));
        this.damping = 0.86 - 0.55 / Math.sqrt(this.jrc03);
        this.buffer = new Array(this.bufferSize).fill(0);
    }

    public update(value: number): number {
        this.barCount++;

        // Store in circular buffer.
        const oldIndex = this.head % this.bufferSize;
        this.buffer[oldIndex] = value;
        this.head++;

        if (this.length < this.jrc06) {
            this.length++;
        }

        const length = this.length;

        // First bar: initialize position.
        if (length < 2) {
            if (!this.initialized) {
                this.position = value;
                this.initialized = true;
            }
            return this.position;
        }

        if (!this.initialized) {
            this.position = value;
            this.initialized = true;
        }

        // Linear regression over buffer.
        let sumValues = 0;
        let sumWeighted = 0;

        for (let k = 0; k < length; k++) {
            let idx = (this.head - length + k) % this.bufferSize;
            if (idx < 0) { idx += this.bufferSize; }
            sumValues += this.buffer[idx];
            sumWeighted += this.buffer[idx] * k;
        }

        const midpoint = (length - 1) / 2.0;
        const sumXSq = length * (length - 1) * (2 * length - 1) / 6.0;
        const regressionDenom = sumXSq - length * midpoint * midpoint;

        let regressionSlope = 0;
        if (Math.abs(regressionDenom) >= this.eps2) {
            regressionSlope = (sumWeighted - midpoint * sumValues) / regressionDenom;
        }

        const intercept = sumValues / length - regressionSlope * midpoint;

        // Compute MAD from regression residuals.
        let sumAbsDev = 0;
        for (let k = 0; k < length; k++) {
            let idx = (this.head - length + k) % this.bufferSize;
            if (idx < 0) { idx += this.bufferSize; }
            const predicted = intercept + regressionSlope * k;
            sumAbsDev += Math.abs(this.buffer[idx] - predicted);
        }

        let rawMAD = sumAbsDev / length;
        const scale = 1.2 * Math.pow(this.jrc06 / length, 0.25);
        rawMAD *= scale;

        // Smooth MAD with EMA.
        if (this.barCount <= this.jrc07 + 1) {
            this.smoothedMAD = rawMAD;
        } else {
            this.smoothedMAD += this.emaFactor * (rawMAD - this.smoothedMAD);
        }

        // Adaptive velocity/position dynamics.
        const predictionError = value - this.position;

        let responseFactor: number;
        if (this.smoothedMAD * this.jrc03 < this.eps2) {
            responseFactor = 1.0;
        } else {
            responseFactor = 1.0 - Math.exp(-Math.abs(predictionError) / (this.smoothedMAD * this.jrc03));
        }

        this.velocity = responseFactor * predictionError + this.velocity * this.damping;
        this.position += this.velocity;

        return this.position;
    }
}

/** Jurik Adaptive Zero Lag Velocity (JAVEL) line indicator. */
export class JurikAdaptiveZeroLagVelocity extends LineIndicator {
    private readonly loLength: number;
    private readonly hiLength: number;
    private readonly sensitivity: number;
    private readonly eps = 0.001;

    private prices: number[] = [];
    private value1Arr: number[] = [];
    private barCount = 0;
    private readonly smooth: VelSmooth;

    public constructor(params: JurikAdaptiveZeroLagVelocityParams) {
        super();

        if (params.loLength < 2) {
            throw new Error('invalid jurik adaptive zero lag velocity parameters: lo_length should be at least 2');
        }
        if (params.hiLength < params.loLength) {
            throw new Error('invalid jurik adaptive zero lag velocity parameters: hi_length should be at least lo_length');
        }
        if (params.period <= 0) {
            throw new Error('invalid jurik adaptive zero lag velocity parameters: period should be positive');
        }

        this.loLength = params.loLength;
        this.hiLength = params.hiLength;
        this.sensitivity = params.sensitivity;
        this.smooth = new VelSmooth(params.period);
        this.mnemonic = jurikAdaptiveZeroLagVelocityMnemonic(params);
        this.description = 'Jurik adaptive zero lag velocity ' + this.mnemonic;
        this.primed = false;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.JurikAdaptiveZeroLagVelocity,
            this.mnemonic,
            this.description,
            [{ mnemonic: this.mnemonic, description: this.description }],
        );
    }

    private computeAdaptiveDepth(bar: number): number {
        let longWindow = bar;
        if (longWindow > 99) { longWindow = 99; }
        longWindow++;

        let shortWindow = bar;
        if (shortWindow > 9) { shortWindow = 9; }
        shortWindow++;

        let avg1 = 0;
        for (let i = bar - longWindow + 1; i <= bar; i++) {
            avg1 += this.value1Arr[i];
        }
        avg1 /= longWindow;

        let avg2 = 0;
        for (let i = bar - shortWindow + 1; i <= bar; i++) {
            avg2 += this.value1Arr[i];
        }
        avg2 /= shortWindow;

        const value2 = this.sensitivity * Math.log((this.eps + avg1) / (this.eps + avg2));
        const value3 = value2 / (1.0 + Math.abs(value2));

        return this.loLength + (this.hiLength - this.loLength) * (1.0 + value3) / 2.0;
    }

    private computeWLSSlope(bar: number, depth: number): number {
        const n = depth + 1;
        const s1 = n * (n + 1) / 2.0;
        const s2 = s1 * (2 * n + 1) / 3.0;
        const denom = s1 * s1 * s1 - s2 * s2;

        let sumXW = 0;
        let sumXW2 = 0;

        for (let i = 0; i <= depth; i++) {
            const w = n - i;
            const p = this.prices[bar - i];
            sumXW += p * w;
            sumXW2 += p * w * w;
        }

        return (sumXW2 * s1 - sumXW * s2) / denom;
    }

    /** Updates the value of the JAVEL indicator given the next sample. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        const bar = this.barCount;
        this.barCount++;

        this.prices.push(sample);

        // Compute value1 (abs diff).
        if (bar === 0) {
            this.value1Arr.push(0.0);
        } else {
            this.value1Arr.push(Math.abs(sample - this.prices[bar - 1]));
        }

        // Compute adaptive depth.
        const adaptiveDepth = this.computeAdaptiveDepth(bar);
        const depth = Math.ceil(adaptiveDepth);

        // Check if we have enough prices for WLS.
        if (bar < depth) {
            return NaN;
        }

        // Stage 1: WLS slope.
        const slope = this.computeWLSSlope(bar, depth);

        // Stage 2: adaptive smoother.
        const result = this.smooth.update(slope);

        if (!this.primed) {
            this.primed = true;
        }

        return result;
    }
}
