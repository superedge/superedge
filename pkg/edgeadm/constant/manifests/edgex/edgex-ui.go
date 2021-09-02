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

package edgex

//ui for edgex, not necessary
const EDGEX_UI = "edgex-ui.yml"

const EDGEX_UI_YAML = `
apiVersion: v1
kind: Service
metadata:
  name: edgex-ui-go
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: edgex-ui-go
  ports:
  - name: http
    port: 4000
    protocol: TCP
    targetPort: 4000
    nodePort: 30040
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-ui-go
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: edgex-ui-go
  template:
    metadata:
      namespace: {{.Namespace}}
      labels: 
        app: edgex-ui-go
    spec:
      hostname: edgex-ui-go
      containers:
      - name: edgex-ui-go
        image: edgexfoundry/docker-edgex-ui-go:1.3.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          protocol: TCP
          containerPort: 4000
`