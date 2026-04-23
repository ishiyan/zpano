import { } from 'jasmine';

import { Bar } from '../../../entities/bar';
import { BarComponent } from '../../../entities/bar-component';
import { Quote } from '../../../entities/quote';
import { QuoteComponent } from '../../../entities/quote-component';
import { Scalar } from '../../../entities/scalar';
import { Trade } from '../../../entities/trade';
import { TradeComponent } from '../../../entities/trade-component';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { Band } from '../../core/outputs/band';
import { Shape } from '../../core/outputs/shape/shape';
import { HilbertTransformerCycleEstimatorType } from '../hilbert-transformer/cycle-estimator-type';
import { MesaAdaptiveMovingAverage } from './mesa-adaptive-moving-average';
import { MesaAdaptiveMovingAverageOutput } from './output';

/* eslint-disable max-len */
// Input data taken from TA-Lib (http://ta-lib.org/) tests, test_MAMA.xsl, Price, D5…D256, 252 entries.
// Expected data taken from TA-Lib (http://ta-lib.org/) tests, test_MAMA_new.xsl,
// MAMA, L5…L256, 252 entries,
// FAMA, M5…M256, 252 entries.
// All parameters have default values.

const input = [
  92.0000, 93.1725, 95.3125, 94.8450, 94.4075, 94.1100, 93.5000, 91.7350, 90.9550, 91.6875,
  94.5000, 97.9700, 97.5775, 90.7825, 89.0325, 92.0950, 91.1550, 89.7175, 90.6100, 91.0000,
  88.9225, 87.5150, 86.4375, 83.8900, 83.0025, 82.8125, 82.8450, 86.7350, 86.8600, 87.5475,
  85.7800, 86.1725, 86.4375, 87.2500, 88.9375, 88.2050, 85.8125, 84.5950, 83.6575, 84.4550,
  83.5000, 86.7825, 88.1725, 89.2650, 90.8600, 90.7825, 91.8600, 90.3600, 89.8600, 90.9225,
  89.5000, 87.6725, 86.5000, 84.2825, 82.9075, 84.2500, 85.6875, 86.6100, 88.2825, 89.5325,
  89.5000, 88.0950, 90.6250, 92.2350, 91.6725, 92.5925, 93.0150, 91.1725, 90.9850, 90.3775,
  88.2500, 86.9075, 84.0925, 83.1875, 84.2525, 97.8600, 99.8750, 103.2650, 105.9375, 103.5000,
  103.1100, 103.6100, 104.6400, 106.8150, 104.9525, 105.5000, 107.1400, 109.7350, 109.8450, 110.9850,
  120.0000, 119.8750, 117.9075, 119.4075, 117.9525, 117.2200, 115.6425, 113.1100, 111.7500, 114.5175,
  114.7450, 115.4700, 112.5300, 112.0300, 113.4350, 114.2200, 119.5950, 117.9650, 118.7150, 115.0300,
  114.5300, 115.0000, 116.5300, 120.1850, 120.5000, 120.5950, 124.1850, 125.3750, 122.9700, 123.0000,
  124.4350, 123.4400, 124.0300, 128.1850, 129.6550, 130.8750, 132.3450, 132.0650, 133.8150, 135.6600,
  137.0350, 137.4700, 137.3450, 136.3150, 136.4400, 136.2850, 129.0950, 128.3100, 126.0000, 124.0300,
  123.9350, 125.0300, 127.2500, 125.6200, 125.5300, 123.9050, 120.6550, 119.9650, 120.7800, 124.0000,
  122.7800, 120.7200, 121.7800, 122.4050, 123.2500, 126.1850, 127.5600, 126.5650, 123.0600, 122.7150,
  123.5900, 122.3100, 122.4650, 123.9650, 123.9700, 124.1550, 124.4350, 127.0000, 125.5000, 128.8750,
  130.5350, 132.3150, 134.0650, 136.0350, 133.7800, 132.7500, 133.4700, 130.9700, 127.5950, 128.4400,
  127.9400, 125.8100, 124.6250, 122.7200, 124.0900, 123.2200, 121.4050, 120.9350, 118.2800, 118.3750,
  121.1550, 120.9050, 117.1250, 113.0600, 114.9050, 112.4350, 107.9350, 105.9700, 106.3700, 106.8450,
  106.9700, 110.0300, 91.0000, 93.5600, 93.6200, 95.3100, 94.1850, 94.7800, 97.6250, 97.5900,
  95.2500, 94.7200, 92.2200, 91.5650, 92.2200, 93.8100, 95.5900, 96.1850, 94.6250, 95.1200,
  94.0000, 93.7450, 95.9050, 101.7450, 106.4400, 107.9350, 103.4050, 105.0600, 104.1550, 103.3100,
  103.3450, 104.8400, 110.4050, 114.5000, 117.3150, 118.2500, 117.1850, 109.7500, 109.6550, 108.5300,
  106.2200, 107.7200, 109.8400, 109.0950, 109.0900, 109.1550, 109.3150, 109.0600, 109.9050, 109.6250,
  109.5300, 108.0600,
];

