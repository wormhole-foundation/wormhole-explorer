package config

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPrefix(t *testing.T) {
	os.Clearenv()
	cfg := Configuration{
		P2pNetwork:  "mainnet",
		Environment: "staging",
	}
	prefix := cfg.GetPrefix()
	assert.Equal(t, "mainnet-staging", prefix)
}

func TestGetPrefixNoP2P(t *testing.T) {
	os.Clearenv()
	os.Setenv("ENVIRONMENT", "staging")

	isLocal := true
	_, err := New(context.TODO(), &isLocal)
	assert.NotNil(t, err)
}
