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

const APP_APPLICATION_GRID_CONTROLLER = "application-grid-controller.yaml"

const ApplicationGridControllerYaml = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: application-grid-controller
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: application-grid-controller
rules:
  - apiGroups:
    - apiextensions.k8s.io
    resources:
      - customresourcedefinitions
    verbs:
      - "*"
  - apiGroups:
    - ""
    resources:
      - nodes
      - services
    verbs:
      - "*"
  - apiGroups:
    - extensions
    - apps
    resources:
      - deployments
    verbs:
      - "*"
  - apiGroups:
    - superedge.io
    resources:
      - deploymentgrids
      - servicegrids
      - deploymentgrids/status
      - deploymentgridlists
      - servicegridlists
    verbs:
      - "*"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: application-grid-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: application-grid-controller
subjects:
  - kind: ServiceAccount
    name: application-grid-controller
    namespace: kube-system

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: application-grid-controller
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: application-grid-controller
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: application-grid-controller
    spec:
      serviceAccountName: application-grid-controller
      containers:
        - name: application-grid-controller
          image: superedge/application-grid-controller:v0.1.0
          imagePullPolicy: IfNotPresent
          command:
            - /usr/local/bin/application-grid-controller
          resources:
            limits:
              cpu: 50m
              memory: 100Mi
            requests:
              cpu: 10m
              memory: 20Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
        - key: "node-role.kubernetes.io/master"
          operator: "Exists"
          effect: "NoSchedule"
`
