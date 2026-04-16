//nolint:testpackage
package cybercycle

//nolint: gofumpt
import (
	"math"
	"testing"
	"time"

	"zpano/entities"
	"zpano/indicators/core"
	"zpano/indicators/core/outputs"
)

func testCyberCycleTime() time.Time {
	return time.Date(2021, time.April, 1, 0, 0, 0, 0, &time.Location{})
}

// Input data taken from TA-Lib Excel simulation (TALib data), test_iTrend.xsl,
// (high + low)/2 median price, D3…D254, 252 entries.

func testCyberCycleInput() []float64 { //nolint:dupl,funlen
	return []float64{
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
	}
}

// Expected cycle (Value) values taken from Excel simulation, test_iTrend.xsl, L3…L254, 252 entries.

func testCyberCycleExpectedCycle() []float64 { //nolint:dupl,funlen
	return []float64{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		-0.40373412499999100, -0.96429470166668600, -1.31092291705417000, -0.68548947158259200, 1.22887159506654000,
		3.36359403162886000, 3.52770513750667000, 1.35106929655656000, -0.81205863600096300, -1.23401810588689000,
		-0.63570246475572100, -0.40339606424739500, -0.33086911773293200, -0.15942842801569900, -0.46666842618197000,
		-1.31484532322436000, -2.34269885230920000, -3.21244993690501000, -3.73050678069778000, -3.62212421375204000,
		-2.36872106670329000, -0.48266873284399200, 1.30149104916854000, 2.06182469474024000, 1.87278386545763000,
		1.67604916543704000, 1.83253367206192000, 2.56031385268201000, 3.22339863260551000, 3.12886049504493000,
		2.08346558510971000, 0.68640958780638800, -0.06198311374153140, -0.17312185863629800, 0.42730521509484600,
		1.69144084769429000, 3.18624269617584000, 4.51935772571629000, 5.16055491607647000, 5.45974946984687000,
		5.20908510866732000, 4.42083032940065000, 3.64950487928218000, 2.97420318398292000, 2.18155594378371000,
		0.93239087786088900, -0.62284932795728600, -2.01808204734577000, -2.75436185972954000, -2.42006418801425000,
		-1.19521901514309000, 0.27780509596404300, 1.57257899844036000, 2.50740254918312000, 2.67778436364620000,
		2.53304696201011000, 2.84903583030455000, 3.49405668109057000, 3.91569513928142000, 3.82699967933818000,
		3.34856839635451000, 2.54781524664315000, 1.51101064649924000, 0.40278091650025700, -0.78119460346670700,
		-2.32326118754582000, -3.79848670046355000, -4.49240553675381000, -2.14806869279784000, 2.89211037388440000,
		8.05123576199249000, 11.13760437993340000, 11.36422622154550000, 10.22682844178690000, 8.26912884062561000,
		6.73638512009549000, 6.31397592453717000, 5.94912457301851000, 5.16285807451561000, 4.29959793581197000,
		4.14068455571173000, 4.60687176894004000, 4.84199951174339000, 5.94044436555316000, 7.82131891305532000,
		8.89099291943263000, 8.12393184182649000, 6.04791259561332000, 4.20798438201175000, 2.31004766951251000,
		0.01232966079133350, -2.23965016445616000, -3.59944173992356000, -3.65516260485302000, -2.93348109666671000,
		-2.85715179661271000, -3.77587335369262000, -4.55934604689459000, -4.53095642944854000, -2.99424056281515000,
		-1.20968339767280000, -0.13676924855923600, -0.61971752750631100, -2.26515191974954000, -3.54526533952726000,
		-3.86913107362935000, -2.67602076312675000, -1.01017048716708000, 0.18879862689753700, 0.99205565038022500,
		1.80764695230356000, 2.25198762843744000, 1.68472588984627000, 0.67134424277851600, 0.04844376527334200,
		-0.18478802383738900, 0.12346492974419700, 1.19420153530787000, 2.59478693377018000, 3.47449422642475000,
		3.52999756296555000, 3.37135632734785000, 3.35808764749140000, 3.64522139514418000, 3.92847120531954000,
		3.64435876973416000, 2.70175433705798000, 1.45187785448478000, 0.31097606655357800, -1.63802746422093000,
		-4.41298774174640000, -7.26712801667701000, -9.22436424234942000, -9.95472184414596000, -9.90206893440348000,
		-8.62923364915530000, -7.07400810356329000, -6.11638622280663000, -6.03884572398178000, -6.79038137333397000,
		-7.75166493772937000, -8.25952026771340000, -7.27542356830480000, -5.46671600750159000, -4.40404811722615000,
		-4.24481980231917000, -4.13519292822477000, -3.28289440780554000, -1.78777931823005000, -0.07848322109686290,
		1.23109899943034000, 1.07263377686706000, -0.33614469963450300, -1.52966946993255000, -1.91889260086063000,
		-1.82319775888943000, -1.55689293355002000, -1.02249715640624000, -0.35119453352147300, 0.04644299989243790,
		0.59810571517601200, 1.20811172545373000, 1.94862405128820000, 2.91944202905113000, 4.11340712790928000,
		5.38277474281827000, 6.25018171754657000, 6.39164059873980000, 5.42495295031661000, 3.93680974207219000,
		2.39839408435878000, 0.65408745932243800, -1.20232161922217000, -2.47222960115456000, -3.05482156884890000,
		-3.66557200685371000, -4.70897507035050000, -5.26318767295749000, -5.12463263752143000, -4.91682799994891000,
		-4.98461260337940000, -5.58143000929650000, -6.05086785579535000, -5.62638189673881000, -4.35399515695675000,
		-3.46032341445017000, -4.33421980629210000, -5.80603491021203000, -6.52208803503231000, -6.97364050548440000,
		-7.98885908620154000, -8.97366814380807000, -8.56388183632729000, -7.07097720048917000, -5.00789640100369000,
		-5.81341331266382000, -9.13095428932659000, -12.04324415819120000, -11.75285393606370000, -8.80020346949228000,
		-6.69222820895413000, -4.57395344622415000, -2.37101187788581000, -0.74042144472835400, -0.30583400567795200,
		-0.83485307634880700, -1.40937014049790000, -1.45392807725866000, -0.66664001001782900, 1.03025916288786000,
		2.86146888346919000, 3.93305559827095000, 4.15226849630482000, 3.85590711618240000, 3.51705650531192000,
		3.66770741009400000, 5.11866954883057000, 8.19700684683454000, 11.48434529649540000, 12.55861305048750000,
		11.34492215196790000, 9.39530010862703000, 7.86247238697589000, 6.81870017999030000, 6.00358921728649000,
		6.63653086681263000, 8.77643072324039000, 11.41922401518750000, 13.13529594405150000, 13.03571481353350000,
		10.49401588282880000, 6.49357679983644000, 2.83260596897050000, 0.43846163227321900, -0.59141026653435400,
		-0.72806039484036000, -0.09614955321084380, 0.46017351652528900, 0.38655565597574900, 0.14250465400549400,
		0.03238539876344730, 0.01924277478399820, 0.06055084637442910, 0.05252433167902510, -0.28914014927285800,
	}
}

