package common

import (
	"context"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoID struct {
	Id string `bson:"_id"`
}

func FindVaasIdsByFromAddressOrToAddress(
	ctx context.Context,
	db *mongo.Database,
	address string,
) ([]string, error) {
	addressHexa := strings.ToLower(address)
	if !utils.StartsWith0x(address) {
		addressHexa = "0x" + strings.ToLower(addressHexa)
	}

	matchForToAddress := bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "standardizedProperties.toAddress", Value: bson.M{"$eq": addressHexa}}},
		bson.D{{Key: "standardizedProperties.toAddress", Value: bson.M{"$eq": address}}},
	}}}}}

	matchForFromAddress := bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "originTx.from", Value: bson.M{"$eq": addressHexa}}},
		bson.D{{Key: "originTx.from", Value: bson.M{"$eq": address}}},
	}}}}}

	toAddressFilter := bson.D{{Key: "$unionWith", Value: bson.D{{Key: "coll", Value: "parsedVaa"}, {Key: "pipeline", Value: bson.A{matchForToAddress}}}}}
	fromAddressFilter := bson.D{{Key: "$unionWith", Value: bson.D{{Key: "coll", Value: "globalTransactions"}, {Key: "pipeline", Value: bson.A{matchForFromAddress}}}}}
	group := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$_id"}}}}

	pipeline := []bson.D{fromAddressFilter, toAddressFilter, group}

	cur, err := db.Collection("_temporal").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var documents []mongoID
	err = cur.All(ctx, &documents)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, doc := range documents {
		ids = append(ids, doc.Id)
	}
	return ids, nil
}
