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

const APP_EDGE_HEALTH_ADMISSION = "edge-health-admission.yaml"

const EdgeHealthAdmissionYaml = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: edge-health-admission
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: edge-health-admission
rules:
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - "*"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: edge-health-admission
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: edge-health-admission
subjects:
  - kind: ServiceAccount
    name: edge-health-admission
    namespace: kube-system
---
apiVersion: v1
data:
  server.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURsekNDQW4rZ0F3SUJBZ0lKQU9kQ3lvTEpROGJDTUEwR0NTcUdTSWIzRFFFQkJRVUFNQlF4RWpBUUJnTlYKQkFNTUNWZHBjMlV5WXlCRFFUQWVGdzB5TURBM01UZ3dORE01TVRKYUZ3MDBOekV5TURRd05ETTVNVEphTUdNeApDekFKQmdOVkJBWVRBa05PTVJFd0R3WURWUVFJREFoVGFHVnVXbWhsYmpFTE1Ba0dBMVVFQnd3Q1Uxb3hEekFOCkJnTlZCQW9NQmxkcGMyVXlZekVQTUEwR0ExVUVDd3dHVjJselpUSmpNUkl3RUFZRFZRUUREQWxYYVhObE1tTWcKUTBFd2dnRWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUURnOWVmeC9SaDNvRURUNHFjRwpyK1NNRDZPdjZQRmkvbVpYWXRNQ3NTMk1USlRMTVJuazZ2ZkJjelBqeGFXVW9TWXJTTVhvaE5yMHM4YWZ5dlVEClVrelA5OVVndnh5dDNsaUZkQzJZZTJWWEtLRitSNGk0cEtoYlhaT2ZqVUVIeWlDS3cyd1V6L2l3L3VzQktVMk0KdytaQytvb0VjMm5MWXI5VFVlSnhZcWVlSU1KNlRjVXNQdFl0VDgyeU8rbVR1U0JSQ1BhK1VkS2d2RHRtQy9iSgptaHlkbDZXbERuMTJpeTdUUWh0dndvWEsyQjE4bUxBVEtOMW45S2RzMnVPOC8wektUZC9JUmNvV3VnNklPUlVZCjVlU2t5L3JyTjRReVp3ejJHemh4eVVmK3kyeHNLSW44MCtWY2FrUnpuRS9Zb2p3UWkvUkhpTlJ4SHFrS3A2NEsKUkxQekFnTUJBQUdqZ1p3d2daa3dMZ1lEVlIwakJDY3dKYUVZcEJZd0ZERVNNQkFHQTFVRUF3d0pWMmx6WlRKagpJRU5CZ2drQW5XZnkrb3hVU1Jvd0NRWURWUjBUQkFJd0FEQUxCZ05WSFE4RUJBTUNCREF3SFFZRFZSMGxCQll3CkZBWUlLd1lCQlFVSEF3RUdDQ3NHQVFVRkJ3TUNNREFHQTFVZEVRUXBNQ2VDSldWa1oyVXRhR1ZoYkhSb0xXRmsKYldsemMybHZiaTVyZFdKbExYTjVjM1JsYlM1emRtTXdEUVlKS29aSWh2Y05BUUVGQlFBRGdnRUJBRUd3RDNscgphZDFXYzZtRE1DMEpaZmxnNmJFdnBLTW1Yd1BhNTJXQTRuZ3hLYjMxSzJXbFo3SDJrTFBaTndLVy9tcElzdFoyCnozbjVrVTFKZGFtM2pPZVptMWI5d0N2U01Lek5rczRDMG5lblY0ODhhTU1JSUgwZEFNbVZSeFlEVXhZM1NweHkKVDdsbnRQQ3dXOWlyQVBmVlNURHk1ZzN5MHR0SVpPRmpXMjRQMnFKaG9Fa3RMVURSNGdMMEhDb3FUZDhMNnlSRwoyV2UraHZtNG0zRVpKaFg5MDY1NDhXSG9CSHlicVcrdll6STBjSGU5bWp1eFIvU0VkUlhoVXV5dzNsSjFyT0RtCjEyM2k1VEprUDRyTVJlZHdvc3YxbW9kbml2SExHa1RkNTBZcmREOVdPdE55WUlBTlFkUDhDcFRNc1dBdHFaejEKL1RlamZxejN2OERXYlpNPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
  server.key: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBNFBYbjhmMFlkNkJBMCtLbkJxL2tqQStqcitqeFl2NW1WMkxUQXJFdGpFeVV5ekVaCjVPcjN3WE16NDhXbGxLRW1LMGpGNklUYTlMUEduOHIxQTFKTXovZlZJTDhjcmQ1WWhYUXRtSHRsVnlpaGZrZUkKdUtTb1cxMlRuNDFCQjhvZ2lzTnNGTS80c1A3ckFTbE5qTVBtUXZxS0JITnB5MksvVTFIaWNXS25uaURDZWszRgpMRDdXTFUvTnNqdnBrN2tnVVFqMnZsSFNvTHc3Wmd2Mnlab2NuWmVscFE1OWRvc3UwMEliYjhLRnl0Z2RmSml3CkV5amRaL1NuYk5yanZQOU15azNmeUVYS0Zyb09pRGtWR09Ya3BNdjY2emVFTW1jTTloczRjY2xIL3N0c2JDaUoKL05QbFhHcEVjNXhQMktJOEVJdjBSNGpVY1I2cENxZXVDa1N6OHdJREFRQUJBb0lCQUdtcFZwNUVvSDlmbDlOdAp1OEhhVCtDeFg5SzUrTmJrWXJGY3kzdVNPTENUTDdnWWdlOFJwZmtJNFRCMG53Y21nY1VHMDE0Wk9MYUMwaUl2CnM5RXhrTDZGeTJjc0hJNVZ4d0kzeFVxL2VxUHJnNTdLZnA4clI2QlNYWW90VUlRV0hoN1BGeTdYV0JuYVFnc3oKbVNjcXhEWmxjdm9RTTNyQ0VOZFR1S1pGRGpHb2s4OGU2WEo1QUQzMnQxZDB4NW80VTdLNm9FQnVZeEp6RDhtNQoySE1xSkRmdVZ6aUp0aThJT3BkMUVja2NvdkRCMmVwbDFPa2hUTWtrU0d4dW83b3JwVkRqalFEZ2YybVd4bG1tCndGd3FwRmtUYW9UdkdjYm03S0VKSGdTRXhqVWdKMlFhZnB2NC9DcU8zLzZWdHhXSWpNL2c5V2p3ZzBGb05RK04KSjIrZG5LRUNnWUVBK2lGbGJRRXkvL1BXV1N5ZGlRQTdLMkNDb2RtTjJGUU45dHlXSHBtM1ZyVWJ6YjdFS2JyMgpaRk93YklhbHhEblIzUW1yVk1wdEdEL2hJN2VQbXlnOG5WMHQ5U1JZL1hPNFRZdlg2SzdXMTIvRllLbyt2Q3lECk1oZUZhQ2ljRzI1am9DUklUQVBqNS9SS3loQTUrYk0xVFI3a1JQTTNJMU1nME9uUjl2U2FuUGtDZ1lFQTVqMU8KcXArZnR6cWhldXljR3psMUtKZloydGtJclRLN1k2YWpTTjR0M0NtaGhqMktNVEk5QVRESDZkY1VHMEhZSTZOcQpMVlhqZFJrSDhRUFRvK3hmU2thbEpnaVh0L3BMdU1QNzlnUTB2T2Vtakw2cjJPd1dGejk1b040dmtDNDNuODA5CnR0TVp2Ti9GMW5UM3ZOeWdJSTZTMVRodEVYMXZ6M0tpZGR3Zkwwc0NnWUFBd0hyeWtlOWFUNXhVVmtyKzcyNCsKR2lNcVkySUd4WEhwVFE1eWR4blMrK1ppZnZGT0FzN2N6RmVhYStreHBzN1hzRURBbDM4dWRIcXp4Y2g3dWVvOAp1dHY1Z2F0Mno1TTlRRzljdHJIVW9mUmc3d0lUUkxyOE9vL2ZHVWdtMlBVWnRTSTJnRWgrR1FEa2pKbndBemJrCnpYUDROUmIwVnpxaEJpTG9jQ0hLMlFLQmdDS0RUaWVGaGd1UlhtTnUxSGZBUlMrd2s1ZWFzUkpGYUpHbmlSS0QKTzV5bElQRmVpRGlYcjAxZVlwbExCRmlScGpTeGFsa2hadGRHeVVuM3FPSUpyTDhWbCt2N25jS1dZb052M1hVagpiRVJrOVRKajRwN0J4UTMzRmVSbmFmblM4OE9nb0grblpWUkt0djFPeTFRa1BseWpBcCt6dGFYSmg5a3c5ZWwwCjliZkJBb0dCQU82Ylhnbmk0VWxTNG5mZGZIYUxUdVJVRmFrWDJRQXFGOTArOHcxa1hTRUhmK2syQjMzN2pWTU4KRVcxOGl1dldnZ29nNlRUM0ZTVWNmVUtUc2tMdGpucmxYc1ROSE90VW9XMTdiY01LcDduTEQveXY2Q2tKd3hacQppSDJ5cU5selB2WE50L2l2T2ErQVdNYm5haGtMQWpBNGRxVWxrVnFZa2JuMFNIUXd5cStLCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==
