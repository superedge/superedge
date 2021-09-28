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

const AppTopolvmLvmd = "topolvm-lvmd.yaml"

const AppTopolvmLvmdYaml = `
# Source: topolvm/templates/lvmd/psp.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: topolvm-lvmd
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
spec:
  privileged: true
  hostPID: true
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'hostPath'
    - 'secret'
  allowedHostPaths: 
  - pathPrefix: /run/topolvm
    readOnly: false
  runAsUser:
    rule: 'RunAsAny'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
---

## psp:topolvm-lvmd RBAC
# Source: topolvm/templates/lvmd/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: topolvm-lvmd
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
---
# Source: topolvm/templates/lvmd/role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: psp:topolvm-lvmd
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
rules:
  - apiGroups: ["policy"]
    resources: ["podsecuritypolicies"]
    verbs: ["use"]
    resourceNames: ["topolvm-lvmd"]
---
# Source: topolvm/templates/lvmd/rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: topolvm-lvmd:psp:topolvm-lvmd
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
subjects:
  - kind: ServiceAccount
    name: topolvm-lvmd
    namespace: topolvm-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: psp:topolvm-lvmd
---

# Source: topolvm/templates/lvmd/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: topolvm-lvmd
  namespace: topolvm-system
  labels:
    idx: "0"
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
data:
  lvmd.yaml: |
    socket-name: /run/topolvm/lvmd.sock
    device-classes: 
      - default: true
        name: ssd
        spare-gb: 10
        volume-group: myvg1
---

# Source: topolvm/templates/lvmd/daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: topolvm-lvmd
  namespace: topolvm-system
  labels:
    idx: "0"
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
spec:
  selector:
    matchLabels:
      idx: "0"
      app.kubernetes.io/name: topolvm-lvmd
  template:
    metadata:
      labels:
        idx: "0"
        app.kubernetes.io/name: topolvm-lvmd
      annotations:
        checksum/config: 61a60308315b12849f6c11282006c292a18466c65d7201d56f856c0c90a94f5e
        prometheus.io/port: "8080"
    spec:
      containers:
        - name: lvmd
          image: "superedge.tencentcloudcr.com/superedge/topolvm-with-sidecar:0.10.0"
          securityContext:
            privileged: true
          command:
            - /lvmd
            - --container
          volumeMounts:
            - name: config
              mountPath: /etc/topolvm
            - mountPath: /run/topolvm
              name: lvmd-socket-dir
      hostPID: true
      nodeSelector:
        superedge.io/local.pv: "topolvm"
      serviceAccountName: topolvm-lvmd
      volumes:
        - name: config
          configMap:
            name: topolvm-lvmd
        - hostPath:
            path: /run/topolvm
            type: DirectoryOrCreate
          name: lvmd-socket-dir
`
