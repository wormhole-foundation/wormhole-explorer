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
		  "emitter_chain": "10",
		  "destination_chain": "1",
		  "volume": 205496454333
		},
		{
		  "emitter_chain": "11",
		  "destination_chain": "1",
		  "volume": 16016186803
		},
		{
		  "emitter_chain": "12",
		  "destination_chain": "1",
		  "volume": 2503044910912
		},
		{
		  "emitter_chain": "13",
		  "destination_chain": "1",
		  "volume": 600755167
		},
		{
		  "emitter_chain": "14",
		  "destination_chain": "1",
		  "volume": 2363682498159
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "1",
		  "volume": 105739893510
		},
		{
		  "emitter_chain": "18",
		  "destination_chain": "1",
		  "volume": 99628422
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "1",
		  "volume": 19949470803943190
		},
		{
		  "emitter_chain": "22",
		  "destination_chain": "1",
		  "volume": 11220687344321
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "1",
		  "volume": 38768459035
		},
		{
		  "emitter_chain": "24",
		  "destination_chain": "1",
		  "volume": 2652106023
		},
		{
		  "emitter_chain": "3",
		  "destination_chain": "1",
		  "volume": 14751477259
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "1",
		  "volume": 159929740971571000
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "1",
		  "volume": 101446967968774080
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "1",
		  "volume": 859660722141
		},
		{
		  "emitter_chain": "7",
		  "destination_chain": "1",
		  "volume": 26067653497
		},
		{
		  "emitter_chain": "8",
		  "destination_chain": "1",
		  "volume": 4117447
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "10",
		  "volume": 76073828853
		},
		{
		  "emitter_chain": "14",
		  "destination_chain": "10",
		  "volume": 100000000
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "10",
		  "volume": 58842786942
		},
		{
		  "emitter_chain": "18",
		  "destination_chain": "10",
		  "volume": 24332691049
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "10",
		  "volume": 874220261252
		},
		{
		  "emitter_chain": "22",
		  "destination_chain": "10",
		  "volume": 188200001
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "10",
		  "volume": 1700823167
		},
		{
		  "emitter_chain": "24",
		  "destination_chain": "10",
		  "volume": 1900000
		},
		{
		  "emitter_chain": "3",
		  "destination_chain": "10",
		  "volume": 4578715120202
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "10",
		  "volume": 3204512800290791
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "10",
		  "volume": 5000054668435456
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "10",
		  "volume": 273201469849
		},
		{
		  "emitter_chain": "7",
		  "destination_chain": "10",
		  "volume": 100000000
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "11",
		  "volume": 4532468876
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "11",
		  "volume": 4319489200
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "11",
		  "volume": 15553306812
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "11",
		  "volume": 61318626
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "12",
		  "volume": 4822889517819
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "12",
		  "volume": 27403293970693
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "12",
		  "volume": 10000000
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "13",
		  "volume": 1252041403
		},
		{
		  "emitter_chain": "14",
		  "destination_chain": "13",
		  "volume": 836501883
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "13",
		  "volume": 2204380096
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "13",
		  "volume": 22814036575
		},
		{
		  "emitter_chain": "22",
		  "destination_chain": "13",
		  "volume": 2
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "13",
		  "volume": 526330757
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "13",
		  "volume": 1006962621380
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "13",
		  "volume": 12870838120
		},
		{
		  "emitter_chain": "7",
		  "destination_chain": "13",
		  "volume": 11708331
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "14",
		  "volume": 3274511881598
		},
		{
		  "emitter_chain": "13",
		  "destination_chain": "14",
		  "volume": 464828947
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "14",
		  "volume": 91461884884
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "14",
		  "volume": 52252378775061
		},
		{
		  "emitter_chain": "22",
		  "destination_chain": "14",
		  "volume": 118529092
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "14",
		  "volume": 6653424203713
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "14",
		  "volume": 1157701141796
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "14",
		  "volume": 19146352568
		},
		{
		  "emitter_chain": "7",
		  "destination_chain": "14",
		  "volume": 11467140597
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "15",
		  "volume": 461229201944
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "16",
		  "volume": 69909074486
		},
		{
		  "emitter_chain": "10",
		  "destination_chain": "16",
		  "volume": 58841786942
		},
		{
		  "emitter_chain": "11",
		  "destination_chain": "16",
		  "volume": 5000000000
		},
		{
		  "emitter_chain": "13",
		  "destination_chain": "16",
		  "volume": 684448517
		},
		{
		  "emitter_chain": "14",
		  "destination_chain": "16",
		  "volume": 40446287150
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "16",
		  "volume": 36156861905088
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "16",
		  "volume": 11864185359
		},
		{
		  "emitter_chain": "24",
		  "destination_chain": "16",
		  "volume": 207629459
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "16",
		  "volume": 75506631236235
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "16",
		  "volume": 26498142175
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "16",
		  "volume": 1901026572
		},
		{
		  "emitter_chain": "7",
		  "destination_chain": "16",
		  "volume": 1215747131
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "18",
		  "volume": 32628422
		},
		{
		  "emitter_chain": "10",
		  "destination_chain": "18",
		  "volume": 1646832637
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "18",
		  "volume": 12099963937
		},
		{
		  "emitter_chain": "3",
		  "destination_chain": "18",
		  "volume": 60327044882
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "18",
		  "volume": 151178186185
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "18",
		  "volume": 289762721
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "18",
		  "volume": 22750009930
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "19",
		  "volume": 18487525156
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "19",
		  "volume": 281893417794
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "19",
		  "volume": 1000000
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "2",
		  "volume": 2374163717301541
		},
		{
		  "emitter_chain": "10",
		  "destination_chain": "2",
		  "volume": 6756432340039027
		},
		{
		  "emitter_chain": "11",
		  "destination_chain": "2",
		  "volume": 21310015468
		},
		{
		  "emitter_chain": "12",
		  "destination_chain": "2",
		  "volume": 25345032496434
		},
		{
		  "emitter_chain": "13",
		  "destination_chain": "2",
		  "volume": 110383184114
		},
		{
		  "emitter_chain": "14",
		  "destination_chain": "2",
		  "volume": 6837250060449
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "2",
		  "volume": 37472983570922
		},
		{
		  "emitter_chain": "18",
		  "destination_chain": "2",
		  "volume": 102963937
		},
		{
		  "emitter_chain": "22",
		  "destination_chain": "2",
		  "volume": 17166876817275
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "2",
		  "volume": 50672151360
		},
		{
		  "emitter_chain": "24",
		  "destination_chain": "2",
		  "volume": 86412901878
		},
		{
		  "emitter_chain": "3",
		  "destination_chain": "2",
		  "volume": 258022587961601
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "2",
		  "volume": 4436612861467056
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "2",
		  "volume": 50122536443883
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "2",
		  "volume": 5195030135663
		},
		{
		  "emitter_chain": "7",
		  "destination_chain": "2",
		  "volume": 119778132583
		},
		{
		  "emitter_chain": "8",
		  "destination_chain": "2",
		  "volume": 671516470243
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "22",
		  "volume": 1001086910982504
		},
		{
		  "emitter_chain": "10",
		  "destination_chain": "22",
		  "volume": 1807608424
		},
		{
		  "emitter_chain": "14",
		  "destination_chain": "22",
		  "volume": 10686248055
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "22",
		  "volume": 6517227036
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "22",
		  "volume": 1916201163463
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "22",
		  "volume": 90054220347
		},
		{
		  "emitter_chain": "24",
		  "destination_chain": "22",
		  "volume": 11348300000
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "22",
		  "volume": 215211408022767
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "22",
		  "volume": 14198744164
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "22",
		  "volume": 125342432
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "23",
		  "volume": 1000043278530407
		},
		{
		  "emitter_chain": "10",
		  "destination_chain": "23",
		  "volume": 115970000
		},
		{
		  "emitter_chain": "11",
		  "destination_chain": "23",
		  "volume": 61318626
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "23",
		  "volume": 14362015815
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "23",
		  "volume": 16935321016
		},
		{
		  "emitter_chain": "22",
		  "destination_chain": "23",
		  "volume": 25706005635
		},
		{
		  "emitter_chain": "24",
		  "destination_chain": "23",
		  "volume": 12988800881
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "23",
		  "volume": 4694175586075199
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "23",
		  "volume": 2566232686905
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "23",
		  "volume": 4128467473
		},
		{
		  "emitter_chain": "7",
		  "destination_chain": "23",
		  "volume": 7193865
		},
		{
		  "emitter_chain": "8",
		  "destination_chain": "23",
		  "volume": 300000000
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "24",
		  "volume": 10023182356
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "24",
		  "volume": 861525445
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "24",
		  "volume": 100000991573911
		},
		{
		  "emitter_chain": "22",
		  "destination_chain": "24",
		  "volume": 11097300000
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "24",
		  "volume": 13976571168
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "24",
		  "volume": 13800986397
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "24",
		  "volume": 1714292385
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "3",
		  "volume": 935963033238
		},
		{
		  "emitter_chain": "10",
		  "destination_chain": "3",
		  "volume": 521965710750
		},
		{
		  "emitter_chain": "18",
		  "destination_chain": "3",
		  "volume": 59146527102
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "3",
		  "volume": 216874999700547
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "3",
		  "volume": 528367141721685
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "3",
		  "volume": 452423912553
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "3",
		  "volume": 50317489503
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "4",
		  "volume": 162896952707632300
		},
		{
		  "emitter_chain": "10",
		  "destination_chain": "4",
		  "volume": 8236696496674905000
		},
		{
		  "emitter_chain": "13",
		  "destination_chain": "4",
		  "volume": 950701148
		},
		{
		  "emitter_chain": "14",
		  "destination_chain": "4",
		  "volume": 24000083799857
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "4",
		  "volume": 99634352200953
		},
		{
		  "emitter_chain": "18",
		  "destination_chain": "4",
		  "volume": 150838334630
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "4",
		  "volume": 11483795247439652
		},
		{
		  "emitter_chain": "22",
		  "destination_chain": "4",
		  "volume": 179607449539439
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "4",
		  "volume": 129048434296329980
		},
		{
		  "emitter_chain": "24",
		  "destination_chain": "4",
		  "volume": 2294765969
		},
		{
		  "emitter_chain": "3",
		  "destination_chain": "4",
		  "volume": 223035687961636
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "4",
		  "volume": 66255509482314400
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "4",
		  "volume": 10000000090551583000
		},
		{
		  "emitter_chain": "7",
		  "destination_chain": "4",
		  "volume": 2822885325419
		},
		{
		  "emitter_chain": "8",
		  "destination_chain": "4",
		  "volume": 1015719330207
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "5",
		  "volume": 2300696448214956
		},
		{
		  "emitter_chain": "10",
		  "destination_chain": "5",
		  "volume": 304251425265
		},
		{
		  "emitter_chain": "13",
		  "destination_chain": "5",
		  "volume": 7309000000
		},
		{
		  "emitter_chain": "14",
		  "destination_chain": "5",
		  "volume": 1137181274145
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "5",
		  "volume": 23903566559
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "5",
		  "volume": 193481448550215650
		},
		{
		  "emitter_chain": "22",
		  "destination_chain": "5",
		  "volume": 15035908111
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "5",
		  "volume": 2543546365654
		},
		{
		  "emitter_chain": "24",
		  "destination_chain": "5",
		  "volume": 3986195946
		},
		{
		  "emitter_chain": "3",
		  "destination_chain": "5",
		  "volume": 43410000000
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "5",
		  "volume": 672853118202627800
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "5",
		  "volume": 1506652683979338
		},
		{
		  "emitter_chain": "7",
		  "destination_chain": "5",
		  "volume": 6478310128
		},
		{
		  "emitter_chain": "8",
		  "destination_chain": "5",
		  "volume": 1000444524372
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "6",
		  "volume": 795628681635
		},
		{
		  "emitter_chain": "10",
		  "destination_chain": "6",
		  "volume": 33472285898
		},
		{
		  "emitter_chain": "13",
		  "destination_chain": "6",
		  "volume": 10459669487
		},
		{
		  "emitter_chain": "14",
		  "destination_chain": "6",
		  "volume": 19106535388
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "6",
		  "volume": 2574592761
		},
		{
		  "emitter_chain": "18",
		  "destination_chain": "6",
		  "volume": 268092413
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "6",
		  "volume": 6860805241062
		},
		{
		  "emitter_chain": "22",
		  "destination_chain": "6",
		  "volume": 126604396
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "6",
		  "volume": 10198890933
		},
		{
		  "emitter_chain": "24",
		  "destination_chain": "6",
		  "volume": 32500000
		},
		{
		  "emitter_chain": "3",
		  "destination_chain": "6",
		  "volume": 480131130672
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "6",
		  "volume": 10001673447916442000
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "6",
		  "volume": 12386555756409
		},
		{
		  "emitter_chain": "7",
		  "destination_chain": "6",
		  "volume": 5775105385
		},
		{
		  "emitter_chain": "1",
		  "destination_chain": "7",
		  "volume": 11165996628
		},
		{
		  "emitter_chain": "13",
		  "destination_chain": "7",
		  "volume": 200000000
		},
		{
		  "emitter_chain": "16",
		  "destination_chain": "7",
		  "volume": 294624003
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "7",
		  "volume": 10738904165
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "7",
		  "volume": 7000000
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "7",
		  "volume": 6029227056104
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "7",
		  "volume": 1380550650
		},
		{
		  "emitter_chain": "6",
		  "destination_chain": "7",
		  "volume": 14022974
		},
		{
		  "emitter_chain": "2",
		  "destination_chain": "8",
		  "volume": 34056349060
		},
		{
		  "emitter_chain": "23",
		  "destination_chain": "8",
		  "volume": 300000000
		},
		{
		  "emitter_chain": "4",
		  "destination_chain": "8",
		  "volume": 1160630809972
		},
		{
		  "emitter_chain": "5",
		  "destination_chain": "8",
		  "volume": 1001456998315
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
