import { } from 'jasmine';

import { FractalAdaptiveMovingAverage } from './fractal-adaptive-moving-average';
import { FractalAdaptiveMovingAverageOutput } from './fractal-adaptive-moving-average-output';
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';
import { BarComponent } from '../../../entities/bar-component';
import { QuoteComponent } from '../../../entities/quote-component';
import { TradeComponent } from '../../../entities/trade-component';
import { Scalar } from '../../../entities/scalar';
import { Bar } from '../../../entities/bar';
import { Quote } from '../../../entities/quote';
import { Trade } from '../../../entities/trade';

/* eslint-disable max-len */
// Input data taken from test_FRAMA.xsl reference implementation:
// Mid-Price, D5...D256, 252 entries,
// High, B5...B256, 252 entries,
// Low, C5...C256, 252 entries.
//
// Expected data taken from test_FRAMA.xsl reference implementation:
// FRAMA, R5...R256, 252 entries,
// FDIM, O5...O256, 252 entries.
//
// All parameters have default values.

const inputMid = [
  92.00000, 93.17250, 95.31250, 94.84500, 94.40750, 94.11000, 93.50000, 91.73500, 90.95500, 91.68750,
  94.50000, 97.97000, 97.57750, 90.78250, 89.03250, 92.09500, 91.15500, 89.71750, 90.61000, 91.00000,
  88.92250, 87.51500, 86.43750, 83.89000, 83.00250, 82.81250, 82.84500, 86.73500, 86.86000, 87.54750,
  85.78000, 86.17250, 86.43750, 87.25000, 88.93750, 88.20500, 85.81250, 84.59500, 83.65750, 84.45500,
  83.50000, 86.78250, 88.17250, 89.26500, 90.86000, 90.78250, 91.86000, 90.36000, 89.86000, 90.92250,
  89.50000, 87.67250, 86.50000, 84.28250, 82.90750, 84.25000, 85.68750, 86.61000, 88.28250, 89.53250,
  89.50000, 88.09500, 90.62500, 92.23500, 91.67250, 92.59250, 93.01500, 91.17250, 90.98500, 90.37750,
  88.25000, 86.90750, 84.09250, 83.18750, 84.25250, 97.86000, 99.87500, 103.26500, 105.93750, 103.50000,
  103.11000, 103.61000, 104.64000, 106.81500, 104.95250, 105.50000, 107.14000, 109.73500, 109.84500, 110.98500,
  120.00000, 119.87500, 117.90750, 119.40750, 117.95250, 117.22000, 115.64250, 113.11000, 111.75000, 114.51750,
  114.74500, 115.47000, 112.53000, 112.03000, 113.43500, 114.22000, 119.59500, 117.96500, 118.71500, 115.03000,
  114.53000, 115.00000, 116.53000, 120.18500, 120.50000, 120.59500, 124.18500, 125.37500, 122.97000, 123.00000,
  124.43500, 123.44000, 124.03000, 128.18500, 129.65500, 130.87500, 132.34500, 132.06500, 133.81500, 135.66000,
  137.03500, 137.47000, 137.34500, 136.31500, 136.44000, 136.28500, 129.09500, 128.31000, 126.00000, 124.03000,
  123.93500, 125.03000, 127.25000, 125.62000, 125.53000, 123.90500, 120.65500, 119.96500, 120.78000, 124.00000,
  122.78000, 120.72000, 121.78000, 122.40500, 123.25000, 126.18500, 127.56000, 126.56500, 123.06000, 122.71500,
  123.59000, 122.31000, 122.46500, 123.96500, 123.97000, 124.15500, 124.43500, 127.00000, 125.50000, 128.87500,
  130.53500, 132.31500, 134.06500, 136.03500, 133.78000, 132.75000, 133.47000, 130.97000, 127.59500, 128.44000,
  127.94000, 125.81000, 124.62500, 122.72000, 124.09000, 123.22000, 121.40500, 120.93500, 118.28000, 118.37500,
  121.15500, 120.90500, 117.12500, 113.06000, 114.90500, 112.43500, 107.93500, 105.97000, 106.37000, 106.84500,
  106.97000, 110.03000, 91.00000, 93.56000, 93.62000, 95.31000, 94.18500, 94.78000, 97.62500, 97.59000,
  95.25000, 94.72000, 92.22000, 91.56500, 92.22000, 93.81000, 95.59000, 96.18500, 94.62500, 95.12000,
  94.00000, 93.74500, 95.90500, 101.74500, 106.44000, 107.93500, 103.40500, 105.06000, 104.15500, 103.31000,
  103.34500, 104.84000, 110.40500, 114.50000, 117.31500, 118.25000, 117.18500, 109.75000, 109.65500, 108.53000,
  106.22000, 107.72000, 109.84000, 109.09500, 109.09000, 109.15500, 109.31500, 109.06000, 109.90500, 109.62500,
  109.53000, 108.06000,
];

