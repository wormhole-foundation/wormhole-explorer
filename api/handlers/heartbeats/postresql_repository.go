package heartbeats

import (
	"context"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"go.uber.org/zap"
)

type PostresqlRepository struct {
	db     *db.DB
	logger *zap.Logger
}

// NewPostresqlRepository creates a new repository.
func NewPostresqlRepository(db *db.DB, logger *zap.Logger) *PostresqlRepository {
	return &PostresqlRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostresqlRepository) FindByIDs(ctx context.Context, ids []string) ([]*HeartbeatDoc, error) {

	if len(ids) == 0 {
		return nil, fmt.Errorf("ids list is empty")
	}

	// normalize the ids (guardian addresses)
	var addresses []string
	for _, id := range ids {
		address := utils.NormalizeHex(id)
		addresses = append(addresses, address)
	}

	// Prepare the query with placeholder for array parameter
	query := `SELECT id, guardian_name, boot_timestamp, timestamp, version, networks, feature, created_at, updated_at
		FROM wormholescan.wh_heartbeats
		WHERE id = ANY($1);`

	var response []*heartbeatSQL
	err := r.db.Select(ctx, &response, query, addresses)
	if err != nil {
		r.logger.Error("failed to select heartbeats", zap.Error(err))
		return nil, err
	}

	// Convert the response to []*HeartbeatDoc
	var heartbeats []*HeartbeatDoc
	for _, h := range response {
		heartbeats = append(heartbeats, &HeartbeatDoc{
			ID:            h.ID,
			NodeName:      h.NodeName,
			BootTimestamp: h.BootTimestamp.UnixNano(),
			Timestamp:     h.Timestamp.UnixNano(),
			Version:       h.Version,
			Networks:      h.Networks,
			Features:      h.Features,
			UpdatedAt:     h.UpdatedAt,
		})
	}

	return heartbeats, nil
}
