import { } from 'jasmine';

import { LinearRegression } from './linear-regression';
import { LinearRegressionOutput } from './output';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Shape } from '../../core/outputs/shape/shape';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { Scalar } from '../../../entities/scalar';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Trade } from '../../../entities/trade';
import {
    input,
    expectedValue,
    expectedForecast,
    expectedIntercept,
    expectedSlopeRad,
    expectedSlopeDeg,
} from './testdata';

/* eslint-disable max-len */
// Input data from Excel verification (period 14), 252 entries.
// Expected Value output from Excel verification (period 14), 252 entries.
// Expected Forecast (TSF) output from Excel verification (period 14), 252 entries.
// Expected Intercept output from Excel verification (period 14), 252 entries.
// Expected SlopeRad output from Excel verification (period 14), 252 entries.
// Expected SlopeDeg output from Excel verification (period 14), 252 entries.
describe('LinearRegression', () => {
  const time = new Date(2021, 3, 1);
  const period = 14;
  const eps = 1e-4;

  it('should have correct output enum values', () => {
    expect(LinearRegressionOutput.Value).toBe(0);
    expect(LinearRegressionOutput.Forecast).toBe(1);
    expect(LinearRegressionOutput.Intercept).toBe(2);
    expect(LinearRegressionOutput.SlopeRad).toBe(3);
    expect(LinearRegressionOutput.SlopeDeg).toBe(4);
  });

  it('should return expected mnemonic with default parameters', () => {
    const lr = new LinearRegression({ length: 14 });
    expect(lr.metadata().mnemonic).toBe('linreg(14)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const lr = new LinearRegression({ length: 14, barComponent: BarComponent.Open });
    expect(lr.metadata().mnemonic).toBe('linreg(14, o)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const lr = new LinearRegression({ length: 14, quoteComponent: QuoteComponent.Bid });
    expect(lr.metadata().mnemonic).toBe('linreg(14, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const lr = new LinearRegression({ length: 14, tradeComponent: TradeComponent.Volume });
    expect(lr.metadata().mnemonic).toBe('linreg(14, v)');
  });

  it('should return expected metadata', () => {
    const lr = new LinearRegression({ length: 14 });
    const meta = lr.metadata();

    expect(meta.identifier).toBe(IndicatorIdentifier.LinearRegression);
    expect(meta.mnemonic).toBe('linreg(14)');
    expect(meta.description).toBe('Linear Regression linreg(14)');
    expect(meta.outputs.length).toBe(5);

    expect(meta.outputs[0].kind).toBe(LinearRegressionOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].description).toBe('Linear Regression linreg(14) value');
    expect(meta.outputs[1].kind).toBe(LinearRegressionOutput.Forecast);
    expect(meta.outputs[1].description).toBe('Linear Regression linreg(14) forecast');
    expect(meta.outputs[2].kind).toBe(LinearRegressionOutput.Intercept);
    expect(meta.outputs[2].description).toBe('Linear Regression linreg(14) intercept');
    expect(meta.outputs[3].kind).toBe(LinearRegressionOutput.SlopeRad);
    expect(meta.outputs[3].description).toBe('Linear Regression linreg(14) slope');
    expect(meta.outputs[4].kind).toBe(LinearRegressionOutput.SlopeDeg);
    expect(meta.outputs[4].description).toBe('Linear Regression linreg(14) angle');
  });

  it('should throw if the length is less than 2', () => {
    expect(() => { new LinearRegression({ length: 1 }); }).toThrow();
    expect(() => { new LinearRegression({ length: 0 }); }).toThrow();
    expect(() => { new LinearRegression({ length: -1 }); }).toThrow();
  });

  it('should not throw for length = 2', () => {
    expect(() => { new LinearRegression({ length: 2 }); }).not.toThrow();
  });

  it('should calculate expected Value output for all 252 rows', () => {
    const lr = new LinearRegression({ length: period });

    for (let i = 0; i < 13; i++) {
      expect(lr.update(input[i])).toBeNaN();
      expect(lr.isPrimed()).toBe(false);
    }

    for (let i = 13; i < input.length; i++) {
      const value = lr.update(input[i]);
      expect(lr.isPrimed()).toBe(true);
      expect(Math.abs(value - expectedValue[i]))
        .withContext(`Value [${i}]`).toBeLessThan(eps);
    }

    expect(lr.update(Number.NaN)).toBeNaN();
  });

  it('should calculate all five outputs via updateScalar for all 252 rows', () => {
    const lr = new LinearRegression({ length: period });

    for (let i = 0; i < 13; i++) {
      const output = lr.updateScalar(new Scalar({ time, value: input[i] }));
      expect(output.length).toBe(5);
      for (let j = 0; j < 5; j++) {
        expect(Number.isNaN((output[j] as Scalar).value)).toBe(true);
      }
    }

    for (let i = 13; i < input.length; i++) {
      const output = lr.updateScalar(new Scalar({ time, value: input[i] }));
      expect(output.length).toBe(5);

      const sValue = output[0] as Scalar;
      const sForecast = output[1] as Scalar;
      const sIntercept = output[2] as Scalar;
      const sSlopeRad = output[3] as Scalar;
      const sSlopeDeg = output[4] as Scalar;

      expect(Math.abs(sValue.value - expectedValue[i]))
        .withContext(`Value [${i}]`).toBeLessThan(eps);
      expect(Math.abs(sForecast.value - expectedForecast[i]))
        .withContext(`Forecast [${i}]`).toBeLessThan(eps);
      expect(Math.abs(sIntercept.value - expectedIntercept[i]))
        .withContext(`Intercept [${i}]`).toBeLessThan(eps);
      expect(Math.abs(sSlopeRad.value - expectedSlopeRad[i]))
        .withContext(`SlopeRad [${i}]`).toBeLessThan(eps);
      expect(Math.abs(sSlopeDeg.value - expectedSlopeDeg[i]))
        .withContext(`SlopeDeg [${i}]`).toBeLessThan(eps);

      expect(sValue.time).toBe(time);
    }
  });

  it('should transition primed state correctly', () => {
    const lr = new LinearRegression({ length: period });

    expect(lr.isPrimed()).toBe(false);

    for (let i = 0; i < 13; i++) {
      lr.update(input[i]);
      expect(lr.isPrimed()).toBe(false);
    }

    lr.update(input[13]);
    expect(lr.isPrimed()).toBe(true);
  });

  it('should produce correct output for bar updates', () => {
    const lr = new LinearRegression({ length: period });

    // Prime with scalars, then check bar update.
    for (let i = 0; i < 14; i++) {
      lr.update(input[i]);
    }

    const bar = new Bar({ time, open: input[14], high: input[14] + 1, low: input[14] - 1, close: input[14], volume: 0 });
    const output = lr.updateBar(bar);
    expect(output.length).toBe(5);

    const s0 = output[0] as Scalar;
    expect(s0.time).toBe(time);
    // Expected Value at index 14.
    expect(Math.abs(s0.value - 93.26071428571430)).toBeLessThan(eps);
  });

  it('should produce correct output for quote updates', () => {
    const lr = new LinearRegression({ length: period });

    for (let i = 0; i < 14; i++) {
      lr.update(input[i]);
    }

    // DefaultQuoteComponent is MidPrice = (bid+ask)/2.
    const quote = new Quote({ time, bidPrice: input[14], askPrice: input[14], bidSize: 0, askSize: 0 });
    const output = lr.updateQuote(quote);
    expect(output.length).toBe(5);

    const s0 = output[0] as Scalar;
    expect(s0.time).toBe(time);
    expect(Math.abs(s0.value - 93.26071428571430)).toBeLessThan(eps);
  });

  it('should produce correct output for trade updates', () => {
    const lr = new LinearRegression({ length: period });

    for (let i = 0; i < 14; i++) {
      lr.update(input[i]);
    }

    const trade = new Trade({ time, price: input[14], volume: 0 });
    const output = lr.updateTrade(trade);
    expect(output.length).toBe(5);

    const s0 = output[0] as Scalar;
    expect(s0.time).toBe(time);
    expect(Math.abs(s0.value - 93.26071428571430)).toBeLessThan(eps);
  });

  it('should return NaN outputs when not primed', () => {
    const lr = new LinearRegression({ length: period });

    const s = new Scalar({ time, value: 100 });
    const output = lr.updateScalar(s);
    expect(output.length).toBe(5);

    for (let j = 0; j < 5; j++) {
      expect(Number.isNaN((output[j] as Scalar).value)).toBe(true);
    }
  });
});