const expectedMama = [
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, 82.8141250000000, 83.0101687500000, 83.2026603125000, 83.4199022968750,
  83.5379071820312, 83.6696368229297, 85.0535684114648, 85.1633899908916, 85.3520954913470,
  85.4947407167797, 85.5106286809407, 85.0528143404703, 84.9830486234468, 84.9566461922745,
  84.8838138826607, 85.8331569413304, 85.9501240942639, 86.1158678895507, 86.3530744950731,
  86.5745457703195, 89.2172728851597, 89.2744092409017, 89.3036887788567, 89.3846293399138,
  89.3903978729181, 89.3045029792722, 89.1642778303086, 86.7233889151543, 86.5325944693966,
  85.3912972346983, 85.4061073729634, 85.4663020043152, 85.6071119040994, 85.8033813088944,
  85.9882122434497, 86.0935516312772, 88.3592758156386, 88.5530620248567, 88.7090339236138,
  88.9032072274331, 89.1087968660615, 90.1406484330307, 90.1828660113792, 90.1925977108102,
  90.0954678252697, 89.9360694340062, 89.6438909623059, 86.4156954811530, 86.2446021188251,
  86.8253720128838, 87.4778534122396, 88.2672107416277, 89.1507252045463, 89.8681889443189,
  96.4890944721595, 96.8451397485515, 97.2348827611239, 97.7138886230677, 98.0758191919143,
  98.4470282323186, 101.8250440834490, 105.7800220417250, 107.8125110208620, 107.9711354698190,
  113.9855677349100, 116.9302838674550, 117.0475622035170, 117.1655590933420, 117.5590295466710,
  117.5404616523790, 117.4455635697600, 117.2287853912720, 116.9548461217090, 116.8329788156230,
  115.7889894078120, 115.7730399374210, 115.6108879405500, 115.4318435435220, 115.3320013663460,
  115.2764012980290, 117.4357006490150, 117.4621656165640, 118.0885828082820, 117.9356536678680,
  116.2328268339340, 116.1711854922370, 116.1891262176250, 116.3889199067440, 116.5944739114070,
  118.5947369557030, 118.8742501079180, 119.1992876025220, 121.0846438012610, 121.1804116111980,
  121.3431410306380, 122.3915705153190, 122.4734919895530, 122.7590673900750, 123.1038640205720,
  126.9894320102860, 129.6672160051430, 130.8661080025710, 131.0135526024430, 131.2458749723210,
  131.5353312237050, 131.8320646625190, 132.1077114293930, 134.2113557146970, 135.3256778573480,
  135.3736439644810, 135.0597117662570, 134.7222261779440, 134.2861148690470, 129.1580574345230,
  127.6079647084670, 127.4790664730440, 127.4676131493920, 127.3752324919220, 127.2829708673260,
  127.1140723239600, 126.7911187077620, 123.3780593538810, 123.2481563861870, 123.6240781930930,
  123.5818742834390, 123.4387805692670, 123.3558415408030, 123.3082994637630, 123.3053844905750,
  124.7451922452880, 124.8859326330230, 125.7254663165120, 124.3927331582560, 124.2939523455320,
  124.2587547282550, 124.1613169918420, 124.0765011422500, 124.0207505711250, 124.0182130425690,
  124.0250523904400, 124.0455497709180, 124.1932722823720, 124.8466361411860, 125.0480543341270,
  125.3224016174210, 128.8187008087100, 129.0810157682750, 129.4287149798610, 131.6043574899310,
  131.6616396154340, 131.7520576346620, 131.7129547529290, 131.5070570152830, 131.3537041645190,
  131.1830189562930, 128.4965094781460, 128.3029340042390, 128.0237873040270, 127.8270979388260,
  126.8972774956690, 124.1511387478340, 123.0563880258110, 122.6897993885170, 122.4740594190910,
  121.8145297095460, 121.3597648547730, 120.9105197717010, 119.9913167237650, 119.7370008875760,
  119.3719008431980, 118.4965356358960, 112.2332678179480, 111.9401044270510, 111.6853492056980,
  109.3276746028490, 109.3627908727070, 108.4446513290710, 107.7004187626180, 106.7428119832590,
  101.0264059916300, 97.6057029958149, 97.4644178460241, 97.4724469537229, 97.4783246060368,
  97.3669083757349, 97.2345629569482, 94.7272814784741, 94.5691674045504, 94.4517090343228,
  94.4196235826067, 95.0048117913034, 95.0638212017382, 95.0418801416513, 95.0457861345687,
  94.9934968278403, 94.3692484139201, 94.4460359932241, 94.8109841935629, 95.3924349838848,
  101.6637174919420, 101.7507816173450, 103.4053908086730, 103.4428712682390, 103.4362277048270,
  103.4316663195860, 103.5020830036060, 103.8472288534260, 109.1736144267130, 109.5806837053770,
  110.0141495201090, 113.5995747600540, 113.4070960220520, 113.2194912209490, 112.9850166599020,
  112.6467658269060, 112.4004275355610, 111.1202137677810, 111.0189530793920, 110.9225054254220,
  110.8341301541510, 110.7581736464430, 110.6732649641210, 110.2891324820610, 110.2559258579580,
  110.2196295650600, 110.1116480868070,
];

