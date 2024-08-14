package consumer

import (
	"context"
)

type DualRepository struct {
	mongoRepository    *MongoRepository
	postgresRepository *PostgreSQLRepository
}

func NewDualRepository(mongoRepository *MongoRepository,
	postgresRepository *PostgreSQLRepository) Repository {
	return &DualRepository{
		mongoRepository:    mongoRepository,
		postgresRepository: postgresRepository,
	}
}

func (r *DualRepository) AlreadyProcessed(ctx context.Context, vaaId string, digest string) (bool, error) {
	processed, err := r.mongoRepository.AlreadyProcessed(ctx, vaaId, digest)
	if err != nil {
		return false, err
	}
	if !processed {
		return false, nil
	}
	processed, err = r.postgresRepository.AlreadyProcessed(ctx, vaaId, digest)
	if err != nil {
		return false, err
	}
	return processed, nil
}

func (r *DualRepository) GetVaaIdTxHash(ctx context.Context, vaaID, vaaDigest string) (*VaaIdTxHash, error) {
	vaaIdTxHash, err := r.mongoRepository.GetVaaIdTxHash(ctx, vaaID, vaaDigest)
	if err == nil && vaaIdTxHash != nil {
		return vaaIdTxHash, nil
	}
	vaaIdTxHash, err = r.postgresRepository.GetVaaIdTxHash(ctx, vaaID, vaaDigest)
	if err != nil {
		return nil, err
	}
	return vaaIdTxHash, nil
}

func (r *DualRepository) UpsertTargetTx(ctx context.Context, globalTx *TargetTxUpdate) error {
	err := r.mongoRepository.UpsertTargetTx(ctx, globalTx)
	if err != nil {
		return err
	}
	return r.postgresRepository.UpsertTargetTx(ctx, globalTx)
}

func (r *DualRepository) GetTxStatus(ctx context.Context, targetTxUpdate *TargetTxUpdate) (string, error) {
	status, err := r.mongoRepository.GetTxStatus(ctx, targetTxUpdate)
	if err != nil {
		status, err = r.postgresRepository.GetTxStatus(ctx, targetTxUpdate)
		if err != nil {
			return "", err
		}
		return status, nil
	}
	return status, nil
}

func (r *DualRepository) FindSourceTxById(ctx context.Context, id string) (*SourceTxDoc, error) {
	sourceTxDoc, err := r.mongoRepository.FindSourceTxById(ctx, id)
	if err == nil && sourceTxDoc != nil {
		return sourceTxDoc, nil
	}
	return r.postgresRepository.FindSourceTxById(ctx, id)
}

func (r *DualRepository) UpsertOriginTx(ctx context.Context, originTx, nestedTx *UpsertOriginTxParams) error {
	err := r.mongoRepository.UpsertOriginTx(ctx, originTx, nestedTx)
	if err != nil {
		return err
	}
	return r.postgresRepository.UpsertOriginTx(ctx, originTx, nestedTx)
}

func (r *DualRepository) RegisterProcessedVaa(ctx context.Context, vaaDigest, vaaId string) error {
	return r.postgresRepository.RegisterProcessedVaa(ctx, vaaDigest, vaaId)
}
