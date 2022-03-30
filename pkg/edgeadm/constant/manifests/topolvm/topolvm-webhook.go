/*
Copyright 2020 The topolvm Authors.
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

const AppTopolvmWebhook = "topolvm-webhook.yaml"

const AppTopolvmWebhookYaml = `
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: topolvm-hook
  annotations:
    cert-manager.io/inject-ca-from: topolvm-system/topolvm-mutatingwebhook
  labels:
    app.kubernetes.io/version: "0.10.0"
    app.kubernetes.io/name: topolvm-hook
webhooks:
  - name: pvc-hook.topolvm.cybozu.com
    admissionReviewVersions:
      - "v1"
      - "v1beta1"
    namespaceSelector:
      matchExpressions:
        - key: topolvm.cybozu.com/webhook
          operator: NotIn
          values: ["ignore"]
    failurePolicy: Fail
    matchPolicy: Equivalent
    clientConfig:
      caBundle: {{.CABundle}}
      service:
        namespace: topolvm-system
        name: topolvm-controller
        path: /pvc/mutate
    rules:
      - operations: ["CREATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["persistentvolumeclaims"]
    sideEffects: None
  - name: pod-hook.topolvm.cybozu.com
    admissionReviewVersions:
      - "v1"
      - "v1beta1"
    namespaceSelector:
      matchExpressions:
        - key: topolvm.cybozu.com/webhook
          operator: NotIn
          values: ["ignore"]
    failurePolicy: Fail
    matchPolicy: Equivalent
    clientConfig:
      caBundle: {{.CABundle}}
      service:
        namespace: topolvm-system
        name: topolvm-controller
        path: /pod/mutate
    rules:
      - operations: ["CREATE"]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    sideEffects: None 
---
apiVersion: v1
data:
  tls.crt: {{.ServerCrt}}
  tls.key: {{.ServerKey}}
kind: Secret
metadata:
  name: topolvm-mutatingwebhook
  namespace: {{.Namespace}}
---

# Source: topolvm/templates/controller/csidriver.yaml
apiVersion: storage.k8s.io/v1
kind: CSIDriver
metadata:
  name: topolvm.cybozu.com
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.10.0"
spec:
  attachRequired: false
  podInfoOnMount: true
  volumeLifecycleModes:
    - Persistent
    - Ephemeral
---
# Source: topolvm/templates/storageclass.yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: topolvm-provisioner
  annotations:
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.10.0"
provisioner: topolvm.cybozu.com
parameters:
  "csi.storage.k8s.io/fstype": "xfs"
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
`
