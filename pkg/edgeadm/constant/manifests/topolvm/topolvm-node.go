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

const AppTopolvmNode = "topolvm-node.yaml"

const AppTopolvmNodeYaml = `
## topolvm-node RBAC
# Source: topolvm/templates/node/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: topolvm-node
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
---
# Source: topolvm/templates/node/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: topolvm-system:node
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch", "update", "patch"]
  - apiGroups: ["topolvm.cybozu.com"]
    resources: ["logicalvolumes", "logicalvolumes/status"]
    verbs: ["get", "list", "watch", "create", "update", "delete", "patch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["csidrivers"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["policy"]
    resources: ["podsecuritypolicies"]
    verbs: ["use"]
    resourceNames: ["topolvm-node"]
---
# Source: topolvm/templates/node/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: topolvm-system:node
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
subjects:
  - kind: ServiceAccount
    name: topolvm-node
    namespace: topolvm-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: topolvm-system:node
---

# Source: topolvm/templates/node/psp.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: topolvm-node
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
spec:
  privileged: true
  allowPrivilegeEscalation: true
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'secret'
    - 'hostPath'
  allowedHostPaths: 
  - pathPrefix: /var/lib/kubelet
    readOnly: false
  - pathPrefix: /run/topolvm
    readOnly: false
  hostNetwork: false
  runAsUser:
    rule: 'RunAsAny'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: true
---

# Source: topolvm/templates/node/daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: topolvm-node
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: topolvm-node
  template:
    metadata:
      labels:
        app.kubernetes.io/name: topolvm-node
      annotations: 
        prometheus.io/port: "8080"
    spec:
      containers:
        - name: topolvm-node
          image: "quay.io/topolvm/topolvm-with-sidecar:0.10.0"
          securityContext: 
            privileged: true
          command:
            - /topolvm-node
            - --lvmd-socket=/run/lvmd/lvmd.sock
          ports:
            - containerPort: 9808
              name: healthz
              protocol: TCP
          livenessProbe:
            httpGet:
              path: /healthz
              port: healthz
            failureThreshold: 3
            initialDelaySeconds: 10
            timeoutSeconds: 3
            periodSeconds: 60
          env:
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          volumeMounts: 
            - mountPath: /run/topolvm
              name: node-plugin-dir
            - mountPath: /run/lvmd
              name: lvmd-socket-dir
            - mountPath: /var/lib/kubelet/pods
              mountPropagation: Bidirectional
              name: pod-volumes-dir
            - mountPath: /var/lib/kubelet/plugins/kubernetes.io/csi
              mountPropagation: Bidirectional
              name: csi-plugin-dir
        - name: csi-registrar
          image: "quay.io/topolvm/topolvm-with-sidecar:0.10.0"
          command:
            - /csi-node-driver-registrar
            - "--csi-address=/run/topolvm/csi-topolvm.sock"
            - "--kubelet-registration-path=/var/lib/kubelet/plugins/topolvm.cybozu.com/node/csi-topolvm.sock"
          lifecycle:
            preStop:
              exec:
                command: ["/bin/sh", "-c", "rm -rf /registration/topolvm.cybozu.com /registration/topolvm.cybozu.com-reg.sock"]
          volumeMounts:
            - name: node-plugin-dir
              mountPath: /run/topolvm
            - name: registration-dir
              mountPath: /registration
        - name: liveness-probe
          image: "quay.io/topolvm/topolvm-with-sidecar:0.10.0"
          command:
            - /livenessprobe
            - "--csi-address=/run/topolvm/csi-topolvm.sock"
          volumeMounts:
            - name: node-plugin-dir
              mountPath: /run/topolvm
      nodeSelector:
        superedge.io/local.pv: "topolvm"
      serviceAccountName: topolvm-node
      volumes: 
        - hostPath:
            path: /var/lib/kubelet/plugins_registry/
            type: Directory
          name: registration-dir
        - hostPath:
            path: /var/lib/kubelet/plugins/topolvm.cybozu.com/node
            type: DirectoryOrCreate
          name: node-plugin-dir
        - hostPath:
            path: /var/lib/kubelet/plugins/kubernetes.io/csi
            type: DirectoryOrCreate
          name: csi-plugin-dir
        - hostPath:
            path: /var/lib/kubelet/pods/
            type: DirectoryOrCreate
          name: pod-volumes-dir
        - hostPath:
            path: /run/topolvm
            type: Directory
          name: lvmd-socket-dir
`
