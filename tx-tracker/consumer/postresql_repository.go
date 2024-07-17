package consumer

import (
	"context"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
)

type PostgreSQLRepository interface {
	UpsertOriginTx(ctx context.Context, params *UpsertOriginTxParams) error
	UpsertTargetTx(ctx context.Context, globalTx *TargetTxUpdate) error
}

func NewPostgreSQLRepository(ctx context.Context, databaseURL string) (PostgreSQLRepository, error) {

	postreSQLClient, err := db.NewDB(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	return &postreSQLRepository{
		dbClient: postreSQLClient,
	}, err
}

type postreSQLRepository struct {
	dbClient *db.DB
}

func (p *postreSQLRepository) UpsertOriginTx(ctx context.Context, params *UpsertOriginTxParams) error {
	//TODO implement me
	panic("implement me")
}

func (p *postreSQLRepository) UpsertTargetTx(ctx context.Context, globalTx *TargetTxUpdate) error {
	//TODO implement me
	panic("implement me")
}
