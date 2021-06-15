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
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"

	"k8s.io/klog/v2"
)

const bucketName = "SuperEdge"

type boltStorage struct {
	db *bolt.DB
}

func NewBoltStorage(dbFile string) Storage {
	mkdir(filepath.Dir(dbFile))

	db, err := bolt.Open(dbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		klog.Fatalf("init bolt storage error: %v", err)
	}

	// init
	// Start a writable transaction.
	tx, err := db.Begin(true)
	if err != nil {
		klog.Fatalf("init bolt storage error: %v", err)
	}
	defer tx.Rollback()

	// Use the transaction...
	_, err = tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		klog.Fatalf("init bolt storage error: %v", err)
	}

	// Commit the transaction and check for error.
	if err := tx.Commit(); err != nil {
		klog.Fatalf("init bolt storage error: %v", err)
	}

	return &boltStorage{
		db: db,
	}
}

func (bs *boltStorage) StoreOne(key string, data []byte) error {
	klog.V(8).Infof("storage one key=%s, cache=%s", key, string(data))

	err := bs.update(bs.oneKey(key), data)
	if err != nil {
		klog.Errorf("write one cache %s error: %v", key, err)
		return err
	}

	return nil
}

func (bs *boltStorage) StoreList(key string, data []byte) error {
	klog.V(8).Infof("storage list key=%s, cache=%s", key, string(data))

	err := bs.update(bs.listKey(key), data)
	if err != nil {
		klog.Errorf("write list cache %s error: %v", key, err)
		return err
	}

	return nil
}

func (bs *boltStorage) LoadOne(key string) ([]byte, error) {
	data, err := bs.get(bs.oneKey(key))
	if err != nil {
		klog.Errorf("read one cache %s error: %v", key, err)
		return nil, err
	}

	klog.V(8).Infof("load one key=%s, cache=%s", key, string(data))
	return data, nil
}

func (bs *boltStorage) LoadList(key string) ([]byte, error) {
	data, err := bs.get(bs.listKey(key))
	if err != nil {
		klog.Errorf("read list cache %s error: %v", key, err)
		return nil, err
	}

	klog.V(8).Infof("load list key=%s, cache=%s", key, string(data))
	return data, nil
}

func (bs *boltStorage) Delete(key string) error {
	return nil
}

func (bs *boltStorage) oneKey(key string) string {
	return strings.ReplaceAll(key, "/", "_")
}

func (bs *boltStorage) listKey(key string) string {
	return strings.ReplaceAll(fmt.Sprintf("%s_%s", key, "list"), "/", "_")
}

func (bs *boltStorage) get(key string) ([]byte, error) {
	var data []byte
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		data = b.Get([]byte(key))

		if data == nil {
			return fmt.Errorf("no data for %s", key)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (bs *boltStorage) update(key string, data []byte) error {
	err := bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		err := b.Put([]byte(key), data)
		return err
	})
	return err
}