const inputHigh = [
  93.2500, 94.9400, 96.3750, 96.1900, 96.0000, 94.7200, 95.0000, 93.7200, 92.4700, 92.7500,
  96.2500, 99.6250, 99.1250, 92.7500, 91.3150, 93.2500, 93.4050, 90.6550, 91.9700, 92.2500,
  90.3450, 88.5000, 88.2500, 85.5000, 84.4400, 84.7500, 84.4400, 89.4050, 88.1250, 89.1250,
  87.1550, 87.2500, 87.3750, 88.9700, 90.0000, 89.8450, 86.9700, 85.9400, 84.7500, 85.4700,
  84.4700, 88.5000, 89.4700, 90.0000, 92.4400, 91.4400, 92.9700, 91.7200, 91.1550, 91.7500,
  90.0000, 88.8750, 89.0000, 85.2500, 83.8150, 85.2500, 86.6250, 87.9400, 89.3750, 90.6250,
  90.7500, 88.8450, 91.9700, 93.3750, 93.8150, 94.0300, 94.0300, 91.8150, 92.0000, 91.9400,
  89.7500, 88.7500, 86.1550, 84.8750, 85.9400, 99.3750, 103.2800, 105.3750, 107.6250, 105.2500,
  104.5000, 105.5000, 106.1250, 107.9400, 106.2500, 107.0000, 108.7500, 110.9400, 110.9400, 114.2200,
  123.0000, 121.7500, 119.8150, 120.3150, 119.3750, 118.1900, 116.6900, 115.3450, 113.0000, 118.3150,
  116.8700, 116.7500, 113.8700, 114.6200, 115.3100, 116.0000, 121.6900, 119.8700, 120.8700, 116.7500,
  116.5000, 116.0000, 118.3100, 121.5000, 122.0000, 121.4400, 125.7500, 127.7500, 124.1900, 124.4400,
  125.7500, 124.6900, 125.3100, 132.0000, 131.3100, 132.2500, 133.8800, 133.5000, 135.5000, 137.4400,
  138.6900, 139.1900, 138.5000, 138.1300, 137.5000, 138.8800, 132.1300, 129.7500, 128.5000, 125.4400,
  125.1200, 126.5000, 128.6900, 126.6200, 126.6900, 126.0000, 123.1200, 121.8700, 124.0000, 127.0000,
  124.4400, 122.5000, 123.7500, 123.8100, 124.5000, 127.8700, 128.5600, 129.6300, 124.8700, 124.3700,
  124.8700, 123.6200, 124.0600, 125.8700, 125.1900, 125.6200, 126.0000, 128.5000, 126.7500, 129.7500,
  132.6900, 133.9400, 136.5000, 137.6900, 135.5600, 133.5600, 135.0000, 132.3800, 131.4400, 130.8800,
  129.6300, 127.2500, 127.8100, 125.0000, 126.8100, 124.7500, 122.8100, 122.2500, 121.0600, 120.0000,
  123.2500, 122.7500, 119.1900, 115.0600, 116.6900, 114.8700, 110.8700, 107.2500, 108.8700, 109.0000,
  108.5000, 113.0600, 93.0000, 94.6200, 95.1200, 96.0000, 95.5600, 95.3100, 99.0000, 98.8100,
  96.8100, 95.9400, 94.4400, 92.9400, 93.9400, 95.5000, 97.0600, 97.5000, 96.2500, 96.3700,
  95.0000, 94.8700, 98.2500, 105.1200, 108.4400, 109.8700, 105.0000, 106.0000, 104.9400, 104.5000,
  104.4400, 106.3100, 112.8700, 116.5000, 119.1900, 121.0000, 122.1200, 111.9400, 112.7500, 110.1900,
  107.9400, 109.6900, 111.0600, 110.4400, 110.1200, 110.3100, 110.4400, 110.0000, 110.7500, 110.5000,
  110.5000, 109.5000,
];

