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

package manifests

const APP_LITE_APISERVER = "lite-apiserver.yaml"

const LiteApiServerYaml = `
apiVersion: v1
kind: Pod
metadata:
  labels:
    k8s-app: lite-apiserver
  name: lite-apiserver
  namespace: {{.Namespace}}
spec:
  containers:
    - command:
        - lite-apiserver
        - --ca-file=/etc/kubernetes/pki/ca.crt
        - --tls-cert-file=/etc/kubernetes/edge/lite-apiserver.crt
        - --tls-private-key-file=/etc/kubernetes/edge/lite-apiserver.key
        - --kube-apiserver-url={{.MasterIP}}
        - --kube-apiserver-port=6443
        - --port=51003
        - --tls-config-file=/etc/kubernetes/edge/tls.json
        - --v=4
        - --file-cache-path=/data/lite-apiserver/cache
        - --timeout=3
      image: superedge.tencentcloudcr.com/superedge/lite-apiserver:v0.5.0
      imagePullPolicy: IfNotPresent
      name: lite-apiserver
      volumeMounts:
        - mountPath: /etc/kubernetes/pki
          name: k8s-certs
        - mountPath: /etc/kubernetes/edge
          name: edge-certs
          readOnly: true
        - mountPath: /var/lib/kubelet/pki
          name: kubelet
          readOnly: true
        - mountPath: /data
          name: cache
  hostNetwork: true
  volumes:
    - hostPath:
        path: /var/lib/kubelet/pki
        type: DirectoryOrCreate
      name: kubelet
    - hostPath:
        path: /data
      name: cache
    - hostPath:
        path: /etc/kubernetes/pki
        type: DirectoryOrCreate
      name: k8s-certs
    - hostPath:
        path: /etc/kubernetes/edge
        type: DirectoryOrCreate
      name: edge-certs
status: {}
`