// Expected signal values taken from Excel simulation, test_iTrend.xsl, N3…N254, 252 entries.

func testCyberCycleExpectedSignal() []float64 { //nolint:dupl,funlen
	return []float64{
		math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(),
		-0.07250995437499910, -0.16168842910416800, -0.27661187789916800, -0.31749963726751100, -0.16286251403410500,
		0.18978314053219100, 0.52357534022963800, 0.60632473586233100, 0.46448639867600200, 0.29463594821971300,
		0.20160210692217000, 0.14110228980521300, 0.09390514905139860, 0.06857179134468880, 0.01504776959202290,
		-0.11794153968961600, -0.34041727095157400, -0.62762053754691800, -0.93790916186200400, -1.20633066705101000,
		-1.32256970701624000, -1.23857960959901000, -0.98457254372225700, -0.67993281987600800, -0.42466115134264400,
		-0.21459011966467500, -0.00987774049201537, 0.24714141882538700, 0.54476714020340000, 0.80317647568755200,
		0.93120538662976800, 0.90672580674743000, 0.80985491469853400, 0.71155723736505100, 0.68313203513803000,
		0.78396291639365600, 1.02419089437188000, 1.37370757750632000, 1.75239231136333000, 2.12312802721169000,
		2.43172373535725000, 2.63063439476159000, 2.73252144321365000, 2.75668961729058000, 2.69917624993989000,
		2.52249771273199000, 2.20796300866306000, 1.78535850306218000, 1.33138646678301000, 0.95624140130328000,
		0.74109535965864300, 0.69476633328918300, 0.78254759980430100, 0.95503309474218300, 1.12730822163258000,
		1.26788209567034000, 1.42599746913376000, 1.63280339032944000, 1.86109256522464000, 2.05768327663599000,
		2.18677178860784000, 2.22287613441137000, 2.15168958562016000, 1.97679871870817000, 1.70099938649068000,
		1.29857332908703000, 0.78886732613197500, 0.26074003984339700, 0.01985916657927310, 0.30708428730978600,
		1.08149943477806000, 2.08710992929359000, 3.01482155851878000, 3.73602224684560000, 4.18933290622360000,
		4.44403812761079000, 4.63103190730343000, 4.76284117387493000, 4.80284286393900000, 4.75251837112630000,
		4.69133498958484000, 4.68288866752036000, 4.69879975194267000, 4.82296421330372000, 5.12279968327888000,
		5.49961900689425000, 5.76205029038748000, 5.79063652091006000, 5.63237130702023000, 5.30013894326946000,
		4.77135801502165000, 4.07025719707386000, 3.30328730337412000, 2.60744231255141000, 2.05334997162960000,
		1.56229979480536000, 1.02848247995557000, 0.46969962727055000, -0.03036597840135870, -0.32675343684273800,
		-0.41504643292574400, -0.38721871448909300, -0.41046859579081500, -0.59593692818668700, -0.89086976932074400,
		-1.18869589975161000, -1.33742838608912000, -1.30470259619692000, -1.15535247388747000, -0.94061166146070200,
		-0.66578580008427600, -0.37400845723210400, -0.16813502252426700, -0.08418709599398860, -0.07092400986725560,
		-0.08231041126426890, -0.06173287716342230, 0.06386056408370650, 0.31695320105235400, 0.63270730358959300,
		0.92243632952719000, 1.16732832930926000, 1.38640426112747000, 1.61228597452914000, 1.84390449760818000,
		2.02394992482078000, 2.09173036604450000, 2.02774511488853000, 1.85606821005503000, 1.50665864262744000,
		0.91469400419005300, 0.09651180210334740, -0.83557580234192900, -1.74749040652233000, -2.56294825931045000,
		-3.16957679829493000, -3.56001992882177000, -3.81565655822025000, -4.03797547479641000, -4.31321606465016000,
		-4.65706095195808000, -5.01730688353362000, -5.24311855201073000, -5.26547829755982000, -5.17933527952645000,
		-5.08588373180572000, -4.99081465144763000, -4.82002262708342000, -4.51679829619808000, -4.07296678868796000,
		-3.54256020987613000, -3.08104081120181000, -2.80655120004508000, -2.67886302703383000, -2.60286598441651000,
		-2.52489916186380000, -2.42809853903242000, -2.28753840076980000, -2.09390401404497000, -1.87986931265123000,
		-1.63207180986851000, -1.34805345633628000, -1.01838570557383000, -0.62460293211133800, -0.15080192610927500,
		0.40255574078347900, 0.98731833845978800, 1.52775056448779000, 1.91747080307067000, 2.11940469697082000,
		2.14730363570962000, 1.99798201807090000, 1.67795165434159000, 1.26293352879198000, 0.83115801902789000,
		0.38148501643973000, -0.12756099223929300, -0.64112366031111200, -1.08947455803214000, -1.47220990222382000,
		-1.82345017233938000, -2.19924815603509000, -2.58441012601112000, -2.88860730308389000, -3.03514608847117000,
		-3.07766382106907000, -3.20331941959138000, -3.46359096865344000, -3.76944067529133000, -4.08986065831064000,
		-4.47976050109973000, -4.92915126537056000, -5.29262432246623000, -5.47045961026853000, -5.42420328934204000,
		-5.46312429167422000, -5.82990729143946000, -6.45124097811463000, -6.98140227390954000, -7.16328239346782000,
		-7.11617697501645000, -6.86195462213722000, -6.41286034771208000, -5.84561645741371000, -5.29163821224013000,
		-4.84595969865100000, -4.50230074283569000, -4.19746347627799000, -3.84438112965197000, -3.35691710039799000,
		-2.73507850201127000, -2.06826509198305000, -1.44621173315426000, -0.91599984822059500, -0.47269421286734400,
		-0.05865405057120970, 0.45907830936896800, 1.23287116311553000, 2.25801857645351000, 3.28807802385691000,
		4.09376243666801000, 4.62391620386391000, 4.94777182217511000, 5.13486465795663000, 5.22173711388962000,
		5.36321648918192000, 5.70453791258777000, 6.27600652284774000, 6.96193546496812000, 7.56931339982466000,
		7.86178364812507000, 7.72496296329621000, 7.23572726386364000, 6.55600070070460000, 5.84125960398070000,
		5.18432760409860000, 4.65627988836765000, 4.23666925118342000, 3.85165789166265000, 3.48074256789693000,
		3.13590685098359000, 2.82424044336363000, 2.54787148366471000, 2.29833676846614000, 2.03958907669224000,
	}
}

