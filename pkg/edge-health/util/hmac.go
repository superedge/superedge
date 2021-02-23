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
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/superedge/superedge/pkg/edge-health/common"
	"github.com/superedge/superedge/pkg/edge-health/metadata"
	"golang.org/x/sys/unix"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/informers/core/v1"
	"os"
	"os/signal"
)

func GenerateHmac(communInfo metadata.CommunInfo, cmInformer corev1.ConfigMapInformer) (string, error) {
	addrBytes, err := json.Marshal(communInfo.SourceIP)
	if err != nil {
		return "", err
	}
	detailBytes, _ := json.Marshal(communInfo.CheckDetail)
	if err != nil {
		return "", err
	}
	hmacBefore := string(addrBytes) + string(detailBytes)
	if hmacConf, err := cmInformer.Lister().ConfigMaps(metav1.NamespaceSystem).Get(common.HmacConfig); err != nil {
		return "", err
	} else {
		return GetHmacCode(hmacBefore, hmacConf.Data[common.HmacKey])
	}
}

func GetHmacCode(s, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))
	if _, err := io.WriteString(h, s); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func SignalWatch() (context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, unix.SIGTERM, unix.SIGINT)
	go func() {
		for range signals {
			cancelFunc()
		}
	}()
	return ctx, cancelFunc
}
