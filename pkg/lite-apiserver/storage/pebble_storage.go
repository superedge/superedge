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
	"github.com/cockroachdb/pebble"
	"k8s.io/klog/v2"
	"strings"
)

var writeOptions = &pebble.WriteOptions{Sync: true}

type pebbleStorage struct {
	db *pebble.DB
}

// just for test
func newPebbleStoreWithDb(db *pebble.DB) Storage {
	ps := &pebbleStorage{
		db: db,
	}

	return ps
}

func NewPebbleStorage(path string) Storage {
	ops := &pebble.Options{}
	db, err := pebble.Open(path, ops)
	if err != nil {
		klog.Fatal(err)
	}

	ps := &pebbleStorage{
		db: db,
	}

	return ps
}

func (ps *pebbleStorage) StoreOne(key string, data []byte) error {
	klog.V(8).Infof("storage one key=%s, cache=%s", key, string(data))

	err := ps.db.Set([]byte(key), data, writeOptions)
	if err != nil {
		klog.Errorf("write one cache %s error: %v", key, err)
		return err
	}

	return nil
}

func (ps *pebbleStorage) StoreList(key string, cache []byte) error {
	klog.V(8).Infof("storage list key=%s, cache=%s", key, string(cache))

	err := ps.db.Set([]byte(ps.listKey(key)), cache, writeOptions)
	if err != nil {
		klog.Errorf("write list cache %s error: %v", key, err)
		return err
	}

	return nil
}

func (ps *pebbleStorage) LoadOne(key string) ([]byte, error) {
	data, closer, err := ps.db.Get([]byte(key))
	if err != nil {
		klog.Errorf("read one cache %s error: %v", key, err)
		return nil, err
	}

	if err := closer.Close(); err != nil {
		klog.Error(err)
		return nil, err
	}

	return data, nil
}

func (ps *pebbleStorage) LoadList(key string) ([]byte, error) {
	listKey := ps.listKey(key)
	data, closer, err := ps.db.Get([]byte(listKey))
	if err != nil {
		klog.Errorf("read list cache %s error: %v", key, err)
	}

	if err := closer.Close(); err != nil {
		klog.Error(err)
		return nil, err
	}

	return data, nil
}

func (ps *pebbleStorage) Delete(key string) error {
	return ps.db.Delete([]byte(key), writeOptions)
}

func (ps *pebbleStorage) listKey(key string) string {
	return strings.ReplaceAll(fmt.Sprintf("%s_%s", key, "list"), "/", "_")
}
