package common

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJsonType(t *testing.T) {
	require.Equal(t, "application/json", JsonType)
}
