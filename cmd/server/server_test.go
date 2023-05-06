package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractNameValue(t *testing.T) {
	r, err := http.NewRequest(http.MethodPost, "http://<АДРЕС_СЕРВЕРА>/update/metricType/name/val", nil)
	require.NoError(t, err)

	valType, name, val, err := getMetricTypeNameValue(r)
	require.NoError(t, err)
	require.Equal(t, valType, "metricType")
	require.Equal(t, name, "name")
	require.Equal(t, val, "val")
}
