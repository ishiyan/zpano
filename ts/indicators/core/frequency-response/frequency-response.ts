/**
 * Represents a filter which frequency response is to be calculated.
 **/
 export interface FrequencyResponseFilter {
    metadata(): { mnemonic: string };
    update(sample: number): number;
}

/**
 * Contains a single calculated filter frequency response component data.
 * All arrays have the same spectrum length.
 **/
 export interface FrequencyResponseComponent {
    data: number[];
    min: number;
    max: number;
}

/**
 * Contains calculated filter frequency response data.
 * All arrays have the same spectrum length.
 **/
 export interface FrequencyResponseResult {
    label: string;
    /**
     * Normalized frequencies in range (0, 1] expressed in units of **cycles per sample**.
     *
     * The maximal value of **1** corresponds to the Nyquist frequency, or the 'one cycle per two samples'.
     *
     * The minimal value of **0** corresponds to the zero frequency, or the 'one cycle per infinite samples'.
     **/
    frequencies: number[];
    /** Spectrum power in perscentages from a maximum value. */
    powerPercent: FrequencyResponseComponent;
    /** Spectrum power in decibels. */
    powerDecibel: FrequencyResponseComponent;
    /** Spectrum amplitude in perscentages from a maximum value. */
    amplitudePercent: FrequencyResponseComponent;
    /** Spectrum amplitude in decibels. */
    amplitudeDecibel: FrequencyResponseComponent;
    /** Phase in degrees in range [-180, 180]. */
    phaseDegrees: FrequencyResponseComponent;
    /** Phase in degrees unwrapped. */
    phaseDegreesUnwrapped: FrequencyResponseComponent;
}

export class FrequencyResponse {

    public static calculate(signalLength: number, filter: FrequencyResponseFilter, warmup: number,
        phaseDegreesUnwrappingLimit = 179, filteredSignal: number[] = []): FrequencyResponseResult {
        if (!FrequencyResponse.isValidSignalLength(signalLength)) {
            throw new Error('signal length should be power of 2 and not less than 4');
        }

        const spectrumLength = signalLength / 2 - 1;
        const fr: FrequencyResponseResult = {
            label: filter.metadata().mnemonic,
            frequencies: FrequencyResponse.prepareFrequencyDomain(spectrumLength),
            powerPercent: FrequencyResponse.createFrequencyResponseComponent(spectrumLength),
            powerDecibel: FrequencyResponse.createFrequencyResponseComponent(spectrumLength),
            amplitudePercent: FrequencyResponse.createFrequencyResponseComponent(spectrumLength),
            amplitudeDecibel: FrequencyResponse.createFrequencyResponseComponent(spectrumLength),
            phaseDegrees: FrequencyResponse.createFrequencyResponseComponent(spectrumLength),
            phaseDegreesUnwrapped: FrequencyResponse.createFrequencyResponseComponent(spectrumLength)
        };

        const signal = filteredSignal.length === signalLength ?
            filteredSignal : FrequencyResponse.prepareFilteredSignal(signalLength, filter, warmup);
        FrequencyResponse.directRealFastFourierTransform(signal);
        FrequencyResponse.parseSpectrum(spectrumLength, signal, fr.powerPercent, fr.amplitudePercent,
            fr.phaseDegrees, fr.phaseDegreesUnwrapped, phaseDegreesUnwrappingLimit);
        FrequencyResponse.toDecibels(spectrumLength, fr.powerPercent, fr.powerDecibel);
        FrequencyResponse.toPercents(spectrumLength, fr.powerPercent, fr.powerPercent);
        FrequencyResponse.toDecibels(spectrumLength, fr.amplitudePercent, fr.amplitudeDecibel);
        FrequencyResponse.toPercents(spectrumLength, fr.amplitudePercent, fr.amplitudePercent);

        return fr;
    }

    protected static isValidSignalLength(len: number): boolean {
        while (len > 4) {
            if (len % 2 !== 0) {
                return false;
            }

            len /= 2;
        }

        return len === 4;
    }

    protected static prepareFrequencyDomain(len: number): number[] {
        const freq = new Array<number>(len);
        for (let i = 0; i < len; ++i) {
            freq[i] = (1 + i) / len;
        }

        return freq;
    }

    protected static prepareFilteredSignal(len: number, filter: FrequencyResponseFilter, warmup: number): number[] {
        for (let i = 0; i < warmup; ++i) {
            filter.update(0);
        }

        const signal = new Array<number>(len);
        signal[0] = filter.update(1000);

        for (let i = 1; i < len; ++i) {
            signal[i] = filter.update(0);
        }

        return signal;
    }

