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
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"

	"k8s.io/klog/v2"
)

const (
	// Default BadgerDB discardRatio. It represents the discard ratio for the
	// BadgerDB GC.
	//
	// Ref: https://godoc.org/github.com/dgraph-io/badger#DB.RunValueLogGC
	badgerDiscardRatio = 0.5

	// Default BadgerDB GC interval
	badgerGCInterval = 10 * time.Minute
)

type badgerStorage struct {
	db *badger.DB
}

func NewBadgerStorage(path string) Storage {
	opts := badger.DefaultOptions(path).
		WithNumMemtables(1).
		WithNumLevelZeroTables(1).
		WithValueLogFileSize(1 << 25).
		WithMemTableSize(4 << 20)

	db, err := badger.Open(opts)
	if err != nil {
		klog.Fatal(err)
	}
	bs := &badgerStorage{
		db: db,
	}

	// run gc
	go bs.runGC()

	return bs
}

func (bs *badgerStorage) StoreOne(key string, data []byte) error {
	klog.V(8).Infof("storage one key=%s, cache=%s", key, string(data))

	err := bs.update(bs.oneKey(key), data)
	if err != nil {
		klog.Errorf("write one cache %s error: %v", key, err)
		return err
	}

	return nil
}

func (bs *badgerStorage) StoreList(key string, data []byte) error {
	klog.V(8).Infof("storage list key=%s, cache=%s", key, string(data))

	err := bs.update(bs.listKey(key), data)
	if err != nil {
		klog.Errorf("write list cache %s error: %v", key, err)
		return err
	}

	return nil
}

func (bs *badgerStorage) LoadOne(key string) ([]byte, error) {
	data, err := bs.get(bs.oneKey(key))
	if err != nil {
		klog.Errorf("read one cache %s error: %v", key, err)
		return nil, err
	}

	klog.V(8).Infof("load one key=%s, cache=%s", key, string(data))
	return data, nil
}

func (bs *badgerStorage) LoadList(key string) ([]byte, error) {
	data, err := bs.get(bs.listKey(key))
	if err != nil {
		klog.Errorf("read list cache %s error: %v", key, err)
		return nil, err
	}

	klog.V(8).Infof("load list key=%s, cache=%s", key, string(data))
	return data, nil
}

func (bs *badgerStorage) Delete(key string) error {
	return nil
}

func (bs *badgerStorage) runGC() {
	ticker := time.NewTicker(badgerGCInterval)
	for {
		select {
		case <-ticker.C:
			err := bs.db.RunValueLogGC(badgerDiscardRatio)
			if err != nil {
				// don't report error when GC didn't result in any cleanup
				if err == badger.ErrNoRewrite {
					klog.V(2).Infof("no BadgerDB GC occurred: %v", err)
				} else {
					klog.Errorf("failed to GC BadgerDB: %v", err)
				}
			}
		}
	}
}

func (bs *badgerStorage) oneKey(key string) string {
	return strings.ReplaceAll(key, "/", "_")
}

func (bs *badgerStorage) listKey(key string) string {
	return strings.ReplaceAll(fmt.Sprintf("%s_%s", key, "list"), "/", "_")
}

func (bs *badgerStorage) get(key string) ([]byte, error) {
	var data []byte
	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(bs.oneKey(key)))
		if err != nil {
			klog.Errorf("get one cache %s error: %v", key, err)
			return err
		}

		data, err = item.ValueCopy(nil)
		if err != nil {
			klog.Errorf("copy one cache %s error: %v", key, err)
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (bs *badgerStorage) update(key string, data []byte) error {
	err := bs.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), data)
		return err
	})
	return err
}
