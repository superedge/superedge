package cache

import (
	"encoding/json"
	"net/http"

	"github.com/superedge/superedge/pkg/lite-apiserver/constant"
)

// EdgeResponse hold all data of a response
type EdgeCache struct {
	Status int         `json:"code"`
	Header http.Header `json:"header"`
	Body   []byte      `json:"body"`
}

func NewEdgeCache(status int, header http.Header, body []byte) *EdgeCache {
	return &EdgeCache{
		Status: status,
		Header: header,
		Body:   body,
	}
}

func NewEmptyEdgeCache() *EdgeCache {
	header := make(http.Header)
	header.Set(constant.ContentType, "application/json")

	return &EdgeCache{
		Status: http.StatusOK,
		Header: header,
		Body:   []byte(""),
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