const inputLow = [
  90.7500, 91.4050, 94.2500, 93.5000, 92.8150, 93.5000, 92.0000, 89.7500, 89.4400, 90.6250,
  92.7500, 96.3150, 96.0300, 88.8150, 86.7500, 90.9400, 88.9050, 88.7800, 89.2500, 89.7500,
  87.5000, 86.5300, 84.6250, 82.2800, 81.5650, 80.8750, 81.2500, 84.0650, 85.5950, 85.9700,
  84.4050, 85.0950, 85.5000, 85.5300, 87.8750, 86.5650, 84.6550, 83.2500, 82.5650, 83.4400,
  82.5300, 85.0650, 86.8750, 88.5300, 89.2800, 90.1250, 90.7500, 89.0000, 88.5650, 90.0950,
  89.0000, 86.4700, 84.0000, 83.3150, 82.0000, 83.2500, 84.7500, 85.2800, 87.1900, 88.4400,
  88.2500, 87.3450, 89.2800, 91.0950, 89.5300, 91.1550, 92.0000, 90.5300, 89.9700, 88.8150,
  86.7500, 85.0650, 82.0300, 81.5000, 82.5650, 96.3450, 96.4700, 101.1550, 104.2500, 101.7500,
  101.7200, 101.7200, 103.1550, 105.6900, 103.6550, 104.0000, 105.5300, 108.5300, 108.7500, 107.7500,
  117.0000, 118.0000, 116.0000, 118.5000, 116.5300, 116.2500, 114.5950, 110.8750, 110.5000, 110.7200,
  112.6200, 114.1900, 111.1900, 109.4400, 111.5600, 112.4400, 117.5000, 116.0600, 116.5600, 113.3100,
  112.5600, 114.0000, 114.7500, 118.8700, 119.0000, 119.7500, 122.6200, 123.0000, 121.7500, 121.5600,
  123.1200, 122.1900, 122.7500, 124.3700, 128.0000, 129.5000, 130.8100, 130.6300, 132.1300, 133.8800,
  135.3800, 135.7500, 136.1900, 134.5000, 135.3800, 133.6900, 126.0600, 126.8700, 123.5000, 122.6200,
  122.7500, 123.5600, 125.8100, 124.6200, 124.3700, 121.8100, 118.1900, 118.0600, 117.5600, 121.0000,
  121.1200, 118.9400, 119.8100, 121.0000, 122.0000, 124.5000, 126.5600, 123.5000, 121.2500, 121.0600,
  122.3100, 121.0000, 120.8700, 122.0600, 122.7500, 122.6900, 122.8700, 125.5000, 124.2500, 128.0000,
  128.3800, 130.6900, 131.6300, 134.3800, 132.0000, 131.9400, 131.9400, 129.5600, 123.7500, 126.0000,
  126.2500, 124.3700, 121.4400, 120.4400, 121.3700, 121.6900, 120.0000, 119.6200, 115.5000, 116.7500,
  119.0600, 119.0600, 115.0600, 111.0600, 113.1200, 110.0000, 105.0000, 104.6900, 103.8700, 104.6900,
  105.4400, 107.0000, 89.0000, 92.5000, 92.1200, 94.6200, 92.8100, 94.2500, 96.2500, 96.3700,
  93.6900, 93.5000, 90.0000, 90.1900, 90.5000, 92.1200, 94.1200, 94.8700, 93.0000, 93.8700,
  93.0000, 92.6200, 93.5600, 98.3700, 104.4400, 106.0000, 101.8100, 104.1200, 103.3700, 102.1200,
  102.2500, 103.3700, 107.9400, 112.5000, 115.4400, 115.5000, 112.2500, 107.5600, 106.5600, 106.8700,
  104.5000, 105.7500, 108.6200, 107.7500, 108.0600, 108.0000, 108.1900, 108.1200, 109.0600, 108.7500,
  108.5600, 106.6200,
];

