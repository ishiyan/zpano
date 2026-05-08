import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { JurikFractalAdaptiveZeroLagVelocityParams } from './params';

/** Function to calculate mnemonic of a JurikFractalAdaptiveZeroLagVelocity indicator. */
export const jurikFractalAdaptiveZeroLagVelocityMnemonic = (params: JurikFractalAdaptiveZeroLagVelocityParams): string =>
    'jvelcfb('.concat(
        params.loDepth.toString(), ', ',
        params.hiDepth.toString(), ', ',
        params.fractalType.toString(), ', ',
        params.smooth.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
        ')');

// Scale sets for different fractal types.
const scaleSets: { [key: number]: number[] } = {
    1: [2, 3, 4, 6, 8, 12, 16, 24],
    2: [2, 3, 4, 6, 8, 12, 16, 24, 32, 48],
    3: [2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96],
    4: [2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192],
};

const weightsEven = [2, 3, 6, 12, 24, 48, 96];
const weightsOdd = [4, 8, 16, 32, 64, 128, 256];

/** Streaming state for a single CFB channel at one depth. */
class CfbAux {
    private readonly depth: number;
    private bar = 0;
    private readonly intA: number[];
    private intAIdx = 0;
    private readonly src: number[];
    private srcIdx = 0;
    private jrc04 = 0;
    private jrc05 = 0;
    private jrc06 = 0;
    private prevSample = 0;
    private firstCall = true;

    constructor(depth: number) {
        this.depth = depth;
        this.intA = new Array(depth).fill(0);
        this.src = new Array(depth + 2).fill(0);
    }

    public update(sample: number): number {
        this.bar++;
        const depth = this.depth;
        const srcSize = depth + 2;

        this.src[this.srcIdx] = sample;
        this.srcIdx = (this.srcIdx + 1) % srcSize;

        if (this.firstCall) {
            this.firstCall = false;
            this.prevSample = sample;
            return 0.0;
        }

        const intAVal = Math.abs(sample - this.prevSample);
        this.prevSample = sample;

        const oldIntA = this.intA[this.intAIdx];
        this.intA[this.intAIdx] = intAVal;
        this.intAIdx = (this.intAIdx + 1) % depth;

        const refBar = this.bar - 1;
        if (refBar < depth) {
            return 0.0;
        }

        if (refBar <= depth * 2) {
            // Recompute from scratch.
            this.jrc04 = 0.0;
            this.jrc05 = 0.0;
            this.jrc06 = 0.0;

            const curIntAPos = (this.intAIdx - 1 + depth) % depth;
            const curSrcPos = (this.srcIdx - 1 + srcSize) % srcSize;

            for (let j = 0; j < depth; j++) {
                const intAPos = (curIntAPos - j + depth) % depth;
                const srcPos = (curSrcPos - j - 1 + srcSize * 2) % srcSize;

                this.jrc04 += this.intA[intAPos];
                this.jrc05 += (depth - j) * this.intA[intAPos];
                this.jrc06 += this.src[srcPos];
            }
        } else {
            // Incremental update.
            this.jrc05 = this.jrc05 - this.jrc04 + intAVal * depth;
            this.jrc04 = this.jrc04 - oldIntA + intAVal;

            const curSrcPos = (this.srcIdx - 1 + srcSize) % srcSize;
            const srcBarMinus1 = (curSrcPos - 1 + srcSize) % srcSize;
            const srcBarMinusDepthMinus1 = (curSrcPos - depth - 1 + srcSize * 2) % srcSize;

            this.jrc06 = this.jrc06 - this.src[srcBarMinusDepthMinus1] + this.src[srcBarMinus1];
        }

        const curSrcPos = (this.srcIdx - 1 + srcSize) % srcSize;
        const jrc08 = Math.abs(depth * this.src[curSrcPos] - this.jrc06);

        if (this.jrc05 === 0.0) {
            return 0.0;
        }

        return jrc08 / this.jrc05;
    }
}

/** Composite Fractal Behavior weighted dominant cycle. */
class Cfb {
    private readonly scales: number[];
    private readonly numChannels: number;
    private readonly auxs: CfbAux[];
    private readonly auxWindows: number[][];
    private auxWinIdx = 0;
    private readonly er23: number[];
    private readonly smooth: number;
    private bar = 0;
    private cfbValue = 0;