// Input high price taken from TA-Lib Excel simulation, test_iTrend.xsl, B3…B254, 252 entries.

func testCyberCycleInputHigh() []float64 { //nolint:dupl,funlen
	return []float64{
		93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000, 96.250000,
		99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000, 90.345000, 88.500000,
		88.250000, 85.500000, 84.440000, 84.750000, 84.440000, 89.405000, 88.125000, 89.125000, 87.155000, 87.250000, 87.375000,
		88.970000, 90.000000, 89.845000, 86.970000, 85.940000, 84.750000, 85.470000, 84.470000, 88.500000, 89.470000, 90.000000,
		92.440000, 91.440000, 92.970000, 91.720000, 91.155000, 91.750000, 90.000000, 88.875000, 89.000000, 85.250000, 83.815000,
		85.250000, 86.625000, 87.940000, 89.375000, 90.625000, 90.750000, 88.845000, 91.970000, 93.375000, 93.815000, 94.030000,
		94.030000, 91.815000, 92.000000, 91.940000, 89.750000, 88.750000, 86.155000, 84.875000, 85.940000, 99.375000, 103.280000,
		105.375000, 107.625000, 105.250000, 104.500000, 105.500000, 106.125000, 107.940000, 106.250000, 107.000000, 108.750000,
		110.940000, 110.940000, 114.220000, 123.000000, 121.750000, 119.815000, 120.315000, 119.375000, 118.190000, 116.690000,
		115.345000, 113.000000, 118.315000, 116.870000, 116.750000, 113.870000, 114.620000, 115.310000, 116.000000, 121.690000,
		119.870000, 120.870000, 116.750000, 116.500000, 116.000000, 118.310000, 121.500000, 122.000000, 121.440000, 125.750000,
		127.750000, 124.190000, 124.440000, 125.750000, 124.690000, 125.310000, 132.000000, 131.310000, 132.250000, 133.880000,
		133.500000, 135.500000, 137.440000, 138.690000, 139.190000, 138.500000, 138.130000, 137.500000, 138.880000, 132.130000,
		129.750000, 128.500000, 125.440000, 125.120000, 126.500000, 128.690000, 126.620000, 126.690000, 126.000000, 123.120000,
		121.870000, 124.000000, 127.000000, 124.440000, 122.500000, 123.750000, 123.810000, 124.500000, 127.870000, 128.560000,
		129.630000, 124.870000, 124.370000, 124.870000, 123.620000, 124.060000, 125.870000, 125.190000, 125.620000, 126.000000,
		128.500000, 126.750000, 129.750000, 132.690000, 133.940000, 136.500000, 137.690000, 135.560000, 133.560000, 135.000000,
		132.380000, 131.440000, 130.880000, 129.630000, 127.250000, 127.810000, 125.000000, 126.810000, 124.750000, 122.810000,
		122.250000, 121.060000, 120.000000, 123.250000, 122.750000, 119.190000, 115.060000, 116.690000, 114.870000, 110.870000,
		107.250000, 108.870000, 109.000000, 108.500000, 113.060000, 93.000000, 94.620000, 95.120000, 96.000000, 95.560000,
		95.310000, 99.000000, 98.810000, 96.810000, 95.940000, 94.440000, 92.940000, 93.940000, 95.500000, 97.060000, 97.500000,
		96.250000, 96.370000, 95.000000, 94.870000, 98.250000, 105.120000, 108.440000, 109.870000, 105.000000, 106.000000,
		104.940000, 104.500000, 104.440000, 106.310000, 112.870000, 116.500000, 119.190000, 121.000000, 122.120000, 111.940000,
		112.750000, 110.190000, 107.940000, 109.690000, 111.060000, 110.440000, 110.120000, 110.310000, 110.440000, 110.000000,
		110.750000, 110.500000, 110.500000, 109.500000,
	}
}

