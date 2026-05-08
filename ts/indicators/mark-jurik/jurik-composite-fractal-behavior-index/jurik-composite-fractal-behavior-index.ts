import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { JurikCompositeFractalBehaviorIndexParams } from './params';

/** Depth sets for each FractalType (1–4). */
const depthSets: number[][] = [
    [2, 3, 4, 6, 8, 12, 16, 24],                             // Type 1: JCFB24
    [2, 3, 4, 6, 8, 12, 16, 24, 32, 48],                     // Type 2: JCFB48
    [2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96],             // Type 3: JCFB96
    [2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 192],   // Type 4: JCFB192
];

/** Weights for even-indexed channels. */
const weightsEven = [2, 3, 6, 12, 24, 48, 96];
/** Weights for odd-indexed channels. */
const weightsOdd = [4, 8, 16, 32, 64, 128, 256];

/** Function to calculate mnemonic of a JurikCompositeFractalBehaviorIndex indicator. */
export const jurikCompositeFractalBehaviorIndexMnemonic = (params: JurikCompositeFractalBehaviorIndexParams): string =>
    'jcfb('.concat(
        params.fractalType.toString(),
        ',',
        params.smooth.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
        ')');

/** Streaming state for a single CFB auxiliary instance. */
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

        const srcSize = this.depth + 2;
        this.src[this.srcIdx] = sample;
        this.srcIdx = (this.srcIdx + 1) % srcSize;

        let intAVal: number;

        if (this.firstCall) {
            this.firstCall = false;
            this.prevSample = sample;
            return 0;
        }

        intAVal = Math.abs(sample - this.prevSample);
        this.prevSample = sample;

        const oldIntA = this.intA[this.intAIdx];
        this.intA[this.intAIdx] = intAVal;
        this.intAIdx = (this.intAIdx + 1) % this.depth;

        const refBar = this.bar - 1;
        if (refBar < this.depth) {
            return 0;
        }

        if (refBar <= this.depth * 2) {
            this.jrc04 = 0;
            this.jrc05 = 0;
            this.jrc06 = 0;

            const curIntAPos = (this.intAIdx - 1 + this.depth) % this.depth;
            const curSrcPos = (this.srcIdx - 1 + srcSize) % srcSize;

            for (let j = 0; j < this.depth; j++) {
                const intAPos = (curIntAPos - j + this.depth) % this.depth;
                const intAV = this.intA[intAPos];

                const srcPos = (curSrcPos - j - 1 + srcSize * 2) % srcSize;
                const srcV = this.src[srcPos];

                this.jrc04 += intAV;
                this.jrc05 += (this.depth - j) * intAV;
                this.jrc06 += srcV;
            }
        } else {
            this.jrc05 = this.jrc05 - this.jrc04 + intAVal * this.depth;
            this.jrc04 = this.jrc04 - oldIntA + intAVal;

            const curSrcPos = (this.srcIdx - 1 + srcSize) % srcSize;
            const srcBarMinus1 = (curSrcPos - 1 + srcSize) % srcSize;
            const srcBarMinusDepthMinus1 = (curSrcPos - this.depth - 1 + srcSize) % srcSize;

            this.jrc06 = this.jrc06 - this.src[srcBarMinusDepthMinus1] + this.src[srcBarMinus1];
        }

        const curSrcPos = (this.srcIdx - 1 + srcSize) % srcSize;
        const jrc08 = Math.abs(this.depth * this.src[curSrcPos] - this.jrc06);

        if (this.jrc05 === 0) {
            return 0;
        }

        return jrc08 / this.jrc05;
    }
}

/** Jurik Composite Fractal Behavior Index line indicator. */
export class JurikCompositeFractalBehaviorIndex extends LineIndicator {
    private readonly paramFractal: number;
    private readonly paramSmooth: number;
    private readonly numChannels: number;
    private readonly auxInstances: CfbAux[];
    private readonly auxWindows: number[][];
    private auxWinIdx = 0;
    private readonly er23: number[];
    private bar = 0;
    private er19 = 20;

    /**
     * Constructs an instance using given parameters.
     */
    public constructor(params: JurikCompositeFractalBehaviorIndexParams) {
        super();

        const fractalType = params.fractalType;
        const smooth = params.smooth;

        if (fractalType < 1 || fractalType > 4) {
            throw new Error('invalid jurik composite fractal behavior index parameters: fractal type should be between 1 and 4');
        }

        if (smooth < 1) {
            throw new Error('invalid jurik composite fractal behavior index parameters: smooth should be at least 1');
        }

        this.paramFractal = fractalType;
        this.paramSmooth = smooth;

        this.mnemonic = jurikCompositeFractalBehaviorIndexMnemonic(params);
        this.description = 'Jurik composite fractal behavior index ' + this.mnemonic;
        this.primed = false;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;

        const depths = depthSets[fractalType - 1];
        this.numChannels = depths.length;

        this.auxInstances = depths.map(d => new CfbAux(d));
        this.auxWindows = Array.from({ length: this.numChannels }, () => new Array(smooth).fill(0));
        this.er23 = new Array(this.numChannels).fill(0);
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.JurikCompositeFractalBehaviorIndex,
            this.mnemonic,
            this.description,
            [
                { mnemonic: this.mnemonic, description: this.description },
            ],
        );
    }

    /** Updates the value of the CFB indicator given the next sample. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        this.bar++;

        // Feed all aux instances.
        const auxValues = new Array(this.numChannels);
        for (let i = 0; i < this.numChannels; i++) {
            auxValues[i] = this.auxInstances[i].update(sample);
        }

        if (this.bar === 1) {
            return NaN;
        }

        const refBar = this.bar - 1;
        const smooth = this.paramSmooth;

        if (refBar <= smooth) {
            const winPos = this.auxWinIdx;
            for (let i = 0; i < this.numChannels; i++) {
                this.auxWindows[i][winPos] = auxValues[i];
            }
            this.auxWinIdx = (this.auxWinIdx + 1) % smooth;

            for (let i = 0; i < this.numChannels; i++) {
                let sum = 0;
                for (let j = 0; j < refBar; j++) {
                    const pos = (this.auxWinIdx - 1 - j + smooth * 2) % smooth;
                    sum += this.auxWindows[i][pos];
                }
                this.er23[i] = sum / refBar;
            }
        } else {
            const winPos = this.auxWinIdx;
            for (let i = 0; i < this.numChannels; i++) {
                const oldVal = this.auxWindows[i][winPos];
                this.auxWindows[i][winPos] = auxValues[i];
                this.er23[i] += (auxValues[i] - oldVal) / smooth;
            }
            this.auxWinIdx = (this.auxWinIdx + 1) % smooth;
        }

        // Compute weighted composite (only when refBar > 5).
        if (refBar > 5) {
            const n = this.numChannels;
            const er22 = new Array(n);

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
                    er17 += sq * weightsEven[Math.floor(idx / 2)];
                } else {
                    er17 += sq * weightsOdd[Math.floor(idx / 2)];
                }
            }

            if (er18 === 0) {
                this.er19 = 0;
            } else {
                this.er19 = er17 / er18;
            }
        }

        if (!this.primed) {
            if (refBar > 5) {
                this.primed = true;
            }
        }

        return this.er19;
    }
}
