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

//A device-random for test
const EDGEX_DEVICE_RANDOM = "edgex-device-random.yml"

const EDGEX_DEVICE_RANDOM_YAML = `
apiVersion: v1
kind: Service
metadata:
  name: edgex-device-random
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: edgex-device-random
  ports:
  - name: http
    port: 49988
    protocol: TCP
    targetPort: 49988
    nodePort: 30088
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-device-random
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: edgex-device-random
  template:
    metadata:
      labels: 
        app: edgex-device-random
    spec:
      hostname: edgex-device-random
      containers:
      - name: edgex-device-random
        image: edgexfoundry/docker-device-random-go:1.3.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          protocol: TCP
          containerPort: 49988
        envFrom: 
        - configMapRef:
            name: common-variables
        env:
          - name: Service_Host
            value: "edgex-device-random"
`