const expectedFrama = [
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  89.2267105034236, 89.3368287632670, 89.3585676433656, 89.4331023968400, 89.4708293163365,
  89.3786991078887, 89.1255621954484, 88.8911587615043, 88.3130074555269, 87.4050766949446,
  86.5735654739459, 85.8984882273912, 86.1518019830071, 86.3493322808939, 86.6530660498765,
  86.5510640984500, 86.5320480827810, 86.5291445959233, 86.5490145722999, 86.6957869186919,
  86.7648461717294, 86.7125962031917, 86.6436759299080, 86.5710514203313, 86.5195885459905,
  86.4472121610614, 86.4543310999726, 86.4924764347400, 86.5600569180083, 86.7797971506857,
  86.9067055814372, 87.0457812264415, 87.1388358835381, 87.3467219291760, 87.9745429969702,
  88.6681194123833, 88.5243479341425, 88.4848284694531, 88.4200869036510, 88.3243232561998,
  88.2535448607335, 88.1448428308959, 87.8545582247850, 88.1159382988993, 88.2836225449494,
  88.3339075161590, 88.3262310495505, 88.3644658131460, 88.5244273243834, 88.6973941261061,
  89.6630232769251, 90.8525688643512, 90.9170423613911, 90.9300171591977, 90.7249066522646,
  90.5453798195497, 90.4934124418101, 90.3130522022209, 89.9186462617956, 89.2240044723459,
  90.2701154779666, 91.8923806944276, 94.9530427185545, 97.0984638845931, 98.0000897849679,
  98.4148534167124, 98.8780464828365, 104.6400000000000, 105.7281135005410, 105.2640337128590,
  105.3638604047110, 105.8464041001890, 107.0007497560500, 107.8450771168450, 109.1506183650330,
  113.4637991525930, 115.8347532798220, 116.5235456965500, 117.9374769010970, 117.9485496987280,
  117.6605266243390, 116.8627248634490, 115.5295343664310, 115.3497433623420, 115.2952692253930,
  115.2606805426730, 115.2743424223860, 115.0986932128960, 114.9529749204760, 114.8808931912170,
  114.8447466307030, 114.9094693543300, 114.9707228174530, 115.0457834111500, 115.0453608771260,
  115.0259380415870, 115.0245360589650, 115.1360352883790, 115.5268497579460, 115.6587496384770,
  115.7896708456130, 116.2275496776190, 116.8498538013430, 117.4455188697390, 117.9647000424260,
  120.0690362614910, 121.2100854001430, 122.3364516152540, 127.3768981361550, 127.9997579600520,
  128.4674127614630, 129.1077605668970, 129.6506439679060, 130.2634243522160, 131.3595250173300,
  133.2658670379710, 134.3464896763640, 135.5889317598570, 135.9938109893880, 136.1388017723700,
  136.2150501151920, 135.8365132427990, 135.6080745909970, 135.4788216502980, 135.0786044987580,
  134.5127512444070, 133.8654963662030, 133.4139489002410, 125.6200000000000, 125.5963341903920,
  124.9331361137190, 123.9468540518040, 123.2192653200150, 122.7575488300800, 122.9927262781350,
  122.9312511703190, 122.5455560593910, 122.4459636945080, 122.4427424463040, 122.4595145864780,
  122.5154345871760, 122.5958812939760, 122.6735494933150, 122.6864691474850, 122.6877483835490,
  122.7495480325670, 122.7194412268340, 122.7028238705180, 122.7632690438560, 122.7962194389690,
  122.9023846742710, 123.0157441087240, 123.1031383889540, 123.1424928523140, 123.2054796604150,
  123.4789498668650, 124.0005457069040, 124.8767652824900, 128.6268145966040, 131.2966442267030,
  131.6219911641770, 133.1431746234470, 131.8514826815970, 131.6406729224900, 131.5384698262220,
  131.4780656002040, 131.3626084307420, 131.1316858570180, 130.7950021389090, 129.3038419034150,
  127.2479978676120, 126.6321157520150, 125.7555252839640, 123.3074419111100, 121.8578512690060,
  121.7182148617720, 121.5938798479940, 120.5729187185010, 117.3879061290530, 116.4825108668860,
  115.1715003505450, 114.1349929815520, 112.9387272954480, 111.8793016022070, 110.0104915738280,
  108.1652419969530, 108.5078121377520, 99.8204338949554, 96.2748506210997, 95.6228459686871,
  95.5496516533106, 95.2741320987552, 95.1895928551543, 95.5523570136138, 97.5900000000000,
  97.1015881952464, 96.6536753055277, 96.1168120004233, 95.5656438010483, 95.1605276694061,
  94.9969954532945, 95.0716039682056, 95.0986398055546, 95.0775867733060, 95.0794720182091,
  95.0586898125545, 95.0299593356558, 95.0577051881919, 95.7026067506480, 97.2720402379149,
  99.7506289364549, 100.6000819045310, 101.6367830489120, 102.2147892036990, 102.5348611533410,
  103.3450000000000, 103.8107505617500, 104.8152343011550, 106.2179387231950, 108.2442461404210,
  110.2561373093930, 111.7271316870480, 111.3163200481330, 110.9833247908240, 110.0835836125560,
  109.6933894692140, 109.6031011050960, 109.6092038238180, 109.5997557696640, 109.5483679235420,
  109.5092127187630, 109.4827494282640, 109.3894652250650, 109.4790456471110, 109.5255507143980,
  109.5271161293130, 109.1649886027460,
];

