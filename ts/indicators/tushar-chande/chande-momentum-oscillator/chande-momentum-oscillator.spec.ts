import { } from 'jasmine';

import { ChandeMomentumOscillator } from './chande-momentum-oscillator';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { ChandeMomentumOscillatorOutput } from './output';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { bookLength10Input, bookLength10Output, input } from './testdata';

/* eslint-disable max-len */
// Input data from Chande's book and TA-Lib test data.
// TA Lib uses incorrect CMO calculation which is incompatible with Chande's book.
// We use Chande's book data for the length=10 test.

describe('ChandeMomentumOscillator', () => {

  it('should have correct output enum value', () => {
    expect(ChandeMomentumOscillatorOutput.ChandeMomentumOscillatorValue).toBe(0);
  });

  it('should return expected mnemonic for default components', () => {
    const cmo = new ChandeMomentumOscillator({length: 14});
    expect(cmo.metadata().mnemonic).toBe('cmo(14)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const cmo = new ChandeMomentumOscillator({length: 14, barComponent: BarComponent.Median});
    expect(cmo.metadata().mnemonic).toBe('cmo(14, hl/2)');
    expect(cmo.metadata().description).toBe('Chande Momentum Oscillator cmo(14, hl/2)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const cmo = new ChandeMomentumOscillator({length: 14, quoteComponent: QuoteComponent.Bid});
    expect(cmo.metadata().mnemonic).toBe('cmo(14, b)');
    expect(cmo.metadata().description).toBe('Chande Momentum Oscillator cmo(14, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const cmo = new ChandeMomentumOscillator({length: 14, tradeComponent: TradeComponent.Volume});
    expect(cmo.metadata().mnemonic).toBe('cmo(14, v)');
    expect(cmo.metadata().description).toBe('Chande Momentum Oscillator cmo(14, v)');
  });

  it('should return expected metadata', () => {
    const cmo = new ChandeMomentumOscillator({length: 5});
    const meta = cmo.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.ChandeMomentumOscillator);
    expect(meta.mnemonic).toBe('cmo(5)');
    expect(meta.description).toBe('Chande Momentum Oscillator cmo(5)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(ChandeMomentumOscillatorOutput.ChandeMomentumOscillatorValue);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('cmo(5)');
    expect(meta.outputs[0].description).toBe('Chande Momentum Oscillator cmo(5)');
  });

  it('should throw if length is less than 1', () => {
    expect(() => { new ChandeMomentumOscillator({length: 0}); }).toThrow();
    expect(() => { new ChandeMomentumOscillator({length: -1}); }).toThrow();
  });

  it('should calculate expected output for length = 10 (book)', () => {
    const cmo = new ChandeMomentumOscillator({length: 10});
    const eps = 1e-13;

    for (let i = 0; i < 10; i++) {
      expect(cmo.update(bookLength10Input[i])).toBeNaN();
    }

    for (let i = 10; i < bookLength10Input.length; i++) {
      const act = cmo.update(bookLength10Input[i]);
      expect(Math.abs(act - bookLength10Output[i])).toBeLessThan(eps);
    }

    expect(cmo.update(Number.NaN)).toBeNaN();
  });

  it('should track primed state correctly for length = 1', () => {
    const cmo = new ChandeMomentumOscillator({length: 1});
    expect(cmo.isPrimed()).toBe(false);

    cmo.update(input[0]);
    expect(cmo.isPrimed()).toBe(false);

    cmo.update(input[1]);
    expect(cmo.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 2', () => {
    const cmo = new ChandeMomentumOscillator({length: 2});
    expect(cmo.isPrimed()).toBe(false);

    for (let i = 0; i < 2; i++) {
      cmo.update(input[i]);
      expect(cmo.isPrimed()).toBe(false);
    }

    cmo.update(input[2]);
    expect(cmo.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 5', () => {
    const cmo = new ChandeMomentumOscillator({length: 5});
    expect(cmo.isPrimed()).toBe(false);

    for (let i = 0; i < 5; i++) {
      cmo.update(input[i]);
      expect(cmo.isPrimed()).toBe(false);
    }

    cmo.update(input[5]);
    expect(cmo.isPrimed()).toBe(true);
  });

  it('should track primed state correctly for length = 10', () => {
    const cmo = new ChandeMomentumOscillator({length: 10});
    expect(cmo.isPrimed()).toBe(false);

    for (let i = 0; i < 10; i++) {
      cmo.update(input[i]);
      expect(cmo.isPrimed()).toBe(false);
    }

    cmo.update(input[10]);
    expect(cmo.isPrimed()).toBe(true);
  });
});
