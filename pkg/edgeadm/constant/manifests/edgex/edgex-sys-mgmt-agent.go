/*
Copyright 2020 The edgex foundry Authors.
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

package edgex

//The components in edgex system management
const EDGEX_SYS_MGMT = "edgex-system-management.yml"

const EDGEX_SYS_MGMT_YAML = `
apiVersion: v1
kind: Service
metadata:
  name: edgex-sys-mgmt-agent
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: edgex-sys-mgmt-agent
  ports:
  - name: http
    port: 48090
    protocol: TCP
    targetPort: 48090
    nodePort: 30990
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-sys-mgmt-agent
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: edgex-sys-mgmt-agent
  template:
    metadata:
      labels: 
        app: edgex-sys-mgmt-agent
    spec:
      hostname: edgex-sys-mgmt-agent
      volumes:
        - name: edgex-sys-mgmt-agent
          hostPath:
            path: /var/run/
            type: DirectoryOrCreate
      containers:
      - name: edgex-sys-mgmt-agent
        image: edgexfoundry/docker-sys-mgmt-agent-go:1.3.0
        imagePullPolicy: IfNotPresent
        envFrom: 
        - configMapRef:
            name: common-variables
        env:
          - name: EXECUTORPATH
            value: "/sys-mgmt-executor"
          - name: METRICSMECHANISM
            value: "executor"
          - name: SERVICE_HOST
            value: "edgex-sys-mgmt-agent"
        ports:
        - name: http
          protocol: TCP
          containerPort: 48090
        volumeMounts:
        - name: edgex-sys-mgmt-agent
          mountPath: /var/run/docker.sock
          subPath: docker.sock
`
