package vaa

import (
	"context"
)

type VAARepository interface {
	GetVaa(ctx context.Context, id string) (*VaaDoc, error)
	GetTxHash(ctx context.Context, vaaDigest string) (string, error)
}

type VaaDoc struct {
	VaaID  string `bson:"vaa_id" json:"vaa_id"`
	ID     string `bson:"_id" json:"id"`
	Vaa    []byte `bson:"vaas" json:"vaa"`
	TxHash string `bson:"txHash" json:"txHash"`
}
