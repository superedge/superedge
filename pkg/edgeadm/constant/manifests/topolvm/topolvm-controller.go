package manifests

const AppTopolvmController = "topolvm-controller.yaml"

const AppTopolvmControllerYaml = `
# Source: topolvm/templates/controller/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: topolvm-controller
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
---

## topolvm-system:controller RBAC
# Source: topolvm/templates/controller/clusterroles.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: topolvm-system:controller
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch", "patch", "update"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update", "delete"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses","csidrivers"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["topolvm.cybozu.com"]
    resources: ["logicalvolumes", "logicalvolumes/status"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
# Source: topolvm/templates/controller/clusterrolebindings.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: topolvm-system:controller
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
subjects:
  - kind: ServiceAccount
    namespace: topolvm-system
    name: topolvm-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: topolvm-system:controller
---


## topolvm-external-provisioner-runner RBAC
# Source: topolvm/templates/controller/clusterroles.yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: topolvm-external-provisioner-runner
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshots"]
    verbs: ["get", "list"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshotcontents"]
    verbs: ["get", "list"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["csinodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch"]
---
# Source: topolvm/templates/controller/clusterrolebindings.yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: topolvm-csi-provisioner-role
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
subjects:
  - kind: ServiceAccount
    namespace: topolvm-system
    name: topolvm-controller
roleRef:
  kind: ClusterRole
  name: topolvm-external-provisioner-runner
  apiGroup: rbac.authorization.k8s.io
---


## external-resizer-cfg RBAC
# Source: topolvm/templates/controller/roles.yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: external-resizer-cfg
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
rules:
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
---
# Source: topolvm/templates/controller/rolebinding.yaml
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-resizer-role-cfg
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
subjects:
  - kind: ServiceAccount
    name: topolvm-controller
    namespace: topolvm-system
roleRef:
  kind: Role
  name: external-resizer-cfg
  apiGroup: rbac.authorization.k8s.io
---

## topolvm-external-resizer-runner RBAC
# Source: topolvm/templates/controller/clusterroles.yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: topolvm-external-resizer-runner
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "patch"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims/status"]
    verbs: ["patch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
---
# Source: topolvm/templates/controller/clusterrolebindings.yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: topolvm-csi-resizer-role
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
subjects:
  - kind: ServiceAccount
    namespace: topolvm-system
    name: topolvm-controller
roleRef:
  kind: ClusterRole
  name: topolvm-external-resizer-runner
  apiGroup: rbac.authorization.k8s.io
---

## leader-election RBAC 
# Source: topolvm/templates/controller/roles.yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: leader-election
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
rules:
  - apiGroups: ["", "coordination.k8s.io"]
    resources: ["configmaps", "leases"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]
---
# Source: topolvm/templates/controller/rolebinding.yaml
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: leader-election
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
subjects:
  - kind: ServiceAccount
    namespace: topolvm-system
    name: topolvm-controller
roleRef:
  kind: Role
  name: leader-election
  apiGroup: rbac.authorization.k8s.io
---


## external-provisioner-cfg RBAC
# Source: topolvm/templates/controller/roles.yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: external-provisioner-cfg
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
rules:
  # Only one of the following rules for endpoints or leases is required based on
  # what is set for --leader-election-type. Endpoints are deprecated in favor of Leases.
  # - apiGroups: [""]
  #   resources: ["endpoints"]
  #   verbs: ["get", "watch", "list", "delete", "update", "create"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
  # Permissions for CSIStorageCapacity are only needed enabling the publishing
  # of storage capacity information.
  - apiGroups: ["storage.k8s.io"]
    resources: ["csistoragecapacities"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  # The GET permissions below are needed for walking up the ownership chain
  # for CSIStorageCapacity. They are sufficient for deployment via
  # StatefulSet (only needs to get Pod) and Deployment (needs to get
  # Pod and then ReplicaSet to find the Deployment).
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get"]
  - apiGroups: ["apps"]
    resources: ["replicasets"]
    verbs: ["get"]
---
# Source: topolvm/templates/controller/rolebinding.yaml
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: csi-provisioner-role-cfg
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
subjects:
  - kind: ServiceAccount
    namespace: topolvm-system
    name: topolvm-controller
roleRef:
  kind: Role
  name: external-provisioner-cfg
  apiGroup: rbac.authorization.k8s.io
---

# Source: topolvm/templates/controller/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: topolvm-controller
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: topolvm-controller
  template:
    metadata:
      labels:
        app.kubernetes.io/name: topolvm-controller
    spec:
      containers:
        - name: topolvm-controller
          image: "quay.io/topolvm/topolvm-with-sidecar:0.10.0"
          command:
            - /topolvm-controller
            - --cert-dir=/certs
          ports:
            - containerPort: 9808
              name: healthz
              protocol: TCP
          readinessProbe:
            httpGet:
              path: /metrics
              port: 8080
              scheme: HTTP
          livenessProbe:
            httpGet:
              path: /healthz
              port: healthz
            failureThreshold: 3
            initialDelaySeconds: 10
            timeoutSeconds: 3
            periodSeconds: 60
          volumeMounts:
            - name: socket-dir
              mountPath: /run/topolvm
            - name: certs
              mountPath: /certs
        - name: csi-provisioner
          image: "quay.io/topolvm/topolvm-with-sidecar:0.10.0"
          command:
            - /csi-provisioner
            - "--csi-address=/run/topolvm/csi-topolvm.sock"
            - "--feature-gates=Topology=true"
            - --leader-election
            - --leader-election-namespace=topolvm-system
          volumeMounts:
            - name: socket-dir
              mountPath: /run/topolvm
        - name: csi-resizer
          image: "quay.io/topolvm/topolvm-with-sidecar:0.10.0"
          command:
            - /csi-resizer
            - "--csi-address=/run/topolvm/csi-topolvm.sock"
            - --leader-election
            - --leader-election-namespace=topolvm-system
          volumeMounts:
            - name: socket-dir
              mountPath: /run/topolvm
        - name: liveness-probe
          image: "quay.io/topolvm/topolvm-with-sidecar:0.10.0"
          command:
            - /livenessprobe
            - "--csi-address=/run/topolvm/csi-topolvm.sock"
          volumeMounts:
            - name: socket-dir
              mountPath: /run/topolvm
      securityContext: 
        runAsGroup: 10000
        runAsUser: 10000
      serviceAccountName: topolvm-controller
      nodeSelector:
        node-role.kubernetes.io/master: ""
      affinity: 
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app.kubernetes.io/name
                operator: In
                values:
                - topolvm-controller
            topologyKey: kubernetes.io/hostname
      tolerations:
        - key: "node-role.kubernetes.io/master"
          operator: "Exists"
          effect: "NoSchedule"
      volumes:
        - name: certs
          secret:
            secretName: topolvm-mutatingwebhook
        - emptyDir: {}
          name: socket-dir
---
# Source: topolvm/templates/controller/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: topolvm-controller
  namespace: topolvm-system
  labels:
    app.kubernetes.io/name: topolvm
    app.kubernetes.io/version: "0.9.0"
spec:
  selector:
    app.kubernetes.io/name: topolvm-controller
  ports:
    - protocol: TCP
      port: 443
      targetPort: 9443
---
`