const expectedFdim = [
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  Number.NaN, Number.NaN, Number.NaN, Number.NaN, Number.NaN,
  1.59890169167903, 1.62165644793154, 1.62165644793154, 1.61252428020110, 1.80919463544085,
  1.38731970610845, 1.43351015935032, 1.52973767168433, 1.46851456520950, 1.38354164634960,
  1.37109415169125, 1.37109415169125, 1.25940664178948, 1.27726049384560, 1.29801226771749,
  1.46621931004478, 1.64950994910223, 1.75636643108938, 1.77982540391337, 1.60573917939322,
  1.66976434212608, 1.63035440353517, 1.74374807938101, 1.80701608891535, 1.80701608891535,
  1.81017544111998, 1.83650126771712, 1.82681221154869, 1.80652697653078, 1.64577656342613,
  1.74943143817906, 1.77582152998323, 1.77582152998323, 1.55846469854408, 1.37776733571198,
  1.17115282708314, 1.42021030883958, 1.85473705279196, 1.90615339833091, 1.88007738492828,
  1.88007738492828, 1.68651329535641, 1.36161995216723, 1.10705610357065, 1.46337161664613,
  1.69181507815183, 1.74653422398087, 1.88951845658424, 1.69187770463767, 1.63004099736752,
  1.30285444944473, 1.22495995189646, 1.34783764032371, 1.35956884043905, 1.21517893110387,
  1.56971472702222, 1.92255877234439, 1.77505054148317, 1.62843753928139, 1.45576343705044,
  1.45836731881707, 1.38618514567424, 1.28502256328983, 1.35463302165334, 1.42562891054656,
  1.54530630036937, 1.52491787818741, 1.00000000000000, 1.15039253126962, 1.11152635307626,
  1.18680163989364, 1.28297022510159, 1.26372847050257, 1.26372847050257, 1.19056419205197,
  1.20030366202746, 1.21600811553991, 1.23922893975500, 1.15478004773904, 1.06625108183884,
  1.20151595057820, 1.20151595057820, 1.22472725764617, 1.66133517039154, 1.59202999088372,
  1.60082060116837, 1.59264957145672, 1.59689604972804, 1.66171969970519, 1.66171969970519,
  1.63103214941714, 1.93282999496790, 1.84897799808605, 1.84897799808605, 1.78616959216698,
  1.71189940689006, 1.63359727597203, 1.56519924589973, 1.55561585030035, 1.78819358878065,
  1.78819358878065, 1.64134188857247, 1.58364869037548, 1.50587983211848, 1.51466228033362,
  1.24390464224437, 1.23522486668123, 1.19927817027484, 1.03228953795651, 1.28159141036807,
  1.39437450609095, 1.39107282597327, 1.36808996279667, 1.41612149890974, 1.34613391399291,
  1.23690069636080, 1.29500119064180, 1.19131469760116, 1.12682597114944, 1.24408923724970,
  1.14135584924554, 1.63718743048263, 1.75891212086056, 1.93559796663019, 1.72823254667111,
  1.64716093602379, 1.58292951290891, 1.58292951290891, 1.00000000000000, 1.29006057157969,
  1.20329308509134, 1.31862673179711, 1.36909968657346, 1.36144182794620, 1.36144182794620,
  1.26956092132206, 1.37919697099394, 1.44287547660124, 1.55218746462954, 1.84121181705624,
  1.91180784958885, 1.89865774377969, 1.85422542556859, 1.73792149616632, 1.67418202708844,
  1.58217083010958, 1.58217083010958, 1.59251277433104, 1.65987916755774, 1.78187491356854,
  1.55358349924761, 1.56548774675230, 1.82943210678232, 1.89232448247385, 1.97954750208345,
  1.71408276792926, 1.61446202611031, 1.53008864019217, 1.23677925693134, 1.14279610968028,
  1.32501260856216, 1.04226122447158, 1.11296783739634, 1.65258007936250, 1.74788862135100,
  1.88752527074159, 1.84550698090478, 1.73251968913501, 1.69883046527944, 1.32643730340908,
  1.23559387774326, 1.48856906743990, 1.40642914975894, 1.24240775524384, 1.26590830128790,
  1.35093238357993, 1.40780593004777, 1.32059473523525, 1.18634876848019, 1.21906104149649,
  1.24479092400705, 1.42197809672361, 1.41706408511445, 1.39620536549780, 1.21518703719479,
  1.10844434488427, 1.36793648676589, 1.15217157353877, 1.12345837208153, 1.30489469762110,
  1.31542660504745, 1.34743468903322, 1.38339235438793, 1.41347360206241, 1.00000000000000,
  1.34021485262743, 1.36283655427296, 1.45845008211624, 1.45845008211624, 1.45845008211624,
  1.45845008211624, 1.45013481390423, 1.80735492205760, 1.67606676922226, 1.67606676922226,
  1.85775987549485, 1.83007499855769, 1.74941508099970, 1.50787852025575, 1.41757797206962,
  1.31683665260466, 1.31683665260466, 1.31683665260466, 1.31958034003094, 1.26712505052561,
  1.00000000000000, 1.25324390264098, 1.40861123011941, 1.41956147614643, 1.36925132500904,
  1.34832266537189, 1.33652547110032, 1.34119642859214, 1.34900762894365, 1.21781872510100,
  1.49785478440575, 1.66979051096921, 1.79451984062430, 1.86789646399265, 1.49825086752783,
  1.50100472182287, 1.43281691184008, 1.32813746131429, 1.38002243066038, 1.24835838733530,
  1.22683040844769, 1.30380147670304,
];

