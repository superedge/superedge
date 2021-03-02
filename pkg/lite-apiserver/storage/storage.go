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
	"github.com/superedge/superedge/pkg/lite-apiserver/config"
	"github.com/superedge/superedge/pkg/lite-apiserver/constant"
	"k8s.io/klog"
	"os"
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
	default:
		// error type, use FileStorage
		return NewFileStorage(config.FileCachePath)
	}
}

func mkdir(dirPath string) {
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		klog.Fatalf("mkdir %s error: %v", dirPath, err)
	}
}
