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
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

type fileStorage struct {
	filePath string
	seedMu   sync.Mutex
	seed     *rand.Rand
}

func NewFileStorage(filePath string) Storage {
	mkdir(filePath)

	fs := &fileStorage{
		seed:     rand.New(rand.NewSource(time.Now().Unix())),
		filePath: filePath,
	}

	return fs
}

func (fs *fileStorage) StoreOne(key string, data []byte) error {
	klog.V(8).Infof("storage one key=%s, cache=%s", key, string(data))

	err := fs.writeFile(fs.oneFileName(key), data)
	if err != nil {
		klog.Errorf("write one cache %s error: %v", key, err)
		return err
	}

	return nil
}

func (fs *fileStorage) StoreList(key string, data []byte) error {
	klog.V(8).Infof("storage list key=%s, cache=%s", key, string(data))

	err := fs.writeFile(fs.listFileName(key), data)
	if err != nil {
		klog.Errorf("write list cache %s error: %v", key, err)
		return err
	}

	return nil
}

func (fs *fileStorage) LoadOne(key string) ([]byte, error) {
	data, err := fs.readFile(fs.oneFileName(key))
	if err != nil {
		klog.Errorf("read one cache %s error: %v", key, err)
		return nil, err
	}

	klog.V(8).Infof("load one key=%s, cache=%s", key, string(data))
	return data, nil
}

func (fs *fileStorage) LoadList(key string) ([]byte, error) {
	data, err := fs.readFile(fs.listFileName(key))
	if err != nil {
		klog.Errorf("read list cache %s error: %v", key, err)
		return nil, err
	}

	klog.V(8).Infof("load list key=%s, cache=%s", key, string(data))
	return data, nil
}

func (fs *fileStorage) Delete(key string) error {
	return nil
}

func (fs *fileStorage) randomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	var result []byte
	for i := 0; i < l; i++ {
		fs.seedMu.Lock()
		result = append(result, bytes[fs.seed.Intn(len(bytes))])
		fs.seedMu.Unlock()
	}
	return string(result)
}

func (fs *fileStorage) writeFile(fileName string, data []byte) error {
	salt := fs.randomString(12)
	tmpFileName := fmt.Sprintf("%s_%s", fileName, salt)

	f, err := os.Create(filepath.Join(fs.filePath, tmpFileName))
	if err != nil {
		klog.Errorf("create file %s error: %v", tmpFileName, err)
		return err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			klog.Errorf("close file error: %v", err)
		}
	}()

	_, err = f.Write(data)
	if err != nil {
		klog.Errorf("write file %s error: %v", tmpFileName, err)
		return err
	}

	err = os.Rename(filepath.Join(fs.filePath, tmpFileName), filepath.Join(fs.filePath, fileName))
	if err != nil {
		klog.Errorf("rename tmp to target file error: %v", err)
		return err
	}
	return nil
}

func (fs *fileStorage) readFile(fileName string) ([]byte, error) {
	f, err := os.Open(filepath.Join(fs.filePath, fileName))
	if err != nil {
		klog.Errorf("open cache file %s error: %v", fileName, err)
		return []byte{}, err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			klog.Errorf("close file error: %v", err)
		}
	}()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		klog.Errorf("read cache %s error: %v", fileName, err)
		return []byte{}, err
	}
	return data, nil
}

func (fs *fileStorage) oneFileName(key string) string {
	return strings.ReplaceAll(key, "/", "_")
}

func (fs *fileStorage) listFileName(key string) string {
	return strings.ReplaceAll(fmt.Sprintf("%s_%s", key, "list"), "/", "_")
}
