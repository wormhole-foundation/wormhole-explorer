package operations_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/operations"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
)

var sortStage = bson.D{{"$sort", bson.D{{"timestamp", -1}, {"_id", -1}}}}
var skipStage = bson.D{{"$skip", int64(0)}}
var limitStage = bson.D{{"$limit", int64(0)}}
var lookupVaasStage = bson.D{
	{"$lookup", bson.D{
		{"from", "vaas"},
		{"localField", "_id"},
		{"foreignField", "_id"},
		{"as", "vaas"}}}}
var lookupTransferPricesStage = bson.D{{"$lookup", bson.D{
	{"from", "transferPrices"},
	{"localField", "_id"},
	{"foreignField", "_id"},
	{"as", "transferPrices"},
}}}
var lookupGlobalTransactionsStage = bson.D{{"$lookup", bson.D{
	{"from", "globalTransactions"},
	{"localField", "_id"},
	{"foreignField", "_id"},
	{"as", "globalTransactions"},
}}}
var addFieldsStage = bson.D{{"$addFields", bson.D{
	{"payload", "$parsedPayload"},
	{"vaa", bson.D{{"$arrayElemAt", bson.A{"$vaas", 0}}}},
	{"symbol", bson.D{{"$arrayElemAt", bson.A{"$transferPrices.symbol", 0}}}},
	{"usdAmount", bson.D{{"$arrayElemAt", bson.A{"$transferPrices.usdAmount", 0}}}},
	{"tokenAmount", bson.D{{"$arrayElemAt", bson.A{"$transferPrices.tokenAmount", 0}}}},
	{"originTx", bson.D{{"$arrayElemAt", bson.A{"$globalTransactions.originTx", 0}}}},
	{"destinationTx", bson.D{{"$arrayElemAt", bson.A{"$globalTransactions.destinationTx", 0}}}},
}}}
var unSetStage = bson.D{{"$unset", bson.A{"transferPrices"}}}

