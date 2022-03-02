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
	"os"

	"k8s.io/klog/v2"

	"github.com/superedge/superedge/pkg/lite-apiserver/config"
	"github.com/superedge/superedge/pkg/lite-apiserver/constant"
)

type Storage interface {
	StoreOne(key string, cache []byte) error

	StoreList(key string, cache []byte) error

	LoadOne(key string) ([]byte, error)

	LoadList(key string) ([]byte, error)

	Delete(key string) error
}

func CreateStorage(config *config.LiteServerConfig) Storage {
	switch config.CacheType {
	case constant.FileStorage:
		return NewFileStorage(config.FileCachePath)
	case constant.MemoryStorage:
		return NewMemoryStorage()
	case constant.BadgerStorage:
		return NewBadgerStorage(config.BadgerCachePath)
	case constant.BoltStorage:
		return NewBoltStorage(config.BoltCacheFile)
	case constant.PebbleStorage:
		return NewPebbleStorage(config.PebbleCachePath)
	default:
		// error type, use FileStorage
		klog.Errorf("%s is not supported, use default %s cache storage", config.CacheType, constant.FileStorage)
		return NewFileStorage(config.FileCachePath)
	}
}

func mkdir(dirPath string) {
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		klog.Fatalf("mkdir %s error: %v", dirPath, err)
	}
}
