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

//The components in application services
const EDGEX_APP = "edgex-application-services.yml"

const EDGEX_APP_YAML = `
apiVersion: v1
kind: Service
metadata:
  name: edgex-app-service-configurable-rules
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: edgex-app-service-configurable-rules
  ports:
  - name: http
    port: 48100
    protocol: TCP
    targetPort: 48100
    nodePort: 30100
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-app-service-configurable-rules
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: edgex-app-service-configurable-rules
  template:
    metadata:
      labels: 
        app: edgex-app-service-configurable-rules
    spec:
      hostname: edgex-app-service-configurable-rules
      containers:
      - name: edgex-app-service-configurable-rules
        image: edgexfoundry/docker-app-service-configurable:1.3.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          protocol: TCP
          containerPort: 48100
        envFrom: 
        - configMapRef:
            name: common-variables
        env: 
        - name: Service_Host
          value: "edgex-app-service-configurable-rules"
        - name: Service_Port
          value: "48100"
        - name: edgex_profile
          value: "rules-engine"
        - name: MessageBus_SubscribeHost_Host
          value: "edgex-core-data"
        - name: Binding_PublishTopic
          value: "events"
`
