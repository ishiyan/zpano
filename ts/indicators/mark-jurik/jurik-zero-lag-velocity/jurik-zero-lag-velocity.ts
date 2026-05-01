import { buildMetadata } from '../../core/build-metadata';
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { LineIndicator } from '../../core/line-indicator';
import { JurikZeroLagVelocityParams } from './params';

/** Function to calculate mnemonic of a JurikZeroLagVelocity indicator. */
export const jurikZeroLagVelocityMnemonic = (params: JurikZeroLagVelocityParams): string =>
    'vel('.concat(
        params.depth.toString(),
        componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
        ')');

/** Linear regression slope over a window of depth+1 points. */
class VelAux1 {
    private readonly depth: number;
    private readonly win: number[];
    private idx = 0;
    private bar = 0;

    private readonly jrc04: number;
    private readonly jrc05: number;
    private readonly jrc06: number;
    private readonly jrc07: number;

    constructor(depth: number) {
        this.depth = depth;
        const size = depth + 1;
        this.win = new Array(size).fill(0);
        const jrc04 = size;
        const jrc05 = jrc04 * (jrc04 + 1) / 2;
        const jrc06 = jrc05 * (2 * jrc04 + 1) / 3;
        this.jrc04 = jrc04;
        this.jrc05 = jrc05;
        this.jrc06 = jrc06;
        this.jrc07 = jrc05 * jrc05 * jrc05 - jrc06 * jrc06;
    }

    public update(sample: number): number {
        const size = this.depth + 1;
        this.win[this.idx] = sample;
        this.idx = (this.idx + 1) % size;
        this.bar++;

        if (this.bar <= this.depth) {
            return 0;
        }

        let jrc08 = 0;
        let jrc09 = 0;

        for (let j = 0; j <= this.depth; j++) {
            const pos = (this.idx - 1 - j + size * 2) % size;
            const w = this.jrc04 - j;
            jrc08 += this.win[pos] * w;
            jrc09 += this.win[pos] * w * w;
        }

        return (jrc09 * this.jrc05 - jrc08 * this.jrc06) / this.jrc07;
    }
}

/** Adaptive smoother state for VEL aux3. */
class VelAux3State {
    private readonly length = 30;
    private readonly eps = 0.0001;
    private readonly decay = 3;
    private readonly beta: number;
    private readonly alpha: number;
    private readonly maxWin: number;

    private readonly srcRing = new Array(100).fill(0);
    private readonly devRing = new Array(100).fill(0);
    private srcIdx = 0;
    private devIdx = 0;

    private jr08 = 0;
    private jr09 = 0;
    private jr10 = 0;
    private jr11 = 0;
    private jr12 = 0;
    private jr13 = 0;
    private jr14 = 0;
    private jr19 = 0;
    private jr20 = 0;
    private jr21 = 0;
    private jr21a = 0;
    private jr21b = 0;
    private jr22 = 0;
    private jr23 = 0;

    private bar = 0;
    private initDone = false;
    private history: number[] = [];

    constructor() {
        this.beta = 0.86 - 0.55 / Math.sqrt(this.decay);
        this.alpha = 1 - Math.exp(-Math.log(4) / this.decay / 2);
        this.maxWin = this.length + 1;
    }

