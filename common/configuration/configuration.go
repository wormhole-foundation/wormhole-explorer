package configuration

import (
	"context"
	"fmt"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

func LoadFromEnv[T any](ctx context.Context) (*T, error) {

	_ = godotenv.Load()

	var settings T

	err := envconfig.Process(ctx, &settings)
	if err != nil {
		return nil, fmt.Errorf("failed to read config from environment: %w", err)
	}

	return &settings, nil
}

func IsMainnet(s string) bool {
	return strings.ToLower(s) == domain.P2pMainNet
}

func IsTestnet(s string) bool {
	return strings.ToLower(s) == domain.P2pTestNet
}