// Input low price taken from TA-Lib Excel simulation, test_iTrend.xsl, C3…C254, 252 entries.

func testCyberCycleInputLow() []float64 { //nolint:dupl,funlen
	return []float64{
		90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000, 92.750000,
		96.315000, 96.030000, 88.815000, 86.750000, 90.940000, 88.905000, 88.780000, 89.250000, 89.750000, 87.500000, 86.530000,
		84.625000, 82.280000, 81.565000, 80.875000, 81.250000, 84.065000, 85.595000, 85.970000, 84.405000, 85.095000, 85.500000,
		85.530000, 87.875000, 86.565000, 84.655000, 83.250000, 82.565000, 83.440000, 82.530000, 85.065000, 86.875000, 88.530000,
		89.280000, 90.125000, 90.750000, 89.000000, 88.565000, 90.095000, 89.000000, 86.470000, 84.000000, 83.315000, 82.000000,
		83.250000, 84.750000, 85.280000, 87.190000, 88.440000, 88.250000, 87.345000, 89.280000, 91.095000, 89.530000, 91.155000,
		92.000000, 90.530000, 89.970000, 88.815000, 86.750000, 85.065000, 82.030000, 81.500000, 82.565000, 96.345000, 96.470000,
		101.155000, 104.250000, 101.750000, 101.720000, 101.720000, 103.155000, 105.690000, 103.655000, 104.000000, 105.530000,
		108.530000, 108.750000, 107.750000, 117.000000, 118.000000, 116.000000, 118.500000, 116.530000, 116.250000, 114.595000,
		110.875000, 110.500000, 110.720000, 112.620000, 114.190000, 111.190000, 109.440000, 111.560000, 112.440000, 117.500000,
		116.060000, 116.560000, 113.310000, 112.560000, 114.000000, 114.750000, 118.870000, 119.000000, 119.750000, 122.620000,
		123.000000, 121.750000, 121.560000, 123.120000, 122.190000, 122.750000, 124.370000, 128.000000, 129.500000, 130.810000,
		130.630000, 132.130000, 133.880000, 135.380000, 135.750000, 136.190000, 134.500000, 135.380000, 133.690000, 126.060000,
		126.870000, 123.500000, 122.620000, 122.750000, 123.560000, 125.810000, 124.620000, 124.370000, 121.810000, 118.190000,
		118.060000, 117.560000, 121.000000, 121.120000, 118.940000, 119.810000, 121.000000, 122.000000, 124.500000, 126.560000,
		123.500000, 121.250000, 121.060000, 122.310000, 121.000000, 120.870000, 122.060000, 122.750000, 122.690000, 122.870000,
		125.500000, 124.250000, 128.000000, 128.380000, 130.690000, 131.630000, 134.380000, 132.000000, 131.940000, 131.940000,
		129.560000, 123.750000, 126.000000, 126.250000, 124.370000, 121.440000, 120.440000, 121.370000, 121.690000, 120.000000,
		119.620000, 115.500000, 116.750000, 119.060000, 119.060000, 115.060000, 111.060000, 113.120000, 110.000000, 105.000000,
		104.690000, 103.870000, 104.690000, 105.440000, 107.000000, 89.000000, 92.500000, 92.120000, 94.620000, 92.810000,
		94.250000, 96.250000, 96.370000, 93.690000, 93.500000, 90.000000, 90.190000, 90.500000, 92.120000, 94.120000, 94.870000,
		93.000000, 93.870000, 93.000000, 92.620000, 93.560000, 98.370000, 104.440000, 106.000000, 101.810000, 104.120000,
		103.370000, 102.120000, 102.250000, 103.370000, 107.940000, 112.500000, 115.440000, 115.500000, 112.250000, 107.560000,
		106.560000, 106.870000, 104.500000, 105.750000, 108.620000, 107.750000, 108.060000, 108.000000, 108.190000, 108.120000,
		109.060000, 108.750000, 108.560000, 106.620000,
	}
}

