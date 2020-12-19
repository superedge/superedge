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
	"time"

	"k8s.io/klog"

	"superedge/pkg/lite-apiserver/config"
)

type FileStorage struct {
	filePath string
	seed     *rand.Rand
}

func NewFileStorage(config *config.LiteServerConfig) *FileStorage {
	s := &FileStorage{
		seed:     rand.New(rand.NewSource(time.Now().Unix())),
		filePath: config.FileCachePath,
	}

	mkdirPath(s.filePath)
	return s
}

func (fs *FileStorage) Store(key string, data []byte) error {
	return fs.writeFile(key, data)
}

func (fs *FileStorage) Load(key string) ([]byte, error) {
	f, err := os.Open(filepath.Join(fs.filePath, key))
	if err != nil {
		klog.Errorf("open cache file %s error: %v", key, err)
		return []byte{}, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		klog.Errorf("read cache %s error: %v", key, err)
		return []byte{}, err
	}
	return b, nil
}

func (fs *FileStorage) randomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	var result []byte
	for i := 0; i < l; i++ {
		result = append(result, bytes[fs.seed.Intn(len(bytes))])
	}
	return string(result)
}

func (fs *FileStorage) writeFile(hFileName string, data []byte) error {
	salt := fs.randomString(12)
	tmpFileName := fmt.Sprintf("%s_%s", hFileName, salt)
	f, err := os.Create(filepath.Join(fs.filePath, tmpFileName))
	if err != nil {
		klog.Errorf("create file %s error: %v", tmpFileName, err)
		return err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			klog.Errorf("rename tmp to target file error: %v", err)
		}
		err = os.Rename(filepath.Join(fs.filePath, tmpFileName), filepath.Join(fs.filePath, hFileName))
		if err != nil {
			klog.Errorf("rename tmp to target file error: %v", err)
		}
	}()

	_, err = f.Write(data)
	if err != nil {
		klog.Errorf("write file %s error: %v", tmpFileName, err)
		return err
	}
	return nil

}

func mkdirPath(filePath string) {
	if _, err := os.Lstat(filePath); err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			klog.Fatalf("mkdir %s error : %v", filePath, err)
		}
	}
}
