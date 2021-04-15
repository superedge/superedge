package cache

import (
	"encoding/json"
	"net/http"

	"github.com/superedge/superedge/pkg/lite-apiserver/constant"
)

// EdgeCache hold all data of a response
type EdgeCache struct {
	StatusCode int         `json:"code"`
	Header     http.Header `json:"header"`
	Body       []byte      `json:"body"`
}

func NewEdgeCache(statusCode int, header http.Header, body []byte) *EdgeCache {
	return &EdgeCache{
		StatusCode: statusCode,
		Header:     header,
		Body:       body,
	}
}

func NewDefaultEdgeCache() *EdgeCache {
	header := make(http.Header)
	header.Set(constant.ContentType, "application/json")

	return &EdgeCache{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       []byte(""),
	}
}

func MarshalEdgeCache(cache *EdgeCache) ([]byte, error) {
	return json.Marshal(cache)
}

func UnmarshalEdgeCache(data []byte) (*EdgeCache, error) {
	cache := &EdgeCache{}
	err := json.Unmarshal(data, cache)

	return cache, err
}
