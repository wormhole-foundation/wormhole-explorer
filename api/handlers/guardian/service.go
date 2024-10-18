package guardian

import (
	"context"
	"time"

	"sort"

	"github.com/certusone/wormhole/node/pkg/common"
	eth_common "github.com/ethereum/go-ethereum/common"
	"github.com/wormhole-foundation/wormhole-explorer/api/cacheable"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"go.uber.org/zap"
)

type Service struct {
	repo       *repository.MongoGuardianSetRepository
	p2pNetwork string
	cache      cache.Cache
	metrics    metrics.Metrics
	logger     *zap.Logger
}

const currentGuardianSetKey = "current-guardian-set"

func NewService(repo *repository.MongoGuardianSetRepository, p2pNetwork string, cache cache.Cache,
	metrics metrics.Metrics, logger *zap.Logger) *Service {
	return &Service{
		repo:       repo,
		p2pNetwork: p2pNetwork,
		cache:      cache,
		metrics:    metrics,
		logger:     logger.With(zap.String("module", "GuardianService")),
	}
}

func (s *Service) GetGuardianSet(ctx context.Context) (*GuardianSet, error) {
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, 1*time.Minute, currentGuardianSetKey, s.metrics,
		func() (*GuardianSet, error) {
			return s.getGuardianSet(ctx)
		})
}

func (s *Service) getGuardianSet(ctx context.Context) (*GuardianSet, error) {

	docs, err := s.repo.FindAll(ctx)
	if err != nil {
		s.logger.Error("failed to get guardian set from repository", zap.Error(err))
		gs := getByEnv(s.p2pNetwork)
		return &gs, nil
	}

	gs := &GuardianSet{}
	if len(docs) == 0 {
		s.logger.Error("guardian set not fetched from chain yet")
		return gs, nil
	}
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].GuardianSetIndex > docs[j].GuardianSetIndex
	})
	var gstByIndex []common.GuardianSet
	var expirationTimeByIndex []time.Time
	for _, doc := range docs {
		sort.Slice(doc.Keys, func(i, j int) bool {
			return doc.Keys[i].Index < doc.Keys[j].Index
		})
		keys := make([]eth_common.Address, len(doc.Keys))
		for i, k := range doc.Keys {
			keys[i] = eth_common.BytesToAddress(k.Address)
		}
		gstByIndex = append(gstByIndex, common.GuardianSet{
			Keys:  keys,
			Index: doc.GuardianSetIndex,
		})
		var expirationTime time.Time
		if doc.ExpirationTime != nil {
			expirationTime = *doc.ExpirationTime
		}
		expirationTimeByIndex = append(expirationTimeByIndex, expirationTime)
	}
	sort.Slice(gstByIndex, func(i, j int) bool {
		return gstByIndex[i].Index < gstByIndex[j].Index
	})
	sort.Slice(expirationTimeByIndex, func(i, j int) bool {
		return gstByIndex[i].Index < gstByIndex[j].Index
	})
	return &GuardianSet{
		GstByIndex:            gstByIndex,
		ExpirationTimeByIndex: expirationTimeByIndex,
	}, nil
}
