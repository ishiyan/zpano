import { } from 'jasmine';

import { T3ExponentialMovingAverage } from './t3-exponential-moving-average';
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';
import { T3ExponentialMovingAverageOutput } from './t3-exponential-moving-average-output';

/* eslint-disable max-len */
// Input data is taken from the TA-Lib (http://ta-lib.org/) tests,
//    test_data.c, TA_SREF_close_daily_ref_0_PRIV[252].
//
// Output data is taken from TA-Lib (http://ta-lib.org/) tests,
//    test_ma.c.
//
// /************/
// /*  T3 TEST */
// /************/
// { 1, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,      0,  85.73, 24,  252-24  }, /* First Value */
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS,      1,  84.37, 24,  252-24  },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 252-26, 109.03, 24,  252-24  },
// { 0, TA_ANY_MA_TEST, 0, 0, 251, 5, TA_MAType_T3, TA_COMPATIBILITY_DEFAULT, TA_SUCCESS, 252-25, 108.88, 24,  252-24  }, /* Last Value */

const input = [
  91.500000,94.815000,94.375000,95.095000,93.780000,94.625000,92.530000,92.750000,90.315000,92.470000,96.125000,
  97.250000,98.500000,89.875000,91.000000,92.815000,89.155000,89.345000,91.625000,89.875000,88.375000,87.625000,
  84.780000,83.000000,83.500000,81.375000,84.440000,89.250000,86.375000,86.250000,85.250000,87.125000,85.815000,
  88.970000,88.470000,86.875000,86.815000,84.875000,84.190000,83.875000,83.375000,85.500000,89.190000,89.440000,
  91.095000,90.750000,91.440000,89.000000,91.000000,90.500000,89.030000,88.815000,84.280000,83.500000,82.690000,
  84.750000,85.655000,86.190000,88.940000,89.280000,88.625000,88.500000,91.970000,91.500000,93.250000,93.500000,
  93.155000,91.720000,90.000000,89.690000,88.875000,85.190000,83.375000,84.875000,85.940000,97.250000,99.875000,
  104.940000,106.000000,102.500000,102.405000,104.595000,106.125000,106.000000,106.065000,104.625000,108.625000,
  109.315000,110.500000,112.750000,123.000000,119.625000,118.750000,119.250000,117.940000,116.440000,115.190000,
  111.875000,110.595000,118.125000,116.000000,116.000000,112.000000,113.750000,112.940000,116.000000,120.500000,
  116.620000,117.000000,115.250000,114.310000,115.500000,115.870000,120.690000,120.190000,120.750000,124.750000,
  123.370000,122.940000,122.560000,123.120000,122.560000,124.620000,129.250000,131.000000,132.250000,131.000000,
  132.810000,134.000000,137.380000,137.810000,137.880000,137.250000,136.310000,136.250000,134.630000,128.250000,
  129.000000,123.870000,124.810000,123.000000,126.250000,128.380000,125.370000,125.690000,122.250000,119.370000,
  118.500000,123.190000,123.500000,122.190000,119.310000,123.310000,121.120000,123.370000,127.370000,128.500000,
  123.870000,122.940000,121.750000,124.440000,122.000000,122.370000,122.940000,124.000000,123.190000,124.560000,
  127.250000,125.870000,128.860000,132.000000,130.750000,134.750000,135.000000,132.380000,133.310000,131.940000,
  130.000000,125.370000,130.130000,127.120000,125.190000,122.000000,125.000000,123.000000,123.500000,120.060000,
  121.000000,117.750000,119.870000,122.000000,119.190000,116.370000,113.500000,114.250000,110.000000,105.060000,
  107.000000,107.870000,107.000000,107.120000,107.000000,91.000000,93.940000,93.870000,95.500000,93.000000,
  94.940000,98.250000,96.750000,94.810000,94.370000,91.560000,90.250000,93.940000,93.620000,97.000000,95.000000,
  95.870000,94.060000,94.620000,93.750000,98.000000,103.940000,107.870000,106.060000,104.500000,105.000000,
  104.190000,103.060000,103.420000,105.270000,111.870000,116.000000,116.620000,118.280000,113.370000,109.000000,
  109.700000,109.250000,107.000000,109.190000,110.000000,109.200000,110.120000,108.000000,108.620000,109.750000,
  109.810000,109.000000,108.750000,107.870000
];