    protected static createFrequencyResponseComponent(len: number): FrequencyResponseComponent {
        return { min: -Infinity, max: Infinity, data: new Array<number>(len) };
    }

    protected static parseSpectrum(len: number, spectrum: number[],
        power: FrequencyResponseComponent, amplitude: FrequencyResponseComponent,
        phase: FrequencyResponseComponent, phaseUnwrapped: FrequencyResponseComponent,
        phaseDegreesUnwrappingLimit: number) {
        const rad2deg = 180 / Math.PI;
        let pmin = Infinity;
        let pmax = -Infinity;
        let amin = Infinity;
        let amax = -Infinity;

        for (let i = 0, k = 2; i < len; ++i) {
            const re = spectrum[k++];
            const im = spectrum[k++];

            // Wrapped phase -- atan2 returns radians in the [-π, π] range.
            // We convert them into [-180, 180] dergee range.
            phase.data[i] = -Math.atan2(im, re) * rad2deg;
            phaseUnwrapped.data[i] = 0;

            const pwr = re * re + im * im;
            power.data[i] = pwr;
            pmin = Math.min(pmin, pwr);
            pmax = Math.max(pmax, pwr);

            const amp = Math.sqrt(pwr);
            amplitude.data[i] = amp;
            amin = Math.min(amin, amp);
            amax = Math.max(amax, amp);
        }

        FrequencyResponse.unwrapPhaseDegrees(len, phase.data, phaseUnwrapped, phaseDegreesUnwrappingLimit);
        phase.min = -180;
        phase.max = 180;
        power.min = pmin;
        power.max = pmax;
        amplitude.min = amin;
        amplitude.max = amax;
    }

    /** Unwraps phase degrees from the [-180, 180] range. */
    protected static unwrapPhaseDegrees(len: number, wrapped: number[],
        unwrapped: FrequencyResponseComponent, phaseDegreesUnwrappingLimit: number) {
        let k = 0;

        let min = wrapped[0];
        let max = min;
        unwrapped.data[0] = min;

        for (let i = 1; i < len; ++i) {
            let w = wrapped[i];
            const increment = wrapped[i] - wrapped[i-1];
            if (increment > phaseDegreesUnwrappingLimit) {
                k -= increment;
            } else if (increment < -phaseDegreesUnwrappingLimit) {
                k += increment;
            }

            w += k;
            min = Math.min(min, w);
            max = Math.max(max, w);
            unwrapped.data[i] = w;
        }

        unwrapped.min = min;
        unwrapped.max = max;
    }

    protected static toDecibels(len: number, src: FrequencyResponseComponent, tgt: FrequencyResponseComponent) {
        const five = 5;
        const ten = 10;
        const twenty = 20;
        const hundreed = 100;
        let dbmin = Infinity;
        let dbmax = -Infinity;
        let base = src.data[0];
        if (base < Number.EPSILON) {
            base = src.max;
        }

        for (let i = 0; i < len; ++i) {
            const db = twenty * Math.log10(src.data[i] / base);
            dbmin = Math.min(dbmin, db);
            dbmax = Math.max(dbmax, db);
            tgt.data[i] = db;
        }

        // If dbmin falls into one of [-100, -90), [-90, -80), ..., [-10, 0)
        // intervals, set it to the minimum value of the interval.
        for (let i = ten; i > 0; --i) {
            const min = -i * ten;
            const max = -(i - 1) * ten;
            if (dbmin >= min && dbmin < max) {
                dbmin = min;
                break;
            }
        }

        // Limit all minimal decibel values to -100.
        if (dbmin < -hundreed) {
            dbmin = -hundreed;
            for (let i = 0; i < len; ++i) {
                if (tgt.data[i] < -hundreed) {
                    tgt.data[i] = -hundreed;
                }
            }
        }

        // If dbmax falls into one of [0, 5), [5, 10)
        // intervals, set it to the maximum value of the interval.
        for (let i = 2; i > 0; --i) {
            const max = i * five;
            const min = (i - 1) * five;
            if (dbmax >= min && dbmax < max) {
                dbmax = max;
                break;
            }
        }

        // Limit all maximal decibel values to 10.
        if (dbmax > ten) {
            dbmax = ten;
            for (let i = 0; i < len; ++i) {
                if (tgt.data[i] > ten) {
                    tgt.data[i] = ten;
                }
            }
        }

        tgt.min = dbmin;
        tgt.max = dbmax;
    }

