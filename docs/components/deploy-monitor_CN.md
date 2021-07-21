# 部署监控系统
监控系统分为2个部分：prometheus-server、node-exporter

## 1. 部署prometheus-server

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus-server
rules:
  - apiGroups:
      - ""
    resources:
      - nodes
      - nodes/proxy
      - nodes/metrics
      - services
      - endpoints
      - pods
      - ingresses
      - configmaps
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - extensions
      - networking.k8s.io
    resources:
      - ingresses/status
      - ingresses
    verbs:
      - get
      - list
      - watch
  - nonResourceURLs:
      - /metrics
    verbs:
      - get
---
apiVersion: v1
data:
  alerting_rules.yml: |
    {}
  alerts: |
    {}
  prometheus.yml: |
    global:
      evaluation_interval: 1m
      scrape_interval: 1m
      scrape_timeout: 10s
    rule_files:
    - /etc/config/recording_rules.yml
    - /etc/config/alerting_rules.yml
    - /etc/config/rules
    - /etc/config/alerts
    scrape_configs:
    - job_name: prometheus
      static_configs:
      - targets:
        - 0.0.0.0:9090
    - bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      job_name: kubernetes-apiservers
      kubernetes_sd_configs:
      - role: endpoints
      relabel_configs:
      - action: keep
        regex: default;kubernetes;https
        source_labels:
        - __meta_kubernetes_namespace
        - __meta_kubernetes_service_name
        - __meta_kubernetes_endpoint_port_name
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        insecure_skip_verify: true
    - bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      job_name: kubernetes-nodes
      kubernetes_sd_configs:
      - role: node
      relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
      - replacement: kubernetes.default.svc:443
        target_label: __address__
      - regex: (.+)
        replacement: /api/v1/nodes/$1/proxy/metrics
        source_labels:
        - __meta_kubernetes_node_name
        target_label: __metrics_path__
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        insecure_skip_verify: true
    - bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      job_name: kubernetes-nodes-cadvisor
      kubernetes_sd_configs:
      - role: node
      relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
      - replacement: kubernetes.default.svc:443
        target_label: __address__
      - regex: (.+)
        replacement: /api/v1/nodes/$1/proxy/metrics/cadvisor
        source_labels:
        - __meta_kubernetes_node_name
        target_label: __metrics_path__
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        insecure_skip_verify: true
    - job_name: tunnel-cloud-metrics
      kubernetes_sd_configs:
      - role: endpoints
      scheme: http
      relabel_configs:
      - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: edge-system;tunnel-metrics
    - job_name: node-cadvisor
      kubernetes_sd_configs:
      - role: node
      scheme: https
      tls_config:
        insecure_skip_verify: true
      relabel_configs:
      - source_labels: [__name__]
        regex: '(container_tasks_state|container_memory_failures_total)'
        action: drop
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __address__
        replacement: ${1}:10250
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /metrics/cadvisor
      - source_labels: [__address__]
        target_label: "unInstanceId"
        replacement: "none"
    - job_name: node-exporter
      kubernetes_sd_configs:
      - role: node
      scheme: https
      tls_config:
        insecure_skip_verify: true
      relabel_configs:
      - source_labels: [__name__]
        regex: '(container_tasks_state|container_memory_failures_total)'
        action: drop
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __address__
        replacement: ${1}:9100
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /metrics
      - source_labels: [__address__]
        target_label: "unInstanceId"
        replacement: "none"
    - job_name: kubernetes-pods-slow
      kubernetes_sd_configs:
      - role: pod
      relabel_configs:
      - action: keep
        regex: true
        source_labels:
        - __meta_kubernetes_pod_annotation_prometheus_io_scrape_slow
      - action: replace
        regex: (https?)
        source_labels:
        - __meta_kubernetes_pod_annotation_prometheus_io_scheme
        target_label: __scheme__
      - action: replace
        regex: (.+)
        source_labels:
        - __meta_kubernetes_pod_annotation_prometheus_io_path
        target_label: __metrics_path__
      - action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        source_labels:
        - __address__
        - __meta_kubernetes_pod_annotation_prometheus_io_port
        target_label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
      - action: replace
        source_labels:
        - __meta_kubernetes_namespace
        target_label: kubernetes_namespace
      - action: replace
        source_labels:
        - __meta_kubernetes_pod_name
        target_label: kubernetes_pod_name
      - action: drop
        regex: Pending|Succeeded|Failed
        source_labels:
        - __meta_kubernetes_pod_phase
      scrape_interval: 5m
      scrape_timeout: 30s
    alerting:
      alertmanagers:
      - kubernetes_sd_configs:
          - role: pod
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
        relabel_configs:
        - source_labels: [__meta_kubernetes_namespace]
          regex: edge-system
          action: keep
        - source_labels: [__meta_kubernetes_pod_label_app]
          regex: prometheus
          action: keep
        - source_labels: [__meta_kubernetes_pod_label_component]
          regex: alertmanager
          action: keep
        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_probe]
          regex: .*
          action: keep
        - source_labels: [__meta_kubernetes_pod_container_port_number]
          regex: "9093"
          action: keep
  recording_rules.yml: |
    {}
  rules: |
    {}
