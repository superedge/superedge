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

//The components in device services
const EDGEX_DEVICE = "edgex-device-services.yml"

const EDGEX_DEVICE_YAML = `
apiVersion: v1
kind: Service
metadata:
  name: edgex-device-rest
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: edgex-device-rest
  ports:
  - name: http
    port: 49986
    protocol: TCP
    targetPort: 49986
    nodePort: 30086
---
apiVersion: v1
kind: Service
metadata:
  name: edgex-device-virtual
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: edgex-device-virtual
  ports:
  - name: http
    port: 49990
    protocol: TCP
    targetPort: 49990
    nodePort: 30090
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-device-rest
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: edgex-device-rest
  template:
    metadata:
      namespace: {{.Namespace}}
      labels: 
        app: edgex-device-rest
    spec:
      hostname: edgex-device-rest
      containers:
      - name: edgex-device-rest
        image: edgexfoundry/docker-device-rest-go:1.2.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          protocol: TCP
          containerPort: 49986
        envFrom: 
        - configMapRef:
            name: common-variables
        env:
          - name: Service_Host
            value: "edgex-device-rest"
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  namespace: {{.Namespace}}
  name: edgex-device-virtual
spec:
  selector:
    matchLabels: 
      app: edgex-device-virtual
  template:
    metadata:
      namespace: {{.Namespace}}
      labels: 
        app: edgex-device-virtual
    spec:
      hostname: edgex-device-virtual
      containers:
      - name: edgex-device-virtual
        image: edgexfoundry/docker-device-virtual-go:1.3.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          protocol: TCP
          containerPort: 49990
        envFrom: 
        - configMapRef:
            name: common-variables
        env:
          - name: Service_Host
            value: "edgex-device-virtual"
`
