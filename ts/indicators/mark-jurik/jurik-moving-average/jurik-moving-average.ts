import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { LineIndicator } from '../../core/line-indicator';
import { OutputType } from '../../core/outputs/output-type';
import { JurikMovingAverageOutput } from './jurik-moving-average-output';
import { JurikMovingAverageParams } from './jurik-moving-average-params';

/** Function to calculate mnemonic of a __JurikMovingAverage__ indicator. */
export const jurikMovingAverageMnemonic = (params: JurikMovingAverageParams): string =>
  'jma('.concat(
    params.length.toString(),
    ', ',
    params.phase.toString(),
    componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent),
    ')');

const c30 = 30
const c64 = 64
const c128 = 128;
const cInit = 1000000.0;
const cEpsilon = 1.0E-10;

/** Jurik Moving Average line indicator. */
export class JurikMovingAverage extends LineIndicator {
  private list: Array<number> = new Array(c128);//.fill(0);
  private ring: Array<number> = new Array(c128).fill(0);
  private ring2: Array<number> = new Array(11).fill(0);
  private buffer: Array<number> = new Array(62).fill(0);

  // Integers
  private s28 = c64-1;
	private s30 = c64;
	private s38 = 0;
	private s40 = 0;
	private s48 = 0;
	private s50 = 0;
	private s70 = 0;
	private f0 = 1;
	private fD8 = 0;
	private fF0 = 0;
	private v5 = 0;

  // Doubles
	private s8 = 0.0;
	private s18 = 0.0;
	private f10 = 0.0;
	private f18 = 0.0;
	private f38 = 0.0;
	private f50 = 0.0;
	private f58 = 0.0;
	private f78 = 0.0;
	private f88 = 0.0;
	private f90 = 0.0;
	private f98 = 0.0;
	private fA8 = 0.0;
	private fB8 = 0.0;
	private fC0 = 0.0;
	private fC8 = 0.0;
	private fF8 = 0.0;
	private v1 = 0.0;
	private v2 = 0.0;
	private v3 = 0.0;

  /**
   * Constructs an instance using given parameters.
   **/
  public constructor(params: JurikMovingAverageParams) {
    super();
    const length = Math.floor(params.length);
    if (length < 1) {
      throw new Error('length should be positive');
    }

    const phase = params.phase;
    if (phase < -100 || phase > 100) {
      throw new Error('phase should be in range [-100, 100]');
    }

    this.mnemonic = jurikMovingAverageMnemonic(params);
    this.description = "Jurik moving average " + this.mnemonic;
    this.primed = false;
    this.barComponent = params.barComponent;
    this.quoteComponent = params.quoteComponent;
    this.tradeComponent = params.tradeComponent;
    for (let i = 0; i < c64; i++) {
      this.list[i] = -cInit;
    }
    for (let i = c64; i < c128; i++) {
      this.list[i] = cInit;
    }

    let f80 = length > 1 ? (length - 1) / 2 : cEpsilon;
 
    this.f10 = phase/100 + 1.5;  
    this.v1 = Math.log(Math.sqrt(f80));
    this.v2 = this.v1;
    this.v3 = Math.max(this.v2 / Math.log(2) + 2, 0)
    this.f98 = this.v3;
    this.f88 = Math.max(this.f98 - 2, 0.5);    
    this.f78 = Math.sqrt(f80) * this.f98;
    this.f90 = this.f78 / (this.f78 + 1);
    f80 *= 0.9;
    this.f50 = f80 / (f80 + 2);
  }

  /** Describes the output data of the indicator. */
  public metadata(): IndicatorMetadata {
    return {
      type: IndicatorType.JurikMovingAverage,
      mnemonic: this.mnemonic,
      description: this.description,
      outputs: [{
        kind: JurikMovingAverageOutput.MovingAverageValue,
        type: OutputType.Scalar,
        mnemonic: this.mnemonic,
        description: this.description,
      }],
    };
  }

