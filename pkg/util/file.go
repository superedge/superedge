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

package util

import (
	"bufio"
	"io/ioutil"
	"os"
)

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func RemoveFile(filePath string) error {
	if IsFileExist(filePath) {
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}
	return nil
}

func WriteFile(file, ctx string) error {
	return ioutil.WriteFile(file, []byte(ctx), 0664)
}

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

func WriteWithBufio(name, content string) error {
	fileObj, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fileObj.Close()

	buf := []byte(content)
	writeObj := bufio.NewWriterSize(fileObj, 4096)
	if _, err := writeObj.Write(buf); err != nil {
		return err
	}
	if err := writeObj.Flush(); err != nil {
		return err
	}

	return nil
}

func WriteWithAppend(name, content string) error {
	fileObj, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer fileObj.Close()

	buf := []byte(content)
	writeObj := bufio.NewWriterSize(fileObj, 4096)
	if _, err := writeObj.Write(buf); err != nil {
		return err
	}
	if err := writeObj.Flush(); err != nil {
		return err
	}

	return nil
}

func CopyFile(sourceFile, destinationFile string) error {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(destinationFile, input, 0644)
	if err != nil {
		return err
	}
	return nil
}
