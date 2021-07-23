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

const APP_APPLICATION_GRID_WRAPPER = "application-grid-wrapper.yaml"

const ApplicationGridWrapperYaml = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: application-grid-wrapper
  namespace: {{.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: application-grid-wrapper
rules:
  - apiGroups:
      - ""
    resources:
      - endpoints
      - services
    verbs:
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ""
      - events.k8s.io
    resources:
      - events
    verbs:
      - create
      - patch
      - update
  - apiGroups:
      - discovery.k8s.io
    resources:
      - endpointslices
    verbs:
      - list
      - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: application-grid-wrapper
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: application-grid-wrapper
subjects:
  - kind: ServiceAccount
    name: application-grid-wrapper
    namespace: {{.Namespace}}
---
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: application-grid-wrapper
  name: application-grid-wrapper
  namespace: {{.Namespace}}
data:
  kubeconfig.conf: |
    apiVersion: v1
    clusters:
    - cluster:
        certificate-authority: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        server: https://127.0.0.1:51003
      name: default
    contexts:
    - context:
        cluster: default
        namespace: default
        user: default
      name: default
    current-context: default
    kind: Config
    preferences: {}
    users:
    - name: default
      user:
        tokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token

---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: application-grid-wrapper-node
    addonmanager.kubernetes.io/mode: Reconcile
  name: application-grid-wrapper-node
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels:
      k8s-app: application-grid-wrapper-node
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 10%
  template:
    metadata:
      labels:
        k8s-app: application-grid-wrapper-node
    spec:
      serviceAccount: application-grid-wrapper
      serviceAccountName: application-grid-wrapper
      priorityClassName: system-node-critical
      hostNetwork: true
      restartPolicy: Always
      nodeSelector:
        kubernetes.io/os: linux 
        superedge.io/edge-node: enable   # TODO select nodes labeled as edges
      containers:
        - name: application-grid-wrapper
          image: superedge.tencentcloudcr.com/superedge/application-grid-wrapper:v0.5.0
          imagePullPolicy: IfNotPresent
          command:
            - /usr/local/bin/application-grid-wrapper
            - --kubeconfig=/var/lib/application-grid-wrapper/kubeconfig.conf
            - --bind-address=127.0.0.1:51006
            - --hostname=$(NODE_NAME)
            - --notify-channel-size=10000
          env:
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
          resources:
            limits:
              cpu: 50m
              memory: 100Mi
            requests:
              cpu: 10m
              memory: 20Mi
          securityContext:
            privileged: true
          volumeMounts:
            - mountPath: /var/lib/application-grid-wrapper
              name: application-grid-wrapper
      volumes:
        - configMap:
            defaultMode: 420
            name: application-grid-wrapper
          name: application-grid-wrapper
        - hostPath:
            path: /var/tmp
            type: Directory
          name: host-var-tmp
---
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: application-grid-wrapper-master
  name: application-grid-wrapper-master
  namespace: {{.Namespace}}
data:
  kubeconfig.conf: |
    apiVersion: v1
    clusters:
    - cluster:
        certificate-authority: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        server: {{.AdvertiseAddress}}
      name: default
    contexts:
    - context:
        cluster: default
        namespace: default
        user: default
      name: default
    current-context: default
    kind: Config
    preferences: {}
    users:
    - name: default
      user:
        tokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token

---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    k8s-app: application-grid-wrapper-master
    addonmanager.kubernetes.io/mode: Reconcile
  name: application-grid-wrapper-master
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels:
      k8s-app: application-grid-wrapper-master
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 10%
  template:
    metadata:
      labels:
        k8s-app: application-grid-wrapper-master
    spec:
      serviceAccount: application-grid-wrapper
      serviceAccountName: application-grid-wrapper
      priorityClassName: system-node-critical
      hostNetwork: true
      restartPolicy: Always
      nodeSelector:
        node-role.kubernetes.io/master: "" # TODO select master node
      tolerations:
        - key: "node-role.kubernetes.io/master"
          operator: "Exists"
          effect: "NoSchedule"
        - key: "CriticalAddonsOnly"
          operator: "Exists"
        - operator: "Exists"
      containers:
        - name: application-grid-wrapper
          image: superedge.tencentcloudcr.com/superedge/application-grid-wrapper:v0.5.0
          imagePullPolicy: IfNotPresent
          command:
            - /usr/local/bin/application-grid-wrapper
            - --kubeconfig=/var/lib/application-grid-wrapper/kubeconfig.conf
            - --bind-address=127.0.0.1:51006
            - --hostname=$(NODE_NAME)
            - --wrapper-in-cluster=false
            - --notify-channel-size=10000
          env:
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
          resources:
            limits:
              cpu: 50m
              memory: 100Mi
            requests:
              cpu: 10m
              memory: 20Mi
          securityContext:
            privileged: true
          volumeMounts:
            - mountPath: /var/lib/application-grid-wrapper
              name: application-grid-wrapper-master
      volumes:
        - configMap:
            defaultMode: 420
            name: application-grid-wrapper-master
          name: application-grid-wrapper-master
        - hostPath:
            path: /var/tmp
            type: Directory
          name: host-var-tmp
`