/** Taken from TA-Lib (http://ta-lib.org/) tests, test_T3.xls, T3(5,0.7) column.
 * Length is 5, volume factor is 0.7, firstIsAverage = true.
 */
const expected5sma = [
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, 85.7298756737179, 84.3653623895779, 83.4821283349961, 83.6487802207527,
  84.2077478000361, 84.8140022624377, 85.2116245986709, 85.6197795358472, 85.8836656465792, 86.3750719606217, 86.9636060637729,
  87.3371558479710, 87.4758153832031, 87.2270424677263, 86.6613834388821, 85.9398842053554, 85.1714918011768, 84.7244766890823,
  85.0186552094379, 85.8361275749773, 87.0356306436977, 88.2725752638162, 89.4175897806162, 90.0755914533444, 90.5278838453218,
  90.7984930563717, 90.7581699961507, 90.4801259590211, 89.5672188532377, 88.1996129474603, 86.6401290448746, 85.4193255238329,
  84.7373428697986, 84.5542144141793, 85.0089060379763, 85.8702240344147, 86.7701665475090, 87.5189361511518, 88.4703458181180,
  89.4547401883457, 90.5257440890420, 91.5716941141709, 92.4261488267763, 92.8706801637941, 92.7736235068667, 92.2939046615678,
  91.5631031818609, 90.3470608503046, 88.7081510949477, 87.1739940007027, 86.0889664824729, 86.7772147701018, 89.0169092028569,
  92.5057804552432, 96.4466124562832, 99.6837411808250, 101.9235391590250, 103.5186452020380, 104.7879113375970,
  105.7542427169020, 106.4292320356980, 106.6733152181040, 107.0568493010970, 107.6555206656350, 108.4583182470370,
  109.5471293301630, 111.8943473068200, 114.5257440534960, 116.7522861312560, 118.4172664729190, 119.3724288777240,
  119.5812174651990, 119.1501687372510, 117.9971201089140, 116.3534343269300, 115.5013310938270, 115.1990014959530,
  115.2083285297160, 114.8904585528640, 114.4950593565460, 114.0487358120780, 113.9829965309400, 114.7612063113720,
  115.5988427739080, 116.2820558532300, 116.5569339013220, 116.3998304939940, 116.1332216622030, 115.9307103028120,
  116.3816663728700, 117.2275795032460, 118.2363153524880, 119.6585225029040, 121.0618935007380, 122.1700155983110,
  122.8820887081840, 123.3207551399340, 123.4984543600820, 123.7352650715480, 124.5754479479500, 125.9877593956320,
  127.7267734776620, 129.2785277930630, 130.6608444538510, 131.9285009815100, 133.3876328585210, 134.9030494029410,
  136.2618952710580, 137.2563561145750, 137.7646531404110, 137.8839771260830, 137.5729479559510, 136.2709748699910,
  134.4840999190830, 132.0833416334720, 129.6490588115330, 127.3492061640200, 125.7880621091180, 125.2052512102680,
  124.9747151996480, 124.9118778165660, 124.5323645295170, 123.5869115064700, 122.2636241397390, 121.4460557356730,
  121.2290496842930, 121.2683301268390, 121.0648053749090, 121.1193616899350, 121.1687367276690, 121.4290677661220,
  122.3015023505060, 123.6599418543550, 124.6036952964740, 124.9216747581580, 124.6510883818380, 124.3746003415850,
  123.9445858612620, 123.4799723118630, 123.1305751279310, 123.0381466442190, 123.0414739643860, 123.2322556687330,
  123.8479012432350, 124.5374199390880, 125.4819317489600, 126.8800108924130, 128.2915644516780, 129.9430787439590,
  131.6306461797650, 132.8096655685730, 133.5564467005160, 133.8091349137510, 133.4892448608330, 132.2930588908100,
  131.1872334543740, 130.0750416897400, 128.8276548825960, 127.2323656922100, 125.9087869921910, 124.7744630847380,
  123.9218599303110, 122.9512269688430, 122.0557031673420, 120.9658538086510, 120.0812010671890, 119.7469811721990,
  119.5244049684770, 118.9898699494460, 117.9192093492870, 116.6944322236370, 115.0900468905000, 112.8160424290250,
  110.5623150888320, 108.8011743021960, 107.5209706801750, 106.6905259237250, 106.2133261998150, 104.1264662149270,
  101.2847821472110, 98.4590980536630, 96.2919654244885, 94.5940438966966, 93.5667878417156, 93.4953907344503, 93.9033603418944,
  94.2582984032293, 94.4092295943194, 94.0788845752426, 93.3002023012997, 92.7807005779447, 92.5740879570196, 92.9881068881615,
  93.5608091110847, 94.1882964748360, 94.5557215309443, 94.7372650538305, 94.7016291992604, 95.0301498197575, 96.3450164034449,
  98.6871801925784, 101.1917630244060, 103.1803496718400, 104.5740156548680, 105.3446911849580, 105.5108202516040,
  105.3291172048920, 105.2095106016370, 105.9763724337760, 107.8213878123870, 110.2176992295730, 112.7781119527130,
  114.5107526150470, 114.8553909679630, 114.2944059224460, 113.2731825801640, 111.8905200107540, 110.7001873898750,
  109.9420061462590, 109.4704380663640, 109.2996995046210, 109.0836408007100, 108.8729425638770, 108.8321814660130,
  108.9382930540510, 109.0247660433720, 109.0321034127580, 108.8791500044930
];