    constructor(fractalType: number, smooth: number) {
        this.scales = scaleSets[fractalType];
        this.numChannels = this.scales.length;
        this.auxs = this.scales.map(d => new CfbAux(d));
        this.auxWindows = Array.from({ length: this.numChannels }, () => new Array(smooth).fill(0));
        this.er23 = new Array(this.numChannels).fill(0);
        this.smooth = smooth;
    }

    public update(sample: number): number {
        this.bar++;
        const refBar = this.bar - 1;

        const auxValues = this.auxs.map(aux => aux.update(sample));

        if (refBar === 0) {
            return 0.0;
        }

        const smooth = this.smooth;
        const n = this.numChannels;

        if (refBar <= smooth) {
            const winPos = this.auxWinIdx;
            for (let i = 0; i < n; i++) {
                this.auxWindows[i][winPos] = auxValues[i];
            }
            this.auxWinIdx = (this.auxWinIdx + 1) % smooth;

            for (let i = 0; i < n; i++) {
                let s = 0.0;
                for (let j = 0; j < refBar; j++) {
                    const pos = (this.auxWinIdx - 1 - j + smooth * 2) % smooth;
                    s += this.auxWindows[i][pos];
                }
                this.er23[i] = s / refBar;
            }
        } else {
            const winPos = this.auxWinIdx;
            for (let i = 0; i < n; i++) {
                const oldVal = this.auxWindows[i][winPos];
                this.auxWindows[i][winPos] = auxValues[i];
                this.er23[i] += (auxValues[i] - oldVal) / smooth;
            }
            this.auxWinIdx = (this.auxWinIdx + 1) % smooth;
        }

        if (refBar > 5) {
            const er22 = new Array(n).fill(0);

            // Odd-indexed channels (descending).
            let er15 = 1.0;
            for (let idx = n - 1; idx >= 1; idx -= 2) {
                er22[idx] = er15 * this.er23[idx];
                er15 *= (1 - er22[idx]);
            }

            // Even-indexed channels (descending).
            let er16 = 1.0;
            for (let idx = n - 2; idx >= 0; idx -= 2) {
                er22[idx] = er16 * this.er23[idx];
                er16 *= (1 - er22[idx]);
            }

            // Weighted sum.
            let er17 = 0;
            let er18 = 0;
            for (let idx = 0; idx < n; idx++) {
                const sq = er22[idx] * er22[idx];
                er18 += sq;

                if (idx % 2 === 0) {
                    er17 += sq * weightsEven[Math.trunc(idx / 2)];
                } else {
                    er17 += sq * weightsOdd[Math.trunc(idx / 2)];
                }
            }

            if (er18 === 0.0) {
                this.cfbValue = 0.0;
            } else {
                this.cfbValue = er17 / er18;
            }
        }

        return this.cfbValue;
    }
}

/** Adaptive smoother (Stage 2) with fixed period=3.0. */
class VelSmooth {
    private readonly jrc03: number;
    private readonly jrc06: number;
    private readonly jrc07: number;
    private readonly emaFactor: number;
    private readonly damping: number;
    private readonly eps2 = 0.0001;
    private readonly bufferSize = 1001;
    private readonly buffer: number[];
    private idx = 0;
    private length = 0;
    private velocity = 0;
    private position = 0;
    private smoothedMAD = 0;
    private madInit = false;
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
        this.buffer[this.idx] = value;
        this.idx = (this.idx + 1) % this.bufferSize;
        this.length++;

        if (this.length > this.bufferSize) {
            this.length = this.bufferSize;
        }

        const length = this.length;

        if (!this.initialized) {
            this.initialized = true;
            this.position = value;
            this.velocity = 0.0;
            this.smoothedMAD = 0.0;
            return this.position;
        }

        // Linear regression over capped window.
        let n = length;
        if (n > this.jrc06) {
            n = this.jrc06;
        }

        let sx = 0; let sy = 0; let sxy = 0; let sx2 = 0;

        for (let i = 0; i < n; i++) {
            let bufIdx = (this.idx - 1 - i + this.bufferSize) % this.bufferSize;
            const x = i;
            const y = this.buffer[bufIdx];
            sx += x;
            sy += y;
            sxy += x * y;
            sx2 += x * x;
        }

        const fn = n;
        let slope = 0;
        if (n > 1) {
            slope = (fn * sxy - sx * sy) / (fn * sx2 - sx * sx);
        }

        const intercept = (sy - slope * sx) / fn;