    protected static toPercents(len: number, src: FrequencyResponseComponent, tgt: FrequencyResponseComponent) {
        const ten = 10;
        const hundreed = 100;
        const twohundred = 200;
        const pctmin = 0;
        let pctmax = -Infinity;
        let base = src.data[0];
        if (base < Number.EPSILON) {
            base = src.max;
        }

        for (let i = 0; i < len; ++i) {
            const pct = hundreed * src.data[i] / base;
            pctmax = Math.max(pctmax, pct);
            tgt.data[i] = pct;
        }

        // If pctmax falls into one of [100, 110), [110, 120), ...,  [190, 200)
        // intervals, set it to the maximum value of the interval.
        for (let i = 0; i < ten; ++i) {
            const min = hundreed + i * ten;
            const max = hundreed + (i + 1) * ten;
            if (pctmax >= min && pctmax < max) {
                pctmax = max;
                break;
            }
        }

        // Limit all maximal percentages values to 200.
        if (pctmax > twohundred) {
            pctmax = twohundred;
            for (let i = 0; i < len; ++i) {
                if (tgt.data[i] > twohundred) {
                    tgt.data[i] = twohundred;
                }
            }
        }

        tgt.min = pctmin;
        tgt.max = pctmax;
    }

    /**
     * Performs a direct real fast Fourier transform.
     *
     * The input parameter is a data array containing real data on input and {re,im} pairs on return.
     *
     * The length of the input data slice must be a power of 2 (128, 256, 512, 1024, 2048, 4096).
     * Since this is an internal function, we don't check the validity of the length here.
     * */
    protected static directRealFastFourierTransform(array: number[]) {
        const half = 0.5;
        const two = 2;
        const twoPi = two * Math.PI;
        const four = 4;
        const len = array.length;
        const ttheta = twoPi / len;
        const nn = len / 2;

        let j = 1;
        for (let ii = 1; ii <= nn; ++ii) {
            const i = two * ii - 1;
            if (j > i) {
                const tempR = array[j - 1];
                const tempI = array[j];
                array[j - 1] = array[i - 1];
                array[j] = array[i];
                array[i - 1] = tempR;
                array[i] = tempI;
            }

            let m = nn;
            while (m >= 2 && j > m) {
                j -= m;
                m /= 2;
            }
            j += m;
        }

        let mMax = two;
        let n = len;
        while (n > mMax) {
            const istep = two * mMax;
            const theta = twoPi / mMax;
            const wpI = Math.sin(theta);
            let wpR = Math.sin(half * theta);
            wpR = -two * wpR * wpR;
            let wR = 1.0;
            let wI = 0.0;
            for (let ii = 1; ii <= mMax / two; ++ii) {
                const m = two * ii - 1;
                for (let jj = 0; jj <= (n - m) / istep; ++jj) {
                    const i = m + jj * istep;
                    j = i + mMax;
                    const tempR = wR * array[j - 1] - wI * array[j];
                    const tempI = wR * array[j] + wI * array[j - 1];
                    array[j - 1] = array[i - 1] - tempR;
                    array[j] = array[i] - tempI;
                    array[i - 1] = array[i - 1] + tempR;
                    array[i] = array[i] + tempI;
                }
                const wtemp = wR;
                wR = wR * wpR - wI * wpI + wR;
                wI = wI * wpR + wtemp * wpI + wI;
            }
            mMax = istep;
        }

        const twpI = Math.sin(ttheta);
        let twpR = Math.sin(half * ttheta);
        twpR = -two * twpR * twpR;
        let twR = 1 + twpR;
        let twI = twpI;
        n = len / four + 1;
        for (let i = 2; i <= n; ++i) {
            const i1 = i + i - 2;
            const i2 = i1 + 1;
            const i3 = len + 1 - i2;
            const i4 = i3 + 1;
            const wRs = twR;
            const wIs = twI;
            const h1R = half * (array[i1] + array[i3]);
            const h1I = half * (array[i2] - array[i4]);
            const h2R = half * (array[i2] + array[i4]);
            const h2I = -half * (array[i1] - array[i3]);
            array[i1] = h1R + wRs * h2R - wIs * h2I;
            array[i2] = h1I + wRs * h2I + wIs * h2R;
            array[i3] = h1R - wRs * h2R + wIs * h2I;
            array[i4] = -h1I + wRs * h2I + wIs * h2R;
            const twTemp = twR;
            twR = twR * twpR - twI * twpI + twR;
            twI = twI * twpR + twTemp * twpI + twI;
        }

        twR = array[0];
        array[0] = twR + array[1];
        array[1] = twR - array[1];
    }
}
