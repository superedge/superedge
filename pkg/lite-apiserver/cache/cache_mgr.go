/*
Copyright 2020 The SuperEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/apimachinery/pkg/watch"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/kubernetes/scheme"
	restclientwatch "k8s.io/client-go/rest/watch"
	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/lite-apiserver/constant"
	"github.com/superedge/superedge/pkg/lite-apiserver/storage"
)

type CacheManager struct {
	storage storage.Storage
}

func NewCacheManager(storage storage.Storage) *CacheManager {
	return &CacheManager{
		storage: storage,
	}
}

func (c CacheManager) Cache(req *http.Request, statusCode int, header http.Header, body io.ReadCloser) error {
	info, ok := apirequest.RequestInfoFrom(req.Context())
	if !ok {
		return fmt.Errorf("no RequestInfo found in the context")
	}

	userAgent := getUserAgent(req)
	key := cacheKey(userAgent, info)
	klog.V(4).Infof("cache for %s", key)

	return c.handleCache(key, info.Verb, statusCode, header, body)
}

func (c CacheManager) handleCache(key string, verb string, statusCode int, header http.Header, body io.ReadCloser) error {
	// handle watch with streaming
	if verb == constant.VerbWatch {
		return c.cacheWatch(key, body)
	}

	// handle get/list
	var buf bytes.Buffer
	n, err := buf.ReadFrom(body)
	if err != nil {
		klog.Errorf("failed to get cache response, %v", err)
		return err
	}
	klog.V(4).Infof("cache %d bytes body from response for %s", n, key)

	header.Del(constant.ContentLength)

	if verb == constant.VerbList {
		return c.cacheList(key, statusCode, header, buf.Bytes())
	} else if verb == constant.VerbGet {
		return c.cacheGet(key, statusCode, header, buf.Bytes())
	}

	return nil
}

func (c CacheManager) cacheWatch(key string, body io.ReadCloser) error {
	decoder := getWatchDecoder(body)
	accessor := meta.NewAccessor()

	for {
		eventType, obj, err := decoder.Decode()
		if err != nil {
			klog.V(2).Infof("watch for key %s decode end: %v", key, err)
			return err
		}

		klog.V(6).Infof("new watch event, key=%s, type=%s, obj=%v", key, eventType, obj)

		switch eventType {
		case watch.Bookmark:
			klog.Infof("watch bookmark for key %s", key)
			continue
		case watch.Error:
			klog.Infof("watch error for key %s", key)
			continue
		case watch.Added, watch.Modified, watch.Deleted:
			// get list object
			listData, err := c.storage.LoadList(key)
			if err != nil {
				klog.Warningf("get list for key %s error: %v", key, err)
				continue
			}

			listCache, err := UnmarshalEdgeCache(listData)
			if err != nil {
				klog.Errorf("unmarshal key %s error: %v", key, err)
				return err
			}

			listObj, _, err := unstructured.UnstructuredJSONScheme.Decode(listCache.Body, nil, nil)
			if err != nil {
				klog.Warningf("decode list for key %s error: %v", key, err)
				continue
			}

			items, err := meta.ExtractList(listObj)
			if err != nil {
				klog.Errorf("extract list error: %v", err)
				continue
			}

			// update
			updated := false
			switch eventType {
			case watch.Added:
				items = append(items, obj)
				updated = true
			case watch.Modified:
				for i, item := range items {
					if sameObj(accessor, obj, item) && newVersion(accessor, obj, item) {
						items[i] = obj
						updated = true
						break
					}
				}
			case watch.Deleted:
				tmp := items[:0]
				for _, item := range items {
					if !sameObj(accessor, obj, item) {
						tmp = append(tmp, item)
					}
				}
				items = tmp
				updated = true
			default:
				// impossible
			}

			if !updated {
				klog.V(4).Infof("list object %s do not updated for event %s", key, eventType)
				continue
			}

			// update items
			err = meta.SetList(listObj, items)
			if err != nil {
				klog.Errorf("set list error: %v", err)
				return err
			}

			// update resource version
			rv, _ := accessor.ResourceVersion(obj)
			err = accessor.SetResourceVersion(listObj, rv)
			if err != nil {
				klog.Errorf("set resource version %s error: %v", rv, err)
			}

			buffer := new(bytes.Buffer)
			err = unstructured.UnstructuredJSONScheme.Encode(listObj, buffer)
			if err != nil {
				klog.Errorf("encode list error: %v", err)
				return err
			}
			listCache.Body = buffer.Bytes()
			// update cache
			data, err := MarshalEdgeCache(listCache)
			if err != nil {
				klog.Errorf("marshal key %s error: %v", key, err)
				return err
			}
			err = c.storage.StoreList(key, data)
			if err != nil {
				klog.Errorf("update cache list for %s error: %v", key, err)
				return err
			}
		}
	}
}

func (c CacheManager) cacheGet(key string, code int, header http.Header, body []byte) error {
	cache := NewEdgeCache(code, header, body)
	data, err := MarshalEdgeCache(cache)
	if err != nil {
		klog.Errorf("marshal key %s error: %v", key, err)
		return err
	}

	err = c.storage.StoreOne(key, data)
	if err != nil {
		klog.Errorf("cache one for %s error: %v", key, err)
		return err
	}
	return nil
}

func (c CacheManager) cacheList(key string, code int, header http.Header, body []byte) error {
	var err error
	// decode gzip data
	if header.Get("Content-Encoding") == "gzip" {
		body, err = gzipDecode(body)
		if err != nil {
			klog.Errorf("read gzip body from list cache failed, key:%s, err: %v", key, err)
			return err
		}
		header.Del("Content-Encoding")
	}
	cache := NewEdgeCache(code, header, body)
	data, err := MarshalEdgeCache(cache)
	if err != nil {
		klog.Errorf("marshal key %s error: %v", key, err)
		return err
	}

	err = c.storage.StoreList(key, data)
	if err != nil {
		klog.Errorf("cache list for %s error: %v", key, err)
		return err
	}
	return nil
}

func (c CacheManager) queryList(key string) (*EdgeCache, error) {
	data, err := c.storage.LoadList(key)
	if err != nil {
		return nil, err
	}

	cache, err := UnmarshalEdgeCache(data)
	if err != nil {
		klog.Errorf("unmarshal key %s error: %v", key, err)
		return nil, err
	}

	// compress data to gzip
	if len(cache.Body) > defaultGzipThresholdBytes && cache.Header.Get("Content-Encoding") == "" {
		body, err := gzipEncode(cache.Body)
		if err != nil {
			return nil, err
		}
		cache.Body = body
		cache.Header.Set("Content-Encoding", "gzip")
	}

	return cache, err
}

func (c CacheManager) Query(req *http.Request) (*EdgeCache, error) {
	info, ok := apirequest.RequestInfoFrom(req.Context())
	if !ok {
		return nil, fmt.Errorf("parse requestInfo error")
	}

	userAgent := getUserAgent(req)
	key := cacheKey(userAgent, info)
	klog.V(4).Infof("query cache for key %s", key)

	return c.handleQuery(key, info.Verb)
}

func (c CacheManager) handleQuery(key string, verb string) (*EdgeCache, error) {
	var data []byte
	var err error

	switch verb {
	case constant.VerbList:
		return c.queryList(key)
	case constant.VerbGet:
		data, err = c.storage.LoadOne(key)
	case constant.VerbWatch:
		return NewDefaultEdgeCache(), nil
	default:
		return nil, fmt.Errorf("unsupported verb %s for query cache", verb)
	}

	if err != nil {
		return nil, err
	}

	cache, err := UnmarshalEdgeCache(data)
	if err != nil {
		klog.Errorf("unmarshal key %s error: %v", key, err)
		return nil, err
	}

	return cache, err
}

func cacheKey(userAgent string, info *apirequest.RequestInfo) string {
	keys := []string{userAgent, info.Namespace, info.Resource, info.Name, info.Subresource}
	if info.IsResourceRequest {
		keys = []string{userAgent, info.Namespace, info.Resource, info.Name, info.Subresource}
	} else {
		keys = []string{userAgent, strings.ReplaceAll(info.Path, "/", "-")}
	}
	key := strings.Join(keys, "_")
	return key
}

func getUserAgent(r *http.Request) string {
	userAgent := constant.DefaultUserAgent
	h := r.UserAgent()
	if len(h) > 0 {
		userAgent = strings.Split(h, " ")[0]
	}
	return userAgent
}

func getWatchDecoder(body io.ReadCloser) *restclientwatch.Decoder {
	framer := jsonserializer.Framer.NewFrameReader(body)
	jsonSerializer := jsonserializer.NewSerializerWithOptions(jsonserializer.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, jsonserializer.SerializerOptions{Yaml: false, Pretty: false, Strict: false})
	streamingDecoder := streaming.NewDecoder(framer, jsonSerializer)
	return restclientwatch.NewDecoder(streamingDecoder, unstructured.UnstructuredJSONScheme)
}

func sameObj(accessor meta.MetadataAccessor, obj1 runtime.Object, obj2 runtime.Object) bool {
	// TODO use accessor.UID(obj1) == accessor.UID(obj2) ?

	kind1, err := accessor.Kind(obj1)
	if err != nil {
		klog.Errorf("get kind for obj1 error: %v", err)
		return false
	}
	kind2, err := accessor.Kind(obj2)
	if err != nil {
		klog.Errorf("get kind for obj2 error: %v", err)
		return false
	}

	ns1, err := accessor.Namespace(obj1)
	if err != nil {
		klog.Errorf("get namespace for obj1 error: %v", err)
		return false
	}
	ns2, err := accessor.Namespace(obj2)
	if err != nil {
		klog.Errorf("get namespace for obj2 error: %v", err)
		return false
	}

	name1, err := accessor.Name(obj1)
	if err != nil {
		klog.Errorf("get name for obj1 error: %v", err)
		return false
	}
	name2, err := accessor.Name(obj2)
	if err != nil {
		klog.Errorf("get name for obj2 error: %v", err)
		return false
	}

	return (kind1 == kind2) && (ns1 == ns2) && (name1 == name2)
}

func newVersion(accessor meta.MetadataAccessor, newOne runtime.Object, oldOne runtime.Object) bool {
	newRvStr, err := accessor.ResourceVersion(newOne)
	if err != nil {
		klog.Errorf("get resource version for new error: %v", err)
		return false
	}

	oldRvStr, err := accessor.ResourceVersion(oldOne)
	if err != nil {
		klog.Errorf("get resource version for old error: %v", err)
		return false
	}

	oldRvInt, _ := strconv.Atoi(oldRvStr)
	newRvInt, _ := strconv.Atoi(newRvStr)

	return newRvInt > oldRvInt
}