        // MAD from regression residuals.
        let mad = 0;
        for (let i = 0; i < n; i++) {
            let bufIdx = (this.idx - 1 - i + this.bufferSize) % this.bufferSize;
            const predicted = intercept + slope * i;
            mad += Math.abs(this.buffer[bufIdx] - predicted);
        }
        mad /= fn;

        // Scale MAD.
        const scaledMAD = mad * 1.2 * Math.pow(this.jrc06 / fn, 0.25);

        // Smooth MAD with EMA.
        if (!this.madInit) {
            this.smoothedMAD = scaledMAD;
            if (scaledMAD > 0) {
                this.madInit = true;
            }
        } else {
            this.smoothedMAD += (scaledMAD - this.smoothedMAD) * this.emaFactor;
        }

        const smoothedMAD = Math.max(this.eps2, this.smoothedMAD);

        // Adaptive velocity/position dynamics.
        const predictionError = value - this.position;
        const responseFactor = 1.0 - Math.exp(-Math.abs(predictionError) / (smoothedMAD * this.jrc03));
        this.velocity = responseFactor * predictionError + this.velocity * this.damping;
        this.position += this.velocity;

        return this.position;
    }
}

/** Jurik Fractal Adaptive Zero Lag Velocity (JVELCFB) line indicator. */
export class JurikFractalAdaptiveZeroLagVelocity extends LineIndicator {
    private readonly loDepth: number;
    private readonly hiDepth: number;

    private prices: number[] = [];
    private barCount = 0;
    private readonly cfbInst: Cfb;
    private cfbMin: number | null = null;
    private cfbMax: number | null = null;
    private readonly velSmooth: VelSmooth;

    public constructor(params: JurikFractalAdaptiveZeroLagVelocityParams) {
        super();

        if (params.loDepth < 2) {
            throw new Error('invalid jurik fractal adaptive zero lag velocity parameters: lo_depth should be at least 2');
        }
        if (params.hiDepth < params.loDepth) {
            throw new Error('invalid jurik fractal adaptive zero lag velocity parameters: hi_depth should be at least lo_depth');
        }
        if (params.fractalType < 1 || params.fractalType > 4) {
            throw new Error('invalid jurik fractal adaptive zero lag velocity parameters: fractal_type should be 1-4');
        }
        if (params.smooth < 1) {
            throw new Error('invalid jurik fractal adaptive zero lag velocity parameters: smooth should be at least 1');
        }

        this.loDepth = params.loDepth;
        this.hiDepth = params.hiDepth;
        this.cfbInst = new Cfb(params.fractalType, params.smooth);
        this.velSmooth = new VelSmooth(3.0);
        this.mnemonic = jurikFractalAdaptiveZeroLagVelocityMnemonic(params);
        this.description = 'Jurik fractal adaptive zero lag velocity ' + this.mnemonic;
        this.primed = false;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.JurikFractalAdaptiveZeroLagVelocity,
            this.mnemonic,
            this.description,
            [{ mnemonic: this.mnemonic, description: this.description }],
        );
    }

    /** Updates the value of the JVELCFB indicator given the next sample. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        const bar = this.barCount;
        this.barCount++;

        this.prices.push(sample);

        // CFB computation.
        const cfbVal = this.cfbInst.update(sample);

        if (bar === 0) {
            return NaN;
        }

        // Stochastic normalization.
        if (this.cfbMin === null) {
            this.cfbMin = cfbVal;
            this.cfbMax = cfbVal;
        } else {
            if (cfbVal < this.cfbMin) { this.cfbMin = cfbVal; }
            if (cfbVal > this.cfbMax!) { this.cfbMax = cfbVal; }
        }

        const cfbRange = this.cfbMax! - this.cfbMin;
        let sr: number;
        if (cfbRange !== 0.0) {
            sr = (cfbVal - this.cfbMin) / cfbRange;
        } else {
            sr = 0.5;
        }

        const depthF = this.loDepth + sr * (this.hiDepth - this.loDepth);
        const depth = Math.round(depthF);

        // Stage 1: WLS slope.
        if (bar < depth) {
            return NaN;
        }

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

        const slope = (sumXW2 * s1 - sumXW * s2) / denom;

        // Stage 2: adaptive smoother.
        const result = this.velSmooth.update(slope);

        if (!this.primed) {
            this.primed = true;
        }

        return result;
    }
}
