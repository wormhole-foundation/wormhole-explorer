package utils

import (
	"testing"

	"github.com/test-go/testify/assert"
)

func TestIsPrivateIPAsString(t *testing.T) {

	is := IsPrivateIPAsString("127.0.0.1")

	assert.Equal(t, true, is)

	is = IsPrivateIPAsString("10.0.1.8")

	assert.Equal(t, true, is)

	is = IsPrivateIPAsString("200.121.10.92")

	assert.Equal(t, false, is)
}