func TestCyberCycleUpdate(t *testing.T) { //nolint: funlen
	t.Parallel()

	check := func(index int, exp, act float64) {
		t.Helper()

		if math.Abs(exp-act) > 1e-8 {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	checkNaN := func(index int, act float64) {
		t.Helper()

		if !math.IsNaN(act) {
			t.Errorf("[%v] is incorrect: expected NaN, actual %v", index, act)
		}
	}

	input := testCyberCycleInput()

	const lprimed = 7 // First 7 values are NaN (primed on sample 8).

	t.Run("reference implementation: cycle value from test_iTrend.xls", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()
		exp := testCyberCycleExpectedCycle()

		for i := range lprimed {
			checkNaN(i, cc.Update(input[i]))
		}

		for i := lprimed; i < len(input); i++ {
			act := cc.Update(input[i])
			check(i, exp[i], act)
		}

		checkNaN(0, cc.Update(math.NaN()))
	})

	t.Run("reference implementation: signal from test_iTrend.xls", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()
		expSig := testCyberCycleExpectedSignal()

		for i := range lprimed {
			cc.Update(input[i])
		}

		for i := lprimed; i < len(input); i++ {
			cc.Update(input[i])
			act := cc.signal
			check(i, expSig[i], act)
		}
	})
}

func TestCyberCycleUpdateEntity(t *testing.T) { //nolint: funlen,cyclop
	t.Parallel()

	const lprimed = 7

	time := testCyberCycleTime()

	input := testCyberCycleInput()
	inputHigh := testCyberCycleInputHigh()
	inputLow := testCyberCycleInputLow()
	expCycle := testCyberCycleExpectedCycle()
	expSignal := testCyberCycleExpectedSignal()

	check := func(index int, expValue, expSig float64, act core.Output) {
		t.Helper()

		const outputLen = 2

		if len(act) != outputLen {
			t.Errorf("[%v] len(output) is incorrect: expected %v, actual %v", index, outputLen, len(act))
		}

		s0, ok := act[0].(entities.Scalar)
		if !ok {
			t.Errorf("[%v] output[0] is not a scalar", index)
		}

		s1, ok := act[1].(entities.Scalar)
		if !ok {
			t.Errorf("[%v] output[1] is not a scalar", index)
		}

		if s0.Time != time {
			t.Errorf("[%v] output[0] time is incorrect: expected %v, actual %v", index, time, s0.Time)
		}

		if s1.Time != time {
			t.Errorf("[%v] output[1] time is incorrect: expected %v, actual %v", index, time, s1.Time)
		}

		if math.IsNaN(expValue) {
			if !math.IsNaN(s0.Value) {
				t.Errorf("[%v] output[0] value: expected NaN, actual %v", index, s0.Value)
			}

			if !math.IsNaN(s1.Value) {
				t.Errorf("[%v] output[1] value: expected NaN, actual %v", index, s1.Value)
			}

			return
		}

		if math.Abs(expValue-s0.Value) > 1e-8 {
			t.Errorf("[%v] output[0] value: expected %v, actual %v", index, expValue, s0.Value)
		}

		if math.Abs(expSig-s1.Value) > 1e-8 {
			t.Errorf("[%v] output[1] value: expected %v, actual %v", index, expSig, s1.Value)
		}
	}

	t.Run("update scalar", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()

		for i := 0; i < len(input); i++ {
			s := entities.Scalar{Time: time, Value: input[i]}
			act := cc.UpdateScalar(&s)
			check(i, expCycle[i], expSignal[i], act)
		}
	})

	t.Run("update bar", func(t *testing.T) {
		t.Parallel()

		// Default bar component for CyberCycle is BarMedianPrice = (High+Low)/2.
		cc := testCyberCycleCreateDefault()

		for i := 0; i < len(input); i++ {
			b := entities.Bar{Time: time, High: inputHigh[i], Low: inputLow[i]}
			act := cc.UpdateBar(&b)
			check(i, expCycle[i], expSignal[i], act)
		}
	})

	t.Run("update quote", func(t *testing.T) {
		t.Parallel()

		// Use QuoteMidPrice = (Ask+Bid)/2, feeding high/low as ask/bid.
		cc := testCyberCycleCreateDefault()

		for i := range lprimed {
			q := entities.Quote{Time: time, Ask: inputHigh[i], Bid: inputLow[i]}
			act := cc.UpdateQuote(&q)
			check(i, expCycle[i], expSignal[i], act)
		}

		for i := lprimed; i < len(input); i++ {
			q := entities.Quote{Time: time, Ask: inputHigh[i], Bid: inputLow[i]}
			act := cc.UpdateQuote(&q)
			check(i, expCycle[i], expSignal[i], act)
		}
	})

	t.Run("update trade", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()

		for i := 0; i < len(input); i++ {
			r := entities.Trade{Time: time, Price: input[i]}
			act := cc.UpdateTrade(&r)
			check(i, expCycle[i], expSignal[i], act)
		}
	})
}

