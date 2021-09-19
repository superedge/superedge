package manifests

const AppTopolvmScheduler = "topolvm-scheduler.yaml"

const AppTopolvmSchedulerYaml = `
## topolvm-scheduler RBAC
# Source: topolvm/templates/scheduler/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: topolvm-scheduler
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
---
# Source: topolvm/templates/scheduler/role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: psp:topolvm-scheduler
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
rules:
  - apiGroups: ["policy"]
    resources: ["podsecuritypolicies"]
    verbs: ["use"]
    resourceNames: ["topolvm-scheduler"]
---
# Source: topolvm/templates/scheduler/rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: topolvm-scheduler:psp:topolvm-scheduler
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: psp:topolvm-scheduler
subjects:
  - kind: ServiceAccount
    namespace: topolvm-system
    name: topolvm-scheduler
---

# Source: topolvm/templates/scheduler/psp.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: topolvm-scheduler
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
spec:
  privileged: false
  allowPrivilegeEscalation: false
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'secret'
  hostNetwork: true
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'MustRunAs'
    ranges:
      - min: 1
        max: 65535
  fsGroup:
    rule: 'MayRunAs'
    ranges:
      - min: 1
        max: 65535
  readOnlyRootFilesystem: true
---

# Source: topolvm/templates/scheduler/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: topolvm-scheduler-options
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
data:
  scheduler-options.yaml: |
    listen: "localhost:9251"
    default-divisor: 1
---

# Source: topolvm/templates/scheduler/daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: topolvm-scheduler
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: topolvm-scheduler
  template:
    metadata:
      annotations:
        checksum/config: 534abbdd3e5b84154285589e5168700e69db27740b6dcac09486eacdab5c1e97
      labels:
        app.kubernetes.io/name: topolvm-scheduler
    spec:
      securityContext: 
        runAsGroup: 10000
        runAsUser: 10000
      serviceAccountName: topolvm-scheduler
      containers:
        - name: topolvm-scheduler
          image: "quay.io/topolvm/topolvm-with-sidecar:0.10.0"
          command:
            - /topolvm-scheduler
            - --config=/etc/topolvm/scheduler-options.yaml
          livenessProbe:
            httpGet:
              host: localhost
              port: 9251
              path: /status
          volumeMounts:
            - mountPath: /etc/topolvm
              name: topolvm-scheduler-options
      hostNetwork: true
      volumes:
        - name: topolvm-scheduler-options
          configMap:
            name: topolvm-scheduler-options
      tolerations: 
        - key: CriticalAddonsOnly
          operator: Exists
        - effect: NoSchedule
          key: node-role.kubernetes.io/control-plane
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
          operator: "Exists"
`