kind: Secret
metadata:
  name: validate-admission-control-server-certs
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: edge-health-admission
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: edge-health-admission
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        k8s-app: edge-health-admission
    spec:
      containers:
        - name: edge-health-admission
          image: superedge/edge-health-admission:v0.1.0
          imagePullPolicy: Always
          command:
            - edge-health-admission
          args:
            - --alsologtostderr
            - --v=4
            - --admission-control-server-cert=/etc/edge-health-admission/certs/server.crt
            - --admission-control-server-key=/etc/edge-health-admission/certs/server.key
          env:
            - name: MASTER_NAMESPACE
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: metadata.namespace
          volumeMounts:
            - mountPath: /etc/edge-health-admission/certs
              name: admission-server-certs
          resources:
            limits:
              cpu: 50m
              memory: 100Mi
            requests:
              cpu: 10m
              memory: 20Mi
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      volumes:
        - secret:
            secretName: validate-admission-control-server-certs
          name: admission-server-certs
      nodeSelector:
        node-role.kubernetes.io/master: ""
      tolerations:
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
      serviceAccountName: edge-health-admission
---
apiVersion: v1
kind: Service
metadata:
  name: edge-health-admission
  namespace: kube-system
spec:
  ports:
    - port: 443
      protocol: TCP
      targetPort: 8443
  selector:
    k8s-app: edge-health-admission
  type: ClusterIP
`
