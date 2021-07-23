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

const APP_EDGE_HEALTH_ADMISSION = "edge-health-admission.yaml"

const EdgeHealthAdmissionYaml = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: edge-health-admission
  namespace: {{.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: edge-health-admission
rules:
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - "*"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: edge-health-admission
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: edge-health-admission
subjects:
  - kind: ServiceAccount
    name: edge-health-admission
    namespace: {{.Namespace}}
---
apiVersion: v1
data:
  server.crt: {{.ServerCrt}}
  server.key: {{.ServerKey}}
kind: Secret
metadata:
  name: validate-admission-control-server-certs
  namespace: {{.Namespace}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: edge-health-admission
  namespace: {{.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: edge-health-admission
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        k8s-app: edge-health-admission
    spec:
      containers:
        - name: edge-health-admission
          image: superedge.tencentcloudcr.com/superedge/edge-health-admission:v0.5.0
          imagePullPolicy: Always
          command:
            - edge-health-admission
          args:
            - --alsologtostderr
            - --v=4
            - --admission-control-server-cert=/etc/edge-health-admission/certs/server.crt
            - --admission-control-server-key=/etc/edge-health-admission/certs/server.key
          env:
            - name: MASTER_NAMESPACE
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: metadata.namespace
          volumeMounts:
            - mountPath: /etc/edge-health-admission/certs
              name: admission-server-certs
          resources:
            limits:
              cpu: 50m
              memory: 100Mi
            requests:
              cpu: 10m
              memory: 20Mi
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      volumes:
        - secret:
            secretName: validate-admission-control-server-certs
          name: admission-server-certs
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
      serviceAccountName: edge-health-admission
---
apiVersion: v1
kind: Service
metadata:
  name: edge-health-admission
  namespace: {{.Namespace}}
spec:
  ports:
    - port: 443
      protocol: TCP
      targetPort: 8443
  selector:
    k8s-app: edge-health-admission
  type: ClusterIP
`
