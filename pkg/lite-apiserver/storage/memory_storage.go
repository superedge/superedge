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

package storage

import (
	"fmt"
)

type memoryStorage struct {
	oneMap  map[string][]byte
	listMap map[string][]byte
}

func NewMemoryStorage() Storage {
	return &memoryStorage{
		oneMap:  make(map[string][]byte),
		listMap: make(map[string][]byte),
	}
}

func (ms *memoryStorage) StoreOne(key string, cache []byte) error {
	ms.oneMap[key] = cache
	return nil
}

func (ms *memoryStorage) StoreList(key string, cache []byte) error {
	ms.listMap[key] = cache
	return nil
}

func (ms *memoryStorage) LoadOne(key string) ([]byte, error) {
	c, ok := ms.oneMap[key]
	if ok {
		return c, nil
	}
	return nil, fmt.Errorf("key %s not found", key)
}

func (ms *memoryStorage) LoadList(key string) ([]byte, error) {
	c, ok := ms.listMap[key]
	if ok {
		return c, nil
	}
	return nil, fmt.Errorf("key %s not found", key)
}

func (ms *memoryStorage) Delete(key string) error {
	return nil
}
