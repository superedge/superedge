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

package token

import (
	"bufio"
	"encoding/json"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"sync"
	"time"
)

var tokenData *TokenCache

const (
	DEFAULT = "default"
)

type TokenCache struct {
	Tokens map[string]string
	Lock   sync.RWMutex
}

type Token struct {
	NodeName string `json:"nodename"`
	Token    string `json:"token"`
}

func GetTokenFromCache(nodeName string) string {
	defer tokenData.Lock.RUnlock()
	tokenData.Lock.RLock()
	token, ok := tokenData.Tokens[nodeName]
	if !ok {
		token = tokenData.Tokens[DEFAULT]
	}
	return token
}

func InitTokenCache(file string) error {
	tokenData = &TokenCache{
		Tokens: nil,
		Lock:   sync.RWMutex{},
	}
	err := GetTokenFromFile(file)
	if err != nil {
		klog.Error("failed to read client token from file")
		return err
	}
	go func() {
		stop := make(chan struct{}, 1)
		for {
			select {
			case <-stop:
				return
			case <-time.After(time.Duration(10) * time.Second):
				err := GetTokenFromFile(file)
				if err != nil {
					klog.Error("failed to read client token from file")
					stop <- struct{}{}
				}
			}
		}
	}()
	return nil
}

func GetTonken(nodeName, token string) (string, error) {
	ttoken := &Token{
		NodeName: nodeName,
		Token:    token,
	}
	data, err := json.Marshal(ttoken)
	if err != nil {
		klog.Error("get token fail !")
		return "", err
	}
	return string(data), nil

}

func GetTokenFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		klog.Errorf("open file fail !")
		return err
	}
	defer f.Close()
	tokenMap := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ParseLine(scanner.Text(), tokenMap)
	}
	tokenData.Lock.Lock()
	tokenData.Tokens = tokenMap
	tokenData.Lock.Unlock()
	return nil
}

func ParseToken(token string) (*Token, error) {
	rtoken := &Token{}
	err := json.Unmarshal([]byte(token), rtoken)
	if err != nil {
		return rtoken, err
	}
	return rtoken, nil
}

func ParseLine(line string, m map[string]string) {
	line = util.ReplaceString(line)
	kv := strings.Split(line, ":")
	if len(kv) < 2 {
		return
	}
	m[kv[0]] = kv[1]
}
