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

//The components in supporting services
const EDGEX_SUPPORT = "edgex-support-services.yml"

const EDGEX_SUPPORT_YAML = `
apiVersion: v1
kind: Service
metadata:
  name: edgex-support-notifications
  namespace: {{.Namespace}}
spec:
  type: NodePort 
  selector:
    app: edgex-support-notifications
  ports:
  - name: http
    port: 48060
    protocol: TCP
    targetPort: 48060
    nodePort: 30060
---
apiVersion: v1
kind: Service
metadata:
 name: edgex-kuiper
 namespace: {{.Namespace}}
spec:
 type: NodePort
 selector:
   app: edgex-kuiper
 ports:
 - name: tcp-48075
   port: 48075
   protocol: TCP
   targetPort: 48075
   nodePort: 30075
 - name: tcp-20498
   port: 20498
   protocol: TCP
   targetPort: 20498
   nodePort: 30098
---
apiVersion: v1
kind: Service
metadata:
  name: edgex-support-scheduler
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: edgex-support-scheduler
  ports:
  - name: http
    port: 48085
    protocol: TCP
    targetPort: 48085  
    nodePort: 30085
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-support-notifications
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: edgex-support-notifications
  template:
    metadata:
      labels: 
        app: edgex-support-notifications
    spec:
      hostname: edgex-support-notifications
      containers:
      - name: edgex-support-notifications
        image: edgexfoundry/docker-support-notifications-go:1.3.0
        imagePullPolicy: IfNotPresent
        envFrom: 
        - configMapRef:
            name: common-variables
        env: 
        - name: Service_Host
          value: "edgex-support-notifications"
        ports:
        - name: http
          protocol: TCP
          containerPort: 48060        
---
apiVersion: apps/v1
kind: Deployment
metadata: 
 namespace: {{.Namespace}}
 name: edgex-kuiper
spec:
 selector:
   matchLabels: 
     app: edgex-kuiper
 template:
   metadata:
     labels: 
       app: edgex-kuiper
   spec:
     hostname: edgex-kuiper
     containers:
     - name: edgex-kuiper
       image: emqx/kuiper:1.0.0-alpine
       imagePullPolicy: IfNotPresent
       ports:
       - name: tcp-48075
         protocol: TCP
         containerPort: 48075
       - name: tcp-20498
         protocol: TCP
         containerPort: 20498
       env:
         - name: KUIPER__BASIC__CONSOLELOG
           value: "true"
         - name: KUIPER__BASIC__RESTPORT
           value: "48075"
         - name: EDGEX__DEFAULT__SERVER
           value: "edgex-app-service-configurable-rules"
         - name: EDGEX__DEFAULT__SERVICESERVER
           value: "http://edgex-core-data:48080"
         - name: EDGEX__DEFAULT__TOPIC
           value: "events"
         - name: EDGEX__DEFAULT__PROTOCOL
           value: "tcp"
         - name: EDGEX__DEFAULT__PORT
           value: "5566"
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-support-scheduler 
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: edgex-support-scheduler
  template:
    metadata:
      labels: 
        app: edgex-support-scheduler
    spec:
      hostname: edgex-support-scheduler
      containers:
      - name: edgex-support-scheduler
        image: edgexfoundry/docker-support-scheduler-go:1.3.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          protocol: TCP
          containerPort: 48085
        envFrom: 
        - configMapRef:
            name: common-variables
        env: 
        - name: Service_Host
          value: "edgex-support-scheduler"
        - name: IntervalActions_ScrubPushed_Host
          value: "edgex-core-data"
        - name: IntervalActions_ScrubAged_Host
          value: "edgex-core-data"
`
