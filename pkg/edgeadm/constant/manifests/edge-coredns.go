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

const APPEdgeCorednsConfig = "edge-coredns-config.yaml"

const EdgeCorednsConfigYaml = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: edge-coredns
  namespace: {{.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:edge-coredns
rules:
  - apiGroups:
    - ""
    resources:
    - endpoints
    - services
    - pods
    - namespaces
    verbs:
    - list
    - watch
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
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:edge-coredns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:edge-coredns
subjects:
- kind: ServiceAccount
  name: edge-coredns
  namespace: {{.Namespace}}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: edge-coredns
  namespace: {{.Namespace}}
data:
  Corefile: |
    .:53 {
        errors
        health {
          lameduck 5s
        }
        hosts /etc/edge/hosts {
            reload 300ms
            fallthrough
        }
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload 2s
        loadbalance
    }
`

const APPEdgeCorednsDeploymentGrid = "edge-coredns-deployment-grid.yaml"

const EdgeCorednsDeploymentGridYaml = `
apiVersion: superedge.io/v1
kind: DeploymentGrid
metadata:
  name: edge-coredns
  namespace: {{.Namespace}}
spec:
  gridUniqKey: superedge.io.hostname
  template:
    replicas: 1
    selector:
      matchLabels:
        k8s-app: edge-coredns
    strategy: {}
    template:
      metadata:
        labels:
          k8s-app: edge-coredns
      selector:
        matchLabels:
          k8s-app: edge-coredns
      spec:
        priorityClassName: system-cluster-critical
        serviceAccountName: edge-coredns
        tolerations:
          - key: "CriticalAddonsOnly"
            operator: "Exists"
        nodeSelector:
          superedge.io/edge-node: enable
        containers:
        - name: coredns
          image: superedge.tencentcloudcr.com/superedge/coredns:1.6.9
          imagePullPolicy: IfNotPresent
          resources:
            limits:
              cpu: 50m
              memory: 100Mi
            requests:
              cpu: 10m
              memory: 20Mi
          args: [ "-conf", "/etc/coredns/Corefile" ]
          volumeMounts:
          - name: config-volume
            mountPath: /etc/coredns
            readOnly: true
          ports:
          - containerPort: 53
            name: dns
            protocol: UDP
          - containerPort: 53
            name: dns-tcp
            protocol: TCP
          - containerPort: 9153
            name: metrics
            protocol: TCP
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              add:
              - NET_BIND_SERVICE
              drop:
              - all
            readOnlyRootFilesystem: true
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
              scheme: HTTP
            initialDelaySeconds: 60
            timeoutSeconds: 5
            successThreshold: 1
            failureThreshold: 5
          readinessProbe:
            httpGet:
              path: /ready
              port: 8181
              scheme: HTTP
        dnsPolicy: Default
        volumes:
          - name: config-volume
            configMap:
              name: edge-coredns
              items:
              - key: Corefile
                path: Corefile
`

const APPEdgeCorednsServiceGrid = "edge-coredns-service-grid.yaml"

const EdgeCorednsServiceGridYaml = `
apiVersion: superedge.io/v1
kind: ServiceGrid
metadata:
  name: edge-coredns
  namespace: {{.Namespace}}
  annotations:
    prometheus.io/port: "9153"
    prometheus.io/scrape: "true"
  labels:
    k8s-app: edge-coredns
    kubernetes.io/name: "CoreDNS"
    kubernetes.io/cluster-service: "true"
spec:
  gridUniqKey: superedge.io.hostname
  template:
    selector:
      k8s-app: edge-coredns
    ports:
    - name: dns
      port: 53
      protocol: UDP
    - name: dns-tcp
      port: 53
      protocol: TCP
    - name: metrics
      port: 9153
      protocol: TCP
`
