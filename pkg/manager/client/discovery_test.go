package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_toTitle(t *testing.T) {
	kind := "CustomResourceDefinition"
	assert := require.New(t)
	assert.Equal(kind, toTitle(kind), "expected kind to be same")
}
