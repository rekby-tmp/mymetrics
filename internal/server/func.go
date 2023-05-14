package server

import (
	"fmt"
	"github.com/rekby-tmp/mymetrics/internal/common"
	"net/http"
	"strings"
)

func getMetricTypeNameValue(r *http.Request) (valType, name, value string, _ error) {
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 4 {
		return "", "", "", fmt.Errorf("bad parts in path for extrace metric name and balue: %v", len(parts))
	}

	return parts[1], parts[2], parts[3], nil
}

func isAcceptGzipResponse(request *http.Request) bool {
	for _, accept := range request.Header.Values("Accept-Encoding") {
		if accept == common.GzipEncoding {
			return true
		}
	}

	return false
}

func needCompress(contentType string) bool {
	return contentType == common.JsonType || contentType == common.HtmlType
}
