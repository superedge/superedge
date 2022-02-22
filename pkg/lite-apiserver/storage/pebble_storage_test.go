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
	"github.com/cockroachdb/pebble"
	"gotest.tools/assert"
	"os"
	"testing"
)

func TestPebbleStorage_StoreOne(t *testing.T) {
	path := "pebble_one"
	db, err := pebble.Open(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	storage := newPebbleStoreWithDb(db)

	key := "one-key"
	data := []byte("one key data")

	err = storage.StoreOne(key, data)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := storage.LoadOne(key)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(data), string(cache))
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
	cleanPath(path, t)
}

func TestPebbleStorage_StoreList(t *testing.T) {
	path := "pebble_one"
	db, err := pebble.Open(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	storage := newPebbleStoreWithDb(db)

	key := "list-key"
	data := []byte("list key data")

	err = storage.StoreList(key, data)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := storage.LoadList(key)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(data), string(cache))

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	cleanPath(path, t)
}

func TestPebbleStorage_Delete(t *testing.T) {
	path := "pebble_delete"
	db, err := pebble.Open(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	storage := newPebbleStoreWithDb(db)

	key := "delete-key"
	data := []byte("delete key data")

	err = storage.StoreOne(key, data)
	if err != nil {
		t.Fatal(err)
	}

	err = storage.Delete(key)
	if err != nil {
		t.Fatal(err)
	}

	cache, _ := storage.LoadOne(key)

	assert.Equal(t, 0, len(cache))

	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}

	cleanPath(path, t)
}

func cleanPath(path string, t *testing.T) {
	err := os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}
