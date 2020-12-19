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

const HelperJobRbacYaml = `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: helper
  namespace: kube-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: helper
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin # TODO add ClusterRole
subjects:
  - kind: ServiceAccount
    name: helper
    namespace: kube-system
`

const APP_HELPER_JOB = "helper-job.yaml"

const HelperJobYaml = `
---
apiVersion: batch/v1
kind: Job
metadata:
  name: helper-{{.NodeName}}
  labels:
    app: helper
    k8s-app: helper-{{.NodeName}}
    superedge.io.node: {{.NodeRole}}-helper
  namespace: kube-system
spec:
  completions: 1
  parallelism: 1
  ttlSecondsAfterFinished: 60
  activeDeadlineSeconds: 300
  backoffLimit: 3
  template:
    metadata:
      labels:
        app: helper
        k8s-app: helper-{{.NodeName}}
      annotations:
        superedge.io/app-class: helper
    spec:
      hostPID: true
      containers:
        - name: helper
          image: superedge/helper-job:v0.1.0
          command:
            - /bin/sh
            - -c
          args:
            - cp /usr/local/bin/helper /tmp/host/ && nsenter -m -u -i -n -t 1 /tmp/helper
          env:
            - name: ACTION
              value: {{.Action}}
            - name: MASTER_IP
              value: {{.MasterIP}}
            - name: KUBECONF
              value: {{.KubeConf}}
            - name: NODE_ROLE
              value: {{.NodeRole}}
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
          resources:
            requests:
              cpu: 10m
            limits:
              cpu: 20m
          securityContext:
            privileged: true
          imagePullPolicy: Always
          volumeMounts:
          - mountPath: /tmp/host
            name: host-tmp
      restartPolicy: OnFailure
      serviceAccount: helper
      serviceAccountName: helper
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      nodeSelector:
        kubernetes.io/hostname: {{.NodeName}}
      volumes:
        - name: host-tmp
          hostPath:
            path: /tmp
            type: Directory
`
