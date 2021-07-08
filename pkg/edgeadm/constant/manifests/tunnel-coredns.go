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

const APP_TUNNEL_CORDDNS = "tunnel-coredns.yaml"

const TunnelCorednsYaml = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-coredns
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
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        reload 2s
        loadbalance
    }
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-nodes
  namespace: {{.Namespace}}
data:
  hosts: ""
---
apiVersion: v1
kind: Service
metadata:
  name: tunnel-coredns
  namespace: {{.Namespace}}
spec:
  ports:
    - name: dns
      port: 53
      protocol: UDP
      targetPort: 53
    - name: dns-tcp
      port: 53
      protocol: TCP
      targetPort: 53
    - name: metrics
      port: 9153
      protocol: TCP
      targetPort: 9153
  selector:
    k8s-app: tunnel-coredns
  type: ClusterIP
  clusterIP: {{.TunnelCoreDNSClusterIP}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tunnel-coredns
  namespace: {{.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: tunnel-coredns
  template:
    metadata:
      labels:
        k8s-app: tunnel-coredns
    spec:
      containers:
        - args:
            - -conf
            - /etc/coredns/Corefile
          image: superedge.tencentcloudcr.com/superedge/coredns:1.6.9
          imagePullPolicy: IfNotPresent
          livenessProbe:
            failureThreshold: 5
            httpGet:
              path: /health
              port: 8080
              scheme: HTTP
            initialDelaySeconds: 60
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 5
          name: tunnel-coredns
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
          readinessProbe:
            failureThreshold: 3
            httpGet:
              path: /ready
              port: 8181
              scheme: HTTP
          volumeMounts:
            - mountPath: /etc/coredns
              name: config-volume
              readOnly: true
            - mountPath: /etc/edge
              name: hosts
              readOnly: true
          resources:
            limits:
              cpu: 50m
              memory: 100Mi
            requests:
              cpu: 10m
              memory: 50Mi
      volumes:
        - configMap:
            defaultMode: 420
            items:
              - key: Corefile
                path: Corefile
            name: tunnel-coredns
          name: config-volume
        - configMap:
            defaultMode: 420
            name: tunnel-nodes
          name: hosts
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
        - key: "node-role.kubernetes.io/master"
          operator: "Exists"
          effect: "NoSchedule"
`
