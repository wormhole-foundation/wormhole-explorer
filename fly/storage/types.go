package storage

import (
	"errors"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type Uint64 uint64

func (u Uint64) MarshalBSONValue() (bsontype.Type, []byte, error) {
	ui64Str := strconv.FormatUint(uint64(u), 10)
	d128, err := primitive.ParseDecimal128(ui64Str)
	return bsontype.Decimal128, bsoncore.AppendDecimal128(nil, d128), err
}

func (u *Uint64) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	d128, _, ok := bsoncore.ReadDecimal128(b)
	if !ok {
		return errors.New("Uint64 UnmarshalBSONValue error")
	}

	ui64, err := strconv.ParseUint(d128.String(), 10, 64)
	if err != nil {
		return err
	}

	*u = Uint64(ui64)
	return nil
}
