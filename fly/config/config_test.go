package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPrefix(t *testing.T) {
	os.Setenv("P2P_NETWORK", "mainnet")
	os.Setenv("ENVIROMENT", "staging")

	prefix := GetPrefix()

	assert.Equal(t, "mainnet-staging:", prefix)
}

func TestGetPrefixNoP2P(t *testing.T) {
	os.Setenv("ENVIROMENT", "staging")

	prefix := GetPrefix()

	assert.Equal(t, "", prefix)
}