describe('FractalAdaptiveMovingAverage', () => {
  const time = new Date(2021, 3, 1); // April 1, 2021

  it('should have correct output enum values', () => {
    expect(FractalAdaptiveMovingAverageOutput.Value).toBe(0);
    expect(FractalAdaptiveMovingAverageOutput.Fdim).toBe(1);
  });

  it('should return expected mnemonic', () => {
    let frama = new FractalAdaptiveMovingAverage(
      { length: 16, slowestSmoothingFactor: 0.01 });
    expect(frama.metadata().mnemonic).toBe('frama(16, 0.010)');

    frama = new FractalAdaptiveMovingAverage(
      { length: 18, slowestSmoothingFactor: 0.005 });
    expect(frama.metadata().mnemonic).toBe('frama(18, 0.005)');

    frama = new FractalAdaptiveMovingAverage(
      { length: 17, slowestSmoothingFactor: 0.01 });
    expect(frama.metadata().mnemonic).toBe('frama(18, 0.010)');
  });

  it('should return expected mnemonic with non-default bar component', () => {
    const frama = new FractalAdaptiveMovingAverage({
      length: 16, slowestSmoothingFactor: 0.01,
      barComponent: BarComponent.Median,
    });
    expect(frama.metadata().mnemonic).toBe('frama(16, 0.010, hl/2)');
    expect(frama.metadata().description).toBe('Fractal adaptive moving average frama(16, 0.010, hl/2)');
  });

  it('should return expected mnemonic with non-default quote component', () => {
    const frama = new FractalAdaptiveMovingAverage({
      length: 16, slowestSmoothingFactor: 0.01,
      quoteComponent: QuoteComponent.Bid,
    });
    expect(frama.metadata().mnemonic).toBe('frama(16, 0.010, b)');
  });

  it('should return expected mnemonic with non-default trade component', () => {
    const frama = new FractalAdaptiveMovingAverage({
      length: 16, slowestSmoothingFactor: 0.01,
      tradeComponent: TradeComponent.Volume,
    });
    expect(frama.metadata().mnemonic).toBe('frama(16, 0.010, v)');
  });

  it('should return expected metadata', () => {
    const frama = new FractalAdaptiveMovingAverage(
      { length: 16, slowestSmoothingFactor: 0.01 });
    const meta = frama.metadata();

    const mn = 'frama(16, 0.010)';
    const mnFdim = 'framaDim(16, 0.010)';
    const descr = 'Fractal adaptive moving average ';

    expect(meta.type).toBe(IndicatorType.FractalAdaptiveMovingAverage);
    expect(meta.mnemonic).toBe(mn);
    expect(meta.description).toBe(descr + mn);
    expect(meta.outputs.length).toBe(2);

    expect(meta.outputs[0].kind).toBe(FractalAdaptiveMovingAverageOutput.Value);
    expect(meta.outputs[0].type).toBe(OutputType.Scalar);
    expect(meta.outputs[0].mnemonic).toBe(mn);
    expect(meta.outputs[0].description).toBe(descr + mn);

    expect(meta.outputs[1].kind).toBe(FractalAdaptiveMovingAverageOutput.Fdim);
    expect(meta.outputs[1].type).toBe(OutputType.Scalar);
    expect(meta.outputs[1].mnemonic).toBe(mnFdim);
    expect(meta.outputs[1].description).toBe(descr + mnFdim);
  });

  it('should throw if the length is less than 2', () => {
    expect(() => { new FractalAdaptiveMovingAverage({ length: 1, slowestSmoothingFactor: 0.01 }); }).toThrow();
    expect(() => { new FractalAdaptiveMovingAverage({ length: 0, slowestSmoothingFactor: 0.01 }); }).toThrow();
    expect(() => { new FractalAdaptiveMovingAverage({ length: -1, slowestSmoothingFactor: 0.01 }); }).toThrow();
  });

  it('should throw if the slowest smoothing factor is less than 0', () => {
    expect(() => {
      new FractalAdaptiveMovingAverage(
        { length: 16, slowestSmoothingFactor: -0.01 });
    }).toThrow();
  });

  it('should throw if the slowest smoothing factor is greater than 1', () => {
    expect(() => {
      new FractalAdaptiveMovingAverage(
        { length: 16, slowestSmoothingFactor: 1.01 });
    }).toThrow();
  });

  it('should calculate expected FRAMA values from reference implementation', () => {
    const lprimed = 15;
    const eps = 1e-9;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      expect(frama.update(inputMid[i], inputHigh[i], inputLow[i])).toBeNaN();
      expect(frama.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < inputMid.length; i++) {
      const act = frama.update(inputMid[i], inputHigh[i], inputLow[i]);
      expect(frama.isPrimed()).toBe(true);
      expect(Math.abs(act - expectedFrama[i])).withContext(`FRAMA [${i}]: expected ${expectedFrama[i]}, actual ${act}`)
        .toBeLessThan(eps);
    }

    expect(frama.update(Number.NaN, Number.NaN, Number.NaN)).toBeNaN();
    expect(frama.update(Number.NaN, 1, 1)).toBeNaN();
    expect(frama.update(1, Number.NaN, 1)).toBeNaN();
    expect(frama.update(1, 1, Number.NaN)).toBeNaN();
  });

  it('should calculate expected Fdim values from reference implementation', () => {
    const lprimed = 15;
    const eps = 1e-9;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    // Access fractal dimension via updateEntity output.
    for (let i = 0; i < lprimed; i++) {
      const output = frama.updateScalar(new Scalar({ time, value: inputMid[i] }));
      // Before primed, scalar uses sample for high/low too, which is different from the reference.
      // Instead, use update directly and check output entity.
    }

    // Re-create to test Fdim properly through the entity path with high/low.
    const frama2 = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama2.update(inputMid[i], inputHigh[i], inputLow[i]);
    }

    for (let i = lprimed; i < inputMid.length; i++) {
      const bar = new Bar({ time, open: inputMid[i], high: inputHigh[i], low: inputLow[i], close: inputMid[i], volume: 0 });
      const output = frama2.updateBar(bar);
      expect(output.length).toBe(2);

      const framaScalar = output[0] as Scalar;
      const fdimScalar = output[1] as Scalar;

      expect(Math.abs(framaScalar.value - expectedFrama[i]))
        .withContext(`FRAMA [${i}]: expected ${expectedFrama[i]}, actual ${framaScalar.value}`)
        .toBeLessThan(eps);

      expect(Math.abs(fdimScalar.value - expectedFdim[i]))
        .withContext(`FDIM [${i}]: expected ${expectedFdim[i]}, actual ${fdimScalar.value}`)
        .toBeLessThan(eps);
    }
  });

  it('should transition primed state correctly', () => {
    const lprimed = 15;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    expect(frama.isPrimed()).toBe(false);

    for (let i = 0; i < lprimed; i++) {
      frama.update(inputMid[i], inputHigh[i], inputLow[i]);
      expect(frama.isPrimed()).toBe(false);
    }

    for (let i = lprimed; i < inputMid.length; i++) {
      frama.update(inputMid[i], inputHigh[i], inputLow[i]);
      expect(frama.isPrimed()).toBe(true);
    }
  });

  it('should produce correct updateEntity output for scalar', () => {
    const lprimed = 15;
    const inp = 3;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama.update(0, 0, 0);
    }

    const s = new Scalar({ time, value: inp });
    const output = frama.updateScalar(s);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(s0.time).toBe(time);
    expect(s1.time).toBe(time);
    // After priming with zeros, update with 3 should produce a specific value.
    expect(Number.isNaN(s0.value)).toBe(false);
    expect(Number.isNaN(s1.value)).toBe(false);
  });

  it('should produce correct updateEntity output for bar', () => {
    const lprimed = 15;
    const inp = 3;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama.update(0, 0, 0);
    }

    const b = new Bar({ time, open: inp, high: inp, low: inp, close: inp, volume: 0 });
    const output = frama.updateBar(b);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(s0.time).toBe(time);
    expect(s1.time).toBe(time);
    expect(Number.isNaN(s0.value)).toBe(false);
    expect(Number.isNaN(s1.value)).toBe(false);
  });

  it('should produce correct updateEntity output for quote', () => {
    const lprimed = 15;
    const inp = 3;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama.update(0, 0, 0);
    }

    const q = new Quote({ time, bidPrice: inp, askPrice: inp, bidSize: 0, askSize: 0 });
    const output = frama.updateQuote(q);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(s0.time).toBe(time);
    expect(s1.time).toBe(time);
    expect(Number.isNaN(s0.value)).toBe(false);
    expect(Number.isNaN(s1.value)).toBe(false);
  });

  it('should produce correct updateEntity output for trade', () => {
    const lprimed = 15;
    const inp = 3;
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama.update(0, 0, 0);
    }

    const r = new Trade({ time, price: inp, volume: 0 });
    const output = frama.updateTrade(r);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(s0.time).toBe(time);
    expect(s1.time).toBe(time);
    expect(Number.isNaN(s0.value)).toBe(false);
    expect(Number.isNaN(s1.value)).toBe(false);
  });

  it('should return NaN for fdim when frama is NaN (not primed)', () => {
    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    const s = new Scalar({ time, value: 100 });
    const output = frama.updateScalar(s);
    expect(output.length).toBe(2);

    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;
    expect(Number.isNaN(s0.value)).toBe(true);
    expect(Number.isNaN(s1.value)).toBe(true);
  });

  it('should match Go updateEntity values exactly', () => {
    const lprimed = 15;
    const inp = 3;
    const expectedFramaValue = 2.999999999999997;
    const expectedFdimValue = 1.0000000000000002;
    const eps = 1e-13;

    const frama = new FractalAdaptiveMovingAverage({ length: 16, slowestSmoothingFactor: 0.01 });

    for (let i = 0; i < lprimed; i++) {
      frama.update(0, 0, 0);
    }

    const s = new Scalar({ time, value: inp });
    const output = frama.updateScalar(s);
    const s0 = output[0] as Scalar;
    const s1 = output[1] as Scalar;

    expect(Math.abs(s0.value - expectedFramaValue))
      .withContext(`FRAMA: expected ${expectedFramaValue}, actual ${s0.value}`)
      .toBeLessThan(eps);
    expect(Math.abs(s1.value - expectedFdimValue))
      .withContext(`FDIM: expected ${expectedFdimValue}, actual ${s1.value}`)
      .toBeLessThan(eps);
  });
});