describe('T3ExponentialMovingAverage', () => {

  it('should have correct output enum value', () => {
    expect(T3ExponentialMovingAverageOutput.T3ExponentialMovingAverageValue).toBe(0);
  });

  it('should return expected mnemonic for length-based', () => {
    let t3 = new T3ExponentialMovingAverage({length: 7, volumeFactor: 0.6781, firstIsAverage: true});
    expect(t3.metadata().mnemonic).toBe('t3(7, 0.67810000)');
    t3 = new T3ExponentialMovingAverage({length: 7, volumeFactor: 0.6789, firstIsAverage: false});
    expect(t3.metadata().mnemonic).toBe('t3(7, 0.67890000)');
  });

  it('should return expected mnemonic for smoothing-factor-based', () => {
    // alpha = 0.12345, length = round(2/0.12345) - 1 = round(16.2) - 1 = 15
    const t3 = new T3ExponentialMovingAverage({smoothingFactor: 0.12345, volumeFactor: 0.56789, firstIsAverage: false});
    expect(t3.metadata().mnemonic).toBe('t3(15, 0.12345000, 0.56789000)');
  });

  it('should return expected metadata', () => {
    const t3 = new T3ExponentialMovingAverage({length: 10, volumeFactor: 0.3333, firstIsAverage: true});
    const meta = t3.metadata();

    expect(meta.type).toBe(IndicatorType.T3ExponentialMovingAverage);
    expect(meta.mnemonic).toBe('t3(10, 0.33330000)');
    expect(meta.description).toBe('T3 exponential moving average t3(10, 0.33330000)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(T3ExponentialMovingAverageOutput.T3ExponentialMovingAverageValue);
    expect(meta.outputs[0].type).toBe(OutputType.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('t3(10, 0.33330000)');
    expect(meta.outputs[0].description).toBe('T3 exponential moving average t3(10, 0.33330000)');
  });

  it('should return expected metadata for smoothing-factor-based', () => {
    // alpha = 2 / (10 + 1) = 2/11 = 0.18181818...
    const alpha = 2 / 11;
    const t3 = new T3ExponentialMovingAverage({smoothingFactor: alpha, volumeFactor: 0.3333333, firstIsAverage: false});
    const meta = t3.metadata();

    expect(meta.type).toBe(IndicatorType.T3ExponentialMovingAverage);
    expect(meta.mnemonic).toBe('t3(10, 0.18181818, 0.33333330)');
    expect(meta.description).toBe('T3 exponential moving average t3(10, 0.18181818, 0.33333330)');
    expect(meta.outputs.length).toBe(1);
    expect(meta.outputs[0].kind).toBe(T3ExponentialMovingAverageOutput.T3ExponentialMovingAverageValue);
    expect(meta.outputs[0].type).toBe(OutputType.Scalar);
    expect(meta.outputs[0].mnemonic).toBe('t3(10, 0.18181818, 0.33333330)');
    expect(meta.outputs[0].description).toBe('T3 exponential moving average t3(10, 0.18181818, 0.33333330)');
  });

  it('should throw if length is less than 2', () => {
    expect(() => { new T3ExponentialMovingAverage({length: 1, volumeFactor: 0.7, firstIsAverage: true}); }).toThrow();
  });

  it('should throw if smoothing factor is less than 0', () => {
    expect(() => { new T3ExponentialMovingAverage({smoothingFactor: -0.1, volumeFactor: 0.7, firstIsAverage: false}); }).toThrow();
  });

  it('should throw if smoothing factor is greater than 1', () => {
    expect(() => { new T3ExponentialMovingAverage({smoothingFactor: 1.1, volumeFactor: 0.7, firstIsAverage: false}); }).toThrow();
  });

  it('should throw if volume factor is less than 0', () => {
    expect(() => { new T3ExponentialMovingAverage({length: 5, volumeFactor: -0.1, firstIsAverage: true}); }).toThrow();
  });

  it('should throw if volume factor is greater than 1', () => {
    expect(() => { new T3ExponentialMovingAverage({length: 5, volumeFactor: 1.1, firstIsAverage: true}); }).toThrow();
  });

  it('should calculate expected output and prime state for length 5, first is NOT SMA', () => {
    const len = 5;
    const lenPrimed = 6*(len - 1);
    const epsilon = 1e-3;
    const t3 = new T3ExponentialMovingAverage({length: len, volumeFactor: 0.7, firstIsAverage: false});

    for (let i = 0; i < lenPrimed; i++) {
      expect(t3.update(input[i])).toBeNaN();
      expect(t3.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = t3.update(input[i]);
      expect(t3.isPrimed()).toBe(true);

      if (i === 24) {
        expect(act).toBeCloseTo(85.749, epsilon);
      } else if (i === 25) {
        expect(act).toBeCloseTo(84.380, epsilon);
      } else if (i === 250) {
        expect(act).toBeCloseTo(109.03, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(108.88, epsilon);
      }
    }

    expect(t3.update(Number.NaN)).toBeNaN();
  });

  it('should calculate expected output and prime state for length 5, first is SMA', () => {
    const len = 5;
    const lenPrimed = 6*(len - 1);
    const epsilon = 1e-3;
    const t3 = new T3ExponentialMovingAverage({length: len, volumeFactor: 0.7, firstIsAverage: true});

    for (let i = 0; i < lenPrimed; i++) {
      expect(t3.update(input[i])).toBeNaN();
      expect(t3.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = t3.update(input[i]);
      expect(t3.isPrimed()).toBe(true);

      if (i === 250) {
        expect(act).toBeCloseTo(109.03, epsilon);
      } else if (i === 251) {
        expect(act).toBeCloseTo(108.88, epsilon);
      }
    }

    expect(t3.update(Number.NaN)).toBeNaN();
  });

  it('should match expected output (Excel) for length 5, first is SMA', () => {
    const eps = 1e-13;
    const len = 5;
    const lenPrimed = 6*(len - 1);
    const t3 = new T3ExponentialMovingAverage({length: len, volumeFactor: 0.7, firstIsAverage: true});

    for (let i = 0; i < lenPrimed; i++) {
      expect(t3.update(input[i])).toBeNaN();
      expect(t3.isPrimed()).toBe(false);
    }

    for (let i = lenPrimed; i < input.length; i++) {
      const act = t3.update(input[i]);
      expect(t3.isPrimed()).toBe(true);
      expect(act).toBeCloseTo(expected5sma[i], eps);
    }

    expect(t3.update(Number.NaN)).toBeNaN();
  });
});
