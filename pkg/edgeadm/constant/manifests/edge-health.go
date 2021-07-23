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

const APP_EDGE_HEALTH = "edge-health.yaml"

const EdgeHealthYaml = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: edge-health
  namespace: {{.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: edge-health
rules:
  - apiGroups:
      - ""
    resources:
      - nodes
      - nodes/proxy
    verbs:
      - get
      - list
      - watch
      - update
      - patch
  - apiGroups:
      - ""
    resources:
      - configmaps
    verbs:
      - get
      - list
      - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: edge-health
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: edge-health
subjects:
  - kind: ServiceAccount
    name: edge-health
    namespace: {{.Namespace}}

---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: edge-health
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels:
      name: edge-health
  template:
    metadata:
      labels:
        name: edge-health
    spec:
      serviceAccountName: edge-health
      containers:
        - name: edge-health
          image: superedge.tencentcloudcr.com/superedge/edge-health:v0.5.0
          imagePullPolicy: IfNotPresent
          resources:
            limits: 
              cpu: 50m
              memory: 100Mi
            requests: 
              cpu: 10m
              memory: 20Mi
          command:
            - edge-health
          args:
            - --kubeletauthplugin=timeout=5,retrytime=3,weight=1,port=10250
            - --v=2
          env:
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
          securityContext:
            procMount: Default
      dnsPolicy: ClusterFirst
      hostNetwork: true
      nodeSelector:
        superedge.io/edge-node: enable
      restartPolicy: Always
      securityContext: {}
      terminationGracePeriodSeconds: 30
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: hmac-config
  namespace: {{.Namespace}}
data:
  hmackey: {{.HmacKey}}
`