    public feed(sample: number, barIdx: number): number {
        if (barIdx < this.length) {
            this.history.push(sample);
            return 0;
        }

        this.bar++;

        if (!this.initDone) {
            this.initDone = true;

            let jr28 = 0;
            for (let j = 1; j <= this.length - 1; j++) {
                if (this.history[this.history.length - j] === this.history[this.history.length - j - 1]) {
                    jr28++;
                }
            }

            let jr26: number;
            if (jr28 < this.length - 1) {
                jr26 = barIdx - this.length;
            } else {
                jr26 = barIdx;
            }

            this.jr11 = Math.trunc(Math.min(1 + (barIdx - jr26), this.maxWin));

            this.jr21 = this.history[this.history.length - 1];
            const jr07 = 3;
            this.jr08 = (sample - this.history[this.history.length - jr07]) / jr07;

            for (let jr15 = this.jr11 - 1; jr15 >= 1; jr15--) {
                if (this.srcIdx <= 0) {
                    this.srcIdx = 100;
                }
                this.srcIdx--;
                this.srcRing[this.srcIdx] = this.history[this.history.length - jr15];
            }

            this.history = [];
        }

        // Push current value to source ring.
        if (this.srcIdx <= 0) {
            this.srcIdx = 100;
        }
        this.srcIdx--;
        this.srcRing[this.srcIdx] = sample;

        if (this.jr11 <= this.length) {
            // Growing phase.
            if (this.bar === 1) {
                this.jr21 = sample;
            } else {
                this.jr21 = Math.sqrt(this.alpha) * sample + (1 - Math.sqrt(this.alpha)) * this.jr21a;
            }

            if (this.bar > 2) {
                this.jr08 = (this.jr21 - this.jr21b) / 2;
            } else {
                this.jr08 = 0;
            }

            this.jr11++;
        } else if (this.jr11 <= this.maxWin) {
            // Transition phase: recompute from scratch.
            this.jr12 = this.jr11 * (this.jr11 + 1) * (this.jr11 - 1) / 12;
            this.jr13 = (this.jr11 + 1) / 2;
            this.jr14 = (this.jr11 - 1) / 2;

            this.jr09 = 0;
            this.jr10 = 0;

            for (let jr15 = this.jr11 - 1; jr15 >= 0; jr15--) {
                const jr24 = (this.srcIdx + jr15) % 100;
                this.jr09 += this.srcRing[jr24];
                this.jr10 += this.srcRing[jr24] * (this.jr14 - jr15);
            }

            const jr16 = this.jr10 / this.jr12;
            let jr17 = (this.jr09 / this.jr11) - (jr16 * this.jr13);

            this.jr19 = 0;
            for (let jr15 = this.jr11 - 1; jr15 >= 0; jr15--) {
                jr17 += jr16;
                const jr24 = (this.srcIdx + jr15) % 100;
                this.jr19 += Math.abs(this.srcRing[jr24] - jr17);
            }

            this.jr20 = (this.jr19 / this.jr11) * Math.pow(this.maxWin / this.jr11, 0.25);
            this.jr11++;

            // Adaptive step.
            this.jr20 = Math.max(this.eps, this.jr20);
            this.jr22 = sample - (this.jr21 + this.jr08 * this.beta);
            this.jr23 = 1 - Math.exp(-Math.abs(this.jr22) / this.jr20 / this.decay);
            this.jr08 = this.jr23 * this.jr22 + this.jr08 * this.beta;
            this.jr21 += this.jr08;
        } else {
            // Steady state.
            const jr24out = (this.srcIdx + this.maxWin) % 100;
            this.jr10 = this.jr10 - this.jr09 + this.srcRing[jr24out] * this.jr13 + sample * this.jr14;
            this.jr09 = this.jr09 - this.srcRing[jr24out] + sample;

            // Deviation ring update.
            if (this.devIdx <= 0) {
                this.devIdx = this.maxWin;
            }
            this.devIdx--;
            this.jr19 -= this.devRing[this.devIdx];

            const jr16 = this.jr10 / this.jr12;
            const jr17 = (this.jr09 / this.maxWin) + (jr16 * this.jr14);
            this.devRing[this.devIdx] = Math.abs(sample - jr17);
            this.jr19 = Math.max(this.eps, this.jr19 + this.devRing[this.devIdx]);
            this.jr20 += ((this.jr19 / this.maxWin) - this.jr20) * this.alpha;

            // Adaptive step.
            this.jr20 = Math.max(this.eps, this.jr20);
            this.jr22 = sample - (this.jr21 + this.jr08 * this.beta);
            this.jr23 = 1 - Math.exp(-Math.abs(this.jr22) / this.jr20 / this.decay);
            this.jr08 = this.jr23 * this.jr22 + this.jr08 * this.beta;
            this.jr21 += this.jr08;
        }

        this.jr21b = this.jr21a;
        this.jr21a = this.jr21;

        return this.jr21;
    }
}

/** Jurik Zero Lag Velocity line indicator. */
export class JurikZeroLagVelocity extends LineIndicator {
    private readonly paramDepth: number;
    private readonly aux1: VelAux1;
    private readonly aux3: VelAux3State;
    private bar = 0;

    /**
     * Constructs an instance using given parameters.
     */
    public constructor(params: JurikZeroLagVelocityParams) {
        super();

        const depth = params.depth;
        if (depth < 2) {
            throw new Error('invalid jurik zero lag velocity parameters: depth should be at least 2');
        }

        this.paramDepth = depth;
        this.mnemonic = jurikZeroLagVelocityMnemonic(params);
        this.description = 'Jurik zero lag velocity ' + this.mnemonic;
        this.primed = false;
        this.barComponent = params.barComponent;
        this.quoteComponent = params.quoteComponent;
        this.tradeComponent = params.tradeComponent;

        this.aux1 = new VelAux1(depth);
        this.aux3 = new VelAux3State();
    }

    /** Describes the output data of the indicator. */
    public metadata(): IndicatorMetadata {
        return buildMetadata(
            IndicatorIdentifier.JurikZeroLagVelocity,
            this.mnemonic,
            this.description,
            [
                { mnemonic: this.mnemonic, description: this.description },
            ],
        );
    }

    /** Updates the value of the VEL indicator given the next sample. */
    public update(sample: number): number {
        if (Number.isNaN(sample)) {
            return sample;
        }

        // Stage 1: compute linear regression slope.
        const aux1Val = this.aux1.update(sample);

        // Stage 2: feed into adaptive smoother.
        const barIdx = this.bar;
        this.bar++;

        const result = this.aux3.feed(aux1Val, barIdx);

        if (barIdx < 30) {
            return NaN;
        }

        if (!this.primed) {
            this.primed = true;
        }

        return result;
    }
}