func TestCyberCycleIsPrimed(t *testing.T) {
	t.Parallel()

	check := func(index int, exp, act bool) {
		t.Helper()

		if exp != act {
			t.Errorf("[%v] is incorrect: expected %v, actual %v", index, exp, act)
		}
	}

	const lprimed = 7

	t.Run("default params", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()
		input := testCyberCycleInput()

		check(0, false, cc.IsPrimed())

		for i := range lprimed {
			cc.Update(input[i])
			check(i+1, false, cc.IsPrimed())
		}

		for i := lprimed; i < len(input); i++ {
			cc.Update(input[i])
			check(i+1, true, cc.IsPrimed())
		}
	})
}

func TestCyberCycleMetadata(t *testing.T) {
	t.Parallel()

	check := func(what string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", what, exp, act)
		}
	}

	t.Run("default params (smoothing factor 0.07)", func(t *testing.T) {
		t.Parallel()

		cc := testCyberCycleCreateDefault()
		act := cc.Metadata()

		mn := "cc(28, hl/2)"
		mnSig := "ccSignal(28, hl/2)"
		descr := "Cyber Cycle "
		descrSig := "Cyber Cycle signal "

		check("Type", core.CyberCycle, act.Type)
		check("Mnemonic", mn, act.Mnemonic)
		check("Description", descr+mn, act.Description)
		check("len(Outputs)", 2, len(act.Outputs))

		check("Outputs[0].Kind", int(Value), act.Outputs[0].Kind)
		check("Outputs[0].Type", outputs.ScalarType, act.Outputs[0].Type)
		check("Outputs[0].Mnemonic", mn, act.Outputs[0].Mnemonic)
		check("Outputs[0].Description", descr+mn, act.Outputs[0].Description)

		check("Outputs[1].Kind", int(Signal), act.Outputs[1].Kind)
		check("Outputs[1].Type", outputs.ScalarType, act.Outputs[1].Type)
		check("Outputs[1].Mnemonic", mnSig, act.Outputs[1].Mnemonic)
		check("Outputs[1].Description", descrSig+mnSig, act.Outputs[1].Description)
	})

	t.Run("length-based with non-default trade component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:         3,
			SignalLag:      2,
			TradeComponent: entities.TradeVolume,
		}

		cc, _ := NewCyberCycleLength(&params)
		act := cc.Metadata()

		check("Mnemonic", "cc(3, hl/2, v)", act.Mnemonic)
		check("Description", "Cyber Cycle cc(3, hl/2, v)", act.Description)
	})
}