kind: ConfigMap
metadata:
  name: prometheus-server
  namespace: edge-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus-server
  namespace: edge-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prometheus-server
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-server
subjects:
  - kind: ServiceAccount
    name: prometheus-server
    namespace: edge-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus-server
  namespace: edge-system
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: prometheus
      component: server
      release: prometheus
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: prometheus
        component: server
        release: prometheus
    spec:
      containers:
        - args:
            - --volume-dir=/etc/config
            - --webhook-url=http://127.0.0.1:9090/-/reload
          image: jimmidyson/configmap-reload:v0.5.0
          imagePullPolicy: IfNotPresent
          name: prometheus-server-configmap-reload
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
            - mountPath: /etc/config
              name: config-volume
              readOnly: true
        - args:
            - --storage.tsdb.retention.time=15d
            - --config.file=/etc/config/prometheus.yml
            - --storage.tsdb.path=/home
            - --web.console.libraries=/etc/prometheus/console_libraries
            - --web.console.templates=/etc/prometheus/consoles
            - --web.enable-lifecycle
          image: quay.io/prometheus/prometheus:v2.26.0
          imagePullPolicy: IfNotPresent
          livenessProbe:
            failureThreshold: 3
            httpGet:
              path: /-/healthy
              port: 9090
              scheme: HTTP
            initialDelaySeconds: 30
            periodSeconds: 15
            successThreshold: 1
            timeoutSeconds: 10
          name: prometheus-server
          ports:
            - containerPort: 9090
              hostPort: 9090
              protocol: TCP
          readinessProbe:
            failureThreshold: 3
            httpGet:
              path: /-/ready
              port: 9090
              scheme: HTTP
            initialDelaySeconds: 30
            periodSeconds: 5
            successThreshold: 1
            timeoutSeconds: 4
          volumeMounts:
            - mountPath: /etc/config
              name: config-volume
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext:
        fsGroup: 65534
        runAsGroup: 65534
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccount: prometheus-server
      serviceAccountName: prometheus-server
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
          operator: Exists
      volumes:
        - configMap:
            defaultMode: 420
            name: prometheus-server
          name: config-volume
---
apiVersion: v1
kind: Service
metadata:
  name: prometheus-server
  namespace: edge-system
spec:
  ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: 9090
  selector:
    app: prometheus
    component: server
    release: prometheus
  type: ClusterIP
```

## 2. 部署prometheus-node-exporter

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus-node-exporter
  namespace: edge-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-node-exporter-conf
  namespace: edge-system
data:
  config.yaml: |
    tls_server_config:
      cert_file: /home/certs/apiserver-kubelet-server.crt
      key_file: /home/certs/apiserver-kubelet-server.key
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: prometheus-node-exporter
  namespace: edge-system
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: prometheus
      component: node-exporter
      release: prometheus
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: prometheus
        chart: prometheus-14.3.1
        component: node-exporter
        heritage: Helm
        release: prometheus
    spec:
      containers:
        - args:
            - --path.procfs=/host/proc
            - --path.sysfs=/host/sys
            - --path.rootfs=/host/root
            - --web.config=/home/conf/config.yaml
            - --web.listen-address=:9100
          image: quay.io/prometheus/node-exporter:v1.1.2
          imagePullPolicy: IfNotPresent
          name: prometheus-node-exporter
          ports:
            - containerPort: 9100
              hostPort: 9100
              name: metrics
              protocol: TCP
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
            - mountPath: /host/proc
              name: proc
              readOnly: true
            - mountPath: /host/sys
              name: sys
              readOnly: true
            - mountPath: /host/root
              mountPropagation: HostToContainer
              name: root
              readOnly: true
            - mountPath: /home/conf
              name: conf
            - mountPath: /home/certs
              name: certs
      dnsPolicy: ClusterFirst
      hostNetwork: true
      hostPID: true
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext:
        fsGroup: 65534
        runAsGroup: 65534
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccount: prometheus-node-exporter
      serviceAccountName: prometheus-node-exporter
      terminationGracePeriodSeconds: 30
      volumes:
        - hostPath:
            path: /proc
            type: ""
          name: proc
        - hostPath:
            path: /sys
            type: ""
          name: sys
        - hostPath:
            path: /
            type: ""
          name: root
        - configMap:
            defaultMode: 420
            name: prometheus-node-exporter-conf
          name: conf
        - name: certs
          secret:
            defaultMode: 420
            secretName: tunnel-cloud-cert
```

## 3. 验证部署是否成功

### 3.1 验证是否采集到kubelet metrics

```shell
$ curl -G  http://<prometheus-server的clusterip>/api/v1/series? --data-urlencode 'match[]=container_processes'
{
  [
    {
      "__name__": "container_processes",
      "id": "/system.slice/docker.service",
      "instance": "edge-7x94bd",
      "job": "node-cadvisor",
      "unInstanceId": "none"
    },
    {
      "__name__": "container_processes",
      "id": "/system.slice/kubelet.service",
      "instance": "edge-7x94bd",
      "job": "node-cadvisor",
      "unInstanceId": "none"
    }
  ]
}
```

### 3.2 验证是否采集到node metrics

```shell
curl -G  http://<prometheus-server的clusterip>/api/v1/series? --data-urlencode 'match[]=node_cpu_guest_seconds_total'
{
  "status": "success",
  "data": [
    {
      "__name__": "node_cpu_guest_seconds_total",
      "cpu": "0",
      "instance": "edge-7x94bd",
      "job": "node-exporter",
      "mode": "nice",
      "unInstanceId": "none"
    },
    {
      "__name__": "node_cpu_guest_seconds_total",
      "cpu": "0",
      "instance": "edge-7x94bd",
      "job": "node-exporter",
      "mode": "user",
      "unInstanceId": "none"
    },
    {
      "__name__": "node_cpu_guest_seconds_total",
      "cpu": "1",
      "instance": "edge-7x94bd",
      "job": "node-exporter",
      "mode": "nice",
      "unInstanceId": "none"
    },
    {
      "__name__": "node_cpu_guest_seconds_total",
      "cpu": "1",
      "instance": "edge-7x94bd",
      "job": "node-exporter",
      "mode": "user",
      "unInstanceId": "none"
    }
  ]
}
```