const expectedFama = [
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, 82.8125406250000, 82.8174813281250, 82.8271108027344, 82.8419305900879,
  82.8593300048865, 82.8795876753376, 83.4230828593694, 83.4665905376574, 83.5137281614997,
  83.5632534753817, 83.6119378555206, 83.9721569767581, 83.9974292679253, 84.0214096910340,
  84.0429697958247, 84.4905165822011, 84.5270067700027, 84.5667282979914, 84.6113869529184,
  84.6604659233535, 85.7996676638050, 85.8865362032324, 85.9719650176230, 86.0572816256803,
  86.1406095318612, 86.2197068680465, 86.2933211421031, 86.4008380853659, 86.4041319949666,
  86.1509233048996, 86.1323029066012, 86.1156528840440, 86.1029393595454, 86.0954504082791,
  86.0927694541584, 86.0927890085864, 86.6594107103494, 86.7067519932121, 86.7568090414721,
  86.8104689961212, 86.8679271928697, 87.6861075029099, 87.7485264656217, 87.8096282467514,
  87.8667742362143, 87.9185066161591, 87.9616412248128, 87.5751547888978, 87.5225361670348,
  87.5051070631811, 87.5044257219075, 87.5234953474005, 87.5641760938292, 87.6217764150914,
  89.8386059293584, 90.0137692748382, 90.1942971119954, 90.3822868997722, 90.5746252070757,
  90.7714352827068, 92.9191050749822, 96.1343343166678, 99.0538784927164, 99.2768099171440,
  102.9539993715850, 106.4480704955530, 107.0841072479250, 107.3361435440610, 109.8918650447130,
  110.1013133578250, 110.2849196131230, 110.4585162575770, 110.6209245041800, 110.7762258619660,
  112.0294167484270, 112.1230073281520, 112.2102043434620, 112.2907453234640, 112.3667767245360,
  112.4395173388730, 113.6885631664080, 113.7829032276620, 114.8593231228170, 114.9362313864430,
  115.2603802483160, 115.2831503794140, 115.3057997753690, 115.3328777786540, 115.3644176819730,
  116.1719975004050, 116.2395538155930, 116.3135471602660, 117.5063213205150, 117.5981735777820,
  117.6917977641040, 118.8667409519070, 118.9569097278490, 119.0519636694040, 119.1532611781830,
  121.1123038862090, 123.2510319159420, 125.1548009376000, 125.3012697292210, 125.4498848602980,
  125.6020210193830, 125.7577721104620, 125.9165205934350, 127.9902293737510, 129.8240914946500,
  129.9628303063960, 130.0902523428920, 130.2060516887690, 130.3080532682760, 130.0205543098380,
  129.6625515845250, 129.6079644567380, 129.5544556740550, 129.4999750945010, 129.4445499888220,
  129.3862880472000, 129.3214088137140, 127.8355714487560, 127.7208860721920, 126.6966841024170,
  126.6188138569430, 126.5393130247510, 126.4597262376520, 126.3809405683050, 126.3040516663620,
  125.9143368110930, 125.8886267066410, 125.8478366091090, 125.4840607463960, 125.4490254147970,
  125.4192686476330, 125.3878198562390, 125.3550368883890, 125.0214653090730, 124.9963840024100,
  124.9721007121110, 124.9489369385810, 124.9300453221760, 124.9091930269290, 124.9126645596090,
  124.9229079860540, 125.8968561917180, 125.9764601811320, 126.0627665511000, 127.4481642858080,
  127.5535011690480, 127.6584650806890, 127.7598273224950, 127.8535080648140, 127.9410129673070,
  128.0220631170320, 128.1406747073100, 128.1447311897340, 128.1417075925910, 128.1338423512470,
  128.0090584528720, 127.0445785266120, 126.3658027068100, 126.2247357499530, 126.1309688416810,
  125.0518590586470, 124.1288355076790, 123.9581279749420, 123.7258946261500, 123.6261722826860,
  123.5198154966990, 123.3275779107930, 120.5540003875810, 120.3386529885680, 120.1223203939960,
  117.4236589462100, 117.2221372443720, 117.0027000964890, 116.7701430631430, 116.4291645602000,
  112.5784749180570, 108.8352819374970, 108.5510103352100, 108.2740462506730, 108.0041532095570,
  107.7382220887110, 107.4756306104170, 104.2885433274310, 104.0455589293590, 103.8057126819840,
  103.5710604544990, 101.4294982887000, 101.2703563615260, 101.1146444560290, 100.9629229979930,
  100.8136873437390, 99.2025776112842, 99.0836640708327, 98.9768470739010, 98.8872367716506,
  99.5813569517235, 99.6355925683640, 100.5780421284410, 100.6496628569360, 100.7193269781330,
  100.7871354616700, 100.8550091502180, 100.9298146427980, 102.9907645887770, 103.1555125666920,
  103.3269784905270, 105.8951275579090, 106.0829267695130, 106.2613408807990, 106.4294327752760,
  106.5848661015670, 106.7302551374170, 107.8277447950080, 107.9075250021170, 107.9828995127000,
  108.0541802787360, 108.1217801129290, 108.1855672342090, 108.7114585461720, 108.7500702289660,
  108.7868092123690, 108.8199301842300,
];