func TestNewCyberCycleLength(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		errLength    = "invalid cyber cycle parameters: length should be a positive integer"
		errSignalLag = "invalid cyber cycle parameters: signal lag should be a positive integer"
		errbc        = "invalid cyber cycle parameters: 9999: unknown bar component"
		errqc        = "invalid cyber cycle parameters: 9999: unknown quote component"
		errtc        = "invalid cyber cycle parameters: 9999: unknown trade component"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	checkInstance := func(cc *CyberCycle,
		length int, signalLag int,
	) {
		t.Helper()

		check("length", length, cc.length)
		check("signalLag", signalLag, cc.signalLag)
		check("primed", false, cc.primed)
		check("value is NaN", true, math.IsNaN(cc.value))
		check("signal is NaN", true, math.IsNaN(cc.signal))
		check("barFunc == nil", false, cc.barFunc == nil)
		check("quoteFunc == nil", false, cc.quoteFunc == nil)
		check("tradeFunc == nil", false, cc.tradeFunc == nil)
	}

	t.Run("length=28, signalLag=14", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 28, SignalLag: 14}
		cc, err := NewCyberCycleLength(&params)

		check("err == nil", true, err == nil)
		checkInstance(cc, 28, 14)
	})

	t.Run("length=1, signalLag=1", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 1, SignalLag: 1}
		cc, err := NewCyberCycleLength(&params)

		check("err == nil", true, err == nil)
		checkInstance(cc, 1, 1)
	})

	t.Run("length=0, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 0, SignalLag: 1}
		cc, err := NewCyberCycleLength(&params)

		check("cc == nil", true, cc == nil)
		check("err", errLength, err.Error())
	})

	t.Run("length=-8, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: -8, SignalLag: 1}
		cc, err := NewCyberCycleLength(&params)

		check("cc == nil", true, cc == nil)
		check("err", errLength, err.Error())
	})

	t.Run("signalLag=0, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 1, SignalLag: 0}
		cc, err := NewCyberCycleLength(&params)

		check("cc == nil", true, cc == nil)
		check("err", errSignalLag, err.Error())
	})

	t.Run("signalLag=-8, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: 1, SignalLag: -8}
		cc, err := NewCyberCycleLength(&params)

		check("cc == nil", true, cc == nil)
		check("err", errSignalLag, err.Error())
	})

	t.Run("length=-8, signalLag=-9, error", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{Length: -8, SignalLag: -9}
		cc, err := NewCyberCycleLength(&params)

		check("cc == nil", true, cc == nil)
		check("err", errLength, err.Error())
	})

	t.Run("invalid bar component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:       10,
			SignalLag:    9,
			BarComponent: entities.BarComponent(9999),
		}

		cc, err := NewCyberCycleLength(&params)
		check("cc == nil", true, cc == nil)
		check("err", errbc, err.Error())
	})

	t.Run("invalid quote component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:         10,
			SignalLag:      9,
			QuoteComponent: entities.QuoteComponent(9999),
		}

		cc, err := NewCyberCycleLength(&params)
		check("cc == nil", true, cc == nil)
		check("err", errqc, err.Error())
	})

	t.Run("invalid trade component", func(t *testing.T) {
		t.Parallel()

		params := LengthParams{
			Length:         10,
			SignalLag:      9,
			TradeComponent: entities.TradeComponent(9999),
		}

		cc, err := NewCyberCycleLength(&params)
		check("cc == nil", true, cc == nil)
		check("err", errtc, err.Error())
	})
}

