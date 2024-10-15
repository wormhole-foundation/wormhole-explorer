package storage

import "context"

type pricesRepositoryComposite struct {
	repos []PricesRepository
}

// NewPricesRepositoryComposite creates a new storage repository.
func NewPricesRepositoryComposite(repos ...PricesRepository) PricesRepository {
	return &pricesRepositoryComposite{
		repos: repos,
	}
}

func (r *pricesRepositoryComposite) Upsert(ctx context.Context, op OperationPrice) error {
	for _, repo := range r.repos {
		if err := repo.Upsert(ctx, op); err != nil {
			return err
		}
	}
	return nil
}

type vaaRepositoryComposite struct {
	repos []VaaRepository
}

// NewVaaRepositoryComposite creates a new storage repository.
func NewVaaRepositoryComposite(repos ...VaaRepository) VaaRepository {
	return &vaaRepositoryComposite{
		repos: repos,
	}
}

func (r *vaaRepositoryComposite) FindByVaaID(ctx context.Context, id string) (*Vaa, error) {
	var rerr error
	for _, repo := range r.repos {
		vaa, rerr := repo.FindByVaaID(ctx, id)
		if rerr == nil && vaa != nil {
			return vaa, nil
		}
	}
	return nil, rerr
}

func (r *vaaRepositoryComposite) FindPage(ctx context.Context, query VaaPageQuery, pagination Pagination) ([]*Vaa, error) {
	var rerr error
	for _, repo := range r.repos {
		result, rerr := repo.FindPage(ctx, query, pagination)
		if rerr == nil {
			return result, nil
		}
	}
	return nil, rerr
}
