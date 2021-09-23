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

//The components in core services
const EDGEX_CORE = "edgex-core-services.yml"

const EDGEX_CORE_YAML = `
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
`