func TestPipeline_FindByChainAndAppId(t *testing.T) {
	cases := []struct {
		name     string
		query    operations.OperationQuery
		expected mongo.Pipeline
	}{
		{
			name:  "Search with no query filters",
			query: operations.OperationQuery{},
			expected: mongo.Pipeline{
				sortStage,
				skipStage,
				limitStage,
				lookupVaasStage,
				lookupTransferPricesStage,
				lookupGlobalTransactionsStage,
				addFieldsStage,
				unSetStage,
			},
		},
		{
			name: "Search with single source_chain ",
			query: operations.OperationQuery{
				SourceChainIDs: []sdk.ChainID{1},
			},
			expected: mongo.Pipeline{
				bson.D{{"$match", bson.M{"$and": bson.A{
					bson.M{"rawStandardizedProperties.fromChain": bson.M{"$in": []sdk.ChainID{1}}},
				}}}},
				sortStage,
				skipStage,
				limitStage,
				lookupVaasStage,
				lookupTransferPricesStage,
				lookupGlobalTransactionsStage,
				addFieldsStage,
				unSetStage,
			},
		},
		{
			name: "Search with multiple source_chain ",
			query: operations.OperationQuery{
				SourceChainIDs: []sdk.ChainID{1, 2},
			},
			expected: mongo.Pipeline{
				bson.D{{"$match", bson.M{"$and": bson.A{
					bson.M{"rawStandardizedProperties.fromChain": bson.M{"$in": []sdk.ChainID{1, 2}}},
				}}}},
				sortStage,
				skipStage,
				limitStage,
				lookupVaasStage,
				lookupTransferPricesStage,
				lookupGlobalTransactionsStage,
				addFieldsStage,
				unSetStage,
			},
		},
		{
			name: "Search with single target_chain ",
			query: operations.OperationQuery{
				TargetChainIDs: []sdk.ChainID{1},
			},
			expected: mongo.Pipeline{
				bson.D{{"$match", bson.M{"$and": bson.A{
					bson.M{"rawStandardizedProperties.toChain": bson.M{"$in": []sdk.ChainID{1}}},
				}}}},
				sortStage,
				skipStage,
				limitStage,
				lookupVaasStage,
				lookupTransferPricesStage,
				lookupGlobalTransactionsStage,
				addFieldsStage,
				unSetStage,
			},
		},
		{
			name: "Search with single target_chain ",
			query: operations.OperationQuery{
				TargetChainIDs: []sdk.ChainID{1, 2},
			},
			expected: mongo.Pipeline{
				bson.D{{"$match", bson.M{"$and": bson.A{
					bson.M{"rawStandardizedProperties.toChain": bson.M{"$in": []sdk.ChainID{1, 2}}},
				}}}},
				sortStage,
				skipStage,
				limitStage,
				lookupVaasStage,
				lookupTransferPricesStage,
				lookupGlobalTransactionsStage,
				addFieldsStage,
				unSetStage,
			},
		},
		{
			name: "Search with same source and target chain",
			query: operations.OperationQuery{
				SourceChainIDs: []sdk.ChainID{1},
				TargetChainIDs: []sdk.ChainID{1},
			},
			expected: mongo.Pipeline{
				bson.D{{"$match", bson.M{"$or": bson.A{
					bson.M{"rawStandardizedProperties.fromChain": bson.M{"$in": []sdk.ChainID{1}}},
					bson.M{"rawStandardizedProperties.toChain": bson.M{"$in": []sdk.ChainID{1}}},
				}}}},
				sortStage,
				skipStage,
				limitStage,
				lookupVaasStage,
				lookupTransferPricesStage,
				lookupGlobalTransactionsStage,
				addFieldsStage,
				unSetStage,
			},
		},
		{
			name: "Search with different source and target chain",
			query: operations.OperationQuery{
				SourceChainIDs: []sdk.ChainID{1},
				TargetChainIDs: []sdk.ChainID{2},
			},
			expected: mongo.Pipeline{
				bson.D{{"$match", bson.M{"$and": bson.A{
					bson.M{"rawStandardizedProperties.fromChain": bson.M{"$in": []sdk.ChainID{1}}},
					bson.M{"rawStandardizedProperties.toChain": bson.M{"$in": []sdk.ChainID{2}}},
				}}}},
				sortStage,
				skipStage,
				limitStage,
				lookupVaasStage,
				lookupTransferPricesStage,
				lookupGlobalTransactionsStage,
				addFieldsStage,
				unSetStage,
			},
		},
		{
			name: "Search by appID exclusive",
			query: operations.OperationQuery{
				AppIDs:         []string{"CCTP_WORMHOLE_INTEGRATION", "PORTAL_TOKEN_BRIDGE"},
				ExclusiveAppId: true,
			},
			expected: mongo.Pipeline{
				bson.D{{"$match", bson.M{"$or": bson.A{
					bson.M{"$and": bson.A{
						bson.M{"rawStandardizedProperties.appIds": bson.M{"$eq": "CCTP_WORMHOLE_INTEGRATION"}},
						bson.M{"rawStandardizedProperties.appIds": bson.M{"$size": 1}},
					}},
					bson.M{"$and": bson.A{
						bson.M{"rawStandardizedProperties.appIds": bson.M{"$eq": "PORTAL_TOKEN_BRIDGE"}},
						bson.M{"rawStandardizedProperties.appIds": bson.M{"$size": 1}},
					}},
				}}}},
				sortStage,
				skipStage,
				limitStage,
				lookupVaasStage,
				lookupTransferPricesStage,
				lookupGlobalTransactionsStage,
				addFieldsStage,
				unSetStage,
			},
		},
		{
			name: "Search by appID exclusive false",
			query: operations.OperationQuery{
				AppIDs:         []string{"CCTP_WORMHOLE_INTEGRATION", "PORTAL_TOKEN_BRIDGE"},
				ExclusiveAppId: false,
			},
			expected: mongo.Pipeline{
				bson.D{{"$match", bson.M{"rawStandardizedProperties.appIds": bson.M{"$in": []string{"CCTP_WORMHOLE_INTEGRATION", "PORTAL_TOKEN_BRIDGE"}}}}},
				sortStage,
				skipStage,
				limitStage,
				lookupVaasStage,
				lookupTransferPricesStage,
				lookupGlobalTransactionsStage,
				addFieldsStage,
				unSetStage,
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result := operations.BuildPipelineSearchByChainAndAppID(testCase.query)
			assert.Equal(t, testCase.expected, result, "Expected pipeline did not match actual pipeline")
		})
	}
}
