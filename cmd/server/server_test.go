package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractNameValue(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/name/val", nil)
	require.NoError(t, err)

	name, val, err := getMetricNameValue(r)
	require.NoError(t, err)
	require.Equal(t, name, "name")
	require.Equal(t, val, "val")
}
