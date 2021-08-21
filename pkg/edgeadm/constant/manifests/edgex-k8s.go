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
const EDGEX = "k8s-hanoi-redis-no-secty.yml"

const EdgexYaml = `
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: {{.Namespace}}
  name: common-variables
data:
  EDGEX_SECURITY_SECRET_STORE: "false"
  Registry_Host: "edgex-core-consul"
  Clients_CoreData_Host: "edgex-core-data"
  Clients_Data_Host: "edgex-core-data"
  Clients_Notifications_Host: "edgex-support-notifications"
  Clients_Metadata_Host: "edgex-core-metadata"
  Clients_Command_Host: "edgex-core-command"
  Clients_Scheduler_Host: "edgex-support-scheduler"
  Clients_RulesEngine_Host: "edgex-kuiper"
  Clients_VirtualDevice_Host: "edgex-device-virtual"
  Databases_Primary_Host: "edgex-redis"
  Service_ServerBindAddr: "0.0.0.0"
  Logging_EnableRemote: "false"

---
apiVersion: v1
kind: Service
metadata:
  name: edgex-core-consul
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: edgex-core-consul
  ports:
  - name: tcp-8500
    port: 8500
    protocol: TCP
    targetPort: 8500
    nodePort: 30850
  - name: tcp-8400
    port: 8400
    protocol: TCP
    targetPort: 8400
    nodePort: 30840
---
apiVersion: v1
kind: Service
metadata:
  name: edgex-redis
  namespace: {{.Namespace}}
spec:
  selector:
    app: edgex-redis
  ports:
  - name: http
    protocol: TCP
    port: 6379
    targetPort: 6379  
---
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
  name: edgex-core-metadata
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: edgex-core-metadata
  ports:
  - name: http
    port: 48081
    protocol: TCP
    targetPort: 48081
    nodePort: 30081
---
apiVersion: v1
kind: Service
metadata:
  name: edgex-core-data
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: edgex-core-data
  ports:
  - name: tcp-5563
    port: 5563
    protocol: TCP
    targetPort: 5563
  - name: tcp-48080
    port: 48080
    protocol: TCP
    targetPort: 48080   
    nodePort: 30080
---
apiVersion: v1
kind: Service
metadata:
  name: edgex-core-command
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: edgex-core-command
  ports:
  - name: http
    port: 48082
    protocol: TCP
    targetPort: 48082   
    nodePort: 30082 
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
  name: edgex-core-consul
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: edgex-core-consul
  template:
    metadata:
      labels: 
        app: edgex-core-consul
    spec:
      hostname: edgex-core-consul
      volumes:
        - name: consul-config
          hostPath:
            path: /consul/config
            type: DirectoryOrCreate
        - name: consul-data
          hostPath:
            path: /consul/data
            type: DirectoryOrCreate
        - name: consul-scripts
          hostPath:
            path: /consul/scripts
            type: DirectoryOrCreate
      containers:
      - name: edgex-core-consul
        image: edgexfoundry/docker-edgex-consul:1.3.0
        imagePullPolicy: IfNotPresent
        env:
          - name: EDGEX_DB
            value: "redis"
          - name: EDGEX_SECURE
            value: "false"
        ports:
        - name: tcp-8500
          protocol: TCP
          containerPort: 8500
        - name: tcp-8400
          protocol: TCP
          containerPort: 8400
        volumeMounts:
        - name: consul-config
          mountPath: /consul/config
        - name: consul-data
          mountPath: /consul/data
        - name: consul-scripts
          mountPath: /consul/scripts
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-redis
  namespace: {{.Namespace}}
spec:
  selector: 
    matchLabels:
      app: edgex-redis
  template:
    metadata:
      labels:
        app: edgex-redis
    spec:
      hostname: edgex-redis
      volumes:
        - name: db-data
          hostPath:
            path: /data
            type: DirectoryOrCreate 
      containers:
      - name: edgex-redis
        image: redis:6.0.9-alpine
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 6379
        envFrom: 
        - configMapRef:
            name: common-variables
        volumeMounts:
        - name: db-data
          mountPath: /data
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
  name: edgex-core-metadata
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: edgex-core-metadata
  template:
    metadata:
      labels: 
        app: edgex-core-metadata
    spec:
      hostname: edgex-core-metadata
      containers:
      - name: edgex-core-metadata
        image: edgexfoundry/docker-core-metadata-go:1.3.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          protocol: TCP
          containerPort: 48081
        envFrom: 
        - configMapRef:
            name: common-variables
        env: 
        - name: Service_Host
          value: "edgex-core-metadata"
        - name: Notifications_Sender
          value: "edgex-core-metadata" 
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-core-data
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: edgex-core-data
  template:
    metadata:
       labels: 
         app: edgex-core-data
    spec:
      hostname: edgex-core-data
      containers:
      - name: edgex-core-data
        image: edgexfoundry/docker-core-data-go:1.3.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: tcp-48080
          protocol: TCP
          containerPort: 48080
        - name: tcp-5563
          protocol: TCP
          containerPort: 5563
        envFrom: 
        - configMapRef:
            name: common-variables
        env: 
        - name: Service_Host
          value: "edgex-core-data"
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: edgex-core-command
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: edgex-core-command
  template:
    metadata:
      labels: 
        app: edgex-core-command
    spec:
      hostname: edgex-core-command
      containers:
      - name: edgex-core-command
        image: edgexfoundry/docker-core-command-go:1.3.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          protocol: TCP
          containerPort: 48082
        envFrom: 
        - configMapRef:
            name: common-variables
        env: 
        - name: Service_Host
          value: "edgex-core-command"
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
---
#apiVersion: apps/v1
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

---
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
---
apiVersion: v1
kind: Service
metadata:
  name: app-service-mqtt
  namespace: {{.Namespace}}
spec:
  type: NodePort 
  selector:
    app: app-service-mqtt
  ports:
  - name: http
    port: 48101
    protocol: TCP
    targetPort: 48101
    nodePort: 30200
---
apiVersion: apps/v1
kind: Deployment
metadata: 
  name: app-service-mqtt
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels: 
      app: app-service-mqtt
  template:
    metadata:
      labels: 
        app: app-service-mqtt
    spec:
      hostname: app-service-mqtt
      containers:
      - name: app-service-mqtt
        image: edgexfoundry/docker-app-service-configurable:1.1.0
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          protocol: TCP
          containerPort: 48101
        envFrom: 
        - configMapRef:
            name: common-variables
        env:
          - name: edgex_profile
            value: "mqtt-export"
          - name: Service_Host
            value: "edgex-app-service-configurable-mqtt"
          - name: Service_Port
            value: "48101"
          - name: MessageBus_SubscribeHost_Host
            value: "edgex-core-data"
          - name: Binding_PublishTopic
            value: "events"
          - name: Writable_Pipeline_Functions_MQTTSend_Addressable_Address
            value: "broker.mqttdashboard.com"
          - name: Writable_Pipeline_Functions_MQTTSend_Addressable_Port
            value: "1883"
          - name: Writable_Pipeline_Functions_MQTTSend_Addressable_Protocol
            value: "tcp"
          - name: Writable_Pipeline_Functions_MQTTSend_Addressable_Publisher
            value: "edgex"
          - name: Writable_Pipeline_Functions_MQTTSend_Addressable_Topic
            value: "EdgeXEvents"

`