func TestNewCyberCycleSmoothingFactor(t *testing.T) { //nolint: funlen
	t.Parallel()

	const (
		errAlpha     = "invalid cyber cycle parameters: smoothing factor should be in range [0, 1]"
		errSignalLag = "invalid cyber cycle parameters: signal lag should be a positive integer"
	)

	check := func(name string, exp, act any) {
		t.Helper()

		if exp != act {
			t.Errorf("%s is incorrect: expected %v, actual %v", name, exp, act)
		}
	}

	t.Run("default (alpha=0.07, signalLag=9)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.07, SignalLag: 9}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.07, cc.smoothingFactor)
		check("length", 28, cc.length)
		check("signalLag", 9, cc.signalLag)
	})

	t.Run("alpha=0.06, signalLag=11", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.06, SignalLag: 11}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.06, cc.smoothingFactor)
		check("length", 32, cc.length)
		check("signalLag", 11, cc.signalLag)
	})

	t.Run("near-zero alpha (epsilon case)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.000000001, SignalLag: 9}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.000000001, cc.smoothingFactor)
		check("length", math.MaxInt, cc.length)
	})

	t.Run("alpha=0 (boundary)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0, SignalLag: 9}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 0.0, cc.smoothingFactor)
		check("length", math.MaxInt, cc.length)
	})

	t.Run("alpha=1 (boundary)", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 1, SignalLag: 9}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("err == nil", true, err == nil)
		check("smoothingFactor", 1.0, cc.smoothingFactor)
		check("length", 1, cc.length)
	})

	t.Run("alpha=-0.0001, error", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: -0.0001, SignalLag: 8}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("cc == nil", true, cc == nil)
		check("err", errAlpha, err.Error())
	})

	t.Run("alpha=1.0001, error", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 1.0001, SignalLag: 8}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("cc == nil", true, cc == nil)
		check("err", errAlpha, err.Error())
	})

	t.Run("signalLag=0, error", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.07, SignalLag: 0}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("cc == nil", true, cc == nil)
		check("err", errSignalLag, err.Error())
	})

	t.Run("signalLag=-8, error", func(t *testing.T) {
		t.Parallel()

		params := SmoothingFactorParams{SmoothingFactor: 0.07, SignalLag: -8}
		cc, err := NewCyberCycleSmoothingFactor(&params)

		check("cc == nil", true, cc == nil)
		check("err", errSignalLag, err.Error())
	})
}

// testCyberCycleCreateDefault creates a CyberCycle with the default parameters
// matching the MBST source: smoothingFactor=0.07, signalLag=9.
func testCyberCycleCreateDefault() *CyberCycle {
	params := SmoothingFactorParams{SmoothingFactor: 0.07, SignalLag: 9}

	cc, _ := NewCyberCycleSmoothingFactor(&params)

	return cc
}
