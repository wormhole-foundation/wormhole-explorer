package transactions

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	"go.uber.org/zap"
)

// Test_Address_ShortString runs several test cases on the method `Address.ShortString()`.
func Test_Controller_createChainActivityResponse(t *testing.T) {

	activityJSON := `
	[
		{
		  "ChainSourceID": "10",
		  "ChainDestinationID": "1",
		  "Volume": 205496454333
		},
		{
		  "ChainSourceID": "11",
		  "ChainDestinationID": "1",
		  "Volume": 16016186803
		},
		{
		  "ChainSourceID": "12",
		  "ChainDestinationID": "1",
		  "Volume": 2503044910912
		},
		{
		  "ChainSourceID": "13",
		  "ChainDestinationID": "1",
		  "Volume": 600755167
		},
		{
		  "ChainSourceID": "14",
		  "ChainDestinationID": "1",
		  "Volume": 2363682498159
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "1",
		  "Volume": 105739893510
		},
		{
		  "ChainSourceID": "18",
		  "ChainDestinationID": "1",
		  "Volume": 99628422
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "1",
		  "Volume": 19949470803943190
		},
		{
		  "ChainSourceID": "22",
		  "ChainDestinationID": "1",
		  "Volume": 11220687344321
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "1",
		  "Volume": 38768459035
		},
		{
		  "ChainSourceID": "24",
		  "ChainDestinationID": "1",
		  "Volume": 2652106023
		},
		{
		  "ChainSourceID": "3",
		  "ChainDestinationID": "1",
		  "Volume": 14751477259
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "1",
		  "Volume": 159929740971571000
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "1",
		  "Volume": 101446967968774080
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "1",
		  "Volume": 859660722141
		},
		{
		  "ChainSourceID": "7",
		  "ChainDestinationID": "1",
		  "Volume": 26067653497
		},
		{
		  "ChainSourceID": "8",
		  "ChainDestinationID": "1",
		  "Volume": 4117447
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "10",
		  "Volume": 76073828853
		},
		{
		  "ChainSourceID": "14",
		  "ChainDestinationID": "10",
		  "Volume": 100000000
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "10",
		  "Volume": 58842786942
		},
		{
		  "ChainSourceID": "18",
		  "ChainDestinationID": "10",
		  "Volume": 24332691049
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "10",
		  "Volume": 874220261252
		},
		{
		  "ChainSourceID": "22",
		  "ChainDestinationID": "10",
		  "Volume": 188200001
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "10",
		  "Volume": 1700823167
		},
		{
		  "ChainSourceID": "24",
		  "ChainDestinationID": "10",
		  "Volume": 1900000
		},
		{
		  "ChainSourceID": "3",
		  "ChainDestinationID": "10",
		  "Volume": 4578715120202
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "10",
		  "Volume": 3204512800290791
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "10",
		  "Volume": 5000054668435456
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "10",
		  "Volume": 273201469849
		},
		{
		  "ChainSourceID": "7",
		  "ChainDestinationID": "10",
		  "Volume": 100000000
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "11",
		  "Volume": 4532468876
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "11",
		  "Volume": 4319489200
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "11",
		  "Volume": 15553306812
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "11",
		  "Volume": 61318626
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "12",
		  "Volume": 4822889517819
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "12",
		  "Volume": 27403293970693
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "12",
		  "Volume": 10000000
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "13",
		  "Volume": 1252041403
		},
		{
		  "ChainSourceID": "14",
		  "ChainDestinationID": "13",
		  "Volume": 836501883
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "13",
		  "Volume": 2204380096
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "13",
		  "Volume": 22814036575
		},
		{
		  "ChainSourceID": "22",
		  "ChainDestinationID": "13",
		  "Volume": 2
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "13",
		  "Volume": 526330757
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "13",
		  "Volume": 1006962621380
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "13",
		  "Volume": 12870838120
		},
		{
		  "ChainSourceID": "7",
		  "ChainDestinationID": "13",
		  "Volume": 11708331
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "14",
		  "Volume": 3274511881598
		},
		{
		  "ChainSourceID": "13",
		  "ChainDestinationID": "14",
		  "Volume": 464828947
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "14",
		  "Volume": 91461884884
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "14",
		  "Volume": 52252378775061
		},
		{
		  "ChainSourceID": "22",
		  "ChainDestinationID": "14",
		  "Volume": 118529092
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "14",
		  "Volume": 6653424203713
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "14",
		  "Volume": 1157701141796
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "14",
		  "Volume": 19146352568
		},
		{
		  "ChainSourceID": "7",
		  "ChainDestinationID": "14",
		  "Volume": 11467140597
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "15",
		  "Volume": 461229201944
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "16",
		  "Volume": 69909074486
		},
		{
		  "ChainSourceID": "10",
		  "ChainDestinationID": "16",
		  "Volume": 58841786942
		},
		{
		  "ChainSourceID": "11",
		  "ChainDestinationID": "16",
		  "Volume": 5000000000
		},
		{
		  "ChainSourceID": "13",
		  "ChainDestinationID": "16",
		  "Volume": 684448517
		},
		{
		  "ChainSourceID": "14",
		  "ChainDestinationID": "16",
		  "Volume": 40446287150
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "16",
		  "Volume": 36156861905088
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "16",
		  "Volume": 11864185359
		},
		{
		  "ChainSourceID": "24",
		  "ChainDestinationID": "16",
		  "Volume": 207629459
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "16",
		  "Volume": 75506631236235
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "16",
		  "Volume": 26498142175
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "16",
		  "Volume": 1901026572
		},
		{
		  "ChainSourceID": "7",
		  "ChainDestinationID": "16",
		  "Volume": 1215747131
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "18",
		  "Volume": 32628422
		},
		{
		  "ChainSourceID": "10",
		  "ChainDestinationID": "18",
		  "Volume": 1646832637
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "18",
		  "Volume": 12099963937
		},
		{
		  "ChainSourceID": "3",
		  "ChainDestinationID": "18",
		  "Volume": 60327044882
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "18",
		  "Volume": 151178186185
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "18",
		  "Volume": 289762721
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "18",
		  "Volume": 22750009930
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "19",
		  "Volume": 18487525156
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "19",
		  "Volume": 281893417794
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "19",
		  "Volume": 1000000
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "2",
		  "Volume": 2374163717301541
		},
		{
		  "ChainSourceID": "10",
		  "ChainDestinationID": "2",
		  "Volume": 6756432340039027
		},
		{
		  "ChainSourceID": "11",
		  "ChainDestinationID": "2",
		  "Volume": 21310015468
		},
		{
		  "ChainSourceID": "12",
		  "ChainDestinationID": "2",
		  "Volume": 25345032496434
		},
		{
		  "ChainSourceID": "13",
		  "ChainDestinationID": "2",
		  "Volume": 110383184114
		},
		{
		  "ChainSourceID": "14",
		  "ChainDestinationID": "2",
		  "Volume": 6837250060449
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "2",
		  "Volume": 37472983570922
		},
		{
		  "ChainSourceID": "18",
		  "ChainDestinationID": "2",
		  "Volume": 102963937
		},
		{
		  "ChainSourceID": "22",
		  "ChainDestinationID": "2",
		  "Volume": 17166876817275
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "2",
		  "Volume": 50672151360
		},
		{
		  "ChainSourceID": "24",
		  "ChainDestinationID": "2",
		  "Volume": 86412901878
		},
		{
		  "ChainSourceID": "3",
		  "ChainDestinationID": "2",
		  "Volume": 258022587961601
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "2",
		  "Volume": 4436612861467056
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "2",
		  "Volume": 50122536443883
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "2",
		  "Volume": 5195030135663
		},
		{
		  "ChainSourceID": "7",
		  "ChainDestinationID": "2",
		  "Volume": 119778132583
		},
		{
		  "ChainSourceID": "8",
		  "ChainDestinationID": "2",
		  "Volume": 671516470243
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "22",
		  "Volume": 1001086910982504
		},
		{
		  "ChainSourceID": "10",
		  "ChainDestinationID": "22",
		  "Volume": 1807608424
		},
		{
		  "ChainSourceID": "14",
		  "ChainDestinationID": "22",
		  "Volume": 10686248055
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "22",
		  "Volume": 6517227036
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "22",
		  "Volume": 1916201163463
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "22",
		  "Volume": 90054220347
		},
		{
		  "ChainSourceID": "24",
		  "ChainDestinationID": "22",
		  "Volume": 11348300000
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "22",
		  "Volume": 215211408022767
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "22",
		  "Volume": 14198744164
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "22",
		  "Volume": 125342432
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "23",
		  "Volume": 1000043278530407
		},
		{
		  "ChainSourceID": "10",
		  "ChainDestinationID": "23",
		  "Volume": 115970000
		},
		{
		  "ChainSourceID": "11",
		  "ChainDestinationID": "23",
		  "Volume": 61318626
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "23",
		  "Volume": 14362015815
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "23",
		  "Volume": 16935321016
		},
		{
		  "ChainSourceID": "22",
		  "ChainDestinationID": "23",
		  "Volume": 25706005635
		},
		{
		  "ChainSourceID": "24",
		  "ChainDestinationID": "23",
		  "Volume": 12988800881
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "23",
		  "Volume": 4694175586075199
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "23",
		  "Volume": 2566232686905
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "23",
		  "Volume": 4128467473
		},
		{
		  "ChainSourceID": "7",
		  "ChainDestinationID": "23",
		  "Volume": 7193865
		},
		{
		  "ChainSourceID": "8",
		  "ChainDestinationID": "23",
		  "Volume": 300000000
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "24",
		  "Volume": 10023182356
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "24",
		  "Volume": 861525445
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "24",
		  "Volume": 100000991573911
		},
		{
		  "ChainSourceID": "22",
		  "ChainDestinationID": "24",
		  "Volume": 11097300000
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "24",
		  "Volume": 13976571168
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "24",
		  "Volume": 13800986397
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "24",
		  "Volume": 1714292385
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "3",
		  "Volume": 935963033238
		},
		{
		  "ChainSourceID": "10",
		  "ChainDestinationID": "3",
		  "Volume": 521965710750
		},
		{
		  "ChainSourceID": "18",
		  "ChainDestinationID": "3",
		  "Volume": 59146527102
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "3",
		  "Volume": 216874999700547
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "3",
		  "Volume": 528367141721685
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "3",
		  "Volume": 452423912553
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "3",
		  "Volume": 50317489503
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "4",
		  "Volume": 162896952707632300
		},
		{
		  "ChainSourceID": "10",
		  "ChainDestinationID": "4",
		  "Volume": 8236696496674905000
		},
		{
		  "ChainSourceID": "13",
		  "ChainDestinationID": "4",
		  "Volume": 950701148
		},
		{
		  "ChainSourceID": "14",
		  "ChainDestinationID": "4",
		  "Volume": 24000083799857
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "4",
		  "Volume": 99634352200953
		},
		{
		  "ChainSourceID": "18",
		  "ChainDestinationID": "4",
		  "Volume": 150838334630
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "4",
		  "Volume": 11483795247439652
		},
		{
		  "ChainSourceID": "22",
		  "ChainDestinationID": "4",
		  "Volume": 179607449539439
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "4",
		  "Volume": 129048434296329980
		},
		{
		  "ChainSourceID": "24",
		  "ChainDestinationID": "4",
		  "Volume": 2294765969
		},
		{
		  "ChainSourceID": "3",
		  "ChainDestinationID": "4",
		  "Volume": 223035687961636
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "4",
		  "Volume": 66255509482314400
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "4",
		  "Volume": 10000000090551583000
		},
		{
		  "ChainSourceID": "7",
		  "ChainDestinationID": "4",
		  "Volume": 2822885325419
		},
		{
		  "ChainSourceID": "8",
		  "ChainDestinationID": "4",
		  "Volume": 1015719330207
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "5",
		  "Volume": 2300696448214956
		},
		{
		  "ChainSourceID": "10",
		  "ChainDestinationID": "5",
		  "Volume": 304251425265
		},
		{
		  "ChainSourceID": "13",
		  "ChainDestinationID": "5",
		  "Volume": 7309000000
		},
		{
		  "ChainSourceID": "14",
		  "ChainDestinationID": "5",
		  "Volume": 1137181274145
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "5",
		  "Volume": 23903566559
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "5",
		  "Volume": 193481448550215650
		},
		{
		  "ChainSourceID": "22",
		  "ChainDestinationID": "5",
		  "Volume": 15035908111
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "5",
		  "Volume": 2543546365654
		},
		{
		  "ChainSourceID": "24",
		  "ChainDestinationID": "5",
		  "Volume": 3986195946
		},
		{
		  "ChainSourceID": "3",
		  "ChainDestinationID": "5",
		  "Volume": 43410000000
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "5",
		  "Volume": 672853118202627800
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "5",
		  "Volume": 1506652683979338
		},
		{
		  "ChainSourceID": "7",
		  "ChainDestinationID": "5",
		  "Volume": 6478310128
		},
		{
		  "ChainSourceID": "8",
		  "ChainDestinationID": "5",
		  "Volume": 1000444524372
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "6",
		  "Volume": 795628681635
		},
		{
		  "ChainSourceID": "10",
		  "ChainDestinationID": "6",
		  "Volume": 33472285898
		},
		{
		  "ChainSourceID": "13",
		  "ChainDestinationID": "6",
		  "Volume": 10459669487
		},
		{
		  "ChainSourceID": "14",
		  "ChainDestinationID": "6",
		  "Volume": 19106535388
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "6",
		  "Volume": 2574592761
		},
		{
		  "ChainSourceID": "18",
		  "ChainDestinationID": "6",
		  "Volume": 268092413
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "6",
		  "Volume": 6860805241062
		},
		{
		  "ChainSourceID": "22",
		  "ChainDestinationID": "6",
		  "Volume": 126604396
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "6",
		  "Volume": 10198890933
		},
		{
		  "ChainSourceID": "24",
		  "ChainDestinationID": "6",
		  "Volume": 32500000
		},
		{
		  "ChainSourceID": "3",
		  "ChainDestinationID": "6",
		  "Volume": 480131130672
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "6",
		  "Volume": 10001673447916442000
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "6",
		  "Volume": 12386555756409
		},
		{
		  "ChainSourceID": "7",
		  "ChainDestinationID": "6",
		  "Volume": 5775105385
		},
		{
		  "ChainSourceID": "1",
		  "ChainDestinationID": "7",
		  "Volume": 11165996628
		},
		{
		  "ChainSourceID": "13",
		  "ChainDestinationID": "7",
		  "Volume": 200000000
		},
		{
		  "ChainSourceID": "16",
		  "ChainDestinationID": "7",
		  "Volume": 294624003
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "7",
		  "Volume": 10738904165
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "7",
		  "Volume": 7000000
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "7",
		  "Volume": 6029227056104
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "7",
		  "Volume": 1380550650
		},
		{
		  "ChainSourceID": "6",
		  "ChainDestinationID": "7",
		  "Volume": 14022974
		},
		{
		  "ChainSourceID": "2",
		  "ChainDestinationID": "8",
		  "Volume": 34056349060
		},
		{
		  "ChainSourceID": "23",
		  "ChainDestinationID": "8",
		  "Volume": 300000000
		},
		{
		  "ChainSourceID": "4",
		  "ChainDestinationID": "8",
		  "Volume": 1160630809972
		},
		{
		  "ChainSourceID": "5",
		  "ChainDestinationID": "8",
		  "Volume": 1001456998315
		}
	  ]
	`

	var activity []transactions.ChainActivityResult

	err := json.Unmarshal([]byte(activityJSON), &activity)
	assert.NoError(t, err)

	controller := NewController(nil, zap.NewExample())
	result, err := controller.createChainActivityResponse(activity, false)
	assert.NoError(t, err)

	totalPercentage := float64(0)
	for _, r := range result {
		totalPercentage += r.Percentage
		totalPercentageByChain := float64(0)
		for _, c := range r.Destinations {
			totalPercentageByChain += c.Percentage
		}
		assert.Equal(t, 100, int(math.Round(totalPercentageByChain)))
	}
	assert.Equal(t, 100, int(math.Round(totalPercentage)))
}