  /** Updates the value of the indicator given the next sample.
   *
   *  The indicator is not primed during the first 30 updates.
   */
  public update(sample: number): number {
    if (Number.isNaN(sample)) {
      return sample;
    }
    
	  if (this.fF0 < 61) {
		  this.fF0++;
		  this.buffer[this.fF0] = sample;
	  }

    if (this.fF0 <= c30) {
      return Number.NaN;
    }

  	this.primed = true;
	  if (this.f0 === 0) {
		  this.fD8 = 0;
	  } else {
		  this.f0 = 0;
		  this.v5 = 0;

		  for (let i = 1; i < c30; i++) {
			  if (this.buffer[i+1] !== this.buffer[i]) {
				  this.v5 = 1;
			  }
		  }

		  this.fD8 = this.v5 * c30
      this.f38 = this.fD8 === 0 ? sample : this.buffer[1];
      this.f18 = this.f38
		  if (this.fD8 > 29) {
			  this.fD8 = 29;
		  }
	  }

    for (let i = this.fD8; i >= 0; i--) {
      const f8 = i !== 0 ? this.buffer[31-i] : sample;
      const f28 = f8 - this.f18;
      const f48 = f8 - this.f38;
      const a28 = Math.abs(f28);
      const a48 = Math.abs(f48);
      this.v2 = Math.max(a28, a48);
      const fA0 = this.v2;
		  const v = fA0 + cEpsilon;

      if (this.s48 <= 1) {
        this.s48 = 127;
      } else {
        this.s48--;
      }
  
      if (this.s50 <= 1) {
        this.s50 = 10;
      } else {
        this.s50--;
      }
  
      if (this.s70 < c128) {
        this.s70++;
      }

      this.s8 += v - this.ring2[this.s50];
		  this.ring2[this.s50] = v;
		  const s20 = this.s70 > 10 ? this.s8 / 10 : this.s8 / this.s70;

		  let s58, s68: number;
      if (this.s70 > 127) {
        const s10 = this.ring[this.s48];
        this.ring[this.s48] = s20;
        s68 = c64;
        s58 = c64;
  
        while (s68 > 1) {
          if (this.list[s58] < s10) {
            s68 /= 2;
            s58 += s68;
          } else if (this.list[s58] <= s10) {
            s68 = 1;
          } else {
            s68 /= 2;
            s58 -= s68;
          }
        }
      } else {
        this.ring[this.s48] = s20;
        if (this.s28 + this.s30 > 127) {
          this.s30--;
          s58 = this.s30;
        } else {
          this.s28++;
          s58 = this.s28;
        }
  
        this.s38 = Math.min(this.s28, 96);
        this.s40 = Math.max(this.s30, 32);
      }

      s68 = c64;
		  let s60 = c64;
		  while (s68 > 1) {
			  if (this.list[s60] >= s20) {
				  if (this.list[s60-1] <= s20) {
					  s68 = 1;
				  } else {
					  s68 /= 2;
					  s60 -= s68;
				  }
			  } else {
				  s68 /= 2;
				  s60 += s68;
			  }

			  if (s60 === 127 && s20 > this.list[127]) {
				  s60 = c128;
			  }
		  }

      if (this.s70 > 127) {
        if (s58 >= s60) {
          if (this.s38+1 > s60 && this.s40-1 < s60) {
            this.s18 += s20;
          } else if (this.s40 > s60 && this.s40-1 < s58) {
            this.s18 += this.list[this.s40-1];
          }
        } else if (this.s40 >= s60) {
          if (this.s38+1 < s60 && this.s38+1 > s58) {
            this.s18 += this.list[this.s38+1];
          }
        } else if (this.s38+2 > s60) {
          this.s18 += s20;
        } else if (this.s38+1 < s60 && this.s38+1 > s58) {
          this.s18 += this.list[this.s38+1];
        }
  
        if (s58 > s60) {
          if (this.s40-1 < s58 && this.s38+1 > s58) {
            this.s18 -= this.list[s58];
          } else if (this.s38 < s58 && this.s38+1 > s60) {
            this.s18 -= this.list[this.s38];
          }
        } else {
          if (this.s38+1 > s58 && this.s40-1 < s58) {
            this.s18 -= this.list[s58];
          } else if (this.s40 > s58 && this.s40 < s60) {
            this.s18 -= this.list[this.s40];
          }
        }
      }
  
      if (s58 <= s60) {
        if (s58 >= s60) {
          this.list[s60] = s20;
        } else {
          for (let k = s58 + 1; k <= s60-1; k++) {
            this.list[k-1] = this.list[k];
          }
  
          this.list[s60-1] = s20;
        }
      } else {
        for (let k = s58 - 1; k >= s60; k--) {
          this.list[k+1] = this.list[k];
        }
  
        this.list[s60] = s20;
      }
  
      if (this.s70 < c128) {
        this.s18 = 0;
        for (let k = this.s40; k <= this.s38; k++) {
          this.s18 += this.list[k];
        }
      }
  
      const f60 = this.s18 / (this.s38 - this.s40 + 1);
      if (this.fF8+1 > 31) {
        this.fF8 = 31;
      } else {
        this.fF8++;
      }
  
      if (this.fF8 <= c30) {
        if (f28 > 0) {
          this.f18 = f8;
        } else {
          this.f18 = f8 - f28 * this.f90;
        }
  
        if (f48 < 0) {
          this.f38 = f8;
        } else {
          this.f38 = f8 - f48 * this.f90;
        }
  
        this.fB8 = sample;
        if (this.fF8 !== c30) {
          continue
        }
  
        this.fC0 = sample;
        let v4 = 1;
  
        if (Math.ceil(this.f78) >= 1) {
          v4 = Math.ceil(this.f78);
        }
  
        const fE8 = v4; // Math.floor(v4);
        let v2 = 1;
  
        if (Math.floor(this.f78) >= 1) {
          v2 = Math.floor(this.f78);
        }
  
        const fE0 = v2; // Math.floor(v2);
        let f68 = 1;
  
        if (fE8 !== fE0) {
          v4 = fE8 - fE0;
          f68 = (this.f78 - fE0) / v4;
        }
  
        const v5 = Math.min(fE0, 29);
        const v6 = Math.min(fE8, 29);
        this.fA8 = (sample - this.buffer[this.fF0-v5]) * (1-f68) / fE0 +
          (sample-this.buffer[this.fF0-v6]) * f68 / fE8;
      } else {
        const p = Math.pow(fA0 / f60, this.f88);
        this.v1 = Math.min(this.f98, p);
  
        if (this.v1 < 1) {
          this.v2 = 1;
        } else {
  				this.v3 = Math.min(this.f98, p);
          this.v2 = this.v3;
        }
  
        this.f58 = this.v2
        const f70 = Math.pow(this.f90, Math.sqrt(this.f58));
  
        if (f28 > 0) {
          this.f18 = f8;
        } else {
          this.f18 = f8 - f28 * f70;
        }
  
        if (f48 < 0) {
          this.f38 = f8;
        } else {
          this.f38 = f8 - f48 * f70;
        }
      }  
    }

    if (this.fF8 > c30) {
      const f30 = Math.pow(this.f50, this.f58);
      this.fC0 = (1 - f30) * sample + f30 * this.fC0;
      this.fC8 = (sample - this.fC0) * (1 - this.f50) + this.f50 * this.fC8;
      const fD0 = this.f10 * this.fC8 + this.fC0;
      const f20 = f30 * -2;
      const f40 = f30 * f30;
      const fB0 = f20 + f40 + 1;
      this.fA8 = (fD0 - this.fB8) * fB0 + f40 * this.fA8;
      this.fB8 += this.fA8;
    }
  
    return this.fB8;
  }
}
