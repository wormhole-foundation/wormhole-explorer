package vaa

import "context"

type DualVaaRepository struct {
	repos []VAARepository
}

func NewDualVaaRepository(repos ...VAARepository) VAARepository {
	return &DualVaaRepository{
		repos: repos,
	}
}

func (r *DualVaaRepository) GetVaa(ctx context.Context, id string) (*VaaDoc, error) {
	for _, repo := range r.repos {
		vaa, err := repo.GetVaa(ctx, id)
		if err != nil {
			return nil, err
		}
		if vaa != nil {
			return vaa, nil
		}
	}
	return nil, nil
}

func (r *DualVaaRepository) GetTxHash(ctx context.Context, vaaDigest string) (string, error) {
	for _, repo := range r.repos {
		txHash, err := repo.GetTxHash(ctx, vaaDigest)
		if err != nil {
			return "", err
		}
		if txHash != "" {
			return txHash, nil
		}
	}
	return "", nil
}