describe('MesaAdaptiveMovingAverage', () => {
  const eps = 1e-10;
  const time = new Date(2021, 3, 1);

  it('should have correct output enum values', () => {
    expect(MesaAdaptiveMovingAverageOutput.Value).toBe(0);
    expect(MesaAdaptiveMovingAverageOutput.Fama).toBe(1);
    expect(MesaAdaptiveMovingAverageOutput.Band).toBe(2);
  });

  it('should return expected mnemonic for length params', () => {
    let mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });
    expect(mama.metadata().mnemonic).toBe('mama(3, 39)');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40)');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminatorUnrolled,
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, hdu(4, 0.200, 0.200))');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.PhaseAccumulator,
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, pa(4, 0.150, 0.250))');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.DualDifferentiator,
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, dd(4, 0.150, 0.150))');
  });

  it('should return expected mnemonic for smoothing factor params', () => {
    let mama = MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: 0.5, slowLimitSmoothingFactor: 0.05,
    });
    expect(mama.metadata().mnemonic).toBe('mama(0.500, 0.050)');

    mama = MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: 0.5, slowLimitSmoothingFactor: 0.05,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminatorUnrolled,
    });
    expect(mama.metadata().mnemonic).toBe('mama(0.500, 0.050, hdu(4, 0.200, 0.200))');
  });

  it('should return expected mnemonic for explicit estimator params', () => {
    // Default HomodyneDiscriminator params produce no moniker suffix.
    let mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40)');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: { smoothingLength: 3, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 },
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, hd(3, 0.200, 0.200))');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.3, alphaEmaPeriod: 0.2 },
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, hd(4, 0.300, 0.200))');

    mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 2, slowLimitLength: 40,
      estimatorType: HilbertTransformerCycleEstimatorType.HomodyneDiscriminator,
      estimatorParams: { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.3 },
    });
    expect(mama.metadata().mnemonic).toBe('mama(2, 40, hd(4, 0.200, 0.300))');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 3, slowLimitLength: 39,
      barComponent: BarComponent.Median,
    });
    expect(mama.metadata().mnemonic).toBe('mama(3, 39, hl/2)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 3, slowLimitLength: 39,
      quoteComponent: QuoteComponent.Bid,
    });
    expect(mama.metadata().mnemonic).toBe('mama(3, 39, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const mama = MesaAdaptiveMovingAverage.fromLength({
      fastLimitLength: 3, slowLimitLength: 39,
      tradeComponent: TradeComponent.Volume,
    });
    expect(mama.metadata().mnemonic).toBe('mama(3, 39, v)');
  });

  it('should return expected metadata', () => {
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });
    const meta = mama.metadata();
    const mn = 'mama(3, 39)';
    const mnFama = 'fama(3, 39)';
    const mnBand = 'mama-fama(3, 39)';
    const descr = 'Mesa adaptive moving average ';

    expect(meta.identifier).toBe(IndicatorIdentifier.MesaAdaptiveMovingAverage);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe(descr + mn);
    expect(meta.outputs.length).toBe(3);

    expect(meta.outputs[0].kind).toBe(MesaAdaptiveMovingAverageOutput.Value);
    expect(meta.outputs[0].shape).toBe(Shape.Scalar);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe(descr + mn);

    expect(meta.outputs[1].kind).toBe(MesaAdaptiveMovingAverageOutput.Fama);
    expect(meta.outputs[1].shape).toBe(Shape.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnFama);
    expect(meta.outputs[1].description).toBe(descr + mnFama);

    expect(meta.outputs[2].kind).toBe(MesaAdaptiveMovingAverageOutput.Band);
    expect(meta.outputs[2].shape).toBe(Shape.Band);
    expect(meta.outputs[2].mnemonic).toBe(mnBand);
    expect(meta.outputs[2].description).toBe(descr + mnBand);
  });

  it('should throw if the fast limit length is less than 2', () => {
    expect(() => MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 1, slowLimitLength: 39 })).toThrow();
  });

  it('should throw if the slow limit length is less than 2', () => {
    expect(() => MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 1 })).toThrow();
  });

  it('should throw if the fast limit smoothing factor is out of range', () => {
    expect(() => MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: -0.01, slowLimitSmoothingFactor: 0.05,
    })).toThrow();
    expect(() => MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: 1.01, slowLimitSmoothingFactor: 0.05,
    })).toThrow();
  });

  it('should throw if the slow limit smoothing factor is out of range', () => {
    expect(() => MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: 0.5, slowLimitSmoothingFactor: -0.01,
    })).toThrow();
    expect(() => MesaAdaptiveMovingAverage.fromSmoothingFactor({
      fastLimitSmoothingFactor: 0.5, slowLimitSmoothingFactor: 1.01,
    })).toThrow();
  });

  it('should calculate expected MAMA update values and prime state', () => {
    const lprimed = 26;
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });

    for (let i = 0; i < lprimed; i++) {
      expect(mama.update(input[i])).toBeNaN();
      expect(mama.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      const act = mama.update(input[i]);
      expect(mama.isPrimed()).toBe(true);
      expect(Math.abs(act - expectedMama[i]))
        .withContext(`MAMA[${i}]: expected ${expectedMama[i]}, actual ${act}`)
        .toBeLessThan(eps);
    }

    expect(mama.update(Number.NaN)).toBeNaN();
  });

  it('should produce expected MAMA/FAMA/Band via updateScalar', () => {
    const lprimed = 26;
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });

    for (let i = 0; i < lprimed; i++) {
      const out = mama.updateScalar(new Scalar({ time, value: input[i] }));
      expect(out.length).toBe(3);
      expect(Number.isNaN((out[0] as Scalar).value)).toBe(true);
      expect(Number.isNaN((out[1] as Scalar).value)).toBe(true);
      expect(Number.isNaN((out[2] as Band).upper)).toBe(true);
      expect(Number.isNaN((out[2] as Band).lower)).toBe(true);
      expect(mama.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < input.length; i++) {
      const out = mama.updateScalar(new Scalar({ time, value: input[i] }));
      expect(out.length).toBe(3);
      expect(mama.isPrimed()).toBe(true);

      const s0 = out[0] as Scalar;
      const s1 = out[1] as Scalar;
      const b = out[2] as Band;

      expect(Math.abs(s0.value - expectedMama[i]))
        .withContext(`MAMA[${i}]: expected ${expectedMama[i]}, actual ${s0.value}`)
        .toBeLessThan(eps);
      expect(Math.abs(s1.value - expectedFama[i]))
        .withContext(`FAMA[${i}]: expected ${expectedFama[i]}, actual ${s1.value}`)
        .toBeLessThan(eps);
      expect(b.upper).toBe(s0.value);
      expect(b.lower).toBe(s1.value);
      expect(b.time).toBe(time);
    }
  });

  it('should produce expected output via updateBar', () => {
    const lprimed = 26;
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });

    for (let i = 0; i < input.length; i++) {
      const bar = new Bar({ time, open: input[i], high: input[i], low: input[i], close: input[i], volume: 0 });
      const out = mama.updateBar(bar);
      expect(out.length).toBe(3);
      if (i >= lprimed) {
        expect(Math.abs((out[0] as Scalar).value - expectedMama[i])).toBeLessThan(eps);
        expect(Math.abs((out[1] as Scalar).value - expectedFama[i])).toBeLessThan(eps);
      }
    }
  });

  it('should produce expected output via updateQuote', () => {
    const lprimed = 26;
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });

    for (let i = 0; i < input.length; i++) {
      const q = new Quote({ time, bidPrice: input[i], askPrice: input[i], bidSize: 0, askSize: 0 });
      const out = mama.updateQuote(q);
      expect(out.length).toBe(3);
      if (i >= lprimed) {
        expect(Math.abs((out[0] as Scalar).value - expectedMama[i])).toBeLessThan(eps);
        expect(Math.abs((out[1] as Scalar).value - expectedFama[i])).toBeLessThan(eps);
      }
    }
  });

  it('should produce expected output via updateTrade', () => {
    const lprimed = 26;
    const mama = MesaAdaptiveMovingAverage.fromLength({ fastLimitLength: 3, slowLimitLength: 39 });

    for (let i = 0; i < input.length; i++) {
      const r = new Trade({ time, price: input[i], volume: 0 });
      const out = mama.updateTrade(r);
      expect(out.length).toBe(3);
      if (i >= lprimed) {
        expect(Math.abs((out[0] as Scalar).value - expectedMama[i])).toBeLessThan(eps);
        expect(Math.abs((out[1] as Scalar).value - expectedFama[i])).toBeLessThan(eps);
      }
    }
  });

  it('should construct via default() factory', () => {
    const mama = MesaAdaptiveMovingAverage.default();
    expect(mama.metadata().mnemonic).toBe('mama(3, 39)');
    expect(mama.isPrimed()).toBe(false);
  });
});
